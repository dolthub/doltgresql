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

func TestDelete(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple delete",
			SetUpScript: []string{
				"CREATE TABLE t123 (id int primary key, c1 varchar(100));",
				"INSERT INTO t123 VALUES (1, 'one'), (2, 'two');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DELETE FROM t123 where id = 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM t123;",
					Expected: []sql.Row{{2, "two"}},
				},
			},
		},
		{
			Name: "delete returning",
			SetUpScript: []string{
				"CREATE TABLE t123 (id int primary key, c1 varchar(100));",
				"INSERT INTO t123 VALUES (1, 'one'), (2, 'two'), (3, 'three');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DELETE FROM t123 where id = 1 RETURNING id, c1;",
					Expected: []sql.Row{{1, "one"}},
				},
				{
					// Test a DELETE with no filter, to test that we don't convert
					// to a TRUNCATE operation
					Query:    "DELETE FROM t123 RETURNING *;",
					Expected: []sql.Row{{2, "two"}, {3, "three"}},
				},
				{
					Query:    "SELECT * FROM t123;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
