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

func TestTime(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_time)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_time,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE TIME_TBL (f1 time(2));`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('00:00');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('01:00');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('02:03 PST');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('11:59 EDT');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('12:00');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('12:01');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('23:59');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('11:59:59.99 PM');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('2003-03-07 15:36:39 America/New_York');`,
			},
			{
				Statement: `INSERT INTO TIME_TBL VALUES ('2003-07-07 15:36:39 America/New_York');`,
			},
			{
				Statement:   `INSERT INTO TIME_TBL VALUES ('15:36:39 America/New_York');`,
				ErrorString: `invalid input syntax for type time: "15:36:39 America/New_York"`,
			},
			{
				Statement: `SELECT f1 AS "Time" FROM TIME_TBL;`,
				Results:   []sql.Row{{`00:00:00`}, {`01:00:00`}, {`02:03:00`}, {`11:59:00`}, {`12:00:00`}, {`12:01:00`}, {`23:59:00`}, {`23:59:59.99`}, {`15:36:39`}, {`15:36:39`}},
			},
			{
				Statement: `SELECT f1 AS "Three" FROM TIME_TBL WHERE f1 < '05:06:07';`,
				Results:   []sql.Row{{`00:00:00`}, {`01:00:00`}, {`02:03:00`}},
			},
			{
				Statement: `SELECT f1 AS "Five" FROM TIME_TBL WHERE f1 > '05:06:07';`,
				Results:   []sql.Row{{`11:59:00`}, {`12:00:00`}, {`12:01:00`}, {`23:59:00`}, {`23:59:59.99`}, {`15:36:39`}, {`15:36:39`}},
			},
			{
				Statement: `SELECT f1 AS "None" FROM TIME_TBL WHERE f1 < '00:00';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT f1 AS "Eight" FROM TIME_TBL WHERE f1 >= '00:00';`,
				Results:   []sql.Row{{`00:00:00`}, {`01:00:00`}, {`02:03:00`}, {`11:59:00`}, {`12:00:00`}, {`12:01:00`}, {`23:59:00`}, {`23:59:59.99`}, {`15:36:39`}, {`15:36:39`}},
			},
			{
				Statement: `SELECT '23:59:59.999999'::time;`,
				Results:   []sql.Row{{`23:59:59.999999`}},
			},
			{
				Statement: `SELECT '23:59:59.9999999'::time;  -- rounds up`,
				Results:   []sql.Row{{`24:00:00`}},
			},
			{
				Statement: `SELECT '23:59:60'::time;  -- rounds up`,
				Results:   []sql.Row{{`24:00:00`}},
			},
			{
				Statement: `SELECT '24:00:00'::time;  -- allowed`,
				Results:   []sql.Row{{`24:00:00`}},
			},
			{
				Statement:   `SELECT '24:00:00.01'::time;  -- not allowed`,
				ErrorString: `date/time field value out of range: "24:00:00.01"`,
			},
			{
				Statement:   `SELECT '23:59:60.01'::time;  -- not allowed`,
				ErrorString: `date/time field value out of range: "23:59:60.01"`,
			},
			{
				Statement:   `SELECT '24:01:00'::time;  -- not allowed`,
				ErrorString: `date/time field value out of range: "24:01:00"`,
			},
			{
				Statement:   `SELECT '25:00:00'::time;  -- not allowed`,
				ErrorString: `date/time field value out of range: "25:00:00"`,
			},
			{
				Statement:   `SELECT f1 + time '00:01' AS "Illegal" FROM TIME_TBL;`,
				ErrorString: `operator is not unique: time without time zone + time without time zone`,
			},
			{
				Statement: `SELECT EXTRACT(MICROSECOND FROM TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{25575401}},
			},
			{
				Statement: `SELECT EXTRACT(MILLISECOND FROM TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{25575.401}},
			},
			{
				Statement: `SELECT EXTRACT(SECOND      FROM TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{25.575401}},
			},
			{
				Statement: `SELECT EXTRACT(MINUTE      FROM TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{30}},
			},
			{
				Statement: `SELECT EXTRACT(HOUR        FROM TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{13}},
			},
			{
				Statement:   `SELECT EXTRACT(DAY         FROM TIME '2020-05-26 13:30:25.575401');  -- error`,
				ErrorString: `unit "day" not supported for type time without time zone`,
			},
			{
				Statement:   `SELECT EXTRACT(FORTNIGHT   FROM TIME '2020-05-26 13:30:25.575401');  -- error`,
				ErrorString: `unit "fortnight" not recognized for type time without time zone`,
			},
			{
				Statement:   `SELECT EXTRACT(TIMEZONE    FROM TIME '2020-05-26 13:30:25.575401');  -- error`,
				ErrorString: `unit "timezone" not supported for type time without time zone`,
			},
			{
				Statement: `SELECT EXTRACT(EPOCH       FROM TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{48625.575401}},
			},
			{
				Statement: `SELECT date_part('microsecond', TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{25575401}},
			},
			{
				Statement: `SELECT date_part('millisecond', TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{25575.401}},
			},
			{
				Statement: `SELECT date_part('second',      TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{25.575401}},
			},
			{
				Statement: `SELECT date_part('epoch',       TIME '2020-05-26 13:30:25.575401');`,
				Results:   []sql.Row{{48625.575401}},
			},
		},
	})
}
