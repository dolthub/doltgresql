#!/usr/bin/env bats
# Types compatibility tests for doltgresql.
#
# Verifies that the current doltgres build can correctly read, write, and alter
# all_types rows that were originally written by an older doltgres version (or
# by the current version in a forward-compatibility run).
#
# Depends on the all_types table created by setup_repo.sh.

load $BATS_TEST_DIRNAME/helper/common.bash

BATS_REPO=""

setup() {
  BATS_REPO="$BATS_TMPDIR/types-compat-repo-$$-$RANDOM"
  copy_repo "$REPO_DIR" "$BATS_REPO"
  start_doltgres "${DOLTGRES_TEST_BIN:-doltgres}" "$BATS_REPO" "$BATS_REPO/server.log"
}

teardown() {
  stop_doltgres
  rm -rf "$BATS_REPO"
}

# ---------------------------------------------------------------------------
# Row counts
# ---------------------------------------------------------------------------

@test "types_compat: all_types has 3 rows" {
  run sql_csv -c "SELECT count(*) FROM all_types;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3" ]] || false
}

# ---------------------------------------------------------------------------
# Numeric types
# ---------------------------------------------------------------------------

@test "types_compat: smallint column readable" {
  run sql_csv -c "SELECT pk, c_smallint FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,100" ]] || false
}

@test "types_compat: negative smallint readable" {
  run sql_csv -c "SELECT pk, c_smallint FROM all_types WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2,-100" ]] || false
}

@test "types_compat: int and bigint columns readable" {
  run sql_csv -c "SELECT pk, c_int, c_bigint FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,2000000,9223372036854775807" ]] || false
}

@test "types_compat: negative int and bigint readable" {
  run sql_csv -c "SELECT pk, c_int, c_bigint FROM all_types WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2,-2000000,-9223372036854775807" ]] || false
}

@test "types_compat: real and double precision readable" {
  run sql_csv -c "SELECT pk, c_real, c_double FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,1.5,2.5" ]] || false
}

@test "types_compat: numeric column readable" {
  run sql_csv -c "SELECT pk, c_numeric FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,12345.67" ]] || false
}

@test "types_compat: negative numeric readable" {
  run sql_csv -c "SELECT pk, c_numeric FROM all_types WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2,-12345.67" ]] || false
}

# ---------------------------------------------------------------------------
# Text types
# ---------------------------------------------------------------------------

@test "types_compat: char column readable" {
  run sql_csv -c "SELECT pk, trim(c_char) FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "hello" ]] || false
}

@test "types_compat: varchar column readable" {
  run sql_csv -c "SELECT pk, c_varchar FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,hello world" ]] || false
}

@test "types_compat: text column readable" {
  run sql_csv -c "SELECT pk, c_text FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,text val" ]] || false
}

@test "types_compat: large text value (500 chars) readable" {
  run sql_csv -c "SELECT pk, length(c_text) FROM all_types WHERE pk=3;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3,500" ]] || false
}

# ---------------------------------------------------------------------------
# Temporal types
# ---------------------------------------------------------------------------

@test "types_compat: date column readable" {
  run sql_csv -c "SELECT pk, c_date FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,2024-01-15" ]] || false
}

@test "types_compat: time column readable" {
  skip "broken functionality"
  run sql_csv -c "SELECT pk, c_time FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "13:30:45" ]] || false
}

@test "types_compat: timestamp column readable" {
  run sql_csv -c "SELECT pk, c_timestamp FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2024-01-15" ]] || false
  [[ "$output" =~ "13:30:45" ]] || false
}

@test "types_compat: timestamptz not null for pk=1" {
  run sql_csv -c "SELECT pk, (c_timestamptz IS NOT NULL) FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "t" ]] || false
}

@test "types_compat: timestamptz is null for pk=2" {
  run sql_csv -c "SELECT pk, (c_timestamptz IS NULL) FROM all_types WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "t" ]] || false
}

# ---------------------------------------------------------------------------
# Boolean type
# ---------------------------------------------------------------------------

@test "types_compat: boolean column readable" {
  run sql_csv -c "SELECT pk, c_boolean FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,t" ]] || false
}

@test "types_compat: false boolean readable" {
  run sql_csv -c "SELECT pk, c_boolean FROM all_types WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2,f" ]] || false
}

# ---------------------------------------------------------------------------
# JSONB type
# ---------------------------------------------------------------------------

