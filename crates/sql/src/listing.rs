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

//! Doltgres' statements that describe and list objects, which Postgres lacks: DESCRIBE answered as psql's `\d`,
//! SHOW TABLES, SEQUENCES, SCHEMAS, DATABASES, and INDEXES as `\dt`, `\ds`, `\dn`, `\l`, and `\di`, and SHOW CREATE
//! TABLE as DOLT_PATCH writes the statement.

use pg_query::NodeEnum;
use pg_query::protobuf::RangeVar;

use crate::error::{PgError, Result, code};
use crate::parse::{Extras, Statement};
use crate::query::{Ctx, column};
use crate::types::Value;
use crate::{Column, Outcome};

/// DESCRIBE_COLUMNS are the columns of a DESCRIBE, which are those of psql's `\d` of a table.
const DESCRIBE_COLUMNS: [&str; 5] = ["Column", "Type", "Collation", "Nullable", "Default"];

/// describe_columns returns the result columns of a DESCRIBE.
pub fn describe_columns() -> Vec<Column> {
    DESCRIBE_COLUMNS.iter().map(|name| column(name.to_string(), crate::expr::typ(crate::oid::TEXT))).collect()
}

/// show_create_columns returns the result columns of a SHOW CREATE TABLE.
pub fn show_create_columns() -> Vec<Column> {
    let text = crate::expr::typ(crate::oid::TEXT);
    vec![column("Table".into(), text), column("Create Table".into(), text)]
}

impl Ctx<'_> {
    /// named_table loads the table that a DESCRIBE or SHOW CREATE TABLE names, as it is now or at its `AS OF` revision.
    fn named_table(&mut self, relation: &RangeVar, extras: &Extras) -> Result<crate::catalog::table::TableDef> {
        match (extras.as_of.first(), self.catalog_root(relation)?) {
            (Some((_, revision)), _) => self.resolve_table_as_of(relation, revision),
            (None, Some(root)) => self.resolve_table_in(relation, &root),
            (None, None) => self.resolve_table(relation),
        }
    }

    /// show_create_table runs a SHOW CREATE TABLE, returning the table's name and the CREATE TABLE statement that
    /// Doltgres' schema formatter writes for it, as DOLT_PATCH writes it without the semicolon.
    pub fn show_create_table(&mut self, relation: &RangeVar, extras: &Extras) -> Result<Outcome> {
        let table = self.named_table(relation, extras)?;
        let foreign_keys = crate::foreign::load(self.db, &self.txn.root.clone())?;
        let statement = crate::dolt::patch::create_table_statement(&table, &foreign_keys);
        Ok(Outcome::Rows {
            columns: show_create_columns(),
            rows: vec![vec![Value::Text(table.name), Value::Text(statement.trim_end_matches(';').to_string())]],
            tag: "SELECT 1".into(),
        })
    }

    /// describe_table runs a DESCRIBE of a table, as it is now or at its `AS OF` revision, returning a row for each
    /// column with the cells of psql's `\d`, or of another relation, such as a view or a Dolt system table, with the
    /// names and types of its columns.
    pub fn describe_table(&mut self, relation: &RangeVar, extras: &Extras) -> Result<Outcome> {
        let table = match self.named_table(relation, extras) {
            Ok(table) => table,
            Err(err) => {
                let parts = [&relation.catalogname, &relation.schemaname, &relation.relname];
                let name: Vec<String> =
                    parts.iter().filter(|p| !p.is_empty()).map(|p| crate::engine::quote_identifier(p)).collect();
                let Ok(Some(Statement::Postgres { node: NodeEnum::SelectStmt(select), .. })) =
                    crate::parse::parse(&format!("TABLE {}", name.join("."))).map(|s| s.into_iter().next())
                else {
                    return Err(err);
                };
                let Ok(query) = crate::plan::Planner { ctx: self, outer: Vec::new() }.plan_query(&select) else {
                    return Err(err);
                };
                let rows: Vec<Vec<Value>> = query
                    .columns
                    .iter()
                    .map(|c| {
                        let type_name = crate::cast::format_type(c.type_oid, Some(c.type_modifier));
                        vec![
                            Value::Text(c.name.clone()),
                            Value::Text(type_name.unwrap_or_else(|| "???".into())),
                            Value::Null,
                            Value::Text(String::new()),
                            Value::Null,
                        ]
                    })
                    .collect();
                let tag = format!("SELECT {}", rows.len());
                return Ok(Outcome::Rows { columns: describe_columns(), rows, tag });
            }
        };
        let columns: Vec<(String, crate::catalog::ColumnType)> =
            table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
        let mut rows = Vec::with_capacity(table.columns.len());
        for column in &table.columns {
            let default = match column.default.is_empty() {
                true => Value::Null,
                false => {
                    let mut analyzer = crate::ruleutils::Analyzer::new(self, columns.clone());
                    let expression = analyzer.deparse(&column.default, Some(column.ty), true)?;
                    Value::Text(match column.generated {
                        true => format!("generated always as ({expression}) stored"),
                        false => expression,
                    })
                }
            };
            rows.push(vec![
                Value::Text(column.name.clone()),
                Value::Text(
                    crate::cast::format_type(column.ty.oid, Some(column.ty.modifier)).unwrap_or_else(|| "???".into()),
                ),
                Value::Null,
                Value::Text(if column.nullable { "" } else { "not null" }.into()),
                default,
            ]);
        }
        let tag = format!("SELECT {}", rows.len());
        Ok(Outcome::Rows { columns: describe_columns(), rows, tag })
    }
}

