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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestUserSpaceDoltTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "dolt branches",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT name FROM dolt.branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:    `SELECT name FROM dolt_branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Skip:     true, // TODO: referencing items outside the schema or database is not yet supported
					Query:    `SELECT dolt.branches.name FROM dolt.branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Skip:     true, // TODO: ERROR: table not found: dolt_branches
					Query:    `SELECT dolt_branches.name FROM dolt_branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:       `SELECT * FROM public.branches`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM branches`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE branches (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO branches VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT name FROM dolt.branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:       `CREATE SCHEMA dolt`,
					ExpectedErr: "schema exists",
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT name FROM branches`,
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:    `SELECT * FROM public.branches`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM branches`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt log",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT count(*) FROM dolt.log`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT count(*) FROM dolt_log`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:       `SELECT * FROM public.log`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM log`,
					ExpectedErr: "table not found",
				},
				{
					Query:    `CREATE TABLE log (id INT PRIMARY KEY)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO log VALUES (1)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM log`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT count(*) FROM dolt.log`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SET search_path = 'dolt'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT count(*) FROM log`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:    `SELECT * FROM public.log`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SET search_path = 'public'",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM log`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "dolt tags",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT tag_name FROM dolt.tags`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_tags`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM public.tags`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM tags`,
					ExpectedErr: "table not found",
				},
			},
		},
		{
			Name: "dolt docs",
			SetUpScript: []string{
				"INSERT INTO dolt.docs values ('README.md', 'testing')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM dolt.docs`,
					Expected: []sql.Row{
						{"README.md", "testing"},
					},
				},
				{
					Query: `SELECT * FROM dolt_docs`,
					Expected: []sql.Row{
						{"README.md", "testing"},
					},
				},
				{
					Query:       `SELECT * FROM public.docs`,
					ExpectedErr: "table not found",
				},
				{
					Query:       `SELECT * FROM docs`,
					ExpectedErr: "table not found",
				},
			},
		},
		{
			Name: "dolt schemas",
			SetUpScript: []string{
				"create view myView as select 2 + 2",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM dolt_schemas`,
					Expected: []sql.Row{
						{
							"view",
							"myview",
							"create view myView as select 2 + 2",
							"{\"CreatedAt\":0}",
							"NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES",
						},
					},
				},
			},
		},
	})
}
