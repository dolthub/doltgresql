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

//! The parts of Postgres' optimizer/util/clauses.c that the planner calls: which functions an expression runs, and
//! which of its relations and Vars it can only be true with when they are not NULL. Doltgres' built-in operators and
//! casts are strict, as Postgres' are, and its user-defined routines are volatile unless marked otherwise, which is
//! Postgres' default.

use super::nodes::{PlannerGlobal, Relids, VarNode};
use crate::expr::Expr;

/// placeholder_expr returns the expression of a PlaceHolderVar, or None for any other expression.
fn placeholder_expr<'g>(glob: &'g PlannerGlobal, e: &Expr) -> Option<&'g Expr> {
    match e {
        Expr::Column(id) => match glob.node(*id) {
            VarNode::PlaceHolderVar(phv) => Some(&glob.placeholder(phv.phid).phexpr),
            VarNode::Var(_) => None,
        },
        _ => None,
    }
}

/// any_node reports whether a test holds for an expression or one of its descendants, descending into the
/// expressions of PlaceHolderVars, as Postgres' expression_tree_walker does.
fn any_node(glob: &PlannerGlobal, e: &Expr, test: &mut dyn FnMut(&Expr) -> bool) -> bool {
    let mut found = false;
    e.visit(&mut |x| {
        if found {
            return;
        }
        found = test(x) || placeholder_expr(glob, x).is_some_and(|p| any_node(glob, p, test));
    });
    found
}

/// contain_volatile_functions reports whether an expression calls a volatile function, as Postgres' function of the
/// same name does.
pub fn contain_volatile_functions(glob: &PlannerGlobal, e: &Expr) -> bool {
    any_node(glob, e, &mut is_volatile_node)
}

/// contain_mutable_functions reports whether an expression calls a function that is not immutable, whose result may
/// change within a statement or between statements, as Postgres' function of the same name does. Stored routines
/// count as mutable, since Doltgres does not record their volatility.
pub fn contain_mutable_functions(glob: &PlannerGlobal, e: &Expr) -> bool {
    any_node(glob, e, &mut |x| match x {
        Expr::Func(f, _) => crate::pgcatalog::is_mutable(crate::functions::function(*f).name),
        Expr::Routine(..) | Expr::Operator(..) => true,
        _ => false,
    })
}

/// is_volatile_node reports whether an expression node calls a volatile function, not counting its arguments.
pub fn is_volatile_node(e: &Expr) -> bool {
    match e {
        Expr::Func(f, _) => crate::pgcatalog::is_volatile(crate::functions::function(*f).name),
        Expr::Routine(..) | Expr::Operator(..) => true,
        _ => false,
    }
}

/// contain_subplans reports whether an expression has a subquery, as Postgres' function of the same name does.
pub fn contain_subplans(e: &Expr) -> bool {
    crate::plan::has_subquery(e)
}

/// contain_nonstrict_functions reports whether an expression can be non-NULL when an input is NULL, as Postgres'
/// function of the same name does.
pub fn contain_nonstrict_functions(glob: &PlannerGlobal, e: &Expr) -> bool {
    any_node(glob, e, &mut |x| match x {
        Expr::Column(_)
        | Expr::Const(_)
        | Expr::Param(_)
        | Expr::Outer(..)
        | Expr::Cast(..)
        | Expr::Arith(..)
        | Expr::Neg(..)
        | Expr::Compare(..)
        | Expr::Concat(..)
        | Expr::DateTime(..)
        | Expr::ArrayOp(..)
        | Expr::Field(..)
        | Expr::Subscript(..)
        | Expr::Spread(_)
        | Expr::Not(_) => false,
        Expr::Func(f, _) => !crate::functions::function(*f).strict,
        Expr::Routine(routine, _) | Expr::Operator(_, routine, ..) => !routine.strict,
        Expr::AnyArray(_, array, all) => !is_strict_saop(array, *all, false),
        _ => true,
    })
}

