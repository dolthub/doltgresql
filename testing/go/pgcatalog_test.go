package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestPgAggregate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_aggregate",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_aggregate";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_aggregate";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_aggregate";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT aggfnoid FROM PG_catalog.pg_AGGREGATE ORDER BY aggfnoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAm(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_am",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_am";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_am";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_am";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT amname FROM PG_catalog.pg_AM ORDER BY amname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAmop(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_amop",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_amop";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_amop";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_amop";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_AMOP ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAmproc(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_amproc",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_amproc";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_amproc";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_amproc";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_AMPROC ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAttribute(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_attribute",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_attribute";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_attribute";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_attribute";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT attname FROM PG_catalog.pg_ATTRIBUTE ORDER BY attname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAttrdef(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_attrdef",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_attrdef";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_attrdef";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_attrdef";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_ATTRDEF ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAuthMembers(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_auth_members",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_auth_members";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_auth_members";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_auth_members";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT member FROM PG_catalog.pg_AUTH_MEMBERS ORDER BY member;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAuthid(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_authid",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_authid";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_authid";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_authid";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT rolname FROM PG_catalog.pg_AUTHID ORDER BY rolname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgCast(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_cast",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_cast";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_cast";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_cast";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_CAST ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgClass(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_class",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_class";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_class";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_class";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relname FROM PG_catalog.pg_CLASS ORDER BY relname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgCollation(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_collation",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_collation";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_collation";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_collation";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT collname FROM PG_catalog.pg_COLLATION ORDER BY collname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgConstraint(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_constraint",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_constraint";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_constraint";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_constraint";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT conname FROM PG_catalog.pg_CONSTRAINT ORDER BY conname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgConversion(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_conversion",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_conversion";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_conversion";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_conversion";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT conname FROM PG_catalog.pg_CONVERSION ORDER BY conname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

// TODO: Figure out why there is not a doltgres database when running these tests locally
func TestPgDatabase(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_database",
			SetUpScript: []string{
				`CREATE DATABASE test;`,
				`CREATE TABLE test (pk INT primary key, v1 INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT datname FROM "pg_catalog"."pg_database";`,
					Expected: []sql.Row{
						{"doltgres"},
						{"postgres"},
						{"test"},
					},
				},
				{
					Query: `SELECT datname FROM "pg_catalog"."pg_database" ORDER BY oid DESC;`,
					Expected: []sql.Row{
						{"test"},
						{"postgres"},
						{"doltgres"},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT datname FROM PG_catalog.pg_DATABASE ORDER BY datname;",
					Expected: []sql.Row{
						{"doltgres"},
						{"postgres"},
						{"test"},
					},
				},
				{
					Query: "SELECT * FROM pg_catalog.pg_database WHERE datname='test';",
					Expected: []sql.Row{
						{3, "test", 0, 0, "i", "f", "t", -1, 0, 0, 0, "", "", nil, "", nil, nil},
					},
				},
			},
		},
	})
}

func TestPgDescription(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_description",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_description";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_description";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_description";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_DESCRIPTION ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgEnum(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_enum",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_enum";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_enum";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_enum";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT enumlabel FROM PG_catalog.pg_ENUM ORDER BY enumlabel;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgEventTrigger(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_event_trigger",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_event_trigger";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_event_trigger";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_event_trigger";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT evtname FROM PG_catalog.pg_EVENT_TRIGGER ORDER BY evtname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgIndex(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_index",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_index";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_index";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_index";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT indexrelid FROM PG_catalog.pg_INDEX ORDER BY indexrelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgNamespace(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_namespace",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_namespace";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_namespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_namespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT nspname FROM PG_catalog.pg_NAMESPACE ORDER BY nspname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgProc(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_proc",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_proc";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_proc";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_proc";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT proname FROM PG_catalog.pg_PROC ORDER BY proname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTrigger(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_trigger",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_trigger";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_trigger";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_trigger";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT tgname FROM PG_catalog.pg_TRIGGER ORDER BY tgname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgType(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_type",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE typname = 'float8';`,
					Expected: []sql.Row{{701, "float8", 0, 0, 8, "t", "b", "N", "t", "t", ",", 0, "-", 0, 0, "float8in", "float8out", "float8rec", "float8send", "-", "-", "-", "d", "x", "f", 0, 0, 0, 0, nil, nil, nil}},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_type";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_type";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT typname FROM PG_catalog.pg_TYPE ORDER BY typname;",
					Expected: []sql.Row{{"anyarray"}, {"bool"}, {"bpchar"}, {"bytea"}, {"char"}, {"date"}, {"float4"}, {"float8"}, {"int2"}, {"int4"}, {"int8"}, {"json"}, {"jsonb"}, {"name"}, {"numeric"}, {"oid"}, {"text"}, {"time"}, {"timestamp"}, {"timestamptz"}, {"timetz"}, {"unknown"}, {"uuid"}, {"varChar"}, {"xid"}},
				},
			},
		},
	})
}
