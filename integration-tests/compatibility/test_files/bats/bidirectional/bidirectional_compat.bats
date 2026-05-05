#!/usr/bin/env bats
# Bidirectional compatibility tests for doltgresql.
#
# Each test creates an isolated repository and alternates reads/writes between
# two doltgres binaries — an older release (DOLTGRES_LEGACY_BIN) and the current
# HEAD build (DOLTGRES_NEW_BIN).  The tests run in both directions: the runner
# swaps LEGACY and NEW to exercise forward and backward compatibility.
#
# Unlike the other compatibility tests, each test body manages its own server
# lifecycle; setup() and teardown() only manage the scratch directory.
#
# Environment variables:
#   DOLTGRES_LEGACY_BIN   — path to the "old" doltgres binary
#   DOLTGRES_NEW_BIN      — path to the "new" (HEAD) doltgres binary
#   REPO_DIR              — scratch directory base (empty, just needs to exist)

# Note: helper is one directory up from this file.
load $BATS_TEST_DIRNAME/../helper/common.bash

BATS_REPO=""

setup() {
  BATS_REPO="$BATS_TMPDIR/bidir-$$-$RANDOM"
  mkdir -p "$BATS_REPO"
}

teardown() {
  stop_doltgres
  rm -rf "$BATS_REPO"
}

# Convenience wrappers so test bodies read cleanly.
old_server_start() { start_doltgres "$DOLTGRES_LEGACY_BIN" "$BATS_REPO" "$BATS_REPO/old.log"; }
new_server_start() { start_doltgres "$DOLTGRES_NEW_BIN"    "$BATS_REPO" "$BATS_REPO/new.log"; }

# ---------------------------------------------------------------------------
# Test 1: Scalar types DML — INT, VARCHAR, NUMERIC, TIMESTAMP round-trip.
# Four rounds: old → HEAD → old → HEAD.
# ---------------------------------------------------------------------------

@test "bidirectional: scalar types round-trip across versions" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- Setup: old doltgres creates schema and seeds two rows ---
  old_server_start
  sql <<SQL
CREATE TABLE scalars (
  pk         INT NOT NULL PRIMARY KEY,
  c_int      INT,
  c_varchar  VARCHAR(255),
  c_numeric  NUMERIC(10,2),
  c_ts       TIMESTAMP
);
INSERT INTO scalars VALUES
  (1, 100, 'old-row-1', 10.50, '2024-01-01 10:00:00'),
  (2, 200, 'old-row-2', 20.75, '2024-06-15 12:30:00');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: initial data');"
  stop_doltgres

  # --- Round 1: HEAD reads old's rows, inserts its own ---
  new_server_start
  run sql_csv -c "SELECT pk, c_varchar, c_numeric FROM scalars WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,old-row-1,10.50" ]] || false

  run sql_csv -c "SELECT count(*) FROM scalars;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2" ]] || false

  sql -c "INSERT INTO scalars VALUES (3, 300, 'head-row-3', 30.25, '2025-01-15 08:00:00');"
  sql -c "UPDATE scalars SET c_varchar='head-updated-1' WHERE pk=1;"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: insert row 3, update row 1');"
  stop_doltgres

  # --- Round 2: old reads HEAD's changes ---
  old_server_start
  run sql_csv -c "SELECT pk, c_varchar FROM scalars WHERE pk=3;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3,head-row-3" ]] || false

  run sql_csv -c "SELECT pk, c_varchar FROM scalars WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,head-updated-1" ]] || false

  sql -c "INSERT INTO scalars VALUES (4, 400, 'old-row-4', 40.00, '2025-03-01 09:00:00');"
  sql -c "DELETE FROM scalars WHERE pk=2;"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: insert row 4, delete row 2');"
  stop_doltgres

  # --- Round 3: HEAD reads old's changes ---
  new_server_start
  run sql_csv -c "SELECT pk, c_varchar, c_numeric FROM scalars WHERE pk=4;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4,old-row-4,40.00" ]] || false

  run sql_csv -c "SELECT count(*) FROM scalars WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "0" ]] || false

  sql -c "INSERT INTO scalars VALUES (5, 500, 'head-row-5', 50.50, '2025-06-01 14:00:00');"
  sql -c "UPDATE scalars SET c_numeric=99.99 WHERE pk=4;"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: insert row 5, update row 4 numeric');"
  stop_doltgres

  # --- Round 4: old reads final state ---
  old_server_start
  run sql_csv -c "SELECT count(*) FROM scalars;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4" ]] || false   # pks 1, 3, 4, 5

  run sql_csv -c "SELECT pk, c_numeric FROM scalars WHERE pk=4;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4,99.99" ]] || false

  run sql_csv -c "SELECT pk, c_varchar FROM scalars WHERE pk=5;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "5,head-row-5" ]] || false
  stop_doltgres
}

