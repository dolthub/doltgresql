package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// Bootstrap functions/operators retain native overload semantics when their
// inputs come from independently bound revisions. User-defined routine versioning
// is a separate gate; these tests do not claim that it is supported.
func TestBranchBuiltinResolution(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "builtin resolution with independent branch inputs",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY,v text)",
			"INSERT INTO t VALUES(1,'main')",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_checkout('-b','other')",
			"ALTER TABLE t ADD COLUMN extra int",
			"UPDATE t SET v='other',extra=7",
			"SELECT dolt_commit('-Am','other')",
			"SELECT dolt_checkout('main')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: `SELECT upper(v),length(v),id+1 FROM "postgres/other".public.t`, Expected: []sql.Row{{"OTHER", 5, 2}}},
			{Query: `SELECT upper(a.v),upper(b.v),a.id=b.id FROM "postgres/main".public.t a JOIN "postgres/other".public.t b ON a.id=b.id`, Expected: []sql.Row{{"MAIN", "OTHER", "t"}}},
			{Query: `SELECT pg_catalog.lower(v),id OPERATOR(pg_catalog.+) 2 FROM "postgres/other".public.t`, Expected: []sql.Row{{"other", 3}}},
			{Query: `SELECT array_length(ARRAY[id,id+1],1),coalesce(extra,0) FROM "postgres/other".public.t`, Expected: []sql.Row{{2, 7}}},
			{Query: `SELECT id::numeric+0.5 FROM "postgres/other".public.t`, Expected: []sql.Row{{Numeric("1.5")}}},
			{Query: `SET search_path=missing,public,pg_catalog`},
			{Query: `SELECT upper(v),id FROM t`, Expected: []sql.Row{{"MAIN", 1}}},
			{Query: `SELECT upper(v),id FROM "postgres/other".public.t`, Expected: []sql.Row{{"OTHER", 1}}},
		},
	}})
}
