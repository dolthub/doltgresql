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

//! Deferred constraints: the modes that SET CONSTRAINTS gives a transaction's constraints, and the foreign key and
//! unique checks that deferred constraints owe until the transaction commits or makes them immediate.

use pg_query::NodeEnum;
use pg_query::protobuf::ConstraintsSetStmt;

use crate::Outcome;
use crate::error::{PgError, Result, code};
use crate::query::Ctx;
use crate::types::Value;

/// Pending is a check that a deferred constraint owes, named by the schema, table, and name of its constraint.
#[derive(Clone, Debug, PartialEq)]
pub enum Pending {
    /// A foreign key's referencing key, whose rows must find a referenced row.
    Child(String, String, String, Vec<Value>),
    /// A foreign key's referenced key, which referencing rows must not refer to once it is gone.
    Parent(String, String, String, Vec<Value>),
    /// A unique index's values, which at most one row may have.
    Unique(String, String, String, Vec<Value>),
}

/// Deferred is a transaction's constraint modes and the checks its deferred constraints owe.
#[derive(Clone, Debug, Default)]
pub struct Deferred {
    /// The mode SET CONSTRAINTS ALL gave every deferrable constraint, where true defers.
    pub all: Option<bool>,
    /// The modes SET CONSTRAINTS gave constraints by schema and name, latest last.
    pub named: Vec<(String, String, bool)>,
    pub pending: Vec<Pending>,
}

/// ConstraintInfo is a constraint that SET CONSTRAINTS can name: its schema, its name, and whether it is
/// DEFERRABLE.
type ConstraintInfo = (String, String, bool);

