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

func TestAlterTable(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_alter_table)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_alter_table,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET client_min_messages TO 'warning';`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_alter_table_user1;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE USER regress_alter_table_user1;`,
			},
			{
				Statement: `CREATE TABLE attmp (initial int4);`,
			},
			{
				Statement:   `COMMENT ON TABLE attmp_wrong IS 'table comment';`,
				ErrorString: `relation "attmp_wrong" does not exist`,
			},
			{
				Statement: `COMMENT ON TABLE attmp IS 'table comment';`,
			},
			{
				Statement: `COMMENT ON TABLE attmp IS NULL;`,
			},
			{
				Statement:   `ALTER TABLE attmp ADD COLUMN xmin integer; -- fails`,
				ErrorString: `column name "xmin" conflicts with a system column name`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN a int4 default 3;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN b name;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN c text;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN d float8;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN e float4;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN f int2;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN g polygon;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN i char;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN k int4;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN l tid;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN m xid;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN n oidvector;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN p boolean;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN q point;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN r lseg;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN s path;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN t box;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN v timestamp;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN w interval;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN x float8[];`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN y float4[];`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN z int2[];`,
			},
			{
				Statement: `INSERT INTO attmp (a, b, c, d, e, f, g,    i,    k, l, m, n, p, q, r, s, t,
	v, w, x, y, z)
   VALUES (4, 'name', 'text', 4.1, 4.1, 2, '(4.1,4.1,3.1,3.1)',
	'c',
	314159, '(1,1)', '512',
	'1 2 3 4 5 6 7 8', true, '(1.1,1.1)', '(4.1,4.1,3.1,3.1)',
	'(0,2,4.1,4.1,3.1,3.1)', '(4.1,4.1,3.1,3.1)',
	'epoch', '01:00:10', '{1.0,2.0,3.0,4.0}', '{1.0,2.0,3.0,4.0}', '{1,2,3,4}');`,
			},
			{
				Statement: `SELECT * FROM attmp;`,
				Results:   []sql.Row{{``, 4, `name`, `text`, 4.1, 4.1, 2, `((4.1,4.1),(3.1,3.1))`, `c`, 314159, `(1,1)`, 512, `1 2 3 4 5 6 7 8`, true, `(1.1,1.1)`, `[(4.1,4.1),(3.1,3.1)]`, `((0,2),(4.1,4.1),(3.1,3.1))`, `(4.1,4.1),(3.1,3.1)`, `Thu Jan 01 00:00:00 1970`, `@ 1 hour 10 secs`, `{1,2,3,4}`, `{1,2,3,4}`, `{1,2,3,4}`}},
			},
			{
				Statement: `DROP TABLE attmp;`,
			},
			{
				Statement: `CREATE TABLE attmp (
	initial 	int4
);`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN a int4;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN b name;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN c text;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN d float8;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN e float4;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN f int2;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN g polygon;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN i char;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN k int4;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN l tid;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN m xid;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN n oidvector;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN p boolean;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN q point;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN r lseg;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN s path;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN t box;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN v timestamp;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN w interval;`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN x float8[];`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN y float4[];`,
			},
			{
				Statement: `ALTER TABLE attmp ADD COLUMN z int2[];`,
			},
			{
				Statement: `INSERT INTO attmp (a, b, c, d, e, f, g,    i,   k, l, m, n, p, q, r, s, t,
	v, w, x, y, z)
   VALUES (4, 'name', 'text', 4.1, 4.1, 2, '(4.1,4.1,3.1,3.1)',
        'c',
	314159, '(1,1)', '512',
	'1 2 3 4 5 6 7 8', true, '(1.1,1.1)', '(4.1,4.1,3.1,3.1)',
	'(0,2,4.1,4.1,3.1,3.1)', '(4.1,4.1,3.1,3.1)',
	'epoch', '01:00:10', '{1.0,2.0,3.0,4.0}', '{1.0,2.0,3.0,4.0}', '{1,2,3,4}');`,
			},
			{
				Statement: `SELECT * FROM attmp;`,
				Results:   []sql.Row{{``, 4, `name`, `text`, 4.1, 4.1, 2, `((4.1,4.1),(3.1,3.1))`, `c`, 314159, `(1,1)`, 512, `1 2 3 4 5 6 7 8`, true, `(1.1,1.1)`, `[(4.1,4.1),(3.1,3.1)]`, `((0,2),(4.1,4.1),(3.1,3.1))`, `(4.1,4.1),(3.1,3.1)`, `Thu Jan 01 00:00:00 1970`, `@ 1 hour 10 secs`, `{1,2,3,4}`, `{1,2,3,4}`, `{1,2,3,4}`}},
			},
			{
				Statement: `CREATE INDEX attmp_idx ON attmp (a, (d + e), b);`,
			},
			{
				Statement:   `ALTER INDEX attmp_idx ALTER COLUMN 0 SET STATISTICS 1000;`,
				ErrorString: `column number must be in range from 1 to 32767`,
			},
			{
				Statement:   `ALTER INDEX attmp_idx ALTER COLUMN 1 SET STATISTICS 1000;`,
				ErrorString: `cannot alter statistics on non-expression column "a" of index "attmp_idx"`,
			},
			{
				Statement: `ALTER INDEX attmp_idx ALTER COLUMN 2 SET STATISTICS 1000;`,
			},
			{
				Statement: `\d+ attmp_idx
                        Index "public.attmp_idx"
 Column |       Type       | Key? | Definition | Storage | Stats target 
--------+------------------+------+------------+---------+--------------
 a      | integer          | yes  | a          | plain   | 
 expr   | double precision | yes  | (d + e)    | plain   | 1000
 b      | cstring          | yes  | b          | plain   | 
btree, for table "public.attmp"
ALTER INDEX attmp_idx ALTER COLUMN 3 SET STATISTICS 1000;`,
				ErrorString: `cannot alter statistics on non-expression column "b" of index "attmp_idx"`,
			},
			{
				Statement:   `ALTER INDEX attmp_idx ALTER COLUMN 4 SET STATISTICS 1000;`,
				ErrorString: `column number 4 of relation "attmp_idx" does not exist`,
			},
			{
				Statement: `ALTER INDEX attmp_idx ALTER COLUMN 2 SET STATISTICS -1;`,
			},
			{
				Statement: `DROP TABLE attmp;`,
			},
			{
				Statement: `CREATE TABLE attmp (regtable int);`,
			},
			{
				Statement: `CREATE TEMP TABLE attmp (attmptable int);`,
			},
			{
				Statement: `ALTER TABLE attmp RENAME TO attmp_new;`,
			},
			{
				Statement: `SELECT * FROM attmp;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM attmp_new;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `ALTER TABLE attmp RENAME TO attmp_new2;`,
			},
			{
				Statement:   `SELECT * FROM attmp;		-- should fail`,
				ErrorString: `relation "attmp" does not exist`,
			},
			{
				Statement: `SELECT * FROM attmp_new;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM attmp_new2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE attmp_new;`,
			},
			{
				Statement: `DROP TABLE attmp_new2;`,
			},
			{
				Statement: `CREATE TABLE part_attmp (a int primary key) partition by range (a);`,
			},
			{
				Statement: `CREATE TABLE part_attmp1 PARTITION OF part_attmp FOR VALUES FROM (0) TO (100);`,
			},
			{
				Statement: `ALTER INDEX part_attmp_pkey RENAME TO part_attmp_index;`,
			},
			{
				Statement: `ALTER INDEX part_attmp1_pkey RENAME TO part_attmp1_index;`,
			},
			{
				Statement: `ALTER TABLE part_attmp RENAME TO part_at2tmp;`,
			},
			{
				Statement: `ALTER TABLE part_attmp1 RENAME TO part_at2tmp1;`,
			},
			{
				Statement: `SET ROLE regress_alter_table_user1;`,
			},
			{
				Statement:   `ALTER INDEX part_attmp_index RENAME TO fail;`,
				ErrorString: `must be owner of index part_attmp_index`,
			},
			{
				Statement:   `ALTER INDEX part_attmp1_index RENAME TO fail;`,
				ErrorString: `must be owner of index part_attmp1_index`,
			},
			{
				Statement:   `ALTER TABLE part_at2tmp RENAME TO fail;`,
				ErrorString: `must be owner of table part_at2tmp`,
			},
			{
				Statement:   `ALTER TABLE part_at2tmp1 RENAME TO fail;`,
				ErrorString: `must be owner of table part_at2tmp1`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `DROP TABLE part_at2tmp;`,
			},
			{
				Statement: `CREATE TABLE attmp_array (id int);`,
			},
			{
				Statement: `CREATE TABLE attmp_array2 (id int);`,
			},
			{
				Statement: `SELECT typname FROM pg_type WHERE oid = 'attmp_array[]'::regtype;`,
				Results:   []sql.Row{{`_attmp_array`}},
			},
			{
				Statement: `SELECT typname FROM pg_type WHERE oid = 'attmp_array2[]'::regtype;`,
				Results:   []sql.Row{{`_attmp_array2`}},
			},
			{
				Statement: `ALTER TABLE attmp_array2 RENAME TO _attmp_array;`,
			},
			{
				Statement: `SELECT typname FROM pg_type WHERE oid = 'attmp_array[]'::regtype;`,
				Results:   []sql.Row{{`__attmp_array`}},
			},
			{
				Statement: `SELECT typname FROM pg_type WHERE oid = '_attmp_array[]'::regtype;`,
				Results:   []sql.Row{{`___attmp_array`}},
			},
			{
				Statement: `DROP TABLE _attmp_array;`,
			},
			{
				Statement: `DROP TABLE attmp_array;`,
			},
			{
				Statement: `CREATE TABLE attmp_array (id int);`,
			},
			{
				Statement: `SELECT typname FROM pg_type WHERE oid = 'attmp_array[]'::regtype;`,
				Results:   []sql.Row{{`_attmp_array`}},
			},
			{
				Statement: `ALTER TABLE attmp_array RENAME TO _attmp_array;`,
			},
			{
				Statement: `SELECT typname FROM pg_type WHERE oid = '_attmp_array[]'::regtype;`,
				Results:   []sql.Row{{`__attmp_array`}},
			},
			{
				Statement: `DROP TABLE _attmp_array;`,
			},
			{
				Statement: `ALTER INDEX IF EXISTS __onek_unique1 RENAME TO attmp_onek_unique1;`,
			},
			{
				Statement: `ALTER INDEX IF EXISTS __attmp_onek_unique1 RENAME TO onek_unique1;`,
			},
			{
				Statement: `ALTER INDEX onek_unique1 RENAME TO attmp_onek_unique1;`,
			},
			{
				Statement: `ALTER INDEX attmp_onek_unique1 RENAME TO onek_unique1;`,
			},
			{
				Statement: `SET ROLE regress_alter_table_user1;`,
			},
			{
				Statement:   `ALTER INDEX onek_unique1 RENAME TO fail;  -- permission denied`,
				ErrorString: `must be owner of index onek_unique1`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `CREATE TABLE alter_idx_rename_test (a INT);`,
			},
			{
				Statement: `CREATE INDEX alter_idx_rename_test_idx ON alter_idx_rename_test (a);`,
			},
			{
				Statement: `CREATE TABLE alter_idx_rename_test_parted (a INT) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE INDEX alter_idx_rename_test_parted_idx ON alter_idx_rename_test_parted (a);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER INDEX alter_idx_rename_test RENAME TO alter_idx_rename_test_2;`,
			},
			{
				Statement: `ALTER INDEX alter_idx_rename_test_parted RENAME TO alter_idx_rename_test_parted_2;`,
			},
			{
				Statement: `SELECT relation::regclass, mode FROM pg_locks
WHERE pid = pg_backend_pid() AND locktype = 'relation'
  AND relation::regclass::text LIKE 'alter\_idx%'
ORDER BY relation::regclass::text COLLATE "C";`,
				Results: []sql.Row{{`alter_idx_rename_test_2`, `AccessExclusiveLock`}, {`alter_idx_rename_test_parted_2`, `AccessExclusiveLock`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER INDEX alter_idx_rename_test_idx RENAME TO alter_idx_rename_test_idx_2;`,
			},
			{
				Statement: `ALTER INDEX alter_idx_rename_test_parted_idx RENAME TO alter_idx_rename_test_parted_idx_2;`,
			},
			{
				Statement: `SELECT relation::regclass, mode FROM pg_locks
WHERE pid = pg_backend_pid() AND locktype = 'relation'
  AND relation::regclass::text LIKE 'alter\_idx%'
ORDER BY relation::regclass::text COLLATE "C";`,
				Results: []sql.Row{{`alter_idx_rename_test_idx_2`, `ShareUpdateExclusiveLock`}, {`alter_idx_rename_test_parted_idx_2`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER TABLE alter_idx_rename_test_idx_2 RENAME TO alter_idx_rename_test_idx_3;`,
			},
			{
				Statement: `ALTER TABLE alter_idx_rename_test_parted_idx_2 RENAME TO alter_idx_rename_test_parted_idx_3;`,
			},
			{
				Statement: `SELECT relation::regclass, mode FROM pg_locks
WHERE pid = pg_backend_pid() AND locktype = 'relation'
  AND relation::regclass::text LIKE 'alter\_idx%'
ORDER BY relation::regclass::text COLLATE "C";`,
				Results: []sql.Row{{`alter_idx_rename_test_idx_3`, `AccessExclusiveLock`}, {`alter_idx_rename_test_parted_idx_3`, `AccessExclusiveLock`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP TABLE alter_idx_rename_test_2;`,
			},
			{
				Statement: `CREATE VIEW attmp_view (unique1) AS SELECT unique1 FROM tenk1;`,
			},
			{
				Statement: `ALTER TABLE attmp_view RENAME TO attmp_view_new;`,
			},
			{
				Statement: `SET ROLE regress_alter_table_user1;`,
			},
			{
				Statement:   `ALTER VIEW attmp_view_new RENAME TO fail;  -- permission denied`,
				ErrorString: `must be owner of view attmp_view_new`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `set enable_seqscan to off;`,
			},
			{
				Statement: `set enable_bitmapscan to off;`,
			},
			{
				Statement: `SELECT unique1 FROM tenk1 WHERE unique1 < 5;`,
				Results:   []sql.Row{{0}, {1}, {2}, {3}, {4}},
			},
			{
				Statement: `reset enable_seqscan;`,
			},
			{
				Statement: `reset enable_bitmapscan;`,
			},
			{
				Statement: `DROP VIEW attmp_view_new;`,
			},
			{
				Statement: `alter table stud_emp rename to pg_toast_stud_emp;`,
			},
			{
				Statement: `alter table pg_toast_stud_emp rename to stud_emp;`,
			},
			{
				Statement: `ALTER TABLE onek ADD CONSTRAINT onek_unique1_constraint UNIQUE (unique1);`,
			},
			{
				Statement: `ALTER INDEX onek_unique1_constraint RENAME TO onek_unique1_constraint_foo;`,
			},
			{
				Statement: `ALTER TABLE onek DROP CONSTRAINT onek_unique1_constraint_foo;`,
			},
			{
				Statement: `ALTER TABLE onek ADD CONSTRAINT onek_check_constraint CHECK (unique1 >= 0);`,
			},
			{
				Statement: `ALTER TABLE onek RENAME CONSTRAINT onek_check_constraint TO onek_check_constraint_foo;`,
			},
			{
				Statement: `ALTER TABLE onek DROP CONSTRAINT onek_check_constraint_foo;`,
			},
			{
				Statement: `ALTER TABLE onek ADD CONSTRAINT onek_unique1_constraint UNIQUE (unique1);`,
			},
			{
				Statement:   `DROP INDEX onek_unique1_constraint;  -- to see whether it's there`,
				ErrorString: `cannot drop index onek_unique1_constraint because constraint onek_unique1_constraint on table onek requires it`,
			},
			{
				Statement: `ALTER TABLE onek RENAME CONSTRAINT onek_unique1_constraint TO onek_unique1_constraint_foo;`,
			},
			{
				Statement:   `DROP INDEX onek_unique1_constraint_foo;  -- to see whether it's there`,
				ErrorString: `cannot drop index onek_unique1_constraint_foo because constraint onek_unique1_constraint_foo on table onek requires it`,
			},
			{
				Statement: `ALTER TABLE onek DROP CONSTRAINT onek_unique1_constraint_foo;`,
			},
			{
				Statement: `CREATE TABLE constraint_rename_test (a int CONSTRAINT con1 CHECK (a > 0), b int, c int);`,
			},
			{
				Statement: `\d constraint_rename_test
       Table "public.constraint_rename_test"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | integer |           |          | 
Check constraints:
    "con1" CHECK (a > 0)
CREATE TABLE constraint_rename_test2 (a int CONSTRAINT con1 CHECK (a > 0), d int) INHERITS (constraint_rename_test);`,
			},
			{
				Statement: `\d constraint_rename_test2
      Table "public.constraint_rename_test2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | integer |           |          | 
 d      | integer |           |          | 
Check constraints:
    "con1" CHECK (a > 0)
Inherits: constraint_rename_test
ALTER TABLE constraint_rename_test2 RENAME CONSTRAINT con1 TO con1foo; -- fail`,
				ErrorString: `cannot rename inherited constraint "con1"`,
			},
			{
				Statement:   `ALTER TABLE ONLY constraint_rename_test RENAME CONSTRAINT con1 TO con1foo; -- fail`,
				ErrorString: `inherited constraint "con1" must be renamed in child tables too`,
			},
			{
				Statement: `ALTER TABLE constraint_rename_test RENAME CONSTRAINT con1 TO con1foo; -- ok`,
			},
			{
				Statement: `\d constraint_rename_test
       Table "public.constraint_rename_test"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | integer |           |          | 
Check constraints:
    "con1foo" CHECK (a > 0)
Number of child tables: 1 (Use \d+ to list them.)
\d constraint_rename_test2
      Table "public.constraint_rename_test2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | integer |           |          | 
 d      | integer |           |          | 
Check constraints:
    "con1foo" CHECK (a > 0)
Inherits: constraint_rename_test
ALTER TABLE constraint_rename_test ADD CONSTRAINT con2 CHECK (b > 0) NO INHERIT;`,
			},
			{
				Statement: `ALTER TABLE ONLY constraint_rename_test RENAME CONSTRAINT con2 TO con2foo; -- ok`,
			},
			{
				Statement: `ALTER TABLE constraint_rename_test RENAME CONSTRAINT con2foo TO con2bar; -- ok`,
			},
			{
				Statement: `\d constraint_rename_test
       Table "public.constraint_rename_test"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | integer |           |          | 
Check constraints:
    "con1foo" CHECK (a > 0)
    "con2bar" CHECK (b > 0) NO INHERIT
Number of child tables: 1 (Use \d+ to list them.)
\d constraint_rename_test2
      Table "public.constraint_rename_test2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | integer |           |          | 
 d      | integer |           |          | 
Check constraints:
    "con1foo" CHECK (a > 0)
Inherits: constraint_rename_test
ALTER TABLE constraint_rename_test ADD CONSTRAINT con3 PRIMARY KEY (a);`,
			},
			{
				Statement: `ALTER TABLE constraint_rename_test RENAME CONSTRAINT con3 TO con3foo; -- ok`,
			},
			{
				Statement: `\d constraint_rename_test
       Table "public.constraint_rename_test"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
 b      | integer |           |          | 
 c      | integer |           |          | 
Indexes:
    "con3foo" PRIMARY KEY, btree (a)
Check constraints:
    "con1foo" CHECK (a > 0)
    "con2bar" CHECK (b > 0) NO INHERIT
Number of child tables: 1 (Use \d+ to list them.)
\d constraint_rename_test2
      Table "public.constraint_rename_test2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
 b      | integer |           |          | 
 c      | integer |           |          | 
 d      | integer |           |          | 
Check constraints:
    "con1foo" CHECK (a > 0)
Inherits: constraint_rename_test
DROP TABLE constraint_rename_test2;`,
			},
			{
				Statement: `DROP TABLE constraint_rename_test;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS constraint_not_exist RENAME CONSTRAINT con3 TO con3foo; -- ok`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS constraint_rename_test ADD CONSTRAINT con4 UNIQUE (a);`,
			},
			{
				Statement: `CREATE TABLE constraint_rename_cache (a int,
  CONSTRAINT chk_a CHECK (a > 0),
  PRIMARY KEY (a));`,
			},
			{
				Statement: `ALTER TABLE constraint_rename_cache
  RENAME CONSTRAINT chk_a TO chk_a_new;`,
			},
			{
				Statement: `ALTER TABLE constraint_rename_cache
  RENAME CONSTRAINT constraint_rename_cache_pkey TO constraint_rename_pkey_new;`,
			},
			{
				Statement: `CREATE TABLE like_constraint_rename_cache
  (LIKE constraint_rename_cache INCLUDING ALL);`,
			},
			{
				Statement: `\d like_constraint_rename_cache
    Table "public.like_constraint_rename_cache"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 
Indexes:
    "like_constraint_rename_cache_pkey" PRIMARY KEY, btree (a)
Check constraints:
    "chk_a_new" CHECK (a > 0)
DROP TABLE constraint_rename_cache;`,
			},
			{
				Statement: `DROP TABLE like_constraint_rename_cache;`,
			},
			{
				Statement: `CREATE TABLE attmp2 (a int primary key);`,
			},
			{
				Statement: `CREATE TABLE attmp3 (a int, b int);`,
			},
			{
				Statement: `CREATE TABLE attmp4 (a int, b int, unique(a,b));`,
			},
			{
				Statement: `CREATE TABLE attmp5 (a int, b int);`,
			},
			{
				Statement: `INSERT INTO attmp2 values (1);`,
			},
			{
				Statement: `INSERT INTO attmp2 values (2);`,
			},
			{
				Statement: `INSERT INTO attmp2 values (3);`,
			},
			{
				Statement: `INSERT INTO attmp2 values (4);`,
			},
			{
				Statement: `INSERT INTO attmp3 values (1,10);`,
			},
			{
				Statement: `INSERT INTO attmp3 values (1,20);`,
			},
			{
				Statement: `INSERT INTO attmp3 values (5,50);`,
			},
			{
				Statement:   `ALTER TABLE attmp3 add constraint attmpconstr foreign key(c) references attmp2 match full;`,
				ErrorString: `column "c" referenced in foreign key constraint does not exist`,
			},
			{
				Statement:   `ALTER TABLE attmp3 add constraint attmpconstr foreign key(a) references attmp2(b) match full;`,
				ErrorString: `column "b" referenced in foreign key constraint does not exist`,
			},
			{
				Statement:   `ALTER TABLE attmp3 add constraint attmpconstr foreign key (a) references attmp2 match full;`,
				ErrorString: `insert or update on table "attmp3" violates foreign key constraint "attmpconstr"`,
			},
			{
				Statement: `DELETE FROM attmp3 where a=5;`,
			},
			{
				Statement: `ALTER TABLE attmp3 add constraint attmpconstr foreign key (a) references attmp2 match full;`,
			},
			{
				Statement: `ALTER TABLE attmp3 drop constraint attmpconstr;`,
			},
			{
				Statement: `INSERT INTO attmp3 values (5,50);`,
			},
			{
				Statement: `ALTER TABLE attmp3 add constraint attmpconstr foreign key (a) references attmp2 match full NOT VALID;`,
			},
			{
				Statement:   `ALTER TABLE attmp3 validate constraint attmpconstr;`,
				ErrorString: `insert or update on table "attmp3" violates foreign key constraint "attmpconstr"`,
			},
			{
				Statement: `DELETE FROM attmp3 where a=5;`,
			},
			{
				Statement: `ALTER TABLE attmp3 validate constraint attmpconstr;`,
			},
			{
				Statement: `ALTER TABLE attmp3 validate constraint attmpconstr;`,
			},
			{
				Statement:   `ALTER TABLE attmp3 ADD CONSTRAINT b_greater_than_ten CHECK (b > 10); -- fail`,
				ErrorString: `check constraint "b_greater_than_ten" of relation "attmp3" is violated by some row`,
			},
			{
				Statement: `ALTER TABLE attmp3 ADD CONSTRAINT b_greater_than_ten CHECK (b > 10) NOT VALID; -- succeeds`,
			},
			{
				Statement:   `ALTER TABLE attmp3 VALIDATE CONSTRAINT b_greater_than_ten; -- fails`,
				ErrorString: `check constraint "b_greater_than_ten" of relation "attmp3" is violated by some row`,
			},
			{
				Statement: `DELETE FROM attmp3 WHERE NOT b > 10;`,
			},
			{
				Statement: `ALTER TABLE attmp3 VALIDATE CONSTRAINT b_greater_than_ten; -- succeeds`,
			},
			{
				Statement: `ALTER TABLE attmp3 VALIDATE CONSTRAINT b_greater_than_ten; -- succeeds`,
			},
			{
				Statement: `select * from attmp3;`,
				Results:   []sql.Row{{1, 20}},
			},
			{
				Statement: `CREATE TABLE attmp6 () INHERITS (attmp3);`,
			},
			{
				Statement: `CREATE TABLE attmp7 () INHERITS (attmp3);`,
			},
			{
				Statement: `INSERT INTO attmp6 VALUES (6, 30), (7, 16);`,
			},
			{
				Statement: `ALTER TABLE attmp3 ADD CONSTRAINT b_le_20 CHECK (b <= 20) NOT VALID;`,
			},
			{
				Statement:   `ALTER TABLE attmp3 VALIDATE CONSTRAINT b_le_20;	-- fails`,
				ErrorString: `check constraint "b_le_20" of relation "attmp6" is violated by some row`,
			},
			{
				Statement: `DELETE FROM attmp6 WHERE b > 20;`,
			},
			{
				Statement: `ALTER TABLE attmp3 VALIDATE CONSTRAINT b_le_20;	-- succeeds`,
			},
			{
				Statement: `CREATE FUNCTION boo(int) RETURNS int IMMUTABLE STRICT LANGUAGE plpgsql AS $$ BEGIN RAISE NOTICE 'boo: %', $1; RETURN $1; END; $$;`,
			},
			{
				Statement: `INSERT INTO attmp7 VALUES (8, 18);`,
			},
			{
				Statement: `ALTER TABLE attmp7 ADD CONSTRAINT identity CHECK (b = boo(b));`,
			},
			{
				Statement: `ALTER TABLE attmp3 ADD CONSTRAINT IDENTITY check (b = boo(b)) NOT VALID;`,
			},
			{
				Statement: `ALTER TABLE attmp3 VALIDATE CONSTRAINT identity;`,
			},
			{
				Statement: `create table parent_noinh_convalid (a int);`,
			},
			{
				Statement: `create table child_noinh_convalid () inherits (parent_noinh_convalid);`,
			},
			{
				Statement: `insert into parent_noinh_convalid values (1);`,
			},
			{
				Statement: `insert into child_noinh_convalid values (1);`,
			},
			{
				Statement: `alter table parent_noinh_convalid add constraint check_a_is_2 check (a = 2) no inherit not valid;`,
			},
			{
				Statement:   `alter table parent_noinh_convalid validate constraint check_a_is_2;`,
				ErrorString: `check constraint "check_a_is_2" of relation "parent_noinh_convalid" is violated by some row`,
			},
			{
				Statement: `delete from only parent_noinh_convalid;`,
			},
			{
				Statement: `alter table parent_noinh_convalid validate constraint check_a_is_2;`,
			},
			{
				Statement: `select convalidated from pg_constraint where conrelid = 'parent_noinh_convalid'::regclass and conname = 'check_a_is_2';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `drop table parent_noinh_convalid, child_noinh_convalid;`,
			},
			{
				Statement:   `ALTER TABLE attmp5 add constraint attmpconstr foreign key(a) references attmp4(a) match full;`,
				ErrorString: `there is no unique constraint matching given keys for referenced table "attmp4"`,
			},
			{
				Statement: `DROP TABLE attmp7;`,
			},
			{
				Statement: `DROP TABLE attmp6;`,
			},
			{
				Statement: `DROP TABLE attmp5;`,
			},
			{
				Statement: `DROP TABLE attmp4;`,
			},
			{
				Statement: `DROP TABLE attmp3;`,
			},
			{
				Statement: `DROP TABLE attmp2;`,
			},
			{
				Statement: `set constraint_exclusion TO 'partition';`,
			},
			{
				Statement: `create table nv_parent (d date, check (false) no inherit not valid);`,
			},
			{
				Statement: `\d nv_parent
            Table "public.nv_parent"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 d      | date |           |          | 
Check constraints:
    "nv_parent_check" CHECK (false) NO INHERIT
create table nv_child_2010 () inherits (nv_parent);`,
			},
			{
				Statement: `create table nv_child_2011 () inherits (nv_parent);`,
			},
			{
				Statement: `alter table nv_child_2010 add check (d between '2010-01-01'::date and '2010-12-31'::date) not valid;`,
			},
			{
				Statement: `alter table nv_child_2011 add check (d between '2011-01-01'::date and '2011-12-31'::date) not valid;`,
			},
			{
				Statement: `explain (costs off) select * from nv_parent where d between '2011-08-01' and '2011-08-31';`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on nv_parent nv_parent_1`}, {`Filter: ((d >= '08-01-2011'::date) AND (d <= '08-31-2011'::date))`}, {`->  Seq Scan on nv_child_2010 nv_parent_2`}, {`Filter: ((d >= '08-01-2011'::date) AND (d <= '08-31-2011'::date))`}, {`->  Seq Scan on nv_child_2011 nv_parent_3`}, {`Filter: ((d >= '08-01-2011'::date) AND (d <= '08-31-2011'::date))`}},
			},
			{
				Statement: `create table nv_child_2009 (check (d between '2009-01-01'::date and '2009-12-31'::date)) inherits (nv_parent);`,
			},
			{
				Statement: `explain (costs off) select * from nv_parent where d between '2011-08-01'::date and '2011-08-31'::date;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on nv_parent nv_parent_1`}, {`Filter: ((d >= '08-01-2011'::date) AND (d <= '08-31-2011'::date))`}, {`->  Seq Scan on nv_child_2010 nv_parent_2`}, {`Filter: ((d >= '08-01-2011'::date) AND (d <= '08-31-2011'::date))`}, {`->  Seq Scan on nv_child_2011 nv_parent_3`}, {`Filter: ((d >= '08-01-2011'::date) AND (d <= '08-31-2011'::date))`}},
			},
			{
				Statement: `explain (costs off) select * from nv_parent where d between '2009-08-01'::date and '2009-08-31'::date;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on nv_parent nv_parent_1`}, {`Filter: ((d >= '08-01-2009'::date) AND (d <= '08-31-2009'::date))`}, {`->  Seq Scan on nv_child_2010 nv_parent_2`}, {`Filter: ((d >= '08-01-2009'::date) AND (d <= '08-31-2009'::date))`}, {`->  Seq Scan on nv_child_2011 nv_parent_3`}, {`Filter: ((d >= '08-01-2009'::date) AND (d <= '08-31-2009'::date))`}, {`->  Seq Scan on nv_child_2009 nv_parent_4`}, {`Filter: ((d >= '08-01-2009'::date) AND (d <= '08-31-2009'::date))`}},
			},
			{
				Statement: `alter table nv_child_2011 VALIDATE CONSTRAINT nv_child_2011_d_check;`,
			},
			{
				Statement: `explain (costs off) select * from nv_parent where d between '2009-08-01'::date and '2009-08-31'::date;`,
				Results:   []sql.Row{{`Append`}, {`->  Seq Scan on nv_parent nv_parent_1`}, {`Filter: ((d >= '08-01-2009'::date) AND (d <= '08-31-2009'::date))`}, {`->  Seq Scan on nv_child_2010 nv_parent_2`}, {`Filter: ((d >= '08-01-2009'::date) AND (d <= '08-31-2009'::date))`}, {`->  Seq Scan on nv_child_2009 nv_parent_3`}, {`Filter: ((d >= '08-01-2009'::date) AND (d <= '08-31-2009'::date))`}},
			},
			{
				Statement: `alter table nv_parent add check (d between '2001-01-01'::date and '2099-12-31'::date) not valid;`,
			},
			{
				Statement: `\d nv_child_2009
          Table "public.nv_child_2009"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 d      | date |           |          | 
Check constraints:
    "nv_child_2009_d_check" CHECK (d >= '01-01-2009'::date AND d <= '12-31-2009'::date)
    "nv_parent_d_check" CHECK (d >= '01-01-2001'::date AND d <= '12-31-2099'::date) NOT VALID
Inherits: nv_parent
CREATE TEMP TABLE PKTABLE (ptest1 int PRIMARY KEY);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES(42);`,
			},
			{
				Statement: `CREATE TEMP TABLE FKTABLE (ftest1 inet);`,
			},
			{
				Statement:   `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1) references pktable;`,
				ErrorString: `foreign key constraint "fktable_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement:   `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1) references pktable(ptest1);`,
				ErrorString: `foreign key constraint "fktable_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `CREATE TEMP TABLE FKTABLE (ftest1 int8);`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1) references pktable;`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES(42);		-- should succeed`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES(43);		-- should fail`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `CREATE TEMP TABLE FKTABLE (ftest1 numeric);`,
			},
			{
				Statement:   `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1) references pktable;`,
				ErrorString: `foreign key constraint "fktable_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TEMP TABLE PKTABLE (ptest1 numeric PRIMARY KEY);`,
			},
			{
				Statement: `INSERT INTO PKTABLE VALUES(42);`,
			},
			{
				Statement: `CREATE TEMP TABLE FKTABLE (ftest1 int);`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1) references pktable;`,
			},
			{
				Statement: `INSERT INTO FKTABLE VALUES(42);		-- should succeed`,
			},
			{
				Statement:   `INSERT INTO FKTABLE VALUES(43);		-- should fail`,
				ErrorString: `insert or update on table "fktable" violates foreign key constraint "fktable_ftest1_fkey"`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TEMP TABLE PKTABLE (ptest1 int, ptest2 inet,
                           PRIMARY KEY(ptest1, ptest2));`,
			},
			{
				Statement: `CREATE TEMP TABLE FKTABLE (ftest1 cidr, ftest2 timestamp);`,
			},
			{
				Statement:   `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1, ftest2) references pktable;`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `CREATE TEMP TABLE FKTABLE (ftest1 cidr, ftest2 timestamp);`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1, ftest2)
     references pktable(ptest1, ptest2);`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `CREATE TEMP TABLE FKTABLE (ftest1 int, ftest2 inet);`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest1, ftest2)
     references pktable(ptest2, ptest1);`,
				ErrorString: `foreign key constraint "fktable_ftest1_ftest2_fkey" cannot be implemented`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD FOREIGN KEY(ftest2, ftest1)
     references pktable(ptest1, ptest2);`,
				ErrorString: `foreign key constraint "fktable_ftest2_ftest1_fkey" cannot be implemented`,
			},
			{
				Statement: `DROP TABLE FKTABLE;`,
			},
			{
				Statement: `DROP TABLE PKTABLE;`,
			},
			{
				Statement: `CREATE TEMP TABLE PKTABLE (ptest1 int primary key);`,
			},
			{
				Statement: `CREATE TEMP TABLE FKTABLE (ftest1 int);`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD CONSTRAINT fknd FOREIGN KEY(ftest1) REFERENCES pktable
  ON DELETE CASCADE ON UPDATE NO ACTION NOT DEFERRABLE;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD CONSTRAINT fkdd FOREIGN KEY(ftest1) REFERENCES pktable
  ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY DEFERRED;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD CONSTRAINT fkdi FOREIGN KEY(ftest1) REFERENCES pktable
  ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD CONSTRAINT fknd2 FOREIGN KEY(ftest1) REFERENCES pktable
  ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY DEFERRED;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ALTER CONSTRAINT fknd2 NOT DEFERRABLE;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD CONSTRAINT fkdd2 FOREIGN KEY(ftest1) REFERENCES pktable
  ON DELETE CASCADE ON UPDATE NO ACTION NOT DEFERRABLE;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ALTER CONSTRAINT fkdd2 DEFERRABLE INITIALLY DEFERRED;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ADD CONSTRAINT fkdi2 FOREIGN KEY(ftest1) REFERENCES pktable
  ON DELETE CASCADE ON UPDATE NO ACTION NOT DEFERRABLE;`,
			},
			{
				Statement: `ALTER TABLE FKTABLE ALTER CONSTRAINT fkdi2 DEFERRABLE INITIALLY IMMEDIATE;`,
			},
			{
				Statement: `SELECT conname, tgfoid::regproc, tgtype, tgdeferrable, tginitdeferred
FROM pg_trigger JOIN pg_constraint con ON con.oid = tgconstraint
WHERE tgrelid = 'pktable'::regclass
ORDER BY 1,2,3;`,
				Results: []sql.Row{{`fkdd`, "RI_FKey_cascade_del", 9, false, false}, {`fkdd`, "RI_FKey_noaction_upd", 17, true, true}, {`fkdd2`, "RI_FKey_cascade_del", 9, false, false}, {`fkdd2`, "RI_FKey_noaction_upd", 17, true, true}, {`fkdi`, "RI_FKey_cascade_del", 9, false, false}, {`fkdi`, "RI_FKey_noaction_upd", 17, true, false}, {`fkdi2`, "RI_FKey_cascade_del", 9, false, false}, {`fkdi2`, "RI_FKey_noaction_upd", 17, true, false}, {`fknd`, "RI_FKey_cascade_del", 9, false, false}, {`fknd`, "RI_FKey_noaction_upd", 17, false, false}, {`fknd2`, "RI_FKey_cascade_del", 9, false, false}, {`fknd2`, "RI_FKey_noaction_upd", 17, false, false}},
			},
			{
				Statement: `SELECT conname, tgfoid::regproc, tgtype, tgdeferrable, tginitdeferred
FROM pg_trigger JOIN pg_constraint con ON con.oid = tgconstraint
WHERE tgrelid = 'fktable'::regclass
ORDER BY 1,2,3;`,
				Results: []sql.Row{{`fkdd`, "RI_FKey_check_ins", 5, true, true}, {`fkdd`, "RI_FKey_check_upd", 17, true, true}, {`fkdd2`, "RI_FKey_check_ins", 5, true, true}, {`fkdd2`, "RI_FKey_check_upd", 17, true, true}, {`fkdi`, "RI_FKey_check_ins", 5, true, false}, {`fkdi`, "RI_FKey_check_upd", 17, true, false}, {`fkdi2`, "RI_FKey_check_ins", 5, true, false}, {`fkdi2`, "RI_FKey_check_upd", 17, true, false}, {`fknd`, "RI_FKey_check_ins", 5, false, false}, {`fknd`, "RI_FKey_check_upd", 17, false, false}, {`fknd2`, "RI_FKey_check_ins", 5, false, false}, {`fknd2`, "RI_FKey_check_upd", 17, false, false}},
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc_test1 check (test>3);`,
			},
			{
				Statement:   `insert into atacc1 (test) values (2);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "atacc_test1"`,
			},
			{
				Statement: `insert into atacc1 (test) values (4);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement: `insert into atacc1 (test) values (2);`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_test1 check (test>3);`,
				ErrorString: `check constraint "atacc_test1" of relation "atacc1" is violated by some row`,
			},
			{
				Statement: `insert into atacc1 (test) values (4);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_test1 check (test1>3);`,
				ErrorString: `column "test1" does not exist`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int, test2 int, test3 int);`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc_test1 check (test+test2<test3*4);`,
			},
			{
				Statement:   `insert into atacc1 (test,test2,test3) values (4,4,2);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "atacc_test1"`,
			},
			{
				Statement: `insert into atacc1 (test,test2,test3) values (4,4,5);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (test int check (test>3), test2 int);`,
			},
			{
				Statement: `alter table atacc1 add check (test2>test);`,
			},
			{
				Statement:   `insert into atacc1 (test2, test) values (3, 4);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "atacc1_check"`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (test int);`,
			},
			{
				Statement: `create table atacc2 (test2 int);`,
			},
			{
				Statement: `create table atacc3 (test3 int) inherits (atacc1, atacc2);`,
			},
			{
				Statement: `alter table atacc2 add constraint foo check (test2>0);`,
			},
			{
				Statement:   `insert into atacc2 (test2) values (-3);`,
				ErrorString: `new row for relation "atacc2" violates check constraint "foo"`,
			},
			{
				Statement: `insert into atacc2 (test2) values (3);`,
			},
			{
				Statement:   `insert into atacc3 (test2) values (-3);`,
				ErrorString: `new row for relation "atacc3" violates check constraint "foo"`,
			},
			{
				Statement: `insert into atacc3 (test2) values (3);`,
			},
			{
				Statement: `drop table atacc3;`,
			},
			{
				Statement: `drop table atacc2;`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (test int);`,
			},
			{
				Statement: `create table atacc2 (test2 int);`,
			},
			{
				Statement: `create table atacc3 (test3 int) inherits (atacc1, atacc2);`,
			},
			{
				Statement: `alter table atacc3 no inherit atacc2;`,
			},
			{
				Statement:   `alter table atacc3 no inherit atacc2;`,
				ErrorString: `relation "atacc2" is not a parent of relation "atacc3"`,
			},
			{
				Statement: `insert into atacc3 (test2) values (3);`,
			},
			{
				Statement: `select test2 from atacc2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `alter table atacc2 add constraint foo check (test2>0);`,
			},
			{
				Statement:   `alter table atacc3 inherit atacc2;`,
				ErrorString: `child table is missing constraint "foo"`,
			},
			{
				Statement: `alter table atacc3 rename test2 to testx;`,
			},
			{
				Statement:   `alter table atacc3 inherit atacc2;`,
				ErrorString: `child table is missing column "test2"`,
			},
			{
				Statement: `alter table atacc3 add test2 bool;`,
			},
			{
				Statement:   `alter table atacc3 inherit atacc2;`,
				ErrorString: `child table "atacc3" has different type for column "test2"`,
			},
			{
				Statement: `alter table atacc3 drop test2;`,
			},
			{
				Statement: `alter table atacc3 add test2 int;`,
			},
			{
				Statement: `update atacc3 set test2 = 4 where test2 is null;`,
			},
			{
				Statement: `alter table atacc3 add constraint foo check (test2>0);`,
			},
			{
				Statement: `alter table atacc3 inherit atacc2;`,
			},
			{
				Statement:   `alter table atacc3 inherit atacc2;`,
				ErrorString: `relation "atacc2" would be inherited from more than once`,
			},
			{
				Statement:   `alter table atacc2 inherit atacc3;`,
				ErrorString: `circular inheritance not allowed`,
			},
			{
				Statement:   `alter table atacc2 inherit atacc2;`,
				ErrorString: `circular inheritance not allowed`,
			},
			{
				Statement: `select test2 from atacc2;`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `drop table atacc2 cascade;`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (test int);`,
			},
			{
				Statement: `create table atacc2 (test2 int) inherits (atacc1);`,
			},
			{
				Statement: `alter table atacc1 add constraint foo check (test>0) no inherit;`,
			},
			{
				Statement: `insert into atacc2 (test) values (-3);`,
			},
			{
				Statement:   `insert into atacc1 (test) values (-3);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "foo"`,
			},
			{
				Statement: `insert into atacc1 (test) values (3);`,
			},
			{
				Statement:   `alter table atacc2 add constraint foo check (test>0) no inherit;`,
				ErrorString: `check constraint "foo" of relation "atacc2" is violated by some row`,
			},
			{
				Statement: `drop table atacc2;`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int ) ;`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc_test1 unique (test);`,
			},
			{
				Statement: `insert into atacc1 (test) values (2);`,
			},
			{
				Statement:   `insert into atacc1 (test) values (2);`,
				ErrorString: `duplicate key value violates unique constraint "atacc_test1"`,
			},
			{
				Statement: `insert into atacc1 (test) values (4);`,
			},
			{
				Statement:   `alter table atacc1 alter column test type integer using 0;`,
				ErrorString: `could not create unique index "atacc_test1"`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement: `insert into atacc1 (test) values (2);`,
			},
			{
				Statement: `insert into atacc1 (test) values (2);`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_test1 unique (test);`,
				ErrorString: `could not create unique index "atacc_test1"`,
			},
			{
				Statement: `insert into atacc1 (test) values (3);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_test1 unique (test1);`,
				ErrorString: `column "test1" named in key does not exist`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int, test2 int);`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc_test1 unique (test, test2);`,
			},
			{
				Statement: `insert into atacc1 (test,test2) values (4,4);`,
			},
			{
				Statement:   `insert into atacc1 (test,test2) values (4,4);`,
				ErrorString: `duplicate key value violates unique constraint "atacc_test1"`,
			},
			{
				Statement: `insert into atacc1 (test,test2) values (4,5);`,
			},
			{
				Statement: `insert into atacc1 (test,test2) values (5,4);`,
			},
			{
				Statement: `insert into atacc1 (test,test2) values (5,5);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (test int, test2 int, unique(test));`,
			},
			{
				Statement: `alter table atacc1 add unique (test2);`,
			},
			{
				Statement: `insert into atacc1 (test2, test) values (3, 3);`,
			},
			{
				Statement:   `insert into atacc1 (test2, test) values (2, 3);`,
				ErrorString: `duplicate key value violates unique constraint "atacc1_test_key"`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( id serial, test int) ;`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc_test1 primary key (test);`,
			},
			{
				Statement: `insert into atacc1 (test) values (2);`,
			},
			{
				Statement:   `insert into atacc1 (test) values (2);`,
				ErrorString: `duplicate key value violates unique constraint "atacc_test1"`,
			},
			{
				Statement: `insert into atacc1 (test) values (4);`,
			},
			{
				Statement:   `insert into atacc1 (test) values(NULL);`,
				ErrorString: `null value in column "test" of relation "atacc1" violates not-null constraint`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_oid1 primary key(id);`,
				ErrorString: `multiple primary keys for table "atacc1" are not allowed`,
			},
			{
				Statement: `alter table atacc1 drop constraint atacc_test1 restrict;`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc_oid1 primary key(id);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement: `insert into atacc1 (test) values (2);`,
			},
			{
				Statement: `insert into atacc1 (test) values (2);`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_test1 primary key (test);`,
				ErrorString: `could not create unique index "atacc_test1"`,
			},
			{
				Statement: `insert into atacc1 (test) values (3);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement: `insert into atacc1 (test) values (NULL);`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_test1 primary key (test);`,
				ErrorString: `column "test" of relation "atacc1" contains null values`,
			},
			{
				Statement: `insert into atacc1 (test) values (3);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_test1 primary key (test1);`,
				ErrorString: `column "test1" of relation "atacc1" does not exist`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int );`,
			},
			{
				Statement: `insert into atacc1 (test) values (0);`,
			},
			{
				Statement:   `alter table atacc1 add column test2 int primary key;`,
				ErrorString: `column "test2" of relation "atacc1" contains null values`,
			},
			{
				Statement: `alter table atacc1 add column test2 int default 0 primary key;`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (a int);`,
			},
			{
				Statement: `insert into atacc1 values(1);`,
			},
			{
				Statement: `alter table atacc1
  add column b float8 not null default random(),
  add primary key(a);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (a int primary key);`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc1_fkey foreign key (a) references atacc1 (a) not valid;`,
			},
			{
				Statement: `alter table atacc1 validate constraint atacc1_fkey, alter a type bigint;`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (a bigint, b int);`,
			},
			{
				Statement: `insert into atacc1 values(1,1);`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc1_chk check(b = 1) not valid;`,
			},
			{
				Statement: `alter table atacc1 validate constraint atacc1_chk, alter a type int;`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (a bigint, b int);`,
			},
			{
				Statement: `insert into atacc1 values(1,2);`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc1_chk check(b = 1) not valid;`,
			},
			{
				Statement:   `alter table atacc1 validate constraint atacc1_chk, alter a type int;`,
				ErrorString: `check constraint "atacc1_chk" of relation "atacc1" is violated by some row`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 ( test int, test2 int);`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc_test1 primary key (test, test2);`,
			},
			{
				Statement:   `alter table atacc1 add constraint atacc_test2 primary key (test);`,
				ErrorString: `multiple primary keys for table "atacc1" are not allowed`,
			},
			{
				Statement: `insert into atacc1 (test,test2) values (4,4);`,
			},
			{
				Statement:   `insert into atacc1 (test,test2) values (4,4);`,
				ErrorString: `duplicate key value violates unique constraint "atacc_test1"`,
			},
			{
				Statement:   `insert into atacc1 (test,test2) values (NULL,3);`,
				ErrorString: `null value in column "test" of relation "atacc1" violates not-null constraint`,
			},
			{
				Statement:   `insert into atacc1 (test,test2) values (3, NULL);`,
				ErrorString: `null value in column "test2" of relation "atacc1" violates not-null constraint`,
			},
			{
				Statement:   `insert into atacc1 (test,test2) values (NULL,NULL);`,
				ErrorString: `null value in column "test" of relation "atacc1" violates not-null constraint`,
			},
			{
				Statement: `insert into atacc1 (test,test2) values (4,5);`,
			},
			{
				Statement: `insert into atacc1 (test,test2) values (5,4);`,
			},
			{
				Statement: `insert into atacc1 (test,test2) values (5,5);`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (test int, test2 int, primary key(test));`,
			},
			{
				Statement: `insert into atacc1 (test2, test) values (3, 3);`,
			},
			{
				Statement:   `insert into atacc1 (test2, test) values (2, 3);`,
				ErrorString: `duplicate key value violates unique constraint "atacc1_pkey"`,
			},
			{
				Statement:   `insert into atacc1 (test2, test) values (1, NULL);`,
				ErrorString: `null value in column "test" of relation "atacc1" violates not-null constraint`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement:   `alter table pg_class alter column relname drop not null;`,
				ErrorString: `permission denied: "pg_class" is a system catalog`,
			},
			{
				Statement:   `alter table pg_class alter relname set not null;`,
				ErrorString: `permission denied: "pg_class" is a system catalog`,
			},
			{
				Statement:   `alter table non_existent alter column bar set not null;`,
				ErrorString: `relation "non_existent" does not exist`,
			},
			{
				Statement:   `alter table non_existent alter column bar drop not null;`,
				ErrorString: `relation "non_existent" does not exist`,
			},
			{
				Statement: `create table atacc1 (test int not null);`,
			},
			{
				Statement: `alter table atacc1 add constraint "atacc1_pkey" primary key (test);`,
			},
			{
				Statement:   `alter table atacc1 alter column test drop not null;`,
				ErrorString: `column "test" is in a primary key`,
			},
			{
				Statement: `alter table atacc1 drop constraint "atacc1_pkey";`,
			},
			{
				Statement: `alter table atacc1 alter column test drop not null;`,
			},
			{
				Statement: `insert into atacc1 values (null);`,
			},
			{
				Statement:   `alter table atacc1 alter test set not null;`,
				ErrorString: `column "test" of relation "atacc1" contains null values`,
			},
			{
				Statement: `delete from atacc1;`,
			},
			{
				Statement: `alter table atacc1 alter test set not null;`,
			},
			{
				Statement:   `alter table atacc1 alter bar set not null;`,
				ErrorString: `column "bar" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter bar drop not null;`,
				ErrorString: `column "bar" of relation "atacc1" does not exist`,
			},
			{
				Statement: `create view myview as select * from atacc1;`,
			},
			{
				Statement:   `alter table myview alter column test drop not null;`,
				ErrorString: `ALTER action ALTER COLUMN ... DROP NOT NULL cannot be performed on relation "myview"`,
			},
			{
				Statement:   `alter table myview alter column test set not null;`,
				ErrorString: `ALTER action ALTER COLUMN ... SET NOT NULL cannot be performed on relation "myview"`,
			},
			{
				Statement: `drop view myview;`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (test_a int, test_b int);`,
			},
			{
				Statement: `insert into atacc1 values (null, 1);`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc1_constr_or check(test_a is not null or test_b < 10);`,
			},
			{
				Statement:   `alter table atacc1 alter test_a set not null;`,
				ErrorString: `column "test_a" of relation "atacc1" contains null values`,
			},
			{
				Statement: `alter table atacc1 drop constraint atacc1_constr_or;`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc1_constr_invalid check(test_a is not null) not valid;`,
			},
			{
				Statement:   `alter table atacc1 alter test_a set not null;`,
				ErrorString: `column "test_a" of relation "atacc1" contains null values`,
			},
			{
				Statement: `alter table atacc1 drop constraint atacc1_constr_invalid;`,
			},
			{
				Statement: `update atacc1 set test_a = 1;`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc1_constr_a_valid check(test_a is not null);`,
			},
			{
				Statement: `alter table atacc1 alter test_a set not null;`,
			},
			{
				Statement: `delete from atacc1;`,
			},
			{
				Statement: `insert into atacc1 values (2, null);`,
			},
			{
				Statement: `alter table atacc1 alter test_a drop not null;`,
			},
			{
				Statement:   `alter table atacc1 alter test_a set not null, alter test_b set not null;`,
				ErrorString: `column "test_b" of relation "atacc1" contains null values`,
			},
			{
				Statement:   `alter table atacc1 alter test_b set not null, alter test_a set not null;`,
				ErrorString: `column "test_b" of relation "atacc1" contains null values`,
			},
			{
				Statement: `update atacc1 set test_b = 1;`,
			},
			{
				Statement: `alter table atacc1 alter test_b set not null, alter test_a set not null;`,
			},
			{
				Statement: `alter table atacc1 alter test_a drop not null, alter test_b drop not null;`,
			},
			{
				Statement: `alter table atacc1 add constraint atacc1_constr_b_valid check(test_b is not null);`,
			},
			{
				Statement: `alter table atacc1 alter test_b set not null, alter test_a set not null;`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table parent (a int);`,
			},
			{
				Statement: `create table child (b varchar(255)) inherits (parent);`,
			},
			{
				Statement: `alter table parent alter a set not null;`,
			},
			{
				Statement:   `insert into parent values (NULL);`,
				ErrorString: `null value in column "a" of relation "parent" violates not-null constraint`,
			},
			{
				Statement:   `insert into child (a, b) values (NULL, 'foo');`,
				ErrorString: `null value in column "a" of relation "child" violates not-null constraint`,
			},
			{
				Statement: `alter table parent alter a drop not null;`,
			},
			{
				Statement: `insert into parent values (NULL);`,
			},
			{
				Statement: `insert into child (a, b) values (NULL, 'foo');`,
			},
			{
				Statement:   `alter table only parent alter a set not null;`,
				ErrorString: `column "a" of relation "parent" contains null values`,
			},
			{
				Statement:   `alter table child alter a set not null;`,
				ErrorString: `column "a" of relation "child" contains null values`,
			},
			{
				Statement: `delete from parent;`,
			},
			{
				Statement: `alter table only parent alter a set not null;`,
			},
			{
				Statement:   `insert into parent values (NULL);`,
				ErrorString: `null value in column "a" of relation "parent" violates not-null constraint`,
			},
			{
				Statement: `alter table child alter a set not null;`,
			},
			{
				Statement:   `insert into child (a, b) values (NULL, 'foo');`,
				ErrorString: `null value in column "a" of relation "child" violates not-null constraint`,
			},
			{
				Statement: `delete from child;`,
			},
			{
				Statement: `alter table child alter a set not null;`,
			},
			{
				Statement:   `insert into child (a, b) values (NULL, 'foo');`,
				ErrorString: `null value in column "a" of relation "child" violates not-null constraint`,
			},
			{
				Statement: `drop table child;`,
			},
			{
				Statement: `drop table parent;`,
			},
			{
				Statement: `create table def_test (
	c1	int4 default 5,
	c2	text default 'initial_default'
);`,
			},
			{
				Statement: `insert into def_test default values;`,
			},
			{
				Statement: `alter table def_test alter column c1 drop default;`,
			},
			{
				Statement: `insert into def_test default values;`,
			},
			{
				Statement: `alter table def_test alter column c2 drop default;`,
			},
			{
				Statement: `insert into def_test default values;`,
			},
			{
				Statement: `alter table def_test alter column c1 set default 10;`,
			},
			{
				Statement: `alter table def_test alter column c2 set default 'new_default';`,
			},
			{
				Statement: `insert into def_test default values;`,
			},
			{
				Statement: `select * from def_test;`,
				Results:   []sql.Row{{5, `initial_default`}, {``, `initial_default`}, {``, ``}, {10, `new_default`}},
			},
			{
				Statement:   `alter table def_test alter column c1 set default 'wrong_datatype';`,
				ErrorString: `invalid input syntax for type integer: "wrong_datatype"`,
			},
			{
				Statement: `alter table def_test alter column c2 set default 20;`,
			},
			{
				Statement:   `alter table def_test alter column c3 set default 30;`,
				ErrorString: `column "c3" of relation "def_test" does not exist`,
			},
			{
				Statement: `create view def_view_test as select * from def_test;`,
			},
			{
				Statement: `create rule def_view_test_ins as
	on insert to def_view_test
	do instead insert into def_test select new.*;`,
			},
			{
				Statement: `insert into def_view_test default values;`,
			},
			{
				Statement: `alter table def_view_test alter column c1 set default 45;`,
			},
			{
				Statement: `insert into def_view_test default values;`,
			},
			{
				Statement: `alter table def_view_test alter column c2 set default 'view_default';`,
			},
			{
				Statement: `insert into def_view_test default values;`,
			},
			{
				Statement: `select * from def_view_test;`,
				Results:   []sql.Row{{5, `initial_default`}, {``, `initial_default`}, {``, ``}, {10, `new_default`}, {``, ``}, {45, ``}, {45, `view_default`}},
			},
			{
				Statement: `drop rule def_view_test_ins on def_view_test;`,
			},
			{
				Statement: `drop view def_view_test;`,
			},
			{
				Statement: `drop table def_test;`,
			},
			{
				Statement:   `alter table pg_class drop column relname;`,
				ErrorString: `permission denied: "pg_class" is a system catalog`,
			},
			{
				Statement:   `alter table nosuchtable drop column bar;`,
				ErrorString: `relation "nosuchtable" does not exist`,
			},
			{
				Statement: `create table atacc1 (a int4 not null, b int4, c int4 not null, d int4);`,
			},
			{
				Statement: `insert into atacc1 values (1, 2, 3, 4);`,
			},
			{
				Statement: `alter table atacc1 drop a;`,
			},
			{
				Statement:   `alter table atacc1 drop a;`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement: `select * from atacc1;`,
				Results:   []sql.Row{{2, 3, 4}},
			},
			{
				Statement:   `select * from atacc1 order by a;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `select * from atacc1 order by "........pg.dropped.1........";`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement:   `select * from atacc1 group by a;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `select * from atacc1 group by "........pg.dropped.1........";`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement: `select atacc1.* from atacc1;`,
				Results:   []sql.Row{{2, 3, 4}},
			},
			{
				Statement:   `select a from atacc1;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `select atacc1.a from atacc1;`,
				ErrorString: `column atacc1.a does not exist`,
			},
			{
				Statement: `select b,c,d from atacc1;`,
				Results:   []sql.Row{{2, 3, 4}},
			},
			{
				Statement:   `select a,b,c,d from atacc1;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `select * from atacc1 where a = 1;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `select "........pg.dropped.1........" from atacc1;`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement:   `select atacc1."........pg.dropped.1........" from atacc1;`,
				ErrorString: `column atacc1.........pg.dropped.1........ does not exist`,
			},
			{
				Statement:   `select "........pg.dropped.1........",b,c,d from atacc1;`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement:   `select * from atacc1 where "........pg.dropped.1........" = 1;`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement:   `update atacc1 set a = 3;`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `update atacc1 set b = 2 where a = 3;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `update atacc1 set "........pg.dropped.1........" = 3;`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `update atacc1 set b = 2 where "........pg.dropped.1........" = 3;`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement:   `insert into atacc1 values (10, 11, 12, 13);`,
				ErrorString: `INSERT has more expressions than target columns`,
			},
			{
				Statement:   `insert into atacc1 values (default, 11, 12, 13);`,
				ErrorString: `INSERT has more expressions than target columns`,
			},
			{
				Statement: `insert into atacc1 values (11, 12, 13);`,
			},
			{
				Statement:   `insert into atacc1 (a) values (10);`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `insert into atacc1 (a) values (default);`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `insert into atacc1 (a,b,c,d) values (10,11,12,13);`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `insert into atacc1 (a,b,c,d) values (default,11,12,13);`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement: `insert into atacc1 (b,c,d) values (11,12,13);`,
			},
			{
				Statement:   `insert into atacc1 ("........pg.dropped.1........") values (10);`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `insert into atacc1 ("........pg.dropped.1........") values (default);`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `insert into atacc1 ("........pg.dropped.1........",b,c,d) values (10,11,12,13);`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `insert into atacc1 ("........pg.dropped.1........",b,c,d) values (default,11,12,13);`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `delete from atacc1 where a = 3;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `delete from atacc1 where "........pg.dropped.1........" = 3;`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement: `delete from atacc1;`,
			},
			{
				Statement:   `alter table atacc1 drop bar;`,
				ErrorString: `column "bar" of relation "atacc1" does not exist`,
			},
			{
				Statement: `alter table atacc1 SET WITHOUT OIDS;`,
			},
			{
				Statement:   `alter table atacc1 SET WITH OIDS;`,
				ErrorString: `syntax error at or near "WITH"`,
			},
			{
				Statement:   `alter table atacc1 drop xmin;`,
				ErrorString: `cannot drop system column "xmin"`,
			},
			{
				Statement: `create view myview as select * from atacc1;`,
			},
			{
				Statement: `select * from myview;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `alter table myview drop d;`,
				ErrorString: `ALTER action DROP COLUMN cannot be performed on relation "myview"`,
			},
			{
				Statement: `drop view myview;`,
			},
			{
				Statement:   `analyze atacc1(a);`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `analyze atacc1("........pg.dropped.1........");`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `vacuum analyze atacc1(a);`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `vacuum analyze atacc1("........pg.dropped.1........");`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `comment on column atacc1.a is 'testing';`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `comment on column atacc1."........pg.dropped.1........" is 'testing';`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter a set storage plain;`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter "........pg.dropped.1........" set storage plain;`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter a set statistics 0;`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter "........pg.dropped.1........" set statistics 0;`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter a set default 3;`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter "........pg.dropped.1........" set default 3;`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter a drop default;`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter "........pg.dropped.1........" drop default;`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter a set not null;`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter "........pg.dropped.1........" set not null;`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter a drop not null;`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 alter "........pg.dropped.1........" drop not null;`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 rename a to x;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `alter table atacc1 rename "........pg.dropped.1........" to x;`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement:   `alter table atacc1 add primary key(a);`,
				ErrorString: `column "a" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 add primary key("........pg.dropped.1........");`,
				ErrorString: `column "........pg.dropped.1........" of relation "atacc1" does not exist`,
			},
			{
				Statement:   `alter table atacc1 add unique(a);`,
				ErrorString: `column "a" named in key does not exist`,
			},
			{
				Statement:   `alter table atacc1 add unique("........pg.dropped.1........");`,
				ErrorString: `column "........pg.dropped.1........" named in key does not exist`,
			},
			{
				Statement:   `alter table atacc1 add check (a > 3);`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `alter table atacc1 add check ("........pg.dropped.1........" > 3);`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement: `create table atacc2 (id int4 unique);`,
			},
			{
				Statement:   `alter table atacc1 add foreign key (a) references atacc2(id);`,
				ErrorString: `column "a" referenced in foreign key constraint does not exist`,
			},
			{
				Statement:   `alter table atacc1 add foreign key ("........pg.dropped.1........") references atacc2(id);`,
				ErrorString: `column "........pg.dropped.1........" referenced in foreign key constraint does not exist`,
			},
			{
				Statement:   `alter table atacc2 add foreign key (id) references atacc1(a);`,
				ErrorString: `column "a" referenced in foreign key constraint does not exist`,
			},
			{
				Statement:   `alter table atacc2 add foreign key (id) references atacc1("........pg.dropped.1........");`,
				ErrorString: `column "........pg.dropped.1........" referenced in foreign key constraint does not exist`,
			},
			{
				Statement: `drop table atacc2;`,
			},
			{
				Statement:   `create index "testing_idx" on atacc1(a);`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `create index "testing_idx" on atacc1("........pg.dropped.1........");`,
				ErrorString: `column "........pg.dropped.1........" does not exist`,
			},
			{
				Statement: `insert into atacc1 values (21, 22, 23);`,
			},
			{
				Statement: `create table attest1 as select * from atacc1;`,
			},
			{
				Statement: `select * from attest1;`,
				Results:   []sql.Row{{21, 22, 23}},
			},
			{
				Statement: `drop table attest1;`,
			},
			{
				Statement: `select * into attest2 from atacc1;`,
			},
			{
				Statement: `select * from attest2;`,
				Results:   []sql.Row{{21, 22, 23}},
			},
			{
				Statement: `drop table attest2;`,
			},
			{
				Statement: `alter table atacc1 drop c;`,
			},
			{
				Statement: `alter table atacc1 drop d;`,
			},
			{
				Statement: `alter table atacc1 drop b;`,
			},
			{
				Statement: `select * from atacc1;`,
			},
			{
				Statement: `(1 row)
drop table atacc1;`,
			},
			{
				Statement: `create table atacc1 (id serial primary key, value int check (value < 10));`,
			},
			{
				Statement:   `insert into atacc1(value) values (100);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "atacc1_value_check"`,
			},
			{
				Statement: `alter table atacc1 drop column value;`,
			},
			{
				Statement: `alter table atacc1 add column value int check (value < 10);`,
			},
			{
				Statement:   `insert into atacc1(value) values (100);`,
				ErrorString: `new row for relation "atacc1" violates check constraint "atacc1_value_check"`,
			},
			{
				Statement:   `insert into atacc1(id, value) values (null, 0);`,
				ErrorString: `null value in column "id" of relation "atacc1" violates not-null constraint`,
			},
			{
				Statement: `drop table atacc1;`,
			},
			{
				Statement: `create table parent (a int, b int, c int);`,
			},
			{
				Statement: `insert into parent values (1, 2, 3);`,
			},
			{
				Statement: `alter table parent drop a;`,
			},
			{
				Statement: `create table child (d varchar(255)) inherits (parent);`,
			},
			{
				Statement: `insert into child values (12, 13, 'testing');`,
			},
			{
				Statement: `select * from parent;`,
				Results:   []sql.Row{{2, 3}, {12, 13}},
			},
			{
				Statement: `select * from child;`,
				Results:   []sql.Row{{12, 13, `testing`}},
			},
			{
				Statement: `alter table parent drop c;`,
			},
			{
				Statement: `select * from parent;`,
				Results:   []sql.Row{{2}, {12}},
			},
			{
				Statement: `select * from child;`,
				Results:   []sql.Row{{12, `testing`}},
			},
			{
				Statement: `drop table child;`,
			},
			{
				Statement: `drop table parent;`,
			},
			{
				Statement: `create table parent (a float8, b numeric(10,4), c text collate "C");`,
			},
			{
				Statement:   `create table child (a float4) inherits (parent); -- fail`,
				ErrorString: `column "a" has a type conflict`,
			},
			{
				Statement:   `create table child (b decimal(10,7)) inherits (parent); -- fail`,
				ErrorString: `column "b" has a type conflict`,
			},
			{
				Statement:   `create table child (c text collate "POSIX") inherits (parent); -- fail`,
				ErrorString: `column "c" has a collation conflict`,
			},
			{
				Statement: `create table child (a double precision, b decimal(10,4)) inherits (parent);`,
			},
			{
				Statement: `drop table child;`,
			},
			{
				Statement: `drop table parent;`,
			},
			{
				Statement: `create table attest (a int4, b int4, c int4);`,
			},
			{
				Statement: `insert into attest values (1,2,3);`,
			},
			{
				Statement: `alter table attest drop a;`,
			},
			{
				Statement: `copy attest to stdout;`,
			},
			{
				Statement: `2	3
copy attest(a) to stdout;`,
				ErrorString: `column "a" of relation "attest" does not exist`,
			},
			{
				Statement:   `copy attest("........pg.dropped.1........") to stdout;`,
				ErrorString: `column "........pg.dropped.1........" of relation "attest" does not exist`,
			},
			{
				Statement:   `copy attest from stdin;`,
				ErrorString: `extra data after last expected column`,
			},
			{
				Statement: `CONTEXT:  COPY attest, line 1: "10	11	12"
select * from attest;`,
				Results: []sql.Row{{2, 3}},
			},
			{
				Statement: `copy attest from stdin;`,
			},
			{
				Statement: `select * from attest;`,
				Results:   []sql.Row{{2, 3}, {21, 22}},
			},
			{
				Statement:   `copy attest(a) from stdin;`,
				ErrorString: `column "a" of relation "attest" does not exist`,
			},
			{
				Statement:   `copy attest("........pg.dropped.1........") from stdin;`,
				ErrorString: `column "........pg.dropped.1........" of relation "attest" does not exist`,
			},
			{
				Statement: `copy attest(b,c) from stdin;`,
			},
			{
				Statement: `select * from attest;`,
				Results:   []sql.Row{{2, 3}, {21, 22}, {31, 32}},
			},
			{
				Statement: `drop table attest;`,
			},
			{
				Statement: `create table dropColumn (a int, b int, e int);`,
			},
			{
				Statement: `create table dropColumnChild (c int) inherits (dropColumn);`,
			},
			{
				Statement: `create table dropColumnAnother (d int) inherits (dropColumnChild);`,
			},
			{
				Statement:   `alter table dropColumnchild drop column a;`,
				ErrorString: `cannot drop inherited column "a"`,
			},
			{
				Statement:   `alter table only dropColumnChild drop column b;`,
				ErrorString: `cannot drop inherited column "b"`,
			},
			{
				Statement: `alter table only dropColumn drop column e;`,
			},
			{
				Statement: `alter table dropColumnChild drop column c;`,
			},
			{
				Statement: `alter table dropColumn drop column a;`,
			},
			{
				Statement: `create table renameColumn (a int);`,
			},
			{
				Statement: `create table renameColumnChild (b int) inherits (renameColumn);`,
			},
			{
				Statement: `create table renameColumnAnother (c int) inherits (renameColumnChild);`,
			},
			{
				Statement:   `alter table renameColumnChild rename column a to d;`,
				ErrorString: `cannot rename inherited column "a"`,
			},
			{
				Statement:   `alter table only renameColumnChild rename column a to d;`,
				ErrorString: `inherited column "a" must be renamed in child tables too`,
			},
			{
				Statement:   `alter table only renameColumn rename column a to d;`,
				ErrorString: `inherited column "a" must be renamed in child tables too`,
			},
			{
				Statement: `alter table renameColumn rename column a to d;`,
			},
			{
				Statement: `alter table renameColumnChild rename column b to a;`,
			},
			{
				Statement: `alter table if exists doesnt_exist_tab rename column a to d;`,
			},
			{
				Statement: `alter table if exists doesnt_exist_tab rename column b to a;`,
			},
			{
				Statement: `alter table renameColumn add column w int;`,
			},
			{
				Statement:   `alter table only renameColumn add column x int;`,
				ErrorString: `column must be added to child tables too`,
			},
			{
				Statement: `create table p1 (f1 int, f2 int);`,
			},
			{
				Statement: `create table c1 (f1 int not null) inherits(p1);`,
			},
			{
				Statement:   `alter table c1 drop column f1;`,
				ErrorString: `cannot drop inherited column "f1"`,
			},
			{
				Statement: `alter table p1 drop column f1;`,
			},
			{
				Statement: `select f1 from c1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `alter table c1 drop column f1;`,
			},
			{
				Statement:   `select f1 from c1;`,
				ErrorString: `column "f1" does not exist`,
			},
			{
				Statement: `drop table p1 cascade;`,
			},
			{
				Statement: `create table p1 (f1 int, f2 int);`,
			},
			{
				Statement: `create table c1 () inherits(p1);`,
			},
			{
				Statement:   `alter table c1 drop column f1;`,
				ErrorString: `cannot drop inherited column "f1"`,
			},
			{
				Statement: `alter table p1 drop column f1;`,
			},
			{
				Statement:   `select f1 from c1;`,
				ErrorString: `column "f1" does not exist`,
			},
			{
				Statement: `drop table p1 cascade;`,
			},
			{
				Statement: `create table p1 (f1 int, f2 int);`,
			},
			{
				Statement: `create table c1 () inherits(p1);`,
			},
			{
				Statement:   `alter table c1 drop column f1;`,
				ErrorString: `cannot drop inherited column "f1"`,
			},
			{
				Statement: `alter table only p1 drop column f1;`,
			},
			{
				Statement: `alter table c1 drop column f1;`,
			},
			{
				Statement: `drop table p1 cascade;`,
			},
			{
				Statement: `create table p1 (f1 int, f2 int);`,
			},
			{
				Statement: `create table c1 (f1 int not null) inherits(p1);`,
			},
			{
				Statement:   `alter table c1 drop column f1;`,
				ErrorString: `cannot drop inherited column "f1"`,
			},
			{
				Statement: `alter table only p1 drop column f1;`,
			},
			{
				Statement: `alter table c1 drop column f1;`,
			},
			{
				Statement: `drop table p1 cascade;`,
			},
			{
				Statement: `create table p1(id int, name text);`,
			},
			{
				Statement: `create table p2(id2 int, name text, height int);`,
			},
			{
				Statement: `create table c1(age int) inherits(p1,p2);`,
			},
			{
				Statement: `create table gc1() inherits (c1);`,
			},
			{
				Statement: `select relname, attname, attinhcount, attislocal
from pg_class join pg_attribute on (pg_class.oid = pg_attribute.attrelid)
where relname in ('p1','p2','c1','gc1') and attnum > 0 and not attisdropped
order by relname, attnum;`,
				Results: []sql.Row{{`c1`, `id`, 1, false}, {`c1`, `name`, 2, false}, {`c1`, `id2`, 1, false}, {`c1`, `height`, 1, false}, {`c1`, `age`, 0, true}, {`gc1`, `id`, 1, false}, {`gc1`, `name`, 1, false}, {`gc1`, `id2`, 1, false}, {`gc1`, `height`, 1, false}, {`gc1`, `age`, 1, false}, {`p1`, `id`, 0, true}, {`p1`, `name`, 0, true}, {`p2`, `id2`, 0, true}, {`p2`, `name`, 0, true}, {`p2`, `height`, 0, true}},
			},
			{
				Statement: `alter table only p1 drop column name;`,
			},
			{
				Statement: `alter table p2 drop column name;`,
			},
			{
				Statement:   `alter table gc1 drop column name;`,
				ErrorString: `cannot drop inherited column "name"`,
			},
			{
				Statement: `alter table c1 drop column name;`,
			},
			{
				Statement:   `alter table gc1 drop column name;`,
				ErrorString: `column "name" of relation "gc1" does not exist`,
			},
			{
				Statement: `alter table p2 drop column height;`,
			},
			{
				Statement: `create table dropColumnExists ();`,
			},
			{
				Statement: `alter table dropColumnExists drop column non_existing; --fail
ERROR:  column "non_existing" of relation "dropcolumnexists" does not exist
alter table dropColumnExists drop column if exists non_existing; --succeed
select relname, attname, attinhcount, attislocal
from pg_class join pg_attribute on (pg_class.oid = pg_attribute.attrelid)
where relname in ('p1','p2','c1','gc1') and attnum > 0 and not attisdropped
order by relname, attnum;`,
				Results: []sql.Row{{`c1`, `id`, 1, false}, {`c1`, `id2`, 1, false}, {`c1`, `age`, 0, true}, {`gc1`, `id`, 1, false}, {`gc1`, `id2`, 1, false}, {`gc1`, `age`, 1, false}, {`p1`, `id`, 0, true}, {`p2`, `id2`, 0, true}},
			},
			{
				Statement: `drop table p1, p2 cascade;`,
			},
			{
				Statement: `create table depth0();`,
			},
			{
				Statement: `create table depth1(c text) inherits (depth0);`,
			},
			{
				Statement: `create table depth2() inherits (depth1);`,
			},
			{
				Statement: `alter table depth0 add c text;`,
			},
			{
				Statement: `select attrelid::regclass, attname, attinhcount, attislocal
from pg_attribute
where attnum > 0 and attrelid::regclass in ('depth0', 'depth1', 'depth2')
order by attrelid::regclass::text, attnum;`,
				Results: []sql.Row{{`depth0`, `c`, 0, true}, {`depth1`, `c`, 1, true}, {`depth2`, `c`, 1, false}},
			},
			{
				Statement: `create table p1 (f1 int);`,
			},
			{
				Statement: `create table c1 (f2 text, f3 int) inherits (p1);`,
			},
			{
				Statement: `alter table p1 add column a1 int check (a1 > 0);`,
			},
			{
				Statement: `alter table p1 add column f2 text;`,
			},
			{
				Statement: `insert into p1 values (1,2,'abc');`,
			},
			{
				Statement:   `insert into c1 values(11,'xyz',33,0); -- should fail`,
				ErrorString: `new row for relation "c1" violates check constraint "p1_a1_check"`,
			},
			{
				Statement: `insert into c1 values(11,'xyz',33,22);`,
			},
			{
				Statement: `select * from p1;`,
				Results:   []sql.Row{{1, 2, `abc`}, {11, 22, `xyz`}},
			},
			{
				Statement: `update p1 set a1 = a1 + 1, f2 = upper(f2);`,
			},
			{
				Statement: `select * from p1;`,
				Results:   []sql.Row{{1, 3, `ABC`}, {11, 23, `XYZ`}},
			},
			{
				Statement: `drop table p1 cascade;`,
			},
			{
				Statement: `create domain mytype as text;`,
			},
			{
				Statement: `create temp table foo (f1 text, f2 mytype, f3 text);`,
			},
			{
				Statement: `insert into foo values('bb','cc','dd');`,
			},
			{
				Statement: `select * from foo;`,
				Results:   []sql.Row{{`bb`, `cc`, `dd`}},
			},
			{
				Statement: `drop domain mytype cascade;`,
			},
			{
				Statement: `select * from foo;`,
				Results:   []sql.Row{{`bb`, `dd`}},
			},
			{
				Statement: `insert into foo values('qq','rr');`,
			},
			{
				Statement: `select * from foo;`,
				Results:   []sql.Row{{`bb`, `dd`}, {`qq`, `rr`}},
			},
			{
				Statement: `update foo set f3 = 'zz';`,
			},
			{
				Statement: `select * from foo;`,
				Results:   []sql.Row{{`bb`, `zz`}, {`qq`, `zz`}},
			},
			{
				Statement: `select f3,max(f1) from foo group by f3;`,
				Results:   []sql.Row{{`zz`, `qq`}},
			},
			{
				Statement:   `alter table foo alter f1 TYPE integer; -- fails`,
				ErrorString: `column "f1" cannot be cast automatically to type integer`,
			},
			{
				Statement: `alter table foo alter f1 TYPE varchar(10);`,
			},
			{
				Statement: `create table anothertab (atcol1 serial8, atcol2 boolean,
	constraint anothertab_chk check (atcol1 <= 3));`,
			},
			{
				Statement: `insert into anothertab (atcol1, atcol2) values (default, true);`,
			},
			{
				Statement: `insert into anothertab (atcol1, atcol2) values (default, false);`,
			},
			{
				Statement: `select * from anothertab;`,
				Results:   []sql.Row{{1, true}, {2, false}},
			},
			{
				Statement:   `alter table anothertab alter column atcol1 type boolean; -- fails`,
				ErrorString: `column "atcol1" cannot be cast automatically to type boolean`,
			},
			{
				Statement:   `alter table anothertab alter column atcol1 type boolean using atcol1::int; -- fails`,
				ErrorString: `result of USING clause for column "atcol1" cannot be cast automatically to type boolean`,
			},
			{
				Statement: `alter table anothertab alter column atcol1 type integer;`,
			},
			{
				Statement: `select * from anothertab;`,
				Results:   []sql.Row{{1, true}, {2, false}},
			},
			{
				Statement:   `insert into anothertab (atcol1, atcol2) values (45, null); -- fails`,
				ErrorString: `new row for relation "anothertab" violates check constraint "anothertab_chk"`,
			},
			{
				Statement: `insert into anothertab (atcol1, atcol2) values (default, null);`,
			},
			{
				Statement: `select * from anothertab;`,
				Results:   []sql.Row{{1, true}, {2, false}, {3, ``}},
			},
			{
				Statement: `alter table anothertab alter column atcol2 type text
      using case when atcol2 is true then 'IT WAS TRUE'
                 when atcol2 is false then 'IT WAS FALSE'
                 else 'IT WAS NULL!' end;`,
			},
			{
				Statement: `select * from anothertab;`,
				Results:   []sql.Row{{1, `IT WAS TRUE`}, {2, `IT WAS FALSE`}, {3, `IT WAS NULL!`}},
			},
			{
				Statement: `alter table anothertab alter column atcol1 type boolean
        using case when atcol1 % 2 = 0 then true else false end; -- fails`,
				ErrorString: `default for column "atcol1" cannot be cast automatically to type boolean`,
			},
			{
				Statement: `alter table anothertab alter column atcol1 drop default;`,
			},
			{
				Statement: `alter table anothertab alter column atcol1 type boolean
        using case when atcol1 % 2 = 0 then true else false end; -- fails`,
				ErrorString: `operator does not exist: boolean <= integer`,
			},
			{
				Statement: `alter table anothertab drop constraint anothertab_chk;`,
			},
			{
				Statement:   `alter table anothertab drop constraint anothertab_chk; -- fails`,
				ErrorString: `constraint "anothertab_chk" of relation "anothertab" does not exist`,
			},
			{
				Statement: `alter table anothertab drop constraint IF EXISTS anothertab_chk; -- succeeds`,
			},
			{
				Statement: `alter table anothertab alter column atcol1 type boolean
        using case when atcol1 % 2 = 0 then true else false end;`,
			},
			{
				Statement: `select * from anothertab;`,
				Results:   []sql.Row{{false, `IT WAS TRUE`}, {true, `IT WAS FALSE`}, {false, `IT WAS NULL!`}},
			},
			{
				Statement: `drop table anothertab;`,
			},
			{
				Statement: `create table anothertab(f1 int primary key, f2 int unique,
                        f3 int, f4 int, f5 int);`,
			},
			{
				Statement: `alter table anothertab
  add exclude using btree (f3 with =);`,
			},
			{
				Statement: `alter table anothertab
  add exclude using btree (f4 with =) where (f4 is not null);`,
			},
			{
				Statement: `alter table anothertab
  add exclude using btree (f4 with =) where (f5 > 0);`,
			},
			{
				Statement: `alter table anothertab
  add unique(f1,f4);`,
			},
			{
				Statement: `create index on anothertab(f2,f3);`,
			},
			{
				Statement: `create unique index on anothertab(f4);`,
			},
			{
				Statement: `\d anothertab
             Table "public.anothertab"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 f1     | integer |           | not null | 
 f2     | integer |           |          | 
 f3     | integer |           |          | 
 f4     | integer |           |          | 
 f5     | integer |           |          | 
Indexes:
    "anothertab_pkey" PRIMARY KEY, btree (f1)
    "anothertab_f1_f4_key" UNIQUE CONSTRAINT, btree (f1, f4)
    "anothertab_f2_f3_idx" btree (f2, f3)
    "anothertab_f2_key" UNIQUE CONSTRAINT, btree (f2)
    "anothertab_f3_excl" EXCLUDE USING btree (f3 WITH =)
    "anothertab_f4_excl" EXCLUDE USING btree (f4 WITH =) WHERE (f4 IS NOT NULL)
    "anothertab_f4_excl1" EXCLUDE USING btree (f4 WITH =) WHERE (f5 > 0)
    "anothertab_f4_idx" UNIQUE, btree (f4)
alter table anothertab alter column f1 type bigint;`,
			},
			{
				Statement: `alter table anothertab
  alter column f2 type bigint,
  alter column f3 type bigint,
  alter column f4 type bigint;`,
			},
			{
				Statement: `alter table anothertab alter column f5 type bigint;`,
			},
			{
				Statement: `\d anothertab
            Table "public.anothertab"
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 f1     | bigint |           | not null | 
 f2     | bigint |           |          | 
 f3     | bigint |           |          | 
 f4     | bigint |           |          | 
 f5     | bigint |           |          | 
Indexes:
    "anothertab_pkey" PRIMARY KEY, btree (f1)
    "anothertab_f1_f4_key" UNIQUE CONSTRAINT, btree (f1, f4)
    "anothertab_f2_f3_idx" btree (f2, f3)
    "anothertab_f2_key" UNIQUE CONSTRAINT, btree (f2)
    "anothertab_f3_excl" EXCLUDE USING btree (f3 WITH =)
    "anothertab_f4_excl" EXCLUDE USING btree (f4 WITH =) WHERE (f4 IS NOT NULL)
    "anothertab_f4_excl1" EXCLUDE USING btree (f4 WITH =) WHERE (f5 > 0)
    "anothertab_f4_idx" UNIQUE, btree (f4)
drop table anothertab;`,
			},
			{
				Statement: `create table another (f1 int, f2 text, f3 text);`,
			},
			{
				Statement: `insert into another values(1, 'one', 'uno');`,
			},
			{
				Statement: `insert into another values(2, 'two', 'due');`,
			},
			{
				Statement: `insert into another values(3, 'three', 'tre');`,
			},
			{
				Statement: `select * from another;`,
				Results:   []sql.Row{{1, `one`, `uno`}, {2, `two`, `due`}, {3, `three`, `tre`}},
			},
			{
				Statement: `alter table another
  alter f1 type text using f2 || ' and ' || f3 || ' more',
  alter f2 type bigint using f1 * 10,
  drop column f3;`,
			},
			{
				Statement: `select * from another;`,
				Results:   []sql.Row{{`one and uno more`, 10}, {`two and due more`, 20}, {`three and tre more`, 30}},
			},
			{
				Statement: `drop table another;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create table skip_wal_skip_rewrite_index (c varchar(10) primary key);`,
			},
			{
				Statement: `alter table skip_wal_skip_rewrite_index alter c type varchar(20);`,
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `create table at_tab1 (a int, b text);`,
			},
			{
				Statement: `create table at_tab2 (x int, y at_tab1);`,
			},
			{
				Statement:   `alter table at_tab1 alter column b type varchar; -- fails`,
				ErrorString: `cannot alter table "at_tab1" because column "at_tab2.y" uses its row type`,
			},
			{
				Statement: `drop table at_tab2;`,
			},
			{
				Statement: `create table at_tab2 (x int, y text, check((x,y)::at_tab1 = (1,'42')::at_tab1));`,
			},
			{
				Statement: `alter table at_tab1 alter column b type varchar; -- allowed, but ...`,
			},
			{
				Statement:   `insert into at_tab2 values(1,'42'); -- ... this will fail`,
				ErrorString: `ROW() column has type text instead of type character varying`,
			},
			{
				Statement: `drop table at_tab1, at_tab2;`,
			},
			{
				Statement: `create table at_tab1 (a int, b text) partition by list(a);`,
			},
			{
				Statement: `create table at_tab2 (x int, y at_tab1);`,
			},
			{
				Statement:   `alter table at_tab1 alter column b type varchar; -- fails`,
				ErrorString: `cannot alter table "at_tab1" because column "at_tab2.y" uses its row type`,
			},
			{
				Statement: `drop table at_tab1, at_tab2;`,
			},
			{
				Statement: `create table at_partitioned (a int, b text) partition by range (a);`,
			},
			{
				Statement: `create table at_part_1 partition of at_partitioned for values from (0) to (1000);`,
			},
			{
				Statement: `insert into at_partitioned values (512, '0.123');`,
			},
			{
				Statement: `create table at_part_2 (b text, a int);`,
			},
			{
				Statement: `insert into at_part_2 values ('1.234', 1024);`,
			},
			{
				Statement: `create index on at_partitioned (b);`,
			},
			{
				Statement: `create index on at_partitioned (a);`,
			},
			{
				Statement: `\d at_part_1
             Table "public.at_part_1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | text    |           |          | 
Partition of: at_partitioned FOR VALUES FROM (0) TO (1000)
Indexes:
    "at_part_1_a_idx" btree (a)
    "at_part_1_b_idx" btree (b)
\d at_part_2
             Table "public.at_part_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 b      | text    |           |          | 
 a      | integer |           |          | 
alter table at_partitioned attach partition at_part_2 for values from (1000) to (2000);`,
			},
			{
				Statement: `\d at_part_2
             Table "public.at_part_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 b      | text    |           |          | 
 a      | integer |           |          | 
Partition of: at_partitioned FOR VALUES FROM (1000) TO (2000)
Indexes:
    "at_part_2_a_idx" btree (a)
    "at_part_2_b_idx" btree (b)
alter table at_partitioned alter column b type numeric using b::numeric;`,
			},
			{
				Statement: `\d at_part_1
             Table "public.at_part_1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | numeric |           |          | 
Partition of: at_partitioned FOR VALUES FROM (0) TO (1000)
Indexes:
    "at_part_1_a_idx" btree (a)
    "at_part_1_b_idx" btree (b)
\d at_part_2
             Table "public.at_part_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 b      | numeric |           |          | 
 a      | integer |           |          | 
Partition of: at_partitioned FOR VALUES FROM (1000) TO (2000)
Indexes:
    "at_part_2_a_idx" btree (a)
    "at_part_2_b_idx" btree (b)
drop table at_partitioned;`,
			},
			{
				Statement: `create table at_partitioned(id int, name varchar(64), unique (id, name))
  partition by hash(id);`,
			},
			{
				Statement: `comment on constraint at_partitioned_id_name_key on at_partitioned is 'parent constraint';`,
			},
			{
				Statement: `comment on index at_partitioned_id_name_key is 'parent index';`,
			},
			{
				Statement: `create table at_partitioned_0 partition of at_partitioned
  for values with (modulus 2, remainder 0);`,
			},
			{
				Statement: `comment on constraint at_partitioned_0_id_name_key on at_partitioned_0 is 'child 0 constraint';`,
			},
			{
				Statement: `comment on index at_partitioned_0_id_name_key is 'child 0 index';`,
			},
			{
				Statement: `create table at_partitioned_1 partition of at_partitioned
  for values with (modulus 2, remainder 1);`,
			},
			{
				Statement: `comment on constraint at_partitioned_1_id_name_key on at_partitioned_1 is 'child 1 constraint';`,
			},
			{
				Statement: `comment on index at_partitioned_1_id_name_key is 'child 1 index';`,
			},
			{
				Statement: `insert into at_partitioned values(1, 'foo');`,
			},
			{
				Statement: `insert into at_partitioned values(3, 'bar');`,
			},
			{
				Statement: `create temp table old_oids as
  select relname, oid as oldoid, relfilenode as oldfilenode
  from pg_class where relname like 'at_partitioned%';`,
			},
			{
				Statement: `select relname,
  c.oid = oldoid as orig_oid,
  case relfilenode
    when 0 then 'none'
    when c.oid then 'own'
    when oldfilenode then 'orig'
    else 'OTHER'
    end as storage,
  obj_description(c.oid, 'pg_class') as desc
  from pg_class c left join old_oids using (relname)
  where relname like 'at_partitioned%'
  order by relname;`,
				Results: []sql.Row{{`at_partitioned`, true, `none`, ``}, {`at_partitioned_0`, true, `own`, ``}, {`at_partitioned_0_id_name_key`, true, `own`, `child 0 index`}, {`at_partitioned_1`, true, `own`, ``}, {`at_partitioned_1_id_name_key`, true, `own`, `child 1 index`}, {`at_partitioned_id_name_key`, true, `none`, `parent index`}},
			},
			{
				Statement: `select conname, obj_description(oid, 'pg_constraint') as desc
  from pg_constraint where conname like 'at_partitioned%'
  order by conname;`,
				Results: []sql.Row{{`at_partitioned_0_id_name_key`, `child 0 constraint`}, {`at_partitioned_1_id_name_key`, `child 1 constraint`}, {`at_partitioned_id_name_key`, `parent constraint`}},
			},
			{
				Statement: `alter table at_partitioned alter column name type varchar(127);`,
			},
			{
				Statement: `select relname,
  c.oid = oldoid as orig_oid,
  case relfilenode
    when 0 then 'none'
    when c.oid then 'own'
    when oldfilenode then 'orig'
    else 'OTHER'
    end as storage,
  obj_description(c.oid, 'pg_class') as desc
  from pg_class c left join old_oids using (relname)
  where relname like 'at_partitioned%'
  order by relname;`,
				Results: []sql.Row{{`at_partitioned`, true, `none`, ``}, {`at_partitioned_0`, true, `own`, ``}, {`at_partitioned_0_id_name_key`, false, `own`, `parent index`}, {`at_partitioned_1`, true, `own`, ``}, {`at_partitioned_1_id_name_key`, false, `own`, `parent index`}, {`at_partitioned_id_name_key`, false, `none`, `parent index`}},
			},
			{
				Statement: `select conname, obj_description(oid, 'pg_constraint') as desc
  from pg_constraint where conname like 'at_partitioned%'
  order by conname;`,
				Results: []sql.Row{{`at_partitioned_0_id_name_key`, ``}, {`at_partitioned_1_id_name_key`, ``}, {`at_partitioned_id_name_key`, `parent constraint`}},
			},
			{
				Statement: `drop table at_partitioned;`,
			},
			{
				Statement: `create temp table recur1 (f1 int);`,
			},
			{
				Statement:   `alter table recur1 add column f2 recur1; -- fails`,
				ErrorString: `composite type recur1 cannot be made a member of itself`,
			},
			{
				Statement:   `alter table recur1 add column f2 recur1[]; -- fails`,
				ErrorString: `composite type recur1 cannot be made a member of itself`,
			},
			{
				Statement: `create domain array_of_recur1 as recur1[];`,
			},
			{
				Statement:   `alter table recur1 add column f2 array_of_recur1; -- fails`,
				ErrorString: `composite type recur1 cannot be made a member of itself`,
			},
			{
				Statement: `create temp table recur2 (f1 int, f2 recur1);`,
			},
			{
				Statement:   `alter table recur1 add column f2 recur2; -- fails`,
				ErrorString: `composite type recur1 cannot be made a member of itself`,
			},
			{
				Statement: `alter table recur1 add column f2 int;`,
			},
			{
				Statement:   `alter table recur1 alter column f2 type recur2; -- fails`,
				ErrorString: `composite type recur1 cannot be made a member of itself`,
			},
			{
				Statement: `create table test_storage (a text);`,
			},
			{
				Statement: `select reltoastrelid <> 0 as has_toast_table
  from pg_class where oid = 'test_storage'::regclass;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `alter table test_storage alter a set storage plain;`,
			},
			{
				Statement: `alter table test_storage add b int default random()::int;`,
			},
			{
				Statement: `select reltoastrelid <> 0 as has_toast_table
  from pg_class where oid = 'test_storage'::regclass;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `alter table test_storage alter a set storage extended; -- re-add TOAST table`,
			},
			{
				Statement: `select reltoastrelid <> 0 as has_toast_table
  from pg_class where oid = 'test_storage'::regclass;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `create index test_storage_idx on test_storage (b, a);`,
			},
			{
				Statement: `alter table test_storage alter column a set storage external;`,
			},
			{
				Statement: `\d+ test_storage
                                     Table "public.test_storage"
 Column |  Type   | Collation | Nullable |      Default      | Storage  | Stats target | Description 
--------+---------+-----------+----------+-------------------+----------+--------------+-------------
 a      | text    |           |          |                   | external |              | 
 b      | integer |           |          | random()::integer | plain    |              | 
Indexes:
    "test_storage_idx" btree (b, a)
\d+ test_storage_idx
                Index "public.test_storage_idx"
 Column |  Type   | Key? | Definition | Storage  | Stats target 
--------+---------+------+------------+----------+--------------
 b      | integer | yes  | b          | plain    | 
 a      | text    | yes  | a          | external | 
btree, for table "public.test_storage"
CREATE TABLE test_inh_check (a float check (a > 10.2), b float);`,
			},
			{
				Statement: `CREATE TABLE test_inh_check_child() INHERITS(test_inh_check);`,
			},
			{
				Statement: `\d test_inh_check
               Table "public.test_inh_check"
 Column |       Type       | Collation | Nullable | Default 
--------+------------------+-----------+----------+---------
 a      | double precision |           |          | 
 b      | double precision |           |          | 
Check constraints:
    "test_inh_check_a_check" CHECK (a > 10.2::double precision)
Number of child tables: 1 (Use \d+ to list them.)
\d test_inh_check_child
            Table "public.test_inh_check_child"
 Column |       Type       | Collation | Nullable | Default 
--------+------------------+-----------+----------+---------
 a      | double precision |           |          | 
 b      | double precision |           |          | 
Check constraints:
    "test_inh_check_a_check" CHECK (a > 10.2::double precision)
Inherits: test_inh_check
select relname, conname, coninhcount, conislocal, connoinherit
  from pg_constraint c, pg_class r
  where relname like 'test_inh_check%' and c.conrelid = r.oid
  order by 1, 2;`,
				Results: []sql.Row{{`test_inh_check`, `test_inh_check_a_check`, 0, true, false}, {`test_inh_check_child`, `test_inh_check_a_check`, 1, false, false}},
			},
			{
				Statement: `ALTER TABLE test_inh_check ALTER COLUMN a TYPE numeric;`,
			},
			{
				Statement: `\d test_inh_check
               Table "public.test_inh_check"
 Column |       Type       | Collation | Nullable | Default 
--------+------------------+-----------+----------+---------
 a      | numeric          |           |          | 
 b      | double precision |           |          | 
Check constraints:
    "test_inh_check_a_check" CHECK (a::double precision > 10.2::double precision)
Number of child tables: 1 (Use \d+ to list them.)
\d test_inh_check_child
            Table "public.test_inh_check_child"
 Column |       Type       | Collation | Nullable | Default 
--------+------------------+-----------+----------+---------
 a      | numeric          |           |          | 
 b      | double precision |           |          | 
Check constraints:
    "test_inh_check_a_check" CHECK (a::double precision > 10.2::double precision)
Inherits: test_inh_check
select relname, conname, coninhcount, conislocal, connoinherit
  from pg_constraint c, pg_class r
  where relname like 'test_inh_check%' and c.conrelid = r.oid
  order by 1, 2;`,
				Results: []sql.Row{{`test_inh_check`, `test_inh_check_a_check`, 0, true, false}, {`test_inh_check_child`, `test_inh_check_a_check`, 1, false, false}},
			},
			{
				Statement: `ALTER TABLE test_inh_check ADD CONSTRAINT bnoinherit CHECK (b > 100) NO INHERIT;`,
			},
			{
				Statement: `ALTER TABLE test_inh_check_child ADD CONSTRAINT blocal CHECK (b < 1000);`,
			},
			{
				Statement: `ALTER TABLE test_inh_check_child ADD CONSTRAINT bmerged CHECK (b > 1);`,
			},
			{
				Statement: `ALTER TABLE test_inh_check ADD CONSTRAINT bmerged CHECK (b > 1);`,
			},
			{
				Statement: `\d test_inh_check
               Table "public.test_inh_check"
 Column |       Type       | Collation | Nullable | Default 
--------+------------------+-----------+----------+---------
 a      | numeric          |           |          | 
 b      | double precision |           |          | 
Check constraints:
    "bmerged" CHECK (b > 1::double precision)
    "bnoinherit" CHECK (b > 100::double precision) NO INHERIT
    "test_inh_check_a_check" CHECK (a::double precision > 10.2::double precision)
Number of child tables: 1 (Use \d+ to list them.)
\d test_inh_check_child
            Table "public.test_inh_check_child"
 Column |       Type       | Collation | Nullable | Default 
--------+------------------+-----------+----------+---------
 a      | numeric          |           |          | 
 b      | double precision |           |          | 
Check constraints:
    "blocal" CHECK (b < 1000::double precision)
    "bmerged" CHECK (b > 1::double precision)
    "test_inh_check_a_check" CHECK (a::double precision > 10.2::double precision)
Inherits: test_inh_check
select relname, conname, coninhcount, conislocal, connoinherit
  from pg_constraint c, pg_class r
  where relname like 'test_inh_check%' and c.conrelid = r.oid
  order by 1, 2;`,
				Results: []sql.Row{{`test_inh_check`, `bmerged`, 0, true, false}, {`test_inh_check`, `bnoinherit`, 0, true, true}, {`test_inh_check`, `test_inh_check_a_check`, 0, true, false}, {`test_inh_check_child`, `blocal`, 0, true, false}, {`test_inh_check_child`, `bmerged`, 1, true, false}, {`test_inh_check_child`, `test_inh_check_a_check`, 1, false, false}},
			},
			{
				Statement: `ALTER TABLE test_inh_check ALTER COLUMN b TYPE numeric;`,
			},
			{
				Statement: `\d test_inh_check
           Table "public.test_inh_check"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | numeric |           |          | 
 b      | numeric |           |          | 
Check constraints:
    "bmerged" CHECK (b::double precision > 1::double precision)
    "bnoinherit" CHECK (b::double precision > 100::double precision) NO INHERIT
    "test_inh_check_a_check" CHECK (a::double precision > 10.2::double precision)
Number of child tables: 1 (Use \d+ to list them.)
\d test_inh_check_child
        Table "public.test_inh_check_child"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | numeric |           |          | 
 b      | numeric |           |          | 
Check constraints:
    "blocal" CHECK (b::double precision < 1000::double precision)
    "bmerged" CHECK (b::double precision > 1::double precision)
    "test_inh_check_a_check" CHECK (a::double precision > 10.2::double precision)
Inherits: test_inh_check
select relname, conname, coninhcount, conislocal, connoinherit
  from pg_constraint c, pg_class r
  where relname like 'test_inh_check%' and c.conrelid = r.oid
  order by 1, 2;`,
				Results: []sql.Row{{`test_inh_check`, `bmerged`, 0, true, false}, {`test_inh_check`, `bnoinherit`, 0, true, true}, {`test_inh_check`, `test_inh_check_a_check`, 0, true, false}, {`test_inh_check_child`, `blocal`, 0, true, false}, {`test_inh_check_child`, `bmerged`, 1, true, false}, {`test_inh_check_child`, `test_inh_check_a_check`, 1, false, false}},
			},
			{
				Statement: `CREATE TABLE test_type_diff (f1 int);`,
			},
			{
				Statement: `CREATE TABLE test_type_diff_c (extra smallint) INHERITS (test_type_diff);`,
			},
			{
				Statement: `ALTER TABLE test_type_diff ADD COLUMN f2 int;`,
			},
			{
				Statement: `INSERT INTO test_type_diff_c VALUES (1, 2, 3);`,
			},
			{
				Statement: `ALTER TABLE test_type_diff ALTER COLUMN f2 TYPE bigint USING f2::bigint;`,
			},
			{
				Statement: `CREATE TABLE test_type_diff2 (int_two int2, int_four int4, int_eight int8);`,
			},
			{
				Statement: `CREATE TABLE test_type_diff2_c1 (int_four int4, int_eight int8, int_two int2);`,
			},
			{
				Statement: `CREATE TABLE test_type_diff2_c2 (int_eight int8, int_two int2, int_four int4);`,
			},
			{
				Statement: `CREATE TABLE test_type_diff2_c3 (int_two int2, int_four int4, int_eight int8);`,
			},
			{
				Statement: `ALTER TABLE test_type_diff2_c1 INHERIT test_type_diff2;`,
			},
			{
				Statement: `ALTER TABLE test_type_diff2_c2 INHERIT test_type_diff2;`,
			},
			{
				Statement: `ALTER TABLE test_type_diff2_c3 INHERIT test_type_diff2;`,
			},
			{
				Statement: `INSERT INTO test_type_diff2_c1 VALUES (1, 2, 3);`,
			},
			{
				Statement: `INSERT INTO test_type_diff2_c2 VALUES (4, 5, 6);`,
			},
			{
				Statement: `INSERT INTO test_type_diff2_c3 VALUES (7, 8, 9);`,
			},
			{
				Statement: `ALTER TABLE test_type_diff2 ALTER COLUMN int_four TYPE int8 USING int_four::int8;`,
			},
			{
				Statement:   `ALTER TABLE test_type_diff2 ALTER COLUMN int_four TYPE int4 USING (pg_column_size(test_type_diff2));`,
				ErrorString: `cannot convert whole-row table reference`,
			},
			{
				Statement: `CREATE TABLE check_fk_presence_1 (id int PRIMARY KEY, t text);`,
			},
			{
				Statement: `CREATE TABLE check_fk_presence_2 (id int REFERENCES check_fk_presence_1, t text);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER TABLE check_fk_presence_2 DROP CONSTRAINT check_fk_presence_2_id_fkey;`,
			},
			{
				Statement: `ANALYZE check_fk_presence_2;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `\d check_fk_presence_2
        Table "public.check_fk_presence_2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           |          | 
 t      | text    |           |          | 
Foreign-key constraints:
    "check_fk_presence_2_id_fkey" FOREIGN KEY (id) REFERENCES check_fk_presence_1(id)
DROP TABLE check_fk_presence_1, check_fk_presence_2;`,
			},
			{
				Statement: `create table at_base_table(id int, stuff text);`,
			},
			{
				Statement: `insert into at_base_table values (23, 'skidoo');`,
			},
			{
				Statement: `create view at_view_1 as select * from at_base_table bt;`,
			},
			{
				Statement: `create view at_view_2 as select *, to_json(v1) as j from at_view_1 v1;`,
			},
			{
				Statement: `\d+ at_view_1
                          View "public.at_view_1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Description 
--------+---------+-----------+----------+---------+----------+-------------
 id     | integer |           |          |         | plain    | 
 stuff  | text    |           |          |         | extended | 
View definition:
 SELECT bt.id,
    bt.stuff
   FROM at_base_table bt;`,
			},
			{
				Statement: `\d+ at_view_2
                          View "public.at_view_2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Description 
--------+---------+-----------+----------+---------+----------+-------------
 id     | integer |           |          |         | plain    | 
 stuff  | text    |           |          |         | extended | 
 j      | json    |           |          |         | extended | 
View definition:
 SELECT v1.id,
    v1.stuff,
    to_json(v1.*) AS j
   FROM at_view_1 v1;`,
			},
			{
				Statement: `explain (verbose, costs off) select * from at_view_2;`,
				Results:   []sql.Row{{`Seq Scan on public.at_base_table bt`}, {`Output: bt.id, bt.stuff, to_json(ROW(bt.id, bt.stuff))`}},
			},
			{
				Statement: `select * from at_view_2;`,
				Results:   []sql.Row{{23, `skidoo`, `{"id":23,"stuff":"skidoo"}`}},
			},
			{
				Statement: `create or replace view at_view_1 as select *, 2+2 as more from at_base_table bt;`,
			},
			{
				Statement: `\d+ at_view_1
                          View "public.at_view_1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Description 
--------+---------+-----------+----------+---------+----------+-------------
 id     | integer |           |          |         | plain    | 
 stuff  | text    |           |          |         | extended | 
 more   | integer |           |          |         | plain    | 
View definition:
 SELECT bt.id,
    bt.stuff,
    2 + 2 AS more
   FROM at_base_table bt;`,
			},
			{
				Statement: `\d+ at_view_2
                          View "public.at_view_2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Description 
--------+---------+-----------+----------+---------+----------+-------------
 id     | integer |           |          |         | plain    | 
 stuff  | text    |           |          |         | extended | 
 j      | json    |           |          |         | extended | 
View definition:
 SELECT v1.id,
    v1.stuff,
    to_json(v1.*) AS j
   FROM at_view_1 v1;`,
			},
			{
				Statement: `explain (verbose, costs off) select * from at_view_2;`,
				Results:   []sql.Row{{`Seq Scan on public.at_base_table bt`}, {`Output: bt.id, bt.stuff, to_json(ROW(bt.id, bt.stuff, 4))`}},
			},
			{
				Statement: `select * from at_view_2;`,
				Results:   []sql.Row{{23, `skidoo`, `{"id":23,"stuff":"skidoo","more":4}`}},
			},
			{
				Statement: `drop view at_view_2;`,
			},
			{
				Statement: `drop view at_view_1;`,
			},
			{
				Statement: `drop table at_base_table;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create temp table t1 as select * from int8_tbl;`,
			},
			{
				Statement: `create temp view v1 as select 1::int8 as q1;`,
			},
			{
				Statement: `create temp view v2 as select * from v1;`,
			},
			{
				Statement: `create or replace temp view v1 with (security_barrier = true)
  as select * from t1;`,
			},
			{
				Statement: `create temp table log (q1 int8, q2 int8);`,
			},
			{
				Statement: `create rule v1_upd_rule as on update to v1
  do also insert into log values (new.*);`,
			},
			{
				Statement: `update v2 set q1 = q1 + 1 where q1 = 123;`,
			},
			{
				Statement: `select * from t1;`,
				Results:   []sql.Row{{4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}, {124, 456}, {124, 4567890123456789}},
			},
			{
				Statement: `select * from log;`,
				Results:   []sql.Row{{124, 456}, {124, 4567890123456789}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `CREATE FUNCTION check_ddl_rewrite(p_tablename regclass, p_ddl text)
RETURNS boolean
LANGUAGE plpgsql AS $$
DECLARE
    v_relfilenode oid;`,
			},
			{
				Statement: `BEGIN
    v_relfilenode := relfilenode FROM pg_class WHERE oid = p_tablename;`,
			},
			{
				Statement: `    EXECUTE p_ddl;`,
			},
			{
				Statement: `    RETURN v_relfilenode <> (SELECT relfilenode FROM pg_class WHERE oid = p_tablename);`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TABLE rewrite_test(col text);`,
			},
			{
				Statement: `INSERT INTO rewrite_test VALUES ('something');`,
			},
			{
				Statement: `INSERT INTO rewrite_test VALUES (NULL);`,
			},
			{
				Statement: `SELECT check_ddl_rewrite('rewrite_test', $$
  ALTER TABLE rewrite_test
      ADD COLUMN empty1 text,
      ADD COLUMN notempty1_rewrite serial;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT check_ddl_rewrite('rewrite_test', $$
    ALTER TABLE rewrite_test
        ADD COLUMN notempty2_rewrite serial,
        ADD COLUMN empty2 text;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT check_ddl_rewrite('rewrite_test', $$
    ALTER TABLE rewrite_test
        ADD COLUMN empty3 text,
        ADD COLUMN notempty3_norewrite int default 42;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT check_ddl_rewrite('rewrite_test', $$
    ALTER TABLE rewrite_test
        ADD COLUMN notempty4_norewrite int default 42,
        ADD COLUMN empty4 text;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT check_ddl_rewrite('rewrite_test', $$
    ALTER TABLE rewrite_test
        ADD COLUMN empty5 text,
        ADD COLUMN notempty5_norewrite int default 42,
        ADD COLUMN notempty5_rewrite serial;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT check_ddl_rewrite('rewrite_test', $$
    ALTER TABLE rewrite_test
        ADD COLUMN notempty6_rewrite serial,
        ADD COLUMN empty6 text,
        ADD COLUMN notempty6_norewrite int default 42;`,
			},
			{
				Statement: `$$);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP FUNCTION check_ddl_rewrite(regclass, text);`,
			},
			{
				Statement: `DROP TABLE rewrite_test;`,
			},
			{
				Statement:   `drop type lockmodes;`,
				ErrorString: `type "lockmodes" does not exist`,
			},
			{
				Statement: `create type lockmodes as enum (
 'SIReadLock'
,'AccessShareLock'
,'RowShareLock'
,'RowExclusiveLock'
,'ShareUpdateExclusiveLock'
,'ShareLock'
,'ShareRowExclusiveLock'
,'ExclusiveLock'
,'AccessExclusiveLock'
);`,
			},
			{
				Statement:   `drop view my_locks;`,
				ErrorString: `view "my_locks" does not exist`,
			},
			{
				Statement: `create or replace view my_locks as
select case when c.relname like 'pg_toast%' then 'pg_toast' else c.relname end, max(mode::lockmodes) as max_lockmode
from pg_locks l join pg_class c on l.relation = c.oid
where virtualtransaction = (
        select virtualtransaction
        from pg_locks
        where transactionid = pg_current_xact_id()::xid)
and locktype = 'relation'
and relnamespace != (select oid from pg_namespace where nspname = 'pg_catalog')
and c.relname != 'my_locks'
group by c.relname;`,
			},
			{
				Statement: `create table alterlock (f1 int primary key, f2 text);`,
			},
			{
				Statement: `insert into alterlock values (1, 'foo');`,
			},
			{
				Statement: `create table alterlock2 (f3 int primary key, f1 int);`,
			},
			{
				Statement: `insert into alterlock2 values (1, 1);`,
			},
			{
				Statement: `begin; alter table alterlock alter column f2 set statistics 150;`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `begin; alter table alterlock cluster on alterlock_pkey;`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}, {`alterlock_pkey`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin; alter table alterlock set without cluster;`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin; alter table alterlock set (fillfactor = 100);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}, {`pg_toast`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin; alter table alterlock reset (fillfactor);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}, {`pg_toast`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin; alter table alterlock set (toast.autovacuum_enabled = off);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}, {`pg_toast`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin; alter table alterlock set (autovacuum_enabled = off);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}, {`pg_toast`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin; alter table alterlock alter column f2 set (n_distinct = 1);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `begin; alter table alterlock set (autovacuum_enabled = off, fillfactor = 80);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareUpdateExclusiveLock`}, {`pg_toast`, `ShareUpdateExclusiveLock`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin; alter table alterlock alter column f2 set storage extended;`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `AccessExclusiveLock`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `begin; alter table alterlock alter column f2 set default 'x';`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `AccessExclusiveLock`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create trigger ttdummy
	before delete or update on alterlock
	for each row
	execute procedure
	ttdummy (1, 1);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareRowExclusiveLock`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `alter table alterlock2 add foreign key (f1) references alterlock (f1);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareRowExclusiveLock`}, {`alterlock2`, `ShareRowExclusiveLock`}, {`alterlock2_pkey`, `AccessShareLock`}, {`alterlock_pkey`, `AccessShareLock`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `alter table alterlock2
add constraint alterlock2nv foreign key (f1) references alterlock (f1) NOT VALID;`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `ShareRowExclusiveLock`}, {`alterlock2`, `ShareRowExclusiveLock`}},
			},
			{
				Statement: `commit;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `alter table alterlock2 validate constraint alterlock2nv;`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`alterlock`, `RowShareLock`}, {`alterlock2`, `ShareUpdateExclusiveLock`}, {`alterlock2_pkey`, `AccessShareLock`}, {`alterlock_pkey`, `AccessShareLock`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `create or replace view my_locks as
select case when c.relname like 'pg_toast%' then 'pg_toast' else c.relname end, max(mode::lockmodes) as max_lockmode
from pg_locks l join pg_class c on l.relation = c.oid
where virtualtransaction = (
        select virtualtransaction
        from pg_locks
        where transactionid = pg_current_xact_id()::xid)
and locktype = 'relation'
and relnamespace != (select oid from pg_namespace where nspname = 'pg_catalog')
and c.relname = 'my_locks'
group by c.relname;`,
			},
			{
				Statement:   `alter table my_locks set (autovacuum_enabled = false);`,
				ErrorString: `unrecognized parameter "autovacuum_enabled"`,
			},
			{
				Statement:   `alter view my_locks set (autovacuum_enabled = false);`,
				ErrorString: `unrecognized parameter "autovacuum_enabled"`,
			},
			{
				Statement: `alter table my_locks reset (autovacuum_enabled);`,
			},
			{
				Statement: `alter view my_locks reset (autovacuum_enabled);`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `alter view my_locks set (security_barrier=off);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`my_locks`, `AccessExclusiveLock`}},
			},
			{
				Statement: `alter view my_locks reset (security_barrier);`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `alter table my_locks set (security_barrier=off);`,
			},
			{
				Statement: `select * from my_locks order by 1;`,
				Results:   []sql.Row{{`my_locks`, `AccessExclusiveLock`}},
			},
			{
				Statement: `alter table my_locks reset (security_barrier);`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `drop table alterlock2;`,
			},
			{
				Statement: `drop table alterlock;`,
			},
			{
				Statement: `drop view my_locks;`,
			},
			{
				Statement: `drop type lockmodes;`,
			},
			{
				Statement: `create function test_strict(text) returns text as
    'select coalesce($1, ''got passed a null'');'
    language sql returns null on null input;`,
			},
			{
				Statement: `select test_strict(NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `alter function test_strict(text) called on null input;`,
			},
			{
				Statement: `select test_strict(NULL);`,
				Results:   []sql.Row{{`got passed a null`}},
			},
			{
				Statement: `create function non_strict(text) returns text as
    'select coalesce($1, ''got passed a null'');'
    language sql called on null input;`,
			},
			{
				Statement: `select non_strict(NULL);`,
				Results:   []sql.Row{{`got passed a null`}},
			},
			{
				Statement: `alter function non_strict(text) returns null on null input;`,
			},
			{
				Statement: `select non_strict(NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create schema alter1;`,
			},
			{
				Statement: `create schema alter2;`,
			},
			{
				Statement: `create table alter1.t1(f1 serial primary key, f2 int check (f2 > 0));`,
			},
			{
				Statement: `create view alter1.v1 as select * from alter1.t1;`,
			},
			{
				Statement: `create function alter1.plus1(int) returns int as 'select $1+1' language sql;`,
			},
			{
				Statement: `create domain alter1.posint integer check (value > 0);`,
			},
			{
				Statement: `create type alter1.ctype as (f1 int, f2 text);`,
			},
			{
				Statement: `create function alter1.same(alter1.ctype, alter1.ctype) returns boolean language sql
as 'select $1.f1 is not distinct from $2.f1 and $1.f2 is not distinct from $2.f2';`,
			},
			{
				Statement: `create operator alter1.=(procedure = alter1.same, leftarg  = alter1.ctype, rightarg = alter1.ctype);`,
			},
			{
				Statement: `create operator class alter1.ctype_hash_ops default for type alter1.ctype using hash as
  operator 1 alter1.=(alter1.ctype, alter1.ctype);`,
			},
			{
				Statement: `create conversion alter1.latin1_to_utf8 for 'latin1' to 'utf8' from iso8859_1_to_utf8;`,
			},
			{
				Statement: `create text search parser alter1.prs(start = prsd_start, gettoken = prsd_nexttoken, end = prsd_end, lextypes = prsd_lextype);`,
			},
			{
				Statement: `create text search configuration alter1.cfg(parser = alter1.prs);`,
			},
			{
				Statement: `create text search template alter1.tmpl(init = dsimple_init, lexize = dsimple_lexize);`,
			},
			{
				Statement: `create text search dictionary alter1.dict(template = alter1.tmpl);`,
			},
			{
				Statement: `insert into alter1.t1(f2) values(11);`,
			},
			{
				Statement: `insert into alter1.t1(f2) values(12);`,
			},
			{
				Statement: `alter table alter1.t1 set schema alter1; -- no-op, same schema`,
			},
			{
				Statement: `alter table alter1.t1 set schema alter2;`,
			},
			{
				Statement: `alter table alter1.v1 set schema alter2;`,
			},
			{
				Statement: `alter function alter1.plus1(int) set schema alter2;`,
			},
			{
				Statement: `alter domain alter1.posint set schema alter2;`,
			},
			{
				Statement: `alter operator class alter1.ctype_hash_ops using hash set schema alter2;`,
			},
			{
				Statement: `alter operator family alter1.ctype_hash_ops using hash set schema alter2;`,
			},
			{
				Statement: `alter operator alter1.=(alter1.ctype, alter1.ctype) set schema alter2;`,
			},
			{
				Statement: `alter function alter1.same(alter1.ctype, alter1.ctype) set schema alter2;`,
			},
			{
				Statement: `alter type alter1.ctype set schema alter1; -- no-op, same schema`,
			},
			{
				Statement: `alter type alter1.ctype set schema alter2;`,
			},
			{
				Statement: `alter conversion alter1.latin1_to_utf8 set schema alter2;`,
			},
			{
				Statement: `alter text search parser alter1.prs set schema alter2;`,
			},
			{
				Statement: `alter text search configuration alter1.cfg set schema alter2;`,
			},
			{
				Statement: `alter text search template alter1.tmpl set schema alter2;`,
			},
			{
				Statement: `alter text search dictionary alter1.dict set schema alter2;`,
			},
			{
				Statement: `drop schema alter1;`,
			},
			{
				Statement: `insert into alter2.t1(f2) values(13);`,
			},
			{
				Statement: `insert into alter2.t1(f2) values(14);`,
			},
			{
				Statement: `select * from alter2.t1;`,
				Results:   []sql.Row{{1, 11}, {2, 12}, {3, 13}, {4, 14}},
			},
			{
				Statement: `select * from alter2.v1;`,
				Results:   []sql.Row{{1, 11}, {2, 12}, {3, 13}, {4, 14}},
			},
			{
				Statement: `select alter2.plus1(41);`,
				Results:   []sql.Row{{42}},
			},
			{
				Statement: `drop schema alter2 cascade;`,
			},
			{
				Statement: `CREATE TYPE test_type AS (a int);`,
			},
			{
				Statement: `\d test_type
         Composite type "public.test_type"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
ALTER TYPE nosuchtype ADD ATTRIBUTE b text; -- fails`,
				ErrorString: `relation "nosuchtype" does not exist`,
			},
			{
				Statement: `ALTER TYPE test_type ADD ATTRIBUTE b text;`,
			},
			{
				Statement: `\d test_type
         Composite type "public.test_type"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | text    |           |          | 
ALTER TYPE test_type ADD ATTRIBUTE b text; -- fails`,
				ErrorString: `column "b" of relation "test_type" already exists`,
			},
			{
				Statement: `ALTER TYPE test_type ALTER ATTRIBUTE b SET DATA TYPE varchar;`,
			},
			{
				Statement: `\d test_type
              Composite type "public.test_type"
 Column |       Type        | Collation | Nullable | Default 
--------+-------------------+-----------+----------+---------
 a      | integer           |           |          | 
 b      | character varying |           |          | 
ALTER TYPE test_type ALTER ATTRIBUTE b SET DATA TYPE integer;`,
			},
			{
				Statement: `\d test_type
         Composite type "public.test_type"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
ALTER TYPE test_type DROP ATTRIBUTE b;`,
			},
			{
				Statement: `\d test_type
         Composite type "public.test_type"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
ALTER TYPE test_type DROP ATTRIBUTE c; -- fails`,
				ErrorString: `column "c" of relation "test_type" does not exist`,
			},
			{
				Statement: `ALTER TYPE test_type DROP ATTRIBUTE IF EXISTS c;`,
			},
			{
				Statement: `ALTER TYPE test_type DROP ATTRIBUTE a, ADD ATTRIBUTE d boolean;`,
			},
			{
				Statement: `\d test_type
         Composite type "public.test_type"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 d      | boolean |           |          | 
ALTER TYPE test_type RENAME ATTRIBUTE a TO aa;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement: `ALTER TYPE test_type RENAME ATTRIBUTE d TO dd;`,
			},
			{
				Statement: `\d test_type
         Composite type "public.test_type"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 dd     | boolean |           |          | 
DROP TYPE test_type;`,
			},
			{
				Statement: `CREATE TYPE test_type1 AS (a int, b text);`,
			},
			{
				Statement: `CREATE TABLE test_tbl1 (x int, y test_type1);`,
			},
			{
				Statement:   `ALTER TYPE test_type1 ALTER ATTRIBUTE b TYPE varchar; -- fails`,
				ErrorString: `cannot alter type "test_type1" because column "test_tbl1.y" uses it`,
			},
			{
				Statement: `DROP TABLE test_tbl1;`,
			},
			{
				Statement: `CREATE TABLE test_tbl1 (x int, y text);`,
			},
			{
				Statement: `CREATE INDEX test_tbl1_idx ON test_tbl1((row(x,y)::test_type1));`,
			},
			{
				Statement:   `ALTER TYPE test_type1 ALTER ATTRIBUTE b TYPE varchar; -- fails`,
				ErrorString: `cannot alter type "test_type1" because column "test_tbl1_idx.row" uses it`,
			},
			{
				Statement: `DROP TABLE test_tbl1;`,
			},
			{
				Statement: `DROP TYPE test_type1;`,
			},
			{
				Statement: `CREATE TYPE test_type2 AS (a int, b text);`,
			},
			{
				Statement: `CREATE TABLE test_tbl2 OF test_type2;`,
			},
			{
				Statement: `CREATE TABLE test_tbl2_subclass () INHERITS (test_tbl2);`,
			},
			{
				Statement: `\d test_type2
        Composite type "public.test_type2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | text    |           |          | 
\d test_tbl2
             Table "public.test_tbl2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | text    |           |          | 
Number of child tables: 1 (Use \d+ to list them.)
Typed table of type: test_type2
ALTER TYPE test_type2 ADD ATTRIBUTE c text; -- fails`,
				ErrorString: `cannot alter type "test_type2" because it is the type of a typed table`,
			},
			{
				Statement: `ALTER TYPE test_type2 ADD ATTRIBUTE c text CASCADE;`,
			},
			{
				Statement: `\d test_type2
        Composite type "public.test_type2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | text    |           |          | 
 c      | text    |           |          | 
\d test_tbl2
             Table "public.test_tbl2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | text    |           |          | 
 c      | text    |           |          | 
Number of child tables: 1 (Use \d+ to list them.)
Typed table of type: test_type2
ALTER TYPE test_type2 ALTER ATTRIBUTE b TYPE varchar; -- fails`,
				ErrorString: `cannot alter type "test_type2" because it is the type of a typed table`,
			},
			{
				Statement: `ALTER TYPE test_type2 ALTER ATTRIBUTE b TYPE varchar CASCADE;`,
			},
			{
				Statement: `\d test_type2
             Composite type "public.test_type2"
 Column |       Type        | Collation | Nullable | Default 
--------+-------------------+-----------+----------+---------
 a      | integer           |           |          | 
 b      | character varying |           |          | 
 c      | text              |           |          | 
\d test_tbl2
                  Table "public.test_tbl2"
 Column |       Type        | Collation | Nullable | Default 
--------+-------------------+-----------+----------+---------
 a      | integer           |           |          | 
 b      | character varying |           |          | 
 c      | text              |           |          | 
Number of child tables: 1 (Use \d+ to list them.)
Typed table of type: test_type2
ALTER TYPE test_type2 DROP ATTRIBUTE b; -- fails`,
				ErrorString: `cannot alter type "test_type2" because it is the type of a typed table`,
			},
			{
				Statement: `ALTER TYPE test_type2 DROP ATTRIBUTE b CASCADE;`,
			},
			{
				Statement: `\d test_type2
        Composite type "public.test_type2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 c      | text    |           |          | 
\d test_tbl2
             Table "public.test_tbl2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 c      | text    |           |          | 
Number of child tables: 1 (Use \d+ to list them.)
Typed table of type: test_type2
ALTER TYPE test_type2 RENAME ATTRIBUTE a TO aa; -- fails`,
				ErrorString: `cannot alter type "test_type2" because it is the type of a typed table`,
			},
			{
				Statement: `ALTER TYPE test_type2 RENAME ATTRIBUTE a TO aa CASCADE;`,
			},
			{
				Statement: `\d test_type2
        Composite type "public.test_type2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 aa     | integer |           |          | 
 c      | text    |           |          | 
\d test_tbl2
             Table "public.test_tbl2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 aa     | integer |           |          | 
 c      | text    |           |          | 
Number of child tables: 1 (Use \d+ to list them.)
Typed table of type: test_type2
\d test_tbl2_subclass
         Table "public.test_tbl2_subclass"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 aa     | integer |           |          | 
 c      | text    |           |          | 
Inherits: test_tbl2
DROP TABLE test_tbl2_subclass, test_tbl2;`,
			},
			{
				Statement: `DROP TYPE test_type2;`,
			},
			{
				Statement: `CREATE TYPE test_typex AS (a int, b text);`,
			},
			{
				Statement: `CREATE TABLE test_tblx (x int, y test_typex check ((y).a > 0));`,
			},
			{
				Statement:   `ALTER TYPE test_typex DROP ATTRIBUTE a; -- fails`,
				ErrorString: `cannot drop column a of composite type test_typex because other objects depend on it`,
			},
			{
				Statement: `ALTER TYPE test_typex DROP ATTRIBUTE a CASCADE;`,
			},
			{
				Statement: `\d test_tblx
               Table "public.test_tblx"
 Column |    Type    | Collation | Nullable | Default 
--------+------------+-----------+----------+---------
 x      | integer    |           |          | 
 y      | test_typex |           |          | 
DROP TABLE test_tblx;`,
			},
			{
				Statement: `DROP TYPE test_typex;`,
			},
			{
				Statement: `CREATE TYPE test_type3 AS (a int);`,
			},
			{
				Statement: `CREATE TABLE test_tbl3 (c) AS SELECT '(1)'::test_type3;`,
			},
			{
				Statement: `ALTER TYPE test_type3 DROP ATTRIBUTE a, ADD ATTRIBUTE b int;`,
			},
			{
				Statement: `CREATE TYPE test_type_empty AS ();`,
			},
			{
				Statement: `DROP TYPE test_type_empty;`,
			},
			{
				Statement: `CREATE TYPE tt_t0 AS (z inet, x int, y numeric(8,2));`,
			},
			{
				Statement: `ALTER TYPE tt_t0 DROP ATTRIBUTE z;`,
			},
			{
				Statement: `CREATE TABLE tt0 (x int NOT NULL, y numeric(8,2));	-- OK`,
			},
			{
				Statement: `CREATE TABLE tt1 (x int, y bigint);					-- wrong base type`,
			},
			{
				Statement: `CREATE TABLE tt2 (x int, y numeric(9,2));			-- wrong typmod`,
			},
			{
				Statement: `CREATE TABLE tt3 (y numeric(8,2), x int);			-- wrong column order`,
			},
			{
				Statement: `CREATE TABLE tt4 (x int);							-- too few columns`,
			},
			{
				Statement: `CREATE TABLE tt5 (x int, y numeric(8,2), z int);	-- too few columns`,
			},
			{
				Statement: `CREATE TABLE tt6 () INHERITS (tt0);					-- can't have a parent`,
			},
			{
				Statement: `CREATE TABLE tt7 (x int, q text, y numeric(8,2));`,
			},
			{
				Statement: `ALTER TABLE tt7 DROP q;								-- OK`,
			},
			{
				Statement: `ALTER TABLE tt0 OF tt_t0;`,
			},
			{
				Statement:   `ALTER TABLE tt1 OF tt_t0;`,
				ErrorString: `table "tt1" has different type for column "y"`,
			},
			{
				Statement:   `ALTER TABLE tt2 OF tt_t0;`,
				ErrorString: `table "tt2" has different type for column "y"`,
			},
			{
				Statement:   `ALTER TABLE tt3 OF tt_t0;`,
				ErrorString: `table has column "y" where type requires "x"`,
			},
			{
				Statement:   `ALTER TABLE tt4 OF tt_t0;`,
				ErrorString: `table is missing column "y"`,
			},
			{
				Statement:   `ALTER TABLE tt5 OF tt_t0;`,
				ErrorString: `table has extra column "z"`,
			},
			{
				Statement:   `ALTER TABLE tt6 OF tt_t0;`,
				ErrorString: `typed tables cannot inherit`,
			},
			{
				Statement: `ALTER TABLE tt7 OF tt_t0;`,
			},
			{
				Statement: `CREATE TYPE tt_t1 AS (x int, y numeric(8,2));`,
			},
			{
				Statement: `ALTER TABLE tt7 OF tt_t1;			-- reassign an already-typed table`,
			},
			{
				Statement: `ALTER TABLE tt7 NOT OF;`,
			},
			{
				Statement: `\d tt7
                   Table "public.tt7"
 Column |     Type     | Collation | Nullable | Default 
--------+--------------+-----------+----------+---------
 x      | integer      |           |          | 
 y      | numeric(8,2) |           |          | 
CREATE TABLE test_drop_constr_parent (c text CHECK (c IS NOT NULL));`,
			},
			{
				Statement: `CREATE TABLE test_drop_constr_child () INHERITS (test_drop_constr_parent);`,
			},
			{
				Statement: `ALTER TABLE ONLY test_drop_constr_parent DROP CONSTRAINT "test_drop_constr_parent_c_check";`,
			},
			{
				Statement:   `INSERT INTO test_drop_constr_child (c) VALUES (NULL);`,
				ErrorString: `new row for relation "test_drop_constr_child" violates check constraint "test_drop_constr_parent_c_check"`,
			},
			{
				Statement: `DROP TABLE test_drop_constr_parent CASCADE;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 ADD COLUMN f int;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 ADD CONSTRAINT xxx PRIMARY KEY(f);`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 ADD CHECK (f BETWEEN 0 AND 10);`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 ALTER COLUMN f SET DEFAULT 0;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 RENAME COLUMN f TO f1;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 SET SCHEMA alter2;`,
			},
			{
				Statement: `CREATE TABLE tt8(a int);`,
			},
			{
				Statement: `CREATE SCHEMA alter2;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 ADD COLUMN f int;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 ADD CONSTRAINT xxx PRIMARY KEY(f);`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 ADD CHECK (f BETWEEN 0 AND 10);`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 ALTER COLUMN f SET DEFAULT 0;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 RENAME COLUMN f TO f1;`,
			},
			{
				Statement: `ALTER TABLE IF EXISTS tt8 SET SCHEMA alter2;`,
			},
			{
				Statement: `\d alter2.tt8
                Table "alter2.tt8"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 f1     | integer |           | not null | 0
Indexes:
    "xxx" PRIMARY KEY, btree (f1)
Check constraints:
    "tt8_f_check" CHECK (f1 >= 0 AND f1 <= 10)
DROP TABLE alter2.tt8;`,
			},
			{
				Statement: `DROP SCHEMA alter2;`,
			},
			{
				Statement: `CREATE TABLE tt9(c integer);`,
			},
			{
				Statement: `ALTER TABLE tt9 ADD CHECK(c > 1);`,
			},
			{
				Statement: `ALTER TABLE tt9 ADD CHECK(c > 2);  -- picks nonconflicting name`,
			},
			{
				Statement: `ALTER TABLE tt9 ADD CONSTRAINT foo CHECK(c > 3);`,
			},
			{
				Statement:   `ALTER TABLE tt9 ADD CONSTRAINT foo CHECK(c > 4);  -- fail, dup name`,
				ErrorString: `constraint "foo" for relation "tt9" already exists`,
			},
			{
				Statement: `ALTER TABLE tt9 ADD UNIQUE(c);`,
			},
			{
				Statement: `ALTER TABLE tt9 ADD UNIQUE(c);  -- picks nonconflicting name`,
			},
			{
				Statement:   `ALTER TABLE tt9 ADD CONSTRAINT tt9_c_key UNIQUE(c);  -- fail, dup name`,
				ErrorString: `relation "tt9_c_key" already exists`,
			},
			{
				Statement:   `ALTER TABLE tt9 ADD CONSTRAINT foo UNIQUE(c);  -- fail, dup name`,
				ErrorString: `constraint "foo" for relation "tt9" already exists`,
			},
			{
				Statement:   `ALTER TABLE tt9 ADD CONSTRAINT tt9_c_key CHECK(c > 5);  -- fail, dup name`,
				ErrorString: `constraint "tt9_c_key" for relation "tt9" already exists`,
			},
			{
				Statement: `ALTER TABLE tt9 ADD CONSTRAINT tt9_c_key2 CHECK(c > 6);`,
			},
			{
				Statement: `ALTER TABLE tt9 ADD UNIQUE(c);  -- picks nonconflicting name`,
			},
			{
				Statement: `\d tt9
                Table "public.tt9"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c      | integer |           |          | 
Indexes:
    "tt9_c_key" UNIQUE CONSTRAINT, btree (c)
    "tt9_c_key1" UNIQUE CONSTRAINT, btree (c)
    "tt9_c_key3" UNIQUE CONSTRAINT, btree (c)
Check constraints:
    "foo" CHECK (c > 3)
    "tt9_c_check" CHECK (c > 1)
    "tt9_c_check1" CHECK (c > 2)
    "tt9_c_key2" CHECK (c > 6)
DROP TABLE tt9;`,
			},
			{
				Statement: `CREATE TABLE comment_test (
  id int,
  positive_col int CHECK (positive_col > 0),
  indexed_col int,
  CONSTRAINT comment_test_pk PRIMARY KEY (id));`,
			},
			{
				Statement: `CREATE INDEX comment_test_index ON comment_test(indexed_col);`,
			},
			{
				Statement: `COMMENT ON COLUMN comment_test.id IS 'Column ''id'' on comment_test';`,
			},
			{
				Statement: `COMMENT ON INDEX comment_test_index IS 'Simple index on comment_test';`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT comment_test_positive_col_check ON comment_test IS 'CHECK constraint on comment_test.positive_col';`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT comment_test_pk ON comment_test IS 'PRIMARY KEY constraint of comment_test';`,
			},
			{
				Statement: `COMMENT ON INDEX comment_test_pk IS 'Index backing the PRIMARY KEY of comment_test';`,
			},
			{
				Statement: `SELECT col_description('comment_test'::regclass, 1) as comment;`,
				Results:   []sql.Row{{`Column 'id' on comment_test`}},
			},
			{
				Statement: `SELECT indexrelid::regclass::text as index, obj_description(indexrelid, 'pg_class') as comment FROM pg_index where indrelid = 'comment_test'::regclass ORDER BY 1, 2;`,
				Results:   []sql.Row{{`comment_test_index`, `Simple index on comment_test`}, {`comment_test_pk`, `Index backing the PRIMARY KEY of comment_test`}},
			},
			{
				Statement: `SELECT conname as constraint, obj_description(oid, 'pg_constraint') as comment FROM pg_constraint where conrelid = 'comment_test'::regclass ORDER BY 1, 2;`,
				Results:   []sql.Row{{`comment_test_pk`, `PRIMARY KEY constraint of comment_test`}, {`comment_test_positive_col_check`, `CHECK constraint on comment_test.positive_col`}},
			},
			{
				Statement: `ALTER TABLE comment_test ALTER COLUMN indexed_col SET DATA TYPE int;`,
			},
			{
				Statement: `ALTER TABLE comment_test ALTER COLUMN indexed_col SET DATA TYPE text;`,
			},
			{
				Statement: `ALTER TABLE comment_test ALTER COLUMN id SET DATA TYPE int;`,
			},
			{
				Statement: `ALTER TABLE comment_test ALTER COLUMN id SET DATA TYPE text;`,
			},
			{
				Statement: `ALTER TABLE comment_test ALTER COLUMN positive_col SET DATA TYPE int;`,
			},
			{
				Statement: `ALTER TABLE comment_test ALTER COLUMN positive_col SET DATA TYPE bigint;`,
			},
			{
				Statement: `SELECT col_description('comment_test'::regclass, 1) as comment;`,
				Results:   []sql.Row{{`Column 'id' on comment_test`}},
			},
			{
				Statement: `SELECT indexrelid::regclass::text as index, obj_description(indexrelid, 'pg_class') as comment FROM pg_index where indrelid = 'comment_test'::regclass ORDER BY 1, 2;`,
				Results:   []sql.Row{{`comment_test_index`, `Simple index on comment_test`}, {`comment_test_pk`, `Index backing the PRIMARY KEY of comment_test`}},
			},
			{
				Statement: `SELECT conname as constraint, obj_description(oid, 'pg_constraint') as comment FROM pg_constraint where conrelid = 'comment_test'::regclass ORDER BY 1, 2;`,
				Results:   []sql.Row{{`comment_test_pk`, `PRIMARY KEY constraint of comment_test`}, {`comment_test_positive_col_check`, `CHECK constraint on comment_test.positive_col`}},
			},
			{
				Statement: `CREATE TABLE comment_test_child (
  id text CONSTRAINT comment_test_child_fk REFERENCES comment_test);`,
			},
			{
				Statement: `CREATE INDEX comment_test_child_fk ON comment_test_child(id);`,
			},
			{
				Statement: `COMMENT ON COLUMN comment_test_child.id IS 'Column ''id'' on comment_test_child';`,
			},
			{
				Statement: `COMMENT ON INDEX comment_test_child_fk IS 'Index backing the FOREIGN KEY of comment_test_child';`,
			},
			{
				Statement: `COMMENT ON CONSTRAINT comment_test_child_fk ON comment_test_child IS 'FOREIGN KEY constraint of comment_test_child';`,
			},
			{
				Statement: `ALTER TABLE comment_test ALTER COLUMN id SET DATA TYPE text;`,
			},
			{
				Statement:   `ALTER TABLE comment_test ALTER COLUMN id SET DATA TYPE int USING id::integer;`,
				ErrorString: `foreign key constraint "comment_test_child_fk" cannot be implemented`,
			},
			{
				Statement: `SELECT col_description('comment_test_child'::regclass, 1) as comment;`,
				Results:   []sql.Row{{`Column 'id' on comment_test_child`}},
			},
			{
				Statement: `SELECT indexrelid::regclass::text as index, obj_description(indexrelid, 'pg_class') as comment FROM pg_index where indrelid = 'comment_test_child'::regclass ORDER BY 1, 2;`,
				Results:   []sql.Row{{`comment_test_child_fk`, `Index backing the FOREIGN KEY of comment_test_child`}},
			},
			{
				Statement: `SELECT conname as constraint, obj_description(oid, 'pg_constraint') as comment FROM pg_constraint where conrelid = 'comment_test_child'::regclass ORDER BY 1, 2;`,
				Results:   []sql.Row{{`comment_test_child_fk`, `FOREIGN KEY constraint of comment_test_child`}},
			},
			{
				Statement: `CREATE TEMP TABLE filenode_mapping AS
SELECT
    oid, mapped_oid, reltablespace, relfilenode, relname
FROM pg_class,
    pg_filenode_relation(reltablespace, pg_relation_filenode(oid)) AS mapped_oid
WHERE relkind IN ('r', 'i', 'S', 't', 'm') AND mapped_oid IS DISTINCT FROM oid;`,
			},
			{
				Statement: `SELECT m.* FROM filenode_mapping m LEFT JOIN pg_class c ON c.oid = m.oid
WHERE c.oid IS NOT NULL OR m.mapped_oid IS NOT NULL;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SHOW allow_system_table_mods;`,
				Results:   []sql.Row{{`off`}},
			},
			{
				Statement:   `CREATE TABLE pg_catalog.new_system_table();`,
				ErrorString: `permission denied to create "pg_catalog.new_system_table"`,
			},
			{
				Statement: `CREATE TABLE new_system_table(id serial primary key, othercol text);`,
			},
			{
				Statement: `ALTER TABLE new_system_table SET SCHEMA pg_catalog;`,
			},
			{
				Statement: `ALTER TABLE new_system_table SET SCHEMA public;`,
			},
			{
				Statement: `ALTER TABLE new_system_table SET SCHEMA pg_catalog;`,
			},
			{
				Statement: `ALTER TABLE new_system_table SET SCHEMA pg_catalog;`,
			},
			{
				Statement: `ALTER TABLE new_system_table RENAME TO old_system_table;`,
			},
			{
				Statement: `CREATE INDEX old_system_table__othercol ON old_system_table (othercol);`,
			},
			{
				Statement: `INSERT INTO old_system_table(othercol) VALUES ('somedata'), ('otherdata');`,
			},
			{
				Statement: `UPDATE old_system_table SET id = -id;`,
			},
			{
				Statement: `DELETE FROM old_system_table WHERE othercol = 'somedata';`,
			},
			{
				Statement: `TRUNCATE old_system_table;`,
			},
			{
				Statement: `ALTER TABLE old_system_table DROP CONSTRAINT new_system_table_pkey;`,
			},
			{
				Statement: `ALTER TABLE old_system_table DROP COLUMN othercol;`,
			},
			{
				Statement: `DROP TABLE old_system_table;`,
			},
			{
				Statement: `CREATE UNLOGGED TABLE unlogged1(f1 SERIAL PRIMARY KEY, f2 TEXT); -- has sequence, toast`,
			},
			{
				Statement: `SELECT relname, relkind, relpersistence FROM pg_class WHERE relname ~ '^unlogged1'
UNION ALL
SELECT r.relname || ' toast table', t.relkind, t.relpersistence FROM pg_class r JOIN pg_class t ON t.oid = r.reltoastrelid WHERE r.relname ~ '^unlogged1'
UNION ALL
SELECT r.relname || ' toast index', ri.relkind, ri.relpersistence FROM pg_class r join pg_class t ON t.oid = r.reltoastrelid JOIN pg_index i ON i.indrelid = t.oid JOIN pg_class ri ON ri.oid = i.indexrelid WHERE r.relname ~ '^unlogged1'
ORDER BY relname;`,
				Results: []sql.Row{{`unlogged1`, `r`, `u`}, {`unlogged1 toast index`, `i`, `u`}, {`unlogged1 toast table`, true, `u`}, {`unlogged1_f1_seq`, `S`, `u`}, {`unlogged1_pkey`, `i`, `u`}},
			},
			{
				Statement: `CREATE UNLOGGED TABLE unlogged2(f1 SERIAL PRIMARY KEY, f2 INTEGER REFERENCES unlogged1); -- foreign key`,
			},
			{
				Statement: `CREATE UNLOGGED TABLE unlogged3(f1 SERIAL PRIMARY KEY, f2 INTEGER REFERENCES unlogged3); -- self-referencing foreign key`,
			},
			{
				Statement: `ALTER TABLE unlogged3 SET LOGGED; -- skip self-referencing foreign key`,
			},
			{
				Statement:   `ALTER TABLE unlogged2 SET LOGGED; -- fails because a foreign key to an unlogged table exists`,
				ErrorString: `could not change table "unlogged2" to logged because it references unlogged table "unlogged1"`,
			},
			{
				Statement: `ALTER TABLE unlogged1 SET LOGGED;`,
			},
			{
				Statement: `SELECT relname, relkind, relpersistence FROM pg_class WHERE relname ~ '^unlogged1'
UNION ALL
SELECT r.relname || ' toast table', t.relkind, t.relpersistence FROM pg_class r JOIN pg_class t ON t.oid = r.reltoastrelid WHERE r.relname ~ '^unlogged1'
UNION ALL
SELECT r.relname || ' toast index', ri.relkind, ri.relpersistence FROM pg_class r join pg_class t ON t.oid = r.reltoastrelid JOIN pg_index i ON i.indrelid = t.oid JOIN pg_class ri ON ri.oid = i.indexrelid WHERE r.relname ~ '^unlogged1'
ORDER BY relname;`,
				Results: []sql.Row{{`unlogged1`, `r`, `p`}, {`unlogged1 toast index`, `i`, `p`}, {`unlogged1 toast table`, true, `p`}, {`unlogged1_f1_seq`, `S`, `p`}, {`unlogged1_pkey`, `i`, `p`}},
			},
			{
				Statement: `ALTER TABLE unlogged1 SET LOGGED; -- silently do nothing`,
			},
			{
				Statement: `DROP TABLE unlogged3;`,
			},
			{
				Statement: `DROP TABLE unlogged2;`,
			},
			{
				Statement: `DROP TABLE unlogged1;`,
			},
			{
				Statement: `CREATE TABLE logged1(f1 SERIAL PRIMARY KEY, f2 TEXT); -- has sequence, toast`,
			},
			{
				Statement: `SELECT relname, relkind, relpersistence FROM pg_class WHERE relname ~ '^logged1'
UNION ALL
SELECT r.relname || ' toast table', t.relkind, t.relpersistence FROM pg_class r JOIN pg_class t ON t.oid = r.reltoastrelid WHERE r.relname ~ '^logged1'
UNION ALL
SELECT r.relname ||' toast index', ri.relkind, ri.relpersistence FROM pg_class r join pg_class t ON t.oid = r.reltoastrelid JOIN pg_index i ON i.indrelid = t.oid JOIN pg_class ri ON ri.oid = i.indexrelid WHERE r.relname ~ '^logged1'
ORDER BY relname;`,
				Results: []sql.Row{{`logged1`, `r`, `p`}, {`logged1 toast index`, `i`, `p`}, {`logged1 toast table`, true, `p`}, {`logged1_f1_seq`, `S`, `p`}, {`logged1_pkey`, `i`, `p`}},
			},
			{
				Statement: `CREATE TABLE logged2(f1 SERIAL PRIMARY KEY, f2 INTEGER REFERENCES logged1); -- foreign key`,
			},
			{
				Statement: `CREATE TABLE logged3(f1 SERIAL PRIMARY KEY, f2 INTEGER REFERENCES logged3); -- self-referencing foreign key`,
			},
			{
				Statement:   `ALTER TABLE logged1 SET UNLOGGED; -- fails because a foreign key from a permanent table exists`,
				ErrorString: `could not change table "logged1" to unlogged because it references logged table "logged2"`,
			},
			{
				Statement: `ALTER TABLE logged3 SET UNLOGGED; -- skip self-referencing foreign key`,
			},
			{
				Statement: `ALTER TABLE logged2 SET UNLOGGED;`,
			},
			{
				Statement: `ALTER TABLE logged1 SET UNLOGGED;`,
			},
			{
				Statement: `SELECT relname, relkind, relpersistence FROM pg_class WHERE relname ~ '^logged1'
UNION ALL
SELECT r.relname || ' toast table', t.relkind, t.relpersistence FROM pg_class r JOIN pg_class t ON t.oid = r.reltoastrelid WHERE r.relname ~ '^logged1'
UNION ALL
SELECT r.relname || ' toast index', ri.relkind, ri.relpersistence FROM pg_class r join pg_class t ON t.oid = r.reltoastrelid JOIN pg_index i ON i.indrelid = t.oid JOIN pg_class ri ON ri.oid = i.indexrelid WHERE r.relname ~ '^logged1'
ORDER BY relname;`,
				Results: []sql.Row{{`logged1`, `r`, `u`}, {`logged1 toast index`, `i`, `u`}, {`logged1 toast table`, true, `u`}, {`logged1_f1_seq`, `S`, `u`}, {`logged1_pkey`, `i`, `u`}},
			},
			{
				Statement: `ALTER TABLE logged1 SET UNLOGGED; -- silently do nothing`,
			},
			{
				Statement: `DROP TABLE logged3;`,
			},
			{
				Statement: `DROP TABLE logged2;`,
			},
			{
				Statement: `DROP TABLE logged1;`,
			},
			{
				Statement: `CREATE TABLE test_add_column(c1 integer);`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
ALTER TABLE test_add_column
	ADD COLUMN c2 integer;`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
ALTER TABLE test_add_column
	ADD COLUMN c2 integer; -- fail because c2 already exists`,
				ErrorString: `column "c2" of relation "test_add_column" already exists`,
			},
			{
				Statement: `ALTER TABLE ONLY test_add_column
	ADD COLUMN c2 integer; -- fail because c2 already exists`,
				ErrorString: `column "c2" of relation "test_add_column" already exists`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
ALTER TABLE test_add_column
	ADD COLUMN IF NOT EXISTS c2 integer; -- skipping because c2 already exists`,
			},
			{
				Statement: `ALTER TABLE ONLY test_add_column
	ADD COLUMN IF NOT EXISTS c2 integer; -- skipping because c2 already exists`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
ALTER TABLE test_add_column
	ADD COLUMN c2 integer, -- fail because c2 already exists
	ADD COLUMN c3 integer primary key;`,
				ErrorString: `column "c2" of relation "test_add_column" already exists`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
ALTER TABLE test_add_column
	ADD COLUMN IF NOT EXISTS c2 integer, -- skipping because c2 already exists
	ADD COLUMN c3 integer primary key;`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
 c3     | integer |           | not null | 
Indexes:
    "test_add_column_pkey" PRIMARY KEY, btree (c3)
ALTER TABLE test_add_column
	ADD COLUMN IF NOT EXISTS c2 integer, -- skipping because c2 already exists
	ADD COLUMN IF NOT EXISTS c3 integer primary key; -- skipping because c3 already exists`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
 c3     | integer |           | not null | 
Indexes:
    "test_add_column_pkey" PRIMARY KEY, btree (c3)
ALTER TABLE test_add_column
	ADD COLUMN IF NOT EXISTS c2 integer, -- skipping because c2 already exists
	ADD COLUMN IF NOT EXISTS c3 integer, -- skipping because c3 already exists
	ADD COLUMN c4 integer REFERENCES test_add_column;`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
 c3     | integer |           | not null | 
 c4     | integer |           |          | 
Indexes:
    "test_add_column_pkey" PRIMARY KEY, btree (c3)
Foreign-key constraints:
    "test_add_column_c4_fkey" FOREIGN KEY (c4) REFERENCES test_add_column(c3)
Referenced by:
    TABLE "test_add_column" CONSTRAINT "test_add_column_c4_fkey" FOREIGN KEY (c4) REFERENCES test_add_column(c3)
ALTER TABLE test_add_column
	ADD COLUMN IF NOT EXISTS c4 integer REFERENCES test_add_column;`,
			},
			{
				Statement: `\d test_add_column
          Table "public.test_add_column"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
 c3     | integer |           | not null | 
 c4     | integer |           |          | 
Indexes:
    "test_add_column_pkey" PRIMARY KEY, btree (c3)
Foreign-key constraints:
    "test_add_column_c4_fkey" FOREIGN KEY (c4) REFERENCES test_add_column(c3)
Referenced by:
    TABLE "test_add_column" CONSTRAINT "test_add_column_c4_fkey" FOREIGN KEY (c4) REFERENCES test_add_column(c3)
ALTER TABLE test_add_column
	ADD COLUMN IF NOT EXISTS c5 SERIAL CHECK (c5 > 8);`,
			},
			{
				Statement: `\d test_add_column
                            Table "public.test_add_column"
 Column |  Type   | Collation | Nullable |                   Default                   
--------+---------+-----------+----------+---------------------------------------------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
 c3     | integer |           | not null | 
 c4     | integer |           |          | 
 c5     | integer |           | not null | nextval('test_add_column_c5_seq'::regclass)
Indexes:
    "test_add_column_pkey" PRIMARY KEY, btree (c3)
Check constraints:
    "test_add_column_c5_check" CHECK (c5 > 8)
Foreign-key constraints:
    "test_add_column_c4_fkey" FOREIGN KEY (c4) REFERENCES test_add_column(c3)
Referenced by:
    TABLE "test_add_column" CONSTRAINT "test_add_column_c4_fkey" FOREIGN KEY (c4) REFERENCES test_add_column(c3)
ALTER TABLE test_add_column
	ADD COLUMN IF NOT EXISTS c5 SERIAL CHECK (c5 > 10);`,
			},
			{
				Statement: `\d test_add_column*
                            Table "public.test_add_column"
 Column |  Type   | Collation | Nullable |                   Default                   
--------+---------+-----------+----------+---------------------------------------------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
 c3     | integer |           | not null | 
 c4     | integer |           |          | 
 c5     | integer |           | not null | nextval('test_add_column_c5_seq'::regclass)
Indexes:
    "test_add_column_pkey" PRIMARY KEY, btree (c3)
Check constraints:
    "test_add_column_c5_check" CHECK (c5 > 8)
Foreign-key constraints:
    "test_add_column_c4_fkey" FOREIGN KEY (c4) REFERENCES test_add_column(c3)
Referenced by:
    TABLE "test_add_column" CONSTRAINT "test_add_column_c4_fkey" FOREIGN KEY (c4) REFERENCES test_add_column(c3)
               Sequence "public.test_add_column_c5_seq"
  Type   | Start | Minimum |  Maximum   | Increment | Cycles? | Cache 
---------+-------+---------+------------+-----------+---------+-------
 integer |     1 |       1 | 2147483647 |         1 | no      |     1
Owned by: public.test_add_column.c5
 Index "public.test_add_column_pkey"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 c3     | integer | yes  | c3
primary key, btree, for table "public.test_add_column"
DROP TABLE test_add_column;`,
			},
			{
				Statement: `\d test_add_column*
CREATE TABLE ataddindex(f1 INT);`,
			},
			{
				Statement: `INSERT INTO ataddindex VALUES (42), (43);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX ataddindexi0 ON ataddindex(f1);`,
			},
			{
				Statement: `ALTER TABLE ataddindex
  ADD PRIMARY KEY USING INDEX ataddindexi0,
  ALTER f1 TYPE BIGINT;`,
			},
			{
				Statement: `\d ataddindex
            Table "public.ataddindex"
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 f1     | bigint |           | not null | 
Indexes:
    "ataddindexi0" PRIMARY KEY, btree (f1)
DROP TABLE ataddindex;`,
			},
			{
				Statement: `CREATE TABLE ataddindex(f1 VARCHAR(10));`,
			},
			{
				Statement: `INSERT INTO ataddindex(f1) VALUES ('foo'), ('a');`,
			},
			{
				Statement: `ALTER TABLE ataddindex
  ALTER f1 SET DATA TYPE TEXT,
  ADD EXCLUDE ((f1 LIKE 'a') WITH =);`,
			},
			{
				Statement: `\d ataddindex
           Table "public.ataddindex"
 Column | Type | Collation | Nullable | Default 
--------+------+-----------+----------+---------
 f1     | text |           |          | 
Indexes:
    "ataddindex_expr_excl" EXCLUDE USING btree ((f1 ~~ 'a'::text) WITH =)
DROP TABLE ataddindex;`,
			},
			{
				Statement: `CREATE TABLE ataddindex(id int, ref_id int);`,
			},
			{
				Statement: `ALTER TABLE ataddindex
  ADD PRIMARY KEY (id),
  ADD FOREIGN KEY (ref_id) REFERENCES ataddindex;`,
			},
			{
				Statement: `\d ataddindex
             Table "public.ataddindex"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           | not null | 
 ref_id | integer |           |          | 
Indexes:
    "ataddindex_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "ataddindex_ref_id_fkey" FOREIGN KEY (ref_id) REFERENCES ataddindex(id)
Referenced by:
    TABLE "ataddindex" CONSTRAINT "ataddindex_ref_id_fkey" FOREIGN KEY (ref_id) REFERENCES ataddindex(id)
DROP TABLE ataddindex;`,
			},
			{
				Statement: `CREATE TABLE ataddindex(id int, ref_id int);`,
			},
			{
				Statement: `ALTER TABLE ataddindex
  ADD UNIQUE (id),
  ADD FOREIGN KEY (ref_id) REFERENCES ataddindex (id);`,
			},
			{
				Statement: `\d ataddindex
             Table "public.ataddindex"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           |          | 
 ref_id | integer |           |          | 
Indexes:
    "ataddindex_id_key" UNIQUE CONSTRAINT, btree (id)
Foreign-key constraints:
    "ataddindex_ref_id_fkey" FOREIGN KEY (ref_id) REFERENCES ataddindex(id)
Referenced by:
    TABLE "ataddindex" CONSTRAINT "ataddindex_ref_id_fkey" FOREIGN KEY (ref_id) REFERENCES ataddindex(id)
DROP TABLE ataddindex;`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int,
	b int
) PARTITION BY RANGE (a, (a+b+1));`,
			},
			{
				Statement:   `ALTER TABLE partitioned ADD EXCLUDE USING gist (a WITH &&);`,
				ErrorString: `exclusion constraints are not supported on partitioned tables`,
			},
			{
				Statement:   `ALTER TABLE partitioned DROP COLUMN a;`,
				ErrorString: `cannot drop column "a" because it is part of the partition key of relation "partitioned"`,
			},
			{
				Statement:   `ALTER TABLE partitioned ALTER COLUMN a TYPE char(5);`,
				ErrorString: `cannot alter column "a" because it is part of the partition key of relation "partitioned"`,
			},
			{
				Statement:   `ALTER TABLE partitioned DROP COLUMN b;`,
				ErrorString: `cannot drop column "b" because it is part of the partition key of relation "partitioned"`,
			},
			{
				Statement:   `ALTER TABLE partitioned ALTER COLUMN b TYPE char(5);`,
				ErrorString: `cannot alter column "b" because it is part of the partition key of relation "partitioned"`,
			},
			{
				Statement: `CREATE TABLE nonpartitioned (
	a int,
	b int
);`,
			},
			{
				Statement:   `ALTER TABLE partitioned INHERIT nonpartitioned;`,
				ErrorString: `cannot change inheritance of partitioned table`,
			},
			{
				Statement:   `ALTER TABLE nonpartitioned INHERIT partitioned;`,
				ErrorString: `cannot inherit from partitioned table "partitioned"`,
			},
			{
				Statement:   `ALTER TABLE partitioned ADD CONSTRAINT chk_a CHECK (a > 0) NO INHERIT;`,
				ErrorString: `cannot add NO INHERIT constraint to partitioned table "partitioned"`,
			},
			{
				Statement: `DROP TABLE partitioned, nonpartitioned;`,
			},
			{
				Statement: `CREATE TABLE unparted (
	a int
);`,
			},
			{
				Statement: `CREATE TABLE fail_part (like unparted);`,
			},
			{
				Statement:   `ALTER TABLE unparted ATTACH PARTITION fail_part FOR VALUES IN ('a');`,
				ErrorString: `table "unparted" is not partitioned`,
			},
			{
				Statement: `DROP TABLE unparted, fail_part;`,
			},
			{
				Statement: `CREATE TABLE list_parted (
	a int NOT NULL,
	b char(2) COLLATE "C",
	CONSTRAINT check_a CHECK (a > 0)
) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE fail_part (LIKE list_parted);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES FROM (1) TO (10);`,
				ErrorString: `invalid bound specification for a list partition`,
			},
			{
				Statement: `DROP TABLE fail_part;`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION nonexistent FOR VALUES IN (1);`,
				ErrorString: `relation "nonexistent" does not exist`,
			},
			{
				Statement: `CREATE ROLE regress_test_me;`,
			},
			{
				Statement: `CREATE ROLE regress_test_not_me;`,
			},
			{
				Statement: `CREATE TABLE not_owned_by_me (LIKE list_parted);`,
			},
			{
				Statement: `ALTER TABLE not_owned_by_me OWNER TO regress_test_not_me;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_test_me;`,
			},
			{
				Statement: `CREATE TABLE owned_by_me (
	a int
) PARTITION BY LIST (a);`,
			},
			{
				Statement:   `ALTER TABLE owned_by_me ATTACH PARTITION not_owned_by_me FOR VALUES IN (1);`,
				ErrorString: `must be owner of table not_owned_by_me`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP TABLE owned_by_me, not_owned_by_me;`,
			},
			{
				Statement: `DROP ROLE regress_test_not_me;`,
			},
			{
				Statement: `DROP ROLE regress_test_me;`,
			},
			{
				Statement: `CREATE TABLE parent (LIKE list_parted);`,
			},
			{
				Statement: `CREATE TABLE child () INHERITS (parent);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION child FOR VALUES IN (1);`,
				ErrorString: `cannot attach inheritance child as partition`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION parent FOR VALUES IN (1);`,
				ErrorString: `cannot attach inheritance parent as partition`,
			},
			{
				Statement: `DROP TABLE parent CASCADE;`,
			},
			{
				Statement: `CREATE TEMP TABLE temp_parted (a int) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE perm_part (a int);`,
			},
			{
				Statement:   `ALTER TABLE temp_parted ATTACH PARTITION perm_part FOR VALUES IN (1);`,
				ErrorString: `cannot attach a permanent relation as partition of temporary relation "temp_parted"`,
			},
			{
				Statement: `DROP TABLE temp_parted, perm_part;`,
			},
			{
				Statement: `CREATE TYPE mytype AS (a int);`,
			},
			{
				Statement: `CREATE TABLE fail_part OF mytype;`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES IN (1);`,
				ErrorString: `cannot attach a typed table as partition`,
			},
			{
				Statement: `DROP TYPE mytype CASCADE;`,
			},
			{
				Statement: `CREATE TABLE fail_part (like list_parted, c int);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES IN (1);`,
				ErrorString: `table "fail_part" contains column "c" not found in parent "list_parted"`,
			},
			{
				Statement: `DROP TABLE fail_part;`,
			},
			{
				Statement: `CREATE TABLE fail_part (a int NOT NULL);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES IN (1);`,
				ErrorString: `child table is missing column "b"`,
			},
			{
				Statement: `DROP TABLE fail_part;`,
			},
			{
				Statement: `CREATE TABLE fail_part (
	b char(3),
	a int NOT NULL
);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES IN (1);`,
				ErrorString: `child table "fail_part" has different type for column "b"`,
			},
			{
				Statement: `ALTER TABLE fail_part ALTER b TYPE char (2) COLLATE "POSIX";`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES IN (1);`,
				ErrorString: `child table "fail_part" has different collation for column "b"`,
			},
			{
				Statement: `DROP TABLE fail_part;`,
			},
			{
				Statement: `CREATE TABLE fail_part (
	b char(2) COLLATE "C",
	a int NOT NULL
);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES IN (1);`,
				ErrorString: `child table is missing constraint "check_a"`,
			},
			{
				Statement: `ALTER TABLE fail_part ADD CONSTRAINT check_a CHECK (a >= 0);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES IN (1);`,
				ErrorString: `child table "fail_part" has different definition for check constraint "check_a"`,
			},
			{
				Statement: `DROP TABLE fail_part;`,
			},
			{
				Statement: `CREATE TABLE part_1 (
	a int NOT NULL,
	b char(2) COLLATE "C",
	CONSTRAINT check_a CHECK (a > 0)
);`,
			},
			{
				Statement: `ALTER TABLE list_parted ATTACH PARTITION part_1 FOR VALUES IN (1);`,
			},
			{
				Statement: `SELECT attislocal, attinhcount FROM pg_attribute WHERE attrelid = 'part_1'::regclass AND attnum > 0;`,
				Results:   []sql.Row{{false, 1}, {false, 1}},
			},
			{
				Statement: `SELECT conislocal, coninhcount FROM pg_constraint WHERE conrelid = 'part_1'::regclass AND conname = 'check_a';`,
				Results:   []sql.Row{{false, 1}},
			},
			{
				Statement: `CREATE TABLE fail_part (LIKE part_1 INCLUDING CONSTRAINTS);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_part FOR VALUES IN (1);`,
				ErrorString: `partition "fail_part" would overlap partition "part_1"`,
			},
			{
				Statement: `DROP TABLE fail_part;`,
			},
			{
				Statement: `CREATE TABLE def_part (LIKE list_parted INCLUDING CONSTRAINTS);`,
			},
			{
				Statement: `ALTER TABLE list_parted ATTACH PARTITION def_part DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE fail_def_part (LIKE part_1 INCLUDING CONSTRAINTS);`,
			},
			{
				Statement:   `ALTER TABLE list_parted ATTACH PARTITION fail_def_part DEFAULT;`,
				ErrorString: `partition "fail_def_part" conflicts with existing default partition "def_part"`,
			},
			{
				Statement: `CREATE TABLE list_parted2 (
	a int,
	b char
) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE part_2 (LIKE list_parted2);`,
			},
			{
				Statement: `INSERT INTO part_2 VALUES (3, 'a');`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 ATTACH PARTITION part_2 FOR VALUES IN (2);`,
				ErrorString: `partition constraint of relation "part_2" is violated by some row`,
			},
			{
				Statement: `DELETE FROM part_2;`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ATTACH PARTITION part_2 FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE TABLE list_parted2_def PARTITION OF list_parted2 DEFAULT;`,
			},
			{
				Statement: `INSERT INTO list_parted2_def VALUES (11, 'z');`,
			},
			{
				Statement: `CREATE TABLE part_3 (LIKE list_parted2);`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 ATTACH PARTITION part_3 FOR VALUES IN (11);`,
				ErrorString: `updated partition constraint for default partition "list_parted2_def" would be violated by some row`,
			},
			{
				Statement: `DELETE FROM list_parted2_def WHERE a = 11;`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ATTACH PARTITION part_3 FOR VALUES IN (11);`,
			},
			{
				Statement: `CREATE TABLE part_3_4 (
	LIKE list_parted2,
	CONSTRAINT check_a CHECK (a IN (3))
);`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ATTACH PARTITION part_3_4 FOR VALUES IN (3, 4);`,
			},
			{
				Statement: `ALTER TABLE list_parted2 DETACH PARTITION part_3_4;`,
			},
			{
				Statement: `ALTER TABLE part_3_4 ALTER a SET NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ATTACH PARTITION part_3_4 FOR VALUES IN (3, 4);`,
			},
			{
				Statement: `ALTER TABLE list_parted2_def ADD CONSTRAINT check_a CHECK (a IN (5, 6));`,
			},
			{
				Statement: `CREATE TABLE part_55_66 PARTITION OF list_parted2 FOR VALUES IN (55, 66);`,
			},
			{
				Statement: `CREATE TABLE range_parted (
	a int,
	b int
) PARTITION BY RANGE (a, b);`,
			},
			{
				Statement: `CREATE TABLE part1 (
	a int NOT NULL CHECK (a = 1),
	b int NOT NULL CHECK (b >= 1 AND b <= 10)
);`,
			},
			{
				Statement: `INSERT INTO part1 VALUES (1, 10);`,
			},
			{
				Statement:   `ALTER TABLE range_parted ATTACH PARTITION part1 FOR VALUES FROM (1, 1) TO (1, 10);`,
				ErrorString: `partition constraint of relation "part1" is violated by some row`,
			},
			{
				Statement: `DELETE FROM part1;`,
			},
			{
				Statement: `ALTER TABLE range_parted ATTACH PARTITION part1 FOR VALUES FROM (1, 1) TO (1, 10);`,
			},
			{
				Statement: `CREATE TABLE part2 (
	a int NOT NULL CHECK (a = 1),
	b int NOT NULL CHECK (b >= 10 AND b < 18)
);`,
			},
			{
				Statement: `ALTER TABLE range_parted ATTACH PARTITION part2 FOR VALUES FROM (1, 10) TO (1, 20);`,
			},
			{
				Statement: `CREATE TABLE partr_def1 PARTITION OF range_parted DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE partr_def2 (LIKE part1 INCLUDING CONSTRAINTS);`,
			},
			{
				Statement:   `ALTER TABLE range_parted ATTACH PARTITION partr_def2 DEFAULT;`,
				ErrorString: `partition "partr_def2" conflicts with existing default partition "partr_def1"`,
			},
			{
				Statement: `INSERT INTO partr_def1 VALUES (2, 10);`,
			},
			{
				Statement: `CREATE TABLE part3 (LIKE range_parted);`,
			},
			{
				Statement:   `ALTER TABLE range_parted ATTACH partition part3 FOR VALUES FROM (2, 10) TO (2, 20);`,
				ErrorString: `updated partition constraint for default partition "partr_def1" would be violated by some row`,
			},
			{
				Statement: `ALTER TABLE range_parted ATTACH partition part3 FOR VALUES FROM (3, 10) TO (3, 20);`,
			},
			{
				Statement: `CREATE TABLE part_5 (
	LIKE list_parted2
) PARTITION BY LIST (b);`,
			},
			{
				Statement: `CREATE TABLE part_5_a PARTITION OF part_5 FOR VALUES IN ('a');`,
			},
			{
				Statement: `INSERT INTO part_5_a (a, b) VALUES (6, 'a');`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 ATTACH PARTITION part_5 FOR VALUES IN (5);`,
				ErrorString: `partition constraint of relation "part_5_a" is violated by some row`,
			},
			{
				Statement: `DELETE FROM part_5_a WHERE a NOT IN (3);`,
			},
			{
				Statement: `ALTER TABLE part_5 ADD CONSTRAINT check_a CHECK (a IS NOT NULL AND a = 5);`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ATTACH PARTITION part_5 FOR VALUES IN (5);`,
			},
			{
				Statement: `ALTER TABLE list_parted2 DETACH PARTITION part_5;`,
			},
			{
				Statement: `ALTER TABLE part_5 DROP CONSTRAINT check_a;`,
			},
			{
				Statement: `ALTER TABLE part_5 ADD CONSTRAINT check_a CHECK (a IN (5)), ALTER a SET NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ATTACH PARTITION part_5 FOR VALUES IN (5);`,
			},
			{
				Statement: `-- attached differs from the parent.  It should not affect the constraint-
CREATE TABLE part_6 (
	c int,
	LIKE list_parted2,
	CONSTRAINT check_a CHECK (a IS NOT NULL AND a = 6)
);`,
			},
			{
				Statement: `ALTER TABLE part_6 DROP c;`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ATTACH PARTITION part_6 FOR VALUES IN (6);`,
			},
			{
				Statement: `CREATE TABLE part_7 (
	LIKE list_parted2,
	CONSTRAINT check_a CHECK (a IS NOT NULL AND a = 7)
) PARTITION BY LIST (b);`,
			},
			{
				Statement: `CREATE TABLE part_7_a_null (
	c int,
	d int,
	e int,
	LIKE list_parted2,  -- 'a' will have attnum = 4
	CONSTRAINT check_b CHECK (b IS NULL OR b = 'a'),
	CONSTRAINT check_a CHECK (a IS NOT NULL AND a = 7)
);`,
			},
			{
				Statement: `ALTER TABLE part_7_a_null DROP c, DROP d, DROP e;`,
			},
			{
				Statement: `ALTER TABLE part_7 ATTACH PARTITION part_7_a_null FOR VALUES IN ('a', null);`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ATTACH PARTITION part_7 FOR VALUES IN (7);`,
			},
			{
				Statement: `ALTER TABLE list_parted2 DETACH PARTITION part_7;`,
			},
			{
				Statement: `ALTER TABLE part_7 DROP CONSTRAINT check_a; -- thusly, scan won't be skipped`,
			},
			{
				Statement: `INSERT INTO part_7 (a, b) VALUES (8, null), (9, 'a');`,
			},
			{
				Statement: `SELECT tableoid::regclass, a, b FROM part_7 order by a;`,
				Results:   []sql.Row{{`part_7_a_null`, 8, ``}, {`part_7_a_null`, 9, `a`}},
			},
			{
				Statement:   `ALTER TABLE list_parted2 ATTACH PARTITION part_7 FOR VALUES IN (7);`,
				ErrorString: `partition constraint of relation "part_7_a_null" is violated by some row`,
			},
			{
				Statement: `ALTER TABLE part_5 DROP CONSTRAINT check_a;`,
			},
			{
				Statement: `CREATE TABLE part5_def PARTITION OF part_5 DEFAULT PARTITION BY LIST(a);`,
			},
			{
				Statement: `CREATE TABLE part5_def_p1 PARTITION OF part5_def FOR VALUES IN (5);`,
			},
			{
				Statement: `INSERT INTO part5_def_p1 VALUES (5, 'y');`,
			},
			{
				Statement: `CREATE TABLE part5_p1 (LIKE part_5);`,
			},
			{
				Statement:   `ALTER TABLE part_5 ATTACH PARTITION part5_p1 FOR VALUES IN ('y');`,
				ErrorString: `updated partition constraint for default partition "part5_def_p1" would be violated by some row`,
			},
			{
				Statement: `DELETE FROM part5_def_p1 WHERE b = 'y';`,
			},
			{
				Statement: `ALTER TABLE part_5 ATTACH PARTITION part5_p1 FOR VALUES IN ('y');`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 ATTACH PARTITION part_2 FOR VALUES IN (2);`,
				ErrorString: `"part_2" is already a partition`,
			},
			{
				Statement:   `ALTER TABLE part_5 ATTACH PARTITION list_parted2 FOR VALUES IN ('b');`,
				ErrorString: `circular inheritance not allowed`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 ATTACH PARTITION list_parted2 FOR VALUES IN (0);`,
				ErrorString: `circular inheritance not allowed`,
			},
			{
				Statement: `CREATE TABLE quuux (a int, b text) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE quuux_default PARTITION OF quuux DEFAULT PARTITION BY LIST (b);`,
			},
			{
				Statement: `CREATE TABLE quuux_default1 PARTITION OF quuux_default (
	CONSTRAINT check_1 CHECK (a IS NOT NULL AND a = 1)
) FOR VALUES IN ('b');`,
			},
			{
				Statement: `CREATE TABLE quuux1 (a int, b text);`,
			},
			{
				Statement: `ALTER TABLE quuux ATTACH PARTITION quuux1 FOR VALUES IN (1); -- validate!`,
			},
			{
				Statement: `CREATE TABLE quuux2 (a int, b text);`,
			},
			{
				Statement: `ALTER TABLE quuux ATTACH PARTITION quuux2 FOR VALUES IN (2); -- skip validation`,
			},
			{
				Statement: `DROP TABLE quuux1, quuux2;`,
			},
			{
				Statement: `CREATE TABLE quuux1 PARTITION OF quuux FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE quuux2 PARTITION OF quuux FOR VALUES IN (2);`,
			},
			{
				Statement: `DROP TABLE quuux;`,
			},
			{
				Statement: `CREATE TABLE hash_parted (
	a int,
	b int
) PARTITION BY HASH (a part_test_int4_ops);`,
			},
			{
				Statement: `CREATE TABLE hpart_1 PARTITION OF hash_parted FOR VALUES WITH (MODULUS 4, REMAINDER 0);`,
			},
			{
				Statement: `CREATE TABLE fail_part (LIKE hpart_1);`,
			},
			{
				Statement:   `ALTER TABLE hash_parted ATTACH PARTITION fail_part FOR VALUES WITH (MODULUS 8, REMAINDER 4);`,
				ErrorString: `partition "fail_part" would overlap partition "hpart_1"`,
			},
			{
				Statement:   `ALTER TABLE hash_parted ATTACH PARTITION fail_part FOR VALUES WITH (MODULUS 8, REMAINDER 0);`,
				ErrorString: `partition "fail_part" would overlap partition "hpart_1"`,
			},
			{
				Statement: `DROP TABLE fail_part;`,
			},
			{
				Statement: `CREATE TABLE hpart_2 (LIKE hash_parted);`,
			},
			{
				Statement: `INSERT INTO hpart_2 VALUES (3, 0);`,
			},
			{
				Statement:   `ALTER TABLE hash_parted ATTACH PARTITION hpart_2 FOR VALUES WITH (MODULUS 4, REMAINDER 1);`,
				ErrorString: `partition constraint of relation "hpart_2" is violated by some row`,
			},
			{
				Statement: `DELETE FROM hpart_2;`,
			},
			{
				Statement: `ALTER TABLE hash_parted ATTACH PARTITION hpart_2 FOR VALUES WITH (MODULUS 4, REMAINDER 1);`,
			},
			{
				Statement: `CREATE TABLE hpart_5 (
	LIKE hash_parted
) PARTITION BY LIST (b);`,
			},
			{
				Statement: `CREATE TABLE hpart_5_a PARTITION OF hpart_5 FOR VALUES IN ('1', '2', '3');`,
			},
			{
				Statement: `INSERT INTO hpart_5_a (a, b) VALUES (7, 1);`,
			},
			{
				Statement:   `ALTER TABLE hash_parted ATTACH PARTITION hpart_5 FOR VALUES WITH (MODULUS 4, REMAINDER 2);`,
				ErrorString: `partition constraint of relation "hpart_5_a" is violated by some row`,
			},
			{
				Statement: `DELETE FROM hpart_5_a;`,
			},
			{
				Statement: `ALTER TABLE hash_parted ATTACH PARTITION hpart_5 FOR VALUES WITH (MODULUS 4, REMAINDER 2);`,
			},
			{
				Statement: `CREATE TABLE fail_part(LIKE hash_parted);`,
			},
			{
				Statement:   `ALTER TABLE hash_parted ATTACH PARTITION fail_part FOR VALUES WITH (MODULUS 0, REMAINDER 1);`,
				ErrorString: `modulus for hash partition must be an integer value greater than zero`,
			},
			{
				Statement:   `ALTER TABLE hash_parted ATTACH PARTITION fail_part FOR VALUES WITH (MODULUS 8, REMAINDER 8);`,
				ErrorString: `remainder for hash partition must be less than modulus`,
			},
			{
				Statement:   `ALTER TABLE hash_parted ATTACH PARTITION fail_part FOR VALUES WITH (MODULUS 3, REMAINDER 2);`,
				ErrorString: `every hash partition modulus must be a factor of the next larger modulus`,
			},
			{
				Statement: `DROP TABLE fail_part;`,
			},
			{
				Statement: `CREATE TABLE regular_table (a int);`,
			},
			{
				Statement:   `ALTER TABLE regular_table DETACH PARTITION any_name;`,
				ErrorString: `table "regular_table" is not partitioned`,
			},
			{
				Statement: `DROP TABLE regular_table;`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 DETACH PARTITION part_4;`,
				ErrorString: `relation "part_4" does not exist`,
			},
			{
				Statement:   `ALTER TABLE hash_parted DETACH PARTITION hpart_4;`,
				ErrorString: `relation "hpart_4" does not exist`,
			},
			{
				Statement: `CREATE TABLE not_a_part (a int);`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 DETACH PARTITION not_a_part;`,
				ErrorString: `relation "not_a_part" is not a partition of relation "list_parted2"`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 DETACH PARTITION part_1;`,
				ErrorString: `relation "part_1" is not a partition of relation "list_parted2"`,
			},
			{
				Statement:   `ALTER TABLE hash_parted DETACH PARTITION not_a_part;`,
				ErrorString: `relation "not_a_part" is not a partition of relation "hash_parted"`,
			},
			{
				Statement: `DROP TABLE not_a_part;`,
			},
			{
				Statement: `ALTER TABLE list_parted2 DETACH PARTITION part_3_4;`,
			},
			{
				Statement: `SELECT attinhcount, attislocal FROM pg_attribute WHERE attrelid = 'part_3_4'::regclass AND attnum > 0;`,
				Results:   []sql.Row{{0, true}, {0, true}},
			},
			{
				Statement: `SELECT coninhcount, conislocal FROM pg_constraint WHERE conrelid = 'part_3_4'::regclass AND conname = 'check_a';`,
				Results:   []sql.Row{{0, true}},
			},
			{
				Statement: `DROP TABLE part_3_4;`,
			},
			{
				Statement: `CREATE TABLE range_parted2 (
    a int
) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE part_rp PARTITION OF range_parted2 FOR VALUES FROM (0) to (100);`,
			},
			{
				Statement: `ALTER TABLE range_parted2 DETACH PARTITION part_rp;`,
			},
			{
				Statement: `DROP TABLE range_parted2;`,
			},
			{
				Statement: `SELECT * from part_rp;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE part_rp;`,
			},
			{
				Statement: `CREATE TABLE range_parted2 (
	a int
) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE part_rp PARTITION OF range_parted2 FOR VALUES FROM (0) to (100);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `ALTER TABLE range_parted2 DETACH PARTITION part_rp CONCURRENTLY;`,
				ErrorString: `ALTER TABLE ... DETACH CONCURRENTLY cannot run inside a transaction block`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE TABLE part_rpd PARTITION OF range_parted2 DEFAULT;`,
			},
			{
				Statement:   `ALTER TABLE range_parted2 DETACH PARTITION part_rp CONCURRENTLY;`,
				ErrorString: `cannot detach partitions concurrently when a default partition exists`,
			},
			{
				Statement:   `ALTER TABLE range_parted2 DETACH PARTITION part_rpd CONCURRENTLY;`,
				ErrorString: `cannot detach partitions concurrently when a default partition exists`,
			},
			{
				Statement: `DROP TABLE part_rpd;`,
			},
			{
				Statement: `ALTER TABLE range_parted2 DETACH PARTITION part_rp CONCURRENTLY;`,
			},
			{
				Statement: `\d+ range_parted2
                         Partitioned table "public.range_parted2"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
Partition key: RANGE (a)
Number of partitions: 0
\d part_rp
              Table "public.part_rp"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Check constraints:
    "part_rp_a_check" CHECK (a IS NOT NULL AND a >= 0 AND a < 100)
CREATE TABLE part_rp100 PARTITION OF range_parted2 (CHECK (a>=123 AND a<133 AND a IS NOT NULL)) FOR VALUES FROM (100) to (200);`,
			},
			{
				Statement: `ALTER TABLE range_parted2 DETACH PARTITION part_rp100 CONCURRENTLY;`,
			},
			{
				Statement: `\d part_rp100
             Table "public.part_rp100"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Check constraints:
    "part_rp100_a_check" CHECK (a >= 123 AND a < 133 AND a IS NOT NULL)
DROP TABLE range_parted2;`,
			},
			{
				Statement:   `ALTER TABLE ONLY list_parted2 ADD COLUMN c int;`,
				ErrorString: `column must be added to child tables too`,
			},
			{
				Statement:   `ALTER TABLE ONLY list_parted2 DROP COLUMN b;`,
				ErrorString: `cannot drop column from only the partitioned table when partitions exist`,
			},
			{
				Statement:   `ALTER TABLE part_2 ADD COLUMN c text;`,
				ErrorString: `cannot add column to a partition`,
			},
			{
				Statement:   `ALTER TABLE part_2 DROP COLUMN b;`,
				ErrorString: `cannot drop inherited column "b"`,
			},
			{
				Statement:   `ALTER TABLE part_2 RENAME COLUMN b to c;`,
				ErrorString: `cannot rename inherited column "b"`,
			},
			{
				Statement:   `ALTER TABLE part_2 ALTER COLUMN b TYPE text;`,
				ErrorString: `cannot alter inherited column "b"`,
			},
			{
				Statement:   `ALTER TABLE ONLY list_parted2 ALTER b SET NOT NULL;`,
				ErrorString: `constraint must be added to child tables too`,
			},
			{
				Statement:   `ALTER TABLE ONLY list_parted2 ADD CONSTRAINT check_b CHECK (b <> 'zz');`,
				ErrorString: `constraint must be added to child tables too`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ALTER b SET NOT NULL;`,
			},
			{
				Statement:   `ALTER TABLE ONLY list_parted2 ALTER b DROP NOT NULL;`,
				ErrorString: `cannot remove constraint from only the partitioned table when partitions exist`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ADD CONSTRAINT check_b CHECK (b <> 'zz');`,
			},
			{
				Statement:   `ALTER TABLE ONLY list_parted2 DROP CONSTRAINT check_b;`,
				ErrorString: `cannot remove constraint from only the partitioned table when partitions exist`,
			},
			{
				Statement: `CREATE TABLE parted_no_parts (a int) PARTITION BY LIST (a);`,
			},
			{
				Statement: `ALTER TABLE ONLY parted_no_parts ALTER a SET NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE ONLY parted_no_parts ADD CONSTRAINT check_a CHECK (a > 0);`,
			},
			{
				Statement: `ALTER TABLE ONLY parted_no_parts ALTER a DROP NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE ONLY parted_no_parts DROP CONSTRAINT check_a;`,
			},
			{
				Statement: `DROP TABLE parted_no_parts;`,
			},
			{
				Statement: `ALTER TABLE list_parted2 ALTER b SET NOT NULL, ADD CONSTRAINT check_a2 CHECK (a > 0);`,
			},
			{
				Statement:   `ALTER TABLE part_2 ALTER b DROP NOT NULL;`,
				ErrorString: `column "b" is marked NOT NULL in parent table`,
			},
			{
				Statement:   `ALTER TABLE part_2 DROP CONSTRAINT check_a2;`,
				ErrorString: `cannot drop inherited constraint "check_a2" of relation "part_2"`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 add constraint check_b2 check (b <> 'zz') NO INHERIT;`,
				ErrorString: `cannot add NO INHERIT constraint to partitioned table "list_parted2"`,
			},
			{
				Statement:   `CREATE TABLE inh_test () INHERITS (part_2);`,
				ErrorString: `cannot inherit from partition "part_2"`,
			},
			{
				Statement: `CREATE TABLE inh_test (LIKE part_2);`,
			},
			{
				Statement:   `ALTER TABLE inh_test INHERIT part_2;`,
				ErrorString: `cannot inherit from a partition`,
			},
			{
				Statement:   `ALTER TABLE part_2 INHERIT inh_test;`,
				ErrorString: `cannot change inheritance of a partition`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 DROP COLUMN b;`,
				ErrorString: `cannot drop column "b" because it is part of the partition key of relation "part_5"`,
			},
			{
				Statement:   `ALTER TABLE list_parted2 ALTER COLUMN b TYPE text;`,
				ErrorString: `cannot alter column "b" because it is part of the partition key of relation "part_5"`,
			},
			{
				Statement: `ALTER TABLE list_parted DROP COLUMN b;`,
			},
			{
				Statement: `SELECT * FROM list_parted;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE list_parted, list_parted2, range_parted;`,
			},
			{
				Statement: `DROP TABLE fail_def_part;`,
			},
			{
				Statement: `DROP TABLE hash_parted;`,
			},
			{
				Statement: `create table p (a int, b int) partition by range (a, b);`,
			},
			{
				Statement: `create table p1 (b int, a int not null) partition by range (b);`,
			},
			{
				Statement: `create table p11 (like p1);`,
			},
			{
				Statement: `alter table p11 drop a;`,
			},
			{
				Statement: `alter table p11 add a int;`,
			},
			{
				Statement: `alter table p11 drop a;`,
			},
			{
				Statement: `alter table p11 add a int not null;`,
			},
			{
				Statement: `select attrelid::regclass, attname, attnum
from pg_attribute
where attname = 'a'
 and (attrelid = 'p'::regclass
   or attrelid = 'p1'::regclass
   or attrelid = 'p11'::regclass)
order by attrelid::regclass::text;`,
				Results: []sql.Row{{`p`, `a`, 1}, {`p1`, `a`, 2}, {`p11`, `a`, 4}},
			},
			{
				Statement: `alter table p1 attach partition p11 for values from (2) to (5);`,
			},
			{
				Statement: `insert into p1 (a, b) values (2, 3);`,
			},
			{
				Statement:   `alter table p attach partition p1 for values from (1, 2) to (1, 10);`,
				ErrorString: `partition constraint of relation "p11" is violated by some row`,
			},
			{
				Statement: `drop table p;`,
			},
			{
				Statement: `drop table p1;`,
			},
			{
				Statement: `create table parted_validate_test (a int) partition by list (a);`,
			},
			{
				Statement: `create table parted_validate_test_1 partition of parted_validate_test for values in (0, 1);`,
			},
			{
				Statement: `alter table parted_validate_test add constraint parted_validate_test_chka check (a > 0) not valid;`,
			},
			{
				Statement: `alter table parted_validate_test validate constraint parted_validate_test_chka;`,
			},
			{
				Statement: `drop table parted_validate_test;`,
			},
			{
				Statement: `CREATE TABLE attmp(i integer);`,
			},
			{
				Statement: `INSERT INTO attmp VALUES (1);`,
			},
			{
				Statement: `ALTER TABLE attmp ALTER COLUMN i SET (n_distinct = 1, n_distinct_inherited = 2);`,
			},
			{
				Statement: `ALTER TABLE attmp ALTER COLUMN i RESET (n_distinct_inherited);`,
			},
			{
				Statement: `ANALYZE attmp;`,
			},
			{
				Statement: `DROP TABLE attmp;`,
			},
			{
				Statement: `DROP USER regress_alter_table_user1;`,
			},
			{
				Statement: `create table defpart_attach_test (a int) partition by list (a);`,
			},
			{
				Statement: `create table defpart_attach_test1 partition of defpart_attach_test for values in (1);`,
			},
			{
				Statement: `create table defpart_attach_test_d (b int, a int);`,
			},
			{
				Statement: `alter table defpart_attach_test_d drop b;`,
			},
			{
				Statement: `insert into defpart_attach_test_d values (1), (2);`,
			},
			{
				Statement:   `alter table defpart_attach_test attach partition defpart_attach_test_d default;`,
				ErrorString: `partition constraint of relation "defpart_attach_test_d" is violated by some row`,
			},
			{
				Statement: `delete from defpart_attach_test_d where a = 1;`,
			},
			{
				Statement: `alter table defpart_attach_test_d add check (a > 1);`,
			},
			{
				Statement: `alter table defpart_attach_test attach partition defpart_attach_test_d default;`,
			},
			{
				Statement: `create table defpart_attach_test_2 (like defpart_attach_test_d);`,
			},
			{
				Statement:   `alter table defpart_attach_test attach partition defpart_attach_test_2 for values in (2);`,
				ErrorString: `updated partition constraint for default partition "defpart_attach_test_d" would be violated by some row`,
			},
			{
				Statement: `drop table defpart_attach_test;`,
			},
			{
				Statement: `create table perm_part_parent (a int) partition by list (a);`,
			},
			{
				Statement: `create temp table temp_part_parent (a int) partition by list (a);`,
			},
			{
				Statement: `create table perm_part_child (a int);`,
			},
			{
				Statement: `create temp table temp_part_child (a int);`,
			},
			{
				Statement:   `alter table temp_part_parent attach partition perm_part_child default; -- error`,
				ErrorString: `cannot attach a permanent relation as partition of temporary relation "temp_part_parent"`,
			},
			{
				Statement:   `alter table perm_part_parent attach partition temp_part_child default; -- error`,
				ErrorString: `cannot attach a temporary relation as partition of permanent relation "perm_part_parent"`,
			},
			{
				Statement: `alter table temp_part_parent attach partition temp_part_child default; -- ok`,
			},
			{
				Statement: `drop table perm_part_parent cascade;`,
			},
			{
				Statement: `drop table temp_part_parent cascade;`,
			},
			{
				Statement: `create table tab_part_attach (a int) partition by list (a);`,
			},
			{
				Statement: `create or replace function func_part_attach() returns trigger
  language plpgsql as $$
  begin
    execute 'create table tab_part_attach_1 (a int)';`,
			},
			{
				Statement: `    execute 'alter table tab_part_attach attach partition tab_part_attach_1 for values in (1)';`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end $$;`,
			},
			{
				Statement: `create trigger trig_part_attach before insert on tab_part_attach
  for each statement execute procedure func_part_attach();`,
			},
			{
				Statement:   `insert into tab_part_attach values (1);`,
				ErrorString: `cannot ALTER TABLE "tab_part_attach" because it is being used by active queries in this session`,
			},
			{
				Statement: `CONTEXT:  SQL statement "alter table tab_part_attach attach partition tab_part_attach_1 for values in (1)"
PL/pgSQL function func_part_attach() line 4 at EXECUTE
drop table tab_part_attach;`,
			},
			{
				Statement: `drop function func_part_attach();`,
			},
			{
				Statement: `create function at_test_sql_partop (int4, int4) returns int language sql
as $$ select case when $1 = $2 then 0 when $1 > $2 then 1 else -1 end; $$;`,
			},
			{
				Statement: `create operator class at_test_sql_partop for type int4 using btree as
    operator 1 < (int4, int4), operator 2 <= (int4, int4),
    operator 3 = (int4, int4), operator 4 >= (int4, int4),
    operator 5 > (int4, int4), function 1 at_test_sql_partop(int4, int4);`,
			},
			{
				Statement: `create table at_test_sql_partop (a int) partition by range (a at_test_sql_partop);`,
			},
			{
				Statement: `create table at_test_sql_partop_1 (a int);`,
			},
			{
				Statement: `alter table at_test_sql_partop attach partition at_test_sql_partop_1 for values from (0) to (10);`,
			},
			{
				Statement: `drop table at_test_sql_partop;`,
			},
			{
				Statement: `drop operator class at_test_sql_partop using btree;`,
			},
			{
				Statement: `drop function at_test_sql_partop;`,
			},
			{
				Statement: `/* Test case for bug #16242 */
create table bar1 (a integer, b integer not null default 1)
  partition by range (a);`,
			},
			{
				Statement: `create table bar2 (a integer);`,
			},
			{
				Statement: `insert into bar2 values (1);`,
			},
			{
				Statement: `alter table bar2 add column b integer not null default 1;`,
			},
			{
				Statement: `alter table bar1 attach partition bar2 default;`,
			},
			{
				Statement: `select * from bar1;`,
				Results:   []sql.Row{{1, 1}},
			},
			{
				Statement: `create function xtrig()
  returns trigger language plpgsql
as $$
  declare
    r record;`,
			},
			{
				Statement: `  begin
    for r in select * from old loop
      raise info 'a=%, b=%', r.a, r.b;`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `    return NULL;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger xtrig
  after update on bar1
  referencing old table as old
  for each statement execute procedure xtrig();`,
			},
			{
				Statement: `update bar1 set a = a + 1;`,
			},
			{
				Statement: `INFO:  a=1, b=1
/* End test case for bug #16242 */
/* Test case for bug #17409 */
create table attbl (p1 int constraint pk_attbl primary key);`,
			},
			{
				Statement: `create table atref (c1 int references attbl(p1));`,
			},
			{
				Statement: `cluster attbl using pk_attbl;`,
			},
			{
				Statement: `alter table attbl alter column p1 set data type bigint;`,
			},
			{
				Statement: `alter table atref alter column c1 set data type bigint;`,
			},
			{
				Statement: `drop table attbl, atref;`,
			},
			{
				Statement: `create table attbl (p1 int constraint pk_attbl primary key);`,
			},
			{
				Statement: `alter table attbl replica identity using index pk_attbl;`,
			},
			{
				Statement: `create table atref (c1 int references attbl(p1));`,
			},
			{
				Statement: `alter table attbl alter column p1 set data type bigint;`,
			},
			{
				Statement: `alter table atref alter column c1 set data type bigint;`,
			},
			{
				Statement: `drop table attbl, atref;`,
			},
			{
				Statement: `/* End test case for bug #17409 */
create table alttype_cluster (a int);`,
			},
			{
				Statement: `alter table alttype_cluster add primary key (a);`,
			},
			{
				Statement: `create index alttype_cluster_ind on alttype_cluster (a);`,
			},
			{
				Statement: `alter table alttype_cluster cluster on alttype_cluster_ind;`,
			},
			{
				Statement: `select indexrelid::regclass, indisclustered from pg_index
  where indrelid = 'alttype_cluster'::regclass
  order by indexrelid::regclass::text;`,
				Results: []sql.Row{{`alttype_cluster_ind`, true}, {`alttype_cluster_pkey`, false}},
			},
			{
				Statement: `alter table alttype_cluster alter a type bigint;`,
			},
			{
				Statement: `select indexrelid::regclass, indisclustered from pg_index
  where indrelid = 'alttype_cluster'::regclass
  order by indexrelid::regclass::text;`,
				Results: []sql.Row{{`alttype_cluster_ind`, true}, {`alttype_cluster_pkey`, false}},
			},
			{
				Statement: `alter table alttype_cluster cluster on alttype_cluster_pkey;`,
			},
			{
				Statement: `select indexrelid::regclass, indisclustered from pg_index
  where indrelid = 'alttype_cluster'::regclass
  order by indexrelid::regclass::text;`,
				Results: []sql.Row{{`alttype_cluster_ind`, false}, {`alttype_cluster_pkey`, true}},
			},
			{
				Statement: `alter table alttype_cluster alter a type int;`,
			},
			{
				Statement: `select indexrelid::regclass, indisclustered from pg_index
  where indrelid = 'alttype_cluster'::regclass
  order by indexrelid::regclass::text;`,
				Results: []sql.Row{{`alttype_cluster_ind`, false}, {`alttype_cluster_pkey`, true}},
			},
			{
				Statement: `drop table alttype_cluster;`,
			},
			{
				Statement: `create table target_parted (a int, b int) partition by list (a);`,
			},
			{
				Statement: `create table attach_parted (a int, b int) partition by list (b);`,
			},
			{
				Statement: `create table attach_parted_part1 partition of attach_parted for values in (1);`,
			},
			{
				Statement: `insert into attach_parted_part1 values (1, 1);`,
			},
			{
				Statement: `alter table target_parted attach partition attach_parted for values in (1);`,
			},
			{
				Statement:   `insert into attach_parted_part1 values (2, 1);`,
				ErrorString: `new row for relation "attach_parted_part1" violates partition constraint`,
			},
			{
				Statement: `alter table target_parted detach partition attach_parted;`,
			},
			{
				Statement: `insert into attach_parted_part1 values (2, 1);`,
			},
			{
				Statement: `create schema alter1;`,
			},
			{
				Statement: `create schema alter2;`,
			},
			{
				Statement: `create table alter1.t1 (a int);`,
			},
			{
				Statement: `set client_min_messages = 'ERROR';`,
			},
			{
				Statement: `create publication pub1 for table alter1.t1, tables in schema alter2;`,
			},
			{
				Statement: `reset client_min_messages;`,
			},
			{
				Statement: `alter table alter1.t1 set schema alter2;`,
			},
			{
				Statement: `\d+ alter2.t1
                                    Table "alter2.t1"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
Publications:
    "pub1"
drop publication pub1;`,
			},
			{
				Statement: `drop schema alter1 cascade;`,
			},
			{
				Statement: `drop schema alter2 cascade;`,
			},
		},
	})
}
