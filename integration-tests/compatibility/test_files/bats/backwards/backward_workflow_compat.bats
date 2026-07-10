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

@test "backward_workflow: fk on varchar column from old repo, then merge" {
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

@test "backward_workflow: dangling child violation surfaces on merge" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- Old: create parent, seed rows ---
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

  # --- New: create child + FK, seed a child that will become dangling ---
  new_server_start
  sql <<SQL
CREATE TABLE child (
  id  INT NOT NULL PRIMARY KEY,
  ref VARCHAR(10) NOT NULL,
  CONSTRAINT child_ref_fk FOREIGN KEY (ref) REFERENCES parent(val)
);
INSERT INTO child VALUES (10, 'apple');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: create child + fk, seed');"

  # No violations at this point.
  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "{0}" ]] || false

  # Branch 'drop_banana': remove banana from parent (no child refs banana yet).
  sql <<SQL
SELECT dolt_branch('drop_banana');
SELECT dolt_checkout('drop_banana');
DELETE FROM parent WHERE val='banana';
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'drop_banana: remove banana');
SQL

  # Back on main: add a child row that references banana (still valid on main).
  sql <<SQL
SELECT dolt_checkout('main');
INSERT INTO child VALUES (20, 'banana');
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'main: add child(20, banana)');
SQL

  # merge should fail
  run sql -c "SELECT dolt_merge('drop_banana');"
  [ "$status" -ne 0 ]

  # merge succeeds with constraint violation if forced
  run sql -c "SET dolt_force_transaction_commit=1; SELECT dolt_merge('drop_banana');"
  [ "$status" -eq 0 ]

  # Verify the violation was recorded against child.
  run sql_csv -c "SELECT count(*) FROM dolt_constraint_violations_child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1" ]] || false

  # And that the violation row is the banana reference (id=20).
  run sql_csv -c "SELECT id, ref FROM dolt_constraint_violations_child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "20,banana" ]] || false

  stop_doltgres
}

@test "backward_workflow: fk on text column from old repo, then merge" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # TEXT is stored with an adaptive encoding.  In v0.56.2 and older that is
  # ExtendedAdaptiveEnc + DoltgresType handler; HEAD stores it as
  # StringAdaptiveEnc + nil handler.  The mixed-handler case exercises the
  # ExtendedAdaptiveEnc branch of convertNativeEncodedFkField.
  old_server_start
  sql <<SQL
CREATE TABLE parent (
  id  INT NOT NULL PRIMARY KEY,
  val TEXT NOT NULL UNIQUE
);
INSERT INTO parent VALUES (1, 'apple'), (2, 'banana'), (3, 'cherry');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: create parent (text)');"
  stop_doltgres

  new_server_start
  sql <<SQL
CREATE TABLE child (
  id  INT NOT NULL PRIMARY KEY,
  ref TEXT NOT NULL,
  CONSTRAINT child_ref_fk FOREIGN KEY (ref) REFERENCES parent(val)
);
SQL

  sql -c "INSERT INTO child VALUES (10, 'apple'), (11, 'banana');"
  run sql_csv -c "SELECT count(*) FROM child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2" ]] || false

  # Invalid insert — 'grape' is not present in parent, FK must reject.
  run sql -c "INSERT INTO child VALUES (12, 'grape');"
  [ "$status" -ne 0 ]

  # No reported violations.
  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "{0}" ]] || false

  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: create child + fk');"

  # Branch, write on both, merge.
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

  run sql_csv -c "SELECT count(*) FROM child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4" ]] || false

  # FK still enforced after merge.
  run sql -c "INSERT INTO child VALUES (99, 'grape');"
  [ "$status" -ne 0 ]

  # Still no violations reported.
  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "{0}" ]] || false

  stop_doltgres
}

@test "backward_workflow: fk from varchar child to text parent" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # Cross-type FK: parent's TEXT column vs child's VARCHAR(10) column.  In the
  # old repo the parent's TEXT lives in ExtendedAdaptiveEnc + DoltgresType
  # handler; HEAD's child VARCHAR(10) lives in StringEnc + nil handler.  Every
  # value stored fits within VARCHAR(10) so no truncation is expected.
  old_server_start
  sql <<SQL
CREATE TABLE parent (
  id  INT NOT NULL PRIMARY KEY,
  val TEXT NOT NULL UNIQUE
);
INSERT INTO parent VALUES (1, 'apple'), (2, 'banana'), (3, 'cherry');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: create parent (text)');"
  stop_doltgres

  new_server_start
  sql <<SQL
