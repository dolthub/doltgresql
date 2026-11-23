#!/usr/bin/env bats
# Backward-only compatibility tests for altering columns of tables created by older releases.
#
# Releases before v0.56.3 store every column with an extended encoding, while newer builds store
# many types (integers, JSONB, etc.) with native encodings. ALTER TABLE statements that don't
# rewrite the table, such as ALTER COLUMN ... SET DEFAULT, must keep a column on the encoding its
# existing values were written with, or those values are read back incorrectly.
#
# Environment variables:
#   DOLTGRES_LEGACY_BIN   — path to the "old" doltgres binary
#   DOLTGRES_NEW_BIN      — path to the "new" (HEAD) doltgres binary
#   REPO_DIR              — scratch directory base (empty, just needs to exist)

load $BATS_TEST_DIRNAME/../helper/common.bash

BATS_REPO=""

setup() {
  BATS_REPO="$BATS_TMPDIR/alter-column-$$-$RANDOM"
  mkdir -p "$BATS_REPO"
}

teardown() {
  stop_doltgres
  rm -rf "$BATS_REPO"
}

old_server_start() { start_doltgres "$DOLTGRES_LEGACY_BIN" "$BATS_REPO" "$BATS_REPO/old.log"; }
new_server_start() { start_doltgres "$DOLTGRES_NEW_BIN"    "$BATS_REPO" "$BATS_REPO/new.log"; }

# old_create_table — the old build creates a table, seeds a row, and commits.
old_create_table() {
  old_server_start
  sql <<SQL
CREATE TABLE t (pk INT PRIMARY KEY, v INT DEFAULT 7, w BIGINT DEFAULT 200, j JSONB DEFAULT '{"a": 1}', s SMALLINT);
INSERT INTO t (pk, s) VALUES (1, 3);
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'old: create t');
SQL
  stop_doltgres
}

@test "alter_column_compat: altering column defaults keeps existing values" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_create_table

  # --- New: alter every column without changing its type ---
  new_server_start
  sql <<SQL
ALTER TABLE t ALTER COLUMN pk SET DEFAULT 5;
ALTER TABLE t ALTER COLUMN v SET DEFAULT 8;
ALTER TABLE t ALTER COLUMN w SET DEFAULT 300;
ALTER TABLE t ALTER COLUMN j SET DEFAULT '{"b": 2}';
INSERT INTO t (pk, s) VALUES (2, 4);
SQL
  run sql_csv -c "SELECT pk, v, w, j, s FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = '1,7,200,"{""a"": 1}",3' ] || false
  [ "${lines[2]}" = '2,8,300,"{""b"": 2}",4' ] || false
  run sql_csv -c "SELECT pk, v FROM t WHERE pk = 1;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "1,7" ] || false
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'new: alter defaults');"
  stop_doltgres

  # --- Old: still reads the same values ---
  old_server_start
  run sql_csv -c "SELECT pk, v, w, j, s FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = '1,7,200,"{""a"": 1}",3' ] || false
  [ "${lines[2]}" = '2,8,300,"{""b"": 2}",4' ] || false
}

# Setting a column NOT NULL rewrites the table with the current encodings, which older releases can't read.
@test "alter_column_compat: setting a column not null keeps existing values" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_create_table

  new_server_start
  sql -c "ALTER TABLE t ALTER COLUMN s SET NOT NULL;"
  sql -c "INSERT INTO t (pk, s) VALUES (2, 4);"
  run sql_csv -c "SELECT pk, v, w, j, jsonb_typeof(j->'a'), s FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = '1,7,200,"{""a"": 1}",number,3' ] || false
  [ "${lines[2]}" = '2,7,200,"{""a"": 1}",number,4' ] || false
  run sql -c "INSERT INTO t (pk) VALUES (3);"
  [ "$status" -ne 0 ]
}

@test "alter_column_compat: dropping a column default keeps existing values" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_create_table

  new_server_start
  sql -c "ALTER TABLE t ALTER COLUMN w DROP DEFAULT;"
  sql -c "INSERT INTO t (pk, s) VALUES (2, 4);"
  run sql_csv -c "SELECT pk, v, w, s FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "1,7,200,3" ] || false
  [ "${lines[2]}" = "2,7,,4" ] || false
}

