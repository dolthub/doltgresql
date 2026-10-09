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

//! Postgres' optimizer/path/equivclass.c: equivalence classes, the sets of expressions that the query's equalities
//! make equal, from which the planner derives the equalities it needs at each relation and join. Doltgres has no
//! inheritance or partitioning, so classes never have child members.

use super::PlannerInfo;
use super::initsplan::{
    add_vars_to_targetlist, build_implied_join_equality, distribute_restrictinfo_to_rels, process_implied_equality,
};
use super::nodes::{
    EcId, EmId, EquivalenceClass, EquivalenceMember, OuterJoinClauseInfo, RelOptKind, Relids, RinfoId, VarNode,
};
use super::restrictinfo::{RestrictInfoArgs, binary_op_args, make_restrictinfo};
use super::var::{pull_var_clause, remove_nulling_relids_fn};
use crate::expr::{CmpOp, Expr};

impl PlannerInfo<'_, '_> {
    /// eq_class_ids returns the IDs of the equivalence classes that were not merged into others, in the order of
    /// Postgres' eq_classes list.
    pub fn eq_class_ids(&self) -> Vec<EcId> {
        (0..self.eq_classes.len()).filter(|&i| self.eq_classes[i].ec_merged.is_none()).collect()
    }

    /// canonical_ec returns the class that an equivalence class was merged into, or the class itself, as Postgres'
    /// update_mergeclause_eclasses follows ec_merged.
    pub fn canonical_ec(&self, mut ec: EcId) -> EcId {
        while let Some(merged) = self.eq_classes[ec].ec_merged {
            ec = merged;
        }
        ec
    }
}

/// process_equivalence adds the sides of a mergejoinable equality to the equivalence classes, merging classes that
/// it makes equal, and reports whether the clause went into a class rather than being distributed on its own, as
/// Postgres' function of the same name does. An equality of an expression with itself becomes an IS NOT NULL test.
pub fn process_equivalence(root: &mut PlannerInfo<'_, '_>, rinfo: &mut RinfoId, jdomain: usize) -> bool {
    let restrictinfo = root.rinfos[*rinfo].clone();
    let Some((item1, item2)) = binary_op_args(&restrictinfo.clause) else { return false };
    let (item1, item2) = (item1.clone(), item2.clone());
    if item1 == item2 {
        let ntest = Expr::IsNull(Box::new(item1), true);
        *rinfo = make_restrictinfo(
            root,
            ntest,
            RestrictInfoArgs {
                is_pushed_down: restrictinfo.is_pushed_down,
                has_clone: restrictinfo.has_clone,
                is_clone: restrictinfo.is_clone,
                pseudoconstant: restrictinfo.pseudoconstant,
                security_level: restrictinfo.security_level,
                required_relids: None,
                incompatible_relids: restrictinfo.incompatible_relids.clone(),
                outer_relids: restrictinfo.outer_relids.clone(),
            },
        );
        return false;
    }
    let item1_type = super::nodefuncs::expr_type(root, &item1).unwrap_or(0);
    let item2_type = super::nodefuncs::expr_type(root, &item2).unwrap_or(0);
    let opfamilies = restrictinfo.mergeopfamilies.clone();
    let (mut ec1, mut ec2, mut em1, mut em2) = (None, None, None, None);
    for cur_ec in root.eq_class_ids() {
        let ec = &root.eq_classes[cur_ec];
        if ec.ec_has_volatile || ec.ec_opfamilies != opfamilies {
            continue;
        }
        for &cur_em in &ec.ec_members {
            let em = &root.eq_members[cur_em];
            if em.em_is_const && em.em_jdomain != jdomain {
                continue;
            }
            if ec1.is_none() && item1_type == em.em_datatype && item1 == em.em_expr {
                (ec1, em1) = (Some(cur_ec), Some(cur_em));
                if ec2.is_some() {
                    break;
                }
            }
            if ec2.is_none() && item2_type == em.em_datatype && item2 == em.em_expr {
                (ec2, em2) = (Some(cur_ec), Some(cur_em));
                if ec1.is_some() {
                    break;
                }
            }
        }
        if ec1.is_some() && ec2.is_some() {
            break;
        }
    }
    let security_level = restrictinfo.security_level;
    let add_source = |root: &mut PlannerInfo<'_, '_>, ec: EcId, rinfo: RinfoId| {
        let class = &mut root.eq_classes[ec];
        class.ec_sources.push(rinfo);
        class.ec_min_security = class.ec_min_security.min(security_level);
        class.ec_max_security = class.ec_max_security.max(security_level);
    };
    let (ec, em1, em2) = match (ec1, ec2) {
        (Some(ec1), Some(ec2)) if ec1 == ec2 => {
            add_source(root, ec1, *rinfo);
            (ec1, em1, em2)
        }
        (Some(ec1), Some(ec2)) => {
            assert!(!root.ec_merging_done, "too late to merge equivalence classes");
            let merged = std::mem::replace(&mut root.eq_classes[ec2], empty_class(Vec::new()));
            let class = &mut root.eq_classes[ec1];
            class.ec_members.extend(merged.ec_members);
            class.ec_sources.extend(merged.ec_sources);
            class.ec_derives.extend(merged.ec_derives);
            class.ec_relids.add_members(&merged.ec_relids);
            class.ec_has_const |= merged.ec_has_const;
            class.ec_min_security = class.ec_min_security.min(merged.ec_min_security);
            class.ec_max_security = class.ec_max_security.max(merged.ec_max_security);
            root.eq_classes[ec2].ec_merged = Some(ec1);
            add_source(root, ec1, *rinfo);
            (ec1, em1, em2)
        }
        (Some(ec1), None) => {
            let em2 = add_eq_member(root, ec1, item2, restrictinfo.right_relids.clone(), jdomain, item2_type);
            add_source(root, ec1, *rinfo);
            (ec1, em1, Some(em2))
        }
        (None, Some(ec2)) => {
            let em1 = add_eq_member(root, ec2, item1, restrictinfo.left_relids.clone(), jdomain, item1_type);
            add_source(root, ec2, *rinfo);
            (ec2, Some(em1), em2)
        }
        (None, None) => {
            let mut class = empty_class(opfamilies);
            class.ec_sources.push(*rinfo);
            class.ec_min_security = security_level;
            class.ec_max_security = security_level;
            root.eq_classes.push(class);
            let ec = root.eq_classes.len() - 1;
            let em1 = add_eq_member(root, ec, item1, restrictinfo.left_relids.clone(), jdomain, item1_type);
            let em2 = add_eq_member(root, ec, item2, restrictinfo.right_relids.clone(), jdomain, item2_type);
            (ec, Some(em1), Some(em2))
        }
    };
    let r = &mut root.rinfos[*rinfo];
    (r.left_ec, r.right_ec, r.left_em, r.right_em) = (Some(ec), Some(ec), em1, em2);
    true
}

