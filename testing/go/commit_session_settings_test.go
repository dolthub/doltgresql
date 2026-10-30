package _go

import (
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestCommitSessionMetadata(t *testing.T) {
	if testTarget(t) == targetPostgres {
		t.Skip("version-control metadata")
	}
	RunScripts(t, []ScriptTest{{
		Name:        "session identity overrides and explicit argument precedence",
		SetUpScript: []string{"CREATE TABLE settings_rows(id int PRIMARY KEY)", "SELECT dolt_commit('-Am','base')"},
		Assertions: []ScriptTestAssertion{
			{Query: "SET dolt_author_name='Author'"},
			{Query: "SET dolt_author_email='author@example.com'"},
			{Query: "SET dolt_committer_name='Committer'"},
			{Query: "SET dolt_committer_email='committer@example.com'"},
			{Query: "SET dolt_author_date='2020-01-02T03:04:05Z'"},
			{Query: "SET dolt_committer_date='2021-02-03T04:05:06Z'"},
			{Query: "INSERT INTO settings_rows VALUES(1)"},
			{Query: "SELECT dolt_commit('-Am','session')", SkipResultsCheck: true},
			{Query: "SELECT author,author_email,committer,email FROM dolt_log WHERE message='session'", Expected: []sql.Row{{"Author", "author@example.com", "Committer", "committer@example.com"}}},
			{Query: "SELECT author_date='2020-01-02T03:04:05Z'::timestamptz, date='2021-02-03T04:05:06Z'::timestamptz FROM dolt_log WHERE message='session'", Expected: []sql.Row{{"t", "t"}}},
			{Query: "INSERT INTO settings_rows VALUES(2)"},
			{Query: "SELECT dolt_commit('-Am','explicit','--author','Explicit <explicit@example.com>','--date','2022-03-04T05:06:07Z')", SkipResultsCheck: true},
			{Query: "SELECT author,author_email,committer,email FROM dolt_log WHERE message='explicit'", Expected: []sql.Row{{"Explicit", "explicit@example.com", "Committer", "committer@example.com"}}},
			{Query: "SELECT author_date='2022-03-04T05:06:07Z'::timestamptz, date='2021-02-03T04:05:06Z'::timestamptz FROM dolt_log WHERE message='explicit'", Expected: []sql.Row{{"t", "t"}}},
			{Query: "SET dolt_committer_name=''"},
			{Query: "SET dolt_committer_email=''"},
			{Query: "SET dolt_committer_date=''"},
			{Query: "INSERT INTO settings_rows VALUES(3)"},
			{Query: "SELECT dolt_commit('-Am','author fallback','--author','Explicit <explicit@example.com>','--date','2022-03-04T05:06:07Z')", SkipResultsCheck: true},
			{Query: "SELECT author,committer,email, date='2022-03-04T05:06:07Z'::timestamptz FROM dolt_log WHERE message='author fallback'", Expected: []sql.Row{{"Explicit", "Explicit", "explicit@example.com", "t"}}},
		},
	}})
}

func TestPostgresCommitSessionSettings(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native PostgreSQL session settings semantics")
	}
	RunScripts(t, []ScriptTest{{
		Name:        "transactional settings and error cleanup",
		SetUpScript: []string{"CREATE TABLE setting_errors(id int PRIMARY KEY)", "SELECT dolt_commit('-Am','base')"},
		Assertions: []ScriptTestAssertion{
			{Query: "SET dolt_author_name='Session author'"},
			{Query: "BEGIN"},
			{Query: "SET LOCAL dolt_author_name='Local author'"},
			{Query: "SHOW dolt_author_name", Expected: []sql.Row{{"Local author"}}},
			{Query: "ROLLBACK"},
			{Query: "SHOW dolt_author_name", Expected: []sql.Row{{"Session author"}}},
			{Query: "INSERT INTO setting_errors VALUES(1)"},
			{Query: "SET dolt_committer_date='not a date'"},
			{Query: "SELECT dolt_commit('-Am','bad','--committer','Bad <bad@example.com>')", ExpectedErrCode: "22007"},
			{Query: "SELECT count(*) FROM dolt_log WHERE message='bad'", Expected: []sql.Row{{0}}},
			{Query: "RESET dolt_committer_date"},
			{Query: "SELECT dolt_commit('-Am','good')", SkipResultsCheck: true},
			{Query: "SELECT author,committer FROM dolt_log WHERE message='good'", Expected: []sql.Row{{"Session author", "postgres"}}},
			{Query: "RESET dolt_author_name"},
			{Query: "SHOW dolt_author_name", Expected: []sql.Row{{""}}},
			{Query: "SET dolt_transactions_disabled=0", ExpectedErrCode: "42704"},
			{Query: "SET dolt_transactions_disabled=1", ExpectedErrCode: "42704"},
		},
	}})
}

