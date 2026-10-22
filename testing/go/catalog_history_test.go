package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// Extends TestAsOf's cross-revision joins to incompatible table definitions.
// No preparation or engine-specific setup is permitted.
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

	RunScripts(t, []ScriptTest{{Name: "incompatible retained schemas", SetUpScript: setup, Assertions: []ScriptTestAssertion{
		{Query: "SELECT old.v,cur.v,cur.note FROM t AS OF 'HEAD~1' old JOIN t AS OF 'HEAD' cur USING(id)", Expected: []sql.Row{{10, "ten", "new schema"}}},
		{Query: "SELECT (old).v,(cur).v FROM t AS OF 'HEAD~1' old JOIN t AS OF 'HEAD' cur USING(id)", Expected: []sql.Row{{10, "ten"}}},
		{Query: "SELECT row_to_json(old)::text,row_to_json(cur)::text FROM t AS OF 'HEAD~1' old JOIN t AS OF 'HEAD' cur USING(id)", Expected: []sql.Row{{`{"id":1,"v":10}`, `{"id":1,"v":"ten","note":"new schema"}`}}},
		{Query: "SELECT * FROM removed AS OF 'HEAD~1'", Expected: []sql.Row{{42}}},
		{Query: "SELECT count(*) FROM t AS OF 'HEAD~1' old JOIN t cur USING(id)", Expected: []sql.Row{{int64(1)}}},
		{Query: "CREATE TABLE restored (id int,v int)"},
		{Query: "INSERT INTO restored SELECT * FROM t AS OF 'HEAD~1'"},
		{Query: "SELECT * FROM restored", Expected: []sql.Row{{1, 10}}},
		{Query: "UPDATE t cur SET note=old.v::text FROM t AS OF 'HEAD~1' old WHERE cur.id=old.id"},
		{Query: "SELECT note FROM t", Expected: []sql.Row{{"10"}}},
	}}})
}