/// strict_args returns the arguments of a strict operator, cast, or function, which is NULL when any of them is,
/// or None for any other expression.
fn strict_args(e: &Expr) -> Option<Vec<&Expr>> {
    match e {
        Expr::Cast(inner, ..) | Expr::Neg(inner, _) => Some(vec![inner]),
        Expr::Arith(_, l, r, _)
        | Expr::Compare(_, l, r)
        | Expr::Concat(l, r)
        | Expr::DateTime(_, l, r)
        | Expr::ArrayOp(_, l, r) => Some(vec![l, r]),
        Expr::Func(f, args) if crate::functions::function(*f).strict => Some(args.iter().collect()),
        Expr::Routine(routine, args) if routine.strict => Some(args.iter().collect()),
        Expr::Operator(_, routine, l, r) if routine.strict => Some(vec![l, r]),
        _ => None,
    }
}

/// is_strict_saop reports whether a comparison with each element of an array is NULL when its left side is, as
/// Postgres' function of the same name does: always for ANY where a false result counts, and otherwise when the
/// array is a constant or ARRAY constructor with elements.
fn is_strict_saop(array: &Expr, all: bool, false_ok: bool) -> bool {
    if !all && false_ok {
        return true;
    }
    match array {
        Expr::Const(crate::types::Value::Array(a)) => !a.values.is_empty(),
        Expr::Array(_, items, false) => !items.is_empty(),
        _ => false,
    }
}

/// is_row reports whether an IS NULL test's argument is a row, which Postgres' NullTest marks argisrow.
fn is_row(e: &Expr) -> bool {
    matches!(e, Expr::Row(..))
}

/// find_nonnullable_rels returns the relations whose Vars a clause can only be true with when they are not NULL, as
/// Postgres' find_nonnullable_rels finds them.
pub fn find_nonnullable_rels(glob: &PlannerGlobal, e: &Expr) -> Relids {
    nonnullable_rels(glob, e, true)
}

/// nonnullable_rels is find_nonnullable_rels_walker, given whether the expression is at the clause's top level.
fn nonnullable_rels(glob: &PlannerGlobal, e: &Expr, top_level: bool) -> Relids {
    if let Expr::Column(id) = e {
        return match glob.node(*id) {
            VarNode::Var(var) => Relids::singleton(var.varno),
            VarNode::PlaceHolderVar(phv) => {
                let placeholder = glob.placeholder(phv.phid);
                let mut result = nonnullable_rels(glob, &placeholder.phexpr, top_level);
                if placeholder.phrels.num_members() == 1 {
                    result.add_members(&placeholder.phrels);
                }
                result
            }
        };
    }
    if let Some(args) = strict_args(e) {
        return args.into_iter().fold(Relids::new(), |relids, a| relids.union(&nonnullable_rels(glob, a, false)));
    }
    match e {
        Expr::AnyArray(l, array, all) if is_strict_saop(array, *all, top_level) => {
            nonnullable_rels(glob, l, false).union(&nonnullable_rels(glob, array, false))
        }
        Expr::And(a, b) if top_level => nonnullable_rels(glob, a, true).union(&nonnullable_rels(glob, b, true)),
        Expr::And(a, b) | Expr::Or(a, b) => {
            nonnullable_rels(glob, a, top_level).intersect(&nonnullable_rels(glob, b, top_level))
        }
        Expr::Not(inner) => nonnullable_rels(glob, inner, false),
        Expr::IsNull(inner, true) if top_level && !is_row(inner) => nonnullable_rels(glob, inner, false),
        Expr::BoolTest(inner, Some(_), false) | Expr::BoolTest(inner, None, true) if top_level => {
            nonnullable_rels(glob, inner, false)
        }
        Expr::AnySubquery(test, _, false) if top_level => nonnullable_rels(glob, test, top_level),
        _ => Relids::new(),
    }
}

/// NonnullableVars is a set of Vars by range table index and attribute, as Postgres' multibitmapsets of Vars are.
pub type NonnullableVars = Vec<(usize, usize)>;

/// find_nonnullable_vars returns the Vars that a clause can only be true with when they are not NULL, as Postgres'
/// find_nonnullable_vars finds them.
pub fn find_nonnullable_vars(glob: &PlannerGlobal, e: &Expr) -> NonnullableVars {
    nonnullable_vars(glob, e, true)
}

