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

func TestInsert(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple insert",
			SetUpScript: []string{
				"CREATE TABLE mytable (id INT PRIMARY KEY, name TEXT)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
					SkipResultsCheck: true,
				},
				{
					Query:            "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM mytable order by id",
					Expected: []sql.Row{
						{1, "hello"},
						{2, "world"},
					},
				},
			},
		},
		{
			Name: "keyless insert",
			SetUpScript: []string{
				"CREATE TABLE mytable (id INT, name TEXT)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
					SkipResultsCheck: true,
				},
				{
					Query:            "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM mytable order by id",
					Expected: []sql.Row{
						{1, "hello"},
						{2, "world"},
					},
				},
			},
		},
		{
			Name: "on conflict clause",
			SetUpScript: []string{
				"CREATE TABLE mytable (id INT primary key, name TEXT)",
				"create table t2 (id int primary key, c1 text, c2 text)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "INSERT INTO mytable (id, name) VALUES (1, 'hello')",
					SkipResultsCheck: true,
				},
				{
					Query:            "INSERT INTO mytable (ID, naME) VALUES (2, 'world')",
					SkipResultsCheck: true,
				},
				{
					Query:            "INSERT INTO mytable (ID, naME) VALUES (1, 'world') ON CONFLICT (id) DO UPDATE SET name = 'world'",
					SkipResultsCheck: true,
				},
				{
					Query:            "INSERT INTO mytable (ID, naME) VALUES (2, 'hello') ON CONFLICT (id) DO UPDATE SET name = 'conflict'",
					SkipResultsCheck: true,
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (1, 'not inserted') ON CONFLICT (id) DO NOTHING",
				},
				{
					Query: "SELECT * FROM mytable order by id",
					Expected: []sql.Row{
						{1, "world"},
						{2, "conflict"},
					},
				},
				{
					Query: "INSERT INTO mytable (ID, naME) VALUES (1, 'hello') ON CONFLICT (id) DO UPDATE set name = concat('new', name)",
				},
				{
					Query: "SELECT * FROM mytable order by id",
					Expected: []sql.Row{
						{1, "newworld"},
						{2, "conflict"},
					},
				},
				{
					Query:            "INSERT INTO t2 (id, c1, c2) VALUES (1, 'hello', 'world'), (2, 'world', 'hello')",
					SkipResultsCheck: true,
				},
				{
					Query:            "INSERT INTO t2 (id, c1, c2) VALUES (1, 'hello', 'world') ON CONFLICT (id) DO UPDATE SET c1 = 'conflict', c2 = c1",
					SkipResultsCheck: true,
				},
				{
					Query:            "INSERT INTO t2 (id, c1, c2) VALUES (2, 'hello', 'world') ON CONFLICT (id) DO UPDATE SET c2 = c1",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM t2 order by id",
					Expected: []sql.Row{
						{1, "conflict", "conflict"},
						{2, "world", "world"},
					},
				},
			},
		},
		{
			Name: "null and unspecified default values",
			SetUpScript: []string{
				"CREATE TABLE t (i INT DEFAULT NULL, j INT)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "INSERT INTO t VALUES (default, default)",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM t",
					Expected: []sql.Row{
						{nil, nil},
					},
				},
			},
		},
		{
			Name: "implicit default values",
			SetUpScript: []string{
				"CREATE TABLE t (i INT DEFAULT 123, j INT default 456);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "INSERT INTO t DEFAULT VALUES;",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM t",
					Expected: []sql.Row{
						{123, 456},
					},
				},
			},
		},
		{
			Name: "types",
			SetUpScript: []string{
				`create table child (i2 int2, i4 int4, i8 int8, f float, d double precision, v varchar, vl varchar(100), t text, j json, ts timestamp);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `insert into child values (1, 2, 3, 4.5, 6.7, 'hello', 'world', 'text', '{"a": 1}', '2021-01-01 00:00:00');`,
				},
				{
					Query: `select * from child;`,
					Expected: []sql.Row{
						{int16(1), int32(2), int64(3), float32(4.5), float64(6.7), "hello", "world", "text", `{"a": 1}`, "2021-01-01 00:00:00"},
					},
				},
			},
		},
		{
			Name: "insert returning",
			SetUpScript: []string{
				"CREATE TABLE t (i serial, j INT)",
				"CREATE TABLE u (u uuid DEFAULT 'ac1f3e2d-1e4b-4d3e-8b1f-2b7f1e7f0e3d', j INT)",
				"CREATE TABLE s (v1 varchar DEFAULT 'hello', v2 varchar DEFAULT 'world')",
				"CREATE SCHEMA ts",
				"CREATE TABLE ts.t (i serial, j INT)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i",
					Expected: []sql.Row{
						{1}, {2}, {3},
					},
				},
				{
					Query: "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i+3",
					Expected: []sql.Row{
						{7}, {8}, {9},
					},
				},
				{
					Query: "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i+j, j-3*i",
					Expected: []sql.Row{
						{12, -16}, {14, -18}, {16, -20},
					},
				},
				{
					Query: "INSERT INTO u (j) VALUES (5), (6), (7) RETURNING u",
					Expected: []sql.Row{
						{"ac1f3e2d-1e4b-4d3e-8b1f-2b7f1e7f0e3d"}, {"ac1f3e2d-1e4b-4d3e-8b1f-2b7f1e7f0e3d"}, {"ac1f3e2d-1e4b-4d3e-8b1f-2b7f1e7f0e3d"},
					},
				},
				{
					Query: "INSERT INTO s (v2) VALUES (' a') RETURNING concat(v1, v2)",
					Expected: []sql.Row{
						{"hello a"},
					},
				},
				{
					Query: "INSERT INTO s (v1) VALUES ('sup ') RETURNING concat(v1, v2)",
					Expected: []sql.Row{
						{"sup world"},
					},
				},
				{
					Query: "INSERT INTO s (v2, v1) VALUES ('def', 'abc'), ('xyz', 'uvw') RETURNING concat(v1, v2), concat(v2, v1), 100",
					Expected: []sql.Row{
						{"abcdef", "defabc", 100},
						{"uvwxyz", "xyzuvw", 100},
					},
				},
				{
					Query:       "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i, doesnotexist",
					ExpectedErr: "could not be found",
				},
				{
					Query:       "INSERT INTO t (j) VALUES (5), (6), (7) RETURNING i, doesnotexist(j)",
					ExpectedErr: "function: 'doesnotexist' not found",
				},
				{
					Query:    "INSERT INTO public.t (j) VALUES (8) RETURNING t.j",
					Expected: []sql.Row{{8}},
				},
				{
					Query:    "INSERT INTO public.t (j) VALUES (9) RETURNING public.t.j",
					Expected: []sql.Row{{9}},
				},
				{
					Query:    "INSERT INTO ts.t (j) VALUES (10) RETURNING ts.t.j",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "INSERT INTO public.t (j) VALUES ($1) RETURNING j;",
					BindVars: []any{11},
					Expected: []sql.Row{{11}},
				},
				{
					Query:    "INSERT INTO public.t (j) VALUES ($1) RETURNING t.j;",
					BindVars: []any{12},
					Expected: []sql.Row{{12}},
				},
				{
					Query:    "INSERT INTO public.t (j) VALUES ($1) RETURNING public.t.j;",
					BindVars: []any{13},
					Expected: []sql.Row{{13}},
				},
			},
		},
		{
			Name: "insert iso8601 timestamptz literal",
			SetUpScript: []string{
				"CREATE TABLE django_migrations (id serial primary key, app varchar, name varchar, applied timestamptz)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO "django_migrations" ("app", "name", "applied") VALUES ('contenttypes', '0001_initial', '2025-03-24T19:21:59.690479+00:00'::timestamptz) RETURNING "django_migrations"."id"`,
					Expected: []sql.Row{{1}},
				},
			},
		},
	})
}
