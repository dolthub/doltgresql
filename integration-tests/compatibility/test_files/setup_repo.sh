#!/bin/bash
# Creates a doltgresql repository populated with test data for compatibility testing.
# Usage: setup_repo.sh <datadir> [<doltgres-binary>]
#
# Starts a temporary doltgres server, creates tables / views / branches expected
# by compatibility.bats and types_compatibility.bats, then stops the server.
# The resulting data directory is the "repo" passed to BATS tests via REPO_DIR.

set -eo pipefail

DATADIR="$1"
DOLTGRES_BIN="${2:-doltgres}"

if [ -z "$DATADIR" ]; then
  echo "Usage: setup_repo.sh <datadir> [<doltgres-binary>]" >&2
  exit 1
fi

mkdir -p "$DATADIR"

# ---------------------------------------------------------------------------
# Server helpers
# ---------------------------------------------------------------------------

pick_port() {
  for i in {0..99}; do
    local port=$((RANDOM % 4096 + 2048))
    if ! nc -z localhost "$port" 2>/dev/null; then
      echo "$port"
      return 0
    fi
  done
  echo "ERROR: could not find a free port" >&2
  return 1
}

PORT=$(pick_port)
CONFIGFILE="$DATADIR/setup-config.yaml"

cat > "$CONFIGFILE" <<EOF
log_level: warning
behavior:
  read_only: false
  disable_client_multi_statements: false
listener:
  host: localhost
  port: $PORT
EOF

"$DOLTGRES_BIN" -data-dir="$DATADIR" --config="$CONFIGFILE" \
  > "$DATADIR/setup.log" 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null; wait "$SERVER_PID" 2>/dev/null; exit' EXIT INT TERM

echo "Waiting for doltgres server on port $PORT ..."
end=$((SECONDS + 20))
while [ $SECONDS -lt $end ]; do
  if PGPASSWORD=password psql -U postgres -h localhost -p "$PORT" \
      -c "SELECT 1;" postgres >/dev/null 2>&1; then
    echo "Server ready."
    break
  fi
  sleep 0.5
done

if ! PGPASSWORD=password psql -U postgres -h localhost -p "$PORT" \
    -c "SELECT 1;" postgres >/dev/null 2>&1; then
  echo "ERROR: server failed to start. Log:" >&2
  cat "$DATADIR/setup.log" >&2
  exit 1
fi

# Q — run SQL against the server (heredoc-friendly, single session).
Q() {
  PGPASSWORD=password psql -U postgres -h localhost -p "$PORT" \
    -v ON_ERROR_STOP=1 "$@" postgres
}

# ---------------------------------------------------------------------------
# Step 1: Create initial schema and data on main, then branch from there.
#
# Branch layout:
#   init         — snapshot of initial data (abc w/o alterations)
#   other        — diverges from init; drops x, adds z to abc
#   check_merge  — diverges from main after its alterations; adds rows to def
#   main         — drops w, adds y to abc; clean-merges check_merge
# ---------------------------------------------------------------------------

echo "Step 1: creating initial schema on main ..."

Q <<'SQL'
CREATE TABLE abc (
  pk BIGINT NOT NULL,
  a  TEXT,
  b  DOUBLE PRECISION,
  w  BIGINT,
  x  BIGINT,
  PRIMARY KEY (pk)
);
INSERT INTO abc VALUES
  (0, 'asdf', 1.1, 0, 0),
  (1, 'asdf', 1.1, 0, 0),
  (2, 'asdf', 1.1, 0, 0);

CREATE VIEW view1 AS SELECT 2+2;

CREATE TABLE def (
  i INT CHECK (i > 0)
);
INSERT INTO def VALUES (1), (2), (3);

CREATE TABLE all_types (
  pk            INT NOT NULL PRIMARY KEY,
  c_smallint    SMALLINT,
  c_int         INT,
  c_bigint      BIGINT,
  c_real        REAL,
  c_double      DOUBLE PRECISION,
  c_numeric     NUMERIC(10,2),
  c_char        CHAR(10),
  c_varchar     VARCHAR(255),
  c_text        TEXT,
  c_bytea       BYTEA,
  c_date        DATE,
  c_time        TIME,
  c_timestamp   TIMESTAMP,
  c_timestamptz TIMESTAMPTZ,
  c_boolean     BOOLEAN,
  c_jsonb       JSONB
);

