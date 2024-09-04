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

func TestSelectDistinctOn(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_select_distinct_on)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_select_distinct_on,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT DISTINCT ON (string4) string4, two, ten
   FROM onek
   ORDER BY string4 using <, two using >, ten using <;`,
				Results: []sql.Row{{`AAAAxx`, 1, 1}, {`HHHHxx`, 1, 1}, {`OOOOxx`, 1, 1}, {`VVVVxx`, 1, 1}},
			},
			{
				Statement: `SELECT DISTINCT ON (string4, ten) string4, two, ten
   FROM onek
   ORDER BY string4 using <, two using <, ten using <;`,
				ErrorString: `SELECT DISTINCT ON expressions must match initial ORDER BY expressions`,
			},
			{
				Statement: `SELECT DISTINCT ON (string4, ten) string4, ten, two
   FROM onek
   ORDER BY string4 using <, ten using >, two using <;`,
				Results: []sql.Row{{`AAAAxx`, 9, 1}, {`AAAAxx`, 8, 0}, {`AAAAxx`, 7, 1}, {`AAAAxx`, 6, 0}, {`AAAAxx`, 5, 1}, {`AAAAxx`, 4, 0}, {`AAAAxx`, 3, 1}, {`AAAAxx`, 2, 0}, {`AAAAxx`, 1, 1}, {`AAAAxx`, 0, 0}, {`HHHHxx`, 9, 1}, {`HHHHxx`, 8, 0}, {`HHHHxx`, 7, 1}, {`HHHHxx`, 6, 0}, {`HHHHxx`, 5, 1}, {`HHHHxx`, 4, 0}, {`HHHHxx`, 3, 1}, {`HHHHxx`, 2, 0}, {`HHHHxx`, 1, 1}, {`HHHHxx`, 0, 0}, {`OOOOxx`, 9, 1}, {`OOOOxx`, 8, 0}, {`OOOOxx`, 7, 1}, {`OOOOxx`, 6, 0}, {`OOOOxx`, 5, 1}, {`OOOOxx`, 4, 0}, {`OOOOxx`, 3, 1}, {`OOOOxx`, 2, 0}, {`OOOOxx`, 1, 1}, {`OOOOxx`, 0, 0}, {`VVVVxx`, 9, 1}, {`VVVVxx`, 8, 0}, {`VVVVxx`, 7, 1}, {`VVVVxx`, 6, 0}, {`VVVVxx`, 5, 1}, {`VVVVxx`, 4, 0}, {`VVVVxx`, 3, 1}, {`VVVVxx`, 2, 0}, {`VVVVxx`, 1, 1}, {`VVVVxx`, 0, 0}},
			},
			{
				Statement: `select distinct on (1) floor(random()) as r, f1 from int4_tbl order by 1,2;`,
				Results:   []sql.Row{{0, -2147483647}},
			},
		},
	})
}
