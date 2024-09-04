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

func TestBox(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_box)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_box,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE BOX_TBL (f1 box);`,
			},
			{
				Statement: `INSERT INTO BOX_TBL (f1) VALUES ('(2.0,2.0,0.0,0.0)');`,
			},
			{
				Statement: `INSERT INTO BOX_TBL (f1) VALUES ('(1.0,1.0,3.0,3.0)');`,
			},
			{
				Statement: `INSERT INTO BOX_TBL (f1) VALUES ('((-8, 2), (-2, -10))');`,
			},
			{
				Statement: `INSERT INTO BOX_TBL (f1) VALUES ('(2.5, 2.5, 2.5,3.5)');`,
			},
			{
				Statement: `INSERT INTO BOX_TBL (f1) VALUES ('(3.0, 3.0,3.0,3.0)');`,
			},
			{
				Statement:   `INSERT INTO BOX_TBL (f1) VALUES ('(2.3, 4.5)');`,
				ErrorString: `invalid input syntax for type box: "(2.3, 4.5)"`,
			},
			{
				Statement:   `INSERT INTO BOX_TBL (f1) VALUES ('[1, 2, 3, 4)');`,
				ErrorString: `invalid input syntax for type box: "[1, 2, 3, 4)"`,
			},
			{
				Statement:   `INSERT INTO BOX_TBL (f1) VALUES ('(1, 2, 3, 4]');`,
				ErrorString: `invalid input syntax for type box: "(1, 2, 3, 4]"`,
			},
			{
				Statement:   `INSERT INTO BOX_TBL (f1) VALUES ('(1, 2, 3, 4) x');`,
				ErrorString: `invalid input syntax for type box: "(1, 2, 3, 4) x"`,
			},
			{
				Statement:   `INSERT INTO BOX_TBL (f1) VALUES ('asdfasdf(ad');`,
				ErrorString: `invalid input syntax for type box: "asdfasdf(ad"`,
			},
			{
				Statement: `SELECT * FROM BOX_TBL;`,
				Results:   []sql.Row{{`(2,2),(0,0)`}, {`(3,3),(1,1)`}, {`(-2,2),(-8,-10)`}, {`(2.5,3.5),(2.5,2.5)`}, {`(3,3),(3,3)`}},
			},
			{
				Statement: `SELECT b.*, area(b.f1) as barea
   FROM BOX_TBL b;`,
				Results: []sql.Row{{`(2,2),(0,0)`, 4}, {`(3,3),(1,1)`, 4}, {`(-2,2),(-8,-10)`, 72}, {`(2.5,3.5),(2.5,2.5)`, 0}, {`(3,3),(3,3)`, 0}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE b.f1 && box '(2.5,2.5,1.0,1.0)';`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(3,3),(1,1)`}, {`(2.5,3.5),(2.5,2.5)`}},
			},
			{
				Statement: `SELECT b1.*
   FROM BOX_TBL b1
   WHERE b1.f1 &< box '(2.0,2.0,2.5,2.5)';`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(-2,2),(-8,-10)`}, {`(2.5,3.5),(2.5,2.5)`}},
			},
			{
				Statement: `SELECT b1.*
   FROM BOX_TBL b1
   WHERE b1.f1 &> box '(2.0,2.0,2.5,2.5)';`,
				Results: []sql.Row{{`(2.5,3.5),(2.5,2.5)`}, {`(3,3),(3,3)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE b.f1 << box '(3.0,3.0,5.0,5.0)';`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(-2,2),(-8,-10)`}, {`(2.5,3.5),(2.5,2.5)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE b.f1 <= box '(3.0,3.0,5.0,5.0)';`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(3,3),(1,1)`}, {`(2.5,3.5),(2.5,2.5)`}, {`(3,3),(3,3)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE b.f1 < box '(3.0,3.0,5.0,5.0)';`,
				Results: []sql.Row{{`(2.5,3.5),(2.5,2.5)`}, {`(3,3),(3,3)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE b.f1 = box '(3.0,3.0,5.0,5.0)';`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(3,3),(1,1)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b				-- zero area
   WHERE b.f1 > box '(3.5,3.0,4.5,3.0)';`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(3,3),(1,1)`}, {`(-2,2),(-8,-10)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b				-- zero area
   WHERE b.f1 >= box '(3.5,3.0,4.5,3.0)';`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(3,3),(1,1)`}, {`(-2,2),(-8,-10)`}, {`(2.5,3.5),(2.5,2.5)`}, {`(3,3),(3,3)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE box '(3.0,3.0,5.0,5.0)' >> b.f1;`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(-2,2),(-8,-10)`}, {`(2.5,3.5),(2.5,2.5)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE b.f1 <@ box '(0,0,3,3)';`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(3,3),(1,1)`}, {`(3,3),(3,3)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE box '(0,0,3,3)' @> b.f1;`,
				Results: []sql.Row{{`(2,2),(0,0)`}, {`(3,3),(1,1)`}, {`(3,3),(3,3)`}},
			},
			{
				Statement: `SELECT b.f1
   FROM BOX_TBL b
   WHERE box '(1,1,3,3)' ~= b.f1;`,
				Results: []sql.Row{{`(3,3),(1,1)`}},
			},
			{
				Statement: `SELECT @@(b1.f1) AS p
   FROM BOX_TBL b1;`,
				Results: []sql.Row{{`(1,1)`}, {`(2,2)`}, {`(-5,-4)`}, {`(2.5,3)`}, {`(3,3)`}},
			},
			{
				Statement: `SELECT b1.*, b2.*
   FROM BOX_TBL b1, BOX_TBL b2
   WHERE b1.f1 @> b2.f1 and not b1.f1 ~= b2.f1;`,
				Results: []sql.Row{{`(3,3),(1,1)`, `(3,3),(3,3)`}},
			},
			{
				Statement: `SELECT height(f1), width(f1) FROM BOX_TBL;`,
				Results:   []sql.Row{{2, 2}, {2, 2}, {12, 6}, {1, 0}, {0, 0}},
			},
			{
				Statement: `CREATE TEMPORARY TABLE box_temp (f1 box);`,
			},
			{
				Statement: `INSERT INTO box_temp
	SELECT box(point(i, i), point(i * 2, i * 2))
	FROM generate_series(1, 50) AS i;`,
			},
			{
				Statement: `CREATE INDEX box_spgist ON box_temp USING spgist (f1);`,
			},
			{
				Statement: `INSERT INTO box_temp
	VALUES (NULL),
		   ('(0,0)(0,100)'),
		   ('(-3,4.3333333333)(40,1)'),
		   ('(0,100)(0,infinity)'),
		   ('(-infinity,0)(0,infinity)'),
		   ('(-infinity,-infinity)(infinity,infinity)');`,
			},
			{
				Statement: `SET enable_seqscan = false;`,
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 << '(10,20),(30,40)';`,
				Results:   []sql.Row{{`(2,2),(1,1)`}, {`(4,4),(2,2)`}, {`(6,6),(3,3)`}, {`(8,8),(4,4)`}, {`(0,100),(0,0)`}, {`(0,Infinity),(0,100)`}, {`(0,Infinity),(-Infinity,0)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 << '(10,20),(30,40)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 << '(30,40),(10,20)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 &< '(10,4.333334),(5,100)';`,
				Results:   []sql.Row{{`(2,2),(1,1)`}, {`(4,4),(2,2)`}, {`(6,6),(3,3)`}, {`(8,8),(4,4)`}, {`(10,10),(5,5)`}, {`(0,100),(0,0)`}, {`(0,Infinity),(0,100)`}, {`(0,Infinity),(-Infinity,0)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 &< '(10,4.333334),(5,100)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 &< '(10,100),(5,4.333334)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 && '(15,20),(25,30)';`,
				Results:   []sql.Row{{`(20,20),(10,10)`}, {`(22,22),(11,11)`}, {`(24,24),(12,12)`}, {`(26,26),(13,13)`}, {`(28,28),(14,14)`}, {`(30,30),(15,15)`}, {`(32,32),(16,16)`}, {`(34,34),(17,17)`}, {`(36,36),(18,18)`}, {`(38,38),(19,19)`}, {`(40,40),(20,20)`}, {`(42,42),(21,21)`}, {`(44,44),(22,22)`}, {`(46,46),(23,23)`}, {`(48,48),(24,24)`}, {`(50,50),(25,25)`}, {`(Infinity,Infinity),(-Infinity,-Infinity)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 && '(15,20),(25,30)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 && '(25,30),(15,20)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 &> '(40,30),(45,50)';`,
				Results:   []sql.Row{{`(80,80),(40,40)`}, {`(82,82),(41,41)`}, {`(84,84),(42,42)`}, {`(86,86),(43,43)`}, {`(88,88),(44,44)`}, {`(90,90),(45,45)`}, {`(92,92),(46,46)`}, {`(94,94),(47,47)`}, {`(96,96),(48,48)`}, {`(98,98),(49,49)`}, {`(100,100),(50,50)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 &> '(40,30),(45,50)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 &> '(45,50),(40,30)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 >> '(30,40),(40,30)';`,
				Results:   []sql.Row{{`(82,82),(41,41)`}, {`(84,84),(42,42)`}, {`(86,86),(43,43)`}, {`(88,88),(44,44)`}, {`(90,90),(45,45)`}, {`(92,92),(46,46)`}, {`(94,94),(47,47)`}, {`(96,96),(48,48)`}, {`(98,98),(49,49)`}, {`(100,100),(50,50)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 >> '(30,40),(40,30)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 >> '(40,40),(30,30)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 <<| '(10,4.33334),(5,100)';`,
				Results:   []sql.Row{{`(2,2),(1,1)`}, {`(4,4),(2,2)`}, {`(40,4.3333333333),(-3,1)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 <<| '(10,4.33334),(5,100)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 <<| '(10,100),(5,4.33334)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 &<| '(10,4.3333334),(5,1)';`,
				Results:   []sql.Row{{`(2,2),(1,1)`}, {`(4,4),(2,2)`}, {`(40,4.3333333333),(-3,1)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 &<| '(10,4.3333334),(5,1)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 &<| '(10,4.3333334),(5,1)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 |&> '(49.99,49.99),(49.99,49.99)';`,
				Results:   []sql.Row{{`(100,100),(50,50)`}, {`(0,Infinity),(0,100)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 |&> '(49.99,49.99),(49.99,49.99)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 |&> '(49.99,49.99),(49.99,49.99)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 |>> '(37,38),(39,40)';`,
				Results:   []sql.Row{{`(82,82),(41,41)`}, {`(84,84),(42,42)`}, {`(86,86),(43,43)`}, {`(88,88),(44,44)`}, {`(90,90),(45,45)`}, {`(92,92),(46,46)`}, {`(94,94),(47,47)`}, {`(96,96),(48,48)`}, {`(98,98),(49,49)`}, {`(100,100),(50,50)`}, {`(0,Infinity),(0,100)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 |>> '(37,38),(39,40)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 |>> '(39,40),(37,38)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 @> '(10,11),(15,16)';`,
				Results:   []sql.Row{{`(16,16),(8,8)`}, {`(18,18),(9,9)`}, {`(20,20),(10,10)`}, {`(Infinity,Infinity),(-Infinity,-Infinity)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 @> '(10,11),(15,15)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 @> '(15,15),(10,11)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 <@ '(10,15),(30,35)';`,
				Results:   []sql.Row{{`(30,30),(15,15)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 <@ '(10,15),(30,35)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 <@ '(30,35),(10,15)'::box)`}},
			},
			{
				Statement: `SELECT * FROM box_temp WHERE f1 ~= '(20,20),(40,40)';`,
				Results:   []sql.Row{{`(40,40),(20,20)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM box_temp WHERE f1 ~= '(20,20),(40,40)';`,
				Results:   []sql.Row{{`Index Only Scan using box_spgist on box_temp`}, {`Index Cond: (f1 ~= '(40,40),(20,20)'::box)`}},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `DROP INDEX box_spgist;`,
			},
			{
				Statement: `CREATE TABLE quad_box_tbl (id int, b box);`,
			},
			{
				Statement: `INSERT INTO quad_box_tbl
  SELECT (x - 1) * 100 + y, box(point(x * 10, y * 10), point(x * 10 + 5, y * 10 + 5))
  FROM generate_series(1, 100) x,
       generate_series(1, 100) y;`,
			},
			{
				Statement: `INSERT INTO quad_box_tbl
  SELECT i, '((200, 300),(210, 310))'
  FROM generate_series(10001, 11000) AS i;`,
			},
			{
				Statement: `INSERT INTO quad_box_tbl
VALUES
  (11001, NULL),
  (11002, NULL),
  (11003, '((-infinity,-infinity),(infinity,infinity))'),
  (11004, '((-infinity,100),(-infinity,500))'),
  (11005, '((-infinity,-infinity),(700,infinity))');`,
			},
			{
				Statement: `CREATE INDEX quad_box_tbl_idx ON quad_box_tbl USING spgist(b);`,
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
				Statement: `CREATE TABLE quad_box_tbl_ord_seq1 AS
SELECT rank() OVER (ORDER BY b <-> point '123,456') n, b <-> point '123,456' dist, id
FROM quad_box_tbl;`,
			},
			{
				Statement: `CREATE TABLE quad_box_tbl_ord_seq2 AS
SELECT rank() OVER (ORDER BY b <-> point '123,456') n, b <-> point '123,456' dist, id
FROM quad_box_tbl WHERE b <@ box '((200,300),(500,600))';`,
			},
			{
				Statement: `SET enable_seqscan = OFF;`,
			},
			{
				Statement: `SET enable_indexscan = ON;`,
			},
			{
				Statement: `SET enable_bitmapscan = ON;`,
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b <<  box '((100,200),(300,500))';`,
				Results:   []sql.Row{{901}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b &<  box '((100,200),(300,500))';`,
				Results:   []sql.Row{{3901}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b &&  box '((100,200),(300,500))';`,
				Results:   []sql.Row{{1653}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b &>  box '((100,200),(300,500))';`,
				Results:   []sql.Row{{10100}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b >>  box '((100,200),(300,500))';`,
				Results:   []sql.Row{{7000}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b >>  box '((100,200),(300,500))';`,
				Results:   []sql.Row{{7000}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b <<| box '((100,200),(300,500))';`,
				Results:   []sql.Row{{1900}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b &<| box '((100,200),(300,500))';`,
				Results:   []sql.Row{{5901}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b |&> box '((100,200),(300,500))';`,
				Results:   []sql.Row{{9100}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b |>> box '((100,200),(300,500))';`,
				Results:   []sql.Row{{5000}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b @>  box '((201,301),(202,303))';`,
				Results:   []sql.Row{{1003}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b <@  box '((100,200),(300,500))';`,
				Results:   []sql.Row{{1600}},
			},
			{
				Statement: `SELECT count(*) FROM quad_box_tbl WHERE b ~=  box '((200,300),(205,305))';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SET enable_indexscan = ON;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT rank() OVER (ORDER BY b <-> point '123,456') n, b <-> point '123,456' dist, id
FROM quad_box_tbl;`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Scan using quad_box_tbl_idx on quad_box_tbl`}, {`Order By: (b <-> '(123,456)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE quad_box_tbl_ord_idx1 AS
SELECT rank() OVER (ORDER BY b <-> point '123,456') n, b <-> point '123,456' dist, id
FROM quad_box_tbl;`,
			},
			{
				Statement: `SELECT *
FROM quad_box_tbl_ord_seq1 seq FULL JOIN quad_box_tbl_ord_idx1 idx
	ON seq.n = idx.n AND seq.id = idx.id AND
		(seq.dist = idx.dist OR seq.dist IS NULL AND idx.dist IS NULL)
WHERE seq.id IS NULL OR idx.id IS NULL;`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT rank() OVER (ORDER BY b <-> point '123,456') n, b <-> point '123,456' dist, id
FROM quad_box_tbl WHERE b <@ box '((200,300),(500,600))';`,
				Results: []sql.Row{{`WindowAgg`}, {`->  Index Scan using quad_box_tbl_idx on quad_box_tbl`}, {`Index Cond: (b <@ '(500,600),(200,300)'::box)`}, {`Order By: (b <-> '(123,456)'::point)`}},
			},
			{
				Statement: `CREATE TEMP TABLE quad_box_tbl_ord_idx2 AS
SELECT rank() OVER (ORDER BY b <-> point '123,456') n, b <-> point '123,456' dist, id
FROM quad_box_tbl WHERE b <@ box '((200,300),(500,600))';`,
			},
			{
				Statement: `SELECT *
FROM quad_box_tbl_ord_seq2 seq FULL JOIN quad_box_tbl_ord_idx2 idx
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
