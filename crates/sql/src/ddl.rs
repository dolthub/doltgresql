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
    ConstrType, CreateSchemaStmt, CreateStmt, CreateTableAsStmt, DropBehavior, DropStmt, IndexStmt, ObjectType,
    ResTarget, SelectStmt, SortByDir, SortByNulls, TruncateStmt,
};
use pg_query::{Node, NodeEnum};
use store::Hash;

use crate::Outcome;
use crate::catalog::ColumnType;
use crate::catalog::table::{Check, ColumnDef, IndexDef, TableDef, schema_message};
use crate::error::{PgError, Result, code};
use crate::expr::{node_name, position, resolve_type_name};
use crate::plan::Planner;
use crate::query::Ctx;

/// constraint_type returns a constraint node's type.
pub(crate) fn constraint_type(constraint: &pg_query::protobuf::Constraint) -> ConstrType {
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

/// TableParts are what a table's definition elements add up to, before the checks and indexes are named.
#[derive(Default)]
pub(crate) struct TableParts {
    pub columns: Vec<ColumnDef>,
    pub primary_key: Vec<usize>,
    /// Each check constraint's name, empty when unnamed, and expression.
    pub checks: Vec<(String, Node)>,
    /// Each unique constraint's name, empty when unnamed, and columns.
    pub uniques: Vec<(String, Vec<usize>)>,
    /// Each serial or identity column with its sequence's data type and options.
    pub generated: Vec<(usize, &'static str, Vec<Node>)>,
}

impl TableParts {
    /// add_column adds a column definition with its column constraints.
    pub fn add_column(&mut self, table: &str, def: &pg_query::protobuf::ColumnDef) -> Result<()> {
        if self.columns.iter().any(|c| c.name == def.colname) {
            return Err(PgError::new(
                code::DUPLICATE_COLUMN,
                format!("column \"{}\" specified more than once", def.colname),
            ));
        }
        let index = self.columns.len();
        let type_name = def.type_name.as_ref().ok_or_else(|| PgError::internal("a column without a type"))?;
        let serial = serial_type(type_name);
        let ty = match serial {
            Some(data_type) => crate::catalog::resolve_type(&[data_type.to_string()], &[], false, None)?,
            None => resolve_type_name(type_name)?,
        };
        if let Some(data_type) = serial {
            self.generated.push((index, data_type, Vec::new()));
        }
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
                    if !self.primary_key.is_empty() {
                        return Err(multiple_primary_keys(table, constraint.location));
                    }
                    self.primary_key.push(index);
                }
                ConstrType::ConstrDefault => {
                    let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("an empty DEFAULT"))?;
                    column.default = expression_text(expr)?;
                }
                ConstrType::ConstrCheck => {
                    let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("an empty CHECK"))?;
                    self.checks.push((constraint.conname.clone(), expr.clone()));
                }
                ConstrType::ConstrUnique => self.uniques.push((constraint.conname.clone(), vec![index])),
                ConstrType::ConstrIdentity => {
                    let data_type = match ty.oid {
                        crate::oid::INT2 => "int2",
                        crate::oid::INT4 => "int4",
                        crate::oid::INT8 => "int8",
                        _ => {
                            return Err(PgError::new(
                                code::INVALID_PARAMETER_VALUE,
                                "identity column type must be smallint, integer, or bigint",
                            ));
                        }
                    };
                    if !column.default.is_empty() {
                        return Err(PgError {
                            position: position(constraint.location),
                            ..PgError::new(
                                code::SYNTAX_ERROR,
                                format!(
                                    "both default and identity specified for column \"{}\" of table \"{table}\"",
                                    def.colname
                                ),
                            )
                        });
                    }
                    self.generated.push((index, data_type, constraint.options.clone()));
                }
                other => return Err(PgError::unsupported(format!("the column constraint {other:?}"))),
            }
        }
        self.columns.push(column);
        Ok(())
    }

    /// key_columns returns the columns a constraint's key names.
    fn key_columns(&self, constraint: &pg_query::protobuf::Constraint) -> Result<Vec<usize>> {
        let mut keys = Vec::new();
        for key in &constraint.keys {
            let key = node_name(key).unwrap_or_default();
            keys.push(self.columns.iter().position(|c| c.name == key).ok_or_else(|| PgError {
                position: position(constraint.location),
                ..PgError::new(code::UNDEFINED_COLUMN, format!("column \"{key}\" named in key does not exist"))
            })?);
        }
        Ok(keys)
    }

    /// add_constraint adds a table constraint.
    pub fn add_constraint(&mut self, table: &str, constraint: &pg_query::protobuf::Constraint) -> Result<()> {
        match constraint_type(constraint) {
            ConstrType::ConstrPrimary => {
                if !self.primary_key.is_empty() {
                    return Err(multiple_primary_keys(table, constraint.location));
                }
                self.primary_key = self.key_columns(constraint)?;
            }
            ConstrType::ConstrCheck => {
                let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("an empty CHECK"))?;
                self.checks.push((constraint.conname.clone(), expr.clone()));
            }
            ConstrType::ConstrUnique => {
                let keys = self.key_columns(constraint)?;
                self.uniques.push((constraint.conname.clone(), keys));
            }
            other => return Err(PgError::unsupported(format!("the table constraint {other:?}"))),
        }
        Ok(())
    }
}

