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
use crate::pgcatalog::rows::{is_builtin_schema, table_indexes};
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

/// InfoColumn is a column as information_schema.columns shows it: name, type, nullability, and default.
type InfoColumn = (String, ColumnType, bool, String);

/// optional returns an integer, or NULL without one.
fn optional(value: Option<i32>) -> Value {
    value.map_or(Value::Null, int4)
}

impl Ctx<'_> {
    /// information_schema_rows fills the rows of an information_schema view.
    pub(super) fn information_schema_rows(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        match rows.table.name {
            "schemata" => {
                let database = self.session.database.clone();
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
            _ => Ok(()),
        }
    }

    /// information_schema_tables lists the user tables and views.
    fn information_schema_tables(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.database.clone();
        let mut push = |schema: &str, name: &str, view: bool| {
            rows.push(vec![
                ("table_catalog", text(database.clone())),
                ("table_schema", text(schema)),
                ("table_name", text(name)),
                ("table_type", text(if view { "VIEW" } else { "BASE TABLE" })),
                ("is_insertable_into", yes_no(!view || schema != "pg_catalog")),
                ("is_typed", yes_no(false)),
            ]);
        };
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            push(&table.schema, &table.name, false);
        }
        for view in &snapshot.views {
            push(&view.schema, &view.name, true);
        }
        Ok(())
    }

    /// information_schema_columns lists the columns of the user tables and views.
    fn information_schema_columns(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.database.clone();
        let snapshot = self.snapshot()?;
        let mut relations: Vec<(String, String, Vec<InfoColumn>)> = Vec::new();
        for table in &snapshot.tables {
            let columns = table.columns.iter().map(|c| (c.name.clone(), c.ty, c.nullable, c.default.clone())).collect();
            relations.push((table.schema.clone(), table.name.clone(), columns));
        }
        for view in &snapshot.views {
            let columns = self.view_columns(&view.schema, &view.name).unwrap_or_default();
            relations.push((
                view.schema.clone(),
                view.name.clone(),
                columns.into_iter().map(|(n, t)| (n, t, true, String::new())).collect(),
            ));
        }
        for (schema, table, columns) in relations {
            for (i, (name, ty, nullable, default)) in columns.into_iter().enumerate() {
                let l = lengths(ty);
                let udt = builtin_type(ty.oid).map_or("unknown", |t| t.name);
                rows.push(vec![
                    ("table_catalog", text(database.clone())),
                    ("table_schema", text(schema.clone())),
                    ("table_name", text(table.clone())),
                    ("column_name", text(name)),
                    ("ordinal_position", int4(i as i32 + 1)),
                    ("column_default", if default.is_empty() { Value::Null } else { text(default) }),
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
                    ("is_generated", text("NEVER")),
                    ("is_updatable", yes_no(true)),
                ]);
            }
        }
        Ok(())
    }

    /// information_schema_views lists the user views.
    fn information_schema_views(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.database.clone();
        for view in self.snapshot()?.views {
            let definition = crate::views::view_definition(&view.statement).unwrap_or_default();
            rows.push(vec![
                ("table_catalog", text(database.clone())),
                ("table_schema", text(view.schema)),
                ("table_name", text(view.name)),
                ("view_definition", text(definition)),
                ("check_option", text("NONE")),
                ("is_updatable", yes_no(true)),
                ("is_insertable_into", yes_no(true)),
                ("is_trigger_updatable", yes_no(false)),
                ("is_trigger_deletable", yes_no(false)),
                ("is_trigger_insertable_into", yes_no(false)),
            ]);
        }
        Ok(())
    }

    /// information_schema_sequences lists the sequences.
    fn information_schema_sequences(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.database.clone();
        for sequence in self.snapshot()?.sequences {
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

    /// information_schema_table_constraints lists the constraints of the user tables.
    fn information_schema_table_constraints(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.database.clone();
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            let mut constraints: Vec<(String, &str)> = Vec::new();
            for index in table_indexes(table).into_iter().filter(|i| i.unique) {
                constraints.push((index.name, if index.primary { "PRIMARY KEY" } else { "UNIQUE" }));
            }
            for fk in
                snapshot.foreign_keys.iter().filter(|f| f.child_schema == table.schema && f.child_table == table.name)
            {
                constraints.push((fk.name.clone(), "FOREIGN KEY"));
            }
            for check in &table.checks {
                constraints.push((check.name.clone(), "CHECK"));
            }
            let relation = crate::pgcatalog::snapshot::table_oid(&table.schema, &table.name);
            let namespace = crate::pgcatalog::snapshot::namespace_oid(&table.schema);
            for (i, _) in table.columns.iter().enumerate().filter(|(_, c)| !c.nullable) {
                constraints.push((format!("{namespace}_{relation}_{}_not_null", i + 1), "CHECK"));
            }
            for (name, kind) in constraints {
                rows.push(vec![
                    ("constraint_catalog", text(database.clone())),
                    ("constraint_schema", text(table.schema.clone())),
                    ("constraint_name", text(name)),
                    ("table_catalog", text(database.clone())),
                    ("table_schema", text(table.schema.clone())),
                    ("table_name", text(table.name.clone())),
                    ("constraint_type", text(kind)),
                    ("is_deferrable", yes_no(false)),
                    ("initially_deferred", yes_no(false)),
                    ("enforced", yes_no(true)),
                    ("nulls_distinct", if kind == "UNIQUE" { yes_no(true) } else { Value::Null }),
                ]);
            }
        }
        Ok(())
    }

    /// information_schema_key_column_usage lists the columns of the primary key, unique, and foreign key constraints.
    fn information_schema_key_column_usage(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.database.clone();
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            let mut keys: Vec<(String, Vec<String>, Option<Vec<usize>>)> = Vec::new();
            for index in table_indexes(table).into_iter().filter(|i| i.unique) {
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
