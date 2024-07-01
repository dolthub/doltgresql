#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server
    query_server <<SQL
    CREATE TABLE test1 (pk BIGINT PRIMARY KEY, v1 SMALLINT);
    CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 INTEGER, v2 SMALLINT);
    INSERT INTO test1 VALUES (1, 2), (6, 7);
    INSERT INTO test2 VALUES (3, 4, 5), (8, 9, 0);
    CREATE VIEW testview AS SELECT * FROM test1;
SQL
}

teardown() {
    teardown_common
}

@test 'workbench-commands: current_schema' {
  run query_server -c "SELECT * FROM current_schema()"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "public" ]] || false

  run query_server <<SQL
  CREATE SCHEMA test_schema;
  SET search_path TO test_schema;
  SELECT * FROM current_schema();
SQL
  [ "$status" -eq 0 ]
  [[ "$output" =~ "test_schema" ]] || false
}

@test 'workbench-commands: current_database' {
  run query_server -c "SELECT * FROM current_database();"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "doltgres" ]] || false

  run query_server -c "CREATE DATABASE newdb;"
  [ "$status" -eq 0 ]

  run query_server_for_db newdb -c "SELECT * FROM current_database()"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "newdb" ]] || false
}