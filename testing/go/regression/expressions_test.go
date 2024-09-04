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

func TestExpressions(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_expressions)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_expressions,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT date(now())::text = current_date::text;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT now()::timetz::text = current_time::text;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT now()::timetz(4)::text = current_time(4)::text;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT now()::time::text = localtime::text;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT now()::time(3)::text = localtime(3)::text;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT current_timestamp = NOW();`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT length(current_timestamp::text) >= length(current_timestamp(0)::text);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT now()::timestamp::text = localtimestamp::text;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT current_catalog = current_database();`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT current_schema;`,
				Results:   []sql.Row{{`public`}},
			},
			{
				Statement: `SET search_path = 'notme';`,
			},
			{
				Statement: `SELECT current_schema;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SET search_path = 'pg_catalog';`,
			},
			{
				Statement: `SELECT current_schema;`,
				Results:   []sql.Row{{`pg_catalog`}},
			},
			{
				Statement: `RESET search_path;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create table numeric_tbl (f1 numeric(18,3), f2 numeric);`,
			},
			{
				Statement: `create view numeric_view as
  select
    f1, f1::numeric(16,4) as f1164, f1::numeric as f1n,
    f2, f2::numeric(16,4) as f2164, f2::numeric as f2n
  from numeric_tbl;`,
			},
			{
				Statement: `\d+ numeric_view
                           View "public.numeric_view"
 Column |     Type      | Collation | Nullable | Default | Storage | Description 
--------+---------------+-----------+----------+---------+---------+-------------
 f1     | numeric(18,3) |           |          |         | main    | 
 f1164  | numeric(16,4) |           |          |         | main    | 
 f1n    | numeric       |           |          |         | main    | 
 f2     | numeric       |           |          |         | main    | 
 f2164  | numeric(16,4) |           |          |         | main    | 
 f2n    | numeric       |           |          |         | main    | 
View definition:
 SELECT numeric_tbl.f1,
    numeric_tbl.f1::numeric(16,4) AS f1164,
    numeric_tbl.f1::numeric AS f1n,
    numeric_tbl.f2,
    numeric_tbl.f2::numeric(16,4) AS f2164,
    numeric_tbl.f2 AS f2n
   FROM numeric_tbl;`,
			},
			{
				Statement: `explain (verbose, costs off) select * from numeric_view;`,
				Results:   []sql.Row{{`Seq Scan on public.numeric_tbl`}, {`Output: numeric_tbl.f1, (numeric_tbl.f1)::numeric(16,4), (numeric_tbl.f1)::numeric, numeric_tbl.f2, (numeric_tbl.f2)::numeric(16,4), numeric_tbl.f2`}},
			},
			{
				Statement: `create table bpchar_tbl (f1 character(16) unique, f2 bpchar);`,
			},
			{
				Statement: `create view bpchar_view as
  select
    f1, f1::character(14) as f114, f1::bpchar as f1n,
    f2, f2::character(14) as f214, f2::bpchar as f2n
  from bpchar_tbl;`,
			},
			{
				Statement: `\d+ bpchar_view
                            View "public.bpchar_view"
 Column |     Type      | Collation | Nullable | Default | Storage  | Description 
--------+---------------+-----------+----------+---------+----------+-------------
 f1     | character(16) |           |          |         | extended | 
 f114   | character(14) |           |          |         | extended | 
 f1n    | bpchar        |           |          |         | extended | 
 f2     | bpchar        |           |          |         | extended | 
 f214   | character(14) |           |          |         | extended | 
 f2n    | bpchar        |           |          |         | extended | 
View definition:
 SELECT bpchar_tbl.f1,
    bpchar_tbl.f1::character(14) AS f114,
    bpchar_tbl.f1::bpchar AS f1n,
    bpchar_tbl.f2,
    bpchar_tbl.f2::character(14) AS f214,
    bpchar_tbl.f2 AS f2n
   FROM bpchar_tbl;`,
			},
			{
				Statement: `explain (verbose, costs off) select * from bpchar_view
  where f1::bpchar = 'foo';`,
				Results: []sql.Row{{`Index Scan using bpchar_tbl_f1_key on public.bpchar_tbl`}, {`Output: bpchar_tbl.f1, (bpchar_tbl.f1)::character(14), (bpchar_tbl.f1)::bpchar, bpchar_tbl.f2, (bpchar_tbl.f2)::character(14), bpchar_tbl.f2`}, {`Index Cond: ((bpchar_tbl.f1)::bpchar = 'foo'::bpchar)`}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `explain (verbose, costs off)
select random() IN (1, 4, 8.0);`,
				Results: []sql.Row{{`Result`}, {`Output: (random() = ANY ('{1,4,8}'::double precision[]))`}},
			},
			{
				Statement: `explain (verbose, costs off)
