#!/usr/bin/env python3
import os
import sys
import traceback
import psycopg2

# ---------------------------------------------------------------------------
# Query list (kept at top for consistency with other tests)
# ---------------------------------------------------------------------------

TEST_QUERIES = [
    "DROP TABLE IF EXISTS test",
    "create table test (pk int, value int, d1 decimal(9, 3), f1 float, c1 char(10), t1 text, primary key(pk))",
    "select * from test",
    "insert into test (pk, value, d1, f1, c1, t1) values (0,0,0.0,0.0,'abc','a1')",
    "select * from test",
    "select dolt_add('-A');",
    "select dolt_commit('-m', 'my commit')",
    "select COUNT(*) FROM dolt.log",
    "select dolt_checkout('-b', 'mybranch')",
    "insert into test (pk, value, d1, f1, c1, t1) values (10,10, 123456.789, 420.42,'example','some text')",
    "select dolt_commit('-a', '-m', 'my commit2')",
    "select dolt_checkout('main')",
    "select dolt_merge('mybranch')",
    "select COUNT(*) FROM dolt.log",
]

# ---------------------------------------------------------------------------

def env(name, default=None):
    return os.getenv(name, default)


def connect(user: str, port: int):
    conn = psycopg2.connect(
        host=env("PGHOST", "localhost"),
        port=port,
        dbname="postgres",
        user=user,
        password=env("PGPASSWORD", "password"),
        connect_timeout=int(env("PGCONNECT_TIMEOUT", "10")),
        sslmode=env("PGSSLMODE"),
    )
    conn.autocommit = True
    return conn


def run(cur, q):
    print(f"SQL> {q}", flush=True)
    cur.execute(q)
    if cur.description is not None:
        cur.fetchall()  # drain result set

# load_test creates a table with |n_rows| and asserts that all rows are correctly returned.
def load_test(cur, n_rows=1000):
    print("\n=== Part 1: Load test ===", flush=True)

    rows = max(1000, int(n_rows))

    run(cur, "DROP TABLE IF EXISTS load_test")
    run(cur, "CREATE TABLE load_test (id INT PRIMARY KEY, val INT NOT NULL)")

    data = [(i, i * 10) for i in range(rows)]
    cur.executemany(
        "INSERT INTO load_test (id, val) VALUES (%s, %s)",
        data,
    )

    cur.execute("SELECT COUNT(*) FROM load_test")
    cnt = cur.fetchone()[0]
    if cnt != rows:
        raise AssertionError(f"COUNT(*) mismatch: expected {rows}, got {cnt}")

    cur.execute("SELECT id FROM load_test ORDER BY id")
    got = cur.fetchall()
    if len(got) != rows:
        raise AssertionError(f"fetchall mismatch: expected {rows}, got {len(got)}")

    print(f"Inserted and selected {rows} rows OK.", flush=True)


def compliance_test(cur):
    print("\n=== Part 2: Test Queries ===", flush=True)
    for q in TEST_QUERIES:
        run(cur, q)
    print("Compliance queries executed OK.", flush=True)


# error_code_test asserts that errors arrive with the correct SQLSTATE and are interpreted
# correctly by the client. psycopg2 constructs its exception classes from the SQLSTATE the
# server reports (psycopg2.errors), so catching the specific class checks both at once.
def error_code_test(user, port):
    print("\n=== Part 3: Error codes ===", flush=True)
    from psycopg2 import errors

    conn = connect(user, port)
    cur = conn.cursor()
    run(cur, "DROP TABLE IF EXISTS err_test")
    run(cur, "CREATE TABLE err_test (id INT PRIMARY KEY, val INT NOT NULL)")
    run(cur, "INSERT INTO err_test VALUES (1, 1)")

    cases = [
        ("INSERT INTO err_test VALUES (1, 1)", errors.UniqueViolation, "23505"),
        ("INSERT INTO err_test VALUES (2, NULL)", errors.NotNullViolation, "23502"),
        ("SELECT * FROM no_such_table", errors.UndefinedTable, "42P01"),
        ("SELECT no_such_col FROM err_test", errors.UndefinedColumn, "42703"),
        ("SELECT 1/0", errors.DivisionByZero, "22012"),
        ("SELEC 1", errors.SyntaxError, "42601"),
        ("SELECT 'abc'::int4", errors.InvalidTextRepresentation, "22P02"),
    ]
    for sql, exc_class, code in cases:
        print(f"SQL> {sql}  (expect {exc_class.__name__} / {code})", flush=True)
        try:
            cur.execute(sql)
        except exc_class as e:
            if e.pgcode != code:
                raise AssertionError(f"{sql}: expected pgcode {code}, got {e.pgcode}")
        else:
            raise AssertionError(f"{sql}: expected {exc_class.__name__}, but query succeeded")

    # the connection must have survived every error above
    cur.execute("SELECT COUNT(*) FROM err_test")
    if cur.fetchone()[0] != 1:
        raise AssertionError("unexpected row count after failed statements")

    # Conflicting concurrent transactions: the losing commit must raise
    # SerializationFailure (SQLSTATE 40001), which retry logic dispatches on.
    conn_a = connect(user, port)
    conn_a.autocommit = False
    conn_b = conn  # autocommit, plays the concurrent writer
    cur_a = conn_a.cursor()
    cur_a.execute("UPDATE err_test SET val = 10 WHERE id = 1")
    cur.execute("UPDATE err_test SET val = 20 WHERE id = 1")
    try:
        conn_a.commit()
        raise AssertionError("expected SerializationFailure on conflicting commit")
    except errors.SerializationFailure as e:
        if e.pgcode != "40001":
            raise AssertionError(f"expected pgcode 40001, got {e.pgcode}")

    # ... and the session must be usable again immediately
    conn_a.rollback()
    cur_a.execute("SELECT COUNT(*) FROM err_test")
    if cur_a.fetchone()[0] != 1:
        raise AssertionError("connection unusable after serialization failure")
    conn_a.close()

    print("Error code tests OK.", flush=True)


def main():
    if len(sys.argv) != 3:
        print("Usage: python3 psycopg2_test.py <user> <port>")
        return 2

    user = sys.argv[1]
    port = int(sys.argv[2])
    load_rows = int(env("LOAD_ROWS", "1000"))

    try:
        with connect(user, port) as conn:
            with conn.cursor() as cur:
                load_test(cur, load_rows)
                compliance_test(cur)

        error_code_test(user, port)

        print("\n✅ All tests passed.", flush=True)
        return 0

    except Exception as e:
        print("\n❌ Test failed.", flush=True)
        print(f"Error: {e}", flush=True)
        traceback.print_exc()
        return 1


if __name__ == "__main__":
    sys.exit(main())