# ---------------------------------------------------------------------------
# Test 2: Large TEXT values — out-of-band storage round-trip.
# ---------------------------------------------------------------------------

@test "bidirectional: large text values round-trip" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # Setup: old creates table, inserts small and large text
  old_server_start
  sql <<SQL
CREATE TABLE texts (
  pk     INT NOT NULL PRIMARY KEY,
  c_text TEXT
);
INSERT INTO texts VALUES
  (1, 'old-small'),
  (2, repeat('A', 70000));
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: initial texts');"
  stop_doltgres

  # Round 1: HEAD reads both rows, inserts its own large text
  new_server_start
  run sql_csv -c "SELECT pk, c_text FROM texts WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "old-small" ]] || false

  run sql_csv -c "SELECT pk, length(c_text) FROM texts WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2,70000" ]] || false

  sql -c "INSERT INTO texts VALUES (3, repeat('H', 80000));"
  sql -c "UPDATE texts SET c_text=repeat('U', 75000) WHERE pk=1;"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: large insert, large update');"
  stop_doltgres

  # Round 2: old reads HEAD's large values
  old_server_start
  run sql_csv -c "SELECT pk, length(c_text) FROM texts WHERE pk=3;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3,80000" ]] || false

  run sql_csv -c "SELECT pk, length(c_text) FROM texts WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,75000" ]] || false

  sql -c "UPDATE texts SET c_text=repeat('V', 90000) WHERE pk=2;"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: update row 2 to even larger');"
  stop_doltgres

  # Round 3: HEAD reads old's in-place update
  new_server_start
  run sql_csv -c "SELECT pk, length(c_text) FROM texts WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2,90000" ]] || false

  run sql_csv -c "SELECT count(*) FROM texts;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3" ]] || false
  stop_doltgres
}

# ---------------------------------------------------------------------------
# Test 3: ADD COLUMN DDL — both versions add columns to the same table.
# ---------------------------------------------------------------------------

@test "bidirectional: add columns from both versions" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # Setup: old creates minimal table
  old_server_start
  sql <<SQL
CREATE TABLE evolving (
  pk     INT NOT NULL PRIMARY KEY,
  c_base INT
);
INSERT INTO evolving VALUES (1, 10), (2, 20), (3, 30);
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: base schema');"
  stop_doltgres

  # Round 1: HEAD adds TEXT and DATE columns
  new_server_start
  sql -c "ALTER TABLE evolving ADD COLUMN c_text TEXT;"
  sql -c "ALTER TABLE evolving ADD COLUMN c_date DATE;"
  sql -c "UPDATE evolving SET c_text='text-' || pk::text, c_date='2025-01-01';"
  sql -c "INSERT INTO evolving VALUES (4, 40, 'text-4', '2025-02-01');"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: add text and date columns');"
  stop_doltgres

  # Round 2: old reads HEAD's new columns, adds its own
  old_server_start
  run sql_csv -c "SELECT pk, c_text, c_date FROM evolving WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,text-1,2025-01-01" ]] || false

  run sql_csv -c "SELECT pk, c_text FROM evolving WHERE pk=4;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4,text-4" ]] || false

  sql -c "ALTER TABLE evolving ADD COLUMN c_int2 INT;"
  sql -c "ALTER TABLE evolving ADD COLUMN c_numeric NUMERIC(8,2);"
  sql -c "UPDATE evolving SET c_int2=pk*100, c_numeric=pk*1.5;"
  sql -c "INSERT INTO evolving VALUES (5, 50, 'text-5', '2025-03-01', 500, 7.50);"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: add int2 and numeric columns');"
  stop_doltgres

  # Round 3: HEAD reads all 4 added columns
  new_server_start
  run sql_csv -c "SELECT pk, c_text, c_date, c_int2, c_numeric FROM evolving WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,text-1,2025-01-01,100,1.50" ]] || false

  run sql_csv -c "SELECT pk, c_text, c_int2, c_numeric FROM evolving WHERE pk=5;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "5,text-5,500,7.50" ]] || false
  stop_doltgres
}