# old_create_mixed_table — the old build creates a table with types whose older encodings can't represent every value.
old_create_mixed_table() {
  old_server_start
  sql <<SQL
CREATE TABLE m (k INT PRIMARY KEY, n NUMERIC, jb JSONB, j JSON, spare INT);
INSERT INTO m VALUES (1, 22.5, '{"x": "a\\\\nb"}', '42', 0);
SQL
  stop_doltgres
}

@test "alter_column_compat: dropping a column rewrites existing values unchanged" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_create_mixed_table

  new_server_start
  run sql_csv -c "SELECT k, n::text, jb->>'x', j FROM m ORDER BY k;"
  [ "$status" -eq 0 ]
  before="${lines[1]}"

  sql -c "ALTER TABLE m DROP COLUMN spare;"
  run sql_csv -c "SELECT k, n::text, jb->>'x', j FROM m ORDER BY k;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "$before" ] || false

  sql -c "INSERT INTO m VALUES (2, 'NaN', jsonb_build_object('x', chr(10)), '43');"
  run sql_csv -c "SELECT k, n::text, ascii(jb->>'x'), length(jb->>'x'), j FROM m WHERE k = 2;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "2,NaN,10,1,43" ] || false
}

@test "alter_column_compat: changing the type of another column rewrites existing values unchanged" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_create_mixed_table

  new_server_start
  run sql_csv -c "SELECT k, n::text, jb->>'x', j, spare FROM m ORDER BY k;"
  [ "$status" -eq 0 ]
  before="${lines[1]}"

  sql -c "ALTER TABLE m ALTER COLUMN spare TYPE BIGINT;"
  run sql_csv -c "SELECT k, n::text, jb->>'x', j, spare FROM m ORDER BY k;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "$before" ] || false

  sql -c "INSERT INTO m VALUES (2, 'NaN', jsonb_build_object('x', chr(10)), '43', 5);"
  run sql_csv -c "SELECT k, n::text, ascii(jb->>'x'), length(jb->>'x'), j, spare FROM m WHERE k = 2;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "2,NaN,10,1,43,5" ] || false
}

@test "alter_column_compat: changing the types of columns converts existing values" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  old_server_start
  sql <<SQL
CREATE TABLE c (k INT PRIMARY KEY, d DATE, big DATE, i INT, b BOOLEAN, n NUMERIC, v VARCHAR(10));
INSERT INTO c VALUES (1, '2026-01-01', '2026-01-02', 7, true, 22.5, 'abc'), (2, '2026-03-04', '2026-03-05', -8, false, 1.25, 'abcdefgh');
CREATE TABLE p (k BOOLEAN PRIMARY KEY);
INSERT INTO p VALUES (true);
SQL
  stop_doltgres

  new_server_start
  sql <<SQL
ALTER TABLE c ALTER COLUMN d TYPE TEXT;
ALTER TABLE c ALTER COLUMN big TYPE CHAR(70000) USING repeat(big::text, 7000)::CHAR(70000);
ALTER TABLE c ALTER COLUMN i TYPE BIGINT;
ALTER TABLE c ALTER COLUMN b TYPE INTEGER USING b::INTEGER;
ALTER TABLE c ALTER COLUMN n TYPE NUMERIC(10, 1);
ALTER TABLE c ALTER COLUMN v TYPE VARCHAR(20);
ALTER TABLE p ALTER COLUMN k TYPE INTEGER USING 1;
SQL
  run sql_csv -c "SELECT k, d, length(big), substr(big, 1, 10), i, b, n, v FROM c ORDER BY k;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "1,2026-01-01,70000,2026-01-02,7,1,22.5,abc" ] || false
  [ "${lines[2]}" = "2,2026-03-04,70000,2026-03-05,-8,0,1.3,abcdefgh" ] || false

  run sql -c "ALTER TABLE c ALTER COLUMN v TYPE VARCHAR(5);"
  [ "$status" -ne 0 ]

  sql <<SQL
INSERT INTO c VALUES (3, 'x', repeat('y', 70000), 9000000000, 5, 3.14, 'abcdefghijklmnop');
CREATE TABLE f (k INTEGER REFERENCES p(k));
INSERT INTO f VALUES (1);
SQL
  run sql_csv -c "SELECT k, d, length(big), i, b, n, v FROM c WHERE k = 3;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "3,x,70000,9000000000,5,3.1,abcdefghijklmnop" ] || false
  run sql_csv -c "SELECT dolt_verify_constraints('--all');"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "0" ] || false
}
