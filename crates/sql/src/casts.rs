// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//! The casts of a root value's cast collection, which CREATE CAST and emulated extensions create.

use std::cell::RefCell;
use std::sync::Arc;

use pg_query::NodeEnum;
use pg_query::protobuf::{CoercionContext, CreateCastStmt, DropStmt};

use crate::Outcome;
use crate::catalog::{ColumnType, id};
use crate::error::{PgError, Result, code};
use crate::oid;
use crate::query::Ctx;
use crate::routines::Routine;

/// COLLECTION is the position of the casts in a root value's root object collections.
pub const COLLECTION: usize = 7;

/// The cast contexts, as Go numbers them.
pub const EXPLICIT: u8 = 0;
pub const ASSIGNMENT: u8 = 1;
pub const IMPLICIT: u8 = 2;

/// UserCast is a cast of the cast collection, with the routine that performs it, or None for a cast through text.
pub struct UserCast {
    pub source: u32,
    pub target: u32,
    pub context: u8,
    pub stored: objects::Cast,
    pub routine: Option<Arc<Routine>>,
}

/// CastCache is the casts of a root value, with the addresses of the cast and function collections.
pub type CastCache = ((Option<store::Hash>, Option<store::Hash>), Arc<Vec<UserCast>>);

thread_local! {
    /// INSTALLED is the casts of the running statement's root value.
    static INSTALLED: RefCell<Arc<Vec<UserCast>>> = RefCell::new(Arc::new(Vec::new()));
}

/// find returns the installed cast from one type to another.
pub fn find(source: u32, target: u32) -> Option<(u8, Option<Arc<Routine>>)> {
    INSTALLED.with(|casts| {
        casts.borrow().iter().find(|c| c.source == source && c.target == target).map(|c| (c.context, c.routine.clone()))
    })
}

/// context returns the context of the installed cast from one type to another.
pub fn context(source: u32, target: u32) -> Option<u8> {
    find(source, target).map(|(context, _)| context)
}

impl Ctx<'_> {
    /// user_casts returns the casts of the working root, reusing them while their collections are unchanged.
    pub fn user_casts(&mut self) -> Result<Arc<Vec<UserCast>>> {
        let address = (self.txn.root.root_objects[COLLECTION], self.txn.root.root_objects[crate::routines::COLLECTION]);
        if let Some((cached, casts)) = &self.session.casts
            && *cached == address
        {
            return Ok(casts.clone());
        }
        let mut casts = Vec::new();
        if address.0.is_some() {
            let routines = self.routines()?;
            for (_, address) in self.txn.root.objects(self.db, COLLECTION)? {
                let stored = objects::Cast::deserialize(&prolly::read_blob(self.db, &address)?)?;
                let segments = id::segments(&stored.id);
                let [source, target] = segments.as_slice() else { continue };
                casts.push(UserCast {
                    source: crate::usertypes::type_oid(source.as_bytes()),
                    target: crate::usertypes::type_oid(target.as_bytes()),
                    context: stored.cast_type,
                    routine: routines.iter().find(|r| r.object.id == stored.function).cloned(),
                    stored,
                });
            }
        }
        let casts = Arc::new(casts);
        self.session.casts = Some((address, casts.clone()));
        Ok(casts)
    }

    /// install_casts makes the working root's casts known to this thread's statement.
    pub fn install_casts(&mut self) -> Result<()> {
        let casts = self.user_casts()?;
        INSTALLED.with(|installed| *installed.borrow_mut() = casts);
        Ok(())
    }
}

/// context_code returns the pg_cast context letter of a Go cast context.
pub fn context_code(context: u8) -> &'static str {
    match context {
        IMPLICIT => "i",
        ASSIGNMENT => "a",
        _ => "e",
    }
}

/// invalid_definition returns Postgres' error for a CREATE CAST that cannot work.
fn invalid_definition(message: &str) -> PgError {
    PgError::new(code::INVALID_OBJECT_DEFINITION, message)
}

/// definition returns a type's stored definition, built in or user-defined.
fn definition(type_oid: u32) -> Option<objects::SerializedType> {
    match crate::catalog::builtin_type(type_oid) {
        Some(t) => Some(t.definition.clone()),
        None => crate::usertypes::get(type_oid).map(|t| t.definition.clone()),
    }
}

