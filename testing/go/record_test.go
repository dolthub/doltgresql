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
	"github.com/dolthub/go-mysql-server/sql/types"
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
					Expected: []sql.Row{{"(1,1)"}},
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
					Expected: []sql.Row{{"(1,2,3)"}},
				},
				{
					Query:    "SELECT (4, 5, 6) as myRow;",
					Expected: []sql.Row{{"(4,5,6)"}},
				},
				{
					Query:    "SELECT (NULL, 'foo', NULL) as myRow;",
					Expected: []sql.Row{{"(,foo,)"}},
				},
				{
					Query:    "SELECT (NULL, (1 > 0), 'baz') as myRow;",
					Expected: []sql.Row{{"(,t,baz)"}},
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
			},
		},
		{
			// TODO: Additional work is needed to support inserting records into tables
			Skip: true,
			Name: "ROW() use inserting and selecting composite rows",
			SetUpScript: []string{
				"CREATE TYPE user_info AS (id INT, name TEXT, email TEXT);",
				"CREATE TABLE accounts (info user_info);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: ERROR: ASSIGNMENT_CAST: target is of type user_info but expression is of type record
					Query:    "INSERT INTO accounts VALUES (ROW(1, 'alice', 'a@example.com'));",
					Expected: []sql.Row{{types.NewOkResult(1)}},
				},
				{
					Query:    "SELECT info FROM accounts;",
					Expected: []sql.Row{{"(1,alice,a@example.com)"}},
				},
				{
					// TODO: ERROR: (E).x is not yet supported (SQLSTATE XX000)
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
					// TODO: ERROR: unknown type with oid: 2249
					Skip:     true,
					Query:    "SELECT ROW(1, 'a')::record;",
					Expected: []sql.Row{{"(1,a)"}},
				},
				{
					// TODO: This does not return an error yet
					Skip:        true,
					Query:       "SELECT ROW(1, 2) = ROW(1, 'two');",
					ExpectedErr: "invalid input syntax",
				},
				{
					// TODO: interface conversion panic
					Skip:     true,
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
					// TODO: expression.IsNull in GMS is used in this evaluation, but returns
					//       false for this case, because the record evaluates to []any{nil}
					//       instead of just nil.
					Skip:     true,
					Query:    "SELECT ROW(NULL) IS NULL",
					Expected: []sql.Row{{"t"}},
				},
				{
					// TODO: expression.IsNull in GMS is used in this evaluation, but returns
					//       false for this case, because the record evaluates to []any{nil}
					//       instead of just nil.
					Skip:     true,
					Query:    "SELECT ROW(NULL, NULL, NULL) IS NULL;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(NULL, 42, NULL) IS NULL;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT ROW(42) IS NULL",
					Expected: []sql.Row{{0}},
				},
				{
					// TODO: expression.IsNull in GMS is used in this evaluation (wrapped with
					//       an expression.Not), but returns true for this case, because the record
					//       evaluates to []any{nil} instead of just nil.
					Skip:     true,
					Query:    "SELECT ROW(NULL) IS NOT NULL;",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT ROW(42) IS NOT NULL;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT ROW(id, name), COUNT(*) FROM users GROUP BY ROW(id, name);",
					Expected: []sql.Row{{"(1,John)", 1}, {"(2,Joe)", 1}},
				},
			},
		},
		{
			Name: "ROW() nesting",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ROW(ROW(1, 'x'), true);",
					Expected: []sql.Row{{`("(1,x)",t)`}},
				},
			},
		},
	})
}
