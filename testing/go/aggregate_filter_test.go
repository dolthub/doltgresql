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

func TestAggregateFilter(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "aggregate FILTER (WHERE ...)",
			SetUpScript: []string{
				"CREATE TABLE posts (id INT PRIMARY KEY, grp TEXT, status TEXT, n INT);",
				`INSERT INTO posts VALUES
					(1, 'a', 'published', 1),
					(2, 'a', 'draft', 2),
					(3, 'a', 'published', NULL),
					(4, 'b', 'published', 4),
					(5, 'b', 'draft', 5);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'published') FROM posts;",
					Expected: []sql.Row{{5, 3}},
				},
				{
					Query:    "SELECT COUNT(n) FILTER (WHERE status = 'published') FROM posts;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT COUNT(DISTINCT grp) FILTER (WHERE status = 'draft') FROM posts;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT MIN(n) FILTER (WHERE status = 'draft'), MAX(n) FILTER (WHERE status = 'published') FROM posts;",
					Expected: []sql.Row{{2, 4}},
				},
				{
					Query: "SELECT grp, COUNT(*) FILTER (WHERE status = 'published'), COUNT(*) FROM posts GROUP BY grp ORDER BY grp;",
					Expected: []sql.Row{
						{"a", 2, 3},
						{"b", 1, 2},
					},
				},
				{
					Query:    "SELECT COUNT(*) FILTER (WHERE status = 'archived'), MAX(n) FILTER (WHERE status = 'archived') FROM posts;",
					Expected: []sql.Row{{0, nil}},
				},
				{
					Query:    "SELECT string_agg(status, ',') FILTER (WHERE id = 4) FROM posts;",
					Expected: []sql.Row{{"published"}},
				},
				{
					Query:       "SELECT array_agg(n) FILTER (WHERE status = 'draft') FROM posts;",
					ExpectedErr: "FILTER is not yet supported for array_agg",
				},
			},
		},
	})
}
