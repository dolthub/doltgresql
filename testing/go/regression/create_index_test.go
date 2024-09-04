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

func TestCreateIndex(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_index)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_index,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv abs_srcdir PG_ABS_SRCDIR
CREATE INDEX onek_unique1 ON onek USING btree(unique1 int4_ops);`,
			},
			{
				Statement: `CREATE INDEX IF NOT EXISTS onek_unique1 ON onek USING btree(unique1 int4_ops);`,
			},
			{
				Statement:   `CREATE INDEX IF NOT EXISTS ON onek USING btree(unique1 int4_ops);`,
				ErrorString: `syntax error at or near "ON"`,
			},
			{
				Statement: `CREATE INDEX onek_unique2 ON onek USING btree(unique2 int4_ops);`,
			},
			{
				Statement: `CREATE INDEX onek_hundred ON onek USING btree(hundred int4_ops);`,
			},
			{
				Statement: `CREATE INDEX onek_stringu1 ON onek USING btree(stringu1 name_ops);`,
			},
			{
				Statement: `CREATE INDEX tenk1_unique1 ON tenk1 USING btree(unique1 int4_ops);`,
			},
			{
				Statement: `CREATE INDEX tenk1_unique2 ON tenk1 USING btree(unique2 int4_ops);`,
			},
			{
				Statement: `CREATE INDEX tenk1_hundred ON tenk1 USING btree(hundred int4_ops);`,
			},
			{
				Statement: `CREATE INDEX tenk1_thous_tenthous ON tenk1 (thousand, tenthous);`,
			},
			{
				Statement: `CREATE INDEX tenk2_unique1 ON tenk2 USING btree(unique1 int4_ops);`,
			},
			{
				Statement: `CREATE INDEX tenk2_unique2 ON tenk2 USING btree(unique2 int4_ops);`,
			},
			{
				Statement: `CREATE INDEX tenk2_hundred ON tenk2 USING btree(hundred int4_ops);`,
			},
			{
				Statement: `CREATE INDEX rix ON road USING btree (name text_ops);`,
			},
			{
				Statement: `CREATE INDEX iix ON ihighway USING btree (name text_ops);`,
			},
			{
				Statement: `CREATE INDEX six ON shighway USING btree (name text_ops);`,
			},
			{
				Statement:   `COMMENT ON INDEX six_wrong IS 'bad index';`,
				ErrorString: `relation "six_wrong" does not exist`,
			},
			{
				Statement: `COMMENT ON INDEX six IS 'good index';`,
			},
			{
				Statement: `COMMENT ON INDEX six IS NULL;`,
			},
			{
				Statement: `CREATE INDEX onek2_u1_prtl ON onek2 USING btree(unique1 int4_ops)
	where unique1 < 20 or unique1 > 980;`,
			},
			{
				Statement: `CREATE INDEX onek2_u2_prtl ON onek2 USING btree(unique2 int4_ops)
	where stringu1 < 'B';`,
			},
			{
				Statement: `CREATE INDEX onek2_stu1_prtl ON onek2 USING btree(stringu1 name_ops)
	where onek2.stringu1 >= 'J' and onek2.stringu1 < 'K';`,
			},
			{
				Statement: `CREATE TABLE slow_emp4000 (
	home_base	 box
);`,
			},
			{
				Statement: `CREATE TABLE fast_emp4000 (
	home_base	 box
);`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/rect.data'
COPY slow_emp4000 FROM :'filename';`,
			},
			{
				Statement: `INSERT INTO fast_emp4000 SELECT * FROM slow_emp4000;`,
			},
			{
				Statement: `ANALYZE slow_emp4000;`,
			},
			{
				Statement: `ANALYZE fast_emp4000;`,
			},
			{
				Statement: `CREATE INDEX grect2ind ON fast_emp4000 USING gist (home_base);`,
			},
			{
				Statement: `CREATE TEMP TABLE point_tbl AS SELECT * FROM public.point_tbl;`,
			},
			{
				Statement: `INSERT INTO POINT_TBL(f1) VALUES (NULL);`,
			},
			{
				Statement: `CREATE INDEX gpointind ON point_tbl USING gist (f1);`,
			},
			{
				Statement: `CREATE TEMP TABLE gpolygon_tbl AS
    SELECT polygon(home_base) AS f1 FROM slow_emp4000;`,
			},
			{
				Statement: `INSERT INTO gpolygon_tbl VALUES ( '(1000,0,0,1000)' );`,
			},
			{
				Statement: `INSERT INTO gpolygon_tbl VALUES ( '(0,1000,1000,1000)' );`,
			},
			{
				Statement: `CREATE TEMP TABLE gcircle_tbl AS
    SELECT circle(home_base) AS f1 FROM slow_emp4000;`,
			},
			{
				Statement: `CREATE INDEX ggpolygonind ON gpolygon_tbl USING gist (f1);`,
			},
			{
				Statement: `CREATE INDEX ggcircleind ON gcircle_tbl USING gist (f1);`,
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
				Statement: `SELECT * FROM fast_emp4000
    WHERE home_base <@ '(200,200),(2000,1000)'::box
    ORDER BY (home_base[0])[0];`,
				Results: []sql.Row{{`(337,455),(240,359)`}, {`(1444,403),(1346,344)`}},
			},
			{
				Statement: `SELECT count(*) FROM fast_emp4000 WHERE home_base && '(1000,1000,0,0)'::box;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(*) FROM fast_emp4000 WHERE home_base IS NULL;`,
				Results:   []sql.Row{{278}},
			},
			{
				Statement: `SELECT count(*) FROM gpolygon_tbl WHERE f1 && '(1000,1000,0,0)'::polygon;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(*) FROM gcircle_tbl WHERE f1 && '<(500,500),500>'::circle;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl WHERE f1 <@ box '(0,0,100,100)';`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl WHERE box '(0,0,100,100)' @> f1;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl WHERE f1 <@ polygon '(0,0),(0,100),(100,100),(50,50),(100,0),(0,0)';`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl WHERE f1 <@ circle '<(50,50),50>';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 << '(0.0, 0.0)';`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 >> '(0.0, 0.0)';`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 <<| '(0.0, 0.0)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 |>> '(0.0, 0.0)';`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 ~= '(-5, -12)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM point_tbl ORDER BY f1 <-> '0,1';`,
				Results:   []sql.Row{{`(0,0)`}, {`(1e-300,-1e-300)`}, {`(-3,4)`}, {`(-10,0)`}, {`(10,10)`}, {`(-5,-12)`}, {`(5.1,34.5)`}, {`(Infinity,1e+300)`}, {`(1e+300,Infinity)`}, {`(NaN,NaN)`}, {``}},
			},
			{
				Statement: `SELECT * FROM point_tbl WHERE f1 IS NULL;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT * FROM point_tbl WHERE f1 IS NOT NULL ORDER BY f1 <-> '0,1';`,
				Results:   []sql.Row{{`(0,0)`}, {`(1e-300,-1e-300)`}, {`(-3,4)`}, {`(-10,0)`}, {`(10,10)`}, {`(-5,-12)`}, {`(5.1,34.5)`}, {`(1e+300,Infinity)`}, {`(Infinity,1e+300)`}, {`(NaN,NaN)`}},
			},
			{
				Statement: `SELECT * FROM point_tbl WHERE f1 <@ '(-10,-10),(10,10)':: box ORDER BY f1 <-> '0,1';`,
				Results:   []sql.Row{{`(0,0)`}, {`(1e-300,-1e-300)`}, {`(-3,4)`}, {`(-10,0)`}, {`(10,10)`}},
			},
			{
				Statement: `SELECT * FROM gpolygon_tbl ORDER BY f1 <-> '(0,0)'::point LIMIT 10;`,
				Results:   []sql.Row{{`((240,359),(240,455),(337,455),(337,359))`}, {`((662,163),(662,187),(759,187),(759,163))`}, {`((1000,0),(0,1000))`}, {`((0,1000),(1000,1000))`}, {`((1346,344),(1346,403),(1444,403),(1444,344))`}, {`((278,1409),(278,1457),(369,1457),(369,1409))`}, {`((907,1156),(907,1201),(948,1201),(948,1156))`}, {`((1517,971),(1517,1043),(1594,1043),(1594,971))`}, {`((175,1820),(175,1850),(259,1850),(259,1820))`}, {`((2424,81),(2424,160),(2424,160),(2424,81))`}},
			},
			{
				Statement: `SELECT circle_center(f1), round(radius(f1)) as radius FROM gcircle_tbl ORDER BY f1 <-> '(200,300)'::point LIMIT 10;`,
				Results:   []sql.Row{{`(288.5,407)`, 68}, {`(710.5,175)`, 50}, {`(323.5,1433)`, 51}, {`(927.5,1178.5)`, 30}, {`(1395,373.5)`, 57}, {`(1555.5,1007)`, 53}, {`(217,1835)`, 45}, {`(489,2421.5)`, 22}, {`(2424,120.5)`, 40}, {`(751.5,2655)`, 20}},
			},
			{
				Statement: `SET enable_seqscan = OFF;`,
			},
			{
				Statement: `SET enable_indexscan = ON;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM fast_emp4000
    WHERE home_base <@ '(200,200),(2000,1000)'::box
    ORDER BY (home_base[0])[0];`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: ((home_base[0])[0])`}, {`->  Index Only Scan using grect2ind on fast_emp4000`}, {`Index Cond: (home_base <@ '(2000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT * FROM fast_emp4000
    WHERE home_base <@ '(200,200),(2000,1000)'::box
    ORDER BY (home_base[0])[0];`,
				Results: []sql.Row{{`(337,455),(240,359)`}, {`(1444,403),(1346,344)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM fast_emp4000 WHERE home_base && '(1000,1000,0,0)'::box;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using grect2ind on fast_emp4000`}, {`Index Cond: (home_base && '(1000,1000),(0,0)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM fast_emp4000 WHERE home_base && '(1000,1000,0,0)'::box;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM fast_emp4000 WHERE home_base IS NULL;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using grect2ind on fast_emp4000`}, {`Index Cond: (home_base IS NULL)`}},
			},
			{
				Statement: `SELECT count(*) FROM fast_emp4000 WHERE home_base IS NULL;`,
				Results:   []sql.Row{{278}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM gpolygon_tbl WHERE f1 && '(1000,1000,0,0)'::polygon;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Scan using ggpolygonind on gpolygon_tbl`}, {`Index Cond: (f1 && '((1000,1000),(0,0))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM gpolygon_tbl WHERE f1 && '(1000,1000,0,0)'::polygon;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM gcircle_tbl WHERE f1 && '<(500,500),500>'::circle;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Scan using ggcircleind on gcircle_tbl`}, {`Index Cond: (f1 && '<(500,500),500>'::circle)`}},
			},
			{
				Statement: `SELECT count(*) FROM gcircle_tbl WHERE f1 && '<(500,500),500>'::circle;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl WHERE f1 <@ box '(0,0,100,100)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl`}, {`Index Cond: (f1 <@ '(100,100),(0,0)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl WHERE f1 <@ box '(0,0,100,100)';`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl WHERE box '(0,0,100,100)' @> f1;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl`}, {`Index Cond: (f1 <@ '(100,100),(0,0)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl WHERE box '(0,0,100,100)' @> f1;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl WHERE f1 <@ polygon '(0,0),(0,100),(100,100),(50,50),(100,0),(0,0)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl`}, {`Index Cond: (f1 <@ '((0,0),(0,100),(100,100),(50,50),(100,0),(0,0))'::polygon)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl WHERE f1 <@ polygon '(0,0),(0,100),(100,100),(50,50),(100,0),(0,0)';`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl WHERE f1 <@ circle '<(50,50),50>';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl`}, {`Index Cond: (f1 <@ '<(50,50),50>'::circle)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl WHERE f1 <@ circle '<(50,50),50>';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl p WHERE p.f1 << '(0.0, 0.0)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl p`}, {`Index Cond: (f1 << '(0,0)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 << '(0.0, 0.0)';`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl p WHERE p.f1 >> '(0.0, 0.0)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl p`}, {`Index Cond: (f1 >> '(0,0)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 >> '(0.0, 0.0)';`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl p WHERE p.f1 <<| '(0.0, 0.0)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl p`}, {`Index Cond: (f1 <<| '(0,0)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 <<| '(0.0, 0.0)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl p WHERE p.f1 |>> '(0.0, 0.0)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl p`}, {`Index Cond: (f1 |>> '(0,0)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 |>> '(0.0, 0.0)';`,
				Results:   []sql.Row{{5}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM point_tbl p WHERE p.f1 ~= '(-5, -12)';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using gpointind on point_tbl p`}, {`Index Cond: (f1 ~= '(-5,-12)'::point)`}},
			},
			{
				Statement: `SELECT count(*) FROM point_tbl p WHERE p.f1 ~= '(-5, -12)';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM point_tbl ORDER BY f1 <-> '0,1';`,
				Results: []sql.Row{{`Index Only Scan using gpointind on point_tbl`}, {`Order By: (f1 <-> '(0,1)'::point)`}},
			},
			{
				Statement: `SELECT * FROM point_tbl ORDER BY f1 <-> '0,1';`,
				Results:   []sql.Row{{`(1e-300,-1e-300)`}, {`(0,0)`}, {`(-3,4)`}, {`(-10,0)`}, {`(10,10)`}, {`(-5,-12)`}, {`(5.1,34.5)`}, {`(Infinity,1e+300)`}, {`(1e+300,Infinity)`}, {`(NaN,NaN)`}, {``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM point_tbl WHERE f1 IS NULL;`,
				Results: []sql.Row{{`Index Only Scan using gpointind on point_tbl`}, {`Index Cond: (f1 IS NULL)`}},
			},
			{
				Statement: `SELECT * FROM point_tbl WHERE f1 IS NULL;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM point_tbl WHERE f1 IS NOT NULL ORDER BY f1 <-> '0,1';`,
				Results: []sql.Row{{`Index Only Scan using gpointind on point_tbl`}, {`Index Cond: (f1 IS NOT NULL)`}, {`Order By: (f1 <-> '(0,1)'::point)`}},
			},
			{
				Statement: `SELECT * FROM point_tbl WHERE f1 IS NOT NULL ORDER BY f1 <-> '0,1';`,
				Results:   []sql.Row{{`(1e-300,-1e-300)`}, {`(0,0)`}, {`(-3,4)`}, {`(-10,0)`}, {`(10,10)`}, {`(-5,-12)`}, {`(5.1,34.5)`}, {`(Infinity,1e+300)`}, {`(1e+300,Infinity)`}, {`(NaN,NaN)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM point_tbl WHERE f1 <@ '(-10,-10),(10,10)':: box ORDER BY f1 <-> '0,1';`,
				Results: []sql.Row{{`Index Only Scan using gpointind on point_tbl`}, {`Index Cond: (f1 <@ '(10,10),(-10,-10)'::box)`}, {`Order By: (f1 <-> '(0,1)'::point)`}},
			},
			{
				Statement: `SELECT * FROM point_tbl WHERE f1 <@ '(-10,-10),(10,10)':: box ORDER BY f1 <-> '0,1';`,
				Results:   []sql.Row{{`(1e-300,-1e-300)`}, {`(0,0)`}, {`(-3,4)`}, {`(-10,0)`}, {`(10,10)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM gpolygon_tbl ORDER BY f1 <-> '(0,0)'::point LIMIT 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using ggpolygonind on gpolygon_tbl`}, {`Order By: (f1 <-> '(0,0)'::point)`}},
			},
			{
				Statement: `SELECT * FROM gpolygon_tbl ORDER BY f1 <-> '(0,0)'::point LIMIT 10;`,
				Results:   []sql.Row{{`((240,359),(240,455),(337,455),(337,359))`}, {`((662,163),(662,187),(759,187),(759,163))`}, {`((1000,0),(0,1000))`}, {`((0,1000),(1000,1000))`}, {`((1346,344),(1346,403),(1444,403),(1444,344))`}, {`((278,1409),(278,1457),(369,1457),(369,1409))`}, {`((907,1156),(907,1201),(948,1201),(948,1156))`}, {`((1517,971),(1517,1043),(1594,1043),(1594,971))`}, {`((175,1820),(175,1850),(259,1850),(259,1820))`}, {`((2424,81),(2424,160),(2424,160),(2424,81))`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT circle_center(f1), round(radius(f1)) as radius FROM gcircle_tbl ORDER BY f1 <-> '(200,300)'::point LIMIT 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using ggcircleind on gcircle_tbl`}, {`Order By: (f1 <-> '(200,300)'::point)`}},
			},
			{
				Statement: `SELECT circle_center(f1), round(radius(f1)) as radius FROM gcircle_tbl ORDER BY f1 <-> '(200,300)'::point LIMIT 10;`,
				Results:   []sql.Row{{`(288.5,407)`, 68}, {`(710.5,175)`, 50}, {`(323.5,1433)`, 51}, {`(927.5,1178.5)`, 30}, {`(1395,373.5)`, 57}, {`(1555.5,1007)`, 53}, {`(217,1835)`, 45}, {`(489,2421.5)`, 22}, {`(2424,120.5)`, 40}, {`(751.5,2655)`, 20}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT point(x,x), (SELECT f1 FROM gpolygon_tbl ORDER BY f1 <-> point(x,x) LIMIT 1) as c FROM generate_series(0,10,1) x;`,
				Results: []sql.Row{{`Function Scan on generate_series x`}, {`SubPlan 1`}, {`->  Limit`}, {`->  Index Scan using ggpolygonind on gpolygon_tbl`}, {`Order By: (f1 <-> point((x.x)::double precision, (x.x)::double precision))`}},
			},
			{
				Statement: `SELECT point(x,x), (SELECT f1 FROM gpolygon_tbl ORDER BY f1 <-> point(x,x) LIMIT 1) as c FROM generate_series(0,10,1) x;`,
				Results:   []sql.Row{{`(0,0)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(1,1)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(2,2)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(3,3)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(4,4)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(5,5)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(6,6)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(7,7)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(8,8)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(9,9)`, `((240,359),(240,455),(337,455),(337,359))`}, {`(10,10)`, `((240,359),(240,455),(337,455),(337,359))`}},
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
SELECT * FROM point_tbl WHERE f1 <@ '(-10,-10),(10,10)':: box ORDER BY f1 <-> '0,1';`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: ((f1 <-> '(0,1)'::point))`}, {`->  Bitmap Heap Scan on point_tbl`}, {`Recheck Cond: (f1 <@ '(10,10),(-10,-10)'::box)`}, {`->  Bitmap Index Scan on gpointind`}, {`Index Cond: (f1 <@ '(10,10),(-10,-10)'::box)`}},
			},
			{
				Statement: `SELECT * FROM point_tbl WHERE f1 <@ '(-10,-10),(10,10)':: box ORDER BY f1 <-> '0,1';`,
				Results:   []sql.Row{{`(0,0)`}, {`(1e-300,-1e-300)`}, {`(-3,4)`}, {`(-10,0)`}, {`(10,10)`}},
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
			{
				Statement: `CREATE TABLE array_index_op_test (
	seqno		int4,
	i			int4[],
	t			text[]
);`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/array.data'
COPY array_index_op_test FROM :'filename';`,
			},
			{
				Statement: `ANALYZE array_index_op_test;`,
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i = '{NULL}' ORDER BY seqno;`,
				Results:   []sql.Row{{102, `{NULL}`, `{NULL}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i @> '{NULL}' ORDER BY seqno;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i && '{NULL}' ORDER BY seqno;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i <@ '{NULL}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
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
				Statement: `CREATE INDEX intarrayidx ON array_index_op_test USING gin (i);`,
			},
			{
				Statement: `explain (costs off)
SELECT * FROM array_index_op_test WHERE i @> '{32}' ORDER BY seqno;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: seqno`}, {`->  Bitmap Heap Scan on array_index_op_test`}, {`Recheck Cond: (i @> '{32}'::integer[])`}, {`->  Bitmap Index Scan on intarrayidx`}, {`Index Cond: (i @> '{32}'::integer[])`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i @> '{32}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i && '{32}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i @> '{17}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i && '{17}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i @> '{32,17}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i && '{32,17}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i <@ '{38,34,32,89}' ORDER BY seqno;`,
				Results:   []sql.Row{{40, `{34}`, `{AAAAAAAAAAAAAA10611,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAA31334,AAAAA70466,AAAAAAAA81587,AAAAAAA74623}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i = '{47,77}' ORDER BY seqno;`,
				Results:   []sql.Row{{95, `{47,77}`, `{AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA74076,AAAAAAAAAA18107,AAAAA40681,AAAAAAAAAAAAAAA35875,AAAAA60038,AAAAAAA56483}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i = '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i @> '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{1, `{92,75,71,52,64,83}`, `{AAAAAAAA44066,AAAAAA1059,AAAAAAAAAAA176,AAAAAAA48038}`}, {2, `{3,6}`, `{AAAAAA98232,AAAAAAAA79710,AAAAAAAAAAAAAAAAA69675,AAAAAAAAAAAAAAAA55798,AAAAAAAAA12793}`}, {3, `{37,64,95,43,3,41,13,30,11,43}`, `{AAAAAAAAAA48845,AAAAA75968,AAAAA95309,AAA54451,AAAAAAAAAA22292,AAAAAAA99836,A96617,AA17009,AAAAAAAAAAAAAA95246}`}, {4, `{71,39,99,55,33,75,45}`, `{AAAAAAAAA53663,AAAAAAAAAAAAAAA67062,AAAAAAAAAA64777,AAA99043,AAAAAAAAAAAAAAAAAAA91804,39557}`}, {5, `{50,42,77,50,4}`, `{AAAAAAAAAAAAAAAAA26540,AAAAAAA79710,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA176,AAAAA95309,AAAAAAAAAAA46154,AAAAAA66777,AAAAAAAAA27249,AAAAAAAAAA64777,AAAAAAAAAAAAAAAAAAA70104}`}, {6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {7, `{12,51,88,64,8}`, `{AAAAAAAAAAAAAAAAAA12591,AAAAAAAAAAAAAAAAA50407,AAAAAAAAAAAA67946}`}, {8, `{60,84}`, `{AAAAAAA81898,AAAAAA1059,AAAAAAAAAAAA81511,AAAAA961,AAAAAAAAAAAAAAAA31334,AAAAA64741,AA6416,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAAAA50407}`}, {9, `{56,52,35,27,80,44,81,22}`, `{AAAAAAAAAAAAAAA73034,AAAAAAAAAAAAA7929,AAAAAAA66161,AA88409,39557,A27153,AAAAAAAA9523,AAAAAAAAAAA99000}`}, {10, `{71,5,45}`, `{AAAAAAAAAAA21658,AAAAAAAAAAAA21089,AAA54451,AAAAAAAAAAAAAAAAAA54141,AAAAAAAAAAAAAA28620,AAAAAAAAAAA21658,AAAAAAAAAAA74076,AAAAAAAAA27249}`}, {11, `{41,86,74,48,22,74,47,50}`, `{AAAAAAAA9523,AAAAAAAAAAAA37562,AAAAAAAAAAAAAAAA14047,AAAAAAAAAAA46154,AAAA41702,AAAAAAAAAAAAAAAAA764,AAAAA62737,39557}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {13, `{3,52,34,23}`, `{AAAAAA98232,AAAA49534,AAAAAAAAAAA21658}`}, {14, `{78,57,19}`, `{AAAA8857,AAAAAAAAAAAAAAA73034,AAAAAAAA81587,AAAAAAAAAAAAAAA68526,AAAAA75968,AAAAAAAAAAAAAA65909,AAAAAAAAA10012,AAAAAAAAAAAAAA65909}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {16, `{14,63,85,11}`, `{AAAAAA66777}`}, {17, `{7,10,81,85}`, `{AAAAAA43678,AAAAAAA12144,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAAAAA15356}`}, {18, `{1}`, `{AAAAAAAAAAA33576,AAAAA95309,64261,AAA59323,AAAAAAAAAAAAAA95246,55847,AAAAAAAAAAAA67946,AAAAAAAAAAAAAAAAAA64374}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {20, `{72,89,70,51,54,37,8,49,79}`, `{AAAAAA58494}`}, {21, `{2,8,65,10,5,79,43}`, `{AAAAAAAAAAAAAAAAA88852,AAAAAAAAAAAAAAAAAAA91804,AAAAA64669,AAAAAAAAAAAAAAAA1443,AAAAAAAAAAAAAAAA23657,AAAAA12179,AAAAAAAAAAAAAAAAA88852,AAAAAAAAAAAAAAAA31334,AAAAAAAAAAAAAAAA41303,AAAAAAAAAAAAAAAAAAA85420}`}, {22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {23, `{40,90,5,38,72,40,30,10,43,55}`, `{A6053,AAAAAAAAAAA6119,AA44673,AAAAAAAAAAAAAAAAA764,AA17009,AAAAA17383,AAAAA70514,AAAAA33250,AAAAA95309,AAAAAAAAAAAA37562}`}, {24, `{94,61,99,35,48}`, `{AAAAAAAAAAA50956,AAAAAAAAAAA15165,AAAA85070,AAAAAAAAAAAAAAA36627,AAAAA961,AAAAAAAAAA55219}`}, {25, `{31,1,10,11,27,79,38}`, `{AAAAAAAAAAAAAAAAAA59334,45449}`}, {26, `{71,10,9,69,75}`, `{47735,AAAAAAA21462,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAAAAA91804,AAAAAAAAA72121,AAAAAAAAAAAAAAAAAAA1205,AAAAA41597,AAAA8857,AAAAAAAAAAAAAAAAAAA15356,AA17009}`}, {27, `{94}`, `{AA6416,A6053,AAAAAAA21462,AAAAAAA57334,AAAAAAAAAAAAAAAAAA12591,AA88409,AAAAAAAAAAAAA70254}`}, {28, `{14,33,6,34,14}`, `{AAAAAAAAAAAAAAA13198,AAAAAAAA69452,AAAAAAAAAAA82945,AAAAAAA12144,AAAAAAAAA72121,AAAAAAAAAA18601}`}, {29, `{39,21}`, `{AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAAAAA38885,AAAA85070,AAAAAAAAAAAAAAAAAAA70104,AAAAA66674,AAAAAAAAAAAAA62007,AAAAAAAA69452,AAAAAAA1242,AAAAAAAAAAAAAAAA1729,AAAA35194}`}, {30, `{26,81,47,91,34}`, `{AAAAAAAAAAAAAAAAAAA70104,AAAAAAA80240}`}, {31, `{80,24,18,21,54}`, `{AAAAAAAAAAAAAAA13198,AAAAAAAAAAAAAAAAAAA70415,A27153,AAAAAAAAA53663,AAAAAAAAAAAAAAAAA50407,A68938}`}, {32, `{58,79,82,80,67,75,98,10,41}`, `{AAAAAAAAAAAAAAAAAA61286,AAA54451,AAAAAAAAAAAAAAAAAAA87527,A96617,51533}`}, {33, `{74,73}`, `{A85417,AAAAAAA56483,AAAAA17383,AAAAAAAAAAAAA62159,AAAAAAAAAAAA52814,AAAAAAAAAAAAA85723,AAAAAAAAAAAAAAAAAA55796}`}, {34, `{70,45}`, `{AAAAAAAAAAAAAAAAAA71621,AAAAAAAAAAAAAA28620,AAAAAAAAAA55219,AAAAAAAA23648,AAAAAAAAAA22292,AAAAAAA1242}`}, {35, `{23,40}`, `{AAAAAAAAAAAA52814,AAAA48949,AAAAAAAAA34727,AAAA8857,AAAAAAAAAAAAAAAAAAA62179,AAAAAAAAAAAAAAA68526,AAAAAAA99836,AAAAAAAA50094,AAAA91194,AAAAAAAAAAAAA73084}`}, {36, `{79,82,14,52,30,5,79}`, `{AAAAAAAAA53663,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA89194,AA88409,AAAAAAAAAAAAAAA81326,AAAAAAAAAAAAAAAAA63050,AAAAAAAAAAAAAAAA33598}`}, {37, `{53,11,81,39,3,78,58,64,74}`, `{AAAAAAAAAAAAAAAAAAA17075,AAAAAAA66161,AAAAAAAA23648,AAAAAAAAAAAAAA10611}`}, {38, `{59,5,4,95,28}`, `{AAAAAAAAAAA82945,A96617,47735,AAAAA12179,AAAAA64669,AAAAAA99807,AA74433,AAAAAAAAAAAAAAAAA59387}`}, {39, `{82,43,99,16,74}`, `{AAAAAAAAAAAAAAA67062,AAAAAAA57334,AAAAAAAAAAAAAA65909,A27153,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAA64777,AAAAAAAAAAAA81511,AAAAAAAAAAAAAA65909,AAAAAAAAAAAAAA28620}`}, {40, `{34}`, `{AAAAAAAAAAAAAA10611,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAA31334,AAAAA70466,AAAAAAAA81587,AAAAAAA74623}`}, {41, `{19,26,63,12,93,73,27,94}`, `{AAAAAAA79710,AAAAAAAAAA55219,AAAA41702,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA71621,AAAAAAAAAAAAAAAAA63050,AAAAAAA99836,AAAAAAAAAAAAAA8666}`}, {42, `{15,76,82,75,8,91}`, `{AAAAAAAAAAA176,AAAAAA38063,45449,AAAAAA54032,AAAAAAA81898,AA6416,AAAAAAAAAAAAAAAAAAA62179,45449,AAAAA60038,AAAAAAAA81587}`}, {43, `{39,87,91,97,79,28}`, `{AAAAAAAAAAA74076,A96617,AAAAAAAAAAAAAAAAAAA89194,AAAAAAAAAAAAAAAAAA55796,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAA67946}`}, {44, `{40,58,68,29,54}`, `{AAAAAAA81898,AAAAAA66777,AAAAAA98232}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {46, `{53,24}`, `{AAAAAAAAAAA53908,AAAAAA54032,AAAAA17383,AAAA48949,AAAAAAAAAA18601,AAAAA64669,45449,AAAAAAAAAAA98051,AAAAAAAAAAAAAAAAAA71621}`}, {47, `{98,23,64,12,75,61}`, `{AAA59323,AAAAA95309,AAAAAAAAAAAAAAAA31334,AAAAAAAAA27249,AAAAA17383,AAAAAAAAAAAA37562,AAAAAA1059,A84822,55847,AAAAA70466}`}, {48, `{76,14}`, `{AAAAAAAAAAAAA59671,AAAAAAAAAAAAAAAAAAA91804,AAAAAA66777,AAAAAAAAAAAAAAAAAAA89194,AAAAAAAAAAAAAAA36627,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAA73084,AAAAAAA79710,AAAAAAAAAAAAAAA40402,AAAAAAAAAAAAAAAAAAA65037}`}, {49, `{56,5,54,37,49}`, `{AA21643,AAAAAAAAAAA92631,AAAAAAAA81587}`}, {50, `{20,12,37,64,93}`, `{AAAAAAAAAA5483,AAAAAAAAAAAAAAAAAAA1205,AA6416,AAAAAAAAAAAAAAAAA63050,AAAAAAAAAAAAAAAAAA47955}`}, {51, `{47}`, `{AAAAAAAAAAAAAA96505,AAAAAAAAAAAAAAAAAA36842,AAAAA95309,AAAAAAAA81587,AA6416,AAAA91194,AAAAAA58494,AAAAAA1059,AAAAAAAA69452}`}, {52, `{89,0}`, `{AAAAAAAAAAAAAAAAAA47955,AAAAAAA48038,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAAAAA73084,AAAAA70466,AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA46154,AA66862}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {54, `{70,47}`, `{AAAAAAAAAAAAAAAAAA54141,AAAAA40681,AAAAAAA48038,AAAAAAAAAAAAAAAA29150,AAAAA41597,AAAAAAAAAAAAAAAAAA59334,AA15322}`}, {55, `{47,79,47,64,72,25,71,24,93}`, `{AAAAAAAAAAAAAAAAAA55796,AAAAA62737}`}, {56, `{33,7,60,54,93,90,77,85,39}`, `{AAAAAAAAAAAAAAAAAA32918,AA42406}`}, {57, `{23,45,10,42,36,21,9,96}`, `{AAAAAAAAAAAAAAAAAAA70415}`}, {58, `{92}`, `{AAAAAAAAAAAAAAAA98414,AAAAAAAA23648,AAAAAAAAAAAAAAAAAA55796,AA25381,AAAAAAAAAAA6119}`}, {59, `{9,69,46,77}`, `{39557,AAAAAAA89932,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAAAAAAAAA26540,AAA20874,AA6416,AAAAAAAAAAAAAAAAAA47955}`}, {60, `{62,2,59,38,89}`, `{AAAAAAA89932,AAAAAAAAAAAAAAAAAAA15356,AA99927,AA17009,AAAAAAAAAAAAAAA35875}`}, {61, `{72,2,44,95,54,54,13}`, `{AAAAAAAAAAAAAAAAAAA91804}`}, {62, `{83,72,29,73}`, `{AAAAAAAAAAAAA15097,AAAA8857,AAAAAAAAAAAA35809,AAAAAAAAAAAA52814,AAAAAAAAAAAAAAAAAAA38885,AAAAAAAAAAAAAAAAAA24183,AAAAAA43678,A96617}`}, {63, `{11,4,61,87}`, `{AAAAAAAAA27249,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAA13198,AAA20874,39557,51533,AAAAAAAAAAA53908,AAAAAAAAAAAAAA96505,AAAAAAAA78938}`}, {64, `{26,19,34,24,81,78}`, `{A96617,AAAAAAAAAAAAAAAAAAA70104,A68938,AAAAAAAAAAA53908,AAAAAAAAAAAAAAA453,AA17009,AAAAAAA80240}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {66, `{31,23,70,52,4,33,48,25}`, `{AAAAAAAAAAAAAAAAA69675,AAAAAAAA50094,AAAAAAAAAAA92631,AAAA35194,39557,AAAAAAA99836}`}, {67, `{31,94,7,10}`, `{AAAAAA38063,A96617,AAAA35194,AAAAAAAAAAAA67946}`}, {68, `{90,43,38}`, `{AA75092,AAAAAAAAAAAAAAAAA69675,AAAAAAAAAAA92631,AAAAAAAAA10012,AAAAAAAAAAAAA7929,AA21643}`}, {69, `{67,35,99,85,72,86,44}`, `{AAAAAAAAAAAAAAAAAAA1205,AAAAAAAA50094,AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAAAAAAA47955}`}, {70, `{56,70,83}`, `{AAAA41702,AAAAAAAAAAA82945,AA21643,AAAAAAAAAAA99000,A27153,AA25381,AAAAAAAAAAAAAA96505,AAAAAAA1242}`}, {71, `{74,26}`, `{AAAAAAAAAAA50956,AA74433,AAAAAAA21462,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAA36627,AAAAAAAAAAAAA70254,AAAAAAAAAA43419,39557}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {73, `{88,25,96,78,65,15,29,19}`, `{AAA54451,AAAAAAAAA27249,AAAAAAA9228,AAAAAAAAAAAAAAA67062,AAAAAAAAAAAAAAAAAAA70415,AAAAA17383,AAAAAAAAAAAAAAAA33598}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {75, `{12,96,83,24,71,89,55}`, `{AAAA48949,AAAAAAAA29716,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAAA67946,AAAAAAAAAAAAAAAA29150,AAA28075,AAAAAAAAAAAAAAAAA43052}`}, {76, `{92,55,10,7}`, `{AAAAAAAAAAAAAAA67062}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {78, `{55,89,44,84,34}`, `{AAAAAAAAAAA6119,AAAAAAAAAAAAAA8666,AA99927,AA42406,AAAAAAA81898,AAAAAAA9228,AAAAAAAAAAA92631,AA21643,AAAAAAAAAAAAAA28620}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {80, `{74,89,44,80,0}`, `{AAAA35194,AAAAAAAA79710,AAA20874,AAAAAAAAAAAAAAAAAAA70104,AAAAAAAAAAAAA73084,AAAAAAA57334,AAAAAAA9228,AAAAAAAAAAAAA62007}`}, {81, `{63,77,54,48,61,53,97}`, `{AAAAAAAAAAAAAAA81326,AAAAAAAAAA22292,AA25381,AAAAAAAAAAA74076,AAAAAAA81898,AAAAAAAAA72121}`}, {82, `{34,60,4,79,78,16,86,89,42,50}`, `{AAAAA40681,AAAAAAAAAAAAAAAAAA12591,AAAAAAA80240,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA70104}`}, {83, `{14,10}`, `{AAAAAAAAAA22292,AAAAAAAAAAAAA70254,AAAAAAAAAAA6119}`}, {84, `{11,83,35,13,96,94}`, `{AAAAA95309,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAAAAA24183}`}, {85, `{39,60}`, `{AAAAAAAAAAAAAAAA55798,AAAAAAAAAA22292,AAAAAAA66161,AAAAAAA21462,AAAAAAAAAAAAAAAAAA12591,55847,AAAAAA98232,AAAAAAAAAAA46154}`}, {86, `{33,81,72,74,45,36,82}`, `{AAAAAAAA81587,AAAAAAAAAAAAAA96505,45449,AAAA80176}`}, {87, `{57,27,50,12,97,68}`, `{AAAAAAAAAAAAAAAAA26540,AAAAAAAAA10012,AAAAAAAAAAAA35809,AAAAAAAAAAAAAAAA29150,AAAAAAAAAAA82945,AAAAAA66777,31228,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAA96505}`}, {88, `{41,90,77,24,6,24}`, `{AAAA35194,AAAA35194,AAAAAAA80240,AAAAAAAAAAA46154,AAAAAA58494,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA59334,AAAAAAAAAAAAAAAAAAA91804,AA74433}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {90, `{88,75}`, `{AAAAA60038,AAAAAAAA23648,AAAAAAAAAAA99000,AAAA41702,AAAAAAAAAAAAA22860,AAAAAAAAAAAAAAA68526}`}, {91, `{78}`, `{AAAAAAAAAAAAA62007,AAA99043}`}, {92, `{85,63,49,45}`, `{AAAAAAA89932,AAAAAAAAAAAAA22860,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAAA21089}`}, {93, `{11}`, `{AAAAAAAAAAA176,AAAAAAAAAAAAAA8666,AAAAAAAAAAAAAAA453,AAAAAAAAAAAAA85723,A68938,AAAAAAAAAAAAA9821,AAAAAAA48038,AAAAAAAAAAAAAAAAA59387,AA99927,AAAAA17383}`}, {94, `{98,9,85,62,88,91,60,61,38,86}`, `{AAAAAAAA81587,AAAAA17383,AAAAAAAA81587}`}, {95, `{47,77}`, `{AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA74076,AAAAAAAAAA18107,AAAAA40681,AAAAAAAAAAAAAAA35875,AAAAA60038,AAAAAAA56483}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}, {97, `{54,2,86,65}`, `{47735,AAAAAAA99836,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAA29150,AAAAAAA80240,AAAAAAAAAAAAAAAA98414,AAAAAAA56483,AAAAAAAAAAAAAAAA29150,AAAAAAA39692,AA21643}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {99, `{37,86}`, `{AAAAAAAAAAAAAAAAAA32918,AAAAA70514,AAAAAAAAA10012,AAAAAAAAAAAAAAAAA59387,AAAAAAAAAA64777,AAAAAAAAAAAAAAAAAAA15356}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}, {101, `{}`, `{}`}, {102, `{NULL}`, `{NULL}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i && '{}' ORDER BY seqno;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i <@ '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `CREATE INDEX textarrayidx ON array_index_op_test USING gin (t);`,
			},
			{
				Statement: `explain (costs off)
SELECT * FROM array_index_op_test WHERE t @> '{AAAAAAAA72908}' ORDER BY seqno;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: seqno`}, {`->  Bitmap Heap Scan on array_index_op_test`}, {`Recheck Cond: (t @> '{AAAAAAAA72908}'::text[])`}, {`->  Bitmap Index Scan on textarrayidx`}, {`Index Cond: (t @> '{AAAAAAAA72908}'::text[])`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t @> '{AAAAAAAA72908}' ORDER BY seqno;`,
				Results:   []sql.Row{{22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t && '{AAAAAAAA72908}' ORDER BY seqno;`,
				Results:   []sql.Row{{22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t @> '{AAAAAAAAAA646}' ORDER BY seqno;`,
				Results:   []sql.Row{{15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t && '{AAAAAAAAAA646}' ORDER BY seqno;`,
				Results:   []sql.Row{{15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t @> '{AAAAAAAA72908,AAAAAAAAAA646}' ORDER BY seqno;`,
				Results:   []sql.Row{{79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t && '{AAAAAAAA72908,AAAAAAAAAA646}' ORDER BY seqno;`,
				Results:   []sql.Row{{15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t <@ '{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}' ORDER BY seqno;`,
				Results:   []sql.Row{{22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t = '{AAAAAAAAAA646,A87088}' ORDER BY seqno;`,
				Results:   []sql.Row{{96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t = '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t @> '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{1, `{92,75,71,52,64,83}`, `{AAAAAAAA44066,AAAAAA1059,AAAAAAAAAAA176,AAAAAAA48038}`}, {2, `{3,6}`, `{AAAAAA98232,AAAAAAAA79710,AAAAAAAAAAAAAAAAA69675,AAAAAAAAAAAAAAAA55798,AAAAAAAAA12793}`}, {3, `{37,64,95,43,3,41,13,30,11,43}`, `{AAAAAAAAAA48845,AAAAA75968,AAAAA95309,AAA54451,AAAAAAAAAA22292,AAAAAAA99836,A96617,AA17009,AAAAAAAAAAAAAA95246}`}, {4, `{71,39,99,55,33,75,45}`, `{AAAAAAAAA53663,AAAAAAAAAAAAAAA67062,AAAAAAAAAA64777,AAA99043,AAAAAAAAAAAAAAAAAAA91804,39557}`}, {5, `{50,42,77,50,4}`, `{AAAAAAAAAAAAAAAAA26540,AAAAAAA79710,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA176,AAAAA95309,AAAAAAAAAAA46154,AAAAAA66777,AAAAAAAAA27249,AAAAAAAAAA64777,AAAAAAAAAAAAAAAAAAA70104}`}, {6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {7, `{12,51,88,64,8}`, `{AAAAAAAAAAAAAAAAAA12591,AAAAAAAAAAAAAAAAA50407,AAAAAAAAAAAA67946}`}, {8, `{60,84}`, `{AAAAAAA81898,AAAAAA1059,AAAAAAAAAAAA81511,AAAAA961,AAAAAAAAAAAAAAAA31334,AAAAA64741,AA6416,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAAAA50407}`}, {9, `{56,52,35,27,80,44,81,22}`, `{AAAAAAAAAAAAAAA73034,AAAAAAAAAAAAA7929,AAAAAAA66161,AA88409,39557,A27153,AAAAAAAA9523,AAAAAAAAAAA99000}`}, {10, `{71,5,45}`, `{AAAAAAAAAAA21658,AAAAAAAAAAAA21089,AAA54451,AAAAAAAAAAAAAAAAAA54141,AAAAAAAAAAAAAA28620,AAAAAAAAAAA21658,AAAAAAAAAAA74076,AAAAAAAAA27249}`}, {11, `{41,86,74,48,22,74,47,50}`, `{AAAAAAAA9523,AAAAAAAAAAAA37562,AAAAAAAAAAAAAAAA14047,AAAAAAAAAAA46154,AAAA41702,AAAAAAAAAAAAAAAAA764,AAAAA62737,39557}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {13, `{3,52,34,23}`, `{AAAAAA98232,AAAA49534,AAAAAAAAAAA21658}`}, {14, `{78,57,19}`, `{AAAA8857,AAAAAAAAAAAAAAA73034,AAAAAAAA81587,AAAAAAAAAAAAAAA68526,AAAAA75968,AAAAAAAAAAAAAA65909,AAAAAAAAA10012,AAAAAAAAAAAAAA65909}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {16, `{14,63,85,11}`, `{AAAAAA66777}`}, {17, `{7,10,81,85}`, `{AAAAAA43678,AAAAAAA12144,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAAAAA15356}`}, {18, `{1}`, `{AAAAAAAAAAA33576,AAAAA95309,64261,AAA59323,AAAAAAAAAAAAAA95246,55847,AAAAAAAAAAAA67946,AAAAAAAAAAAAAAAAAA64374}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {20, `{72,89,70,51,54,37,8,49,79}`, `{AAAAAA58494}`}, {21, `{2,8,65,10,5,79,43}`, `{AAAAAAAAAAAAAAAAA88852,AAAAAAAAAAAAAAAAAAA91804,AAAAA64669,AAAAAAAAAAAAAAAA1443,AAAAAAAAAAAAAAAA23657,AAAAA12179,AAAAAAAAAAAAAAAAA88852,AAAAAAAAAAAAAAAA31334,AAAAAAAAAAAAAAAA41303,AAAAAAAAAAAAAAAAAAA85420}`}, {22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {23, `{40,90,5,38,72,40,30,10,43,55}`, `{A6053,AAAAAAAAAAA6119,AA44673,AAAAAAAAAAAAAAAAA764,AA17009,AAAAA17383,AAAAA70514,AAAAA33250,AAAAA95309,AAAAAAAAAAAA37562}`}, {24, `{94,61,99,35,48}`, `{AAAAAAAAAAA50956,AAAAAAAAAAA15165,AAAA85070,AAAAAAAAAAAAAAA36627,AAAAA961,AAAAAAAAAA55219}`}, {25, `{31,1,10,11,27,79,38}`, `{AAAAAAAAAAAAAAAAAA59334,45449}`}, {26, `{71,10,9,69,75}`, `{47735,AAAAAAA21462,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAAAAA91804,AAAAAAAAA72121,AAAAAAAAAAAAAAAAAAA1205,AAAAA41597,AAAA8857,AAAAAAAAAAAAAAAAAAA15356,AA17009}`}, {27, `{94}`, `{AA6416,A6053,AAAAAAA21462,AAAAAAA57334,AAAAAAAAAAAAAAAAAA12591,AA88409,AAAAAAAAAAAAA70254}`}, {28, `{14,33,6,34,14}`, `{AAAAAAAAAAAAAAA13198,AAAAAAAA69452,AAAAAAAAAAA82945,AAAAAAA12144,AAAAAAAAA72121,AAAAAAAAAA18601}`}, {29, `{39,21}`, `{AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAAAAA38885,AAAA85070,AAAAAAAAAAAAAAAAAAA70104,AAAAA66674,AAAAAAAAAAAAA62007,AAAAAAAA69452,AAAAAAA1242,AAAAAAAAAAAAAAAA1729,AAAA35194}`}, {30, `{26,81,47,91,34}`, `{AAAAAAAAAAAAAAAAAAA70104,AAAAAAA80240}`}, {31, `{80,24,18,21,54}`, `{AAAAAAAAAAAAAAA13198,AAAAAAAAAAAAAAAAAAA70415,A27153,AAAAAAAAA53663,AAAAAAAAAAAAAAAAA50407,A68938}`}, {32, `{58,79,82,80,67,75,98,10,41}`, `{AAAAAAAAAAAAAAAAAA61286,AAA54451,AAAAAAAAAAAAAAAAAAA87527,A96617,51533}`}, {33, `{74,73}`, `{A85417,AAAAAAA56483,AAAAA17383,AAAAAAAAAAAAA62159,AAAAAAAAAAAA52814,AAAAAAAAAAAAA85723,AAAAAAAAAAAAAAAAAA55796}`}, {34, `{70,45}`, `{AAAAAAAAAAAAAAAAAA71621,AAAAAAAAAAAAAA28620,AAAAAAAAAA55219,AAAAAAAA23648,AAAAAAAAAA22292,AAAAAAA1242}`}, {35, `{23,40}`, `{AAAAAAAAAAAA52814,AAAA48949,AAAAAAAAA34727,AAAA8857,AAAAAAAAAAAAAAAAAAA62179,AAAAAAAAAAAAAAA68526,AAAAAAA99836,AAAAAAAA50094,AAAA91194,AAAAAAAAAAAAA73084}`}, {36, `{79,82,14,52,30,5,79}`, `{AAAAAAAAA53663,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA89194,AA88409,AAAAAAAAAAAAAAA81326,AAAAAAAAAAAAAAAAA63050,AAAAAAAAAAAAAAAA33598}`}, {37, `{53,11,81,39,3,78,58,64,74}`, `{AAAAAAAAAAAAAAAAAAA17075,AAAAAAA66161,AAAAAAAA23648,AAAAAAAAAAAAAA10611}`}, {38, `{59,5,4,95,28}`, `{AAAAAAAAAAA82945,A96617,47735,AAAAA12179,AAAAA64669,AAAAAA99807,AA74433,AAAAAAAAAAAAAAAAA59387}`}, {39, `{82,43,99,16,74}`, `{AAAAAAAAAAAAAAA67062,AAAAAAA57334,AAAAAAAAAAAAAA65909,A27153,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAA64777,AAAAAAAAAAAA81511,AAAAAAAAAAAAAA65909,AAAAAAAAAAAAAA28620}`}, {40, `{34}`, `{AAAAAAAAAAAAAA10611,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAA31334,AAAAA70466,AAAAAAAA81587,AAAAAAA74623}`}, {41, `{19,26,63,12,93,73,27,94}`, `{AAAAAAA79710,AAAAAAAAAA55219,AAAA41702,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA71621,AAAAAAAAAAAAAAAAA63050,AAAAAAA99836,AAAAAAAAAAAAAA8666}`}, {42, `{15,76,82,75,8,91}`, `{AAAAAAAAAAA176,AAAAAA38063,45449,AAAAAA54032,AAAAAAA81898,AA6416,AAAAAAAAAAAAAAAAAAA62179,45449,AAAAA60038,AAAAAAAA81587}`}, {43, `{39,87,91,97,79,28}`, `{AAAAAAAAAAA74076,A96617,AAAAAAAAAAAAAAAAAAA89194,AAAAAAAAAAAAAAAAAA55796,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAA67946}`}, {44, `{40,58,68,29,54}`, `{AAAAAAA81898,AAAAAA66777,AAAAAA98232}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {46, `{53,24}`, `{AAAAAAAAAAA53908,AAAAAA54032,AAAAA17383,AAAA48949,AAAAAAAAAA18601,AAAAA64669,45449,AAAAAAAAAAA98051,AAAAAAAAAAAAAAAAAA71621}`}, {47, `{98,23,64,12,75,61}`, `{AAA59323,AAAAA95309,AAAAAAAAAAAAAAAA31334,AAAAAAAAA27249,AAAAA17383,AAAAAAAAAAAA37562,AAAAAA1059,A84822,55847,AAAAA70466}`}, {48, `{76,14}`, `{AAAAAAAAAAAAA59671,AAAAAAAAAAAAAAAAAAA91804,AAAAAA66777,AAAAAAAAAAAAAAAAAAA89194,AAAAAAAAAAAAAAA36627,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAA73084,AAAAAAA79710,AAAAAAAAAAAAAAA40402,AAAAAAAAAAAAAAAAAAA65037}`}, {49, `{56,5,54,37,49}`, `{AA21643,AAAAAAAAAAA92631,AAAAAAAA81587}`}, {50, `{20,12,37,64,93}`, `{AAAAAAAAAA5483,AAAAAAAAAAAAAAAAAAA1205,AA6416,AAAAAAAAAAAAAAAAA63050,AAAAAAAAAAAAAAAAAA47955}`}, {51, `{47}`, `{AAAAAAAAAAAAAA96505,AAAAAAAAAAAAAAAAAA36842,AAAAA95309,AAAAAAAA81587,AA6416,AAAA91194,AAAAAA58494,AAAAAA1059,AAAAAAAA69452}`}, {52, `{89,0}`, `{AAAAAAAAAAAAAAAAAA47955,AAAAAAA48038,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAAAAA73084,AAAAA70466,AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA46154,AA66862}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {54, `{70,47}`, `{AAAAAAAAAAAAAAAAAA54141,AAAAA40681,AAAAAAA48038,AAAAAAAAAAAAAAAA29150,AAAAA41597,AAAAAAAAAAAAAAAAAA59334,AA15322}`}, {55, `{47,79,47,64,72,25,71,24,93}`, `{AAAAAAAAAAAAAAAAAA55796,AAAAA62737}`}, {56, `{33,7,60,54,93,90,77,85,39}`, `{AAAAAAAAAAAAAAAAAA32918,AA42406}`}, {57, `{23,45,10,42,36,21,9,96}`, `{AAAAAAAAAAAAAAAAAAA70415}`}, {58, `{92}`, `{AAAAAAAAAAAAAAAA98414,AAAAAAAA23648,AAAAAAAAAAAAAAAAAA55796,AA25381,AAAAAAAAAAA6119}`}, {59, `{9,69,46,77}`, `{39557,AAAAAAA89932,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAAAAAAAAA26540,AAA20874,AA6416,AAAAAAAAAAAAAAAAAA47955}`}, {60, `{62,2,59,38,89}`, `{AAAAAAA89932,AAAAAAAAAAAAAAAAAAA15356,AA99927,AA17009,AAAAAAAAAAAAAAA35875}`}, {61, `{72,2,44,95,54,54,13}`, `{AAAAAAAAAAAAAAAAAAA91804}`}, {62, `{83,72,29,73}`, `{AAAAAAAAAAAAA15097,AAAA8857,AAAAAAAAAAAA35809,AAAAAAAAAAAA52814,AAAAAAAAAAAAAAAAAAA38885,AAAAAAAAAAAAAAAAAA24183,AAAAAA43678,A96617}`}, {63, `{11,4,61,87}`, `{AAAAAAAAA27249,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAA13198,AAA20874,39557,51533,AAAAAAAAAAA53908,AAAAAAAAAAAAAA96505,AAAAAAAA78938}`}, {64, `{26,19,34,24,81,78}`, `{A96617,AAAAAAAAAAAAAAAAAAA70104,A68938,AAAAAAAAAAA53908,AAAAAAAAAAAAAAA453,AA17009,AAAAAAA80240}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {66, `{31,23,70,52,4,33,48,25}`, `{AAAAAAAAAAAAAAAAA69675,AAAAAAAA50094,AAAAAAAAAAA92631,AAAA35194,39557,AAAAAAA99836}`}, {67, `{31,94,7,10}`, `{AAAAAA38063,A96617,AAAA35194,AAAAAAAAAAAA67946}`}, {68, `{90,43,38}`, `{AA75092,AAAAAAAAAAAAAAAAA69675,AAAAAAAAAAA92631,AAAAAAAAA10012,AAAAAAAAAAAAA7929,AA21643}`}, {69, `{67,35,99,85,72,86,44}`, `{AAAAAAAAAAAAAAAAAAA1205,AAAAAAAA50094,AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAAAAAAA47955}`}, {70, `{56,70,83}`, `{AAAA41702,AAAAAAAAAAA82945,AA21643,AAAAAAAAAAA99000,A27153,AA25381,AAAAAAAAAAAAAA96505,AAAAAAA1242}`}, {71, `{74,26}`, `{AAAAAAAAAAA50956,AA74433,AAAAAAA21462,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAA36627,AAAAAAAAAAAAA70254,AAAAAAAAAA43419,39557}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {73, `{88,25,96,78,65,15,29,19}`, `{AAA54451,AAAAAAAAA27249,AAAAAAA9228,AAAAAAAAAAAAAAA67062,AAAAAAAAAAAAAAAAAAA70415,AAAAA17383,AAAAAAAAAAAAAAAA33598}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {75, `{12,96,83,24,71,89,55}`, `{AAAA48949,AAAAAAAA29716,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAAA67946,AAAAAAAAAAAAAAAA29150,AAA28075,AAAAAAAAAAAAAAAAA43052}`}, {76, `{92,55,10,7}`, `{AAAAAAAAAAAAAAA67062}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {78, `{55,89,44,84,34}`, `{AAAAAAAAAAA6119,AAAAAAAAAAAAAA8666,AA99927,AA42406,AAAAAAA81898,AAAAAAA9228,AAAAAAAAAAA92631,AA21643,AAAAAAAAAAAAAA28620}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {80, `{74,89,44,80,0}`, `{AAAA35194,AAAAAAAA79710,AAA20874,AAAAAAAAAAAAAAAAAAA70104,AAAAAAAAAAAAA73084,AAAAAAA57334,AAAAAAA9228,AAAAAAAAAAAAA62007}`}, {81, `{63,77,54,48,61,53,97}`, `{AAAAAAAAAAAAAAA81326,AAAAAAAAAA22292,AA25381,AAAAAAAAAAA74076,AAAAAAA81898,AAAAAAAAA72121}`}, {82, `{34,60,4,79,78,16,86,89,42,50}`, `{AAAAA40681,AAAAAAAAAAAAAAAAAA12591,AAAAAAA80240,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA70104}`}, {83, `{14,10}`, `{AAAAAAAAAA22292,AAAAAAAAAAAAA70254,AAAAAAAAAAA6119}`}, {84, `{11,83,35,13,96,94}`, `{AAAAA95309,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAAAAA24183}`}, {85, `{39,60}`, `{AAAAAAAAAAAAAAAA55798,AAAAAAAAAA22292,AAAAAAA66161,AAAAAAA21462,AAAAAAAAAAAAAAAAAA12591,55847,AAAAAA98232,AAAAAAAAAAA46154}`}, {86, `{33,81,72,74,45,36,82}`, `{AAAAAAAA81587,AAAAAAAAAAAAAA96505,45449,AAAA80176}`}, {87, `{57,27,50,12,97,68}`, `{AAAAAAAAAAAAAAAAA26540,AAAAAAAAA10012,AAAAAAAAAAAA35809,AAAAAAAAAAAAAAAA29150,AAAAAAAAAAA82945,AAAAAA66777,31228,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAA96505}`}, {88, `{41,90,77,24,6,24}`, `{AAAA35194,AAAA35194,AAAAAAA80240,AAAAAAAAAAA46154,AAAAAA58494,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA59334,AAAAAAAAAAAAAAAAAAA91804,AA74433}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {90, `{88,75}`, `{AAAAA60038,AAAAAAAA23648,AAAAAAAAAAA99000,AAAA41702,AAAAAAAAAAAAA22860,AAAAAAAAAAAAAAA68526}`}, {91, `{78}`, `{AAAAAAAAAAAAA62007,AAA99043}`}, {92, `{85,63,49,45}`, `{AAAAAAA89932,AAAAAAAAAAAAA22860,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAAA21089}`}, {93, `{11}`, `{AAAAAAAAAAA176,AAAAAAAAAAAAAA8666,AAAAAAAAAAAAAAA453,AAAAAAAAAAAAA85723,A68938,AAAAAAAAAAAAA9821,AAAAAAA48038,AAAAAAAAAAAAAAAAA59387,AA99927,AAAAA17383}`}, {94, `{98,9,85,62,88,91,60,61,38,86}`, `{AAAAAAAA81587,AAAAA17383,AAAAAAAA81587}`}, {95, `{47,77}`, `{AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA74076,AAAAAAAAAA18107,AAAAA40681,AAAAAAAAAAAAAAA35875,AAAAA60038,AAAAAAA56483}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}, {97, `{54,2,86,65}`, `{47735,AAAAAAA99836,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAA29150,AAAAAAA80240,AAAAAAAAAAAAAAAA98414,AAAAAAA56483,AAAAAAAAAAAAAAAA29150,AAAAAAA39692,AA21643}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {99, `{37,86}`, `{AAAAAAAAAAAAAAAAAA32918,AAAAA70514,AAAAAAAAA10012,AAAAAAAAAAAAAAAAA59387,AAAAAAAAAA64777,AAAAAAAAAAAAAAAAAAA15356}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}, {101, `{}`, `{}`}, {102, `{NULL}`, `{NULL}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t && '{}' ORDER BY seqno;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t <@ '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `DROP INDEX intarrayidx, textarrayidx;`,
			},
			{
				Statement: `CREATE INDEX botharrayidx ON array_index_op_test USING gin (i, t);`,
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i @> '{32}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i && '{32}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t @> '{AAAAAAA80240}' ORDER BY seqno;`,
				Results:   []sql.Row{{19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {30, `{26,81,47,91,34}`, `{AAAAAAAAAAAAAAAAAAA70104,AAAAAAA80240}`}, {64, `{26,19,34,24,81,78}`, `{A96617,AAAAAAAAAAAAAAAAAAA70104,A68938,AAAAAAAAAAA53908,AAAAAAAAAAAAAAA453,AA17009,AAAAAAA80240}`}, {82, `{34,60,4,79,78,16,86,89,42,50}`, `{AAAAA40681,AAAAAAAAAAAAAAAAAA12591,AAAAAAA80240,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA70104}`}, {88, `{41,90,77,24,6,24}`, `{AAAA35194,AAAA35194,AAAAAAA80240,AAAAAAAAAAA46154,AAAAAA58494,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA59334,AAAAAAAAAAAAAAAAAAA91804,AA74433}`}, {97, `{54,2,86,65}`, `{47735,AAAAAAA99836,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAA29150,AAAAAAA80240,AAAAAAAAAAAAAAAA98414,AAAAAAA56483,AAAAAAAAAAAAAAAA29150,AAAAAAA39692,AA21643}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t && '{AAAAAAA80240}' ORDER BY seqno;`,
				Results:   []sql.Row{{19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {30, `{26,81,47,91,34}`, `{AAAAAAAAAAAAAAAAAAA70104,AAAAAAA80240}`}, {64, `{26,19,34,24,81,78}`, `{A96617,AAAAAAAAAAAAAAAAAAA70104,A68938,AAAAAAAAAAA53908,AAAAAAAAAAAAAAA453,AA17009,AAAAAAA80240}`}, {82, `{34,60,4,79,78,16,86,89,42,50}`, `{AAAAA40681,AAAAAAAAAAAAAAAAAA12591,AAAAAAA80240,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA70104}`}, {88, `{41,90,77,24,6,24}`, `{AAAA35194,AAAA35194,AAAAAAA80240,AAAAAAAAAAA46154,AAAAAA58494,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA59334,AAAAAAAAAAAAAAAAAAA91804,AA74433}`}, {97, `{54,2,86,65}`, `{47735,AAAAAAA99836,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAA29150,AAAAAAA80240,AAAAAAAAAAAAAAAA98414,AAAAAAA56483,AAAAAAAAAAAAAAAA29150,AAAAAAA39692,AA21643}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i @> '{32}' AND t && '{AAAAAAA80240}' ORDER BY seqno;`,
				Results:   []sql.Row{{100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE i && '{32}' AND t @> '{AAAAAAA80240}' ORDER BY seqno;`,
				Results:   []sql.Row{{100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_index_op_test WHERE t = '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
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
			{
				Statement: `CREATE TABLE array_gin_test (a int[]);`,
			},
			{
				Statement: `INSERT INTO array_gin_test SELECT ARRAY[1, g%5, g] FROM generate_series(1, 10000) g;`,
			},
			{
				Statement: `CREATE INDEX array_gin_test_idx ON array_gin_test USING gin (a);`,
			},
			{
				Statement: `SELECT COUNT(*) FROM array_gin_test WHERE a @> '{2}';`,
				Results:   []sql.Row{{2000}},
			},
			{
				Statement: `DROP TABLE array_gin_test;`,
			},
			{
				Statement: `CREATE INDEX gin_relopts_test ON array_index_op_test USING gin (i)
  WITH (FASTUPDATE=on, GIN_PENDING_LIST_LIMIT=128);`,
			},
			{
				Statement: `\d+ gin_relopts_test
                Index "public.gin_relopts_test"
 Column |  Type   | Key? | Definition | Storage | Stats target 
--------+---------+------+------------+---------+--------------
 i      | integer | yes  | i          | plain   | 
gin, for table "public.array_index_op_test"
Options: fastupdate=on, gin_pending_list_limit=128
CREATE UNLOGGED TABLE unlogged_hash_table (id int4);`,
			},
			{
				Statement: `CREATE INDEX unlogged_hash_index ON unlogged_hash_table USING hash (id int4_ops);`,
			},
			{
				Statement: `DROP TABLE unlogged_hash_table;`,
			},
			{
				Statement: `SET maintenance_work_mem = '1MB';`,
			},
			{
				Statement: `CREATE INDEX hash_tuplesort_idx ON tenk1 USING hash (stringu1 name_ops) WITH (fillfactor = 10);`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM tenk1 WHERE stringu1 = 'TVAAAA';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on tenk1`}, {`Recheck Cond: (stringu1 = 'TVAAAA'::name)`}, {`->  Bitmap Index Scan on hash_tuplesort_idx`}, {`Index Cond: (stringu1 = 'TVAAAA'::name)`}},
			},
			{
				Statement: `SELECT count(*) FROM tenk1 WHERE stringu1 = 'TVAAAA';`,
				Results:   []sql.Row{{14}},
			},
			{
				Statement: `DROP INDEX hash_tuplesort_idx;`,
			},
			{
				Statement: `RESET maintenance_work_mem;`,
			},
			{
				Statement: `CREATE TABLE unique_tbl (i int, t text);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX unique_idx1 ON unique_tbl (i) NULLS DISTINCT;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX unique_idx2 ON unique_tbl (i) NULLS NOT DISTINCT;`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (1, 'one');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (2, 'two');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (3, 'three');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (4, 'four');`,
			},
			{
				Statement: `INSERT INTO unique_tbl VALUES (5, 'one');`,
			},
			{
				Statement: `INSERT INTO unique_tbl (t) VALUES ('six');`,
			},
			{
				Statement:   `INSERT INTO unique_tbl (t) VALUES ('seven');  -- error from unique_idx2`,
				ErrorString: `duplicate key value violates unique constraint "unique_idx2"`,
			},
			{
				Statement: `DETAIL:  Key (i)=(null) already exists.
DROP INDEX unique_idx1, unique_idx2;`,
			},
			{
				Statement: `INSERT INTO unique_tbl (t) VALUES ('seven');`,
			},
			{
				Statement: `CREATE UNIQUE INDEX unique_idx3 ON unique_tbl (i) NULLS DISTINCT;  -- ok`,
			},
			{
				Statement:   `CREATE UNIQUE INDEX unique_idx4 ON unique_tbl (i) NULLS NOT DISTINCT;  -- error`,
				ErrorString: `could not create unique index "unique_idx4"`,
			},
			{
				Statement: `DETAIL:  Key (i)=(null) is duplicated.
DELETE FROM unique_tbl WHERE t = 'seven';`,
			},
			{
				Statement: `CREATE UNIQUE INDEX unique_idx4 ON unique_tbl (i) NULLS NOT DISTINCT;  -- ok now`,
			},
			{
				Statement: `\d unique_tbl
             Table "public.unique_tbl"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 i      | integer |           |          | 
 t      | text    |           |          | 
Indexes:
    "unique_idx3" UNIQUE, btree (i)
    "unique_idx4" UNIQUE, btree (i) NULLS NOT DISTINCT
\d unique_idx3
      Index "public.unique_idx3"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 i      | integer | yes  | i
unique, btree, for table "public.unique_tbl"
\d unique_idx4
      Index "public.unique_idx4"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 i      | integer | yes  | i
unique nulls not distinct, btree, for table "public.unique_tbl"
SELECT pg_get_indexdef('unique_idx3'::regclass);`,
				Results: []sql.Row{{`CREATE UNIQUE INDEX unique_idx3 ON public.unique_tbl USING btree (i)`}},
			},
			{
				Statement: `SELECT pg_get_indexdef('unique_idx4'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX unique_idx4 ON public.unique_tbl USING btree (i) NULLS NOT DISTINCT`}},
			},
			{
				Statement: `DROP TABLE unique_tbl;`,
			},
			{
				Statement: `CREATE TABLE func_index_heap (f1 text, f2 text);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX func_index_index on func_index_heap (textcat(f1,f2));`,
			},
			{
				Statement: `INSERT INTO func_index_heap VALUES('ABC','DEF');`,
			},
			{
				Statement: `INSERT INTO func_index_heap VALUES('AB','CDEFG');`,
			},
			{
				Statement: `INSERT INTO func_index_heap VALUES('QWE','RTY');`,
			},
			{
				Statement:   `INSERT INTO func_index_heap VALUES('ABCD', 'EF');`,
				ErrorString: `duplicate key value violates unique constraint "func_index_index"`,
			},
			{
				Statement: `DETAIL:  Key (textcat(f1, f2))=(ABCDEF) already exists.
INSERT INTO func_index_heap VALUES('QWERTY');`,
			},
			{
				Statement: `\d func_index_heap
         Table "public.func_index_heap"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 f1     | text |           |          | 
 f2     | text |           |          | 
Indexes:
    "func_index_index" UNIQUE, btree (textcat(f1, f2))
\d func_index_index
     Index "public.func_index_index"
 Column  | Type | Key? |   Definition    
---------+------+------+-----------------
 textcat | text | yes  | textcat(f1, f2)
unique, btree, for table "public.func_index_heap"
DROP TABLE func_index_heap;`,
			},
			{
				Statement: `CREATE TABLE func_index_heap (f1 text, f2 text);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX func_index_index on func_index_heap ((f1 || f2) text_ops);`,
			},
			{
				Statement: `INSERT INTO func_index_heap VALUES('ABC','DEF');`,
			},
			{
				Statement: `INSERT INTO func_index_heap VALUES('AB','CDEFG');`,
			},
			{
				Statement: `INSERT INTO func_index_heap VALUES('QWE','RTY');`,
			},
			{
				Statement:   `INSERT INTO func_index_heap VALUES('ABCD', 'EF');`,
				ErrorString: `duplicate key value violates unique constraint "func_index_index"`,
			},
			{
				Statement: `DETAIL:  Key ((f1 || f2))=(ABCDEF) already exists.
INSERT INTO func_index_heap VALUES('QWERTY');`,
			},
			{
				Statement: `\d func_index_heap
         Table "public.func_index_heap"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 f1     | text |           |          | 
 f2     | text |           |          | 
Indexes:
    "func_index_index" UNIQUE, btree ((f1 || f2))
\d func_index_index
  Index "public.func_index_index"
 Column | Type | Key? | Definition 
--------+------+------+------------
 expr   | text | yes  | (f1 || f2)
unique, btree, for table "public.func_index_heap"
create index on func_index_heap ((f1 || f2), (row(f1, f2)));`,
				ErrorString: `column "row" has pseudo-type record`,
			},
			{
				Statement: `CREATE TABLE covering_index_heap (f1 int, f2 int, f3 text);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX covering_index_index on covering_index_heap (f1,f2) INCLUDE(f3);`,
			},
			{
				Statement: `INSERT INTO covering_index_heap VALUES(1,1,'AAA');`,
			},
			{
				Statement: `INSERT INTO covering_index_heap VALUES(1,2,'AAA');`,
			},
			{
				Statement:   `INSERT INTO covering_index_heap VALUES(1,2,'BBB');`,
				ErrorString: `duplicate key value violates unique constraint "covering_index_index"`,
			},
			{
				Statement: `DETAIL:  Key (f1, f2)=(1, 2) already exists.
INSERT INTO covering_index_heap VALUES(1,4,'AAA');`,
			},
			{
				Statement: `CREATE UNIQUE INDEX covering_pkey on covering_index_heap (f1,f2) INCLUDE(f3);`,
			},
			{
				Statement: `ALTER TABLE covering_index_heap ADD CONSTRAINT covering_pkey PRIMARY KEY USING INDEX
covering_pkey;`,
			},
			{
				Statement: `DROP TABLE covering_index_heap;`,
			},
			{
				Statement: `CREATE TABLE concur_heap (f1 text, f2 text);`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY concur_index1 ON concur_heap(f2,f1);`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY IF NOT EXISTS concur_index1 ON concur_heap(f2,f1);`,
			},
			{
				Statement: `INSERT INTO concur_heap VALUES  ('a','b');`,
			},
			{
				Statement: `INSERT INTO concur_heap VALUES  ('b','b');`,
			},
			{
				Statement: `CREATE UNIQUE INDEX CONCURRENTLY concur_index2 ON concur_heap(f1);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS concur_index2 ON concur_heap(f1);`,
			},
			{
				Statement:   `INSERT INTO concur_heap VALUES ('b','x');`,
				ErrorString: `duplicate key value violates unique constraint "concur_index2"`,
			},
			{
				Statement: `DETAIL:  Key (f1)=(b) already exists.
CREATE UNIQUE INDEX CONCURRENTLY concur_index3 ON concur_heap(f2);`,
				ErrorString: `could not create unique index "concur_index3"`,
			},
			{
				Statement: `DETAIL:  Key (f2)=(b) is duplicated.
CREATE INDEX CONCURRENTLY concur_index4 on concur_heap(f2) WHERE f1='a';`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY concur_index5 on concur_heap(f2) WHERE f1='x';`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY on concur_heap((f2||f1));`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `CREATE INDEX CONCURRENTLY concur_index7 ON concur_heap(f1);`,
				ErrorString: `CREATE INDEX CONCURRENTLY cannot run inside a transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE FUNCTION predicate_stable() RETURNS bool IMMUTABLE
LANGUAGE plpgsql AS $$
BEGIN
  EXECUTE 'SELECT txid_current()';`,
			},
			{
				Statement: `  RETURN true;`,
			},
			{
				Statement: `END; $$;`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY concur_index8 ON concur_heap (f1)
  WHERE predicate_stable();`,
			},
			{
				Statement: `DROP INDEX concur_index8;`,
			},
			{
				Statement: `DROP FUNCTION predicate_stable();`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE INDEX std_index on concur_heap(f2);`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `VACUUM FULL concur_heap;`,
			},
			{
				Statement:   `REINDEX TABLE concur_heap;`,
				ErrorString: `could not create unique index "concur_index3"`,
			},
			{
				Statement: `DETAIL:  Key (f2)=(b) is duplicated.
DELETE FROM concur_heap WHERE f1 = 'b';`,
			},
			{
				Statement: `VACUUM FULL concur_heap;`,
			},
			{
				Statement: `\d concur_heap
           Table "public.concur_heap"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 f1     | text |           |          | 
 f2     | text |           |          | 
Indexes:
    "concur_heap_expr_idx" btree ((f2 || f1))
    "concur_index1" btree (f2, f1)
    "concur_index2" UNIQUE, btree (f1)
    "concur_index3" UNIQUE, btree (f2) INVALID
    "concur_index4" btree (f2) WHERE f1 = 'a'::text
    "concur_index5" btree (f2) WHERE f1 = 'x'::text
    "std_index" btree (f2)
REINDEX TABLE concur_heap;`,
			},
			{
				Statement: `\d concur_heap
           Table "public.concur_heap"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 f1     | text |           |          | 
 f2     | text |           |          | 
Indexes:
    "concur_heap_expr_idx" btree ((f2 || f1))
    "concur_index1" btree (f2, f1)
    "concur_index2" UNIQUE, btree (f1)
    "concur_index3" UNIQUE, btree (f2)
    "concur_index4" btree (f2) WHERE f1 = 'a'::text
    "concur_index5" btree (f2) WHERE f1 = 'x'::text
    "std_index" btree (f2)
CREATE TEMP TABLE concur_temp (f1 int, f2 text)
  ON COMMIT PRESERVE ROWS;`,
			},
			{
				Statement: `INSERT INTO concur_temp VALUES (1, 'foo'), (2, 'bar');`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY concur_temp_ind ON concur_temp(f1);`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY concur_temp_ind;`,
			},
			{
				Statement: `DROP TABLE concur_temp;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TEMP TABLE concur_temp (f1 int, f2 text)
  ON COMMIT DROP;`,
			},
			{
				Statement: `INSERT INTO concur_temp VALUES (1, 'foo'), (2, 'bar');`,
			},
			{
				Statement:   `CREATE INDEX CONCURRENTLY concur_temp_ind ON concur_temp(f1);`,
				ErrorString: `CREATE INDEX CONCURRENTLY cannot run inside a transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE TEMP TABLE concur_temp (f1 int, f2 text)
  ON COMMIT DELETE ROWS;`,
			},
			{
				Statement: `INSERT INTO concur_temp VALUES (1, 'foo'), (2, 'bar');`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY concur_temp_ind ON concur_temp(f1);`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY concur_temp_ind;`,
			},
			{
				Statement: `DROP TABLE concur_temp;`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY "concur_index2";				-- works`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY IF EXISTS "concur_index2";		-- notice`,
			},
			{
				Statement:   `DROP INDEX CONCURRENTLY "concur_index2", "concur_index3";`,
				ErrorString: `DROP INDEX CONCURRENTLY does not support dropping multiple objects`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `DROP INDEX CONCURRENTLY "concur_index5";`,
				ErrorString: `DROP INDEX CONCURRENTLY cannot run inside a transaction block`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY IF EXISTS "concur_index3";`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY "concur_index4";`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY "concur_index5";`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY "concur_index1";`,
			},
			{
				Statement: `DROP INDEX CONCURRENTLY "concur_heap_expr_idx";`,
			},
			{
				Statement: `\d concur_heap
           Table "public.concur_heap"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 f1     | text |           |          | 
 f2     | text |           |          | 
Indexes:
    "std_index" btree (f2)
DROP TABLE concur_heap;`,
			},
			{
				Statement: `CREATE TABLE cwi_test( a int , b varchar(10), c char);`,
			},
			{
				Statement: `INSERT INTO cwi_test VALUES(1, 2), (3, 4), (5, 6);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX cwi_uniq_idx ON cwi_test(a , b);`,
			},
			{
				Statement: `ALTER TABLE cwi_test ADD primary key USING INDEX cwi_uniq_idx;`,
			},
			{
				Statement: `\d cwi_test
                     Table "public.cwi_test"
 Column |         Type          | Collation | Nullable | Default 
--------+-----------------------+-----------+----------+---------
 a      | integer               |           | not null | 
 b      | character varying(10) |           | not null | 
 c      | character(1)          |           |          | 
Indexes:
    "cwi_uniq_idx" PRIMARY KEY, btree (a, b)
\d cwi_uniq_idx
            Index "public.cwi_uniq_idx"
 Column |         Type          | Key? | Definition 
--------+-----------------------+------+------------
 a      | integer               | yes  | a
 b      | character varying(10) | yes  | b
primary key, btree, for table "public.cwi_test"
CREATE UNIQUE INDEX cwi_uniq2_idx ON cwi_test(b , a);`,
			},
			{
				Statement: `ALTER TABLE cwi_test DROP CONSTRAINT cwi_uniq_idx,
	ADD CONSTRAINT cwi_replaced_pkey PRIMARY KEY
		USING INDEX cwi_uniq2_idx;`,
			},
			{
				Statement: `\d cwi_test
                     Table "public.cwi_test"
 Column |         Type          | Collation | Nullable | Default 
--------+-----------------------+-----------+----------+---------
 a      | integer               |           | not null | 
 b      | character varying(10) |           | not null | 
 c      | character(1)          |           |          | 
Indexes:
    "cwi_replaced_pkey" PRIMARY KEY, btree (b, a)
\d cwi_replaced_pkey
          Index "public.cwi_replaced_pkey"
 Column |         Type          | Key? | Definition 
--------+-----------------------+------+------------
 b      | character varying(10) | yes  | b
 a      | integer               | yes  | a
primary key, btree, for table "public.cwi_test"
DROP INDEX cwi_replaced_pkey;	-- Should fail; a constraint depends on it`,
				ErrorString: `cannot drop index cwi_replaced_pkey because constraint cwi_replaced_pkey on table cwi_test requires it`,
			},
			{
				Statement: `CREATE UNIQUE INDEX cwi_uniq3_idx ON cwi_test(a desc);`,
			},
			{
				Statement:   `ALTER TABLE cwi_test ADD UNIQUE USING INDEX cwi_uniq3_idx;  -- fail`,
				ErrorString: `index "cwi_uniq3_idx" column number 1 does not have default sorting behavior`,
			},
			{
				Statement: `DETAIL:  Cannot create a primary key or unique constraint using such an index.
CREATE UNIQUE INDEX cwi_uniq4_idx ON cwi_test(b collate "POSIX");`,
			},
			{
				Statement:   `ALTER TABLE cwi_test ADD UNIQUE USING INDEX cwi_uniq4_idx;  -- fail`,
				ErrorString: `index "cwi_uniq4_idx" column number 1 does not have default sorting behavior`,
			},
			{
				Statement: `DETAIL:  Cannot create a primary key or unique constraint using such an index.
DROP TABLE cwi_test;`,
			},
			{
				Statement: `CREATE TABLE cwi_test(a int) PARTITION BY hash (a);`,
			},
			{
				Statement: `create unique index on cwi_test (a);`,
			},
			{
				Statement:   `alter table cwi_test add primary key using index cwi_test_a_idx ;`,
				ErrorString: `ALTER TABLE / ADD CONSTRAINT USING INDEX is not supported on partitioned tables`,
			},
			{
				Statement: `DROP TABLE cwi_test;`,
			},
			{
				Statement: `CREATE TABLE syscol_table (a INT);`,
			},
			{
				Statement:   `CREATE INDEX ON syscolcol_table (ctid);`,
				ErrorString: `relation "syscolcol_table" does not exist`,
			},
			{
				Statement:   `CREATE INDEX ON syscol_table ((ctid >= '(1000,0)'));`,
				ErrorString: `index creation on system columns is not supported`,
			},
			{
				Statement:   `CREATE INDEX ON syscol_table (a) WHERE ctid >= '(1000,0)';`,
				ErrorString: `index creation on system columns is not supported`,
			},
			{
				Statement: `DROP TABLE syscol_table;`,
			},
			{
				Statement: `CREATE TABLE onek_with_null AS SELECT unique1, unique2 FROM onek;`,
			},
			{
				Statement: `INSERT INTO onek_with_null (unique1,unique2) VALUES (NULL, -1), (NULL, NULL);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX onek_nulltest ON onek_with_null (unique2,unique1);`,
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
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique2 IS NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NOT NULL;`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique2 IS NOT NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NOT NULL AND unique1 > 500;`,
				Results:   []sql.Row{{499}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique1 > 500;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `DROP INDEX onek_nulltest;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX onek_nulltest ON onek_with_null (unique2 desc,unique1);`,
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique2 IS NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NOT NULL;`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique2 IS NOT NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NOT NULL AND unique1 > 500;`,
				Results:   []sql.Row{{499}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique1 > 500;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `DROP INDEX onek_nulltest;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX onek_nulltest ON onek_with_null (unique2 desc nulls last,unique1);`,
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique2 IS NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NOT NULL;`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique2 IS NOT NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NOT NULL AND unique1 > 500;`,
				Results:   []sql.Row{{499}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique1 > 500;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `DROP INDEX onek_nulltest;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX onek_nulltest ON onek_with_null (unique2  nulls first,unique1);`,
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique2 IS NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NOT NULL;`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique2 IS NOT NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NOT NULL AND unique1 > 500;`,
				Results:   []sql.Row{{499}},
			},
			{
				Statement: `SELECT count(*) FROM onek_with_null WHERE unique1 IS NULL AND unique1 > 500;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `DROP INDEX onek_nulltest;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX onek_nulltest ON onek_with_null (unique2);`,
			},
			{
				Statement: `SET enable_seqscan = OFF;`,
			},
			{
				Statement: `SET enable_indexscan = ON;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `SELECT unique1, unique2 FROM onek_with_null
  ORDER BY unique2 LIMIT 2;`,
				Results: []sql.Row{{``, -1}, {147, 0}},
			},
			{
				Statement: `SELECT unique1, unique2 FROM onek_with_null WHERE unique2 >= -1
  ORDER BY unique2 LIMIT 2;`,
				Results: []sql.Row{{``, -1}, {147, 0}},
			},
			{
				Statement: `SELECT unique1, unique2 FROM onek_with_null WHERE unique2 >= 0
  ORDER BY unique2 LIMIT 2;`,
				Results: []sql.Row{{147, 0}, {931, 1}},
			},
			{
				Statement: `SELECT unique1, unique2 FROM onek_with_null
  ORDER BY unique2 DESC LIMIT 2;`,
				Results: []sql.Row{{``, ``}, {278, 999}},
			},
			{
				Statement: `SELECT unique1, unique2 FROM onek_with_null WHERE unique2 >= -1
  ORDER BY unique2 DESC LIMIT 2;`,
				Results: []sql.Row{{278, 999}, {0, 998}},
			},
			{
				Statement: `SELECT unique1, unique2 FROM onek_with_null WHERE unique2 < 999
  ORDER BY unique2 DESC LIMIT 2;`,
				Results: []sql.Row{{0, 998}, {744, 997}},
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
			{
				Statement: `DROP TABLE onek_with_null;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM tenk1
  WHERE thousand = 42 AND (tenthous = 1 OR tenthous = 3 OR tenthous = 42);`,
				Results: []sql.Row{{`Bitmap Heap Scan on tenk1`}, {`Recheck Cond: (((thousand = 42) AND (tenthous = 1)) OR ((thousand = 42) AND (tenthous = 3)) OR ((thousand = 42) AND (tenthous = 42)))`}, {`->  BitmapOr`}, {`->  Bitmap Index Scan on tenk1_thous_tenthous`}, {`Index Cond: ((thousand = 42) AND (tenthous = 1))`}, {`->  Bitmap Index Scan on tenk1_thous_tenthous`}, {`Index Cond: ((thousand = 42) AND (tenthous = 3))`}, {`->  Bitmap Index Scan on tenk1_thous_tenthous`}, {`Index Cond: ((thousand = 42) AND (tenthous = 42))`}},
			},
			{
				Statement: `SELECT * FROM tenk1
  WHERE thousand = 42 AND (tenthous = 1 OR tenthous = 3 OR tenthous = 42);`,
				Results: []sql.Row{{42, 5530, 0, 2, 2, 2, 42, 42, 42, 42, 42, 84, 85, `QBAAAA`, `SEIAAA`, `OOOOxx`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM tenk1
  WHERE hundred = 42 AND (thousand = 42 OR thousand = 99);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on tenk1`}, {`Recheck Cond: ((hundred = 42) AND ((thousand = 42) OR (thousand = 99)))`}, {`->  BitmapAnd`}, {`->  Bitmap Index Scan on tenk1_hundred`}, {`Index Cond: (hundred = 42)`}, {`->  BitmapOr`}, {`->  Bitmap Index Scan on tenk1_thous_tenthous`}, {`Index Cond: (thousand = 42)`}, {`->  Bitmap Index Scan on tenk1_thous_tenthous`}, {`Index Cond: (thousand = 99)`}},
			},
			{
				Statement: `SELECT count(*) FROM tenk1
  WHERE hundred = 42 AND (thousand = 42 OR thousand = 99);`,
				Results: []sql.Row{{10}},
			},
			{
				Statement: `CREATE TABLE dupindexcols AS
  SELECT unique1 as id, stringu2::text as f1 FROM tenk1;`,
			},
			{
				Statement: `CREATE INDEX dupindexcols_i ON dupindexcols (f1, id, f1 text_pattern_ops);`,
			},
			{
				Statement: `ANALYZE dupindexcols;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
  SELECT count(*) FROM dupindexcols
    WHERE f1 BETWEEN 'WA' AND 'ZZZ' and id < 1000 and f1 ~<~ 'YX';`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on dupindexcols`}, {`Recheck Cond: ((f1 >= 'WA'::text) AND (f1 <= 'ZZZ'::text) AND (id < 1000) AND (f1 ~<~ 'YX'::text))`}, {`->  Bitmap Index Scan on dupindexcols_i`}, {`Index Cond: ((f1 >= 'WA'::text) AND (f1 <= 'ZZZ'::text) AND (id < 1000) AND (f1 ~<~ 'YX'::text))`}},
			},
			{
				Statement: `SELECT count(*) FROM dupindexcols
  WHERE f1 BETWEEN 'WA' AND 'ZZZ' and id < 1000 and f1 ~<~ 'YX';`,
				Results: []sql.Row{{97}},
			},
			{
				Statement: `explain (costs off)
SELECT unique1 FROM tenk1
WHERE unique1 IN (1,42,7)
ORDER BY unique1;`,
				Results: []sql.Row{{`Index Only Scan using tenk1_unique1 on tenk1`}, {`Index Cond: (unique1 = ANY ('{1,42,7}'::integer[]))`}},
			},
			{
				Statement: `SELECT unique1 FROM tenk1
WHERE unique1 IN (1,42,7)
ORDER BY unique1;`,
				Results: []sql.Row{{1}, {7}, {42}},
			},
			{
				Statement: `explain (costs off)
SELECT thousand, tenthous FROM tenk1
WHERE thousand < 2 AND tenthous IN (1001,3000)
ORDER BY thousand;`,
				Results: []sql.Row{{`Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`Index Cond: (thousand < 2)`}, {`Filter: (tenthous = ANY ('{1001,3000}'::integer[]))`}},
			},
			{
				Statement: `SELECT thousand, tenthous FROM tenk1
WHERE thousand < 2 AND tenthous IN (1001,3000)
ORDER BY thousand;`,
				Results: []sql.Row{{0, 3000}, {1, 1001}},
			},
			{
				Statement: `SET enable_indexonlyscan = OFF;`,
			},
			{
				Statement: `explain (costs off)
SELECT thousand, tenthous FROM tenk1
WHERE thousand < 2 AND tenthous IN (1001,3000)
ORDER BY thousand;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: thousand`}, {`->  Index Scan using tenk1_thous_tenthous on tenk1`}, {`Index Cond: ((thousand < 2) AND (tenthous = ANY ('{1001,3000}'::integer[])))`}},
			},
			{
				Statement: `SELECT thousand, tenthous FROM tenk1
WHERE thousand < 2 AND tenthous IN (1001,3000)
ORDER BY thousand;`,
				Results: []sql.Row{{0, 3000}, {1, 1001}},
			},
			{
				Statement: `RESET enable_indexonlyscan;`,
			},
			{
				Statement: `explain (costs off)
  select * from tenk1 where (thousand, tenthous) in ((1,1001), (null,null));`,
				Results: []sql.Row{{`Index Scan using tenk1_thous_tenthous on tenk1`}, {`Index Cond: ((thousand = 1) AND (tenthous = 1001))`}},
			},
			{
				Statement: `create temp table boolindex (b bool, i int, unique(b, i), junk float);`,
			},
			{
				Statement: `explain (costs off)
  select * from boolindex order by b, i limit 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using boolindex_b_i_key on boolindex`}},
			},
			{
				Statement: `explain (costs off)
  select * from boolindex where b order by i limit 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using boolindex_b_i_key on boolindex`}, {`Index Cond: (b = true)`}},
			},
			{
				Statement: `explain (costs off)
  select * from boolindex where b = true order by i desc limit 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan Backward using boolindex_b_i_key on boolindex`}, {`Index Cond: (b = true)`}},
			},
			{
				Statement: `explain (costs off)
  select * from boolindex where not b order by i limit 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan using boolindex_b_i_key on boolindex`}, {`Index Cond: (b = false)`}},
			},
			{
				Statement: `explain (costs off)
  select * from boolindex where b is true order by i desc limit 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan Backward using boolindex_b_i_key on boolindex`}, {`Index Cond: (b = true)`}},
			},
			{
				Statement: `explain (costs off)
  select * from boolindex where b is false order by i desc limit 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Scan Backward using boolindex_b_i_key on boolindex`}, {`Index Cond: (b = false)`}},
			},
			{
				Statement: `CREATE TABLE reindex_verbose(id integer primary key);`,
			},
			{
				Statement: `\set VERBOSITY terse \\ -- suppress machine-dependent details
REINDEX (VERBOSE) TABLE reindex_verbose;`,
			},
			{
				Statement: `INFO:  index "reindex_verbose_pkey" was reindexed
\set VERBOSITY default
DROP TABLE reindex_verbose;`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_tab (c1 int);`,
			},
			{
				Statement: `REINDEX TABLE concur_reindex_tab; -- notice`,
			},
			{
				Statement: `REINDEX (CONCURRENTLY) TABLE concur_reindex_tab; -- notice`,
			},
			{
				Statement: `ALTER TABLE concur_reindex_tab ADD COLUMN c2 text; -- add toast index`,
			},
			{
				Statement: `CREATE UNIQUE INDEX concur_reindex_ind1 ON concur_reindex_tab(c1);`,
			},
			{
				Statement: `CREATE INDEX concur_reindex_ind2 ON concur_reindex_tab(c2);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX concur_reindex_ind3 ON concur_reindex_tab(abs(c1));`,
			},
			{
				Statement: `CREATE INDEX concur_reindex_ind4 ON concur_reindex_tab(c1, c1, c2);`,
			},
			{
				Statement: `ALTER TABLE concur_reindex_tab ADD PRIMARY KEY USING INDEX concur_reindex_ind1;`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_tab2 (c1 int REFERENCES concur_reindex_tab);`,
			},
			{
				Statement: `INSERT INTO concur_reindex_tab VALUES  (1, 'a');`,
			},
			{
				Statement: `INSERT INTO concur_reindex_tab VALUES  (2, 'a');`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_tab3 (c1 int, c2 int4range, EXCLUDE USING gist (c2 WITH &&));`,
			},
			{
				Statement: `INSERT INTO concur_reindex_tab3 VALUES  (3, '[1,2]');`,
			},
			{
				Statement:   `REINDEX INDEX CONCURRENTLY  concur_reindex_tab3_c2_excl;  -- error`,
				ErrorString: `concurrent index creation for exclusion constraints is not supported`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_reindex_tab3;  -- succeeds with warning`,
			},
			{
				Statement:   `INSERT INTO concur_reindex_tab3 VALUES  (4, '[2,4]');`,
				ErrorString: `conflicting key value violates exclusion constraint "concur_reindex_tab3_c2_excl"`,
			},
			{
				Statement: `DETAIL:  Key (c2)=([2,5)) conflicts with existing key (c2)=([1,3)).
CREATE MATERIALIZED VIEW concur_reindex_matview AS SELECT * FROM concur_reindex_tab;`,
			},
			{
				Statement: `SELECT pg_describe_object(classid, objid, objsubid) as obj,
       pg_describe_object(refclassid,refobjid,refobjsubid) as objref,
       deptype
FROM pg_depend
WHERE classid = 'pg_class'::regclass AND
  objid in ('concur_reindex_tab'::regclass,
            'concur_reindex_ind1'::regclass,
	    'concur_reindex_ind2'::regclass,
	    'concur_reindex_ind3'::regclass,
	    'concur_reindex_ind4'::regclass,
	    'concur_reindex_matview'::regclass)
  ORDER BY 1, 2;`,
				Results: []sql.Row{{`index concur_reindex_ind1`, `constraint concur_reindex_ind1 on table concur_reindex_tab`, `i`}, {`index concur_reindex_ind2`, `column c2 of table concur_reindex_tab`, `a`}, {`index concur_reindex_ind3`, `column c1 of table concur_reindex_tab`, `a`}, {`index concur_reindex_ind3`, `table concur_reindex_tab`, `a`}, {`index concur_reindex_ind4`, `column c1 of table concur_reindex_tab`, `a`}, {`index concur_reindex_ind4`, `column c2 of table concur_reindex_tab`, `a`}, {`materialized view concur_reindex_matview`, `schema public`, `n`}, {`table concur_reindex_tab`, `schema public`, `n`}},
			},
			{
				Statement: `REINDEX INDEX CONCURRENTLY concur_reindex_ind1;`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_reindex_tab;`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_reindex_matview;`,
			},
			{
				Statement: `SELECT pg_describe_object(classid, objid, objsubid) as obj,
       pg_describe_object(refclassid,refobjid,refobjsubid) as objref,
       deptype
FROM pg_depend
WHERE classid = 'pg_class'::regclass AND
  objid in ('concur_reindex_tab'::regclass,
            'concur_reindex_ind1'::regclass,
	    'concur_reindex_ind2'::regclass,
	    'concur_reindex_ind3'::regclass,
	    'concur_reindex_ind4'::regclass,
	    'concur_reindex_matview'::regclass)
  ORDER BY 1, 2;`,
				Results: []sql.Row{{`index concur_reindex_ind1`, `constraint concur_reindex_ind1 on table concur_reindex_tab`, `i`}, {`index concur_reindex_ind2`, `column c2 of table concur_reindex_tab`, `a`}, {`index concur_reindex_ind3`, `column c1 of table concur_reindex_tab`, `a`}, {`index concur_reindex_ind3`, `table concur_reindex_tab`, `a`}, {`index concur_reindex_ind4`, `column c1 of table concur_reindex_tab`, `a`}, {`index concur_reindex_ind4`, `column c2 of table concur_reindex_tab`, `a`}, {`materialized view concur_reindex_matview`, `schema public`, `n`}, {`table concur_reindex_tab`, `schema public`, `n`}},
			},
			{
				Statement: `CREATE TABLE testcomment (i int);`,
			},
			{
				Statement: `CREATE INDEX testcomment_idx1 ON testcomment (i);`,
			},
			{
				Statement: `COMMENT ON INDEX testcomment_idx1 IS 'test comment';`,
			},
			{
				Statement: `SELECT obj_description('testcomment_idx1'::regclass, 'pg_class');`,
				Results:   []sql.Row{{`test comment`}},
			},
			{
				Statement: `REINDEX TABLE testcomment;`,
			},
			{
				Statement: `SELECT obj_description('testcomment_idx1'::regclass, 'pg_class');`,
				Results:   []sql.Row{{`test comment`}},
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY testcomment ;`,
			},
			{
				Statement: `SELECT obj_description('testcomment_idx1'::regclass, 'pg_class');`,
				Results:   []sql.Row{{`test comment`}},
			},
			{
				Statement: `DROP TABLE testcomment;`,
			},
			{
				Statement: `CREATE TABLE concur_clustered(i int);`,
			},
			{
				Statement: `CREATE INDEX concur_clustered_i_idx ON concur_clustered(i);`,
			},
			{
				Statement: `ALTER TABLE concur_clustered CLUSTER ON concur_clustered_i_idx;`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_clustered;`,
			},
			{
				Statement: `SELECT indexrelid::regclass, indisclustered FROM pg_index
  WHERE indrelid = 'concur_clustered'::regclass;`,
				Results: []sql.Row{{`concur_clustered_i_idx`, true}},
			},
			{
				Statement: `DROP TABLE concur_clustered;`,
			},
			{
				Statement: `CREATE TABLE concur_replident(i int NOT NULL);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX concur_replident_i_idx ON concur_replident(i);`,
			},
			{
				Statement: `ALTER TABLE concur_replident REPLICA IDENTITY
  USING INDEX concur_replident_i_idx;`,
			},
			{
				Statement: `SELECT indexrelid::regclass, indisreplident FROM pg_index
  WHERE indrelid = 'concur_replident'::regclass;`,
				Results: []sql.Row{{`concur_replident_i_idx`, true}},
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_replident;`,
			},
			{
				Statement: `SELECT indexrelid::regclass, indisreplident FROM pg_index
  WHERE indrelid = 'concur_replident'::regclass;`,
				Results: []sql.Row{{`concur_replident_i_idx`, true}},
			},
			{
				Statement: `DROP TABLE concur_replident;`,
			},
			{
				Statement: `CREATE TABLE concur_appclass_tab(i tsvector, j tsvector, k tsvector);`,
			},
			{
				Statement: `CREATE INDEX concur_appclass_ind on concur_appclass_tab
  USING gist (i tsvector_ops (siglen='1000'), j tsvector_ops (siglen='500'));`,
			},
			{
				Statement: `CREATE INDEX concur_appclass_ind_2 on concur_appclass_tab
  USING gist (k tsvector_ops (siglen='300'), j tsvector_ops);`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_appclass_tab;`,
			},
			{
				Statement: `\d concur_appclass_tab
         Table "public.concur_appclass_tab"
 Column |   Type   | Collation | Nullable | Default 
--------+----------+-----------+----------+---------
 i      | tsvector |           |          | 
 j      | tsvector |           |          | 
 k      | tsvector |           |          | 
Indexes:
    "concur_appclass_ind" gist (i tsvector_ops (siglen='1000'), j tsvector_ops (siglen='500'))
    "concur_appclass_ind_2" gist (k tsvector_ops (siglen='300'), j)
DROP TABLE concur_appclass_tab;`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_part (c1 int, c2 int) PARTITION BY RANGE (c1);`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_part_0 PARTITION OF concur_reindex_part
  FOR VALUES FROM (0) TO (10) PARTITION BY list (c2);`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_part_0_1 PARTITION OF concur_reindex_part_0
  FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_part_0_2 PARTITION OF concur_reindex_part_0
  FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_part_10 PARTITION OF concur_reindex_part
  FOR VALUES FROM (10) TO (20) PARTITION BY list (c2);`,
			},
			{
				Statement: `CREATE INDEX concur_reindex_part_index ON ONLY concur_reindex_part (c1);`,
			},
			{
				Statement: `CREATE INDEX concur_reindex_part_index_0 ON ONLY concur_reindex_part_0 (c1);`,
			},
			{
				Statement: `ALTER INDEX concur_reindex_part_index ATTACH PARTITION concur_reindex_part_index_0;`,
			},
			{
				Statement: `CREATE INDEX concur_reindex_part_index_10 ON ONLY concur_reindex_part_10 (c1);`,
			},
			{
				Statement: `ALTER INDEX concur_reindex_part_index ATTACH PARTITION concur_reindex_part_index_10;`,
			},
			{
				Statement: `CREATE INDEX concur_reindex_part_index_0_1 ON ONLY concur_reindex_part_0_1 (c1);`,
			},
			{
				Statement: `ALTER INDEX concur_reindex_part_index_0 ATTACH PARTITION concur_reindex_part_index_0_1;`,
			},
			{
				Statement: `CREATE INDEX concur_reindex_part_index_0_2 ON ONLY concur_reindex_part_0_2 (c1);`,
			},
			{
				Statement: `ALTER INDEX concur_reindex_part_index_0 ATTACH PARTITION concur_reindex_part_index_0_2;`,
			},
			{
				Statement: `SELECT relid, parentrelid, level FROM pg_partition_tree('concur_reindex_part_index')
  ORDER BY relid, level;`,
				Results: []sql.Row{{`concur_reindex_part_index`, ``, 0}, {`concur_reindex_part_index_0`, `concur_reindex_part_index`, 1}, {`concur_reindex_part_index_10`, `concur_reindex_part_index`, 1}, {`concur_reindex_part_index_0_1`, `concur_reindex_part_index_0`, 2}, {`concur_reindex_part_index_0_2`, `concur_reindex_part_index_0`, 2}},
			},
			{
				Statement: `SELECT relid, parentrelid, level FROM pg_partition_tree('concur_reindex_part_index')
  ORDER BY relid, level;`,
				Results: []sql.Row{{`concur_reindex_part_index`, ``, 0}, {`concur_reindex_part_index_0`, `concur_reindex_part_index`, 1}, {`concur_reindex_part_index_10`, `concur_reindex_part_index`, 1}, {`concur_reindex_part_index_0_1`, `concur_reindex_part_index_0`, 2}, {`concur_reindex_part_index_0_2`, `concur_reindex_part_index_0`, 2}},
			},
			{
				Statement: `SELECT pg_describe_object(classid, objid, objsubid) as obj,
       pg_describe_object(refclassid,refobjid,refobjsubid) as objref,
       deptype
FROM pg_depend
WHERE classid = 'pg_class'::regclass AND
  objid in ('concur_reindex_part'::regclass,
            'concur_reindex_part_0'::regclass,
            'concur_reindex_part_0_1'::regclass,
            'concur_reindex_part_0_2'::regclass,
            'concur_reindex_part_index'::regclass,
            'concur_reindex_part_index_0'::regclass,
            'concur_reindex_part_index_0_1'::regclass,
            'concur_reindex_part_index_0_2'::regclass)
  ORDER BY 1, 2;`,
				Results: []sql.Row{{`column c1 of table concur_reindex_part`, `table concur_reindex_part`, `i`}, {`column c2 of table concur_reindex_part_0`, `table concur_reindex_part_0`, `i`}, {`index concur_reindex_part_index`, `column c1 of table concur_reindex_part`, `a`}, {`index concur_reindex_part_index_0`, `column c1 of table concur_reindex_part_0`, `a`}, {`index concur_reindex_part_index_0`, `index concur_reindex_part_index`, `P`}, {`index concur_reindex_part_index_0`, `table concur_reindex_part_0`, `S`}, {`index concur_reindex_part_index_0_1`, `column c1 of table concur_reindex_part_0_1`, `a`}, {`index concur_reindex_part_index_0_1`, `index concur_reindex_part_index_0`, `P`}, {`index concur_reindex_part_index_0_1`, `table concur_reindex_part_0_1`, `S`}, {`index concur_reindex_part_index_0_2`, `column c1 of table concur_reindex_part_0_2`, `a`}, {`index concur_reindex_part_index_0_2`, `index concur_reindex_part_index_0`, `P`}, {`index concur_reindex_part_index_0_2`, `table concur_reindex_part_0_2`, `S`}, {`table concur_reindex_part`, `schema public`, `n`}, {`table concur_reindex_part_0`, `schema public`, `n`}, {`table concur_reindex_part_0`, `table concur_reindex_part`, `a`}, {`table concur_reindex_part_0_1`, `schema public`, `n`}, {`table concur_reindex_part_0_1`, `table concur_reindex_part_0`, `a`}, {`table concur_reindex_part_0_2`, `schema public`, `n`}, {`table concur_reindex_part_0_2`, `table concur_reindex_part_0`, `a`}},
			},
			{
				Statement: `REINDEX INDEX CONCURRENTLY concur_reindex_part_index_0_1;`,
			},
			{
				Statement: `REINDEX INDEX CONCURRENTLY concur_reindex_part_index_0_2;`,
			},
			{
				Statement: `SELECT relid, parentrelid, level FROM pg_partition_tree('concur_reindex_part_index')
  ORDER BY relid, level;`,
				Results: []sql.Row{{`concur_reindex_part_index`, ``, 0}, {`concur_reindex_part_index_0`, `concur_reindex_part_index`, 1}, {`concur_reindex_part_index_10`, `concur_reindex_part_index`, 1}, {`concur_reindex_part_index_0_1`, `concur_reindex_part_index_0`, 2}, {`concur_reindex_part_index_0_2`, `concur_reindex_part_index_0`, 2}},
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_reindex_part_0_1;`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_reindex_part_0_2;`,
			},
			{
				Statement: `SELECT pg_describe_object(classid, objid, objsubid) as obj,
       pg_describe_object(refclassid,refobjid,refobjsubid) as objref,
       deptype
FROM pg_depend
WHERE classid = 'pg_class'::regclass AND
  objid in ('concur_reindex_part'::regclass,
            'concur_reindex_part_0'::regclass,
            'concur_reindex_part_0_1'::regclass,
            'concur_reindex_part_0_2'::regclass,
            'concur_reindex_part_index'::regclass,
            'concur_reindex_part_index_0'::regclass,
            'concur_reindex_part_index_0_1'::regclass,
            'concur_reindex_part_index_0_2'::regclass)
  ORDER BY 1, 2;`,
				Results: []sql.Row{{`column c1 of table concur_reindex_part`, `table concur_reindex_part`, `i`}, {`column c2 of table concur_reindex_part_0`, `table concur_reindex_part_0`, `i`}, {`index concur_reindex_part_index`, `column c1 of table concur_reindex_part`, `a`}, {`index concur_reindex_part_index_0`, `column c1 of table concur_reindex_part_0`, `a`}, {`index concur_reindex_part_index_0`, `index concur_reindex_part_index`, `P`}, {`index concur_reindex_part_index_0`, `table concur_reindex_part_0`, `S`}, {`index concur_reindex_part_index_0_1`, `column c1 of table concur_reindex_part_0_1`, `a`}, {`index concur_reindex_part_index_0_1`, `index concur_reindex_part_index_0`, `P`}, {`index concur_reindex_part_index_0_1`, `table concur_reindex_part_0_1`, `S`}, {`index concur_reindex_part_index_0_2`, `column c1 of table concur_reindex_part_0_2`, `a`}, {`index concur_reindex_part_index_0_2`, `index concur_reindex_part_index_0`, `P`}, {`index concur_reindex_part_index_0_2`, `table concur_reindex_part_0_2`, `S`}, {`table concur_reindex_part`, `schema public`, `n`}, {`table concur_reindex_part_0`, `schema public`, `n`}, {`table concur_reindex_part_0`, `table concur_reindex_part`, `a`}, {`table concur_reindex_part_0_1`, `schema public`, `n`}, {`table concur_reindex_part_0_1`, `table concur_reindex_part_0`, `a`}, {`table concur_reindex_part_0_2`, `schema public`, `n`}, {`table concur_reindex_part_0_2`, `table concur_reindex_part_0`, `a`}},
			},
			{
				Statement: `SELECT relid, parentrelid, level FROM pg_partition_tree('concur_reindex_part_index')
  ORDER BY relid, level;`,
				Results: []sql.Row{{`concur_reindex_part_index`, ``, 0}, {`concur_reindex_part_index_0`, `concur_reindex_part_index`, 1}, {`concur_reindex_part_index_10`, `concur_reindex_part_index`, 1}, {`concur_reindex_part_index_0_1`, `concur_reindex_part_index_0`, 2}, {`concur_reindex_part_index_0_2`, `concur_reindex_part_index_0`, 2}},
			},
			{
				Statement:   `REINDEX TABLE concur_reindex_part_index; -- error`,
				ErrorString: `"concur_reindex_part_index" is not a table or materialized view`,
			},
			{
				Statement:   `REINDEX TABLE CONCURRENTLY concur_reindex_part_index; -- error`,
				ErrorString: `"concur_reindex_part_index" is not a table or materialized view`,
			},
			{
				Statement:   `REINDEX TABLE concur_reindex_part_index_10; -- error`,
				ErrorString: `"concur_reindex_part_index_10" is not a table or materialized view`,
			},
			{
				Statement:   `REINDEX TABLE CONCURRENTLY concur_reindex_part_index_10; -- error`,
				ErrorString: `"concur_reindex_part_index_10" is not a table or materialized view`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `REINDEX INDEX concur_reindex_part_index;`,
				ErrorString: `REINDEX INDEX cannot run inside a transaction block`,
			},
			{
				Statement: `CONTEXT:  while reindexing partitioned index "public.concur_reindex_part_index"
ROLLBACK;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION create_relfilenode_part(relname text, indname text)
  RETURNS VOID AS
  $func$
  BEGIN
  EXECUTE format('
    CREATE TABLE %I AS
      SELECT oid, relname, relfilenode, relkind, reltoastrelid
      FROM pg_class
      WHERE oid IN
         (SELECT relid FROM pg_partition_tree(''%I''));',
	 relname, indname);`,
			},
			{
				Statement: `  END
  $func$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION compare_relfilenode_part(tabname text)
  RETURNS TABLE (relname name, relkind "char", state text) AS
  $func$
  BEGIN
    RETURN QUERY EXECUTE
      format(
        'SELECT  b.relname,
                 b.relkind,
                 CASE WHEN a.relfilenode = b.relfilenode THEN ''relfilenode is unchanged''
                 ELSE ''relfilenode has changed'' END
           -- Do not join with OID here as CONCURRENTLY changes it.
           FROM %I b JOIN pg_class a ON b.relname = a.relname
           ORDER BY 1;', tabname);`,
			},
			{
				Statement: `  END
  $func$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT create_relfilenode_part('reindex_index_status', 'concur_reindex_part_index');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `REINDEX INDEX concur_reindex_part_index;`,
			},
			{
				Statement: `SELECT * FROM compare_relfilenode_part('reindex_index_status');`,
				Results:   []sql.Row{{`concur_reindex_part_index`, `I`, `relfilenode is unchanged`}, {`concur_reindex_part_index_0`, `I`, `relfilenode is unchanged`}, {`concur_reindex_part_index_0_1`, `i`, `relfilenode has changed`}, {`concur_reindex_part_index_0_2`, `i`, `relfilenode has changed`}, {`concur_reindex_part_index_10`, `I`, `relfilenode is unchanged`}},
			},
			{
				Statement: `DROP TABLE reindex_index_status;`,
			},
			{
				Statement: `SELECT create_relfilenode_part('reindex_index_status', 'concur_reindex_part_index');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `REINDEX INDEX CONCURRENTLY concur_reindex_part_index;`,
			},
			{
				Statement: `SELECT * FROM compare_relfilenode_part('reindex_index_status');`,
				Results:   []sql.Row{{`concur_reindex_part_index`, `I`, `relfilenode is unchanged`}, {`concur_reindex_part_index_0`, `I`, `relfilenode is unchanged`}, {`concur_reindex_part_index_0_1`, `i`, `relfilenode has changed`}, {`concur_reindex_part_index_0_2`, `i`, `relfilenode has changed`}, {`concur_reindex_part_index_10`, `I`, `relfilenode is unchanged`}},
			},
			{
				Statement: `DROP TABLE reindex_index_status;`,
			},
			{
				Statement:   `REINDEX INDEX concur_reindex_part; -- error`,
				ErrorString: `"concur_reindex_part" is not an index`,
			},
			{
				Statement:   `REINDEX INDEX CONCURRENTLY concur_reindex_part; -- error`,
				ErrorString: `"concur_reindex_part" is not an index`,
			},
			{
				Statement:   `REINDEX INDEX concur_reindex_part_10; -- error`,
				ErrorString: `"concur_reindex_part_10" is not an index`,
			},
			{
				Statement:   `REINDEX INDEX CONCURRENTLY concur_reindex_part_10; -- error`,
				ErrorString: `"concur_reindex_part_10" is not an index`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `REINDEX TABLE concur_reindex_part;`,
				ErrorString: `REINDEX TABLE cannot run inside a transaction block`,
			},
			{
				Statement: `CONTEXT:  while reindexing partitioned table "public.concur_reindex_part"
ROLLBACK;`,
			},
			{
				Statement: `SELECT create_relfilenode_part('reindex_index_status', 'concur_reindex_part_index');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `REINDEX TABLE concur_reindex_part;`,
			},
			{
				Statement: `SELECT * FROM compare_relfilenode_part('reindex_index_status');`,
				Results:   []sql.Row{{`concur_reindex_part_index`, `I`, `relfilenode is unchanged`}, {`concur_reindex_part_index_0`, `I`, `relfilenode is unchanged`}, {`concur_reindex_part_index_0_1`, `i`, `relfilenode has changed`}, {`concur_reindex_part_index_0_2`, `i`, `relfilenode has changed`}, {`concur_reindex_part_index_10`, `I`, `relfilenode is unchanged`}},
			},
			{
				Statement: `DROP TABLE reindex_index_status;`,
			},
			{
				Statement: `SELECT create_relfilenode_part('reindex_index_status', 'concur_reindex_part_index');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_reindex_part;`,
			},
			{
				Statement: `SELECT * FROM compare_relfilenode_part('reindex_index_status');`,
				Results:   []sql.Row{{`concur_reindex_part_index`, `I`, `relfilenode is unchanged`}, {`concur_reindex_part_index_0`, `I`, `relfilenode is unchanged`}, {`concur_reindex_part_index_0_1`, `i`, `relfilenode has changed`}, {`concur_reindex_part_index_0_2`, `i`, `relfilenode has changed`}, {`concur_reindex_part_index_10`, `I`, `relfilenode is unchanged`}},
			},
			{
				Statement: `DROP TABLE reindex_index_status;`,
			},
			{
				Statement: `DROP FUNCTION create_relfilenode_part;`,
			},
			{
				Statement: `DROP FUNCTION compare_relfilenode_part;`,
			},
			{
				Statement: `DROP TABLE concur_reindex_part;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `REINDEX TABLE CONCURRENTLY concur_reindex_tab;`,
				ErrorString: `REINDEX CONCURRENTLY cannot run inside a transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement:   `REINDEX TABLE CONCURRENTLY pg_class; -- no catalog relation`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement:   `REINDEX INDEX CONCURRENTLY pg_class_oid_index; -- no catalog index`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement:   `REINDEX TABLE CONCURRENTLY pg_toast.pg_toast_1260; -- no catalog toast table`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement:   `REINDEX INDEX CONCURRENTLY pg_toast.pg_toast_1260_index; -- no catalog toast index`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement:   `REINDEX SYSTEM CONCURRENTLY postgres; -- not allowed for SYSTEM`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement: `REINDEX SCHEMA CONCURRENTLY pg_catalog;`,
			},
			{
				Statement: `\d concur_reindex_tab
         Table "public.concur_reindex_tab"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           | not null | 
 c2     | text    |           |          | 
Indexes:
    "concur_reindex_ind1" PRIMARY KEY, btree (c1)
    "concur_reindex_ind2" btree (c2)
    "concur_reindex_ind3" UNIQUE, btree (abs(c1))
    "concur_reindex_ind4" btree (c1, c1, c2)
Referenced by:
    TABLE "concur_reindex_tab2" CONSTRAINT "concur_reindex_tab2_c1_fkey" FOREIGN KEY (c1) REFERENCES concur_reindex_tab(c1)
DROP MATERIALIZED VIEW concur_reindex_matview;`,
			},
			{
				Statement: `DROP TABLE concur_reindex_tab, concur_reindex_tab2, concur_reindex_tab3;`,
			},
			{
				Statement: `CREATE TABLE concur_reindex_tab4 (c1 int);`,
			},
			{
				Statement: `INSERT INTO concur_reindex_tab4 VALUES (1), (1), (2);`,
			},
			{
				Statement:   `CREATE UNIQUE INDEX CONCURRENTLY concur_reindex_ind5 ON concur_reindex_tab4 (c1);`,
				ErrorString: `could not create unique index "concur_reindex_ind5"`,
			},
			{
				Statement: `DETAIL:  Key (c1)=(1) is duplicated.
REINDEX INDEX CONCURRENTLY concur_reindex_ind5;`,
				ErrorString: `could not create unique index "concur_reindex_ind5_ccnew"`,
			},
			{
				Statement: `DETAIL:  Key (c1)=(1) is duplicated.
\d concur_reindex_tab4
        Table "public.concur_reindex_tab4"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
Indexes:
    "concur_reindex_ind5" UNIQUE, btree (c1) INVALID
    "concur_reindex_ind5_ccnew" UNIQUE, btree (c1) INVALID
DROP INDEX concur_reindex_ind5_ccnew;`,
			},
			{
				Statement: `DELETE FROM concur_reindex_tab4 WHERE c1 = 1;`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_reindex_tab4;`,
			},
			{
				Statement: `\d concur_reindex_tab4
        Table "public.concur_reindex_tab4"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
Indexes:
    "concur_reindex_ind5" UNIQUE, btree (c1) INVALID
REINDEX INDEX CONCURRENTLY concur_reindex_ind5;`,
			},
			{
				Statement: `\d concur_reindex_tab4
        Table "public.concur_reindex_tab4"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
Indexes:
    "concur_reindex_ind5" UNIQUE, btree (c1)
DROP TABLE concur_reindex_tab4;`,
			},
			{
				Statement: `CREATE TABLE concur_exprs_tab (c1 int , c2 boolean);`,
			},
			{
				Statement: `INSERT INTO concur_exprs_tab (c1, c2) VALUES (1369652450, FALSE),
  (414515746, TRUE),
  (897778963, FALSE);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX concur_exprs_index_expr
  ON concur_exprs_tab ((c1::text COLLATE "C"));`,
			},
			{
				Statement: `CREATE UNIQUE INDEX concur_exprs_index_pred ON concur_exprs_tab (c1)
  WHERE (c1::text > 500000000::text COLLATE "C");`,
			},
			{
				Statement: `CREATE UNIQUE INDEX concur_exprs_index_pred_2
  ON concur_exprs_tab ((1 / c1))
  WHERE ('-H') >= (c2::TEXT) COLLATE "C";`,
			},
			{
				Statement: `ALTER INDEX concur_exprs_index_expr ALTER COLUMN 1 SET STATISTICS 100;`,
			},
			{
				Statement: `ANALYZE concur_exprs_tab;`,
			},
			{
				Statement: `SELECT starelid::regclass, count(*) FROM pg_statistic WHERE starelid IN (
  'concur_exprs_index_expr'::regclass,
  'concur_exprs_index_pred'::regclass,
  'concur_exprs_index_pred_2'::regclass)
  GROUP BY starelid ORDER BY starelid::regclass::text;`,
				Results: []sql.Row{{`concur_exprs_index_expr`, 1}},
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_expr'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_expr ON public.concur_exprs_tab USING btree (((c1)::text) COLLATE "C")`}},
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_pred'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_pred ON public.concur_exprs_tab USING btree (c1) WHERE ((c1)::text > ((500000000)::text COLLATE "C"))`}},
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_pred_2'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_pred_2 ON public.concur_exprs_tab USING btree (((1 / c1))) WHERE ('-H'::text >= ((c2)::text COLLATE "C"))`}},
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_exprs_tab;`,
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_expr'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_expr ON public.concur_exprs_tab USING btree (((c1)::text) COLLATE "C")`}},
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_pred'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_pred ON public.concur_exprs_tab USING btree (c1) WHERE ((c1)::text > ((500000000)::text COLLATE "C"))`}},
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_pred_2'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_pred_2 ON public.concur_exprs_tab USING btree (((1 / c1))) WHERE ('-H'::text >= ((c2)::text COLLATE "C"))`}},
			},
			{
				Statement: `ALTER TABLE concur_exprs_tab ALTER c2 TYPE TEXT;`,
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_expr'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_expr ON public.concur_exprs_tab USING btree (((c1)::text) COLLATE "C")`}},
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_pred'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_pred ON public.concur_exprs_tab USING btree (c1) WHERE ((c1)::text > ((500000000)::text COLLATE "C"))`}},
			},
			{
				Statement: `SELECT pg_get_indexdef('concur_exprs_index_pred_2'::regclass);`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX concur_exprs_index_pred_2 ON public.concur_exprs_tab USING btree (((1 / c1))) WHERE ('-H'::text >= (c2 COLLATE "C"))`}},
			},
			{
				Statement: `SELECT starelid::regclass, count(*) FROM pg_statistic WHERE starelid IN (
  'concur_exprs_index_expr'::regclass,
  'concur_exprs_index_pred'::regclass,
  'concur_exprs_index_pred_2'::regclass)
  GROUP BY starelid ORDER BY starelid::regclass::text;`,
				Results: []sql.Row{{`concur_exprs_index_expr`, 1}},
			},
			{
				Statement: `SELECT attrelid::regclass, attnum, attstattarget
  FROM pg_attribute WHERE attrelid IN (
    'concur_exprs_index_expr'::regclass,
    'concur_exprs_index_pred'::regclass,
    'concur_exprs_index_pred_2'::regclass)
  ORDER BY attrelid::regclass::text, attnum;`,
				Results: []sql.Row{{`concur_exprs_index_expr`, 1, 100}, {`concur_exprs_index_pred`, 1, -1}, {`concur_exprs_index_pred_2`, 1, -1}},
			},
			{
				Statement: `DROP TABLE concur_exprs_tab;`,
			},
			{
				Statement: `CREATE TEMP TABLE concur_temp_tab_1 (c1 int, c2 text)
  ON COMMIT PRESERVE ROWS;`,
			},
			{
				Statement: `INSERT INTO concur_temp_tab_1 VALUES (1, 'foo'), (2, 'bar');`,
			},
			{
				Statement: `CREATE INDEX concur_temp_ind_1 ON concur_temp_tab_1(c2);`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_temp_tab_1;`,
			},
			{
				Statement: `REINDEX INDEX CONCURRENTLY concur_temp_ind_1;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `REINDEX INDEX CONCURRENTLY concur_temp_ind_1;`,
				ErrorString: `REINDEX CONCURRENTLY cannot run inside a transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE TEMP TABLE concur_temp_tab_2 (c1 int, c2 text)
  ON COMMIT DELETE ROWS;`,
			},
			{
				Statement: `CREATE INDEX concur_temp_ind_2 ON concur_temp_tab_2(c2);`,
			},
			{
				Statement: `REINDEX TABLE CONCURRENTLY concur_temp_tab_2;`,
			},
			{
				Statement: `REINDEX INDEX CONCURRENTLY concur_temp_ind_2;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TEMP TABLE concur_temp_tab_3 (c1 int, c2 text)
  ON COMMIT PRESERVE ROWS;`,
			},
			{
				Statement: `INSERT INTO concur_temp_tab_3 VALUES (1, 'foo'), (2, 'bar');`,
			},
			{
				Statement: `CREATE INDEX concur_temp_ind_3 ON concur_temp_tab_3(c2);`,
			},
			{
				Statement:   `REINDEX INDEX CONCURRENTLY concur_temp_ind_3;`,
				ErrorString: `REINDEX CONCURRENTLY cannot run inside a transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE TABLE reindex_temp_before AS
SELECT oid, relname, relfilenode, relkind, reltoastrelid
  FROM pg_class
  WHERE relname IN ('concur_temp_ind_1', 'concur_temp_ind_2');`,
			},
			{
				Statement: `SELECT pg_my_temp_schema()::regnamespace as temp_schema_name \gset
REINDEX SCHEMA CONCURRENTLY :temp_schema_name;`,
			},
			{
				Statement: `SELECT  b.relname,
        b.relkind,
        CASE WHEN a.relfilenode = b.relfilenode THEN 'relfilenode is unchanged'
        ELSE 'relfilenode has changed' END
  FROM reindex_temp_before b JOIN pg_class a ON b.oid = a.oid
  ORDER BY 1;`,
				Results: []sql.Row{{`concur_temp_ind_1`, `i`, `relfilenode has changed`}, {`concur_temp_ind_2`, `i`, `relfilenode has changed`}},
			},
			{
				Statement: `DROP TABLE concur_temp_tab_1, concur_temp_tab_2, reindex_temp_before;`,
			},
			{
				Statement:   `REINDEX SCHEMA schema_to_reindex; -- failure, schema does not exist`,
				ErrorString: `schema "schema_to_reindex" does not exist`,
			},
			{
				Statement: `CREATE SCHEMA schema_to_reindex;`,
			},
			{
				Statement: `SET search_path = 'schema_to_reindex';`,
			},
			{
				Statement: `CREATE TABLE table1(col1 SERIAL PRIMARY KEY);`,
			},
			{
				Statement: `INSERT INTO table1 SELECT generate_series(1,400);`,
			},
			{
				Statement: `CREATE TABLE table2(col1 SERIAL PRIMARY KEY, col2 TEXT NOT NULL);`,
			},
			{
				Statement: `INSERT INTO table2 SELECT generate_series(1,400), 'abc';`,
			},
			{
				Statement: `CREATE INDEX ON table2(col2);`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW matview AS SELECT col1 FROM table2;`,
			},
			{
				Statement: `CREATE INDEX ON matview(col1);`,
			},
			{
				Statement: `CREATE VIEW view AS SELECT col2 FROM table2;`,
			},
			{
				Statement: `CREATE TABLE reindex_before AS
SELECT oid, relname, relfilenode, relkind, reltoastrelid
	FROM pg_class
	where relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'schema_to_reindex');`,
			},
			{
				Statement: `INSERT INTO reindex_before
SELECT oid, 'pg_toast_TABLE', relfilenode, relkind, reltoastrelid
FROM pg_class WHERE oid IN
	(SELECT reltoastrelid FROM reindex_before WHERE reltoastrelid > 0);`,
			},
			{
				Statement: `INSERT INTO reindex_before
SELECT oid, 'pg_toast_TABLE_index', relfilenode, relkind, reltoastrelid
FROM pg_class where oid in
	(select indexrelid from pg_index where indrelid in
		(select reltoastrelid from reindex_before where reltoastrelid > 0));`,
			},
			{
				Statement: `REINDEX SCHEMA schema_to_reindex;`,
			},
			{
				Statement: `CREATE TABLE reindex_after AS SELECT oid, relname, relfilenode, relkind
	FROM pg_class
	where relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'schema_to_reindex');`,
			},
			{
				Statement: `SELECT  b.relname,
        b.relkind,
        CASE WHEN a.relfilenode = b.relfilenode THEN 'relfilenode is unchanged'
        ELSE 'relfilenode has changed' END
  FROM reindex_before b JOIN pg_class a ON b.oid = a.oid
  ORDER BY 1;`,
				Results: []sql.Row{{`matview`, `m`, `relfilenode is unchanged`}, {`matview_col1_idx`, `i`, `relfilenode has changed`}, {`pg_toast_TABLE`, true, `relfilenode is unchanged`}, {`pg_toast_TABLE_index`, `i`, `relfilenode has changed`}, {`table1`, `r`, `relfilenode is unchanged`}, {`table1_col1_seq`, `S`, `relfilenode is unchanged`}, {`table1_pkey`, `i`, `relfilenode has changed`}, {`table2`, `r`, `relfilenode is unchanged`}, {`table2_col1_seq`, `S`, `relfilenode is unchanged`}, {`table2_col2_idx`, `i`, `relfilenode has changed`}, {`table2_pkey`, `i`, `relfilenode has changed`}, {`view`, `v`, `relfilenode is unchanged`}},
			},
			{
				Statement: `REINDEX SCHEMA schema_to_reindex;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `REINDEX SCHEMA schema_to_reindex; -- failure, cannot run in a transaction`,
				ErrorString: `REINDEX SCHEMA cannot run inside a transaction block`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `REINDEX SCHEMA CONCURRENTLY schema_to_reindex;`,
			},
			{
				Statement: `CREATE ROLE regress_reindexuser NOLOGIN;`,
			},
			{
				Statement: `SET SESSION ROLE regress_reindexuser;`,
			},
			{
				Statement:   `REINDEX SCHEMA schema_to_reindex;`,
				ErrorString: `must be owner of schema schema_to_reindex`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `GRANT USAGE ON SCHEMA pg_toast TO regress_reindexuser;`,
			},
			{
				Statement: `SET SESSION ROLE regress_reindexuser;`,
			},
			{
				Statement:   `REINDEX TABLE pg_toast.pg_toast_1260;`,
				ErrorString: `must be owner of table pg_toast_1260`,
			},
			{
				Statement:   `REINDEX INDEX pg_toast.pg_toast_1260_index;`,
				ErrorString: `must be owner of index pg_toast_1260_index`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `REVOKE USAGE ON SCHEMA pg_toast FROM regress_reindexuser;`,
			},
			{
				Statement: `DROP ROLE regress_reindexuser;`,
			},
			{
				Statement: `DROP SCHEMA schema_to_reindex CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to table table1
drop cascades to table table2
drop cascades to materialized view matview
drop cascades to view view
drop cascades to table reindex_before
drop cascades to table reindex_after`,
			},
		},
	})
}
