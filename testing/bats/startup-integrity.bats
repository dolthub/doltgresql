#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

# These tests exercise the startup integrity check: the server scans every database for
# adaptive-encoded out-of-band values missing from their tree node's chunk reference index
# (value_address_offsets), a corruption written by earlier releases, and refuses to start against a
# corrupted database so that garbage collection cannot destroy the affected values.

setup() {
    setup_common
}

teardown() {
    teardown_common
}

# generate_corrupt_data_dir builds a data directory at $1 (absolute path) containing a database
# corrupted the way earlier releases wrote it, using the corruption generator in cmd/admin. Requires
# the Go toolchain and the doltgresql source tree, both of which the bats CI environment provides.
generate_corrupt_data_dir() {
    local dest=$1
    local repo_root
    repo_root=$(cd "$BATS_TEST_DIRNAME/../.." && pwd)
    (cd "$repo_root" && ADMIN_TEST_GEN_DIR="$dest" go test -count=1 -run TestGenerateCorruptedDataDir ./cmd/admin/)
}

write_config() {
    local port=$1
    local skip=$2
    cat > config.yaml <<EOF
log_level: warn

behavior:
  skip_startup_integrity_check: $skip

listener:
  host: localhost
  port: $port
EOF
}

@test "startup-integrity: server refuses to start against a corrupt database" {
    generate_corrupt_data_dir "$(pwd)/data"
    PORT=$( definePORT )
    write_config $PORT false

    run doltgres -data-dir=./data -config=config.yaml
    log_status_eq 1
    log_output_has "failed a startup integrity check"
    # the error must tell the user to back up before repairing
    log_output_has "MAKE A BACKUP COPY"
    log_output_has "doltgres-admin repair"
    log_output_has "skip_startup_integrity_check"
}

@test "startup-integrity: skip_startup_integrity_check starts against a corrupt database" {
    generate_corrupt_data_dir "$(pwd)/data"
    PORT=$( definePORT )
    write_config $PORT true

    start_sql_server_with_args "-data-dir=./data" "-config=config.yaml"

    run query_server_for_db corruption_test -c "SELECT count(*) FROM t2;" -t
    log_status_eq 0
    log_output_has "2"

    # a skipped check records no passed-check sentinel
    [ ! -f data/corruption_test/.dolt/.integrity_check_passed ]
}

@test "startup-integrity: healthy database passes the startup check" {
    PORT=$( definePORT )
    write_config $PORT false

    # an empty data dir (plus the default database the server creates) is healthy
    mkdir data
    start_sql_server_with_args "-data-dir=./data" "-config=config.yaml"

    run query_server -c "SELECT 1;" -t
    log_status_eq 0
}

@test "startup-integrity: repaired database passes the startup check" {
    generate_corrupt_data_dir "$(pwd)/data"
    local repo_root
    repo_root=$(cd "$BATS_TEST_DIRNAME/../.." && pwd)
    (cd "$repo_root" && go build -o "$OLDPWD/doltgres-admin" ./cmd/admin)
    ./doltgres-admin repair -dir ./data -out report.html

    PORT=$( definePORT )
    write_config $PORT false

    start_sql_server_with_args "-data-dir=./data" "-config=config.yaml"

    run query_server_for_db corruption_test -c "SELECT count(*) FROM t1 WHERE length(big) = 20000;" -t
    log_status_eq 0
    # 7 committed out-of-band values on main plus 2 repaired uncommitted ones in the working set
    log_output_has "9"

    # the passed check dropped a sentinel so future startups skip re-checking this database
    [ -f data/corruption_test/.dolt/.integrity_check_passed ]
}

@test "startup-integrity: sentinel from a passed check skips future checks" {
    generate_corrupt_data_dir "$(pwd)/data"
    PORT=$( definePORT )
    write_config $PORT false

    # the corrupt database refuses to start
    run doltgres -data-dir=./data -config=config.yaml
    log_status_eq 1
    log_output_has "failed a startup integrity check"

    # a sentinel recorded by a passed check suppresses the check for that database entirely, so
    # the server now starts even though the database is corrupt. (The file content is the version
    # of the check that passed; this couples the test to that format on purpose.)
    echo "1" > data/corruption_test/.dolt/.integrity_check_passed
    start_sql_server_with_args "-data-dir=./data" "-config=config.yaml"
    run query_server_for_db corruption_test -c "SELECT count(*) FROM t2;" -t
    log_status_eq 0
    stop_sql_server 1

    # a sentinel recorded by a different (older) version of the check does not suppress it
    echo "0" > data/corruption_test/.dolt/.integrity_check_passed
    PORT=$( definePORT )
    write_config $PORT false
    run doltgres -data-dir=./data -config=config.yaml
    log_status_eq 1
    log_output_has "failed a startup integrity check"
}
