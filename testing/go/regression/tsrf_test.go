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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestTsrf(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_tsrf)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_tsrf,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT generate_series(1, 3);`,
				Results:   []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement: `SELECT generate_series(1, 3), generate_series(3,5);`,
				Results:   []sql.Row{{1, 3}, {2, 4}, {3, 5}},
			},
			{
				Statement: `SELECT generate_series(1, 2), generate_series(1,4);`,
				Results:   []sql.Row{{1, 1}, {2, 2}, {``, 3}, {``, 4}},
			},
			{
				Statement: `SELECT generate_series(1, generate_series(1, 3));`,
				Results:   []sql.Row{{1}, {1}, {2}, {1}, {2}, {3}},
			},
			{
				Statement:   `SELECT * FROM generate_series(1, generate_series(1, 3));`,
				ErrorString: `set-returning functions must appear at top level of FROM`,
			},
			{
				Statement: `SELECT generate_series(generate_series(1,3), generate_series(2, 4));`,
				Results:   []sql.Row{{1}, {2}, {2}, {3}, {3}, {4}},
			},
			{
				Statement: `explain (verbose, costs off)
SELECT generate_series(1, generate_series(1, 3)), generate_series(2, 4);`,
				Results: []sql.Row{{`ProjectSet`}, {`Output: generate_series(1, (generate_series(1, 3))), (generate_series(2, 4))`}, {`->  ProjectSet`}, {`Output: generate_series(1, 3), generate_series(2, 4)`}, {`->  Result`}},
			},
			{
				Statement: `SELECT generate_series(1, generate_series(1, 3)), generate_series(2, 4);`,
				Results:   []sql.Row{{1, 2}, {1, 3}, {2, 3}, {1, 4}, {2, 4}, {3, 4}},
			},
			{
				Statement: `CREATE TABLE few(id int, dataa text, datab text);`,
			},
			{
				Statement: `INSERT INTO few VALUES(1, 'a', 'foo'),(2, 'a', 'bar'),(3, 'b', 'bar');`,
			},
			{
				Statement: `explain (verbose, costs off)
SELECT unnest(ARRAY[1, 2]) FROM few WHERE false;`,
				Results: []sql.Row{{`ProjectSet`}, {`Output: unnest('{1,2}'::integer[])`}, {`->  Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `SELECT unnest(ARRAY[1, 2]) FROM few WHERE false;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `explain (verbose, costs off)
