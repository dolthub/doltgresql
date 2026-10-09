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

//! The parts of Postgres' optimizer/plan/planner.c that choose the final path.

use std::rc::Rc;

use super::nodes::{Path, RelOptInfo};
use super::pathnode::compare_fractional_path_costs;

/// get_cheapest_fractional_path returns the path of a relation that reads the rows a query reads most cheaply,
/// given as a count, or as a fraction below one, or zero for all of them, as Postgres' function of the same name
/// does.
pub fn get_cheapest_fractional_path(rel: &RelOptInfo, tuple_fraction: f64) -> Rc<Path> {
    let mut best_path = rel.cheapest_total_path.clone().expect("every relation has a path");
    if tuple_fraction <= 0.0 {
        return best_path;
    }
    let fraction = match tuple_fraction >= 1.0 && best_path.rows > 0.0 {
        true => tuple_fraction / best_path.rows,
        false => tuple_fraction,
    };
    for path in rel.pathlist.iter().filter(|p| p.param == 0) {
        if !Rc::ptr_eq(path, &best_path) && compare_fractional_path_costs(&best_path, path, fraction).is_gt() {
            best_path = path.clone();
        }
    }
    best_path
}
