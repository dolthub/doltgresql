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

//! The parts of Postgres' optimizer/util/clauses.c that the planner calls: which functions an expression runs,
//! which of its relations and Vars it can only be true with when they are not NULL, and simplifying its constant
//! parts. Doltgres' built-in operators and casts are strict, as Postgres' are, and its user-defined routines are
//! volatile unless marked otherwise, which is Postgres' default.

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
/// for any other expression, as Postgres' function of the same name does.
pub fn expression_returns_set_rows(root: &super::PlannerInfo<'_, '_>, e: &Expr) -> f64 {
    match e {
        Expr::SetRef(i) => root.parse.target_srfs.get(*i).map_or(1000.0, |srf| expression_returns_set_rows(root, srf)),
        Expr::Func(index, args) => match crate::pgcatalog::function_rows(crate::functions::function(*index).name) {
            Some(prorows) => {
                let rows = super::plancat::get_function_rows(crate::functions::function(*index).name, args, prorows);
                super::costsize::clamp_row_est(rows)
            }
            None => 1.0,
        },
        Expr::Routine(routine, _) if routine.set_of => 1000.0,
        _ => 1.0,
    }
}

/// EvalConstContext is what eval_const_expressions works with, as Postgres' eval_const_expressions_context holds it:
/// the context to evaluate expressions in, and whether stable functions may be evaluated too, for estimates.
struct EvalConstContext<'c, 'a> {
    ctx: &'c mut crate::query::Ctx<'a>,
    estimate: bool,
}

/// eval_const_expressions simplifies an expression, as Postgres' function of the same name does: it evaluates the
/// parts whose inputs are constants and whose functions are immutable, makes a strict call with a NULL argument NULL,
/// drops the constant arguments of ANDs and ORs that do not decide them, pushes NOTs down, compares booleans
/// directly, and drops the CASE branches and COALESCE arguments that cannot be reached. A part whose evaluation fails
/// stays, so that its error waits until the query runs it.
pub fn eval_const_expressions(ctx: &mut crate::query::Ctx<'_>, e: Expr) -> Expr {
    eval_const_expressions_mutator(&mut EvalConstContext { ctx, estimate: false }, e)
}

/// estimate_expression_value simplifies an expression as eval_const_expressions does, also evaluating its stable
/// functions and casts, for estimating what it will be, as Postgres' function of the same name does.
pub fn estimate_expression_value(root: &super::PlannerInfo<'_, '_>, e: Expr) -> Expr {
    let mut ctx = root.ctx.borrow_mut();
    eval_const_expressions_mutator(&mut EvalConstContext { ctx: &mut ctx, estimate: true }, e)
}

