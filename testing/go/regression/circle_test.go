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

func TestCircle(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_circle)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_circle,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET extra_float_digits = -1;`,
			},
			{
				Statement: `CREATE TABLE CIRCLE_TBL (f1 circle);`,
			},
			{
				Statement: `INSERT INTO CIRCLE_TBL VALUES ('<(5,1),3>');`,
			},
			{
				Statement: `INSERT INTO CIRCLE_TBL VALUES ('((1,2),100)');`,
			},
			{
				Statement: `INSERT INTO CIRCLE_TBL VALUES (' 1 , 3 , 5 ');`,
			},
			{
				Statement: `INSERT INTO CIRCLE_TBL VALUES (' ( ( 1 , 2 ) , 3 ) ');`,
			},
			{
				Statement: `INSERT INTO CIRCLE_TBL VALUES (' ( 100 , 200 ) , 10 ');`,
			},
			{
				Statement: `INSERT INTO CIRCLE_TBL VALUES (' < ( 100 , 1 ) , 115 > ');`,
			},
			{
				Statement: `INSERT INTO CIRCLE_TBL VALUES ('<(3,5),0>');	-- Zero radius`,
			},
			{
				Statement: `INSERT INTO CIRCLE_TBL VALUES ('<(3,5),NaN>');	-- NaN radius`,
			},
			{
				Statement:   `INSERT INTO CIRCLE_TBL VALUES ('<(-100,0),-100>');`,
				ErrorString: `invalid input syntax for type circle: "<(-100,0),-100>"`,
			},
			{
				Statement:   `INSERT INTO CIRCLE_TBL VALUES ('<(100,200),10');`,
				ErrorString: `invalid input syntax for type circle: "<(100,200),10"`,
			},
			{
				Statement:   `INSERT INTO CIRCLE_TBL VALUES ('<(100,200),10> x');`,
				ErrorString: `invalid input syntax for type circle: "<(100,200),10> x"`,
			},
			{
				Statement:   `INSERT INTO CIRCLE_TBL VALUES ('1abc,3,5');`,
				ErrorString: `invalid input syntax for type circle: "1abc,3,5"`,
			},
			{
				Statement:   `INSERT INTO CIRCLE_TBL VALUES ('(3,(1,2),3)');`,
				ErrorString: `invalid input syntax for type circle: "(3,(1,2),3)"`,
			},
			{
				Statement: `SELECT * FROM CIRCLE_TBL;`,
				Results:   []sql.Row{{`<(5,1),3>`}, {`<(1,2),100>`}, {`<(1,3),5>`}, {`<(1,2),3>`}, {`<(100,200),10>`}, {`<(100,1),115>`}, {`<(3,5),0>`}, {`<(3,5),NaN>`}},
			},
			{
				Statement: `SELECT center(f1) AS center
  FROM CIRCLE_TBL;`,
				Results: []sql.Row{{`(5,1)`}, {`(1,2)`}, {`(1,3)`}, {`(1,2)`}, {`(100,200)`}, {`(100,1)`}, {`(3,5)`}, {`(3,5)`}},
			},
			{
				Statement: `SELECT radius(f1) AS radius
  FROM CIRCLE_TBL;`,
				Results: []sql.Row{{3}, {100}, {5}, {3}, {10}, {115}, {0}, {`NaN`}},
			},
			{
				Statement: `SELECT diameter(f1) AS diameter
  FROM CIRCLE_TBL;`,
				Results: []sql.Row{{6}, {200}, {10}, {6}, {20}, {230}, {0}, {`NaN`}},
			},
			{
				Statement: `SELECT f1 FROM CIRCLE_TBL WHERE radius(f1) < 5;`,
				Results:   []sql.Row{{`<(5,1),3>`}, {`<(1,2),3>`}, {`<(3,5),0>`}},
			},
			{
				Statement: `SELECT f1 FROM CIRCLE_TBL WHERE diameter(f1) >= 10;`,
				Results:   []sql.Row{{`<(1,2),100>`}, {`<(1,3),5>`}, {`<(100,200),10>`}, {`<(100,1),115>`}, {`<(3,5),NaN>`}},
			},
			{
				Statement: `SELECT c1.f1 AS one, c2.f1 AS two, (c1.f1 <-> c2.f1) AS distance
  FROM CIRCLE_TBL c1, CIRCLE_TBL c2
  WHERE (c1.f1 < c2.f1) AND ((c1.f1 <-> c2.f1) > 0)
  ORDER BY distance, area(c1.f1), area(c2.f1);`,
				Results: []sql.Row{{`<(3,5),0>`, `<(1,2),3>`, 0.60555127546399}, {`<(3,5),0>`, `<(5,1),3>`, 1.4721359549996}, {`<(100,200),10>`, `<(100,1),115>`, 74}, {`<(100,200),10>`, `<(1,2),100>`, 111.37072977248}, {`<(1,3),5>`, `<(100,200),10>`, 205.4767561445}, {`<(5,1),3>`, `<(100,200),10>`, 207.51303816328}, {`<(3,5),0>`, `<(100,200),10>`, 207.79348015953}, {`<(1,2),3>`, `<(100,200),10>`, 208.37072977248}},
			},
		},
	})
}
