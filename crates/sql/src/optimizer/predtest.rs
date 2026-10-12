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

//! Postgres' optimizer/util/predtest.c: proving that clauses imply or refute a predicate, as partial indexes and
//! constraint exclusion need. A comparison's btree operator family is the one of its operands' type, and two
//! comparisons are the same operator when they compare by the same operation in the same family.

use super::PlannerInfo;
use crate::expr::{CmpOp, Expr};
use crate::types::Value;

/// MAX_SAOP_ARRAY_SIZE is the most elements of an IN list or ANY array that a proof treats as an AND or OR of its
/// comparisons, as Postgres' constant of the same name is.
const MAX_SAOP_ARRAY_SIZE: usize = 100;

/// PredClass is how a proof takes a clause apart, as Postgres' PredClass is: an AND or OR of its items, or an atom.
enum PredClass {
    Atom,
    And(Vec<Expr>),
    Or(Vec<Expr>),
}

/// predicate_implied_by reports whether clauses, taken as an AND, imply a predicate, an AND of expressions, as
/// Postgres' function of the same name does: always true, or with `weak`, never false.
pub fn predicate_implied_by(root: &PlannerInfo<'_, '_>, predicate: &[Expr], clauses: &[Expr], weak: bool) -> bool {
    if predicate.is_empty() {
        return true;
    }
    if clauses.is_empty() {
        return false;
    }
    predicate_implied_by_recurse(root, &as_node(clauses), &as_node(predicate), weak)
}

/// predicate_refuted_by reports whether clauses, taken as an AND, refute a predicate, an AND of expressions, as
/// Postgres' function of the same name does: make it false, or with `weak`, never true.
pub fn predicate_refuted_by(root: &PlannerInfo<'_, '_>, predicate: &[Expr], clauses: &[Expr], weak: bool) -> bool {
    if predicate.is_empty() || clauses.is_empty() {
        return false;
    }
    predicate_refuted_by_recurse(root, &as_node(clauses), &as_node(predicate), weak)
}

/// as_node returns a list of expressions as one expression, an AND of them when there are several.
fn as_node(list: &[Expr]) -> Expr {
    list.iter().cloned().reduce(|a, b| Expr::And(Box::new(a), Box::new(b))).expect("a list of expressions")
}

/// predicate_implied_by_recurse reports whether a clause implies a predicate, taking apart ANDs, ORs, and arrays of
/// either, as Postgres' function of the same name does.
fn predicate_implied_by_recurse(root: &PlannerInfo<'_, '_>, clause: &Expr, predicate: &Expr, weak: bool) -> bool {
    let pclass = predicate_classify(predicate);
    match (predicate_classify(clause), pclass) {
        (PredClass::And(_), PredClass::And(pitems)) => {
            pitems.iter().all(|pitem| predicate_implied_by_recurse(root, clause, pitem, weak))
        }
        (PredClass::And(citems), PredClass::Or(pitems)) => {
            pitems.iter().any(|pitem| predicate_implied_by_recurse(root, clause, pitem, weak))
                || citems.iter().any(|citem| predicate_implied_by_recurse(root, citem, predicate, weak))
        }
        (PredClass::And(citems), PredClass::Atom) => {
            citems.iter().any(|citem| predicate_implied_by_recurse(root, citem, predicate, weak))
        }
        (PredClass::Or(citems), PredClass::Or(pitems)) => {
            citems.iter().all(|citem| pitems.iter().any(|pitem| predicate_implied_by_recurse(root, citem, pitem, weak)))
        }
        (PredClass::Or(citems), PredClass::And(_) | PredClass::Atom) => {
            citems.iter().all(|citem| predicate_implied_by_recurse(root, citem, predicate, weak))
        }
        (PredClass::Atom, PredClass::And(pitems)) => {
            pitems.iter().all(|pitem| predicate_implied_by_recurse(root, clause, pitem, weak))
        }
        (PredClass::Atom, PredClass::Or(pitems)) => {
            pitems.iter().any(|pitem| predicate_implied_by_recurse(root, clause, pitem, weak))
        }
        (PredClass::Atom, PredClass::Atom) => predicate_implied_by_simple_clause(root, predicate, clause, weak),
    }
}

