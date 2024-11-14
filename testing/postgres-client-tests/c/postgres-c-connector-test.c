#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <libpq-fe.h>

#define QUERIES_SIZE 13

char *queries[QUERIES_SIZE] = {
    "create table test (pk int, value int, d1 decimal(9, 3), f1 float, c1 char(10), t1 text, primary key(pk))",
    "select * from test",
    "insert into test (pk, value, d1, f1, c1, t1) values (0,0,0.0,0.0,'abc','a1')",
    "select * from test",
    "call dolt_add('-A');",
    "call dolt_commit('-m', 'my commit')",
    "select COUNT(*) FROM dolt.log",
    "call dolt_checkout('-b', 'mybranch')",
    "insert into test (pk, value, d1, f1, c1, t1) values (10,10, 123456.789, 420.42,'example','some text')",
    "call dolt_commit('-a', '-m', 'my commit2')",
    "call dolt_checkout('main')",
    "call dolt_merge('mybranch')",
    "select COUNT(*) FROM dolt.log",
};

int main(int argc, char *argv[]) {

    char* user = argv[1];
    int   port = atoi(argv[2]);

    // Connect to the database
    // conninfo is a string of keywords and values separated by spaces.
    char conninfo[100];
    sprintf(conninfo, "dbname=postgres user=%s password=password host=localhost port=%d", user, port);

    // Create a connection
    PGconn *conn = PQconnectdb(conninfo);

    // Check if the connection is successful
    if (PQstatus(conn) != CONNECTION_OK) {
        // If not successful, print the error message and finish the connection
        printf("Error while connecting to the database server: %s\n", PQerrorMessage(conn));

        // Finish the connection
        PQfinish(conn);

        // Exit the program
        exit(1);
    }

    // We have successfully established a connection to the database server
    printf("Connection Established\n");
    printf("Port: %s\n", PQport(conn));
    printf("Host: %s\n", PQhost(conn));
    printf("DBName: %s\n", PQdb(conn));

    for ( int i = 0; i < QUERIES_SIZE; i++ ) {
        // Submit the query and retrieve the result
        PGresult *res = PQexec(conn, queries[i]);

        // Check the status of the query result
        ExecStatusType resStatus = PQresultStatus(res);

        if (resStatus == PGRES_COMMAND_OK || resStatus == PGRES_TUPLES_OK) {
            // Successful completion of a command returning no data OR data (such as a SELECT or SHOW).

            // Clear the result
            PQclear(res);
        } else {
            printf("QUERY FAILED: %s\n", queries[i]);
            // If not successful, print the error message and finish the connection
            printf("Error while executing the query: %s\n", PQerrorMessage(conn));

            // Clear the result
            PQclear(res);

            // Finish the connection
            PQfinish(conn);

            // Exit the program
            exit(1);
        }
    }

    // Submit the query and retrieve the result
    PGresult *res = PQexec(conn, "SELECT * FROM test WHERE pk = 10");

    // Check the status of the query result
    ExecStatusType resStatus = PQresultStatus(res);

    if (resStatus != PGRES_TUPLES_OK) {
        printf("QUERY FAILED: %s\n", "SELECT * FROM test WHERE pk = 10");
        // If not successful, print the error message and finish the connection
        printf("Error while executing the query: %s\n", PQerrorMessage(conn));

        // Clear the result
        PQclear(res);

        // Finish the connection
        PQfinish(conn);

        // Exit the program
        exit(1);
    }


    // Get the number of columns in the query result
    int cols = PQnfields(res);
    printf("Number of cols: %d\n", cols);
    assert(cols == 6);

    char *expectedCols[6] = {"pk", "value", "d1", "f1", "c1", "t1"};
    // Assert the column names
    for (int i = 0; i < cols; i++) {
        assert(strcmp(PQfname(res, i), expectedCols[i]) == 0);
    }

    // Get the number of rows in the query result
    int rows = PQntuples(res);
    printf("Number of rows: %d\n", rows);
    assert(rows == 1);

    char *expectedRowResults[6] = {"10", "10", "123456.789", "420.42", "example   ", "some text"};
    // Assert query result
    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < cols; j++) {
            char *actual = PQgetvalue(res, i, j);
            printf("EXPECTED: '%s'\n", expectedRowResults[j]);
            printf("ACTUAL: '%s'\n", actual);
            assert(strcmp(actual, expectedRowResults[j]) == 0);
        }
    }

    // Clear the result
    PQclear(res);

    // Close the connection and free the memory
    PQfinish(conn);

    return 0;
}
