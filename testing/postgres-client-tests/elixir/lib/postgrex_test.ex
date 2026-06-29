defmodule PostgrexTest do
  def main(args) do
    [user, port_str] = args
    port = String.to_integer(port_str)

    {:ok, _} = Application.ensure_all_started(:postgrex)

    {:ok, conn} = Postgrex.start_link(
      hostname: "localhost",
      port: port,
      database: "postgres",
      username: user,
      password: "password",
      ssl: false
    )

    {:ok, %{rows: [[pk]]}} = Postgrex.query(conn, "SELECT pk FROM test_table LIMIT 1", [])
    if pk != 1, do: raise("expected pk=1, got #{pk}")

    {:ok, _} = Postgrex.query(conn, "INSERT INTO test_table VALUES (2)", [])

    {:ok, %{rows: [[count]]}} = Postgrex.query(conn, "SELECT COUNT(*) FROM test_table", [])
    if count != 2, do: raise("expected count=2, got #{count}")

    # Dolt workflow: create table, insert, commit, branch, insert, commit, merge
    Enum.each([
      "DROP TABLE IF EXISTS test",
      "CREATE TABLE test (pk int, value int, PRIMARY KEY(pk))",
      "INSERT INTO test (pk, value) VALUES (0, 0)",
      "SELECT dolt_add('-A')",
      "SELECT dolt_commit('-m', 'added table test')",
      "SELECT dolt_checkout('-b', 'mybranch')",
      "INSERT INTO test VALUES (1, 1)",
      "SELECT dolt_commit('-a', '-m', 'updated test')",
      "SELECT dolt_checkout('main')",
      "SELECT dolt_merge('mybranch')"
    ], fn q -> {:ok, _} = Postgrex.query(conn, q, []) end)

    {:ok, %{rows: [[log_count]]}} = Postgrex.query(conn, "SELECT COUNT(*) FROM dolt_log", [])
    if log_count != 4, do: raise("expected 4 dolt_log entries, got #{log_count}")

    {:ok, %{rows: [[test_count]]}} = Postgrex.query(conn, "SELECT COUNT(*) FROM test", [])
    if test_count != 2, do: raise("expected 2 rows in test, got #{test_count}")

    IO.puts("Postgrex test passed")
  end
end
