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

// TestLateralIndexPushdown asserts that the analyzer pushes an indexed lookup into a LATERAL subquery's
// inner filter when it references an outer-scope column and an index exists for the comparison.
func TestLateralIndexPushdown(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "LATERAL subquery pushes an outer-scope equality into an indexed lookup",
			SetUpScript: []string{
				`CREATE TABLE a (id INT PRIMARY KEY, x INT);`,
				`CREATE TABLE b (id INT PRIMARY KEY, x INT);`,
				`CREATE INDEX b_x_idx ON b(x);`,
				`INSERT INTO a VALUES (1,1),(2,2),(3,3);`,
				`INSERT INTO b VALUES (1,1),(2,2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `EXPLAIN SELECT * FROM a, LATERAL (SELECT b.x FROM b WHERE a.x = b.x) sq ORDER BY a.id;`,
					Expected: []sql.Row{
						{"Sort(a.id ASC)"},
						{" └─ LateralCrossJoin"},
						{"     ├─ Table"},
						{"     │   └─ name: a"},
						{"     └─ SubqueryAlias"},
						{"         ├─ name: sq"},
						{"         ├─ outerVisibility: false"},
						{"         ├─ isLateral: true"},
						{"         ├─ cacheable: false"},
						{"         ├─ colSet: (5)"},
						{"         ├─ tableId: 3"},
						{"         └─ Filter"},
						{"             ├─ a.x = b.x"},
						{"             └─ IndexedTableAccess(b)"},
						{"                 ├─ index: [b.x]"},
						{"                 ├─ columns: [x]"},
						{"                 └─ keys: a.x"},
					},
				},
				{
					Query: `SELECT a.id, a.x, sq.x FROM a, LATERAL (SELECT b.x FROM b WHERE a.x = b.x) sq ORDER BY a.id;`,
					Expected: []sql.Row{
						{1, 1, 1},
						{2, 2, 2},
					},
				},
			},
		},
	})
}
