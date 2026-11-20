#!/usr/bin/env bats
# Backward-only compatibility tests for function calls in column expressions.
#
# Column defaults, generated columns, and check constraints are persisted as the
# text of their expressions.  Older releases wrote function calls without a
# schema, while HEAD writes every call with the schema of the function it
# resolved to (pg_catalog for built-in functions), and qualifies the sequence
# named by nextval.  These tests check that expressions written by an older
# release keep calling the same functions, and producing the same values, once
# HEAD reads them, and that HEAD persists them with their schemas once they are
# set again.
#
# Environment variables:
#   DOLTGRES_LEGACY_BIN   — path to the "old" doltgres binary
#   DOLTGRES_NEW_BIN      — path to the "new" (HEAD) doltgres binary
#   REPO_DIR              — scratch directory base (empty, just needs to exist)

load $BATS_TEST_DIRNAME/../helper/common.bash

BATS_REPO=""

setup() {
  BATS_REPO="$BATS_TMPDIR/function-default-$$-$RANDOM"
  mkdir -p "$BATS_REPO"
}

teardown() {
  stop_doltgres
  rm -rf "$BATS_REPO"
}

old_server_start() { start_doltgres "$DOLTGRES_LEGACY_BIN" "$BATS_REPO" "$BATS_REPO/old.log"; }
new_server_start() { start_doltgres "$DOLTGRES_NEW_BIN"    "$BATS_REPO" "$BATS_REPO/new.log"; }

@test "function_default_compat: built-in functions in old column expressions" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- Old: defaults, a generated column, and a check constraint that call built-in functions ---
  old_server_start
  sql <<SQL
CREATE TABLE t (
  pk INT PRIMARY KEY,
  a TEXT DEFAULT upper('q'),
  b INT GENERATED ALWAYS AS (length(a)) STORED,
  c TEXT DEFAULT concat('x', 'y'),
  CHECK (abs(pk) < 100)
);
INSERT INTO t (pk) VALUES (1);
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'old: create t');
SQL
  stop_doltgres

  # --- New: the old expressions produce the same values ---
  new_server_start
  sql -c "INSERT INTO t (pk) VALUES (2);"
  sql -c "INSERT INTO t (pk, a) VALUES (3, 'abcd');"
  run sql -c "INSERT INTO t (pk) VALUES (100);"
  [ "$status" -ne 0 ]
  run sql_csv -c "SELECT pk, a, b, c FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "1,Q,1,xy" ] || false
  [ "${lines[2]}" = "2,Q,1,xy" ] || false
  [ "${lines[3]}" = "3,abcd,4,xy" ] || false
  run sql_csv -c "SELECT column_name, column_default FROM information_schema.columns WHERE table_name = 't' AND column_default IS NOT NULL ORDER BY column_name;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "a,upper('q'::text)" ] || false
  [ "${lines[2]}" = "c,\"concat('x', 'y')\"" ] || false

  # --- New: a column added by HEAD, whose default is persisted with its schema, sits beside the old ones ---
  sql -c "ALTER TABLE t ADD COLUMN d TEXT DEFAULT lower('Z');"
  sql -c "INSERT INTO t (pk) VALUES (4);"
  run sql_csv -c "SELECT pk, a, b, c, d FROM t WHERE pk = 4;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "4,Q,1,xy,z" ] || false
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'new: alter t');"
  stop_doltgres

  # --- New, after a restart: the old and new expressions produce the same values ---
  new_server_start
  sql -c "INSERT INTO t (pk) VALUES (5);"
  run sql -c "INSERT INTO t (pk) VALUES (100);"
  [ "$status" -ne 0 ]
  run sql_csv -c "SELECT pk, a, b, c, d FROM t WHERE pk = 5;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "5,Q,1,xy,z" ] || false
}

@test "function_default_compat: user-defined function in an old column default" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- Old: a default that calls a function in public ---
  old_server_start
  sql <<SQL
CREATE FUNCTION f() RETURNS INT AS \$\$ SELECT 42 \$\$ LANGUAGE sql;
CREATE TABLE t (pk INT PRIMARY KEY, v INT DEFAULT f());
INSERT INTO t (pk) VALUES (1);
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'old: create t');
SQL
  stop_doltgres

  # --- New: the default calls the same function, before and after the table is altered ---
  new_server_start
  sql -c "INSERT INTO t (pk) VALUES (2);"
  sql -c "ALTER TABLE t ADD COLUMN w INT;"
  sql -c "INSERT INTO t (pk) VALUES (3);"
  run sql_csv -c "SELECT pk, v FROM t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "1,42" ] || false
  [ "${lines[2]}" = "2,42" ] || false
  [ "${lines[3]}" = "3,42" ] || false
}

@test "function_default_compat: user-defined function in an old column default outside public" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- Old: defaults that call a function in the table's own schema. Older releases persist the call without its
  # schema, so the default only finds the function when the schema is on the search path. ---
  old_server_start
  sql <<SQL
