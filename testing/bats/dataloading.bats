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
# NOTE: This data dump still has one issue that needs to be fixed in Doltgres, before it will load cleanly without
#       modifications:
#         TEXT columns are replaced with VARCHAR because unique TEXT indexes don't work properly yet
@test 'dataloading: tabular import, french towns dataset' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/french-towns-communes-francaises.sql
  [ "$status" -eq 0 ]
  echo "OUTPUT: $output"
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
@test 'dataloading: tabular import, windows line endings' {
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

# Tests that we can load tabular data dump files that contain a header
@test 'dataloading: tabular import, with header' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/tab-load-with-header.sql
  [ "$status" -eq 0 ]
  [[ "$output" =~ "COPY 3" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the row count of imported tables
  run query_server -c "SELECT count(*) from Regions;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "3" ]] || false

  # Check the inserted rows
  run query_server -c "SELECT * from Regions;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1 | 01   | 97105   | Guadeloupe" ]] || false
  [[ "$output" =~ "2 | 02   | 97209   | Martinique" ]] || false
  [[ "$output" =~ "3 | 03   | 97302   | Guyane" ]] || false
}

# Tests that we can load tabular data dump files that do not explicitly manage the session's transaction.
@test 'dataloading: tabular import, no explicit tx management' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/tab-load-with-no-tx-control.sql
  [ "$status" -eq 0 ]
  [[ "$output" =~ "COPY 3" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the inserted rows
  run query_server -c "SELECT * FROM test_info ORDER BY id;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4 | string for 4 |       1" ]] || false
  [[ "$output" =~ "5 | string for 5 |       0" ]] || false
  [[ "$output" =~ "6 | string for 6 |       0" ]] || false
}

# Tests that we can load tabular data dump files that specify a delimiter.
@test "dataloading: tabular import, delimiter='|', no explicit tx management" {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/tab-load-with-delimiter-no-tx-control.sql
  [ "$status" -eq 0 ]
  echo "OUTPUT: $output"
  [[ "$output" =~ "COPY 3" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the inserted rows
  run query_server -c "SELECT * FROM test_info ORDER BY id;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4 | string for 4 |       1" ]] || false
  [[ "$output" =~ "5 | string for 5 |       0" ]] || false
  [[ "$output" =~ "6 | string for 6 |       0" ]] || false
}


# Tests loading in data via different CSV data files.
@test 'dataloading: csv import' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/csv-load-basic-cases.sql
  [ "$status" -eq 0 ]
  [[ "$output" =~ "COPY 9" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the row count of imported tables
  run query_server -c "SELECT count(*) from tbl1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "9" ]] || false

  # Assert the data was loaded correctly
  run query_server -c "SELECT * from tbl1 order by pk;"
  [ "$status" -eq 0 ]
  [ "${#lines[@]}" -eq 17 ]
  [[ "$output" =~ "1 | green | " ]] || false
  [[ "$output" =~ "2 | blue  | a   +" ]] || false
  [[ "$output" =~ "  |       | q   +" ]] || false
  [[ "$output" =~ "  |       | u   +" ]] || false
  [[ "$output" =~ "  |       | a" ]] || false
  [[ "$output" =~ "3 | brown |" ]] || false
  [[ "$output" =~ "4 | NULL  | NULL" ]] || false
  [[ "$output" =~ "5 | ?     |" ]] || false
  [[ "$output" =~ "6 | foo  +| baz" ]] || false
  # NOTE: \. has to be escaped as \\\\.
  [[ "$output" =~ "  | \\\\.  +|" ]] || false
  [[ "$output" =~ "  | bar   |" ]] || false
  [[ "$output" =~ "7 |       | ' '" ]] || false
  [[ "$output" =~ "8 |       |" ]] || false
  [[ "$output" =~ "9 |       | ''" ]] || false

  # Assert NULL values were properly identified
  run query_server -c "SELECT * from tbl1 where c2 is NULL;"
  [[ "$output" =~ " 1 | green | " ]] || false
  [[ "$output" =~ " 3 | brown | " ]] || false
  run query_server -c "SELECT * from tbl1 where c1 is NULL;"
  [ "${#lines[@]}" -eq 4 ]
  [[ "$output" =~ " 9 |    | ''" ]] || false
}

# Tests loading in CSV data that includes a header row.
@test 'dataloading: csv import with header' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/csv-load-with-header.sql
  [ "$status" -eq 0 ]
  echo "OUTPUT: $output"
  [[ "$output" =~ "COPY 9" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the row count of imported tables
  run query_server -c "SELECT count(*) from tbl1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "9" ]] || false
}

# Tests loading in data via a CSV data file that is large enough to be split across multiple chunks.
@test 'dataloading: csv import across multiple chunks' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/csv-load-multi-chunk.sql
  [ "$status" -eq 0 ]
  [[ "$output" =~ "COPY 100" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the row count of imported tables
  run query_server -c "SELECT count(*) from tbl1;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "100" ]] || false
}

# Tests that we can load CSV data dump files that do not explicitly manage the session's transaction.
@test 'dataloading: csv import, no explicit tx management' {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/csv-load-with-no-tx-control.sql
  [ "$status" -eq 0 ]
  [[ "$output" =~ "COPY 3" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the inserted rows
  run query_server -c "SELECT * FROM test_info ORDER BY id;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4 | string for 4 |       1" ]] || false
  [[ "$output" =~ "5 | string for 5 |       0" ]] || false
  [[ "$output" =~ "6 | string for 6 |       0" ]] || false
}

# Tests that we can load CSV data dump files that do not explicitly manage the session's transaction.
@test "dataloading: csv import, delimiter='|', no explicit tx management" {
  # Import the data dump and assert the expected output
  run query_server -f $BATS_TEST_DIRNAME/dataloading/psv-load-with-no-tx-control.sql
  [ "$status" -eq 0 ]
  [[ "$output" =~ "COPY 3" ]] || false
  [[ ! "$output" =~ "ERROR" ]] || false

  # Check the inserted rows
  run query_server -c "SELECT * FROM test_info ORDER BY id;"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4 | string for 4 |       1" ]] || false
  [[ "$output" =~ "5 | string for 5 |       0" ]] || false
  [[ "$output" =~ "6 | string for 6 |       0" ]] || false
}