/// predicate_refuted_by_recurse reports whether a clause refutes a predicate, taking apart ANDs, ORs, NOTs, and
/// arrays, as Postgres' function of the same name does.
fn predicate_refuted_by_recurse(root: &PlannerInfo<'_, '_>, clause: &Expr, predicate: &Expr, weak: bool) -> bool {
    let pclass = predicate_classify(predicate);
    let implied_by_not = |root: &PlannerInfo<'_, '_>| {
        extract_not_arg(predicate).is_some_and(|not_arg| predicate_implied_by_recurse(root, clause, not_arg, false))
    };
    match predicate_classify(clause) {
        PredClass::And(citems) => match pclass {
            PredClass::And(pitems) => {
                pitems.iter().any(|pitem| predicate_refuted_by_recurse(root, clause, pitem, weak))
                    || citems.iter().any(|citem| predicate_refuted_by_recurse(root, citem, predicate, weak))
            }
            PredClass::Or(pitems) => pitems.iter().all(|pitem| predicate_refuted_by_recurse(root, clause, pitem, weak)),
            PredClass::Atom => {
                implied_by_not(root)
                    || citems.iter().any(|citem| predicate_refuted_by_recurse(root, citem, predicate, weak))
            }
        },
        PredClass::Or(citems) => match pclass {
            PredClass::Or(pitems) => pitems.iter().all(|pitem| predicate_refuted_by_recurse(root, clause, pitem, weak)),
            PredClass::And(pitems) => citems
                .iter()
                .all(|citem| pitems.iter().any(|pitem| predicate_refuted_by_recurse(root, citem, pitem, weak))),
            PredClass::Atom => {
                implied_by_not(root)
                    || citems.iter().all(|citem| predicate_refuted_by_recurse(root, citem, predicate, weak))
            }
        },
        PredClass::Atom => {
            if extract_strong_not_arg(clause)
                .is_some_and(|not_arg| predicate_implied_by_recurse(root, predicate, not_arg, !weak))
            {
                return true;
            }
            match pclass {
                PredClass::And(pitems) => {
                    pitems.iter().any(|pitem| predicate_refuted_by_recurse(root, clause, pitem, weak))
                }
                PredClass::Or(pitems) => {
                    pitems.iter().all(|pitem| predicate_refuted_by_recurse(root, clause, pitem, weak))
                }
                PredClass::Atom => {
                    implied_by_not(root) || predicate_refuted_by_simple_clause(root, predicate, clause, weak)
                }
            }
        }
    }
}

/// predicate_classify returns how a proof takes an expression apart: an AND or OR into its arguments, an IN list or
/// ANY array of at most MAX_SAOP_ARRAY_SIZE elements into the comparison with each element, joined by OR, or by AND
/// for ALL, and anything else as an atom, as Postgres' function of the same name does.
fn predicate_classify(clause: &Expr) -> PredClass {
    use super::restrictinfo::{and_args, or_args};
    match clause {
        Expr::And(..) => PredClass::And(and_args(clause).into_iter().cloned().collect()),
        Expr::Or(..) => PredClass::Or(or_args(clause).into_iter().cloned().collect()),
        Expr::AnyArray(comparison, array, all) => {
            let Expr::Compare(op, left, value) = &**comparison else { return PredClass::Atom };
            if !matches!(**value, Expr::SubqueryValue) {
                return PredClass::Atom;
            }
            let elements: Vec<Expr> = match &**array {
                Expr::Const(Value::Array(a)) if a.values.len() <= MAX_SAOP_ARRAY_SIZE => {
                    a.values.iter().map(|v| Expr::Const(v.clone())).collect()
                }
                Expr::Array(_, items, false) if items.len() <= MAX_SAOP_ARRAY_SIZE => items.clone(),
                _ => return PredClass::Atom,
            };
            let items = elements.into_iter().map(|e| Expr::Compare(*op, left.clone(), Box::new(e))).collect();
            match all {
                true => PredClass::And(items),
                false => PredClass::Or(items),
            }
        }
        _ => PredClass::Atom,
    }
}

