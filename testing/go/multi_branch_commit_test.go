package _go

import (
	"errors"
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// These cases adapt DoltMultiBranchCommitTests to PostgreSQL transaction errors:
// a failed COMMIT rolls back, rather than retaining a transaction for retry.
func TestMultiBranchAutomaticCommit(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native PostgreSQL transaction integration")
	}
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() { conn.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
	c := conn.Default
	exec := func(q string) { t.Helper(); _, err := c.Exec(ctx, q); require.NoError(t, err, q) }
	other, err := pgx.ConnectConfig(ctx, c.Config().Copy())
	require.NoError(t, err)
	defer other.Close(ctx)
	check := func(c *pgx.Conn, branch string, head bool, want int) {
		t.Helper()
		q := fmt.Sprintf(`SELECT count(*) FROM "postgres/%s".public.multi_rows`, branch)
		if head {
			q += " AS OF 'HEAD'"
		}
		var got int
		require.NoError(t, c.QueryRow(ctx, q).Scan(&got), q)
		require.Equal(t, want, got, q)
	}
	exec("CREATE TABLE multi_rows(id int PRIMARY KEY)")
	exec("SELECT dolt_commit('-Am','base')")
	exec("SELECT dolt_branch('b1')")
	exec("SELECT dolt_branch('b2')")
	var enabled bool
	require.NoError(t, c.QueryRow(ctx, "SELECT current_setting('dolt_multi_branch_commit')::boolean").Scan(&enabled))
	require.False(t, enabled)
	exec("BEGIN")
	exec(`INSERT INTO "postgres/main".public.multi_rows VALUES(1)`)
	exec(`INSERT INTO "postgres/b1".public.multi_rows VALUES(2)`)
	exec("COMMIT")
	for _, b := range []string{"main", "b1"} {
		check(c, b, false, 1)
		check(c, b, true, 0)
	}
	exec("SET dolt_transaction_commit=1")
	exec("BEGIN")
	exec(`INSERT INTO "postgres/main".public.multi_rows VALUES(3)`)
	exec(`INSERT INTO "postgres/b1".public.multi_rows VALUES(4)`)
	_, err = c.Exec(ctx, "COMMIT")
	require.ErrorContains(t, err, "dolt_multi_branch_commit")
	for _, b := range []string{"main", "b1"} {
		check(c, b, false, 1)
		check(c, b, true, 0)
	}
	exec("SET dolt_multi_branch_commit=1")
	exec("SET dolt_transaction_commit_message=' across branches '")
	exec("SELECT dolt_checkout('b2')")
	exec("BEGIN")
	exec(`INSERT INTO "postgres/b1".public.multi_rows VALUES(5)`)
	exec(`INSERT INTO "postgres/main".public.multi_rows VALUES(6)`)
	exec("SAVEPOINT discard_branch")
	exec(`INSERT INTO "postgres/b2".public.multi_rows VALUES(7)`)
	exec("ROLLBACK TO discard_branch")
	for _, b := range []string{"main", "b1"} {
		check(other, b, false, 1)
		check(other, b, true, 0)
	}
	exec("COMMIT")
	for _, b := range []string{"main", "b1"} {
		check(other, b, false, 2)
		check(other, b, true, 2)
	}
	check(c, "b2", true, 0)
	var sameDate bool
	require.NoError(t, c.QueryRow(ctx, `SELECT m.date=b.date FROM dolt_log('main') m JOIN dolt_log('b1') b USING(message) WHERE m.message='across branches'`).Scan(&sameDate))
	require.True(t, sameDate)
	var branch string
	require.NoError(t, c.QueryRow(ctx, "SELECT active_branch()").Scan(&branch))
	require.Equal(t, "b2", branch)
	// Every head and data write stays hidden until COMMIT PREPARED.
	exec("BEGIN")
	exec(`INSERT INTO "postgres/main".public.multi_rows VALUES(8)`)
	exec(`INSERT INTO "postgres/b1".public.multi_rows VALUES(9)`)
	exec("PREPARE TRANSACTION 'multi_success'")
	for _, b := range []string{"main", "b1"} {
		check(other, b, false, 2)
		check(other, b, true, 2)
	}
	exec("COMMIT PREPARED 'multi_success'")
	for _, b := range []string{"main", "b1"} {
		check(other, b, false, 3)
		check(other, b, true, 3)
	}
	exec("BEGIN")
	exec(`INSERT INTO "postgres/main".public.multi_rows VALUES(10)`)
	exec(`INSERT INTO "postgres/b1".public.multi_rows VALUES(11)`)
	exec("PREPARE TRANSACTION 'multi_abort'")
	exec("ROLLBACK PREPARED 'multi_abort'")
	for _, b := range []string{"main", "b1"} {
		check(other, b, false, 3)
		check(other, b, true, 3)
	}
	// Invalid metadata must roll back every branch's data as well as its HEAD.
	exec("SET dolt_committer_date='invalid'")
	exec("BEGIN")
	exec(`INSERT INTO "postgres/main".public.multi_rows VALUES(12)`)
	exec(`INSERT INTO "postgres/b1".public.multi_rows VALUES(13)`)
	_, err = c.Exec(ctx, "COMMIT")
	require.Error(t, err)
	exec("RESET dolt_committer_date")
	for _, b := range []string{"main", "b1"} {
		check(c, b, false, 3)
		check(c, b, true, 3)
	}
	require.NoError(t, other.QueryRow(ctx, "SELECT current_setting('dolt_multi_branch_commit')::boolean").Scan(&enabled))
	require.False(t, enabled)
}

// Derived from Dolt's concurrent multi-branch insert tests. PostgreSQL may
// choose a deadlock victim; retry that entire transaction, never one branch.
func TestMultiBranchConcurrentCommit(t *testing.T) {
	testMultiBranchConcurrentCommit(t, false)
}

func TestMultiBranchExplicitConcurrentCommit(t *testing.T) {
	testMultiBranchConcurrentCommit(t, true)
}

func testMultiBranchConcurrentCommit(t *testing.T, explicit bool) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native PostgreSQL transaction integration")
	}
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() { conn.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
	c := conn.Default
	for _, q := range []string{"CREATE TABLE multi_concurrent(id int PRIMARY KEY)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('b1')", "SELECT dolt_branch('b2')"} {
		_, err := c.Exec(ctx, q)
		require.NoError(t, err)
	}
	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	done := make(chan error, 2)
	for worker := 0; worker < 2; worker++ {
		go func(worker int) {
			db, err := pgx.ConnectConfig(ctx, c.Config().Copy())
			if err != nil {
				done <- err
				ready <- struct{}{}
				return
			}
			defer db.Close(ctx)
			_, err = db.Exec(ctx, "SET dolt_transaction_commit=1; SET dolt_multi_branch_commit=1; SET statement_timeout='10s'")
			if err != nil {
				done <- err
				ready <- struct{}{}
				return
			}
			branches := []string{"main", "b1", "b2"}
			if worker == 1 {
				branches = []string{"b2", "b1", "main"}
			}
			for attempt := 0; attempt < 5; attempt++ {
				_, err = db.Exec(ctx, "BEGIN")
				for _, branch := range branches {
					if err == nil {
						_, err = db.Exec(ctx, fmt.Sprintf(`INSERT INTO "postgres/%s".public.multi_concurrent VALUES(%d)`, branch, worker+1))
					}
				}
				if attempt == 0 {
					ready <- struct{}{}
					<-start
				}
				if err == nil {
					query := "COMMIT"
					if explicit {
						query = "SELECT * FROM dolt_commit_all('-Am','concurrent')"
					}
					_, err = db.Exec(ctx, query)
				}
				if err == nil {
					done <- nil
					return
				}
				_, _ = db.Exec(ctx, "ROLLBACK")
				var pgerr *pgconn.PgError
				if !errors.As(err, &pgerr) || (pgerr.Code != "40P01" && pgerr.Code != "40001") {
					done <- err
					return
				}
			}
			done <- err
		}(worker)
	}
	<-ready
	<-ready
	close(start)
	require.NoError(t, <-done)
	require.NoError(t, <-done)
	for _, branch := range []string{"main", "b1", "b2"} {
		for _, suffix := range []string{"", " AS OF 'HEAD'"} {
			var n int
			require.NoError(t, c.QueryRow(ctx, fmt.Sprintf(`SELECT count(*) FROM "postgres/%s".public.multi_concurrent%s`, branch, suffix)).Scan(&n))
			require.Equal(t, 2, n, branch+suffix)
		}
	}
}