CREATE TABLE child (
  id  INT NOT NULL PRIMARY KEY,
  ref VARCHAR(10) NOT NULL,
  CONSTRAINT child_ref_fk FOREIGN KEY (ref) REFERENCES parent(val)
);
SQL

  sql -c "INSERT INTO child VALUES (10, 'apple'), (11, 'banana');"
  run sql_csv -c "SELECT count(*) FROM child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2" ]] || false

  run sql -c "INSERT INTO child VALUES (12, 'grape');"
  [ "$status" -ne 0 ]

  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "{0}" ]] || false

  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: create child + fk (varchar → text)');"
  stop_doltgres
}

@test "backward_workflow: fk from text child to varchar parent" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # Opposite mixing from the previous test: parent's VARCHAR(10) is
  # ExtendedEnc + DoltgresType handler in the old repo; HEAD's child TEXT is
  # StringAdaptiveEnc + nil handler.  Exercises StringAdaptiveEnc source ↔
  # ExtendedEnc target in convertNativeEncodedFkField (and the reverse on
  # merge / parent-diff).
  old_server_start
  sql <<SQL
CREATE TABLE parent (
  id  INT NOT NULL PRIMARY KEY,
  val VARCHAR(10) NOT NULL UNIQUE
);
INSERT INTO parent VALUES (1, 'apple'), (2, 'banana'), (3, 'cherry');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: create parent (varchar)');"
  stop_doltgres

  new_server_start
  sql <<SQL
CREATE TABLE child (
  id  INT NOT NULL PRIMARY KEY,
  ref TEXT NOT NULL,
  CONSTRAINT child_ref_fk FOREIGN KEY (ref) REFERENCES parent(val)
);
SQL

  sql -c "INSERT INTO child VALUES (10, 'apple'), (11, 'banana');"
  run sql_csv -c "SELECT count(*) FROM child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2" ]] || false

  run sql -c "INSERT INTO child VALUES (12, 'grape');"
  [ "$status" -ne 0 ]

  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "{0}" ]] || false

  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: create child + fk (text → varchar)');"
  stop_doltgres
}

@test "backward_workflow: dangling child violation surfaces on merge (text)" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # TEXT analogue of the VARCHAR dangling-child test above.  Exercises the
  # parent-diff direction (ExtendedAdaptiveEnc → StringAdaptiveEnc, i.e. the
  # parent's stored key format → the child's index key format) for adaptive
  # encodings.
  old_server_start
  sql <<SQL
CREATE TABLE parent (
  id  INT NOT NULL PRIMARY KEY,
  val TEXT NOT NULL UNIQUE
);
INSERT INTO parent VALUES (1, 'apple'), (2, 'banana'), (3, 'cherry');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: create parent (text)');"
  stop_doltgres

  new_server_start
  sql <<SQL
CREATE TABLE child (
  id  INT NOT NULL PRIMARY KEY,
  ref TEXT NOT NULL,
  CONSTRAINT child_ref_fk FOREIGN KEY (ref) REFERENCES parent(val)
);
INSERT INTO child VALUES (10, 'apple');
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'head: create child + fk, seed');"

  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "{0}" ]] || false

  # Branch 'drop_banana': remove banana from parent (no child refs banana yet).
  sql <<SQL
SELECT dolt_branch('drop_banana');
SELECT dolt_checkout('drop_banana');
DELETE FROM parent WHERE val='banana';
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'drop_banana: remove banana');
SQL

  # Back on main: add a child row that references banana (still valid on main).
  sql <<SQL
SELECT dolt_checkout('main');
INSERT INTO child VALUES (20, 'banana');
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'main: add child(20, banana)');
SQL

  # Merge should fail without forced commit due to constraint violation.
  run sql -c "SELECT dolt_merge('drop_banana');"
  [ "$status" -ne 0 ]

  # Merge succeeds with constraint violation when forced.
  run sql -c "SET dolt_force_transaction_commit=1; SELECT dolt_merge('drop_banana');"
  [ "$status" -eq 0 ]

  run sql_csv -c "SELECT count(*) FROM dolt_constraint_violations_child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1" ]] || false

  run sql_csv -c "SELECT id, ref FROM dolt_constraint_violations_child;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "20,banana" ]] || false

  stop_doltgres
}

