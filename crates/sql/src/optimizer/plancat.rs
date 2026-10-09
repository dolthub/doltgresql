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

//! Postgres' optimizer/util/plancat.c: what the planner learns of a table from the catalog, its constraints, and
//! whether they prove that a relation returns no rows.

use super::PlannerInfo;
use super::nodes::{RelOptKind, Relids, RteKind};
use crate::expr::Expr;
use crate::types::Value;

/// get_relation_constraints returns the check constraints of a base relation's table, and with `include_notnull`
/// an IS NOT NULL test of each NOT NULL column, over the relation's Vars, as Postgres' function of the same name does.
/// Doltgres' tables have no inheritance or partitions, and every check constraint is valid.
pub fn get_relation_constraints(root: &mut PlannerInfo<'_, '_>, rel: usize, include_notnull: bool) -> Vec<Expr> {
    let Some(table) = root.parse.rte(rel).table().cloned() else { return Vec::new() };
    let checks = root.ctx.check_rules(&table).unwrap_or_default();
    let glob = &mut *root.glob;
    let mut result: Vec<Expr> = checks
        .into_iter()
        .map(|check| super::var::replace_columns(check, &mut |c| glob.var(rel, c, Relids::new())))
        .map(|check| check.fold(root.ctx))
        .collect();
    if include_notnull {
        for (attno, column) in table.columns.iter().enumerate() {
            if !column.nullable {
                result.push(Expr::IsNull(Box::new(root.glob.var(rel, attno, Relids::new())), true));
            }
        }
    }
    result
}

/// relation_excluded_by_constraints reports whether a base relation's restrictions prove it returns no rows: one is
/// a false or NULL constant, or with constraint_exclusion on, they refute each other or its table's constraints, as
/// Postgres' function of the same name does.
pub fn relation_excluded_by_constraints(root: &mut PlannerInfo<'_, '_>, rel: usize) -> bool {
    let baserestrictinfo = root.rels[rel].baserestrictinfo.clone();
    if baserestrictinfo.is_empty() {
        return false;
    }
    if baserestrictinfo
        .iter()
        .any(|&r| matches!(root.rinfos[r].clause, Expr::Const(Value::Null) | Expr::Const(Value::Bool(false))))
    {
        return true;
    }
    let constraint_exclusion = root.ctx.session.settings.get("constraint_exclusion");
    match constraint_exclusion.as_deref() {
        Some("on") if root.rels[rel].reloptkind == RelOptKind::BaseRel => {}
        _ => return false,
    }
    let clauses: Vec<Expr> = baserestrictinfo.iter().map(|&r| root.rinfos[r].clause.clone()).collect();
    let safe_restrictions: Vec<Expr> =
        clauses.iter().filter(|c| !super::clauses::contain_mutable_functions(root.glob, c)).cloned().collect();
    if super::predtest::predicate_refuted_by(root, &safe_restrictions, &safe_restrictions, true) {
        return true;
    }
    if !matches!(root.parse.rte(rel).kind, RteKind::Relation(..)) {
        return false;
    }
    let constraint_pred = get_relation_constraints(root, rel, true);
    let safe_constraints: Vec<Expr> =
        constraint_pred.into_iter().filter(|c| !super::clauses::contain_mutable_functions(root.glob, c)).collect();
    super::predtest::predicate_refuted_by(root, &safe_constraints, &clauses, false)
}
