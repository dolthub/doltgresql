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

func TestInherit(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_inherit)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_inherit,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE a (aa TEXT);`,
			},
			{
				Statement: `CREATE TABLE b (bb TEXT) INHERITS (a);`,
			},
			{
				Statement: `CREATE TABLE c (cc TEXT) INHERITS (a);`,
			},
			{
				Statement: `CREATE TABLE d (dd TEXT) INHERITS (b,c,a);`,
			},
			{
				Statement: `INSERT INTO a(aa) VALUES('aaa');`,
			},
			{
				Statement: `INSERT INTO a(aa) VALUES('aaaa');`,
			},
			{
				Statement: `INSERT INTO a(aa) VALUES('aaaaa');`,
			},
			{
				Statement: `INSERT INTO a(aa) VALUES('aaaaaa');`,
			},
			{
				Statement: `INSERT INTO a(aa) VALUES('aaaaaaa');`,
			},
			{
				Statement: `INSERT INTO a(aa) VALUES('aaaaaaaa');`,
			},
			{
				Statement: `INSERT INTO b(aa) VALUES('bbb');`,
			},
			{
				Statement: `INSERT INTO b(aa) VALUES('bbbb');`,
			},
			{
				Statement: `INSERT INTO b(aa) VALUES('bbbbb');`,
			},
			{
				Statement: `INSERT INTO b(aa) VALUES('bbbbbb');`,
			},
			{
				Statement: `INSERT INTO b(aa) VALUES('bbbbbbb');`,
			},
			{
				Statement: `INSERT INTO b(aa) VALUES('bbbbbbbb');`,
			},
			{
				Statement: `INSERT INTO c(aa) VALUES('ccc');`,
			},
			{
				Statement: `INSERT INTO c(aa) VALUES('cccc');`,
			},
			{
				Statement: `INSERT INTO c(aa) VALUES('ccccc');`,
			},
			{
				Statement: `INSERT INTO c(aa) VALUES('cccccc');`,
			},
			{
				Statement: `INSERT INTO c(aa) VALUES('ccccccc');`,
			},
			{
				Statement: `INSERT INTO c(aa) VALUES('cccccccc');`,
			},
			{
				Statement: `INSERT INTO d(aa) VALUES('ddd');`,
			},
			{
				Statement: `INSERT INTO d(aa) VALUES('dddd');`,
			},
			{
				Statement: `INSERT INTO d(aa) VALUES('ddddd');`,
			},
			{
				Statement: `INSERT INTO d(aa) VALUES('dddddd');`,
			},
			{
				Statement: `INSERT INTO d(aa) VALUES('ddddddd');`,
			},
			{
				Statement: `INSERT INTO d(aa) VALUES('dddddddd');`,
			},
			{
				Statement: `SELECT relname, a.* FROM a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`a`, `aaa`}, {`a`, `aaaa`}, {`a`, `aaaaa`}, {`a`, `aaaaaa`}, {`a`, `aaaaaaa`}, {`a`, `aaaaaaaa`}, {`b`, `bbb`}, {`b`, `bbbb`}, {`b`, `bbbbb`}, {`b`, `bbbbbb`}, {`b`, `bbbbbbb`}, {`b`, `bbbbbbbb`}, {`c`, `ccc`}, {`c`, `cccc`}, {`c`, `ccccc`}, {`c`, `cccccc`}, {`c`, `ccccccc`}, {`c`, `cccccccc`}, {`d`, `ddd`}, {`d`, `dddd`}, {`d`, `ddddd`}, {`d`, `dddddd`}, {`d`, `ddddddd`}, {`d`, `dddddddd`}},
			},
			{
				Statement: `SELECT relname, b.* FROM b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`b`, `bbb`, ``}, {`b`, `bbbb`, ``}, {`b`, `bbbbb`, ``}, {`b`, `bbbbbb`, ``}, {`b`, `bbbbbbb`, ``}, {`b`, `bbbbbbbb`, ``}, {`d`, `ddd`, ``}, {`d`, `dddd`, ``}, {`d`, `ddddd`, ``}, {`d`, `dddddd`, ``}, {`d`, `ddddddd`, ``}, {`d`, `dddddddd`, ``}},
			},
			{
				Statement: `SELECT relname, c.* FROM c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`c`, `ccc`, ``}, {`c`, `cccc`, ``}, {`c`, `ccccc`, ``}, {`c`, `cccccc`, ``}, {`c`, `ccccccc`, ``}, {`c`, `cccccccc`, ``}, {`d`, `ddd`, ``}, {`d`, `dddd`, ``}, {`d`, `ddddd`, ``}, {`d`, `dddddd`, ``}, {`d`, `ddddddd`, ``}, {`d`, `dddddddd`, ``}},
			},
			{
				Statement: `SELECT relname, d.* FROM d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `ddd`, ``, ``, ``}, {`d`, `dddd`, ``, ``, ``}, {`d`, `ddddd`, ``, ``, ``}, {`d`, `dddddd`, ``, ``, ``}, {`d`, `ddddddd`, ``, ``, ``}, {`d`, `dddddddd`, ``, ``, ``}},
			},
			{
				Statement: `SELECT relname, a.* FROM ONLY a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`a`, `aaa`}, {`a`, `aaaa`}, {`a`, `aaaaa`}, {`a`, `aaaaaa`}, {`a`, `aaaaaaa`}, {`a`, `aaaaaaaa`}},
			},
			{
				Statement: `SELECT relname, b.* FROM ONLY b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`b`, `bbb`, ``}, {`b`, `bbbb`, ``}, {`b`, `bbbbb`, ``}, {`b`, `bbbbbb`, ``}, {`b`, `bbbbbbb`, ``}, {`b`, `bbbbbbbb`, ``}},
			},
			{
				Statement: `SELECT relname, c.* FROM ONLY c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`c`, `ccc`, ``}, {`c`, `cccc`, ``}, {`c`, `ccccc`, ``}, {`c`, `cccccc`, ``}, {`c`, `ccccccc`, ``}, {`c`, `cccccccc`, ``}},
			},
			{
				Statement: `SELECT relname, d.* FROM ONLY d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `ddd`, ``, ``, ``}, {`d`, `dddd`, ``, ``, ``}, {`d`, `ddddd`, ``, ``, ``}, {`d`, `dddddd`, ``, ``, ``}, {`d`, `ddddddd`, ``, ``, ``}, {`d`, `dddddddd`, ``, ``, ``}},
			},
			{
				Statement: `UPDATE a SET aa='zzzz' WHERE aa='aaaa';`,
			},
			{
				Statement: `UPDATE ONLY a SET aa='zzzzz' WHERE aa='aaaaa';`,
			},
			{
				Statement: `UPDATE b SET aa='zzz' WHERE aa='aaa';`,
			},
			{
				Statement: `UPDATE ONLY b SET aa='zzz' WHERE aa='aaa';`,
			},
			{
				Statement: `UPDATE a SET aa='zzzzzz' WHERE aa LIKE 'aaa%';`,
			},
			{
				Statement: `SELECT relname, a.* FROM a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`a`, `zzzz`}, {`a`, `zzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`b`, `bbb`}, {`b`, `bbbb`}, {`b`, `bbbbb`}, {`b`, `bbbbbb`}, {`b`, `bbbbbbb`}, {`b`, `bbbbbbbb`}, {`c`, `ccc`}, {`c`, `cccc`}, {`c`, `ccccc`}, {`c`, `cccccc`}, {`c`, `ccccccc`}, {`c`, `cccccccc`}, {`d`, `ddd`}, {`d`, `dddd`}, {`d`, `ddddd`}, {`d`, `dddddd`}, {`d`, `ddddddd`}, {`d`, `dddddddd`}},
			},
			{
				Statement: `SELECT relname, b.* FROM b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`b`, `bbb`, ``}, {`b`, `bbbb`, ``}, {`b`, `bbbbb`, ``}, {`b`, `bbbbbb`, ``}, {`b`, `bbbbbbb`, ``}, {`b`, `bbbbbbbb`, ``}, {`d`, `ddd`, ``}, {`d`, `dddd`, ``}, {`d`, `ddddd`, ``}, {`d`, `dddddd`, ``}, {`d`, `ddddddd`, ``}, {`d`, `dddddddd`, ``}},
			},
			{
				Statement: `SELECT relname, c.* FROM c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`c`, `ccc`, ``}, {`c`, `cccc`, ``}, {`c`, `ccccc`, ``}, {`c`, `cccccc`, ``}, {`c`, `ccccccc`, ``}, {`c`, `cccccccc`, ``}, {`d`, `ddd`, ``}, {`d`, `dddd`, ``}, {`d`, `ddddd`, ``}, {`d`, `dddddd`, ``}, {`d`, `ddddddd`, ``}, {`d`, `dddddddd`, ``}},
			},
			{
				Statement: `SELECT relname, d.* FROM d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `ddd`, ``, ``, ``}, {`d`, `dddd`, ``, ``, ``}, {`d`, `ddddd`, ``, ``, ``}, {`d`, `dddddd`, ``, ``, ``}, {`d`, `ddddddd`, ``, ``, ``}, {`d`, `dddddddd`, ``, ``, ``}},
			},
			{
				Statement: `SELECT relname, a.* FROM ONLY a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`a`, `zzzz`}, {`a`, `zzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}},
			},
			{
				Statement: `SELECT relname, b.* FROM ONLY b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`b`, `bbb`, ``}, {`b`, `bbbb`, ``}, {`b`, `bbbbb`, ``}, {`b`, `bbbbbb`, ``}, {`b`, `bbbbbbb`, ``}, {`b`, `bbbbbbbb`, ``}},
			},
			{
				Statement: `SELECT relname, c.* FROM ONLY c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`c`, `ccc`, ``}, {`c`, `cccc`, ``}, {`c`, `ccccc`, ``}, {`c`, `cccccc`, ``}, {`c`, `ccccccc`, ``}, {`c`, `cccccccc`, ``}},
			},
			{
				Statement: `SELECT relname, d.* FROM ONLY d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `ddd`, ``, ``, ``}, {`d`, `dddd`, ``, ``, ``}, {`d`, `ddddd`, ``, ``, ``}, {`d`, `dddddd`, ``, ``, ``}, {`d`, `ddddddd`, ``, ``, ``}, {`d`, `dddddddd`, ``, ``, ``}},
			},
			{
				Statement: `UPDATE b SET aa='new';`,
			},
			{
				Statement: `SELECT relname, a.* FROM a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`a`, `zzzz`}, {`a`, `zzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`b`, `new`}, {`b`, `new`}, {`b`, `new`}, {`b`, `new`}, {`b`, `new`}, {`b`, `new`}, {`c`, `ccc`}, {`c`, `cccc`}, {`c`, `ccccc`}, {`c`, `cccccc`}, {`c`, `ccccccc`}, {`c`, `cccccccc`}, {`d`, `new`}, {`d`, `new`}, {`d`, `new`}, {`d`, `new`}, {`d`, `new`}, {`d`, `new`}},
			},
			{
				Statement: `SELECT relname, b.* FROM b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}},
			},
			{
				Statement: `SELECT relname, c.* FROM c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`c`, `ccc`, ``}, {`c`, `cccc`, ``}, {`c`, `ccccc`, ``}, {`c`, `cccccc`, ``}, {`c`, `ccccccc`, ``}, {`c`, `cccccccc`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}},
			},
			{
				Statement: `SELECT relname, d.* FROM d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}},
			},
			{
				Statement: `SELECT relname, a.* FROM ONLY a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`a`, `zzzz`}, {`a`, `zzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}, {`a`, `zzzzzz`}},
			},
			{
				Statement: `SELECT relname, b.* FROM ONLY b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}},
			},
			{
				Statement: `SELECT relname, c.* FROM ONLY c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`c`, `ccc`, ``}, {`c`, `cccc`, ``}, {`c`, `ccccc`, ``}, {`c`, `cccccc`, ``}, {`c`, `ccccccc`, ``}, {`c`, `cccccccc`, ``}},
			},
			{
				Statement: `SELECT relname, d.* FROM ONLY d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}},
			},
			{
				Statement: `UPDATE a SET aa='new';`,
			},
			{
				Statement: `DELETE FROM ONLY c WHERE aa='new';`,
			},
			{
				Statement: `SELECT relname, a.* FROM a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`a`, `new`}, {`a`, `new`}, {`a`, `new`}, {`a`, `new`}, {`a`, `new`}, {`a`, `new`}, {`b`, `new`}, {`b`, `new`}, {`b`, `new`}, {`b`, `new`}, {`b`, `new`}, {`b`, `new`}, {`d`, `new`}, {`d`, `new`}, {`d`, `new`}, {`d`, `new`}, {`d`, `new`}, {`d`, `new`}},
			},
			{
				Statement: `SELECT relname, b.* FROM b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}},
			},
			{
				Statement: `SELECT relname, c.* FROM c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}, {`d`, `new`, ``}},
			},
			{
				Statement: `SELECT relname, d.* FROM d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}},
			},
			{
				Statement: `SELECT relname, a.* FROM ONLY a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`a`, `new`}, {`a`, `new`}, {`a`, `new`}, {`a`, `new`}, {`a`, `new`}, {`a`, `new`}},
			},
			{
				Statement: `SELECT relname, b.* FROM ONLY b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}, {`b`, `new`, ``}},
			},
			{
				Statement: `SELECT relname, c.* FROM ONLY c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT relname, d.* FROM ONLY d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{{`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}, {`d`, `new`, ``, ``, ``}},
			},
			{
				Statement: `DELETE FROM a;`,
			},
			{
				Statement: `SELECT relname, a.* FROM a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT relname, b.* FROM b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT relname, c.* FROM c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT relname, d.* FROM d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT relname, a.* FROM ONLY a, pg_class where a.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT relname, b.* FROM ONLY b, pg_class where b.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT relname, c.* FROM ONLY c, pg_class where c.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT relname, d.* FROM ONLY d, pg_class where d.tableoid = pg_class.oid;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `CREATE TEMP TABLE z (b TEXT, PRIMARY KEY(aa, b)) inherits (a);`,
			},
			{
				Statement:   `INSERT INTO z VALUES (NULL, 'text'); -- should fail`,
				ErrorString: `null value in column "aa" of relation "z" violates not-null constraint`,
			},
			{
				Statement: `create table some_tab (a int, b int);`,
			},
			{
				Statement: `create table some_tab_child () inherits (some_tab);`,
			},
			{
				Statement: `insert into some_tab_child values(1,2);`,
			},
			{
				Statement: `explain (verbose, costs off)
update some_tab set a = a + 1 where false;`,
				Results: []sql.Row{{`Update on public.some_tab`}, {`->  Result`}, {`Output: (some_tab.a + 1), NULL::oid, NULL::tid`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `update some_tab set a = a + 1 where false;`,
			},
			{
				Statement: `explain (verbose, costs off)
update some_tab set a = a + 1 where false returning b, a;`,
				Results: []sql.Row{{`Update on public.some_tab`}, {`Output: some_tab.b, some_tab.a`}, {`->  Result`}, {`Output: (some_tab.a + 1), NULL::oid, NULL::tid`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `update some_tab set a = a + 1 where false returning b, a;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `table some_tab;`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `drop table some_tab cascade;`,
			},
			{
				Statement: `create temp table foo(f1 int, f2 int);`,
			},
			{
				Statement: `create temp table foo2(f3 int) inherits (foo);`,
			},
			{
				Statement: `create temp table bar(f1 int, f2 int);`,
			},
			{
				Statement: `create temp table bar2(f3 int) inherits (bar);`,
			},
			{
				Statement: `insert into foo values(1,1);`,
			},
			{
				Statement: `insert into foo values(3,3);`,
			},
			{
				Statement: `insert into foo2 values(2,2,2);`,
			},
			{
				Statement: `insert into foo2 values(3,3,3);`,
			},
			{
				Statement: `insert into bar values(1,1);`,
			},
			{
				Statement: `insert into bar values(2,2);`,
			},
			{
				Statement: `insert into bar values(3,3);`,
			},
			{
				Statement: `insert into bar values(4,4);`,
			},
			{
				Statement: `insert into bar2 values(1,1,1);`,
			},
			{
				Statement: `insert into bar2 values(2,2,2);`,
			},
			{
				Statement: `insert into bar2 values(3,3,3);`,
			},
			{
				Statement: `insert into bar2 values(4,4,4);`,
			},
			{
				Statement: `update bar set f2 = f2 + 100 where f1 in (select f1 from foo);`,
			},
			{
				Statement: `select tableoid::regclass::text as relname, bar.* from bar order by 1,2;`,
				Results:   []sql.Row{{`bar`, 1, 101}, {`bar`, 2, 102}, {`bar`, 3, 103}, {`bar`, 4, 4}, {`bar2`, 1, 101}, {`bar2`, 2, 102}, {`bar2`, 3, 103}, {`bar2`, 4, 4}},
			},
			{
				Statement: `update bar set f2 = f2 + 100
from
  ( select f1 from foo union all select f1+3 from foo ) ss
where bar.f1 = ss.f1;`,
			},
			{
				Statement: `select tableoid::regclass::text as relname, bar.* from bar order by 1,2;`,
				Results:   []sql.Row{{`bar`, 1, 201}, {`bar`, 2, 202}, {`bar`, 3, 203}, {`bar`, 4, 104}, {`bar2`, 1, 201}, {`bar2`, 2, 202}, {`bar2`, 3, 203}, {`bar2`, 4, 104}},
			},
			{
				Statement: `create table some_tab (a int);`,
			},
			{
				Statement: `insert into some_tab values (0);`,
			},
			{
				Statement: `create table some_tab_child () inherits (some_tab);`,
			},
			{
				Statement: `insert into some_tab_child values (1);`,
			},
			{
				Statement: `create table parted_tab (a int, b char) partition by list (a);`,
			},
			{
				Statement: `create table parted_tab_part1 partition of parted_tab for values in (1);`,
			},
			{
				Statement: `create table parted_tab_part2 partition of parted_tab for values in (2);`,
			},
			{
				Statement: `create table parted_tab_part3 partition of parted_tab for values in (3);`,
			},
			{
				Statement: `insert into parted_tab values (1, 'a'), (2, 'a'), (3, 'a');`,
			},
			{
				Statement: `update parted_tab set b = 'b'
from
  (select a from some_tab union all select a+1 from some_tab) ss (a)
where parted_tab.a = ss.a;`,
			},
			{
				Statement: `select tableoid::regclass::text as relname, parted_tab.* from parted_tab order by 1,2;`,
				Results:   []sql.Row{{`parted_tab_part1`, 1, `b`}, {`parted_tab_part2`, 2, `b`}, {`parted_tab_part3`, 3, `a`}},
			},
			{
				Statement: `truncate parted_tab;`,
			},
			{
				Statement: `insert into parted_tab values (1, 'a'), (2, 'a'), (3, 'a');`,
			},
			{
				Statement: `update parted_tab set b = 'b'
from
  (select 0 from parted_tab union all select 1 from parted_tab) ss (a)
where parted_tab.a = ss.a;`,
			},
			{
				Statement: `select tableoid::regclass::text as relname, parted_tab.* from parted_tab order by 1,2;`,
				Results:   []sql.Row{{`parted_tab_part1`, 1, `b`}, {`parted_tab_part2`, 2, `a`}, {`parted_tab_part3`, 3, `a`}},
			},
			{
				Statement: `explain update parted_tab set a = 2 where false;`,
				Results:   []sql.Row{{`Update on parted_tab  (cost=0.00..0.00 rows=0 width=0)`}, {`->  Result  (cost=0.00..0.00 rows=0 width=10)`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `drop table parted_tab;`,
			},
			{
				Statement: `create table mlparted_tab (a int, b char, c text) partition by list (a);`,
			},
			{
				Statement: `create table mlparted_tab_part1 partition of mlparted_tab for values in (1);`,
			},
			{
				Statement: `create table mlparted_tab_part2 partition of mlparted_tab for values in (2) partition by list (b);`,
			},
			{
				Statement: `create table mlparted_tab_part3 partition of mlparted_tab for values in (3);`,
			},
			{
				Statement: `create table mlparted_tab_part2a partition of mlparted_tab_part2 for values in ('a');`,
			},
			{
				Statement: `create table mlparted_tab_part2b partition of mlparted_tab_part2 for values in ('b');`,
			},
			{
				Statement: `insert into mlparted_tab values (1, 'a'), (2, 'a'), (2, 'b'), (3, 'a');`,
			},
			{
				Statement: `update mlparted_tab mlp set c = 'xxx'
from
  (select a from some_tab union all select a+1 from some_tab) ss (a)
where (mlp.a = ss.a and mlp.b = 'b') or mlp.a = 3;`,
			},
			{
				Statement: `select tableoid::regclass::text as relname, mlparted_tab.* from mlparted_tab order by 1,2;`,
				Results:   []sql.Row{{`mlparted_tab_part1`, 1, `a`, ``}, {`mlparted_tab_part2a`, 2, `a`, ``}, {`mlparted_tab_part2b`, 2, `b`, `xxx`}, {`mlparted_tab_part3`, 3, `a`, `xxx`}},
			},
			{
				Statement: `drop table mlparted_tab;`,
			},
			{
				Statement: `drop table some_tab cascade;`,
			},
			{
				Statement: `/* Test multiple inheritance of column defaults */
CREATE TABLE firstparent (tomorrow date default now()::date + 1);`,
			},
			{
				Statement: `CREATE TABLE secondparent (tomorrow date default  now() :: date  +  1);`,
			},
			{
				Statement: `CREATE TABLE jointchild () INHERITS (firstparent, secondparent);  -- ok`,
			},
			{
				Statement: `CREATE TABLE thirdparent (tomorrow date default now()::date - 1);`,
			},
			{
				Statement:   `CREATE TABLE otherchild () INHERITS (firstparent, thirdparent);  -- not ok`,
				ErrorString: `column "tomorrow" inherits conflicting default values`,
			},
			{
				Statement: `CREATE TABLE otherchild (tomorrow date default now())
  INHERITS (firstparent, thirdparent);  -- ok, child resolves ambiguous default`,
			},
			{
				Statement: `DROP TABLE firstparent, secondparent, jointchild, thirdparent, otherchild;`,
			},
			{
				Statement: `insert into d values('test','one','two','three');`,
			},
			{
				Statement: `alter table a alter column aa type integer using bit_length(aa);`,
			},
			{
				Statement: `select * from d;`,
				Results:   []sql.Row{{32, `one`, `two`, `three`}},
			},
			{
				Statement: `create temp table parent1(f1 int, f2 int);`,
			},
			{
				Statement: `create temp table parent2(f1 int, f3 bigint);`,
			},
			{
				Statement: `create temp table childtab(f4 int) inherits(parent1, parent2);`,
			},
			{
				Statement:   `alter table parent1 alter column f1 type bigint;  -- fail, conflict w/parent2`,
				ErrorString: `cannot alter inherited column "f1" of relation "childtab"`,
			},
			{
				Statement: `alter table parent1 alter column f2 type bigint;  -- ok`,
			},
			{
				Statement: `create table p1(ff1 int);`,
			},
			{
				Statement: `alter table p1 add constraint p1chk check (ff1 > 0) no inherit;`,
			},
			{
				Statement: `alter table p1 add constraint p2chk check (ff1 > 10);`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pgc.connoinherit from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname = 'p1' order by 1,2;`,
				Results:   []sql.Row{{`p1`, `p1chk`, `c`, true, 0, true}, {`p1`, `p2chk`, `c`, true, 0, false}},
			},
			{
				Statement: `create table c1 () inherits (p1);`,
			},
			{
				Statement: `\d p1
                 Table "public.p1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 ff1    | integer |           |          | 
Check constraints:
    "p1chk" CHECK (ff1 > 0) NO INHERIT
    "p2chk" CHECK (ff1 > 10)
Number of child tables: 1 (Use \d+ to list them.)
\d c1
                 Table "public.c1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 ff1    | integer |           |          | 
Check constraints:
    "p2chk" CHECK (ff1 > 10)
Inherits: p1
create table c2 (constraint p2chk check (ff1 > 10) no inherit) inherits (p1);	--fails
ERROR:  constraint "p2chk" conflicts with inherited constraint on relation "c2"
drop table p1 cascade;`,
			},
			{
				Statement: `create table base (i integer);`,
			},
			{
				Statement: `create table derived () inherits (base);`,
			},
			{
				Statement: `create table more_derived (like derived, b int) inherits (derived);`,
			},
			{
				Statement: `insert into derived (i) values (0);`,
			},
			{
				Statement: `select derived::base from derived;`,
				Results:   []sql.Row{{`(0)`}},
			},
			{
				Statement: `select NULL::derived::base;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `explain (verbose on, costs off) select row(i, b)::more_derived::derived::base from more_derived;`,
				Results:   []sql.Row{{`Seq Scan on public.more_derived`}, {`Output: (ROW(i, b)::more_derived)::base`}},
			},
			{
				Statement: `explain (verbose on, costs off) select (1, 2)::more_derived::derived::base;`,
				Results:   []sql.Row{{`Result`}, {`Output: '(1)'::base`}},
			},
			{
				Statement: `drop table more_derived;`,
			},
			{
				Statement: `drop table derived;`,
			},
			{
				Statement: `drop table base;`,
			},
			{
				Statement: `create table p1(ff1 int);`,
			},
			{
				Statement: `create table p2(f1 text);`,
			},
			{
				Statement: `create function p2text(p2) returns text as 'select $1.f1' language sql;`,
			},
			{
				Statement: `create table c1(f3 int) inherits(p1,p2);`,
			},
			{
				Statement: `insert into c1 values(123456789, 'hi', 42);`,
			},
			{
				Statement: `select p2text(c1.*) from c1;`,
				Results:   []sql.Row{{`hi`}},
			},
			{
				Statement: `drop function p2text(p2);`,
			},
			{
				Statement: `drop table c1;`,
			},
			{
				Statement: `drop table p2;`,
			},
			{
				Statement: `drop table p1;`,
			},
			{
				Statement: `CREATE TABLE ac (aa TEXT);`,
			},
			{
				Statement: `alter table ac add constraint ac_check check (aa is not null);`,
			},
			{
				Statement: `CREATE TABLE bc (bb TEXT) INHERITS (ac);`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc') order by 1,2;`,
				Results:   []sql.Row{{`ac`, `ac_check`, `c`, true, 0, `(aa IS NOT NULL)`}, {`bc`, `ac_check`, `c`, false, 1, `(aa IS NOT NULL)`}},
			},
			{
				Statement:   `insert into ac (aa) values (NULL);`,
				ErrorString: `new row for relation "ac" violates check constraint "ac_check"`,
			},
			{
				Statement:   `insert into bc (aa) values (NULL);`,
				ErrorString: `new row for relation "bc" violates check constraint "ac_check"`,
			},
			{
				Statement:   `alter table bc drop constraint ac_check;  -- fail, disallowed`,
				ErrorString: `cannot drop inherited constraint "ac_check" of relation "bc"`,
			},
			{
				Statement: `alter table ac drop constraint ac_check;`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc') order by 1,2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `alter table ac add check (aa is not null);`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc') order by 1,2;`,
				Results:   []sql.Row{{`ac`, `ac_aa_check`, `c`, true, 0, `(aa IS NOT NULL)`}, {`bc`, `ac_aa_check`, `c`, false, 1, `(aa IS NOT NULL)`}},
			},
			{
				Statement:   `insert into ac (aa) values (NULL);`,
				ErrorString: `new row for relation "ac" violates check constraint "ac_aa_check"`,
			},
			{
				Statement:   `insert into bc (aa) values (NULL);`,
				ErrorString: `new row for relation "bc" violates check constraint "ac_aa_check"`,
			},
			{
				Statement:   `alter table bc drop constraint ac_aa_check;  -- fail, disallowed`,
				ErrorString: `cannot drop inherited constraint "ac_aa_check" of relation "bc"`,
			},
			{
				Statement: `alter table ac drop constraint ac_aa_check;`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc') order by 1,2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `alter table ac add constraint ac_check check (aa is not null);`,
			},
			{
				Statement: `alter table bc no inherit ac;`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc') order by 1,2;`,
				Results:   []sql.Row{{`ac`, `ac_check`, `c`, true, 0, `(aa IS NOT NULL)`}, {`bc`, `ac_check`, `c`, true, 0, `(aa IS NOT NULL)`}},
			},
			{
				Statement: `alter table bc drop constraint ac_check;`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc') order by 1,2;`,
				Results:   []sql.Row{{`ac`, `ac_check`, `c`, true, 0, `(aa IS NOT NULL)`}},
			},
			{
				Statement: `alter table ac drop constraint ac_check;`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc') order by 1,2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `drop table bc;`,
			},
			{
				Statement: `drop table ac;`,
			},
			{
				Statement: `create table ac (a int constraint check_a check (a <> 0));`,
			},
			{
				Statement: `create table bc (a int constraint check_a check (a <> 0), b int constraint check_b check (b <> 0)) inherits (ac);`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc') order by 1,2;`,
				Results:   []sql.Row{{`ac`, `check_a`, `c`, true, 0, `(a <> 0)`}, {`bc`, `check_a`, `c`, true, 1, `(a <> 0)`}, {`bc`, `check_b`, `c`, true, 0, `(b <> 0)`}},
			},
			{
				Statement: `drop table bc;`,
			},
			{
				Statement: `drop table ac;`,
			},
			{
				Statement: `create table ac (a int constraint check_a check (a <> 0));`,
			},
			{
				Statement: `create table bc (b int constraint check_b check (b <> 0));`,
			},
			{
				Statement: `create table cc (c int constraint check_c check (c <> 0)) inherits (ac, bc);`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc', 'cc') order by 1,2;`,
				Results:   []sql.Row{{`ac`, `check_a`, `c`, true, 0, `(a <> 0)`}, {`bc`, `check_b`, `c`, true, 0, `(b <> 0)`}, {`cc`, `check_a`, `c`, false, 1, `(a <> 0)`}, {`cc`, `check_b`, `c`, false, 1, `(b <> 0)`}, {`cc`, `check_c`, `c`, true, 0, `(c <> 0)`}},
			},
			{
				Statement: `alter table cc no inherit bc;`,
			},
			{
				Statement: `select pc.relname, pgc.conname, pgc.contype, pgc.conislocal, pgc.coninhcount, pg_get_expr(pgc.conbin, pc.oid) as consrc from pg_class as pc inner join pg_constraint as pgc on (pgc.conrelid = pc.oid) where pc.relname in ('ac', 'bc', 'cc') order by 1,2;`,
				Results:   []sql.Row{{`ac`, `check_a`, `c`, true, 0, `(a <> 0)`}, {`bc`, `check_b`, `c`, true, 0, `(b <> 0)`}, {`cc`, `check_a`, `c`, false, 1, `(a <> 0)`}, {`cc`, `check_b`, `c`, true, 0, `(b <> 0)`}, {`cc`, `check_c`, `c`, true, 0, `(c <> 0)`}},
			},
			{
				Statement: `drop table cc;`,
			},
			{
				Statement: `drop table bc;`,
			},
			{
				Statement: `drop table ac;`,
			},
			{
				Statement: `create table p1(f1 int);`,
			},
			{
				Statement: `create table p2(f2 int);`,
			},
			{
				Statement: `create table c1(f3 int) inherits(p1,p2);`,
			},
			{
				Statement: `insert into c1 values(1,-1,2);`,
			},
			{
				Statement:   `alter table p2 add constraint cc check (f2>0);  -- fail`,
				ErrorString: `check constraint "cc" of relation "c1" is violated by some row`,
			},
			{
				Statement:   `alter table p2 add check (f2>0);  -- check it without a name, too`,
				ErrorString: `check constraint "p2_f2_check" of relation "c1" is violated by some row`,
			},
			{
				Statement: `delete from c1;`,
			},
			{
				Statement: `insert into c1 values(1,1,2);`,
			},
			{
				Statement: `alter table p2 add check (f2>0);`,
			},
			{
				Statement:   `insert into c1 values(1,-1,2);  -- fail`,
				ErrorString: `new row for relation "c1" violates check constraint "p2_f2_check"`,
			},
			{
				Statement: `create table c2(f3 int) inherits(p1,p2);`,
			},
			{
				Statement: `\d c2
                 Table "public.c2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 f1     | integer |           |          | 
 f2     | integer |           |          | 
 f3     | integer |           |          | 
Check constraints:
    "p2_f2_check" CHECK (f2 > 0)
Inherits: p1,
          p2
create table c3 (f4 int) inherits(c1,c2);`,
			},
			{
				Statement: `\d c3
                 Table "public.c3"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 f1     | integer |           |          | 
 f2     | integer |           |          | 
 f3     | integer |           |          | 
 f4     | integer |           |          | 
Check constraints:
    "p2_f2_check" CHECK (f2 > 0)
Inherits: c1,
          c2
drop table p1 cascade;`,
			},
			{
				Statement: `drop table p2 cascade;`,
			},
			{
				Statement: `create table pp1 (f1 int);`,
			},
			{
				Statement: `create table cc1 (f2 text, f3 int) inherits (pp1);`,
			},
			{
				Statement: `alter table pp1 add column a1 int check (a1 > 0);`,
			},
			{
				Statement: `\d cc1
                Table "public.cc1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 f1     | integer |           |          | 
 f2     | text    |           |          | 
 f3     | integer |           |          | 
 a1     | integer |           |          | 
Check constraints:
    "pp1_a1_check" CHECK (a1 > 0)
Inherits: pp1
create table cc2(f4 float) inherits(pp1,cc1);`,
			},
			{
				Statement: `\d cc2
                     Table "public.cc2"
 Column |       Type       | Collation | Nullable | Default 
--------+------------------+-----------+----------+---------
 f1     | integer          |           |          | 
 a1     | integer          |           |          | 
 f2     | text             |           |          | 
 f3     | integer          |           |          | 
 f4     | double precision |           |          | 
Check constraints:
    "pp1_a1_check" CHECK (a1 > 0)
Inherits: pp1,
          cc1
alter table pp1 add column a2 int check (a2 > 0);`,
			},
			{
				Statement: `\d cc2
                     Table "public.cc2"
 Column |       Type       | Collation | Nullable | Default 
--------+------------------+-----------+----------+---------
 f1     | integer          |           |          | 
 a1     | integer          |           |          | 
 f2     | text             |           |          | 
 f3     | integer          |           |          | 
 f4     | double precision |           |          | 
 a2     | integer          |           |          | 
Check constraints:
    "pp1_a1_check" CHECK (a1 > 0)
    "pp1_a2_check" CHECK (a2 > 0)
Inherits: pp1,
          cc1
drop table pp1 cascade;`,
			},
			{
				Statement: `CREATE TABLE inht1 (a int, b int);`,
			},
			{
				Statement: `CREATE TABLE inhs1 (b int, c int);`,
			},
			{
				Statement: `CREATE TABLE inhts (d int) INHERITS (inht1, inhs1);`,
			},
			{
				Statement: `ALTER TABLE inht1 RENAME a TO aa;`,
			},
			{
				Statement:   `ALTER TABLE inht1 RENAME b TO bb;                -- to be failed`,
				ErrorString: `cannot rename inherited column "b"`,
			},
			{
				Statement:   `ALTER TABLE inhts RENAME aa TO aaa;      -- to be failed`,
				ErrorString: `cannot rename inherited column "aa"`,
			},
			{
				Statement: `ALTER TABLE inhts RENAME d TO dd;`,
			},
			{
				Statement: `\d+ inhts
                                   Table "public.inhts"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 aa     | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
 c      | integer |           |          |         | plain   |              | 
 dd     | integer |           |          |         | plain   |              | 
Inherits: inht1,
          inhs1
DROP TABLE inhts;`,
			},
			{
				Statement: `CREATE TABLE inht2 (x int) INHERITS (inht1);`,
			},
			{
				Statement: `CREATE TABLE inht3 (y int) INHERITS (inht1);`,
			},
			{
				Statement: `CREATE TABLE inht4 (z int) INHERITS (inht2, inht3);`,
			},
			{
				Statement: `ALTER TABLE inht1 RENAME aa TO aaa;`,
			},
			{
				Statement: `\d+ inht4
                                   Table "public.inht4"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 aaa    | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
 x      | integer |           |          |         | plain   |              | 
 y      | integer |           |          |         | plain   |              | 
 z      | integer |           |          |         | plain   |              | 
Inherits: inht2,
          inht3
CREATE TABLE inhts (d int) INHERITS (inht2, inhs1);`,
			},
			{
				Statement: `ALTER TABLE inht1 RENAME aaa TO aaaa;`,
			},
			{
				Statement:   `ALTER TABLE inht1 RENAME b TO bb;                -- to be failed`,
				ErrorString: `cannot rename inherited column "b"`,
			},
			{
				Statement: `\d+ inhts
                                   Table "public.inhts"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 aaaa   | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
 x      | integer |           |          |         | plain   |              | 
 c      | integer |           |          |         | plain   |              | 
 d      | integer |           |          |         | plain   |              | 
Inherits: inht2,
          inhs1
WITH RECURSIVE r AS (
  SELECT 'inht1'::regclass AS inhrelid
UNION ALL
  SELECT c.inhrelid FROM pg_inherits c, r WHERE r.inhrelid = c.inhparent
)
SELECT a.attrelid::regclass, a.attname, a.attinhcount, e.expected
  FROM (SELECT inhrelid, count(*) AS expected FROM pg_inherits
        WHERE inhparent IN (SELECT inhrelid FROM r) GROUP BY inhrelid) e
  JOIN pg_attribute a ON e.inhrelid = a.attrelid WHERE NOT attislocal
  ORDER BY a.attrelid::regclass::name, a.attnum;`,
				Results: []sql.Row{{`inht2`, `aaaa`, 1, 1}, {`inht2`, `b`, 1, 1}, {`inht3`, `aaaa`, 1, 1}, {`inht3`, `b`, 1, 1}, {`inht4`, `aaaa`, 2, 2}, {`inht4`, `b`, 2, 2}, {`inht4`, `x`, 1, 2}, {`inht4`, `y`, 1, 2}, {`inhts`, `aaaa`, 1, 1}, {`inhts`, `b`, 2, 1}, {`inhts`, `x`, 1, 1}, {`inhts`, `c`, 1, 1}},
			},
			{
				Statement: `DROP TABLE inht1, inhs1 CASCADE;`,
			},
			{
				Statement: `CREATE TABLE test_constraints (id int, val1 varchar, val2 int, UNIQUE(val1, val2));`,
			},
			{
				Statement: `CREATE TABLE test_constraints_inh () INHERITS (test_constraints);`,
			},
			{
				Statement: `\d+ test_constraints
                                   Table "public.test_constraints"
 Column |       Type        | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+-------------------+-----------+----------+---------+----------+--------------+-------------
 id     | integer           |           |          |         | plain    |              | 
 val1   | character varying |           |          |         | extended |              | 
 val2   | integer           |           |          |         | plain    |              | 
Indexes:
    "test_constraints_val1_val2_key" UNIQUE CONSTRAINT, btree (val1, val2)
Child tables: test_constraints_inh
ALTER TABLE ONLY test_constraints DROP CONSTRAINT test_constraints_val1_val2_key;`,
			},
			{
				Statement: `\d+ test_constraints
                                   Table "public.test_constraints"
 Column |       Type        | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+-------------------+-----------+----------+---------+----------+--------------+-------------
 id     | integer           |           |          |         | plain    |              | 
 val1   | character varying |           |          |         | extended |              | 
 val2   | integer           |           |          |         | plain    |              | 
Child tables: test_constraints_inh
\d+ test_constraints_inh
                                 Table "public.test_constraints_inh"
 Column |       Type        | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+-------------------+-----------+----------+---------+----------+--------------+-------------
 id     | integer           |           |          |         | plain    |              | 
 val1   | character varying |           |          |         | extended |              | 
 val2   | integer           |           |          |         | plain    |              | 
Inherits: test_constraints
DROP TABLE test_constraints_inh;`,
			},
			{
				Statement: `DROP TABLE test_constraints;`,
			},
			{
				Statement: `CREATE TABLE test_ex_constraints (
    c circle,
    EXCLUDE USING gist (c WITH &&)
);`,
			},
			{
				Statement: `CREATE TABLE test_ex_constraints_inh () INHERITS (test_ex_constraints);`,
			},
			{
				Statement: `\d+ test_ex_constraints
                           Table "public.test_ex_constraints"
 Column |  Type  | Collation | Nullable | Default | Storage | Stats target | Description 
--------+--------+-----------+----------+---------+---------+--------------+-------------
 c      | circle |           |          |         | plain   |              | 
Indexes:
    "test_ex_constraints_c_excl" EXCLUDE USING gist (c WITH &&)
Child tables: test_ex_constraints_inh
ALTER TABLE test_ex_constraints DROP CONSTRAINT test_ex_constraints_c_excl;`,
			},
			{
				Statement: `\d+ test_ex_constraints
                           Table "public.test_ex_constraints"
 Column |  Type  | Collation | Nullable | Default | Storage | Stats target | Description 
--------+--------+-----------+----------+---------+---------+--------------+-------------
 c      | circle |           |          |         | plain   |              | 
Child tables: test_ex_constraints_inh
\d+ test_ex_constraints_inh
                         Table "public.test_ex_constraints_inh"
 Column |  Type  | Collation | Nullable | Default | Storage | Stats target | Description 
--------+--------+-----------+----------+---------+---------+--------------+-------------
 c      | circle |           |          |         | plain   |              | 
Inherits: test_ex_constraints
DROP TABLE test_ex_constraints_inh;`,
			},
			{
				Statement: `DROP TABLE test_ex_constraints;`,
			},
			{
				Statement: `CREATE TABLE test_primary_constraints(id int PRIMARY KEY);`,
			},
			{
				Statement: `CREATE TABLE test_foreign_constraints(id1 int REFERENCES test_primary_constraints(id));`,
			},
			{
				Statement: `CREATE TABLE test_foreign_constraints_inh () INHERITS (test_foreign_constraints);`,
			},
			{
				Statement: `\d+ test_primary_constraints
                         Table "public.test_primary_constraints"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 id     | integer |           | not null |         | plain   |              | 
Indexes:
    "test_primary_constraints_pkey" PRIMARY KEY, btree (id)
Referenced by:
    TABLE "test_foreign_constraints" CONSTRAINT "test_foreign_constraints_id1_fkey" FOREIGN KEY (id1) REFERENCES test_primary_constraints(id)
\d+ test_foreign_constraints
                         Table "public.test_foreign_constraints"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 id1    | integer |           |          |         | plain   |              | 
Foreign-key constraints:
    "test_foreign_constraints_id1_fkey" FOREIGN KEY (id1) REFERENCES test_primary_constraints(id)
Child tables: test_foreign_constraints_inh
ALTER TABLE test_foreign_constraints DROP CONSTRAINT test_foreign_constraints_id1_fkey;`,
			},
			{
				Statement: `\d+ test_foreign_constraints
                         Table "public.test_foreign_constraints"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 id1    | integer |           |          |         | plain   |              | 
Child tables: test_foreign_constraints_inh
\d+ test_foreign_constraints_inh
                       Table "public.test_foreign_constraints_inh"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 id1    | integer |           |          |         | plain   |              | 
Inherits: test_foreign_constraints
DROP TABLE test_foreign_constraints_inh;`,
			},
			{
				Statement: `DROP TABLE test_foreign_constraints;`,
			},
			{
				Statement: `DROP TABLE test_primary_constraints;`,
			},
			{
				Statement: `create table inh_fk_1 (a int primary key);`,
			},
			{
				Statement: `insert into inh_fk_1 values (1), (2), (3);`,
			},
			{
				Statement: `create table inh_fk_2 (x int primary key, y int references inh_fk_1 on delete cascade);`,
			},
			{
				Statement: `insert into inh_fk_2 values (11, 1), (22, 2), (33, 3);`,
			},
			{
				Statement: `create table inh_fk_2_child () inherits (inh_fk_2);`,
			},
			{
				Statement: `insert into inh_fk_2_child values (111, 1), (222, 2);`,
			},
			{
				Statement: `delete from inh_fk_1 where a = 1;`,
			},
			{
				Statement: `select * from inh_fk_1 order by 1;`,
				Results:   []sql.Row{{2}, {3}},
			},
			{
				Statement: `select * from inh_fk_2 order by 1, 2;`,
				Results:   []sql.Row{{22, 2}, {33, 3}, {111, 1}, {222, 2}},
			},
			{
				Statement: `drop table inh_fk_1, inh_fk_2, inh_fk_2_child;`,
			},
			{
				Statement: `create table p1(f1 int);`,
			},
			{
				Statement: `create table p1_c1() inherits(p1);`,
			},
			{
				Statement: `alter table p1 add constraint inh_check_constraint1 check (f1 > 0);`,
			},
			{
				Statement: `alter table p1_c1 add constraint inh_check_constraint1 check (f1 > 0);`,
			},
			{
				Statement: `alter table p1_c1 add constraint inh_check_constraint2 check (f1 < 10);`,
			},
			{
				Statement: `alter table p1 add constraint inh_check_constraint2 check (f1 < 10);`,
			},
			{
				Statement: `select conrelid::regclass::text as relname, conname, conislocal, coninhcount
from pg_constraint where conname like 'inh\_check\_constraint%'
order by 1, 2;`,
				Results: []sql.Row{{`p1`, `inh_check_constraint1`, true, 0}, {`p1`, `inh_check_constraint2`, true, 0}, {`p1_c1`, `inh_check_constraint1`, true, 1}, {`p1_c1`, `inh_check_constraint2`, true, 1}},
			},
			{
				Statement: `drop table p1 cascade;`,
			},
			{
				Statement: `create table invalid_check_con(f1 int);`,
			},
			{
				Statement: `create table invalid_check_con_child() inherits(invalid_check_con);`,
			},
			{
				Statement: `alter table invalid_check_con_child add constraint inh_check_constraint check(f1 > 0) not valid;`,
			},
			{
				Statement:   `alter table invalid_check_con add constraint inh_check_constraint check(f1 > 0); -- fail`,
				ErrorString: `constraint "inh_check_constraint" conflicts with NOT VALID constraint on relation "invalid_check_con_child"`,
			},
			{
				Statement: `alter table invalid_check_con_child drop constraint inh_check_constraint;`,
			},
			{
				Statement: `insert into invalid_check_con values(0);`,
			},
			{
				Statement: `alter table invalid_check_con_child add constraint inh_check_constraint check(f1 > 0);`,
			},
			{
				Statement: `alter table invalid_check_con add constraint inh_check_constraint check(f1 > 0) not valid;`,
			},
			{
				Statement:   `insert into invalid_check_con values(0); -- fail`,
				ErrorString: `new row for relation "invalid_check_con" violates check constraint "inh_check_constraint"`,
			},
			{
				Statement:   `insert into invalid_check_con_child values(0); -- fail`,
				ErrorString: `new row for relation "invalid_check_con_child" violates check constraint "inh_check_constraint"`,
			},
			{
				Statement: `select conrelid::regclass::text as relname, conname,
       convalidated, conislocal, coninhcount, connoinherit
from pg_constraint where conname like 'inh\_check\_constraint%'
order by 1, 2;`,
				Results: []sql.Row{{`invalid_check_con`, `inh_check_constraint`, false, true, 0, false}, {`invalid_check_con_child`, `inh_check_constraint`, true, true, 1, false}},
			},
			{
				Statement: `create temp table patest0 (id, x) as
  select x, x from generate_series(0,1000) x;`,
			},
			{
				Statement: `create temp table patest1() inherits (patest0);`,
			},
			{
				Statement: `insert into patest1
  select x, x from generate_series(0,1000) x;`,
			},
			{
				Statement: `create temp table patest2() inherits (patest0);`,
			},
			{
				Statement: `insert into patest2
  select x, x from generate_series(0,1000) x;`,
			},
			{
				Statement: `create index patest0i on patest0(id);`,
			},
			{
				Statement: `create index patest1i on patest1(id);`,
			},
			{
				Statement: `create index patest2i on patest2(id);`,
			},
			{
				Statement: `analyze patest0;`,
			},
			{
				Statement: `analyze patest1;`,
			},
			{
				Statement: `analyze patest2;`,
			},
			{
				Statement: `explain (costs off)
select * from patest0 join (select f1 from int4_tbl limit 1) ss on id = f1;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Limit`}, {`->  Seq Scan on int4_tbl`}, {`->  Append`}, {`->  Index Scan using patest0i on patest0 patest0_1`}, {`Index Cond: (id = int4_tbl.f1)`}, {`->  Index Scan using patest1i on patest1 patest0_2`}, {`Index Cond: (id = int4_tbl.f1)`}, {`->  Index Scan using patest2i on patest2 patest0_3`}, {`Index Cond: (id = int4_tbl.f1)`}},
			},
			{
				Statement: `select * from patest0 join (select f1 from int4_tbl limit 1) ss on id = f1;`,
				Results:   []sql.Row{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			},
			{
				Statement: `drop index patest2i;`,
			},
			{
				Statement: `explain (costs off)
select * from patest0 join (select f1 from int4_tbl limit 1) ss on id = f1;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Limit`}, {`->  Seq Scan on int4_tbl`}, {`->  Append`}, {`->  Index Scan using patest0i on patest0 patest0_1`}, {`Index Cond: (id = int4_tbl.f1)`}, {`->  Index Scan using patest1i on patest1 patest0_2`}, {`Index Cond: (id = int4_tbl.f1)`}, {`->  Seq Scan on patest2 patest0_3`}, {`Filter: (int4_tbl.f1 = id)`}},
			},
			{
				Statement: `select * from patest0 join (select f1 from int4_tbl limit 1) ss on id = f1;`,
				Results:   []sql.Row{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			},
			{
				Statement: `drop table patest0 cascade;`,
			},
			{
				Statement: `create table matest0 (id serial primary key, name text);`,
			},
			{
				Statement: `create table matest1 (id integer primary key) inherits (matest0);`,
			},
			{
				Statement: `create table matest2 (id integer primary key) inherits (matest0);`,
			},
			{
				Statement: `create table matest3 (id integer primary key) inherits (matest0);`,
			},
			{
				Statement: `create index matest0i on matest0 ((1-id));`,
			},
			{
				Statement: `create index matest1i on matest1 ((1-id));`,
			},
			{
				Statement: `create index matest3i on matest3 ((1-id));`,
			},
			{
				Statement: `insert into matest1 (name) values ('Test 1');`,
			},
			{
				Statement: `insert into matest1 (name) values ('Test 2');`,
			},
			{
				Statement: `insert into matest2 (name) values ('Test 3');`,
			},
			{
				Statement: `insert into matest2 (name) values ('Test 4');`,
			},
			{
				Statement: `insert into matest3 (name) values ('Test 5');`,
			},
			{
				Statement: `insert into matest3 (name) values ('Test 6');`,
			},
			{
				Statement: `set enable_indexscan = off;  -- force use of seqscan/sort, so no merge`,
			},
			{
				Statement: `explain (verbose, costs off) select * from matest0 order by 1-id;`,
				Results:   []sql.Row{{`Sort`}, {`Output: matest0.id, matest0.name, ((1 - matest0.id))`}, {`Sort Key: ((1 - matest0.id))`}, {`->  Result`}, {`Output: matest0.id, matest0.name, (1 - matest0.id)`}, {`->  Append`}, {`->  Seq Scan on public.matest0 matest0_1`}, {`Output: matest0_1.id, matest0_1.name`}, {`->  Seq Scan on public.matest1 matest0_2`}, {`Output: matest0_2.id, matest0_2.name`}, {`->  Seq Scan on public.matest2 matest0_3`}, {`Output: matest0_3.id, matest0_3.name`}, {`->  Seq Scan on public.matest3 matest0_4`}, {`Output: matest0_4.id, matest0_4.name`}},
			},
			{
				Statement: `select * from matest0 order by 1-id;`,
				Results:   []sql.Row{{6, `Test 6`}, {5, `Test 5`}, {4, `Test 4`}, {3, `Test 3`}, {2, `Test 2`}, {1, `Test 1`}},
			},
			{
				Statement: `explain (verbose, costs off) select min(1-id) from matest0;`,
				Results:   []sql.Row{{`Aggregate`}, {`Output: min((1 - matest0.id))`}, {`->  Append`}, {`->  Seq Scan on public.matest0 matest0_1`}, {`Output: matest0_1.id`}, {`->  Seq Scan on public.matest1 matest0_2`}, {`Output: matest0_2.id`}, {`->  Seq Scan on public.matest2 matest0_3`}, {`Output: matest0_3.id`}, {`->  Seq Scan on public.matest3 matest0_4`}, {`Output: matest0_4.id`}},
			},
			{
				Statement: `select min(1-id) from matest0;`,
				Results:   []sql.Row{{-5}},
			},
			{
				Statement: `reset enable_indexscan;`,
			},
			{
				Statement: `set enable_seqscan = off;  -- plan with fewest seqscans should be merge`,
			},
			{
				Statement: `set enable_parallel_append = off; -- Don't let parallel-append interfere`,
			},
			{
				Statement: `explain (verbose, costs off) select * from matest0 order by 1-id;`,
				Results:   []sql.Row{{`Merge Append`}, {`Sort Key: ((1 - matest0.id))`}, {`->  Index Scan using matest0i on public.matest0 matest0_1`}, {`Output: matest0_1.id, matest0_1.name, (1 - matest0_1.id)`}, {`->  Index Scan using matest1i on public.matest1 matest0_2`}, {`Output: matest0_2.id, matest0_2.name, (1 - matest0_2.id)`}, {`->  Sort`}, {`Output: matest0_3.id, matest0_3.name, ((1 - matest0_3.id))`}, {`Sort Key: ((1 - matest0_3.id))`}, {`->  Seq Scan on public.matest2 matest0_3`}, {`Output: matest0_3.id, matest0_3.name, (1 - matest0_3.id)`}, {`->  Index Scan using matest3i on public.matest3 matest0_4`}, {`Output: matest0_4.id, matest0_4.name, (1 - matest0_4.id)`}},
			},
			{
				Statement: `select * from matest0 order by 1-id;`,
				Results:   []sql.Row{{6, `Test 6`}, {5, `Test 5`}, {4, `Test 4`}, {3, `Test 3`}, {2, `Test 2`}, {1, `Test 1`}},
			},
			{
				Statement: `explain (verbose, costs off) select min(1-id) from matest0;`,
				Results:   []sql.Row{{`Result`}, {`Output: $0`}, {`InitPlan 1 (returns $0)`}, {`->  Limit`}, {`Output: ((1 - matest0.id))`}, {`->  Result`}, {`Output: ((1 - matest0.id))`}, {`->  Merge Append`}, {`Sort Key: ((1 - matest0.id))`}, {`->  Index Scan using matest0i on public.matest0 matest0_1`}, {`Output: matest0_1.id, (1 - matest0_1.id)`}, {`Index Cond: ((1 - matest0_1.id) IS NOT NULL)`}, {`->  Index Scan using matest1i on public.matest1 matest0_2`}, {`Output: matest0_2.id, (1 - matest0_2.id)`}, {`Index Cond: ((1 - matest0_2.id) IS NOT NULL)`}, {`->  Sort`}, {`Output: matest0_3.id, ((1 - matest0_3.id))`}, {`Sort Key: ((1 - matest0_3.id))`}, {`->  Bitmap Heap Scan on public.matest2 matest0_3`}, {`Output: matest0_3.id, (1 - matest0_3.id)`}, {`Filter: ((1 - matest0_3.id) IS NOT NULL)`}, {`->  Bitmap Index Scan on matest2_pkey`}, {`->  Index Scan using matest3i on public.matest3 matest0_4`}, {`Output: matest0_4.id, (1 - matest0_4.id)`}, {`Index Cond: ((1 - matest0_4.id) IS NOT NULL)`}},
			},
			{
				Statement: `select min(1-id) from matest0;`,
				Results:   []sql.Row{{-5}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_parallel_append;`,
			},
			{
				Statement: `drop table matest0 cascade;`,
			},
			{
				Statement: `create table matest0 (a int, b int, c int, d int);`,
			},
			{
				Statement: `create table matest1 () inherits(matest0);`,
			},
			{
				Statement: `create index matest0i on matest0 (b, c);`,
			},
			{
				Statement: `create index matest1i on matest1 (b, c);`,
			},
			{
				Statement: `set enable_nestloop = off;  -- we want a plan with two MergeAppends`,
			},
			{
				Statement: `explain (costs off)
select t1.* from matest0 t1, matest0 t2
where t1.b = t2.b and t2.c = t2.d
order by t1.b limit 10;`,
				Results: []sql.Row{{`Limit`}, {`->  Merge Join`}, {`Merge Cond: (t1.b = t2.b)`}, {`->  Merge Append`}, {`Sort Key: t1.b`}, {`->  Index Scan using matest0i on matest0 t1_1`}, {`->  Index Scan using matest1i on matest1 t1_2`}, {`->  Materialize`}, {`->  Merge Append`}, {`Sort Key: t2.b`}, {`->  Index Scan using matest0i on matest0 t2_1`}, {`Filter: (c = d)`}, {`->  Index Scan using matest1i on matest1 t2_2`}, {`Filter: (c = d)`}},
			},
			{
				Statement: `reset enable_nestloop;`,
			},
			{
				Statement: `drop table matest0 cascade;`,
			},
			{
				Statement: `set enable_seqscan = off;`,
			},
			{
				Statement: `set enable_indexscan = on;`,
			},
			{
				Statement: `set enable_bitmapscan = off;`,
			},
			{
				Statement: `explain (costs off)
SELECT thousand, tenthous FROM tenk1
UNION ALL
SELECT thousand, thousand FROM tenk1
ORDER BY thousand, tenthous;`,
				Results: []sql.Row{{`Merge Append`}, {`Sort Key: tenk1.thousand, tenk1.tenthous`}, {`->  Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`->  Sort`}, {`Sort Key: tenk1_1.thousand, tenk1_1.thousand`}, {`->  Index Only Scan using tenk1_thous_tenthous on tenk1 tenk1_1`}},
			},
			{
				Statement: `explain (costs off)
SELECT thousand, tenthous, thousand+tenthous AS x FROM tenk1
UNION ALL
SELECT 42, 42, hundred FROM tenk1
ORDER BY thousand, tenthous;`,
				Results: []sql.Row{{`Merge Append`}, {`Sort Key: tenk1.thousand, tenk1.tenthous`}, {`->  Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`->  Sort`}, {`Sort Key: 42, 42`}, {`->  Index Only Scan using tenk1_hundred on tenk1 tenk1_1`}},
			},
			{
				Statement: `explain (costs off)
SELECT thousand, tenthous FROM tenk1
UNION ALL
SELECT thousand, random()::integer FROM tenk1
ORDER BY thousand, tenthous;`,
				Results: []sql.Row{{`Merge Append`}, {`Sort Key: tenk1.thousand, tenk1.tenthous`}, {`->  Index Only Scan using tenk1_thous_tenthous on tenk1`}, {`->  Sort`}, {`Sort Key: tenk1_1.thousand, ((random())::integer)`}, {`->  Index Only Scan using tenk1_thous_tenthous on tenk1 tenk1_1`}},
			},
			{
				Statement: `explain (costs off)
SELECT min(x) FROM
  (SELECT unique1 AS x FROM tenk1 a
   UNION ALL
   SELECT unique2 AS x FROM tenk1 b) s;`,
				Results: []sql.Row{{`Result`}, {`InitPlan 1 (returns $0)`}, {`->  Limit`}, {`->  Merge Append`}, {`Sort Key: a.unique1`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 a`}, {`Index Cond: (unique1 IS NOT NULL)`}, {`->  Index Only Scan using tenk1_unique2 on tenk1 b`}, {`Index Cond: (unique2 IS NOT NULL)`}},
			},
			{
				Statement: `explain (costs off)
SELECT min(y) FROM
  (SELECT unique1 AS x, unique1 AS y FROM tenk1 a
   UNION ALL
   SELECT unique2 AS x, unique2 AS y FROM tenk1 b) s;`,
				Results: []sql.Row{{`Result`}, {`InitPlan 1 (returns $0)`}, {`->  Limit`}, {`->  Merge Append`}, {`Sort Key: a.unique1`}, {`->  Index Only Scan using tenk1_unique1 on tenk1 a`}, {`Index Cond: (unique1 IS NOT NULL)`}, {`->  Index Only Scan using tenk1_unique2 on tenk1 b`}, {`Index Cond: (unique2 IS NOT NULL)`}},
			},
			{
				Statement: `explain (costs off)
SELECT x, y FROM
  (SELECT thousand AS x, tenthous AS y FROM tenk1 a
   UNION ALL
   SELECT unique2 AS x, unique2 AS y FROM tenk1 b) s
ORDER BY x, y;`,
				Results: []sql.Row{{`Merge Append`}, {`Sort Key: a.thousand, a.tenthous`}, {`->  Index Only Scan using tenk1_thous_tenthous on tenk1 a`}, {`->  Sort`}, {`Sort Key: b.unique2, b.unique2`}, {`->  Index Only Scan using tenk1_unique2 on tenk1 b`}},
			},
			{
				Statement: `explain (costs off)
SELECT
    ARRAY(SELECT f.i FROM (
        (SELECT d + g.i FROM generate_series(4, 30, 3) d ORDER BY 1)
        UNION ALL
        (SELECT d + g.i FROM generate_series(0, 30, 5) d ORDER BY 1)
    ) f(i)
    ORDER BY f.i LIMIT 10)
FROM generate_series(1, 3) g(i);`,
				Results: []sql.Row{{`Function Scan on generate_series g`}, {`SubPlan 1`}, {`->  Limit`}, {`->  Merge Append`}, {`Sort Key: ((d.d + g.i))`}, {`->  Sort`}, {`Sort Key: ((d.d + g.i))`}, {`->  Function Scan on generate_series d`}, {`->  Sort`}, {`Sort Key: ((d_1.d + g.i))`}, {`->  Function Scan on generate_series d_1`}},
			},
			{
				Statement: `SELECT
    ARRAY(SELECT f.i FROM (
        (SELECT d + g.i FROM generate_series(4, 30, 3) d ORDER BY 1)
        UNION ALL
        (SELECT d + g.i FROM generate_series(0, 30, 5) d ORDER BY 1)
    ) f(i)
    ORDER BY f.i LIMIT 10)
FROM generate_series(1, 3) g(i);`,
				Results: []sql.Row{{`{1,5,6,8,11,11,14,16,17,20}`}, {`{2,6,7,9,12,12,15,17,18,21}`}, {`{3,7,8,10,13,13,16,18,19,22}`}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_indexscan;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `create table inhpar(f1 int, f2 name);`,
			},
			{
				Statement: `create table inhcld(f2 name, f1 int);`,
			},
			{
				Statement: `alter table inhcld inherit inhpar;`,
			},
			{
				Statement: `insert into inhpar select x, x::text from generate_series(1,5) x;`,
			},
			{
				Statement: `insert into inhcld select x::text, x from generate_series(6,10) x;`,
			},
			{
				Statement: `explain (verbose, costs off)
update inhpar i set (f1, f2) = (select i.f1, i.f2 || '-' from int4_tbl limit 1);`,
				Results: []sql.Row{{`Update on public.inhpar i`}, {`Update on public.inhpar i_1`}, {`Update on public.inhcld i_2`}, {`->  Result`}, {`Output: $2, $3, (SubPlan 1 (returns $2,$3)), i.tableoid, i.ctid`}, {`->  Append`}, {`->  Seq Scan on public.inhpar i_1`}, {`Output: i_1.f1, i_1.f2, i_1.tableoid, i_1.ctid`}, {`->  Seq Scan on public.inhcld i_2`}, {`Output: i_2.f1, i_2.f2, i_2.tableoid, i_2.ctid`}, {`SubPlan 1 (returns $2,$3)`}, {`->  Limit`}, {`Output: (i.f1), (((i.f2)::text || '-'::text))`}, {`->  Seq Scan on public.int4_tbl`}, {`Output: i.f1, ((i.f2)::text || '-'::text)`}},
			},
			{
				Statement: `update inhpar i set (f1, f2) = (select i.f1, i.f2 || '-' from int4_tbl limit 1);`,
			},
			{
				Statement: `select * from inhpar;`,
				Results:   []sql.Row{{1, `1-`}, {2, `2-`}, {3, `3-`}, {4, `4-`}, {5, `5-`}, {6, `6-`}, {7, `7-`}, {8, `8-`}, {9, `9-`}, {10, `10-`}},
			},
			{
				Statement: `drop table inhpar cascade;`,
			},
			{
				Statement: `create table inhpar(f1 int primary key, f2 name) partition by range (f1);`,
			},
			{
				Statement: `create table inhcld1(f2 name, f1 int primary key);`,
			},
			{
				Statement: `create table inhcld2(f1 int primary key, f2 name);`,
			},
			{
				Statement: `alter table inhpar attach partition inhcld1 for values from (1) to (5);`,
			},
			{
				Statement: `alter table inhpar attach partition inhcld2 for values from (5) to (100);`,
			},
			{
				Statement: `insert into inhpar select x, x::text from generate_series(1,10) x;`,
			},
			{
				Statement: `explain (verbose, costs off)
update inhpar i set (f1, f2) = (select i.f1, i.f2 || '-' from int4_tbl limit 1);`,
				Results: []sql.Row{{`Update on public.inhpar i`}, {`Update on public.inhcld1 i_1`}, {`Update on public.inhcld2 i_2`}, {`->  Append`}, {`->  Seq Scan on public.inhcld1 i_1`}, {`Output: $2, $3, (SubPlan 1 (returns $2,$3)), i_1.tableoid, i_1.ctid`}, {`SubPlan 1 (returns $2,$3)`}, {`->  Limit`}, {`Output: (i_1.f1), (((i_1.f2)::text || '-'::text))`}, {`->  Seq Scan on public.int4_tbl`}, {`Output: i_1.f1, ((i_1.f2)::text || '-'::text)`}, {`->  Seq Scan on public.inhcld2 i_2`}, {`Output: $2, $3, (SubPlan 1 (returns $2,$3)), i_2.tableoid, i_2.ctid`}},
			},
			{
				Statement: `update inhpar i set (f1, f2) = (select i.f1, i.f2 || '-' from int4_tbl limit 1);`,
			},
			{
				Statement: `select * from inhpar;`,
				Results:   []sql.Row{{1, `1-`}, {2, `2-`}, {3, `3-`}, {4, `4-`}, {5, `5-`}, {6, `6-`}, {7, `7-`}, {8, `8-`}, {9, `9-`}, {10, `10-`}},
			},
			{
				Statement: `insert into inhpar as i values (3), (7) on conflict (f1)
  do update set (f1, f2) = (select i.f1, i.f2 || '+');`,
			},
			{
				Statement: `select * from inhpar order by f1;  -- tuple order might be unstable here`,
				Results:   []sql.Row{{1, `1-`}, {2, `2-`}, {3, `3-+`}, {4, `4-`}, {5, `5-`}, {6, `6-`}, {7, `7-+`}, {8, `8-`}, {9, `9-`}, {10, `10-`}},
			},
			{
				Statement: `drop table inhpar cascade;`,
			},
			{
				Statement: `create table cnullparent (f1 int);`,
			},
			{
				Statement: `create table cnullchild (check (f1 = 1 or f1 = null)) inherits(cnullparent);`,
			},
			{
				Statement: `insert into cnullchild values(1);`,
			},
			{
				Statement: `insert into cnullchild values(2);`,
			},
			{
				Statement: `insert into cnullchild values(null);`,
			},
			{
				Statement: `select * from cnullparent;`,
				Results:   []sql.Row{{1}, {2}, {``}},
			},
			{
				Statement: `select * from cnullparent where f1 = 2;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `drop table cnullparent cascade;`,
			},
			{
				Statement: `create table inh_perm_parent (a1 int);`,
			},
			{
				Statement: `create temp table inh_temp_parent (a1 int);`,
			},
			{
				Statement: `create temp table inh_temp_child () inherits (inh_perm_parent); -- ok`,
			},
			{
				Statement:   `create table inh_perm_child () inherits (inh_temp_parent); -- error`,
				ErrorString: `cannot inherit from temporary relation "inh_temp_parent"`,
			},
			{
				Statement: `create temp table inh_temp_child_2 () inherits (inh_temp_parent); -- ok`,
			},
			{
				Statement: `insert into inh_perm_parent values (1);`,
			},
			{
				Statement: `insert into inh_temp_parent values (2);`,
			},
			{
				Statement: `insert into inh_temp_child values (3);`,
			},
			{
				Statement: `insert into inh_temp_child_2 values (4);`,
			},
			{
				Statement: `select tableoid::regclass, a1 from inh_perm_parent;`,
				Results:   []sql.Row{{`inh_perm_parent`, 1}, {`inh_temp_child`, 3}},
			},
			{
				Statement: `select tableoid::regclass, a1 from inh_temp_parent;`,
				Results:   []sql.Row{{`inh_temp_parent`, 2}, {`inh_temp_child_2`, 4}},
			},
			{
				Statement: `drop table inh_perm_parent cascade;`,
			},
			{
				Statement: `drop table inh_temp_parent cascade;`,
			},
			{
				Statement: `create table list_parted (
	a	varchar
) partition by list (a);`,
			},
			{
				Statement: `create table part_ab_cd partition of list_parted for values in ('ab', 'cd');`,
			},
			{
				Statement: `create table part_ef_gh partition of list_parted for values in ('ef', 'gh');`,
			},
			{
				Statement: `create table part_null_xy partition of list_parted for values in (null, 'xy');`,
			},
			{
				Statement: `explain (costs off) select * from list_parted;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_ab_cd list_parted_1`}, {`->  Seq Scan on part_ef_gh list_parted_2`}, {`->  Seq Scan on part_null_xy list_parted_3`}},
			},
			{
				Statement: `explain (costs off) select * from list_parted where a is null;`,
				Results:   []sql.Row{{`Seq Scan on part_null_xy list_parted`}, {`Filter: (a IS NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from list_parted where a is not null;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_ab_cd list_parted_1`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on part_ef_gh list_parted_2`}, {`Filter: (a IS NOT NULL)`}, {`->  Seq Scan on part_null_xy list_parted_3`}, {`Filter: (a IS NOT NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from list_parted where a in ('ab', 'cd', 'ef');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_ab_cd list_parted_1`}, {`Filter: ((a)::text = ANY ('{ab,cd,ef}'::text[]))`}, {`->  Seq Scan on part_ef_gh list_parted_2`}, {`Filter: ((a)::text = ANY ('{ab,cd,ef}'::text[]))`}},
			},
			{
				Statement: `explain (costs off) select * from list_parted where a = 'ab' or a in (null, 'cd');`,
				Results:   []sql.Row{{`Seq Scan on part_ab_cd list_parted`}, {`Filter: (((a)::text = 'ab'::text) OR ((a)::text = ANY ('{NULL,cd}'::text[])))`}},
			},
			{
				Statement: `explain (costs off) select * from list_parted where a = 'ab';`,
				Results:   []sql.Row{{`Seq Scan on part_ab_cd list_parted`}, {`Filter: ((a)::text = 'ab'::text)`}},
			},
			{
				Statement: `create table range_list_parted (
	a	int,
	b	char(2)
) partition by range (a);`,
			},
			{
				Statement: `create table part_1_10 partition of range_list_parted for values from (1) to (10) partition by list (b);`,
			},
			{
				Statement: `create table part_1_10_ab partition of part_1_10 for values in ('ab');`,
			},
			{
				Statement: `create table part_1_10_cd partition of part_1_10 for values in ('cd');`,
			},
			{
				Statement: `create table part_10_20 partition of range_list_parted for values from (10) to (20) partition by list (b);`,
			},
			{
				Statement: `create table part_10_20_ab partition of part_10_20 for values in ('ab');`,
			},
			{
				Statement: `create table part_10_20_cd partition of part_10_20 for values in ('cd');`,
			},
			{
				Statement: `create table part_21_30 partition of range_list_parted for values from (21) to (30) partition by list (b);`,
			},
			{
				Statement: `create table part_21_30_ab partition of part_21_30 for values in ('ab');`,
			},
			{
				Statement: `create table part_21_30_cd partition of part_21_30 for values in ('cd');`,
			},
			{
				Statement: `create table part_40_inf partition of range_list_parted for values from (40) to (maxvalue) partition by list (b);`,
			},
			{
				Statement: `create table part_40_inf_ab partition of part_40_inf for values in ('ab');`,
			},
			{
				Statement: `create table part_40_inf_cd partition of part_40_inf for values in ('cd');`,
			},
			{
				Statement: `create table part_40_inf_null partition of part_40_inf for values in (null);`,
			},
			{
				Statement: `explain (costs off) select * from range_list_parted;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_1_10_ab range_list_parted_1`}, {`->  Seq Scan on part_1_10_cd range_list_parted_2`}, {`->  Seq Scan on part_10_20_ab range_list_parted_3`}, {`->  Seq Scan on part_10_20_cd range_list_parted_4`}, {`->  Seq Scan on part_21_30_ab range_list_parted_5`}, {`->  Seq Scan on part_21_30_cd range_list_parted_6`}, {`->  Seq Scan on part_40_inf_ab range_list_parted_7`}, {`->  Seq Scan on part_40_inf_cd range_list_parted_8`}, {`->  Seq Scan on part_40_inf_null range_list_parted_9`}},
			},
			{
				Statement: `explain (costs off) select * from range_list_parted where a = 5;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_1_10_ab range_list_parted_1`}, {`Filter: (a = 5)`}, {`->  Seq Scan on part_1_10_cd range_list_parted_2`}, {`Filter: (a = 5)`}},
			},
			{
				Statement: `explain (costs off) select * from range_list_parted where b = 'ab';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_1_10_ab range_list_parted_1`}, {`Filter: (b = 'ab'::bpchar)`}, {`->  Seq Scan on part_10_20_ab range_list_parted_2`}, {`Filter: (b = 'ab'::bpchar)`}, {`->  Seq Scan on part_21_30_ab range_list_parted_3`}, {`Filter: (b = 'ab'::bpchar)`}, {`->  Seq Scan on part_40_inf_ab range_list_parted_4`}, {`Filter: (b = 'ab'::bpchar)`}},
			},
			{
				Statement: `explain (costs off) select * from range_list_parted where a between 3 and 23 and b in ('ab');`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_1_10_ab range_list_parted_1`}, {`Filter: ((a >= 3) AND (a <= 23) AND (b = 'ab'::bpchar))`}, {`->  Seq Scan on part_10_20_ab range_list_parted_2`}, {`Filter: ((a >= 3) AND (a <= 23) AND (b = 'ab'::bpchar))`}, {`->  Seq Scan on part_21_30_ab range_list_parted_3`}, {`Filter: ((a >= 3) AND (a <= 23) AND (b = 'ab'::bpchar))`}},
			},
			{
				Statement: `/* Should select no rows because range partition key cannot be null */
explain (costs off) select * from range_list_parted where a is null;`,
				Results: []sql.Row{{`Result`}, {`One-Time Filter: false`}},
			},
			{
				Statement: `/* Should only select rows from the null-accepting partition */
explain (costs off) select * from range_list_parted where b is null;`,
				Results: []sql.Row{{`Seq Scan on part_40_inf_null range_list_parted`}, {`Filter: (b IS NULL)`}},
			},
			{
				Statement: `explain (costs off) select * from range_list_parted where a is not null and a < 67;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_1_10_ab range_list_parted_1`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}, {`->  Seq Scan on part_1_10_cd range_list_parted_2`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}, {`->  Seq Scan on part_10_20_ab range_list_parted_3`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}, {`->  Seq Scan on part_10_20_cd range_list_parted_4`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}, {`->  Seq Scan on part_21_30_ab range_list_parted_5`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}, {`->  Seq Scan on part_21_30_cd range_list_parted_6`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}, {`->  Seq Scan on part_40_inf_ab range_list_parted_7`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}, {`->  Seq Scan on part_40_inf_cd range_list_parted_8`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}, {`->  Seq Scan on part_40_inf_null range_list_parted_9`}, {`Filter: ((a IS NOT NULL) AND (a < 67))`}},
			},
			{
				Statement: `explain (costs off) select * from range_list_parted where a >= 30;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on part_40_inf_ab range_list_parted_1`}, {`Filter: (a >= 30)`}, {`->  Seq Scan on part_40_inf_cd range_list_parted_2`}, {`Filter: (a >= 30)`}, {`->  Seq Scan on part_40_inf_null range_list_parted_3`}, {`Filter: (a >= 30)`}},
			},
			{
				Statement: `drop table list_parted;`,
			},
			{
				Statement: `drop table range_list_parted;`,
			},
			{
				Statement: `create table mcrparted (a int, b int, c int) partition by range (a, abs(b), c);`,
			},
			{
				Statement: `create table mcrparted_def partition of mcrparted default;`,
			},
			{
				Statement: `create table mcrparted0 partition of mcrparted for values from (minvalue, minvalue, minvalue) to (1, 1, 1);`,
			},
			{
				Statement: `create table mcrparted1 partition of mcrparted for values from (1, 1, 1) to (10, 5, 10);`,
			},
			{
				Statement: `create table mcrparted2 partition of mcrparted for values from (10, 5, 10) to (10, 10, 10);`,
			},
			{
				Statement: `create table mcrparted3 partition of mcrparted for values from (11, 1, 1) to (20, 10, 10);`,
			},
			{
				Statement: `create table mcrparted4 partition of mcrparted for values from (20, 10, 10) to (20, 20, 20);`,
			},
			{
				Statement: `create table mcrparted5 partition of mcrparted for values from (20, 20, 20) to (maxvalue, maxvalue, maxvalue);`,
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a = 0;	-- scans mcrparted0, mcrparted_def`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mcrparted0 mcrparted_1`}, {`Filter: (a = 0)`}, {`->  Seq Scan on mcrparted_def mcrparted_2`}, {`Filter: (a = 0)`}},
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a = 10 and abs(b) < 5;	-- scans mcrparted1, mcrparted_def`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mcrparted1 mcrparted_1`}, {`Filter: ((a = 10) AND (abs(b) < 5))`}, {`->  Seq Scan on mcrparted_def mcrparted_2`}, {`Filter: ((a = 10) AND (abs(b) < 5))`}},
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a = 10 and abs(b) = 5;	-- scans mcrparted1, mcrparted2, mcrparted_def`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mcrparted1 mcrparted_1`}, {`Filter: ((a = 10) AND (abs(b) = 5))`}, {`->  Seq Scan on mcrparted2 mcrparted_2`}, {`Filter: ((a = 10) AND (abs(b) = 5))`}, {`->  Seq Scan on mcrparted_def mcrparted_3`}, {`Filter: ((a = 10) AND (abs(b) = 5))`}},
			},
			{
				Statement: `explain (costs off) select * from mcrparted where abs(b) = 5;	-- scans all partitions`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mcrparted0 mcrparted_1`}, {`Filter: (abs(b) = 5)`}, {`->  Seq Scan on mcrparted1 mcrparted_2`}, {`Filter: (abs(b) = 5)`}, {`->  Seq Scan on mcrparted2 mcrparted_3`}, {`Filter: (abs(b) = 5)`}, {`->  Seq Scan on mcrparted3 mcrparted_4`}, {`Filter: (abs(b) = 5)`}, {`->  Seq Scan on mcrparted4 mcrparted_5`}, {`Filter: (abs(b) = 5)`}, {`->  Seq Scan on mcrparted5 mcrparted_6`}, {`Filter: (abs(b) = 5)`}, {`->  Seq Scan on mcrparted_def mcrparted_7`}, {`Filter: (abs(b) = 5)`}},
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a > -1;	-- scans all partitions`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mcrparted0 mcrparted_1`}, {`Filter: (a > '-1'::integer)`}, {`->  Seq Scan on mcrparted1 mcrparted_2`}, {`Filter: (a > '-1'::integer)`}, {`->  Seq Scan on mcrparted2 mcrparted_3`}, {`Filter: (a > '-1'::integer)`}, {`->  Seq Scan on mcrparted3 mcrparted_4`}, {`Filter: (a > '-1'::integer)`}, {`->  Seq Scan on mcrparted4 mcrparted_5`}, {`Filter: (a > '-1'::integer)`}, {`->  Seq Scan on mcrparted5 mcrparted_6`}, {`Filter: (a > '-1'::integer)`}, {`->  Seq Scan on mcrparted_def mcrparted_7`}, {`Filter: (a > '-1'::integer)`}},
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a = 20 and abs(b) = 10 and c > 10;	-- scans mcrparted4`,
				Results:   []sql.Row{{`Seq Scan on mcrparted4 mcrparted`}, {`Filter: ((c > 10) AND (a = 20) AND (abs(b) = 10))`}},
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a = 20 and c > 20; -- scans mcrparted3, mcrparte4, mcrparte5, mcrparted_def`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on mcrparted3 mcrparted_1`}, {`Filter: ((c > 20) AND (a = 20))`}, {`->  Seq Scan on mcrparted4 mcrparted_2`}, {`Filter: ((c > 20) AND (a = 20))`}, {`->  Seq Scan on mcrparted5 mcrparted_3`}, {`Filter: ((c > 20) AND (a = 20))`}, {`->  Seq Scan on mcrparted_def mcrparted_4`}, {`Filter: ((c > 20) AND (a = 20))`}},
			},
			{
				Statement: `create table parted_minmax (a int, b varchar(16)) partition by range (a);`,
			},
			{
				Statement: `create table parted_minmax1 partition of parted_minmax for values from (1) to (10);`,
			},
			{
				Statement: `create index parted_minmax1i on parted_minmax1 (a, b);`,
			},
			{
				Statement: `insert into parted_minmax values (1,'12345');`,
			},
			{
				Statement: `explain (costs off) select min(a), max(a) from parted_minmax where b = '12345';`,
				Results:   []sql.Row{{`Result`}, {`InitPlan 1 (returns $0)`}, {`->  Limit`}, {`->  Index Only Scan using parted_minmax1i on parted_minmax1 parted_minmax`}, {`Index Cond: ((a IS NOT NULL) AND (b = '12345'::text))`}, {`InitPlan 2 (returns $1)`}, {`->  Limit`}, {`->  Index Only Scan Backward using parted_minmax1i on parted_minmax1 parted_minmax_1`}, {`Index Cond: ((a IS NOT NULL) AND (b = '12345'::text))`}},
			},
			{
				Statement: `select min(a), max(a) from parted_minmax where b = '12345';`,
				Results:   []sql.Row{{1, 1}},
			},
			{
				Statement: `drop table parted_minmax;`,
			},
			{
				Statement: `create index mcrparted_a_abs_c_idx on mcrparted (a, abs(b), c);`,
			},
			{
				Statement: `explain (costs off) select * from mcrparted order by a, abs(b), c;`,
				Results:   []sql.Row{{`Merge Append`}, {`Sort Key: mcrparted.a, (abs(mcrparted.b)), mcrparted.c`}, {`->  Index Scan using mcrparted0_a_abs_c_idx on mcrparted0 mcrparted_1`}, {`->  Index Scan using mcrparted1_a_abs_c_idx on mcrparted1 mcrparted_2`}, {`->  Index Scan using mcrparted2_a_abs_c_idx on mcrparted2 mcrparted_3`}, {`->  Index Scan using mcrparted3_a_abs_c_idx on mcrparted3 mcrparted_4`}, {`->  Index Scan using mcrparted4_a_abs_c_idx on mcrparted4 mcrparted_5`}, {`->  Index Scan using mcrparted5_a_abs_c_idx on mcrparted5 mcrparted_6`}, {`->  Index Scan using mcrparted_def_a_abs_c_idx on mcrparted_def mcrparted_7`}},
			},
			{
				Statement: `drop table mcrparted_def;`,
			},
			{
				Statement: `explain (costs off) select * from mcrparted order by a, abs(b), c;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Scan using mcrparted0_a_abs_c_idx on mcrparted0 mcrparted_1`}, {`->  Index Scan using mcrparted1_a_abs_c_idx on mcrparted1 mcrparted_2`}, {`->  Index Scan using mcrparted2_a_abs_c_idx on mcrparted2 mcrparted_3`}, {`->  Index Scan using mcrparted3_a_abs_c_idx on mcrparted3 mcrparted_4`}, {`->  Index Scan using mcrparted4_a_abs_c_idx on mcrparted4 mcrparted_5`}, {`->  Index Scan using mcrparted5_a_abs_c_idx on mcrparted5 mcrparted_6`}},
			},
			{
				Statement: `explain (costs off) select * from mcrparted order by a desc, abs(b) desc, c desc;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Scan Backward using mcrparted5_a_abs_c_idx on mcrparted5 mcrparted_6`}, {`->  Index Scan Backward using mcrparted4_a_abs_c_idx on mcrparted4 mcrparted_5`}, {`->  Index Scan Backward using mcrparted3_a_abs_c_idx on mcrparted3 mcrparted_4`}, {`->  Index Scan Backward using mcrparted2_a_abs_c_idx on mcrparted2 mcrparted_3`}, {`->  Index Scan Backward using mcrparted1_a_abs_c_idx on mcrparted1 mcrparted_2`}, {`->  Index Scan Backward using mcrparted0_a_abs_c_idx on mcrparted0 mcrparted_1`}},
			},
			{
				Statement: `drop table mcrparted5;`,
			},
			{
				Statement: `create table mcrparted5 partition of mcrparted for values from (20, 20, 20) to (maxvalue, maxvalue, maxvalue) partition by list (a);`,
			},
			{
				Statement: `create table mcrparted5a partition of mcrparted5 for values in(20);`,
			},
			{
				Statement: `create table mcrparted5_def partition of mcrparted5 default;`,
			},
			{
				Statement: `explain (costs off) select * from mcrparted order by a, abs(b), c;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Scan using mcrparted0_a_abs_c_idx on mcrparted0 mcrparted_1`}, {`->  Index Scan using mcrparted1_a_abs_c_idx on mcrparted1 mcrparted_2`}, {`->  Index Scan using mcrparted2_a_abs_c_idx on mcrparted2 mcrparted_3`}, {`->  Index Scan using mcrparted3_a_abs_c_idx on mcrparted3 mcrparted_4`}, {`->  Index Scan using mcrparted4_a_abs_c_idx on mcrparted4 mcrparted_5`}, {`->  Merge Append`}, {`Sort Key: mcrparted_7.a, (abs(mcrparted_7.b)), mcrparted_7.c`}, {`->  Index Scan using mcrparted5a_a_abs_c_idx on mcrparted5a mcrparted_7`}, {`->  Index Scan using mcrparted5_def_a_abs_c_idx on mcrparted5_def mcrparted_8`}},
			},
			{
				Statement: `drop table mcrparted5_def;`,
			},
			{
				Statement: `explain (costs off) select a, abs(b) from mcrparted order by a, abs(b), c;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Scan using mcrparted0_a_abs_c_idx on mcrparted0 mcrparted_1`}, {`->  Index Scan using mcrparted1_a_abs_c_idx on mcrparted1 mcrparted_2`}, {`->  Index Scan using mcrparted2_a_abs_c_idx on mcrparted2 mcrparted_3`}, {`->  Index Scan using mcrparted3_a_abs_c_idx on mcrparted3 mcrparted_4`}, {`->  Index Scan using mcrparted4_a_abs_c_idx on mcrparted4 mcrparted_5`}, {`->  Index Scan using mcrparted5a_a_abs_c_idx on mcrparted5a mcrparted_6`}},
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a < 20 order by a, abs(b), c;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Scan using mcrparted0_a_abs_c_idx on mcrparted0 mcrparted_1`}, {`Index Cond: (a < 20)`}, {`->  Index Scan using mcrparted1_a_abs_c_idx on mcrparted1 mcrparted_2`}, {`Index Cond: (a < 20)`}, {`->  Index Scan using mcrparted2_a_abs_c_idx on mcrparted2 mcrparted_3`}, {`Index Cond: (a < 20)`}, {`->  Index Scan using mcrparted3_a_abs_c_idx on mcrparted3 mcrparted_4`}, {`Index Cond: (a < 20)`}},
			},
			{
				Statement: `set enable_bitmapscan to off;`,
			},
			{
				Statement: `set enable_sort to off;`,
			},
			{
				Statement: `create table mclparted (a int) partition by list(a);`,
			},
			{
				Statement: `create table mclparted1 partition of mclparted for values in(1);`,
			},
			{
				Statement: `create table mclparted2 partition of mclparted for values in(2);`,
			},
			{
				Statement: `create index on mclparted (a);`,
			},
			{
				Statement: `explain (costs off) select * from mclparted order by a;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using mclparted1_a_idx on mclparted1 mclparted_1`}, {`->  Index Only Scan using mclparted2_a_idx on mclparted2 mclparted_2`}},
			},
			{
				Statement: `create table mclparted3_5 partition of mclparted for values in(3,5);`,
			},
			{
				Statement: `create table mclparted4 partition of mclparted for values in(4);`,
			},
			{
				Statement: `explain (costs off) select * from mclparted order by a;`,
				Results:   []sql.Row{{`Merge Append`}, {`Sort Key: mclparted.a`}, {`->  Index Only Scan using mclparted1_a_idx on mclparted1 mclparted_1`}, {`->  Index Only Scan using mclparted2_a_idx on mclparted2 mclparted_2`}, {`->  Index Only Scan using mclparted3_5_a_idx on mclparted3_5 mclparted_3`}, {`->  Index Only Scan using mclparted4_a_idx on mclparted4 mclparted_4`}},
			},
			{
				Statement: `explain (costs off) select * from mclparted where a in(3,4,5) order by a;`,
				Results:   []sql.Row{{`Merge Append`}, {`Sort Key: mclparted.a`}, {`->  Index Only Scan using mclparted3_5_a_idx on mclparted3_5 mclparted_1`}, {`Index Cond: (a = ANY ('{3,4,5}'::integer[]))`}, {`->  Index Only Scan using mclparted4_a_idx on mclparted4 mclparted_2`}, {`Index Cond: (a = ANY ('{3,4,5}'::integer[]))`}},
			},
			{
				Statement: `create table mclparted_null partition of mclparted for values in(null);`,
			},
			{
				Statement: `create table mclparted_def partition of mclparted default;`,
			},
			{
				Statement: `explain (costs off) select * from mclparted where a in(1,2,4) order by a;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using mclparted1_a_idx on mclparted1 mclparted_1`}, {`Index Cond: (a = ANY ('{1,2,4}'::integer[]))`}, {`->  Index Only Scan using mclparted2_a_idx on mclparted2 mclparted_2`}, {`Index Cond: (a = ANY ('{1,2,4}'::integer[]))`}, {`->  Index Only Scan using mclparted4_a_idx on mclparted4 mclparted_3`}, {`Index Cond: (a = ANY ('{1,2,4}'::integer[]))`}},
			},
			{
				Statement: `explain (costs off) select * from mclparted where a in(1,2,4) or a is null order by a;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using mclparted1_a_idx on mclparted1 mclparted_1`}, {`Filter: ((a = ANY ('{1,2,4}'::integer[])) OR (a IS NULL))`}, {`->  Index Only Scan using mclparted2_a_idx on mclparted2 mclparted_2`}, {`Filter: ((a = ANY ('{1,2,4}'::integer[])) OR (a IS NULL))`}, {`->  Index Only Scan using mclparted4_a_idx on mclparted4 mclparted_3`}, {`Filter: ((a = ANY ('{1,2,4}'::integer[])) OR (a IS NULL))`}, {`->  Index Only Scan using mclparted_null_a_idx on mclparted_null mclparted_4`}, {`Filter: ((a = ANY ('{1,2,4}'::integer[])) OR (a IS NULL))`}},
			},
			{
				Statement: `drop table mclparted_null;`,
			},
			{
				Statement: `create table mclparted_0_null partition of mclparted for values in(0,null);`,
			},
			{
				Statement: `explain (costs off) select * from mclparted where a in(1,2,4) or a is null order by a;`,
				Results:   []sql.Row{{`Merge Append`}, {`Sort Key: mclparted.a`}, {`->  Index Only Scan using mclparted_0_null_a_idx on mclparted_0_null mclparted_1`}, {`Filter: ((a = ANY ('{1,2,4}'::integer[])) OR (a IS NULL))`}, {`->  Index Only Scan using mclparted1_a_idx on mclparted1 mclparted_2`}, {`Filter: ((a = ANY ('{1,2,4}'::integer[])) OR (a IS NULL))`}, {`->  Index Only Scan using mclparted2_a_idx on mclparted2 mclparted_3`}, {`Filter: ((a = ANY ('{1,2,4}'::integer[])) OR (a IS NULL))`}, {`->  Index Only Scan using mclparted4_a_idx on mclparted4 mclparted_4`}, {`Filter: ((a = ANY ('{1,2,4}'::integer[])) OR (a IS NULL))`}},
			},
			{
				Statement: `explain (costs off) select * from mclparted where a in(0,1,2,4) order by a;`,
				Results:   []sql.Row{{`Merge Append`}, {`Sort Key: mclparted.a`}, {`->  Index Only Scan using mclparted_0_null_a_idx on mclparted_0_null mclparted_1`}, {`Index Cond: (a = ANY ('{0,1,2,4}'::integer[]))`}, {`->  Index Only Scan using mclparted1_a_idx on mclparted1 mclparted_2`}, {`Index Cond: (a = ANY ('{0,1,2,4}'::integer[]))`}, {`->  Index Only Scan using mclparted2_a_idx on mclparted2 mclparted_3`}, {`Index Cond: (a = ANY ('{0,1,2,4}'::integer[]))`}, {`->  Index Only Scan using mclparted4_a_idx on mclparted4 mclparted_4`}, {`Index Cond: (a = ANY ('{0,1,2,4}'::integer[]))`}},
			},
			{
				Statement: `explain (costs off) select * from mclparted where a in(1,2,4) order by a;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using mclparted1_a_idx on mclparted1 mclparted_1`}, {`Index Cond: (a = ANY ('{1,2,4}'::integer[]))`}, {`->  Index Only Scan using mclparted2_a_idx on mclparted2 mclparted_2`}, {`Index Cond: (a = ANY ('{1,2,4}'::integer[]))`}, {`->  Index Only Scan using mclparted4_a_idx on mclparted4 mclparted_3`}, {`Index Cond: (a = ANY ('{1,2,4}'::integer[]))`}},
			},
			{
				Statement: `explain (costs off) select * from mclparted where a in(1,2,4,100) order by a;`,
				Results:   []sql.Row{{`Merge Append`}, {`Sort Key: mclparted.a`}, {`->  Index Only Scan using mclparted1_a_idx on mclparted1 mclparted_1`}, {`Index Cond: (a = ANY ('{1,2,4,100}'::integer[]))`}, {`->  Index Only Scan using mclparted2_a_idx on mclparted2 mclparted_2`}, {`Index Cond: (a = ANY ('{1,2,4,100}'::integer[]))`}, {`->  Index Only Scan using mclparted4_a_idx on mclparted4 mclparted_3`}, {`Index Cond: (a = ANY ('{1,2,4,100}'::integer[]))`}, {`->  Index Only Scan using mclparted_def_a_idx on mclparted_def mclparted_4`}, {`Index Cond: (a = ANY ('{1,2,4,100}'::integer[]))`}},
			},
			{
				Statement: `drop table mclparted;`,
			},
			{
				Statement: `reset enable_sort;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `drop index mcrparted_a_abs_c_idx;`,
			},
			{
				Statement: `create index on mcrparted1 (a, abs(b), c);`,
			},
			{
				Statement: `create index on mcrparted2 (a, abs(b), c);`,
			},
			{
				Statement: `create index on mcrparted3 (a, abs(b), c);`,
			},
			{
				Statement: `create index on mcrparted4 (a, abs(b), c);`,
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a < 20 order by a, abs(b), c limit 1;`,
				Results:   []sql.Row{{`Limit`}, {`->  Append`}, {`->  Sort`}, {`Sort Key: mcrparted_1.a, (abs(mcrparted_1.b)), mcrparted_1.c`}, {`->  Seq Scan on mcrparted0 mcrparted_1`}, {`Filter: (a < 20)`}, {`->  Index Scan using mcrparted1_a_abs_c_idx on mcrparted1 mcrparted_2`}, {`Index Cond: (a < 20)`}, {`->  Index Scan using mcrparted2_a_abs_c_idx on mcrparted2 mcrparted_3`}, {`Index Cond: (a < 20)`}, {`->  Index Scan using mcrparted3_a_abs_c_idx on mcrparted3 mcrparted_4`}, {`Index Cond: (a < 20)`}},
			},
			{
				Statement: `set enable_bitmapscan = 0;`,
			},
			{
				Statement: `explain (costs off) select * from mcrparted where a = 10 order by a, abs(b), c;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Scan using mcrparted1_a_abs_c_idx on mcrparted1 mcrparted_1`}, {`Index Cond: (a = 10)`}, {`->  Index Scan using mcrparted2_a_abs_c_idx on mcrparted2 mcrparted_2`}, {`Index Cond: (a = 10)`}},
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `drop table mcrparted;`,
			},
			{
				Statement: `create table bool_lp (b bool) partition by list(b);`,
			},
			{
				Statement: `create table bool_lp_true partition of bool_lp for values in(true);`,
			},
			{
				Statement: `create table bool_lp_false partition of bool_lp for values in(false);`,
			},
			{
				Statement: `create index on bool_lp (b);`,
			},
			{
				Statement: `explain (costs off) select * from bool_lp order by b;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using bool_lp_false_b_idx on bool_lp_false bool_lp_1`}, {`->  Index Only Scan using bool_lp_true_b_idx on bool_lp_true bool_lp_2`}},
			},
			{
				Statement: `drop table bool_lp;`,
			},
			{
				Statement: `create table bool_rp (b bool, a int) partition by range(b,a);`,
			},
			{
				Statement: `create table bool_rp_false_1k partition of bool_rp for values from (false,0) to (false,1000);`,
			},
			{
				Statement: `create table bool_rp_true_1k partition of bool_rp for values from (true,0) to (true,1000);`,
			},
			{
				Statement: `create table bool_rp_false_2k partition of bool_rp for values from (false,1000) to (false,2000);`,
			},
			{
				Statement: `create table bool_rp_true_2k partition of bool_rp for values from (true,1000) to (true,2000);`,
			},
			{
				Statement: `create index on bool_rp (b,a);`,
			},
			{
				Statement: `explain (costs off) select * from bool_rp where b = true order by b,a;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using bool_rp_true_1k_b_a_idx on bool_rp_true_1k bool_rp_1`}, {`Index Cond: (b = true)`}, {`->  Index Only Scan using bool_rp_true_2k_b_a_idx on bool_rp_true_2k bool_rp_2`}, {`Index Cond: (b = true)`}},
			},
			{
				Statement: `explain (costs off) select * from bool_rp where b = false order by b,a;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using bool_rp_false_1k_b_a_idx on bool_rp_false_1k bool_rp_1`}, {`Index Cond: (b = false)`}, {`->  Index Only Scan using bool_rp_false_2k_b_a_idx on bool_rp_false_2k bool_rp_2`}, {`Index Cond: (b = false)`}},
			},
			{
				Statement: `explain (costs off) select * from bool_rp where b = true order by a;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using bool_rp_true_1k_b_a_idx on bool_rp_true_1k bool_rp_1`}, {`Index Cond: (b = true)`}, {`->  Index Only Scan using bool_rp_true_2k_b_a_idx on bool_rp_true_2k bool_rp_2`}, {`Index Cond: (b = true)`}},
			},
			{
				Statement: `explain (costs off) select * from bool_rp where b = false order by a;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using bool_rp_false_1k_b_a_idx on bool_rp_false_1k bool_rp_1`}, {`Index Cond: (b = false)`}, {`->  Index Only Scan using bool_rp_false_2k_b_a_idx on bool_rp_false_2k bool_rp_2`}, {`Index Cond: (b = false)`}},
			},
			{
				Statement: `drop table bool_rp;`,
			},
			{
				Statement: `create table range_parted (a int, b int, c int) partition by range(a, b);`,
			},
			{
				Statement: `create table range_parted1 partition of range_parted for values from (0,0) to (10,10);`,
			},
			{
				Statement: `create table range_parted2 partition of range_parted for values from (10,10) to (20,20);`,
			},
			{
				Statement: `create index on range_parted (a,b,c);`,
			},
			{
				Statement: `explain (costs off) select * from range_parted order by a,b,c;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan using range_parted1_a_b_c_idx on range_parted1 range_parted_1`}, {`->  Index Only Scan using range_parted2_a_b_c_idx on range_parted2 range_parted_2`}},
			},
			{
				Statement: `explain (costs off) select * from range_parted order by a desc,b desc,c desc;`,
				Results:   []sql.Row{{`Append`}, {`->  Index Only Scan Backward using range_parted2_a_b_c_idx on range_parted2 range_parted_2`}, {`->  Index Only Scan Backward using range_parted1_a_b_c_idx on range_parted1 range_parted_1`}},
			},
			{
				Statement: `drop table range_parted;`,
			},
			{
				Statement: `create table permtest_parent (a int, b text, c text) partition by list (a);`,
			},
			{
				Statement: `create table permtest_child (b text, c text, a int) partition by list (b);`,
			},
			{
				Statement: `create table permtest_grandchild (c text, b text, a int);`,
			},
			{
				Statement: `alter table permtest_child attach partition permtest_grandchild for values in ('a');`,
			},
			{
				Statement: `alter table permtest_parent attach partition permtest_child for values in (1);`,
			},
			{
				Statement: `create index on permtest_parent (left(c, 3));`,
			},
			{
				Statement: `insert into permtest_parent
  select 1, 'a', left(md5(i::text), 5) from generate_series(0, 100) i;`,
			},
			{
				Statement: `analyze permtest_parent;`,
			},
			{
				Statement: `create role regress_no_child_access;`,
			},
			{
				Statement: `revoke all on permtest_grandchild from regress_no_child_access;`,
			},
			{
				Statement: `grant select on permtest_parent to regress_no_child_access;`,
			},
			{
				Statement: `set session authorization regress_no_child_access;`,
			},
			{
				Statement: `explain (costs off)
  select * from permtest_parent p1 inner join permtest_parent p2
  on p1.a = p2.a and p1.c ~ 'a1$';`,
				Results: []sql.Row{{`Nested Loop`}, {`Join Filter: (p1.a = p2.a)`}, {`->  Seq Scan on permtest_grandchild p1`}, {`Filter: (c ~ 'a1$'::text)`}, {`->  Seq Scan on permtest_grandchild p2`}},
			},
			{
				Statement: `explain (costs off)
  select * from permtest_parent p1 inner join permtest_parent p2
  on p1.a = p2.a and left(p1.c, 3) ~ 'a1$';`,
				Results: []sql.Row{{`Nested Loop`}, {`Join Filter: (p1.a = p2.a)`}, {`->  Seq Scan on permtest_grandchild p1`}, {`Filter: ("left"(c, 3) ~ 'a1$'::text)`}, {`->  Seq Scan on permtest_grandchild p2`}},
			},
			{
				Statement: `reset session authorization;`,
			},
			{
				Statement: `revoke all on permtest_parent from regress_no_child_access;`,
			},
			{
				Statement: `grant select(a,c) on permtest_parent to regress_no_child_access;`,
			},
			{
				Statement: `set session authorization regress_no_child_access;`,
			},
			{
				Statement: `explain (costs off)
  select p2.a, p1.c from permtest_parent p1 inner join permtest_parent p2
  on p1.a = p2.a and p1.c ~ 'a1$';`,
				Results: []sql.Row{{`Nested Loop`}, {`Join Filter: (p1.a = p2.a)`}, {`->  Seq Scan on permtest_grandchild p1`}, {`Filter: (c ~ 'a1$'::text)`}, {`->  Seq Scan on permtest_grandchild p2`}},
			},
			{
				Statement: `explain (costs off)
  select p2.a, p1.c from permtest_parent p1 inner join permtest_parent p2
  on p1.a = p2.a and left(p1.c, 3) ~ 'a1$';`,
				Results: []sql.Row{{`Hash Join`}, {`Hash Cond: (p2.a = p1.a)`}, {`->  Seq Scan on permtest_grandchild p2`}, {`->  Hash`}, {`->  Seq Scan on permtest_grandchild p1`}, {`Filter: ("left"(c, 3) ~ 'a1$'::text)`}},
			},
			{
				Statement: `reset session authorization;`,
			},
			{
				Statement: `revoke all on permtest_parent from regress_no_child_access;`,
			},
			{
				Statement: `drop role regress_no_child_access;`,
			},
			{
				Statement: `drop table permtest_parent;`,
			},
			{
				Statement: `CREATE TABLE errtst_parent (
    partid int not null,
    shdata int not null,
    data int NOT NULL DEFAULT 0,
    CONSTRAINT shdata_small CHECK(shdata < 3)
) PARTITION BY RANGE (partid);`,
			},
			{
				Statement: `CREATE TABLE errtst_child_fastdef (
    partid int not null,
    shdata int not null,
    CONSTRAINT shdata_small CHECK(shdata < 3)
);`,
			},
			{
				Statement: `CREATE TABLE errtst_child_plaindef (
    partid int not null,
    shdata int not null,
    data int NOT NULL DEFAULT 0,
    CONSTRAINT shdata_small CHECK(shdata < 3),
    CHECK(data < 10)
);`,
			},
			{
				Statement: `CREATE TABLE errtst_child_reorder (
    data int NOT NULL DEFAULT 0,
    shdata int not null,
    partid int not null,
    CONSTRAINT shdata_small CHECK(shdata < 3),
    CHECK(data < 10)
);`,
			},
			{
				Statement: `ALTER TABLE errtst_child_fastdef ADD COLUMN data int NOT NULL DEFAULT 0;`,
			},
			{
				Statement: `ALTER TABLE errtst_child_fastdef ADD CONSTRAINT errtest_child_fastdef_data_check CHECK (data < 10);`,
			},
			{
				Statement: `ALTER TABLE errtst_parent ATTACH PARTITION errtst_child_fastdef FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `ALTER TABLE errtst_parent ATTACH PARTITION errtst_child_plaindef FOR VALUES FROM (10) TO (20);`,
			},
			{
				Statement: `ALTER TABLE errtst_parent ATTACH PARTITION errtst_child_reorder FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `INSERT INTO errtst_parent(partid, shdata, data) VALUES ( '0', '1', '5');`,
			},
			{
				Statement: `INSERT INTO errtst_parent(partid, shdata, data) VALUES ('10', '1', '5');`,
			},
			{
				Statement: `INSERT INTO errtst_parent(partid, shdata, data) VALUES ('20', '1', '5');`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ( '0', '1', '10');`,
				ErrorString: `new row for relation "errtst_child_fastdef" violates check constraint "errtest_child_fastdef_data_check"`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ('10', '1', '10');`,
				ErrorString: `new row for relation "errtst_child_plaindef" violates check constraint "errtst_child_plaindef_data_check"`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ('20', '1', '10');`,
				ErrorString: `new row for relation "errtst_child_reorder" violates check constraint "errtst_child_reorder_data_check"`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ( '0', '1', NULL);`,
				ErrorString: `null value in column "data" of relation "errtst_child_fastdef" violates not-null constraint`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ('10', '1', NULL);`,
				ErrorString: `null value in column "data" of relation "errtst_child_plaindef" violates not-null constraint`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ('20', '1', NULL);`,
				ErrorString: `null value in column "data" of relation "errtst_child_reorder" violates not-null constraint`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ( '0', '5', '5');`,
				ErrorString: `new row for relation "errtst_child_fastdef" violates check constraint "shdata_small"`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ('10', '5', '5');`,
				ErrorString: `new row for relation "errtst_child_plaindef" violates check constraint "shdata_small"`,
			},
			{
				Statement:   `INSERT INTO errtst_parent(partid, shdata, data) VALUES ('20', '5', '5');`,
				ErrorString: `new row for relation "errtst_child_reorder" violates check constraint "shdata_small"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE errtst_parent SET data = data + 1 WHERE partid = 0;`,
			},
			{
				Statement: `UPDATE errtst_parent SET data = data + 1 WHERE partid = 10;`,
			},
			{
				Statement: `UPDATE errtst_parent SET data = data + 1 WHERE partid = 20;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement:   `UPDATE errtst_parent SET data = data + 10 WHERE partid = 0;`,
				ErrorString: `new row for relation "errtst_child_fastdef" violates check constraint "errtest_child_fastdef_data_check"`,
			},
			{
				Statement:   `UPDATE errtst_parent SET data = data + 10 WHERE partid = 10;`,
				ErrorString: `new row for relation "errtst_child_plaindef" violates check constraint "errtst_child_plaindef_data_check"`,
			},
			{
				Statement:   `UPDATE errtst_parent SET data = data + 10 WHERE partid = 20;`,
				ErrorString: `new row for relation "errtst_child_reorder" violates check constraint "errtst_child_reorder_data_check"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE errtst_child_fastdef SET partid = 1 WHERE partid = 0;`,
			},
			{
				Statement: `UPDATE errtst_child_plaindef SET partid = 11 WHERE partid = 10;`,
			},
			{
				Statement: `UPDATE errtst_child_reorder SET partid = 21 WHERE partid = 20;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement:   `UPDATE errtst_child_fastdef SET partid = partid + 10 WHERE partid = 0;`,
				ErrorString: `new row for relation "errtst_child_fastdef" violates partition constraint`,
			},
			{
				Statement:   `UPDATE errtst_child_plaindef SET partid = partid + 10 WHERE partid = 10;`,
				ErrorString: `new row for relation "errtst_child_plaindef" violates partition constraint`,
			},
			{
				Statement:   `UPDATE errtst_child_reorder SET partid = partid + 10 WHERE partid = 20;`,
				ErrorString: `new row for relation "errtst_child_reorder" violates partition constraint`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `UPDATE errtst_parent SET partid = 10, data = data + 1 WHERE partid = 0;`,
			},
			{
				Statement: `UPDATE errtst_parent SET partid = 20, data = data + 1 WHERE partid = 10;`,
			},
			{
				Statement: `UPDATE errtst_parent SET partid = 0, data = data + 1 WHERE partid = 20;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement:   `UPDATE errtst_parent SET partid = 10, data = data + 10 WHERE partid = 0;`,
				ErrorString: `new row for relation "errtst_child_plaindef" violates check constraint "errtst_child_plaindef_data_check"`,
			},
			{
				Statement:   `UPDATE errtst_parent SET partid = 20, data = data + 10 WHERE partid = 10;`,
				ErrorString: `new row for relation "errtst_child_reorder" violates check constraint "errtst_child_reorder_data_check"`,
			},
			{
				Statement:   `UPDATE errtst_parent SET partid = 0, data = data + 10 WHERE partid = 20;`,
				ErrorString: `new row for relation "errtst_child_fastdef" violates check constraint "errtest_child_fastdef_data_check"`,
			},
			{
				Statement:   `UPDATE errtst_parent SET partid = 30, data = data + 10 WHERE partid = 20;`,
				ErrorString: `no partition of relation "errtst_parent" found for row`,
			},
			{
				Statement: `DROP TABLE errtst_parent;`,
			},
		},
	})
}
