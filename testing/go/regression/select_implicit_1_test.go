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

func TestSelectImplicit1(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_select_implicit_1)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_select_implicit_1,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE test_missing_target (a int, b int, c char(8), d char);`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (0, 1, 'XXXX', 'A');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (1, 2, 'ABAB', 'b');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (2, 2, 'ABAB', 'c');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (3, 3, 'BBBB', 'D');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (4, 3, 'BBBB', 'e');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (5, 3, 'bbbb', 'F');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (6, 4, 'cccc', 'g');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (7, 4, 'cccc', 'h');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (8, 4, 'CCCC', 'I');`,
			},
			{
				Statement: `INSERT INTO test_missing_target VALUES (9, 4, 'CCCC', 'j');`,
			},
			{
				Statement: `SELECT c, count(*) FROM test_missing_target GROUP BY test_missing_target.c ORDER BY c;`,
				Results:   []sql.Row{{`ABAB`, 2}, {`BBBB`, 2}, {`bbbb`, 1}, {`CCCC`, 2}, {`cccc`, 2}, {`XXXX`, 1}},
			},
			{
				Statement: `SELECT count(*) FROM test_missing_target GROUP BY test_missing_target.c ORDER BY c;`,
				Results:   []sql.Row{{2}, {2}, {1}, {2}, {2}, {1}},
			},
			{
				Statement:   `SELECT count(*) FROM test_missing_target GROUP BY a ORDER BY b;`,
				ErrorString: `column "test_missing_target.b" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `SELECT count(*) FROM test_missing_target GROUP BY b ORDER BY b;`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}},
			},
			{
				Statement: `SELECT test_missing_target.b, count(*)
  FROM test_missing_target GROUP BY b ORDER BY b;`,
				Results: []sql.Row{{1, 1}, {2, 2}, {3, 3}, {4, 4}},
			},
			{
				Statement: `SELECT c FROM test_missing_target ORDER BY a;`,
				Results:   []sql.Row{{`XXXX`}, {`ABAB`}, {`ABAB`}, {`BBBB`}, {`BBBB`}, {`bbbb`}, {`cccc`}, {`cccc`}, {`CCCC`}, {`CCCC`}},
			},
			{
				Statement: `SELECT count(*) FROM test_missing_target GROUP BY b ORDER BY b desc;`,
				Results:   []sql.Row{{4}, {3}, {2}, {1}},
			},
			{
				Statement: `SELECT count(*) FROM test_missing_target ORDER BY 1 desc;`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `SELECT c, count(*) FROM test_missing_target GROUP BY 1 ORDER BY 1;`,
				Results:   []sql.Row{{`ABAB`, 2}, {`BBBB`, 2}, {`bbbb`, 1}, {`CCCC`, 2}, {`cccc`, 2}, {`XXXX`, 1}},
			},
			{
				Statement:   `SELECT c, count(*) FROM test_missing_target GROUP BY 3;`,
				ErrorString: `GROUP BY position 3 is not in select list`,
			},
			{
				Statement: `SELECT count(*) FROM test_missing_target x, test_missing_target y
	WHERE x.a = y.a
	GROUP BY b ORDER BY b;`,
				ErrorString: `column reference "b" is ambiguous`,
			},
			{
				Statement: `SELECT a, a FROM test_missing_target
	ORDER BY a;`,
				Results: []sql.Row{{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}},
			},
			{
				Statement: `SELECT a/2, a/2 FROM test_missing_target
	ORDER BY a/2;`,
				Results: []sql.Row{{0, 0}, {0, 0}, {1, 1}, {1, 1}, {2, 2}, {2, 2}, {3, 3}, {3, 3}, {4, 4}, {4, 4}},
			},
			{
				Statement: `SELECT a/2, a/2 FROM test_missing_target
	GROUP BY a/2 ORDER BY a/2;`,
				Results: []sql.Row{{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}},
			},
			{
				Statement: `SELECT x.b, count(*) FROM test_missing_target x, test_missing_target y
	WHERE x.a = y.a
	GROUP BY x.b ORDER BY x.b;`,
				Results: []sql.Row{{1, 1}, {2, 2}, {3, 3}, {4, 4}},
			},
			{
				Statement: `SELECT count(*) FROM test_missing_target x, test_missing_target y
	WHERE x.a = y.a
	GROUP BY x.b ORDER BY x.b;`,
				Results: []sql.Row{{1}, {2}, {3}, {4}},
			},
			{
				Statement: `CREATE TABLE test_missing_target2 AS
SELECT count(*)
FROM test_missing_target x, test_missing_target y
	WHERE x.a = y.a
	GROUP BY x.b ORDER BY x.b;`,
			},
			{
				Statement: `SELECT * FROM test_missing_target2;`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}},
			},
			{
				Statement: `SELECT a%2, count(b) FROM test_missing_target
GROUP BY test_missing_target.a%2
ORDER BY test_missing_target.a%2;`,
				Results: []sql.Row{{0, 5}, {1, 5}},
			},
			{
				Statement: `SELECT count(c) FROM test_missing_target
GROUP BY lower(test_missing_target.c)
ORDER BY lower(test_missing_target.c);`,
				Results: []sql.Row{{2}, {3}, {4}, {1}},
			},
			{
				Statement:   `SELECT count(a) FROM test_missing_target GROUP BY a ORDER BY b;`,
				ErrorString: `column "test_missing_target.b" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `SELECT count(b) FROM test_missing_target GROUP BY b/2 ORDER BY b/2;`,
				Results:   []sql.Row{{1}, {5}, {4}},
			},
			{
				Statement: `SELECT lower(test_missing_target.c), count(c)
  FROM test_missing_target GROUP BY lower(c) ORDER BY lower(c);`,
				Results: []sql.Row{{`abab`, 2}, {`bbbb`, 3}, {`cccc`, 4}, {`xxxx`, 1}},
			},
			{
				Statement: `SELECT a FROM test_missing_target ORDER BY upper(d);`,
				Results:   []sql.Row{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}},
			},
			{
				Statement: `SELECT count(b) FROM test_missing_target
	GROUP BY (b + 1) / 2 ORDER BY (b + 1) / 2 desc;`,
				Results: []sql.Row{{7}, {3}},
			},
			{
				Statement: `SELECT count(x.a) FROM test_missing_target x, test_missing_target y
	WHERE x.a = y.a
	GROUP BY b/2 ORDER BY b/2;`,
				ErrorString: `column reference "b" is ambiguous`,
			},
			{
				Statement: `SELECT x.b/2, count(x.b) FROM test_missing_target x, test_missing_target y
	WHERE x.a = y.a
	GROUP BY x.b/2 ORDER BY x.b/2;`,
				Results: []sql.Row{{0, 1}, {1, 5}, {2, 4}},
			},
			{
				Statement: `SELECT count(b) FROM test_missing_target x, test_missing_target y
	WHERE x.a = y.a
	GROUP BY x.b/2;`,
				ErrorString: `column reference "b" is ambiguous`,
			},
			{
				Statement: `CREATE TABLE test_missing_target3 AS
SELECT count(x.b)
FROM test_missing_target x, test_missing_target y
	WHERE x.a = y.a
	GROUP BY x.b/2 ORDER BY x.b/2;`,
			},
			{
				Statement: `SELECT * FROM test_missing_target3;`,
				Results:   []sql.Row{{1}, {5}, {4}},
			},
			{
				Statement: `DROP TABLE test_missing_target;`,
			},
			{
				Statement: `DROP TABLE test_missing_target2;`,
			},
			{
				Statement: `DROP TABLE test_missing_target3;`,
			},
		},
	})
}
