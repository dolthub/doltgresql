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

// TestSelect covers SELECT syntax not covered by our MySQL select tests
func TestSelect(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "SELECT DISTINCT ON",
			SetUpScript: []string{
				"CREATE TABLE test (v1 INT4, v2 INT4);",
				"INSERT INTO test VALUES (1, 3), (1, 4), (2, 3), (2, 4);",
				"CREATE TABLE test2 (v1 INT4, v2 INT4, v3 INT4);",
				"INSERT INTO test2 VALUES (1, 3, 5), (2, 3, 5), (1, 4, 5), (2, 4, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY v1, v2;",
					Expected: []sql.Row{
						{1, 3},
						{1, 4},
						{2, 3},
						{2, 4},
					},
				},
				{
					Query: "SELECT DISTINCT * FROM test ORDER BY v1, v2;",
					Expected: []sql.Row{
						{1, 3},
						{1, 4},
						{2, 3},
						{2, 4},
					},
				},
				{
					Query: "SELECT DISTINCT ON(v1) * FROM test ORDER BY v1, v2;",
					Expected: []sql.Row{
						{1, 3},
						{2, 3},
					},
				},
				{
					Query: "SELECT DISTINCT ON(v2) * FROM test ORDER BY v2, v1;",
					Expected: []sql.Row{
						{1, 3},
						{1, 4},
					},
				},
				{
					Query:       "SELECT DISTINCT ON(v1) * FROM test ORDER BY v2, v1;",
					ExpectedErr: sql.ErrDistinctOnMatchOrderBy.Message,
				},
				{
					Query: "SELECT DISTINCT ON(v2) * FROM test ORDER BY v2 DESC, v1 DESC;",
					Expected: []sql.Row{
						{2, 4},
						{2, 3},
					},
				},
				{
					Query: "SELECT DISTINCT ON(v2, v1) * FROM test2 ORDER BY v1, v2;",
					Expected: []sql.Row{
						{1, 3, 5},
						{1, 4, 5},
						{2, 3, 5},
						{2, 4, 5},
					},
				},
				{
					Query: "SELECT DISTINCT ON(v2, v1) * FROM test2 ORDER BY v1, v2 DESC;",
					Expected: []sql.Row{
						{1, 4, 5},
						{1, 3, 5},
						{2, 4, 5},
						{2, 3, 5},
					},
				},
				{
					Query: "SELECT DISTINCT ON(v2, v1) * FROM test2 ORDER BY v1, v2 LIMIT 1;",
					Expected: []sql.Row{
						{1, 3, 5},
					},
				},
				{
					Query: "SELECT DISTINCT ON(v2, v1) * FROM test2 ORDER BY v1, v2 DESC LIMIT 1;",
					Expected: []sql.Row{
						{1, 4, 5},
					},
				},
				{
					Query: "SELECT DISTINCT ON(v1, v2, v3) * FROM test2 ORDER BY v1, v2;",
					Expected: []sql.Row{
						{1, 3, 5},
						{1, 4, 5},
						{2, 3, 5},
						{2, 4, 5},
					},
				},
				{
					Query: "SELECT DISTINCT ON(v3) v1 FROM test2;",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query:       "SELECT DISTINCT ON(v1, v3) * FROM test2 ORDER BY v1, v2;",
					ExpectedErr: sql.ErrDistinctOnMatchOrderBy.Message,
				},
				{
					Query:       "SELECT DISTINCT ON(v2) * FROM test2 ORDER BY v1, v2;",
					ExpectedErr: sql.ErrDistinctOnMatchOrderBy.Message,
				},
			},
		},
		{
			Name: "select values",
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22));",
					Expected: []sql.Row{
						{1, "峰哥", 18},
						{2, "王哥", 20},
						{3, "张哥", 22},
					},
				},
				{
					Query: "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22)) x(id,name,age);",
					Expected: []sql.Row{
						{1, "峰哥", 18},
						{2, "王哥", 20},
						{3, "张哥", 22},
					},
					ExpectedColNames: []string{
						"id",
						"name",
						"age",
					},
				},
				{
					Query:    "select * from (values(1,'峰哥',18),(2,'王哥',20),(3,'张哥',22)) x(id,name,age) limit $1;",
					BindVars: []any{2}, // forcing this to use prepared statements
					Expected: []sql.Row{
						{1, "峰哥", 18},
						{2, "王哥", 20},
					},
					ExpectedColNames: []string{
						"id",
						"name",
						"age",
					},
				},
			},
		},
		{
			Name: "SELECT with no select expressions",
			SetUpScript: []string{
				"CREATE TABLE mytable (pk int primary key);",
				"INSERT INTO mytable VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select from mytable;",
					Expected: []sql.Row{{}, {}},
				},
				{
					// https://github.com/dolthub/doltgresql/issues/1470
					Query:    "SELECT EXISTS (SELECT FROM mytable where pk > 0);",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
	})
}