/// empty_class returns an equivalence class without members of operator families.
fn empty_class(ec_opfamilies: Vec<u32>) -> EquivalenceClass {
    EquivalenceClass {
        ec_opfamilies,
        ec_members: Vec::new(),
        ec_sources: Vec::new(),
        ec_derives: Vec::new(),
        ec_derives_hash: std::collections::HashMap::new(),
        ec_relids: Relids::new(),
        ec_has_const: false,
        ec_has_volatile: false,
        ec_broken: false,
        ec_sortref: 0,
        ec_min_security: usize::MAX,
        ec_max_security: 0,
        ec_merged: None,
    }
}

/// add_eq_member adds a member to an equivalence class, which is a constant when it reads no relation, as Postgres'
/// add_eq_member and make_eq_member do.
fn add_eq_member(
    root: &mut PlannerInfo<'_, '_>,
    ec: EcId,
    em_expr: Expr,
    em_relids: Relids,
    em_jdomain: usize,
    em_datatype: u32,
) -> EmId {
    let em_is_const = em_relids.is_empty();
    let class = &mut root.eq_classes[ec];
    class.ec_has_const |= em_is_const;
    class.ec_relids.add_members(&em_relids);
    root.eq_members.push(EquivalenceMember { em_expr, em_relids, em_is_const, em_datatype, em_jdomain });
    let em = root.eq_members.len() - 1;
    root.eq_classes[ec].ec_members.push(em);
    em
}

/// generate_base_implied_equalities derives the equalities that each equivalence class implies within a base
/// relation, or between its members and a constant, and records which classes mention each relation, as Postgres'
/// function of the same name does.
pub fn generate_base_implied_equalities(root: &mut PlannerInfo<'_, '_>) {
    root.ec_merging_done = true;
    for ec in root.eq_class_ids() {
        let mut can_generate_joinclause = false;
        if root.eq_classes[ec].ec_members.len() > 1 {
            if root.eq_classes[ec].ec_has_const {
                generate_base_implied_equalities_const(root, ec);
            } else {
                generate_base_implied_equalities_no_const(root, ec);
            }
            if root.eq_classes[ec].ec_broken {
                generate_base_implied_equalities_broken(root, ec);
            }
            can_generate_joinclause = root.eq_classes[ec].ec_relids.num_members() > 1;
        }
        for i in root.eq_classes[ec].ec_relids.clone().members() {
            let rel = &mut root.rels[i];
            if rel.reloptkind != RelOptKind::BaseRel {
                continue;
            }
            rel.eclass_indexes.add_member(ec);
            if can_generate_joinclause {
                rel.has_eclass_joins = true;
            }
        }
    }
}

