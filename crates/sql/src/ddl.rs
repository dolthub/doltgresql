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

//! Statements that define schema objects.

use doltdb::table::{Table, empty_rows};
use doltdb::tags::{EXTENDED_KIND, auto_generate_tag};
use pg_query::protobuf::{
    ConstrType, CreateSchemaStmt, CreateStmt, CreateTableAsStmt, DropBehavior, DropStmt, ObjectType, ResTarget,
    SelectStmt, TruncateStmt,
};
use pg_query::{Node, NodeEnum};

use crate::Outcome;
use crate::catalog::ColumnType;
use crate::catalog::table::{Check, ColumnDef, schema_message};
use crate::error::{PgError, Result, code};
use crate::expr::{node_name, position, resolve_type_name};
use crate::plan::Planner;
use crate::query::Ctx;

/// constraint_type returns a constraint node's type.
fn constraint_type(constraint: &pg_query::protobuf::Constraint) -> ConstrType {
    ConstrType::try_from(constraint.contype).unwrap_or(ConstrType::Undefined)
}

/// expression_text returns an expression as SQL text, which defaults and check constraints store, by deparsing it
/// inside a SELECT.
pub fn expression_text(expr: &Node) -> Result<String> {
    let select = SelectStmt {
        target_list: vec![Node {
            node: Some(NodeEnum::ResTarget(Box::new(ResTarget {
                val: Some(Box::new(expr.clone())),
                ..Default::default()
            }))),
        }],
        limit_option: pg_query::protobuf::LimitOption::Default as i32,
        op: pg_query::protobuf::SetOperation::SetopNone as i32,
        ..Default::default()
    };
    let text = pg_query::NodeRef::SelectStmt(&select).deparse().map_err(PgError::internal)?;
    Ok(text.strip_prefix("SELECT ").unwrap_or(&text).to_string())
}

/// unique_name returns the first of `base`, `base1`, `base2`, ... that no existing name takes.
fn unique_name(base: &str, existing: &[String]) -> String {
    if !existing.iter().any(|e| e == base) {
        return base.to_string();
    }
    (1..).map(|i| format!("{base}{i}")).find(|n| !existing.iter().any(|e| e == n)).unwrap_or_default()
}

/// object_names returns the schema and name a qualified object name list names, with the schema empty when it is
/// unqualified.
fn object_names(names: &[Node]) -> (String, String) {
    let parts: Vec<&str> = names.iter().filter_map(node_name).collect();
    match parts.as_slice() {
        [name] => (String::new(), name.to_string()),
        [schema, name] => (schema.to_string(), name.to_string()),
        [_, schema, name] => (schema.to_string(), name.to_string()),
        _ => (String::new(), String::new()),
    }
}