/// serial_type returns the integer type of a serial pseudo-type name, or None for any other type.
fn serial_type(type_name: &pg_query::protobuf::TypeName) -> Option<&'static str> {
    if !type_name.array_bounds.is_empty() {
        return None;
    }
    let names: Vec<&str> = type_name.names.iter().filter_map(node_name).collect();
    match names.as_slice() {
        ["smallserial" | "serial2"] | ["pg_catalog", "smallserial" | "serial2"] => Some("int2"),
        ["serial" | "serial4"] | ["pg_catalog", "serial" | "serial4"] => Some("int4"),
        ["bigserial" | "serial8"] | ["pg_catalog", "bigserial" | "serial8"] => Some("int8"),
        _ => None,
    }
}

/// check_column returns the column a check expression references when it references exactly one, which Postgres
/// names an unnamed check constraint after.
pub(crate) fn check_column(expr: &Node) -> Option<String> {
    let node = expr.node.as_ref()?;
    let mut names: Vec<String> = Vec::new();
    for (node, ..) in node.nodes() {
        if let pg_query::NodeRef::ColumnRef(column) = node
            && let Some(name) = column.fields.iter().filter_map(node_name).next_back()
            && !names.iter().any(|n| n == name)
        {
            names.push(name.to_string());
        }
    }
    if names.len() == 1 { names.pop() } else { None }
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
        if self.relation_names(&schema)?.iter().any(|n| n == name) {
            let message = format!("relation \"{name}\" already exists");
            if create.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE TABLE"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        let mut parts = TableParts::default();
        for element in &create.table_elts {
            match element.node.as_ref() {
                Some(NodeEnum::ColumnDef(def)) => parts.add_column(name, def)?,
                Some(NodeEnum::Constraint(constraint)) => parts.add_constraint(name, constraint)?,
                _ => return Err(PgError::unsupported("this table element")),
            }
        }
        let TableParts { mut columns, primary_key, checks: pending_checks, uniques, generated } = parts;
        for &i in &primary_key {
            columns[i].primary_key = true;
            columns[i].nullable = false;
        }
        let mut constraints = self.constraint_names(&schema)?;
        let mut checks: Vec<Check> = Vec::new();
        for (constraint, expr) in pending_checks {
            let check_name = if constraint.is_empty() {
                let column = check_column(&expr);
                choose_relation_name(name, column.as_deref().unwrap_or(""), "check", &constraints)
            } else if checks.iter().any(|c| c.name == constraint) {
                return Err(PgError::new(
                    code::DUPLICATE_OBJECT,
                    format!("check constraint \"{constraint}\" already exists"),
                ));
            } else {
                constraint
            };
            constraints.push(check_name.clone());
            checks.push(Check { name: check_name, expression: expression_text(&expr)? });
        }
        let mut taken = self.relation_names(&schema)?;
        taken.push(name.to_string());
        if !primary_key.is_empty() {
            taken.push(format!("{name}_pkey"));
        }
        for (column, data_type, options) in generated {
            let column_name = columns[column].name.clone();
            columns[column].default =
                self.create_owned_sequence(&schema, name, &column_name, data_type, &options, &mut taken)?;
            columns[column].nullable = false;
        }
        let mut indexes = Vec::new();
        for (constraint, keys) in uniques {
            let index_name = if constraint.is_empty() {
                let names: Vec<&str> = keys.iter().map(|&k| columns[k].name.as_str()).collect();
                choose_relation_name(name, &names.join("_"), "key", &taken)
            } else {
                constraint
            };
            taken.push(index_name.clone());
            indexes.push(new_index(index_name, keys, true));
        }
        self.write_new_table(&schema, name, columns, primary_key, checks, indexes)?;
        Ok(Outcome::command("CREATE TABLE"))
    }

    /// target_schema returns the schema a new object goes in: the named one, which must exist, or the first existing
    /// schema of the search path.
    pub(crate) fn target_schema(&self, named: &str, location: i32) -> Result<String> {
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
        indexes: Vec<IndexDef>,
    ) -> Result<()> {
        let mut tags = self.txn.all_tags(self.db)?;
        let mut kinds = Vec::new();
        for column in &mut columns {
            column.tag = auto_generate_tag(&tags, name, &kinds, &column.name, EXTENDED_KIND);
            tags.insert(column.tag);
            kinds.push(EXTENDED_KIND);
        }
        let value_columns: Vec<usize> = (0..columns.len()).filter(|i| !primary_key.contains(i)).collect();
        let message = schema_message(&columns, &primary_key, &value_columns, &checks, &indexes)?;
        let (mut address, mut table) = Table::create(self.db, message)?;
        if !indexes.is_empty() {
            let empty = Hash::of(&empty_rows());
            for index in &indexes {
                table.put_index(self.db, &index.name, Some(empty))?;
            }
            address = table.write(self.db)?;
        }
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
        self.write_new_table(&schema, &name, columns, Vec::new(), Vec::new(), Vec::new())?;
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
            ObjectType::ObjectIndex => self.drop_indexes(drop),
            ObjectType::ObjectSequence => self.drop_sequences(drop),
            ObjectType::ObjectView => self.drop_views(drop),
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
                    let shown = if schema.is_empty() { name.clone() } else { format!("{schema}.{name}") };
                    if self.find_view(&schema, &name)?.is_some() {
                        return Err(PgError {
                            hint: Some("Use DROP VIEW to remove a view.".into()),
                            ..PgError::new(code::WRONG_OBJECT_TYPE, format!("\"{shown}\" is not a table"))
                        });
                    }
                    if !drop.missing_ok {
                        return Err(PgError::new(code::UNDEFINED_TABLE, format!("table \"{shown}\" does not exist")));
                    }
                    self.session
                        .notice(PgError::notice("00000", format!("table \"{shown}\" does not exist, skipping")));
                }
            }
        }
        for (schema, name) in &doomed {
            self.drop_dependents(schema, name, "table", drop.behavior)?;
        }
        for (schema, name) in doomed {
            self.txn.root.put_table(self.db, &schema, &name, None)?;
            self.drop_owned_sequences(&schema, &name)?;
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
            let empty = Hash::of(&empty_rows());
            for index in &table.indexes {
                stored.put_index(self.db, &index.name, Some(empty))?;
            }
            let address = stored.write(self.db)?;
            self.txn.root.put_table(self.db, &table.schema, &table.name, Some(address))?;
        }
        Ok(Outcome::command("TRUNCATE TABLE"))
    }
}

