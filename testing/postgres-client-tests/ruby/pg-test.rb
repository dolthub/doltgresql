#!/usr/bin/ruby

require 'pg'
require 'test/unit'

extend Test::Unit::Assertions

user = ARGV[0]
port = ARGV[1]
db   = "postgres"

queries = [
  "create table test (pk int, value int, d1 decimal(9, 3), f1 float, c1 char(10), t1 text, primary key(pk))",
  "select * from test",
  "insert into test (pk, value, d1, f1, c1, t1) values (0,0,0.0,0.0,'abc','a1')",
  "select * from test",
  "call dolt_add('-A');",
  "call dolt_commit('-m', 'my commit')",
  "select COUNT(*) FROM dolt.log",
  "call dolt_checkout('-b', 'mybranch')",
  "insert into test (pk, value, d1, f1, c1, t1) values (1,1, 123456.789, 420.42,'example','some text')",
  "call dolt_commit('-a', '-m', 'my commit2')",
  "call dolt_checkout('main')",
  "call dolt_merge('mybranch')",
  "select COUNT(*) FROM dolt.log",
]

# Smoke test the queries to make sure nothing blows up
conn = PG::Connection.new(:host => "localhost", :user => user, :dbname => db, :port => port, :password => "password")
queries.each do |query|
  res = conn.query(query)
end

# Then make sure we can read some data back
res = conn.query("SELECT * from test where pk = 1;")
rowCount = 0
res.each do |row|
  rowCount += 1
  assert_equal 1, row["pk"].to_i
  assert_equal 1, row["value"].to_i
  assert_equal 123456.789, row["d1"].to_f
  assert_equal 420.42, row["f1"].to_f
  assert_equal "example   ", row["c1"]
  assert_equal "some text", row["t1"]
end
assert_equal 1, rowCount

conn.close()
exit(0)
