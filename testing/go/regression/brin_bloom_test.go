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

func TestBrinBloom(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_brin_bloom)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_brin_bloom,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE brintest_bloom (byteacol bytea,
	charcol "char",
	namecol name,
	int8col bigint,
	int2col smallint,
	int4col integer,
	textcol text,
	oidcol oid,
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
	numericcol numeric,
	uuidcol uuid,
	lsncol pg_lsn
) WITH (fillfactor=10);`,
			},
			{
				Statement: `INSERT INTO brintest_bloom SELECT
	repeat(stringu1, 8)::bytea,
	substr(stringu1, 1, 1)::"char",
	stringu1::name, 142857 * tenthous,
	thousand,
	twothousand,
	repeat(stringu1, 8),
	unique1::oid,
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
	tenthous::numeric(36,30) * fivethous * even / (hundred + 1),
	format('%s%s-%s-%s-%s-%s%s%s', to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'))::uuid,
	format('%s/%s%s', odd, even, tenthous)::pg_lsn
FROM tenk1 ORDER BY unique2 LIMIT 100;`,
			},
			{
				Statement: `INSERT INTO brintest_bloom (inetcol, cidrcol) SELECT
	inet 'fe80::6e40:8ff:fea9:8c46' + tenthous,
	cidr 'fe80::6e40:8ff:fea9:8c46' + tenthous
FROM tenk1 ORDER BY thousand, tenthous LIMIT 25;`,
			},
			{
				Statement: `CREATE INDEX brinidx_bloom ON brintest_bloom USING brin (
	byteacol bytea_bloom_ops(n_distinct_per_range = -1.1)
);`,
				ErrorString: `value -1.1 out of bounds for option "n_distinct_per_range"`,
			},
			{
				Statement: `CREATE INDEX brinidx_bloom ON brintest_bloom USING brin (
	byteacol bytea_bloom_ops(false_positive_rate = 0.00009)
);`,
				ErrorString: `value 0.00009 out of bounds for option "false_positive_rate"`,
			},
			{
				Statement: `CREATE INDEX brinidx_bloom ON brintest_bloom USING brin (
	byteacol bytea_bloom_ops(false_positive_rate = 0.26)
);`,
				ErrorString: `value 0.26 out of bounds for option "false_positive_rate"`,
			},
			{
				Statement: `CREATE INDEX brinidx_bloom ON brintest_bloom USING brin (
	byteacol bytea_bloom_ops,
	charcol char_bloom_ops,
	namecol name_bloom_ops,
	int8col int8_bloom_ops,
	int2col int2_bloom_ops,
	int4col int4_bloom_ops,
	textcol text_bloom_ops,
	oidcol oid_bloom_ops,
	float4col float4_bloom_ops,
	float8col float8_bloom_ops,
	macaddrcol macaddr_bloom_ops,
	inetcol inet_bloom_ops,
	cidrcol inet_bloom_ops,
	bpcharcol bpchar_bloom_ops,
	datecol date_bloom_ops,
	timecol time_bloom_ops,
	timestampcol timestamp_bloom_ops,
	timestamptzcol timestamptz_bloom_ops,
	intervalcol interval_bloom_ops,
	timetzcol timetz_bloom_ops,
	numericcol numeric_bloom_ops,
	uuidcol uuid_bloom_ops,
	lsncol pg_lsn_bloom_ops
) with (pages_per_range = 1);`,
			},
			{
				Statement: `CREATE TABLE brinopers_bloom (colname name, typ text,
	op text[], value text[], matches int[],
	check (cardinality(op) = cardinality(value)),
	check (cardinality(op) = cardinality(matches)));`,
			},
			{
				Statement: `INSERT INTO brinopers_bloom VALUES
	('byteacol', 'bytea',
	 '{=}',
	 '{BNAAAABNAAAABNAAAABNAAAABNAAAABNAAAABNAAAABNAAAA}',
	 '{1}'),
	('charcol', '"char"',
	 '{=}',
	 '{M}',
	 '{6}'),
	('namecol', 'name',
	 '{=}',
	 '{MAAAAA}',
	 '{2}'),
	('int2col', 'int2',
	 '{=}',
	 '{800}',
	 '{1}'),
	('int4col', 'int4',
	 '{=}',
	 '{800}',
	 '{1}'),
	('int8col', 'int8',
	 '{=}',
	 '{1257141600}',
	 '{1}'),
	('textcol', 'text',
	 '{=}',
	 '{BNAAAABNAAAABNAAAABNAAAABNAAAABNAAAABNAAAABNAAAA}',
	 '{1}'),
	('oidcol', 'oid',
	 '{=}',
	 '{8800}',
	 '{1}'),
	('float4col', 'float4',
	 '{=}',
	 '{1}',
	 '{4}'),
	('float8col', 'float8',
	 '{=}',
	 '{0}',
	 '{1}'),
	('macaddrcol', 'macaddr',
	 '{=}',
	 '{2c:00:2d:00:16:00}',
	 '{2}'),
	('inetcol', 'inet',
	 '{=}',
	 '{10.2.14.231/24}',
	 '{1}'),
	('inetcol', 'cidr',
	 '{=}',
	 '{fe80::6e40:8ff:fea9:8c46}',
	 '{1}'),
	('cidrcol', 'inet',
	 '{=}',
	 '{10.2.14/24}',
	 '{2}'),
	('cidrcol', 'inet',
	 '{=}',
	 '{fe80::6e40:8ff:fea9:8c46}',
	 '{1}'),
	('cidrcol', 'cidr',
	 '{=}',
	 '{10.2.14/24}',
	 '{2}'),
	('cidrcol', 'cidr',
	 '{=}',
	 '{fe80::6e40:8ff:fea9:8c46}',
	 '{1}'),
	('bpcharcol', 'bpchar',
	 '{=}',
	 '{W}',
	 '{6}'),
	('datecol', 'date',
	 '{=}',
	 '{2009-12-01}',
	 '{1}'),
	('timecol', 'time',
	 '{=}',
	 '{02:28:57}',
	 '{1}'),
	('timestampcol', 'timestamp',
	 '{=}',
	 '{1964-03-24 19:26:45}',
	 '{1}'),
	('timestamptzcol', 'timestamptz',
	 '{=}',
	 '{1972-10-19 09:00:00-07}',
	 '{1}'),
	('intervalcol', 'interval',
	 '{=}',
	 '{1 mons 13 days 12:24}',
	 '{1}'),
	('timetzcol', 'timetz',
	 '{=}',
	 '{01:35:50+02}',
	 '{2}'),
	('numericcol', 'numeric',
	 '{=}',
	 '{2268164.347826086956521739130434782609}',
	 '{1}'),
	('uuidcol', 'uuid',
	 '{=}',
	 '{52225222-5222-5222-5222-522252225222}',
	 '{1}'),
	('lsncol', 'pg_lsn',
	 '{=, IS, IS NOT}',
	 '{44/455222, NULL, NULL}',
	 '{1, 25, 100}');`,
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
	FOR r IN SELECT colname, oper, typ, value[ordinality], matches[ordinality] FROM brinopers_bloom, unnest(op) WITH ORDINALITY AS oper LOOP
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
				Statement: `		FOR plan_line IN EXECUTE format($y$EXPLAIN SELECT array_agg(ctid) FROM brintest_bloom WHERE %s $y$, cond) LOOP
			IF plan_line LIKE '%Bitmap Heap Scan on brintest_bloom%' THEN
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
				Statement: `		EXECUTE format($y$SELECT array_agg(ctid) FROM brintest_bloom WHERE %s $y$, cond)
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
				Statement: `		FOR plan_line IN EXECUTE format($y$EXPLAIN SELECT array_agg(ctid) FROM brintest_bloom WHERE %s $y$, cond) LOOP
			IF plan_line LIKE '%Seq Scan on brintest_bloom%' THEN
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
				Statement: `		EXECUTE format($y$SELECT array_agg(ctid) FROM brintest_bloom WHERE %s $y$, cond)
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
				Statement: `			FOR r2 IN EXECUTE 'SELECT ' || r.colname || ' FROM brintest_bloom WHERE ' || cond LOOP
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
				Statement: `			FOR r2 IN EXECUTE 'SELECT ' || r.colname || ' FROM brintest_bloom WHERE ' || cond LOOP
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
				Statement: `INSERT INTO brintest_bloom SELECT
	repeat(stringu1, 42)::bytea,
	substr(stringu1, 1, 1)::"char",
	stringu1::name, 142857 * tenthous,
	thousand,
	twothousand,
	repeat(stringu1, 42),
	unique1::oid,
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
	tenthous::numeric(36,30) * fivethous * even / (hundred + 1),
	format('%s%s-%s-%s-%s-%s%s%s', to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'), to_char(tenthous, 'FM0000'))::uuid,
	format('%s/%s%s', odd, even, tenthous)::pg_lsn
FROM tenk1 ORDER BY unique2 LIMIT 5 OFFSET 5;`,
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx_bloom', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `VACUUM brintest_bloom;  -- force a summarization cycle in brinidx`,
			},
			{
				Statement: `UPDATE brintest_bloom SET int8col = int8col * int4col;`,
			},
			{
				Statement: `UPDATE brintest_bloom SET textcol = '' WHERE textcol IS NOT NULL;`,
			},
			{
				Statement:   `SELECT brin_summarize_new_values('brintest_bloom'); -- error, not an index`,
				ErrorString: `"brintest_bloom" is not an index`,
			},
			{
				Statement:   `SELECT brin_summarize_new_values('tenk1_unique1'); -- error, not a BRIN index`,
				ErrorString: `"tenk1_unique1" is not a BRIN index`,
			},
			{
				Statement: `SELECT brin_summarize_new_values('brinidx_bloom'); -- ok, no change expected`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT brin_desummarize_range('brinidx_bloom', -1); -- error, invalid range`,
				ErrorString: `block number out of range: -1`,
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx_bloom', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx_bloom', 0);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT brin_desummarize_range('brinidx_bloom', 100000000);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `CREATE TABLE brin_summarize_bloom (
    value int
) WITH (fillfactor=10, autovacuum_enabled=false);`,
			},
			{
				Statement: `CREATE INDEX brin_summarize_bloom_idx ON brin_summarize_bloom USING brin (value) WITH (pages_per_range=2);`,
			},
			{
				Statement: `DO $$
DECLARE curtid tid;`,
			},
			{
				Statement: `BEGIN
  LOOP
    INSERT INTO brin_summarize_bloom VALUES (1) RETURNING ctid INTO curtid;`,
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
				Statement: `SELECT brin_summarize_range('brin_summarize_bloom_idx', 0);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_bloom_idx', 1);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_bloom_idx', 2);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT brin_summarize_range('brin_summarize_bloom_idx', 4294967295);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT brin_summarize_range('brin_summarize_bloom_idx', -1);`,
				ErrorString: `block number out of range: -1`,
			},
			{
				Statement:   `SELECT brin_summarize_range('brin_summarize_bloom_idx', 4294967296);`,
				ErrorString: `block number out of range: 4294967296`,
			},
			{
				Statement: `CREATE TABLE brin_test_bloom (a INT, b INT);`,
			},
			{
				Statement: `INSERT INTO brin_test_bloom SELECT x/100,x%100 FROM generate_series(1,10000) x(x);`,
			},
			{
				Statement: `CREATE INDEX brin_test_bloom_a_idx ON brin_test_bloom USING brin (a) WITH (pages_per_range = 2);`,
			},
			{
				Statement: `CREATE INDEX brin_test_bloom_b_idx ON brin_test_bloom USING brin (b) WITH (pages_per_range = 2);`,
			},
			{
				Statement: `VACUUM ANALYZE brin_test_bloom;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM brin_test_bloom WHERE a = 1;`,
				Results:   []sql.Row{{`Bitmap Heap Scan on brin_test_bloom`}, {`Recheck Cond: (a = 1)`}, {`->  Bitmap Index Scan on brin_test_bloom_a_idx`}, {`Index Cond: (a = 1)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF) SELECT * FROM brin_test_bloom WHERE b = 1;`,
				Results:   []sql.Row{{`Seq Scan on brin_test_bloom`}, {`Filter: (b = 1)`}},
			},
		},
	})
}
