// Copyright 2024 Dolthub, Inc.
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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCombocid(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_combocid)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_combocid,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TEMP TABLE combocidtest (foobar int);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest VALUES (1);`,
			},
			{
				Statement: `INSERT INTO combocidtest VALUES (2);`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,1)`, 10, 1}, {`(0,2)`, 11, 2}},
			},
			{
				Statement: `SAVEPOINT s1;`,
			},
			{
				Statement: `UPDATE combocidtest SET foobar = foobar + 10;`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,3)`, 12, 11}, {`(0,4)`, 12, 12}},
			},
			{
				Statement: `ROLLBACK TO s1;`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,1)`, 0, 1}, {`(0,2)`, 1, 2}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,1)`, 0, 1}, {`(0,2)`, 1, 2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO combocidtest VALUES (333);`,
			},
			{
				Statement: `DECLARE c CURSOR FOR SELECT ctid,cmin,* FROM combocidtest;`,
			},
			{
				Statement: `DELETE FROM combocidtest;`,
			},
			{
				Statement: `FETCH ALL FROM c;`,
				Results:   []sql.Row{{`(0,1)`, 1, 1}, {`(0,2)`, 1, 2}, {`(0,5)`, 0, 333}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,1)`, 1, 1}, {`(0,2)`, 1, 2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest SELECT 1 LIMIT 0;`,
			},
			{
				Statement: `INSERT INTO combocidtest VALUES (444);`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,1)`, 1, 1}, {`(0,2)`, 1, 2}, {`(0,6)`, 10, 444}},
			},
			{
				Statement: `SAVEPOINT s1;`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest FOR UPDATE;`,
				Results:   []sql.Row{{`(0,1)`, 1, 1}, {`(0,2)`, 1, 2}, {`(0,6)`, 10, 444}},
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,1)`, 1, 1}, {`(0,2)`, 1, 2}, {`(0,6)`, 10, 444}},
			},
			{
				Statement: `UPDATE combocidtest SET foobar = foobar + 10;`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,7)`, 12, 11}, {`(0,8)`, 12, 12}, {`(0,9)`, 12, 454}},
			},
			{
				Statement: `ROLLBACK TO s1;`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,1)`, 12, 1}, {`(0,2)`, 12, 2}, {`(0,6)`, 0, 444}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT ctid,cmin,* FROM combocidtest;`,
				Results:   []sql.Row{{`(0,1)`, 12, 1}, {`(0,2)`, 12, 2}, {`(0,6)`, 0, 444}},
			},
			{
				Statement: `CREATE TABLE IF NOT EXISTS testcase(
	id int PRIMARY KEY,
	balance numeric
);`,
			},
			{
				Statement: `INSERT INTO testcase VALUES (1, 0);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT * FROM testcase WHERE testcase.id = 1 FOR UPDATE;`,
				Results:   []sql.Row{{1, 0}},
			},
			{
				Statement: `UPDATE testcase SET balance = balance + 400 WHERE id=1;`,
			},
			{
				Statement: `SAVEPOINT subxact;`,
			},
			{
				Statement: `UPDATE testcase SET balance = balance - 100 WHERE id=1;`,
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT subxact;`,
			},
			{
				Statement: `SELECT * FROM testcase WHERE id = 1 FOR UPDATE;`,
				Results:   []sql.Row{{1, 400}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE testcase;`,
			},
		},
	})
}
