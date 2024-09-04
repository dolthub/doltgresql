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

func TestInsert(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_insert)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_insert,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create table inserttest (col1 int4, col2 int4 NOT NULL, col3 text default 'testing');`,
			},
			{
				Statement:   `insert into inserttest (col1, col2, col3) values (DEFAULT, DEFAULT, DEFAULT);`,
				ErrorString: `null value in column "col2" of relation "inserttest" violates not-null constraint`,
			},
			{
				Statement: `insert into inserttest (col2, col3) values (3, DEFAULT);`,
			},
			{
				Statement: `insert into inserttest (col1, col2, col3) values (DEFAULT, 5, DEFAULT);`,
			},
			{
				Statement: `insert into inserttest values (DEFAULT, 5, 'test');`,
			},
			{
				Statement: `insert into inserttest values (DEFAULT, 7);`,
			},
			{
				Statement: `select * from inserttest;`,
				Results:   []sql.Row{{``, 3, `testing`}, {``, 5, `testing`}, {``, 5, `test`}, {``, 7, `testing`}},
			},
			{
				Statement:   `insert into inserttest (col1, col2, col3) values (DEFAULT, DEFAULT);`,
				ErrorString: `INSERT has more target columns than expressions`,
			},
			{
				Statement:   `insert into inserttest (col1, col2, col3) values (1, 2);`,
				ErrorString: `INSERT has more target columns than expressions`,
			},
			{
				Statement:   `insert into inserttest (col1) values (1, 2);`,
				ErrorString: `INSERT has more expressions than target columns`,
			},
			{
				Statement:   `insert into inserttest (col1) values (DEFAULT, DEFAULT);`,
				ErrorString: `INSERT has more expressions than target columns`,
			},
			{
				Statement: `select * from inserttest;`,
				Results:   []sql.Row{{``, 3, `testing`}, {``, 5, `testing`}, {``, 5, `test`}, {``, 7, `testing`}},
			},
			{
				Statement: `insert into inserttest values(10, 20, '40'), (-1, 2, DEFAULT),
    ((select 2), (select i from (values(3)) as foo (i)), 'values are fun!');`,
			},
			{
				Statement: `select * from inserttest;`,
				Results:   []sql.Row{{``, 3, `testing`}, {``, 5, `testing`}, {``, 5, `test`}, {``, 7, `testing`}, {10, 20, 40}, {-1, 2, `testing`}, {2, 3, `values are fun!`}},
			},
			{
				Statement: `insert into inserttest values(30, 50, repeat('x', 10000));`,
			},
			{
				Statement: `select col1, col2, char_length(col3) from inserttest;`,
				Results:   []sql.Row{{``, 3, 7}, {``, 5, 7}, {``, 5, 4}, {``, 7, 7}, {10, 20, 2}, {-1, 2, 7}, {2, 3, 15}, {30, 50, 10000}},
			},
			{
				Statement: `drop table inserttest;`,
			},
			{
				Statement: `CREATE TABLE large_tuple_test (a int, b text) WITH (fillfactor = 10);`,
			},
			{
				Statement: `ALTER TABLE large_tuple_test ALTER COLUMN b SET STORAGE plain;`,
			},
			{
				Statement: `INSERT INTO large_tuple_test (select 1, NULL);`,
			},
			{
				Statement: `INSERT INTO large_tuple_test (select 2, repeat('a', 1000));`,
			},
			{
				Statement: `SELECT pg_size_pretty(pg_relation_size('large_tuple_test'::regclass, 'main'));`,
				Results:   []sql.Row{{`8192 bytes`}},
			},
			{
				Statement: `INSERT INTO large_tuple_test (select 3, NULL);`,
			},
			{
				Statement: `INSERT INTO large_tuple_test (select 4, repeat('a', 8126));`,
			},
			{
				Statement: `DROP TABLE large_tuple_test;`,
			},
			{
				Statement: `create type insert_test_type as (if1 int, if2 text[]);`,
			},
			{
				Statement: `create table inserttest (f1 int, f2 int[],
                         f3 insert_test_type, f4 insert_test_type[]);`,
			},
			{
				Statement: `insert into inserttest (f2[1], f2[2]) values (1,2);`,
			},
			{
				Statement: `insert into inserttest (f2[1], f2[2]) values (3,4), (5,6);`,
			},
			{
				Statement: `insert into inserttest (f2[1], f2[2]) select 7,8;`,
			},
			{
				Statement:   `insert into inserttest (f2[1], f2[2]) values (1,default);  -- not supported`,
				ErrorString: `cannot set an array element to DEFAULT`,
			},
			{
				Statement: `insert into inserttest (f3.if1, f3.if2) values (1,array['foo']);`,
			},
			{
				Statement: `insert into inserttest (f3.if1, f3.if2) values (1,'{foo}'), (2,'{bar}');`,
			},
			{
				Statement: `insert into inserttest (f3.if1, f3.if2) select 3, '{baz,quux}';`,
			},
			{
				Statement:   `insert into inserttest (f3.if1, f3.if2) values (1,default);  -- not supported`,
				ErrorString: `cannot set a subfield to DEFAULT`,
			},
			{
				Statement: `insert into inserttest (f3.if2[1], f3.if2[2]) values ('foo', 'bar');`,
			},
			{
				Statement: `insert into inserttest (f3.if2[1], f3.if2[2]) values ('foo', 'bar'), ('baz', 'quux');`,
			},
			{
				Statement: `insert into inserttest (f3.if2[1], f3.if2[2]) select 'bear', 'beer';`,
			},
			{
				Statement: `insert into inserttest (f4[1].if2[1], f4[1].if2[2]) values ('foo', 'bar');`,
			},
			{
				Statement: `insert into inserttest (f4[1].if2[1], f4[1].if2[2]) values ('foo', 'bar'), ('baz', 'quux');`,
			},
			{
				Statement: `insert into inserttest (f4[1].if2[1], f4[1].if2[2]) select 'bear', 'beer';`,
			},
			{
				Statement: `select * from inserttest;`,
				Results:   []sql.Row{{``, `{1,2}`, ``, ``}, {``, `{3,4}`, ``, ``}, {``, `{5,6}`, ``, ``}, {``, `{7,8}`, ``, ``}, {``, ``, `(1,{foo})`, ``}, {``, ``, `(1,{foo})`, ``}, {``, ``, `(2,{bar})`, ``}, {``, ``, `(3,"{baz,quux}")`, ``}, {``, ``, `(,"{foo,bar}")`, ``}, {``, ``, `(,"{foo,bar}")`, ``}, {``, ``, `(,"{baz,quux}")`, ``}, {``, ``, `(,"{bear,beer}")`, ``}, {``, ``, ``, `{"(,\"{foo,bar}\")"}`}, {``, ``, ``, `{"(,\"{foo,bar}\")"}`}, {``, ``, ``, `{"(,\"{baz,quux}\")"}`}, {``, ``, ``, `{"(,\"{bear,beer}\")"}`}},
			},
			{
				Statement: `create table inserttest2 (f1 bigint, f2 text);`,
			},
			{
				Statement: `create rule irule1 as on insert to inserttest2 do also
  insert into inserttest (f3.if2[1], f3.if2[2])
  values (new.f1,new.f2);`,
			},
			{
				Statement: `create rule irule2 as on insert to inserttest2 do also
  insert into inserttest (f4[1].if1, f4[1].if2[2])
  values (1,'fool'),(new.f1,new.f2);`,
			},
			{
				Statement: `create rule irule3 as on insert to inserttest2 do also
  insert into inserttest (f4[1].if1, f4[1].if2[2])
  select new.f1, new.f2;`,
			},
			{
				Statement: `\d+ inserttest2
                                Table "public.inserttest2"
 Column |  Type  | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+--------+-----------+----------+---------+----------+--------------+-------------
 f1     | bigint |           |          |         | plain    |              | 
 f2     | text   |           |          |         | extended |              | 
