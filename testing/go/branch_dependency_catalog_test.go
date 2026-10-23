package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestBranchDependencyCatalog(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "routine dependency cascade and history",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY)",
			"INSERT INTO t VALUES(1)",
			"CREATE FUNCTION f(input t) RETURNS int LANGUAGE plpgsql AS $$BEGIN RETURN input.id; END$$",
			"SELECT dolt_commit('-Am','dependency base')",
			"SELECT dolt_checkout('-b','other')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "DROP TABLE t", ExpectedErr: "cannot drop table"},
			{Query: "SELECT f(t) FROM t", Expected: []sql.Row{{1}}},
			{Query: "DROP TABLE t CASCADE"},
			{Query: "SELECT count(*) FROM pg_proc WHERE proname='f'", Expected: []sql.Row{{0}}},
			{Query: "SELECT id FROM t AS OF 'main'", Expected: []sql.Row{{1}}},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "SELECT f(t) FROM t", Expected: []sql.Row{{1}}},
		},
	}, {
		Name: "index and constraint deletion are branch local",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY,x int CONSTRAINT unique_x UNIQUE,y int)",
			"CREATE UNIQUE INDEX unique_y ON t(y)",
			"INSERT INTO t VALUES(1,2,3)",
			"SELECT dolt_commit('-Am','indexes')",
			"SELECT dolt_checkout('-b','other')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "DROP INDEX unique_y"},
			{Query: "ALTER TABLE t DROP CONSTRAINT unique_x"},
			{Query: "INSERT INTO t VALUES(2,2,3)"},
			{Query: "SELECT count(*) FROM t", Expected: []sql.Row{{2}}},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "SELECT count(*) FROM t", Expected: []sql.Row{{1}}},
			{Query: "INSERT INTO t VALUES(2,2,4)", ExpectedErr: "duplicate key"},
			{Query: "INSERT INTO t VALUES(2,4,3)", ExpectedErr: "duplicate key"},
		},
	}})
}