INSERT INTO all_types VALUES (
  1,
  100, 2000000, 9223372036854775807,
  1.5, 2.5, 12345.67,
  'hello', 'hello world', 'text val', E'\\xDEADBEEF',
  '2024-01-15', '13:30:45', '2024-01-15 13:30:45', '2024-01-15 13:30:45+00',
  true, '{"k":"v"}'
);
INSERT INTO all_types VALUES (
  2,
  -100, -2000000, -9223372036854775807,
  -1.5, -2.5, -12345.67,
  'hi', 'hi there', 'text val2', E'\\xC0FFEE',
  '2023-12-31', '23:59:59', '2023-12-31 23:59:59', NULL,
  false, '[1,2,3]'
);
INSERT INTO all_types (pk, c_text) VALUES (3, repeat('t', 500));

CREATE VIEW all_types_view AS SELECT * FROM all_types;
SQL

Q -c "SELECT dolt_add('.');"
Q -c "SELECT dolt_commit('-m', 'initialized data');"

Q <<'EOF'
CREATE TABLE big (
  pk  INT PRIMARY KEY,
  str TEXT
);
EOF

for i in $(seq 0 9); do
    start=$((i * 100 + 1))
    end=$((start + 99))
    stmt="INSERT INTO big VALUES "
    for j in $(seq $start $end); do
        stmt+=$(printf "(%d, 'row %d')" $j $j)
      [ $j -lt $end ] && stmt+=", "
    done
    stmt+=";"
    Q -c "$stmt"
done

# ---------------------------------------------------------------------------
# Step 2: Branch 'init' and 'other' from this initial-data commit.
# ---------------------------------------------------------------------------

echo "Step 2: branching init and other from initial data commit ..."

Q -c "SELECT dolt_branch('init');"
Q -c "SELECT dolt_branch('other');"

# ---------------------------------------------------------------------------
# Step 3: Advance main — alter abc (drop w, add y).
# ---------------------------------------------------------------------------

echo "Step 3: advancing main with abc alterations ..."

Q <<'SQL'
DELETE FROM abc WHERE pk=1;
UPDATE abc SET x = 1 WHERE pk = 0;
INSERT INTO abc VALUES (3, 'data', 1.1, 0, 0);
ALTER TABLE abc DROP COLUMN w;
ALTER TABLE abc ADD COLUMN y BIGINT;
UPDATE abc SET y = 121;
SQL

Q -c "SELECT dolt_add('.');"
Q -c "SELECT dolt_commit('-m', 'made changes to main');"

# Branch 'check_merge' from main's current commit (has abc alterations).
Q -c "SELECT dolt_branch('check_merge');"

# ---------------------------------------------------------------------------
# Step 4: Populate 'other' — different alterations to abc.
# ---------------------------------------------------------------------------

echo "Step 4: populating other branch ..."

Q <<'SQL'
SELECT dolt_checkout('other');
DELETE FROM abc WHERE pk=2;
UPDATE abc SET w = 1 WHERE pk = 0;
INSERT INTO abc VALUES (4, 'data', 1.1, 0, 0);
ALTER TABLE abc DROP COLUMN x;
ALTER TABLE abc ADD COLUMN z BIGINT;
UPDATE abc SET z = 122;
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'made changes to other');
SQL

# ---------------------------------------------------------------------------
# Step 5: Populate 'check_merge' — add rows to def only.
# ---------------------------------------------------------------------------

echo "Step 5: populating check_merge branch ..."

Q <<'SQL'
SELECT dolt_checkout('check_merge');
INSERT INTO def VALUES (5), (6), (7);
SELECT dolt_add('.');
SELECT dolt_commit('-m', 'made changes to check_merge');
SQL

# ---------------------------------------------------------------------------
# Done
# ---------------------------------------------------------------------------

echo ""
echo "Repository setup complete."
echo "  Data dir : $DATADIR"
echo "  Branches : main, init, other, check_merge"

kill "$SERVER_PID"
wait "$SERVER_PID" 2>/dev/null || true
trap - EXIT INT TERM
