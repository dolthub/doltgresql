package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// Extends TestAsOf's cross-revision joins to incompatible table definitions.
// Preparation is specific to the native catalog experiment, not query semantics.
func TestCatalogHistory(t *testing.T) {
	setup := []string{
		"CREATE TABLE t (id int PRIMARY KEY, v int)",
		"INSERT INTO t VALUES (1,10)",
		"CREATE TABLE removed (id int)",
		"INSERT INTO removed VALUES (42)",
		"SELECT dolt_commit('-Am','integer schema')",
		"DROP TABLE t", "DROP TABLE removed",
		"CREATE TABLE t (id int PRIMARY KEY, v text, note text)",
		"INSERT INTO t VALUES (1,'ten','new schema')",
		"SELECT dolt_commit('-Am','text schema')",
	}
	if externalPostgres() {
		setup = append(setup, "SELECT dolt_catalog_materialize('HEAD~1')", "SELECT dolt_catalog_materialize('HEAD')")
	}
	RunScripts(t, []ScriptTest{{Name: "incompatible retained schemas", SetUpScript: setup, Assertions: []ScriptTestAssertion{
		{Query: "SELECT old.v,cur.v,cur.note FROM t AS OF 'HEAD~1' old JOIN t AS OF 'HEAD' cur USING(id)", Expected: []sql.Row{{10, "ten", "new schema"}}},
		{Query: "SELECT * FROM removed AS OF 'HEAD~1'", Expected: []sql.Row{{42}}},
		{Query: "SELECT count(*) FROM t AS OF 'HEAD~1' old JOIN t cur USING(id)", Expected: []sql.Row{{int64(1)}}},
	}}})
}
