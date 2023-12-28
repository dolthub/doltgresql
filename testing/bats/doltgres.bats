#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    # dolt config was being used to create a new doltgres database
    # so unset user name and email for testing
    dolt config --global --unset user.name
    dolt config --global --unset user.email
}

teardown() {
    dolt config --global --add user.name "Bats Tests"
    dolt config --global --add user.email "bats@email.fake"
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
