package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestPgCatalog(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_catalog",
			SetUpScript: []string{
				`CREATE DATABASE test;`,
				`CREATE TABLE test (pk INT primary key, v1 INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT datname FROM pg_catalog.pg_database;",
					Expected: []sql.Row{
						{"postgres"},
						{"test"},
					},
				},
			},
		},
	})
}
