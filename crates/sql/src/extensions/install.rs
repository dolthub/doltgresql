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

//! CREATE EXTENSION and DROP EXTENSION, which write and remove the objects of an extension as Go does.

use pg_query::NodeEnum;
use pg_query::protobuf::{CreateExtensionStmt, DropStmt};
use store::Hash;

use super::{Extension, Routine, get};
use crate::Outcome;
use crate::catalog::id::{self, SECTION_FUNCTION, SECTION_NAMESPACE, SECTION_TYPE};
use crate::error::{PgError, Result, code};
use crate::query::Ctx;

/// COLLECTION is the position of the extensions in a root value's root object collections.
pub const COLLECTION: usize = 4;

/// The positions of the casts, operators, and aggregates in a root value's root object collections.
const CASTS: usize = 7;
const OPERATORS: usize = 8;
const AGGREGATES: usize = 9;

/// SECTION_EXTENSION is the ID section of extensions.
const SECTION_EXTENSION: u8 = 10;

/// Objects names the objects of an installed extension in a schema.
struct Objects<'e> {
    extension: &'e Extension,
    schema: String,
}

impl Objects<'_> {
    /// type_id returns the ID of a type that the extension names, which is either its own or a built-in type.
    fn type_id(&self, name: &str) -> Result<Vec<u8>> {
        if self.extension.types.iter().any(|t| t.name == name) {
            return Ok(id::new(SECTION_TYPE, &[&self.schema, name]));
        }
        crate::catalog::builtin_type_named(name)
            .map(|t| t.definition.id.clone())
            .ok_or_else(|| PgError::internal(format!("the extension type {name} does not exist")))
    }

    /// routine returns the extension's routine with the symbol.
    fn routine(&self, symbol: &str) -> Result<&Routine> {
        self.extension.routines.iter().find(|r| r.symbol == symbol).ok_or_else(|| {
            PgError::internal(format!(
                "extension \"{}\" does not declare the function \"{symbol}\"",
                self.extension.name
            ))
        })
    }

    /// function_id returns the ID of a function in the schema with parameters of the named types.
    fn function_id(&self, name: &str, params: &[String]) -> Result<Vec<u8>> {
        let types = params.iter().map(|p| self.type_id(p)).collect::<Result<Vec<_>>>()?;
        let types: Vec<String> = types.iter().map(|t| String::from_utf8_lossy(t).into_owned()).collect();
        let mut segments = vec![self.schema.as_str(), name];
        segments.extend(types.iter().map(String::as_str));
        Ok(id::new(SECTION_FUNCTION, &segments))
    }

    /// routine_id returns the ID of the routine with the symbol, or an empty ID without a symbol.
    fn routine_id(&self, symbol: &str) -> Result<Vec<u8>> {
        if symbol.is_empty() {
            return Ok(Vec::new());
        }
        let routine = self.routine(symbol)?;
        let params: Vec<String> = routine.params.iter().map(|(_, t)| t.clone()).collect();
        self.function_id(&routine.name, &params)
    }

    /// functions returns the stored functions of the extension's routines.
    fn functions(&self) -> Result<Vec<objects::Function>> {
        let mut functions = Vec::with_capacity(self.extension.routines.len());
        for routine in &self.extension.routines {
            let mut all_params = Vec::with_capacity(routine.params.len());
            for (name, ty) in &routine.params {
                all_params.push(objects::Parameter {
                    name: name.as_bytes().to_vec(),
                    type_id: self.type_id(ty)?,
                    ..objects::Parameter::default()
                });
            }
            functions.push(objects::Function {
                id: self.routine_id(&routine.symbol)?,
                return_type: self.type_id(&routine.returns)?,
                all_params,
                is_non_deterministic: true,
                strict: routine.strict,
                extension_name: self.extension.name.as_bytes().to_vec(),
                extension_symbol: routine.symbol.as_bytes().to_vec(),
                ..objects::Function::default()
            });
        }
        Ok(functions)
    }

    /// operators returns the stored operators of the extension.
    fn operators(&self) -> Result<Vec<objects::Operator>> {
        let mut operators = Vec::with_capacity(self.extension.operators.len());
        for operator in &self.extension.operators {
            let (left, right) = (self.type_id(&operator.left)?, self.type_id(&operator.right)?);
            let id = id::new(
                id::SECTION_OPERATOR,
                &[&self.schema, operator.name, &String::from_utf8_lossy(&left), &String::from_utf8_lossy(&right)],
            );
            operators.push(objects::Operator {
                id,
                function: self.routine_id(&operator.routine)?,
                return_type: self.type_id(&self.routine(&operator.routine)?.returns)?,
                commutator: operator.commutator.as_bytes().to_vec(),
                negator: operator.negator.as_bytes().to_vec(),
                ..objects::Operator::default()
            });
        }
        Ok(operators)
    }

    /// casts returns the stored casts of the extension.
    fn casts(&self) -> Result<Vec<objects::Cast>> {
        let mut casts = Vec::with_capacity(self.extension.casts.len());
        for cast in &self.extension.casts {
            let (source, target) = (self.type_id(&cast.source)?, self.type_id(&cast.target)?);
            let function = self.routine_id(&cast.routine)?;
            casts.push(objects::Cast {
                id: id::new(id::SECTION_CAST, &[&String::from_utf8_lossy(&source), &String::from_utf8_lossy(&target)]),
                cast_type: cast.context,
                use_in_out: function.is_empty(),
                function,
            });
        }
        Ok(casts)
    }

    /// aggregates returns the stored aggregates of the extension.
    fn aggregates(&self) -> Result<Vec<objects::Aggregate>> {
        let mut aggregates = Vec::with_capacity(self.extension.aggregates.len());
        for aggregate in &self.extension.aggregates {
            aggregates.push(objects::Aggregate {
                id: self.function_id(aggregate.name, &aggregate.params)?,
                return_type: self.type_id(&aggregate.returns)?,
                s_func: self.routine_id(&aggregate.transition)?,
                s_type: self.type_id(&aggregate.state_type)?,
                final_func: self.routine_id(&aggregate.final_routine)?,
                combine_func: self.routine_id(&aggregate.combine)?,
                init_cond: aggregate.init_cond.unwrap_or_default().as_bytes().to_vec(),
                has_init_cond: aggregate.init_cond.is_some(),
            });
        }
        Ok(aggregates)
    }
}

