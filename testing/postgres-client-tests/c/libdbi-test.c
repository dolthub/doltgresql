#include <dbi/dbi.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void conn_fail(dbi_conn conn, const char *op) {
    const char *errstr = NULL;
    dbi_conn_error(conn, &errstr);
    fprintf(stderr, "%s: %s\n", op, errstr ? errstr : "(no message)");
    exit(1);
}

static dbi_result exec_query(dbi_conn conn, const char *sql) {
    dbi_result res = dbi_conn_query(conn, sql);
    if (!res)
        conn_fail(conn, sql);
    return res;
}

int main(int argc, char *argv[]) {
    if (argc < 3) {
        fprintf(stderr, "Usage: %s <user> <port>\n", argv[0]);
        return 1;
    }
    const char *user = argv[1];
    int port = atoi(argv[2]);

    dbi_inst dbi;
    if (dbi_initialize_r(NULL, &dbi) < 0) {
        fprintf(stderr, "dbi_initialize_r failed: no drivers found\n");
        return 1;
    }

    dbi_conn conn = dbi_conn_new_r("pgsql", dbi);
    if (!conn) {
        fprintf(stderr, "dbi_conn_new_r(pgsql) failed: driver not installed?\n");
        dbi_shutdown_r(dbi);
        return 1;
    }

    dbi_conn_set_option(conn, "host", "localhost");
    dbi_conn_set_option_numeric(conn, "port", (long)port);
    dbi_conn_set_option(conn, "dbname", "postgres");
    dbi_conn_set_option(conn, "username", user);
    dbi_conn_set_option(conn, "password", "password");

    if (dbi_conn_connect(conn) < 0)
        conn_fail(conn, "dbi_conn_connect");

    // SELECT pk from test_table (set up by bats setup())
    dbi_result res = exec_query(conn, "SELECT pk FROM test_table LIMIT 1");
    if (!dbi_result_next_row(res)) {
        fprintf(stderr, "expected at least one row in test_table\n");
        exit(1);
    }
    int pk = dbi_result_get_int(res, "pk");
    dbi_result_free(res);
    if (pk != 1) {
        fprintf(stderr, "expected pk=1, got %d\n", pk);
        exit(1);
    }

    // INSERT
    res = exec_query(conn, "INSERT INTO test_table VALUES (2)");
    dbi_result_free(res);

    // COUNT
    res = exec_query(conn, "SELECT COUNT(*) FROM test_table");
    if (!dbi_result_next_row(res)) {
        fprintf(stderr, "expected count row\n");
        exit(1);
    }
    long long count = dbi_result_get_longlong(res, "count");
    dbi_result_free(res);
    if (count != 2) {
        fprintf(stderr, "expected count=2, got %lld\n", count);
        exit(1);
    }

    // Dolt workflow
    const char *dolt_queries[] = {
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
        NULL
    };
    for (int i = 0; dolt_queries[i]; i++) {
        res = exec_query(conn, dolt_queries[i]);
        dbi_result_free(res);
    }

    res = exec_query(conn, "SELECT COUNT(*) FROM dolt_log");
    dbi_result_next_row(res);
    long long log_count = dbi_result_get_longlong(res, "count");
    dbi_result_free(res);
    if (log_count != 4) {
        fprintf(stderr, "expected 4 dolt_log entries, got %lld\n", log_count);
        exit(1);
    }

    res = exec_query(conn, "SELECT COUNT(*) FROM test");
    dbi_result_next_row(res);
    long long test_count = dbi_result_get_longlong(res, "count");
    dbi_result_free(res);
    if (test_count != 2) {
        fprintf(stderr, "expected 2 rows in test, got %lld\n", test_count);
        exit(1);
    }

    dbi_conn_close(conn);
    dbi_shutdown_r(dbi);

    printf("libdbi pgsql test passed\n");
    return 0;
}
