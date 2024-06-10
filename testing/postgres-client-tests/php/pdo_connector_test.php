<?php
    $user = $argv[1];
    $port = $argv[2];
    $db = 'doltgres';

    $conn = new PDO("pgsql:host=localhost;port={$port};dbname={$db}", $user, '');
    $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    $queries = [
        "create table test (pk int, value int, primary key(pk))" => 0,
        "insert into test (pk, value) values (0,0)" => 1,
        "select * from test" => 1,
        "call dolt_add('-A');" => 0,
        "call dolt_commit('-m', 'my commit')" => 0,
        "call dolt_checkout('-b', 'mybranch')" => 0,
        "insert into test (pk, value) values (1,1)" => 1,
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

    exit(0)
?>
