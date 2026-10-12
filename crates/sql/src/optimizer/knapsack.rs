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

//! Postgres' lib/knapsack.c: the discrete knapsack problem, which consider_groupingsets_paths solves to choose the
//! rollups that fit in a hash table's memory.

use std::collections::BTreeSet;

/// discrete_knapsack returns the items whose weights fit in a capacity with the greatest total value, where each item
/// is worth one when no values are given, as Postgres' DiscreteKnapsack chooses them.
pub fn discrete_knapsack(max_weight: usize, item_weights: &[usize], item_values: Option<&[f64]>) -> BTreeSet<usize> {
    let mut values = vec![0.0; max_weight + 1];
    let mut sets: Vec<BTreeSet<usize>> = vec![BTreeSet::new(); max_weight + 1];
    for (i, &iw) in item_weights.iter().enumerate() {
        let iv = item_values.map_or(1.0, |v| v[i]);
        for j in (iw..=max_weight).rev() {
            let ow = j - iw;
            if values[j] <= values[ow] + iv {
                if j != ow {
                    sets[j] = sets[ow].clone();
                }
                sets[j].insert(i);
                values[j] = values[ow] + iv;
            }
        }
    }
    std::mem::take(&mut sets[max_weight])
}
