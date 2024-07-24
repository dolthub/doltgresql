package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestInfoSchemaColumns(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "information_schema.columns",
			SetUpScript: []string{
				"create table test_table (id int primary key, col1 varchar(255));",
				"create view test_view as select * from test_table;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.columns ORDER BY table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name, column_name FROM information_schema.columns WHERE table_schema='public' ORDER BY table_name;`,
					Expected: []sql.Row{
						{"postgres", "public", "test_table", "id"},
						{"postgres", "public", "test_table", "col1"},
						{"postgres", "public", "test_view", ""},
					},
				},
				{
					Query:    `CREATE SCHEMA test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SET SEARCH_PATH TO test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_table2 (id2 INT);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.columns order by table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"}, {"test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name, column_name FROM information_schema.columns WHERE table_schema='test_schema';`,
					Expected: []sql.Row{
						{"postgres", "test_schema", "test_table2", "id2"},
					},
				},
				{
					Query: "SELECT * FROM information_schema.columns WHERE table_name='test_table';",
					Expected: []sql.Row{
						{"postgres", "public", "test_table", "id", 1, nil, "NO", "integer", nil, nil, 32, 2, 0, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "postgres", "pg_catalog", "int4", nil, nil, nil, nil, nil, "NO", "NO", nil, nil, nil, nil, nil, "NO", "NEVER", nil, "YES"},
						{"postgres", "public", "test_table", "col1", 2, nil, "YES", "character varying", 255, 1020, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "postgres", "pg_catalog", "varchar", nil, nil, nil, nil, nil, "NO", "NO", nil, nil, nil, nil, nil, "NO", "NEVER", nil, "YES"},
					},
				},
				{
					Skip:  true, // TODO: Don't have complete view information to fill out these rows
					Query: "SELECT * FROM information_schema.columns WHERE table_name='test_view';",
					Expected: []sql.Row{
						{"postgres", "public", "test_view", "id", 1, nil, "YES", "integer", nil, nil, 32, 2, 0, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "postgres", "pg_catalog", "int4", nil, nil, nil, nil, 1, "NO", "NO", nil, nil, nil, nil, nil, "NO", "NEVER", nil, "YES"},
						{"postgres", "public", "test_view", "col1", 2, nil, "YES", "character varying", 255, 1020, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "postgres", "pg_catalog", "varchar", nil, nil, nil, nil, 2, "NO", "NO", nil, nil, nil, nil, nil, "NO", "NEVER", nil, "YES"},
					},
				},
				{
					Query: `SELECT columns.table_name, columns.column_name from "information_schema"."columns" WHERE table_name='test_table';`,
					Expected: []sql.Row{
						{"test_table", "id"},
						{"test_table", "col1"},
					},
				},
			},
		},
	})
}

func TestInfoSchemaSchemata(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:     "information_schema.schemata",
			Database: "newdb",
			SetUpScript: []string{
				"create schema test_schema",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT catalog_name, schema_name FROM information_schema.schemata order by schema_name;`,
					Expected: []sql.Row{
						{"newdb", "information_schema"},
						{"newdb", "pg_catalog"},
						{"newdb", "public"},
						{"newdb", "test_schema"},
					},
				},
			},
		},
	})
}

func TestInfoSchemaTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "information_schema.tables",
			SetUpScript: []string{
				"create table test_table (id int)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.tables order by table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema FROM information_schema.tables group by table_catalog, table_schema order by table_schema;`,
					Expected: []sql.Row{
						{"postgres", "information_schema"},
						{"postgres", "pg_catalog"},
						{"postgres", "public"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name FROM information_schema.tables WHERE table_schema='public';`,
					Expected: []sql.Row{
						{"postgres", "public", "test_table"},
					},
				},
				{
					Query:    `CREATE SCHEMA test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SET SEARCH_PATH TO test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_table2 (id INT);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.tables order by table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"}, {"test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema FROM information_schema.tables group by table_catalog, table_schema order by table_schema;`,
					Expected: []sql.Row{
						{"postgres", "information_schema"},
						{"postgres", "pg_catalog"},
						{"postgres", "public"},
						{"postgres", "test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name FROM information_schema.tables WHERE table_schema='test_schema';`,
					Expected: []sql.Row{
						{"postgres", "test_schema", "test_table2"},
					},
				},
				{
					Skip:     true, // TODO: need ENUM type for table_type column
					Query:    "SELECT * FROM information_schema.tables ORDER BY table_name;",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT "table_schema", "table_name", obj_description(('"' || "table_schema" || '"."' || "table_name" || '"')::regclass, 'pg_class') AS table_comment FROM "information_schema"."tables" WHERE ("table_schema" = 'test_schema' AND "table_name" = 'test_table2')`,
					Expected: []sql.Row{{"test_schema", "test_table2", ""}},
				},
			},
		},
	})
}
