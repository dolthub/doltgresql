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

func TestRecords(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Record cannot be used as column type",
			SetUpScript: []string{
				"CREATE TABLE t2 (pk INT PRIMARY KEY, c1 VARCHAR(100));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "CREATE TABLE t (pk INT PRIMARY KEY, r RECORD);",
					ExpectedErr: `column "r" has pseudo-type record`,
				},
				{
					Query:       "ALTER TABLE t2 ADD COLUMN c2 RECORD;",
					ExpectedErr: `column "c2" has pseudo-type record`,
				},
				{
					Query:       "ALTER TABLE t2 ALTER COLUMN c1 TYPE RECORD;",
					ExpectedErr: `column "c1" has pseudo-type record`,
				},
				{
					Query:       "CREATE DOMAIN my_domain AS record;",
					ExpectedErr: `"record" is not a valid base type for a domain`,
				},
				{
					Query:       "CREATE SEQUENCE my_seq AS record;",
					ExpectedErr: "sequence type must be smallint, integer, or bigint",
				},
				{
					Query:       "CREATE TYPE outer_type AS (id int, payload record);",
					ExpectedErr: `column "payload" has pseudo-type record`,
				},
			},
		},
		{
			Name: "Casting to record",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select row(1, 1)::record;",
					Expected: []sql.Row{{[]any{1, 1}}},
				},
			},
		},
		{
			// TODO: Wrapping table rows with ROW() is not supported yet. Planbuilder assumes the
			//       table alias is a column name and not a table.
			Name: "ROW() wrapping table rows",
			SetUpScript: []string{
				"create table users (name text, location text, age int);",
				"insert into users values ('jason', 'SEA', 42), ('max', 'SFO', 31);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: ERROR: column "p" could not be found in any table in scope
					Skip:     true,
					Query:    "select row(p) from users p;",
					Expected: []sql.Row{{`("(jason,SEA,44)")`}, {`("(max,SFO,31)")`}},
				},
				{
					// TODO: ERROR: name resolution on this statement is not yet supported
					Skip:     true,
					Query:    "select row(p.*, 42) from users p;",
					Expected: []sql.Row{{`(jason,SEA,42,42)`}, {`(max,SFO,31,42)`}},
				},
				{
					// TODO: ERROR: (E).x is not yet supported
					Skip:     true,
					Query:    "SELECT (u).location FROM users u;",
					Expected: []sql.Row{{"SEA"}, {"SFO"}},
				},
			},
		},
		{
			Name: "ROW() wrapping values",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ROW(1, 2, 3) as myRow;",
					Expected: []sql.Row{{[]any{1, 2, 3}}},
				},
				{
					Query:    "SELECT (4, 5, 6) as myRow;",
					Expected: []sql.Row{{[]any{4, 5, 6}}},
				},
				{
					Query:    "SELECT (NULL, 'foo', NULL) as myRow;",
					Expected: []sql.Row{{[]any{nil, "foo", nil}}},
				},
				{
					Query:    "SELECT (NULL, (1 > 0), 'baz') as myRow;",
					Expected: []sql.Row{{[]any{nil, true, "baz"}}},
				},
			},
		},
		{
			Name: "ROW() equality and comparison",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ROW(1, 'x') = ROW(1, 'x');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, 'x') = ROW(1, 'y');",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(1, NULL) = ROW(1, 1);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT ROW(1, 2) < ROW(1, 3);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, 2) < ROW(2, NULL);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(2, 2) < ROW(2, NULL);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT ROW(2, 2, 1) < ROW(2, NULL, 2);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT ROW(1, 2) < ROW(NULL, 3);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT ROW(NULL, NULL, NULL) < ROW(NULL, NULL, NULL);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT ROW(1, 2) <= ROW(1, 3);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, 2) <= ROW(1, 2);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, NULL) <= ROW(1, 2);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT ROW(2, 1) > ROW(1, 999);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(2, 1) > ROW(1, NULL);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(2, 1) >= ROW(1, 999);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(2, 1) >= ROW(2, 1);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(NULL, 1) >= ROW(2, 1);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT ROW(1, 2) != ROW(3, 4);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, 2) != ROW(NULL, 4);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(NULL, 4) != ROW(NULL, 4);",
					Expected: []sql.Row{{nil}},
				},
				{
					// TODO: IS NOT DISTINCT FROM is not yet supported
					Skip:     true,
					Query:    "SELECT ROW(1, NULL) IS NOT DISTINCT FROM ROW(1, NULL);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, '2') = ROW(1, 2::TEXT);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, 1) = ROW(1, 1);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, 1) != ROW(1, 1);",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(1, 1) < ROW(1, 1);",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(1, 1) <= ROW(1, 1);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, 1) > ROW(1, 1);",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(1, 1) = ROW(1, 1);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(1, 2, null, 4) = ROW(null, 2, 3, 5);",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(1, 2, null, 4) != ROW(null, 2, 3, 5);",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "ROW() use inserting and selecting composite rows",
			SetUpScript: []string{
				"CREATE TYPE user_info AS (id INT, name TEXT, email TEXT);",
				"CREATE TABLE accounts (info user_info);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO accounts VALUES (ROW(1, 'alice', 'a@example.com'));",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT info FROM accounts;",
					Expected: []sql.Row{{"(1,alice,a@example.com)"}},
				},
				{
					Query:    "SELECT (a.info).name FROM accounts a;",
					Expected: []sql.Row{{"alice"}},
				},
			},
		},
		{
			Name: "ROW() use in WHERE clause",
			SetUpScript: []string{
				"create table users (id int primary key, name text, email text);",
				"insert into users values (1, 'John', 'j@a.com'), (2, 'Joe', 'joe@joe.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM users WHERE ROW(id, name, email) = ROW(1, 'John', 'j@a.com');",
					Expected: []sql.Row{{1, "John", "j@a.com"}},
				},
				{
					// TODO: IS NOT DISTINCT FROM is not yet supported
					Skip:     true,
					Query:    "SELECT * FROM users WHERE ROW(id, name) IS NOT DISTINCT FROM ROW(2, 'Jane');",
					Expected: []sql.Row{{2, "Joe", "joe@joe.com"}},
				},
			},
		},
		{
			Name: "ROW() casting and type inference",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ROW(1, 'a')::record;",
					Expected: []sql.Row{{[]any{1, "a"}}},
				},
				{
					Query:       "SELECT ROW(1, 2) = ROW(1, 'two');",
					ExpectedErr: "invalid input syntax",
				},
				{
					Query:    "SELECT ROW(1, 2) = ROW(1, '2');",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "ROW() error cases and edge conditions",
			SetUpScript: []string{
				"create table users (id int primary key, name text, email text);",
				"insert into users values (1, 'John', 'j@a.com'), (2, 'Joe', 'joe@joe.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT ROW(1, 2) = ROW(1);",
					ExpectedErr: "unequal number of entries",
				},
				{
					Query:       "SELECT ROW(1, 2) = ROW(1, 2, 3);",
					ExpectedErr: "unequal number of entries",
				},
				{
					Query:       "SELECT ROW(1, 2) < ROW(1);",
					ExpectedErr: "unequal number of entries",
				},
				{
					Query:       "SELECT ROW(1, 2) <= ROW(1);",
					ExpectedErr: "unequal number of entries",
				},
				{
					Query:       "SELECT ROW(1, 2) > ROW(1);",
					ExpectedErr: "unequal number of entries",
				},
				{
					Query:       "SELECT ROW(1, 2) >= ROW(1);",
					ExpectedErr: "unequal number of entries",
				},
				{
					Query:       "SELECT ROW(1, 2) != ROW(1);",
					ExpectedErr: "unequal number of entries",
				},
				{
					Query:    "SELECT NULL::record IS NULL",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(NULL) IS NULL",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(NULL, NULL, NULL) IS NULL;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(NULL, 42, NULL) IS NULL;",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(42) IS NULL",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(NULL) IS NOT NULL;",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(NULL, NULL) IS NOT NULL;",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(NULL, 1) IS NOT NULL;",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(1, 1) IS NOT NULL;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(42) IS NOT NULL;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(id, name), COUNT(*) FROM users GROUP BY ROW(id, name);",
					Expected: []sql.Row{{[]any{1, "John"}, 1}, {[]any{2, "Joe"}, 1}},
				},
			},
		},
		{
			Name: "ROW() nesting",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ROW(ROW(1, 'x'), true);",
					Expected: []sql.Row{{[]any{[]any{1, "x"}, true}}},
				},
			},
		},
		{
			Name: "ROW() NULL handling depends on the comparison context",
			SetUpScript: []string{
				"CREATE TYPE ct AS (a INT4, b INT4);",
				"CREATE TABLE ctt (id INT4, c ct);",
				"INSERT INTO ctt VALUES (1, ROW(1, NULL)), (2, ROW(1, 2)), (3, ROW(NULL, NULL)), (4, NULL);",
				"CREATE TABLE rf (id INT4 PRIMARY KEY, a INT4, b INT4);",
				"CREATE INDEX rfi ON rf (a, b);",
				"INSERT INTO rf VALUES (1, 1, NULL), (2, 1, 2), (3, 2, 1);",
				"CREATE TABLE ck (a INT4, b INT4, CHECK ((a, b) < (5, 5)));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ROW(NULL::INT4) = ROW(NULL::INT4), ROW(NULL::INT4) = ANY(ARRAY[ROW(NULL::INT4)]);",
					Expected: []sql.Row{{nil, "t"}},
				},
				{
					Query:    "SELECT ROW(1, NULL::INT4) = ANY(ARRAY[ROW(1, NULL::INT4)]), ROW(1, NULL::INT4) <> ALL(ARRAY[ROW(1, NULL::INT4)]), ROW(1, NULL::INT4) < ANY(ARRAY[ROW(1, 2)]), ROW(1, 2) < ANY(ARRAY[ROW(1, NULL::INT4)]), ROW(1, NULL::INT4) > ANY(ARRAY[ROW(1, 2)]);",
					Expected: []sql.Row{{"t", "f", "f", "t", "t"}},
				},
				{
					Query:    "SELECT ROW(1, NULL::INT4) IS DISTINCT FROM ROW(1, NULL::INT4), ROW(1, NULL::INT4) = ALL(ARRAY[ROW(1, NULL::INT4)]), ROW(1, NULL::INT4) <= ANY(ARRAY[ROW(1, NULL::INT4)]), ROW(NULL::INT4, 1) < ANY(ARRAY[ROW(1, 1)]);",
					Expected: []sql.Row{{"f", "t", "t", "f"}},
				},
				{
					Query:    "SELECT ROW(ROW(NULL::INT4)) = ROW(ROW(NULL::INT4)), ARRAY[ROW(NULL::INT4)] = ARRAY[ROW(NULL::INT4)];",
					Expected: []sql.Row{{"t", "t"}},
				},
				{
					Query:    "SELECT record_eq(ROW(NULL::INT4), ROW(NULL::INT4)), record_lt(ROW(NULL::INT4), ROW(1)), record_gt(ROW(NULL::INT4), ROW(1)), record_ne(ROW(1, NULL::INT4), ROW(1, NULL::INT4));",
					Expected: []sql.Row{{"t", "f", "t", "f"}},
				},
				{
					Query:    "SELECT ROW(1, NULL)::ct = ROW(1, NULL)::ct, (ROW(1, NULL::INT4)) = (ROW(1, NULL::INT4)), (1, NULL::INT4) = (1, NULL::INT4), ROW(1, NULL)::ct = ROW(1, NULL::INT4);",
					Expected: []sql.Row{{"t", nil, nil, "t"}},
				},
				{
					Query:    "SELECT ROW(1, NULL::INT4) < ROW(1, NULL::INT4), ROW(NULL::INT4, 1) <= ROW(NULL::INT4, 1), ROW(1, 2, NULL::INT4) >= ROW(1, 1, NULL::INT4), ROW(1, 2, 3) > ROW(1, 2, NULL::INT4);",
					Expected: []sql.Row{{nil, nil, "t", nil}},
				},
				{
					Query:    "SELECT ROW(1, NULL::INT4) IN (ROW(1, NULL::INT4), ROW(2, 3)), ROW(1, NULL::INT4) NOT IN (ROW(1, NULL::INT4), ROW(2, 3)), ROW(1, NULL::INT4) IN (ROW(2, NULL::INT4), ROW(2, 3));",
					Expected: []sql.Row{{nil, nil, "f"}},
				},
				{
					Query:    "SELECT ROW(1, 2) IN (ROW(1, 2)), ROW(1, 2) NOT IN (ROW(1, 3), ROW(2, 2));",
					Expected: []sql.Row{{"t", "t"}},
				},
				{
					Query:       "SELECT ROW(1, 2) IN (ROW(1, 2), ROW(1));",
					ExpectedErr: "unequal number of entries",
				},
				{
					Query: "SELECT id, c = c, c = ROW(1, NULL)::ct, c < ROW(1, 3)::ct, c > ROW(1, 3)::ct, c <> ROW(1, NULL)::ct, c >= ROW(NULL, NULL)::ct FROM ctt ORDER BY id;",
					Expected: []sql.Row{
						{1, "t", "t", "f", "t", "f", "f"},
						{2, "t", "f", "t", "f", "t", "f"},
						{3, "t", "f", "f", "t", "t", "t"},
						{4, nil, nil, nil, nil, nil, nil},
					},
				},
				{
					Query: "SELECT id, ROW(1, NULL::INT4) = c, c = ROW(1, NULL::INT4), c IN (ROW(1, NULL::INT4)), ROW(1, NULL::INT4) IN (c) FROM ctt ORDER BY id;",
					Expected: []sql.Row{
						{1, "t", "t", "t", "t"},
						{2, "f", "f", "f", "f"},
						{3, "f", "f", "f", "f"},
						{4, nil, nil, nil, nil},
					},
				},
				{
					Query:    "SELECT id FROM ctt WHERE c = ANY(ARRAY[ROW(1, NULL)::ct, ROW(NULL, NULL)::ct]) ORDER BY id;",
					Expected: []sql.Row{{1}, {3}},
				},
				{
					Query:    "SELECT id FROM ctt WHERE c IN (ROW(1, NULL)::ct, ROW(NULL, NULL)::ct) ORDER BY id;",
					Expected: []sql.Row{{1}, {3}},
				},
				{
					Query:    "SELECT id FROM ctt WHERE c IN (SELECT c FROM ctt WHERE id = 1) ORDER BY id;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT id FROM rf WHERE ROW(ROW(a, b)) = ROW(ROW(1, NULL::INT4)) ORDER BY id;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT id FROM rf WHERE ROW(a, b) = ROW(1, 2) ORDER BY id;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT id FROM rf WHERE (a, b) < (1, 3) ORDER BY id;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT id FROM rf WHERE (a, b) >= (1, 2) ORDER BY id;",
					Expected: []sql.Row{{2}, {3}},
				},
				{
					Query:    "SELECT id FROM rf WHERE (a, b) > (1, 1) ORDER BY id;",
					Expected: []sql.Row{{2}, {3}},
				},
				{
					Query:    "SELECT id FROM rf WHERE (a, b) <= (2, 0) ORDER BY id;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    "SELECT id FROM rf WHERE (a, b) <> (1, 2) ORDER BY id;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT id FROM rf WHERE (a, b) IN ((1, 2), (2, 1)) ORDER BY id;",
					Expected: []sql.Row{{2}, {3}},
				},
				{
					Query:    "SELECT id FROM rf WHERE (a, b) NOT IN ((1, 2), (5, 5)) ORDER BY id;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:       "SELECT ROW() = ROW();",
					ExpectedErr: "cannot compare rows of zero length",
				},
				{
					Query:    "INSERT INTO ck VALUES (1, 9);",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO ck VALUES (5, 6);",
					ExpectedErr: "Check constraint",
				},
				{
					Query:    "SELECT * FROM ck;",
					Expected: []sql.Row{{1, 9}},
				},
			},
		},
		{
			Name: "ROW() compared to subqueries",
			SetUpScript: []string{
				"CREATE TABLE sq (x INT4, y INT4);",
				"INSERT INTO sq VALUES (1, 2), (1, NULL), (3, 4);",
				"CREATE TABLE rf (id INT4 PRIMARY KEY, a INT4, b INT4);",
				"CREATE INDEX rfi ON rf (a, b);",
				"INSERT INTO rf VALUES (1, 1, NULL), (2, 1, 2), (3, 2, 1);",
				"CREATE TABLE ck (a INT4, b INT4, CHECK (ROW(a, b) IS DISTINCT FROM ROW(1, 1)));",
				"CREATE SEQUENCE rseq;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ROW(1, NULL::INT4) IN (SELECT 1, NULL::INT4), ROW(1, 2) IN (SELECT 1, 2), ROW(1, 2) IN (SELECT 1, 3), ROW(1, 2) NOT IN (SELECT 1, 3);",
					Expected: []sql.Row{{nil, "t", "f", "t"}},
				},
				{
					Query:    "SELECT ROW(1, NULL::INT4) = ANY(SELECT 1, NULL::INT4), ROW(1, 2) = ANY(SELECT 1, 2), ROW(1, 2) < ANY(SELECT 1, 3), ROW(1, 2) < ALL(SELECT 1, 1), ROW(1, 2) <> ALL(SELECT 1, NULL::INT4);",
					Expected: []sql.Row{{nil, "t", "t", "f", nil}},
				},
				{
					Query:    "SELECT ROW(1, NULL::INT4) = (SELECT 1, NULL::INT4), ROW(1, 2) = (SELECT 1, 2), ROW(1, 2) < (SELECT 1, 3), ROW(1, 2) <> (SELECT 1, 2), ROW(1, 2) >= (SELECT 1, NULL::INT4);",
					Expected: []sql.Row{{nil, "t", "t", "f", nil}},
				},
				{
					Query:    "SELECT ROW(1, 2) IN (SELECT x, y FROM sq), ROW(1, 5) IN (SELECT x, y FROM sq), ROW(9, 9) IN (SELECT x, y FROM sq), ROW(1, 2) = ANY(SELECT x, y FROM sq), ROW(1, 5) < ALL(SELECT x, y FROM sq), ROW(0, 0) < ALL(SELECT x, y FROM sq);",
					Expected: []sql.Row{{"t", nil, "f", "t", "f", "t"}},
				},
				{
					Query:    "SELECT ROW(1, 2) = (SELECT x, y FROM sq WHERE x = 3), ROW(1, 2) = (SELECT x, y FROM sq WHERE x = 99), ROW(1, 2) IN (SELECT x, y FROM sq WHERE x = 99), ROW(1, 2) = ALL(SELECT x, y FROM sq WHERE x = 99);",
					Expected: []sql.Row{{"f", nil, "f", "t"}},
				},
				{
					Query:    "SELECT x, y, ROW(x, y) IN (SELECT 1, 2), ROW(x, y) < (SELECT 2, 0) FROM sq ORDER BY x, y;",
					Expected: []sql.Row{{1, 2, "t", "t"}, {1, nil, nil, "t"}, {3, 4, "f", "f"}},
				},
				{
					Query:    "SELECT x FROM sq WHERE ROW(x, y) IN (SELECT 1, y FROM sq WHERE y IS NOT NULL) ORDER BY x;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT ROW(1) IN (SELECT 1), ROW(NULL::INT4) = ANY(SELECT NULL::INT4);",
					Expected: []sql.Row{{"t", nil}},
				},
				{
					Query:       "SELECT ROW(1, 2) = (SELECT x, y FROM sq);",
					ExpectedErr: "more than",
				},
				{
					Query:       "SELECT ROW(1, 2) IN (SELECT x FROM sq);",
					ExpectedErr: "subquery has too few columns",
				},
				{
					Query:       "SELECT ROW(1, 2) = ANY(SELECT 1, 2, 3);",
					ExpectedErr: "subquery has too many columns",
				},
				{
					Query:       "SELECT ROW(1, 2) = (SELECT 1);",
					ExpectedErr: "subquery has too few columns",
				},
				{
					Query:       "SELECT ROW(NULL::INT4) = (SELECT ROW(NULL::INT4));",
					ExpectedErr: "operator does not exist: integer = record",
				},
				{
					Query:    "SELECT id FROM rf WHERE ROW(ROW(a, b)) < ROW(ROW(1, 3)) ORDER BY id;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT id, (a, b) < (1, 3), (a, b) >= (1, 2) FROM rf ORDER BY id;",
					Expected: []sql.Row{{1, nil, nil}, {2, "t", "t"}, {3, "f", "t"}},
				},
				{
					Query:    "SELECT ROW(nextval('rseq'), 1) < ROW(100, 2);",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT nextval('rseq');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "INSERT INTO ck VALUES (1, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO ck VALUES (1, 1);",
					ExpectedErr: "Check constraint",
				},
				{
					Query:    "SELECT * FROM ck;",
					Expected: []sql.Row{{1, 2}},
				},
			},
		},
	})
}
