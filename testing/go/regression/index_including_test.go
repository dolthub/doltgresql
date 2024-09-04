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

func TestIndexIncluding(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_index_including)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_index_including,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `/*
 * 1.test CREATE INDEX
 *
 * Deliberately avoid dropping objects in this section, to get some pg_dump
 * coverage.
 */
CREATE TABLE tbl_include_reg (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_include_reg SELECT x, 2*x, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `CREATE INDEX tbl_include_reg_idx ON tbl_include_reg (c1, c2) INCLUDE (c3, c4);`,
			},
			{
				Statement: `CREATE INDEX ON tbl_include_reg (c1, c2) INCLUDE (c1, c3);`,
			},
			{
				Statement: `SELECT pg_get_indexdef(i.indexrelid)
FROM pg_index i JOIN pg_class c ON i.indexrelid = c.oid
WHERE i.indrelid = 'tbl_include_reg'::regclass ORDER BY c.relname;`,
				Results: []sql.Row{{`CREATE INDEX tbl_include_reg_c1_c2_c11_c3_idx ON public.tbl_include_reg USING btree (c1, c2) INCLUDE (c1, c3)`}, {`CREATE INDEX tbl_include_reg_idx ON public.tbl_include_reg USING btree (c1, c2) INCLUDE (c3, c4)`}},
			},
			{
				Statement: `\d tbl_include_reg_idx
  Index "public.tbl_include_reg_idx"
 Column |  Type   | Key? | Definition 
--------+---------+------+------------
 c1     | integer | yes  | c1
 c2     | integer | yes  | c2
 c3     | integer | no   | c3
 c4     | box     | no   | c4
