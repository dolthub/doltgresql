// Copyright 2026 Dolthub, Inc.
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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestStatsUsage(t *testing.T) {
	RunScripts(t, StatsUsageTests)
}

// StatsUsageTests verify that table statistics are collected when ANALYZE forces a refresh, that
// the dolt_statistics system table reflects the analyzed data, and that the query planner chooses
// sensible plans for joins and filters over tables of varying sizes and value cardinalities.
//
// Note that the row_count / distinct_count / null_count columns of dolt_statistics currently
// return uint64 values, which sum() rejects ("sum: expected int64, got uint64"), so the queries
// below cast to bigint before aggregating.
var StatsUsageTests = []ScriptTest{
	{
		Name: "dolt_statistics contains reasonable values after ANALYZE",
		SetUpScript: []string{
			"CREATE TABLE big (pk int primary key, lowcard int, highcard int);",
			"CREATE INDEX big_lowcard_idx ON big(lowcard);",
			"CREATE INDEX big_highcard_idx ON big(highcard);",
			"INSERT INTO big SELECT i, i % 10, i FROM generate_series(1, 5000) g(i);",
			"CREATE TABLE small (pk int primary key, c1 int);",
			"INSERT INTO small SELECT i, i % 5 FROM generate_series(1, 10) g(i);",
			"ANALYZE big;",
			"ANALYZE small;",
		},
		Assertions: []ScriptTestAssertion{
			{
				// Every index on big should account for all 5000 rows across its histogram buckets
				Query: "SELECT index_name, sum(row_count::bigint)::int FROM dolt_statistics WHERE table_name = 'big' GROUP BY index_name ORDER BY index_name;",
				Expected: []sql.Row{
					{"big_highcard_idx", 5000},
					{"big_lowcard_idx", 5000},
					{"primary", 5000},
				},
			},
			{
				// The primary key is unique, so distinct count should equal row count
				Query:    "SELECT sum(distinct_count::bigint)::int FROM dolt_statistics WHERE table_name = 'big' AND index_name = 'primary';",
				Expected: []sql.Row{{5000}},
			},
			{
				// highcard is also unique
				Query:    "SELECT sum(distinct_count::bigint)::int FROM dolt_statistics WHERE table_name = 'big' AND index_name = 'big_highcard_idx';",
				Expected: []sql.Row{{5000}},
			},
			{
				// lowcard has only 10 distinct values, so no bucket can see more than 10 of them
				Query:    "SELECT max(distinct_count::bigint) <= 10 AND sum(distinct_count::bigint) >= 10 FROM dolt_statistics WHERE table_name = 'big' AND index_name = 'big_lowcard_idx';",
				Expected: []sql.Row{{"t"}},
			},
			{
				// No NULL values were inserted anywhere
				Query:    "SELECT sum(null_count::bigint)::int FROM dolt_statistics WHERE table_name IN ('big', 'small');",
				Expected: []sql.Row{{0}},
			},
			{
				// The columns tracked for each index are correct
				Query: "SELECT DISTINCT index_name, columns FROM dolt_statistics WHERE table_name = 'big' ORDER BY index_name;",
				Expected: []sql.Row{
					{"big_highcard_idx", "highcard"},
					{"big_lowcard_idx", "lowcard"},
					{"primary", "pk"},
				},
			},
			{
				// The largest upper bound over the primary index buckets is the max pk value
				Query:    "SELECT max(upper_bound::int) FROM dolt_statistics WHERE table_name = 'big' AND index_name = 'primary';",
				Expected: []sql.Row{{5000}},
			},
			{
				// The small table's stats account for all of its rows too
				Query: "SELECT index_name, sum(row_count::bigint)::int, sum(distinct_count::bigint)::int FROM dolt_statistics WHERE table_name = 'small' GROUP BY index_name ORDER BY index_name;",
				Expected: []sql.Row{
					{"primary", 10, 10},
				},
			},
		},
	},
	{
		Name: "ANALYZE refreshes statistics after data changes",
		SetUpScript: []string{
			"CREATE TABLE t (pk int primary key, c1 int);",
			"INSERT INTO t SELECT i, i % 7 FROM generate_series(1, 100) g(i);",
			"ANALYZE t;",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT sum(row_count::bigint)::int FROM dolt_statistics WHERE table_name = 't';",
				Expected: []sql.Row{{100}},
			},
			{
				Query:    "INSERT INTO t SELECT i, i % 7 FROM generate_series(101, 300) g(i);",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE t;",
				Expected: []sql.Row{},
			},
			{
				// The refreshed stats reflect the new row count
				Query:    "SELECT sum(row_count::bigint)::int FROM dolt_statistics WHERE table_name = 't';",
				Expected: []sql.Row{{300}},
			},
		},
	},
	{
		Name: "join planner puts small table on the scan side of a lookup join",
		SetUpScript: []string{
			"CREATE TABLE big (pk int primary key, lowcard int, highcard int);",
			"CREATE INDEX big_lowcard_idx ON big(lowcard);",
			"CREATE INDEX big_highcard_idx ON big(highcard);",
			"INSERT INTO big SELECT i, i % 10, i FROM generate_series(1, 5000) g(i);",
			"CREATE TABLE small (pk int primary key, c1 int);",
			"INSERT INTO small SELECT i, i % 5 FROM generate_series(1, 10) g(i);",
			"ANALYZE big;",
			"ANALYZE small;",
		},
		Assertions: []ScriptTestAssertion{
			{
				// The planner should scan the small table (10 rows) and use the index on the
				// big table (5000 rows) for lookups, not the other way around.
				Query: "EXPLAIN SELECT * FROM small JOIN big ON small.pk = big.highcard;",
				Expected: []sql.Row{
					{"LookupJoin"},
					{" ├─ Table"},
					{" │   ├─ name: small"},
					{" │   └─ columns: [pk c1]"},
					{" └─ IndexedTableAccess(big)"},
					{"     ├─ index: [big.highcard]"},
					{"     ├─ columns: [pk lowcard highcard]"},
					{"     └─ keys: small.pk"},
				},
			},
			{
				// With the FROM order reversed, the planner should still put the small table on
				// the scan side of the join. The Project node restores the SELECT * column order,
				// which shows the join inputs were reordered from the query text.
				Query: "EXPLAIN SELECT * FROM big JOIN small ON small.pk = big.highcard;",
				Expected: []sql.Row{
					{"Project"},
					{" ├─ columns: [big.pk, big.lowcard, big.highcard, small.pk, small.c1]"},
					{" └─ LookupJoin"},
					{"     ├─ Table"},
					{"     │   ├─ name: small"},
					{"     │   └─ columns: [pk c1]"},
					{"     └─ IndexedTableAccess(big)"},
					{"         ├─ index: [big.highcard]"},
					{"         ├─ columns: [pk lowcard highcard]"},
					{"         └─ keys: small.pk"},
				},
			},
		},
	},
	{
		Name: "join planner builds hash table from the small table in a hash join",
		SetUpScript: []string{
			// No indexes on the join columns, so the planner must use a hash join, and
			// which side gets materialized into the hash table depends on relative size.
			"CREATE TABLE big (pk int primary key, val int);",
			"INSERT INTO big SELECT i, i % 100 FROM generate_series(1, 5000) g(i);",
			"CREATE TABLE small (pk int primary key, val int);",
			"INSERT INTO small SELECT i, i FROM generate_series(1, 10) g(i);",
			"ANALYZE big;",
			"ANALYZE small;",
		},
		Assertions: []ScriptTestAssertion{
			{
				// The hash table (the HashLookup child) should be built from the small table
				Query: "EXPLAIN SELECT * FROM big JOIN small ON big.val = small.val;",
				Expected: []sql.Row{
					{"HashJoin"},
					{" ├─ big.val = small.val"},
					{" ├─ Table"},
					{" │   ├─ name: big"},
					{" │   └─ columns: [pk val]"},
					{" └─ HashLookup"},
					{"     ├─ left-key: (big.val)"},
					{"     ├─ right-key: (small.val)"},
					{"     └─ Table"},
					{"         ├─ name: small"},
					{"         └─ columns: [pk val]"},
				},
			},
			{
				// Same plan when the FROM order is reversed
				Query: "EXPLAIN SELECT * FROM small JOIN big ON big.val = small.val;",
				Expected: []sql.Row{
					{"Project"},
					{" ├─ columns: [small.pk, small.val, big.pk, big.val]"},
					{" └─ HashJoin"},
					{"     ├─ big.val = small.val"},
					{"     ├─ Table"},
					{"     │   ├─ name: big"},
					{"     │   └─ columns: [pk val]"},
					{"     └─ HashLookup"},
					{"         ├─ left-key: (big.val)"},
					{"         ├─ right-key: (small.val)"},
					{"         └─ Table"},
					{"             ├─ name: small"},
					{"             └─ columns: [pk val]"},
				},
			},
		},
	},
	{
		Name: "planner uses cardinality to select more selective index",
		SetUpScript: []string{
			"CREATE TABLE t (pk int primary key, lowcard int, highcard int);",
			"CREATE INDEX t_lowcard_idx ON t(lowcard);",
			"CREATE INDEX t_highcard_idx ON t(highcard);",
			"INSERT INTO t SELECT i, i % 10, i FROM generate_series(1, 5000) g(i);",
			"ANALYZE t;",
		},
		Assertions: []ScriptTestAssertion{
			{
				// With a choice between an index that matches ~500 rows (lowcard = 3) and one
				// that matches 1 row (highcard = 42), the planner should pick the
				// high-cardinality index. The matched filter is converted to an index range and
				// removed from the Filter node above the scan.
				Query: "EXPLAIN SELECT * FROM t WHERE lowcard = 3 AND highcard = 42;",
				Expected: []sql.Row{
					{"Filter"},
					{" ├─ t.lowcard = 3"},
					{" └─ IndexedTableAccess(t)"},
					{"     ├─ index: [t.highcard]"},
					{"     ├─ filters: [{[42, 42]}]"},
					{"     └─ columns: [pk lowcard highcard]"},
				},
			},
			{
				// When the highcard predicate matches every row (highcard > 0) and the lowcard
				// predicate matches ~500 rows, the index choice should flip to lowcard.
				Query: "EXPLAIN SELECT * FROM t WHERE lowcard = 3 AND highcard > 0;",
				Expected: []sql.Row{
					{"Filter"},
					{" ├─ t.highcard > 0"},
					{" └─ IndexedTableAccess(t)"},
					{"     ├─ index: [t.lowcard]"},
					{"     ├─ filters: [{[3, 3]}]"},
					{"     └─ columns: [pk lowcard highcard]"},
				},
			},
		},
	},
	{
		// This test distinguishes stats-informed join planning from the size heuristics the
		// planner falls back to without statistics. Table big has 5000 rows but the indexed
		// filter val > 4950 matches only 50, fewer than small's 1000 rows, so a planner that
		// consults the val histogram scans filtered big on the outside of the lookup join and
		// probes small's primary key. Dolt produces exactly this plan for the equivalent MySQL
		// script. Doltgres instead keeps small on the scan side, because join costing looks up
		// row counts and index statistics with an empty schema name while Doltgres stats are
		// keyed under schema 'public' (memo/rel_props.go RowCount and indexed_joins.go GetStats
		// in go-mysql-server), so join planning currently never sees collected statistics.
		// Unskip once join costing resolves the table's schema when querying the stats provider.
		Name: "join planner uses histograms to reorder a filtered lookup join",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE big (pk int primary key, val int, jc int);",
			"CREATE INDEX big_val_idx ON big(val);",
			"CREATE INDEX big_jc_idx ON big(jc);",
			"INSERT INTO big SELECT i, i, i % 1000 + 1 FROM generate_series(1, 5000) g(i);",
			"CREATE TABLE small (pk int primary key, c1 int);",
			"INSERT INTO small SELECT i, i FROM generate_series(1, 1000) g(i);",
			"ANALYZE big;",
			"ANALYZE small;",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "EXPLAIN SELECT * FROM small JOIN big ON big.jc = small.pk WHERE big.val > 4950;",
				Expected: []sql.Row{
					{"Project"},
					{" ├─ columns: [small.pk, small.c1, big.pk, big.val, big.jc]"},
					{" └─ LookupJoin"},
					{"     ├─ IndexedTableAccess(big)"},
					{"     │   ├─ index: [big.val]"},
					{"     │   ├─ filters: [{(4950, ∞)}]"},
					{"     │   └─ columns: [pk val jc]"},
					{"     └─ IndexedTableAccess(small)"},
					{"         ├─ index: [small.pk]"},
					{"         ├─ columns: [pk c1]"},
					{"         └─ keys: big.jc"},
				},
			},
		},
	},
}
