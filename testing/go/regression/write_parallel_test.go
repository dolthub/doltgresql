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

func TestWriteParallel(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_write_parallel)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_write_parallel,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup, RegressionFileName_select_parallel},
		Statements: []RegressionFileStatement{
			{
				Statement: `begin;`,
			},
			{
				Statement: `set parallel_setup_cost=0;`,
			},
			{
				Statement: `set parallel_tuple_cost=0;`,
			},
			{
				Statement: `set min_parallel_table_scan_size=0;`,
			},
			{
				Statement: `set max_parallel_workers_per_gather=4;`,
			},
			{
				Statement: `explain (costs off) create table parallel_write as
    select length(stringu1) from tenk1 group by length(stringu1);`,
				Results: []sql.Row{{`Finalize HashAggregate`}, {`Group Key: (length((stringu1)::text))`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial HashAggregate`}, {`Group Key: length((stringu1)::text)`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `create table parallel_write as
    select length(stringu1) from tenk1 group by length(stringu1);`,
			},
			{
				Statement: `drop table parallel_write;`,
			},
			{
				Statement: `explain (costs off) select length(stringu1) into parallel_write
    from tenk1 group by length(stringu1);`,
				Results: []sql.Row{{`Finalize HashAggregate`}, {`Group Key: (length((stringu1)::text))`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial HashAggregate`}, {`Group Key: length((stringu1)::text)`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `select length(stringu1) into parallel_write
    from tenk1 group by length(stringu1);`,
			},
			{
				Statement: `drop table parallel_write;`,
			},
			{
				Statement: `explain (costs off) create materialized view parallel_mat_view as
    select length(stringu1) from tenk1 group by length(stringu1);`,
				Results: []sql.Row{{`Finalize HashAggregate`}, {`Group Key: (length((stringu1)::text))`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial HashAggregate`}, {`Group Key: length((stringu1)::text)`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `create materialized view parallel_mat_view as
    select length(stringu1) from tenk1 group by length(stringu1);`,
			},
			{
				Statement: `create unique index on parallel_mat_view(length);`,
			},
			{
				Statement: `refresh materialized view parallel_mat_view;`,
			},
			{
				Statement: `refresh materialized view concurrently parallel_mat_view;`,
			},
			{
				Statement: `drop materialized view parallel_mat_view;`,
			},
			{
				Statement: `prepare prep_stmt as select length(stringu1) from tenk1 group by length(stringu1);`,
			},
			{
				Statement: `explain (costs off) create table parallel_write as execute prep_stmt;`,
				Results:   []sql.Row{{`Finalize HashAggregate`}, {`Group Key: (length((stringu1)::text))`}, {`->  Gather`}, {`Workers Planned: 4`}, {`->  Partial HashAggregate`}, {`Group Key: length((stringu1)::text)`}, {`->  Parallel Seq Scan on tenk1`}},
			},
			{
				Statement: `create table parallel_write as execute prep_stmt;`,
			},
			{
				Statement: `drop table parallel_write;`,
			},
			{
				Statement: `rollback;`,
			},
		},
	})
}