/// USER_SCHEMAS is the condition of psql's listings that leaves out the system schemas of a `pg_namespace` `n`.
const USER_SCHEMAS: &str =
    "n.nspname <> 'pg_catalog' AND n.nspname !~ '^pg_toast' AND n.nspname <> 'information_schema'";

impl Ctx<'_> {
    /// list_objects runs a `SHOW` of tables, sequences, schemas, databases, or a table's indexes, returning the rows
    /// of psql's `\dt`, `\ds`, `\dn`, `\l`, or `\di` from the query psql runs for it.
    pub fn list_objects(&mut self, kind: &str, from: &[String]) -> Result<Outcome> {
        let query = self.plan_listing(kind, from)?;
        let rows = query.plan.run(self)?;
        let tag = format!("SELECT {}", rows.len());
        Ok(Outcome::Rows { columns: query.columns, rows, tag })
    }

    /// plan_listing plans the query that psql runs for a `SHOW` of tables, sequences, schemas, databases, or a
    /// table's indexes.
    pub fn plan_listing(&mut self, kind: &str, from: &[String]) -> Result<crate::plan::Query> {
        let quote = crate::pgcatalog::definitions::quote_literal;
        let relations = |kinds: &str, filter: &str| {
            format!(
                "SELECT n.nspname AS \"Schema\", c.relname AS \"Name\", CASE c.relkind WHEN 'r' THEN 'table' WHEN 'S' \
                 THEN 'sequence' WHEN 'p' THEN 'partitioned table' END AS \"Type\", \
                 pg_catalog.pg_get_userbyid(c.relowner) AS \"Owner\" FROM pg_catalog.pg_class c LEFT JOIN \
                 pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE c.relkind IN ({kinds}) AND {filter} \
                 ORDER BY 1, 2"
            )
        };
        let sql = match kind {
            "tables" => match from {
                [] => relations("'r', 'p'", &format!("{USER_SCHEMAS} AND pg_catalog.pg_table_is_visible(c.oid)")),
                [schema] | [_, schema] => {
                    self.check_database(&from[..from.len() - 1], from)?;
                    if !self.txn.root.schemas.iter().any(|s| s == schema.as_bytes()) {
                        return Err(PgError::new(
                            code::INVALID_SCHEMA_NAME,
                            format!("schema \"{schema}\" does not exist"),
                        ));
                    }
                    relations("'r', 'p'", &format!("n.nspname = {}", quote(schema)))
                }
                _ => return Err(cross_database(from)),
            },
            "sequences" => {
                self.check_database(from, from)?;
                relations("'S'", USER_SCHEMAS)
            }
            "schemas" => {
                self.check_database(from, from)?;
                "SELECT n.nspname AS \"Name\", pg_catalog.pg_get_userbyid(n.nspowner) AS \"Owner\" FROM \
                 pg_catalog.pg_namespace n WHERE n.nspname !~ '^pg_' AND n.nspname <> 'information_schema' ORDER BY 1"
                    .to_string()
            }
            "databases" => "SELECT d.datname AS \"Name\", pg_catalog.pg_get_userbyid(d.datdba) AS \"Owner\", \
                 pg_catalog.pg_encoding_to_char(d.encoding) AS \"Encoding\", d.datcollate AS \"Collate\", d.datctype \
                 AS \"Ctype\", d.daticulocale AS \"ICU Locale\", CASE d.datlocprovider WHEN 'c' THEN 'libc' WHEN 'i' \
                 THEN 'icu' END AS \"Locale Provider\", NULL::text AS \"Access privileges\" FROM \
                 pg_catalog.pg_database d ORDER BY 1"
                .to_string(),
            _ => {
                if from.len() == 3 {
                    self.check_database(&from[..1], from)?;
                }
                let table: Vec<String> =
                    from.iter().rev().take(2).rev().map(|p| crate::engine::quote_identifier(p)).collect();
                format!(
                    "SELECT n.nspname AS \"Schema\", c.relname AS \"Name\", 'index' AS \"Type\", \
                     pg_catalog.pg_get_userbyid(c.relowner) AS \"Owner\", c2.relname AS \"Table\" FROM \
                     pg_catalog.pg_class c LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace LEFT JOIN \
                     pg_catalog.pg_index i ON i.indexrelid = c.oid LEFT JOIN pg_catalog.pg_class c2 ON i.indrelid = \
                     c2.oid WHERE c.relkind IN ('i', 'I') AND i.indrelid = {}::pg_catalog.regclass ORDER BY 1, 2",
                    quote(&table.join("."))
                )
            }
        };
        let Some(Statement::Postgres { node: NodeEnum::SelectStmt(select), .. }) =
            crate::parse::parse(&sql)?.into_iter().next()
        else {
            return Err(PgError::internal("a listing query is not a SELECT"));
        };
        crate::plan::Planner { ctx: self, outer: Vec::new() }.plan_query(&select)
    }

    /// check_database fails unless a listing's database, the first of the names when there is one, is the session's
    /// database, since Postgres reads no other database's catalogs.
    fn check_database(&self, database: &[String], from: &[String]) -> Result<()> {
        match database.first() {
            None => Ok(()),
            Some(name) if *name == self.session.database => Ok(()),
            Some(name) if self.session.engine.database_exists(name) => Err(cross_database(from)),
            Some(name) => Err(PgError::new(code::INVALID_CATALOG_NAME, format!("database \"{name}\" does not exist"))),
        }
    }
}

/// cross_database returns Postgres' error for a name in another database.
fn cross_database(from: &[String]) -> PgError {
    PgError::new(
        code::FEATURE_NOT_SUPPORTED,
        format!("cross-database references are not implemented: \"{}\"", from.join(".")),
    )
}
