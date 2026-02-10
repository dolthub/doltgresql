start_doltgres_server() {
  REPO_NAME="doltgres_repo_$$"
  mkdir $REPO_NAME
  cd $REPO_NAME

  CONFIG=$( defineCONFIG )
  echo "$CONFIG" > config.yaml

  doltgres --data-dir=. -config=config.yaml &
  SERVER_PID=$!
  # Give the server a chance to start
  sleep 2
}

setup_doltgres_repo() {
  run psql --version
  if [[ ! "$output" =~ "(PostgreSQL) 15" ]] && [[ ! "$output" =~ "(PostgreSQL) 16" ]] && [[ ! "$output" =~ "(PostgreSQL) 17" ]]; then
    echo "PSQL must be version 15, got $output"
    return 1
  fi

  start_doltgres_server
}

teardown_doltgres_repo() {
  cd ..
  kill $SERVER_PID
  rm -rf $REPO_NAME
}

query_server() {
  PGPASSWORD="password" psql -U "postgres" -h localhost -p 5432 "$@" postgres
}

defineCONFIG() {
  cat <<EOF
  behavior:
    read_only: false
    disable_client_multi_statements: false
    dolt_transaction_commit: false

  user:
    name: "postgres"
    password: "password"

  listener:
    host: localhost
    port: 5432
EOF
}
