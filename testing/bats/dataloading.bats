#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server
}

teardown() {
    teardown_common
}

# Tests that we can successfully load the french towns dataset into Doltgres
# https://github.com/morenoh149/postgresDBSamples/blob/master/french-towns-communes-francaises/french-towns-communes-francaises.sql
# NOTE: This data dump still has two issues that need to be fixed in Doltgres, before it will load cleanly without
#       modifications:
#         TEXT columns are replaced with VARCHAR
#         ANALYZE statements are commented out
@test 'dataloading: french towns dataset' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/french-towns-communes-francaises.sql
  [ "$status" -eq 0 ]
  [[ "$output" =~ "COPY 26" ]] || false
  [[ "$output" =~ "COPY 100" ]] || false
  [[ "$output" =~ "COPY 36684" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false
  [[ ! "$output" =~ "is not yet supported" ]] || false

  # Check the row count of imported tables
  run query_server -c "SELECT count(*) from Regions;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "26" ]] || false
  run query_server -c "SELECT count(*) from Departments;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "100" ]] || false
  run query_server -c "SELECT count(*) from Towns;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "36684" ]] || false

  # Spot check a row from each table
  run query_server -c "SELECT * from Regions where id=21;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "21 | 74   | 87085   | Limousin" ]] || false
  run query_server -c "SELECT * from Departments where id=42;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "42 | 41   | 41018   | 24     | Loir-et-Cher" ]] || false
  run query_server -c "SELECT * from Towns where id=420;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "420 | 001  |         | Abbécourt | 02" ]] || false
}

# Tests that we can load data dump files with windows line endings.
@test 'dataloading: windows line endings' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/windows-line-endings.sql
  [ "$status" -eq 0 ]
  [[ "$output" =~ "COPY 26" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the row count of imported tables
  run query_server -c "SELECT count(*) from Regions;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "26" ]] || false

  # Spot check a row
  run query_server -c "SELECT * from Regions where id=21;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "21 | 74   | 87085   | Limousin" ]] || false
}