SELECT * FROM few f1,
  (SELECT unnest(ARRAY[1,2]) FROM few f2 WHERE false OFFSET 0) ss;`,
				Results: []sql.Row{{`Result`}, {`Output: f1.id, f1.dataa, f1.datab, ss.unnest`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `SELECT * FROM few f1,
  (SELECT unnest(ARRAY[1,2]) FROM few f2 WHERE false OFFSET 0) ss;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT few.id, generate_series(1,3) g FROM few ORDER BY id DESC;`,
				Results:   []sql.Row{{3, 1}, {3, 2}, {3, 3}, {2, 1}, {2, 2}, {2, 3}, {1, 1}, {1, 2}, {1, 3}},
			},
			{
				Statement: `SELECT few.id, generate_series(1,3) g FROM few ORDER BY id, g DESC;`,
				Results:   []sql.Row{{1, 3}, {1, 2}, {1, 1}, {2, 3}, {2, 2}, {2, 1}, {3, 3}, {3, 2}, {3, 1}},
			},
			{
				Statement: `SELECT few.id, generate_series(1,3) g FROM few ORDER BY id, generate_series(1,3) DESC;`,
				Results:   []sql.Row{{1, 3}, {1, 2}, {1, 1}, {2, 3}, {2, 2}, {2, 1}, {3, 3}, {3, 2}, {3, 1}},
			},
			{
				Statement: `SELECT few.id FROM few ORDER BY id, generate_series(1,3) DESC;`,
				Results:   []sql.Row{{1}, {1}, {1}, {2}, {2}, {2}, {3}, {3}, {3}},
			},
			{
				Statement: `SET enable_hashagg TO 0; -- stable output order`,
			},
			{
				Statement: `SELECT few.dataa, count(*), min(id), max(id), unnest('{1,1,3}'::int[]) FROM few WHERE few.id = 1 GROUP BY few.dataa;`,
				Results:   []sql.Row{{`a`, 1, 1, 1, 1}, {`a`, 1, 1, 1, 1}, {`a`, 1, 1, 1, 3}},
			},
			{
				Statement: `SELECT few.dataa, count(*), min(id), max(id), unnest('{1,1,3}'::int[]) FROM few WHERE few.id = 1 GROUP BY few.dataa, unnest('{1,1,3}'::int[]);`,
				Results:   []sql.Row{{`a`, 2, 1, 1, 1}, {`a`, 1, 1, 1, 3}},
			},
			{
				Statement: `SELECT few.dataa, count(*), min(id), max(id), unnest('{1,1,3}'::int[]) FROM few WHERE few.id = 1 GROUP BY few.dataa, 5;`,
				Results:   []sql.Row{{`a`, 2, 1, 1, 1}, {`a`, 1, 1, 1, 3}},
			},
			{
				Statement: `RESET enable_hashagg;`,
			},
			{
				Statement: `SELECT dataa, generate_series(1,1), count(*) FROM few GROUP BY 1 HAVING count(*) > 1;`,
				Results:   []sql.Row{{`a`, 1, 2}},
			},
			{
				Statement: `SELECT dataa, generate_series(1,1), count(*) FROM few GROUP BY 1, 2 HAVING count(*) > 1;`,
				Results:   []sql.Row{{`a`, 1, 2}},
			},
			{
				Statement: `SELECT few.dataa, count(*) FROM few WHERE dataa = 'a' GROUP BY few.dataa ORDER BY 2;`,
				Results:   []sql.Row{{`a`, 2}},
			},
			{
				Statement: `SELECT few.dataa, count(*) FROM few WHERE dataa = 'a' GROUP BY few.dataa, unnest('{1,1,3}'::int[]) ORDER BY 2;`,
				Results:   []sql.Row{{`a`, 2}, {`a`, 4}},
			},
			{
				Statement:   `SELECT q1, case when q1 > 0 then generate_series(1,3) else 0 end FROM int8_tbl;`,
				ErrorString: `set-returning functions are not allowed in CASE`,
			},
			{
				Statement:   `SELECT q1, coalesce(generate_series(1,3), 0) FROM int8_tbl;`,
				ErrorString: `set-returning functions are not allowed in COALESCE`,
			},
			{
				Statement:   `SELECT min(generate_series(1, 3)) FROM few;`,
				ErrorString: `aggregate function calls cannot contain set-returning function calls`,
			},
			{
				Statement: `SELECT sum((3 = ANY(SELECT generate_series(1,4)))::int);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT sum((3 = ANY(SELECT lag(x) over(order by x)
                    FROM generate_series(1,4) x))::int);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement:   `SELECT min(generate_series(1, 3)) OVER() FROM few;`,
				ErrorString: `window function calls cannot contain set-returning function calls`,
			},
			{
				Statement: `SELECT id,lag(id) OVER(), count(*) OVER(), generate_series(1,3) FROM few;`,
				Results:   []sql.Row{{1, ``, 3, 1}, {1, ``, 3, 2}, {1, ``, 3, 3}, {2, 1, 3, 1}, {2, 1, 3, 2}, {2, 1, 3, 3}, {3, 2, 3, 1}, {3, 2, 3, 2}, {3, 2, 3, 3}},
			},
			{
				Statement: `SELECT SUM(count(*)) OVER(PARTITION BY generate_series(1,3) ORDER BY generate_series(1,3)), generate_series(1,3) g FROM few GROUP BY g;`,
				Results:   []sql.Row{{3, 1}, {3, 2}, {3, 3}},
			},
			{
				Statement: `SELECT few.dataa, count(*), min(id), max(id), generate_series(1,3) FROM few GROUP BY few.dataa ORDER BY 5, 1;`,
				Results:   []sql.Row{{`a`, 2, 1, 2, 1}, {`b`, 1, 3, 3, 1}, {`a`, 2, 1, 2, 2}, {`b`, 1, 3, 3, 2}, {`a`, 2, 1, 2, 3}, {`b`, 1, 3, 3, 3}},
			},
			{
				Statement: `set enable_hashagg = false;`,
			},
			{
				Statement: `SELECT dataa, datab b, generate_series(1,2) g, count(*) FROM few GROUP BY CUBE(dataa, datab);`,
				Results:   []sql.Row{{`a`, `bar`, 1, 1}, {`a`, `bar`, 2, 1}, {`a`, `foo`, 1, 1}, {`a`, `foo`, 2, 1}, {`a`, ``, 1, 2}, {`a`, ``, 2, 2}, {`b`, `bar`, 1, 1}, {`b`, `bar`, 2, 1}, {`b`, ``, 1, 1}, {`b`, ``, 2, 1}, {``, ``, 1, 3}, {``, ``, 2, 3}, {``, `bar`, 1, 2}, {``, `bar`, 2, 2}, {``, `foo`, 1, 1}, {``, `foo`, 2, 1}},
			},
			{
				Statement: `SELECT dataa, datab b, generate_series(1,2) g, count(*) FROM few GROUP BY CUBE(dataa, datab) ORDER BY dataa;`,
				Results:   []sql.Row{{`a`, `bar`, 1, 1}, {`a`, `bar`, 2, 1}, {`a`, `foo`, 1, 1}, {`a`, `foo`, 2, 1}, {`a`, ``, 1, 2}, {`a`, ``, 2, 2}, {`b`, `bar`, 1, 1}, {`b`, `bar`, 2, 1}, {`b`, ``, 1, 1}, {`b`, ``, 2, 1}, {``, ``, 1, 3}, {``, ``, 2, 3}, {``, `bar`, 1, 2}, {``, `bar`, 2, 2}, {``, `foo`, 1, 1}, {``, `foo`, 2, 1}},
			},
			{
				Statement: `SELECT dataa, datab b, generate_series(1,2) g, count(*) FROM few GROUP BY CUBE(dataa, datab) ORDER BY g;`,
				Results:   []sql.Row{{`a`, `bar`, 1, 1}, {`a`, `foo`, 1, 1}, {`a`, ``, 1, 2}, {`b`, `bar`, 1, 1}, {`b`, ``, 1, 1}, {``, ``, 1, 3}, {``, `bar`, 1, 2}, {``, `foo`, 1, 1}, {``, `foo`, 2, 1}, {`a`, `bar`, 2, 1}, {`b`, ``, 2, 1}, {`a`, `foo`, 2, 1}, {``, `bar`, 2, 2}, {`a`, ``, 2, 2}, {``, ``, 2, 3}, {`b`, `bar`, 2, 1}},
			},
			{
				Statement: `SELECT dataa, datab b, generate_series(1,2) g, count(*) FROM few GROUP BY CUBE(dataa, datab, g);`,
				Results:   []sql.Row{{`a`, `bar`, 1, 1}, {`a`, `bar`, 2, 1}, {`a`, `bar`, ``, 2}, {`a`, `foo`, 1, 1}, {`a`, `foo`, 2, 1}, {`a`, `foo`, ``, 2}, {`a`, ``, ``, 4}, {`b`, `bar`, 1, 1}, {`b`, `bar`, 2, 1}, {`b`, `bar`, ``, 2}, {`b`, ``, ``, 2}, {``, ``, ``, 6}, {``, `bar`, 1, 2}, {``, `bar`, 2, 2}, {``, `bar`, ``, 4}, {``, `foo`, 1, 1}, {``, `foo`, 2, 1}, {``, `foo`, ``, 2}, {`a`, ``, 1, 2}, {`b`, ``, 1, 1}, {``, ``, 1, 3}, {`a`, ``, 2, 2}, {`b`, ``, 2, 1}, {``, ``, 2, 3}},
			},
			{
				Statement: `SELECT dataa, datab b, generate_series(1,2) g, count(*) FROM few GROUP BY CUBE(dataa, datab, g) ORDER BY dataa;`,
				Results:   []sql.Row{{`a`, `foo`, ``, 2}, {`a`, ``, ``, 4}, {`a`, ``, 2, 2}, {`a`, `bar`, 1, 1}, {`a`, `bar`, 2, 1}, {`a`, `bar`, ``, 2}, {`a`, `foo`, 1, 1}, {`a`, `foo`, 2, 1}, {`a`, ``, 1, 2}, {`b`, `bar`, 1, 1}, {`b`, ``, ``, 2}, {`b`, ``, 1, 1}, {`b`, `bar`, 2, 1}, {`b`, `bar`, ``, 2}, {`b`, ``, 2, 1}, {``, ``, 2, 3}, {``, ``, ``, 6}, {``, `bar`, 1, 2}, {``, `bar`, 2, 2}, {``, `bar`, ``, 4}, {``, `foo`, 1, 1}, {``, `foo`, 2, 1}, {``, `foo`, ``, 2}, {``, ``, 1, 3}},
			},
			{
				Statement: `SELECT dataa, datab b, generate_series(1,2) g, count(*) FROM few GROUP BY CUBE(dataa, datab, g) ORDER BY g;`,
				Results:   []sql.Row{{`a`, `bar`, 1, 1}, {`a`, `foo`, 1, 1}, {`b`, `bar`, 1, 1}, {``, `bar`, 1, 2}, {``, `foo`, 1, 1}, {`a`, ``, 1, 2}, {`b`, ``, 1, 1}, {``, ``, 1, 3}, {`a`, ``, 2, 2}, {`b`, ``, 2, 1}, {``, `bar`, 2, 2}, {``, ``, 2, 3}, {``, `foo`, 2, 1}, {`a`, `bar`, 2, 1}, {`a`, `foo`, 2, 1}, {`b`, `bar`, 2, 1}, {`a`, ``, ``, 4}, {`b`, `bar`, ``, 2}, {`b`, ``, ``, 2}, {``, ``, ``, 6}, {`a`, `foo`, ``, 2}, {`a`, `bar`, ``, 2}, {``, `bar`, ``, 4}, {``, `foo`, ``, 2}},
			},
			{
				Statement: `reset enable_hashagg;`,
			},
			{
				Statement: `explain (verbose, costs off)
select 'foo' as f, generate_series(1,2) as g from few order by 1;`,
				Results: []sql.Row{{`ProjectSet`}, {`Output: 'foo'::text, generate_series(1, 2)`}, {`->  Seq Scan on public.few`}, {`Output: id, dataa, datab`}},
			},
			{
				Statement: `select 'foo' as f, generate_series(1,2) as g from few order by 1;`,
				Results:   []sql.Row{{`foo`, 1}, {`foo`, 2}, {`foo`, 1}, {`foo`, 2}, {`foo`, 1}, {`foo`, 2}},
			},
			{
				Statement: `CREATE TABLE fewmore AS SELECT generate_series(1,3) AS data;`,
			},
			{
				Statement: `INSERT INTO fewmore VALUES(generate_series(4,5));`,
			},
			{
				Statement: `SELECT * FROM fewmore;`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}, {5}},
			},
			{
				Statement:   `UPDATE fewmore SET data = generate_series(4,9);`,
				ErrorString: `set-returning functions are not allowed in UPDATE`,
			},
			{
				Statement:   `INSERT INTO fewmore VALUES(1) RETURNING generate_series(1,3);`,
				ErrorString: `set-returning functions are not allowed in RETURNING`,
			},
			{
				Statement:   `VALUES(1, generate_series(1,2));`,
				ErrorString: `set-returning functions are not allowed in VALUES`,
			},
			{
				Statement: `SELECT int4mul(generate_series(1,2), 10);`,
				Results:   []sql.Row{{10}, {20}},
			},
			{
				Statement: `SELECT generate_series(1,3) IS DISTINCT FROM 2;`,
				Results:   []sql.Row{{true}, {false}, {true}},
			},
			{
				Statement:   `SELECT * FROM int4mul(generate_series(1,2), 10);`,
				ErrorString: `set-returning functions must appear at top level of FROM`,
			},
			{
				Statement: `SELECT DISTINCT ON (a) a, b, generate_series(1,3) g
FROM (VALUES (3, 2), (3,1), (1,1), (1,4), (5,3), (5,1)) AS t(a, b);`,
				Results: []sql.Row{{1, 1, 1}, {3, 2, 1}, {5, 3, 1}},
			},
			{
				Statement: `SELECT DISTINCT ON (a) a, b, generate_series(1,3) g
FROM (VALUES (3, 2), (3,1), (1,1), (1,4), (5,3), (5,1)) AS t(a, b)
ORDER BY a, b DESC;`,
				Results: []sql.Row{{1, 4, 1}, {1, 4, 2}, {1, 4, 3}, {3, 2, 1}, {3, 2, 2}, {3, 2, 3}, {5, 3, 1}, {5, 3, 2}, {5, 3, 3}},
			},
			{
				Statement: `SELECT DISTINCT ON (a) a, b, generate_series(1,3) g
FROM (VALUES (3, 2), (3,1), (1,1), (1,4), (5,3), (5,1)) AS t(a, b)
ORDER BY a, b DESC, g DESC;`,
				Results: []sql.Row{{1, 4, 3}, {3, 2, 3}, {5, 3, 3}},
			},
			{
				Statement: `SELECT DISTINCT ON (a, b, g) a, b, generate_series(1,3) g
FROM (VALUES (3, 2), (3,1), (1,1), (1,4), (5,3), (5,1)) AS t(a, b)
ORDER BY a, b DESC, g DESC;`,
				Results: []sql.Row{{1, 4, 3}, {1, 4, 2}, {1, 4, 1}, {1, 1, 3}, {1, 1, 2}, {1, 1, 1}, {3, 2, 3}, {3, 2, 2}, {3, 2, 1}, {3, 1, 3}, {3, 1, 2}, {3, 1, 1}, {5, 3, 3}, {5, 3, 2}, {5, 3, 1}, {5, 1, 3}, {5, 1, 2}, {5, 1, 1}},
			},
			{
				Statement: `SELECT DISTINCT ON (g) a, b, generate_series(1,3) g
FROM (VALUES (3, 2), (3,1), (1,1), (1,4), (5,3), (5,1)) AS t(a, b);`,
				Results: []sql.Row{{3, 2, 1}, {5, 1, 2}, {3, 1, 3}},
			},
			{
				Statement: `SELECT a, generate_series(1,2) FROM (VALUES(1),(2),(3)) r(a) LIMIT 2 OFFSET 2;`,
				Results:   []sql.Row{{2, 1}, {2, 2}},
			},
			{
				Statement:   `SELECT 1 LIMIT generate_series(1,3);`,
				ErrorString: `set-returning functions are not allowed in LIMIT`,
			},
			{
				Statement: `SELECT (SELECT generate_series(1,3) LIMIT 1 OFFSET few.id) FROM few;`,
				Results:   []sql.Row{{2}, {3}, {``}},
			},
			{
				Statement: `SELECT (SELECT generate_series(1,3) LIMIT 1 OFFSET g.i) FROM generate_series(0,3) g(i);`,
				Results:   []sql.Row{{1}, {2}, {3}, {``}},
			},
			{
				Statement: `CREATE OPERATOR |@| (PROCEDURE = unnest, RIGHTARG = ANYARRAY);`,
			},
			{
				Statement: `SELECT |@|ARRAY[1,2,3];`,
				Results:   []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement: `explain (verbose, costs off)
select generate_series(1,3) as x, generate_series(1,3) + 1 as xp1;`,
				Results: []sql.Row{{`Result`}, {`Output: (generate_series(1, 3)), ((generate_series(1, 3)) + 1)`}, {`->  ProjectSet`}, {`Output: generate_series(1, 3)`}, {`->  Result`}},
			},
			{
				Statement: `select generate_series(1,3) as x, generate_series(1,3) + 1 as xp1;`,
				Results:   []sql.Row{{1, 2}, {2, 3}, {3, 4}},
			},
			{
				Statement: `explain (verbose, costs off)
select generate_series(1,3)+1 order by generate_series(1,3);`,
				Results: []sql.Row{{`Sort`}, {`Output: (((generate_series(1, 3)) + 1)), (generate_series(1, 3))`}, {`Sort Key: (generate_series(1, 3))`}, {`->  Result`}, {`Output: ((generate_series(1, 3)) + 1), (generate_series(1, 3))`}, {`->  ProjectSet`}, {`Output: generate_series(1, 3)`}, {`->  Result`}},
			},
			{
				Statement: `select generate_series(1,3)+1 order by generate_series(1,3);`,
				Results:   []sql.Row{{2}, {3}, {4}},
			},
			{
				Statement: `explain (verbose, costs off)
select generate_series(1,3) as x, generate_series(3,6) + 1 as y;`,
				Results: []sql.Row{{`Result`}, {`Output: (generate_series(1, 3)), ((generate_series(3, 6)) + 1)`}, {`->  ProjectSet`}, {`Output: generate_series(1, 3), generate_series(3, 6)`}, {`->  Result`}},
			},
			{
				Statement: `select generate_series(1,3) as x, generate_series(3,6) + 1 as y;`,
				Results:   []sql.Row{{1, 4}, {2, 5}, {3, 6}, {``, 7}},
			},
			{
				Statement: `DROP TABLE few;`,
			},
			{
				Statement: `DROP TABLE fewmore;`,
			},
		},
	})
}