/// generate_base_implied_equalities_const derives an equality of each member of a class with a constant, as
/// Postgres' function of the same name does.
fn generate_base_implied_equalities_const(root: &mut PlannerInfo<'_, '_>, ec: EcId) {
    let class = &root.eq_classes[ec];
    if class.ec_members.len() == 2 && class.ec_sources.len() == 1 {
        let restrictinfo = class.ec_sources[0];
        distribute_restrictinfo_to_rels(root, restrictinfo);
        return;
    }
    let mut const_em = None;
    for &em in &class.ec_members {
        if root.eq_members[em].em_is_const {
            const_em = Some(em);
            if matches!(root.eq_members[em].em_expr, Expr::Const(_)) {
                break;
            }
        }
    }
    let const_em = const_em.expect("a class with a constant has a constant member");
    for cur_em in root.eq_classes[ec].ec_members.clone() {
        if cur_em == const_em {
            continue;
        }
        let (cur, constant) = (&root.eq_members[cur_em], &root.eq_members[const_em]);
        if !select_equality_operator(root, ec, cur.em_datatype, constant.em_datatype) {
            root.eq_classes[ec].ec_broken = true;
            break;
        }
        let (item1, item2) = (cur.em_expr.clone(), constant.em_expr.clone());
        let qualscope = root.join_domains[constant.em_jdomain].jd_relids.clone();
        let both_const = cur.em_is_const;
        let min_security = root.eq_classes[ec].ec_min_security;
        let rinfo = process_implied_equality(root, item1, item2, &qualscope, min_security, both_const);
        if let Some(rinfo) = rinfo
            && !root.rinfos[rinfo].mergeopfamilies.is_empty()
        {
            let r = &mut root.rinfos[rinfo];
            (r.left_ec, r.right_ec, r.left_em, r.right_em) = (Some(ec), Some(ec), Some(cur_em), Some(const_em));
            ec_add_derived_clause(root, ec, rinfo);
        }
    }
}

/// generate_base_implied_equalities_no_const derives an equality of each pair of consecutive members of a class
/// that read the same one relation, and marks every member's Vars as needed by the class's relations, as Postgres'
/// function of the same name does.
fn generate_base_implied_equalities_no_const(root: &mut PlannerInfo<'_, '_>, ec: EcId) {
    let mut prev_ems: std::collections::HashMap<usize, EmId> = std::collections::HashMap::new();
    for cur_em in root.eq_classes[ec].ec_members.clone() {
        let Some(relid) = root.eq_members[cur_em].em_relids.singleton_member() else { continue };
        if let Some(&prev_em) = prev_ems.get(&relid) {
            let (prev, cur) = (&root.eq_members[prev_em], &root.eq_members[cur_em]);
            if !select_equality_operator(root, ec, prev.em_datatype, cur.em_datatype) {
                root.eq_classes[ec].ec_broken = true;
                break;
            }
            let (item1, item2, qualscope) = (prev.em_expr.clone(), cur.em_expr.clone(), cur.em_relids.clone());
            let min_security = root.eq_classes[ec].ec_min_security;
            let rinfo = process_implied_equality(root, item1, item2, &qualscope, min_security, false);
            if let Some(rinfo) = rinfo
                && !root.rinfos[rinfo].mergeopfamilies.is_empty()
            {
                let r = &mut root.rinfos[rinfo];
                (r.left_ec, r.right_ec, r.left_em, r.right_em) = (Some(ec), Some(ec), Some(prev_em), Some(cur_em));
            }
        }
        prev_ems.insert(relid, cur_em);
    }
    let relids = root.eq_classes[ec].ec_relids.clone();
    for cur_em in root.eq_classes[ec].ec_members.clone() {
        let vars = pull_var_clause(root.glob, &root.eq_members[cur_em].em_expr, true);
        add_vars_to_targetlist(root, &vars, &relids);
    }
}

/// generate_base_implied_equalities_broken distributes the clauses that a class came from, when the class cannot
/// derive equalities, as Postgres' function of the same name does.
fn generate_base_implied_equalities_broken(root: &mut PlannerInfo<'_, '_>, ec: EcId) {
    for restrictinfo in root.eq_classes[ec].ec_sources.clone() {
        if root.eq_classes[ec].ec_has_const || root.rinfos[restrictinfo].required_relids.num_members() <= 1 {
            distribute_restrictinfo_to_rels(root, restrictinfo);
        }
    }
}

