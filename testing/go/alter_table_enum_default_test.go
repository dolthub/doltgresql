// Copyright 2026 Dolthub, Inc.
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
