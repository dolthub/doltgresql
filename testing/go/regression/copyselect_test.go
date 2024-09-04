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

func TestCopyselect(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_copyselect)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_copyselect,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create table test1 (id serial, t text);`,
			},
			{
				Statement: `insert into test1 (t) values ('a');`,
			},
			{
				Statement: `insert into test1 (t) values ('b');`,
			},
			{
				Statement: `insert into test1 (t) values ('c');`,
			},
			{
				Statement: `insert into test1 (t) values ('d');`,
			},
			{
				Statement: `insert into test1 (t) values ('e');`,
			},
			{
				Statement: `create table test2 (id serial, t text);`,
			},
			{
				Statement: `insert into test2 (t) values ('A');`,
			},
			{
				Statement: `insert into test2 (t) values ('B');`,
			},
			{
				Statement: `insert into test2 (t) values ('C');`,
			},
			{
				Statement: `insert into test2 (t) values ('D');`,
			},
			{
				Statement: `insert into test2 (t) values ('E');`,
			},
			{
				Statement: `create view v_test1
as select 'v_'||t from test1;`,
			},
			{
				Statement: `copy test1 to stdout;`,
			},
			{
				Statement: `1	a
2	b
3	c
4	d
5	e
copy v_test1 to stdout;`,
				ErrorString: `cannot copy from view "v_test1"`,
			},
			{
				Statement: `copy (select t from test1 where id=1) to stdout;`,
			},
			{
				Statement: `a
copy (select t from test1 where id=3 for update) to stdout;`,
			},
			{
				Statement: `c
copy (select t into temp test3 from test1 where id=3) to stdout;`,
				ErrorString: `COPY (SELECT INTO) is not supported`,
			},
			{
				Statement:   `copy (select * from test1) from stdin;`,
				ErrorString: `syntax error at or near "from"`,
			},
			{
				Statement:   `copy (select * from test1) (t,id) to stdout;`,
				ErrorString: `syntax error at or near "("`,
			},
			{
				Statement: `copy (select * from test1 join test2 using (id)) to stdout;`,
			},
			{
				Statement: `1	a	A
2	b	B
3	c	C
4	d	D
5	e	E
copy (select t from test1 where id = 1 UNION select * from v_test1 ORDER BY 1) to stdout;`,
			},
			{
				Statement: `a
v_a
v_b
v_c
v_d
v_e
copy (select * from (select t from test1 where id = 1 UNION select * from v_test1 ORDER BY 1) t1) to stdout;`,
			},
			{
				Statement: `a
v_a
v_b
v_c
v_d
v_e
copy (select t from test1 where id = 1) to stdout csv header force quote t;`,
			},
			{
				Statement: `t
"a"
\copy test1 to stdout
1	a
2	b
3	c
4	d
5	e
\copy v_test1 to stdout
ERROR:  cannot copy from view "v_test1"
HINT:  Try the COPY (SELECT ...) TO variant.
\copy (select "id",'id','id""'||t,(id + 1)*id,t,"test1"."t" from test1 where id=3) to stdout
3	id	id""c	12	c	c
drop table test2;`,
			},
			{
				Statement: `drop view v_test1;`,
			},
			{
				Statement: `drop table test1;`,
			},
			{
				Statement: `copy (select 1) to stdout\; select 1/0;	-- row, then error`,
			},
			{
				Statement: `1
ERROR:  division by zero
select 1/0\; copy (select 1) to stdout; -- error only`,
				ErrorString: `division by zero`,
			},
			{
				Statement: `copy (select 1) to stdout\; copy (select 2) to stdout\; select 3\; select 4; -- 1 2 3 4`,
			},
			{
				Statement: `1
2
 ?column? 
----------
        3
(1 row)
 ?column? 
----------
        4
(1 row)
create table test3 (c int);`,
			},
			{
				Statement: `select 0\; copy test3 from stdin\; copy test3 from stdin\; select 1; -- 0 1`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: ` ?column? 
----------
        1
(1 row)
select * from test3;`,
				Results: []sql.Row{{1}, {2}},
			},
			{
				Statement: `drop table test3;`,
			},
		},
	})
}
