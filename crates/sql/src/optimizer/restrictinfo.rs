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

//! Postgres' optimizer/util/restrictinfo.c: building the RestrictInfo of a clause.

use super::clauses::{contain_volatile_functions, pull_varnos};
use super::nodes::{Relids, RestrictInfo, is_subset, overlap};
use crate::expr::{CmpOp, Expr};

/// make_restrictinfo builds a RestrictInfo of a clause, as Postgres' make_restrictinfo does, finding the relations
/// of each side of a binary operator clause, which a join can test, and whether it is an equality that a hash join
/// can use.
pub fn make_restrictinfo(
    clause: Expr,
    is_pushed_down: bool,
    pseudoconstant: bool,
    required_relids: Relids,
) -> RestrictInfo {
    let clause_relids = pull_varnos(&clause);
    let (mut left_relids, mut right_relids, mut can_join, mut hashjoinable) = (0, 0, false, false);
    if let Expr::Compare(op, l, r) = &clause {
        let (left, right) = (pull_varnos(l), pull_varnos(r));
        if left != 0 && right != 0 && !overlap(left, right) && !contain_volatile_functions(&clause) {
            (left_relids, right_relids, can_join) = (left, right, true);
            hashjoinable = *op == CmpOp::Eq && !crate::plan::has_subquery(&clause);
        }
    }
    RestrictInfo {
        clause,
        is_pushed_down,
        pseudoconstant,
        clause_relids,
        required_relids,
        left_relids,
        right_relids,
        can_join,
        hashjoinable,
    }
}

/// is_pushed_down reports whether a clause is evaluated as a filter on a join's result rather than as the join's own
/// condition, as Postgres' RINFO_IS_PUSHED_DOWN does.
pub fn is_pushed_down(rinfo: &RestrictInfo, joinrelids: Relids) -> bool {
    rinfo.is_pushed_down || !is_subset(rinfo.required_relids, joinrelids)
}
