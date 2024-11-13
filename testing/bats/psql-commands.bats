#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server
    query_server <<SQL
    CREATE TABLE test1 (pk BIGINT PRIMARY KEY, v1 SMALLINT);
    CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 INTEGER, v2 SMALLINT);
    INSERT INTO test1 VALUES (1, 2), (6, 7);
    INSERT INTO test2 VALUES (3, 4, 5), (8, 9, 0);
    CREATE VIEW testview AS SELECT * FROM test1;
SQL
}

teardown() {
    teardown_common
}

@test 'psql-commands: \l' {
    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "postgres" ]] || false
}

@test 'psql-commands: \dt' {
    run query_server --csv -c "\dt"
    echo "$output"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "public,test1,table,postgres" ]] || false
    [[ "$output" =~ "public,test2,table,postgres" ]] || false
    [ "${#lines[@]}" -eq 3 ]
}

@test 'psql-commands: \d' {
    run query_server --csv -c "\d"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "public,test1,table,postgres" ]] || false
    [[ "$output" =~ "public,test2,table,postgres" ]] || false
    [[ "$output" =~ "public,testview,view,postgres" ]] || false
    [ "${#lines[@]}" -eq 4 ]
}

@test 'psql-commands: \d table' {
    skip "this command has not yet been implemented"
}

@test 'psql-commands: \dn' {
    run query_server --csv -c "\dn"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "public,pg_database_owner" ]] || false
    [ "${#lines[@]}" -eq 2 ]
}

@test 'psql-commands: \df' {
    run query_server -c "\df"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "0 rows" ]] || false
}

@test 'psql-commands: \dv' {
    run query_server --csv -c "\dv"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "public,testview,view,postgres" ]] || false
    [ "${#lines[@]}" -eq 2 ]
}

@test 'psql-commands: \du' {
    skip "users have not yet been implemented"
}