// Qualified DML must keep defaults, constraints, and row layouts on its target
// branch, even when a statement reads the same logical table on another branch.
func TestMultiBranchWriteBinding(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native branch-qualified write integration")
	}
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() { conn.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
	c := conn.Default
	exec := func(q string) { t.Helper(); _, err := c.Exec(ctx, q); require.NoError(t, err, q) }
	exec("CREATE TABLE branch_parent(id int PRIMARY KEY)")
	exec("CREATE TABLE branch_child(id int PRIMARY KEY, p int REFERENCES branch_parent, v int DEFAULT 10 CHECK(v>0))")
	exec("CREATE TABLE branch_audit(id int)")
	exec("CREATE FUNCTION record_branch_write() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN INSERT INTO branch_audit VALUES(NEW.id); RETURN NEW; END $$")
	exec("CREATE TRIGGER branch_write AFTER INSERT ON branch_child FOR EACH ROW EXECUTE FUNCTION record_branch_write()")
	exec("INSERT INTO branch_parent VALUES(1)")
	exec("SELECT dolt_commit('-Am','base')")
	exec("SELECT dolt_branch('b1')")
	exec("SELECT dolt_checkout('b1')")
	exec("ALTER TABLE branch_child ADD COLUMN extra text DEFAULT 'branch'")
	exec("ALTER TABLE branch_child ALTER COLUMN v SET DEFAULT 20")
	exec("INSERT INTO branch_parent VALUES(2)")
	exec("SELECT dolt_commit('-Am','branch schema')")
	exec("SELECT dolt_checkout('main')")
	exec("SET dolt_transaction_commit=1")
	exec("SET dolt_multi_branch_commit=1")
	exec("BEGIN")
	exec(`INSERT INTO "postgres/b1".public.branch_child(id,p) VALUES(2,2)`)
	exec(`INSERT INTO "postgres/main".public.branch_child(id,p) VALUES(1,1)`)
	exec(`UPDATE "postgres/b1".public.branch_child SET v=(SELECT v+1 FROM "postgres/main".public.branch_child WHERE id=1) WHERE id=2`)
	exec("COMMIT")
	var v int
	var extra string
	require.NoError(t, c.QueryRow(ctx, `SELECT v,extra FROM "postgres/b1".public.branch_child AS OF 'HEAD' WHERE id=2`).Scan(&v, &extra))
	require.Equal(t, 11, v)
	require.Equal(t, "branch", extra)
	_, err := c.Exec(ctx, `INSERT INTO "postgres/main".public.branch_child(id,p) VALUES(3,2)`)
	require.ErrorContains(t, err, "foreign key")
	exec(`PREPARE bound_write(int) AS INSERT INTO "postgres/b1".public.branch_child(id,p) VALUES($1,2)`)
	exec("EXECUTE bound_write(3)")
	exec("SELECT dolt_checkout('b1')")
	exec("EXECUTE bound_write(4)")
	exec("SELECT dolt_checkout('main')")
	require.NoError(t, c.QueryRow(ctx, `SELECT v FROM "postgres/b1".public.branch_child WHERE id=4`).Scan(&v))
	require.Equal(t, 20, v)
	exec(`DELETE FROM "postgres/b1".public.branch_child WHERE id IN (3,4)`)
	var count int
	require.NoError(t, c.QueryRow(ctx, `SELECT count(*) FROM "postgres/b1".public.branch_child AS OF 'HEAD'`).Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, c.QueryRow(ctx, `SELECT count(*) FROM "postgres/b1".public.branch_audit`).Scan(&count))
	require.Equal(t, 3, count)
	require.NoError(t, c.QueryRow(ctx, `SELECT count(*) FROM "postgres/main".public.branch_audit`).Scan(&count))
	require.Equal(t, 1, count)
	// A historical qualifier must never become a writable working-set target.
	_, err = c.Exec(ctx, `INSERT INTO "postgres/HEAD~1".public.branch_child(id,p) VALUES(5,1)`)
	require.Error(t, err)
}

