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

//! The parts of Postgres' optimizer/util/clauses.c and var.c that the planner calls: the relations and Vars that an
//! expression reads, and which of them it can only be true with when they are not NULL.

use super::nodes::{Relids, singleton, var_parts};
use crate::expr::Expr;

/// pull_varnos returns the range table indexes that an expression's Vars read, as Postgres' pull_varnos does for
/// the current query level.
pub fn pull_varnos(e: &Expr) -> Relids {
    let mut relids = 0;
    e.visit(&mut |x| {
        if let Expr::Column(c) = x {
            relids |= singleton(var_parts(*c).0);
        }
    });
    relids
}

/// contain_volatile_functions reports whether an expression calls a volatile function, as Postgres' function of the
/// same name does, counting user-defined routines and operators as volatile, which is their default.
pub fn contain_volatile_functions(e: &Expr) -> bool {
    let mut volatile = false;
    e.visit(&mut |x| match x {
        Expr::Func(f, _) => volatile |= crate::pgcatalog::is_volatile(crate::functions::function(*f).name),
        Expr::Routine(..) | Expr::Operator(..) => volatile = true,
        _ => {}
    });
    volatile
}

/// contain_vars reports whether an expression reads a Var of the query, as Postgres' contain_vars_of_level does for
/// the current level.
pub fn contain_vars(e: &Expr) -> bool {
    pull_varnos(e) != 0
}

/// contain_nonstrict_functions reports whether an expression can be non-NULL when an input is NULL, because one of
/// its parts is not a value, a Var, or a strict operator, cast, or function, as Postgres' function of the same name
/// does.
pub fn contain_nonstrict_functions(e: &Expr) -> bool {
    let mut nonstrict = false;
    e.visit(&mut |x| {
        nonstrict |= match x {
            Expr::Column(_)
            | Expr::Const(_)
            | Expr::Param(_)
            | Expr::Outer(..)
            | Expr::Cast(..)
            | Expr::Arith(..)
            | Expr::Neg(..)
            | Expr::Compare(..)
            | Expr::DateTime(..) => false,
            Expr::Func(f, _) => !crate::functions::function(*f).strict,
            _ => true,
        }
    });
    nonstrict
}

/// strict_args returns the arguments of a strict operator or function, which is NULL when any of them is, or None
/// for any other expression.
fn strict_args(e: &Expr) -> Option<Vec<&Expr>> {
    match e {
        Expr::Cast(inner, ..) | Expr::Neg(inner, _) => Some(vec![inner]),
        Expr::Arith(_, l, r, _) | Expr::Compare(_, l, r) | Expr::DateTime(_, l, r) => Some(vec![l, r]),
        Expr::Func(f, args) if crate::functions::function(*f).strict => Some(args.iter().collect()),
        _ => None,
    }
}

/// find_nonnullable_rels returns the relations whose Vars a clause can only be true with when they are not NULL, as
/// Postgres' find_nonnullable_rels finds them.
pub fn find_nonnullable_rels(e: &Expr) -> Relids {
    nonnullable_rels(e, true)
}

/// nonnullable_rels is find_nonnullable_rels_walker, given whether the expression is at the clause's top level,
/// where an AND is true only when each argument is and a NULL test or boolean test is true or false.
fn nonnullable_rels(e: &Expr, top_level: bool) -> Relids {
    if let Some(args) = strict_args(e) {
        return args.into_iter().fold(0, |relids, a| relids | nonnullable_rels(a, false));
    }
    match e {
        Expr::Column(c) => singleton(var_parts(*c).0),
        Expr::And(a, b) if top_level => nonnullable_rels(a, true) | nonnullable_rels(b, true),
        Expr::And(a, b) | Expr::Or(a, b) => nonnullable_rels(a, top_level) & nonnullable_rels(b, top_level),
        Expr::Not(inner) => nonnullable_rels(inner, false),
        Expr::IsNull(inner, true) | Expr::BoolTest(inner, Some(_), false) | Expr::BoolTest(inner, None, true)
            if top_level =>
        {
            nonnullable_rels(inner, false)
        }
        _ => 0,
    }
}

/// find_nonnullable_vars returns the Vars that a clause can only be true with when they are not NULL, as Postgres'
/// find_nonnullable_vars finds them.
pub fn find_nonnullable_vars(e: &Expr) -> Vec<usize> {
    nonnullable_vars(e, true)
}

/// nonnullable_vars is find_nonnullable_vars_walker, given whether the expression is at the clause's top level.
fn nonnullable_vars(e: &Expr, top_level: bool) -> Vec<usize> {
    if let Some(args) = strict_args(e) {
        return args.into_iter().flat_map(|a| nonnullable_vars(a, false)).collect();
    }
    match e {
        Expr::Column(c) => vec![*c],
        Expr::And(a, b) if top_level => [nonnullable_vars(a, true), nonnullable_vars(b, true)].concat(),
        Expr::And(a, b) | Expr::Or(a, b) => {
            let right = nonnullable_vars(b, top_level);
            nonnullable_vars(a, top_level).into_iter().filter(|v| right.contains(v)).collect()
        }
        Expr::Not(inner) => nonnullable_vars(inner, false),
        Expr::IsNull(inner, true) | Expr::BoolTest(inner, Some(_), false) | Expr::BoolTest(inner, None, true)
            if top_level =>
        {
            nonnullable_vars(inner, false)
        }
        _ => Vec::new(),
    }
}

/// find_forced_null_vars returns the Vars that a clause can only be true with when they are NULL, from its top-level
/// IS NULL and IS UNKNOWN tests of Vars, as Postgres' find_forced_null_vars finds them.
pub fn find_forced_null_vars(e: &Expr) -> Vec<usize> {
    match e {
        Expr::And(a, b) => [find_forced_null_vars(a), find_forced_null_vars(b)].concat(),
        Expr::IsNull(inner, false) | Expr::BoolTest(inner, None, false) => match **inner {
            Expr::Column(c) => vec![c],
            _ => Vec::new(),
        },
        _ => Vec::new(),
    }
}
