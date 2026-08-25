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

func TestCountDistinctUuid(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "COUNT DISTINCT uuid",
			SetUpScript: []string{
				"CREATE TABLE uuid_distinct (g text, u uuid, uuid_text text, n integer);",
				`INSERT INTO uuid_distinct VALUES
					('a','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001',1),
					('a','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001',1),
					('a','00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002',2),
					('a',NULL,NULL,NULL),
					('b','00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002',2),
					('b','00000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000003',3),
					('b',NULL,NULL,NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(DISTINCT u) FROM uuid_distinct;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT COUNT(DISTINCT u) FROM uuid_distinct WHERE u IS NULL;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT COUNT(DISTINCT u) FROM uuid_distinct WHERE false;",
					Expected: []sql.Row{{0}},
				},
				{
					Query: "SELECT g, COUNT(DISTINCT u) FROM uuid_distinct GROUP BY g ORDER BY g;",
					Expected: []sql.Row{
						{"a", 2},
						{"b", 2},
					},
				},
				{
					Query: "SELECT COUNT(DISTINCT (u::text)::uuid), COUNT(DISTINCT uuid_text::uuid) FROM uuid_distinct;",
					Expected: []sql.Row{
						{3, 3},
					},
				},
				{
					Query:    "SELECT COUNT(DISTINCT ('{' || uuid_text || '}')::uuid) FROM uuid_distinct;",
					Expected: []sql.Row{{3}},
				},
				{
					Query: "SELECT COUNT(DISTINCT u), COUNT(DISTINCT uuid_text), COUNT(DISTINCT n) FROM uuid_distinct;",
					Expected: []sql.Row{
						{3, 3, 3},
					},
				},
				{
					Query:    "SELECT pg_typeof(COUNT(DISTINCT u)) FROM uuid_distinct;",
					Expected: []sql.Row{{"bigint"}},
				},
			},
		},
	})
}