impl Ctx<'_> {
    /// is_deferred reports whether a constraint is deferred now, by the modes SET CONSTRAINTS set or else as it was
    /// declared.
    pub fn is_deferred(&self, schema: &str, name: &str, deferrable: bool, initially_deferred: bool) -> bool {
        if !deferrable {
            return false;
        }
        let deferred = &self.session.deferred;
        match deferred.named.iter().rev().find(|(s, n, _)| s == schema && n == name) {
            Some((_, _, mode)) => *mode,
            None => deferred.all.unwrap_or(initially_deferred),
        }
    }

    /// defer records a check that a deferred constraint owes, once.
    pub fn defer(&mut self, pending: Pending) {
        if !self.session.deferred.pending.contains(&pending) {
            self.session.deferred.pending.push(pending);
        }
    }

    /// constraint_infos returns the constraints that SET CONSTRAINTS can name in a schema.
    fn constraint_infos(&mut self, schema: &str) -> Result<Vec<ConstraintInfo>> {
        let snapshot = self.snapshot()?;
        let mut out = Vec::new();
        for table in snapshot.tables.iter().filter(|t| t.schema == schema) {
            if !table.key_columns.is_empty() {
                out.push((schema.to_string(), table.primary_name(), table.primary.deferrable));
            }
            for index in table.indexes.iter().filter(|i| i.unique) {
                out.push((schema.to_string(), index.name.clone(), index.deferrable));
            }
            for check in &table.checks {
                out.push((schema.to_string(), check.name.clone(), false));
            }
        }
        for fk in snapshot.foreign_keys.iter().filter(|fk| fk.child_schema == schema) {
            out.push((schema.to_string(), fk.name.clone(), fk.deferrable));
        }
        for user_type in self.user_types()?.values().filter(|t| t.schema == schema) {
            if let crate::usertypes::Kind::Domain(domain) = &user_type.kind {
                out.extend(domain.checks.iter().map(|(name, _)| (schema.to_string(), name.clone(), false)));
            }
        }
        Ok(out)
    }

    /// set_constraints runs SET CONSTRAINTS, then runs the checks of the constraints it made immediate, as Postgres'
    /// AfterTriggerSetState does.
    pub fn set_constraints(&mut self, stmt: &ConstraintsSetStmt) -> Result<Outcome> {
        if !self.session.explicit {
            self.session.notice(PgError {
                severity: "WARNING",
                ..PgError::new(
                    code::NO_ACTIVE_SQL_TRANSACTION,
                    "SET CONSTRAINTS can only be used in transaction blocks",
                )
            });
        }
        if stmt.constraints.is_empty() {
            self.session.deferred.all = Some(stmt.deferred);
            self.session.deferred.named.clear();
        }
        for node in &stmt.constraints {
            let Some(NodeEnum::RangeVar(relation)) = node.node.as_ref() else { continue };
            if !relation.catalogname.is_empty() && relation.catalogname != self.session.database {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    format!(
                        "cross-database references are not implemented: \"{}.{}.{}\"",
                        relation.catalogname, relation.schemaname, relation.relname
                    ),
                ));
            }
            let schemas = if relation.schemaname.is_empty() {
                self.session.search_path()
            } else {
                if !self.txn.root.schemas.iter().any(|s| s == relation.schemaname.as_bytes()) {
                    return Err(PgError::new(
                        code::INVALID_SCHEMA_NAME,
                        format!("schema \"{}\" does not exist", relation.schemaname),
                    ));
                }
                vec![relation.schemaname.clone()]
            };
            let mut found = None;
            for schema in schemas {
                let matches: Vec<ConstraintInfo> =
                    self.constraint_infos(&schema)?.into_iter().filter(|(_, n, _)| *n == relation.relname).collect();
                if !matches.is_empty() {
                    found = Some((schema, matches));
                    break;
                }
            }
            let Some((schema, matches)) = found else {
                return Err(PgError::new(
                    code::UNDEFINED_OBJECT,
                    format!("constraint \"{}\" does not exist", relation.relname),
                ));
            };
            if stmt.deferred && matches.iter().any(|(_, _, deferrable)| !deferrable) {
                return Err(PgError::new(
                    code::WRONG_OBJECT_TYPE,
                    format!("constraint \"{}\" is not deferrable", relation.relname),
                ));
            }
            self.session.deferred.named.push((schema, relation.relname.clone(), stmt.deferred));
        }
        if !stmt.deferred {
            self.run_deferred(false)?;
        }
        Ok(Outcome::command("SET CONSTRAINTS"))
    }

    /// run_deferred runs the owed checks, in the order their changes happened: all of them when the transaction
    /// commits, and otherwise those of the constraints that are no longer deferred.
    pub fn run_deferred(&mut self, commit: bool) -> Result<()> {
        let pending = std::mem::take(&mut self.session.deferred.pending);
        let fks = self.foreign_keys()?;
        let mut kept = Vec::new();
        for check in pending {
            match &check {
                Pending::Child(schema, table, name, key) | Pending::Parent(schema, table, name, key) => {
                    let found = fks
                        .iter()
                        .find(|fk| fk.child_schema == *schema && fk.child_table == *table && fk.name == *name);
                    let Some(fk) = found else { continue };
                    if !commit && self.is_deferred(schema, name, fk.deferrable, fk.initially_deferred) {
                        kept.push(check);
                        continue;
                    }
                    let result = if matches!(check, Pending::Child(..)) {
                        self.recheck_child(fk, key)
                    } else {
                        self.recheck_parent(fk, key)
                    };
                    if let Err(err) = result {
                        self.session.deferred.pending.clear();
                        return Err(err);
                    }
                }
                Pending::Unique(schema, table, name, values) => {
                    let Some(def) = self.txn.table(self.db, schema, table)? else { continue };
                    let Some(index) = def.indexes.iter().find(|i| i.name == *name).cloned() else { continue };
                    if !commit && self.is_deferred(schema, name, index.deferrable, index.initially_deferred) {
                        kept.push(check);
                        continue;
                    }
                    let rows = crate::query::scan(self.db, &def)?;
                    let same = rows
                        .iter()
                        .filter(|row| {
                            index.columns.iter().zip(values).all(|(&c, v)| {
                                !v.is_null()
                                    && crate::plan::rows_equal(std::slice::from_ref(&row[c]), std::slice::from_ref(v))
                            })
                        })
                        .count();
                    if same > 1 {
                        self.session.deferred.pending.clear();
                        let mut row = vec![Value::Null; def.columns.len()];
                        for (&c, v) in index.columns.iter().zip(values) {
                            row[c] = v.clone();
                        }
                        return Err(crate::dml::unique_violation(&def, &index, &row));
                    }
                }
            }
        }
        self.session.deferred.pending = kept;
        Ok(())
    }
}
