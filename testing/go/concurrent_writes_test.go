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

// TestConcurrentWrites tests that a transaction writing a row that another transaction committed first fails with a
// serialization failure, rather than losing the other transaction's write.
func TestConcurrentWrites(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "concurrent increments of the same row",
			SetUpScript: []string{
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
			Name: "concurrent updates of different columns in the same row",
			SetUpScript: []string{
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
					Query:           `COMMIT;`,
					ExpectedErr:     "serialization failure",
					ExpectedErrCode: "40001",
				},
				{
					Query:    `SELECT * FROM t;`,
					Expected: []sql.Row{{1, 1, 0}},
				},
			},
		},
		{
			Name: "concurrent deletes of the same row",
			SetUpScript: []string{
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
	})
}
