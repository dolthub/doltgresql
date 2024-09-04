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

func TestMatview(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_matview)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_matview,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE mvtest_t (id int NOT NULL PRIMARY KEY, type text NOT NULL, amt numeric NOT NULL);`,
			},
			{
				Statement: `INSERT INTO mvtest_t VALUES
  (1, 'x', 2),
  (2, 'x', 3),
  (3, 'y', 5),
  (4, 'y', 7),
  (5, 'z', 11);`,
			},
			{
				Statement: `CREATE VIEW mvtest_tv AS SELECT type, sum(amt) AS totamt FROM mvtest_t GROUP BY type;`,
			},
			{
				Statement: `SELECT * FROM mvtest_tv ORDER BY type;`,
				Results:   []sql.Row{{`x`, 5}, {`y`, 12}, {`z`, 11}},
			},
			{
				Statement: `EXPLAIN (costs off)
  CREATE MATERIALIZED VIEW mvtest_tm AS SELECT type, sum(amt) AS totamt FROM mvtest_t GROUP BY type WITH NO DATA;`,
				Results: []sql.Row{{`HashAggregate`}, {`Group Key: type`}, {`->  Seq Scan on mvtest_t`}},
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_tm AS SELECT type, sum(amt) AS totamt FROM mvtest_t GROUP BY type WITH NO DATA;`,
			},
			{
				Statement: `SELECT relispopulated FROM pg_class WHERE oid = 'mvtest_tm'::regclass;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `SELECT * FROM mvtest_tm ORDER BY type;`,
				ErrorString: `materialized view "mvtest_tm" has not been populated`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_tm;`,
			},
			{
				Statement: `SELECT relispopulated FROM pg_class WHERE oid = 'mvtest_tm'::regclass;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `CREATE UNIQUE INDEX mvtest_tm_type ON mvtest_tm (type);`,
			},
			{
				Statement: `SELECT * FROM mvtest_tm ORDER BY type;`,
				Results:   []sql.Row{{`x`, 5}, {`y`, 12}, {`z`, 11}},
			},
			{
				Statement: `EXPLAIN (costs off)
  CREATE MATERIALIZED VIEW mvtest_tvm AS SELECT * FROM mvtest_tv ORDER BY type;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: mvtest_t.type`}, {`->  HashAggregate`}, {`Group Key: mvtest_t.type`}, {`->  Seq Scan on mvtest_t`}},
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_tvm AS SELECT * FROM mvtest_tv ORDER BY type;`,
			},
			{
				Statement: `SELECT * FROM mvtest_tvm;`,
				Results:   []sql.Row{{`x`, 5}, {`y`, 12}, {`z`, 11}},
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_tmm AS SELECT sum(totamt) AS grandtot FROM mvtest_tm;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_tvmm AS SELECT sum(totamt) AS grandtot FROM mvtest_tvm;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX mvtest_tvmm_expr ON mvtest_tvmm ((grandtot > 0));`,
			},
			{
				Statement: `CREATE UNIQUE INDEX mvtest_tvmm_pred ON mvtest_tvmm (grandtot) WHERE grandtot < 0;`,
			},
			{
				Statement: `CREATE VIEW mvtest_tvv AS SELECT sum(totamt) AS grandtot FROM mvtest_tv;`,
			},
			{
				Statement: `EXPLAIN (costs off)
  CREATE MATERIALIZED VIEW mvtest_tvvm AS SELECT * FROM mvtest_tvv;`,
				Results: []sql.Row{{`Aggregate`}, {`->  HashAggregate`}, {`Group Key: mvtest_t.type`}, {`->  Seq Scan on mvtest_t`}},
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_tvvm AS SELECT * FROM mvtest_tvv;`,
			},
			{
				Statement: `CREATE VIEW mvtest_tvvmv AS SELECT * FROM mvtest_tvvm;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_bb AS SELECT * FROM mvtest_tvvmv;`,
			},
			{
				Statement: `CREATE INDEX mvtest_aa ON mvtest_bb (grandtot);`,
			},
			{
				Statement: `\d+ mvtest_tvm
                           Materialized view "public.mvtest_tvm"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 type   | text    |           |          |         | extended |              | 
 totamt | numeric |           |          |         | main     |              | 
View definition:
 SELECT mvtest_tv.type,
    mvtest_tv.totamt
   FROM mvtest_tv
  ORDER BY mvtest_tv.type;`,
			},
			{
				Statement: `\d+ mvtest_tvm
                           Materialized view "public.mvtest_tvm"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 type   | text    |           |          |         | extended |              | 
 totamt | numeric |           |          |         | main     |              | 
View definition:
 SELECT mvtest_tv.type,
    mvtest_tv.totamt
   FROM mvtest_tv
  ORDER BY mvtest_tv.type;`,
			},
			{
				Statement: `\d+ mvtest_tvvm
                           Materialized view "public.mvtest_tvvm"
  Column  |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
----------+---------+-----------+----------+---------+---------+--------------+-------------
 grandtot | numeric |           |          |         | main    |              | 
View definition:
 SELECT mvtest_tvv.grandtot
   FROM mvtest_tvv;`,
			},
			{
				Statement: `\d+ mvtest_bb
                            Materialized view "public.mvtest_bb"
  Column  |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
----------+---------+-----------+----------+---------+---------+--------------+-------------
 grandtot | numeric |           |          |         | main    |              | 
Indexes:
    "mvtest_aa" btree (grandtot)
View definition:
 SELECT mvtest_tvvmv.grandtot
   FROM mvtest_tvvmv;`,
			},
			{
				Statement: `CREATE SCHEMA mvtest_mvschema;`,
			},
			{
				Statement: `ALTER MATERIALIZED VIEW mvtest_tvm SET SCHEMA mvtest_mvschema;`,
			},
			{
				Statement: `\d+ mvtest_tvm
\d+ mvtest_tvmm
                           Materialized view "public.mvtest_tvmm"
  Column  |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
----------+---------+-----------+----------+---------+---------+--------------+-------------
 grandtot | numeric |           |          |         | main    |              | 
Indexes:
    "mvtest_tvmm_expr" UNIQUE, btree ((grandtot > 0::numeric))
    "mvtest_tvmm_pred" UNIQUE, btree (grandtot) WHERE grandtot < 0::numeric
View definition:
 SELECT sum(mvtest_tvm.totamt) AS grandtot
   FROM mvtest_mvschema.mvtest_tvm;`,
			},
			{
				Statement: `SET search_path = mvtest_mvschema, public;`,
			},
			{
				Statement: `\d+ mvtest_tvm
                      Materialized view "mvtest_mvschema.mvtest_tvm"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 type   | text    |           |          |         | extended |              | 
 totamt | numeric |           |          |         | main     |              | 
View definition:
 SELECT mvtest_tv.type,
    mvtest_tv.totamt
   FROM mvtest_tv
  ORDER BY mvtest_tv.type;`,
			},
			{
				Statement: `INSERT INTO mvtest_t VALUES (6, 'z', 13);`,
			},
			{
				Statement: `SELECT * FROM mvtest_tm ORDER BY type;`,
				Results:   []sql.Row{{`x`, 5}, {`y`, 12}, {`z`, 11}},
			},
			{
				Statement: `SELECT * FROM mvtest_tvm ORDER BY type;`,
				Results:   []sql.Row{{`x`, 5}, {`y`, 12}, {`z`, 11}},
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW CONCURRENTLY mvtest_tm;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_tvm;`,
			},
			{
				Statement: `SELECT * FROM mvtest_tm ORDER BY type;`,
				Results:   []sql.Row{{`x`, 5}, {`y`, 12}, {`z`, 24}},
			},
			{
				Statement: `SELECT * FROM mvtest_tvm ORDER BY type;`,
				Results:   []sql.Row{{`x`, 5}, {`y`, 12}, {`z`, 24}},
			},
			{
				Statement: `RESET search_path;`,
			},
			{
				Statement: `EXPLAIN (costs off)
  SELECT * FROM mvtest_tmm;`,
				Results: []sql.Row{{`Seq Scan on mvtest_tmm`}},
			},
			{
				Statement: `EXPLAIN (costs off)
  SELECT * FROM mvtest_tvmm;`,
				Results: []sql.Row{{`Seq Scan on mvtest_tvmm`}},
			},
			{
				Statement: `EXPLAIN (costs off)
  SELECT * FROM mvtest_tvvm;`,
				Results: []sql.Row{{`Seq Scan on mvtest_tvvm`}},
			},
			{
				Statement: `SELECT * FROM mvtest_tmm;`,
				Results:   []sql.Row{{28}},
			},
			{
				Statement: `SELECT * FROM mvtest_tvmm;`,
				Results:   []sql.Row{{28}},
			},
			{
				Statement: `SELECT * FROM mvtest_tvvm;`,
				Results:   []sql.Row{{28}},
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_tmm;`,
			},
			{
				Statement:   `REFRESH MATERIALIZED VIEW CONCURRENTLY mvtest_tvmm;`,
				ErrorString: `cannot refresh materialized view "public.mvtest_tvmm" concurrently`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_tvmm;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_tvvm;`,
			},
			{
				Statement: `EXPLAIN (costs off)
  SELECT * FROM mvtest_tmm;`,
				Results: []sql.Row{{`Seq Scan on mvtest_tmm`}},
			},
			{
				Statement: `EXPLAIN (costs off)
  SELECT * FROM mvtest_tvmm;`,
				Results: []sql.Row{{`Seq Scan on mvtest_tvmm`}},
			},
			{
				Statement: `EXPLAIN (costs off)
  SELECT * FROM mvtest_tvvm;`,
				Results: []sql.Row{{`Seq Scan on mvtest_tvvm`}},
			},
			{
				Statement: `SELECT * FROM mvtest_tmm;`,
				Results:   []sql.Row{{41}},
			},
			{
				Statement: `SELECT * FROM mvtest_tvmm;`,
				Results:   []sql.Row{{41}},
			},
			{
				Statement: `SELECT * FROM mvtest_tvvm;`,
				Results:   []sql.Row{{41}},
			},
			{
				Statement: `DROP MATERIALIZED VIEW IF EXISTS no_such_mv;`,
			},
			{
				Statement:   `REFRESH MATERIALIZED VIEW CONCURRENTLY mvtest_tvmm WITH NO DATA;`,
				ErrorString: `CONCURRENTLY and WITH NO DATA options cannot be used together`,
			},
			{
				Statement:   `SELECT * FROM mvtest_tvvm FOR SHARE;`,
				ErrorString: `cannot lock rows in materialized view "mvtest_tvvm"`,
			},
			{
				Statement: `SELECT type, m.totamt AS mtot, v.totamt AS vtot FROM mvtest_tm m LEFT JOIN mvtest_tv v USING (type) ORDER BY type;`,
				Results:   []sql.Row{{`x`, 5, 5}, {`y`, 12, 12}, {`z`, 24, 24}},
			},
			{
				Statement:   `DROP TABLE mvtest_t;`,
				ErrorString: `cannot drop table mvtest_t because other objects depend on it`,
			},
			{
				Statement: `materialized view mvtest_mvschema.mvtest_tvm depends on view mvtest_tv
materialized view mvtest_tvmm depends on materialized view mvtest_mvschema.mvtest_tvm
view mvtest_tvv depends on view mvtest_tv
materialized view mvtest_tvvm depends on view mvtest_tvv
view mvtest_tvvmv depends on materialized view mvtest_tvvm
materialized view mvtest_bb depends on view mvtest_tvvmv
materialized view mvtest_tm depends on table mvtest_t
materialized view mvtest_tmm depends on materialized view mvtest_tm
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
BEGIN;`,
			},
			{
				Statement: `DROP TABLE mvtest_t CASCADE;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE VIEW mvtest_vt1 AS SELECT 1 moo;`,
			},
			{
				Statement: `CREATE VIEW mvtest_vt2 AS SELECT moo, 2*moo FROM mvtest_vt1 UNION ALL SELECT moo, 3*moo FROM mvtest_vt1;`,
			},
			{
				Statement: `\d+ mvtest_vt2
                          View "public.mvtest_vt2"
  Column  |  Type   | Collation | Nullable | Default | Storage | Description 
----------+---------+-----------+----------+---------+---------+-------------
 moo      | integer |           |          |         | plain   | 
 ?column? | integer |           |          |         | plain   | 
View definition:
 SELECT mvtest_vt1.moo,
    2 * mvtest_vt1.moo AS "?column?"
   FROM mvtest_vt1
UNION ALL
 SELECT mvtest_vt1.moo,
    3 * mvtest_vt1.moo
   FROM mvtest_vt1;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mv_test2 AS SELECT moo, 2*moo FROM mvtest_vt2 UNION ALL SELECT moo, 3*moo FROM mvtest_vt2;`,
			},
			{
				Statement: `\d+ mv_test2
                            Materialized view "public.mv_test2"
  Column  |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
----------+---------+-----------+----------+---------+---------+--------------+-------------
 moo      | integer |           |          |         | plain   |              | 
 ?column? | integer |           |          |         | plain   |              | 
View definition:
 SELECT mvtest_vt2.moo,
    2 * mvtest_vt2.moo AS "?column?"
   FROM mvtest_vt2
UNION ALL
 SELECT mvtest_vt2.moo,
    3 * mvtest_vt2.moo
   FROM mvtest_vt2;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mv_test3 AS SELECT * FROM mv_test2 WHERE moo = 12345;`,
			},
			{
				Statement: `SELECT relispopulated FROM pg_class WHERE oid = 'mv_test3'::regclass;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP VIEW mvtest_vt1 CASCADE;`,
			},
			{
				Statement: `CREATE TABLE mvtest_foo(a, b) AS VALUES(1, 10);`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv AS SELECT * FROM mvtest_foo;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX ON mvtest_mv(a);`,
			},
			{
				Statement: `INSERT INTO mvtest_foo SELECT * FROM mvtest_foo;`,
			},
			{
				Statement:   `REFRESH MATERIALIZED VIEW mvtest_mv;`,
				ErrorString: `could not create unique index "mvtest_mv_a_idx"`,
			},
			{
				Statement:   `REFRESH MATERIALIZED VIEW CONCURRENTLY mvtest_mv;`,
				ErrorString: `new data for materialized view "mvtest_mv" contains duplicate rows without any null columns`,
			},
			{
				Statement: `DROP TABLE mvtest_foo CASCADE;`,
			},
			{
				Statement: `CREATE TABLE mvtest_foo(a, b, c) AS VALUES(1, 2, 3);`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv AS SELECT * FROM mvtest_foo;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX ON mvtest_mv (a);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX ON mvtest_mv (b);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX on mvtest_mv (c);`,
			},
			{
				Statement: `INSERT INTO mvtest_foo VALUES(2, 3, 4);`,
			},
			{
				Statement: `INSERT INTO mvtest_foo VALUES(3, 4, 5);`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_mv;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW CONCURRENTLY mvtest_mv;`,
			},
			{
				Statement: `DROP TABLE mvtest_foo CASCADE;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv1 AS SELECT 1 AS col1 WITH NO DATA;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv2 AS SELECT * FROM mvtest_mv1
  WHERE col1 = (SELECT LEAST(col1) FROM mvtest_mv1) WITH NO DATA;`,
			},
			{
				Statement: `DROP MATERIALIZED VIEW mvtest_mv1 CASCADE;`,
			},
			{
				Statement: `CREATE TABLE mvtest_boxes (id serial primary key, b box);`,
			},
			{
				Statement: `INSERT INTO mvtest_boxes (b) VALUES
  ('(32,32),(31,31)'),
  ('(2.0000004,2.0000004),(1,1)'),
  ('(1.9999996,1.9999996),(1,1)');`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_boxmv AS SELECT * FROM mvtest_boxes;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX mvtest_boxmv_id ON mvtest_boxmv (id);`,
			},
			{
				Statement: `UPDATE mvtest_boxes SET b = '(2,2),(1,1)' WHERE id = 2;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW CONCURRENTLY mvtest_boxmv;`,
			},
			{
				Statement: `SELECT * FROM mvtest_boxmv ORDER BY id;`,
				Results:   []sql.Row{{1, `(32,32),(31,31)`}, {2, `(2,2),(1,1)`}, {3, `(1.9999996,1.9999996),(1,1)`}},
			},
			{
				Statement: `DROP TABLE mvtest_boxes CASCADE;`,
			},
			{
				Statement: `CREATE TABLE mvtest_v (i int, j int);`,
			},
			{
				Statement:   `CREATE MATERIALIZED VIEW mvtest_mv_v (ii, jj, kk) AS SELECT i, j FROM mvtest_v; -- error`,
				ErrorString: `too many column names were specified`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv_v (ii, jj) AS SELECT i, j FROM mvtest_v; -- ok`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv_v_2 (ii) AS SELECT i, j FROM mvtest_v; -- ok`,
			},
			{
				Statement:   `CREATE MATERIALIZED VIEW mvtest_mv_v_3 (ii, jj, kk) AS SELECT i, j FROM mvtest_v WITH NO DATA; -- error`,
				ErrorString: `too many column names were specified`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv_v_3 (ii, jj) AS SELECT i, j FROM mvtest_v WITH NO DATA; -- ok`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv_v_4 (ii) AS SELECT i, j FROM mvtest_v WITH NO DATA; -- ok`,
			},
			{
				Statement: `ALTER TABLE mvtest_v RENAME COLUMN i TO x;`,
			},
			{
				Statement: `INSERT INTO mvtest_v values (1, 2);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX mvtest_mv_v_ii ON mvtest_mv_v (ii);`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_mv_v;`,
			},
			{
				Statement: `UPDATE mvtest_v SET j = 3 WHERE x = 1;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW CONCURRENTLY mvtest_mv_v;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_mv_v_2;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_mv_v_3;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_mv_v_4;`,
			},
			{
				Statement: `SELECT * FROM mvtest_v;`,
				Results:   []sql.Row{{1, 3}},
			},
			{
				Statement: `SELECT * FROM mvtest_mv_v;`,
				Results:   []sql.Row{{1, 3}},
			},
			{
				Statement: `SELECT * FROM mvtest_mv_v_2;`,
				Results:   []sql.Row{{1, 3}},
			},
			{
				Statement: `SELECT * FROM mvtest_mv_v_3;`,
				Results:   []sql.Row{{1, 3}},
			},
			{
				Statement: `SELECT * FROM mvtest_mv_v_4;`,
				Results:   []sql.Row{{1, 3}},
			},
			{
				Statement: `DROP TABLE mvtest_v CASCADE;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mv_unspecified_types AS
  SELECT 42 as i, 42.5 as num, 'foo' as u, 'foo'::unknown as u2, null as n;`,
			},
			{
				Statement: `\d+ mv_unspecified_types
                      Materialized view "public.mv_unspecified_types"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 i      | integer |           |          |         | plain    |              | 
 num    | numeric |           |          |         | main     |              | 
 u      | text    |           |          |         | extended |              | 
 u2     | text    |           |          |         | extended |              | 
 n      | text    |           |          |         | extended |              | 
View definition:
 SELECT 42 AS i,
    42.5 AS num,
    'foo'::text AS u,
    'foo'::text AS u2,
    NULL::text AS n;`,
			},
			{
				Statement: `SELECT * FROM mv_unspecified_types;`,
				Results:   []sql.Row{{42, 42.5, `foo`, `foo`, ``}},
			},
			{
				Statement: `DROP MATERIALIZED VIEW mv_unspecified_types;`,
			},
			{
				Statement:   `create materialized view mvtest_error as select 1/0 as x;  -- fail`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `create materialized view mvtest_error as select 1/0 as x with no data;`,
			},
			{
				Statement:   `refresh materialized view mvtest_error;  -- fail here`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `drop materialized view mvtest_error;`,
			},
			{
				Statement: `CREATE TABLE mvtest_v AS SELECT generate_series(1,10) AS a;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv_v AS SELECT a FROM mvtest_v WHERE a <= 5;`,
			},
			{
				Statement: `DELETE FROM mvtest_v WHERE EXISTS ( SELECT * FROM mvtest_mv_v WHERE mvtest_mv_v.a = mvtest_v.a );`,
			},
			{
				Statement: `SELECT * FROM mvtest_v;`,
				Results:   []sql.Row{{6}, {7}, {8}, {9}, {10}},
			},
			{
				Statement: `SELECT * FROM mvtest_mv_v;`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}, {5}},
			},
			{
				Statement: `DROP TABLE mvtest_v CASCADE;`,
			},
			{
				Statement: `CREATE ROLE regress_user_mvtest;`,
			},
			{
				Statement: `SET ROLE regress_user_mvtest;`,
			},
			{
				Statement: `CREATE TABLE mvtest_foo_data AS SELECT i,
  i+1 AS tid,
  md5(random()::text) AS mv,
  md5(random()::text) AS newdata,
  md5(random()::text) AS newdata2,
  md5(random()::text) AS diff
  FROM generate_series(1, 10) i;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW mvtest_mv_foo AS SELECT * FROM mvtest_foo_data;`,
			},
			{
				Statement:   `CREATE MATERIALIZED VIEW mvtest_mv_foo AS SELECT * FROM mvtest_foo_data;`,
				ErrorString: `relation "mvtest_mv_foo" already exists`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW IF NOT EXISTS mvtest_mv_foo AS SELECT * FROM mvtest_foo_data;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX ON mvtest_mv_foo (i);`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW mvtest_mv_foo;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW CONCURRENTLY mvtest_mv_foo;`,
			},
			{
				Statement: `DROP OWNED BY regress_user_mvtest CASCADE;`,
			},
			{
				Statement: `DROP ROLE regress_user_mvtest;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE FUNCTION mvtest_func()
  RETURNS void AS $$
BEGIN
  CREATE MATERIALIZED VIEW mvtest1 AS SELECT 1 AS x;`,
			},
			{
				Statement: `  CREATE MATERIALIZED VIEW mvtest2 AS SELECT 1 AS x WITH NO DATA;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT mvtest_func();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT * FROM mvtest1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT * FROM mvtest2;`,
				ErrorString: `materialized view "mvtest2" has not been populated`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE SCHEMA matview_schema;`,
			},
			{
				Statement: `CREATE USER regress_matview_user;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_matview_user
  REVOKE INSERT ON TABLES FROM regress_matview_user;`,
			},
			{
				Statement: `GRANT ALL ON SCHEMA matview_schema TO public;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_matview_user;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW matview_schema.mv_withdata1 (a) AS
  SELECT generate_series(1, 10) WITH DATA;`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE MATERIALIZED VIEW matview_schema.mv_withdata2 (a) AS
  SELECT generate_series(1, 10) WITH DATA;`,
				Results: []sql.Row{{`ProjectSet (actual rows=10 loops=1)`}, {`->  Result (actual rows=1 loops=1)`}},
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW matview_schema.mv_withdata2;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW matview_schema.mv_nodata1 (a) AS
  SELECT generate_series(1, 10) WITH NO DATA;`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE MATERIALIZED VIEW matview_schema.mv_nodata2 (a) AS
  SELECT generate_series(1, 10) WITH NO DATA;`,
				Results: []sql.Row{{`ProjectSet (never executed)`}, {`->  Result (never executed)`}},
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW matview_schema.mv_nodata2;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_matview_user
  GRANT INSERT ON TABLES TO regress_matview_user;`,
			},
			{
				Statement: `DROP SCHEMA matview_schema CASCADE;`,
			},
			{
				Statement: `DROP USER regress_matview_user;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW matview_ine_tab AS SELECT 1;`,
			},
			{
				Statement:   `CREATE MATERIALIZED VIEW matview_ine_tab AS SELECT 1 / 0; -- error`,
				ErrorString: `relation "matview_ine_tab" already exists`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW IF NOT EXISTS matview_ine_tab AS
  SELECT 1 / 0; -- ok`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW matview_ine_tab AS
  SELECT 1 / 0 WITH NO DATA; -- error`,
				ErrorString: `relation "matview_ine_tab" already exists`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW IF NOT EXISTS matview_ine_tab AS
  SELECT 1 / 0 WITH NO DATA; -- ok`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE MATERIALIZED VIEW matview_ine_tab AS
    SELECT 1 / 0; -- error`,
				ErrorString: `relation "matview_ine_tab" already exists`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE MATERIALIZED VIEW IF NOT EXISTS matview_ine_tab AS
    SELECT 1 / 0; -- ok`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE MATERIALIZED VIEW matview_ine_tab AS
    SELECT 1 / 0 WITH NO DATA; -- error`,
				ErrorString: `relation "matview_ine_tab" already exists`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE MATERIALIZED VIEW IF NOT EXISTS matview_ine_tab AS
    SELECT 1 / 0 WITH NO DATA; -- ok`,
				Results: []sql.Row{},
			},
			{
				Statement: `DROP MATERIALIZED VIEW matview_ine_tab;`,
			},
		},
	})
}
