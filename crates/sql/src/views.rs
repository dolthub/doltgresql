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

//! Views, which Dolt stores as statements in each schema's dolt_schemas table.

use doltdb::table::Table;
use pg_query::protobuf::{DropBehavior, DropStmt, SelectStmt, ViewStmt};
use pg_query::{NodeEnum, NodeRef};
use prolly::val::encoding;
use serial::write::{ColumnFields, DEFAULT_TARGET_ROW_SIZE, SchemaFields, write_schema};

use crate::Outcome;
use crate::error::{PgError, Result, code};
use crate::expr::node_name;
use crate::plan::Planner;
use crate::query::{Ctx, scan};
use crate::types::Value;

/// DOLT_SCHEMAS is the table that holds a schema's views.
pub const DOLT_SCHEMAS: &str = "dolt_schemas";

/// SQL_MODE is the sql_mode that Doltgres records with each view.
const SQL_MODE: &str = "ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION";

/// dolt_schemas_schema returns the schema of the dolt_schemas table, as Dolt defines it.
fn dolt_schemas_schema() -> Vec<u8> {
    const TAG: u64 = 2251799813689256;
    let column = |i: u64, name: &'static [u8], sql_type: &'static [u8], encoding: u8, key: bool| ColumnFields {
        name,
        sql_type,
        default_value: b"",
        comment: b"",
        on_update: b"",
        tag: TAG + i,
        encoding,
        primary_key: key,
        auto_increment: false,
        nullable: !key,
        generated: false,
        is_virtual: false,
        adaptive_encoding: crate::storage::marks_adaptive(encoding),
        hidden: false,
        hidden_system: false,
    };
    write_schema(&SchemaFields {
        columns: vec![
            column(0, b"type", b"varchar(64) COLLATE utf8mb4_0900_ai_ci", encoding::STRING, true),
            column(1, b"name", b"varchar(64) COLLATE utf8mb4_0900_ai_ci", encoding::STRING, true),
            column(2, b"fragment", b"longtext", encoding::STRING_ADAPTIVE, false),
            column(3, b"extra", b"json", encoding::JSON_ADAPTIVE, false),
            column(4, b"sql_mode", b"varchar(256) COLLATE utf8mb4_0900_ai_ci", encoding::STRING, false),
        ],
        keyless: false,
        key_columns: vec![0, 1],
        value_columns: vec![2, 3, 4],
        indexes: Vec::new(),
        checks: Vec::new(),
        collation: 309,
        comment: b"",
        target_row_size: DEFAULT_TARGET_ROW_SIZE,
    })
}

/// relations returns the tables a query reads, by name.
fn relations(select: &SelectStmt) -> Vec<String> {
    let mut names = Vec::new();
    for (node, ..) in NodeEnum::SelectStmt(Box::new(select.clone())).nodes() {
        if let NodeRef::RangeVar(relation) = node
            && !names.contains(&relation.relname)
        {
            names.push(relation.relname.clone());
        }
    }
    names
}

/// view_definition returns the query of a stored CREATE VIEW statement as text, as the catalogs show it.
pub fn view_definition(fragment: &str) -> Result<String> {
    let (select, _) = view_query(fragment)?;
    let text = pg_query::NodeRef::SelectStmt(&select).deparse().map_err(PgError::internal)?;
    Ok(format!(" {text};"))
}

/// view_query returns the query of a stored CREATE VIEW statement, with its column names.
pub fn view_query(fragment: &str) -> Result<(SelectStmt, Vec<String>)> {
    let parsed = pg_query::parse(fragment).map_err(PgError::internal)?;
    let statement = parsed.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node);
    let Some(NodeEnum::ViewStmt(view)) = statement else {
        return Err(PgError::internal(format!("a stored view that is not one: {fragment}")));
    };
    let Some(NodeEnum::SelectStmt(select)) = view.query.and_then(|q| q.node) else {
        return Err(PgError::internal(format!("a stored view without a query: {fragment}")));
    };
    Ok((*select, view.aliases.iter().filter_map(node_name).map(str::to_string).collect()))
}

