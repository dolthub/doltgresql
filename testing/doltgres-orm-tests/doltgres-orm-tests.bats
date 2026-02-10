#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/helpers.bash

setup() {
  setup_doltgres_repo
}

teardown() {
  teardown_doltgres_repo

  # Check if postgresql is still running. If so stop it
  active=$(service postgresql status)
  if echo "$active" | grep "online"; then
      service postgresql stop
  fi
}

@test "Drizzle smoke test" {
  # the schema should be empty
  # the dolt system tables are filtered out
  cd $BATS_TEST_DIRNAME/drizzle
  npm i drizzle-orm pg dotenv
  npm i -D drizzle-kit tsx @types/pg
  npx drizzle-kit push

  # we can check if 'components table was created'
  query_server -c "SELECT * FROM users" -t
  run query_server -c "SELECT * FROM users" -t
  [ "$status" -eq 0 ]

  npx tsx src/index.ts
  query_server -c "SELECT * FROM users" -t
  run query_server -c "SELECT age FROM users" -t
  [ "$status" -eq 0 ]
  [[ "$output" =~ "31" ]] || false
}
