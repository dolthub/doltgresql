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

// TestLookupJoinExtendedKeyStoragePosition tests lookup joins whose key column is an
// extended-encoded type (e.g. uuid) that sits at a non-zero storage position among the
// source table's non-primary-key columns.
// Regression test for https://github.com/dolthub/doltgresql/issues/2979
func TestLookupJoinExtendedKeyStoragePosition(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "join key is the second non-PK column",
			SetUpScript: []string{
				"CREATE TABLE c (id uuid PRIMARY KEY, name text);",
				"CREATE TABLE m (seq int PRIMARY KEY, filler text, company_id uuid, label text);",
				"INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
				"INSERT INTO m VALUES (1, 'f', '11111111-1111-1111-1111-111111111111', 'L');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "join key is the first non-PK column",
			SetUpScript: []string{
				"CREATE TABLE c (id uuid PRIMARY KEY, name text);",
				"CREATE TABLE m (seq int PRIMARY KEY, company_id uuid, label text);",
				"INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
				"INSERT INTO m VALUES (1, '11111111-1111-1111-1111-111111111111', 'L');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "join key is the third non-PK column",
			SetUpScript: []string{
				"CREATE TABLE c (id uuid PRIMARY KEY, name text);",
				"CREATE TABLE m (seq int PRIMARY KEY, filler1 text, filler2 text, company_id uuid, label text);",
				"INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
				"INSERT INTO m VALUES (1, 'f1', 'f2', '11111111-1111-1111-1111-111111111111', 'L');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "composite key with both columns at non-zero storage positions",
			SetUpScript: []string{
				"CREATE TABLE c (id uuid PRIMARY KEY, tag text, code uuid, label text);",
				"CREATE TABLE m (seq int PRIMARY KEY, filler text, ref_code uuid, ref_label text);",
				"INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'c1', '22222222-2222-2222-2222-222222222222', 'match');",
				"INSERT INTO c VALUES ('33333333-3333-3333-3333-333333333333', 'c2', '22222222-2222-2222-2222-222222222222', 'other');",
				"INSERT INTO m VALUES (1, 'f', '22222222-2222-2222-2222-222222222222', 'match');",
				"INSERT INTO m VALUES (2, 'f', '22222222-2222-2222-2222-222222222222', 'other');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select /*+ lookup_join(m, c) */ HINT m.seq, c.tag from m join c on c.code = m.ref_code and c.label = m.ref_label order by m.seq;",
					Expected: []sql.Row{
						{1, "c1"},
						{2, "c2"},
					},
				},
			},
		},
	})
}
