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

func TestSubselect(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_subselect)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_subselect,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT 1 AS one WHERE 1 IN (SELECT 1);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT 1 AS zero WHERE 1 NOT IN (SELECT 1);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT 1 AS zero WHERE 1 IN (SELECT 2);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM (SELECT 1 AS x) ss;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM ((SELECT 1 AS x)) ss;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `(SELECT 2) UNION SELECT 2;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `((SELECT 2)) UNION SELECT 2;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT ((SELECT 2) UNION SELECT 2);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT (((SELECT 2)) UNION SELECT 2);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT (SELECT ARRAY[1,2,3])[1];`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT ((SELECT ARRAY[1,2,3]))[2];`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT (((SELECT ARRAY[1,2,3])))[3];`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `CREATE TABLE SUBSELECT_TBL (
  f1 integer,
  f2 integer,
  f3 float
);`,
			},
			{
				Statement: `INSERT INTO SUBSELECT_TBL VALUES (1, 2, 3);`,
			},
			{
				Statement: `INSERT INTO SUBSELECT_TBL VALUES (2, 3, 4);`,
			},
			{
				Statement: `INSERT INTO SUBSELECT_TBL VALUES (3, 4, 5);`,
			},
			{
				Statement: `INSERT INTO SUBSELECT_TBL VALUES (1, 1, 1);`,
			},
			{
				Statement: `INSERT INTO SUBSELECT_TBL VALUES (2, 2, 2);`,
			},
			{
				Statement: `INSERT INTO SUBSELECT_TBL VALUES (3, 3, 3);`,
			},
			{
				Statement: `INSERT INTO SUBSELECT_TBL VALUES (6, 7, 8);`,
			},
			{
				Statement: `INSERT INTO SUBSELECT_TBL VALUES (8, 9, NULL);`,
			},
			{
				Statement: `SELECT * FROM SUBSELECT_TBL;`,
				Results:   []sql.Row{{1, 2, 3}, {2, 3, 4}, {3, 4, 5}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {6, 7, 8}, {8, 9, ``}},
			},
			{
				Statement: `SELECT f1 AS "Constant Select" FROM SUBSELECT_TBL
  WHERE f1 IN (SELECT 1);`,
				Results: []sql.Row{{1}, {1}},
			},
			{
				Statement: `SELECT f1 AS "Uncorrelated Field" FROM SUBSELECT_TBL
  WHERE f1 IN (SELECT f2 FROM SUBSELECT_TBL);`,
				Results: []sql.Row{{1}, {2}, {3}, {1}, {2}, {3}},
			},
			{
				Statement: `SELECT f1 AS "Uncorrelated Field" FROM SUBSELECT_TBL
  WHERE f1 IN (SELECT f2 FROM SUBSELECT_TBL WHERE
    f2 IN (SELECT f1 FROM SUBSELECT_TBL));`,
				Results: []sql.Row{{1}, {2}, {3}, {1}, {2}, {3}},
			},
			{
				Statement: `SELECT f1, f2
  FROM SUBSELECT_TBL
  WHERE (f1, f2) NOT IN (SELECT f2, CAST(f3 AS int4) FROM SUBSELECT_TBL
                         WHERE f3 IS NOT NULL);`,
				Results: []sql.Row{{1, 2}, {6, 7}, {8, 9}},
			},
			{
				Statement: `SELECT f1 AS "Correlated Field", f2 AS "Second Field"
  FROM SUBSELECT_TBL upper
  WHERE f1 IN (SELECT f2 FROM SUBSELECT_TBL WHERE f1 = upper.f1);`,
				Results: []sql.Row{{1, 2}, {2, 3}, {3, 4}, {1, 1}, {2, 2}, {3, 3}},
			},
			{
				Statement: `SELECT f1 AS "Correlated Field", f3 AS "Second Field"
  FROM SUBSELECT_TBL upper
  WHERE f1 IN
    (SELECT f2 FROM SUBSELECT_TBL WHERE CAST(upper.f2 AS float) = f3);`,
				Results: []sql.Row{{2, 4}, {3, 5}, {1, 1}, {2, 2}, {3, 3}},
			},
			{
				Statement: `SELECT f1 AS "Correlated Field", f3 AS "Second Field"
  FROM SUBSELECT_TBL upper
  WHERE f3 IN (SELECT upper.f1 + f2 FROM SUBSELECT_TBL
               WHERE f2 = CAST(f3 AS integer));`,
				Results: []sql.Row{{1, 3}, {2, 4}, {3, 5}, {6, 8}},
			},
			{
				Statement: `SELECT f1 AS "Correlated Field"
  FROM SUBSELECT_TBL
  WHERE (f1, f2) IN (SELECT f2, CAST(f3 AS int4) FROM SUBSELECT_TBL
                     WHERE f3 IS NOT NULL);`,
				Results: []sql.Row{{2}, {3}, {1}, {2}, {3}},
			},
			{
				Statement: `SELECT ss.f1 AS "Correlated Field", ss.f3 AS "Second Field"
  FROM SUBSELECT_TBL ss
  WHERE f1 NOT IN (SELECT f1+1 FROM INT4_TBL
                   WHERE f1 != ss.f1 AND f1 < 2147483647);`,
				Results: []sql.Row{{2, 4}, {3, 5}, {2, 2}, {3, 3}, {6, 8}, {8, ``}},
			},
			{
				Statement: `select q1, float8(count(*)) / (select count(*) from int8_tbl)
from int8_tbl group by q1 order by q1;`,
				Results: []sql.Row{{123, 0.4}, {4567890123456789, 0.6}},
			},
			{
				Statement: `SELECT *, pg_typeof(f1) FROM
  (SELECT 'foo' AS f1 FROM generate_series(1,3)) ss ORDER BY 1;`,
				Results: []sql.Row{{`foo`, `text`}, {`foo`, `text`}, {`foo`, `text`}},
			},
			{
				Statement: `explain (verbose, costs off) select '42' union all select '43';`,
				Results:   []sql.Row{{`Append`}, {`->  Result`}, {`Output: '42'::text`}, {`->  Result`}, {`Output: '43'::text`}},
			},
			{
				Statement: `explain (verbose, costs off) select '42' union all select 43;`,
				Results:   []sql.Row{{`Append`}, {`->  Result`}, {`Output: 42`}, {`->  Result`}, {`Output: 43`}},
			},
			{
				Statement: `explain (verbose, costs off)