func TestAutomaticCommitSessionSettings(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native transaction boundary integration")
	}
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() { conn.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
	c := conn.Default
	exec := func(q string) { t.Helper(); _, err := c.Exec(ctx, q); require.NoError(t, err, q) }
	count := func(want int) {
		t.Helper()
		var got int
		require.NoError(t, c.QueryRow(ctx, "SELECT count(*) FROM dolt_log").Scan(&got))
		require.Equal(t, want, got)
	}
	exec("CREATE TABLE auto_rows(id int PRIMARY KEY)")
	exec("SELECT dolt_commit('-Am','base')")
	count(2)
	exec("SET dolt_transaction_commit=1")
	exec("SET dolt_transaction_commit_message='  automatic message  '")
	count(2)
	exec("BEGIN")
	exec("INSERT INTO auto_rows VALUES(1)")
	exec("SAVEPOINT discarded")
	exec("INSERT INTO auto_rows VALUES(2)")
	exec("ROLLBACK TO discarded")
	exec("INSERT INTO auto_rows VALUES(3)")
	exec("COMMIT")
	count(3)
	var msg string
	require.NoError(t, c.QueryRow(ctx, "SELECT message FROM dolt_log ORDER BY commit_order DESC LIMIT 1").Scan(&msg))
	require.Equal(t, "automatic message", msg)
	exec("BEGIN")
	exec("INSERT INTO auto_rows VALUES(4)")
	exec("ROLLBACK")
	count(3)
	exec("BEGIN")
	exec("SAVEPOINT only_write")
	exec("INSERT INTO auto_rows VALUES(4)")
	exec("ROLLBACK TO only_write")
	exec("COMMIT")
	count(3)
	exec("UPDATE auto_rows SET id=id")
	count(3)
	other, err := pgx.ConnectConfig(ctx, c.Config().Copy())
	require.NoError(t, err)
	defer other.Close(ctx)
	var setting string
	require.NoError(t, other.QueryRow(ctx, "SHOW dolt_transaction_commit").Scan(&setting))
	require.Equal(t, "off", setting)
	_, err = other.Exec(ctx, "INSERT INTO auto_rows VALUES(4)")
	require.NoError(t, err)
	count(3) // Reading another session's dirty working set must not create a commit.
	exec("SET dolt_transaction_commit_message=''")
	count(3)
	exec("INSERT INTO auto_rows VALUES(5)")
	count(4)
	require.NoError(t, c.QueryRow(ctx, "SELECT message FROM dolt_log ORDER BY commit_order DESC LIMIT 1").Scan(&msg))
	require.Equal(t, "Transaction commit", msg)
	var rows int
	require.NoError(t, c.QueryRow(ctx, "SELECT count(*) FROM auto_rows AS OF 'HEAD'").Scan(&rows))
	require.Equal(t, 4, rows)
	exec("SET dolt_transaction_commit=0")
	exec("BEGIN")
	exec("SET LOCAL dolt_transaction_commit=1")
	exec("INSERT INTO auto_rows VALUES(6)")
	exec("COMMIT")
	count(5)
	require.NoError(t, c.QueryRow(ctx, "SHOW dolt_transaction_commit").Scan(&setting))
	require.Equal(t, "off", setting)
	exec("SET dolt_transaction_commit=1")
	exec("INSERT INTO auto_rows VALUES(7)")
	count(6)
	exec("SELECT dolt_commit('--allow-empty','-m','explicit')")
	count(7) // An explicit version commit must not cause a second automatic commit.
	exec("CREATE TABLE auto_parent(id int PRIMARY KEY)")
	exec("CREATE TABLE auto_child(id int REFERENCES auto_parent DEFERRABLE INITIALLY DEFERRED)")
	count(9)
	exec("BEGIN")
	exec("INSERT INTO auto_child VALUES(99)")
	_, err = c.Exec(ctx, "COMMIT")
	require.Error(t, err)
	count(9)
	exec("SET dolt_committer_date='bad date'")
	_, err = c.Exec(ctx, "INSERT INTO auto_rows VALUES(8)")
	require.Error(t, err)
	exec("RESET dolt_committer_date")
	count(9)
	require.NoError(t, c.QueryRow(ctx, "SELECT count(*) FROM auto_rows WHERE id=8").Scan(&rows))
	require.Zero(t, rows)
	exec("INSERT INTO auto_rows VALUES(8)")
	count(10)
	exec("BEGIN")
	exec("INSERT INTO auto_rows VALUES(9)")
	exec("PREPARE TRANSACTION 'automatic_prepare'")
	count(10)
	require.NoError(t, other.QueryRow(ctx, "SELECT count(*) FROM auto_rows WHERE id=9").Scan(&rows))
	require.Zero(t, rows)
	exec("COMMIT PREPARED 'automatic_prepare'")
	count(11)
	require.NoError(t, other.QueryRow(ctx, "SELECT count(*) FROM auto_rows AS OF 'HEAD' WHERE id=9").Scan(&rows))
	require.Equal(t, 1, rows)
	exec("BEGIN")
	exec("INSERT INTO auto_rows VALUES(10)")
	exec("PREPARE TRANSACTION 'automatic_rollback'")
	exec("ROLLBACK PREPARED 'automatic_rollback'")
	count(11)
	require.NoError(t, c.QueryRow(ctx, "SELECT count(*) FROM auto_rows WHERE id=10").Scan(&rows))
	require.Zero(t, rows)
	exec("BEGIN")
	exec("INSERT INTO auto_rows SELECT generate_series(100,2500)")
	_, err = other.Exec(ctx, "INSERT INTO auto_rows VALUES(3000)")
	require.NoError(t, err)
	exec("COMMIT")
	count(12)
	require.NoError(t, c.QueryRow(ctx, "SELECT count(*) FROM auto_rows AS OF 'HEAD' WHERE id>=100").Scan(&rows))
	require.Equal(t, 2402, rows)
}

