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

//! Ports of cmd/admin's tests, which start the doltgres binary that DOLTGRES_BIN_PATH names.

use std::collections::{HashMap, HashSet};
use std::path::{Path, PathBuf};
use std::sync::Arc;

use admin::repair::Repairer;
use admin::report::TableReport;
use admin::rewrite::{TreeRewriter, reserialize_internal, reserialize_leaf};
use admin::{Exit, scan_branch_heads};
use doltdb::database::Database;
use doltdb::root::Root;
use driver::client::Db;
use driver::process::{DoltUser, RepoStore};
use prolly::Node;
use serial::{Commit, Message};
use sql::integrity::{Scanner, Stats, TableInfo, analyze_leaf, check_database, tables_for_root, tree_node};
use sql::txn::read;
use store::Hash;

const TEST_DB_NAME: &str = "corruption_test";

/// Scans are the table reports of every branch head, by branch and table name.
type Scans = HashMap<String, HashMap<String, TableReport>>;

#[test]
fn test_report_and_repair_end_to_end() {
    let store = new_store();
    create_test_database(&store);

    let mut db = open_test_database(&store.dir);
    let baseline = scan_all_branches(&mut db);
    assert_table_stats(&baseline, "main", "t1", expect(1007, 1007, 7, 0, 0));
    assert_table_stats(&baseline, "main", "t4", expect(3, 3, 2, 0, 0));
    assert_table_stats(&baseline, "main", "t5", expect(2, 2, 2, 0, 0));
    assert_table_stats(&baseline, "b2", "t1", expect(1005, 1005, 5, 0, 0));
    assert_table_stats(&baseline, "b3", "t1", expect(1008, 1008, 8, 0, 0));
    assert_key_stats(&baseline, "main", "t3", 3, 3, 0);
    assert_not_scanned(&baseline, "main", "t2");

    assert!(map_height(&mut db, "main", "t6") > 1);
    assert_key_stats(&baseline, "main", "t6", 300, 300, 0);
    assert_ne!(stats(&baseline, "main", "t6").internal_key_out_of_band_values, 0);
    assert_eq!(stats(&baseline, "main", "t6").internal_key_corrupt_values, 0);

    let original_heads = branch_head_hashes(&mut db);
    assert_eq!(original_heads.len(), 3);
    let (original_addresses, _) = table_oob_addrs(&mut db, "main", "t1");
    assert_eq!(original_addresses.len(), 7);
    assert!(map_height(&mut db, "main", "t1") > 1);

    let summary = new_corrupter(&mut db).repair_database().unwrap();
    assert!(summary.commits_examined >= 5);
    assert_eq!(summary.commits_rewritten, 3, "only commits with out-of-band data are rewritten");
    assert_eq!(summary.branches_updated, 3);
    assert_eq!(summary.working_sets_fixed, 1);
    assert_ne!(summary.leaf_chunks_rewritten, 0);
    assert_ne!(summary.internal_chunks_rewritten, 0);
    db.close().unwrap();

    let mut db = open_test_database(&store.dir);
    let corrupted = scan_all_branches(&mut db);
    assert_table_stats(&corrupted, "main", "t1", expect(1007, 1007, 7, 7, 7));
    assert_table_stats(&corrupted, "main", "t4", expect(3, 3, 2, 2, 2));
    assert_table_stats(&corrupted, "main", "t5", expect(2, 2, 2, 2, 2));
    assert_table_stats(&corrupted, "b2", "t1", expect(1005, 1005, 5, 5, 5));
    assert_table_stats(&corrupted, "b3", "t1", expect(1008, 1008, 8, 8, 8));
    assert_key_stats(&corrupted, "main", "t3", 3, 3, 3);
    assert_key_stats(&corrupted, "main", "t6", 300, 300, 300);
    let t6 = stats(&corrupted, "main", "t6");
    assert_ne!(t6.internal_key_out_of_band_values, 0);
    assert_eq!(t6.internal_key_out_of_band_values, t6.internal_key_corrupt_values);

    let corrupt_heads = branch_head_hashes(&mut db);
    assert_ne!(original_heads["main"], corrupt_heads["main"]);
    assert_ne!(original_heads["b2"], corrupt_heads["b2"]);
    assert_eq!(root_commit_hash(&mut db, "main"), root_commit_hash(&mut db, "b2"));

    let walked = walked_addrs(&mut db, "main", "t1");
    for address in &original_addresses {
        assert!(!walked.contains(address), "corrupt tree should not reach out-of-band chunk {address}");
    }

    let corruption = check_database(&mut db, TEST_DB_NAME).unwrap().expect("the check refuses the corrupt database");
    assert_eq!(corruption.database, TEST_DB_NAME);
    assert_ne!(corruption.stats.corrupt_values, 0);
    assert!(corruption.message().contains("BACKUP"));
    db.close().unwrap();

    let report_path = store.dir.join("report.html");
    run_admin(&["repair", "-dir", &store.dir.display().to_string(), "-out", &report_path.display().to_string()]);
    let html = std::fs::read_to_string(&report_path).unwrap();
    assert!(html.contains("Repair summary"));
    assert!(html.contains("Post-repair verification scan"));

    let mut db = open_test_database(&store.dir);
    let repaired = scan_all_branches(&mut db);
    assert_table_stats(&repaired, "main", "t1", expect(1007, 1007, 7, 0, 0));
    assert_table_stats(&repaired, "main", "t4", expect(3, 3, 2, 0, 0));
    assert_table_stats(&repaired, "main", "t5", expect(2, 2, 2, 0, 0));
    assert_table_stats(&repaired, "b2", "t1", expect(1005, 1005, 5, 0, 0));
    assert_table_stats(&repaired, "b3", "t1", expect(1008, 1008, 8, 0, 0));
    assert_key_stats(&repaired, "main", "t3", 3, 3, 0);
    assert_key_stats(&repaired, "main", "t6", 300, 300, 0);
    assert_ne!(stats(&repaired, "main", "t6").internal_key_out_of_band_values, 0);
    assert_eq!(stats(&repaired, "main", "t6").internal_key_corrupt_values, 0);

    assert_eq!(original_heads, branch_head_hashes(&mut db));
    assert!(check_database(&mut db, TEST_DB_NAME).unwrap().is_none());
    let walked = walked_addrs(&mut db, "main", "t1");
    for address in &original_addresses {
        assert!(walked.contains(address), "repaired tree should reach out-of-band chunk {address}");
    }
    db.close().unwrap();
}

