package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// Assert logical catalog relationships, without depending on engine-local OIDs,
// access method names, system-object inventories, or deparser formatting.
func TestBranchCatalogIntrospection(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "stored catalog views follow branch schema and membership",
		SetUpScript: []string{
			"CREATE TABLE t(id int PRIMARY KEY, value text)",
			"INSERT INTO t VALUES(1,'one')",
			"SELECT dolt_commit('-Am','base')",
			"SELECT dolt_checkout('-b','other')",
			"ALTER TABLE t ADD COLUMN extra int",
			"CREATE TABLE only_other(x int)",
			"SELECT dolt_commit('-Am','other schema')",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "WITH t AS (SELECT 42 AS id) SELECT id FROM t", Expected: []sql.Row{{42}}},
			{Query: "WITH cols AS (SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='t') SELECT column_name FROM cols ORDER BY column_name", Expected: []sql.Row{{"extra"}, {"id"}, {"value"}}},
			{Query: "SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='t' ORDER BY ordinal_position", Expected: []sql.Row{{"id"}, {"value"}, {"extra"}}},
			{Query: "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename", Expected: []sql.Row{{"only_other"}, {"t"}}},
			{Query: "SELECT indexname FROM pg_indexes WHERE schemaname='public' AND tablename='t'", Expected: []sql.Row{{"t_pkey"}}},
			{Query: "SELECT constraint_type FROM information_schema.table_constraints WHERE table_schema='public' AND table_name='t' AND constraint_type='PRIMARY KEY'", Expected: []sql.Row{{"PRIMARY KEY"}}},
			{Query: "SELECT column_name,ordinal_position FROM information_schema.key_column_usage WHERE table_schema='public' AND table_name='t' ORDER BY ordinal_position", Expected: []sql.Row{{"id", 1}}},
			{Query: "SELECT c.relname,a.attname FROM pg_class c JOIN pg_attribute a ON a.attrelid=c.oid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relname='t' AND a.attnum>0 ORDER BY a.attnum", Expected: []sql.Row{{"t", "id"}, {"t", "value"}, {"t", "extra"}}},
			{Query: "SELECT dolt_checkout('main')", SkipResultsCheck: true},
			{Query: "SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='t' ORDER BY ordinal_position", Expected: []sql.Row{{"id"}, {"value"}}},
			{Query: "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename", Expected: []sql.Row{{"t"}}},
			{Query: "SELECT table_name FROM information_schema.tables WHERE table_schema='public' ORDER BY table_name", Expected: []sql.Row{{"t"}}},
			{Query: "SELECT dolt_checkout('other')", SkipResultsCheck: true},
			{Query: "ALTER TABLE t DROP COLUMN value"},
			{Query: "SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='t' ORDER BY ordinal_position", Expected: []sql.Row{{"id"}, {"extra"}}},
			{Query: "SELECT indexname FROM pg_indexes WHERE schemaname='public' AND tablename='t'", Expected: []sql.Row{{"t_pkey"}}},
		},
	}})
}
