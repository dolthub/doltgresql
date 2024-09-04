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

func TestDomain(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_domain)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_domain,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create domain domaindroptest int4;`,
			},
			{
				Statement: `comment on domain domaindroptest is 'About to drop this..';`,
			},
			{
				Statement: `create domain dependenttypetest domaindroptest;`,
			},
			{
				Statement:   `drop domain domaindroptest;`,
				ErrorString: `cannot drop type domaindroptest because other objects depend on it`,
			},
			{
				Statement: `drop domain domaindroptest cascade;`,
			},
			{
				Statement:   `drop domain domaindroptest cascade;`,
				ErrorString: `type "domaindroptest" does not exist`,
			},
			{
				Statement: `create domain domainvarchar varchar(5);`,
			},
			{
				Statement: `create domain domainnumeric numeric(8,2);`,
			},
			{
				Statement: `create domain domainint4 int4;`,
			},
			{
				Statement: `create domain domaintext text;`,
			},
			{
				Statement: `SELECT cast('123456' as domainvarchar);`,
				Results:   []sql.Row{{12345}},
			},
			{
				Statement: `SELECT cast('12345' as domainvarchar);`,
				Results:   []sql.Row{{12345}},
			},
			{
				Statement: `create table basictest
           ( testint4 domainint4
           , testtext domaintext
           , testvarchar domainvarchar
           , testnumeric domainnumeric
           );`,
			},
			{
				Statement: `INSERT INTO basictest values ('88', 'haha', 'short', '123.12');      -- Good`,
			},
			{
				Statement:   `INSERT INTO basictest values ('88', 'haha', 'short text', '123.12'); -- Bad varchar`,
				ErrorString: `value too long for type character varying(5)`,
			},
			{
				Statement: `INSERT INTO basictest values ('88', 'haha', 'short', '123.1212');    -- Truncate numeric`,
			},
			{
				Statement:   `COPY basictest (testvarchar) FROM stdin; -- fail`,
				ErrorString: `value too long for type character varying(5)`,
			},
			{
				Statement: `CONTEXT:  COPY basictest, line 1, column testvarchar: "notsoshorttext"
COPY basictest (testvarchar) FROM stdin;`,
			},
			{
				Statement: `select * from basictest;`,
				Results:   []sql.Row{{88, `haha`, `short`, 123.12}, {88, `haha`, `short`, 123.12}, {``, ``, `short`, ``}},
			},
			{
				Statement: `select testtext || testvarchar as concat, testnumeric + 42 as sum
from basictest;`,
				Results: []sql.Row{{`hahashort`, 165.12}, {`hahashort`, 165.12}, {``, ``}},
			},
			{
				Statement: `select pg_typeof(coalesce(4::domainint4, 7));`,
				Results:   []sql.Row{{`integer`}},
			},
			{
				Statement: `select pg_typeof(coalesce(4::domainint4, 7::domainint4));`,
				Results:   []sql.Row{{`domainint4`}},
			},
			{
				Statement: `drop table basictest;`,
			},
			{
				Statement: `drop domain domainvarchar restrict;`,
			},
			{
				Statement: `drop domain domainnumeric restrict;`,
			},
			{
				Statement: `drop domain domainint4 restrict;`,
			},
			{
				Statement: `drop domain domaintext;`,
			},
			{
				Statement: `create domain domainint4arr int4[1];`,
			},
			{
				Statement: `create domain domainchar4arr varchar(4)[2][3];`,
			},
			{
				Statement: `create table domarrtest
           ( testint4arr domainint4arr
           , testchar4arr domainchar4arr
            );`,
			},
			{
				Statement: `INSERT INTO domarrtest values ('{2,2}', '{{"a","b"},{"c","d"}}');`,
			},
			{
				Statement: `INSERT INTO domarrtest values ('{{2,2},{2,2}}', '{{"a","b"}}');`,
			},
			{
				Statement: `INSERT INTO domarrtest values ('{2,2}', '{{"a","b"},{"c","d"},{"e","f"}}');`,
			},
			{
				Statement: `INSERT INTO domarrtest values ('{2,2}', '{{"a"},{"c"}}');`,
			},
			{
				Statement: `INSERT INTO domarrtest values (NULL, '{{"a","b","c"},{"d","e","f"}}');`,
			},
			{
				Statement:   `INSERT INTO domarrtest values (NULL, '{{"toolong","b","c"},{"d","e","f"}}');`,
				ErrorString: `value too long for type character varying(4)`,
			},
			{
				Statement: `INSERT INTO domarrtest (testint4arr[1], testint4arr[3]) values (11,22);`,
			},
			{
				Statement: `select * from domarrtest;`,
				Results:   []sql.Row{{`{2,2}`, `{{a,b},{c,d}}`}, {`{{2,2},{2,2}}`, `{{a,b}}`}, {`{2,2}`, `{{a,b},{c,d},{e,f}}`}, {`{2,2}`, `{{a},{c}}`}, {``, `{{a,b,c},{d,e,f}}`}, {`{11,NULL,22}`, ``}},
			},
			{
				Statement: `select testint4arr[1], testchar4arr[2:2] from domarrtest;`,
				Results:   []sql.Row{{2, `{{c,d}}`}, {``, `{}`}, {2, `{{c,d}}`}, {2, `{{c}}`}, {``, `{{d,e,f}}`}, {11, ``}},
			},
			{
				Statement: `select array_dims(testint4arr), array_dims(testchar4arr) from domarrtest;`,
				Results:   []sql.Row{{`[1:2]`, `[1:2][1:2]`}, {`[1:2][1:2]`, `[1:1][1:2]`}, {`[1:2]`, `[1:3][1:2]`}, {`[1:2]`, `[1:2][1:1]`}, {``, `[1:2][1:3]`}, {`[1:3]`, ``}},
			},
			{
				Statement: `COPY domarrtest FROM stdin;`,
			},
			{
				Statement:   `COPY domarrtest FROM stdin;	-- fail`,
				ErrorString: `value too long for type character varying(4)`,
			},
			{
				Statement: `CONTEXT:  COPY domarrtest, line 1, column testchar4arr: "{qwerty,w,e}"
select * from domarrtest;`,
				Results: []sql.Row{{`{2,2}`, `{{a,b},{c,d}}`}, {`{{2,2},{2,2}}`, `{{a,b}}`}, {`{2,2}`, `{{a,b},{c,d},{e,f}}`}, {`{2,2}`, `{{a},{c}}`}, {``, `{{a,b,c},{d,e,f}}`}, {`{11,NULL,22}`, ``}, {`{3,4}`, `{q,w,e}`}, {``, ``}},
			},
			{
				Statement: `update domarrtest set
  testint4arr[1] = testint4arr[1] + 1,
  testint4arr[3] = testint4arr[3] - 1
