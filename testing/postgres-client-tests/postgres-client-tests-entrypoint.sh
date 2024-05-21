#!/bin/sh

echo "Running mysql-client-tests:"
bats doltgresql/testing/postgres-client-tests/postgres-client-tests.bats