btree, for table "public.tbl_include_reg"
CREATE TABLE tbl_include_unique1 (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_include_unique1 SELECT x, 2*x, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX tbl_include_unique1_idx_unique ON tbl_include_unique1 using btree (c1, c2) INCLUDE (c3, c4);`,
			},
			{
				Statement: `ALTER TABLE tbl_include_unique1 add UNIQUE USING INDEX tbl_include_unique1_idx_unique;`,
			},
			{
				Statement: `ALTER TABLE tbl_include_unique1 add UNIQUE (c1, c2) INCLUDE (c3, c4);`,
			},
			{
				Statement: `SELECT pg_get_indexdef(i.indexrelid)
FROM pg_index i JOIN pg_class c ON i.indexrelid = c.oid
WHERE i.indrelid = 'tbl_include_unique1'::regclass ORDER BY c.relname;`,
				Results: []sql.Row{{`CREATE UNIQUE INDEX tbl_include_unique1_c1_c2_c3_c4_key ON public.tbl_include_unique1 USING btree (c1, c2) INCLUDE (c3, c4)`}, {`CREATE UNIQUE INDEX tbl_include_unique1_idx_unique ON public.tbl_include_unique1 USING btree (c1, c2) INCLUDE (c3, c4)`}},
			},
			{
				Statement: `CREATE TABLE tbl_include_unique2 (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_include_unique2 SELECT 1, 2, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
			},
			{
				Statement:   `CREATE UNIQUE INDEX tbl_include_unique2_idx_unique ON tbl_include_unique2 using btree (c1, c2) INCLUDE (c3, c4);`,
				ErrorString: `could not create unique index "tbl_include_unique2_idx_unique"`,
			},
			{
				Statement:   `ALTER TABLE tbl_include_unique2 add UNIQUE (c1, c2) INCLUDE (c3, c4);`,
				ErrorString: `could not create unique index "tbl_include_unique2_c1_c2_c3_c4_key"`,
			},
			{
				Statement: `CREATE TABLE tbl_include_pk (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_include_pk SELECT 1, 2*x, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `ALTER TABLE tbl_include_pk add PRIMARY KEY (c1, c2) INCLUDE (c3, c4);`,
			},
			{
				Statement: `SELECT pg_get_indexdef(i.indexrelid)
FROM pg_index i JOIN pg_class c ON i.indexrelid = c.oid
WHERE i.indrelid = 'tbl_include_pk'::regclass ORDER BY c.relname;`,
				Results: []sql.Row{{`CREATE UNIQUE INDEX tbl_include_pk_pkey ON public.tbl_include_pk USING btree (c1, c2) INCLUDE (c3, c4)`}},
			},
			{
				Statement: `CREATE TABLE tbl_include_box (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_include_box SELECT 1, 2*x, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX tbl_include_box_idx_unique ON tbl_include_box using btree (c1, c2) INCLUDE (c3, c4);`,
			},
			{
				Statement: `ALTER TABLE tbl_include_box add PRIMARY KEY USING INDEX tbl_include_box_idx_unique;`,
			},
			{
				Statement: `SELECT pg_get_indexdef(i.indexrelid)
FROM pg_index i JOIN pg_class c ON i.indexrelid = c.oid
WHERE i.indrelid = 'tbl_include_box'::regclass ORDER BY c.relname;`,
				Results: []sql.Row{{`CREATE UNIQUE INDEX tbl_include_box_idx_unique ON public.tbl_include_box USING btree (c1, c2) INCLUDE (c3, c4)`}},
			},
			{
				Statement: `CREATE TABLE tbl_include_box_pk (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_include_box_pk SELECT 1, 2, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
			},
			{
				Statement:   `ALTER TABLE tbl_include_box_pk add PRIMARY KEY (c1, c2) INCLUDE (c3, c4);`,
				ErrorString: `could not create unique index "tbl_include_box_pk_pkey"`,
			},
			{
				Statement: `/*
 * 2. Test CREATE TABLE with constraint
 */
CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box,
				CONSTRAINT covering UNIQUE(c1,c2) INCLUDE(c3,c4));`,
			},
			{
				Statement: `SELECT indexrelid::regclass, indnatts, indnkeyatts, indisunique, indisprimary, indkey, indclass FROM pg_index WHERE indrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`covering`, 4, 2, true, false, `1 2 3 4`, `1978 1978`}},
			},
			{
				Statement: `SELECT pg_get_constraintdef(oid), conname, conkey FROM pg_constraint WHERE conrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`UNIQUE (c1, c2) INCLUDE (c3, c4)`, `covering`, `{1,2}`}},
			},
			{
				Statement:   `INSERT INTO tbl SELECT 1, 2, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
				ErrorString: `duplicate key value violates unique constraint "covering"`,
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box,
				CONSTRAINT covering PRIMARY KEY(c1,c2) INCLUDE(c3,c4));`,
			},
			{
				Statement: `SELECT indexrelid::regclass, indnatts, indnkeyatts, indisunique, indisprimary, indkey, indclass FROM pg_index WHERE indrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`covering`, 4, 2, true, true, `1 2 3 4`, `1978 1978`}},
			},
			{
				Statement: `SELECT pg_get_constraintdef(oid), conname, conkey FROM pg_constraint WHERE conrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`PRIMARY KEY (c1, c2) INCLUDE (c3, c4)`, `covering`, `{1,2}`}},
			},
			{
				Statement:   `INSERT INTO tbl SELECT 1, 2, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
				ErrorString: `duplicate key value violates unique constraint "covering"`,
			},
			{
				Statement:   `INSERT INTO tbl SELECT 1, NULL, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
				ErrorString: `null value in column "c2" of relation "tbl" violates not-null constraint`,
			},
			{
				Statement: `INSERT INTO tbl SELECT x, 2*x, NULL, NULL FROM generate_series(1,300) AS x;`,
			},
			{
				Statement: `explain (costs off)
select * from tbl where (c1,c2,c3) < (2,5,1);`,
				Results: []sql.Row{{`Bitmap Heap Scan on tbl`}, {`Filter: (ROW(c1, c2, c3) < ROW(2, 5, 1))`}, {`->  Bitmap Index Scan on covering`}, {`Index Cond: (ROW(c1, c2) <= ROW(2, 5))`}},
			},
			{
				Statement: `select * from tbl where (c1,c2,c3) < (2,5,1);`,
				Results:   []sql.Row{{1, 2, ``, ``}, {2, 4, ``, ``}},
			},
			{
				Statement: `SET enable_seqscan = off;`,
			},
			{
				Statement: `explain (costs off)
select * from tbl where (c1,c2,c3) < (262,1,1) limit 1;`,
				Results: []sql.Row{{`Limit`}, {`->  Index Only Scan using covering on tbl`}, {`Index Cond: (ROW(c1, c2) <= ROW(262, 1))`}, {`Filter: (ROW(c1, c2, c3) < ROW(262, 1, 1))`}},
			},
			{
				Statement: `select * from tbl where (c1,c2,c3) < (262,1,1) limit 1;`,
				Results:   []sql.Row{{1, 2, ``, ``}},
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box,
				UNIQUE(c1,c2) INCLUDE(c3,c4));`,
			},
			{
				Statement: `SELECT indexrelid::regclass, indnatts, indnkeyatts, indisunique, indisprimary, indkey, indclass FROM pg_index WHERE indrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`tbl_c1_c2_c3_c4_key`, 4, 2, true, false, `1 2 3 4`, `1978 1978`}},
			},
			{
				Statement: `SELECT pg_get_constraintdef(oid), conname, conkey FROM pg_constraint WHERE conrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`UNIQUE (c1, c2) INCLUDE (c3, c4)`, `tbl_c1_c2_c3_c4_key`, `{1,2}`}},
			},
			{
				Statement:   `INSERT INTO tbl SELECT 1, 2, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
				ErrorString: `duplicate key value violates unique constraint "tbl_c1_c2_c3_c4_key"`,
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box,
				PRIMARY KEY(c1,c2) INCLUDE(c3,c4));`,
			},
			{
				Statement: `SELECT indexrelid::regclass, indnatts, indnkeyatts, indisunique, indisprimary, indkey, indclass FROM pg_index WHERE indrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`tbl_pkey`, 4, 2, true, true, `1 2 3 4`, `1978 1978`}},
			},
			{
				Statement: `SELECT pg_get_constraintdef(oid), conname, conkey FROM pg_constraint WHERE conrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`PRIMARY KEY (c1, c2) INCLUDE (c3, c4)`, `tbl_pkey`, `{1,2}`}},
			},
			{
				Statement:   `INSERT INTO tbl SELECT 1, 2, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
				ErrorString: `duplicate key value violates unique constraint "tbl_pkey"`,
			},
			{
				Statement:   `INSERT INTO tbl SELECT 1, NULL, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
				ErrorString: `null value in column "c2" of relation "tbl" violates not-null constraint`,
			},
			{
				Statement: `INSERT INTO tbl SELECT x, 2*x, NULL, NULL FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box,
				EXCLUDE USING btree (c1 WITH =) INCLUDE(c3,c4));`,
			},
			{
				Statement: `SELECT indexrelid::regclass, indnatts, indnkeyatts, indisunique, indisprimary, indkey, indclass FROM pg_index WHERE indrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`tbl_c1_c3_c4_excl`, 3, 1, false, false, `1 3 4`, 1978}},
			},
			{
				Statement: `SELECT pg_get_constraintdef(oid), conname, conkey FROM pg_constraint WHERE conrelid = 'tbl'::regclass::oid;`,
				Results:   []sql.Row{{`EXCLUDE USING btree (c1 WITH =) INCLUDE (c3, c4)`, `tbl_c1_c3_c4_excl`, `{1}`}},
			},
			{
				Statement:   `INSERT INTO tbl SELECT 1, 2, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
				ErrorString: `conflicting key value violates exclusion constraint "tbl_c1_c3_c4_excl"`,
			},
			{
				Statement: `INSERT INTO tbl SELECT x, 2*x, NULL, NULL FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 3.0 Test ALTER TABLE DROP COLUMN.
 * Any column deletion leads to index deletion.
 */
CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 int);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX tbl_idx ON tbl using btree(c1, c2, c3, c4);`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX tbl_idx ON public.tbl USING btree (c1, c2, c3, c4)`}},
			},
			{
				Statement: `ALTER TABLE tbl DROP COLUMN c3;`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 3.1 Test ALTER TABLE DROP COLUMN.
 * Included column deletion leads to the index deletion,
 * AS well AS key columns deletion. It's explained in documentation.
 */
CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX tbl_idx ON tbl using btree(c1, c2) INCLUDE(c3,c4);`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX tbl_idx ON public.tbl USING btree (c1, c2) INCLUDE (c3, c4)`}},
			},
			{
				Statement: `ALTER TABLE tbl DROP COLUMN c3;`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 3.2 Test ALTER TABLE DROP COLUMN.
 * Included column deletion leads to the index deletion.
 * AS well AS key columns deletion. It's explained in documentation.
 */
CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box, UNIQUE(c1, c2) INCLUDE(c3,c4));`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX tbl_c1_c2_c3_c4_key ON public.tbl USING btree (c1, c2) INCLUDE (c3, c4)`}},
			},
			{
				Statement: `ALTER TABLE tbl DROP COLUMN c3;`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `ALTER TABLE tbl DROP COLUMN c1;`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 3.3 Test ALTER TABLE SET STATISTICS
 */
CREATE TABLE tbl (c1 int, c2 int);`,
			},
			{
				Statement: `CREATE INDEX tbl_idx ON tbl (c1, (c1+0)) INCLUDE (c2);`,
			},
			{
				Statement:   `ALTER INDEX tbl_idx ALTER COLUMN 1 SET STATISTICS 1000;`,
				ErrorString: `cannot alter statistics on non-expression column "c1" of index "tbl_idx"`,
			},
			{
				Statement: `ALTER INDEX tbl_idx ALTER COLUMN 2 SET STATISTICS 1000;`,
			},
			{
				Statement:   `ALTER INDEX tbl_idx ALTER COLUMN 3 SET STATISTICS 1000;`,
				ErrorString: `cannot alter statistics on included column "c2" of index "tbl_idx"`,
			},
			{
				Statement:   `ALTER INDEX tbl_idx ALTER COLUMN 4 SET STATISTICS 1000;`,
				ErrorString: `column number 4 of relation "tbl_idx" does not exist`,
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 4. CREATE INDEX CONCURRENTLY
 */
CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box, UNIQUE(c1, c2) INCLUDE(c3,c4));`,
			},
			{
				Statement: `INSERT INTO tbl SELECT x, 2*x, 3*x, box('4,4,4,4') FROM generate_series(1,1000) AS x;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX CONCURRENTLY on tbl (c1, c2) INCLUDE (c3, c4);`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX tbl_c1_c2_c3_c4_idx ON public.tbl USING btree (c1, c2) INCLUDE (c3, c4)`}, {`CREATE UNIQUE INDEX tbl_c1_c2_c3_c4_key ON public.tbl USING btree (c1, c2) INCLUDE (c3, c4)`}},
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 5. REINDEX
 */
CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box, UNIQUE(c1, c2) INCLUDE(c3,c4));`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{{`CREATE UNIQUE INDEX tbl_c1_c2_c3_c4_key ON public.tbl USING btree (c1, c2) INCLUDE (c3, c4)`}},
			},
			{
				Statement: `ALTER TABLE tbl DROP COLUMN c3;`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `REINDEX INDEX tbl_c1_c2_c3_c4_key;`,
				ErrorString: `relation "tbl_c1_c2_c3_c4_key" does not exist`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `ALTER TABLE tbl DROP COLUMN c1;`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl' ORDER BY indexname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 7. Check various AMs. All but btree, gist and spgist must fail.
 */
CREATE TABLE tbl (c1 int,c2 int, c3 box, c4 box);`,
			},
			{
				Statement:   `CREATE INDEX on tbl USING brin(c1, c2) INCLUDE (c3, c4);`,
				ErrorString: `access method "brin" does not support included columns`,
			},
			{
				Statement: `CREATE INDEX on tbl USING gist(c3) INCLUDE (c1, c4);`,
			},
			{
				Statement: `CREATE INDEX on tbl USING spgist(c3) INCLUDE (c4);`,
			},
			{
				Statement:   `CREATE INDEX on tbl USING gin(c1, c2) INCLUDE (c3, c4);`,
				ErrorString: `access method "gin" does not support included columns`,
			},
			{
				Statement:   `CREATE INDEX on tbl USING hash(c1, c2) INCLUDE (c3, c4);`,
				ErrorString: `access method "hash" does not support included columns`,
			},
			{
				Statement: `CREATE INDEX on tbl USING rtree(c3) INCLUDE (c1, c4);`,
			},
			{
				Statement: `CREATE INDEX on tbl USING btree(c1, c2) INCLUDE (c3, c4);`,
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 8. Update, delete values in indexed table.
 */
CREATE TABLE tbl (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl SELECT x, 2*x, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX tbl_idx_unique ON tbl using btree(c1, c2) INCLUDE (c3,c4);`,
			},
			{
				Statement: `UPDATE tbl SET c1 = 100 WHERE c1 = 2;`,
			},
			{
				Statement: `UPDATE tbl SET c1 = 1 WHERE c1 = 3;`,
			},
			{
				Statement:   `UPDATE tbl SET c2 = 2 WHERE c1 = 1;`,
				ErrorString: `duplicate key value violates unique constraint "tbl_idx_unique"`,
			},
			{
				Statement: `UPDATE tbl SET c3 = 1;`,
			},
			{
				Statement: `DELETE FROM tbl WHERE c1 = 5 OR c3 = 12;`,
			},
			{
				Statement: `DROP TABLE tbl;`,
			},
			{
				Statement: `/*
 * 9. Alter column type.
 */
CREATE TABLE tbl (c1 int,c2 int, c3 int, c4 box, UNIQUE(c1, c2) INCLUDE(c3,c4));`,
			},
			{
				Statement: `INSERT INTO tbl SELECT x, 2*x, 3*x, box('4,4,4,4') FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `ALTER TABLE tbl ALTER c1 TYPE bigint;`,
			},
			{
				Statement: `ALTER TABLE tbl ALTER c3 TYPE bigint;`,
			},
			{
				Statement: `\d tbl
                Table "public.tbl"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | bigint  |           |          | 
 c2     | integer |           |          | 
 c3     | bigint  |           |          | 
 c4     | box     |           |          | 
Indexes:
    "tbl_c1_c2_c3_c4_key" UNIQUE CONSTRAINT, btree (c1, c2) INCLUDE (c3, c4)
DROP TABLE tbl;`,
			},
		},
	})
}
