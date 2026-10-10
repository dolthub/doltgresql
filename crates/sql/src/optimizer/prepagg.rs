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

//! Postgres' optimizer/prep/prepagg.c: what a query's aggregate calls cost. Doltgres' built-in aggregates are not
//! catalog entries with transition and final functions, so each call's transition function costs one operator, as a
//! built-in function does, and so does the final function of an aggregate whose Postgres counterpart has one.

use super::PlannerInfo;
use super::costsize::{CPU_OPERATOR_COST, QualCost, cost_qual_eval_node};
use crate::functions::aggregate::{AGGREGATES, AggCall};

/// AggClauseCosts are the costs of a query's aggregate calls, as Postgres' AggClauseCosts holds them: of their
/// transition functions and arguments for each row, of their final functions for each group, and the memory of their
/// transition states.
#[derive(Clone, Copy, Debug, Default)]
pub struct AggClauseCosts {
    pub trans_cost: QualCost,
    pub final_cost: QualCost,
    pub transition_space: f64,
}

/// get_agg_clause_costs returns the costs of a query's aggregate calls, as Postgres' function of the same name does.
pub fn get_agg_clause_costs(root: &PlannerInfo<'_, '_>) -> AggClauseCosts {
    let mut costs = AggClauseCosts::default();
    for call in &root.parse.aggregates {
        costs.trans_cost.per_tuple += CPU_OPERATOR_COST;
        for arg in call.args.iter().chain(&call.filter) {
            let argcosts = cost_qual_eval_node(arg);
            costs.trans_cost.startup += argcosts.startup;
            costs.trans_cost.per_tuple += argcosts.per_tuple;
        }
        if has_final_function(call) {
            costs.final_cost.per_tuple += CPU_OPERATOR_COST;
        }
    }
    costs
}

/// has_final_function reports whether the Postgres counterpart of an aggregate call has a final function: every
/// aggregate but those whose state is their result.
fn has_final_function(call: &AggCall) -> bool {
    if call.user.is_some() {
        return true;
    }
    let name = AGGREGATES[call.index].name;
    !matches!(name, "count" | "min" | "max" | "bool_and" | "bool_or" | "every" | "bit_and" | "bit_or" | "bit_xor")
}

/// count_ordered_aggs returns how many of a query's aggregate calls take ordered or DISTINCT input, as Postgres'
/// preprocess_aggrefs counts numOrderedAggs.
pub fn count_ordered_aggs(root: &PlannerInfo<'_, '_>) -> usize {
    root.parse.aggregates.iter().filter(|call| call.distinct || !call.order.is_empty()).count()
}
