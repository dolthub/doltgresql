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

func TestArrays(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_arrays)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_arrays,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv abs_srcdir PG_ABS_SRCDIR
CREATE TABLE arrtest (
	a 			int2[],
	b 			int4[][][],
	c 			name[],
	d			text[][],
	e 			float8[],
	f			char(5)[],
	g			varchar(5)[]
);`,
			},
			{
				Statement: `CREATE TABLE array_op_test (
	seqno		int4,
	i			int4[],
	t			text[]
);`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/array.data'
COPY array_op_test FROM :'filename';`,
			},
			{
				Statement: `ANALYZE array_op_test;`,
			},
			{
				Statement: `INSERT INTO arrtest (a[1:5], b[1:1][1:2][1:2], c, d, f, g)
   VALUES ('{1,2,3,4,5}', '{{{0,0},{1,2}}}', '{}', '{}', '{}', '{}');`,
			},
			{
				Statement: `UPDATE arrtest SET e[0] = '1.1';`,
			},
			{
				Statement: `UPDATE arrtest SET e[1] = '2.2';`,
			},
			{
				Statement: `INSERT INTO arrtest (f)
   VALUES ('{"too long"}');`,
				ErrorString: `value too long for type character(5)`,
			},
			{
				Statement: `INSERT INTO arrtest (a, b[1:2][1:2], c, d, e, f, g)
   VALUES ('{11,12,23}', '{{3,4},{4,5}}', '{"foobar"}',
           '{{"elt1", "elt2"}}', '{"3.4", "6.7"}',
           '{"abc","abcde"}', '{"abc","abcde"}');`,
			},
			{
				Statement: `INSERT INTO arrtest (a, b[1:2], c, d[1:2])
   VALUES ('{}', '{3,4}', '{foo,bar}', '{bar,foo}');`,
			},
			{
				Statement:   `INSERT INTO arrtest (b[2]) VALUES(now());  -- error, type mismatch`,
				ErrorString: `subscripted assignment to "b" requires type integer but expression is of type timestamp with time zone`,
			},
			{
				Statement:   `INSERT INTO arrtest (b[1:2]) VALUES(now());  -- error, type mismatch`,
				ErrorString: `subscripted assignment to "b" requires type integer[] but expression is of type timestamp with time zone`,
			},
			{
				Statement: `SELECT * FROM arrtest;`,
				Results:   []sql.Row{{`{1,2,3,4,5}`, `{{{0,0},{1,2}}}`, `{}`, `{}`, `[0:1]={1.1,2.2}`, `{}`, `{}`}, {`{11,12,23}`, `{{3,4},{4,5}}`, `{foobar}`, `{{elt1,elt2}}`, `{3.4,6.7}`, `{"abc  ",abcde}`, `{abc,abcde}`}, {`{}`, `{3,4}`, `{foo,bar}`, `{bar,foo}`, ``, ``, ``}},
			},
			{
				Statement: `SELECT arrtest.a[1],
          arrtest.b[1][1][1],
          arrtest.c[1],
          arrtest.d[1][1],
          arrtest.e[0]
   FROM arrtest;`,
				Results: []sql.Row{{1, 0, ``, ``, 1.1}, {11, ``, `foobar`, `elt1`, ``}, {``, ``, `foo`, ``, ``}},
			},
			{
				Statement: `SELECT a[1], b[1][1][1], c[1], d[1][1], e[0]
   FROM arrtest;`,
				Results: []sql.Row{{1, 0, ``, ``, 1.1}, {11, ``, `foobar`, `elt1`, ``}, {``, ``, `foo`, ``, ``}},
			},
			{
				Statement: `SELECT a[1:3],
          b[1:1][1:2][1:2],
          c[1:2],
          d[1:1][1:2]
   FROM arrtest;`,
				Results: []sql.Row{{`{1,2,3}`, `{{{0,0},{1,2}}}`, `{}`, `{}`}, {`{11,12,23}`, `{}`, `{foobar}`, `{{elt1,elt2}}`}, {`{}`, `{}`, `{foo,bar}`, `{}`}},
			},
			{
				Statement: `SELECT array_ndims(a) AS a,array_ndims(b) AS b,array_ndims(c) AS c
   FROM arrtest;`,
				Results: []sql.Row{{1, 3, ``}, {1, 2, 1}, {``, 1, 1}},
			},
			{
				Statement: `SELECT array_dims(a) AS a,array_dims(b) AS b,array_dims(c) AS c
   FROM arrtest;`,
				Results: []sql.Row{{`[1:5]`, `[1:1][1:2][1:2]`, ``}, {`[1:3]`, `[1:2][1:2]`, `[1:1]`}, {``, `[1:2]`, `[1:2]`}},
			},
			{
				Statement: `SELECT *
   FROM arrtest
   WHERE a[1] < 5 and
         c = '{"foobar"}'::_name;`,
				Results: []sql.Row{},
			},
			{
				Statement: `UPDATE arrtest
  SET a[1:2] = '{16,25}'
  WHERE NOT a = '{}'::_int2;`,
			},
			{
				Statement: `UPDATE arrtest
  SET b[1:1][1:1][1:2] = '{113, 117}',
      b[1:1][1:2][2:2] = '{142, 147}'
  WHERE array_dims(b) = '[1:1][1:2][1:2]';`,
			},
			{
				Statement: `UPDATE arrtest
  SET c[2:2] = '{"new_word"}'
  WHERE array_dims(c) is not null;`,
			},
			{
				Statement: `SELECT a,b,c FROM arrtest;`,
				Results:   []sql.Row{{`{16,25,3,4,5}`, `{{{113,142},{1,147}}}`, `{}`}, {`{}`, `{3,4}`, `{foo,new_word}`}, {`{16,25,23}`, `{{3,4},{4,5}}`, `{foobar,new_word}`}},
			},
			{
				Statement: `SELECT a[1:3],
          b[1:1][1:2][1:2],
          c[1:2],
          d[1:1][2:2]
   FROM arrtest;`,
				Results: []sql.Row{{`{16,25,3}`, `{{{113,142},{1,147}}}`, `{}`, `{}`}, {`{}`, `{}`, `{foo,new_word}`, `{}`}, {`{16,25,23}`, `{}`, `{foobar,new_word}`, `{{elt2}}`}},
			},
			{
				Statement: `SELECT b[1:1][2][2],
       d[1:1][2]
   FROM arrtest;`,
				Results: []sql.Row{{`{{{113,142},{1,147}}}`, `{}`}, {`{}`, `{}`}, {`{}`, `{{elt1,elt2}}`}},
			},
			{
				Statement: `INSERT INTO arrtest(a) VALUES('{1,null,3}');`,
			},
			{
				Statement: `SELECT a FROM arrtest;`,
				Results:   []sql.Row{{`{16,25,3,4,5}`}, {`{}`}, {`{16,25,23}`}, {`{1,NULL,3}`}},
			},
			{
				Statement: `UPDATE arrtest SET a[4] = NULL WHERE a[2] IS NULL;`,
			},
			{
				Statement: `SELECT a FROM arrtest WHERE a[2] IS NULL;`,
				Results:   []sql.Row{{`[4:4]={NULL}`}, {`{1,NULL,3,NULL}`}},
			},
			{
				Statement: `DELETE FROM arrtest WHERE a[2] IS NULL AND b IS NULL;`,
			},
			{
				Statement: `SELECT a,b,c FROM arrtest;`,
				Results:   []sql.Row{{`{16,25,3,4,5}`, `{{{113,142},{1,147}}}`, `{}`}, {`{16,25,23}`, `{{3,4},{4,5}}`, `{foobar,new_word}`}, {`[4:4]={NULL}`, `{3,4}`, `{foo,new_word}`}},
			},
			{
				Statement: `select '{{1,2,3},{4,5,6},{7,8,9}}'::int[];`,
				Results:   []sql.Row{{`{{1,2,3},{4,5,6},{7,8,9}}`}},
			},
			{
				Statement: `select ('{{1,2,3},{4,5,6},{7,8,9}}'::int[])[1:2][2];`,
				Results:   []sql.Row{{`{{1,2},{4,5}}`}},
			},
			{
				Statement: `select '[0:2][0:2]={{1,2,3},{4,5,6},{7,8,9}}'::int[];`,
				Results:   []sql.Row{{`[0:2][0:2]={{1,2,3},{4,5,6},{7,8,9}}`}},
			},
			{
				Statement: `select ('[0:2][0:2]={{1,2,3},{4,5,6},{7,8,9}}'::int[])[1:2][2];`,
				Results:   []sql.Row{{`{{5,6},{8,9}}`}},
			},
			{
				Statement:   `SELECT ('{}'::int[])[1][2][3][4][5][6][7];`,
				ErrorString: `number of array dimensions (7) exceeds the maximum allowed (6)`,
			},
			{
				Statement: `SELECT ('{{{1},{2},{3}},{{4},{5},{6}}}'::int[])[1][NULL][1];`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT ('{{{1},{2},{3}},{{4},{5},{6}}}'::int[])[1][NULL:1][1];`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT ('{{{1},{2},{3}},{{4},{5},{6}}}'::int[])[1][1:NULL][1];`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `UPDATE arrtest
  SET c[NULL] = '{"can''t assign"}'
  WHERE array_dims(c) is not null;`,
				ErrorString: `array subscript in assignment must not be null`,
			},
			{
				Statement: `UPDATE arrtest
  SET c[NULL:1] = '{"can''t assign"}'
  WHERE array_dims(c) is not null;`,
				ErrorString: `array subscript in assignment must not be null`,
			},
			{
				Statement: `UPDATE arrtest
  SET c[1:NULL] = '{"can''t assign"}'
  WHERE array_dims(c) is not null;`,
				ErrorString: `array subscript in assignment must not be null`,
			},
			{
				Statement:   `SELECT (now())[1];`,
				ErrorString: `cannot subscript type timestamp with time zone because it does not support subscripting`,
			},
			{
				Statement: `CREATE TEMP TABLE arrtest_s (
  a       int2[],
  b       int2[][]
);`,
			},
			{
				Statement: `INSERT INTO arrtest_s VALUES ('{1,2,3,4,5}', '{{1,2,3}, {4,5,6}, {7,8,9}}');`,
			},
			{
				Statement: `INSERT INTO arrtest_s VALUES ('[0:4]={1,2,3,4,5}', '[0:2][0:2]={{1,2,3}, {4,5,6}, {7,8,9}}');`,
			},
			{
				Statement: `SELECT * FROM arrtest_s;`,
				Results:   []sql.Row{{`{1,2,3,4,5}`, `{{1,2,3},{4,5,6},{7,8,9}}`}, {`[0:4]={1,2,3,4,5}`, `[0:2][0:2]={{1,2,3},{4,5,6},{7,8,9}}`}},
			},
			{
				Statement: `SELECT a[:3], b[:2][:2] FROM arrtest_s;`,
				Results:   []sql.Row{{`{1,2,3}`, `{{1,2},{4,5}}`}, {`{1,2,3,4}`, `{{1,2,3},{4,5,6},{7,8,9}}`}},
			},
			{
				Statement: `SELECT a[2:], b[2:][2:] FROM arrtest_s;`,
				Results:   []sql.Row{{`{2,3,4,5}`, `{{5,6},{8,9}}`}, {`{3,4,5}`, `{{9}}`}},
			},
			{
				Statement: `SELECT a[:], b[:] FROM arrtest_s;`,
				Results:   []sql.Row{{`{1,2,3,4,5}`, `{{1,2,3},{4,5,6},{7,8,9}}`}, {`{1,2,3,4,5}`, `{{1,2,3},{4,5,6},{7,8,9}}`}},
			},
			{
				Statement: `UPDATE arrtest_s SET a[:3] = '{11, 12, 13}', b[:2][:2] = '{{11,12}, {14,15}}'
  WHERE array_lower(a,1) = 1;`,
			},
			{
				Statement: `SELECT * FROM arrtest_s;`,
				Results:   []sql.Row{{`[0:4]={1,2,3,4,5}`, `[0:2][0:2]={{1,2,3},{4,5,6},{7,8,9}}`}, {`{11,12,13,4,5}`, `{{11,12,3},{14,15,6},{7,8,9}}`}},
			},
			{
				Statement: `UPDATE arrtest_s SET a[3:] = '{23, 24, 25}', b[2:][2:] = '{{25,26}, {28,29}}';`,
			},
			{
				Statement: `SELECT * FROM arrtest_s;`,
				Results:   []sql.Row{{`[0:4]={1,2,3,23,24}`, `[0:2][0:2]={{1,2,3},{4,5,6},{7,8,25}}`}, {`{11,12,23,24,25}`, `{{11,12,3},{14,25,26},{7,28,29}}`}},
			},
			{
				Statement: `UPDATE arrtest_s SET a[:] = '{11, 12, 13, 14, 15}';`,
			},
			{
				Statement: `SELECT * FROM arrtest_s;`,
				Results:   []sql.Row{{`[0:4]={11,12,13,14,15}`, `[0:2][0:2]={{1,2,3},{4,5,6},{7,8,25}}`}, {`{11,12,13,14,15}`, `{{11,12,3},{14,25,26},{7,28,29}}`}},
			},
			{
				Statement:   `UPDATE arrtest_s SET a[:] = '{23, 24, 25}';  -- fail, too small`,
				ErrorString: `source array too small`,
			},
			{
				Statement: `INSERT INTO arrtest_s VALUES(NULL, NULL);`,
			},
			{
				Statement:   `UPDATE arrtest_s SET a[:] = '{11, 12, 13, 14, 15}';  -- fail, no good with null`,
				ErrorString: `array slice subscript must provide both boundaries`,
			},
			{
				Statement: `CREATE TEMP TABLE point_tbl AS SELECT * FROM public.point_tbl;`,
			},
			{
				Statement: `INSERT INTO POINT_TBL(f1) VALUES (NULL);`,
			},
			{
				Statement:   `SELECT f1[0:1] FROM POINT_TBL;`,
				ErrorString: `slices of fixed-length arrays not implemented`,
			},
			{
				Statement:   `SELECT f1[0:] FROM POINT_TBL;`,
				ErrorString: `slices of fixed-length arrays not implemented`,
			},
			{
				Statement:   `SELECT f1[:1] FROM POINT_TBL;`,
				ErrorString: `slices of fixed-length arrays not implemented`,
			},
			{
				Statement:   `SELECT f1[:] FROM POINT_TBL;`,
				ErrorString: `slices of fixed-length arrays not implemented`,
			},
			{
				Statement: `UPDATE point_tbl SET f1[0] = 10 WHERE f1 IS NULL RETURNING *;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `INSERT INTO point_tbl(f1[0]) VALUES(0) RETURNING *;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `UPDATE point_tbl SET f1[0] = NULL WHERE f1::text = '(10,10)'::point::text RETURNING *;`,
				Results:   []sql.Row{{`(10,10)`}},
			},
			{
				Statement: `UPDATE point_tbl SET f1[0] = -10, f1[1] = -10 WHERE f1::text = '(10,10)'::point::text RETURNING *;`,
				Results:   []sql.Row{{`(-10,-10)`}},
			},
			{
				Statement:   `UPDATE point_tbl SET f1[3] = 10 WHERE f1::text = '(-10,-10)'::point::text RETURNING *;`,
				ErrorString: `array subscript out of range`,
			},
			{
				Statement: `CREATE TEMP TABLE arrtest1 (i int[], t text[]);`,
			},
			{
				Statement: `insert into arrtest1 values(array[1,2,null,4], array['one','two',null,'four']);`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`{1,2,NULL,4}`, `{one,two,NULL,four}`}},
			},
			{
				Statement: `update arrtest1 set i[2] = 22, t[2] = 'twenty-two';`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`{1,22,NULL,4}`, `{one,twenty-two,NULL,four}`}},
			},
			{
				Statement: `update arrtest1 set i[5] = 5, t[5] = 'five';`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`{1,22,NULL,4,5}`, `{one,twenty-two,NULL,four,five}`}},
			},
			{
				Statement: `update arrtest1 set i[8] = 8, t[8] = 'eight';`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`{1,22,NULL,4,5,NULL,NULL,8}`, `{one,twenty-two,NULL,four,five,NULL,NULL,eight}`}},
			},
			{
				Statement: `update arrtest1 set i[0] = 0, t[0] = 'zero';`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[0:8]={0,1,22,NULL,4,5,NULL,NULL,8}`, `[0:8]={zero,one,twenty-two,NULL,four,five,NULL,NULL,eight}`}},
			},
			{
				Statement: `update arrtest1 set i[-3] = -3, t[-3] = 'minus-three';`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[-3:8]={-3,NULL,NULL,0,1,22,NULL,4,5,NULL,NULL,8}`, `[-3:8]={minus-three,NULL,NULL,zero,one,twenty-two,NULL,four,five,NULL,NULL,eight}`}},
			},
			{
				Statement: `update arrtest1 set i[0:2] = array[10,11,12], t[0:2] = array['ten','eleven','twelve'];`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[-3:8]={-3,NULL,NULL,10,11,12,NULL,4,5,NULL,NULL,8}`, `[-3:8]={minus-three,NULL,NULL,ten,eleven,twelve,NULL,four,five,NULL,NULL,eight}`}},
			},
			{
				Statement: `update arrtest1 set i[8:10] = array[18,null,20], t[8:10] = array['p18',null,'p20'];`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[-3:10]={-3,NULL,NULL,10,11,12,NULL,4,5,NULL,NULL,18,NULL,20}`, `[-3:10]={minus-three,NULL,NULL,ten,eleven,twelve,NULL,four,five,NULL,NULL,p18,NULL,p20}`}},
			},
			{
				Statement: `update arrtest1 set i[11:12] = array[null,22], t[11:12] = array[null,'p22'];`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[-3:12]={-3,NULL,NULL,10,11,12,NULL,4,5,NULL,NULL,18,NULL,20,NULL,22}`, `[-3:12]={minus-three,NULL,NULL,ten,eleven,twelve,NULL,four,five,NULL,NULL,p18,NULL,p20,NULL,p22}`}},
			},
			{
				Statement: `update arrtest1 set i[15:16] = array[null,26], t[15:16] = array[null,'p26'];`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[-3:16]={-3,NULL,NULL,10,11,12,NULL,4,5,NULL,NULL,18,NULL,20,NULL,22,NULL,NULL,NULL,26}`, `[-3:16]={minus-three,NULL,NULL,ten,eleven,twelve,NULL,four,five,NULL,NULL,p18,NULL,p20,NULL,p22,NULL,NULL,NULL,p26}`}},
			},
			{
				Statement: `update arrtest1 set i[-5:-3] = array[-15,-14,-13], t[-5:-3] = array['m15','m14','m13'];`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[-5:16]={-15,-14,-13,NULL,NULL,10,11,12,NULL,4,5,NULL,NULL,18,NULL,20,NULL,22,NULL,NULL,NULL,26}`, `[-5:16]={m15,m14,m13,NULL,NULL,ten,eleven,twelve,NULL,four,five,NULL,NULL,p18,NULL,p20,NULL,p22,NULL,NULL,NULL,p26}`}},
			},
			{
				Statement: `update arrtest1 set i[-7:-6] = array[-17,null], t[-7:-6] = array['m17',null];`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[-7:16]={-17,NULL,-15,-14,-13,NULL,NULL,10,11,12,NULL,4,5,NULL,NULL,18,NULL,20,NULL,22,NULL,NULL,NULL,26}`, `[-7:16]={m17,NULL,m15,m14,m13,NULL,NULL,ten,eleven,twelve,NULL,four,five,NULL,NULL,p18,NULL,p20,NULL,p22,NULL,NULL,NULL,p26}`}},
			},
			{
				Statement: `update arrtest1 set i[-12:-10] = array[-22,null,-20], t[-12:-10] = array['m22',null,'m20'];`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[-12:16]={-22,NULL,-20,NULL,NULL,-17,NULL,-15,-14,-13,NULL,NULL,10,11,12,NULL,4,5,NULL,NULL,18,NULL,20,NULL,22,NULL,NULL,NULL,26}`, `[-12:16]={m22,NULL,m20,NULL,NULL,m17,NULL,m15,m14,m13,NULL,NULL,ten,eleven,twelve,NULL,four,five,NULL,NULL,p18,NULL,p20,NULL,p22,NULL,NULL,NULL,p26}`}},
			},
			{
				Statement: `delete from arrtest1;`,
			},
			{
				Statement: `insert into arrtest1 values(array[1,2,null,4], array['one','two',null,'four']);`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`{1,2,NULL,4}`, `{one,two,NULL,four}`}},
			},
			{
				Statement: `update arrtest1 set i[0:5] = array[0,1,2,null,4,5], t[0:5] = array['z','p1','p2',null,'p4','p5'];`,
			},
			{
				Statement: `select * from arrtest1;`,
				Results:   []sql.Row{{`[0:5]={0,1,2,NULL,4,5}`, `[0:5]={z,p1,p2,NULL,p4,p5}`}},
			},
			{
				Statement: `CREATE TEMP TABLE arrtest2 (i integer ARRAY[4], f float8[], n numeric[], t text[], d timestamp[]);`,
			},
			{
				Statement: `INSERT INTO arrtest2 VALUES(
  ARRAY[[[113,142],[1,147]]],
  ARRAY[1.1,1.2,1.3]::float8[],
  ARRAY[1.1,1.2,1.3],
  ARRAY[[['aaa','aab'],['aba','abb'],['aca','acb']],[['baa','bab'],['bba','bbb'],['bca','bcb']]],
  ARRAY['19620326','19931223','19970117']::timestamp[]
);`,
			},
			{
				Statement: `CREATE TEMP TABLE arrtest_f (f0 int, f1 text, f2 float8);`,
			},
			{
				Statement: `insert into arrtest_f values(1,'cat1',1.21);`,
			},
			{
				Statement: `insert into arrtest_f values(2,'cat1',1.24);`,
			},
			{
				Statement: `insert into arrtest_f values(3,'cat1',1.18);`,
			},
			{
				Statement: `insert into arrtest_f values(4,'cat1',1.26);`,
			},
			{
				Statement: `insert into arrtest_f values(5,'cat1',1.15);`,
			},
			{
				Statement: `insert into arrtest_f values(6,'cat2',1.15);`,
			},
			{
				Statement: `insert into arrtest_f values(7,'cat2',1.26);`,
			},
			{
				Statement: `insert into arrtest_f values(8,'cat2',1.32);`,
			},
			{
				Statement: `insert into arrtest_f values(9,'cat2',1.30);`,
			},
			{
				Statement: `CREATE TEMP TABLE arrtest_i (f0 int, f1 text, f2 int);`,
			},
			{
				Statement: `insert into arrtest_i values(1,'cat1',21);`,
			},
			{
				Statement: `insert into arrtest_i values(2,'cat1',24);`,
			},
			{
				Statement: `insert into arrtest_i values(3,'cat1',18);`,
			},
			{
				Statement: `insert into arrtest_i values(4,'cat1',26);`,
			},
			{
				Statement: `insert into arrtest_i values(5,'cat1',15);`,
			},
			{
				Statement: `insert into arrtest_i values(6,'cat2',15);`,
			},
			{
				Statement: `insert into arrtest_i values(7,'cat2',26);`,
			},
			{
				Statement: `insert into arrtest_i values(8,'cat2',32);`,
			},
			{
				Statement: `insert into arrtest_i values(9,'cat2',30);`,
			},
			{
				Statement: `SELECT t.f[1][3][1] AS "131", t.f[2][2][1] AS "221" FROM (
  SELECT ARRAY[[[111,112],[121,122],[131,132]],[[211,212],[221,122],[231,232]]] AS f
) AS t;`,
				Results: []sql.Row{{131, 221}},
			},
			{
				Statement: `SELECT ARRAY[[[[[['hello'],['world']]]]]];`,
				Results:   []sql.Row{{`{{{{{{hello},{world}}}}}}`}},
			},
			{
				Statement: `SELECT ARRAY[ARRAY['hello'],ARRAY['world']];`,
				Results:   []sql.Row{{`{{hello},{world}}`}},
			},
			{
				Statement: `SELECT ARRAY(select f2 from arrtest_f order by f2) AS "ARRAY";`,
				Results:   []sql.Row{{`{1.15,1.15,1.18,1.21,1.24,1.26,1.26,1.3,1.32}`}},
			},
			{
				Statement: `SELECT '{1,null,3}'::int[];`,
				Results:   []sql.Row{{`{1,NULL,3}`}},
			},
			{
				Statement: `SELECT ARRAY[1,NULL,3];`,
				Results:   []sql.Row{{`{1,NULL,3}`}},
			},
			{
				Statement: `SELECT array_append(array[42], 6) AS "{42,6}";`,
				Results:   []sql.Row{{`{42,6}`}},
			},
			{
				Statement: `SELECT array_prepend(6, array[42]) AS "{6,42}";`,
				Results:   []sql.Row{{`{6,42}`}},
			},
			{
				Statement: `SELECT array_cat(ARRAY[1,2], ARRAY[3,4]) AS "{1,2,3,4}";`,
				Results:   []sql.Row{{`{1,2,3,4}`}},
			},
			{
				Statement: `SELECT array_cat(ARRAY[1,2], ARRAY[[3,4],[5,6]]) AS "{{1,2},{3,4},{5,6}}";`,
				Results:   []sql.Row{{`{{1,2},{3,4},{5,6}}`}},
			},
			{
				Statement: `SELECT array_cat(ARRAY[[3,4],[5,6]], ARRAY[1,2]) AS "{{3,4},{5,6},{1,2}}";`,
				Results:   []sql.Row{{`{{3,4},{5,6},{1,2}}`}},
			},
			{
				Statement: `SELECT array_position(ARRAY[1,2,3,4,5], 4);`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `SELECT array_position(ARRAY[5,3,4,2,1], 4);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement:   `SELECT array_position(ARRAY[[1,2],[3,4]], 3);`,
				ErrorString: `searching for elements in multidimensional arrays is not supported`,
			},
			{
				Statement: `SELECT array_position(ARRAY['sun','mon','tue','wed','thu','fri','sat'], 'mon');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT array_position(ARRAY['sun','mon','tue','wed','thu','fri','sat'], 'sat');`,
				Results:   []sql.Row{{7}},
			},
			{
				Statement: `SELECT array_position(ARRAY['sun','mon','tue','wed','thu','fri','sat'], NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT array_position(ARRAY['sun','mon','tue','wed','thu',NULL,'fri','sat'], NULL);`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `SELECT array_position(ARRAY['sun','mon','tue','wed','thu',NULL,'fri','sat'], 'sat');`,
				Results:   []sql.Row{{8}},
			},
			{
				Statement: `SELECT array_positions(NULL, 10);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT array_positions(NULL, NULL::int);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT array_positions(ARRAY[1,2,3,4,5,6,1,2,3,4,5,6], 4);`,
				Results:   []sql.Row{{`{4,10}`}},
			},
			{
				Statement:   `SELECT array_positions(ARRAY[[1,2],[3,4]], 4);`,
				ErrorString: `searching for elements in multidimensional arrays is not supported`,
			},
			{
				Statement: `SELECT array_positions(ARRAY[1,2,3,4,5,6,1,2,3,4,5,6], NULL);`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `SELECT array_positions(ARRAY[1,2,3,NULL,5,6,1,2,3,NULL,5,6], NULL);`,
				Results:   []sql.Row{{`{4,10}`}},
			},
			{
				Statement: `SELECT array_length(array_positions(ARRAY(SELECT 'AAAAAAAAAAAAAAAAAAAAAAAAA'::text || i % 10
                                          FROM generate_series(1,100) g(i)),
                                  'AAAAAAAAAAAAAAAAAAAAAAAAA5'), 1);`,
				Results: []sql.Row{{10}},
			},
			{
				Statement: `DO $$
DECLARE
  o int;`,
			},
			{
				Statement: `  a int[] := ARRAY[1,2,3,2,3,1,2];`,
			},
			{
				Statement: `BEGIN
  o := array_position(a, 2);`,
			},
			{
				Statement: `  WHILE o IS NOT NULL
  LOOP
    RAISE NOTICE '%', o;`,
			},
			{
				Statement: `    o := array_position(a, 2, o + 1);`,
			},
			{
				Statement: `  END LOOP;`,
			},
			{
				Statement: `END
$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `SELECT array_position('[2:4]={1,2,3}'::int[], 1);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT array_positions('[2:4]={1,2,3}'::int[], 1);`,
				Results:   []sql.Row{{`{2}`}},
			},
			{
				Statement: `SELECT
    array_position(ids, (1, 1)),
    array_positions(ids, (1, 1))
        FROM
(VALUES
    (ARRAY[(0, 0), (1, 1)]),
    (ARRAY[(1, 1)])
) AS f (ids);`,
				Results: []sql.Row{{2, `{2}`}, {1, `{1}`}},
			},
			{
				Statement: `SELECT a FROM arrtest WHERE b = ARRAY[[[113,142],[1,147]]];`,
				Results:   []sql.Row{{`{16,25,3,4,5}`}},
			},
			{
				Statement: `SELECT NOT ARRAY[1.1,1.2,1.3] = ARRAY[1.1,1.2,1.3] AS "FALSE";`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT ARRAY[1,2] || 3 AS "{1,2,3}";`,
				Results:   []sql.Row{{`{1,2,3}`}},
			},
			{
				Statement: `SELECT 0 || ARRAY[1,2] AS "{0,1,2}";`,
				Results:   []sql.Row{{`{0,1,2}`}},
			},
			{
				Statement: `SELECT ARRAY[1,2] || ARRAY[3,4] AS "{1,2,3,4}";`,
				Results:   []sql.Row{{`{1,2,3,4}`}},
			},
			{
				Statement: `SELECT ARRAY[[['hello','world']]] || ARRAY[[['happy','birthday']]] AS "ARRAY";`,
				Results:   []sql.Row{{`{{{hello,world}},{{happy,birthday}}}`}},
			},
			{
				Statement: `SELECT ARRAY[[1,2],[3,4]] || ARRAY[5,6] AS "{{1,2},{3,4},{5,6}}";`,
				Results:   []sql.Row{{`{{1,2},{3,4},{5,6}}`}},
			},
			{
				Statement: `SELECT ARRAY[0,0] || ARRAY[1,1] || ARRAY[2,2] AS "{0,0,1,1,2,2}";`,
				Results:   []sql.Row{{`{0,0,1,1,2,2}`}},
			},
			{
				Statement: `SELECT 0 || ARRAY[1,2] || 3 AS "{0,1,2,3}";`,
				Results:   []sql.Row{{`{0,1,2,3}`}},
			},
			{
				Statement: `SELECT ARRAY[1.1] || ARRAY[2,3,4];`,
				Results:   []sql.Row{{`{1.1,2,3,4}`}},
			},
			{
				Statement: `SELECT array_agg(x) || array_agg(x) FROM (VALUES (ROW(1,2)), (ROW(3,4))) v(x);`,
				Results:   []sql.Row{{`{"(1,2)","(3,4)","(1,2)","(3,4)"}`}},
			},
			{
				Statement: `SELECT ROW(1,2) || array_agg(x) FROM (VALUES (ROW(3,4)), (ROW(5,6))) v(x);`,
				Results:   []sql.Row{{`{"(1,2)","(3,4)","(5,6)"}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i @> '{32}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i && '{32}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i @> '{17}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i && '{17}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i @> '{32,17}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i && '{32,17}' ORDER BY seqno;`,
				Results:   []sql.Row{{6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i <@ '{38,34,32,89}' ORDER BY seqno;`,
				Results:   []sql.Row{{40, `{34}`, `{AAAAAAAAAAAAAA10611,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAA31334,AAAAA70466,AAAAAAAA81587,AAAAAAA74623}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i = '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i @> '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{1, `{92,75,71,52,64,83}`, `{AAAAAAAA44066,AAAAAA1059,AAAAAAAAAAA176,AAAAAAA48038}`}, {2, `{3,6}`, `{AAAAAA98232,AAAAAAAA79710,AAAAAAAAAAAAAAAAA69675,AAAAAAAAAAAAAAAA55798,AAAAAAAAA12793}`}, {3, `{37,64,95,43,3,41,13,30,11,43}`, `{AAAAAAAAAA48845,AAAAA75968,AAAAA95309,AAA54451,AAAAAAAAAA22292,AAAAAAA99836,A96617,AA17009,AAAAAAAAAAAAAA95246}`}, {4, `{71,39,99,55,33,75,45}`, `{AAAAAAAAA53663,AAAAAAAAAAAAAAA67062,AAAAAAAAAA64777,AAA99043,AAAAAAAAAAAAAAAAAAA91804,39557}`}, {5, `{50,42,77,50,4}`, `{AAAAAAAAAAAAAAAAA26540,AAAAAAA79710,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA176,AAAAA95309,AAAAAAAAAAA46154,AAAAAA66777,AAAAAAAAA27249,AAAAAAAAAA64777,AAAAAAAAAAAAAAAAAAA70104}`}, {6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {7, `{12,51,88,64,8}`, `{AAAAAAAAAAAAAAAAAA12591,AAAAAAAAAAAAAAAAA50407,AAAAAAAAAAAA67946}`}, {8, `{60,84}`, `{AAAAAAA81898,AAAAAA1059,AAAAAAAAAAAA81511,AAAAA961,AAAAAAAAAAAAAAAA31334,AAAAA64741,AA6416,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAAAA50407}`}, {9, `{56,52,35,27,80,44,81,22}`, `{AAAAAAAAAAAAAAA73034,AAAAAAAAAAAAA7929,AAAAAAA66161,AA88409,39557,A27153,AAAAAAAA9523,AAAAAAAAAAA99000}`}, {10, `{71,5,45}`, `{AAAAAAAAAAA21658,AAAAAAAAAAAA21089,AAA54451,AAAAAAAAAAAAAAAAAA54141,AAAAAAAAAAAAAA28620,AAAAAAAAAAA21658,AAAAAAAAAAA74076,AAAAAAAAA27249}`}, {11, `{41,86,74,48,22,74,47,50}`, `{AAAAAAAA9523,AAAAAAAAAAAA37562,AAAAAAAAAAAAAAAA14047,AAAAAAAAAAA46154,AAAA41702,AAAAAAAAAAAAAAAAA764,AAAAA62737,39557}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {13, `{3,52,34,23}`, `{AAAAAA98232,AAAA49534,AAAAAAAAAAA21658}`}, {14, `{78,57,19}`, `{AAAA8857,AAAAAAAAAAAAAAA73034,AAAAAAAA81587,AAAAAAAAAAAAAAA68526,AAAAA75968,AAAAAAAAAAAAAA65909,AAAAAAAAA10012,AAAAAAAAAAAAAA65909}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {16, `{14,63,85,11}`, `{AAAAAA66777}`}, {17, `{7,10,81,85}`, `{AAAAAA43678,AAAAAAA12144,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAAAAA15356}`}, {18, `{1}`, `{AAAAAAAAAAA33576,AAAAA95309,64261,AAA59323,AAAAAAAAAAAAAA95246,55847,AAAAAAAAAAAA67946,AAAAAAAAAAAAAAAAAA64374}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {20, `{72,89,70,51,54,37,8,49,79}`, `{AAAAAA58494}`}, {21, `{2,8,65,10,5,79,43}`, `{AAAAAAAAAAAAAAAAA88852,AAAAAAAAAAAAAAAAAAA91804,AAAAA64669,AAAAAAAAAAAAAAAA1443,AAAAAAAAAAAAAAAA23657,AAAAA12179,AAAAAAAAAAAAAAAAA88852,AAAAAAAAAAAAAAAA31334,AAAAAAAAAAAAAAAA41303,AAAAAAAAAAAAAAAAAAA85420}`}, {22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {23, `{40,90,5,38,72,40,30,10,43,55}`, `{A6053,AAAAAAAAAAA6119,AA44673,AAAAAAAAAAAAAAAAA764,AA17009,AAAAA17383,AAAAA70514,AAAAA33250,AAAAA95309,AAAAAAAAAAAA37562}`}, {24, `{94,61,99,35,48}`, `{AAAAAAAAAAA50956,AAAAAAAAAAA15165,AAAA85070,AAAAAAAAAAAAAAA36627,AAAAA961,AAAAAAAAAA55219}`}, {25, `{31,1,10,11,27,79,38}`, `{AAAAAAAAAAAAAAAAAA59334,45449}`}, {26, `{71,10,9,69,75}`, `{47735,AAAAAAA21462,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAAAAA91804,AAAAAAAAA72121,AAAAAAAAAAAAAAAAAAA1205,AAAAA41597,AAAA8857,AAAAAAAAAAAAAAAAAAA15356,AA17009}`}, {27, `{94}`, `{AA6416,A6053,AAAAAAA21462,AAAAAAA57334,AAAAAAAAAAAAAAAAAA12591,AA88409,AAAAAAAAAAAAA70254}`}, {28, `{14,33,6,34,14}`, `{AAAAAAAAAAAAAAA13198,AAAAAAAA69452,AAAAAAAAAAA82945,AAAAAAA12144,AAAAAAAAA72121,AAAAAAAAAA18601}`}, {29, `{39,21}`, `{AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAAAAA38885,AAAA85070,AAAAAAAAAAAAAAAAAAA70104,AAAAA66674,AAAAAAAAAAAAA62007,AAAAAAAA69452,AAAAAAA1242,AAAAAAAAAAAAAAAA1729,AAAA35194}`}, {30, `{26,81,47,91,34}`, `{AAAAAAAAAAAAAAAAAAA70104,AAAAAAA80240}`}, {31, `{80,24,18,21,54}`, `{AAAAAAAAAAAAAAA13198,AAAAAAAAAAAAAAAAAAA70415,A27153,AAAAAAAAA53663,AAAAAAAAAAAAAAAAA50407,A68938}`}, {32, `{58,79,82,80,67,75,98,10,41}`, `{AAAAAAAAAAAAAAAAAA61286,AAA54451,AAAAAAAAAAAAAAAAAAA87527,A96617,51533}`}, {33, `{74,73}`, `{A85417,AAAAAAA56483,AAAAA17383,AAAAAAAAAAAAA62159,AAAAAAAAAAAA52814,AAAAAAAAAAAAA85723,AAAAAAAAAAAAAAAAAA55796}`}, {34, `{70,45}`, `{AAAAAAAAAAAAAAAAAA71621,AAAAAAAAAAAAAA28620,AAAAAAAAAA55219,AAAAAAAA23648,AAAAAAAAAA22292,AAAAAAA1242}`}, {35, `{23,40}`, `{AAAAAAAAAAAA52814,AAAA48949,AAAAAAAAA34727,AAAA8857,AAAAAAAAAAAAAAAAAAA62179,AAAAAAAAAAAAAAA68526,AAAAAAA99836,AAAAAAAA50094,AAAA91194,AAAAAAAAAAAAA73084}`}, {36, `{79,82,14,52,30,5,79}`, `{AAAAAAAAA53663,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA89194,AA88409,AAAAAAAAAAAAAAA81326,AAAAAAAAAAAAAAAAA63050,AAAAAAAAAAAAAAAA33598}`}, {37, `{53,11,81,39,3,78,58,64,74}`, `{AAAAAAAAAAAAAAAAAAA17075,AAAAAAA66161,AAAAAAAA23648,AAAAAAAAAAAAAA10611}`}, {38, `{59,5,4,95,28}`, `{AAAAAAAAAAA82945,A96617,47735,AAAAA12179,AAAAA64669,AAAAAA99807,AA74433,AAAAAAAAAAAAAAAAA59387}`}, {39, `{82,43,99,16,74}`, `{AAAAAAAAAAAAAAA67062,AAAAAAA57334,AAAAAAAAAAAAAA65909,A27153,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAA64777,AAAAAAAAAAAA81511,AAAAAAAAAAAAAA65909,AAAAAAAAAAAAAA28620}`}, {40, `{34}`, `{AAAAAAAAAAAAAA10611,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAA31334,AAAAA70466,AAAAAAAA81587,AAAAAAA74623}`}, {41, `{19,26,63,12,93,73,27,94}`, `{AAAAAAA79710,AAAAAAAAAA55219,AAAA41702,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA71621,AAAAAAAAAAAAAAAAA63050,AAAAAAA99836,AAAAAAAAAAAAAA8666}`}, {42, `{15,76,82,75,8,91}`, `{AAAAAAAAAAA176,AAAAAA38063,45449,AAAAAA54032,AAAAAAA81898,AA6416,AAAAAAAAAAAAAAAAAAA62179,45449,AAAAA60038,AAAAAAAA81587}`}, {43, `{39,87,91,97,79,28}`, `{AAAAAAAAAAA74076,A96617,AAAAAAAAAAAAAAAAAAA89194,AAAAAAAAAAAAAAAAAA55796,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAA67946}`}, {44, `{40,58,68,29,54}`, `{AAAAAAA81898,AAAAAA66777,AAAAAA98232}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {46, `{53,24}`, `{AAAAAAAAAAA53908,AAAAAA54032,AAAAA17383,AAAA48949,AAAAAAAAAA18601,AAAAA64669,45449,AAAAAAAAAAA98051,AAAAAAAAAAAAAAAAAA71621}`}, {47, `{98,23,64,12,75,61}`, `{AAA59323,AAAAA95309,AAAAAAAAAAAAAAAA31334,AAAAAAAAA27249,AAAAA17383,AAAAAAAAAAAA37562,AAAAAA1059,A84822,55847,AAAAA70466}`}, {48, `{76,14}`, `{AAAAAAAAAAAAA59671,AAAAAAAAAAAAAAAAAAA91804,AAAAAA66777,AAAAAAAAAAAAAAAAAAA89194,AAAAAAAAAAAAAAA36627,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAA73084,AAAAAAA79710,AAAAAAAAAAAAAAA40402,AAAAAAAAAAAAAAAAAAA65037}`}, {49, `{56,5,54,37,49}`, `{AA21643,AAAAAAAAAAA92631,AAAAAAAA81587}`}, {50, `{20,12,37,64,93}`, `{AAAAAAAAAA5483,AAAAAAAAAAAAAAAAAAA1205,AA6416,AAAAAAAAAAAAAAAAA63050,AAAAAAAAAAAAAAAAAA47955}`}, {51, `{47}`, `{AAAAAAAAAAAAAA96505,AAAAAAAAAAAAAAAAAA36842,AAAAA95309,AAAAAAAA81587,AA6416,AAAA91194,AAAAAA58494,AAAAAA1059,AAAAAAAA69452}`}, {52, `{89,0}`, `{AAAAAAAAAAAAAAAAAA47955,AAAAAAA48038,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAAAAA73084,AAAAA70466,AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA46154,AA66862}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {54, `{70,47}`, `{AAAAAAAAAAAAAAAAAA54141,AAAAA40681,AAAAAAA48038,AAAAAAAAAAAAAAAA29150,AAAAA41597,AAAAAAAAAAAAAAAAAA59334,AA15322}`}, {55, `{47,79,47,64,72,25,71,24,93}`, `{AAAAAAAAAAAAAAAAAA55796,AAAAA62737}`}, {56, `{33,7,60,54,93,90,77,85,39}`, `{AAAAAAAAAAAAAAAAAA32918,AA42406}`}, {57, `{23,45,10,42,36,21,9,96}`, `{AAAAAAAAAAAAAAAAAAA70415}`}, {58, `{92}`, `{AAAAAAAAAAAAAAAA98414,AAAAAAAA23648,AAAAAAAAAAAAAAAAAA55796,AA25381,AAAAAAAAAAA6119}`}, {59, `{9,69,46,77}`, `{39557,AAAAAAA89932,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAAAAAAAAA26540,AAA20874,AA6416,AAAAAAAAAAAAAAAAAA47955}`}, {60, `{62,2,59,38,89}`, `{AAAAAAA89932,AAAAAAAAAAAAAAAAAAA15356,AA99927,AA17009,AAAAAAAAAAAAAAA35875}`}, {61, `{72,2,44,95,54,54,13}`, `{AAAAAAAAAAAAAAAAAAA91804}`}, {62, `{83,72,29,73}`, `{AAAAAAAAAAAAA15097,AAAA8857,AAAAAAAAAAAA35809,AAAAAAAAAAAA52814,AAAAAAAAAAAAAAAAAAA38885,AAAAAAAAAAAAAAAAAA24183,AAAAAA43678,A96617}`}, {63, `{11,4,61,87}`, `{AAAAAAAAA27249,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAA13198,AAA20874,39557,51533,AAAAAAAAAAA53908,AAAAAAAAAAAAAA96505,AAAAAAAA78938}`}, {64, `{26,19,34,24,81,78}`, `{A96617,AAAAAAAAAAAAAAAAAAA70104,A68938,AAAAAAAAAAA53908,AAAAAAAAAAAAAAA453,AA17009,AAAAAAA80240}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {66, `{31,23,70,52,4,33,48,25}`, `{AAAAAAAAAAAAAAAAA69675,AAAAAAAA50094,AAAAAAAAAAA92631,AAAA35194,39557,AAAAAAA99836}`}, {67, `{31,94,7,10}`, `{AAAAAA38063,A96617,AAAA35194,AAAAAAAAAAAA67946}`}, {68, `{90,43,38}`, `{AA75092,AAAAAAAAAAAAAAAAA69675,AAAAAAAAAAA92631,AAAAAAAAA10012,AAAAAAAAAAAAA7929,AA21643}`}, {69, `{67,35,99,85,72,86,44}`, `{AAAAAAAAAAAAAAAAAAA1205,AAAAAAAA50094,AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAAAAAAA47955}`}, {70, `{56,70,83}`, `{AAAA41702,AAAAAAAAAAA82945,AA21643,AAAAAAAAAAA99000,A27153,AA25381,AAAAAAAAAAAAAA96505,AAAAAAA1242}`}, {71, `{74,26}`, `{AAAAAAAAAAA50956,AA74433,AAAAAAA21462,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAA36627,AAAAAAAAAAAAA70254,AAAAAAAAAA43419,39557}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {73, `{88,25,96,78,65,15,29,19}`, `{AAA54451,AAAAAAAAA27249,AAAAAAA9228,AAAAAAAAAAAAAAA67062,AAAAAAAAAAAAAAAAAAA70415,AAAAA17383,AAAAAAAAAAAAAAAA33598}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {75, `{12,96,83,24,71,89,55}`, `{AAAA48949,AAAAAAAA29716,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAAA67946,AAAAAAAAAAAAAAAA29150,AAA28075,AAAAAAAAAAAAAAAAA43052}`}, {76, `{92,55,10,7}`, `{AAAAAAAAAAAAAAA67062}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {78, `{55,89,44,84,34}`, `{AAAAAAAAAAA6119,AAAAAAAAAAAAAA8666,AA99927,AA42406,AAAAAAA81898,AAAAAAA9228,AAAAAAAAAAA92631,AA21643,AAAAAAAAAAAAAA28620}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {80, `{74,89,44,80,0}`, `{AAAA35194,AAAAAAAA79710,AAA20874,AAAAAAAAAAAAAAAAAAA70104,AAAAAAAAAAAAA73084,AAAAAAA57334,AAAAAAA9228,AAAAAAAAAAAAA62007}`}, {81, `{63,77,54,48,61,53,97}`, `{AAAAAAAAAAAAAAA81326,AAAAAAAAAA22292,AA25381,AAAAAAAAAAA74076,AAAAAAA81898,AAAAAAAAA72121}`}, {82, `{34,60,4,79,78,16,86,89,42,50}`, `{AAAAA40681,AAAAAAAAAAAAAAAAAA12591,AAAAAAA80240,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA70104}`}, {83, `{14,10}`, `{AAAAAAAAAA22292,AAAAAAAAAAAAA70254,AAAAAAAAAAA6119}`}, {84, `{11,83,35,13,96,94}`, `{AAAAA95309,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAAAAA24183}`}, {85, `{39,60}`, `{AAAAAAAAAAAAAAAA55798,AAAAAAAAAA22292,AAAAAAA66161,AAAAAAA21462,AAAAAAAAAAAAAAAAAA12591,55847,AAAAAA98232,AAAAAAAAAAA46154}`}, {86, `{33,81,72,74,45,36,82}`, `{AAAAAAAA81587,AAAAAAAAAAAAAA96505,45449,AAAA80176}`}, {87, `{57,27,50,12,97,68}`, `{AAAAAAAAAAAAAAAAA26540,AAAAAAAAA10012,AAAAAAAAAAAA35809,AAAAAAAAAAAAAAAA29150,AAAAAAAAAAA82945,AAAAAA66777,31228,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAA96505}`}, {88, `{41,90,77,24,6,24}`, `{AAAA35194,AAAA35194,AAAAAAA80240,AAAAAAAAAAA46154,AAAAAA58494,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA59334,AAAAAAAAAAAAAAAAAAA91804,AA74433}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {90, `{88,75}`, `{AAAAA60038,AAAAAAAA23648,AAAAAAAAAAA99000,AAAA41702,AAAAAAAAAAAAA22860,AAAAAAAAAAAAAAA68526}`}, {91, `{78}`, `{AAAAAAAAAAAAA62007,AAA99043}`}, {92, `{85,63,49,45}`, `{AAAAAAA89932,AAAAAAAAAAAAA22860,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAAA21089}`}, {93, `{11}`, `{AAAAAAAAAAA176,AAAAAAAAAAAAAA8666,AAAAAAAAAAAAAAA453,AAAAAAAAAAAAA85723,A68938,AAAAAAAAAAAAA9821,AAAAAAA48038,AAAAAAAAAAAAAAAAA59387,AA99927,AAAAA17383}`}, {94, `{98,9,85,62,88,91,60,61,38,86}`, `{AAAAAAAA81587,AAAAA17383,AAAAAAAA81587}`}, {95, `{47,77}`, `{AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA74076,AAAAAAAAAA18107,AAAAA40681,AAAAAAAAAAAAAAA35875,AAAAA60038,AAAAAAA56483}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}, {97, `{54,2,86,65}`, `{47735,AAAAAAA99836,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAA29150,AAAAAAA80240,AAAAAAAAAAAAAAAA98414,AAAAAAA56483,AAAAAAAAAAAAAAAA29150,AAAAAAA39692,AA21643}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {99, `{37,86}`, `{AAAAAAAAAAAAAAAAAA32918,AAAAA70514,AAAAAAAAA10012,AAAAAAAAAAAAAAAAA59387,AAAAAAAAAA64777,AAAAAAAAAAAAAAAAAAA15356}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}, {101, `{}`, `{}`}, {102, `{NULL}`, `{NULL}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i && '{}' ORDER BY seqno;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i <@ '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i = '{NULL}' ORDER BY seqno;`,
				Results:   []sql.Row{{102, `{NULL}`, `{NULL}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i @> '{NULL}' ORDER BY seqno;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i && '{NULL}' ORDER BY seqno;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE i <@ '{NULL}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t @> '{AAAAAAAA72908}' ORDER BY seqno;`,
				Results:   []sql.Row{{22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t && '{AAAAAAAA72908}' ORDER BY seqno;`,
				Results:   []sql.Row{{22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t @> '{AAAAAAAAAA646}' ORDER BY seqno;`,
				Results:   []sql.Row{{15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t && '{AAAAAAAAAA646}' ORDER BY seqno;`,
				Results:   []sql.Row{{15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t @> '{AAAAAAAA72908,AAAAAAAAAA646}' ORDER BY seqno;`,
				Results:   []sql.Row{{79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t && '{AAAAAAAA72908,AAAAAAAAAA646}' ORDER BY seqno;`,
				Results:   []sql.Row{{15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t <@ '{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}' ORDER BY seqno;`,
				Results:   []sql.Row{{22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t = '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t @> '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{1, `{92,75,71,52,64,83}`, `{AAAAAAAA44066,AAAAAA1059,AAAAAAAAAAA176,AAAAAAA48038}`}, {2, `{3,6}`, `{AAAAAA98232,AAAAAAAA79710,AAAAAAAAAAAAAAAAA69675,AAAAAAAAAAAAAAAA55798,AAAAAAAAA12793}`}, {3, `{37,64,95,43,3,41,13,30,11,43}`, `{AAAAAAAAAA48845,AAAAA75968,AAAAA95309,AAA54451,AAAAAAAAAA22292,AAAAAAA99836,A96617,AA17009,AAAAAAAAAAAAAA95246}`}, {4, `{71,39,99,55,33,75,45}`, `{AAAAAAAAA53663,AAAAAAAAAAAAAAA67062,AAAAAAAAAA64777,AAA99043,AAAAAAAAAAAAAAAAAAA91804,39557}`}, {5, `{50,42,77,50,4}`, `{AAAAAAAAAAAAAAAAA26540,AAAAAAA79710,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA176,AAAAA95309,AAAAAAAAAAA46154,AAAAAA66777,AAAAAAAAA27249,AAAAAAAAAA64777,AAAAAAAAAAAAAAAAAAA70104}`}, {6, `{39,35,5,94,17,92,60,32}`, `{AAAAAAAAAAAAAAA35875,AAAAAAAAAAAAAAAA23657}`}, {7, `{12,51,88,64,8}`, `{AAAAAAAAAAAAAAAAAA12591,AAAAAAAAAAAAAAAAA50407,AAAAAAAAAAAA67946}`}, {8, `{60,84}`, `{AAAAAAA81898,AAAAAA1059,AAAAAAAAAAAA81511,AAAAA961,AAAAAAAAAAAAAAAA31334,AAAAA64741,AA6416,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAAAA50407}`}, {9, `{56,52,35,27,80,44,81,22}`, `{AAAAAAAAAAAAAAA73034,AAAAAAAAAAAAA7929,AAAAAAA66161,AA88409,39557,A27153,AAAAAAAA9523,AAAAAAAAAAA99000}`}, {10, `{71,5,45}`, `{AAAAAAAAAAA21658,AAAAAAAAAAAA21089,AAA54451,AAAAAAAAAAAAAAAAAA54141,AAAAAAAAAAAAAA28620,AAAAAAAAAAA21658,AAAAAAAAAAA74076,AAAAAAAAA27249}`}, {11, `{41,86,74,48,22,74,47,50}`, `{AAAAAAAA9523,AAAAAAAAAAAA37562,AAAAAAAAAAAAAAAA14047,AAAAAAAAAAA46154,AAAA41702,AAAAAAAAAAAAAAAAA764,AAAAA62737,39557}`}, {12, `{17,99,18,52,91,72,0,43,96,23}`, `{AAAAA33250,AAAAAAAAAAAAAAAAAAA85420,AAAAAAAAAAA33576}`}, {13, `{3,52,34,23}`, `{AAAAAA98232,AAAA49534,AAAAAAAAAAA21658}`}, {14, `{78,57,19}`, `{AAAA8857,AAAAAAAAAAAAAAA73034,AAAAAAAA81587,AAAAAAAAAAAAAAA68526,AAAAA75968,AAAAAAAAAAAAAA65909,AAAAAAAAA10012,AAAAAAAAAAAAAA65909}`}, {15, `{17,14,16,63,67}`, `{AA6416,AAAAAAAAAA646,AAAAA95309}`}, {16, `{14,63,85,11}`, `{AAAAAA66777}`}, {17, `{7,10,81,85}`, `{AAAAAA43678,AAAAAAA12144,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAAAAA15356}`}, {18, `{1}`, `{AAAAAAAAAAA33576,AAAAA95309,64261,AAA59323,AAAAAAAAAAAAAA95246,55847,AAAAAAAAAAAA67946,AAAAAAAAAAAAAAAAAA64374}`}, {19, `{52,82,17,74,23,46,69,51,75}`, `{AAAAAAAAAAAAA73084,AAAAA75968,AAAAAAAAAAAAAAAA14047,AAAAAAA80240,AAAAAAAAAAAAAAAAAAA1205,A68938}`}, {20, `{72,89,70,51,54,37,8,49,79}`, `{AAAAAA58494}`}, {21, `{2,8,65,10,5,79,43}`, `{AAAAAAAAAAAAAAAAA88852,AAAAAAAAAAAAAAAAAAA91804,AAAAA64669,AAAAAAAAAAAAAAAA1443,AAAAAAAAAAAAAAAA23657,AAAAA12179,AAAAAAAAAAAAAAAAA88852,AAAAAAAAAAAAAAAA31334,AAAAAAAAAAAAAAAA41303,AAAAAAAAAAAAAAAAAAA85420}`}, {22, `{11,6,56,62,53,30}`, `{AAAAAAAA72908}`}, {23, `{40,90,5,38,72,40,30,10,43,55}`, `{A6053,AAAAAAAAAAA6119,AA44673,AAAAAAAAAAAAAAAAA764,AA17009,AAAAA17383,AAAAA70514,AAAAA33250,AAAAA95309,AAAAAAAAAAAA37562}`}, {24, `{94,61,99,35,48}`, `{AAAAAAAAAAA50956,AAAAAAAAAAA15165,AAAA85070,AAAAAAAAAAAAAAA36627,AAAAA961,AAAAAAAAAA55219}`}, {25, `{31,1,10,11,27,79,38}`, `{AAAAAAAAAAAAAAAAAA59334,45449}`}, {26, `{71,10,9,69,75}`, `{47735,AAAAAAA21462,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAAAAA91804,AAAAAAAAA72121,AAAAAAAAAAAAAAAAAAA1205,AAAAA41597,AAAA8857,AAAAAAAAAAAAAAAAAAA15356,AA17009}`}, {27, `{94}`, `{AA6416,A6053,AAAAAAA21462,AAAAAAA57334,AAAAAAAAAAAAAAAAAA12591,AA88409,AAAAAAAAAAAAA70254}`}, {28, `{14,33,6,34,14}`, `{AAAAAAAAAAAAAAA13198,AAAAAAAA69452,AAAAAAAAAAA82945,AAAAAAA12144,AAAAAAAAA72121,AAAAAAAAAA18601}`}, {29, `{39,21}`, `{AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAAAAA38885,AAAA85070,AAAAAAAAAAAAAAAAAAA70104,AAAAA66674,AAAAAAAAAAAAA62007,AAAAAAAA69452,AAAAAAA1242,AAAAAAAAAAAAAAAA1729,AAAA35194}`}, {30, `{26,81,47,91,34}`, `{AAAAAAAAAAAAAAAAAAA70104,AAAAAAA80240}`}, {31, `{80,24,18,21,54}`, `{AAAAAAAAAAAAAAA13198,AAAAAAAAAAAAAAAAAAA70415,A27153,AAAAAAAAA53663,AAAAAAAAAAAAAAAAA50407,A68938}`}, {32, `{58,79,82,80,67,75,98,10,41}`, `{AAAAAAAAAAAAAAAAAA61286,AAA54451,AAAAAAAAAAAAAAAAAAA87527,A96617,51533}`}, {33, `{74,73}`, `{A85417,AAAAAAA56483,AAAAA17383,AAAAAAAAAAAAA62159,AAAAAAAAAAAA52814,AAAAAAAAAAAAA85723,AAAAAAAAAAAAAAAAAA55796}`}, {34, `{70,45}`, `{AAAAAAAAAAAAAAAAAA71621,AAAAAAAAAAAAAA28620,AAAAAAAAAA55219,AAAAAAAA23648,AAAAAAAAAA22292,AAAAAAA1242}`}, {35, `{23,40}`, `{AAAAAAAAAAAA52814,AAAA48949,AAAAAAAAA34727,AAAA8857,AAAAAAAAAAAAAAAAAAA62179,AAAAAAAAAAAAAAA68526,AAAAAAA99836,AAAAAAAA50094,AAAA91194,AAAAAAAAAAAAA73084}`}, {36, `{79,82,14,52,30,5,79}`, `{AAAAAAAAA53663,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA89194,AA88409,AAAAAAAAAAAAAAA81326,AAAAAAAAAAAAAAAAA63050,AAAAAAAAAAAAAAAA33598}`}, {37, `{53,11,81,39,3,78,58,64,74}`, `{AAAAAAAAAAAAAAAAAAA17075,AAAAAAA66161,AAAAAAAA23648,AAAAAAAAAAAAAA10611}`}, {38, `{59,5,4,95,28}`, `{AAAAAAAAAAA82945,A96617,47735,AAAAA12179,AAAAA64669,AAAAAA99807,AA74433,AAAAAAAAAAAAAAAAA59387}`}, {39, `{82,43,99,16,74}`, `{AAAAAAAAAAAAAAA67062,AAAAAAA57334,AAAAAAAAAAAAAA65909,A27153,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAA64777,AAAAAAAAAAAA81511,AAAAAAAAAAAAAA65909,AAAAAAAAAAAAAA28620}`}, {40, `{34}`, `{AAAAAAAAAAAAAA10611,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAA50956,AAAAAAAAAAAAAAAA31334,AAAAA70466,AAAAAAAA81587,AAAAAAA74623}`}, {41, `{19,26,63,12,93,73,27,94}`, `{AAAAAAA79710,AAAAAAAAAA55219,AAAA41702,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA71621,AAAAAAAAAAAAAAAAA63050,AAAAAAA99836,AAAAAAAAAAAAAA8666}`}, {42, `{15,76,82,75,8,91}`, `{AAAAAAAAAAA176,AAAAAA38063,45449,AAAAAA54032,AAAAAAA81898,AA6416,AAAAAAAAAAAAAAAAAAA62179,45449,AAAAA60038,AAAAAAAA81587}`}, {43, `{39,87,91,97,79,28}`, `{AAAAAAAAAAA74076,A96617,AAAAAAAAAAAAAAAAAAA89194,AAAAAAAAAAAAAAAAAA55796,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAA67946}`}, {44, `{40,58,68,29,54}`, `{AAAAAAA81898,AAAAAA66777,AAAAAA98232}`}, {45, `{99,45}`, `{AAAAAAAA72908,AAAAAAAAAAAAAAAAAAA17075,AA88409,AAAAAAAAAAAAAAAAAA36842,AAAAAAA48038,AAAAAAAAAAAAAA10611}`}, {46, `{53,24}`, `{AAAAAAAAAAA53908,AAAAAA54032,AAAAA17383,AAAA48949,AAAAAAAAAA18601,AAAAA64669,45449,AAAAAAAAAAA98051,AAAAAAAAAAAAAAAAAA71621}`}, {47, `{98,23,64,12,75,61}`, `{AAA59323,AAAAA95309,AAAAAAAAAAAAAAAA31334,AAAAAAAAA27249,AAAAA17383,AAAAAAAAAAAA37562,AAAAAA1059,A84822,55847,AAAAA70466}`}, {48, `{76,14}`, `{AAAAAAAAAAAAA59671,AAAAAAAAAAAAAAAAAAA91804,AAAAAA66777,AAAAAAAAAAAAAAAAAAA89194,AAAAAAAAAAAAAAA36627,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAA73084,AAAAAAA79710,AAAAAAAAAAAAAAA40402,AAAAAAAAAAAAAAAAAAA65037}`}, {49, `{56,5,54,37,49}`, `{AA21643,AAAAAAAAAAA92631,AAAAAAAA81587}`}, {50, `{20,12,37,64,93}`, `{AAAAAAAAAA5483,AAAAAAAAAAAAAAAAAAA1205,AA6416,AAAAAAAAAAAAAAAAA63050,AAAAAAAAAAAAAAAAAA47955}`}, {51, `{47}`, `{AAAAAAAAAAAAAA96505,AAAAAAAAAAAAAAAAAA36842,AAAAA95309,AAAAAAAA81587,AA6416,AAAA91194,AAAAAA58494,AAAAAA1059,AAAAAAAA69452}`}, {52, `{89,0}`, `{AAAAAAAAAAAAAAAAAA47955,AAAAAAA48038,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAAAAA73084,AAAAA70466,AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA46154,AA66862}`}, {53, `{38,17}`, `{AAAAAAAAAAA21658}`}, {54, `{70,47}`, `{AAAAAAAAAAAAAAAAAA54141,AAAAA40681,AAAAAAA48038,AAAAAAAAAAAAAAAA29150,AAAAA41597,AAAAAAAAAAAAAAAAAA59334,AA15322}`}, {55, `{47,79,47,64,72,25,71,24,93}`, `{AAAAAAAAAAAAAAAAAA55796,AAAAA62737}`}, {56, `{33,7,60,54,93,90,77,85,39}`, `{AAAAAAAAAAAAAAAAAA32918,AA42406}`}, {57, `{23,45,10,42,36,21,9,96}`, `{AAAAAAAAAAAAAAAAAAA70415}`}, {58, `{92}`, `{AAAAAAAAAAAAAAAA98414,AAAAAAAA23648,AAAAAAAAAAAAAAAAAA55796,AA25381,AAAAAAAAAAA6119}`}, {59, `{9,69,46,77}`, `{39557,AAAAAAA89932,AAAAAAAAAAAAAAAAA43052,AAAAAAAAAAAAAAAAA26540,AAA20874,AA6416,AAAAAAAAAAAAAAAAAA47955}`}, {60, `{62,2,59,38,89}`, `{AAAAAAA89932,AAAAAAAAAAAAAAAAAAA15356,AA99927,AA17009,AAAAAAAAAAAAAAA35875}`}, {61, `{72,2,44,95,54,54,13}`, `{AAAAAAAAAAAAAAAAAAA91804}`}, {62, `{83,72,29,73}`, `{AAAAAAAAAAAAA15097,AAAA8857,AAAAAAAAAAAA35809,AAAAAAAAAAAA52814,AAAAAAAAAAAAAAAAAAA38885,AAAAAAAAAAAAAAAAAA24183,AAAAAA43678,A96617}`}, {63, `{11,4,61,87}`, `{AAAAAAAAA27249,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAA13198,AAA20874,39557,51533,AAAAAAAAAAA53908,AAAAAAAAAAAAAA96505,AAAAAAAA78938}`}, {64, `{26,19,34,24,81,78}`, `{A96617,AAAAAAAAAAAAAAAAAAA70104,A68938,AAAAAAAAAAA53908,AAAAAAAAAAAAAAA453,AA17009,AAAAAAA80240}`}, {65, `{61,5,76,59,17}`, `{AAAAAA99807,AAAAA64741,AAAAAAAAAAA53908,AA21643,AAAAAAAAA10012}`}, {66, `{31,23,70,52,4,33,48,25}`, `{AAAAAAAAAAAAAAAAA69675,AAAAAAAA50094,AAAAAAAAAAA92631,AAAA35194,39557,AAAAAAA99836}`}, {67, `{31,94,7,10}`, `{AAAAAA38063,A96617,AAAA35194,AAAAAAAAAAAA67946}`}, {68, `{90,43,38}`, `{AA75092,AAAAAAAAAAAAAAAAA69675,AAAAAAAAAAA92631,AAAAAAAAA10012,AAAAAAAAAAAAA7929,AA21643}`}, {69, `{67,35,99,85,72,86,44}`, `{AAAAAAAAAAAAAAAAAAA1205,AAAAAAAA50094,AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAAAAAAA47955}`}, {70, `{56,70,83}`, `{AAAA41702,AAAAAAAAAAA82945,AA21643,AAAAAAAAAAA99000,A27153,AA25381,AAAAAAAAAAAAAA96505,AAAAAAA1242}`}, {71, `{74,26}`, `{AAAAAAAAAAA50956,AA74433,AAAAAAA21462,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAA36627,AAAAAAAAAAAAA70254,AAAAAAAAAA43419,39557}`}, {72, `{22,1,16,78,20,91,83}`, `{47735,AAAAAAA56483,AAAAAAAAAAAAA93788,AA42406,AAAAAAAAAAAAA73084,AAAAAAAA72908,AAAAAAAAAAAAAAAAAA61286,AAAAA66674,AAAAAAAAAAAAAAAAA50407}`}, {73, `{88,25,96,78,65,15,29,19}`, `{AAA54451,AAAAAAAAA27249,AAAAAAA9228,AAAAAAAAAAAAAAA67062,AAAAAAAAAAAAAAAAAAA70415,AAAAA17383,AAAAAAAAAAAAAAAA33598}`}, {74, `{32}`, `{AAAAAAAAAAAAAAAA1729,AAAAAAAAAAAAA22860,AAAAAA99807,AAAAA17383,AAAAAAAAAAAAAAA67062,AAAAAAAAAAA15165,AAAAAAAAAAA50956}`}, {75, `{12,96,83,24,71,89,55}`, `{AAAA48949,AAAAAAAA29716,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAAA67946,AAAAAAAAAAAAAAAA29150,AAA28075,AAAAAAAAAAAAAAAAA43052}`}, {76, `{92,55,10,7}`, `{AAAAAAAAAAAAAAA67062}`}, {77, `{97,15,32,17,55,59,18,37,50,39}`, `{AAAAAAAAAAAA67946,AAAAAA54032,AAAAAAAA81587,55847,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAAAAA43052,AAAAAA75463,AAAA49534,AAAAAAAA44066}`}, {78, `{55,89,44,84,34}`, `{AAAAAAAAAAA6119,AAAAAAAAAAAAAA8666,AA99927,AA42406,AAAAAAA81898,AAAAAAA9228,AAAAAAAAAAA92631,AA21643,AAAAAAAAAAAAAA28620}`}, {79, `{45}`, `{AAAAAAAAAA646,AAAAAAAAAAAAAAAAAAA70415,AAAAAA43678,AAAAAAAA72908}`}, {80, `{74,89,44,80,0}`, `{AAAA35194,AAAAAAAA79710,AAA20874,AAAAAAAAAAAAAAAAAAA70104,AAAAAAAAAAAAA73084,AAAAAAA57334,AAAAAAA9228,AAAAAAAAAAAAA62007}`}, {81, `{63,77,54,48,61,53,97}`, `{AAAAAAAAAAAAAAA81326,AAAAAAAAAA22292,AA25381,AAAAAAAAAAA74076,AAAAAAA81898,AAAAAAAAA72121}`}, {82, `{34,60,4,79,78,16,86,89,42,50}`, `{AAAAA40681,AAAAAAAAAAAAAAAAAA12591,AAAAAAA80240,AAAAAAAAAAAAAAAA55798,AAAAAAAAAAAAAAAAAAA70104}`}, {83, `{14,10}`, `{AAAAAAAAAA22292,AAAAAAAAAAAAA70254,AAAAAAAAAAA6119}`}, {84, `{11,83,35,13,96,94}`, `{AAAAA95309,AAAAAAAAAAAAAAAAAA32918,AAAAAAAAAAAAAAAAAA24183}`}, {85, `{39,60}`, `{AAAAAAAAAAAAAAAA55798,AAAAAAAAAA22292,AAAAAAA66161,AAAAAAA21462,AAAAAAAAAAAAAAAAAA12591,55847,AAAAAA98232,AAAAAAAAAAA46154}`}, {86, `{33,81,72,74,45,36,82}`, `{AAAAAAAA81587,AAAAAAAAAAAAAA96505,45449,AAAA80176}`}, {87, `{57,27,50,12,97,68}`, `{AAAAAAAAAAAAAAAAA26540,AAAAAAAAA10012,AAAAAAAAAAAA35809,AAAAAAAAAAAAAAAA29150,AAAAAAAAAAA82945,AAAAAA66777,31228,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAA28620,AAAAAAAAAAAAAA96505}`}, {88, `{41,90,77,24,6,24}`, `{AAAA35194,AAAA35194,AAAAAAA80240,AAAAAAAAAAA46154,AAAAAA58494,AAAAAAAAAAAAAAAAAAA17075,AAAAAAAAAAAAAAAAAA59334,AAAAAAAAAAAAAAAAAAA91804,AA74433}`}, {89, `{40,32,17,6,30,88}`, `{AA44673,AAAAAAAAAAA6119,AAAAAAAAAAAAAAAA23657,AAAAAAAAAAAAAAAAAA47955,AAAAAAAAAAAAAAAA33598,AAAAAAAAAAA33576,AA44673}`}, {90, `{88,75}`, `{AAAAA60038,AAAAAAAA23648,AAAAAAAAAAA99000,AAAA41702,AAAAAAAAAAAAA22860,AAAAAAAAAAAAAAA68526}`}, {91, `{78}`, `{AAAAAAAAAAAAA62007,AAA99043}`}, {92, `{85,63,49,45}`, `{AAAAAAA89932,AAAAAAAAAAAAA22860,AAAAAAAAAAAAAAAAAAA1205,AAAAAAAAAAAA21089}`}, {93, `{11}`, `{AAAAAAAAAAA176,AAAAAAAAAAAAAA8666,AAAAAAAAAAAAAAA453,AAAAAAAAAAAAA85723,A68938,AAAAAAAAAAAAA9821,AAAAAAA48038,AAAAAAAAAAAAAAAAA59387,AA99927,AAAAA17383}`}, {94, `{98,9,85,62,88,91,60,61,38,86}`, `{AAAAAAAA81587,AAAAA17383,AAAAAAAA81587}`}, {95, `{47,77}`, `{AAAAAAAAAAAAAAAAA764,AAAAAAAAAAA74076,AAAAAAAAAA18107,AAAAA40681,AAAAAAAAAAAAAAA35875,AAAAA60038,AAAAAAA56483}`}, {96, `{23,97,43}`, `{AAAAAAAAAA646,A87088}`}, {97, `{54,2,86,65}`, `{47735,AAAAAAA99836,AAAAAAAAAAAAAAAAA6897,AAAAAAAAAAAAAAAA29150,AAAAAAA80240,AAAAAAAAAAAAAAAA98414,AAAAAAA56483,AAAAAAAAAAAAAAAA29150,AAAAAAA39692,AA21643}`}, {98, `{38,34,32,89}`, `{AAAAAAAAAAAAAAAAAA71621,AAAA8857,AAAAAAAAAAAAAAAAAAA65037,AAAAAAAAAAAAAAAA31334,AAAAAAAAAA48845}`}, {99, `{37,86}`, `{AAAAAAAAAAAAAAAAAA32918,AAAAA70514,AAAAAAAAA10012,AAAAAAAAAAAAAAAAA59387,AAAAAAAAAA64777,AAAAAAAAAAAAAAAAAAA15356}`}, {100, `{85,32,57,39,49,84,32,3,30}`, `{AAAAAAA80240,AAAAAAAAAAAAAAAA1729,AAAAA60038,AAAAAAAAAAA92631,AAAAAAAA9523}`}, {101, `{}`, `{}`}, {102, `{NULL}`, `{NULL}`}},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t && '{}' ORDER BY seqno;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM array_op_test WHERE t <@ '{}' ORDER BY seqno;`,
				Results:   []sql.Row{{101, `{}`, `{}`}},
			},
			{
				Statement: `SELECT ARRAY[1,2,3]::text[]::int[]::float8[] AS "{1,2,3}";`,
				Results:   []sql.Row{{`{1,2,3}`}},
			},
			{
				Statement: `SELECT pg_typeof(ARRAY[1,2,3]::text[]::int[]::float8[]) AS "double precision[]";`,
				Results:   []sql.Row{{`double precision[]`}},
			},
			{
				Statement: `SELECT ARRAY[['a','bc'],['def','hijk']]::text[]::varchar[] AS "{{a,bc},{def,hijk}}";`,
				Results:   []sql.Row{{`{{a,bc},{def,hijk}}`}},
			},
			{
				Statement: `SELECT pg_typeof(ARRAY[['a','bc'],['def','hijk']]::text[]::varchar[]) AS "character varying[]";`,
				Results:   []sql.Row{{`character varying[]`}},
			},
			{
				Statement: `SELECT CAST(ARRAY[[[[[['a','bb','ccc']]]]]] as text[]) as "{{{{{{a,bb,ccc}}}}}}";`,
				Results:   []sql.Row{{`{{{{{{a,bb,ccc}}}}}}`}},
			},
			{
				Statement: `SELECT NULL::text[]::int[] AS "NULL";`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select 33 = any ('{1,2,3}');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 33 = any ('{1,2,33}');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 33 = all ('{1,2,33}');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 33 >= all ('{1,2,33}');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select null::int >= all ('{1,2,33}');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select null::int >= all ('{}');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select null::int >= any ('{}');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 33.4 = any (array[1,2,3]);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 33.4 > all (array[1,2,3]);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select 33 * any ('{1,2,3}');`,
				ErrorString: `op ANY/ALL (array) requires operator to yield boolean`,
			},
			{
				Statement:   `select 33 * any (44);`,
				ErrorString: `op ANY/ALL (array) requires array on right side`,
			},
			{
				Statement: `select 33 = any (null::int[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select null::int = any ('{1,2,3}');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select 33 = any ('{1,null,3}');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select 33 = any ('{1,null,33}');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 33 = all (null::int[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select null::int = all ('{1,2,3}');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select 33 = all ('{1,null,3}');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 33 = all ('{33,null,33}');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT -1 != ALL(ARRAY(SELECT NULLIF(g.i, 900) FROM generate_series(1,1000) g(i)));`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create temp table arr_tbl (f1 int[] unique);`,
			},
			{
				Statement: `insert into arr_tbl values ('{1,2,3}');`,
			},
			{
				Statement: `insert into arr_tbl values ('{1,2}');`,
			},
			{
				Statement:   `insert into arr_tbl values ('{1,2,3}');`,
				ErrorString: `duplicate key value violates unique constraint "arr_tbl_f1_key"`,
			},
			{
				Statement: `insert into arr_tbl values ('{2,3,4}');`,
			},
			{
				Statement: `insert into arr_tbl values ('{1,5,3}');`,
			},
			{
				Statement: `insert into arr_tbl values ('{1,2,10}');`,
			},
			{
				Statement: `set enable_seqscan to off;`,
			},
			{
				Statement: `set enable_bitmapscan to off;`,
			},
			{
				Statement: `select * from arr_tbl where f1 > '{1,2,3}' and f1 <= '{1,5,3}';`,
				Results:   []sql.Row{{`{1,2,10}`}, {`{1,5,3}`}},
			},
			{
				Statement: `select * from arr_tbl where f1 >= '{1,2,3}' and f1 < '{1,5,3}';`,
				Results:   []sql.Row{{`{1,2,3}`}, {`{1,2,10}`}},
			},
			{
				Statement: `create temp table arr_pk_tbl (pk int4 primary key, f1 int[]);`,
			},
			{
				Statement: `insert into arr_pk_tbl values (1, '{1,2,3}');`,
			},
			{
				Statement: `insert into arr_pk_tbl values (1, '{3,4,5}') on conflict (pk)
  do update set f1[1] = excluded.f1[1], f1[3] = excluded.f1[3]
  returning pk, f1;`,
				Results: []sql.Row{{1, `{3,2,5}`}},
			},
			{
				Statement: `insert into arr_pk_tbl(pk, f1[1:2]) values (1, '{6,7,8}') on conflict (pk)
  do update set f1[1] = excluded.f1[1],
    f1[2] = excluded.f1[2],
    f1[3] = excluded.f1[3]
  returning pk, f1;`,
				Results: []sql.Row{{1, `{6,7,NULL}`}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `select 'foo' like any (array['%a', '%o']); -- t`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'foo' like any (array['%a', '%b']); -- f`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'foo' like all (array['f%', '%o']); -- t`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'foo' like all (array['f%', '%b']); -- f`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'foo' not like any (array['%a', '%b']); -- t`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'foo' not like all (array['%a', '%o']); -- f`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'foo' ilike any (array['%A', '%O']); -- t`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'foo' ilike all (array['F%', '%O']); -- t`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select '{{1,{2}},{2,3}}'::text[];`,
				ErrorString: `malformed array literal: "{{1,{2}},{2,3}}"`,
			},
			{
				Statement:   `select '{{},{}}'::text[];`,
				ErrorString: `malformed array literal: "{{},{}}"`,
			},
			{
				Statement:   `select E'{{1,2},\\{2,3}}'::text[];`,
				ErrorString: `malformed array literal: "{{1,2},\{2,3}}"`,
			},
			{
				Statement:   `select '{{"1 2" x},{3}}'::text[];`,
				ErrorString: `malformed array literal: "{{"1 2" x},{3}}"`,
			},
			{
				Statement:   `select '{}}'::text[];`,
				ErrorString: `malformed array literal: "{}}"`,
			},
			{
				Statement:   `select '{ }}'::text[];`,
				ErrorString: `malformed array literal: "{ }}"`,
			},
			{
				Statement:   `select array[];`,
				ErrorString: `cannot determine type of empty array`,
			},
			{
				Statement: `select '{}'::text[];`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select '{{{1,2,3,4},{2,3,4,5}},{{3,4,5,6},{4,5,6,7}}}'::text[];`,
				Results:   []sql.Row{{`{{{1,2,3,4},{2,3,4,5}},{{3,4,5,6},{4,5,6,7}}}`}},
			},
			{
				Statement: `select '{0 second  ,0 second}'::interval[];`,
				Results:   []sql.Row{{`{"@ 0","@ 0"}`}},
			},
			{
				Statement: `select '{ { "," } , { 3 } }'::text[];`,
				Results:   []sql.Row{{`{{","},{3}}`}},
			},
			{
				Statement: `select '  {   {  "  0 second  "   ,  0 second  }   }'::text[];`,
				Results:   []sql.Row{{`{{"  0 second  ","0 second"}}`}},
			},
			{
				Statement: `select '{
           0 second,
           @ 1 hour @ 42 minutes @ 20 seconds
         }'::interval[];`,
				Results: []sql.Row{{`{"@ 0","@ 1 hour 42 mins 20 secs"}`}},
			},
			{
				Statement: `select array[]::text[];`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select '[0:1]={1.1,2.2}'::float8[];`,
				Results:   []sql.Row{{`[0:1]={1.1,2.2}`}},
			},
			{
				Statement: `CREATE TEMP TABLE arraggtest ( f1 INT[], f2 TEXT[][], f3 FLOAT[]);`,
			},
			{
				Statement: `INSERT INTO arraggtest (f1, f2, f3) VALUES
('{1,2,3,4}','{{grey,red},{blue,blue}}','{1.6, 0.0}');`,
			},
			{
				Statement: `INSERT INTO arraggtest (f1, f2, f3) VALUES
('{1,2,3}','{{grey,red},{grey,blue}}','{1.6}');`,
			},
			{
				Statement: `SELECT max(f1), min(f1), max(f2), min(f2), max(f3), min(f3) FROM arraggtest;`,
				Results:   []sql.Row{{`{1,2,3,4}`, `{1,2,3}`, `{{grey,red},{grey,blue}}`, `{{grey,red},{blue,blue}}`, `{1.6,0}`, `{1.6}`}},
			},
			{
				Statement: `INSERT INTO arraggtest (f1, f2, f3) VALUES
('{3,3,2,4,5,6}','{{white,yellow},{pink,orange}}','{2.1,3.3,1.8,1.7,1.6}');`,
			},
			{
				Statement: `SELECT max(f1), min(f1), max(f2), min(f2), max(f3), min(f3) FROM arraggtest;`,
				Results:   []sql.Row{{`{3,3,2,4,5,6}`, `{1,2,3}`, `{{white,yellow},{pink,orange}}`, `{{grey,red},{blue,blue}}`, `{2.1,3.3,1.8,1.7,1.6}`, `{1.6}`}},
			},
			{
				Statement: `INSERT INTO arraggtest (f1, f2, f3) VALUES
('{2}','{{black,red},{green,orange}}','{1.6,2.2,2.6,0.4}');`,
			},
			{
				Statement: `SELECT max(f1), min(f1), max(f2), min(f2), max(f3), min(f3) FROM arraggtest;`,
				Results:   []sql.Row{{`{3,3,2,4,5,6}`, `{1,2,3}`, `{{white,yellow},{pink,orange}}`, `{{black,red},{green,orange}}`, `{2.1,3.3,1.8,1.7,1.6}`, `{1.6}`}},
			},
			{
				Statement: `INSERT INTO arraggtest (f1, f2, f3) VALUES
('{4,2,6,7,8,1}','{{red},{black},{purple},{blue},{blue}}',NULL);`,
			},
			{
				Statement: `SELECT max(f1), min(f1), max(f2), min(f2), max(f3), min(f3) FROM arraggtest;`,
				Results:   []sql.Row{{`{4,2,6,7,8,1}`, `{1,2,3}`, `{{white,yellow},{pink,orange}}`, `{{black,red},{green,orange}}`, `{2.1,3.3,1.8,1.7,1.6}`, `{1.6}`}},
			},
			{
				Statement: `INSERT INTO arraggtest (f1, f2, f3) VALUES
('{}','{{pink,white,blue,red,grey,orange}}','{2.1,1.87,1.4,2.2}');`,
			},
			{
				Statement: `SELECT max(f1), min(f1), max(f2), min(f2), max(f3), min(f3) FROM arraggtest;`,
				Results:   []sql.Row{{`{4,2,6,7,8,1}`, `{}`, `{{white,yellow},{pink,orange}}`, `{{black,red},{green,orange}}`, `{2.1,3.3,1.8,1.7,1.6}`, `{1.6}`}},
			},
			{
				Statement: `create type comptype as (f1 int, f2 text);`,
			},
			{
				Statement: `create table comptable (c1 comptype, c2 comptype[]);`,
			},
			{
				Statement: `insert into comptable
  values (row(1,'foo'), array[row(2,'bar')::comptype, row(3,'baz')::comptype]);`,
			},
			{
				Statement: `create type _comptype as enum('fooey');`,
			},
			{
				Statement: `select * from comptable;`,
				Results:   []sql.Row{{`(1,foo)`, `{"(2,bar)","(3,baz)"}`}},
			},
			{
				Statement: `select c2[2].f2 from comptable;`,
				Results:   []sql.Row{{`baz`}},
			},
			{
				Statement: `drop type _comptype;`,
			},
			{
				Statement: `drop table comptable;`,
			},
			{
				Statement: `drop type comptype;`,
			},
			{
				Statement: `create or replace function unnest1(anyarray)
returns setof anyelement as $$
select $1[s] from generate_subscripts($1,1) g(s);`,
			},
			{
				Statement: `$$ language sql immutable;`,
			},
			{
				Statement: `create or replace function unnest2(anyarray)
returns setof anyelement as $$
select $1[s1][s2] from generate_subscripts($1,1) g1(s1),
                   generate_subscripts($1,2) g2(s2);`,
			},
			{
				Statement: `$$ language sql immutable;`,
			},
			{
				Statement: `select * from unnest1(array[1,2,3]);`,
				Results:   []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement: `select * from unnest2(array[[1,2,3],[4,5,6]]);`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}, {5}, {6}},
			},
			{
				Statement: `drop function unnest1(anyarray);`,
			},
			{
				Statement: `drop function unnest2(anyarray);`,
			},
			{
				Statement: `select array_fill(null::integer, array[3,3],array[2,2]);`,
				Results:   []sql.Row{{`[2:4][2:4]={{NULL,NULL,NULL},{NULL,NULL,NULL},{NULL,NULL,NULL}}`}},
			},
			{
				Statement: `select array_fill(null::integer, array[3,3]);`,
				Results:   []sql.Row{{`{{NULL,NULL,NULL},{NULL,NULL,NULL},{NULL,NULL,NULL}}`}},
			},
			{
				Statement: `select array_fill(null::text, array[3,3],array[2,2]);`,
				Results:   []sql.Row{{`[2:4][2:4]={{NULL,NULL,NULL},{NULL,NULL,NULL},{NULL,NULL,NULL}}`}},
			},
			{
				Statement: `select array_fill(null::text, array[3,3]);`,
				Results:   []sql.Row{{`{{NULL,NULL,NULL},{NULL,NULL,NULL},{NULL,NULL,NULL}}`}},
			},
			{
				Statement: `select array_fill(7, array[3,3],array[2,2]);`,
				Results:   []sql.Row{{`[2:4][2:4]={{7,7,7},{7,7,7},{7,7,7}}`}},
			},
			{
				Statement: `select array_fill(7, array[3,3]);`,
				Results:   []sql.Row{{`{{7,7,7},{7,7,7},{7,7,7}}`}},
			},
			{
				Statement: `select array_fill('juhu'::text, array[3,3],array[2,2]);`,
				Results:   []sql.Row{{`[2:4][2:4]={{juhu,juhu,juhu},{juhu,juhu,juhu},{juhu,juhu,juhu}}`}},
			},
			{
				Statement: `select array_fill('juhu'::text, array[3,3]);`,
				Results:   []sql.Row{{`{{juhu,juhu,juhu},{juhu,juhu,juhu},{juhu,juhu,juhu}}`}},
			},
			{
				Statement: `select a, a = '{}' as is_eq, array_dims(a)
  from (select array_fill(42, array[0]) as a) ss;`,
				Results: []sql.Row{{`{}`, true, ``}},
			},
			{
				Statement: `select a, a = '{}' as is_eq, array_dims(a)
  from (select array_fill(42, '{}') as a) ss;`,
				Results: []sql.Row{{`{}`, true, ``}},
			},
			{
				Statement: `select a, a = '{}' as is_eq, array_dims(a)
  from (select array_fill(42, '{}', '{}') as a) ss;`,
				Results: []sql.Row{{`{}`, true, ``}},
			},
			{
				Statement:   `select array_fill(1, null, array[2,2]);`,
				ErrorString: `dimension array or low bound array cannot be null`,
			},
			{
				Statement:   `select array_fill(1, array[2,2], null);`,
				ErrorString: `dimension array or low bound array cannot be null`,
			},
			{
				Statement:   `select array_fill(1, array[2,2], '{}');`,
				ErrorString: `wrong number of array subscripts`,
			},
			{
				Statement:   `select array_fill(1, array[3,3], array[1,1,1]);`,
				ErrorString: `wrong number of array subscripts`,
			},
			{
				Statement:   `select array_fill(1, array[1,2,null]);`,
				ErrorString: `dimension values cannot be null`,
			},
			{
				Statement:   `select array_fill(1, array[[1,2],[3,4]]);`,
				ErrorString: `wrong number of array subscripts`,
			},
			{
				Statement: `select string_to_array('1|2|3', '|');`,
				Results:   []sql.Row{{`{1,2,3}`}},
			},
			{
				Statement: `select string_to_array('1|2|3|', '|');`,
				Results:   []sql.Row{{`{1,2,3,""}`}},
			},
			{
				Statement: `select string_to_array('1||2|3||', '||');`,
				Results:   []sql.Row{{`{1,2|3,""}`}},
			},
			{
				Statement: `select string_to_array('1|2|3', '');`,
				Results:   []sql.Row{{`{1|2|3}`}},
			},
			{
				Statement: `select string_to_array('', '|');`,
				Results:   []sql.Row{{`{}`}},
			},
			{
				Statement: `select string_to_array('1|2|3', NULL);`,
				Results:   []sql.Row{{`{1,|,2,|,3}`}},
			},
			{
				Statement: `select string_to_array(NULL, '|') IS NULL;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select string_to_array('abc', '');`,
				Results:   []sql.Row{{`{abc}`}},
			},
			{
				Statement: `select string_to_array('abc', '', 'abc');`,
				Results:   []sql.Row{{`{NULL}`}},
			},
			{
				Statement: `select string_to_array('abc', ',');`,
				Results:   []sql.Row{{`{abc}`}},
			},
			{
				Statement: `select string_to_array('abc', ',', 'abc');`,
				Results:   []sql.Row{{`{NULL}`}},
			},
			{
				Statement: `select string_to_array('1,2,3,4,,6', ',');`,
				Results:   []sql.Row{{`{1,2,3,4,"",6}`}},
			},
			{
				Statement: `select string_to_array('1,2,3,4,,6', ',', '');`,
				Results:   []sql.Row{{`{1,2,3,4,NULL,6}`}},
			},
			{
				Statement: `select string_to_array('1,2,3,4,*,6', ',', '*');`,
				Results:   []sql.Row{{`{1,2,3,4,NULL,6}`}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('1|2|3', '|') g(v);`,
				Results:   []sql.Row{{1, false}, {2, false}, {3, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('1|2|3|', '|') g(v);`,
				Results:   []sql.Row{{1, false}, {2, false}, {3, false}, {``, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('1||2|3||', '||') g(v);`,
				Results:   []sql.Row{{1, false}, {`2|3`, false}, {``, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('1|2|3', '') g(v);`,
				Results:   []sql.Row{{`1|2|3`, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('', '|') g(v);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('1|2|3', NULL) g(v);`,
				Results:   []sql.Row{{1, false}, {``, `| f`}, {2, false}, {``, `| f`}, {3, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table(NULL, '|') g(v);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('abc', '') g(v);`,
				Results:   []sql.Row{{`abc`, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('abc', '', 'abc') g(v);`,
				Results:   []sql.Row{{``, true}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('abc', ',') g(v);`,
				Results:   []sql.Row{{`abc`, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('abc', ',', 'abc') g(v);`,
				Results:   []sql.Row{{``, true}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('1,2,3,4,,6', ',') g(v);`,
				Results:   []sql.Row{{1, false}, {2, false}, {3, false}, {4, false}, {``, false}, {6, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('1,2,3,4,,6', ',', '') g(v);`,
				Results:   []sql.Row{{1, false}, {2, false}, {3, false}, {4, false}, {``, true}, {6, false}},
			},
			{
				Statement: `select v, v is null as "is null" from string_to_table('1,2,3,4,*,6', ',', '*') g(v);`,
				Results:   []sql.Row{{1, false}, {2, false}, {3, false}, {4, false}, {``, true}, {6, false}},
			},
			{
				Statement: `select array_to_string(NULL::int4[], ',') IS NULL;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select array_to_string('{}'::int4[], ',');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select array_to_string(array[1,2,3,4,NULL,6], ',');`,
				Results:   []sql.Row{{`1,2,3,4,6`}},
			},
			{
				Statement: `select array_to_string(array[1,2,3,4,NULL,6], ',', '*');`,
				Results:   []sql.Row{{`1,2,3,4,*,6`}},
			},
			{
				Statement: `select array_to_string(array[1,2,3,4,NULL,6], NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select array_to_string(array[1,2,3,4,NULL,6], ',', NULL);`,
				Results:   []sql.Row{{`1,2,3,4,6`}},
			},
			{
				Statement: `select array_to_string(string_to_array('1|2|3', '|'), '|');`,
				Results:   []sql.Row{{`1|2|3`}},
			},
			{
				Statement: `select array_length(array[1,2,3], 1);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select array_length(array[[1,2,3], [4,5,6]], 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select array_length(array[[1,2,3], [4,5,6]], 1);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select array_length(array[[1,2,3], [4,5,6]], 2);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select array_length(array[[1,2,3], [4,5,6]], 3);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select cardinality(NULL::int[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select cardinality('{}'::int[]);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `select cardinality(array[1,2,3]);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select cardinality('[2:4]={5,6,7}'::int[]);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select cardinality('{{1,2}}'::int[]);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `select cardinality('{{1,2},{3,4},{5,6}}'::int[]);`,
				Results:   []sql.Row{{6}},
			},
			{
				Statement: `select cardinality('{{{1,9},{5,6}},{{2,3},{3,4}}}'::int[]);`,
				Results:   []sql.Row{{8}},
			},
			{
				Statement: `select array_agg(unique1) from (select unique1 from tenk1 where unique1 < 15 order by unique1) ss;`,
				Results:   []sql.Row{{`{0,1,2,3,4,5,6,7,8,9,10,11,12,13,14}`}},
			},
			{
				Statement: `select array_agg(ten) from (select ten from tenk1 where unique1 < 15 order by unique1) ss;`,
				Results:   []sql.Row{{`{0,1,2,3,4,5,6,7,8,9,0,1,2,3,4}`}},
			},
			{
				Statement: `select array_agg(nullif(ten, 4)) from (select ten from tenk1 where unique1 < 15 order by unique1) ss;`,
				Results:   []sql.Row{{`{0,1,2,3,NULL,5,6,7,8,9,0,1,2,3,NULL}`}},
			},
			{
				Statement: `select array_agg(unique1) from tenk1 where unique1 < -15;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select array_agg(ar)
  from (values ('{1,2}'::int[]), ('{3,4}'::int[])) v(ar);`,
				Results: []sql.Row{{`{{1,2},{3,4}}`}},
			},
			{
				Statement: `select array_agg(distinct ar order by ar desc)
  from (select array[i / 2] from generate_series(1,10) a(i)) b(ar);`,
				Results: []sql.Row{{`{{5},{4},{3},{2},{1},{0}}`}},
			},
			{
				Statement: `select array_agg(ar)
  from (select array_agg(array[i, i+1, i-1])
        from generate_series(1,2) a(i)) b(ar);`,
				Results: []sql.Row{{`{{{1,2,0},{2,3,1}}}`}},
			},
			{
				Statement: `select array_agg(array[i+1.2, i+1.3, i+1.4]) from generate_series(1,3) g(i);`,
				Results:   []sql.Row{{`{{2.2,2.3,2.4},{3.2,3.3,3.4},{4.2,4.3,4.4}}`}},
			},
			{
				Statement: `select array_agg(array['Hello', i::text]) from generate_series(9,11) g(i);`,
				Results:   []sql.Row{{`{{Hello,9},{Hello,10},{Hello,11}}`}},
			},
			{
				Statement: `select array_agg(array[i, nullif(i, 3), i+1]) from generate_series(1,4) g(i);`,
				Results:   []sql.Row{{`{{1,1,2},{2,2,3},{3,NULL,4},{4,4,5}}`}},
			},
			{
				Statement:   `select array_agg('{}'::int[]) from generate_series(1,2);`,
				ErrorString: `cannot accumulate empty arrays`,
			},
			{
				Statement:   `select array_agg(null::int[]) from generate_series(1,2);`,
				ErrorString: `cannot accumulate null arrays`,
			},
			{
				Statement: `select array_agg(ar)
  from (values ('{1,2}'::int[]), ('{3}'::int[])) v(ar);`,
				ErrorString: `cannot accumulate arrays of different dimensionality`,
			},
			{
				Statement: `select unnest(array[1,2,3]);`,
				Results:   []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement: `select * from unnest(array[1,2,3]);`,
				Results:   []sql.Row{{1}, {2}, {3}},
			},
			{
				Statement: `select unnest(array[1,2,3,4.5]::float8[]);`,
				Results:   []sql.Row{{1}, {2}, {3}, {4.5}},
			},
			{
				Statement: `select unnest(array[1,2,3,4.5]::numeric[]);`,
				Results:   []sql.Row{{1}, {2}, {3}, {4.5}},
			},
			{
				Statement: `select unnest(array[1,2,3,null,4,null,null,5,6]);`,
				Results:   []sql.Row{{1}, {2}, {3}, {``}, {4}, {``}, {``}, {5}, {6}},
			},
			{
				Statement: `select unnest(array[1,2,3,null,4,null,null,5,6]::text[]);`,
				Results:   []sql.Row{{1}, {2}, {3}, {``}, {4}, {``}, {``}, {5}, {6}},
			},
			{
				Statement: `select abs(unnest(array[1,2,null,-3]));`,
				Results:   []sql.Row{{1}, {2}, {``}, {3}},
			},
			{
				Statement: `select array_remove(array[1,2,2,3], 2);`,
				Results:   []sql.Row{{`{1,3}`}},
			},
			{
				Statement: `select array_remove(array[1,2,2,3], 5);`,
				Results:   []sql.Row{{`{1,2,2,3}`}},
			},
			{
				Statement: `select array_remove(array[1,NULL,NULL,3], NULL);`,
				Results:   []sql.Row{{`{1,3}`}},
			},
			{
				Statement: `select array_remove(array['A','CC','D','C','RR'], 'RR');`,
				Results:   []sql.Row{{`{A,CC,D,C}`}},
			},
			{
				Statement: `select array_remove(array[1.0, 2.1, 3.3], 1);`,
				Results:   []sql.Row{{`{2.1,3.3}`}},
			},
			{
				Statement:   `select array_remove('{{1,2,2},{1,4,3}}', 2); -- not allowed`,
				ErrorString: `removing elements from multidimensional arrays is not supported`,
			},
			{
				Statement: `select array_remove(array['X','X','X'], 'X') = '{}';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select array_replace(array[1,2,5,4],5,3);`,
				Results:   []sql.Row{{`{1,2,3,4}`}},
			},
			{
				Statement: `select array_replace(array[1,2,5,4],5,NULL);`,
				Results:   []sql.Row{{`{1,2,NULL,4}`}},
			},
			{
				Statement: `select array_replace(array[1,2,NULL,4,NULL],NULL,5);`,
				Results:   []sql.Row{{`{1,2,5,4,5}`}},
			},
			{
				Statement: `select array_replace(array['A','B','DD','B'],'B','CC');`,
				Results:   []sql.Row{{`{A,CC,DD,CC}`}},
			},
			{
				Statement: `select array_replace(array[1,NULL,3],NULL,NULL);`,
				Results:   []sql.Row{{`{1,NULL,3}`}},
			},
			{
				Statement: `select array_replace(array['AB',NULL,'CDE'],NULL,'12');`,
				Results:   []sql.Row{{`{AB,12,CDE}`}},
			},
			{
				Statement: `select array(select array[i,i/2] from generate_series(1,5) i);`,
				Results:   []sql.Row{{`{{1,0},{2,1},{3,1},{4,2},{5,2}}`}},
			},
			{
				Statement: `select array(select array['Hello', i::text] from generate_series(9,11) i);`,
				Results:   []sql.Row{{`{{Hello,9},{Hello,10},{Hello,11}}`}},
			},
			{
				Statement: `create temp table t1 (f1 int8_tbl[]);`,
			},
			{
				Statement: `insert into t1 (f1[5].q1) values(42);`,
			},
			{
				Statement: `select * from t1;`,
				Results:   []sql.Row{{`[5:5]={"(42,)"}`}},
			},
			{
				Statement: `update t1 set f1[5].q2 = 43;`,
			},
			{
				Statement: `select * from t1;`,
				Results:   []sql.Row{{`[5:5]={"(42,43)"}`}},
			},
			{
				Statement: `create temp table src (f1 text);`,
			},
			{
				Statement: `insert into src
  select string_agg(random()::text,'') from generate_series(1,10000);`,
			},
			{
				Statement: `create type textandtext as (c1 text, c2 text);`,
			},
			{
				Statement: `create temp table dest (f1 textandtext[]);`,
			},
			{
				Statement: `insert into dest select array[row(f1,f1)::textandtext] from src;`,
			},
			{
				Statement: `select length(md5((f1[1]).c2)) from dest;`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement: `delete from src;`,
			},
			{
				Statement: `select length(md5((f1[1]).c2)) from dest;`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement: `truncate table src;`,
			},
			{
				Statement: `drop table src;`,
			},
			{
				Statement: `select length(md5((f1[1]).c2)) from dest;`,
				Results:   []sql.Row{{32}},
			},
			{
				Statement: `drop table dest;`,
			},
			{
				Statement: `drop type textandtext;`,
			},
			{
				Statement: `SELECT
    op,
    width_bucket(op::numeric, ARRAY[1, 3, 5, 10.0]::numeric[]) AS wb_n1,
    width_bucket(op::numeric, ARRAY[0, 5.5, 9.99]::numeric[]) AS wb_n2,
    width_bucket(op::numeric, ARRAY[-6, -5, 2.0]::numeric[]) AS wb_n3,
    width_bucket(op::float8, ARRAY[1, 3, 5, 10.0]::float8[]) AS wb_f1,
    width_bucket(op::float8, ARRAY[0, 5.5, 9.99]::float8[]) AS wb_f2,
    width_bucket(op::float8, ARRAY[-6, -5, 2.0]::float8[]) AS wb_f3
FROM (VALUES
  (-5.2),
  (-0.0000000001),
  (0.000000000001),
  (1),
  (1.99999999999999),
  (2),
  (2.00000000000001),
  (3),
  (4),
  (4.5),
  (5),
  (5.5),
  (6),
  (7),
  (8),
  (9),
  (9.99999999999999),
  (10),
  (10.0000000000001)
) v(op);`,
				Results: []sql.Row{{-5.2, 0, 0, 1, 0, 0, 1}, {-0.0000000001, 0, 0, 2, 0, 0, 2}, {0.000000000001, 0, 1, 2, 0, 1, 2}, {1, 1, 1, 2, 1, 1, 2}, {1.99999999999999, 1, 1, 2, 1, 1, 2}, {2, 1, 1, 3, 1, 1, 3}, {2.00000000000001, 1, 1, 3, 1, 1, 3}, {3, 2, 1, 3, 2, 1, 3}, {4, 2, 1, 3, 2, 1, 3}, {4.5, 2, 1, 3, 2, 1, 3}, {5, 3, 1, 3, 3, 1, 3}, {5.5, 3, 2, 3, 3, 2, 3}, {6, 3, 2, 3, 3, 2, 3}, {7, 3, 2, 3, 3, 2, 3}, {8, 3, 2, 3, 3, 2, 3}, {9, 3, 2, 3, 3, 2, 3}, {9.99999999999999, 3, 3, 3, 3, 3, 3}, {10, 4, 3, 3, 4, 3, 3}, {10.0000000000001, 4, 3, 3, 4, 3, 3}},
			},
			{
				Statement: `SELECT
    op,
    width_bucket(op, ARRAY[1, 3, 9, 'NaN', 'NaN']::float8[]) AS wb
FROM (VALUES
  (-5.2::float8),
  (4::float8),
  (77::float8),
  ('NaN'::float8)
) v(op);`,
				Results: []sql.Row{{-5.2, 0}, {4, 2}, {77, 3}, {`NaN`, 5}},
			},
			{
				Statement: `SELECT
    op,
    width_bucket(op, ARRAY[1, 3, 5, 10]) AS wb_1
FROM generate_series(0,11) as op;`,
				Results: []sql.Row{{0, 0}, {1, 1}, {2, 1}, {3, 2}, {4, 2}, {5, 3}, {6, 3}, {7, 3}, {8, 3}, {9, 3}, {10, 4}, {11, 4}},
			},
			{
				Statement: `SELECT width_bucket(now(),
                    array['yesterday', 'today', 'tomorrow']::timestamptz[]);`,
				Results: []sql.Row{{2}},
			},
			{
				Statement: `SELECT width_bucket(5, ARRAY[3]);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT width_bucket(5, '{}');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT width_bucket('5'::text, ARRAY[3, 4]::integer[]);`,
				ErrorString: `function width_bucket(text, integer[]) does not exist`,
			},
			{
				Statement:   `SELECT width_bucket(5, ARRAY[3, 4, NULL]);`,
				ErrorString: `thresholds array must not contain NULLs`,
			},
			{
				Statement:   `SELECT width_bucket(5, ARRAY[ARRAY[1, 2], ARRAY[3, 4]]);`,
				ErrorString: `thresholds must be one-dimensional array`,
			},
			{
				Statement: `SELECT arr, trim_array(arr, 2)
FROM
(VALUES ('{1,2,3,4,5,6}'::bigint[]),
        ('{1,2}'),
        ('[10:16]={1,2,3,4,5,6,7}'),
        ('[-15:-10]={1,2,3,4,5,6}'),
        ('{{1,10},{2,20},{3,30},{4,40}}')) v(arr);`,
				Results: []sql.Row{{`{1,2,3,4,5,6}`, `{1,2,3,4}`}, {`{1,2}`, `{}`}, {`[10:16]={1,2,3,4,5,6,7}`, `{1,2,3,4,5}`}, {`[-15:-10]={1,2,3,4,5,6}`, `{1,2,3,4}`}, {`{{1,10},{2,20},{3,30},{4,40}}`, `{{1,10},{2,20}}`}},
			},
			{
				Statement:   `SELECT trim_array(ARRAY[1, 2, 3], -1); -- fail`,
				ErrorString: `number of elements to trim must be between 0 and 3`,
			},
			{
				Statement:   `SELECT trim_array(ARRAY[1, 2, 3], 10); -- fail`,
				ErrorString: `number of elements to trim must be between 0 and 3`,
			},
			{
				Statement:   `SELECT trim_array(ARRAY[]::int[], 1); -- fail`,
				ErrorString: `number of elements to trim must be between 0 and 0`,
			},
		},
	})
}
