#!/bin/bash
set -eo pipefail

# _log prints a timestamped (ISO 8601), color-coded structured log message.
# The message includes a log level and the message itself.
# If <MESSAGE> is omitted, it reads from stdin, allowing multi-line input.
#
# Arguments:
#   $1 - <LEVEL>   : Log level (e.g., Warn, Error, Debug)
#   $2 - <MESSAGE> : Message to log; if omitted, the function reads from stdin
#
# Usage:
#   _log <LEVEL> [MESSAGE]
#   _log Warn "Disk space low"
#   echo "Database connection lost" | _log Error
#
# Output:
#   2025-10-16T12:34:56+00:00 [Warn] [Entrypoint] Disk space low
#   2025-10-16T12:35:01+00:00 [Error] [Entrypoint] Database connection lost
_log() {
  local level="$1"; shift

  local dt
  dt="$(date --rfc-3339=seconds)"

  local color_reset="\033[0m"
  local color=""
  case "$level" in
  Warn)  color="\033[1;33m" ;; # yellow
  Error) color="\033[1;31m" ;; # red
  Debug) color="\033[1;34m" ;; # blue
  esac

  local msg="$*"
  if [ "$#" -eq 0 ]; then
    msg="$(cat)"
  fi

  printf '%b%s [%s] [Entrypoint] %s%b\n' "$color" "$dt" "$level" "$msg" "$color_reset"
}


# _dbg logs a message of type 'Debug' using log.
_dbg() {
  _log Debug "$@"
}

# mysql_note logs a message of type 'Note' using log.
note() {
  _log Note "$@"
}

# mysql_warn logs a message of type 'Warning' using log and writes to stderr.
warn() {
  _log Warn "$@" >&2
}

# log_error logs a message of type 'ERROR' using log, writes to stderr, prints a container removal hint, and
# exits with status 1.
log_error() {
  _log Error "$@" >&2
  note "Remove this container with 'docker rm -f <container_name>' before retrying"
  exit 1
}

# exec_sql executes a local SQL query, retrying until success or timeout. Ensures reliability during slow
# container or resource startup. On timeout, it prints the provided error prefix followed by filtered Doltgres output.
# Errors are parsed to remove blank lines and extract only relevant error text. Use --show-result to display successful
# query results.
#
# Usage:
#   exec_sql [--show-result] "<ERROR_MESSAGE>" "<QUERY>"
#   exec_sql [--show-result] "<ERROR_MESSAGE>" < /docker-entrypoint-initdb.d/init.sql
#   cat /docker-entrypoint-initdb.d/init.sql | exec_sql [--show-result] "<ERROR_MESSAGE>"
#
# Output:
#   Prints query output only if --show-result is specified.
exec_sql() {
  local show_result=0
  if [ "$1" = "--show-result" ]; then
    show_result=1
    shift
  fi

  local error_message="$1"
  local query="${2:-}"
  local timeout="${DOLTGRES_SERVER_TIMEOUT:-300}"
  local start_time now output status

  start_time=$(date +%s)

  while true; do
    if [ -n "$query" ]; then
      output=$(psql -c "$query" 2>&1)
      status=$?
    else
      set +e # tmp disabled to initdb.d/ file err
      output=$(psql < /dev/stdin 2>&1)
      status=$?
      set -e
    fi

    if [ "$status" -eq 0 ]; then
      [ "$show_result" -eq 1 ] && echo "$output" | grep -v "^$" || true
      return 0
    fi

    if echo "$output" | grep -qiE "Error [0-9]+ \([A-Z0-9]+\)"; then
      log_error "$error_message$(echo "$output" | grep -iE "Error|error")"
    fi

    if [ "$timeout" -ne 0 ]; then
      now=$(date +%s)
      if [ $((now - start_time)) -ge "$timeout" ]; then
        log_error "$error_message$(echo "$output" | grep -iE "Error|error" || true)"
      fi
    fi

    sleep 1
  done
}

CONTAINER_DATA_DIR="/var/lib/doltgres"
INIT_COMPLETED="$CONTAINER_DATA_DIR/.init_completed"

# TODO: remove
DOLT_CONFIG_DIR="/etc/dolt/doltcfg.d"
SERVER_CONFIG_DIR="/etc/doltgres/servercfg.d"
DOLT_ROOT_PATH="/.dolt"
SERVER_PID=-1

# check_for_doltgres_binary verifies that the dolt binary is present and executable in the system PATH.
# If not found or not executable, it logs an error and exits.
check_for_doltgres_binary() {
  local doltgres_bin
  doltgres_bin=$(which doltgres)
  if [ ! -x "$doltgres_bin" ]; then
    log_error "doltgres binary executable not found"
  fi
}

