// Copyright 2026 Dolthub, Inc.
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

// TestDoltProcedureRecordResults tests that the functions wrapping Dolt stored procedures return their results with
// the same semantics as Postgres functions with OUT parameters: invoking them in a FROM clause explodes the result
// into one column per OUT parameter, while invoking them in a SELECT list returns a single record value (or a bare
// value for procedures with a single result column).
func TestDoltProcedureRecordResults(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "dolt procedures in FROM clause explode into columns",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM dolt_checkout('-b', 'newbranch');",
					Expected: []sql.Row{{int64(0), "Switched to branch 'newbranch'"}},
				},
				{
					Query:    "SELECT message FROM dolt_checkout('main');",
					Expected: []sql.Row{{"Switched to branch 'main'"}},
				},
				{
					Query:    "SELECT * FROM dolt_checkout('newbranch') AS t(a, b);",
					Expected: []sql.Row{{int64(0), "Switched to branch 'newbranch'"}},
				},
				{
					Query:    "SELECT * FROM dolt_add('.');",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "CREATE TABLE t1 (pk int primary key);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT count(*) FROM dolt_commit('-Am', 'new table') WHERE length(hash) = 32;",
					Expected: []sql.Row{{int64(1)}},
				},
			},
		},
		{
			Name: "dolt procedures in SELECT list return records or bare values",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT dolt_checkout('-b', 'newbranch');",
					Expected: []sql.Row{{[]any{int64(0), "Switched to branch 'newbranch'"}}},
				},
				{
					Query:    "SELECT dolt_add('.');",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT (dolt_checkout('main')).message;",
					Expected: []sql.Row{{"Switched to branch 'main'"}},
					Skip:     true, // TODO: field selection on record-returning function calls is not yet supported
				},
			},
		},
	})
}