#[test]
fn test_repaired_database_is_servable() {
    let store = new_store();
    create_test_database(&store);

    let mut db = open_test_database(&store.dir);
    new_corrupter(&mut db).repair_database().unwrap();
    db.close().unwrap();

    let report_path = store.dir.join("report.html");
    run_admin(&["repair", "-dir", &store.dir.display().to_string(), "-out", &report_path.display().to_string()]);

    store
        .init_database(
            TEST_DB_NAME,
            Some(&|db: &mut Db| {
                check_count(db, "SELECT count(*) FROM t1 WHERE length(big) = 20000", "9")?;
                check_count(db, "SELECT count(*) FROM t3 WHERE length(big) = 20000", "3")?;
                check_count(db, "SELECT count(*) FROM t4 WHERE length(j::text) > 20000", "2")?;
                check_count(db, "SELECT count(*) FROM t1 AS OF 'b2' WHERE length(big) = 20000", "5")?;
                Ok(())
            }),
        )
        .unwrap();
}

#[test]
fn test_key_column_corruption_repair() {
    let store = new_store();
    create_test_database(&store);

    let mut db = open_test_database(&store.dir);
    let original_heads = branch_head_hashes(&mut db);
    let (_, original_key_addresses) = table_oob_addrs(&mut db, "main", "t3");
    assert_eq!(original_key_addresses.len(), 3);

    let mut corrupter = Repairer::new(Scanner::new(&mut db), false);
    corrupter.rewriter.transform_leaf = |_, node, _, values| {
        if node.address_offsets()?.0.is_empty() {
            return Ok(None);
        }
        Ok(Some(reserialize_leaf(node, &[], values)?))
    };
    corrupter.rewriter.transform_internal = corrupt_internal;
    corrupter.rewriter.should_rewrite =
        |stats| stats.key_out_of_band_values > 0 || stats.internal_key_out_of_band_values > 0;
    let summary = corrupter.repair_database().unwrap();
    assert_ne!(summary.leaf_chunks_rewritten, 0);
    db.close().unwrap();

    let mut db = open_test_database(&store.dir);
    let corrupted = scan_all_branches(&mut db);
    assert_table_stats(&corrupted, "main", "t1", expect(1007, 1007, 7, 0, 0));
    assert_key_stats(&corrupted, "main", "t3", 3, 3, 3);
    assert_key_stats(&corrupted, "main", "t6", 300, 300, 300);
    assert_ne!(stats(&corrupted, "main", "t6").internal_key_corrupt_values, 0);

    let corruption = check_database(&mut db, TEST_DB_NAME).unwrap().expect("the check refuses the corrupt database");
    assert_eq!(corruption.stats.corrupt_values, 0);
    assert_ne!(corruption.stats.key_corrupt_values, 0);

    let walked = walked_addrs(&mut db, "main", "t3");
    for address in &original_key_addresses {
        assert!(!walked.contains(address), "corrupt tree should not reach out-of-band key chunk {address}");
    }
    db.close().unwrap();

    let report_path = store.dir.join("report.html");
    run_admin(&["repair", "-dir", &store.dir.display().to_string(), "-out", &report_path.display().to_string()]);

    let mut db = open_test_database(&store.dir);
    let repaired = scan_all_branches(&mut db);
    assert_key_stats(&repaired, "main", "t3", 3, 3, 0);
    assert_key_stats(&repaired, "main", "t6", 300, 300, 0);
    assert_eq!(stats(&repaired, "main", "t6").internal_key_corrupt_values, 0);
    assert!(check_database(&mut db, TEST_DB_NAME).unwrap().is_none());
    assert_eq!(original_heads, branch_head_hashes(&mut db));

    let walked = walked_addrs(&mut db, "main", "t3");
    for address in &original_key_addresses {
        assert!(walked.contains(address), "repaired tree should reach out-of-band key chunk {address}");
    }
    db.close().unwrap();
}

