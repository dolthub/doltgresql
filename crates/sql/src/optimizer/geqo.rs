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

//! Postgres' optimizer/geqo: the genetic query optimizer, which searches the orders of joining many relations with
//! a genetic algorithm instead of trying every one. A tour is an order of the initial relations, numbered from 1; a
//! pool of tours evolves by edge recombination crossover (ERX, the one Postgres builds), keeping the cheapest. Its
//! random numbers come from Postgres' pg_prng, so a seed gives the same plans that Postgres chooses.

use super::PlannerInfo;
use super::joinrels::{have_join_order_restriction, make_join_rel};
use super::pathnode::set_cheapest;

/// Gene is a relation's number in a tour, as Postgres' Gene is.
type Gene = i32;

/// Chromosome is a tour with the cost of its plan, as Postgres' Chromosome is.
#[derive(Clone, Debug)]
struct Chromosome {
    string: Vec<Gene>,
    worth: f64,
}

/// Pool is the tours of a generation, cheapest first, as Postgres' Pool is.
struct Pool {
    data: Vec<Chromosome>,
    string_length: usize,
}

/// Edge is a relation's neighbors in the two tours being recombined, as Postgres' Edge is: a neighbor that both
/// tours share is negated.
#[derive(Clone, Copy, Debug, Default)]
struct Edge {
    edge_list: [Gene; 4],
    total_edges: i32,
    unused_edges: i32,
}

/// GeqoPrivateData is the state of one search, as Postgres' GeqoPrivateData is: the relations to join and the
/// random number generator.
struct GeqoPrivateData {
    initial_rels: Vec<usize>,
    random_state: PgPrng,
}

/// GeqoSettings are the geqo_* settings that the search reads.
struct GeqoSettings {
    effort: i32,
    pool_size: i32,
    generations: i32,
    selection_bias: f64,
    seed: f64,
}

/// geqo returns the relation that joins the initial relations, built in the order of the cheapest tour that the
/// genetic search finds, as Postgres' function of the same name does.
pub fn geqo(root: &mut PlannerInfo<'_, '_>, number_of_rels: usize, initial_rels: Vec<usize>) -> usize {
    let setting = |name: &str| root.ctx.session.settings.get(name).unwrap_or_default();
    let settings = GeqoSettings {
        effort: setting("geqo_effort").parse().unwrap_or(5),
        pool_size: setting("geqo_pool_size").parse().unwrap_or(0),
        generations: setting("geqo_generations").parse().unwrap_or(0),
        selection_bias: setting("geqo_selection_bias").parse().unwrap_or(2.0),
        seed: setting("geqo_seed").parse().unwrap_or(0.0),
    };
    let mut private = GeqoPrivateData { initial_rels, random_state: PgPrng::default() };
    geqo_set_seed(&mut private, settings.seed);
    let pool_size = gimme_pool_size(&settings, number_of_rels);
    let number_generations = gimme_number_generations(&settings, pool_size);
    let mut pool = alloc_pool(pool_size, number_of_rels);
    random_init_pool(root, &mut private, &mut pool);
    sort_pool(&mut pool);
    let mut momma = alloc_chromo(pool.string_length);
    let mut daddy = alloc_chromo(pool.string_length);
    let mut edge_table = alloc_edge_table(pool.string_length);
    for _ in 0..number_generations {
        geqo_selection(&mut private, &mut momma, &mut daddy, &pool, settings.selection_bias);
        gimme_edge_table(&momma.string, &daddy.string, pool.string_length, &mut edge_table);
        let kid = &mut momma;
        gimme_tour(&mut private, &mut edge_table, &mut kid.string, pool.string_length);
        kid.worth = geqo_eval(root, &private, &kid.string);
        spread_chromo(kid, &mut pool);
    }
    let best_tour = pool.data[0].string.clone();
    gimme_tree(root, &private, &best_tour).expect("geqo failed to make a valid plan")
}

/// gimme_pool_size returns the number of tours in a pool for a number of relations, as Postgres' function of the
/// same name does: geqo_pool_size when it is set, and otherwise two to the power of one more than the relations,
/// between ten and fifty times geqo_effort.
fn gimme_pool_size(settings: &GeqoSettings, nr_rel: usize) -> usize {
    if settings.pool_size >= 2 {
        return settings.pool_size as usize;
    }
    let size = 2f64.powf(nr_rel as f64 + 1.0);
    let maxsize = 50 * settings.effort;
    if size > maxsize as f64 {
        return maxsize as usize;
    }
    let minsize = 10 * settings.effort;
    if size < minsize as f64 {
        return minsize as usize;
    }
    size.ceil() as usize
}

