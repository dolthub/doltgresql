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

func TestWindowFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "native sum and row_number as window functions",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, grp INT, amt INT);",
				"INSERT INTO t VALUES (1, 1, 10), (2, 1, 20), (3, 1, 30), (4, 2, 5), (5, 2, 15);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// SUM(int4) OVER (...) must return a bigint (int64), not GMS's float64.
					Query: "SELECT id, sum(amt) OVER (PARTITION BY grp ORDER BY id) FROM t ORDER BY id",
					Expected: []sql.Row{
						{1, int64(10)},
						{2, int64(30)},
						{3, int64(60)},
						{4, int64(5)},
						{5, int64(20)},
					},
				},
				{
					// row_number() must return a bigint directly and reset per partition.
					Query: "SELECT id, row_number() OVER (PARTITION BY grp ORDER BY id) FROM t ORDER BY id",
					Expected: []sql.Row{
						{1, int64(1)},
						{2, int64(2)},
						{3, int64(3)},
						{4, int64(1)},
						{5, int64(2)},
					},
				},
				{
					// sum(amt) as a regular GROUP BY aggregate still works and is also bigint.
					Query: "SELECT grp, sum(amt) FROM t GROUP BY grp ORDER BY grp",
					Expected: []sql.Row{
						{1, int64(60)},
						{2, int64(20)},
					},
				},
			},
		},
		{
			// https://github.com/dolthub/doltgresql/issues/1796
			Name: "basic window functions",
			SetUpScript: []string{
				"CREATE TABLE c (c_id INT PRIMARY KEY, bill TEXT);",
				"CREATE TABLE o (o_id INT PRIMARY KEY, c_id INT, ship TEXT);",
				"INSERT INTO c VALUES (1, 'CA'), (2, 'TX'), (3, 'MA'), (4, 'TX'), (5, NULL), (6, 'FL');",
				"INSERT INTO o VALUES (10, 1, 'CA'), (20, 1, 'CA'), (30, 1, 'CA'), (40, 2, 'CA'), (50, 2, 'TX'), (60, 2, NULL), (70, 4, 'WY'), (80, 4, NULL), (90, 6, 'WA');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT row_number() OVER () AS rn FROM o WHERE c_id=-999",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT row_number() OVER () AS rn FROM o WHERE c_id=1",
					Expected: []sql.Row{
						{int64(1)},
						{int64(2)},
						{int64(3)},
					},
				},
				{
					Query:    "SELECT rank() OVER () AS rnk FROM o WHERE c_id=-999",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT o_id, c_id, rank() OVER (ORDER BY o_id) AS rnk FROM o WHERE c_id=1",
					Expected: []sql.Row{
						{10, 1, int64(1)},
						{20, 1, int64(2)},
						{30, 1, int64(3)},
					},
				},
				{
					Query:    "SELECT dense_rank() OVER () AS drnk FROM o WHERE c_id=-999",
					Expected: []sql.Row{},
				},
				{
					// Doltgres inherits GMS/MySQL null-first sort order, so NULLs sort before
					// non-NULLs in window ORDER BY and outer ORDER BY. Postgres default is NULLS LAST.
					// TODO: update expected values once Doltgres adopts Postgres NULLS LAST ordering.
					Query: "SELECT ship, dense_rank() OVER (ORDER BY ship) AS drnk FROM o WHERE c_id IN (1, 2) ORDER BY ship",
					Expected: []sql.Row{
						{nil, int64(1)},
						{"CA", int64(2)},
						{"CA", int64(2)},
						{"CA", int64(2)},
						{"CA", int64(2)},
						{"TX", int64(3)},
					},
				},
				{
					Query: "SELECT * FROM (SELECT c_id AS c_c_id, bill FROM c) sq1, LATERAL (SELECT row_number() OVER () AS rownum FROM o WHERE c_id = c_c_id) sq2 ORDER BY c_c_id, bill, rownum",
					Expected: []sql.Row{
						{1, "CA", int64(1)},
						{1, "CA", int64(2)},
						{1, "CA", int64(3)},
						{2, "TX", int64(1)},
						{2, "TX", int64(2)},
						{2, "TX", int64(3)},
						{4, "TX", int64(1)},
						{4, "TX", int64(2)},
						{6, "FL", int64(1)},
					},
				},
				// ORDER BY on rank alias with LIMIT (multi-hop Limit→Sort→Window chain)
				{
					Query: "SELECT c_id, rank() OVER (ORDER BY c_id) AS rnk FROM c ORDER BY rnk DESC LIMIT 3",
					Expected: []sql.Row{
						{6, int64(6)},
						{5, int64(5)},
						{4, int64(4)},
					},
				},
				// ORDER BY on rank alias (Sort→Window chain)
				{
					Query: "SELECT c_id, rank() OVER (ORDER BY c_id) AS r FROM c ORDER BY r",
					Expected: []sql.Row{
						{1, int64(1)},
						{2, int64(2)},
						{3, int64(3)},
						{4, int64(4)},
						{5, int64(5)},
						{6, int64(6)},
					},
				},
				// DISTINCT + ORDER BY on rank alias
				{
					Query: "SELECT DISTINCT c_id, rank() OVER (ORDER BY c_id) AS r FROM c ORDER BY r",
					Expected: []sql.Row{
						{1, int64(1)},
						{2, int64(2)},
						{3, int64(3)},
						{4, int64(4)},
						{5, int64(5)},
						{6, int64(6)},
					},
				},
				// CASE expression over rank() subquery (int64 rank values flow into Numeric cast)
				{
					Query: "SELECT sum(CASE WHEN r > 0 THEN 1 ELSE 0 END) FROM (SELECT rank() OVER (ORDER BY c_id) AS r FROM c) t",
					Expected: []sql.Row{
						{int64(6)},
					},
				},
				// window SUM with non-empty result (GMS SumAgg.Compute returns float64, not int32)
				{
					Query: "SELECT c_id, SUM(o_id) OVER (PARTITION BY c_id) AS s FROM o WHERE c_id = 1 ORDER BY o_id",
					Expected: []sql.Row{
						{1, int64(60)},
						{1, int64(60)},
						{1, int64(60)},
					},
				},
			},
		},
		{
			Name: "window SUM/AVG wrapped in a subquery projection",
			SetUpScript: []string{
				"CREATE TABLE wrapper_probe (grp INT, val INT);",
				"INSERT INTO wrapper_probe VALUES (1, 10), (1, 20), (2, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT grp, val, grp_total FROM (SELECT grp, val, SUM(val) OVER (PARTITION BY grp) AS grp_total FROM wrapper_probe) sub ORDER BY grp, val;",
					Expected: []sql.Row{
						{1, 10, int64(30)},
						{1, 20, int64(30)},
						{2, 5, int64(5)},
					},
				},
				{
					Query: "SELECT grp, val, grp_avg FROM (SELECT grp, val, AVG(val) OVER (PARTITION BY grp) AS grp_avg FROM wrapper_probe) sub ORDER BY grp, val;",
					Expected: []sql.Row{
						{1, 10, Numeric("15")},
						{1, 20, Numeric("15")},
						{2, 5, Numeric("5")},
					},
				},
			},
		},
	})
}
