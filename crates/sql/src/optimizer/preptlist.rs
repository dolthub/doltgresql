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

//! Postgres' optimizer/prep/preptlist.c: the target list that the plan of a query computes.

use super::PlannerInfo;
use super::nodes::CmdType;

/// preprocess_targetlist sets the target list that the plan of a query computes: the query's own, followed for an
/// UPDATE, DELETE, or MERGE by the columns that identify the result relation's rows, as Postgres' function of the same
/// name does. Doltgres' UPDATE, DELETE, and MERGE evaluate their assignments and RETURNING lists over each row that the
/// plan returns, so the query's own target list holds the columns of the other relations that they read.
pub fn preprocess_targetlist(root: &mut PlannerInfo<'_, '_>) {
    root.processed_tlist = root.parse.target_list.clone();
    if matches!(root.parse.command_type, CmdType::Update | CmdType::Delete | CmdType::Merge) {
        super::appendinfo::add_row_identity_columns(root, root.parse.result_relation);
    }
}
