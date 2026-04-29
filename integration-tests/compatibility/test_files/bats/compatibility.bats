#!/usr/bin/env bats
# Backward / forward compatibility tests for doltgresql.
#
# Each test starts a fresh doltgres server against a copy of the pre-created repo
# (REPO_DIR) and verifies that the binary under test can correctly read and write
# the data created by the other version.
#
# Environment variables consumed by this file:
#   REPO_DIR               — path to the data directory created by setup_repo.sh
#   DOLTGRES_TEST_BIN      — doltgres binary to use (default: doltgres from PATH)

load $BATS_TEST_DIRNAME/helper/common.bash

BATS_REPO=""

setup() {
  BATS_REPO="$BATS_TMPDIR/compat-repo-$$-$RANDOM"
  copy_repo "$REPO_DIR" "$BATS_REPO"
  start_doltgres "${DOLTGRES_TEST_BIN:-doltgres}" "$BATS_REPO" "$BATS_REPO/server.log"
}

teardown() {
  stop_doltgres
  rm -rf "$BATS_REPO"
}

# ---------------------------------------------------------------------------
# Sanity / version checks
# ---------------------------------------------------------------------------

@test "compatibility: server is accessible" {
  run sql -c "SELECT 1;"
  [ "$status" -eq 0 ]
}

@test "compatibility: dolt_version returns a version string" {
  run sql -c "SELECT dolt_version();"
  [ "$status" -eq 0 ]
  [[ "$output" =~ [0-9]+\.[0-9]+\.[0-9]+ ]] || false
}

# ---------------------------------------------------------------------------
# Branch enumeration
# ---------------------------------------------------------------------------

@test "compatibility: expected branches exist" {
  run sql_csv -c "SELECT name FROM dolt_branches ORDER BY name;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "check_merge" ]] || false
  [[ "$output" =~ "init" ]]        || false
  [[ "$output" =~ "main" ]]        || false
  [[ "$output" =~ "other" ]]       || false
}

# ---------------------------------------------------------------------------
# Working set is clean on main
# ---------------------------------------------------------------------------

@test "compatibility: working set is clean on main" {
  run sql -c "SELECT * FROM dolt_status;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "(0 rows)" ]] || false
}

# ---------------------------------------------------------------------------
# Schema on branch init (original schema before alterations)
# ---------------------------------------------------------------------------

@test "compatibility: schema on branch init has original abc columns" {
  run sql <<SQL
SELECT dolt_checkout('init');
SELECT column_name FROM information_schema.columns
  WHERE table_name = 'abc' ORDER BY ordinal_position;
SQL
  [ "$status" -eq 0 ]
  [[ "$output" =~ "pk" ]] || false
  [[ "$output" =~ " a" ]] || false
  [[ "$output" =~ " b" ]] || false
  [[ "$output" =~ " w" ]] || false
  [[ "$output" =~ " x" ]] || false
}

# ---------------------------------------------------------------------------
# Data on branch init
# ---------------------------------------------------------------------------

@test "compatibility: data on branch init matches initial insert" {
  run sql_csv <<SQL
SELECT dolt_checkout('init');
SELECT pk, a, b, w, x FROM abc ORDER BY pk;
SQL
  [ "$status" -eq 0 ]
  [[ "$output" =~ "0,asdf,1.1,0,0" ]] || false
  [[ "$output" =~ "1,asdf,1.1,0,0" ]] || false
  [[ "$output" =~ "2,asdf,1.1,0,0" ]] || false
}

# ---------------------------------------------------------------------------
# Schema on main (after ALTER TABLE DROP/ADD)
# ---------------------------------------------------------------------------

@test "compatibility: schema on main has altered abc columns" {
  run sql <<SQL
SELECT column_name FROM information_schema.columns
  WHERE table_name = 'abc' ORDER BY ordinal_position;
SQL
  [ "$status" -eq 0 ]
  [[ "$output" =~ "pk" ]] || false
  [[ "$output" =~ " a" ]] || false
  [[ "$output" =~ " b" ]] || false
  [[ "$output" =~ " x" ]] || false
  [[ "$output" =~ " y" ]] || false
  # 'w' was dropped on main
  [[ ! "$output" =~ " w" ]] || false
}

# ---------------------------------------------------------------------------
# Data on main
# ---------------------------------------------------------------------------

@test "compatibility: data on main matches expected changes" {
  run sql_csv -c "SELECT pk, a, x, y FROM abc ORDER BY pk;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "0,asdf,1,121" ]] || false
  [[ "$output" =~ "2,asdf,0,121" ]] || false
  [[ "$output" =~ "3,data,0,121" ]] || false
}

