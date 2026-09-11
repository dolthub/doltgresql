#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server
}

teardown() {
    teardown_common
}

@test 'auth: CREATE DATABASE requires SUPERUSER or CREATEDB' {
    query_server -c "CREATE ROLE demo LOGIN PASSWORD 'password'"
    query_server -c "GRANT ALL PRIVILEGES ON DATABASE postgres TO demo"

    run query_server -At -c "SELECT rolsuper, rolcreatedb, rolcreaterole FROM pg_roles WHERE rolname = 'demo'"
    [ "$status" -eq 0 ]
    [ "$output" = "f|f|f" ]

    SQL_USER=demo run query_server -c "CREATE DATABASE made_by_demo"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "permission denied to create database" ]] || false

    run query_server -At -c "SELECT datname FROM pg_database WHERE datname = 'made_by_demo'"
    [ "$status" -eq 0 ]
    [ -z "$output" ]

    query_server -c "ALTER ROLE demo CREATEDB"
    SQL_USER=demo query_server -c "CREATE DATABASE made_by_demo"
    run query_server -At -c "SELECT datname FROM pg_database WHERE datname = 'made_by_demo'"
    [ "$status" -eq 0 ]
    [ "$output" = "made_by_demo" ]
    query_server -c "DROP DATABASE made_by_demo"
}

@test 'auth: DROP DATABASE requires SUPERUSER' {
    query_server -c "CREATE DATABASE victim"
    query_server -c "CREATE ROLE demo LOGIN PASSWORD 'password'"
    query_server -c "GRANT ALL PRIVILEGES ON DATABASE victim TO demo"

    SQL_USER=demo run query_server -c "DROP DATABASE victim"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "must be owner of database victim" ]] || false

    query_server -c "ALTER ROLE demo CREATEDB"
    SQL_USER=demo run query_server -c "DROP DATABASE IF EXISTS victim"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "must be owner of database victim" ]] || false

    run query_server -At -c "SELECT datname FROM pg_database WHERE datname = 'victim'"
    [ "$status" -eq 0 ]
    [ "$output" = "victim" ]

    query_server -c "DROP DATABASE victim"
    run query_server -At -c "SELECT datname FROM pg_database WHERE datname = 'victim'"
    [ "$status" -eq 0 ]
    [ -z "$output" ]
}
