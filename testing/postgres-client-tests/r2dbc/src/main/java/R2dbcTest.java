import io.r2dbc.postgresql.PostgresqlConnectionConfiguration;
import io.r2dbc.postgresql.PostgresqlConnectionFactory;
import io.r2dbc.postgresql.client.SSLMode;
import io.r2dbc.spi.Connection;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

public class R2dbcTest {

    public static void main(String[] args) {
        String user = args[0];
        int port = Integer.parseInt(args[1]);

        PostgresqlConnectionFactory factory = new PostgresqlConnectionFactory(
            PostgresqlConnectionConfiguration.builder()
                .host("localhost")
                .port(port)
                .database("postgres")
                .username(user)
                .password("password")
                .sslMode(SSLMode.DISABLE)
                .build()
        );

        try {
            Mono.usingWhen(
                factory.create(),
                R2dbcTest::runTests,
                Connection::close
            ).block();
            System.out.println("r2dbc test passed");
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
            System.exit(1);
        }
    }

    // Execute a statement and discard all results (works for both DML and SELECT).
    static Mono<Void> exec(Connection conn, String sql) {
        return Flux.from(conn.createStatement(sql).execute())
            .flatMap(r -> Flux.from(r.map((row, meta) -> 0)))
            .then();
    }

    static <T> Mono<T> queryOne(Connection conn, String sql, Class<T> type) {
        return Mono.from(conn.createStatement(sql).execute())
            .flatMap(r -> Mono.from(r.map((row, meta) -> row.get(0, type))));
    }

    static Mono<Void> runTests(Connection conn) {
        return queryOne(conn, "SELECT pk FROM test_table LIMIT 1", Integer.class)
            .doOnNext(pk -> {
                if (pk != 1) throw new RuntimeException("expected pk=1, got " + pk);
            })
            .then(exec(conn, "INSERT INTO test_table VALUES (2)"))
            .then(queryOne(conn, "SELECT COUNT(*) FROM test_table", Long.class))
            .doOnNext(n -> {
                if (n.longValue() != 2L)
                    throw new RuntimeException("expected count=2, got " + n);
            })
            .thenMany(Flux.fromArray(new String[]{
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
            }).concatMap(sql -> exec(conn, sql)))
            .then(queryOne(conn, "SELECT COUNT(*) FROM dolt_log", Long.class))
            .doOnNext(n -> {
                if (n.longValue() != 4L)
                    throw new RuntimeException("expected 4 dolt_log entries, got " + n);
            })
            .then(queryOne(conn, "SELECT COUNT(*) FROM test", Long.class))
            .doOnNext(n -> {
                if (n.longValue() != 2L)
                    throw new RuntimeException("expected 2 rows in test, got " + n);
            })
            .then();
    }
}
