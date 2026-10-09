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

use std::collections::{HashMap, HashSet};
use std::path::Path;

use doltdb::database::Database;
use doltdb::root::Root;
use doltdb::table::{ADAPTIVE_ENCODINGS, ADDRESS_ENCODINGS};
use prolly::{Node, Tuple};
use serial::{Message, WorkingSet};
use store::Hash;

use crate::catalog::table::TableDef;
use crate::error::{PgError, Result};

/// SENTINEL is the file in a database's .dolt directory that records that the database passed the startup integrity
/// check, so that later starts skip it.
pub const SENTINEL: &str = ".integrity_check_passed";

/// CHECK_VERSION is what a sentinel records, which changes whenever the check learns to find a new kind of
/// corruption.
const CHECK_VERSION: &str = "1";

/// Stats counts what a tree and the trees below it hold, and the chunk addresses their nodes fail to record, as
/// Doltgres' integrity.Stats does.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub struct Stats {
    pub chunks: u64,
    pub leaf_chunks: u64,
    /// The nodes that fail to record at least one address of their keys or values.
    pub corrupt_chunks: u64,
    pub rows: u64,
    /// The rows that hold an address that their node fails to record.
    pub corrupt_rows: u64,
    /// The adaptive values that are not NULL in value tuples.
    pub adaptive_values: u64,
    /// The adaptive values stored out of band in value tuples.
    pub out_of_band_values: u64,
    /// The addresses in value tuples that their nodes fail to record.
    pub corrupt_values: u64,
    /// The recorded value or key positions that hold no address.
    pub unexpected_offsets: u64,
    pub key_adaptive_values: u64,
    pub key_out_of_band_values: u64,
    pub key_corrupt_values: u64,
    /// The out-of-band adaptive values in the keys of internal nodes, which copy keys of the leaves.
    pub internal_key_out_of_band_values: u64,
    pub internal_key_corrupt_values: u64,
    /// The out-of-band values whose chunks the database no longer holds, which are lost.
    pub missing_chunks: u64,
}

impl Stats {
    /// add adds another tree's counts to these.
    pub fn add(&mut self, other: &Stats) {
        self.chunks += other.chunks;
        self.leaf_chunks += other.leaf_chunks;
        self.corrupt_chunks += other.corrupt_chunks;
        self.rows += other.rows;
        self.corrupt_rows += other.corrupt_rows;
        self.adaptive_values += other.adaptive_values;
        self.out_of_band_values += other.out_of_band_values;
        self.corrupt_values += other.corrupt_values;
        self.unexpected_offsets += other.unexpected_offsets;
        self.key_adaptive_values += other.key_adaptive_values;
        self.key_out_of_band_values += other.key_out_of_band_values;
        self.key_corrupt_values += other.key_corrupt_values;
        self.internal_key_out_of_band_values += other.internal_key_out_of_band_values;
        self.internal_key_corrupt_values += other.internal_key_corrupt_values;
        self.missing_chunks += other.missing_chunks;
    }

    /// corrupt returns the addresses that the nodes fail to record, in keys and values at every level.
    pub fn corrupt(&self) -> u64 {
        self.corrupt_values + self.key_corrupt_values + self.internal_key_corrupt_values
    }
}

/// CacheKey identifies a tree that was scanned: its address and the encodings of its keys and values.
pub type CacheKey = (Hash, Vec<u8>, Vec<u8>);

/// TableInfo is a table of a root value, with the encodings that its rows are read with.
pub struct TableInfo {
    pub schema: String,
    pub name: String,
    pub def: TableDef,
    pub keys: Vec<u8>,
    pub values: Vec<u8>,
    /// The value columns with an adaptive encoding.
    pub adaptive_value_columns: Vec<String>,
    /// The key columns with an adaptive encoding.
    pub adaptive_key_columns: Vec<String>,
}

impl TableInfo {
    /// values_impacted reports whether the table's value tuples hold adaptive values.
    pub fn values_impacted(&self) -> bool {
        !self.adaptive_value_columns.is_empty()
    }

    /// keys_impacted reports whether the table's key tuples hold adaptive values.
    pub fn keys_impacted(&self) -> bool {
        !self.adaptive_key_columns.is_empty()
    }

