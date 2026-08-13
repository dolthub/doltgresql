#!/usr/bin/ruby

# Tests that Sequel's built-in transaction retry semantics work against Doltgres.
#
# When two transactions modify the same row concurrently, Doltgres reports the losing
# COMMIT with SQLSTATE 40001 (serialization_failure). The pg driver surfaces that
# SQLSTATE as PG::TRSerializationFailure, Sequel maps it to Sequel::SerializationFailure,
# and Sequel's `transaction(retry_on: ...)` re-runs the block — the standard ORM recipe
# for handling optimistic-concurrency conflicts. All three layers key off the SQLSTATE,
# so this test fails if the server reports the conflict with any other code.

require 'sequel'

user = ARGV[0]
port = ARGV[1].to_i

def connect(user, port)
  Sequel.connect(
    adapter: 'postgres',
    host: 'localhost',
    port: port,
    database: 'postgres',
    user: user,
    password: 'password'
  )
end

db1 = connect(user, port)
db2 = connect(user, port)

db1.run("DROP TABLE IF EXISTS retry_test")
db1.run("CREATE TABLE retry_test (id INT PRIMARY KEY, val INT, who TEXT)")
db1[:retry_test].insert(id: 1, val: 0, who: 'nobody')

# Part 1: interpretation. A conflicting concurrent commit must surface as
# Sequel::SerializationFailure — the exception class Sequel documents for
# retry_on — which it derives from SQLSTATE 40001.
begin
  db1.transaction do
    db1[:retry_test].where(id: 1).update(val: 100, who: 'txn1')
    # conflicting write from a second connection, autocommitted while txn1 is open
    db2[:retry_test].where(id: 1).update(val: 200, who: 'txn2')
  end
  raise "expected the conflicting transaction to raise, but it committed"
rescue Sequel::SerializationFailure => e
  puts "conflict surfaced as #{e.class} (driver: #{e.wrapped_exception.class})"
end

# The session must remain usable after the failed commit.
one = db1["SELECT 1 AS one"].first[:one]
raise "connection unusable after serialization failure" unless one == 1

# Part 2: retry. Sequel re-runs the block on Sequel::SerializationFailure; the
# second attempt has no concurrent writer, so it must commit.
db1[:retry_test].where(id: 1).update(val: 0, who: 'nobody')
attempts = 0
db1.transaction(retry_on: Sequel::SerializationFailure, num_retries: 5) do
  attempts += 1
  db1[:retry_test].where(id: 1).update(val: 100, who: 'txn1')
  if attempts == 1
    db2[:retry_test].where(id: 1).update(val: 200, who: 'txn2')
  end
end

raise "expected the transaction to be attempted twice, got #{attempts}" unless attempts == 2

row = db1[:retry_test].where(id: 1).first
unless row[:val] == 100 && row[:who] == 'txn1'
  raise "expected the retried transaction's write to win, got #{row}"
end

db1.disconnect
db2.disconnect
puts "Sequel serialization-failure retry test passed (attempts=#{attempts})"
