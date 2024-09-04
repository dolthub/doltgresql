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

func TestRowtypes(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_rowtypes)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_rowtypes,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create type complex as (r float8, i float8);`,
			},
			{
				Statement: `create temp table fullname (first text, last text);`,
			},
			{
				Statement: `create type quad as (c1 complex, c2 complex);`,
			},
			{
				Statement: `select (1.1,2.2)::complex, row((3.3,4.4),(5.5,null))::quad;`,
				Results:   []sql.Row{{`(1.1,2.2)`, `("(3.3,4.4)","(5.5,)")`}},
			},
			{
				Statement: `select row('Joe', 'Blow')::fullname, '(Joe,Blow)'::fullname;`,
				Results:   []sql.Row{{`(Joe,Blow)`, `(Joe,Blow)`}},
			},
			{
				Statement: `select '(Joe,von Blow)'::fullname, '(Joe,d''Blow)'::fullname;`,
				Results:   []sql.Row{{`(Joe,"von Blow")`, `(Joe,d'Blow)`}},
			},
			{
				Statement: `select '(Joe,"von""Blow")'::fullname, E'(Joe,d\\\\Blow)'::fullname;`,
				Results:   []sql.Row{{`(Joe,"von""Blow")`, `(Joe,"d\\Blow")`}},
			},
			{
				Statement: `select '(Joe,"Blow,Jr")'::fullname;`,
				Results:   []sql.Row{{`(Joe,"Blow,Jr")`}},
			},
			{
				Statement: `select '(Joe,)'::fullname;	-- ok, null 2nd column`,
				Results:   []sql.Row{{`(Joe,)`}},
			},
			{
				Statement:   `select '(Joe)'::fullname;	-- bad`,
				ErrorString: `malformed record literal: "(Joe)"`,
			},
			{
				Statement:   `select '(Joe,,)'::fullname;	-- bad`,
				ErrorString: `malformed record literal: "(Joe,,)"`,
			},
			{
				Statement:   `select '[]'::fullname;          -- bad`,
				ErrorString: `malformed record literal: "[]"`,
			},
			{
				Statement: `select ' (Joe,Blow)  '::fullname;  -- ok, extra whitespace`,
				Results:   []sql.Row{{`(Joe,Blow)`}},
			},
			{
				Statement:   `select '(Joe,Blow) /'::fullname;  -- bad`,
				ErrorString: `malformed record literal: "(Joe,Blow) /"`,
			},
			{
				Statement: `create temp table quadtable(f1 int, q quad);`,
			},
			{
				Statement: `insert into quadtable values (1, ((3.3,4.4),(5.5,6.6)));`,
			},
			{
				Statement: `insert into quadtable values (2, ((null,4.4),(5.5,6.6)));`,
			},
			{
				Statement: `select * from quadtable;`,
				Results:   []sql.Row{{1, `("(3.3,4.4)","(5.5,6.6)")`}, {2, `("(,4.4)","(5.5,6.6)")`}},
			},
			{
				Statement:   `select f1, q.c1 from quadtable;		-- fails, q is a table reference`,
				ErrorString: `missing FROM-clause entry for table "q"`,
			},
			{
				Statement: `select f1, (q).c1, (qq.q).c1.i from quadtable qq;`,
				Results:   []sql.Row{{1, `(3.3,4.4)`, 4.4}, {2, `(,4.4)`, 4.4}},
			},
			{
				Statement: `create temp table people (fn fullname, bd date);`,
			},
			{
				Statement: `insert into people values ('(Joe,Blow)', '1984-01-10');`,
			},
			{
				Statement: `select * from people;`,
				Results:   []sql.Row{{`(Joe,Blow)`, `01-10-1984`}},
			},
			{
				Statement:   `alter table fullname add column suffix text default '';`,
				ErrorString: `cannot alter table "fullname" because column "people.fn" uses its row type`,
			},
			{
				Statement: `alter table fullname add column suffix text default null;`,
			},
			{
				Statement: `select * from people;`,
				Results:   []sql.Row{{`(Joe,Blow,)`, `01-10-1984`}},
			},
			{
				Statement: `update people set fn.suffix = 'Jr';`,
			},
			{
				Statement: `select * from people;`,
				Results:   []sql.Row{{`(Joe,Blow,Jr)`, `01-10-1984`}},
			},
			{
				Statement: `insert into quadtable (f1, q.c1.r, q.c2.i) values(44,55,66);`,
			},
			{
				Statement: `update quadtable set q.c1.r = 12 where f1 = 2;`,
			},
			{
				Statement:   `update quadtable set q.c1 = 12;  -- error, type mismatch`,
				ErrorString: `subfield "c1" is of type complex but expression is of type integer`,
			},
			{
				Statement: `select * from quadtable;`,
				Results:   []sql.Row{{1, `("(3.3,4.4)","(5.5,6.6)")`}, {44, `("(55,)","(,66)")`}, {2, `("(12,4.4)","(5.5,6.6)")`}},
			},
			{
				Statement: `create temp table pp (f1 text);`,
			},
			{
				Statement: `insert into pp values (repeat('abcdefghijkl', 100000));`,
			},
			{
				Statement: `insert into people select ('Jim', f1, null)::fullname, current_date from pp;`,
			},
			{
				Statement: `select (fn).first, substr((fn).last, 1, 20), length((fn).last) from people;`,
				Results:   []sql.Row{{`Joe`, `Blow`, 4}, {`Jim`, `abcdefghijklabcdefgh`, 1200000}},
			},
			{
				Statement: `update people set fn.first = 'Jack';`,
			},
			{
				Statement: `select (fn).first, substr((fn).last, 1, 20), length((fn).last) from people;`,
				Results:   []sql.Row{{`Jack`, `Blow`, 4}, {`Jack`, `abcdefghijklabcdefgh`, 1200000}},
			},
			{
				Statement: `select ROW(1,2) < ROW(1,3) as true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select ROW(1,2) < ROW(1,1) as false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select ROW(1,2) < ROW(1,NULL) as null;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select ROW(1,2,3) < ROW(1,3,NULL) as true; -- the NULL is not examined`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select ROW(11,'ABC') < ROW(11,'DEF') as true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select ROW(11,'ABC') > ROW(11,'DEF') as false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select ROW(12,'ABC') > ROW(11,'DEF') as true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select ROW(1,2,3) < ROW(1,NULL,4) as null;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select ROW(1,2,3) = ROW(1,NULL,4) as false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select ROW(1,2,3) <> ROW(1,NULL,4) as true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select ROW('ABC','DEF') ~<=~ ROW('DEF','ABC') as true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select ROW('ABC','DEF') ~>=~ ROW('DEF','ABC') as false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `select ROW('ABC','DEF') ~~ ROW('DEF','ABC') as fail;`,
				ErrorString: `could not determine interpretation of row comparison operator ~~`,
			},
			{
				Statement: `select ROW(1,2) = ROW(1,2::int8);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select ROW(1,2) in (ROW(3,4), ROW(1,2));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select ROW(1,2) in (ROW(3,4), ROW(1,2::int8));`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select unique1, unique2 from tenk1
where (unique1, unique2) < any (select ten, ten from tenk1 where hundred < 3)
      and unique1 <= 20
order by 1;`,
				Results: []sql.Row{{0, 9998}, {1, 2838}},
			},
			{
				Statement: `explain (costs off)
select thousand, tenthous from tenk1
where (thousand, tenthous) >= (997, 5000)
order by thousand, tenthous;`,
				Results: []sql.Row{{`Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`Index Cond: (ROW(thousand, tenthous) >= ROW(997, 5000))`}},
			},
			{
				Statement: `select thousand, tenthous from tenk1
where (thousand, tenthous) >= (997, 5000)
order by thousand, tenthous;`,
				Results: []sql.Row{{997, 5997}, {997, 6997}, {997, 7997}, {997, 8997}, {997, 9997}, {998, 998}, {998, 1998}, {998, 2998}, {998, 3998}, {998, 4998}, {998, 5998}, {998, 6998}, {998, 7998}, {998, 8998}, {998, 9998}, {999, 999}, {999, 1999}, {999, 2999}, {999, 3999}, {999, 4999}, {999, 5999}, {999, 6999}, {999, 7999}, {999, 8999}, {999, 9999}},
			},
			{
				Statement: `explain (costs off)
select thousand, tenthous, four from tenk1
where (thousand, tenthous, four) > (998, 5000, 3)
order by thousand, tenthous;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: thousand, tenthous`}, {`->  Bitmap Heap Scan on tenk1`}, {`Filter: (ROW(thousand, tenthous, four) > ROW(998, 5000, 3))`}, {`->  Bitmap Index Scan on tenk1_thous_tenthous`}, {`Index Cond: (ROW(thousand, tenthous) >= ROW(998, 5000))`}},
			},
			{
				Statement: `select thousand, tenthous, four from tenk1
where (thousand, tenthous, four) > (998, 5000, 3)
order by thousand, tenthous;`,
				Results: []sql.Row{{998, 5998, 2}, {998, 6998, 2}, {998, 7998, 2}, {998, 8998, 2}, {998, 9998, 2}, {999, 999, 3}, {999, 1999, 3}, {999, 2999, 3}, {999, 3999, 3}, {999, 4999, 3}, {999, 5999, 3}, {999, 6999, 3}, {999, 7999, 3}, {999, 8999, 3}, {999, 9999, 3}},
			},
			{
				Statement: `explain (costs off)
select thousand, tenthous from tenk1
where (998, 5000) < (thousand, tenthous)
order by thousand, tenthous;`,
				Results: []sql.Row{{`Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`Index Cond: (ROW(thousand, tenthous) > ROW(998, 5000))`}},
			},
			{
				Statement: `select thousand, tenthous from tenk1
where (998, 5000) < (thousand, tenthous)
order by thousand, tenthous;`,
				Results: []sql.Row{{998, 5998}, {998, 6998}, {998, 7998}, {998, 8998}, {998, 9998}, {999, 999}, {999, 1999}, {999, 2999}, {999, 3999}, {999, 4999}, {999, 5999}, {999, 6999}, {999, 7999}, {999, 8999}, {999, 9999}},
			},
			{
				Statement: `explain (costs off)
select thousand, hundred from tenk1
where (998, 5000) < (thousand, hundred)
order by thousand, hundred;`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: thousand, hundred`}, {`->  Bitmap Heap Scan on tenk1`}, {`Filter: (ROW(998, 5000) < ROW(thousand, hundred))`}, {`->  Bitmap Index Scan on tenk1_thous_tenthous`}, {`Index Cond: (thousand >= 998)`}},
			},
			{
				Statement: `select thousand, hundred from tenk1
where (998, 5000) < (thousand, hundred)
order by thousand, hundred;`,
				Results: []sql.Row{{999, 99}, {999, 99}, {999, 99}, {999, 99}, {999, 99}, {999, 99}, {999, 99}, {999, 99}, {999, 99}, {999, 99}},
			},
			{
				Statement: `create temp table test_table (a text, b text);`,
			},
			{
				Statement: `insert into test_table values ('a', 'b');`,
			},
			{
				Statement: `insert into test_table select 'a', null from generate_series(1,1000);`,
			},
			{
				Statement: `insert into test_table values ('b', 'a');`,
			},
			{
				Statement: `create index on test_table (a,b);`,
			},
			{
				Statement: `set enable_sort = off;`,
			},
			{
				Statement: `explain (costs off)
select a,b from test_table where (a,b) > ('a','a') order by a,b;`,
				Results: []sql.Row{{`Index Only Scan using test_table_a_b_idx on test_table`}, {`Index Cond: (ROW(a, b) > ROW('a'::text, 'a'::text))`}},
			},
			{
				Statement: `select a,b from test_table where (a,b) > ('a','a') order by a,b;`,
				Results:   []sql.Row{{`a`, `b`}, {`b`, `a`}},
			},
			{
				Statement: `reset enable_sort;`,
			},
			{
				Statement:   `select * from int8_tbl i8 where i8 in (row(123,456));  -- fail, type mismatch`,
				ErrorString: `cannot compare dissimilar column types bigint and integer at record column 1`,
			},
			{
				Statement: `explain (costs off)
select * from int8_tbl i8
where i8 in (row(123,456)::int8_tbl, '(4567890123456789,123)');`,
				Results: []sql.Row{{`Seq Scan on int8_tbl i8`}, {`Filter: (i8.* = ANY ('{"(123,456)","(4567890123456789,123)"}'::int8_tbl[]))`}},
			},
			{
				Statement: `select * from int8_tbl i8
where i8 in (row(123,456)::int8_tbl, '(4567890123456789,123)');`,
				Results: []sql.Row{{123, 456}, {4567890123456789, 123}},
			},
			{
				Statement: `select (row(1, 2.0)).f1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select (row(1, 2.0)).f2;`,
				Results:   []sql.Row{{2.0}},
			},
			{
				Statement:   `select (row(1, 2.0)).nosuch;  -- fail`,
				ErrorString: `could not identify column "nosuch" in record data type`,
			},
			{
				Statement: `select (row(1, 2.0)).*;`,
				Results:   []sql.Row{{1, 2.0}},
			},
			{
				Statement: `select (r).f1 from (select row(1, 2.0) as r) ss;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `select (r).f3 from (select row(1, 2.0) as r) ss;  -- fail`,
				ErrorString: `could not identify column "f3" in record data type`,
			},
			{
				Statement: `select (r).* from (select row(1, 2.0) as r) ss;`,
				Results:   []sql.Row{{1, 2.0}},
			},
			{
				Statement: `select ROW();`,
				Results:   []sql.Row{{`()`}},
			},
			{
				Statement: `select ROW() IS NULL;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select ROW() = ROW();`,
				ErrorString: `cannot compare rows of zero length`,
			},
			{
				Statement: `select array[ row(1,2), row(3,4), row(5,6) ];`,
				Results:   []sql.Row{{`{"(1,2)","(3,4)","(5,6)"}`}},
			},
			{
				Statement: `select row(1,1.1) = any (array[ row(7,7.7), row(1,1.1), row(0,0.0) ]);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1,1.1) = any (array[ row(7,7.7), row(1,1.0), row(0,0.0) ]);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `create type cantcompare as (p point, r float8);`,
			},
			{
				Statement: `create temp table cc (f1 cantcompare);`,
			},
			{
				Statement: `insert into cc values('("(1,2)",3)');`,
			},
			{
				Statement: `insert into cc values('("(4,5)",6)');`,
			},
			{
				Statement:   `select * from cc order by f1; -- fail, but should complain about cantcompare`,
				ErrorString: `could not identify an ordering operator for type cantcompare`,
			},
			{
				Statement: `create type testtype1 as (a int, b int);`,
			},
			{
				Statement: `select row(1, 2)::testtype1 < row(1, 3)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 2)::testtype1 <= row(1, 3)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 2)::testtype1 = row(1, 2)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 2)::testtype1 <> row(1, 3)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 3)::testtype1 >= row(1, 2)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 3)::testtype1 > row(1, 2)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, -2)::testtype1 < row(1, -3)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -2)::testtype1 <= row(1, -3)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -2)::testtype1 = row(1, -3)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -2)::testtype1 <> row(1, -2)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -3)::testtype1 >= row(1, -2)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -3)::testtype1 > row(1, -2)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -2)::testtype1 < row(1, 3)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create type testtype3 as (a int, b text);`,
			},
			{
				Statement:   `select row(1, 2)::testtype1 < row(1, 'abc')::testtype3;`,
				ErrorString: `cannot compare dissimilar column types integer and text at record column 2`,
			},
			{
				Statement:   `select row(1, 2)::testtype1 <> row(1, 'abc')::testtype3;`,
				ErrorString: `cannot compare dissimilar column types integer and text at record column 2`,
			},
			{
				Statement: `create type testtype5 as (a int);`,
			},
			{
				Statement:   `select row(1, 2)::testtype1 < row(1)::testtype5;`,
				ErrorString: `cannot compare record types with different numbers of columns`,
			},
			{
				Statement:   `select row(1, 2)::testtype1 <> row(1)::testtype5;`,
				ErrorString: `cannot compare record types with different numbers of columns`,
			},
			{
				Statement: `create type testtype6 as (a int, b point);`,
			},
			{
				Statement:   `select row(1, '(1,2)')::testtype6 < row(1, '(1,3)')::testtype6;`,
				ErrorString: `could not identify a comparison function for type point`,
			},
			{
				Statement:   `select row(1, '(1,2)')::testtype6 <> row(1, '(1,3)')::testtype6;`,
				ErrorString: `could not identify an equality operator for type point`,
			},
			{
				Statement: `drop type testtype1, testtype3, testtype5, testtype6;`,
			},
			{
				Statement: `create type testtype1 as (a int, b int);`,
			},
			{
				Statement: `select row(1, 2)::testtype1 *< row(1, 3)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 2)::testtype1 *<= row(1, 3)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 2)::testtype1 *= row(1, 2)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 2)::testtype1 *<> row(1, 3)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 3)::testtype1 *>= row(1, 2)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 3)::testtype1 *> row(1, 2)::testtype1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, -2)::testtype1 *< row(1, -3)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -2)::testtype1 *<= row(1, -3)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -2)::testtype1 *= row(1, -3)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -2)::testtype1 *<> row(1, -2)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -3)::testtype1 *>= row(1, -2)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -3)::testtype1 *> row(1, -2)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, -2)::testtype1 *< row(1, 3)::testtype1;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `create type testtype2 as (a smallint, b bool);  -- byval different sizes`,
			},
			{
				Statement: `select row(1, true)::testtype2 *< row(2, true)::testtype2;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(-2, true)::testtype2 *< row(-1, true)::testtype2;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(0, false)::testtype2 *< row(0, true)::testtype2;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(0, false)::testtype2 *<> row(0, true)::testtype2;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create type testtype3 as (a int, b text);  -- variable length`,
			},
			{
				Statement: `select row(1, 'abc')::testtype3 *< row(1, 'abd')::testtype3;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 'abc')::testtype3 *< row(1, 'abcd')::testtype3;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, 'abc')::testtype3 *> row(1, 'abd')::testtype3;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, 'abc')::testtype3 *<> row(1, 'abd')::testtype3;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `create type testtype4 as (a int, b point);  -- by ref, fixed length`,
			},
			{
				Statement: `select row(1, '(1,2)')::testtype4 *< row(1, '(1,3)')::testtype4;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, '(1,2)')::testtype4 *<> row(1, '(1,3)')::testtype4;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select row(1, 2)::testtype1 *< row(1, 'abc')::testtype3;`,
				ErrorString: `cannot compare dissimilar column types integer and text at record column 2`,
			},
			{
				Statement:   `select row(1, 2)::testtype1 *<> row(1, 'abc')::testtype3;`,
				ErrorString: `cannot compare dissimilar column types integer and text at record column 2`,
			},
			{
				Statement: `create type testtype5 as (a int);`,
			},
			{
				Statement:   `select row(1, 2)::testtype1 *< row(1)::testtype5;`,
				ErrorString: `cannot compare record types with different numbers of columns`,
			},
			{
				Statement:   `select row(1, 2)::testtype1 *<> row(1)::testtype5;`,
				ErrorString: `cannot compare record types with different numbers of columns`,
			},
			{
				Statement: `create type testtype6 as (a int, b point);`,
			},
			{
				Statement: `select row(1, '(1,2)')::testtype6 *< row(1, '(1,3)')::testtype6;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select row(1, '(1,2)')::testtype6 *>= row(1, '(1,3)')::testtype6;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select row(1, '(1,2)')::testtype6 *<> row(1, '(1,3)')::testtype6;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select q.a, q.b = row(2), q.c = array[row(3)], q.d = row(row(4)) from
    unnest(array[row(1, row(2), array[row(3)], row(row(4))),
                 row(2, row(3), array[row(4)], row(row(5)))])
      as q(a int, b record, c record[], d record);`,
				Results: []sql.Row{{1, true, true, true}, {2, false, false, false}},
			},
			{
				Statement: `drop type testtype1, testtype2, testtype3, testtype4, testtype5, testtype6;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TABLE price (
    id SERIAL PRIMARY KEY,
    active BOOLEAN NOT NULL,
    price NUMERIC
);`,
			},
			{
				Statement: `CREATE TYPE price_input AS (
    id INTEGER,
    price NUMERIC
);`,
			},
			{
				Statement: `CREATE TYPE price_key AS (
    id INTEGER
);`,
			},
			{
				Statement: `CREATE FUNCTION price_key_from_table(price) RETURNS price_key AS $$
    SELECT $1.id
$$ LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION price_key_from_input(price_input) RETURNS price_key AS $$
    SELECT $1.id
$$ LANGUAGE SQL;`,
			},
			{
				Statement: `insert into price values (1,false,42), (10,false,100), (11,true,17.99);`,
			},
			{
				Statement: `UPDATE price
    SET active = true, price = input_prices.price
    FROM unnest(ARRAY[(10, 123.00), (11, 99.99)]::price_input[]) input_prices
    WHERE price_key_from_table(price.*) = price_key_from_input(input_prices.*);`,
			},
			{
				Statement: `select * from price;`,
				Results:   []sql.Row{{1, false, 42}, {10, true, 123.00}, {11, true, 99.99}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `create temp table compos (f1 int, f2 text);`,
			},
			{
				Statement: `create function fcompos1(v compos) returns void as $$
insert into compos values (v);  -- fail`,
			},
			{
				Statement:   `$$ language sql;`,
				ErrorString: `column "f1" is of type integer but expression is of type compos`,
			},
			{
				Statement: `create function fcompos1(v compos) returns void as $$
insert into compos values (v.*);`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `create function fcompos2(v compos) returns void as $$
select fcompos1(v);`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `create function fcompos3(v compos) returns void as $$
select fcompos1(fcompos3.v.*);`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select fcompos1(row(1,'one'));`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select fcompos2(row(2,'two'));`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select fcompos3(row(3,'three'));`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select * from compos;`,
				Results:   []sql.Row{{1, `one`}, {2, `two`}, {3, `three`}},
			},
			{
				Statement: `select cast (fullname as text) from fullname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select fullname::text from fullname;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `select text(fullname) from fullname;  -- error`,
				ErrorString: `function text(fullname) does not exist`,
			},
			{
				Statement:   `select fullname.text from fullname;  -- error`,
				ErrorString: `column fullname.text does not exist`,
			},
			{
				Statement: `select cast (row('Jim', 'Beam') as text);`,
				Results:   []sql.Row{{`(Jim,Beam)`}},
			},
			{
				Statement: `select (row('Jim', 'Beam'))::text;`,
				Results:   []sql.Row{{`(Jim,Beam)`}},
			},
			{
				Statement:   `select text(row('Jim', 'Beam'));  -- error`,
				ErrorString: `function text(record) does not exist`,
			},
			{
				Statement:   `select (row('Jim', 'Beam')).text;  -- error`,
				ErrorString: `could not identify column "text" in record data type`,
			},
			{
				Statement: `insert into fullname values ('Joe', 'Blow');`,
			},
			{
				Statement: `select f.last from fullname f;`,
				Results:   []sql.Row{{`Blow`}},
			},
			{
				Statement: `select last(f) from fullname f;`,
				Results:   []sql.Row{{`Blow`}},
			},
			{
				Statement: `create function longname(fullname) returns text language sql
as $$select $1.first || ' ' || $1.last$$;`,
			},
			{
				Statement: `select f.longname from fullname f;`,
				Results:   []sql.Row{{`Joe Blow`}},
			},
			{
				Statement: `select longname(f) from fullname f;`,
				Results:   []sql.Row{{`Joe Blow`}},
			},
			{
				Statement: `alter table fullname add column longname text;`,
			},
			{
				Statement: `select f.longname from fullname f;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select longname(f) from fullname f;`,
				Results:   []sql.Row{{`Joe Blow`}},
			},
			{
				Statement: `select row_to_json(i) from int8_tbl i;`,
				Results:   []sql.Row{{`{"q1":123,"q2":456}`}, {`{"q1":123,"q2":4567890123456789}`}, {`{"q1":4567890123456789,"q2":123}`}, {`{"q1":4567890123456789,"q2":4567890123456789}`}, {`{"q1":4567890123456789,"q2":-4567890123456789}`}},
			},
			{
				Statement: `select row_to_json(i) from int8_tbl i(x,y);`,
				Results:   []sql.Row{{`{"q1":123,"q2":456}`}, {`{"q1":123,"q2":4567890123456789}`}, {`{"q1":4567890123456789,"q2":123}`}, {`{"q1":4567890123456789,"q2":4567890123456789}`}, {`{"q1":4567890123456789,"q2":-4567890123456789}`}},
			},
			{
				Statement: `select row_to_json(ss) from
  (select q1, q2 from int8_tbl) as ss;`,
				Results: []sql.Row{{`{"q1":123,"q2":456}`}, {`{"q1":123,"q2":4567890123456789}`}, {`{"q1":4567890123456789,"q2":123}`}, {`{"q1":4567890123456789,"q2":4567890123456789}`}, {`{"q1":4567890123456789,"q2":-4567890123456789}`}},
			},
			{
				Statement: `select row_to_json(ss) from
  (select q1, q2 from int8_tbl offset 0) as ss;`,
				Results: []sql.Row{{`{"q1":123,"q2":456}`}, {`{"q1":123,"q2":4567890123456789}`}, {`{"q1":4567890123456789,"q2":123}`}, {`{"q1":4567890123456789,"q2":4567890123456789}`}, {`{"q1":4567890123456789,"q2":-4567890123456789}`}},
			},
			{
				Statement: `select row_to_json(ss) from
  (select q1 as a, q2 as b from int8_tbl) as ss;`,
				Results: []sql.Row{{`{"a":123,"b":456}`}, {`{"a":123,"b":4567890123456789}`}, {`{"a":4567890123456789,"b":123}`}, {`{"a":4567890123456789,"b":4567890123456789}`}, {`{"a":4567890123456789,"b":-4567890123456789}`}},
			},
			{
				Statement: `select row_to_json(ss) from
  (select q1 as a, q2 as b from int8_tbl offset 0) as ss;`,
				Results: []sql.Row{{`{"a":123,"b":456}`}, {`{"a":123,"b":4567890123456789}`}, {`{"a":4567890123456789,"b":123}`}, {`{"a":4567890123456789,"b":4567890123456789}`}, {`{"a":4567890123456789,"b":-4567890123456789}`}},
			},
			{
				Statement: `select row_to_json(ss) from
  (select q1 as a, q2 as b from int8_tbl) as ss(x,y);`,
				Results: []sql.Row{{`{"x":123,"y":456}`}, {`{"x":123,"y":4567890123456789}`}, {`{"x":4567890123456789,"y":123}`}, {`{"x":4567890123456789,"y":4567890123456789}`}, {`{"x":4567890123456789,"y":-4567890123456789}`}},
			},
			{
				Statement: `select row_to_json(ss) from
  (select q1 as a, q2 as b from int8_tbl offset 0) as ss(x,y);`,
				Results: []sql.Row{{`{"x":123,"y":456}`}, {`{"x":123,"y":4567890123456789}`}, {`{"x":4567890123456789,"y":123}`}, {`{"x":4567890123456789,"y":4567890123456789}`}, {`{"x":4567890123456789,"y":-4567890123456789}`}},
			},
			{
				Statement: `explain (costs off)
select row_to_json(q) from
  (select thousand, tenthous from tenk1
   where thousand = 42 and tenthous < 2000 offset 0) q;`,
				Results: []sql.Row{{`Subquery Scan on q`}, {`->  Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`Index Cond: ((thousand = 42) AND (tenthous < 2000))`}},
			},
			{
				Statement: `select row_to_json(q) from
  (select thousand, tenthous from tenk1
   where thousand = 42 and tenthous < 2000 offset 0) q;`,
				Results: []sql.Row{{`{"thousand":42,"tenthous":42}`}, {`{"thousand":42,"tenthous":1042}`}},
			},
			{
				Statement: `select row_to_json(q) from
  (select thousand as x, tenthous as y from tenk1
   where thousand = 42 and tenthous < 2000 offset 0) q;`,
				Results: []sql.Row{{`{"x":42,"y":42}`}, {`{"x":42,"y":1042}`}},
			},
			{
				Statement: `select row_to_json(q) from
  (select thousand as x, tenthous as y from tenk1
   where thousand = 42 and tenthous < 2000 offset 0) q(a,b);`,
				Results: []sql.Row{{`{"a":42,"b":42}`}, {`{"a":42,"b":1042}`}},
			},
			{
				Statement: `create temp table tt1 as select * from int8_tbl limit 2;`,
			},
			{
				Statement: `create temp table tt2 () inherits(tt1);`,
			},
			{
				Statement: `insert into tt2 values(0,0);`,
			},
			{
				Statement: `select row_to_json(r) from (select q2,q1 from tt1 offset 0) r;`,
				Results:   []sql.Row{{`{"q2":456,"q1":123}`}, {`{"q2":4567890123456789,"q1":123}`}, {`{"q2":0,"q1":0}`}},
			},
			{
				Statement: `create temp table tt3 () inherits(tt2);`,
			},
			{
				Statement: `insert into tt3 values(33,44);`,
			},
			{
				Statement: `select row_to_json(tt3::tt2::tt1) from tt3;`,
				Results:   []sql.Row{{`{"q1":33,"q2":44}`}},
			},
			{
				Statement: `explain (verbose, costs off)
select r, r is null as isnull, r is not null as isnotnull
from (values (1,row(1,2)), (1,row(null,null)), (1,null),
             (null,row(1,2)), (null,row(null,null)), (null,null) ) r(a,b);`,
				Results: []sql.Row{{`Values Scan on "*VALUES*"`}, {`Output: ROW("*VALUES*".column1, "*VALUES*".column2), (("*VALUES*".column1 IS NULL) AND ("*VALUES*".column2 IS NOT DISTINCT FROM NULL)), (("*VALUES*".column1 IS NOT NULL) AND ("*VALUES*".column2 IS DISTINCT FROM NULL))`}},
			},
			{
				Statement: `select r, r is null as isnull, r is not null as isnotnull
from (values (1,row(1,2)), (1,row(null,null)), (1,null),
             (null,row(1,2)), (null,row(null,null)), (null,null) ) r(a,b);`,
				Results: []sql.Row{{`(1,"(1,2)")`, false, true}, {`(1,"(,)")`, false, true}, {`(1,)`, false, false}, {`(,"(1,2)")`, false, false}, {`(,"(,)")`, false, false}, {`(,)`, true, false}},
			},
			{
				Statement: `explain (verbose, costs off)
with r(a,b) as materialized
  (values (1,row(1,2)), (1,row(null,null)), (1,null),
          (null,row(1,2)), (null,row(null,null)), (null,null) )
select r, r is null as isnull, r is not null as isnotnull from r;`,
				Results: []sql.Row{{`CTE Scan on r`}, {`Output: r.*, (r.* IS NULL), (r.* IS NOT NULL)`}, {`CTE r`}, {`->  Values Scan on "*VALUES*"`}, {`Output: "*VALUES*".column1, "*VALUES*".column2`}},
			},
			{
				Statement: `with r(a,b) as materialized
  (values (1,row(1,2)), (1,row(null,null)), (1,null),
          (null,row(1,2)), (null,row(null,null)), (null,null) )
select r, r is null as isnull, r is not null as isnotnull from r;`,
				Results: []sql.Row{{`(1,"(1,2)")`, false, true}, {`(1,"(,)")`, false, true}, {`(1,)`, false, false}, {`(,"(1,2)")`, false, false}, {`(,"(,)")`, false, false}, {`(,)`, true, false}},
			},
			{
				Statement: `CREATE TABLE compositetable(a text, b text);`,
			},
			{
				Statement: `INSERT INTO compositetable(a, b) VALUES('fa', 'fb');`,
			},
			{
				Statement:   `SELECT d.a FROM (SELECT compositetable AS d FROM compositetable) s;`,
				ErrorString: `missing FROM-clause entry for table "d"`,
			},
			{
				Statement: `SELECT (d).a, (d).b FROM (SELECT compositetable AS d FROM compositetable) s;`,
				Results:   []sql.Row{{`fa`, `fb`}},
			},
			{
				Statement:   `SELECT (d).ctid FROM (SELECT compositetable AS d FROM compositetable) s;`,
				ErrorString: `column "ctid" not found in data type compositetable`,
			},
			{
				Statement:   `SELECT (NULL::compositetable).nonexistent;`,
				ErrorString: `column "nonexistent" not found in data type compositetable`,
			},
			{
				Statement: `SELECT (NULL::compositetable).a;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT (NULL::compositetable).oid;`,
				ErrorString: `column "oid" not found in data type compositetable`,
			},
			{
				Statement: `DROP TABLE compositetable;`,
			},
		},
	})
}
