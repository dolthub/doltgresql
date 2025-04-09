// Copyright 2025 Dolthub, Inc.
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

func TestPsqlCommands(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			// Many of the psql commands use the OPERATOR(pg_catalog.+) syntax, testing it here directly in a simpler context
			Name: "operator keyword",
			Assertions: []ScriptTestAssertion{
				{
					Query: "select 1 OPERATOR(pg_catalog.+) 1",
					Expected: []sql.Row{
						{2},
					},
				},
				{
					Query: "select 1 OPERATOR(PG_CATALOG.+) 1",
					Expected: []sql.Row{
						{2},
					},
				},
				{
					Query: "select 1 OPERATOR(myschema.+) 1",
					ExpectedErr: "schema \"myschema\" not allowed",
				},
				{
					Query: "select 1 OPERATOR(pg_catalog.<) 1",
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query: "select 1 OPERATOR(pg_catalog.<=) 1",
					Expected: []sql.Row{
						{"t"},
					},
				},
				{
					Query: "select 1 OPERATOR(pg_catalog.=) 1",
					Expected: []sql.Row{
						{"t"},
					},
				},
				{
					Query: "select 'hello' OPERATOR(pg_catalog.~) 'hello';",
					Expected: []sql.Row{
						{"t"},
					},
				},
			},
		},
		{
			Name: `\dt tablename`,
			Focus: true,
			SetUpScript: []string{
				"CREATE TABLE test_table (id INT PRIMARY KEY, name TEXT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT n.nspname as \"Schema\", " +
						"  c.relname as \"Name\", " +
						"  CASE c.relkind WHEN 'r' THEN 'table' WHEN 'v' THEN 'view' WHEN 'm' THEN 'materialized view' WHEN 'i' THEN 'index' WHEN 'S' THEN 'sequence' WHEN 't' THEN 'TOAST table' WHEN 'f' THEN 'foreign table' WHEN 'p' THEN 'partitioned table' WHEN 'I' THEN 'partitioned index' END as \"Type\", " +
						"  pg_catalog.pg_get_userbyid(c.relowner) as \"Owner\" " +
						"FROM pg_catalog.pg_class c " +
						"     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace " +
						"     LEFT JOIN pg_catalog.pg_am am ON am.oid = c.relam " +
						"WHERE c.relkind IN ('r','p','t','s','') " +
						"  AND c.relname OPERATOR(pg_catalog.~) '^(test_table)$' COLLATE pg_catalog.default " +
						"  AND pg_catalog.pg_table_is_visible(c.oid) " +
						"ORDER BY 1,2;",
				},
			},
		},
	})
}
