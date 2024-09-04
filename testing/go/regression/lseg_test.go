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

func TestLseg(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_lseg)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_lseg,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE LSEG_TBL (s lseg);`,
			},
			{
				Statement: `INSERT INTO LSEG_TBL VALUES ('[(1,2),(3,4)]');`,
			},
			{
				Statement: `INSERT INTO LSEG_TBL VALUES ('(0,0),(6,6)');`,
			},
			{
				Statement: `INSERT INTO LSEG_TBL VALUES ('10,-10 ,-3,-4');`,
			},
			{
				Statement: `INSERT INTO LSEG_TBL VALUES ('[-1e6,2e2,3e5, -4e1]');`,
			},
			{
				Statement: `INSERT INTO LSEG_TBL VALUES (lseg(point(11, 22), point(33,44)));`,
			},
			{
				Statement: `INSERT INTO LSEG_TBL VALUES ('[(-10,2),(-10,3)]');	-- vertical`,
			},
			{
				Statement: `INSERT INTO LSEG_TBL VALUES ('[(0,-20),(30,-20)]');	-- horizontal`,
			},
			{
				Statement: `INSERT INTO LSEG_TBL VALUES ('[(NaN,1),(NaN,90)]');	-- NaN`,
			},
			{
				Statement:   `INSERT INTO LSEG_TBL VALUES ('(3asdf,2 ,3,4r2)');`,
				ErrorString: `invalid input syntax for type lseg: "(3asdf,2 ,3,4r2)"`,
			},
			{
				Statement:   `INSERT INTO LSEG_TBL VALUES ('[1,2,3, 4');`,
				ErrorString: `invalid input syntax for type lseg: "[1,2,3, 4"`,
			},
			{
				Statement:   `INSERT INTO LSEG_TBL VALUES ('[(,2),(3,4)]');`,
				ErrorString: `invalid input syntax for type lseg: "[(,2),(3,4)]"`,
			},
			{
				Statement:   `INSERT INTO LSEG_TBL VALUES ('[(1,2),(3,4)');`,
				ErrorString: `invalid input syntax for type lseg: "[(1,2),(3,4)"`,
			},
			{
				Statement: `select * from LSEG_TBL;`,
				Results:   []sql.Row{{`[(1,2),(3,4)]`}, {`[(0,0),(6,6)]`}, {`[(10,-10),(-3,-4)]`}, {`[(-1000000,200),(300000,-40)]`}, {`[(11,22),(33,44)]`}, {`[(-10,2),(-10,3)]`}, {`[(0,-20),(30,-20)]`}, {`[(NaN,1),(NaN,90)]`}},
			},
		},
	})
}
