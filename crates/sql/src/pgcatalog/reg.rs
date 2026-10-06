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

//! The reg types, whose values are OIDs that print as the names of the catalog objects they identify.

use crate::catalog::builtin_type;
use crate::error::{PgError, Result, code};
use crate::oid as types;
use crate::pgcatalog::snapshot::{index_oid, namespace_oid, sequence_oid, table_oid, view_oid};
use crate::pgcatalog::{builtin, lookup};
use crate::query::Ctx;
use crate::types::{Reg, Value};

/// Relation is a relation that regclass can name: its schema, name, and OID.
struct Relation {
    schema: String,
    name: String,
    oid: u32,
}

/// builtin_column returns a column of a built-in catalog's rows by name, as pairs of OID and value.
fn builtin_column(catalog: &str, column: &str) -> Vec<(u32, Value)> {
    let Some(table) = lookup("pg_catalog", catalog) else { return Vec::new() };
    let (Some(oid), Some(i)) = (table.column("oid"), table.column(column)) else { return Vec::new() };
    builtin::rows(table)
        .into_iter()
        .filter_map(|row| match &row[oid] {
            Value::Oid(o) => Some((*o, row[i].clone())),
            _ => None,
        })
        .collect()
}

/// text_of returns the text of a value.
fn text_of(value: &Value) -> String {
    value.output().unwrap_or_default()
}

