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

//! DOLT_SCHEMA_DIFF and DOLT_PATCH: the CREATE TABLE statements of changed tables as Doltgres' schema formatter writes
//! them, and the statements that turn one revision's tables into another's, written as Postgres runs them.

use std::cmp::Ordering;

use doltdb::root::Root;
use store::Hash;

use crate::catalog::ColumnType;
use crate::catalog::table::{ColumnDef, TableDef};
use crate::dolt::diff::{Delta, Name, changes, deltas, diff_refs, full_name, matches, project, ref_root};
use crate::dolt::history;
use crate::error::{PgError, Result, code};
use crate::foreign::{ForeignKeyDef, Rule};
use crate::query::Ctx;
use crate::types::Value;

/// quote_identifier double-quotes a name, as Doltgres' schema formatter quotes every identifier.
fn quote_identifier(name: &str) -> String {
    format!("\"{}\"", name.replace('"', "\"\""))
}

/// type_text returns a column type as Doltgres names a type: its internal name or its name, with the modifier as the
/// type's typmodout prints it.
fn type_text(ty: ColumnType) -> String {
    if crate::array::is_array_type(ty.oid) && !crate::array::is_vector_type(ty.oid) {
        let element = ColumnType { oid: crate::expr::element_type(ty.oid), modifier: ty.modifier };
        return format!("{}[]", type_text(element));
    }
    let name = match crate::catalog::builtin_type(ty.oid) {
        Some(b) if !b.definition.internal_name.is_empty() => {
            String::from_utf8_lossy(&b.definition.internal_name).into()
        }
        Some(b) => b.name.to_string(),
        None => crate::usertypes::get(ty.oid).map(|t| t.name.clone()).unwrap_or_default(),
    };
    let m = ty.modifier;
    let suffix = match ty.oid {
        _ if m < 0 => String::new(),
        crate::oid::BPCHAR | crate::oid::VARCHAR if m > 4 => format!("({})", m - 4),
        crate::oid::NUMERIC if m >= 4 => format!("({},{})", (m - 4) >> 16, (((m - 4) & 0x7ff) ^ 1024) - 1024),
        crate::oid::BIT
        | crate::oid::VARBIT
        | crate::oid::TIME
        | crate::oid::TIMETZ
        | crate::oid::TIMESTAMP
        | crate::oid::TIMESTAMPTZ => format!("({m})"),
        crate::oid::INTERVAL if m & 0xffff != 0xffff => format!("({})", m & 0xffff),
        _ => String::new(),
    };
    format!("{name}{suffix}")
}

/// column_definition returns a column's definition as Doltgres' schema formatter writes it, quoting a stored default
/// that is a bare constant as Dolt does, and parenthesizing one that is an expression.
fn column_definition(column: &ColumnDef) -> String {
    let mut out = format!("{} {}", quote_identifier(&column.name), type_text(column.ty));
    if !column.nullable {
        out.push_str(" NOT NULL");
    }
    let stored = &column.default;
    if column.generated {
        out.push_str(&format!(" GENERATED ALWAYS AS ({stored}) STORED"));
    } else if !stored.is_empty() {
        let constant = crate::parse::expression_node(stored)
            .is_ok_and(|node| matches!(node.node, Some(pg_query::NodeEnum::AConst(_))));
        let enclosed = stored == "NULL" || (stored.starts_with('(') && stored.ends_with(')'));
        match (enclosed || stored.starts_with('\''), constant) {
            (true, _) => out.push_str(&format!(" DEFAULT {stored}")),
            (false, true) => out.push_str(&format!(" DEFAULT '{stored}'")),
            (false, false) => out.push_str(&format!(" DEFAULT ({stored})")),
        }
    }
    out
}

/// rule_text returns the words of a foreign key action.
fn rule_text(rule: Rule) -> &'static str {
    match rule {
        Rule::NoAction => "NO ACTION",
        Rule::Restrict => "RESTRICT",
        Rule::Cascade => "CASCADE",
        Rule::SetNull => "SET NULL",
        Rule::SetDefault => "SET DEFAULT",
    }
}

