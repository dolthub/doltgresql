#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
}

teardown() {
    teardown_common
}

@test 'doltgres: --help' {
    # just a smoke test
    doltgres --help
}

@test 'doltgres: --config-help' {
    # just a smoke test
    doltgres --config-help
}

@test 'doltgres: no arguments' {
    PORT=5432
    mkdir test-home
    # TODO: DOLT_ROOT_PATH behavior overrides the HOME behavior, which is confusing and not
    # applicable to Doltgres, fix it
    HOME=test-home DOLTGRES_DATA_DIR='' DOLT_ROOT_PATH='' doltgres > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    cat server.out
    echo "$output"
    [ "$status" -eq 0 ]
    
    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"

    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    # databases should get created in home/doltgres/databases by default
    [ -d test-home/doltgres/databases/postgres ]
}

@test 'doltgres: data-dir param' {
    PORT=5432
    DOLTGRES_DATA_DIR=fake doltgres --data-dir test > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    cat server.out
    echo "$output"
    [ "$status" -eq 0 ]
    
    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"

    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    [ -d test/postgres ]
}

@test 'doltgres: data dir in env var' {
    PORT=5432
    DOLTGRES_DATA_DIR=test doltgres > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    cat server.out
    echo "$output"
    [ "$status" -eq 0 ]
    
    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"

    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    [ -d test/postgres ]
}

@test 'doltgres: implicit config.yaml' {
    PORT=5434

    cat > config.yaml <<EOF
log_level: info

behavior:
  read_only: false
  disable_client_multi_statements: false
  dolt_transaction_commit: false

user:
  name: "postgres"
  password: "password"

listener:
  host: localhost
  port: $PORT
  read_timeout_millis: 28800000
  write_timeout_millis: 28800000

data_dir: test
EOF

    doltgres > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    cat server.out
    echo "$output"
    [ "$status" -eq 0 ]
    
    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"

    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    [ -d test/postgres ]
}

@test 'doltgres: config.yaml without data dir' {
    PORT=5434

    cat > config.yaml <<EOF
log_level: info

behavior:
  read_only: false
  disable_client_multi_statements: false
  dolt_transaction_commit: false

user:
  name: "postgres"
  password: "password"

listener:
  host: localhost
  port: $PORT
  read_timeout_millis: 28800000
  write_timeout_millis: 28800000

EOF

    mkdir test-home
    HOME=test-home DOLTGRES_DATA_DIR='' DOLT_ROOT_PATH='' doltgres --config config.yaml > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    cat server.out
    echo "$output"
    [ "$status" -eq 0 ]
    
    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"

    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    [ -d test-home/doltgres/databases/postgres ]
}

@test 'doltgres: config file override with explicit config.yaml' {
    PORT=5434

    cat > config-test.yaml <<EOF
log_level: info

behavior:
  read_only: false
  disable_client_multi_statements: false
  dolt_transaction_commit: false

user:
  name: "postgres"
  password: "password"

listener:
  host: localhost
  port: $PORT
  read_timeout_millis: 28800000
  write_timeout_millis: 28800000

data_dir: test
EOF

    # The only supported override right now is the data dir, add more here as we add more overrides
    doltgres --config config-test.yaml --data-dir local-override > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    cat server.out
    echo "$output"
    [ "$status" -eq 0 ]
    
    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"

    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    [ ! -d test/postgres ]
    [ -d local-override/postgres ]
}

@test 'doltgres: config file override with implicit config.yaml' {
    PORT=5434

    cat > config.yaml <<EOF
log_level: info

behavior:
  read_only: false
  disable_client_multi_statements: false
  dolt_transaction_commit: false

user:
  name: "postgres"
  password: "password"

listener:
  host: localhost
  port: $PORT
  read_timeout_millis: 28800000
  write_timeout_millis: 28800000

data_dir: test
EOF

    # The only supported override right now is the data dir, add more here as we add more overrides
    doltgres --data-dir local-override > server.out 2>&1 &
    SERVER_PID=$!
    run wait_for_connection $PORT 7500

    cat server.out
    echo "$output"
    [ "$status" -eq 0 ]
    
    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"

    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false

    [ ! -d test/postgres ]
    [ -d local-override/postgres ]
}

@test 'doltgres: config file' {
    PORT=$( definePORT )
    CONFIG=$( defineCONFIG $PORT )
    echo "$CONFIG" > config.yaml

    cat config.yaml
    start_sql_server_with_args -config config.yaml > log.txt 2>&1
    
    run cat log.txt
    [[ ! "$output" =~ "Author identity unknown" ]] || false
    [ -d "postgres" ]

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
  disable_client_multi_statements: false
  dolt_transaction_commit: false

user:
  name: "postgres"
  password: "password"

listener:
  host: localhost
  port: $PORT
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

    start_sql_server_with_args -config config.yaml

    query_server -c "create table t1 (a int primary key, b int)"
    query_server -c "insert into t1 values (1,2)"
    
    run query_server -c "select * from t1" -t
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | 2" ]] || false
}

@test 'doltgres: DOLTGRES_DATA_DIR set to current dir' {
    [ ! -d "postgres" ]
    export DOLTGRES_DATA_DIR="$(pwd)"
    start_sql_server > log.txt 2>&1

    run cat log.txt
    [[ ! "$output" =~ "Author identity unknown" ]] || false
    [ -d "postgres" ]

    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "postgres" ]] || false
}
