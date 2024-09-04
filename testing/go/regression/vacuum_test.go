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

func TestVacuum(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_vacuum)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_vacuum,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE vactst (i INT);`,
			},
			{
				Statement: `INSERT INTO vactst VALUES (1);`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst VALUES (0);`,
			},
			{
				Statement: `SELECT count(*) FROM vactst;`,
				Results:   []sql.Row{{2049}},
			},
			{
				Statement: `DELETE FROM vactst WHERE i != 0;`,
			},
			{
				Statement: `SELECT * FROM vactst;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `VACUUM FULL vactst;`,
			},
			{
				Statement: `UPDATE vactst SET i = i + 1;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT * FROM vactst;`,
			},
			{
				Statement: `INSERT INTO vactst VALUES (0);`,
			},
			{
				Statement: `SELECT count(*) FROM vactst;`,
				Results:   []sql.Row{{2049}},
			},
			{
				Statement: `DELETE FROM vactst WHERE i != 0;`,
			},
			{
				Statement: `VACUUM (FULL) vactst;`,
			},
			{
				Statement: `DELETE FROM vactst;`,
			},
			{
				Statement: `SELECT * FROM vactst;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `VACUUM (FULL, FREEZE) vactst;`,
			},
			{
				Statement: `VACUUM (ANALYZE, FULL) vactst;`,
			},
			{
				Statement: `CREATE TABLE vaccluster (i INT PRIMARY KEY);`,
			},
			{
				Statement: `ALTER TABLE vaccluster CLUSTER ON vaccluster_pkey;`,
			},
			{
				Statement: `CLUSTER vaccluster;`,
			},
			{
				Statement: `CREATE FUNCTION do_analyze() RETURNS VOID VOLATILE LANGUAGE SQL
	AS 'ANALYZE pg_am';`,
			},
			{
				Statement: `CREATE FUNCTION wrap_do_analyze(c INT) RETURNS INT IMMUTABLE LANGUAGE SQL
	AS 'SELECT $1 FROM do_analyze()';`,
			},
			{
				Statement: `CREATE INDEX ON vaccluster(wrap_do_analyze(i));`,
			},
			{
				Statement: `INSERT INTO vaccluster VALUES (1), (2);`,
			},
			{
				Statement:   `ANALYZE vaccluster;`,
				ErrorString: `ANALYZE cannot be executed from VACUUM or ANALYZE`,
			},
			{
				Statement: `CONTEXT:  SQL function "do_analyze" statement 1
SQL function "wrap_do_analyze" statement 1
INSERT INTO vactst SELECT generate_series(1, 300);`,
			},
			{
				Statement: `DELETE FROM vactst WHERE i % 7 = 0; -- delete a few rows outside`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO vactst SELECT generate_series(301, 400);`,
			},
			{
				Statement: `DELETE FROM vactst WHERE i % 5 <> 0; -- delete a few rows inside`,
			},
			{
				Statement: `ANALYZE vactst;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `VACUUM FULL pg_am;`,
			},
			{
				Statement: `VACUUM FULL pg_class;`,
			},
			{
				Statement: `VACUUM FULL pg_database;`,
			},
			{
				Statement:   `VACUUM FULL vaccluster;`,
				ErrorString: `ANALYZE cannot be executed from VACUUM or ANALYZE`,
			},
			{
				Statement: `CONTEXT:  SQL function "do_analyze" statement 1
SQL function "wrap_do_analyze" statement 1
VACUUM FULL vactst;`,
			},
			{
				Statement: `VACUUM (DISABLE_PAGE_SKIPPING) vaccluster;`,
			},
			{
				Statement: `CREATE TABLE pvactst (i INT, a INT[], p POINT) with (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO pvactst SELECT i, array[1,2,3], point(i, i+1) FROM generate_series(1,1000) i;`,
			},
			{
				Statement: `CREATE INDEX btree_pvactst ON pvactst USING btree (i);`,
			},
			{
				Statement: `CREATE INDEX hash_pvactst ON pvactst USING hash (i);`,
			},
			{
				Statement: `CREATE INDEX brin_pvactst ON pvactst USING brin (i);`,
			},
			{
				Statement: `CREATE INDEX gin_pvactst ON pvactst USING gin (a);`,
			},
			{
				Statement: `CREATE INDEX gist_pvactst ON pvactst USING gist (p);`,
			},
			{
				Statement: `CREATE INDEX spgist_pvactst ON pvactst USING spgist (p);`,
			},
			{
				Statement: `SET min_parallel_index_scan_size to 0;`,
			},
			{
				Statement: `VACUUM (PARALLEL 2) pvactst;`,
			},
			{
				Statement: `UPDATE pvactst SET i = i WHERE i < 1000;`,
			},
			{
				Statement: `VACUUM (PARALLEL 2) pvactst;`,
			},
			{
				Statement: `UPDATE pvactst SET i = i WHERE i < 1000;`,
			},
			{
				Statement: `VACUUM (PARALLEL 0) pvactst; -- disable parallel vacuum`,
			},
			{
				Statement:   `VACUUM (PARALLEL -1) pvactst; -- error`,
				ErrorString: `parallel workers for vacuum must be between 0 and 1024`,
			},
			{
				Statement: `VACUUM (PARALLEL 2, INDEX_CLEANUP FALSE) pvactst;`,
			},
			{
				Statement:   `VACUUM (PARALLEL 2, FULL TRUE) pvactst; -- error, cannot use both PARALLEL and FULL`,
				ErrorString: `VACUUM FULL cannot be performed in parallel`,
			},
			{
				Statement:   `VACUUM (PARALLEL) pvactst; -- error, cannot use PARALLEL option without parallel degree`,
				ErrorString: `parallel option requires a value between 0 and 1024`,
			},
			{
				Statement: `CREATE TEMPORARY TABLE tmp (a int PRIMARY KEY);`,
			},
			{
				Statement: `CREATE INDEX tmp_idx1 ON tmp (a);`,
			},
			{
				Statement: `VACUUM (PARALLEL 1, FULL FALSE) tmp; -- parallel vacuum disabled for temp tables`,
			},
			{
				Statement: `VACUUM (PARALLEL 0, FULL TRUE) tmp; -- can specify parallel disabled (even though that's implied by FULL)`,
			},
			{
				Statement: `RESET min_parallel_index_scan_size;`,
			},
			{
				Statement: `DROP TABLE pvactst;`,
			},
			{
				Statement: `CREATE TABLE no_index_cleanup (i INT PRIMARY KEY, t TEXT);`,
			},
			{
				Statement: `CREATE INDEX no_index_cleanup_idx ON no_index_cleanup(t);`,
			},
			{
				Statement: `ALTER TABLE no_index_cleanup ALTER COLUMN t SET STORAGE EXTERNAL;`,
			},
			{
				Statement: `INSERT INTO no_index_cleanup(i, t) VALUES (generate_series(1,30),
    repeat('1234567890',269));`,
			},
			{
				Statement: `VACUUM (INDEX_CLEANUP TRUE, FULL TRUE) no_index_cleanup;`,
			},
			{
				Statement: `VACUUM (FULL TRUE) no_index_cleanup;`,
			},
			{
				Statement: `ALTER TABLE no_index_cleanup SET (vacuum_index_cleanup = false);`,
			},
			{
				Statement: `DELETE FROM no_index_cleanup WHERE i < 15;`,
			},
			{
				Statement: `VACUUM no_index_cleanup;`,
			},
			{
				Statement: `ALTER TABLE no_index_cleanup SET (vacuum_index_cleanup = true);`,
			},
			{
				Statement: `VACUUM no_index_cleanup;`,
			},
			{
				Statement: `ALTER TABLE no_index_cleanup SET (vacuum_index_cleanup = auto);`,
			},
			{
				Statement: `VACUUM no_index_cleanup;`,
			},
			{
				Statement: `INSERT INTO no_index_cleanup(i, t) VALUES (generate_series(31,60),
    repeat('1234567890',269));`,
			},
			{
				Statement: `DELETE FROM no_index_cleanup WHERE i < 45;`,
			},
			{
				Statement: `ALTER TABLE no_index_cleanup SET (vacuum_index_cleanup = off,
    toast.vacuum_index_cleanup = yes);`,
			},
			{
				Statement: `VACUUM no_index_cleanup;`,
			},
			{
				Statement: `ALTER TABLE no_index_cleanup SET (vacuum_index_cleanup = true,
    toast.vacuum_index_cleanup = false);`,
			},
			{
				Statement: `VACUUM no_index_cleanup;`,
			},
			{
				Statement: `VACUUM (INDEX_CLEANUP FALSE) vaccluster;`,
			},
			{
				Statement: `VACUUM (INDEX_CLEANUP AUTO) vactst; -- index cleanup option is ignored if no indexes`,
			},
			{
				Statement: `VACUUM (INDEX_CLEANUP FALSE, FREEZE TRUE) vaccluster;`,
			},
			{
				Statement: `CREATE TEMP TABLE vac_truncate_test(i INT NOT NULL, j text)
	WITH (vacuum_truncate=true, autovacuum_enabled=false);`,
			},
			{
				Statement:   `INSERT INTO vac_truncate_test VALUES (1, NULL), (NULL, NULL);`,
				ErrorString: `null value in column "i" of relation "vac_truncate_test" violates not-null constraint`,
			},
			{
				Statement: `VACUUM (TRUNCATE FALSE, DISABLE_PAGE_SKIPPING) vac_truncate_test;`,
			},
			{
				Statement: `SELECT pg_relation_size('vac_truncate_test') > 0;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `VACUUM (DISABLE_PAGE_SKIPPING) vac_truncate_test;`,
			},
			{
				Statement: `SELECT pg_relation_size('vac_truncate_test') = 0;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `VACUUM (TRUNCATE FALSE, FULL TRUE) vac_truncate_test;`,
			},
			{
				Statement: `DROP TABLE vac_truncate_test;`,
			},
			{
				Statement: `CREATE TABLE vacparted (a int, b char) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE vacparted1 PARTITION OF vacparted FOR VALUES IN (1);`,
			},
			{
				Statement: `INSERT INTO vacparted VALUES (1, 'a');`,
			},
			{
				Statement: `UPDATE vacparted SET b = 'b';`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacparted;`,
			},
			{
				Statement: `VACUUM (FULL) vacparted;`,
			},
			{
				Statement: `VACUUM (FREEZE) vacparted;`,
			},
			{
				Statement:   `VACUUM ANALYZE vacparted(a,b,a);`,
				ErrorString: `column "a" of relation "vacparted" appears more than once`,
			},
			{
				Statement:   `ANALYZE vacparted(a,b,b);`,
				ErrorString: `column "b" of relation "vacparted" appears more than once`,
			},
			{
				Statement: `CREATE TABLE vacparted_i (a int primary key, b varchar(100))
  PARTITION BY HASH (a);`,
			},
			{
				Statement: `CREATE TABLE vacparted_i1 PARTITION OF vacparted_i
  FOR VALUES WITH (MODULUS 2, REMAINDER 0);`,
			},
			{
				Statement: `CREATE TABLE vacparted_i2 PARTITION OF vacparted_i
  FOR VALUES WITH (MODULUS 2, REMAINDER 1);`,
			},
			{
				Statement: `INSERT INTO vacparted_i SELECT i, 'test_'|| i from generate_series(1,10) i;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacparted_i;`,
			},
			{
				Statement: `VACUUM (FULL) vacparted_i;`,
			},
			{
				Statement: `VACUUM (FREEZE) vacparted_i;`,
			},
			{
				Statement: `SELECT relname, relhasindex FROM pg_class
  WHERE relname LIKE 'vacparted_i%' AND relkind IN ('p','r')
  ORDER BY relname;`,
				Results: []sql.Row{{`vacparted_i`, true}, {`vacparted_i1`, true}, {`vacparted_i2`, true}},
			},
			{
				Statement: `DROP TABLE vacparted_i;`,
			},
			{
				Statement: `VACUUM vaccluster, vactst;`,
			},
			{
				Statement:   `VACUUM vacparted, does_not_exist;`,
				ErrorString: `relation "does_not_exist" does not exist`,
			},
			{
				Statement: `VACUUM (FREEZE) vacparted, vaccluster, vactst;`,
			},
			{
				Statement:   `VACUUM (FREEZE) does_not_exist, vaccluster;`,
				ErrorString: `relation "does_not_exist" does not exist`,
			},
			{
				Statement: `VACUUM ANALYZE vactst, vacparted (a);`,
			},
			{
				Statement:   `VACUUM ANALYZE vactst (does_not_exist), vacparted (b);`,
				ErrorString: `column "does_not_exist" of relation "vactst" does not exist`,
			},
			{
				Statement: `VACUUM FULL vacparted, vactst;`,
			},
			{
				Statement:   `VACUUM FULL vactst, vacparted (a, b), vaccluster (i);`,
				ErrorString: `ANALYZE option must be specified when a column list is provided`,
			},
			{
				Statement: `ANALYZE vactst, vacparted;`,
			},
			{
				Statement: `ANALYZE vacparted (b), vactst;`,
			},
			{
				Statement:   `ANALYZE vactst, does_not_exist, vacparted;`,
				ErrorString: `relation "does_not_exist" does not exist`,
			},
			{
				Statement:   `ANALYZE vactst (i), vacparted (does_not_exist);`,
				ErrorString: `column "does_not_exist" of relation "vacparted" does not exist`,
			},
			{
				Statement: `ANALYZE vactst, vactst;`,
			},
			{
				Statement: `BEGIN;  -- ANALYZE behaves differently inside a transaction block`,
			},
			{
				Statement: `ANALYZE vactst, vactst;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement:   `ANALYZE (VERBOSE) does_not_exist;`,
				ErrorString: `relation "does_not_exist" does not exist`,
			},
			{
				Statement:   `ANALYZE (nonexistent-arg) does_not_exist;`,
				ErrorString: `syntax error at or near "arg"`,
			},
			{
				Statement:   `ANALYZE (nonexistentarg) does_not_exit;`,
				ErrorString: `unrecognized ANALYZE option "nonexistentarg"`,
			},
			{
				Statement: `SET client_min_messages TO 'ERROR';`,
			},
			{
				Statement:   `ANALYZE (SKIP_LOCKED, VERBOSE) does_not_exist;`,
				ErrorString: `relation "does_not_exist" does not exist`,
			},
			{
				Statement:   `ANALYZE (VERBOSE, SKIP_LOCKED) does_not_exist;`,
				ErrorString: `relation "does_not_exist" does not exist`,
			},
			{
				Statement: `VACUUM (SKIP_LOCKED) vactst;`,
			},
			{
				Statement: `VACUUM (SKIP_LOCKED, FULL) vactst;`,
			},
			{
				Statement: `ANALYZE (SKIP_LOCKED) vactst;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `SET default_transaction_isolation = serializable;`,
			},
			{
				Statement: `VACUUM vactst;`,
			},
			{
				Statement: `ANALYZE vactst;`,
			},
			{
				Statement: `RESET default_transaction_isolation;`,
			},
			{
				Statement: `BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;`,
			},
			{
				Statement: `ANALYZE vactst;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `ALTER TABLE vactst ADD COLUMN t TEXT;`,
			},
			{
				Statement: `ALTER TABLE vactst ALTER COLUMN t SET STORAGE EXTERNAL;`,
			},
			{
				Statement: `VACUUM (PROCESS_TOAST FALSE) vactst;`,
			},
			{
				Statement:   `VACUUM (PROCESS_TOAST FALSE, FULL) vactst;`,
				ErrorString: `PROCESS_TOAST required with VACUUM FULL`,
			},
			{
				Statement: `DROP TABLE vaccluster;`,
			},
			{
				Statement: `DROP TABLE vactst;`,
			},
			{
				Statement: `DROP TABLE vacparted;`,
			},
			{
				Statement: `DROP TABLE no_index_cleanup;`,
			},
			{
				Statement: `CREATE TABLE vacowned (a int);`,
			},
			{
				Statement: `CREATE TABLE vacowned_parted (a int) PARTITION BY LIST (a);`,
			},
			{
				Statement: `CREATE TABLE vacowned_part1 PARTITION OF vacowned_parted FOR VALUES IN (1);`,
			},
			{
				Statement: `CREATE TABLE vacowned_part2 PARTITION OF vacowned_parted FOR VALUES IN (2);`,
			},
			{
				Statement: `CREATE ROLE regress_vacuum;`,
			},
			{
				Statement: `SET ROLE regress_vacuum;`,
			},
			{
				Statement: `VACUUM vacowned;`,
			},
			{
				Statement: `ANALYZE vacowned;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned;`,
			},
			{
				Statement: `VACUUM pg_catalog.pg_class;`,
			},
			{
				Statement: `ANALYZE pg_catalog.pg_class;`,
			},
			{
				Statement: `VACUUM (ANALYZE) pg_catalog.pg_class;`,
			},
			{
				Statement: `VACUUM pg_catalog.pg_authid;`,
			},
			{
				Statement: `ANALYZE pg_catalog.pg_authid;`,
			},
			{
				Statement: `VACUUM (ANALYZE) pg_catalog.pg_authid;`,
			},
			{
				Statement: `VACUUM vacowned_parted;`,
			},
			{
				Statement: `VACUUM vacowned_part1;`,
			},
			{
				Statement: `VACUUM vacowned_part2;`,
			},
			{
				Statement: `ANALYZE vacowned_parted;`,
			},
			{
				Statement: `ANALYZE vacowned_part1;`,
			},
			{
				Statement: `ANALYZE vacowned_part2;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_parted;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_part1;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_part2;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `ALTER TABLE vacowned_parted OWNER TO regress_vacuum;`,
			},
			{
				Statement: `ALTER TABLE vacowned_part1 OWNER TO regress_vacuum;`,
			},
			{
				Statement: `SET ROLE regress_vacuum;`,
			},
			{
				Statement: `VACUUM vacowned_parted;`,
			},
			{
				Statement: `VACUUM vacowned_part1;`,
			},
			{
				Statement: `VACUUM vacowned_part2;`,
			},
			{
				Statement: `ANALYZE vacowned_parted;`,
			},
			{
				Statement: `ANALYZE vacowned_part1;`,
			},
			{
				Statement: `ANALYZE vacowned_part2;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_parted;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_part1;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_part2;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `ALTER TABLE vacowned_parted OWNER TO CURRENT_USER;`,
			},
			{
				Statement: `SET ROLE regress_vacuum;`,
			},
			{
				Statement: `VACUUM vacowned_parted;`,
			},
			{
				Statement: `VACUUM vacowned_part1;`,
			},
			{
				Statement: `VACUUM vacowned_part2;`,
			},
			{
				Statement: `ANALYZE vacowned_parted;`,
			},
			{
				Statement: `ANALYZE vacowned_part1;`,
			},
			{
				Statement: `ANALYZE vacowned_part2;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_parted;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_part1;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_part2;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `ALTER TABLE vacowned_parted OWNER TO regress_vacuum;`,
			},
			{
				Statement: `ALTER TABLE vacowned_part1 OWNER TO CURRENT_USER;`,
			},
			{
				Statement: `SET ROLE regress_vacuum;`,
			},
			{
				Statement: `VACUUM vacowned_parted;`,
			},
			{
				Statement: `VACUUM vacowned_part1;`,
			},
			{
				Statement: `VACUUM vacowned_part2;`,
			},
			{
				Statement: `ANALYZE vacowned_parted;`,
			},
			{
				Statement: `ANALYZE vacowned_part1;`,
			},
			{
				Statement: `ANALYZE vacowned_part2;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_parted;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_part1;`,
			},
			{
				Statement: `VACUUM (ANALYZE) vacowned_part2;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `DROP TABLE vacowned;`,
			},
			{
				Statement: `DROP TABLE vacowned_parted;`,
			},
			{
				Statement: `DROP ROLE regress_vacuum;`,
			},
		},
	})
}
