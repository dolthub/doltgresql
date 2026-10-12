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

//! Postgres' optimizer/util/joininfo.c: the join clauses of base relations.

use super::PlannerInfo;
use super::equivclass::have_relevant_eclass_joinclause;
use super::initsplan::{restriction_is_always_false, restriction_is_always_true};
use super::nodes::{RelOptKind, Relids, RinfoId};
use super::restrictinfo::{RestrictInfoArgs, make_restrictinfo};
use crate::expr::Expr;

/// have_relevant_joinclause reports whether a join clause or equivalence class connects two relations, as Postgres'
/// function of the same name does.
pub fn have_relevant_joinclause(root: &PlannerInfo<'_, '_>, rel1: usize, rel2: usize) -> bool {
    let (r1, r2) = (&root.rels[rel1], &root.rels[rel2]);
    let (joininfo, other_relids) = match r1.joininfo.len() <= r2.joininfo.len() {
        true => (&r1.joininfo, &r2.relids),
        false => (&r2.joininfo, &r1.relids),
    };
    joininfo.iter().any(|&r| other_relids.overlap(&root.rinfos[r].required_relids))
        || (r1.has_eclass_joins && r2.has_eclass_joins && have_relevant_eclass_joinclause(root, rel1, rel2))
}

/// add_join_clause_to_rels adds a join clause to the join clauses of each base relation it needs, dropping one that
/// NOT NULL columns make always true and making one that they make always false the constant false, as Postgres'
/// function of the same name does.
pub fn add_join_clause_to_rels(root: &mut PlannerInfo<'_, '_>, mut restrictinfo: RinfoId, join_relids: &Relids) {
    if restriction_is_always_true(root, &root.rinfos[restrictinfo]) {
        return;
    }
    if restriction_is_always_false(root, &root.rinfos[restrictinfo]) {
        let r = root.rinfos[restrictinfo].clone();
        let (save_rinfo_serial, save_last_rinfo_serial) = (r.rinfo_serial, root.last_rinfo_serial);
        restrictinfo = make_restrictinfo(
            root,
            Expr::Const(crate::types::Value::Bool(false)),
            RestrictInfoArgs {
                is_pushed_down: r.is_pushed_down,
                has_clone: r.has_clone,
                is_clone: r.is_clone,
                pseudoconstant: r.pseudoconstant,
                security_level: 0,
                required_relids: Some(r.required_relids.clone()),
                incompatible_relids: r.incompatible_relids.clone(),
                outer_relids: r.outer_relids.clone(),
            },
        );
        root.rinfos[restrictinfo].rinfo_serial = save_rinfo_serial;
        root.last_rinfo_serial = save_last_rinfo_serial;
    }
    for relid in join_relids.members() {
        let rel = &mut root.rels[relid];
        if rel.reloptkind == RelOptKind::BaseRel {
            rel.joininfo.push(restrictinfo);
        }
    }
}
