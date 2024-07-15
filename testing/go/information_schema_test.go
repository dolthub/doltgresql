package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

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
				"create schema test_schema;",
				"SET search_path TO test_schema;",
				"create table test_table (id int);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.tables order by table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema FROM information_schema.tables group by table_catalog, table_schema order by table_schema;`,
					Expected: []sql.Row{
						{"postgres", "information_schema"},
						{"postgres", "pg_catalog"},
						{"postgres", "test_schema"},
					},
				},
				{
					Skip:     true, // TODO: need ENUM type for table_type column
					Query:    "SELECT * FROM PG_catalog.pg_AGGREGATE ORDER BY aggfnoid;",
					Expected: []sql.Row{},
				},
				{
					Skip:     true, // TODO: regclass failing here
					Query:    `SELECT "table_schema", "table_name", ('"' || "table_schema" || '"."' || "table_name" || '"')::regclass AS table_oid FROM "information_schema"."tables" WHERE ("table_schema" = 'test_schema' AND "table_name" = 'test_table');`,
					Expected: []sql.Row{{"test_schema", "test_table", ""}},
				},
			},
		},
	})
}
