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

func TestInterval(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_interval)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_interval,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET DATESTYLE = 'ISO';`,
			},
			{
				Statement: `SET IntervalStyle to postgres;`,
			},
			{
				Statement: `SELECT INTERVAL '01:00' AS "One hour";`,
				Results:   []sql.Row{{`01:00:00`}},
			},
			{
				Statement: `SELECT INTERVAL '+02:00' AS "Two hours";`,
				Results:   []sql.Row{{`02:00:00`}},
			},
			{
				Statement: `SELECT INTERVAL '-08:00' AS "Eight hours";`,
				Results:   []sql.Row{{`-08:00:00`}},
			},
			{
				Statement: `SELECT INTERVAL '-1 +02:03' AS "22 hours ago...";`,
				Results:   []sql.Row{{`-1 days +02:03:00`}},
			},
			{
				Statement: `SELECT INTERVAL '-1 days +02:03' AS "22 hours ago...";`,
				Results:   []sql.Row{{`-1 days +02:03:00`}},
			},
			{
				Statement: `SELECT INTERVAL '1.5 weeks' AS "Ten days twelve hours";`,
				Results:   []sql.Row{{`10 days 12:00:00`}},
			},
			{
				Statement: `SELECT INTERVAL '1.5 months' AS "One month 15 days";`,
				Results:   []sql.Row{{`1 mon 15 days`}},
			},
			{
				Statement: `SELECT INTERVAL '10 years -11 month -12 days +13:14' AS "9 years...";`,
				Results:   []sql.Row{{`9 years 1 mon -12 days +13:14:00`}},
			},
			{
				Statement: `CREATE TABLE INTERVAL_TBL (f1 interval);`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('@ 1 minute');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('@ 5 hour');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('@ 10 day');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('@ 34 year');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('@ 3 months');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('@ 14 seconds ago');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('1 day 2 hours 3 minutes 4 seconds');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('6 years');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('5 months');`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL (f1) VALUES ('5 months 12 hours');`,
			},
			{
				Statement:   `INSERT INTO INTERVAL_TBL (f1) VALUES ('badly formatted interval');`,
				ErrorString: `invalid input syntax for type interval: "badly formatted interval"`,
			},
			{
				Statement:   `INSERT INTO INTERVAL_TBL (f1) VALUES ('@ 30 eons ago');`,
				ErrorString: `invalid input syntax for type interval: "@ 30 eons ago"`,
			},
			{
				Statement: `SELECT * FROM INTERVAL_TBL;`,
				Results:   []sql.Row{{`00:01:00`}, {`05:00:00`}, {`10 days`}, {`34 years`}, {`3 mons`}, {`-00:00:14`}, {`1 day 02:03:04`}, {`6 years`}, {`5 mons`}, {`5 mons 12:00:00`}},
			},
			{
				Statement: `SELECT * FROM INTERVAL_TBL
   WHERE INTERVAL_TBL.f1 <> interval '@ 10 days';`,
				Results: []sql.Row{{`00:01:00`}, {`05:00:00`}, {`34 years`}, {`3 mons`}, {`-00:00:14`}, {`1 day 02:03:04`}, {`6 years`}, {`5 mons`}, {`5 mons 12:00:00`}},
			},
			{
				Statement: `SELECT * FROM INTERVAL_TBL
   WHERE INTERVAL_TBL.f1 <= interval '@ 5 hours';`,
				Results: []sql.Row{{`00:01:00`}, {`05:00:00`}, {`-00:00:14`}},
			},
			{
				Statement: `SELECT * FROM INTERVAL_TBL
   WHERE INTERVAL_TBL.f1 < interval '@ 1 day';`,
				Results: []sql.Row{{`00:01:00`}, {`05:00:00`}, {`-00:00:14`}},
			},
			{
				Statement: `SELECT * FROM INTERVAL_TBL
   WHERE INTERVAL_TBL.f1 = interval '@ 34 years';`,
				Results: []sql.Row{{`34 years`}},
			},
			{
				Statement: `SELECT * FROM INTERVAL_TBL
   WHERE INTERVAL_TBL.f1 >= interval '@ 1 month';`,
				Results: []sql.Row{{`34 years`}, {`3 mons`}, {`6 years`}, {`5 mons`}, {`5 mons 12:00:00`}},
			},
			{
				Statement: `SELECT * FROM INTERVAL_TBL
   WHERE INTERVAL_TBL.f1 > interval '@ 3 seconds ago';`,
				Results: []sql.Row{{`00:01:00`}, {`05:00:00`}, {`10 days`}, {`34 years`}, {`3 mons`}, {`1 day 02:03:04`}, {`6 years`}, {`5 mons`}, {`5 mons 12:00:00`}},
			},
			{
				Statement: `SELECT r1.*, r2.*
   FROM INTERVAL_TBL r1, INTERVAL_TBL r2
   WHERE r1.f1 > r2.f1
   ORDER BY r1.f1, r2.f1;`,
				Results: []sql.Row{{`00:01:00`, `-00:00:14`}, {`05:00:00`, `-00:00:14`}, {`05:00:00`, `00:01:00`}, {`1 day 02:03:04`, `-00:00:14`}, {`1 day 02:03:04`, `00:01:00`}, {`1 day 02:03:04`, `05:00:00`}, {`10 days`, `-00:00:14`}, {`10 days`, `00:01:00`}, {`10 days`, `05:00:00`}, {`10 days`, `1 day 02:03:04`}, {`3 mons`, `-00:00:14`}, {`3 mons`, `00:01:00`}, {`3 mons`, `05:00:00`}, {`3 mons`, `1 day 02:03:04`}, {`3 mons`, `10 days`}, {`5 mons`, `-00:00:14`}, {`5 mons`, `00:01:00`}, {`5 mons`, `05:00:00`}, {`5 mons`, `1 day 02:03:04`}, {`5 mons`, `10 days`}, {`5 mons`, `3 mons`}, {`5 mons 12:00:00`, `-00:00:14`}, {`5 mons 12:00:00`, `00:01:00`}, {`5 mons 12:00:00`, `05:00:00`}, {`5 mons 12:00:00`, `1 day 02:03:04`}, {`5 mons 12:00:00`, `10 days`}, {`5 mons 12:00:00`, `3 mons`}, {`5 mons 12:00:00`, `5 mons`}, {`6 years`, `-00:00:14`}, {`6 years`, `00:01:00`}, {`6 years`, `05:00:00`}, {`6 years`, `1 day 02:03:04`}, {`6 years`, `10 days`}, {`6 years`, `3 mons`}, {`6 years`, `5 mons`}, {`6 years`, `5 mons 12:00:00`}, {`34 years`, `-00:00:14`}, {`34 years`, `00:01:00`}, {`34 years`, `05:00:00`}, {`34 years`, `1 day 02:03:04`}, {`34 years`, `10 days`}, {`34 years`, `3 mons`}, {`34 years`, `5 mons`}, {`34 years`, `5 mons 12:00:00`}, {`34 years`, `6 years`}},
			},
			{
				Statement: `CREATE TEMP TABLE INTERVAL_TBL_OF (f1 interval);`,
			},
			{
				Statement: `INSERT INTO INTERVAL_TBL_OF (f1) VALUES
  ('2147483647 days 2147483647 months'),
  ('2147483647 days -2147483648 months'),
  ('1 year'),
  ('-2147483648 days 2147483647 months'),
  ('-2147483648 days -2147483648 months');`,
			},
			{
				Statement:   `INSERT INTO INTERVAL_TBL_OF (f1) VALUES ('2147483648 days');`,
				ErrorString: `interval field value out of range: "2147483648 days"`,
			},
			{
				Statement:   `INSERT INTO INTERVAL_TBL_OF (f1) VALUES ('-2147483649 days');`,
				ErrorString: `interval field value out of range: "-2147483649 days"`,
			},
			{
				Statement:   `INSERT INTO INTERVAL_TBL_OF (f1) VALUES ('2147483647 years');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement:   `INSERT INTO INTERVAL_TBL_OF (f1) VALUES ('-2147483648 years');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement:   `select extract(epoch from '256 microseconds'::interval * (2^55)::float8);`,
				ErrorString: `interval out of range`,
			},
			{
				Statement: `SELECT r1.*, r2.*
   FROM INTERVAL_TBL_OF r1, INTERVAL_TBL_OF r2
   WHERE r1.f1 > r2.f1
   ORDER BY r1.f1, r2.f1;`,
				Results: []sql.Row{{`-178956970 years -8 mons +2147483647 days`, `-178956970 years -8 mons -2147483648 days`}, {`1 year`, `-178956970 years -8 mons -2147483648 days`}, {`1 year`, `-178956970 years -8 mons +2147483647 days`}, {`178956970 years 7 mons -2147483648 days`, `-178956970 years -8 mons -2147483648 days`}, {`178956970 years 7 mons -2147483648 days`, `-178956970 years -8 mons +2147483647 days`}, {`178956970 years 7 mons -2147483648 days`, `1 year`}, {`178956970 years 7 mons 2147483647 days`, `-178956970 years -8 mons -2147483648 days`}, {`178956970 years 7 mons 2147483647 days`, `-178956970 years -8 mons +2147483647 days`}, {`178956970 years 7 mons 2147483647 days`, `1 year`}, {`178956970 years 7 mons 2147483647 days`, `178956970 years 7 mons -2147483648 days`}},
			},
			{
				Statement: `CREATE INDEX ON INTERVAL_TBL_OF USING btree (f1);`,
			},
			{
				Statement: `SET enable_seqscan TO false;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT f1 FROM INTERVAL_TBL_OF r1 ORDER BY f1;`,
				Results: []sql.Row{{`Index Only Scan using interval_tbl_of_f1_idx on interval_tbl_of r1`}},
			},
			{
				Statement: `SELECT f1 FROM INTERVAL_TBL_OF r1 ORDER BY f1;`,
				Results:   []sql.Row{{`-178956970 years -8 mons -2147483648 days`}, {`-178956970 years -8 mons +2147483647 days`}, {`1 year`}, {`178956970 years 7 mons -2147483648 days`}, {`178956970 years 7 mons 2147483647 days`}},
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `DROP TABLE INTERVAL_TBL_OF;`,
			},
			{
				Statement: `CREATE TABLE INTERVAL_MULDIV_TBL (span interval);`,
			},
			{
				Statement: `COPY INTERVAL_MULDIV_TBL FROM STDIN;`,
			},
			{
				Statement: `SELECT span * 0.3 AS product
