#!/bin/bash
set -euo pipefail

# Tests that DuckDB's postgres extension can import a table from a running Doltgres server, and
# export a table back to it. DuckDB attaches to the server and reads remote tables by issuing
# COPY (SELECT ...) TO STDOUT (FORMAT "binary") and writes them with COPY ... FROM STDIN (FORMAT "binary"),
# so this exercises Doltgres's catalog tables and the binary COPY TO / COPY FROM implementations.
# The duckdb_test table is created by the caller.

USER=$1
PORT=$2

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT
DB="$WORKDIR/local.duckdb"

# Import the table from the running server into a local DuckDB table
duckdb "$DB" -c "
LOAD postgres;
ATTACH 'host=localhost port=$PORT dbname=postgres user=$USER password=password' AS pg (TYPE postgres, READ_ONLY);
CREATE TABLE local_copy AS SELECT * FROM pg.public.duckdb_test;
DETACH pg;
"

# Then check the contents of the imported table in a fresh DuckDB session
actual=$(duckdb "$DB" -csv -c "SELECT * FROM local_copy ORDER BY pk;")
expected=$(cat <<'EOF'
pk,name,price,active,created
1,first,1.5,true,2025-01-01 12:00:00
2,NULL,-2.25,false,2025-06-15 08:30:00
3,"",0.0,NULL,NULL
4,"héllo, ""world""",10000000000.0,true,1999-12-31 23:59:59
EOF
)

if [ "$actual" != "$expected" ]; then
    echo "DuckDB imported table did not match expected contents"
    echo "expected:"
    echo "$expected"
    echo "actual:"
    echo "$actual"
    exit 1
fi

# Now write the local table back to the server as a new table (DuckDB does this via binary
# COPY FROM STDIN), then read it back from the server and check that nothing was lost
duckdb "$DB" -c "
LOAD postgres;
ATTACH 'host=localhost port=$PORT dbname=postgres user=$USER password=password' AS pg (TYPE postgres);
CREATE TABLE pg.public.duckdb_written AS SELECT * FROM local_copy;
DETACH pg;
"

actual=$(duckdb -csv -c "
LOAD postgres;
ATTACH 'host=localhost port=$PORT dbname=postgres user=$USER password=password' AS pg (TYPE postgres, READ_ONLY);
SELECT * FROM pg.public.duckdb_written ORDER BY pk;
")

if [ "$actual" != "$expected" ]; then
    echo "Table written back to Doltgres by DuckDB did not match expected contents"
    echo "expected:"
    echo "$expected"
    echo "actual:"
    echo "$actual"
    exit 1
fi

echo "DuckDB postgres extension test passed"
