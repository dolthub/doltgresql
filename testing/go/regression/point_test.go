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

func TestPoint(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_point)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_point,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET extra_float_digits = 0;`,
			},
			{
				Statement:   `INSERT INTO POINT_TBL(f1) VALUES ('asdfasdf');`,
				ErrorString: `invalid input syntax for type point: "asdfasdf"`,
			},
			{
				Statement:   `INSERT INTO POINT_TBL(f1) VALUES ('(10.0 10.0)');`,
				ErrorString: `invalid input syntax for type point: "(10.0 10.0)"`,
			},
			{
				Statement:   `INSERT INTO POINT_TBL(f1) VALUES ('(10.0, 10.0) x');`,
				ErrorString: `invalid input syntax for type point: "(10.0, 10.0) x"`,
			},
			{
				Statement:   `INSERT INTO POINT_TBL(f1) VALUES ('(10.0,10.0');`,
				ErrorString: `invalid input syntax for type point: "(10.0,10.0"`,
			},
			{
				Statement:   `INSERT INTO POINT_TBL(f1) VALUES ('(10.0, 1e+500)');	-- Out of range`,
				ErrorString: `"1e+500" is out of range for type double precision`,
			},
			{
				Statement: `SELECT * FROM POINT_TBL;`,
				Results:   []sql.Row{{`(0,0)`}, {`(-10,0)`}, {`(-3,4)`}, {`(5.1,34.5)`}, {`(-5,-12)`}, {`(1e-300,-1e-300)`}, {`(1e+300,Infinity)`}, {`(Infinity,1e+300)`}, {`(NaN,NaN)`}, {`(10,10)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p WHERE p.f1 << '(0.0, 0.0)';`,
				Results:   []sql.Row{{`(-10,0)`}, {`(-3,4)`}, {`(-5,-12)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p WHERE '(0.0,0.0)' >> p.f1;`,
				Results:   []sql.Row{{`(-10,0)`}, {`(-3,4)`}, {`(-5,-12)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p WHERE '(0.0,0.0)' |>> p.f1;`,
				Results:   []sql.Row{{`(-5,-12)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p WHERE p.f1 <<| '(0.0, 0.0)';`,
				Results:   []sql.Row{{`(-5,-12)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p WHERE p.f1 ~= '(5.1, 34.5)';`,
				Results:   []sql.Row{{`(5.1,34.5)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p
   WHERE p.f1 <@ box '(0,0,100,100)';`,
				Results: []sql.Row{{`(0,0)`}, {`(5.1,34.5)`}, {`(10,10)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p
   WHERE box '(0,0,100,100)' @> p.f1;`,
				Results: []sql.Row{{`(0,0)`}, {`(5.1,34.5)`}, {`(10,10)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p
   WHERE not p.f1 <@ box '(0,0,100,100)';`,
				Results: []sql.Row{{`(-10,0)`}, {`(-3,4)`}, {`(-5,-12)`}, {`(1e-300,-1e-300)`}, {`(1e+300,Infinity)`}, {`(Infinity,1e+300)`}, {`(NaN,NaN)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p
   WHERE p.f1 <@ path '[(0,0),(-10,0),(-10,10)]';`,
				Results: []sql.Row{{`(0,0)`}, {`(-10,0)`}, {`(1e-300,-1e-300)`}},
			},
			{
				Statement: `SELECT p.* FROM POINT_TBL p
   WHERE not box '(0,0,100,100)' @> p.f1;`,
				Results: []sql.Row{{`(-10,0)`}, {`(-3,4)`}, {`(-5,-12)`}, {`(1e-300,-1e-300)`}, {`(1e+300,Infinity)`}, {`(Infinity,1e+300)`}, {`(NaN,NaN)`}},
			},
			{
				Statement: `SELECT p.f1, p.f1 <-> point '(0,0)' AS dist
   FROM POINT_TBL p
   ORDER BY dist;`,
				Results: []sql.Row{{`(0,0)`, 0}, {`(1e-300,-1e-300)`, 1.4142135623731e-300}, {`(-3,4)`, 5}, {`(-10,0)`, 10}, {`(-5,-12)`, 13}, {`(10,10)`, 14.142135623731}, {`(5.1,34.5)`, 34.8749193547455}, {`(1e+300,Infinity)`, `Infinity`}, {`(Infinity,1e+300)`, `Infinity`}, {`(NaN,NaN)`, `NaN`}},
			},
			{
				Statement: `SELECT p1.f1 AS point1, p2.f1 AS point2, p1.f1 <-> p2.f1 AS dist
   FROM POINT_TBL p1, POINT_TBL p2
   ORDER BY dist, p1.f1[0], p2.f1[0];`,
				Results: []sql.Row{{`(-10,0)`, `(-10,0)`, 0}, {`(-5,-12)`, `(-5,-12)`, 0}, {`(-3,4)`, `(-3,4)`, 0}, {`(0,0)`, `(0,0)`, 0}, {`(1e-300,-1e-300)`, `(1e-300,-1e-300)`, 0}, {`(5.1,34.5)`, `(5.1,34.5)`, 0}, {`(10,10)`, `(10,10)`, 0}, {`(0,0)`, `(1e-300,-1e-300)`, 1.4142135623731e-300}, {`(1e-300,-1e-300)`, `(0,0)`, 1.4142135623731e-300}, {`(-3,4)`, `(0,0)`, 5}, {`(-3,4)`, `(1e-300,-1e-300)`, 5}, {`(0,0)`, `(-3,4)`, 5}, {`(1e-300,-1e-300)`, `(-3,4)`, 5}, {`(-10,0)`, `(-3,4)`, 8.06225774829855}, {`(-3,4)`, `(-10,0)`, 8.06225774829855}, {`(-10,0)`, `(0,0)`, 10}, {`(-10,0)`, `(1e-300,-1e-300)`, 10}, {`(0,0)`, `(-10,0)`, 10}, {`(1e-300,-1e-300)`, `(-10,0)`, 10}, {`(-10,0)`, `(-5,-12)`, 13}, {`(-5,-12)`, `(-10,0)`, 13}, {`(-5,-12)`, `(0,0)`, 13}, {`(-5,-12)`, `(1e-300,-1e-300)`, 13}, {`(0,0)`, `(-5,-12)`, 13}, {`(1e-300,-1e-300)`, `(-5,-12)`, 13}, {`(0,0)`, `(10,10)`, 14.142135623731}, {`(1e-300,-1e-300)`, `(10,10)`, 14.142135623731}, {`(10,10)`, `(0,0)`, 14.142135623731}, {`(10,10)`, `(1e-300,-1e-300)`, 14.142135623731}, {`(-3,4)`, `(10,10)`, 14.3178210632764}, {`(10,10)`, `(-3,4)`, 14.3178210632764}, {`(-5,-12)`, `(-3,4)`, 16.1245154965971}, {`(-3,4)`, `(-5,-12)`, 16.1245154965971}, {`(-10,0)`, `(10,10)`, 22.3606797749979}, {`(10,10)`, `(-10,0)`, 22.3606797749979}, {`(5.1,34.5)`, `(10,10)`, 24.9851956166046}, {`(10,10)`, `(5.1,34.5)`, 24.9851956166046}, {`(-5,-12)`, `(10,10)`, 26.6270539113887}, {`(10,10)`, `(-5,-12)`, 26.6270539113887}, {`(-3,4)`, `(5.1,34.5)`, 31.5572495632937}, {`(5.1,34.5)`, `(-3,4)`, 31.5572495632937}, {`(0,0)`, `(5.1,34.5)`, 34.8749193547455}, {`(1e-300,-1e-300)`, `(5.1,34.5)`, 34.8749193547455}, {`(5.1,34.5)`, `(0,0)`, 34.8749193547455}, {`(5.1,34.5)`, `(1e-300,-1e-300)`, 34.8749193547455}, {`(-10,0)`, `(5.1,34.5)`, 37.6597928831267}, {`(5.1,34.5)`, `(-10,0)`, 37.6597928831267}, {`(-5,-12)`, `(5.1,34.5)`, 47.5842410888311}, {`(5.1,34.5)`, `(-5,-12)`, 47.5842410888311}, {`(-10,0)`, `(1e+300,Infinity)`, `Infinity`}, {`(-10,0)`, `(Infinity,1e+300)`, `Infinity`}, {`(-5,-12)`, `(1e+300,Infinity)`, `Infinity`}, {`(-5,-12)`, `(Infinity,1e+300)`, `Infinity`}, {`(-3,4)`, `(1e+300,Infinity)`, `Infinity`}, {`(-3,4)`, `(Infinity,1e+300)`, `Infinity`}, {`(0,0)`, `(1e+300,Infinity)`, `Infinity`}, {`(0,0)`, `(Infinity,1e+300)`, `Infinity`}, {`(1e-300,-1e-300)`, `(1e+300,Infinity)`, `Infinity`}, {`(1e-300,-1e-300)`, `(Infinity,1e+300)`, `Infinity`}, {`(5.1,34.5)`, `(1e+300,Infinity)`, `Infinity`}, {`(5.1,34.5)`, `(Infinity,1e+300)`, `Infinity`}, {`(10,10)`, `(1e+300,Infinity)`, `Infinity`}, {`(10,10)`, `(Infinity,1e+300)`, `Infinity`}, {`(1e+300,Infinity)`, `(-10,0)`, `Infinity`}, {`(1e+300,Infinity)`, `(-5,-12)`, `Infinity`}, {`(1e+300,Infinity)`, `(-3,4)`, `Infinity`}, {`(1e+300,Infinity)`, `(0,0)`, `Infinity`}, {`(1e+300,Infinity)`, `(1e-300,-1e-300)`, `Infinity`}, {`(1e+300,Infinity)`, `(5.1,34.5)`, `Infinity`}, {`(1e+300,Infinity)`, `(10,10)`, `Infinity`}, {`(1e+300,Infinity)`, `(Infinity,1e+300)`, `Infinity`}, {`(Infinity,1e+300)`, `(-10,0)`, `Infinity`}, {`(Infinity,1e+300)`, `(-5,-12)`, `Infinity`}, {`(Infinity,1e+300)`, `(-3,4)`, `Infinity`}, {`(Infinity,1e+300)`, `(0,0)`, `Infinity`}, {`(Infinity,1e+300)`, `(1e-300,-1e-300)`, `Infinity`}, {`(Infinity,1e+300)`, `(5.1,34.5)`, `Infinity`}, {`(Infinity,1e+300)`, `(10,10)`, `Infinity`}, {`(Infinity,1e+300)`, `(1e+300,Infinity)`, `Infinity`}, {`(-10,0)`, `(NaN,NaN)`, `NaN`}, {`(-5,-12)`, `(NaN,NaN)`, `NaN`}, {`(-3,4)`, `(NaN,NaN)`, `NaN`}, {`(0,0)`, `(NaN,NaN)`, `NaN`}, {`(1e-300,-1e-300)`, `(NaN,NaN)`, `NaN`}, {`(5.1,34.5)`, `(NaN,NaN)`, `NaN`}, {`(10,10)`, `(NaN,NaN)`, `NaN`}, {`(1e+300,Infinity)`, `(1e+300,Infinity)`, `NaN`}, {`(1e+300,Infinity)`, `(NaN,NaN)`, `NaN`}, {`(Infinity,1e+300)`, `(Infinity,1e+300)`, `NaN`}, {`(Infinity,1e+300)`, `(NaN,NaN)`, `NaN`}, {`(NaN,NaN)`, `(-10,0)`, `NaN`}, {`(NaN,NaN)`, `(-5,-12)`, `NaN`}, {`(NaN,NaN)`, `(-3,4)`, `NaN`}, {`(NaN,NaN)`, `(0,0)`, `NaN`}, {`(NaN,NaN)`, `(1e-300,-1e-300)`, `NaN`}, {`(NaN,NaN)`, `(5.1,34.5)`, `NaN`}, {`(NaN,NaN)`, `(10,10)`, `NaN`}, {`(NaN,NaN)`, `(1e+300,Infinity)`, `NaN`}, {`(NaN,NaN)`, `(Infinity,1e+300)`, `NaN`}, {`(NaN,NaN)`, `(NaN,NaN)`, `NaN`}},
			},
			{
				Statement: `SELECT p1.f1 AS point1, p2.f1 AS point2
   FROM POINT_TBL p1, POINT_TBL p2
   WHERE (p1.f1 <-> p2.f1) > 3;`,
				Results: []sql.Row{{`(0,0)`, `(-10,0)`}, {`(0,0)`, `(-3,4)`}, {`(0,0)`, `(5.1,34.5)`}, {`(0,0)`, `(-5,-12)`}, {`(0,0)`, `(1e+300,Infinity)`}, {`(0,0)`, `(Infinity,1e+300)`}, {`(0,0)`, `(NaN,NaN)`}, {`(0,0)`, `(10,10)`}, {`(-10,0)`, `(0,0)`}, {`(-10,0)`, `(-3,4)`}, {`(-10,0)`, `(5.1,34.5)`}, {`(-10,0)`, `(-5,-12)`}, {`(-10,0)`, `(1e-300,-1e-300)`}, {`(-10,0)`, `(1e+300,Infinity)`}, {`(-10,0)`, `(Infinity,1e+300)`}, {`(-10,0)`, `(NaN,NaN)`}, {`(-10,0)`, `(10,10)`}, {`(-3,4)`, `(0,0)`}, {`(-3,4)`, `(-10,0)`}, {`(-3,4)`, `(5.1,34.5)`}, {`(-3,4)`, `(-5,-12)`}, {`(-3,4)`, `(1e-300,-1e-300)`}, {`(-3,4)`, `(1e+300,Infinity)`}, {`(-3,4)`, `(Infinity,1e+300)`}, {`(-3,4)`, `(NaN,NaN)`}, {`(-3,4)`, `(10,10)`}, {`(5.1,34.5)`, `(0,0)`}, {`(5.1,34.5)`, `(-10,0)`}, {`(5.1,34.5)`, `(-3,4)`}, {`(5.1,34.5)`, `(-5,-12)`}, {`(5.1,34.5)`, `(1e-300,-1e-300)`}, {`(5.1,34.5)`, `(1e+300,Infinity)`}, {`(5.1,34.5)`, `(Infinity,1e+300)`}, {`(5.1,34.5)`, `(NaN,NaN)`}, {`(5.1,34.5)`, `(10,10)`}, {`(-5,-12)`, `(0,0)`}, {`(-5,-12)`, `(-10,0)`}, {`(-5,-12)`, `(-3,4)`}, {`(-5,-12)`, `(5.1,34.5)`}, {`(-5,-12)`, `(1e-300,-1e-300)`}, {`(-5,-12)`, `(1e+300,Infinity)`}, {`(-5,-12)`, `(Infinity,1e+300)`}, {`(-5,-12)`, `(NaN,NaN)`}, {`(-5,-12)`, `(10,10)`}, {`(1e-300,-1e-300)`, `(-10,0)`}, {`(1e-300,-1e-300)`, `(-3,4)`}, {`(1e-300,-1e-300)`, `(5.1,34.5)`}, {`(1e-300,-1e-300)`, `(-5,-12)`}, {`(1e-300,-1e-300)`, `(1e+300,Infinity)`}, {`(1e-300,-1e-300)`, `(Infinity,1e+300)`}, {`(1e-300,-1e-300)`, `(NaN,NaN)`}, {`(1e-300,-1e-300)`, `(10,10)`}, {`(1e+300,Infinity)`, `(0,0)`}, {`(1e+300,Infinity)`, `(-10,0)`}, {`(1e+300,Infinity)`, `(-3,4)`}, {`(1e+300,Infinity)`, `(5.1,34.5)`}, {`(1e+300,Infinity)`, `(-5,-12)`}, {`(1e+300,Infinity)`, `(1e-300,-1e-300)`}, {`(1e+300,Infinity)`, `(1e+300,Infinity)`}, {`(1e+300,Infinity)`, `(Infinity,1e+300)`}, {`(1e+300,Infinity)`, `(NaN,NaN)`}, {`(1e+300,Infinity)`, `(10,10)`}, {`(Infinity,1e+300)`, `(0,0)`}, {`(Infinity,1e+300)`, `(-10,0)`}, {`(Infinity,1e+300)`, `(-3,4)`}, {`(Infinity,1e+300)`, `(5.1,34.5)`}, {`(Infinity,1e+300)`, `(-5,-12)`}, {`(Infinity,1e+300)`, `(1e-300,-1e-300)`}, {`(Infinity,1e+300)`, `(1e+300,Infinity)`}, {`(Infinity,1e+300)`, `(Infinity,1e+300)`}, {`(Infinity,1e+300)`, `(NaN,NaN)`}, {`(Infinity,1e+300)`, `(10,10)`}, {`(NaN,NaN)`, `(0,0)`}, {`(NaN,NaN)`, `(-10,0)`}, {`(NaN,NaN)`, `(-3,4)`}, {`(NaN,NaN)`, `(5.1,34.5)`}, {`(NaN,NaN)`, `(-5,-12)`}, {`(NaN,NaN)`, `(1e-300,-1e-300)`}, {`(NaN,NaN)`, `(1e+300,Infinity)`}, {`(NaN,NaN)`, `(Infinity,1e+300)`}, {`(NaN,NaN)`, `(NaN,NaN)`}, {`(NaN,NaN)`, `(10,10)`}, {`(10,10)`, `(0,0)`}, {`(10,10)`, `(-10,0)`}, {`(10,10)`, `(-3,4)`}, {`(10,10)`, `(5.1,34.5)`}, {`(10,10)`, `(-5,-12)`}, {`(10,10)`, `(1e-300,-1e-300)`}, {`(10,10)`, `(1e+300,Infinity)`}, {`(10,10)`, `(Infinity,1e+300)`}, {`(10,10)`, `(NaN,NaN)`}},
			},
			{
				Statement: `SELECT p1.f1 AS point1, p2.f1 AS point2, (p1.f1 <-> p2.f1) AS distance
   FROM POINT_TBL p1, POINT_TBL p2
   WHERE (p1.f1 <-> p2.f1) > 3 and p1.f1 << p2.f1
   ORDER BY distance, p1.f1[0], p2.f1[0];`,
				Results: []sql.Row{{`(-3,4)`, `(0,0)`, 5}, {`(-3,4)`, `(1e-300,-1e-300)`, 5}, {`(-10,0)`, `(-3,4)`, 8.06225774829855}, {`(-10,0)`, `(0,0)`, 10}, {`(-10,0)`, `(1e-300,-1e-300)`, 10}, {`(-10,0)`, `(-5,-12)`, 13}, {`(-5,-12)`, `(0,0)`, 13}, {`(-5,-12)`, `(1e-300,-1e-300)`, 13}, {`(0,0)`, `(10,10)`, 14.142135623731}, {`(1e-300,-1e-300)`, `(10,10)`, 14.142135623731}, {`(-3,4)`, `(10,10)`, 14.3178210632764}, {`(-5,-12)`, `(-3,4)`, 16.1245154965971}, {`(-10,0)`, `(10,10)`, 22.3606797749979}, {`(5.1,34.5)`, `(10,10)`, 24.9851956166046}, {`(-5,-12)`, `(10,10)`, 26.6270539113887}, {`(-3,4)`, `(5.1,34.5)`, 31.5572495632937}, {`(0,0)`, `(5.1,34.5)`, 34.8749193547455}, {`(1e-300,-1e-300)`, `(5.1,34.5)`, 34.8749193547455}, {`(-10,0)`, `(5.1,34.5)`, 37.6597928831267}, {`(-5,-12)`, `(5.1,34.5)`, 47.5842410888311}, {`(-10,0)`, `(1e+300,Infinity)`, `Infinity`}, {`(-10,0)`, `(Infinity,1e+300)`, `Infinity`}, {`(-5,-12)`, `(1e+300,Infinity)`, `Infinity`}, {`(-5,-12)`, `(Infinity,1e+300)`, `Infinity`}, {`(-3,4)`, `(1e+300,Infinity)`, `Infinity`}, {`(-3,4)`, `(Infinity,1e+300)`, `Infinity`}, {`(0,0)`, `(1e+300,Infinity)`, `Infinity`}, {`(0,0)`, `(Infinity,1e+300)`, `Infinity`}, {`(1e-300,-1e-300)`, `(1e+300,Infinity)`, `Infinity`}, {`(1e-300,-1e-300)`, `(Infinity,1e+300)`, `Infinity`}, {`(5.1,34.5)`, `(1e+300,Infinity)`, `Infinity`}, {`(5.1,34.5)`, `(Infinity,1e+300)`, `Infinity`}, {`(10,10)`, `(1e+300,Infinity)`, `Infinity`}, {`(10,10)`, `(Infinity,1e+300)`, `Infinity`}, {`(1e+300,Infinity)`, `(Infinity,1e+300)`, `Infinity`}},
			},
			{
				Statement: `SELECT p1.f1 AS point1, p2.f1 AS point2, (p1.f1 <-> p2.f1) AS distance
   FROM POINT_TBL p1, POINT_TBL p2
   WHERE (p1.f1 <-> p2.f1) > 3 and p1.f1 << p2.f1 and p1.f1 |>> p2.f1
   ORDER BY distance;`,
				Results: []sql.Row{{`(-3,4)`, `(0,0)`, 5}, {`(-3,4)`, `(1e-300,-1e-300)`, 5}, {`(-10,0)`, `(-5,-12)`, 13}, {`(5.1,34.5)`, `(10,10)`, 24.9851956166046}, {`(1e+300,Infinity)`, `(Infinity,1e+300)`, `Infinity`}},
			},
			{
				Statement: `CREATE TEMP TABLE point_gist_tbl(f1 point);`,
			},
			{
				Statement: `INSERT INTO point_gist_tbl SELECT '(0,0)' FROM generate_series(0,1000);`,
			},
			{
				Statement: `CREATE INDEX point_gist_tbl_index ON point_gist_tbl USING gist (f1);`,
			},
			{
				Statement: `INSERT INTO point_gist_tbl VALUES ('(0.0000009,0.0000009)');`,
			},
			{
				Statement: `SET enable_seqscan TO true;`,
			},
			{
				Statement: `SET enable_indexscan TO false;`,
			},
			{
				Statement: `SET enable_bitmapscan TO false;`,
			},
			{
				Statement: `SELECT COUNT(*) FROM point_gist_tbl WHERE f1 ~= '(0.0000009,0.0000009)'::point;`,
				Results:   []sql.Row{{1002}},
			},
			{
				Statement: `SELECT COUNT(*) FROM point_gist_tbl WHERE f1 <@ '(0.0000009,0.0000009),(0.0000009,0.0000009)'::box;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT COUNT(*) FROM point_gist_tbl WHERE f1 ~= '(0.0000018,0.0000018)'::point;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SET enable_seqscan TO false;`,
			},
			{
				Statement: `SET enable_indexscan TO true;`,
			},
			{
				Statement: `SET enable_bitmapscan TO true;`,
			},
			{
				Statement: `SELECT COUNT(*) FROM point_gist_tbl WHERE f1 ~= '(0.0000009,0.0000009)'::point;`,
				Results:   []sql.Row{{1002}},
			},
			{
				Statement: `SELECT COUNT(*) FROM point_gist_tbl WHERE f1 <@ '(0.0000009,0.0000009),(0.0000009,0.0000009)'::box;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT COUNT(*) FROM point_gist_tbl WHERE f1 ~= '(0.0000018,0.0000018)'::point;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `RESET enable_indexscan;`,
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
		},
	})
}