impl Ctx<'_> {
    /// reg_value converts a value to a reg type: text names the object, and an OID or an integer identifies it.
    pub fn reg_value(&mut self, value: Value, type_oid: u32) -> Result<Value> {
        let reg = match value {
            Value::Null => return Ok(Value::Null),
            Value::Text(text) => self.reg_from_name(text.trim(), type_oid)?,
            Value::Reg(reg) => self.reg_from_oid(reg.oid, type_oid)?,
            Value::Oid(o) => self.reg_from_oid(o, type_oid)?,
            Value::Int2(i) => self.reg_from_oid(i as u32, type_oid)?,
            Value::Int4(i) => self.reg_from_oid(i as u32, type_oid)?,
            Value::Int8(i) => self.reg_from_oid(i as u32, type_oid)?,
            other => {
                return Err(PgError::new(
                    code::CANNOT_COERCE,
                    format!(
                        "cannot cast type {} to {}",
                        crate::cast::type_display(crate::functions::value_type(&other)),
                        crate::cast::type_display(type_oid)
                    ),
                ));
            }
        };
        Ok(Value::Reg(Box::new(reg)))
    }

    /// reg_from_oid returns the reg value of an OID, which prints as the number when nothing has the OID.
    fn reg_from_oid(&mut self, oid: u32, type_oid: u32) -> Result<Reg> {
        let name = match type_oid {
            _ if oid == 0 => Some("-".to_string()),
            types::REGCLASS => self.relations()?.into_iter().find(|r| r.oid == oid).map(|r| self.visible_name(&r)),
            types::REGTYPE => builtin_type(oid).map(|_| crate::cast::type_display(oid).into_owned()),
            types::REGNAMESPACE => self.namespaces().into_iter().find(|(_, o)| *o == oid).map(|(n, _)| n),
            types::REGROLE => self.roles().into_iter().find(|(_, o)| *o == oid).map(|(n, _)| n),
            types::REGPROC | types::REGPROCEDURE => {
                builtin_column("pg_proc", "proname").into_iter().find(|(o, _)| *o == oid).map(|(_, n)| text_of(&n))
            }
            _ => {
                builtin_column("pg_operator", "oprname").into_iter().find(|(o, _)| *o == oid).map(|(_, n)| text_of(&n))
            }
        };
        Ok(Reg { type_oid, oid, name: name.unwrap_or_else(|| oid.to_string()) })
    }

    /// reg_from_name returns the reg value of the object that text names, failing as Postgres does when there is
    /// none, where a number is an OID and `-` is 0.
    fn reg_from_name(&mut self, text: &str, type_oid: u32) -> Result<Reg> {
        if text == "-" {
            return Ok(Reg { type_oid, oid: 0, name: "-".into() });
        }
        if !text.is_empty() && text.bytes().all(|b| b.is_ascii_digit()) {
            let oid = text.parse().map_err(|_| crate::cast::out_of_range(types::OID, text))?;
            return self.reg_from_oid(oid, type_oid);
        }
        match type_oid {
            types::REGCLASS => {
                let names = crate::sequences::parse_qualified_name(text)?;
                let (schema, name) = match names.as_slice() {
                    [name] => (None, name.clone()),
                    [schema, name] | [_, schema, name] => (Some(schema.clone()), name.clone()),
                    _ => {
                        return Err(PgError::new(
                            code::SYNTAX_ERROR,
                            format!("improper relation name (too many dotted names): {text}"),
                        ));
                    }
                };
                let relations = self.relations()?;
                let schemas = match &schema {
                    Some(schema) => vec![schema.clone()],
                    None => self.effective_search_path(),
                };
                for s in &schemas {
                    if let Some(r) = relations.iter().find(|r| r.schema == *s && r.name == name) {
                        return Ok(Reg { type_oid, oid: r.oid, name: self.visible_name(r) });
                    }
                }
                Err(PgError::new(code::UNDEFINED_TABLE, format!("relation \"{text}\" does not exist")))
            }
            types::REGTYPE => {
                let ty = parse_type_name(text)?;
                Ok(Reg { type_oid, oid: ty, name: crate::cast::type_display(ty).into_owned() })
            }
            types::REGNAMESPACE => {
                let name = crate::sequences::parse_qualified_name(text)?.pop().unwrap_or_default();
                let (name, oid) = self.namespaces().into_iter().find(|(n, _)| *n == name).ok_or_else(|| {
                    PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{name}\" does not exist"))
                })?;
                Ok(Reg { type_oid, oid, name })
            }
            types::REGROLE => {
                let name = crate::sequences::parse_qualified_name(text)?.pop().unwrap_or_default();
                let (name, oid) =
                    self.roles().into_iter().find(|(n, _)| *n == name).ok_or_else(|| {
                        PgError::new(code::UNDEFINED_OBJECT, format!("role \"{name}\" does not exist"))
                    })?;
                Ok(Reg { type_oid, oid, name })
            }
            types::REGPROC | types::REGPROCEDURE => {
                let name = text.split('(').next().unwrap_or(text).trim();
                let name = crate::sequences::parse_qualified_name(name)?.pop().unwrap_or_default();
                let matches: Vec<u32> = builtin_column("pg_proc", "proname")
                    .into_iter()
                    .filter(|(_, n)| text_of(n) == name)
                    .map(|(o, _)| o)
                    .collect();
                match matches.as_slice() {
                    [oid] => Ok(Reg { type_oid, oid: *oid, name }),
                    [] => Err(PgError::new(code::UNDEFINED_FUNCTION, format!("function \"{text}\" does not exist"))),
                    _ => {
                        Err(PgError::new(code::AMBIGUOUS_FUNCTION, format!("more than one function named \"{text}\"")))
                    }
                }
            }
            _ => Err(PgError::unsupported(format!("reading values of type {}", crate::cast::type_display(type_oid)))),
        }
    }

    /// effective_search_path returns the schemas that unqualified names resolve in, with pg_catalog first unless the
    /// search path places it.
    pub fn effective_search_path(&self) -> Vec<String> {
        let mut path = self.session.search_path();
        if !path.iter().any(|s| s == "pg_catalog") {
            path.insert(0, "pg_catalog".into());
        }
        path
    }

    /// visible_name returns a relation's name, qualified with its schema unless an unqualified name finds it.
    fn visible_name(&self, relation: &Relation) -> String {
        let path = self.effective_search_path();
        if path.contains(&relation.schema) {
            crate::engine::quote_identifier(&relation.name)
        } else {
            format!(
                "{}.{}",
                crate::engine::quote_identifier(&relation.schema),
                crate::engine::quote_identifier(&relation.name)
            )
        }
    }

    /// relations returns every relation that regclass can name: the system catalogs, their indexes, and the user
    /// relations.
    fn relations(&mut self) -> Result<Vec<Relation>> {
        let mut out = Vec::new();
        if let Some(class) = lookup("pg_catalog", "pg_class") {
            let (Some(oid), Some(name), Some(namespace)) =
                (class.column("oid"), class.column("relname"), class.column("relnamespace"))
            else {
                return Ok(out);
            };
            for row in builtin::rows(class) {
                let schema = match &row[namespace] {
                    Value::Oid(11) => "pg_catalog",
                    Value::Oid(99) => "pg_toast",
                    _ => "information_schema",
                };
                if let Value::Oid(o) = row[oid] {
                    out.push(Relation { schema: schema.into(), name: text_of(&row[name]), oid: o });
                }
            }
        }
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            out.push(Relation {
                schema: table.schema.clone(),
                name: table.name.clone(),
                oid: table_oid(&table.schema, &table.name),
            });
            for index in crate::pgcatalog::rows::table_indexes(table) {
                out.push(Relation {
                    schema: table.schema.clone(),
                    oid: index_oid(&table.schema, &table.name, &index.name),
                    name: index.name,
                });
            }
        }
        for view in &snapshot.views {
            out.push(Relation {
                schema: view.schema.clone(),
                name: view.name.clone(),
                oid: view_oid(&view.schema, &view.name),
            });
        }
        for sequence in &snapshot.sequences {
            let (schema, name) = crate::sequences::schema_and_name(sequence);
            let oid = sequence_oid(&schema, &name);
            out.push(Relation { schema, name, oid });
        }
        Ok(out)
    }

    /// namespaces returns the name and OID of every schema.
    fn namespaces(&self) -> Vec<(String, u32)> {
        let mut out: Vec<(String, u32)> = ["pg_toast", "pg_catalog", "information_schema"]
            .iter()
            .map(|s| (s.to_string(), namespace_oid(s)))
            .collect();
        out.extend(self.schema_names().into_iter().filter(|s| s != "pg_catalog").map(|s| {
            let oid = namespace_oid(&s);
            (s, oid)
        }));
        out
    }

    /// roles returns the name and OID of every role.
    fn roles(&self) -> Vec<(String, u32)> {
        let mut out = vec![(self.session.superuser.clone(), crate::pgcatalog::rows::SUPERUSER)];
        out.extend(crate::pgcatalog::rows::PREDEFINED_ROLES.iter().map(|&(o, n)| (n.to_string(), o)));
        out
    }
}

/// parse_type_name reads a type name as Postgres' parseTypeString does, returning the type's OID.
fn parse_type_name(text: &str) -> Result<u32> {
    let undefined = || PgError::new(code::UNDEFINED_OBJECT, format!("type \"{text}\" does not exist"));
    let parsed = pg_query::parse(&format!("SELECT NULL::{text}"))
        .map_err(|_| PgError::new(code::SYNTAX_ERROR, format!("invalid type name \"{text}\"")))?;
    let mut type_name = None;
    for (node, ..) in parsed.protobuf.nodes() {
        if let pg_query::NodeRef::TypeCast(cast) = node {
            type_name = cast.type_name.clone();
        }
    }
    let type_name = type_name.ok_or_else(undefined)?;
    crate::expr::resolve_type_name(&type_name).map(|t| t.oid).map_err(|_| undefined())
}
