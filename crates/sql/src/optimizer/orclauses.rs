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

//! Postgres' optimizer/util/orclauses.c: restrictions of one relation that an OR join clause implies, such as
//! `a.x = 1 OR a.x = 3` from `(a.x = 1 AND b.y = 2) OR (a.x = 3 AND b.y = 4)`, which filter the relation's rows before
//! the join.

use super::PlannerInfo;
use super::clausesel::clause_selectivity;
use super::nodes::{JoinType, RelOptKind, RinfoId};
use super::prepqual::{make_andclause, make_orclause};
use super::restrictinfo::{RestrictInfoArgs, join_clause_is_movable_to, make_restrictinfo, or_args};
use crate::expr::Expr;

/// extract_restriction_or_clauses adds to each base relation the restriction that each of its OR join clauses
/// implies, when it is selective enough to be worth testing, as Postgres' function of the same name does.
pub fn extract_restriction_or_clauses(root: &mut PlannerInfo<'_, '_>) {
    for rti in 1..root.rels.len() {
        if root.rels[rti].reloptkind != RelOptKind::BaseRel || root.rels[rti].relid != rti {
            continue;
        }
        for rinfo in root.rels[rti].joininfo.clone() {
            let r = &root.rinfos[rinfo];
            if r.orclause.is_some()
                && join_clause_is_movable_to(r, &root.rels[rti])
                && let Some(orclause) = extract_or_clause(root, rinfo, rti)
            {
                consider_new_or_clause(root, rti, orclause, rinfo);
            }
        }
    }
}

/// is_safe_restriction_clause_for reports whether a clause reads only a relation and could filter its rows, as
/// Postgres' function of the same name does.
fn is_safe_restriction_clause_for(root: &PlannerInfo<'_, '_>, rinfo: RinfoId, rel: usize) -> bool {
    let r = &root.rinfos[rinfo];
    !r.pseudoconstant
        && r.clause_relids == root.rels[rel].relids
        && !super::clauses::contain_volatile_functions(root.glob, &r.clause)
}

/// extract_or_clause returns the OR of what each argument of an OR clause implies of a relation, from the conjuncts
/// of the argument that read only that relation, or None when an argument implies nothing, as Postgres' function of
/// the same name does.
fn extract_or_clause(root: &PlannerInfo<'_, '_>, or_rinfo: RinfoId, rel: usize) -> Option<Expr> {
    let mut clauselist = Vec::new();
    for orarg in root.rinfos[or_rinfo].orclause.as_ref().expect("an OR clause") {
        let mut subclauses = Vec::new();
        for &rinfo in orarg {
            if root.rinfos[rinfo].orclause.is_some() {
                if let Some(suborclause) = extract_or_clause(root, rinfo, rel) {
                    subclauses.push(suborclause);
                }
            } else if is_safe_restriction_clause_for(root, rinfo, rel) {
                subclauses.push(root.rinfos[rinfo].clause.clone());
            }
        }
        if subclauses.is_empty() {
            return None;
        }
        let subclause = make_andclause(subclauses);
        clauselist.extend(or_args(&subclause).into_iter().cloned());
    }
    (!clauselist.is_empty()).then(|| make_orclause(clauselist))
}

/// consider_new_or_clause adds an OR restriction that an OR join clause implies to a relation unless it keeps most
/// rows, raising the join clause's selectivity so that the join's size counts its restriction once, as Postgres'
/// function of the same name does.
fn consider_new_or_clause(root: &mut PlannerInfo<'_, '_>, rel: usize, orclause: Expr, join_or_rinfo: RinfoId) {
    let security_level = root.rinfos[join_or_rinfo].security_level;
    let args = RestrictInfoArgs {
        is_pushed_down: true,
        has_clone: false,
        is_clone: false,
        pseudoconstant: false,
        security_level,
        required_relids: None,
        incompatible_relids: Default::default(),
        outer_relids: Default::default(),
    };
    let or_rinfo = make_restrictinfo(root, orclause, args);
    let r = &root.rinfos[or_rinfo];
    let or_selec = clause_selectivity(root, &r.clause, Some(r), 0, JoinType::Inner, None);
    if or_selec > 0.9 {
        return;
    }
    root.rels[rel].baserestrictinfo.push(or_rinfo);
    root.rels[rel].baserestrict_min_security = root.rels[rel].baserestrict_min_security.min(security_level);
    if or_selec > 0.0 {
        let join_or = &root.rinfos[join_or_rinfo];
        let sjinfo = super::joinrels::init_dummy_sjinfo(
            &join_or.clause_relids.difference(&root.rels[rel].relids),
            &root.rels[rel].relids,
        );
        let orig_selec = clause_selectivity(root, &join_or.clause, Some(join_or), 0, JoinType::Inner, Some(&sjinfo));
        join_or.norm_selec.set((orig_selec / or_selec).min(1.0));
    }
}
