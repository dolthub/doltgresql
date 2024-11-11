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
)

func TestExplain(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "basic explain tests",
			SetUpScript: []string{
				`CREATE TABLE t (i INT PRIMARY KEY)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Skip:  true, // Our explain output is very different
					Query: `EXPLAIN SELECT * FROM T;`,
					Expected: []sql.Row{
						{"Seq Scan on t  (cost=0.00..35.50 rows=2550 width=4)"},
					},
				},
				{
					Skip: true, // Need to properly support explain options
					Query: `
EXPLAIN 
(
	ANALYZE, 
	VERBOSE, 
	COSTS, 
	SETTINGS,
	BUFFERS,
	WAL,
	TIMING,
	SUMMARY,
	FORMAT TEXT
) 
	SELECT * FROM t;
`,
					Expected: []sql.Row{
						{"Seq Scan on t  (cost=0.00..35.50 rows=2550 width=4)"},
					},
				},
				{
					Skip: true, // Need to properly support explain options
					Query: `
EXPLAIN 
(
	NOTAVALIDOPTION
) 
	SELECT * FROM t;
`,
					ExpectedErr: "ERROR:  unrecognized EXPLAIN option \"NOTAVALIDOPTION\"",
				},
			},
		},
	})
}
