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

func TestBrinMulti(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_brin_multi)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_brin_multi,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE brintest_multi (
	int8col bigint,
	int2col smallint,
	int4col integer,
	oidcol oid,
	tidcol tid,
	float4col real,
	float8col double precision,
	macaddrcol macaddr,
	macaddr8col macaddr8,
	inetcol inet,
	cidrcol cidr,
	datecol date,
	timecol time without time zone,
	timestampcol timestamp without time zone,
	timestamptzcol timestamp with time zone,
	intervalcol interval,
	timetzcol time with time zone,
	numericcol numeric,
	uuidcol uuid,
	lsncol pg_lsn
) WITH (fillfactor=10);`,
			},
			{
				Statement: `INSERT INTO brintest_multi SELECT
	142857 * tenthous,
	thousand,
	twothousand,
	unique1::oid,
	format('(%s,%s)', tenthous, twenty)::tid,
	(four + 1.0)/(hundred+1),
	odd::float8 / (tenthous + 1),
	format('%s:00:%s:00:%s:00', to_hex(odd), to_hex(even), to_hex(hundred))::macaddr,
	substr(md5(unique1::text), 1, 16)::macaddr8,
	inet '10.2.3.4/24' + tenthous,
	cidr '10.2.3/24' + tenthous,
	date '1995-08-15' + tenthous,
	time '01:20:30' + thousand * interval '18.5 second',
	timestamp '1942-07-23 03:05:09' + tenthous * interval '36.38 hours',
	timestamptz '1972-10-10 03:00' + thousand * interval '1 hour',
	justify_days(justify_hours(tenthous * interval '12 minutes')),
	timetz '01:30:20+02' + hundred * interval '15 seconds',
	tenthous::numeric(36,30) * fivethous * even / (hundred + 1),
	format('%s%s-%s-%s-%s-%s%s%s', to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'))::uuid,
	format('%s/%s%s', odd, even, tenthous)::pg_lsn
FROM tenk1 ORDER BY unique2 LIMIT 100;`,
			},
			{
				Statement: `INSERT INTO brintest_multi (inetcol, cidrcol) SELECT
	inet 'fe80::6e40:8ff:fea9:8c46' + tenthous,
	cidr 'fe80::6e40:8ff:fea9:8c46' + tenthous
FROM tenk1 ORDER BY thousand, tenthous LIMIT 25;`,
			},
			{
				Statement: `CREATE INDEX brinidx_multi ON brintest_multi USING brin (
	int8col int8_minmax_multi_ops(values_per_range = 7)
);`,
				ErrorString: `value 7 out of bounds for option "values_per_range"`,
			},
			{
				Statement: `CREATE INDEX brinidx_multi ON brintest_multi USING brin (
	int8col int8_minmax_multi_ops(values_per_range = 257)
);`,
				ErrorString: `value 257 out of bounds for option "values_per_range"`,
			},
			{
				Statement: `CREATE INDEX brinidx_multi ON brintest_multi USING brin (
	int8col int8_minmax_multi_ops,
	int2col int2_minmax_multi_ops,
	int4col int4_minmax_multi_ops,
	oidcol oid_minmax_multi_ops,
	tidcol tid_minmax_multi_ops,
	float4col float4_minmax_multi_ops,
	float8col float8_minmax_multi_ops,
	macaddrcol macaddr_minmax_multi_ops,
	macaddr8col macaddr8_minmax_multi_ops,
	inetcol inet_minmax_multi_ops,
	cidrcol inet_minmax_multi_ops,
	datecol date_minmax_multi_ops,
	timecol time_minmax_multi_ops,
	timestampcol timestamp_minmax_multi_ops,
	timestamptzcol timestamptz_minmax_multi_ops,
	intervalcol interval_minmax_multi_ops,
	timetzcol timetz_minmax_multi_ops,
	numericcol numeric_minmax_multi_ops,
	uuidcol uuid_minmax_multi_ops,
	lsncol pg_lsn_minmax_multi_ops
);`,
			},
			{
				Statement: `DROP INDEX brinidx_multi;`,
			},
			{
				Statement: `CREATE INDEX brinidx_multi ON brintest_multi USING brin (
	int8col int8_minmax_multi_ops,
	int2col int2_minmax_multi_ops,
	int4col int4_minmax_multi_ops,
	oidcol oid_minmax_multi_ops,
	tidcol tid_minmax_multi_ops,
	float4col float4_minmax_multi_ops,
	float8col float8_minmax_multi_ops,
	macaddrcol macaddr_minmax_multi_ops,
	macaddr8col macaddr8_minmax_multi_ops,
	inetcol inet_minmax_multi_ops,
	cidrcol inet_minmax_multi_ops,
	datecol date_minmax_multi_ops,
	timecol time_minmax_multi_ops,
	timestampcol timestamp_minmax_multi_ops,
	timestamptzcol timestamptz_minmax_multi_ops,
	intervalcol interval_minmax_multi_ops,
	timetzcol timetz_minmax_multi_ops,
	numericcol numeric_minmax_multi_ops,
	uuidcol uuid_minmax_multi_ops,
	lsncol pg_lsn_minmax_multi_ops
) with (pages_per_range = 1);`,
			},
			{
				Statement: `CREATE TABLE brinopers_multi (colname name, typ text,
	op text[], value text[], matches int[],
	check (cardinality(op) = cardinality(value)),
	check (cardinality(op) = cardinality(matches)));`,
			},
			{
				Statement: `INSERT INTO brinopers_multi VALUES
	('int2col', 'int2',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 800, 999, 999}',
	 '{100, 100, 1, 100, 100}'),
	('int2col', 'int4',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 800, 999, 1999}',
	 '{100, 100, 1, 100, 100}'),
	('int2col', 'int8',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 800, 999, 1428427143}',
	 '{100, 100, 1, 100, 100}'),
	('int4col', 'int2',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 800, 1999, 1999}',
	 '{100, 100, 1, 100, 100}'),
	('int4col', 'int4',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 800, 1999, 1999}',
	 '{100, 100, 1, 100, 100}'),
	('int4col', 'int8',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 800, 1999, 1428427143}',
	 '{100, 100, 1, 100, 100}'),
	('int8col', 'int2',
	 '{>, >=}',
	 '{0, 0}',
	 '{100, 100}'),
	('int8col', 'int4',
	 '{>, >=}',
	 '{0, 0}',
	 '{100, 100}'),
	('int8col', 'int8',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 1257141600, 1428427143, 1428427143}',
	 '{100, 100, 1, 100, 100}'),
	('oidcol', 'oid',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 8800, 9999, 9999}',
	 '{100, 100, 1, 100, 100}'),
	('tidcol', 'tid',
	 '{>, >=, =, <=, <}',
	 '{"(0,0)", "(0,0)", "(8800,0)", "(9999,19)", "(9999,19)"}',
	 '{100, 100, 1, 100, 100}'),
	('float4col', 'float4',
	 '{>, >=, =, <=, <}',
	 '{0.0103093, 0.0103093, 1, 1, 1}',
	 '{100, 100, 4, 100, 96}'),
	('float4col', 'float8',
	 '{>, >=, =, <=, <}',
	 '{0.0103093, 0.0103093, 1, 1, 1}',
	 '{100, 100, 4, 100, 96}'),
	('float8col', 'float4',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 0, 1.98, 1.98}',
	 '{99, 100, 1, 100, 100}'),
	('float8col', 'float8',
	 '{>, >=, =, <=, <}',
	 '{0, 0, 0, 1.98, 1.98}',
	 '{99, 100, 1, 100, 100}'),
	('macaddrcol', 'macaddr',
	 '{>, >=, =, <=, <}',
	 '{00:00:01:00:00:00, 00:00:01:00:00:00, 2c:00:2d:00:16:00, ff:fe:00:00:00:00, ff:fe:00:00:00:00}',
	 '{99, 100, 2, 100, 100}'),
	('macaddr8col', 'macaddr8',
	 '{>, >=, =, <=, <}',
	 '{b1:d1:0e:7b:af:a4:42:12, d9:35:91:bd:f7:86:0e:1e, 72:8f:20:6c:2a:01:bf:57, 23:e8:46:63:86:07:ad:cb, 13:16:8e:6a:2e:6c:84:b4}',
	 '{33, 15, 1, 13, 6}'),
	('inetcol', 'inet',
	 '{=, <, <=, >, >=}',
	 '{10.2.14.231/24, 255.255.255.255, 255.255.255.255, 0.0.0.0, 0.0.0.0}',
	 '{1, 100, 100, 125, 125}'),
	('inetcol', 'cidr',
	 '{<, <=, >, >=}',
	 '{255.255.255.255, 255.255.255.255, 0.0.0.0, 0.0.0.0}',
	 '{100, 100, 125, 125}'),
	('cidrcol', 'inet',
	 '{=, <, <=, >, >=}',
	 '{10.2.14/24, 255.255.255.255, 255.255.255.255, 0.0.0.0, 0.0.0.0}',
	 '{2, 100, 100, 125, 125}'),
	('cidrcol', 'cidr',
	 '{=, <, <=, >, >=}',
	 '{10.2.14/24, 255.255.255.255, 255.255.255.255, 0.0.0.0, 0.0.0.0}',
	 '{2, 100, 100, 125, 125}'),
	('datecol', 'date',
	 '{>, >=, =, <=, <}',
	 '{1995-08-15, 1995-08-15, 2009-12-01, 2022-12-30, 2022-12-30}',
	 '{100, 100, 1, 100, 100}'),
	('timecol', 'time',
	 '{>, >=, =, <=, <}',
	 '{01:20:30, 01:20:30, 02:28:57, 06:28:31.5, 06:28:31.5}',
	 '{100, 100, 1, 100, 100}'),
	('timestampcol', 'timestamp',
	 '{>, >=, =, <=, <}',
	 '{1942-07-23 03:05:09, 1942-07-23 03:05:09, 1964-03-24 19:26:45, 1984-01-20 22:42:21, 1984-01-20 22:42:21}',
	 '{100, 100, 1, 100, 100}'),
	('timestampcol', 'timestamptz',
	 '{>, >=, =, <=, <}',
	 '{1942-07-23 03:05:09, 1942-07-23 03:05:09, 1964-03-24 19:26:45, 1984-01-20 22:42:21, 1984-01-20 22:42:21}',
	 '{100, 100, 1, 100, 100}'),
	('timestamptzcol', 'timestamptz',
	 '{>, >=, =, <=, <}',
	 '{1972-10-10 03:00:00-04, 1972-10-10 03:00:00-04, 1972-10-19 09:00:00-07, 1972-11-20 19:00:00-03, 1972-11-20 19:00:00-03}',
	 '{100, 100, 1, 100, 100}'),
	('intervalcol', 'interval',
	 '{>, >=, =, <=, <}',
	 '{00:00:00, 00:00:00, 1 mons 13 days 12:24, 2 mons 23 days 07:48:00, 1 year}',
	 '{100, 100, 1, 100, 100}'),
	('timetzcol', 'timetz',
	 '{>, >=, =, <=, <}',
	 '{01:30:20+02, 01:30:20+02, 01:35:50+02, 23:55:05+02, 23:55:05+02}',
	 '{99, 100, 2, 100, 100}'),
	('numericcol', 'numeric',
	 '{>, >=, =, <=, <}',
	 '{0.00, 0.01, 2268164.347826086956521739130434782609, 99470151.9, 99470151.9}',
	 '{100, 100, 1, 100, 100}'),
	('uuidcol', 'uuid',
	 '{>, >=, =, <=, <}',
	 '{00040004-0004-0004-0004-000400040004, 00040004-0004-0004-0004-000400040004, 52225222-5222-5222-5222-522252225222, 99989998-9998-9998-9998-999899989998, 99989998-9998-9998-9998-999899989998}',
	 '{100, 100, 1, 100, 100}'),
	('lsncol', 'pg_lsn',
	 '{>, >=, =, <=, <, IS, IS NOT}',
	 '{0/1200, 0/1200, 44/455222, 198/1999799, 198/1999799, NULL, NULL}',
	 '{100, 100, 1, 100, 100, 25, 100}');`,
			},
			{
				Statement: `DO $x$
