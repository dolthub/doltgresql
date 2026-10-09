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

//! Postgres' optimizer/path/pathkeys.c: the canonical pathkeys that describe the order of a path's rows, and which
//! orders are useful.

use super::PlannerInfo;
use super::equivclass::{eclass_useful_for_merging, get_eclass_for_sort_expr};
use super::nodes::{EcId, PathKey, PkId, RinfoId};
use super::restrictinfo::binary_op_args;
use crate::plan::SortKey;

/// PathKeysComparison is how the orders of two paths' rows compare, as Postgres' PathKeysComparison is.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum PathKeysComparison {
    Equal,
    /// The first order sorts by the second's keys and then more.
    Better1,
    /// The second order sorts by the first's keys and then more.
    Better2,
    Different,
}

/// make_canonical_pathkey returns the canonical pathkey of an equivalence class sorted by an operator family in a
/// direction, building it when it is new, as Postgres' function of the same name does.
pub fn make_canonical_pathkey(
    root: &mut PlannerInfo<'_, '_>,
    eclass: EcId,
    opfamily: u32,
    descending: bool,
    nulls_first: bool,
) -> PkId {
    assert!(root.ec_merging_done, "too soon to build canonical pathkeys");
    let pk = PathKey {
        pk_eclass: root.canonical_ec(eclass),
        pk_opfamily: opfamily,
        pk_descending: descending,
        pk_nulls_first: nulls_first,
    };
    if let Some(i) = root.canon_pathkeys.iter().position(|p| *p == pk) {
        return i;
    }
    root.canon_pathkeys.push(pk);
    root.canon_pathkeys.len() - 1
}

/// pathkey_is_redundant reports whether a pathkey adds nothing to an order, because its class holds a constant or
/// the order already sorts by its class, as Postgres' function of the same name does.
fn pathkey_is_redundant(root: &PlannerInfo<'_, '_>, new_pathkey: PkId, pathkeys: &[PkId]) -> bool {
    let new_ec = root.canon_pathkeys[new_pathkey].pk_eclass;
    root.eq_classes[new_ec].ec_has_const || pathkeys.iter().any(|&pk| root.canon_pathkeys[pk].pk_eclass == new_ec)
}

/// make_pathkey_from_sortinfo returns the canonical pathkey of an expression sorted by an operator family, as
/// Postgres' function of the same name does.
fn make_pathkey_from_sortinfo(
    root: &mut PlannerInfo<'_, '_>,
    expr: crate::expr::Expr,
    opfamily: u32,
    opcintype: u32,
    descending: bool,
    nulls_first: bool,
    sortref: usize,
) -> PkId {
    let eclass = get_eclass_for_sort_expr(root, expr, &[opfamily], opcintype, sortref, true)
        .expect("a class is created when none has the expression");
    make_canonical_pathkey(root, eclass, opfamily, descending, nulls_first)
}

/// compare_pathkeys compares two orders of rows, as Postgres' function of the same name does.
pub fn compare_pathkeys(keys1: &[PkId], keys2: &[PkId]) -> PathKeysComparison {
    let common = keys1.len().min(keys2.len());
    if keys1[..common] != keys2[..common] {
        return PathKeysComparison::Different;
    }
    match keys1.len().cmp(&keys2.len()) {
        std::cmp::Ordering::Equal => PathKeysComparison::Equal,
        std::cmp::Ordering::Greater => PathKeysComparison::Better1,
        std::cmp::Ordering::Less => PathKeysComparison::Better2,
    }
}

/// pathkeys_contained_in reports whether rows in the second order are also in the first, as Postgres' function of
/// the same name does.
pub fn pathkeys_contained_in(keys1: &[PkId], keys2: &[PkId]) -> bool {
    matches!(compare_pathkeys(keys1, keys2), PathKeysComparison::Equal | PathKeysComparison::Better2)
}

/// pathkeys_count_contained_in counts the keys that two orders share from the start, reporting whether the second
/// holds all of the first, as Postgres' function of the same name does.
pub fn pathkeys_count_contained_in(keys1: &[PkId], keys2: &[PkId]) -> (bool, usize) {
    let n = keys1.iter().zip(keys2).take_while(|(a, b)| a == b).count();
    (n == keys1.len(), n)
}

/// build_join_pathkeys returns the order of a join's rows that its outer path's order gives, as far as it is
/// useful, as Postgres' function of the same name does.
pub fn build_join_pathkeys(
    root: &PlannerInfo<'_, '_>,
    joinrel: usize,
    jointype: super::nodes::JoinType,
    outer_pathkeys: &[PkId],
) -> Vec<PkId> {
    match jointype {
        super::nodes::JoinType::Full | super::nodes::JoinType::Right => Vec::new(),
        _ => truncate_useless_pathkeys(root, joinrel, outer_pathkeys),
    }
}