impl Ctx<'_> {
    /// create_table runs CREATE TABLE.
    pub fn create_table(&mut self, create: &CreateStmt) -> Result<Outcome> {
        let relation = create.relation.as_ref().ok_or_else(|| PgError::internal("CREATE TABLE without a name"))?;
        let schema = self.target_schema(&relation.schemaname, relation.location)?;
        let name = relation.relname.as_str();
        if self.txn.root.table(self.db, &schema, name)?.is_some() {
            let message = format!("relation \"{name}\" already exists");
            if create.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE TABLE"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        let mut columns: Vec<ColumnDef> = Vec::new();
        let mut primary_key: Vec<usize> = Vec::new();
        let mut checks: Vec<Check> = Vec::new();
        let mut check_names: Vec<String> = Vec::new();
        for element in &create.table_elts {
            match element.node.as_ref() {
                Some(NodeEnum::ColumnDef(def)) => {
                    if columns.iter().any(|c| c.name == def.colname) {
                        return Err(PgError::new(
                            code::DUPLICATE_COLUMN,
                            format!("column \"{}\" specified more than once", def.colname),
                        ));
                    }
                    let type_name =
                        def.type_name.as_ref().ok_or_else(|| PgError::internal("a column without a type"))?;
                    let ty = resolve_type_name(type_name)?;
                    let mut column = ColumnDef {
                        name: def.colname.clone(),
                        ty,
                        tag: 0,
                        encoding: ty.encoding(),
                        nullable: true,
                        primary_key: false,
                        default: String::new(),
                    };
                    for constraint in &def.constraints {
                        let Some(NodeEnum::Constraint(constraint)) = constraint.node.as_ref() else { continue };
                        match constraint_type(constraint) {
                            ConstrType::ConstrNotnull => column.nullable = false,
                            ConstrType::ConstrNull => column.nullable = true,
                            ConstrType::ConstrPrimary => {
                                if !primary_key.is_empty() {
                                    return Err(multiple_primary_keys(name, constraint.location));
                                }
                                primary_key.push(columns.len());
                            }
                            ConstrType::ConstrDefault => {
                                let expr = constraint
                                    .raw_expr
                                    .as_deref()
                                    .ok_or_else(|| PgError::internal("an empty DEFAULT"))?;
                                column.default = expression_text(expr)?;
                            }
                            ConstrType::ConstrCheck => {
                                let expr = constraint
                                    .raw_expr
                                    .as_deref()
                                    .ok_or_else(|| PgError::internal("an empty CHECK"))?;
                                let check_name = if constraint.conname.is_empty() {
                                    unique_name(&format!("{name}_{}_check", def.colname), &check_names)
                                } else {
                                    constraint.conname.clone()
                                };
                                check_names.push(check_name.clone());
                                checks.push(Check { name: check_name, expression: expression_text(expr)? });
                            }
                            other => return Err(PgError::unsupported(format!("the column constraint {other:?}"))),
                        }
                    }
                    columns.push(column);
                }
                Some(NodeEnum::Constraint(constraint)) => match constraint_type(constraint) {
                    ConstrType::ConstrPrimary => {
                        if !primary_key.is_empty() {
                            return Err(multiple_primary_keys(name, constraint.location));
                        }
                        for key in &constraint.keys {
                            let key = node_name(key).unwrap_or_default();
                            let i = columns.iter().position(|c| c.name == key).ok_or_else(|| PgError {
                                position: position(constraint.location),
                                ..PgError::new(
                                    code::UNDEFINED_COLUMN,
                                    format!("column \"{key}\" named in key does not exist"),
                                )
                            })?;
                            primary_key.push(i);
                        }
                    }
                    ConstrType::ConstrCheck => {
                        let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("an empty CHECK"))?;
                        let check_name = if constraint.conname.is_empty() {
                            unique_name(&format!("{name}_check"), &check_names)
                        } else {
                            constraint.conname.clone()
                        };
                        check_names.push(check_name.clone());
                        checks.push(Check { name: check_name, expression: expression_text(expr)? });
                    }
                    other => return Err(PgError::unsupported(format!("the table constraint {other:?}"))),
                },
                _ => return Err(PgError::unsupported("this table element")),
            }
        }
        for &i in &primary_key {
            columns[i].primary_key = true;
            columns[i].nullable = false;
        }
        self.write_new_table(&schema, name, columns, primary_key, checks)?;
        Ok(Outcome::command("CREATE TABLE"))
    }

    /// target_schema returns the schema a new object goes in: the named one, which must exist, or the first existing
    /// schema of the search path.
    fn target_schema(&self, named: &str, location: i32) -> Result<String> {
        if named.is_empty() {
            return self.creation_schema();
        }
        if !self.txn.root.schemas.iter().any(|s| s == named.as_bytes()) {
            return Err(PgError {
                position: position(location),
                ..PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{named}\" does not exist"))
            });
        }
        Ok(named.to_string())
    }

    /// write_new_table chooses the columns' tags and writes a new empty table to the working root.
    fn write_new_table(
        &mut self,
        schema: &str,
        name: &str,
        mut columns: Vec<ColumnDef>,
        primary_key: Vec<usize>,
        checks: Vec<Check>,
    ) -> Result<()> {
        let mut tags = self.txn.all_tags(self.db)?;
        let mut kinds = Vec::new();
        for column in &mut columns {
            column.tag = auto_generate_tag(&tags, name, &kinds, &column.name, EXTENDED_KIND);
            tags.insert(column.tag);
            kinds.push(EXTENDED_KIND);
        }
        let value_columns: Vec<usize> = (0..columns.len()).filter(|i| !primary_key.contains(i)).collect();
        let message = schema_message(&columns, &primary_key, &value_columns, &checks)?;
        let (address, _) = Table::create(self.db, message)?;
        self.txn.root.put_table(self.db, schema, name, Some(address))?;
        Ok(())
    }

    /// create_table_as runs CREATE TABLE AS, which makes a keyless table of the query's columns and inserts its rows.
    pub fn create_table_as(&mut self, create: &CreateTableAsStmt) -> Result<Outcome> {
        let into = create.into.as_ref().ok_or_else(|| PgError::internal("CREATE TABLE AS without a target"))?;
        let relation = into.rel.as_ref().ok_or_else(|| PgError::internal("CREATE TABLE AS without a name"))?;
        let schema = self.target_schema(&relation.schemaname, relation.location)?;
        let name = relation.relname.clone();
        if self.txn.root.table(self.db, &schema, &name)?.is_some() {
            let message = format!("relation \"{name}\" already exists");
            if create.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE TABLE AS"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        let Some(NodeEnum::SelectStmt(select)) = create.query.as_deref().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::unsupported("CREATE TABLE AS without a SELECT"));
        };
        let query = Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
        let renames: Vec<&str> = into.col_names.iter().filter_map(node_name).collect();
        let columns: Vec<ColumnDef> = query
            .columns
            .iter()
            .zip(&query.types)
            .enumerate()
            .map(|(i, (c, ty))| {
                let ty = if ty.oid == crate::oid::UNKNOWN {
                    ColumnType { oid: crate::oid::TEXT, modifier: -1 }
                } else {
                    *ty
                };
                ColumnDef {
                    name: renames.get(i).map_or(c.name.clone(), |r| r.to_string()),
                    ty,
                    tag: 0,
                    encoding: ty.encoding(),
                    nullable: true,
                    primary_key: false,
                    default: String::new(),
                }
            })
            .collect();
        self.write_new_table(&schema, &name, columns, Vec::new(), Vec::new())?;
        let rows = if into.skip_data { Vec::new() } else { query.plan.run(self)? };
        let count = rows.len();
        let table =
            self.txn.table(self.db, &schema, &name)?.ok_or_else(|| PgError::internal("a new table vanished"))?;
        crate::dml::insert_rows(self, &table, rows)?;
        Ok(Outcome::command(format!("SELECT {count}")))
    }

    /// create_schema runs CREATE SCHEMA.
    pub fn create_schema(&mut self, create: &CreateSchemaStmt) -> Result<Outcome> {
        let name = if create.schemaname.is_empty() {
            create.authrole.as_ref().map(|r| r.rolename.clone()).unwrap_or_default()
        } else {
            create.schemaname.clone()
        };
        if self.txn.root.schemas.iter().any(|s| s == name.as_bytes()) {
            if create.if_not_exists {
                self.session.notice(PgError::notice(
                    code::DUPLICATE_SCHEMA,
                    format!("schema \"{name}\" already exists, skipping"),
                ));
                return Ok(Outcome::command("CREATE SCHEMA"));
            }
            return Err(PgError::new(code::DUPLICATE_SCHEMA, format!("schema \"{name}\" already exists")));
        }
        if !create.schema_elts.is_empty() {
            return Err(PgError::unsupported("CREATE SCHEMA with elements"));
        }
        self.txn.root.schemas.push(name.into_bytes());
        self.txn.root.schemas.sort();
        Ok(Outcome::command("CREATE SCHEMA"))
    }

    /// drop runs DROP TABLE and DROP SCHEMA.
    pub fn drop(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let kind = ObjectType::try_from(drop.remove_type).unwrap_or(ObjectType::Undefined);
        let cascade = DropBehavior::try_from(drop.behavior) == Ok(DropBehavior::DropCascade);
        match kind {
            ObjectType::ObjectTable => self.drop_tables(drop),
            ObjectType::ObjectSchema => self.drop_schemas(drop, cascade),
            other => Err(PgError::unsupported(format!("DROP {other:?}"))),
        }
    }

    /// drop_tables runs DROP TABLE, resolving every table first so that a missing one drops none.
    fn drop_tables(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(NodeEnum::List(list)) = object.node.as_ref() else { continue };
            let (schema, name) = object_names(&list.items);
            let schemas = if schema.is_empty() { self.session.search_path() } else { vec![schema.clone()] };
            let mut found = None;
            for s in schemas {
                if self.txn.root.table(self.db, &s, &name)?.is_some() {
                    found = Some(s);
                    break;
                }
            }
            match found {
                Some(s) => doomed.push((s, name)),
                None => {
                    let shown = if schema.is_empty() { name } else { format!("{schema}.{name}") };
                    if !drop.missing_ok {
                        return Err(PgError::new(code::UNDEFINED_TABLE, format!("table \"{shown}\" does not exist")));
                    }
                    self.session
                        .notice(PgError::notice("00000", format!("table \"{shown}\" does not exist, skipping")));
                }
            }
        }
        for (schema, name) in doomed {
            self.txn.root.put_table(self.db, &schema, &name, None)?;
        }
        Ok(Outcome::command("DROP TABLE"))
    }

    /// drop_schemas runs DROP SCHEMA, which fails for a schema with tables unless it cascades to them.
    fn drop_schemas(&mut self, drop: &DropStmt, cascade: bool) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(name) = node_name(object) else { continue };
            if !self.txn.root.schemas.iter().any(|s| s == name.as_bytes()) {
                if !drop.missing_ok {
                    return Err(PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{name}\" does not exist")));
                }
                self.session.notice(PgError::notice("00000", format!("schema \"{name}\" does not exist, skipping")));
                continue;
            }
            let prefix = doltdb::root::table_key(name, "");
            let tables: Vec<String> = self
                .txn
                .root
                .tables(self.db)?
                .into_iter()
                .filter_map(|(key, _)| {
                    key.strip_prefix(prefix.as_slice()).map(|n| String::from_utf8_lossy(n).into_owned())
                })
                .collect();
            if !tables.is_empty() && !cascade {
                let detail =
                    tables.iter().map(|t| format!("table {name}.{t} depends on schema {name}")).collect::<Vec<_>>();
                return Err(PgError {
                    detail: Some(detail.join("\n")),
                    hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                    ..PgError::new(
                        code::DEPENDENT_OBJECTS_STILL_EXIST,
                        format!("cannot drop schema {name} because other objects depend on it"),
                    )
                });
            }
            doomed.push((name.to_string(), tables));
        }
        for (name, tables) in doomed {
            match tables.len() {
                0 => {}
                1 => self
                    .session
                    .notice(PgError::notice("00000", format!("drop cascades to table {name}.{}", tables[0]))),
                n => {
                    let detail =
                        tables.iter().map(|t| format!("drop cascades to table {name}.{t}")).collect::<Vec<_>>();
                    self.session.notice(PgError {
                        detail: Some(detail.join("\n")),
                        ..PgError::notice("00000", format!("drop cascades to {n} other objects"))
                    });
                }
            }
            for table in tables {
                self.txn.root.put_table(self.db, &name, &table, None)?;
            }
            self.txn.root.schemas.retain(|s| s != name.as_bytes());
        }
        Ok(Outcome::command("DROP SCHEMA"))
    }

    /// truncate runs TRUNCATE, which empties the tables.
    pub fn truncate(&mut self, truncate: &TruncateStmt) -> Result<Outcome> {
        let mut tables = Vec::new();
        for relation in &truncate.relations {
            let Some(NodeEnum::RangeVar(relation)) = relation.node.as_ref() else { continue };
            tables.push(self.resolve_table(relation)?);
        }
        for table in tables {
            let mut stored = table.table.clone();
            stored.primary_index = empty_rows();
            let address = stored.write(self.db)?;
            self.txn.root.put_table(self.db, &table.schema, &table.name, Some(address))?;
        }
        Ok(Outcome::command("TRUNCATE TABLE"))
    }
}

/// multiple_primary_keys returns Postgres' error for a second primary key.
fn multiple_primary_keys(table: &str, location: i32) -> PgError {
    PgError {
        position: position(location),
        ..PgError::new(
            code::INVALID_TABLE_DEFINITION,
            format!("multiple primary keys for table \"{table}\" are not allowed"),
        )
    }
}
