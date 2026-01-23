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

        print("\n✅ All tests passed.", flush=True)
        return 0

    except Exception as e:
        print("\n❌ Test failed.", flush=True)
        print(f"Error: {e}", flush=True)
        traceback.print_exc()
        return 1


if __name__ == "__main__":
    sys.exit(main())
