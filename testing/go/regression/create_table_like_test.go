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

func TestCreateTableLike(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_table_like)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_table_like,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `/* Test inheritance of structure (LIKE) */
CREATE TABLE inhx (xx text DEFAULT 'text');`,
			},
			{
				Statement: `/*
 * Test double inheritance
 *
 * Ensure that defaults are NOT included unless
 * INCLUDING DEFAULTS is specified
 */
CREATE TABLE ctla (aa TEXT);`,
			},
			{
				Statement: `CREATE TABLE ctlb (bb TEXT) INHERITS (ctla);`,
			},
			{
				Statement:   `CREATE TABLE foo (LIKE nonexistent);`,
				ErrorString: `relation "nonexistent" does not exist`,
			},
			{
				Statement: `CREATE TABLE inhe (ee text, LIKE inhx) inherits (ctlb);`,
			},
			{
				Statement: `INSERT INTO inhe VALUES ('ee-col1', 'ee-col2', DEFAULT, 'ee-col4');`,
			},
			{
				Statement: `SELECT * FROM inhe; /* Columns aa, bb, xx value NULL, ee */
   aa    |   bb    | ee |   xx    
---------+---------+----+---------
 ee-col1 | ee-col2 |    | ee-col4
(1 row)
SELECT * FROM inhx; /* Empty set since LIKE inherits structure only */
 xx 
----
(0 rows)
SELECT * FROM ctlb; /* Has ee entry */
   aa    |   bb    
---------+---------
 ee-col1 | ee-col2
(1 row)
SELECT * FROM ctla; /* Has ee entry */
   aa    
---------
 ee-col1
(1 row)
CREATE TABLE inhf (LIKE inhx, LIKE inhx); /* Throw error */
ERROR:  column "xx" specified more than once
CREATE TABLE inhf (LIKE inhx INCLUDING DEFAULTS INCLUDING CONSTRAINTS);`,
			},
			{
				Statement: `INSERT INTO inhf DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM inhf; /* Single entry with value 'text' */
  xx  
------
 text
(1 row)
ALTER TABLE inhx add constraint foo CHECK (xx = 'text');`,
			},
			{
				Statement: `ALTER TABLE inhx ADD PRIMARY KEY (xx);`,
			},
			{
				Statement: `CREATE TABLE inhg (LIKE inhx); /* Doesn't copy constraint */
INSERT INTO inhg VALUES ('foo');`,
			},
			{
				Statement: `DROP TABLE inhg;`,
			},
			{
				Statement: `CREATE TABLE inhg (x text, LIKE inhx INCLUDING CONSTRAINTS, y text); /* Copies constraints */
INSERT INTO inhg VALUES ('x', 'text', 'y'); /* Succeeds */
INSERT INTO inhg VALUES ('x', 'text', 'y'); /* Succeeds -- Unique constraints not copied */
INSERT INTO inhg VALUES ('x', 'foo',  'y');  /* fails due to constraint */
ERROR:  new row for relation "inhg" violates check constraint "foo"
SELECT * FROM inhg; /* Two records with three columns in order x=x, xx=text, y=y */
 x |  xx  | y 
---+------+---
 x | text | y
 x | text | y
(2 rows)
DROP TABLE inhg;`,
			},
			{
				Statement: `CREATE TABLE test_like_id_1 (a bigint GENERATED ALWAYS AS IDENTITY, b text);`,
			},
			{
				Statement: `\d test_like_id_1
                     Table "public.test_like_id_1"
 Column |  Type  | Collation | Nullable |           Default            
--------+--------+-----------+----------+------------------------------
 a      | bigint |           | not null | generated always as identity
 b      | text   |           |          | 
INSERT INTO test_like_id_1 (b) VALUES ('b1');`,
			},
			{
				Statement: `SELECT * FROM test_like_id_1;`,
				Results:   []sql.Row{{1, `b1`}},
			},
			{
				Statement: `CREATE TABLE test_like_id_2 (LIKE test_like_id_1);`,
			},
			{
				Statement: `\d test_like_id_2
          Table "public.test_like_id_2"
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 a      | bigint |           | not null | 
 b      | text   |           |          | 
INSERT INTO test_like_id_2 (b) VALUES ('b2');`,
				ErrorString: `null value in column "a" of relation "test_like_id_2" violates not-null constraint`,
			},
			{
				Statement: `SELECT * FROM test_like_id_2;  -- identity was not copied`,
				Results:   []sql.Row{},
			},
			{
				Statement: `CREATE TABLE test_like_id_3 (LIKE test_like_id_1 INCLUDING IDENTITY);`,
			},
			{
				Statement: `\d test_like_id_3
                     Table "public.test_like_id_3"
 Column |  Type  | Collation | Nullable |           Default            
--------+--------+-----------+----------+------------------------------
 a      | bigint |           | not null | generated always as identity
 b      | text   |           |          | 
INSERT INTO test_like_id_3 (b) VALUES ('b3');`,
			},
			{
				Statement: `SELECT * FROM test_like_id_3;  -- identity was copied and applied`,
				Results:   []sql.Row{{1, `b3`}},
			},
			{
				Statement: `DROP TABLE test_like_id_1, test_like_id_2, test_like_id_3;`,
			},
			{
				Statement: `CREATE TABLE test_like_gen_1 (a int, b int GENERATED ALWAYS AS (a * 2) STORED);`,
			},
			{
				Statement: `\d test_like_gen_1
                        Table "public.test_like_gen_1"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           |          | 
 b      | integer |           |          | generated always as (a * 2) stored
INSERT INTO test_like_gen_1 (a) VALUES (1);`,
			},
			{
				Statement: `SELECT * FROM test_like_gen_1;`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `CREATE TABLE test_like_gen_2 (LIKE test_like_gen_1);`,
			},
			{
				Statement: `\d test_like_gen_2
          Table "public.test_like_gen_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
INSERT INTO test_like_gen_2 (a) VALUES (1);`,
			},
			{
				Statement: `SELECT * FROM test_like_gen_2;`,
				Results:   []sql.Row{{1, ``}},
			},
			{
				Statement: `CREATE TABLE test_like_gen_3 (LIKE test_like_gen_1 INCLUDING GENERATED);`,
			},
			{
				Statement: `\d test_like_gen_3
                        Table "public.test_like_gen_3"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 a      | integer |           |          | 
 b      | integer |           |          | generated always as (a * 2) stored
INSERT INTO test_like_gen_3 (a) VALUES (1);`,
			},
			{
				Statement: `SELECT * FROM test_like_gen_3;`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `DROP TABLE test_like_gen_1, test_like_gen_2, test_like_gen_3;`,
			},
			{
				Statement: `CREATE TABLE test_like_4 (b int DEFAULT 42,
  c int GENERATED ALWAYS AS (a * 2) STORED,
  a int CHECK (a > 0));`,
			},
			{
				Statement: `\d test_like_4
                          Table "public.test_like_4"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 b      | integer |           |          | 42
 c      | integer |           |          | generated always as (a * 2) stored
 a      | integer |           |          | 
Check constraints:
    "test_like_4_a_check" CHECK (a > 0)
CREATE TABLE test_like_4a (LIKE test_like_4);`,
			},
			{
				Statement: `CREATE TABLE test_like_4b (LIKE test_like_4 INCLUDING DEFAULTS);`,
			},
			{
				Statement: `CREATE TABLE test_like_4c (LIKE test_like_4 INCLUDING GENERATED);`,
			},
			{
				Statement: `CREATE TABLE test_like_4d (LIKE test_like_4 INCLUDING DEFAULTS INCLUDING GENERATED);`,
			},
			{
				Statement: `\d test_like_4a
            Table "public.test_like_4a"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 b      | integer |           |          | 
 c      | integer |           |          | 
 a      | integer |           |          | 
INSERT INTO test_like_4a (a) VALUES(11);`,
			},
			{
				Statement: `SELECT a, b, c FROM test_like_4a;`,
				Results:   []sql.Row{{11, ``, ``}},
			},
			{
				Statement: `\d test_like_4b
            Table "public.test_like_4b"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 b      | integer |           |          | 42
 c      | integer |           |          | 
 a      | integer |           |          | 
INSERT INTO test_like_4b (a) VALUES(11);`,
			},
			{
				Statement: `SELECT a, b, c FROM test_like_4b;`,
				Results:   []sql.Row{{11, 42, ``}},
			},
			{
				Statement: `\d test_like_4c
                         Table "public.test_like_4c"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 b      | integer |           |          | 
 c      | integer |           |          | generated always as (a * 2) stored
 a      | integer |           |          | 
INSERT INTO test_like_4c (a) VALUES(11);`,
			},
			{
				Statement: `SELECT a, b, c FROM test_like_4c;`,
				Results:   []sql.Row{{11, ``, 22}},
			},
			{
				Statement: `\d test_like_4d
                         Table "public.test_like_4d"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 b      | integer |           |          | 42
 c      | integer |           |          | generated always as (a * 2) stored
 a      | integer |           |          | 
INSERT INTO test_like_4d (a) VALUES(11);`,
			},
			{
				Statement: `SELECT a, b, c FROM test_like_4d;`,
				Results:   []sql.Row{{11, 42, 22}},
			},
			{
				Statement: `CREATE TABLE test_like_5 (x point, y point, z point);`,
			},
			{
				Statement: `CREATE TABLE test_like_5x (p int CHECK (p > 0),
   q int GENERATED ALWAYS AS (p * 2) STORED);`,
			},
			{
				Statement: `CREATE TABLE test_like_5c (LIKE test_like_4 INCLUDING ALL)
  INHERITS (test_like_5, test_like_5x);`,
			},
			{
				Statement: `\d test_like_5c
                         Table "public.test_like_5c"
 Column |  Type   | Collation | Nullable |              Default               
--------+---------+-----------+----------+------------------------------------
 x      | point   |           |          | 
 y      | point   |           |          | 
 z      | point   |           |          | 
 p      | integer |           |          | 
 q      | integer |           |          | generated always as (p * 2) stored
 b      | integer |           |          | 42
 c      | integer |           |          | generated always as (a * 2) stored
 a      | integer |           |          | 
Check constraints:
    "test_like_4_a_check" CHECK (a > 0)
    "test_like_5x_p_check" CHECK (p > 0)
Inherits: test_like_5,
          test_like_5x
DROP TABLE test_like_4, test_like_4a, test_like_4b, test_like_4c, test_like_4d;`,
			},
			{
				Statement: `DROP TABLE test_like_5, test_like_5x, test_like_5c;`,
			},
			{
				Statement: `CREATE TABLE inhg (x text, LIKE inhx INCLUDING INDEXES, y text); /* copies indexes */
INSERT INTO inhg VALUES (5, 10);`,
			},
			{
				Statement:   `INSERT INTO inhg VALUES (20, 10); -- should fail`,
				ErrorString: `duplicate key value violates unique constraint "inhg_pkey"`,
			},
			{
				Statement: `DROP TABLE inhg;`,
			},
			{
				Statement: `/* Multiple primary keys creation should fail */
CREATE TABLE inhg (x text, LIKE inhx INCLUDING INDEXES, PRIMARY KEY(x)); /* fails */
ERROR:  multiple primary keys for table "inhg" are not allowed
CREATE TABLE inhz (xx text DEFAULT 'text', yy int UNIQUE);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX inhz_xx_idx on inhz (xx) WHERE xx <> 'test';`,
			},
			{
				Statement: `/* Ok to create multiple unique indexes */
CREATE TABLE inhg (x text UNIQUE, LIKE inhz INCLUDING INDEXES);`,
			},
			{
				Statement: `INSERT INTO inhg (xx, yy, x) VALUES ('test', 5, 10);`,
			},
			{
				Statement: `INSERT INTO inhg (xx, yy, x) VALUES ('test', 10, 15);`,
			},
			{
				Statement:   `INSERT INTO inhg (xx, yy, x) VALUES ('foo', 10, 15); -- should fail`,
				ErrorString: `duplicate key value violates unique constraint "inhg_x_key"`,
			},
			{
				Statement: `DROP TABLE inhg;`,
			},
			{
				Statement: `DROP TABLE inhz;`,
			},
			{
				Statement: `/* Use primary key imported by LIKE for self-referential FK constraint */
CREATE TABLE inhz (x text REFERENCES inhz, LIKE inhx INCLUDING INDEXES);`,
			},
			{
				Statement: `\d inhz
              Table "public.inhz"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 x      | text |           |          | 
 xx     | text |           | not null | 
Indexes:
    "inhz_pkey" PRIMARY KEY, btree (xx)
Foreign-key constraints:
    "inhz_x_fkey" FOREIGN KEY (x) REFERENCES inhz(xx)
Referenced by:
    TABLE "inhz" CONSTRAINT "inhz_x_fkey" FOREIGN KEY (x) REFERENCES inhz(xx)
DROP TABLE inhz;`,
			},
			{
				Statement: `CREATE TABLE ctlt1 (a text CHECK (length(a) > 2) PRIMARY KEY, b text);`,
			},
			{
				Statement: `CREATE INDEX ctlt1_b_key ON ctlt1 (b);`,
			},
			{
				Statement: `CREATE INDEX ctlt1_fnidx ON ctlt1 ((a || b));`,
			},
			{
				Statement: `CREATE STATISTICS ctlt1_a_b_stat ON a,b FROM ctlt1;`,
			},
			{
				Statement: `CREATE STATISTICS ctlt1_expr_stat ON (a || b) FROM ctlt1;`,
			},
			{
				Statement: `COMMENT ON STATISTICS ctlt1_a_b_stat IS 'ab stats';`,
			},
			{
				Statement: `COMMENT ON STATISTICS ctlt1_expr_stat IS 'ab expr stats';`,
			},
			{
				Statement: `COMMENT ON COLUMN ctlt1.a IS 'A';`,
			},
			{
				Statement: `COMMENT ON COLUMN ctlt1.b IS 'B';`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT ctlt1_a_check ON ctlt1 IS 't1_a_check';`,
			},
			{
				Statement: `COMMENT ON INDEX ctlt1_pkey IS 'index pkey';`,
			},
			{
				Statement: `COMMENT ON INDEX ctlt1_b_key IS 'index b_key';`,
			},
			{
				Statement: `ALTER TABLE ctlt1 ALTER COLUMN a SET STORAGE MAIN;`,
			},
			{
				Statement: `CREATE TABLE ctlt2 (c text);`,
			},
			{
				Statement: `ALTER TABLE ctlt2 ALTER COLUMN c SET STORAGE EXTERNAL;`,
			},
			{
				Statement: `COMMENT ON COLUMN ctlt2.c IS 'C';`,
			},
			{
				Statement: `CREATE TABLE ctlt3 (a text CHECK (length(a) < 5), c text CHECK (length(c) < 7));`,
			},
			{
				Statement: `ALTER TABLE ctlt3 ALTER COLUMN c SET STORAGE EXTERNAL;`,
			},
			{
				Statement: `ALTER TABLE ctlt3 ALTER COLUMN a SET STORAGE MAIN;`,
			},
			{
				Statement: `CREATE INDEX ctlt3_fnidx ON ctlt3 ((a || c));`,
			},
			{
				Statement: `COMMENT ON COLUMN ctlt3.a IS 'A3';`,
			},
			{
				Statement: `COMMENT ON COLUMN ctlt3.c IS 'C';`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT ctlt3_a_check ON ctlt3 IS 't3_a_check';`,
			},
			{
				Statement: `CREATE TABLE ctlt4 (a text, c text);`,
			},
			{
				Statement: `ALTER TABLE ctlt4 ALTER COLUMN c SET STORAGE EXTERNAL;`,
			},
			{
				Statement: `CREATE TABLE ctlt12_storage (LIKE ctlt1 INCLUDING STORAGE, LIKE ctlt2 INCLUDING STORAGE);`,
			},
			{
				Statement: `\d+ ctlt12_storage
                             Table "public.ctlt12_storage"
 Column | Type | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------+-----------+----------+---------+----------+--------------+-------------
 a      | text |           | not null |         | main     |              | 
 b      | text |           |          |         | extended |              | 
 c      | text |           |          |         | external |              | 
CREATE TABLE ctlt12_comments (LIKE ctlt1 INCLUDING COMMENTS, LIKE ctlt2 INCLUDING COMMENTS);`,
			},
			{
				Statement: `\d+ ctlt12_comments
                             Table "public.ctlt12_comments"
 Column | Type | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------+-----------+----------+---------+----------+--------------+-------------
 a      | text |           | not null |         | extended |              | A
 b      | text |           |          |         | extended |              | B
 c      | text |           |          |         | extended |              | C
CREATE TABLE ctlt1_inh (LIKE ctlt1 INCLUDING CONSTRAINTS INCLUDING COMMENTS) INHERITS (ctlt1);`,
			},
			{
				Statement: `\d+ ctlt1_inh
                                Table "public.ctlt1_inh"
 Column | Type | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------+-----------+----------+---------+----------+--------------+-------------
 a      | text |           | not null |         | main     |              | A
 b      | text |           |          |         | extended |              | B
Check constraints:
    "ctlt1_a_check" CHECK (length(a) > 2)
Inherits: ctlt1
SELECT description FROM pg_description, pg_constraint c WHERE classoid = 'pg_constraint'::regclass AND objoid = c.oid AND c.conrelid = 'ctlt1_inh'::regclass;`,
				Results: []sql.Row{{`t1_a_check`}},
			},
			{
				Statement: `CREATE TABLE ctlt13_inh () INHERITS (ctlt1, ctlt3);`,
			},
			{
				Statement: `\d+ ctlt13_inh
                               Table "public.ctlt13_inh"
 Column | Type | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------+-----------+----------+---------+----------+--------------+-------------
 a      | text |           | not null |         | main     |              | 
 b      | text |           |          |         | extended |              | 
 c      | text |           |          |         | external |              | 
Check constraints:
    "ctlt1_a_check" CHECK (length(a) > 2)
    "ctlt3_a_check" CHECK (length(a) < 5)
    "ctlt3_c_check" CHECK (length(c) < 7)
Inherits: ctlt1,
          ctlt3
CREATE TABLE ctlt13_like (LIKE ctlt3 INCLUDING CONSTRAINTS INCLUDING INDEXES INCLUDING COMMENTS INCLUDING STORAGE) INHERITS (ctlt1);`,
			},
			{
				Statement: `\d+ ctlt13_like
                               Table "public.ctlt13_like"
 Column | Type | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------+-----------+----------+---------+----------+--------------+-------------
 a      | text |           | not null |         | main     |              | A3
 b      | text |           |          |         | extended |              | 
 c      | text |           |          |         | external |              | C
Indexes:
    "ctlt13_like_expr_idx" btree ((a || c))
Check constraints:
    "ctlt1_a_check" CHECK (length(a) > 2)
    "ctlt3_a_check" CHECK (length(a) < 5)
    "ctlt3_c_check" CHECK (length(c) < 7)
Inherits: ctlt1
SELECT description FROM pg_description, pg_constraint c WHERE classoid = 'pg_constraint'::regclass AND objoid = c.oid AND c.conrelid = 'ctlt13_like'::regclass;`,
				Results: []sql.Row{{`t3_a_check`}},
			},
			{
				Statement: `CREATE TABLE ctlt_all (LIKE ctlt1 INCLUDING ALL);`,
			},
			{
				Statement: `\d+ ctlt_all
                                Table "public.ctlt_all"
 Column | Type | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------+-----------+----------+---------+----------+--------------+-------------
 a      | text |           | not null |         | main     |              | A
 b      | text |           |          |         | extended |              | B
Indexes:
    "ctlt_all_pkey" PRIMARY KEY, btree (a)
    "ctlt_all_b_idx" btree (b)
    "ctlt_all_expr_idx" btree ((a || b))
Check constraints:
    "ctlt1_a_check" CHECK (length(a) > 2)
Statistics objects:
    "public.ctlt_all_a_b_stat" ON a, b FROM ctlt_all
    "public.ctlt_all_expr_stat" ON (a || b) FROM ctlt_all
SELECT c.relname, objsubid, description FROM pg_description, pg_index i, pg_class c WHERE classoid = 'pg_class'::regclass AND objoid = i.indexrelid AND c.oid = i.indexrelid AND i.indrelid = 'ctlt_all'::regclass ORDER BY c.relname, objsubid;`,
				Results: []sql.Row{{`ctlt_all_b_idx`, 0, `index b_key`}, {`ctlt_all_pkey`, 0, `index pkey`}},
			},
			{
				Statement: `SELECT s.stxname, objsubid, description FROM pg_description, pg_statistic_ext s WHERE classoid = 'pg_statistic_ext'::regclass AND objoid = s.oid AND s.stxrelid = 'ctlt_all'::regclass ORDER BY s.stxname, objsubid;`,
				Results:   []sql.Row{{`ctlt_all_a_b_stat`, 0, `ab stats`}, {`ctlt_all_expr_stat`, 0, `ab expr stats`}},
			},
			{
				Statement:   `CREATE TABLE inh_error1 () INHERITS (ctlt1, ctlt4);`,
				ErrorString: `inherited column "a" has a storage parameter conflict`,
			},
			{
				Statement:   `CREATE TABLE inh_error2 (LIKE ctlt4 INCLUDING STORAGE) INHERITS (ctlt1);`,
				ErrorString: `column "a" has a storage parameter conflict`,
			},
			{
				Statement: `CREATE TABLE pg_attrdef (LIKE ctlt1 INCLUDING ALL);`,
			},
			{
				Statement: `\d+ public.pg_attrdef
                               Table "public.pg_attrdef"
 Column | Type | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------+-----------+----------+---------+----------+--------------+-------------
 a      | text |           | not null |         | main     |              | A
 b      | text |           |          |         | extended |              | B
Indexes:
    "pg_attrdef_pkey" PRIMARY KEY, btree (a)
    "pg_attrdef_b_idx" btree (b)
    "pg_attrdef_expr_idx" btree ((a || b))
Check constraints:
    "ctlt1_a_check" CHECK (length(a) > 2)
Statistics objects:
    "public.pg_attrdef_a_b_stat" ON a, b FROM public.pg_attrdef
    "public.pg_attrdef_expr_stat" ON (a || b) FROM public.pg_attrdef
DROP TABLE public.pg_attrdef;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE SCHEMA ctl_schema;`,
			},
			{
				Statement: `SET LOCAL search_path = ctl_schema, public;`,
			},
			{
				Statement: `CREATE TABLE ctlt1 (LIKE ctlt1 INCLUDING ALL);`,
			},
			{
				Statement: `\d+ ctlt1
                                Table "ctl_schema.ctlt1"
 Column | Type | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+------+-----------+----------+---------+----------+--------------+-------------
 a      | text |           | not null |         | main     |              | A
 b      | text |           |          |         | extended |              | B
Indexes:
    "ctlt1_pkey" PRIMARY KEY, btree (a)
    "ctlt1_b_idx" btree (b)
    "ctlt1_expr_idx" btree ((a || b))
Check constraints:
    "ctlt1_a_check" CHECK (length(a) > 2)
Statistics objects:
    "ctl_schema.ctlt1_a_b_stat" ON a, b FROM ctlt1
    "ctl_schema.ctlt1_expr_stat" ON (a || b) FROM ctlt1
ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE ctlt1, ctlt2, ctlt3, ctlt4, ctlt12_storage, ctlt12_comments, ctlt1_inh, ctlt13_inh, ctlt13_like, ctlt_all, ctla, ctlb CASCADE;`,
			},
			{
				Statement: `CREATE TABLE noinh_con_copy (a int CHECK (a > 0) NO INHERIT);`,
			},
			{
				Statement: `CREATE TABLE noinh_con_copy1 (LIKE noinh_con_copy INCLUDING CONSTRAINTS);`,
			},
			{
				Statement: `\d noinh_con_copy1
          Table "public.noinh_con_copy1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Check constraints:
    "noinh_con_copy_a_check" CHECK (a > 0) NO INHERIT
CREATE TABLE noinh_con_copy1_parted (LIKE noinh_con_copy INCLUDING ALL)
  PARTITION BY LIST (a);`,
				ErrorString: `cannot add NO INHERIT constraint to partitioned table "noinh_con_copy1_parted"`,
			},
			{
				Statement: `DROP TABLE noinh_con_copy, noinh_con_copy1;`,
			},
			{
				Statement: `/* LIKE with other relation kinds */
CREATE TABLE ctlt4 (a int, b text);`,
			},
			{
				Statement: `CREATE SEQUENCE ctlseq1;`,
			},
			{
				Statement:   `CREATE TABLE ctlt10 (LIKE ctlseq1);  -- fail`,
				ErrorString: `relation "ctlseq1" is invalid in LIKE clause`,
			},
			{
				Statement: `CREATE VIEW ctlv1 AS SELECT * FROM ctlt4;`,
			},
			{
				Statement: `CREATE TABLE ctlt11 (LIKE ctlv1);`,
			},
			{
				Statement: `CREATE TABLE ctlt11a (LIKE ctlv1 INCLUDING ALL);`,
			},
			{
				Statement: `CREATE TYPE ctlty1 AS (a int, b text);`,
			},
			{
				Statement: `CREATE TABLE ctlt12 (LIKE ctlty1);`,
			},
			{
				Statement: `DROP SEQUENCE ctlseq1;`,
			},
			{
				Statement: `DROP TYPE ctlty1;`,
			},
			{
				Statement: `DROP VIEW ctlv1;`,
			},
			{
				Statement: `DROP TABLE IF EXISTS ctlt4, ctlt10, ctlt11, ctlt11a, ctlt12;`,
			},
		},
	})
}