CREATE SCHEMA s;
CREATE FUNCTION s.g() RETURNS INT AS \$\$ SELECT 7 \$\$ LANGUAGE sql;
SET search_path = s;
CREATE TABLE t (pk INT PRIMARY KEY, v INT DEFAULT s.g());
CREATE TABLE t2 (pk INT PRIMARY KEY, v INT DEFAULT g());
INSERT INTO t (pk) VALUES (1);
INSERT INTO t2 (pk) VALUES (1);
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'old: create s.t and s.t2');
SQL
  run sql -c "INSERT INTO s.t (pk) VALUES (2);"
  [ "$status" -ne 0 ]
  stop_doltgres

  new_server_start
  for table in t t2; do
    # --- New: as with the old release, the default only finds the function when its schema is on the search path ---
    run sql -c "INSERT INTO s.$table (pk) VALUES (2);"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "g" ]] || false
    sql -c "SET search_path = s, public; INSERT INTO s.$table (pk) VALUES (2);"
    run sql_csv -c "SELECT pk, v FROM s.$table ORDER BY pk;"
    [ "$status" -eq 0 ]
    [ "${lines[1]}" = "1,7" ] || false
    [ "${lines[2]}" = "2,7" ] || false

    # --- New: setting the default again persists it with the function's schema, so any search path finds it ---
    sql -c "SET search_path = s, public; ALTER TABLE s.$table ALTER COLUMN v SET DEFAULT g();"
    sql -c "INSERT INTO s.$table (pk) VALUES (3);"
    run sql_csv -c "SELECT v FROM s.$table WHERE pk = 3;"
    [ "$status" -eq 0 ]
    [ "${lines[1]}" = "7" ] || false
  done
}

@test "function_default_compat: sequences in old column defaults" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- Old: a SERIAL column, and defaults that name sequences in the tables' own schema. Older releases persist an
  # unqualified sequence name as written, so the default only finds the sequence when the schema is on the search
  # path. ---
  old_server_start
  sql <<SQL
CREATE TABLE t (id SERIAL PRIMARY KEY, v TEXT);
INSERT INTO t (v) VALUES ('a');
CREATE SCHEMA s;
CREATE SEQUENCE s.seq START 100;
CREATE TABLE s.t (pk INT PRIMARY KEY, n BIGINT DEFAULT nextval('s.seq'));
INSERT INTO s.t (pk) VALUES (1);
SET search_path = s;
CREATE SEQUENCE seq2 START 200;
CREATE TABLE t2 (pk INT PRIMARY KEY, n BIGINT DEFAULT nextval('seq2'));
INSERT INTO t2 (pk) VALUES (1);
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'old: create tables');
SQL
  run sql -c "INSERT INTO s.t2 (pk) VALUES (2);"
  [ "$status" -ne 0 ]
  stop_doltgres

  # --- New: the defaults keep drawing from their sequences, before and after the tables are altered ---
  new_server_start
  sql -c "INSERT INTO t (v) VALUES ('b');"
  sql -c "INSERT INTO s.t (pk) VALUES (2);"
  sql -c "ALTER TABLE t ADD COLUMN w INT;"
  sql -c "ALTER TABLE s.t ADD COLUMN w INT;"
  sql -c "INSERT INTO t (v) VALUES ('c');"
  sql -c "INSERT INTO s.t (pk) VALUES (3);"
  run sql_csv -c "SELECT id, v FROM t ORDER BY id;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "1,a" ] || false
  [ "${lines[2]}" = "2,b" ] || false
  [ "${lines[3]}" = "3,c" ] || false
  run sql_csv -c "SELECT pk, n FROM s.t ORDER BY pk;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "1,100" ] || false
  [ "${lines[2]}" = "2,101" ] || false
  [ "${lines[3]}" = "3,102" ] || false

  # --- New: as with the old release, the unqualified sequence is only found when its schema is on the search path ---
  run sql -c "INSERT INTO s.t2 (pk) VALUES (2);"
  [ "$status" -ne 0 ]
  [[ "$output" =~ "seq2" ]] || false
  sql -c "SET search_path = s, public; INSERT INTO s.t2 (pk) VALUES (2);"
  run sql_csv -c "SELECT pk, n FROM s.t2 ORDER BY pk;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "1,200" ] || false
  [ "${lines[2]}" = "2,201" ] || false

  # --- New: setting the default again persists it with the sequence's schema, so any search path finds it ---
  sql -c "SET search_path = s, public; ALTER TABLE s.t2 ALTER COLUMN n SET DEFAULT nextval('seq2');"
  sql -c "INSERT INTO s.t2 (pk) VALUES (3);"
  run sql_csv -c "SELECT n FROM s.t2 WHERE pk = 3;"
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "202" ] || false
}