impl Ctx<'_> {
    /// cast_type resolves a type that CREATE CAST or DROP CAST names, which may be a table's row type.
    fn cast_type(&mut self, name: &Option<pg_query::protobuf::TypeName>) -> Result<ColumnType> {
        let name = name.as_ref().ok_or_else(|| PgError::internal("a cast without a type"))?;
        self.prepare_type(name)?;
        crate::expr::resolve_type_name(name)
    }

    /// create_cast runs CREATE CAST, checking the cast as Postgres' CreateCast does, and stores it as Go stores a
    /// cast.
    pub fn create_cast(&mut self, create: &CreateCastStmt) -> Result<Outcome> {
        let (source, target) = (self.cast_type(&create.sourcetype)?, self.cast_type(&create.targettype)?);
        let function_args = create.func.as_ref().map_or(0, |f| f.objargs.len());
        if source.oid == target.oid && function_args < 2 {
            return Err(invalid_definition("source data type and target data type are the same"));
        }
        let context = match CoercionContext::try_from(create.context) {
            Ok(CoercionContext::CoercionImplicit) => IMPLICIT,
            Ok(CoercionContext::CoercionAssignment) => ASSIGNMENT,
            _ => EXPLICIT,
        };
        let mut function = Vec::new();
        if let Some(func) = &create.func {
            let routine =
                self.find_routine(func, false, Some(false))?.ok_or_else(|| PgError::internal("a missing function"))?;
            let params = &routine.params;
            if params.is_empty() || params.len() > 3 {
                return Err(invalid_definition("cast function must take one to three arguments"));
            }
            if params[0].ty.oid != source.oid {
                return Err(invalid_definition(
                    "argument of cast function must match or be binary-coercible from source data type",
                ));
            }
            if params.len() > 1 && params[1].ty.oid != oid::INT4 {
                return Err(invalid_definition("second argument of cast function must be type integer"));
            }
            if params.len() > 2 && params[2].ty.oid != oid::BOOL {
                return Err(invalid_definition("third argument of cast function must be type boolean"));
            }
            if routine.ret.oid != target.oid {
                return Err(invalid_definition(
                    "return data type of cast function must match or be binary-coercible to target data type",
                ));
            }
            function = routine.object.id.clone();
        } else if !create.inout {
            let (from, to) = (definition(source.oid), definition(target.oid));
            if [&from, &to].iter().any(|d| d.as_ref().is_some_and(|d| d.typ_type == b"c")) {
                return Err(invalid_definition("composite data types are not binary-compatible"));
            }
            let physical = |d: &Option<objects::SerializedType>| {
                d.as_ref().map(|d| (d.typ_length, d.passed_by_val, d.align.clone()))
            };
            if physical(&from) != physical(&to) {
                return Err(invalid_definition("source and target data types are not physically compatible"));
            }
            return Err(PgError::unsupported("binary-coercible casts"));
        }
        let (source_id, target_id) = (crate::usertypes::type_id(source.oid), crate::usertypes::type_id(target.oid));
        let cast = objects::Cast {
            id: id::new(
                id::SECTION_CAST,
                &[&String::from_utf8_lossy(&source_id), &String::from_utf8_lossy(&target_id)],
            ),
            cast_type: context,
            function,
            use_in_out: create.inout,
        };
        if self.txn.root.objects(self.db, COLLECTION)?.iter().any(|(key, _)| *key == cast.id) {
            return Err(PgError::new(
                code::DUPLICATE_OBJECT,
                format!(
                    "cast from type {} to type {} already exists",
                    crate::cast::type_display(source.oid),
                    crate::cast::type_display(target.oid)
                ),
            ));
        }
        let data = cast.serialize();
        let db = &mut *self.db;
        let mut sink = |_: store::Hash, bytes: &[u8]| {
            db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
        };
        let (address, _) = prolly::write_blob(&data, &mut sink)?.ok_or_else(|| PgError::internal("an empty cast"))?;
        self.txn.root.put_object(self.db, COLLECTION, &cast.id, Some(address))?;
        Ok(Outcome::command("CREATE CAST"))
    }

    /// drop_casts runs DROP CAST.
    pub fn drop_casts(&mut self, drop: &DropStmt) -> Result<Outcome> {
        for object in &drop.objects {
            let Some(NodeEnum::List(list)) = object.node.as_ref() else { continue };
            let mut types = Vec::new();
            for item in &list.items {
                if let Some(NodeEnum::TypeName(name)) = item.node.as_ref() {
                    types.push(self.cast_type(&Some(name.clone()))?);
                }
            }
            let [source, target] = types.as_slice() else { continue };
            let found = self
                .user_casts()?
                .iter()
                .find(|c| c.source == source.oid && c.target == target.oid)
                .map(|c| c.stored.id.clone());
            match found {
                Some(key) => self.txn.root.put_object(self.db, COLLECTION, &key, None)?,
                None => {
                    let message = format!(
                        "cast from type {} to type {} does not exist",
                        crate::cast::type_display(source.oid),
                        crate::cast::type_display(target.oid)
                    );
                    if !drop.missing_ok {
                        return Err(PgError::new(code::UNDEFINED_OBJECT, message));
                    }
                    self.session.notice(PgError::notice("00000", format!("{message}, skipping")));
                }
            }
        }
        Ok(Outcome::command("DROP CAST"))
    }
}