/// make_pathkeys_for_sortclauses returns the pathkeys of sort keys over Vars, without redundant ones, as Postgres'
/// function of the same name does, or None when the planner does not know a key's btree operator family.
pub fn make_pathkeys_for_sortclauses(root: &mut PlannerInfo<'_, '_>, sortclauses: &[SortKey]) -> Option<Vec<PkId>> {
    let mut pathkeys = Vec::new();
    for (i, sortcl) in sortclauses.iter().enumerate() {
        let opcintype = super::nodefuncs::expr_type(root, &sortcl.expr)?;
        let opfamily = super::nodefuncs::btree_opfamily(opcintype)?;
        let pathkey = make_pathkey_from_sortinfo(
            root,
            sortcl.expr.clone(),
            opfamily,
            opcintype,
            sortcl.descending,
            sortcl.nulls_first,
            i + 1,
        );
        if !pathkey_is_redundant(root, pathkey, &pathkeys) {
            pathkeys.push(pathkey);
        }
    }
    Some(pathkeys)
}

/// initialize_mergeclause_eclasses gives each side of a mergejoinable clause that went into no equivalence class a
/// class of its own, as Postgres' function of the same name does.
pub fn initialize_mergeclause_eclasses(root: &mut PlannerInfo<'_, '_>, rinfo: RinfoId) {
    let r = root.rinfos[rinfo].clone();
    let Some((left, right)) = binary_op_args(&r.clause) else { return };
    let lefttype = super::nodefuncs::expr_type(root, left).unwrap_or(0);
    let righttype = super::nodefuncs::expr_type(root, right).unwrap_or(0);
    let left_ec = get_eclass_for_sort_expr(root, left.clone(), &r.mergeopfamilies, lefttype, 0, true);
    let right_ec = get_eclass_for_sort_expr(root, right.clone(), &r.mergeopfamilies, righttype, 0, true);
    let r = &mut root.rinfos[rinfo];
    (r.left_ec, r.right_ec) = (left_ec, right_ec);
}

/// pathkeys_useful_for_merging counts the leading pathkeys of an order that a merge join of a relation could use, as
/// Postgres' function of the same name does.
fn pathkeys_useful_for_merging(root: &PlannerInfo<'_, '_>, rel: usize, pathkeys: &[PkId]) -> usize {
    let mut useful = 0;
    for &pk in pathkeys {
        let pathkey = &root.canon_pathkeys[pk];
        if !right_merge_direction(root, pathkey) {
            break;
        }
        let matched = (root.rels[rel].has_eclass_joins && eclass_useful_for_merging(root, pathkey.pk_eclass, rel))
            || root.rels[rel].joininfo.iter().any(|&r| {
                let r = &root.rinfos[r];
                !r.mergeopfamilies.is_empty()
                    && (r.left_ec.map(|ec| root.canonical_ec(ec)) == Some(pathkey.pk_eclass)
                        || r.right_ec.map(|ec| root.canonical_ec(ec)) == Some(pathkey.pk_eclass))
            });
        if !matched {
            break;
        }
        useful += 1;
    }
    useful
}

/// right_merge_direction reports whether a pathkey sorts in the direction that the query's order asks for its class,
/// or ascending when it asks for none, as Postgres' function of the same name does.
fn right_merge_direction(root: &PlannerInfo<'_, '_>, pathkey: &PathKey) -> bool {
    for &qpk in &root.query_pathkeys {
        let query_pathkey = &root.canon_pathkeys[qpk];
        if pathkey.pk_eclass == query_pathkey.pk_eclass && pathkey.pk_opfamily == query_pathkey.pk_opfamily {
            return pathkey.pk_descending == query_pathkey.pk_descending;
        }
    }
    !pathkey.pk_descending
}

/// truncate_useless_pathkeys returns the leading pathkeys of an order that a merge join or the query's order could
/// use, as Postgres' function of the same name does.
pub fn truncate_useless_pathkeys(root: &PlannerInfo<'_, '_>, rel: usize, pathkeys: &[PkId]) -> Vec<PkId> {
    let nuseful = pathkeys_useful_for_merging(root, rel, pathkeys)
        .max(pathkeys_count_contained_in(&root.query_pathkeys, pathkeys).1);
    pathkeys[..nuseful].to_vec()
}
