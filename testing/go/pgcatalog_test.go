package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

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
