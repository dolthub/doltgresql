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

// TestMergeJoinNullKey verifies that merge joins skip NULL join keys rather than panicking.
// Regression test for https://github.com/dolthub/doltgresql/issues/3879.
func TestMergeJoinNullKey(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "NOT EXISTS over a nullable foreign key",
			SetUpScript: []string{
				"CREATE TABLE users (id int PRIMARY KEY, name text NOT NULL);",
				"CREATE TABLE tasks (id int PRIMARY KEY, title text NOT NULL, assignee_id int REFERENCES users (id));",
				"INSERT INTO users VALUES (1, 'ana'), (2, 'bo'), (3, 'cy');",
				"INSERT INTO tasks VALUES (10, 'write spec', 1), (11, 'review', 2), (12, 'triage', null), (13, 'deploy', 1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT u.id FROM users u WHERE NOT EXISTS (SELECT 1 FROM tasks t WHERE t.assignee_id = u.id) ORDER BY u.id;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT u.id FROM users u WHERE EXISTS (SELECT 1 FROM tasks t WHERE t.assignee_id = u.id) ORDER BY u.id;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    "SELECT t.id, u.id FROM tasks t LEFT JOIN users u ON t.assignee_id = u.id ORDER BY t.id;",
					Expected: []sql.Row{{10, 1}, {11, 2}, {12, nil}, {13, 1}},
				},
			},
		},
		{
			Name: "EXISTS over two indexed nullable columns",
			SetUpScript: []string{
				"CREATE TABLE stores (id int PRIMARY KEY, region_id int);",
				"CREATE TABLE orders (id int PRIMARY KEY, region_id int);",
				"CREATE INDEX stores_region_id ON stores (region_id);",
				"CREATE INDEX orders_region_id ON orders (region_id);",
				"INSERT INTO stores SELECT g, g % 30 FROM generate_series(1, 10) g;",
				"INSERT INTO orders SELECT g, CASE WHEN g = 1 THEN NULL ELSE g % 20 END FROM generate_series(1, 10) g;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT o.id FROM orders o WHERE EXISTS (SELECT 1 FROM stores s WHERE s.region_id = o.region_id) ORDER BY o.id;",
					Expected: []sql.Row{{2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}},
				},
				{
					Query:    "SELECT o.id FROM orders o WHERE NOT EXISTS (SELECT 1 FROM stores s WHERE s.region_id = o.region_id) ORDER BY o.id;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "INSERT INTO stores VALUES (11, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT o.id FROM orders o WHERE EXISTS (SELECT 1 FROM stores s WHERE s.region_id = o.region_id) ORDER BY o.id;",
					Expected: []sql.Row{{2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}},
				},
				{
					Query:    "SELECT o.id FROM orders o WHERE NOT EXISTS (SELECT 1 FROM stores s WHERE s.region_id = o.region_id) ORDER BY o.id;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT count(*) FROM orders o JOIN stores s ON s.region_id = o.region_id;",
					Expected: []sql.Row{{9}},
				},
			},
		},
		{
			Name: "NULL and duplicate keys on both sides",
			SetUpScript: []string{
				"CREATE TABLE a (id int PRIMARY KEY, k int);",
				"CREATE TABLE b (id int PRIMARY KEY, k int);",
				"CREATE INDEX a_k ON a (k);",
				"CREATE INDEX b_k ON b (k);",
				"INSERT INTO a VALUES (1, NULL), (2, NULL), (3, 1), (4, 2), (5, 2), (6, 5);",
				"INSERT INTO b VALUES (1, NULL), (2, 2), (3, 2), (4, 3), (5, 5), (6, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT a.id FROM a WHERE EXISTS (SELECT 1 FROM b WHERE b.k = a.k) ORDER BY a.id;",
					Expected: []sql.Row{{4}, {5}, {6}},
				},
				{
					Query:    "SELECT a.id FROM a WHERE NOT EXISTS (SELECT 1 FROM b WHERE b.k = a.k) ORDER BY a.id;",
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    "SELECT b.id FROM b WHERE NOT EXISTS (SELECT 1 FROM a WHERE a.k = b.k) ORDER BY b.id;",
					Expected: []sql.Row{{1}, {4}, {6}},
				},
				{
					Query:    "SELECT count(*) FROM a JOIN b ON a.k = b.k;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "SELECT count(*), count(b.id) FROM a LEFT JOIN b ON a.k = b.k;",
					Expected: []sql.Row{{8, 5}},
				},
			},
		},
	})
}