/// predicate_implied_by_simple_clause reports whether an atom clause implies an atom predicate, as Postgres'
/// function of the same name does: an equal expression, a boolean column's equality with true or false, a strict
/// clause for an IS NOT NULL test, or a comparison of the same expression with constants.
fn predicate_implied_by_simple_clause(root: &PlannerInfo<'_, '_>, predicate: &Expr, clause: &Expr, weak: bool) -> bool {
    if predicate == clause {
        return true;
    }
    if let Expr::Compare(CmpOp::Eq, leftop, rightop) = clause
        && let Expr::Const(Value::Bool(b)) = &**rightop
    {
        match b {
            true if predicate == &**leftop => return true,
            false if matches!(predicate, Expr::Not(arg) if **arg == **leftop) => return true,
            _ => {}
        }
    }
    if let Some(arg) = is_not_null_test(predicate)
        && !weak
        && !matches!(arg, Expr::Row(..))
        && clause_is_strict_for(clause, arg, true)
    {
        return true;
    }
    operator_predicate_proof(root, predicate, clause, false, weak)
}

/// predicate_refuted_by_simple_clause reports whether an atom clause refutes an atom predicate, as Postgres'
/// function of the same name does: opposite NULL tests of one expression, a strict expression of what a clause tests
/// IS NULL, or comparisons of the same expression with constants.
fn predicate_refuted_by_simple_clause(root: &PlannerInfo<'_, '_>, predicate: &Expr, clause: &Expr, weak: bool) -> bool {
    if let Expr::IsNull(clause_arg, false) = clause {
        if matches!(**clause_arg, Expr::Row(..)) {
            return false;
        }
        if let Some(pred_arg) = is_not_null_test(predicate) {
            if matches!(pred_arg, Expr::Row(..)) {
                return false;
            }
            if pred_arg == &**clause_arg {
                return true;
            }
        }
        return weak && clause_is_strict_for(predicate, clause_arg, true);
    }
    if let Expr::IsNull(pred_arg, false) = predicate {
        if matches!(**pred_arg, Expr::Row(..)) {
            return false;
        }
        if let Some(clause_arg) = is_not_null_test(clause) {
            if matches!(clause_arg, Expr::Row(..)) {
                return false;
            }
            if clause_arg == &**pred_arg {
                return true;
            }
        }
        return clause_is_strict_for(clause, pred_arg, true);
    }
    if is_not_null_test(predicate).is_some() {
        return false;
    }
    operator_predicate_proof(root, predicate, clause, true, weak)
}

/// is_not_null_test returns the argument of an IS NOT NULL test.
fn is_not_null_test(e: &Expr) -> Option<&Expr> {
    match e {
        Expr::IsNull(arg, true) => Some(arg),
        Expr::Not(inner) => match &**inner {
            Expr::IsNull(arg, false) => Some(arg),
            _ => None,
        },
        _ => None,
    }
}

/// extract_not_arg returns what a NOT, IS NOT TRUE, IS FALSE, or IS UNKNOWN test negates, as Postgres' function of
/// the same name does.
fn extract_not_arg(clause: &Expr) -> Option<&Expr> {
    match clause {
        Expr::Not(arg) => Some(arg),
        Expr::BoolTest(arg, Some(true), true)
        | Expr::BoolTest(arg, Some(false), false)
        | Expr::BoolTest(arg, None, false) => Some(arg),
        _ => None,
    }
}

