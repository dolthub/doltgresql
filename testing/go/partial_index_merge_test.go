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

// partialIndexMergeSetup creates a table with a partial unique index that only covers rows without a slot.
var partialIndexMergeSetup = []string{
	"CREATE TABLE t (id INT PRIMARY KEY, p INT, m INT, slot INT);",
	"CREATE UNIQUE INDEX t_partial ON t (p, m) WHERE slot IS NULL;",
}

// TestPartialIndexTransactionMerge checks that the merge run when a transaction commits on top of a working set
// changed by a concurrent transaction respects the predicate of partial indexes.
func TestPartialIndexTransactionMerge(t *testing.T) {
	RunTransactionTests(t, []ScriptTest{
		{
			Name:        "concurrent insert of a row outside the predicate",
			SetUpScript: partialIndexMergeSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (1, 9, 9, NULL);"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 1, 1, 5);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{
					Query:    "/* client a */ SELECT id FROM t WHERE p = 1 AND m = 1 AND slot IS NULL;",
					Expected: []sql.Row{},
				},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, 1, NULL);"},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (4, 1, 1, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
				{
					Query:    "/* client b */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 9, 9, nil}, {2, 1, 1, 5}, {3, 1, 1, nil}},
				},
			},
		},
		{
			Name: "concurrent insert outside the predicate that shares the key of a row inside it",
			SetUpScript: append(partialIndexMergeSetup,
				"INSERT INTO t VALUES (1, 1, 1, NULL);",
			),
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (2, 8, 8, NULL);"},
				{Query: "/* client b */ INSERT INTO t VALUES (3, 1, 1, 6);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{
					Query:    "/* client a */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 1, 1, nil}, {2, 8, 8, nil}, {3, 1, 1, 6}},
				},
			},
		},
		{
			Name:        "delete of a row outside the predicate followed by another concurrent insert",
			SetUpScript: partialIndexMergeSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (1, 9, 9, NULL);"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 1, 1, 5);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ DELETE FROM t WHERE id = 2;"},
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 7, 7, NULL);"},
				{Query: "/* client b */ INSERT INTO t VALUES (4, 1, 1, 7);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ INSERT INTO t VALUES (5, 1, 1, NULL);"},
				{
					Query:    "/* client b */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 9, 9, nil}, {3, 7, 7, nil}, {4, 1, 1, 7}, {5, 1, 1, nil}},
				},
			},
		},
		{
			Name: "concurrent update that moves a row out of the predicate",
			SetUpScript: append(partialIndexMergeSetup,
				"INSERT INTO t VALUES (1, 1, 1, NULL);",
			),
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (2, 9, 9, NULL);"},
				{Query: "/* client b */ UPDATE t SET slot = 5 WHERE id = 1;"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, 1, NULL);"},
				{
					Query:    "/* client b */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 1, 1, 5}, {2, 9, 9, nil}, {3, 1, 1, nil}},
				},
			},
		},
		{
			Name: "update that moves a row out of the predicate frees its key for an insert in the same transaction",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot IS NULL;",
				"INSERT INTO t VALUES (1, 5, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 9, NULL);"},
				{Query: "/* client b */ UPDATE t SET slot = 10 WHERE id = 1;"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 5, NULL);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{
					Query:    "/* client a */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 5, 10}, {2, 5, nil}, {3, 9, nil}},
				},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (4, 5, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "concurrent update that moves a row into the predicate",
			SetUpScript: append(partialIndexMergeSetup,
				"INSERT INTO t VALUES (1, 1, 1, 5);",
			),
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (2, 9, 9, NULL);"},
				{Query: "/* client b */ UPDATE t SET slot = NULL WHERE id = 1;"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (3, 1, 1, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
				{Query: "/* client a */ DELETE FROM t WHERE id = 1;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, 1, NULL);"},
			},
		},
		{
			Name: "concurrent updates of the same row that merge to a row outside the predicate",
			SetUpScript: append(partialIndexMergeSetup,
				"INSERT INTO t VALUES (1, 1, 1, NULL);",
			),
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ UPDATE t SET m = 2 WHERE id = 1;"},
				{Query: "/* client b */ UPDATE t SET slot = 5 WHERE id = 1;"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{
					Query:    "/* client a */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 1, 2, 5}},
				},
				{Query: "/* client a */ INSERT INTO t VALUES (2, 1, 2, NULL);"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, 1, NULL);"},
			},
		},
		{
			Name: "concurrent deletes of a row inside the predicate",
			SetUpScript: append(partialIndexMergeSetup,
				"INSERT INTO t VALUES (1, 1, 1, NULL);",
			),
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (2, 9, 9, NULL);"},
				{Query: "/* client b */ DELETE FROM t WHERE id = 1;"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, 1, NULL);"},
			},
		},
		{
			Name:        "concurrent duplicate inserts inside the predicate still violate the index",
			SetUpScript: partialIndexMergeSetup,
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (1, 1, 1, NULL);"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 1, 1, NULL);"},
				{Query: "/* client a */ COMMIT;"},
				{
					Query:       "/* client b */ COMMIT;",
					ExpectedErr: "Unique Key Constraint Violation",
				},
				{
					Query:    "/* client a */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 1, 1, nil}},
				},
			},
		},
		{
			Name: "concurrent insert outside a predicate that evaluates to NULL",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot > 5;",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (1, 9, 9);"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 1, NULL);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, 6);"},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (4, 1, 7);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "concurrent insert outside a predicate on an enum column",
			SetUpScript: []string{
				"CREATE TYPE status AS ENUM ('open', 'closed');",
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, s status);",
				"CREATE UNIQUE INDEX t_partial ON t (p) WHERE s = 'open';",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (1, 9, 'open');"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 1, 'closed');"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, 'open');"},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (4, 1, 'open');",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "concurrent insert outside the predicate and a column drop",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, extra INT, p INT, m INT, slot INT);",
				"CREATE UNIQUE INDEX t_partial ON t (p, m) WHERE slot IS NULL;",
				"INSERT INTO t VALUES (1, 0, 1, 1, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ ALTER TABLE t DROP COLUMN extra;"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 0, 1, 1, 5);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{
					Query:    "/* client a */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 1, 1, nil}, {2, 1, 1, 5}},
				},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (3, 1, 1, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 2, 2, NULL);"},
				{Query: "/* client a */ DELETE FROM t WHERE id = 1;"},
				{Query: "/* client a */ INSERT INTO t VALUES (4, 1, 1, NULL);"},
			},
		},
		{
			Name: "concurrent insert on the old schema of a row whose new column default is inside the predicate",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT);",
				"INSERT INTO t VALUES (1, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client a */ ALTER TABLE t ADD COLUMN slot INT DEFAULT 10;"},
				{Query: "/* client a */ CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot > 5;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 5);"},
				{Query: "/* client a */ COMMIT;"},
				{
					Query:       "/* client b */ COMMIT;",
					ExpectedErr: "Unique Key Constraint Violation",
				},
				{
					Query:    "/* client a */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 5, 10}},
				},
			},
		},
		{
			Name: "concurrent insert on the old schema of a row whose new column default is outside the predicate",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT);",
				"INSERT INTO t VALUES (1, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client a */ ALTER TABLE t ADD COLUMN slot INT DEFAULT 1;"},
				{Query: "/* client a */ UPDATE t SET slot = 10 WHERE id = 1;"},
				{Query: "/* client a */ CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot > 5;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 5);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{
					Query:    "/* client a */ SELECT * FROM t ORDER BY id;",
					Expected: []sql.Row{{1, 5, 10}, {2, 5, 1}},
				},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (3, 5, 6);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "concurrent insert outside the predicate on a keyless table",
			SetUpScript: []string{
				"CREATE TABLE t (p INT, slot INT);",
				"CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot IS NULL;",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (9, NULL);"},
				{Query: "/* client b */ INSERT INTO t VALUES (1, 5);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ INSERT INTO t VALUES (1, NULL);"},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (1, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
				{
					Query:    "/* client b */ SELECT * FROM t ORDER BY p, slot;",
					Expected: []sql.Row{{1, 5}, {1, nil}, {9, nil}},
				},
			},
		},
		{
			Name: "partial index created concurrently with an insert outside the predicate",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"INSERT INTO t VALUES (1, 1, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot IS NULL;"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 2, 5);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, NULL);"},
				{Query: "/* client a */ INSERT INTO t VALUES (4, 2, NULL);"},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (5, 2, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "partial index created by the last committer while another transaction inserts outside the predicate",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"INSERT INTO t VALUES (1, 1, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (2, 2, 5);"},
				{Query: "/* client b */ CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot IS NULL;"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{Query: "/* client a */ INSERT INTO t VALUES (3, 1, NULL);"},
				{Query: "/* client a */ INSERT INTO t VALUES (4, 2, NULL);"},
				{
					Query:       "/* client a */ INSERT INTO t VALUES (5, 2, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "partial non-unique index with a concurrent insert outside the predicate",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"CREATE INDEX t_partial ON t (p) WHERE slot IS NULL;",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ BEGIN;"},
				{Query: "/* client b */ BEGIN;"},
				{Query: "/* client a */ INSERT INTO t VALUES (1, 1, NULL);"},
				{Query: "/* client b */ INSERT INTO t VALUES (2, 1, 5);"},
				{Query: "/* client a */ COMMIT;"},
				{Query: "/* client b */ COMMIT;"},
				{
					Query:    "/* client a */ SELECT id FROM t WHERE p = 1 AND slot IS NULL;",
					Expected: []sql.Row{{1}},
				},
				{Query: "/* client a */ DELETE FROM t WHERE id = 2;"},
				{
					Query:    "/* client a */ SELECT id FROM t WHERE p = 1 AND slot IS NULL;",
					Expected: []sql.Row{{1}},
				},
			},
		},
	})
}

