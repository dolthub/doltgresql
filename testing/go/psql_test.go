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
					Query:       "select 1 OPERATOR(myschema.+) 1",
					ExpectedErr: "schema \"myschema\" not allowed",
				},
				{
					Query: "select 1 OPERATOR(pg_catalog.<) 1",
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query:       "select 1 OPERATOR(myschema.<) 1",
					ExpectedErr: "schema \"myschema\" not allowed",
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
					Expected: []sql.Row{{"public", "test_table", "table", "postgres"}},
				},
			},
		},
		{
			Name: `\d tablename`,
			SetUpScript: []string{
				"CREATE TABLE test_table (id INT PRIMARY KEY, name TEXT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// these queries return no rows because of the hard-coded oids, should fix
					Query: `SELECT pol.polname, pol.polpermissive,
       CASE WHEN pol.polroles = '{0}' THEN NULL ELSE pg_catalog.array_to_string(array(select rolname from pg_catalog.pg_roles where oid = any (pol.polroles) order by 1),',') END,
       pg_catalog.pg_get_expr(pol.polqual, pol.polrelid),
       pg_catalog.pg_get_expr(pol.polwithcheck, pol.polrelid),
       CASE pol.polcmd
           WHEN 'r' THEN 'SELECT'
           WHEN 'a' THEN 'INSERT'
           WHEN 'w' THEN 'UPDATE'
           WHEN 'd' THEN 'DELETE'
           END AS cmd
FROM pg_catalog.pg_policy pol
WHERE pol.polrelid = '4131846889' ORDER BY 1;`,
				},
				{
					Query: `SELECT oid, stxrelid::pg_catalog.regclass,
stxnamespace::pg_catalog.regnamespace::pg_catalog.text AS nsp,
 stxname, pg_catalog.pg_get_statisticsobjdef_columns(oid) AS columns,
          'd' = any(stxkind) AS ndist_enabled,
          'f' = any(stxkind) AS deps_enabled,
          'm' = any(stxkind) AS mcv_enabled,
          stxstattarget FROM pg_catalog.pg_statistic_ext 
                        WHERE stxrelid = '4131846889' ORDER BY nsp, stxname;`,
				},
				{
					Skip: true, // lots that we don't support yet
					Query: `SELECT pubname
, NULL
, NULL
FROM pg_catalog.pg_publication p
JOIN pg_catalog.pg_publication_namespace pn ON p.oid = pn.pnpubid
JOIN pg_catalog.pg_class pc ON pc.relnamespace = pn.pnnspid
WHERE pc.oid ='4131846889' and pg_catalog.pg_relation_is_publishable('4131846889')
UNION
SELECT pubname
, pg_get_expr(pr.prqual, c.oid)
, (CASE WHEN pr.prattrs IS NOT NULL THEN
(SELECT string_agg(attname, ', ')
FROM pg_catalog.generate_series(0, pg_catalog.array_upper(pr.prattrs::pg_catalog.int2[], 1)) s,
pg_catalog.pg_attribute
WHERE attrelid = pr.prrelid AND attnum = prattrs[s])
ELSE NULL END) FROM pg_catalog.pg_publication p
JOIN pg_catalog.pg_publication_rel pr ON p.oid = pr.prpubid
JOIN pg_catalog.pg_class c ON c.oid = pr.prrelid
WHERE pr.prrelid = '4131846889'
UNION
SELECT pubname
, NULL
, NULL
FROM pg_catalog.pg_publication p
WHERE p.puballtables AND pg_catalog.pg_relation_is_publishable('4131846889')
ORDER BY 1;`,
				},
			},
		},
	})
}