/// nonnullable_vars is find_nonnullable_vars_walker, given whether the expression is at the clause's top level.
fn nonnullable_vars(glob: &PlannerGlobal, e: &Expr, top_level: bool) -> NonnullableVars {
    if let Expr::Column(id) = e {
        return match glob.node(*id) {
            VarNode::Var(var) => vec![(var.varno, var.varattno)],
            VarNode::PlaceHolderVar(phv) => nonnullable_vars(glob, &glob.placeholder(phv.phid).phexpr, top_level),
        };
    }
    let union = |mut a: NonnullableVars, b: NonnullableVars| {
        for v in b {
            if !a.contains(&v) {
                a.push(v);
            }
        }
        a
    };
    if let Some(args) = strict_args(e) {
        return args.into_iter().fold(Vec::new(), |vars, a| union(vars, nonnullable_vars(glob, a, false)));
    }
    match e {
        Expr::AnyArray(l, array, all) if is_strict_saop(array, *all, top_level) => {
            union(nonnullable_vars(glob, l, false), nonnullable_vars(glob, array, false))
        }
        Expr::And(a, b) if top_level => union(nonnullable_vars(glob, a, true), nonnullable_vars(glob, b, true)),
        Expr::And(a, b) | Expr::Or(a, b) => {
            let right = nonnullable_vars(glob, b, top_level);
            nonnullable_vars(glob, a, top_level).into_iter().filter(|v| right.contains(v)).collect()
        }
        Expr::Not(inner) => nonnullable_vars(glob, inner, false),
        Expr::IsNull(inner, true) if top_level && !is_row(inner) => nonnullable_vars(glob, inner, false),
        Expr::BoolTest(inner, Some(_), false) | Expr::BoolTest(inner, None, true) if top_level => {
            nonnullable_vars(glob, inner, false)
        }
        Expr::AnySubquery(test, _, false) if top_level => nonnullable_vars(glob, test, top_level),
        _ => Vec::new(),
    }
}

/// find_forced_null_vars returns the Vars that a clause can only be true with when they are NULL, from its top-level
/// IS NULL and IS UNKNOWN tests of Vars, as Postgres' find_forced_null_vars finds them.
pub fn find_forced_null_vars(glob: &PlannerGlobal, e: &Expr) -> NonnullableVars {
    if let Some(var) = find_forced_null_var(glob, e) {
        let VarNode::Var(var) = glob.node(var) else { unreachable!("a forced NULL Var is a Var") };
        return vec![(var.varno, var.varattno)];
    }
    match e {
        Expr::And(a, b) => {
            let mut vars = find_forced_null_vars(glob, a);
            vars.extend(find_forced_null_vars(glob, b).into_iter().filter(|v| !vars.contains(v)).collect::<Vec<_>>());
            vars
        }
        _ => Vec::new(),
    }
}

/// find_forced_null_var returns the ID of the Var that a clause tests IS NULL or IS UNKNOWN, as Postgres' function of
/// the same name does.
pub fn find_forced_null_var(glob: &PlannerGlobal, e: &Expr) -> Option<usize> {
    let inner = match e {
        Expr::IsNull(inner, false) if !is_row(inner) => inner,
        Expr::BoolTest(inner, None, false) => inner,
        _ => return None,
    };
    match **inner {
        Expr::Column(id) if matches!(glob.node(id), VarNode::Var(_)) => Some(id),
        _ => None,
    }
}

/// expression_returns_set reports whether an expression calls a set-returning function, as Postgres' function of
/// the same name does.
pub fn expression_returns_set(e: &Expr) -> bool {
    let mut found = false;
    e.visit(&mut |x| found |= matches!(x, Expr::SetRef(_)));
    found
}

/// expression_returns_set_rows estimates the rows that a set-returning function returns for each input row, or one
/// for any other expression, as Postgres' function of the same name does with each function's default of 1000.
pub fn expression_returns_set_rows(_root: &super::PlannerInfo<'_, '_>, e: &Expr) -> f64 {
    match e {
        Expr::SetRef(_) => 1000.0,
        _ => 1.0,
    }
}
