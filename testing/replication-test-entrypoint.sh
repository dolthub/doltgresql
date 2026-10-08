#!/bin/bash

# exit if any command exits wtih non-zero
set -e

# run the original postgres setup
/usr/local/bin/docker-ensure-initdb.sh -c wal_level=logical

# Modify the WAL replication settings
echo "wal_level = logical" >> /var/lib/postgresql/data/postgresql.conf

# Start PostgreSQL as the postgres user
sudo -u postgres /usr/lib/postgresql/16/bin/pg_ctl \
     -D /var/lib/postgresql/data \
     -l /var/lib/postgresql/data/logfile \
     start

# Wait for PostgreSQL to become ready for requests
until pg_isready -h localhost -p 5432
do
  echo "Waiting for PostgreSQL to become ready..."
  sleep 1
done

# Run the replication tests against the Rust server
DOLTGRES_REPLICATION_PRIMARY="postgres://postgres:password@localhost:5432/postgres?sslmode=disable" \
  DOLTGRES_TEST_TARGET="doltgres:$(pwd)/target/release/doltgres" \
  cargo test --profile quick -p server --test replication

# Run the bats test
cd testing/bats
bats replication.bats