/// generate_join_implied_equalities returns the equalities that the equivalence classes imply between the two sides
/// of a join, as Postgres' function of the same name does.
pub fn generate_join_implied_equalities(
    root: &mut PlannerInfo<'_, '_>,
    join_relids: &Relids,
    outer_relids: &Relids,
    inner_rel: usize,
    ojrelid: usize,
) -> Vec<RinfoId> {
    let inner_relids = root.rels[inner_rel].relids.clone();
    let matching_ecs = match ojrelid {
        0 => get_common_eclass_indexes(root, &inner_relids, outer_relids),
        _ => get_eclass_indexes_for_relids(root, join_relids),
    };
    let mut result = Vec::new();
    for ec in matching_ecs.members() {
        if root.eq_classes[ec].ec_has_const || root.eq_classes[ec].ec_members.len() <= 1 {
            continue;
        }
        let mut sublist = Vec::new();
        if !root.eq_classes[ec].ec_broken {
            sublist = generate_join_implied_equalities_normal(root, ec, join_relids, outer_relids, &inner_relids);
        }
        if root.eq_classes[ec].ec_broken {
            sublist = generate_join_implied_equalities_broken(root, ec, join_relids, outer_relids, &inner_relids);
        }
        result.extend(sublist);
    }
    result
}

/// generate_join_implied_equalities_normal derives one equality between the members of a class on the two sides of
/// a join, preferring Vars and hashable equalities, and equalities that tie in members that only the join can
/// compute, as Postgres' function of the same name does.
fn generate_join_implied_equalities_normal(
    root: &mut PlannerInfo<'_, '_>,
    ec: EcId,
    join_relids: &Relids,
    outer_relids: &Relids,
    inner_relids: &Relids,
) -> Vec<RinfoId> {
    let (mut new_members, mut outer_members, mut inner_members) = (Vec::new(), Vec::new(), Vec::new());
    for &cur_em in &root.eq_classes[ec].ec_members {
        let relids = &root.eq_members[cur_em].em_relids;
        if !relids.is_subset(join_relids) {
            continue;
        }
        if relids.is_subset(outer_relids) {
            outer_members.push(cur_em);
        } else if relids.is_subset(inner_relids) {
            inner_members.push(cur_em);
        } else {
            new_members.push(cur_em);
        }
    }
    let mut result = Vec::new();
    if !outer_members.is_empty() && !inner_members.is_empty() {
        let mut best: Option<(EmId, EmId)> = None;
        let mut best_score = -1;
        'outer: for &outer_em in &outer_members {
            for &inner_em in &inner_members {
                let (o, i) = (&root.eq_members[outer_em], &root.eq_members[inner_em]);
                if !select_equality_operator(root, ec, o.em_datatype, i.em_datatype) {
                    continue;
                }
                let is_var = |e: &Expr| matches!(e, Expr::Column(id) if matches!(root.glob.node(*id), VarNode::Var(_)));
                let score = 1 + i32::from(is_var(&o.em_expr)) + i32::from(is_var(&i.em_expr));
                if score > best_score {
                    best = Some((outer_em, inner_em));
                    best_score = score;
                    if best_score == 3 {
                        break 'outer;
                    }
                }
            }
        }
        let Some((best_outer_em, best_inner_em)) = best else {
            root.eq_classes[ec].ec_broken = true;
            return Vec::new();
        };
        result.push(create_join_clause(root, ec, best_outer_em, best_inner_em, Some(ec)));
    }
    if !new_members.is_empty() {
        if let Some(&first) = outer_members.first().or(inner_members.first()) {
            new_members.push(first);
        }
        let mut prev_em: Option<EmId> = None;
        for cur_em in new_members {
            if let Some(prev_em) = prev_em {
                let (p, c) = (&root.eq_members[prev_em], &root.eq_members[cur_em]);
                if !select_equality_operator(root, ec, p.em_datatype, c.em_datatype) {
                    root.eq_classes[ec].ec_broken = true;
                    return Vec::new();
                }
                result.push(create_join_clause(root, ec, prev_em, cur_em, None));
            }
            prev_em = Some(cur_em);
        }
    }
    result
}

/// generate_join_implied_equalities_broken returns the clauses that a broken class came from that a join must
/// evaluate, as Postgres' function of the same name does.
fn generate_join_implied_equalities_broken(
    root: &PlannerInfo<'_, '_>,
    ec: EcId,
    join_relids: &Relids,
    outer_relids: &Relids,
    inner_relids: &Relids,
) -> Vec<RinfoId> {
    root.eq_classes[ec]
        .ec_sources
        .iter()
        .copied()
        .filter(|&r| {
            let clause_relids = &root.rinfos[r].required_relids;
            clause_relids.is_subset(join_relids)
                && !clause_relids.is_subset(outer_relids)
                && !clause_relids.is_subset(inner_relids)
        })
        .collect()
}