Rules:
    irule1 AS
    ON INSERT TO inserttest2 DO  INSERT INTO inserttest (f3.if2[1], f3.if2[2])
  VALUES (new.f1, new.f2)
    irule2 AS
    ON INSERT TO inserttest2 DO  INSERT INTO inserttest (f4[1].if1, f4[1].if2[2]) VALUES (1,'fool'::text), (new.f1,new.f2)
    irule3 AS
    ON INSERT TO inserttest2 DO  INSERT INTO inserttest (f4[1].if1, f4[1].if2[2])  SELECT new.f1,
            new.f2
drop table inserttest2;`,
			},
			{
				Statement: `drop table inserttest;`,
			},
			{
				Statement: `drop type insert_test_type;`,
			},
			{
				Statement: `create table range_parted (
	a text,
	b int
) partition by range (a, (b+0));`,
			},
			{
				Statement:   `insert into range_parted values ('a', 11);`,
				ErrorString: `no partition of relation "range_parted" found for row`,
			},
			{
				Statement: `create table part1 partition of range_parted for values from ('a', 1) to ('a', 10);`,
			},
			{
				Statement: `create table part2 partition of range_parted for values from ('a', 10) to ('a', 20);`,
			},
			{
				Statement: `create table part3 partition of range_parted for values from ('b', 1) to ('b', 10);`,
			},
			{
				Statement: `create table part4 partition of range_parted for values from ('b', 10) to ('b', 20);`,
			},
			{
				Statement:   `insert into part1 values ('a', 11);`,
				ErrorString: `new row for relation "part1" violates partition constraint`,
			},
			{
				Statement:   `insert into part1 values ('b', 1);`,
				ErrorString: `new row for relation "part1" violates partition constraint`,
			},
			{
				Statement: `insert into part1 values ('a', 1);`,
			},
			{
				Statement:   `insert into part4 values ('b', 21);`,
				ErrorString: `new row for relation "part4" violates partition constraint`,
			},
			{
				Statement:   `insert into part4 values ('a', 10);`,
				ErrorString: `new row for relation "part4" violates partition constraint`,
			},
			{
				Statement: `insert into part4 values ('b', 10);`,
			},
			{
				Statement:   `insert into part1 values (null);`,
				ErrorString: `new row for relation "part1" violates partition constraint`,
			},
			{
				Statement:   `insert into part1 values (1);`,
				ErrorString: `new row for relation "part1" violates partition constraint`,
			},
			{
				Statement: `create table list_parted (
	a text,
	b int
) partition by list (lower(a));`,
			},
			{
				Statement: `create table part_aa_bb partition of list_parted FOR VALUES IN ('aa', 'bb');`,
			},
			{
				Statement: `create table part_cc_dd partition of list_parted FOR VALUES IN ('cc', 'dd');`,
			},
			{
				Statement: `create table part_null partition of list_parted FOR VALUES IN (null);`,
			},
			{
				Statement:   `insert into part_aa_bb values ('cc', 1);`,
				ErrorString: `new row for relation "part_aa_bb" violates partition constraint`,
			},
			{
				Statement:   `insert into part_aa_bb values ('AAa', 1);`,
				ErrorString: `new row for relation "part_aa_bb" violates partition constraint`,
			},
			{
				Statement:   `insert into part_aa_bb values (null);`,
				ErrorString: `new row for relation "part_aa_bb" violates partition constraint`,
			},
			{
				Statement: `insert into part_cc_dd values ('cC', 1);`,
			},
			{
				Statement: `insert into part_null values (null, 0);`,
			},
			{
				Statement: `create table part_ee_ff partition of list_parted for values in ('ee', 'ff') partition by range (b);`,
			},
			{
				Statement: `create table part_ee_ff1 partition of part_ee_ff for values from (1) to (10);`,
			},
			{
				Statement: `create table part_ee_ff2 partition of part_ee_ff for values from (10) to (20);`,
			},
			{
				Statement: `create table part_default partition of list_parted default;`,
			},
			{
				Statement:   `insert into part_default values ('aa', 2);`,
				ErrorString: `new row for relation "part_default" violates partition constraint`,
			},
			{
				Statement:   `insert into part_default values (null, 2);`,
				ErrorString: `new row for relation "part_default" violates partition constraint`,
			},
			{
				Statement: `insert into part_default values ('Zz', 2);`,
			},
			{
				Statement: `drop table part_default;`,
			},
			{
				Statement: `create table part_xx_yy partition of list_parted for values in ('xx', 'yy') partition by list (a);`,
			},
			{
				Statement: `create table part_xx_yy_p1 partition of part_xx_yy for values in ('xx');`,
			},
			{
				Statement: `create table part_xx_yy_defpart partition of part_xx_yy default;`,
			},
			{
				Statement: `create table part_default partition of list_parted default partition by range(b);`,
			},
			{
				Statement: `create table part_default_p1 partition of part_default for values from (20) to (30);`,
			},
			{
				Statement: `create table part_default_p2 partition of part_default for values from (30) to (40);`,
			},
			{
				Statement:   `insert into part_ee_ff1 values ('EE', 11);`,
				ErrorString: `new row for relation "part_ee_ff1" violates partition constraint`,
			},
			{
				Statement:   `insert into part_default_p2 values ('gg', 43);`,
				ErrorString: `new row for relation "part_default_p2" violates partition constraint`,
			},
			{
				Statement:   `insert into part_ee_ff1 values ('cc', 1);`,
				ErrorString: `new row for relation "part_ee_ff1" violates partition constraint`,
			},
			{
				Statement:   `insert into part_default values ('gg', 43);`,
				ErrorString: `no partition of relation "part_default" found for row`,
			},
			{
				Statement: `insert into part_ee_ff1 values ('ff', 1);`,
			},
			{
				Statement: `insert into part_ee_ff2 values ('ff', 11);`,
			},
			{
				Statement: `insert into part_default_p1 values ('cd', 25);`,
			},
			{
				Statement: `insert into part_default_p2 values ('de', 35);`,
			},
			{
				Statement: `insert into list_parted values ('ab', 21);`,
			},
			{
				Statement: `insert into list_parted values ('xx', 1);`,
			},
			{
				Statement: `insert into list_parted values ('yy', 2);`,
			},
			{
				Statement: `select tableoid::regclass, * from list_parted;`,
				Results:   []sql.Row{{`part_cc_dd`, `cC`, 1}, {`part_ee_ff1`, `ff`, 1}, {`part_ee_ff2`, `ff`, 11}, {`part_xx_yy_p1`, `xx`, 1}, {`part_xx_yy_defpart`, `yy`, 2}, {`part_null`, ``, 0}, {`part_default_p1`, `cd`, 25}, {`part_default_p1`, `ab`, 21}, {`part_default_p2`, `de`, 35}},
			},
			{
				Statement:   `insert into range_parted values ('a', 0);`,
				ErrorString: `no partition of relation "range_parted" found for row`,
			},
			{
				Statement: `insert into range_parted values ('a', 1);`,
			},
			{
				Statement: `insert into range_parted values ('a', 10);`,
			},
			{
				Statement:   `insert into range_parted values ('a', 20);`,
				ErrorString: `no partition of relation "range_parted" found for row`,
			},
			{
				Statement: `insert into range_parted values ('b', 1);`,
			},
			{
				Statement: `insert into range_parted values ('b', 10);`,
			},
			{
				Statement:   `insert into range_parted values ('a');`,
				ErrorString: `no partition of relation "range_parted" found for row`,
			},
			{
				Statement: `create table part_def partition of range_parted default;`,
			},
			{
				Statement:   `insert into part_def values ('b', 10);`,
				ErrorString: `new row for relation "part_def" violates partition constraint`,
			},
			{
				Statement: `insert into part_def values ('c', 10);`,
			},
			{
				Statement: `insert into range_parted values (null, null);`,
			},
			{
				Statement: `insert into range_parted values ('a', null);`,
			},
			{
				Statement: `insert into range_parted values (null, 19);`,
			},
			{
				Statement: `insert into range_parted values ('b', 20);`,
			},
			{
				Statement: `select tableoid::regclass, * from range_parted;`,
				Results:   []sql.Row{{`part1`, `a`, 1}, {`part1`, `a`, 1}, {`part2`, `a`, 10}, {`part3`, `b`, 1}, {`part4`, `b`, 10}, {`part4`, `b`, 10}, {`part_def`, `c`, 10}, {`part_def`, ``, ``}, {`part_def`, `a`, ``}, {`part_def`, ``, 19}, {`part_def`, `b`, 20}},
			},
			{
				Statement: `insert into list_parted values (null, 1);`,
			},
			{
				Statement: `insert into list_parted (a) values ('aA');`,
			},
			{
				Statement:   `insert into list_parted values ('EE', 0);`,
				ErrorString: `no partition of relation "part_ee_ff" found for row`,
			},
			{
				Statement:   `insert into part_ee_ff values ('EE', 0);`,
				ErrorString: `no partition of relation "part_ee_ff" found for row`,
			},
			{
				Statement: `insert into list_parted values ('EE', 1);`,
			},
			{
				Statement: `insert into part_ee_ff values ('EE', 10);`,
			},
			{
				Statement: `select tableoid::regclass, * from list_parted;`,
				Results:   []sql.Row{{`part_aa_bb`, `aA`, ``}, {`part_cc_dd`, `cC`, 1}, {`part_ee_ff1`, `ff`, 1}, {`part_ee_ff1`, `EE`, 1}, {`part_ee_ff2`, `ff`, 11}, {`part_ee_ff2`, `EE`, 10}, {`part_xx_yy_p1`, `xx`, 1}, {`part_xx_yy_defpart`, `yy`, 2}, {`part_null`, ``, 0}, {`part_null`, ``, 1}, {`part_default_p1`, `cd`, 25}, {`part_default_p1`, `ab`, 21}, {`part_default_p2`, `de`, 35}},
			},
			{
				Statement: `create table part_gg partition of list_parted for values in ('gg') partition by range (b);`,
			},
			{
				Statement: `create table part_gg1 partition of part_gg for values from (minvalue) to (1);`,
			},
			{
				Statement: `create table part_gg2 partition of part_gg for values from (1) to (10) partition by range (b);`,
			},
			{
				Statement: `create table part_gg2_1 partition of part_gg2 for values from (1) to (5);`,
			},
			{
				Statement: `create table part_gg2_2 partition of part_gg2 for values from (5) to (10);`,
			},
			{
				Statement: `create table part_ee_ff3 partition of part_ee_ff for values from (20) to (30) partition by range (b);`,
			},
			{
				Statement: `create table part_ee_ff3_1 partition of part_ee_ff3 for values from (20) to (25);`,
			},
			{
				Statement: `create table part_ee_ff3_2 partition of part_ee_ff3 for values from (25) to (30);`,
			},
			{
				Statement: `truncate list_parted;`,
			},
			{
				Statement: `insert into list_parted values ('aa'), ('cc');`,
			},
			{
				Statement: `insert into list_parted select 'Ff', s.a from generate_series(1, 29) s(a);`,
			},
			{
				Statement: `insert into list_parted select 'gg', s.a from generate_series(1, 9) s(a);`,
			},
			{
				Statement: `insert into list_parted (b) values (1);`,
			},
			{
				Statement: `select tableoid::regclass::text, a, min(b) as min_b, max(b) as max_b from list_parted group by 1, 2 order by 1;`,
				Results:   []sql.Row{{`part_aa_bb`, `aa`, ``, ``}, {`part_cc_dd`, `cc`, ``, ``}, {`part_ee_ff1`, `Ff`, 1, 9}, {`part_ee_ff2`, `Ff`, 10, 19}, {`part_ee_ff3_1`, `Ff`, 20, 24}, {`part_ee_ff3_2`, `Ff`, 25, 29}, {`part_gg2_1`, `gg`, 1, 4}, {`part_gg2_2`, `gg`, 5, 9}, {`part_null`, ``, 1, 1}},
			},
			{
				Statement: `create table hash_parted (
	a int
) partition by hash (a part_test_int4_ops);`,
			},
			{
				Statement: `create table hpart0 partition of hash_parted for values with (modulus 4, remainder 0);`,
			},
			{
				Statement: `create table hpart1 partition of hash_parted for values with (modulus 4, remainder 1);`,
			},
			{
				Statement: `create table hpart2 partition of hash_parted for values with (modulus 4, remainder 2);`,
			},
			{
				Statement: `create table hpart3 partition of hash_parted for values with (modulus 4, remainder 3);`,
			},
			{
				Statement: `insert into hash_parted values(generate_series(1,10));`,
			},
			{
				Statement: `insert into hpart0 values(12),(16);`,
			},
			{
				Statement:   `insert into hpart0 values(11);`,
				ErrorString: `new row for relation "hpart0" violates partition constraint`,
			},
			{
				Statement: `insert into hpart3 values(11);`,
			},
			{
				Statement: `select tableoid::regclass as part, a, a%4 as "remainder = a % 4"
