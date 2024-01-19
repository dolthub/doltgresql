// Copyright 2023 Dolthub, Inc.
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var preparedStatementTests = []ScriptTest{
	{
		Name: "expressions without tables",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT CONCAT($1::text, $2::text)",
				BindVars: []any{"hello", "world"},
				Expected: []sql.Row{
					{"helloworld"},
				},
			},
			{
				Query:    "SELECT $1::integer + $2::integer",
				BindVars: []any{1, 2},
				Expected: []sql.Row{
					{3},
				},
			},
		},
	},
	{
		Name: "Integer insert",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 2, 3, 4},
			},
			{
				Query: "SELECT * FROM test order by pk;",
				Expected: []sql.Row{
					{1, 2},
					{3, 4},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{2},
				Expected: []sql.Row{
					{1, 2},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{3},
				Expected: []sql.Row{},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 + $1 = $2;",
				BindVars: []any{1, 3},
				Expected: []sql.Row{
					{1, 2},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE pk + v1 = $1;",
				BindVars: []any{3},
				Expected: []sql.Row{
					{1, 2},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1::integer + $2::integer;",
				BindVars: []any{1, 3},
				Expected: []sql.Row{
					{3, 4},
				},
			},
		},
	},
	{
		Name: "Integer update",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 2, 3, 4},
			},
			{
				Query:    "UPDATE test set v1 = $1 WHERE pk = $2;",
				BindVars: []any{5, 1},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{5},
				Expected: []sql.Row{
					{1, 5},
				},
			},
		},
	},
	{
		Name: "Integer delete",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 2, 3, 4},
			},
			{
				Query:    "DELETE FROM test WHERE pk = $1;",
				BindVars: []any{1},
			},
			{
				Query: "SELECT * FROM test order by 1;",
				Expected: []sql.Row{
					{3, 4},
				},
			},
		},
	},
	{
		Name: "String insert",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, "hello", 3, "goodbye"},
			},
			{
				Query: "SELECT * FROM test order by pk;",
				Expected: []sql.Row{
					{1, "hello"},
					{3, "goodbye"},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE s = $1;",
				BindVars: []any{"hello"},
				Expected: []sql.Row{
					{1, "hello"},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE s = concat($1::text, $2::text);",
				BindVars: []any{"he", "llo"},
				Expected: []sql.Row{
					{1, "hello"},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE concat(s, '!') = $1",
				BindVars: []any{"hello!"},
				Expected: []sql.Row{
					{1, "hello"},
				},
			},
		},
	},
	{
		Name: "String update",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, "hello", 3, "goodbye"},
			},
			{
				Query:    "UPDATE test set s = $1 WHERE pk = $2;",
				BindVars: []any{"new value", 1},
			},
			{
				Query:    "SELECT * FROM test WHERE s = $1;",
				BindVars: []any{"new value"},
				Expected: []sql.Row{
					{1, "new value"},
				},
			},
		},
	},
	{
		Name: "String delete",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, "hello", 3, "goodbye"},
			},
			{
				Query:    "DELETE FROM test WHERE s = $1;",
				BindVars: []any{"hello"},
			},
			{
				Query: "SELECT * FROM test ORDER BY 1;",
				Expected: []sql.Row{
					{3, "goodbye"},
				},
			},
		},
	},
	{
		Name: "Float insert",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 1.1, 3, 3.3},
			},
			{
				Query: "SELECT * FROM test ORDER BY 1;",
				Expected: []sql.Row{
					{1, 1.1},
					{3, 3.3},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE f1 = $1;",
				BindVars: []any{1.1},
				Expected: []sql.Row{
					{1, 1.1},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE f1 + $1 = $2;",
				BindVars: []any{1.0, 2.1},
				Expected: []sql.Row{
					{1, 1.1},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE f1 = $1::decimal + $2::decimal;",
				BindVars: []any{1.0, 0.1},
				Expected: []sql.Row{
					{1, 1.1},
				},
			},
		},
	},
	{
		Name: "Float update",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 1.1, 3, 3.3},
			},
			{
				Query:    "UPDATE test set f1 = $1 WHERE f1 = $2;",
				BindVars: []any{2.2, 1.1},
			},
			{
				Query:    "SELECT * FROM test WHERE f1 = $1;",
				BindVars: []any{2.2},
				Expected: []sql.Row{
					{1, 2.2},
				},
			},
		},
	},
	{
		Name: "Float delete",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 1.1, 3, 3.3},
			},
			{
				Query:    "DELETE FROM test WHERE f1 = $1;",
				BindVars: []any{1.1},
			},
			{
				Query: "SELECT * FROM test order by 1;",
				Expected: []sql.Row{
					{3, 3.3},
				},
			},
		},
	},
}

func TestPreparedErrorHandling(t *testing.T) {
	tt := ScriptTest{
		Name: "error handling doesn't foul session",
		SetUpScript: []string{
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"insert into test values (1, 1), (2, 2), (3, 3), (4, 4), (5, 5), (6, 6), (7, 7);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select v1 from doesNotExist where pk = 1;",
				ExpectedErr: true,
			},
			{
				Query:    "select v1 from test where pk = 1;",
				Expected: []sql.Row{{1}},
			},
			{
				Query:    "select v1 from test where pk = 2;",
				Expected: []sql.Row{{2}},
			},
			{
				Query:    "select v1 from test where pk = 3;",
				Expected: []sql.Row{{3}},
			},
			{
				Query:    "select v1 from test where pk = 4;",
				Expected: []sql.Row{{4}},
			},
			{
				Query:    "select v1 from test where pk = 5;",
				Expected: []sql.Row{{5}},
			},
			{
				Query:    "select v1 from test where pk = 6;",
				Expected: []sql.Row{{6}},
			},
			{
				Query:    "select v1 from test where pk = 7;",
				Expected: []sql.Row{{7}},
			},
		},
	}

	RunScriptN(t, tt, 20)
}

func TestPreparedStatements(t *testing.T) {
	RunScripts(t, preparedStatementTests)
}

// RunScriptN runs the assertios of the given script n times using the same connection
func RunScriptN(t *testing.T, script ScriptTest, n int) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}
	ctx, conn, controller := CreateServer(t, scriptDatabase)
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		err := controller.WaitForStop()
		require.NoError(t, err)
	}()

	// Run the setup
	for _, query := range script.SetUpScript {
		rows, err := conn.Query(ctx, query)
		require.NoError(t, err)
		_, err = ReadRows(rows)
		assert.NoError(t, err)
	}

	for i := 0; i < n; i++ {
		t.Run(script.Name, func(t *testing.T) {
			// Run the assertions
			for _, assertion := range script.Assertions {
				t.Run(assertion.Query, func(t *testing.T) {
					if assertion.Skip {
						t.Skip("Skip has been set in the assertion")
					}
					rows, err := conn.Query(ctx, assertion.Query)
					if assertion.ExpectedErr {
						rows.Close()
						require.Error(t, err)
						return
					} else {
						require.NoError(t, err)
					}

					foundRows, err := ReadRows(rows)
					if assertion.ExpectedErr {
						require.Error(t, err)
						return
					} else {
						require.NoError(t, err)
					}
					assert.Equal(t, NormalizeRows(assertion.Expected), foundRows)
				})
			}
		})
	}
}
