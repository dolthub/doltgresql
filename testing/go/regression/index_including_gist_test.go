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

func TestIndexIncludingGist(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_index_including_gist)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_index_including_gist,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `/*
 * 1.1. test CREATE INDEX with buffered build
 */
CREATE TABLE tbl_gist (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_gist SELECT x, 2*x, 3*x, box(point(x,x+1),point(2*x,2*x+1)) FROM generate_series(1,8000) AS x;`,
			},
			{
				Statement: `CREATE INDEX tbl_gist_idx ON tbl_gist using gist (c4) INCLUDE (c1,c2,c3);`,
			},
			{
				Statement: `SELECT pg_get_indexdef(i.indexrelid)
FROM pg_index i JOIN pg_class c ON i.indexrelid = c.oid
WHERE i.indrelid = 'tbl_gist'::regclass ORDER BY c.relname;`,
				Results: []sql.Row{{`CREATE INDEX tbl_gist_idx ON public.tbl_gist USING gist (c4) INCLUDE (c1, c2, c3)`}},
			},
			{
				Statement: `SELECT * FROM tbl_gist where c4 <@ box(point(1,1),point(10,10));`,
				Results:   []sql.Row{{1, 2, 3, `(2,3),(1,2)`}, {2, 4, 6, `(4,5),(2,3)`}, {3, 6, 9, `(6,7),(3,4)`}, {4, 8, 12, `(8,9),(4,5)`}},
			},
			{
				Statement: `SET enable_bitmapscan TO off;`,
			},
			{
				Statement: `EXPLAIN  (costs off) SELECT * FROM tbl_gist where c4 <@ box(point(1,1),point(10,10));`,
				Results:   []sql.Row{{`Index Only Scan using tbl_gist_idx on tbl_gist`}, {`Index Cond: (c4 <@ '(10,10),(1,1)'::box)`}},
			},
			{
				Statement: `SET enable_bitmapscan TO default;`,
			},
			{
				Statement: `DROP TABLE tbl_gist;`,
			},
			{
				Statement: `/*
 * 1.2. test CREATE INDEX with inserts
 */
CREATE TABLE tbl_gist (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `CREATE INDEX tbl_gist_idx ON tbl_gist using gist (c4) INCLUDE (c1,c2,c3);`,
			},
			{
				Statement: `INSERT INTO tbl_gist SELECT x, 2*x, 3*x, box(point(x,x+1),point(2*x,2*x+1)) FROM generate_series(1,8000) AS x;`,
			},
			{
				Statement: `SELECT pg_get_indexdef(i.indexrelid)
FROM pg_index i JOIN pg_class c ON i.indexrelid = c.oid
WHERE i.indrelid = 'tbl_gist'::regclass ORDER BY c.relname;`,
				Results: []sql.Row{{`CREATE INDEX tbl_gist_idx ON public.tbl_gist USING gist (c4) INCLUDE (c1, c2, c3)`}},
			},
			{
				Statement: `SELECT * FROM tbl_gist where c4 <@ box(point(1,1),point(10,10));`,
				Results:   []sql.Row{{1, 2, 3, `(2,3),(1,2)`}, {2, 4, 6, `(4,5),(2,3)`}, {3, 6, 9, `(6,7),(3,4)`}, {4, 8, 12, `(8,9),(4,5)`}},
			},
			{
				Statement: `SET enable_bitmapscan TO off;`,
			},
			{
				Statement: `EXPLAIN  (costs off) SELECT * FROM tbl_gist where c4 <@ box(point(1,1),point(10,10));`,
				Results:   []sql.Row{{`Index Only Scan using tbl_gist_idx on tbl_gist`}, {`Index Cond: (c4 <@ '(10,10),(1,1)'::box)`}},
			},
			{
				Statement: `SET enable_bitmapscan TO default;`,
			},
			{
				Statement: `DROP TABLE tbl_gist;`,
			},
			{
				Statement: `/*
 * 2. CREATE INDEX CONCURRENTLY
 */
CREATE TABLE tbl_gist (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_gist SELECT x, 2*x, 3*x, box(point(x,x+1),point(2*x,2*x+1)) FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY tbl_gist_idx ON tbl_gist using gist (c4) INCLUDE (c1,c2,c3);`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl_gist' ORDER BY indexname;`,
				Results:   []sql.Row{{`CREATE INDEX tbl_gist_idx ON public.tbl_gist USING gist (c4) INCLUDE (c1, c2, c3)`}},
			},
			{
				Statement: `DROP TABLE tbl_gist;`,
			},
			{
				Statement: `/*
 * 3. REINDEX
 */
CREATE TABLE tbl_gist (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_gist SELECT x, 2*x, 3*x, box(point(x,x+1),point(2*x,2*x+1)) FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `CREATE INDEX tbl_gist_idx ON tbl_gist using gist (c4) INCLUDE (c1,c3);`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl_gist' ORDER BY indexname;`,
				Results:   []sql.Row{{`CREATE INDEX tbl_gist_idx ON public.tbl_gist USING gist (c4) INCLUDE (c1, c3)`}},
			},
			{
				Statement: `REINDEX INDEX tbl_gist_idx;`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl_gist' ORDER BY indexname;`,
				Results:   []sql.Row{{`CREATE INDEX tbl_gist_idx ON public.tbl_gist USING gist (c4) INCLUDE (c1, c3)`}},
			},
			{
				Statement: `ALTER TABLE tbl_gist DROP COLUMN c1;`,
			},
			{
				Statement: `SELECT indexdef FROM pg_indexes WHERE tablename = 'tbl_gist' ORDER BY indexname;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE tbl_gist;`,
			},
			{
				Statement: `/*
 * 4. Update, delete values in indexed table.
 */
CREATE TABLE tbl_gist (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_gist SELECT x, 2*x, 3*x, box(point(x,x+1),point(2*x,2*x+1)) FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `CREATE INDEX tbl_gist_idx ON tbl_gist using gist (c4) INCLUDE (c1,c3);`,
			},
			{
				Statement: `UPDATE tbl_gist SET c1 = 100 WHERE c1 = 2;`,
			},
			{
				Statement: `UPDATE tbl_gist SET c1 = 1 WHERE c1 = 3;`,
			},
			{
				Statement: `DELETE FROM tbl_gist WHERE c1 = 5 OR c3 = 12;`,
			},
			{
				Statement: `DROP TABLE tbl_gist;`,
			},
			{
				Statement: `/*
 * 5. Alter column type.
 */
CREATE TABLE tbl_gist (c1 int, c2 int, c3 int, c4 box);`,
			},
			{
				Statement: `INSERT INTO tbl_gist SELECT x, 2*x, 3*x, box(point(x,x+1),point(2*x,2*x+1)) FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `CREATE INDEX tbl_gist_idx ON tbl_gist using gist (c4) INCLUDE (c1,c3);`,
			},
			{
				Statement: `ALTER TABLE tbl_gist ALTER c1 TYPE bigint;`,
			},
			{
				Statement: `ALTER TABLE tbl_gist ALTER c3 TYPE bigint;`,
			},
			{
				Statement: `\d tbl_gist
              Table "public.tbl_gist"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | bigint  |           |          | 
 c2     | integer |           |          | 
 c3     | bigint  |           |          | 
 c4     | box     |           |          | 
Indexes:
    "tbl_gist_idx" gist (c4) INCLUDE (c1, c3)
DROP TABLE tbl_gist;`,
			},
			{
				Statement: `/*
 * 6. EXCLUDE constraint.
 */
CREATE TABLE tbl_gist (c1 int, c2 int, c3 int, c4 box, EXCLUDE USING gist (c4 WITH &&) INCLUDE (c1, c2, c3));`,
			},
			{
				Statement:   `INSERT INTO tbl_gist SELECT x, 2*x, 3*x, box(point(x,x+1),point(2*x,2*x+1)) FROM generate_series(1,10) AS x;`,
				ErrorString: `conflicting key value violates exclusion constraint "tbl_gist_c4_c1_c2_c3_excl"`,
			},
			{
				Statement: `INSERT INTO tbl_gist SELECT x, 2*x, 3*x, box(point(3*x,2*x),point(3*x+1,2*x+1)) FROM generate_series(1,10) AS x;`,
			},
			{
				Statement: `EXPLAIN  (costs off) SELECT * FROM tbl_gist where c4 <@ box(point(1,1),point(10,10));`,
				Results:   []sql.Row{{`Index Only Scan using tbl_gist_c4_c1_c2_c3_excl on tbl_gist`}, {`Index Cond: (c4 <@ '(10,10),(1,1)'::box)`}},
			},
			{
				Statement: `\d tbl_gist
              Table "public.tbl_gist"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 c1     | integer |           |          | 
 c2     | integer |           |          | 
 c3     | integer |           |          | 
 c4     | box     |           |          | 
Indexes:
    "tbl_gist_c4_c1_c2_c3_excl" EXCLUDE USING gist (c4 WITH &&) INCLUDE (c1, c2, c3)
DROP TABLE tbl_gist;`,
			},
		},
	})
}
