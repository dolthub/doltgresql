package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// Native metadata statements retain PostgreSQL validation. Their branch and
// revision behavior uses the same harness as the version-control scenarios.
func TestBranchMetadataCatalog(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "native attribute proposals are transactional and branch local",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY,x int)",
			"INSERT INTO t VALUES(1,10)",
			"SELECT dolt_commit('-Am','metadata base')",
			"SELECT dolt_checkout('-b','other')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "ALTER TABLE t ALTER COLUMN x SET STATISTICS 200"},
			{Query: "SELECT attstattarget FROM pg_attribute WHERE attrelid='t'::regclass AND attname='x'", Expected: []sql.Row{{200}}},
			{Query: "ALTER TABLE t ALTER COLUMN x SET STATISTICS 300, ALTER COLUMN x SET (invalid_option=1)", ExpectedErr: "unrecognized parameter"},
			{Query: "SELECT attstattarget FROM pg_attribute WHERE attrelid='t'::regclass AND attname='x'", Expected: []sql.Row{{200}}},
			{Query: "ALTER TABLE t ALTER COLUMN x SET (n_distinct=-0.5)"},
			{Query: "SELECT attoptions::text FROM pg_attribute WHERE attrelid='t'::regclass AND attname='x'", Expected: []sql.Row{{"{n_distinct=-0.5}"}}},
			{Query: "SELECT attstattarget IS NULL FROM pg_attribute AS OF 'HEAD' WHERE attname='x'", Expected: []sql.Row{{"t"}}},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "SELECT attstattarget IS NULL,attoptions IS NULL FROM pg_attribute WHERE attrelid='t'::regclass AND attname='x'", Expected: []sql.Row{{"t", "t"}}},
			{Query: "SELECT x FROM t", Expected: []sql.Row{{10}}},
		},
	}})
}
