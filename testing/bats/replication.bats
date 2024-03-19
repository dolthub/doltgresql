#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    # tests are run without setting doltgres config user.name and user.email
    stash_current_dolt_user
    unset_dolt_user
}

teardown() {
    restore_stashed_dolt_user
    teardown_common
}

@test 'replication: test postgres connection' {
    if [[ ! -v "RUN_DOLTGRES_REPLICATION_TESTS" ]]; then
       skip "RUN_DOLTGRES_REPLICATION_TESTS not set, skipping"
    fi

    postgres_primary_query "drop table if exists t1"
    postgres_primary_query "create table t1 (a int primary key, b int)"
    postgres_primary_query "DROP PUBLICATION IF EXISTS doltgres_slot"
    postgres_primary_query "CREATE PUBLICATION doltgres_slot FOR ALL TABLES"

     cp "$BATS_TEST_DIRNAME/replication-config.yaml" "$BATS_TMPDIR/dolt-repo-$$" 
    start_sql_server_with_config_file "--host 0.0.0.0" "--config=replication-config.yaml" > log.txt 2>&1
    PORT=5433
    
    run cat log.txt
    [[ ! "$output" =~ "Author identity unknown" ]] || false
    [ -d "doltgres" ]

    # Create the table that already exists on the primaryu before doing any inserts on the primary
    query_server -c "create table t1 (a int primary key, b int)"
    
    postgres_primary_query "insert into t1 values (1, 2)"
    sleep 1

    query_server -c "select 'abc123'"
    query_server -c "select * from t1"

    cat log.txt
    run query_server -c "select * from t1"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    stop_sql_server
}

postgres_primary_query() {
    PGPASSWORD=password psql -U "postgres" -h 127.0.0.1 -p 5432 postgres -c "$@"
}

start_sql_server_with_config_file() {
    DEFAULT_DB=""
    nativevar DEFAULT_DB "$DEFAULT_DB" /w
    doltgresql "$@" &
    SERVER_PID=$!
    wait_for_connection 5433 3000
}