/// select_equality_operator reports whether a class's operator families have an equality of two types, as Postgres'
/// function of the same name finds one. A family here is the btree family of both types, so each of the class's
/// members has it.
fn select_equality_operator(root: &PlannerInfo<'_, '_>, ec: EcId, lefttype: u32, righttype: u32) -> bool {
    let families = &root.eq_classes[ec].ec_opfamilies;
    let family = |t| super::nodefuncs::btree_opfamily(t);
    families.iter().any(|f| family(lefttype) == Some(*f) && family(righttype) == Some(*f))
}

/// create_join_clause returns the RestrictInfo of an equality of two members of a class, building it unless the
/// class already has one, as Postgres' function of the same name does.
fn create_join_clause(
    root: &mut PlannerInfo<'_, '_>,
    ec: EcId,
    leftem: EmId,
    rightem: EmId,
    parent_ec: Option<EcId>,
) -> RinfoId {
    if let Some(rinfo) = ec_search_clause_for_ems(root, ec, leftem, Some(rightem), parent_ec) {
        return rinfo;
    }
    let (left, right) = (&root.eq_members[leftem], &root.eq_members[rightem]);
    let clause = Expr::Compare(CmpOp::Eq, Box::new(left.em_expr.clone()), Box::new(right.em_expr.clone()));
    let qualscope = left.em_relids.union(&right.em_relids);
    let min_security = root.eq_classes[ec].ec_min_security;
    let rinfo = build_implied_join_equality(root, clause, qualscope, min_security);
    let r = &mut root.rinfos[rinfo];
    r.parent_ec = parent_ec;
    (r.left_ec, r.right_ec, r.left_em, r.right_em) = (Some(ec), Some(ec), Some(leftem), Some(rightem));
    ec_add_derived_clause(root, ec, rinfo);
    rinfo
}

/// reconsider_outer_join_clauses derives the equalities that outer join clauses imply with constants of the
/// equivalence classes, distributing the clauses that remain, as Postgres' function of the same name does.
pub fn reconsider_outer_join_clauses(root: &mut PlannerInfo<'_, '_>) {
    loop {
        let mut found = false;
        for side in 0..3 {
            let mut i = 0;
            while i < root.outer_join_clauses(side).len() {
                let ojcinfo = root.outer_join_clauses(side)[i].clone();
                let reconsidered = match side {
                    2 => reconsider_full_join_clause(root, &ojcinfo),
                    _ => reconsider_outer_join_clause(root, &ojcinfo, side == 0),
                };
                if !reconsidered {
                    i += 1;
                    continue;
                }
                found = true;
                root.outer_join_clauses_mut(side).remove(i);
                let rinfo = root.rinfos[ojcinfo.rinfo].clone();
                let rinfo = make_restrictinfo(
                    root,
                    Expr::Const(crate::types::Value::Bool(true)),
                    RestrictInfoArgs {
                        is_pushed_down: rinfo.is_pushed_down,
                        has_clone: rinfo.has_clone,
                        is_clone: rinfo.is_clone,
                        pseudoconstant: false,
                        security_level: 0,
                        required_relids: Some(rinfo.required_relids.clone()),
                        incompatible_relids: rinfo.incompatible_relids.clone(),
                        outer_relids: rinfo.outer_relids.clone(),
                    },
                );
                distribute_restrictinfo_to_rels(root, rinfo);
            }
        }
        if !found {
            break;
        }
    }
    for side in 0..3 {
        for ojcinfo in root.outer_join_clauses(side).clone() {
            distribute_restrictinfo_to_rels(root, ojcinfo.rinfo);
        }
    }
}

impl PlannerInfo<'_, '_> {
    /// outer_join_clauses returns the left, right, or full join clauses, by 0, 1, or 2.
    fn outer_join_clauses(&self, side: usize) -> &Vec<OuterJoinClauseInfo> {
        match side {
            0 => &self.left_join_clauses,
            1 => &self.right_join_clauses,
            _ => &self.full_join_clauses,
        }
    }

    /// outer_join_clauses_mut is outer_join_clauses for changing them.
    fn outer_join_clauses_mut(&mut self, side: usize) -> &mut Vec<OuterJoinClauseInfo> {
        match side {
            0 => &mut self.left_join_clauses,
            1 => &mut self.right_join_clauses,
            _ => &mut self.full_join_clauses,
        }
    }
}

