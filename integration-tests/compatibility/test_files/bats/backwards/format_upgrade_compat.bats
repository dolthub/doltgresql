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
