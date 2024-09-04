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

func TestStatsExt(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_stats_ext)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_stats_ext,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create function check_estimated_rows(text) returns table (estimated int, actual int)
language plpgsql as
$$
declare
    ln text;`,
			},
			{
				Statement: `    tmp text[];`,
			},
			{
				Statement: `    first_row bool := true;`,
			},
			{
				Statement: `begin
    for ln in
        execute format('explain analyze %s', $1)
    loop
        if first_row then
            first_row := false;`,
			},
			{
				Statement: `            tmp := regexp_match(ln, 'rows=(\d*) .* rows=(\d*)');`,
			},
			{
				Statement: `            return query select tmp[1]::int, tmp[2]::int;`,
			},
			{
				Statement: `        end if;`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TABLE ext_stats_test (x text, y int, z int);`,
			},
			{
				Statement:   `CREATE STATISTICS tst;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON a, b;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement:   `CREATE STATISTICS tst FROM sometab;`,
				ErrorString: `syntax error at or near "FROM"`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON a, b FROM nonexistent;`,
				ErrorString: `relation "nonexistent" does not exist`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON a, b FROM ext_stats_test;`,
				ErrorString: `column "a" does not exist`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON x, x, y FROM ext_stats_test;`,
				ErrorString: `duplicate column name in statistics definition`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON x, x, y, x, x, y, x, x, y FROM ext_stats_test;`,
				ErrorString: `cannot have more than 8 columns in statistics`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON x, x, y, x, x, (x || 'x'), (y + 1), (x || 'x'), (x || 'x'), (y + 1) FROM ext_stats_test;`,
				ErrorString: `cannot have more than 8 columns in statistics`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON (x || 'x'), (x || 'x'), (y + 1), (x || 'x'), (x || 'x'), (y + 1), (x || 'x'), (x || 'x'), (y + 1) FROM ext_stats_test;`,
				ErrorString: `cannot have more than 8 columns in statistics`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON (x || 'x'), (x || 'x'), y FROM ext_stats_test;`,
				ErrorString: `duplicate expression in statistics definition`,
			},
			{
				Statement:   `CREATE STATISTICS tst (unrecognized) ON x, y FROM ext_stats_test;`,
				ErrorString: `unrecognized statistics kind "unrecognized"`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON (y) FROM ext_stats_test; -- single column reference`,
				ErrorString: `extended statistics require at least 2 columns`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON y + z FROM ext_stats_test; -- missing parentheses`,
				ErrorString: `syntax error at or near "+"`,
			},
			{
				Statement:   `CREATE STATISTICS tst ON (x, y) FROM ext_stats_test; -- tuple expression`,
				ErrorString: `syntax error at or near ","`,
			},
			{
				Statement: `DROP TABLE ext_stats_test;`,
			},
			{
				Statement: `CREATE TABLE ab1 (a INTEGER, b INTEGER, c INTEGER);`,
			},
			{
				Statement: `CREATE STATISTICS IF NOT EXISTS ab1_a_b_stats ON a, b FROM ab1;`,
			},
			{
				Statement: `COMMENT ON STATISTICS ab1_a_b_stats IS 'new comment';`,
			},
			{
				Statement: `CREATE ROLE regress_stats_ext;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_stats_ext;`,
			},
			{
				Statement:   `COMMENT ON STATISTICS ab1_a_b_stats IS 'changed comment';`,
				ErrorString: `must be owner of statistics object ab1_a_b_stats`,
			},
			{
				Statement:   `DROP STATISTICS ab1_a_b_stats;`,
				ErrorString: `must be owner of statistics object ab1_a_b_stats`,
			},
			{
				Statement:   `ALTER STATISTICS ab1_a_b_stats RENAME TO ab1_a_b_stats_new;`,
				ErrorString: `must be owner of statistics object ab1_a_b_stats`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP ROLE regress_stats_ext;`,
			},
			{
				Statement: `CREATE STATISTICS IF NOT EXISTS ab1_a_b_stats ON a, b FROM ab1;`,
			},
			{
				Statement: `DROP STATISTICS ab1_a_b_stats;`,
			},
			{
				Statement: `CREATE SCHEMA regress_schema_2;`,
			},
			{
				Statement: `CREATE STATISTICS regress_schema_2.ab1_a_b_stats ON a, b FROM ab1;`,
			},
			{
				Statement: `SELECT pg_get_statisticsobjdef(oid) FROM pg_statistic_ext WHERE stxname = 'ab1_a_b_stats';`,
				Results:   []sql.Row{{`CREATE STATISTICS regress_schema_2.ab1_a_b_stats ON a, b FROM ab1`}},
			},
			{
				Statement: `DROP STATISTICS regress_schema_2.ab1_a_b_stats;`,
			},
			{
				Statement: `CREATE STATISTICS ab1_b_c_stats ON b, c FROM ab1;`,
			},
			{
				Statement: `CREATE STATISTICS ab1_a_b_c_stats ON a, b, c FROM ab1;`,
			},
			{
				Statement: `CREATE STATISTICS ab1_b_a_stats ON b, a FROM ab1;`,
			},
			{
				Statement: `ALTER TABLE ab1 DROP COLUMN a;`,
			},
			{
				Statement: `\d ab1
                Table "public.ab1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 b      | integer |           |          | 
 c      | integer |           |          | 
Statistics objects:
    "public.ab1_b_c_stats" ON b, c FROM ab1
SELECT stxname FROM pg_statistic_ext WHERE stxname LIKE 'ab1%';`,
				Results: []sql.Row{{`ab1_b_c_stats`}},
			},
			{
				Statement: `DROP TABLE ab1;`,
			},
			{
				Statement: `SELECT stxname FROM pg_statistic_ext WHERE stxname LIKE 'ab1%';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `CREATE TABLE ab1 (a INTEGER, b INTEGER);`,
			},
			{
				Statement: `ALTER TABLE ab1 ALTER a SET STATISTICS 0;`,
			},
			{
				Statement: `INSERT INTO ab1 SELECT a, a%23 FROM generate_series(1, 1000) a;`,
			},
			{
				Statement: `CREATE STATISTICS ab1_a_b_stats ON a, b FROM ab1;`,
			},
			{
				Statement: `ANALYZE ab1;`,
			},
			{
				Statement: `ALTER TABLE ab1 ALTER a SET STATISTICS -1;`,
			},
			{
				Statement: `ALTER STATISTICS ab1_a_b_stats SET STATISTICS 0;`,
			},
			{
				Statement: `\d ab1
                Table "public.ab1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | integer |           |          | 
Statistics objects:
    "public.ab1_a_b_stats" ON a, b FROM ab1; STATISTICS 0
ANALYZE ab1;`,
			},
			{
				Statement: `SELECT stxname, stxdndistinct, stxddependencies, stxdmcv, stxdinherit
  FROM pg_statistic_ext s LEFT JOIN pg_statistic_ext_data d ON (d.stxoid = s.oid)
 WHERE s.stxname = 'ab1_a_b_stats';`,
				Results: []sql.Row{{`ab1_a_b_stats`, ``, ``, ``, ``}},
			},
			{
				Statement: `ALTER STATISTICS ab1_a_b_stats SET STATISTICS -1;`,
			},
			{
				Statement: `\d+ ab1
                                    Table "public.ab1"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           |          |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
Statistics objects:
    "public.ab1_a_b_stats" ON a, b FROM ab1
ANALYZE ab1 (a);`,
			},
			{
				Statement: `ANALYZE ab1;`,
			},
			{
				Statement: `DROP TABLE ab1;`,
			},
			{
				Statement:   `ALTER STATISTICS ab1_a_b_stats SET STATISTICS 0;`,
				ErrorString: `statistics object "ab1_a_b_stats" does not exist`,
			},
			{
				Statement: `ALTER STATISTICS IF EXISTS ab1_a_b_stats SET STATISTICS 0;`,
			},
			{
				Statement: `CREATE TABLE ab1 (a INTEGER, b INTEGER);`,
			},
			{
				Statement: `CREATE TABLE ab1c () INHERITS (ab1);`,
			},
			{
				Statement: `INSERT INTO ab1 VALUES (1,1);`,
			},
			{
				Statement: `CREATE STATISTICS ab1_a_b_stats ON a, b FROM ab1;`,
			},
			{
				Statement: `ANALYZE ab1;`,
			},
			{
				Statement: `DROP TABLE ab1 CASCADE;`,
			},
			{
				Statement: `CREATE TABLE stxdinh(a int, b int);`,
			},
			{
				Statement: `CREATE TABLE stxdinh1() INHERITS(stxdinh);`,
			},
			{
				Statement: `CREATE TABLE stxdinh2() INHERITS(stxdinh);`,
			},
			{
				Statement: `INSERT INTO stxdinh SELECT mod(a,50), mod(a,100) FROM generate_series(0, 1999) a;`,
			},
			{
				Statement: `INSERT INTO stxdinh1 SELECT mod(a,100), mod(a,100) FROM generate_series(0, 999) a;`,
			},
			{
				Statement: `INSERT INTO stxdinh2 SELECT mod(a,100), mod(a,100) FROM generate_series(0, 999) a;`,
			},
			{
				Statement: `VACUUM ANALYZE stxdinh, stxdinh1, stxdinh2;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT a, b FROM stxdinh* GROUP BY 1, 2');`,
				Results:   []sql.Row{{400, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT a, b FROM stxdinh* WHERE a = 0 AND b = 0');`,
				Results:   []sql.Row{{3, 40}},
			},
			{
				Statement: `CREATE STATISTICS stxdinh ON a, b FROM stxdinh;`,
			},
			{
				Statement: `VACUUM ANALYZE stxdinh, stxdinh1, stxdinh2;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT a, b FROM stxdinh* GROUP BY 1, 2');`,
				Results:   []sql.Row{{150, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT a, b FROM stxdinh* WHERE a = 0 AND b = 0');`,
				Results:   []sql.Row{{22, 40}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT a, b FROM ONLY stxdinh GROUP BY 1, 2');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT a, b FROM ONLY stxdinh WHERE a = 0 AND b = 0');`,
				Results:   []sql.Row{{20, 20}},
			},
			{
				Statement: `DROP TABLE stxdinh, stxdinh1, stxdinh2;`,
			},
			{
				Statement: `CREATE TABLE stxdinp(i int, a int, b int) PARTITION BY RANGE (i);`,
			},
			{
				Statement: `CREATE TABLE stxdinp1 PARTITION OF stxdinp FOR VALUES FROM (1) TO (100);`,
			},
			{
				Statement: `INSERT INTO stxdinp SELECT 1, a/100, a/100 FROM generate_series(1, 999) a;`,
			},
			{
				Statement: `CREATE STATISTICS stxdinp ON (a + 1), a, b FROM stxdinp;`,
			},
			{
				Statement: `VACUUM ANALYZE stxdinp; -- partitions are processed recursively`,
			},
			{
				Statement: `SELECT 1 FROM pg_statistic_ext WHERE stxrelid = 'stxdinp'::regclass;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT a, b FROM stxdinp GROUP BY 1, 2');`,
				Results:   []sql.Row{{10, 10}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT a + 1, b FROM ONLY stxdinp GROUP BY 1, 2');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `DROP TABLE stxdinp;`,
			},
			{
				Statement: `CREATE TABLE ab1 (a INTEGER, b INTEGER, c TIMESTAMP, d TIMESTAMPTZ);`,
			},
			{
				Statement: `CREATE STATISTICS ab1_exprstat_1 ON (a+b) FROM ab1;`,
			},
			{
				Statement: `CREATE STATISTICS ab1_exprstat_2 ON (a+b) FROM ab1;`,
			},
			{
				Statement: `SELECT stxkind FROM pg_statistic_ext WHERE stxname = 'ab1_exprstat_2';`,
				Results:   []sql.Row{{`{e}`}},
			},
			{
				Statement: `CREATE STATISTICS ab1_exprstat_3 ON (a+b), a FROM ab1;`,
			},
			{
				Statement: `SELECT stxkind FROM pg_statistic_ext WHERE stxname = 'ab1_exprstat_3';`,
				Results:   []sql.Row{{`{d,f,m,e}`}},
			},
			{
				Statement: `CREATE STATISTICS ab1_exprstat_4 ON date_trunc('day', d) FROM ab1;`,
			},
			{
				Statement: `CREATE STATISTICS ab1_exprstat_5 ON date_trunc('day', c) FROM ab1;`,
			},
			{
				Statement: `CREATE STATISTICS ab1_exprstat_6 ON
  (case a when 1 then true else false end), b FROM ab1;`,
			},
			{
				Statement: `INSERT INTO ab1
SELECT x / 10, x / 3,
    '2020-10-01'::timestamp + x * interval '1 day',
    '2020-10-01'::timestamptz + x * interval '1 day'
FROM generate_series(1, 100) x;`,
			},
			{
				Statement: `ANALYZE ab1;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM ab1 WHERE (case a when 1 then true else false end) AND b=2');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `DROP TABLE ab1;`,
			},
			{
				Statement: `CREATE schema tststats;`,
			},
			{
				Statement: `CREATE TABLE tststats.t (a int, b int, c text);`,
			},
			{
				Statement: `CREATE INDEX ti ON tststats.t (a, b);`,
			},
			{
				Statement: `CREATE SEQUENCE tststats.s;`,
			},
			{
				Statement: `CREATE VIEW tststats.v AS SELECT * FROM tststats.t;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW tststats.mv AS SELECT * FROM tststats.t;`,
			},
			{
				Statement: `CREATE TYPE tststats.ty AS (a int, b int, c text);`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER extstats_dummy_fdw;`,
			},
			{
				Statement: `CREATE SERVER extstats_dummy_srv FOREIGN DATA WRAPPER extstats_dummy_fdw;`,
			},
			{
				Statement: `CREATE FOREIGN TABLE tststats.f (a int, b int, c text) SERVER extstats_dummy_srv;`,
			},
			{
				Statement: `CREATE TABLE tststats.pt (a int, b int, c text) PARTITION BY RANGE (a, b);`,
			},
			{
				Statement: `CREATE TABLE tststats.pt1 PARTITION OF tststats.pt FOR VALUES FROM (-10, -10) TO (10, 10);`,
			},
			{
				Statement: `CREATE STATISTICS tststats.s1 ON a, b FROM tststats.t;`,
			},
			{
				Statement:   `CREATE STATISTICS tststats.s2 ON a, b FROM tststats.ti;`,
				ErrorString: `cannot define statistics for relation "ti"`,
			},
			{
				Statement:   `CREATE STATISTICS tststats.s3 ON a, b FROM tststats.s;`,
				ErrorString: `cannot define statistics for relation "s"`,
			},
			{
				Statement:   `CREATE STATISTICS tststats.s4 ON a, b FROM tststats.v;`,
				ErrorString: `cannot define statistics for relation "v"`,
			},
			{
				Statement: `CREATE STATISTICS tststats.s5 ON a, b FROM tststats.mv;`,
			},
			{
				Statement:   `CREATE STATISTICS tststats.s6 ON a, b FROM tststats.ty;`,
				ErrorString: `cannot define statistics for relation "ty"`,
			},
			{
				Statement: `CREATE STATISTICS tststats.s7 ON a, b FROM tststats.f;`,
			},
			{
				Statement: `CREATE STATISTICS tststats.s8 ON a, b FROM tststats.pt;`,
			},
			{
				Statement: `CREATE STATISTICS tststats.s9 ON a, b FROM tststats.pt1;`,
			},
			{
				Statement: `DO $$
DECLARE
	relname text := reltoastrelid::regclass FROM pg_class WHERE oid = 'tststats.t'::regclass;`,
			},
			{
				Statement: `BEGIN
	EXECUTE 'CREATE STATISTICS tststats.s10 ON a, b FROM ' || relname;`,
			},
			{
				Statement: `EXCEPTION WHEN wrong_object_type THEN
	RAISE NOTICE 'stats on toast table not created';`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `DROP SCHEMA tststats CASCADE;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER extstats_dummy_fdw CASCADE;`,
			},
			{
				Statement: `CREATE TABLE ndistinct (
    filler1 TEXT,
    filler2 NUMERIC,
    a INT,
    b INT,
    filler3 DATE,
    c INT,
    d INT
)
WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO ndistinct (a, b, c, filler1)
     SELECT i/100, i/100, i/100, cash_words((i/100)::money)
       FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{100, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY b, c');`,
				Results:   []sql.Row{{100, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, c');`,
				Results:   []sql.Row{{100, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, c, d');`,
				Results:   []sql.Row{{200, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY b, c, d');`,
				Results:   []sql.Row{{200, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (a+1)');`,
				Results:   []sql.Row{{100, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100)');`,
				Results:   []sql.Row{{100, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100), (2*c)');`,
				Results:   []sql.Row{{100, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (a+1), (b+100)');`,
				Results:   []sql.Row{{100, 11}},
			},
			{
				Statement: `CREATE STATISTICS s10 ON a, b, c FROM ndistinct;`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT s.stxkind, d.stxdndistinct
  FROM pg_statistic_ext s, pg_statistic_ext_data d
 WHERE s.stxrelid = 'ndistinct'::regclass
   AND d.stxoid = s.oid;`,
				Results: []sql.Row{{`{d,f,m}`, `{"3, 4": 11, "3, 6": 11, "4, 6": 11, "3, 4, 6": 11}`}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY ctid, a, b');`,
				Results:   []sql.Row{{1000, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{11, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY b, c');`,
				Results:   []sql.Row{{11, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, c');`,
				Results:   []sql.Row{{11, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (a+1)');`,
				Results:   []sql.Row{{11, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100)');`,
				Results:   []sql.Row{{11, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100), (2*c)');`,
				Results:   []sql.Row{{11, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (a+1), (b+100)');`,
				Results:   []sql.Row{{11, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, c, d');`,
				Results:   []sql.Row{{200, 11}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY b, c, d');`,
				Results:   []sql.Row{{200, 11}},
			},
			{
				Statement: `TRUNCATE TABLE ndistinct;`,
			},
			{
				Statement: `INSERT INTO ndistinct (a, b, c, filler1)
     SELECT mod(i,13), mod(i,17), mod(i,19),
            cash_words(mod(i,23)::int::money)
       FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT s.stxkind, d.stxdndistinct
  FROM pg_statistic_ext s, pg_statistic_ext_data d
 WHERE s.stxrelid = 'ndistinct'::regclass
   AND d.stxoid = s.oid;`,
				Results: []sql.Row{{`{d,f,m}`, `{"3, 4": 221, "3, 6": 247, "4, 6": 323, "3, 4, 6": 1000}`}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{221, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, c');`,
				Results:   []sql.Row{{1000, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, c, d');`,
				Results:   []sql.Row{{1000, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY b, c, d');`,
				Results:   []sql.Row{{323, 323}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, d');`,
				Results:   []sql.Row{{200, 13}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (a+1)');`,
				Results:   []sql.Row{{221, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100)');`,
				Results:   []sql.Row{{221, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100), (2*c)');`,
				Results:   []sql.Row{{1000, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (a+1), (b+100)');`,
				Results:   []sql.Row{{221, 221}},
			},
			{
				Statement: `DROP STATISTICS s10;`,
			},
			{
				Statement: `SELECT s.stxkind, d.stxdndistinct
  FROM pg_statistic_ext s, pg_statistic_ext_data d
 WHERE s.stxrelid = 'ndistinct'::regclass
   AND d.stxoid = s.oid;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{100, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, c');`,
				Results:   []sql.Row{{100, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, c, d');`,
				Results:   []sql.Row{{200, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY b, c, d');`,
				Results:   []sql.Row{{200, 323}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, d');`,
				Results:   []sql.Row{{200, 13}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (a+1)');`,
				Results:   []sql.Row{{100, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100)');`,
				Results:   []sql.Row{{100, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100), (2*c)');`,
				Results:   []sql.Row{{100, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (a+1), (b+100)');`,
				Results:   []sql.Row{{100, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100)');`,
				Results:   []sql.Row{{100, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100), (2*c)');`,
				Results:   []sql.Row{{100, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (a+1), (b+100)');`,
				Results:   []sql.Row{{100, 221}},
			},
			{
				Statement: `CREATE STATISTICS s10 (ndistinct) ON (a+1), (b+100), (2*c) FROM ndistinct;`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT s.stxkind, d.stxdndistinct
  FROM pg_statistic_ext s, pg_statistic_ext_data d
 WHERE s.stxrelid = 'ndistinct'::regclass
   AND d.stxoid = s.oid;`,
				Results: []sql.Row{{`{d,e}`, `{"-1, -2": 221, "-1, -3": 247, "-2, -3": 323, "-1, -2, -3": 1000}`}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100)');`,
				Results:   []sql.Row{{221, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a+1), (b+100), (2*c)');`,
				Results:   []sql.Row{{1000, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (a+1), (b+100)');`,
				Results:   []sql.Row{{221, 221}},
			},
			{
				Statement: `DROP STATISTICS s10;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{100, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (2*c)');`,
				Results:   []sql.Row{{100, 247}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (2*c)');`,
				Results:   []sql.Row{{100, 1000}},
			},
			{
				Statement: `CREATE STATISTICS s10 (ndistinct) ON a, b, (2*c) FROM ndistinct;`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT s.stxkind, d.stxdndistinct
  FROM pg_statistic_ext s, pg_statistic_ext_data d
 WHERE s.stxrelid = 'ndistinct'::regclass
   AND d.stxoid = s.oid;`,
				Results: []sql.Row{{`{d,e}`, `{"3, 4": 221, "3, -1": 247, "4, -1": 323, "3, 4, -1": 1000}`}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{221, 221}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (2*c)');`,
				Results:   []sql.Row{{247, 247}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (2*c)');`,
				Results:   []sql.Row{{1000, 1000}},
			},
			{
				Statement: `DROP STATISTICS s10;`,
			},
			{
				Statement: `TRUNCATE ndistinct;`,
			},
			{
				Statement: `INSERT INTO ndistinct (a, b, c, d)
     SELECT mod(i,3), mod(i,9), mod(i,5), mod(i,20)
       FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{27, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1)');`,
				Results:   []sql.Row{{27, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), b');`,
				Results:   []sql.Row{{27, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1)');`,
				Results:   []sql.Row{{27, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1), c');`,
				Results:   []sql.Row{{100, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (c*10)');`,
				Results:   []sql.Row{{100, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1), c, (d - 1)');`,
				Results:   []sql.Row{{100, 180}},
			},
			{
				Statement: `CREATE STATISTICS s11 (ndistinct) ON a, b FROM ndistinct;`,
			},
			{
				Statement: `CREATE STATISTICS s12 (ndistinct) ON c, d FROM ndistinct;`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1), c');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (c*10)');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1), c, (d - 1)');`,
				Results:   []sql.Row{{100, 180}},
			},
			{
				Statement: `DROP STATISTICS s12;`,
			},
			{
				Statement: `CREATE STATISTICS s12 (ndistinct) ON (c * 10), (d - 1) FROM ndistinct;`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1), c');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (c*10)');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1), c, (d - 1)');`,
				Results:   []sql.Row{{100, 180}},
			},
			{
				Statement: `DROP STATISTICS s12;`,
			},
			{
				Statement: `CREATE STATISTICS s12 (ndistinct) ON c, d, (c * 10), (d - 1) FROM ndistinct;`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1), c');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (c*10)');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1), c, (d - 1)');`,
				Results:   []sql.Row{{100, 180}},
			},
			{
				Statement: `DROP STATISTICS s11;`,
			},
			{
				Statement: `CREATE STATISTICS s11 (ndistinct) ON a, b, (a*5), (b+1) FROM ndistinct;`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1), c');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (c*10)');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1), c, (d - 1)');`,
				Results:   []sql.Row{{100, 180}},
			},
			{
				Statement: `DROP STATISTICS s11;`,
			},
			{
				Statement: `DROP STATISTICS s12;`,
			},
			{
				Statement: `CREATE STATISTICS s11 (ndistinct) ON a, b, (a*5), (b+1) FROM ndistinct;`,
			},
			{
				Statement: `CREATE STATISTICS s12 (ndistinct) ON a, (b+1), (c * 10) FROM ndistinct;`,
			},
			{
				Statement: `ANALYZE ndistinct;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), b');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1)');`,
				Results:   []sql.Row{{9, 9}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY (a*5), (b+1), c');`,
				Results:   []sql.Row{{45, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, b, (c*10)');`,
				Results:   []sql.Row{{100, 45}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT COUNT(*) FROM ndistinct GROUP BY a, (b+1), c, (d - 1)');`,
				Results:   []sql.Row{{100, 180}},
			},
			{
				Statement: `DROP STATISTICS s11;`,
			},
			{
				Statement: `DROP STATISTICS s12;`,
			},
			{
				Statement: `CREATE TABLE functional_dependencies (
    filler1 TEXT,
    filler2 NUMERIC,
    a INT,
    b TEXT,
    filler3 DATE,
    c INT,
    d TEXT
)
WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `CREATE INDEX fdeps_ab_idx ON functional_dependencies (a, b);`,
			},
			{
				Statement: `CREATE INDEX fdeps_abc_idx ON functional_dependencies (a, b, c);`,
			},
			{
				Statement: `INSERT INTO functional_dependencies (a, b, c, filler1)
     SELECT mod(i, 5), mod(i, 7), mod(i, 11), i FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE functional_dependencies;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{29, 29}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{3, 3}},
			},
			{
				Statement: `CREATE STATISTICS func_deps_stat (dependencies) ON a, b, c FROM functional_dependencies;`,
			},
			{
				Statement: `ANALYZE functional_dependencies;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{29, 29}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{3, 3}},
			},
			{
				Statement: `TRUNCATE functional_dependencies;`,
			},
			{
				Statement: `DROP STATISTICS func_deps_stat;`,
			},
			{
				Statement: `INSERT INTO functional_dependencies (a, b, c, filler1)
     SELECT i, i, i, i FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE functional_dependencies;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE mod(a, 11) = 1 AND mod(b::int, 13) = 1');`,
				Results:   []sql.Row{{1, 35}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE mod(a, 11) = 1 AND mod(b::int, 13) = 1 AND mod(c, 7) = 1');`,
				Results:   []sql.Row{{1, 5}},
			},
			{
				Statement: `CREATE STATISTICS func_deps_stat (dependencies) ON (mod(a,11)), (mod(b::int, 13)), (mod(c, 7)) FROM functional_dependencies;`,
			},
			{
				Statement: `ANALYZE functional_dependencies;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE mod(a, 11) = 1 AND mod(b::int, 13) = 1');`,
				Results:   []sql.Row{{35, 35}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE mod(a, 11) = 1 AND mod(b::int, 13) = 1 AND mod(c, 7) = 1');`,
				Results:   []sql.Row{{5, 5}},
			},
			{
				Statement: `TRUNCATE functional_dependencies;`,
			},
			{
				Statement: `DROP STATISTICS func_deps_stat;`,
			},
			{
				Statement: `INSERT INTO functional_dependencies (a, b, c, filler1)
     SELECT mod(i,100), mod(i,50), mod(i,25), i FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE functional_dependencies;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 51) AND b = ''1''');`,
				Results:   []sql.Row{{2, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 51) AND b IN (''1'', ''2'')');`,
				Results:   []sql.Row{{4, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 2, 51, 52) AND b IN (''1'', ''2'')');`,
				Results:   []sql.Row{{8, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 2, 51, 52) AND b = ''1''');`,
				Results:   []sql.Row{{4, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 26, 51, 76) AND b IN (''1'', ''26'') AND c = 1');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 26, 51, 76) AND b IN (''1'', ''26'') AND c IN (1)');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 2, 26, 27, 51, 52, 76, 77) AND b IN (''1'', ''2'', ''26'', ''27'') AND c IN (1, 2)');`,
				Results:   []sql.Row{{3, 400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a = 1 OR a = 51) AND b = ''1''');`,
				Results:   []sql.Row{{2, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a = 1 OR a = 51) AND (b = ''1'' OR b = ''2'')');`,
				Results:   []sql.Row{{4, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a = 1 OR a = 2 OR a = 51 OR a = 52) AND (b = ''1'' OR b = ''2'')');`,
				Results:   []sql.Row{{8, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a = 1 OR b = ''1'') AND b = ''1''');`,
				Results:   []sql.Row{{3, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 51]) AND b = ''1''');`,
				Results:   []sql.Row{{2, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 51]) AND b = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{4, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 2, 51, 52]) AND b = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{8, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 26, 51, 76]) AND b = ANY (ARRAY[''1'', ''26'']) AND c = 1');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 26, 51, 76]) AND b = ANY (ARRAY[''1'', ''26'']) AND c = ANY (ARRAY[1])');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 2, 26, 27, 51, 52, 76, 77]) AND b = ANY (ARRAY[''1'', ''2'', ''26'', ''27'']) AND c = ANY (ARRAY[1, 2])');`,
				Results:   []sql.Row{{3, 400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a < ANY (ARRAY[1, 51]) AND b > ''1''');`,
				Results:   []sql.Row{{2472, 2400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a >= ANY (ARRAY[1, 51]) AND b <= ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1441, 1250}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a <= ANY (ARRAY[1, 2, 51, 52]) AND b >= ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{3909, 2550}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 51) AND b = ALL (ARRAY[''1''])');`,
				Results:   []sql.Row{{2, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 51) AND b = ALL (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 2, 51, 52) AND b = ALL (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `CREATE STATISTICS func_deps_stat (dependencies) ON a, b, c FROM functional_dependencies;`,
			},
			{
				Statement: `ANALYZE functional_dependencies;`,
			},
			{
				Statement: `SELECT dependencies FROM pg_stats_ext WHERE statistics_name = 'func_deps_stat';`,
				Results:   []sql.Row{{`{"3 => 4": 1.000000, "3 => 6": 1.000000, "4 => 6": 1.000000, "3, 4 => 6": 1.000000, "3, 6 => 4": 1.000000}`}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 51) AND b = ''1''');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 51) AND b IN (''1'', ''2'')');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 2, 51, 52) AND b IN (''1'', ''2'')');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 2, 51, 52) AND b = ''1''');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 26, 51, 76) AND b IN (''1'', ''26'') AND c = 1');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 26, 51, 76) AND b IN (''1'', ''26'') AND c IN (1)');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 2, 26, 27, 51, 52, 76, 77) AND b IN (''1'', ''2'', ''26'', ''27'') AND c IN (1, 2)');`,
				Results:   []sql.Row{{400, 400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a = 1 OR a = 51) AND b = ''1''');`,
				Results:   []sql.Row{{99, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a = 1 OR a = 51) AND (b = ''1'' OR b = ''2'')');`,
				Results:   []sql.Row{{99, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a = 1 OR a = 2 OR a = 51 OR a = 52) AND (b = ''1'' OR b = ''2'')');`,
				Results:   []sql.Row{{197, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a = 1 OR b = ''1'') AND b = ''1''');`,
				Results:   []sql.Row{{3, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 51]) AND b = ''1''');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 51]) AND b = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 2, 51, 52]) AND b = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 26, 51, 76]) AND b = ANY (ARRAY[''1'', ''26'']) AND c = 1');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 26, 51, 76]) AND b = ANY (ARRAY[''1'', ''26'']) AND c = ANY (ARRAY[1])');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = ANY (ARRAY[1, 2, 26, 27, 51, 52, 76, 77]) AND b = ANY (ARRAY[''1'', ''2'', ''26'', ''27'']) AND c = ANY (ARRAY[1, 2])');`,
				Results:   []sql.Row{{400, 400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a < ANY (ARRAY[1, 51]) AND b > ''1''');`,
				Results:   []sql.Row{{2472, 2400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a >= ANY (ARRAY[1, 51]) AND b <= ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1441, 1250}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a <= ANY (ARRAY[1, 2, 51, 52]) AND b >= ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{3909, 2550}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 51) AND b = ALL (ARRAY[''1''])');`,
				Results:   []sql.Row{{2, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 51) AND b = ALL (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a IN (1, 2, 51, 52) AND b = ALL (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `ALTER TABLE functional_dependencies ALTER COLUMN c TYPE numeric;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `ANALYZE functional_dependencies;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `DROP STATISTICS func_deps_stat;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = 2 AND upper(b) = ''1''');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = 2 AND upper(b) = ''1'' AND (c + 1) = 2');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 102) AND upper(b) = ''1''');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 102) AND upper(b) IN (''1'', ''2'')');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 4, 102, 104) AND upper(b) IN (''1'', ''2'')');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 4, 102, 104) AND upper(b) = ''1''');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 52, 102, 152) AND upper(b) IN (''1'', ''26'') AND (c + 1) = 2');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 52, 102, 152) AND upper(b) IN (''1'', ''26'') AND (c + 1) IN (2)');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 4, 52, 54, 102, 104, 152, 154) AND upper(b) IN (''1'', ''2'', ''26'', ''27'') AND (c + 1) IN (2, 3)');`,
				Results:   []sql.Row{{1, 400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE ((a * 2) = 2 OR (a * 2) = 102) AND upper(b) = ''1''');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE ((a * 2) = 2 OR (a * 2) = 102) AND (upper(b) = ''1'' OR upper(b) = ''2'')');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE ((a * 2) = 2 OR (a * 2) = 4 OR (a * 2) = 102 OR (a * 2) = 104) AND (upper(b) = ''1'' OR upper(b) = ''2'')');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE ((a * 2) = 2 OR upper(b) = ''1'') AND upper(b) = ''1''');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 102]) AND upper(b) = ''1''');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 102]) AND upper(b) = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 4, 102, 104]) AND upper(b) = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 52, 102, 152]) AND upper(b) = ANY (ARRAY[''1'', ''26'']) AND (c + 1) = 2');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 52, 102, 152]) AND upper(b) = ANY (ARRAY[''1'', ''26'']) AND (c + 1) = ANY (ARRAY[2])');`,
				Results:   []sql.Row{{1, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 4, 52, 54, 102, 104, 152, 154]) AND upper(b) = ANY (ARRAY[''1'', ''2'', ''26'', ''27'']) AND (c + 1) = ANY (ARRAY[2, 3])');`,
				Results:   []sql.Row{{1, 400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) < ANY (ARRAY[2, 102]) AND upper(b) > ''1''');`,
				Results:   []sql.Row{{926, 2400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) >= ANY (ARRAY[2, 102]) AND upper(b) <= ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1543, 1250}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) <= ANY (ARRAY[2, 4, 102, 104]) AND upper(b) >= ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{2229, 2550}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 102) AND upper(b) = ALL (ARRAY[''1''])');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 102) AND upper(b) = ALL (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 4, 102, 104) AND upper(b) = ALL (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `CREATE STATISTICS func_deps_stat (dependencies) ON (a * 2), upper(b), (c + 1) FROM functional_dependencies;`,
			},
			{
				Statement: `ANALYZE functional_dependencies;`,
			},
			{
				Statement: `SELECT dependencies FROM pg_stats_ext WHERE statistics_name = 'func_deps_stat';`,
				Results:   []sql.Row{{`{"-1 => -2": 1.000000, "-1 => -3": 1.000000, "-2 => -3": 1.000000, "-1, -2 => -3": 1.000000, "-1, -3 => -2": 1.000000}`}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = 2 AND upper(b) = ''1''');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = 2 AND upper(b) = ''1'' AND (c + 1) = 2');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 102) AND upper(b) = ''1''');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 102) AND upper(b) IN (''1'', ''2'')');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 4, 102, 104) AND upper(b) IN (''1'', ''2'')');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 4, 102, 104) AND upper(b) = ''1''');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 52, 102, 152) AND upper(b) IN (''1'', ''26'') AND (c + 1) = 2');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 52, 102, 152) AND upper(b) IN (''1'', ''26'') AND (c + 1) IN (2)');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 4, 52, 54, 102, 104, 152, 154) AND upper(b) IN (''1'', ''2'', ''26'', ''27'') AND (c + 1) IN (2, 3)');`,
				Results:   []sql.Row{{400, 400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE ((a * 2) = 2 OR (a * 2) = 102) AND upper(b) = ''1''');`,
				Results:   []sql.Row{{99, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE ((a * 2) = 2 OR (a * 2) = 102) AND (upper(b) = ''1'' OR upper(b) = ''2'')');`,
				Results:   []sql.Row{{99, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE ((a * 2) = 2 OR (a * 2) = 4 OR (a * 2) = 102 OR (a * 2) = 104) AND (upper(b) = ''1'' OR upper(b) = ''2'')');`,
				Results:   []sql.Row{{197, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE ((a * 2) = 2 OR upper(b) = ''1'') AND upper(b) = ''1''');`,
				Results:   []sql.Row{{3, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 102]) AND upper(b) = ''1''');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 102]) AND upper(b) = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 4, 102, 104]) AND upper(b) = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 52, 102, 152]) AND upper(b) = ANY (ARRAY[''1'', ''26'']) AND (c + 1) = 2');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 52, 102, 152]) AND upper(b) = ANY (ARRAY[''1'', ''26'']) AND (c + 1) = ANY (ARRAY[2])');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) = ANY (ARRAY[2, 4, 52, 54, 102, 104, 152, 154]) AND upper(b) = ANY (ARRAY[''1'', ''2'', ''26'', ''27'']) AND (c + 1) = ANY (ARRAY[2, 3])');`,
				Results:   []sql.Row{{400, 400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) < ANY (ARRAY[2, 102]) AND upper(b) > ''1''');`,
				Results:   []sql.Row{{2472, 2400}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) >= ANY (ARRAY[2, 102]) AND upper(b) <= ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1441, 1250}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) <= ANY (ARRAY[2, 4, 102, 104]) AND upper(b) >= ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{3909, 2550}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 102) AND upper(b) = ALL (ARRAY[''1''])');`,
				Results:   []sql.Row{{2, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 102) AND upper(b) = ALL (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies WHERE (a * 2) IN (2, 4, 102, 104) AND upper(b) = ALL (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `CREATE TABLE functional_dependencies_multi (
	a INTEGER,
	b INTEGER,
	c INTEGER,
	d INTEGER
)
WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO functional_dependencies_multi (a, b, c, d)
    SELECT
         mod(i,7),
         mod(i,7),
         mod(i,11),
         mod(i,11)
    FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE functional_dependencies_multi;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE a = 0 AND b = 0');`,
				Results:   []sql.Row{{102, 714}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE 0 = a AND 0 = b');`,
				Results:   []sql.Row{{102, 714}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE c = 0 AND d = 0');`,
				Results:   []sql.Row{{41, 454}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE a = 0 AND b = 0 AND c = 0 AND d = 0');`,
				Results:   []sql.Row{{1, 64}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE 0 = a AND b = 0 AND 0 = c AND d = 0');`,
				Results:   []sql.Row{{1, 64}},
			},
			{
				Statement: `CREATE STATISTICS functional_dependencies_multi_1 (dependencies) ON a, b FROM functional_dependencies_multi;`,
			},
			{
				Statement: `CREATE STATISTICS functional_dependencies_multi_2 (dependencies) ON c, d FROM functional_dependencies_multi;`,
			},
			{
				Statement: `ANALYZE functional_dependencies_multi;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE a = 0 AND b = 0');`,
				Results:   []sql.Row{{714, 714}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE 0 = a AND 0 = b');`,
				Results:   []sql.Row{{714, 714}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE c = 0 AND d = 0');`,
				Results:   []sql.Row{{454, 454}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE a = 0 AND b = 0 AND c = 0 AND d = 0');`,
				Results:   []sql.Row{{65, 64}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM functional_dependencies_multi WHERE 0 = a AND b = 0 AND 0 = c AND d = 0');`,
				Results:   []sql.Row{{65, 64}},
			},
			{
				Statement: `DROP TABLE functional_dependencies_multi;`,
			},
			{
				Statement: `CREATE TABLE mcv_lists (
    filler1 TEXT,
    filler2 NUMERIC,
    a INT,
    b VARCHAR,
    filler3 DATE,
    c INT,
    d TEXT,
    ia INT[]
)
WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO mcv_lists (a, b, c, filler1)
     SELECT mod(i,37), mod(i,41), mod(i,43), mod(i,47) FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{3, 4}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{1, 1}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats (mcv) ON a, b, c FROM mcv_lists;`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{3, 4}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{1, 1}},
			},
			{
				Statement: `TRUNCATE mcv_lists;`,
			},
			{
				Statement: `DROP STATISTICS mcv_lists_stats;`,
			},
			{
				Statement: `INSERT INTO mcv_lists (a, b, c, filler1)
     SELECT i, i, i, i FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,7) = 1 AND mod(b::int,11) = 1');`,
				Results:   []sql.Row{{1, 13}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,7) = 1 AND mod(b::int,11) = 1 AND mod(c,13) = 1');`,
				Results:   []sql.Row{{1, 1}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats (mcv) ON (mod(a,7)), (mod(b::int,11)), (mod(c,13)) FROM mcv_lists;`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,7) = 1 AND mod(b::int,11) = 1');`,
				Results:   []sql.Row{{13, 13}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,7) = 1 AND mod(b::int,11) = 1 AND mod(c,13) = 1');`,
				Results:   []sql.Row{{1, 1}},
			},
			{
				Statement: `TRUNCATE mcv_lists;`,
			},
			{
				Statement: `DROP STATISTICS mcv_lists_stats;`,
			},
			{
				Statement: `INSERT INTO mcv_lists (a, b, c, ia, filler1)
     SELECT mod(i,100), mod(i,50), mod(i,25), array[mod(i,25)], i
       FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 = a AND ''1'' = b');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < 1 AND b < ''1''');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 > a AND ''1'' > b');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a <= 0 AND b <= ''0''');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 0 >= a AND ''0'' >= b');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < 5 AND b < ''1'' AND c < 5');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < 5 AND ''1'' > b AND 5 > c');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a <= 4 AND b <= ''0'' AND c <= 4');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 4 >= a AND ''0'' >= b AND 4 >= c');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 OR b = ''1'' OR c = 1');`,
				Results:   []sql.Row{{343, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 OR b = ''1'' OR c = 1 OR d IS NOT NULL');`,
				Results:   []sql.Row{{343, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IN (1, 2, 51, 52) AND b IN ( ''1'', ''2'')');`,
				Results:   []sql.Row{{8, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IN (1, 2, 51, 52, NULL) AND b IN ( ''1'', ''2'', NULL)');`,
				Results:   []sql.Row{{8, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = ANY (ARRAY[1, 2, 51, 52]) AND b = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{8, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = ANY (ARRAY[NULL, 1, 2, 51, 52]) AND b = ANY (ARRAY[''1'', ''2'', NULL])');`,
				Results:   []sql.Row{{8, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a <= ANY (ARRAY[1, 2, 3]) AND b IN (''1'', ''2'', ''3'')');`,
				Results:   []sql.Row{{26, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a <= ANY (ARRAY[1, NULL, 2, 3]) AND b IN (''1'', ''2'', NULL, ''3'')');`,
				Results:   []sql.Row{{26, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < ALL (ARRAY[4, 5]) AND c > ANY (ARRAY[1, 2, 3])');`,
				Results:   []sql.Row{{10, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < ALL (ARRAY[4, 5]) AND c > ANY (ARRAY[1, 2, 3, NULL])');`,
				Results:   []sql.Row{{10, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < ALL (ARRAY[4, 5]) AND b IN (''1'', ''2'', ''3'') AND c > ANY (ARRAY[1, 2, 3])');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < ALL (ARRAY[4, 5]) AND b IN (''1'', ''2'', NULL, ''3'') AND c > ANY (ARRAY[1, 2, NULL, 3])');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = ANY (ARRAY[4,5]) AND 4 = ANY(ia)');`,
				Results:   []sql.Row{{4, 50}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats (mcv) ON a, b, c, ia FROM mcv_lists;`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 = a AND ''1'' = b');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < 1 AND b < ''1''');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 > a AND ''1'' > b');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a <= 0 AND b <= ''0''');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 0 >= a AND ''0'' >= b');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1'' AND c = 1');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < 5 AND b < ''1'' AND c < 5');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < 5 AND ''1'' > b AND 5 > c');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a <= 4 AND b <= ''0'' AND c <= 4');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 4 >= a AND ''0'' >= b AND 4 >= c');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 OR b = ''1'' OR c = 1');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 OR b = ''1'' OR c = 1 OR d IS NOT NULL');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 OR b = ''1'' OR c = 1 OR d IS NOT NULL');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IN (1, 2, 51, 52) AND b IN ( ''1'', ''2'')');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IN (1, 2, 51, 52, NULL) AND b IN ( ''1'', ''2'', NULL)');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = ANY (ARRAY[1, 2, 51, 52]) AND b = ANY (ARRAY[''1'', ''2''])');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = ANY (ARRAY[NULL, 1, 2, 51, 52]) AND b = ANY (ARRAY[''1'', ''2'', NULL])');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a <= ANY (ARRAY[1, 2, 3]) AND b IN (''1'', ''2'', ''3'')');`,
				Results:   []sql.Row{{150, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a <= ANY (ARRAY[1, NULL, 2, 3]) AND b IN (''1'', ''2'', NULL, ''3'')');`,
				Results:   []sql.Row{{150, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < ALL (ARRAY[4, 5]) AND c > ANY (ARRAY[1, 2, 3])');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < ALL (ARRAY[4, 5]) AND c > ANY (ARRAY[1, 2, 3, NULL])');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < ALL (ARRAY[4, 5]) AND b IN (''1'', ''2'', ''3'') AND c > ANY (ARRAY[1, 2, 3])');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a < ALL (ARRAY[4, 5]) AND b IN (''1'', ''2'', NULL, ''3'') AND c > ANY (ARRAY[1, 2, NULL, 3])');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = ANY (ARRAY[4,5]) AND 4 = ANY(ia)');`,
				Results:   []sql.Row{{4, 50}},
			},
			{
				Statement: `ALTER TABLE mcv_lists ALTER COLUMN d TYPE VARCHAR(64);`,
			},
			{
				Statement: `SELECT d.stxdmcv IS NOT NULL
  FROM pg_statistic_ext s, pg_statistic_ext_data d
 WHERE s.stxname = 'mcv_lists_stats'
   AND d.stxoid = s.oid;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `ALTER TABLE mcv_lists ALTER COLUMN c TYPE numeric;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 AND b = ''1''');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `TRUNCATE mcv_lists;`,
			},
			{
				Statement: `DROP STATISTICS mcv_lists_stats;`,
			},
			{
				Statement: `INSERT INTO mcv_lists (a, b, c, filler1)
     SELECT i, i, i, i FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 AND mod(b::int,10) = 1');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 = mod(a,20) AND 1 = mod(b::int,10)');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) < 1 AND mod(b::int,10) < 1');`,
				Results:   []sql.Row{{111, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 > mod(a,20) AND 1 > mod(b::int,10)');`,
				Results:   []sql.Row{{111, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 AND mod(b::int,10) = 1 AND mod(c,5) = 1');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 OR mod(b::int,10) = 1 OR mod(c,25) = 1 OR d IS NOT NULL');`,
				Results:   []sql.Row{{15, 120}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) IN (1, 2, 51, 52, NULL) AND mod(b::int,10) IN ( 1, 2, NULL)');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = ANY (ARRAY[1, 2, 51, 52]) AND mod(b::int,10) = ANY (ARRAY[1, 2])');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) <= ANY (ARRAY[1, NULL, 2, 3]) AND mod(b::int,10) IN (1, 2, NULL, 3)');`,
				Results:   []sql.Row{{11, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) < ALL (ARRAY[4, 5]) AND mod(b::int,10) IN (1, 2, 3) AND mod(c,5) > ANY (ARRAY[1, 2, 3])');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats_1 ON (mod(a,20)) FROM mcv_lists;`,
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats_2 ON (mod(b::int,10)) FROM mcv_lists;`,
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats_3 ON (mod(c,5)) FROM mcv_lists;`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 AND mod(b::int,10) = 1');`,
				Results:   []sql.Row{{5, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 = mod(a,20) AND 1 = mod(b::int,10)');`,
				Results:   []sql.Row{{5, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) < 1 AND mod(b::int,10) < 1');`,
				Results:   []sql.Row{{5, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 > mod(a,20) AND 1 > mod(b::int,10)');`,
				Results:   []sql.Row{{5, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 AND mod(b::int,10) = 1 AND mod(c,5) = 1');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 OR mod(b::int,10) = 1 OR mod(c,25) = 1 OR d IS NOT NULL');`,
				Results:   []sql.Row{{149, 120}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) IN (1, 2, 51, 52, NULL) AND mod(b::int,10) IN ( 1, 2, NULL)');`,
				Results:   []sql.Row{{20, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = ANY (ARRAY[1, 2, 51, 52]) AND mod(b::int,10) = ANY (ARRAY[1, 2])');`,
				Results:   []sql.Row{{20, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) <= ANY (ARRAY[1, NULL, 2, 3]) AND mod(b::int,10) IN (1, 2, NULL, 3)');`,
				Results:   []sql.Row{{116, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) < ALL (ARRAY[4, 5]) AND mod(b::int,10) IN (1, 2, 3) AND mod(c,5) > ANY (ARRAY[1, 2, 3])');`,
				Results:   []sql.Row{{12, 100}},
			},
			{
				Statement: `DROP STATISTICS mcv_lists_stats_1;`,
			},
			{
				Statement: `DROP STATISTICS mcv_lists_stats_2;`,
			},
			{
				Statement: `DROP STATISTICS mcv_lists_stats_3;`,
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats (mcv) ON (mod(a,20)), (mod(b::int,10)), (mod(c,5)) FROM mcv_lists;`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 AND mod(b::int,10) = 1');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 = mod(a,20) AND 1 = mod(b::int,10)');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) < 1 AND mod(b::int,10) < 1');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE 1 > mod(a,20) AND 1 > mod(b::int,10)');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 AND mod(b::int,10) = 1 AND mod(c,5) = 1');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 OR mod(b::int,10) = 1 OR mod(c,25) = 1 OR d IS NOT NULL');`,
				Results:   []sql.Row{{105, 120}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) IN (1, 2, 51, 52, NULL) AND mod(b::int,10) IN ( 1, 2, NULL)');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = ANY (ARRAY[1, 2, 51, 52]) AND mod(b::int,10) = ANY (ARRAY[1, 2])');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) <= ANY (ARRAY[1, NULL, 2, 3]) AND mod(b::int,10) IN (1, 2, NULL, 3)');`,
				Results:   []sql.Row{{150, 150}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) < ALL (ARRAY[4, 5]) AND mod(b::int,10) IN (1, 2, 3) AND mod(c,5) > ANY (ARRAY[1, 2, 3])');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE mod(a,20) = 1 OR mod(b::int,10) = 1 OR mod(c,5) = 1 OR d IS NOT NULL');`,
				Results:   []sql.Row{{200, 200}},
			},
			{
				Statement: `TRUNCATE mcv_lists;`,
			},
			{
				Statement: `DROP STATISTICS mcv_lists_stats;`,
			},
			{
				Statement: `INSERT INTO mcv_lists (a, b, c, filler1)
     SELECT
         (CASE WHEN mod(i,100) = 1 THEN NULL ELSE mod(i,100) END),
         (CASE WHEN mod(i,50) = 1  THEN NULL ELSE mod(i,50) END),
         (CASE WHEN mod(i,25) = 1  THEN NULL ELSE mod(i,25) END),
         i
     FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NULL AND b IS NULL');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NULL AND b IS NULL AND c IS NULL');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NULL AND b IS NOT NULL');`,
				Results:   []sql.Row{{49, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NOT NULL AND b IS NULL AND c IS NOT NULL');`,
				Results:   []sql.Row{{95, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IN (0, 1) AND b IN (''0'', ''1'')');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats (mcv) ON a, b, c FROM mcv_lists;`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NULL AND b IS NULL');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NULL AND b IS NULL AND c IS NULL');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NULL AND b IS NOT NULL');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NOT NULL AND b IS NULL AND c IS NOT NULL');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IN (0, 1) AND b IN (''0'', ''1'')');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `TRUNCATE mcv_lists;`,
			},
			{
				Statement: `INSERT INTO mcv_lists (a, b, c) SELECT 1, 2, 3 FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT m.*
  FROM pg_statistic_ext s, pg_statistic_ext_data d,
       pg_mcv_list_items(d.stxdmcv) m
 WHERE s.stxname = 'mcv_lists_stats'
   AND d.stxoid = s.oid;`,
				Results: []sql.Row{{0, `{1,2,3}`, `{f,f,f}`, 1, 1}},
			},
			{
				Statement: `TRUNCATE mcv_lists;`,
			},
			{
				Statement: `DROP STATISTICS mcv_lists_stats;`,
			},
			{
				Statement: `INSERT INTO mcv_lists (a, b, c, d)
     SELECT
         NULL, -- always NULL
         (CASE WHEN mod(i,2) = 0 THEN NULL ELSE 'x' END),
         (CASE WHEN mod(i,2) = 0 THEN NULL ELSE 0 END),
         (CASE WHEN mod(i,2) = 0 THEN NULL ELSE 'x' END)
     FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE b = ''x'' OR d = ''x''');`,
				Results:   []sql.Row{{3750, 2500}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 OR b = ''x'' OR d = ''x''');`,
				Results:   []sql.Row{{3750, 2500}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NULL AND (b = ''x'' OR d = ''x'')');`,
				Results:   []sql.Row{{3750, 2500}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_stats (mcv) ON a, b, d FROM mcv_lists;`,
			},
			{
				Statement: `ANALYZE mcv_lists;`,
			},
			{
				Statement: `SELECT m.*
  FROM pg_statistic_ext s, pg_statistic_ext_data d,
       pg_mcv_list_items(d.stxdmcv) m
 WHERE s.stxname = 'mcv_lists_stats'
   AND d.stxoid = s.oid;`,
				Results: []sql.Row{{0, `{NULL,x,x}`, `{t,f,f}`, 0.5, 0.25}, {1, `{NULL,NULL,NULL}`, `{t,t,t}`, 0.5, 0.25}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE b = ''x'' OR d = ''x''');`,
				Results:   []sql.Row{{2500, 2500}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a = 1 OR b = ''x'' OR d = ''x''');`,
				Results:   []sql.Row{{2500, 2500}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists WHERE a IS NULL AND (b = ''x'' OR d = ''x'')');`,
				Results:   []sql.Row{{2500, 2500}},
			},
			{
				Statement: `CREATE TABLE mcv_lists_uuid (
    a UUID,
    b UUID,
    c UUID
)
WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO mcv_lists_uuid (a, b, c)
     SELECT
         md5(mod(i,100)::text)::uuid,
         md5(mod(i,50)::text)::uuid,
         md5(mod(i,25)::text)::uuid
     FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists_uuid;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_uuid WHERE a = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc'' AND b = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc''');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_uuid WHERE a = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc'' AND b = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc'' AND c = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc''');`,
				Results:   []sql.Row{{1, 50}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_uuid_stats (mcv) ON a, b, c
  FROM mcv_lists_uuid;`,
			},
			{
				Statement: `ANALYZE mcv_lists_uuid;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_uuid WHERE a = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc'' AND b = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc''');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_uuid WHERE a = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc'' AND b = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc'' AND c = ''1679091c-5a88-0faf-6fb5-e6087eb1b2dc''');`,
				Results:   []sql.Row{{50, 50}},
			},
			{
				Statement: `DROP TABLE mcv_lists_uuid;`,
			},
			{
				Statement: `CREATE TABLE mcv_lists_arrays (
    a TEXT[],
    b NUMERIC[],
    c INT[]
)
WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO mcv_lists_arrays (a, b, c)
     SELECT
         ARRAY[md5((i/100)::text), md5((i/100-1)::text), md5((i/100+1)::text)],
         ARRAY[(i/100-1)::numeric/1000, (i/100)::numeric/1000, (i/100+1)::numeric/1000],
         ARRAY[(i/100-1), i/100, (i/100+1)]
     FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_arrays_stats (mcv) ON a, b, c
  FROM mcv_lists_arrays;`,
			},
			{
				Statement: `ANALYZE mcv_lists_arrays;`,
			},
			{
				Statement: `CREATE TABLE mcv_lists_bool (
    a BOOL,
    b BOOL,
    c BOOL
)
WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO mcv_lists_bool (a, b, c)
     SELECT
         (mod(i,2) = 0), (mod(i,4) = 0), (mod(i,8) = 0)
     FROM generate_series(1,10000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists_bool;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_bool WHERE a AND b AND c');`,
				Results:   []sql.Row{{156, 1250}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_bool WHERE NOT a AND b AND c');`,
				Results:   []sql.Row{{156, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_bool WHERE NOT a AND NOT b AND c');`,
				Results:   []sql.Row{{469, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_bool WHERE NOT a AND b AND NOT c');`,
				Results:   []sql.Row{{1094, 0}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_bool_stats (mcv) ON a, b, c
  FROM mcv_lists_bool;`,
			},
			{
				Statement: `ANALYZE mcv_lists_bool;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_bool WHERE a AND b AND c');`,
				Results:   []sql.Row{{1250, 1250}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_bool WHERE NOT a AND b AND c');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_bool WHERE NOT a AND NOT b AND c');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_bool WHERE NOT a AND b AND NOT c');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `CREATE TABLE mcv_lists_partial (
    a INT,
    b INT,
    c INT
);`,
			},
			{
				Statement: `INSERT INTO mcv_lists_partial (a, b, c)
     SELECT
         mod(i,10),
         mod(i,10),
         mod(i,10)
     FROM generate_series(0,999) s(i);`,
			},
			{
				Statement: `INSERT INTO mcv_lists_partial (a, b, c)
     SELECT
         i,
         i,
         i
     FROM generate_series(0,99) s(i);`,
			},
			{
				Statement: `INSERT INTO mcv_lists_partial (a, b, c)
     SELECT
         i,
         i,
         i
     FROM generate_series(0,3999) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists_partial;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 0 AND b = 0 AND c = 0');`,
				Results:   []sql.Row{{1, 102}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 0 OR b = 0 OR c = 0');`,
				Results:   []sql.Row{{300, 102}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 10 AND b = 10 AND c = 10');`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 10 OR b = 10 OR c = 10');`,
				Results:   []sql.Row{{6, 2}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 0 AND b = 0 AND c = 10');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 0 OR b = 0 OR c = 10');`,
				Results:   []sql.Row{{204, 104}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE (a = 0 AND b = 0 AND c = 0) OR (a = 1 AND b = 1 AND c = 1) OR (a = 2 AND b = 2 AND c = 2)');`,
				Results:   []sql.Row{{1, 306}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE (a = 0 AND b = 0) OR (a = 0 AND c = 0) OR (b = 0 AND c = 0)');`,
				Results:   []sql.Row{{6, 102}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_partial_stats (mcv) ON a, b, c
  FROM mcv_lists_partial;`,
			},
			{
				Statement: `ANALYZE mcv_lists_partial;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 0 AND b = 0 AND c = 0');`,
				Results:   []sql.Row{{102, 102}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 0 OR b = 0 OR c = 0');`,
				Results:   []sql.Row{{96, 102}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 10 AND b = 10 AND c = 10');`,
				Results:   []sql.Row{{2, 2}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 10 OR b = 10 OR c = 10');`,
				Results:   []sql.Row{{2, 2}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 0 AND b = 0 AND c = 10');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE a = 0 OR b = 0 OR c = 10');`,
				Results:   []sql.Row{{102, 104}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE (a = 0 AND b = 0 AND c = 0) OR (a = 1 AND b = 1 AND c = 1) OR (a = 2 AND b = 2 AND c = 2)');`,
				Results:   []sql.Row{{306, 306}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_partial WHERE (a = 0 AND b = 0) OR (a = 0 AND c = 0) OR (b = 0 AND c = 0)');`,
				Results:   []sql.Row{{108, 102}},
			},
			{
				Statement: `DROP TABLE mcv_lists_partial;`,
			},
			{
				Statement: `CREATE TABLE mcv_lists_multi (
	a INTEGER,
	b INTEGER,
	c INTEGER,
	d INTEGER
)
WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO mcv_lists_multi (a, b, c, d)
    SELECT
         mod(i,5),
         mod(i,5),
         mod(i,7),
         mod(i,7)
    FROM generate_series(1,5000) s(i);`,
			},
			{
				Statement: `ANALYZE mcv_lists_multi;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE a = 0 AND b = 0');`,
				Results:   []sql.Row{{200, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE c = 0 AND d = 0');`,
				Results:   []sql.Row{{102, 714}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE b = 0 AND c = 0');`,
				Results:   []sql.Row{{143, 142}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE b = 0 OR c = 0');`,
				Results:   []sql.Row{{1571, 1572}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE a = 0 AND b = 0 AND c = 0 AND d = 0');`,
				Results:   []sql.Row{{4, 142}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE (a = 0 AND b = 0) OR (c = 0 AND d = 0)');`,
				Results:   []sql.Row{{298, 1572}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE a = 0 OR b = 0 OR c = 0 OR d = 0');`,
				Results:   []sql.Row{{2649, 1572}},
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_multi_1 (mcv) ON a, b FROM mcv_lists_multi;`,
			},
			{
				Statement: `CREATE STATISTICS mcv_lists_multi_2 (mcv) ON c, d FROM mcv_lists_multi;`,
			},
			{
				Statement: `ANALYZE mcv_lists_multi;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE a = 0 AND b = 0');`,
				Results:   []sql.Row{{1000, 1000}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE c = 0 AND d = 0');`,
				Results:   []sql.Row{{714, 714}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE b = 0 AND c = 0');`,
				Results:   []sql.Row{{143, 142}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE b = 0 OR c = 0');`,
				Results:   []sql.Row{{1571, 1572}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE a = 0 AND b = 0 AND c = 0 AND d = 0');`,
				Results:   []sql.Row{{143, 142}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE (a = 0 AND b = 0) OR (c = 0 AND d = 0)');`,
				Results:   []sql.Row{{1571, 1572}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM mcv_lists_multi WHERE a = 0 OR b = 0 OR c = 0 OR d = 0');`,
				Results:   []sql.Row{{1571, 1572}},
			},
			{
				Statement: `DROP TABLE mcv_lists_multi;`,
			},
			{
				Statement: `CREATE TABLE expr_stats (a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO expr_stats SELECT mod(i,10), mod(i,10), mod(i,10) FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE expr_stats;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE (2*a) = 0 AND (3*b) = 0');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE (a+b) = 0 AND (a-b) = 0');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `CREATE STATISTICS expr_stats_1 (mcv) ON (a+b), (a-b), (2*a), (3*b) FROM expr_stats;`,
			},
			{
				Statement: `ANALYZE expr_stats;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE (2*a) = 0 AND (3*b) = 0');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE (a+b) = 0 AND (a-b) = 0');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `DROP STATISTICS expr_stats_1;`,
			},
			{
				Statement: `DROP TABLE expr_stats;`,
			},
			{
				Statement: `CREATE TABLE expr_stats (a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO expr_stats SELECT mod(i,10), mod(i,10), mod(i,10) FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE expr_stats;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE a = 0 AND (2*a) = 0 AND (3*b) = 0');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE a = 3 AND b = 3 AND (a-b) = 0');`,
				Results:   []sql.Row{{1, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE a = 0 AND b = 1 AND (a-b) = 0');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `CREATE STATISTICS expr_stats_1 (mcv) ON a, b, (2*a), (3*b), (a+b), (a-b) FROM expr_stats;`,
			},
			{
				Statement: `ANALYZE expr_stats;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE a = 0 AND (2*a) = 0 AND (3*b) = 0');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE a = 3 AND b = 3 AND (a-b) = 0');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE a = 0 AND b = 1 AND (a-b) = 0');`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `DROP TABLE expr_stats;`,
			},
			{
				Statement: `CREATE TABLE expr_stats (a int, b name, c text);`,
			},
			{
				Statement: `INSERT INTO expr_stats SELECT mod(i,10), md5(mod(i,10)::text), md5(mod(i,10)::text) FROM generate_series(1,1000) s(i);`,
			},
			{
				Statement: `ANALYZE expr_stats;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE a = 0 AND (b || c) <= ''z'' AND (c || b) >= ''0''');`,
				Results:   []sql.Row{{11, 100}},
			},
			{
				Statement: `CREATE STATISTICS expr_stats_1 (mcv) ON a, b, (b || c), (c || b) FROM expr_stats;`,
			},
			{
				Statement: `ANALYZE expr_stats;`,
			},
			{
				Statement: `SELECT * FROM check_estimated_rows('SELECT * FROM expr_stats WHERE a = 0 AND (b || c) <= ''z'' AND (c || b) >= ''0''');`,
				Results:   []sql.Row{{100, 100}},
			},
			{
				Statement: `DROP TABLE expr_stats;`,
			},
			{
				Statement: `CREATE TABLE expr_stats_incompatible_test (
    c0 double precision,
    c1 boolean NOT NULL
);`,
			},
			{
				Statement: `CREATE STATISTICS expr_stat_comp_1 ON c0, c1 FROM expr_stats_incompatible_test;`,
			},
			{
				Statement: `INSERT INTO expr_stats_incompatible_test VALUES (1234,false), (5678,true);`,
			},
			{
				Statement: `ANALYZE expr_stats_incompatible_test;`,
			},
			{
				Statement: `SELECT c0 FROM ONLY expr_stats_incompatible_test WHERE
(
  upper('x') LIKE ('x'||('[0,1]'::int4range))
  AND
  (c0 IN (0, 1) OR c1)
);`,
				Results: []sql.Row{},
			},
			{
				Statement: `DROP TABLE expr_stats_incompatible_test;`,
			},
			{
				Statement: `CREATE SCHEMA tststats;`,
			},
			{
				Statement: `CREATE TABLE tststats.priv_test_tbl (
    a int,
    b int
);`,
			},
			{
				Statement: `INSERT INTO tststats.priv_test_tbl
     SELECT mod(i,5), mod(i,10) FROM generate_series(1,100) s(i);`,
			},
			{
				Statement: `CREATE STATISTICS tststats.priv_test_stats (mcv) ON a, b
  FROM tststats.priv_test_tbl;`,
			},
			{
				Statement: `ANALYZE tststats.priv_test_tbl;`,
			},
			{
				Statement: `create table stts_t1 (a int, b int);`,
			},
			{
				Statement: `create statistics stts_1 (ndistinct) on a, b from stts_t1;`,
			},
			{
				Statement: `create statistics stts_2 (ndistinct, dependencies) on a, b from stts_t1;`,
			},
			{
				Statement: `create statistics stts_3 (ndistinct, dependencies, mcv) on a, b from stts_t1;`,
			},
			{
				Statement: `create table stts_t2 (a int, b int, c int);`,
			},
			{
				Statement: `create statistics stts_4 on b, c from stts_t2;`,
			},
			{
				Statement: `create table stts_t3 (col1 int, col2 int, col3 int);`,
			},
			{
				Statement: `create statistics stts_hoge on col1, col2, col3 from stts_t3;`,
			},
			{
				Statement: `create schema stts_s1;`,
			},
			{
				Statement: `create schema stts_s2;`,
			},
			{
				Statement: `create statistics stts_s1.stts_foo on col1, col2 from stts_t3;`,
			},
			{
				Statement: `create statistics stts_s2.stts_yama (dependencies, mcv) on col1, col3 from stts_t3;`,
			},
			{
				Statement: `insert into stts_t1 select i,i from generate_series(1,100) i;`,
			},
			{
				Statement: `analyze stts_t1;`,
			},
			{
				Statement: `set search_path to public, stts_s1, stts_s2, tststats;`,
			},
			{
				Statement: `\dX
                                                        List of extended statistics
  Schema  |          Name          |                            Definition                            | Ndistinct | Dependencies |   MCV   
----------+------------------------+------------------------------------------------------------------+-----------+--------------+---------
 public   | func_deps_stat         | (a * 2), upper(b), (c + 1::numeric) FROM functional_dependencies |           | defined      | 
 public   | mcv_lists_arrays_stats | a, b, c FROM mcv_lists_arrays                                    |           |              | defined
 public   | mcv_lists_bool_stats   | a, b, c FROM mcv_lists_bool                                      |           |              | defined
 public   | mcv_lists_stats        | a, b, d FROM mcv_lists                                           |           |              | defined
 public   | stts_1                 | a, b FROM stts_t1                                                | defined   |              | 
 public   | stts_2                 | a, b FROM stts_t1                                                | defined   | defined      | 
 public   | stts_3                 | a, b FROM stts_t1                                                | defined   | defined      | defined
 public   | stts_4                 | b, c FROM stts_t2                                                | defined   | defined      | defined
 public   | stts_hoge              | col1, col2, col3 FROM stts_t3                                    | defined   | defined      | defined
 stts_s1  | stts_foo               | col1, col2 FROM stts_t3                                          | defined   | defined      | defined
 stts_s2  | stts_yama              | col1, col3 FROM stts_t3                                          |           | defined      | defined
 tststats | priv_test_stats        | a, b FROM priv_test_tbl                                          |           |              | defined
(12 rows)
\dX stts_?
                       List of extended statistics
 Schema |  Name  |    Definition     | Ndistinct | Dependencies |   MCV   
--------+--------+-------------------+-----------+--------------+---------
 public | stts_1 | a, b FROM stts_t1 | defined   |              | 
 public | stts_2 | a, b FROM stts_t1 | defined   | defined      | 
 public | stts_3 | a, b FROM stts_t1 | defined   | defined      | defined
 public | stts_4 | b, c FROM stts_t2 | defined   | defined      | defined
(4 rows)
\dX *stts_hoge
                               List of extended statistics
 Schema |   Name    |          Definition           | Ndistinct | Dependencies |   MCV   
--------+-----------+-------------------------------+-----------+--------------+---------
 public | stts_hoge | col1, col2, col3 FROM stts_t3 | defined   | defined      | defined
(1 row)
\dX+
                                                        List of extended statistics
  Schema  |          Name          |                            Definition                            | Ndistinct | Dependencies |   MCV   
----------+------------------------+------------------------------------------------------------------+-----------+--------------+---------
 public   | func_deps_stat         | (a * 2), upper(b), (c + 1::numeric) FROM functional_dependencies |           | defined      | 
 public   | mcv_lists_arrays_stats | a, b, c FROM mcv_lists_arrays                                    |           |              | defined
 public   | mcv_lists_bool_stats   | a, b, c FROM mcv_lists_bool                                      |           |              | defined
 public   | mcv_lists_stats        | a, b, d FROM mcv_lists                                           |           |              | defined
 public   | stts_1                 | a, b FROM stts_t1                                                | defined   |              | 
 public   | stts_2                 | a, b FROM stts_t1                                                | defined   | defined      | 
 public   | stts_3                 | a, b FROM stts_t1                                                | defined   | defined      | defined
 public   | stts_4                 | b, c FROM stts_t2                                                | defined   | defined      | defined
 public   | stts_hoge              | col1, col2, col3 FROM stts_t3                                    | defined   | defined      | defined
 stts_s1  | stts_foo               | col1, col2 FROM stts_t3                                          | defined   | defined      | defined
 stts_s2  | stts_yama              | col1, col3 FROM stts_t3                                          |           | defined      | defined
 tststats | priv_test_stats        | a, b FROM priv_test_tbl                                          |           |              | defined
(12 rows)
\dX+ stts_?
                       List of extended statistics
 Schema |  Name  |    Definition     | Ndistinct | Dependencies |   MCV   
--------+--------+-------------------+-----------+--------------+---------
 public | stts_1 | a, b FROM stts_t1 | defined   |              | 
 public | stts_2 | a, b FROM stts_t1 | defined   | defined      | 
 public | stts_3 | a, b FROM stts_t1 | defined   | defined      | defined
 public | stts_4 | b, c FROM stts_t2 | defined   | defined      | defined
(4 rows)
\dX+ *stts_hoge
                               List of extended statistics
 Schema |   Name    |          Definition           | Ndistinct | Dependencies |   MCV   
--------+-----------+-------------------------------+-----------+--------------+---------
 public | stts_hoge | col1, col2, col3 FROM stts_t3 | defined   | defined      | defined
(1 row)
\dX+ stts_s2.stts_yama
                            List of extended statistics
 Schema  |   Name    |       Definition        | Ndistinct | Dependencies |   MCV   
---------+-----------+-------------------------+-----------+--------------+---------
 stts_s2 | stts_yama | col1, col3 FROM stts_t3 |           | defined      | defined
(1 row)
set search_path to public, stts_s1;`,
			},
			{
				Statement: `\dX
                                                       List of extended statistics
 Schema  |          Name          |                            Definition                            | Ndistinct | Dependencies |   MCV   
---------+------------------------+------------------------------------------------------------------+-----------+--------------+---------
 public  | func_deps_stat         | (a * 2), upper(b), (c + 1::numeric) FROM functional_dependencies |           | defined      | 
 public  | mcv_lists_arrays_stats | a, b, c FROM mcv_lists_arrays                                    |           |              | defined
 public  | mcv_lists_bool_stats   | a, b, c FROM mcv_lists_bool                                      |           |              | defined
 public  | mcv_lists_stats        | a, b, d FROM mcv_lists                                           |           |              | defined
 public  | stts_1                 | a, b FROM stts_t1                                                | defined   |              | 
 public  | stts_2                 | a, b FROM stts_t1                                                | defined   | defined      | 
 public  | stts_3                 | a, b FROM stts_t1                                                | defined   | defined      | defined
 public  | stts_4                 | b, c FROM stts_t2                                                | defined   | defined      | defined
 public  | stts_hoge              | col1, col2, col3 FROM stts_t3                                    | defined   | defined      | defined
 stts_s1 | stts_foo               | col1, col2 FROM stts_t3                                          | defined   | defined      | defined
(10 rows)
create role regress_stats_ext nosuperuser;`,
			},
			{
				Statement: `set role regress_stats_ext;`,
			},
			{
				Statement: `\dX
                                                       List of extended statistics
 Schema |          Name          |                            Definition                            | Ndistinct | Dependencies |   MCV   
--------+------------------------+------------------------------------------------------------------+-----------+--------------+---------
 public | func_deps_stat         | (a * 2), upper(b), (c + 1::numeric) FROM functional_dependencies |           | defined      | 
 public | mcv_lists_arrays_stats | a, b, c FROM mcv_lists_arrays                                    |           |              | defined
 public | mcv_lists_bool_stats   | a, b, c FROM mcv_lists_bool                                      |           |              | defined
 public | mcv_lists_stats        | a, b, d FROM mcv_lists                                           |           |              | defined
 public | stts_1                 | a, b FROM stts_t1                                                | defined   |              | 
 public | stts_2                 | a, b FROM stts_t1                                                | defined   | defined      | 
 public | stts_3                 | a, b FROM stts_t1                                                | defined   | defined      | defined
 public | stts_4                 | b, c FROM stts_t2                                                | defined   | defined      | defined
 public | stts_hoge              | col1, col2, col3 FROM stts_t3                                    | defined   | defined      | defined
(9 rows)
reset role;`,
			},
			{
				Statement: `drop table stts_t1, stts_t2, stts_t3;`,
			},
			{
				Statement: `drop schema stts_s1, stts_s2 cascade;`,
			},
			{
				Statement: `drop user regress_stats_ext;`,
			},
			{
				Statement: `reset search_path;`,
			},
			{
				Statement: `CREATE USER regress_stats_user1;`,
			},
			{
				Statement: `GRANT USAGE ON SCHEMA tststats TO regress_stats_user1;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_stats_user1;`,
			},
			{
				Statement:   `SELECT * FROM tststats.priv_test_tbl; -- Permission denied`,
				ErrorString: `permission denied for table priv_test_tbl`,
			},
			{
				Statement: `SELECT * FROM tststats.priv_test_tbl
  WHERE a = 1 and tststats.priv_test_tbl.* > (1, 1) is not null;`,
				ErrorString: `permission denied for table priv_test_tbl`,
			},
			{
				Statement: `CREATE FUNCTION op_leak(int, int) RETURNS bool
    AS 'BEGIN RAISE NOTICE ''op_leak => %, %'', $1, $2; RETURN $1 < $2; END'
    LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE OPERATOR <<< (procedure = op_leak, leftarg = int, rightarg = int,
                     restrict = scalarltsel);`,
			},
			{
				Statement:   `SELECT * FROM tststats.priv_test_tbl WHERE a <<< 0 AND b <<< 0; -- Permission denied`,
				ErrorString: `permission denied for table priv_test_tbl`,
			},
			{
				Statement:   `DELETE FROM tststats.priv_test_tbl WHERE a <<< 0 AND b <<< 0; -- Permission denied`,
				ErrorString: `permission denied for table priv_test_tbl`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `CREATE VIEW tststats.priv_test_view WITH (security_barrier=true)
    AS SELECT * FROM tststats.priv_test_tbl WHERE false;`,
			},
			{
				Statement: `GRANT SELECT, DELETE ON tststats.priv_test_view TO regress_stats_user1;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_stats_user1;`,
			},
			{
				Statement: `SELECT * FROM tststats.priv_test_view WHERE a <<< 0 AND b <<< 0; -- Should not leak`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DELETE FROM tststats.priv_test_view WHERE a <<< 0 AND b <<< 0; -- Should not leak`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `ALTER TABLE tststats.priv_test_tbl ENABLE ROW LEVEL SECURITY;`,
			},
			{
				Statement: `GRANT SELECT, DELETE ON tststats.priv_test_tbl TO regress_stats_user1;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_stats_user1;`,
			},
			{
				Statement: `SELECT * FROM tststats.priv_test_tbl WHERE a <<< 0 AND b <<< 0; -- Should not leak`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DELETE FROM tststats.priv_test_tbl WHERE a <<< 0 AND b <<< 0; -- Should not leak`,
			},
			{
				Statement: `DROP OPERATOR <<< (int, int);`,
			},
			{
				Statement: `DROP FUNCTION op_leak(int, int);`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP SCHEMA tststats CASCADE;`,
			},
			{
				Statement: `DROP USER regress_stats_user1;`,
			},
		},
	})
}