#[test]
fn test_internal_only_key_bookkeeping_repair() {
    let store = new_store();
    create_test_database(&store);

    let mut db = open_test_database(&store.dir);
    let original_heads = branch_head_hashes(&mut db);

    let mut corrupter = Repairer::new(Scanner::new(&mut db), false);
    corrupter.rewriter.transform_leaf = |_, _, _, _| Ok(None);
    corrupter.rewriter.transform_internal = corrupt_internal;
    corrupter.rewriter.should_rewrite = |stats| stats.internal_key_out_of_band_values > 0;
    let summary = corrupter.repair_database().unwrap();
    assert_eq!(summary.leaf_chunks_rewritten, 0);
    assert_ne!(summary.internal_chunks_rewritten, 0);
    db.close().unwrap();

    let mut db = open_test_database(&store.dir);
    let corrupted = scan_all_branches(&mut db);
    let t6 = stats(&corrupted, "main", "t6");
    assert_eq!(t6.key_corrupt_values, 0, "leaf bookkeeping is intact");
    assert_ne!(t6.internal_key_corrupt_values, 0);
    assert!(check_database(&mut db, TEST_DB_NAME).unwrap().is_some());
    db.close().unwrap();

    let report_path = store.dir.join("report.html");
    run_admin(&["repair", "-dir", &store.dir.display().to_string(), "-out", &report_path.display().to_string()]);

    let mut db = open_test_database(&store.dir);
    let repaired = scan_all_branches(&mut db);
    assert_eq!(stats(&repaired, "main", "t6").internal_key_corrupt_values, 0);
    assert_ne!(stats(&repaired, "main", "t6").internal_key_out_of_band_values, 0);
    assert!(check_database(&mut db, TEST_DB_NAME).unwrap().is_none());
    assert_eq!(original_heads, branch_head_hashes(&mut db));
    db.close().unwrap();
}

#[test]
fn test_scan_after_garbage_collection() {
    let store = new_store();
    create_test_database(&store);
    store.init_database(TEST_DB_NAME, Some(&|db: &mut Db| db.exec("SELECT dolt_gc();", &[]))).unwrap();

    let mut db = open_test_database(&store.dir);
    let scans = scan_all_branches(&mut db);
    assert_table_stats(&scans, "main", "t1", expect(1007, 1007, 7, 0, 0));
    assert_key_stats(&scans, "main", "t6", 300, 300, 0);
    assert!(check_database(&mut db, TEST_DB_NAME).unwrap().is_none());
    db.close().unwrap();
}

