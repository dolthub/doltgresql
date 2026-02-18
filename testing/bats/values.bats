#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server

}

teardown() {
    teardown_common
}

@test 'values: mixed int and decimal' {
    # Integer first, then decimal - should resolve to numeric
    run query_server -t -c "SELECT * FROM (VALUES(1),(2.01),(3)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "2.01" ]] || false
    [[ "$output" =~ "3" ]] || false
}

@test 'values: decimal first then int' {
    # Decimal first, then integers - should resolve to numeric
    run query_server -t -c "SELECT * FROM (VALUES(1.01),(2),(3)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1.01" ]] || false
    [[ "$output" =~ "2" ]] || false
    [[ "$output" =~ "3" ]] || false
}

@test 'values: SUM with mixed types' {
    # SUM should work directly now that VALUES has correct type
    run query_server -t -c "SELECT SUM(n) FROM (VALUES(1),(2.01),(3)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "6.01" ]] || false
}

@test 'values: multiple columns mixed types' {
    run query_server -t -c "SELECT * FROM (VALUES(1, 'a'), (2.5, 'b')) v(num, str);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "a" ]] || false
    [[ "$output" =~ "2.5" ]] || false
    [[ "$output" =~ "b" ]] || false
}

@test 'values: SUM with explicit cast' {
    run query_server -t -c "SELECT SUM(n::numeric) FROM (VALUES(1),(2.01),(3)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "6.01" ]] || false
}

@test 'values: MIN and MAX with mixed types' {
    run query_server -t -c "SELECT MIN(n), MAX(n) FROM (VALUES(1),(2.5),(3),(0.5)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "0.5" ]] || false
    [[ "$output" =~ "3" ]] || false
}

@test 'values: GROUP BY with mixed types' {
    run query_server -t -c "SELECT n, COUNT(*) FROM (VALUES(1),(2.5),(1),(3.5),(2.5)) v(n) GROUP BY n ORDER BY n;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "2.5" ]] || false
    [[ "$output" =~ "3.5" ]] || false
}

@test 'values: SUM GROUP BY with mixed types' {
    run query_server -t -c "SELECT category, SUM(amount) FROM (VALUES('a', 1),('b', 2.5),('a', 3),('b', 4.5)) v(category, amount) GROUP BY category ORDER BY category;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "a" ]] || false
    [[ "$output" =~ "4" ]] || false
    [[ "$output" =~ "b" ]] || false
    [[ "$output" =~ "7.0" ]] || false
}

@test 'values: DISTINCT with mixed types' {
    run query_server -t -c "SELECT DISTINCT n FROM (VALUES(1),(2.5),(1),(2.5),(3)) v(n) ORDER BY n;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "2.5" ]] || false
    [[ "$output" =~ "3" ]] || false
}

@test 'values: ORDER BY with mixed types' {
    run query_server -t -c "SELECT * FROM (VALUES(3),(1.5),(2),(4.5)) v(n) ORDER BY n;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1.5" ]] || false
    [[ "$output" =~ "2" ]] || false
    [[ "$output" =~ "3" ]] || false
    [[ "$output" =~ "4.5" ]] || false
}

@test 'values: ORDER BY DESC with mixed types' {
    run query_server -t -c "SELECT * FROM (VALUES(3),(1.5),(2),(4.5)) v(n) ORDER BY n DESC;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "4.5" ]] || false
    [[ "$output" =~ "3" ]] || false
    [[ "$output" =~ "2" ]] || false
    [[ "$output" =~ "1.5" ]] || false
}

@test 'values: LIMIT with mixed types' {
    run query_server -t -c "SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) LIMIT 3;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "2.5" ]] || false
    [[ "$output" =~ "3" ]] || false
    ! [[ "$output" =~ "4.5" ]] || false
    ! [[ "$output" =~ " 5" ]] || false
}