impl Ctx<'_> {
    /// views returns the name and statement of each view in a schema.
    pub fn views(&mut self, schema: &str) -> Result<Vec<(String, String)>> {
        let Some(table) = self.txn.table(self.db, schema, DOLT_SCHEMAS)? else { return Ok(Vec::new()) };
        Ok(scan(self.db, &table)?
            .into_iter()
            .filter(|row| row[0] == Value::Text("view".into()))
            .filter_map(|row| match (&row[1], &row[2]) {
                (Value::Text(name), Value::Text(fragment)) => Some((name.clone(), fragment.clone())),
                _ => None,
            })
            .collect())
    }

    /// find_view returns the schema and statement of a view, searching the session's schemas for an unqualified name.
    pub fn find_view(&mut self, schema: &str, name: &str) -> Result<Option<(String, String)>> {
        let schemas = if schema.is_empty() { self.session.search_path() } else { vec![schema.to_string()] };
        for schema in schemas {
            if let Some((_, fragment)) = self.views(&schema)?.into_iter().find(|(n, _)| n == name) {
                return Ok(Some((schema, fragment)));
            }
        }
        Ok(None)
    }

    /// put_view stores a view's statement in its schema's dolt_schemas table, or removes the view without one.
    fn put_view(&mut self, schema: &str, name: &str, fragment: Option<&str>) -> Result<()> {
        let table = match self.txn.table(self.db, schema, DOLT_SCHEMAS)? {
            Some(table) => table,
            None => {
                let (address, _) = Table::create(self.db, dolt_schemas_schema())?;
                self.txn.root.put_table(self.db, schema, DOLT_SCHEMAS, Some(address))?;
                self.txn
                    .table(self.db, schema, DOLT_SCHEMAS)?
                    .ok_or_else(|| PgError::internal("dolt_schemas vanished"))?
            }
        };
        let existing: Vec<Vec<Value>> = scan(self.db, &table)?
            .into_iter()
            .filter(|row| row[0] == Value::Text("view".into()) && row[1] == Value::Text(name.into()))
            .collect();
        crate::dml::delete_rows(self, &table, &existing)?;
        if let Some(fragment) = fragment {
            let table = self
                .txn
                .table(self.db, schema, DOLT_SCHEMAS)?
                .ok_or_else(|| PgError::internal("dolt_schemas vanished"))?;
            let row = vec![
                Value::Text("view".into()),
                Value::Text(name.into()),
                Value::Text(fragment.into()),
                Value::Text("{\"CreatedAt\":0}".into()),
                Value::Text(SQL_MODE.into()),
            ];
            crate::dml::write_rows(self, &table, &[row])?;
        }
        Ok(())
    }

    /// create_view runs CREATE [OR REPLACE] VIEW, checking its query first.
    pub fn create_view(&mut self, stmt: &ViewStmt) -> Result<Outcome> {
        let relation = stmt.view.as_ref().ok_or_else(|| PgError::internal("CREATE VIEW without a name"))?;
        if relation.relpersistence == "t" {
            return Err(PgError::unsupported("temporary views"));
        }
        let schema = self.target_schema(&relation.schemaname, relation.location)?;
        let name = relation.relname.clone();
        let Some(NodeEnum::SelectStmt(select)) = stmt.query.as_deref().and_then(|q| q.node.as_ref()) else {
            return Err(PgError::unsupported("this view query"));
        };
        let query = Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
        let aliases: Vec<String> = stmt.aliases.iter().filter_map(node_name).map(str::to_string).collect();
        if aliases.len() > query.columns.len() {
            return Err(PgError::new(code::SYNTAX_ERROR, "CREATE VIEW specifies more column names than columns"));
        }
        let names: Vec<String> = query
            .columns
            .iter()
            .enumerate()
            .map(|(i, c)| aliases.get(i).cloned().unwrap_or_else(|| c.name.clone()))
            .collect();
        let existing = self.views(&schema)?.into_iter().find(|(n, _)| *n == name);
        match &existing {
            Some((_, fragment)) if stmt.replace => {
                let (old_select, old_aliases) = view_query(fragment)?;
                let old = Planner { ctx: self, outer: Vec::new() }.plan_query(&old_select)?;
                if old.columns.len() > names.len() {
                    return Err(PgError::new(code::INVALID_TABLE_DEFINITION, "cannot drop columns from view"));
                }
                for (i, column) in old.columns.iter().enumerate() {
                    let old_name = old_aliases.get(i).cloned().unwrap_or_else(|| column.name.clone());
                    if old_name != names[i] {
                        return Err(PgError::new(
                            code::INVALID_TABLE_DEFINITION,
                            format!("cannot change name of view column \"{old_name}\" to \"{}\"", names[i]),
                        ));
                    }
                    if old.types[i].oid != query.types[i].oid {
                        return Err(PgError::new(
                            code::INVALID_TABLE_DEFINITION,
                            format!(
                                "cannot change data type of view column \"{old_name}\" from {} to {}",
                                crate::cast::type_display(old.types[i].oid),
                                crate::cast::type_display(query.types[i].oid)
                            ),
                        ));
                    }
                }
            }
            _ => {
                if self.relation_names(&schema)?.contains(&name) {
                    return Err(PgError::new(code::DUPLICATE_TABLE, format!("relation \"{name}\" already exists")));
                }
            }
        }
        let fragment = NodeRef::ViewStmt(stmt).deparse().map_err(PgError::internal)?;
        self.put_view(&schema, &name, Some(&fragment))?;
        self.own(crate::auth::Object::Table(schema.clone(), name.clone()))?;
        Ok(Outcome::command("CREATE VIEW"))
    }

    /// rename_view renames a view, keeping its privileges.
    pub fn rename_view(&mut self, schema: &str, old: &str, new: &str, fragment: &str) -> Result<()> {
        self.require_owner(&crate::auth::Object::Table(schema.to_string(), old.to_string()))
            .map_err(|err| PgError { message: format!("must be owner of view {old}"), ..err })?;
        if self.relation_names(schema)?.iter().any(|n| n == new) {
            return Err(PgError::new(code::DUPLICATE_TABLE, format!("relation \"{new}\" already exists")));
        }
        let parsed = pg_query::parse(fragment).map_err(PgError::internal)?;
        let Some(NodeEnum::ViewStmt(mut view)) =
            parsed.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node)
        else {
            return Err(PgError::internal(format!("a stored view that is not one: {fragment}")));
        };
        if let Some(relation) = view.view.as_mut() {
            relation.relname = new.to_string();
        }
        let renamed = NodeRef::ViewStmt(&view).deparse().map_err(PgError::internal)?;
        self.put_view(schema, old, None)?;
        self.put_view(schema, new, Some(&renamed))?;
        let mut auth = self.auth()?;
        auth.rename_object(
            &crate::auth::Object::Table(schema.to_string(), old.to_string()),
            &crate::auth::Object::Table(schema.to_string(), new.to_string()),
        );
        auth.persist()
    }

    /// drop_views runs DROP VIEW.
    pub fn drop_views(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(NodeEnum::List(list)) = object.node.as_ref() else { continue };
            let parts: Vec<&str> = list.items.iter().filter_map(node_name).collect();
            let (schema, name) = match parts.as_slice() {
                [name] => (String::new(), name.to_string()),
                [.., schema, name] => (schema.to_string(), name.to_string()),
                [] => continue,
            };
            match self.find_view(&schema, &name)? {
                Some((schema, _)) => {
                    self.require_owner(&crate::auth::Object::Table(schema.clone(), name.clone()))
                        .map_err(|err| PgError { message: format!("must be owner of view {name}"), ..err })?;
                    doomed.push((schema, name))
                }
                None => {
                    let shown = parts.join(".");
                    let schemas = if schema.is_empty() { self.session.search_path() } else { vec![schema.clone()] };
                    for s in &schemas {
                        if self.txn.root.table(self.db, s, &name)?.is_some() {
                            return Err(PgError {
                                hint: Some("Use DROP TABLE to remove a table.".into()),
                                ..PgError::new(code::WRONG_OBJECT_TYPE, format!("\"{shown}\" is not a view"))
                            });
                        }
                    }
                    if !drop.missing_ok {
                        return Err(PgError::new(code::UNDEFINED_TABLE, format!("view \"{shown}\" does not exist")));
                    }
                    self.session.notice(PgError::notice("00000", format!("view \"{shown}\" does not exist, skipping")));
                }
            }
        }
        for (schema, name) in &doomed {
            self.drop_dependents(schema, name, "view", drop.behavior, &[])?;
        }
        for (schema, name) in doomed {
            self.put_view(&schema, &name, None)?;
        }
        Ok(Outcome::command("DROP VIEW"))
    }

    /// view_dependents returns the views of a schema that read a relation, which dropping it must cascade to.
    pub fn view_dependents(&mut self, schema: &str, relation: &str) -> Result<Vec<String>> {
        let mut dependents = Vec::new();
        for (name, fragment) in self.views(schema)? {
            if name != relation && relations(&view_query(&fragment)?.0).iter().any(|r| r == relation) {
                dependents.push(name);
            }
        }
        Ok(dependents)
    }

    /// drop_dependents fails as Postgres does when views or the given foreign keys depend on a relation being dropped
    /// without CASCADE, and drops the views, and the views that depend on them, with it otherwise.
    pub fn drop_dependents(
        &mut self,
        schema: &str,
        relation: &str,
        kind: &str,
        behavior: i32,
        constraints: &[(String, String)],
    ) -> Result<()> {
        let mut found: Vec<(String, String, &str)> = Vec::new();
        let mut pending = vec![(relation.to_string(), kind)];
        while let Some((name, kind)) = pending.pop() {
            for view in self.view_dependents(schema, &name)? {
                if !found.iter().any(|(v, ..)| *v == view) {
                    found.push((view.clone(), name.clone(), kind));
                    pending.insert(0, (view, "view"));
                }
            }
        }
        if found.is_empty() && constraints.is_empty() {
            return Ok(());
        }
        if DropBehavior::try_from(behavior) != Ok(DropBehavior::DropCascade) {
            let detail: Vec<String> = constraints
                .iter()
                .map(|(c, t)| format!("constraint {c} on table {t} depends on {kind} {relation}"))
                .chain(found.iter().map(|(v, on, kind)| format!("view {v} depends on {kind} {on}")))
                .collect();
            return Err(PgError {
                detail: Some(detail.join("\n")),
                hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                ..PgError::new(
                    code::DEPENDENT_OBJECTS_STILL_EXIST,
                    format!("cannot drop {kind} {relation} because other objects depend on it"),
                )
            });
        }
        let detail: Vec<String> = constraints
            .iter()
            .map(|(c, t)| format!("drop cascades to constraint {c} on table {t}"))
            .chain(found.iter().map(|(v, ..)| format!("drop cascades to view {v}")))
            .collect();
        self.notice_cascades(detail);
        for (view, ..) in found {
            self.put_view(schema, &view, None)?;
        }
        Ok(())
    }

    /// notice_cascades sends Postgres' notice for the objects a drop cascades to, one per line.
    pub(crate) fn notice_cascades(&mut self, lines: Vec<String>) {
        if lines.len() == 1 {
            self.session.notice(PgError::notice("00000", lines[0].clone()));
        } else {
            self.session.notice(PgError {
                detail: Some(lines.join("\n")),
                ..PgError::notice("00000", format!("drop cascades to {} other objects", lines.len()))
            });
        }
    }
}