#[test]
fn test_generate_corrupted_data_dir() {
    let Ok(dir) = std::env::var("ADMIN_TEST_GEN_DIR") else { return };
    let dir = PathBuf::from(dir);
    assert!(dir.is_absolute(), "ADMIN_TEST_GEN_DIR must be an absolute path, got {}", dir.display());
    std::fs::create_dir_all(&dir).unwrap();
    let store = RepoStore { user: Arc::new(DoltUser::new().unwrap()), dir: dir.clone() };
    create_test_database(&store);
    let mut db = open_test_database(&dir);
    new_corrupter(&mut db).repair_database().unwrap();
    db.close().unwrap();
}

/// new_store returns a new data directory, deleted with its user when the test ends.
fn new_store() -> RepoStore {
    Arc::new(DoltUser::new().unwrap()).make_repo_store().unwrap()
}

/// create_test_database creates a database with an initial commit holding only t2, which has no adaptive columns;
/// t1 with out-of-band and inline text values over several commits and branches, and uncommitted ones; t3 and t6 with
/// out-of-band text primary keys, t6 in a tree with internal nodes; t4 with jsonb; and t5 with an enum column.
fn create_test_database(store: &RepoStore) {
    const STATEMENTS: &[&str] = &[
        "CREATE TABLE t2 (id int primary key, n int)",
        "INSERT INTO t2 VALUES (1, 1), (2, 2)",
        "SELECT dolt_commit('-Am', 'commit 0: no adaptive data')",
        "CREATE TABLE t1 (id int primary key, big text)",
        "INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'x') FROM generate_series(1, 5) AS g(i)",
        "INSERT INTO t1 SELECT i, 'small-' || i::text FROM generate_series(101, 1100) AS g(i)",
        "CREATE TABLE t3 (big text primary key, n int)",
        "INSERT INTO t3 SELECT rpad(i::text, 20000, 'z'), i FROM generate_series(1, 3) AS g(i)",
        "CREATE TABLE t4 (id int primary key, j jsonb)",
        r#"INSERT INTO t4 SELECT i, ('{"k": "' || rpad(i::text, 20000, 'j') || '"}')::jsonb FROM generate_series(1, 2) AS g(i)"#,
        r#"INSERT INTO t4 VALUES (3, '{"k": "small"}')"#,
        "CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')",
        "CREATE TABLE t5 (id int primary key, m mood, big text)",
        "INSERT INTO t5 SELECT i, 'happy', rpad(i::text, 20000, 'e') FROM generate_series(1, 2) AS g(i)",
        "CREATE TABLE t6 (big text primary key, n int)",
        "INSERT INTO t6 SELECT lpad(i::text, 8, '0') || repeat('m', 19992), i FROM generate_series(1, 300) AS g(i)",
        "SELECT dolt_commit('-Am', 'commit 1: initial data')",
        "SELECT dolt_branch('b2')",
        "SELECT dolt_checkout('-b', 'b3')",
        "INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'v') FROM generate_series(11, 13) AS g(i)",
        "SELECT dolt_commit('-Am', 'commit on b3: divergent out-of-band values')",
        "SELECT dolt_checkout('main')",
        "INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'y') FROM generate_series(6, 7) AS g(i)",
        "SELECT dolt_commit('-Am', 'commit 2: more out-of-band values')",
        "INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'w') FROM generate_series(8, 9) AS g(i)",
    ];
    store
        .init_database(
            TEST_DB_NAME,
            Some(&|db: &mut Db| {
                for statement in STATEMENTS {
                    db.exec(statement, &[]).map_err(|e| format!("statement: {statement}: {e}"))?;
                }
                Ok(())
            }),
        )
        .unwrap();
}

/// check_count checks the count that a query returns.
fn check_count(db: &mut Db, query: &str, want: &str) -> Result<(), String> {
    let got = db.query(query, &[])?.rows[0][0].clone();
    if got != want {
        return Err(format!("{query} returned {got}, expected {want}"));
    }
    Ok(())
}

/// open_test_database opens the test database offline.
fn open_test_database(dir: &Path) -> Database {
    Database::open(&dir.join(TEST_DB_NAME).join(".dolt").join("noms")).unwrap()
}

