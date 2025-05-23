#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server
}

teardown() {
    teardown_common
}

@test 'regression: correct column name displayed for dolt_ tables' {
    run query_server -c "select name from dolt_branches"
    [ "$status" -eq 0 ]
    [[ ! "$output" =~ "dolt_branches.name" ]] || false
}