select 1 = all (select (select 1));`,
				Results: []sql.Row{{`Result`}, {`Output: (SubPlan 2)`}, {`SubPlan 2`}, {`->  Materialize`}, {`Output: ($0)`}, {`InitPlan 1 (returns $0)`}, {`->  Result`}, {`Output: 1`}, {`->  Result`}, {`Output: $0`}},
			},
			{
				Statement: `select 1 = all (select (select 1));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `explain (costs off)
select * from int4_tbl o where exists
  (select 1 from int4_tbl i where i.f1=o.f1 limit null);`,
				Results: []sql.Row{{`Hash Semi Join`}, {`Hash Cond: (o.f1 = i.f1)`}, {`->  Seq Scan on int4_tbl o`}, {`->  Hash`}, {`->  Seq Scan on int4_tbl i`}},
			},
			{
				Statement: `explain (costs off)
select * from int4_tbl o where not exists
  (select 1 from int4_tbl i where i.f1=o.f1 limit 1);`,
				Results: []sql.Row{{`Hash Anti Join`}, {`Hash Cond: (o.f1 = i.f1)`}, {`->  Seq Scan on int4_tbl o`}, {`->  Hash`}, {`->  Seq Scan on int4_tbl i`}},
			},
			{
				Statement: `explain (costs off)
select * from int4_tbl o where exists
  (select 1 from int4_tbl i where i.f1=o.f1 limit 0);`,
				Results: []sql.Row{{`Seq Scan on int4_tbl o`}, {`Filter: (SubPlan 1)`}, {`SubPlan 1`}, {`->  Limit`}, {`->  Seq Scan on int4_tbl i`}, {`Filter: (f1 = o.f1)`}},
			},
			{
				Statement: `select count(*) from
  (select 1 from tenk1 a
   where unique1 IN (select hundred from tenk1 b)) ss;`,
				Results: []sql.Row{{100}},
			},
			{
				Statement: `select count(distinct ss.ten) from
  (select ten from tenk1 a
   where unique1 IN (select hundred from tenk1 b)) ss;`,
				Results: []sql.Row{{10}},
			},
			{
				Statement: `select count(*) from
  (select 1 from tenk1 a
   where unique1 IN (select distinct hundred from tenk1 b)) ss;`,
				Results: []sql.Row{{100}},
			},
			{
				Statement: `select count(distinct ss.ten) from
  (select ten from tenk1 a
   where unique1 IN (select distinct hundred from tenk1 b)) ss;`,
				Results: []sql.Row{{10}},
			},
			{
				Statement: `CREATE TEMP TABLE foo (id integer);`,
			},
			{
				Statement: `CREATE TEMP TABLE bar (id1 integer, id2 integer);`,
			},
			{
				Statement: `INSERT INTO foo VALUES (1);`,
			},
			{
				Statement: `INSERT INTO bar VALUES (1, 1);`,
			},
			{
				Statement: `INSERT INTO bar VALUES (2, 2);`,
			},
			{
				Statement: `INSERT INTO bar VALUES (3, 1);`,
			},
			{
				Statement: `SELECT * FROM foo WHERE id IN
    (SELECT id2 FROM (SELECT DISTINCT id1, id2 FROM bar) AS s);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM foo WHERE id IN
    (SELECT id2 FROM (SELECT id1,id2 FROM bar GROUP BY id1,id2) AS s);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM foo WHERE id IN
    (SELECT id2 FROM (SELECT id1, id2 FROM bar UNION
                      SELECT id1, id2 FROM bar) AS s);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM foo WHERE id IN
    (SELECT id2 FROM (SELECT DISTINCT ON (id2) id1, id2 FROM bar) AS s);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM foo WHERE id IN
    (SELECT id2 FROM (SELECT id2 FROM bar GROUP BY id2) AS s);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM foo WHERE id IN
    (SELECT id2 FROM (SELECT id2 FROM bar UNION
                      SELECT id2 FROM bar) AS s);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `CREATE TABLE orderstest (
    approver_ref integer,
    po_ref integer,
    ordercanceled boolean
);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (1, 1, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (66, 5, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (66, 6, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (66, 7, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (66, 1, true);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (66, 8, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (66, 1, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (77, 1, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (1, 1, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (66, 1, false);`,
			},
			{
				Statement: `INSERT INTO orderstest VALUES (1, 1, false);`,
			},
			{
				Statement: `CREATE VIEW orders_view AS
SELECT *,
(SELECT CASE
   WHEN ord.approver_ref=1 THEN '---' ELSE 'Approved'
 END) AS "Approved",
(SELECT CASE
 WHEN ord.ordercanceled
 THEN 'Canceled'
 ELSE
  (SELECT CASE
		WHEN ord.po_ref=1
		THEN
		 (SELECT CASE
				WHEN ord.approver_ref=1
				THEN '---'
				ELSE 'Approved'
			END)
		ELSE 'PO'
	END)
END) AS "Status",
(CASE
 WHEN ord.ordercanceled
 THEN 'Canceled'
 ELSE
  (CASE
		WHEN ord.po_ref=1
		THEN
		 (CASE
				WHEN ord.approver_ref=1
				THEN '---'
				ELSE 'Approved'
			END)
		ELSE 'PO'
	END)
END) AS "Status_OK"
FROM orderstest ord;`,
			},
			{
				Statement: `SELECT * FROM orders_view;`,
				Results:   []sql.Row{{1, 1, false, `---`, `---`, `---`}, {66, 5, false, `Approved`, `PO`, `PO`}, {66, 6, false, `Approved`, `PO`, `PO`}, {66, 7, false, `Approved`, `PO`, `PO`}, {66, 1, true, `Approved`, `Canceled`, `Canceled`}, {66, 8, false, `Approved`, `PO`, `PO`}, {66, 1, false, `Approved`, `Approved`, `Approved`}, {77, 1, false, `Approved`, `Approved`, `Approved`}, {1, 1, false, `---`, `---`, `---`}, {66, 1, false, `Approved`, `Approved`, `Approved`}, {1, 1, false, `---`, `---`, `---`}},
			},
			{
				Statement: `DROP TABLE orderstest cascade;`,
			},
			{
				Statement: `create temp table parts (
    partnum     text,
    cost        float8
);`,
			},
			{
				Statement: `create temp table shipped (
    ttype       char(2),
    ordnum      int4,
    partnum     text,
    value       float8
);`,
			},
			{
				Statement: `create temp view shipped_view as
    select * from shipped where ttype = 'wt';`,
			},
			{
				Statement: `create rule shipped_view_insert as on insert to shipped_view do instead
    insert into shipped values('wt', new.ordnum, new.partnum, new.value);`,
			},
			{
				Statement: `insert into parts (partnum, cost) values (1, 1234.56);`,
			},
			{
				Statement: `insert into shipped_view (ordnum, partnum, value)
    values (0, 1, (select cost from parts where partnum = '1'));`,
			},
			{
				Statement: `select * from shipped_view;`,
				Results:   []sql.Row{{`wt`, 0, 1, 1234.56}},
			},
			{
				Statement: `create rule shipped_view_update as on update to shipped_view do instead
    update shipped set partnum = new.partnum, value = new.value
        where ttype = new.ttype and ordnum = new.ordnum;`,
			},
			{
				Statement: `update shipped_view set value = 11
    from int4_tbl a join int4_tbl b
      on (a.f1 = (select f1 from int4_tbl c where c.f1=b.f1))
    where ordnum = a.f1;`,
			},
			{
				Statement: `select * from shipped_view;`,
				Results:   []sql.Row{{`wt`, 0, 1, 11}},
			},
			{
				Statement: `select f1, ss1 as relabel from
    (select *, (select sum(f1) from int4_tbl b where f1 >= a.f1) as ss1
     from int4_tbl a) ss;`,
				Results: []sql.Row{{0, 2147607103}, {123456, 2147607103}, {-123456, 2147483647}, {2147483647, 2147483647}, {-2147483647, 0}},
			},
			{
				Statement: `select * from (
  select max(unique1) from tenk1 as a
  where exists (select 1 from tenk1 as b where b.thousand = a.unique2)
) ss;`,
				Results: []sql.Row{{9997}},
			},
			{
				Statement: `select * from (
  select min(unique1) from tenk1 as a
  where not exists (select 1 from tenk1 as b where b.unique2 = 10000)
) ss;`,
				Results: []sql.Row{{0}},
			},
			{
				Statement: `create temp table numeric_table (num_col numeric);`,
			},
			{
				Statement: `insert into numeric_table values (1), (1.000000000000000000001), (2), (3);`,
			},
			{
				Statement: `create temp table float_table (float_col float8);`,
			},
			{
				Statement: `insert into float_table values (1), (2), (3);`,
			},
			{
				Statement: `select * from float_table
  where float_col in (select num_col from numeric_table);`,
				Results: []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement: `select * from numeric_table
  where num_col in (select float_col from float_table);`,
				Results: []sql.Row{{1}, {1.000000000000000000001}, {2}, {3}},
			},
			{
				Statement: `create temp table ta (id int primary key, val int);`,
			},
			{
				Statement: `insert into ta values(1,1);`,
			},
			{
				Statement: `insert into ta values(2,2);`,
			},
			{
				Statement: `create temp table tb (id int primary key, aval int);`,
			},
			{
				Statement: `insert into tb values(1,1);`,
			},
			{
				Statement: `insert into tb values(2,1);`,
			},
			{
				Statement: `insert into tb values(3,2);`,
			},
			{
				Statement: `insert into tb values(4,2);`,
			},
			{
				Statement: `create temp table tc (id int primary key, aid int);`,
			},
			{
				Statement: `insert into tc values(1,1);`,
			},
			{
				Statement: `insert into tc values(2,2);`,
			},
			{
				Statement: `select
  ( select min(tb.id) from tb
    where tb.aval = (select ta.val from ta where ta.id = tc.aid) ) as min_tb_id
from tc;`,
				Results: []sql.Row{{1}, {3}},
			},
			{
				Statement: `create temp table t1 (f1 numeric(14,0), f2 varchar(30));`,
			},
			{
				Statement: `select * from
  (select distinct f1, f2, (select f2 from t1 x where x.f1 = up.f1) as fs
   from t1 up) ss
group by f1,f2,fs;`,
				Results: []sql.Row{},
			},
			{
				Statement: `create temp table table_a(id integer);`,
			},
			{
				Statement: `insert into table_a values (42);`,
			},
			{
				Statement: `create temp view view_a as select * from table_a;`,
			},
			{
				Statement: `select view_a from view_a;`,
				Results:   []sql.Row{{`(42)`}},
			},
			{
				Statement: `select (select view_a) from view_a;`,
				Results:   []sql.Row{{`(42)`}},
			},
			{
				Statement: `select (select (select view_a)) from view_a;`,
				Results:   []sql.Row{{`(42)`}},
			},
			{
				Statement: `select (select (a.*)::text) from view_a a;`,
				Results:   []sql.Row{{`(42)`}},
			},
			{
				Statement: `select q from (select max(f1) from int4_tbl group by f1 order by f1) q;`,
				Results:   []sql.Row{{`(-2147483647)`}, {`(-123456)`}, {`(0)`}, {`(123456)`}, {`(2147483647)`}},
			},
			{
				Statement: `with q as (select max(f1) from int4_tbl group by f1 order by f1)
  select q from q;`,
				Results: []sql.Row{{`(-2147483647)`}, {`(-123456)`}, {`(0)`}, {`(123456)`}, {`(2147483647)`}},
			},
			{
				Statement: `begin;  --  this shouldn't delete anything, but be safe`,
			},
			{
				Statement: `delete from road
where exists (
  select 1
  from
    int4_tbl cross join
    ( select f1, array(select q1 from int8_tbl) as arr
      from text_tbl ) ss
  where road.name = ss.f1 );`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `select
  (select sq1) as qq1
from
  (select exists(select 1 from int4_tbl where f1 = q2) as sq1, 42 as dummy
   from int8_tbl) sq0
  join
  int4_tbl i4 on dummy = i4.f1;`,
				Results: []sql.Row{},
			},
			{
				Statement: `create temp table upsert(key int4 primary key, val text);`,
			},
			{
				Statement: `insert into upsert values(1, 'val') on conflict (key) do update set val = 'not seen';`,
			},
			{
				Statement: `insert into upsert values(1, 'val') on conflict (key) do update set val = 'seen with subselect ' || (select f1 from int4_tbl where f1 != 0 limit 1)::text;`,
			},
			{
				Statement: `select * from upsert;`,
				Results:   []sql.Row{{1, `seen with subselect 123456`}},
			},
			{
				Statement: `with aa as (select 'int4_tbl' u from int4_tbl limit 1)
insert into upsert values (1, 'x'), (999, 'y')
on conflict (key) do update set val = (select u from aa)
returning *;`,
				Results: []sql.Row{{1, `int4_tbl`}, {999, `y`}},
			},
			{
				Statement: `create temp table outer_7597 (f1 int4, f2 int4);`,
			},
			{
				Statement: `insert into outer_7597 values (0, 0);`,
			},
			{
				Statement: `insert into outer_7597 values (1, 0);`,
			},
			{
				Statement: `insert into outer_7597 values (0, null);`,
			},
			{
				Statement: `insert into outer_7597 values (1, null);`,
			},
			{
				Statement: `create temp table inner_7597(c1 int8, c2 int8);`,
			},
			{
				Statement: `insert into inner_7597 values(0, null);`,
			},
			{
				Statement: `select * from outer_7597 where (f1, f2) not in (select * from inner_7597);`,
				Results:   []sql.Row{{1, 0}, {1, ``}},
			},
			{
				Statement: `create temp table outer_text (f1 text, f2 text);`,
			},
			{
				Statement: `insert into outer_text values ('a', 'a');`,
			},
			{
				Statement: `insert into outer_text values ('b', 'a');`,
			},
			{
				Statement: `insert into outer_text values ('a', null);`,
			},
			{
				Statement: `insert into outer_text values ('b', null);`,
			},
			{
				Statement: `create temp table inner_text (c1 text, c2 text);`,
			},
			{
				Statement: `insert into inner_text values ('a', null);`,
			},
			{
				Statement: `insert into inner_text values ('123', '456');`,
			},
			{
				Statement: `select * from outer_text where (f1, f2) not in (select * from inner_text);`,
				Results:   []sql.Row{{`b`, `a`}, {`b`, ``}},
			},
			{
				Statement: `explain (verbose, costs off)
select 'foo'::text in (select 'bar'::name union all select 'bar'::name);`,
				Results: []sql.Row{{`Result`}, {`Output: (hashed SubPlan 1)`}, {`SubPlan 1`}, {`->  Append`}, {`->  Result`}, {`Output: 'bar'::name`}, {`->  Result`}, {`Output: 'bar'::name`}},
			},
			{
				Statement: `select 'foo'::text in (select 'bar'::name union all select 'bar'::name);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `explain (verbose, costs off)
select row(row(row(1))) = any (select row(row(1)));`,
				Results: []sql.Row{{`Result`}, {`Output: (SubPlan 1)`}, {`SubPlan 1`}, {`->  Materialize`}, {`Output: '("(1)")'::record`}, {`->  Result`}, {`Output: '("(1)")'::record`}},
			},
			{
				Statement: `select row(row(row(1))) = any (select row(row(1)));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select '1'::text in (select '1'::name union all select '1'::name);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select * from int8_tbl where q1 in (select c1 from inner_text);`,
				ErrorString: `operator does not exist: bigint = text`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create function bogus_int8_text_eq(int8, text) returns boolean
language sql as 'select $1::text = $2';`,
			},
			{
				Statement: `create operator = (procedure=bogus_int8_text_eq, leftarg=int8, rightarg=text);`,
			},
			{
				Statement: `explain (costs off)
select * from int8_tbl where q1 in (select c1 from inner_text);`,
				Results: []sql.Row{{`Seq Scan on int8_tbl`}, {`Filter: (hashed SubPlan 1)`}, {`SubPlan 1`}, {`->  Seq Scan on inner_text`}},
			},
			{
				Statement: `select * from int8_tbl where q1 in (select c1 from inner_text);`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `create or replace function bogus_int8_text_eq(int8, text) returns boolean
language sql as 'select $1::text = $2 and $1::text = $2';`,
			},
			{
				Statement: `explain (costs off)
select * from int8_tbl where q1 in (select c1 from inner_text);`,
				Results: []sql.Row{{`Seq Scan on int8_tbl`}, {`Filter: (hashed SubPlan 1)`}, {`SubPlan 1`}, {`->  Seq Scan on inner_text`}},
			},
			{
				Statement: `select * from int8_tbl where q1 in (select c1 from inner_text);`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `create or replace function bogus_int8_text_eq(int8, text) returns boolean
language sql as 'select $2 = $1::text';`,
			},
			{
				Statement: `explain (costs off)
select * from int8_tbl where q1 in (select c1 from inner_text);`,
				Results: []sql.Row{{`Seq Scan on int8_tbl`}, {`Filter: (SubPlan 1)`}, {`SubPlan 1`}, {`->  Materialize`}, {`->  Seq Scan on inner_text`}},
			},
			{
				Statement: `select * from int8_tbl where q1 in (select c1 from inner_text);`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}},
			},
			{
				Statement: `rollback;  -- to get rid of the bogus operator`,
			},
			{
				Statement: `explain (costs off)
select count(*) from tenk1 t
where (exists(select 1 from tenk1 k where k.unique1 = t.unique2) or ten < 0);`,
				Results: []sql.Row{{`Aggregate`}, {`->  Seq Scan on tenk1 t`}, {`Filter: ((hashed SubPlan 2) OR (ten < 0))`}, {`SubPlan 2`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 k`}},
			},
			{
				Statement: `select count(*) from tenk1 t
where (exists(select 1 from tenk1 k where k.unique1 = t.unique2) or ten < 0);`,
				Results: []sql.Row{{10000}},
			},
			{
				Statement: `explain (costs off)
select count(*) from tenk1 t
where (exists(select 1 from tenk1 k where k.unique1 = t.unique2) or ten < 0)
  and thousand = 1;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Bitmap Heap Scan on tenk1 t`}, {`Recheck Cond: (thousand = 1)`}, {`Filter: ((SubPlan 1) OR (ten < 0))`}, {`->  Bitmap Index Scan on tenk1_thous_tenthous`}, {`Index Cond: (thousand = 1)`}, {`SubPlan 1`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 k`}, {`Index Cond: (unique1 = t.unique2)`}},
			},
			{
				Statement: `select count(*) from tenk1 t
where (exists(select 1 from tenk1 k where k.unique1 = t.unique2) or ten < 0)
  and thousand = 1;`,
				Results: []sql.Row{{10}},
			},
			{
				Statement: `create temp table exists_tbl (c1 int, c2 int, c3 int) partition by list (c1);`,
			},
			{
				Statement: `create temp table exists_tbl_null partition of exists_tbl for values in (null);`,
			},
			{
				Statement: `create temp table exists_tbl_def partition of exists_tbl default;`,
			},
			{
				Statement: `insert into exists_tbl select x, x/2, x+1 from generate_series(0,10) x;`,
			},
			{
				Statement: `analyze exists_tbl;`,
			},
			{
				Statement: `explain (costs off)
select * from exists_tbl t1
  where (exists(select 1 from exists_tbl t2 where t1.c1 = t2.c2) or c3 < 0);`,
				Results: []sql.Row{{`Append`}, {`->  Seq Scan on exists_tbl_null t1_1`}, {`Filter: ((SubPlan 1) OR (c3 < 0))`}, {`SubPlan 1`}, {`->  Append`}, {`->  Seq Scan on exists_tbl_null t2_1`}, {`Filter: (t1_1.c1 = c2)`}, {`->  Seq Scan on exists_tbl_def t2_2`}, {`Filter: (t1_1.c1 = c2)`}, {`->  Seq Scan on exists_tbl_def t1_2`}, {`Filter: ((hashed SubPlan 2) OR (c3 < 0))`}, {`SubPlan 2`}, {`->  Append`}, {`->  Seq Scan on exists_tbl_null t2_4`}, {`->  Seq Scan on exists_tbl_def t2_5`}},
			},
			{
				Statement: `select * from exists_tbl t1
  where (exists(select 1 from exists_tbl t2 where t1.c1 = t2.c2) or c3 < 0);`,
				Results: []sql.Row{{0, 0, 1}, {1, 0, 2}, {2, 1, 3}, {3, 1, 4}, {4, 2, 5}, {5, 2, 6}},
			},
			{
				Statement: `select a.thousand from tenk1 a, tenk1 b
where a.thousand = b.thousand
  and exists ( select 1 from tenk1 c where b.hundred = c.hundred
                   and not exists ( select 1 from tenk1 d
                                    where a.thousand = d.thousand ) );`,
				Results: []sql.Row{},
			},
			{
				Statement: `explain (verbose, costs off)
  select x, x from
    (select (select now()) as x from (values(1),(2)) v(y)) ss;`,
				Results: []sql.Row{{`Values Scan on "*VALUES*"`}, {`Output: $0, $1`}, {`InitPlan 1 (returns $0)`}, {`->  Result`}, {`Output: now()`}, {`InitPlan 2 (returns $1)`}, {`->  Result`}, {`Output: now()`}},
			},
			{
				Statement: `explain (verbose, costs off)
  select x, x from
    (select (select random()) as x from (values(1),(2)) v(y)) ss;`,
				Results: []sql.Row{{`Subquery Scan on ss`}, {`Output: ss.x, ss.x`}, {`->  Values Scan on "*VALUES*"`}, {`Output: $0`}, {`InitPlan 1 (returns $0)`}, {`->  Result`}, {`Output: random()`}},
			},
			{
				Statement: `explain (verbose, costs off)
  select x, x from
    (select (select now() where y=y) as x from (values(1),(2)) v(y)) ss;`,
				Results: []sql.Row{{`Values Scan on "*VALUES*"`}, {`Output: (SubPlan 1), (SubPlan 2)`}, {`SubPlan 1`}, {`->  Result`}, {`Output: now()`}, {`One-Time Filter: ("*VALUES*".column1 = "*VALUES*".column1)`}, {`SubPlan 2`}, {`->  Result`}, {`Output: now()`}, {`One-Time Filter: ("*VALUES*".column1 = "*VALUES*".column1)`}},
			},
			{
				Statement: `explain (verbose, costs off)
  select x, x from
    (select (select random() where y=y) as x from (values(1),(2)) v(y)) ss;`,
				Results: []sql.Row{{`Subquery Scan on ss`}, {`Output: ss.x, ss.x`}, {`->  Values Scan on "*VALUES*"`}, {`Output: (SubPlan 1)`}, {`SubPlan 1`}, {`->  Result`}, {`Output: random()`}, {`One-Time Filter: ("*VALUES*".column1 = "*VALUES*".column1)`}},
			},
			{
				Statement: `explain (verbose, costs off)
select sum(ss.tst::int) from
  onek o cross join lateral (
  select i.ten in (select f1 from int4_tbl where f1 <= o.hundred) as tst,
         random() as r
  from onek i where i.unique1 = o.unique1 ) ss
where o.ten = 0;`,
				Results: []sql.Row{{`Aggregate`}, {`Output: sum((((hashed SubPlan 1)))::integer)`}, {`->  Nested Loop`}, {`Output: ((hashed SubPlan 1))`}, {`->  Seq Scan on public.onek o`}, {`Output: o.unique1, o.unique2, o.two, o.four, o.ten, o.twenty, o.hundred, o.thousand, o.twothousand, o.fivethous, o.tenthous, o.odd, o.even, o.stringu1, o.stringu2, o.string4`}, {`Filter: (o.ten = 0)`}, {`->  Index Scan using onek_unique1 on public.onek i`}, {`Output: (hashed SubPlan 1), random()`}, {`Index Cond: (i.unique1 = o.unique1)`}, {`SubPlan 1`}, {`->  Seq Scan on public.int4_tbl`}, {`Output: int4_tbl.f1`}, {`Filter: (int4_tbl.f1 <= $0)`}},
			},
			{
				Statement: `select sum(ss.tst::int) from
  onek o cross join lateral (
  select i.ten in (select f1 from int4_tbl where f1 <= o.hundred) as tst,
         random() as r
  from onek i where i.unique1 = o.unique1 ) ss
where o.ten = 0;`,
				Results: []sql.Row{{100}},
			},
			{
				Statement: `explain (costs off)
select count(*) from
  onek o cross join lateral (
    select * from onek i1 where i1.unique1 = o.unique1
    except
    select * from onek i2 where i2.unique1 = o.unique2
  ) ss
where o.ten = 1;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Nested Loop`}, {`->  Seq Scan on onek o`}, {`Filter: (ten = 1)`}, {`->  Subquery Scan on ss`}, {`->  HashSetOp Except`}, {`->  Append`}, {`->  Subquery Scan on "*SELECT* 1"`}, {`->  Index Scan using onek_unique1 on onek i1`}, {`Index Cond: (unique1 = o.unique1)`}, {`->  Subquery Scan on "*SELECT* 2"`}, {`->  Index Scan using onek_unique1 on onek i2`}, {`Index Cond: (unique1 = o.unique2)`}},
			},
			{
				Statement: `select count(*) from
  onek o cross join lateral (
    select * from onek i1 where i1.unique1 = o.unique1
    except
    select * from onek i2 where i2.unique1 = o.unique2
  ) ss
where o.ten = 1;`,
				Results: []sql.Row{{100}},
			},
			{
				Statement: `explain (costs off)
select sum(o.four), sum(ss.a) from
  onek o cross join lateral (
    with recursive x(a) as
      (select o.four as a
       union
       select a + 1 from x
       where a < 10)
    select * from x
  ) ss
where o.ten = 1;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Nested Loop`}, {`->  Seq Scan on onek o`}, {`Filter: (ten = 1)`}, {`->  Memoize`}, {`Cache Key: o.four`}, {`Cache Mode: binary`}, {`->  CTE Scan on x`}, {`CTE x`}, {`->  Recursive Union`}, {`->  Result`}, {`->  WorkTable Scan on x x_1`}, {`Filter: (a < 10)`}},
			},
			{
				Statement: `select sum(o.four), sum(ss.a) from
  onek o cross join lateral (
    with recursive x(a) as
      (select o.four as a
       union
       select a + 1 from x
       where a < 10)
    select * from x
  ) ss
where o.ten = 1;`,
				Results: []sql.Row{{1700, 5350}},
			},
			{
				Statement: `create temp table notinouter (a int);`,
			},
			{
				Statement: `create temp table notininner (b int not null);`,
			},
			{
				Statement: `insert into notinouter values (null), (1);`,
			},
			{
				Statement: `select * from notinouter where a not in (select b from notininner);`,
				Results:   []sql.Row{{``}, {1}},
			},
			{
				Statement: `create temp table nocolumns();`,
			},
			{
				Statement: `select exists(select * from nocolumns);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select val.x
  from generate_series(1,10) as s(i),
  lateral (
    values ((select s.i + 1)), (s.i + 101)
  ) as val(x)
where s.i < 10 and (select val.x) < 110;`,
				Results: []sql.Row{{2}, {102}, {3}, {103}, {4}, {104}, {5}, {105}, {6}, {106}, {7}, {107}, {8}, {108}, {9}, {109}, {10}},
			},
			{
				Statement: `explain (verbose, costs off)
select * from
(values
  (3 not in (select * from (values (1), (2)) ss1)),
  (false)
) ss;`,
				Results: []sql.Row{{`Values Scan on "*VALUES*"`}, {`Output: "*VALUES*".column1`}, {`SubPlan 1`}, {`->  Values Scan on "*VALUES*_1"`}, {`Output: "*VALUES*_1".column1`}},
			},
			{
				Statement: `select * from
(values
  (3 not in (select * from (values (1), (2)) ss1)),
  (false)
) ss;`,
				Results: []sql.Row{{true}, {false}},
			},
			{
				Statement: `explain (verbose, costs off)
select * from int4_tbl where
  (case when f1 in (select unique1 from tenk1 a) then f1 else null end) in
  (select ten from tenk1 b);`,
				Results: []sql.Row{{`Nested Loop Semi Join`}, {`Output: int4_tbl.f1`}, {`Join Filter: (CASE WHEN (hashed SubPlan 1) THEN int4_tbl.f1 ELSE NULL::integer END = b.ten)`}, {`->  Seq Scan on public.int4_tbl`}, {`Output: int4_tbl.f1`}, {`->  Seq Scan on public.tenk1 b`}, {`Output: b.unique1, b.unique2, b.two, b.four, b.ten, b.twenty, b.hundred, b.thousand, b.twothousand, b.fivethous, b.tenthous, b.odd, b.even, b.stringu1, b.stringu2, b.string4`}, {`SubPlan 1`}, {`->  Index Only Scan using tenk1_unique1 on public.tenk1 a`}, {`Output: a.unique1`}},
			},
			{
				Statement: `select * from int4_tbl where
  (case when f1 in (select unique1 from tenk1 a) then f1 else null end) in
  (select ten from tenk1 b);`,
				Results: []sql.Row{{0}},
			},
			{
				Statement: `explain (verbose, costs off)
select * from int4_tbl o where (f1, f1) in
  (select f1, generate_series(1,50) / 10 g from int4_tbl i group by f1);`,
				Results: []sql.Row{{`Nested Loop Semi Join`}, {`Output: o.f1`}, {`Join Filter: (o.f1 = "ANY_subquery".f1)`}, {`->  Seq Scan on public.int4_tbl o`}, {`Output: o.f1`}, {`->  Materialize`}, {`Output: "ANY_subquery".f1, "ANY_subquery".g`}, {`->  Subquery Scan on "ANY_subquery"`}, {`Output: "ANY_subquery".f1, "ANY_subquery".g`}, {`Filter: ("ANY_subquery".f1 = "ANY_subquery".g)`}, {`->  Result`}, {`Output: i.f1, ((generate_series(1, 50)) / 10)`}, {`->  ProjectSet`}, {`Output: generate_series(1, 50), i.f1`}, {`->  HashAggregate`}, {`Output: i.f1`}, {`Group Key: i.f1`}, {`->  Seq Scan on public.int4_tbl i`}, {`Output: i.f1`}},
			},
			{
				Statement: `select * from int4_tbl o where (f1, f1) in
  (select f1, generate_series(1,50) / 10 g from int4_tbl i group by f1);`,
				Results: []sql.Row{{0}},
			},
			{
				Statement: `select (select q from
         (select 1,2,3 where f1 > 0
          union all
          select 4,5,6.0 where f1 <= 0
         ) q )
from int4_tbl;`,
				Results: []sql.Row{{`(4,5,6.0)`}, {`(1,2,3)`}, {`(4,5,6.0)`}, {`(1,2,3)`}, {`(4,5,6.0)`}},
			},
			{
				Statement: `explain (verbose, costs off)
select * from
    int4_tbl i4,
    lateral (
        select i4.f1 > 1 as b, 1 as id
        from (select random() order by 1) as t1
      union all
        select true as b, 2 as id
    ) as t2
where b and f1 >= 0;`,
				Results: []sql.Row{{`Nested Loop`}, {`Output: i4.f1, ((i4.f1 > 1)), (1)`}, {`->  Seq Scan on public.int4_tbl i4`}, {`Output: i4.f1`}, {`Filter: (i4.f1 >= 0)`}, {`->  Append`}, {`->  Subquery Scan on t1`}, {`Output: (i4.f1 > 1), 1`}, {`Filter: (i4.f1 > 1)`}, {`->  Sort`}, {`Output: (random())`}, {`Sort Key: (random())`}, {`->  Result`}, {`Output: random()`}, {`->  Result`}, {`Output: true, 2`}},
			},
			{
				Statement: `select * from
    int4_tbl i4,
    lateral (
        select i4.f1 > 1 as b, 1 as id
        from (select random() order by 1) as t1
      union all
        select true as b, 2 as id
    ) as t2
where b and f1 >= 0;`,
				Results: []sql.Row{{0, true, 2}, {123456, true, 1}, {123456, true, 2}, {2147483647, true, 1}, {2147483647, true, 2}},
			},
			{
				Statement: `create temp sequence ts1;`,
			},
			{
				Statement: `select * from
  (select distinct ten from tenk1) ss
  where ten < 10 + nextval('ts1')
  order by 1;`,
				Results: []sql.Row{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}},
			},
			{
				Statement: `select nextval('ts1');`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `create function tattle(x int, y int) returns bool
volatile language plpgsql as $$
begin
  raise notice 'x = %, y = %', x, y;`,
			},
			{
				Statement: `  return x > y;`,
			},
			{
				Statement: `end$$;`,
			},
			{
				Statement: `explain (verbose, costs off)
select * from
  (select 9 as x, unnest(array[1,2,3,11,12,13]) as u) ss
  where tattle(x, 8);`,
				Results: []sql.Row{{`Subquery Scan on ss`}, {`Output: ss.x, ss.u`}, {`Filter: tattle(ss.x, 8)`}, {`->  ProjectSet`}, {`Output: 9, unnest('{1,2,3,11,12,13}'::integer[])`}, {`->  Result`}},
			},
			{
				Statement: `select * from
  (select 9 as x, unnest(array[1,2,3,11,12,13]) as u) ss
  where tattle(x, 8);`,
				Results: []sql.Row{{9, 1}, {9, 2}, {9, 3}, {9, 11}, {9, 12}, {9, 13}},
			},
			{
				Statement: `alter function tattle(x int, y int) stable;`,
			},
			{
				Statement: `explain (verbose, costs off)
select * from
  (select 9 as x, unnest(array[1,2,3,11,12,13]) as u) ss
  where tattle(x, 8);`,
				Results: []sql.Row{{`ProjectSet`}, {`Output: 9, unnest('{1,2,3,11,12,13}'::integer[])`}, {`->  Result`}, {`One-Time Filter: tattle(9, 8)`}},
			},
			{
				Statement: `select * from
  (select 9 as x, unnest(array[1,2,3,11,12,13]) as u) ss
  where tattle(x, 8);`,
				Results: []sql.Row{{9, 1}, {9, 2}, {9, 3}, {9, 11}, {9, 12}, {9, 13}},
			},
			{
				Statement: `explain (verbose, costs off)
select * from
  (select 9 as x, unnest(array[1,2,3,11,12,13]) as u) ss
  where tattle(x, u);`,
				Results: []sql.Row{{`Subquery Scan on ss`}, {`Output: ss.x, ss.u`}, {`Filter: tattle(ss.x, ss.u)`}, {`->  ProjectSet`}, {`Output: 9, unnest('{1,2,3,11,12,13}'::integer[])`}, {`->  Result`}},
			},
			{
				Statement: `select * from
  (select 9 as x, unnest(array[1,2,3,11,12,13]) as u) ss
  where tattle(x, u);`,
				Results: []sql.Row{{9, 1}, {9, 2}, {9, 3}},
			},
			{
				Statement: `drop function tattle(x int, y int);`,
			},
			{
				Statement: `create table sq_limit (pk int primary key, c1 int, c2 int);`,
			},
			{
				Statement: `insert into sq_limit values
    (1, 1, 1),
    (2, 2, 2),
    (3, 3, 3),
    (4, 4, 4),
    (5, 1, 1),
    (6, 2, 2),
    (7, 3, 3),
    (8, 4, 4);`,
			},
			{
				Statement: `create function explain_sq_limit() returns setof text language plpgsql as
$$
declare ln text;`,
			},
			{
				Statement: `begin
    for ln in
        explain (analyze, summary off, timing off, costs off)
        select * from (select pk,c2 from sq_limit order by c1,pk) as x limit 3
    loop
        ln := regexp_replace(ln, 'Memory: \S*',  'Memory: xxx');`,
			},
			{
				Statement: `        return next ln;`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `select * from explain_sq_limit();`,
				Results:   []sql.Row{{`Limit (actual rows=3 loops=1)`}, {`->  Subquery Scan on x (actual rows=3 loops=1)`}, {`->  Sort (actual rows=3 loops=1)`}, {`Sort Key: sq_limit.c1, sq_limit.pk`}, {`Sort Method: top-N heapsort  Memory: xxx`}, {`->  Seq Scan on sq_limit (actual rows=8 loops=1)`}},
			},
			{
				Statement: `select * from (select pk,c2 from sq_limit order by c1,pk) as x limit 3;`,
				Results:   []sql.Row{{1, 1}, {5, 1}, {2, 2}},
			},
			{
				Statement: `drop function explain_sq_limit();`,
			},
			{
				Statement: `drop table sq_limit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `declare c1 scroll cursor for
 select * from generate_series(1,4) i
  where i <> all (values (2),(3));`,
			},
			{
				Statement: `move forward all in c1;`,
			},
			{
				Statement: `fetch backward all in c1;`,
				Results:   []sql.Row{{4}, {1}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `explain (verbose, costs off)
with x as (select * from (select f1 from subselect_tbl) ss)
select * from x where f1 = 1;`,
				Results: []sql.Row{{`Seq Scan on public.subselect_tbl`}, {`Output: subselect_tbl.f1`}, {`Filter: (subselect_tbl.f1 = 1)`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as materialized (select * from (select f1 from subselect_tbl) ss)
select * from x where f1 = 1;`,
				Results: []sql.Row{{`CTE Scan on x`}, {`Output: x.f1`}, {`Filter: (x.f1 = 1)`}, {`CTE x`}, {`->  Seq Scan on public.subselect_tbl`}, {`Output: subselect_tbl.f1`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as (select * from (select f1, now() from subselect_tbl) ss)
select * from x where f1 = 1;`,
				Results: []sql.Row{{`Seq Scan on public.subselect_tbl`}, {`Output: subselect_tbl.f1, now()`}, {`Filter: (subselect_tbl.f1 = 1)`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as (select * from (select f1, random() from subselect_tbl) ss)
select * from x where f1 = 1;`,
				Results: []sql.Row{{`CTE Scan on x`}, {`Output: x.f1, x.random`}, {`Filter: (x.f1 = 1)`}, {`CTE x`}, {`->  Seq Scan on public.subselect_tbl`}, {`Output: subselect_tbl.f1, random()`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as (select * from (select f1 from subselect_tbl for update) ss)
select * from x where f1 = 1;`,
				Results: []sql.Row{{`CTE Scan on x`}, {`Output: x.f1`}, {`Filter: (x.f1 = 1)`}, {`CTE x`}, {`->  Subquery Scan on ss`}, {`Output: ss.f1`}, {`->  LockRows`}, {`Output: subselect_tbl.f1, subselect_tbl.ctid`}, {`->  Seq Scan on public.subselect_tbl`}, {`Output: subselect_tbl.f1, subselect_tbl.ctid`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as (select * from (select f1, now() as n from subselect_tbl) ss)
select * from x, x x2 where x.n = x2.n;`,
				Results: []sql.Row{{`Merge Join`}, {`Output: x.f1, x.n, x2.f1, x2.n`}, {`Merge Cond: (x.n = x2.n)`}, {`CTE x`}, {`->  Seq Scan on public.subselect_tbl`}, {`Output: subselect_tbl.f1, now()`}, {`->  Sort`}, {`Output: x.f1, x.n`}, {`Sort Key: x.n`}, {`->  CTE Scan on x`}, {`Output: x.f1, x.n`}, {`->  Sort`}, {`Output: x2.f1, x2.n`}, {`Sort Key: x2.n`}, {`->  CTE Scan on x x2`}, {`Output: x2.f1, x2.n`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as not materialized (select * from (select f1, now() as n from subselect_tbl) ss)
select * from x, x x2 where x.n = x2.n;`,
				Results: []sql.Row{{`Result`}, {`Output: subselect_tbl.f1, now(), subselect_tbl_1.f1, now()`}, {`One-Time Filter: (now() = now())`}, {`->  Nested Loop`}, {`Output: subselect_tbl.f1, subselect_tbl_1.f1`}, {`->  Seq Scan on public.subselect_tbl`}, {`Output: subselect_tbl.f1, subselect_tbl.f2, subselect_tbl.f3`}, {`->  Materialize`}, {`Output: subselect_tbl_1.f1`}, {`->  Seq Scan on public.subselect_tbl subselect_tbl_1`}, {`Output: subselect_tbl_1.f1`}},
			},
			{
				Statement: `explain (verbose, costs off)
with recursive x(a) as
  ((values ('a'), ('b'))
   union all
   (with z as not materialized (select * from x)
    select z.a || z1.a as a from z cross join z as z1
    where length(z.a || z1.a) < 5))
select * from x;`,
				Results: []sql.Row{{`CTE Scan on x`}, {`Output: x.a`}, {`CTE x`}, {`->  Recursive Union`}, {`->  Values Scan on "*VALUES*"`}, {`Output: "*VALUES*".column1`}, {`->  Nested Loop`}, {`Output: (z.a || z1.a)`}, {`Join Filter: (length((z.a || z1.a)) < 5)`}, {`CTE z`}, {`->  WorkTable Scan on x x_1`}, {`Output: x_1.a`}, {`->  CTE Scan on z`}, {`Output: z.a`}, {`->  CTE Scan on z z1`}, {`Output: z1.a`}},
			},
			{
				Statement: `with recursive x(a) as
  ((values ('a'), ('b'))
   union all
   (with z as not materialized (select * from x)
    select z.a || z1.a as a from z cross join z as z1
    where length(z.a || z1.a) < 5))
select * from x;`,
				Results: []sql.Row{{`a`}, {`b`}, {`aa`}, {`ab`}, {`ba`}, {`bb`}, {`aaaa`}, {`aaab`}, {`aaba`}, {`aabb`}, {`abaa`}, {`abab`}, {`abba`}, {`abbb`}, {`baaa`}, {`baab`}, {`baba`}, {`babb`}, {`bbaa`}, {`bbab`}, {`bbba`}, {`bbbb`}},
			},
			{
				Statement: `explain (verbose, costs off)
with recursive x(a) as
  ((values ('a'), ('b'))
   union all
   (with z as not materialized (select * from x)
    select z.a || z.a as a from z
    where length(z.a || z.a) < 5))
select * from x;`,
				Results: []sql.Row{{`CTE Scan on x`}, {`Output: x.a`}, {`CTE x`}, {`->  Recursive Union`}, {`->  Values Scan on "*VALUES*"`}, {`Output: "*VALUES*".column1`}, {`->  WorkTable Scan on x x_1`}, {`Output: (x_1.a || x_1.a)`}, {`Filter: (length((x_1.a || x_1.a)) < 5)`}},
			},
			{
				Statement: `with recursive x(a) as
  ((values ('a'), ('b'))
   union all
   (with z as not materialized (select * from x)
    select z.a || z.a as a from z
    where length(z.a || z.a) < 5))
select * from x;`,
				Results: []sql.Row{{`a`}, {`b`}, {`aa`}, {`bb`}, {`aaaa`}, {`bbbb`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as (select * from int4_tbl)
select * from (with y as (select * from x) select * from y) ss;`,
				Results: []sql.Row{{`Seq Scan on public.int4_tbl`}, {`Output: int4_tbl.f1`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as materialized (select * from int4_tbl)
select * from (with y as (select * from x) select * from y) ss;`,
				Results: []sql.Row{{`CTE Scan on x`}, {`Output: x.f1`}, {`CTE x`}, {`->  Seq Scan on public.int4_tbl`}, {`Output: int4_tbl.f1`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as (select 1 as y)
select * from (with x as (select 2 as y) select * from x) ss;`,
				Results: []sql.Row{{`Result`}, {`Output: 2`}},
			},
			{
				Statement: `explain (verbose, costs off)
with x as (select * from subselect_tbl)
select * from x for update;`,
				Results: []sql.Row{{`Seq Scan on public.subselect_tbl`}, {`Output: subselect_tbl.f1, subselect_tbl.f2, subselect_tbl.f3`}},
			},
		},
	})
}