@test "types_compat: jsonb column readable" {
  run sql_csv -c "SELECT pk, c_jsonb->'k' FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "v" ]] || false
}

@test "types_compat: jsonb array readable" {
  skip "jsonb_array_length not implemented"
  run sql_csv -c "SELECT pk, jsonb_array_length(c_jsonb) FROM all_types WHERE pk=2;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3" ]] || false
}

# ---------------------------------------------------------------------------
# DML: insert and read back all column types
# ---------------------------------------------------------------------------

@test "types_compat: full-column-set insert round-trips correctly" {
    skip "spurious error message for numeric type"
  sql <<SQL
INSERT INTO all_types (pk, c_smallint, c_int, c_bigint,
    c_real, c_double, c_numeric,
    c_char, c_varchar, c_text, c_bytea,
    c_date, c_time, c_timestamp, c_timestamptz,
    c_boolean, c_jsonb)
  VALUES (50, 55, 555555, 5555555555,
    0.5, 0.25, 55.55,
    'round', 'round trip', 'rt text', '\x526F756E64',
    '2025-03-18', '09:00:00', '2025-03-18 09:00:00', '2025-03-18 09:00:00+00',
    true, '{"round":true}');
SQL
  run sql_csv -c "SELECT pk, c_smallint, c_int FROM all_types WHERE pk=50;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "50,55,555555" ]] || false

  run sql_csv -c "SELECT pk, c_varchar, c_text FROM all_types WHERE pk=50;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "50,round trip,rt text" ]] || false

  run sql_csv -c "SELECT pk, c_boolean FROM all_types WHERE pk=50;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "50,t" ]] || false
}

@test "types_compat: update all_types row written by older version" {
  sql -c "UPDATE all_types SET c_varchar='updated', c_text='updated text' WHERE pk=1;"
  run sql_csv -c "SELECT pk, c_varchar, c_text FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,updated,updated text" ]] || false
}

@test "types_compat: update text to large value" {
  sql -c "UPDATE all_types SET c_text=repeat('u', 2000) WHERE pk=1;"
  run sql_csv -c "SELECT pk, length(c_text) FROM all_types WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,2000" ]] || false
}

@test "types_compat: delete a row from old table" {
  sql -c "DELETE FROM all_types WHERE pk=2;"
  run sql_csv -c "SELECT count(*) FROM all_types;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2" ]] || false
}

# ---------------------------------------------------------------------------
# ALTER TABLE: add columns of various types to a table written by older version
# ---------------------------------------------------------------------------

@test "types_compat: add text column to abc and use dml" {
  sql -c "ALTER TABLE abc ADD COLUMN new_text TEXT;"
  sql -c "UPDATE abc SET new_text='text for row';"
  sql -c "INSERT INTO abc (pk, a, b, x, y, new_text) VALUES (99, 'test', 1.0, 0, 0, 'inserted');"

  run sql_csv -c "SELECT pk, new_text FROM abc WHERE pk=99;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "99,inserted" ]] || false

  sql -c "DELETE FROM abc WHERE pk=99;"
}

@test "types_compat: add integer columns to abc and use dml" {
  sql -c "ALTER TABLE abc ADD COLUMN new_smallint SMALLINT, ADD COLUMN new_bigint BIGINT;"
  sql -c "UPDATE abc SET new_smallint=42, new_bigint=9999999999;"
  sql -c "INSERT INTO abc (pk, a, b, x, y, new_smallint, new_bigint) VALUES (99, 'test', 1.0, 0, 0, -1, -9999999999);"

  run sql_csv -c "SELECT pk, new_smallint, new_bigint FROM abc WHERE pk=99;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "99,-1,-9999999999" ]] || false

  sql -c "DELETE FROM abc WHERE pk=99;"
}

@test "types_compat: add numeric column to abc and use dml" {
  sql -c "ALTER TABLE abc ADD COLUMN new_numeric NUMERIC(12,4);"
  sql -c "UPDATE abc SET new_numeric=9876.5432;"
  sql -c "INSERT INTO abc (pk, a, b, x, y, new_numeric) VALUES (99, 'test', 1.0, 0, 0, -9876.5432);"

  run sql_csv -c "SELECT pk, new_numeric FROM abc WHERE pk=99;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "99,-9876.5432" ]] || false

  sql -c "DELETE FROM abc WHERE pk=99;"
}

