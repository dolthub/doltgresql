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

@test "node postgres client" {
    node $BATS_TEST_DIRNAME/node/index.js $USER $PORT
}

@test "knex node postgres client" {
    DOLTGRES_VERSION=$( doltgres --version | sed -nre 's/^[^0-9]*(([0-9]+\.)*[0-9]+).*/\1/p' )
    echo $DOLTGRES_VERSION
    node $BATS_TEST_DIRNAME/node/knex.js $USER $PORT $DOLTGRES_VERSION
}

@test "node postgres client, workbench stability" {
    DOLTGRES_VERSION=$( doltgres --version | sed -nre 's/^[^0-9]*(([0-9]+\.)*[0-9]+).*/\1/p' )
    echo $DOLTGRES_VERSION
    node $BATS_TEST_DIRNAME/node/workbench.js $USER $PORT $DOLTGRES_VERSION
}


@test "perl DBI:Pg client" {
    perl $BATS_TEST_DIRNAME/perl/postgres-test.pl $USER $PORT
}

@test "ruby pg test" {
    ruby $BATS_TEST_DIRNAME/ruby/pg-test.rb $USER $PORT
}

@test "php pg_connect client" {
    cd $BATS_TEST_DIRNAME/php
    php pg_connect_test.php $USER $PORT
}

@test "php pdo pgsql client" {
    cd $BATS_TEST_DIRNAME/php
    php pdo_connector_test.php $USER $PORT
}

@test "c postgres: libpq connector" {
    (cd $BATS_TEST_DIRNAME/c; make clean; make)
    $BATS_TEST_DIRNAME/c/postgres-c-connector-test $USER $PORT
}