func TestMultiBranchPendingMerge(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("fork preserves non-aborting merge conflicts")
	}
	RunScripts(t, []ScriptTest{{
		Name: "a pending merge remains inspectable while another branch commits",
		SetUpScript: []string{
			"CREATE TABLE multi_conflict(id int PRIMARY KEY,v text)",
			"INSERT INTO multi_conflict VALUES(1,'base')", "SELECT dolt_commit('-Am','base')",
			"SELECT dolt_checkout('-b','other')", "UPDATE multi_conflict SET v='theirs'", "SELECT dolt_commit('-Am','theirs')",
			"SELECT dolt_checkout('main')", "UPDATE multi_conflict SET v='ours'", "SELECT dolt_commit('-Am','ours')",
			"SET dolt_transaction_commit=1", "SET dolt_multi_branch_commit=1",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT conflicts FROM dolt_merge('other')", Expected: []sql.Row{{1}}},
			{Query: "BEGIN"},
			{Query: `INSERT INTO "postgres/main".public.multi_conflict VALUES(2,'main work')`},
			{Query: `INSERT INTO "postgres/other".public.multi_conflict VALUES(3,'other work')`},
			{Query: "COMMIT"},
			{Query: "SELECT count(*) FROM multi_conflict", Expected: []sql.Row{{2}}},
			{Query: "SELECT count(*) FROM multi_conflict AS OF 'HEAD'", Expected: []sql.Row{{1}}},
			{Query: `SELECT count(*) FROM "postgres/other".public.multi_conflict AS OF 'HEAD'`, Expected: []sql.Row{{2}}},
			{Query: "SELECT count(*) FROM dolt_conflicts_multi_conflict", Expected: []sql.Row{{1}}},
			{Query: "SELECT dolt_commit('-Am','unresolved')", ExpectedErr: "unresolved merge conflicts"},
		},
	}})
}

