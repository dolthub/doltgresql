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

func TestTablespace(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_tablespace)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_tablespace,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `CREATE TABLESPACE regress_tblspace LOCATION 'relative'; -- fail`,
				ErrorString: `tablespace location must be an absolute path`,
			},
			{
				Statement:   `CREATE TABLESPACE regress_tblspace LOCATION ''; -- fail`,
				ErrorString: `tablespace location must be an absolute path`,
			},
			{
				Statement: `SET allow_in_place_tablespaces = true;`,
			},
			{
				Statement:   `CREATE TABLESPACE regress_tblspacewith LOCATION '' WITH (some_nonexistent_parameter = true); -- fail`,
				ErrorString: `unrecognized parameter "some_nonexistent_parameter"`,
			},
			{
				Statement: `CREATE TABLESPACE regress_tblspacewith LOCATION '' WITH (random_page_cost = 3.0); -- ok`,
			},
			{
				Statement: `SELECT spcoptions FROM pg_tablespace WHERE spcname = 'regress_tblspacewith';`,
				Results:   []sql.Row{{`{random_page_cost=3.0}`}},
			},
			{
				Statement: `DROP TABLESPACE regress_tblspacewith;`,
			},
			{
				Statement: `CREATE TABLESPACE regress_tblspace LOCATION '';`,
			},
			{
				Statement: `SELECT regexp_replace(pg_tablespace_location(oid), '(pg_tblspc)/(\d+)', '\1/NNN')
  FROM pg_tablespace  WHERE spcname = 'regress_tblspace';`,
				Results: []sql.Row{{`pg_tblspc/NNN`}},
			},
			{
				Statement: `ALTER TABLESPACE regress_tblspace SET (random_page_cost = 1.0, seq_page_cost = 1.1);`,
			},
			{
				Statement:   `ALTER TABLESPACE regress_tblspace SET (some_nonexistent_parameter = true);  -- fail`,
				ErrorString: `unrecognized parameter "some_nonexistent_parameter"`,
			},
			{
				Statement:   `ALTER TABLESPACE regress_tblspace RESET (random_page_cost = 2.0); -- fail`,
				ErrorString: `RESET must not include values for parameters`,
			},
			{
				Statement: `ALTER TABLESPACE regress_tblspace RESET (random_page_cost, effective_io_concurrency); -- ok`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) TABLE pg_am;`,
				ErrorString: `cannot move system relation "pg_am_name_index"`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) TABLE CONCURRENTLY pg_am;`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) TABLE pg_authid;`,
				ErrorString: `cannot move system relation "pg_authid_rolname_index"`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) TABLE CONCURRENTLY pg_authid;`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) INDEX pg_toast.pg_toast_1260_index;`,
				ErrorString: `cannot move system relation "pg_toast_1260_index"`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) INDEX CONCURRENTLY pg_toast.pg_toast_1260_index;`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) TABLE pg_toast.pg_toast_1260;`,
				ErrorString: `cannot move system relation "pg_toast_1260_index"`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) TABLE CONCURRENTLY pg_toast.pg_toast_1260;`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement:   `REINDEX (TABLESPACE pg_global) TABLE pg_authid;`,
				ErrorString: `cannot move system relation "pg_authid_rolname_index"`,
			},
			{
				Statement:   `REINDEX (TABLESPACE pg_global) TABLE CONCURRENTLY pg_authid;`,
				ErrorString: `cannot reindex system catalogs concurrently`,
			},
			{
				Statement: `CREATE TABLE regress_tblspace_test_tbl (num1 bigint, num2 double precision, t text);`,
			},
			{
				Statement: `INSERT INTO regress_tblspace_test_tbl (num1, num2, t)
  SELECT round(random()*100), random(), 'text'
  FROM generate_series(1, 10) s(i);`,
			},
			{
				Statement: `CREATE INDEX regress_tblspace_test_tbl_idx ON regress_tblspace_test_tbl (num1);`,
			},
			{
				Statement:   `REINDEX (TABLESPACE pg_global) INDEX regress_tblspace_test_tbl_idx;`,
				ErrorString: `only shared relations can be placed in pg_global tablespace`,
			},
			{
				Statement:   `REINDEX (TABLESPACE pg_global) INDEX CONCURRENTLY regress_tblspace_test_tbl_idx;`,
				ErrorString: `cannot move non-shared relation to tablespace "pg_global"`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `REINDEX (TABLESPACE regress_tblspace) INDEX regress_tblspace_test_tbl_idx;`,
			},
			{
				Statement: `REINDEX (TABLESPACE regress_tblspace) TABLE regress_tblspace_test_tbl;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT c.relname FROM pg_class c, pg_tablespace s
  WHERE c.reltablespace = s.oid AND s.spcname = 'regress_tblspace';`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT relfilenode as main_filenode FROM pg_class
  WHERE relname = 'regress_tblspace_test_tbl_idx' \gset
SELECT relfilenode as toast_filenode FROM pg_class
  WHERE oid =
    (SELECT i.indexrelid
       FROM pg_class c,
            pg_index i
       WHERE i.indrelid = c.reltoastrelid AND
             c.relname = 'regress_tblspace_test_tbl') \gset
REINDEX (TABLESPACE regress_tblspace) TABLE regress_tblspace_test_tbl;`,
			},
			{
				Statement: `SELECT c.relname FROM pg_class c, pg_tablespace s
  WHERE c.reltablespace = s.oid AND s.spcname = 'regress_tblspace'
  ORDER BY c.relname;`,
				Results: []sql.Row{{`regress_tblspace_test_tbl_idx`}},
			},
			{
				Statement: `ALTER TABLE regress_tblspace_test_tbl SET TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `ALTER TABLE regress_tblspace_test_tbl SET TABLESPACE pg_default;`,
			},
			{
				Statement: `SELECT c.relname FROM pg_class c, pg_tablespace s
  WHERE c.reltablespace = s.oid AND s.spcname = 'regress_tblspace'
  ORDER BY c.relname;`,
				Results: []sql.Row{{`regress_tblspace_test_tbl_idx`}},
			},
			{
				Statement: `ALTER INDEX regress_tblspace_test_tbl_idx SET TABLESPACE pg_default;`,
			},
			{
				Statement: `SELECT c.relname FROM pg_class c, pg_tablespace s
  WHERE c.reltablespace = s.oid AND s.spcname = 'regress_tblspace'
  ORDER BY c.relname;`,
				Results: []sql.Row{},
			},
			{
				Statement: `REINDEX (TABLESPACE regress_tblspace, CONCURRENTLY) TABLE regress_tblspace_test_tbl;`,
			},
			{
				Statement: `SELECT c.relname FROM pg_class c, pg_tablespace s
  WHERE c.reltablespace = s.oid AND s.spcname = 'regress_tblspace'
  ORDER BY c.relname;`,
				Results: []sql.Row{{`regress_tblspace_test_tbl_idx`}},
			},
			{
				Statement: `SELECT relfilenode = :main_filenode AS main_same FROM pg_class
  WHERE relname = 'regress_tblspace_test_tbl_idx';`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT relfilenode = :toast_filenode as toast_same FROM pg_class
  WHERE oid =
    (SELECT i.indexrelid
       FROM pg_class c,
            pg_index i
       WHERE i.indrelid = c.reltoastrelid AND
             c.relname = 'regress_tblspace_test_tbl');`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `DROP TABLE regress_tblspace_test_tbl;`,
			},
			{
				Statement: `CREATE TABLE tbspace_reindex_part (c1 int, c2 int) PARTITION BY RANGE (c1);`,
			},
			{
				Statement: `CREATE TABLE tbspace_reindex_part_0 PARTITION OF tbspace_reindex_part
  FOR VALUES FROM (0) TO (10) PARTITION BY list (c2);`,
			},
			{
				Statement: `CREATE TABLE tbspace_reindex_part_0_1 PARTITION OF tbspace_reindex_part_0
  FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE tbspace_reindex_part_0_2 PARTITION OF tbspace_reindex_part_0
  FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE TABLE tbspace_reindex_part_10 PARTITION OF tbspace_reindex_part
   FOR VALUES FROM (10) TO (20) PARTITION BY list (c2);`,
			},
			{
				Statement: `CREATE INDEX tbspace_reindex_part_index ON ONLY tbspace_reindex_part (c1);`,
			},
			{
				Statement: `CREATE INDEX tbspace_reindex_part_index_0 ON ONLY tbspace_reindex_part_0 (c1);`,
			},
			{
				Statement: `ALTER INDEX tbspace_reindex_part_index ATTACH PARTITION tbspace_reindex_part_index_0;`,
			},
			{
				Statement: `CREATE INDEX tbspace_reindex_part_index_10 ON ONLY tbspace_reindex_part_10 (c1);`,
			},
			{
				Statement: `ALTER INDEX tbspace_reindex_part_index ATTACH PARTITION tbspace_reindex_part_index_10;`,
			},
			{
				Statement: `CREATE INDEX tbspace_reindex_part_index_0_1 ON ONLY tbspace_reindex_part_0_1 (c1);`,
			},
			{
				Statement: `ALTER INDEX tbspace_reindex_part_index_0 ATTACH PARTITION tbspace_reindex_part_index_0_1;`,
			},
			{
				Statement: `CREATE INDEX tbspace_reindex_part_index_0_2 ON ONLY tbspace_reindex_part_0_2 (c1);`,
			},
			{
				Statement: `ALTER INDEX tbspace_reindex_part_index_0 ATTACH PARTITION tbspace_reindex_part_index_0_2;`,
			},
			{
				Statement: `SELECT relid, parentrelid, level FROM pg_partition_tree('tbspace_reindex_part_index')
  ORDER BY relid, level;`,
				Results: []sql.Row{{`tbspace_reindex_part_index`, ``, 0}, {`tbspace_reindex_part_index_0`, `tbspace_reindex_part_index`, 1}, {`tbspace_reindex_part_index_10`, `tbspace_reindex_part_index`, 1}, {`tbspace_reindex_part_index_0_1`, `tbspace_reindex_part_index_0`, 2}, {`tbspace_reindex_part_index_0_2`, `tbspace_reindex_part_index_0`, 2}},
			},
			{
				Statement: `CREATE TEMP TABLE reindex_temp_before AS
  SELECT oid, relname, relfilenode, reltablespace
  FROM pg_class
    WHERE relname ~ 'tbspace_reindex_part_index';`,
			},
			{
				Statement: `REINDEX (TABLESPACE regress_tblspace, CONCURRENTLY) TABLE tbspace_reindex_part;`,
			},
			{
				Statement: `SELECT b.relname,
       CASE WHEN a.relfilenode = b.relfilenode THEN 'relfilenode is unchanged'
       ELSE 'relfilenode has changed' END AS filenode,
       CASE WHEN a.reltablespace = b.reltablespace THEN 'reltablespace is unchanged'
       ELSE 'reltablespace has changed' END AS tbspace
  FROM reindex_temp_before b JOIN pg_class a ON b.relname = a.relname
  ORDER BY 1;`,
				Results: []sql.Row{{`tbspace_reindex_part_index`, `relfilenode is unchanged`, `reltablespace is unchanged`}, {`tbspace_reindex_part_index_0`, `relfilenode is unchanged`, `reltablespace is unchanged`}, {`tbspace_reindex_part_index_0_1`, `relfilenode has changed`, `reltablespace has changed`}, {`tbspace_reindex_part_index_0_2`, `relfilenode has changed`, `reltablespace has changed`}, {`tbspace_reindex_part_index_10`, `relfilenode is unchanged`, `reltablespace is unchanged`}},
			},
			{
				Statement: `DROP TABLE tbspace_reindex_part;`,
			},
			{
				Statement: `CREATE SCHEMA testschema;`,
			},
			{
				Statement: `CREATE TABLE testschema.foo (i int) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `SELECT relname, spcname FROM pg_catalog.pg_tablespace t, pg_catalog.pg_class c
    where c.reltablespace = t.oid AND c.relname = 'foo';`,
				Results: []sql.Row{{`foo`, `regress_tblspace`}},
			},
			{
				Statement: `INSERT INTO testschema.foo VALUES(1);`,
			},
			{
				Statement: `INSERT INTO testschema.foo VALUES(2);`,
			},
			{
				Statement: `CREATE TABLE testschema.asselect TABLESPACE regress_tblspace AS SELECT 1;`,
			},
			{
				Statement: `SELECT relname, spcname FROM pg_catalog.pg_tablespace t, pg_catalog.pg_class c
    where c.reltablespace = t.oid AND c.relname = 'asselect';`,
				Results: []sql.Row{{`asselect`, `regress_tblspace`}},
			},
			{
				Statement: `PREPARE selectsource(int) AS SELECT $1;`,
			},
			{
				Statement: `CREATE TABLE testschema.asexecute TABLESPACE regress_tblspace
    AS EXECUTE selectsource(2);`,
			},
			{
				Statement: `SELECT relname, spcname FROM pg_catalog.pg_tablespace t, pg_catalog.pg_class c
    where c.reltablespace = t.oid AND c.relname = 'asexecute';`,
				Results: []sql.Row{{`asexecute`, `regress_tblspace`}},
			},
			{
				Statement: `CREATE INDEX foo_idx on testschema.foo(i) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `SELECT relname, spcname FROM pg_catalog.pg_tablespace t, pg_catalog.pg_class c
    where c.reltablespace = t.oid AND c.relname = 'foo_idx';`,
				Results: []sql.Row{{`foo_idx`, `regress_tblspace`}},
			},
			{
				Statement: `\d testschema.foo
              Table "testschema.foo"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 i      | integer |           |          | 
Indexes:
    "foo_idx" btree (i), tablespace "regress_tblspace"
Tablespace: "regress_tblspace"
\d testschema.foo_idx
      Index "testschema.foo_idx"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 i      | integer | yes  | i
btree, for table "testschema.foo"
Tablespace: "regress_tblspace"
CREATE TABLE testschema.part (a int) PARTITION BY LIST (a);`,
			},
			{
				Statement: `SET default_tablespace TO pg_global;`,
			},
			{
				Statement:   `CREATE TABLE testschema.part_1 PARTITION OF testschema.part FOR VALUES IN (1);`,
				ErrorString: `only shared relations can be placed in pg_global tablespace`,
			},
			{
				Statement: `RESET default_tablespace;`,
			},
			{
				Statement: `CREATE TABLE testschema.part_1 PARTITION OF testschema.part FOR VALUES IN (1);`,
			},
			{
				Statement: `SET default_tablespace TO regress_tblspace;`,
			},
			{
				Statement: `CREATE TABLE testschema.part_2 PARTITION OF testschema.part FOR VALUES IN (2);`,
			},
			{
				Statement: `SET default_tablespace TO pg_global;`,
			},
			{
				Statement:   `CREATE TABLE testschema.part_3 PARTITION OF testschema.part FOR VALUES IN (3);`,
				ErrorString: `only shared relations can be placed in pg_global tablespace`,
			},
			{
				Statement: `ALTER TABLE testschema.part SET TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `CREATE TABLE testschema.part_3 PARTITION OF testschema.part FOR VALUES IN (3);`,
			},
			{
				Statement: `CREATE TABLE testschema.part_4 PARTITION OF testschema.part FOR VALUES IN (4)
  TABLESPACE pg_default;`,
			},
			{
				Statement: `CREATE TABLE testschema.part_56 PARTITION OF testschema.part FOR VALUES IN (5, 6)
  PARTITION BY LIST (a);`,
			},
			{
				Statement: `ALTER TABLE testschema.part SET TABLESPACE pg_default;`,
			},
			{
				Statement: `CREATE TABLE testschema.part_78 PARTITION OF testschema.part FOR VALUES IN (7, 8)
  PARTITION BY LIST (a);`,
				ErrorString: `only shared relations can be placed in pg_global tablespace`,
			},
			{
				Statement: `CREATE TABLE testschema.part_910 PARTITION OF testschema.part FOR VALUES IN (9, 10)
  PARTITION BY LIST (a) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `RESET default_tablespace;`,
			},
			{
				Statement: `CREATE TABLE testschema.part_78 PARTITION OF testschema.part FOR VALUES IN (7, 8)
  PARTITION BY LIST (a);`,
			},
			{
				Statement: `SELECT relname, spcname FROM pg_catalog.pg_class c
    JOIN pg_catalog.pg_namespace n ON (c.relnamespace = n.oid)
    LEFT JOIN pg_catalog.pg_tablespace t ON c.reltablespace = t.oid
    where c.relname LIKE 'part%' AND n.nspname = 'testschema' order by relname;`,
				Results: []sql.Row{{`part`, ``}, {`part_1`, ``}, {`part_2`, `regress_tblspace`}, {`part_3`, `regress_tblspace`}, {`part_4`, ``}, {`part_56`, `regress_tblspace`}, {`part_78`, ``}, {`part_910`, `regress_tblspace`}},
			},
			{
				Statement: `RESET default_tablespace;`,
			},
			{
				Statement: `DROP TABLE testschema.part;`,
			},
			{
				Statement: `CREATE TABLE testschema.part (a int) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE testschema.part1 PARTITION OF testschema.part FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE INDEX part_a_idx ON testschema.part (a) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `CREATE TABLE testschema.part2 PARTITION OF testschema.part FOR VALUES IN (2);`,
			},
			{
				Statement: `SELECT relname, spcname FROM pg_catalog.pg_tablespace t, pg_catalog.pg_class c
    where c.reltablespace = t.oid AND c.relname LIKE 'part%_idx';`,
				Results: []sql.Row{{`part1_a_idx`, `regress_tblspace`}, {`part2_a_idx`, `regress_tblspace`}, {`part_a_idx`, `regress_tblspace`}},
			},
			{
				Statement: `\d testschema.part
        Partitioned table "testschema.part"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition key: LIST (a)
Indexes:
    "part_a_idx" btree (a), tablespace "regress_tblspace"
Number of partitions: 2 (Use \d+ to list them.)
\d+ testschema.part
                           Partitioned table "testschema.part"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
Partition key: LIST (a)
Indexes:
    "part_a_idx" btree (a), tablespace "regress_tblspace"
Partitions: testschema.part1 FOR VALUES IN (1),
            testschema.part2 FOR VALUES IN (2)
\d testschema.part1
             Table "testschema.part1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition of: testschema.part FOR VALUES IN (1)
Indexes:
    "part1_a_idx" btree (a), tablespace "regress_tblspace"
\d+ testschema.part1
                                 Table "testschema.part1"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
Partition of: testschema.part FOR VALUES IN (1)
Partition constraint: ((a IS NOT NULL) AND (a = 1))
Indexes:
    "part1_a_idx" btree (a), tablespace "regress_tblspace"
\d testschema.part_a_idx
Partitioned index "testschema.part_a_idx"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 a      | integer | yes  | a
btree, for table "testschema.part"
Number of partitions: 2 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
\d+ testschema.part_a_idx
           Partitioned index "testschema.part_a_idx"
 Column |  Type   | Key? | Definition | Storage | Stats target 
--------+---------+------+------------+---------+--------------
 a      | integer | yes  | a          | plain   | 
btree, for table "testschema.part"
Partitions: testschema.part1_a_idx,
            testschema.part2_a_idx
Tablespace: "regress_tblspace"
CREATE TABLE testschema.dflt (a int PRIMARY KEY) PARTITION BY LIST (a) TABLESPACE pg_default;`,
				ErrorString: `cannot specify default tablespace for partitioned relations`,
			},
			{
				Statement:   `CREATE TABLE testschema.dflt (a int PRIMARY KEY USING INDEX TABLESPACE pg_default) PARTITION BY LIST (a);`,
				ErrorString: `cannot specify default tablespace for partitioned relations`,
			},
			{
				Statement: `SET default_tablespace TO 'pg_default';`,
			},
			{
				Statement:   `CREATE TABLE testschema.dflt (a int PRIMARY KEY) PARTITION BY LIST (a) TABLESPACE regress_tblspace;`,
				ErrorString: `cannot specify default tablespace for partitioned relations`,
			},
			{
				Statement:   `CREATE TABLE testschema.dflt (a int PRIMARY KEY USING INDEX TABLESPACE regress_tblspace) PARTITION BY LIST (a);`,
				ErrorString: `cannot specify default tablespace for partitioned relations`,
			},
			{
				Statement: `CREATE TABLE testschema.dflt (a int PRIMARY KEY USING INDEX TABLESPACE regress_tblspace) PARTITION BY LIST (a) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `SET default_tablespace TO '';`,
			},
			{
				Statement: `CREATE TABLE testschema.dflt2 (a int PRIMARY KEY) PARTITION BY LIST (a);`,
			},
			{
				Statement: `DROP TABLE testschema.dflt, testschema.dflt2;`,
			},
			{
				Statement: `CREATE TABLE testschema.test_default_tab(id bigint) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `INSERT INTO testschema.test_default_tab VALUES (1);`,
			},
			{
				Statement: `CREATE INDEX test_index1 on testschema.test_default_tab (id);`,
			},
			{
				Statement: `CREATE INDEX test_index2 on testschema.test_default_tab (id) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab ADD CONSTRAINT test_index3 PRIMARY KEY (id);`,
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab ADD CONSTRAINT test_index4 UNIQUE (id) USING INDEX TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `\d testschema.test_index1
   Index "testschema.test_index1"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
btree, for table "testschema.test_default_tab"
\d testschema.test_index2
   Index "testschema.test_index2"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
\d testschema.test_index3
   Index "testschema.test_index3"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
primary key, btree, for table "testschema.test_default_tab"
\d testschema.test_index4
   Index "testschema.test_index4"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
unique, btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
SET default_tablespace TO regress_tblspace;`,
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab ALTER id TYPE bigint;`,
			},
			{
				Statement: `\d testschema.test_index1
   Index "testschema.test_index1"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
btree, for table "testschema.test_default_tab"
\d testschema.test_index2
   Index "testschema.test_index2"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
\d testschema.test_index3
   Index "testschema.test_index3"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
primary key, btree, for table "testschema.test_default_tab"
\d testschema.test_index4
   Index "testschema.test_index4"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
unique, btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
SELECT * FROM testschema.test_default_tab;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab ALTER id TYPE int;`,
			},
			{
				Statement: `\d testschema.test_index1
    Index "testschema.test_index1"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
btree, for table "testschema.test_default_tab"
\d testschema.test_index2
    Index "testschema.test_index2"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
\d testschema.test_index3
    Index "testschema.test_index3"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
primary key, btree, for table "testschema.test_default_tab"
\d testschema.test_index4
    Index "testschema.test_index4"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
unique, btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
SELECT * FROM testschema.test_default_tab;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `SET default_tablespace TO '';`,
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab ALTER id TYPE int;`,
			},
			{
				Statement: `\d testschema.test_index1
    Index "testschema.test_index1"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
btree, for table "testschema.test_default_tab"
\d testschema.test_index2
    Index "testschema.test_index2"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
\d testschema.test_index3
    Index "testschema.test_index3"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
primary key, btree, for table "testschema.test_default_tab"
\d testschema.test_index4
    Index "testschema.test_index4"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
unique, btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
ALTER TABLE testschema.test_default_tab ALTER id TYPE bigint;`,
			},
			{
				Statement: `\d testschema.test_index1
   Index "testschema.test_index1"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
btree, for table "testschema.test_default_tab"
\d testschema.test_index2
   Index "testschema.test_index2"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
\d testschema.test_index3
   Index "testschema.test_index3"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
primary key, btree, for table "testschema.test_default_tab"
\d testschema.test_index4
   Index "testschema.test_index4"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
unique, btree, for table "testschema.test_default_tab"
Tablespace: "regress_tblspace"
DROP TABLE testschema.test_default_tab;`,
			},
			{
				Statement: `CREATE TABLE testschema.test_default_tab_p(id bigint, val bigint)
    PARTITION BY LIST (id) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `CREATE TABLE testschema.test_default_tab_p1 PARTITION OF testschema.test_default_tab_p
    FOR VALUES IN (1);`,
			},
			{
				Statement: `INSERT INTO testschema.test_default_tab_p VALUES (1);`,
			},
			{
				Statement: `CREATE INDEX test_index1 on testschema.test_default_tab_p (val);`,
			},
			{
				Statement: `CREATE INDEX test_index2 on testschema.test_default_tab_p (val) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab_p ADD CONSTRAINT test_index3 PRIMARY KEY (id);`,
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab_p ADD CONSTRAINT test_index4 UNIQUE (id) USING INDEX TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `\d testschema.test_index1
Partitioned index "testschema.test_index1"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 val    | bigint | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index2
Partitioned index "testschema.test_index2"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 val    | bigint | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
\d testschema.test_index3
Partitioned index "testschema.test_index3"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
primary key, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index4
Partitioned index "testschema.test_index4"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
unique, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
SET default_tablespace TO regress_tblspace;`,
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab_p ALTER val TYPE bigint;`,
			},
			{
				Statement: `\d testschema.test_index1
Partitioned index "testschema.test_index1"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 val    | bigint | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index2
Partitioned index "testschema.test_index2"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 val    | bigint | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
\d testschema.test_index3
Partitioned index "testschema.test_index3"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
primary key, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index4
Partitioned index "testschema.test_index4"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
unique, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
SELECT * FROM testschema.test_default_tab_p;`,
				Results: []sql.Row{{1, ``}},
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab_p ALTER val TYPE int;`,
			},
			{
				Statement: `\d testschema.test_index1
Partitioned index "testschema.test_index1"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 val    | integer | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index2
Partitioned index "testschema.test_index2"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 val    | integer | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
\d testschema.test_index3
Partitioned index "testschema.test_index3"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
primary key, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index4
Partitioned index "testschema.test_index4"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
unique, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
SELECT * FROM testschema.test_default_tab_p;`,
				Results: []sql.Row{{1, ``}},
			},
			{
				Statement: `SET default_tablespace TO '';`,
			},
			{
				Statement: `ALTER TABLE testschema.test_default_tab_p ALTER val TYPE int;`,
			},
			{
				Statement: `\d testschema.test_index1
Partitioned index "testschema.test_index1"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 val    | integer | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index2
Partitioned index "testschema.test_index2"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 val    | integer | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
\d testschema.test_index3
Partitioned index "testschema.test_index3"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
primary key, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index4
Partitioned index "testschema.test_index4"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
unique, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
ALTER TABLE testschema.test_default_tab_p ALTER val TYPE bigint;`,
			},
			{
				Statement: `\d testschema.test_index1
Partitioned index "testschema.test_index1"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 val    | bigint | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index2
Partitioned index "testschema.test_index2"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 val    | bigint | yes  | val
btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
\d testschema.test_index3
Partitioned index "testschema.test_index3"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
primary key, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
\d testschema.test_index4
Partitioned index "testschema.test_index4"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 id     | bigint | yes  | id
unique, btree, for table "testschema.test_default_tab_p"
Number of partitions: 1 (Use \d+ to list them.)
Tablespace: "regress_tblspace"
DROP TABLE testschema.test_default_tab_p;`,
			},
			{
				Statement: `CREATE TABLE testschema.test_tab(id int) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `INSERT INTO testschema.test_tab VALUES (1);`,
			},
			{
				Statement: `SET default_tablespace TO regress_tblspace;`,
			},
			{
				Statement: `ALTER TABLE testschema.test_tab ADD CONSTRAINT test_tab_unique UNIQUE (id);`,
			},
			{
				Statement: `SET default_tablespace TO '';`,
			},
			{
				Statement: `ALTER TABLE testschema.test_tab ADD CONSTRAINT test_tab_pkey PRIMARY KEY (id);`,
			},
			{
				Statement: `\d testschema.test_tab_unique
  Index "testschema.test_tab_unique"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
unique, btree, for table "testschema.test_tab"
Tablespace: "regress_tblspace"
\d testschema.test_tab_pkey
   Index "testschema.test_tab_pkey"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 id     | integer | yes  | id
primary key, btree, for table "testschema.test_tab"
SELECT * FROM testschema.test_tab;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `DROP TABLE testschema.test_tab;`,
			},
			{
				Statement: `CREATE TABLE testschema.test_tab(a int, b int, c int);`,
			},
			{
				Statement: `SET default_tablespace TO regress_tblspace;`,
			},
			{
				Statement: `ALTER TABLE testschema.test_tab ADD CONSTRAINT test_tab_unique UNIQUE (a);`,
			},
			{
				Statement: `CREATE INDEX test_tab_a_idx ON testschema.test_tab (a);`,
			},
			{
				Statement: `SET default_tablespace TO '';`,
			},
			{
				Statement: `CREATE INDEX test_tab_b_idx ON testschema.test_tab (b);`,
			},
			{
				Statement: `\d testschema.test_tab_unique
  Index "testschema.test_tab_unique"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 a      | integer | yes  | a
unique, btree, for table "testschema.test_tab"
Tablespace: "regress_tblspace"
\d testschema.test_tab_a_idx
  Index "testschema.test_tab_a_idx"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 a      | integer | yes  | a
btree, for table "testschema.test_tab"
Tablespace: "regress_tblspace"
\d testschema.test_tab_b_idx
  Index "testschema.test_tab_b_idx"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 b      | integer | yes  | b
btree, for table "testschema.test_tab"
ALTER TABLE testschema.test_tab ALTER b TYPE bigint, ADD UNIQUE (c);`,
			},
			{
				Statement: `\d testschema.test_tab_unique
  Index "testschema.test_tab_unique"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 a      | integer | yes  | a
unique, btree, for table "testschema.test_tab"
Tablespace: "regress_tblspace"
\d testschema.test_tab_a_idx
  Index "testschema.test_tab_a_idx"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 a      | integer | yes  | a
btree, for table "testschema.test_tab"
Tablespace: "regress_tblspace"
\d testschema.test_tab_b_idx
  Index "testschema.test_tab_b_idx"
 Column |  Type  | Key? | Definition 
--------+--------+------+------------
 b      | bigint | yes  | b
btree, for table "testschema.test_tab"
DROP TABLE testschema.test_tab;`,
			},
			{
				Statement: `CREATE TABLE testschema.atable AS VALUES (1), (2);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX anindex ON testschema.atable(column1);`,
			},
			{
				Statement: `ALTER TABLE testschema.atable SET TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `ALTER INDEX testschema.anindex SET TABLESPACE regress_tblspace;`,
			},
			{
				Statement:   `ALTER INDEX testschema.part_a_idx SET TABLESPACE pg_global;`,
				ErrorString: `only shared relations can be placed in pg_global tablespace`,
			},
			{
				Statement: `ALTER INDEX testschema.part_a_idx SET TABLESPACE pg_default;`,
			},
			{
				Statement: `ALTER INDEX testschema.part_a_idx SET TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `INSERT INTO testschema.atable VALUES(3);	-- ok`,
			},
			{
				Statement:   `INSERT INTO testschema.atable VALUES(1);	-- fail (checks index)`,
				ErrorString: `duplicate key value violates unique constraint "anindex"`,
			},
			{
				Statement: `SELECT COUNT(*) FROM testschema.atable;		-- checks heap`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `CREATE MATERIALIZED VIEW testschema.amv AS SELECT * FROM testschema.atable;`,
			},
			{
				Statement: `ALTER MATERIALIZED VIEW testschema.amv SET TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW testschema.amv;`,
			},
			{
				Statement: `SELECT COUNT(*) FROM testschema.amv;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement:   `CREATE TABLESPACE regress_badspace LOCATION '/no/such/location';`,
				ErrorString: `directory "/no/such/location" does not exist`,
			},
			{
				Statement:   `CREATE TABLE bar (i int) TABLESPACE regress_nosuchspace;`,
				ErrorString: `tablespace "regress_nosuchspace" does not exist`,
			},
			{
				Statement:   `DROP TABLESPACE regress_tblspace;`,
				ErrorString: `tablespace "regress_tblspace" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `ALTER INDEX testschema.part_a_idx SET TABLESPACE pg_default;`,
			},
			{
				Statement:   `DROP TABLESPACE regress_tblspace;`,
				ErrorString: `tablespace "regress_tblspace" is not empty`,
			},
			{
				Statement: `CREATE ROLE regress_tablespace_user1 login;`,
			},
			{
				Statement: `CREATE ROLE regress_tablespace_user2 login;`,
			},
			{
				Statement: `GRANT USAGE ON SCHEMA testschema TO regress_tablespace_user2;`,
			},
			{
				Statement: `ALTER TABLESPACE regress_tblspace OWNER TO regress_tablespace_user1;`,
			},
			{
				Statement: `CREATE TABLE testschema.tablespace_acl (c int);`,
			},
			{
				Statement: `CREATE INDEX k ON testschema.tablespace_acl (c) TABLESPACE regress_tblspace;`,
			},
			{
				Statement: `ALTER TABLE testschema.tablespace_acl OWNER TO regress_tablespace_user2;`,
			},
			{
				Statement: `SET SESSION ROLE regress_tablespace_user2;`,
			},
			{
				Statement:   `CREATE TABLE tablespace_table (i int) TABLESPACE regress_tblspace; -- fail`,
				ErrorString: `permission denied for tablespace regress_tblspace`,
			},
			{
				Statement: `ALTER TABLE testschema.tablespace_acl ALTER c TYPE bigint;`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace) TABLE tablespace_table; -- fail`,
				ErrorString: `permission denied for tablespace regress_tblspace`,
			},
			{
				Statement:   `REINDEX (TABLESPACE regress_tblspace, CONCURRENTLY) TABLE tablespace_table; -- fail`,
				ErrorString: `permission denied for tablespace regress_tblspace`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `ALTER TABLESPACE regress_tblspace RENAME TO regress_tblspace_renamed;`,
			},
			{
				Statement: `ALTER TABLE ALL IN TABLESPACE regress_tblspace_renamed SET TABLESPACE pg_default;`,
			},
			{
				Statement: `ALTER INDEX ALL IN TABLESPACE regress_tblspace_renamed SET TABLESPACE pg_default;`,
			},
			{
				Statement: `ALTER MATERIALIZED VIEW ALL IN TABLESPACE regress_tblspace_renamed SET TABLESPACE pg_default;`,
			},
			{
				Statement: `ALTER TABLE ALL IN TABLESPACE regress_tblspace_renamed SET TABLESPACE pg_default;`,
			},
			{
				Statement: `ALTER MATERIALIZED VIEW ALL IN TABLESPACE regress_tblspace_renamed SET TABLESPACE pg_default;`,
			},
			{
				Statement: `DROP TABLESPACE regress_tblspace_renamed;`,
			},
			{
				Statement: `DROP SCHEMA testschema CASCADE;`,
			},
			{
				Statement: `DROP ROLE regress_tablespace_user1;`,
			},
			{
				Statement: `DROP ROLE regress_tablespace_user2;`,
			},
		},
	})
}