/// gimme_number_generations returns the number of generations to evolve, geqo_generations when it is set and
/// otherwise the pool's size, as Postgres' function of the same name does.
fn gimme_number_generations(settings: &GeqoSettings, pool_size: usize) -> usize {
    match settings.generations > 0 {
        true => settings.generations as usize,
        false => pool_size,
    }
}

/// geqo_eval returns the cost of the cheapest path of joining the relations in a tour's order, or the largest cost
/// when the order joins them in no valid way, as Postgres' function of the same name does. The join relations it
/// builds are discarded afterward.
fn geqo_eval(root: &mut PlannerInfo<'_, '_>, private: &GeqoPrivateData, tour: &[Gene]) -> f64 {
    let savelength = root.rels.len();
    let savehash = root.join_rel_hash.clone();
    let fitness = match gimme_tree(root, private, tour) {
        Some(joinrel) => root.rels[joinrel].cheapest_total_path.as_ref().map_or(f64::MAX, |path| path.total_cost),
        None => f64::MAX,
    };
    root.rels.truncate(savelength);
    root.join_rel_hash = savehash;
    fitness
}

/// Clump is a join relation of some of a tour's relations and how many it joins, as Postgres' Clump is.
struct Clump {
    joinrel: usize,
    size: usize,
}

/// gimme_tree returns the relation that joins a tour's relations, joining each one in turn to the first clump it
/// has a join clause or a join order restriction with, and then the clumps that remain in any way, as Postgres'
/// function of the same name does, or None when they cannot all be joined.
fn gimme_tree(root: &mut PlannerInfo<'_, '_>, private: &GeqoPrivateData, tour: &[Gene]) -> Option<usize> {
    let mut clumps = Vec::new();
    for &gene in tour {
        let cur_rel = private.initial_rels[gene as usize - 1];
        clumps = merge_clump(root, clumps, Clump { joinrel: cur_rel, size: 1 }, false);
    }
    if clumps.len() > 1 {
        let mut fclumps = Vec::new();
        for clump in clumps {
            fclumps = merge_clump(root, fclumps, clump, true);
        }
        clumps = fclumps;
    }
    match clumps.as_slice() {
        [clump] => Some(clump.joinrel),
        _ => None,
    }
}

/// merge_clump joins a new clump to the first clump of a list that it is desirable to join with, or to any when
/// forced, merging the result into the list again, and otherwise adds it to the list among clumps of its size, as
/// Postgres' function of the same name does.
fn merge_clump(root: &mut PlannerInfo<'_, '_>, mut clumps: Vec<Clump>, new_clump: Clump, force: bool) -> Vec<Clump> {
    for i in 0..clumps.len() {
        if (force || desirable_join(root, clumps[i].joinrel, new_clump.joinrel))
            && let Some(joinrel) = make_join_rel(root, clumps[i].joinrel, new_clump.joinrel)
        {
            set_cheapest(&mut root.rels[joinrel]);
            let mut old_clump = clumps.remove(i);
            old_clump.joinrel = joinrel;
            old_clump.size += new_clump.size;
            return merge_clump(root, clumps, old_clump, force);
        }
    }
    if clumps.is_empty() || new_clump.size == 1 {
        clumps.push(new_clump);
        return clumps;
    }
    let pos = clumps.iter().position(|old_clump| new_clump.size > old_clump.size).unwrap_or(clumps.len());
    clumps.insert(pos, new_clump);
    clumps
}

/// desirable_join reports whether two relations have a join clause between them or must be joined to each other
/// first, as Postgres' function of the same name does.
fn desirable_join(root: &PlannerInfo<'_, '_>, outer_rel: usize, inner_rel: usize) -> bool {
    super::joininfo::have_relevant_joinclause(root, outer_rel, inner_rel)
        || have_join_order_restriction(root, outer_rel, inner_rel)
}

/// alloc_pool returns a pool of tours of a length, as Postgres' function of the same name does.
fn alloc_pool(pool_size: usize, string_length: usize) -> Pool {
    Pool { data: vec![alloc_chromo(string_length); pool_size], string_length }
}

