// Copyright 2023 Dolthub, Inc.
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreparedStatements(t *testing.T) {
	RunScripts(t, preparedStatementTests)
}

func TestPreparedPgCatalog(t *testing.T) {
	RunScripts(t, pgCatalogTests)
}

var preparedStatementTests = []ScriptTest{
	{
		Name:        "Expressions without tables",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT CONCAT($1::text, $2::text)",
				BindVars: []any{"hello", "world"},
				Expected: []sql.Row{
					{"helloworld"},
				},
			},
			{
				Query:    "SELECT $1::integer + $2::integer",
				BindVars: []any{1, 2},
				Expected: []sql.Row{
					{3},
				},
			},
			{
				Query:    "select $1 as test",
				BindVars: []any{"hello"},
				Expected: []sql.Row{
					{"hello"},
				},
			},
		},
	},
	{
		Name: "Expressions with tables",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1);",
				BindVars: []any{"public"},
				Expected: []sql.Row{{"t"}},
			},
			{
				Query:    "SELECT nspname FROM pg_namespace LIMIT $1;",
				BindVars: []any{1},
				Expected: []sql.Row{{"dolt"}},
			},
			{
				Skip:     true, // TODO: ERROR: unsupported syntax: <nil>
				Query:    "SELECT nspname FROM pg_namespace OFFSET $1;",
				BindVars: []any{1},
				Expected: []sql.Row{{"dolt"}},
			},
		},
	},
	{
		Name: "Integer insert",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 2, 3, 4},
			},
			{
				Query: "SELECT * FROM test order by pk;",
				Expected: []sql.Row{
					{1, 2},
					{3, 4},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{2},
				Expected: []sql.Row{
					{1, 2},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{3},
				Expected: []sql.Row{},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 + $1 = $2;",
				BindVars: []any{1, 3},
				Expected: []sql.Row{
					{1, 2},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE pk + v1 = $1;",
				BindVars: []any{3},
				Expected: []sql.Row{
					{1, 2},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1::integer + $2::integer;",
				BindVars: []any{1, 3},
				Expected: []sql.Row{
					{3, 4},
				},
			},
		},
	},
	{
		Name: "Integer update",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 2, 3, 4},
			},
			{
				Query:    "UPDATE test set v1 = $1 WHERE pk = $2;",
				BindVars: []any{5, 1},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{5},
				Expected: []sql.Row{
					{1, 5},
				},
			},
		},
	},
	{
		Name: "Integer delete",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 2, 3, 4},
			},
			{
				Query:    "DELETE FROM test WHERE pk = $1;",
				BindVars: []any{1},
			},
			{
				Query: "SELECT * FROM test order by 1;",
				Expected: []sql.Row{
					{3, 4},
				},
			},
		},
	},
	{
		Name: "String insert",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, "hello", 3, "goodbye"},
			},
			{
				Query: "SELECT * FROM test order by pk;",
				Expected: []sql.Row{
					{1, "hello"},
					{3, "goodbye"},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE s = $1;",
				BindVars: []any{"hello"},
				Expected: []sql.Row{
					{1, "hello"},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE s = concat($1::text, $2::text);",
				BindVars: []any{"he", "llo"},
				Expected: []sql.Row{
					{1, "hello"},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE concat(s, '!') = $1",
				BindVars: []any{"hello!"},
				Expected: []sql.Row{
					{1, "hello"},
				},
			},
		},
	},
	{
		Name: "String update",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, "hello", 3, "goodbye"},
			},
			{
				Query:    "UPDATE test set s = $1 WHERE pk = $2;",
				BindVars: []any{"new value", 1},
			},
			{
				Query:    "SELECT * FROM test WHERE s = $1;",
				BindVars: []any{"new value"},
				Expected: []sql.Row{
					{1, "new value"},
				},
			},
		},
	},
	{
		Name: "String delete",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, s character varying(20));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, "hello", 3, "goodbye"},
			},
			{
				Query:    "DELETE FROM test WHERE s = $1;",
				BindVars: []any{"hello"},
			},
			{
				Query: "SELECT * FROM test ORDER BY 1;",
				Expected: []sql.Row{
					{3, "goodbye"},
				},
			},
		},
	},
	{
		Name: "Float insert",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 1.1, 3, 3.3},
			},
			{
				Query: "SELECT * FROM test ORDER BY 1;",
				Expected: []sql.Row{
					{1, 1.1},
					{3, 3.3},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE f1 = $1;",
				BindVars: []any{1.1},
				Expected: []sql.Row{
					{1, 1.1},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE f1 + $1 = $2;",
				BindVars: []any{1.0, 2.1},
				Expected: []sql.Row{
					{1, 1.1},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE f1 = $1::decimal + $2::decimal;",
				BindVars: []any{1.0, 0.1},
				Expected: []sql.Row{
					{1, 1.1},
				},
			},
		},
	},
	{
		Name: "Float update",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 1.1, 3, 3.3},
			},
			{
				Query:    "UPDATE test set f1 = $1 WHERE f1 = $2;",
				BindVars: []any{2.2, 1.1},
			},
			{
				Query:    "SELECT * FROM test WHERE f1 = $1;",
				BindVars: []any{2.2},
				Expected: []sql.Row{
					{1, 2.2},
				},
			},
		},
	},
	{
		Name: "Float delete",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, f1 DOUBLE PRECISION);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, 1.1, 3, 3.3},
			},
			{
				Query:    "DELETE FROM test WHERE f1 = $1;",
				BindVars: []any{1.1},
			},
			{
				Query: "SELECT * FROM test order by 1;",
				Expected: []sql.Row{
					{3, 3.3},
				},
			},
		},
	},
	{
		Name: "Date type insert, update, delete",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 DATE);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "INSERT INTO test VALUES ($1, $2), ($3, $4);",
				BindVars: []any{1, "2022-02-02", 3, "2024-04-01 -07"},
			},
			{
				Query: "SELECT * FROM test order by pk;",
				Expected: []sql.Row{
					{1, "2022-02-02"},
					{3, "2024-04-01"},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{"2022-02-02"},
				Expected: []sql.Row{
					{1, "2022-02-02"},
				},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{"2022-02-03"},
				Expected: []sql.Row{},
			},
			{
				Query:    "UPDATE test set v1 = $1 WHERE pk = $2;",
				BindVars: []any{"2022-02-03", 1},
			},
			{
				Query:    "SELECT * FROM test WHERE v1 = $1;",
				BindVars: []any{"2022-02-03"},
				Expected: []sql.Row{
					{1, "2022-02-03"},
				},
			},
			{
				Query:    "DELETE FROM test WHERE pk = $1;",
				BindVars: []any{1},
			},
			{
				Query: "SELECT * FROM test order by 1;",
				Expected: []sql.Row{
					{3, "2024-04-01"},
				},
			},
		},
	},
	{
		Name: "pg_get_viewdef function",
		SetUpScript: []string{
			"CREATE TABLE test (id int, name text)",
			"INSERT INTO test VALUES (1,'desk'), (2,'chair')",
			"CREATE VIEW test_view AS SELECT name FROM test",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    `select pg_get_viewdef($1::regclass);`,
				BindVars: []any{"test_view"},
				Expected: []sql.Row{{"SELECT name FROM test"}},
			},
		},
	},
}

