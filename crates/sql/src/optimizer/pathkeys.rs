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

//! The parts of Postgres' optimizer/path/pathkeys.c that compare the orders of paths' rows.

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

/// compare_pathkeys compares two orders of rows, as Postgres' function of the same name does.
pub fn compare_pathkeys(keys1: &[SortKey], keys2: &[SortKey]) -> PathKeysComparison {
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
pub fn pathkeys_contained_in(keys1: &[SortKey], keys2: &[SortKey]) -> bool {
    matches!(compare_pathkeys(keys1, keys2), PathKeysComparison::Equal | PathKeysComparison::Better2)
}
