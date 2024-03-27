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

    # setup the postgres primary
    postgres_primary_query "drop table if exists t1"
    postgres_primary_query "create table t1 (a int primary key, b int)"
    postgres_primary_query "DROP PUBLICATION IF EXISTS doltgres_slot"
    postgres_primary_query "CREATE PUBLICATION doltgres_slot FOR TABLE t1"
    run postgres_primary_query "DROP_REPLICATION_SLOT doltgres_slot" # ignore errors if the slot doesn't exist
    postgres_primary_query "CREATE_REPLICATION_SLOT doltgres_slot LOGICAL pgoutput"
    
    # This host may have a history, and we don't want to start replicating from the beginning of
    # history, just from the current WAL position. So seed that state here.
    LSN=$(postgres_primary_query "SELECT pg_current_wal_lsn()" -t)

    if [[ ! -d ./.doltcfg ]]; then
        mkdir ./.doltcfg
    fi
    echo $LSN > ./.doltcfg/pg_wal_location 

    cat ./.doltcfg/pg_wal_location  
    
    cp "$BATS_TEST_DIRNAME/replication-config.yaml" "$BATS_TMPDIR/dolt-repo-$$" 
    PORT=5433
    start_sql_server_with_args "--config=replication-config.yaml"
        
    # Create the table that already exists on the primary before doing any inserts on the primary
    query_server doltgres -c "create table t1 (a int primary key, b int)"

    # this insert on the primary should now replicate to the replica
    postgres_primary_query "insert into t1 values (1, 2)"
    sleep 1

    query_server doltgres -c "select * from t1" -t
    run query_server doltgres -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    stop_sql_server
}

postgres_primary_query() {
    PGPASSWORD=password psql -U "postgres" -h 127.0.0.1 -p 5432 "dbname=postgres replication=database" -c "$@"
}
