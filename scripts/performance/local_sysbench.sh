#!/bin/bash
#set -e
#set -o pipefail

SYSBENCH_TEST="covering_index_scan_postgres"
PPROF=0
PORT=54171

while [[ $# -gt 0 ]]; do
  case "$1" in
    --pprof)
      PPROF=1
      ;;
    *)
      SYSBENCH_TEST="$1"
      ;;
  esac
  shift
done

mkdir -p sbtest
cd sbtest

if [ ! -d "./sysbench-lua-scripts" ]; then
  git clone https://github.com/dolthub/sysbench-lua-scripts.git
fi
cp ./sysbench-lua-scripts/*.lua ./

go build -o doltgres.exe ../../../cmd/doltgres/

cat <<YAML > dolt-config.yaml
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

data_dir: .
YAML

rm -rf ./.dolt
rm -rf ./postgres
./doltgres.exe -config="dolt-config.yaml" 2> prepare.log &
SERVER_PID="$!"
sleep 1

echo "----$SYSBENCH_TEST----" 1> results.log
sysbench \
  --db-driver="pgsql" \
  --pgsql-host="0.0.0.0" \
  --pgsql-port="$PORT" \
  --pgsql-user="postgres" \
  --pgsql-password="password" \
  --pgsql-db="postgres" \
  --db-ps-mode=disable \
  --table-size=10000 \
  --percentile=50 \
  --rand-type=uniform \
  --rand-seed=1 \
  "$SYSBENCH_TEST" prepare

kill -15 "$SERVER_PID"

if [ "$PPROF" -eq 1 ]; then
  ./doltgres.exe --prof cpu -config="dolt-config.yaml" 2> run.log &
else
  ./doltgres.exe -config="dolt-config.yaml" 2> run.log &
fi
SERVER_PID="$!"
sleep 1

sysbench \
  --db-driver="pgsql" \
  --pgsql-host="0.0.0.0" \
  --pgsql-port="$PORT" \
  --pgsql-user="postgres" \
  --pgsql-password="password" \
  --pgsql-db="postgres" \
  --db-ps-mode=disable \
  --table-size=10000 \
  --percentile=50 \
  --rand-type=uniform \
  --rand-seed=1 \
  --report-interval=1 \
  --time=120 \
  "$SYSBENCH_TEST" run 1>> results.log

sleep 1
kill -15 "$SERVER_PID"