/// extract_strong_not_arg returns what a NOT or IS FALSE test negates, as Postgres' function of the same name does.
fn extract_strong_not_arg(clause: &Expr) -> Option<&Expr> {
    match clause {
        Expr::Not(arg) => Some(arg),
        Expr::BoolTest(arg, Some(false), false) => Some(arg),
        _ => None,
    }
}

/// clause_is_strict_for reports whether a clause is NULL, or with `allow_false` NULL or false, whenever an expression
/// is NULL, as Postgres' function of the same name does.
fn clause_is_strict_for(clause: &Expr, subexpr: &Expr, allow_false: bool) -> bool {
    if clause == subexpr {
        return true;
    }
    match clause {
        Expr::Compare(_, l, r) | Expr::Arith(_, l, r, _) | Expr::Concat(l, r) | Expr::DateTime(_, l, r) => {
            clause_is_strict_for(l, subexpr, false) || clause_is_strict_for(r, subexpr, false)
        }
        Expr::RowCompare(..) => false,
        Expr::Operator(_, routine, l, r) if routine.strict => {
            clause_is_strict_for(l, subexpr, false) || clause_is_strict_for(r, subexpr, false)
        }
        Expr::Func(f, args) if crate::functions::function(*f).strict => {
            args.iter().any(|a| clause_is_strict_for(a, subexpr, false))
        }
        Expr::Routine(routine, args) if routine.strict => args.iter().any(|a| clause_is_strict_for(a, subexpr, false)),
        Expr::Cast(arg, ..) | Expr::Neg(arg, _) => clause_is_strict_for(arg, subexpr, false),
        Expr::AnyArray(comparison, array, all) => {
            if let Expr::Compare(_, scalar, value) = &**comparison
                && matches!(**value, Expr::SubqueryValue)
                && clause_is_strict_for(scalar, subexpr, false)
            {
                if allow_false && !all {
                    return true;
                }
                let nelems = match &**array {
                    Expr::Const(Value::Null) => return true,
                    Expr::Const(Value::Array(a)) => a.values.len(),
                    Expr::Array(_, items, false) => items.len(),
                    _ => 0,
                };
                if nelems > 0 {
                    return true;
                }
            }
            clause_is_strict_for(array, subexpr, false)
        }
        Expr::Const(value) => value.is_null(),
        _ => false,
    }
}

/// CompareType is a btree strategy, as Postgres' CompareType numbers them from 1.
#[derive(Clone, Copy, PartialEq, Eq)]
enum CompareType {
    Lt = 1,
    Le,
    Eq,
    Ge,
    Gt,
    Ne,
}

/// compare_type returns the btree strategy of a comparison, where an inequality's is that of the negator of its
/// family's equality.
fn compare_type(op: CmpOp) -> CompareType {
    match op {
        CmpOp::Lt => CompareType::Lt,
        CmpOp::Le => CompareType::Le,
        CmpOp::Eq => CompareType::Eq,
        CmpOp::Ge => CompareType::Ge,
        CmpOp::Gt => CompareType::Gt,
        CmpOp::Ne => CompareType::Ne,
    }
}

/// cmp_op returns the comparison of a btree strategy.
fn cmp_op(cmptype: CompareType) -> CmpOp {
    match cmptype {
        CompareType::Lt => CmpOp::Lt,
        CompareType::Le => CmpOp::Le,
        CompareType::Eq => CmpOp::Eq,
        CompareType::Ge => CmpOp::Ge,
        CompareType::Gt => CmpOp::Gt,
        CompareType::Ne => CmpOp::Ne,
    }
}

/// RC_IMPLIES_TABLE and RC_REFUTES_TABLE say whether a clause `A op1 B` implies or refutes a predicate `A op2 B`, by
/// the clause's strategy and then the predicate's, as Postgres' tables of the same names do.
const RC_IMPLIES_TABLE: [[bool; 6]; 6] = [
    [true, true, false, false, false, true],
    [false, true, false, false, false, false],
    [false, true, true, true, false, false],
    [false, false, false, true, false, false],
    [false, false, false, true, true, true],
    [false, false, false, false, false, true],
];
const RC_REFUTES_TABLE: [[bool; 6]; 6] = [
    [false, false, true, true, true, false],
    [false, false, false, false, true, false],
    [true, false, false, false, true, true],
    [true, false, false, false, false, false],
    [true, true, true, false, false, false],
    [false, false, true, false, false, false],
];

