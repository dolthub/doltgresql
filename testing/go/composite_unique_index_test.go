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

// TestCompositeUniqueIndexExtendedAndAdaptive tests composite UNIQUE indexes that combine an
// extended-encoded type (uuid) with an adaptive-encoded type (text, unbounded varchar).
// Regression test for https://github.com/dolthub/doltgresql/issues/2886
func TestCompositeUniqueIndexExtendedAndAdaptive(t *testing.T) {
	RunScripts(t, []ScriptTest{
		// Keyless tables
		{
			Name: "keyless: UNIQUE(uuid, text) — insert into empty table",
			SetUpScript: []string{
				"CREATE TABLE t (iid uuid, slug text, UNIQUE(iid, slug));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES ('22222222-2222-2222-2222-222222222222', 'hello');",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'world');",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
					ExpectedErr: "unique",
				},
			},
		},
		{
			Name: "keyless: UNIQUE(text, uuid) — column order reversed",
			SetUpScript: []string{
				"CREATE TABLE t (slug text, iid uuid, UNIQUE(slug, iid));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO t VALUES ('hello', '11111111-1111-1111-1111-111111111111');",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t VALUES ('hello', '11111111-1111-1111-1111-111111111111');",
					ExpectedErr: "unique",
				},
				{
					Query:    "INSERT INTO t VALUES ('world', '11111111-1111-1111-1111-111111111111');",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "keyless: UNIQUE(uuid, varchar) — unbounded varchar",
			SetUpScript: []string{
				"CREATE TABLE t (iid uuid, slug varchar, UNIQUE(iid, slug));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t VALUES ('11111111-1111-1111-1111-111111111111', 'hello');",
					ExpectedErr: "unique",
				},
			},
		},
		{
			Name: "keyless: UNIQUE(uuid, text) — NULL uuid skips unique check",
			SetUpScript: []string{
				"CREATE TABLE t (iid uuid, slug text, UNIQUE(iid, slug));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO t VALUES (NULL, 'hello');",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (NULL, 'hello');",
					Expected: []sql.Row{},
				},
			},
		},

		// Keyed tables
		{
			Name: "keyed: UNIQUE(uuid, text) on table with primary key",
			SetUpScript: []string{
				"CREATE TABLE t (pk int primary key, iid uuid, slug text, UNIQUE(iid, slug));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO t VALUES (1, '11111111-1111-1111-1111-111111111111', 'hello');",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t VALUES (2, '11111111-1111-1111-1111-111111111111', 'hello');",
					ExpectedErr: "unique",
				},
				{
					Query:    "INSERT INTO t VALUES (3, '11111111-1111-1111-1111-111111111111', 'world');",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "keyed: UNIQUE(uuid, varchar) — unbounded varchar on table with primary key",
			SetUpScript: []string{
				"CREATE TABLE t (pk int primary key, iid uuid, slug varchar, UNIQUE(iid, slug));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO t VALUES (1, '11111111-1111-1111-1111-111111111111', 'hello');",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t VALUES (2, '11111111-1111-1111-1111-111111111111', 'hello');",
					ExpectedErr: "unique",
				},
			},
		},
	})
}
