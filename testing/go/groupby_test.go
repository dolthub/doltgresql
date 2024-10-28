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

func TestGroupBy(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Basic order by/group by cases",
			SetUpScript: []string{
				"create table members (id bigint primary key, team text);",
				"insert into members values (3,'red'), (4,'red'),(5,'orange'),(6,'orange'),(7,'orange'),(8,'purple');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select team as f from members order by id, f",
					Expected: []sql.Row{{"red"}, {"red"}, {"orange"}, {"orange"}, {"orange"}, {"purple"}},
				},
				{
					Query: "SELECT team, COUNT(*) FROM members GROUP BY team ORDER BY 2",
					Expected: []sql.Row{
						{"purple", int64(1)},
						{"red", int64(2)},
						{"orange", int64(3)},
					},
				},
				{
					Query: "SELECT team, COUNT(*) FROM members GROUP BY 1 ORDER BY 2",
					Expected: []sql.Row{
						{"purple", int64(1)},
						{"red", int64(2)},
						{"orange", int64(3)},
					},
				},
				{
					Query:       "SELECT team, COUNT(*) FROM members GROUP BY team ORDER BY columndoesnotexist",
					ExpectedErr: "not be found",
				},
				{
					Query:    "SELECT DISTINCT t1.id as id FROM members AS t1 JOIN members AS t2 ON t1.id = t2.id WHERE t2.id > 0 ORDER BY t1.id",
					Expected: []sql.Row{{3}, {4}, {5}, {6}, {7}, {8}},
				},
				{
					Query:    "SELECT id as alias1, (SELECT alias1+1 group by alias1 having alias1 > 0) FROM members where id < 6;",
					Expected: []sql.Row{{3, 4}, {4, 5}, {5, 6}},
				},
				{
					Query:    "SELECT id, (SELECT UPPER(team) having id > 3) as upper_team FROM members where id < 6;",
					Expected: []sql.Row{{3, nil}, {4, "RED"}, {5, "ORANGE"}},
				},
				{
					Query:    "SELECT id, (SELECT -1 as id having id < 10) as upper_team FROM members where id < 6;",
					Expected: []sql.Row{{3, -1}, {4, -1}, {5, -1}},
				},
			},
		},
	})
}