DECLARE
	r record;`,
			},
			{
				Statement: `	r2 record;`,
			},
			{
				Statement: `	cond text;`,
			},
			{
				Statement: `	idx_ctids tid[];`,
			},
			{
				Statement: `	ss_ctids tid[];`,
			},
			{
				Statement: `	count int;`,
			},
			{
				Statement: `	plan_ok bool;`,
			},
			{
				Statement: `	plan_line text;`,
			},
			{
				Statement: `BEGIN
	FOR r IN SELECT colname, oper, typ, value[ordinality], matches[ordinality] FROM brinopers_multi, unnest(op) WITH ORDINALITY AS oper LOOP
		-- prepare the condition
		IF r.value IS NULL THEN
			cond := format('%I %s %L', r.colname, r.oper, r.value);`,
			},
			{
				Statement: `		ELSE
			cond := format('%I %s %L::%s', r.colname, r.oper, r.value, r.typ);`,
			},
			{
				Statement: `		END IF;`,
			},
			{
				Statement: `		-- run the query using the brin index
		SET enable_seqscan = 0;`,
			},
			{
				Statement: `		SET enable_bitmapscan = 1;`,
			},
			{
				Statement: `		plan_ok := false;`,
			},
			{
				Statement: `		FOR plan_line IN EXECUTE format($y$EXPLAIN SELECT array_agg(ctid) FROM brintest_multi WHERE %s $y$, cond) LOOP
			IF plan_line LIKE '%Bitmap Heap Scan on brintest_multi%' THEN
				plan_ok := true;`,
			},
			{
				Statement: `			END IF;`,
			},
			{
				Statement: `		END LOOP;`,
			},
			{
				Statement: `		IF NOT plan_ok THEN
			RAISE WARNING 'did not get bitmap indexscan plan for %', r;`,
			},
			{
				Statement: `		END IF;`,
			},
			{
				Statement: `		EXECUTE format($y$SELECT array_agg(ctid) FROM brintest_multi WHERE %s $y$, cond)
			INTO idx_ctids;`,
			},
			{
				Statement: `		-- run the query using a seqscan
		SET enable_seqscan = 1;`,
			},
			{
				Statement: `		SET enable_bitmapscan = 0;`,
			},
			{
				Statement: `		plan_ok := false;`,
			},
			{
				Statement: `		FOR plan_line IN EXECUTE format($y$EXPLAIN SELECT array_agg(ctid) FROM brintest_multi WHERE %s $y$, cond) LOOP
			IF plan_line LIKE '%Seq Scan on brintest_multi%' THEN
				plan_ok := true;`,
			},
			{
				Statement: `			END IF;`,
			},
			{
				Statement: `		END LOOP;`,
			},
			{
				Statement: `		IF NOT plan_ok THEN
			RAISE WARNING 'did not get seqscan plan for %', r;`,
			},
			{
				Statement: `		END IF;`,
			},
			{
				Statement: `		EXECUTE format($y$SELECT array_agg(ctid) FROM brintest_multi WHERE %s $y$, cond)
			INTO ss_ctids;`,
			},
			{
				Statement: `		-- make sure both return the same results
		count := array_length(idx_ctids, 1);`,
			},
			{
				Statement: `		IF NOT (count = array_length(ss_ctids, 1) AND
				idx_ctids @> ss_ctids AND
				idx_ctids <@ ss_ctids) THEN
			-- report the results of each scan to make the differences obvious
			RAISE WARNING 'something not right in %: count %', r, count;`,
			},
			{
				Statement: `			SET enable_seqscan = 1;`,
			},
			{
				Statement: `			SET enable_bitmapscan = 0;`,
			},
			{
				Statement: `			FOR r2 IN EXECUTE 'SELECT ' || r.colname || ' FROM brintest_multi WHERE ' || cond LOOP
				RAISE NOTICE 'seqscan: %', r2;`,
			},
			{
				Statement: `			END LOOP;`,
			},
			{
				Statement: `			SET enable_seqscan = 0;`,
			},
			{
				Statement: `			SET enable_bitmapscan = 1;`,
			},
			{
				Statement: `			FOR r2 IN EXECUTE 'SELECT ' || r.colname || ' FROM brintest_multi WHERE ' || cond LOOP
				RAISE NOTICE 'bitmapscan: %', r2;`,
			},
			{
				Statement: `			END LOOP;`,
			},
			{
				Statement: `		END IF;`,
			},
			{
				Statement: `		-- make sure we found expected number of matches
		IF count != r.matches THEN RAISE WARNING 'unexpected number of results % for %', count, r; END IF;`,
			},
			{
				Statement: `	END LOOP;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$x$;`,
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `RESET enable_bitmapscan;`,
			},
			{
				Statement: `INSERT INTO brintest_multi SELECT
	142857 * tenthous,
	thousand,
	twothousand,
	unique1::oid,
	format('(%s,%s)', tenthous, twenty)::tid,
	(four + 1.0)/(hundred+1),
	odd::float8 / (tenthous + 1),
	format('%s:00:%s:00:%s:00', to_hex(odd), to_hex(even), to_hex(hundred))::macaddr,
	substr(md5(unique1::text), 1, 16)::macaddr8,
	inet '10.2.3.4' + tenthous,
	cidr '10.2.3/24' + tenthous,
	date '1995-08-15' + tenthous,
	time '01:20:30' + thousand * interval '18.5 second',
	timestamp '1942-07-23 03:05:09' + tenthous * interval '36.38 hours',
	timestamptz '1972-10-10 03:00' + thousand * interval '1 hour',
	justify_days(justify_hours(tenthous * interval '12 minutes')),
	timetz '01:30:20' + hundred * interval '15 seconds',
	tenthous::numeric(36,30) * fivethous * even / (hundred + 1),
	format('%s%s-%s-%s-%s-%s%s%s', to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'))::uuid,
	format('%s/%s%s', odd, even, tenthous)::pg_lsn
FROM tenk1 ORDER BY unique2 LIMIT 5 OFFSET 5;`,
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx_multi', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `VACUUM brintest_multi;  -- force a summarization cycle in brinidx`,
			},
			{
				Statement: `insert into public.brintest_multi (float4col) values (real 'nan');`,
			},
			{
				Statement: `insert into public.brintest_multi (float8col) values (real 'nan');`,
			},
			{
				Statement: `UPDATE brintest_multi SET int8col = int8col * int4col;`,
			},
			{
				Statement: `CREATE TABLE brin_test_inet (a inet);`,
			},
			{
				Statement: `CREATE INDEX ON brin_test_inet USING brin (a inet_minmax_multi_ops);`,
			},
			{
				Statement: `INSERT INTO brin_test_inet VALUES ('127.0.0.1/0');`,
			},
			{
				Statement: `INSERT INTO brin_test_inet VALUES ('0.0.0.0/12');`,
			},
			{
				Statement: `DROP TABLE brin_test_inet;`,
			},
			{
				Statement:   `SELECT brin_summarize_new_values('brintest_multi'); -- error, not an index`,
				ErrorString: `"brintest_multi" is not an index`,
			},
			{
				Statement:   `SELECT brin_summarize_new_values('tenk1_unique1'); -- error, not a BRIN index`,
				ErrorString: `"tenk1_unique1" is not a BRIN index`,
			},
			{
				Statement: `SELECT brin_summarize_new_values('brinidx_multi'); -- ok, no change expected`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT brin_desummarize_range('brinidx_multi', -1); -- error, invalid range`,
				ErrorString: `block number out of range: -1`,
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx_multi', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx_multi', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx_multi', 100000000);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `CREATE TABLE brin_large_range (a int4);`,
			},
			{
				Statement: `INSERT INTO brin_large_range SELECT i FROM generate_series(1,10000) s(i);`,
			},
			{
				Statement: `CREATE INDEX brin_large_range_idx ON brin_large_range USING brin (a int4_minmax_multi_ops);`,
			},
			{
				Statement: `DROP TABLE brin_large_range;`,
			},
			{
				Statement: `CREATE TABLE brin_summarize_multi (
    value int
) WITH (fillfactor=10, autovacuum_enabled=false);`,
			},
			{
				Statement: `CREATE INDEX brin_summarize_multi_idx ON brin_summarize_multi USING brin (value) WITH (pages_per_range=2);`,
			},
			{
				Statement: `DO $$
DECLARE curtid tid;`,
			},
			{
				Statement: `BEGIN
  LOOP
    INSERT INTO brin_summarize_multi VALUES (1) RETURNING ctid INTO curtid;`,
			},
			{
				Statement: `    EXIT WHEN curtid > tid '(2, 0)';`,
			},
			{
				Statement: `  END LOOP;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_multi_idx', 0);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_multi_idx', 1);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_multi_idx', 2);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_multi_idx', 4294967295);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT brin_summarize_range('brin_summarize_multi_idx', -1);`,
				ErrorString: `block number out of range: -1`,
			},
			{
				Statement:   `SELECT brin_summarize_range('brin_summarize_multi_idx', 4294967296);`,
				ErrorString: `block number out of range: 4294967296`,
			},
			{
				Statement: `CREATE TABLE brin_test_multi (a INT, b INT);`,
			},
			{
				Statement: `INSERT INTO brin_test_multi SELECT x/100,x%100 FROM generate_series(1,10000) x(x);`,
			},
			{
				Statement: `CREATE INDEX brin_test_multi_a_idx ON brin_test_multi USING brin (a) WITH (pages_per_range = 2);`,
			},
			{
				Statement: `CREATE INDEX brin_test_multi_b_idx ON brin_test_multi USING brin (b) WITH (pages_per_range = 2);`,
			},
			{
				Statement: `VACUUM ANALYZE brin_test_multi;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM brin_test_multi WHERE a = 1;`,
				Results:   []sql.Row{{`Bitmap Heap Scan on brin_test_multi`}, {`Recheck Cond: (a = 1)`}, {`->  Bitmap Index Scan on brin_test_multi_a_idx`}, {`Index Cond: (a = 1)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM brin_test_multi WHERE b = 1;`,
				Results:   []sql.Row{{`Seq Scan on brin_test_multi`}, {`Filter: (b = 1)`}},
			},
		},
	})
}
