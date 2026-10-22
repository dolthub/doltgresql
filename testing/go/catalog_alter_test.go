package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// Canonical definitions must change with the selected branch without preparation.
func TestBranchCatalogAlter(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name:        "plain ALTER needs no catalog preparation",
		SetUpScript: []string{"CREATE TABLE plain(a int,b text)", "INSERT INTO plain VALUES(1,'old')"},
		Assertions: []ScriptTestAssertion{
			{Query: "ALTER TABLE plain ADD COLUMN c int"},
			{Query: "SELECT * FROM plain", Expected: []sql.Row{{1, "old", nil}}},
			{Query: "ALTER TABLE plain DROP COLUMN b"},
			{Query: "SELECT * FROM plain", Expected: []sql.Row{{1, nil}}},
			{Query: "INSERT INTO plain VALUES(2,20)"},
			{Query: "SELECT * FROM plain ORDER BY a", Expected: []sql.Row{{1, nil}, {2, 20}}},
		},
	}, {
		Name: "add and drop columns preserve other branches and history",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY, old text, v int)",
			"INSERT INTO t VALUES(1,'one',10),(2,'two',20)",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_branch('other')",
			"SELECT dolt_checkout('other')",
			"SELECT dolt_checkout('main')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "ALTER TABLE t ADD COLUMN extra text"},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1, "one", 10, nil}, {2, "two", 20, nil}}},
			{Query: "UPDATE t SET extra='new' WHERE id=1"},
			{Query: "ALTER TABLE t DROP COLUMN old"},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1, 10, "new"}, {2, 20, nil}}},
			{Query: "SELECT * FROM t AS OF 'HEAD' ORDER BY id", Expected: []sql.Row{{1, "one", 10}, {2, "two", 20}}},
			{Query: `SELECT a.extra,b.old FROM "postgres/main".public.t a JOIN "postgres/other".public.t b USING(id) WHERE id=1`, Expected: []sql.Row{{"new", "one"}}},
			{Query: "SELECT dolt_commit('-Am','changed layout')", SkipResultsCheck: true},
			{Query: "SELECT * FROM t AS OF 'HEAD~1' ORDER BY id", Expected: []sql.Row{{1, "one", 10}, {2, "two", 20}}},
			{Query: "SELECT dolt_checkout('other')", SkipResultsCheck: true},
			{Query: "ALTER TABLE t DROP COLUMN v"},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1, "one"}, {2, "two"}}},
			{Query: "INSERT INTO t VALUES(3,'three')"},
			{Query: "SELECT dolt_commit('-Am','other layout')", SkipResultsCheck: true},
			{Query: `SELECT a.extra,b.old FROM "postgres/main".public.t a JOIN "postgres/other".public.t b USING(id) WHERE id=1`, Expected: []sql.Row{{"new", "one"}}},
		},
	}, {
		Name: "drop column preserves duplicate row multiplicity",
		SetUpScript: []string{
			"CREATE TABLE bag(a int,b text)", "INSERT INTO bag VALUES(1,'a'),(1,'b'),(1,'b')",
			"SELECT dolt_commit('-Am','base')", "SELECT dolt_checkout('-b','other')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "ALTER TABLE bag DROP COLUMN b"},
			{Query: "SELECT a FROM bag", Expected: []sql.Row{{1}, {1}, {1}}},
			{Query: "SELECT * FROM bag AS OF 'HEAD' ORDER BY b", Expected: []sql.Row{{1, "a"}, {1, "b"}, {1, "b"}}},
			{Query: "SELECT dolt_commit('-Am','dropped')", SkipResultsCheck: true},
		},
	}})
}
