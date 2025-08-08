#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server
}

teardown() {
    teardown_common
}

@test 'root-objects: dolt_add, dolt_branch, dolt_checkout, dolt_commit, dolt_reset' {
    query_server <<SQL
CREATE SEQUENCE test;
SELECT setval('test', 10);
SQL
    run query_server -c "SELECT nextval('test');"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "11" ]] || false

    query_server -c "SELECT dolt_add('test');"
    run query_server -c "SELECT length(dolt_commit('-m', 'initial')::text);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "34" ]] || false

    query_server -c "SELECT dolt_branch('other');"
    query_server -c "SELECT setval('test', 20);"
    query_server -c "SELECT dolt_add('.');"
    run query_server -c "SELECT length(dolt_commit('-m', 'next')::text);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "34" ]] || false

    run query_server -c "SELECT nextval('test');"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "21" ]] || false

    run query_server <<SQL
SELECT dolt_checkout('other');
SELECT nextval('test');
SQL
    [ "$status" -eq 0 ]
    [[ "$output" =~ "12" ]] || false
}

@test 'root-objects: start and stop' {
    query_server <<SQL
CREATE SEQUENCE test;
SELECT setval('test', 10);
SQL
    run query_server -c "SELECT nextval('test');"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "11" ]] || false

    stop_sql_server
    start_sql_server
    query_server -c "SELECT dolt_add('test');"
    run query_server -c "SELECT length(dolt_commit('-m', 'initial')::text);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "34" ]] || false

    stop_sql_server
    start_sql_server
    query_server -c "SELECT dolt_branch('other');"
    query_server -c "SELECT setval('test', 20);"
    query_server -c "SELECT dolt_add('.');"
    run query_server -c "SELECT length(dolt_commit('-m', 'next')::text);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "34" ]] || false

    stop_sql_server
    start_sql_server
    run query_server -c "SELECT nextval('test');"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "21" ]] || false

    stop_sql_server
    start_sql_server
    run query_server <<SQL
SELECT dolt_checkout('other');
SELECT nextval('test');
SQL
    [ "$status" -eq 0 ]
    [[ "$output" =~ "12" ]] || false
}

@test 'root-objects: \d does not break' {
    query_server <<SQL
CREATE TABLE "t" ("id" SERIAL);
SQL
    run query_server -c "\d"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "sequence" ]] || false

    stop_sql_server
    start_sql_server
    run query_server -c "\d"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "sequence" ]] || false
}
