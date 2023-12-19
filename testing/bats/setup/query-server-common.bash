SERVER_REQS_INSTALLED="FALSE"
SERVER_PID=""
DEFAULT_DB=""

# wait_for_connection(<PORT>, <TIMEOUT IN MS>) attempts to connect to the sql-server at the specified
# port on localhost, using the $SQL_USER (or 'postgres' if unspecified) as the user name, and trying once
# per second until the millisecond timeout is reached. If a connection is successfully established,
# this function returns 0. If a connection was not able to be established within the timeout period,
# this function returns 1.
wait_for_connection() {
  port=$1
  timeout=$2
  user=${SQL_USER:-postgres}
  end_time=$((SECONDS+($timeout/1000)))

  while [ $SECONDS -lt $end_time ]; do
    run psql -U $user -h localhost -p $port -c "SELECT 1;"
    if [ $status -eq 0 ]; then
      echo "Connected successfully!"
      return 0
    fi
    sleep 1
  done

  echo "Failed to connect to database $DEFAULT_DB on port $port within $timeout ms."
  return 1
}

start_sql_server() {
    DEFAULT_DB="$1"
    DEFAULT_DB="${DEFAULT_DB:=postgres}"
    nativevar DEFAULT_DB "$DEFAULT_DB" /w
    nativevar DOLTGRES_DATA_DIR "$(pwd)" /p # This has to be set in every function that calls doltgresql
    nativevar DOLTGRES_DATA_DIR_CWD "" /w
    logFile="$2"
    PORT=$( definePORT )
    if [[ $logFile ]]
    then
        doltgresql --host 0.0.0.0 --port=$PORT --user "${SQL_USER:-postgres}" > $logFile 2>&1 &
    else
        doltgresql --host 0.0.0.0 --port=$PORT --user "${SQL_USER:-postgres}" &
    fi
    SERVER_PID=$!
    wait_for_connection $PORT 7500
}

# like start_sql_server, but the second argument is a string with all
# arguments to dolt-sql-server (excluding --port, which is defined in
# this func)
start_sql_server_with_args() {
    DEFAULT_DB=""
    nativevar DEFAULT_DB "$DEFAULT_DB" /w
    nativevar DOLTGRES_DATA_DIR "$(pwd)" /p # This has to be set in every function that calls doltgresql
    nativevar DOLTGRES_DATA_DIR_CWD "" /w
    PORT=$( definePORT )
    doltgresql "$@" --port=$PORT &
    SERVER_PID=$!
    wait_for_connection $PORT 7500
}

start_sql_server_with_config() {
    DEFAULT_DB="$1"
    DEFAULT_DB="${DEFAULT_DB:=postgres}"
    nativevar DEFAULT_DB "$DEFAULT_DB" /w
    nativevar DOLTGRES_DATA_DIR "$(pwd)" /p # This has to be set in every function that calls doltgresql
    nativevar DOLTGRES_DATA_DIR_CWD "" /w
    PORT=$( definePORT )
    echo "
log_level: debug

user:
  name: postgres

listener:
  host: 0.0.0.0
  port: $PORT
  max_connections: 10

behavior:
  autocommit: false
" > .cliconfig.yaml
    cat "$2" >> .cliconfig.yaml
    doltgresql --config .cliconfig.yaml &
    SERVER_PID=$!
    wait_for_connection $PORT 7500
}

start_sql_multi_user_server() {
    DEFAULT_DB="$1"
    DEFAULT_DB="${DEFAULT_DB:=postgres}"
    nativevar DEFAULT_DB "$DEFAULT_DB" /w
    nativevar DOLTGRES_DATA_DIR "$(pwd)" /p # This has to be set in every function that calls doltgresql
    nativevar DOLTGRES_DATA_DIR_CWD "" /w
    PORT=$( definePORT )
    echo "
log_level: debug

user:
  name: postgres

listener:
  host: 0.0.0.0
  port: $PORT
  max_connections: 10

behavior:
  autocommit: false
" > .cliconfig.yaml
    doltgresql --config .cliconfig.yaml &
    SERVER_PID=$!
    wait_for_connection $PORT 7500
}

start_multi_db_server() {
    DEFAULT_DB="$1"
    DEFAULT_DB="${DEFAULT_DB:=postgres}"
    nativevar DEFAULT_DB "$DEFAULT_DB" /w
    nativevar DOLTGRES_DATA_DIR "$(pwd)" /p # This has to be set in every function that calls doltgresql
    nativevar DOLTGRES_DATA_DIR_CWD "" /w
    PORT=$( definePORT )
    doltgresql --host 0.0.0.0 --port=$PORT --user postgres --data-dir ./ &
    SERVER_PID=$!
    wait_for_connection $PORT 7500
}

# stop_sql_server stops the SQL server. For cases where it's important
# to wait for the process to exit after the kill signal (e.g. waiting
# for an async replication push), pass 1.
# kill the process if it's still running
stop_sql_server() {
    # Clean up any mysql.sock file in the default, global location
    if [ -f "/tmp/mysql.sock" ]; then rm -f /tmp/mysql.sock; fi
    if [ -f "/tmp/postgres.sock" ]; then rm -f /tmp/mysql.sock; fi
    if [ -f "/tmp/dolt.$PORT.sock" ]; then rm -f /tmp/dolt.$PORT.sock; fi

    wait=$1
    if [ ! -z "$SERVER_PID" ]; then
        # ignore failures of kill command in the case the server is already dead
        run kill $SERVER_PID
        if [ $wait ]; then
            while ps -p $SERVER_PID > /dev/null; do
                sleep .1;
            done
        fi;
    fi
    SERVER_PID=
}

definePORT() {
  for i in {0..99}
  do
    port=$((RANDOM % 4096 + 2048))
    # nc (netcat) returns 0 when it _can_ connect to a port (therefore in use), 1 otherwise.
    run nc -z localhost $port
    if [ "$status" -eq 1 ]; then
      echo $port
      break
    fi
  done
}