from hash_parted order by part;`,
				Results: []sql.Row{{`hpart0`, 4, 0}, {`hpart0`, 8, 0}, {`hpart0`, 12, 0}, {`hpart0`, 16, 0}, {`hpart1`, 1, 1}, {`hpart1`, 5, 1}, {`hpart1`, 9, 1}, {`hpart2`, 2, 2}, {`hpart2`, 6, 2}, {`hpart2`, 10, 2}, {`hpart3`, 3, 3}, {`hpart3`, 7, 3}, {`hpart3`, 11, 3}},
			},
			{
				Statement: `\d+ list_parted
                          Partitioned table "public.list_parted"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition key: LIST (lower(a))
Partitions: part_aa_bb FOR VALUES IN ('aa', 'bb'),
            part_cc_dd FOR VALUES IN ('cc', 'dd'),
            part_ee_ff FOR VALUES IN ('ee', 'ff'), PARTITIONED,
            part_gg FOR VALUES IN ('gg'), PARTITIONED,
            part_null FOR VALUES IN (NULL),
            part_xx_yy FOR VALUES IN ('xx', 'yy'), PARTITIONED,
            part_default DEFAULT, PARTITIONED
drop table range_parted, list_parted;`,
			},
			{
				Statement: `drop table hash_parted;`,
			},
			{
				Statement: `create table list_parted (a int) partition by list (a);`,
			},
			{
				Statement: `create table part_default partition of list_parted default;`,
			},
			{
				Statement: `\d+ part_default
                               Table "public.part_default"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
