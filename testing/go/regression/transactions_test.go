// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestTransactions(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_transactions)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_transactions,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TABLE xacttest (a smallint, b real);`,
			},
			{
				Statement: `INSERT INTO xacttest VALUES
  (56, 7.8),
  (100, 99.097),
  (0, 0.09561),
  (42, 324.78);`,
			},
			{
				Statement: `INSERT INTO xacttest (a, b) VALUES (777, 777.777);`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `-- should retrieve one value--
SELECT a FROM xacttest WHERE a > 100;`,
				Results: []sql.Row{{777}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TABLE disappear (a int4);`,
			},
			{
				Statement: `DELETE FROM xacttest;`,
			},
			{
				Statement: `SELECT * FROM xacttest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `ABORT;`,
			},
			{
				Statement: `SELECT oid FROM pg_class WHERE relname = 'disappear';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM xacttest;`,
				Results:   []sql.Row{{56, 7.8}, {100, 99.097}, {0, 0.09561}, {42, 324.78}, {777, 777.777}},
			},
			{
				Statement: `CREATE TABLE writetest (a int);`,
			},
			{
				Statement: `CREATE TEMPORARY TABLE temptest (a int);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET TRANSACTION ISOLATION LEVEL SERIALIZABLE, READ ONLY, DEFERRABLE; -- ok`,
			},
			{
				Statement: `SELECT * FROM writetest; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SET TRANSACTION READ WRITE; --fail
ERROR:  transaction read-write mode must be set before any query
COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET TRANSACTION READ ONLY; -- ok`,
			},
			{
				Statement: `SET TRANSACTION READ WRITE; -- ok`,
			},
			{
				Statement: `SET TRANSACTION READ ONLY; -- ok`,
			},
			{
				Statement: `SELECT * FROM writetest; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SAVEPOINT x;`,
			},
			{
				Statement: `SET TRANSACTION READ ONLY; -- ok`,
			},
			{
				Statement: `SELECT * FROM writetest; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SET TRANSACTION READ ONLY; -- ok`,
			},
			{
				Statement: `SET TRANSACTION READ WRITE; --fail
ERROR:  cannot set transaction read-write mode inside a read-only transaction
COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET TRANSACTION READ WRITE; -- ok`,
			},
			{
				Statement: `SAVEPOINT x;`,
			},
			{
				Statement: `SET TRANSACTION READ WRITE; -- ok`,
			},
			{
				Statement: `SET TRANSACTION READ ONLY; -- ok`,
			},
			{
				Statement: `SELECT * FROM writetest; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SET TRANSACTION READ ONLY; -- ok`,
			},
			{
				Statement: `SET TRANSACTION READ WRITE; --fail
ERROR:  cannot set transaction read-write mode inside a read-only transaction
COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET TRANSACTION READ WRITE; -- ok`,
			},
			{
				Statement: `SAVEPOINT x;`,
			},
			{
				Statement: `SET TRANSACTION READ ONLY; -- ok`,
			},
			{
				Statement: `SELECT * FROM writetest; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT x;`,
			},
			{
				Statement: `SHOW transaction_read_only;  -- off`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SAVEPOINT y;`,
			},
			{
				Statement: `SET TRANSACTION READ ONLY; -- ok`,
			},
			{
				Statement: `SELECT * FROM writetest; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `RELEASE SAVEPOINT y;`,
			},
			{
				Statement: `SHOW transaction_read_only;  -- off`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY;`,
			},
			{
				Statement:   `DROP TABLE writetest; -- fail`,
				ErrorString: `cannot execute DROP TABLE in a read-only transaction`,
			},
			{
				Statement:   `INSERT INTO writetest VALUES (1); -- fail`,
				ErrorString: `cannot execute INSERT in a read-only transaction`,
			},
			{
				Statement: `SELECT * FROM writetest; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DELETE FROM temptest; -- ok`,
			},
			{
				Statement: `UPDATE temptest SET a = 0 FROM writetest WHERE temptest.a = 1 AND writetest.a = temptest.a; -- ok`,
			},
			{
				Statement: `PREPARE test AS UPDATE writetest SET a = 0; -- ok`,
			},
			{
				Statement:   `EXECUTE test; -- fail`,
				ErrorString: `cannot execute UPDATE in a read-only transaction`,
			},
			{
				Statement: `SELECT * FROM writetest, temptest; -- ok`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `CREATE TABLE test AS SELECT * FROM writetest; -- fail`,
				ErrorString: `cannot execute CREATE TABLE AS in a read-only transaction`,
			},
			{
				Statement: `START TRANSACTION READ WRITE;`,
			},
			{
				Statement: `DROP TABLE writetest; -- ok`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SET SESSION CHARACTERISTICS AS TRANSACTION READ WRITE;`,
			},
			{
				Statement: `CREATE TABLE trans_foobar (a int);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	CREATE TABLE trans_foo (a int);`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement: `		DROP TABLE trans_foo;`,
			},
			{
				Statement: `		CREATE TABLE trans_bar (a int);`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT one;`,
			},
			{
				Statement: `	RELEASE SAVEPOINT one;`,
			},
			{
				Statement: `	SAVEPOINT two;`,
			},
			{
				Statement: `		CREATE TABLE trans_baz (a int);`,
			},
			{
				Statement: `	RELEASE SAVEPOINT two;`,
			},
			{
				Statement: `	drop TABLE trans_foobar;`,
			},
			{
				Statement: `	CREATE TABLE trans_barbaz (a int);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM trans_foo;		-- should be empty`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT * FROM trans_bar;		-- shouldn't exist`,
				ErrorString: `relation "trans_bar" does not exist`,
			},
			{
				Statement: `SELECT * FROM trans_barbaz;	-- should be empty`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM trans_baz;		-- should be empty`,
				Results:   []sql.Row{},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	INSERT INTO trans_foo VALUES (1);`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement:   `		INSERT into trans_bar VALUES (1);`,
				ErrorString: `relation "trans_bar" does not exist`,
			},
			{
				Statement: `	ROLLBACK TO one;`,
			},
			{
				Statement: `	RELEASE SAVEPOINT one;`,
			},
			{
				Statement: `	SAVEPOINT two;`,
			},
			{
				Statement: `		INSERT into trans_barbaz VALUES (1);`,
			},
			{
				Statement: `	RELEASE two;`,
			},
			{
				Statement: `	SAVEPOINT three;`,
			},
			{
				Statement: `		SAVEPOINT four;`,
			},
			{
				Statement: `			INSERT INTO trans_foo VALUES (2);`,
			},
			{
				Statement: `		RELEASE SAVEPOINT four;`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT three;`,
			},
			{
				Statement: `	RELEASE SAVEPOINT three;`,
			},
			{
				Statement: `	INSERT INTO trans_foo VALUES (3);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM trans_foo;		-- should have 1 and 3`,
				Results:   []sql.Row{{1}, {3}},
			},
			{
				Statement: `SELECT * FROM trans_barbaz;	-- should have 1`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement:   `		SELECT trans_foo;`,
				ErrorString: `column "trans_foo" does not exist`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT one;`,
			},
			{
				Statement: `	RELEASE SAVEPOINT one;`,
			},
			{
				Statement: `	SAVEPOINT two;`,
			},
			{
				Statement: `		CREATE TABLE savepoints (a int);`,
			},
			{
				Statement: `		SAVEPOINT three;`,
			},
			{
				Statement: `			INSERT INTO savepoints VALUES (1);`,
			},
			{
				Statement: `			SAVEPOINT four;`,
			},
			{
				Statement: `				INSERT INTO savepoints VALUES (2);`,
			},
			{
				Statement: `				SAVEPOINT five;`,
			},
			{
				Statement: `					INSERT INTO savepoints VALUES (3);`,
			},
			{
				Statement: `				ROLLBACK TO SAVEPOINT five;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `COMMIT;		-- should not be in a transaction block`,
			},
			{
				Statement: `SELECT * FROM savepoints;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement: `		DELETE FROM savepoints WHERE a=1;`,
			},
			{
				Statement: `	RELEASE SAVEPOINT one;`,
			},
			{
				Statement: `	SAVEPOINT two;`,
			},
			{
				Statement: `		DELETE FROM savepoints WHERE a=1;`,
			},
			{
				Statement: `		SAVEPOINT three;`,
			},
			{
				Statement: `			DELETE FROM savepoints WHERE a=2;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `COMMIT;		-- should not be in a transaction block`,
			},
			{
				Statement: `SELECT * FROM savepoints;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	INSERT INTO savepoints VALUES (4);`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (5);`,
			},
			{
				Statement:   `		SELECT trans_foo;`,
				ErrorString: `column "trans_foo" does not exist`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM savepoints;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	INSERT INTO savepoints VALUES (6);`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (7);`,
			},
			{
				Statement: `	RELEASE SAVEPOINT one;`,
			},
			{
				Statement: `	INSERT INTO savepoints VALUES (8);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT a.xmin = b.xmin FROM savepoints a, savepoints b WHERE a.a=6 AND b.a=8;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT a.xmin = b.xmin FROM savepoints a, savepoints b WHERE a.a=6 AND b.a=7;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	INSERT INTO savepoints VALUES (9);`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (10);`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (11);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT a FROM savepoints WHERE a in (9, 10, 11);`,
				Results:   []sql.Row{{9}, {11}},
			},
			{
				Statement: `SELECT a.xmin = b.xmin FROM savepoints a, savepoints b WHERE a.a=9 AND b.a=11;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	INSERT INTO savepoints VALUES (12);`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (13);`,
			},
			{
				Statement: `		SAVEPOINT two;`,
			},
			{
				Statement: `			INSERT INTO savepoints VALUES (14);`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (15);`,
			},
			{
				Statement: `		SAVEPOINT two;`,
			},
			{
				Statement: `			INSERT INTO savepoints VALUES (16);`,
			},
			{
				Statement: `			SAVEPOINT three;`,
			},
			{
				Statement: `				INSERT INTO savepoints VALUES (17);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT a FROM savepoints WHERE a BETWEEN 12 AND 17;`,
				Results:   []sql.Row{{12}, {15}, {16}, {17}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	INSERT INTO savepoints VALUES (18);`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (19);`,
			},
			{
				Statement: `		SAVEPOINT two;`,
			},
			{
				Statement: `			INSERT INTO savepoints VALUES (20);`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (21);`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT one;`,
			},
			{
				Statement: `		INSERT INTO savepoints VALUES (22);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT a FROM savepoints WHERE a BETWEEN 18 AND 22;`,
				Results:   []sql.Row{{18}, {22}},
			},
			{
				Statement: `DROP TABLE savepoints;`,
			},
			{
				Statement:   `SAVEPOINT one;`,
				ErrorString: `SAVEPOINT can only be used in transaction blocks`,
			},
			{
				Statement:   `ROLLBACK TO SAVEPOINT one;`,
				ErrorString: `ROLLBACK TO SAVEPOINT can only be used in transaction blocks`,
			},
			{
				Statement:   `RELEASE SAVEPOINT one;`,
				ErrorString: `RELEASE SAVEPOINT can only be used in transaction blocks`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `  SAVEPOINT one;`,
			},
			{
				Statement:   `  SELECT 0/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `  SAVEPOINT two;    -- ignored till the end of ...`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement:   `  RELEASE SAVEPOINT one;      -- ignored till the end of ...`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `  ROLLBACK TO SAVEPOINT one;`,
			},
			{
				Statement: `  SELECT 1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT 1;			-- this should work`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	DECLARE c CURSOR FOR SELECT unique2 FROM tenk1 ORDER BY unique2;`,
			},
			{
				Statement: `	SAVEPOINT one;`,
			},
			{
				Statement: `		FETCH 10 FROM c;`,
				Results:   []sql.Row{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}},
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT one;`,
			},
			{
				Statement: `		FETCH 10 FROM c;`,
				Results:   []sql.Row{{10}, {11}, {12}, {13}, {14}, {15}, {16}, {17}, {18}, {19}},
			},
			{
				Statement: `	RELEASE SAVEPOINT one;`,
			},
			{
				Statement: `	FETCH 10 FROM c;`,
				Results:   []sql.Row{{20}, {21}, {22}, {23}, {24}, {25}, {26}, {27}, {28}, {29}},
			},
			{
				Statement: `	CLOSE c;`,
			},
			{
				Statement: `	DECLARE c CURSOR FOR SELECT unique2/0 FROM tenk1 ORDER BY unique2;`,
			},
			{
				Statement: `	SAVEPOINT two;`,
			},
			{
				Statement:   `		FETCH 10 FROM c;`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT two;`,
			},
			{
				Statement: `	-- c is now dead to the world ...
		FETCH 10 FROM c;`,
				ErrorString: `portal "c" cannot be run`,
			},
			{
				Statement: `	ROLLBACK TO SAVEPOINT two;`,
			},
			{
				Statement: `	RELEASE SAVEPOINT two;`,
			},
			{
				Statement:   `	FETCH 10 FROM c;`,
				ErrorString: `portal "c" cannot be run`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `select * from xacttest;`,
				Results:   []sql.Row{{56, 7.8}, {100, 99.097}, {0, 0.09561}, {42, 324.78}, {777, 777.777}},
			},
			{
				Statement: `create or replace function max_xacttest() returns smallint language sql as
'select max(a) from xacttest' stable;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `update xacttest set a = max_xacttest() + 10 where a > 0;`,
			},
			{
				Statement: `select * from xacttest;`,
				Results:   []sql.Row{{0, 0.09561}, {787, 7.8}, {787, 99.097}, {787, 324.78}, {787, 777.777}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `create or replace function max_xacttest() returns smallint language sql as
'select max(a) from xacttest' volatile;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `update xacttest set a = max_xacttest() + 10 where a > 0;`,
			},
			{
				Statement: `select * from xacttest;`,
				Results:   []sql.Row{{0, 0.09561}, {787, 7.8}, {797, 99.097}, {807, 324.78}, {817, 777.777}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `create or replace function max_xacttest() returns smallint language plpgsql as
'begin return max(a) from xacttest; end' stable;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `update xacttest set a = max_xacttest() + 10 where a > 0;`,
			},
			{
				Statement: `select * from xacttest;`,
				Results:   []sql.Row{{0, 0.09561}, {787, 7.8}, {787, 99.097}, {787, 324.78}, {787, 777.777}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `create or replace function max_xacttest() returns smallint language plpgsql as
'begin return max(a) from xacttest; end' volatile;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `update xacttest set a = max_xacttest() + 10 where a > 0;`,
			},
			{
				Statement: `select * from xacttest;`,
				Results:   []sql.Row{{0, 0.09561}, {787, 7.8}, {797, 99.097}, {807, 324.78}, {817, 777.777}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `	savepoint x;`,
			},
			{
				Statement: `		CREATE TABLE koju (a INT UNIQUE);`,
			},
			{
				Statement: `		INSERT INTO koju VALUES (1);`,
			},
			{
				Statement:   `		INSERT INTO koju VALUES (1);`,
				ErrorString: `duplicate key value violates unique constraint "koju_a_key"`,
			},
			{
				Statement: `	rollback to x;`,
			},
			{
				Statement: `	CREATE TABLE koju (a INT UNIQUE);`,
			},
			{
				Statement: `	INSERT INTO koju VALUES (1);`,
			},
			{
				Statement:   `	INSERT INTO koju VALUES (1);`,
				ErrorString: `duplicate key value violates unique constraint "koju_a_key"`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE trans_foo;`,
			},
			{
				Statement: `DROP TABLE trans_baz;`,
			},
			{
				Statement: `DROP TABLE trans_barbaz;`,
			},
			{
				Statement: `create function inverse(int) returns float8 as
$$
begin
  analyze revalidate_bug;`,
			},
			{
				Statement: `  return 1::float8/$1;`,
			},
			{
				Statement: `exception
  when division_by_zero then return 0;`,
			},
			{
				Statement: `end$$ language plpgsql volatile;`,
			},
			{
				Statement: `create table revalidate_bug (c float8 unique);`,
			},
			{
				Statement: `insert into revalidate_bug values (1);`,
			},
			{
				Statement: `insert into revalidate_bug values (inverse(0));`,
			},
			{
				Statement: `drop table revalidate_bug;`,
			},
			{
				Statement: `drop function inverse(int);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `savepoint x;`,
			},
			{
				Statement: `create table trans_abc (a int);`,
			},
			{
				Statement: `insert into trans_abc values (5);`,
			},
			{
				Statement: `insert into trans_abc values (10);`,
			},
			{
				Statement: `declare foo cursor for select * from trans_abc;`,
			},
			{
				Statement: `fetch from foo;`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `rollback to x;`,
			},
			{
				Statement:   `fetch from foo;`,
				ErrorString: `cursor "foo" does not exist`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create table trans_abc (a int);`,
			},
			{
				Statement: `insert into trans_abc values (5);`,
			},
			{
				Statement: `insert into trans_abc values (10);`,
			},
			{
				Statement: `insert into trans_abc values (15);`,
			},
			{
				Statement: `declare foo cursor for select * from trans_abc;`,
			},
			{
				Statement: `fetch from foo;`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `savepoint x;`,
			},
			{
				Statement: `fetch from foo;`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `rollback to x;`,
			},
			{
				Statement: `fetch from foo;`,
				Results:   []sql.Row{{15}},
			},
			{
				Statement: `abort;`,
			},
			{
				Statement: `CREATE FUNCTION invert(x float8) RETURNS float8 LANGUAGE plpgsql AS
$$ begin return 1/x; end $$;`,
			},
			{
				Statement: `CREATE FUNCTION create_temp_tab() RETURNS text
LANGUAGE plpgsql AS $$
BEGIN
  CREATE TEMP TABLE new_table (f1 float8);`,
			},
			{
				Statement: `  -- case of interest is that we fail while holding an open
  -- relcache reference to new_table
  INSERT INTO new_table SELECT invert(0.0);`,
			},
			{
				Statement: `  RETURN 'foo';`,
			},
			{
				Statement: `END $$;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DECLARE ok CURSOR FOR SELECT * FROM int8_tbl;`,
			},
			{
				Statement: `DECLARE ctt CURSOR FOR SELECT create_temp_tab();`,
			},
			{
				Statement: `FETCH ok;`,
				Results:   []sql.Row{{123, 456}},
			},
			{
				Statement: `SAVEPOINT s1;`,
			},
			{
				Statement: `FETCH ok;  -- should work`,
				Results:   []sql.Row{{123, 4567890123456789}},
			},
			{
				Statement:   `FETCH ctt; -- error occurs here`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function invert(double precision) line 1 at RETURN
SQL statement "INSERT INTO new_table SELECT invert(0.0)"
PL/pgSQL function create_temp_tab() line 6 at SQL statement
ROLLBACK TO s1;`,
			},
			{
				Statement: `FETCH ok;  -- should work`,
				Results:   []sql.Row{{4567890123456789, 123}},
			},
			{
				Statement:   `FETCH ctt; -- must be rejected`,
				ErrorString: `portal "ctt" cannot be run`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP FUNCTION create_temp_tab();`,
			},
			{
				Statement: `DROP FUNCTION invert(x float8);`,
			},
			{
				Statement: `CREATE TABLE trans_abc (a int);`,
			},
			{
				Statement: `SET default_transaction_read_only = on;`,
			},
			{
				Statement: `START TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ WRITE, DEFERRABLE;`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`on`}},
			},
			{
				Statement: `INSERT INTO trans_abc VALUES (1);`,
			},
			{
				Statement: `INSERT INTO trans_abc VALUES (2);`,
			},
			{
				Statement: `COMMIT AND CHAIN;  -- TBLOCK_END`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`on`}},
			},
			{
				Statement:   `INSERT INTO trans_abc VALUES ('error');`,
				ErrorString: `invalid input syntax for type integer: "error"`,
			},
			{
				Statement:   `INSERT INTO trans_abc VALUES (3);  -- check it's really aborted`,
				ErrorString: `current transaction is aborted, commands ignored until end of transaction block`,
			},
			{
				Statement: `COMMIT AND CHAIN;  -- TBLOCK_ABORT_END`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`on`}},
			},
			{
				Statement: `INSERT INTO trans_abc VALUES (4);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `START TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ WRITE, DEFERRABLE;`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`on`}},
			},
			{
				Statement: `SAVEPOINT x;`,
			},
			{
				Statement:   `INSERT INTO trans_abc VALUES ('error');`,
				ErrorString: `invalid input syntax for type integer: "error"`,
			},
			{
				Statement: `COMMIT AND CHAIN;  -- TBLOCK_ABORT_PENDING`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`on`}},
			},
			{
				Statement: `INSERT INTO trans_abc VALUES (5);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `START TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ WRITE, DEFERRABLE;`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`on`}},
			},
			{
				Statement: `SAVEPOINT x;`,
			},
			{
				Statement: `COMMIT AND CHAIN;  -- TBLOCK_SUBCOMMIT`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`on`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `START TRANSACTION ISOLATION LEVEL SERIALIZABLE, READ WRITE, NOT DEFERRABLE;`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`serializable`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `INSERT INTO trans_abc VALUES (6);`,
			},
			{
				Statement: `ROLLBACK AND CHAIN;  -- TBLOCK_ABORT_PENDING`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`serializable`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement:   `INSERT INTO trans_abc VALUES ('error');`,
				ErrorString: `invalid input syntax for type integer: "error"`,
			},
			{
				Statement: `ROLLBACK AND CHAIN;  -- TBLOCK_ABORT_END`,
			},
			{
				Statement: `SHOW transaction_isolation;`,
				Results:   []sql.Row{{`serializable`}},
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `SHOW transaction_deferrable;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement:   `COMMIT AND CHAIN;  -- error`,
				ErrorString: `COMMIT AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement:   `ROLLBACK AND CHAIN;  -- error`,
				ErrorString: `ROLLBACK AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement: `SELECT * FROM trans_abc ORDER BY 1;`,
				Results:   []sql.Row{{1}, {2}, {4}, {5}},
			},
			{
				Statement: `RESET default_transaction_read_only;`,
			},
			{
				Statement: `DROP TABLE trans_abc;`,
			},
			{
				Statement: `create temp table i_table (f1 int);`,
			},
			{
				Statement: `SELECT 1\; SELECT 2\; SELECT 3;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: ` ?column? 
----------
        2
(1 row)
 ?column? 
----------
        3
(1 row)
insert into i_table values(1)\; select * from i_table;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `insert into i_table values(2)\; select * from i_table\; select 1/0;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `ERROR:  division by zero
select * from i_table;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `rollback;  -- we are not in a transaction at this point`,
			},
			{
				Statement: `begin\; insert into i_table values(3)\; commit;`,
			},
			{
				Statement: `rollback;  -- we are not in a transaction at this point`,
			},
			{
				Statement: `begin\; insert into i_table values(4)\; rollback;`,
			},
			{
				Statement: `rollback;  -- we are not in a transaction at this point`,
			},
			{
				Statement: `select 1\; begin\; insert into i_table values(5);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `select 1\; begin\; insert into i_table values(6);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement:   `insert into i_table values(7)\; commit\; insert into i_table values(8)\; select 1/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `insert into i_table values(9)\; rollback\; select 2;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select * from i_table;`,
				Results:   []sql.Row{{1}, {3}, {5}, {7}},
			},
			{
				Statement: `rollback;  -- we are not in a transaction at this point`,
			},
			{
				Statement: `SELECT 1\; VACUUM;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `ERROR:  VACUUM cannot run inside a transaction block
SELECT 1\; COMMIT\; VACUUM;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `ERROR:  VACUUM cannot run inside a transaction block
SELECT 1\; SAVEPOINT sp;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `ERROR:  SAVEPOINT can only be used in transaction blocks
SELECT 1\; COMMIT\; SAVEPOINT sp;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `ERROR:  SAVEPOINT can only be used in transaction blocks
ROLLBACK TO SAVEPOINT sp\; SELECT 2;`,
				ErrorString: `ROLLBACK TO SAVEPOINT can only be used in transaction blocks`,
			},
			{
				Statement: `SELECT 2\; RELEASE SAVEPOINT sp\; SELECT 3;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `ERROR:  RELEASE SAVEPOINT can only be used in transaction blocks
SELECT 1\; BEGIN\; SAVEPOINT sp\; ROLLBACK TO SAVEPOINT sp\; COMMIT;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement:   `SET TRANSACTION READ ONLY\; COMMIT AND CHAIN;  -- error`,
				ErrorString: `COMMIT AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement:   `SET TRANSACTION READ ONLY\; ROLLBACK AND CHAIN;  -- error`,
				ErrorString: `ROLLBACK AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement: `SHOW transaction_read_only;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement: `CREATE TABLE trans_abc (a int);`,
			},
			{
				Statement:   `INSERT INTO trans_abc VALUES (7)\; COMMIT\; INSERT INTO trans_abc VALUES (8)\; COMMIT AND CHAIN;  -- 7 commit, 8 error`,
				ErrorString: `COMMIT AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement:   `INSERT INTO trans_abc VALUES (9)\; ROLLBACK\; INSERT INTO trans_abc VALUES (10)\; ROLLBACK AND CHAIN;  -- 9 rollback, 10 error`,
				ErrorString: `ROLLBACK AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement:   `INSERT INTO trans_abc VALUES (11)\; COMMIT AND CHAIN\; INSERT INTO trans_abc VALUES (12)\; COMMIT;  -- 11 error, 12 not reached`,
				ErrorString: `COMMIT AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement:   `INSERT INTO trans_abc VALUES (13)\; ROLLBACK AND CHAIN\; INSERT INTO trans_abc VALUES (14)\; ROLLBACK;  -- 13 error, 14 not reached`,
				ErrorString: `ROLLBACK AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement: `START TRANSACTION ISOLATION LEVEL REPEATABLE READ\; INSERT INTO trans_abc VALUES (15)\; COMMIT AND CHAIN;  -- 15 ok`,
			},
			{
				Statement: `SHOW transaction_isolation;  -- transaction is active at this point`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `START TRANSACTION ISOLATION LEVEL REPEATABLE READ\; INSERT INTO trans_abc VALUES (16)\; ROLLBACK AND CHAIN;  -- 16 ok`,
			},
			{
				Statement: `SHOW transaction_isolation;  -- transaction is active at this point`,
				Results:   []sql.Row{{`repeatable read`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SET default_transaction_isolation = 'read committed';`,
			},
			{
				Statement:   `START TRANSACTION ISOLATION LEVEL REPEATABLE READ\; INSERT INTO trans_abc VALUES (17)\; COMMIT\; INSERT INTO trans_abc VALUES (18)\; COMMIT AND CHAIN;  -- 17 commit, 18 error`,
				ErrorString: `COMMIT AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement: `SHOW transaction_isolation;  -- out of transaction block`,
				Results:   []sql.Row{{`read committed`}},
			},
			{
				Statement:   `START TRANSACTION ISOLATION LEVEL REPEATABLE READ\; INSERT INTO trans_abc VALUES (19)\; ROLLBACK\; INSERT INTO trans_abc VALUES (20)\; ROLLBACK AND CHAIN;  -- 19 rollback, 20 error`,
				ErrorString: `ROLLBACK AND CHAIN can only be used in transaction blocks`,
			},
			{
				Statement: `SHOW transaction_isolation;  -- out of transaction block`,
				Results:   []sql.Row{{`read committed`}},
			},
			{
				Statement: `RESET default_transaction_isolation;`,
			},
			{
				Statement: `SELECT * FROM trans_abc ORDER BY 1;`,
				Results:   []sql.Row{{7}, {15}, {17}},
			},
			{
				Statement: `DROP TABLE trans_abc;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement:   `select 1/0;`,
				ErrorString: `division by zero`,
			},
			{
				Statement:   `rollback to X;`,
				ErrorString: `savepoint "x" does not exist`,
			},
		},
	})
}
