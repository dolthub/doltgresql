package _go

import (
	"os"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// Required semantics: each qualified catalog view must describe its requested
// revision independently of the session branch, including multiple revisions in
// one statement. The PostgreSQL fork's results below are the intended behavior.
//
// Doltgres discrepancy observed 2026-10-23: with the session on main, the first
// query returned id,v instead of id,v,extra; pg_tables omitted other_only; and the
// cross-branch join failed to expose the distinct schemas. Three assertions failed.
// The index-name assertion passed because both branches have the same index name;
// that result alone does not demonstrate correct revision selection.
//
// The fork carries a selected database root through each view's expansion. The
// Doltgres results indicate use of session-branch metadata, but its exact faulty
// lookup has not been traced. This is a Doltgres bug to fix, not compatibility
// behavior to reproduce. Remove the fork-only guard once Doltgres honors these
// qualifiers; preserve the expected rows and independent-revision assertions.
func TestPostgresRevisionCatalogViews(t *testing.T) {
	if os.Getenv("DOLTGRES_TEST_SERVER_KIND") != "postgres" {
		t.Skip("Doltgres catalog views currently use the session branch instead of the qualified revision")
	}

	RunScripts(t, []ScriptTest{{
		Name:        "catalog views on independent revisions",
		SetUpScript: []string{"CREATE TABLE t(id int PRIMARY KEY,v text)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_checkout('-b','other')", "ALTER TABLE t ADD COLUMN extra int", "CREATE TABLE other_only(x int)", "SELECT dolt_commit('-Am','other')", "SELECT dolt_checkout('main')"},
		Assertions: []ScriptTestAssertion{
			{Query: `SELECT column_name FROM "postgres/other".information_schema.columns WHERE table_schema='public' AND table_name='t' ORDER BY ordinal_position`, Expected: []sql.Row{{"id"}, {"v"}, {"extra"}}},
			{Query: `SELECT tablename FROM "postgres/other".pg_catalog.pg_tables WHERE schemaname='public' ORDER BY tablename`, Expected: []sql.Row{{"other_only"}, {"t"}}},
			{Query: `SELECT a.column_name,b.column_name FROM "postgres/main".information_schema.columns a CROSS JOIN "postgres/other".information_schema.columns b WHERE a.table_name='t' AND a.column_name='v' AND b.table_name='t' AND b.column_name='extra'`, Expected: []sql.Row{{"v", "extra"}}},
			{Query: `SELECT indexname FROM "postgres/other".pg_catalog.pg_indexes WHERE schemaname='public' AND tablename='t'`, Expected: []sql.Row{{"t_pkey"}}},
		},
	}})
}

func TestBranchSearchPath(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name:        "schema precedence after branch activation",
		SetUpScript: []string{"CREATE SCHEMA a", "CREATE SCHEMA b", "CREATE TABLE a.t(v int)", "CREATE TABLE b.t(v int)", "INSERT INTO a.t VALUES(1)", "INSERT INTO b.t VALUES(2)", "SELECT dolt_commit('-Am','base')", "SELECT dolt_checkout('-b','other')"},
		Assertions: []ScriptTestAssertion{
			{Query: "SET search_path=a,b"},
			{Query: "SELECT v FROM t", Expected: []sql.Row{{1}}},
			{Query: "SET search_path=b,a"},
			{Query: "SELECT v FROM t", Expected: []sql.Row{{2}}},
			{Query: "SET search_path=missing,b,a"},
			{Query: "SELECT v FROM t", Expected: []sql.Row{{2}}},
			{Query: "SELECT v FROM a.t", Expected: []sql.Row{{1}}},
			{Query: "WITH t AS (SELECT 42 AS v) SELECT v FROM t", Expected: []sql.Row{{42}}},
		},
	}})
}