where testchar4arr is null;`,
			},
			{
				Statement: `select * from domarrtest where testchar4arr is null;`,
				Results:   []sql.Row{{`{12,NULL,21}`, ``}, {`{NULL,NULL,NULL}`, ``}},
			},
			{
				Statement: `drop table domarrtest;`,
			},
			{
				Statement: `drop domain domainint4arr restrict;`,
			},
			{
				Statement: `drop domain domainchar4arr restrict;`,
			},
			{
				Statement: `create domain dia as int[];`,
			},
			{
				Statement: `select '{1,2,3}'::dia;`,
				Results:   []sql.Row{{`{1,2,3}`}},
			},
			{
				Statement: `select array_dims('{1,2,3}'::dia);`,
				Results:   []sql.Row{{`[1:3]`}},
			},
			{
				Statement: `select pg_typeof('{1,2,3}'::dia);`,
				Results:   []sql.Row{{`dia`}},
			},
			{
				Statement: `select pg_typeof('{1,2,3}'::dia || 42); -- should be int[] not dia`,
				Results:   []sql.Row{{`integer[]`}},
			},
			{
				Statement: `drop domain dia;`,
			},
			{
				Statement: `create type comptype as (r float8, i float8);`,
			},
			{
				Statement: `create domain dcomptype as comptype;`,
			},
			{
				Statement: `create table dcomptable (d1 dcomptype unique);`,
			},
			{
				Statement: `insert into dcomptable values (row(1,2)::dcomptype);`,
			},
			{
				Statement: `insert into dcomptable values (row(3,4)::comptype);`,
			},
			{
				Statement:   `insert into dcomptable values (row(1,2)::dcomptype);  -- fail on uniqueness`,
				ErrorString: `duplicate key value violates unique constraint "dcomptable_d1_key"`,
			},
			{
				Statement: `insert into dcomptable (d1.r) values(11);`,
			},
			{
				Statement: `select * from dcomptable;`,
				Results:   []sql.Row{{`(1,2)`}, {`(3,4)`}, {`(11,)`}},
			},
			{
				Statement: `select (d1).r, (d1).i, (d1).* from dcomptable;`,
				Results:   []sql.Row{{1, 2, 1, 2}, {3, 4, 3, 4}, {11, ``, 11, ``}},
			},
			{
				Statement: `update dcomptable set d1.r = (d1).r + 1 where (d1).i > 0;`,
			},
			{
				Statement: `select * from dcomptable;`,
				Results:   []sql.Row{{`(11,)`}, {`(2,2)`}, {`(4,4)`}},
			},
			{
				Statement: `alter domain dcomptype add constraint c1 check ((value).r <= (value).i);`,
			},
			{
				Statement:   `alter domain dcomptype add constraint c2 check ((value).r > (value).i);  -- fail`,
				ErrorString: `column "d1" of table "dcomptable" contains values that violate the new constraint`,
			},
			{
				Statement:   `select row(2,1)::dcomptype;  -- fail`,
				ErrorString: `value for domain dcomptype violates check constraint "c1"`,
			},
			{
				Statement: `insert into dcomptable values (row(1,2)::comptype);`,
			},
			{
				Statement:   `insert into dcomptable values (row(2,1)::comptype);  -- fail`,
				ErrorString: `value for domain dcomptype violates check constraint "c1"`,
			},
			{
				Statement: `insert into dcomptable (d1.r) values(99);`,
			},
			{
				Statement: `insert into dcomptable (d1.r, d1.i) values(99, 100);`,
			},
			{
				Statement:   `insert into dcomptable (d1.r, d1.i) values(100, 99);  -- fail`,
				ErrorString: `value for domain dcomptype violates check constraint "c1"`,
			},
			{
				Statement:   `update dcomptable set d1.r = (d1).r + 1 where (d1).i > 0;  -- fail`,
				ErrorString: `value for domain dcomptype violates check constraint "c1"`,
			},
			{
				Statement: `update dcomptable set d1.r = (d1).r - 1, d1.i = (d1).i + 1 where (d1).i > 0;`,
			},
			{
				Statement: `select * from dcomptable;`,
				Results:   []sql.Row{{`(11,)`}, {`(99,)`}, {`(1,3)`}, {`(3,5)`}, {`(0,3)`}, {`(98,101)`}},
			},
			{
				Statement: `explain (verbose, costs off)
  update dcomptable set d1.r = (d1).r - 1, d1.i = (d1).i + 1 where (d1).i > 0;`,
				Results: []sql.Row{{`Update on public.dcomptable`}, {`->  Seq Scan on public.dcomptable`}, {`Output: ROW(((d1).r - '1'::double precision), ((d1).i + '1'::double precision)), ctid`}, {`Filter: ((dcomptable.d1).i > '0'::double precision)`}},
			},
			{
				Statement: `create rule silly as on delete to dcomptable do instead
  update dcomptable set d1.r = (d1).r - 1, d1.i = (d1).i + 1 where (d1).i > 0;`,
			},
			{
				Statement: `\d+ dcomptable
                                  Table "public.dcomptable"
 Column |   Type    | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+-----------+-----------+----------+---------+----------+--------------+-------------
 d1     | dcomptype |           |          |         | extended |              | 
Indexes:
    "dcomptable_d1_key" UNIQUE CONSTRAINT, btree (d1)
Rules:
    silly AS
    ON DELETE TO dcomptable DO INSTEAD  UPDATE dcomptable SET d1.r = (dcomptable.d1).r - 1::double precision, d1.i = (dcomptable.d1).i + 1::double precision
  WHERE (dcomptable.d1).i > 0::double precision
