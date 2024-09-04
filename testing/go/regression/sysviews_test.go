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

func TestSysviews(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_sysviews)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_sysviews,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `select count(*) >= 0 as ok from pg_available_extension_versions;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) >= 0 as ok from pg_available_extensions;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select name, ident, parent, level, total_bytes >= free_bytes
  from pg_backend_memory_contexts where level = 0;`,
				Results: []sql.Row{{`TopMemoryContext`, ``, ``, 0, true}},
			},
			{
				Statement: `select count(*) > 20 as ok from pg_config;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) = 0 as ok from pg_cursors;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) >= 0 as ok from pg_file_settings;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) > 0 as ok, count(*) FILTER (WHERE error IS NOT NULL) = 0 AS no_err
  from pg_hba_file_rules;`,
				Results: []sql.Row{{true, true}},
			},
			{
				Statement: `select count(*) >= 0 as ok, count(*) FILTER (WHERE error IS NOT NULL) = 0 AS no_err
  from pg_ident_file_mappings;`,
				Results: []sql.Row{{true, true}},
			},
			{
				Statement: `select count(*) > 0 as ok from pg_locks;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) = 0 as ok from pg_prepared_statements;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) >= 0 as ok from pg_prepared_xacts;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) > 0 as ok from pg_stat_slru;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) = 1 as ok from pg_stat_wal;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(*) = 0 as ok from pg_stat_wal_receiver;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select name, setting from pg_settings where name like 'enable%';`,
				Results:   []sql.Row{{`enable_async_append`, `on`}, {`enable_bitmapscan`, `on`}, {`enable_gathermerge`, `on`}, {`enable_hashagg`, `on`}, {`enable_hashjoin`, `on`}, {`enable_incremental_sort`, `on`}, {`enable_indexonlyscan`, `on`}, {`enable_indexscan`, `on`}, {`enable_material`, `on`}, {`enable_memoize`, `on`}, {`enable_mergejoin`, `on`}, {`enable_nestloop`, `on`}, {`enable_parallel_append`, `on`}, {`enable_parallel_hash`, `on`}, {`enable_partition_pruning`, `on`}, {`enable_partitionwise_aggregate`, `off`}, {`enable_partitionwise_join`, `off`}, {`enable_seqscan`, `on`}, {`enable_sort`, `on`}, {`enable_tidscan`, `on`}},
			},
			{
				Statement: `select count(distinct utc_offset) >= 24 as ok from pg_timezone_names;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select count(distinct utc_offset) >= 24 as ok from pg_timezone_abbrevs;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `set timezone_abbreviations = 'Australia';`,
			},
			{
				Statement: `select count(distinct utc_offset) >= 24 as ok from pg_timezone_abbrevs;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `set timezone_abbreviations = 'India';`,
			},
			{
				Statement: `select count(distinct utc_offset) >= 24 as ok from pg_timezone_abbrevs;`,
				Results:   []sql.Row{{true}},
			},
		},
	})
}
