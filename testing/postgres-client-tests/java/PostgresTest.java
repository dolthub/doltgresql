import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.sql.Statement;
import java.sql.ResultSet;

public class PostgresTest {
    // test queries to be run against Doltgres
    private static final String[] queries = {
        "create table test (pk int, value int, primary key(pk))",
        // "describe test",
        "select * from test",
        "insert into test (pk, value) values (0,0)",
        "select pk from test", // can make 'pk', 'test.pk' instead for more test coverage
        "call dolt_add('-A')",
        "call dolt_commit('-m', 'my commit')",
        "select COUNT(*) FROM dolt_log",
        "call dolt_checkout('-b', 'mybranch')",
        "insert into test (pk, value) values (1,1)",
        "call dolt_commit('-a', '-m', 'my commit2')",
        "call dolt_checkout('main')",
        "call dolt_merge('mybranch')",
        "select COUNT(*) FROM dolt_log",
};

// We currently only test a single field value in the first row
private static final String[] expectedResults = {
        "0",
        // "pk",
        null,
        "1",
        "0",
        "0",
        "0",
        "2",
        "0",
        "1",
        "0",
        "0",
        "0",
        "3"
};

// fieldAccessors are the value used to access a field in a row in a result set. Currently, only
// String (i.e column name) and Integer (i.e. field position) values are supported.
private static final Object[] fieldAccessors = {
        1,
        // 1,
        "pk",
        1,
        "pk",
        1,
        1,
        1,
        1,
        1,
        1,
        1,
        1,
        "COUNT(*)",
};

public static void main(String[] args) {
    Connection conn = null;

    String user = args[0];
    String port = args[1];

    try {
        String url = "jdbc:postgresql://127.0.0.1:" + port + "/doltgres";
        String password = "";

        conn = DriverManager.getConnection(url, user, password);
        Statement st = conn.createStatement();

        for (int i = 0; i < queries.length; i++) {
            String query    = queries[i];
            String expected = expectedResults[i];
            if ( st.execute(query) ) {
                ResultSet rs = st.getResultSet();
                if (rs.next()) {
                    String result = "";
                    Object fieldAccessor = fieldAccessors[i];
                    if (fieldAccessor instanceof String) {
                        result = rs.getString((String)fieldAccessor);
                    } else if (fieldAccessor instanceof Integer) {
                        result = rs.getString((Integer)fieldAccessor);
                    } else {
                        System.out.println("Unsupported field accessor value: " + fieldAccessor);
                        System.exit(1);
                    }

                    if (!expected.equals(result) && !(query.contains("dolt_commit")) && !(query.contains("dolt_merge"))) {
                        System.out.println("Query: \n" + query);
                        System.out.println("Expected:\n" + expected);
                        System.out.println("Result:\n" + result);
                        System.exit(1);
                    }
                }
            } else {
                String result = Integer.toString(st.getUpdateCount());
                if ( !expected.equals(result) ) {
                    System.out.println("Query: \n" + query);
                    System.out.println("Expected:\n" + expected);
                    System.out.println("Rows Updated:\n" + result);
                    System.exit(1);
                }
            }
        }
        System.exit(0);
    } catch (SQLException ex) {
        System.out.println("An error occurred.");
        ex.printStackTrace();
        System.exit(1);
    }
}

}