/// reconsider_outer_join_clause derives an equality of an outer join clause's inner side with each constant of the
/// class of its outer side, reporting whether it did, as Postgres' function of the same name does.
fn reconsider_outer_join_clause(
    root: &mut PlannerInfo<'_, '_>,
    ojcinfo: &OuterJoinClauseInfo,
    outer_on_left: bool,
) -> bool {
    let rinfo = root.rinfos[ojcinfo.rinfo].clone();
    let Some((left, right)) = binary_op_args(&rinfo.clause) else { return false };
    let (outervar, innervar, inner_relids) = match outer_on_left {
        true => (left.clone(), right.clone(), rinfo.right_relids.clone()),
        false => (right.clone(), left.clone(), rinfo.left_relids.clone()),
    };
    let inner_datatype = super::nodefuncs::expr_type(root, &innervar).unwrap_or(0);
    let syn_righthand = root.sjinfos[ojcinfo.sjinfo].syn_righthand.clone();
    for cur_ec in root.eq_class_ids() {
        let class = &root.eq_classes[cur_ec];
        if !class.ec_has_const || class.ec_has_volatile || class.ec_opfamilies != rinfo.mergeopfamilies {
            continue;
        }
        if !class.ec_members.iter().any(|&em| root.eq_members[em].em_expr == outervar) {
            continue;
        }
        let mut matched = false;
        for cur_em in class.ec_members.clone() {
            let em = &root.eq_members[cur_em];
            if !em.em_is_const || !select_equality_operator(root, cur_ec, inner_datatype, em.em_datatype) {
                continue;
            }
            let clause = Expr::Compare(CmpOp::Eq, Box::new(innervar.clone()), Box::new(em.em_expr.clone()));
            let min_security = root.eq_classes[cur_ec].ec_min_security;
            let mut newrinfo = build_implied_join_equality(root, clause, inner_relids.clone(), min_security);
            let jdomain = find_join_domain(root, &syn_righthand);
            if process_equivalence(root, &mut newrinfo, jdomain) {
                matched = true;
            }
        }
        return matched;
    }
    false
}

/// reconsider_full_join_clause derives equalities of each side of a FULL JOIN clause with the constants of the
/// class of the COALESCE of its sides, reporting whether it did, as Postgres' function of the same name does.
fn reconsider_full_join_clause(root: &mut PlannerInfo<'_, '_>, ojcinfo: &OuterJoinClauseInfo) -> bool {
    let rinfo = root.rinfos[ojcinfo.rinfo].clone();
    let Some((leftvar, rightvar)) = binary_op_args(&rinfo.clause) else { return false };
    let (leftvar, rightvar) = (leftvar.clone(), rightvar.clone());
    let left_type = super::nodefuncs::expr_type(root, &leftvar).unwrap_or(0);
    let right_type = super::nodefuncs::expr_type(root, &rightvar).unwrap_or(0);
    let sjinfo = root.sjinfos[ojcinfo.sjinfo].clone();
    let fjrelids = Relids::singleton(sjinfo.ojrelid);
    for cur_ec in root.eq_class_ids() {
        let class = &root.eq_classes[cur_ec];
        if !class.ec_has_const || class.ec_has_volatile || class.ec_opfamilies != rinfo.mergeopfamilies {
            continue;
        }
        let mut coal_idx = None;
        for (idx, &em) in class.ec_members.clone().iter().enumerate() {
            let Expr::Coalesce(args) = root.eq_members[em].em_expr.clone() else { continue };
            let Ok([cfirst, csecond]) = <[Expr; 2]>::try_from(args) else { continue };
            let mut strip = remove_nulling_relids_fn(root.glob, &fjrelids, &Relids::new());
            let (cfirst, csecond) = (strip(cfirst), strip(csecond));
            if leftvar == cfirst && rightvar == csecond {
                coal_idx = Some(idx);
                break;
            }
        }
        let Some(coal_idx) = coal_idx else { continue };
        let (mut matchleft, mut matchright) = (false, false);
        for cur_em in root.eq_classes[cur_ec].ec_members.clone() {
            let em = root.eq_members[cur_em].clone();
            if !em.em_is_const {
                continue;
            }
            let min_security = root.eq_classes[cur_ec].ec_min_security;
            if select_equality_operator(root, cur_ec, left_type, em.em_datatype) {
                let clause = Expr::Compare(CmpOp::Eq, Box::new(leftvar.clone()), Box::new(em.em_expr.clone()));
                let mut newrinfo = build_implied_join_equality(root, clause, rinfo.left_relids.clone(), min_security);
                let jdomain = find_join_domain(root, &sjinfo.syn_lefthand);
                matchleft |= process_equivalence(root, &mut newrinfo, jdomain);
            }
            if select_equality_operator(root, cur_ec, right_type, em.em_datatype) {
                let clause = Expr::Compare(CmpOp::Eq, Box::new(rightvar.clone()), Box::new(em.em_expr.clone()));
                let mut newrinfo = build_implied_join_equality(root, clause, rinfo.right_relids.clone(), min_security);
                let jdomain = find_join_domain(root, &sjinfo.syn_righthand);
                matchright |= process_equivalence(root, &mut newrinfo, jdomain);
            }
        }
        if matchleft && matchright {
            root.eq_classes[cur_ec].ec_members.remove(coal_idx);
            return true;
        }
        break;
    }
    false
}

