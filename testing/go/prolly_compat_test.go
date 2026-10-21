package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// This subset deliberately uses only implemented version-control operations.
// Expected values follow Doltgres tests; commit hash shape is engine-specific.
// Run it both in-process and against the external PostgreSQL fork.
func TestProllyVersionCompatibility(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{Name: "staging preserves later unstaged changes", SetUpScript: []string{
			"CREATE TABLE t (pk int PRIMARY KEY, v text)",
			"INSERT INTO t VALUES (1,'one')",
			"SELECT dolt_commit('-Am','initial')",
			"UPDATE t SET v='two' WHERE pk=1",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT status FROM dolt_add('.')", Expected: []sql.Row{{int64(0)}}},
			{Query: "UPDATE t SET v='three' WHERE pk=1", Expected: []sql.Row{}},
			{Query: "SELECT hash FROM dolt_commit('-m','staged')", SkipResultsCheck: true},
			{Query: "SELECT message FROM dolt_log() LIMIT 1", Expected: []sql.Row{{"staged"}}},
			{Query: "SELECT count(*) FROM dolt_log()", Expected: []sql.Row{{int64(4)}}},
			{Query: "SELECT * FROM dolt_diff_summary('HEAD','WORKING','t')", Expected: []sql.Row{{"public.t", "public.t", "modified", "t", "f"}}},
			{Query: "SELECT count(*) FROM dolt_diff_summary('HEAD','STAGED')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT v FROM t", Expected: []sql.Row{{"three"}}},
		}},
		{Name: "commit empty flags and named outputs", SetUpScript: []string{
			"CREATE TABLE t (pk int PRIMARY KEY)", "INSERT INTO t VALUES (1)", "SELECT dolt_commit('-Am','initial')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT dolt_commit('-m','empty')", ExpectedErr: "nothing to commit"},
			{Query: "SELECT dolt_commit('--skip-empty','-m','empty')", Expected: []sql.Row{{nil}}},
			{Query: "SELECT hash FROM dolt_commit('--allow-empty','-m','allowed')", SkipResultsCheck: true},
			{Query: "SELECT message FROM dolt_log() LIMIT 1", Expected: []sql.Row{{"allowed"}}},
			{Query: "SELECT count(*) FROM dolt_diff_summary('HEAD~1','HEAD')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT status FROM dolt_branch('saved')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT status FROM dolt_tag('v1','saved')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT message FROM dolt_log('v1') LIMIT 1", Expected: []sql.Row{{"allowed"}}},
		}},
		{Name: "history table changes", SetUpScript: []string{
			"CREATE TABLE t (pk int PRIMARY KEY, v text)", "INSERT INTO t VALUES (1,'a')", "SELECT dolt_commit('-Am','initial')", "SELECT dolt_branch('original')",
			"UPDATE t SET v='b'", "SELECT dolt_commit('-Am','modified')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT * FROM dolt_diff_summary('original','HEAD','t')", Expected: []sql.Row{{"public.t", "public.t", "modified", "t", "f"}}},
			{Query: "SELECT message FROM dolt_log('original') LIMIT 1", Expected: []sql.Row{{"initial"}}},
			{Query: "SELECT message FROM dolt_log('HEAD~1') LIMIT 1", Expected: []sql.Row{{"initial"}}},
			{Query: "SELECT * FROM dolt_diff_summary('HEAD','WORKING')", Expected: []sql.Row{}},
		}},
	})
}
