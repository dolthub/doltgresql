package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestInfoSchemaTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "information_schema.tables",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT table_catalog, table_schema, table_name FROM information_schema.tables;`,
					Expected: []sql.Row{
						{},
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