/// eval_const_expressions_mutator is eval_const_expressions for each kind of expression, as Postgres' function of the
/// same name is.
fn eval_const_expressions_mutator(cx: &mut EvalConstContext<'_, '_>, e: Expr) -> Expr {
    use crate::types::Value;
    match e {
        Expr::Func(index, args) => {
            let args = args.into_iter().map(|a| eval_const_expressions_mutator(cx, a)).collect();
            let function = crate::functions::function(index);
            let safe = match crate::pgcatalog::is_mutable(function.name) {
                false => true,
                true => cx.estimate && !crate::pgcatalog::is_volatile(function.name),
            };
            simplify_function(cx, Expr::Func(index, args), function.strict, safe)
        }
        Expr::Routine(..) | Expr::Operator(..) => {
            let strict = match &e {
                Expr::Routine(routine, _) | Expr::Operator(_, routine, ..) => routine.strict,
                _ => unreachable!("a routine call"),
            };
            let e = e.map_children(&mut |c| eval_const_expressions_mutator(cx, c));
            simplify_function(cx, e, strict, false)
        }
        Expr::Arith(..) | Expr::Neg(..) => {
            let e = e.map_children(&mut |c| eval_const_expressions_mutator(cx, c));
            simplify_function(cx, e, true, true)
        }
        Expr::Compare(..) => {
            let e = e.map_children(&mut |c| eval_const_expressions_mutator(cx, c));
            let e = simplify_function(cx, e, true, true);
            match e {
                Expr::Compare(op @ (crate::expr::CmpOp::Eq | crate::expr::CmpOp::Ne), l, r) => {
                    simplify_boolean_equality(op, *l, *r)
                }
                other => other,
            }
        }
        Expr::Cast(..) => {
            let e = e.map_children(&mut |c| eval_const_expressions_mutator(cx, c));
            let safe = match &e {
                Expr::Cast(inner, ty, _) => cast_is_immutable(inner, ty.oid) || cx.estimate,
                _ => unreachable!("a cast"),
            };
            simplify_function(cx, e, true, safe)
        }
        Expr::Concat(..) | Expr::ArrayOp(..) | Expr::DistinctFrom(..) => {
            let e = e.map_children(&mut |c| eval_const_expressions_mutator(cx, c));
            simplify_function(cx, e, false, true)
        }
        Expr::DateTime(..) => {
            let e = e.map_children(&mut |c| eval_const_expressions_mutator(cx, c));
            let safe = cx.estimate;
            simplify_function(cx, e, false, safe)
        }
        Expr::NullIf(value, test) => {
            let (value, test) = (eval_const_expressions_mutator(cx, *value), eval_const_expressions_mutator(cx, *test));
            if matches!(value, Expr::Const(Value::Null)) || matches!(test, Expr::Const(Value::Null)) {
                return value;
            }
            simplify_function(cx, Expr::NullIf(Box::new(value), Box::new(test)), false, true)
        }
        Expr::AnyArray(..) => {
            let e = e.map_children(&mut |c| eval_const_expressions_mutator(cx, c));
            let mut all_const = true;
            e.visit_children(&mut |c| all_const &= const_except_element(c));
            match all_const && !contain_volatile_node(&e) {
                true => evaluate_expr(cx, e),
                false => e,
            }
        }
        Expr::Or(..) => {
            let args: Vec<Expr> = super::restrictinfo::or_args(&e).into_iter().cloned().collect();
            let (newargs, have_null, force_true) = simplify_or_arguments(cx, args);
            if force_true {
                return Expr::Const(Value::Bool(true));
            }
            let mut newargs = newargs;
            if have_null {
                newargs.push(Expr::Const(Value::Null));
            }
            match newargs.len() {
                0 => Expr::Const(Value::Bool(false)),
                _ => super::prepqual::make_orclause(newargs),
            }
        }
        Expr::And(..) => {
            let args: Vec<Expr> = super::restrictinfo::and_args(&e).into_iter().cloned().collect();
            let (newargs, have_null, force_false) = simplify_and_arguments(cx, args);
            if force_false {
                return Expr::Const(Value::Bool(false));
            }
            let mut newargs = newargs;
            if have_null {
                newargs.push(Expr::Const(Value::Null));
            }
            match newargs.len() {
                0 => Expr::Const(Value::Bool(true)),
                _ => super::prepqual::make_andclause(newargs),
            }
        }
        Expr::Not(arg) => super::prepqual::negate_clause(eval_const_expressions_mutator(cx, *arg)),
        Expr::SubPlan(_)
        | Expr::AlternativeSubPlan(_)
        | Expr::Exists(_)
        | Expr::Scalar(_)
        | Expr::ArraySubquery(..) => e,
        Expr::Case(whens, default) => {
            let mut newargs = Vec::new();
            let mut defresult = None;
            for (cond, result) in whens {
                let casecond = eval_const_expressions_mutator(cx, cond);
                match casecond {
                    Expr::Const(Value::Bool(true)) => {
                        defresult = Some(eval_const_expressions_mutator(cx, result));
                        break;
                    }
                    Expr::Const(_) => continue,
                    casecond => newargs.push((casecond, eval_const_expressions_mutator(cx, result))),
                }
            }
            let defresult = match defresult {
                Some(defresult) => defresult,
                None => eval_const_expressions_mutator(cx, *default),
            };
            match newargs.is_empty() {
                true => defresult,
                false => Expr::Case(newargs, Box::new(defresult)),
            }
        }
        Expr::Coalesce(args) => {
            let mut newargs = Vec::new();
            for arg in args {
                match eval_const_expressions_mutator(cx, arg) {
                    Expr::Const(Value::Null) => continue,
                    e @ Expr::Const(_) => {
                        if newargs.is_empty() {
                            return e;
                        }
                        newargs.push(e);
                        break;
                    }
                    e => newargs.push(e),
                }
            }
            match newargs.is_empty() {
                true => Expr::Const(Value::Null),
                false => Expr::Coalesce(newargs),
            }
        }
        Expr::MinMax(..) | Expr::Array(..) | Expr::Row(..) | Expr::Subscript(..) | Expr::Field(..) => {
            let e = e.map_children(&mut |c| eval_const_expressions_mutator(cx, c));
            let mut all_const = true;
            e.visit_children(&mut |c| all_const &= matches!(c, Expr::Const(_)));
            match all_const {
                true => evaluate_expr(cx, e),
                false => e,
            }
        }
        Expr::IsNull(arg, negated) => match eval_const_expressions_mutator(cx, *arg) {
            Expr::Row(args, _) => {
                let mut newargs = Vec::new();
                for relem in args {
                    match relem {
                        Expr::Const(Value::Null) if negated => return Expr::Const(Value::Bool(false)),
                        Expr::Const(_) if !negated => return Expr::Const(Value::Bool(false)),
                        Expr::Const(_) => continue,
                        relem => newargs.push(Expr::IsNull(Box::new(relem), negated)),
                    }
                }
                match newargs.is_empty() {
                    true => Expr::Const(Value::Bool(true)),
                    false => super::prepqual::make_andclause(newargs),
                }
            }
            Expr::Const(value) => Expr::Const(Value::Bool(matches!(value, Value::Null) != negated)),
            arg => Expr::IsNull(Box::new(arg), negated),
        },
        Expr::BoolTest(arg, test, negated) => match eval_const_expressions_mutator(cx, *arg) {
            Expr::Const(value) => {
                let result = match (test, value) {
                    (Some(want), Value::Bool(b)) => b == want,
                    (Some(_), _) => false,
                    (None, value) => matches!(value, Value::Null),
                };
                Expr::Const(Value::Bool(result != negated))
            }
            arg => Expr::BoolTest(Box::new(arg), test, negated),
        },
        other => other.map_children(&mut |c| eval_const_expressions_mutator(cx, c)),
    }
}

