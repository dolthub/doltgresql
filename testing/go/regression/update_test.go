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

func TestUpdate(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_update)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_update,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE update_test (
    a   INT DEFAULT 10,
    b   INT,
    c   TEXT
);`,
			},
			{
				Statement: `CREATE TABLE upsert_test (
    a   INT PRIMARY KEY,
    b   TEXT
);`,
			},
			{
				Statement: `INSERT INTO update_test VALUES (5, 10, 'foo');`,
			},
			{
				Statement: `INSERT INTO update_test(b, a) VALUES (15, 10);`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{5, 10, `foo`}, {10, 15, ``}},
			},
			{
				Statement: `UPDATE update_test SET a = DEFAULT, b = DEFAULT;`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{10, ``, `foo`}, {10, ``, ``}},
			},
			{
				Statement: `UPDATE update_test AS t SET b = 10 WHERE t.a = 10;`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{10, 10, `foo`}, {10, 10, ``}},
			},
			{
				Statement: `UPDATE update_test t SET b = t.b + 10 WHERE t.a = 10;`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{10, 20, `foo`}, {10, 20, ``}},
			},
			{
				Statement: `UPDATE update_test SET a=v.i FROM (VALUES(100, 20)) AS v(i, j)
  WHERE update_test.b = v.j;`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{100, 20, `foo`}, {100, 20, ``}},
			},
			{
				Statement: `UPDATE update_test SET a = v.* FROM (VALUES(100, 20)) AS v(i, j)
  WHERE update_test.b = v.j;`,
				ErrorString: `column "a" is of type integer but expression is of type record`,
			},
			{
				Statement: `INSERT INTO update_test SELECT a,b+1,c FROM update_test;`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{100, 20, `foo`}, {100, 20, ``}, {100, 21, `foo`}, {100, 21, ``}},
			},
			{
				Statement: `UPDATE update_test SET (c,b,a) = ('bugle', b+11, DEFAULT) WHERE c = 'foo';`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{100, 20, ``}, {100, 21, ``}, {10, 31, `bugle`}, {10, 32, `bugle`}},
			},
			{
				Statement: `UPDATE update_test SET (c,b) = ('car', a+b), a = a + 1 WHERE a = 10;`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{100, 20, ``}, {100, 21, ``}, {11, 41, `car`}, {11, 42, `car`}},
			},
			{
				Statement:   `UPDATE update_test SET (c,b) = ('car', a+b), b = a + 1 WHERE a = 10;`,
				ErrorString: `multiple assignments to same column "b"`,
			},
			{
				Statement: `UPDATE update_test
  SET (b,a) = (select a,b from update_test where b = 41 and c = 'car')
  WHERE a = 100 AND b = 20;`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{100, 21, ``}, {11, 41, `car`}, {11, 42, `car`}, {41, 11, ``}},
			},
			{
				Statement: `UPDATE update_test o
  SET (b,a) = (select a+1,b from update_test i
               where i.a=o.a and i.b=o.b and i.c is not distinct from o.c);`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{21, 101, ``}, {41, 12, `car`}, {42, 12, `car`}, {11, 42, ``}},
			},
			{
				Statement:   `UPDATE update_test SET (b,a) = (select a+1,b from update_test);`,
				ErrorString: `more than one row returned by a subquery used as an expression`,
			},
			{
				Statement: `UPDATE update_test SET (b,a) = (select a+1,b from update_test where a = 1000)
  WHERE a = 11;`,
			},
			{
				Statement: `SELECT * FROM update_test;`,
				Results:   []sql.Row{{21, 101, ``}, {41, 12, `car`}, {42, 12, `car`}, {``, ``, ``}},
			},
			{
				Statement: `UPDATE update_test SET (a,b) = ROW(v.*) FROM (VALUES(21, 100)) AS v(i, j)
  WHERE update_test.a = v.i;`,
			},
			{
				Statement: `UPDATE update_test SET (a,b) = (v.*) FROM (VALUES(21, 101)) AS v(i, j)
  WHERE update_test.a = v.i;`,
				ErrorString: `source for a multiple-column UPDATE item must be a sub-SELECT or ROW() expression`,
			},
			{
				Statement:   `UPDATE update_test AS t SET b = update_test.b + 10 WHERE t.a = 10;`,
				ErrorString: `invalid reference to FROM-clause entry for table "update_test"`,
			},
			{
				Statement: `UPDATE update_test SET c = repeat('x', 10000) WHERE c = 'car';`,
			},
			{
				Statement: `SELECT a, b, char_length(c) FROM update_test;`,
				Results:   []sql.Row{{``, ``, ``}, {21, 100, ``}, {41, 12, 10000}, {42, 12, 10000}},
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
UPDATE update_test t
  SET (a, b) = (SELECT b, a FROM update_test s WHERE s.a = t.a)
  WHERE CURRENT_USER = SESSION_USER;`,
				Results: []sql.Row{{`Update on public.update_test t`}, {`->  Result`}, {`Output: $1, $2, (SubPlan 1 (returns $1,$2)), t.ctid`}, {`One-Time Filter: (CURRENT_USER = SESSION_USER)`}, {`->  Seq Scan on public.update_test t`}, {`Output: t.a, t.ctid`}, {`SubPlan 1 (returns $1,$2)`}, {`->  Seq Scan on public.update_test s`}, {`Output: s.b, s.a`}, {`Filter: (s.a = t.a)`}},
			},
			{
				Statement: `UPDATE update_test t
  SET (a, b) = (SELECT b, a FROM update_test s WHERE s.a = t.a)
  WHERE CURRENT_USER = SESSION_USER;`,
			},
			{
				Statement: `SELECT a, b, char_length(c) FROM update_test;`,
				Results:   []sql.Row{{``, ``, ``}, {100, 21, ``}, {12, 41, 10000}, {12, 42, 10000}},
			},
			{
				Statement: `INSERT INTO upsert_test VALUES(1, 'Boo'), (3, 'Zoo');`,
			},
			{
				Statement: `WITH aaa AS (SELECT 1 AS a, 'Foo' AS b) INSERT INTO upsert_test
  VALUES (1, 'Bar') ON CONFLICT(a)
  DO UPDATE SET (b, a) = (SELECT b, a FROM aaa) RETURNING *;`,
				Results: []sql.Row{{1, `Foo`}},
			},
			{
				Statement: `INSERT INTO upsert_test VALUES (1, 'Baz'), (3, 'Zaz') ON CONFLICT(a)
  DO UPDATE SET (b, a) = (SELECT b || ', Correlated', a from upsert_test i WHERE i.a = upsert_test.a)
  RETURNING *;`,
				Results: []sql.Row{{1, `Foo, Correlated`}, {3, `Zoo, Correlated`}},
			},
			{
				Statement: `INSERT INTO upsert_test VALUES (1, 'Bat'), (3, 'Zot') ON CONFLICT(a)
  DO UPDATE SET (b, a) = (SELECT b || ', Excluded', a from upsert_test i WHERE i.a = excluded.a)
  RETURNING *;`,
				Results: []sql.Row{{1, `Foo, Correlated, Excluded`}, {3, `Zoo, Correlated, Excluded`}},
			},
			{
				Statement: `INSERT INTO upsert_test VALUES (2, 'Beeble') ON CONFLICT(a)
  DO UPDATE SET (b, a) = (SELECT b || ', Excluded', a from upsert_test i WHERE i.a = excluded.a)
  RETURNING tableoid::regclass, xmin = pg_current_xact_id()::xid AS xmin_correct, xmax = 0 AS xmax_correct;`,
				Results: []sql.Row{{`upsert_test`, true, true}},
			},
			{
				Statement: `INSERT INTO upsert_test VALUES (2, 'Brox') ON CONFLICT(a)
  DO UPDATE SET (b, a) = (SELECT b || ', Excluded', a from upsert_test i WHERE i.a = excluded.a)
  RETURNING tableoid::regclass, xmin = pg_current_xact_id()::xid AS xmin_correct, xmax = pg_current_xact_id()::xid AS xmax_correct;`,
				Results: []sql.Row{{`upsert_test`, true, true}},
			},
			{
				Statement: `DROP TABLE update_test;`,
			},
			{
				Statement: `DROP TABLE upsert_test;`,
			},
			{
				Statement: `CREATE TABLE upsert_test (
    a   INT PRIMARY KEY,
    b   TEXT
) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE upsert_test_1 PARTITION OF upsert_test FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE upsert_test_2 (b TEXT, a INT PRIMARY KEY);`,
			},
			{
				Statement: `ALTER TABLE upsert_test ATTACH PARTITION upsert_test_2 FOR VALUES IN (2);`,
			},
			{
				Statement: `INSERT INTO upsert_test VALUES(1, 'Boo'), (2, 'Zoo');`,
			},
			{
				Statement: `WITH aaa AS (SELECT 1 AS a, 'Foo' AS b) INSERT INTO upsert_test
  VALUES (1, 'Bar') ON CONFLICT(a)
  DO UPDATE SET (b, a) = (SELECT b, a FROM aaa) RETURNING *;`,
				Results: []sql.Row{{1, `Foo`}},
			},
			{
				Statement: `WITH aaa AS (SELECT 1 AS ctea, ' Foo' AS cteb) INSERT INTO upsert_test
  VALUES (1, 'Bar'), (2, 'Baz') ON CONFLICT(a)
  DO UPDATE SET (b, a) = (SELECT upsert_test.b||cteb, upsert_test.a FROM aaa) RETURNING *;`,
				Results: []sql.Row{{1, `Foo Foo`}, {2, `Zoo Foo`}},
			},
			{
				Statement: `DROP TABLE upsert_test;`,
			},
			{
				Statement: `---------------------------
---------------------------
CREATE TABLE range_parted (
	a text,
	b bigint,
	c numeric,
	d int,
	e varchar
) PARTITION BY RANGE (a, b);`,
			},
			{
				Statement: `CREATE TABLE part_b_20_b_30 (e varchar, c numeric, a text, b bigint, d int);`,
			},
			{
				Statement: `ALTER TABLE range_parted ATTACH PARTITION part_b_20_b_30 FOR VALUES FROM ('b', 20) TO ('b', 30);`,
			},
			{
				Statement: `CREATE TABLE part_b_10_b_20 (e varchar, c numeric, a text, b bigint, d int) PARTITION BY RANGE (c);`,
			},
			{
				Statement: `CREATE TABLE part_b_1_b_10 PARTITION OF range_parted FOR VALUES FROM ('b', 1) TO ('b', 10);`,
			},
			{
				Statement: `ALTER TABLE range_parted ATTACH PARTITION part_b_10_b_20 FOR VALUES FROM ('b', 10) TO ('b', 20);`,
			},
			{
				Statement: `CREATE TABLE part_a_10_a_20 PARTITION OF range_parted FOR VALUES FROM ('a', 10) TO ('a', 20);`,
			},
			{
				Statement: `CREATE TABLE part_a_1_a_10 PARTITION OF range_parted FOR VALUES FROM ('a', 1) TO ('a', 10);`,
			},
			{
				Statement: `UPDATE part_b_10_b_20 set b = b - 6;`,
			},
			{
				Statement: `CREATE TABLE part_c_100_200 (e varchar, c numeric, a text, b bigint, d int) PARTITION BY range (abs(d));`,
			},
			{
				Statement: `ALTER TABLE part_c_100_200 DROP COLUMN e, DROP COLUMN c, DROP COLUMN a;`,
			},
			{
				Statement: `ALTER TABLE part_c_100_200 ADD COLUMN c numeric, ADD COLUMN e varchar, ADD COLUMN a text;`,
			},
			{
				Statement: `ALTER TABLE part_c_100_200 DROP COLUMN b;`,
			},
			{
				Statement: `ALTER TABLE part_c_100_200 ADD COLUMN b bigint;`,
			},
			{
				Statement: `CREATE TABLE part_d_1_15 PARTITION OF part_c_100_200 FOR VALUES FROM (1) TO (15);`,
			},
			{
				Statement: `CREATE TABLE part_d_15_20 PARTITION OF part_c_100_200 FOR VALUES FROM (15) TO (20);`,
			},
			{
				Statement: `ALTER TABLE part_b_10_b_20 ATTACH PARTITION part_c_100_200 FOR VALUES FROM (100) TO (200);`,
			},
			{
				Statement: `CREATE TABLE part_c_1_100 (e varchar, d int, c numeric, b bigint, a text);`,
			},
			{
				Statement: `ALTER TABLE part_b_10_b_20 ATTACH PARTITION part_c_1_100 FOR VALUES FROM (1) TO (100);`,
			},
			{
				Statement: `\set init_range_parted 'truncate range_parted; insert into range_parted VALUES (''a'', 1, 1, 1), (''a'', 10, 200, 1), (''b'', 12, 96, 1), (''b'', 13, 97, 2), (''b'', 15, 105, 16), (''b'', 17, 105, 19)'
\set show_data 'select tableoid::regclass::text COLLATE "C" partname, * from range_parted ORDER BY 1, 2, 3, 4, 5, 6'
:init_range_parted;`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_c_1_100`, `b`, 12, 96, 1, ``}, {`part_c_1_100`, `b`, 13, 97, 2, ``}, {`part_d_15_20`, `b`, 15, 105, 16, ``}, {`part_d_15_20`, `b`, 17, 105, 19, ``}},
			},
			{
				Statement: `EXPLAIN (costs off) UPDATE range_parted set c = c - 50 WHERE c > 97;`,
				Results:   []sql.Row{{`Update on range_parted`}, {`Update on part_a_1_a_10 range_parted_1`}, {`Update on part_a_10_a_20 range_parted_2`}, {`Update on part_b_1_b_10 range_parted_3`}, {`Update on part_c_1_100 range_parted_4`}, {`Update on part_d_1_15 range_parted_5`}, {`Update on part_d_15_20 range_parted_6`}, {`Update on part_b_20_b_30 range_parted_7`}, {`->  Append`}, {`->  Seq Scan on part_a_1_a_10 range_parted_1`}, {`Filter: (c > '97'::numeric)`}, {`->  Seq Scan on part_a_10_a_20 range_parted_2`}, {`Filter: (c > '97'::numeric)`}, {`->  Seq Scan on part_b_1_b_10 range_parted_3`}, {`Filter: (c > '97'::numeric)`}, {`->  Seq Scan on part_c_1_100 range_parted_4`}, {`Filter: (c > '97'::numeric)`}, {`->  Seq Scan on part_d_1_15 range_parted_5`}, {`Filter: (c > '97'::numeric)`}, {`->  Seq Scan on part_d_15_20 range_parted_6`}, {`Filter: (c > '97'::numeric)`}, {`->  Seq Scan on part_b_20_b_30 range_parted_7`}, {`Filter: (c > '97'::numeric)`}},
			},
			{
				Statement:   `UPDATE part_c_100_200 set c = c - 20, d = c WHERE c = 105;`,
				ErrorString: `new row for relation "part_c_100_200" violates partition constraint`,
			},
			{
				Statement:   `UPDATE part_b_10_b_20 set a = 'a';`,
				ErrorString: `new row for relation "part_b_10_b_20" violates partition constraint`,
			},
			{
				Statement: `UPDATE range_parted set d = d - 10 WHERE d > 10;`,
			},
			{
				Statement: `UPDATE range_parted set e = d;`,
			},
			{
				Statement: `UPDATE part_c_1_100 set c = c + 20 WHERE c = 98;`,
			},
			{
				Statement: `UPDATE part_b_10_b_20 set c = c + 20 returning c, b, a;`,
				Results:   []sql.Row{{116, 12, `b`}, {117, 13, `b`}, {125, 15, `b`}, {125, 17, `b`}},
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, 1}, {`part_a_1_a_10`, `a`, 1, 1, 1, 1}, {`part_d_1_15`, `b`, 12, 116, 1, 1}, {`part_d_1_15`, `b`, 13, 117, 2, 2}, {`part_d_1_15`, `b`, 15, 125, 6, 6}, {`part_d_1_15`, `b`, 17, 125, 9, 9}},
			},
			{
				Statement:   `UPDATE part_b_10_b_20 set b = b - 6 WHERE c > 116 returning *;`,
				ErrorString: `new row for relation "part_b_10_b_20" violates partition constraint`,
			},
			{
				Statement: `UPDATE range_parted set b = b - 6 WHERE c > 116 returning a, b + c;`,
				Results:   []sql.Row{{`a`, 204}, {`b`, 124}, {`b`, 134}, {`b`, 136}},
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_1_a_10`, `a`, 1, 1, 1, 1}, {`part_a_1_a_10`, `a`, 4, 200, 1, 1}, {`part_b_1_b_10`, `b`, 7, 117, 2, 2}, {`part_b_1_b_10`, `b`, 9, 125, 6, 6}, {`part_d_1_15`, `b`, 11, 125, 9, 9}, {`part_d_1_15`, `b`, 12, 116, 1, 1}},
			},
			{
				Statement: `CREATE TABLE mintab(c1 int);`,
			},
			{
				Statement: `INSERT into mintab VALUES (120);`,
			},
			{
				Statement: `CREATE VIEW upview AS SELECT * FROM range_parted WHERE (select c > c1 FROM mintab) WITH CHECK OPTION;`,
			},
			{
				Statement: `UPDATE upview set c = 199 WHERE b = 4;`,
			},
			{
				Statement:   `UPDATE upview set c = 120 WHERE b = 4;`,
				ErrorString: `new row violates check option for view "upview"`,
			},
			{
				Statement:   `UPDATE upview set a = 'b', b = 15, c = 120 WHERE b = 4;`,
				ErrorString: `new row violates check option for view "upview"`,
			},
			{
				Statement: `UPDATE upview set a = 'b', b = 15 WHERE b = 4;`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_1_a_10`, `a`, 1, 1, 1, 1}, {`part_b_1_b_10`, `b`, 7, 117, 2, 2}, {`part_b_1_b_10`, `b`, 9, 125, 6, 6}, {`part_d_1_15`, `b`, 11, 125, 9, 9}, {`part_d_1_15`, `b`, 12, 116, 1, 1}, {`part_d_1_15`, `b`, 15, 199, 1, 1}},
			},
			{
				Statement: `DROP VIEW upview;`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `UPDATE range_parted set c = 95 WHERE a = 'b' and b > 10 and c > 100 returning (range_parted), *;`,
				Results:   []sql.Row{{`(b,15,95,16,)`, `b`, 15, 95, 16, ``}, {`(b,17,95,19,)`, `b`, 17, 95, 19, ``}},
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_c_1_100`, `b`, 12, 96, 1, ``}, {`part_c_1_100`, `b`, 13, 97, 2, ``}, {`part_c_1_100`, `b`, 15, 95, 16, ``}, {`part_c_1_100`, `b`, 17, 95, 19, ``}},
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `CREATE FUNCTION trans_updatetrigfunc() RETURNS trigger LANGUAGE plpgsql AS
$$
  begin
    raise notice 'trigger = %, old table = %, new table = %',
                 TG_NAME,
                 (select string_agg(old_table::text, ', ' ORDER BY a) FROM old_table),
                 (select string_agg(new_table::text, ', ' ORDER BY a) FROM new_table);`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER trans_updatetrig
  AFTER UPDATE ON range_parted REFERENCING OLD TABLE AS old_table NEW TABLE AS new_table
  FOR EACH STATEMENT EXECUTE PROCEDURE trans_updatetrigfunc();`,
			},
			{
				Statement: `UPDATE range_parted set c = (case when c = 96 then 110 else c + 1 end ) WHERE a = 'b' and b > 10 and c >= 96;`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_c_1_100`, `b`, 13, 98, 2, ``}, {`part_d_15_20`, `b`, 15, 106, 16, ``}, {`part_d_15_20`, `b`, 17, 106, 19, ``}, {`part_d_1_15`, `b`, 12, 110, 1, ``}},
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `CREATE TRIGGER trans_deletetrig
  AFTER DELETE ON range_parted REFERENCING OLD TABLE AS old_table
  FOR EACH STATEMENT EXECUTE PROCEDURE trans_updatetrigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER trans_inserttrig
  AFTER INSERT ON range_parted REFERENCING NEW TABLE AS new_table
  FOR EACH STATEMENT EXECUTE PROCEDURE trans_updatetrigfunc();`,
			},
			{
				Statement: `UPDATE range_parted set c = c + 50 WHERE a = 'b' and b > 10 and c >= 96;`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_d_15_20`, `b`, 15, 155, 16, ``}, {`part_d_15_20`, `b`, 17, 155, 19, ``}, {`part_d_1_15`, `b`, 12, 146, 1, ``}, {`part_d_1_15`, `b`, 13, 147, 2, ``}},
			},
			{
				Statement: `DROP TRIGGER trans_deletetrig ON range_parted;`,
			},
			{
				Statement: `DROP TRIGGER trans_inserttrig ON range_parted;`,
			},
			{
				Statement: `CREATE FUNCTION func_parted_mod_b() RETURNS trigger AS $$
BEGIN
   NEW.b = NEW.b + 1;`,
			},
			{
				Statement: `   return NEW;`,
			},
			{
				Statement: `END $$ language plpgsql;`,
			},
			{
				Statement: `CREATE TRIGGER trig_c1_100 BEFORE UPDATE OR INSERT ON part_c_1_100
   FOR EACH ROW EXECUTE PROCEDURE func_parted_mod_b();`,
			},
			{
				Statement: `CREATE TRIGGER trig_d1_15 BEFORE UPDATE OR INSERT ON part_d_1_15
   FOR EACH ROW EXECUTE PROCEDURE func_parted_mod_b();`,
			},
			{
				Statement: `CREATE TRIGGER trig_d15_20 BEFORE UPDATE OR INSERT ON part_d_15_20
   FOR EACH ROW EXECUTE PROCEDURE func_parted_mod_b();`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `UPDATE range_parted set c = (case when c = 96 then 110 else c + 1 end) WHERE a = 'b' and b > 10 and c >= 96;`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_c_1_100`, `b`, 15, 98, 2, ``}, {`part_d_15_20`, `b`, 17, 106, 16, ``}, {`part_d_15_20`, `b`, 19, 106, 19, ``}, {`part_d_1_15`, `b`, 15, 110, 1, ``}},
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `UPDATE range_parted set c = c + 50 WHERE a = 'b' and b > 10 and c >= 96;`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_d_15_20`, `b`, 17, 155, 16, ``}, {`part_d_15_20`, `b`, 19, 155, 19, ``}, {`part_d_1_15`, `b`, 15, 146, 1, ``}, {`part_d_1_15`, `b`, 16, 147, 2, ``}},
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `UPDATE range_parted set b = 15 WHERE b = 1;`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_10_a_20`, `a`, 15, 1, 1, ``}, {`part_c_1_100`, `b`, 13, 96, 1, ``}, {`part_c_1_100`, `b`, 14, 97, 2, ``}, {`part_d_15_20`, `b`, 16, 105, 16, ``}, {`part_d_15_20`, `b`, 18, 105, 19, ``}},
			},
			{
				Statement: `DROP TRIGGER trans_updatetrig ON range_parted;`,
			},
			{
				Statement: `DROP TRIGGER trig_c1_100 ON part_c_1_100;`,
			},
			{
				Statement: `DROP TRIGGER trig_d1_15 ON part_d_1_15;`,
			},
			{
				Statement: `DROP TRIGGER trig_d15_20 ON part_d_15_20;`,
			},
			{
				Statement: `DROP FUNCTION func_parted_mod_b();`,
			},
			{
				Statement: `-----------------------------------------
ALTER TABLE range_parted ENABLE ROW LEVEL SECURITY;`,
			},
			{
				Statement: `CREATE USER regress_range_parted_user;`,
			},
			{
				Statement: `GRANT ALL ON range_parted, mintab TO regress_range_parted_user;`,
			},
			{
				Statement: `CREATE POLICY seeall ON range_parted AS PERMISSIVE FOR SELECT USING (true);`,
			},
			{
				Statement: `CREATE POLICY policy_range_parted ON range_parted for UPDATE USING (true) WITH CHECK (c % 2 = 0);`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_range_parted_user;`,
			},
			{
				Statement:   `UPDATE range_parted set a = 'b', c = 151 WHERE a = 'a' and c = 200;`,
				ErrorString: `new row violates row-level security policy for table "range_parted"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `CREATE FUNCTION func_d_1_15() RETURNS trigger AS $$
BEGIN
   NEW.c = NEW.c + 1; -- Make even numbers odd, or vice versa`,
			},
			{
				Statement: `   return NEW;`,
			},
			{
				Statement: `END $$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE TRIGGER trig_d_1_15 BEFORE INSERT ON part_d_1_15
   FOR EACH ROW EXECUTE PROCEDURE func_d_1_15();`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_range_parted_user;`,
			},
			{
				Statement: `UPDATE range_parted set a = 'b', c = 151 WHERE a = 'a' and c = 200;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_range_parted_user;`,
			},
			{
				Statement:   `UPDATE range_parted set a = 'b', c = 150 WHERE a = 'a' and c = 200;`,
				ErrorString: `new row violates row-level security policy for table "range_parted"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP TRIGGER trig_d_1_15 ON part_d_1_15;`,
			},
			{
				Statement: `DROP FUNCTION func_d_1_15();`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `CREATE POLICY policy_range_parted_subplan on range_parted
    AS RESTRICTIVE for UPDATE USING (true)
    WITH CHECK ((SELECT range_parted.c <= c1 FROM mintab));`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_range_parted_user;`,
			},
			{
				Statement:   `UPDATE range_parted set a = 'b', c = 122 WHERE a = 'a' and c = 200;`,
				ErrorString: `new row violates row-level security policy "policy_range_parted_subplan" for table "range_parted"`,
			},
			{
				Statement: `UPDATE range_parted set a = 'b', c = 120 WHERE a = 'a' and c = 200;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `CREATE POLICY policy_range_parted_wholerow on range_parted AS RESTRICTIVE for UPDATE USING (true)
   WITH CHECK (range_parted = row('b', 10, 112, 1, NULL)::range_parted);`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_range_parted_user;`,
			},
			{
				Statement: `UPDATE range_parted set a = 'b', c = 112 WHERE a = 'a' and c = 200;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_range_parted_user;`,
			},
			{
				Statement:   `UPDATE range_parted set a = 'b', c = 116 WHERE a = 'a' and c = 200;`,
				ErrorString: `new row violates row-level security policy "policy_range_parted_wholerow" for table "range_parted"`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP POLICY policy_range_parted ON range_parted;`,
			},
			{
				Statement: `DROP POLICY policy_range_parted_subplan ON range_parted;`,
			},
			{
				Statement: `DROP POLICY policy_range_parted_wholerow ON range_parted;`,
			},
			{
				Statement: `REVOKE ALL ON range_parted, mintab FROM regress_range_parted_user;`,
			},
			{
				Statement: `DROP USER regress_range_parted_user;`,
			},
			{
				Statement: `DROP TABLE mintab;`,
			},
			{
				Statement: `---------------------------------------------------
:init_range_parted;`,
			},
			{
				Statement: `CREATE FUNCTION trigfunc() returns trigger language plpgsql as
$$
  begin
    raise notice 'trigger = % fired on table % during %',
                 TG_NAME, TG_TABLE_NAME, TG_OP;`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER parent_delete_trig
  AFTER DELETE ON range_parted for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER parent_update_trig
  AFTER UPDATE ON range_parted for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER parent_insert_trig
  AFTER INSERT ON range_parted for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER c1_delete_trig
  AFTER DELETE ON part_c_1_100 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER c1_update_trig
  AFTER UPDATE ON part_c_1_100 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER c1_insert_trig
  AFTER INSERT ON part_c_1_100 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER d1_delete_trig
  AFTER DELETE ON part_d_1_15 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER d1_update_trig
  AFTER UPDATE ON part_d_1_15 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER d1_insert_trig
  AFTER INSERT ON part_d_1_15 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER d15_delete_trig
  AFTER DELETE ON part_d_15_20 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER d15_update_trig
  AFTER UPDATE ON part_d_15_20 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `CREATE TRIGGER d15_insert_trig
  AFTER INSERT ON part_d_15_20 for each statement execute procedure trigfunc();`,
			},
			{
				Statement: `UPDATE range_parted set c = c - 50 WHERE c > 97;`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 150, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_c_1_100`, `b`, 12, 96, 1, ``}, {`part_c_1_100`, `b`, 13, 97, 2, ``}, {`part_c_1_100`, `b`, 15, 55, 16, ``}, {`part_c_1_100`, `b`, 17, 55, 19, ``}},
			},
			{
				Statement: `DROP TRIGGER parent_delete_trig ON range_parted;`,
			},
			{
				Statement: `DROP TRIGGER parent_update_trig ON range_parted;`,
			},
			{
				Statement: `DROP TRIGGER parent_insert_trig ON range_parted;`,
			},
			{
				Statement: `DROP TRIGGER c1_delete_trig ON part_c_1_100;`,
			},
			{
				Statement: `DROP TRIGGER c1_update_trig ON part_c_1_100;`,
			},
			{
				Statement: `DROP TRIGGER c1_insert_trig ON part_c_1_100;`,
			},
			{
				Statement: `DROP TRIGGER d1_delete_trig ON part_d_1_15;`,
			},
			{
				Statement: `DROP TRIGGER d1_update_trig ON part_d_1_15;`,
			},
			{
				Statement: `DROP TRIGGER d1_insert_trig ON part_d_1_15;`,
			},
			{
				Statement: `DROP TRIGGER d15_delete_trig ON part_d_15_20;`,
			},
			{
				Statement: `DROP TRIGGER d15_update_trig ON part_d_15_20;`,
			},
			{
				Statement: `DROP TRIGGER d15_insert_trig ON part_d_15_20;`,
			},
			{
				Statement: `:init_range_parted;`,
			},
			{
				Statement: `create table part_def partition of range_parted default;`,
			},
			{
				Statement: `\d+ part_def
                                       Table "public.part_def"
 Column |       Type        | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+-------------------+-----------+----------+---------+----------+--------------+-------------
 a      | text              |           |          |         | extended |              | 
 b      | bigint            |           |          |         | plain    |              | 
 c      | numeric           |           |          |         | main     |              | 
 d      | integer           |           |          |         | plain    |              | 
 e      | character varying |           |          |         | extended |              | 
Partition of: range_parted DEFAULT
Partition constraint: (NOT ((a IS NOT NULL) AND (b IS NOT NULL) AND (((a = 'a'::text) AND (b >= '1'::bigint) AND (b < '10'::bigint)) OR ((a = 'a'::text) AND (b >= '10'::bigint) AND (b < '20'::bigint)) OR ((a = 'b'::text) AND (b >= '1'::bigint) AND (b < '10'::bigint)) OR ((a = 'b'::text) AND (b >= '10'::bigint) AND (b < '20'::bigint)) OR ((a = 'b'::text) AND (b >= '20'::bigint) AND (b < '30'::bigint)))))
insert into range_parted values ('c', 9);`,
			},
			{
				Statement: `update part_def set a = 'd' where a = 'c';`,
			},
			{
				Statement:   `update part_def set a = 'a' where a = 'd';`,
				ErrorString: `new row for relation "part_def" violates partition constraint`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_c_1_100`, `b`, 12, 96, 1, ``}, {`part_c_1_100`, `b`, 13, 97, 2, ``}, {`part_d_15_20`, `b`, 15, 105, 16, ``}, {`part_d_15_20`, `b`, 17, 105, 19, ``}, {`part_def`, `d`, 9, ``, ``, ``}},
			},
			{
				Statement:   `UPDATE part_a_10_a_20 set a = 'ad' WHERE a = 'a';`,
				ErrorString: `new row for relation "part_a_10_a_20" violates partition constraint`,
			},
			{
				Statement: `UPDATE range_parted set a = 'ad' WHERE a = 'a';`,
			},
			{
				Statement: `UPDATE range_parted set a = 'bd' WHERE a = 'b';`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_def`, `ad`, 1, 1, 1, ``}, {`part_def`, `ad`, 10, 200, 1, ``}, {`part_def`, `bd`, 12, 96, 1, ``}, {`part_def`, `bd`, 13, 97, 2, ``}, {`part_def`, `bd`, 15, 105, 16, ``}, {`part_def`, `bd`, 17, 105, 19, ``}, {`part_def`, `d`, 9, ``, ``, ``}},
			},
			{
				Statement: `UPDATE range_parted set a = 'a' WHERE a = 'ad';`,
			},
			{
				Statement: `UPDATE range_parted set a = 'b' WHERE a = 'bd';`,
			},
			{
				Statement: `:show_data;`,
				Results:   []sql.Row{{`part_a_10_a_20`, `a`, 10, 200, 1, ``}, {`part_a_1_a_10`, `a`, 1, 1, 1, ``}, {`part_c_1_100`, `b`, 12, 96, 1, ``}, {`part_c_1_100`, `b`, 13, 97, 2, ``}, {`part_d_15_20`, `b`, 15, 105, 16, ``}, {`part_d_15_20`, `b`, 17, 105, 19, ``}, {`part_def`, `d`, 9, ``, ``, ``}},
			},
			{
				Statement: `DROP TABLE range_parted;`,
			},
			{
				Statement: `CREATE TABLE list_parted (
	a text,
	b int
) PARTITION BY list (a);`,
			},
			{
				Statement: `CREATE TABLE list_part1  PARTITION OF list_parted for VALUES in ('a', 'b');`,
			},
			{
				Statement: `CREATE TABLE list_default PARTITION OF list_parted default;`,
			},
			{
				Statement: `INSERT into list_part1 VALUES ('a', 1);`,
			},
			{
				Statement: `INSERT into list_default VALUES ('d', 10);`,
			},
			{
				Statement:   `UPDATE list_default set a = 'a' WHERE a = 'd';`,
				ErrorString: `new row for relation "list_default" violates partition constraint`,
			},
			{
				Statement: `UPDATE list_default set a = 'x' WHERE a = 'd';`,
			},
			{
				Statement: `DROP TABLE list_parted;`,
			},
			{
				Statement: `create table utrtest (a int, b text) partition by list (a);`,
			},
			{
				Statement: `create table utr1 (a int check (a in (1)), q text, b text);`,
			},
			{
				Statement: `create table utr2 (a int check (a in (2)), b text);`,
			},
			{
				Statement: `alter table utr1 drop column q;`,
			},
			{
				Statement: `alter table utrtest attach partition utr1 for values in (1);`,
			},
			{
				Statement: `alter table utrtest attach partition utr2 for values in (2);`,
			},
			{
				Statement: `insert into utrtest values (1, 'foo')
  returning *, tableoid::regclass, xmin = pg_current_xact_id()::xid as xmin_ok;`,
				Results: []sql.Row{{1, `foo`, `utr1`, true}},
			},
			{
				Statement: `insert into utrtest values (2, 'bar')
  returning *, tableoid::regclass, xmin = pg_current_xact_id()::xid as xmin_ok;  -- fails`,
				ErrorString: `cannot retrieve a system column in this context`,
			},
			{
				Statement: `insert into utrtest values (2, 'bar')
  returning *, tableoid::regclass;`,
				Results: []sql.Row{{2, `bar`, `utr2`}},
			},
			{
				Statement: `update utrtest set b = b || b from (values (1), (2)) s(x) where a = s.x
  returning *, tableoid::regclass, xmin = pg_current_xact_id()::xid as xmin_ok;`,
				Results: []sql.Row{{1, `foofoo`, 1, `utr1`, true}, {2, `barbar`, 2, `utr2`, true}},
			},
			{
				Statement: `update utrtest set a = 3 - a from (values (1), (2)) s(x) where a = s.x
  returning *, tableoid::regclass, xmin = pg_current_xact_id()::xid as xmin_ok;  -- fails`,
				ErrorString: `cannot retrieve a system column in this context`,
			},
			{
				Statement: `update utrtest set a = 3 - a from (values (1), (2)) s(x) where a = s.x
  returning *, tableoid::regclass;`,
				Results: []sql.Row{{2, `foofoo`, 1, `utr2`}, {1, `barbar`, 2, `utr1`}},
			},
			{
				Statement: `delete from utrtest
  returning *, tableoid::regclass, xmax = pg_current_xact_id()::xid as xmax_ok;`,
				Results: []sql.Row{{1, `barbar`, `utr1`, true}, {2, `foofoo`, `utr2`, true}},
			},
			{
				Statement: `drop table utrtest;`,
			},
			{
				Statement: `--------------
--------------
CREATE TABLE list_parted (a numeric, b int, c int8) PARTITION BY list (a);`,
			},
			{
				Statement: `CREATE TABLE sub_parted PARTITION OF list_parted for VALUES in (1) PARTITION BY list (b);`,
			},
			{
				Statement: `CREATE TABLE sub_part1(b int, c int8, a numeric);`,
			},
			{
				Statement: `ALTER TABLE sub_parted ATTACH PARTITION sub_part1 for VALUES in (1);`,
			},
			{
				Statement: `CREATE TABLE sub_part2(b int, c int8, a numeric);`,
			},
			{
				Statement: `ALTER TABLE sub_parted ATTACH PARTITION sub_part2 for VALUES in (2);`,
			},
			{
				Statement: `CREATE TABLE list_part1(a numeric, b int, c int8);`,
			},
			{
				Statement: `ALTER TABLE list_parted ATTACH PARTITION list_part1 for VALUES in (2,3);`,
			},
			{
				Statement: `INSERT into list_parted VALUES (2,5,50);`,
			},
			{
				Statement: `INSERT into list_parted VALUES (3,6,60);`,
			},
			{
				Statement: `INSERT into sub_parted VALUES (1,1,60);`,
			},
			{
				Statement: `INSERT into sub_parted VALUES (1,2,10);`,
			},
			{
				Statement:   `UPDATE sub_parted set a = 2 WHERE c = 10;`,
				ErrorString: `new row for relation "sub_parted" violates partition constraint`,
			},
			{
				Statement: `SELECT tableoid::regclass::text, * FROM list_parted WHERE a = 2 ORDER BY 1;`,
				Results:   []sql.Row{{`list_part1`, 2, 5, 50}},
			},
			{
				Statement: `UPDATE list_parted set b = c + a WHERE a = 2;`,
			},
			{
				Statement: `SELECT tableoid::regclass::text, * FROM list_parted WHERE a = 2 ORDER BY 1;`,
				Results:   []sql.Row{{`list_part1`, 2, 52, 50}},
			},
			{
				Statement: `CREATE FUNCTION func_parted_mod_b() returns trigger as $$
BEGIN
   NEW.b = 2; -- This is changing partition key column.`,
			},
			{
				Statement: `   return NEW;`,
			},
			{
				Statement: `END $$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE TRIGGER parted_mod_b before update on sub_part1
   for each row execute procedure func_parted_mod_b();`,
			},
			{
				Statement: `SELECT tableoid::regclass::text, * FROM list_parted ORDER BY 1, 2, 3, 4;`,
				Results:   []sql.Row{{`list_part1`, 2, 52, 50}, {`list_part1`, 3, 6, 60}, {`sub_part1`, 1, 1, 60}, {`sub_part2`, 1, 2, 10}},
			},
			{
				Statement: `UPDATE list_parted set c = 70 WHERE b  = 1;`,
			},
			{
				Statement: `SELECT tableoid::regclass::text, * FROM list_parted ORDER BY 1, 2, 3, 4;`,
				Results:   []sql.Row{{`list_part1`, 2, 52, 50}, {`list_part1`, 3, 6, 60}, {`sub_part2`, 1, 2, 10}, {`sub_part2`, 1, 2, 70}},
			},
			{
				Statement: `DROP TRIGGER parted_mod_b ON sub_part1;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION func_parted_mod_b() returns trigger as $$
BEGIN
   raise notice 'Trigger: Got OLD row %, but returning NULL', OLD;`,
			},
			{
				Statement: `   return NULL;`,
			},
			{
				Statement: `END $$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE TRIGGER trig_skip_delete before delete on sub_part2
   for each row execute procedure func_parted_mod_b();`,
			},
			{
				Statement: `UPDATE list_parted set b = 1 WHERE c = 70;`,
			},
			{
				Statement: `SELECT tableoid::regclass::text, * FROM list_parted ORDER BY 1, 2, 3, 4;`,
				Results:   []sql.Row{{`list_part1`, 2, 52, 50}, {`list_part1`, 3, 6, 60}, {`sub_part2`, 1, 2, 10}, {`sub_part2`, 1, 2, 70}},
			},
			{
				Statement: `DROP TRIGGER trig_skip_delete ON sub_part2;`,
			},
			{
				Statement: `UPDATE list_parted set b = 1 WHERE c = 70;`,
			},
			{
				Statement: `SELECT tableoid::regclass::text, * FROM list_parted ORDER BY 1, 2, 3, 4;`,
				Results:   []sql.Row{{`list_part1`, 2, 52, 50}, {`list_part1`, 3, 6, 60}, {`sub_part1`, 1, 1, 70}, {`sub_part2`, 1, 2, 10}},
			},
			{
				Statement: `DROP FUNCTION func_parted_mod_b();`,
			},
			{
				Statement: `CREATE TABLE non_parted (id int);`,
			},
			{
				Statement: `INSERT into non_parted VALUES (1), (1), (1), (2), (2), (2), (3), (3), (3);`,
			},
			{
				Statement: `UPDATE list_parted t1 set a = 2 FROM non_parted t2 WHERE t1.a = t2.id and a = 1;`,
			},
			{
				Statement: `SELECT tableoid::regclass::text, * FROM list_parted ORDER BY 1, 2, 3, 4;`,
				Results:   []sql.Row{{`list_part1`, 2, 1, 70}, {`list_part1`, 2, 2, 10}, {`list_part1`, 2, 52, 50}, {`list_part1`, 3, 6, 60}},
			},
			{
				Statement: `DROP TABLE non_parted;`,
			},
			{
				Statement: `DROP TABLE list_parted;`,
			},
			{
				Statement: `create or replace function dummy_hashint4(a int4, seed int8) returns int8 as
$$ begin return (a + seed); end; $$ language 'plpgsql' immutable;`,
			},
			{
				Statement: `create operator class custom_opclass for type int4 using hash as
operator 1 = , function 2 dummy_hashint4(int4, int8);`,
			},
			{
				Statement: `create table hash_parted (
	a int,
	b int
) partition by hash (a custom_opclass, b custom_opclass);`,
			},
			{
				Statement: `create table hpart1 partition of hash_parted for values with (modulus 2, remainder 1);`,
			},
			{
				Statement: `create table hpart2 partition of hash_parted for values with (modulus 4, remainder 2);`,
			},
			{
				Statement: `create table hpart3 partition of hash_parted for values with (modulus 8, remainder 0);`,
			},
			{
				Statement: `create table hpart4 partition of hash_parted for values with (modulus 8, remainder 4);`,
			},
			{
				Statement: `insert into hpart1 values (1, 1);`,
			},
			{
				Statement: `insert into hpart2 values (2, 5);`,
			},
			{
				Statement: `insert into hpart4 values (3, 4);`,
			},
			{
				Statement:   `update hpart1 set a = 3, b=4 where a = 1;`,
				ErrorString: `new row for relation "hpart1" violates partition constraint`,
			},
			{
				Statement: `update hash_parted set b = b - 1 where b = 1;`,
			},
			{
				Statement: `update hash_parted set b = b + 8 where b = 1;`,
			},
			{
				Statement: `drop table hash_parted;`,
			},
			{
				Statement: `drop operator class custom_opclass using hash;`,
			},
			{
				Statement: `drop function dummy_hashint4(a int4, seed int8);`,
			},
		},
	})
}