var pgCatalogTests = []ScriptTest{
	{
		Name: "pg_namespace",
		SetUpScript: []string{
			`CREATE SCHEMA testschema;`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    `SELECT * FROM "pg_catalog"."pg_namespace" WHERE nspname=$1;`,
				BindVars: []any{"testschema"},
				Expected: []sql.Row{{2638679668, "testschema", 0, nil}},
			},
			{
				Query:    `SELECT * FROM "pg_catalog"."pg_namespace" WHERE oid=$1;`,
				BindVars: []any{2638679668},
				Expected: []sql.Row{{2638679668, "testschema", 0, nil}},
			},
		},
	},
	{
		Name: "pg_tables",
		SetUpScript: []string{
			`CREATE TABLE testing (pk INT primary key, v1 INT);`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    `SELECT * FROM "pg_catalog"."pg_tables" WHERE tablename=$1;`,
				BindVars: []any{"testing"},
				Expected: []sql.Row{{"public", "testing", "", "", "t", "f", "f", "f"}},
			},
			{
				Query:    `SELECT count(*) FROM "pg_catalog"."pg_tables" WHERE schemaname=$1;`,
				BindVars: []any{"pg_catalog"},
				Expected: []sql.Row{{139}},
			},
		},
	},
	{
		Name: "pg_class",
		SetUpScript: []string{
			`CREATE SCHEMA testschema;`,
			`CREATE TABLE testschema.testtable (id int primary key, v1 text)`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT c.oid,d.description,pg_catalog.pg_get_expr(c.relpartbound, c.oid) as partition_expr,  pg_catalog.pg_get_partkeydef(c.oid) as partition_key 
FROM pg_catalog.pg_class c
LEFT OUTER JOIN pg_catalog.pg_description d ON d.objoid=c.oid AND d.objsubid=0 AND d.classoid='pg_class'::regclass
WHERE c.relnamespace=$1 AND c.relkind not in ('i','I','c');`,
				BindVars: []any{2638679668},
				Expected: []sql.Row{{1712283605, nil, nil, ""}},
			},
			{
				Query:    `select c.oid,pg_catalog.pg_total_relation_size(c.oid) as total_rel_size,pg_catalog.pg_relation_size(c.oid) as rel_size FROM pg_class c WHERE c.relnamespace=$1;`,
				BindVars: []any{2638679668},
				Expected: []sql.Row{{444447634, 0, 0}, {1712283605, 0, 0}},
			},
			{
				Query: `SELECT c.relname, a.attrelid, a.attname, a.atttypid, pg_catalog.pg_get_expr(ad.adbin, ad.adrelid, true) as def_value,dsc.description,dep.objid 
FROM pg_catalog.pg_attribute a 
INNER JOIN pg_catalog.pg_class c ON (a.attrelid=c.oid) 
LEFT OUTER JOIN pg_catalog.pg_attrdef ad ON (a.attrelid=ad.adrelid AND a.attnum = ad.adnum) 
LEFT OUTER JOIN pg_catalog.pg_description dsc ON (c.oid=dsc.objoid AND a.attnum = dsc.objsubid) 
LEFT OUTER JOIN pg_depend dep on dep.refobjid = a.attrelid AND dep.deptype = 'i' and dep.refobjsubid = a.attnum and dep.classid = dep.refclassid 
WHERE NOT a.attisdropped AND c.relkind not in ('i','I','c') AND c.oid=$1 ORDER BY a.attnum`,
				BindVars: []any{1712283605},
				Expected: []sql.Row{{"testtable", 1712283605, "id", 23, nil, nil, nil}, {"testtable", 1712283605, "v1", 25, nil, nil, nil}},
			},
		},
	},
}

func TestPreparedErrorHandling(t *testing.T) {
	tt := ScriptTest{
		Name: "error handling doesn't foul session",
		SetUpScript: []string{
			"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			"insert into test values (1, 1), (2, 2), (3, 3), (4, 4), (5, 5), (6, 6), (7, 7);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select v1 from doesNotExist where pk = 1;",
				ExpectedErr: "table not found",
			},
			{
				Query:    "select v1 from test where pk = 1;",
				Expected: []sql.Row{{1}},
			},
			{
				Query:    "select v1 from test where pk = 2;",
				Expected: []sql.Row{{2}},
			},
			{
				Query:    "select v1 from test where pk = 3;",
				Expected: []sql.Row{{3}},
			},
			{
				Query:    "select v1 from test where pk = 4;",
				Expected: []sql.Row{{4}},
			},
			{
				Query:    "select v1 from test where pk = 5;",
				Expected: []sql.Row{{5}},
			},
			{
				Query:    "select v1 from test where pk = 6;",
				Expected: []sql.Row{{6}},
			},
			{
				Query:    "select v1 from test where pk = 7;",
				Expected: []sql.Row{{7}},
			},
		},
	}

	RunScriptN(t, tt, 20)
}

// RunScriptN runs the assertions of the given script n times using the same connection
func RunScriptN(t *testing.T, script ScriptTest, n int) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}
	ctx, conn, controller := CreateServer(t, scriptDatabase)
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		err := controller.WaitForStop()
		require.NoError(t, err)
	}()

	// Run the setup
	for _, query := range script.SetUpScript {
		rows, err := conn.Query(ctx, query)
		require.NoError(t, err)
		_, err = ReadRows(rows, true)
		assert.NoError(t, err)
	}

	for i := 0; i < n; i++ {
		t.Run(script.Name, func(t *testing.T) {
			// Run the assertions
			for _, assertion := range script.Assertions {
				t.Run(assertion.Query, func(t *testing.T) {
					if assertion.Skip {
						t.Skip("Skip has been set in the assertion")
					}
					rows, err := conn.Query(ctx, assertion.Query)
					if err == nil {
						defer rows.Close()
					}

					var errorSeen string

					if assertion.ExpectedErr == "" {
						require.NoError(t, err)
					} else if err != nil {
						errorSeen = err.Error()
					}

					if errorSeen == "" {
						foundRows, err := ReadRows(rows, true)
						if assertion.ExpectedErr == "" {
							require.NoError(t, err)
							assert.Equal(t, NormalizeExpectedRow(rows.FieldDescriptions(), assertion.Expected), foundRows)
						} else if err != nil {
							errorSeen = err.Error()
						}
					}

					if assertion.ExpectedErr != "" {
						require.False(t, errorSeen == "", "Expected error but got none")
						assert.Contains(t, errorSeen, assertion.ExpectedErr)
					}
				})
			}
		})
	}
}