func TestMultiBranchExplicitCommit(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native explicit transaction commit integration")
	}
	for _, fn := range []string{"dolt_commit", "dolt_commit_all"} {
		t.Run(fn, func(t *testing.T) {
			ctx, conn, controller := CreateServer(t, "postgres")
			defer func() { conn.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
			c := conn.Default
			exec := func(q string) { t.Helper(); _, err := c.Exec(ctx, q); require.NoError(t, err, q) }
			other, err := pgx.ConnectConfig(ctx, c.Config().Copy())
			require.NoError(t, err)
			defer other.Close(ctx)
			count := func(db *pgx.Conn, branch string, head bool, want int) {
				t.Helper()
				q := fmt.Sprintf(`SELECT count(*) FROM "postgres/%s".public.explicit_rows`, branch)
				if head {
					q += " AS OF 'HEAD'"
				}
				var got int
				require.NoError(t, db.QueryRow(ctx, q).Scan(&got))
				require.Equal(t, want, got, q)
			}
			exec("CREATE TABLE explicit_rows(id int PRIMARY KEY)")
			exec("SELECT dolt_commit('-Am','base')")
			exec("SELECT dolt_branch('b1')")
			exec("SELECT dolt_branch('b2')")
			exec("SELECT dolt_checkout('b2')")
			if fn == "dolt_commit" {
				exec("SET dolt_multi_branch_commit=1")
			}
			exec("BEGIN")
			exec(`INSERT INTO "postgres/main".public.explicit_rows VALUES(1)`)
			exec("SAVEPOINT retained")
			exec(`INSERT INTO "postgres/b1".public.explicit_rows VALUES(2)`)
			count(other, "main", false, 0)
			count(other, "b1", false, 0)
			rows, err := c.Query(ctx, "SELECT * FROM "+fn+"('-Am','both')")
			require.NoError(t, err)
			var names, hashes []string
			for rows.Next() {
				var name, hash string
				if fn == "dolt_commit_all" {
					require.NoError(t, rows.Scan(&name, &hash))
					names = append(names, name)
				} else {
					require.NoError(t, rows.Scan(&hash))
				}
				hashes = append(hashes, hash)
			}
			require.NoError(t, rows.Err())
			rows.Close()
			require.Len(t, hashes, 2)
			if fn == "dolt_commit_all" {
				require.Equal(t, []string{"b1", "main"}, names)
			}
			for _, h := range hashes {
				require.Len(t, h, 32)
			}
			require.Equal(t, byte('I'), c.PgConn().TxStatus())
			for _, b := range []string{"main", "b1"} {
				count(other, b, false, 1)
				count(other, b, true, 1)
			}
			count(other, "b2", true, 0)
			// An error in a later branch must roll back an earlier ref publication.
			exec("SELECT dolt_checkout('main')")
			exec("BEGIN")
			exec(`INSERT INTO "postgres/b1".public.explicit_rows VALUES(3)`)
			exec("INSERT INTO explicit_rows VALUES(4)")
			exec("DELETE FROM explicit_rows WHERE id=4")
			_, err = c.Exec(ctx, "SELECT * FROM "+fn+"('-Am','must rollback')")
			require.ErrorContains(t, err, "nothing to commit")
			exec("ROLLBACK")
			for _, b := range []string{"main", "b1"} {
				count(other, b, false, 1)
				count(other, b, true, 1)
			}
			// --skip-empty omits unchanged branches but still commits the SQL transaction.
			exec("BEGIN")
			exec(`INSERT INTO "postgres/b1".public.explicit_rows VALUES(5)`)
			exec("INSERT INTO explicit_rows VALUES(6)")
			exec("DELETE FROM explicit_rows WHERE id=6")
			rows, err = c.Query(ctx, "SELECT * FROM "+fn+"('-Am','skip unchanged','--skip-empty')")
			require.NoError(t, err)
			got := 0
			for rows.Next() {
				got++
			}
			require.NoError(t, rows.Err())
			rows.Close()
			require.Equal(t, 1, got)
			require.Equal(t, byte('I'), c.PgConn().TxStatus())
			count(other, "main", true, 1)
			count(other, "b1", true, 2)
			// Native simple-query batches also end at the explicit commit function.
			_, err = c.Exec(ctx, `BEGIN; INSERT INTO explicit_rows VALUES(7); SELECT * FROM `+fn+`('-Am','simple'); SELECT 1/0`, pgx.QueryExecModeSimpleProtocol)
			require.Error(t, err)
			exec("ROLLBACK")
			count(other, "main", true, 2)
		})
	}
}

