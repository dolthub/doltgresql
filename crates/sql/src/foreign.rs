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

//! Foreign keys: Dolt's foreign key collection, defining them, and enforcing them at the end of each statement.

use doltdb::database::Database;
use doltdb::root::{Root, table_key};
use pg_query::protobuf::Constraint;
use serial::write::{ForeignKeyFields, write_foreign_keys};
use serial::{Message, foreign_keys};
use store::Hash;

use crate::catalog::table::{IndexDef, TableDef};
use crate::error::{ErrorObjects, PgError, Result, code};
use crate::expr::{compare_values, node_name};
use crate::query::{Ctx, scan};
use crate::types::Value;

/// Rule is what happens to referencing rows when a referenced row is deleted or its key changes.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Rule {
    NoAction,
    Restrict,
    Cascade,
    SetNull,
    SetDefault,
}

impl Rule {
    /// from_dolt reads Dolt's code for an action, where 0 is the default, no action.
    fn from_dolt(code: u8) -> Rule {
        match code {
            1 => Rule::Cascade,
            3 => Rule::Restrict,
            4 => Rule::SetNull,
            5 => Rule::SetDefault,
            _ => Rule::NoAction,
        }
    }

    /// to_dolt returns Dolt's code for an action.
    fn to_dolt(self) -> u8 {
        match self {
            Rule::Cascade => 1,
            Rule::NoAction => 2,
            Rule::Restrict => 3,
            Rule::SetNull => 4,
            Rule::SetDefault => 5,
        }
    }

    /// from_postgres reads the action letter of Postgres' parser.
    fn from_postgres(letter: &str) -> Rule {
        match letter {
            "r" => Rule::Restrict,
            "c" => Rule::Cascade,
            "n" => Rule::SetNull,
            "d" => Rule::SetDefault,
            _ => Rule::NoAction,
        }
    }
}

/// ForeignKeyDef is a foreign key: the referencing table and columns, the referenced ones, and the actions.
#[derive(Clone, Debug, PartialEq)]
pub struct ForeignKeyDef {
    pub name: String,
    pub child_schema: String,
    pub child_table: String,
    pub child_index: String,
    pub child_columns: Vec<String>,
    pub parent_schema: String,
    pub parent_table: String,
    pub parent_index: String,
    pub parent_columns: Vec<String>,
    pub on_update: Rule,
    pub on_delete: Rule,
    pub match_full: bool,
    pub not_valid: bool,
}

/// split_key returns the schema and name of a table key.
fn split_key(key: &[u8]) -> (String, String) {
    let text = String::from_utf8_lossy(key);
    let mut parts = text.splitn(3, '\0').skip(1);
    (parts.next().unwrap_or_default().to_string(), parts.next().unwrap_or_default().to_string())
}

/// columns returns the names of a foreign key's columns, found by tag in the table when it has them all, otherwise
/// the names stored with the foreign key.
fn columns(
    db: &mut Database,
    root: &Root,
    key: &[u8],
    tags: &[u64],
    names: &[&[u8]],
) -> Result<(String, String, Vec<String>)> {
    let (schema, table) = split_key(key);
    let def = root.table(db, &schema, &table)?.map(|a| TableDef::load(db, &schema, &table, a)).transpose()?;
    let found: Option<Vec<String>> = def.as_ref().and_then(|def| {
        tags.iter().map(|tag| def.columns.iter().find(|c| c.tag == *tag).map(|c| c.name.clone())).collect()
    });
    let columns = match found {
        Some(found) if !found.is_empty() => found,
        _ => names.iter().map(|n| String::from_utf8_lossy(n).into_owned()).collect(),
    };
    Ok((schema, table, columns))
}

