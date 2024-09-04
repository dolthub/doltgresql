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

func TestVacuumParallel(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_vacuum_parallel)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_vacuum_parallel,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup, RegressionFileName_write_parallel},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET max_parallel_maintenance_workers TO 4;`,
			},
			{
				Statement: `SET min_parallel_index_scan_size TO '128kB';`,
			},
			{
				Statement: `CREATE TABLE parallel_vacuum_table (a int) WITH (autovacuum_enabled = off);`,
			},
			{
				Statement: `INSERT INTO parallel_vacuum_table SELECT i from generate_series(1, 10000) i;`,
			},
			{
				Statement: `CREATE INDEX regular_sized_index ON parallel_vacuum_table(a);`,
			},
			{
				Statement: `CREATE INDEX typically_sized_index ON parallel_vacuum_table(a);`,
			},
			{
				Statement: `CREATE INDEX vacuum_in_leader_small_index ON parallel_vacuum_table((1));`,
			},
			{
				Statement: `SELECT EXISTS (
SELECT 1
FROM pg_class
WHERE oid = 'vacuum_in_leader_small_index'::regclass AND
  pg_relation_size(oid) <
  pg_size_bytes(current_setting('min_parallel_index_scan_size'))
) as leader_will_handle_small_index;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT count(*) as trigger_parallel_vacuum_nindexes
FROM pg_class
WHERE oid in ('regular_sized_index'::regclass, 'typically_sized_index'::regclass) AND
  pg_relation_size(oid) >=
  pg_size_bytes(current_setting('min_parallel_index_scan_size'));`,
				Results: []sql.Row{{2}},
			},
			{
				Statement: `DELETE FROM parallel_vacuum_table;`,
			},
			{
				Statement: `VACUUM (PARALLEL 4, INDEX_CLEANUP ON) parallel_vacuum_table;`,
			},
			{
				Statement: `INSERT INTO parallel_vacuum_table SELECT i FROM generate_series(1, 10000) i;`,
			},
			{
				Statement: `RESET max_parallel_maintenance_workers;`,
			},
			{
				Statement: `RESET min_parallel_index_scan_size;`,
			},
		},
	})
}
