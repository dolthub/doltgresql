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

//! The operators of a root value's operator collection, which CREATE OPERATOR and emulated extensions create.

use std::collections::HashSet;
use std::sync::{Arc, OnceLock};

use pg_query::NodeEnum;
use pg_query::protobuf::{DefineStmt, DropStmt, Node};

use crate::Outcome;
use crate::catalog::{ColumnType, id};
use crate::error::{PgError, Result, code};
use crate::query::Ctx;
use crate::routines::Routine;

/// COLLECTION is the position of the operators in a root value's root object collections.
pub const COLLECTION: usize = 8;

/// UserOperator is an operator of the operator collection, with the routine that computes it.
pub struct UserOperator {
    pub schema: String,
    pub name: String,
    pub left: u32,
    pub right: u32,
    pub stored: objects::Operator,
    pub routine: Arc<Routine>,
}

/// OperatorCache is the operators of a root value, with the addresses of the operator and function collections.
pub type OperatorCache = ((Option<store::Hash>, Option<store::Hash>), Arc<Vec<Arc<UserOperator>>>);

impl Ctx<'_> {
    /// user_operators returns the operators of the working root, reusing them while their collections are unchanged.
    pub fn user_operators(&mut self) -> Result<Arc<Vec<Arc<UserOperator>>>> {
        let address = (self.txn.root.root_objects[COLLECTION], self.txn.root.root_objects[crate::routines::COLLECTION]);
        if let Some((cached, operators)) = &self.session.operators
            && *cached == address
        {
            return Ok(operators.clone());
        }
        let routines = self.routines()?;
        let mut operators = Vec::new();
        for (_, address) in self.txn.root.objects(self.db, COLLECTION)? {
            let stored = objects::Operator::deserialize(&prolly::read_blob(self.db, &address)?)?;
            let Some(routine) = routines.iter().find(|r| r.object.id == stored.function) else { continue };
            let mut segments = id::segments(&stored.id).into_iter();
            let (schema, name) = (segments.next().unwrap_or_default(), segments.next().unwrap_or_default());
            let mut operand =
                || segments.next().filter(|t| !t.is_empty()).map_or(0, |t| crate::usertypes::type_oid(t.as_bytes()));
            let (left, right) = (operand(), operand());
            operators.push(Arc::new(UserOperator { schema, name, left, right, stored, routine: routine.clone() }));
        }
        let operators = Arc::new(operators);
        self.session.operators = Some((address, operators.clone()));
        Ok(operators)
    }
}

/// is_builtin reports whether Postgres has a built-in operator of the name.
pub fn is_builtin(name: &str) -> bool {
    static NAMES: OnceLock<HashSet<String>> = OnceLock::new();
    NAMES
        .get_or_init(|| {
            crate::pgcatalog::reg::builtin_column("pg_operator", "oprname")
                .into_iter()
                .filter_map(|(_, name)| name.output())
                .collect()
        })
        .contains(name)
}

/// operator_id returns the ID Go gives an operator from its schema, name, and operand types, where a missing left
/// operand is an empty type ID.
fn operator_id(schema: &str, name: &str, left: &[u8], right: &[u8]) -> Vec<u8> {
    id::new(id::SECTION_OPERATOR, &[schema, name, &String::from_utf8_lossy(left), &String::from_utf8_lossy(right)])
}

/// operator_name returns the name of an operator that a CREATE OPERATOR option gives, written alone or as
/// `OPERATOR(schema.name)`.
fn operator_name(arg: Option<&Node>) -> String {
    match arg.and_then(|a| a.node.as_ref()) {
        Some(NodeEnum::List(list)) => {
            list.items.iter().filter_map(crate::expr::node_name).next_back().unwrap_or_default().to_string()
        }
        Some(NodeEnum::TypeName(name)) => {
            name.names.iter().filter_map(crate::expr::node_name).next_back().unwrap_or_default().to_string()
        }
        Some(NodeEnum::String(s)) => s.sval.clone(),
        _ => String::new(),
    }
}

/// is_set reports whether a CREATE OPERATOR flag option, written alone or with a boolean, is set.
fn is_set(arg: Option<&Node>) -> bool {
    match arg.and_then(|a| a.node.as_ref()) {
        None => true,
        Some(NodeEnum::Boolean(b)) => b.boolval,
        Some(NodeEnum::String(s)) => matches!(s.sval.to_lowercase().as_str(), "true" | "on" | "yes" | "1"),
        Some(NodeEnum::Integer(i)) => i.ival != 0,
        _ => true,
    }
}