/// load returns the foreign keys of a root value in their stored order.
pub fn load(db: &mut Database, root: &Root) -> Result<Vec<ForeignKeyDef>> {
    if root.foreign_keys.iter().all(|&b| b == 0) {
        return Ok(Vec::new());
    }
    let address = serial::hash(&root.foreign_keys)?;
    let data = crate::txn::read(db, &address)?;
    let mut out = Vec::new();
    for fk in foreign_keys(Message(&data))? {
        let (child_schema, child_table, child_columns) =
            columns(db, root, fk.child_table_name, &fk.child_table_columns, &fk.unresolved_child_columns)?;
        let (parent_schema, parent_table, parent_columns) =
            columns(db, root, fk.parent_table_name, &fk.parent_table_columns, &fk.unresolved_parent_columns)?;
        let lossy = |b: &[u8]| String::from_utf8_lossy(b).into_owned();
        out.push(ForeignKeyDef {
            name: lossy(fk.name),
            child_schema,
            child_table,
            child_index: lossy(fk.child_table_index),
            child_columns,
            parent_schema,
            parent_table,
            parent_index: lossy(fk.parent_table_index),
            parent_columns,
            on_update: Rule::from_dolt(fk.on_update),
            on_delete: Rule::from_dolt(fk.on_delete),
            match_full: fk.match_type == 1,
            not_valid: fk.is_not_valid,
        });
    }
    Ok(out)
}

/// store writes the foreign keys into a root value in creation order, the order Postgres checks them in.
pub fn store(db: &mut Database, root: &mut Root, fks: &[ForeignKeyDef]) -> Result<()> {
    if fks.is_empty() {
        root.foreign_keys = vec![0; Hash::LEN];
        return Ok(());
    }
    let mut prepared = Vec::new();
    for fk in fks {
        let tags = |db: &mut Database, schema: &str, table: &str, columns: &[String]| -> Result<Vec<u64>> {
            let def = root.table(db, schema, table)?.map(|a| TableDef::load(db, schema, table, a)).transpose()?;
            Ok(columns
                .iter()
                .map(|c| {
                    def.as_ref().and_then(|d| d.columns.iter().find(|col| col.name == *c)).map_or(0, |col| col.tag)
                })
                .collect())
        };
        let child_tags = tags(db, &fk.child_schema, &fk.child_table, &fk.child_columns)?;
        let parent_tags = tags(db, &fk.parent_schema, &fk.parent_table, &fk.parent_columns)?;
        prepared.push((
            fk,
            child_tags,
            parent_tags,
            table_key(&fk.child_schema, &fk.child_table),
            table_key(&fk.parent_schema, &fk.parent_table),
        ));
    }
    let child_names: Vec<Vec<Vec<u8>>> =
        prepared.iter().map(|(fk, ..)| fk.child_columns.iter().map(|c| c.as_bytes().to_vec()).collect()).collect();
    let parent_names: Vec<Vec<Vec<u8>>> =
        prepared.iter().map(|(fk, ..)| fk.parent_columns.iter().map(|c| c.as_bytes().to_vec()).collect()).collect();
    let fields: Vec<ForeignKeyFields<'_>> = prepared
        .iter()
        .zip(child_names)
        .zip(parent_names)
        .map(|(((fk, child_tags, parent_tags, child_key, parent_key), child_names), parent_names)| ForeignKeyFields {
            name: fk.name.as_bytes(),
            child_table_name: child_key,
            child_table_index: fk.child_index.as_bytes(),
            child_table_columns: child_tags.clone(),
            parent_table_name: parent_key,
            parent_table_index: fk.parent_index.as_bytes(),
            parent_table_columns: parent_tags.clone(),
            on_update: fk.on_update.to_dolt(),
            on_delete: fk.on_delete.to_dolt(),
            unresolved_child_columns: Some(child_names),
            unresolved_parent_columns: Some(parent_names),
            is_not_valid: fk.not_valid,
            match_type: fk.match_full as u8,
        })
        .collect();
    let address = db.write_value(write_foreign_keys(&fields))?;
    root.foreign_keys = address.0.to_vec();
    Ok(())
}

/// key_text renders key columns and values as Postgres' foreign key errors show them.
fn key_text(columns: &[String], values: &[Value]) -> String {
    let shown: Vec<String> = values.iter().map(|v| v.output().unwrap_or_else(|| "null".into())).collect();
    format!("Key ({})=({})", columns.join(", "), shown.join(", "))
}

