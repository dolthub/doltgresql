#!/bin/bash
# Doltgresql compatibility test runner.
#
# Downloads specified doltgresql release binaries, creates test repositories
# using them, and runs BATS test suites to verify backward, forward, and
# bidirectional compatibility.
#
# Usage: ./runner.sh
# Run from the integration-tests/compatibility directory.
#
# Environment variables (optional):
#   DOLTGRES_SKIP_BACKWARD  — skip backward-compatibility tests if set
#   DOLTGRES_SKIP_FORWARD   — skip forward-compatibility tests if set
#   DOLTGRES_SKIP_BIDIR     — skip bidirectional-compatibility tests if set

set -eo pipefail

PLATFORM_TUPLE=""

# ---------------------------------------------------------------------------
# Platform detection
# ---------------------------------------------------------------------------

get_platform_tuple() {
  local OS ARCH
  OS=$(uname)
  ARCH=$(uname -m)

  if [ "$OS" != Linux ] && [ "$OS" != Darwin ]; then
    echo "tests only support linux or macOS." >&2
    exit 1
  fi

  if [ "$OS" = Linux ]; then
    PLATFORM_TUPLE=linux
  else
    PLATFORM_TUPLE=darwin
  fi

  if [ "$ARCH" = x86_64 ]; then
    PLATFORM_TUPLE="${PLATFORM_TUPLE}-amd64"
  elif [ "$ARCH" = arm64 ] || [ "$ARCH" = aarch64 ]; then
    PLATFORM_TUPLE="${PLATFORM_TUPLE}-arm64"
  else
    echo "unsupported architecture: $ARCH" >&2
    exit 1
  fi

  echo "$PLATFORM_TUPLE"
}

# ---------------------------------------------------------------------------
# Release download
# ---------------------------------------------------------------------------

# download_release <version>
# Downloads doltgresql-<platform>.tar.gz for the given version tag, extracts it
# into binaries/<version>/, and prints the path to the bin directory.
download_release() {
  local ver="$1"
  local dirname="binaries/$ver"
  mkdir -p "$dirname"

  local basename="doltgresql-${PLATFORM_TUPLE}"
  local filename="${basename}.tar.gz"
  local filepath="${dirname}/${filename}"
  local url="https://github.com/dolthub/doltgresql/releases/download/${ver}/${filename}"

  echo "Downloading doltgresql ${ver} for ${PLATFORM_TUPLE} ..." >&2
  curl -L -o "$filepath" "$url"
  tar -zxf "$filepath" -C "$dirname"

  # Binary lives at doltgresql-<os>-<arch>/bin/doltgres
  echo "${dirname}/${basename}/bin"
}

# ---------------------------------------------------------------------------
# Server management helpers used by the runner (not BATS)
# ---------------------------------------------------------------------------

