# Common helpers for doltgresql compatibility BATS tests.
# Loaded by all test files in this directory tree via:
#   load $BATS_TEST_DIRNAME/../helper/common.bash   (from subdirs)
# or
#   load $BATS_TEST_DIRNAME/helper/common.bash      (from top level)

# ---------------------------------------------------------------------------
# Server lifecycle
# ---------------------------------------------------------------------------

COMPAT_SERVER_PID=""
COMPAT_SERVER_PORT=""

# pick_port — find an unused TCP port in the range 2048–6144.
pick_port() {
  for i in {0..99}; do
    port=$((RANDOM % 4096 + 2048))
    if ! nc -z localhost "$port" 2>/dev/null; then
      echo "$port"
      return 0
    fi
  done
  echo "ERROR: could not find a free port" >&2
  return 1
}

# write_config <dir> <port> — write a minimal doltgres config.yaml into <dir>.
write_config() {
  local dir="$1"
  local port="$2"
  cat > "$dir/compat-config.yaml" <<EOF
log_level: warning
behavior:
  read_only: false
  disable_client_multi_statements: false
listener:
  host: localhost
  port: $port
EOF
}

# start_doltgres <binary> <datadir> [logfile]
# Starts a doltgres server and waits for it to accept connections.
# Sets COMPAT_SERVER_PID and COMPAT_SERVER_PORT.
start_doltgres() {
  local binary="${1:?start_doltgres: binary required}"
  local datadir="${2:?start_doltgres: datadir required}"
  local logfile="${3:-/dev/null}"

  COMPAT_SERVER_PORT=$(pick_port)
  write_config "$datadir" "$COMPAT_SERVER_PORT"

  PGPASSWORD=password "$binary" -data-dir="$datadir" --config="$datadir/compat-config.yaml" \
    > "$logfile" 2>&1 &
  COMPAT_SERVER_PID=$!

  # Wait up to 15 s
  local end=$((SECONDS + 15))
  while [ $SECONDS -lt $end ]; do
    if PGPASSWORD=password psql -U postgres -h localhost -p "$COMPAT_SERVER_PORT" \
        -c "SELECT 1;" postgres >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.5
  done

  echo "ERROR: doltgres server failed to start on port $COMPAT_SERVER_PORT" >&2
  if [ -f "$logfile" ]; then cat "$logfile" >&2; fi
  kill "$COMPAT_SERVER_PID" 2>/dev/null
  COMPAT_SERVER_PID=""
  return 1
}

# stop_doltgres — kill the server started by start_doltgres.
stop_doltgres() {
  if [ -n "$COMPAT_SERVER_PID" ]; then
    kill "$COMPAT_SERVER_PID" 2>/dev/null || true
    wait "$COMPAT_SERVER_PID" 2>/dev/null || true
    COMPAT_SERVER_PID=""
    COMPAT_SERVER_PORT=""
  fi
}

# ---------------------------------------------------------------------------
# Query helpers
# ---------------------------------------------------------------------------

# sql [psql-args...] — run psql against the current server's postgres database.
# Use with heredoc for multi-statement sessions:
#   sql <<SQL
#     SELECT dolt_checkout('other');
#     SELECT * FROM abc;
#   SQL
# Or with -c for a single statement:
#   run sql -c "SELECT count(*) FROM abc;"
sql() {
  PGPASSWORD=password psql -U postgres -h localhost -p "$COMPAT_SERVER_PORT" \
    -v ON_ERROR_STOP=1 "$@" postgres
}

# sql_csv [psql-args...] — same as sql but outputs CSV (--csv flag).
sql_csv() {
  PGPASSWORD=password psql --csv -U postgres -h localhost -p "$COMPAT_SERVER_PORT" \
    -v ON_ERROR_STOP=1 "$@" postgres
}

# ---------------------------------------------------------------------------
# Repo isolation
# ---------------------------------------------------------------------------

# copy_repo <src> <dst> — copy a data directory for test isolation.
copy_repo() {
  cp -Rpf "$1" "$2"
}
