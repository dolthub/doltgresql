package _go

import (
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/sql"
)

// Same SELECT queries and expectations used by enginetest.TestVersionedQueries.
// Setup is the PostgreSQL spelling of Dolt's VersionedQuerySetup; all tables
// are staged together because this slice does not implement table-specific add.
// SHOW/DESCRIBE and MySQL @variables remain outside this PostgreSQL query slice.
func TestNativeVersionedQueries(t *testing.T) {
	setup := []string{
		"CREATE TABLE myhistorytable (i bigint PRIMARY KEY,s text NOT NULL)",
		"INSERT INTO myhistorytable VALUES (1,'first row, 1'),(2,'second row, 1'),(3,'third row, 1')",
		"SELECT dolt_commit('-Am','create table')", "SELECT dolt_tag('2019-01-01','HEAD')",
		"TRUNCATE myhistorytable",
		"INSERT INTO myhistorytable VALUES (1,'first row, 2'),(2,'second row, 2'),(3,'third row, 2')",
		"SELECT dolt_commit('-Am','create table')", "SELECT dolt_tag('2019-01-02','HEAD')",
		"DROP TABLE myhistorytable", "CREATE TABLE myhistorytable (i bigint PRIMARY KEY,s text NOT NULL,c text NOT NULL)",
		"INSERT INTO myhistorytable VALUES (1,'first row, 3','1'),(2,'second row, 3','2'),(3,'third row, 3','3')",
		"SELECT dolt_commit('-Am','create table')", "SELECT dolt_tag('2019-01-03','HEAD')",
	}
	var assertions []ScriptTestAssertion
	for _, q := range queries.VersionedQueries {
		if strings.HasPrefix(q.Query, "SELECT ") {
			assertions = append(assertions, ScriptTestAssertion{Query: q.Query, Expected: q.Expected})
		} else {
			t.Logf("Outside SELECT slice (also excluded by Doltgres enginetest): %s", q.Query)
		}
	}
	RunScripts(t, []ScriptTest{{Name: "upstream versioned SELECT queries", SetUpScript: setup, Assertions: assertions}})
}

// Isolates the historical-cursor assertions from TestCursors without requiring
// dolt_history tables or Doltgres's implicit SQL commit inside BEGIN.
func TestNativeHistoricalCursors(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "retained cursor bindings", SetUpScript: []string{
		"CREATE TABLE hist (id int PRIMARY KEY,v text)", "INSERT INTO hist VALUES (1,'one')", "SELECT dolt_commit('-Am','first')",
		"INSERT INTO hist VALUES (2,'two')", "UPDATE hist SET v='ONE' WHERE id=1", "SELECT dolt_commit('-am','second')",
		"DELETE FROM hist WHERE id=1", "INSERT INTO hist VALUES (3,'three')", "SELECT dolt_commit('-am','third')",
	}, Assertions: []ScriptTestAssertion{
		{Query: "DECLARE wh CURSOR WITH HOLD FOR SELECT * FROM hist AS OF 'HEAD~1' ORDER BY id"},
		{Query: "BEGIN"},
		{Query: "DECLARE h2 CURSOR FOR SELECT * FROM hist AS OF 'HEAD~2' ORDER BY id"},
		{Query: "DECLARE h1 SCROLL CURSOR FOR SELECT * FROM hist AS OF 'HEAD~1' ORDER BY id"},
		{Query: "DECLARE head CURSOR FOR SELECT * FROM hist AS OF 'HEAD' ORDER BY id"},
		{Query: "DECLARE j CURSOR FOR SELECT cur.id,cur.v,old.v FROM hist cur JOIN hist AS OF 'HEAD~1' old ON cur.id=old.id ORDER BY cur.id"},
		{Query: "INSERT INTO hist VALUES (4,'four')"},
		{Query: "UPDATE hist SET v='TWO'"},
		{Query: "FETCH ALL FROM h2", Expected: []sql.Row{{1, "one"}}},
		{Query: "FETCH 1 FROM h1", Expected: []sql.Row{{1, "ONE"}}},
		{Query: "FETCH ALL FROM h1", Expected: []sql.Row{{2, "two"}}},
		{Query: "FETCH BACKWARD ALL FROM h1", Expected: []sql.Row{{2, "two"}, {1, "ONE"}}},
		{Query: "FETCH ALL FROM head", Expected: []sql.Row{{2, "two"}, {3, "three"}}},
		{Query: "FETCH ALL FROM j", Expected: []sql.Row{{2, "two", "two"}}},
		{Query: "ROLLBACK"},
		{Query: "FETCH ALL FROM wh", Expected: []sql.Row{{1, "ONE"}, {2, "two"}}},
	}}})
}

func TestRevisionWorkingSets(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "independent branch references", SetUpScript: []string{
		"CREATE TABLE t (id int PRIMARY KEY,v text)", "INSERT INTO t VALUES (1,'base')", "SELECT dolt_commit('-Am','base')", "SELECT dolt_branch('other')",
		"UPDATE t SET v='main working'", "SELECT dolt_checkout('other')", "UPDATE t SET v='other working'",
	}, Assertions: []ScriptTestAssertion{
		{Query: `SELECT a.v,b.v FROM "postgres/main".public.t a JOIN "postgres/other".public.t b USING(id)`, Expected: []sql.Row{{"main working", "other working"}}},
		{Query: "SELECT v FROM t AS OF 'main'", Expected: []sql.Row{{"base"}}},
		{Query: "SELECT active_branch()", Expected: []sql.Row{{"other"}}},
		{Query: `SELECT v FROM "postgres/main".public.t`, Expected: []sql.Row{{"main working"}}},
		{Query: "SELECT v FROM t", Expected: []sql.Row{{"other working"}}},
	}}})
}
