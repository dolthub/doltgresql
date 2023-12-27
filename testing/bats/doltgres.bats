#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
}

teardown() {
    teardown_common
}

@test 'doltgres: no args' {
    start_sql_server_with_args "--host 0.0.0.0"
    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "information_schema" ]] || false
    [[ "$output" =~ "doltgres" ]] || false
    [[ "$output" =~ "postgres" ]] || false

    [ ! -d "doltgres" ]
}

@test 'doltgres: with --data-dir' {
    start_sql_server_with_args "--host 0.0.0.0" "--data-dir=."
    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "information_schema" ]] || false
    [[ "$output" =~ "doltgres" ]] || false
    [[ "$output" =~ "postgres" ]] || false

    [ -d "doltgres" ]
}

@test 'doltgres: with DOLTGRES_DATA_DIR' {
    DOLTGRES_DATA_DIR="$BATS_TEST_DIRNAME/test"
    start_sql_server_with_args "--host 0.0.0.0"
    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "information_schema" ]] || false
    [[ "$output" =~ "doltgres" ]] || false
    [[ "$output" =~ "postgres" ]] || false

    [ -d "test/doltgres" ]
    [ ! -d "doltgres" ]
}

@test 'doltgres: with both --data-dir and DOLTGRES_DATA_DIR' {
    DOLTGRES_DATA_DIR="$BATS_TEST_DIRNAME/test1"
    start_sql_server_with_args "--host 0.0.0.0" "--data-dir=./test2"
    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "information_schema" ]] || false
    [[ "$output" =~ "doltgres" ]] || false
    [[ "$output" =~ "postgres" ]] || false

    [ -d "test2/doltgres" ]
    [ ! -d "test1/doltgres" ]
}
