#!/bin/bash
#set -e
#set -o pipefail

PORT=54171

# Set the working directory to the directory of the script's location
cd "$(cd -P -- "$(dirname -- "$0")" && pwd -P)"

mkdir -p mini_sysbench
cd mini_sysbench

if [ ! -d "./sysbench-lua-scripts" ]; then
  git clone https://github.com/dolthub/sysbench-lua-scripts.git
fi
cp ./sysbench-lua-scripts/*.lua ./

go build -o doltgres.exe ../../cmd/doltgres/

values=("covering_index_scan_postgres" "index_join_postgres" "index_join_scan_postgres" "index_scan_postgres" "oltp_point_select" "oltp_read_only" "select_random_points" "select_random_ranges" "table_scan_postgres" "types_table_scan_postgres")
for value in "${values[@]}"; do
  SYSBENCH_TEST="$value"
  cat <<YAML > dolt-config.yaml
log_level: debug

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

data_dir: .
YAML

  rm -rf ./.dolt
  rm -rf ./postgres
  ./doltgres.exe -config="dolt-config.yaml" 2> prepare.log &
  SERVER_PID="$!"

  sleep 1
  echo "----$SYSBENCH_TEST----"
  sysbench \
    --db-driver="pgsql" \
    --pgsql-host="0.0.0.0" \
    --pgsql-port="$PORT" \
    --pgsql-user="postgres" \
    --pgsql-password="password" \
    --pgsql-db="postgres" \
    "$SYSBENCH_TEST" prepare

  kill -15 "$SERVER_PID"

  echo "----$SYSBENCH_TEST----" 1>> results.log
  ./doltgres.exe -config="dolt-config.yaml" 2> run.log &
  SERVER_PID="$!"
  sleep 1

  sysbench \
    --db-driver="pgsql" \
    --pgsql-host="0.0.0.0" \
    --pgsql-port="$PORT" \
    --pgsql-user="postgres" \
    --pgsql-password="password" \
    --pgsql-db="postgres" \
    --time=15 \
    --db-ps-mode=disable \
    "$SYSBENCH_TEST" run 1>> results.log

  sleep 1
  kill -15 "$SERVER_PID"
  echo "----$SYSBENCH_TEST----" 1>> results.log
done
