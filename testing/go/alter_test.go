// Copyright 2025 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// TestAlterStatements tests ALTER statements other than ALTER TABLE, mostly for stub functionality.
// These tests should move into their respective test files as real functionality for these ALTER statements is added.
func TestAlterStatements(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "alter database",
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER DATABASE postgres OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "OWNER TO is unsupported and ignored",
						},
					},
				},
			},
		},
		{
			Name: "alter sequence",
			SetUpScript: []string{
				"CREATE SEQUENCE testseq",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER SEQUENCE testseq OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "OWNER TO is unsupported and ignored",
						},
					},
				},
			},
		},
		{
			Name: "alter type",
			SetUpScript: []string{
				"CREATE TYPE testtype AS ENUM ('a', 'b', 'c')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER TYPE testtype OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "OWNER TO is unsupported and ignored",
						},
					},
				},
			},
		},
		{
			Name: "alter function",
			SetUpScript: []string{
				"CREATE FUNCTION testfunc() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER FUNCTION testfunc() OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "OWNER TO is unsupported and ignored",
						},
					},
				},
			},
		},
		{
			Name: "alter procedure",
			SetUpScript: []string{
				"CREATE TABLE test (v1 INT8);",
				"CREATE PROCEDURE testproc() AS $$ BEGIN INSERT INTO test VALUES (1); END; $$ LANGUAGE plpgsql;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER PROCEDURE testproc() OWNER TO foo;",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "OWNER TO is unsupported and ignored",
						},
					},
				},
			},
		},
		{
			Name: "alter schema",
			SetUpScript: []string{
				"CREATE SCHEMA testschema",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER schema testschema OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "OWNER TO is unsupported and ignored",
						},
					},
				},
			},
		},
		{
			Name: "alter view",
			SetUpScript: []string{
				"CREATE VIEW testview AS SELECT 1",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER VIEW testview OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "OWNER TO is unsupported and ignored",
						},
					},
				},
			},
		},
	})
}

// TestAlterTableAddEnumColumnDefault covers the panic reported in issue #3530.
// ADD COLUMN must resolve the default's output type as well as the column's type
// before casting and validating a string-literal default.
func TestAlterTableAddEnumColumnDefault(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "enum default controls",
			SetUpScript: []string{
				`CREATE TYPE enum_default_probe AS ENUM ('x');`,
				`CREATE TABLE enum_default_no_default_control (id integer);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE enum_default_create_control (value enum_default_probe DEFAULT 'x');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO enum_default_create_control DEFAULT VALUES;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT value FROM enum_default_create_control;`,
					Expected: []sql.Row{{"x"}},
				},
				{
					Query:    `ALTER TABLE enum_default_no_default_control ADD COLUMN value enum_default_probe;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "add enum column with string literal default",
			SetUpScript: []string{
				`CREATE TYPE enum_default_probe AS ENUM ('x');`,
				`CREATE TABLE enum_default_probe_table (id integer);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// The minimal reproduction panics during default validation, even on an empty table.
					Query:    `ALTER TABLE enum_default_probe_table ADD COLUMN value enum_default_probe DEFAULT 'x';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO enum_default_probe_table (id) VALUES (1);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, value FROM enum_default_probe_table;`,
					Expected: []sql.Row{{1, "x"}},
				},
			},
		},
		{
			Name: "add schema qualified enum column with not null default",
			SetUpScript: []string{
				`CREATE SCHEMA auth;`,
				`CREATE TYPE auth.oauth_client_type AS ENUM ('confidential', 'public');`,
				`CREATE TABLE auth.oauth_clients (id integer);`,
				`INSERT INTO auth.oauth_clients VALUES (1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// The Supabase migration must also apply the default to existing rows.
					Query:    `ALTER TABLE auth.oauth_clients ADD COLUMN IF NOT EXISTS client_type auth.oauth_client_type NOT NULL DEFAULT 'confidential';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, client_type FROM auth.oauth_clients;`,
					Expected: []sql.Row{{1, "confidential"}},
				},
				{
					Query:    `INSERT INTO auth.oauth_clients (id) VALUES (2);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id, client_type FROM auth.oauth_clients ORDER BY id;`,
					Expected: []sql.Row{{1, "confidential"}, {2, "confidential"}},
				},
			},
		},
	})
}
