// Copyright 2025 Dolthub, Inc.
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

func TestMerge(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple merge",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT, b INT, c INT, PRIMARY KEY (a))",
				"CREATE TABLE t2 (a INT, b INT, c INT, PRIMARY KEY (a))",
				"INSERT INTO t1 VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9)",
				"INSERT INTO t2 VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9)",
				"SELECT DOLT_COMMIT('-Am', 'intial commit')",
				"SELECT DOLT_BRANCH('branch1')",
				"SELECT DOLT_BRANCH('branch2')",
				"SELECT DOLT_CHECKOUT('branch1')",
				"INSERT INTO t1 VALUES (10, 11, 12)",
				"INSERT INTO t2 VALUES (10, 11, 12)",
				"SELECT DOLT_COMMIT('-Am', 'added 10')",
				"SELECT DOLT_CHECKOUT('branch2')",
				"INSERT INTO t1 VALUES (20, 21, 22)",
				"INSERT INTO t2 VALUES (20, 21, 22)",
				"SELECT DOLT_COMMIT('-Am', 'added 20')",
				"SELECT DOLT_CHECKOUT('main')",
				"SELECT DOLT_MERGE('branch1')",
				"SELECT DOLT_MERGE('branch2')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM t1",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
						{7, 8, 9},
						{10, 11, 12},
						{20, 21, 22},
					},
				},
				{
					Query: "SELECT * FROM t2",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
						{7, 8, 9},
						{10, 11, 12},
						{20, 21, 22},
					},
				},
			},
		},
		{
			Name: "merge independent JSON document changes",
			SetUpScript: []string{
				"CREATE TABLE documents_json (id INT PRIMARY KEY, doc JSON)",
				"CREATE TABLE documents_jsonb (id INT PRIMARY KEY, doc JSONB)",
				"INSERT INTO documents_json VALUES (1, '{\"a\": 1, \"b\": 2}')",
				"INSERT INTO documents_jsonb VALUES (1, '{\"a\": 1, \"b\": 2}')",
				"SELECT DOLT_COMMIT('-Am', 'base')",
				"SELECT DOLT_BRANCH('left')",
				"SELECT DOLT_BRANCH('right')",
				"SELECT DOLT_CHECKOUT('left')",
				"UPDATE documents_json SET doc = '{\"a\": 100, \"b\": 2}' WHERE id = 1",
				"UPDATE documents_jsonb SET doc = doc || '{\"a\": 100}' WHERE id = 1",
				"SELECT DOLT_COMMIT('-am', 'left')",
				"SELECT DOLT_CHECKOUT('right')",
				"UPDATE documents_json SET doc = '{\"a\": 1, \"b\": 200}' WHERE id = 1",
				"UPDATE documents_jsonb SET doc = doc || '{\"b\": 200}' WHERE id = 1",
				"SELECT DOLT_COMMIT('-am', 'right')",
				"SELECT DOLT_CHECKOUT('left')",
				"SELECT DOLT_MERGE('right')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT doc::TEXT FROM documents_json",
					Expected: []sql.Row{{"{\"a\":100,\"b\":200}"}},
				},
				{
					Query:    "SELECT doc::TEXT FROM documents_jsonb",
					Expected: []sql.Row{{"{\"a\": 100, \"b\": 200}"}},
				},
				{
					Query:    "SELECT COUNT(*) FROM dolt_conflicts",
					Expected: []sql.Row{{int64(0)}},
				},
			},
		},
		{
			Name: "merge with check expressions and column defaults",
			SetUpScript: []string{
				"SET timezone TO 'UTC';",
				"CREATE TABLE t1 (a INT, b timestamptz default '2020-01-01 00:00:00'::timestamptz, PRIMARY KEY (a))",
				"ALTER TABLE t1 ADD CONSTRAINT check_b CHECK (b >= '2020-01-01 00:00:00'::timestamptz)",
				"INSERT INTO t1 VALUES (1, '2020-01-02 00:00:00'), (2, '2020-01-03 00:00:00')",
				"SELECT DOLT_COMMIT('-Am', 'intial commit')",
				"SELECT DOLT_BRANCH('branch1')",
				"SELECT DOLT_BRANCH('branch2')",
				"SELECT DOLT_CHECKOUT('branch1')",
				"INSERT INTO t1 VALUES (3, '2020-01-04 00:00:00')",
				"SELECT DOLT_COMMIT('-Am', 'added 3')",
				"SELECT DOLT_CHECKOUT('branch2')",
				"INSERT INTO t1 VALUES (4, '2020-01-05 00:00:00')",
				"SELECT DOLT_COMMIT('-Am', 'added 4')",
				"SELECT DOLT_CHECKOUT('main')",
				"SELECT DOLT_MERGE('branch1')",
				"SELECT DOLT_MERGE('branch2')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM t1 order by a",
					Expected: []sql.Row{
						{1, "2020-01-02 00:00:00+00"},
						{2, "2020-01-03 00:00:00+00"},
						{3, "2020-01-04 00:00:00+00"},
						{4, "2020-01-05 00:00:00+00"},
					},
				},
				{
					// make sure the check constraint is still there
					Query:       "INSERT INTO t1 VALUES (5, '2019-12-31 00:00:00')",
					ExpectedErr: "Check constraint",
				},
			},
		},
		{
			Name: "merge with unique constraints and foreign keys",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT, b INT, PRIMARY KEY (a), unique (b))",
				"CREATE TABLE t2 (a INT, b INT, PRIMARY KEY (a), foreign key (b) references t1(b))",
				"INSERT INTO t1 VALUES (1, 2), (4, 5), (7, 8)",
				"INSERT INTO t2 VALUES (1, 2), (4, 5), (7, 8)",
				"SELECT DOLT_COMMIT('-Am', 'intial commit')",
				"SELECT DOLT_BRANCH('branch1')",
				"SELECT DOLT_BRANCH('branch2')",
				"SELECT DOLT_CHECKOUT('branch1')",
				"INSERT INTO t1 VALUES (10, 11)",
				"INSERT INTO t2 VALUES (10, 11)",
				"SELECT DOLT_COMMIT('-Am', 'added 10')",
				"SELECT DOLT_CHECKOUT('branch2')",
				"INSERT INTO t1 VALUES (20, 21)",
				"INSERT INTO t2 VALUES (20, 21)",
				"SELECT DOLT_COMMIT('-Am', 'added 20')",
				"SELECT DOLT_CHECKOUT('main')",
				"SELECT DOLT_MERGE('branch1')",
				"SELECT DOLT_MERGE('branch2')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM t1 order by a",
					Expected: []sql.Row{
						{1, 2},
						{4, 5},
						{7, 8},
						{10, 11},
						{20, 21},
					},
				},
				{
					// make sure the unique constraint is still there
					Query:       "INSERT INTO t1 VALUES (100, 2)",
					ExpectedErr: "unique key",
				},
				{
					// make sure the foreign key constraint is still there
					Query:       "INSERT INTO t2 VALUES (100, 200)",
					ExpectedErr: "Foreign key violation",
				},
			},
		},
		{
			Name: "merge a branch that created a type",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT PRIMARY KEY)",
				"SELECT DOLT_COMMIT('-Am', 'initial commit')",
				"SELECT DOLT_CHECKOUT('-b', 'branch1')",
				"CREATE TYPE type1 AS ENUM ('a', 'b')",
				"CREATE TABLE t2 (a INT PRIMARY KEY, b type1)",
				"INSERT INTO t2 VALUES (1, 'a')",
				"SELECT DOLT_COMMIT('-Am', 'added type1')",
				"SELECT DOLT_CHECKOUT('main')",
				"INSERT INTO t1 VALUES (1)",
				"SELECT DOLT_COMMIT('-Am', 'added 1')",
				"SELECT DOLT_MERGE('branch1')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM t2",
					Expected: []sql.Row{{1, "a"}},
				},
				{
					Query:    "SELECT 'b'::type1",
					Expected: []sql.Row{{"b"}},
				},
			},
		},
		{
			Name: "merge columns with defaults into rows added on the other branch",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, a TEXT NOT NULL)",
				"INSERT INTO t VALUES (1, 'a1')",
				"SELECT DOLT_COMMIT('-Am', 'base')",
				"SELECT DOLT_BRANCH('br')",
				"ALTER TABLE t ADD COLUMN n1 NUMERIC DEFAULT 0",
				"ALTER TABLE t ADD COLUMN n2 NUMERIC(10,2) NOT NULL DEFAULT 0.00",
				"ALTER TABLE t ADD COLUMN n3 NUMERIC DEFAULT 1 + 1",
				"ALTER TABLE t ADD COLUMN n4 NUMERIC DEFAULT 2.5::numeric",
				"ALTER TABLE t ADD COLUMN n5 NUMERIC DEFAULT abs(-2)",
				"ALTER TABLE t ADD COLUMN n6 NUMERIC DEFAULT '3'",
				"ALTER TABLE t ADD COLUMN n7 NUMERIC(10,2) DEFAULT 1 + 1",
				"ALTER TABLE t ADD COLUMN i1 INT DEFAULT 7",
				"ALTER TABLE t ADD COLUMN i2 INT DEFAULT 2.5",
				"ALTER TABLE t ADD COLUMN i3 BIGINT DEFAULT 5",
				"ALTER TABLE t ADD COLUMN i4 SMALLINT DEFAULT -1",
				"ALTER TABLE t ADD COLUMN i5 BIGINT DEFAULT 1 + 1",
				"ALTER TABLE t ADD COLUMN i6 INT DEFAULT '4'",
				"ALTER TABLE t ADD COLUMN f1 DOUBLE PRECISION DEFAULT 1.5",
				"ALTER TABLE t ADD COLUMN f2 REAL DEFAULT 0",
				"SELECT DOLT_COMMIT('-Am', 'add columns')",
				"SELECT DOLT_CHECKOUT('br')",
				"INSERT INTO t VALUES (5, 'b5')",
				"SELECT DOLT_COMMIT('-Am', 'row 5')",
				"SELECT DOLT_CHECKOUT('-b', 'br2')",
				"INSERT INTO t VALUES (6, 'b6')",
				"SELECT DOLT_COMMIT('-Am', 'row 6')",
				"SELECT DOLT_CHECKOUT('main')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT DOLT_MERGE('br', '--no-ff', '-m', 'merge br')",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM t ORDER BY id",
					Expected: []sql.Row{
						{1, "a1", Numeric("0"), Numeric("0.00"), Numeric("2"), Numeric("2.5"), Numeric("2"), Numeric("3"), Numeric("2.00"), 7, 3, 5, -1, 2, 4, 1.5, 0.0},
						{5, "b5", Numeric("0"), Numeric("0.00"), Numeric("2"), Numeric("2.5"), Numeric("2"), Numeric("3"), Numeric("2.00"), 7, 3, 5, -1, 2, 4, 1.5, 0.0},
					},
				},
				{
					Query:            "SELECT DOLT_CHECKOUT('br2')",
					SkipResultsCheck: true,
				},
				{
					Query:            "SELECT DOLT_MERGE('main', '--no-ff', '-m', 'merge main')",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM t ORDER BY id",
					Expected: []sql.Row{
						{1, "a1", Numeric("0"), Numeric("0.00"), Numeric("2"), Numeric("2.5"), Numeric("2"), Numeric("3"), Numeric("2.00"), 7, 3, 5, -1, 2, 4, 1.5, 0.0},
						{5, "b5", Numeric("0"), Numeric("0.00"), Numeric("2"), Numeric("2.5"), Numeric("2"), Numeric("3"), Numeric("2.00"), 7, 3, 5, -1, 2, 4, 1.5, 0.0},
						{6, "b6", Numeric("0"), Numeric("0.00"), Numeric("2"), Numeric("2.5"), Numeric("2"), Numeric("3"), Numeric("2.00"), 7, 3, 5, -1, 2, 4, 1.5, 0.0},
					},
				},
			},
		},
		{
			Name: "merge a numeric default without committing the merge",
			SetUpScript: []string{
				"CREATE TABLE public.t (id integer PRIMARY KEY, a text NOT NULL)",
				"SELECT dolt_commit('-Am', 'base')",
				"SELECT dolt_branch('br')",
				"ALTER TABLE public.t ADD COLUMN x numeric DEFAULT 0",
				"SELECT dolt_commit('-Am', 'add x')",
				"SELECT dolt_checkout('br')",
				"INSERT INTO public.t VALUES (5, 'b5')",
				"SELECT dolt_commit('-Am', 'row 5')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "BEGIN",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT dolt_merge('main', '--no-ff', '--no-commit')",
					Expected: []sql.Row{{[]any{"", int64(0), int64(0), "merge successful"}}},
				},
				{
					Query:    "SELECT * FROM public.t",
					Expected: []sql.Row{{5, "b5", Numeric("0")}},
				},
			},
		},
		{
			Name: "merge a default set in a later commit",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, a TEXT NOT NULL)",
				"INSERT INTO t VALUES (1, 'a1')",
				"SELECT DOLT_COMMIT('-Am', 'base')",
				"SELECT DOLT_BRANCH('br')",
				"ALTER TABLE t ADD COLUMN x NUMERIC",
				"SELECT DOLT_COMMIT('-Am', 'add x')",
				"ALTER TABLE t ALTER COLUMN x SET DEFAULT 0",
				"SELECT DOLT_COMMIT('-Am', 'default x')",
				"SELECT DOLT_CHECKOUT('br')",
				"INSERT INTO t VALUES (5, 'b5')",
				"SELECT DOLT_COMMIT('-Am', 'row 5')",
				"SELECT DOLT_CHECKOUT('main')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT DOLT_MERGE('br', '--no-ff', '-m', 'merge br')",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT * FROM t ORDER BY id",
					Expected: []sql.Row{{1, "a1", nil}, {5, "b5", Numeric("0")}},
				},
			},
		},
		{
			Name: "merge a default set in the same commit as its column",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, a TEXT NOT NULL)",
				"INSERT INTO t VALUES (1, 'a1')",
				"SELECT DOLT_COMMIT('-Am', 'base')",
				"SELECT DOLT_BRANCH('br')",
				"ALTER TABLE t ADD COLUMN x NUMERIC",
				"ALTER TABLE t ALTER COLUMN x SET DEFAULT 0",
				"SELECT DOLT_COMMIT('-Am', 'add x')",
				"SELECT DOLT_CHECKOUT('br')",
				"INSERT INTO t VALUES (5, 'b5')",
				"SELECT DOLT_COMMIT('-Am', 'row 5')",
				"SELECT DOLT_CHECKOUT('main')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT DOLT_MERGE('br', '--no-ff', '-m', 'merge br')",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT * FROM t ORDER BY id",
					Expected: []sql.Row{{1, "a1", nil}, {5, "b5", Numeric("0")}},
				},
			},
		},
		{
			Name: "merge a default whose column was backfilled",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, a TEXT NOT NULL)",
				"INSERT INTO t VALUES (1, 'a1')",
				"SELECT DOLT_COMMIT('-Am', 'base')",
				"SELECT DOLT_BRANCH('br')",
				"ALTER TABLE t ADD COLUMN x NUMERIC DEFAULT 0",
				"UPDATE t SET x = 0 WHERE x IS NULL",
				"SELECT DOLT_COMMIT('-Am', 'add x')",
				"SELECT DOLT_CHECKOUT('br')",
				"INSERT INTO t VALUES (5, 'b5')",
				"SELECT DOLT_COMMIT('-Am', 'row 5')",
				"SELECT DOLT_CHECKOUT('main')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT DOLT_MERGE('br', '--no-ff', '-m', 'merge br')",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT * FROM t ORDER BY id",
					Expected: []sql.Row{{1, "a1", Numeric("0")}, {5, "b5", Numeric("0")}},
				},
			},
		},
	})
}
