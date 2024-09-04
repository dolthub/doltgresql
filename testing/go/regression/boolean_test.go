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

func TestBoolean(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_boolean)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_boolean,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT 1 AS one;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT true AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT false AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT bool 't' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool '   f           ' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT bool 'true' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `SELECT bool 'test' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "test"`,
			},
			{
				Statement: `SELECT bool 'false' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `SELECT bool 'foo' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "foo"`,
			},
			{
				Statement: `SELECT bool 'y' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool 'yes' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `SELECT bool 'yeah' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "yeah"`,
			},
			{
				Statement: `SELECT bool 'n' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT bool 'no' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `SELECT bool 'nay' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "nay"`,
			},
			{
				Statement: `SELECT bool 'on' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool 'off' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT bool 'of' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `SELECT bool 'o' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "o"`,
			},
			{
				Statement:   `SELECT bool 'on_' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "on_"`,
			},
			{
				Statement:   `SELECT bool 'off_' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "off_"`,
			},
			{
				Statement: `SELECT bool '1' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `SELECT bool '11' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "11"`,
			},
			{
				Statement: `SELECT bool '0' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `SELECT bool '000' AS error;`,
				ErrorString: `invalid input syntax for type boolean: "000"`,
			},
			{
				Statement:   `SELECT bool '' AS error;`,
				ErrorString: `invalid input syntax for type boolean: ""`,
			},
			{
				Statement: `SELECT bool 't' or bool 'f' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool 't' and bool 'f' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT not bool 'f' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool 't' = bool 'f' AS false;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT bool 't' <> bool 'f' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool 't' > bool 'f' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool 't' >= bool 'f' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool 'f' < bool 't' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT bool 'f' <= bool 't' AS true;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT 'TrUe'::text::boolean AS true, 'fAlse'::text::boolean AS false;`,
				Results:   []sql.Row{{true, false}},
			},
			{
				Statement: `SELECT '    true   '::text::boolean AS true,
       '     FALSE'::text::boolean AS false;`,
				Results: []sql.Row{{true, false}},
			},
			{
				Statement: `SELECT true::boolean::text AS true, false::boolean::text AS false;`,
				Results:   []sql.Row{{`true`, `false`}},
			},
			{
				Statement:   `SELECT '  tru e '::text::boolean AS invalid;    -- error`,
				ErrorString: `invalid input syntax for type boolean: "  tru e "`,
			},
			{
				Statement:   `SELECT ''::text::boolean AS invalid;            -- error`,
				ErrorString: `invalid input syntax for type boolean: ""`,
			},
			{
				Statement: `CREATE TABLE BOOLTBL1 (f1 bool);`,
			},
			{
				Statement: `INSERT INTO BOOLTBL1 (f1) VALUES (bool 't');`,
			},
			{
				Statement: `INSERT INTO BOOLTBL1 (f1) VALUES (bool 'True');`,
			},
			{
				Statement: `INSERT INTO BOOLTBL1 (f1) VALUES (bool 'true');`,
			},
			{
				Statement: `SELECT BOOLTBL1.* FROM BOOLTBL1;`,
				Results:   []sql.Row{{true}, {true}, {true}},
			},
			{
				Statement: `SELECT BOOLTBL1.*
   FROM BOOLTBL1
   WHERE f1 = bool 'true';`,
				Results: []sql.Row{{true}, {true}, {true}},
			},
			{
				Statement: `SELECT BOOLTBL1.*
   FROM BOOLTBL1
   WHERE f1 <> bool 'false';`,
				Results: []sql.Row{{true}, {true}, {true}},
			},
			{
				Statement: `SELECT BOOLTBL1.*
   FROM BOOLTBL1
   WHERE booleq(bool 'false', f1);`,
				Results: []sql.Row{},
			},
			{
				Statement: `INSERT INTO BOOLTBL1 (f1) VALUES (bool 'f');`,
			},
			{
				Statement: `SELECT BOOLTBL1.*
   FROM BOOLTBL1
   WHERE f1 = bool 'false';`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `CREATE TABLE BOOLTBL2 (f1 bool);`,
			},
			{
				Statement: `INSERT INTO BOOLTBL2 (f1) VALUES (bool 'f');`,
			},
			{
				Statement: `INSERT INTO BOOLTBL2 (f1) VALUES (bool 'false');`,
			},
			{
				Statement: `INSERT INTO BOOLTBL2 (f1) VALUES (bool 'False');`,
			},
			{
				Statement: `INSERT INTO BOOLTBL2 (f1) VALUES (bool 'FALSE');`,
			},
			{
				Statement: `INSERT INTO BOOLTBL2 (f1)
   VALUES (bool 'XXX');`,
				ErrorString: `invalid input syntax for type boolean: "XXX"`,
			},
			{
				Statement: `SELECT BOOLTBL2.* FROM BOOLTBL2;`,
				Results:   []sql.Row{{false}, {false}, {false}, {false}},
			},
			{
				Statement: `SELECT BOOLTBL1.*, BOOLTBL2.*
   FROM BOOLTBL1, BOOLTBL2
   WHERE BOOLTBL2.f1 <> BOOLTBL1.f1;`,
				Results: []sql.Row{{true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}},
			},
			{
				Statement: `SELECT BOOLTBL1.*, BOOLTBL2.*
   FROM BOOLTBL1, BOOLTBL2
   WHERE boolne(BOOLTBL2.f1,BOOLTBL1.f1);`,
				Results: []sql.Row{{true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}},
			},
			{
				Statement: `SELECT BOOLTBL1.*, BOOLTBL2.*
   FROM BOOLTBL1, BOOLTBL2
   WHERE BOOLTBL2.f1 = BOOLTBL1.f1 and BOOLTBL1.f1 = bool 'false';`,
				Results: []sql.Row{{false, false}, {false, false}, {false, false}, {false, false}},
			},
			{
				Statement: `SELECT BOOLTBL1.*, BOOLTBL2.*
   FROM BOOLTBL1, BOOLTBL2
   WHERE BOOLTBL2.f1 = BOOLTBL1.f1 or BOOLTBL1.f1 = bool 'true'
   ORDER BY BOOLTBL1.f1, BOOLTBL2.f1;`,
				Results: []sql.Row{{false, false}, {false, false}, {false, false}, {false, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}, {true, false}},
			},
			{
				Statement: `SELECT f1
   FROM BOOLTBL1
   WHERE f1 IS TRUE;`,
				Results: []sql.Row{{true}, {true}, {true}},
			},
			{
				Statement: `SELECT f1
   FROM BOOLTBL1
   WHERE f1 IS NOT FALSE;`,
				Results: []sql.Row{{true}, {true}, {true}},
			},
			{
				Statement: `SELECT f1
   FROM BOOLTBL1
   WHERE f1 IS FALSE;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT f1
   FROM BOOLTBL1
   WHERE f1 IS NOT TRUE;`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT f1
   FROM BOOLTBL2
   WHERE f1 IS TRUE;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT f1
   FROM BOOLTBL2
   WHERE f1 IS NOT FALSE;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT f1
   FROM BOOLTBL2
   WHERE f1 IS FALSE;`,
				Results: []sql.Row{{false}, {false}, {false}, {false}},
			},
			{
				Statement: `SELECT f1
   FROM BOOLTBL2
   WHERE f1 IS NOT TRUE;`,
				Results: []sql.Row{{false}, {false}, {false}, {false}},
			},
			{
				Statement: `CREATE TABLE BOOLTBL3 (d text, b bool, o int);`,
			},
			{
				Statement: `INSERT INTO BOOLTBL3 (d, b, o) VALUES ('true', true, 1);`,
			},
			{
				Statement: `INSERT INTO BOOLTBL3 (d, b, o) VALUES ('false', false, 2);`,
			},
			{
				Statement: `INSERT INTO BOOLTBL3 (d, b, o) VALUES ('null', null, 3);`,
			},
			{
				Statement: `SELECT
    d,
    b IS TRUE AS istrue,
    b IS NOT TRUE AS isnottrue,
    b IS FALSE AS isfalse,
    b IS NOT FALSE AS isnotfalse,
    b IS UNKNOWN AS isunknown,
    b IS NOT UNKNOWN AS isnotunknown
FROM booltbl3 ORDER BY o;`,
				Results: []sql.Row{{`true`, true, false, false, true, false, true}, {`false`, false, true, true, false, false, true}, {`null`, false, true, false, true, true, false}},
			},
			{
				Statement: `CREATE TABLE booltbl4(isfalse bool, istrue bool, isnul bool);`,
			},
			{
				Statement: `INSERT INTO booltbl4 VALUES (false, true, null);`,
			},
			{
				Statement: `\pset null '(null)'
SELECT istrue AND isnul AND istrue FROM booltbl4;`,
				Results: []sql.Row{{`(null)`}},
			},
			{
				Statement: `SELECT istrue AND istrue AND isnul FROM booltbl4;`,
				Results:   []sql.Row{{`(null)`}},
			},
			{
				Statement: `SELECT isnul AND istrue AND istrue FROM booltbl4;`,
				Results:   []sql.Row{{`(null)`}},
			},
			{
				Statement: `SELECT isfalse AND isnul AND istrue FROM booltbl4;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT istrue AND isfalse AND isnul FROM booltbl4;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT isnul AND istrue AND isfalse FROM booltbl4;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT isfalse OR isnul OR isfalse FROM booltbl4;`,
				Results:   []sql.Row{{`(null)`}},
			},
			{
				Statement: `SELECT isfalse OR isfalse OR isnul FROM booltbl4;`,
				Results:   []sql.Row{{`(null)`}},
			},
			{
				Statement: `SELECT isnul OR isfalse OR isfalse FROM booltbl4;`,
				Results:   []sql.Row{{`(null)`}},
			},
			{
				Statement: `SELECT isfalse OR isnul OR istrue FROM booltbl4;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT istrue OR isfalse OR isnul FROM booltbl4;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT isnul OR istrue OR isfalse FROM booltbl4;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP TABLE  BOOLTBL1;`,
			},
			{
				Statement: `DROP TABLE  BOOLTBL2;`,
			},
			{
				Statement: `DROP TABLE  BOOLTBL3;`,
			},
			{
				Statement: `DROP TABLE  BOOLTBL4;`,
			},
		},
	})
}
