#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    stash_current_dolt_user
    unset_dolt_user
}

teardown() {
    restore_stashed_dolt_user
    teardown_common
}

@test 'doltgres: DOLTGRES_DATA_DIR set to current dir' {
    [ ! -d "doltgres" ]

    export SQL_USER="doltgres"
    start_sql_server_with_args "--host 0.0.0.0" "--user doltgres" > log.txt 2>&1

    run cat log.txt
    [[ ! "$output" =~ "Author identity unknown" ]] || false
    [ -d "doltgres" ]

    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "information_schema" ]] || false
    [[ "$output" =~ "doltgres" ]] || false
    [[ "$output" =~ "postgres" ]] || false
}

@test 'doltgres: setting both --data-dir and DOLTGRES_DATA_DIR should use --data-dir value' {
    [ ! -d "doltgres" ]

    export SQL_USER="doltgres"
    start_sql_server_with_args "--host 0.0.0.0" "--user doltgres" "--data-dir=./test" > log.txt 2>&1

    run cat log.txt
    [[ ! "$output" =~ "Author identity unknown" ]] || false
    [ ! -d "doltgres" ]
    [ -d "test/doltgres" ]

    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "information_schema" ]] || false
    [[ "$output" =~ "doltgres" ]] || false
}