Partition of: list_parted DEFAULT
No partition constraint
insert into part_default values (null);`,
			},
			{
				Statement: `insert into part_default values (1);`,
			},
			{
				Statement: `insert into part_default values (-1);`,
			},
			{
				Statement: `select tableoid::regclass, a from list_parted;`,
				Results:   []sql.Row{{`part_default`, ``}, {`part_default`, 1}, {`part_default`, -1}},
			},
			{
				Statement: `drop table list_parted;`,
			},
			{
				Statement: `create table mlparted (a int, b int) partition by range (a, b);`,
			},
			{
				Statement: `create table mlparted1 (b int not null, a int not null) partition by range ((b+0));`,
			},
			{
				Statement: `create table mlparted11 (like mlparted1);`,
			},
			{
				Statement: `alter table mlparted11 drop a;`,
			},
			{
				Statement: `alter table mlparted11 add a int;`,
			},
			{
				Statement: `alter table mlparted11 drop a;`,
			},
			{
				Statement: `alter table mlparted11 add a int not null;`,
			},
			{
				Statement: `select attrelid::regclass, attname, attnum
from pg_attribute
where attname = 'a'
 and (attrelid = 'mlparted'::regclass
   or attrelid = 'mlparted1'::regclass
   or attrelid = 'mlparted11'::regclass)
order by attrelid::regclass::text;`,
				Results: []sql.Row{{`mlparted`, `a`, 1}, {`mlparted1`, `a`, 2}, {`mlparted11`, `a`, 4}},
			},
			{
				Statement: `alter table mlparted1 attach partition mlparted11 for values from (2) to (5);`,
			},
			{
				Statement: `alter table mlparted attach partition mlparted1 for values from (1, 2) to (1, 10);`,
			},
			{
				Statement: `insert into mlparted values (1, 2);`,
			},
			{
				Statement: `select tableoid::regclass, * from mlparted;`,
				Results:   []sql.Row{{`mlparted11`, 1, 2}},
			},
			{
				Statement:   `insert into mlparted (a, b) values (1, 5);`,
				ErrorString: `no partition of relation "mlparted1" found for row`,
			},
			{
				Statement: `truncate mlparted;`,
			},
			{
				Statement: `alter table mlparted add constraint check_b check (b = 3);`,
			},
			{
				Statement: `create function mlparted11_trig_fn()
returns trigger AS
$$
begin
  NEW.b := 4;`,
			},
			{
				Statement: `  return NEW;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$
