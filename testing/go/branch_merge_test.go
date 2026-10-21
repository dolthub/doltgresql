package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"os"
	"testing"
)

// These isolate the data cases from TestDoltMerge's mixed schema-feature scripts.
// Hashes are deliberately not compared: the engines use different object formats.
func TestDoltMergeData(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{Name: "fast forward ignores no-commit", SetUpScript: []string{
			"CREATE TABLE t (id int PRIMARY KEY, v int)", "INSERT INTO t VALUES (1,10)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"INSERT INTO t VALUES (2,20)", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT fast_forward,conflicts FROM dolt_merge('main','--no-commit')", Expected: []sql.Row{{int64(1), int64(0)}}},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1, 10}, {2, 20}}},
			{Query: "SELECT count(*) FROM dolt_diff_summary('HEAD','WORKING')", Expected: []sql.Row{{int64(0)}}},
		}},
		{Name: "forced two parent merge", SetUpScript: []string{
			"CREATE TABLE t (id int PRIMARY KEY)", "INSERT INTO t VALUES (1)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"INSERT INTO t VALUES (2)", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT fast_forward,conflicts FROM dolt_merge('main','--no-ff','-m','merge_commit')", Expected: []sql.Row{{int64(0), int64(0)}}},
			{Query: "SELECT message FROM dolt_log() LIMIT 1", Expected: []sql.Row{{"merge_commit"}}},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1}, {2}}},
		}},
		{Name: "composite keys and independent columns", SetUpScript: []string{
			"CREATE TABLE t (k1 smallint,k2 bigint,a int,b int,PRIMARY KEY(k1,k2))", "INSERT INTO t VALUES (1,10,100,200)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"UPDATE t SET a=101", "INSERT INTO t VALUES (2,10,300,400)", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
			"UPDATE t SET b=201", "INSERT INTO t VALUES (2,20,500,600)", "SELECT dolt_commit('-Am','other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT fast_forward,conflicts FROM dolt_merge('main')", Expected: []sql.Row{{int64(0), int64(0)}}},
			{Query: "SELECT * FROM t ORDER BY k1,k2", Expected: []sql.Row{{1, 10, 101, 201}, {2, 10, 300, 400}, {2, 20, 500, 600}}},
		}},
		{Name: "delete and unrelated update", SetUpScript: []string{
			"CREATE TABLE t (id int PRIMARY KEY, v int)", "INSERT INTO t VALUES (1,10),(2,20)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"DELETE FROM t WHERE id=1", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
			"UPDATE t SET v=21 WHERE id=2", "SELECT dolt_commit('-Am','other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT conflicts FROM dolt_merge('main')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT * FROM t", Expected: []sql.Row{{2, 21}}},
		}},
		{Name: "keyless independent values merge", SetUpScript: []string{
			"CREATE TABLE t (v text)", "INSERT INTO t VALUES ('x')", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"INSERT INTO t VALUES ('x')", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
			"INSERT INTO t VALUES ('y')", "SELECT dolt_commit('-Am','other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT conflicts FROM dolt_merge('main')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT count(*) FROM t", Expected: []sql.Row{{int64(3)}}},
		}},
		{Name: "identical keyed updates coalesce", SetUpScript: []string{
			"CREATE TABLE t (id int PRIMARY KEY,v text)", "INSERT INTO t VALUES (1,'old')", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"UPDATE t SET v='new'", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
			"UPDATE t SET v='new'", "SELECT dolt_commit('-Am','other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT conflicts FROM dolt_merge('main')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT * FROM t", Expected: []sql.Row{{1, "new"}}},
		}},
		{Name: "empty table branches merge independent inserts", SetUpScript: []string{
			"CREATE TABLE t (id int PRIMARY KEY)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"INSERT INTO t VALUES (1)", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
			"INSERT INTO t VALUES (2)", "SELECT dolt_commit('-Am','other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT conflicts FROM dolt_merge('main')", Expected: []sql.Row{{int64(0)}}},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1}, {2}}},
		}},
		{Name: "no-ff no-commit stages merge parents", SetUpScript: []string{
			"CREATE TABLE t (id int PRIMARY KEY)", "INSERT INTO t VALUES (1)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"INSERT INTO t VALUES (2)", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT fast_forward,conflicts FROM dolt_merge('main','--no-ff','--no-commit')", Expected: []sql.Row{{int64(0), int64(0)}}},
			{Query: "SELECT message FROM dolt_log() LIMIT 1", Expected: []sql.Row{{"base"}}},
			{Query: "SELECT is_merging FROM dolt_merge_status", Expected: []sql.Row{{"t"}}},
			{Query: "SELECT dolt_commit('-m','complete merge')", SkipResultsCheck: true},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1}, {2}}},
		}},
		{Name: "squash fast forward stages without moving HEAD", SetUpScript: []string{
			"CREATE TABLE t (id int PRIMARY KEY)", "INSERT INTO t VALUES (1)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
			"INSERT INTO t VALUES (2)", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')",
		}, Assertions: []ScriptTestAssertion{
			{Query: "SELECT fast_forward,conflicts FROM dolt_merge('main','--squash')", Expected: []sql.Row{{int64(1), int64(0)}}},
			{Query: "SELECT message FROM dolt_log() LIMIT 1", Expected: []sql.Row{{"base"}}},
			{Query: "SELECT dolt_commit('-m','squash commit')", SkipResultsCheck: true},
			{Query: "SELECT * FROM t ORDER BY id", Expected: []sql.Row{{1}, {2}}},
		}},
	})
}

// Intentional divergence: conflicts return status instead of aborting SQL.
// Doltgres currently skips conflict workflows because its merge returns ERROR.
func TestPostgresMergeConflictStatus(t *testing.T) {
	if os.Getenv("DOLTGRES_TEST_SERVER_KIND") != "postgres" {
		t.Skip("PostgreSQL fork's intentional non-aborting conflict contract")
	}
	RunScripts(t, []ScriptTest{{Name: "resolve and commit", SetUpScript: []string{
		"CREATE TABLE t (id int PRIMARY KEY,v int)", "INSERT INTO t VALUES (1,10)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
		"UPDATE t SET v=20", "SELECT dolt_commit('-Am','main')", "SELECT dolt_checkout('other')", "UPDATE t SET v=30", "SELECT dolt_commit('-Am','other')",
	}, Assertions: []ScriptTestAssertion{
		{Query: "SELECT conflicts FROM dolt_merge('main')", Expected: []sql.Row{{int64(1)}}},
		{Query: "SELECT * FROM t", Expected: []sql.Row{{1, 30}}},
		{Query: "SELECT num_conflicts FROM dolt_conflicts", Expected: []sql.Row{{int64(1)}}},
		{Query: "SELECT status FROM dolt_conflicts_resolve('--theirs','t')", Expected: []sql.Row{{int64(0)}}},
		{Query: "SELECT * FROM t", Expected: []sql.Row{{1, 20}}},
		{Query: "SELECT dolt_commit('-Am','resolved')", SkipResultsCheck: true},
		{Query: "SELECT is_merging FROM dolt_merge_status", Expected: []sql.Row{{"f"}}},
	}}})
}
