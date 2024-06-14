package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestPgDatabase(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_database",
			SetUpScript: []string{
				`CREATE DATABASE test;`,
				`CREATE TABLE test (pk INT primary key, v1 INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_database";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT * FROM PG_catalog.pg_DATABASE ORDER BY datname;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM pg_catalog.pg_database WHERE datname='test';",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
