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

func TestSelectHaving2(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_select_having_2)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_select_having_2,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE test_having (a int, b int, c char(8), d char);`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (0, 1, 'XXXX', 'A');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (1, 2, 'AAAA', 'b');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (2, 2, 'AAAA', 'c');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (3, 3, 'BBBB', 'D');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (4, 3, 'BBBB', 'e');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (5, 3, 'bbbb', 'F');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (6, 4, 'cccc', 'g');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (7, 4, 'cccc', 'h');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (8, 4, 'CCCC', 'I');`,
			},
			{
				Statement: `INSERT INTO test_having VALUES (9, 4, 'CCCC', 'j');`,
			},
			{
				Statement: `SELECT b, c FROM test_having
	GROUP BY b, c HAVING count(*) = 1 ORDER BY b, c;`,
				Results: []sql.Row{{1, `XXXX`}, {3, `bbbb`}},
			},
			{
				Statement: `SELECT b, c FROM test_having
	GROUP BY b, c HAVING b = 3 ORDER BY b, c;`,
				Results: []sql.Row{{3, `bbbb`}, {3, `BBBB`}},
			},
			{
				Statement: `SELECT lower(c), count(c) FROM test_having
	GROUP BY lower(c) HAVING count(*) > 2 OR min(a) = max(a)
	ORDER BY lower(c);`,
				Results: []sql.Row{{`bbbb`, 3}, {`cccc`, 4}, {`xxxx`, 1}},
			},
			{
				Statement: `SELECT c, max(a) FROM test_having
	GROUP BY c HAVING count(*) > 2 OR min(a) = max(a)
	ORDER BY c;`,
				Results: []sql.Row{{`bbbb`, 5}, {`XXXX`, 0}},
			},
			{
				Statement: `SELECT min(a), max(a) FROM test_having HAVING min(a) = max(a);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT min(a), max(a) FROM test_having HAVING min(a) < max(a);`,
				Results:   []sql.Row{{0, 9}},
			},
			{
				Statement:   `SELECT a FROM test_having HAVING min(a) < max(a);`,
				ErrorString: `column "test_having.a" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement:   `SELECT 1 AS one FROM test_having HAVING a > 1;`,
				ErrorString: `column "test_having.a" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `SELECT 1 AS one FROM test_having HAVING 1 > 2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT 1 AS one FROM test_having HAVING 1 < 2;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT 1 AS one FROM test_having WHERE 1/a = 1 HAVING 1 < 2;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP TABLE test_having;`,
			},
		},
	})
}