    /// shown returns the table's name, qualified by its schema when it has one.
    pub fn shown(&self) -> String {
        if self.schema.is_empty() { self.name.clone() } else { format!("{}.{}", self.schema, self.name) }
    }
}

/// tables_for_root returns every table of a root value.
pub fn tables_for_root(db: &mut Database, root: &Root) -> Result<Vec<TableInfo>> {
    let mut tables = Vec::new();
    for (key, address) in root.tables(db)? {
        let text = String::from_utf8_lossy(&key).into_owned();
        let mut parts = text.splitn(3, '\0').skip(1);
        let (Some(schema), Some(name)) = (parts.next(), parts.next()) else { continue };
        let def = TableDef::load(db, schema, name, address)?;
        let adaptive = |columns: &[usize]| -> Vec<String> {
            columns
                .iter()
                .map(|&i| &def.columns[i])
                .filter(|c| ADAPTIVE_ENCODINGS.contains(&c.encoding))
                .map(|c| c.name.clone())
                .collect()
        };
        let (adaptive_value_columns, adaptive_key_columns) = (adaptive(&def.value_columns), adaptive(&def.key_columns));
        tables.push(TableInfo {
            schema: schema.to_string(),
            name: name.to_string(),
            keys: def.key_encodings(),
            values: def.value_encodings(),
            def,
            adaptive_value_columns,
            adaptive_key_columns,
        });
    }
    Ok(tables)
}

/// Scanner checks trees, remembering what it found in each tree with the encodings it read the tree with, and which
/// chunks the database holds.
pub struct Scanner<'d> {
    pub db: &'d mut Database,
    cache: HashMap<CacheKey, Stats>,
    present: HashMap<Hash, bool>,
    /// The scans that the cache answered.
    pub cache_hits: u64,
}

