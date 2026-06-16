#include <sql.h>
#include <sqlext.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void die(SQLSMALLINT type, SQLHANDLE handle, const char *op) {
    SQLCHAR state[6], msg[256];
    SQLINTEGER native;
    SQLSMALLINT len;
    SQLGetDiagRec(type, handle, 1, state, &native, msg, sizeof(msg), &len);
    fprintf(stderr, "%s failed: %s (state=%s)\n", op, msg, state);
    exit(1);
}

#define CHECK_DBC(ret, h, op)  do { if ((ret) != SQL_SUCCESS && (ret) != SQL_SUCCESS_WITH_INFO) die(SQL_HANDLE_DBC,  (h), (op)); } while (0)
#define CHECK_STMT(ret, h, op) do { if ((ret) != SQL_SUCCESS && (ret) != SQL_SUCCESS_WITH_INFO) die(SQL_HANDLE_STMT, (h), (op)); } while (0)

static void exec_query(SQLHDBC dbc, const char *sql) {
    SQLHSTMT stmt;
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    SQLRETURN ret = SQLExecDirect(stmt, (SQLCHAR *)sql, SQL_NTS);
    CHECK_STMT(ret, stmt, sql);
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
}

static long long fetch_count(SQLHDBC dbc, const char *sql) {
    SQLHSTMT stmt;
    SQLCHAR buf[32];
    SQLLEN indicator;
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    SQLRETURN ret = SQLExecDirect(stmt, (SQLCHAR *)sql, SQL_NTS);
    CHECK_STMT(ret, stmt, sql);
    SQLFetch(stmt);
    SQLGetData(stmt, 1, SQL_C_CHAR, buf, sizeof(buf), &indicator);
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
    return atoll((char *)buf);
}

int main(int argc, char *argv[]) {
    if (argc < 3) {
        fprintf(stderr, "Usage: %s <user> <port>\n", argv[0]);
        return 1;
    }
    const char *user = argv[1];
    const char *port = argv[2];

    SQLHENV env;
    SQLHDBC dbc;
    SQLRETURN ret;

    SQLAllocHandle(SQL_HANDLE_ENV, SQL_NULL_HANDLE, &env);
    SQLSetEnvAttr(env, SQL_ATTR_ODBC_VERSION, (SQLPOINTER)SQL_OV_ODBC3, 0);
    SQLAllocHandle(SQL_HANDLE_DBC, env, &dbc);

    char connStr[512];
    snprintf(connStr, sizeof(connStr),
        "Driver={PostgreSQL Unicode};Server=localhost;Port=%s;Database=postgres;UID=%s;PWD=password;",
        port, user);

    SQLCHAR outStr[1024];
    SQLSMALLINT outLen;
    ret = SQLDriverConnect(dbc, NULL, (SQLCHAR *)connStr, SQL_NTS,
                           outStr, sizeof(outStr), &outLen, SQL_DRIVER_NOPROMPT);
    CHECK_DBC(ret, dbc, "SQLDriverConnect");

    // SELECT from test_table (set up by bats setup())
    {
        SQLHSTMT stmt;
        SQLINTEGER pk;
        SQLLEN indicator;
        SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
        ret = SQLExecDirect(stmt, (SQLCHAR *)"SELECT pk FROM test_table LIMIT 1", SQL_NTS);
        CHECK_STMT(ret, stmt, "SELECT pk FROM test_table LIMIT 1");
        SQLFetch(stmt);
        SQLGetData(stmt, 1, SQL_C_SLONG, &pk, sizeof(pk), &indicator);
        if (pk != 1) {
            fprintf(stderr, "expected pk=1, got %d\n", pk);
            return 1;
        }
        SQLFreeHandle(SQL_HANDLE_STMT, stmt);
    }

    exec_query(dbc, "INSERT INTO test_table VALUES (2)");

    long long count = fetch_count(dbc, "SELECT COUNT(*) FROM test_table");
    if (count != 2) {
        fprintf(stderr, "expected count=2, got %lld\n", count);
        return 1;
    }

    // Dolt workflow: create table, insert, commit, branch, insert, commit, merge
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
        exec_query(dbc, dolt_queries[i]);

    count = fetch_count(dbc, "SELECT COUNT(*) FROM dolt_log");
    if (count != 4) {
        fprintf(stderr, "expected 4 dolt_log entries, got %lld\n", count);
        return 1;
    }

    count = fetch_count(dbc, "SELECT COUNT(*) FROM test");
    if (count != 2) {
        fprintf(stderr, "expected 2 rows in test, got %lld\n", count);
        return 1;
    }

    SQLDisconnect(dbc);
    SQLFreeHandle(SQL_HANDLE_DBC, dbc);
    SQLFreeHandle(SQL_HANDLE_ENV, env);

    printf("psqlODBC test passed\n");
    return 0;
}