/// find_join_domain returns the first join domain within a set of relations, as Postgres' function of the same name
/// does.
fn find_join_domain(root: &PlannerInfo<'_, '_>, relids: &Relids) -> usize {
    root.join_domains.iter().position(|jd| jd.jd_relids.is_subset(relids)).expect("an appropriate JoinDomain")
}

/// have_relevant_eclass_joinclause reports whether an equivalence class can give a join clause between two
/// relations, as Postgres' function of the same name does.
pub fn have_relevant_eclass_joinclause(root: &PlannerInfo<'_, '_>, rel1: usize, rel2: usize) -> bool {
    let matching_ecs = get_common_eclass_indexes(root, &root.rels[rel1].relids, &root.rels[rel2].relids);
    matching_ecs.members().any(|ec| root.eq_classes[ec].ec_members.len() > 1)
}

/// has_relevant_eclass_joinclause reports whether an equivalence class can give a join clause between a relation
/// and some other, as Postgres' function of the same name does.
pub fn has_relevant_eclass_joinclause(root: &PlannerInfo<'_, '_>, rel1: usize) -> bool {
    let relids = &root.rels[rel1].relids;
    get_eclass_indexes_for_relids(root, relids).members().any(|ec| {
        let class = &root.eq_classes[ec];
        class.ec_members.len() > 1 && !class.ec_relids.is_subset(relids)
    })
}

/// get_eclass_indexes_for_relids returns the equivalence classes that mention any of a set of relations, as Postgres'
/// function of the same name does.
fn get_eclass_indexes_for_relids(root: &PlannerInfo<'_, '_>, relids: &Relids) -> Relids {
    let mut ec_indexes = Relids::new();
    for i in relids.members() {
        let rel = &root.rels[i];
        if rel.reloptkind == RelOptKind::BaseRel {
            ec_indexes.add_members(&rel.eclass_indexes);
        }
    }
    ec_indexes
}

/// get_common_eclass_indexes returns the equivalence classes that mention both of two sets of relations, as Postgres'
/// function of the same name does.
fn get_common_eclass_indexes(root: &PlannerInfo<'_, '_>, relids1: &Relids, relids2: &Relids) -> Relids {
    get_eclass_indexes_for_relids(root, relids1).intersect(&get_eclass_indexes_for_relids(root, relids2))
}

/// derives_key returns the key of a derived clause between two members, or between a member and a constant, as
/// Postgres' fill_ec_derives_key does.
fn derives_key(leftem: EmId, rightem: Option<EmId>, parent_ec: Option<EcId>) -> (Option<EmId>, EmId, Option<EcId>) {
    match rightem {
        None => (None, leftem, parent_ec),
        Some(rightem) => (Some(leftem.min(rightem)), leftem.max(rightem), parent_ec),
    }
}

/// ec_add_derived_clause records a clause derived from a class, as Postgres' function of the same name does.
fn ec_add_derived_clause(root: &mut PlannerInfo<'_, '_>, ec: EcId, rinfo: RinfoId) {
    let r = &root.rinfos[rinfo];
    let right_em = r.right_em.expect("a derived clause has members");
    let rightem = (!root.eq_members[right_em].em_is_const).then_some(right_em);
    let key = derives_key(r.left_em.expect("a derived clause has members"), rightem, r.parent_ec);
    let class = &mut root.eq_classes[ec];
    class.ec_derives.push(rinfo);
    class.ec_derives_hash.insert(key, rinfo);
}

