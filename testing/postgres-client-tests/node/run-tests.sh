#!/bin/sh
source ../helpers.bash

echo "Running $1 tests"
start_doltgres_server
query_server -c "CREATE TABLE IF NOT EXISTS test_table(pk int)" -t
query_server -c "DELETE FROM test_table" -t
query_server -c "INSERT INTO test_table VALUES (1)" -t

cd ..
DOLTGRES_VERSION=$( doltgres --version | sed -nre 's/^[^0-9]*(([0-9]+\.)*[0-9]+).*/\1/p' )
node $1 $USER $PORT $DOLTGRES_VERSION $PWD/testdata
teardown_doltgres_repo