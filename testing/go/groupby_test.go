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
		{
			Name: "Equal values with different representations are grouped together",
			SetUpScript: []string{
				"CREATE TABLE waits (id INT, w INTERVAL);",
				"INSERT INTO waits VALUES (1, INTERVAL '1 day'), (2, '2024-01-02 00:00'::TIMESTAMP - '2024-01-01 00:00'::TIMESTAMP), (3, INTERVAL '24 hours'), (4, INTERVAL '1 month'), (5, INTERVAL '30 days'), (6, INTERVAL '25 hours');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM waits GROUP BY w ORDER BY 1;",
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    "SELECT MIN(id), COUNT(*) FROM waits GROUP BY w ORDER BY 1;",
					Expected: []sql.Row{{1, 3}, {4, 2}, {6, 1}},
				},
				{
					Query:    "SELECT COUNT(DISTINCT w) FROM waits;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT DISTINCT w FROM waits) t;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT id FROM waits WHERE w IN (SELECT INTERVAL '1 day') ORDER BY id;",
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    "SELECT id FROM waits WHERE w IN (SELECT w FROM waits WHERE id IN (3, 5)) ORDER BY id;",
					Expected: []sql.Row{{1}, {2}, {3}, {4}, {5}},
				},
				{
					Query:    "SELECT id FROM waits WHERE w NOT IN (SELECT INTERVAL '24 hours') ORDER BY id;",
					Expected: []sql.Row{{4}, {5}, {6}},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT w FROM waits WHERE id = 1 UNION SELECT w FROM waits WHERE id IN (2, 3)) t;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT INTERVAL '1 day' UNION SELECT INTERVAL '24 hours' UNION SELECT INTERVAL '1440 minutes') t;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT a.id, b.id FROM waits a JOIN waits b ON a.w = b.w WHERE a.id < b.id ORDER BY 1, 2;",
					Expected: []sql.Row{{1, 2}, {1, 3}, {2, 3}, {4, 5}},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT DISTINCT x FROM (VALUES (ARRAY[1.0]), (ARRAY[1.00]), (ARRAY[1])) t(x)) s;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM (VALUES (ARRAY[1.0, 2.50]), (ARRAY[1.00, 2.5]), (ARRAY[1.0, 2.51])) t(x) GROUP BY x ORDER BY 1;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    "SELECT COUNT(DISTINCT x) FROM (VALUES (ARRAY[INTERVAL '1 day']), (ARRAY[INTERVAL '24 hours'])) t(x);",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT COUNT(*) FROM (VALUES (1.0), (1.00), (1)) t(x) GROUP BY x;",
					Expected: []sql.Row{{3}},
				},
			},
		},
	})
}
