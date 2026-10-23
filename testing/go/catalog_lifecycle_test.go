package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

func TestCatalogLifecycle(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "retained composite routine signatures survive schema changes",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY,x int)",
			"INSERT INTO t VALUES(1,10)",
			"CREATE FUNCTION f(t) RETURNS int LANGUAGE sql AS $$SELECT ($1).x$$",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_checkout('-b','other')",
			"PREPARE retained AS SELECT f(t) FROM t AS OF 'main'",
			"ALTER TABLE t ADD COLUMN extra text",
			"INSERT INTO t VALUES(2,20,'other')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "EXECUTE retained", Expected: []sql.Row{{10}}},
			{Query: "SELECT f(t) FROM t AS OF 'main'", Expected: []sql.Row{{10}}},
			{Query: "SELECT f(t) FROM t ORDER BY id", Expected: []sql.Row{{10}, {20}}},
			{Query: "SELECT dolt_commit('-Am','other')", SkipResultsCheck: true},
			{Query: "EXECUTE retained", Expected: []sql.Row{{10}}},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "SELECT f(t) FROM t", Expected: []sql.Row{{10}}},
		},
	}})
}