# get_env_var returns the value of an environment variable, preferring DOLT_* over MYSQL_*.
# Arguments:
#   $1 - The base variable name (e.g., "USER" for MYSQL_USER or DOLT_USER)
# Output:
#   Prints the value of the first set variable, or an empty string if neither is set.
get_env_var() {
  local var_name="$1"
  local dolt_var="DOLT_${var_name}"
  local mysql_var="MYSQL_${var_name}"

  if [ -n "${!dolt_var}" ]; then
    echo "${!dolt_var}"
  elif [ -n "${!mysql_var}" ]; then
    echo "${!mysql_var}"
  else
    echo ""
  fi
}

# get_env_var_name returns the name of the environment variable that is set, preferring DOLT_* over MYSQL_*.
# Arguments:
#   $1 - The base variable name (e.g., "USER" for MYSQL_USER or DOLT_USER)
# Output:
#   Prints the name of the first set variable, or both names if neither is set.
get_env_var_name() {
  local var_name="$1"
  local dolt_var="DOLT_${var_name}"
  local mysql_var="MYSQL_${var_name}"

  if [ -n "${!dolt_var}" ]; then
    echo "DOLT_${var_name}"
  elif [ -n "${!mysql_var}" ]; then
    echo "MYSQL_${var_name}"
  else
    echo "MYSQL_${var_name}/DOLT_${var_name}"
  fi
}

