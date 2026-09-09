#!/usr/bin/env bats
# Storage format upgrade tests for doltgresql.
#
# These tests verify the backwards-INcompatibility contract of storage format extensions: when a new
# doltgresql release writes data using a storage feature an older release doesn't understand, the older
# release must fail loudly and force an upgrade, rather than silently misreading the data or leaking
# chunk references. Data that doesn't use the new feature must remain fully readable by the older
# release.
#
# Like the backward workflow tests, these are only meaningful in one direction — new writes, old
# encounters — so they are not run with the LEGACY / NEW roles swapped. Each test manages its own
# server lifecycle.
#
# Environment variables:
#   DOLTGRES_LEGACY_BIN   — path to the "old" doltgres binary
#   DOLTGRES_NEW_BIN      — path to the "new" (HEAD) doltgres binary
#   REPO_DIR              — scratch directory base (empty, just needs to exist)

load $BATS_TEST_DIRNAME/../helper/common.bash

BATS_REPO=""

setup() {
  BATS_REPO="$BATS_TMPDIR/format-upgrade-$$-$RANDOM"
  mkdir -p "$BATS_REPO"
}

teardown() {
  stop_doltgres
  rm -rf "$BATS_REPO"
}

# Convenience wrappers so test bodies read cleanly.
old_server_start() { start_doltgres "$DOLTGRES_LEGACY_BIN" "$BATS_REPO" "$BATS_REPO/old.log"; }
new_server_start() { start_doltgres "$DOLTGRES_NEW_BIN"    "$BATS_REPO" "$BATS_REPO/new.log"; }

# Out-of-band key values: adaptive-encoded primary key columns (TEXT etc.) whose content is large
# enough to be stored out of band embed a chunk address in the key tuple. New releases record those
# addresses in the key_address_offsets field of tree nodes (so gc/push/clone retain the chunks); the
# field did not exist in older releases, whose readers reject nodes carrying unknown fields.
@test "format_upgrade: old clients must upgrade after a new client writes out-of-band key values" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- New: write a table whose text primary keys are stored out of band, plus a control table ---
  new_server_start
  sql <<SQL
CREATE TABLE control (id INT NOT NULL PRIMARY KEY, val VARCHAR(100));
INSERT INTO control VALUES (1, 'readable-by-all');
CREATE TABLE oob_keys (big TEXT NOT NULL PRIMARY KEY, n INT);
INSERT INTO oob_keys SELECT lpad(i::text, 8, '0') || repeat('k', 19992), i FROM generate_series(1, 3) AS g(i);
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'new: out-of-band key values');"

  run sql_csv -c "SELECT count(*) FROM oob_keys WHERE length(big) = 20000;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3" ]] || false
  stop_doltgres

  # --- Old: the new storage field forces an upgrade. Depending on how old the release is, this
  # surfaces in one of two ways, both loud rather than silent: releases that understand the rest of
  # the data start up, serve everything that predates the new field, and refuse to read the data
  # that uses it (v0.56.5); releases too old to even read the schemas involved fail to open the
  # database at all, because startup iterates every table (v0.56.2 and earlier).
  if old_server_start; then
    # data that doesn't use the new storage field is still fully readable
    run sql_csv -c "SELECT val FROM control WHERE id = 1;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "readable-by-all" ]] || false

    # ... but the table with out-of-band key values is not: reads fail with an explicit
    # unknown-fields error rather than silently misreading the data
    run sql -c "SELECT count(*) FROM oob_keys;"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "unknown fields" ]] || false
    stop_doltgres
  else
    # the old release could not open the database: require the loud failure in its log
    grep -Eq "unknown fields|panic" "$BATS_REPO/old.log"
  fi

  # --- New again: after upgrading, everything is readable ---
  new_server_start
  run sql_csv -c "SELECT count(*) FROM oob_keys WHERE length(big) = 20000;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3" ]] || false
  run sql_csv -c "SELECT val FROM control WHERE id = 1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "readable-by-all" ]] || false
}

# Index column options: every index created by a new release records the sort order of its columns in
# the index schema, because Postgres places NULLs last in an ascending index while the storage default
# places them first. Older releases reject schemas carrying the unknown fields, so any table with a
# secondary index created by a new release forces an upgrade, while tables without one stay readable.
@test "format_upgrade: old clients must upgrade after a new client creates a secondary index" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- New: write a table with a plain secondary index, plus a control table ---
  new_server_start
  sql <<SQL
