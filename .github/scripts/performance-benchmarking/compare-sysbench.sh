#!/bin/bash
# Runs DoltHub's published sysbench tests for Doltgres (its Lua scripts plus sysbench's oltp tests, uniform keys,
# prepared statements off, one thread) on a fresh server for each test, once with the main branch's server and once
# with the pull request's, and prints a markdown table of their average latencies and transactions per second. Given
# the bin directory of a Postgres 15 install, it also runs each test on a fresh Postgres database for comparison.
#   compare-sysbench.sh <main binary> <pull request binary> <lua scripts dir> <seconds per test> [postgres bin dir]
set -uo pipefail

main_bin=$1
pr_bin=$2
lua=$3
secs=$4
pg_bin=${5:-}
tests="oltp_read_only oltp_point_select select_random_points select_random_ranges covering_index_scan_postgres
index_scan_postgres table_scan_postgres groupby_scan_postgres index_join_scan_postgres types_table_scan_postgres
index_join_postgres oltp_read_write oltp_update_index oltp_update_non_index oltp_insert oltp_write_only
oltp_delete_insert_postgres types_delete_insert_postgres"
port=54397
pg_port=54398
work=$(mktemp -d)

# run_sysbench prepares and runs one sysbench test against the sbtest database of the server on a port, and prints its
# average latency and transactions per second.
run_sysbench() {
  local port=$1 test=$2
  local common="--db-driver=pgsql --pgsql-host=127.0.0.1 --pgsql-port=$port --pgsql-user=postgres"
  common="$common --pgsql-password=password --pgsql-db=sbtest --table-size=10000 --rand-type=uniform --db-ps-mode=disable"
  local script=$test args="$common --tables=1"
  if [ -f "$lua/$test.lua" ]; then
    script="$lua/$test.lua"
    args=$common
  fi
  (cd "$lua" && sysbench "$script" $args prepare > /dev/null 2>&1)
  local out
  out=$(cd "$lua" && sysbench "$script" $args --time="$secs" --threads=1 run 2>&1)
  local avg tps
  avg=$(echo "$out" | grep "avg:" | awk '{print $2}')
  tps=$(echo "$out" | grep "transactions:" | sed -E 's/.*\(([0-9.]+) per sec.*/\1/')
  echo "${avg:-error} ${tps:-error}"
}

# run_test runs one sysbench test against a fresh server and prints its average latency and transactions per second.
run_test() {
  local bin=$1 test=$2
  rm -rf "$work/data" && mkdir -p "$work/data"
  printf 'log_level: warn\nlistener:\n  host: 127.0.0.1\n  port: %s\n' $port > "$work/config.yaml"
  (cd "$work" && exec "$bin" --config config.yaml --data-dir data > server.log 2>&1) &
  local pid=$!
  for _ in $(seq 1 150); do
    PGPASSWORD=password psql -h 127.0.0.1 -p $port -U postgres -c 'SELECT 1' postgres > /dev/null 2>&1 && break
    sleep 0.2
  done
  PGPASSWORD=password psql -h 127.0.0.1 -p $port -U postgres -c 'CREATE DATABASE sbtest' postgres > /dev/null 2>&1
  run_sysbench $port "$test"
  kill $pid && wait $pid 2> /dev/null
}

# run_postgres_test runs one sysbench test against a fresh database of the Postgres server and prints its tps.
run_postgres_test() {
  local test=$1
  psql -h 127.0.0.1 -p $pg_port -U postgres -c 'DROP DATABASE IF EXISTS sbtest' postgres > /dev/null 2>&1
  psql -h 127.0.0.1 -p $pg_port -U postgres -c 'CREATE DATABASE sbtest' postgres > /dev/null 2>&1
  run_sysbench $pg_port "$test" | awk '{print $2}'
}

if [ -n "$pg_bin" ]; then
  LC_ALL=C "$pg_bin/initdb" -D "$work/pg" -U postgres --auth=trust > /dev/null
  LC_ALL=C "$pg_bin/pg_ctl" -D "$work/pg" -o "-p $pg_port -k $work" -l "$work/pg.log" -w start > /dev/null
  trap '"$pg_bin/pg_ctl" -D "$work/pg" -m fast stop > /dev/null' EXIT
fi

echo "| Test | main latency (ms) | PR latency (ms) | Change | main tps | PR tps | Postgres tps | PR / Postgres |"
echo "| --- | --- | --- | --- | --- | --- | --- | --- |"
for test in $tests; do
  read -r main_avg main_tps <<< "$(run_test "$main_bin" "$test")"
  read -r pr_avg pr_tps <<< "$(run_test "$pr_bin" "$test")"
  pg_tps=n/a
  if [ -n "$pg_bin" ]; then
    pg_tps=$(run_postgres_test "$test")
  fi
  change=$(awk -v a="$main_avg" -v b="$pr_avg" 'BEGIN { if (a + 0 > 0 && b + 0 > 0) printf "%+.1f%%", 100 * (b - a) / a; else print "n/a" }')
  ratio=$(awk -v a="$pg_tps" -v b="$pr_tps" 'BEGIN { if (a + 0 > 0 && b + 0 > 0) printf "%.1f%%", 100 * b / a; else print "n/a" }')
  echo "| $test | $main_avg | $pr_avg | $change | $main_tps | $pr_tps | $pg_tps | $ratio |"
done
