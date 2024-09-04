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

func TestPath(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_path)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_path,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE PATH_TBL (f1 path);`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES ('[(1,2),(3,4)]');`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES (' ( ( 1 , 2 ) , ( 3 , 4 ) ) ');`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES ('[ (0,0),(3,0),(4,5),(1,6) ]');`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES ('((1,2) ,(3,4 ))');`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES ('1,2 ,3,4 ');`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES (' [1,2,3, 4] ');`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES ('((10,20))');	-- Only one point`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES ('[ 11,12,13,14 ]');`,
			},
			{
				Statement: `INSERT INTO PATH_TBL VALUES ('( 11,12,13,14) ');`,
			},
			{
				Statement:   `INSERT INTO PATH_TBL VALUES ('[]');`,
				ErrorString: `invalid input syntax for type path: "[]"`,
			},
			{
				Statement:   `INSERT INTO PATH_TBL VALUES ('[(,2),(3,4)]');`,
				ErrorString: `invalid input syntax for type path: "[(,2),(3,4)]"`,
			},
			{
				Statement:   `INSERT INTO PATH_TBL VALUES ('[(1,2),(3,4)');`,
				ErrorString: `invalid input syntax for type path: "[(1,2),(3,4)"`,
			},
			{
				Statement:   `INSERT INTO PATH_TBL VALUES ('(1,2,3,4');`,
				ErrorString: `invalid input syntax for type path: "(1,2,3,4"`,
			},
			{
				Statement:   `INSERT INTO PATH_TBL VALUES ('(1,2),(3,4)]');`,
				ErrorString: `invalid input syntax for type path: "(1,2),(3,4)]"`,
			},
			{
				Statement: `SELECT f1 AS open_path FROM PATH_TBL WHERE isopen(f1);`,
				Results:   []sql.Row{{`[(1,2),(3,4)]`}, {`[(0,0),(3,0),(4,5),(1,6)]`}, {`[(1,2),(3,4)]`}, {`[(11,12),(13,14)]`}},
			},
			{
				Statement: `SELECT f1 AS closed_path FROM PATH_TBL WHERE isclosed(f1);`,
				Results:   []sql.Row{{`((1,2),(3,4))`}, {`((1,2),(3,4))`}, {`((1,2),(3,4))`}, {`((10,20))`}, {`((11,12),(13,14))`}},
			},
			{
				Statement: `SELECT pclose(f1) AS closed_path FROM PATH_TBL;`,
				Results:   []sql.Row{{`((1,2),(3,4))`}, {`((1,2),(3,4))`}, {`((0,0),(3,0),(4,5),(1,6))`}, {`((1,2),(3,4))`}, {`((1,2),(3,4))`}, {`((1,2),(3,4))`}, {`((10,20))`}, {`((11,12),(13,14))`}, {`((11,12),(13,14))`}},
			},
			{
				Statement: `SELECT popen(f1) AS open_path FROM PATH_TBL;`,
				Results:   []sql.Row{{`[(1,2),(3,4)]`}, {`[(1,2),(3,4)]`}, {`[(0,0),(3,0),(4,5),(1,6)]`}, {`[(1,2),(3,4)]`}, {`[(1,2),(3,4)]`}, {`[(1,2),(3,4)]`}, {`[(10,20)]`}, {`[(11,12),(13,14)]`}, {`[(11,12),(13,14)]`}},
			},
		},
	})
}