/// simplify_function returns a call or operation with its arguments simplified already as NULL when it is strict and
/// an argument is NULL, or as its value when every argument is a constant and evaluating it is safe, as Postgres'
/// simplify_function and evaluate_function do.
fn simplify_function(cx: &mut EvalConstContext<'_, '_>, e: Expr, strict: bool, safe: bool) -> Expr {
    let (mut has_null_input, mut has_nonconst_input) = (false, false);
    e.visit_children(&mut |c| match c {
        Expr::Const(crate::types::Value::Null) => has_null_input = true,
        Expr::Const(_) => {}
        _ => has_nonconst_input = true,
    });
    if strict && has_null_input {
        return Expr::Const(crate::types::Value::Null);
    }
    match !has_nonconst_input && safe {
        true => evaluate_expr(cx, e),
        false => e,
    }
}

/// evaluate_expr returns an expression's value as a constant, or the expression itself when evaluating it fails, as
/// Postgres' function of the same name evaluates it at planning.
fn evaluate_expr(cx: &mut EvalConstContext<'_, '_>, e: Expr) -> Expr {
    match e.eval(cx.ctx, &[]) {
        Ok(value) => Expr::Const(value),
        Err(_) => e,
    }
}

/// simplify_or_arguments returns an OR's arguments simplified, without those that are FALSE, and whether one was NULL
/// and whether one was TRUE, which makes the OR TRUE, as Postgres' function of the same name does.
fn simplify_or_arguments(cx: &mut EvalConstContext<'_, '_>, args: Vec<Expr>) -> (Vec<Expr>, bool, bool) {
    use crate::types::Value;
    let (mut newargs, mut have_null) = (Vec::new(), false);
    let mut unprocessed_args: std::collections::VecDeque<Expr> = args.into();
    while let Some(arg) = unprocessed_args.pop_front() {
        let arg = eval_const_expressions_mutator(cx, arg);
        match arg {
            Expr::Or(..) => {
                let subargs: Vec<Expr> = super::restrictinfo::or_args(&arg).into_iter().cloned().collect();
                subargs.into_iter().rev().for_each(|a| unprocessed_args.push_front(a));
            }
            Expr::Const(Value::Null) => have_null = true,
            Expr::Const(Value::Bool(true)) => return (Vec::new(), have_null, true),
            Expr::Const(Value::Bool(false)) => {}
            arg => newargs.push(arg),
        }
    }
    (newargs, have_null, false)
}

