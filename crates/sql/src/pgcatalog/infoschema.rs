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

//! The rows of the information_schema views.

use crate::catalog::{ColumnType, builtin_type};
use crate::error::Result;
use crate::oid as types;
use crate::pgcatalog::rows::{TableIndex, is_builtin_schema, table_indexes};
use crate::pgcatalog::{Rows, int4, text};
use crate::query::Ctx;
use crate::types::Value;

/// yes_no returns `YES` or `NO`.
fn yes_no(value: bool) -> Value {
    text(if value { "YES" } else { "NO" })
}

/// data_type returns a type's name as information_schema shows it.
fn data_type(type_oid: u32) -> String {
    if crate::array::is_array_type(type_oid) {
        return "ARRAY".into();
    }
    match type_oid {
        types::BPCHAR => "character".into(),
        types::CHAR => "\"char\"".into(),
        _ => crate::cast::type_display(type_oid).into_owned(),
    }
}

/// Lengths are the character, numeric, and datetime facets of a column type that information_schema shows.
#[derive(Default)]
struct Lengths {
    character_maximum: Option<i32>,
    character_octet: Option<i32>,
    numeric_precision: Option<i32>,
    numeric_radix: Option<i32>,
    numeric_scale: Option<i32>,
    datetime_precision: Option<i32>,
}

/// lengths returns the facets of a column type.
fn lengths(ty: ColumnType) -> Lengths {
    let modifier = (ty.modifier >= 0).then_some(ty.modifier);
    match ty.oid {
        types::VARCHAR | types::BPCHAR => {
            let max = modifier.map(|m| m - 4);
            Lengths {
                character_maximum: max,
                character_octet: Some(max.map_or(1073741824, |m| m * 4)),
                ..Lengths::default()
            }
        }
        types::TEXT => Lengths { character_octet: Some(1073741824), ..Lengths::default() },
        types::INT2 | types::INT4 | types::INT8 => Lengths {
            numeric_precision: Some(match ty.oid {
                types::INT2 => 16,
                types::INT4 => 32,
                _ => 64,
            }),
            numeric_radix: Some(2),
            numeric_scale: Some(0),
            ..Lengths::default()
        },
        types::FLOAT4 | types::FLOAT8 => Lengths {
            numeric_precision: Some(if ty.oid == types::FLOAT4 { 24 } else { 53 }),
            numeric_radix: Some(2),
            ..Lengths::default()
        },
        types::NUMERIC => Lengths {
            numeric_precision: modifier.map(|m| ((m - 4) >> 16) & 0xffff),
            numeric_radix: Some(10),
            numeric_scale: modifier.map(|m| (m - 4) & 0xffff),
            ..Lengths::default()
        },
        types::DATE => Lengths { datetime_precision: Some(0), ..Lengths::default() },
        types::TIME | types::TIMETZ | types::TIMESTAMP | types::TIMESTAMPTZ | types::INTERVAL => {
            Lengths { datetime_precision: Some(modifier.unwrap_or(6)), ..Lengths::default() }
        }
        _ => Lengths::default(),
    }
}

/// InfoColumn is a column as information_schema.columns shows it: name, type, nullability, default or generation
/// expression, and whether it is generated.
type InfoColumn = (String, ColumnType, bool, String, bool);

/// optional returns an integer, or NULL without one.
fn optional(value: Option<i32>) -> Value {
    value.map_or(Value::Null, int4)
}