impl Ctx<'_> {
    /// relation_names returns the names of the tables, indexes, sequences, and views in a schema, which new relations
    /// must avoid.
    pub(crate) fn relation_names(&mut self, schema: &str) -> Result<Vec<String>> {
        let mut names = Vec::new();
        let prefix = doltdb::root::table_key(schema, "");
        for (key, address) in self.txn.root.tables(self.db)? {
            let Some(name) = key.strip_prefix(prefix.as_slice()) else { continue };
            let name = String::from_utf8_lossy(name).into_owned();
            let table = TableDef::load(self.db, schema, &name, address)?;
            names.extend(table.indexes.iter().map(|i| i.name.clone()));
            if !table.key_columns.is_empty() {
                names.push(format!("{name}_pkey"));
            }
            names.push(name);
        }
        for sequence in crate::sequences::all(self.db, &self.txn.root)? {
            let (sequence_schema, name) = crate::sequences::schema_and_name(&sequence);
            if sequence_schema == schema {
                names.push(name);
            }
        }
        names.extend(self.views(schema)?.into_iter().map(|(name, _)| name));
        Ok(names)
    }

    /// constraint_names returns the names of the constraints of every table in a schema, which new constraints must
    /// avoid.
    pub(crate) fn constraint_names(&mut self, schema: &str) -> Result<Vec<String>> {
        let mut names = Vec::new();
        let prefix = doltdb::root::table_key(schema, "");
        for (key, address) in self.txn.root.tables(self.db)? {
            let Some(name) = key.strip_prefix(prefix.as_slice()) else { continue };
            let name = String::from_utf8_lossy(name).into_owned();
            let table = TableDef::load(self.db, schema, &name, address)?;
            names.extend(table.checks.iter().map(|c| c.name.clone()));
            names.extend(table.indexes.iter().filter(|i| i.unique).map(|i| i.name.clone()));
            if !table.key_columns.is_empty() {
                names.push(format!("{name}_pkey"));
            }
        }
        Ok(names)
    }

    /// create_index runs CREATE INDEX, building the index from the table's rows.
    pub fn create_index(&mut self, stmt: &IndexStmt) -> Result<Outcome> {
        let relation = stmt.relation.as_ref().ok_or_else(|| PgError::internal("CREATE INDEX without a table"))?;
        let table = self.resolve_table(relation)?;
        if !matches!(stmt.access_method.as_str(), "" | "btree" | "hash") {
            return Err(PgError::unsupported(format!("indexes using {}", stmt.access_method)));
        }
        if stmt.where_clause.is_some() {
            return Err(PgError::unsupported("partial indexes"));
        }
        if !stmt.index_including_params.is_empty() {
            return Err(PgError::unsupported("indexes with INCLUDE"));
        }
        let mut columns = Vec::new();
        let mut descending = Vec::new();
        let mut nulls_last = Vec::new();
        let mut op_classes = Vec::new();
        for param in &stmt.index_params {
            let Some(NodeEnum::IndexElem(elem)) = param.node.as_ref() else { continue };
            if elem.expr.is_some() {
                return Err(PgError::unsupported("indexes on expressions"));
            }
            let column = table.columns.iter().position(|c| c.name == elem.name).ok_or_else(|| {
                PgError::new(code::UNDEFINED_COLUMN, format!("column \"{}\" does not exist", elem.name))
            })?;
            let desc = SortByDir::try_from(elem.ordering) == Ok(SortByDir::SortbyDesc);
            let last = match SortByNulls::try_from(elem.nulls_ordering) {
                Ok(SortByNulls::SortbyNullsFirst) => false,
                Ok(SortByNulls::SortbyNullsLast) => true,
                _ => !desc,
            };
            columns.push(column);
            descending.push(desc);
            nulls_last.push(last);
            op_classes.push(elem.opclass.iter().filter_map(node_name).next_back().unwrap_or_default().to_string());
        }
        let taken = self.relation_names(&table.schema)?;
        let name = if stmt.idxname.is_empty() {
            let names: Vec<&str> = columns.iter().map(|&c| table.columns[c].name.as_str()).collect();
            choose_relation_name(&table.name, &names.join("_"), "idx", &taken)
        } else {
            stmt.idxname.clone()
        };
        if taken.contains(&name) {
            let message = format!("relation \"{name}\" already exists");
            if stmt.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE INDEX"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        let index = IndexDef { descending, nulls_last, op_classes, ..new_index(name, columns, stmt.unique) };
        let mut table = table;
        table.indexes.push(index.clone());
        let mut keys = Vec::new();
        for row in crate::query::scan(self.db, &table)? {
            let (primary, _) = table.encode_row(self.db, &row)?;
            keys.push((table.index_key(self.db, &index, &row, &primary)?, row));
        }
        keys.sort_by(|a, b| table.compare_index_keys(&index, &a.0, &b.0));
        if index.unique {
            let width = index.columns.len();
            for pair in keys.windows(2) {
                let null = index.columns.iter().any(|&c| pair[0].1[c].is_null());
                if !null
                    && table.compare_index_prefix(&index, width, &pair[0].0, &pair[1].0) == std::cmp::Ordering::Equal
                {
                    let names: Vec<&str> = index.columns.iter().map(|&c| table.columns[c].name.as_str()).collect();
                    let values: Vec<String> =
                        index.columns.iter().map(|&c| pair[0].1[c].output().unwrap_or_default()).collect();
                    return Err(PgError {
                        detail: Some(format!("Key ({})=({}) is duplicated.", names.join(", "), values.join(", "))),
                        objects: Some(Box::new(crate::error::ErrorObjects {
                            schema: Some(table.schema.clone()),
                            table: Some(table.name.clone()),
                            constraint: Some(index.name.clone()),
                            ..Default::default()
                        })),
                        ..PgError::new(
                            code::UNIQUE_VIOLATION,
                            format!("could not create unique index \"{}\"", index.name),
                        )
                    });
                }
            }
        }
        let empty = Hash::of(&empty_rows());
        let mut stored = table.table.clone();
        stored.put_index(self.db, &index.name, Some(empty))?;
        let edits = keys.into_iter().map(|(k, _)| (k, Some(prolly::val::build_tuple(&[])))).collect();
        let compare = |a: &[u8], b: &[u8]| table.compare_index_keys(&index, a, b);
        stored.edit_index(self.db, &index.name, empty, edits, &compare, &table.index_encodings(&index))?;
        stored.schema = self.db.write_value(table.schema_message()?)?;
        let address = stored.write(self.db)?;
        self.txn.root.put_table(self.db, &table.schema, &table.name, Some(address))?;
        Ok(Outcome::command("CREATE INDEX"))
    }

    /// drop_indexes runs DROP INDEX.
    fn drop_indexes(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(NodeEnum::List(list)) = object.node.as_ref() else { continue };
            let (schema, name) = object_names(&list.items);
            let schemas = if schema.is_empty() { self.session.search_path() } else { vec![schema.clone()] };
            let mut found = None;
            'search: for s in schemas {
                let prefix = doltdb::root::table_key(&s, "");
                for (key, address) in self.txn.root.tables(self.db)? {
                    let Some(table) = key.strip_prefix(prefix.as_slice()) else { continue };
                    let table = TableDef::load(self.db, &s, &String::from_utf8_lossy(table), address)?;
                    if table.indexes.iter().any(|i| i.name == name) {
                        found = Some(table);
                        break 'search;
                    }
                }
            }
            match found {
                Some(table) => doomed.push((table, name)),
                None => {
                    let shown = if schema.is_empty() { name } else { format!("{schema}.{name}") };
                    if !drop.missing_ok {
                        return Err(PgError::new(code::UNDEFINED_OBJECT, format!("index \"{shown}\" does not exist")));
                    }
                    self.session
                        .notice(PgError::notice("00000", format!("index \"{shown}\" does not exist, skipping")));
                }
            }
        }
        for (table, name) in doomed {
            let table = match self.txn.table(self.db, &table.schema, &table.name)? {
                Some(table) => table,
                None => continue,
            };
            let mut table = table;
            table.indexes.retain(|i| i.name != name);
            let mut stored = table.table.clone();
            stored.put_index(self.db, &name, None)?;
            stored.schema = self.db.write_value(table.schema_message()?)?;
            let address = stored.write(self.db)?;
            self.txn.root.put_table(self.db, &table.schema, &table.name, Some(address))?;
        }
        Ok(Outcome::command("DROP INDEX"))
    }
}