/// quoted_list returns names quoted and joined with commas.
fn quoted_list(names: &[String]) -> String {
    names.iter().map(|n| quote_identifier(n)).collect::<Vec<_>>().join(",")
}

/// foreign_key_clause returns a foreign key's REFERENCES clause.
fn foreign_key_clause(key: &ForeignKeyDef) -> String {
    format!(
        "FOREIGN KEY ({}) REFERENCES {} ({}) ON DELETE {} ON UPDATE {}",
        quoted_list(&key.child_columns),
        quote_identifier(&key.parent_table),
        quoted_list(&key.parent_columns),
        rule_text(key.on_delete),
        rule_text(key.on_update)
    )
}

/// index_columns returns an index's quoted columns with the directions of descending ones.
fn index_columns(table: &TableDef, columns: &[usize], descending: &[bool]) -> String {
    let parts: Vec<String> = columns
        .iter()
        .enumerate()
        .map(|(i, &c)| {
            let desc = if descending.get(i).copied().unwrap_or(false) { " DESC" } else { "" };
            format!("{}{desc}", quote_identifier(&table.columns[c].name))
        })
        .collect();
    parts.join(",")
}

/// is_primary_index reports whether an index covers exactly the primary key's columns, which Dolt leaves out of a
/// CREATE TABLE statement.
fn is_primary_index(table: &TableDef, columns: &[usize]) -> bool {
    let mut index = columns.to_vec();
    let mut key = table.key_columns.clone();
    index.sort_unstable();
    key.sort_unstable();
    index == key
}

/// create_table_statement returns a table's CREATE TABLE statement as Doltgres' schema formatter writes it: its
/// columns, primary key, unique indexes, foreign keys, and checks.
pub fn create_table_statement(table: &TableDef, foreign_keys: &[ForeignKeyDef]) -> String {
    let mut parts: Vec<String> = table.columns.iter().map(|c| format!("  {}", column_definition(c))).collect();
    if !table.key_columns.is_empty() {
        let names: Vec<String> = table.key_columns.iter().map(|&c| table.columns[c].name.clone()).collect();
        parts.push(format!("  PRIMARY KEY ({})", quoted_list(&names)));
    }
    for index in table.indexes.iter().filter(|i| i.unique && !is_primary_index(table, &i.columns)) {
        let columns = index_columns(table, &index.columns, &index.descending);
        parts.push(format!("  CONSTRAINT {} UNIQUE ({columns})", quote_identifier(&index.name)));
    }
    for key in foreign_keys.iter().filter(|k| k.child_schema == table.schema && k.child_table == table.name) {
        parts.push(format!("  CONSTRAINT {} {}", quote_identifier(&key.name), foreign_key_clause(key)));
    }
    for check in &table.checks {
        parts.push(format!("  CONSTRAINT {} CHECK ({})", quote_identifier(&check.name), check.expression));
    }
    format!("CREATE TABLE {} (\n{}\n);", quote_identifier(&table.name), parts.join(",\n"))
}

/// Side is one revision of a changed table: its definition and the foreign keys of its root.
struct Side {
    table: TableDef,
    foreign_keys: Vec<ForeignKeyDef>,
}

/// sides loads the tables of a delta at each revision, with each revision's foreign keys.
fn sides(ctx: &mut Ctx<'_>, delta: &Delta, from: &Root, to: &Root) -> Result<(Option<Side>, Option<Side>)> {
    let mut load = |side: &Option<(Name, Hash)>, root: &Root| -> Result<Option<Side>> {
        let Some((name, address)) = side else { return Ok(None) };
        let table = TableDef::load(ctx.db, &name.0, &name.1, *address)?;
        Ok(Some(Side { table, foreign_keys: crate::foreign::load(ctx.db, root)? }))
    };
    Ok((load(&delta.from, from)?, load(&delta.to, to)?))
}

/// sorted_deltas returns the table changes between two roots in the order of their new names, which dropped tables
/// lack, as Dolt sorts them.
fn sorted_deltas(ctx: &mut Ctx<'_>, from: &Root, to: &Root) -> Result<Vec<Delta>> {
    let key = |side: &Option<(Name, Hash)>| side.as_ref().map(|(n, _)| n.clone()).unwrap_or_default();
    let mut all: Vec<Delta> = deltas(ctx.db, from, to)?.into_iter().filter(|d| !d.object).collect();
    all.sort_by_key(|d| key(&d.to));
    Ok(all)
}

