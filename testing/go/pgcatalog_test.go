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
					Query:    `SELECT * FROM "pg_catalog"."pg_type";`,
					Expected: []sql.Row{},
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
					Expected: []sql.Row{},
				},
			},
		},
	})
}