# ---------------------------------------------------------------------------
# Schema on branch other (different ALTER TABLE than main)
# ---------------------------------------------------------------------------

@test "compatibility: schema on branch other has w and z columns" {
  run sql <<SQL
SELECT dolt_checkout('other');
SELECT column_name FROM information_schema.columns
  WHERE table_name = 'abc' ORDER BY ordinal_position;
SQL
  [ "$status" -eq 0 ]
  [[ "$output" =~ "pk" ]] || false
  [[ "$output" =~ " w" ]] || false
  [[ "$output" =~ " z" ]] || false
  # 'x' was dropped on other
  [[ ! "$output" =~ " x" ]] || false
}

# ---------------------------------------------------------------------------
# Data on branch other
# ---------------------------------------------------------------------------

@test "compatibility: data on branch other matches expected changes" {
  run sql_csv <<SQL
SELECT dolt_checkout('other');
SELECT pk, a, w, z FROM abc ORDER BY pk;
SQL
  [ "$status" -eq 0 ]
  [[ "$output" =~ "0,asdf,1,122" ]] || false
  [[ "$output" =~ "1,asdf,0,122" ]] || false
  [[ "$output" =~ "4,data,0,122" ]] || false
  # pk=2 was deleted on other
  [[ ! "$output" =~ ",2," ]] || false
}

# ---------------------------------------------------------------------------
# big table
# ---------------------------------------------------------------------------

@test "compatibility: big table has 1000 rows" {
  run sql_csv -c "SELECT count(*) FROM big;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1000" ]] || false
}

@test "compatibility: big table supports delete and insert" {
  sql -c "DELETE FROM big WHERE pk IN (71, 331, 881);"
  run sql_csv -c "SELECT count(*) FROM big;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "997" ]] || false

  sql -c "INSERT INTO big VALUES (1001, 'foo'), (1002, 'bar'), (1003, 'baz');"
  run sql_csv -c "SELECT count(*) FROM big;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1000" ]] || false

  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'modified big');"
}

# ---------------------------------------------------------------------------
# View
# ---------------------------------------------------------------------------

@test "compatibility: view1 is queryable" {
  run sql -c "SELECT * FROM view1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4" ]] || false
}

@test "compatibility: all_types_view has 3 rows" {
  run sql_csv -c "SELECT count(*) FROM all_types_view;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3" ]] || false
}

# ---------------------------------------------------------------------------
# DML on existing tables (write test)
# ---------------------------------------------------------------------------

@test "compatibility: can insert and read back on main" {
  sql -c "INSERT INTO abc (pk, a, b, x, y) VALUES (99, 'new', 9.9, 9, 99);"
  run sql_csv -c "SELECT pk, a, x, y FROM abc WHERE pk=99;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "99,new,9,99" ]] || false

  sql -c "DELETE FROM abc WHERE pk=99;"
}

@test "compatibility: can update a row on main" {
  sql -c "UPDATE abc SET a='updated' WHERE pk=0;"
  run sql_csv -c "SELECT pk, a FROM abc WHERE pk=0;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "0,updated" ]] || false
}

@test "compatibility: dml is committable" {
  sql -c "INSERT INTO abc (pk, a, b, x, y) VALUES (200, 'commit-test', 1.0, 1, 1);"
  sql -c "SELECT dolt_add('.');"
  run sql -c "SELECT length(dolt_commit('-m', 'compat dml commit')::text);"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "34" ]] || false
}

# ---------------------------------------------------------------------------
# Check constraint (def table)
# ---------------------------------------------------------------------------

@test "compatibility: check constraint is enforced" {
  run sql -c "INSERT INTO def VALUES (-1);"
  [ "$status" -ne 0 ]
}

@test "compatibility: valid insert into def succeeds" {
  run sql -c "INSERT INTO def VALUES (100);"
  [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# Merge: check_merge into main (should succeed cleanly)
# ---------------------------------------------------------------------------

@test "compatibility: clean merge of check_merge into main succeeds" {
  run sql <<SQL
SELECT dolt_merge('check_merge');
SQL
  [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# Merge: other into main (should produce a conflict on abc)
# ---------------------------------------------------------------------------

@test "compatibility: conflicting merge of other into main is detected" {
  run sql -c "SELECT dolt_merge('other');"
  # The merge should either return a non-zero status or report conflicts
  # in dolt_conflicts — either outcome confirms conflict detection works.
  if [ "$status" -eq 0 ]; then
    run sql_csv -c "SELECT count(*) FROM dolt_conflicts;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ [1-9] ]] || false
  fi
}
