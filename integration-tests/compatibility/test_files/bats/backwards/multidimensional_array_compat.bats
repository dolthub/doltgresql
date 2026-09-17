#!/usr/bin/env bats
# Backward-only compatibility tests for multidimensional arrays.
#
# Array columns created by older releases serialize their type with version 0,
# which cannot hold multidimensional values, while the current HEAD build
# creates array types with version 1, which older releases refuse to read.
# These tests check that both builds agree on what an old repo may hold, and
# that altering an old column's type upgrades it.  Every table is created by
# the old release, since older releases cannot open tables created by HEAD at
# all (their column encoding predates them), which has nothing to do with
# arrays.
#
# Environment variables:
#   DOLTGRES_LEGACY_BIN   — path to the "old" doltgres binary
#   DOLTGRES_NEW_BIN      — path to the "new" (HEAD) doltgres binary
#   REPO_DIR              — scratch directory base (empty, just needs to exist)

load $BATS_TEST_DIRNAME/../helper/common.bash

BATS_REPO=""

setup() {
  BATS_REPO="$BATS_TMPDIR/multidimensional-array-$$-$RANDOM"
  mkdir -p "$BATS_REPO"
}

teardown() {
  stop_doltgres
  rm -rf "$BATS_REPO"
}

old_server_start() { start_doltgres "$DOLTGRES_LEGACY_BIN" "$BATS_REPO" "$BATS_REPO/old.log"; }
new_server_start() { start_doltgres "$DOLTGRES_NEW_BIN"    "$BATS_REPO" "$BATS_REPO/new.log"; }

# old_create_table — the old build creates an array column, seeds a row, and commits.
old_create_table() {
  old_server_start
  sql <<SQL
CREATE TABLE t (pk INT PRIMARY KEY, v INT[]);
INSERT INTO t VALUES (1, '{1,2}');
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'old: create t');
SQL
  stop_doltgres
}

# old_server_refuses <table> — the old build must reject the version 1 array type in <table>, either by
# refusing to start (its startup integrity check reads the schema) or by failing every query on the table.
old_server_refuses() {
  if old_server_start; then
    run sql -c "SELECT * FROM $1;"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "version 1 of types is not supported" ]] || false
    stop_doltgres
  else
    grep -q "version 1 of types is not supported" "$BATS_REPO/old.log"
  fi
}

@test "multidimensional_array_compat: old array column stays readable by old after new writes" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_create_table

  # --- New: one-dimensional writes succeed, multidimensional writes are refused ---
  new_server_start
  sql -c "INSERT INTO t VALUES (2, '{3}');"
  run sql -c "INSERT INTO t VALUES (3, '{{1,2},{3,4}}');"
  [ "$status" -ne 0 ]
  [[ "$output" =~ "multidimensional arrays are not supported by the column's type version" ]] || false
  run sql -c "UPDATE t SET v = '{{1,2},{3,4}}' WHERE pk = 1;"
  [ "$status" -ne 0 ]
  [[ "$output" =~ "multidimensional arrays are not supported by the column's type version" ]] || false
  run sql_csv -c "SELECT pk, v FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ '1,"{1,2}"' ]] || false
  [[ "$output" =~ "2,{3}" ]] || false
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'new: add row');"
  stop_doltgres

  # --- Old: the repo is still readable and writable ---
  old_server_start
  run sql_csv -c "SELECT pk, v FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ '1,"{1,2}"' ]] || false
  [[ "$output" =~ "2,{3}" ]] || false
  sql -c "INSERT INTO t VALUES (3, '{4,5}');"
  run sql_csv -c "SELECT count(*) FROM t;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "3" ] || false
}

@test "multidimensional_array_compat: altering the column type upgrades an old array column" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_create_table

  # --- New: the upgrade path is an ALTER to the same type ---
  new_server_start
  run sql -c "INSERT INTO t VALUES (2, '{{1,2},{3,4}}');"
  [ "$status" -ne 0 ]
  [[ "$output" =~ "multidimensional arrays are not supported by the column's type version" ]] || false
  sql -c "ALTER TABLE t ALTER COLUMN v TYPE INT[];"
  sql -c "INSERT INTO t VALUES (2, '{{1,2},{3,4}}');"
  sql -c "UPDATE t SET v = '{{5},{6}}' WHERE pk = 1;"
  run sql_csv -c "SELECT pk, v FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ '1,"{{5},{6}}"' ]] || false
  [[ "$output" =~ '2,"{{1,2},{3,4}}"' ]] || false
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'new: upgrade v');"
  stop_doltgres

  # --- Old: the upgraded column is now refused ---
  old_server_refuses t
}

@test "multidimensional_array_compat: upgraded array column is refused by old even without multidimensional values" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_create_table

  # --- New: upgrade the column type without writing any multidimensional value ---
  new_server_start
  sql -c "ALTER TABLE t ALTER COLUMN v TYPE INT[];"
  sql -c "INSERT INTO t VALUES (2, '{3}');"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'new: upgrade v');"
  stop_doltgres

  # --- Old: the column's type version, not its data, decides readability ---
  old_server_refuses t

  # --- New: still reads both rows ---
  new_server_start
  run sql_csv -c "SELECT pk, v FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ '1,"{1,2}"' ]] || false
  [[ "$output" =~ "2,{3}" ]] || false
}
