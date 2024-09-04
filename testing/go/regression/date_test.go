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

func TestDate(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_date)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_date,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE DATE_TBL (f1 date);`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1957-04-09');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1957-06-13');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1996-02-28');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1996-02-29');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1996-03-01');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1996-03-02');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1997-02-28');`,
			},
			{
				Statement:   `INSERT INTO DATE_TBL VALUES ('1997-02-29');`,
				ErrorString: `date/time field value out of range: "1997-02-29"`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1997-03-01');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('1997-03-02');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('2000-04-01');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('2000-04-02');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('2000-04-03');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('2038-04-08');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('2039-04-09');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('2040-04-10');`,
			},
			{
				Statement: `INSERT INTO DATE_TBL VALUES ('2040-04-10 BC');`,
			},
			{
				Statement: `SELECT f1 FROM DATE_TBL;`,
				Results:   []sql.Row{{`04-09-1957`}, {`06-13-1957`}, {`02-28-1996`}, {`02-29-1996`}, {`03-01-1996`}, {`03-02-1996`}, {`02-28-1997`}, {`03-01-1997`}, {`03-02-1997`}, {`04-01-2000`}, {`04-02-2000`}, {`04-03-2000`}, {`04-08-2038`}, {`04-09-2039`}, {`04-10-2040`}, {`04-10-2040 BC`}},
			},
			{
				Statement: `SELECT f1 FROM DATE_TBL WHERE f1 < '2000-01-01';`,
				Results:   []sql.Row{{`04-09-1957`}, {`06-13-1957`}, {`02-28-1996`}, {`02-29-1996`}, {`03-01-1996`}, {`03-02-1996`}, {`02-28-1997`}, {`03-01-1997`}, {`03-02-1997`}, {`04-10-2040 BC`}},
			},
			{
				Statement: `SELECT f1 FROM DATE_TBL
  WHERE f1 BETWEEN '2000-01-01' AND '2001-01-01';`,
				Results: []sql.Row{{`04-01-2000`}, {`04-02-2000`}, {`04-03-2000`}},
			},
			{
				Statement: `SET datestyle TO iso;  -- display results in ISO`,
			},
			{
				Statement: `SET datestyle TO ymd;`,
			},
			{
				Statement: `SELECT date 'January 8, 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999-01-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999-01-18';`,
				Results:   []sql.Row{{`1999-01-18`}},
			},
			{
				Statement:   `SELECT date '1/8/1999';`,
				ErrorString: `date/time field value out of range: "1/8/1999"`,
			},
			{
				Statement:   `SELECT date '1/18/1999';`,
				ErrorString: `date/time field value out of range: "1/18/1999"`,
			},
			{
				Statement:   `SELECT date '18/1/1999';`,
				ErrorString: `date/time field value out of range: "18/1/1999"`,
			},
			{
				Statement: `SELECT date '01/02/03';`,
				Results:   []sql.Row{{`2001-02-03`}},
			},
			{
				Statement: `SELECT date '19990108';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '990108';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999.008';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'J2451187';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date 'January 8, 99 BC';`,
				ErrorString: `date/time field value out of range: "January 8, 99 BC"`,
			},
			{
				Statement: `SELECT date '99-Jan-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999-Jan-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '08-Jan-99';`,
				ErrorString: `date/time field value out of range: "08-Jan-99"`,
			},
			{
				Statement: `SELECT date '08-Jan-1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date 'Jan-08-99';`,
				ErrorString: `date/time field value out of range: "Jan-08-99"`,
			},
			{
				Statement: `SELECT date 'Jan-08-1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99-08-Jan';`,
				ErrorString: `invalid input syntax for type date: "99-08-Jan"`,
			},
			{
				Statement:   `SELECT date '1999-08-Jan';`,
				ErrorString: `invalid input syntax for type date: "1999-08-Jan"`,
			},
			{
				Statement: `SELECT date '99 Jan 08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999 Jan 08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '08 Jan 99';`,
				ErrorString: `date/time field value out of range: "08 Jan 99"`,
			},
			{
				Statement: `SELECT date '08 Jan 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date 'Jan 08 99';`,
				ErrorString: `date/time field value out of range: "Jan 08 99"`,
			},
			{
				Statement: `SELECT date 'Jan 08 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '99 08 Jan';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999 08 Jan';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '99-01-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999-01-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '08-01-99';`,
				ErrorString: `date/time field value out of range: "08-01-99"`,
			},
			{
				Statement:   `SELECT date '08-01-1999';`,
				ErrorString: `date/time field value out of range: "08-01-1999"`,
			},
			{
				Statement:   `SELECT date '01-08-99';`,
				ErrorString: `date/time field value out of range: "01-08-99"`,
			},
			{
				Statement:   `SELECT date '01-08-1999';`,
				ErrorString: `date/time field value out of range: "01-08-1999"`,
			},
			{
				Statement: `SELECT date '99-08-01';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '1999-08-01';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '99 01 08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999 01 08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '08 01 99';`,
				ErrorString: `date/time field value out of range: "08 01 99"`,
			},
			{
				Statement:   `SELECT date '08 01 1999';`,
				ErrorString: `date/time field value out of range: "08 01 1999"`,
			},
			{
				Statement:   `SELECT date '01 08 99';`,
				ErrorString: `date/time field value out of range: "01 08 99"`,
			},
			{
				Statement:   `SELECT date '01 08 1999';`,
				ErrorString: `date/time field value out of range: "01 08 1999"`,
			},
			{
				Statement: `SELECT date '99 08 01';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '1999 08 01';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SET datestyle TO dmy;`,
			},
			{
				Statement: `SELECT date 'January 8, 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999-01-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999-01-18';`,
				Results:   []sql.Row{{`1999-01-18`}},
			},
			{
				Statement: `SELECT date '1/8/1999';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement:   `SELECT date '1/18/1999';`,
				ErrorString: `date/time field value out of range: "1/18/1999"`,
			},
			{
				Statement: `SELECT date '18/1/1999';`,
				Results:   []sql.Row{{`1999-01-18`}},
			},
			{
				Statement: `SELECT date '01/02/03';`,
				Results:   []sql.Row{{`2003-02-01`}},
			},
			{
				Statement: `SELECT date '19990108';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '990108';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999.008';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'J2451187';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'January 8, 99 BC';`,
				Results:   []sql.Row{{`0099-01-08 BC`}},
			},
			{
				Statement:   `SELECT date '99-Jan-08';`,
				ErrorString: `date/time field value out of range: "99-Jan-08"`,
			},
			{
				Statement: `SELECT date '1999-Jan-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08-Jan-99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08-Jan-1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'Jan-08-99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'Jan-08-1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99-08-Jan';`,
				ErrorString: `invalid input syntax for type date: "99-08-Jan"`,
			},
			{
				Statement:   `SELECT date '1999-08-Jan';`,
				ErrorString: `invalid input syntax for type date: "1999-08-Jan"`,
			},
			{
				Statement:   `SELECT date '99 Jan 08';`,
				ErrorString: `date/time field value out of range: "99 Jan 08"`,
			},
			{
				Statement: `SELECT date '1999 Jan 08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08 Jan 99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08 Jan 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'Jan 08 99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'Jan 08 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99 08 Jan';`,
				ErrorString: `invalid input syntax for type date: "99 08 Jan"`,
			},
			{
				Statement: `SELECT date '1999 08 Jan';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99-01-08';`,
				ErrorString: `date/time field value out of range: "99-01-08"`,
			},
			{
				Statement: `SELECT date '1999-01-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08-01-99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08-01-1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '01-08-99';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '01-08-1999';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement:   `SELECT date '99-08-01';`,
				ErrorString: `date/time field value out of range: "99-08-01"`,
			},
			{
				Statement: `SELECT date '1999-08-01';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement:   `SELECT date '99 01 08';`,
				ErrorString: `date/time field value out of range: "99 01 08"`,
			},
			{
				Statement: `SELECT date '1999 01 08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08 01 99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08 01 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '01 08 99';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '01 08 1999';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement:   `SELECT date '99 08 01';`,
				ErrorString: `date/time field value out of range: "99 08 01"`,
			},
			{
				Statement: `SELECT date '1999 08 01';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SET datestyle TO mdy;`,
			},
			{
				Statement: `SELECT date 'January 8, 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999-01-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999-01-18';`,
				Results:   []sql.Row{{`1999-01-18`}},
			},
			{
				Statement: `SELECT date '1/8/1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1/18/1999';`,
				Results:   []sql.Row{{`1999-01-18`}},
			},
			{
				Statement:   `SELECT date '18/1/1999';`,
				ErrorString: `date/time field value out of range: "18/1/1999"`,
			},
			{
				Statement: `SELECT date '01/02/03';`,
				Results:   []sql.Row{{`2003-01-02`}},
			},
			{
				Statement: `SELECT date '19990108';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '990108';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '1999.008';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'J2451187';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'January 8, 99 BC';`,
				Results:   []sql.Row{{`0099-01-08 BC`}},
			},
			{
				Statement:   `SELECT date '99-Jan-08';`,
				ErrorString: `date/time field value out of range: "99-Jan-08"`,
			},
			{
				Statement: `SELECT date '1999-Jan-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08-Jan-99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08-Jan-1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'Jan-08-99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'Jan-08-1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99-08-Jan';`,
				ErrorString: `invalid input syntax for type date: "99-08-Jan"`,
			},
			{
				Statement:   `SELECT date '1999-08-Jan';`,
				ErrorString: `invalid input syntax for type date: "1999-08-Jan"`,
			},
			{
				Statement:   `SELECT date '99 Jan 08';`,
				ErrorString: `invalid input syntax for type date: "99 Jan 08"`,
			},
			{
				Statement: `SELECT date '1999 Jan 08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08 Jan 99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08 Jan 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'Jan 08 99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date 'Jan 08 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99 08 Jan';`,
				ErrorString: `invalid input syntax for type date: "99 08 Jan"`,
			},
			{
				Statement: `SELECT date '1999 08 Jan';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99-01-08';`,
				ErrorString: `date/time field value out of range: "99-01-08"`,
			},
			{
				Statement: `SELECT date '1999-01-08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08-01-99';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '08-01-1999';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '01-08-99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '01-08-1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99-08-01';`,
				ErrorString: `date/time field value out of range: "99-08-01"`,
			},
			{
				Statement: `SELECT date '1999-08-01';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement:   `SELECT date '99 01 08';`,
				ErrorString: `date/time field value out of range: "99 01 08"`,
			},
			{
				Statement: `SELECT date '1999 01 08';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '08 01 99';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '08 01 1999';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '01 08 99';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement: `SELECT date '01 08 1999';`,
				Results:   []sql.Row{{`1999-01-08`}},
			},
			{
				Statement:   `SELECT date '99 08 01';`,
				ErrorString: `date/time field value out of range: "99 08 01"`,
			},
			{
				Statement: `SELECT date '1999 08 01';`,
				Results:   []sql.Row{{`1999-08-01`}},
			},
			{
				Statement: `SELECT date '4714-11-24 BC';`,
				Results:   []sql.Row{{`4714-11-24 BC`}},
			},
			{
				Statement:   `SELECT date '4714-11-23 BC';  -- out of range`,
				ErrorString: `date out of range: "4714-11-23 BC"`,
			},
			{
				Statement: `SELECT date '5874897-12-31';`,
				Results:   []sql.Row{{`5874897-12-31`}},
			},
			{
				Statement:   `SELECT date '5874898-01-01';  -- out of range`,
				ErrorString: `date out of range: "5874898-01-01"`,
			},
			{
				Statement: `RESET datestyle;`,
			},
			{
				Statement: `SELECT f1 - date '2000-01-01' AS "Days From 2K" FROM DATE_TBL;`,
				Results:   []sql.Row{{-15607}, {-15542}, {-1403}, {-1402}, {-1401}, {-1400}, {-1037}, {-1036}, {-1035}, {91}, {92}, {93}, {13977}, {14343}, {14710}, {-1475115}},
			},
			{
				Statement: `SELECT f1 - date 'epoch' AS "Days From Epoch" FROM DATE_TBL;`,
				Results:   []sql.Row{{-4650}, {-4585}, {9554}, {9555}, {9556}, {9557}, {9920}, {9921}, {9922}, {11048}, {11049}, {11050}, {24934}, {25300}, {25667}, {-1464158}},
			},
			{
				Statement: `SELECT date 'yesterday' - date 'today' AS "One day";`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT date 'today' - date 'tomorrow' AS "One day";`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT date 'yesterday' - date 'tomorrow' AS "Two days";`,
				Results:   []sql.Row{{-2}},
			},
			{
				Statement: `SELECT date 'tomorrow' - date 'today' AS "One day";`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT date 'today' - date 'yesterday' AS "One day";`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT date 'tomorrow' - date 'yesterday' AS "Two days";`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT f1 as "date",
    date_part('year', f1) AS year,
    date_part('month', f1) AS month,
    date_part('day', f1) AS day,
    date_part('quarter', f1) AS quarter,
    date_part('decade', f1) AS decade,
    date_part('century', f1) AS century,
    date_part('millennium', f1) AS millennium,
    date_part('isoyear', f1) AS isoyear,
    date_part('week', f1) AS week,
    date_part('dow', f1) AS dow,
    date_part('isodow', f1) AS isodow,
    date_part('doy', f1) AS doy,
    date_part('julian', f1) AS julian,
    date_part('epoch', f1) AS epoch
    FROM date_tbl;`,
				Results: []sql.Row{{`04-09-1957`, 1957, 4, 9, 2, 195, 20, 2, 1957, 15, 2, 2, 99, 2435938, -401760000}, {`06-13-1957`, 1957, 6, 13, 2, 195, 20, 2, 1957, 24, 4, 4, 164, 2436003, -396144000}, {`02-28-1996`, 1996, 2, 28, 1, 199, 20, 2, 1996, 9, 3, 3, 59, 2450142, 825465600}, {`02-29-1996`, 1996, 2, 29, 1, 199, 20, 2, 1996, 9, 4, 4, 60, 2450143, 825552000}, {`03-01-1996`, 1996, 3, 1, 1, 199, 20, 2, 1996, 9, 5, 5, 61, 2450144, 825638400}, {`03-02-1996`, 1996, 3, 2, 1, 199, 20, 2, 1996, 9, 6, 6, 62, 2450145, 825724800}, {`02-28-1997`, 1997, 2, 28, 1, 199, 20, 2, 1997, 9, 5, 5, 59, 2450508, 857088000}, {`03-01-1997`, 1997, 3, 1, 1, 199, 20, 2, 1997, 9, 6, 6, 60, 2450509, 857174400}, {`03-02-1997`, 1997, 3, 2, 1, 199, 20, 2, 1997, 9, 0, 7, 61, 2450510, 857260800}, {`04-01-2000`, 2000, 4, 1, 2, 200, 20, 2, 2000, 13, 6, 6, 92, 2451636, 954547200}, {`04-02-2000`, 2000, 4, 2, 2, 200, 20, 2, 2000, 13, 0, 7, 93, 2451637, 954633600}, {`04-03-2000`, 2000, 4, 3, 2, 200, 20, 2, 2000, 14, 1, 1, 94, 2451638, 954720000}, {`04-08-2038`, 2038, 4, 8, 2, 203, 21, 3, 2038, 14, 4, 4, 98, 2465522, 2154297600}, {`04-09-2039`, 2039, 4, 9, 2, 203, 21, 3, 2039, 14, 6, 6, 99, 2465888, 2185920000}, {`04-10-2040`, 2040, 4, 10, 2, 204, 21, 3, 2040, 15, 2, 2, 101, 2466255, 2217628800}, {`04-10-2040 BC`, -2040, 4, 10, 2, -204, -21, -3, -2040, 15, 1, 1, 100, 976430, -126503251200}},
			},
			{
				Statement: `SELECT EXTRACT(EPOCH FROM DATE        '1970-01-01');     --  0`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '0101-12-31 BC'); -- -2`,
				Results:   []sql.Row{{-2}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '0100-12-31 BC'); -- -1`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '0001-12-31 BC'); -- -1`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '0001-01-01');    --  1`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '0001-01-01 AD'); --  1`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '1900-12-31');    -- 19`,
				Results:   []sql.Row{{19}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '1901-01-01');    -- 20`,
				Results:   []sql.Row{{20}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '2000-12-31');    -- 20`,
				Results:   []sql.Row{{20}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM DATE '2001-01-01');    -- 21`,
				Results:   []sql.Row{{21}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM CURRENT_DATE)>=21 AS True;     -- true`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM FROM DATE '0001-12-31 BC'); -- -1`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM FROM DATE '0001-01-01 AD'); --  1`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM FROM DATE '1000-12-31');    --  1`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM FROM DATE '1001-01-01');    --  2`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM FROM DATE '2000-12-31');    --  2`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM FROM DATE '2001-01-01');    --  3`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM FROM CURRENT_DATE);         --  3`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM DATE '1994-12-25');    -- 199`,
				Results:   []sql.Row{{199}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM DATE '0010-01-01');    --   1`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM DATE '0009-12-31');    --   0`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM DATE '0001-01-01 BC'); --   0`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM DATE '0002-12-31 BC'); --  -1`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM DATE '0011-01-01 BC'); --  -1`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM DATE '0012-12-31 BC'); --  -2`,
				Results:   []sql.Row{{-2}},
			},
			{
				Statement:   `SELECT EXTRACT(MICROSECONDS  FROM DATE '2020-08-11');`,
				ErrorString: `unit "microseconds" not supported for type date`,
			},
			{
				Statement:   `SELECT EXTRACT(MILLISECONDS  FROM DATE '2020-08-11');`,
				ErrorString: `unit "milliseconds" not supported for type date`,
			},
			{
				Statement:   `SELECT EXTRACT(SECOND        FROM DATE '2020-08-11');`,
				ErrorString: `unit "second" not supported for type date`,
			},
			{
				Statement:   `SELECT EXTRACT(MINUTE        FROM DATE '2020-08-11');`,
				ErrorString: `unit "minute" not supported for type date`,
			},
			{
				Statement:   `SELECT EXTRACT(HOUR          FROM DATE '2020-08-11');`,
				ErrorString: `unit "hour" not supported for type date`,
			},
			{
				Statement: `SELECT EXTRACT(DAY           FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{11}},
			},
			{
				Statement: `SELECT EXTRACT(MONTH         FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{8}},
			},
			{
				Statement: `SELECT EXTRACT(YEAR          FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{2020}},
			},
			{
				Statement: `SELECT EXTRACT(YEAR          FROM DATE '2020-08-11 BC');`,
				Results:   []sql.Row{{-2020}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE        FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{202}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY       FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{21}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM    FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT EXTRACT(ISOYEAR       FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{2020}},
			},
			{
				Statement: `SELECT EXTRACT(ISOYEAR       FROM DATE '2020-08-11 BC');`,
				Results:   []sql.Row{{-2020}},
			},
			{
				Statement: `SELECT EXTRACT(QUARTER       FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT EXTRACT(WEEK          FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{33}},
			},
			{
				Statement: `SELECT EXTRACT(DOW           FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT EXTRACT(DOW           FROM DATE '2020-08-16');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT EXTRACT(ISODOW        FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT EXTRACT(ISODOW        FROM DATE '2020-08-16');`,
				Results:   []sql.Row{{7}},
			},
			{
				Statement: `SELECT EXTRACT(DOY           FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{224}},
			},
			{
				Statement:   `SELECT EXTRACT(TIMEZONE      FROM DATE '2020-08-11');`,
				ErrorString: `unit "timezone" not supported for type date`,
			},
			{
				Statement:   `SELECT EXTRACT(TIMEZONE_M    FROM DATE '2020-08-11');`,
				ErrorString: `unit "timezone_m" not supported for type date`,
			},
			{
				Statement:   `SELECT EXTRACT(TIMEZONE_H    FROM DATE '2020-08-11');`,
				ErrorString: `unit "timezone_h" not supported for type date`,
			},
			{
				Statement: `SELECT EXTRACT(EPOCH         FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{1597104000}},
			},
			{
				Statement: `SELECT EXTRACT(JULIAN        FROM DATE '2020-08-11');`,
				Results:   []sql.Row{{2459073}},
			},
			{
				Statement: `SELECT DATE_TRUNC('MILLENNIUM', TIMESTAMP '1970-03-20 04:30:00.00000'); -- 1001`,
				Results:   []sql.Row{{`Thu Jan 01 00:00:00 1001`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('MILLENNIUM', DATE '1970-03-20'); -- 1001-01-01`,
				Results:   []sql.Row{{`Thu Jan 01 00:00:00 1001 PST`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('CENTURY', TIMESTAMP '1970-03-20 04:30:00.00000'); -- 1901`,
				Results:   []sql.Row{{`Tue Jan 01 00:00:00 1901`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('CENTURY', DATE '1970-03-20'); -- 1901`,
				Results:   []sql.Row{{`Tue Jan 01 00:00:00 1901 PST`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('CENTURY', DATE '2004-08-10'); -- 2001-01-01`,
				Results:   []sql.Row{{`Mon Jan 01 00:00:00 2001 PST`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('CENTURY', DATE '0002-02-04'); -- 0001-01-01`,
				Results:   []sql.Row{{`Mon Jan 01 00:00:00 0001 PST`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('CENTURY', DATE '0055-08-10 BC'); -- 0100-01-01 BC`,
				Results:   []sql.Row{{`Tue Jan 01 00:00:00 0100 PST BC`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('DECADE', DATE '1993-12-25'); -- 1990-01-01`,
				Results:   []sql.Row{{`Mon Jan 01 00:00:00 1990 PST`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('DECADE', DATE '0004-12-25'); -- 0001-01-01 BC`,
				Results:   []sql.Row{{`Sat Jan 01 00:00:00 0001 PST BC`}},
			},
			{
				Statement: `SELECT DATE_TRUNC('DECADE', DATE '0002-12-31 BC'); -- 0011-01-01 BC`,
				Results:   []sql.Row{{`Mon Jan 01 00:00:00 0011 PST BC`}},
			},
			{
				Statement: `select 'infinity'::date, '-infinity'::date;`,
				Results:   []sql.Row{{`infinity`, `-infinity`}},
			},
			{
				Statement: `select 'infinity'::date > 'today'::date as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select '-infinity'::date < 'today'::date as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select isfinite('infinity'::date), isfinite('-infinity'::date), isfinite('today'::date);`,
				Results:   []sql.Row{{false, false, true}},
			},
			{
				Statement: `SELECT EXTRACT(DAY FROM DATE 'infinity');      -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(DAY FROM DATE '-infinity');     -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(DAY           FROM DATE 'infinity');    -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(MONTH         FROM DATE 'infinity');    -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(QUARTER       FROM DATE 'infinity');    -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(WEEK          FROM DATE 'infinity');    -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(DOW           FROM DATE 'infinity');    -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(ISODOW        FROM DATE 'infinity');    -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(DOY           FROM DATE 'infinity');    -- NULL`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT EXTRACT(EPOCH FROM DATE 'infinity');         --  Infinity`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT EXTRACT(EPOCH FROM DATE '-infinity');        -- -Infinity`,
				Results:   []sql.Row{{`-Infinity`}},
			},
			{
				Statement: `SELECT EXTRACT(YEAR       FROM DATE 'infinity');    --  Infinity`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE     FROM DATE 'infinity');    --  Infinity`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY    FROM DATE 'infinity');    --  Infinity`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT EXTRACT(MILLENNIUM FROM DATE 'infinity');    --  Infinity`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT EXTRACT(JULIAN     FROM DATE 'infinity');    --  Infinity`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT EXTRACT(ISOYEAR    FROM DATE 'infinity');    --  Infinity`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement: `SELECT EXTRACT(EPOCH      FROM DATE 'infinity');    --  Infinity`,
				Results:   []sql.Row{{`Infinity`}},
			},
			{
				Statement:   `SELECT EXTRACT(MICROSEC  FROM DATE 'infinity');     -- error`,
				ErrorString: `unit "microsec" not recognized for type date`,
			},
			{
				Statement: `select make_date(2013, 7, 15);`,
				Results:   []sql.Row{{`07-15-2013`}},
			},
			{
				Statement: `select make_date(-44, 3, 15);`,
				Results:   []sql.Row{{`03-15-0044 BC`}},
			},
			{
				Statement: `select make_time(8, 20, 0.0);`,
				Results:   []sql.Row{{`08:20:00`}},
			},
			{
				Statement:   `select make_date(0, 7, 15);`,
				ErrorString: `date field value out of range: 0-07-15`,
			},
			{
				Statement:   `select make_date(2013, 2, 30);`,
				ErrorString: `date field value out of range: 2013-02-30`,
			},
			{
				Statement:   `select make_date(2013, 13, 1);`,
				ErrorString: `date field value out of range: 2013-13-01`,
			},
			{
				Statement:   `select make_date(2013, 11, -1);`,
				ErrorString: `date field value out of range: 2013-11--1`,
			},
			{
				Statement:   `select make_time(10, 55, 100.1);`,
				ErrorString: `time field value out of range: 10:55:100.1`,
			},
			{
				Statement:   `select make_time(24, 0, 2.1);`,
				ErrorString: `time field value out of range: 24:00:2.1`,
			},
		},
	})
}
