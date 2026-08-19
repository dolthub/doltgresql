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
			Name: "named window reference honors ORDER BY from the WINDOW clause",
			SetUpScript: []string{
				"CREATE TABLE t_named(id int, grp int, amt int);",
				"INSERT INTO t_named VALUES (1,1,10),(2,1,20),(3,2,5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT id, SUM(amt) OVER w AS s FROM t_named WINDOW w AS (PARTITION BY grp ORDER BY id) ORDER BY id;",
					Expected: []sql.Row{
						{1, int64(10)},
						{2, int64(30)},
						{3, int64(5)},
					},
				},
				{
					// Inline equivalent must match the named-window-reference result above.
					Query: "SELECT id, SUM(amt) OVER (PARTITION BY grp ORDER BY id) AS s FROM t_named ORDER BY id;",
					Expected: []sql.Row{
						{1, int64(10)},
						{2, int64(30)},
						{3, int64(5)},
					},
				},
			},
		},
		{
			// A named window that only has PARTITION BY (no ORDER BY) gets a "default to
			// full-partition frame" baked in when it's built on its own. If another window
			// then inherits from it and adds an ORDER BY -- either via a second named window
			// (w2 AS (w1 ORDER BY id)) or an inline override (OVER (w1 ORDER BY id)) -- that
			// stale full-partition frame must not survive the merge; the running sum implied
			// by the newly-added ORDER BY has to win.
			Name: "named window inheritance chain honors ORDER BY added by the child",
			SetUpScript: []string{
				"CREATE TABLE t_inherit(id int, grp int, amt int);",
				"INSERT INTO t_inherit VALUES (1,1,10),(2,1,20),(3,1,30),(4,2,5),(5,2,15);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT id, SUM(amt) OVER w2 AS s FROM t_inherit WINDOW w1 AS (PARTITION BY grp), w2 AS (w1 ORDER BY id) ORDER BY id;",
					Expected: []sql.Row{
						{1, int64(10)},
						{2, int64(30)},
						{3, int64(60)},
						{4, int64(5)},
						{5, int64(20)},
					},
				},
				{
					Query: "SELECT id, SUM(amt) OVER (w1 ORDER BY id) AS s FROM t_inherit WINDOW w1 AS (PARTITION BY grp) ORDER BY id;",
					Expected: []sql.Row{
						{1, int64(10)},
						{2, int64(30)},
						{3, int64(60)},
						{4, int64(5)},
						{5, int64(20)},
					},
				},
				{
					// Inline equivalent must match both forms above.
					Query: "SELECT id, SUM(amt) OVER (PARTITION BY grp ORDER BY id) AS s FROM t_inherit ORDER BY id;",
					Expected: []sql.Row{
						{1, int64(10)},
						{2, int64(30)},
						{3, int64(60)},
						{4, int64(5)},
						{5, int64(20)},
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
		{
			Name: "RANGE frame with INTERVAL month boundary is calendar-correct",
			SetUpScript: []string{
				"CREATE TABLE month_edge (d DATE, v INT);",
				"INSERT INTO month_edge VALUES ('2022-01-31', 1), ('2022-02-28', 2), ('2022-03-01', 3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// Jan 31 + 1 month clamps to Feb 28 (2022 isn't a leap year), so the window for the
					// Jan 31 row must include Feb 28 but NOT Mar 1.
					Query: "SELECT sum(v) OVER (ORDER BY d RANGE BETWEEN UNBOUNDED PRECEDING AND INTERVAL '1' MONTH FOLLOWING) FROM month_edge ORDER BY d",
					Expected: []sql.Row{
						{int64(3)},
						{int64(6)},
						{int64(6)},
					},
				},
			},
		},
		{
			Name: "ntile and cume_dist ignore ties/frame and operate over the whole partition",
			SetUpScript: []string{
				"CREATE TABLE rank_ext (id INT PRIMARY KEY, grp INT, val INT);",
				// grp 1 has a tied peer group (val=10 for id 1 and 2); grp 2 has no ties.
				"INSERT INTO rank_ext VALUES (1,1,10),(2,1,10),(3,1,20),(4,1,30),(5,2,5),(6,2,15);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT id, ntile(2) OVER (PARTITION BY grp ORDER BY val) FROM rank_ext ORDER BY id",
					Expected: []sql.Row{
						{1, 1},
						{2, 1},
						{3, 2},
						{4, 2},
						{5, 1},
						{6, 2},
					},
				},
				{
					Query: "SELECT id, cume_dist() OVER (PARTITION BY grp ORDER BY val) FROM rank_ext ORDER BY id",
					Expected: []sql.Row{
						{1, float64(0.5)},
						{2, float64(0.5)},
						{3, float64(0.75)},
						{4, float64(1)},
						{5, float64(0.5)},
						{6, float64(1)},
					},
				},
				{
					// cume_dist (and the other window-only functions with no required arguments) must reject a
					// bare call with a clean error rather than reaching execution with a nil window.
					Query:       "SELECT cume_dist() FROM rank_ext",
					ExpectedErr: "requires an OVER clause",
				},
				{
					Query:       "SELECT row_number() FROM rank_ext",
					ExpectedErr: "requires an OVER clause",
				},
				{
					Query:       "SELECT rank() FROM rank_ext",
					ExpectedErr: "requires an OVER clause",
				},
			},
		},
		{
			Name: "multiple differently-framed numeric RANGE windows in one SELECT don't collide",
			SetUpScript: []string{
				"CREATE TABLE boundary_2 (id INT PRIMARY KEY, val INT);",
				"INSERT INTO boundary_2 VALUES (1,10),(2,20),(3,30);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT
					  sum(val) over (order by id range between 0 preceding and 0 following) as r0,
					  sum(val) over (order by id range between current row and 1 following) as r1foll,
					  sum(val) over (order by id range between unbounded preceding and current row) as runbndprec,
					  sum(val) over (order by id range between current row and unbounded following) as runbndfoll
					FROM boundary_2 ORDER BY id`,
					Expected: []sql.Row{
						{int64(10), int64(30), int64(10), int64(60)},
						{int64(20), int64(50), int64(30), int64(50)},
						{int64(30), int64(30), int64(60), int64(30)},
					},
				},
				{
					Query: "SELECT row_number() over (order by id) as rn1, row_number() over (order by id desc) as rn2 FROM boundary_2 ORDER BY id",
					Expected: []sql.Row{
						{int64(1), int64(3)},
						{int64(2), int64(2)},
						{int64(3), int64(1)},
					},
				},
			},
		},
		{
			Name: "nth_value respects the window's frame and returns the polymorphic argument type",
			SetUpScript: []string{
				"CREATE TABLE nv (id INT PRIMARY KEY, grp INT, val INT);",
				"INSERT INTO nv VALUES (1,1,100),(2,1,200),(3,1,300),(4,2,999);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT id, nth_value(val, 2) OVER (PARTITION BY grp ORDER BY id) FROM nv ORDER BY id",
					Expected: []sql.Row{
						{1, nil},
						{2, 200},
						{3, 200},
						{4, nil},
					},
				},
				{
					Query: "SELECT id, nth_value(val, 2) OVER (PARTITION BY grp ORDER BY id ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) FROM nv ORDER BY id",
					Expected: []sql.Row{
						{1, 200},
						{2, 200},
						{3, 200},
						{4, nil},
					},
				},
			},
		},
		{
			Name: "variance/stddev window functions over an int column",
			SetUpScript: []string{
				"CREATE TABLE t3038 (id BIGINT PRIMARY KEY, grp VARCHAR(10), val INT);",
				"INSERT INTO t3038 VALUES (1,'a',10), (2,'a',20), (3,'b',30), (4,'b',5), (5,'c',15), (6,'c',25);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT id, STDDEV_POP(val) OVER (ORDER BY grp) FROM t3038 ORDER BY id;",
					Expected: []sql.Row{
						{1, Numeric("5.0000000000000000")},
						{2, Numeric("5.0000000000000000")},
						{3, Numeric("9.6014321848357602")},
						{4, Numeric("9.6014321848357602")},
						{5, Numeric("8.5391256382996653")},
						{6, Numeric("8.5391256382996653")},
					},
				},
				{
					Query: "SELECT id, STDDEV_SAMP(val) OVER (ORDER BY grp) FROM t3038 ORDER BY id;",
					Expected: []sql.Row{
						{1, Numeric("7.0710678118654752")},
						{2, Numeric("7.0710678118654752")},
						{3, Numeric("11.0867789130417256")},
						{4, Numeric("11.0867789130417256")},
						{5, Numeric("9.3541434669348535")},
						{6, Numeric("9.3541434669348535")},
					},
				},
				{
					Query: "SELECT id, VAR_POP(val) OVER (ORDER BY grp) FROM t3038 ORDER BY id;",
					Expected: []sql.Row{
						{1, Numeric("25.0000000000000000")},
						{2, Numeric("25.0000000000000000")},
						{3, Numeric("92.1875000000000000")},
						{4, Numeric("92.1875000000000000")},
						{5, Numeric("72.9166666666666667")},
						{6, Numeric("72.9166666666666667")},
					},
				},
				{
					Query: "SELECT id, VAR_SAMP(val) OVER (ORDER BY grp) FROM t3038 ORDER BY id;",
					Expected: []sql.Row{
						{1, Numeric("50.0000000000000000")},
						{2, Numeric("50.0000000000000000")},
						{3, Numeric("122.9166666666666667")},
						{4, Numeric("122.9166666666666667")},
						{5, Numeric("87.5000000000000000")},
						{6, Numeric("87.5000000000000000")},
					},
				},
			},
		},
		{
			Name: "variance/stddev as GROUP BY aggregates, single-row, float8, and aliases",
			SetUpScript: []string{
				"CREATE TABLE t3038b (grp VARCHAR(10), val INT);",
				"INSERT INTO t3038b VALUES ('a',10), ('a',20), ('b',30), ('b',5), ('c',15), ('c',25);",
				"CREATE TABLE t3038_one (val INT);",
				"INSERT INTO t3038_one VALUES (42);",
				"CREATE TABLE t3038_f (val DOUBLE PRECISION);",
				"INSERT INTO t3038_f VALUES (10.0), (20.0);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT grp, VAR_POP(val), VAR_SAMP(val), STDDEV_POP(val), STDDEV_SAMP(val) FROM t3038b GROUP BY grp ORDER BY grp;",
					Expected: []sql.Row{
						{"a", Numeric("25.0000000000000000"), Numeric("50.0000000000000000"), Numeric("5.0000000000000000"), Numeric("7.0710678118654752")},
						{"b", Numeric("156.2500000000000000"), Numeric("312.5000000000000000"), Numeric("12.5000000000000000"), Numeric("17.6776695296636881")},
						{"c", Numeric("25.0000000000000000"), Numeric("50.0000000000000000"), Numeric("5.0000000000000000"), Numeric("7.0710678118654752")},
					},
				},
				{
					// A single row has a well-defined population variance/stddev (0, since there's no
					// spread) but an undefined sample variance/stddev (NULL, not a divide-by-zero panic).
					Query: "SELECT VAR_POP(val), VAR_SAMP(val), STDDEV_POP(val), STDDEV_SAMP(val) FROM t3038_one;",
					Expected: []sql.Row{
						{Numeric("0"), nil, Numeric("0"), nil},
					},
				},
				{
					Query: "SELECT VAR_POP(val), VAR_SAMP(val), STDDEV_POP(val), STDDEV_SAMP(val) FROM t3038_f;",
					Expected: []sql.Row{
						{float64(25), float64(50), float64(5), float64(7.0710678118654755)},
					},
				},
				{
					// variance/stddev must match var_samp/stddev_samp (Postgres semantics)
					Query: "SELECT VARIANCE(val), STDDEV(val) FROM t3038b WHERE grp = 'a';",
					Expected: []sql.Row{
						{Numeric("50.0000000000000000"), Numeric("7.0710678118654752")},
					},
				},
			},
		},
		{
			Name: "variance/stddev over float8 avoid cancellation for large nearly-equal values",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT avg(x::float8), var_pop(x::float8), var_samp(x::float8), stddev_pop(x::float8), stddev_samp(x::float8) FROM (VALUES (100000003), (100000004), (100000006), (100000007)) v(x);",
					Expected: []sql.Row{
						{float64(100000005), float64(2.5), float64(3.3333333333333335), float64(1.5811388300841898), float64(1.8257418583505538)},
					},
				},
				{
					Query: "SELECT avg(x::float8), var_pop(x::float8), var_samp(x::float8), stddev_pop(x::float8), stddev_samp(x::float8) FROM (VALUES (7000000000005), (7000000000007)) v(x);",
					Expected: []sql.Row{
						{float64(7000000000006), float64(1), float64(2), float64(1), float64(1.4142135623730951)},
					},
				},
			},
		},
		{
			Name: "variance/stddev over a real column, as GROUP BY aggregates and window functions",
			SetUpScript: []string{
				"CREATE TABLE t3038r (id INT PRIMARY KEY, grp VARCHAR(10), val REAL);",
				"INSERT INTO t3038r VALUES (1,'a',10), (2,'a',20), (3,'b',30), (4,'b',5), (5,'c',15), (6,'c',25);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT grp, VAR_POP(val), VAR_SAMP(val), STDDEV_POP(val), STDDEV_SAMP(val) FROM t3038r GROUP BY grp ORDER BY grp;",
					Expected: []sql.Row{
						{"a", float64(25), float64(50), float64(5), float64(7.0710678118654755)},
						{"b", float64(156.25), float64(312.5), float64(12.5), float64(17.67766952966369)},
						{"c", float64(25), float64(50), float64(5), float64(7.0710678118654755)},
					},
				},
				{
					Query: "SELECT id, VAR_POP(val) OVER (ORDER BY grp) FROM t3038r ORDER BY id;",
					Expected: []sql.Row{
						{1, float64(25)},
						{2, float64(25)},
						{3, float64(92.1875)},
						{4, float64(92.1875)},
						{5, float64(72.91666666666667)},
						{6, float64(72.91666666666667)},
					},
				},
			},
		},
	})
}
