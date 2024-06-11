<?php
    $user = $argv[1];
    $port = $argv[2];
    $db = 'doltgres';

    $conn = pg_connect("host = localhost port = $port dbname = $db user = $user")
    or die('Could not connect: ' . pg_result_error());

    $queries = [
        "create table test (pk int, value int, d1 decimal(9, 3), f1 float, primary key(pk))" => 0,
        "insert into test (pk, value, d1, f1) values (0,0,0.0,0.0)" => 0,
        "select * from test" => 1,
        "call dolt_add('-A');" => 0,
        "call dolt_commit('-m', 'my commit')" => 0,
        "call dolt_checkout('-b', 'mybranch')" => 0,
        "insert into test (pk, value, d1, f1) values (1,1, 123456.789, 420.42)" => 0,
        "call dolt_commit('-a', '-m', 'my commit2')" => 0,
        "call dolt_checkout('main')" => 0,
        "call dolt_merge('mybranch')" => 0,
        "select COUNT(*) FROM dolt_log" => 1
    ];

    foreach ($queries as $query => $expected) {
        $result = pg_query($conn, $query);
        if (is_bool($result)) {
            if (!$result) {
                echo "LENGTH: {pg_num_rows($result)}\n";
                echo "QUERY: {$query}\n";
                echo "EXPECTED: {$expected}\n";
                echo "RESULT: {$result}";
                exit(1);
            }
        } else if (pg_num_rows($result) != $expected) {
            echo "LENGTH: {pg_num_rows($result)}\n";
            echo "QUERY: {$query}\n";
            echo "EXPECTED: {$expected}\n";
            echo "RESULT: {$result}";
            exit(1);
        }
    }

    $result = pg_query($conn, "SELECT * FROM test WHERE pk = 1");
    assert(1 == pg_num_rows($result));
    while($row = pg_fetch_assoc($result)) {
        assert(1 == $row['pk']);
        assert(1 == $row['value']);
        assert(123456.789 == $row['d1']);
        assert(420.42 == $row['f1']);
    }

    pg_close($conn);

    exit(0)
?>