/// ec_search_clause_for_ems returns a clause that a class came from or derived between two members, as Postgres'
/// function of the same name does.
fn ec_search_clause_for_ems(
    root: &PlannerInfo<'_, '_>,
    ec: EcId,
    leftem: EmId,
    rightem: Option<EmId>,
    parent_ec: Option<EcId>,
) -> Option<RinfoId> {
    for &rinfo in &root.eq_classes[ec].ec_sources {
        let r = &root.rinfos[rinfo];
        if r.parent_ec == parent_ec
            && ((r.left_em == Some(leftem) && r.right_em == rightem)
                || (r.left_em == rightem && r.right_em == Some(leftem)))
        {
            return Some(rinfo);
        }
    }
    root.eq_classes[ec].ec_derives_hash.get(&derives_key(leftem, rightem, parent_ec)).copied()
}

/// get_eclass_for_sort_expr returns the equivalence class of an expression sorted by operator families, building a
/// class of the expression alone when none has it and asked to, as Postgres' function of the same name does.
pub fn get_eclass_for_sort_expr(
    root: &mut PlannerInfo<'_, '_>,
    expr: Expr,
    opfamilies: &[u32],
    opcintype: u32,
    sortref: usize,
    create_it: bool,
) -> Option<EcId> {
    for cur_ec in root.eq_class_ids() {
        let class = &root.eq_classes[cur_ec];
        if (class.ec_has_volatile && (sortref == 0 || sortref != class.ec_sortref)) || class.ec_opfamilies != opfamilies
        {
            continue;
        }
        for &em in &class.ec_members {
            let em = &root.eq_members[em];
            if em.em_is_const && em.em_jdomain != 0 {
                continue;
            }
            if opcintype == em.em_datatype && expr == em.em_expr {
                return Some(cur_ec);
            }
        }
    }
    if !create_it {
        return None;
    }
    let mut class = empty_class(opfamilies.to_vec());
    class.ec_has_volatile = super::clauses::contain_volatile_functions(root.glob, &expr);
    class.ec_sortref = sortref;
    root.eq_classes.push(class);
    let ec = root.eq_classes.len() - 1;
    let expr_relids = super::var::pull_varnos(root, &expr);
    let newem = add_eq_member(root, ec, expr, expr_relids, 0, opcintype);
    if root.eq_classes[ec].ec_has_const && root.eq_classes[ec].ec_has_volatile {
        root.eq_classes[ec].ec_has_const = false;
        root.eq_members[newem].em_is_const = false;
    }
    if root.ec_merging_done {
        for i in root.eq_classes[ec].ec_relids.clone().members() {
            if root.rels[i].reloptkind == RelOptKind::BaseRel {
                root.rels[i].eclass_indexes.add_member(ec);
            }
        }
    }
    Some(ec)
}

/// eclass_useful_for_merging reports whether an equivalence class has a member that a relation could merge join
/// with, as Postgres' function of the same name does.
pub fn eclass_useful_for_merging(root: &PlannerInfo<'_, '_>, ec: EcId, rel: usize) -> bool {
    let class = &root.eq_classes[ec];
    let relids = &root.rels[rel].relids;
    if class.ec_has_const || class.ec_members.len() <= 1 || class.ec_relids.is_subset(relids) {
        return false;
    }
    class.ec_members.iter().any(|&em| !root.eq_members[em].em_relids.overlap(relids))
}

/// generate_implied_equalities_for_column returns the join clauses that equate the first member of a base relation
/// that a test accepts, in the first class that has one, with each member of other relations outside the prohibited
/// ones, as Postgres' function of the same name does for an index column.
pub fn generate_implied_equalities_for_column(
    root: &mut PlannerInfo<'_, '_>,
    rel: usize,
    callback: &dyn Fn(&PlannerInfo<'_, '_>, EcId, EmId) -> bool,
    prohibited_rels: &Relids,
) -> Vec<RinfoId> {
    let mut result = Vec::new();
    let relids = root.rels[rel].relids.clone();
    for cur_ec in root.rels[rel].eclass_indexes.clone().members() {
        let class = &root.eq_classes[cur_ec];
        if class.ec_has_const || class.ec_members.len() <= 1 {
            continue;
        }
        let members = class.ec_members.clone();
        let Some(cur_em) =
            members.iter().copied().find(|&em| root.eq_members[em].em_relids == relids && callback(root, cur_ec, em))
        else {
            continue;
        };
        for other_em in members {
            let other = &root.eq_members[other_em];
            if other_em == cur_em || other.em_relids.overlap(&relids) || other.em_relids.overlap(prohibited_rels) {
                continue;
            }
            let (cur_type, other_type) = (root.eq_members[cur_em].em_datatype, other.em_datatype);
            if !select_equality_operator(root, cur_ec, cur_type, other_type) {
                continue;
            }
            result.push(create_join_clause(root, cur_ec, cur_em, other_em, Some(cur_ec)));
        }
        if !result.is_empty() {
            break;
        }
    }
    result
}