/// dolt_schema_diff returns the CREATE TABLE statements of each table whose definition differs between two
/// revisions, or only of the table the third argument names, as rows of records.
pub fn dolt_schema_diff(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let args: Vec<String> = args.iter().map(|a| a.output().unwrap_or_default()).collect();
    let (from_ref, to_ref, rest) = diff_refs(ctx, &args, "dolt_schema_diff", 2..=3, 1..=2)?;
    let (from, to) = (ref_root(ctx, &from_ref)?, ref_root(ctx, &to_ref)?);
    let mut rows = Vec::new();
    for delta in sorted_deltas(ctx, &from, &to)? {
        let named = |side: &Option<(Name, Hash)>| side.as_ref().is_some_and(|(n, _)| n.1 == rest[0]);
        if !rest.is_empty() && !named(&delta.to) && !named(&delta.from) {
            continue;
        }
        let (old, new) = sides(ctx, &delta, &from, &to)?;
        let statement = |side: &Option<Side>| {
            side.as_ref().map(|s| create_table_statement(&s.table, &s.foreign_keys)).unwrap_or_default()
        };
        let (old_statement, new_statement) = (statement(&old), statement(&new));
        if old_statement == new_statement {
            continue;
        }
        let name = |side: &Option<(Name, Hash)>| side.as_ref().map(|(n, _)| full_name(n)).unwrap_or_default();
        rows.push(Value::Record(vec![
            Value::Text(name(&delta.from)),
            Value::Text(name(&delta.to)),
            Value::Text(old_statement),
            Value::Text(new_statement),
        ]));
    }
    Ok(Value::Set(rows))
}

/// literal returns a value as a Postgres literal: a number as it prints, NULL, or anything else as a quoted string.
fn literal(value: &Value) -> String {
    let text = value.output();
    match value {
        Value::Null => "NULL".into(),
        Value::Int2(_) | Value::Int4(_) | Value::Int8(_) | Value::Oid(_) => text.unwrap_or_default(),
        Value::Float4(_) | Value::Float8(_) | Value::Numeric(_)
            if text.as_deref().is_some_and(|t| t.parse::<f64>().is_ok_and(f64::is_finite)) =>
        {
            text.unwrap_or_default()
        }
        _ => format!("'{}'", text.unwrap_or_default().replace('\'', "''")),
    }
}

/// key_condition returns the WHERE condition that finds a row by its primary key, or by all its columns in a keyless
/// table.
fn key_condition(table: &TableDef, row: &[Value]) -> String {
    let columns: Vec<usize> =
        if table.keyless() { (0..table.columns.len()).collect() } else { table.key_columns.clone() };
    let parts: Vec<String> =
        columns.iter().map(|&c| format!("{}={}", quote_identifier(&table.columns[c].name), literal(&row[c]))).collect();
    parts.join(" AND ")
}

/// data_statements returns the INSERT, UPDATE, and DELETE statements that turn a table's rows at one revision into
/// its rows at another, in key order, or nothing when the revisions' primary keys differ.
fn data_statements(ctx: &mut Ctx<'_>, old: Option<&TableDef>, new: &TableDef) -> Result<Vec<String>> {
    let Some(changes) = changes(ctx.db, old, Some(new))? else { return Ok(Vec::new()) };
    let name = quote_identifier(&new.name);
    let inserted: Vec<usize> = (0..new.columns.len()).filter(|&c| !new.columns[c].generated).collect();
    let mut out = Vec::new();
    for change in changes {
        match change {
            (None, Some(row)) => {
                let columns: Vec<String> = inserted.iter().map(|&c| new.columns[c].name.clone()).collect();
                let values: Vec<String> = inserted.iter().map(|&c| literal(&row[c])).collect();
                out.push(format!("INSERT INTO {name} ({}) VALUES ({});", quoted_list(&columns), values.join(",")));
            }
            (Some(row), None) => {
                let row = match old {
                    Some(table) => project(table, &row, &new.columns),
                    None => row,
                };
                out.push(format!("DELETE FROM {name} WHERE {};", key_condition(new, &row)));
            }
            (Some(before), Some(after)) => {
                let before = match old {
                    Some(table) => project(table, &before, &new.columns),
                    None => before,
                };
                let set: Vec<String> = inserted
                    .iter()
                    .filter(|&&c| crate::expr::compare_values(&before[c], &after[c]) != Ordering::Equal)
                    .map(|&c| format!("{}={}", quote_identifier(&new.columns[c].name), literal(&after[c])))
                    .collect();
                if !set.is_empty() {
                    out.push(format!("UPDATE {name} SET {} WHERE {};", set.join(","), key_condition(new, &after)));
                }
            }
            (None, None) => {}
        }
    }
    Ok(out)
}

