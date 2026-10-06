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

//! CREATE INDEX with the access methods that extensions provide, which build Dolt's vector indexes.

use pg_query::NodeEnum;
use pg_query::protobuf::IndexStmt;

use super::{AccessMethod, Extension, OperatorClass};
use crate::Outcome;
use crate::catalog::table::{IndexDef, TableDef};
use crate::error::{PgError, Result, code};
use crate::expr::{Binder, node_name};
use crate::query::Ctx;

/// no_default_class returns Postgres' error for a type without a default operator class for an access method.
pub fn no_default_class(type_oid: u32, method: &str) -> PgError {
    PgError {
        hint: Some(
            "You must specify an operator class for the index or define a default operator class for the data type."
                .into(),
        ),
        ..PgError::new(
            code::UNDEFINED_OBJECT,
            format!(
                "data type {} has no default operator class for access method \"{method}\"",
                crate::cast::type_display(type_oid)
            ),
        )
    }
}

/// storage_params checks the storage parameters of an index against those its access method takes, returning each
/// parameter's value with the defaults filled in.
fn storage_params(method: &AccessMethod, options: &[pg_query::Node]) -> Result<Vec<(&'static str, i64)>> {
    let mut values: Vec<(&'static str, i64)> = method.params.iter().map(|p| (p.0, p.3)).collect();
    for option in options {
        let Some(NodeEnum::DefElem(def)) = option.node.as_ref() else { continue };
        let name = def.defname.to_ascii_lowercase();
        let Some(&(param, min, max, _)) = method.params.iter().find(|p| p.0 == name) else {
            return Err(PgError::new(code::INVALID_PARAMETER_VALUE, format!("unrecognized parameter \"{name}\"")));
        };
        let text = match def.arg.as_deref().and_then(|a| a.node.as_ref()) {
            Some(NodeEnum::Integer(i)) => i.ival.to_string(),
            Some(NodeEnum::Float(f)) => f.fval.clone(),
            Some(NodeEnum::String(s)) => s.sval.clone(),
            Some(NodeEnum::Boolean(b)) => b.boolval.to_string(),
            _ => String::new(),
        };
        let value = match text.trim().parse::<i64>() {
            Ok(v) => v,
            Err(_) => match text.trim().parse::<f64>() {
                Ok(f) if f.is_finite() => f.round_ties_even() as i64,
                _ => {
                    return Err(PgError::new(
                        code::INVALID_PARAMETER_VALUE,
                        format!("invalid value for integer option \"{param}\": {text}"),
                    ));
                }
            },
        };
        if value < min || value > max {
            return Err(PgError {
                detail: Some(format!("Valid values are between \"{min}\" and \"{max}\".")),
                ..PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!("value {value} out of bounds for option \"{param}\""),
                )
            });
        }
        if let Some(entry) = values.iter_mut().find(|(p, _)| *p == param) {
            entry.1 = value;
        }
    }
    Ok(values)
}