/// child_violation returns Postgres' error for a referencing row whose key is missing from the referenced table.
fn child_violation(fk: &ForeignKeyDef, values: &[Value]) -> PgError {
    PgError {
        detail: Some(format!(
            "{} is not present in table \"{}\".",
            key_text(&fk.child_columns, values),
            fk.parent_table
        )),
        objects: Some(Box::new(ErrorObjects {
            schema: Some(fk.child_schema.clone()),
            table: Some(fk.child_table.clone()),
            constraint: Some(fk.name.clone()),
            ..ErrorObjects::default()
        })),
        ..PgError::new(
            code::FOREIGN_KEY_VIOLATION,
            format!("insert or update on table \"{}\" violates foreign key constraint \"{}\"", fk.child_table, fk.name),
        )
    }
}

/// parent_violation returns Postgres' error for removing a key that rows still refer to.
fn parent_violation(fk: &ForeignKeyDef, values: &[Value]) -> PgError {
    PgError {
        detail: Some(format!(
            "{} is still referenced from table \"{}\".",
            key_text(&fk.parent_columns, values),
            fk.child_table
        )),
        objects: Some(Box::new(ErrorObjects {
            schema: Some(fk.child_schema.clone()),
            table: Some(fk.child_table.clone()),
            constraint: Some(fk.name.clone()),
            ..ErrorObjects::default()
        })),
        ..PgError::new(
            code::FOREIGN_KEY_VIOLATION,
            format!(
                "update or delete on table \"{}\" violates foreign key constraint \"{}\" on table \"{}\"",
                fk.parent_table, fk.name, fk.child_table
            ),
        )
    }
}

/// family returns the operator family that compares a type's values with the other types of its family, or the type
/// itself for a family of one.
fn family(ty: u32) -> u32 {
    use crate::oid::*;
    match ty {
        INT2 | INT4 | INT8 => INT8,
        FLOAT4 | FLOAT8 => FLOAT8,
        VARCHAR | NAME => TEXT,
        other => other,
    }
}

/// converted returns key values converted to the types of the columns they are compared with.
fn converted(key: Vec<Value>, table: &TableDef, columns: &[usize]) -> Result<Vec<Value>> {
    key.into_iter().zip(columns).map(|(v, &c)| crate::cast::cast_value(v, table.columns[c].ty, false)).collect()
}

/// positions returns the positions of columns in a table.
fn positions(table: &TableDef, columns: &[String]) -> Vec<usize> {
    columns.iter().filter_map(|c| table.columns.iter().position(|col| col.name == *c)).collect()
}

/// values returns a row's values in columns.
fn values(row: &[Value], columns: &[usize]) -> Vec<Value> {
    columns.iter().map(|&c| row[c].clone()).collect()
}

/// same_key reports whether two keys are equal, with NULLs equal to nothing.
fn same_key(a: &[Value], b: &[Value]) -> bool {
    a.len() == b.len()
        && a.iter()
            .zip(b)
            .all(|(x, y)| !x.is_null() && !y.is_null() && compare_values(x, y) == std::cmp::Ordering::Equal)
}

/// covering_index returns the index whose leading columns are a foreign key's columns in order, empty for the
/// primary key.
fn covering_index(table: &TableDef, columns: &[usize]) -> Option<String> {
    let leads = |index: &[usize]| index.starts_with(columns);
    if !table.keyless() && leads(&table.key_columns) {
        return Some(String::new());
    }
    table.indexes.iter().find(|i| leads(&i.columns)).map(|i| i.name.clone())
}

/// ParentRows are the referenced table of a foreign key, the positions of its referenced columns, and its rows.
type ParentRows = (Option<TableDef>, Vec<usize>, Vec<Vec<Value>>);

/// Change is a row change of a statement: the old row, the new row, or both for an update.
pub type Change = (Option<Vec<Value>>, Option<Vec<Value>>);