language plpgsql;`,
			},
			{
				Statement: `create trigger mlparted11_trig before insert ON mlparted11
  for each row execute procedure mlparted11_trig_fn();`,
			},
			{
				Statement:   `insert into mlparted values (1, 2);`,
				ErrorString: `new row for relation "mlparted11" violates check constraint "check_b"`,
			},
			{
				Statement: `drop trigger mlparted11_trig on mlparted11;`,
			},
			{
				Statement: `drop function mlparted11_trig_fn();`,
			},
			{
				Statement:   `insert into mlparted1 (a, b) values (2, 3);`,
				ErrorString: `new row for relation "mlparted1" violates partition constraint`,
			},
			{
				Statement: `create table lparted_nonullpart (a int, b char) partition by list (b);`,
			},
			{
				Statement: `create table lparted_nonullpart_a partition of lparted_nonullpart for values in ('a');`,
			},
			{
				Statement:   `insert into lparted_nonullpart values (1);`,
				ErrorString: `no partition of relation "lparted_nonullpart" found for row`,
			},
			{
				Statement: `drop table lparted_nonullpart;`,
			},
			{
				Statement: `alter table mlparted drop constraint check_b;`,
			},
			{
				Statement: `create table mlparted12 partition of mlparted1 for values from (5) to (10);`,
			},
			{
				Statement: `create table mlparted2 (b int not null, a int not null);`,
			},
			{
				Statement: `alter table mlparted attach partition mlparted2 for values from (1, 10) to (1, 20);`,
			},
			{
				Statement: `create table mlparted3 partition of mlparted for values from (1, 20) to (1, 30);`,
			},
			{
				Statement: `create table mlparted4 (like mlparted);`,
			},
			{
				Statement: `alter table mlparted4 drop a;`,
			},
			{
				Statement: `alter table mlparted4 add a int not null;`,
			},
			{
				Statement: `alter table mlparted attach partition mlparted4 for values from (1, 30) to (1, 40);`,
			},
			{
				Statement: `with ins (a, b, c) as
  (insert into mlparted (b, a) select s.a, 1 from generate_series(2, 39) s(a) returning tableoid::regclass, *)
  select a, b, min(c), max(c) from ins group by a, b order by 1;`,
				Results: []sql.Row{{`mlparted11`, 1, 2, 4}, {`mlparted12`, 1, 5, 9}, {`mlparted2`, 1, 10, 19}, {`mlparted3`, 1, 20, 29}, {`mlparted4`, 1, 30, 39}},
			},
			{
				Statement: `alter table mlparted add c text;`,
			},
			{
				Statement: `create table mlparted5 (c text, a int not null, b int not null) partition by list (c);`,
			},
			{
				Statement: `create table mlparted5a (a int not null, c text, b int not null);`,
			},
			{
				Statement: `alter table mlparted5 attach partition mlparted5a for values in ('a');`,
			},
			{
				Statement: `alter table mlparted attach partition mlparted5 for values from (1, 40) to (1, 50);`,
			},
			{
				Statement: `alter table mlparted add constraint check_b check (a = 1 and b < 45);`,
			},
			{
				Statement:   `insert into mlparted values (1, 45, 'a');`,
				ErrorString: `new row for relation "mlparted5a" violates check constraint "check_b"`,
			},
			{
				Statement: `create function mlparted5abrtrig_func() returns trigger as $$ begin new.c = 'b'; return new; end; $$ language plpgsql;`,
			},
			{
				Statement: `create trigger mlparted5abrtrig before insert on mlparted5a for each row execute procedure mlparted5abrtrig_func();`,
			},
			{
				Statement:   `insert into mlparted5 (a, b, c) values (1, 40, 'a');`,
				ErrorString: `new row for relation "mlparted5a" violates partition constraint`,
			},
			{
				Statement: `drop table mlparted5;`,
			},
			{
				Statement: `alter table mlparted drop constraint check_b;`,
			},
			{
				Statement: `create table mlparted_def partition of mlparted default partition by range(a);`,
			},
			{
				Statement: `create table mlparted_def1 partition of mlparted_def for values from (40) to (50);`,
			},
			{
				Statement: `create table mlparted_def2 partition of mlparted_def for values from (50) to (60);`,
			},
			{
				Statement: `insert into mlparted values (40, 100);`,
			},
			{
				Statement: `insert into mlparted_def1 values (42, 100);`,
			},
			{
				Statement: `insert into mlparted_def2 values (54, 50);`,
			},
			{
				Statement:   `insert into mlparted values (70, 100);`,
				ErrorString: `no partition of relation "mlparted_def" found for row`,
			},
			{
				Statement:   `insert into mlparted_def1 values (52, 50);`,
				ErrorString: `new row for relation "mlparted_def1" violates partition constraint`,
			},
			{
				Statement:   `insert into mlparted_def2 values (34, 50);`,
				ErrorString: `new row for relation "mlparted_def2" violates partition constraint`,
			},
			{
				Statement: `create table mlparted_defd partition of mlparted_def default;`,
			},
			{
				Statement: `insert into mlparted values (70, 100);`,
			},
			{
				Statement: `select tableoid::regclass, * from mlparted_def;`,
				Results:   []sql.Row{{`mlparted_def1`, 40, 100, ``}, {`mlparted_def1`, 42, 100, ``}, {`mlparted_def2`, 54, 50, ``}, {`mlparted_defd`, 70, 100, ``}},
			},
			{
				Statement: `alter table mlparted add d int, add e int;`,
			},
			{
				Statement: `alter table mlparted drop e;`,
			},
			{
				Statement: `create table mlparted5 partition of mlparted
  for values from (1, 40) to (1, 50) partition by range (c);`,
			},
			{
				Statement: `create table mlparted5_ab partition of mlparted5
  for values from ('a') to ('c') partition by list (c);`,
			},
			{
				Statement: `create table mlparted5_cd partition of mlparted5
  for values from ('c') to ('e') partition by list (c);`,
			},
			{
				Statement: `create table mlparted5_a partition of mlparted5_ab for values in ('a');`,
			},
			{
				Statement: `create table mlparted5_b (d int, b int, c text, a int);`,
			},
			{
				Statement: `alter table mlparted5_ab attach partition mlparted5_b for values in ('b');`,
			},
			{
				Statement: `truncate mlparted;`,
			},
			{
				Statement: `insert into mlparted values (1, 2, 'a', 1);`,
			},
			{
				Statement: `insert into mlparted values (1, 40, 'a', 1);  -- goes to mlparted5_a`,
			},
			{
				Statement: `insert into mlparted values (1, 45, 'b', 1);  -- goes to mlparted5_b`,
			},
			{
				Statement:   `insert into mlparted values (1, 45, 'c', 1);  -- goes to mlparted5_cd, fails`,
				ErrorString: `no partition of relation "mlparted5_cd" found for row`,
			},
			{
				Statement:   `insert into mlparted values (1, 45, 'f', 1);  -- goes to mlparted5, fails`,
				ErrorString: `no partition of relation "mlparted5" found for row`,
			},
			{
				Statement: `select tableoid::regclass, * from mlparted order by a, b, c, d;`,
				Results:   []sql.Row{{`mlparted11`, 1, 2, `a`, 1}, {`mlparted5_a`, 1, 40, `a`, 1}, {`mlparted5_b`, 1, 45, `b`, 1}},
			},
			{
				Statement: `alter table mlparted drop d;`,
			},
			{
				Statement: `truncate mlparted;`,
			},
			{
				Statement: `alter table mlparted add e int, add d int;`,
			},
			{
				Statement: `alter table mlparted drop e;`,
			},
			{
				Statement: `insert into mlparted values (1, 2, 'a', 1);`,
			},
			{
				Statement: `insert into mlparted values (1, 40, 'a', 1);  -- goes to mlparted5_a`,
			},
			{
				Statement: `insert into mlparted values (1, 45, 'b', 1);  -- goes to mlparted5_b`,
			},
			{
				Statement:   `insert into mlparted values (1, 45, 'c', 1);  -- goes to mlparted5_cd, fails`,
				ErrorString: `no partition of relation "mlparted5_cd" found for row`,
			},
			{
				Statement:   `insert into mlparted values (1, 45, 'f', 1);  -- goes to mlparted5, fails`,
				ErrorString: `no partition of relation "mlparted5" found for row`,
			},
			{
				Statement: `select tableoid::regclass, * from mlparted order by a, b, c, d;`,
				Results:   []sql.Row{{`mlparted11`, 1, 2, `a`, 1}, {`mlparted5_a`, 1, 40, `a`, 1}, {`mlparted5_b`, 1, 45, `b`, 1}},
			},
			{
				Statement: `alter table mlparted drop d;`,
			},
			{
				Statement: `drop table mlparted5;`,
			},
			{
				Statement: `create table key_desc (a int, b int) partition by list ((a+0));`,
			},
			{
				Statement: `create table key_desc_1 partition of key_desc for values in (1) partition by range (b);`,
			},
			{
				Statement: `create user regress_insert_other_user;`,
			},
			{
				Statement: `grant select (a) on key_desc_1 to regress_insert_other_user;`,
			},
			{
				Statement: `grant insert on key_desc to regress_insert_other_user;`,
			},
			{
				Statement: `set role regress_insert_other_user;`,
			},
			{
				Statement:   `insert into key_desc values (1, 1);`,
				ErrorString: `no partition of relation "key_desc_1" found for row`,
			},
			{
				Statement: `reset role;`,
			},
			{
				Statement: `grant select (b) on key_desc_1 to regress_insert_other_user;`,
			},
			{
				Statement: `set role regress_insert_other_user;`,
			},
			{
				Statement:   `insert into key_desc values (1, 1);`,
				ErrorString: `no partition of relation "key_desc_1" found for row`,
			},
			{
				Statement:   `insert into key_desc values (2, 1);`,
				ErrorString: `no partition of relation "key_desc" found for row`,
			},
			{
				Statement: `reset role;`,
			},
			{
				Statement: `revoke all on key_desc from regress_insert_other_user;`,
			},
			{
				Statement: `revoke all on key_desc_1 from regress_insert_other_user;`,
			},
			{
				Statement: `drop role regress_insert_other_user;`,
			},
			{
				Statement: `drop table key_desc, key_desc_1;`,
			},
			{
				Statement: `create table mcrparted (a int, b int, c int) partition by range (a, abs(b), c);`,
			},
			{
				Statement:   `create table mcrparted0 partition of mcrparted for values from (minvalue, 0, 0) to (1, maxvalue, maxvalue);`,
				ErrorString: `every bound following MINVALUE must also be MINVALUE`,
			},
			{
				Statement:   `create table mcrparted2 partition of mcrparted for values from (10, 6, minvalue) to (10, maxvalue, minvalue);`,
				ErrorString: `every bound following MAXVALUE must also be MAXVALUE`,
			},
			{
				Statement:   `create table mcrparted4 partition of mcrparted for values from (21, minvalue, 0) to (30, 20, minvalue);`,
				ErrorString: `every bound following MINVALUE must also be MINVALUE`,
			},
			{
				Statement: `create table mcrparted0 partition of mcrparted for values from (minvalue, minvalue, minvalue) to (1, maxvalue, maxvalue);`,
			},
			{
				Statement: `create table mcrparted1 partition of mcrparted for values from (2, 1, minvalue) to (10, 5, 10);`,
			},
			{
				Statement: `create table mcrparted2 partition of mcrparted for values from (10, 6, minvalue) to (10, maxvalue, maxvalue);`,
			},
			{
				Statement: `create table mcrparted3 partition of mcrparted for values from (11, 1, 1) to (20, 10, 10);`,
			},
			{
				Statement: `create table mcrparted4 partition of mcrparted for values from (21, minvalue, minvalue) to (30, 20, maxvalue);`,
			},
			{
				Statement: `create table mcrparted5 partition of mcrparted for values from (30, 21, 20) to (maxvalue, maxvalue, maxvalue);`,
			},
			{
				Statement:   `insert into mcrparted values (null, null, null);`,
				ErrorString: `no partition of relation "mcrparted" found for row`,
			},
			{
				Statement: `insert into mcrparted values (0, 1, 1);`,
			},
			{
				Statement: `insert into mcrparted0 values (0, 1, 1);`,
			},
			{
				Statement: `insert into mcrparted values (9, 1000, 1);`,
			},
			{
				Statement: `insert into mcrparted1 values (9, 1000, 1);`,
			},
			{
				Statement: `insert into mcrparted values (10, 5, -1);`,
			},
			{
				Statement: `insert into mcrparted1 values (10, 5, -1);`,
			},
			{
				Statement: `insert into mcrparted values (2, 1, 0);`,
			},
			{
				Statement: `insert into mcrparted1 values (2, 1, 0);`,
			},
			{
				Statement: `insert into mcrparted values (10, 6, 1000);`,
			},
			{
				Statement: `insert into mcrparted2 values (10, 6, 1000);`,
			},
			{
				Statement: `insert into mcrparted values (10, 1000, 1000);`,
			},
			{
				Statement: `insert into mcrparted2 values (10, 1000, 1000);`,
			},
			{
				Statement:   `insert into mcrparted values (11, 1, -1);`,
				ErrorString: `no partition of relation "mcrparted" found for row`,
			},
			{
				Statement:   `insert into mcrparted3 values (11, 1, -1);`,
				ErrorString: `new row for relation "mcrparted3" violates partition constraint`,
			},
			{
				Statement: `insert into mcrparted values (30, 21, 20);`,
			},
			{
				Statement: `insert into mcrparted5 values (30, 21, 20);`,
			},
			{
				Statement:   `insert into mcrparted4 values (30, 21, 20);	-- error`,
				ErrorString: `new row for relation "mcrparted4" violates partition constraint`,
			},
			{
				Statement: `select tableoid::regclass::text, * from mcrparted order by 1;`,
				Results:   []sql.Row{{`mcrparted0`, 0, 1, 1}, {`mcrparted0`, 0, 1, 1}, {`mcrparted1`, 9, 1000, 1}, {`mcrparted1`, 9, 1000, 1}, {`mcrparted1`, 10, 5, -1}, {`mcrparted1`, 10, 5, -1}, {`mcrparted1`, 2, 1, 0}, {`mcrparted1`, 2, 1, 0}, {`mcrparted2`, 10, 6, 1000}, {`mcrparted2`, 10, 6, 1000}, {`mcrparted2`, 10, 1000, 1000}, {`mcrparted2`, 10, 1000, 1000}, {`mcrparted5`, 30, 21, 20}, {`mcrparted5`, 30, 21, 20}},
			},
			{
				Statement: `drop table mcrparted;`,
			},
			{
				Statement: `create table brtrigpartcon (a int, b text) partition by list (a);`,
			},
			{
				Statement: `create table brtrigpartcon1 partition of brtrigpartcon for values in (1);`,
			},
			{
				Statement: `create or replace function brtrigpartcon1trigf() returns trigger as $$begin new.a := 2; return new; end$$ language plpgsql;`,
			},
			{
				Statement: `create trigger brtrigpartcon1trig before insert on brtrigpartcon1 for each row execute procedure brtrigpartcon1trigf();`,
			},
			{
				Statement:   `insert into brtrigpartcon values (1, 'hi there');`,
				ErrorString: `new row for relation "brtrigpartcon1" violates partition constraint`,
			},
			{
				Statement:   `insert into brtrigpartcon1 values (1, 'hi there');`,
				ErrorString: `new row for relation "brtrigpartcon1" violates partition constraint`,
			},
			{
				Statement: `create table inserttest3 (f1 text default 'foo', f2 text default 'bar', f3 int);`,
			},
			{
				Statement: `create role regress_coldesc_role;`,
			},
			{
				Statement: `grant insert on inserttest3 to regress_coldesc_role;`,
			},
			{
				Statement: `grant insert on brtrigpartcon to regress_coldesc_role;`,
			},
			{
				Statement: `revoke select on brtrigpartcon from regress_coldesc_role;`,
			},
			{
				Statement: `set role regress_coldesc_role;`,
			},
			{
				Statement: `with result as (insert into brtrigpartcon values (1, 'hi there') returning 1)
  insert into inserttest3 (f3) select * from result;`,
				ErrorString: `new row for relation "brtrigpartcon1" violates partition constraint`,
			},
			{
				Statement: `reset role;`,
			},
			{
				Statement: `revoke all on inserttest3 from regress_coldesc_role;`,
			},
			{
				Statement: `revoke all on brtrigpartcon from regress_coldesc_role;`,
			},
			{
				Statement: `drop role regress_coldesc_role;`,
			},
			{
				Statement: `drop table inserttest3;`,
			},
			{
				Statement: `drop table brtrigpartcon;`,
			},
			{
				Statement: `drop function brtrigpartcon1trigf();`,
			},
			{
				Statement: `create table donothingbrtrig_test (a int, b text) partition by list (a);`,
			},
			{
				Statement: `create table donothingbrtrig_test1 (b text, a int);`,
			},
			{
				Statement: `create table donothingbrtrig_test2 (c text, b text, a int);`,
			},
			{
				Statement: `alter table donothingbrtrig_test2 drop column c;`,
			},
			{
				Statement: `create or replace function donothingbrtrig_func() returns trigger as $$begin raise notice 'b: %', new.b; return NULL; end$$ language plpgsql;`,
			},
			{
				Statement: `create trigger donothingbrtrig1 before insert on donothingbrtrig_test1 for each row execute procedure donothingbrtrig_func();`,
			},
			{
				Statement: `create trigger donothingbrtrig2 before insert on donothingbrtrig_test2 for each row execute procedure donothingbrtrig_func();`,
			},
			{
				Statement: `alter table donothingbrtrig_test attach partition donothingbrtrig_test1 for values in (1);`,
			},
			{
				Statement: `alter table donothingbrtrig_test attach partition donothingbrtrig_test2 for values in (2);`,
			},
			{
				Statement: `insert into donothingbrtrig_test values (1, 'foo'), (2, 'bar');`,
			},
			{
				Statement: `copy donothingbrtrig_test from stdout;`,
			},
			{
				Statement: `select tableoid::regclass, * from donothingbrtrig_test;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `drop table donothingbrtrig_test;`,
			},
			{
				Statement: `drop function donothingbrtrig_func();`,
			},
			{
				Statement: `create table mcrparted (a text, b int) partition by range(a, b);`,
			},
			{
				Statement: `create table mcrparted1_lt_b partition of mcrparted for values from (minvalue, minvalue) to ('b', minvalue);`,
			},
			{
				Statement: `create table mcrparted2_b partition of mcrparted for values from ('b', minvalue) to ('c', minvalue);`,
			},
			{
				Statement: `create table mcrparted3_c_to_common partition of mcrparted for values from ('c', minvalue) to ('common', minvalue);`,
			},
			{
				Statement: `create table mcrparted4_common_lt_0 partition of mcrparted for values from ('common', minvalue) to ('common', 0);`,
			},
			{
				Statement: `create table mcrparted5_common_0_to_10 partition of mcrparted for values from ('common', 0) to ('common', 10);`,
			},
			{
				Statement: `create table mcrparted6_common_ge_10 partition of mcrparted for values from ('common', 10) to ('common', maxvalue);`,
			},
			{
				Statement: `create table mcrparted7_gt_common_lt_d partition of mcrparted for values from ('common', maxvalue) to ('d', minvalue);`,
			},
			{
				Statement: `create table mcrparted8_ge_d partition of mcrparted for values from ('d', minvalue) to (maxvalue, maxvalue);`,
			},
			{
				Statement: `\d+ mcrparted
                           Partitioned table "public.mcrparted"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition key: RANGE (a, b)
Partitions: mcrparted1_lt_b FOR VALUES FROM (MINVALUE, MINVALUE) TO ('b', MINVALUE),
            mcrparted2_b FOR VALUES FROM ('b', MINVALUE) TO ('c', MINVALUE),
            mcrparted3_c_to_common FOR VALUES FROM ('c', MINVALUE) TO ('common', MINVALUE),
            mcrparted4_common_lt_0 FOR VALUES FROM ('common', MINVALUE) TO ('common', 0),
            mcrparted5_common_0_to_10 FOR VALUES FROM ('common', 0) TO ('common', 10),
            mcrparted6_common_ge_10 FOR VALUES FROM ('common', 10) TO ('common', MAXVALUE),
            mcrparted7_gt_common_lt_d FOR VALUES FROM ('common', MAXVALUE) TO ('d', MINVALUE),
            mcrparted8_ge_d FOR VALUES FROM ('d', MINVALUE) TO (MAXVALUE, MAXVALUE)
\d+ mcrparted1_lt_b
                              Table "public.mcrparted1_lt_b"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition of: mcrparted FOR VALUES FROM (MINVALUE, MINVALUE) TO ('b', MINVALUE)
Partition constraint: ((a IS NOT NULL) AND (b IS NOT NULL) AND (a < 'b'::text))
\d+ mcrparted2_b
                                Table "public.mcrparted2_b"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition of: mcrparted FOR VALUES FROM ('b', MINVALUE) TO ('c', MINVALUE)
Partition constraint: ((a IS NOT NULL) AND (b IS NOT NULL) AND (a >= 'b'::text) AND (a < 'c'::text))
\d+ mcrparted3_c_to_common
                           Table "public.mcrparted3_c_to_common"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition of: mcrparted FOR VALUES FROM ('c', MINVALUE) TO ('common', MINVALUE)
Partition constraint: ((a IS NOT NULL) AND (b IS NOT NULL) AND (a >= 'c'::text) AND (a < 'common'::text))
\d+ mcrparted4_common_lt_0
                           Table "public.mcrparted4_common_lt_0"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition of: mcrparted FOR VALUES FROM ('common', MINVALUE) TO ('common', 0)
Partition constraint: ((a IS NOT NULL) AND (b IS NOT NULL) AND (a = 'common'::text) AND (b < 0))
\d+ mcrparted5_common_0_to_10
                         Table "public.mcrparted5_common_0_to_10"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition of: mcrparted FOR VALUES FROM ('common', 0) TO ('common', 10)
Partition constraint: ((a IS NOT NULL) AND (b IS NOT NULL) AND (a = 'common'::text) AND (b >= 0) AND (b < 10))
\d+ mcrparted6_common_ge_10
                          Table "public.mcrparted6_common_ge_10"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition of: mcrparted FOR VALUES FROM ('common', 10) TO ('common', MAXVALUE)
Partition constraint: ((a IS NOT NULL) AND (b IS NOT NULL) AND (a = 'common'::text) AND (b >= 10))
\d+ mcrparted7_gt_common_lt_d
                         Table "public.mcrparted7_gt_common_lt_d"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition of: mcrparted FOR VALUES FROM ('common', MAXVALUE) TO ('d', MINVALUE)
Partition constraint: ((a IS NOT NULL) AND (b IS NOT NULL) AND (a > 'common'::text) AND (a < 'd'::text))
\d+ mcrparted8_ge_d
                              Table "public.mcrparted8_ge_d"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           |          |         | plain    |              | 
Partition of: mcrparted FOR VALUES FROM ('d', MINVALUE) TO (MAXVALUE, MAXVALUE)
Partition constraint: ((a IS NOT NULL) AND (b IS NOT NULL) AND (a >= 'd'::text))
insert into mcrparted values ('aaa', 0), ('b', 0), ('bz', 10), ('c', -10),
    ('comm', -10), ('common', -10), ('common', 0), ('common', 10),
    ('commons', 0), ('d', -10), ('e', 0);`,
			},
			{
				Statement: `select tableoid::regclass, * from mcrparted order by a, b;`,
				Results:   []sql.Row{{`mcrparted1_lt_b`, `aaa`, 0}, {`mcrparted2_b`, `b`, 0}, {`mcrparted2_b`, `bz`, 10}, {`mcrparted3_c_to_common`, `c`, -10}, {`mcrparted3_c_to_common`, `comm`, -10}, {`mcrparted4_common_lt_0`, `common`, -10}, {`mcrparted5_common_0_to_10`, `common`, 0}, {`mcrparted6_common_ge_10`, `common`, 10}, {`mcrparted7_gt_common_lt_d`, `commons`, 0}, {`mcrparted8_ge_d`, `d`, -10}, {`mcrparted8_ge_d`, `e`, 0}},
			},
			{
				Statement: `drop table mcrparted;`,
			},
			{
				Statement: `create table returningwrtest (a int) partition by list (a);`,
			},
			{
				Statement: `create table returningwrtest1 partition of returningwrtest for values in (1);`,
			},
			{
				Statement: `insert into returningwrtest values (1) returning returningwrtest;`,
				Results:   []sql.Row{{`(1)`}},
			},
			{
				Statement: `alter table returningwrtest add b text;`,
			},
			{
				Statement: `create table returningwrtest2 (b text, c int, a int);`,
			},
			{
				Statement: `alter table returningwrtest2 drop c;`,
			},
			{
				Statement: `alter table returningwrtest attach partition returningwrtest2 for values in (2);`,
			},
			{
				Statement: `insert into returningwrtest values (2, 'foo') returning returningwrtest;`,
				Results:   []sql.Row{{`(2,foo)`}},
			},
			{
				Statement: `drop table returningwrtest;`,
			},
		},
	})
}
