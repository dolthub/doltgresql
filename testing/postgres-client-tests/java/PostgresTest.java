import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.sql.Statement;
import java.sql.ResultSet;

public class PostgresTest {
    public String query;
    public Integer expectedUpdateCount; // for queries that don't return set of results
    public Object fieldAccessor; // can be String (i.e column name), Integer (i.e. field position) and `null` for no result set queries.
    public String[] expectedResults;

    // Parameterized constructor
    public PostgresTest(String query, Integer expectedUpdateCount, Object fieldAccessor, String[] expectedResults) {
        this.query = query;
        this.expectedUpdateCount = expectedUpdateCount;
        this.fieldAccessor = fieldAccessor;
        this.expectedResults = expectedResults;
    }
    // Getters and Setters (optional)
    public String getQuery() {
        return this.query;
    }

    public Integer getExpectedUpdateCount() {
        return this.expectedUpdateCount;
    }

    public Object getFieldAccessor() {
        return this.fieldAccessor;
    }

    public String[] getExpectedResults() {
        return this.expectedResults;
    }

    // test queries to be run against Doltgres
    private static final PostgresTest[] tests = {
        new PostgresTest("create table test (pk int, value int, d1 decimal(4,2), c1 char(10), primary key(pk))", 0, null, null),
        new PostgresTest("select pk from test", null, "pk", new String[]{}), // the table has no rows
        new PostgresTest("insert into test (pk, value, d1, c1) values (0,1,2.3,'hi'), (2,3,4.56,'hello')", 2, null, null),
        new PostgresTest("select * from test", null, "pk", new String[]{"0","2"}),
        new PostgresTest("select * from test", null, "value", new String[]{"1","3"}),
        // TODO: doltgres DECIMAL type result is returned as "2.3", should be "2.30"
        // new PostgresTest("select * from test", null, "d1", new String[]{"2.30","4.56"}),
        new PostgresTest("select * from test", null, "c1", new String[]{"hi        ","hello     "}),
        new PostgresTest("call dolt_add('-A')", 0, null, null),
        new PostgresTest("call dolt_commit('-m', 'my commit')", 0, null, null),
        new PostgresTest("select COUNT(*) FROM dolt_log", null, 1, new String[]{"2"}),
        new PostgresTest("call dolt_checkout('-b', 'mybranch')", 0, null, null),
        new PostgresTest("insert into test (pk, value, d1, c1) values (1,1,12.34,'bye')", 1, null, null),
        new PostgresTest("call dolt_commit('-a', '-m', 'my commit2')", 0, null, null),
        new PostgresTest("call dolt_checkout('main')", 0, null, null),
        new PostgresTest("call dolt_merge('mybranch')", 0, null, null),
        new PostgresTest("select COUNT(*) FROM dolt_log", null, "COUNT(*)", new String[]{"3"}), // returns res
    };

    public static void main(String[] args) {
        Connection conn = null;

        String user = args[0];
        String port = args[1];

        try {
            String url = "jdbc:postgresql://127.0.0.1:" + port + "/doltgres";
            String password = "";

            conn = DriverManager.getConnection(url, user, password);
            Statement st = conn.createStatement(ResultSet.TYPE_SCROLL_INSENSITIVE, ResultSet.CONCUR_READ_ONLY);

            for (int i = 0; i < tests.length; i++) {
                PostgresTest t = tests[i];
                String query    = t.getQuery();
                if ( st.execute(query) ) {
                    String[] expectedResults = t.getExpectedResults();
                    Object fieldAccessor = t.getFieldAccessor();
                    ResultSet rs = st.getResultSet();
                    int rowCount = getRowCount(rs);
                    // Compare the row count to the length of the array
                    if (rowCount != expectedResults.length) {
                        System.out.println("Row count in ResultSet: " + rowCount + " does not match length of the expected results: " + expectedResults.length);
                        System.exit(1);
                    }

                    Integer j = 0;
                    if (rs.next()) {
                        String expected = expectedResults[j];
                        String result = "";
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
                            System.out.println("Expected:\n'" + expected + "'");
                            System.out.println("Result:\n'" + result + "'");
                            System.exit(1);
                        }
                        j++;
                    }
                } else {
                    Integer expectedUpdateCount = t.getExpectedUpdateCount();
                    Integer result = st.getUpdateCount();
                    if ( !expectedUpdateCount.equals(result) ) {
                        System.out.println("Query: \n" + query);
                        System.out.println("Expected:\n" + expectedUpdateCount);
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

    public static int getRowCount(ResultSet rs) throws SQLException {
        int rowCount = 0;
        if (rs != null) {
            rs.last(); // Move to the last row
            rowCount = rs.getRow(); // Get the row number (which is the row count)
            rs.beforeFirst(); // Move back to the beginning of the ResultSet
        }
        return rowCount;
    }
}