/// RC_IMPLIC_TABLE and RC_REFUTE_TABLE give the test between the constants of a clause `A op1 c1` and a predicate
/// `A op2 c2` that proves the implication or refutation, by the clause's strategy and then the predicate's, as
/// Postgres' tables of the same names do, where None is no test.
const RC_IMPLIC_TABLE: [[Option<CompareType>; 6]; 6] = {
    use CompareType::*;
    [
        [Some(Ge), Some(Ge), None, None, None, Some(Ge)],
        [Some(Gt), Some(Ge), None, None, None, Some(Gt)],
        [Some(Gt), Some(Ge), Some(Eq), Some(Le), Some(Lt), Some(Ne)],
        [None, None, None, Some(Le), Some(Lt), Some(Lt)],
        [None, None, None, Some(Le), Some(Le), Some(Le)],
        [None, None, None, None, None, Some(Eq)],
    ]
};
const RC_REFUTE_TABLE: [[Option<CompareType>; 6]; 6] = {
    use CompareType::*;
    [
        [None, None, Some(Ge), Some(Ge), Some(Ge), None],
        [None, None, Some(Gt), Some(Gt), Some(Ge), None],
        [Some(Le), Some(Lt), Some(Ne), Some(Gt), Some(Ge), Some(Eq)],
        [Some(Le), Some(Lt), Some(Lt), None, None, None],
        [Some(Le), Some(Le), Some(Le), None, None, None],
        [None, None, Some(Eq), None, None, None],
    ]
};

/// Operator is a comparison as a btree operator: its comparison and the operator families of its operands' types.
#[derive(Clone, Copy, PartialEq, Eq)]
struct Operator {
    op: CmpOp,
    lefttype: Option<u32>,
    righttype: Option<u32>,
}

impl Operator {
    /// commutator returns the operator with its operands swapped, as Postgres' get_commutator does.
    fn commutator(self) -> Operator {
        Operator { op: crate::indexscan::swap(self.op), lefttype: self.righttype, righttype: self.lefttype }
    }

    /// negator returns the operator that is false where this one is true, as Postgres' get_negator does.
    fn negator(self) -> Operator {
        let op = match self.op {
            CmpOp::Eq => CmpOp::Ne,
            CmpOp::Ne => CmpOp::Eq,
            CmpOp::Lt => CmpOp::Ge,
            CmpOp::Ge => CmpOp::Lt,
            CmpOp::Le => CmpOp::Gt,
            CmpOp::Gt => CmpOp::Le,
        };
        Operator { op, ..self }
    }

    /// opfamily returns the btree operator family that the operator belongs to, when both operand types share one.
    fn opfamily(self) -> Option<u32> {
        let family = |ty: Option<u32>| ty.and_then(super::nodefuncs::btree_opfamily);
        family(self.lefttype).filter(|f| family(self.righttype) == Some(*f))
    }
}

