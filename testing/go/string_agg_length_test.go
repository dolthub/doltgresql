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

func TestStringAggLength(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "string_agg is not cut at 1024 characters",
			SetUpScript: []string{
				"CREATE TABLE parts (id INT PRIMARY KEY, s TEXT);",
				"INSERT INTO parts SELECT g, repeat('x', 100) FROM generate_series(1, 50) AS g;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT length(string_agg(s, '')) FROM parts;",
					Expected: []sql.Row{{5000}},
				},
				{
					Query:    "SELECT length(string_agg(s, ',' ORDER BY id)) FROM parts;",
					Expected: []sql.Row{{5049}},
				},
			},
		},
	})
}
