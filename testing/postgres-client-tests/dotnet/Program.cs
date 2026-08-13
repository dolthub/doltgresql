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

// Error codes: Npgsql exposes the SQLSTATE as PostgresException.SqlState, and treats
// XX-class (internal error) codes as critical failures that break the connection. These
// checks confirm both that the correct codes arrive and that the connection stays open
// after each error, which is what makes ORM error handling (EF Core et al.) work.
foreach (var (sql, wantCode, label) in new[]
{
    ("INSERT INTO test (pk, value) VALUES (0, 0)", "23505", "unique violation"),
    ("SELECT * FROM no_such_table", "42P01", "undefined table"),
    ("SELECT 1/0", "22012", "division by zero"),
    ("SELEC 1", "42601", "syntax error"),
    ("SELECT 'abc'::int4", "22P02", "invalid text representation"),
})
{
    try
    {
        await using var cmd = new NpgsqlCommand(sql, conn);
        await cmd.ExecuteNonQueryAsync();
        throw new Exception($"{label}: expected SQLSTATE {wantCode}, but the query succeeded");
    }
    catch (PostgresException e)
    {
        if (e.SqlState != wantCode)
            throw new Exception($"{label}: expected SQLSTATE {wantCode}, got {e.SqlState} ({e.MessageText})");
    }
    if (conn.State != System.Data.ConnectionState.Open)
        throw new Exception($"{label}: connection did not survive the error");
}

// A failed statement inside a driver-API transaction must be rollback-able, with the
// connection intact — the mechanism ORMs rely on for atomic writes.
await using (var tx = await conn.BeginTransactionAsync())
{
    await using (var cmd = new NpgsqlCommand("INSERT INTO test (pk, value) VALUES (100, 100)", conn))
        await cmd.ExecuteNonQueryAsync();
    try
    {
        await using var cmd = new NpgsqlCommand("INSERT INTO test (pk, value) VALUES (100, 100)", conn);
        await cmd.ExecuteNonQueryAsync();
        throw new Exception("expected duplicate key error inside transaction");
    }
    catch (PostgresException e)
    {
        if (e.SqlState != "23505")
            throw new Exception($"expected SQLSTATE 23505 inside transaction, got {e.SqlState}");
    }
    await tx.RollbackAsync();
}
await using (var cmd = new NpgsqlCommand("SELECT COUNT(*) FROM test WHERE pk = 100", conn))
{
    var count = (long)(await cmd.ExecuteScalarAsync())!;
    if (count != 0)
        throw new Exception($"expected rolled-back insert to leave 0 rows, got {count}");
}

// Conflicting concurrent transactions: the losing COMMIT must report 40001
// (serialization_failure), the code retry logic dispatches on.
await using (var conn2 = new NpgsqlConnection(connStr))
{
    await conn2.OpenAsync();
    await using var tx = await conn.BeginTransactionAsync();
    await using (var cmd = new NpgsqlCommand("UPDATE test SET value = 10 WHERE pk = 0", conn))
        await cmd.ExecuteNonQueryAsync();
    await using (var cmd = new NpgsqlCommand("UPDATE test SET value = 20 WHERE pk = 0", conn2))
        await cmd.ExecuteNonQueryAsync();
    try
    {
        await tx.CommitAsync();
        throw new Exception("expected serialization failure on conflicting commit");
    }
    catch (PostgresException e)
    {
        if (e.SqlState != "40001")
            throw new Exception($"expected SQLSTATE 40001 on conflicting commit, got {e.SqlState}");
    }
}

// ... and the session must be usable again immediately.
await using (var cmd = new NpgsqlCommand("SELECT COUNT(*) FROM test", conn))
{
    var count = (long)(await cmd.ExecuteScalarAsync())!;
    if (count != 2)
        throw new Exception($"expected 2 rows after recovery, got {count}");
}

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
