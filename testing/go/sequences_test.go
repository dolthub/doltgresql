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
			Name: "identity generated by default",
			SetUpScript: []string{
				`CREATE TABLE "django_migrations" (
    "id" bigint NOT NULL PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
		"app" varchar(255) NOT NULL,
		"name" varchar(255) NOT NULL,
		"applied" timestamp with time zone NOT NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id"`,
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query: `INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id"`,
					Expected: []sql.Row{
						{2},
					},
				},
				{
					Query: `INSERT INTO "django_migrations" ("id", "app", "name", "applied") VALUES (100, 'contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id"`,
					Expected: []sql.Row{
						{100},
					},
				},
			},
		},
		{
			Name: "identity generated by default with sequence options",
			Skip: true, // not supported yet, need to add sequence info into DML node given to GMS
			SetUpScript: []string{
				`CREATE TABLE "django_migrations" (
    "id" bigint NOT NULL PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY (START WITH 100 INCREMENT BY 2),
		"app" varchar(255) NOT NULL,
		"name" varchar(255) NOT NULL,
		"applied" timestamp with time zone NOT NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id"`,
					Expected: []sql.Row{
						{100},
					},
				},
				{
					Query: `INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-25T17:45:54.794344+00:00'::timestamptz) RETURNING "django_migrations"."id"`,
					Expected: []sql.Row{
						{102},
					},
				},
			},
		},
		{
			Name: "insert on a different branch",
			Skip: true, // currently dies on creating the sequence on a non-current DB table
			SetUpScript: []string{
				"create table test (pk serial primary key, v1 int);",
				"insert into test (v1) values (2), (3), (5), (7), (11);",
				"call dolt_branch('b1');",
				`create table "postgres/b1".public.test2 (pk serial primary key, v1 int);`,
				`insert into "postgres/b1".public.test2 (v1) values (2), (3), (5), (7), (11);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT pk FROM test ORDER BY v1;",
					Expected: []sql.Row{
						{1},
						{2},
						{3},
						{4},
						{5},
					},
				},
				{
					Query: `SELECT pk FROM "postgres/b1".public.test2 ORDER BY v1;`,
					Expected: []sql.Row{
						{1},
						{2},
						{3},
						{4},
						{5},
					},
				},
			},
		},
		{
			Name: "dolt_add, dolt_branch, dolt_checkout, dolt_commit, dolt_reset",
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
		{
			Name: "dolt_clean",
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
					Query:    "SELECT setval('test1', 10);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{11}},
				},
				{
					Query:    "SELECT setval('test2', 10);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{11}},
				},
				{
					Query:    "SELECT dolt_add('test1');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.test1", "t", "new table"},
						{"public.test2", "f", "new table"},
					},
				},
				{
					Query:    "SELECT dolt_clean('test2');", // TODO: dolt_clean() requires a param, need to fix procedure to func conversion
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query: "SELECT * FROM dolt.status;",
					Expected: []sql.Row{
						{"public.test1", "t", "new table"},
					},
				},
			},
		},
		{
			Name: "dolt_merge",
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
					Query:    "SELECT length(dolt_commit('-Am', 'initial')::text) = 34;",
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
					Query:    "SELECT length(dolt_commit('-am', 'next')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT dolt_checkout('other');",
					Expected: []sql.Row{{`{0,"Switched to branch 'other'"}`}},
				},
				{
					Query:    "SELECT setval('test', 30);",
					Expected: []sql.Row{{30}},
				},
				{
					Query:    "SELECT length(dolt_commit('-am', 'next2')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT dolt_checkout('main');",
					Expected: []sql.Row{{`{0,"Switched to branch 'main'"}`}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{21}},
				},
				{
					Query:    "SELECT dolt_reset('--hard');",
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT strpos(dolt_merge('other')::text, 'merge successful') > 32;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{31}},
				},
			},
		},
		{
			Name: "Information Schema & DIFF_STAT regression testing",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE "user" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, PRIMARY KEY ("id"));`,
					Expected: []sql.Row{},
				},
				{
					Query: `CREATE TABLE "call" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "state" character varying NOT NULL,
  "content" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "ended_at" timestamptz NULL,
  "user" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "call_user_user_fk" FOREIGN KEY ("user") REFERENCES "user" ("id") ON DELETE NO ACTION
);`,
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM information_schema.key_column_usage where constraint_schema <> 'pg_catalog';",
					Expected: []sql.Row{
						{"postgres", "public", "PRIMARY", "postgres", "public", "call", "id", 1, nil, nil, nil, nil},
						{"postgres", "public", "PRIMARY", "postgres", "public", "user", "id", 1, nil, nil, nil, nil},
						{"postgres", "public", "call_user_user_fk", "postgres", "public", "call", "user", 1, 1, "postgres", "user", "id"},
					},
				},
				{
					Query: "SELECT * FROM DOLT_DIFF_STAT('HEAD', 'WORKING');",
					Expected: []sql.Row{
						{"public.call_id_seq", 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{"public.user_id_seq", 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					},
				},
				{ // This is the same as running "\d" in PSQL
					Query: `SELECT n.nspname as "Schema",
  c.relname as "Name",
  CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'materialized view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 't' THEN 'TOAST table' WHEN 'f' THEN 'foreign table' WHEN 'p' THEN 'partitioned table' WHEN 'I' THEN 'partitioned index' END as "Type",
  pg_catalog.pg_get_userbyid(c.relowner) as "Owner"
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
     LEFT JOIN pg_catalog.pg_am am ON am.oid = c.relam
WHERE c.relkind IN ('r','p','v','m','S','f','')
      AND n.nspname <> 'pg_catalog'
      AND n.nspname !~ '^pg_toast'
      AND n.nspname <> 'information_schema'
  AND pg_catalog.pg_table_is_visible(c.oid)
ORDER BY 1,2;`,
					Expected: []sql.Row{
						{"public", "call", "table", "postgres"},
						{"public", "call_id_seq", "sequence", "postgres"},
						{"public", "dolt_branches", "table", "postgres"},
						{"public", "dolt_column_diff", "table", "postgres"},
						{"public", "dolt_commit_ancestors", "table", "postgres"},
						{"public", "dolt_commit_diff_call", "table", "postgres"},
						{"public", "dolt_commit_diff_user", "table", "postgres"},
						{"public", "dolt_commits", "table", "postgres"},
						{"public", "dolt_conflicts", "table", "postgres"},
						{"public", "dolt_conflicts_call", "table", "postgres"},
						{"public", "dolt_conflicts_user", "table", "postgres"},
						{"public", "dolt_constraint_violations", "table", "postgres"},
						{"public", "dolt_constraint_violations_call", "table", "postgres"},
						{"public", "dolt_constraint_violations_user", "table", "postgres"},
						{"public", "dolt_diff", "table", "postgres"},
						{"public", "dolt_diff_call", "table", "postgres"},
						{"public", "dolt_diff_user", "table", "postgres"},
						{"public", "dolt_history_call", "table", "postgres"},
						{"public", "dolt_history_user", "table", "postgres"},
						{"public", "dolt_log", "table", "postgres"},
						{"public", "dolt_merge_status", "table", "postgres"},
						{"public", "dolt_remote_branches", "table", "postgres"},
						{"public", "dolt_remotes", "table", "postgres"},
						{"public", "dolt_schema_conflicts", "table", "postgres"},
						{"public", "dolt_status", "table", "postgres"},
						{"public", "dolt_tags", "table", "postgres"},
						{"public", "dolt_workspace_call", "table", "postgres"},
						{"public", "dolt_workspace_user", "table", "postgres"},
						{"public", "user", "table", "postgres"},
						{"public", "user_id_seq", "sequence", "postgres"},
					},
				},
			},
		},
		{
			Name: "ALTER COLUMN ADD GENERATED BY DEFAULT",
			SetUpScript: []string{
				"CREATE TABLE public.test1 (id int2 NOT NULL, name character varying(150) NOT NULL);",
				"CREATE TABLE public.test2 (id int4 NOT NULL, name character varying(150) NOT NULL);",
				"CREATE TABLE public.test3 (id int8 NOT NULL, name character varying(150) NOT NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE public.test1 ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (SEQUENCE NAME public.test1_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "ALTER TABLE public.test2 ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (SEQUENCE NAME public.test2_id_seq START WITH 10 INCREMENT BY 5);",
					Expected: []sql.Row{},
				},
				{
					Query:    "ALTER TABLE public.test3 ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (SEQUENCE NAME public.test3_id_seq START WITH 100 INCREMENT BY -1 MAXVALUE 100);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO public.test1 (name) VALUES ('abc'), ('def');",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO public.test2 (name) VALUES ('abc'), ('def');",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO public.test3 (name) VALUES ('abc'), ('def');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM public.test1;",
					Expected: []sql.Row{{1, "abc"}, {2, "def"}},
				},
				{
					Query:    "SELECT * FROM public.test2;",
					Expected: []sql.Row{{10, "abc"}, {15, "def"}},
				},
				{
					Query:    "SELECT * FROM public.test3;",
					Expected: []sql.Row{{100, "abc"}, {99, "def"}},
				},
			},
		},
		{
			Name: "ALTER SEQUENCE OWNED BY",
			SetUpScript: []string{
				"CREATE TABLE test (id int4 NOT NULL);",
				"CREATE SEQUENCE seq1;",
				"CREATE SEQUENCE seq2;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT nextval('seq1');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('seq2');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "ALTER SEQUENCE seq1 OWNED BY test.id;",
					Expected: []sql.Row{},
				},
				{
					Query:    "ALTER SEQUENCE seq2 OWNED BY test.id;",
					Expected: []sql.Row{},
				},
				{ // Setting OWNED BY back to NONE ensures that we properly handle this case
					Query:    "ALTER SEQUENCE seq2 OWNED BY NONE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP TABLE test;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT nextval('seq1');",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "SELECT nextval('seq2');",
					Expected: []sql.Row{{2}},
				},
			},
		},
	})
}
