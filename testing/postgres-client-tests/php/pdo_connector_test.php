<?php
    $user = $argv[1];
    $port = $argv[2];
    $db = 'doltgres';

    $conn = new PDO("pgsql:host=localhost;port={$port};dbname={$db}", $user, '');
    $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    $queries = [
        "create table test (pk int, value int, d1 decimal(9, 3), f1 float, c1 char(10), t1 text, primary key(pk))" => 0,
        "insert into test (pk, value, d1, f1, c1, t1) values (0,0,0.0,0.0,'abc','a1')" => 1,
        "select * from test" => 1,
        "call dolt_add('-A');" => 0,
        "call dolt_commit('-m', 'my commit')" => 0,
        "call dolt_checkout('-b', 'mybranch')" => 0,
        "insert into test (pk, value, d1, f1, c1, t1) values (1,1, 123456.789, 420.42,'example','some text')" => 1,
        "call dolt_commit('-a', '-m', 'my commit2')" => 0,
        "call dolt_checkout('main')" => 0,
        "call dolt_merge('mybranch')" => 0,
        "select COUNT(*) FROM dolt_log" => 1
    ];

    foreach ($queries as $query => $expected) {
        $result = $conn->query($query);
        if ($result->rowCount() != $expected) {
            echo "LENGTH: {$result->rowCount()}\n";
            echo "QUERY: {$query}\n";
            echo "EXPECTED: {$expected}\n";
            echo "RESULT: {$result}";
            exit(1);
        }
    }

    $result = $conn->query("SELECT * FROM test WHERE pk = 1");
    assert(1 == $result->rowCount());
    while($row = $result->fetch(PDO::FETCH_ASSOC)) {
        assert(1 == $row['pk']);
        assert(1 == $row['value']);
        assert(123456.789 == $row['d1']);
        assert(420.42 == $row['f1']);
        assert("example   " == $row['c1']);
        assert("some text" == $row['t1']);
    }

    exit(0)
?>