impl Ctx<'_> {
    /// information_schema_rows fills the rows of an information_schema view.
    pub(super) fn information_schema_rows(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        match rows.table.name {
            "schemata" => {
                let database = self.session.display.clone();
                for schema in self.schema_names().into_iter().filter(|s| !is_builtin_schema(s)) {
                    let owner = self.session.superuser.clone();
                    rows.push(vec![
                        ("catalog_name", text(database.clone())),
                        ("schema_name", text(schema)),
                        ("schema_owner", text(owner)),
                    ]);
                }
                Ok(())
            }
            "tables" => self.information_schema_tables(rows),
            "columns" => self.information_schema_columns(rows),
            "views" => self.information_schema_views(rows),
            "sequences" => self.information_schema_sequences(rows),
            "table_constraints" => self.information_schema_table_constraints(rows),
            "key_column_usage" => self.information_schema_key_column_usage(rows),
            "referential_constraints" => self.information_schema_referential_constraints(rows),
            "triggers" => self.information_schema_triggers(rows),
            "check_constraints" => self.information_schema_check_constraints(rows),
            _ => Ok(()),
        }
    }

    /// information_schema_tables lists the user tables and views.
    fn information_schema_tables(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.display.clone();
        let mut push = |schema: &str, name: &str, view: bool, insertable: bool| {
            rows.push(vec![
                ("table_catalog", text(database.clone())),
                ("table_schema", text(schema)),
                ("table_name", text(name)),
                (
                    "table_type",
                    text(match view {
                        true => "VIEW",
                        false if schema.starts_with("pg_temp_") => "LOCAL TEMPORARY",
                        false => "BASE TABLE",
                    }),
                ),
                ("is_insertable_into", yes_no(insertable)),
                ("is_typed", yes_no(false)),
            ]);
        };
        let snapshot = self.snapshot()?;
        for (table, _) in snapshot.listed() {
            push(&table.schema, &table.name, false, true);
        }
        for view in &snapshot.views {
            let insertable = self.view_updatability(&view.schema, &view.name, &view.statement).insertable;
            push(&view.schema, &view.name, true, insertable);
        }
        Ok(())
    }

    /// information_schema_columns lists the columns of the user tables and views.
    fn information_schema_columns(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.display.clone();
        let snapshot = self.snapshot()?;
        let mut relations: Vec<(String, String, Vec<InfoColumn>)> = Vec::new();
        let mut updatable: std::collections::HashMap<(String, String), Vec<bool>> = std::collections::HashMap::new();
        for (table, _) in snapshot.listed() {
            let columns = table
                .columns
                .iter()
                .map(|c| (c.name.clone(), c.ty, c.nullable, c.default.clone(), c.generated))
                .collect();
            relations.push((table.schema.clone(), table.name.clone(), columns));
        }
        for view in &snapshot.views {
            let columns = self.view_columns(&view.schema, &view.name).unwrap_or_default();
            let updatability = self.view_updatability(&view.schema, &view.name, &view.statement);
            updatable.insert((view.schema.clone(), view.name.clone()), updatability.columns);
            relations.push((
                view.schema.clone(),
                view.name.clone(),
                columns.into_iter().map(|(n, t)| (n, t, true, String::new(), false)).collect(),
            ));
        }
        for (schema, table, columns) in relations {
            let types: Vec<(String, ColumnType)> = columns.iter().map(|(name, ty, ..)| (name.clone(), *ty)).collect();
            for (i, (name, ty, nullable, default, generated)) in columns.into_iter().enumerate() {
                let column_default = match default.is_empty() || generated {
                    true => Value::Null,
                    false => {
                        let mut analyzer = crate::ruleutils::Analyzer::new(self, types.clone());
                        let shown = analyzer.deparse(&default, Some(ty), false)?;
                        if shown == "NULL" || shown.starts_with("NULL::") { Value::Null } else { text(shown) }
                    }
                };
                let generation = match generated {
                    true => text(self.expression_definition(
                        &default,
                        super::snapshot::table_oid(&schema, &table),
                        false,
                    )?),
                    false => Value::Null,
                };
                let l = lengths(ty);
                let udt = builtin_type(ty.oid).map_or("unknown", |t| t.name);
                rows.push(vec![
                    ("table_catalog", text(database.clone())),
                    ("table_schema", text(schema.clone())),
                    ("table_name", text(table.clone())),
                    ("column_name", text(name)),
                    ("ordinal_position", int4(i as i32 + 1)),
                    ("column_default", column_default),
                    ("is_nullable", yes_no(nullable)),
                    ("data_type", text(data_type(ty.oid))),
                    ("character_maximum_length", optional(l.character_maximum)),
                    ("character_octet_length", optional(l.character_octet)),
                    ("numeric_precision", optional(l.numeric_precision)),
                    ("numeric_precision_radix", optional(l.numeric_radix)),
                    ("numeric_scale", optional(l.numeric_scale)),
                    ("datetime_precision", optional(l.datetime_precision)),
                    ("udt_catalog", text(database.clone())),
                    ("udt_schema", text("pg_catalog")),
                    ("udt_name", text(udt)),
                    ("dtd_identifier", text((i + 1).to_string())),
                    ("is_self_referencing", yes_no(false)),
                    ("is_identity", yes_no(false)),
                    ("identity_cycle", yes_no(false)),
                    ("is_generated", text(if generated { "ALWAYS" } else { "NEVER" })),
                    ("generation_expression", generation),
                    (
                        "is_updatable",
                        yes_no(
                            updatable
                                .get(&(schema.clone(), table.clone()))
                                .is_none_or(|c| c.get(i).copied().unwrap_or(false)),
                        ),
                    ),
                ]);
            }
        }
        Ok(())
    }

    /// information_schema_views lists the user views.
    fn information_schema_views(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.display.clone();
        for view in self.snapshot()?.views.clone() {
            let definition = self.view_definition(&view.statement, false, 0).unwrap_or_default();
            let updatability = self.view_updatability(&view.schema, &view.name, &view.statement);
            let check = self.view_check_option(&view.schema, &view.name, &view.statement);
            rows.push(vec![
                ("table_catalog", text(database.clone())),
                ("table_schema", text(view.schema)),
                ("table_name", text(view.name)),
                ("view_definition", text(definition)),
                ("check_option", text(check)),
                ("is_updatable", yes_no(updatability.updatable && updatability.deletable)),
                ("is_insertable_into", yes_no(updatability.insertable)),
                ("is_trigger_updatable", yes_no(false)),
                ("is_trigger_deletable", yes_no(false)),
                ("is_trigger_insertable_into", yes_no(false)),
            ]);
        }
        Ok(())
    }

    /// information_schema_sequences lists the sequences.
    fn information_schema_sequences(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.display.clone();
        for sequence in self.snapshot()?.sequences.clone() {
            let (schema, name) = crate::sequences::schema_and_name(&sequence);
            let data_type = crate::catalog::builtin_type_by_id(&sequence.data_type_id).map_or(types::INT8, |t| t.oid);
            let precision = match data_type {
                types::INT2 => 16,
                types::INT4 => 32,
                _ => 64,
            };
            rows.push(vec![
                ("sequence_catalog", text(database.clone())),
                ("sequence_schema", text(schema)),
                ("sequence_name", text(name)),
                ("data_type", text(crate::cast::type_display(data_type))),
                ("numeric_precision", int4(precision)),
                ("numeric_precision_radix", int4(2)),
                ("numeric_scale", int4(0)),
                ("start_value", text(sequence.start.to_string())),
                ("minimum_value", text(sequence.minimum.to_string())),
                ("maximum_value", text(sequence.maximum.to_string())),
                ("increment", text(sequence.increment.to_string())),
                ("cycle_option", yes_no(sequence.cycle)),
            ]);
        }
        Ok(())
    }

    /// information_schema_check_constraints lists the check constraints of the user tables, and a NOT NULL check for each
    /// of their columns that cannot be NULL.
    fn information_schema_check_constraints(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        //TODO: list the checks of domains too, which needs the deparser to print VALUE
        let database = self.session.display.clone();
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            let mut checks = Vec::new();
            for check in &table.checks {
                let columns = table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
                let clause = crate::ruleutils::Analyzer::new(self, columns).deparse(&check.expression, None, false)?;
                checks.push((check.name.clone(), clause));
            }
            for (name, number) in crate::pgcatalog::snapshot::not_null_constraints(table) {
                checks.push((name, format!("{} IS NOT NULL", table.columns[number as usize - 1].name)));
            }
            for (name, clause) in checks {
                rows.push(vec![
                    ("constraint_catalog", text(database.clone())),
                    ("constraint_schema", text(table.schema.clone())),
                    ("constraint_name", text(name)),
                    ("check_clause", text(clause)),
                ]);
            }
        }
        Ok(())
    }

    /// information_schema_table_constraints lists the constraints of the user tables.
    fn information_schema_table_constraints(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.display.clone();
        let snapshot = self.snapshot()?;
        for (table, indexes) in snapshot.listed() {
            let mut constraints: Vec<(String, &str, (bool, bool))> = Vec::new();
            for index in indexes.clone().into_iter().filter(TableIndex::constraint) {
                let kind = if index.primary { "PRIMARY KEY" } else { "UNIQUE" };
                constraints.push((index.name, kind, (index.deferrable, index.initially_deferred)));
            }
            for fk in
                snapshot.foreign_keys.iter().filter(|f| f.child_schema == table.schema && f.child_table == table.name)
            {
                constraints.push((fk.name.clone(), "FOREIGN KEY", (fk.deferrable, fk.initially_deferred)));
            }
            for check in &table.checks {
                constraints.push((check.name.clone(), "CHECK", (false, false)));
            }
            for (name, _) in crate::pgcatalog::snapshot::not_null_constraints(table) {
                constraints.push((name, "CHECK", (false, false)));
            }
            for (name, kind, (deferrable, deferred)) in constraints {
                rows.push(vec![
                    ("constraint_catalog", text(database.clone())),
                    ("constraint_schema", text(table.schema.clone())),
                    ("constraint_name", text(name)),
                    ("table_catalog", text(database.clone())),
                    ("table_schema", text(table.schema.clone())),
                    ("table_name", text(table.name.clone())),
                    ("constraint_type", text(kind)),
                    ("is_deferrable", yes_no(deferrable)),
                    ("initially_deferred", yes_no(deferred)),
                    ("enforced", yes_no(true)),
                    ("nulls_distinct", if kind == "UNIQUE" { yes_no(true) } else { Value::Null }),
                ]);
            }
        }
        Ok(())
    }

    /// information_schema_referential_constraints lists the foreign keys with the unique constraints they refer to.
    fn information_schema_referential_constraints(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.display.clone();
        let snapshot = self.snapshot()?;
        let rule = |rule: crate::foreign::Rule| match rule {
            crate::foreign::Rule::NoAction => "NO ACTION",
            crate::foreign::Rule::Restrict => "RESTRICT",
            crate::foreign::Rule::Cascade => "CASCADE",
            crate::foreign::Rule::SetNull => "SET NULL",
            crate::foreign::Rule::SetDefault => "SET DEFAULT",
        };
        for fk in &snapshot.foreign_keys {
            let unique = snapshot.table(&fk.parent_schema, &fk.parent_table).and_then(|parent| {
                table_indexes(parent)
                    .into_iter()
                    .find(|i| if fk.parent_index.is_empty() { i.primary } else { i.name == fk.parent_index })
                    .map(|i| i.name)
            });
            rows.push(vec![
                ("constraint_catalog", text(database.clone())),
                ("constraint_schema", text(fk.child_schema.clone())),
                ("constraint_name", text(fk.name.clone())),
                ("unique_constraint_catalog", text(database.clone())),
                ("unique_constraint_schema", text(fk.parent_schema.clone())),
                ("unique_constraint_name", unique.map_or(Value::Null, text)),
                ("match_option", text("NONE")),
                ("update_rule", text(rule(fk.on_update))),
                ("delete_rule", text(rule(fk.on_delete))),
            ]);
        }
        Ok(())
    }

    /// information_schema_key_column_usage lists the columns of the primary key, unique, and foreign key constraints.
    fn information_schema_key_column_usage(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.display.clone();
        let snapshot = self.snapshot()?;
        for (table, indexes) in snapshot.listed() {
            let mut keys: Vec<(String, Vec<String>, Option<Vec<usize>>)> = Vec::new();
            for index in indexes.clone().into_iter().filter(TableIndex::constraint) {
                keys.push((index.name, index.columns.iter().map(|&c| table.columns[c].name.clone()).collect(), None));
            }
            for fk in
                snapshot.foreign_keys.iter().filter(|f| f.child_schema == table.schema && f.child_table == table.name)
            {
                let parent = snapshot.table(&fk.parent_schema, &fk.parent_table);
                let positions = parent.map(|p| {
                    let key: Vec<usize> = if fk.parent_index.is_empty() {
                        p.key_columns.clone()
                    } else {
                        p.indexes
                            .iter()
                            .find(|i| i.name == fk.parent_index)
                            .map(|i| i.columns.clone())
                            .unwrap_or_default()
                    };
                    fk.parent_columns
                        .iter()
                        .map(|c| {
                            let column = p.columns.iter().position(|col| col.name == *c).unwrap_or(0);
                            key.iter().position(|&k| k == column).unwrap_or(0) + 1
                        })
                        .collect()
                });
                keys.push((fk.name.clone(), fk.child_columns.clone(), positions));
            }
            for (name, columns, positions) in keys {
                for (i, column) in columns.into_iter().enumerate() {
                    rows.push(vec![
                        ("constraint_catalog", text(database.clone())),
                        ("constraint_schema", text(table.schema.clone())),
                        ("constraint_name", text(name.clone())),
                        ("table_catalog", text(database.clone())),
                        ("table_schema", text(table.schema.clone())),
                        ("table_name", text(table.name.clone())),
                        ("column_name", text(column)),
                        ("ordinal_position", int4(i as i32 + 1)),
                        (
                            "position_in_unique_constraint",
                            positions.as_ref().map_or(Value::Null, |p| int4(p[i] as i32)),
                        ),
                    ]);
                }
            }
        }
        Ok(())
    }
}
