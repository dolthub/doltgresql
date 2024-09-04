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

func TestPolygon(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_polygon)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_polygon,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE POLYGON_TBL(f1 polygon);`,
			},
			{
				Statement: `INSERT INTO POLYGON_TBL(f1) VALUES ('(2.0,0.0),(2.0,4.0),(0.0,0.0)');`,
			},
			{
				Statement: `INSERT INTO POLYGON_TBL(f1) VALUES ('(3.0,1.0),(3.0,3.0),(1.0,0.0)');`,
			},
			{
				Statement: `INSERT INTO POLYGON_TBL(f1) VALUES ('(1,2),(3,4),(5,6),(7,8)');`,
			},
			{
				Statement: `INSERT INTO POLYGON_TBL(f1) VALUES ('(7,8),(5,6),(3,4),(1,2)'); -- Reverse`,
			},
			{
				Statement: `INSERT INTO POLYGON_TBL(f1) VALUES ('(1,2),(7,8),(5,6),(3,-4)');`,
			},
			{
				Statement: `INSERT INTO POLYGON_TBL(f1) VALUES ('(0.0,0.0)');`,
			},
			{
				Statement: `INSERT INTO POLYGON_TBL(f1) VALUES ('(0.0,1.0),(0.0,1.0)');`,
			},
			{
				Statement:   `INSERT INTO POLYGON_TBL(f1) VALUES ('0.0');`,
				ErrorString: `invalid input syntax for type polygon: "0.0"`,
			},
			{
				Statement:   `INSERT INTO POLYGON_TBL(f1) VALUES ('(0.0 0.0');`,
				ErrorString: `invalid input syntax for type polygon: "(0.0 0.0"`,
			},
			{
				Statement:   `INSERT INTO POLYGON_TBL(f1) VALUES ('(0,1,2)');`,
				ErrorString: `invalid input syntax for type polygon: "(0,1,2)"`,
			},
			{
				Statement:   `INSERT INTO POLYGON_TBL(f1) VALUES ('(0,1,2,3');`,
				ErrorString: `invalid input syntax for type polygon: "(0,1,2,3"`,
			},
			{
				Statement:   `INSERT INTO POLYGON_TBL(f1) VALUES ('asdf');`,
				ErrorString: `invalid input syntax for type polygon: "asdf"`,
			},
			{
				Statement: `SELECT * FROM POLYGON_TBL;`,
				Results:   []sql.Row{{`((2,0),(2,4),(0,0))`}, {`((3,1),(3,3),(1,0))`}, {`((1,2),(3,4),(5,6),(7,8))`}, {`((7,8),(5,6),(3,4),(1,2))`}, {`((1,2),(7,8),(5,6),(3,-4))`}, {`((0,0))`}, {`((0,1),(0,1))`}},
			},
			{
				Statement: `CREATE TABLE quad_poly_tbl (id int, p polygon);`,
			},
			{
				Statement: `INSERT INTO quad_poly_tbl
	SELECT (x - 1) * 100 + y, polygon(circle(point(x * 10, y * 10), 1 + (x + y) % 10))
	FROM generate_series(1, 100) x,
		 generate_series(1, 100) y;`,
			},
			{
				Statement: `INSERT INTO quad_poly_tbl
	SELECT i, polygon '((200, 300),(210, 310),(230, 290))'
	FROM generate_series(10001, 11000) AS i;`,
			},
			{
				Statement: `INSERT INTO quad_poly_tbl
	VALUES
		(11001, NULL),
		(11002, NULL),
		(11003, NULL);`,
			},
			{
				Statement: `CREATE INDEX quad_poly_tbl_idx ON quad_poly_tbl USING spgist(p);`,
			},
			{
				Statement: `SET enable_seqscan = ON;`,
			},
			{
				Statement: `SET enable_indexscan = OFF;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `CREATE TEMP TABLE quad_poly_tbl_ord_seq2 AS
SELECT rank() OVER (ORDER BY p <-> point '123,456') n, p <-> point '123,456' dist, id
FROM quad_poly_tbl WHERE p <@ polygon '((300,300),(400,600),(600,500),(700,200))';`,
			},
			{
				Statement: `SET enable_seqscan = OFF;`,
			},
			{
				Statement: `SET enable_indexscan = OFF;`,
			},
			{
				Statement: `SET enable_bitmapscan = ON;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p << polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p << '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p << '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p << polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{3890}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p &< polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p &< '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p &< '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p &< polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{7900}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p && polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p && '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p && '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p && polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{977}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p &> polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p &> '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p &> '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p &> polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{7000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p >> polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p >> '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p >> '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p >> polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{2990}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p <<| polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p <<| '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p <<| '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p <<| polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{1890}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p &<| polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p &<| '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p &<| '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p &<| polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{6900}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p |&> polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p |&> '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p |&> '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p |&> polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{9000}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p |>> polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p |>> '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p |>> '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p |>> polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{3990}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p <@ polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p <@ '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p <@ '((300,300),(400,600),(600,500),(700,200))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p <@ polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results:   []sql.Row{{831}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p @> polygon '((340,550),(343,552),(341,553))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p @> '((340,550),(343,552),(341,553))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p @> '((340,550),(343,552),(341,553))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p @> polygon '((340,550),(343,552),(341,553))';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM quad_poly_tbl WHERE p ~= polygon '((200, 300),(210, 310),(230, 290))';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on quad_poly_tbl`}, {`Recheck Cond: (p ~= '((200,300),(210,310),(230,290))'::polygon)`}, {`->  Bitmap Index Scan on quad_poly_tbl_idx`}, {`Index Cond: (p ~= '((200,300),(210,310),(230,290))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM quad_poly_tbl WHERE p ~= polygon '((200, 300),(210, 310),(230, 290))';`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `SET enable_indexscan = ON;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT rank() OVER (ORDER BY p <-> point '123,456') n, p <-> point '123,456' dist, id
FROM quad_poly_tbl WHERE p <@ polygon '((300,300),(400,600),(600,500),(700,200))';`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Scan using quad_poly_tbl_idx on quad_poly_tbl`}, {`Index Cond: (p <@ '((300,300),(400,600),(600,500),(700,200))'::polygon)`}, {`Order By: (p <-> '(123,456)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE quad_poly_tbl_ord_idx2 AS
SELECT rank() OVER (ORDER BY p <-> point '123,456') n, p <-> point '123,456' dist, id
FROM quad_poly_tbl WHERE p <@ polygon '((300,300),(400,600),(600,500),(700,200))';`,
			},
			{
				Statement: `SELECT *
FROM quad_poly_tbl_ord_seq2 seq FULL JOIN quad_poly_tbl_ord_idx2 idx
	ON seq.n = idx.n AND seq.id = idx.id AND
		(seq.dist = idx.dist OR seq.dist IS NULL AND idx.dist IS NULL)
WHERE seq.id IS NULL OR idx.id IS NULL;`,
				Results: []sql.Row{},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `RESET enable_indexscan;`,
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
		},
	})
}