# ---------------------------------------------------------------------------
# Test 4: Branch and merge — both versions create branches, merge across
# version boundaries.
# ---------------------------------------------------------------------------

@test "bidirectional: branch and merge across versions" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # Setup: old creates repo with base table
  old_server_start
  sql <<SQL
CREATE TABLE shared (
  pk  INT NOT NULL PRIMARY KEY,
  val VARCHAR(100),
  src VARCHAR(20)
);
INSERT INTO shared VALUES (1, 'base-1', 'old'), (2, 'base-2', 'old');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: base data');"
  stop_doltgres

  # Round 1: HEAD creates a feature branch
  new_server_start
  sql <<SQL
SELECT dolt_branch('head_feature');
SELECT dolt_checkout('head_feature');
INSERT INTO shared VALUES (10, 'head-feature-10', 'head'), (11, 'head-feature-11', 'head');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: feature branch inserts');"
  sql -c "SELECT dolt_checkout('main');"
  stop_doltgres

  # Round 2: old creates its own branch, merges HEAD's feature branch
  old_server_start
  sql <<SQL
SELECT dolt_branch('old_branch');
SELECT dolt_checkout('old_branch');
INSERT INTO shared VALUES (20, 'old-branch-20', 'old');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: old_branch insert');"
  sql <<SQL
SELECT dolt_checkout('main');
INSERT INTO shared VALUES (3, 'base-3', 'old');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: main insert');"

  sql -c "SELECT dolt_merge('head_feature');"
  run sql_csv -c "SELECT count(*) FROM shared;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "5" ]] || false   # 1,2,3,10,11

  run sql_csv -c "SELECT pk, val FROM shared WHERE pk=10;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "10,head-feature-10" ]] || false
  stop_doltgres

  # Round 3: HEAD reads merged state, merges old_branch
  new_server_start
  run sql_csv -c "SELECT count(*) FROM shared;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "5" ]] || false

  run sql_csv -c "SELECT pk, val FROM shared WHERE pk=3;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3,base-3" ]] || false

  sql -c "SELECT dolt_merge('old_branch');"
  run sql_csv -c "SELECT count(*) FROM shared;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "6" ]] || false   # 1,2,3,10,11,20

  run sql_csv -c "SELECT pk, val FROM shared WHERE pk=20;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "20,old-branch-20" ]] || false
  stop_doltgres
}

# ---------------------------------------------------------------------------
# Test 5: Comprehensive type coverage — both versions add columns of different
# type families across rounds.
# ---------------------------------------------------------------------------

@test "bidirectional: comprehensive type coverage across versions" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # Setup: old creates minimal table
  old_server_start
  sql <<SQL
