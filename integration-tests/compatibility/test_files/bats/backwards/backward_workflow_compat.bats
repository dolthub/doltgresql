#!/usr/bin/env bats
# Backward-only compatibility tests for doltgresql.
#
# These tests exercise workflows where an older doltgresql release writes the
# initial state and the current HEAD build then continues operating on the same
# repo (schema changes, DML, branching, merging).  Unlike bidirectional tests,
# they are only meaningful in one direction — old writes, then new operates —
# so we do not run them with the LEGACY / NEW roles swapped.
#
# Each test manages its own server lifecycle so that it can switch between the
# LEGACY and NEW binaries mid-test.  setup() and teardown() only manage the
# scratch directory.
#
# Environment variables:
#   DOLTGRES_LEGACY_BIN   — path to the "old" doltgres binary
#   DOLTGRES_NEW_BIN      — path to the "new" (HEAD) doltgres binary
#   REPO_DIR              — scratch directory base (empty, just needs to exist)

# Note: helper is one directory up from this file.
load $BATS_TEST_DIRNAME/../helper/common.bash

BATS_REPO=""

setup() {
  BATS_REPO="$BATS_TMPDIR/backward-$$-$RANDOM"
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
# Foreign key on a VARCHAR column from an old repo, then branch + merge.
# Regression: this workflow broke for repos written by v0.56.2 or older because
# those releases stored VARCHAR(N) with ExtendedEnc + a DoltgresType handler,
# while HEAD stores VARCHAR(N) with StringEnc + a nil handler.  Foreign-key
# validation across the mismatched encodings failed.
# ---------------------------------------------------------------------------

@test "backward workflow: fk on varchar column from old repo, then merge" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- Old: create parent with a VARCHAR(10) column and seed rows ---
  old_server_start
  sql <<SQL
CREATE TABLE parent (
  id  INT NOT NULL PRIMARY KEY,
  val VARCHAR(10) NOT NULL UNIQUE
);
INSERT INTO parent VALUES (1, 'apple'), (2, 'banana'), (3, 'cherry');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: create parent');"
  stop_doltgres

  # --- New: create child with FK on VARCHAR(10) referencing parent(val) ---
  new_server_start
  sql <<SQL
CREATE TABLE child (
  id  INT NOT NULL PRIMARY KEY,
  ref VARCHAR(10) NOT NULL,
  CONSTRAINT child_ref_fk FOREIGN KEY (ref) REFERENCES parent(val)
);
SQL

  # Valid inserts — ref values that exist in parent.
  sql -c "INSERT INTO child VALUES (10, 'apple'), (11, 'banana');"
  run sql_csv -c "SELECT count(*) FROM child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2" ]] || false

  # Invalid insert — 'grape' is not present in parent, FK must reject.
  run sql -c "INSERT INTO child VALUES (12, 'grape');"
  [ "$status" -ne 0 ]

  # Row count unchanged after the rejected insert.
  run sql_csv -c "SELECT count(*) FROM child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2" ]] || false

  # dolt_verify_constraints reports no violations. Returns {0} on success.
  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "{0}" ]] || false

  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: create child + fk');"

  # --- Branch, write on both branches, then merge new into main ---
  sql -c "SELECT dolt_branch('new');"

  sql -c "INSERT INTO child VALUES (20, 'cherry');"
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'main: add cherry');"

  sql <<SQL
SELECT dolt_checkout('new');
INSERT INTO child VALUES (30, 'banana');
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'new: add banana on new branch');
SQL

  run sql -c "SELECT dolt_merge('new');"
  [ "$status" -eq 0 ]

  # After merge, main should see rows from both branches.
  run sql_csv -c "SELECT count(*) FROM child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4" ]] || false

  run sql_csv -c "SELECT id, ref FROM child ORDER BY id;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "10,apple" ]]  || false
  [[ "$output" =~ "11,banana" ]] || false
  [[ "$output" =~ "20,cherry" ]] || false
  [[ "$output" =~ "30,banana" ]] || false

  # FK still enforced after merge.
  run sql -c "INSERT INTO child VALUES (99, 'grape');"
  [ "$status" -ne 0 ]

  # And still no reported constraint violations.
  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "{0}" ]] || false

  stop_doltgres
}
