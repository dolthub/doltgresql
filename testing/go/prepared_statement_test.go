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

func TestPreparedStatements(t *testing.T) {
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