impl Ctx<'_> {
    /// foreign_keys returns the foreign keys of the working root.
    pub fn foreign_keys(&mut self) -> Result<Vec<ForeignKeyDef>> {
        load(self.db, &self.txn.root)
    }

    /// add_foreign_key adds a foreign key to a table, creating an index on its columns when none leads with them, and
    /// checks the table's rows against it unless it is NOT VALID.
    pub fn add_foreign_key(&mut self, child: &TableDef, columns: &[usize], constraint: &Constraint) -> Result<()> {
        let relation = constraint.pktable.as_ref().ok_or_else(|| PgError::internal("a foreign key without a table"))?;
        let parent = if relation.relname == child.name
            && (relation.schemaname.is_empty() || relation.schemaname == child.schema)
        {
            child.clone()
        } else {
            self.resolve_table(relation)?
        };
        let referenced: Vec<usize> = if constraint.pk_attrs.is_empty() {
            if parent.keyless() {
                return Err(PgError::new(
                    code::UNDEFINED_OBJECT,
                    format!("there is no primary key for referenced table \"{}\"", parent.name),
                ));
            }
            parent.key_columns.clone()
        } else {
            let mut out = Vec::new();
            for name in constraint.pk_attrs.iter().filter_map(node_name) {
                out.push(parent.columns.iter().position(|c| c.name == name).ok_or_else(|| {
                    PgError::new(
                        code::UNDEFINED_COLUMN,
                        format!("column \"{name}\" referenced in foreign key constraint does not exist"),
                    )
                })?);
            }
            out
        };
        if referenced.len() != columns.len() {
            return Err(PgError::new(
                code::INVALID_FOREIGN_KEY,
                "number of referencing and referenced columns for foreign key disagree",
            ));
        }
        let sorted = |v: &[usize]| {
            let mut v = v.to_vec();
            v.sort_unstable();
            v
        };
        let parent_index = if !parent.keyless() && sorted(&parent.key_columns) == sorted(&referenced) {
            String::new()
        } else if let Some(index) =
            parent.indexes.iter().find(|i| i.unique && sorted(&i.columns) == sorted(&referenced))
        {
            index.name.clone()
        } else {
            return Err(PgError::new(
                code::INVALID_FOREIGN_KEY,
                format!("there is no unique constraint matching given keys for referenced table \"{}\"", parent.name),
            ));
        };
        let name = if constraint.conname.is_empty() {
            let names: Vec<&str> = columns.iter().map(|&c| child.columns[c].name.as_str()).collect();
            let taken = self.constraint_names(&child.schema)?;
            crate::ddl::choose_relation_name(&child.name, &names.join("_"), "fkey", &taken)
        } else {
            constraint.conname.clone()
        };
        for (&c, &p) in columns.iter().zip(&referenced) {
            let (ct, pt) = (child.columns[c].ty.oid, parent.columns[p].ty.oid);
            if family(ct) != family(pt) && !crate::expr::implicitly_converts(ct, pt) {
                return Err(PgError {
                    detail: Some(format!(
                        "Key columns \"{}\" and \"{}\" are of incompatible types: {} and {}.",
                        child.columns[c].name,
                        parent.columns[p].name,
                        crate::cast::type_display(ct),
                        crate::cast::type_display(pt)
                    )),
                    ..PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!("foreign key constraint \"{name}\" cannot be implemented"),
                    )
                });
            }
        }
        let mut fks = self.foreign_keys()?;
        if fks.iter().any(|fk| fk.name == name && fk.child_table == child.name && fk.child_schema == child.schema) {
            return Err(PgError::new(
                code::DUPLICATE_OBJECT,
                format!("constraint \"{name}\" for relation \"{}\" already exists", child.name),
            ));
        }
        let child_index = if let Some(index) = covering_index(child, columns) {
            index
        } else {
            let index = IndexDef { system: true, ..crate::ddl::new_index(name.clone(), columns.to_vec(), false) };
            self.build_index(child.clone(), index)?;
            name.clone()
        };
        let fk = ForeignKeyDef {
            name,
            child_schema: child.schema.clone(),
            child_table: child.name.clone(),
            child_index,
            child_columns: columns.iter().map(|&c| child.columns[c].name.clone()).collect(),
            parent_schema: parent.schema.clone(),
            parent_table: parent.name.clone(),
            parent_index,
            parent_columns: referenced.iter().map(|&p| parent.columns[p].name.clone()).collect(),
            on_update: Rule::from_postgres(&constraint.fk_upd_action),
            on_delete: Rule::from_postgres(&constraint.fk_del_action),
            match_full: constraint.fk_matchtype == "f",
            not_valid: constraint.skip_validation,
        };
        if !fk.not_valid {
            let rows = match self.txn.table(self.db, &child.schema, &child.name)? {
                Some(table) => scan(self.db, &table)?,
                None => Vec::new(),
            };
            let changes: Vec<Change> = rows.into_iter().map(|r| (None, Some(r))).collect();
            self.check_children(&fk, child, &changes)?;
        }
        fks.push(fk);
        store(self.db, &mut self.txn.root, &fks)
    }

    /// check_children fails when a new or changed row of a foreign key's table refers to a missing key.
    fn check_children(&mut self, fk: &ForeignKeyDef, child: &TableDef, changes: &[Change]) -> Result<()> {
        let columns = positions(child, &fk.child_columns);
        let mut parent_rows: Option<ParentRows> = None;
        for (_, new) in changes {
            let Some(row) = new else { continue };
            let key = values(row, &columns);
            let nulls = key.iter().filter(|v| v.is_null()).count();
            if nulls == key.len() || (nulls > 0 && !fk.match_full) {
                continue;
            }
            if nulls > 0 {
                return Err(PgError {
                    detail: Some("MATCH FULL does not allow mixing of null and nonnull key values.".into()),
                    ..child_violation(fk, &key)
                });
            }
            if parent_rows.is_none() {
                let parent = self.txn.table(self.db, &fk.parent_schema, &fk.parent_table)?;
                parent_rows = Some(match parent {
                    Some(parent) => {
                        let rows = scan(self.db, &parent)?;
                        let columns = positions(&parent, &fk.parent_columns);
                        (Some(parent), columns, rows)
                    }
                    None => (None, Vec::new(), Vec::new()),
                });
            }
            let (parent, parent_columns, rows) = parent_rows.as_ref().expect("parent rows were read");
            let wanted = match parent {
                Some(parent) => converted(key.clone(), parent, parent_columns).ok(),
                None => Some(key.clone()),
            };
            if !wanted.is_some_and(|wanted| rows.iter().any(|r| same_key(&values(r, parent_columns), &wanted))) {
                return Err(child_violation(fk, &key));
            }
        }
        Ok(())
    }

    /// enforce_foreign_keys checks a statement's changes to a table against the foreign keys it is part of, carrying
    /// out the referential actions of rows that refer to removed or changed keys.
    pub fn enforce_foreign_keys(&mut self, table: &TableDef, changes: &[Change]) -> Result<()> {
        if changes.is_empty() {
            return Ok(());
        }
        let fks = self.foreign_keys()?;
        for fk in fks.iter().filter(|fk| fk.child_table == table.name && fk.child_schema == table.schema) {
            self.check_children(fk, table, changes)?;
        }
        let referencing: Vec<&ForeignKeyDef> =
            fks.iter().filter(|fk| fk.parent_table == table.name && fk.parent_schema == table.schema).collect();
        for checks in [true, false] {
            for fk in &referencing {
                self.enforce_parent(fk, table, changes, checks)?;
            }
        }
        Ok(())
    }

    /// enforce_parent carries out a foreign key's action for the referenced keys that a statement removed or changed,
    /// either the checks of NO ACTION and RESTRICT or the changes of the other actions.
    fn enforce_parent(
        &mut self,
        fk: &ForeignKeyDef,
        parent: &TableDef,
        changes: &[Change],
        checks: bool,
    ) -> Result<()> {
        let parent_columns = positions(parent, &fk.parent_columns);
        let current = self.txn.table(self.db, &parent.schema, &parent.name)?;
        let remaining = match &current {
            Some(t) => scan(self.db, t)?,
            None => Vec::new(),
        };
        let mut removed: Vec<(Vec<Value>, Option<Vec<Value>>)> = Vec::new();
        for (old, new) in changes {
            let Some(old) = old else { continue };
            let key = values(old, &parent_columns);
            if key.iter().any(Value::is_null) || remaining.iter().any(|r| same_key(&values(r, &parent_columns), &key)) {
                continue;
            }
            removed.push((key, new.as_ref().map(|n| values(n, &parent_columns))));
        }
        removed.retain(|(_, new_key)| {
            let action = if new_key.is_some() { fk.on_update } else { fk.on_delete };
            matches!(action, Rule::NoAction | Rule::Restrict) == checks
        });
        if removed.is_empty() {
            return Ok(());
        }
        let Some(child) = self.txn.table(self.db, &fk.child_schema, &fk.child_table)? else { return Ok(()) };
        let child_columns = positions(&child, &fk.child_columns);
        let child_rows = scan(self.db, &child)?;
        let mut child_changes: Vec<Change> = Vec::new();
        for (key, new_key) in &removed {
            let action = if new_key.is_some() { fk.on_update } else { fk.on_delete };
            let mut referencing = Vec::new();
            for row in &child_rows {
                if converted(values(row, &child_columns), parent, &parent_columns).is_ok_and(|v| same_key(&v, key)) {
                    referencing.push(row);
                }
            }
            for row in referencing {
                match action {
                    Rule::NoAction | Rule::Restrict => return Err(parent_violation(fk, key)),
                    Rule::Cascade => match new_key {
                        None => child_changes.push((Some(row.clone()), None)),
                        Some(new_key) => {
                            let mut updated = row.clone();
                            for (&c, v) in child_columns.iter().zip(converted(new_key.clone(), &child, &child_columns)?)
                            {
                                updated[c] = v;
                            }
                            child_changes.push((Some(row.clone()), Some(updated)));
                        }
                    },
                    Rule::SetNull | Rule::SetDefault => {
                        let rules = self.row_rules(&child)?;
                        let mut updated = row.clone();
                        for &c in &child_columns {
                            updated[c] = if action == Rule::SetNull {
                                Value::Null
                            } else {
                                crate::dml::default_value(self, &rules, c)?
                            };
                        }
                        child_changes.push((Some(row.clone()), Some(updated)));
                    }
                }
            }
        }
        crate::dml::apply_changes(self, &child, &child_changes)?;
        let child = self
            .txn
            .table(self.db, &fk.child_schema, &fk.child_table)?
            .ok_or_else(|| PgError::internal("a referencing table vanished"))?;
        self.enforce_foreign_keys(&child, &child_changes)
    }

    /// drop_table_foreign_keys drops the views and foreign keys that depend on a table being dropped, failing as
    /// Postgres does unless the drop cascades, and drops the table's own foreign keys.
    pub fn drop_table_foreign_keys(
        &mut self,
        schema: &str,
        table: &str,
        behavior: i32,
        dropping: &[(String, String)],
    ) -> Result<()> {
        let fks = self.foreign_keys()?;
        let referencing: Vec<ForeignKeyDef> = fks
            .iter()
            .filter(|fk| {
                fk.parent_schema == schema
                    && fk.parent_table == table
                    && !dropping.iter().any(|(s, t)| *s == fk.child_schema && *t == fk.child_table)
            })
            .cloned()
            .collect();
        let constraints: Vec<(String, String)> =
            referencing.iter().map(|fk| (fk.name.clone(), fk.child_table.clone())).collect();
        self.drop_dependents(schema, table, "table", behavior, &constraints)?;
        for fk in &referencing {
            self.drop_other_foreign_key(fk)?;
        }
        let fks = self.foreign_keys()?;
        let kept: Vec<ForeignKeyDef> =
            fks.iter().filter(|fk| !(fk.child_schema == schema && fk.child_table == table)).cloned().collect();
        if kept.len() == fks.len() { Ok(()) } else { store(self.db, &mut self.txn.root, &kept) }
    }

    /// drop_referencing_foreign_keys fails as Postgres does when foreign keys refer to a unique index being dropped,
    /// an empty index meaning the primary key, unless the drop cascades, which drops those foreign keys.
    pub fn drop_referencing_foreign_keys(
        &mut self,
        parent: &mut TableDef,
        index: &str,
        object: &str,
        cascade: bool,
    ) -> Result<()> {
        let referencing: Vec<ForeignKeyDef> = self
            .foreign_keys()?
            .into_iter()
            .filter(|fk| {
                fk.parent_schema == parent.schema && fk.parent_table == parent.name && fk.parent_index == index
            })
            .collect();
        if referencing.is_empty() {
            return Ok(());
        }
        let index_name = if index.is_empty() { format!("{}_pkey", parent.name) } else { index.to_string() };
        if !cascade {
            let detail: Vec<String> = referencing
                .iter()
                .map(|fk| format!("constraint {} on table {} depends on index {index_name}", fk.name, fk.child_table))
                .collect();
            return Err(dependents_error(object.to_string(), detail));
        }
        self.notice_cascades(
            referencing
                .iter()
                .map(|fk| format!("drop cascades to constraint {} on table {}", fk.name, fk.child_table))
                .collect(),
        );
        for fk in &referencing {
            if fk.child_schema == parent.schema && fk.child_table == parent.name {
                self.drop_foreign_key(parent, &fk.name)?;
            } else {
                self.drop_other_foreign_key(fk)?;
            }
        }
        Ok(())
    }

    /// drop_column_foreign_keys drops the foreign keys that a column being dropped is part of: those of its table
    /// silently, and those that refer to it only when the drop cascades, failing as Postgres does otherwise.
    pub fn drop_column_foreign_keys(&mut self, table: &mut TableDef, column: &str, cascade: bool) -> Result<()> {
        let fks = self.foreign_keys()?;
        let referencing: Vec<&ForeignKeyDef> = fks
            .iter()
            .filter(|fk| {
                fk.parent_schema == table.schema
                    && fk.parent_table == table.name
                    && fk.parent_columns.iter().any(|c| c == column)
                    && !(fk.child_schema == table.schema
                        && fk.child_table == table.name
                        && fk.child_columns.iter().any(|c| c == column))
            })
            .collect();
        if !referencing.is_empty() {
            let object = format!("column {column} of table {}", table.name);
            if !cascade {
                let detail = referencing
                    .iter()
                    .map(|fk| format!("constraint {} on table {} depends on {object}", fk.name, fk.child_table))
                    .collect();
                return Err(dependents_error(object, detail));
            }
            self.notice_cascades(
                referencing
                    .iter()
                    .map(|fk| format!("drop cascades to constraint {} on table {}", fk.name, fk.child_table))
                    .collect(),
            );
        }
        for fk in &fks {
            let child = fk.child_schema == table.schema && fk.child_table == table.name;
            if child && (fk.child_columns.iter().any(|c| c == column) || referencing.contains(&fk)) {
                self.drop_foreign_key(table, &fk.name)?;
            } else if referencing.contains(&fk) {
                self.drop_other_foreign_key(fk)?;
            }
        }
        Ok(())
    }

    /// drop_other_foreign_key drops a foreign key of a table that no statement is altering.
    fn drop_other_foreign_key(&mut self, fk: &ForeignKeyDef) -> Result<()> {
        let Some(mut child) = self.txn.table(self.db, &fk.child_schema, &fk.child_table)? else { return Ok(()) };
        self.drop_foreign_key(&mut child, &fk.name)?;
        self.finish_alteration(crate::alter::Alteration::new(child))
    }

    /// drop_foreign_key removes a table's foreign key by name, with the index made for it, reporting whether it had
    /// one, and leaves writing the table to the caller.
    pub fn drop_foreign_key(&mut self, table: &mut TableDef, name: &str) -> Result<bool> {
        let fks = self.foreign_keys()?;
        let Some(fk) = fks
            .iter()
            .find(|fk| fk.child_schema == table.schema && fk.child_table == table.name && fk.name == name)
            .cloned()
        else {
            return Ok(false);
        };
        let kept: Vec<ForeignKeyDef> = fks.into_iter().filter(|f| *f != fk).collect();
        store(self.db, &mut self.txn.root, &kept)?;
        if let Some(i) = table.indexes.iter().position(|ix| ix.system && ix.name == fk.child_index)
            && !kept.iter().any(|f| {
                f.child_schema == table.schema && f.child_table == table.name && f.child_index == fk.child_index
            })
        {
            table.indexes.remove(i);
            table.table.put_index(self.db, &fk.child_index, None)?;
        }
        Ok(true)
    }

    /// rename_in_foreign_keys follows a table rename, or a column rename when the names match, in foreign keys read
    /// before it.
    pub fn rename_in_foreign_keys(
        &mut self,
        mut fks: Vec<ForeignKeyDef>,
        schema: &str,
        old: &str,
        new: &str,
    ) -> Result<()> {
        if fks.is_empty() {
            return Ok(());
        }
        let current = self.foreign_keys()?;
        for (fk, now) in fks.iter_mut().zip(current) {
            if fk.child_schema == schema && fk.child_table == old {
                fk.child_table = new.to_string();
                fk.child_columns = now.child_columns.clone();
            }
            if fk.parent_schema == schema && fk.parent_table == old {
                fk.parent_table = new.to_string();
                fk.parent_columns = now.parent_columns;
            }
        }
        store(self.db, &mut self.txn.root, &fks)
    }

    /// check_truncate fails as Postgres does when a foreign key of a table outside a TRUNCATE refers to one inside it.
    pub fn check_truncate(&mut self, tables: &[TableDef]) -> Result<()> {
        let inside = |schema: &str, name: &str| tables.iter().any(|t| t.schema == schema && t.name == name);
        for fk in self.foreign_keys()? {
            if inside(&fk.parent_schema, &fk.parent_table) && !inside(&fk.child_schema, &fk.child_table) {
                return Err(PgError {
                    detail: Some(format!("Table \"{}\" references \"{}\".", fk.child_table, fk.parent_table)),
                    hint: Some(format!(
                        "Truncate table \"{}\" at the same time, or use TRUNCATE ... CASCADE.",
                        fk.child_table
                    )),
                    ..PgError::new(
                        code::FEATURE_NOT_SUPPORTED,
                        "cannot truncate a table referenced in a foreign key constraint",
                    )
                });
            }
        }
        Ok(())
    }

    /// replace_child_index points the foreign keys that used a dropped index of their table at another index that
    /// leads with their columns, creating one when none does.
    pub fn replace_child_index(&mut self, schema: &str, table: &str, dropped: &str) -> Result<()> {
        let mut fks = self.foreign_keys()?;
        for fk in &mut fks {
            if fk.child_schema != schema || fk.child_table != table || fk.child_index != dropped {
                continue;
            }
            let child = self.txn.table(self.db, schema, table)?.ok_or_else(|| PgError::internal("a table vanished"))?;
            let columns = positions(&child, &fk.child_columns);
            let index = match covering_index(&child, &columns) {
                Some(index) => index,
                None => {
                    let index = IndexDef { system: true, ..crate::ddl::new_index(fk.name.clone(), columns, false) };
                    self.build_index(child, index)?;
                    fk.name.clone()
                }
            };
            fk.child_index = index;
        }
        store(self.db, &mut self.txn.root, &fks)
    }

    /// rename_foreign_key renames a table's foreign key, reporting whether it had one by the old name.
    pub fn rename_foreign_key(&mut self, table: &TableDef, old: &str, new: &str) -> Result<bool> {
        let mut fks = self.foreign_keys()?;
        let Some(fk) =
            fks.iter_mut().find(|fk| fk.child_schema == table.schema && fk.child_table == table.name && fk.name == old)
        else {
            return Ok(false);
        };
        fk.name = new.to_string();
        store(self.db, &mut self.txn.root, &fks)?;
        Ok(true)
    }
}

/// dependents_error returns Postgres' error for dropping an object that foreign keys depend on.
fn dependents_error(object: String, detail: Vec<String>) -> PgError {
    PgError {
        detail: Some(detail.join("\n")),
        hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
        ..PgError::new(
            code::DEPENDENT_OBJECTS_STILL_EXIST,
            format!("cannot drop {object} because other objects depend on it"),
        )
    }
}
