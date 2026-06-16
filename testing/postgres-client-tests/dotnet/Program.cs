using Npgsql;

var user = args[0];
var port = args[1];

var connStr = $"Host=localhost;Port={port};Username={user};Password=password;Database=postgres;SSL Mode=Disable";
await using var conn = new NpgsqlConnection(connStr);
await conn.OpenAsync();

// Basic SELECT
await using (var cmd = new NpgsqlCommand("SELECT pk FROM test_table LIMIT 1", conn))
{
    var pk = (int)(await cmd.ExecuteScalarAsync())!;
    if (pk != 1)
        throw new Exception($"expected pk=1, got {pk}");
}

// INSERT
await using (var cmd = new NpgsqlCommand("INSERT INTO test_table VALUES (2)", conn))
    await cmd.ExecuteNonQueryAsync();

// COUNT
await using (var cmd = new NpgsqlCommand("SELECT COUNT(*) FROM test_table", conn))
{
    var count = (long)(await cmd.ExecuteScalarAsync())!;
    if (count != 2)
        throw new Exception($"expected count=2, got {count}");
}

// Prepared SELECT
await using (var cmd = new NpgsqlCommand("SELECT pk FROM test_table WHERE pk = $1", conn))
{
    cmd.Parameters.AddWithValue(1);
    await cmd.PrepareAsync();
    var pk = (int)(await cmd.ExecuteScalarAsync())!;
    if (pk != 1)
        throw new Exception($"expected pk=1 from prepared stmt, got {pk}");
}

// Dolt workflow: create table, insert, commit, branch, insert, commit, merge
foreach (var q in new[]
{
    "DROP TABLE IF EXISTS test",
    "CREATE TABLE test (pk int, value int, PRIMARY KEY(pk))",
    "INSERT INTO test (pk, value) VALUES (0, 0)",
    "SELECT dolt_add('-A')",
    "SELECT dolt_commit('-m', 'added table test')",
    "SELECT dolt_checkout('-b', 'mybranch')",
    "INSERT INTO test VALUES (1, 1)",
    "SELECT dolt_commit('-a', '-m', 'updated test')",
    "SELECT dolt_checkout('main')",
    "SELECT dolt_merge('mybranch')",
})
{
    await using var cmd = new NpgsqlCommand(q, conn);
    await cmd.ExecuteNonQueryAsync();
}

await RunPreparedQuery(
    "SELECT pk, value FROM test WHERE pk = $1",
    [0],
    async r =>
    {
        if (!await r.ReadAsync()) throw new Exception("no rows");
        var pk = r.GetInt32(0);
        var value = r.GetInt32(1);
        if (pk != 0 || value != 0)
            throw new Exception($"expected pk=0 value=0, got pk={pk} value={value}");
    });

await RunPreparedQuery(
    "SELECT COUNT(*) FROM dolt_log",
    [],
    async r =>
    {
        if (!await r.ReadAsync()) throw new Exception("no rows");
        var size = r.GetInt64(0);
        if (size != 4)
            throw new Exception($"expected 4 dolt_log entries, got {size}");
    });

await RunPreparedQuery(
    "SELECT COUNT(*) FROM test",
    [],
    async r =>
    {
        if (!await r.ReadAsync()) throw new Exception("no rows");
        var size = r.GetInt64(0);
        if (size != 2)
            throw new Exception($"expected 2 rows in test, got {size}");
    });

Console.WriteLine("Npgsql test passed");

async Task RunPreparedQuery(string query, object[] queryArgs, Func<NpgsqlDataReader, Task> check)
{
    await using var cmd = new NpgsqlCommand(query, conn);
    foreach (var arg in queryArgs)
        cmd.Parameters.AddWithValue(arg);
    await cmd.PrepareAsync();
    await using var reader = await cmd.ExecuteReaderAsync();
    await check(reader);
}
