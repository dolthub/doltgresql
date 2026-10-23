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

// TestPostgresDoltDiffFunctions isolates the row-level diff cases of
// TestDoltDiff, TestDoltDiffStat and TestDoltDiffSummary from their mixed
// schema-feature setup scripts (array keys, serial, defaults, foreign keys,
// generated columns, triggers), which the PostgreSQL fork's storage does not
// support yet. The expectations hold on both engines.
func TestPostgresDoltDiffFunctions(t *testing.T) {
	setup := []string{
		"CREATE TABLE t_simple (pk INT4 PRIMARY KEY);",
		"CREATE TABLE t_composite (pk1 INT2, pk2 INT8, PRIMARY KEY(pk1, pk2));",
		"CREATE TABLE t_values (pk INT4 PRIMARY KEY, v1 INT8, v2 TEXT);",
		"CREATE TABLE t_keyless (v1 INT4);",
		"INSERT INTO t_simple VALUES (1), (2), (3);",
		"INSERT INTO t_composite VALUES (1, 100), (1, 101), (2, 100), (2, 101);",
		"INSERT INTO t_values VALUES (1, 10, 'a'), (2, 20, 'b');",
		"INSERT INTO t_keyless VALUES (7), (8), (9);",
		"SELECT DOLT_COMMIT('-A', '-m', 'initial');",
		"SELECT DOLT_BRANCH('original');",
		"INSERT INTO t_simple VALUES (4);",
		"INSERT INTO t_composite VALUES (3, 100);",
		"UPDATE t_values SET v1 = 11 WHERE pk = 1;",
		"DELETE FROM t_values WHERE pk = 2;",
		"INSERT INTO t_keyless VALUES (10);",
		"SELECT DOLT_COMMIT('-A', '-m', 'changes');",
	}
	RunScripts(t, []ScriptTest{
		{
			Name:        "dolt_diff rows between branches",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT from_pk, to_pk, diff_type FROM DOLT_DIFF('main', 'original', 't_simple');",
					Expected: []sql.Row{{4, nil, "removed"}},
				},
				{
					Query:    "SELECT from_pk1, from_pk2 FROM DOLT_DIFF('main', 'original', 't_composite');",
					Expected: []sql.Row{{3, 100}},
				},
				{
					Query:    "SELECT diff_type, from_pk, from_v1, to_pk, to_v1, to_v2 FROM DOLT_DIFF('original', 'main', 't_values') ORDER BY coalesce(to_pk, from_pk);",
					Expected: []sql.Row{{"modified", 1, 10, 1, 11, "a"}, {"removed", 2, 20, nil, nil, nil}},
				},
				{
					Query:    "SELECT diff_type, from_v1, to_v1 FROM DOLT_DIFF('original', 'main', 't_keyless');",
					Expected: []sql.Row{{"added", nil, 10}},
				},
				{
					Query:    "SELECT diff_type, to_pk FROM DOLT_DIFF('original..main', 't_simple');",
					Expected: []sql.Row{{"added", 4}},
				},
				{
					Query:    "SELECT to_commit, from_commit FROM DOLT_DIFF('original', 'main', 't_simple');",
					Expected: []sql.Row{{"main", "original"}},
				},
				{
					Query:    "SELECT count(*) FROM DOLT_DIFF('main', 'main', 't_simple');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:       "SELECT * FROM DOLT_DIFF('main', 'original', 'missing');",
					ExpectedErr: "table not found: public.missing",
				},
			},
		},
		{
			Name:        "dolt_diff_stat and dolt_diff_summary between branches",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT COUNT(*) FROM DOLT_DIFF_STAT('main', 'original');",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT table_name, rows_unmodified, rows_added, rows_deleted, rows_modified, old_row_count, new_row_count FROM DOLT_DIFF_STAT('original', 'main', 't_values');",
					Expected: []sql.Row{{"public.t_values", 0, 0, 1, 1, 2, 1}},
				},
				{
					Query:    "SELECT table_name, rows_added, cells_added, new_cell_count FROM DOLT_DIFF_STAT('original', 'main', 't_composite');",
					Expected: []sql.Row{{"public.t_composite", 1, 2, 10}},
				},
				{
					Query:    "SELECT table_name, rows_added, rows_deleted, old_row_count, new_row_count FROM DOLT_DIFF_STAT('original', 'main', 't_keyless');",
					Expected: []sql.Row{{"public.t_keyless", 1, 0, 3, 4}},
				},
				{
					Query:    "SELECT COUNT(*) FROM DOLT_DIFF_SUMMARY('main', 'original');",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT from_table_name, to_table_name, diff_type, data_change, schema_change FROM DOLT_DIFF_SUMMARY('main', 'original', 't_simple');",
					Expected: []sql.Row{{"public.t_simple", "public.t_simple", "modified", "t", "f"}},
				},
			},
		},
	})
}
