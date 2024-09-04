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

func TestSequence(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_sequence)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_sequence,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `CREATE SEQUENCE sequence_testx INCREMENT BY 0;`,
				ErrorString: `INCREMENT must not be zero`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx INCREMENT BY -1 MINVALUE 20;`,
				ErrorString: `MINVALUE (20) must be less than MAXVALUE (-1)`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx INCREMENT BY 1 MAXVALUE -20;`,
				ErrorString: `MINVALUE (1) must be less than MAXVALUE (-20)`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx INCREMENT BY -1 START 10;`,
				ErrorString: `START value (10) cannot be greater than MAXVALUE (-1)`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx INCREMENT BY 1 START -10;`,
				ErrorString: `START value (-10) cannot be less than MINVALUE (1)`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx CACHE 0;`,
				ErrorString: `CACHE (0) must be greater than zero`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx OWNED BY nobody;  -- nonsense word`,
				ErrorString: `invalid OWNED BY option`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx OWNED BY pg_class_oid_index.oid;  -- not a table`,
				ErrorString: `sequence cannot be owned by relation "pg_class_oid_index"`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx OWNED BY pg_class.relname;  -- not same schema`,
				ErrorString: `sequence must be in same schema as table it is linked to`,
			},
			{
				Statement: `CREATE TABLE sequence_test_table (a int);`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx OWNED BY sequence_test_table.b;  -- wrong column`,
				ErrorString: `column "b" of relation "sequence_test_table" does not exist`,
			},
			{
				Statement: `DROP TABLE sequence_test_table;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test5 AS integer;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test6 AS smallint;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test7 AS bigint;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test8 AS integer MAXVALUE 100000;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test9 AS integer INCREMENT BY -1;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test10 AS integer MINVALUE -100000 START 1;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test11 AS smallint;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test12 AS smallint INCREMENT -1;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test13 AS smallint MINVALUE -32768;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test14 AS smallint MAXVALUE 32767 INCREMENT -1;`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx AS text;`,
				ErrorString: `sequence type must be smallint, integer, or bigint`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx AS nosuchtype;`,
				ErrorString: `type "nosuchtype" does not exist`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx AS smallint MAXVALUE 100000;`,
				ErrorString: `MAXVALUE (100000) is out of range for sequence data type smallint`,
			},
			{
				Statement:   `CREATE SEQUENCE sequence_testx AS smallint MINVALUE -100000;`,
				ErrorString: `MINVALUE (-100000) is out of range for sequence data type smallint`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test5 AS smallint;  -- success, max will be adjusted`,
			},
			{
				Statement:   `ALTER SEQUENCE sequence_test8 AS smallint;  -- fail, max has to be adjusted`,
				ErrorString: `MAXVALUE (100000) is out of range for sequence data type smallint`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test8 AS smallint MAXVALUE 20000;  -- ok now`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test9 AS smallint;  -- success, min will be adjusted`,
			},
			{
				Statement:   `ALTER SEQUENCE sequence_test10 AS smallint;  -- fail, min has to be adjusted`,
				ErrorString: `MINVALUE (-100000) is out of range for sequence data type smallint`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test10 AS smallint MINVALUE -20000;  -- ok now`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test11 AS int;  -- max will be adjusted`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test12 AS int;  -- min will be adjusted`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test13 AS int;  -- min and max will be adjusted`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test14 AS int;  -- min and max will be adjusted`,
			},
			{
				Statement: `---
---
CREATE TABLE serialTest1 (f1 text, f2 serial);`,
			},
			{
				Statement: `INSERT INTO serialTest1 VALUES ('foo');`,
			},
			{
				Statement: `INSERT INTO serialTest1 VALUES ('bar');`,
			},
			{
				Statement: `INSERT INTO serialTest1 VALUES ('force', 100);`,
			},
			{
				Statement:   `INSERT INTO serialTest1 VALUES ('wrong', NULL);`,
				ErrorString: `null value in column "f2" of relation "serialtest1" violates not-null constraint`,
			},
			{
				Statement: `SELECT * FROM serialTest1;`,
				Results:   []sql.Row{{`foo`, 1}, {`bar`, 2}, {`force`, 100}},
			},
			{
				Statement: `SELECT pg_get_serial_sequence('serialTest1', 'f2');`,
				Results:   []sql.Row{{`public.serialtest1_f2_seq`}},
			},
			{
				Statement: `CREATE TABLE serialTest2 (f1 text, f2 serial, f3 smallserial, f4 serial2,
  f5 bigserial, f6 serial8);`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1)
  VALUES ('test_defaults');`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f2, f3, f4, f5, f6)
  VALUES ('test_max_vals', 2147483647, 32767, 32767, 9223372036854775807,
          9223372036854775807),
         ('test_min_vals', -2147483648, -32768, -32768, -9223372036854775808,
          -9223372036854775808);`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f3)
  VALUES ('bogus', -32769);`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f4)
  VALUES ('bogus', -32769);`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f3)
  VALUES ('bogus', 32768);`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f4)
  VALUES ('bogus', 32768);`,
				ErrorString: `smallint out of range`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f5)
  VALUES ('bogus', -9223372036854775809);`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f6)
  VALUES ('bogus', -9223372036854775809);`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f5)
  VALUES ('bogus', 9223372036854775808);`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `INSERT INTO serialTest2 (f1, f6)
  VALUES ('bogus', 9223372036854775808);`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement: `SELECT * FROM serialTest2 ORDER BY f2 ASC;`,
				Results:   []sql.Row{{`test_min_vals`, -2147483648, -32768, -32768, -9223372036854775808, -9223372036854775808}, {`test_defaults`, 1, 1, 1, 1, 1}, {`test_max_vals`, 2147483647, 32767, 32767, 9223372036854775807, 9223372036854775807}},
			},
			{
				Statement: `SELECT nextval('serialTest2_f2_seq');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT nextval('serialTest2_f3_seq');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT nextval('serialTest2_f4_seq');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT nextval('serialTest2_f5_seq');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT nextval('serialTest2_f6_seq');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `CREATE SEQUENCE sequence_test;`,
			},
			{
				Statement: `CREATE SEQUENCE IF NOT EXISTS sequence_test;`,
			},
			{
				Statement: `SELECT nextval('sequence_test'::text);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT nextval('sequence_test'::regclass);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT currval('sequence_test'::text);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT currval('sequence_test'::regclass);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT setval('sequence_test'::text, 32);`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement: `SELECT nextval('sequence_test'::regclass);`,
				Results:   []sql.Row{{33}},
			},
			{
				Statement: `SELECT setval('sequence_test'::text, 99, false);`,
				Results:   []sql.Row{{99}},
			},
			{
				Statement: `SELECT nextval('sequence_test'::regclass);`,
				Results:   []sql.Row{{99}},
			},
			{
				Statement: `SELECT setval('sequence_test'::regclass, 32);`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement: `SELECT nextval('sequence_test'::text);`,
				Results:   []sql.Row{{33}},
			},
			{
				Statement: `SELECT setval('sequence_test'::regclass, 99, false);`,
				Results:   []sql.Row{{99}},
			},
			{
				Statement: `SELECT nextval('sequence_test'::text);`,
				Results:   []sql.Row{{99}},
			},
			{
				Statement: `DISCARD SEQUENCES;`,
			},
			{
				Statement:   `SELECT currval('sequence_test'::regclass);`,
				ErrorString: `currval of sequence "sequence_test" is not yet defined in this session`,
			},
			{
				Statement: `DROP SEQUENCE sequence_test;`,
			},
			{
				Statement: `CREATE SEQUENCE foo_seq;`,
			},
			{
				Statement: `ALTER TABLE foo_seq RENAME TO foo_seq_new;`,
			},
			{
				Statement: `SELECT * FROM foo_seq_new;`,
				Results:   []sql.Row{{1, 0, false}},
			},
			{
				Statement: `SELECT nextval('foo_seq_new');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT nextval('foo_seq_new');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT last_value, log_cnt IN (31, 32) AS log_cnt_ok, is_called FROM foo_seq_new;`,
				Results:   []sql.Row{{2, true, true}},
			},
			{
				Statement: `DROP SEQUENCE foo_seq_new;`,
			},
			{
				Statement: `ALTER TABLE serialtest1_f2_seq RENAME TO serialtest1_f2_foo;`,
			},
			{
				Statement: `INSERT INTO serialTest1 VALUES ('more');`,
			},
			{
				Statement: `SELECT * FROM serialTest1;`,
				Results:   []sql.Row{{`foo`, 1}, {`bar`, 2}, {`force`, 100}, {`more`, 3}},
			},
			{
				Statement: `CREATE TEMP SEQUENCE myseq2;`,
			},
			{
				Statement: `CREATE TEMP SEQUENCE myseq3;`,
			},
			{
				Statement: `CREATE TEMP TABLE t1 (
  f1 serial,
  f2 int DEFAULT nextval('myseq2'),
  f3 int DEFAULT nextval('myseq3'::text)
);`,
			},
			{
				Statement:   `DROP SEQUENCE t1_f1_seq;`,
				ErrorString: `cannot drop sequence t1_f1_seq because other objects depend on it`,
			},
			{
				Statement:   `DROP SEQUENCE myseq2;`,
				ErrorString: `cannot drop sequence myseq2 because other objects depend on it`,
			},
			{
				Statement: `DROP SEQUENCE myseq3;`,
			},
			{
				Statement: `DROP TABLE t1;`,
			},
			{
				Statement:   `DROP SEQUENCE t1_f1_seq;`,
				ErrorString: `sequence "t1_f1_seq" does not exist`,
			},
			{
				Statement: `DROP SEQUENCE myseq2;`,
			},
			{
				Statement: `ALTER SEQUENCE IF EXISTS sequence_test2 RESTART WITH 24
  INCREMENT BY 4 MAXVALUE 36 MINVALUE 5 CYCLE;`,
			},
			{
				Statement:   `ALTER SEQUENCE serialTest1 CYCLE;  -- error, not a sequence`,
				ErrorString: `"serialtest1" is not a sequence`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test2 START WITH 32;`,
			},
			{
				Statement: `CREATE SEQUENCE sequence_test4 INCREMENT BY -1;`,
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement: `SELECT nextval('sequence_test4');`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `ALTER SEQUENCE sequence_test2 RESTART;`,
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement:   `ALTER SEQUENCE sequence_test2 RESTART WITH 0;  -- error`,
				ErrorString: `RESTART value (0) cannot be less than MINVALUE (1)`,
			},
			{
				Statement:   `ALTER SEQUENCE sequence_test4 RESTART WITH 40;  -- error`,
				ErrorString: `RESTART value (40) cannot be greater than MAXVALUE (-1)`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test2 RESTART WITH 24
  INCREMENT BY 4 MAXVALUE 36 MINVALUE 5 CYCLE;`,
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{24}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{28}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{36}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');  -- cycled`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `ALTER SEQUENCE sequence_test2 RESTART WITH 24
  NO CYCLE;`,
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{24}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{28}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{36}},
			},
			{
				Statement:   `SELECT nextval('sequence_test2');  -- error`,
				ErrorString: `nextval: reached maximum value of sequence "sequence_test2" (36)`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test2 RESTART WITH -24 START WITH -24
  INCREMENT BY -4 MINVALUE -36 MAXVALUE -5 CYCLE;`,
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{-24}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{-28}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{-32}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{-36}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');  -- cycled`,
				Results:   []sql.Row{{-5}},
			},
			{
				Statement: `ALTER SEQUENCE sequence_test2 RESTART WITH -24
  NO CYCLE;`,
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{-24}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{-28}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{-32}},
			},
			{
				Statement: `SELECT nextval('sequence_test2');`,
				Results:   []sql.Row{{-36}},
			},
			{
				Statement:   `SELECT nextval('sequence_test2');  -- error`,
				ErrorString: `nextval: reached minimum value of sequence "sequence_test2" (-36)`,
			},
			{
				Statement: `ALTER SEQUENCE IF EXISTS sequence_test2 RESTART WITH 32 START WITH 32
  INCREMENT BY 4 MAXVALUE 36 MINVALUE 5 CYCLE;`,
			},
			{
				Statement:   `SELECT setval('sequence_test2', -100);  -- error`,
				ErrorString: `setval: value -100 is out of bounds for sequence "sequence_test2" (5..36)`,
			},
			{
				Statement:   `SELECT setval('sequence_test2', 100);  -- error`,
				ErrorString: `setval: value 100 is out of bounds for sequence "sequence_test2" (5..36)`,
			},
			{
				Statement: `SELECT setval('sequence_test2', 5);`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `CREATE SEQUENCE sequence_test3;  -- not read from, to test is_called`,
			},
			{
				Statement: `SELECT * FROM information_schema.sequences
  WHERE sequence_name ~ ANY(ARRAY['sequence_test', 'serialtest'])
  ORDER BY sequence_name ASC;`,
				Results: []sql.Row{{`regression`, `public`, `sequence_test10`, `smallint`, 16, 2, 0, 1, -20000, 32767, 1, `NO`}, {`regression`, `public`, `sequence_test11`, `integer`, 32, 2, 0, 1, 1, 2147483647, 1, `NO`}, {`regression`, `public`, `sequence_test12`, `integer`, 32, 2, 0, -1, -2147483648, -1, -1, `NO`}, {`regression`, `public`, `sequence_test13`, `integer`, 32, 2, 0, -32768, -2147483648, 2147483647, 1, `NO`}, {`regression`, `public`, `sequence_test14`, `integer`, 32, 2, 0, 32767, -2147483648, 2147483647, -1, `NO`}, {`regression`, `public`, `sequence_test2`, `bigint`, 64, 2, 0, 32, 5, 36, 4, `YES`}, {`regression`, `public`, `sequence_test3`, `bigint`, 64, 2, 0, 1, 1, 9223372036854775807, 1, `NO`}, {`regression`, `public`, `sequence_test4`, `bigint`, 64, 2, 0, -1, -9223372036854775808, -1, -1, `NO`}, {`regression`, `public`, `sequence_test5`, `smallint`, 16, 2, 0, 1, 1, 32767, 1, `NO`}, {`regression`, `public`, `sequence_test6`, `smallint`, 16, 2, 0, 1, 1, 32767, 1, `NO`}, {`regression`, `public`, `sequence_test7`, `bigint`, 64, 2, 0, 1, 1, 9223372036854775807, 1, `NO`}, {`regression`, `public`, `sequence_test8`, `smallint`, 16, 2, 0, 1, 1, 20000, 1, `NO`}, {`regression`, `public`, `sequence_test9`, `smallint`, 16, 2, 0, -1, -32768, -1, -1, `NO`}, {`regression`, `public`, `serialtest1_f2_foo`, `integer`, 32, 2, 0, 1, 1, 2147483647, 1, `NO`}, {`regression`, `public`, `serialtest2_f2_seq`, `integer`, 32, 2, 0, 1, 1, 2147483647, 1, `NO`}, {`regression`, `public`, `serialtest2_f3_seq`, `smallint`, 16, 2, 0, 1, 1, 32767, 1, `NO`}, {`regression`, `public`, `serialtest2_f4_seq`, `smallint`, 16, 2, 0, 1, 1, 32767, 1, `NO`}, {`regression`, `public`, `serialtest2_f5_seq`, `bigint`, 64, 2, 0, 1, 1, 9223372036854775807, 1, `NO`}, {`regression`, `public`, `serialtest2_f6_seq`, `bigint`, 64, 2, 0, 1, 1, 9223372036854775807, 1, `NO`}},
			},
			{
				Statement: `SELECT schemaname, sequencename, start_value, min_value, max_value, increment_by, cycle, cache_size, last_value
FROM pg_sequences
WHERE sequencename ~ ANY(ARRAY['sequence_test', 'serialtest'])
  ORDER BY sequencename ASC;`,
				Results: []sql.Row{{`public`, `sequence_test10`, 1, -20000, 32767, 1, false, 1, ``}, {`public`, `sequence_test11`, 1, 1, 2147483647, 1, false, 1, ``}, {`public`, `sequence_test12`, -1, -2147483648, -1, -1, false, 1, ``}, {`public`, `sequence_test13`, -32768, -2147483648, 2147483647, 1, false, 1, ``}, {`public`, `sequence_test14`, 32767, -2147483648, 2147483647, -1, false, 1, ``}, {`public`, `sequence_test2`, 32, 5, 36, 4, true, 1, 5}, {`public`, `sequence_test3`, 1, 1, 9223372036854775807, 1, false, 1, ``}, {`public`, `sequence_test4`, -1, -9223372036854775808, -1, -1, false, 1, -1}, {`public`, `sequence_test5`, 1, 1, 32767, 1, false, 1, ``}, {`public`, `sequence_test6`, 1, 1, 32767, 1, false, 1, ``}, {`public`, `sequence_test7`, 1, 1, 9223372036854775807, 1, false, 1, ``}, {`public`, `sequence_test8`, 1, 1, 20000, 1, false, 1, ``}, {`public`, `sequence_test9`, -1, -32768, -1, -1, false, 1, ``}, {`public`, `serialtest1_f2_foo`, 1, 1, 2147483647, 1, false, 1, 3}, {`public`, `serialtest2_f2_seq`, 1, 1, 2147483647, 1, false, 1, 2}, {`public`, `serialtest2_f3_seq`, 1, 1, 32767, 1, false, 1, 2}, {`public`, `serialtest2_f4_seq`, 1, 1, 32767, 1, false, 1, 2}, {`public`, `serialtest2_f5_seq`, 1, 1, 9223372036854775807, 1, false, 1, 2}, {`public`, `serialtest2_f6_seq`, 1, 1, 9223372036854775807, 1, false, 1, 2}},
			},
			{
				Statement: `SELECT * FROM pg_sequence_parameters('sequence_test4'::regclass);`,
				Results:   []sql.Row{{-1, -9223372036854775808, -1, -1, false, 1, 20}},
			},
			{
				Statement: `\d sequence_test4
                       Sequence "public.sequence_test4"
  Type  | Start |       Minimum        | Maximum | Increment | Cycles? | Cache 
--------+-------+----------------------+---------+-----------+---------+-------
 bigint |    -1 | -9223372036854775808 |      -1 |        -1 | no      |     1
\d serialtest2_f2_seq
                 Sequence "public.serialtest2_f2_seq"
  Type   | Start | Minimum |  Maximum   | Increment | Cycles? | Cache 
---------+-------+---------+------------+-----------+---------+-------
 integer |     1 |       1 | 2147483647 |         1 | no      |     1
Owned by: public.serialtest2.f2
COMMENT ON SEQUENCE asdf IS 'won''t work';`,
				ErrorString: `relation "asdf" does not exist`,
			},
			{
				Statement: `COMMENT ON SEQUENCE sequence_test2 IS 'will work';`,
			},
			{
				Statement: `COMMENT ON SEQUENCE sequence_test2 IS NULL;`,
			},
			{
				Statement: `CREATE SEQUENCE seq;`,
			},
			{
				Statement: `SELECT nextval('seq');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT lastval();`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT setval('seq', 99);`,
				Results:   []sql.Row{{99}},
			},
			{
				Statement: `SELECT lastval();`,
				Results:   []sql.Row{{99}},
			},
			{
				Statement: `DISCARD SEQUENCES;`,
			},
			{
				Statement:   `SELECT lastval();`,
				ErrorString: `lastval is not yet defined in this session`,
			},
			{
				Statement: `CREATE SEQUENCE seq2;`,
			},
			{
				Statement: `SELECT nextval('seq2');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT lastval();`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP SEQUENCE seq2;`,
			},
			{
				Statement:   `SELECT lastval();`,
				ErrorString: `lastval is not yet defined in this session`,
			},
			{
				Statement: `CREATE UNLOGGED SEQUENCE sequence_test_unlogged;`,
			},
			{
				Statement: `ALTER SEQUENCE sequence_test_unlogged SET LOGGED;`,
			},
			{
				Statement: `\d sequence_test_unlogged
                   Sequence "public.sequence_test_unlogged"
  Type  | Start | Minimum |       Maximum       | Increment | Cycles? | Cache 
--------+-------+---------+---------------------+-----------+---------+-------
 bigint |     1 |       1 | 9223372036854775807 |         1 | no      |     1
ALTER SEQUENCE sequence_test_unlogged SET UNLOGGED;`,
			},
			{
				Statement: `\d sequence_test_unlogged
              Unlogged sequence "public.sequence_test_unlogged"
  Type  | Start | Minimum |       Maximum       | Increment | Cycles? | Cache 
--------+-------+---------+---------------------+-----------+---------+-------
 bigint |     1 |       1 | 9223372036854775807 |         1 | no      |     1
DROP SEQUENCE sequence_test_unlogged;`,
			},
			{
				Statement: `CREATE TEMPORARY SEQUENCE sequence_test_temp1;`,
			},
			{
				Statement: `START TRANSACTION READ ONLY;`,
			},
			{
				Statement: `SELECT nextval('sequence_test_temp1');  -- ok`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT nextval('sequence_test2');  -- error`,
				ErrorString: `cannot execute nextval() in a read-only transaction`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `START TRANSACTION READ ONLY;`,
			},
			{
				Statement: `SELECT setval('sequence_test_temp1', 1);  -- ok`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT setval('sequence_test2', 1);  -- error`,
				ErrorString: `cannot execute setval() in a read-only transaction`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE USER regress_seq_user;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT SELECT ON seq3 TO regress_seq_user;`,
			},
			{
				Statement:   `SELECT nextval('seq3');`,
				ErrorString: `permission denied for sequence seq3`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT UPDATE ON seq3 TO regress_seq_user;`,
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT USAGE ON seq3 TO regress_seq_user;`,
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT SELECT ON seq3 TO regress_seq_user;`,
			},
			{
				Statement: `SELECT currval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT UPDATE ON seq3 TO regress_seq_user;`,
			},
			{
				Statement:   `SELECT currval('seq3');`,
				ErrorString: `permission denied for sequence seq3`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT USAGE ON seq3 TO regress_seq_user;`,
			},
			{
				Statement: `SELECT currval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT SELECT ON seq3 TO regress_seq_user;`,
			},
			{
				Statement: `SELECT lastval();`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT UPDATE ON seq3 TO regress_seq_user;`,
			},
			{
				Statement:   `SELECT lastval();`,
				ErrorString: `permission denied for sequence seq3`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `GRANT USAGE ON seq3 TO regress_seq_user;`,
			},
			{
				Statement: `SELECT lastval();`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement: `CREATE SEQUENCE seq3;`,
			},
			{
				Statement: `REVOKE ALL ON seq3 FROM regress_seq_user;`,
			},
			{
				Statement: `SAVEPOINT save;`,
			},
			{
				Statement:   `SELECT setval('seq3', 5);`,
				ErrorString: `permission denied for sequence seq3`,
			},
			{
				Statement: `ROLLBACK TO save;`,
			},
			{
				Statement: `GRANT UPDATE ON seq3 TO regress_seq_user;`,
			},
			{
				Statement: `SELECT setval('seq3', 5);`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `SELECT nextval('seq3');`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL SESSION AUTHORIZATION regress_seq_user;`,
			},
			{
				Statement:   `ALTER SEQUENCE sequence_test2 START WITH 1;`,
				ErrorString: `must be owner of sequence sequence_test2`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE serialTest1, serialTest2;`,
			},
			{
				Statement: `SELECT * FROM information_schema.sequences WHERE sequence_name IN
  ('sequence_test2', 'serialtest2_f2_seq', 'serialtest2_f3_seq',
   'serialtest2_f4_seq', 'serialtest2_f5_seq', 'serialtest2_f6_seq')
  ORDER BY sequence_name ASC;`,
				Results: []sql.Row{{`regression`, `public`, `sequence_test2`, `bigint`, 64, 2, 0, 32, 5, 36, 4, `YES`}},
			},
			{
				Statement: `DROP USER regress_seq_user;`,
			},
			{
				Statement: `DROP SEQUENCE seq;`,
			},
			{
				Statement: `CREATE SEQUENCE test_seq1 CACHE 10;`,
			},
			{
				Statement: `SELECT nextval('test_seq1');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT nextval('test_seq1');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT nextval('test_seq1');`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `DROP SEQUENCE test_seq1;`,
			},
		},
	})
}
