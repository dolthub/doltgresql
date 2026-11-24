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

// TestDoltTableFunctionAuth checks that Dolt's table functions require SELECT on the tables whose data or metadata
// they return.
func TestDoltTableFunctionAuth(t *testing.T) {
	// denied asserts that the limited user is refused because of the given table
	denied := func(query string, table string, bindVars ...any) ScriptTestAssertion {
		return ScriptTestAssertion{
			Query:           query,
			BindVars:        bindVars,
			Username:        `limited`,
			Password:        `password`,
			ExpectedErr:     "permission denied for table " + table,
			ExpectedErrCode: "42501",
		}
	}
	// allowed asserts that the limited user gets the expected rows
	allowed := func(query string, expected []sql.Row, bindVars ...any) ScriptTestAssertion {
		return ScriptTestAssertion{
			Query:    query,
			BindVars: bindVars,
			Username: `limited`,
			Password: `password`,
			Expected: expected,
		}
	}
	setUp := []string{
		`CREATE USER limited PASSWORD 'password';`,
		`CREATE TABLE secret (pk INT PRIMARY KEY, ssn TEXT);`,
		`INSERT INTO secret VALUES (1, '111');`,
		`CREATE TABLE open_t (pk INT PRIMARY KEY, v TEXT);`,
		`INSERT INTO open_t VALUES (1, 'a');`,
		`CREATE SCHEMA s2;`,
		`CREATE TABLE s2.hidden (pk INT PRIMARY KEY, ssn TEXT);`,
		`INSERT INTO s2.hidden VALUES (1, '333');`,
		`GRANT SELECT ON open_t TO limited;`,
		`SELECT dolt_commit('-Am', 'init');`,
		`SELECT dolt_branch('other');`,
		`UPDATE secret SET ssn = '222';`,
		`UPDATE open_t SET v = 'b';`,
		`UPDATE s2.hidden SET ssn = '444';`,
	}
	RunScripts(t, []ScriptTest{
		{
			Name:        "table functions on a table without a grant",
			SetUpScript: setUp,
			Assertions: []ScriptTestAssertion{
				denied(`SELECT to_ssn FROM dolt_diff('HEAD', 'WORKING', 'secret');`, "secret"),
				denied(`SELECT to_ssn FROM dolt_diff('HEAD..WORKING', 'secret');`, "secret"),
				denied(`SELECT statement FROM dolt_patch('HEAD', 'WORKING', 'secret');`, "secret"),
				denied(`SELECT statement FROM dolt_patch('HEAD', 'WORKING', $1);`, "secret", "secret"),
				denied(`SELECT table_name FROM dolt_diff_stat('HEAD', 'WORKING', 'secret');`, "secret"),
				denied(`SELECT to_table_name FROM dolt_diff_summary('HEAD', 'WORKING', 'secret');`, "secret"),
				denied(`SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'secret');`, "secret"),
				denied(`SELECT * FROM dolt_preview_merge_conflicts('main', 'other', 'secret');`, "secret"),
			},
		},
		{
			Name:        "table functions on a table with a grant",
			SetUpScript: setUp,
			Assertions: []ScriptTestAssertion{
				allowed(`SELECT from_v, to_v FROM dolt_diff('HEAD', 'WORKING', 'open_t');`, []sql.Row{{"a", "b"}}),
				allowed(`SELECT from_v, to_v FROM dolt_diff('HEAD..WORKING', 'open_t');`, []sql.Row{{"a", "b"}}),
				allowed(`SELECT statement FROM dolt_patch('HEAD', 'WORKING', 'open_t');`,
					[]sql.Row{{`UPDATE "open_t" SET "v"='b' WHERE "pk"=1;`}}),
				allowed(`SELECT statement FROM dolt_patch('HEAD', 'WORKING', $1);`,
					[]sql.Row{{`UPDATE "open_t" SET "v"='b' WHERE "pk"=1;`}}, "open_t"),
				allowed(`SELECT table_name, rows_modified FROM dolt_diff_stat('HEAD', 'WORKING', 'open_t');`,
					[]sql.Row{{"public.open_t", 1}}),
				allowed(`SELECT to_table_name, diff_type FROM dolt_diff_summary('HEAD', 'WORKING', 'open_t');`,
					[]sql.Row{{"public.open_t", "modified"}}),
				allowed(`SELECT count(*) FROM dolt_schema_diff('HEAD', 'WORKING', 'open_t');`, []sql.Row{{0}}),
			},
		},
		{
			Name:        "table functions that read every table, including tables in other schemas",
			SetUpScript: setUp,
			Assertions: []ScriptTestAssertion{
				denied(`SELECT statement FROM dolt_patch('HEAD', 'WORKING');`, "secret"),
				denied(`SELECT * FROM dolt_diff_stat('HEAD..WORKING');`, "secret"),
				denied(`SELECT * FROM dolt_diff_summary('HEAD', 'WORKING');`, "secret"),
				denied(`SELECT * FROM dolt_schema_diff('HEAD', 'WORKING');`, "secret"),
				denied(`SELECT * FROM dolt_preview_merge_conflicts_summary('main', 'other');`, "secret"),
				denied(`SELECT count(*) FROM dolt_log();`, "secret"),
				denied(`SELECT * FROM dolt_merge_bases('main', 'other');`, "secret"),
				{
					Query:    `GRANT SELECT ON secret TO limited;`,
					Expected: []sql.Row{},
				},
				// The public schema is now readable, but s2.hidden is not
				denied(`SELECT statement FROM dolt_patch('HEAD', 'WORKING');`, "hidden"),
				denied(`SELECT * FROM dolt_diff_stat('HEAD..WORKING');`, "hidden"),
				denied(`SELECT * FROM dolt_diff_summary('HEAD', 'WORKING');`, "hidden"),
				denied(`SELECT * FROM dolt_schema_diff('HEAD', 'WORKING');`, "hidden"),
				denied(`SELECT count(*) FROM dolt_log();`, "hidden"),
				{
					Query:    `GRANT SELECT ON s2.hidden TO limited;`,
					Expected: []sql.Row{},
				},
				allowed(`SELECT count(*) FROM dolt_patch('HEAD', 'WORKING');`, []sql.Row{{3}}),
				allowed(`SELECT count(*) FROM dolt_diff_summary('HEAD', 'WORKING');`, []sql.Row{{3}}),
				allowed(`SELECT count(*) FROM dolt_log();`, []sql.Row{{3}}),
			},
		},
		{
			Name:        "unqualified table names cover every schema with a table of that name",
			SetUpScript: setUp,
			Assertions: []ScriptTestAssertion{
				// s2 is not on the search path, but the table functions still find s2.hidden
				denied(`SELECT to_ssn FROM dolt_diff('HEAD', 'WORKING', 'hidden');`, "hidden"),
				denied(`SELECT statement FROM dolt_patch('HEAD', 'WORKING', 'hidden');`, "hidden"),
				denied(`SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 'hidden');`, "hidden"),
				denied(`SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'hidden');`, "hidden"),
				denied(`SELECT * FROM dolt_preview_merge_conflicts('main', 'other', 'hidden');`, "hidden"),
				{
					Query:    `SET search_path = s2, public;`,
					Username: `limited`,
					Password: `password`,
					Expected: []sql.Row{},
				},
				denied(`SELECT to_ssn FROM dolt_diff('HEAD', 'WORKING', 'hidden');`, "hidden"),
			},
		},
		{
			Name: "table functions on a table that only exists in a commit",
			SetUpScript: []string{
				`CREATE USER limited PASSWORD 'password';`,
				`CREATE TABLE gone (pk INT PRIMARY KEY, ssn TEXT);`,
				`INSERT INTO gone VALUES (1, '111');`,
				`SELECT dolt_commit('-Am', 'init');`,
				`DROP TABLE gone;`,
			},
			Assertions: []ScriptTestAssertion{
				denied(`SELECT from_ssn FROM dolt_diff('HEAD', 'WORKING', 'gone');`, "gone"),
				denied(`SELECT statement FROM dolt_patch('HEAD', 'WORKING', 'gone');`, "gone"),
				{
					Query:    `SELECT from_ssn FROM dolt_diff('HEAD', 'WORKING', 'gone');`,
					Expected: []sql.Row{{"111"}},
				},
			},
		},
		{
			Name:        "superusers and commit-only table functions are unaffected",
			SetUpScript: setUp,
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT to_ssn FROM dolt_diff('HEAD', 'WORKING', 'secret');`,
					Expected: []sql.Row{{"222"}},
				},
				{
					Query:    `SELECT count(*) FROM dolt_patch('HEAD', 'WORKING');`,
					Expected: []sql.Row{{3}},
				},
				{
					Query:    `SELECT count(*) FROM dolt_log();`,
					Expected: []sql.Row{{3}},
				},
				allowed(`SELECT count(*) FROM dolt_branch_status('main', 'other');`, []sql.Row{{1}}),
				allowed(`SELECT count(*) >= 0 FROM dolt_reflog();`, []sql.Row{{"t"}}),
			},
		},
	})
}
