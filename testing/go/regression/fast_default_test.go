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

func TestFastDefault(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_fast_default)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_fast_default,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET search_path = fast_default;`,
			},
			{
				Statement: `CREATE SCHEMA fast_default;`,
			},
			{
				Statement: `CREATE TABLE m(id OID);`,
			},
			{
				Statement: `INSERT INTO m VALUES (NULL::OID);`,
			},
			{
				Statement: `CREATE FUNCTION set(tabname name) RETURNS VOID
AS $$
BEGIN
  UPDATE m
  SET id = (SELECT c.relfilenode
            FROM pg_class AS c, pg_namespace AS s
            WHERE c.relname = tabname
                AND c.relnamespace = s.oid
                AND s.nspname = 'fast_default');`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE 'plpgsql';`,
			},
			{
				Statement: `CREATE FUNCTION comp() RETURNS TEXT
AS $$
BEGIN
  RETURN (SELECT CASE
               WHEN m.id = c.relfilenode THEN 'Unchanged'
               ELSE 'Rewritten'
               END
           FROM m, pg_class AS c, pg_namespace AS s
           WHERE c.relname = 't'
               AND c.relnamespace = s.oid
               AND s.nspname = 'fast_default');`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE 'plpgsql';`,
			},
			{
				Statement: `CREATE FUNCTION log_rewrite() RETURNS event_trigger
LANGUAGE plpgsql as
$func$
declare
   this_schema text;`,
			},
			{
				Statement: `begin
    select into this_schema relnamespace::regnamespace::text
    from pg_class
    where oid = pg_event_trigger_table_rewrite_oid();`,
			},
			{
				Statement: `    if this_schema = 'fast_default'
    then
        RAISE NOTICE 'rewriting table % for reason %',
          pg_event_trigger_table_rewrite_oid()::regclass,
          pg_event_trigger_table_rewrite_reason();`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$func$;`,
			},
			{
				Statement: `CREATE TABLE has_volatile AS
SELECT * FROM generate_series(1,10) id;`,
			},
			{
				Statement: `CREATE EVENT TRIGGER has_volatile_rewrite
                  ON table_rewrite
   EXECUTE PROCEDURE log_rewrite();`,
			},
			{
				Statement: `ALTER TABLE has_volatile ADD col1 int;`,
			},
			{
				Statement: `ALTER TABLE has_volatile ADD col2 int DEFAULT 1;`,
			},
			{
				Statement: `ALTER TABLE has_volatile ADD col3 timestamptz DEFAULT current_timestamp;`,
			},
			{
				Statement: `ALTER TABLE has_volatile ADD col4 int DEFAULT (random() * 10000)::int;`,
			},
			{
				Statement: `CREATE TABLE T(pk INT NOT NULL PRIMARY KEY, c_int INT DEFAULT 1);`,
			},
			{
				Statement: `SELECT set('t');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `INSERT INTO T VALUES (1), (2);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_bpchar BPCHAR(5) DEFAULT 'hello',
              ALTER COLUMN c_int SET DEFAULT 2;`,
			},
			{
				Statement: `INSERT INTO T VALUES (3), (4);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_text TEXT  DEFAULT 'world',
              ALTER COLUMN c_bpchar SET DEFAULT 'dog';`,
			},
			{
				Statement: `INSERT INTO T VALUES (5), (6);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_date DATE DEFAULT '2016-06-02',
              ALTER COLUMN c_text SET DEFAULT 'cat';`,
			},
			{
				Statement: `INSERT INTO T VALUES (7), (8);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_timestamp TIMESTAMP DEFAULT '2016-09-01 12:00:00',
              ADD COLUMN c_timestamp_null TIMESTAMP,
              ALTER COLUMN c_date SET DEFAULT '2010-01-01';`,
			},
			{
				Statement: `INSERT INTO T VALUES (9), (10);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_array TEXT[]
                  DEFAULT '{"This", "is", "the", "real", "world"}',
              ALTER COLUMN c_timestamp SET DEFAULT '1970-12-31 11:12:13',
              ALTER COLUMN c_timestamp_null SET DEFAULT '2016-09-29 12:00:00';`,
			},
			{
				Statement: `INSERT INTO T VALUES (11), (12);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_small SMALLINT DEFAULT -5,
              ADD COLUMN c_small_null SMALLINT,
              ALTER COLUMN c_array
                  SET DEFAULT '{"This", "is", "no", "fantasy"}';`,
			},
			{
				Statement: `INSERT INTO T VALUES (13), (14);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_big BIGINT DEFAULT 180000000000018,
              ALTER COLUMN c_small SET DEFAULT 9,
              ALTER COLUMN c_small_null SET DEFAULT 13;`,
			},
			{
				Statement: `INSERT INTO T VALUES (15), (16);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_num NUMERIC DEFAULT 1.00000000001,
              ALTER COLUMN c_big SET DEFAULT -9999999999999999;`,
			},
			{
				Statement: `INSERT INTO T VALUES (17), (18);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_time TIME DEFAULT '12:00:00',
              ALTER COLUMN c_num SET DEFAULT 2.000000000000002;`,
			},
			{
				Statement: `INSERT INTO T VALUES (19), (20);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_interval INTERVAL DEFAULT '1 day',
              ALTER COLUMN c_time SET DEFAULT '23:59:59';`,
			},
			{
				Statement: `INSERT INTO T VALUES (21), (22);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_hugetext TEXT DEFAULT repeat('abcdefg',1000),
              ALTER COLUMN c_interval SET DEFAULT '3 hours';`,
			},
			{
				Statement: `INSERT INTO T VALUES (23), (24);`,
			},
			{
				Statement: `ALTER TABLE T ALTER COLUMN c_interval DROP DEFAULT,
              ALTER COLUMN c_hugetext SET DEFAULT repeat('poiuyt', 1000);`,
			},
			{
				Statement: `INSERT INTO T VALUES (25), (26);`,
			},
			{
				Statement: `ALTER TABLE T ALTER COLUMN c_bpchar    DROP DEFAULT,
              ALTER COLUMN c_date      DROP DEFAULT,
              ALTER COLUMN c_text      DROP DEFAULT,
              ALTER COLUMN c_timestamp DROP DEFAULT,
              ALTER COLUMN c_array     DROP DEFAULT,
              ALTER COLUMN c_small     DROP DEFAULT,
              ALTER COLUMN c_big       DROP DEFAULT,
              ALTER COLUMN c_num       DROP DEFAULT,
              ALTER COLUMN c_time      DROP DEFAULT,
              ALTER COLUMN c_hugetext  DROP DEFAULT;`,
			},
			{
				Statement: `INSERT INTO T VALUES (27), (28);`,
			},
			{
				Statement: `SELECT pk, c_int, c_bpchar, c_text, c_date, c_timestamp,
       c_timestamp_null, c_array, c_small, c_small_null,
       c_big, c_num, c_time, c_interval,
       c_hugetext = repeat('abcdefg',1000) as c_hugetext_origdef,
       c_hugetext = repeat('poiuyt', 1000) as c_hugetext_newdef
FROM T ORDER BY pk;`,
				Results: []sql.Row{{1, 1, `hello`, `world`, `06-02-2016`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {2, 1, `hello`, `world`, `06-02-2016`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {3, 2, `hello`, `world`, `06-02-2016`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {4, 2, `hello`, `world`, `06-02-2016`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {5, 2, `dog`, `world`, `06-02-2016`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {6, 2, `dog`, `world`, `06-02-2016`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {7, 2, `dog`, `cat`, `06-02-2016`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {8, 2, `dog`, `cat`, `06-02-2016`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {9, 2, `dog`, `cat`, `01-01-2010`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {10, 2, `dog`, `cat`, `01-01-2010`, `Thu Sep 01 12:00:00 2016`, ``, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {11, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {12, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,the,real,world}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {13, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {14, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, -5, ``, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {15, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {16, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, 180000000000018, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {17, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {18, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 1.00000000001, `12:00:00`, `@ 1 day`, true, false}, {19, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 2.000000000000002, `12:00:00`, `@ 1 day`, true, false}, {20, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 2.000000000000002, `12:00:00`, `@ 1 day`, true, false}, {21, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 2.000000000000002, `23:59:59`, `@ 1 day`, true, false}, {22, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 2.000000000000002, `23:59:59`, `@ 1 day`, true, false}, {23, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 2.000000000000002, `23:59:59`, `@ 3 hours`, true, false}, {24, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 2.000000000000002, `23:59:59`, `@ 3 hours`, true, false}, {25, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 2.000000000000002, `23:59:59`, ``, false, true}, {26, 2, `dog`, `cat`, `01-01-2010`, `Thu Dec 31 11:12:13 1970`, `Thu Sep 29 12:00:00 2016`, `{This,is,no,fantasy}`, 9, 13, -9999999999999999, 2.000000000000002, `23:59:59`, ``, false, true}, {27, 2, ``, ``, ``, ``, `Thu Sep 29 12:00:00 2016`, ``, ``, 13, ``, ``, ``, ``, ``, ``}, {28, 2, ``, ``, ``, ``, `Thu Sep 29 12:00:00 2016`, ``, ``, 13, ``, ``, ``, ``, ``, ``}},
			},
			{
				Statement: `SELECT comp();`,
				Results:   []sql.Row{{`Unchanged`}},
			},
			{
				Statement: `DROP TABLE T;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION foo(a INT) RETURNS TEXT AS $$
DECLARE res TEXT := '';`,
			},
			{
				Statement: `        i INT;`,
			},
			{
				Statement: `BEGIN
  i := 0;`,
			},
			{
				Statement: `  WHILE (i < a) LOOP
    res := res || chr(ascii('a') + i);`,
			},
			{
				Statement: `    i := i + 1;`,
			},
			{
				Statement: `  END LOOP;`,
			},
			{
				Statement: `  RETURN res;`,
			},
			{
				Statement: `END; $$ LANGUAGE PLPGSQL STABLE;`,
			},
			{
				Statement: `CREATE TABLE T(pk INT NOT NULL PRIMARY KEY, c_int INT DEFAULT LENGTH(foo(6)));`,
			},
			{
				Statement: `SELECT set('t');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `INSERT INTO T VALUES (1), (2);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_bpchar BPCHAR(5) DEFAULT foo(4),
              ALTER COLUMN c_int SET DEFAULT LENGTH(foo(8));`,
			},
			{
				Statement: `INSERT INTO T VALUES (3), (4);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_text TEXT  DEFAULT foo(6),
              ALTER COLUMN c_bpchar SET DEFAULT foo(3);`,
			},
			{
				Statement: `INSERT INTO T VALUES (5), (6);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_date DATE
                  DEFAULT '2016-06-02'::DATE  + LENGTH(foo(10)),
              ALTER COLUMN c_text SET DEFAULT foo(12);`,
			},
			{
				Statement: `INSERT INTO T VALUES (7), (8);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_timestamp TIMESTAMP
                  DEFAULT '2016-09-01'::DATE + LENGTH(foo(10)),
              ALTER COLUMN c_date
                  SET DEFAULT '2010-01-01'::DATE - LENGTH(foo(4));`,
			},
			{
				Statement: `INSERT INTO T VALUES (9), (10);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_array TEXT[]
                  DEFAULT ('{"This", "is", "' || foo(4) ||
                           '","the", "real", "world"}')::TEXT[],
              ALTER COLUMN c_timestamp
                  SET DEFAULT '1970-12-31'::DATE + LENGTH(foo(30));`,
			},
			{
				Statement: `INSERT INTO T VALUES (11), (12);`,
			},
			{
				Statement: `ALTER TABLE T ALTER COLUMN c_int DROP DEFAULT,
              ALTER COLUMN c_array
                  SET DEFAULT ('{"This", "is", "' || foo(1) ||
                               '", "fantasy"}')::text[];`,
			},
			{
				Statement: `INSERT INTO T VALUES (13), (14);`,
			},
			{
				Statement: `ALTER TABLE T ALTER COLUMN c_bpchar    DROP DEFAULT,
              ALTER COLUMN c_date      DROP DEFAULT,
              ALTER COLUMN c_text      DROP DEFAULT,
              ALTER COLUMN c_timestamp DROP DEFAULT,
              ALTER COLUMN c_array     DROP DEFAULT;`,
			},
			{
				Statement: `INSERT INTO T VALUES (15), (16);`,
			},
			{
				Statement: `SELECT * FROM T;`,
				Results:   []sql.Row{{1, 6, `abcd`, `abcdef`, `06-12-2016`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {2, 6, `abcd`, `abcdef`, `06-12-2016`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {3, 8, `abcd`, `abcdef`, `06-12-2016`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {4, 8, `abcd`, `abcdef`, `06-12-2016`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {5, 8, `abc`, `abcdef`, `06-12-2016`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {6, 8, `abc`, `abcdef`, `06-12-2016`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {7, 8, `abc`, `abcdefghijkl`, `06-12-2016`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {8, 8, `abc`, `abcdefghijkl`, `06-12-2016`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {9, 8, `abc`, `abcdefghijkl`, `12-28-2009`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {10, 8, `abc`, `abcdefghijkl`, `12-28-2009`, `Sun Sep 11 00:00:00 2016`, `{This,is,abcd,the,real,world}`}, {11, 8, `abc`, `abcdefghijkl`, `12-28-2009`, `Sat Jan 30 00:00:00 1971`, `{This,is,abcd,the,real,world}`}, {12, 8, `abc`, `abcdefghijkl`, `12-28-2009`, `Sat Jan 30 00:00:00 1971`, `{This,is,abcd,the,real,world}`}, {13, ``, `abc`, `abcdefghijkl`, `12-28-2009`, `Sat Jan 30 00:00:00 1971`, `{This,is,a,fantasy}`}, {14, ``, `abc`, `abcdefghijkl`, `12-28-2009`, `Sat Jan 30 00:00:00 1971`, `{This,is,a,fantasy}`}, {15, ``, ``, ``, ``, ``, ``}, {16, ``, ``, ``, ``, ``, ``}},
			},
			{
				Statement: `SELECT comp();`,
				Results:   []sql.Row{{`Unchanged`}},
			},
			{
				Statement: `DROP TABLE T;`,
			},
			{
				Statement: `DROP FUNCTION foo(INT);`,
			},
			{
				Statement: `CREATE TABLE T(pk INT NOT NULL PRIMARY KEY);`,
			},
			{
				Statement: `INSERT INTO T VALUES (1);`,
			},
			{
				Statement: `SELECT set('t');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c1 TIMESTAMP DEFAULT now();`,
			},
			{
				Statement: `SELECT comp();`,
				Results:   []sql.Row{{`Unchanged`}},
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c2 TIMESTAMP DEFAULT clock_timestamp();`,
			},
			{
				Statement: `SELECT comp();`,
				Results:   []sql.Row{{`Rewritten`}},
			},
			{
				Statement: `DROP TABLE T;`,
			},
			{
				Statement: `CREATE TABLE T (pk INT NOT NULL PRIMARY KEY);`,
			},
			{
				Statement: `SELECT set('t');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `INSERT INTO T SELECT * FROM generate_series(1, 10) a;`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_bigint BIGINT NOT NULL DEFAULT -1;`,
			},
			{
				Statement: `INSERT INTO T SELECT b, b - 10 FROM generate_series(11, 20) a(b);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_text TEXT DEFAULT 'hello';`,
			},
			{
				Statement: `INSERT INTO T SELECT b, b - 10, (b + 10)::text FROM generate_series(21, 30) a(b);`,
			},
			{
				Statement: `SELECT c_bigint, c_text FROM T WHERE c_bigint = -1 LIMIT 1;`,
				Results:   []sql.Row{{-1, `hello`}},
			},
			{
				Statement: `EXPLAIN (VERBOSE TRUE, COSTS FALSE)
SELECT c_bigint, c_text FROM T WHERE c_bigint = -1 LIMIT 1;`,
				Results: []sql.Row{{`Limit`}, {`Output: c_bigint, c_text`}, {`->  Seq Scan on fast_default.t`}, {`Output: c_bigint, c_text`}, {`Filter: (t.c_bigint = '-1'::integer)`}},
			},
			{
				Statement: `SELECT c_bigint, c_text FROM T WHERE c_text = 'hello' LIMIT 1;`,
				Results:   []sql.Row{{-1, `hello`}},
			},
			{
				Statement: `EXPLAIN (VERBOSE TRUE, COSTS FALSE) SELECT c_bigint, c_text FROM T WHERE c_text = 'hello' LIMIT 1;`,
				Results:   []sql.Row{{`Limit`}, {`Output: c_bigint, c_text`}, {`->  Seq Scan on fast_default.t`}, {`Output: c_bigint, c_text`}, {`Filter: (t.c_text = 'hello'::text)`}},
			},
			{
				Statement: `SELECT COALESCE(c_bigint, pk), COALESCE(c_text, pk::text)
FROM T
ORDER BY pk LIMIT 10;`,
				Results: []sql.Row{{-1, `hello`}, {-1, `hello`}, {-1, `hello`}, {-1, `hello`}, {-1, `hello`}, {-1, `hello`}, {-1, `hello`}, {-1, `hello`}, {-1, `hello`}, {-1, `hello`}},
			},
			{
				Statement: `SELECT SUM(c_bigint), MAX(c_text COLLATE "C" ), MIN(c_text COLLATE "C") FROM T;`,
				Results:   []sql.Row{{200, `hello`, 31}},
			},
			{
				Statement: `SELECT * FROM T ORDER BY c_bigint, c_text, pk LIMIT 10;`,
				Results:   []sql.Row{{1, -1, `hello`}, {2, -1, `hello`}, {3, -1, `hello`}, {4, -1, `hello`}, {5, -1, `hello`}, {6, -1, `hello`}, {7, -1, `hello`}, {8, -1, `hello`}, {9, -1, `hello`}, {10, -1, `hello`}},
			},
			{
				Statement: `EXPLAIN (VERBOSE TRUE, COSTS FALSE)
SELECT * FROM T ORDER BY c_bigint, c_text, pk LIMIT 10;`,
				Results: []sql.Row{{`Limit`}, {`Output: pk, c_bigint, c_text`}, {`->  Sort`}, {`Output: pk, c_bigint, c_text`}, {`Sort Key: t.c_bigint, t.c_text, t.pk`}, {`->  Seq Scan on fast_default.t`}, {`Output: pk, c_bigint, c_text`}},
			},
			{
				Statement: `SELECT * FROM T WHERE c_bigint > -1 ORDER BY c_bigint, c_text, pk LIMIT 10;`,
				Results:   []sql.Row{{11, 1, `hello`}, {12, 2, `hello`}, {13, 3, `hello`}, {14, 4, `hello`}, {15, 5, `hello`}, {16, 6, `hello`}, {17, 7, `hello`}, {18, 8, `hello`}, {19, 9, `hello`}, {20, 10, `hello`}},
			},
			{
				Statement: `EXPLAIN (VERBOSE TRUE, COSTS FALSE)
SELECT * FROM T WHERE c_bigint > -1 ORDER BY c_bigint, c_text, pk LIMIT 10;`,
				Results: []sql.Row{{`Limit`}, {`Output: pk, c_bigint, c_text`}, {`->  Sort`}, {`Output: pk, c_bigint, c_text`}, {`Sort Key: t.c_bigint, t.c_text, t.pk`}, {`->  Seq Scan on fast_default.t`}, {`Output: pk, c_bigint, c_text`}, {`Filter: (t.c_bigint > '-1'::integer)`}},
			},
			{
				Statement: `DELETE FROM T WHERE pk BETWEEN 10 AND 20 RETURNING *;`,
				Results:   []sql.Row{{10, -1, `hello`}, {11, 1, `hello`}, {12, 2, `hello`}, {13, 3, `hello`}, {14, 4, `hello`}, {15, 5, `hello`}, {16, 6, `hello`}, {17, 7, `hello`}, {18, 8, `hello`}, {19, 9, `hello`}, {20, 10, `hello`}},
			},
			{
				Statement: `EXPLAIN (VERBOSE TRUE, COSTS FALSE)
DELETE FROM T WHERE pk BETWEEN 10 AND 20 RETURNING *;`,
				Results: []sql.Row{{`Delete on fast_default.t`}, {`Output: pk, c_bigint, c_text`}, {`->  Bitmap Heap Scan on fast_default.t`}, {`Output: ctid`}, {`Recheck Cond: ((t.pk >= 10) AND (t.pk <= 20))`}, {`->  Bitmap Index Scan on t_pkey`}, {`Index Cond: ((t.pk >= 10) AND (t.pk <= 20))`}},
			},
			{
				Statement: `UPDATE T SET c_text = '"' || c_text || '"'  WHERE pk < 10;`,
			},
			{
				Statement: `SELECT * FROM T WHERE c_text LIKE '"%"' ORDER BY PK;`,
				Results:   []sql.Row{{1, -1, "hello"}, {2, -1, "hello"}, {3, -1, "hello"}, {4, -1, "hello"}, {5, -1, "hello"}, {6, -1, "hello"}, {7, -1, "hello"}, {8, -1, "hello"}, {9, -1, "hello"}},
			},
			{
				Statement: `SELECT comp();`,
				Results:   []sql.Row{{`Unchanged`}},
			},
			{
				Statement: `DROP TABLE T;`,
			},
			{
				Statement: `CREATE TABLE T(pk INT NOT NULL PRIMARY KEY);`,
			},
			{
				Statement: `SELECT set('t');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `INSERT INTO T VALUES (1), (2);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_int INT NOT NULL DEFAULT -1;`,
			},
			{
				Statement: `INSERT INTO T VALUES (3), (4);`,
			},
			{
				Statement: `ALTER TABLE T ADD COLUMN c_text TEXT DEFAULT 'Hello';`,
			},
			{
				Statement: `INSERT INTO T VALUES (5), (6);`,
			},
			{
				Statement: `ALTER TABLE T ALTER COLUMN c_text SET DEFAULT 'world',
              ALTER COLUMN c_int  SET DEFAULT 1;`,
			},
			{
				Statement: `INSERT INTO T VALUES (7), (8);`,
			},
			{
				Statement: `SELECT * FROM T ORDER BY pk;`,
				Results:   []sql.Row{{1, -1, `Hello`}, {2, -1, `Hello`}, {3, -1, `Hello`}, {4, -1, `Hello`}, {5, -1, `Hello`}, {6, -1, `Hello`}, {7, 1, `world`}, {8, 1, `world`}},
			},
			{
				Statement: `CREATE INDEX i ON T(c_int, c_text);`,
			},
			{
				Statement: `SELECT c_text FROM T WHERE c_int = -1;`,
				Results:   []sql.Row{{`Hello`}, {`Hello`}, {`Hello`}, {`Hello`}, {`Hello`}, {`Hello`}},
			},
			{
				Statement: `SELECT comp();`,
				Results:   []sql.Row{{`Unchanged`}},
			},
			{
				Statement: `CREATE TABLE t1 AS
SELECT 1::int AS a , 2::int AS b
FROM generate_series(1,20) q;`,
			},
			{
				Statement: `ALTER TABLE t1 ADD COLUMN c text;`,
			},
			{
				Statement: `SELECT a,
       stddev(cast((SELECT sum(1) FROM generate_series(1,20) x) AS float4))
          OVER (PARTITION BY a,b,c ORDER BY b)
       AS z
FROM t1;`,
				Results: []sql.Row{{1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}, {1, 0}},
			},
			{
				Statement: `DROP TABLE T;`,
			},
			{
				Statement: `CREATE FUNCTION test_trigger()
RETURNS trigger
LANGUAGE plpgsql
AS $$
begin
    raise notice 'old tuple: %', to_json(OLD)::text;`,
			},
			{
				Statement: `    if TG_OP = 'DELETE'
    then
       return OLD;`,
			},
			{
				Statement: `    else
       return NEW;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TABLE t (id serial PRIMARY KEY, a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO t (a,b,c) VALUES (1,2,3);`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN x int NOT NULL DEFAULT 4;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN y int NOT NULL DEFAULT 5;`,
			},
			{
				Statement: `CREATE TRIGGER a BEFORE UPDATE ON t FOR EACH ROW EXECUTE PROCEDURE test_trigger();`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, 3, 4, 5}},
			},
			{
				Statement: `UPDATE t SET y = 2;`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, 3, 4, 2}},
			},
			{
				Statement: `DROP TABLE t;`,
			},
			{
				Statement: `CREATE TABLE t (id serial PRIMARY KEY, a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO t (a,b,c) VALUES (1,2,3);`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN x int NOT NULL DEFAULT 4;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN y int;`,
			},
			{
				Statement: `CREATE TRIGGER a BEFORE UPDATE ON t FOR EACH ROW EXECUTE PROCEDURE test_trigger();`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, 3, 4, ``}},
			},
			{
				Statement: `UPDATE t SET y = 2;`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, 3, 4, 2}},
			},
			{
				Statement: `DROP TABLE t;`,
			},
			{
				Statement: `CREATE TABLE t (id serial PRIMARY KEY, a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO t (a,b,c) VALUES (1,2,3);`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN x int;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN y int NOT NULL DEFAULT 5;`,
			},
			{
				Statement: `CREATE TRIGGER a BEFORE UPDATE ON t FOR EACH ROW EXECUTE PROCEDURE test_trigger();`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, 3, ``, 5}},
			},
			{
				Statement: `UPDATE t SET y = 2;`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, 3, ``, 2}},
			},
			{
				Statement: `DROP TABLE t;`,
			},
			{
				Statement: `CREATE TABLE t (id serial PRIMARY KEY, a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO t (a,b,c) VALUES (1,2,3);`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN x int;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN y int;`,
			},
			{
				Statement: `CREATE TRIGGER a BEFORE UPDATE ON t FOR EACH ROW EXECUTE PROCEDURE test_trigger();`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, 3, ``, ``}},
			},
			{
				Statement: `UPDATE t SET y = 2;`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, 3, ``, 2}},
			},
			{
				Statement: `DROP TABLE t;`,
			},
			{
				Statement: `CREATE TABLE t (id serial PRIMARY KEY, a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO t (a,b,c) VALUES (1,2,NULL);`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN x int NOT NULL DEFAULT 4;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN y int NOT NULL DEFAULT 5;`,
			},
			{
				Statement: `CREATE TRIGGER a BEFORE UPDATE ON t FOR EACH ROW EXECUTE PROCEDURE test_trigger();`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, ``, 4, 5}},
			},
			{
				Statement: `UPDATE t SET y = 2;`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, ``, 4, 2}},
			},
			{
				Statement: `DROP TABLE t;`,
			},
			{
				Statement: `CREATE TABLE t (id serial PRIMARY KEY, a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO t (a,b,c) VALUES (1,2,NULL);`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN x int NOT NULL DEFAULT 4;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN y int;`,
			},
			{
				Statement: `CREATE TRIGGER a BEFORE UPDATE ON t FOR EACH ROW EXECUTE PROCEDURE test_trigger();`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, ``, 4, ``}},
			},
			{
				Statement: `UPDATE t SET y = 2;`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, ``, 4, 2}},
			},
			{
				Statement: `DROP TABLE t;`,
			},
			{
				Statement: `CREATE TABLE t (id serial PRIMARY KEY, a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO t (a,b,c) VALUES (1,2,NULL);`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN x int;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN y int NOT NULL DEFAULT 5;`,
			},
			{
				Statement: `CREATE TRIGGER a BEFORE UPDATE ON t FOR EACH ROW EXECUTE PROCEDURE test_trigger();`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, ``, ``, 5}},
			},
			{
				Statement: `UPDATE t SET y = 2;`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, ``, ``, 2}},
			},
			{
				Statement: `DROP TABLE t;`,
			},
			{
				Statement: `CREATE TABLE t (id serial PRIMARY KEY, a int, b int, c int);`,
			},
			{
				Statement: `INSERT INTO t (a,b,c) VALUES (1,2,NULL);`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN x int;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN y int;`,
			},
			{
				Statement: `CREATE TRIGGER a BEFORE UPDATE ON t FOR EACH ROW EXECUTE PROCEDURE test_trigger();`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, ``, ``, ``}},
			},
			{
				Statement: `UPDATE t SET y = 2;`,
			},
			{
				Statement: `SELECT * FROM t;`,
				Results:   []sql.Row{{1, 1, 2, ``, ``, 2}},
			},
			{
				Statement: `DROP TABLE t;`,
			},
			{
				Statement: `CREATE TABLE leader (a int PRIMARY KEY, b int);`,
			},
			{
				Statement: `CREATE TABLE follower (a int REFERENCES leader ON DELETE CASCADE, b int);`,
			},
			{
				Statement: `INSERT INTO leader VALUES (1, 1), (2, 2);`,
			},
			{
				Statement: `ALTER TABLE leader ADD c int;`,
			},
			{
				Statement: `ALTER TABLE leader DROP c;`,
			},
			{
				Statement: `DELETE FROM leader;`,
			},
			{
				Statement: `CREATE TABLE vtype( a integer);`,
			},
			{
				Statement: `INSERT INTO vtype VALUES (1);`,
			},
			{
				Statement: `ALTER TABLE vtype ADD COLUMN b DOUBLE PRECISION DEFAULT 0.2;`,
			},
			{
				Statement: `ALTER TABLE vtype ADD COLUMN c BOOLEAN DEFAULT true;`,
			},
			{
				Statement: `SELECT * FROM vtype;`,
				Results:   []sql.Row{{1, 0.2, true}},
			},
			{
				Statement: `ALTER TABLE vtype
      ALTER b TYPE text USING b::text,
      ALTER c TYPE text USING c::text;`,
			},
			{
				Statement: `SELECT * FROM vtype;`,
				Results:   []sql.Row{{1, 0.2, `true`}},
			},
			{
				Statement: `CREATE TABLE vtype2 (a int);`,
			},
			{
				Statement: `INSERT INTO vtype2 VALUES (1);`,
			},
			{
				Statement: `ALTER TABLE vtype2 ADD COLUMN b varchar(10) DEFAULT 'xxx';`,
			},
			{
				Statement: `ALTER TABLE vtype2 ALTER COLUMN b SET DEFAULT 'yyy';`,
			},
			{
				Statement: `INSERT INTO vtype2 VALUES (2);`,
			},
			{
				Statement: `ALTER TABLE vtype2 ALTER COLUMN b TYPE varchar(20) USING b::varchar(20);`,
			},
			{
				Statement: `SELECT * FROM vtype2;`,
				Results:   []sql.Row{{1, `xxx`}, {2, `yyy`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TABLE t();`,
			},
			{
				Statement: `INSERT INTO t DEFAULT VALUES;`,
			},
			{
				Statement: `ALTER TABLE t ADD COLUMN a int DEFAULT 1;`,
			},
			{
				Statement: `CREATE INDEX ON t(a);`,
			},
			{
				Statement: `UPDATE t SET a = NULL;`,
			},
			{
				Statement: `SET LOCAL enable_seqscan = true;`,
			},
			{
				Statement: `SELECT * FROM t WHERE a IS NULL;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SET LOCAL enable_seqscan = false;`,
			},
			{
				Statement: `SELECT * FROM t WHERE a IS NULL;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER dummy;`,
			},
			{
				Statement: `CREATE SERVER s0 FOREIGN DATA WRAPPER dummy;`,
			},
			{
				Statement: `CREATE FOREIGN TABLE ft1 (c1 integer NOT NULL) SERVER s0;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ADD COLUMN c8 integer DEFAULT 0;`,
			},
			{
				Statement: `ALTER FOREIGN TABLE ft1 ALTER COLUMN c8 TYPE char(10);`,
			},
			{
				Statement: `SELECT count(*)
  FROM pg_attribute
  WHERE attrelid = 'ft1'::regclass AND
    (attmissingval IS NOT NULL OR atthasmissing);`,
				Results: []sql.Row{{0}},
			},
			{
				Statement: `DROP FOREIGN TABLE ft1;`,
			},
			{
				Statement: `DROP SERVER s0;`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER dummy;`,
			},
			{
				Statement: `DROP TABLE vtype;`,
			},
			{
				Statement: `DROP TABLE vtype2;`,
			},
			{
				Statement: `DROP TABLE follower;`,
			},
			{
				Statement: `DROP TABLE leader;`,
			},
			{
				Statement: `DROP FUNCTION test_trigger();`,
			},
			{
				Statement: `DROP TABLE t1;`,
			},
			{
				Statement: `DROP FUNCTION set(name);`,
			},
			{
				Statement: `DROP FUNCTION comp();`,
			},
			{
				Statement: `DROP TABLE m;`,
			},
			{
				Statement: `DROP TABLE has_volatile;`,
			},
			{
				Statement: `DROP EVENT TRIGGER has_volatile_rewrite;`,
			},
			{
				Statement: `DROP FUNCTION log_rewrite;`,
			},
			{
				Statement: `DROP SCHEMA fast_default;`,
			},
			{
				Statement: `set search_path = public;`,
			},
			{
				Statement: `create table has_fast_default(f1 int);`,
			},
			{
				Statement: `insert into has_fast_default values(1);`,
			},
			{
				Statement: `alter table has_fast_default add column f2 int default 42;`,
			},
			{
				Statement: `table has_fast_default;`,
				Results:   []sql.Row{{1, 42}},
			},
		},
	})
}
