#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    # Disable the auto-GC load-average throttling so GC runs immediately instead of
    # backing off under CI load (see loadAvgGCScheduler in dolt/go's auto_gc.go).
    export DOLT_GC_SCHEDULER=NONE
}

teardown() {
    teardown_common
}

# make_gen_vals_sql writes a series of INSERT statements to $1, structured as a set of
# separate statements (rather than one large one) so that each is its own auto-committed
# transaction, giving the auto-GC commit hook multiple chances to fire during the import.
make_gen_vals_sql() {
    local out=$1
    echo "DROP TABLE IF EXISTS vals;" > "$out"
    echo "CREATE TABLE vals (c1 int, c2 int, c3 int, c4 int);" >> "$out"
    for i in $(seq 1 256); do
        echo "INSERT INTO vals SELECT (random()*65536)::int, (random()*65536)::int, (random()*65536)::int, (random()*65536)::int FROM generate_series(1, 1024);" >> "$out"
    done
}

@test "sql-auto-gc: importing data through a running server runs auto gc" {
    PORT=$( definePORT )
    cat > config.yaml <<EOF
log_level: debug

behavior:
  auto_gc_behavior:
    enable: true

listener:
  host: localhost
  port: $PORT
EOF
    start_sql_server_with_args "-config=config.yaml"

    make_gen_vals_sql gen_vals.sql
    query_server -f gen_vals.sql

    # auto GC runs asynchronously against the live server (unlike Dolt's one-shot
    # batch `dolt sql`, which blocks until GC finishes before the process exits), so
    # poll until the store size settles instead of asserting immediately after import.
    before=-1
    after=$(du -sk postgres/.dolt/noms | awk '{print $1}')
    for i in $(seq 1 30); do
        if [ "$before" -eq "$after" ]; then
            break
        fi
        before=$after
        sleep 1
        after=$(du -sk postgres/.dolt/noms | awk '{print $1}')
    done

    [[ "$after" -lt 524288 ]] || (echo "postgres/.dolt/noms should be less than 512MB after auto GC" && false)

    tablefiles=$(ls -1 postgres/.dolt/noms/ | egrep '[0-9a-v]{32}' | egrep -v 'v{32}')
    [[ $(echo "$tablefiles" | wc -l) -eq 1 ]] || (echo "postgres/.dolt/noms should have one table file after auto GC" && false)

    stop_sql_server
}

@test "sql-auto-gc: auto gc runs by default" {
    PORT=$( definePORT )
    cat > config.yaml <<EOF
log_level: debug

listener:
  host: localhost
  port: $PORT
EOF
    start_sql_server_with_args "-config=config.yaml" > log.txt 2>&1

    make_gen_vals_sql gen_vals.sql
    query_server -f gen_vals.sql

    # auto GC runs asynchronously against the live server, so poll until the store
    # size settles instead of asserting immediately after import.
    before=-1
    after=$(du -sk postgres/.dolt/noms | awk '{print $1}')
    for i in $(seq 1 30); do
        if [ "$before" -eq "$after" ]; then
            break
        fi
        before=$after
        sleep 1
        after=$(du -sk postgres/.dolt/noms | awk '{print $1}')
    done

    run grep -c "Successfully completed auto GC" log.txt
    [[ "$output" -gt 0 ]] || (echo "auto GC should run by default when auto_gc_behavior is not configured" && false)

    [[ "$after" -lt 524288 ]] || (echo "postgres/.dolt/noms should be less than 512MB after auto GC" && false)

    tablefiles=$(ls -1 postgres/.dolt/noms/ | egrep '[0-9a-v]{32}' | egrep -v 'v{32}')
    [[ $(echo "$tablefiles" | wc -l) -eq 1 ]] || (echo "postgres/.dolt/noms should have one table file after auto GC" && false)

    stop_sql_server
}

@test "sql-auto-gc: auto gc can be disabled" {
    PORT=$( definePORT )
    cat > config.yaml <<EOF
log_level: debug

behavior:
  auto_gc_behavior:
    enable: false

listener:
  host: localhost
  port: $PORT
EOF
    start_sql_server_with_args "-config=config.yaml" > log.txt 2>&1

    make_gen_vals_sql gen_vals.sql
    query_server -f gen_vals.sql
    sleep 3

    run grep -c "Successfully completed auto GC" log.txt
    [[ "$output" -eq 0 ]] || (echo "auto GC should not have run when disabled" && false)

    stop_sql_server
}
