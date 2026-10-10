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

//! Postgres' optimizer/prep/prepqual.c: negating a clause by pushing the NOT down, and taking the conditions that
//! every argument of an OR shares out of it. Doltgres' ANDs and ORs are binary, so a nest of them stands for one
//! Postgres BoolExpr with all of their arguments.

use super::restrictinfo::{and_args, or_args};
use crate::expr::{CmpOp, Expr};
use crate::types::Value;

/// negate_clause returns the negation of a clause, pushing the NOT into it where that keeps its meaning, as Postgres'
/// function of the same name does: a constant flips, a comparison takes its negator, an AND becomes an OR of the
/// negated arguments and an OR an AND, a NOT goes away, and an IS NULL, IS DISTINCT FROM, or boolean test flips.
pub fn negate_clause(node: Expr) -> Expr {
    match node {
        Expr::Const(Value::Null) => Expr::Const(Value::Null),
        Expr::Const(Value::Bool(b)) => Expr::Const(Value::Bool(!b)),
        Expr::Compare(op, l, r) => Expr::Compare(negator(op), l, r),
        Expr::AnyArray(comparison, array, all) => match *comparison {
            Expr::Compare(op, l, r) => Expr::AnyArray(Box::new(Expr::Compare(negator(op), l, r)), array, !all),
            comparison => Expr::Not(Box::new(Expr::AnyArray(Box::new(comparison), array, all))),
        },
        Expr::And(a, b) => make_orclause(vec![negate_clause(*a), negate_clause(*b)]),
        Expr::Or(a, b) => make_andclause(vec![negate_clause(*a), negate_clause(*b)]),
        Expr::Not(inner) => *inner,
        Expr::IsNull(arg, negated) if !matches!(*arg, Expr::Row(..)) => Expr::IsNull(arg, !negated),
        Expr::DistinctFrom(l, r, negated) => Expr::DistinctFrom(l, r, !negated),
        Expr::BoolTest(arg, value, negated) => Expr::BoolTest(arg, value, !negated),
        other => Expr::Not(Box::new(other)),
    }
}

/// negator returns the comparison that holds exactly when another one is false for values that are not NULL, as
/// Postgres' get_negator finds an operator's negator.
fn negator(op: CmpOp) -> CmpOp {
    match op {
        CmpOp::Eq => CmpOp::Ne,
        CmpOp::Ne => CmpOp::Eq,
        CmpOp::Lt => CmpOp::Ge,
        CmpOp::Le => CmpOp::Gt,
        CmpOp::Gt => CmpOp::Le,
        CmpOp::Ge => CmpOp::Lt,
    }
}

/// make_andclause returns the AND of a list of clauses, as Postgres' function of the same name does.
pub fn make_andclause(args: Vec<Expr>) -> Expr {
    args.into_iter().reduce(|a, b| Expr::And(Box::new(a), Box::new(b))).expect("an AND has arguments")
}

/// make_orclause returns the OR of a list of clauses, as Postgres' function of the same name does.
pub fn make_orclause(args: Vec<Expr>) -> Expr {
    args.into_iter().reduce(|a, b| Expr::Or(Box::new(a), Box::new(b))).expect("an OR has arguments")
}

/// canonicalize_qual simplifies a WHERE or JOIN condition by taking out of each OR the conditions that all of its
/// arguments share, as Postgres' function of the same name does for a qual that is not a CHECK constraint.
pub fn canonicalize_qual(qual: Expr) -> Expr {
    find_duplicate_ors(qual)
}

/// find_duplicate_ors simplifies the ORs of a condition with process_duplicate_ors, dropping the constant arguments of
/// ANDs and ORs that do not decide them, as Postgres' function of the same name does.
fn find_duplicate_ors(qual: Expr) -> Expr {
    match qual {
        Expr::Or(..) => {
            let mut orlist = Vec::new();
            for arg in or_args(&qual) {
                let arg = find_duplicate_ors(arg.clone());
                match arg {
                    Expr::Const(Value::Null) | Expr::Const(Value::Bool(false)) => continue,
                    Expr::Const(_) => return arg,
                    _ => orlist.push(arg),
                }
            }
            process_duplicate_ors(pull_ors(orlist))
        }
        Expr::And(..) => {
            let mut andlist = Vec::new();
            for arg in and_args(&qual) {
                let arg = find_duplicate_ors(arg.clone());
                match arg {
                    Expr::Const(Value::Bool(true)) => continue,
                    Expr::Const(_) => return Expr::Const(Value::Bool(false)),
                    _ => andlist.push(arg),
                }
            }
            match pull_ands(andlist) {
                andlist if andlist.is_empty() => Expr::Const(Value::Bool(true)),
                andlist => make_andclause(andlist),
            }
        }
        other => other,
    }
}

/// pull_ands flattens the ANDs in a list of AND arguments, as Postgres' function of the same name does.
fn pull_ands(andlist: Vec<Expr>) -> Vec<Expr> {
    andlist.iter().flat_map(and_args).cloned().collect()
}

/// pull_ors flattens the ORs in a list of OR arguments, as Postgres' function of the same name does.
fn pull_ors(orlist: Vec<Expr>) -> Vec<Expr> {
    orlist.iter().flat_map(or_args).cloned().collect()
}

/// process_duplicate_ors returns an OR of a list of arguments with the conditions that each of them has, as itself or
/// as an argument of its AND, taken out of the OR into an AND around it, as Postgres' function of the same name does:
/// `(A AND B) OR (A AND C)` becomes `A AND (B OR C)`, and an OR left with an argument that had nothing else goes away.
fn process_duplicate_ors(orlist: Vec<Expr>) -> Expr {
    match orlist.len() {
        0 => return Expr::Const(Value::Bool(false)),
        1 => return orlist.into_iter().next().expect("one argument"),
        _ => {}
    }
    let mut reference: Option<Vec<&Expr>> = None;
    for clause in &orlist {
        match clause {
            Expr::And(..) => {
                let subclauses = and_args(clause);
                if reference.as_ref().is_none_or(|r| subclauses.len() < r.len()) {
                    reference = Some(subclauses);
                }
            }
            other => {
                reference = Some(vec![other]);
                break;
            }
        }
    }
    let mut unique_reference: Vec<&Expr> = Vec::new();
    for refclause in reference.unwrap_or_default() {
        if !unique_reference.contains(&refclause) {
            unique_reference.push(refclause);
        }
    }
    let winners: Vec<Expr> = unique_reference
        .into_iter()
        .filter(|refclause| {
            orlist.iter().all(|clause| match clause {
                Expr::And(..) => and_args(clause).contains(refclause),
                other => other == *refclause,
            })
        })
        .cloned()
        .collect();
    if winners.is_empty() {
        return make_orclause(orlist);
    }
    let mut neworlist = Vec::new();
    for clause in &orlist {
        match clause {
            Expr::And(..) => {
                let subclauses: Vec<Expr> =
                    and_args(clause).into_iter().filter(|c| !winners.contains(c)).cloned().collect();
                if subclauses.is_empty() {
                    neworlist.clear();
                    break;
                }
                neworlist.push(make_andclause(subclauses));
            }
            other => {
                if winners.contains(other) {
                    neworlist.clear();
                    break;
                }
                neworlist.push(other.clone());
            }
        }
    }
    let mut winners = winners;
    match neworlist.len() {
        0 => {}
        1 => winners.extend(neworlist),
        _ => winners.push(make_orclause(pull_ors(neworlist))),
    }
    match winners.len() {
        1 => winners.into_iter().next().expect("one winner"),
        _ => make_andclause(pull_ands(winners)),
    }
}
