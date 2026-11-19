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

// TestConcurrentWritesWithoutRowLocks tests concurrent writes with row-level locking turned off. No statement waits
// for another, and conflicting changes are found when the later of two transactions commits.
func TestConcurrentWritesWithoutRowLocks(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "concurrent increments of the same row",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE c (k TEXT, w INT, n INT NOT NULL, PRIMARY KEY (k, w));`,
				`INSERT INTO c VALUES ('x', 1, 0), ('y', 1, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON c TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO c (k, w, n) VALUES ('x', 1, 1) ON CONFLICT (k, w) DO UPDATE SET n = c.n + 1 RETURNING n;`,
					Expected: []sql.Row{{1}},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `INSERT INTO c (k, w, n) VALUES ('x', 1, 1) ON CONFLICT (k, w) DO UPDATE SET n = c.n + 1 RETURNING n;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:           `COMMIT;`,
					ExpectedErr:     "serialization failure",
					ExpectedErrCode: "40001",
				},
				{
					Query:    `SELECT k, n FROM c ORDER BY k;`,
					Expected: []sql.Row{{"x", 1}, {"y", 0}},
				},
			},
		},
		{
			Name: "concurrent increments of different rows",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE c (k TEXT, w INT, n INT NOT NULL, PRIMARY KEY (k, w));`,
				`INSERT INTO c VALUES ('x', 1, 0), ('y', 1, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON c TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `UPDATE c SET n = n + 1 WHERE k = 'x' RETURNING n;`,
					Expected: []sql.Row{{1}},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `UPDATE c SET n = n + 1 WHERE k = 'y' RETURNING n;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `COMMIT;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT k, n FROM c ORDER BY k;`,
					Expected: []sql.Row{{"x", 1}, {"y", 1}},
				},
			},
		},
		{
			Name: "concurrent increments of the same column",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE t (pk INT PRIMARY KEY, a INT, b INT);`,
				`INSERT INTO t VALUES (1, 0, 0), (2, 0, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON t TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `UPDATE t SET a = a + 1 WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `UPDATE t SET a = a + 1 WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `COMMIT;`,
					ExpectedErr:     "serialization failure",
					ExpectedErrCode: "40001",
				},
				{
					Query:    `SELECT * FROM t ORDER BY pk;`,
					Expected: []sql.Row{{1, 1, 0}, {2, 0, 0}},
				},
			},
		},
		{
			Name: "concurrent updates of the same column to different values",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE t (pk INT PRIMARY KEY, a INT, b INT);`,
				`INSERT INTO t VALUES (1, 0, 0), (2, 0, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON t TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `UPDATE t SET a = 5 WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `UPDATE t SET a = 7 WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `COMMIT;`,
					ExpectedErr:     "serialization failure",
					ExpectedErrCode: "40001",
				},
				{
					Query:    `SELECT * FROM t ORDER BY pk;`,
					Expected: []sql.Row{{1, 7, 0}, {2, 0, 0}},
				},
			},
		},
		{
			Name: "concurrent updates that agree on a shared column but differ in another",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE t (pk INT PRIMARY KEY, a INT, b INT);`,
				`INSERT INTO t VALUES (1, 0, 0), (2, 0, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON t TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `UPDATE t SET a = a + 1, b = 5 WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `UPDATE t SET a = a + 1 WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `COMMIT;`,
					ExpectedErr:     "serialization failure",
					ExpectedErrCode: "40001",
				},
				{
					Query:    `SELECT * FROM t ORDER BY pk;`,
					Expected: []sql.Row{{1, 1, 0}, {2, 0, 0}},
				},
			},
		},
		{
			Name: "concurrent updates of different columns in the same row",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE t (pk INT PRIMARY KEY, a INT, b INT);`,
				`INSERT INTO t VALUES (1, 0, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON t TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `UPDATE t SET b = a + 1;`,
					Expected: []sql.Row{},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `UPDATE t SET a = a + 1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `COMMIT;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM t;`,
					Expected: []sql.Row{{1, 1, 1}},
				},
			},
		},
		{
			Name: "concurrent deletes of the same row",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE t (pk INT PRIMARY KEY, a INT, b INT);`,
				`INSERT INTO t VALUES (1, 0, 0), (2, 0, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON t TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DELETE FROM t WHERE pk = 1 RETURNING *;`,
					Expected: []sql.Row{{1, 0, 0}},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `DELETE FROM t WHERE pk = 1 RETURNING *;`,
					Expected: []sql.Row{{1, 0, 0}},
				},
				{
					Query:           `COMMIT;`,
					ExpectedErr:     "serialization failure",
					ExpectedErrCode: "40001",
				},
				{
					Query:    `SELECT * FROM t;`,
					Expected: []sql.Row{{2, 0, 0}},
				},
			},
		},
		{
			Name: "concurrent delete and update of the same row",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE t (pk INT PRIMARY KEY, a INT, b INT);`,
				`INSERT INTO t VALUES (1, 0, 0), (2, 0, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON t TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DELETE FROM t WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `UPDATE t SET b = 1 WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `COMMIT;`,
					ExpectedErr:     "serialization failure",
					ExpectedErrCode: "40001",
				},
				{
					Query:    `SELECT * FROM t ORDER BY pk;`,
					Expected: []sql.Row{{1, 0, 1}, {2, 0, 0}},
				},
			},
		},
		{
			Name: "concurrent inserts of the same row",
			SetUpScript: []string{
				`SET dolt_row_level_locking = 0;`,
				`CREATE USER user1 PASSWORD 'password';`,
				`CREATE TABLE t (pk INT PRIMARY KEY, a INT, b INT);`,
				`INSERT INTO t VALUES (1, 0, 0), (2, 0, 0);`,
				`GRANT ALL PRIVILEGES ON SCHEMA public TO user1;`,
				`GRANT ALL PRIVILEGES ON t TO user1;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `BEGIN;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO t VALUES (3, 0, 0);`,
					Expected: []sql.Row{},
				},
				{
					Username: "user1",
					Password: "password",
					Query:    `INSERT INTO t VALUES (3, 0, 0);`,
					Expected: []sql.Row{},
				},
				{
					Query:           `COMMIT;`,
					ExpectedErr:     "serialization failure",
					ExpectedErrCode: "40001",
				},
				{
					Query:    `SELECT * FROM t ORDER BY pk;`,
					Expected: []sql.Row{{1, 0, 0}, {2, 0, 0}, {3, 0, 0}},
				},
			},
		},
	})
}

// TestConcurrentWrites tests concurrent writes of the same rows. A statement that changes a row waits for a
// transaction that has already changed it. If that transaction commits, the statement fails with a serialization
// failure, since what it read of the row is out of date; if it rolls back, the statement goes ahead.
func TestConcurrentWrites(t *testing.T) {
	counterSetup := []string{
		`CREATE TABLE c (k TEXT, w INT, n INT NOT NULL, PRIMARY KEY (k, w));`,
		`INSERT INTO c VALUES ('x', 1, 0), ('y', 1, 0);`,
	}
	tableSetup := []string{
		`CREATE TABLE t (pk INT PRIMARY KEY, a INT, b INT);`,
		`INSERT INTO t VALUES (1, 0, 0), (2, 0, 0);`,
	}
	const concurrentUpdate = "could not serialize access due to concurrent update"
	const concurrentDelete = "could not serialize access due to concurrent delete"

	RunTransactionTests(t, []ScriptTest{
		{
			Name:        "concurrent increments of the same row",
			SetUpScript: counterSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{
					Query:    `/* client a */ INSERT INTO c (k, w, n) VALUES ('x', 1, 1) ON CONFLICT (k, w) DO UPDATE SET n = c.n + 1 RETURNING n;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:            `/* client b */ INSERT INTO c (k, w, n) VALUES ('x', 1, 1) ON CONFLICT (k, w) DO UPDATE SET n = c.n + 1 RETURNING n;`,
					ExpectedBlocking: true,
					ExpectedErr:      concurrentUpdate,
					ExpectedErrCode:  "40001",
				},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{
					// Trying again, the statement sees the committed increment and adds its own.
					Query:    `/* client b */ INSERT INTO c (k, w, n) VALUES ('x', 1, 1) ON CONFLICT (k, w) DO UPDATE SET n = c.n + 1 RETURNING n;`,
					Expected: []sql.Row{{2}},
				},
				{Query: `/* client a */ SELECT k, n FROM c ORDER BY k;`, Expected: []sql.Row{{"x", 2}, {"y", 0}}},
			},
		},
		{
			Name:        "concurrent increments of different rows",
			SetUpScript: counterSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ UPDATE c SET n = n + 1 WHERE k = 'x' RETURNING n;`, Expected: []sql.Row{{1}}},
				{Query: `/* client b */ UPDATE c SET n = n + 1 WHERE k = 'y' RETURNING n;`, Expected: []sql.Row{{1}}},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{Query: `/* client a */ SELECT k, n FROM c ORDER BY k;`, Expected: []sql.Row{{"x", 1}, {"y", 1}}},
			},
		},
		{
			Name:        "concurrent increments of the same column",
			SetUpScript: tableSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ UPDATE t SET a = a + 1 WHERE pk = 1;`, Expected: []sql.Row{}},
				{
					Query:            `/* client b */ UPDATE t SET a = a + 1 WHERE pk = 1;`,
					ExpectedBlocking: true,
					ExpectedErr:      concurrentUpdate,
					ExpectedErrCode:  "40001",
				},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{Query: `/* client b */ UPDATE t SET a = a + 1 WHERE pk = 1;`, Expected: []sql.Row{}},
				{Query: `/* client a */ SELECT * FROM t ORDER BY pk;`, Expected: []sql.Row{{1, 2, 0}, {2, 0, 0}}},
			},
		},
		{
			Name:        "concurrent updates of the same column to different values",
			SetUpScript: tableSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ UPDATE t SET a = 5 WHERE pk = 1;`, Expected: []sql.Row{}},
				{
					Query:            `/* client b */ UPDATE t SET a = 7 WHERE pk = 1;`,
					ExpectedBlocking: true,
					ExpectedErr:      concurrentUpdate,
					ExpectedErrCode:  "40001",
				},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{Query: `/* client a */ SELECT * FROM t ORDER BY pk;`, Expected: []sql.Row{{1, 5, 0}, {2, 0, 0}}},
			},
		},
		{
			Name:        "concurrent updates of the same row proceed when the first rolls back",
			SetUpScript: tableSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ UPDATE t SET a = 5 WHERE pk = 1;`, Expected: []sql.Row{}},
				{Query: `/* client b */ UPDATE t SET a = 7 WHERE pk = 1;`, ExpectedBlocking: true},
				{Query: `/* client a */ ROLLBACK;`, Expected: []sql.Row{}},
				{Query: `/* client b */ SELECT * FROM t ORDER BY pk;`, Expected: []sql.Row{{1, 7, 0}, {2, 0, 0}}},
			},
		},
		{
			Name:        "concurrent updates that agree on a shared column but differ in another",
			SetUpScript: tableSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ UPDATE t SET a = a + 1, b = 5 WHERE pk = 1;`, Expected: []sql.Row{}},
				{
					Query:            `/* client b */ UPDATE t SET a = a + 1 WHERE pk = 1;`,
					ExpectedBlocking: true,
					ExpectedErr:      concurrentUpdate,
					ExpectedErrCode:  "40001",
				},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{Query: `/* client a */ SELECT * FROM t ORDER BY pk;`, Expected: []sql.Row{{1, 1, 5}, {2, 0, 0}}},
			},
		},
		{
			// Without row locks these changes are merged, because they are to different columns. With them, the
			// second statement can't know that until it has waited, and by then its view of the row is out of date.
			Name:        "concurrent updates of different columns in the same row",
			SetUpScript: tableSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ UPDATE t SET b = a + 1 WHERE pk = 1;`, Expected: []sql.Row{}},
				{
					Query:            `/* client b */ UPDATE t SET a = a + 1 WHERE pk = 1;`,
					ExpectedBlocking: true,
					ExpectedErr:      concurrentUpdate,
					ExpectedErrCode:  "40001",
				},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{Query: `/* client b */ UPDATE t SET a = a + 1 WHERE pk = 1;`, Expected: []sql.Row{}},
				{Query: `/* client a */ SELECT * FROM t ORDER BY pk;`, Expected: []sql.Row{{1, 1, 1}, {2, 0, 0}}},
			},
		},
		{
			Name:        "concurrent deletes of the same row",
			SetUpScript: tableSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ DELETE FROM t WHERE pk = 1 RETURNING *;`, Expected: []sql.Row{{1, 0, 0}}},
				{
					Query:            `/* client b */ DELETE FROM t WHERE pk = 1 RETURNING *;`,
					ExpectedBlocking: true,
					ExpectedErr:      concurrentDelete,
					ExpectedErrCode:  "40001",
				},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{Query: `/* client b */ DELETE FROM t WHERE pk = 1 RETURNING *;`, Expected: []sql.Row{}},
				{Query: `/* client a */ SELECT * FROM t;`, Expected: []sql.Row{{2, 0, 0}}},
			},
		},
		{
			Name:        "concurrent delete and update of the same row",
			SetUpScript: tableSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ DELETE FROM t WHERE pk = 1;`, Expected: []sql.Row{}},
				{
					Query:            `/* client b */ UPDATE t SET b = 1 WHERE pk = 1;`,
					ExpectedBlocking: true,
					ExpectedErr:      concurrentDelete,
					ExpectedErrCode:  "40001",
				},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{Query: `/* client a */ SELECT * FROM t ORDER BY pk;`, Expected: []sql.Row{{2, 0, 0}}},
			},
		},
		{
			Name:        "concurrent inserts of the same row",
			SetUpScript: tableSetup,
			Assertions: []ScriptTestAssertion{
				{Query: `/* client a */ BEGIN;`, Expected: []sql.Row{}},
				{Query: `/* client a */ INSERT INTO t VALUES (3, 0, 0);`, Expected: []sql.Row{}},
				{
					Query:            `/* client b */ INSERT INTO t VALUES (3, 1, 1);`,
					ExpectedBlocking: true,
					ExpectedErr:      "could not serialize access due to concurrent insert of the same key",
					ExpectedErrCode:  "40001",
				},
				{Query: `/* client a */ COMMIT;`, Expected: []sql.Row{}},
				{Query: `/* client b */ INSERT INTO t VALUES (3, 1, 1);`, ExpectedErr: "duplicate"},
				{Query: `/* client a */ SELECT * FROM t ORDER BY pk;`, Expected: []sql.Row{{1, 0, 0}, {2, 0, 0}, {3, 0, 0}}},
			},
		},
	})
}
