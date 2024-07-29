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
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE test (pk INT primary key, v1 TEXT DEFAULT 'hey');`,

				// Should show attributes for all schemas
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_attribute" WHERE attname='pk';`,
					Expected: []sql.Row{{2686451712, "pk", 23, 0, 1, -1, -1, 0, "f", "i", "p", "", "t", "f", "f", "", "", "f", "t", 0, -1, 0, nil, nil, nil, nil}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_attribute" WHERE attname='v1';`,
					Expected: []sql.Row{{2686451712, "v1", 25, 0, 2, -1, -1, 0, "f", "i", "p", "", "f", "t", "f", "", "", "f", "t", 0, -1, 0, nil, nil, nil, nil}},
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
					Query: "SELECT attname FROM PG_catalog.pg_ATTRIBUTE ORDER BY attname LIMIT 3;",
					Expected: []sql.Row{
						{"abbrev"},
						{"abbrev"},
						{"active"},
					},
				},
				{
					Query: `SELECT attname FROM "pg_catalog"."pg_attribute" a JOIN "pg_catalog"."pg_class" c ON a.attrelid = c.oid WHERE c.relname = 'test';`,
					Expected: []sql.Row{
						{"pk"},
						{"v1"},
					},
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

func TestPgAvailableExtensionVersions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_available_extension_versions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_available_extension_versions";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_available_extension_versions";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_available_extension_versions";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_AVAILABLE_EXTENSION_VERSIONS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAvailableExtensions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_available_extensions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_available_extensions";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_available_extensions";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_available_extensions";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_AVAILABLE_EXTENSIONS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgBackendMemoryContexts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_backend_memory_contexts",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_backend_memory_contexts";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_backend_memory_contexts";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_backend_memory_contexts";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_BACKEND_MEMORY_CONTEXTS ORDER BY name;",
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
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE VIEW testview AS SELECT * FROM testing LIMIT 1;`,

				// Should show classes for all schemas
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				// Table
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE relname='testing';`,
					Expected: []sql.Row{
						{2686451712, "testing", 1879048194, 0, 0, 0, 0, 0, 0, 0, float32(0), 0, 0, "t", "f", "p", "r", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				// Index
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE relname='testing_pkey';`,
					Expected: []sql.Row{
						{1612709888, "testing_pkey", 1879048194, 0, 0, 0, 0, 0, 0, 0, float32(0), 0, 0, "f", "f", "p", "i", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				// View
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE relname='testview';`,
					Expected: []sql.Row{
						{2954887168, "testview", 1879048194, 0, 0, 0, 0, 0, 0, 0, float32(0), 0, 0, "f", "f", "p", "v", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
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
					Query: "SELECT relname FROM PG_catalog.pg_CLASS ORDER BY relname ASC LIMIT 3;",
					Expected: []sql.Row{
						{"pg_aggregate"},
						{"pg_am"},
						{"pg_amop"},
					},
				},
				{
					Query: "SELECT relname from pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid  WHERE n.nspname = 'testschema' ORDER BY relname;",
					Expected: []sql.Row{
						{"testing"},
						{"testing_pkey"},
						{"testview"},
						{"v1"},
					},
				},
				{
					Query: "SELECT relname from pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid  WHERE n.nspname = 'pg_catalog' LIMIT 3;",
					Expected: []sql.Row{
						{"pg_aggregate"},
						{"pg_am"},
						{"pg_amop"},
					},
				},
				{
					Skip:  true, // TODO: Should be able to select from pg_class without specifying pg_catalog
					Query: `SELECT relname FROM "pg_class" WHERE relname='testing';`,
					Expected: []sql.Row{
						{"testing"},
					},
				},
			},
		},
		{
			Name: "pg_class with regclass",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE VIEW testview AS SELECT * FROM testing LIMIT 1;`,
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT * FROM "pg_catalog"."pg_class" WHERE oid='testing'::regclass;`,
					ExpectedErr: "does not exist",
				},
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE oid='testschema.testing'::regclass;`,
					Expected: []sql.Row{
						{2686451712, "testing", 1879048194, 0, 0, 0, 0, 0, 0, 0, float32(0), 0, 0, "t", "f", "p", "r", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE oid='testschema.testing_pkey'::regclass;`,
					Expected: []sql.Row{
						{1612709888, "testing_pkey", 1879048194, 0, 0, 0, 0, 0, 0, 0, float32(0), 0, 0, "f", "f", "p", "i", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE oid='testschema.testview'::regclass;`,
					Expected: []sql.Row{
						{2954887168, "testview", 1879048194, 0, 0, 0, 0, 0, 0, 0, float32(0), 0, 0, "f", "f", "p", "v", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
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

func TestPgConfig(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_config",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_config";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_config";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_config";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_CONFIG ORDER BY name;",
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
			SetUpScript: []string{
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE TABLE testing2 (pk INT primary key, pktesting INT REFERENCES testing(pk), v1 TEXT);`,
				// TODO: Uncomment when check constraints supported
				// `ALTER TABLE testing2 ADD CONSTRAINT v1_check CHECK (v1 != '')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM "pg_catalog"."pg_constraint" LIMIT 3;`,
					Expected: []sql.Row{
						{1611661312, "testing_pkey", 1879048193, "p", "f", "f", "t", 2685403136, 0, 1611661312, 0, 0, "", "", "", "t", 0, "t", "{1}", nil, nil, nil, nil, nil, nil, nil},
						{1611661313, "v1", 1879048193, "u", "f", "f", "t", 2685403136, 0, 1611661313, 0, 0, "", "", "", "t", 0, "t", "{2}", nil, nil, nil, nil, nil, nil, nil},
						{1611661314, "testing2_pkey", 1879048193, "p", "f", "f", "t", 2685403137, 0, 1611661314, 0, 0, "", "", "", "t", 0, "t", "{1}", nil, nil, nil, nil, nil, nil, nil},
					},
				},
				{
					Skip:  true, // TODO: Foreign keys don't work
					Query: `SELECT * FROM "pg_catalog"."pg_constraint" LIMIT 2;`,
					Expected: []sql.Row{
						{1611661312, "testing_pkey", 1879048193, "p", "f", "f", "t", 2685403136, 0, 1611661312, 0, 0, "", "", "", "t", 0, "t", "{1}", nil, nil, nil, nil, nil, nil, nil},
						{1611661313, "v1", 1879048193, "u", "f", "f", "t", 2685403136, 0, 1611661313, 0, 0, "", "", "", "t", 0, "t", "{2}", nil, nil, nil, nil, nil, nil, nil},
						{1611661314, "testing2_pkey", 1879048193, "p", "f", "f", "t", 2685403137, 0, 1611661314, 0, 0, "", "", "", "t", 0, "t", "{1}", nil, nil, nil, nil, nil, nil, nil},
						{1611661314, "testing2_pktesting_fkey", 1879048193, "f", "f", "t", 2685403137, 0, 1611661314, 0, 0, "", "", "", "t", 0, "t", "{2}", "{1}", nil, nil, nil, nil, nil, nil}},
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
					Query: "SELECT conname FROM PG_catalog.pg_CONSTRAINT ORDER BY conname;",
					Expected: []sql.Row{
						{"testing2_pkey"},
						{"testing_pkey"},
						{"v1"},
					},
				},
				{
					Query: "SELECT co.oid, co.conname, co.conrelid, cl.relname FROM pg_catalog.pg_constraint co JOIN pg_catalog.pg_class cl ON co.conrelid = cl.oid WHERE cl.relname = 'testing2';",
					Expected: []sql.Row{
						{1611661314, "testing2_pkey", 2685403137, "testing2"},
					},
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

func TestPgCursors(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_cursors",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_cursors";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_cursors";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_cursors";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_CURSORS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

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
					Query: `SELECT oid, datname FROM "pg_catalog"."pg_database" ORDER BY datname DESC;`,
					Expected: []sql.Row{
						{805306370, "test"},
						{805306369, "postgres"},
						{805306368, "doltgres"},
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
					Query: "SELECT oid, datname FROM PG_catalog.pg_DATABASE ORDER BY datname ASC;",
					Expected: []sql.Row{
						{805306368, "doltgres"},
						{805306369, "postgres"},
						{805306370, "test"},
					},
				},
				{
					Query: "SELECT * FROM pg_catalog.pg_database WHERE datname='test';",
					Expected: []sql.Row{
						{805306370, "test", 0, 0, "i", "f", "t", -1, 0, 0, 0, "", "", nil, "", nil, nil},
					},
				},
			},
		},
	})
}

