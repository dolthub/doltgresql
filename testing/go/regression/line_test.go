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

func TestLine(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_line)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_line,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE LINE_TBL (s line);`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES ('{0,-1,5}');	-- A == 0`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES ('{1,0,5}');	-- B == 0`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES ('{0,3,0}');	-- A == C == 0`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES (' (0,0), (6,6)');`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES ('10,-10 ,-5,-4');`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES ('[-1e6,2e2,3e5, -4e1]');`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES ('{3,NaN,5}');`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES ('{NaN,NaN,NaN}');`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES ('[(1,3),(2,3)]');`,
			},
			{
				Statement: `INSERT INTO LINE_TBL VALUES (line(point '(3,1)', point '(3,2)'));`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('{}');`,
				ErrorString: `invalid input syntax for type line: "{}"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('{0');`,
				ErrorString: `invalid input syntax for type line: "{0"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('{0,0}');`,
				ErrorString: `invalid input syntax for type line: "{0,0}"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('{0,0,1');`,
				ErrorString: `invalid input syntax for type line: "{0,0,1"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('{0,0,1}');`,
				ErrorString: `invalid line specification: A and B cannot both be zero`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('{0,0,1} x');`,
				ErrorString: `invalid input syntax for type line: "{0,0,1} x"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('(3asdf,2 ,3,4r2)');`,
				ErrorString: `invalid input syntax for type line: "(3asdf,2 ,3,4r2)"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('[1,2,3, 4');`,
				ErrorString: `invalid input syntax for type line: "[1,2,3, 4"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('[(,2),(3,4)]');`,
				ErrorString: `invalid input syntax for type line: "[(,2),(3,4)]"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('[(1,2),(3,4)');`,
				ErrorString: `invalid input syntax for type line: "[(1,2),(3,4)"`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES ('[(1,2),(1,2)]');`,
				ErrorString: `invalid line specification: must be two distinct points`,
			},
			{
				Statement:   `INSERT INTO LINE_TBL VALUES (line(point '(1,0)', point '(1,0)'));`,
				ErrorString: `invalid line specification: must be two distinct points`,
			},
			{
				Statement: `select * from LINE_TBL;`,
				Results:   []sql.Row{{`{0,-1,5}`}, {`{1,0,5}`}, {`{0,3,0}`}, {`{1,-1,0}`}, {`{-0.4,-1,-6}`}, {`{-0.0001846153846153846,-1,15.384615384615387}`}, {`{3,NaN,5}`}, {`{NaN,NaN,NaN}`}, {`{0,-1,3}`}, {`{-1,0,3}`}},
			},
			{
				Statement: `select '{nan, 1, nan}'::line = '{nan, 1, nan}'::line as true,
	   '{nan, 1, nan}'::line = '{nan, 2, nan}'::line as false;`,
				Results: []sql.Row{{true, false}},
			},
		},
	})
}
