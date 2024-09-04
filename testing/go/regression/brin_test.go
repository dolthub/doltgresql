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

func TestBrin(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_brin)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_brin,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE brintest (byteacol bytea,
	charcol "char",
	namecol name,
	int8col bigint,
	int2col smallint,
	int4col integer,
	textcol text,
	oidcol oid,
	tidcol tid,
	float4col real,
	float8col double precision,
	macaddrcol macaddr,
	inetcol inet,
	cidrcol cidr,
	bpcharcol character,
	datecol date,
	timecol time without time zone,
	timestampcol timestamp without time zone,
	timestamptzcol timestamp with time zone,
	intervalcol interval,
	timetzcol time with time zone,
	bitcol bit(10),
	varbitcol bit varying(16),
	numericcol numeric,
	uuidcol uuid,
	int4rangecol int4range,
	lsncol pg_lsn,
	boxcol box
) WITH (fillfactor=10, autovacuum_enabled=off);`,
			},
			{
				Statement: `INSERT INTO brintest SELECT
	repeat(stringu1, 8)::bytea,
	substr(stringu1, 1, 1)::"char",
	stringu1::name, 142857 * tenthous,
	thousand,
	twothousand,
	repeat(stringu1, 8),
	unique1::oid,
	format('(%s,%s)', tenthous, twenty)::tid,
	(four + 1.0)/(hundred+1),
	odd::float8 / (tenthous + 1),
	format('%s:00:%s:00:%s:00', to_hex(odd), to_hex(even), to_hex(hundred))::macaddr,
	inet '10.2.3.4/24' + tenthous,
	cidr '10.2.3/24' + tenthous,
	substr(stringu1, 1, 1)::bpchar,
	date '1995-08-15' + tenthous,
	time '01:20:30' + thousand * interval '18.5 second',
	timestamp '1942-07-23 03:05:09' + tenthous * interval '36.38 hours',
	timestamptz '1972-10-10 03:00' + thousand * interval '1 hour',
	justify_days(justify_hours(tenthous * interval '12 minutes')),
	timetz '01:30:20+02' + hundred * interval '15 seconds',
	thousand::bit(10),
	tenthous::bit(16)::varbit,
	tenthous::numeric(36,30) * fivethous * even / (hundred + 1),
	format('%s%s-%s-%s-%s-%s%s%s', to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'))::uuid,
	int4range(thousand, twothousand),
	format('%s/%s%s', odd, even, tenthous)::pg_lsn,
	box(point(odd, even), point(thousand, twothousand))
FROM tenk1 ORDER BY unique2 LIMIT 100;`,
			},
			{
				Statement: `INSERT INTO brintest (inetcol, cidrcol, int4rangecol) SELECT
	inet 'fe80::6e40:8ff:fea9:8c46' + tenthous,
	cidr 'fe80::6e40:8ff:fea9:8c46' + tenthous,
	'empty'::int4range
FROM tenk1 ORDER BY thousand, tenthous LIMIT 25;`,
			},
			{
				Statement: `CREATE INDEX brinidx ON brintest USING brin (
	byteacol,
	charcol,
	namecol,
	int8col,
	int2col,
	int4col,
	textcol,
	oidcol,
	tidcol,
	float4col,
	float8col,
	macaddrcol,
	inetcol inet_inclusion_ops,
	inetcol inet_minmax_ops,
	cidrcol inet_inclusion_ops,
	cidrcol inet_minmax_ops,
	bpcharcol,
	datecol,
	timecol,
	timestampcol,
	timestamptzcol,
	intervalcol,
	timetzcol,
	bitcol,
	varbitcol,
	numericcol,
	uuidcol,
	int4rangecol,
	lsncol,
	boxcol
) with (pages_per_range = 1);`,
			},
			{
				Statement: `CREATE TABLE brinopers (colname name, typ text,
	op text[], value text[], matches int[],
	check (cardinality(op) = cardinality(value)),
	check (cardinality(op) = cardinality(matches)));`,
			},
			{
				Statement: `INSERT INTO brinopers VALUES
	('byteacol', 'bytea',
	 '{>, >=, =, <=, <}',
	 '{AAAAAA, AAAAAA, BNAAAABNAAAABNAAAABNAAAABNAAAABNAAAABNAAAABNAAAA, ZZZZZZ, ZZZZZZ}',
	 '{100, 100, 1, 100, 100}'),
	('charcol', '"char"',
	 '{>, >=, =, <=, <}',
	 '{A, A, M, Z, Z}',
	 '{97, 100, 6, 100, 98}'),
	('namecol', 'name',
	 '{>, >=, =, <=, <}',
	 '{AAAAAA, AAAAAA, MAAAAA, ZZAAAA, ZZAAAA}',
	 '{100, 100, 2, 100, 100}'),
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
	('textcol', 'text',
	 '{>, >=, =, <=, <}',
	 '{ABABAB, ABABAB, BNAAAABNAAAABNAAAABNAAAABNAAAABNAAAABNAAAABNAAAA, ZZAAAA, ZZAAAA}',
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
	('inetcol', 'inet',
	 '{&&, =, <, <=, >, >=, >>=, >>, <<=, <<}',
	 '{10/8, 10.2.14.231/24, 255.255.255.255, 255.255.255.255, 0.0.0.0, 0.0.0.0, 10.2.14.231/24, 10.2.14.231/25, 10.2.14.231/8, 0/0}',
	 '{100, 1, 100, 100, 125, 125, 2, 2, 100, 100}'),
	('inetcol', 'inet',
	 '{&&, >>=, <<=, =}',
	 '{fe80::6e40:8ff:fea9:a673/32, fe80::6e40:8ff:fea9:8c46, fe80::6e40:8ff:fea9:a673/32, fe80::6e40:8ff:fea9:8c46}',
	 '{25, 1, 25, 1}'),
	('inetcol', 'cidr',
	 '{&&, <, <=, >, >=, >>=, >>, <<=, <<}',
	 '{10/8, 255.255.255.255, 255.255.255.255, 0.0.0.0, 0.0.0.0, 10.2.14/24, 10.2.14/25, 10/8, 0/0}',
	 '{100, 100, 100, 125, 125, 2, 2, 100, 100}'),
	('inetcol', 'cidr',
	 '{&&, >>=, <<=, =}',
	 '{fe80::/32, fe80::6e40:8ff:fea9:8c46, fe80::/32, fe80::6e40:8ff:fea9:8c46}',
	 '{25, 1, 25, 1}'),
	('cidrcol', 'inet',
	 '{&&, =, <, <=, >, >=, >>=, >>, <<=, <<}',
	 '{10/8, 10.2.14/24, 255.255.255.255, 255.255.255.255, 0.0.0.0, 0.0.0.0, 10.2.14.231/24, 10.2.14.231/25, 10.2.14.231/8, 0/0}',
	 '{100, 2, 100, 100, 125, 125, 2, 2, 100, 100}'),
	('cidrcol', 'inet',
	 '{&&, >>=, <<=, =}',
	 '{fe80::6e40:8ff:fea9:a673/32, fe80::6e40:8ff:fea9:8c46, fe80::6e40:8ff:fea9:a673/32, fe80::6e40:8ff:fea9:8c46}',
	 '{25, 1, 25, 1}'),
	('cidrcol', 'cidr',
	 '{&&, =, <, <=, >, >=, >>=, >>, <<=, <<}',
	 '{10/8, 10.2.14/24, 255.255.255.255, 255.255.255.255, 0.0.0.0, 0.0.0.0, 10.2.14/24, 10.2.14/25, 10/8, 0/0}',
	 '{100, 2, 100, 100, 125, 125, 2, 2, 100, 100}'),
	('cidrcol', 'cidr',
	 '{&&, >>=, <<=, =}',
	 '{fe80::/32, fe80::6e40:8ff:fea9:8c46, fe80::/32, fe80::6e40:8ff:fea9:8c46}',
	 '{25, 1, 25, 1}'),
	('bpcharcol', 'bpchar',
	 '{>, >=, =, <=, <}',
	 '{A, A, W, Z, Z}',
	 '{97, 100, 6, 100, 98}'),
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
	('bitcol', 'bit(10)',
	 '{>, >=, =, <=, <}',
	 '{0000000010, 0000000010, 0011011110, 1111111000, 1111111000}',
	 '{100, 100, 1, 100, 100}'),
	('varbitcol', 'varbit(16)',
	 '{>, >=, =, <=, <}',
	 '{0000000000000100, 0000000000000100, 0001010001100110, 1111111111111000, 1111111111111000}',
	 '{100, 100, 1, 100, 100}'),
	('numericcol', 'numeric',
	 '{>, >=, =, <=, <}',
	 '{0.00, 0.01, 2268164.347826086956521739130434782609, 99470151.9, 99470151.9}',
	 '{100, 100, 1, 100, 100}'),
	('uuidcol', 'uuid',
	 '{>, >=, =, <=, <}',
	 '{00040004-0004-0004-0004-000400040004, 00040004-0004-0004-0004-000400040004, 52225222-5222-5222-5222-522252225222, 99989998-9998-9998-9998-999899989998, 99989998-9998-9998-9998-999899989998}',
	 '{100, 100, 1, 100, 100}'),
	('int4rangecol', 'int4range',
	 '{<<, &<, &&, &>, >>, @>, <@, =, <, <=, >, >=}',
	 '{"[10000,)","[10000,)","(,]","[3,4)","[36,44)","(1500,1501]","[3,4)","[222,1222)","[36,44)","[43,1043)","[367,4466)","[519,)"}',
	 '{53, 53, 53, 53, 50, 22, 72, 1, 74, 75, 34, 21}'),
	('int4rangecol', 'int4range',
	 '{@>, <@, =, <=, >, >=}',
	 '{empty, empty, empty, empty, empty, empty}',
	 '{125, 72, 72, 72, 53, 125}'),
	('int4rangecol', 'int4',
	 '{@>}',
	 '{1500}',
	 '{22}'),
	('lsncol', 'pg_lsn',
	 '{>, >=, =, <=, <, IS, IS NOT}',
	 '{0/1200, 0/1200, 44/455222, 198/1999799, 198/1999799, NULL, NULL}',
	 '{100, 100, 1, 100, 100, 25, 100}'),
	('boxcol', 'point',
	 '{@>}',
	 '{"(500,43)"}',
	 '{11}'),
	('boxcol', 'box',
	 '{<<, &<, &&, &>, >>, <<|, &<|, |&>, |>>, @>, <@, ~=}',
	 '{"((1000,2000),(3000,4000))","((1,2),(3000,4000))","((1,2),(3000,4000))","((1,2),(3000,4000))","((1,2),(3,4))","((1000,2000),(3000,4000))","((1,2000),(3,4000))","((1000,2),(3000,4))","((1,2),(3,4))","((1,2),(300,400))","((1,2),(3000,4000))","((222,1222),(44,45))"}',
	 '{100, 100, 100, 99, 96, 100, 100, 99, 96, 1, 99, 1}');`,
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
	FOR r IN SELECT colname, oper, typ, value[ordinality], matches[ordinality] FROM brinopers, unnest(op) WITH ORDINALITY AS oper LOOP
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
				Statement: `		FOR plan_line IN EXECUTE format($y$EXPLAIN SELECT array_agg(ctid) FROM brintest WHERE %s $y$, cond) LOOP
			IF plan_line LIKE '%Bitmap Heap Scan on brintest%' THEN
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
				Statement: `		EXECUTE format($y$SELECT array_agg(ctid) FROM brintest WHERE %s $y$, cond)
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
				Statement: `		FOR plan_line IN EXECUTE format($y$EXPLAIN SELECT array_agg(ctid) FROM brintest WHERE %s $y$, cond) LOOP
			IF plan_line LIKE '%Seq Scan on brintest%' THEN
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
				Statement: `		EXECUTE format($y$SELECT array_agg(ctid) FROM brintest WHERE %s $y$, cond)
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
				Statement: `			FOR r2 IN EXECUTE 'SELECT ' || r.colname || ' FROM brintest WHERE ' || cond LOOP
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
				Statement: `			FOR r2 IN EXECUTE 'SELECT ' || r.colname || ' FROM brintest WHERE ' || cond LOOP
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
				Statement: `INSERT INTO brintest SELECT
	repeat(stringu1, 42)::bytea,
	substr(stringu1, 1, 1)::"char",
	stringu1::name, 142857 * tenthous,
	thousand,
	twothousand,
	repeat(stringu1, 42),
	unique1::oid,
	format('(%s,%s)', tenthous, twenty)::tid,
	(four + 1.0)/(hundred+1),
	odd::float8 / (tenthous + 1),
	format('%s:00:%s:00:%s:00', to_hex(odd), to_hex(even), to_hex(hundred))::macaddr,
	inet '10.2.3.4' + tenthous,
	cidr '10.2.3/24' + tenthous,
	substr(stringu1, 1, 1)::bpchar,
	date '1995-08-15' + tenthous,
	time '01:20:30' + thousand * interval '18.5 second',
	timestamp '1942-07-23 03:05:09' + tenthous * interval '36.38 hours',
	timestamptz '1972-10-10 03:00' + thousand * interval '1 hour',
	justify_days(justify_hours(tenthous * interval '12 minutes')),
	timetz '01:30:20' + hundred * interval '15 seconds',
	thousand::bit(10),
	tenthous::bit(16)::varbit,
	tenthous::numeric(36,30) * fivethous * even / (hundred + 1),
	format('%s%s-%s-%s-%s-%s%s%s', to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'))::uuid,
	int4range(thousand, twothousand),
	format('%s/%s%s', odd, even, tenthous)::pg_lsn,
	box(point(odd, even), point(thousand, twothousand))
FROM tenk1 ORDER BY unique2 LIMIT 5 OFFSET 5;`,
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `VACUUM brintest;  -- force a summarization cycle in brinidx`,
			},
			{
				Statement: `UPDATE brintest SET int8col = int8col * int4col;`,
			},
			{
				Statement: `UPDATE brintest SET textcol = '' WHERE textcol IS NOT NULL;`,
			},
			{
				Statement:   `SELECT brin_summarize_new_values('brintest'); -- error, not an index`,
				ErrorString: `"brintest" is not an index`,
			},
			{
				Statement:   `SELECT brin_summarize_new_values('tenk1_unique1'); -- error, not a BRIN index`,
				ErrorString: `"tenk1_unique1" is not a BRIN index`,
			},
			{
				Statement: `SELECT brin_summarize_new_values('brinidx'); -- ok, no change expected`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT brin_desummarize_range('brinidx', -1); -- error, invalid range`,
				ErrorString: `block number out of range: -1`,
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx', 100000000);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `CREATE TABLE brin_summarize (
    value int
) WITH (fillfactor=10, autovacuum_enabled=false);`,
			},
			{
				Statement: `CREATE INDEX brin_summarize_idx ON brin_summarize USING brin (value) WITH (pages_per_range=2);`,
			},
			{
				Statement: `DO $$
DECLARE curtid tid;`,
			},
			{
				Statement: `BEGIN
  LOOP
    INSERT INTO brin_summarize VALUES (1) RETURNING ctid INTO curtid;`,
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
				Statement: `SELECT brin_summarize_range('brin_summarize_idx', 0);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_idx', 1);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_idx', 2);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_idx', 4294967295);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT brin_summarize_range('brin_summarize_idx', -1);`,
				ErrorString: `block number out of range: -1`,
			},
			{
				Statement:   `SELECT brin_summarize_range('brin_summarize_idx', 4294967296);`,
				ErrorString: `block number out of range: 4294967296`,
			},
			{
				Statement: `CREATE TABLE brintest_2 (n numrange);`,
			},
			{
				Statement: `CREATE INDEX brinidx_2 ON brintest_2 USING brin (n);`,
			},
			{
				Statement: `INSERT INTO brintest_2 VALUES ('empty');`,
			},
			{
				Statement: `INSERT INTO brintest_2 VALUES (numrange(0, 2^1000::numeric));`,
			},
			{
				Statement: `INSERT INTO brintest_2 VALUES ('(-1, 0)');`,
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT brin_summarize_range('brinidx', 0);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP TABLE brintest_2;`,
			},
			{
				Statement: `CREATE TABLE brin_test (a INT, b INT);`,
			},
			{
				Statement: `INSERT INTO brin_test SELECT x/100,x%100 FROM generate_series(1,10000) x(x);`,
			},
			{
				Statement: `CREATE INDEX brin_test_a_idx ON brin_test USING brin (a) WITH (pages_per_range = 2);`,
			},
			{
				Statement: `CREATE INDEX brin_test_b_idx ON brin_test USING brin (b) WITH (pages_per_range = 2);`,
			},
			{
				Statement: `VACUUM ANALYZE brin_test;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM brin_test WHERE a = 1;`,
				Results:   []sql.Row{{`Bitmap Heap Scan on brin_test`}, {`Recheck Cond: (a = 1)`}, {`->  Bitmap Index Scan on brin_test_a_idx`}, {`Index Cond: (a = 1)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM brin_test WHERE b = 1;`,
				Results:   []sql.Row{{`Seq Scan on brin_test`}, {`Filter: (b = 1)`}},
			},
			{
				Statement: `CREATE TABLE brintest_3 (a text, b text, c text, d text);`,
			},
			{
				Statement: `WITH rand_value AS (SELECT string_agg(md5(i::text),'') AS val FROM generate_series(1,60) s(i))
INSERT INTO brintest_3
SELECT val, val, val, val FROM rand_value;`,
			},
			{
				Statement: `CREATE INDEX brin_test_toast_idx ON brintest_3 USING brin (b, c);`,
			},
			{
				Statement: `DELETE FROM brintest_3;`,
			},
			{
				Statement: `CREATE INDEX CONCURRENTLY brin_test_temp_idx ON brintest_3(a);`,
			},
			{
				Statement: `DROP INDEX brin_test_temp_idx;`,
			},
			{
				Statement: `VACUUM brintest_3;`,
			},
			{
				Statement: `WITH rand_value AS (SELECT string_agg(md5((-i)::text),'') AS val FROM generate_series(1,60) s(i))
INSERT INTO brintest_3
SELECT val, val, val, val FROM rand_value;`,
			},
			{
				Statement: `SET enable_seqscan = off;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM brintest_3 WHERE b < '0';`,
				Results: []sql.Row{{`Bitmap Heap Scan on brintest_3`}, {`Recheck Cond: (b < '0'::text)`}, {`->  Bitmap Index Scan on brin_test_toast_idx`}, {`Index Cond: (b < '0'::text)`}},
			},
			{
				Statement: `SELECT * FROM brintest_3 WHERE b < '0';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE brintest_3;`,
			},
			{
				Statement: `RESET enable_seqscan;`,
			},
			{
				Statement: `CREATE UNLOGGED TABLE brintest_unlogged (n numrange);`,
			},
			{
				Statement: `CREATE INDEX brinidx_unlogged ON brintest_unlogged USING brin (n);`,
			},
			{
				Statement: `INSERT INTO brintest_unlogged VALUES (numrange(0, 2^1000::numeric));`,
			},
			{
				Statement: `DROP TABLE brintest_unlogged;`,
			},
		},
	})
}
