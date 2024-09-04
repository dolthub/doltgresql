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

func TestPortalsP2(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_portals_p2)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_portals_p2,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DECLARE foo13 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 50;`,
			},
			{
				Statement: `DECLARE foo14 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 51;`,
			},
			{
				Statement: `DECLARE foo15 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 52;`,
			},
			{
				Statement: `DECLARE foo16 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 53;`,
			},
			{
				Statement: `DECLARE foo17 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 54;`,
			},
			{
				Statement: `DECLARE foo18 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 55;`,
			},
			{
				Statement: `DECLARE foo19 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 56;`,
			},
			{
				Statement: `DECLARE foo20 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 57;`,
			},
			{
				Statement: `DECLARE foo21 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 58;`,
			},
			{
				Statement: `DECLARE foo22 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 59;`,
			},
			{
				Statement: `DECLARE foo23 CURSOR FOR
   SELECT * FROM onek WHERE unique1 = 60;`,
			},
			{
				Statement: `DECLARE foo24 CURSOR FOR
   SELECT * FROM onek2 WHERE unique1 = 50;`,
			},
			{
				Statement: `DECLARE foo25 CURSOR FOR
   SELECT * FROM onek2 WHERE unique1 = 60;`,
			},
			{
				Statement: `FETCH all in foo13;`,
				Results:   []sql.Row{{50, 253, 0, 2, 0, 10, 0, 50, 50, 50, 50, 0, 1, `YBAAAA`, `TJAAAA`, `HHHHxx`}},
			},
			{
				Statement: `FETCH all in foo14;`,
				Results:   []sql.Row{{51, 76, 1, 3, 1, 11, 1, 51, 51, 51, 51, 2, 3, `ZBAAAA`, `YCAAAA`, `AAAAxx`}},
			},
			{
				Statement: `FETCH all in foo15;`,
				Results:   []sql.Row{{52, 985, 0, 0, 2, 12, 2, 52, 52, 52, 52, 4, 5, `ACAAAA`, `XLBAAA`, `HHHHxx`}},
			},
			{
				Statement: `FETCH all in foo16;`,
				Results:   []sql.Row{{53, 196, 1, 1, 3, 13, 3, 53, 53, 53, 53, 6, 7, `BCAAAA`, `OHAAAA`, `AAAAxx`}},
			},
			{
				Statement: `FETCH all in foo17;`,
				Results:   []sql.Row{{54, 356, 0, 2, 4, 14, 4, 54, 54, 54, 54, 8, 9, `CCAAAA`, `SNAAAA`, `AAAAxx`}},
			},
			{
				Statement: `FETCH all in foo18;`,
				Results:   []sql.Row{{55, 627, 1, 3, 5, 15, 5, 55, 55, 55, 55, 10, 11, `DCAAAA`, `DYAAAA`, `VVVVxx`}},
			},
			{
				Statement: `FETCH all in foo19;`,
				Results:   []sql.Row{{56, 54, 0, 0, 6, 16, 6, 56, 56, 56, 56, 12, 13, `ECAAAA`, `CCAAAA`, `OOOOxx`}},
			},
			{
				Statement: `FETCH all in foo20;`,
				Results:   []sql.Row{{57, 942, 1, 1, 7, 17, 7, 57, 57, 57, 57, 14, 15, `FCAAAA`, `GKBAAA`, `OOOOxx`}},
			},
			{
				Statement: `FETCH all in foo21;`,
				Results:   []sql.Row{{58, 114, 0, 2, 8, 18, 8, 58, 58, 58, 58, 16, 17, `GCAAAA`, `KEAAAA`, `OOOOxx`}},
			},
			{
				Statement: `FETCH all in foo22;`,
				Results:   []sql.Row{{59, 593, 1, 3, 9, 19, 9, 59, 59, 59, 59, 18, 19, `HCAAAA`, `VWAAAA`, `HHHHxx`}},
			},
			{
				Statement: `FETCH all in foo23;`,
				Results:   []sql.Row{{60, 483, 0, 0, 0, 0, 0, 60, 60, 60, 60, 0, 1, `ICAAAA`, `PSAAAA`, `VVVVxx`}},
			},
			{
				Statement: `FETCH all in foo24;`,
				Results:   []sql.Row{{50, 253, 0, 2, 0, 10, 0, 50, 50, 50, 50, 0, 1, `YBAAAA`, `TJAAAA`, `HHHHxx`}},
			},
			{
				Statement: `FETCH all in foo25;`,
				Results:   []sql.Row{{60, 483, 0, 0, 0, 0, 0, 60, 60, 60, 60, 0, 1, `ICAAAA`, `PSAAAA`, `VVVVxx`}},
			},
			{
				Statement: `CLOSE foo13;`,
			},
			{
				Statement: `CLOSE foo14;`,
			},
			{
				Statement: `CLOSE foo15;`,
			},
			{
				Statement: `CLOSE foo16;`,
			},
			{
				Statement: `CLOSE foo17;`,
			},
			{
				Statement: `CLOSE foo18;`,
			},
			{
				Statement: `CLOSE foo19;`,
			},
			{
				Statement: `CLOSE foo20;`,
			},
			{
				Statement: `CLOSE foo21;`,
			},
			{
				Statement: `CLOSE foo22;`,
			},
			{
				Statement: `CLOSE foo23;`,
			},
			{
				Statement: `CLOSE foo24;`,
			},
			{
				Statement: `CLOSE foo25;`,
			},
			{
				Statement: `END;`,
			},
		},
	})
}
