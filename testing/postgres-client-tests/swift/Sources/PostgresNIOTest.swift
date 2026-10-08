import Logging
import NIOPosix
import PostgresNIO

struct TestError: Error, CustomStringConvertible {
    let description: String
    init(_ msg: String) { description = msg }
}

@main
struct PostgresNIOTest {
    static func main() async {
        let args = CommandLine.arguments
        guard args.count >= 3 else {
            fputs("Usage: postgresnio-test <user> <port>\n", stderr)
            exit(1)
        }
        let user = args[1]
        let port = Int(args[2])!
        do {
            try await runTests(user: user, port: port)
            print("PostgresNIO test passed")
        } catch {
            fputs("Error: \(error)\n", stderr)
            exit(1)
        }
    }

    // Execute a statement and drain any result rows.
    static func exec(_ conn: PostgresConnection, _ sql: String, logger: Logger) async throws {
        for try await _ in try await conn.query(PostgresQuery(unsafeSQL: sql), logger: logger) {}
    }

    // Execute a single-row query and return the first row.
    static func queryRow(_ conn: PostgresConnection, _ sql: String, logger: Logger) async throws -> PostgresRow {
        for try await row in try await conn.query(PostgresQuery(unsafeSQL: sql), logger: logger) {
            return row
        }
        throw TestError("expected at least one row for: \(sql)")
    }

    static func runTests(user: String, port: Int) async throws {
        let eventLoopGroup = MultiThreadedEventLoopGroup(numberOfThreads: 1)
        defer { try! eventLoopGroup.syncShutdownGracefully() }

        var logger = Logger(label: "postgres-test")
        logger.logLevel = .critical

        let config = PostgresConnection.Configuration(
            host: "localhost",
            port: port,
            username: user,
            password: "password",
            database: "postgres",
            tls: .disable
        )
        let conn = try await PostgresConnection.connect(
            on: eventLoopGroup.next(),
            configuration: config,
            id: 1,
            logger: logger
        )
        defer { try! conn.close().wait() }

        // SELECT pk from test_table (set up by bats setup())
        let pkRow = try await queryRow(conn, "SELECT pk FROM test_table LIMIT 1", logger: logger)
        let pk = try pkRow.decode(Int32.self, context: .default)
        guard pk == 1 else { throw TestError("expected pk=1, got \(pk)") }

        // INSERT
        try await exec(conn, "INSERT INTO test_table VALUES (2)", logger: logger)

        // COUNT
        let countRow = try await queryRow(conn, "SELECT COUNT(*) FROM test_table", logger: logger)
        let count = try countRow.decode(Int64.self, context: .default)
        guard count == 2 else { throw TestError("expected count=2, got \(count)") }

        // Dolt workflow: create table, insert, commit, branch, insert, commit, merge
        for q in [
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
        ] {
            try await exec(conn, q, logger: logger)
        }

        let logRow = try await queryRow(conn, "SELECT COUNT(*) FROM dolt_log", logger: logger)
        let logCount = try logRow.decode(Int64.self, context: .default)
        guard logCount == 4 else { throw TestError("expected 4 dolt_log entries, got \(logCount)") }

        let testRow = try await queryRow(conn, "SELECT COUNT(*) FROM test", logger: logger)
        let testCount = try testRow.decode(Int64.self, context: .default)
        guard testCount == 2 else { throw TestError("expected 2 rows in test, got \(testCount)") }
    }
}
