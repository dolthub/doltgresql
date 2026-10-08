#include <iostream>
#include <stdexcept>
#include <string>
#include <pqxx/pqxx>

int main(int argc, char *argv[]) {
    if (argc < 3) {
        std::cerr << "Usage: " << argv[0] << " <user> <port>\n";
        return 1;
    }
    std::string user = argv[1];
    std::string port = argv[2];

    try {
        std::string connStr = "host=localhost port=" + port + " dbname=postgres user=" + user + " password=password sslmode=disable";
        pqxx::connection conn(connStr);
        pqxx::nontransaction ntxn(conn);

        // SELECT from test_table (set up by bats setup())
        pqxx::result r = ntxn.exec("SELECT pk FROM test_table LIMIT 1");
        if (r.empty() || r[0][0].as<int>() != 1)
            throw std::runtime_error("expected pk=1");

        // INSERT
        ntxn.exec("INSERT INTO test_table VALUES (2)");

        // COUNT
        r = ntxn.exec("SELECT COUNT(*) FROM test_table");
        if (r[0][0].as<long long>() != 2)
            throw std::runtime_error("expected count=2, got " + std::to_string(r[0][0].as<long long>()));

        // Prepared statement
        conn.prepare("select_pk", "SELECT pk FROM test_table WHERE pk = $1");
        r = ntxn.exec_prepared("select_pk", 1);
        if (r.empty() || r[0][0].as<int>() != 1)
            throw std::runtime_error("expected pk=1 from prepared stmt");

        // Dolt workflow: create table, insert, commit, branch, insert, commit, merge
        for (const char *q : {
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
        }) {
            ntxn.exec(q);
        }

        conn.prepare("select_test_by_pk", "SELECT pk, value FROM test WHERE pk = $1");
        r = ntxn.exec_prepared("select_test_by_pk", 0);
        if (r.empty())
            throw std::runtime_error("no rows for select_test_by_pk");
        if (r[0]["pk"].as<int>() != 0 || r[0]["value"].as<int>() != 0)
            throw std::runtime_error("expected pk=0 value=0");

        conn.prepare("count_dolt_log", "SELECT COUNT(*) FROM dolt_log");
        r = ntxn.exec_prepared("count_dolt_log");
        if (r[0][0].as<long long>() != 4)
            throw std::runtime_error("expected 4 dolt_log entries, got " + std::to_string(r[0][0].as<long long>()));

        conn.prepare("count_test", "SELECT COUNT(*) FROM test");
        r = ntxn.exec_prepared("count_test");
        if (r[0][0].as<long long>() != 2)
            throw std::runtime_error("expected 2 rows in test, got " + std::to_string(r[0][0].as<long long>()));

        std::cout << "libpqxx test passed\n";
    } catch (const std::exception &e) {
        std::cerr << "Error: " << e.what() << "\n";
        return 1;
    }
    return 0;
}
