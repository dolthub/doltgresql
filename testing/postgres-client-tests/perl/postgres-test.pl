use strict;

use DBI;

my $QUERY_RESPONSE = [
    { "create table test (pk int, value int, d1 decimal(9, 3), f1 float, c1 char(10), t1 text, primary key(pk))" => '0E0' },
    { "insert into test (pk, value, d1, f1, c1, t1) values (0,0,0.0,0.0,'abc','a1')" => 1 },
    { "select * from test" => 1 },
    {"call dolt_add('-A');" => '0E0' },
    {"call dolt_commit('-m', 'my commit')" => '0E0' },
    {"call dolt_checkout('-b', 'mybranch')" => '0E0' },
    {"insert into test (pk, value, d1, f1, c1, t1) values (10,10, 123456.789, 420.42,'example','some text')" => 1 },
    {"call dolt_commit('-a', '-m', 'my commit2')" => '0E0' },
    {"call dolt_checkout('main')" => '0E0' },
    {"call dolt_merge('mybranch')" => '0E0' },
    {"select COUNT(*) FROM dolt.log" => 1 },
];

my $user = $ARGV[0];
my $port = $ARGV[1];
my $db   = "doltgres";

my $dsn = "DBI:Pg:database=$db;host=127.0.0.1;port=$port";
# Connect to the database
my $dbh = DBI->connect($dsn, $user, "password", { PrintError => 0, RaiseError => 1 });
die "failed to connect to database:DBI->errstr()" unless($dbh);

foreach my $query_response ( @{$QUERY_RESPONSE} ) {
    my @query_keys = keys %{$query_response};
    my $query      = $query_keys[0];
    my $exp_result = $query_response->{$query};

    my $result = $dbh->do($query);
    if ( $result != $exp_result ) {
	print "QUERY: $query\n";
	print "EXPECTED: $exp_result\n";
	print "RESULT: $result\n";
	exit 1
    }
}

# Define the SQL query
my $query = "SELECT * FROM test WHERE pk = 10";

# Prepare the query
my $sth = $dbh->prepare($query);

# Execute the query
$sth->execute();

my @cols = ("pk", "value", "d1", "f1", "c1", "t1");
my @expectedRowResults = ("10", "10", "123456.789", "420.42", "example   ", "some text");

# Fetch and process the results
while (my $row = $sth->fetchrow_hashref()) {
    if ($row) {
        # Process the row if it's defined
        # Access individual column values using hash dereferencing
        for my $i (0 .. $#cols) {
            my $val = $row->{$cols[$i]};
            # Comparing the strings using eq operator
            my $c = $val eq $expectedRowResults[$i];
            if ( $c == 0 ) {
                print "Expected: '$expectedRowResults[$i]'\n";
                print "Actual: '$val'\n";
                # Disconnect from the database
                $dbh->disconnect();
                exit 1;
            }
        }
    } else {
        print("no rows?");
        last; # Exit the loop when there are no more rows
    }
}

# Disconnect from the database
$dbh->disconnect();

exit 0;
