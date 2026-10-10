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

use std::path::Path;

use driver::runner::{Outcome, run_tests_file};

/// run_file runs a test definition file, printing every test's outcome and failing when any test fails.
fn run_file(name: &str) {
    let path =
        Path::new(env!("CARGO_MANIFEST_DIR")).join("../../integration-tests/go-sql-server-driver/tests").join(name);
    let outcomes = run_tests_file(&path).unwrap();
    let mut failures = Vec::new();
    for (test, outcome) in outcomes {
        match outcome {
            Outcome::Pass => println!("--- PASS: {test}"),
            Outcome::Skip(reason) => println!("--- SKIP: {test}: {reason}"),
            Outcome::Fail(problems) => {
                println!("--- FAIL: {test}");
                failures.push(format!("{test}:\n  {}", problems.join("\n  ")));
            }
        }
    }
    assert!(failures.is_empty(), "{}", failures.join("\n"));
}

#[test]
fn test_config() {
    run_file("sql-server-config.yaml");
}

#[test]
fn test_cluster() {
    run_file("sql-server-cluster.yaml");
}

#[test]
fn test_cluster_root_objects() {
    run_file("sql-server-cluster-root-objects.yaml");
}

#[test]
fn test_cluster_auth_replication() {
    run_file("sql-server-cluster-auth-replication.yaml");
}

#[test]
fn test_remotes_api() {
    run_file("sql-server-remotesapi.yaml");
}

#[test]
fn test_cluster_tls() {
    run_file("sql-server-cluster-tls.yaml");
}

#[test]
fn test_original() {
    run_file("sql-server-orig.yaml");
}

#[test]
fn test_tls() {
    run_file("sql-server-tls.yaml");
}

#[test]
fn test_cluster_read_only() {
    run_file("sql-server-cluster-read-only.yaml");
}

#[test]
fn test_large_text_replication() {
    run_file("sql-server-large-text-replication.yaml");
}

#[test]
fn test_large_blob_replication() {
    run_file("sql-server-large-blob-replication.yaml");
}

#[test]
fn test_large_json_replication() {
    run_file("sql-server-large-json-replication.yaml");
}

#[test]
fn test_large_multi_column_replication() {
    run_file("sql-server-large-multi-column-replication.yaml");
}

#[test]
fn test_large_values_gc() {
    run_file("sql-server-large-values-gc.yaml");
}

#[test]
fn test_type_diversity_cluster() {
    run_file("sql-server-type-diversity.yaml");
}

#[test]
fn test_wide_int_table() {
    run_file("sql-server-wide-int-table.yaml");
}

#[test]
fn test_wide_varchar_table() {
    run_file("sql-server-wide-varchar-table.yaml");
}

#[test]
fn test_wide_text_table() {
    run_file("sql-server-wide-text-table.yaml");
}

#[test]
fn test_large_values_failover() {
    run_file("sql-server-large-values-failover.yaml");
}

#[test]
fn test_wide_table_failover() {
    run_file("sql-server-wide-table-failover.yaml");
}
