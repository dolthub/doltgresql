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

//! Postgres' optimizer/util/inherit.c: building the members of append relations. Doltgres has no table inheritance
//! or partitioning, so its only append relations are the UNION ALL subqueries that the planner flattened.

use super::PlannerInfo;
use super::initsplan::{restriction_is_always_false, restriction_is_always_true};
use super::nodes::{AppendRelInfo, Relids};
use super::restrictinfo::{RestrictInfoArgs, and_args, make_restrictinfo};
use crate::expr::Expr;
use crate::types::Value;

/// expand_appendrel_subquery builds the relation of each member of a UNION ALL subquery's append relation, and of
/// their own members, as Postgres' function of the same name does.
pub fn expand_appendrel_subquery(root: &mut PlannerInfo<'_, '_>, rti: usize) {
    let children: Vec<usize> =
        root.parse.append_rel_list.iter().filter(|a| a.parent_relid == rti).map(|a| a.child_relid).collect();
    for child_rtindex in children {
        super::relnode::build_simple_rel(root, child_rtindex, Some(rti));
        if root.parse.rte(child_rtindex).inh {
            expand_appendrel_subquery(root, child_rtindex);
        }
    }
}

/// apply_child_basequals gives a member of an append relation its parent's restrictions over its own columns,
/// simplified, as Postgres' function of the same name does, returning false when one is always false.
pub fn apply_child_basequals(
    root: &mut PlannerInfo<'_, '_>,
    parentrel: usize,
    childrel: usize,
    appinfo: &AppendRelInfo,
) -> bool {
    let mut childquals = Vec::new();
    let mut cq_min_security = usize::MAX;
    for rinfo in root.rels[parentrel].baserestrictinfo.clone() {
        let clause = root.rinfos[rinfo].clause.clone();
        let childqual = super::appendinfo::adjust_appendrel_attrs(root.glob, clause, appinfo);
        let childqual = super::clauses::eval_const_expressions(root.ctx.get_mut(), childqual);
        match childqual {
            Expr::Const(Value::Bool(true)) => continue,
            Expr::Const(_) => return false,
            _ => {}
        }
        let parent = root.rinfos[rinfo].clone();
        for onecq in and_args(&childqual) {
            let pseudoconstant =
                !super::var::contain_var_clause(onecq) && !super::clauses::contain_volatile_functions(root.glob, onecq);
            if pseudoconstant {
                root.has_pseudo_constant_quals = true;
            }
            let args = RestrictInfoArgs {
                is_pushed_down: parent.is_pushed_down,
                has_clone: parent.has_clone,
                is_clone: parent.is_clone,
                pseudoconstant,
                security_level: parent.security_level,
                required_relids: None,
                incompatible_relids: Relids::new(),
                outer_relids: Relids::new(),
            };
            let childrinfo = make_restrictinfo(root, onecq.clone(), args);
            if restriction_is_always_false(root, &root.rinfos[childrinfo]) {
                return false;
            }
            if restriction_is_always_true(root, &root.rinfos[childrinfo]) {
                continue;
            }
            childquals.push(childrinfo);
            cq_min_security = cq_min_security.min(parent.security_level);
        }
    }
    root.rels[childrel].baserestrictinfo = childquals;
    root.rels[childrel].baserestrict_min_security = cq_min_security;
    true
}
