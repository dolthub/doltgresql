#!/bin/sh

echo "Running postgres-client-tests:"
bats /postgres-client-tests/postgres-client-tests.bats
