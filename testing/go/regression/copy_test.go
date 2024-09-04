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

func TestCopy(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_copy)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_copy,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv abs_srcdir PG_ABS_SRCDIR
\getenv abs_builddir PG_ABS_BUILDDIR
create temp table copytest (
	style	text,
	test 	text,
	filler	int);`,
			},
			{
				Statement: `insert into copytest values('DOS',E'abc\r\ndef',1);`,
			},
			{
				Statement: `insert into copytest values('Unix',E'abc\ndef',2);`,
			},
			{
				Statement: `insert into copytest values('Mac',E'abc\rdef',3);`,
			},
			{
				Statement: `insert into copytest values(E'esc\\ape',E'a\\r\\\r\\\n\\nb',4);`,
			},
			{
				Statement: `\set filename :abs_builddir '/results/copytest.csv'
copy copytest to :'filename' csv;`,
			},
			{
				Statement: `create temp table copytest2 (like copytest);`,
			},
			{
				Statement: `copy copytest2 from :'filename' csv;`,
			},
			{
				Statement: `select * from copytest except select * from copytest2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `truncate copytest2;`,
			},
			{
				Statement: `copy copytest to :'filename' csv quote '''' escape E'\\';`,
			},
			{
				Statement: `copy copytest2 from :'filename' csv quote '''' escape E'\\';`,
			},
			{
				Statement: `select * from copytest except select * from copytest2;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `create temp table copytest3 (
	c1 int,
	"col with , comma" text,
	"col with "" quote"  int);`,
			},
			{
				Statement: `copy copytest3 from stdin csv header;`,
			},
			{
				Statement: `copy copytest3 to stdout csv header;`,
			},
			{
				Statement: `c1,"col with , comma","col with "" quote"
1,a,1
2,b,2
create temp table copytest4 (
	c1 int,
	"colname with tab: 	" text);`,
			},
			{
				Statement: `copy copytest4 from stdin (header);`,
			},
			{
				Statement: `copy copytest4 to stdout (header);`,
			},
			{
				Statement: `c1	colname with tab: \t
1	a
2	b
create table parted_copytest (
	a int,
	b int,
	c text
) partition by list (b);`,
			},
			{
				Statement: `create table parted_copytest_a1 (c text, b int, a int);`,
			},
			{
				Statement: `create table parted_copytest_a2 (a int, c text, b int);`,
			},
			{
				Statement: `alter table parted_copytest attach partition parted_copytest_a1 for values in(1);`,
			},
			{
				Statement: `alter table parted_copytest attach partition parted_copytest_a2 for values in(2);`,
			},
			{
				Statement: `insert into parted_copytest select x,1,'One' from generate_series(1,1000) x;`,
			},
			{
				Statement: `insert into parted_copytest select x,2,'Two' from generate_series(1001,1010) x;`,
			},
			{
				Statement: `insert into parted_copytest select x,1,'One' from generate_series(1011,1020) x;`,
			},
			{
				Statement: `\set filename :abs_builddir '/results/parted_copytest.csv'
copy (select * from parted_copytest order by a) to :'filename';`,
			},
			{
				Statement: `truncate parted_copytest;`,
			},
			{
				Statement: `copy parted_copytest from :'filename';`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `truncate parted_copytest;`,
			},
			{
				Statement:   `copy parted_copytest from :'filename' (freeze);`,
				ErrorString: `cannot perform COPY FREEZE on a partitioned table`,
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `select tableoid::regclass,count(*),sum(a) from parted_copytest
group by tableoid order by tableoid::regclass::name;`,
				Results: []sql.Row{{`parted_copytest_a1`, 1010, 510655}, {`parted_copytest_a2`, 10, 10055}},
			},
			{
				Statement: `truncate parted_copytest;`,
			},
			{
				Statement: `create function part_ins_func() returns trigger language plpgsql as $$
begin
  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `create trigger part_ins_trig
	before insert on parted_copytest_a2
	for each row
	execute procedure part_ins_func();`,
			},
			{
				Statement: `copy parted_copytest from :'filename';`,
			},
			{
				Statement: `select tableoid::regclass,count(*),sum(a) from parted_copytest
group by tableoid order by tableoid::regclass::name;`,
				Results: []sql.Row{{`parted_copytest_a1`, 1010, 510655}, {`parted_copytest_a2`, 10, 10055}},
			},
			{
				Statement: `truncate table parted_copytest;`,
			},
			{
				Statement: `create index on parted_copytest (b);`,
			},
			{
				Statement: `drop trigger part_ins_trig on parted_copytest_a2;`,
			},
			{
				Statement: `copy parted_copytest from stdin;`,
			},
			{
				Statement: `select * from parted_copytest where b = 1;`,
				Results:   []sql.Row{{1, 1, `str1`}},
			},
			{
				Statement: `select * from parted_copytest where b = 2;`,
				Results:   []sql.Row{{2, 2, `str2`}},
			},
			{
				Statement: `drop table parted_copytest;`,
			},
			{
				Statement: `create table tab_progress_reporting (
	name text,
	age int4,
	location point,
	salary int4,
	manager name
);`,
			},
			{
				Statement: `create function notice_after_tab_progress_reporting() returns trigger AS
$$
declare report record;`,
			},
			{
				Statement: `begin
  -- The fields ignored here are the ones that may not remain
  -- consistent across multiple runs.  The sizes reported may differ
  -- across platforms, so just check if these are strictly positive.
  with progress_data as (
    select
       relid::regclass::text as relname,
       command,
       type,
       bytes_processed > 0 as has_bytes_processed,
       bytes_total > 0 as has_bytes_total,
       tuples_processed,
       tuples_excluded
      from pg_stat_progress_copy
      where pid = pg_backend_pid())
  select into report (to_jsonb(r)) as value
    from progress_data r;`,
			},
			{
				Statement: `  raise info 'progress: %', report.value::text;`,
			},
			{
				Statement: `  return new;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `create trigger check_after_tab_progress_reporting
	after insert on tab_progress_reporting
	for each statement
	execute function notice_after_tab_progress_reporting();`,
			},
			{
				Statement: `copy tab_progress_reporting from stdin;`,
			},
			{
				Statement: `INFO:  progress: {"type": "PIPE", "command": "COPY FROM", "relname": "tab_progress_reporting", "has_bytes_total": false, "tuples_excluded": 0, "tuples_processed": 3, "has_bytes_processed": true}
truncate tab_progress_reporting;`,
			},
			{
				Statement: `\set filename :abs_srcdir '/data/emp.data'
copy tab_progress_reporting from :'filename'
	where (salary < 2000);`,
			},
			{
				Statement: `INFO:  progress: {"type": "FILE", "command": "COPY FROM", "relname": "tab_progress_reporting", "has_bytes_total": true, "tuples_excluded": 1, "tuples_processed": 2, "has_bytes_processed": true}
drop trigger check_after_tab_progress_reporting on tab_progress_reporting;`,
			},
			{
				Statement: `drop function notice_after_tab_progress_reporting();`,
			},
			{
				Statement: `drop table tab_progress_reporting;`,
			},
			{
				Statement: `create table header_copytest (
	a int,
	b int,
	c text
);`,
			},
			{
				Statement: `alter table header_copytest drop column c;`,
			},
			{
				Statement: `alter table header_copytest add column c text;`,
			},
			{
				Statement:   `copy header_copytest to stdout with (header match);`,
				ErrorString: `cannot use "match" with HEADER in COPY TO`,
			},
			{
				Statement:   `copy header_copytest from stdin with (header wrong_choice);`,
				ErrorString: `header requires a Boolean value or "match"`,
			},
			{
				Statement: `copy header_copytest from stdin with (header match);`,
			},
			{
				Statement: `copy header_copytest (c, a, b) from stdin with (header match);`,
			},
			{
				Statement: `copy header_copytest from stdin with (header match, format csv);`,
			},
			{
				Statement:   `copy header_copytest (c, b, a) from stdin with (header match);`,
				ErrorString: `column name mismatch in header line field 1: got "a", expected "c"`,
			},
			{
				Statement: `CONTEXT:  COPY header_copytest, line 1: "a	b	c"
copy header_copytest from stdin with (header match);`,
				ErrorString: `column name mismatch in header line field 3: got null value ("\N"), expected "c"`,
			},
			{
				Statement: `CONTEXT:  COPY header_copytest, line 1: "a	b	\N"
copy header_copytest from stdin with (header match);`,
				ErrorString: `wrong number of fields in header line: got 2, expected 3`,
			},
			{
				Statement: `CONTEXT:  COPY header_copytest, line 1: "a	b"
copy header_copytest from stdin with (header match);`,
				ErrorString: `wrong number of fields in header line: got 4, expected 3`,
			},
			{
				Statement: `CONTEXT:  COPY header_copytest, line 1: "a	b	c	d"
copy header_copytest from stdin with (header match);`,
				ErrorString: `column name mismatch in header line field 3: got "d", expected "c"`,
			},
			{
				Statement: `CONTEXT:  COPY header_copytest, line 1: "a	b	d"
SELECT * FROM header_copytest ORDER BY a;`,
				Results: []sql.Row{{1, 2, `foo`}, {3, 4, `bar`}, {5, 6, `baz`}},
			},
			{
				Statement: `alter table header_copytest drop column b;`,
			},
			{
				Statement: `copy header_copytest (c, a) from stdin with (header match);`,
			},
			{
				Statement: `copy header_copytest (a, c) from stdin with (header match);`,
			},
			{
				Statement:   `copy header_copytest from stdin with (header match);`,
				ErrorString: `wrong number of fields in header line: got 3, expected 2`,
			},
			{
				Statement: `CONTEXT:  COPY header_copytest, line 1: "a	........pg.dropped.2........	c"
copy header_copytest (a, c) from stdin with (header match);`,
				ErrorString: `wrong number of fields in header line: got 3, expected 2`,
			},
			{
				Statement: `CONTEXT:  COPY header_copytest, line 1: "a	c	b"
SELECT * FROM header_copytest ORDER BY a;`,
				Results: []sql.Row{{1, `foo`}, {3, `bar`}, {5, `baz`}, {7, `foo`}, {8, `foo`}},
			},
			{
				Statement: `drop table header_copytest;`,
			},
		},
	})
}