func TestMergeCommitSessionControls(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("fork preserves non-aborting merge conflicts")
	}
	for _, setting := range []int{0, 1} {
		RunScripts(t, []ScriptTest{{
			Name: fmt.Sprintf("merge conflicts remain inspectable with allow=%d", setting),
			SetUpScript: []string{
				"CREATE TABLE conflict_settings(id int PRIMARY KEY,v text)",
				"INSERT INTO conflict_settings VALUES(1,'base')", "SELECT dolt_commit('-Am','base')",
				"SELECT dolt_checkout('-b','other')", "UPDATE conflict_settings SET v='theirs'", "SELECT dolt_commit('-Am','theirs')",
				"SELECT dolt_checkout('main')", "UPDATE conflict_settings SET v='ours'", "SELECT dolt_commit('-Am','ours')",
				fmt.Sprintf("SET dolt_allow_commit_conflicts=%d", setting), "SET dolt_transaction_commit=1",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "SELECT conflicts FROM dolt_merge('other')", Expected: []sql.Row{{1}}},
				{Query: "SELECT count(*) FROM dolt_conflicts_conflict_settings", Expected: []sql.Row{{1}}},
				{Query: "SELECT count(*) FROM dolt_log", Expected: []sql.Row{{3}}},
				{Query: "SELECT dolt_commit('-Am','unresolved')", ExpectedErr: "unresolved merge conflicts"},
				{Query: "SELECT dolt_merge('--abort')", SkipResultsCheck: true},
				{Query: "SELECT v FROM conflict_settings", Expected: []sql.Row{{"ours"}}},
			},
		}})
	}
	RunScripts(t, []ScriptTest{{
		Name: "forced merge violations preserve normal DML checks and can be repaired",
		SetUpScript: []string{
			"CREATE TABLE force_parent(a int,b int,PRIMARY KEY(a,b))",
			"CREATE TABLE force_child(id int PRIMARY KEY,p int,q int,FOREIGN KEY(p,q) REFERENCES force_parent(a,b) MATCH FULL)",
			"INSERT INTO force_parent VALUES(1,1),(2,2)", "INSERT INTO force_child VALUES(1,1,1)", "SELECT dolt_commit('-Am','base')",
			"SELECT dolt_checkout('-b','other')", "INSERT INTO force_child VALUES(2,2,2)", "SELECT dolt_commit('-Am','new reference')",
			"SELECT dolt_checkout('main')", "DELETE FROM force_parent WHERE a=2", "SELECT dolt_commit('-Am','remove parent')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT dolt_merge('other')", ExpectedErrCode: "23503"},
			{Query: "SELECT count(*) FROM force_child", Expected: []sql.Row{{1}}},
			{Query: "SET dolt_force_transaction_commit=1"},
			{Query: "SELECT conflicts FROM dolt_merge('other')", Expected: []sql.Row{{1}}},
			{Query: "SELECT violation_type,id,p,q FROM dolt_constraint_violations_force_child", Expected: []sql.Row{{"foreign key", 2, 2, 2}}},
			{Query: "INSERT INTO force_child VALUES(3,99,99)", ExpectedErrCode: "23503"},
			{Query: "INSERT INTO force_child VALUES(3,NULL,1)", ExpectedErrCode: "23503"},
			{Query: "SET dolt_force_transaction_commit=0"},
			{Query: "SELECT dolt_commit('-Am','unresolved')", ExpectedErr: "unresolved constraint violations"},
			{Query: "DELETE FROM force_child WHERE id=2"},
			{Query: "DELETE FROM dolt_constraint_violations_force_child"},
			{Query: "SELECT dolt_commit('-Am','repair')", SkipResultsCheck: true},
			{Query: "SELECT count(*) FROM dolt_constraint_violations", Expected: []sql.Row{{0}}},
			{Query: "SELECT count(*) FROM force_child AS OF 'HEAD'", Expected: []sql.Row{{1}}},
			{Query: "BEGIN"}, {Query: "SET LOCAL dolt_force_transaction_commit=1"}, {Query: "ROLLBACK"},
			{Query: "SHOW dolt_force_transaction_commit", Expected: []sql.Row{{"off"}}},
		},
	}})
}

