#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
}

teardown() {
    teardown_common
}

@test 'foreign-keys: survives restarts' {
    PORT=$( definePORT )

    # stopping the server undefines the port, so save it
    port=$PORT
    mkdir test-home

    CONFIG=$( defineCONFIG $PORT )
    echo "$CONFIG" > config.yaml
    
    doltgres > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    cat server.out
    echo "$output"
    [ "$status" -eq 0 ]
    
    query_server -c "create table parent (a int primary key, b int)"
    query_server -c "insert into parent values (1,2)"
    query_server -c "create table child (c int primary key, d int, foreign key (d) references public.parent(a))"
    query_server -c "insert into child values (2,1)"

    stop_sql_server

    PORT=$port
    doltgres > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    run query_server -c "insert into child values (100,100)"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "violation" ]] || false
}
