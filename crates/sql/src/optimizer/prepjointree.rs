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

//! The parts of Postgres' optimizer/prep/prepjointree.c that prepare the join tree: reduce_outer_joins.

use super::clauses::{find_forced_null_vars, find_nonnullable_rels, find_nonnullable_vars};
use super::nodes::{FromExpr, JoinTreeNode, JoinType, Relids, overlap, singleton, var_parts};

/// ReduceState is what the first pass of reduce_outer_joins learns about a join tree node: the relations under it,
/// whether an outer join is under it, and the same of its children.
struct ReduceState {
    relids: Relids,
    contains_outer: bool,
    sub_states: Vec<ReduceState>,
}

/// reduce_outer_joins turns outer joins into inner joins where a stricter condition above rejects the rows that
/// they pad with NULLs, and left joins into anti joins where a condition above keeps only those rows, as Postgres'
/// reduce_outer_joins does.
pub fn reduce_outer_joins(jointree: &mut FromExpr) {
    let state = pass1_from(jointree);
    if state.contains_outer {
        pass2_from(jointree, &state, 0, &[], &[]);
    }
}

/// pass1_from is reduce_outer_joins_pass1 for a FROM list.
fn pass1_from(f: &FromExpr) -> ReduceState {
    let sub_states: Vec<ReduceState> = f.fromlist.iter().map(pass1).collect();
    ReduceState {
        relids: sub_states.iter().fold(0, |relids, s| relids | s.relids),
        contains_outer: sub_states.iter().any(|s| s.contains_outer),
        sub_states,
    }
}

/// pass1 is reduce_outer_joins_pass1, which finds the relations and outer joins under a join tree node.
fn pass1(node: &JoinTreeNode) -> ReduceState {
    match node {
        JoinTreeNode::Rel(varno) => {
            ReduceState { relids: singleton(*varno), contains_outer: false, sub_states: Vec::new() }
        }
        JoinTreeNode::From(f) => pass1_from(f),
        JoinTreeNode::Join(j) => {
            let sub_states = vec![pass1(&j.larg), pass1(&j.rarg)];
            ReduceState {
                relids: sub_states[0].relids | sub_states[1].relids,
                contains_outer: j.jointype.is_outer() || sub_states.iter().any(|s| s.contains_outer),
                sub_states,
            }
        }
    }
}

/// pass2_from is reduce_outer_joins_pass2 for a FROM list, given the relations and Vars that conditions above make
/// non-nullable and the Vars they force to be NULL.
fn pass2_from(
    f: &mut FromExpr,
    state: &ReduceState,
    nonnullable_rels: Relids,
    nonnullable_vars: &[usize],
    forced_null_vars: &[usize],
) {
    let pass_nonnullable_rels = f.quals.iter().fold(nonnullable_rels, |relids, q| relids | find_nonnullable_rels(q));
    let mut pass_nonnullable_vars: Vec<usize> = f.quals.iter().flat_map(find_nonnullable_vars).collect();
    pass_nonnullable_vars.extend_from_slice(nonnullable_vars);
    let mut pass_forced_null_vars: Vec<usize> = f.quals.iter().flat_map(find_forced_null_vars).collect();
    pass_forced_null_vars.extend_from_slice(forced_null_vars);
    for (node, sub_state) in f.fromlist.iter_mut().zip(&state.sub_states) {
        if sub_state.contains_outer {
            pass2(node, sub_state, pass_nonnullable_rels, &pass_nonnullable_vars, &pass_forced_null_vars);
        }
    }
}

/// pass2 is reduce_outer_joins_pass2, which reduces the outer joins under a join tree node.
fn pass2(
    node: &mut JoinTreeNode,
    state: &ReduceState,
    nonnullable_rels: Relids,
    nonnullable_vars: &[usize],
    forced_null_vars: &[usize],
) {
    let j = match node {
        JoinTreeNode::From(f) => return pass2_from(f, state, nonnullable_rels, nonnullable_vars, forced_null_vars),
        JoinTreeNode::Rel(_) => return,
        JoinTreeNode::Join(j) => j,
    };
    let (mut left_state, mut right_state) = (&state.sub_states[0], &state.sub_states[1]);
    let mut jointype = j.jointype;
    let mut right = false;
    match jointype {
        JoinType::Left if overlap(nonnullable_rels, right_state.relids) => jointype = JoinType::Inner,
        JoinType::Full => {
            match (overlap(nonnullable_rels, left_state.relids), overlap(nonnullable_rels, right_state.relids)) {
                (true, true) => jointype = JoinType::Inner,
                (true, false) => jointype = JoinType::Left,
                (false, true) => right = true,
                (false, false) => {}
            }
        }
        _ => {}
    }
    if right {
        std::mem::swap(&mut j.larg, &mut j.rarg);
        std::mem::swap(&mut left_state, &mut right_state);
        jointype = JoinType::Left;
    }
    let mut local_nonnullable_vars = None;
    if jointype == JoinType::Left {
        let vars: Vec<usize> = j.quals.iter().flat_map(find_nonnullable_vars).collect();
        let forced = vars.iter().filter(|v| forced_null_vars.contains(v));
        if forced.fold(0, |relids, &v| relids | singleton(var_parts(v).0)) & right_state.relids != 0 {
            jointype = JoinType::Anti;
        }
        local_nonnullable_vars = Some(vars);
    }
    j.jointype = jointype;
    if !left_state.contains_outer && !right_state.contains_outer {
        return;
    }
    let (mut local_nonnullable_rels, mut local_forced_null_vars) = (0, Vec::new());
    let mut local_nonnullable_vars = local_nonnullable_vars.unwrap_or_default();
    if jointype != JoinType::Full {
        local_nonnullable_rels = j.quals.iter().fold(0, |relids, q| relids | find_nonnullable_rels(q));
        if local_nonnullable_vars.is_empty() {
            local_nonnullable_vars = j.quals.iter().flat_map(find_nonnullable_vars).collect();
        }
        local_forced_null_vars = j.quals.iter().flat_map(find_forced_null_vars).collect();
        if matches!(jointype, JoinType::Inner | JoinType::Semi) {
            local_nonnullable_rels |= nonnullable_rels;
            local_nonnullable_vars.extend_from_slice(nonnullable_vars);
            local_forced_null_vars.extend_from_slice(forced_null_vars);
        }
    } else {
        local_nonnullable_vars.clear();
    }
    if left_state.contains_outer {
        match jointype {
            JoinType::Inner | JoinType::Semi => {
                pass2(&mut j.larg, left_state, local_nonnullable_rels, &local_nonnullable_vars, &local_forced_null_vars)
            }
            JoinType::Full => pass2(&mut j.larg, left_state, 0, &[], &[]),
            _ => pass2(&mut j.larg, left_state, nonnullable_rels, nonnullable_vars, forced_null_vars),
        }
    }
    if right_state.contains_outer {
        match jointype {
            JoinType::Full => pass2(&mut j.rarg, right_state, 0, &[], &[]),
            _ => pass2(
                &mut j.rarg,
                right_state,
                local_nonnullable_rels,
                &local_nonnullable_vars,
                &local_forced_null_vars,
            ),
        }
    }
}