/// schema_statements returns the statements that turn a table's definition at one revision into its definition at
/// another, as Dolt's GenerateSqlPatchSchemaStatements does, but in Postgres' syntax.
fn schema_statements(old: Option<&Side>, new: Option<&Side>) -> Vec<String> {
    let (old, new) = match (old, new) {
        (Some(old), None) => return vec![format!("DROP TABLE {};", quote_identifier(&old.table.name))],
        (None, Some(new)) => return vec![create_table_statement(&new.table, &new.foreign_keys)],
        (Some(old), Some(new)) => (old, new),
        (None, None) => return Vec::new(),
    };
    let (from, to) = (&old.table, &new.table);
    let name = quote_identifier(&to.name);
    let mut out = Vec::new();
    if from.name != to.name {
        out.push(format!("ALTER TABLE {} RENAME TO {name};", quote_identifier(&from.name)));
    }
    let mut tags: Vec<u64> = from.columns.iter().map(|c| c.tag).collect();
    tags.extend(to.columns.iter().map(|c| c.tag).filter(|t| !from.columns.iter().any(|c| c.tag == *t)));
    for tag in tags {
        let before = from.columns.iter().find(|c| c.tag == tag);
        let after = to.columns.iter().find(|c| c.tag == tag);
        match (before, after) {
            (None, Some(after)) => out.push(format!("ALTER TABLE {name} ADD {};", column_definition(after))),
            (Some(before), None) => out.push(format!("ALTER TABLE {name} DROP {};", quote_identifier(&before.name))),
            (Some(before), Some(after)) if before.primary_key == after.primary_key => {
                if before.name != after.name {
                    out.push(format!(
                        "ALTER TABLE {name} RENAME COLUMN {} TO {};",
                        quote_identifier(&before.name),
                        quote_identifier(&after.name)
                    ));
                }
                if before.ty != after.ty {
                    out.push(format!(
                        "ALTER TABLE {name} ALTER COLUMN {} TYPE {};",
                        quote_identifier(&after.name),
                        type_text(after.ty)
                    ));
                }
            }
            _ => {}
        }
    }
    let key = |t: &TableDef| t.key_columns.iter().map(|&c| t.columns[c].tag).collect::<Vec<_>>();
    if key(from) != key(to) {
        out.push(format!("ALTER TABLE {name} DROP CONSTRAINT {};", quote_identifier(&from.primary_name())));
        if !to.key_columns.is_empty() {
            let names: Vec<String> = to.key_columns.iter().map(|&c| to.columns[c].name.clone()).collect();
            out.push(format!("ALTER TABLE {name} ADD PRIMARY KEY ({});", quoted_list(&names)));
        }
    }
    let index_text =
        |t: &TableDef, i: &crate::catalog::table::IndexDef| (i.unique, index_columns(t, &i.columns, &i.descending));
    for index in from.indexes.iter().filter(|i| !i.system) {
        let kept = to.indexes.iter().find(|i| i.name == index.name);
        if kept.is_none_or(|k| index_text(to, k) != index_text(from, index)) {
            out.push(format!("DROP INDEX {};", quote_identifier(&index.name)));
        }
    }
    for index in to.indexes.iter().filter(|i| !i.system) {
        let kept = from.indexes.iter().find(|i| i.name == index.name);
        if kept.is_none_or(|k| index_text(from, k) != index_text(to, index)) {
            let unique = if index.unique { "UNIQUE " } else { "" };
            let columns = index_columns(to, &index.columns, &index.descending);
            out.push(format!("CREATE {unique}INDEX {} ON {name} ({columns});", quote_identifier(&index.name)));
        }
    }
    let owned = |side: &Side| -> Vec<ForeignKeyDef> {
        side.foreign_keys
            .iter()
            .filter(|k| k.child_schema == side.table.schema && k.child_table == side.table.name)
            .cloned()
            .collect()
    };
    let (old_keys, new_keys) = (owned(old), owned(new));
    for key in &old_keys {
        if new_keys.iter().find(|k| k.name == key.name).is_none_or(|k| k != key) {
            out.push(format!("ALTER TABLE {name} DROP CONSTRAINT {};", quote_identifier(&key.name)));
        }
    }
    for key in &new_keys {
        if old_keys.iter().find(|k| k.name == key.name).is_none_or(|k| k != key) {
            out.push(format!(
                "ALTER TABLE {name} ADD CONSTRAINT {} {};",
                quote_identifier(&key.name),
                foreign_key_clause(key)
            ));
        }
    }
    out
}