/// random_init_pool fills a pool with random tours that join the relations in some valid way, as Postgres' function
/// of the same name does.
fn random_init_pool(root: &mut PlannerInfo<'_, '_>, private: &mut GeqoPrivateData, pool: &mut Pool) {
    let mut bad = 0;
    let mut i = 0;
    while i < pool.data.len() {
        init_tour(private, &mut pool.data[i].string);
        pool.data[i].worth = geqo_eval(root, private, &pool.data[i].string);
        if pool.data[i].worth < f64::MAX {
            i += 1;
        } else {
            bad += 1;
            assert!(i != 0 || bad < 10000, "geqo failed to make a valid plan");
        }
    }
}

/// sort_pool sorts a pool's tours from the cheapest, as Postgres' function of the same name does, though its qsort may
/// order tours of equal cost differently.
fn sort_pool(pool: &mut Pool) {
    pool.data.sort_by(|a, b| a.worth.total_cmp(&b.worth));
}

/// alloc_chromo returns a tour of a length, as Postgres' function of the same name does.
fn alloc_chromo(string_length: usize) -> Chromosome {
    Chromosome { string: vec![0; string_length], worth: 0.0 }
}

/// spread_chromo puts a new tour into its place in a pool by cost, dropping the most expensive one, unless the new
/// one costs more than all of them, as Postgres' function of the same name does.
fn spread_chromo(chromo: &Chromosome, pool: &mut Pool) {
    let size = pool.data.len();
    if chromo.worth > pool.data[size - 1].worth {
        return;
    }
    let (mut top, mut mid, mut bot) = (0, size / 2, size - 1);
    let index = loop {
        if chromo.worth <= pool.data[top].worth {
            break top;
        } else if chromo.worth == pool.data[mid].worth {
            break mid;
        } else if chromo.worth == pool.data[bot].worth || bot - top <= 1 {
            break bot;
        } else if chromo.worth < pool.data[mid].worth {
            bot = mid;
            mid = top + (bot - top) / 2;
        } else {
            top = mid;
            mid = top + (bot - top) / 2;
        }
    };
    pool.data.pop();
    pool.data.insert(index, chromo.clone());
}

/// geqo_selection copies two different tours of a pool, chosen at random with a bias toward the cheaper ones, as
/// Postgres' function of the same name does.
fn geqo_selection(
    private: &mut GeqoPrivateData,
    momma: &mut Chromosome,
    daddy: &mut Chromosome,
    pool: &Pool,
    bias: f64,
) {
    let size = pool.data.len();
    let first = linear_rand(private, size, bias);
    let mut second = linear_rand(private, size, bias);
    if size > 1 {
        while first == second {
            second = linear_rand(private, size, bias);
        }
    }
    *momma = pool.data[first].clone();
    *daddy = pool.data[second].clone();
}

/// linear_rand returns a random position in a pool of a size, more likely the nearer it is to the start by a bias,
/// as Postgres' function of the same name does.
fn linear_rand(private: &mut GeqoPrivateData, pool_size: usize, bias: f64) -> usize {
    let max = pool_size as f64;
    loop {
        let mut sqrtval = bias * bias - 4.0 * (bias - 1.0) * geqo_rand(private);
        if sqrtval > 0.0 {
            sqrtval = sqrtval.sqrt();
        }
        let index = max * (bias - sqrtval) / 2.0 / (bias - 1.0);
        if (0.0..max).contains(&index) {
            return index as usize;
        }
    }
}

/// init_tour fills a tour with a random order of its relations, as Postgres' function of the same name does.
fn init_tour(private: &mut GeqoPrivateData, tour: &mut [Gene]) {
    if !tour.is_empty() {
        tour[0] = 1;
    }
    for i in 1..tour.len() {
        let j = geqo_randint(private, i as u64, 0) as usize;
        if i != j {
            tour[i] = tour[j];
        }
        tour[j] = i as Gene + 1;
    }
}

/// alloc_edge_table returns the edge table of tours of a length, numbered from 1, as Postgres' function of the same
/// name does.
fn alloc_edge_table(num_gene: usize) -> Vec<Edge> {
    vec![Edge::default(); num_gene + 1]
}

/// gimme_edge_table fills the edge table with the neighbors of each relation in two tours, which are cycles, as
/// Postgres' function of the same name does.
fn gimme_edge_table(tour1: &[Gene], tour2: &[Gene], num_gene: usize, edge_table: &mut [Edge]) {
    for edge in &mut edge_table[1..=num_gene] {
        edge.total_edges = 0;
        edge.unused_edges = 0;
    }
    for index1 in 0..num_gene {
        let index2 = (index1 + 1) % num_gene;
        gimme_edge(tour1[index1], tour1[index2], edge_table);
        gimme_edge(tour1[index2], tour1[index1], edge_table);
        gimme_edge(tour2[index1], tour2[index2], edge_table);
        gimme_edge(tour2[index2], tour2[index1], edge_table);
    }
}