/// invalid_definition returns Postgres' error for a CREATE OPERATOR that cannot work.
fn invalid_definition(message: &str) -> PgError {
    PgError::new(code::INVALID_FUNCTION_DEFINITION, message)
}

impl Ctx<'_> {
    /// operand_type resolves a CREATE OPERATOR or DROP OPERATOR operand type, which may be a table's row type.
    fn operand_type(&mut self, name: &pg_query::protobuf::TypeName) -> Result<ColumnType> {
        self.prepare_type(name)?;
        crate::expr::resolve_type_name(name)
    }

    /// create_operator runs CREATE OPERATOR, checking the operator as Postgres' DefineOperator does, and stores it as
    /// Go stores an operator, recording it as the commutator and negator of the operators it names that have none.
    pub fn create_operator(&mut self, define: &DefineStmt) -> Result<Outcome> {
        let names: Vec<&str> = define.defnames.iter().filter_map(crate::expr::node_name).collect();
        let (named_schema, symbol) = match names.as_slice() {
            [.., schema, symbol] => (schema.to_string(), symbol.to_string()),
            [symbol] => (String::new(), symbol.to_string()),
            [] => return Err(PgError::internal("an operator without a name")),
        };
        let schema = self.target_schema(&named_schema, -1)?;
        let (mut left, mut right, mut function) = (None, None, Vec::new());
        let (mut commutator, mut negator, mut hashes, mut merges) = (String::new(), String::new(), false, false);
        let (mut restrict, mut join) = (false, false);
        for item in &define.definition {
            let Some(NodeEnum::DefElem(def)) = item.node.as_ref() else { continue };
            let arg = def.arg.as_deref();
            match def.defname.to_lowercase().as_str() {
                "leftarg" | "rightarg" => {
                    let Some(NodeEnum::TypeName(type_name)) = arg.and_then(|a| a.node.as_ref()) else { continue };
                    let ty = self.operand_type(type_name)?;
                    if def.defname.eq_ignore_ascii_case("leftarg") { left = Some(ty) } else { right = Some(ty) }
                }
                "function" | "procedure" => {
                    function = match arg.and_then(|a| a.node.as_ref()) {
                        Some(NodeEnum::TypeName(name)) => {
                            name.names.iter().filter_map(crate::expr::node_name).map(str::to_string).collect()
                        }
                        _ => Vec::new(),
                    }
                }
                "commutator" => commutator = operator_name(arg),
                "negator" => negator = operator_name(arg),
                "hashes" => hashes = is_set(arg),
                "merges" => merges = is_set(arg),
                "restrict" => restrict = true,
                "join" => join = true,
                other => self.session.notice(PgError {
                    severity: "WARNING",
                    ..PgError::new(code::SYNTAX_ERROR, format!("operator attribute \"{other}\" not recognized"))
                }),
            }
        }
        if function.is_empty() {
            return Err(invalid_definition("operator function must be specified"));
        }
        let right = match (left, right) {
            (None, None) => return Err(invalid_definition("operator argument types must be specified")),
            (_, None) => {
                return Err(PgError {
                    detail: Some("Postfix operators are not supported.".into()),
                    ..invalid_definition("operator right argument type must be specified")
                });
            }
            (_, Some(right)) => right,
        };
        let inputs: Vec<ColumnType> = left.into_iter().chain([right]).collect();
        let routine = self.support_function(&function, &inputs, None)?;
        if left.is_none() {
            for (set, message) in [
                (!commutator.is_empty(), "only binary operators can have commutators"),
                (join, "only binary operators can have join selectivity"),
                (merges, "only binary operators can merge join"),
                (hashes, "only binary operators can hash"),
            ] {
                if set {
                    return Err(invalid_definition(message));
                }
            }
        }
        if routine.ret.oid != crate::oid::BOOL {
            for (set, message) in [
                (!negator.is_empty(), "only boolean operators can have negators"),
                (restrict, "only boolean operators can have restriction selectivity"),
                (join, "only boolean operators can have join selectivity"),
                (merges, "only boolean operators can merge join"),
                (hashes, "only boolean operators can hash"),
            ] {
                if set {
                    return Err(invalid_definition(message));
                }
            }
        }
        let left_id = left.map_or(Vec::new(), |t| crate::usertypes::type_id(t.oid));
        let right_id = crate::usertypes::type_id(right.oid);
        let id = operator_id(&schema, &symbol, &left_id, &right_id);
        let existing = self.txn.root.objects(self.db, COLLECTION)?;
        if existing.iter().any(|(key, _)| *key == id) {
            return Err(PgError::new(code::DUPLICATE_FUNCTION, format!("operator {symbol} already exists")));
        }
        let operator = objects::Operator {
            id: id.clone(),
            function: routine.object.id.clone(),
            return_type: crate::usertypes::type_id(routine.ret.oid),
            commutator: commutator.clone().into_bytes(),
            negator: negator.clone().into_bytes(),
            hashes,
            merges,
        };
        self.put_operator(&operator)?;
        let links = [
            (operator_id(&schema, &commutator, &right_id, &left_id), commutator.is_empty(), true),
            (operator_id(&schema, &negator, &left_id, &right_id), negator.is_empty(), false),
        ];
        for (linked, unnamed, is_commutator) in links {
            if unnamed || linked == id {
                continue;
            }
            let Some((_, address)) = existing.iter().find(|(key, _)| *key == linked) else { continue };
            let mut other = objects::Operator::deserialize(&prolly::read_blob(self.db, address)?)?;
            let field = if is_commutator { &mut other.commutator } else { &mut other.negator };
            if field.is_empty() {
                *field = symbol.clone().into_bytes();
                self.put_operator(&other)?;
            }
        }
        Ok(Outcome::command("CREATE OPERATOR"))
    }

    /// put_operator writes an operator into the working root.
    fn put_operator(&mut self, operator: &objects::Operator) -> Result<()> {
        let data = operator.serialize();
        let db = &mut *self.db;
        let mut sink = |_: store::Hash, bytes: &[u8]| {
            db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
        };
        let (address, _) =
            prolly::write_blob(&data, &mut sink)?.ok_or_else(|| PgError::internal("an empty operator"))?;
        self.txn.root.put_object(self.db, COLLECTION, &operator.id, Some(address))?;
        Ok(())
    }

    /// drop_operators runs DROP OPERATOR.
    pub fn drop_operators(&mut self, drop: &DropStmt) -> Result<Outcome> {
        for object in &drop.objects {
            let Some(NodeEnum::ObjectWithArgs(target)) = object.node.as_ref() else { continue };
            let names: Vec<&str> = target.objname.iter().filter_map(crate::expr::node_name).collect();
            let (schema, symbol) = match names.as_slice() {
                [.., schema, symbol] => (Some(schema.to_string()), symbol.to_string()),
                [symbol] => (None, symbol.to_string()),
                [] => continue,
            };
            let mut operands = Vec::new();
            for arg in &target.objargs {
                operands.push(match arg.node.as_ref() {
                    Some(NodeEnum::TypeName(type_name)) => self.operand_type(type_name)?.oid,
                    _ => 0,
                });
            }
            let (left, right) = match operands.as_slice() {
                [left, right] => (*left, *right),
                _ => (0, operands.first().copied().unwrap_or_default()),
            };
            let found = self
                .user_operators()?
                .iter()
                .find(|o| {
                    o.name == symbol
                        && o.left == left
                        && o.right == right
                        && match &schema {
                            Some(schema) => o.schema == *schema,
                            None => crate::usertypes::in_search_path(&o.schema),
                        }
                })
                .map(|o| o.stored.id.clone());
            match found {
                Some(key) => self.txn.root.put_object(self.db, COLLECTION, &key, None)?,
                None if drop.missing_ok => {
                    self.session.notice(PgError::notice("00000", format!("operator {symbol} does not exist, skipping")))
                }
                None => {
                    let shown = |t: u32| if t == 0 { "NONE".into() } else { crate::cast::type_display(t) };
                    return Err(PgError::new(
                        code::UNDEFINED_FUNCTION,
                        format!("operator does not exist: {} {symbol} {}", shown(left), shown(right)),
                    ));
                }
            }
        }
        Ok(Outcome::command("DROP OPERATOR"))
    }
}
