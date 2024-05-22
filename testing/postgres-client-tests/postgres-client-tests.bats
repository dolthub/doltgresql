#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/helpers.bash

# PostgreSQL client tests are set up to test Doltgres as a PostgreSQL server and
# standard PostgreSQL Clients in a wide array of languages.

setup() {
    setup_doltgres_repo

    query_server -c "CREATE TABLE IF NOT EXISTS test_table(pk int)" -t
    query_server -c "DELETE FROM test_table" -t
    query_server -c "INSERT INTO test_table VALUES (1)" -t
}

teardown() {
    cd ..
    teardown_doltgres_repo

    # Check if postgresql is still running. If so stop it
    active=$(service postgresql status)
    if echo "$active" | grep "online"; then
        service postgresql stop
    fi
}

@test "postgres-connector-java client" {
    javac $BATS_TEST_DIRNAME/java/PostgresTest.java
    java -cp $BATS_TEST_DIRNAME/java:$BATS_TEST_DIRNAME/java/postgresql-42.7.3.jar PostgresTest $USER $PORT
}