/// gimme_edge adds a neighbor to a relation's edges, marking it shared when it is there already, and reports whether
/// it was new, as Postgres' function of the same name does.
fn gimme_edge(gene1: Gene, gene2: Gene, edge_table: &mut [Edge]) -> bool {
    let city1 = &mut edge_table[gene1 as usize];
    let edges = city1.total_edges as usize;
    if let Some(edge) = city1.edge_list[..edges].iter_mut().find(|edge| edge.abs() == gene2) {
        *edge = -gene2;
        return false;
    }
    city1.edge_list[edges] = gene2;
    city1.total_edges += 1;
    city1.unused_edges += 1;
    true
}

/// gimme_tour fills a new tour from the edge table, starting at a random relation and following each one's
/// neighbors, and returns how many times it ran out of them, as Postgres' function of the same name does.
fn gimme_tour(private: &mut GeqoPrivateData, edge_table: &mut [Edge], new_gene: &mut [Gene], num_gene: usize) -> usize {
    let mut edge_failures = 0;
    new_gene[0] = geqo_randint(private, num_gene as u64, 1) as Gene;
    for i in 1..num_gene {
        let previous = new_gene[i - 1] as usize;
        remove_gene(new_gene[i - 1], edge_table[previous], edge_table);
        if edge_table[previous].unused_edges > 0 {
            new_gene[i] = gimme_gene(private, edge_table[previous], edge_table);
        } else {
            edge_failures += 1;
            new_gene[i] = edge_failure(private, new_gene, i - 1, edge_table, num_gene);
        }
        edge_table[previous].unused_edges = -1;
    }
    edge_failures
}

/// remove_gene removes a relation from the unused edges of each of its neighbors, as Postgres' function of the same
/// name does.
fn remove_gene(gene: Gene, edge: Edge, edge_table: &mut [Edge]) {
    for &possess_edge in &edge.edge_list[..edge.unused_edges.max(0) as usize] {
        let neighbor = &mut edge_table[possess_edge.unsigned_abs() as usize];
        let genes_remaining = neighbor.unused_edges.max(0) as usize;
        if let Some(j) = neighbor.edge_list[..genes_remaining].iter().position(|e| e.abs() == gene) {
            neighbor.unused_edges -= 1;
            neighbor.edge_list[j] = neighbor.edge_list[genes_remaining - 1];
        }
    }
}

/// gimme_gene returns the next relation of a tour from a relation's neighbors: one that both tours share, or else
/// one chosen at random among those with the fewest neighbors left, as Postgres' function of the same name does.
fn gimme_gene(private: &mut GeqoPrivateData, edge: Edge, edge_table: &[Edge]) -> Gene {
    let mut minimum_edges = 5;
    let mut minimum_count: i64 = -1;
    let unused = &edge.edge_list[..edge.unused_edges as usize];
    for &friend in unused {
        if friend < 0 {
            return friend.abs();
        }
        let friend_edges = edge_table[friend as usize].unused_edges;
        if friend_edges < minimum_edges {
            minimum_edges = friend_edges;
            minimum_count = 1;
        } else {
            assert!(minimum_count != -1, "minimum_count not set");
            if friend_edges == minimum_edges {
                minimum_count += 1;
            }
        }
    }
    let rand_decision = geqo_randint(private, (minimum_count - 1) as u64, 0) as i64;
    for &friend in unused {
        if edge_table[friend as usize].unused_edges == minimum_edges {
            minimum_count -= 1;
            if minimum_count == rand_decision {
                return friend;
            }
        }
    }
    panic!("neither shared nor minimum number nor random edge found");
}

