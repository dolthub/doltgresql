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

# Function to extract and verify the first line (column name)
verify_column_name() {
  local output=$1
  local expected_column_name=$2

  # Extract the first line and trim leading and trailing whitespace
  local first_line=$(echo "$output" | head -n 1 | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

  # Verify the first line matches the expected column name
  [ "$first_line" = "$expected_column_name" ] || return 1
}

@test 'workbench-commands: subqueries' {
  run query_server -c "SELECT \"con\".\"conname\" AS \"constraint_name\", \"con\".\"nspname\" AS \"table_schema\", \"con\".\"relname\" AS \"table_name\", \"att2\".\"attname\" AS \"column_name\", \"ns\".\"nspname\" AS \"referenced_table_schema\", \"cl\".\"relname\" AS \"referenced_table_name\", \"att\".\"attname\" AS \"referenced_column_name\", \"con\".\"confdeltype\" AS \"on_delete\", \"con\".\"confupdtype\" AS \"on_update\", \"con\".\"condeferrable\" AS \"deferrable\", \"con\".\"condeferred\" AS \"deferred\" FROM ( SELECT UNNEST (\"con1\".\"conkey\") AS \"parent\", UNNEST (\"con1\".\"confkey\") AS \"child\", \"con1\".\"confrelid\", \"con1\".\"conrelid\", \"con1\".\"conname\", \"con1\".\"contype\", \"ns\".\"nspname\", \"cl\".\"relname\", \"con1\".\"condeferrable\", CASE WHEN \"con1\".\"condeferred\" THEN 'INITIALLY DEFERRED' ELSE 'INITIALLY IMMEDIATE' END as condeferred, CASE \"con1\".\"confdeltype\" WHEN 'a' THEN 'NO ACTION' WHEN 'r' THEN 'RESTRICT' WHEN 'c' THEN 'CASCADE' WHEN 'n' THEN 'SET NULL' WHEN 'd' THEN 'SET DEFAULT' END as \"confdeltype\", CASE \"con1\".\"confupdtype\" WHEN 'a' THEN 'NO ACTION' WHEN 'r' THEN 'RESTRICT' WHEN 'c' THEN 'CASCADE' WHEN 'n' THEN 'SET NULL' WHEN 'd' THEN 'SET DEFAULT' END as \"confupdtype\" FROM \"pg_class\" \"cl\" INNER JOIN \"pg_namespace\" \"ns\" ON \"cl\".\"relnamespace\" = \"ns\".\"oid\" INNER JOIN \"pg_constraint\" \"con1\" ON \"con1\".\"conrelid\" = \"cl\".\"oid\" WHERE \"con1\".\"contype\" = 'f' AND ((\"ns\".\"nspname\" = 'public' AND \"cl\".\"relname\" = 'testing')) ) \"con\" INNER JOIN \"pg_attribute\" \"att\" ON \"att\".\"attrelid\" = \"con\".\"confrelid\" AND \"att\".\"attnum\" = \"con\".\"child\" INNER JOIN \"pg_class\" \"cl\" ON \"cl\".\"oid\" = \"con\".\"confrelid\"  AND \"cl\".\"relispartition\" = 'f'INNER JOIN \"pg_namespace\" \"ns\" ON \"cl\".\"relnamespace\" = \"ns\".\"oid\" INNER JOIN \"pg_attribute\" \"att2\" ON \"att2\".\"attrelid\" = \"con\".\"conrelid\" AND \"att2\".\"attnum\" = \"con\".\"parent\";"
  [ "$status" -eq 0 ]

  run query_server --csv -c "SELECT columns.*, pg_catalog.col_description(('\"' || table_catalog || '\".\"' || table_schema || '\".\"' || table_name || '\"')::regclass::oid, ordinal_position) AS description, ('\"' || \"udt_schema\" || '\".\"' || \"udt_name\" || '\"')::\"regtype\" AS \"regtype\", pg_catalog.format_type(\"col_attr\".\"atttypid\", \"col_attr\".\"atttypmod\") AS \"format_type\" FROM \"information_schema\".\"columns\" LEFT JOIN \"pg_catalog\".\"pg_attribute\" AS \"col_attr\" ON \"col_attr\".\"attname\" = \"columns\".\"column_name\" AND \"col_attr\".\"attrelid\" = ( SELECT \"cls\".\"oid\" FROM \"pg_catalog\".\"pg_class\" AS \"cls\" LEFT JOIN \"pg_catalog\".\"pg_namespace\" AS \"ns\" ON \"ns\".\"oid\" = \"cls\".\"relnamespace\" WHERE \"cls\".\"relname\" = \"columns\".\"table_name\" AND \"ns\".\"nspname\" = \"columns\".\"table_schema\" ) WHERE (\"table_schema\" = 'public' AND \"table_name\" = 'test1');"
  [ "$status" -eq 0 ]
  [ "${#lines[@]}" -eq 3 ]
  [[ "$output" =~ "pk" ]] || false
  [[ "$output" =~ "v1" ]] || false
}

@test 'workbench-commands: version' {
  run query_server -c "SELECT version();"
  [ "$status" -eq 0 ]
  # Ensure the column name is 'version' and not 'version()'
  verify_column_name "$output" "version"
  [[ "$output" =~ "PostgreSQL 15.5" ]] || false
}

@test 'workbench-commands: current_schema' {
  run query_server -c "SELECT * FROM current_schema()"
  [ "$status" -eq 0 ]
  verify_column_name "$output" "current_schema"
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
  verify_column_name "$output" "current_database"
  [[ "$output" =~ "doltgres" ]] || false

  run query_server -c "CREATE DATABASE newdb;"
  [ "$status" -eq 0 ]

  run query_server_for_db newdb -c "SELECT * FROM current_database()"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "newdb" ]] || false
}