create function makedcomp(r float8, i float8) returns dcomptype
as 'select row(r, i)' language sql;`,
			},
			{
				Statement: `select makedcomp(1,2);`,
				Results:   []sql.Row{{`(1,2)`}},
			},
			{
				Statement:   `select makedcomp(2,1);  -- fail`,
				ErrorString: `value for domain dcomptype violates check constraint "c1"`,
			},
			{
				Statement: `select * from makedcomp(1,2) m;`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `select m, m is not null from makedcomp(1,2) m;`,
				Results:   []sql.Row{{`(1,2)`, true}},
			},
			{
				Statement: `drop function makedcomp(float8, float8);`,
			},
			{
				Statement: `drop table dcomptable;`,
			},
			{
				Statement: `drop type comptype cascade;`,
			},
			{
				Statement: `create type comptype as (r float8, i float8);`,
			},
			{
				Statement: `create domain dcomptype as comptype;`,
			},
			{
				Statement: `alter domain dcomptype add constraint c1 check ((value).r > 0);`,
			},
			{
				Statement: `comment on constraint c1 on domain dcomptype is 'random commentary';`,
			},
			{
				Statement:   `select row(0,1)::dcomptype;  -- fail`,
				ErrorString: `value for domain dcomptype violates check constraint "c1"`,
			},
			{
				Statement:   `alter type comptype alter attribute r type varchar;  -- fail`,
				ErrorString: `operator does not exist: character varying > double precision`,
			},
			{
				Statement: `alter type comptype alter attribute r type bigint;`,
			},
			{
				Statement:   `alter type comptype drop attribute r;  -- fail`,
				ErrorString: `cannot drop column r of composite type comptype because other objects depend on it`,
			},
			{
				Statement: `alter type comptype drop attribute i;`,
			},
			{
				Statement: `select conname, obj_description(oid, 'pg_constraint') from pg_constraint
  where contypid = 'dcomptype'::regtype;  -- check comment is still there`,
				Results: []sql.Row{{`c1`, `random commentary`}},
			},
			{
				Statement: `drop type comptype cascade;`,
			},
			{
				Statement: `create type comptype as (r float8, i float8);`,
			},
			{
				Statement: `create domain dcomptypea as comptype[];`,
			},
			{
				Statement: `create table dcomptable (d1 dcomptypea unique);`,
			},
			{
				Statement: `insert into dcomptable values (array[row(1,2)]::dcomptypea);`,
			},
			{
				Statement: `insert into dcomptable values (array[row(3,4), row(5,6)]::comptype[]);`,
			},
			{
				Statement: `insert into dcomptable values (array[row(7,8)::comptype, row(9,10)::comptype]);`,
			},
			{
				Statement:   `insert into dcomptable values (array[row(1,2)]::dcomptypea);  -- fail on uniqueness`,
				ErrorString: `duplicate key value violates unique constraint "dcomptable_d1_key"`,
			},
			{
				Statement: `insert into dcomptable (d1[1]) values(row(9,10));`,
			},
			{
				Statement: `insert into dcomptable (d1[1].r) values(11);`,
			},
			{
				Statement: `select * from dcomptable;`,
				Results:   []sql.Row{{`{"(1,2)"}`}, {`{"(3,4)","(5,6)"}`}, {`{"(7,8)","(9,10)"}`}, {`{"(9,10)"}`}, {`{"(11,)"}`}},
			},
			{
				Statement: `select d1[2], d1[1].r, d1[1].i from dcomptable;`,
				Results:   []sql.Row{{``, 1, 2}, {`(5,6)`, 3, 4}, {`(9,10)`, 7, 8}, {``, 9, 10}, {``, 11, ``}},
			},
			{
				Statement: `update dcomptable set d1[2] = row(d1[2].i, d1[2].r);`,
			},
			{
				Statement: `select * from dcomptable;`,
				Results:   []sql.Row{{`{"(1,2)","(,)"}`}, {`{"(3,4)","(6,5)"}`}, {`{"(7,8)","(10,9)"}`}, {`{"(9,10)","(,)"}`}, {`{"(11,)","(,)"}`}},
			},
			{
				Statement: `update dcomptable set d1[1].r = d1[1].r + 1 where d1[1].i > 0;`,
			},
			{
				Statement: `select * from dcomptable;`,
				Results:   []sql.Row{{`{"(11,)","(,)"}`}, {`{"(2,2)","(,)"}`}, {`{"(4,4)","(6,5)"}`}, {`{"(8,8)","(10,9)"}`}, {`{"(10,10)","(,)"}`}},
			},
			{
				Statement: `alter domain dcomptypea add constraint c1 check (value[1].r <= value[1].i);`,
			},
			{
				Statement:   `alter domain dcomptypea add constraint c2 check (value[1].r > value[1].i);  -- fail`,
				ErrorString: `column "d1" of table "dcomptable" contains values that violate the new constraint`,
			},
			{
				Statement:   `select array[row(2,1)]::dcomptypea;  -- fail`,
				ErrorString: `value for domain dcomptypea violates check constraint "c1"`,
			},
			{
				Statement: `insert into dcomptable values (array[row(1,2)]::comptype[]);`,
			},
			{
				Statement:   `insert into dcomptable values (array[row(2,1)]::comptype[]);  -- fail`,
				ErrorString: `value for domain dcomptypea violates check constraint "c1"`,
			},
			{
				Statement: `insert into dcomptable (d1[1].r) values(99);`,
			},
			{
				Statement: `insert into dcomptable (d1[1].r, d1[1].i) values(99, 100);`,
			},
			{
				Statement:   `insert into dcomptable (d1[1].r, d1[1].i) values(100, 99);  -- fail`,
				ErrorString: `value for domain dcomptypea violates check constraint "c1"`,
			},
			{
				Statement:   `update dcomptable set d1[1].r = d1[1].r + 1 where d1[1].i > 0;  -- fail`,
				ErrorString: `value for domain dcomptypea violates check constraint "c1"`,
			},
			{
				Statement: `update dcomptable set d1[1].r = d1[1].r - 1, d1[1].i = d1[1].i + 1
  where d1[1].i > 0;`,
			},
			{
				Statement: `select * from dcomptable;`,
				Results:   []sql.Row{{`{"(11,)","(,)"}`}, {`{"(99,)"}`}, {`{"(1,3)","(,)"}`}, {`{"(3,5)","(6,5)"}`}, {`{"(7,9)","(10,9)"}`}, {`{"(9,11)","(,)"}`}, {`{"(0,3)"}`}, {`{"(98,101)"}`}},
			},
			{
				Statement: `explain (verbose, costs off)
  update dcomptable set d1[1].r = d1[1].r - 1, d1[1].i = d1[1].i + 1
    where d1[1].i > 0;`,
				Results: []sql.Row{{`Update on public.dcomptable`}, {`->  Seq Scan on public.dcomptable`}, {`Output: (d1[1].r := (d1[1].r - '1'::double precision))[1].i := (d1[1].i + '1'::double precision), ctid`}, {`Filter: (dcomptable.d1[1].i > '0'::double precision)`}},
			},
			{
				Statement: `create rule silly as on delete to dcomptable do instead
  update dcomptable set d1[1].r = d1[1].r - 1, d1[1].i = d1[1].i + 1
    where d1[1].i > 0;`,
			},
			{
				Statement: `\d+ dcomptable
                                  Table "public.dcomptable"
 Column |    Type    | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------------+-----------+----------+---------+----------+--------------+-------------
 d1     | dcomptypea |           |          |         | extended |              | 
