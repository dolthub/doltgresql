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

/// infer_arbiter_indexes returns the unique indexes of a table that an ON CONFLICT clause's inference elements and
/// WHERE clause name as its arbiters, None for the primary key, as Postgres' function of the same name does: each
/// whose key columns are the elements' columns, whose expressions are the elements' expressions, whose collations
/// and operator classes are those that the elements give, and whose predicate the WHERE clause implies. The elements
/// and the WHERE clause read the table's columns.
pub(crate) fn infer_arbiter_indexes(
    ctx: &mut crate::query::Ctx<'_>,
    table: &crate::catalog::table::TableDef,
    elems: &[super::nodes::InferenceElem],
    arbiter_where: Option<&Expr>,
) -> crate::error::Result<Vec<Option<usize>>> {
    use crate::catalog::table::HIDDEN_BASE;
    let rules = ctx.index_rules(table)?;
    let mut glob = super::nodes::PlannerGlobal::default();
    let parse = super::query::unbind(&mut glob, ctx, crate::plan::Plan::Scan(Box::new(table.clone()), None));
    let mut root = super::new_planner_info(ctx, &mut glob, parse, 0.0, false);
    let to_vars = |root: &mut PlannerInfo<'_, '_>, exprs: Vec<&Expr>| -> Vec<Expr> {
        let conjuncts = exprs.into_iter().flat_map(crate::indexscan::conjuncts);
        conjuncts.map(|e| super::var::replace_columns(e.clone(), &mut |c| root.glob.var(1, c, Relids::new()))).collect()
    };
    let clauses = to_vars(&mut root, arbiter_where.into_iter().collect());
    let infer_attrs: std::collections::BTreeSet<usize> =
        elems.iter().filter_map(|elem| if let Expr::Column(c) = elem.expr { Some(c) } else { None }).collect();
    let infer_elems: Vec<&Expr> =
        elems.iter().map(|elem| &elem.expr).filter(|e| !matches!(e, Expr::Column(_))).collect();
    let mut candidates: Vec<(Option<usize>, &[usize], &[String])> = Vec::new();
    if !table.keyless() {
        candidates.push((None, &table.key_columns, &[]));
    }
    for (i, index) in table.indexes.iter().enumerate().filter(|(_, index)| index.unique && index.vector.is_none()) {
        candidates.push((Some(i), &index.columns, &index.op_classes));
    }
    let mut results = Vec::new();
    'indexes: for (index, columns, op_classes) in candidates {
        let indexed_attrs: std::collections::BTreeSet<usize> =
            columns.iter().copied().filter(|&c| c < HIDDEN_BASE).collect();
        if indexed_attrs != infer_attrs {
            continue;
        }
        let idx_exprs: Vec<&Expr> =
            columns.iter().filter_map(|c| c.checked_sub(HIDDEN_BASE)).map(|k| &rules.hidden()[k]).collect();
        for elem in elems {
            if !infer_collation_opclass_match(elem, table, columns, op_classes, rules.hidden()) {
                continue 'indexes;
            }
            if matches!(elem.expr, Expr::Column(_))
                || elem.infercollid.is_some()
                || elem.inferopclass.is_some()
                || idx_exprs.contains(&&elem.expr)
            {
                continue;
            }
            continue 'indexes;
        }
        if idx_exprs.iter().any(|e| !infer_elems.contains(e)) {
            continue;
        }
        let predicate = index.and_then(|i| rules.predicates()[i].as_ref());
        let predicate = to_vars(&mut root, predicate.into_iter().collect());
        if !super::predtest::predicate_implied_by(&root, &predicate, &clauses, false) {
            continue;
        }
        results.push(index);
    }
    if results.is_empty() {
        return Err(crate::error::PgError::new(
            crate::error::code::INVALID_COLUMN_REFERENCE,
            "there is no unique or exclusion constraint matching the ON CONFLICT specification",
        ));
    }
    Ok(results)
}

