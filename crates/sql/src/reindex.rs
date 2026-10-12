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

//! REINDEX, ported from Postgres' ExecReindex: it checks what Postgres checks of its options and target, and leaves
//! the indexes as they are, since Doltgres keeps every index consistent with its table as it writes.

use pg_query::NodeEnum;
use pg_query::protobuf::{RangeVar, ReindexObjectType, ReindexStmt};

use crate::Outcome;
use crate::auth::Object;
use crate::error::{PgError, Result, code};
use crate::query::Ctx;

/// Relation is a relation that a statement names: its schema and name, its relkind, and whether it is a system
/// catalog or one of their indexes.
pub(crate) struct Relation {
    pub schema: String,
    pub name: String,
    pub kind: char,
    pub system: bool,
}

impl Ctx<'_> {
    /// reindex runs REINDEX.
    pub fn reindex(&mut self, stmt: &ReindexStmt) -> Result<Outcome> {
        let (mut concurrently, mut verbose, mut tablespace) = (false, false, None);
        for param in &stmt.params {
            let Some(NodeEnum::DefElem(option)) = param.node.as_ref() else { continue };
            let enabled = !matches!(option.arg.as_deref().and_then(|a| a.node.as_ref()), Some(NodeEnum::Boolean(b)) if !b.boolval)
                && !matches!(option.arg.as_deref().and_then(|a| a.node.as_ref()), Some(NodeEnum::String(s)) if matches!(s.sval.as_str(), "false" | "off"));
            match option.defname.as_str() {
                "concurrently" => concurrently = enabled,
                "verbose" => verbose = enabled,
                "tablespace" => {
                    tablespace = option.arg.as_deref().and_then(|a| match a.node.as_ref() {
                        Some(NodeEnum::String(s)) => Some(s.sval.clone()),
                        _ => None,
                    })
                }
                other => {
                    return Err(PgError {
                        position: crate::expr::position(option.location),
                        ..PgError::new(code::SYNTAX_ERROR, format!("unrecognized REINDEX option \"{other}\""))
                    });
                }
            }
        }
        if concurrently && self.session.explicit {
            return Err(PgError::new(
                code::ACTIVE_SQL_TRANSACTION,
                "REINDEX CONCURRENTLY cannot run inside a transaction block",
            ));
        }
        if let Some(name) = &tablespace {
            match name.as_str() {
                "pg_default" => {}
                "pg_global" => {
                    return Err(PgError::new(
                        code::INVALID_PARAMETER_VALUE,
                        "only shared relations can be placed in pg_global tablespace",
                    ));
                }
                _ => {
                    return Err(PgError::new(code::UNDEFINED_OBJECT, format!("tablespace \"{name}\" does not exist")));
                }
            }
        }
        let kind = ReindexObjectType::try_from(stmt.kind).unwrap_or(ReindexObjectType::Undefined);
        match kind {
            ReindexObjectType::ReindexObjectIndex | ReindexObjectType::ReindexObjectTable => {
                let relation = stmt.relation.as_ref().ok_or_else(|| PgError::internal("REINDEX without a relation"))?;
                let index = kind == ReindexObjectType::ReindexObjectIndex;
                let Some(found) = self.find_relation(relation)? else {
                    return Err(PgError::new(
                        code::UNDEFINED_TABLE,
                        format!("relation \"{}\" does not exist", shown(relation)),
                    ));
                };
                if index && found.kind != 'i' {
                    return Err(PgError::new(
                        code::WRONG_OBJECT_TYPE,
                        format!("\"{}\" is not an index", relation.relname),
                    ));
                }
                if !index && !matches!(found.kind, 'r' | 't') {
                    return Err(PgError::new(
                        code::WRONG_OBJECT_TYPE,
                        format!("\"{}\" is not a table or materialized view", relation.relname),
                    ));
                }
                if found.system && concurrently {
                    return Err(PgError::new(
                        code::FEATURE_NOT_SUPPORTED,
                        "cannot reindex system catalogs concurrently",
                    ));
                }
                if !index && !found.system {
                    self.require_owner(&Object::Table(found.schema.clone(), found.name.clone()))?;
                    let snapshot = self.snapshot()?;
                    let indexes: Vec<String> = snapshot
                        .listed()
                        .into_iter()
                        .find(|(t, _)| t.schema == found.schema && t.name == found.name)
                        .map(|(_, indexes)| indexes.into_iter().map(|i| i.name).collect())
                        .unwrap_or_default();
                    let name = &found.name;
                    if indexes.is_empty() {
                        let message = match concurrently {
                            true => format!("table \"{name}\" has no indexes that can be reindexed concurrently"),
                            false => format!("table \"{name}\" has no indexes to reindex"),
                        };
                        self.session.notice(PgError::notice("00000", message));
                    }
                    if verbose {
                        for index in indexes {
                            self.session.notice(PgError {
                                severity: "INFO",
                                detail: Some("CPU: user: 0.00 s, system: 0.00 s, elapsed: 0.00 s".into()),
                                ..PgError::new("00000", format!("index \"{index}\" was reindexed"))
                            });
                        }
                    }
                }
            }
            _ => {
                let command = match kind {
                    ReindexObjectType::ReindexObjectSchema => "REINDEX SCHEMA",
                    ReindexObjectType::ReindexObjectSystem => "REINDEX SYSTEM",
                    _ => "REINDEX DATABASE",
                };
                if self.session.explicit {
                    return Err(PgError::new(
                        code::ACTIVE_SQL_TRANSACTION,
                        format!("{command} cannot run inside a transaction block"),
                    ));
                }
                if kind == ReindexObjectType::ReindexObjectSystem && concurrently {
                    return Err(PgError::new(
                        code::FEATURE_NOT_SUPPORTED,
                        "cannot reindex system catalogs concurrently",
                    ));
                }
                if kind == ReindexObjectType::ReindexObjectSchema {
                    let schema = self.session.named_schema(&stmt.name);
                    let system = matches!(schema.as_str(), "pg_catalog" | "pg_toast" | "information_schema");
                    if !system && !self.txn.root.schemas.iter().any(|s| *s == schema.as_bytes()) {
                        return Err(PgError::new(
                            code::INVALID_SCHEMA_NAME,
                            format!("schema \"{}\" does not exist", stmt.name),
                        ));
                    }
                    self.require_owner(&Object::Schema(schema))?;
                    if system && concurrently {
                        self.session.notice(PgError {
                            severity: "WARNING",
                            ..PgError::new(
                                code::FEATURE_NOT_SUPPORTED,
                                "cannot reindex system catalogs concurrently, skipping all",
                            )
                        });
                    }
                } else {
                    if !stmt.name.is_empty() && stmt.name != self.session.database {
                        return Err(PgError::new(
                            code::FEATURE_NOT_SUPPORTED,
                            "can only reindex the currently open database",
                        ));
                    }
                    let database = self.session.database.clone();
                    self.require_owner(&Object::Database(database))?;
                }
            }
        }
        Ok(Outcome::command("REINDEX"))
    }

    /// find_relation finds the relation that a statement names, searching the system catalogs first and then the
    /// search path for an unqualified name.
    pub(crate) fn find_relation(&mut self, relation: &RangeVar) -> Result<Option<Relation>> {
        let schemas = match relation.schemaname.as_str() {
            "" => [vec!["pg_catalog".to_string()], self.session.search_path()].concat(),
            schema => vec![self.session.named_schema(schema)],
        };
        let snapshot = self.snapshot()?;
        let name = relation.relname.as_str();
        let found = |schema: &str, kind: char, system: bool| {
            Some(Relation { schema: schema.to_string(), name: name.to_string(), kind, system })
        };
        for schema in schemas {
            if let Some(kind) = crate::pgcatalog::reg::builtin_relkind(&schema, name) {
                return Ok(found(&schema, kind.chars().next().unwrap_or('r'), true));
            }
            for (table, indexes) in snapshot.listed() {
                if table.schema != schema {
                    continue;
                }
                if table.name == name {
                    return Ok(found(&schema, 'r', false));
                }
                if indexes.iter().any(|i| i.name == name) {
                    return Ok(found(&schema, 'i', false));
                }
            }
            if snapshot.views.iter().any(|v| v.schema == schema && v.name == name) {
                return Ok(found(&schema, 'v', false));
            }
            if snapshot
                .sequences
                .iter()
                .any(|s| crate::sequences::schema_and_name(s) == (schema.clone(), name.to_string()))
            {
                return Ok(found(&schema, 'S', false));
            }
        }
        Ok(None)
    }
}

/// shown returns a relation's name as an error names it.
fn shown(relation: &RangeVar) -> String {
    match relation.schemaname.as_str() {
        "" => relation.relname.clone(),
        schema => format!("{schema}.{}", relation.relname),
    }
}
