#!/bin/bash

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

# Run the Go test
go test -run="TestReplication" ./...

# Run the bats test
cd testing/bats
bats replication.bats
