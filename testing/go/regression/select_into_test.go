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

func TestSelectInto(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_select_into)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_select_into,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT *
   INTO TABLE sitmp1
   FROM onek
   WHERE onek.unique1 < 2;`,
			},
			{
				Statement: `DROP TABLE sitmp1;`,
			},
			{
				Statement: `SELECT *
   INTO TABLE sitmp1
   FROM onek2
   WHERE onek2.unique1 < 2;`,
			},
			{
				Statement: `DROP TABLE sitmp1;`,
			},
			{
				Statement: `CREATE SCHEMA selinto_schema;`,
			},
			{
				Statement: `CREATE USER regress_selinto_user;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_selinto_user
	  REVOKE INSERT ON TABLES FROM regress_selinto_user;`,
			},
			{
				Statement: `GRANT ALL ON SCHEMA selinto_schema TO public;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_selinto_user;`,
			},
			{
				Statement: `CREATE TABLE selinto_schema.tbl_withdata1 (a)
  AS SELECT generate_series(1,3) WITH DATA;`,
			},
			{
				Statement:   `INSERT INTO selinto_schema.tbl_withdata1 VALUES (4);`,
				ErrorString: `permission denied for table tbl_withdata1`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE selinto_schema.tbl_withdata2 (a) AS
  SELECT generate_series(1,3) WITH DATA;`,
				Results: []sql.Row{{`ProjectSet (actual rows=3 loops=1)`}, {`->  Result (actual rows=1 loops=1)`}},
			},
			{
				Statement: `CREATE TABLE selinto_schema.tbl_nodata1 (a) AS
  SELECT generate_series(1,3) WITH NO DATA;`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE selinto_schema.tbl_nodata2 (a) AS
  SELECT generate_series(1,3) WITH NO DATA;`,
				Results: []sql.Row{{`ProjectSet (never executed)`}, {`->  Result (never executed)`}},
			},
			{
				Statement: `PREPARE data_sel AS SELECT generate_series(1,3);`,
			},
			{
				Statement: `CREATE TABLE selinto_schema.tbl_withdata3 (a) AS
  EXECUTE data_sel WITH DATA;`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE selinto_schema.tbl_withdata4 (a) AS
  EXECUTE data_sel WITH DATA;`,
				Results: []sql.Row{{`ProjectSet (actual rows=3 loops=1)`}, {`->  Result (actual rows=1 loops=1)`}},
			},
			{
				Statement: `CREATE TABLE selinto_schema.tbl_nodata3 (a) AS
  EXECUTE data_sel WITH NO DATA;`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE selinto_schema.tbl_nodata4 (a) AS
  EXECUTE data_sel WITH NO DATA;`,
				Results: []sql.Row{{`ProjectSet (never executed)`}, {`->  Result (never executed)`}},
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `ALTER DEFAULT PRIVILEGES FOR ROLE regress_selinto_user
	  GRANT INSERT ON TABLES TO regress_selinto_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_selinto_user;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DEALLOCATE data_sel;`,
			},
			{
				Statement: `DROP SCHEMA selinto_schema CASCADE;`,
			},
			{
				Statement: `DROP USER regress_selinto_user;`,
			},
			{
				Statement: `CREATE TABLE ctas_base (i int, j int);`,
			},
			{
				Statement: `INSERT INTO ctas_base VALUES (1, 2);`,
			},
			{
				Statement:   `CREATE TABLE ctas_nodata (ii, jj, kk) AS SELECT i, j FROM ctas_base; -- Error`,
				ErrorString: `too many column names were specified`,
			},
			{
				Statement:   `CREATE TABLE ctas_nodata (ii, jj, kk) AS SELECT i, j FROM ctas_base WITH NO DATA; -- Error`,
				ErrorString: `too many column names were specified`,
			},
			{
				Statement: `CREATE TABLE ctas_nodata (ii, jj) AS SELECT i, j FROM ctas_base; -- OK`,
			},
			{
				Statement: `CREATE TABLE ctas_nodata_2 (ii, jj) AS SELECT i, j FROM ctas_base WITH NO DATA; -- OK`,
			},
			{
				Statement: `CREATE TABLE ctas_nodata_3 (ii) AS SELECT i, j FROM ctas_base; -- OK`,
			},
			{
				Statement: `CREATE TABLE ctas_nodata_4 (ii) AS SELECT i, j FROM ctas_base WITH NO DATA; -- OK`,
			},
			{
				Statement: `SELECT * FROM ctas_nodata;`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `SELECT * FROM ctas_nodata_2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM ctas_nodata_3;`,
				Results:   []sql.Row{{1, 2}},
			},
			{
				Statement: `SELECT * FROM ctas_nodata_4;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE ctas_base;`,
			},
			{
				Statement: `DROP TABLE ctas_nodata;`,
			},
			{
				Statement: `DROP TABLE ctas_nodata_2;`,
			},
			{
				Statement: `DROP TABLE ctas_nodata_3;`,
			},
			{
				Statement: `DROP TABLE ctas_nodata_4;`,
			},
			{
				Statement: `CREATE FUNCTION make_table() RETURNS VOID
AS $$
  CREATE TABLE created_table AS SELECT * FROM int8_tbl;`,
			},
			{
				Statement: `$$ LANGUAGE SQL;`,
			},
			{
				Statement: `SELECT make_table();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT * FROM created_table;`,
				Results:   []sql.Row{{123, 456}, {123, 4567890123456789}, {4567890123456789, 123}, {4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789}},
			},
			{
				Statement: `DO $$
BEGIN
	EXECUTE 'EXPLAIN ANALYZE SELECT * INTO TABLE easi FROM int8_tbl';`,
			},
			{
				Statement: `	EXECUTE 'EXPLAIN ANALYZE CREATE TABLE easi2 AS SELECT * FROM int8_tbl WITH NO DATA';`,
			},
			{
				Statement: `END$$;`,
			},
			{
				Statement: `DROP TABLE created_table;`,
			},
			{
				Statement: `DROP TABLE easi, easi2;`,
			},
			{
				Statement:   `DECLARE foo CURSOR FOR SELECT 1 INTO int4_tbl;`,
				ErrorString: `SELECT ... INTO is not allowed here`,
			},
			{
				Statement:   `COPY (SELECT 1 INTO frak UNION SELECT 2) TO 'blob';`,
				ErrorString: `COPY (SELECT INTO) is not supported`,
			},
			{
				Statement:   `SELECT * FROM (SELECT 1 INTO f) bar;`,
				ErrorString: `SELECT ... INTO is not allowed here`,
			},
			{
				Statement:   `CREATE VIEW foo AS SELECT 1 INTO int4_tbl;`,
				ErrorString: `views must not contain SELECT INTO`,
			},
			{
				Statement:   `INSERT INTO int4_tbl SELECT 1 INTO f;`,
				ErrorString: `SELECT ... INTO is not allowed here`,
			},
			{
				Statement: `CREATE TABLE ctas_ine_tbl AS SELECT 1;`,
			},
			{
				Statement:   `CREATE TABLE ctas_ine_tbl AS SELECT 1 / 0; -- error`,
				ErrorString: `relation "ctas_ine_tbl" already exists`,
			},
			{
				Statement: `CREATE TABLE IF NOT EXISTS ctas_ine_tbl AS SELECT 1 / 0; -- ok`,
			},
			{
				Statement:   `CREATE TABLE ctas_ine_tbl AS SELECT 1 / 0 WITH NO DATA; -- error`,
				ErrorString: `relation "ctas_ine_tbl" already exists`,
			},
			{
				Statement: `CREATE TABLE IF NOT EXISTS ctas_ine_tbl AS SELECT 1 / 0 WITH NO DATA; -- ok`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE ctas_ine_tbl AS SELECT 1 / 0; -- error`,
				ErrorString: `relation "ctas_ine_tbl" already exists`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE IF NOT EXISTS ctas_ine_tbl AS SELECT 1 / 0; -- ok`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE ctas_ine_tbl AS SELECT 1 / 0 WITH NO DATA; -- error`,
				ErrorString: `relation "ctas_ine_tbl" already exists`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE IF NOT EXISTS ctas_ine_tbl AS SELECT 1 / 0 WITH NO DATA; -- ok`,
				Results: []sql.Row{},
			},
			{
				Statement: `PREPARE ctas_ine_query AS SELECT 1 / 0;`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE ctas_ine_tbl AS EXECUTE ctas_ine_query; -- error`,
				ErrorString: `relation "ctas_ine_tbl" already exists`,
			},
			{
				Statement: `EXPLAIN (ANALYZE, COSTS OFF, SUMMARY OFF, TIMING OFF)
  CREATE TABLE IF NOT EXISTS ctas_ine_tbl AS EXECUTE ctas_ine_query; -- ok`,
				Results: []sql.Row{},
			},
			{
				Statement: `DROP TABLE ctas_ine_tbl;`,
			},
		},
	})
}