func TestAutomaticEmptySchemaCommit(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native branch namespace lifecycle")
	}
	RunScripts(t, []ScriptTest{{
		Name:        "automatic commits before the first versioned table",
		SetUpScript: []string{"SET dolt_transaction_commit=1", "CREATE SCHEMA automatic_empty"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT count(*) FROM dolt_log", Expected: []sql.Row{{2}}},
			{Query: "SELECT message FROM dolt_log ORDER BY commit_order DESC LIMIT 1", Expected: []sql.Row{{"Transaction commit"}}},
			{Query: "DROP SCHEMA automatic_empty"},
			{Query: "SELECT count(*) FROM dolt_log", Expected: []sql.Row{{3}}},
			{Query: "SELECT dolt_checkout('-b','restore','HEAD~1')", SkipResultsCheck: true},
			{Query: "SELECT count(*) FROM pg_namespace WHERE nspname='automatic_empty'", Expected: []sql.Row{{1}}},
		},
	}})
}

func TestForcedMergeCollation(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native PostgreSQL RI collation semantics")
	}
	RunScripts(t, []ScriptTest{{
		Name: "forced foreign key reporting uses the referenced collation",
		SetUpScript: []string{
			"CREATE TABLE coll_parent(k text COLLATE \"C\" PRIMARY KEY)",
			"CREATE TABLE coll_child(id int PRIMARY KEY,f text COLLATE \"POSIX\" REFERENCES coll_parent(k))",
			"INSERT INTO coll_parent VALUES('one'),('two')", "SELECT dolt_commit('-Am','base')",
			"SELECT dolt_checkout('-b','other')", "INSERT INTO coll_child VALUES(1,'two')", "SELECT dolt_commit('-Am','child')",
			"SELECT dolt_checkout('main')", "DELETE FROM coll_parent WHERE k='two'", "SELECT dolt_commit('-Am','parent')",
			"SET dolt_force_transaction_commit=1",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT conflicts FROM dolt_merge('other')", Expected: []sql.Row{{1}}},
			{Query: "SELECT id,f FROM dolt_constraint_violations_coll_child", Expected: []sql.Row{{1, "two"}}},
		},
	}})
}
