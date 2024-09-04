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

func TestRandom(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_random)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_random,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT count(*) FROM onek;`,
				Results:   []sql.Row{{1000}},
			},
			{
				Statement: `(SELECT unique1 AS random
  FROM onek ORDER BY random() LIMIT 1)
INTERSECT
(SELECT unique1 AS random
  FROM onek ORDER BY random() LIMIT 1)
INTERSECT
(SELECT unique1 AS random
  FROM onek ORDER BY random() LIMIT 1);`,
				Results: []sql.Row{},
			},
			{
				Statement: `CREATE TABLE RANDOM_TBL AS
  SELECT count(*) AS random
  FROM onek WHERE random() < 1.0/10;`,
			},
			{
				Statement: `INSERT INTO RANDOM_TBL (random)
  SELECT count(*)
  FROM onek WHERE random() < 1.0/10;`,
			},
			{
				Statement: `INSERT INTO RANDOM_TBL (random)
  SELECT count(*)
  FROM onek WHERE random() < 1.0/10;`,
			},
			{
				Statement: `INSERT INTO RANDOM_TBL (random)
  SELECT count(*)
  FROM onek WHERE random() < 1.0/10;`,
			},
			{
				Statement: `SELECT random, count(random) FROM RANDOM_TBL
  GROUP BY random HAVING count(random) > 3;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT AVG(random) FROM RANDOM_TBL
  HAVING AVG(random) NOT BETWEEN 80 AND 120;`,
				Results: []sql.Row{},
			},
		},
	})
}
