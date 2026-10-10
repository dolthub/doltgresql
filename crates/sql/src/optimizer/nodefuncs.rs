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

//! The parts of Postgres' nodes/nodeFuncs.c and utils/cache/lsyscache.c that the planner calls: the type of an
//! expression, and the btree operator families of comparisons. Doltgres has no pg_opfamily, so a family is the
//! btree family of Postgres' default operator class for the compared type.

use super::PlannerInfo;
use super::nodes::{Query, VarNode};
use crate::expr::Expr;
use crate::types::Value;

/// BOOLOID is the OID of the boolean type.
pub const BOOLOID: u32 = 16;

/// expr_type returns the OID of the type of an expression, or None where the planner does not know it, as Postgres'
/// exprType does.
pub fn expr_type(root: &PlannerInfo<'_, '_>, e: &Expr) -> Option<u32> {
    query_expr_type(root.glob, &root.parse, e)
}

/// query_expr_type is expr_type for an expression of a query.
pub fn query_expr_type(glob: &super::nodes::PlannerGlobal, parse: &Query, e: &Expr) -> Option<u32> {
    match e {
        Expr::Const(value) => value_type(value),
        Expr::Column(id) => match glob.node(*id) {
            VarNode::Var(var) if var.varno == 0 => {
                parse.set_operations.as_ref().and_then(|op| op.col_types.get(var.varattno).copied().flatten())
            }
            VarNode::Var(var) => parse.rte(var.varno).coltypes.get(var.varattno).copied().flatten(),
            VarNode::PlaceHolderVar(phv) => query_expr_type(glob, parse, &glob.placeholder(phv.phid).phexpr),
        },
        Expr::Cast(_, ty, _) | Expr::Arith(.., ty) | Expr::Neg(_, ty) => Some(ty.oid),
        Expr::Compare(..)
        | Expr::RowCompare(..)
        | Expr::And(..)
        | Expr::Or(..)
        | Expr::Not(_)
        | Expr::IsNull(..)
        | Expr::BoolTest(..)
        | Expr::DistinctFrom(..)
        | Expr::Exists(_)
        | Expr::AnySubquery(..)
        | Expr::AnyArray(..) => Some(BOOLOID),
        Expr::SubPlan(subplan) => query_expr_type(glob, parse, &subplan.link),
        Expr::AlternativeSubPlan(subplans) => query_expr_type(glob, parse, &subplans[0].link),
        Expr::Func(f, _) => known_type(crate::functions::function(*f).ret),
        Expr::AggRef(k) => parse.aggregates.get(*k).map(|call| call.ret),
        Expr::WindowRef(k) => parse.window_funcs.get(*k).map(|call| call.ret.oid),
        Expr::SetRef(k) => parse.target_srfs.get(*k).and_then(|call| query_expr_type(glob, parse, call)),
        Expr::DateTime(op, l, r) => {
            use crate::expr::DateOp as D;
            Some(match op {
                D::DatePlusDays | D::DateMinusDays => 1082,
                D::DateMinusDate => 23,
                D::TimestampPlusInterval(true) | D::TimestampMinusInterval(true) | D::DatePlusTimeTz => 1184,
                D::TimestampPlusInterval(false) | D::TimestampMinusInterval(false) | D::DatePlusTime => 1114,
                D::TimePlusInterval | D::TimeMinusInterval => {
                    let lt = query_expr_type(glob, parse, l)?;
                    if lt == 1186 { query_expr_type(glob, parse, r)? } else { lt }
                }
                D::TimestampMinusTimestamp
                | D::TimeMinusTime
                | D::IntervalPlusInterval
                | D::IntervalMinusInterval
                | D::IntervalTimesFloat
                | D::IntervalDivFloat => 1186,
            })
        }
        Expr::Grouping(..) => Some(23),
        Expr::Routine(routine, _) | Expr::Operator(_, routine, ..) => known_type(routine.ret.oid),
        Expr::Coalesce(args) | Expr::MinMax(_, args) => query_expr_type(glob, parse, args.first()?),
        Expr::NullIf(value, _) => query_expr_type(glob, parse, value),
        Expr::Case(whens, otherwise) => match whens.first() {
            Some((_, result)) => query_expr_type(glob, parse, result),
            None => query_expr_type(glob, parse, otherwise),
        },
        _ => None,
    }
}

/// known_type returns a type OID unless it is a pseudo-type, whose actual type depends on the arguments.
fn known_type(oid: u32) -> Option<u32> {
    let pseudo = matches!(oid, 2276..=2283 | 2776 | 3500 | 3831 | 4537 | 4538 | 5077..=5080);
    (!pseudo && oid != 0).then_some(oid)
}

/// value_type returns the OID of the type of a constant.
pub fn value_type(value: &Value) -> Option<u32> {
    Some(match value {
        Value::Bool(_) => BOOLOID,
        Value::Int2(_) => 21,
        Value::Int4(_) => 23,
        Value::Int8(_) => 20,
        Value::Float4(_) => 700,
        Value::Float8(_) => 701,
        Value::Numeric(_) => 1700,
        Value::Date(_) => 1082,
        Value::Time(_) => 1083,
        Value::TimeTz(..) => 1266,
        Value::Timestamp(_) => 1114,
        Value::TimestampTz(_) => 1184,
        Value::Interval(_) => 1186,
        Value::Text(_) => 25,
        Value::Bytea(_) => 17,
        Value::Uuid(_) => 2950,
        Value::Oid(_) => 26,
        Value::Array(array) => array.array_type(),
        _ => return None,
    })
}

/// btree_opfamily returns the btree operator family of the default operator class of a type, as Postgres' catalog
/// assigns them, or None for a type without one.
pub fn btree_opfamily(oid: u32) -> Option<u32> {
    Some(match oid {
        21 | 23 | 20 => 1976,
        700 | 701 => 1970,
        1700 => 1988,
        25 | 1043 => 1994,
        19 => 1986,
        1042 => 426,
        16 => 424,
        1082 | 1114 | 1184 => 434,
        1083 => 1996,
        1266 => 2000,
        1186 => 1982,
        17 => 428,
        18 => 429,
        26 => 1989,
        2950 => 2968,
        3802 => 4033,
        869 | 650 => 1974,
        790 => 2099,
        1560 => 423,
        1562 => 2002,
        3220 => 3260,
        3614 => 3626,
        3615 => 3683,
        2249 => 2994,
        _ => match crate::catalog::builtin_type(oid) {
            Some(t) if t.elem != 0 => btree_opfamily(t.elem).map(|_| 397)?,
            Some(_) => return None,
            None => match &crate::usertypes::get(oid)?.kind {
                crate::usertypes::Kind::Enum(_) => 3522,
                crate::usertypes::Kind::Composite(_) => 2994,
                crate::usertypes::Kind::Array(elem) => btree_opfamily(*elem).map(|_| 397)?,
                crate::usertypes::Kind::Domain(domain) => btree_opfamily(domain.base.oid)?,
                _ => return None,
            },
        },
    })
}

/// get_mergejoin_opfamilies returns the btree operator families of an equality of two expressions of known types,
/// when a merge join can use it, as Postgres' op_mergejoinable and get_mergejoin_opfamilies decide.
pub fn get_mergejoin_opfamilies(left: Option<u32>, right: Option<u32>) -> Vec<u32> {
    match (left.and_then(btree_opfamily), right.and_then(btree_opfamily)) {
        (Some(l), Some(r)) if l == r => vec![l],
        _ => Vec::new(),
    }
}