/// operator_predicate_proof reports whether a comparison clause implies, or refutes with `refute_it`, a comparison
/// predicate that compares the same expression, either with the same other expression or with constants that a
/// test of the btree family relates, as Postgres' function of the same name does.
fn operator_predicate_proof(
    root: &PlannerInfo<'_, '_>,
    predicate: &Expr,
    clause: &Expr,
    refute_it: bool,
    weak: bool,
) -> bool {
    let (Expr::Compare(pop, pred_leftop, pred_rightop), Expr::Compare(cop, clause_leftop, clause_rightop)) =
        (predicate, clause)
    else {
        return false;
    };
    let operator = |op: CmpOp, l: &Expr, r: &Expr| Operator {
        op,
        lefttype: super::nodefuncs::expr_type(root, l),
        righttype: super::nodefuncs::expr_type(root, r),
    };
    let mut pred_op = operator(*pop, pred_leftop, pred_rightop);
    let mut clause_op = operator(*cop, clause_leftop, clause_rightop);
    let constant = |e: &Expr| match e {
        Expr::Const(value) => Some(value.clone()),
        _ => None,
    };
    let (pred_const, clause_const) = if pred_leftop == clause_leftop {
        if pred_rightop == clause_rightop {
            return operator_same_subexprs_proof(pred_op, clause_op, refute_it);
        }
        let (Some(p), Some(c)) = (constant(pred_rightop), constant(clause_rightop)) else { return false };
        (p, c)
    } else if pred_rightop == clause_rightop {
        let (Some(p), Some(c)) = (constant(pred_leftop), constant(clause_leftop)) else { return false };
        pred_op = pred_op.commutator();
        clause_op = clause_op.commutator();
        (p, c)
    } else if pred_leftop == clause_rightop {
        if pred_rightop == clause_leftop {
            return operator_same_subexprs_proof(pred_op.commutator(), clause_op, refute_it);
        }
        let (Some(p), Some(c)) = (constant(pred_rightop), constant(clause_leftop)) else { return false };
        clause_op = clause_op.commutator();
        (p, c)
    } else if pred_rightop == clause_leftop {
        let (Some(p), Some(c)) = (constant(pred_leftop), constant(clause_rightop)) else { return false };
        pred_op = pred_op.commutator();
        (p, c)
    } else {
        return false;
    };
    if clause_const.is_null() {
        return !(weak && !refute_it) || pred_const.is_null();
    }
    if pred_const.is_null() {
        return weak;
    }
    let Some(test_op) = get_btree_test_op(pred_op, clause_op, refute_it) else { return false };
    test_op.test(crate::expr::compare_values(&pred_const, &clause_const))
}

/// operator_same_subexprs_proof reports whether a comparison clause implies, or refutes with `refute_it`, a
/// comparison predicate of the same two expressions, as Postgres' function of the same name does.
fn operator_same_subexprs_proof(pred_op: Operator, clause_op: Operator, refute_it: bool) -> bool {
    if refute_it && pred_op.negator() == clause_op || !refute_it && pred_op == clause_op {
        return true;
    }
    lookup_proof(pred_op, clause_op, refute_it).0
}

/// lookup_proof returns whether a clause's operator implies, or refutes with `refute_it`, a predicate's operator over
/// the same expressions, and the comparison between their constants that proves it otherwise, when both operators
/// belong to one btree family, as Postgres' lookup_proof_cache finds them.
fn lookup_proof(pred_op: Operator, clause_op: Operator, refute_it: bool) -> (bool, Option<CmpOp>) {
    if pred_op.opfamily().is_none_or(|f| clause_op.opfamily() != Some(f)) {
        return (false, None);
    }
    let (pred_cmptype, clause_cmptype) =
        (compare_type(pred_op.op) as usize - 1, compare_type(clause_op.op) as usize - 1);
    let (same_subexprs, test_cmptype) = match refute_it {
        true => (RC_REFUTES_TABLE[clause_cmptype][pred_cmptype], RC_REFUTE_TABLE[clause_cmptype][pred_cmptype]),
        false => (RC_IMPLIES_TABLE[clause_cmptype][pred_cmptype], RC_IMPLIC_TABLE[clause_cmptype][pred_cmptype]),
    };
    (same_subexprs, test_cmptype.map(cmp_op))
}

/// get_btree_test_op returns the comparison between a predicate's constant and a clause's that proves the clause
/// implies, or refutes with `refute_it`, the predicate, as Postgres' function of the same name does.
fn get_btree_test_op(pred_op: Operator, clause_op: Operator, refute_it: bool) -> Option<CmpOp> {
    lookup_proof(pred_op, clause_op, refute_it).1
}
