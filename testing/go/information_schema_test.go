package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestInfoSchemaTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "information_schema.tables",
			SetUpScript: []string{
				"create table test_table (id int)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT table_schema FROM information_schema.tables group by table_schema order by table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"},
					},
				},
				{
					// TODO: all table_catalog values should be "doltgres"
					Query: `SELECT table_catalog, table_schema FROM information_schema.tables group by table_catalog, table_schema order by table_catalog;`,
					Expected: []sql.Row{
						{"def", "information_schema"},
						{"doltgres", "pg_catalog"},
						{"postgres", "pg_catalog"},
						{"postgres", "public"},
					},
				},
				{
					Skip:     true, // TODO: need ENUM type for table_type column
					Query:    "SELECT * FROM PG_catalog.pg_AGGREGATE ORDER BY aggfnoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
