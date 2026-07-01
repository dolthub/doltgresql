#include <apr.h>
#include <apr_pools.h>
#include <apr_dbd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void fail(const char *op, const char *msg) {
    fprintf(stderr, "%s: %s\n", op, msg ? msg : "(no message)");
    exit(1);
}

static void exec_query(const apr_dbd_driver_t *drv, apr_dbd_t *handle, const char *sql) {
    int nrows = 0;
    int rv = apr_dbd_query(drv, handle, &nrows, sql);
    if (rv != 0)
        fail(sql, apr_dbd_error(drv, handle, rv));
}

static const char *exec_scalar(const apr_dbd_driver_t *drv, apr_pool_t *pool,
                                apr_dbd_t *handle, const char *sql) {
    apr_dbd_results_t *res = NULL;
    int rv = apr_dbd_select(drv, pool, handle, &res, sql, 1);
    if (rv != 0)
        fail(sql, apr_dbd_error(drv, handle, rv));
    apr_dbd_row_t *row = NULL;
    if (apr_dbd_get_row(drv, pool, res, &row, -1) != 0)
        fail(sql, "no rows returned");
    return apr_dbd_get_entry(drv, row, 0);
}

int main(int argc, char *argv[]) {
    if (argc < 3) {
        fprintf(stderr, "Usage: %s <user> <port>\n", argv[0]);
        return 1;
    }
    const char *user = argv[1];
    const char *port = argv[2];

    apr_initialize();
    apr_pool_t *pool;
    apr_pool_create(&pool, NULL);

    apr_status_t rv = apr_dbd_init(pool);
    if (rv != APR_SUCCESS) {
        char buf[256];
        apr_strerror(rv, buf, sizeof(buf));
        fail("apr_dbd_init", buf);
    }

    const apr_dbd_driver_t *driver;
    rv = apr_dbd_get_driver(pool, "pgsql", &driver);
    if (rv != APR_SUCCESS) {
        char buf[256];
        apr_strerror(rv, buf, sizeof(buf));
        fail("apr_dbd_get_driver(pgsql)", buf);
    }

    char params[256];
    snprintf(params, sizeof(params),
             "host=localhost port=%s dbname=postgres user=%s password=password",
             port, user);
    apr_dbd_t *handle = NULL;
    const char *open_err = NULL;
    rv = apr_dbd_open_ex(driver, pool, params, &handle, &open_err);
    if (rv != APR_SUCCESS)
        fail("apr_dbd_open_ex", open_err);

    const char *pk_str = exec_scalar(driver, pool, handle, "SELECT pk FROM test_table LIMIT 1");
    if (!pk_str || atoi(pk_str) != 1) {
        fprintf(stderr, "expected pk=1, got %s\n", pk_str ? pk_str : "NULL");
        exit(1);
    }

    exec_query(driver, handle, "INSERT INTO test_table VALUES (2)");

    const char *count_str = exec_scalar(driver, pool, handle, "SELECT COUNT(*) FROM test_table");
    if (!count_str || atoi(count_str) != 2) {
        fprintf(stderr, "expected count=2, got %s\n", count_str ? count_str : "NULL");
        exit(1);
    }

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
    for (int i = 0; dolt_queries[i]; i++)
        exec_query(driver, handle, dolt_queries[i]);

    const char *log_count = exec_scalar(driver, pool, handle, "SELECT COUNT(*) FROM dolt_log");
    if (!log_count || atoi(log_count) != 4) {
        fprintf(stderr, "expected 4 dolt_log entries, got %s\n", log_count ? log_count : "NULL");
        exit(1);
    }

    const char *test_count = exec_scalar(driver, pool, handle, "SELECT COUNT(*) FROM test");
    if (!test_count || atoi(test_count) != 2) {
        fprintf(stderr, "expected 2 rows in test, got %s\n", test_count ? test_count : "NULL");
        exit(1);
    }

    apr_dbd_close(driver, handle);
    apr_pool_destroy(pool);
    apr_terminate();

    printf("libaprutil apr_dbd test passed\n");
    return 0;
}
