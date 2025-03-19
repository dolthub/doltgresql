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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Basic CREATE SEQUENCE and DROP SEQUENCE",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT nextval('test'::regclass);",
					Expected: []sql.Row{{4}},
				},
				{
					Query:       "SELECT nextval('doesnotexist'::regclass);",
					Skip:        true, // TODO: error is valid but text changed, need to adjust
					ExpectedErr: "does not exist",
				},
				{
					Query:    "DROP SEQUENCE test;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT nextval('test');",
					ExpectedErr: "does not exist",
				},
			},
		},
		{
			Name: "CREATE SEQUENCE IF NOT EXISTS",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test1;",
					ExpectedErr: "already exists",
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "CREATE SEQUENCE IF NOT EXISTS test1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "CREATE SEQUENCE IF NOT EXISTS test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "CREATE SEQUENCE IF NOT EXISTS test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "DROP SEQUENCE IF NOT EXISTS",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE SEQUENCE test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "DROP SEQUENCE test1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "DROP SEQUENCE test1;",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "DROP SEQUENCE IF EXISTS test1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT nextval('test1');",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "DROP SEQUENCE IF EXISTS test2;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT nextval('test2');",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "DROP SEQUENCE IF EXISTS test2;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "MINVALUE and MAXVALUE with DATA TYPE",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1 AS SMALLINT MINVALUE -32768;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test2 AS SMALLINT MINVALUE -32769;",
					ExpectedErr: "out of range",
				},
				{
					Query:    "CREATE SEQUENCE test3 AS SMALLINT MAXVALUE 32767;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test4 AS SMALLINT MINVALUE 32768;",
					ExpectedErr: "out of range",
				},
				{
					Query:    "CREATE SEQUENCE test5 AS INTEGER MINVALUE -2147483648;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test6 AS INTEGER MINVALUE -2147483649;",
					ExpectedErr: "out of range",
				},
				{
					Query:    "CREATE SEQUENCE test7 AS INTEGER MAXVALUE 2147483647;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test8 AS INTEGER MINVALUE 2147483648;",
					ExpectedErr: "out of range",
				},
				{
					Query:    "CREATE SEQUENCE test9 AS BIGINT MINVALUE -9223372036854775808;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test10 AS BIGINT MINVALUE -9223372036854775809;",
					ExpectedErr: "out of int64 range",
				},
				{
					Query:    "CREATE SEQUENCE test11 AS BIGINT MAXVALUE 9223372036854775807;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test12 AS BIGINT MINVALUE 9223372036854775808;",
					ExpectedErr: "out of int64 range",
				},
			},
		},
		{
			Name: "CREATE SEQUENCE START",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1 START 39;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{39}},
				},
				{
					Query:       "CREATE SEQUENCE test2 START 0;",
					ExpectedErr: "cannot be less than",
				},
				{
					Query:    "CREATE SEQUENCE test2 MINVALUE 0 START 0;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "CREATE SEQUENCE test3 MINVALUE -100 START -7;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test3');",
					Expected: []sql.Row{{-7}},
				},
				{
					Query:       "CREATE SEQUENCE test4 START -5 INCREMENT 1;",
					ExpectedErr: "cannot be less than",
				},
				{
					Query:    "CREATE SEQUENCE test4 START -5 INCREMENT -1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{-5}},
				},
				{
					Query:       "CREATE SEQUENCE test5 START 25 INCREMENT -1;",
					ExpectedErr: "cannot be greater than",
				},
				{
					Query:    "CREATE SEQUENCE test5 START 25 MAXVALUE 25 INCREMENT -1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{25}},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{24}},
				},
			},
		},
		{
			Name: "CYCLE and NO CYCLE",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1 MINVALUE 0 MAXVALUE 3 START 2 INCREMENT 1 NO CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:       "SELECT nextval('test1');",
					ExpectedErr: "reached maximum value",
				},
				{
					Query:    "CREATE SEQUENCE test2 MINVALUE 0 MAXVALUE 3 START 2 INCREMENT 1 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "CREATE SEQUENCE test3 MINVALUE 0 MAXVALUE 3 START 1 INCREMENT -1 NO CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test3');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test3');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:       "SELECT nextval('test3');",
					ExpectedErr: "reached minimum value",
				},
				{
					Query:    "CREATE SEQUENCE test4 MINVALUE 0 MAXVALUE 3 START 1 INCREMENT -1 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "CREATE SEQUENCE test5 MINVALUE 1 MAXVALUE 7 START 1 INCREMENT 5 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{6}},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "CREATE SEQUENCE test6 MINVALUE 1 MAXVALUE 7 START 6 INCREMENT -5 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test6');",
					Expected: []sql.Row{{6}},
				},
				{
					Query:    "SELECT nextval('test6');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test6');",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT nextval('test6');",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "nextval() over multiple rows/columns",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TABLE test (v1 INTEGER, v2 INTEGER);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE SEQUENCE seq1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test VALUES (nextval('seq1'), 7), (nextval('seq1'), 11), (nextval('seq1'), 17);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test ORDER BY v1;",
					Expected: []sql.Row{
						{1, 7},
						{2, 11},
						{3, 17},
					},
				},
				{
					Query:    "INSERT INTO test VALUES (nextval('seq1'), nextval('seq1')), (nextval('seq1'), nextval('seq1'));",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test ORDER BY v1;",
					Expected: []sql.Row{
						{1, 7},
						{2, 11},
						{3, 17},
						{4, 5},
						{6, 7},
					},
				},
			},
		},
		{
			Name: "nextval() with double-quoted identifiers",
			SetUpScript: []string{
				"CREATE SEQUENCE test_sequence;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT nextval('test_sequence');",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query: "SELECT nextval('public.test_sequence');",
					Expected: []sql.Row{
						{2},
					},
				},
				{
					Query: `SELECT nextval('"test_sequence"');`,
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: `SELECT nextval('public."test_sequence"');`,
					Expected: []sql.Row{
						{4},
					},
				},
			},
		},
		{
			Name: "nextval() in filter",
			Skip: true, // GMS seems to call nextval once and cache the value, which is incorrect here
			SetUpScript: []string{
				"CREATE TABLE test_serial (v1 SERIAL, v2 INTEGER);",
				"INSERT INTO test_serial (v2) VALUES (4), (5), (6);",
				"CREATE TABLE test_seq (v1 INTEGER, v2 INTEGER);",
				"CREATE SEQUENCE test_sequence OWNED BY test_seq.v1;",
				"INSERT INTO test_seq VALUES (nextval('test_sequence'), 4), (nextval('test_sequence'), 5), (nextval('test_sequence'), 6);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test_serial WHERE nextval('test_serial_v1_seq') = v2 ORDER BY v1;",
					Expected: []sql.Row{
						{1, 4},
						{2, 5},
						{3, 6},
					},
				},
				{
					Query:    "SELECT nextval('test_serial_v1_seq');",
					Expected: []sql.Row{{7}},
				},
				{
					Query: "SELECT * FROM test_seq WHERE nextval('test_sequence') = v2 ORDER BY v1;",
					Expected: []sql.Row{
						{1, 4},
						{2, 5},
						{3, 6},
					},
				},
				{
					Query:    "SELECT nextval('test_sequence');",
					Expected: []sql.Row{{7}},
				},
			},
		},
		{
			Name: "setval()",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1 MINVALUE 1 MAXVALUE 10 START 5 INCREMENT 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE SEQUENCE test2 MINVALUE 1 MAXVALUE 10 START 5 INCREMENT -1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT setval('test1', 2);",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT setval('test1', 10);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:       "SELECT nextval('test1');",
					ExpectedErr: "reached maximum value",
				},
				{
					Query:    "SELECT setval('test1', 10, false);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "SELECT setval('test1', 10, true);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:       "SELECT nextval('test1');",
					ExpectedErr: "reached maximum value",
				},
				{
					Query:    "SELECT setval('test2', 9);",
					Expected: []sql.Row{{9}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{8}},
				},
				{
					Query:    "SELECT setval('test2', 1);",
					Expected: []sql.Row{{1}},
				},
				{
					Query:       "SELECT nextval('test2');",
					ExpectedErr: "reached minimum value",
				},
				{
					Query:    "SELECT setval('test2', 1, false);",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT setval('test2', 1, true);",
					Expected: []sql.Row{{1}},
				},
				{
					Query:       "SELECT nextval('test2');",
					ExpectedErr: "reached minimum value",
				},
				{
					Query:    "CREATE SEQUENCE test3 MINVALUE 3 MAXVALUE 7 START 5 INCREMENT 1 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE SEQUENCE test4 MINVALUE 3 MAXVALUE 7 START 5 INCREMENT -1 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT setval('test3', 7, true);",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT nextval('test3');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT setval('test4', 3, true);",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "CREATE SEQUENCE test5;",
					Expected: []sql.Row{},
				},
				{
					// test with a double-quoted identifier
					Query:    `SELECT setval('public."test5"', 100, true);`,
					Expected: []sql.Row{{100}},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{101}},
				},
			},
		},
		{
			Name: "SERIAL",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TABLE test (pk SERIAL PRIMARY KEY, v1 INTEGER);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE test_small (pk SMALLSERIAL PRIMARY KEY, v1 INTEGER);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE test_big (pk BIGSERIAL PRIMARY KEY, v1 INTEGER);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test (v1) VALUES (2), (3), (5), (7), (11);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test_small (v1) VALUES (2), (3), (5), (7), (11);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test_big (v1) VALUES (2), (3), (5), (7), (11);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 2},
						{2, 3},
						{3, 5},
						{4, 7},
						{5, 11},
					},
				},
				{
					Query: "SELECT * FROM test_small;",
					Expected: []sql.Row{
						{1, 2},
						{2, 3},
						{3, 5},
						{4, 7},
						{5, 11},
					},
				},
				{
					Query: "SELECT * FROM test_big;",
					Expected: []sql.Row{
						{1, 2},
						{2, 3},
						{3, 5},
						{4, 7},
						{5, 11},
					},
				},
			},
		},
		{
			Name: "SERIAL type created in table of different schema",
			SetUpScript: []string{
				"CREATE SCHEMA myschema",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TABLE myschema.test (pk SERIAL PRIMARY KEY, v1 INTEGER);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO myschema.test (v1) VALUES (2), (3), (5), (7), (11);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM myschema.test;",
					Expected: []sql.Row{
						{1, 2},
						{2, 3},
						{3, 5},
						{4, 7},
						{5, 11},
					},
				},
				{
					Query:       "SELECT nextval('test_pk_seq');",
					ExpectedErr: `relation "test_pk_seq" does not exist`,
				},
				{
					Query: "SELECT nextval('myschema.test_pk_seq');",
					Expected: []sql.Row{
						{6},
					},
				},
				{
					Query: "SELECT nextval('postgres.myschema.test_pk_seq');",
					Expected: []sql.Row{
						{7},
					},
				},
			},
		},
		{
			Name: "Default emulating SERIAL",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE seq1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE test (pk INTEGER DEFAULT (nextval('seq1')), v1 INTEGER);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test (v1) VALUES (2), (3), (5), (7), (11);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test ORDER BY v1;",
					Expected: []sql.Row{
						{1, 2},
						{2, 3},
						{3, 5},
						{4, 7},
						{5, 11},
					},
				},
			},
		},
		{
			Name: "Default emulating SERIAL in non default schema",
			SetUpScript: []string{
				"CREATE SCHEMA myschema",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE myschema.seq1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE myschema.test (pk INTEGER DEFAULT (nextval('seq1')), v1 INTEGER);",
					Expected: []sql.Row{},
				},
				{
					Skip:     true, // TODO: relation "seq1" does not exist
					Query:    "INSERT INTO myschema.test (v1) VALUES (2), (3), (5), (7), (11);",
					Expected: []sql.Row{},
				},
				{
					Skip:  true, // TODO: unskip when INSERT above is unskipped
					Query: "SELECT * FROM myschema.test ORDER BY v1;",
					Expected: []sql.Row{
						{1, 2},
						{2, 3},
						{3, 5},
						{4, 7},
						{5, 11},
					},
				},
			},
		},
		{
			Name: "pg_sequence",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_sequence";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_sequence";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_sequence";`,
					ExpectedErr: "not",
				},
				{
					Query:    "CREATE SEQUENCE some_sequence;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE SEQUENCE another_sequence INCREMENT 3 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM pg_catalog.pg_sequence ORDER BY seqrelid;",
					Expected: []sql.Row{
						{2795592127, 20, 1, 1, 9223372036854775807, 1, 1, "f"},
						{3473123643, 20, 1, 3, 9223372036854775807, 1, 1, "t"},
					},
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT * FROM PG_catalog.pg_SEQUENCE ORDER BY seqrelid;",
					Expected: []sql.Row{
						{2795592127, 20, 1, 1, 9223372036854775807, 1, 1, "f"},
						{3473123643, 20, 1, 3, 9223372036854775807, 1, 1, "t"},
					},
				},
				{
					Query: "SELECT * FROM pg_catalog.pg_sequence WHERE seqrelid = 'some_sequence'::regclass;",
					Expected: []sql.Row{
						{2795592127, 20, 1, 1, 9223372036854775807, 1, 1, "f"},
					},
				},
				{
					Query: "SELECT * FROM pg_catalog.pg_sequence WHERE seqrelid = 'another_sequence'::regclass;",
					Expected: []sql.Row{
						{3473123643, 20, 1, 3, 9223372036854775807, 1, 1, "t"},
					},
				},
				{
					Query: "SELECT nextval('another_sequence');",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query: "SELECT nextval('another_sequence');",
					Expected: []sql.Row{
						{4},
					},
				},
			},
		},
		{
			Name: "DROP TABLE",
			SetUpScript: []string{
				"CREATE TABLE test (pk SERIAL PRIMARY KEY, v1 INTEGER);",
				"INSERT INTO test (v1) VALUES (2), (3), (5), (7), (11);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 2},
						{2, 3},
						{3, 5},
						{4, 7},
						{5, 11},
					},
				},
				{
					Query: "SELECT * FROM pg_catalog.pg_sequence;",
					Expected: []sql.Row{
						{3822699147, 23, 1, 1, 2147483647, 1, 1, "f"},
					},
				},
				{
					Query:    "DROP TABLE test;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM pg_catalog.pg_sequence;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "dolt_add, dolt_commit, dolt_branch",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT setval('test', 10);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{11}},
				},
				{
					Query: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING')",
					Expected: []sql.Row{
						{"", "public.test", "added", 1, 1},
					},
				},
				{
					Query:    "SELECT dolt_add('test');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT dolt_branch('other');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT setval('test', 20);",
					Expected: []sql.Row{{20}},
				},
				{
					Query:    "SELECT dolt_add('.');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'next')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{21}},
				},
				{
					Query:    "SELECT dolt_checkout('other');",
					Expected: []sql.Row{{`{0,"Switched to branch 'other'"}`}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{12}},
				},
				{
					Query:    "SELECT dolt_reset('--hard');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{12}},
				},
			},
		},
	})
}
