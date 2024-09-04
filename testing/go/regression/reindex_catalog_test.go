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
)

func TestReindexCatalog(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_reindex_catalog)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_reindex_catalog,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `REINDEX TABLE pg_class; -- mapped, non-shared, critical`,
			},
			{
				Statement: `REINDEX TABLE pg_index; -- non-mapped, non-shared, critical`,
			},
			{
				Statement: `REINDEX TABLE pg_operator; -- non-mapped, non-shared, critical`,
			},
			{
				Statement: `REINDEX TABLE pg_database; -- mapped, shared, critical`,
			},
			{
				Statement: `REINDEX TABLE pg_shdescription; -- mapped, shared non-critical`,
			},
			{
				Statement: `REINDEX INDEX pg_class_oid_index; -- mapped, non-shared, critical`,
			},
			{
				Statement: `REINDEX INDEX pg_class_relname_nsp_index; -- mapped, non-shared, non-critical`,
			},
			{
				Statement: `REINDEX INDEX pg_index_indexrelid_index; -- non-mapped, non-shared, critical`,
			},
			{
				Statement: `REINDEX INDEX pg_index_indrelid_index; -- non-mapped, non-shared, non-critical`,
			},
			{
				Statement: `REINDEX INDEX pg_database_oid_index; -- mapped, shared, critical`,
			},
			{
				Statement: `REINDEX INDEX pg_shdescription_o_c_index; -- mapped, shared, non-critical`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET min_parallel_table_scan_size = 0;`,
			},
			{
				Statement: `REINDEX INDEX pg_class_oid_index; -- mapped, non-shared, critical`,
			},
			{
				Statement: `REINDEX INDEX pg_class_relname_nsp_index; -- mapped, non-shared, non-critical`,
			},
			{
				Statement: `REINDEX INDEX pg_index_indexrelid_index; -- non-mapped, non-shared, critical`,
			},
			{
				Statement: `REINDEX INDEX pg_index_indrelid_index; -- non-mapped, non-shared, non-critical`,
			},
			{
				Statement: `REINDEX INDEX pg_database_oid_index; -- mapped, shared, critical`,
			},
			{
				Statement: `REINDEX INDEX pg_shdescription_o_c_index; -- mapped, shared, non-critical`,
			},
			{
				Statement: `ROLLBACK;`,
			},
		},
	})
}
