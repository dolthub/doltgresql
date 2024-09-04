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

func TestTimestamptz(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_timestamptz)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_timestamptz,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE TIMESTAMPTZ_TBL (d1 timestamp(2) with time zone);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('today');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('yesterday');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('tomorrow');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('tomorrow EST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('tomorrow zulu');`,
			},
			{
				Statement: `SELECT count(*) AS One FROM TIMESTAMPTZ_TBL WHERE d1 = timestamp with time zone 'today';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) AS One FROM TIMESTAMPTZ_TBL WHERE d1 = timestamp with time zone 'tomorrow';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) AS One FROM TIMESTAMPTZ_TBL WHERE d1 = timestamp with time zone 'yesterday';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) AS One FROM TIMESTAMPTZ_TBL WHERE d1 = timestamp with time zone 'tomorrow EST';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT count(*) AS One FROM TIMESTAMPTZ_TBL WHERE d1 = timestamp with time zone 'tomorrow zulu';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DELETE FROM TIMESTAMPTZ_TBL;`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('now');`,
			},
			{
				Statement: `SELECT pg_sleep(0.1);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('now');`,
			},
			{
				Statement: `SELECT pg_sleep(0.1);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('now');`,
			},
			{
				Statement: `SELECT pg_sleep(0.1);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT count(*) AS two FROM TIMESTAMPTZ_TBL WHERE d1 = timestamp(2) with time zone 'now';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT count(d1) AS three, count(DISTINCT d1) AS two FROM TIMESTAMPTZ_TBL;`,
				Results:   []sql.Row{{3, 2}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `TRUNCATE TIMESTAMPTZ_TBL;`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('-infinity');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('infinity');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('epoch');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Mon Feb 10 17:32:01 1997 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Mon Feb 10 17:32:01.000001 1997 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Mon Feb 10 17:32:01.999999 1997 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Mon Feb 10 17:32:01.4 1997 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Mon Feb 10 17:32:01.5 1997 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Mon Feb 10 17:32:01.6 1997 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997-01-02');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997-01-02 03:04:05');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997-02-10 17:32:01-08');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997-02-10 17:32:01-0800');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997-02-10 17:32:01 -08:00');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('19970210 173201 -0800');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997-06-10 17:32:01 -07:00');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('2001-09-22T18:19:20');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('2000-03-15 08:14:01 GMT+8');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('2000-03-15 13:14:02 GMT-1');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('2000-03-15 12:14:03 GMT-2');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('2000-03-15 03:14:04 PST+8');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('2000-03-15 02:14:05 MST+7:00');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 10 17:32:01 1997 -0800');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 10 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 10 5:32PM 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997/02/10 17:32:01-0800');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997-02-10 17:32:01 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb-10-1997 17:32:01 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('02-10-1997 17:32:01 PST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('19970210 173201 PST');`,
			},
			{
				Statement: `set datestyle to ymd;`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('97FEB10 5:32:01PM UTC');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('97/02/10 17:32:01 UTC');`,
			},
			{
				Statement: `reset datestyle;`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997.041 17:32:01 UTC');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('19970210 173201 America/New_York');`,
			},
			{
				Statement: `SELECT '19970210 173201' AT TIME ZONE 'America/New_York';`,
				Results:   []sql.Row{{`Mon Feb 10 20:32:01 1997`}},
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('19970710 173201 America/New_York');`,
			},
			{
				Statement: `SELECT '19970710 173201' AT TIME ZONE 'America/New_York';`,
				Results:   []sql.Row{{`Thu Jul 10 20:32:01 1997`}},
			},
			{
				Statement:   `INSERT INTO TIMESTAMPTZ_TBL VALUES ('19970710 173201 America/Does_not_exist');`,
				ErrorString: `time zone "america/does_not_exist" not recognized`,
			},
			{
				Statement:   `SELECT '19970710 173201' AT TIME ZONE 'America/Does_not_exist';`,
				ErrorString: `time zone "America/Does_not_exist" not recognized`,
			},
			{
				Statement: `SELECT '20500710 173201 Europe/Helsinki'::timestamptz; -- DST`,
				Results:   []sql.Row{{`Sun Jul 10 07:32:01 2050 PDT`}},
			},
			{
				Statement: `SELECT '20500110 173201 Europe/Helsinki'::timestamptz; -- non-DST`,
				Results:   []sql.Row{{`Mon Jan 10 07:32:01 2050 PST`}},
			},
			{
				Statement: `SELECT '205000-07-10 17:32:01 Europe/Helsinki'::timestamptz; -- DST`,
				Results:   []sql.Row{{`Thu Jul 10 07:32:01 205000 PDT`}},
			},
			{
				Statement: `SELECT '205000-01-10 17:32:01 Europe/Helsinki'::timestamptz; -- non-DST`,
				Results:   []sql.Row{{`Fri Jan 10 07:32:01 205000 PST`}},
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('1997-06-10 18:32:01 PDT');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 10 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 11 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 12 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 13 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 14 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 15 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 0097 BC');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 0097');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 0597');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 1097');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 1697');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 1797');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 1897');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 2097');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 28 17:32:01 1996');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 29 17:32:01 1996');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Mar 01 17:32:01 1996');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Dec 30 17:32:01 1996');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Dec 31 17:32:01 1996');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Jan 01 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 28 17:32:01 1997');`,
			},
			{
				Statement:   `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 29 17:32:01 1997');`,
				ErrorString: `date/time field value out of range: "Feb 29 17:32:01 1997"`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Mar 01 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Dec 30 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Dec 31 17:32:01 1997');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Dec 31 17:32:01 1999');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Jan 01 17:32:01 2000');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Dec 31 17:32:01 2000');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Jan 01 17:32:01 2001');`,
			},
			{
				Statement:   `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 -0097');`,
				ErrorString: `time zone displacement out of range: "Feb 16 17:32:01 -0097"`,
			},
			{
				Statement:   `INSERT INTO TIMESTAMPTZ_TBL VALUES ('Feb 16 17:32:01 5097 BC');`,
				ErrorString: `timestamp out of range: "Feb 16 17:32:01 5097 BC"`,
			},
			{
				Statement: `SELECT 'Wed Jul 11 10:51:14 America/New_York 2001'::timestamptz;`,
				Results:   []sql.Row{{`Wed Jul 11 07:51:14 2001 PDT`}},
			},
			{
				Statement: `SELECT 'Wed Jul 11 10:51:14 GMT-4 2001'::timestamptz;`,
				Results:   []sql.Row{{`Tue Jul 10 23:51:14 2001 PDT`}},
			},
			{
				Statement: `SELECT 'Wed Jul 11 10:51:14 GMT+4 2001'::timestamptz;`,
				Results:   []sql.Row{{`Wed Jul 11 07:51:14 2001 PDT`}},
			},
			{
				Statement: `SELECT 'Wed Jul 11 10:51:14 PST-03:00 2001'::timestamptz;`,
				Results:   []sql.Row{{`Wed Jul 11 00:51:14 2001 PDT`}},
			},
			{
				Statement: `SELECT 'Wed Jul 11 10:51:14 PST+03:00 2001'::timestamptz;`,
				Results:   []sql.Row{{`Wed Jul 11 06:51:14 2001 PDT`}},
			},
			{
				Statement: `SELECT d1 FROM TIMESTAMPTZ_TBL;`,
				Results:   []sql.Row{{`-infinity`}, {`infinity`}, {`Wed Dec 31 16:00:00 1969 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:02 1997 PST`}, {`Mon Feb 10 17:32:01.4 1997 PST`}, {`Mon Feb 10 17:32:01.5 1997 PST`}, {`Mon Feb 10 17:32:01.6 1997 PST`}, {`Thu Jan 02 00:00:00 1997 PST`}, {`Thu Jan 02 03:04:05 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Tue Jun 10 17:32:01 1997 PDT`}, {`Sat Sep 22 18:19:20 2001 PDT`}, {`Wed Mar 15 08:14:01 2000 PST`}, {`Wed Mar 15 04:14:02 2000 PST`}, {`Wed Mar 15 02:14:03 2000 PST`}, {`Wed Mar 15 03:14:04 2000 PST`}, {`Wed Mar 15 01:14:05 2000 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:00 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 14:32:01 1997 PST`}, {`Thu Jul 10 14:32:01 1997 PDT`}, {`Tue Jun 10 18:32:01 1997 PDT`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Tue Feb 11 17:32:01 1997 PST`}, {`Wed Feb 12 17:32:01 1997 PST`}, {`Thu Feb 13 17:32:01 1997 PST`}, {`Fri Feb 14 17:32:01 1997 PST`}, {`Sat Feb 15 17:32:01 1997 PST`}, {`Sun Feb 16 17:32:01 1997 PST`}, {`Tue Feb 16 17:32:01 0097 PST BC`}, {`Sat Feb 16 17:32:01 0097 PST`}, {`Thu Feb 16 17:32:01 0597 PST`}, {`Tue Feb 16 17:32:01 1097 PST`}, {`Sat Feb 16 17:32:01 1697 PST`}, {`Thu Feb 16 17:32:01 1797 PST`}, {`Tue Feb 16 17:32:01 1897 PST`}, {`Sun Feb 16 17:32:01 1997 PST`}, {`Sat Feb 16 17:32:01 2097 PST`}, {`Wed Feb 28 17:32:01 1996 PST`}, {`Thu Feb 29 17:32:01 1996 PST`}, {`Fri Mar 01 17:32:01 1996 PST`}, {`Mon Dec 30 17:32:01 1996 PST`}, {`Tue Dec 31 17:32:01 1996 PST`}, {`Wed Jan 01 17:32:01 1997 PST`}, {`Fri Feb 28 17:32:01 1997 PST`}, {`Sat Mar 01 17:32:01 1997 PST`}, {`Tue Dec 30 17:32:01 1997 PST`}, {`Wed Dec 31 17:32:01 1997 PST`}, {`Fri Dec 31 17:32:01 1999 PST`}, {`Sat Jan 01 17:32:01 2000 PST`}, {`Sun Dec 31 17:32:01 2000 PST`}, {`Mon Jan 01 17:32:01 2001 PST`}},
			},
			{
				Statement: `SELECT '4714-11-24 00:00:00+00 BC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Nov 23 16:00:00 4714 PST BC`}},
			},
			{
				Statement: `SELECT '4714-11-23 16:00:00-08 BC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Nov 23 16:00:00 4714 PST BC`}},
			},
			{
				Statement: `SELECT 'Sun Nov 23 16:00:00 4714 PST BC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Nov 23 16:00:00 4714 PST BC`}},
			},
			{
				Statement:   `SELECT '4714-11-23 23:59:59+00 BC'::timestamptz;  -- out of range`,
				ErrorString: `timestamp out of range: "4714-11-23 23:59:59+00 BC"`,
			},
			{
				Statement: `SELECT '294276-12-31 23:59:59+00'::timestamptz;`,
				Results:   []sql.Row{{`Sun Dec 31 15:59:59 294276 PST`}},
			},
			{
				Statement: `SELECT '294276-12-31 15:59:59-08'::timestamptz;`,
				Results:   []sql.Row{{`Sun Dec 31 15:59:59 294276 PST`}},
			},
			{
				Statement:   `SELECT '294277-01-01 00:00:00+00'::timestamptz;  -- out of range`,
				ErrorString: `timestamp out of range: "294277-01-01 00:00:00+00"`,
			},
			{
				Statement:   `SELECT '294277-12-31 16:00:00-08'::timestamptz;  -- out of range`,
				ErrorString: `timestamp out of range: "294277-12-31 16:00:00-08"`,
			},
			{
				Statement: `SELECT d1 FROM TIMESTAMPTZ_TBL
   WHERE d1 > timestamp with time zone '1997-01-02';`,
				Results: []sql.Row{{`infinity`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:02 1997 PST`}, {`Mon Feb 10 17:32:01.4 1997 PST`}, {`Mon Feb 10 17:32:01.5 1997 PST`}, {`Mon Feb 10 17:32:01.6 1997 PST`}, {`Thu Jan 02 03:04:05 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Tue Jun 10 17:32:01 1997 PDT`}, {`Sat Sep 22 18:19:20 2001 PDT`}, {`Wed Mar 15 08:14:01 2000 PST`}, {`Wed Mar 15 04:14:02 2000 PST`}, {`Wed Mar 15 02:14:03 2000 PST`}, {`Wed Mar 15 03:14:04 2000 PST`}, {`Wed Mar 15 01:14:05 2000 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:00 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 14:32:01 1997 PST`}, {`Thu Jul 10 14:32:01 1997 PDT`}, {`Tue Jun 10 18:32:01 1997 PDT`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Tue Feb 11 17:32:01 1997 PST`}, {`Wed Feb 12 17:32:01 1997 PST`}, {`Thu Feb 13 17:32:01 1997 PST`}, {`Fri Feb 14 17:32:01 1997 PST`}, {`Sat Feb 15 17:32:01 1997 PST`}, {`Sun Feb 16 17:32:01 1997 PST`}, {`Sun Feb 16 17:32:01 1997 PST`}, {`Sat Feb 16 17:32:01 2097 PST`}, {`Fri Feb 28 17:32:01 1997 PST`}, {`Sat Mar 01 17:32:01 1997 PST`}, {`Tue Dec 30 17:32:01 1997 PST`}, {`Wed Dec 31 17:32:01 1997 PST`}, {`Fri Dec 31 17:32:01 1999 PST`}, {`Sat Jan 01 17:32:01 2000 PST`}, {`Sun Dec 31 17:32:01 2000 PST`}, {`Mon Jan 01 17:32:01 2001 PST`}},
			},
			{
				Statement: `SELECT d1 FROM TIMESTAMPTZ_TBL
   WHERE d1 < timestamp with time zone '1997-01-02';`,
				Results: []sql.Row{{`-infinity`}, {`Wed Dec 31 16:00:00 1969 PST`}, {`Tue Feb 16 17:32:01 0097 PST BC`}, {`Sat Feb 16 17:32:01 0097 PST`}, {`Thu Feb 16 17:32:01 0597 PST`}, {`Tue Feb 16 17:32:01 1097 PST`}, {`Sat Feb 16 17:32:01 1697 PST`}, {`Thu Feb 16 17:32:01 1797 PST`}, {`Tue Feb 16 17:32:01 1897 PST`}, {`Wed Feb 28 17:32:01 1996 PST`}, {`Thu Feb 29 17:32:01 1996 PST`}, {`Fri Mar 01 17:32:01 1996 PST`}, {`Mon Dec 30 17:32:01 1996 PST`}, {`Tue Dec 31 17:32:01 1996 PST`}, {`Wed Jan 01 17:32:01 1997 PST`}},
			},
			{
				Statement: `SELECT d1 FROM TIMESTAMPTZ_TBL
   WHERE d1 = timestamp with time zone '1997-01-02';`,
				Results: []sql.Row{{`Thu Jan 02 00:00:00 1997 PST`}},
			},
			{
				Statement: `SELECT d1 FROM TIMESTAMPTZ_TBL
   WHERE d1 != timestamp with time zone '1997-01-02';`,
				Results: []sql.Row{{`-infinity`}, {`infinity`}, {`Wed Dec 31 16:00:00 1969 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:02 1997 PST`}, {`Mon Feb 10 17:32:01.4 1997 PST`}, {`Mon Feb 10 17:32:01.5 1997 PST`}, {`Mon Feb 10 17:32:01.6 1997 PST`}, {`Thu Jan 02 03:04:05 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Tue Jun 10 17:32:01 1997 PDT`}, {`Sat Sep 22 18:19:20 2001 PDT`}, {`Wed Mar 15 08:14:01 2000 PST`}, {`Wed Mar 15 04:14:02 2000 PST`}, {`Wed Mar 15 02:14:03 2000 PST`}, {`Wed Mar 15 03:14:04 2000 PST`}, {`Wed Mar 15 01:14:05 2000 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:00 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 14:32:01 1997 PST`}, {`Thu Jul 10 14:32:01 1997 PDT`}, {`Tue Jun 10 18:32:01 1997 PDT`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Tue Feb 11 17:32:01 1997 PST`}, {`Wed Feb 12 17:32:01 1997 PST`}, {`Thu Feb 13 17:32:01 1997 PST`}, {`Fri Feb 14 17:32:01 1997 PST`}, {`Sat Feb 15 17:32:01 1997 PST`}, {`Sun Feb 16 17:32:01 1997 PST`}, {`Tue Feb 16 17:32:01 0097 PST BC`}, {`Sat Feb 16 17:32:01 0097 PST`}, {`Thu Feb 16 17:32:01 0597 PST`}, {`Tue Feb 16 17:32:01 1097 PST`}, {`Sat Feb 16 17:32:01 1697 PST`}, {`Thu Feb 16 17:32:01 1797 PST`}, {`Tue Feb 16 17:32:01 1897 PST`}, {`Sun Feb 16 17:32:01 1997 PST`}, {`Sat Feb 16 17:32:01 2097 PST`}, {`Wed Feb 28 17:32:01 1996 PST`}, {`Thu Feb 29 17:32:01 1996 PST`}, {`Fri Mar 01 17:32:01 1996 PST`}, {`Mon Dec 30 17:32:01 1996 PST`}, {`Tue Dec 31 17:32:01 1996 PST`}, {`Wed Jan 01 17:32:01 1997 PST`}, {`Fri Feb 28 17:32:01 1997 PST`}, {`Sat Mar 01 17:32:01 1997 PST`}, {`Tue Dec 30 17:32:01 1997 PST`}, {`Wed Dec 31 17:32:01 1997 PST`}, {`Fri Dec 31 17:32:01 1999 PST`}, {`Sat Jan 01 17:32:01 2000 PST`}, {`Sun Dec 31 17:32:01 2000 PST`}, {`Mon Jan 01 17:32:01 2001 PST`}},
			},
			{
				Statement: `SELECT d1 FROM TIMESTAMPTZ_TBL
   WHERE d1 <= timestamp with time zone '1997-01-02';`,
				Results: []sql.Row{{`-infinity`}, {`Wed Dec 31 16:00:00 1969 PST`}, {`Thu Jan 02 00:00:00 1997 PST`}, {`Tue Feb 16 17:32:01 0097 PST BC`}, {`Sat Feb 16 17:32:01 0097 PST`}, {`Thu Feb 16 17:32:01 0597 PST`}, {`Tue Feb 16 17:32:01 1097 PST`}, {`Sat Feb 16 17:32:01 1697 PST`}, {`Thu Feb 16 17:32:01 1797 PST`}, {`Tue Feb 16 17:32:01 1897 PST`}, {`Wed Feb 28 17:32:01 1996 PST`}, {`Thu Feb 29 17:32:01 1996 PST`}, {`Fri Mar 01 17:32:01 1996 PST`}, {`Mon Dec 30 17:32:01 1996 PST`}, {`Tue Dec 31 17:32:01 1996 PST`}, {`Wed Jan 01 17:32:01 1997 PST`}},
			},
			{
				Statement: `SELECT d1 FROM TIMESTAMPTZ_TBL
   WHERE d1 >= timestamp with time zone '1997-01-02';`,
				Results: []sql.Row{{`infinity`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:02 1997 PST`}, {`Mon Feb 10 17:32:01.4 1997 PST`}, {`Mon Feb 10 17:32:01.5 1997 PST`}, {`Mon Feb 10 17:32:01.6 1997 PST`}, {`Thu Jan 02 00:00:00 1997 PST`}, {`Thu Jan 02 03:04:05 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Tue Jun 10 17:32:01 1997 PDT`}, {`Sat Sep 22 18:19:20 2001 PDT`}, {`Wed Mar 15 08:14:01 2000 PST`}, {`Wed Mar 15 04:14:02 2000 PST`}, {`Wed Mar 15 02:14:03 2000 PST`}, {`Wed Mar 15 03:14:04 2000 PST`}, {`Wed Mar 15 01:14:05 2000 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:00 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 09:32:01 1997 PST`}, {`Mon Feb 10 14:32:01 1997 PST`}, {`Thu Jul 10 14:32:01 1997 PDT`}, {`Tue Jun 10 18:32:01 1997 PDT`}, {`Mon Feb 10 17:32:01 1997 PST`}, {`Tue Feb 11 17:32:01 1997 PST`}, {`Wed Feb 12 17:32:01 1997 PST`}, {`Thu Feb 13 17:32:01 1997 PST`}, {`Fri Feb 14 17:32:01 1997 PST`}, {`Sat Feb 15 17:32:01 1997 PST`}, {`Sun Feb 16 17:32:01 1997 PST`}, {`Sun Feb 16 17:32:01 1997 PST`}, {`Sat Feb 16 17:32:01 2097 PST`}, {`Fri Feb 28 17:32:01 1997 PST`}, {`Sat Mar 01 17:32:01 1997 PST`}, {`Tue Dec 30 17:32:01 1997 PST`}, {`Wed Dec 31 17:32:01 1997 PST`}, {`Fri Dec 31 17:32:01 1999 PST`}, {`Sat Jan 01 17:32:01 2000 PST`}, {`Sun Dec 31 17:32:01 2000 PST`}, {`Mon Jan 01 17:32:01 2001 PST`}},
			},
			{
				Statement: `SELECT d1 - timestamp with time zone '1997-01-02' AS diff
   FROM TIMESTAMPTZ_TBL WHERE d1 BETWEEN '1902-01-01' AND '2038-01-01';`,
				Results: []sql.Row{{`@ 9863 days 8 hours ago`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 2 secs`}, {`@ 39 days 17 hours 32 mins 1.4 secs`}, {`@ 39 days 17 hours 32 mins 1.5 secs`}, {`@ 39 days 17 hours 32 mins 1.6 secs`}, {`@ 0`}, {`@ 3 hours 4 mins 5 secs`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 159 days 16 hours 32 mins 1 sec`}, {`@ 1724 days 17 hours 19 mins 20 secs`}, {`@ 1168 days 8 hours 14 mins 1 sec`}, {`@ 1168 days 4 hours 14 mins 2 secs`}, {`@ 1168 days 2 hours 14 mins 3 secs`}, {`@ 1168 days 3 hours 14 mins 4 secs`}, {`@ 1168 days 1 hour 14 mins 5 secs`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 9 hours 32 mins 1 sec`}, {`@ 39 days 9 hours 32 mins 1 sec`}, {`@ 39 days 9 hours 32 mins 1 sec`}, {`@ 39 days 14 hours 32 mins 1 sec`}, {`@ 189 days 13 hours 32 mins 1 sec`}, {`@ 159 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 40 days 17 hours 32 mins 1 sec`}, {`@ 41 days 17 hours 32 mins 1 sec`}, {`@ 42 days 17 hours 32 mins 1 sec`}, {`@ 43 days 17 hours 32 mins 1 sec`}, {`@ 44 days 17 hours 32 mins 1 sec`}, {`@ 45 days 17 hours 32 mins 1 sec`}, {`@ 45 days 17 hours 32 mins 1 sec`}, {`@ 308 days 6 hours 27 mins 59 secs ago`}, {`@ 307 days 6 hours 27 mins 59 secs ago`}, {`@ 306 days 6 hours 27 mins 59 secs ago`}, {`@ 2 days 6 hours 27 mins 59 secs ago`}, {`@ 1 day 6 hours 27 mins 59 secs ago`}, {`@ 6 hours 27 mins 59 secs ago`}, {`@ 57 days 17 hours 32 mins 1 sec`}, {`@ 58 days 17 hours 32 mins 1 sec`}, {`@ 362 days 17 hours 32 mins 1 sec`}, {`@ 363 days 17 hours 32 mins 1 sec`}, {`@ 1093 days 17 hours 32 mins 1 sec`}, {`@ 1094 days 17 hours 32 mins 1 sec`}, {`@ 1459 days 17 hours 32 mins 1 sec`}, {`@ 1460 days 17 hours 32 mins 1 sec`}},
			},
			{
				Statement: `SELECT date_trunc( 'week', timestamp with time zone '2004-02-29 15:44:17.71393' ) AS week_trunc;`,
				Results:   []sql.Row{{`Mon Feb 23 00:00:00 2004 PST`}},
			},
			{
				Statement: `SELECT date_trunc('day', timestamp with time zone '2001-02-16 20:38:40+00', 'Australia/Sydney') as sydney_trunc;  -- zone name`,
				Results:   []sql.Row{{`Fri Feb 16 05:00:00 2001 PST`}},
			},
			{
				Statement: `SELECT date_trunc('day', timestamp with time zone '2001-02-16 20:38:40+00', 'GMT') as gmt_trunc;  -- fixed-offset abbreviation`,
				Results:   []sql.Row{{`Thu Feb 15 16:00:00 2001 PST`}},
			},
			{
				Statement: `SELECT date_trunc('day', timestamp with time zone '2001-02-16 20:38:40+00', 'VET') as vet_trunc;  -- variable-offset abbreviation`,
				Results:   []sql.Row{{`Thu Feb 15 20:00:00 2001 PST`}},
			},
			{
				Statement: `SELECT
  str,
  interval,
  date_trunc(str, ts, 'Australia/Sydney') = date_bin(interval::interval, ts, timestamp with time zone '2001-01-01+11') AS equal
FROM (
  VALUES
  ('day', '1 d'),
  ('hour', '1 h'),
  ('minute', '1 m'),
  ('second', '1 s'),
  ('millisecond', '1 ms'),
  ('microsecond', '1 us')
) intervals (str, interval),
(VALUES (timestamptz '2020-02-29 15:44:17.71393+00')) ts (ts);`,
				Results: []sql.Row{{`day`, `1 d`, true}, {`hour`, `1 h`, true}, {`minute`, `1 m`, true}, {`second`, `1 s`, true}, {`millisecond`, `1 ms`, true}, {`microsecond`, `1 us`, true}},
			},
			{
				Statement: `SELECT
  interval,
  ts,
  origin,
  date_bin(interval::interval, ts, origin)
FROM (
  VALUES
  ('15 days'),
  ('2 hours'),
  ('1 hour 30 minutes'),
  ('15 minutes'),
  ('10 seconds'),
  ('100 milliseconds'),
  ('250 microseconds')
) intervals (interval),
(VALUES (timestamptz '2020-02-11 15:44:17.71393')) ts (ts),
(VALUES (timestamptz '2001-01-01')) origin (origin);`,
				Results: []sql.Row{{`15 days`, `Tue Feb 11 15:44:17.71393 2020 PST`, `Mon Jan 01 00:00:00 2001 PST`, `Thu Feb 06 00:00:00 2020 PST`}, {`2 hours`, `Tue Feb 11 15:44:17.71393 2020 PST`, `Mon Jan 01 00:00:00 2001 PST`, `Tue Feb 11 14:00:00 2020 PST`}, {`1 hour 30 minutes`, `Tue Feb 11 15:44:17.71393 2020 PST`, `Mon Jan 01 00:00:00 2001 PST`, `Tue Feb 11 15:00:00 2020 PST`}, {`15 minutes`, `Tue Feb 11 15:44:17.71393 2020 PST`, `Mon Jan 01 00:00:00 2001 PST`, `Tue Feb 11 15:30:00 2020 PST`}, {`10 seconds`, `Tue Feb 11 15:44:17.71393 2020 PST`, `Mon Jan 01 00:00:00 2001 PST`, `Tue Feb 11 15:44:10 2020 PST`}, {`100 milliseconds`, `Tue Feb 11 15:44:17.71393 2020 PST`, `Mon Jan 01 00:00:00 2001 PST`, `Tue Feb 11 15:44:17.7 2020 PST`}, {`250 microseconds`, `Tue Feb 11 15:44:17.71393 2020 PST`, `Mon Jan 01 00:00:00 2001 PST`, `Tue Feb 11 15:44:17.71375 2020 PST`}},
			},
			{
				Statement: `SELECT date_bin('5 min'::interval, timestamptz '2020-02-01 01:01:01+00', timestamptz '2020-02-01 00:02:30+00');`,
				Results:   []sql.Row{{`Fri Jan 31 16:57:30 2020 PST`}},
			},
			{
				Statement:   `SELECT date_bin('5 months'::interval, timestamp with time zone '2020-02-01 01:01:01+00', timestamp with time zone '2001-01-01+00');`,
				ErrorString: `timestamps cannot be binned into intervals containing months or years`,
			},
			{
				Statement:   `SELECT date_bin('5 years'::interval,  timestamp with time zone '2020-02-01 01:01:01+00', timestamp with time zone '2001-01-01+00');`,
				ErrorString: `timestamps cannot be binned into intervals containing months or years`,
			},
			{
				Statement:   `SELECT date_bin('0 days'::interval, timestamp with time zone '1970-01-01 01:00:00+00' , timestamp with time zone '1970-01-01 00:00:00+00');`,
				ErrorString: `stride must be greater than zero`,
			},
			{
				Statement:   `SELECT date_bin('-2 days'::interval, timestamp with time zone '1970-01-01 01:00:00+00' , timestamp with time zone '1970-01-01 00:00:00+00');`,
				ErrorString: `stride must be greater than zero`,
			},
			{
				Statement: `SELECT d1 - timestamp with time zone '1997-01-02' AS diff
  FROM TIMESTAMPTZ_TBL
  WHERE d1 BETWEEN timestamp with time zone '1902-01-01' AND timestamp with time zone '2038-01-01';`,
				Results: []sql.Row{{`@ 9863 days 8 hours ago`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 2 secs`}, {`@ 39 days 17 hours 32 mins 1.4 secs`}, {`@ 39 days 17 hours 32 mins 1.5 secs`}, {`@ 39 days 17 hours 32 mins 1.6 secs`}, {`@ 0`}, {`@ 3 hours 4 mins 5 secs`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 159 days 16 hours 32 mins 1 sec`}, {`@ 1724 days 17 hours 19 mins 20 secs`}, {`@ 1168 days 8 hours 14 mins 1 sec`}, {`@ 1168 days 4 hours 14 mins 2 secs`}, {`@ 1168 days 2 hours 14 mins 3 secs`}, {`@ 1168 days 3 hours 14 mins 4 secs`}, {`@ 1168 days 1 hour 14 mins 5 secs`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 39 days 9 hours 32 mins 1 sec`}, {`@ 39 days 9 hours 32 mins 1 sec`}, {`@ 39 days 9 hours 32 mins 1 sec`}, {`@ 39 days 14 hours 32 mins 1 sec`}, {`@ 189 days 13 hours 32 mins 1 sec`}, {`@ 159 days 17 hours 32 mins 1 sec`}, {`@ 39 days 17 hours 32 mins 1 sec`}, {`@ 40 days 17 hours 32 mins 1 sec`}, {`@ 41 days 17 hours 32 mins 1 sec`}, {`@ 42 days 17 hours 32 mins 1 sec`}, {`@ 43 days 17 hours 32 mins 1 sec`}, {`@ 44 days 17 hours 32 mins 1 sec`}, {`@ 45 days 17 hours 32 mins 1 sec`}, {`@ 45 days 17 hours 32 mins 1 sec`}, {`@ 308 days 6 hours 27 mins 59 secs ago`}, {`@ 307 days 6 hours 27 mins 59 secs ago`}, {`@ 306 days 6 hours 27 mins 59 secs ago`}, {`@ 2 days 6 hours 27 mins 59 secs ago`}, {`@ 1 day 6 hours 27 mins 59 secs ago`}, {`@ 6 hours 27 mins 59 secs ago`}, {`@ 57 days 17 hours 32 mins 1 sec`}, {`@ 58 days 17 hours 32 mins 1 sec`}, {`@ 362 days 17 hours 32 mins 1 sec`}, {`@ 363 days 17 hours 32 mins 1 sec`}, {`@ 1093 days 17 hours 32 mins 1 sec`}, {`@ 1094 days 17 hours 32 mins 1 sec`}, {`@ 1459 days 17 hours 32 mins 1 sec`}, {`@ 1460 days 17 hours 32 mins 1 sec`}},
			},
			{
				Statement: `SELECT d1 as timestamptz,
   date_part( 'year', d1) AS year, date_part( 'month', d1) AS month,
   date_part( 'day', d1) AS day, date_part( 'hour', d1) AS hour,
   date_part( 'minute', d1) AS minute, date_part( 'second', d1) AS second
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{`-infinity`, `-Infinity`, ``, ``, ``, ``, ``}, {`infinity`, `Infinity`, ``, ``, ``, ``, ``}, {`Wed Dec 31 16:00:00 1969 PST`, 1969, 12, 31, 16, 0, 0}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:02 1997 PST`, 1997, 2, 10, 17, 32, 2}, {`Mon Feb 10 17:32:01.4 1997 PST`, 1997, 2, 10, 17, 32, 1.4}, {`Mon Feb 10 17:32:01.5 1997 PST`, 1997, 2, 10, 17, 32, 1.5}, {`Mon Feb 10 17:32:01.6 1997 PST`, 1997, 2, 10, 17, 32, 1.6}, {`Thu Jan 02 00:00:00 1997 PST`, 1997, 1, 2, 0, 0, 0}, {`Thu Jan 02 03:04:05 1997 PST`, 1997, 1, 2, 3, 4, 5}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Tue Jun 10 17:32:01 1997 PDT`, 1997, 6, 10, 17, 32, 1}, {`Sat Sep 22 18:19:20 2001 PDT`, 2001, 9, 22, 18, 19, 20}, {`Wed Mar 15 08:14:01 2000 PST`, 2000, 3, 15, 8, 14, 1}, {`Wed Mar 15 04:14:02 2000 PST`, 2000, 3, 15, 4, 14, 2}, {`Wed Mar 15 02:14:03 2000 PST`, 2000, 3, 15, 2, 14, 3}, {`Wed Mar 15 03:14:04 2000 PST`, 2000, 3, 15, 3, 14, 4}, {`Wed Mar 15 01:14:05 2000 PST`, 2000, 3, 15, 1, 14, 5}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:00 1997 PST`, 1997, 2, 10, 17, 32, 0}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Mon Feb 10 09:32:01 1997 PST`, 1997, 2, 10, 9, 32, 1}, {`Mon Feb 10 09:32:01 1997 PST`, 1997, 2, 10, 9, 32, 1}, {`Mon Feb 10 09:32:01 1997 PST`, 1997, 2, 10, 9, 32, 1}, {`Mon Feb 10 14:32:01 1997 PST`, 1997, 2, 10, 14, 32, 1}, {`Thu Jul 10 14:32:01 1997 PDT`, 1997, 7, 10, 14, 32, 1}, {`Tue Jun 10 18:32:01 1997 PDT`, 1997, 6, 10, 18, 32, 1}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 2, 10, 17, 32, 1}, {`Tue Feb 11 17:32:01 1997 PST`, 1997, 2, 11, 17, 32, 1}, {`Wed Feb 12 17:32:01 1997 PST`, 1997, 2, 12, 17, 32, 1}, {`Thu Feb 13 17:32:01 1997 PST`, 1997, 2, 13, 17, 32, 1}, {`Fri Feb 14 17:32:01 1997 PST`, 1997, 2, 14, 17, 32, 1}, {`Sat Feb 15 17:32:01 1997 PST`, 1997, 2, 15, 17, 32, 1}, {`Sun Feb 16 17:32:01 1997 PST`, 1997, 2, 16, 17, 32, 1}, {`Tue Feb 16 17:32:01 0097 PST BC`, -97, 2, 16, 17, 32, 1}, {`Sat Feb 16 17:32:01 0097 PST`, 97, 2, 16, 17, 32, 1}, {`Thu Feb 16 17:32:01 0597 PST`, 597, 2, 16, 17, 32, 1}, {`Tue Feb 16 17:32:01 1097 PST`, 1097, 2, 16, 17, 32, 1}, {`Sat Feb 16 17:32:01 1697 PST`, 1697, 2, 16, 17, 32, 1}, {`Thu Feb 16 17:32:01 1797 PST`, 1797, 2, 16, 17, 32, 1}, {`Tue Feb 16 17:32:01 1897 PST`, 1897, 2, 16, 17, 32, 1}, {`Sun Feb 16 17:32:01 1997 PST`, 1997, 2, 16, 17, 32, 1}, {`Sat Feb 16 17:32:01 2097 PST`, 2097, 2, 16, 17, 32, 1}, {`Wed Feb 28 17:32:01 1996 PST`, 1996, 2, 28, 17, 32, 1}, {`Thu Feb 29 17:32:01 1996 PST`, 1996, 2, 29, 17, 32, 1}, {`Fri Mar 01 17:32:01 1996 PST`, 1996, 3, 1, 17, 32, 1}, {`Mon Dec 30 17:32:01 1996 PST`, 1996, 12, 30, 17, 32, 1}, {`Tue Dec 31 17:32:01 1996 PST`, 1996, 12, 31, 17, 32, 1}, {`Wed Jan 01 17:32:01 1997 PST`, 1997, 1, 1, 17, 32, 1}, {`Fri Feb 28 17:32:01 1997 PST`, 1997, 2, 28, 17, 32, 1}, {`Sat Mar 01 17:32:01 1997 PST`, 1997, 3, 1, 17, 32, 1}, {`Tue Dec 30 17:32:01 1997 PST`, 1997, 12, 30, 17, 32, 1}, {`Wed Dec 31 17:32:01 1997 PST`, 1997, 12, 31, 17, 32, 1}, {`Fri Dec 31 17:32:01 1999 PST`, 1999, 12, 31, 17, 32, 1}, {`Sat Jan 01 17:32:01 2000 PST`, 2000, 1, 1, 17, 32, 1}, {`Sun Dec 31 17:32:01 2000 PST`, 2000, 12, 31, 17, 32, 1}, {`Mon Jan 01 17:32:01 2001 PST`, 2001, 1, 1, 17, 32, 1}},
			},
			{
				Statement: `SELECT d1 as timestamptz,
   date_part( 'quarter', d1) AS quarter, date_part( 'msec', d1) AS msec,
   date_part( 'usec', d1) AS usec
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{`-infinity`, ``, ``, ``}, {`infinity`, ``, ``, ``}, {`Wed Dec 31 16:00:00 1969 PST`, 4, 0, 0}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:02 1997 PST`, 1, 2000, 2000000}, {`Mon Feb 10 17:32:01.4 1997 PST`, 1, 1400, 1400000}, {`Mon Feb 10 17:32:01.5 1997 PST`, 1, 1500, 1500000}, {`Mon Feb 10 17:32:01.6 1997 PST`, 1, 1600, 1600000}, {`Thu Jan 02 00:00:00 1997 PST`, 1, 0, 0}, {`Thu Jan 02 03:04:05 1997 PST`, 1, 5000, 5000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Tue Jun 10 17:32:01 1997 PDT`, 2, 1000, 1000000}, {`Sat Sep 22 18:19:20 2001 PDT`, 3, 20000, 20000000}, {`Wed Mar 15 08:14:01 2000 PST`, 1, 1000, 1000000}, {`Wed Mar 15 04:14:02 2000 PST`, 1, 2000, 2000000}, {`Wed Mar 15 02:14:03 2000 PST`, 1, 3000, 3000000}, {`Wed Mar 15 03:14:04 2000 PST`, 1, 4000, 4000000}, {`Wed Mar 15 01:14:05 2000 PST`, 1, 5000, 5000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:00 1997 PST`, 1, 0, 0}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 09:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 09:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 09:32:01 1997 PST`, 1, 1000, 1000000}, {`Mon Feb 10 14:32:01 1997 PST`, 1, 1000, 1000000}, {`Thu Jul 10 14:32:01 1997 PDT`, 3, 1000, 1000000}, {`Tue Jun 10 18:32:01 1997 PDT`, 2, 1000, 1000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Tue Feb 11 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Wed Feb 12 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Thu Feb 13 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Fri Feb 14 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Sat Feb 15 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Sun Feb 16 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Tue Feb 16 17:32:01 0097 PST BC`, 1, 1000, 1000000}, {`Sat Feb 16 17:32:01 0097 PST`, 1, 1000, 1000000}, {`Thu Feb 16 17:32:01 0597 PST`, 1, 1000, 1000000}, {`Tue Feb 16 17:32:01 1097 PST`, 1, 1000, 1000000}, {`Sat Feb 16 17:32:01 1697 PST`, 1, 1000, 1000000}, {`Thu Feb 16 17:32:01 1797 PST`, 1, 1000, 1000000}, {`Tue Feb 16 17:32:01 1897 PST`, 1, 1000, 1000000}, {`Sun Feb 16 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Sat Feb 16 17:32:01 2097 PST`, 1, 1000, 1000000}, {`Wed Feb 28 17:32:01 1996 PST`, 1, 1000, 1000000}, {`Thu Feb 29 17:32:01 1996 PST`, 1, 1000, 1000000}, {`Fri Mar 01 17:32:01 1996 PST`, 1, 1000, 1000000}, {`Mon Dec 30 17:32:01 1996 PST`, 4, 1000, 1000000}, {`Tue Dec 31 17:32:01 1996 PST`, 4, 1000, 1000000}, {`Wed Jan 01 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Fri Feb 28 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Sat Mar 01 17:32:01 1997 PST`, 1, 1000, 1000000}, {`Tue Dec 30 17:32:01 1997 PST`, 4, 1000, 1000000}, {`Wed Dec 31 17:32:01 1997 PST`, 4, 1000, 1000000}, {`Fri Dec 31 17:32:01 1999 PST`, 4, 1000, 1000000}, {`Sat Jan 01 17:32:01 2000 PST`, 1, 1000, 1000000}, {`Sun Dec 31 17:32:01 2000 PST`, 4, 1000, 1000000}, {`Mon Jan 01 17:32:01 2001 PST`, 1, 1000, 1000000}},
			},
			{
				Statement: `SELECT d1 as timestamptz,
   date_part( 'isoyear', d1) AS isoyear, date_part( 'week', d1) AS week,
   date_part( 'isodow', d1) AS isodow, date_part( 'dow', d1) AS dow,
   date_part( 'doy', d1) AS doy
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{`-infinity`, `-Infinity`, ``, ``, ``, ``}, {`infinity`, `Infinity`, ``, ``, ``, ``}, {`Wed Dec 31 16:00:00 1969 PST`, 1970, 1, 3, 3, 365}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:02 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01.4 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01.5 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01.6 1997 PST`, 1997, 7, 1, 1, 41}, {`Thu Jan 02 00:00:00 1997 PST`, 1997, 1, 4, 4, 2}, {`Thu Jan 02 03:04:05 1997 PST`, 1997, 1, 4, 4, 2}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Tue Jun 10 17:32:01 1997 PDT`, 1997, 24, 2, 2, 161}, {`Sat Sep 22 18:19:20 2001 PDT`, 2001, 38, 6, 6, 265}, {`Wed Mar 15 08:14:01 2000 PST`, 2000, 11, 3, 3, 75}, {`Wed Mar 15 04:14:02 2000 PST`, 2000, 11, 3, 3, 75}, {`Wed Mar 15 02:14:03 2000 PST`, 2000, 11, 3, 3, 75}, {`Wed Mar 15 03:14:04 2000 PST`, 2000, 11, 3, 3, 75}, {`Wed Mar 15 01:14:05 2000 PST`, 2000, 11, 3, 3, 75}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:00 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 09:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 09:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 09:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Mon Feb 10 14:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Thu Jul 10 14:32:01 1997 PDT`, 1997, 28, 4, 4, 191}, {`Tue Jun 10 18:32:01 1997 PDT`, 1997, 24, 2, 2, 161}, {`Mon Feb 10 17:32:01 1997 PST`, 1997, 7, 1, 1, 41}, {`Tue Feb 11 17:32:01 1997 PST`, 1997, 7, 2, 2, 42}, {`Wed Feb 12 17:32:01 1997 PST`, 1997, 7, 3, 3, 43}, {`Thu Feb 13 17:32:01 1997 PST`, 1997, 7, 4, 4, 44}, {`Fri Feb 14 17:32:01 1997 PST`, 1997, 7, 5, 5, 45}, {`Sat Feb 15 17:32:01 1997 PST`, 1997, 7, 6, 6, 46}, {`Sun Feb 16 17:32:01 1997 PST`, 1997, 7, 7, 0, 47}, {`Tue Feb 16 17:32:01 0097 PST BC`, -97, 7, 2, 2, 47}, {`Sat Feb 16 17:32:01 0097 PST`, 97, 7, 6, 6, 47}, {`Thu Feb 16 17:32:01 0597 PST`, 597, 7, 4, 4, 47}, {`Tue Feb 16 17:32:01 1097 PST`, 1097, 7, 2, 2, 47}, {`Sat Feb 16 17:32:01 1697 PST`, 1697, 7, 6, 6, 47}, {`Thu Feb 16 17:32:01 1797 PST`, 1797, 7, 4, 4, 47}, {`Tue Feb 16 17:32:01 1897 PST`, 1897, 7, 2, 2, 47}, {`Sun Feb 16 17:32:01 1997 PST`, 1997, 7, 7, 0, 47}, {`Sat Feb 16 17:32:01 2097 PST`, 2097, 7, 6, 6, 47}, {`Wed Feb 28 17:32:01 1996 PST`, 1996, 9, 3, 3, 59}, {`Thu Feb 29 17:32:01 1996 PST`, 1996, 9, 4, 4, 60}, {`Fri Mar 01 17:32:01 1996 PST`, 1996, 9, 5, 5, 61}, {`Mon Dec 30 17:32:01 1996 PST`, 1997, 1, 1, 1, 365}, {`Tue Dec 31 17:32:01 1996 PST`, 1997, 1, 2, 2, 366}, {`Wed Jan 01 17:32:01 1997 PST`, 1997, 1, 3, 3, 1}, {`Fri Feb 28 17:32:01 1997 PST`, 1997, 9, 5, 5, 59}, {`Sat Mar 01 17:32:01 1997 PST`, 1997, 9, 6, 6, 60}, {`Tue Dec 30 17:32:01 1997 PST`, 1998, 1, 2, 2, 364}, {`Wed Dec 31 17:32:01 1997 PST`, 1998, 1, 3, 3, 365}, {`Fri Dec 31 17:32:01 1999 PST`, 1999, 52, 5, 5, 365}, {`Sat Jan 01 17:32:01 2000 PST`, 1999, 52, 6, 6, 1}, {`Sun Dec 31 17:32:01 2000 PST`, 2000, 52, 7, 0, 366}, {`Mon Jan 01 17:32:01 2001 PST`, 2001, 1, 1, 1, 1}},
			},
			{
				Statement: `SELECT d1 as timestamptz,
   date_part( 'decade', d1) AS decade,
   date_part( 'century', d1) AS century,
   date_part( 'millennium', d1) AS millennium,
   round(date_part( 'julian', d1)) AS julian,
   date_part( 'epoch', d1) AS epoch
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{`-infinity`, `-Infinity`, `-Infinity`, `-Infinity`, `-Infinity`, `-Infinity`}, {`infinity`, `Infinity`, `Infinity`, `Infinity`, `Infinity`, `Infinity`}, {`Wed Dec 31 16:00:00 1969 PST`, 196, 20, 2, 2440588, 0}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:02 1997 PST`, 199, 20, 2, 2450491, 855624722}, {`Mon Feb 10 17:32:01.4 1997 PST`, 199, 20, 2, 2450491, 855624721.4}, {`Mon Feb 10 17:32:01.5 1997 PST`, 199, 20, 2, 2450491, 855624721.5}, {`Mon Feb 10 17:32:01.6 1997 PST`, 199, 20, 2, 2450491, 855624721.6}, {`Thu Jan 02 00:00:00 1997 PST`, 199, 20, 2, 2450451, 852192000}, {`Thu Jan 02 03:04:05 1997 PST`, 199, 20, 2, 2450451, 852203045}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Tue Jun 10 17:32:01 1997 PDT`, 199, 20, 2, 2450611, 865989121}, {`Sat Sep 22 18:19:20 2001 PDT`, 200, 21, 3, 2452176, 1001207960}, {`Wed Mar 15 08:14:01 2000 PST`, 200, 20, 2, 2451619, 953136841}, {`Wed Mar 15 04:14:02 2000 PST`, 200, 20, 2, 2451619, 953122442}, {`Wed Mar 15 02:14:03 2000 PST`, 200, 20, 2, 2451619, 953115243}, {`Wed Mar 15 03:14:04 2000 PST`, 200, 20, 2, 2451619, 953118844}, {`Wed Mar 15 01:14:05 2000 PST`, 200, 20, 2, 2451619, 953111645}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:00 1997 PST`, 199, 20, 2, 2450491, 855624720}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Mon Feb 10 09:32:01 1997 PST`, 199, 20, 2, 2450490, 855595921}, {`Mon Feb 10 09:32:01 1997 PST`, 199, 20, 2, 2450490, 855595921}, {`Mon Feb 10 09:32:01 1997 PST`, 199, 20, 2, 2450490, 855595921}, {`Mon Feb 10 14:32:01 1997 PST`, 199, 20, 2, 2450491, 855613921}, {`Thu Jul 10 14:32:01 1997 PDT`, 199, 20, 2, 2450641, 868570321}, {`Tue Jun 10 18:32:01 1997 PDT`, 199, 20, 2, 2450611, 865992721}, {`Mon Feb 10 17:32:01 1997 PST`, 199, 20, 2, 2450491, 855624721}, {`Tue Feb 11 17:32:01 1997 PST`, 199, 20, 2, 2450492, 855711121}, {`Wed Feb 12 17:32:01 1997 PST`, 199, 20, 2, 2450493, 855797521}, {`Thu Feb 13 17:32:01 1997 PST`, 199, 20, 2, 2450494, 855883921}, {`Fri Feb 14 17:32:01 1997 PST`, 199, 20, 2, 2450495, 855970321}, {`Sat Feb 15 17:32:01 1997 PST`, 199, 20, 2, 2450496, 856056721}, {`Sun Feb 16 17:32:01 1997 PST`, 199, 20, 2, 2450497, 856143121}, {`Tue Feb 16 17:32:01 0097 PST BC`, -10, -1, -1, 1686043, -65192682479}, {`Sat Feb 16 17:32:01 0097 PST`, 9, 1, 1, 1756537, -59102000879}, {`Thu Feb 16 17:32:01 0597 PST`, 59, 6, 1, 1939158, -43323546479}, {`Tue Feb 16 17:32:01 1097 PST`, 109, 11, 2, 2121779, -27545092079}, {`Sat Feb 16 17:32:01 1697 PST`, 169, 17, 2, 2340925, -8610877679}, {`Thu Feb 16 17:32:01 1797 PST`, 179, 18, 2, 2377449, -5455204079}, {`Tue Feb 16 17:32:01 1897 PST`, 189, 19, 2, 2413973, -2299530479}, {`Sun Feb 16 17:32:01 1997 PST`, 199, 20, 2, 2450497, 856143121}, {`Sat Feb 16 17:32:01 2097 PST`, 209, 21, 3, 2487022, 4011903121}, {`Wed Feb 28 17:32:01 1996 PST`, 199, 20, 2, 2450143, 825557521}, {`Thu Feb 29 17:32:01 1996 PST`, 199, 20, 2, 2450144, 825643921}, {`Fri Mar 01 17:32:01 1996 PST`, 199, 20, 2, 2450145, 825730321}, {`Mon Dec 30 17:32:01 1996 PST`, 199, 20, 2, 2450449, 851995921}, {`Tue Dec 31 17:32:01 1996 PST`, 199, 20, 2, 2450450, 852082321}, {`Wed Jan 01 17:32:01 1997 PST`, 199, 20, 2, 2450451, 852168721}, {`Fri Feb 28 17:32:01 1997 PST`, 199, 20, 2, 2450509, 857179921}, {`Sat Mar 01 17:32:01 1997 PST`, 199, 20, 2, 2450510, 857266321}, {`Tue Dec 30 17:32:01 1997 PST`, 199, 20, 2, 2450814, 883531921}, {`Wed Dec 31 17:32:01 1997 PST`, 199, 20, 2, 2450815, 883618321}, {`Fri Dec 31 17:32:01 1999 PST`, 199, 20, 2, 2451545, 946690321}, {`Sat Jan 01 17:32:01 2000 PST`, 200, 20, 2, 2451546, 946776721}, {`Sun Dec 31 17:32:01 2000 PST`, 200, 20, 2, 2451911, 978312721}, {`Mon Jan 01 17:32:01 2001 PST`, 200, 21, 3, 2451912, 978399121}},
			},
			{
				Statement: `SELECT d1 as timestamptz,
   date_part( 'timezone', d1) AS timezone,
   date_part( 'timezone_hour', d1) AS timezone_hour,
   date_part( 'timezone_minute', d1) AS timezone_minute
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{`-infinity`, ``, ``, ``}, {`infinity`, ``, ``, ``}, {`Wed Dec 31 16:00:00 1969 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:02 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01.4 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01.5 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01.6 1997 PST`, -28800, -8, 0}, {`Thu Jan 02 00:00:00 1997 PST`, -28800, -8, 0}, {`Thu Jan 02 03:04:05 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Tue Jun 10 17:32:01 1997 PDT`, -25200, -7, 0}, {`Sat Sep 22 18:19:20 2001 PDT`, -25200, -7, 0}, {`Wed Mar 15 08:14:01 2000 PST`, -28800, -8, 0}, {`Wed Mar 15 04:14:02 2000 PST`, -28800, -8, 0}, {`Wed Mar 15 02:14:03 2000 PST`, -28800, -8, 0}, {`Wed Mar 15 03:14:04 2000 PST`, -28800, -8, 0}, {`Wed Mar 15 01:14:05 2000 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:00 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 09:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 09:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 09:32:01 1997 PST`, -28800, -8, 0}, {`Mon Feb 10 14:32:01 1997 PST`, -28800, -8, 0}, {`Thu Jul 10 14:32:01 1997 PDT`, -25200, -7, 0}, {`Tue Jun 10 18:32:01 1997 PDT`, -25200, -7, 0}, {`Mon Feb 10 17:32:01 1997 PST`, -28800, -8, 0}, {`Tue Feb 11 17:32:01 1997 PST`, -28800, -8, 0}, {`Wed Feb 12 17:32:01 1997 PST`, -28800, -8, 0}, {`Thu Feb 13 17:32:01 1997 PST`, -28800, -8, 0}, {`Fri Feb 14 17:32:01 1997 PST`, -28800, -8, 0}, {`Sat Feb 15 17:32:01 1997 PST`, -28800, -8, 0}, {`Sun Feb 16 17:32:01 1997 PST`, -28800, -8, 0}, {`Tue Feb 16 17:32:01 0097 PST BC`, -28800, -8, 0}, {`Sat Feb 16 17:32:01 0097 PST`, -28800, -8, 0}, {`Thu Feb 16 17:32:01 0597 PST`, -28800, -8, 0}, {`Tue Feb 16 17:32:01 1097 PST`, -28800, -8, 0}, {`Sat Feb 16 17:32:01 1697 PST`, -28800, -8, 0}, {`Thu Feb 16 17:32:01 1797 PST`, -28800, -8, 0}, {`Tue Feb 16 17:32:01 1897 PST`, -28800, -8, 0}, {`Sun Feb 16 17:32:01 1997 PST`, -28800, -8, 0}, {`Sat Feb 16 17:32:01 2097 PST`, -28800, -8, 0}, {`Wed Feb 28 17:32:01 1996 PST`, -28800, -8, 0}, {`Thu Feb 29 17:32:01 1996 PST`, -28800, -8, 0}, {`Fri Mar 01 17:32:01 1996 PST`, -28800, -8, 0}, {`Mon Dec 30 17:32:01 1996 PST`, -28800, -8, 0}, {`Tue Dec 31 17:32:01 1996 PST`, -28800, -8, 0}, {`Wed Jan 01 17:32:01 1997 PST`, -28800, -8, 0}, {`Fri Feb 28 17:32:01 1997 PST`, -28800, -8, 0}, {`Sat Mar 01 17:32:01 1997 PST`, -28800, -8, 0}, {`Tue Dec 30 17:32:01 1997 PST`, -28800, -8, 0}, {`Wed Dec 31 17:32:01 1997 PST`, -28800, -8, 0}, {`Fri Dec 31 17:32:01 1999 PST`, -28800, -8, 0}, {`Sat Jan 01 17:32:01 2000 PST`, -28800, -8, 0}, {`Sun Dec 31 17:32:01 2000 PST`, -28800, -8, 0}, {`Mon Jan 01 17:32:01 2001 PST`, -28800, -8, 0}},
			},
			{
				Statement: `SELECT d1 as "timestamp",
   extract(microseconds from d1) AS microseconds,
   extract(milliseconds from d1) AS milliseconds,
   extract(seconds from d1) AS seconds,
   round(extract(julian from d1)) AS julian,
   extract(epoch from d1) AS epoch
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{`-infinity`, ``, ``, ``, `-Infinity`, `-Infinity`}, {`infinity`, ``, ``, ``, `Infinity`, `Infinity`}, {`Wed Dec 31 16:00:00 1969 PST`, 0, 0.000, 0.000000, 2440588, 0.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:02 1997 PST`, 2000000, 2000.000, 2.000000, 2450491, 855624722.000000}, {`Mon Feb 10 17:32:01.4 1997 PST`, 1400000, 1400.000, 1.400000, 2450491, 855624721.400000}, {`Mon Feb 10 17:32:01.5 1997 PST`, 1500000, 1500.000, 1.500000, 2450491, 855624721.500000}, {`Mon Feb 10 17:32:01.6 1997 PST`, 1600000, 1600.000, 1.600000, 2450491, 855624721.600000}, {`Thu Jan 02 00:00:00 1997 PST`, 0, 0.000, 0.000000, 2450451, 852192000.000000}, {`Thu Jan 02 03:04:05 1997 PST`, 5000000, 5000.000, 5.000000, 2450451, 852203045.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Tue Jun 10 17:32:01 1997 PDT`, 1000000, 1000.000, 1.000000, 2450611, 865989121.000000}, {`Sat Sep 22 18:19:20 2001 PDT`, 20000000, 20000.000, 20.000000, 2452176, 1001207960.000000}, {`Wed Mar 15 08:14:01 2000 PST`, 1000000, 1000.000, 1.000000, 2451619, 953136841.000000}, {`Wed Mar 15 04:14:02 2000 PST`, 2000000, 2000.000, 2.000000, 2451619, 953122442.000000}, {`Wed Mar 15 02:14:03 2000 PST`, 3000000, 3000.000, 3.000000, 2451619, 953115243.000000}, {`Wed Mar 15 03:14:04 2000 PST`, 4000000, 4000.000, 4.000000, 2451619, 953118844.000000}, {`Wed Mar 15 01:14:05 2000 PST`, 5000000, 5000.000, 5.000000, 2451619, 953111645.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:00 1997 PST`, 0, 0.000, 0.000000, 2450491, 855624720.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Mon Feb 10 09:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450490, 855595921.000000}, {`Mon Feb 10 09:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450490, 855595921.000000}, {`Mon Feb 10 09:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450490, 855595921.000000}, {`Mon Feb 10 14:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855613921.000000}, {`Thu Jul 10 14:32:01 1997 PDT`, 1000000, 1000.000, 1.000000, 2450641, 868570321.000000}, {`Tue Jun 10 18:32:01 1997 PDT`, 1000000, 1000.000, 1.000000, 2450611, 865992721.000000}, {`Mon Feb 10 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450491, 855624721.000000}, {`Tue Feb 11 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450492, 855711121.000000}, {`Wed Feb 12 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450493, 855797521.000000}, {`Thu Feb 13 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450494, 855883921.000000}, {`Fri Feb 14 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450495, 855970321.000000}, {`Sat Feb 15 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450496, 856056721.000000}, {`Sun Feb 16 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450497, 856143121.000000}, {`Tue Feb 16 17:32:01 0097 PST BC`, 1000000, 1000.000, 1.000000, 1686043, -65192682479.000000}, {`Sat Feb 16 17:32:01 0097 PST`, 1000000, 1000.000, 1.000000, 1756537, -59102000879.000000}, {`Thu Feb 16 17:32:01 0597 PST`, 1000000, 1000.000, 1.000000, 1939158, -43323546479.000000}, {`Tue Feb 16 17:32:01 1097 PST`, 1000000, 1000.000, 1.000000, 2121779, -27545092079.000000}, {`Sat Feb 16 17:32:01 1697 PST`, 1000000, 1000.000, 1.000000, 2340925, -8610877679.000000}, {`Thu Feb 16 17:32:01 1797 PST`, 1000000, 1000.000, 1.000000, 2377449, -5455204079.000000}, {`Tue Feb 16 17:32:01 1897 PST`, 1000000, 1000.000, 1.000000, 2413973, -2299530479.000000}, {`Sun Feb 16 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450497, 856143121.000000}, {`Sat Feb 16 17:32:01 2097 PST`, 1000000, 1000.000, 1.000000, 2487022, 4011903121.000000}, {`Wed Feb 28 17:32:01 1996 PST`, 1000000, 1000.000, 1.000000, 2450143, 825557521.000000}, {`Thu Feb 29 17:32:01 1996 PST`, 1000000, 1000.000, 1.000000, 2450144, 825643921.000000}, {`Fri Mar 01 17:32:01 1996 PST`, 1000000, 1000.000, 1.000000, 2450145, 825730321.000000}, {`Mon Dec 30 17:32:01 1996 PST`, 1000000, 1000.000, 1.000000, 2450449, 851995921.000000}, {`Tue Dec 31 17:32:01 1996 PST`, 1000000, 1000.000, 1.000000, 2450450, 852082321.000000}, {`Wed Jan 01 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450451, 852168721.000000}, {`Fri Feb 28 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450509, 857179921.000000}, {`Sat Mar 01 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450510, 857266321.000000}, {`Tue Dec 30 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450814, 883531921.000000}, {`Wed Dec 31 17:32:01 1997 PST`, 1000000, 1000.000, 1.000000, 2450815, 883618321.000000}, {`Fri Dec 31 17:32:01 1999 PST`, 1000000, 1000.000, 1.000000, 2451545, 946690321.000000}, {`Sat Jan 01 17:32:01 2000 PST`, 1000000, 1000.000, 1.000000, 2451546, 946776721.000000}, {`Sun Dec 31 17:32:01 2000 PST`, 1000000, 1000.000, 1.000000, 2451911, 978312721.000000}, {`Mon Jan 01 17:32:01 2001 PST`, 1000000, 1000.000, 1.000000, 2451912, 978399121.000000}},
			},
			{
				Statement: `SELECT date_part('epoch', '294270-01-01 00:00:00+00'::timestamptz);`,
				Results:   []sql.Row{{9224097091200}},
			},
			{
				Statement: `SELECT extract(epoch from '294270-01-01 00:00:00+00'::timestamptz);`,
				Results:   []sql.Row{{9224097091200.000000}},
			},
			{
				Statement: `SELECT extract(epoch from '5000-01-01 00:00:00+00'::timestamptz);`,
				Results:   []sql.Row{{95617584000.000000}},
			},
			{
				Statement: `SELECT to_char(d1, 'DAY Day day DY Dy dy MONTH Month month RM MON Mon mon')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`WEDNESDAY Wednesday wednesday WED Wed wed DECEMBER  December  december  XII  DEC Dec dec`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`THURSDAY  Thursday  thursday  THU Thu thu JANUARY   January   january   I    JAN Jan jan`}, {`THURSDAY  Thursday  thursday  THU Thu thu JANUARY   January   january   I    JAN Jan jan`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`TUESDAY   Tuesday   tuesday   TUE Tue tue JUNE      June      june      VI   JUN Jun jun`}, {`SATURDAY  Saturday  saturday  SAT Sat sat SEPTEMBER September september IX   SEP Sep sep`}, {`WEDNESDAY Wednesday wednesday WED Wed wed MARCH     March     march     III  MAR Mar mar`}, {`WEDNESDAY Wednesday wednesday WED Wed wed MARCH     March     march     III  MAR Mar mar`}, {`WEDNESDAY Wednesday wednesday WED Wed wed MARCH     March     march     III  MAR Mar mar`}, {`WEDNESDAY Wednesday wednesday WED Wed wed MARCH     March     march     III  MAR Mar mar`}, {`WEDNESDAY Wednesday wednesday WED Wed wed MARCH     March     march     III  MAR Mar mar`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`THURSDAY  Thursday  thursday  THU Thu thu JULY      July      july      VII  JUL Jul jul`}, {`TUESDAY   Tuesday   tuesday   TUE Tue tue JUNE      June      june      VI   JUN Jun jun`}, {`MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb`}, {`TUESDAY   Tuesday   tuesday   TUE Tue tue FEBRUARY  February  february  II   FEB Feb feb`}, {`WEDNESDAY Wednesday wednesday WED Wed wed FEBRUARY  February  february  II   FEB Feb feb`}, {`THURSDAY  Thursday  thursday  THU Thu thu FEBRUARY  February  february  II   FEB Feb feb`}, {`FRIDAY    Friday    friday    FRI Fri fri FEBRUARY  February  february  II   FEB Feb feb`}, {`SATURDAY  Saturday  saturday  SAT Sat sat FEBRUARY  February  february  II   FEB Feb feb`}, {`SUNDAY    Sunday    sunday    SUN Sun sun FEBRUARY  February  february  II   FEB Feb feb`}, {`TUESDAY   Tuesday   tuesday   TUE Tue tue FEBRUARY  February  february  II   FEB Feb feb`}, {`SATURDAY  Saturday  saturday  SAT Sat sat FEBRUARY  February  february  II   FEB Feb feb`}, {`THURSDAY  Thursday  thursday  THU Thu thu FEBRUARY  February  february  II   FEB Feb feb`}, {`TUESDAY   Tuesday   tuesday   TUE Tue tue FEBRUARY  February  february  II   FEB Feb feb`}, {`SATURDAY  Saturday  saturday  SAT Sat sat FEBRUARY  February  february  II   FEB Feb feb`}, {`THURSDAY  Thursday  thursday  THU Thu thu FEBRUARY  February  february  II   FEB Feb feb`}, {`TUESDAY   Tuesday   tuesday   TUE Tue tue FEBRUARY  February  february  II   FEB Feb feb`}, {`SUNDAY    Sunday    sunday    SUN Sun sun FEBRUARY  February  february  II   FEB Feb feb`}, {`SATURDAY  Saturday  saturday  SAT Sat sat FEBRUARY  February  february  II   FEB Feb feb`}, {`WEDNESDAY Wednesday wednesday WED Wed wed FEBRUARY  February  february  II   FEB Feb feb`}, {`THURSDAY  Thursday  thursday  THU Thu thu FEBRUARY  February  february  II   FEB Feb feb`}, {`FRIDAY    Friday    friday    FRI Fri fri MARCH     March     march     III  MAR Mar mar`}, {`MONDAY    Monday    monday    MON Mon mon DECEMBER  December  december  XII  DEC Dec dec`}, {`TUESDAY   Tuesday   tuesday   TUE Tue tue DECEMBER  December  december  XII  DEC Dec dec`}, {`WEDNESDAY Wednesday wednesday WED Wed wed JANUARY   January   january   I    JAN Jan jan`}, {`FRIDAY    Friday    friday    FRI Fri fri FEBRUARY  February  february  II   FEB Feb feb`}, {`SATURDAY  Saturday  saturday  SAT Sat sat MARCH     March     march     III  MAR Mar mar`}, {`TUESDAY   Tuesday   tuesday   TUE Tue tue DECEMBER  December  december  XII  DEC Dec dec`}, {`WEDNESDAY Wednesday wednesday WED Wed wed DECEMBER  December  december  XII  DEC Dec dec`}, {`FRIDAY    Friday    friday    FRI Fri fri DECEMBER  December  december  XII  DEC Dec dec`}, {`SATURDAY  Saturday  saturday  SAT Sat sat JANUARY   January   january   I    JAN Jan jan`}, {`SUNDAY    Sunday    sunday    SUN Sun sun DECEMBER  December  december  XII  DEC Dec dec`}, {`MONDAY    Monday    monday    MON Mon mon JANUARY   January   january   I    JAN Jan jan`}},
			},
			{
				Statement: `SELECT to_char(d1, 'FMDAY FMDay FMday FMMONTH FMMonth FMmonth FMRM')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`WEDNESDAY Wednesday wednesday DECEMBER December december XII`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`THURSDAY Thursday thursday JANUARY January january I`}, {`THURSDAY Thursday thursday JANUARY January january I`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`TUESDAY Tuesday tuesday JUNE June june VI`}, {`SATURDAY Saturday saturday SEPTEMBER September september IX`}, {`WEDNESDAY Wednesday wednesday MARCH March march III`}, {`WEDNESDAY Wednesday wednesday MARCH March march III`}, {`WEDNESDAY Wednesday wednesday MARCH March march III`}, {`WEDNESDAY Wednesday wednesday MARCH March march III`}, {`WEDNESDAY Wednesday wednesday MARCH March march III`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`THURSDAY Thursday thursday JULY July july VII`}, {`TUESDAY Tuesday tuesday JUNE June june VI`}, {`MONDAY Monday monday FEBRUARY February february II`}, {`TUESDAY Tuesday tuesday FEBRUARY February february II`}, {`WEDNESDAY Wednesday wednesday FEBRUARY February february II`}, {`THURSDAY Thursday thursday FEBRUARY February february II`}, {`FRIDAY Friday friday FEBRUARY February february II`}, {`SATURDAY Saturday saturday FEBRUARY February february II`}, {`SUNDAY Sunday sunday FEBRUARY February february II`}, {`TUESDAY Tuesday tuesday FEBRUARY February february II`}, {`SATURDAY Saturday saturday FEBRUARY February february II`}, {`THURSDAY Thursday thursday FEBRUARY February february II`}, {`TUESDAY Tuesday tuesday FEBRUARY February february II`}, {`SATURDAY Saturday saturday FEBRUARY February february II`}, {`THURSDAY Thursday thursday FEBRUARY February february II`}, {`TUESDAY Tuesday tuesday FEBRUARY February february II`}, {`SUNDAY Sunday sunday FEBRUARY February february II`}, {`SATURDAY Saturday saturday FEBRUARY February february II`}, {`WEDNESDAY Wednesday wednesday FEBRUARY February february II`}, {`THURSDAY Thursday thursday FEBRUARY February february II`}, {`FRIDAY Friday friday MARCH March march III`}, {`MONDAY Monday monday DECEMBER December december XII`}, {`TUESDAY Tuesday tuesday DECEMBER December december XII`}, {`WEDNESDAY Wednesday wednesday JANUARY January january I`}, {`FRIDAY Friday friday FEBRUARY February february II`}, {`SATURDAY Saturday saturday MARCH March march III`}, {`TUESDAY Tuesday tuesday DECEMBER December december XII`}, {`WEDNESDAY Wednesday wednesday DECEMBER December december XII`}, {`FRIDAY Friday friday DECEMBER December december XII`}, {`SATURDAY Saturday saturday JANUARY January january I`}, {`SUNDAY Sunday sunday DECEMBER December december XII`}, {`MONDAY Monday monday JANUARY January january I`}},
			},
			{
				Statement: `SELECT to_char(d1, 'Y,YYY YYYY YYY YY Y CC Q MM WW DDD DD D J')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`1,969 1969 969 69 9 20 4 12 53 365 31 4 2440587`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 01 01 002 02 5 2450451`}, {`1,997 1997 997 97 7 20 1 01 01 002 02 5 2450451`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 2 06 23 161 10 3 2450610`}, {`2,001 2001 001 01 1 21 3 09 38 265 22 7 2452175`}, {`2,000 2000 000 00 0 20 1 03 11 075 15 4 2451619`}, {`2,000 2000 000 00 0 20 1 03 11 075 15 4 2451619`}, {`2,000 2000 000 00 0 20 1 03 11 075 15 4 2451619`}, {`2,000 2000 000 00 0 20 1 03 11 075 15 4 2451619`}, {`2,000 2000 000 00 0 20 1 03 11 075 15 4 2451619`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 3 07 28 191 10 5 2450640`}, {`1,997 1997 997 97 7 20 2 06 23 161 10 3 2450610`}, {`1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 02 06 042 11 3 2450491`}, {`1,997 1997 997 97 7 20 1 02 07 043 12 4 2450492`}, {`1,997 1997 997 97 7 20 1 02 07 044 13 5 2450493`}, {`1,997 1997 997 97 7 20 1 02 07 045 14 6 2450494`}, {`1,997 1997 997 97 7 20 1 02 07 046 15 7 2450495`}, {`1,997 1997 997 97 7 20 1 02 07 047 16 1 2450496`}, {`0,097 0097 097 97 7 -01 1 02 07 047 16 3 1686042`}, {`0,097 0097 097 97 7 01 1 02 07 047 16 7 1756536`}, {`0,597 0597 597 97 7 06 1 02 07 047 16 5 1939157`}, {`1,097 1097 097 97 7 11 1 02 07 047 16 3 2121778`}, {`1,697 1697 697 97 7 17 1 02 07 047 16 7 2340924`}, {`1,797 1797 797 97 7 18 1 02 07 047 16 5 2377448`}, {`1,897 1897 897 97 7 19 1 02 07 047 16 3 2413972`}, {`1,997 1997 997 97 7 20 1 02 07 047 16 1 2450496`}, {`2,097 2097 097 97 7 21 1 02 07 047 16 7 2487021`}, {`1,996 1996 996 96 6 20 1 02 09 059 28 4 2450142`}, {`1,996 1996 996 96 6 20 1 02 09 060 29 5 2450143`}, {`1,996 1996 996 96 6 20 1 03 09 061 01 6 2450144`}, {`1,996 1996 996 96 6 20 4 12 53 365 30 2 2450448`}, {`1,996 1996 996 96 6 20 4 12 53 366 31 3 2450449`}, {`1,997 1997 997 97 7 20 1 01 01 001 01 4 2450450`}, {`1,997 1997 997 97 7 20 1 02 09 059 28 6 2450508`}, {`1,997 1997 997 97 7 20 1 03 09 060 01 7 2450509`}, {`1,997 1997 997 97 7 20 4 12 52 364 30 3 2450813`}, {`1,997 1997 997 97 7 20 4 12 53 365 31 4 2450814`}, {`1,999 1999 999 99 9 20 4 12 53 365 31 6 2451544`}, {`2,000 2000 000 00 0 20 1 01 01 001 01 7 2451545`}, {`2,000 2000 000 00 0 20 4 12 53 366 31 1 2451910`}, {`2,001 2001 001 01 1 21 1 01 01 001 01 2 2451911`}},
			},
			{
				Statement: `SELECT to_char(d1, 'FMY,YYY FMYYYY FMYYY FMYY FMY FMCC FMQ FMMM FMWW FMDDD FMDD FMD FMJ')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`1,969 1969 969 69 9 20 4 12 53 365 31 4 2440587`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 1 1 2 2 5 2450451`}, {`1,997 1997 997 97 7 20 1 1 1 2 2 5 2450451`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 2 6 23 161 10 3 2450610`}, {`2,001 2001 1 1 1 21 3 9 38 265 22 7 2452175`}, {`2,000 2000 0 0 0 20 1 3 11 75 15 4 2451619`}, {`2,000 2000 0 0 0 20 1 3 11 75 15 4 2451619`}, {`2,000 2000 0 0 0 20 1 3 11 75 15 4 2451619`}, {`2,000 2000 0 0 0 20 1 3 11 75 15 4 2451619`}, {`2,000 2000 0 0 0 20 1 3 11 75 15 4 2451619`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 3 7 28 191 10 5 2450640`}, {`1,997 1997 997 97 7 20 2 6 23 161 10 3 2450610`}, {`1,997 1997 997 97 7 20 1 2 6 41 10 2 2450490`}, {`1,997 1997 997 97 7 20 1 2 6 42 11 3 2450491`}, {`1,997 1997 997 97 7 20 1 2 7 43 12 4 2450492`}, {`1,997 1997 997 97 7 20 1 2 7 44 13 5 2450493`}, {`1,997 1997 997 97 7 20 1 2 7 45 14 6 2450494`}, {`1,997 1997 997 97 7 20 1 2 7 46 15 7 2450495`}, {`1,997 1997 997 97 7 20 1 2 7 47 16 1 2450496`}, {`0,097 97 97 97 7 -1 1 2 7 47 16 3 1686042`}, {`0,097 97 97 97 7 1 1 2 7 47 16 7 1756536`}, {`0,597 597 597 97 7 6 1 2 7 47 16 5 1939157`}, {`1,097 1097 97 97 7 11 1 2 7 47 16 3 2121778`}, {`1,697 1697 697 97 7 17 1 2 7 47 16 7 2340924`}, {`1,797 1797 797 97 7 18 1 2 7 47 16 5 2377448`}, {`1,897 1897 897 97 7 19 1 2 7 47 16 3 2413972`}, {`1,997 1997 997 97 7 20 1 2 7 47 16 1 2450496`}, {`2,097 2097 97 97 7 21 1 2 7 47 16 7 2487021`}, {`1,996 1996 996 96 6 20 1 2 9 59 28 4 2450142`}, {`1,996 1996 996 96 6 20 1 2 9 60 29 5 2450143`}, {`1,996 1996 996 96 6 20 1 3 9 61 1 6 2450144`}, {`1,996 1996 996 96 6 20 4 12 53 365 30 2 2450448`}, {`1,996 1996 996 96 6 20 4 12 53 366 31 3 2450449`}, {`1,997 1997 997 97 7 20 1 1 1 1 1 4 2450450`}, {`1,997 1997 997 97 7 20 1 2 9 59 28 6 2450508`}, {`1,997 1997 997 97 7 20 1 3 9 60 1 7 2450509`}, {`1,997 1997 997 97 7 20 4 12 52 364 30 3 2450813`}, {`1,997 1997 997 97 7 20 4 12 53 365 31 4 2450814`}, {`1,999 1999 999 99 9 20 4 12 53 365 31 6 2451544`}, {`2,000 2000 0 0 0 20 1 1 1 1 1 7 2451545`}, {`2,000 2000 0 0 0 20 4 12 53 366 31 1 2451910`}, {`2,001 2001 1 1 1 21 1 1 1 1 1 2 2451911`}},
			},
			{
				Statement: `SELECT to_char(d1, 'HH HH12 HH24 MI SS SSSS')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`04 04 16 00 00 57600`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 02 63122`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`12 12 00 00 00 0`}, {`03 03 03 04 05 11045`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`06 06 18 19 20 65960`}, {`08 08 08 14 01 29641`}, {`04 04 04 14 02 15242`}, {`02 02 02 14 03 8043`}, {`03 03 03 14 04 11644`}, {`01 01 01 14 05 4445`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 00 63120`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`09 09 09 32 01 34321`}, {`09 09 09 32 01 34321`}, {`09 09 09 32 01 34321`}, {`02 02 14 32 01 52321`}, {`02 02 14 32 01 52321`}, {`06 06 18 32 01 66721`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}, {`05 05 17 32 01 63121`}},
			},
			{
				Statement: `SELECT to_char(d1, E'"HH:MI:SS is" HH:MI:SS "\\"text between quote marks\\""')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`HH:MI:SS is 04:00:00 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:02 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 12:00:00 "text between quote marks"`}, {`HH:MI:SS is 03:04:05 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 06:19:20 "text between quote marks"`}, {`HH:MI:SS is 08:14:01 "text between quote marks"`}, {`HH:MI:SS is 04:14:02 "text between quote marks"`}, {`HH:MI:SS is 02:14:03 "text between quote marks"`}, {`HH:MI:SS is 03:14:04 "text between quote marks"`}, {`HH:MI:SS is 01:14:05 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:00 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 09:32:01 "text between quote marks"`}, {`HH:MI:SS is 09:32:01 "text between quote marks"`}, {`HH:MI:SS is 09:32:01 "text between quote marks"`}, {`HH:MI:SS is 02:32:01 "text between quote marks"`}, {`HH:MI:SS is 02:32:01 "text between quote marks"`}, {`HH:MI:SS is 06:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}, {`HH:MI:SS is 05:32:01 "text between quote marks"`}},
			},
			{
				Statement: `SELECT to_char(d1, 'HH24--text--MI--text--SS')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`16--text--00--text--00`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--02`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`00--text--00--text--00`}, {`03--text--04--text--05`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`18--text--19--text--20`}, {`08--text--14--text--01`}, {`04--text--14--text--02`}, {`02--text--14--text--03`}, {`03--text--14--text--04`}, {`01--text--14--text--05`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--00`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`09--text--32--text--01`}, {`09--text--32--text--01`}, {`09--text--32--text--01`}, {`14--text--32--text--01`}, {`14--text--32--text--01`}, {`18--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}, {`17--text--32--text--01`}},
			},
			{
				Statement: `SELECT to_char(d1, 'YYYYTH YYYYth Jth')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`1969TH 1969th 2440587th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450451st`}, {`1997TH 1997th 2450451st`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450610th`}, {`2001ST 2001st 2452175th`}, {`2000TH 2000th 2451619th`}, {`2000TH 2000th 2451619th`}, {`2000TH 2000th 2451619th`}, {`2000TH 2000th 2451619th`}, {`2000TH 2000th 2451619th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450640th`}, {`1997TH 1997th 2450610th`}, {`1997TH 1997th 2450490th`}, {`1997TH 1997th 2450491st`}, {`1997TH 1997th 2450492nd`}, {`1997TH 1997th 2450493rd`}, {`1997TH 1997th 2450494th`}, {`1997TH 1997th 2450495th`}, {`1997TH 1997th 2450496th`}, {`0097TH 0097th 1686042nd`}, {`0097TH 0097th 1756536th`}, {`0597TH 0597th 1939157th`}, {`1097TH 1097th 2121778th`}, {`1697TH 1697th 2340924th`}, {`1797TH 1797th 2377448th`}, {`1897TH 1897th 2413972nd`}, {`1997TH 1997th 2450496th`}, {`2097TH 2097th 2487021st`}, {`1996TH 1996th 2450142nd`}, {`1996TH 1996th 2450143rd`}, {`1996TH 1996th 2450144th`}, {`1996TH 1996th 2450448th`}, {`1996TH 1996th 2450449th`}, {`1997TH 1997th 2450450th`}, {`1997TH 1997th 2450508th`}, {`1997TH 1997th 2450509th`}, {`1997TH 1997th 2450813th`}, {`1997TH 1997th 2450814th`}, {`1999TH 1999th 2451544th`}, {`2000TH 2000th 2451545th`}, {`2000TH 2000th 2451910th`}, {`2001ST 2001st 2451911th`}},
			},
			{
				Statement: `SELECT to_char(d1, 'YYYY A.D. YYYY a.d. YYYY bc HH:MI:SS P.M. HH:MI:SS p.m. HH:MI:SS pm')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`1969 A.D. 1969 a.d. 1969 ad 04:00:00 P.M. 04:00:00 p.m. 04:00:00 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:02 P.M. 05:32:02 p.m. 05:32:02 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 12:00:00 A.M. 12:00:00 a.m. 12:00:00 am`}, {`1997 A.D. 1997 a.d. 1997 ad 03:04:05 A.M. 03:04:05 a.m. 03:04:05 am`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`2001 A.D. 2001 a.d. 2001 ad 06:19:20 P.M. 06:19:20 p.m. 06:19:20 pm`}, {`2000 A.D. 2000 a.d. 2000 ad 08:14:01 A.M. 08:14:01 a.m. 08:14:01 am`}, {`2000 A.D. 2000 a.d. 2000 ad 04:14:02 A.M. 04:14:02 a.m. 04:14:02 am`}, {`2000 A.D. 2000 a.d. 2000 ad 02:14:03 A.M. 02:14:03 a.m. 02:14:03 am`}, {`2000 A.D. 2000 a.d. 2000 ad 03:14:04 A.M. 03:14:04 a.m. 03:14:04 am`}, {`2000 A.D. 2000 a.d. 2000 ad 01:14:05 A.M. 01:14:05 a.m. 01:14:05 am`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:00 P.M. 05:32:00 p.m. 05:32:00 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 09:32:01 A.M. 09:32:01 a.m. 09:32:01 am`}, {`1997 A.D. 1997 a.d. 1997 ad 09:32:01 A.M. 09:32:01 a.m. 09:32:01 am`}, {`1997 A.D. 1997 a.d. 1997 ad 09:32:01 A.M. 09:32:01 a.m. 09:32:01 am`}, {`1997 A.D. 1997 a.d. 1997 ad 02:32:01 P.M. 02:32:01 p.m. 02:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 02:32:01 P.M. 02:32:01 p.m. 02:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 06:32:01 P.M. 06:32:01 p.m. 06:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`0097 B.C. 0097 b.c. 0097 bc 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`0097 A.D. 0097 a.d. 0097 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`0597 A.D. 0597 a.d. 0597 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1097 A.D. 1097 a.d. 1097 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1697 A.D. 1697 a.d. 1697 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1797 A.D. 1797 a.d. 1797 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1897 A.D. 1897 a.d. 1897 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`2097 A.D. 2097 a.d. 2097 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1996 A.D. 1996 a.d. 1996 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1996 A.D. 1996 a.d. 1996 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1996 A.D. 1996 a.d. 1996 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1996 A.D. 1996 a.d. 1996 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1996 A.D. 1996 a.d. 1996 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1997 A.D. 1997 a.d. 1997 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`1999 A.D. 1999 a.d. 1999 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`2000 A.D. 2000 a.d. 2000 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`2000 A.D. 2000 a.d. 2000 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}, {`2001 A.D. 2001 a.d. 2001 ad 05:32:01 P.M. 05:32:01 p.m. 05:32:01 pm`}},
			},
			{
				Statement: `SELECT to_char(d1, 'IYYY IYY IY I IW IDDD ID')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`1970 970 70 0 01 003 3`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 01 004 4`}, {`1997 997 97 7 01 004 4`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 24 163 2`}, {`2001 001 01 1 38 265 6`}, {`2000 000 00 0 11 073 3`}, {`2000 000 00 0 11 073 3`}, {`2000 000 00 0 11 073 3`}, {`2000 000 00 0 11 073 3`}, {`2000 000 00 0 11 073 3`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 28 193 4`}, {`1997 997 97 7 24 163 2`}, {`1997 997 97 7 07 043 1`}, {`1997 997 97 7 07 044 2`}, {`1997 997 97 7 07 045 3`}, {`1997 997 97 7 07 046 4`}, {`1997 997 97 7 07 047 5`}, {`1997 997 97 7 07 048 6`}, {`1997 997 97 7 07 049 7`}, {`0097 097 97 7 07 044 2`}, {`0097 097 97 7 07 048 6`}, {`0597 597 97 7 07 046 4`}, {`1097 097 97 7 07 044 2`}, {`1697 697 97 7 07 048 6`}, {`1797 797 97 7 07 046 4`}, {`1897 897 97 7 07 044 2`}, {`1997 997 97 7 07 049 7`}, {`2097 097 97 7 07 048 6`}, {`1996 996 96 6 09 059 3`}, {`1996 996 96 6 09 060 4`}, {`1996 996 96 6 09 061 5`}, {`1997 997 97 7 01 001 1`}, {`1997 997 97 7 01 002 2`}, {`1997 997 97 7 01 003 3`}, {`1997 997 97 7 09 061 5`}, {`1997 997 97 7 09 062 6`}, {`1998 998 98 8 01 002 2`}, {`1998 998 98 8 01 003 3`}, {`1999 999 99 9 52 362 5`}, {`1999 999 99 9 52 363 6`}, {`2000 000 00 0 52 364 7`}, {`2001 001 01 1 01 001 1`}},
			},
			{
				Statement: `SELECT to_char(d1, 'FMIYYY FMIYY FMIY FMI FMIW FMIDDD FMID')
   FROM TIMESTAMPTZ_TBL;`,
				Results: []sql.Row{{``}, {``}, {`1970 970 70 0 1 3 3`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 1 4 4`}, {`1997 997 97 7 1 4 4`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 24 163 2`}, {`2001 1 1 1 38 265 6`}, {`2000 0 0 0 11 73 3`}, {`2000 0 0 0 11 73 3`}, {`2000 0 0 0 11 73 3`}, {`2000 0 0 0 11 73 3`}, {`2000 0 0 0 11 73 3`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 28 193 4`}, {`1997 997 97 7 24 163 2`}, {`1997 997 97 7 7 43 1`}, {`1997 997 97 7 7 44 2`}, {`1997 997 97 7 7 45 3`}, {`1997 997 97 7 7 46 4`}, {`1997 997 97 7 7 47 5`}, {`1997 997 97 7 7 48 6`}, {`1997 997 97 7 7 49 7`}, {`97 97 97 7 7 44 2`}, {`97 97 97 7 7 48 6`}, {`597 597 97 7 7 46 4`}, {`1097 97 97 7 7 44 2`}, {`1697 697 97 7 7 48 6`}, {`1797 797 97 7 7 46 4`}, {`1897 897 97 7 7 44 2`}, {`1997 997 97 7 7 49 7`}, {`2097 97 97 7 7 48 6`}, {`1996 996 96 6 9 59 3`}, {`1996 996 96 6 9 60 4`}, {`1996 996 96 6 9 61 5`}, {`1997 997 97 7 1 1 1`}, {`1997 997 97 7 1 2 2`}, {`1997 997 97 7 1 3 3`}, {`1997 997 97 7 9 61 5`}, {`1997 997 97 7 9 62 6`}, {`1998 998 98 8 1 2 2`}, {`1998 998 98 8 1 3 3`}, {`1999 999 99 9 52 362 5`}, {`1999 999 99 9 52 363 6`}, {`2000 0 0 0 52 364 7`}, {`2001 1 1 1 1 1 1`}},
			},
			{
				Statement: `SELECT to_char(d, 'FF1 FF2 FF3 FF4 FF5 FF6  ff1 ff2 ff3 ff4 ff5 ff6  MS US')
   FROM (VALUES
       ('2018-11-02 12:34:56'::timestamptz),
       ('2018-11-02 12:34:56.78'),
       ('2018-11-02 12:34:56.78901'),
       ('2018-11-02 12:34:56.78901234')
   ) d(d);`,
				Results: []sql.Row{{`0 00 000 0000 00000 000000  0 00 000 0000 00000 000000  000 000000`}, {`7 78 780 7800 78000 780000  7 78 780 7800 78000 780000  780 780000`}, {`7 78 789 7890 78901 789010  7 78 789 7890 78901 789010  789 789010`}, {`7 78 789 7890 78901 789012  7 78 789 7890 78901 789012  789 789012`}},
			},
			{
				Statement: `SET timezone = '00:00';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{+00, `+00:00`}},
			},
			{
				Statement: `SET timezone = '+02:00';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{-02, `-02:00`}},
			},
			{
				Statement: `SET timezone = '-13:00';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{+13, `+13:00`}},
			},
			{
				Statement: `SET timezone = '-00:30';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{`+00:30`, `+00:30`}},
			},
			{
				Statement: `SET timezone = '00:30';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{`-00:30`, `-00:30`}},
			},
			{
				Statement: `SET timezone = '-04:30';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{`+04:30`, `+04:30`}},
			},
			{
				Statement: `SET timezone = '04:30';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{`-04:30`, `-04:30`}},
			},
			{
				Statement: `SET timezone = '-04:15';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{`+04:15`, `+04:15`}},
			},
			{
				Statement: `SET timezone = '04:15';`,
			},
			{
				Statement: `SELECT to_char(now(), 'OF') as "OF", to_char(now(), 'TZH:TZM') as "TZH:TZM";`,
				Results:   []sql.Row{{`-04:15`, `-04:15`}},
			},
			{
				Statement: `RESET timezone;`,
			},
			{
				Statement: `SET timezone = '00:00';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "Of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{+00, `+00:00`}},
			},
			{
				Statement: `SET timezone = '+02:00';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{-02, `-02:00`}},
			},
			{
				Statement: `SET timezone = '-13:00';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{+13, `+13:00`}},
			},
			{
				Statement: `SET timezone = '-00:30';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{`+00:30`, `+00:30`}},
			},
			{
				Statement: `SET timezone = '00:30';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{`-00:30`, `-00:30`}},
			},
			{
				Statement: `SET timezone = '-04:30';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{`+04:30`, `+04:30`}},
			},
			{
				Statement: `SET timezone = '04:30';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{`-04:30`, `-04:30`}},
			},
			{
				Statement: `SET timezone = '-04:15';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{`+04:15`, `+04:15`}},
			},
			{
				Statement: `SET timezone = '04:15';`,
			},
			{
				Statement: `SELECT to_char(now(), 'of') as "of", to_char(now(), 'tzh:tzm') as "tzh:tzm";`,
				Results:   []sql.Row{{`-04:15`, `-04:15`}},
			},
			{
				Statement: `RESET timezone;`,
			},
			{
				Statement: `CREATE TABLE TIMESTAMPTZ_TST (a int , b timestamptz);`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TST VALUES(1, 'Sat Mar 12 23:58:48 1000 IST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TST VALUES(2, 'Sat Mar 12 23:58:48 10000 IST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TST VALUES(3, 'Sat Mar 12 23:58:48 100000 IST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TST VALUES(3, '10000 Mar 12 23:58:48 IST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TST VALUES(4, '100000312 23:58:48 IST');`,
			},
			{
				Statement: `INSERT INTO TIMESTAMPTZ_TST VALUES(4, '1000000312 23:58:48 IST');`,
			},
			{
				Statement: `SELECT * FROM TIMESTAMPTZ_TST ORDER BY a;`,
				Results:   []sql.Row{{1, `Wed Mar 12 13:58:48 1000 PST`}, {2, `Sun Mar 12 14:58:48 10000 PDT`}, {3, `Sun Mar 12 14:58:48 100000 PDT`}, {3, `Sun Mar 12 14:58:48 10000 PDT`}, {4, `Sun Mar 12 14:58:48 10000 PDT`}, {4, `Sun Mar 12 14:58:48 100000 PDT`}},
			},
			{
				Statement: `DROP TABLE TIMESTAMPTZ_TST;`,
			},
			{
				Statement: `set TimeZone to 'America/New_York';`,
			},
			{
				Statement: `SELECT make_timestamptz(1973, 07, 15, 08, 15, 55.33);`,
				Results:   []sql.Row{{`Sun Jul 15 08:15:55.33 1973 EDT`}},
			},
			{
				Statement: `SELECT make_timestamptz(1973, 07, 15, 08, 15, 55.33, '+2');`,
				Results:   []sql.Row{{`Sun Jul 15 02:15:55.33 1973 EDT`}},
			},
			{
				Statement: `SELECT make_timestamptz(1973, 07, 15, 08, 15, 55.33, '-2');`,
				Results:   []sql.Row{{`Sun Jul 15 06:15:55.33 1973 EDT`}},
			},
			{
				Statement: `WITH tzs (tz) AS (VALUES
    ('+1'), ('+1:'), ('+1:0'), ('+100'), ('+1:00'), ('+01:00'),
    ('+10'), ('+1000'), ('+10:'), ('+10:0'), ('+10:00'), ('+10:00:'),
    ('+10:00:1'), ('+10:00:01'),
    ('+10:00:10'))
     SELECT make_timestamptz(2010, 2, 27, 3, 45, 00, tz), tz FROM tzs;`,
				Results: []sql.Row{{`Fri Feb 26 21:45:00 2010 EST`, +1}, {`Fri Feb 26 21:45:00 2010 EST`, `+1:`}, {`Fri Feb 26 21:45:00 2010 EST`, `+1:0`}, {`Fri Feb 26 21:45:00 2010 EST`, +100}, {`Fri Feb 26 21:45:00 2010 EST`, `+1:00`}, {`Fri Feb 26 21:45:00 2010 EST`, `+01:00`}, {`Fri Feb 26 12:45:00 2010 EST`, +10}, {`Fri Feb 26 12:45:00 2010 EST`, +1000}, {`Fri Feb 26 12:45:00 2010 EST`, `+10:`}, {`Fri Feb 26 12:45:00 2010 EST`, `+10:0`}, {`Fri Feb 26 12:45:00 2010 EST`, `+10:00`}, {`Fri Feb 26 12:45:00 2010 EST`, `+10:00:`}, {`Fri Feb 26 12:44:59 2010 EST`, `+10:00:1`}, {`Fri Feb 26 12:44:59 2010 EST`, `+10:00:01`}, {`Fri Feb 26 12:44:50 2010 EST`, `+10:00:10`}},
			},
			{
				Statement:   `SELECT make_timestamptz(1973, 07, 15, 08, 15, 55.33, '2');`,
				ErrorString: `invalid input syntax for type numeric time zone: "2"`,
			},
			{
				Statement:   `SELECT make_timestamptz(2014, 12, 10, 10, 10, 10, '+16');`,
				ErrorString: `numeric time zone "+16" out of range`,
			},
			{
				Statement:   `SELECT make_timestamptz(2014, 12, 10, 10, 10, 10, '-16');`,
				ErrorString: `numeric time zone "-16" out of range`,
			},
			{
				Statement: `SELECT make_timestamptz(1973, 07, 15, 08, 15, 55.33, '+2') = '1973-07-15 08:15:55.33+02'::timestamptz;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT make_timestamptz(2014, 12, 10, 0, 0, 0, 'Europe/Prague') = timestamptz '2014-12-10 00:00:00 Europe/Prague';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT make_timestamptz(2014, 12, 10, 0, 0, 0, 'Europe/Prague') AT TIME ZONE 'UTC';`,
				Results:   []sql.Row{{`Tue Dec 09 23:00:00 2014`}},
			},
			{
				Statement: `SELECT make_timestamptz(1846, 12, 10, 0, 0, 0, 'Asia/Manila') AT TIME ZONE 'UTC';`,
				Results:   []sql.Row{{`Wed Dec 09 15:56:00 1846`}},
			},
			{
				Statement: `SELECT make_timestamptz(1881, 12, 10, 0, 0, 0, 'Europe/Paris') AT TIME ZONE 'UTC';`,
				Results:   []sql.Row{{`Fri Dec 09 23:50:39 1881`}},
			},
			{
				Statement:   `SELECT make_timestamptz(1910, 12, 24, 0, 0, 0, 'Nehwon/Lankhmar');`,
				ErrorString: `time zone "Nehwon/Lankhmar" not recognized`,
			},
			{
				Statement: `SELECT make_timestamptz(2008, 12, 10, 10, 10, 10, 'EST');`,
				Results:   []sql.Row{{`Wed Dec 10 10:10:10 2008 EST`}},
			},
			{
				Statement: `SELECT make_timestamptz(2008, 12, 10, 10, 10, 10, 'EDT');`,
				Results:   []sql.Row{{`Wed Dec 10 09:10:10 2008 EST`}},
			},
			{
				Statement: `SELECT make_timestamptz(2014, 12, 10, 10, 10, 10, 'PST8PDT');`,
				Results:   []sql.Row{{`Wed Dec 10 13:10:10 2014 EST`}},
			},
			{
				Statement: `RESET TimeZone;`,
			},
			{
				Statement: `select * from generate_series('2020-01-01 00:00'::timestamptz,
                              '2020-01-02 03:00'::timestamptz,
                              '1 hour'::interval);`,
				Results: []sql.Row{{`Wed Jan 01 00:00:00 2020 PST`}, {`Wed Jan 01 01:00:00 2020 PST`}, {`Wed Jan 01 02:00:00 2020 PST`}, {`Wed Jan 01 03:00:00 2020 PST`}, {`Wed Jan 01 04:00:00 2020 PST`}, {`Wed Jan 01 05:00:00 2020 PST`}, {`Wed Jan 01 06:00:00 2020 PST`}, {`Wed Jan 01 07:00:00 2020 PST`}, {`Wed Jan 01 08:00:00 2020 PST`}, {`Wed Jan 01 09:00:00 2020 PST`}, {`Wed Jan 01 10:00:00 2020 PST`}, {`Wed Jan 01 11:00:00 2020 PST`}, {`Wed Jan 01 12:00:00 2020 PST`}, {`Wed Jan 01 13:00:00 2020 PST`}, {`Wed Jan 01 14:00:00 2020 PST`}, {`Wed Jan 01 15:00:00 2020 PST`}, {`Wed Jan 01 16:00:00 2020 PST`}, {`Wed Jan 01 17:00:00 2020 PST`}, {`Wed Jan 01 18:00:00 2020 PST`}, {`Wed Jan 01 19:00:00 2020 PST`}, {`Wed Jan 01 20:00:00 2020 PST`}, {`Wed Jan 01 21:00:00 2020 PST`}, {`Wed Jan 01 22:00:00 2020 PST`}, {`Wed Jan 01 23:00:00 2020 PST`}, {`Thu Jan 02 00:00:00 2020 PST`}, {`Thu Jan 02 01:00:00 2020 PST`}, {`Thu Jan 02 02:00:00 2020 PST`}, {`Thu Jan 02 03:00:00 2020 PST`}},
			},
			{
				Statement: `select generate_series('2022-01-01 00:00'::timestamptz,
                       'infinity'::timestamptz,
                       '1 month'::interval) limit 10;`,
				Results: []sql.Row{{`Sat Jan 01 00:00:00 2022 PST`}, {`Tue Feb 01 00:00:00 2022 PST`}, {`Tue Mar 01 00:00:00 2022 PST`}, {`Fri Apr 01 00:00:00 2022 PDT`}, {`Sun May 01 00:00:00 2022 PDT`}, {`Wed Jun 01 00:00:00 2022 PDT`}, {`Fri Jul 01 00:00:00 2022 PDT`}, {`Mon Aug 01 00:00:00 2022 PDT`}, {`Thu Sep 01 00:00:00 2022 PDT`}, {`Sat Oct 01 00:00:00 2022 PDT`}},
			},
			{
				Statement: `select * from generate_series('2020-01-01 00:00'::timestamptz,
                              '2020-01-02 03:00'::timestamptz,
                              '0 hour'::interval);`,
				ErrorString: `step size cannot equal zero`,
			},
			{
				Statement: `SET TimeZone to 'UTC';`,
			},
			{
				Statement: `SELECT '2011-03-27 00:00:00 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 21:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 01:00:00 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 22:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 01:59:59 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 22:59:59 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:00:00 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:00:01 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:01 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:59:59 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 23:59:59 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 03:00:00 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 03:00:01 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:01 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 04:00:00 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 00:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 00:00:00 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 21:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 01:00:00 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 22:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 01:59:59 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 22:59:59 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:00:00 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 22:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:00:01 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 22:00:01 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:59:59 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 22:59:59 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 03:00:00 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 03:00:01 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:01 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 04:00:00 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 00:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 00:00:00 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 20:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 00:59:59 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 20:59:59 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 01:00:00 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 01:00:01 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:01 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 02:00:00 Europe/Moscow'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 23:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 00:00:00 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 20:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 00:59:59 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 20:59:59 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 01:00:00 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 01:00:01 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:01 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 02:00:00 MSK'::timestamptz;`,
				Results:   []sql.Row{{`Sat Oct 25 23:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 00:00:00'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Mar 26 21:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 01:00:00'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Mar 26 22:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 01:59:59'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Mar 26 22:59:59 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:00:00'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:00:01'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:01 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:59:59'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Mar 26 23:59:59 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 03:00:00'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 03:00:01'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:01 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 04:00:00'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Mar 27 00:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 00:00:00'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Mar 26 21:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 01:00:00'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Mar 26 22:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 01:59:59'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Mar 26 22:59:59 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:00:00'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Mar 26 22:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:00:01'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Mar 26 22:00:01 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 02:59:59'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Mar 26 22:59:59 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 03:00:00'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 03:00:01'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Mar 26 23:00:01 2011 UTC`}},
			},
			{
				Statement: `SELECT '2011-03-27 04:00:00'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Mar 27 00:00:00 2011 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 00:00:00'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Oct 25 20:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 00:59:59'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Oct 25 20:59:59 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 01:00:00'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 01:00:01'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:01 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 02:00:00'::timestamp AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sat Oct 25 23:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 00:00:00'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Oct 25 20:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 00:59:59'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Oct 25 20:59:59 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 01:00:00'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 01:00:01'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:01 2014 UTC`}},
			},
			{
				Statement: `SELECT '2014-10-26 02:00:00'::timestamp AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sat Oct 25 23:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT make_timestamptz(2014, 10, 26, 0, 0, 0, 'MSK');`,
				Results:   []sql.Row{{`Sat Oct 25 20:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT make_timestamptz(2014, 10, 26, 1, 0, 0, 'MSK');`,
				Results:   []sql.Row{{`Sat Oct 25 22:00:00 2014 UTC`}},
			},
			{
				Statement: `SELECT to_timestamp(         0);          -- 1970-01-01 00:00:00+00`,
				Results:   []sql.Row{{`Thu Jan 01 00:00:00 1970 UTC`}},
			},
			{
				Statement: `SELECT to_timestamp( 946684800);          -- 2000-01-01 00:00:00+00`,
				Results:   []sql.Row{{`Sat Jan 01 00:00:00 2000 UTC`}},
			},
			{
				Statement: `SELECT to_timestamp(1262349296.7890123);  -- 2010-01-01 12:34:56.789012+00`,
				Results:   []sql.Row{{`Fri Jan 01 12:34:56.789012 2010 UTC`}},
			},
			{
				Statement: `SELECT to_timestamp(-210866803200);       --   4714-11-24 00:00:00+00 BC`,
				Results:   []sql.Row{{`Mon Nov 24 00:00:00 4714 UTC BC`}},
			},
			{
				Statement: `SELECT to_timestamp(' Infinity'::float);`,
				Results:   []sql.Row{{`infinity`}},
			},
			{
				Statement: `SELECT to_timestamp('-Infinity'::float);`,
				Results:   []sql.Row{{`-infinity`}},
			},
			{
				Statement:   `SELECT to_timestamp('NaN'::float);`,
				ErrorString: `timestamp cannot be NaN`,
			},
			{
				Statement: `SET TimeZone to 'Europe/Moscow';`,
			},
			{
				Statement: `SELECT '2011-03-26 21:00:00 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 00:00:00 2011 MSK`}},
			},
			{
				Statement: `SELECT '2011-03-26 22:00:00 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 01:00:00 2011 MSK`}},
			},
			{
				Statement: `SELECT '2011-03-26 22:59:59 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 01:59:59 2011 MSK`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:00:00 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 03:00:00 2011 MSK`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:00:01 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 03:00:01 2011 MSK`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:59:59 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 03:59:59 2011 MSK`}},
			},
			{
				Statement: `SELECT '2011-03-27 00:00:00 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Mar 27 04:00:00 2011 MSK`}},
			},
			{
				Statement: `SELECT '2014-10-25 21:00:00 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:00 2014 MSK`}},
			},
			{
				Statement: `SELECT '2014-10-25 21:59:59 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Oct 26 01:59:59 2014 MSK`}},
			},
			{
				Statement: `SELECT '2014-10-25 22:00:00 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:00 2014 MSK`}},
			},
			{
				Statement: `SELECT '2014-10-25 22:00:01 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:01 2014 MSK`}},
			},
			{
				Statement: `SELECT '2014-10-25 23:00:00 UTC'::timestamptz;`,
				Results:   []sql.Row{{`Sun Oct 26 02:00:00 2014 MSK`}},
			},
			{
				Statement: `RESET TimeZone;`,
			},
			{
				Statement: `SELECT '2011-03-26 21:00:00 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Mar 27 00:00:00 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 22:00:00 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Mar 27 01:00:00 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 22:59:59 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Mar 27 01:59:59 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:00:00 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Mar 27 03:00:00 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:00:01 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Mar 27 03:00:01 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:59:59 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Mar 27 03:59:59 2011`}},
			},
			{
				Statement: `SELECT '2011-03-27 00:00:00 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Mar 27 04:00:00 2011`}},
			},
			{
				Statement: `SELECT '2014-10-25 21:00:00 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:00 2014`}},
			},
			{
				Statement: `SELECT '2014-10-25 21:59:59 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Oct 26 01:59:59 2014`}},
			},
			{
				Statement: `SELECT '2014-10-25 22:00:00 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:00 2014`}},
			},
			{
				Statement: `SELECT '2014-10-25 22:00:01 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:01 2014`}},
			},
			{
				Statement: `SELECT '2014-10-25 23:00:00 UTC'::timestamptz AT TIME ZONE 'Europe/Moscow';`,
				Results:   []sql.Row{{`Sun Oct 26 02:00:00 2014`}},
			},
			{
				Statement: `SELECT '2011-03-26 21:00:00 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Mar 27 00:00:00 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 22:00:00 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Mar 27 01:00:00 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 22:59:59 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Mar 27 01:59:59 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:00:00 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Mar 27 03:00:00 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:00:01 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Mar 27 03:00:01 2011`}},
			},
			{
				Statement: `SELECT '2011-03-26 23:59:59 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Mar 27 03:59:59 2011`}},
			},
			{
				Statement: `SELECT '2011-03-27 00:00:00 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Mar 27 04:00:00 2011`}},
			},
			{
				Statement: `SELECT '2014-10-25 21:00:00 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:00 2014`}},
			},
			{
				Statement: `SELECT '2014-10-25 21:59:59 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Oct 26 01:59:59 2014`}},
			},
			{
				Statement: `SELECT '2014-10-25 22:00:00 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:00 2014`}},
			},
			{
				Statement: `SELECT '2014-10-25 22:00:01 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Oct 26 01:00:01 2014`}},
			},
			{
				Statement: `SELECT '2014-10-25 23:00:00 UTC'::timestamptz AT TIME ZONE 'MSK';`,
				Results:   []sql.Row{{`Sun Oct 26 02:00:00 2014`}},
			},
			{
				Statement: `create temp table tmptz (f1 timestamptz primary key);`,
			},
			{
				Statement: `insert into tmptz values ('2017-01-18 00:00+00');`,
			},
			{
				Statement: `explain (costs off)
select * from tmptz where f1 at time zone 'utc' = '2017-01-18 00:00';`,
				Results: []sql.Row{{`Seq Scan on tmptz`}, {`Filter: ((f1 AT TIME ZONE 'utc'::text) = 'Wed Jan 18 00:00:00 2017'::timestamp without time zone)`}},
			},
			{
				Statement: `select * from tmptz where f1 at time zone 'utc' = '2017-01-18 00:00';`,
				Results:   []sql.Row{{`Tue Jan 17 16:00:00 2017 PST`}},
			},
		},
	})
}
