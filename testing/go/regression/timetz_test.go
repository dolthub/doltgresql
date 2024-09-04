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

func TestTimetz(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_timetz)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_timetz,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE TIMETZ_TBL (f1 time(2) with time zone);`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('00:01 PDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('01:00 PDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('02:03 PDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('07:07 PST');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('08:08 EDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('11:59 PDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('12:00 PDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('12:01 PDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('23:59 PDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('11:59:59.99 PM PDT');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('2003-03-07 15:36:39 America/New_York');`,
			},
			{
				Statement: `INSERT INTO TIMETZ_TBL VALUES ('2003-07-07 15:36:39 America/New_York');`,
			},
			{
				Statement:   `INSERT INTO TIMETZ_TBL VALUES ('15:36:39 America/New_York');`,
				ErrorString: `invalid input syntax for type time with time zone: "15:36:39 America/New_York"`,
			},
			{
				Statement:   `INSERT INTO TIMETZ_TBL VALUES ('15:36:39 m2');`,
				ErrorString: `invalid input syntax for type time with time zone: "15:36:39 m2"`,
			},
			{
				Statement:   `INSERT INTO TIMETZ_TBL VALUES ('15:36:39 MSK m2');`,
				ErrorString: `invalid input syntax for type time with time zone: "15:36:39 MSK m2"`,
			},
			{
				Statement: `SELECT f1 AS "Time TZ" FROM TIMETZ_TBL;`,
				Results:   []sql.Row{{`00:01:00-07`}, {`01:00:00-07`}, {`02:03:00-07`}, {`07:07:00-08`}, {`08:08:00-04`}, {`11:59:00-07`}, {`12:00:00-07`}, {`12:01:00-07`}, {`23:59:00-07`}, {`23:59:59.99-07`}, {`15:36:39-05`}, {`15:36:39-04`}},
			},
			{
				Statement: `SELECT f1 AS "Three" FROM TIMETZ_TBL WHERE f1 < '05:06:07-07';`,
				Results:   []sql.Row{{`00:01:00-07`}, {`01:00:00-07`}, {`02:03:00-07`}},
			},
			{
				Statement: `SELECT f1 AS "Seven" FROM TIMETZ_TBL WHERE f1 > '05:06:07-07';`,
				Results:   []sql.Row{{`07:07:00-08`}, {`08:08:00-04`}, {`11:59:00-07`}, {`12:00:00-07`}, {`12:01:00-07`}, {`23:59:00-07`}, {`23:59:59.99-07`}, {`15:36:39-05`}, {`15:36:39-04`}},
			},
			{
				Statement: `SELECT f1 AS "None" FROM TIMETZ_TBL WHERE f1 < '00:00-07';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT f1 AS "Ten" FROM TIMETZ_TBL WHERE f1 >= '00:00-07';`,
				Results:   []sql.Row{{`00:01:00-07`}, {`01:00:00-07`}, {`02:03:00-07`}, {`07:07:00-08`}, {`08:08:00-04`}, {`11:59:00-07`}, {`12:00:00-07`}, {`12:01:00-07`}, {`23:59:00-07`}, {`23:59:59.99-07`}, {`15:36:39-05`}, {`15:36:39-04`}},
			},
			{
				Statement: `SELECT '23:59:59.999999 PDT'::timetz;`,
				Results:   []sql.Row{{`23:59:59.999999-07`}},
			},
			{
				Statement: `SELECT '23:59:59.9999999 PDT'::timetz;  -- rounds up`,
				Results:   []sql.Row{{`24:00:00-07`}},
			},
			{
				Statement: `SELECT '23:59:60 PDT'::timetz;  -- rounds up`,
				Results:   []sql.Row{{`24:00:00-07`}},
			},
			{
				Statement: `SELECT '24:00:00 PDT'::timetz;  -- allowed`,
				Results:   []sql.Row{{`24:00:00-07`}},
			},
			{
				Statement:   `SELECT '24:00:00.01 PDT'::timetz;  -- not allowed`,
				ErrorString: `date/time field value out of range: "24:00:00.01 PDT"`,
			},
			{
				Statement:   `SELECT '23:59:60.01 PDT'::timetz;  -- not allowed`,
				ErrorString: `date/time field value out of range: "23:59:60.01 PDT"`,
			},
			{
				Statement:   `SELECT '24:01:00 PDT'::timetz;  -- not allowed`,
				ErrorString: `date/time field value out of range: "24:01:00 PDT"`,
			},
			{
				Statement:   `SELECT '25:00:00 PDT'::timetz;  -- not allowed`,
				ErrorString: `date/time field value out of range: "25:00:00 PDT"`,
			},
			{
				Statement:   `SELECT f1 + time with time zone '00:01' AS "Illegal" FROM TIMETZ_TBL;`,
				ErrorString: `operator does not exist: time with time zone + time with time zone`,
			},
			{
				Statement: `SELECT EXTRACT(MICROSECOND FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{25575401}},
			},
			{
				Statement: `SELECT EXTRACT(MILLISECOND FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{25575.401}},
			},
			{
				Statement: `SELECT EXTRACT(SECOND      FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{25.575401}},
			},
			{
				Statement: `SELECT EXTRACT(MINUTE      FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{30}},
			},
			{
				Statement: `SELECT EXTRACT(HOUR        FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{13}},
			},
			{
				Statement:   `SELECT EXTRACT(DAY         FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');  -- error`,
				ErrorString: `unit "day" not supported for type time with time zone`,
			},
			{
				Statement:   `SELECT EXTRACT(FORTNIGHT   FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');  -- error`,
				ErrorString: `unit "fortnight" not recognized for type time with time zone`,
			},
			{
				Statement: `SELECT EXTRACT(TIMEZONE    FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04:30');`,
				Results:   []sql.Row{{-16200}},
			},
			{
				Statement: `SELECT EXTRACT(TIMEZONE_HOUR   FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04:30');`,
				Results:   []sql.Row{{-4}},
			},
			{
				Statement: `SELECT EXTRACT(TIMEZONE_MINUTE FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04:30');`,
				Results:   []sql.Row{{-30}},
			},
			{
				Statement: `SELECT EXTRACT(EPOCH       FROM TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{63025.575401}},
			},
			{
				Statement: `SELECT date_part('microsecond', TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{25575401}},
			},
			{
				Statement: `SELECT date_part('millisecond', TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{25575.401}},
			},
			{
				Statement: `SELECT date_part('second',      TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{25.575401}},
			},
			{
				Statement: `SELECT date_part('epoch',       TIME WITH TIME ZONE '2020-05-26 13:30:25.575401-04');`,
				Results:   []sql.Row{{63025.575401}},
			},
		},
	})
}