select random()::int IN (1, 4, 8.0);`,
				Results: []sql.Row{{`Result`}, {`Output: (((random())::integer)::numeric = ANY ('{1,4,8.0}'::numeric[]))`}},
			},
			{
				Statement:   `select '(0,0)'::point in ('(0,0,0,0)'::box, point(0,0));`,
				ErrorString: `operator does not exist: point = box`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create function return_int_input(int) returns int as $$
begin
	return $1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql stable;`,
			},
			{
				Statement: `create function return_text_input(text) returns text as $$
begin
	return $1;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql stable;`,
			},
			{
				Statement: `select return_int_input(1) in (10, 9, 2, 8, 3, 7, 4, 6, 5, 1);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select return_int_input(1) in (10, 9, 2, 8, 3, 7, 4, 6, 5, null);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select return_int_input(1) in (null, null, null, null, null, null, null, null, null, null, null);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select return_int_input(1) in (10, 9, 2, 8, 3, 7, 4, 6, 5, 1, null);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select return_int_input(null::int) in (10, 9, 2, 8, 3, 7, 4, 6, 5, 1);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select return_int_input(null::int) in (10, 9, 2, 8, 3, 7, 4, 6, 5, null);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select return_text_input('a') in ('a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j');`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select return_int_input(1) not in (10, 9, 2, 8, 3, 7, 4, 6, 5, 1);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select return_int_input(1) not in (10, 9, 2, 8, 3, 7, 4, 6, 5, 0);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select return_int_input(1) not in (10, 9, 2, 8, 3, 7, 4, 6, 5, 2, null);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select return_int_input(1) not in (10, 9, 2, 8, 3, 7, 4, 6, 5, 1, null);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select return_int_input(1) not in (null, null, null, null, null, null, null, null, null, null, null);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select return_int_input(null::int) not in (10, 9, 2, 8, 3, 7, 4, 6, 5, 1);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select return_int_input(null::int) not in (10, 9, 2, 8, 3, 7, 4, 6, 5, null);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select return_text_input('a') not in ('a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j');`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `rollback;`,
			},
			{
				Statement: `begin;`,
			},
			{
				Statement: `create type myint;`,
			},
			{
				Statement: `create function myintin(cstring) returns myint strict immutable language
  internal as 'int4in';`,
			},
			{
				Statement: `create function myintout(myint) returns cstring strict immutable language
  internal as 'int4out';`,
			},
			{
				Statement: `create function myinthash(myint) returns integer strict immutable language
  internal as 'hashint4';`,
			},
			{
				Statement: `create type myint (input = myintin, output = myintout, like = int4);`,
			},
			{
				Statement: `create cast (int4 as myint) without function;`,
			},
			{
				Statement: `create cast (myint as int4) without function;`,
			},
			{
				Statement: `create function myinteq(myint, myint) returns bool as $$
begin
  if $1 is null and $2 is null then
    return true;`,
			},
			{
				Statement: `  else
    return $1::int = $2::int;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql immutable;`,
			},
			{
				Statement: `create function myintne(myint, myint) returns bool as $$
begin
  return not myinteq($1, $2);`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ language plpgsql immutable;`,
			},
			{
				Statement: `create operator = (
  leftarg    = myint,
  rightarg   = myint,
  commutator = =,
  negator    = <>,
  procedure  = myinteq,
  restrict   = eqsel,
  join       = eqjoinsel,
  merges
);`,
			},
			{
				Statement: `create operator <> (
  leftarg    = myint,
  rightarg   = myint,
  commutator = <>,
  negator    = =,
  procedure  = myintne,
  restrict   = eqsel,
  join       = eqjoinsel,
  merges
);`,
			},
			{
				Statement: `create operator class myint_ops
default for type myint using hash as
  operator    1   =  (myint, myint),
  function    1   myinthash(myint);`,
			},
			{
				Statement: `create table inttest (a myint);`,
			},
			{
				Statement: `insert into inttest values(1::myint),(null);`,
			},
			{
				Statement: `select * from inttest where a in (1::myint,2::myint,3::myint,4::myint,5::myint,6::myint,7::myint,8::myint,9::myint, null);`,
				Results:   []sql.Row{{1}, {``}},
			},
			{
				Statement: `select * from inttest where a not in (1::myint,2::myint,3::myint,4::myint,5::myint,6::myint,7::myint,8::myint,9::myint, null);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from inttest where a not in (0::myint,2::myint,3::myint,4::myint,5::myint,6::myint,7::myint,8::myint,9::myint, null);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from inttest where a in (1::myint,2::myint,3::myint,4::myint,5::myint, null);`,
				Results:   []sql.Row{{1}, {``}},
			},
			{
				Statement: `select * from inttest where a not in (1::myint,2::myint,3::myint,4::myint,5::myint, null);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select * from inttest where a not in (0::myint,2::myint,3::myint,4::myint,5::myint, null);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `rollback;`,
			},
		},
	})
}