Indexes:
    "dcomptable_d1_key" UNIQUE CONSTRAINT, btree (d1)
Rules:
    silly AS
    ON DELETE TO dcomptable DO INSTEAD  UPDATE dcomptable SET d1[1].r = dcomptable.d1[1].r - 1::double precision, d1[1].i = dcomptable.d1[1].i + 1::double precision
  WHERE dcomptable.d1[1].i > 0::double precision
drop table dcomptable;`,
			},
			{
				Statement: `drop type comptype cascade;`,
			},
			{
				Statement: `create domain posint as int check (value > 0);`,
			},
			{
				Statement: `create table pitable (f1 posint[]);`,
			},
			{
				Statement: `insert into pitable values(array[42]);`,
			},
			{
				Statement:   `insert into pitable values(array[-1]);  -- fail`,
				ErrorString: `value for domain posint violates check constraint "posint_check"`,
			},
			{
				Statement:   `insert into pitable values('{0}');  -- fail`,
				ErrorString: `value for domain posint violates check constraint "posint_check"`,
			},
			{
				Statement: `update pitable set f1[1] = f1[1] + 1;`,
			},
			{
				Statement:   `update pitable set f1[1] = 0;  -- fail`,
				ErrorString: `value for domain posint violates check constraint "posint_check"`,
			},
			{
				Statement: `select * from pitable;`,
				Results:   []sql.Row{{`{43}`}},
			},
			{
				Statement: `drop table pitable;`,
			},
			{
				Statement: `create domain vc4 as varchar(4);`,
			},
			{
				Statement: `create table vc4table (f1 vc4[]);`,
			},
			{
				Statement:   `insert into vc4table values(array['too long']);  -- fail`,
				ErrorString: `value too long for type character varying(4)`,
			},
			{
				Statement: `insert into vc4table values(array['too long']::vc4[]);  -- cast truncates`,
			},
			{
				Statement: `select * from vc4table;`,
				Results:   []sql.Row{{`{"too "}`}},
			},
			{
				Statement: `drop table vc4table;`,
			},
			{
				Statement: `drop type vc4;`,
			},
			{
				Statement: `create domain dposinta as posint[];`,
			},
			{
				Statement: `create table dposintatable (f1 dposinta[]);`,
			},
			{
				Statement:   `insert into dposintatable values(array[array[42]]);  -- fail`,
				ErrorString: `column "f1" is of type dposinta[] but expression is of type integer[]`,
			},
			{
				Statement:   `insert into dposintatable values(array[array[42]::posint[]]); -- still fail`,
				ErrorString: `column "f1" is of type dposinta[] but expression is of type posint[]`,
			},
			{
				Statement: `insert into dposintatable values(array[array[42]::dposinta]); -- but this works`,
			},
			{
				Statement: `select f1, f1[1], (f1[1])[1] from dposintatable;`,
				Results:   []sql.Row{{`{"{42}"}`, `{42}`, 42}},
			},
			{
				Statement: `select pg_typeof(f1) from dposintatable;`,
				Results:   []sql.Row{{`dposinta[]`}},
			},
			{
				Statement: `select pg_typeof(f1[1]) from dposintatable;`,
				Results:   []sql.Row{{`dposinta`}},
			},
			{
				Statement: `select pg_typeof(f1[1][1]) from dposintatable;`,
				Results:   []sql.Row{{`dposinta`}},
			},
			{
				Statement: `select pg_typeof((f1[1])[1]) from dposintatable;`,
				Results:   []sql.Row{{`posint`}},
			},
			{
				Statement: `update dposintatable set f1[2] = array[99];`,
			},
			{
				Statement: `select f1, f1[1], (f1[2])[1] from dposintatable;`,
				Results:   []sql.Row{{`{"{42}","{99}"}`, `{42}`, 99}},
			},
			{
				Statement:   `update dposintatable set f1[2][1] = array[97];`,
				ErrorString: `wrong number of array subscripts`,
			},
			{
				Statement:   `update dposintatable set (f1[2])[1] = array[98];`,
				ErrorString: `syntax error at or near "["`,
			},
			{
				Statement: `drop table dposintatable;`,
			},
			{
				Statement: `drop domain posint cascade;`,
			},
			{
				Statement: `create type comptype as (cf1 int, cf2 int);`,
			},
			{
				Statement: `create domain dcomptype as comptype check ((value).cf1 > 0);`,
			},
			{
				Statement: `create table dcomptable (f1 dcomptype[]);`,
			},
			{
				Statement: `insert into dcomptable values (null);`,
			},
			{
				Statement: `update dcomptable set f1[1].cf2 = 5;`,
			},
			{
				Statement: `table dcomptable;`,
				Results:   []sql.Row{{`{"(,5)"}`}},
			},
			{
				Statement:   `update dcomptable set f1[1].cf1 = -1;  -- fail`,
				ErrorString: `value for domain dcomptype violates check constraint "dcomptype_check"`,
			},
			{
				Statement: `update dcomptable set f1[1].cf1 = 1;`,
			},
			{
				Statement: `table dcomptable;`,
				Results:   []sql.Row{{`{"(1,5)"}`}},
			},
			{
				Statement: `alter domain dcomptype drop constraint dcomptype_check;`,
			},
			{
				Statement: `update dcomptable set f1[1].cf1 = -1;  -- now ok`,
			},
			{
				Statement: `table dcomptable;`,
				Results:   []sql.Row{{`{"(-1,5)"}`}},
			},
			{
				Statement: `drop table dcomptable;`,
			},
			{
				Statement: `drop type comptype cascade;`,
			},
			{
				Statement: `create domain dnotnull varchar(15) NOT NULL;`,
			},
			{
				Statement: `create domain dnull    varchar(15);`,
			},
			{
				Statement: `create domain dcheck   varchar(15) NOT NULL CHECK (VALUE = 'a' OR VALUE = 'c' OR VALUE = 'd');`,
			},
			{
				Statement: `create table nulltest
           ( col1 dnotnull
           , col2 dnotnull NULL  -- NOT NULL in the domain cannot be overridden
           , col3 dnull    NOT NULL
           , col4 dnull
           , col5 dcheck CHECK (col5 IN ('c', 'd'))
           );`,
			},
			{
				Statement:   `INSERT INTO nulltest DEFAULT VALUES;`,
				ErrorString: `domain dnotnull does not allow null values`,
			},
			{
				Statement: `INSERT INTO nulltest values ('a', 'b', 'c', 'd', 'c');  -- Good`,
			},
			{
				Statement:   `insert into nulltest values ('a', 'b', 'c', 'd', NULL);`,
				ErrorString: `domain dcheck does not allow null values`,
			},
			{
				Statement:   `insert into nulltest values ('a', 'b', 'c', 'd', 'a');`,
				ErrorString: `new row for relation "nulltest" violates check constraint "nulltest_col5_check"`,
			},
			{
				Statement:   `INSERT INTO nulltest values (NULL, 'b', 'c', 'd', 'd');`,
				ErrorString: `domain dnotnull does not allow null values`,
			},
			{
				Statement:   `INSERT INTO nulltest values ('a', NULL, 'c', 'd', 'c');`,
				ErrorString: `domain dnotnull does not allow null values`,
			},
			{
				Statement:   `INSERT INTO nulltest values ('a', 'b', NULL, 'd', 'c');`,
				ErrorString: `null value in column "col3" of relation "nulltest" violates not-null constraint`,
			},
			{
				Statement: `INSERT INTO nulltest values ('a', 'b', 'c', NULL, 'd'); -- Good`,
			},
			{
				Statement: `COPY nulltest FROM stdin; --fail
