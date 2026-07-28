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

// TestFunctionalIndexMultiExpr tests indexes containing more than one functional expression,
// optionally mixed with plain column references. These tests cover postgres specific features
// that are not used in the shared enginetests for functional indexes.
func TestFunctionalIndexMultiExpr(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "mixed expressions and a plain column, filtering across all key parts",
			SetUpScript: []string{
				"CREATE TABLE t (pk int primary key, name text, age int, c1 int, c2 int);",
				"INSERT INTO t VALUES (1, 'alice', 30, 1, 2), (2, 'bob', 40, 3, 4);",
				"CREATE INDEX idx1 ON t (upper(name), age, (c1 + c2));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT pk FROM t WHERE upper(name) = 'ALICE' AND age = 30 AND c1 + c2 = 3;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT pk FROM t WHERE upper(name) = 'BOB' AND age = 40 AND c1 + c2 = 7;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT pg_get_indexdef('idx1'::regclass);",
					Expected: []sql.Row{{"CREATE INDEX idx1 ON public.t USING btree ((upper(name)), age, (c1 + c2))"}},
				},
				{
					Query:    "SELECT indexdef FROM pg_indexes WHERE indexname = 'idx1';",
					Expected: []sql.Row{{"CREATE INDEX idx1 ON public.t USING btree ((upper(name)), age, (c1 + c2))"}},
				},
			},
		},
		{
			Name: "expression-first ordering",
			SetUpScript: []string{
				"CREATE TABLE t (pk int primary key, c1 int, c2 int, c3 int);",
				"INSERT INTO t VALUES (1, 1, 2, 3), (2, 4, 5, 6);",
				"CREATE INDEX idx1 ON t ((c1 + c2), c3, (c1 * c2));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT pk FROM t WHERE c1 + c2 = 3 AND c3 = 3 AND c1 * c2 = 2;",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "two expressions sharing columns but with different operators",
			SetUpScript: []string{
				"CREATE TABLE t (pk int primary key, c1 int, c2 int);",
				"INSERT INTO t VALUES (1, 1, 2), (2, 4, 5);",
				"CREATE INDEX idx1 ON t ((c1 + c2), (c1 * c2));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT pk FROM t WHERE c1 + c2 = 3;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT pk FROM t WHERE c1 * c2 = 2;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT pk FROM t WHERE c1 + c2 = 9;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT pk FROM t WHERE c1 * c2 = 20;",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "composite range scan across a mixed expression/column key",
			SetUpScript: []string{
				"CREATE TABLE t (pk int primary key, c1 int, c2 int);",
				"INSERT INTO t VALUES (1, 1, 10), (2, 1, 20), (3, 1, 30), (4, 2, 10);",
				"CREATE INDEX idx1 ON t ((c1 * 10), c2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT pk FROM t WHERE c1 * 10 = 10 AND c2 > 15;",
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [t.pk]"},
						{" └─ IndexedTableAccess(t)"},
						{"     ├─ index: [t.!hidden!idx1!0!0,t.c2]"},
						{"     └─ filters: [{[10, 10], (15, ∞)}]"},
					},
				},
				{
					Query:    "SELECT pk FROM t WHERE c1 * 10 = 10 AND c2 > 15 ORDER BY pk;",
					Expected: []sql.Row{{2}, {3}},
				},
			},
		},
		{
			Name: "UNIQUE constraint enforced across all expressions",
			SetUpScript: []string{
				"CREATE TABLE t (pk int primary key, c1 int, c2 int, c3 int);",
				"CREATE UNIQUE INDEX idx1 ON t ((c1 * 10), c2, (c3 * 10));",
				"INSERT INTO t VALUES (1, 1, 2, 3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO t VALUES (2, 1, 2, 3);",
					ExpectedErr: "unique",
				},
				{
					Query:    "INSERT INTO t VALUES (3, 1, 3, 3);",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DROP INDEX removes only its own hidden columns, a second index survives",
			SetUpScript: []string{
				"CREATE TABLE t (pk int primary key, c1 int, c2 int, c3 int);",
				"INSERT INTO t VALUES (1, 10, 20, 30);",
				"CREATE INDEX idx2 ON t ((c2 * 100));",
				"CREATE INDEX idx1 ON t ((c1 * 10), c2, (c3 * 10));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT pk FROM t WHERE c1 * 10 = 100 AND c2 = 20 AND c3 * 10 = 300;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "DROP INDEX idx1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT pk, c2 * 100 FROM t;",
					Expected: []sql.Row{{1, 2000}},
				},
				{
					Query:    "DROP INDEX idx2;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