CREATE TABLE control (id INT NOT NULL PRIMARY KEY, val VARCHAR(100));
INSERT INTO control VALUES (1, 'readable-by-all');
CREATE TABLE indexed (id INT NOT NULL PRIMARY KEY, n INT);
CREATE INDEX indexed_n ON indexed (n);
INSERT INTO indexed VALUES (1, 10), (2, NULL), (3, 5);
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'new: secondary index');"

  run sql_csv -c "SELECT id FROM indexed WHERE n > 6;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1" ]] || false
  stop_doltgres

  # --- Old: the index schema fields force an upgrade, loudly, in one of the two ways described above ---
  if old_server_start; then
    run sql_csv -c "SELECT val FROM control WHERE id = 1;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "readable-by-all" ]] || false

    run sql -c "SELECT count(*) FROM indexed;"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "unknown fields" ]] || false
    stop_doltgres
  else
    grep -Eq "unknown fields|panic" "$BATS_REPO/old.log"
  fi

  # --- New again: after upgrading, everything is readable ---
  new_server_start
  run sql_csv -c "SELECT id FROM indexed WHERE n > 6;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1" ]] || false
  run sql_csv -c "SELECT val FROM control WHERE id = 1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "readable-by-all" ]] || false
}

# Indexes created by older releases store their columns ascending with NULLs first and carry no order fields. A new
# release must leave them that way, also when it changes the indexed column's type or the rest of the table, so that
# tables which never gain a new index stay readable by the older release.
@test "format_upgrade: new clients keep old indexes readable by old clients through schema changes" {
  [ -n "$DOLTGRES_LEGACY_BIN" ] || skip "requires DOLTGRES_LEGACY_BIN"
  [ -n "$DOLTGRES_NEW_BIN"    ] || skip "requires DOLTGRES_NEW_BIN"

  # --- Old: write a table with a secondary index on a nullable column ---
  # Every column type here is stored the same way by every release. Releases before v0.56.3 cannot read
  # INT, TEXT, and similar columns once a new release rewrites the table with Dolt's native encodings, which
  # has nothing to do with indexes.
  old_server_start
  sql <<SQL
CREATE TABLE indexed (id UUID NOT NULL PRIMARY KEY, n DATE, flag BOOLEAN);
CREATE INDEX indexed_n ON indexed (n);
INSERT INTO indexed VALUES
  ('00000000-0000-0000-0000-000000000001', '2024-01-10', true),
  ('00000000-0000-0000-0000-000000000002', NULL, false),
  ('00000000-0000-0000-0000-000000000003', '2024-01-05', true);
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'old: table with a secondary index');"
  stop_doltgres

  # --- New: change the indexed column's type and the rest of the table, and write through the index ---
  new_server_start
  sql <<SQL
ALTER TABLE indexed ALTER COLUMN n TYPE TIMESTAMP;
ALTER TABLE indexed ADD COLUMN extra DATE;
ALTER TABLE indexed RENAME COLUMN flag TO active;
INSERT INTO indexed VALUES
  ('00000000-0000-0000-0000-000000000004', '2024-01-07 12:00:00', true, '2024-02-01'),
  ('00000000-0000-0000-0000-000000000005', NULL, false, '2024-02-02');
UPDATE indexed SET n = '2024-01-11' WHERE id = '00000000-0000-0000-0000-000000000001';
DELETE FROM indexed WHERE id = '00000000-0000-0000-0000-000000000003';
SQL
  sql -c "SELECT dolt_add('.'); SELECT dolt_commit('-m', 'new: schema changes around the old index');"

  # the old index keeps its original order, which is NULLS FIRST in Postgres terms
  run sql_csv -c "SELECT indexdef FROM pg_indexes WHERE indexname = 'indexed_n';"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "CREATE INDEX indexed_n ON public.indexed USING btree (n NULLS FIRST)" ]] || false
  run sql_csv -c "SELECT id FROM indexed WHERE n > '2024-01-06' ORDER BY n;"
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" == "00000000-0000-0000-0000-000000000004" ]] || false
  [[ "${lines[2]}" == "00000000-0000-0000-0000-000000000001" ]] || false
  stop_doltgres

  # --- Old: everything the new release wrote is readable, through the index too ---
  old_server_start
  run sql_csv -c "SELECT id, n, active, extra FROM indexed ORDER BY id;"
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" == "00000000-0000-0000-0000-000000000001,2024-01-11 00:00:00,t," ]] || false
  [[ "${lines[2]}" == "00000000-0000-0000-0000-000000000002,,f," ]] || false
  [[ "${lines[3]}" == "00000000-0000-0000-0000-000000000004,2024-01-07 12:00:00,t,2024-02-01" ]] || false
  [[ "${lines[4]}" == "00000000-0000-0000-0000-000000000005,,f,2024-02-02" ]] || false
  run sql_csv -c "SELECT id FROM indexed WHERE n = '2024-01-07 12:00:00';"
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" == "00000000-0000-0000-0000-000000000004" ]] || false
  run sql_csv -c "SELECT id FROM indexed WHERE n IS NULL ORDER BY id;"
  [ "$status" -eq 0 ]
  [[ "${lines[1]}" == "00000000-0000-0000-0000-000000000002" ]] || false
  [[ "${lines[2]}" == "00000000-0000-0000-0000-000000000005" ]] || false
  stop_doltgres
}