@test 'values: WHERE filter with mixed types' {
    run query_server -t -c "SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) WHERE n > 2;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "2.5" ]] || false
    [[ "$output" =~ "3" ]] || false
    [[ "$output" =~ "4.5" ]] || false
    [[ "$output" =~ "5" ]] || false
}

@test 'values: NULLs with mixed types' {
    run query_server -t -c "SELECT * FROM (VALUES(1),(NULL),(2.5)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "2.5" ]] || false
}

@test 'values: all same type no cast needed' {
    run query_server -t -c "SELECT * FROM (VALUES(1),(2),(3)) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "2" ]] || false
    [[ "$output" =~ "3" ]] || false
}

@test 'values: all string literals' {
    run query_server -t -c "SELECT * FROM (VALUES('a'),('b'),('c')) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "a" ]] || false
    [[ "$output" =~ "b" ]] || false
    [[ "$output" =~ "c" ]] || false
}

@test 'values: string concatenation' {
    run query_server -t -c "SELECT n || '!' FROM (VALUES('hello'),('world')) v(n);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "hello!" ]] || false
    [[ "$output" =~ "world!" ]] || false
}

@test 'values: type mismatch bool and int errors' {
    run query_server -t -c "SELECT * FROM (VALUES(true),(1),(false)) v(n);"
    [ "$status" -ne 0 ]
    [[ "$output" =~ "cannot be matched" ]] || false
}

@test 'values: JOIN with same types' {
    run query_server -t -c "SELECT a.n, b.label FROM (VALUES(1),(2),(3)) a(n) JOIN (VALUES(1, 'one'),(2, 'two'),(3, 'three')) b(id, label) ON a.n = b.id;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "one" ]] || false
    [[ "$output" =~ "2" ]] || false
    [[ "$output" =~ "two" ]] || false
    [[ "$output" =~ "3" ]] || false
    [[ "$output" =~ "three" ]] || false
}

@test 'values: JOIN with mixed types' {
    run query_server -t -c "SELECT a.n, b.label FROM (VALUES(1),(2.5),(3)) a(n) JOIN (VALUES(1, 'one'),(3, 'three')) b(id, label) ON a.n = b.id;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "one" ]] || false
    [[ "$output" =~ "3" ]] || false
    [[ "$output" =~ "three" ]] || false
}

@test 'values: CTE with mixed types' {
    run query_server -t -c "WITH nums AS (SELECT * FROM (VALUES(1),(2.5),(3)) v(n)) SELECT * FROM nums;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "2.5" ]] || false
    [[ "$output" =~ "3" ]] || false
}

@test 'values: CTE SUM with mixed types' {
    run query_server -t -c "WITH nums AS (SELECT * FROM (VALUES(1),(2.5),(3)) v(n)) SELECT SUM(n) FROM nums;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "6.5" ]] || false
}

@test 'values: multi-column partial cast' {
    # Only second column needs cast, first stays int
    run query_server -t -c "SELECT * FROM (VALUES(1, 10),(2, 20.5),(3, 30)) v(a, b);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "10" ]] || false
    [[ "$output" =~ "2" ]] || false
    [[ "$output" =~ "20.5" ]] || false
    [[ "$output" =~ "3" ]] || false
    [[ "$output" =~ "30" ]] || false
}

@test 'values: combined GROUP BY ORDER BY LIMIT' {
    run query_server -t -c "SELECT n, COUNT(*) as cnt FROM (VALUES(1),(2.5),(1),(2.5),(3),(1)) v(n) GROUP BY n ORDER BY cnt DESC LIMIT 2;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1" ]] || false
    [[ "$output" =~ "3" ]] || false
    [[ "$output" =~ "2.5" ]] || false
    [[ "$output" =~ "2" ]] || false
}

@test 'values: combined WHERE ORDER BY LIMIT' {
    run query_server -t -c "SELECT * FROM (VALUES(1),(2.5),(3),(4.5),(5)) v(n) WHERE n > 1 ORDER BY n DESC LIMIT 2;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "5" ]] || false
    [[ "$output" =~ "4.5" ]] || false
}
