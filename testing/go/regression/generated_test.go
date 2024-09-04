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

func TestGenerated(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_generated)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_generated,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT attrelid, attname, attgenerated FROM pg_attribute WHERE attgenerated NOT IN ('', 's');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `CREATE TABLE gtest0 (a int PRIMARY KEY, b int GENERATED ALWAYS AS (55) STORED);`,
			},
			{
				Statement: `CREATE TABLE gtest1 (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED);`,
			},
			{
				Statement: `SELECT table_name, column_name, column_default, is_nullable, is_generated, generation_expression FROM information_schema.columns WHERE table_name LIKE 'gtest_' ORDER BY 1, 2;`,
				Results:   []sql.Row{{`gtest0`, `a`, ``, `NO`, `NEVER`, ``}, {`gtest0`, `b`, ``, `YES`, `ALWAYS`, 55}, {`gtest1`, `a`, ``, `NO`, `NEVER`, ``}, {`gtest1`, `b`, ``, `YES`, `ALWAYS`, `(a * 2)`}},
			},
			{
				Statement: `SELECT table_name, column_name, dependent_column FROM information_schema.column_column_usage ORDER BY 1, 2, 3;`,
				Results:   []sql.Row{{`gtest1`, `a`, `b`}},
			},
			{
				Statement: `\d gtest1
                            Table "public.gtest1"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           | not null | 
 b      | integer |           |          | generated always as (a * 2) stored
Indexes:
    "gtest1_pkey" PRIMARY KEY, btree (a)
CREATE TABLE gtest_err_1 (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED GENERATED ALWAYS AS (a * 3) STORED);`,
				ErrorString: `multiple generation clauses specified for column "b" of table "gtest_err_1"`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_2a (a int PRIMARY KEY, b int GENERATED ALWAYS AS (b * 2) STORED);`,
				ErrorString: `cannot use generated column "b" in column generation expression`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_2b (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED, c int GENERATED ALWAYS AS (b * 3) STORED);`,
				ErrorString: `cannot use generated column "b" in column generation expression`,
			},
			{
				Statement: `CREATE TABLE gtest_err_2c (a int PRIMARY KEY,
    b int GENERATED ALWAYS AS (num_nulls(gtest_err_2c)) STORED);`,
				ErrorString: `cannot use whole-row variable in column generation expression`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_3 (a int PRIMARY KEY, b int GENERATED ALWAYS AS (c * 2) STORED);`,
				ErrorString: `column "c" does not exist`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_4 (a int PRIMARY KEY, b double precision GENERATED ALWAYS AS (random()) STORED);`,
				ErrorString: `generation expression is not immutable`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_5a (a int PRIMARY KEY, b int DEFAULT 5 GENERATED ALWAYS AS (a * 2) STORED);`,
				ErrorString: `both default and generation expression specified for column "b" of table "gtest_err_5a"`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_5b (a int PRIMARY KEY, b int GENERATED ALWAYS AS identity GENERATED ALWAYS AS (a * 2) STORED);`,
				ErrorString: `both identity and generation expression specified for column "b" of table "gtest_err_5b"`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_6a (a int PRIMARY KEY, b bool GENERATED ALWAYS AS (xmin <> 37) STORED);`,
				ErrorString: `cannot use system column "xmin" in column generation expression`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_7a (a int PRIMARY KEY, b int GENERATED ALWAYS AS (avg(a)) STORED);`,
				ErrorString: `aggregate functions are not allowed in column generation expressions`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_7b (a int PRIMARY KEY, b int GENERATED ALWAYS AS (row_number() OVER (ORDER BY a)) STORED);`,
				ErrorString: `window functions are not allowed in column generation expressions`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_7c (a int PRIMARY KEY, b int GENERATED ALWAYS AS ((SELECT a)) STORED);`,
				ErrorString: `cannot use subquery in column generation expression`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_7d (a int PRIMARY KEY, b int GENERATED ALWAYS AS (generate_series(1, a)) STORED);`,
				ErrorString: `set-returning functions are not allowed in column generation expressions`,
			},
			{
				Statement:   `CREATE TABLE gtest_err_8 (a int PRIMARY KEY, b int GENERATED BY DEFAULT AS (a * 2) STORED);`,
				ErrorString: `for a generated column, GENERATED ALWAYS must be specified`,
			},
			{
				Statement: `INSERT INTO gtest1 VALUES (1);`,
			},
			{
				Statement: `INSERT INTO gtest1 VALUES (2, DEFAULT);  -- ok`,
			},
			{
				Statement:   `INSERT INTO gtest1 VALUES (3, 33);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement:   `INSERT INTO gtest1 VALUES (3, 33), (4, 44);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement:   `INSERT INTO gtest1 VALUES (3, DEFAULT), (4, 44);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement:   `INSERT INTO gtest1 VALUES (3, 33), (4, DEFAULT);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement: `INSERT INTO gtest1 VALUES (3, DEFAULT), (4, DEFAULT);  -- ok`,
			},
			{
				Statement: `SELECT * FROM gtest1 ORDER BY a;`,
				Results:   []sql.Row{{1, 2}, {2, 4}, {3, 6}, {4, 8}},
			},
			{
				Statement: `DELETE FROM gtest1 WHERE a >= 3;`,
			},
			{
				Statement: `UPDATE gtest1 SET b = DEFAULT WHERE a = 1;`,
			},
			{
				Statement:   `UPDATE gtest1 SET b = 11 WHERE a = 1;  -- error`,
				ErrorString: `column "b" can only be updated to DEFAULT`,
			},
			{
				Statement: `SELECT * FROM gtest1 ORDER BY a;`,
				Results:   []sql.Row{{1, 2}, {2, 4}},
			},
			{
				Statement: `SELECT a, b, b * 2 AS b2 FROM gtest1 ORDER BY a;`,
				Results:   []sql.Row{{1, 2, 4}, {2, 4, 8}},
			},
			{
				Statement: `SELECT a, b FROM gtest1 WHERE b = 4 ORDER BY a;`,
				Results:   []sql.Row{{2, 4}},
			},
			{
				Statement:   `INSERT INTO gtest1 VALUES (2000000000);`,
				ErrorString: `integer out of range`,
			},
			{
				Statement: `SELECT * FROM gtest1;`,
				Results:   []sql.Row{{2, 4}, {1, 2}},
			},
			{
				Statement: `DELETE FROM gtest1 WHERE a = 2000000000;`,
			},
			{
				Statement: `CREATE TABLE gtestx (x int, y int);`,
			},
			{
				Statement: `INSERT INTO gtestx VALUES (11, 1), (22, 2), (33, 3);`,
			},
			{
				Statement: `SELECT * FROM gtestx, gtest1 WHERE gtestx.y = gtest1.a;`,
				Results:   []sql.Row{{11, 1, 1, 2}, {22, 2, 2, 4}},
			},
			{
				Statement: `DROP TABLE gtestx;`,
			},
			{
				Statement: `SELECT * FROM gtest1 ORDER BY a;`,
				Results:   []sql.Row{{1, 2}, {2, 4}},
			},
			{
				Statement: `UPDATE gtest1 SET a = 3 WHERE b = 4;`,
			},
			{
				Statement: `SELECT * FROM gtest1 ORDER BY a;`,
				Results:   []sql.Row{{1, 2}, {3, 6}},
			},
			{
				Statement: `DELETE FROM gtest1 WHERE b = 2;`,
			},
			{
				Statement: `SELECT * FROM gtest1 ORDER BY a;`,
				Results:   []sql.Row{{3, 6}},
			},
			{
				Statement: `CREATE TABLE gtestm (
  id int PRIMARY KEY,
  f1 int,
  f2 int,
  f3 int GENERATED ALWAYS AS (f1 * 2) STORED,
  f4 int GENERATED ALWAYS AS (f2 * 2) STORED
);`,
			},
			{
				Statement: `INSERT INTO gtestm VALUES (1, 5, 100);`,
			},
			{
				Statement: `MERGE INTO gtestm t USING (VALUES (1, 10), (2, 20)) v(id, f1) ON t.id = v.id
  WHEN MATCHED THEN UPDATE SET f1 = v.f1
  WHEN NOT MATCHED THEN INSERT VALUES (v.id, v.f1, 200);`,
			},
			{
				Statement: `SELECT * FROM gtestm ORDER BY id;`,
				Results:   []sql.Row{{1, 10, 100, 20, 200}, {2, 20, 200, 40, 400}},
			},
			{
				Statement: `DROP TABLE gtestm;`,
			},
			{
				Statement: `CREATE VIEW gtest1v AS SELECT * FROM gtest1;`,
			},
			{
				Statement: `SELECT * FROM gtest1v;`,
				Results:   []sql.Row{{3, 6}},
			},
			{
				Statement:   `INSERT INTO gtest1v VALUES (4, 8);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement: `INSERT INTO gtest1v VALUES (5, DEFAULT);  -- ok`,
			},
			{
				Statement:   `INSERT INTO gtest1v VALUES (6, 66), (7, 77);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement:   `INSERT INTO gtest1v VALUES (6, DEFAULT), (7, 77);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement:   `INSERT INTO gtest1v VALUES (6, 66), (7, DEFAULT);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement: `INSERT INTO gtest1v VALUES (6, DEFAULT), (7, DEFAULT);  -- ok`,
			},
			{
				Statement: `ALTER VIEW gtest1v ALTER COLUMN b SET DEFAULT 100;`,
			},
			{
				Statement:   `INSERT INTO gtest1v VALUES (8, DEFAULT);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement:   `INSERT INTO gtest1v VALUES (8, DEFAULT), (9, DEFAULT);  -- error`,
				ErrorString: `cannot insert a non-DEFAULT value into column "b"`,
			},
			{
				Statement: `SELECT * FROM gtest1v;`,
				Results:   []sql.Row{{3, 6}, {5, 10}, {6, 12}, {7, 14}},
			},
			{
				Statement: `DELETE FROM gtest1v WHERE a >= 5;`,
			},
			{
				Statement: `DROP VIEW gtest1v;`,
			},
			{
				Statement: `WITH foo AS (SELECT * FROM gtest1) SELECT * FROM foo;`,
				Results:   []sql.Row{{3, 6}},
			},
			{
				Statement: `CREATE TABLE gtest1_1 () INHERITS (gtest1);`,
			},
			{
				Statement: `SELECT * FROM gtest1_1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `\d gtest1_1
                           Table "public.gtest1_1"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           | not null | 
 b      | integer |           |          | generated always as (a * 2) stored
Inherits: gtest1
INSERT INTO gtest1_1 VALUES (4);`,
			},
			{
				Statement: `SELECT * FROM gtest1_1;`,
				Results:   []sql.Row{{4, 8}},
			},
			{
				Statement: `SELECT * FROM gtest1;`,
				Results:   []sql.Row{{3, 6}, {4, 8}},
			},
			{
				Statement: `CREATE TABLE gtest_normal (a int, b int);`,
			},
			{
				Statement: `CREATE TABLE gtest_normal_child (a int, b int GENERATED ALWAYS AS (a * 2) STORED) INHERITS (gtest_normal);`,
			},
			{
				Statement: `\d gtest_normal_child
                      Table "public.gtest_normal_child"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           |          | 
 b      | integer |           |          | generated always as (a * 2) stored
Inherits: gtest_normal
INSERT INTO gtest_normal (a) VALUES (1);`,
			},
			{
				Statement: `INSERT INTO gtest_normal_child (a) VALUES (2);`,
			},
			{
				Statement: `SELECT * FROM gtest_normal;`,
				Results:   []sql.Row{{1, ``}, {2, 4}},
			},
			{
				Statement: `CREATE TABLE gtest_normal_child2 (a int, b int GENERATED ALWAYS AS (a * 3) STORED);`,
			},
			{
				Statement: `ALTER TABLE gtest_normal_child2 INHERIT gtest_normal;`,
			},
			{
				Statement: `INSERT INTO gtest_normal_child2 (a) VALUES (3);`,
			},
			{
				Statement: `SELECT * FROM gtest_normal;`,
				Results:   []sql.Row{{1, ``}, {2, 4}, {3, 9}},
			},
			{
				Statement:   `CREATE TABLE gtestx (x int, b int GENERATED ALWAYS AS (a * 22) STORED) INHERITS (gtest1);  -- error`,
				ErrorString: `child column "b" specifies generation expression`,
			},
			{
				Statement:   `CREATE TABLE gtestx (x int, b int DEFAULT 10) INHERITS (gtest1);  -- error`,
				ErrorString: `column "b" inherits from generated column but specifies default`,
			},
			{
				Statement:   `CREATE TABLE gtestx (x int, b int GENERATED ALWAYS AS IDENTITY) INHERITS (gtest1);  -- error`,
				ErrorString: `column "b" inherits from generated column but specifies identity`,
			},
			{
				Statement: `CREATE TABLE gtestxx_1 (a int NOT NULL, b int);`,
			},
			{
				Statement:   `ALTER TABLE gtestxx_1 INHERIT gtest1;  -- error`,
				ErrorString: `column "b" in child table must be a generated column`,
			},
			{
				Statement: `CREATE TABLE gtestxx_2 (a int NOT NULL, b int GENERATED ALWAYS AS (a * 22) STORED);`,
			},
			{
				Statement:   `ALTER TABLE gtestxx_2 INHERIT gtest1;  -- error`,
				ErrorString: `column "b" in child table has a conflicting generation expression`,
			},
			{
				Statement: `CREATE TABLE gtestxx_3 (a int NOT NULL, b int GENERATED ALWAYS AS (a * 2) STORED);`,
			},
			{
				Statement: `ALTER TABLE gtestxx_3 INHERIT gtest1;  -- ok`,
			},
			{
				Statement: `CREATE TABLE gtestxx_4 (b int GENERATED ALWAYS AS (a * 2) STORED, a int NOT NULL);`,
			},
			{
				Statement: `ALTER TABLE gtestxx_4 INHERIT gtest1;  -- ok`,
			},
			{
				Statement: `CREATE TABLE gtesty (x int, b int);`,
			},
			{
				Statement:   `CREATE TABLE gtest1_2 () INHERITS (gtest1, gtesty);  -- error`,
				ErrorString: `inherited column "b" has a generation conflict`,
			},
			{
				Statement: `DROP TABLE gtesty;`,
			},
			{
				Statement: `CREATE TABLE gtesty (x int, b int GENERATED ALWAYS AS (x * 22) STORED);`,
			},
			{
				Statement:   `CREATE TABLE gtest1_2 () INHERITS (gtest1, gtesty);  -- error`,
				ErrorString: `column "b" inherits conflicting generation expressions`,
			},
			{
				Statement: `DROP TABLE gtesty;`,
			},
			{
				Statement: `CREATE TABLE gtesty (x int, b int DEFAULT 55);`,
			},
			{
				Statement:   `CREATE TABLE gtest1_2 () INHERITS (gtest0, gtesty);  -- error`,
				ErrorString: `inherited column "b" has a generation conflict`,
			},
			{
				Statement: `DROP TABLE gtesty;`,
			},
			{
				Statement: `CREATE TABLE gtestp (f1 int);`,
			},
			{
				Statement: `CREATE TABLE gtestc (f2 int GENERATED ALWAYS AS (f1+1) STORED) INHERITS(gtestp);`,
			},
			{
				Statement: `INSERT INTO gtestc values(42);`,
			},
			{
				Statement: `TABLE gtestc;`,
				Results:   []sql.Row{{42, 43}},
			},
			{
				Statement: `UPDATE gtestp SET f1 = f1 * 10;`,
			},
			{
				Statement: `TABLE gtestc;`,
				Results:   []sql.Row{{420, 421}},
			},
			{
				Statement: `DROP TABLE gtestp CASCADE;`,
			},
			{
				Statement: `CREATE TABLE gtest3 (a int, b int GENERATED ALWAYS AS (a * 3) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest3 (a) VALUES (1), (2), (3), (NULL);`,
			},
			{
				Statement: `SELECT * FROM gtest3 ORDER BY a;`,
				Results:   []sql.Row{{1, 3}, {2, 6}, {3, 9}, {``, ``}},
			},
			{
				Statement: `UPDATE gtest3 SET a = 22 WHERE a = 2;`,
			},
			{
				Statement: `SELECT * FROM gtest3 ORDER BY a;`,
				Results:   []sql.Row{{1, 3}, {3, 9}, {22, 66}, {``, ``}},
			},
			{
				Statement: `CREATE TABLE gtest3a (a text, b text GENERATED ALWAYS AS (a || '+' || a) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest3a (a) VALUES ('a'), ('b'), ('c'), (NULL);`,
			},
			{
				Statement: `SELECT * FROM gtest3a ORDER BY a;`,
				Results:   []sql.Row{{`a`, `a+a`}, {`b`, `b+b`}, {`c`, `c+c`}, {``, ``}},
			},
			{
				Statement: `UPDATE gtest3a SET a = 'bb' WHERE a = 'b';`,
			},
			{
				Statement: `SELECT * FROM gtest3a ORDER BY a;`,
				Results:   []sql.Row{{`a`, `a+a`}, {`bb`, `bb+bb`}, {`c`, `c+c`}, {``, ``}},
			},
			{
				Statement: `TRUNCATE gtest1;`,
			},
			{
				Statement: `INSERT INTO gtest1 (a) VALUES (1), (2);`,
			},
			{
				Statement: `COPY gtest1 TO stdout;`,
			},
			{
				Statement: `1
2
COPY gtest1 (a, b) TO stdout;`,
				ErrorString: `column "b" is a generated column`,
			},
			{
				Statement: `COPY gtest1 FROM stdin;`,
			},
			{
				Statement:   `COPY gtest1 (a, b) FROM stdin;`,
				ErrorString: `column "b" is a generated column`,
			},
			{
				Statement: `SELECT * FROM gtest1 ORDER BY a;`,
				Results:   []sql.Row{{1, 2}, {2, 4}, {3, 6}, {4, 8}},
			},
			{
				Statement: `TRUNCATE gtest3;`,
			},
			{
				Statement: `INSERT INTO gtest3 (a) VALUES (1), (2);`,
			},
			{
				Statement: `COPY gtest3 TO stdout;`,
			},
			{
				Statement: `1
2
COPY gtest3 (a, b) TO stdout;`,
				ErrorString: `column "b" is a generated column`,
			},
			{
				Statement: `COPY gtest3 FROM stdin;`,
			},
			{
				Statement:   `COPY gtest3 (a, b) FROM stdin;`,
				ErrorString: `column "b" is a generated column`,
			},
			{
				Statement: `SELECT * FROM gtest3 ORDER BY a;`,
				Results:   []sql.Row{{1, 3}, {2, 6}, {3, 9}, {4, 12}},
			},
			{
				Statement: `CREATE TABLE gtest2 (a int PRIMARY KEY, b int GENERATED ALWAYS AS (NULL) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest2 VALUES (1);`,
			},
			{
				Statement: `SELECT * FROM gtest2;`,
				Results:   []sql.Row{{1, ``}},
			},
			{
				Statement: `CREATE TABLE gtest_varlena (a varchar, b varchar GENERATED ALWAYS AS (a) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest_varlena (a) VALUES('01234567890123456789');`,
			},
			{
				Statement: `INSERT INTO gtest_varlena (a) VALUES(NULL);`,
			},
			{
				Statement: `SELECT * FROM gtest_varlena ORDER BY a;`,
				Results:   []sql.Row{{"01234567890123456789", "01234567890123456789"}, {``, ``}},
			},
			{
				Statement: `DROP TABLE gtest_varlena;`,
			},
			{
				Statement: `CREATE TYPE double_int as (a int, b int);`,
			},
			{
				Statement: `CREATE TABLE gtest4 (
    a int,
    b double_int GENERATED ALWAYS AS ((a * 2, a * 3)) STORED
);`,
			},
			{
				Statement: `INSERT INTO gtest4 VALUES (1), (6);`,
			},
			{
				Statement: `SELECT * FROM gtest4;`,
				Results:   []sql.Row{{1, `(2,3)`}, {6, `(12,18)`}},
			},
			{
				Statement: `DROP TABLE gtest4;`,
			},
			{
				Statement: `DROP TYPE double_int;`,
			},
			{
				Statement: `CREATE TABLE gtest_tableoid (
  a int PRIMARY KEY,
  b bool GENERATED ALWAYS AS (tableoid = 'gtest_tableoid'::regclass) STORED
);`,
			},
			{
				Statement: `INSERT INTO gtest_tableoid VALUES (1), (2);`,
			},
			{
				Statement: `ALTER TABLE gtest_tableoid ADD COLUMN
  c regclass GENERATED ALWAYS AS (tableoid) STORED;`,
			},
			{
				Statement: `SELECT * FROM gtest_tableoid;`,
				Results:   []sql.Row{{1, true, `gtest_tableoid`}, {2, true, `gtest_tableoid`}},
			},
			{
				Statement: `CREATE TABLE gtest10 (a int PRIMARY KEY, b int, c int GENERATED ALWAYS AS (b * 2) STORED);`,
			},
			{
				Statement:   `ALTER TABLE gtest10 DROP COLUMN b;  -- fails`,
				ErrorString: `cannot drop column b of table gtest10 because other objects depend on it`,
			},
			{
				Statement: `ALTER TABLE gtest10 DROP COLUMN b CASCADE;  -- drops c too`,
			},
			{
				Statement: `\d gtest10
              Table "public.gtest10"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
Indexes:
    "gtest10_pkey" PRIMARY KEY, btree (a)
CREATE TABLE gtest10a (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED);`,
			},
			{
				Statement: `ALTER TABLE gtest10a DROP COLUMN b;`,
			},
			{
				Statement: `INSERT INTO gtest10a (a) VALUES (1);`,
			},
			{
				Statement: `CREATE USER regress_user11;`,
			},
			{
				Statement: `CREATE TABLE gtest11s (a int PRIMARY KEY, b int, c int GENERATED ALWAYS AS (b * 2) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest11s VALUES (1, 10), (2, 20);`,
			},
			{
				Statement: `GRANT SELECT (a, c) ON gtest11s TO regress_user11;`,
			},
			{
				Statement: `CREATE FUNCTION gf1(a int) RETURNS int AS $$ SELECT a * 3 $$ IMMUTABLE LANGUAGE SQL;`,
			},
			{
				Statement: `REVOKE ALL ON FUNCTION gf1(int) FROM PUBLIC;`,
			},
			{
				Statement: `CREATE TABLE gtest12s (a int PRIMARY KEY, b int, c int GENERATED ALWAYS AS (gf1(b)) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest12s VALUES (1, 10), (2, 20);`,
			},
			{
				Statement: `GRANT SELECT (a, c) ON gtest12s TO regress_user11;`,
			},
			{
				Statement: `SET ROLE regress_user11;`,
			},
			{
				Statement:   `SELECT a, b FROM gtest11s;  -- not allowed`,
				ErrorString: `permission denied for table gtest11s`,
			},
			{
				Statement: `SELECT a, c FROM gtest11s;  -- allowed`,
				Results:   []sql.Row{{1, 20}, {2, 40}},
			},
			{
				Statement:   `SELECT gf1(10);  -- not allowed`,
				ErrorString: `permission denied for function gf1`,
			},
			{
				Statement: `SELECT a, c FROM gtest12s;  -- allowed`,
				Results:   []sql.Row{{1, 30}, {2, 60}},
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement:   `DROP FUNCTION gf1(int);  -- fail`,
				ErrorString: `cannot drop function gf1(integer) because other objects depend on it`,
			},
			{
				Statement: `DROP TABLE gtest11s, gtest12s;`,
			},
			{
				Statement: `DROP FUNCTION gf1(int);`,
			},
			{
				Statement: `DROP USER regress_user11;`,
			},
			{
				Statement: `CREATE TABLE gtest20 (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED CHECK (b < 50));`,
			},
			{
				Statement: `INSERT INTO gtest20 (a) VALUES (10);  -- ok`,
			},
			{
				Statement:   `INSERT INTO gtest20 (a) VALUES (30);  -- violates constraint`,
				ErrorString: `new row for relation "gtest20" violates check constraint "gtest20_b_check"`,
			},
			{
				Statement: `CREATE TABLE gtest20a (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest20a (a) VALUES (10);`,
			},
			{
				Statement: `INSERT INTO gtest20a (a) VALUES (30);`,
			},
			{
				Statement:   `ALTER TABLE gtest20a ADD CHECK (b < 50);  -- fails on existing row`,
				ErrorString: `check constraint "gtest20a_b_check" of relation "gtest20a" is violated by some row`,
			},
			{
				Statement: `CREATE TABLE gtest20b (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest20b (a) VALUES (10);`,
			},
			{
				Statement: `INSERT INTO gtest20b (a) VALUES (30);`,
			},
			{
				Statement: `ALTER TABLE gtest20b ADD CONSTRAINT chk CHECK (b < 50) NOT VALID;`,
			},
			{
				Statement:   `ALTER TABLE gtest20b VALIDATE CONSTRAINT chk;  -- fails on existing row`,
				ErrorString: `check constraint "chk" of relation "gtest20b" is violated by some row`,
			},
			{
				Statement: `CREATE TABLE gtest21a (a int PRIMARY KEY, b int GENERATED ALWAYS AS (nullif(a, 0)) STORED NOT NULL);`,
			},
			{
				Statement: `INSERT INTO gtest21a (a) VALUES (1);  -- ok`,
			},
			{
				Statement:   `INSERT INTO gtest21a (a) VALUES (0);  -- violates constraint`,
				ErrorString: `null value in column "b" of relation "gtest21a" violates not-null constraint`,
			},
			{
				Statement: `CREATE TABLE gtest21b (a int PRIMARY KEY, b int GENERATED ALWAYS AS (nullif(a, 0)) STORED);`,
			},
			{
				Statement: `ALTER TABLE gtest21b ALTER COLUMN b SET NOT NULL;`,
			},
			{
				Statement: `INSERT INTO gtest21b (a) VALUES (1);  -- ok`,
			},
			{
				Statement:   `INSERT INTO gtest21b (a) VALUES (0);  -- violates constraint`,
				ErrorString: `null value in column "b" of relation "gtest21b" violates not-null constraint`,
			},
			{
				Statement: `ALTER TABLE gtest21b ALTER COLUMN b DROP NOT NULL;`,
			},
			{
				Statement: `INSERT INTO gtest21b (a) VALUES (0);  -- ok now`,
			},
			{
				Statement: `CREATE TABLE gtest22a (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a / 2) STORED UNIQUE);`,
			},
			{
				Statement: `INSERT INTO gtest22a VALUES (2);`,
			},
			{
				Statement:   `INSERT INTO gtest22a VALUES (3);`,
				ErrorString: `duplicate key value violates unique constraint "gtest22a_b_key"`,
			},
			{
				Statement: `INSERT INTO gtest22a VALUES (4);`,
			},
			{
				Statement: `CREATE TABLE gtest22b (a int, b int GENERATED ALWAYS AS (a / 2) STORED, PRIMARY KEY (a, b));`,
			},
			{
				Statement: `INSERT INTO gtest22b VALUES (2);`,
			},
			{
				Statement:   `INSERT INTO gtest22b VALUES (2);`,
				ErrorString: `duplicate key value violates unique constraint "gtest22b_pkey"`,
			},
			{
				Statement: `CREATE TABLE gtest22c (a int, b int GENERATED ALWAYS AS (a * 2) STORED);`,
			},
			{
				Statement: `CREATE INDEX gtest22c_b_idx ON gtest22c (b);`,
			},
			{
				Statement: `CREATE INDEX gtest22c_expr_idx ON gtest22c ((b * 3));`,
			},
			{
				Statement: `CREATE INDEX gtest22c_pred_idx ON gtest22c (a) WHERE b > 0;`,
			},
			{
				Statement: `\d gtest22c
                           Table "public.gtest22c"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           |          | 
 b      | integer |           |          | generated always as (a * 2) stored
Indexes:
    "gtest22c_b_idx" btree (b)
    "gtest22c_expr_idx" btree ((b * 3))
    "gtest22c_pred_idx" btree (a) WHERE b > 0
INSERT INTO gtest22c VALUES (1), (2), (3);`,
			},
			{
				Statement: `SET enable_seqscan TO off;`,
			},
			{
				Statement: `SET enable_bitmapscan TO off;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM gtest22c WHERE b = 4;`,
				Results:   []sql.Row{{`Index Scan using gtest22c_b_idx on gtest22c`}, {`Index Cond: (b = 4)`}},
			},
			{
				Statement: `SELECT * FROM gtest22c WHERE b = 4;`,
				Results:   []sql.Row{{2, 4}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM gtest22c WHERE b * 3 = 6;`,
				Results:   []sql.Row{{`Index Scan using gtest22c_expr_idx on gtest22c`}, {`Index Cond: ((b * 3) = 6)`}},
			},
			{
				Statement: `SELECT * FROM gtest22c WHERE b * 3 = 6;`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM gtest22c WHERE a = 1 AND b > 0;`,
				Results:   []sql.Row{{`Index Scan using gtest22c_pred_idx on gtest22c`}, {`Index Cond: (a = 1)`}},
			},
			{
				Statement: `SELECT * FROM gtest22c WHERE a = 1 AND b > 0;`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
			{
				Statement: `CREATE TABLE gtest23a (x int PRIMARY KEY, y int);`,
			},
			{
				Statement: `INSERT INTO gtest23a VALUES (1, 11), (2, 22), (3, 33);`,
			},
			{
				Statement:   `CREATE TABLE gtest23x (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED REFERENCES gtest23a (x) ON UPDATE CASCADE);  -- error`,
				ErrorString: `invalid ON UPDATE action for foreign key constraint containing generated column`,
			},
			{
				Statement:   `CREATE TABLE gtest23x (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED REFERENCES gtest23a (x) ON DELETE SET NULL);  -- error`,
				ErrorString: `invalid ON DELETE action for foreign key constraint containing generated column`,
			},
			{
				Statement: `CREATE TABLE gtest23b (a int PRIMARY KEY, b int GENERATED ALWAYS AS (a * 2) STORED REFERENCES gtest23a (x));`,
			},
			{
				Statement: `\d gtest23b
                           Table "public.gtest23b"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           | not null | 
 b      | integer |           |          | generated always as (a * 2) stored
Indexes:
    "gtest23b_pkey" PRIMARY KEY, btree (a)
Foreign-key constraints:
    "gtest23b_b_fkey" FOREIGN KEY (b) REFERENCES gtest23a(x)
INSERT INTO gtest23b VALUES (1);  -- ok`,
			},
			{
				Statement:   `INSERT INTO gtest23b VALUES (5);  -- error`,
				ErrorString: `insert or update on table "gtest23b" violates foreign key constraint "gtest23b_b_fkey"`,
			},
			{
				Statement: `DROP TABLE gtest23b;`,
			},
			{
				Statement: `DROP TABLE gtest23a;`,
			},
			{
				Statement: `CREATE TABLE gtest23p (x int, y int GENERATED ALWAYS AS (x * 2) STORED, PRIMARY KEY (y));`,
			},
			{
				Statement: `INSERT INTO gtest23p VALUES (1), (2), (3);`,
			},
			{
				Statement: `CREATE TABLE gtest23q (a int PRIMARY KEY, b int REFERENCES gtest23p (y));`,
			},
			{
				Statement: `INSERT INTO gtest23q VALUES (1, 2);  -- ok`,
			},
			{
				Statement:   `INSERT INTO gtest23q VALUES (2, 5);  -- error`,
				ErrorString: `insert or update on table "gtest23q" violates foreign key constraint "gtest23q_b_fkey"`,
			},
			{
				Statement: `CREATE DOMAIN gtestdomain1 AS int CHECK (VALUE < 10);`,
			},
			{
				Statement: `CREATE TABLE gtest24 (a int PRIMARY KEY, b gtestdomain1 GENERATED ALWAYS AS (a * 2) STORED);`,
			},
			{
				Statement: `INSERT INTO gtest24 (a) VALUES (4);  -- ok`,
			},
			{
				Statement:   `INSERT INTO gtest24 (a) VALUES (6);  -- error`,
				ErrorString: `value for domain gtestdomain1 violates check constraint "gtestdomain1_check"`,
			},
			{
				Statement: `CREATE TYPE gtest_type AS (f1 integer, f2 text, f3 bigint);`,
			},
			{
				Statement:   `CREATE TABLE gtest28 OF gtest_type (f1 WITH OPTIONS GENERATED ALWAYS AS (f2 *2) STORED);`,
				ErrorString: `generated columns are not supported on typed tables`,
			},
			{
				Statement: `DROP TYPE gtest_type CASCADE;`,
			},
			{
				Statement: `CREATE TABLE gtest_parent (f1 date NOT NULL, f2 text, f3 bigint) PARTITION BY RANGE (f1);`,
			},
			{
				Statement: `CREATE TABLE gtest_child PARTITION OF gtest_parent (
    f3 WITH OPTIONS GENERATED ALWAYS AS (f2 * 2) STORED
) FOR VALUES FROM ('2016-07-01') TO ('2016-08-01'); -- error`,
				ErrorString: `generated columns are not supported on partitions`,
			},
			{
				Statement: `DROP TABLE gtest_parent;`,
			},
			{
				Statement: `CREATE TABLE gtest_parent (f1 date NOT NULL, f2 bigint, f3 bigint GENERATED ALWAYS AS (f2 * 2) STORED) PARTITION BY RANGE (f1);`,
			},
			{
				Statement: `CREATE TABLE gtest_child PARTITION OF gtest_parent FOR VALUES FROM ('2016-07-01') TO ('2016-08-01');`,
			},
			{
				Statement: `CREATE TABLE gtest_child3 PARTITION OF gtest_parent FOR VALUES FROM ('2016-09-01') TO ('2016-10-01');`,
			},
			{
				Statement: `INSERT INTO gtest_parent (f1, f2) VALUES ('2016-07-15', 1);`,
			},
			{
				Statement: `SELECT * FROM gtest_parent;`,
				Results:   []sql.Row{{`07-15-2016`, 1, 2}},
			},
			{
				Statement: `SELECT * FROM gtest_child;`,
				Results:   []sql.Row{{`07-15-2016`, 1, 2}},
			},
			{
				Statement: `UPDATE gtest_parent SET f1 = f1 + 60, f2 = f2 + 1;`,
			},
			{
				Statement: `SELECT * FROM gtest_parent;`,
				Results:   []sql.Row{{`09-13-2016`, 2, 4}},
			},
			{
				Statement: `SELECT * FROM gtest_child3;`,
				Results:   []sql.Row{{`09-13-2016`, 2, 4}},
			},
			{
				Statement: `DROP TABLE gtest_parent;`,
			},
			{
				Statement:   `CREATE TABLE gtest_parent (f1 date NOT NULL, f2 bigint, f3 bigint GENERATED ALWAYS AS (f2 * 2) STORED) PARTITION BY RANGE (f3);`,
				ErrorString: `cannot use generated column in partition key`,
			},
			{
				Statement:   `CREATE TABLE gtest_parent (f1 date NOT NULL, f2 bigint, f3 bigint GENERATED ALWAYS AS (f2 * 2) STORED) PARTITION BY RANGE ((f3 * 3));`,
				ErrorString: `cannot use generated column in partition key`,
			},
			{
				Statement: `CREATE TABLE gtest25 (a int PRIMARY KEY);`,
			},
			{
				Statement: `INSERT INTO gtest25 VALUES (3), (4);`,
			},
			{
				Statement: `ALTER TABLE gtest25 ADD COLUMN b int GENERATED ALWAYS AS (a * 3) STORED;`,
			},
			{
				Statement: `SELECT * FROM gtest25 ORDER BY a;`,
				Results:   []sql.Row{{3, 9}, {4, 12}},
			},
			{
				Statement:   `ALTER TABLE gtest25 ADD COLUMN x int GENERATED ALWAYS AS (b * 4) STORED;  -- error`,
				ErrorString: `cannot use generated column "b" in column generation expression`,
			},
			{
				Statement:   `ALTER TABLE gtest25 ADD COLUMN x int GENERATED ALWAYS AS (z * 4) STORED;  -- error`,
				ErrorString: `column "z" does not exist`,
			},
			{
				Statement: `ALTER TABLE gtest25 ADD COLUMN c int DEFAULT 42,
  ADD COLUMN x int GENERATED ALWAYS AS (c * 4) STORED;`,
			},
			{
				Statement: `ALTER TABLE gtest25 ADD COLUMN d int DEFAULT 101;`,
			},
			{
				Statement: `ALTER TABLE gtest25 ALTER COLUMN d SET DATA TYPE float8,
  ADD COLUMN y float8 GENERATED ALWAYS AS (d * 4) STORED;`,
			},
			{
				Statement: `SELECT * FROM gtest25 ORDER BY a;`,
				Results:   []sql.Row{{3, 9, 42, 168, 101, 404}, {4, 12, 42, 168, 101, 404}},
			},
			{
				Statement: `\d gtest25
                                         Table "public.gtest25"
 Column |       Type       | Collation | Nullable |                       Default                        
--------+------------------+-----------+----------+------------------------------------------------------
 a      | integer          |           | not null | 
 b      | integer          |           |          | generated always as (a * 3) stored
 c      | integer          |           |          | 42
 x      | integer          |           |          | generated always as (c * 4) stored
 d      | double precision |           |          | 101
 y      | double precision |           |          | generated always as (d * 4::double precision) stored
Indexes:
    "gtest25_pkey" PRIMARY KEY, btree (a)
CREATE TABLE gtest27 (
    a int,
    b int,
    x int GENERATED ALWAYS AS ((a + b) * 2) STORED
);`,
			},
			{
				Statement: `INSERT INTO gtest27 (a, b) VALUES (3, 7), (4, 11);`,
			},
			{
				Statement:   `ALTER TABLE gtest27 ALTER COLUMN a TYPE text;  -- error`,
				ErrorString: `cannot alter type of a column used by a generated column`,
			},
			{
				Statement: `ALTER TABLE gtest27 ALTER COLUMN x TYPE numeric;`,
			},
			{
				Statement: `\d gtest27
                                Table "public.gtest27"
 Column |  Type   | Collation | Nullable |                  Default                   
--------+---------+-----------+----------+--------------------------------------------
 a      | integer |           |          | 
 b      | integer |           |          | 
 x      | numeric |           |          | generated always as (((a + b) * 2)) stored
SELECT * FROM gtest27;`,
				Results: []sql.Row{{3, 7, 20}, {4, 11, 30}},
			},
			{
				Statement:   `ALTER TABLE gtest27 ALTER COLUMN x TYPE boolean USING x <> 0;  -- error`,
				ErrorString: `generation expression for column "x" cannot be cast automatically to type boolean`,
			},
			{
				Statement:   `ALTER TABLE gtest27 ALTER COLUMN x DROP DEFAULT;  -- error`,
				ErrorString: `column "x" of relation "gtest27" is a generated column`,
			},
			{
				Statement: `ALTER TABLE gtest27
  DROP COLUMN x,
  ALTER COLUMN a TYPE bigint,
  ALTER COLUMN b TYPE bigint,
  ADD COLUMN x bigint GENERATED ALWAYS AS ((a + b) * 2) STORED;`,
			},
			{
				Statement: `\d gtest27
                              Table "public.gtest27"
 Column |  Type  | Collation | Nullable |                 Default                  
--------+--------+-----------+----------+------------------------------------------
 a      | bigint |           |          | 
 b      | bigint |           |          | 
 x      | bigint |           |          | generated always as ((a + b) * 2) stored
ALTER TABLE gtest27
  ALTER COLUMN a TYPE float8,
  ALTER COLUMN b TYPE float8;  -- error`,
				ErrorString: `cannot alter type of a column used by a generated column`,
			},
			{
				Statement: `\d gtest27
                              Table "public.gtest27"
 Column |  Type  | Collation | Nullable |                 Default                  
--------+--------+-----------+----------+------------------------------------------
 a      | bigint |           |          | 
 b      | bigint |           |          | 
 x      | bigint |           |          | generated always as ((a + b) * 2) stored
SELECT * FROM gtest27;`,
				Results: []sql.Row{{3, 7, 20}, {4, 11, 30}},
			},
			{
				Statement: `CREATE TABLE gtest29 (
    a int,
    b int GENERATED ALWAYS AS (a * 2) STORED
);`,
			},
			{
				Statement: `INSERT INTO gtest29 (a) VALUES (3), (4);`,
			},
			{
				Statement:   `ALTER TABLE gtest29 ALTER COLUMN a DROP EXPRESSION;  -- error`,
				ErrorString: `column "a" of relation "gtest29" is not a stored generated column`,
			},
			{
				Statement: `ALTER TABLE gtest29 ALTER COLUMN a DROP EXPRESSION IF EXISTS;  -- notice`,
			},
			{
				Statement: `ALTER TABLE gtest29 ALTER COLUMN b DROP EXPRESSION;`,
			},
			{
				Statement: `INSERT INTO gtest29 (a) VALUES (5);`,
			},
			{
				Statement: `INSERT INTO gtest29 (a, b) VALUES (6, 66);`,
			},
			{
				Statement: `SELECT * FROM gtest29;`,
				Results:   []sql.Row{{3, 6}, {4, 8}, {5, ``}, {6, 66}},
			},
			{
				Statement: `\d gtest29
              Table "public.gtest29"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
ALTER TABLE gtest29 DROP COLUMN a;  -- should not drop b`,
			},
			{
				Statement: `\d gtest29
              Table "public.gtest29"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 b      | integer |           |          | 
CREATE TABLE gtest30 (
    a int,
    b int GENERATED ALWAYS AS (a * 2) STORED
);`,
			},
			{
				Statement: `CREATE TABLE gtest30_1 () INHERITS (gtest30);`,
			},
			{
				Statement: `ALTER TABLE gtest30 ALTER COLUMN b DROP EXPRESSION;`,
			},
			{
				Statement: `\d gtest30
              Table "public.gtest30"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Number of child tables: 1 (Use \d+ to list them.)
\d gtest30_1
             Table "public.gtest30_1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Inherits: gtest30
DROP TABLE gtest30 CASCADE;`,
			},
			{
				Statement: `CREATE TABLE gtest30 (
    a int,
    b int GENERATED ALWAYS AS (a * 2) STORED
);`,
			},
			{
				Statement: `CREATE TABLE gtest30_1 () INHERITS (gtest30);`,
			},
			{
				Statement:   `ALTER TABLE ONLY gtest30 ALTER COLUMN b DROP EXPRESSION;  -- error`,
				ErrorString: `ALTER TABLE / DROP EXPRESSION must be applied to child tables too`,
			},
			{
				Statement: `\d gtest30
                            Table "public.gtest30"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           |          | 
 b      | integer |           |          | generated always as (a * 2) stored
Number of child tables: 1 (Use \d+ to list them.)
\d gtest30_1
                           Table "public.gtest30_1"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           |          | 
 b      | integer |           |          | generated always as (a * 2) stored
Inherits: gtest30
ALTER TABLE gtest30_1 ALTER COLUMN b DROP EXPRESSION;  -- error`,
				ErrorString: `cannot drop generation expression from inherited column`,
			},
			{
				Statement: `CREATE TABLE gtest26 (
    a int PRIMARY KEY,
    b int GENERATED ALWAYS AS (a * 2) STORED
);`,
			},
			{
				Statement: `CREATE FUNCTION gtest_trigger_func() RETURNS trigger
  LANGUAGE plpgsql
AS $$
BEGIN
  IF tg_op IN ('DELETE', 'UPDATE') THEN
    RAISE INFO '%: %: old = %', TG_NAME, TG_WHEN, OLD;`,
			},
			{
				Statement: `  END IF;`,
			},
			{
				Statement: `  IF tg_op IN ('INSERT', 'UPDATE') THEN
    RAISE INFO '%: %: new = %', TG_NAME, TG_WHEN, NEW;`,
			},
			{
				Statement: `  END IF;`,
			},
			{
				Statement: `  IF tg_op = 'DELETE' THEN
    RETURN OLD;`,
			},
			{
				Statement: `  ELSE
    RETURN NEW;`,
			},
			{
				Statement: `  END IF;`,
			},
			{
				Statement: `END
$$;`,
			},
			{
				Statement: `CREATE TRIGGER gtest1 BEFORE DELETE OR UPDATE ON gtest26
  FOR EACH ROW
  WHEN (OLD.b < 0)  -- ok
  EXECUTE PROCEDURE gtest_trigger_func();`,
			},
			{
				Statement: `CREATE TRIGGER gtest2a BEFORE INSERT OR UPDATE ON gtest26
  FOR EACH ROW
  WHEN (NEW.b < 0)  -- error
  EXECUTE PROCEDURE gtest_trigger_func();`,
				ErrorString: `BEFORE trigger's WHEN condition cannot reference NEW generated columns`,
			},
			{
				Statement: `CREATE TRIGGER gtest2b BEFORE INSERT OR UPDATE ON gtest26
  FOR EACH ROW
  WHEN (NEW.* IS NOT NULL)  -- error
  EXECUTE PROCEDURE gtest_trigger_func();`,
				ErrorString: `BEFORE trigger's WHEN condition cannot reference NEW generated columns`,
			},
			{
				Statement: `CREATE TRIGGER gtest2 BEFORE INSERT ON gtest26
  FOR EACH ROW
  WHEN (NEW.a < 0)
  EXECUTE PROCEDURE gtest_trigger_func();`,
			},
			{
				Statement: `CREATE TRIGGER gtest3 AFTER DELETE OR UPDATE ON gtest26
  FOR EACH ROW
  WHEN (OLD.b < 0)  -- ok
  EXECUTE PROCEDURE gtest_trigger_func();`,
			},
			{
				Statement: `CREATE TRIGGER gtest4 AFTER INSERT OR UPDATE ON gtest26
  FOR EACH ROW
  WHEN (NEW.b < 0)  -- ok
  EXECUTE PROCEDURE gtest_trigger_func();`,
			},
			{
				Statement: `INSERT INTO gtest26 (a) VALUES (-2), (0), (3);`,
			},
			{
				Statement: `INFO:  gtest2: BEFORE: new = (-2,)
INFO:  gtest4: AFTER: new = (-2,-4)
SELECT * FROM gtest26 ORDER BY a;`,
				Results: []sql.Row{{-2, -4}, {0, 0}, {3, 6}},
			},
			{
				Statement: `UPDATE gtest26 SET a = a * -2;`,
			},
			{
				Statement: `INFO:  gtest1: BEFORE: old = (-2,-4)
INFO:  gtest1: BEFORE: new = (4,)
INFO:  gtest3: AFTER: old = (-2,-4)
INFO:  gtest3: AFTER: new = (4,8)
INFO:  gtest4: AFTER: old = (3,6)
INFO:  gtest4: AFTER: new = (-6,-12)
SELECT * FROM gtest26 ORDER BY a;`,
				Results: []sql.Row{{-6, -12}, {0, 0}, {4, 8}},
			},
			{
				Statement: `DELETE FROM gtest26 WHERE a = -6;`,
			},
			{
				Statement: `INFO:  gtest1: BEFORE: old = (-6,-12)
INFO:  gtest3: AFTER: old = (-6,-12)
SELECT * FROM gtest26 ORDER BY a;`,
				Results: []sql.Row{{0, 0}, {4, 8}},
			},
			{
				Statement: `DROP TRIGGER gtest1 ON gtest26;`,
			},
			{
				Statement: `DROP TRIGGER gtest2 ON gtest26;`,
			},
			{
				Statement: `DROP TRIGGER gtest3 ON gtest26;`,
			},
			{
				Statement: `CREATE FUNCTION gtest_trigger_func3() RETURNS trigger
  LANGUAGE plpgsql
AS $$
BEGIN
  RAISE NOTICE 'OK';`,
			},
			{
				Statement: `  RETURN NEW;`,
			},
			{
				Statement: `END
$$;`,
			},
			{
				Statement: `CREATE TRIGGER gtest11 BEFORE UPDATE OF b ON gtest26
  FOR EACH ROW
  EXECUTE PROCEDURE gtest_trigger_func3();`,
			},
			{
				Statement: `UPDATE gtest26 SET a = 1 WHERE a = 0;`,
			},
			{
				Statement: `DROP TRIGGER gtest11 ON gtest26;`,
			},
			{
				Statement: `TRUNCATE gtest26;`,
			},
			{
				Statement: `CREATE FUNCTION gtest_trigger_func4() RETURNS trigger
  LANGUAGE plpgsql
AS $$
BEGIN
  NEW.a = 10;`,
			},
			{
				Statement: `  NEW.b = 300;`,
			},
			{
				Statement: `  RETURN NEW;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TRIGGER gtest12_01 BEFORE UPDATE ON gtest26
  FOR EACH ROW
  EXECUTE PROCEDURE gtest_trigger_func();`,
			},
			{
				Statement: `CREATE TRIGGER gtest12_02 BEFORE UPDATE ON gtest26
  FOR EACH ROW
  EXECUTE PROCEDURE gtest_trigger_func4();`,
			},
			{
				Statement: `CREATE TRIGGER gtest12_03 BEFORE UPDATE ON gtest26
  FOR EACH ROW
  EXECUTE PROCEDURE gtest_trigger_func();`,
			},
			{
				Statement: `INSERT INTO gtest26 (a) VALUES (1);`,
			},
			{
				Statement: `UPDATE gtest26 SET a = 11 WHERE a = 1;`,
			},
			{
				Statement: `INFO:  gtest12_01: BEFORE: old = (1,2)
INFO:  gtest12_01: BEFORE: new = (11,)
INFO:  gtest12_03: BEFORE: old = (1,2)
INFO:  gtest12_03: BEFORE: new = (10,)
SELECT * FROM gtest26 ORDER BY a;`,
				Results: []sql.Row{{10, 20}},
			},
			{
				Statement: `CREATE TABLE gtest28a (
  a int,
  b int,
  c int,
  x int GENERATED ALWAYS AS (b * 2) STORED
);`,
			},
			{
				Statement: `ALTER TABLE gtest28a DROP COLUMN a;`,
			},
			{
				Statement: `CREATE TABLE gtest28b (LIKE gtest28a INCLUDING GENERATED);`,
			},
			{
				Statement: `\d gtest28*
                           Table "public.gtest28a"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 b      | integer |           |          | 
 c      | integer |           |          | 
 x      | integer |           |          | generated always as (b * 2) stored
                           Table "public.gtest28b"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 b      | integer |           |          | 
 c      | integer |           |          | 
 x      | integer |           |          | generated always as (b * 2) stored`,
			},
		},
	})
}