/// run_admin runs the tool, failing the test when it fails.
fn run_admin(args: &[&str]) {
    let args: Vec<String> = args.iter().map(|a| a.to_string()).collect();
    match admin::run(&args) {
        Ok(()) => {}
        Err(Exit::Error(message)) => panic!("admin failed: {message}"),
        Err(Exit::Status(status)) => panic!("admin exited with {status}"),
    }
}

/// new_corrupter returns a repairer turned around to simulate old releases, which strips the recorded addresses from
/// every node that holds out-of-band values.
fn new_corrupter(db: &mut Database) -> Repairer<'_> {
    let mut corrupter = Repairer::new(Scanner::new(db), false);
    corrupter.rewriter.transform_leaf = corrupt_leaf;
    corrupter.rewriter.transform_internal = corrupt_internal;
    corrupter.rewriter.should_rewrite = |stats| {
        stats.out_of_band_values > 0 || stats.key_out_of_band_values > 0 || stats.internal_key_out_of_band_values > 0
    };
    corrupter
}

/// corrupt_leaf serializes a leaf again without recording the addresses of its keys and values, as old releases did.
fn corrupt_leaf(_: &mut TreeRewriter<'_>, node: &Node, _: &[u8], _: &[u8]) -> sql::Result<Option<Vec<u8>>> {
    let (keys, values) = node.address_offsets()?;
    if keys.is_empty() && values.is_empty() {
        return Ok(None);
    }
    Ok(Some(reserialize_leaf(node, &[], &[])?))
}

/// corrupt_internal serializes an internal node again without recording the addresses of its keys, as old releases
/// did at every level.
fn corrupt_internal(
    _: &mut TreeRewriter<'_>,
    node: &Node,
    children: &[Hash],
    changed: bool,
    _: &[u8],
    _: &[u8],
) -> sql::Result<Option<Vec<u8>>> {
    if !changed && node.address_offsets()?.0.is_empty() {
        return Ok(None);
    }
    Ok(Some(reserialize_internal(node, children, &[], &[])?))
}

/// scan_all_branches scans every branch head with a new scanner.
fn scan_all_branches(db: &mut Database) -> Scans {
    let mut scanner = Scanner::new(db);
    let mut scans = HashMap::new();
    for branch in scan_branch_heads(&mut scanner, false).unwrap() {
        let mut tables = HashMap::new();
        for table in branch.tables {
            assert!(table.error.is_empty(), "table {}.{} on branch {}", table.schema, table.table, branch.branch);
            tables.insert(table.table.clone(), table);
        }
        scans.insert(branch.branch, tables);
    }
    scans
}

/// Expect is what a table's scan should count.
struct Expect {
    rows: u64,
    adaptive: u64,
    out_of_band: u64,
    corrupt_values: u64,
    corrupt_rows: u64,
}

/// expect returns the counts a table's scan should find.
fn expect(rows: u64, adaptive: u64, out_of_band: u64, corrupt_values: u64, corrupt_rows: u64) -> Expect {
    Expect { rows, adaptive, out_of_band, corrupt_values, corrupt_rows }
}

/// stats returns the scan of a table on a branch.
fn stats(scans: &Scans, branch: &str, table: &str) -> Stats {
    let report = &scans[branch][table];
    report.stats.unwrap_or_else(|| panic!("table {table} on branch {branch} was not scanned"))
}

/// assert_table_stats checks the value-side counts of a table.
fn assert_table_stats(scans: &Scans, branch: &str, table: &str, expect: Expect) {
    let s = stats(scans, branch, table);
    assert_eq!(s.rows, expect.rows, "rows of {table} on {branch}");
    assert_eq!(s.adaptive_values, expect.adaptive, "adaptive values of {table} on {branch}");
    assert_eq!(s.out_of_band_values, expect.out_of_band, "out-of-band values of {table} on {branch}");
    assert_eq!(s.corrupt_values, expect.corrupt_values, "corrupt values of {table} on {branch}");
    assert_eq!(s.corrupt_rows, expect.corrupt_rows, "corrupt rows of {table} on {branch}");
    assert_eq!(s.unexpected_offsets, 0, "unexpected offsets of {table} on {branch}");
    assert_eq!(s.missing_chunks, 0, "missing chunks of {table} on {branch}");
    assert_eq!(s.corrupt_chunks > 0, expect.corrupt_values > 0, "corrupt chunks of {table} on {branch}");
}