ERROR:  null value in column "col3" of relation "nulltest" violates not-null constraint
CONTEXT:  COPY nulltest, line 1: "a	b	\N	d	d"
COPY nulltest FROM stdin; --fail
ERROR:  domain dcheck does not allow null values
CONTEXT:  COPY nulltest, line 1, column col5: null input
COPY nulltest FROM stdin;`,
				ErrorString: `new row for relation "nulltest" violates check constraint "nulltest_col5_check"`,
			},
			{
				Statement: `CONTEXT:  COPY nulltest, line 3: "a	b	c	\N	a"
select * from nulltest;`,
				Results: []sql.Row{{`a`, `b`, `c`, `d`, `c`}, {`a`, `b`, `c`, ``, `d`}},
			},
			{
				Statement: `SELECT cast('1' as dnotnull);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT cast(NULL as dnotnull); -- fail`,
				ErrorString: `domain dnotnull does not allow null values`,
			},
			{
				Statement:   `SELECT cast(cast(NULL as dnull) as dnotnull); -- fail`,
				ErrorString: `domain dnotnull does not allow null values`,
			},
			{
				Statement:   `SELECT cast(col4 as dnotnull) from nulltest; -- fail`,
				ErrorString: `domain dnotnull does not allow null values`,
			},
			{
				Statement: `drop table nulltest;`,
			},
			{
				Statement: `drop domain dnotnull restrict;`,
			},
			{
				Statement: `drop domain dnull restrict;`,
			},
			{
				Statement: `drop domain dcheck restrict;`,
			},
			{
				Statement: `create domain ddef1 int4 DEFAULT 3;`,
			},
			{
				Statement: `create domain ddef2 oid DEFAULT '12';`,
			},
			{
				Statement: `create domain ddef3 text DEFAULT 5;`,
			},
			{
				Statement: `create sequence ddef4_seq;`,
			},
			{
				Statement: `create domain ddef4 int4 DEFAULT nextval('ddef4_seq');`,
			},
			{
				Statement: `create domain ddef5 numeric(8,2) NOT NULL DEFAULT '12.12';`,
			},
			{
				Statement: `create table defaulttest
            ( col1 ddef1
            , col2 ddef2
            , col3 ddef3
            , col4 ddef4 PRIMARY KEY
            , col5 ddef1 NOT NULL DEFAULT NULL
            , col6 ddef2 DEFAULT '88'
            , col7 ddef4 DEFAULT 8000
            , col8 ddef5
            );`,
			},
			{
				Statement:   `insert into defaulttest(col4) values(0); -- fails, col5 defaults to null`,
				ErrorString: `null value in column "col5" of relation "defaulttest" violates not-null constraint`,
			},
			{
				Statement: `alter table defaulttest alter column col5 drop default;`,
			},
			{
				Statement: `insert into defaulttest default values; -- succeeds, inserts domain default`,
			},
			{
				Statement: `alter table defaulttest alter column col5 set default null;`,
			},
			{
				Statement:   `insert into defaulttest(col4) values(0); -- fails`,
				ErrorString: `null value in column "col5" of relation "defaulttest" violates not-null constraint`,
			},
			{
				Statement: `alter table defaulttest alter column col5 drop default;`,
			},
			{
				Statement: `insert into defaulttest default values;`,
			},
			{
				Statement: `insert into defaulttest default values;`,
			},
			{
				Statement: `COPY defaulttest(col5) FROM stdin;`,
			},
			{
				Statement: `select * from defaulttest;`,
				Results:   []sql.Row{{3, 12, 5, 1, 3, 88, 8000, 12.12}, {3, 12, 5, 2, 3, 88, 8000, 12.12}, {3, 12, 5, 3, 3, 88, 8000, 12.12}, {3, 12, 5, 4, 42, 88, 8000, 12.12}},
			},
			{
				Statement: `drop table defaulttest cascade;`,
			},
			{
				Statement: `create domain dnotnulltest integer;`,
			},
			{
				Statement: `create table domnotnull
