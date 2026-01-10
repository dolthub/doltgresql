#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server

}

teardown() {
    teardown_common
}

@test 'types: boolean type' {
    query_server <<SQL
    CREATE TABLE t_boolean (id INTEGER primary key, v1 BOOLEAN);
    INSERT INTO t_boolean VALUES (1, 'true'), (2, 'false');
SQL

    run query_server --csv -c "SELECT * FROM t_boolean;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1,t" ]] || false
    [[ "$output" =~ "2,f" ]] || false
}

@test 'types: boolean array type' {
    query_server <<SQL
    CREATE TABLE t_boolean_array (id INTEGER primary key, v1 BOOLEAN[]);
    INSERT INTO t_boolean_array VALUES (1, ARRAY[true, false]), (2, ARRAY[false, true]), (3, ARRAY[true, true]), (4, ARRAY[false, false]), (5, ARRAY[true]), (6, ARRAY[false]);
SQL

    run query_server --csv -c "SELECT * FROM t_boolean_array;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ '1,"{t,f}"' ]] || false
		[[ "$output" =~ '2,"{f,t}"' ]] || false
		[[ "$output" =~ '3,"{t,t}"' ]] || false
		[[ "$output" =~ '4,"{f,f}"' ]] || false
		[[ "$output" =~ '5,{t}' ]] || false
		[[ "$output" =~ '6,{f}' ]] || false
}

@test 'types: VALUES clause mixed int and decimal' {
    # Integer first, then decimal - should resolve to numeric
    run query_server -t -c "SELECT * FROM (VALUES(1),(2.01),(3)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "2.01" ]] || false
    [[ "$output" =~ "3" ]] || false
}

@test 'types: VALUES clause decimal first then int' {
    # Decimal first, then integers - should resolve to numeric
    run query_server -t -c "SELECT * FROM (VALUES(1.01),(2),(3)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1.01" ]] || false
    [[ "$output" =~ "2" ]] || false
    [[ "$output" =~ "3" ]] || false
}

@test 'types: VALUES clause SUM with mixed types' {
    # SUM should work directly now that VALUES has correct type
    run query_server -t -c "SELECT SUM(n) FROM (VALUES(1),(2.01),(3)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "6.01" ]] || false
}

@test 'types: VALUES clause multiple columns mixed types' {
    run query_server -t -c "SELECT * FROM (VALUES(1, 'a'), (2.5, 'b')) v(num, str);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "a" ]] || false
    [[ "$output" =~ "2.5" ]] || false
    [[ "$output" =~ "b" ]] || false
}
