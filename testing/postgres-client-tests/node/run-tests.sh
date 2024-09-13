#!/bin/sh
source ../helpers.bash

echo "Running $1 tests"
start_doltgres_server
cd ..
node $1 $USER $PORT $REPO_NAME $PWD/testdata
teardown_doltgres_repo