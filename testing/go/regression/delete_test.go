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

func TestDelete(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_delete)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_delete,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE delete_test (
    id SERIAL PRIMARY KEY,
    a INT,
    b text
);`,
			},
			{
				Statement: `INSERT INTO delete_test (a) VALUES (10);`,
			},
			{
				Statement: `INSERT INTO delete_test (a, b) VALUES (50, repeat('x', 10000));`,
			},
			{
				Statement: `INSERT INTO delete_test (a) VALUES (100);`,
			},
			{
				Statement: `DELETE FROM delete_test AS dt WHERE dt.a > 75;`,
			},
			{
				Statement:   `DELETE FROM delete_test dt WHERE delete_test.a > 25;`,
				ErrorString: `invalid reference to FROM-clause entry for table "delete_test"`,
			},
			{
				Statement: `SELECT id, a, char_length(b) FROM delete_test;`,
				Results:   []sql.Row{{1, 10, ``}, {2, 50, 10000}},
			},
			{
				Statement: `DELETE FROM delete_test WHERE a > 25;`,
			},
			{
				Statement: `SELECT id, a, char_length(b) FROM delete_test;`,
				Results:   []sql.Row{{1, 10, ``}},
			},
			{
				Statement: `DROP TABLE delete_test;`,
			},
		},
	})
}
