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
					_, err = db.Exec(ctx, "COMMIT")
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