( col1 dnotnulltest
, col2 dnotnulltest
);`,
			},
			{
				Statement: `insert into domnotnull default values;`,
			},
			{
				Statement:   `alter domain dnotnulltest set not null; -- fails`,
				ErrorString: `column "col1" of table "domnotnull" contains null values`,
			},
			{
				Statement: `update domnotnull set col1 = 5;`,
			},
			{
				Statement:   `alter domain dnotnulltest set not null; -- fails`,
				ErrorString: `column "col2" of table "domnotnull" contains null values`,
			},
			{
				Statement: `update domnotnull set col2 = 6;`,
			},
			{
				Statement: `alter domain dnotnulltest set not null;`,
			},
			{
				Statement:   `update domnotnull set col1 = null; -- fails`,
				ErrorString: `domain dnotnulltest does not allow null values`,
			},
			{
				Statement: `alter domain dnotnulltest drop not null;`,
			},
			{
				Statement: `update domnotnull set col1 = null;`,
			},
			{
				Statement: `drop domain dnotnulltest cascade;`,
			},
			{
				Statement: `create table domdeftest (col1 ddef1);`,
			},
			{
				Statement: `insert into domdeftest default values;`,
			},
			{
				Statement: `select * from domdeftest;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `alter domain ddef1 set default '42';`,
			},
			{
				Statement: `insert into domdeftest default values;`,
			},
			{
				Statement: `select * from domdeftest;`,
				Results:   []sql.Row{{3}, {42}},
			},
			{
				Statement: `alter domain ddef1 drop default;`,
			},
			{
				Statement: `insert into domdeftest default values;`,
			},
			{
				Statement: `select * from domdeftest;`,
				Results:   []sql.Row{{3}, {42}, {``}},
			},
			{
				Statement: `drop table domdeftest;`,
			},
			{
				Statement: `create domain con as integer;`,
			},
			{
				Statement: `create table domcontest (col1 con);`,
			},
			{
				Statement: `insert into domcontest values (1);`,
			},
			{
				Statement: `insert into domcontest values (2);`,
			},
			{
				Statement:   `alter domain con add constraint t check (VALUE < 1); -- fails`,
				ErrorString: `column "col1" of table "domcontest" contains values that violate the new constraint`,
			},
			{
				Statement: `alter domain con add constraint t check (VALUE < 34);`,
			},
			{
				Statement: `alter domain con add check (VALUE > 0);`,
			},
			{
				Statement:   `insert into domcontest values (-5); -- fails`,
				ErrorString: `value for domain con violates check constraint "con_check"`,
			},
			{
				Statement:   `insert into domcontest values (42); -- fails`,
				ErrorString: `value for domain con violates check constraint "t"`,
			},
			{
				Statement: `insert into domcontest values (5);`,
			},
			{
				Statement: `alter domain con drop constraint t;`,
			},
			{
				Statement: `insert into domcontest values (-5); --fails
ERROR:  value for domain con violates check constraint "con_check"
insert into domcontest values (42);`,
			},
			{
				Statement:   `alter domain con drop constraint nonexistent;`,
				ErrorString: `constraint "nonexistent" of domain "con" does not exist`,
			},
			{
				Statement: `alter domain con drop constraint if exists nonexistent;`,
			},
			{
				Statement: `create domain things AS INT;`,
			},
			{
				Statement: `CREATE TABLE thethings (stuff things);`,
			},
			{
				Statement: `INSERT INTO thethings (stuff) VALUES (55);`,
			},
			{
				Statement:   `ALTER DOMAIN things ADD CONSTRAINT meow CHECK (VALUE < 11);`,
				ErrorString: `column "stuff" of table "thethings" contains values that violate the new constraint`,
			},
			{
				Statement: `ALTER DOMAIN things ADD CONSTRAINT meow CHECK (VALUE < 11) NOT VALID;`,
			},
			{
				Statement:   `ALTER DOMAIN things VALIDATE CONSTRAINT meow;`,
				ErrorString: `column "stuff" of table "thethings" contains values that violate the new constraint`,
			},
			{
				Statement: `UPDATE thethings SET stuff = 10;`,
			},
			{
				Statement: `ALTER DOMAIN things VALIDATE CONSTRAINT meow;`,
			},
			{
				Statement: `create table domtab (col1 integer);`,
			},
			{
				Statement: `create domain dom as integer;`,
			},
			{
				Statement: `create view domview as select cast(col1 as dom) from domtab;`,
			},
			{
				Statement: `insert into domtab (col1) values (null);`,
			},
			{
				Statement: `insert into domtab (col1) values (5);`,
			},
			{
				Statement: `select * from domview;`,
				Results:   []sql.Row{{``}, {5}},
			},
			{
				Statement: `alter domain dom set not null;`,
			},
			{
				Statement:   `select * from domview; -- fail`,
				ErrorString: `domain dom does not allow null values`,
			},
			{
				Statement: `alter domain dom drop not null;`,
			},
			{
				Statement: `select * from domview;`,
				Results:   []sql.Row{{``}, {5}},
			},
			{
				Statement: `alter domain dom add constraint domchkgt6 check(value > 6);`,
			},
			{
				Statement: `select * from domview; --fail
ERROR:  value for domain dom violates check constraint "domchkgt6"
alter domain dom drop constraint domchkgt6 restrict;`,
			},
			{
				Statement: `select * from domview;`,
				Results:   []sql.Row{{``}, {5}},
			},
			{
				Statement: `drop domain ddef1 restrict;`,
			},
			{
				Statement: `drop domain ddef2 restrict;`,
			},
			{
				Statement: `drop domain ddef3 restrict;`,
			},
			{
				Statement: `drop domain ddef4 restrict;`,
			},
			{
				Statement: `drop domain ddef5 restrict;`,
			},
			{
				Statement: `drop sequence ddef4_seq;`,
			},
			{
				Statement: `create domain vchar4 varchar(4);`,
			},
			{
				Statement: `create domain dinter vchar4 check (substring(VALUE, 1, 1) = 'x');`,
			},
			{
				Statement: `create domain dtop dinter check (substring(VALUE, 2, 1) = '1');`,
			},
			{
				Statement: `select 'x123'::dtop;`,
				Results:   []sql.Row{{`x123`}},
			},
			{
				Statement: `select 'x1234'::dtop; -- explicit coercion should truncate`,
				Results:   []sql.Row{{`x123`}},
			},
			{
				Statement:   `select 'y1234'::dtop; -- fail`,
				ErrorString: `value for domain dtop violates check constraint "dinter_check"`,
			},
			{
				Statement:   `select 'y123'::dtop; -- fail`,
				ErrorString: `value for domain dtop violates check constraint "dinter_check"`,
			},
			{
				Statement:   `select 'yz23'::dtop; -- fail`,
				ErrorString: `value for domain dtop violates check constraint "dinter_check"`,
			},
			{
				Statement:   `select 'xz23'::dtop; -- fail`,
				ErrorString: `value for domain dtop violates check constraint "dtop_check"`,
			},
			{
				Statement: `create temp table dtest(f1 dtop);`,
			},
			{
				Statement: `insert into dtest values('x123');`,
			},
			{
				Statement:   `insert into dtest values('x1234'); -- fail, implicit coercion`,
				ErrorString: `value too long for type character varying(4)`,
			},
			{
				Statement:   `insert into dtest values('y1234'); -- fail, implicit coercion`,
				ErrorString: `value too long for type character varying(4)`,
			},
			{
				Statement:   `insert into dtest values('y123'); -- fail`,
				ErrorString: `value for domain dtop violates check constraint "dinter_check"`,
			},
			{
				Statement:   `insert into dtest values('yz23'); -- fail`,
				ErrorString: `value for domain dtop violates check constraint "dinter_check"`,
			},
			{
				Statement:   `insert into dtest values('xz23'); -- fail`,
				ErrorString: `value for domain dtop violates check constraint "dtop_check"`,
			},
			{
				Statement: `drop table dtest;`,
			},
			{
				Statement: `drop domain vchar4 cascade;`,
			},
			{
				Statement: `create domain str_domain as text not null;`,
			},
			{
				Statement: `create table domain_test (a int, b int);`,
			},
			{
				Statement: `insert into domain_test values (1, 2);`,
			},
			{
				Statement: `insert into domain_test values (1, 2);`,
			},
			{
				Statement:   `alter table domain_test add column c str_domain;`,
				ErrorString: `domain str_domain does not allow null values`,
			},
			{
				Statement: `create domain str_domain2 as text check (value <> 'foo') default 'foo';`,
			},
			{
				Statement:   `alter table domain_test add column d str_domain2;`,
				ErrorString: `value for domain str_domain2 violates check constraint "str_domain2_check"`,
			},
			{
				Statement: `create domain pos_int as int4 check (value > 0) not null;`,
			},
			{
				Statement: `prepare s1 as select $1::pos_int = 10 as "is_ten";`,
			},
			{
				Statement: `execute s1(10);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `execute s1(0); -- should fail`,
				ErrorString: `value for domain pos_int violates check constraint "pos_int_check"`,
			},
			{
				Statement:   `execute s1(NULL); -- should fail`,
				ErrorString: `domain pos_int does not allow null values`,
			},
			{
				Statement: `create function doubledecrement(p1 pos_int) returns pos_int as $$
declare v pos_int;`,
			},
			{
				Statement: `begin
    return p1;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select doubledecrement(3); -- fail because of implicit null assignment`,
				ErrorString: `domain pos_int does not allow null values`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function doubledecrement(pos_int) line 2 during statement block local variable initialization
create or replace function doubledecrement(p1 pos_int) returns pos_int as $$
declare v pos_int := 0;`,
			},
			{
				Statement: `begin
    return p1;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select doubledecrement(3); -- fail at initialization assignment`,
				ErrorString: `value for domain pos_int violates check constraint "pos_int_check"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function doubledecrement(pos_int) line 2 during statement block local variable initialization
create or replace function doubledecrement(p1 pos_int) returns pos_int as $$
declare v pos_int := 1;`,
			},
			{
				Statement: `begin
    v := p1 - 1;`,
			},
			{
				Statement: `    return v - 1;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select doubledecrement(null); -- fail before call`,
				ErrorString: `domain pos_int does not allow null values`,
			},
			{
				Statement:   `select doubledecrement(0); -- fail before call`,
				ErrorString: `value for domain pos_int violates check constraint "pos_int_check"`,
			},
			{
				Statement:   `select doubledecrement(1); -- fail at assignment to v`,
				ErrorString: `value for domain pos_int violates check constraint "pos_int_check"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function doubledecrement(pos_int) line 4 at assignment
select doubledecrement(2); -- fail at return`,
				ErrorString: `value for domain pos_int violates check constraint "pos_int_check"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function doubledecrement(pos_int) while casting return value to function's return type
select doubledecrement(3); -- good`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `create domain posint as int4;`,
			},
			{
				Statement: `create type ddtest1 as (f1 posint);`,
			},
			{
				Statement: `create table ddtest2(f1 ddtest1);`,
			},
			{
				Statement: `insert into ddtest2 values(row(-1));`,
			},
			{
				Statement:   `alter domain posint add constraint c1 check(value >= 0);`,
				ErrorString: `cannot alter type "posint" because column "ddtest2.f1" uses it`,
			},
			{
				Statement: `drop table ddtest2;`,
			},
			{
				Statement: `create table ddtest2(f1 ddtest1[]);`,
			},
			{
				Statement: `insert into ddtest2 values('{(-1)}');`,
			},
			{
				Statement:   `alter domain posint add constraint c1 check(value >= 0);`,
				ErrorString: `cannot alter type "posint" because column "ddtest2.f1" uses it`,
			},
			{
				Statement: `drop table ddtest2;`,
			},
			{
				Statement: `create domain ddtest1d as ddtest1;`,
			},
			{
				Statement: `create table ddtest2(f1 ddtest1d);`,
			},
			{
				Statement: `insert into ddtest2 values('(-1)');`,
			},
			{
				Statement:   `alter domain posint add constraint c1 check(value >= 0);`,
				ErrorString: `cannot alter type "posint" because column "ddtest2.f1" uses it`,
			},
			{
				Statement: `drop table ddtest2;`,
			},
			{
				Statement: `drop domain ddtest1d;`,
			},
			{
				Statement: `create domain ddtest1d as ddtest1[];`,
			},
			{
				Statement: `create table ddtest2(f1 ddtest1d);`,
			},
			{
				Statement: `insert into ddtest2 values('{(-1)}');`,
			},
			{
				Statement:   `alter domain posint add constraint c1 check(value >= 0);`,
				ErrorString: `cannot alter type "posint" because column "ddtest2.f1" uses it`,
			},
			{
				Statement: `drop table ddtest2;`,
			},
			{
				Statement: `drop domain ddtest1d;`,
			},
			{
				Statement: `create type rposint as range (subtype = posint);`,
			},
			{
				Statement: `create table ddtest2(f1 rposint);`,
			},
			{
				Statement: `insert into ddtest2 values('(-1,3]');`,
			},
			{
				Statement:   `alter domain posint add constraint c1 check(value >= 0);`,
				ErrorString: `cannot alter type "posint" because column "ddtest2.f1" uses it`,
			},
			{
				Statement: `drop table ddtest2;`,
			},
			{
				Statement: `drop type rposint;`,
			},
			{
				Statement: `alter domain posint add constraint c1 check(value >= 0);`,
			},
			{
				Statement: `create domain posint2 as posint check (value % 2 = 0);`,
			},
			{
				Statement: `create table ddtest2(f1 posint2);`,
			},
			{
				Statement:   `insert into ddtest2 values(11); -- fail`,
				ErrorString: `value for domain posint2 violates check constraint "posint2_check"`,
			},
			{
				Statement:   `insert into ddtest2 values(-2); -- fail`,
				ErrorString: `value for domain posint2 violates check constraint "c1"`,
			},
			{
				Statement: `insert into ddtest2 values(2);`,
			},
			{
				Statement:   `alter domain posint add constraint c2 check(value >= 10); -- fail`,
				ErrorString: `column "f1" of table "ddtest2" contains values that violate the new constraint`,
			},
			{
				Statement: `alter domain posint add constraint c2 check(value > 0); -- OK`,
			},
			{
				Statement: `drop table ddtest2;`,
			},
			{
				Statement: `drop type ddtest1;`,
			},
			{
				Statement: `drop domain posint cascade;`,
			},
			{
				Statement: `create or replace function array_elem_check(numeric) returns numeric as $$
declare
  x numeric(4,2)[1];`,
			},
			{
				Statement: `begin
  x[1] := $1;`,
			},
			{
				Statement: `  return x[1];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select array_elem_check(121.00);`,
				ErrorString: `numeric field overflow`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function array_elem_check(numeric) line 5 at assignment
select array_elem_check(1.23456);`,
				Results: []sql.Row{{1.23}},
			},
			{
				Statement: `create domain mynums as numeric(4,2)[1];`,
			},
			{
				Statement: `create or replace function array_elem_check(numeric) returns numeric as $$
declare
  x mynums;`,
			},
			{
				Statement: `begin
  x[1] := $1;`,
			},
			{
				Statement: `  return x[1];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select array_elem_check(121.00);`,
				ErrorString: `numeric field overflow`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function array_elem_check(numeric) line 5 at assignment
select array_elem_check(1.23456);`,
				Results: []sql.Row{{1.23}},
			},
			{
				Statement: `create domain mynums2 as mynums;`,
			},
			{
				Statement: `create or replace function array_elem_check(numeric) returns numeric as $$
declare
  x mynums2;`,
			},
			{
				Statement: `begin
  x[1] := $1;`,
			},
			{
				Statement: `  return x[1];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement:   `select array_elem_check(121.00);`,
				ErrorString: `numeric field overflow`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function array_elem_check(numeric) line 5 at assignment
select array_elem_check(1.23456);`,
				Results: []sql.Row{{1.23}},
			},
			{
				Statement: `drop function array_elem_check(numeric);`,
			},
			{
				Statement: `create domain orderedpair as int[2] check (value[1] < value[2]);`,
			},
			{
				Statement: `select array[1,2]::orderedpair;`,
				Results:   []sql.Row{{`{1,2}`}},
			},
			{
				Statement:   `select array[2,1]::orderedpair;  -- fail`,
				ErrorString: `value for domain orderedpair violates check constraint "orderedpair_check"`,
			},
			{
				Statement: `create temp table op (f1 orderedpair);`,
			},
			{
				Statement: `insert into op values (array[1,2]);`,
			},
			{
				Statement:   `insert into op values (array[2,1]);  -- fail`,
				ErrorString: `value for domain orderedpair violates check constraint "orderedpair_check"`,
			},
			{
				Statement: `update op set f1[2] = 3;`,
			},
			{
				Statement:   `update op set f1[2] = 0;  -- fail`,
				ErrorString: `value for domain orderedpair violates check constraint "orderedpair_check"`,
			},
			{
				Statement: `select * from op;`,
				Results:   []sql.Row{{`{1,3}`}},
			},
			{
				Statement: `create or replace function array_elem_check(int) returns int as $$
declare
  x orderedpair := '{1,2}';`,
			},
			{
				Statement: `begin
  x[2] := $1;`,
			},
			{
				Statement: `  return x[2];`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `select array_elem_check(3);`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement:   `select array_elem_check(-1);`,
				ErrorString: `value for domain orderedpair violates check constraint "orderedpair_check"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function array_elem_check(integer) line 5 at assignment
drop function array_elem_check(int);`,
			},
			{
				Statement: `create domain di as int;`,
			},
			{
				Statement: `create function dom_check(int) returns di as $$
declare d di;`,
			},
			{
				Statement: `begin
  d := $1::di;`,
			},
			{
				Statement: `  return d;`,
			},
			{
				Statement: `end
$$ language plpgsql immutable;`,
			},
			{
				Statement: `select dom_check(0);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `alter domain di add constraint pos check (value > 0);`,
			},
			{
				Statement:   `select dom_check(0); -- fail`,
				ErrorString: `value for domain di violates check constraint "pos"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function dom_check(integer) line 4 at assignment
alter domain di drop constraint pos;`,
			},
			{
				Statement: `select dom_check(0);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `create or replace function dom_check(int) returns di as $$
declare d di;`,
			},
			{
				Statement: `begin
  d := $1;`,
			},
			{
				Statement: `  return d;`,
			},
			{
				Statement: `end
$$ language plpgsql immutable;`,
			},
			{
				Statement: `select dom_check(0);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `alter domain di add constraint pos check (value > 0);`,
			},
			{
				Statement:   `select dom_check(0); -- fail`,
				ErrorString: `value for domain di violates check constraint "pos"`,
			},
			{
				Statement: `CONTEXT:  PL/pgSQL function dom_check(integer) line 4 at assignment
alter domain di drop constraint pos;`,
			},
			{
				Statement: `select dom_check(0);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `drop function dom_check(int);`,
			},
			{
				Statement: `drop domain di;`,
			},
			{
				Statement: `create function sql_is_distinct_from(anyelement, anyelement)
returns boolean language sql
as 'select $1 is distinct from $2 limit 1';`,
			},
			{
				Statement: `create domain inotnull int
  check (sql_is_distinct_from(value, null));`,
			},
			{
				Statement: `select 1::inotnull;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `select null::inotnull;`,
				ErrorString: `value for domain inotnull violates check constraint "inotnull_check"`,
			},
			{
				Statement: `create table dom_table (x inotnull);`,
			},
			{
				Statement: `insert into dom_table values ('1');`,
			},
			{
				Statement: `insert into dom_table values (1);`,
			},
			{
				Statement:   `insert into dom_table values (null);`,
				ErrorString: `value for domain inotnull violates check constraint "inotnull_check"`,
			},
			{
				Statement: `drop table dom_table;`,
			},
			{
				Statement: `drop domain inotnull;`,
			},
			{
				Statement: `drop function sql_is_distinct_from(anyelement, anyelement);`,
			},
			{
				Statement: `create domain testdomain1 as int;`,
			},
			{
				Statement: `alter domain testdomain1 rename to testdomain2;`,
			},
			{
				Statement: `alter type testdomain2 rename to testdomain3;  -- alter type also works`,
			},
			{
				Statement: `drop domain testdomain3;`,
			},
			{
				Statement: `create domain testdomain1 as int constraint unsigned check (value > 0);`,
			},
			{
				Statement: `alter domain testdomain1 rename constraint unsigned to unsigned_foo;`,
			},
			{
				Statement: `alter domain testdomain1 drop constraint unsigned_foo;`,
			},
			{
				Statement: `drop domain testdomain1;`,
			},
		},
	})
}
