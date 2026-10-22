package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// Focused branch-DDL slice: real DROP, historical reads, staging and checkout.
// No catalog preparation and no engine-specific expected rows.
func TestBranchCatalogDrop(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "drop changes only selected branch",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY,v text)",
			"INSERT INTO t VALUES(1,'base')",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_branch('other')",
			"SELECT dolt_checkout('other')",
			"SELECT dolt_checkout('main')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "BEGIN"},
			{Query: "DROP TABLE t"},
			{Query: "ROLLBACK"},
			{Query: "SELECT * FROM t", Expected: []sql.Row{{1, "base"}}},
			{Query: "DROP TABLE t"},
			{Query: `SELECT * FROM "postgres/other".public.t`, Expected: []sql.Row{{1, "base"}}},
			{Query: "SELECT * FROM t AS OF 'HEAD'", Expected: []sql.Row{{1, "base"}}},
			{Query: "SELECT from_table_name,to_table_name,diff_type FROM dolt_diff_summary('HEAD','WORKING')", Expected: []sql.Row{{"public.t", "", "dropped"}}},
			{Query: "SELECT dolt_commit('-am','drop')", SkipResultsCheck: true},
			{Query: "SELECT * FROM t AS OF 'HEAD~1'", Expected: []sql.Row{{1, "base"}}},
			{Query: "SELECT count(*) FROM dolt_diff_summary('HEAD','WORKING')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT dolt_checkout('other')", SkipResultsCheck: true},
			{Query: "INSERT INTO t VALUES(2,'other')"},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1, "base"}, {2, "other"}}},
			{Query: "SELECT dolt_commit('-am','other writes')", SkipResultsCheck: true},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: `SELECT * FROM "postgres/other".public.t ORDER BY id`, Expected: []sql.Row{{1, "base"}, {2, "other"}}},
		},
	}})
}
