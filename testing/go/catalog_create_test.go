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

// Same logical names must never require a catalog preparation call.
func TestBranchCatalogSameName(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "independent same-name schemas and retained history",
		SetUpScript: []string{
			"CREATE TABLE seed(id int)",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_branch('other')",
			"SELECT dolt_checkout('other')",
			"SELECT dolt_checkout('main')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "CREATE TABLE t(id int PRIMARY KEY,v int UNIQUE)"},
			{Query: "INSERT INTO t VALUES(1,10)"},
			{Query: "SELECT dolt_commit('-Am','main table')", SkipResultsCheck: true},
			{Query: "SELECT dolt_tag('before_recreate')", SkipResultsCheck: true},
			{Query: "SELECT dolt_checkout('other')", SkipResultsCheck: true},
			{Query: "CREATE TABLE t(id int PRIMARY KEY,v text UNIQUE,note text)"},
			{Query: "INSERT INTO t VALUES(1,'ten','other')"},
			{Query: "CREATE TABLE IF NOT EXISTS t(wrong int)"},
			{Query: "UPDATE t SET note='changed' WHERE id=1"},
			{Query: "INSERT INTO t VALUES(2,'two','remove')"},
			{Query: "DELETE FROM t WHERE id=2"},
			{Query: `SELECT a.v,b.v,b.note FROM "postgres/main".public.t a JOIN "postgres/other".public.t b USING(id)`, Expected: []sql.Row{{10, "ten", "changed"}}},
			{Query: "SELECT dolt_commit('-Am','other table')", SkipResultsCheck: true},
			{Query: "SELECT a.v,b.v FROM t AS OF 'main' a JOIN t AS OF 'other' b USING(id)", Expected: []sql.Row{{10, "ten"}}},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "BEGIN"},
			{Query: "DROP TABLE t"},
			{Query: "CREATE TABLE t(id int,v text)"},
			{Query: "INSERT INTO t VALUES(2,'aborted')"},
			{Query: "ROLLBACK"},
			{Query: "SELECT * FROM t", Expected: []sql.Row{{1, 10}}},
			{Query: "DROP TABLE t"},
			{Query: "CREATE TABLE t(id int PRIMARY KEY,v text)"},
			{Query: "INSERT INTO t VALUES(3,'recreated')"},
			{Query: "SELECT * FROM t AS OF 'HEAD'", Expected: []sql.Row{{1, 10}}},
			{Query: `SELECT * FROM "postgres/other".public.t`, Expected: []sql.Row{{1, "ten", "changed"}}},
			{Query: "SELECT dolt_commit('-Am','recreated')", SkipResultsCheck: true},
			{Query: "SELECT * FROM t", Expected: []sql.Row{{3, "recreated"}}},
			{Query: "SELECT dolt_checkout('-b','oldgeneration','before_recreate')", SkipResultsCheck: true},
			{Query: "SELECT * FROM t", Expected: []sql.Row{{1, 10}}},
			{Query: "INSERT INTO t VALUES(4,40)"},
			{Query: "SELECT dolt_commit('-Am','old generation writes')", SkipResultsCheck: true},
			{Query: `SELECT * FROM "postgres/main".public.t`, Expected: []sql.Row{{3, "recreated"}}},
		},
	}, {
		Name: "clone independently created identical schemas",
		SetUpScript: []string{
			"CREATE TABLE seed(id int)",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_branch('other')",
			"SELECT dolt_checkout('other')",
			"SELECT dolt_checkout('main')",
			"CREATE TABLE t(id int PRIMARY KEY,v text UNIQUE)",
			"INSERT INTO t VALUES(1,'main')",
			"SELECT dolt_commit('-Am','main table')",
			"SELECT dolt_checkout('other')",
			"CREATE TABLE t(id int PRIMARY KEY,v text UNIQUE)",
			"INSERT INTO t VALUES(2,'other')",
			"SELECT dolt_commit('-Am','other table')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT dolt_branch('clone')", SkipResultsCheck: true},
			{Query: "SELECT dolt_checkout('clone')", SkipResultsCheck: true},
			{Query: "SELECT * FROM t", Expected: []sql.Row{{2, "other"}}},
			{Query: "INSERT INTO t VALUES(3,'clone')"},
			{Query: "SELECT dolt_commit('-Am','clone writes')", SkipResultsCheck: true},
			{Query: `SELECT * FROM "postgres/other".public.t`, Expected: []sql.Row{{2, "other"}}},
			{Query: `SELECT * FROM "postgres/main".public.t`, Expected: []sql.Row{{1, "main"}}},
			{Query: "SELECT dolt_checkout('-b','mainclone','main')", SkipResultsCheck: true},
			{Query: "SELECT * FROM t", Expected: []sql.Row{{1, "main"}}},
			{Query: "INSERT INTO t VALUES(4,'mainclone')"},
			{Query: "SELECT dolt_commit('-Am','mainclone writes')", SkipResultsCheck: true},
			{Query: `SELECT * FROM "postgres/main".public.t`, Expected: []sql.Row{{1, "main"}}},
			{Query: "INSERT INTO t VALUES(5,'pending')"},
			{Query: "DROP TABLE seed"},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1, "main"}, {4, "mainclone"}, {5, "pending"}}},
			{Query: "SELECT dolt_commit('-Am','drop preserves unrelated writes')", SkipResultsCheck: true},
		},
	}})
}