impl<'d> Scanner<'d> {
    /// new returns a scanner of the database's trees.
    pub fn new(db: &'d mut Database) -> Scanner<'d> {
        Scanner { db, cache: HashMap::new(), present: HashMap::new(), cache_hits: 0 }
    }

    /// scan_table scans the rows of a table.
    pub fn scan_table(&mut self, table: &TableInfo) -> Result<Stats> {
        self.scan_root_node(&table.def.table.primary_index, &table.keys, &table.values)
    }

    /// scan_root_node scans a tree from its root node, which a table message holds rather than the database, so it
    /// need not be a chunk of its own.
    pub fn scan_root_node(&mut self, root: &[u8], keys: &[u8], values: &[u8]) -> Result<Stats> {
        let key = (Hash::of(root), keys.to_vec(), values.to_vec());
        if let Some(&stats) = self.cache.get(&key) {
            self.cache_hits += 1;
            return Ok(stats);
        }
        self.scan_node(key, Node::decode(root.to_vec())?)
    }

    /// scan_tree scans the tree whose root node is the chunk at the address.
    pub fn scan_tree(&mut self, address: Hash, keys: &[u8], values: &[u8]) -> Result<Stats> {
        let key = (address, keys.to_vec(), values.to_vec());
        if let Some(&stats) = self.cache.get(&key) {
            self.cache_hits += 1;
            return Ok(stats);
        }
        let node = tree_node(self.db, &address)?;
        self.scan_node(key, node)
    }

    /// scan_node scans a node and the nodes below it, caching what it found under the key.
    fn scan_node(&mut self, key: CacheKey, node: Node) -> Result<Stats> {
        let mut stats = Stats { chunks: 1, ..Stats::default() };
        if node.is_leaf() {
            let leaf = analyze_leaf(&node, &key.1, &key.2)?;
            stats.add(&leaf.stats);
            stats.leaf_chunks = 1;
            stats.corrupt_chunks = u64::from(leaf.corrupt);
            let addresses: Vec<Hash> = leaf.value_addresses.into_iter().chain(leaf.key_addresses).collect();
            stats.missing_chunks = self.count_missing(&addresses);
        } else {
            for i in 0..node.count() {
                let below = self.scan_tree(node.child(i)?, &key.1, &key.2)?;
                stats.add(&below);
            }
            let internal = analyze_internal_keys(&node, &key.1)?;
            stats.internal_key_out_of_band_values += internal.out_of_band_values;
            stats.internal_key_corrupt_values += internal.corrupt_values;
            stats.unexpected_offsets += internal.unexpected_offsets;
            stats.corrupt_chunks += u64::from(internal.corrupt);
        }
        self.cache.insert(key, stats);
        Ok(stats)
    }

    /// count_missing returns how many of the addresses name chunks that the database lacks.
    fn count_missing(&mut self, addresses: &[Hash]) -> u64 {
        let unknown: Vec<Hash> = addresses.iter().filter(|a| !self.present.contains_key(*a)).copied().collect();
        for (address, held) in unknown.iter().zip(self.db.has_many(&unknown)) {
            self.present.insert(*address, held);
        }
        addresses.iter().filter(|a| !self.present[*a]).count() as u64
    }
}

/// tree_node reads the tree node at the address, failing when the chunk is missing or holds something else.
pub fn tree_node(db: &Database, address: &Hash) -> Result<Node> {
    let data = db.read_value(address)?.ok_or_else(|| PgError::internal(format!("chunk {address} not found")))?;
    let node = Node::decode(data)?;
    if node.file_id() != serial::PROLLY_TREE_NODE {
        return Err(PgError::internal(format!(
            "chunk {address}: expected a {} message, found {}",
            serial::PROLLY_TREE_NODE,
            node.file_id()
        )));
    }
    Ok(node)
}

/// LeafAnalysis is what a leaf's tuples hold, compared with the addresses that the leaf records.
#[derive(Default)]
pub struct LeafAnalysis {
    pub stats: Stats,
    /// Whether the leaf fails to record an address of its keys or values.
    pub corrupt: bool,
    /// The addresses that value tuples hold.
    pub value_addresses: Vec<Hash>,
    /// The addresses of out-of-band adaptive values that key tuples hold.
    pub key_addresses: Vec<Hash>,
}

/// analyze_leaf finds the addresses that a leaf's tuples hold and checks that the leaf records each of them.
pub fn analyze_leaf(node: &Node, keys: &[u8], values: &[u8]) -> Result<LeafAnalysis> {
    let mut leaf = LeafAnalysis::default();
    let count = node.count();
    if count == 0 {
        return Ok(leaf);
    }
    let (recorded_keys, recorded_values) = node.address_offsets()?;
    let value_side = analyze_tuples(node, false, values, &recorded_values)?;
    let key_side = analyze_tuples(node, true, keys, &recorded_keys)?;
    leaf.stats.adaptive_values = value_side.adaptive_values;
    leaf.stats.out_of_band_values = value_side.out_of_band_values;
    leaf.stats.corrupt_values = value_side.corrupt_values;
    leaf.stats.key_adaptive_values = key_side.adaptive_values;
    leaf.stats.key_out_of_band_values = key_side.out_of_band_values;
    leaf.stats.key_corrupt_values = key_side.corrupt_values;
    leaf.stats.rows = count as u64;
    leaf.stats.corrupt_rows =
        value_side.corrupt_rows.iter().zip(&key_side.corrupt_rows).filter(|(v, k)| **v || **k).count() as u64;
    leaf.corrupt = leaf.stats.corrupt_rows > 0;
    leaf.stats.unexpected_offsets =
        (recorded_values.len() - value_side.matched + recorded_keys.len() - key_side.matched) as u64;
    leaf.value_addresses = value_side.addresses;
    leaf.key_addresses = key_side.addresses;
    Ok(leaf)
}

/// InternalKeyAnalysis is what an internal node's keys hold, compared with the addresses that the node records.
#[derive(Default)]
pub struct InternalKeyAnalysis {
    pub out_of_band_values: u64,
    pub corrupt_values: u64,
    pub unexpected_offsets: u64,
    /// Whether the node fails to record an address of its keys.
    pub corrupt: bool,
}

/// analyze_internal_keys finds the addresses that an internal node's keys hold, which copy keys of the leaves, and
/// checks that the node records each of them.
pub fn analyze_internal_keys(node: &Node, keys: &[u8]) -> Result<InternalKeyAnalysis> {
    if node.count() == 0 {
        return Ok(InternalKeyAnalysis::default());
    }
    let recorded = node.address_offsets()?.0;
    let side = analyze_tuples(node, true, keys, &recorded)?;
    Ok(InternalKeyAnalysis {
        out_of_band_values: side.out_of_band_values,
        corrupt_values: side.corrupt_values,
        unexpected_offsets: (recorded.len() - side.matched) as u64,
        corrupt: side.corrupt_values > 0,
    })
}

/// TupleAnalysis is what the keys or the values of a node hold, compared with the positions that the node records.
#[derive(Default)]
struct TupleAnalysis {
    adaptive_values: u64,
    out_of_band_values: u64,
    corrupt_values: u64,
    /// The recorded positions that an address matched.
    matched: usize,
    corrupt_rows: Vec<bool>,
    addresses: Vec<Hash>,
}

/// analyze_tuples finds where a node's keys or values hold addresses, and matches each against the recorded
/// positions, as Doltgres' analyzeTupleAddresses does.
fn analyze_tuples(node: &Node, keys: bool, encodings: &[u8], recorded: &[u16]) -> Result<TupleAnalysis> {
    let items = if keys { node.key_items() } else { node.value_items() };
    let mut remaining: HashMap<u16, usize> = HashMap::new();
    for &offset in recorded {
        *remaining.entry(offset).or_default() += 1;
    }
    let mut out = TupleAnalysis { corrupt_rows: vec![false; node.count()], ..TupleAnalysis::default() };
    for i in 0..node.count() {
        let item = if keys { node.key(i)? } else { node.value(i)? };
        let start = item.as_ptr() as usize - items.as_ptr() as usize;
        let tuple = Tuple(item);
        let mut expected = Vec::new();
        for (j, encoding) in encodings.iter().enumerate() {
            let Ok(Some((from, to))) = tuple.field_range(j) else { continue };
            let field = &item[from..to];
            if ADDRESS_ENCODINGS.contains(encoding) {
                if field.len() < Hash::LEN || field[..Hash::LEN].iter().all(|&b| b == 0) {
                    continue;
                }
                out.addresses.push(address(&field[..Hash::LEN]));
                expected.push(from);
            } else if ADAPTIVE_ENCODINGS.contains(encoding) && !field.is_empty() {
                out.adaptive_values += 1;
                if field[0] == 0 || field.len() < Hash::LEN {
                    continue;
                }
                out.out_of_band_values += 1;
                out.addresses.push(address(&field[field.len() - Hash::LEN..]));
                expected.push(to - Hash::LEN);
            }
        }
        for at in expected {
            match remaining.get_mut(&((start + at) as u16)) {
                Some(count) if *count > 0 => {
                    *count -= 1;
                    out.matched += 1;
                }
                _ => {
                    out.corrupt_values += 1;
                    out.corrupt_rows[i] = true;
                }
            }
        }
    }
    Ok(out)
}

/// address reads the address that a field holds.
fn address(bytes: &[u8]) -> Hash {
    let mut hash = [0; Hash::LEN];
    hash.copy_from_slice(bytes);
    Hash(hash)
}

/// check_data_dir checks every database of a data directory that has not yet passed the check, as the Go server's
/// startup does, recording a pass in each healthy database's sentinel, and returns the error of the first corrupt
/// database.
pub fn check_data_dir(data_dir: &Path) -> Result<Option<String>> {
    let mut databases: Vec<_> = std::fs::read_dir(data_dir)
        .map_err(PgError::internal)?
        .filter_map(|e| e.ok())
        .filter(|e| e.path().join(".dolt").is_dir())
        .collect();
    databases.sort_by_key(|e| e.file_name());
    for entry in databases {
        let dolt = entry.path().join(".dolt");
        let passed = std::fs::read_to_string(dolt.join(SENTINEL)).is_ok_and(|text| text.trim() == CHECK_VERSION);
        if passed {
            continue;
        }
        let name = entry.file_name().to_string_lossy().into_owned();
        let mut db = Database::open(&dolt.join("noms"))?;
        if let Some(corruption) = check_database(&mut db, &name)? {
            return Ok(Some(corruption.message()));
        }
        std::fs::write(dolt.join(SENTINEL), format!("{CHECK_VERSION}\n")).map_err(PgError::internal)?;
    }
    Ok(None)
}

/// Corruption is the first corrupt table that the check found, as Doltgres' CorruptionError describes it.
pub struct Corruption {
    pub database: String,
    pub branch: String,
    /// The corrupt commit, or None when the corruption is in the branch's working set.
    pub commit: Option<Hash>,
    pub table: String,
    pub stats: Stats,
}

impl Corruption {
    /// message returns the Go server's error for the corrupt table, with how to repair it.
    pub fn message(&self) -> String {
        let location = match self.commit {
            Some(commit) => format!("commit {commit} (reachable from branch {})", self.branch),
            None => format!("the working set of branch {}", self.branch),
        };
        format!(
            "database \"{}\" failed a startup integrity check: table {} at {location} has {} values (in {} rows) that \
             were serialized incorrectly due to an error in a previous release of Doltgres. To avoid data loss with \
             this version, the server will now exit.

To repair this database:
  1. MAKE A BACKUP COPY of the database directory before doing anything else.
  2. Build the repair tool from the DoltgreSQL repository at https://github.com/dolthub/doltgresql/ with this command:
\t\t\t\tgo build -o doltgres-admin ./cmd/admin
  3. Run: doltgres-admin report -dir <data-dir> to see the full extent of the corruption, then
     doltgres-admin repair -dir <data-dir> to repair it.

To start the server without this check (unsafe until the database is repaired), set \
             behavior.skip_startup_integrity_check to true in config.yaml.",
            self.database,
            self.table,
            self.stats.corrupt(),
            self.stats.corrupt_rows
        )
    }
}

/// check_database scans every table of every commit that a branch reaches, and of each branch's working set, for
/// values whose chunk addresses their nodes fail to record, as earlier releases wrote them, returning the first
/// corrupt table found.
pub fn check_database(db: &mut Database, name: &str) -> Result<Option<Corruption>> {
    let branches: Vec<(String, Hash)> = db
        .datasets()?
        .into_iter()
        .filter_map(|(r, head)| r.strip_prefix("refs/heads/").map(|branch| (branch.to_string(), head)))
        .collect();
    let mut scanner = Scanner::new(db);
    let mut visited = HashSet::new();
    for (branch, head) in branches {
        let corruption = |commit, (table, stats)| Corruption {
            database: name.to_string(),
            branch: branch.clone(),
            commit,
            table,
            stats,
        };
        let mut stack = vec![head];
        while let Some(commit) = stack.pop() {
            if !visited.insert(commit) {
                continue;
            }
            let Ok(info) = crate::dolt::history::load(scanner.db, commit) else { continue };
            if let Some(found) = check_root(&mut scanner, info.root)? {
                return Ok(Some(corruption(Some(commit), found)));
            }
            stack.extend(info.parents);
        }
        let Some(address) = scanner.db.head(&doltdb::create::working_set_ref(&branch))? else { continue };
        let data = crate::txn::read(scanner.db, &address)?;
        let working_set = WorkingSet::new(Message(&data))?;
        for root in [Some(working_set.working_root()?), working_set.staged_root()?].into_iter().flatten() {
            if let Some(found) = check_root(&mut scanner, root)? {
                return Ok(Some(corruption(None, found)));
            }
        }
    }
    Ok(None)
}

/// check_root scans the tables of a root value that hold adaptive values, returning the first corrupt one's name and
/// counts.
fn check_root(scanner: &mut Scanner<'_>, address: Hash) -> Result<Option<(String, Stats)>> {
    let root = Root::decode(&crate::txn::read(scanner.db, &address)?)?;
    for table in tables_for_root(scanner.db, &root)? {
        if !table.values_impacted() && !table.keys_impacted() {
            continue;
        }
        let stats = scanner.scan_table(&table)?;
        if stats.corrupt() > 0 {
            return Ok(Some((table.shown(), stats)));
        }
    }
    Ok(None)
}