@test "types_compat: add date/timestamp columns to abc and use dml" {
  sql -c "ALTER TABLE abc ADD COLUMN new_date DATE, ADD COLUMN new_ts TIMESTAMP;"
  sql -c "UPDATE abc SET new_date='2025-03-18', new_ts='2025-03-18 10:00:00';"
  sql -c "INSERT INTO abc (pk, a, b, x, y, new_date, new_ts) VALUES (99, 'test', 1.0, 0, 0, '2025-06-01', '2025-06-01 12:00:00');"

  run sql_csv -c "SELECT pk, new_date FROM abc WHERE pk=99;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "99,2025-06-01" ]] || false

  sql -c "DELETE FROM abc WHERE pk=99;"
}

@test "types_compat: add boolean column to abc and use dml" {
  sql -c "ALTER TABLE abc ADD COLUMN new_bool BOOLEAN;"
  sql -c "UPDATE abc SET new_bool=true;"
  sql -c "INSERT INTO abc (pk, a, b, x, y, new_bool) VALUES (99, 'test', 1.0, 0, 0, false);"

  run sql_csv -c "SELECT pk, new_bool FROM abc WHERE pk=99;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "99,f" ]] || false

  sql -c "DELETE FROM abc WHERE pk=99;"
}

@test "types_compat: add jsonb column to abc and use dml" {
  sql -c "ALTER TABLE abc ADD COLUMN new_jsonb JSONB;"
  sql -c "UPDATE abc SET new_jsonb='{\"updated\":true}';"
  sql -c "INSERT INTO abc (pk, a, b, x, y, new_jsonb) VALUES (99, 'test', 1.0, 0, 0, '{\"inserted\":1}');"

  run sql_csv -c "SELECT pk, new_jsonb->>'inserted' FROM abc WHERE pk=99;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1" ]] || false

  sql -c "DELETE FROM abc WHERE pk=99;"
}

@test "types_compat: add bytea column to abc and use dml" {
  skip "encode not implemented"
  sql -c "ALTER TABLE abc ADD COLUMN new_bytea BYTEA;"
  sql -c "UPDATE abc SET new_bytea='\xDEAD';"
  sql -c "INSERT INTO abc (pk, a, b, x, y, new_bytea) VALUES (99, 'test', 1.0, 0, 0, '\xBEEF');"

  run sql_csv -c "SELECT pk, encode(new_bytea, 'hex') FROM abc WHERE pk=99;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "beef" ]] || false

  sql -c "DELETE FROM abc WHERE pk=99;"
}

# ---------------------------------------------------------------------------
# Schema inspection
# ---------------------------------------------------------------------------

@test "types_compat: all_types column types visible in information_schema" {
  run sql -c "SELECT column_name, data_type FROM information_schema.columns WHERE table_name='all_types' ORDER BY ordinal_position;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "smallint" ]]          || false
  [[ "$output" =~ "integer" ]]           || false
  [[ "$output" =~ "bigint" ]]            || false
  [[ "$output" =~ "real" ]]              || false
  [[ "$output" =~ "double precision" ]]  || false
  [[ "$output" =~ "numeric" ]]           || false
  [[ "$output" =~ "character" ]]         || false
  [[ "$output" =~ "text" ]]              || false
  [[ "$output" =~ "bytea" ]]             || false
  [[ "$output" =~ "date" ]]              || false
  [[ "$output" =~ "timestamp" ]]         || false
  [[ "$output" =~ "boolean" ]]           || false
  [[ "$output" =~ "jsonb" ]]             || false
}

# ---------------------------------------------------------------------------
# Commit works after DML on old table
# ---------------------------------------------------------------------------

@test "types_compat: commit works after dml on all_types" {
  sql -c "INSERT INTO all_types (pk, c_text) VALUES (98, 'commit test');"
  sql -c "SELECT dolt_add('.');"
  run sql -c "SELECT length(dolt_commit('-m', 'types compat commit')::text);"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "34" ]] || false
}

# ---------------------------------------------------------------------------
# View over all_types
# ---------------------------------------------------------------------------

@test "types_compat: all_types_view returns same rows as underlying table" {
  run sql_csv -c "SELECT pk, c_smallint, c_varchar FROM all_types_view WHERE pk=1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1,100,hello world" ]] || false
}

@test "types_compat: all_types_view supports filtering" {
  run sql_csv -c "SELECT count(*) FROM all_types_view WHERE pk < 3;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2" ]] || false
}