impl Ctx<'_> {
    /// access_method returns an installed extension's access method of the name.
    pub fn access_method(&mut self, name: &str) -> Result<Option<(&'static Extension, &'static AccessMethod)>> {
        for installed in self.installed_extensions()? {
            let extension_name = crate::catalog::id::segments(&installed.ext_name).into_iter().next();
            let Some(extension) = extension_name.and_then(|n| super::get(&n)) else { continue };
            if let Some(method) = extension.access_methods.iter().find(|m| m.name == name) {
                return Ok(Some((extension, method)));
            }
        }
        Ok(None)
    }

    /// create_vector_index runs CREATE INDEX with an extension's access method, checking it as pgvector does and
    /// building a Dolt vector index.
    pub fn create_vector_index(
        &mut self,
        stmt: &IndexStmt,
        table: TableDef,
        extension: &'static Extension,
        method: &'static AccessMethod,
    ) -> Result<Outcome> {
        let unsupported = |what: &str| {
            PgError::new(
                code::FEATURE_NOT_SUPPORTED,
                format!("access method \"{}\" does not support {what}", method.name),
            )
        };
        if stmt.unique {
            return Err(unsupported("unique indexes"));
        }
        if !stmt.index_including_params.is_empty() {
            return Err(unsupported("included columns"));
        }
        let [param] = stmt.index_params.as_slice() else { return Err(unsupported("multicolumn indexes")) };
        let params = storage_params(method, &stmt.options)?;
        let Some(NodeEnum::IndexElem(elem)) = param.node.as_ref() else {
            return Err(PgError::internal("an index without columns"));
        };
        let (column, ty) = match &elem.expr {
            Some(expr) => {
                let scope = crate::dml::table_scope(&table, None);
                (None, Binder::new(self, scope).bind(expr)?.1)
            }
            None => {
                let column = table.columns.iter().position(|c| c.name == elem.name).ok_or_else(|| {
                    PgError::new(code::UNDEFINED_COLUMN, format!("column \"{}\" does not exist", elem.name))
                })?;
                (Some(column), table.columns[column].ty)
            }
        };
        let type_name = crate::usertypes::get(ty.oid)
            .map(|t| t.name.clone())
            .unwrap_or_else(|| crate::catalog::builtin_type(ty.oid).map_or(String::new(), |t| t.name.to_string()));
        let class_name = elem.opclass.iter().filter_map(node_name).next_back();
        let class: &OperatorClass = match class_name {
            Some(name) => {
                let class = extension
                    .operator_classes
                    .iter()
                    .find(|c| c.name == name && c.access_methods.contains(&method.name))
                    .ok_or_else(|| {
                        PgError::new(
                            code::UNDEFINED_OBJECT,
                            format!("operator class \"{name}\" does not exist for access method \"{}\"", method.name),
                        )
                    })?;
                if class.type_name != type_name {
                    return Err(PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!(
                            "operator class \"{name}\" does not accept data type {}",
                            crate::cast::type_display(ty.oid)
                        ),
                    ));
                }
                class
            }
            None => extension
                .operator_classes
                .iter()
                .find(|c| c.type_name == type_name && c.default_for.contains(&method.name))
                .ok_or_else(|| no_default_class(ty.oid, method.name))?,
        };
        let Some(distance) = class.distance else {
            return Err(PgError::unsupported(format!("indexes with the operator class {}", class.name)));
        };
        if stmt.where_clause.is_some() {
            return Err(PgError::unsupported("partial vector indexes"));
        }
        if ty.modifier < 1 {
            return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "column does not have dimensions"));
        }
        if ty.modifier > class.max_dimensions {
            return Err(PgError::new(
                code::PROGRAM_LIMIT_EXCEEDED,
                format!("column cannot have more than {} dimensions for {} index", class.max_dimensions, method.name),
            ));
        }
        let Some(column) = column else { return Err(PgError::unsupported("expressions in vector indexes")) };
        let param = |name: &str| params.iter().find(|(p, _)| *p == name).map(|(_, v)| *v);
        if let (Some(m), Some(ef_construction)) = (param("m"), param("ef_construction"))
            && ef_construction < 2 * m
        {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                "ef_construction must be greater than or equal to 2 * m",
            ));
        }
        let name = self.index_name(stmt, &table, &[column])?;
        let Some(name) = name else { return Ok(Outcome::command("CREATE INDEX")) };
        if let Some(lists) = param("lists") {
            let rows = crate::query::scan(self.db, &table)?;
            if (rows.iter().filter(|r| !r[column].is_null()).count() as i64) < lists {
                self.session.notice(PgError {
                    detail: Some("This will cause low recall.".into()),
                    hint: Some("Drop the index until the table has more data.".into()),
                    ..PgError::notice("00000", "ivfflat index created with little data")
                });
            }
        }
        let index = IndexDef { vector: Some(distance), ..crate::ddl::new_index(name, vec![column], false) };
        self.build_index(table, index)?;
        Ok(Outcome::command("CREATE INDEX"))
    }
}

/// vector_rendering returns the access method and operator class that a vector index of a column renders with, which
/// Doltgres derives from the index's distance and the column's type, as Go does, since Dolt stores only the distance.
pub fn vector_rendering(distance: prolly::Distance, column_type: u32) -> Option<(&'static str, String)> {
    let type_name = crate::usertypes::get(column_type)?.name.clone();
    let class = super::all()
        .iter()
        .flat_map(|e| &e.operator_classes)
        .find(|c| c.type_name == type_name && c.distance == Some(distance))?;
    Some(("hnsw", class.name.clone()))
}