// TestPartialIndexBranchMerge checks that branch merges and conflict resolution respect the predicate of partial
// indexes.
func TestPartialIndexBranchMerge(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "merge of a branch that adds a partial unique index",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"INSERT INTO t VALUES (1, 1, 5);",
				"SELECT dolt_commit('-Am', 'create t');",
				"SELECT dolt_checkout('-b', 'other');",
				"CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot IS NULL;",
				"SELECT dolt_commit('-Am', 'add index');",
				"SELECT dolt_checkout('main');",
				"INSERT INTO t VALUES (2, 2, 5);",
				"SELECT dolt_commit('-Am', 'add row');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT dolt_merge('other');",
					SkipResultsCheck: true,
				},
				{Query: "INSERT INTO t VALUES (3, 1, NULL);"},
				{Query: "INSERT INTO t VALUES (4, 2, NULL);"},
				{
					Query:       "INSERT INTO t VALUES (5, 2, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
		{
			Name: "merge of a branch that adds a partial non-unique index",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"INSERT INTO t VALUES (1, 1, 5);",
				"SELECT dolt_commit('-Am', 'create t');",
				"SELECT dolt_checkout('-b', 'other');",
				"CREATE INDEX t_partial ON t (p) WHERE slot IS NULL;",
				"SELECT dolt_commit('-Am', 'add index');",
				"SELECT dolt_checkout('main');",
				"INSERT INTO t VALUES (2, 1, 5), (3, 1, NULL);",
				"SELECT dolt_commit('-Am', 'add rows');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT dolt_merge('other');",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT id FROM t WHERE p = 1 AND slot IS NULL;",
					Expected: []sql.Row{{3}},
				},
				{Query: "DELETE FROM t WHERE slot IS NOT NULL;"},
				{
					Query:    "SELECT id FROM t WHERE p = 1 AND slot IS NULL;",
					Expected: []sql.Row{{3}},
				},
			},
		},
		{
			Name: "merge of a branch that inserts a row outside the predicate",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot IS NULL;",
				"INSERT INTO t VALUES (1, 1, NULL);",
				"SELECT dolt_commit('-Am', 'create t');",
				"SELECT dolt_checkout('-b', 'other');",
				"INSERT INTO t VALUES (2, 1, 5), (3, 2, 5);",
				"SELECT dolt_commit('-Am', 'add rows');",
				"SELECT dolt_checkout('main');",
				"INSERT INTO t VALUES (4, 9, NULL);",
				"SELECT dolt_commit('-Am', 'add row');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT dolt_merge('other');",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT count(*) FROM dolt_constraint_violations;",
					Expected: []sql.Row{{0}},
				},
				{Query: "INSERT INTO t VALUES (5, 2, NULL);"},
			},
		},
		{
			Name: "resolving a conflict with their row outside the predicate",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot IS NULL;",
				"INSERT INTO t VALUES (1, 1, 10);",
				"SELECT dolt_commit('-Am', 'create t');",
				"SELECT dolt_checkout('-b', 'other');",
				"UPDATE t SET slot = 5 WHERE id = 1;",
				"SELECT dolt_commit('-Am', 'their slot');",
				"SELECT dolt_checkout('main');",
				"UPDATE t SET slot = NULL WHERE id = 1;",
				"SELECT dolt_commit('-Am', 'our slot');",
				"SET dolt_allow_commit_conflicts TO 1;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT dolt_merge('other');",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT dolt_conflicts_resolve('--theirs', 't');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT * FROM t;",
					Expected: []sql.Row{{1, 1, 5}},
				},
				{Query: "INSERT INTO t VALUES (2, 1, NULL);"},
			},
		},
		{
			Name: "resolving a conflict with their row inside the predicate",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, p INT, slot INT);",
				"CREATE UNIQUE INDEX t_partial ON t (p) WHERE slot IS NULL;",
				"INSERT INTO t VALUES (1, 1, 10);",
				"SELECT dolt_commit('-Am', 'create t');",
				"SELECT dolt_checkout('-b', 'other');",
				"UPDATE t SET slot = NULL WHERE id = 1;",
				"SELECT dolt_commit('-Am', 'their slot');",
				"SELECT dolt_checkout('main');",
				"UPDATE t SET slot = 5 WHERE id = 1;",
				"SELECT dolt_commit('-Am', 'our slot');",
				"SET dolt_allow_commit_conflicts TO 1;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT dolt_merge('other');",
					SkipResultsCheck: true,
				},
				{
					Query:    "SELECT dolt_conflicts_resolve('--theirs', 't');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT * FROM t;",
					Expected: []sql.Row{{1, 1, nil}},
				},
				{
					Query:       "INSERT INTO t VALUES (2, 1, NULL);",
					ExpectedErr: "duplicate unique key given",
				},
			},
		},
	})
}
