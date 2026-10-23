package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestBranchRoutineCatalog(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "routine additions overloads and drops stay on their branch",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY)",
			"CREATE FUNCTION f(input text) RETURNS text LANGUAGE plpgsql AS $$BEGIN RETURN input||'_main'; END$$",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_checkout('-b','other')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "CREATE OR REPLACE FUNCTION f(input text) RETURNS text LANGUAGE plpgsql AS $$BEGIN RETURN input||'_other'; END$$"},
			{Query: "CREATE FUNCTION f(input int) RETURNS int LANGUAGE plpgsql AS $$BEGIN RETURN input+1; END$$"},
			{Query: "SELECT f('x'::text),f(4)", Expected: []sql.Row{{"x_other", 5}}},
			{Query: "SELECT dolt_commit('-Am','other functions')", SkipResultsCheck: true},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "SELECT f('x'::text)", Expected: []sql.Row{{"x_main"}}},
			{Query: "SELECT f(4)", ExpectedErr: "does not exist"},
			{Query: "SELECT dolt_checkout('other')", SkipResultsCheck: true},
			{Query: "DROP FUNCTION f(text)"},
			{Query: "SELECT f(4)", Expected: []sql.Row{{5}}},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "SELECT f('x'::text)", Expected: []sql.Row{{"x_main"}}},
		},
	}})
}