func TestPgDbRoleSetting(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_db_role_setting",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_db_role_setting";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_db_role_setting";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_db_role_setting";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT setdatabase FROM PG_catalog.pg_DB_ROLE_SETTING ORDER BY setdatabase;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgDefaultAcl(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_default_acl",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_default_acl";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_default_acl";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_default_acl";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_DEFAULT_ACL ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgDepend(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_depend",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_depend";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_depend";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_depend";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT classid FROM PG_catalog.pg_DEPEND ORDER BY classid;",
					Expected: []sql.Row{},
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

func TestPgExtension(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_extension",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_extension";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_extension";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_extension";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT extname FROM PG_catalog.pg_EXTENSION ORDER BY extname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgFileSettings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_file_settings",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_file_settings";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_file_settings";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_file_settings";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_FILE_SETTINGS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgForeignDataWrapper(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_foreign_data_wrapper",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_foreign_data_wrapper";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_foreign_data_wrapper";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_foreign_data_wrapper";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT fdwname FROM PG_catalog.pg_FOREIGN_DATA_WRAPPER ORDER BY fdwname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgForeignServer(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_foreign_server",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_foreign_server";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_foreign_server";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_foreign_server";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT srvname FROM PG_catalog.pg_FOREIGN_SERVER ORDER BY srvname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgForeignTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_foreign_table",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_foreign_table";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_foreign_table";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_foreign_table";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT ftrelid FROM PG_catalog.pg_FOREIGN_TABLE ORDER BY ftrelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgGroup(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_group",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_group";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_group";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_group";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT groname FROM PG_catalog.pg_GROUP ORDER BY groname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgHbaFileRules(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_hba_file_rules",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_hba_file_rules";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_hba_file_rules";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_hba_file_rules";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT line_number FROM PG_catalog.pg_HBA_FILE_RULES ORDER BY line_number;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgIdentFileMappings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ident_file_mappings",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ident_file_mappings";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ident_file_mappings";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ident_file_mappings";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT line_number FROM PG_catalog.pg_IDENT_FILE_MAPPINGS ORDER BY line_number;",
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
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE TABLE testing2 (pk INT, v1 INT, PRIMARY KEY (pk, v1));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM "pg_catalog"."pg_index";`,
					Expected: []sql.Row{
						{1612709888, 2686451712, 1, 0, "t", "f", "t", "f", "f", "f", "t", "f", "t", "t", "f", "{}", "{}", "{}", "{}", nil, nil},
						{1612709889, 2686451712, 1, 0, "t", "f", "f", "f", "f", "f", "t", "f", "t", "t", "f", "{}", "{}", "{}", "{}", nil, nil},
						{1612709890, 2686451713, 2, 0, "t", "f", "t", "f", "f", "f", "t", "f", "t", "t", "f", "{}", "{}", "{}", "{}", nil, nil},
					},
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
					Query:    "SELECT indexrelid FROM PG_catalog.pg_INDEX ORDER BY indexrelid ASC;",
					Expected: []sql.Row{{1612709888}, {1612709889}, {1612709890}},
				},
				{
					Query: "SELECT i.indexrelid, i.indrelid, c.relname, t.relname  FROM pg_catalog.pg_index i JOIN pg_catalog.pg_class c ON i.indexrelid = c.oid JOIN pg_catalog.pg_class t ON i.indrelid = t.oid;",
					Expected: []sql.Row{
						{1612709888, 2686451712, "testing_pkey", "testing"},
						{1612709889, 2686451712, "v1", "testing"},
						{1612709890, 2686451713, "testing2_pkey", "testing2"},
					},
				},
			},
		},
	})
}

func TestPgIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_indexes",
			SetUpScript: []string{
				"CREATE SCHEMA testschema;",
				"SET search_path TO testschema;",
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE TABLE testing2 (pk INT, v1 INT, PRIMARY KEY (pk, v1));`,
				"CREATE INDEX my_index ON testing2(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM "pg_catalog"."pg_indexes";`,
					Expected: []sql.Row{
						{"testschema", "testing", "testing_pkey", "", "CREATE UNIQUE INDEX testing_pkey ON testschema.testing USING btree (pk)"},
						{"testschema", "testing", "v1", "", "CREATE UNIQUE INDEX v1 ON testschema.testing USING btree (v1)"},
						{"testschema", "testing2", "testing2_pkey", "", "CREATE UNIQUE INDEX testing2_pkey ON testschema.testing2 USING btree (pk, v1)"},
						{"testschema", "testing2", "my_index", "", "CREATE INDEX my_index ON testschema.testing2 USING btree (v1)"},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT indexname FROM PG_catalog.pg_INDEXES ORDER BY indexname;",
					Expected: []sql.Row{{"my_index"}, {"testing2_pkey"}, {"testing_pkey"}, {"v1"}},
				},
			},
		},
	})
}

func TestPgInherits(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_inherits",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_inherits";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_inherits";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_inherits";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT inhrelid FROM PG_catalog.pg_INHERITS ORDER BY inhrelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgInitPrivs(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_init_privs",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_init_privs";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_init_privs";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_init_privs";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_INIT_PRIVS ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgLanguage(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_language",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_language";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_language";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_language";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT lanname FROM PG_catalog.pg_LANGUAGE ORDER BY lanname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgLargeobject(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_largeobject",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_largeobject";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_largeobject";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_largeobject";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT loid FROM PG_catalog.pg_LARGEOBJECT ORDER BY loid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgLargeobjectMetadata(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_largeobject_metadata",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_largeobject_metadata";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_largeobject_metadata";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_largeobject_metadata";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_LARGEOBJECT_METADATA ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgLocks(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_locks",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_locks";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_locks";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_locks";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objid FROM PG_catalog.pg_LOCKS ORDER BY objid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgMatviews(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_matviews",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_matviews";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_matviews";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_matviews";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT matviewname FROM PG_catalog.pg_MATVIEWS ORDER BY matviewname;",
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
					Query: `SELECT * FROM "pg_catalog"."pg_namespace" ORDER BY nspname;`,
					Expected: []sql.Row{
						{1879048194, "information_schema", 0, nil},
						{1879048192, "pg_catalog", 0, nil},
						{1879048193, "public", 0, nil},
					},
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
					Query: "SELECT nspname FROM PG_catalog.pg_NAMESPACE ORDER BY nspname;",
					Expected: []sql.Row{
						{"information_schema"},
						{"pg_catalog"},
						{"public"},
					},
				},
				{
					Query:    "CREATE SCHEMA testschema;",
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM "pg_catalog"."pg_namespace" ORDER BY nspname;`,
					Expected: []sql.Row{
						{1879048195, "information_schema", 0, nil},
						{1879048192, "pg_catalog", 0, nil},
						{1879048193, "public", 0, nil},
						{1879048194, "testschema", 0, nil},
					},
				},
			},
		},
	})
}

func TestPgOpclass(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_opclass",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_opclass";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_opclass";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_opclass";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT opcname FROM PG_catalog.pg_OPCLASS ORDER BY opcname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgOpfamily(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_opfamily",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_opfamily";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_opfamily";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_opfamily";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT opfname FROM PG_catalog.pg_OPFAMILY ORDER BY opfname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgParameterAcl(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_parameter_acl",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_parameter_acl";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_parameter_acl";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_parameter_acl";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT parname FROM PG_catalog.pg_PARAMETER_ACL ORDER BY parname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPartitionedTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_partitioned_table",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_partitioned_table";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_partitioned_table";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_partitioned_table";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT partrelid FROM PG_catalog.pg_PARTITIONED_TABLE ORDER BY partrelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPolicies(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_policies",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_policies";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_policies";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_policies";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT policyname FROM PG_catalog.pg_POLICIES ORDER BY policyname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPolicy(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_policy",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_policy";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_policy";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_policy";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT polname FROM PG_catalog.pg_POLICY ORDER BY polname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPreparedStatements(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_prepared_statements",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_prepared_statements";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_prepared_statements";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_prepared_statements";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_PREPARED_STATEMENTS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPreparedXacts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_prepared_xacts",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_prepared_xacts";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_prepared_xacts";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_prepared_xacts";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT gid FROM PG_catalog.pg_PREPARED_XACTS ORDER BY gid;",
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

func TestPgPublication(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_publication",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_publication";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_publication";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_publication";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pubname FROM PG_catalog.pg_PUBLICATION ORDER BY pubname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPublicationNamespace(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_publication_namespace",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_publication_namespace";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_publication_namespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_publication_namespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_PUBLICATION_NAMESPACE ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPublicationRel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_publication_rel",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_publication_rel";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_publication_rel";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_publication_rel";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_PUBLICATION_REL ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPublicationTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_publication_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_publication_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_publication_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_publication_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pubname FROM PG_catalog.pg_PUBLICATION_TABLES ORDER BY pubname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgRange(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_range",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_range";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_range";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_range";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT rngtypid FROM PG_catalog.pg_RANGE ORDER BY rngtypid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgReplicationOrigin(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_replication_origin",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_replication_origin";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_replication_origin";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_replication_origin";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT roname FROM PG_catalog.pg_REPLICATION_ORIGIN ORDER BY roname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgReplicationOriginStatus(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_replication_origin_status",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_replication_origin_status";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_replication_origin_status";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_replication_origin_status";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT local_id FROM PG_catalog.pg_REPLICATION_ORIGIN_STATUS ORDER BY local_id;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgReplicationSlots(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_replication_slot",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_replication_slots";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_replication_slots";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_replication_slots";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT slot_name FROM PG_catalog.pg_REPLICATION_SLOTS ORDER BY slot_name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgRewrite(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_rewrite",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_rewrite";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_rewrite";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_rewrite";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_REWRITE ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgRoles(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_roles",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_roles";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_roles";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_roles";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT rolname FROM PG_catalog.pg_ROLES ORDER BY rolname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgRules(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_rules",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_rules";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_rules";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_rules";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT rulename FROM PG_catalog.pg_RULES ORDER BY rulename;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSeclabel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_seclabel",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_seclabel";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_seclabel";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_seclabel";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_SECLABEL ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSeclabels(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_seclabels",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_seclabels";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_seclabels";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_seclabels";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_SECLABELS ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_sequences",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_sequences";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT sequencename FROM PG_catalog.pg_SEQUENCES ORDER BY sequencename;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSettings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_settings",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_settings";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_settings";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_settings";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_SETTINGS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShadow(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shadow",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shadow";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shadow";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shadow";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT usename FROM PG_catalog.pg_SHADOW ORDER BY usename;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShdepend(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shdepend",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shdepend";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shdepend";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shdepend";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT dbid FROM PG_catalog.pg_SHDEPEND ORDER BY dbid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShdescription(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shdescription",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shdescription";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shdescription";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shdescription";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_SHDESCRIPTION ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShmemAllocations(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shmem_allocations",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shmem_allocations";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shmem_allocations";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shmem_allocations";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_SHMEM_ALLOCATIONS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShseclabel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shseclabel",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shseclabel";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shseclabel";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shseclabel";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_SHSECLABEL ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatActivity(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_activity",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_activity";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_activity";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_activity";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT datname FROM PG_catalog.pg_STAT_ACTIVITY ORDER BY datname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatAllIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_all_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_all_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_all_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_all_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relname FROM PG_catalog.pg_STAT_ALL_INDEXES ORDER BY relname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatAllTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_all_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_all_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relname FROM PG_catalog.pg_STAT_ALL_TABLES ORDER BY relname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatArchiver(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_archiver",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_archiver";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_archiver";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_archiver";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT archived_count FROM PG_catalog.pg_STAT_ARCHIVER ORDER BY archived_count;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatBgwriter(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_bgwriter",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_bgwriter";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_bgwriter";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_bgwriter";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT checkpoints_timed FROM PG_catalog.pg_STAT_BGWRITER ORDER BY checkpoints_timed;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatDatabase(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_database",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_database";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT datname FROM PG_catalog.pg_STAT_DATABASE ORDER BY datname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatDatabaseConflicts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_database_conflicts",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_database_conflicts";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_database_conflicts";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_database_conflicts";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT datname FROM PG_catalog.pg_STAT_DATABASE_CONFLICTS ORDER BY datname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatGssapi(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_gssapi",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_gssapi";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_gssapi";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_gssapi";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_GSSAPI ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatIo(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_io",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_io";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_io";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_io";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT object FROM PG_catalog.pg_STAT_IO ORDER BY object;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressAnalyze(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_analyze",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_analyze";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_analyze";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_analyze";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT datname FROM PG_catalog.pg_STAT_PROGRESS_ANALYZE ORDER BY datname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressBasebackup(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_basebackup",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_basebackup";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_basebackup";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_basebackup";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_BASEBACKUP ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressCluster(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_cluster",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_cluster";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_cluster";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_cluster";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_CLUSTER ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressCopy(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_copy",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_copy";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_copy";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_copy";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_COPY ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressCreateIndex(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_create_index",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_create_index";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_create_index";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_create_index";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_CREATE_INDEX ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressVacuum(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_vacuum",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_vacuum";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_vacuum";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_vacuum";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_VACUUM ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatRecoveryPrefetch(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_recovery_prefetch",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_recovery_prefetch";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_recovery_prefetch";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_recovery_prefetch";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT hit FROM PG_catalog.pg_STAT_RECOVERY_PREFETCH ORDER BY hit;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatReplication(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_replication",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_replication";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_replication";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_replication";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_REPLICATION ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatReplicationSlots(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_replication_slots",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_replication_slots";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_replication_slots";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_replication_slots";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT slot_name FROM PG_catalog.pg_STAT_REPLICATION_SLOTS ORDER BY slot_name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSlru(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_slru",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_slru";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_slru";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_slru";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_STAT_SLRU ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSsl(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_ssl",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_ssl";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_ssl";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_ssl";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_SSL ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSubscription(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_subscription",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_subscription";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_subscription";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_subscription";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT subid FROM PG_catalog.pg_STAT_SUBSCRIPTION ORDER BY subid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSubscriptionStats(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_subscription_stats",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_subscription_stats";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_subscription_stats";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_subscription_stats";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT subid FROM PG_catalog.pg_STAT_SUBSCRIPTION_STATS ORDER BY subid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSysIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_sys_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_sys_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_sys_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_sys_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_SYS_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSysTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_sys_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_sys_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_SYS_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatUserFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_user_functions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_user_functions";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_user_functions";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_user_functions";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT funcid FROM PG_catalog.pg_STAT_USER_FUNCTIONS ORDER BY funcid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatUserIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_user_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_user_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_user_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_user_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_USER_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatUserTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_user_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_user_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_USER_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatWal(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_wal",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_wal";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_wal";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_wal";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT wal_records FROM PG_catalog.pg_STAT_WAL ORDER BY wal_records;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatWalReceiver(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_wal_receiver",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_wal_receiver";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_wal_receiver";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_wal_receiver";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_WAL_RECEIVER ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatXactAllTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_xact_all_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_xact_all_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_xact_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_xact_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_XACT_ALL_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatXactSysTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_xact_sys_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_xact_sys_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_xact_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_xact_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_XACT_SYS_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatXactUserFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_xact_user_functions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_xact_user_functions";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_xact_user_functions";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_xact_user_functions";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT funcid FROM PG_catalog.pg_STAT_XACT_USER_FUNCTIONS ORDER BY funcid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatXactUserTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_xact_user_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_xact_user_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_xact_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_xact_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_XACT_USER_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioAllIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_all_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_all_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_all_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_all_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_ALL_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioAllSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_all_sequences",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_all_sequences";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_all_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_all_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_ALL_SEQUENCES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioAllTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_all_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_all_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_ALL_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioSysIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_sys_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_sys_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_sys_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_sys_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_SYS_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioSysSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_sys_sequences",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_sys_sequences";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_sys_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_sys_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_SYS_SEQUENCES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioSysTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_sys_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_sys_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_SYS_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioUserIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_user_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_user_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_user_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_user_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_USER_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioUserSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_user_sequences",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_user_sequences";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_user_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_user_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_USER_SEQUENCES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioUserTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_user_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_user_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_USER_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatistic(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statistic",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statistic";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statistic";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statistic";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT starelid FROM PG_catalog.pg_STATISTIC ORDER BY starelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatisticExt(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statistic_ext",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statistic_ext";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statistic_ext";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statistic_ext";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT stxname FROM PG_catalog.pg_STATISTIC_EXT ORDER BY stxname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatisticExtData(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statistic_ext_data",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statistic_ext_data";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statistic_ext_data";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statistic_ext_data";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT stxoid FROM PG_catalog.pg_STATISTIC_EXT_DATA ORDER BY stxoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStats(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stats",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stats";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stats";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stats";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT attname FROM PG_catalog.pg_STATS ORDER BY attname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatsExt(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stats_ext",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stats_ext";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stats_ext";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stats_ext";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT statistics_name FROM PG_catalog.pg_STATS_EXT ORDER BY statistics_name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatsExtExprs(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stats_ext_exprs",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stats_ext_exprs";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stats_ext_exprs";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stats_ext_exprs";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT statistics_name FROM PG_catalog.pg_STATS_EXT_EXPRS ORDER BY statistics_name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSubscription(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_subscription",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_subscription";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_subscription";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_subscription";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT subname FROM PG_catalog.pg_SUBSCRIPTION ORDER BY subname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSubscriptionRel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_subscription_rel",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_subscription_rel";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_subscription_rel";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_subscription_rel";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT srsubid FROM PG_catalog.pg_SUBSCRIPTION_REL ORDER BY srsubid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_tables",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE testing (pk INT primary key, v1 INT);`,

				// Should show classes for all schemas
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_tables" WHERE tablename='testing';`,
					Expected: []sql.Row{{"testschema", "testing", "", "", "t", "f", "f", "f"}},
				},
				{
					Query:    `SELECT count(*) FROM "pg_catalog"."pg_tables" WHERE schemaname='pg_catalog';`,
					Expected: []sql.Row{{139}},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT schemaname, tablename FROM PG_catalog.pg_TABLES ORDER BY tablename DESC LIMIT 3;",
					Expected: []sql.Row{
						{"testschema", "testing"},
						{"pg_catalog", "pg_views"},
						{"pg_catalog", "pg_user_mappings"},
					},
				},
			},
		},
	})
}

func TestPgTablespace(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_tablespace",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_tablespace";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_tablespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_tablespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT spcname FROM PG_catalog.pg_TABLESPACE ORDER BY spcname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTimezoneAbbrevs(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_timezone_abbrevs",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_timezone_abbrevs";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_timezone_abbrevs";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_timezone_abbrevs";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT abbrev FROM PG_catalog.pg_TIMEZONE_ABBREVS ORDER BY abbrev;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPgTimezoneNames(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_timezone_names",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_timezone_names";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_timezone_names";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_timezone_names";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_TIMEZONE_NAMES ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTransform(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_transform",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_transform";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_transform";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_transform";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_TRANSFORM ORDER BY oid;",
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

func TestPgTsConfig(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_config",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_config";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_config";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_config";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT cfgname FROM PG_catalog.pg_TS_CONFIG ORDER BY cfgname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsConfigMap(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_config_map",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_config_map";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_config_map";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_config_map";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT mapcfg FROM PG_catalog.pg_TS_CONFIG_MAP ORDER BY mapcfg;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsDict(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_dict",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_dict";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_dict";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_dict";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT dictname FROM PG_catalog.pg_TS_DICT ORDER BY dictname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsParser(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_parser",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_parser";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_parser";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_parser";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT prsname FROM PG_catalog.pg_TS_PARSER ORDER BY prsname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsTemplate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_template",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_template";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_template";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_template";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT tmplname FROM PG_catalog.pg_TS_TEMPLATE ORDER BY tmplname;",
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
					Query:    "SELECT typname FROM PG_catalog.pg_TYPE WHERE typname LIKE '%char' ORDER BY typname;",
					Expected: []sql.Row{{"bpchar"}, {"char"}, {"varchar"}},
				},
			},
		},
		{
			Name: "pg_type with regtype",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE oid='float8'::regtype;`,
					Expected: []sql.Row{{701, "float8", 0, 0, 8, "t", "b", "N", "t", "t", ",", 0, "-", 0, 0, "float8in", "float8out", "float8rec", "float8send", "-", "-", "-", "d", "x", "f", 0, 0, 0, 0, nil, nil, nil}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='double precision'::regtype;`,
					Expected: []sql.Row{{701, "float8"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='DOUBLE PRECISION'::regtype;`,
					Expected: []sql.Row{{701, "float8"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='pg_catalog.float8'::regtype;`,
					Expected: []sql.Row{{701, "float8"}},
				},
				{
					Query:       `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='public.float8'::regtype;`,
					ExpectedErr: `type "public.float8" does not exist`,
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='VARCHAR'::regtype;`,
					Expected: []sql.Row{{1043, "varchar"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='1043'::regtype;`,
					Expected: []sql.Row{{1043, "varchar"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='VARCHAR(10)'::regtype;`,
					Expected: []sql.Row{{1043, "varchar"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='character varying'::regtype;`,
					Expected: []sql.Row{{1043, "varchar"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='timestamptz'::regtype;`,
					Expected: []sql.Row{{1184, "timestamptz"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='timestamp with time zone'::regtype;`,
					Expected: []sql.Row{{1184, "timestamptz"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='regtype'::regtype;`,
					Expected: []sql.Row{{2206, "regtype"}},
				},
			},
		},
	})
}

func TestPgUser(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_user",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_user";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_user";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_user";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT usename FROM PG_catalog.pg_USER ORDER BY usename;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgUserMapping(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_user_mapping",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_user_mapping";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_user_mapping";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_user_mapping";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT umuser FROM PG_catalog.pg_USER_MAPPING ORDER BY umuser;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgUserMappings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_user_mappings",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_user_mappings";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_user_mappings";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_user_mappings";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT umid FROM PG_catalog.pg_USER_MAPPINGS ORDER BY umid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgViews(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_views",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				"CREATE TABLE testing (pk INT primary key, v1 INT);",
				`CREATE VIEW testview AS SELECT * FROM testing LIMIT 1;`,
				`CREATE VIEW testview2 AS SELECT * FROM testing LIMIT 2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_views" WHERE viewname='testview';`,
					Expected: []sql.Row{{"testschema", "testview", "", "SELECT * FROM testing LIMIT 1"}},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_views";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_views";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT viewname FROM PG_catalog.pg_VIEWS ORDER BY viewname;",
					Expected: []sql.Row{{"testview"}, {"testview2"}},
				},
			},
		},
	})
}
