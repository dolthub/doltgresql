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

use doltdb::table::Table;
use doltdb::tags::{EXTENDED_KIND, auto_generate_tag};
use pg_query::NodeEnum;
use pg_query::protobuf::{ConstrType, CreateStmt};

use crate::Outcome;
use crate::catalog::table::{ColumnDef, schema_message};
use crate::error::{PgError, Result, code};
use crate::expr::{node_name, position, resolve_type_name};
use crate::query::Ctx;

/// constraint_type returns a constraint node's type.
fn constraint_type(constraint: &pg_query::protobuf::Constraint) -> ConstrType {
    ConstrType::try_from(constraint.contype).unwrap_or(ConstrType::Undefined)
}

impl Ctx<'_> {
    /// create_table runs CREATE TABLE.
    pub fn create_table(&mut self, create: &CreateStmt) -> Result<Outcome> {
        let relation = create.relation.as_ref().ok_or_else(|| PgError::internal("CREATE TABLE without a name"))?;
        let schema = if relation.schemaname.is_empty() { self.creation_schema()? } else { relation.schemaname.clone() };
        let schema = schema.as_str();
        if !self.txn.root.schemas.iter().any(|s| s == schema.as_bytes()) {
            return Err(PgError {
                position: position(relation.location),
                ..PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{schema}\" does not exist"))
            });
        }
        let name = relation.relname.as_str();
        if self.txn.root.table(self.db, schema, name)?.is_some() {
            let message = format!("relation \"{name}\" already exists");
            if create.if_not_exists {
                self.notices.push(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE TABLE"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        let mut columns: Vec<ColumnDef> = Vec::new();
        let mut primary_key: Vec<usize> = Vec::new();
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
                    other => return Err(PgError::unsupported(format!("the table constraint {other:?}"))),
                },
                _ => return Err(PgError::unsupported("this table element")),
            }
        }
        for &i in &primary_key {
            columns[i].primary_key = true;
            columns[i].nullable = false;
        }
        let mut tags = self.txn.all_tags(self.db)?;
        let mut kinds = Vec::new();
        for column in &mut columns {
            column.tag = auto_generate_tag(&tags, name, &kinds, &column.name, EXTENDED_KIND);
            tags.insert(column.tag);
            kinds.push(EXTENDED_KIND);
        }
        let value_columns: Vec<usize> = (0..columns.len()).filter(|i| !primary_key.contains(i)).collect();
        let (address, _) = Table::create(self.db, schema_message(&columns, &primary_key, &value_columns)?)?;
        self.txn.root.put_table(self.db, schema, name, Some(address))?;
        Ok(Outcome::command("CREATE TABLE"))
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
