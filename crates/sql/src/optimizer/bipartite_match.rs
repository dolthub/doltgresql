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

//! Postgres' lib/bipartite_match.c: the Hopcroft-Karp maximum cardinality matching of a bipartite graph, which
//! extract_rollup_sets runs to chain grouping sets into rollups. Vertices are numbered from 1, and 0 stands for none.

/// BipartiteMatchState is a maximum matching of a bipartite graph, as Postgres' BipartiteMatchState holds it: the
/// vertex of V that each vertex of U is matched to, and the vertex of U that each vertex of V is matched to.
pub struct BipartiteMatchState {
    pub pair_uv: Vec<usize>,
    pub pair_vu: Vec<usize>,
    distance: Vec<usize>,
}

/// HK_INFINITY is the distance of a vertex that the breadth-first search has not reached.
const HK_INFINITY: usize = usize::MAX;

/// bipartite_match returns a maximum matching of a bipartite graph of `u_size` and `v_size` vertices, whose edges
/// are the vertices of V adjacent to each vertex of U, as Postgres' BipartiteMatch finds it.
pub fn bipartite_match(u_size: usize, v_size: usize, adjacency: &[Vec<usize>]) -> BipartiteMatchState {
    let mut state = BipartiteMatchState {
        pair_uv: vec![0; u_size + 1],
        pair_vu: vec![0; v_size + 1],
        distance: vec![0; u_size + 1],
    };
    while hk_breadth_search(&mut state, adjacency) {
        for u in 1..=u_size {
            if state.pair_uv[u] == 0 {
                hk_depth_search(&mut state, adjacency, u);
            }
        }
    }
    state
}

/// hk_breadth_search sets the distance of each vertex of U from the unmatched ones along alternating paths, reporting
/// whether a path reaches an unmatched vertex of V, as Postgres' function of the same name does.
fn hk_breadth_search(state: &mut BipartiteMatchState, adjacency: &[Vec<usize>]) -> bool {
    let mut queue = Vec::with_capacity(state.pair_uv.len());
    state.distance[0] = HK_INFINITY;
    for u in 1..state.pair_uv.len() {
        match state.pair_uv[u] {
            0 => {
                state.distance[u] = 0;
                queue.push(u);
            }
            _ => state.distance[u] = HK_INFINITY,
        }
    }
    let mut qtail = 0;
    while qtail < queue.len() {
        let u = queue[qtail];
        qtail += 1;
        if state.distance[u] < state.distance[0] {
            for &v in adjacency[u].iter().rev() {
                let u_next = state.pair_vu[v];
                if state.distance[u_next] == HK_INFINITY {
                    state.distance[u_next] = 1 + state.distance[u];
                    queue.push(u_next);
                }
            }
        }
    }
    state.distance[0] != HK_INFINITY
}

/// hk_depth_search extends the matching along an alternating path from a vertex of U, reporting whether it found one,
/// as Postgres' function of the same name does.
fn hk_depth_search(state: &mut BipartiteMatchState, adjacency: &[Vec<usize>], u: usize) -> bool {
    if u == 0 {
        return true;
    }
    if state.distance[u] == HK_INFINITY {
        return false;
    }
    let nextdist = state.distance[u] + 1;
    for &v in adjacency[u].iter().rev() {
        if state.distance[state.pair_vu[v]] == nextdist && hk_depth_search(state, adjacency, state.pair_vu[v]) {
            state.pair_vu[v] = u;
            state.pair_uv[u] = v;
            return true;
        }
    }
    state.distance[u] = HK_INFINITY;
    false
}
