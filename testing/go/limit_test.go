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

func TestLimitOffset(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "basic limit tests",
			SetUpScript: []string{
				`CREATE TABLE t (i INT PRIMARY KEY, c int)`,
				`INSERT INTO t VALUES (1, 1), (2, 2), (3, 3), (4, 4)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM t LIMIT 2`,
					Expected: []sql.Row{{1, 1}, {2, 2}},
				},
				{
					Query:    `SELECT * FROM t LIMIT 2 OFFSET 2`,
					Expected: []sql.Row{{3, 3}, {4, 4}},
				},
				{
					Query:    `SELECT * FROM t order by c asc LIMIT 2 OFFSET 2`,
					Expected: []sql.Row{{3, 3}, {4, 4}},
				},
			},
		},
	})
}
