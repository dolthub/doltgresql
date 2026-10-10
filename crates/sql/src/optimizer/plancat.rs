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
    let checks = root.ctx.get_mut().check_rules(&table).unwrap_or_default();
    let glob = &mut *root.glob;
    let mut result: Vec<Expr> = checks
        .into_iter()
        .map(|check| super::var::replace_columns(check, &mut |c| glob.var(rel, c, Relids::new())))
        .map(|check| check.fold(root.ctx.get_mut()))
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
    let constraint_exclusion = root.ctx.get_mut().session.settings.get("constraint_exclusion");
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

/// get_relation_foreign_keys adds each foreign key of a base relation's table that references the table of another
/// of the query's relations to the query's foreign keys, as Postgres' function of the same name does.
pub fn get_relation_foreign_keys(root: &mut PlannerInfo<'_, '_>, rel: usize, table: &crate::catalog::table::TableDef) {
    if root.rels[rel].reloptkind != super::nodes::RelOptKind::BaseRel || root.parse.rtable.len() < 2 {
        return;
    }
    let txn_root = root.ctx.get_mut().txn.root.clone();
    let Ok(fkeys) = crate::foreign::load(root.ctx.get_mut().db, &txn_root) else { return };
    let column =
        |table: &crate::catalog::table::TableDef, name: &String| table.columns.iter().position(|c| c.name == *name);
    for fk in fkeys.iter().filter(|fk| fk.child_schema == table.schema && fk.child_table == table.name) {
        let Some(conkey) = fk.child_columns.iter().map(|name| column(table, name)).collect::<Option<Vec<usize>>>()
        else {
            continue;
        };
        for rti in 1..=root.parse.rtable.len() {
            let Some(parent) = root.parse.rte(rti).table() else { continue };
            if parent.schema != fk.parent_schema || parent.name != fk.parent_table || rti == rel {
                continue;
            }
            let Some(confkey) =
                fk.parent_columns.iter().map(|name| column(parent, name)).collect::<Option<Vec<usize>>>()
            else {
                continue;
            };
            let nkeys = conkey.len();
            root.fkey_list.push(super::nodes::ForeignKeyOptInfo {
                con_relid: rel,
                ref_relid: rti,
                conkey: conkey.clone(),
                confkey,
                eclass: vec![None; nkeys],
                fk_eclass_member: vec![None; nkeys],
                rinfos: vec![Vec::new(); nkeys],
                ..Default::default()
            });
        }
    }
}

/// get_function_rows returns the rows that a call of a set-returning function returns, as Postgres' function of the
/// same name finds them: from the support functions of generate_series and unnest when their arguments are constants,
/// and otherwise from the function's prorows.
pub fn get_function_rows(name: &str, args: &[crate::expr::Expr], prorows: f64) -> f64 {
    use crate::expr::Expr;
    use crate::types::Value;
    let number = |e: &Expr| match e {
        Expr::Const(Value::Int2(v)) => Some(f64::from(*v)),
        Expr::Const(Value::Int4(v)) => Some(f64::from(*v)),
        Expr::Const(Value::Int8(v)) => Some(*v as f64),
        Expr::Const(Value::Numeric(n)) => Some(n.to_f64()),
        _ => None,
    };
    match (name, args) {
        ("generate_series", [start, finish, step @ ..]) if step.len() <= 1 => {
            if args.iter().any(|a| matches!(a, Expr::Const(Value::Null))) {
                return 0.0;
            }
            let step = match step.first() {
                Some(step) => number(step),
                None => Some(1.0),
            };
            match (number(start), number(finish), step) {
                (Some(start), Some(finish), Some(step)) if step != 0.0 => ((finish - start + step) / step).floor(),
                _ => prorows,
            }
        }
        ("unnest", [array]) => super::selfuncs::estimate_array_length(array),
        _ => prorows,
    }
}
