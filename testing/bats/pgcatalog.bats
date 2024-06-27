#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server
    query_server <<SQL
    CREATE TABLE test1 (pk BIGINT PRIMARY KEY, v1 SMALLINT);
    INSERT INTO test1 VALUES (1, 2), (6, 7);
SQL
}

teardown() {
    teardown_common
}


@test 'pgcatalog: tables do not include data from other databases' {
  run query_server --csv -c "SELECT current_database();"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "current_database" ]] || false
  [[ "$output" =~ "doltgres" ]] || false
  [ "${#lines[@]}" -eq 2 ]

  run query_server --csv -c "SELECT attname FROM pg_catalog.pg_attribute WHERE attname = 'pk';"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "attname" ]] || false
  [[ "$output" =~ "pk" ]] || false
  [ "${#lines[@]}" -eq 2 ]

    run query_server --csv -c "SELECT relname FROM pg_catalog.pg_class WHERE relname = 'test1';"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "relname" ]] || false
  [[ "$output" =~ "test1" ]] || false
  [ "${#lines[@]}" -eq 2 ]

  run query_server -c "CREATE DATABASE newdb;"
  [ "$status" -eq 0 ]

  run query_server_for_db newdb --csv -c "SELECT current_database();"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "current_database" ]] || false
  [[ "$output" =~ "newdb" ]] || false
  [ "${#lines[@]}" -eq 2 ]

  run query_server_for_db newdb --csv -c "SELECT attname FROM pg_catalog.pg_attribute WHERE attname = 'pk';"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "attname" ]] || false
  [ "${#lines[@]}" -eq 1 ]

    run query_server_for_db newdb --csv -c "SELECT relname FROM pg_catalog.pg_class WHERE relname = 'test1';"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "relname" ]] || false
  [ "${#lines[@]}" -eq 1 ]
}
