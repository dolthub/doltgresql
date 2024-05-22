#!/bin/sh

echo "Running mysql-client-tests:"
bats /postgres-client-tests/postgres-client-tests.bats
