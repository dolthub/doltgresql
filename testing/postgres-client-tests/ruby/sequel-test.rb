#!/usr/bin/ruby

require 'sequel'

user = ARGV[0]
port = ARGV[1].to_i

DB = Sequel.connect(
  adapter: 'postgres',
  host: 'localhost',
  port: port,
  database: 'postgres',
  user: user,
  password: 'password'
)

# SELECT pk from test_table (set up by bats setup())
pk = DB["SELECT pk FROM test_table LIMIT 1"].first[:pk]
raise "expected pk=1, got #{pk}" unless pk == 1

# INSERT
DB.run("INSERT INTO test_table VALUES (2)")

# COUNT
count = DB["SELECT COUNT(*) FROM test_table"].first[:count]
raise "expected count=2, got #{count}" unless count == 2

# Dolt workflow
[
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
].each { |sql| DB.run(sql) }

log_count = DB["SELECT COUNT(*) FROM dolt_log"].first[:count]
raise "expected 4 dolt_log entries, got #{log_count}" unless log_count == 4

test_count = DB["SELECT COUNT(*) FROM test"].first[:count]
raise "expected 2 rows in test, got #{test_count}" unless test_count == 2

DB.disconnect
puts "Sequel test passed"