FROM INTERVAL_MULDIV_TBL;`,
				Results: []sql.Row{{`1 year 12 days 122:24:00`}, {`-1 years -12 days +93:36:00`}, {`-3 days -14:24:00`}, {`2 mons 13 days 01:22:28.8`}, {`-10 mons +120 days 37:28:21.6567`}, {`1 mon 6 days`}, {`4 mons 6 days`}, {`24 years 11 mons 320 days 16:48:00`}},
			},
			{
				Statement: `SELECT span * 8.2 AS product
FROM INTERVAL_MULDIV_TBL;`,
				Results: []sql.Row{{`28 years 104 days 2961:36:00`}, {`-28 years -104 days +2942:24:00`}, {`-98 days -09:36:00`}, {`6 years 1 mon -197 days +93:34:27.2`}, {`-24 years -7 mons +3946 days 640:15:11.9498`}, {`2 years 8 mons 24 days`}, {`9 years 6 mons 24 days`}, {`682 years 7 mons 8215 days 19:12:00`}},
			},
			{
				Statement: `SELECT span / 10 AS quotient
FROM INTERVAL_MULDIV_TBL;`,
				Results: []sql.Row{{`4 mons 4 days 40:48:00`}, {`-4 mons -4 days +31:12:00`}, {`-1 days -04:48:00`}, {`25 days -15:32:30.4`}, {`-3 mons +30 days 12:29:27.2189`}, {`12 days`}, {`1 mon 12 days`}, {`8 years 3 mons 126 days 21:36:00`}},
			},
			{
				Statement: `SELECT span / 100 AS quotient
