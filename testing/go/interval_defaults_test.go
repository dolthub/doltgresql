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

// TestIntervalColumnDefaults covers interval defaults being serialized and reparsed by later statements.
// Regression test for https://github.com/dolthub/doltgresql/issues/3529.
func TestIntervalColumnDefaults(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "timestamp default survives index creation",
			SetUpScript: []string{
				`CREATE TABLE repro (expires_at timestamptz DEFAULT (now() + interval '3 minutes'));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE INDEX repro_expires_idx ON repro (expires_at);`,
					Expected: []sql.Row{},
				},
				{
					// A second index reloads the default after the first schema rewrite.
					Query:    `CREATE INDEX repro_expires_idx2 ON repro (expires_at);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO repro DEFAULT VALUES RETURNING expires_at = now() + interval '3 minutes';`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "timestamp default is applied by default values insert",
			SetUpScript: []string{
				`CREATE TABLE repro (expires_at timestamptz NOT NULL DEFAULT (now() + interval '3 minutes'));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// The expression already works without a stored default.
					Query:    `SELECT now() + interval '3 minutes' > now();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// Compare within the INSERT statement so the assertion does not depend on elapsed time.
					Query:    `INSERT INTO repro DEFAULT VALUES RETURNING expires_at = now() + interval '3 minutes';`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "timestamp default allows omitted and explicit column inserts",
			SetUpScript: []string{
				`CREATE TABLE repro (id int PRIMARY KEY, expires_at timestamptz NOT NULL DEFAULT (now() + interval '3 minutes'));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO repro (id) VALUES (1) RETURNING id, expires_at = now() + interval '3 minutes';`,
					Expected: []sql.Row{{1, "t"}},
				},
				{
					// Loading the schema must also succeed when the default is not used.
					Query:    `INSERT INTO repro VALUES (2, now()) RETURNING id, expires_at = now();`,
					Expected: []sql.Row{{2, "t"}},
				},
			},
		},
		{
			Name: "stored timestamp defaults preserve values before and after index creation",
			SetUpScript: []string{
				`CREATE TABLE repro (
					id int PRIMARY KEY DEFAULT 0,
					inserted_at timestamptz NOT NULL DEFAULT now(),
					expires_at timestamptz NOT NULL DEFAULT (now() + interval '3 minutes')
				);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO repro DEFAULT VALUES;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO repro (id) VALUES (1);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO repro (id, expires_at) VALUES (2, '2000-01-01 00:00:00+00');`,
					Expected: []sql.Row{},
				},
				{
					// Both defaults use the same statement time. Read the stored values in a later
					// statement to verify the exact offset without depending on elapsed time.
					Query: `SELECT id, CASE WHEN id = 2
						THEN expires_at = '2000-01-01 00:00:00+00'::timestamptz
						ELSE expires_at - inserted_at = interval '3 minutes' END
						FROM repro ORDER BY id;`,
					Expected: []sql.Row{{0, "t"}, {1, "t"}, {2, "t"}},
				},
				{
					Query: `SELECT id, (expires_at - now()) <= interval '3 minutes'
						FROM repro WHERE id <> 2 ORDER BY id;`,
					Expected: []sql.Row{{0, "t"}, {1, "t"}},
				},
				{
					Query:    `CREATE INDEX repro_expires_idx ON repro (expires_at);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE INDEX repro_expires_idx2 ON repro (expires_at);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO repro (id) VALUES (3);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT id, CASE WHEN id = 2
						THEN expires_at = '2000-01-01 00:00:00+00'::timestamptz
						ELSE expires_at - inserted_at = interval '3 minutes' END
						FROM repro ORDER BY id;`,
					Expected: []sql.Row{{0, "t"}, {1, "t"}, {2, "t"}, {3, "t"}},
				},
				{
					Query: `SELECT id, (expires_at - now()) <= interval '3 minutes'
						FROM repro WHERE id <> 2 ORDER BY id;`,
					Expected: []sql.Row{{0, "t"}, {1, "t"}, {3, "t"}},
				},
			},
		},
		{
			Name: "interval column defaults preserve duration values",
			SetUpScript: []string{
				`CREATE TABLE repro (
					minutes interval DEFAULT interval '3 minutes',
					calendar interval DEFAULT interval '1 year 2 mons -3 days 04:05:06.123456',
					negative interval DEFAULT interval '-3 minutes',
					zero interval DEFAULT interval '0 seconds',
					precise interval DEFAULT interval (3) '00:00:01.123456'
				);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `INSERT INTO repro DEFAULT VALUES RETURNING
						minutes = '3 minutes'::interval,
						calendar = '1 year 2 mons -3 days 04:05:06.123456'::interval,
						negative = '-3 minutes'::interval,
						zero = '0 seconds'::interval,
						precise = '00:00:01.123'::interval;`,
					Expected: []sql.Row{{"t", "t", "t", "t", "t"}},
				},
				{
					Query: `SELECT
						minutes = '3 minutes'::interval,
						calendar = '1 year 2 mons -3 days 04:05:06.123456'::interval,
						negative = '-3 minutes'::interval,
						zero = '0 seconds'::interval,
						precise = '00:00:01.123'::interval
						FROM repro;`,
					Expected: []sql.Row{{"t", "t", "t", "t", "t"}},
				},
			},
		},
		{
			Name: "interval timestamp default is readable through information schema",
			SetUpScript: []string{
				`CREATE TABLE repro (expires_at timestamptz DEFAULT (now() + interval '3 minutes'));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT column_name, column_default IS NOT NULL
						FROM information_schema.columns
						WHERE table_schema = 'public' AND table_name = 'repro';`,
					Expected: []sql.Row{{"expires_at", "t"}},
				},
			},
		},
	})
}