/// simplify_and_arguments returns an AND's arguments simplified, without those that are TRUE, and whether one was
/// NULL and whether one was FALSE, which makes the AND FALSE, as Postgres' function of the same name does.
fn simplify_and_arguments(cx: &mut EvalConstContext<'_, '_>, args: Vec<Expr>) -> (Vec<Expr>, bool, bool) {
    use crate::types::Value;
    let (mut newargs, mut have_null) = (Vec::new(), false);
    let mut unprocessed_args: std::collections::VecDeque<Expr> = args.into();
    while let Some(arg) = unprocessed_args.pop_front() {
        let arg = eval_const_expressions_mutator(cx, arg);
        match arg {
            Expr::And(..) => {
                let subargs: Vec<Expr> = super::restrictinfo::and_args(&arg).into_iter().cloned().collect();
                subargs.into_iter().rev().for_each(|a| unprocessed_args.push_front(a));
            }
            Expr::Const(Value::Null) => have_null = true,
            Expr::Const(Value::Bool(false)) => return (Vec::new(), have_null, true),
            Expr::Const(Value::Bool(true)) => {}
            arg => newargs.push(arg),
        }
    }
    (newargs, have_null, false)
}

/// simplify_boolean_equality returns a comparison of a boolean with a constant boolean as the boolean itself or its
/// negation, as Postgres' function of the same name does, or the comparison when neither side is such a constant.
fn simplify_boolean_equality(op: crate::expr::CmpOp, left: Expr, right: Expr) -> Expr {
    use crate::types::Value;
    let equal = op == crate::expr::CmpOp::Eq;
    match (left, right) {
        (Expr::Const(Value::Bool(b)), other) | (other, Expr::Const(Value::Bool(b))) => match b == equal {
            true => other,
            false => super::prepqual::negate_clause(other),
        },
        (left, right) => Expr::Compare(op, Box::new(left), Box::new(right)),
    }
}

/// cast_is_immutable reports whether a cast of a constant to a type gives the same value in every session, which it
/// does unless it reads the time zone, a cast to or from a type with one.
fn cast_is_immutable(inner: &Expr, target: u32) -> bool {
    use crate::types::Value;
    let zoned = |oid: u32| matches!(oid, crate::oid::TIMESTAMPTZ | crate::oid::TIMETZ);
    !zoned(target) && !matches!(inner, Expr::Const(Value::TimestampTz(_) | Value::TimeTz(..)))
}

/// const_except_element reports whether an argument of a comparison with each element of an array is a constant, or
/// is built of constants and the array element that the comparison reads.
fn const_except_element(e: &Expr) -> bool {
    match e {
        Expr::Const(_) | Expr::SubqueryValue => true,
        Expr::Column(_) | Expr::Param(_) | Expr::Outer(..) | Expr::InputColumn(_) => false,
        other => {
            let mut all = true;
            other.visit_children(&mut |c| all &= const_except_element(c));
            all
        }
    }
}

/// contain_volatile_node reports whether an expression calls a volatile function, outside the expressions of
/// PlaceHolderVars.
fn contain_volatile_node(e: &Expr) -> bool {
    let mut found = false;
    e.visit(&mut |x| found |= is_volatile_node(x));
    found
}