FROM INTERVAL_MULDIV_TBL;`,
				Results: []sql.Row{{`12 days 13:40:48`}, {`-12 days -06:28:48`}, {`-02:52:48`}, {`2 days 10:26:44.96`}, {`-6 days +01:14:56.72189`}, {`1 day 04:48:00`}, {`4 days 04:48:00`}, {`9 mons 39 days 16:33:36`}},
			},
			{
				Statement: `DROP TABLE INTERVAL_MULDIV_TBL;`,
			},
			{
				Statement: `SET DATESTYLE = 'postgres';`,
			},
			{
				Statement: `SET IntervalStyle to postgres_verbose;`,
			},
			{
				Statement: `SELECT * FROM INTERVAL_TBL;`,
				Results:   []sql.Row{{`@ 1 min`}, {`@ 5 hours`}, {`@ 10 days`}, {`@ 34 years`}, {`@ 3 mons`}, {`@ 14 secs ago`}, {`@ 1 day 2 hours 3 mins 4 secs`}, {`@ 6 years`}, {`@ 5 mons`}, {`@ 5 mons 12 hours`}},
			},
			{
				Statement: `select avg(f1) from interval_tbl;`,
				Results:   []sql.Row{{`@ 4 years 1 mon 10 days 4 hours 18 mins 23 secs`}},
			},
			{
				Statement: `select '4 millenniums 5 centuries 4 decades 1 year 4 months 4 days 17 minutes 31 seconds'::interval;`,
				Results:   []sql.Row{{`@ 4541 years 4 mons 4 days 17 mins 31 secs`}},
			},
			{
				Statement: `select '100000000y 10mon -1000000000d -100000h -10min -10.000001s ago'::interval;`,
				Results:   []sql.Row{{`@ 100000000 years 10 mons -1000000000 days -100000 hours -10 mins -10.000001 secs ago`}},
			},
			{
				Statement: `SELECT justify_hours(interval '6 months 3 days 52 hours 3 minutes 2 seconds') as "6 mons 5 days 4 hours 3 mins 2 seconds";`,
				Results:   []sql.Row{{`@ 6 mons 5 days 4 hours 3 mins 2 secs`}},
			},
			{
				Statement: `SELECT justify_days(interval '6 months 36 days 5 hours 4 minutes 3 seconds') as "7 mons 6 days 5 hours 4 mins 3 seconds";`,
				Results:   []sql.Row{{`@ 7 mons 6 days 5 hours 4 mins 3 secs`}},
			},
			{
				Statement:   `SELECT justify_hours(interval '2147483647 days 24 hrs');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement:   `SELECT justify_days(interval '2147483647 months 30 days');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement: `SELECT justify_interval(interval '1 month -1 hour') as "1 month -1 hour";`,
				Results:   []sql.Row{{`@ 29 days 23 hours`}},
			},
			{
				Statement: `SELECT justify_interval(interval '2147483647 days 24 hrs');`,
				Results:   []sql.Row{{`@ 5965232 years 4 mons 8 days`}},
			},
			{
				Statement: `SELECT justify_interval(interval '-2147483648 days -24 hrs');`,
				Results:   []sql.Row{{`@ 5965232 years 4 mons 9 days ago`}},
			},
			{
				Statement:   `SELECT justify_interval(interval '2147483647 months 30 days');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement:   `SELECT justify_interval(interval '-2147483648 months -30 days');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement: `SELECT justify_interval(interval '2147483647 months 30 days -24 hrs');`,
				Results:   []sql.Row{{`@ 178956970 years 7 mons 29 days`}},
			},
			{
				Statement: `SELECT justify_interval(interval '-2147483648 months -30 days 24 hrs');`,
				Results:   []sql.Row{{`@ 178956970 years 8 mons 29 days ago`}},
			},
			{
				Statement:   `SELECT justify_interval(interval '2147483647 months -30 days 1440 hrs');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement:   `SELECT justify_interval(interval '-2147483648 months 30 days -1440 hrs');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement: `SET DATESTYLE = 'ISO';`,
			},
			{
				Statement: `SET IntervalStyle TO postgres;`,
			},
			{
				Statement: `SELECT '1 millisecond'::interval, '1 microsecond'::interval,
       '500 seconds 99 milliseconds 51 microseconds'::interval;`,
				Results: []sql.Row{{`00:00:00.001`, `00:00:00.000001`, `00:08:20.099051`}},
			},
			{
				Statement: `SELECT '3 days 5 milliseconds'::interval;`,
				Results:   []sql.Row{{`3 days 00:00:00.005`}},
			},
			{
				Statement:   `SELECT '1 second 2 seconds'::interval;              -- error`,
				ErrorString: `invalid input syntax for type interval: "1 second 2 seconds"`,
			},
			{
				Statement:   `SELECT '10 milliseconds 20 milliseconds'::interval; -- error`,
				ErrorString: `invalid input syntax for type interval: "10 milliseconds 20 milliseconds"`,
			},
			{
				Statement:   `SELECT '5.5 seconds 3 milliseconds'::interval;      -- error`,
				ErrorString: `invalid input syntax for type interval: "5.5 seconds 3 milliseconds"`,
			},
			{
				Statement:   `SELECT '1:20:05 5 microseconds'::interval;          -- error`,
				ErrorString: `invalid input syntax for type interval: "1:20:05 5 microseconds"`,
			},
			{
				Statement:   `SELECT '1 day 1 day'::interval;                     -- error`,
				ErrorString: `invalid input syntax for type interval: "1 day 1 day"`,
			},
			{
				Statement: `SELECT interval '1-2';  -- SQL year-month literal`,
				Results:   []sql.Row{{`1 year 2 mons`}},
			},
			{
				Statement: `SELECT interval '999' second;  -- oversize leading field is ok`,
				Results:   []sql.Row{{`00:16:39`}},
			},
			{
				Statement: `SELECT interval '999' minute;`,
				Results:   []sql.Row{{`16:39:00`}},
			},
			{
				Statement: `SELECT interval '999' hour;`,
				Results:   []sql.Row{{`999:00:00`}},
			},
			{
				Statement: `SELECT interval '999' day;`,
				Results:   []sql.Row{{`999 days`}},
			},
			{
				Statement: `SELECT interval '999' month;`,
				Results:   []sql.Row{{`83 years 3 mons`}},
			},
			{
				Statement: `SELECT interval '1' year;`,
				Results:   []sql.Row{{`1 year`}},
			},
			{
				Statement: `SELECT interval '2' month;`,
				Results:   []sql.Row{{`2 mons`}},
			},
			{
				Statement: `SELECT interval '3' day;`,
				Results:   []sql.Row{{`3 days`}},
			},
			{
				Statement: `SELECT interval '4' hour;`,
				Results:   []sql.Row{{`04:00:00`}},
			},
			{
				Statement: `SELECT interval '5' minute;`,
				Results:   []sql.Row{{`00:05:00`}},
			},
			{
				Statement: `SELECT interval '6' second;`,
				Results:   []sql.Row{{`00:00:06`}},
			},
			{
				Statement: `SELECT interval '1' year to month;`,
				Results:   []sql.Row{{`1 mon`}},
			},
			{
				Statement: `SELECT interval '1-2' year to month;`,
				Results:   []sql.Row{{`1 year 2 mons`}},
			},
			{
				Statement: `SELECT interval '1 2' day to hour;`,
				Results:   []sql.Row{{`1 day 02:00:00`}},
			},
			{
				Statement: `SELECT interval '1 2:03' day to hour;`,
				Results:   []sql.Row{{`1 day 02:00:00`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04' day to hour;`,
				Results:   []sql.Row{{`1 day 02:00:00`}},
			},
			{
				Statement:   `SELECT interval '1 2' day to minute;`,
				ErrorString: `invalid input syntax for type interval: "1 2"`,
			},
			{
				Statement: `SELECT interval '1 2:03' day to minute;`,
				Results:   []sql.Row{{`1 day 02:03:00`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04' day to minute;`,
				Results:   []sql.Row{{`1 day 02:03:00`}},
			},
			{
				Statement:   `SELECT interval '1 2' day to second;`,
				ErrorString: `invalid input syntax for type interval: "1 2"`,
			},
			{
				Statement: `SELECT interval '1 2:03' day to second;`,
				Results:   []sql.Row{{`1 day 02:03:00`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04' day to second;`,
				Results:   []sql.Row{{`1 day 02:03:04`}},
			},
			{
				Statement:   `SELECT interval '1 2' hour to minute;`,
				ErrorString: `invalid input syntax for type interval: "1 2"`,
			},
			{
				Statement: `SELECT interval '1 2:03' hour to minute;`,
				Results:   []sql.Row{{`1 day 02:03:00`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04' hour to minute;`,
				Results:   []sql.Row{{`1 day 02:03:00`}},
			},
			{
				Statement:   `SELECT interval '1 2' hour to second;`,
				ErrorString: `invalid input syntax for type interval: "1 2"`,
			},
			{
				Statement: `SELECT interval '1 2:03' hour to second;`,
				Results:   []sql.Row{{`1 day 02:03:00`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04' hour to second;`,
				Results:   []sql.Row{{`1 day 02:03:04`}},
			},
			{
				Statement:   `SELECT interval '1 2' minute to second;`,
				ErrorString: `invalid input syntax for type interval: "1 2"`,
			},
			{
				Statement: `SELECT interval '1 2:03' minute to second;`,
				Results:   []sql.Row{{`1 day 00:02:03`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04' minute to second;`,
				Results:   []sql.Row{{`1 day 02:03:04`}},
			},
			{
				Statement: `SELECT interval '1 +2:03' minute to second;`,
				Results:   []sql.Row{{`1 day 00:02:03`}},
			},
			{
				Statement: `SELECT interval '1 +2:03:04' minute to second;`,
				Results:   []sql.Row{{`1 day 02:03:04`}},
			},
			{
				Statement: `SELECT interval '1 -2:03' minute to second;`,
				Results:   []sql.Row{{`1 day -00:02:03`}},
			},
			{
				Statement: `SELECT interval '1 -2:03:04' minute to second;`,
				Results:   []sql.Row{{`1 day -02:03:04`}},
			},
			{
				Statement: `SELECT interval '123 11' day to hour; -- ok`,
				Results:   []sql.Row{{`123 days 11:00:00`}},
			},
			{
				Statement:   `SELECT interval '123 11' day; -- not ok`,
				ErrorString: `invalid input syntax for type interval: "123 11"`,
			},
			{
				Statement:   `SELECT interval '123 11'; -- not ok, too ambiguous`,
				ErrorString: `invalid input syntax for type interval: "123 11"`,
			},
			{
				Statement:   `SELECT interval '123 2:03 -2:04'; -- not ok, redundant hh:mm fields`,
				ErrorString: `invalid input syntax for type interval: "123 2:03 -2:04"`,
			},
			{
				Statement: `SELECT interval(0) '1 day 01:23:45.6789';`,
				Results:   []sql.Row{{`1 day 01:23:46`}},
			},
			{
				Statement: `SELECT interval(2) '1 day 01:23:45.6789';`,
				Results:   []sql.Row{{`1 day 01:23:45.68`}},
			},
			{
				Statement: `SELECT interval '12:34.5678' minute to second(2);  -- per SQL spec`,
				Results:   []sql.Row{{`00:12:34.57`}},
			},
			{
				Statement: `SELECT interval '1.234' second;`,
				Results:   []sql.Row{{`00:00:01.234`}},
			},
			{
				Statement: `SELECT interval '1.234' second(2);`,
				Results:   []sql.Row{{`00:00:01.23`}},
			},
			{
				Statement:   `SELECT interval '1 2.345' day to second(2);`,
				ErrorString: `invalid input syntax for type interval: "1 2.345"`,
			},
			{
				Statement: `SELECT interval '1 2:03' day to second(2);`,
				Results:   []sql.Row{{`1 day 02:03:00`}},
			},
			{
				Statement: `SELECT interval '1 2:03.4567' day to second(2);`,
				Results:   []sql.Row{{`1 day 00:02:03.46`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04.5678' day to second(2);`,
				Results:   []sql.Row{{`1 day 02:03:04.57`}},
			},
			{
				Statement:   `SELECT interval '1 2.345' hour to second(2);`,
				ErrorString: `invalid input syntax for type interval: "1 2.345"`,
			},
			{
				Statement: `SELECT interval '1 2:03.45678' hour to second(2);`,
				Results:   []sql.Row{{`1 day 00:02:03.46`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04.5678' hour to second(2);`,
				Results:   []sql.Row{{`1 day 02:03:04.57`}},
			},
			{
				Statement:   `SELECT interval '1 2.3456' minute to second(2);`,
				ErrorString: `invalid input syntax for type interval: "1 2.3456"`,
			},
			{
				Statement: `SELECT interval '1 2:03.5678' minute to second(2);`,
				Results:   []sql.Row{{`1 day 00:02:03.57`}},
			},
			{
				Statement: `SELECT interval '1 2:03:04.5678' minute to second(2);`,
				Results:   []sql.Row{{`1 day 02:03:04.57`}},
			},
			{
				Statement: `SELECT f1, f1::INTERVAL DAY TO MINUTE AS "minutes",
  (f1 + INTERVAL '1 month')::INTERVAL MONTH::INTERVAL YEAR AS "years"
  FROM interval_tbl;`,
				Results: []sql.Row{{`00:01:00`, `00:01:00`, `00:00:00`}, {`05:00:00`, `05:00:00`, `00:00:00`}, {`10 days`, `10 days`, `00:00:00`}, {`34 years`, `34 years`, `34 years`}, {`3 mons`, `3 mons`, `00:00:00`}, {`-00:00:14`, `00:00:00`, `00:00:00`}, {`1 day 02:03:04`, `1 day 02:03:00`, `00:00:00`}, {`6 years`, `6 years`, `6 years`}, {`5 mons`, `5 mons`, `00:00:00`}, {`5 mons 12:00:00`, `5 mons 12:00:00`, `00:00:00`}},
			},
			{
				Statement: `SET IntervalStyle TO sql_standard;`,
			},
			{
				Statement: `SELECT  interval '0'                       AS "zero",
        interval '1-2' year to month       AS "year-month",
        interval '1 2:03:04' day to second AS "day-time",
        - interval '1-2'                   AS "negative year-month",
        - interval '1 2:03:04'             AS "negative day-time";`,
				Results: []sql.Row{{0, `1-2`, `1 2:03:04`, `-1-2`, `-1 2:03:04`}},
			},
			{
				Statement: `SET IntervalStyle TO postgres;`,
			},
			{
				Statement: `SELECT  interval '+1 -1:00:00',
        interval '-1 +1:00:00',
        interval '+1-2 -3 +4:05:06.789',
        interval '-1-2 +3 -4:05:06.789';`,
				Results: []sql.Row{{`1 day -01:00:00`, `-1 days +01:00:00`, `1 year 2 mons -3 days +04:05:06.789`, `-1 years -2 mons +3 days -04:05:06.789`}},
			},
			{
				Statement: `SELECT  interval '-23 hours 45 min 12.34 sec',
        interval '-1 day 23 hours 45 min 12.34 sec',
        interval '-1 year 2 months 1 day 23 hours 45 min 12.34 sec',
        interval '-1 year 2 months 1 day 23 hours 45 min +12.34 sec';`,
				Results: []sql.Row{{`-22:14:47.66`, `-1 days +23:45:12.34`, `-10 mons +1 day 23:45:12.34`, `-10 mons +1 day 23:45:12.34`}},
			},
			{
				Statement: `SET IntervalStyle TO sql_standard;`,
			},
			{
				Statement: `SELECT  interval '1 day -1 hours',
        interval '-1 days +1 hours',
        interval '1 years 2 months -3 days 4 hours 5 minutes 6.789 seconds',
        - interval '1 years 2 months -3 days 4 hours 5 minutes 6.789 seconds';`,
				Results: []sql.Row{{`+0-0 +1 -1:00:00`, `+0-0 -1 +1:00:00`, `+1-2 -3 +4:05:06.789`, `-1-2 +3 -4:05:06.789`}},
			},
			{
				Statement: `SELECT  interval '-23 hours 45 min 12.34 sec',
        interval '-1 day 23 hours 45 min 12.34 sec',
        interval '-1 year 2 months 1 day 23 hours 45 min 12.34 sec',
        interval '-1 year 2 months 1 day 23 hours 45 min +12.34 sec';`,
				Results: []sql.Row{{`-23:45:12.34`, `-1 23:45:12.34`, `-1-2 -1 -23:45:12.34`, `-0-10 +1 +23:45:12.34`}},
			},
			{
				Statement:   `SELECT  interval '';  -- error`,
				ErrorString: `invalid input syntax for type interval: ""`,
			},
			{
				Statement: `SET IntervalStyle to iso_8601;`,
			},
			{
				Statement: `select  interval '0'                                AS "zero",
        interval '1-2'                              AS "a year 2 months",
        interval '1 2:03:04'                        AS "a bit over a day",
        interval '2:03:04.45679'                    AS "a bit over 2 hours",
        (interval '1-2' + interval '3 4:05:06.7')   AS "all fields",
        (interval '1-2' - interval '3 4:05:06.7')   AS "mixed sign",
        (- interval '1-2' + interval '3 4:05:06.7') AS "negative";`,
				Results: []sql.Row{{`PT0S`, `P1Y2M`, `P1DT2H3M4S`, `PT2H3M4.45679S`, `P1Y2M3DT4H5M6.7S`, `P1Y2M-3DT-4H-5M-6.7S`, `P-1Y-2M3DT4H5M6.7S`}},
			},
			{
				Statement: `SET IntervalStyle to sql_standard;`,
			},
			{
				Statement: `select  interval 'P0Y'                    AS "zero",
        interval 'P1Y2M'                  AS "a year 2 months",
        interval 'P1W'                    AS "a week",
        interval 'P1DT2H3M4S'             AS "a bit over a day",
        interval 'P1Y2M3DT4H5M6.7S'       AS "all fields",
        interval 'P-1Y-2M-3DT-4H-5M-6.7S' AS "negative",
        interval 'PT-0.1S'                AS "fractional second";`,
				Results: []sql.Row{{0, `1-2`, `7 0:00:00`, `1 2:03:04`, `+1-2 +3 +4:05:06.7`, `-1-2 -3 -4:05:06.7`, `-0:00:00.1`}},
			},
			{
				Statement: `SET IntervalStyle to postgres;`,
			},
			{
				Statement: `select  interval 'P00021015T103020'       AS "ISO8601 Basic Format",
        interval 'P0002-10-15T10:30:20'   AS "ISO8601 Extended Format";`,
				Results: []sql.Row{{`2 years 10 mons 15 days 10:30:20`, `2 years 10 mons 15 days 10:30:20`}},
			},
			{
				Statement: `select  interval 'P0002'                  AS "year only",
        interval 'P0002-10'               AS "year month",
        interval 'P0002-10-15'            AS "year month day",
        interval 'P0002T1S'               AS "year only plus time",
        interval 'P0002-10T1S'            AS "year month plus time",
        interval 'P0002-10-15T1S'         AS "year month day plus time",
        interval 'PT10'                   AS "hour only",
        interval 'PT10:30'                AS "hour minute";`,
				Results: []sql.Row{{`2 years`, `2 years 10 mons`, `2 years 10 mons 15 days`, `2 years 00:00:01`, `2 years 10 mons 00:00:01`, `2 years 10 mons 15 days 00:00:01`, `10:00:00`, `10:30:00`}},
			},
			{
				Statement: `select interval 'P1Y0M3DT4H5M6S';`,
				Results:   []sql.Row{{`1 year 3 days 04:05:06`}},
			},
			{
				Statement: `select interval 'P1.0Y0M3DT4H5M6S';`,
				Results:   []sql.Row{{`1 year 3 days 04:05:06`}},
			},
			{
				Statement: `select interval 'P1.1Y0M3DT4H5M6S';`,
				Results:   []sql.Row{{`1 year 1 mon 3 days 04:05:06`}},
			},
			{
				Statement: `select interval 'P1.Y0M3DT4H5M6S';`,
				Results:   []sql.Row{{`1 year 3 days 04:05:06`}},
			},
			{
				Statement: `select interval 'P.1Y0M3DT4H5M6S';`,
				Results:   []sql.Row{{`1 mon 3 days 04:05:06`}},
			},
			{
				Statement: `select interval 'P10.5e4Y';  -- not per spec, but we've historically taken it`,
				Results:   []sql.Row{{`105000 years`}},
			},
			{
				Statement:   `select interval 'P.Y0M3DT4H5M6S';  -- error`,
				ErrorString: `invalid input syntax for type interval: "P.Y0M3DT4H5M6S"`,
			},
			{
				Statement: `SET IntervalStyle to postgres_verbose;`,
			},
			{
				Statement: `select interval '-10 mons -3 days +03:55:06.70';`,
				Results:   []sql.Row{{`@ 10 mons 3 days -3 hours -55 mins -6.7 secs ago`}},
			},
			{
				Statement: `select interval '1 year 2 mons 3 days 04:05:06.699999';`,
				Results:   []sql.Row{{`@ 1 year 2 mons 3 days 4 hours 5 mins 6.699999 secs`}},
			},
			{
				Statement: `select interval '0:0:0.7', interval '@ 0.70 secs', interval '0.7 seconds';`,
				Results:   []sql.Row{{`@ 0.7 secs`, `@ 0.7 secs`, `@ 0.7 secs`}},
			},
			{
				Statement: `select interval '2562047788.01521550194 hours';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775807 secs`}},
			},
			{
				Statement: `select interval '-2562047788.01521550222 hours';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775808 secs ago`}},
			},
			{
				Statement: `select interval '153722867280.912930117 minutes';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775807 secs`}},
			},
			{
				Statement: `select interval '-153722867280.912930133 minutes';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775808 secs ago`}},
			},
			{
				Statement: `select interval '9223372036854.775807 seconds';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775807 secs`}},
			},
			{
				Statement: `select interval '-9223372036854.775808 seconds';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775808 secs ago`}},
			},
			{
				Statement: `select interval '9223372036854775.807 milliseconds';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775807 secs`}},
			},
			{
				Statement: `select interval '-9223372036854775.808 milliseconds';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775808 secs ago`}},
			},
			{
				Statement: `select interval '9223372036854775807 microseconds';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775807 secs`}},
			},
			{
				Statement: `select interval '-9223372036854775808 microseconds';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775808 secs ago`}},
			},
			{
				Statement: `select interval 'PT2562047788H54.775807S';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775807 secs`}},
			},
			{
				Statement: `select interval 'PT-2562047788H-54.775808S';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775808 secs ago`}},
			},
			{
				Statement: `select interval 'PT2562047788:00:54.775807';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775807 secs`}},
			},
			{
				Statement: `select interval 'PT2562047788.0152155019444';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775429 secs`}},
			},
			{
				Statement: `select interval 'PT-2562047788.0152155022222';`,
				Results:   []sql.Row{{`@ 2562047788 hours 54.775429 secs ago`}},
			},
			{
				Statement:   `select interval '2147483648 years';`,
				ErrorString: `interval field value out of range: "2147483648 years"`,
			},
			{
				Statement:   `select interval '-2147483649 years';`,
				ErrorString: `interval field value out of range: "-2147483649 years"`,
			},
			{
				Statement:   `select interval '2147483648 months';`,
				ErrorString: `interval field value out of range: "2147483648 months"`,
			},
			{
				Statement:   `select interval '-2147483649 months';`,
				ErrorString: `interval field value out of range: "-2147483649 months"`,
			},
			{
				Statement:   `select interval '2147483648 days';`,
				ErrorString: `interval field value out of range: "2147483648 days"`,
			},
			{
				Statement:   `select interval '-2147483649 days';`,
				ErrorString: `interval field value out of range: "-2147483649 days"`,
			},
			{
				Statement:   `select interval '2562047789 hours';`,
				ErrorString: `interval field value out of range: "2562047789 hours"`,
			},
			{
				Statement:   `select interval '-2562047789 hours';`,
				ErrorString: `interval field value out of range: "-2562047789 hours"`,
			},
			{
				Statement:   `select interval '153722867281 minutes';`,
				ErrorString: `interval field value out of range: "153722867281 minutes"`,
			},
			{
				Statement:   `select interval '-153722867281 minutes';`,
				ErrorString: `interval field value out of range: "-153722867281 minutes"`,
			},
			{
				Statement:   `select interval '9223372036855 seconds';`,
				ErrorString: `interval field value out of range: "9223372036855 seconds"`,
			},
			{
				Statement:   `select interval '-9223372036855 seconds';`,
				ErrorString: `interval field value out of range: "-9223372036855 seconds"`,
			},
			{
				Statement:   `select interval '9223372036854777 millisecond';`,
				ErrorString: `interval field value out of range: "9223372036854777 millisecond"`,
			},
			{
				Statement:   `select interval '-9223372036854777 millisecond';`,
				ErrorString: `interval field value out of range: "-9223372036854777 millisecond"`,
			},
			{
				Statement:   `select interval '9223372036854775808 microsecond';`,
				ErrorString: `interval field value out of range: "9223372036854775808 microsecond"`,
			},
			{
				Statement:   `select interval '-9223372036854775809 microsecond';`,
				ErrorString: `interval field value out of range: "-9223372036854775809 microsecond"`,
			},
			{
				Statement:   `select interval 'P2147483648';`,
				ErrorString: `interval field value out of range: "P2147483648"`,
			},
			{
				Statement:   `select interval 'P-2147483649';`,
				ErrorString: `interval field value out of range: "P-2147483649"`,
			},
			{
				Statement:   `select interval 'P1-2147483647-2147483647';`,
				ErrorString: `interval out of range`,
			},
			{
				Statement:   `select interval 'PT2562047789';`,
				ErrorString: `interval field value out of range: "PT2562047789"`,
			},
			{
				Statement:   `select interval 'PT-2562047789';`,
				ErrorString: `interval field value out of range: "PT-2562047789"`,
			},
			{
				Statement:   `select interval '2147483647 weeks';`,
				ErrorString: `interval field value out of range: "2147483647 weeks"`,
			},
			{
				Statement:   `select interval '-2147483648 weeks';`,
				ErrorString: `interval field value out of range: "-2147483648 weeks"`,
			},
			{
				Statement:   `select interval '2147483647 decades';`,
				ErrorString: `interval field value out of range: "2147483647 decades"`,
			},
			{
				Statement:   `select interval '-2147483648 decades';`,
				ErrorString: `interval field value out of range: "-2147483648 decades"`,
			},
			{
				Statement:   `select interval '2147483647 centuries';`,
				ErrorString: `interval field value out of range: "2147483647 centuries"`,
			},
			{
				Statement:   `select interval '-2147483648 centuries';`,
				ErrorString: `interval field value out of range: "-2147483648 centuries"`,
			},
			{
				Statement:   `select interval '2147483647 millennium';`,
				ErrorString: `interval field value out of range: "2147483647 millennium"`,
			},
			{
				Statement:   `select interval '-2147483648 millennium';`,
				ErrorString: `interval field value out of range: "-2147483648 millennium"`,
			},
			{
				Statement:   `select interval '1 week 2147483647 days';`,
				ErrorString: `interval field value out of range: "1 week 2147483647 days"`,
			},
			{
				Statement:   `select interval '-1 week -2147483648 days';`,
				ErrorString: `interval field value out of range: "-1 week -2147483648 days"`,
			},
			{
				Statement:   `select interval '2147483647 days 1 week';`,
				ErrorString: `interval field value out of range: "2147483647 days 1 week"`,
			},
			{
				Statement:   `select interval '-2147483648 days -1 week';`,
				ErrorString: `interval field value out of range: "-2147483648 days -1 week"`,
			},
			{
				Statement:   `select interval 'P1W2147483647D';`,
				ErrorString: `interval field value out of range: "P1W2147483647D"`,
			},
			{
				Statement:   `select interval 'P-1W-2147483648D';`,
				ErrorString: `interval field value out of range: "P-1W-2147483648D"`,
			},
			{
				Statement:   `select interval 'P2147483647D1W';`,
				ErrorString: `interval field value out of range: "P2147483647D1W"`,
			},
			{
				Statement:   `select interval 'P-2147483648D-1W';`,
				ErrorString: `interval field value out of range: "P-2147483648D-1W"`,
			},
			{
				Statement:   `select interval '1 decade 2147483647 years';`,
				ErrorString: `interval field value out of range: "1 decade 2147483647 years"`,
			},
			{
				Statement:   `select interval '1 century 2147483647 years';`,
				ErrorString: `interval field value out of range: "1 century 2147483647 years"`,
			},
			{
				Statement:   `select interval '1 millennium 2147483647 years';`,
				ErrorString: `interval field value out of range: "1 millennium 2147483647 years"`,
			},
			{
				Statement:   `select interval '-1 decade -2147483648 years';`,
				ErrorString: `interval field value out of range: "-1 decade -2147483648 years"`,
			},
			{
				Statement:   `select interval '-1 century -2147483648 years';`,
				ErrorString: `interval field value out of range: "-1 century -2147483648 years"`,
			},
			{
				Statement:   `select interval '-1 millennium -2147483648 years';`,
				ErrorString: `interval field value out of range: "-1 millennium -2147483648 years"`,
			},
			{
				Statement:   `select interval '2147483647 years 1 decade';`,
				ErrorString: `interval field value out of range: "2147483647 years 1 decade"`,
			},
			{
				Statement:   `select interval '2147483647 years 1 century';`,
				ErrorString: `interval field value out of range: "2147483647 years 1 century"`,
			},
			{
				Statement:   `select interval '2147483647 years 1 millennium';`,
				ErrorString: `interval field value out of range: "2147483647 years 1 millennium"`,
			},
			{
				Statement:   `select interval '-2147483648 years -1 decade';`,
				ErrorString: `interval field value out of range: "-2147483648 years -1 decade"`,
			},
			{
				Statement:   `select interval '-2147483648 years -1 century';`,
				ErrorString: `interval field value out of range: "-2147483648 years -1 century"`,
			},
			{
				Statement:   `select interval '-2147483648 years -1 millennium';`,
				ErrorString: `interval field value out of range: "-2147483648 years -1 millennium"`,
			},
			{
				Statement:   `select interval '0.1 millennium 2147483647 months';`,
				ErrorString: `interval field value out of range: "0.1 millennium 2147483647 months"`,
			},
			{
				Statement:   `select interval '0.1 centuries 2147483647 months';`,
				ErrorString: `interval field value out of range: "0.1 centuries 2147483647 months"`,
			},
			{
				Statement:   `select interval '0.1 decades 2147483647 months';`,
				ErrorString: `interval field value out of range: "0.1 decades 2147483647 months"`,
			},
			{
				Statement:   `select interval '0.1 yrs 2147483647 months';`,
				ErrorString: `interval field value out of range: "0.1 yrs 2147483647 months"`,
			},
			{
				Statement:   `select interval '-0.1 millennium -2147483648 months';`,
				ErrorString: `interval field value out of range: "-0.1 millennium -2147483648 months"`,
			},
			{
				Statement:   `select interval '-0.1 centuries -2147483648 months';`,
				ErrorString: `interval field value out of range: "-0.1 centuries -2147483648 months"`,
			},
			{
				Statement:   `select interval '-0.1 decades -2147483648 months';`,
				ErrorString: `interval field value out of range: "-0.1 decades -2147483648 months"`,
			},
			{
				Statement:   `select interval '-0.1 yrs -2147483648 months';`,
				ErrorString: `interval field value out of range: "-0.1 yrs -2147483648 months"`,
			},
			{
				Statement:   `select interval '2147483647 months 0.1 millennium';`,
				ErrorString: `interval field value out of range: "2147483647 months 0.1 millennium"`,
			},
			{
				Statement:   `select interval '2147483647 months 0.1 centuries';`,
				ErrorString: `interval field value out of range: "2147483647 months 0.1 centuries"`,
			},
			{
				Statement:   `select interval '2147483647 months 0.1 decades';`,
				ErrorString: `interval field value out of range: "2147483647 months 0.1 decades"`,
			},
			{
				Statement:   `select interval '2147483647 months 0.1 yrs';`,
				ErrorString: `interval field value out of range: "2147483647 months 0.1 yrs"`,
			},
			{
				Statement:   `select interval '-2147483648 months -0.1 millennium';`,
				ErrorString: `interval field value out of range: "-2147483648 months -0.1 millennium"`,
			},
			{
				Statement:   `select interval '-2147483648 months -0.1 centuries';`,
				ErrorString: `interval field value out of range: "-2147483648 months -0.1 centuries"`,
			},
			{
				Statement:   `select interval '-2147483648 months -0.1 decades';`,
				ErrorString: `interval field value out of range: "-2147483648 months -0.1 decades"`,
			},
			{
				Statement:   `select interval '-2147483648 months -0.1 yrs';`,
				ErrorString: `interval field value out of range: "-2147483648 months -0.1 yrs"`,
			},
			{
				Statement:   `select interval '0.1 months 2147483647 days';`,
				ErrorString: `interval field value out of range: "0.1 months 2147483647 days"`,
			},
			{
				Statement:   `select interval '-0.1 months -2147483648 days';`,
				ErrorString: `interval field value out of range: "-0.1 months -2147483648 days"`,
			},
			{
				Statement:   `select interval '2147483647 days 0.1 months';`,
				ErrorString: `interval field value out of range: "2147483647 days 0.1 months"`,
			},
			{
				Statement:   `select interval '-2147483648 days -0.1 months';`,
				ErrorString: `interval field value out of range: "-2147483648 days -0.1 months"`,
			},
			{
				Statement:   `select interval '0.5 weeks 2147483647 days';`,
				ErrorString: `interval field value out of range: "0.5 weeks 2147483647 days"`,
			},
			{
				Statement:   `select interval '-0.5 weeks -2147483648 days';`,
				ErrorString: `interval field value out of range: "-0.5 weeks -2147483648 days"`,
			},
			{
				Statement:   `select interval '2147483647 days 0.5 weeks';`,
				ErrorString: `interval field value out of range: "2147483647 days 0.5 weeks"`,
			},
			{
				Statement:   `select interval '-2147483648 days -0.5 weeks';`,
				ErrorString: `interval field value out of range: "-2147483648 days -0.5 weeks"`,
			},
			{
				Statement:   `select interval '0.01 months 9223372036854775807 microseconds';`,
				ErrorString: `interval field value out of range: "0.01 months 9223372036854775807 microseconds"`,
			},
			{
				Statement:   `select interval '-0.01 months -9223372036854775808 microseconds';`,
				ErrorString: `interval field value out of range: "-0.01 months -9223372036854775808 microseconds"`,
			},
			{
				Statement:   `select interval '9223372036854775807 microseconds 0.01 months';`,
				ErrorString: `interval field value out of range: "9223372036854775807 microseconds 0.01 months"`,
			},
			{
				Statement:   `select interval '-9223372036854775808 microseconds -0.01 months';`,
				ErrorString: `interval field value out of range: "-9223372036854775808 microseconds -0.01 months"`,
			},
			{
				Statement:   `select interval '0.1 weeks 9223372036854775807 microseconds';`,
				ErrorString: `interval field value out of range: "0.1 weeks 9223372036854775807 microseconds"`,
			},
			{
				Statement:   `select interval '-0.1 weeks -9223372036854775808 microseconds';`,
				ErrorString: `interval field value out of range: "-0.1 weeks -9223372036854775808 microseconds"`,
			},
			{
				Statement:   `select interval '9223372036854775807 microseconds 0.1 weeks';`,
				ErrorString: `interval field value out of range: "9223372036854775807 microseconds 0.1 weeks"`,
			},
			{
				Statement:   `select interval '-9223372036854775808 microseconds -0.1 weeks';`,
				ErrorString: `interval field value out of range: "-9223372036854775808 microseconds -0.1 weeks"`,
			},
			{
				Statement:   `select interval '0.1 days 9223372036854775807 microseconds';`,
				ErrorString: `interval field value out of range: "0.1 days 9223372036854775807 microseconds"`,
			},
			{
				Statement:   `select interval '-0.1 days -9223372036854775808 microseconds';`,
				ErrorString: `interval field value out of range: "-0.1 days -9223372036854775808 microseconds"`,
			},
			{
				Statement:   `select interval '9223372036854775807 microseconds 0.1 days';`,
				ErrorString: `interval field value out of range: "9223372036854775807 microseconds 0.1 days"`,
			},
			{
				Statement:   `select interval '-9223372036854775808 microseconds -0.1 days';`,
				ErrorString: `interval field value out of range: "-9223372036854775808 microseconds -0.1 days"`,
			},
			{
				Statement:   `select interval 'P0.1Y2147483647M';`,
				ErrorString: `interval field value out of range: "P0.1Y2147483647M"`,
			},
			{
				Statement:   `select interval 'P-0.1Y-2147483648M';`,
				ErrorString: `interval field value out of range: "P-0.1Y-2147483648M"`,
			},
			{
				Statement:   `select interval 'P2147483647M0.1Y';`,
				ErrorString: `interval field value out of range: "P2147483647M0.1Y"`,
			},
			{
				Statement:   `select interval 'P-2147483648M-0.1Y';`,
				ErrorString: `interval field value out of range: "P-2147483648M-0.1Y"`,
			},
			{
				Statement:   `select interval 'P0.1M2147483647D';`,
				ErrorString: `interval field value out of range: "P0.1M2147483647D"`,
			},
			{
				Statement:   `select interval 'P-0.1M-2147483648D';`,
				ErrorString: `interval field value out of range: "P-0.1M-2147483648D"`,
			},
			{
				Statement:   `select interval 'P2147483647D0.1M';`,
				ErrorString: `interval field value out of range: "P2147483647D0.1M"`,
			},
			{
				Statement:   `select interval 'P-2147483648D-0.1M';`,
				ErrorString: `interval field value out of range: "P-2147483648D-0.1M"`,
			},
			{
				Statement:   `select interval 'P0.5W2147483647D';`,
				ErrorString: `interval field value out of range: "P0.5W2147483647D"`,
			},
			{
				Statement:   `select interval 'P-0.5W-2147483648D';`,
				ErrorString: `interval field value out of range: "P-0.5W-2147483648D"`,
			},
			{
				Statement:   `select interval 'P2147483647D0.5W';`,
				ErrorString: `interval field value out of range: "P2147483647D0.5W"`,
			},
			{
				Statement:   `select interval 'P-2147483648D-0.5W';`,
				ErrorString: `interval field value out of range: "P-2147483648D-0.5W"`,
			},
			{
				Statement:   `select interval 'P0.01MT2562047788H54.775807S';`,
				ErrorString: `interval field value out of range: "P0.01MT2562047788H54.775807S"`,
			},
			{
				Statement:   `select interval 'P-0.01MT-2562047788H-54.775808S';`,
				ErrorString: `interval field value out of range: "P-0.01MT-2562047788H-54.775808S"`,
			},
			{
				Statement:   `select interval 'P0.1DT2562047788H54.775807S';`,
				ErrorString: `interval field value out of range: "P0.1DT2562047788H54.775807S"`,
			},
			{
				Statement:   `select interval 'P-0.1DT-2562047788H-54.775808S';`,
				ErrorString: `interval field value out of range: "P-0.1DT-2562047788H-54.775808S"`,
			},
			{
				Statement:   `select interval 'PT2562047788.1H54.775807S';`,
				ErrorString: `interval field value out of range: "PT2562047788.1H54.775807S"`,
			},
			{
				Statement:   `select interval 'PT-2562047788.1H-54.775808S';`,
				ErrorString: `interval field value out of range: "PT-2562047788.1H-54.775808S"`,
			},
			{
				Statement:   `select interval 'PT2562047788H0.1M54.775807S';`,
				ErrorString: `interval field value out of range: "PT2562047788H0.1M54.775807S"`,
			},
			{
				Statement:   `select interval 'PT-2562047788H-0.1M-54.775808S';`,
				ErrorString: `interval field value out of range: "PT-2562047788H-0.1M-54.775808S"`,
			},
			{
				Statement:   `select interval 'P0.1-2147483647-00';`,
				ErrorString: `interval field value out of range: "P0.1-2147483647-00"`,
			},
			{
				Statement:   `select interval 'P00-0.1-2147483647';`,
				ErrorString: `interval field value out of range: "P00-0.1-2147483647"`,
			},
			{
				Statement:   `select interval 'P00-0.01-00T2562047788:00:54.775807';`,
				ErrorString: `interval field value out of range: "P00-0.01-00T2562047788:00:54.775807"`,
			},
			{
				Statement:   `select interval 'P00-00-0.1T2562047788:00:54.775807';`,
				ErrorString: `interval field value out of range: "P00-00-0.1T2562047788:00:54.775807"`,
			},
			{
				Statement:   `select interval 'PT2562047788.1:00:54.775807';`,
				ErrorString: `interval field value out of range: "PT2562047788.1:00:54.775807"`,
			},
			{
				Statement:   `select interval 'PT2562047788:01.:54.775807';`,
				ErrorString: `interval field value out of range: "PT2562047788:01.:54.775807"`,
			},
			{
				Statement:   `select interval '0.1 2562047788:0:54.775807';`,
				ErrorString: `interval field value out of range: "0.1 2562047788:0:54.775807"`,
			},
			{
				Statement:   `select interval '0.1 2562047788:0:54.775808 ago';`,
				ErrorString: `interval field value out of range: "0.1 2562047788:0:54.775808 ago"`,
			},
			{
				Statement:   `select interval '2562047788.1:0:54.775807';`,
				ErrorString: `interval field value out of range: "2562047788.1:0:54.775807"`,
			},
			{
				Statement:   `select interval '2562047788.1:0:54.775808 ago';`,
				ErrorString: `interval field value out of range: "2562047788.1:0:54.775808 ago"`,
			},
			{
				Statement:   `select interval '2562047788:0.1:54.775807';`,
				ErrorString: `invalid input syntax for type interval: "2562047788:0.1:54.775807"`,
			},
			{
				Statement:   `select interval '2562047788:0.1:54.775808 ago';`,
				ErrorString: `invalid input syntax for type interval: "2562047788:0.1:54.775808 ago"`,
			},
			{
				Statement:   `select interval '-2147483648 months ago';`,
				ErrorString: `interval field value out of range: "-2147483648 months ago"`,
			},
			{
				Statement:   `select interval '-2147483648 days ago';`,
				ErrorString: `interval field value out of range: "-2147483648 days ago"`,
			},
			{
				Statement:   `select interval '-9223372036854775808 microseconds ago';`,
				ErrorString: `interval field value out of range: "-9223372036854775808 microseconds ago"`,
			},
			{
				Statement:   `select interval '-2147483648 months -2147483648 days -9223372036854775808 microseconds ago';`,
				ErrorString: `interval field value out of range: "-2147483648 months -2147483648 days -9223372036854775808 microseconds ago"`,
			},
			{
				Statement: `SET IntervalStyle to postgres;`,
			},
			{
				Statement: `select interval '-2147483648 months -2147483648 days -9223372036854775808 us';`,
				Results:   []sql.Row{{`-178956970 years -8 mons -2147483648 days -2562047788:00:54.775808`}},
			},
			{
				Statement: `SET IntervalStyle to sql_standard;`,
			},
			{
				Statement: `select interval '-2147483648 months -2147483648 days -9223372036854775808 us';`,
				Results:   []sql.Row{{`-178956970-8 -2147483648 -2562047788:00:54.775808`}},
			},
			{
				Statement: `SET IntervalStyle to iso_8601;`,
			},
			{
				Statement: `select interval '-2147483648 months -2147483648 days -9223372036854775808 us';`,
				Results:   []sql.Row{{`P-178956970Y-8M-2147483648DT-2562047788H-54.775808S`}},
			},
			{
				Statement: `SET IntervalStyle to postgres_verbose;`,
			},
			{
				Statement: `select interval '-2147483648 months -2147483648 days -9223372036854775808 us';`,
				Results:   []sql.Row{{`@ 178956970 years 8 mons 2147483648 days 2562047788 hours 54.775808 secs ago`}},
			},
			{
				Statement: `select '30 days'::interval = '1 month'::interval as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select interval_hash('30 days'::interval) = interval_hash('1 month'::interval) as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select make_interval(years := 2);`,
				Results:   []sql.Row{{`@ 2 years`}},
			},
			{
				Statement: `select make_interval(years := 1, months := 6);`,
				Results:   []sql.Row{{`@ 1 year 6 mons`}},
			},
			{
				Statement: `select make_interval(years := 1, months := -1, weeks := 5, days := -7, hours := 25, mins := -180);`,
				Results:   []sql.Row{{`@ 11 mons 28 days 22 hours`}},
			},
			{
				Statement: `select make_interval() = make_interval(years := 0, months := 0, weeks := 0, days := 0, mins := 0, secs := 0.0);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select make_interval(hours := -2, mins := -10, secs := -25.3);`,
				Results:   []sql.Row{{`@ 2 hours 10 mins 25.3 secs ago`}},
			},
			{
				Statement:   `select make_interval(years := 'inf'::float::int);`,
				ErrorString: `integer out of range`,
			},
			{
				Statement:   `select make_interval(months := 'NaN'::float::int);`,
				ErrorString: `integer out of range`,
			},
			{
				Statement:   `select make_interval(secs := 'inf');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement:   `select make_interval(secs := 'NaN');`,
				ErrorString: `interval out of range`,
			},
			{
				Statement: `select make_interval(secs := 7e12);`,
				Results:   []sql.Row{{`@ 1944444444 hours 26 mins 40 secs`}},
			},
			{
				Statement: `SELECT f1,
    EXTRACT(MICROSECOND FROM f1) AS MICROSECOND,
    EXTRACT(MILLISECOND FROM f1) AS MILLISECOND,
    EXTRACT(SECOND FROM f1) AS SECOND,
    EXTRACT(MINUTE FROM f1) AS MINUTE,
    EXTRACT(HOUR FROM f1) AS HOUR,
    EXTRACT(DAY FROM f1) AS DAY,
    EXTRACT(MONTH FROM f1) AS MONTH,
    EXTRACT(QUARTER FROM f1) AS QUARTER,
    EXTRACT(YEAR FROM f1) AS YEAR,
    EXTRACT(DECADE FROM f1) AS DECADE,
    EXTRACT(CENTURY FROM f1) AS CENTURY,
    EXTRACT(MILLENNIUM FROM f1) AS MILLENNIUM,
    EXTRACT(EPOCH FROM f1) AS EPOCH
    FROM INTERVAL_TBL;`,
				Results: []sql.Row{{`@ 1 min`, 0, 0.000, 0.000000, 1, 0, 0, 0, 1, 0, 0, 0, 0, 60.000000}, {`@ 5 hours`, 0, 0.000, 0.000000, 0, 5, 0, 0, 1, 0, 0, 0, 0, 18000.000000}, {`@ 10 days`, 0, 0.000, 0.000000, 0, 0, 10, 0, 1, 0, 0, 0, 0, 864000.000000}, {`@ 34 years`, 0, 0.000, 0.000000, 0, 0, 0, 0, 1, 34, 3, 0, 0, 1072958400.000000}, {`@ 3 mons`, 0, 0.000, 0.000000, 0, 0, 0, 3, 2, 0, 0, 0, 0, 7776000.000000}, {`@ 14 secs ago`, -14000000, -14000.000, -14.000000, 0, 0, 0, 0, 1, 0, 0, 0, 0, -14.000000}, {`@ 1 day 2 hours 3 mins 4 secs`, 4000000, 4000.000, 4.000000, 3, 2, 1, 0, 1, 0, 0, 0, 0, 93784.000000}, {`@ 6 years`, 0, 0.000, 0.000000, 0, 0, 0, 0, 1, 6, 0, 0, 0, 189345600.000000}, {`@ 5 mons`, 0, 0.000, 0.000000, 0, 0, 0, 5, 2, 0, 0, 0, 0, 12960000.000000}, {`@ 5 mons 12 hours`, 0, 0.000, 0.000000, 0, 12, 0, 5, 2, 0, 0, 0, 0, 13003200.000000}},
			},
			{
				Statement:   `SELECT EXTRACT(FORTNIGHT FROM INTERVAL '2 days');  -- error`,
				ErrorString: `unit "fortnight" not recognized for type interval`,
			},
			{
				Statement:   `SELECT EXTRACT(TIMEZONE FROM INTERVAL '2 days');  -- error`,
				ErrorString: `unit "timezone" not supported for type interval`,
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM INTERVAL '100 y');`,
				Results:   []sql.Row{{10}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM INTERVAL '99 y');`,
				Results:   []sql.Row{{9}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM INTERVAL '-99 y');`,
				Results:   []sql.Row{{-9}},
			},
			{
				Statement: `SELECT EXTRACT(DECADE FROM INTERVAL '-100 y');`,
				Results:   []sql.Row{{-10}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM INTERVAL '100 y');`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM INTERVAL '99 y');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM INTERVAL '-99 y');`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT EXTRACT(CENTURY FROM INTERVAL '-100 y');`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `SELECT f1,
    date_part('microsecond', f1) AS microsecond,
    date_part('millisecond', f1) AS millisecond,
    date_part('second', f1) AS second,
    date_part('epoch', f1) AS epoch
    FROM INTERVAL_TBL;`,
				Results: []sql.Row{{`@ 1 min`, 0, 0, 0, 60}, {`@ 5 hours`, 0, 0, 0, 18000}, {`@ 10 days`, 0, 0, 0, 864000}, {`@ 34 years`, 0, 0, 0, 1072958400}, {`@ 3 mons`, 0, 0, 0, 7776000}, {`@ 14 secs ago`, -14000000, -14000, -14, -14}, {`@ 1 day 2 hours 3 mins 4 secs`, 4000000, 4000, 4, 93784}, {`@ 6 years`, 0, 0, 0, 189345600}, {`@ 5 mons`, 0, 0, 0, 12960000}, {`@ 5 mons 12 hours`, 0, 0, 0, 13003200}},
			},
			{
				Statement: `SELECT extract(epoch from interval '1000000000 days');`,
				Results:   []sql.Row{{86400000000000.000000}},
			},
		},
	})
}