CREATE TABLE typed (
  pk INT NOT NULL PRIMARY KEY
);
INSERT INTO typed (pk) VALUES (1), (2), (3);
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: pk-only base');"
  stop_doltgres

  # Round 1: HEAD adds numeric columns
  new_server_start
  sql -c "ALTER TABLE typed ADD COLUMN c_smallint SMALLINT;"
  sql -c "ALTER TABLE typed ADD COLUMN c_bigint BIGINT;"
  sql -c "ALTER TABLE typed ADD COLUMN c_real REAL;"
  sql -c "ALTER TABLE typed ADD COLUMN c_double DOUBLE PRECISION;"
  sql -c "UPDATE typed SET c_smallint=pk*10, c_bigint=pk*1000000, c_real=pk*1.5, c_double=pk*2.5;"
  sql -c "INSERT INTO typed (pk, c_smallint, c_bigint, c_real, c_double) VALUES (4, 40, 4000000, 6.0, 10.0);"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: add numeric columns');"
  stop_doltgres

  # Round 2: old reads HEAD's numeric columns, adds text/binary columns
  old_server_start
  run sql_csv -c "SELECT pk, c_smallint, c_bigint FROM typed WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,10,1000000" ]] || false

  sql -c "ALTER TABLE typed ADD COLUMN c_varchar VARCHAR(255);"
  sql -c "ALTER TABLE typed ADD COLUMN c_bytea BYTEA;"
  sql -c "UPDATE typed SET c_varchar='varchar-' || pk::text, c_bytea='\xDEAD';"
  sql -c "INSERT INTO typed (pk, c_varchar, c_bytea) VALUES (5, 'varchar-5', '\xBEEF');"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: add varchar and bytea columns');"
  stop_doltgres

  # Round 3: HEAD reads old's string/binary columns, adds temporal/numeric
  new_server_start
  run sql_csv -c "SELECT pk, c_varchar FROM typed WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,varchar-1" ]] || false

  run sql_csv -c "SELECT pk, c_varchar FROM typed WHERE pk=5;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "5,varchar-5" ]] || false

  sql -c "ALTER TABLE typed ADD COLUMN c_date DATE;"
  sql -c "ALTER TABLE typed ADD COLUMN c_numeric NUMERIC(10,3);"
  sql -c "UPDATE typed SET c_date='2025-01-01', c_numeric=pk*3.141;"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: add temporal and numeric columns');"
  stop_doltgres

  # Round 4: old reads HEAD's temporal/numeric columns, adds boolean/jsonb
  old_server_start
  run sql_csv -c "SELECT pk, c_date, c_numeric FROM typed WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,2025-01-01,3.141" ]] || false

  sql -c "ALTER TABLE typed ADD COLUMN c_boolean BOOLEAN;"
  sql -c "ALTER TABLE typed ADD COLUMN c_jsonb JSONB;"
  sql -c "UPDATE typed SET c_boolean=(pk % 2 = 1), c_jsonb=json_build_object('pk', pk);"
  sql -c "INSERT INTO typed (pk, c_varchar, c_date, c_numeric, c_boolean, c_jsonb) VALUES (6, 'varchar-6', '2025-06-01', 18.847, false, '{\"old\":true}');"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: add boolean and jsonb columns');"
  stop_doltgres

  # Round 5: HEAD reads old's boolean/jsonb, does a final insert using all columns
  new_server_start
  run sql_csv -c "SELECT pk, c_boolean FROM typed WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,t" ]] || false

  run sql_csv -c "SELECT pk, c_boolean FROM typed WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2,f" ]] || false

  run sql_csv -c "SELECT pk, c_jsonb->>'old' FROM typed WHERE pk=6;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "true" ]] || false

  sql -c "INSERT INTO typed (pk, c_smallint, c_bigint, c_varchar, c_date, c_numeric, c_boolean, c_jsonb) VALUES (7, 70, 7000000, 'varchar-7', '2025-07-01', 21.988, true, '{\"head\":true}');"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: final insert using all columns');"
  stop_doltgres

  # Round 6: old reads HEAD's final insert — all columns visible
  old_server_start
  run sql_csv -c "SELECT pk, c_smallint, c_bigint, c_varchar, c_numeric FROM typed WHERE pk=7;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "7,70,7000000,varchar-7,21.988" ]] || false

  run sql_csv -c "SELECT count(*) FROM typed;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "7" ]] || false
  stop_doltgres
}

# ---------------------------------------------------------------------------
# Test 6: JSONB round-trip — inline and large JSONB documents across versions.
# ---------------------------------------------------------------------------

@test "bidirectional: jsonb round-trip across versions" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # Setup: old creates table with small and large JSONB
  old_server_start
  sql <<SQL
CREATE TABLE jsondocs (
  pk       INT NOT NULL PRIMARY KEY,
  c_small  JSONB,
  c_big    JSONB
);
INSERT INTO jsondocs VALUES
  (1, '{"key":"val1","num":100}', '{"meta":"small"}'),
  (2, '{"key":"val2"}', json_build_object('big', repeat('x', 5000))::jsonb);
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: initial jsonb');"
  stop_doltgres

  # Round 1: HEAD reads both rows, inserts its own
  new_server_start
  run sql_csv -c "SELECT pk, c_small->>'key', c_small->>'meta' FROM jsondocs WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "val1" ]] || false

  run sql_csv -c "SELECT pk, length(c_big->>'big') FROM jsondocs WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2,5000" ]] || false

  sql -c "INSERT INTO jsondocs VALUES (3, '{\"head\":true}', json_build_object('hbig', repeat('H', 6000))::jsonb);"
  sql -c "UPDATE jsondocs SET c_small=c_small || '{\"updated\":true}' WHERE pk=1;"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: insert row 3, update row 1');"
  stop_doltgres

  # Round 2: old reads HEAD's changes
  old_server_start
  run sql_csv -c "SELECT pk, c_small->>'head' FROM jsondocs WHERE pk=3;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "true" ]] || false

  run sql_csv -c "SELECT pk, c_small->>'updated' FROM jsondocs WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "true" ]] || false

  run sql_csv -c "SELECT pk, length(c_big->>'hbig') FROM jsondocs WHERE pk=3;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3,6000" ]] || false

  run sql_csv -c "SELECT count(*) FROM jsondocs;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3" ]] || false
  stop_doltgres
}
