setup_doltgres_repo() {
    run psql --version
    if [[ ! "$output" =~ "(PostgreSQL) 15" ]] && [[ ! "$output" =~ "(PostgreSQL) 16" ]]; then
        echo "PSQL must be version 15"
        return 1
    fi

    REPO_NAME="doltgres_repo_$$"
    mkdir $REPO_NAME
    cd $REPO_NAME

    USER="doltgres"
    PORT=$( definePORT )
    CONFIG=$( defineCONFIG $PORT )
    echo "$CONFIG" > config.yaml

    doltgres sql-server --data-dir=.  --socket "dolt.$PORT.sock" -config=config.yaml &
    SERVER_PID=$!
    # Give the server a chance to start
    sleep 2

#    export PGSQL_PWD=""
}

teardown_doltgres_repo() {
    # Clean up any mysql.sock file in the default, global location
    if [ -f "/tmp/mysql.sock" ]; then rm -f /tmp/mysql.sock; fi
    if [ -f "/tmp/postgres.sock" ]; then rm -f /tmp/mysql.sock; fi
    if [ -f "/tmp/dolt.$PORT.sock" ]; then rm -f /tmp/dolt.$PORT.sock; fi

    kill $SERVER_PID
    rm -rf $REPO_NAME
}

query_server() {
    psql -U "${SQL_USER:-postgres}" -h localhost -p $PORT "$@" doltgres
}

definePORT() {
  getPORT=""
  for i in {0..9}
  do
    let getPORT="($$ + $i) % 4096 + 2048"
    portinuse=$(lsof -i -P -n | grep LISTEN | grep $attemptedPORT | wc -l)
      if [ $portinuse -eq 0 ]
      then
        echo "$getPORT"
        break
      fi
  done
}

defineCONFIG() {
    PORT=$1
    cat <<EOF
    behavior:
      read_only: false
      disable_client_multi_statements: false
      dolt_transaction_commit: false

    user:
      name: "doltgres"
      password: "password"

    listener:
      host: localhost
      port: $PORT
EOF
}