# get_config_file_path_if_exists checks for config files of a given type in a directory.
# Arguments:
#   $1 - Directory to search in
#   $2 - File type/extension to search for (e.g., 'json', 'yaml')
# Output:
#   Sets CONFIG_PROVIDED to the path of the config file if exactly one is found, or empty otherwise.
#   Logs a warning if multiple config files are found and uses the default config.
get_config_file_path_if_exists() {
  CONFIG_PROVIDED=
  local CONFIG_DIR=$1
  local FILE_TYPE=$2
  if [ -d "$CONFIG_DIR" ]; then
    note "Checking for config provided in $CONFIG_DIR"
    local number_of_files_found
    number_of_files_found=$(find "$CONFIG_DIR" -type f -name "*.$FILE_TYPE" | wc -l)
    if [ "$number_of_files_found" -gt 1 ]; then
      CONFIG_PROVIDED=
      warn "Multiple config files found in $CONFIG_DIR, using default config"
    elif [ "$number_of_files_found" -eq 1 ]; then
      local files_found
      files_found=$(ls "$CONFIG_DIR"/*."$FILE_TYPE")
      note "$files_found file is found"
      CONFIG_PROVIDED=$files_found
    else
      CONFIG_PROVIDED=
    fi
  else
      note "No config dir found in $CONFIG_DIR"
  fi
}

# docker_process_init_files Runs files found in /docker-entrypoint-initdb.d before the server is started.
# Taken from https://github.com/docker-library/mysql/blob/master/8.0/docker-entrypoint.sh
# Usage:
#   docker_process_init_files [file [file ...]]
#   e.g., docker_process_init_files /always-initdb.d/*
# Processes initializer files based on file extensions.
docker_process_init_files() {
  local f
  echo
  for f; do
    case "$f" in
    *.sh)
      if [ -x "$f" ]; then
        note "$0: running $f"
        if ! "$f"; then
          log_error "Failed to execute $f: "
        fi
      else
        note "$0: sourcing $f"
        if ! . "$f"; then
          log_error "Failed to execute $f: "
        fi
      fi
      ;;
    *.sql)
      note "$0: running $f"
      exec_sql --show-result "Failed to execute $f: " < "$f"
      ;;
    *.sql.bz2)
      note "$0: running $f"
      bunzip2 -c "$f" | exec_sql --show-result "Failed to execute $f: "
      ;;
    *.sql.gz)
      note "$0: running $f"
      gunzip -c "$f" | exec_sql --show-result "Failed to execute $f: "
      ;;
    *.sql.xz)
      note "$0: running $f"
      xzcat "$f" | exec_sql --show-result "Failed to execute $f: "
      ;;
    *.sql.zst)
      note "$0: running $f"
      zstd -dc "$f" | exec_sql --show-result "Failed to execute $f: "
      ;;
    *)
      warn "$0: ignoring $f"
      ;;
    esac
    echo
  done
}

# create_database_from_env creates a database if the DATABASE environment variable is set.
# It retrieves the database name from environment the env var DATABASE
# and attempts to create the database using exec_sql.
create_database_from_env() {
  local database
  database=$(get_env_var "DATABASE")

  if [ -n "$database" ]; then
    note "Creating database '${database}'"
    exec_sql "Failed to create database '$database': " "CREATE DATABASE IF NOT EXISTS \"$database\";"
  fi
}

# create_user_from_env creates a new database user from environment variables.
# It looks for USER/PASSWORD and optionally grants access to a database.
# Requires both USER and PASSWORD to be set; if only the password is set, it logs a warning and does nothing.
# It does not allow creating a 'root' user via these environment variables.
create_user_from_env() {
  local user
  local password
  local database

  user=$(get_env_var "USER")
  password=$(get_env_var "PASSWORD")
  database=$(get_env_var "DATABASE")

  if [ "$user" = 'root' ]; then
    log_error "$(get_env_var_name "USER")="root", $(get_env_var_name "USER") and $(get_env_var_name "PASSWORD") are for configuring the regular user and cannot be used for the root user."
  fi

  if [ -n "$user" ] && [ -z "$password" ]; then
    log_error "$(get_env_var_name "USER") specified, but missing $(get_env_var_name "PASSWORD"); user creation requires a password."
  elif [ -z "$user" ] && [ -n "$password" ]; then
    warn "$(get_env_var_name "PASSWORD") specified, but missing $(get_env_var_name "USER"); password will be ignored"
    return
  fi

  if [ -n "$user" ]; then
    local user_host
    user_host=$(get_env_var "USER_HOST")
    user_host="${user_host:-${DOLT_ROOT_HOST:-localhost}}"

    note "Creating user '${user}@${user_host}'"
    exec_sql "Failed to create user '$user': " "CREATE USER IF NOT EXISTS '$user'@'$user_host' IDENTIFIED BY '$password';"
    exec_sql "Failed to grant server access to user '$user': " "GRANT USAGE ON *.* TO '$user'@'$user_host';"

    if [ -n "$database" ]; then
      exec_sql "Failed to grant permissions to user '$user' on database '$database': " "GRANT ALL ON \`$database\`.* TO '$user'@'$user_host';"
    fi
  fi
}

# is_port_open checks if a TCP port is open on a given host.
# Arguments:
#   $1 - Host (IP or hostname)
#   $2 - Port number
# Returns:
#   0 if the port is open, non-zero otherwise.
is_port_open() {
  local host="$1"
  local port="$2"
  timeout 1 bash -c "cat < /dev/null > /dev/tcp/$host/$port" &>/dev/null
  return $?
}

# start_server starts the Doltgres server in the background and waits until it is ready to accept connections.
# It manages the server process, restarts it if necessary, and checks for readiness by probing the configured port.
# The function retries until the server is available or a timeout is reached, handling process management and logging.
# Arguments:
#   $@ - Additional arguments to pass to `doltgres`
# Returns:
#   0 if the server starts successfully and is ready to accept connections; exits with error otherwise.
start_server() {
  local timeout="${DOLTGRES_SERVER_TIMEOUT:-300}"
  local start_time
  start_time=$(date +%s)

  SERVER_PID=-1

  trap 'note "Caught Ctrl+C, shutting down Doltgres server..."; [ $SERVER_PID -ne -1 ] && kill "$SERVER_PID"; exit 1' INT TERM

  while true; do
    if [ "$SERVER_PID" -eq -1 ] || ! kill -0 "$SERVER_PID" 2>/dev/null; then
      [ "$SERVER_PID" -ne -1 ] && wait "$SERVER_PID" 2>/dev/null || true
      SERVER_PID=-1
      # echo "running dlv --listen=:2345 --headless=true --api-version=2 exec /usr/local/bin/doltgres -- $@"
      # dlv --listen=:2345 --headless=true --api-version=2 exec /usr/local/bin/doltgres -- "$@" 2>&1 &
      echo "running doltgres $@"
      doltgres "$@" 2>&1 &
      SERVER_PID=$!

    fi

    if is_port_open "0.0.0.0" 5432; then
      note "Doltgres server started."
      return 0
    fi

    local now elapsed
    now=$(date +%s)
    elapsed=$((now - start_time))
    if [ "$elapsed" -ge "$timeout" ]; then
      kill "$SERVER_PID" 2>/dev/null || true
      wait "$SERVER_PID" 2>/dev/null || true
      SERVER_PID=-1
      log_error "Doltgres server failed to start within $timeout seconds"
    fi

    sleep 1
  done
}

# _main is the main entrypoint for the Dolt Docker container initialization.
_main() {
  check_for_doltgres_binary

  local doltgres_version
  doltgres_version=$(doltgres --version | cut -f3 -d " ")
  note "Entrypoint script for Doltgres Server $doltgres_version starting..."

  declare -g CONFIG_PROVIDED

  CONFIG_PROVIDED=

  # if there is a single yaml provided in /etc/doltgres/servercfg.d directory,
  # it will be used to start the server with --config flag.
  get_config_file_path_if_exists "$SERVER_CONFIG_DIR" "yaml"
  if [ -n "$CONFIG_PROVIDED" ]; then
    set -- "$@" --config="$CONFIG_PROVIDED"
  fi

  note "Starting Doltgres server"
  
  start_server "$@"

  create_database_from_env

  create_user_from_env

  if [[ ! -f $INIT_COMPLETED ]]; then
    if ls /docker-entrypoint-initdb.d/* >/dev/null 2>&1; then
      docker_process_init_files /docker-entrypoint-initdb.d/*
    else
      warn "No files found in /docker-entrypoint-initdb.d/ to process"
    fi
    touch "$INIT_COMPLETED"
  fi

  note "Doltgres running. Ready for connections."
  wait "$SERVER_PID"
}

_main "$@"
