package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// Both engines run this branch-local CREATE scenario without catalog preparation.
func TestBranchCatalogCreate(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "create publishes only selected branch",
		SetUpScript: []string{
			"CREATE TABLE seed(id int)",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_branch('other')",
			"SELECT dolt_checkout('other')",
			"SELECT dolt_checkout('main')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "BEGIN"},
			{Query: "CREATE TABLE aborted(id int)"},
			{Query: "ROLLBACK"},
			{Query: "CREATE TABLE added(id int PRIMARY KEY, v text UNIQUE)"},
			{Query: "INSERT INTO added VALUES(1,'main')"},
			{Query: "SELECT * FROM added", Expected: []sql.Row{{1, "main"}}},
			{Query: "CREATE TABLE IF NOT EXISTS added(id int)"},
			{Query: "SELECT * FROM added", Expected: []sql.Row{{1, "main"}}},
			{Query: "SELECT dolt_commit('-Am','main creation')", SkipResultsCheck: true},
			{Query: "SELECT * FROM added AS OF 'HEAD'", Expected: []sql.Row{{1, "main"}}},
			{Query: "SELECT dolt_checkout('other')", SkipResultsCheck: true},
			{Query: "CREATE TABLE separate(id int PRIMARY KEY, v text UNIQUE)"},
			{Query: "INSERT INTO separate VALUES(2,'other')"},
			{Query: "SELECT dolt_commit('-Am','other creation')", SkipResultsCheck: true},
			{Query: `SELECT a.v,b.v FROM "postgres/main".public.added a CROSS JOIN "postgres/other".public.separate b`, Expected: []sql.Row{{"main", "other"}}},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "SELECT * FROM added", Expected: []sql.Row{{1, "main"}}},
		},
	}})
}
