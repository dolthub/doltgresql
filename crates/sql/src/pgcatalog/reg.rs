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
pub struct Relation {
    schema: String,
    name: String,
    oid: u32,
}

/// builtin_column returns a column of a built-in catalog's rows by name, as pairs of OID and value.
pub(crate) fn builtin_column(catalog: &str, column: &str) -> Vec<(u32, Value)> {
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
            types::REGCLASS => self.relations()?.iter().find(|r| r.oid == oid).map(|r| self.visible_name(r)),
            types::REGTYPE => match builtin_type(oid).is_some() || crate::usertypes::get(oid).is_some() {
                true => Some(crate::cast::type_display(oid).into_owned()),
                false => self.snapshot()?.tables.iter().find_map(|t| {
                    let array = crate::catalog::oids::oid(&crate::catalog::id::new(
                        crate::catalog::id::SECTION_TYPE,
                        &[&t.schema, &format!("_{}", t.name)],
                    ));
                    if crate::pgcatalog::rows::row_type_oid(&t.schema, &t.name) == oid {
                        Some(t.name.clone())
                    } else {
                        (array == oid).then(|| format!("{}[]", t.name))
                    }
                }),
            },
            types::REGNAMESPACE => self.namespaces().into_iter().find(|(_, o)| *o == oid).map(|(n, _)| n),
            types::REGROLE => self.roles().into_iter().find(|(_, o)| *o == oid).map(|(n, _)| n),
            types::REGPROC | types::REGPROCEDURE => {
                match builtin_column("pg_proc", "proname").into_iter().find(|(o, _)| *o == oid) {
                    Some((_, name)) => Some(text_of(&name)),
                    None => self
                        .routines()?
                        .iter()
                        .find(|r| crate::pgcatalog::routines::routine_oid(r) == oid)
                        .map(|r| r.name.clone()),
                }
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
        self.reg_named(text, type_oid)
    }

    /// reg_named returns the reg value of the object that text names, reading digits as a name rather than an OID.
    fn reg_named(&mut self, text: &str, type_oid: u32) -> Result<Reg> {
        match type_oid {
            types::REGCLASS => {
                let names = crate::sequences::parse_qualified_name(text)?;
                if let [catalog, ..] = names.as_slice()
                    && names.len() == 3
                    && *catalog != self.session.database
                {
                    return Err(PgError::new(
                        code::FEATURE_NOT_SUPPORTED,
                        format!("cross-database references are not implemented: \"{}\"", names.join(".")),
                    ));
                }
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
                let shown = schema.map_or(name.clone(), |schema| format!("{schema}.{name}"));
                Err(PgError::new(code::UNDEFINED_TABLE, format!("relation \"{shown}\" does not exist")))
            }
            types::REGTYPE => {
                let ty = parse_type_name(text)?;
                Ok(Reg { type_oid, oid: ty, name: crate::cast::type_display(ty).into_owned() })
            }
            types::REGNAMESPACE => {
                let name = single_name(text)?;
                let (name, oid) = self.namespaces().into_iter().find(|(n, _)| *n == name).ok_or_else(|| {
                    PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{name}\" does not exist"))
                })?;
                Ok(Reg { type_oid, oid, name })
            }
            types::REGROLE => {
                let name = single_name(text)?;
                let (name, oid) =
                    self.roles().into_iter().find(|(n, _)| *n == name).ok_or_else(|| {
                        PgError::new(code::UNDEFINED_OBJECT, format!("role \"{name}\" does not exist"))
                    })?;
                Ok(Reg { type_oid, oid, name })
            }
            types::REGPROCEDURE => {
                let (names, args) = name_and_arg_types(text)?;
                let mut names = crate::sequences::parse_qualified_name(names)?;
                let name = names.pop().unwrap_or_default();
                let schemas = match names.pop() {
                    Some(schema) => vec![schema],
                    None => self.effective_search_path(),
                };
                let mut found = None;
                if schemas.iter().any(|s| s == "pg_catalog") {
                    let arg_types = builtin_column("pg_proc", "proargtypes");
                    found = builtin_column("pg_proc", "proname")
                        .into_iter()
                        .zip(arg_types)
                        .find(|((_, n), (_, a))| {
                            text_of(n) == name
                                && a.output().unwrap_or_default().split_whitespace().eq(args.iter().map(u32::to_string))
                        })
                        .map(|((o, _), _)| o);
                }
                if found.is_none() {
                    found = self
                        .routines()?
                        .iter()
                        .find(|r| {
                            r.name == name
                                && schemas.contains(&r.schema)
                                && r.inputs().map(|p| p.ty.oid).eq(args.iter().copied())
                        })
                        .map(|r| crate::pgcatalog::routines::routine_oid(r));
                }
                let Some(oid) = found else {
                    return Err(PgError::new(code::UNDEFINED_FUNCTION, format!("function \"{text}\" does not exist")));
                };
                let types: Vec<String> =
                    args.iter().map(|&a| crate::cast::format_type(a, None).unwrap_or_default()).collect();
                Ok(Reg { type_oid, oid, name: format!("{name}({})", types.join(",")) })
            }
            types::REGPROC => {
                let name = text.split('(').next().unwrap_or(text).trim();
                let mut names = crate::sequences::parse_qualified_name(name)?;
                let name = names.pop().unwrap_or_default();
                let schemas = match names.pop() {
                    Some(schema) => vec![schema],
                    None => self.effective_search_path(),
                };
                let mut matches: Vec<u32> = Vec::new();
                if schemas.iter().any(|s| s == "pg_catalog") {
                    matches.extend(
                        builtin_column("pg_proc", "proname")
                            .into_iter()
                            .filter(|(_, n)| text_of(n) == name)
                            .map(|(o, _)| o),
                    );
                }
                matches.extend(
                    self.routines()?
                        .iter()
                        .filter(|r| r.name == name && schemas.contains(&r.schema))
                        .map(|r| crate::pgcatalog::routines::routine_oid(r)),
                );
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

    /// to_reg returns the reg value of the object that text names, or NULL when there is none, as the to_regclass
    /// family of functions does.
    pub fn to_reg(&mut self, text: &str, type_oid: u32) -> Result<Value> {
        let missing = [
            code::UNDEFINED_TABLE,
            code::UNDEFINED_OBJECT,
            code::INVALID_SCHEMA_NAME,
            code::UNDEFINED_FUNCTION,
            code::AMBIGUOUS_FUNCTION,
        ];
        match self.reg_named(text.trim(), type_oid) {
            Ok(reg) => Ok(Value::Reg(Box::new(reg))),
            Err(err) if missing.contains(&err.code) => Ok(Value::Null),
            Err(err) => Err(err),
        }
    }

    /// is_visible reports whether the object of an OID in a catalog is visible in the search path, as the
    /// pg_*_is_visible functions do, returning NULL when no such object exists. The objects of catalogs other than
    /// pg_class, pg_type, and pg_proc all live in pg_catalog.
    pub fn is_visible(&mut self, oid: u32, catalog: &str) -> Result<Value> {
        let path = self.effective_search_path();
        let builtin = |catalog: &str, column: &str| builtin_column(catalog, column).iter().any(|(o, _)| *o == oid);
        let schema = match catalog {
            "pg_class" => self.relations()?.iter().find(|r| r.oid == oid).map(|r| r.schema.clone()),
            "pg_type" => match builtin_type(oid) {
                Some(_) => Some("pg_catalog".to_string()),
                None => self.user_types()?.get(&oid).map(|t| t.schema.clone()),
            },
            "pg_proc" if builtin("pg_proc", "proname") => Some("pg_catalog".to_string()),
            "pg_proc" => self
                .routines()?
                .iter()
                .find(|r| crate::catalog::oids::oid(&r.object.id) == oid)
                .map(|r| r.schema.clone()),
            other => lookup("pg_catalog", other)
                .and_then(|table| table.columns.get(1).map(|c| c.name))
                .filter(|column| builtin(other, column))
                .map(|_| "pg_catalog".to_string()),
        };
        Ok(schema.map_or(Value::Null, |s| Value::Bool(path.contains(&s))))
    }

    /// description returns the comment on an object of a catalog, from pg_description, or from pg_shdescription for a
    /// shared catalog, with a column number for a column's comment.
    pub fn description(&mut self, oid: u32, catalog: &str, column: i32) -> Result<Option<String>> {
        if catalog == "pg_class" {
            let snapshot = self.snapshot()?;
            if let Some(table) = snapshot.tables.iter().find(|t| table_oid(&t.schema, &t.name) == oid) {
                let comment = match usize::try_from(column) {
                    Ok(0) => Some(&table.comment),
                    Ok(i) => table.columns.get(i - 1).map(|c| &c.comment),
                    Err(_) => None,
                };
                return Ok(comment.filter(|c| !c.is_empty()).cloned());
            }
        }
        Ok(self.builtin_description(oid, catalog, column))
    }

    /// builtin_description returns the built-in comment on an object of a catalog, from pg_description, or from
    /// pg_shdescription for a shared catalog, with a column number for a column's comment.
    fn builtin_description(&self, oid: u32, catalog: &str, column: i32) -> Option<String> {
        let shared = catalog == "pg_database";
        let description = lookup("pg_catalog", if shared { "pg_shdescription" } else { "pg_description" })?;
        let class = lookup("pg_catalog", catalog)?.oid;
        let (objoid, classoid, text) =
            (description.column("objoid")?, description.column("classoid")?, description.column("description")?);
        let subid = description.column("objsubid");
        builtin::rows(description)
            .into_iter()
            .find(|row| {
                row[objoid] == Value::Oid(oid)
                    && row[classoid] == Value::Oid(class)
                    && subid.is_none_or(|i| row[i] == Value::Int4(column))
            })
            .map(|row| text_of(&row[text]))
    }

    /// relation_exists reports whether a relation has the OID.
    pub fn relation_exists(&mut self, oid: u32) -> Result<bool> {
        Ok(self.relations()?.iter().any(|r| r.oid == oid))
    }

    /// is_publishable reports whether the relation with the OID is a user table, which logical replication can
    /// publish, or None when no relation has the OID.
    pub fn is_publishable(&mut self, oid: u32) -> Result<Option<bool>> {
        if !self.relation_exists(oid)? {
            return Ok(None);
        }
        let snapshot = self.snapshot()?;
        Ok(Some(snapshot.tables.iter().any(|t| table_oid(&t.schema, &t.name) == oid)))
    }

    /// role_of_oid returns the name of the role of an OID.
    pub fn role_of_oid(&self, oid: u32) -> Option<String> {
        self.roles().into_iter().find(|(_, o)| *o == oid).map(|(n, _)| n)
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
    /// relations, reading them once for each root value that a statement sees.
    fn relations(&mut self) -> Result<std::sync::Arc<Vec<Relation>>> {
        let snapshot = self.snapshot()?;
        if let Some(relations) = self.catalog.as_ref().and_then(|c| c.relations.clone()) {
            return Ok(relations);
        }
        let mut out = Vec::new();
        if let Some(class) = lookup("pg_catalog", "pg_class") {
            let (Some(oid), Some(name), Some(namespace)) =
                (class.column("oid"), class.column("relname"), class.column("relnamespace"))
            else {
                return Ok(std::sync::Arc::new(out));
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
        for user_type in self.user_types()?.values() {
            if matches!(user_type.kind, crate::usertypes::Kind::Composite(_)) {
                out.push(Relation {
                    schema: user_type.schema.clone(),
                    name: user_type.name.clone(),
                    oid: table_oid(&user_type.schema, &user_type.name),
                });
            }
        }
        for sequence in &snapshot.sequences {
            let (schema, name) = crate::sequences::schema_and_name(sequence);
            let oid = sequence_oid(&schema, &name);
            out.push(Relation { schema, name, oid });
        }
        let relations = std::sync::Arc::new(out);
        if let Some(cache) = self.catalog.as_mut() {
            cache.relations = Some(relations.clone());
        }
        Ok(relations)
    }

    /// namespaces returns the name and OID of every schema.
    pub(crate) fn namespaces(&self) -> Vec<(String, u32)> {
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
        let superuser = self.session.superuser.clone();
        let mut out: Vec<(String, u32)> = self
            .auth()
            .map(|auth| {
                auth.roles
                    .values()
                    .filter(|r| r.name != crate::auth::PUBLIC)
                    .map(|r| (r.name.clone(), crate::pgcatalog::rows::role_oid(&r.name, &superuser)))
                    .collect()
            })
            .unwrap_or_default();
        out.extend(crate::pgcatalog::rows::PREDEFINED_ROLES.iter().map(|&(o, n)| (n.to_string(), o)));
        out
    }
}

/// parse_type_name reads a type name as Postgres' parseTypeString does, returning the type's OID.
fn parse_type_name(text: &str) -> Result<u32> {
    let invalid = || PgError::new(code::SYNTAX_ERROR, format!("invalid type name \"{text}\""));
    const PREFIX: &str = "SELECT NULL::";
    let statements = crate::parse::parse(&format!("{PREFIX}{text}")).map_err(|err| PgError {
        position: err.position.and_then(|p| p.checked_sub(PREFIX.len() as u32)).filter(|p| *p > 0),
        ..err
    })?;
    let Some(crate::parse::Statement::Postgres { node: pg_query::NodeEnum::SelectStmt(select), .. }) =
        statements.first()
    else {
        return Err(invalid());
    };
    let target = select.target_list.first().and_then(|t| t.node.as_ref());
    let Some(pg_query::NodeEnum::ResTarget(target)) = target else { return Err(invalid()) };
    if !target.name.is_empty() {
        return Err(PgError {
            position: text.to_lowercase().rfind(&target.name).map(|p| p as u32 + 1),
            ..PgError::new(code::SYNTAX_ERROR, format!("syntax error at or near \"{}\"", target.name))
        });
    }
    let Some(pg_query::NodeEnum::TypeCast(cast)) = target.val.as_ref().and_then(|v| v.node.as_ref()) else {
        return Err(invalid());
    };
    let type_name = cast.type_name.clone().ok_or_else(invalid)?;
    let names: Vec<String> = type_name.names.iter().filter_map(crate::expr::node_name).map(str::to_string).collect();
    crate::expr::resolve_type_name(&type_name)
        .map(|t| t.oid)
        .map_err(|_| PgError::new(code::UNDEFINED_OBJECT, format!("type \"{}\" does not exist", names.join("."))))
}

/// name_and_arg_types splits text such as `f(integer, text)` into the function's name and the OIDs of its argument
/// types, as Postgres' parseNameAndArgTypes does.
fn name_and_arg_types(text: &str) -> Result<(&str, Vec<u32>)> {
    let invalid = |message: &str| PgError::new(code::INVALID_TEXT_REPRESENTATION, message);
    let open = text.find('(').ok_or_else(|| invalid("expected a left parenthesis"))?;
    let args = text[open + 1..].trim_end().strip_suffix(')').ok_or_else(|| invalid("expected a right parenthesis"))?;
    let (mut types, mut start, mut depth, mut quoted) = (Vec::new(), 0, 0, false);
    for (i, c) in args.char_indices() {
        match c {
            '"' => quoted = !quoted,
            '(' if !quoted => depth += 1,
            ')' if !quoted => depth -= 1,
            ',' if !quoted && depth == 0 => {
                types.push(parse_type_name(args[start..i].trim())?);
                start = i + 1;
            }
            _ => {}
        }
    }
    if !args[start..].trim().is_empty() || !types.is_empty() {
        types.push(parse_type_name(args[start..].trim())?);
    }
    Ok((text[..open].trim(), types))
}

/// single_name reads the one name that regnamespace and regrole take, failing as Postgres does for a qualified name.
fn single_name(text: &str) -> Result<String> {
    let mut names = crate::sequences::parse_qualified_name(text)?;
    if names.len() != 1 {
        return Err(PgError::new(code::INVALID_NAME, "invalid name syntax"));
    }
    Ok(names.pop().unwrap_or_default())
}