func TestExplicitCommitDeferredTriggers(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native deferred-trigger commit integration")
	}
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() { conn.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
	c := conn.Default
	exec := func(q string) { t.Helper(); _, err := c.Exec(ctx, q); require.NoError(t, err, q) }
	exec("CREATE TABLE deferred_source(id int PRIMARY KEY)")
	exec("CREATE TABLE deferred_sink(id int PRIMARY KEY)")
	exec("CREATE FUNCTION deferred_copy() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN INSERT INTO deferred_sink VALUES(NEW.id); RETURN NEW; END $$")
	exec("CREATE CONSTRAINT TRIGGER deferred_copy AFTER INSERT ON deferred_source DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION deferred_copy()")
	exec("SELECT dolt_commit('-Am','base')")
	exec("BEGIN")
	exec("INSERT INTO deferred_source VALUES(1)")
	exec("SELECT dolt_commit('-Am','with trigger')")
	require.Equal(t, byte('I'), c.PgConn().TxStatus())
	var n int
	require.NoError(t, c.QueryRow(ctx, "SELECT count(*) FROM deferred_sink AS OF 'HEAD'").Scan(&n))
	require.Equal(t, 1, n)
	exec("CREATE FUNCTION cursor_copy(n int) RETURNS int LANGUAGE plpgsql AS $$ BEGIN INSERT INTO deferred_sink VALUES(n); RETURN n; END $$")
	exec("BEGIN")
	exec("DECLARE pending_cursor CURSOR WITH HOLD FOR SELECT cursor_copy(n) FROM generate_series(2,3) n")
	exec("SELECT dolt_commit('-Am','held cursor writes')")
	require.Equal(t, byte('I'), c.PgConn().TxStatus())
	require.NoError(t, c.QueryRow(ctx, "SELECT count(*) FROM deferred_sink AS OF 'HEAD'").Scan(&n))
	require.Equal(t, 3, n)
	rows, err := c.Query(ctx, "FETCH ALL FROM pending_cursor")
	require.NoError(t, err)
	values := []int{}
	for rows.Next() {
		var v int
		require.NoError(t, rows.Scan(&v))
		values = append(values, v)
	}
	require.NoError(t, rows.Err())
	rows.Close()
	require.Equal(t, []int{2, 3}, values)
	exec("CLOSE pending_cursor")
	exec("CREATE TABLE deferred_fk(id int REFERENCES deferred_source DEFERRABLE INITIALLY DEFERRED)")
	exec("SELECT dolt_commit('-Am','constraint')")
	exec("BEGIN")
	exec("INSERT INTO deferred_fk VALUES(99)")
	_, err = c.Exec(ctx, "SELECT dolt_commit('-Am','bad constraint')")
	require.ErrorContains(t, err, "foreign key")
	exec("ROLLBACK")
	require.NoError(t, c.QueryRow(ctx, "SELECT count(*) FROM dolt_log WHERE message='bad constraint'").Scan(&n))
	require.Equal(t, 0, n)
}