_pick_port() {
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

_write_config() {
  local dir="$1" port="$2"
  cat > "$dir/runner-config.yaml" <<EOF
log_level: warning
behavior:
  read_only: false
  disable_client_multi_statements: false
listener:
  host: localhost
  port: $port
EOF
}

# start_server <binary> <datadir> <logfile> → sets RUNNER_SERVER_PID and RUNNER_SERVER_PORT
start_server() {
  local binary="$1" datadir="$2" logfile="$3"
  RUNNER_SERVER_PORT=$(_pick_port)
  _write_config "$datadir" "$RUNNER_SERVER_PORT"

  PGPASSWORD=password "$binary" -data-dir="$datadir" \
    --config="$datadir/runner-config.yaml" > "$logfile" 2>&1 &
  RUNNER_SERVER_PID=$!

  local end=$((SECONDS + 20))
  while [ $SECONDS -lt $end ]; do
    if PGPASSWORD=password psql -U postgres -h localhost -p "$RUNNER_SERVER_PORT" \
        -c "SELECT 1;" postgres >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.5
  done

  echo "ERROR: server failed to start on port $RUNNER_SERVER_PORT" >&2
  cat "$logfile" >&2
  kill "$RUNNER_SERVER_PID" 2>/dev/null
  return 1
}

stop_server() {
  if [ -n "$RUNNER_SERVER_PID" ]; then
    kill "$RUNNER_SERVER_PID" 2>/dev/null || true
    wait "$RUNNER_SERVER_PID" 2>/dev/null || true
    RUNNER_SERVER_PID=""
    RUNNER_SERVER_PORT=""
  fi
}

RUNNER_SERVER_PID=""
RUNNER_SERVER_PORT=""

# ---------------------------------------------------------------------------
# Repository setup
# ---------------------------------------------------------------------------

# setup_repo <label> [<binary>]
# Creates a test repository under repos/<label>/ using the given binary
# (defaults to doltgres from PATH).  Sets REPO_DIR to the resulting directory.
setup_repo() {
  local label="$1"
  local binary="${2:-doltgres}"
  REPO_DIR="$(pwd)/repos/${label}"
  mkdir -p "$REPO_DIR"
  ./test_files/setup_repo.sh "$REPO_DIR" "$binary"
}

# ---------------------------------------------------------------------------
# Version list helpers
# ---------------------------------------------------------------------------

list_backward_compatible_versions() {
  grep -v '^ *#' < test_files/backward_compatible_versions.txt
}

list_forward_compatible_versions() {
  grep -v '^ *#' < test_files/forward_compatible_versions.txt
}

# ---------------------------------------------------------------------------
# Test runners
# ---------------------------------------------------------------------------

test_backward_compatibility() {
  local ver="$1"
  local bin
  bin=$(download_release "$ver")

  echo "=== Backward compat: creating repo with doltgresql ${ver} ==="
  setup_repo "$ver" "${bin}/doltgres"

  echo "=== Backward compat: testing HEAD doltgresql against repo from ${ver} ==="
  DOLTGRES_TEST_BIN="$(which doltgres)" \
    REPO_DIR="$(pwd)/repos/${ver}" \
    bats --print-output-on-failure ./test_files/bats/compatibility.bats

  DOLTGRES_TEST_BIN="$(which doltgres)" \
    REPO_DIR="$(pwd)/repos/${ver}" \
    bats --print-output-on-failure ./test_files/bats/types_compatibility.bats
}

test_forward_compatibility() {
  if [ -z $1 ]; then
    return
  fi
    
  local ver="$1"
  local bin
  bin=$(download_release "$ver")

  echo "=== Forward compat: testing doltgresql ${ver} against repo from HEAD ==="
  # repos/HEAD was already created by the main flow (see _main).
  DOLTGRES_TEST_BIN="${bin}/doltgres" \
    REPO_DIR="$(pwd)/repos/HEAD" \
    bats --print-output-on-failure ./test_files/bats/compatibility.bats

  DOLTGRES_TEST_BIN="${bin}/doltgres" \
    REPO_DIR="$(pwd)/repos/HEAD" \
    bats --print-output-on-failure ./test_files/bats/types_compatibility.bats
}

test_bidirectional_compatibility() {
  if [ -z $1 ]; then
    return
  fi

  local ver="$1"
  local bin
  bin=$(download_release "$ver")

  local head_bin
  head_bin="$(which doltgres)"

  # Forward direction: old = released version, new = HEAD
  local scratch_fwd="$(pwd)/repos/${ver}-bidir-forward"
  mkdir -p "$scratch_fwd"
  echo "=== Bidirectional (forward): old=${ver}, new=HEAD ==="
  DOLTGRES_LEGACY_BIN="${bin}/doltgres" \
    DOLTGRES_NEW_BIN="$head_bin" \
    REPO_DIR="$scratch_fwd" \
    bats --print-output-on-failure ./test_files/bats/bidirectional/bidirectional_compat.bats

  # Reverse direction: old = HEAD, new = released version
  local scratch_rev="$(pwd)/repos/${ver}-bidir-reverse"
  mkdir -p "$scratch_rev"
  echo "=== Bidirectional (reverse): old=HEAD, new=${ver} ==="
  DOLTGRES_LEGACY_BIN="$head_bin" \
    DOLTGRES_NEW_BIN="${bin}/doltgres" \
    REPO_DIR="$scratch_rev" \
    bats --print-output-on-failure ./test_files/bats/bidirectional/bidirectional_compat.bats
}

# ---------------------------------------------------------------------------
# Cleanup
# ---------------------------------------------------------------------------

cleanup() {
  stop_server
  rm -rf repos binaries
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

_main() {
  PLATFORM_TUPLE=$(get_platform_tuple)

  mkdir -p repos binaries
  trap cleanup EXIT

  # --- Backward compatibility ---
  if [ -z "$DOLTGRES_SKIP_BACKWARD" ] && [ -s "test_files/backward_compatible_versions.txt" ]; then
    echo "=== Running backward compatibility tests ==="
    while IFS= read -r ver; do
      test_backward_compatibility "$ver"
    done < <(list_backward_compatible_versions)
  fi

  # --- Create HEAD repo (used for forward compat and the sanity check) ---
  echo "=== Creating HEAD repo ==="
  setup_repo HEAD

  # --- Forward compatibility ---
  if [ -z "$DOLTGRES_SKIP_FORWARD" ] && [ -s "test_files/forward_compatible_versions.txt" ]; then
    echo "=== Running forward compatibility tests ==="
    while IFS= read -r ver; do
      test_forward_compatibility "$ver"
    done < <(list_forward_compatible_versions)
  fi

  # --- Bidirectional compatibility (uses forward_compatible_versions list) ---
  if [ -z "$DOLTGRES_SKIP_BIDIR" ] && [ -s "test_files/forward_compatible_versions.txt" ]; then
    echo "=== Running bidirectional compatibility tests ==="
    while IFS= read -r ver; do
      test_bidirectional_compatibility "$ver"
    done < <(list_forward_compatible_versions)
  fi

  # --- Sanity check: HEAD against HEAD ---
  echo "=== Sanity check: HEAD doltgresql against HEAD repo ==="
  DOLTGRES_TEST_BIN="$(which doltgres)" \
    REPO_DIR="$(pwd)/repos/HEAD" \
    bats --print-output-on-failure ./test_files/bats/compatibility.bats

  DOLTGRES_TEST_BIN="$(which doltgres)" \
    REPO_DIR="$(pwd)/repos/HEAD" \
    bats --print-output-on-failure ./test_files/bats/types_compatibility.bats
}

_main
