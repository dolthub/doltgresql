use strict;

use DBI;

my $QUERY_RESPONSE = [
    { "create table test (pk int, value int, primary key(pk))" => '0E0' },
    { "insert into test (pk, value) values (0,0)" => 1 },
    { "select * from test" => 1 },
    {"call dolt_add('-A');" => '0E0' },
    {"call dolt_commit('-m', 'my commit')" => '0E0' },
    {"call dolt_checkout('-b', 'mybranch')" => '0E0' },
    {"insert into test (pk, value) values (1,1)" => 1 },
    {"call dolt_commit('-a', '-m', 'my commit2')" => '0E0' },
    {"call dolt_checkout('main')" => '0E0' },
    {"call dolt_merge('mybranch')" => '0E0' },
    {"select COUNT(*) FROM dolt_log" => 1 },
];

my $user = $ARGV[0];
my $port = $ARGV[1];
my $db   = "doltgres";

my $dsn = "DBI:Pg:database=$db;host=127.0.0.1;port=$port";
# Connect to the database
my $dbh = DBI->connect($dsn, $user, "", { PrintError => 0, RaiseError => 1 });
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

# Disconnect from the database
$dbh->disconnect();

exit 0;