/// assert_key_stats checks the key-side counts of a table whose primary key holds adaptive values.
fn assert_key_stats(scans: &Scans, branch: &str, table: &str, adaptive: u64, out_of_band: u64, corrupt: u64) {
    assert!(scans[branch][table].keys_impacted);
    let s = stats(scans, branch, table);
    assert_eq!(s.key_adaptive_values, adaptive, "key adaptive values of {table} on {branch}");
    assert_eq!(s.key_out_of_band_values, out_of_band, "key out-of-band values of {table} on {branch}");
    assert_eq!(s.key_corrupt_values, corrupt, "key corrupt values of {table} on {branch}");
    assert_eq!(s.corrupt_values, 0);
    assert_eq!(s.missing_chunks, 0, "missing chunks of {table} on {branch}");
}

/// assert_not_scanned checks that a table without adaptive values was not scanned.
fn assert_not_scanned(scans: &Scans, branch: &str, table: &str) {
    let report = &scans[branch][table];
    assert!(!report.values_impacted);
    assert!(!report.keys_impacted);
    assert!(report.stats.is_none());
}

/// branch_head_hashes returns the head commit of every branch.
fn branch_head_hashes(db: &mut Database) -> HashMap<String, Hash> {
    db.datasets()
        .unwrap()
        .into_iter()
        .filter_map(|(r, head)| r.strip_prefix("refs/heads/").map(|branch| (branch.to_string(), head)))
        .collect()
}

/// root_commit_hash returns the first commit, which has no parents, that a branch reaches.
fn root_commit_hash(db: &mut Database, branch: &str) -> Hash {
    let mut commit = branch_head_hashes(db)[branch];
    loop {
        let data = read(db, &commit).unwrap();
        match Commit::new(Message(&data)).unwrap().parents().unwrap().first() {
            Some(&parent) => commit = parent,
            None => return commit,
        }
    }
}

/// table_at_branch_head returns a table at the head of a branch.
fn table_at_branch_head(db: &mut Database, branch: &str, table: &str) -> TableInfo {
    let head = branch_head_hashes(db)[branch];
    let commit = sql::dolt::history::load(db, head).unwrap();
    let root = Root::decode(&read(db, &commit.root).unwrap()).unwrap();
    tables_for_root(db, &root).unwrap().into_iter().find(|t| t.name == table).expect("the table exists")
}

/// table_oob_addrs returns the addresses of out-of-band values that a table's value tuples and key tuples hold.
fn table_oob_addrs(db: &mut Database, branch: &str, table: &str) -> (HashSet<Hash>, HashSet<Hash>) {
    let info = table_at_branch_head(db, branch, table);
    let (mut values, mut keys) = (HashSet::new(), HashSet::new());
    let mut nodes = vec![Node::decode(info.def.table.primary_index.clone()).unwrap()];
    while let Some(node) = nodes.pop() {
        if node.is_leaf() {
            let leaf = analyze_leaf(&node, &info.keys, &info.values).unwrap();
            values.extend(leaf.value_addresses);
            keys.extend(leaf.key_addresses);
        } else {
            for i in 0..node.count() {
                nodes.push(tree_node(db, &node.child(i).unwrap()).unwrap());
            }
        }
    }
    (values, keys)
}

/// walked_addrs returns every address that the walk of a table's rows reaches, which is the walk that push, clone,
/// and garbage collection make.
fn walked_addrs(db: &mut Database, branch: &str, table: &str) -> HashSet<Hash> {
    let info = table_at_branch_head(db, branch, table);
    let mut addresses = HashSet::new();
    let mut nodes = vec![info.def.table.primary_index.clone()];
    while let Some(bytes) = nodes.pop() {
        let mut found = Vec::new();
        serial::walk::walk_addrs(Message(&bytes), &mut |address| {
            found.push(address);
            Ok(())
        })
        .unwrap();
        let node = Node::decode(bytes).unwrap();
        for address in found {
            addresses.insert(address);
            if !node.is_leaf() && (0..node.count()).any(|i| node.child(i).unwrap() == address) {
                nodes.push(read(db, &address).unwrap());
            }
        }
    }
    addresses
}

/// map_height returns the number of levels in a table's row map.
fn map_height(db: &mut Database, branch: &str, table: &str) -> usize {
    let info = table_at_branch_head(db, branch, table);
    Node::decode(info.def.table.primary_index.clone()).unwrap().level() as usize + 1
}
