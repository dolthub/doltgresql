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

func TestCreateTable(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_table)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_table,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE unknowntab (
	u unknown    -- fail
);`,
				ErrorString: `column "u" has pseudo-type unknown`,
			},
			{
				Statement: `CREATE TYPE unknown_comptype AS (
	u unknown    -- fail
);`,
				ErrorString: `column "u" has pseudo-type unknown`,
			},
			{
				Statement:   `CREATE TABLE tas_case WITH ("Fillfactor" = 10) AS SELECT 1 a;`,
				ErrorString: `unrecognized parameter "Fillfactor"`,
			},
			{
				Statement: `CREATE UNLOGGED TABLE unlogged1 (a int primary key);			-- OK`,
			},
			{
				Statement: `CREATE TEMPORARY TABLE unlogged2 (a int primary key);			-- OK`,
			},
			{
				Statement: `SELECT relname, relkind, relpersistence FROM pg_class WHERE relname ~ '^unlogged\d' ORDER BY relname;`,
				Results:   []sql.Row{{`unlogged1`, `r`, `u`}, {`unlogged1_pkey`, `i`, `u`}, {`unlogged2`, `r`, true}, {`unlogged2_pkey`, `i`, true}},
			},
			{
				Statement: `REINDEX INDEX unlogged1_pkey;`,
			},
			{
				Statement: `REINDEX INDEX unlogged2_pkey;`,
			},
			{
				Statement: `SELECT relname, relkind, relpersistence FROM pg_class WHERE relname ~ '^unlogged\d' ORDER BY relname;`,
				Results:   []sql.Row{{`unlogged1`, `r`, `u`}, {`unlogged1_pkey`, `i`, `u`}, {`unlogged2`, `r`, true}, {`unlogged2_pkey`, `i`, true}},
			},
			{
				Statement: `DROP TABLE unlogged2;`,
			},
			{
				Statement: `INSERT INTO unlogged1 VALUES (42);`,
			},
			{
				Statement: `CREATE UNLOGGED TABLE public.unlogged2 (a int primary key);		-- also OK`,
			},
			{
				Statement:   `CREATE UNLOGGED TABLE pg_temp.unlogged3 (a int primary key);	-- not OK`,
				ErrorString: `only temporary relations may be created in temporary schemas`,
			},
			{
				Statement: `CREATE TABLE pg_temp.implicitly_temp (a int primary key);		-- OK`,
			},
			{
				Statement: `CREATE TEMP TABLE explicitly_temp (a int primary key);			-- also OK`,
			},
			{
				Statement: `CREATE TEMP TABLE pg_temp.doubly_temp (a int primary key);		-- also OK`,
			},
			{
				Statement:   `CREATE TEMP TABLE public.temp_to_perm (a int primary key);		-- not OK`,
				ErrorString: `cannot create temporary relation in non-temporary schema`,
			},
			{
				Statement: `DROP TABLE unlogged1, public.unlogged2;`,
			},
			{
				Statement: `CREATE TABLE as_select1 AS SELECT * FROM pg_class WHERE relkind = 'r';`,
			},
			{
				Statement:   `CREATE TABLE as_select1 AS SELECT * FROM pg_class WHERE relkind = 'r';`,
				ErrorString: `relation "as_select1" already exists`,
			},
			{
				Statement: `CREATE TABLE IF NOT EXISTS as_select1 AS SELECT * FROM pg_class WHERE relkind = 'r';`,
			},
			{
				Statement: `DROP TABLE as_select1;`,
			},
			{
				Statement: `PREPARE select1 AS SELECT 1 as a;`,
			},
			{
				Statement: `CREATE TABLE as_select1 AS EXECUTE select1;`,
			},
			{
				Statement:   `CREATE TABLE as_select1 AS EXECUTE select1;`,
				ErrorString: `relation "as_select1" already exists`,
			},
			{
				Statement: `SELECT * FROM as_select1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `CREATE TABLE IF NOT EXISTS as_select1 AS EXECUTE select1;`,
			},
			{
				Statement: `DROP TABLE as_select1;`,
			},
			{
				Statement: `DEALLOCATE select1;`,
			},
			{
				Statement: `\set ECHO none
INSERT INTO extra_wide_table(firstc, lastc) VALUES('first col', 'last col');`,
			},
			{
				Statement: `SELECT firstc, lastc FROM extra_wide_table;`,
				Results:   []sql.Row{{`first col`, `last col`}},
			},
			{
				Statement:   `CREATE TABLE withoid() WITH OIDS;`,
				ErrorString: `syntax error at or near "OIDS"`,
			},
			{
				Statement:   `CREATE TABLE withoid() WITH (oids);`,
				ErrorString: `tables declared WITH OIDS are not supported`,
			},
			{
				Statement:   `CREATE TABLE withoid() WITH (oids = true);`,
				ErrorString: `tables declared WITH OIDS are not supported`,
			},
			{
				Statement: `CREATE TEMP TABLE withoutoid() WITHOUT OIDS; DROP TABLE withoutoid;`,
			},
			{
				Statement: `CREATE TEMP TABLE withoutoid() WITH (oids = false); DROP TABLE withoutoid;`,
			},
			{
				Statement:   `CREATE TABLE default_expr_column (id int DEFAULT (id));`,
				ErrorString: `cannot use column reference in DEFAULT expression`,
			},
			{
				Statement:   `CREATE TABLE default_expr_column (id int DEFAULT (bar.id));`,
				ErrorString: `cannot use column reference in DEFAULT expression`,
			},
			{
				Statement:   `CREATE TABLE default_expr_agg_column (id int DEFAULT (avg(id)));`,
				ErrorString: `cannot use column reference in DEFAULT expression`,
			},
			{
				Statement:   `CREATE TABLE default_expr_non_column (a int DEFAULT (avg(non_existent)));`,
				ErrorString: `cannot use column reference in DEFAULT expression`,
			},
			{
				Statement:   `CREATE TABLE default_expr_agg (a int DEFAULT (avg(1)));`,
				ErrorString: `aggregate functions are not allowed in DEFAULT expressions`,
			},
			{
				Statement:   `CREATE TABLE default_expr_agg (a int DEFAULT (select 1));`,
				ErrorString: `cannot use subquery in DEFAULT expression`,
			},
			{
				Statement:   `CREATE TABLE default_expr_agg (a int DEFAULT (generate_series(1,3)));`,
				ErrorString: `set-returning functions are not allowed in DEFAULT expressions`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TABLE remember_create_subid (c int);`,
			},
			{
				Statement: `SAVEPOINT q; DROP TABLE remember_create_subid; ROLLBACK TO q;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP TABLE remember_create_subid;`,
			},
			{
				Statement: `CREATE TABLE remember_node_subid (c int);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `ALTER TABLE remember_node_subid ALTER c TYPE bigint;`,
			},
			{
				Statement: `SAVEPOINT q; DROP TABLE remember_node_subid; ROLLBACK TO q;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP TABLE remember_node_subid;`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) INHERITS (some_table) PARTITION BY LIST (a);`,
				ErrorString: `cannot create partitioned table as inheritance child`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a1 int,
	a2 int
) PARTITION BY LIST (a1, a2);	-- fail`,
				ErrorString: `cannot use "list" partition strategy with more than one column`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int,
	EXCLUDE USING gist (a WITH &&)
) PARTITION BY RANGE (a);`,
				ErrorString: `exclusion constraints are not supported on partitioned tables`,
			},
			{
				Statement: `CREATE FUNCTION retset (a int) RETURNS SETOF int AS $$ SELECT 1; $$ LANGUAGE SQL IMMUTABLE;`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) PARTITION BY RANGE (retset(a));`,
				ErrorString: `set-returning functions are not allowed in partition key expressions`,
			},
			{
				Statement: `DROP FUNCTION retset(int);`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) PARTITION BY RANGE ((avg(a)));`,
				ErrorString: `aggregate functions are not allowed in partition key expressions`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int,
	b int
) PARTITION BY RANGE ((avg(a) OVER (PARTITION BY b)));`,
				ErrorString: `window functions are not allowed in partition key expressions`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) PARTITION BY LIST ((a LIKE (SELECT 1)));`,
				ErrorString: `cannot use subquery in partition key expression`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) PARTITION BY RANGE ((42));`,
				ErrorString: `cannot use constant expression as partition key`,
			},
			{
				Statement: `CREATE FUNCTION const_func () RETURNS int AS $$ SELECT 1; $$ LANGUAGE SQL IMMUTABLE;`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) PARTITION BY RANGE (const_func());`,
				ErrorString: `cannot use constant expression as partition key`,
			},
			{
				Statement: `DROP FUNCTION const_func();`,
			},
			{
				Statement: `CREATE TABLE partitioned (
    a int
) PARTITION BY MAGIC (a);`,
				ErrorString: `unrecognized partitioning strategy "magic"`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) PARTITION BY RANGE (b);`,
				ErrorString: `column "b" named in partition key does not exist`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) PARTITION BY RANGE (xmin);`,
				ErrorString: `cannot use system column "xmin" in partition key`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int,
	b int
) PARTITION BY RANGE (((a, b)));`,
				ErrorString: `partition key column 1 has pseudo-type record`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int,
	b int
) PARTITION BY RANGE (a, ('unknown'));`,
				ErrorString: `partition key column 2 has pseudo-type unknown`,
			},
			{
				Statement: `CREATE FUNCTION immut_func (a int) RETURNS int AS $$ SELECT a + random()::int; $$ LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int
) PARTITION BY RANGE (immut_func(a));`,
				ErrorString: `functions in partition key expression must be marked IMMUTABLE`,
			},
			{
				Statement: `DROP FUNCTION immut_func(int);`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a point
) PARTITION BY LIST (a);`,
				ErrorString: `data type point has no default operator class for access method "btree"`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a point
) PARTITION BY LIST (a point_ops);`,
				ErrorString: `operator class "point_ops" does not exist for access method "btree"`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a point
) PARTITION BY RANGE (a);`,
				ErrorString: `data type point has no default operator class for access method "btree"`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a point
) PARTITION BY RANGE (a point_ops);`,
				ErrorString: `operator class "point_ops" does not exist for access method "btree"`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int,
	CONSTRAINT check_a CHECK (a > 0) NO INHERIT
) PARTITION BY RANGE (a);`,
				ErrorString: `cannot add NO INHERIT constraint to partitioned table "partitioned"`,
			},
			{
				Statement: `CREATE FUNCTION plusone(a int) RETURNS INT AS $$ SELECT a+1; $$ LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE TABLE partitioned (
	a int,
	b int,
	c text,
	d text
) PARTITION BY RANGE (a oid_ops, plusone(b), c collate "default", d collate "C");`,
			},
			{
				Statement: `SELECT relkind FROM pg_class WHERE relname = 'partitioned';`,
				Results:   []sql.Row{{`p`}},
			},
			{
				Statement:   `DROP FUNCTION plusone(int);`,
				ErrorString: `cannot drop function plusone(integer) because other objects depend on it`,
			},
			{
				Statement: `CREATE TABLE partitioned2 (
	a int,
	b text
) PARTITION BY RANGE ((a+1), substr(b, 1, 5));`,
			},
			{
				Statement:   `CREATE TABLE fail () INHERITS (partitioned2);`,
				ErrorString: `cannot inherit from partitioned table "partitioned2"`,
			},
			{
				Statement: `\d partitioned
      Partitioned table "public.partitioned"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
 c      | text    |           |          | 
 d      | text    |           |          | 
Partition key: RANGE (a oid_ops, plusone(b), c, d COLLATE "C")
Number of partitions: 0
\d+ partitioned2
                          Partitioned table "public.partitioned2"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | integer |           |          |         | plain    |              | 
 b      | text    |           |          |         | extended |              | 
Partition key: RANGE (((a + 1)), substr(b, 1, 5))
Number of partitions: 0
INSERT INTO partitioned2 VALUES (1, 'hello');`,
				ErrorString: `no partition of relation "partitioned2" found for row`,
			},
			{
				Statement: `CREATE TABLE part2_1 PARTITION OF partitioned2 FOR VALUES FROM (-1, 'aaaaa') TO (100, 'ccccc');`,
			},
			{
				Statement: `\d+ part2_1
                                  Table "public.part2_1"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | integer |           |          |         | plain    |              | 
 b      | text    |           |          |         | extended |              | 
Partition of: partitioned2 FOR VALUES FROM ('-1', 'aaaaa') TO (100, 'ccccc')
Partition constraint: (((a + 1) IS NOT NULL) AND (substr(b, 1, 5) IS NOT NULL) AND (((a + 1) > '-1'::integer) OR (((a + 1) = '-1'::integer) AND (substr(b, 1, 5) >= 'aaaaa'::text))) AND (((a + 1) < 100) OR (((a + 1) = 100) AND (substr(b, 1, 5) < 'ccccc'::text))))
DROP TABLE partitioned, partitioned2;`,
			},
			{
				Statement: `create table partitioned (a int, b int)
  partition by list ((row(a, b)::partitioned));`,
			},
			{
				Statement: `create table partitioned1
  partition of partitioned for values in ('(1,2)'::partitioned);`,
			},
			{
				Statement: `create table partitioned2
  partition of partitioned for values in ('(2,4)'::partitioned);`,
			},
			{
				Statement: `explain (costs off)
select * from partitioned where row(a,b)::partitioned = '(1,2)'::partitioned;`,
				Results: []sql.Row{{`Seq Scan on partitioned1 partitioned`}, {`Filter: (ROW(a, b)::partitioned = '(1,2)'::partitioned)`}},
			},
			{
				Statement: `drop table partitioned;`,
			},
			{
				Statement: `create table partitioned (a int, b int)
  partition by list ((partitioned));`,
			},
			{
				Statement: `create table partitioned1
  partition of partitioned for values in ('(1,2)');`,
			},
			{
				Statement: `create table partitioned2
  partition of partitioned for values in ('(2,4)');`,
			},
			{
				Statement: `explain (costs off)
select * from partitioned where partitioned = '(1,2)'::partitioned;`,
				Results: []sql.Row{{`Seq Scan on partitioned1 partitioned`}, {`Filter: ((partitioned.*)::partitioned = '(1,2)'::partitioned)`}},
			},
			{
				Statement: `\d+ partitioned1
                               Table "public.partitioned1"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
Partition of: partitioned FOR VALUES IN ('(1,2)')
Partition constraint: (((partitioned1.*)::partitioned IS DISTINCT FROM NULL) AND ((partitioned1.*)::partitioned = '(1,2)'::partitioned))
drop table partitioned;`,
			},
			{
				Statement: `create domain intdom1 as int;`,
			},
			{
				Statement: `create table partitioned (
	a intdom1,
	b text
) partition by range (a);`,
			},
			{
				Statement:   `alter table partitioned drop column a;  -- fail`,
				ErrorString: `cannot drop column "a" because it is part of the partition key of relation "partitioned"`,
			},
			{
				Statement:   `drop domain intdom1;  -- fail, requires cascade`,
				ErrorString: `cannot drop type intdom1 because other objects depend on it`,
			},
			{
				Statement: `drop domain intdom1 cascade;`,
			},
			{
				Statement:   `table partitioned;  -- gone`,
				ErrorString: `relation "partitioned" does not exist`,
			},
			{
				Statement: `create domain intdom1 as int;`,
			},
			{
				Statement: `create table partitioned (
	a intdom1,
	b text
) partition by range (plusone(a));`,
			},
			{
				Statement:   `alter table partitioned drop column a;  -- fail`,
				ErrorString: `cannot drop column "a" because it is part of the partition key of relation "partitioned"`,
			},
			{
				Statement:   `drop domain intdom1;  -- fail, requires cascade`,
				ErrorString: `cannot drop type intdom1 because other objects depend on it`,
			},
			{
				Statement: `drop domain intdom1 cascade;`,
			},
			{
				Statement:   `table partitioned;  -- gone`,
				ErrorString: `relation "partitioned" does not exist`,
			},
			{
				Statement: `CREATE TABLE list_parted (
	a int
) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE part_p1 PARTITION OF list_parted FOR VALUES IN ('1');`,
			},
			{
				Statement: `CREATE TABLE part_p2 PARTITION OF list_parted FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE TABLE part_p3 PARTITION OF list_parted FOR VALUES IN ((2+1));`,
			},
			{
				Statement: `CREATE TABLE part_null PARTITION OF list_parted FOR VALUES IN (null);`,
			},
			{
				Statement: `\d+ list_parted
                          Partitioned table "public.list_parted"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
Partition key: LIST (a)
Partitions: part_null FOR VALUES IN (NULL),
            part_p1 FOR VALUES IN (1),
            part_p2 FOR VALUES IN (2),
            part_p3 FOR VALUES IN (3)
CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN (somename);`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement:   `CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN (somename.somename);`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement:   `CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN (a);`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement:   `CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN (sum(a));`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement:   `CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN (sum(somename));`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement:   `CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN (sum(1));`,
				ErrorString: `aggregate functions are not allowed in partition bound`,
			},
			{
				Statement:   `CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN ((select 1));`,
				ErrorString: `cannot use subquery in partition bound`,
			},
			{
				Statement:   `CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN (generate_series(4, 6));`,
				ErrorString: `set-returning functions are not allowed in partition bound`,
			},
			{
				Statement:   `CREATE TABLE part_bogus_expr_fail PARTITION OF list_parted FOR VALUES IN ((1+1) collate "POSIX");`,
				ErrorString: `collations are not supported by type integer`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF list_parted FOR VALUES IN ();`,
				ErrorString: `syntax error at or near ")"`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF list_parted FOR VALUES FROM (1) TO (2);`,
				ErrorString: `invalid bound specification for a list partition`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF list_parted FOR VALUES WITH (MODULUS 10, REMAINDER 1);`,
				ErrorString: `invalid bound specification for a list partition`,
			},
			{
				Statement: `CREATE TABLE part_default PARTITION OF list_parted DEFAULT;`,
			},
			{
				Statement:   `CREATE TABLE fail_default_part PARTITION OF list_parted DEFAULT;`,
				ErrorString: `partition "fail_default_part" conflicts with existing default partition "part_default"`,
			},
			{
				Statement: `CREATE TABLE bools (
	a bool
) PARTITION BY LIST (a);`,
			},
			{
				Statement:   `CREATE TABLE bools_true PARTITION OF bools FOR VALUES IN (1);`,
				ErrorString: `specified value cannot be cast to type boolean for column "a"`,
			},
			{
				Statement: `DROP TABLE bools;`,
			},
			{
				Statement: `CREATE TABLE moneyp (
	a money
) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE moneyp_10 PARTITION OF moneyp FOR VALUES IN (10);`,
			},
			{
				Statement: `CREATE TABLE moneyp_11 PARTITION OF moneyp FOR VALUES IN ('11');`,
			},
			{
				Statement: `CREATE TABLE moneyp_12 PARTITION OF moneyp FOR VALUES IN (to_char(12, '99')::int);`,
			},
			{
				Statement: `DROP TABLE moneyp;`,
			},
			{
				Statement: `CREATE TABLE bigintp (
	a bigint
) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE bigintp_10 PARTITION OF bigintp FOR VALUES IN (10);`,
			},
			{
				Statement:   `CREATE TABLE bigintp_10_2 PARTITION OF bigintp FOR VALUES IN ('10');`,
				ErrorString: `partition "bigintp_10_2" would overlap partition "bigintp_10"`,
			},
			{
				Statement: `DROP TABLE bigintp;`,
			},
			{
				Statement: `CREATE TABLE range_parted (
	a date
) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE part_bogus_expr_fail PARTITION OF range_parted
  FOR VALUES FROM (somename) TO ('2019-01-01');`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement: `CREATE TABLE part_bogus_expr_fail PARTITION OF range_parted
  FOR VALUES FROM (somename.somename) TO ('2019-01-01');`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement: `CREATE TABLE part_bogus_expr_fail PARTITION OF range_parted
  FOR VALUES FROM (a) TO ('2019-01-01');`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement: `CREATE TABLE part_bogus_expr_fail PARTITION OF range_parted
  FOR VALUES FROM (max(a)) TO ('2019-01-01');`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement: `CREATE TABLE part_bogus_expr_fail PARTITION OF range_parted
  FOR VALUES FROM (max(somename)) TO ('2019-01-01');`,
				ErrorString: `cannot use column reference in partition bound expression`,
			},
			{
				Statement: `CREATE TABLE part_bogus_expr_fail PARTITION OF range_parted
  FOR VALUES FROM (max('2019-02-01'::date)) TO ('2019-01-01');`,
				ErrorString: `aggregate functions are not allowed in partition bound`,
			},
			{
				Statement: `CREATE TABLE part_bogus_expr_fail PARTITION OF range_parted
  FOR VALUES FROM ((select 1)) TO ('2019-01-01');`,
				ErrorString: `cannot use subquery in partition bound`,
			},
			{
				Statement: `CREATE TABLE part_bogus_expr_fail PARTITION OF range_parted
  FOR VALUES FROM (generate_series(1, 3)) TO ('2019-01-01');`,
				ErrorString: `set-returning functions are not allowed in partition bound`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted FOR VALUES IN ('a');`,
				ErrorString: `invalid bound specification for a range partition`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted FOR VALUES WITH (MODULUS 10, REMAINDER 1);`,
				ErrorString: `invalid bound specification for a range partition`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted FOR VALUES FROM ('a', 1) TO ('z');`,
				ErrorString: `FROM must specify exactly one value per partitioning column`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted FOR VALUES FROM ('a') TO ('z', 1);`,
				ErrorString: `TO must specify exactly one value per partitioning column`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted FOR VALUES FROM (null) TO (maxvalue);`,
				ErrorString: `cannot specify NULL in range bound`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted FOR VALUES WITH (MODULUS 10, REMAINDER 1);`,
				ErrorString: `invalid bound specification for a range partition`,
			},
			{
				Statement: `CREATE TABLE hash_parted (
	a int
) PARTITION BY HASH (a);`,
			},
			{
				Statement: `CREATE TABLE hpart_1 PARTITION OF hash_parted FOR VALUES WITH (MODULUS 10, REMAINDER 0);`,
			},
			{
				Statement: `CREATE TABLE hpart_2 PARTITION OF hash_parted FOR VALUES WITH (MODULUS 50, REMAINDER 1);`,
			},
			{
				Statement: `CREATE TABLE hpart_3 PARTITION OF hash_parted FOR VALUES WITH (MODULUS 200, REMAINDER 2);`,
			},
			{
				Statement: `CREATE TABLE hpart_4 PARTITION OF hash_parted FOR VALUES WITH (MODULUS 10, REMAINDER 3);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF hash_parted FOR VALUES WITH (MODULUS 25, REMAINDER 3);`,
				ErrorString: `every hash partition modulus must be a factor of the next larger modulus`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF hash_parted FOR VALUES WITH (MODULUS 150, REMAINDER 3);`,
				ErrorString: `every hash partition modulus must be a factor of the next larger modulus`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF hash_parted FOR VALUES WITH (MODULUS 100, REMAINDER 3);`,
				ErrorString: `partition "fail_part" would overlap partition "hpart_4"`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF hash_parted FOR VALUES FROM ('a', 1) TO ('z');`,
				ErrorString: `invalid bound specification for a hash partition`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF hash_parted FOR VALUES IN (1000);`,
				ErrorString: `invalid bound specification for a hash partition`,
			},
			{
				Statement:   `CREATE TABLE fail_default_part PARTITION OF hash_parted DEFAULT;`,
				ErrorString: `a hash-partitioned table may not have a default partition`,
			},
			{
				Statement: `CREATE TABLE unparted (
	a int
);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF unparted FOR VALUES IN ('a');`,
				ErrorString: `"unparted" is not partitioned`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF unparted FOR VALUES WITH (MODULUS 2, REMAINDER 1);`,
				ErrorString: `"unparted" is not partitioned`,
			},
			{
				Statement: `DROP TABLE unparted;`,
			},
			{
				Statement: `CREATE TEMP TABLE temp_parted (
	a int
) PARTITION BY LIST (a);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF temp_parted FOR VALUES IN ('a');`,
				ErrorString: `cannot create a permanent relation as partition of temporary relation "temp_parted"`,
			},
			{
				Statement: `DROP TABLE temp_parted;`,
			},
			{
				Statement: `CREATE TABLE list_parted2 (
	a varchar
) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE part_null_z PARTITION OF list_parted2 FOR VALUES IN (null, 'z');`,
			},
			{
				Statement: `CREATE TABLE part_ab PARTITION OF list_parted2 FOR VALUES IN ('a', 'b');`,
			},
			{
				Statement: `CREATE TABLE list_parted2_def PARTITION OF list_parted2 DEFAULT;`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF list_parted2 FOR VALUES IN (null);`,
				ErrorString: `partition "fail_part" would overlap partition "part_null_z"`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF list_parted2 FOR VALUES IN ('b', 'c');`,
				ErrorString: `partition "fail_part" would overlap partition "part_ab"`,
			},
			{
				Statement: `INSERT INTO list_parted2 VALUES('X');`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF list_parted2 FOR VALUES IN ('W', 'X', 'Y');`,
				ErrorString: `updated partition constraint for default partition "list_parted2_def" would be violated by some row`,
			},
			{
				Statement: `CREATE TABLE range_parted2 (
	a int
) PARTITION BY RANGE (a);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted2 FOR VALUES FROM (1) TO (0);`,
				ErrorString: `empty range bound specified for partition "fail_part"`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted2 FOR VALUES FROM (1) TO (1);`,
				ErrorString: `empty range bound specified for partition "fail_part"`,
			},
			{
				Statement: `CREATE TABLE part0 PARTITION OF range_parted2 FOR VALUES FROM (minvalue) TO (1);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted2 FOR VALUES FROM (minvalue) TO (2);`,
				ErrorString: `partition "fail_part" would overlap partition "part0"`,
			},
			{
				Statement: `CREATE TABLE part1 PARTITION OF range_parted2 FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted2 FOR VALUES FROM (-1) TO (1);`,
				ErrorString: `partition "fail_part" would overlap partition "part0"`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted2 FOR VALUES FROM (9) TO (maxvalue);`,
				ErrorString: `partition "fail_part" would overlap partition "part1"`,
			},
			{
				Statement: `CREATE TABLE part2 PARTITION OF range_parted2 FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `CREATE TABLE part3 PARTITION OF range_parted2 FOR VALUES FROM (30) TO (40);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted2 FOR VALUES FROM (10) TO (30);`,
				ErrorString: `partition "fail_part" would overlap partition "part2"`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted2 FOR VALUES FROM (10) TO (50);`,
				ErrorString: `partition "fail_part" would overlap partition "part2"`,
			},
			{
				Statement: `CREATE TABLE range2_default PARTITION OF range_parted2 DEFAULT;`,
			},
			{
				Statement:   `CREATE TABLE fail_default_part PARTITION OF range_parted2 DEFAULT;`,
				ErrorString: `partition "fail_default_part" conflicts with existing default partition "range2_default"`,
			},
			{
				Statement: `INSERT INTO range_parted2 VALUES (85);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted2 FOR VALUES FROM (80) TO (90);`,
				ErrorString: `updated partition constraint for default partition "range2_default" would be violated by some row`,
			},
			{
				Statement: `CREATE TABLE part4 PARTITION OF range_parted2 FOR VALUES FROM (90) TO (100);`,
			},
			{
				Statement: `CREATE TABLE range_parted3 (
	a int,
	b int
) PARTITION BY RANGE (a, (b+1));`,
			},
			{
				Statement: `CREATE TABLE part00 PARTITION OF range_parted3 FOR VALUES FROM (0, minvalue) TO (0, maxvalue);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted3 FOR VALUES FROM (0, minvalue) TO (0, 1);`,
				ErrorString: `partition "fail_part" would overlap partition "part00"`,
			},
			{
				Statement: `CREATE TABLE part10 PARTITION OF range_parted3 FOR VALUES FROM (1, minvalue) TO (1, 1);`,
			},
			{
				Statement: `CREATE TABLE part11 PARTITION OF range_parted3 FOR VALUES FROM (1, 1) TO (1, 10);`,
			},
			{
				Statement: `CREATE TABLE part12 PARTITION OF range_parted3 FOR VALUES FROM (1, 10) TO (1, maxvalue);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted3 FOR VALUES FROM (1, 10) TO (1, 20);`,
				ErrorString: `partition "fail_part" would overlap partition "part12"`,
			},
			{
				Statement: `CREATE TABLE range3_default PARTITION OF range_parted3 DEFAULT;`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF range_parted3 FOR VALUES FROM (1, minvalue) TO (1, maxvalue);`,
				ErrorString: `partition "fail_part" would overlap partition "part10"`,
			},
			{
				Statement: `CREATE TABLE hash_parted2 (
	a varchar
) PARTITION BY HASH (a);`,
			},
			{
				Statement: `CREATE TABLE h2part_1 PARTITION OF hash_parted2 FOR VALUES WITH (MODULUS 4, REMAINDER 2);`,
			},
			{
				Statement: `CREATE TABLE h2part_2 PARTITION OF hash_parted2 FOR VALUES WITH (MODULUS 8, REMAINDER 0);`,
			},
			{
				Statement: `CREATE TABLE h2part_3 PARTITION OF hash_parted2 FOR VALUES WITH (MODULUS 8, REMAINDER 4);`,
			},
			{
				Statement: `CREATE TABLE h2part_4 PARTITION OF hash_parted2 FOR VALUES WITH (MODULUS 8, REMAINDER 5);`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF hash_parted2 FOR VALUES WITH (MODULUS 2, REMAINDER 1);`,
				ErrorString: `partition "fail_part" would overlap partition "h2part_4"`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF hash_parted2 FOR VALUES WITH (MODULUS 0, REMAINDER 1);`,
				ErrorString: `modulus for hash partition must be an integer value greater than zero`,
			},
			{
				Statement:   `CREATE TABLE fail_part PARTITION OF hash_parted2 FOR VALUES WITH (MODULUS 8, REMAINDER 8);`,
				ErrorString: `remainder for hash partition must be less than modulus`,
			},
			{
				Statement: `CREATE TABLE parted (
	a text,
	b int NOT NULL DEFAULT 0,
	CONSTRAINT check_a CHECK (length(a) > 0)
) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE part_a PARTITION OF parted FOR VALUES IN ('a');`,
			},
			{
				Statement: `SELECT attname, attislocal, attinhcount FROM pg_attribute
  WHERE attrelid = 'part_a'::regclass and attnum > 0
  ORDER BY attnum;`,
				Results: []sql.Row{{`a`, false, 1}, {`b`, false, 1}},
			},
			{
				Statement: `CREATE TABLE part_b PARTITION OF parted (
	b NOT NULL,
	b DEFAULT 1,
	b CHECK (b >= 0),
	CONSTRAINT check_a CHECK (length(a) > 0)
) FOR VALUES IN ('b');`,
				ErrorString: `column "b" specified more than once`,
			},
			{
				Statement: `CREATE TABLE part_b PARTITION OF parted (
	b NOT NULL DEFAULT 1,
	CONSTRAINT check_a CHECK (length(a) > 0),
	CONSTRAINT check_b CHECK (b >= 0)
) FOR VALUES IN ('b');`,
			},
			{
				Statement: `SELECT conislocal, coninhcount FROM pg_constraint WHERE conrelid = 'part_b'::regclass ORDER BY conislocal, coninhcount;`,
				Results:   []sql.Row{{false, 1}, {true, 0}},
			},
			{
				Statement: `ALTER TABLE parted ADD CONSTRAINT check_b CHECK (b >= 0);`,
			},
			{
				Statement: `SELECT conislocal, coninhcount FROM pg_constraint WHERE conrelid = 'part_b'::regclass;`,
				Results:   []sql.Row{{false, 1}, {false, 1}},
			},
			{
				Statement:   `ALTER TABLE part_b DROP CONSTRAINT check_a;`,
				ErrorString: `cannot drop inherited constraint "check_a" of relation "part_b"`,
			},
			{
				Statement:   `ALTER TABLE part_b DROP CONSTRAINT check_b;`,
				ErrorString: `cannot drop inherited constraint "check_b" of relation "part_b"`,
			},
			{
				Statement: `ALTER TABLE parted DROP CONSTRAINT check_a, DROP CONSTRAINT check_b;`,
			},
			{
				Statement: `SELECT conislocal, coninhcount FROM pg_constraint WHERE conrelid = 'part_b'::regclass;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `CREATE TABLE fail_part_col_not_found PARTITION OF parted FOR VALUES IN ('c') PARTITION BY RANGE (c);`,
				ErrorString: `column "c" named in partition key does not exist`,
			},
			{
				Statement: `CREATE TABLE part_c PARTITION OF parted (b WITH OPTIONS NOT NULL DEFAULT 0) FOR VALUES IN ('c') PARTITION BY RANGE ((b));`,
			},
			{
				Statement: `CREATE TABLE part_c_1_10 PARTITION OF part_c FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `create table parted_notnull_inh_test (a int default 1, b int not null default 0) partition by list (a);`,
			},
			{
				Statement: `create table parted_notnull_inh_test1 partition of parted_notnull_inh_test (a not null, b default 1) for values in (1);`,
			},
			{
				Statement:   `insert into parted_notnull_inh_test (b) values (null);`,
				ErrorString: `null value in column "b" of relation "parted_notnull_inh_test1" violates not-null constraint`,
			},
			{
				Statement: `\d parted_notnull_inh_test1
      Table "public.parted_notnull_inh_test1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           | not null | 1
 b      | integer |           | not null | 1
Partition of: parted_notnull_inh_test FOR VALUES IN (1)
drop table parted_notnull_inh_test;`,
			},
			{
				Statement: `create table parted_boolean_col (a bool, b text) partition by list(a);`,
			},
			{
				Statement: `create table parted_boolean_less partition of parted_boolean_col
  for values in ('foo' < 'bar');`,
			},
			{
				Statement: `create table parted_boolean_greater partition of parted_boolean_col
  for values in ('foo' > 'bar');`,
			},
			{
				Statement: `drop table parted_boolean_col;`,
			},
			{
				Statement: `create table parted_collate_must_match (a text collate "C", b text collate "C")
  partition by range (a);`,
			},
			{
				Statement: `create table parted_collate_must_match1 partition of parted_collate_must_match
  (a collate "POSIX") for values from ('a') to ('m');`,
			},
			{
				Statement: `create table parted_collate_must_match2 partition of parted_collate_must_match
  (b collate "POSIX") for values from ('m') to ('z');`,
			},
			{
				Statement: `drop table parted_collate_must_match;`,
			},
			{
				Statement: `create table test_part_coll_posix (a text) partition by range (a collate "POSIX");`,
			},
			{
				Statement: `create table test_part_coll partition of test_part_coll_posix for values from ('a' collate "C") to ('g');`,
			},
			{
				Statement: `create table test_part_coll2 partition of test_part_coll_posix for values from ('g') to ('m');`,
			},
			{
				Statement: `create table test_part_coll_cast partition of test_part_coll_posix for values from (name 'm' collate "C") to ('s');`,
			},
			{
				Statement: `create table test_part_coll_cast2 partition of test_part_coll_posix for values from (name 's') to ('z');`,
			},
			{
				Statement: `drop table test_part_coll_posix;`,
			},
			{
				Statement: `\d+ part_b
                                   Table "public.part_b"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           | not null | 1       | plain    |              | 
Partition of: parted FOR VALUES IN ('b')
Partition constraint: ((a IS NOT NULL) AND (a = 'b'::text))
\d+ part_c
                             Partitioned table "public.part_c"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           | not null | 0       | plain    |              | 
Partition of: parted FOR VALUES IN ('c')
Partition constraint: ((a IS NOT NULL) AND (a = 'c'::text))
Partition key: RANGE (b)
Partitions: part_c_1_10 FOR VALUES FROM (1) TO (10)
\d+ part_c_1_10
                                Table "public.part_c_1_10"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | text    |           |          |         | extended |              | 
 b      | integer |           | not null | 0       | plain    |              | 
Partition of: part_c FOR VALUES FROM (1) TO (10)
Partition constraint: ((a IS NOT NULL) AND (a = 'c'::text) AND (b IS NOT NULL) AND (b >= 1) AND (b < 10))
\d parted
         Partitioned table "public.parted"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | text    |           |          | 
 b      | integer |           | not null | 0
Partition key: LIST (a)
Number of partitions: 3 (Use \d+ to list them.)
\d hash_parted
      Partitioned table "public.hash_parted"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
Partition key: HASH (a)
Number of partitions: 4 (Use \d+ to list them.)
CREATE TABLE range_parted4 (a int, b int, c int) PARTITION BY RANGE (abs(a), abs(b), c);`,
			},
			{
				Statement: `CREATE TABLE unbounded_range_part PARTITION OF range_parted4 FOR VALUES FROM (MINVALUE, MINVALUE, MINVALUE) TO (MAXVALUE, MAXVALUE, MAXVALUE);`,
			},
			{
				Statement: `\d+ unbounded_range_part
                           Table "public.unbounded_range_part"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
 c      | integer |           |          |         | plain   |              | 
Partition of: range_parted4 FOR VALUES FROM (MINVALUE, MINVALUE, MINVALUE) TO (MAXVALUE, MAXVALUE, MAXVALUE)
Partition constraint: ((abs(a) IS NOT NULL) AND (abs(b) IS NOT NULL) AND (c IS NOT NULL))
DROP TABLE unbounded_range_part;`,
			},
			{
				Statement: `CREATE TABLE range_parted4_1 PARTITION OF range_parted4 FOR VALUES FROM (MINVALUE, MINVALUE, MINVALUE) TO (1, MAXVALUE, MAXVALUE);`,
			},
			{
				Statement: `\d+ range_parted4_1
                              Table "public.range_parted4_1"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
 c      | integer |           |          |         | plain   |              | 
Partition of: range_parted4 FOR VALUES FROM (MINVALUE, MINVALUE, MINVALUE) TO (1, MAXVALUE, MAXVALUE)
Partition constraint: ((abs(a) IS NOT NULL) AND (abs(b) IS NOT NULL) AND (c IS NOT NULL) AND (abs(a) <= 1))
CREATE TABLE range_parted4_2 PARTITION OF range_parted4 FOR VALUES FROM (3, 4, 5) TO (6, 7, MAXVALUE);`,
			},
			{
				Statement: `\d+ range_parted4_2
                              Table "public.range_parted4_2"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
 c      | integer |           |          |         | plain   |              | 
Partition of: range_parted4 FOR VALUES FROM (3, 4, 5) TO (6, 7, MAXVALUE)
Partition constraint: ((abs(a) IS NOT NULL) AND (abs(b) IS NOT NULL) AND (c IS NOT NULL) AND ((abs(a) > 3) OR ((abs(a) = 3) AND (abs(b) > 4)) OR ((abs(a) = 3) AND (abs(b) = 4) AND (c >= 5))) AND ((abs(a) < 6) OR ((abs(a) = 6) AND (abs(b) <= 7))))
CREATE TABLE range_parted4_3 PARTITION OF range_parted4 FOR VALUES FROM (6, 8, MINVALUE) TO (9, MAXVALUE, MAXVALUE);`,
			},
			{
				Statement: `\d+ range_parted4_3
                              Table "public.range_parted4_3"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
 c      | integer |           |          |         | plain   |              | 
Partition of: range_parted4 FOR VALUES FROM (6, 8, MINVALUE) TO (9, MAXVALUE, MAXVALUE)
Partition constraint: ((abs(a) IS NOT NULL) AND (abs(b) IS NOT NULL) AND (c IS NOT NULL) AND ((abs(a) > 6) OR ((abs(a) = 6) AND (abs(b) >= 8))) AND (abs(a) <= 9))
DROP TABLE range_parted4;`,
			},
			{
				Statement: `CREATE FUNCTION my_int4_sort(int4,int4) RETURNS int LANGUAGE sql
  AS $$ SELECT CASE WHEN $1 = $2 THEN 0 WHEN $1 > $2 THEN 1 ELSE -1 END; $$;`,
			},
			{
				Statement: `CREATE OPERATOR CLASS test_int4_ops FOR TYPE int4 USING btree AS
  OPERATOR 1 < (int4,int4), OPERATOR 2 <= (int4,int4),
  OPERATOR 3 = (int4,int4), OPERATOR 4 >= (int4,int4),
  OPERATOR 5 > (int4,int4), FUNCTION 1 my_int4_sort(int4,int4);`,
			},
			{
				Statement: `CREATE TABLE partkey_t (a int4) PARTITION BY RANGE (a test_int4_ops);`,
			},
			{
				Statement: `CREATE TABLE partkey_t_1 PARTITION OF partkey_t FOR VALUES FROM (0) TO (1000);`,
			},
			{
				Statement: `INSERT INTO partkey_t VALUES (100);`,
			},
			{
				Statement: `INSERT INTO partkey_t VALUES (200);`,
			},
			{
				Statement: `DROP TABLE parted, list_parted, range_parted, list_parted2, range_parted2, range_parted3;`,
			},
			{
				Statement: `DROP TABLE partkey_t, hash_parted, hash_parted2;`,
			},
			{
				Statement: `DROP OPERATOR CLASS test_int4_ops USING btree;`,
			},
			{
				Statement: `DROP FUNCTION my_int4_sort(int4,int4);`,
			},
			{
				Statement: `CREATE TABLE parted_col_comment (a int, b text) PARTITION BY LIST (a);`,
			},
			{
				Statement: `COMMENT ON TABLE parted_col_comment IS 'Am partitioned table';`,
			},
			{
				Statement: `COMMENT ON COLUMN parted_col_comment.a IS 'Partition key';`,
			},
			{
				Statement: `SELECT obj_description('parted_col_comment'::regclass);`,
				Results:   []sql.Row{{`Am partitioned table`}},
			},
			{
				Statement: `\d+ parted_col_comment
                        Partitioned table "public.parted_col_comment"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target |  Description  
--------+---------+-----------+----------+---------+----------+--------------+---------------
 a      | integer |           |          |         | plain    |              | Partition key
 b      | text    |           |          |         | extended |              | 
Partition key: LIST (a)
Number of partitions: 0
DROP TABLE parted_col_comment;`,
			},
			{
				Statement: `CREATE TABLE arrlp (a int[]) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE arrlp12 PARTITION OF arrlp FOR VALUES IN ('{1}', '{2}');`,
			},
			{
				Statement: `\d+ arrlp12
                                   Table "public.arrlp12"
 Column |   Type    | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+-----------+-----------+----------+---------+----------+--------------+-------------
 a      | integer[] |           |          |         | extended |              | 
Partition of: arrlp FOR VALUES IN ('{1}', '{2}')
Partition constraint: ((a IS NOT NULL) AND ((a = '{1}'::integer[]) OR (a = '{2}'::integer[])))
DROP TABLE arrlp;`,
			},
			{
				Statement: `create table boolspart (a bool) partition by list (a);`,
			},
			{
				Statement: `create table boolspart_t partition of boolspart for values in (true);`,
			},
			{
				Statement: `create table boolspart_f partition of boolspart for values in (false);`,
			},
			{
				Statement: `\d+ boolspart
                           Partitioned table "public.boolspart"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | boolean |           |          |         | plain   |              | 
Partition key: LIST (a)
Partitions: boolspart_f FOR VALUES IN (false),
            boolspart_t FOR VALUES IN (true)
drop table boolspart;`,
			},
			{
				Statement: `create table perm_parted (a int) partition by list (a);`,
			},
			{
				Statement: `create temporary table temp_parted (a int) partition by list (a);`,
			},
			{
				Statement:   `create table perm_part partition of temp_parted default; -- error`,
				ErrorString: `cannot create a permanent relation as partition of temporary relation "temp_parted"`,
			},
			{
				Statement:   `create temp table temp_part partition of perm_parted default; -- error`,
				ErrorString: `cannot create a temporary relation as partition of permanent relation "perm_parted"`,
			},
			{
				Statement: `create temp table temp_part partition of temp_parted default; -- ok`,
			},
			{
				Statement: `drop table perm_parted cascade;`,
			},
			{
				Statement: `drop table temp_parted cascade;`,
			},
			{
				Statement: `create table tab_part_create (a int) partition by list (a);`,
			},
			{
				Statement: `create or replace function func_part_create() returns trigger
  language plpgsql as $$
  begin
    execute 'create table tab_part_create_1 partition of tab_part_create for values in (1)';`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `  end $$;`,
			},
			{
				Statement: `create trigger trig_part_create before insert on tab_part_create
  for each statement execute procedure func_part_create();`,
			},
			{
				Statement:   `insert into tab_part_create values (1);`,
				ErrorString: `cannot CREATE TABLE .. PARTITION OF "tab_part_create" because it is being used by active queries in this session`,
			},
			{
				Statement: `CONTEXT:  SQL statement "create table tab_part_create_1 partition of tab_part_create for values in (1)"
PL/pgSQL function func_part_create() line 3 at EXECUTE
drop table tab_part_create;`,
			},
			{
				Statement: `drop function func_part_create();`,
			},
			{
				Statement: `create table volatile_partbound_test (partkey timestamp) partition by range (partkey);`,
			},
			{
				Statement: `create table volatile_partbound_test1 partition of volatile_partbound_test for values from (minvalue) to (current_timestamp);`,
			},
			{
				Statement: `create table volatile_partbound_test2 partition of volatile_partbound_test for values from (current_timestamp) to (maxvalue);`,
			},
			{
				Statement: `insert into volatile_partbound_test values (current_timestamp);`,
			},
			{
				Statement: `select tableoid::regclass from volatile_partbound_test;`,
				Results:   []sql.Row{{`volatile_partbound_test2`}},
			},
			{
				Statement: `drop table volatile_partbound_test;`,
			},
			{
				Statement: `create table defcheck (a int, b int) partition by list (b);`,
			},
			{
				Statement: `create table defcheck_def (a int, c int, b int);`,
			},
			{
				Statement: `alter table defcheck_def drop c;`,
			},
			{
				Statement: `alter table defcheck attach partition defcheck_def default;`,
			},
			{
				Statement: `alter table defcheck_def add check (b <= 0 and b is not null);`,
			},
			{
				Statement: `create table defcheck_1 partition of defcheck for values in (1, null);`,
			},
			{
				Statement: `insert into defcheck_def values (0, 0);`,
			},
			{
				Statement:   `create table defcheck_0 partition of defcheck for values in (0);`,
				ErrorString: `updated partition constraint for default partition "defcheck_def" would be violated by some row`,
			},
			{
				Statement: `drop table defcheck;`,
			},
			{
				Statement: `create table part_column_drop (
  useless_1 int,
  id int,
  useless_2 int,
  d int,
  b int,
  useless_3 int
) partition by range (id);`,
			},
			{
				Statement: `alter table part_column_drop drop column useless_1;`,
			},
			{
				Statement: `alter table part_column_drop drop column useless_2;`,
			},
			{
				Statement: `alter table part_column_drop drop column useless_3;`,
			},
			{
				Statement: `create index part_column_drop_b_pred on part_column_drop(b) where b = 1;`,
			},
			{
				Statement: `create index part_column_drop_b_expr on part_column_drop((b = 1));`,
			},
			{
				Statement: `create index part_column_drop_d_pred on part_column_drop(d) where d = 2;`,
			},
			{
				Statement: `create index part_column_drop_d_expr on part_column_drop((d = 2));`,
			},
			{
				Statement: `create table part_column_drop_1_10 partition of
  part_column_drop for values from (1) to (10);`,
			},
			{
				Statement: `\d part_column_drop
    Partitioned table "public.part_column_drop"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           |          | 
 d      | integer |           |          | 
 b      | integer |           |          | 
Partition key: RANGE (id)
Indexes:
    "part_column_drop_b_expr" btree ((b = 1))
    "part_column_drop_b_pred" btree (b) WHERE b = 1
    "part_column_drop_d_expr" btree ((d = 2))
    "part_column_drop_d_pred" btree (d) WHERE d = 2
Number of partitions: 1 (Use \d+ to list them.)
\d part_column_drop_1_10
       Table "public.part_column_drop_1_10"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           |          | 
 d      | integer |           |          | 
 b      | integer |           |          | 
Partition of: part_column_drop FOR VALUES FROM (1) TO (10)
Indexes:
    "part_column_drop_1_10_b_idx" btree (b) WHERE b = 1
    "part_column_drop_1_10_d_idx" btree (d) WHERE d = 2
    "part_column_drop_1_10_expr_idx" btree ((b = 1))
    "part_column_drop_1_10_expr_idx1" btree ((d = 2))
drop table part_column_drop;`,
			},
		},
	})
}