impl Ctx<'_> {
    /// installed_extensions returns the extensions installed in the working root.
    pub fn installed_extensions(&mut self) -> Result<Vec<objects::Extension>> {
        let mut installed = Vec::new();
        for (_, address) in self.txn.root.objects(self.db, COLLECTION)? {
            let data = prolly::read_blob(self.db, &address)?;
            installed.push(objects::Extension::deserialize(&data)?);
        }
        Ok(installed)
    }

    /// put_object writes a root object to a collection of the working root, or removes it without data.
    fn put_object(&mut self, collection: usize, id: &[u8], data: Option<Vec<u8>>) -> Result<()> {
        let address = match data {
            Some(data) => {
                let mut sink = |_: Hash, bytes: &[u8]| {
                    self.db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
                };
                prolly::write_blob(&data, &mut sink)?.map(|(address, _)| address)
            }
            None => None,
        };
        self.txn.root.put_object(self.db, collection, id, address)?;
        Ok(())
    }

    /// create_extension runs CREATE EXTENSION, writing the objects that the extension's installation script creates.
    pub fn create_extension(&mut self, stmt: &CreateExtensionStmt) -> Result<Outcome> {
        let name = stmt.extname.as_str();
        let extension_id = id::new(SECTION_EXTENSION, &[name]);
        if name == "plpgsql" || self.installed_extensions()?.iter().any(|e| e.ext_name == extension_id) {
            if stmt.if_not_exists {
                self.session.notice(PgError::notice(
                    code::DUPLICATE_OBJECT,
                    format!("extension \"{name}\" already exists, skipping"),
                ));
                return Ok(Outcome::command("CREATE EXTENSION"));
            }
            return Err(PgError::new(code::DUPLICATE_OBJECT, format!("extension \"{name}\" already exists")));
        }
        let extension = get(name).ok_or_else(|| {
            PgError::new(code::FEATURE_NOT_SUPPORTED, format!("extension \"{name}\" is not available"))
        })?;
        let (mut schema, mut version) = (String::new(), None);
        for option in &stmt.options {
            let Some(NodeEnum::DefElem(def)) = option.node.as_ref() else { continue };
            let value = def.arg.as_deref().and_then(crate::expr::node_name).unwrap_or_default().to_string();
            match def.defname.as_str() {
                "schema" => schema = value,
                "new_version" => version = Some(value),
                _ => {}
            }
        }
        if let Some(version) = version.filter(|v| v != extension.control.default_version) {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                format!("extension \"{name}\" has no installation script nor update path for version \"{version}\""),
            ));
        }
        let schema = self.target_schema(&schema, -1)?;
        let objects = Objects { extension, schema: schema.clone() };
        for base in &extension.types {
            self.store_type(crate::usertypes::extension_type(&schema, base.name))?;
        }
        for function in objects.functions()? {
            crate::routines::store(self.db, &mut self.txn.root, &function, false)?;
        }
        for operator in objects.operators()? {
            self.put_object(OPERATORS, &operator.id.clone(), Some(operator.serialize()))?;
        }
        for cast in objects.casts()? {
            self.put_object(CASTS, &cast.id.clone(), Some(cast.serialize()))?;
        }
        for aggregate in objects.aggregates()? {
            self.put_object(AGGREGATES, &aggregate.id.clone(), Some(aggregate.serialize()))?;
        }
        let installed = objects::Extension {
            ext_name: extension_id.clone(),
            namespace: id::new(SECTION_NAMESPACE, &[&schema]),
            relocatable: extension.control.relocatable,
            version: extension.control.default_version.as_bytes().to_vec(),
        };
        self.put_object(COLLECTION, &extension_id, Some(installed.serialize()))?;
        Ok(Outcome::command("CREATE EXTENSION"))
    }

    /// drop_extensions runs DROP EXTENSION, removing the objects of each extension, which no table column may use.
    pub fn drop_extensions(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let installed = self.installed_extensions()?;
        let mut dropping = Vec::new();
        for object in &drop.objects {
            let name = crate::expr::node_name(object).unwrap_or_default();
            let extension_id = id::new(SECTION_EXTENSION, &[name]);
            match installed.iter().find(|e| e.ext_name == extension_id) {
                Some(found) => dropping.push((name.to_string(), found.clone())),
                None if drop.missing_ok => self
                    .session
                    .notice(PgError::notice("00000", format!("extension \"{name}\" does not exist, skipping"))),
                None => {
                    return Err(PgError::new(code::UNDEFINED_OBJECT, format!("extension \"{name}\" does not exist")));
                }
            }
        }
        for (name, installed) in dropping {
            let Some(extension) = get(&name) else { continue };
            let schema = id::segments(&installed.namespace).into_iter().next().unwrap_or_default();
            let snapshot = self.snapshot()?;
            let mut dependents = Vec::new();
            for base in &extension.types {
                for type_name in [format!("_{}", base.name), base.name.to_string()] {
                    let oid = crate::catalog::oids::oid(&id::new(SECTION_TYPE, &[&schema, &type_name]));
                    for table in &snapshot.tables {
                        for column in table.columns.iter().filter(|c| c.ty.oid == oid) {
                            let shown = crate::cast::type_display(oid);
                            dependents.push(format!(
                                "column {} of table {} depends on type {shown}",
                                column.name, table.name
                            ));
                        }
                    }
                }
            }
            if !dependents.is_empty() {
                return Err(PgError {
                    detail: Some(dependents.join("\n")),
                    hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                    ..PgError::new(
                        code::DEPENDENT_OBJECTS_STILL_EXIST,
                        format!("cannot drop extension {name} because other objects depend on it"),
                    )
                });
            }
            let objects = Objects { extension, schema: schema.clone() };
            for base in &extension.types {
                for type_name in [base.name.to_string(), format!("_{}", base.name)] {
                    let type_id = id::new(SECTION_TYPE, &[&schema, &type_name]);
                    self.put_object(crate::usertypes::COLLECTION, &type_id, None)?;
                }
            }
            for function in objects.functions()? {
                self.put_object(crate::routines::COLLECTION, &function.id, None)?;
            }
            for operator in objects.operators()? {
                self.put_object(OPERATORS, &operator.id, None)?;
            }
            for cast in objects.casts()? {
                self.put_object(CASTS, &cast.id, None)?;
            }
            for aggregate in objects.aggregates()? {
                self.put_object(AGGREGATES, &aggregate.id, None)?;
            }
            self.put_object(COLLECTION, &installed.ext_name, None)?;
        }
        Ok(Outcome::command("DROP EXTENSION"))
    }
}
