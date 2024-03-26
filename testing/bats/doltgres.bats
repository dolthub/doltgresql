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

@test 'doltgres: config file' {
    [ ! -d "doltgres" ]
    export DOLTGRES_DATA_DIR="$(pwd)"
    export SQL_USER="doltgres"

    PORT=$( definePORT )
    cat > config.yaml <<EOF
behavior:
  read_only: false
  autocommit: true
  persistence_behavior: load
  disable_client_multi_statements: false
  dolt_transaction_commit: false

user:
  name: "doltgres"
  password: "password"

listener:
  host: localhost
  port: $PORT
EOF

    cat config.yaml
    start_sql_server_with_args --config config.yaml > log.txt 2>&1
    
    run cat log.txt
    [[ ! "$output" =~ "Author identity unknown" ]] || false
    [ -d "doltgres" ]

    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"
    
    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false
}

@test 'doltgres: DOLTGRES_DATA_DIR set to current dir' {
    [ ! -d "doltgres" ]
    export DOLTGRES_DATA_DIR="$(pwd)"
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

    export DOLTGRES_DATA_DIR="$(pwd)"
    export SQL_USER="doltgres"
    start_sql_server_with_args "--host 0.0.0.0" "--user doltgres" "--data-dir=./test" #> log.txt 2>&1

    run cat log.txt
    [[ ! "$output" =~ "Author identity unknown" ]] || false
    [ ! -d "doltgres" ]
    [ -d "test/doltgres" ]

    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "information_schema" ]] || false
    [[ "$output" =~ "doltgres" ]] || false
}