/// new_index returns an index of the columns, ascending with NULLs last, whose root the caller sets.
pub(crate) fn new_index(name: String, columns: Vec<usize>, unique: bool) -> IndexDef {
    let count = columns.len();
    IndexDef {
        name,
        columns,
        unique,
        descending: vec![false; count],
        nulls_last: vec![true; count],
        op_classes: vec![String::new(); count],
        comment: String::new(),
        predicate: String::new(),
        root: Hash::of(&empty_rows()),
    }
}

/// NAMEDATALEN_MAX is the longest identifier Postgres keeps, in bytes.
const NAMEDATALEN_MAX: usize = 63;

/// clip returns the longest prefix of a name within a byte length that ends on a character boundary.
fn clip(name: &str, len: usize) -> &str {
    let mut end = len.min(name.len());
    while !name.is_char_boundary(end) {
        end -= 1;
    }
    &name[..end]
}

/// make_object_name joins two names and a label with underscores, shortening the longer name until the result fits,
/// as Postgres' makeObjectName does.
fn make_object_name(name1: &str, name2: &str, label: &str) -> String {
    let overhead = label.len() + 1 + if name2.is_empty() { 0 } else { 1 };
    let available = NAMEDATALEN_MAX - overhead;
    let (mut len1, mut len2) = (name1.len(), name2.len());
    while len1 + len2 > available {
        if len1 > len2 {
            len1 -= 1;
        } else {
            len2 -= 1;
        }
    }
    let (part1, part2) = (clip(name1, len1), clip(name2, len2));
    if part2.is_empty() { format!("{part1}_{label}") } else { format!("{part1}_{part2}_{label}") }
}

/// choose_relation_name returns a name from two names and a label that no existing relation takes, adding a number
/// to the label when needed, as Postgres' ChooseRelationName does.
pub(crate) fn choose_relation_name(name1: &str, name2: &str, label: &str, taken: &[String]) -> String {
    let mut name = make_object_name(name1, name2, label);
    let mut pass = 0;
    while taken.contains(&name) {
        pass += 1;
        name = make_object_name(name1, name2, &format!("{label}{pass}"));
    }
    name
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