/// edge_failure returns the next relation of a tour when the last one has no neighbors left: one chosen at random
/// among the unused relations with four edges, or else among all unused ones, as Postgres' function of the same name
/// does.
fn edge_failure(
    private: &mut GeqoPrivateData,
    gene: &[Gene],
    index: usize,
    edge_table: &[Edge],
    num_gene: usize,
) -> Gene {
    let fail_gene = gene[index];
    let candidate = |i: usize| edge_table[i].unused_edges != -1 && i as Gene != fail_gene;
    let remaining_edges = (1..=num_gene).filter(|&i| candidate(i)).count() as u64;
    let mut four_count = (1..=num_gene).filter(|&i| candidate(i) && edge_table[i].total_edges == 4).count() as u64;
    if four_count != 0 {
        let rand_decision = geqo_randint(private, four_count - 1, 0);
        for (i, edge) in edge_table.iter().enumerate().take(num_gene + 1).skip(1) {
            if candidate(i) && edge.total_edges == 4 {
                four_count -= 1;
                if rand_decision == four_count {
                    return i as Gene;
                }
            }
        }
    } else if remaining_edges != 0 {
        let mut remaining_edges = remaining_edges;
        let rand_decision = geqo_randint(private, remaining_edges - 1, 0);
        for i in 1..=num_gene {
            if candidate(i) {
                remaining_edges -= 1;
                if rand_decision == remaining_edges {
                    return i as Gene;
                }
            }
        }
    } else if let Some(i) = (1..=num_gene).find(|&i| edge_table[i].unused_edges >= 0) {
        return i as Gene;
    }
    panic!("no edge found");
}

/// geqo_set_seed seeds a search's random numbers, as Postgres' function of the same name does.
fn geqo_set_seed(private: &mut GeqoPrivateData, seed: f64) {
    private.random_state.fseed(seed);
}

/// geqo_rand returns a random number in [0, 1), as Postgres' function of the same name does.
fn geqo_rand(private: &mut GeqoPrivateData) -> f64 {
    private.random_state.double()
}

/// geqo_randint returns a random integer between two bounds, inclusive, as Postgres' function of the same name does.
fn geqo_randint(private: &mut GeqoPrivateData, upper: u64, lower: u64) -> u64 {
    private.random_state.uint64_range(lower, upper)
}

/// PgPrng is Postgres' pg_prng random number generator, xoroshiro128**, seeded by splitmix64.
#[derive(Clone, Copy, Debug, Default)]
struct PgPrng {
    s0: u64,
    s1: u64,
}

impl PgPrng {
    /// fseed seeds the generator from a number in [-1, 1], as pg_prng_fseed does.
    fn fseed(&mut self, fseed: f64) {
        let seed = (((1u64 << 52) - 1) as f64 * fseed) as i64;
        self.seed(seed as u64);
    }

    /// seed seeds the generator, as pg_prng_seed does, with a fixed state for one that would be all zeroes.
    fn seed(&mut self, mut seed: u64) {
        self.s0 = splitmix64(&mut seed);
        self.s1 = splitmix64(&mut seed);
        if self.s0 == 0 && self.s1 == 0 {
            (self.s0, self.s1) = (0x5851F42D4C957F2D, 0x14057B7EF767814F);
        }
    }

    /// next returns the next random 64 bits, as xoroshiro128ss does.
    fn next(&mut self) -> u64 {
        let s0 = self.s0;
        let sx = self.s1 ^ s0;
        let val = s0.wrapping_mul(5).rotate_left(7).wrapping_mul(9);
        self.s0 = s0.rotate_left(24) ^ sx ^ (sx << 16);
        self.s1 = sx.rotate_left(37);
        val
    }

    /// uint64_range returns a random integer between two bounds, inclusive, as pg_prng_uint64_range does.
    fn uint64_range(&mut self, rmin: u64, rmax: u64) -> u64 {
        if rmax <= rmin {
            return rmin;
        }
        let range = rmax - rmin;
        let rshift = range.leading_zeros();
        loop {
            let val = self.next() >> rshift;
            if val <= range {
                return rmin + val;
            }
        }
    }

    /// double returns a random number in [0, 1), as pg_prng_double does.
    fn double(&mut self) -> f64 {
        (self.next() >> (64 - 52)) as f64 * 2f64.powi(-52)
    }
}

/// splitmix64 returns the next number of a splitmix64 sequence, as pg_prng's function of the same name does.
fn splitmix64(state: &mut u64) -> u64 {
    *state = state.wrapping_add(0x9E3779B97F4A7C15);
    let mut val = *state;
    val = (val ^ (val >> 30)).wrapping_mul(0xBF58476D1CE4E5B9);
    val = (val ^ (val >> 27)).wrapping_mul(0x94D049BB133111EB);
    val ^ (val >> 31)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn random_numbers_match_postgres() {
        let mut prng = PgPrng::default();
        prng.fseed(0.5);
        assert_eq!(prng.double(), 0.9851677175347999);
        assert_eq!(prng.double(), 0.825301858027981);
        prng.fseed(0.0);
        assert_eq!(prng.double(), 0.8702553105818676);
    }
}