func TestMultiBranchCommitProtocol(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native commit function protocol integration")
	}
	RunMessageFlowTests(t, []MessageFlowTest{{
		Name:        "explicit commit publishes before Sync and completes its result",
		SetUpScript: []string{"CREATE TABLE protocol_rows(id int PRIMARY KEY)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('b1')"},
		Steps: []FlowStep{
			SimpleQuery{Query: `BEGIN; INSERT INTO protocol_rows VALUES(1); INSERT INTO "postgres/b1".public.protocol_rows VALUES(2)`, Expected: []StatementResult{{Tag: "BEGIN"}, {Tag: "INSERT 0 1"}, {Tag: "INSERT 0 1"}}, ExpectedReadyStatus: 'T'},
			Parse{Name: "commit_all", Query: "SELECT branch FROM dolt_commit_all('-Am','wire')"},
			Bind{PreparedStatement: "commit_all", Portal: "commit_result"},
			Execute{Portal: "commit_result", MaxRows: 1, Tag: "SELECT 2", Rows: [][]string{{"b1"}, {"main"}}},
			Flush{},
			QueryOnOtherConnection{Query: `SELECT count(*) FROM "postgres/b1".public.protocol_rows AS OF 'HEAD'`, Expected: [][]string{{"1"}}},
			QueryOnOtherConnection{Query: "SELECT count(*) FROM protocol_rows AS OF 'HEAD'", Expected: [][]string{{"1"}}},
			Sync{ExpectedReadyStatus: 'I'},
		},
	}})
}

func TestMultiBranchPartialStaging(t *testing.T) {
	if testTarget(t) != targetPostgresFork {
		t.Skip("native multi-branch staging")
	}
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() { conn.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
	c := conn.Default
	exec := func(q string) { t.Helper(); _, err := c.Exec(ctx, q); require.NoError(t, err, q) }
	exec("CREATE TABLE staged_rows(id int PRIMARY KEY)")
	exec("SELECT dolt_commit('-Am','base')")
	exec("SELECT dolt_branch('b1')")
	exec("INSERT INTO staged_rows VALUES(1)")
	exec("SELECT dolt_add('-A')")
	exec("SELECT dolt_checkout('b1')")
	exec("INSERT INTO staged_rows VALUES(2)")
	exec("SELECT dolt_add('-A')")
	exec("SELECT dolt_checkout('main')")
	exec("BEGIN")
	exec("INSERT INTO staged_rows VALUES(3)")
	exec(`INSERT INTO "postgres/b1".public.staged_rows VALUES(4)`)
	exec("SET LOCAL dolt_transaction_commit=1")
	exec("SET LOCAL dolt_multi_branch_commit=1")
	exec("SELECT * FROM dolt_commit_all('-m','staged only')")
	require.Equal(t, byte('I'), c.PgConn().TxStatus())
	for _, b := range []string{"main", "b1"} {
		var n int
		require.NoError(t, c.QueryRow(ctx, fmt.Sprintf(`SELECT count(*) FROM "postgres/%s".public.staged_rows`, b)).Scan(&n))
		require.Equal(t, 2, n)
		require.NoError(t, c.QueryRow(ctx, fmt.Sprintf(`SELECT count(*) FROM "postgres/%s".public.staged_rows AS OF 'HEAD'`, b)).Scan(&n))
		require.Equal(t, 1, n)
	}
	exec("BEGIN")
	exec("INSERT INTO staged_rows VALUES(5)")
	exec(`INSERT INTO "postgres/b1".public.staged_rows VALUES(6)`)
	_, err := c.Exec(ctx, "SELECT dolt_commit('-Am','disabled')")
	require.ErrorContains(t, err, "multiple dirty branches")
	exec("ROLLBACK")
	// Empty all-branch commits use the selected branch, and skip-empty returns no rows.
	exec("SELECT dolt_commit_all('--allow-empty','-m','empty selected')")
	rows, err := c.Query(ctx, "SELECT * FROM dolt_commit_all('--skip-empty','-m','skip')")
	require.NoError(t, err)
	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
	rows.Close()
	require.Equal(t, byte('I'), c.PgConn().TxStatus())
}