/// infer_collation_opclass_match reports whether an index column matches an inference element that names a
/// collation or operator class, as Postgres' function of the same name does: a column of the index with that
/// collation and operator class holds the element's column or expression.
fn infer_collation_opclass_match(
    elem: &super::nodes::InferenceElem,
    table: &crate::catalog::table::TableDef,
    columns: &[usize],
    op_classes: &[String],
    hidden: &[Expr],
) -> bool {
    if elem.infercollid.is_none() && elem.inferopclass.is_none() {
        return true;
    }
    for (natt, &column) in columns.iter().enumerate() {
        let Some(def) = table.index_column(column) else { continue };
        let opclass = match op_classes.get(natt).filter(|name| !name.is_empty()) {
            Some(name) => Some(name.clone()),
            None => crate::pgcatalog::default_operator_class(def.ty.oid).map(|class| class.name),
        };
        if elem.inferopclass.is_some() && elem.inferopclass != opclass {
            continue;
        }
        let collation = crate::expr::is_string(def.ty.oid).then(|| "default".to_string());
        if elem.infercollid.is_some() && elem.infercollid != collation {
            continue;
        }
        let holds = match column.checked_sub(crate::catalog::table::HIDDEN_BASE) {
            Some(k) => hidden[k] == elem.expr,
            None => elem.expr == Expr::Column(column),
        };
        if holds {
            return true;
        }
    }
    false
}

/// AUTOVACUUM_ANALYZE_THRESHOLD is how many rows a table must hold for Postgres' autovacuum to have analyzed it, at
/// its default autovacuum_analyze_threshold.
const AUTOVACUUM_ANALYZE_THRESHOLD: f64 = 50.0;

/// get_relation_info fills in what the planner learns of a base relation's table: its NOT NULL columns, its size, its
/// statistics, its indexes, and its foreign keys, as Postgres' function of the same name does. Doltgres' tables have
/// no inheritance, partitions, or extended statistics.
pub fn get_relation_info(root: &mut PlannerInfo<'_, '_>, relid: usize, table: &crate::catalog::table::TableDef) {
    let rows = prolly::Node::decode(table.table.primary_index.clone()).map_or(0.0, |r| r.tree_count() as f64);
    let session = &root.ctx.get_mut().session;
    let vacuumed = session.engine.vacuumed(&session.database, &table.schema, &table.name);
    let autovacuumed = rows > AUTOVACUUM_ANALYZE_THRESHOLD;
    let measured = autovacuumed || vacuumed.is_some();
    let analyzed = autovacuumed || vacuumed == Some(true);
    let stats = analyzed.then(|| crate::colstats::table_stats(root.ctx.get_mut(), table)).flatten();
    let rel = &mut root.rels[relid];
    rel.notnullattnums = (0..table.columns.len()).filter(|&c| !table.columns[c].nullable).collect();
    let data_width = match measured {
        true => 0.0,
        false => get_rel_data_width(table, &mut rel.attr_widths),
    };
    (rel.pages, rel.tuples) = super::costsize::estimate_rel_size(rows, data_width, measured);
    rel.stats = stats;
    root.rels[relid].indexlist = super::indxpath::get_relation_indexes(root, relid, table);
    get_relation_foreign_keys(root, relid, table);
}

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
/// a false or NULL constant, or with constraint_exclusion on, or partition for a member of an append relation, they
/// refute each other or its table's constraints, as Postgres' function of the same name does.
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
    match (constraint_exclusion.as_deref(), root.rels[rel].reloptkind) {
        (Some("on"), _) | (Some("partition"), RelOptKind::OtherMemberRel) => {}
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

/// get_rel_data_width returns the estimated width of a table's rows, caching each column's width, as Postgres'
/// function of the same name does for a table that was never measured, which has no statistics of its columns'
/// widths, so each takes its type's average width.
pub fn get_rel_data_width(table: &crate::catalog::table::TableDef, attr_widths: &mut [f64]) -> f64 {
    let mut tuple_width = 0.0;
    for (column, width) in table.columns.iter().zip(attr_widths.iter_mut()) {
        if *width <= 0.0 {
            *width = super::costsize::get_typavgwidth(Some(column.ty.oid), column.ty.modifier);
        }
        tuple_width += *width;
    }
    super::costsize::clamp_width_est(tuple_width)
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

/// ROUTINE_COST is the cost of a call of a user-defined routine in units of an operator's, as Postgres' default COST
/// of a function written in SQL or PL/pgSQL.
const ROUTINE_COST: f64 = 100.0;

/// add_function_cost adds the cost of a call of a function or operator to a cost, as Postgres' function of the same
/// name does from the function's procost: a built-in function's from pg_proc, and a user-defined routine's default.
/// No Doltgres function has a support function that estimates its own cost.
pub fn add_function_cost(e: &Expr, cost: &mut super::costsize::QualCost) {
    let procost = match e {
        Expr::Func(index, _) => crate::pgcatalog::function_cost(crate::functions::function(*index).name),
        _ => ROUTINE_COST,
    };
    cost.per_tuple += procost * super::costsize::CPU_OPERATOR_COST;
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
