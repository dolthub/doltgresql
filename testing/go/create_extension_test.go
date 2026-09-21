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

func TestCreateExtension(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "create extension uuid-ossp after setting search_path to empty",
			SetUpScript: []string{
				`SELECT pg_catalog.set_config('search_path', '', false);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;`,
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT uuid_nil();",
					ExpectedErr: `function: 'uuid_nil' not found`,
				},
				{
					Query:    "SELECT public.uuid_nil();",
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
			},
		},
		{
			Name: "alter table with default expr using extension function when search_path is empty",
			SetUpScript: []string{
				`SELECT pg_catalog.set_config('search_path', '', false);`,
				`CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;`,
				`CREATE TABLE public.goals (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    note_id uuid,
    completion_timestamp timestamp without time zone,
    due_date timestamp without time zone
);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `ALTER TABLE ONLY public.goals ADD CONSTRAINT goals_pkey PRIMARY KEY (id);`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "uuid-ossp is not available before it is created",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT count(*) FROM pg_catalog.pg_extension;`,
					Expected: []sql.Row{{0}},
				},
				{
					Query: `SELECT name, installed_version FROM pg_catalog.pg_available_extensions WHERE name = 'uuid-ossp';`,
					Expected: []sql.Row{
						{"uuid-ossp", nil},
					},
				},
				{
					Query:       "SELECT uuid_nil();",
					ExpectedErr: `function: 'uuid_nil' not found`,
				},
				{
					Query:    `CREATE EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT uuid_nil();",
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
			},
		},
		{
			Name: "only emulated extensions may be created",
			Assertions: []ScriptTestAssertion{
				{
					Query:       `CREATE EXTENSION "doltgres_no_such_extension";`,
					ExpectedErr: `extension "doltgres_no_such_extension" is not available`,
				},
				{
					Query:       `CREATE EXTENSION IF NOT EXISTS "doltgres_no_such_extension";`,
					ExpectedErr: `extension "doltgres_no_such_extension" is not available`,
				},
				{
					Query:       `CREATE EXTENSION "UUID-OSSP";`,
					ExpectedErr: `extension "UUID-OSSP" is not available`,
				},
				{
					Query:    `CREATE EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
				{
					Query:       `CREATE EXTENSION "uuid-ossp";`,
					ExpectedErr: `extension "uuid-ossp" already exists`,
				},
				{
					Query:    `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT uuid_nil();",
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
			},
		},
		{
			Name: "uuid-ossp options that are not yet supported",
			Assertions: []ScriptTestAssertion{
				{
					Query:       `CREATE EXTENSION "uuid-ossp" VERSION oldversion;`,
					ExpectedErr: "VERSION is not yet supported",
				},
				{
					Query:       `CREATE EXTENSION "uuid-ossp" WITH SCHEMA myschema;`,
					ExpectedErr: "non public SCHEMA is not yet supported",
				},
				{
					Query:       `CREATE EXTENSION "uuid-ossp" CASCADE;`,
					ExpectedErr: "CASCADE is not yet supported",
				},
				{
					Query:       `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr: "DROP EXTENSION is not yet implemented",
				},
			},
		},
		{
			Name: "uuid-ossp installation participates in branches",
			SetUpScript: []string{
				`SELECT dolt_commit('--allow-empty', '-m', 'initial commit');`,
				`SELECT dolt_checkout('-b', 'ext');`,
				`CREATE EXTENSION "uuid-ossp";`,
				`SELECT dolt_commit('-Am', 'create the extension');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT extname FROM pg_catalog.pg_extension;`,
					Expected: []sql.Row{{"uuid-ossp"}},
				},
				{
					Query:            `SELECT dolt_checkout('main');`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT extname FROM pg_catalog.pg_extension;`,
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT uuid_nil();",
					ExpectedErr: `function: 'uuid_nil' not found`,
				},
				{
					Query:            `SELECT dolt_merge('ext');`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT extname, extversion FROM pg_catalog.pg_extension;`,
					Expected: []sql.Row{{"uuid-ossp", "1.1"}},
				},
				{
					Query:    "SELECT uuid_nil();",
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
			},
		},
	})
}
