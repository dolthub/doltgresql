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

func TestBitmapops(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_bitmapops)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_bitmapops,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE bmscantest (a int, b int, t text);`,
			},
			{
				Statement: `INSERT INTO bmscantest
  SELECT (r%53), (r%59), 'foooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo'
  FROM generate_series(1,70000) r;`,
			},
			{
				Statement: `CREATE INDEX i_bmtest_a ON bmscantest(a);`,
			},
			{
				Statement: `CREATE INDEX i_bmtest_b ON bmscantest(b);`,
			},
			{
				Statement: `set enable_indexscan=false;`,
			},
			{
				Statement: `set enable_seqscan=false;`,
			},
			{
				Statement: `set work_mem = 64;`,
			},
			{
				Statement: `SELECT count(*) FROM bmscantest WHERE a = 1 AND b = 1;`,
				Results:   []sql.Row{{23}},
			},
			{
				Statement: `SELECT count(*) FROM bmscantest WHERE a = 1 OR b = 1;`,
				Results:   []sql.Row{{2485}},
			},
			{
				Statement: `DROP TABLE bmscantest;`,
			},
		},
	})
}