/// ref_hash returns the hash that DOLT_PATCH reports for a revision: a commit's hash, or the name of the working set
/// or staged root.
fn ref_hash(ctx: &mut Ctx<'_>, name: &str) -> Result<String> {
    match name {
        "WORKING" | "STAGED" => Ok(name.to_string()),
        _ => Ok(history::resolve(ctx.db, ctx.txn.head, name)?.to_string()),
    }
}

/// dolt_patch returns the statements that turn the tables of one revision into those of another, or only the table
/// the third argument names, schema statements before data statements for each table, as rows of records.
pub fn dolt_patch(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let args: Vec<String> = args.iter().map(|a| a.output().unwrap_or_default()).collect();
    let (from_ref, to_ref, rest) = diff_refs(ctx, &args, "dolt_patch", 2..=3, 1..=2)?;
    let (from, to) = (ref_root(ctx, &from_ref)?, ref_root(ctx, &to_ref)?);
    let (from_hash, to_hash) = (ref_hash(ctx, &from_ref)?, ref_hash(ctx, &to_ref)?);
    let mut all = sorted_deltas(ctx, &from, &to)?;
    if let Some(table) = rest.first() {
        let found = all.iter().find(|d| matches(&d.to, table)).or_else(|| all.iter().find(|d| matches(&d.from, table)));
        let exists = |ctx: &mut Ctx<'_>, root: &Root| -> Result<bool> {
            Ok(crate::dolt::procedures::table_map(ctx.db, root)?.keys().any(|(_, n)| n.eq_ignore_ascii_case(table)))
        };
        if found.is_none() && !exists(ctx, &from)? && !exists(ctx, &to)? {
            return Err(PgError::new(code::UNDEFINED_TABLE, format!("table not found: {table}")));
        }
        all = found.into_iter().cloned().collect();
    }
    let mut rows = Vec::new();
    for delta in all {
        let (old, new) = sides(ctx, &delta, &from, &to)?;
        let name = delta.to.as_ref().or(delta.from.as_ref()).map(|(n, _)| full_name(n)).unwrap_or_default();
        let mut statements: Vec<(&str, String)> =
            schema_statements(old.as_ref(), new.as_ref()).into_iter().map(|s| ("schema", s)).collect();
        if let Some(new) = &new {
            let data = data_statements(ctx, old.as_ref().map(|s| &s.table), &new.table)?;
            statements.extend(data.into_iter().map(|s| ("data", s)));
        }
        for (kind, statement) in statements {
            let order = Value::Numeric(crate::numeric::Numeric::from_i64(rows.len() as i64 + 1));
            rows.push(Value::Record(vec![
                order,
                Value::Text(from_hash.clone()),
                Value::Text(to_hash.clone()),
                Value::Text(name.clone()),
                Value::Text(kind.into()),
                Value::Text(statement),
            ]));
        }
    }
    Ok(Value::Set(rows))
}
