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

@test 'doltgres: config file with all options' {
    PORT=$( definePORT )
    cat > config.yaml <<EOF
log_level: info

behavior:
  read_only: false
  autocommit: true
  persistence_behavior: load
  disable_client_multi_statements: false
  dolt_transaction_commit: false
  event_scheduler: "ON"

user:
  name: ""
  password: ""

listener:
  host: localhost
  port: $PORT
  max_connections: 100
  read_timeout_millis: 28800000
  write_timeout_millis: 28800000
  tls_key: null
  tls_cert: null
  require_secure_transport: null
  allow_cleartext_passwords: null

performance:
  query_parallelism: null

data_dir: .

cfg_dir: .doltcfg

metrics:
  labels: {}
  host: null
  port: -1

remotesapi: {}

privilege_file: .doltcfg/privileges.db

branch_control_file: .doltcfg/branch_control.db

user_session_vars: []

jwks: []
EOF

    cat config.yaml

    skip "Bad error: Variable 'max_connections' is a SESSION variable and can't be used with SET GLOBAL"
    
    start_sql_server_with_args --config config.yaml

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
