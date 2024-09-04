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

func TestEquivclass(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_equivclass)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_equivclass,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create type int8alias1;`,
			},
			{
				Statement: `create function int8alias1in(cstring) returns int8alias1
  strict immutable language internal as 'int8in';`,
			},
			{
				Statement: `create function int8alias1out(int8alias1) returns cstring
  strict immutable language internal as 'int8out';`,
			},
			{
				Statement: `create type int8alias1 (
    input = int8alias1in,
    output = int8alias1out,
    like = int8
);`,
			},
			{
				Statement: `create type int8alias2;`,
			},
			{
				Statement: `create function int8alias2in(cstring) returns int8alias2
  strict immutable language internal as 'int8in';`,
			},
			{
				Statement: `create function int8alias2out(int8alias2) returns cstring
  strict immutable language internal as 'int8out';`,
			},
			{
				Statement: `create type int8alias2 (
    input = int8alias2in,
    output = int8alias2out,
    like = int8
);`,
			},
			{
				Statement: `create cast (int8 as int8alias1) without function;`,
			},
			{
				Statement: `create cast (int8 as int8alias2) without function;`,
			},
			{
				Statement: `create cast (int8alias1 as int8) without function;`,
			},
			{
				Statement: `create cast (int8alias2 as int8) without function;`,
			},
			{
				Statement: `create function int8alias1eq(int8alias1, int8alias1) returns bool
  strict immutable language internal as 'int8eq';`,
			},
			{
				Statement: `create operator = (
    procedure = int8alias1eq,
    leftarg = int8alias1, rightarg = int8alias1,
    commutator = =,
    restrict = eqsel, join = eqjoinsel,
    merges
);`,
			},
			{
				Statement: `alter operator family integer_ops using btree add
  operator 3 = (int8alias1, int8alias1);`,
			},
			{
				Statement: `create function int8alias2eq(int8alias2, int8alias2) returns bool
  strict immutable language internal as 'int8eq';`,
			},
			{
				Statement: `create operator = (
    procedure = int8alias2eq,
    leftarg = int8alias2, rightarg = int8alias2,
    commutator = =,
    restrict = eqsel, join = eqjoinsel,
    merges
);`,
			},
			{
				Statement: `alter operator family integer_ops using btree add
  operator 3 = (int8alias2, int8alias2);`,
			},
			{
				Statement: `create function int8alias1eq(int8, int8alias1) returns bool
  strict immutable language internal as 'int8eq';`,
			},
			{
				Statement: `create operator = (
    procedure = int8alias1eq,
    leftarg = int8, rightarg = int8alias1,
    restrict = eqsel, join = eqjoinsel,
    merges
);`,
			},
			{
				Statement: `alter operator family integer_ops using btree add
  operator 3 = (int8, int8alias1);`,
			},
			{
				Statement: `create function int8alias1eq(int8alias1, int8alias2) returns bool
  strict immutable language internal as 'int8eq';`,
			},
			{
				Statement: `create operator = (
    procedure = int8alias1eq,
    leftarg = int8alias1, rightarg = int8alias2,
    restrict = eqsel, join = eqjoinsel,
    merges
);`,
			},
			{
				Statement: `alter operator family integer_ops using btree add
  operator 3 = (int8alias1, int8alias2);`,
			},
			{
				Statement: `create function int8alias1lt(int8alias1, int8alias1) returns bool
  strict immutable language internal as 'int8lt';`,
			},
			{
				Statement: `create operator < (
    procedure = int8alias1lt,
    leftarg = int8alias1, rightarg = int8alias1
);`,
			},
			{
				Statement: `alter operator family integer_ops using btree add
  operator 1 < (int8alias1, int8alias1);`,
			},
			{
				Statement: `create function int8alias1cmp(int8, int8alias1) returns int
  strict immutable language internal as 'btint8cmp';`,
			},
			{
				Statement: `alter operator family integer_ops using btree add
  function 1 int8alias1cmp (int8, int8alias1);`,
			},
			{
				Statement: `create table ec0 (ff int8 primary key, f1 int8, f2 int8);`,
			},
			{
				Statement: `create table ec1 (ff int8 primary key, f1 int8alias1, f2 int8alias2);`,
			},
			{
				Statement: `create table ec2 (xf int8 primary key, x1 int8alias1, x2 int8alias2);`,
			},
			{
				Statement: `set enable_hashjoin = off;`,
			},
			{
				Statement: `set enable_mergejoin = off;`,
			},
			{
				Statement: `explain (costs off)
  select * from ec0 where ff = f1 and f1 = '42'::int8;`,
				Results: []sql.Row{{`Index Scan using ec0_pkey on ec0`}, {`Index Cond: (ff = '42'::bigint)`}, {`Filter: (f1 = '42'::bigint)`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec0 where ff = f1 and f1 = '42'::int8alias1;`,
				Results: []sql.Row{{`Index Scan using ec0_pkey on ec0`}, {`Index Cond: (ff = '42'::int8alias1)`}, {`Filter: (f1 = '42'::int8alias1)`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1 where ff = f1 and f1 = '42'::int8alias1;`,
				Results: []sql.Row{{`Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::int8alias1)`}, {`Filter: (f1 = '42'::int8alias1)`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1 where ff = f1 and f1 = '42'::int8alias2;`,
				Results: []sql.Row{{`Seq Scan on ec1`}, {`Filter: ((ff = f1) AND (f1 = '42'::int8alias2))`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1, ec2 where ff = x1 and ff = '42'::int8;`,
				Results: []sql.Row{{`Nested Loop`}, {`Join Filter: (ec1.ff = ec2.x1)`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: ((ff = '42'::bigint) AND (ff = '42'::bigint))`}, {`->  Seq Scan on ec2`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1, ec2 where ff = x1 and ff = '42'::int8alias1;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::int8alias1)`}, {`->  Seq Scan on ec2`}, {`Filter: (x1 = '42'::int8alias1)`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1, ec2 where ff = x1 and '42'::int8 = x1;`,
				Results: []sql.Row{{`Nested Loop`}, {`Join Filter: (ec1.ff = ec2.x1)`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::bigint)`}, {`->  Seq Scan on ec2`}, {`Filter: ('42'::bigint = x1)`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1, ec2 where ff = x1 and x1 = '42'::int8alias1;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::int8alias1)`}, {`->  Seq Scan on ec2`}, {`Filter: (x1 = '42'::int8alias1)`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1, ec2 where ff = x1 and x1 = '42'::int8alias2;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Seq Scan on ec2`}, {`Filter: (x1 = '42'::int8alias2)`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = ec2.x1)`}},
			},
			{
				Statement: `create unique index ec1_expr1 on ec1((ff + 1));`,
			},
			{
				Statement: `create unique index ec1_expr2 on ec1((ff + 2 + 1));`,
			},
			{
				Statement: `create unique index ec1_expr3 on ec1((ff + 3 + 1));`,
			},
			{
				Statement: `create unique index ec1_expr4 on ec1((ff + 4));`,
			},
			{
				Statement: `explain (costs off)
  select * from ec1,
    (select ff + 1 as x from
       (select ff + 2 as ff from ec1
        union all
        select ff + 3 as ff from ec1) ss0
     union all
     select ff + 4 as x from ec1) as ss1
  where ss1.x = ec1.f1 and ec1.ff = 42::int8;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::bigint)`}, {`->  Append`}, {`->  Index Scan using ec1_expr2 on ec1 ec1_1`}, {`Index Cond: (((ff + 2) + 1) = ec1.f1)`}, {`->  Index Scan using ec1_expr3 on ec1 ec1_2`}, {`Index Cond: (((ff + 3) + 1) = ec1.f1)`}, {`->  Index Scan using ec1_expr4 on ec1 ec1_3`}, {`Index Cond: ((ff + 4) = ec1.f1)`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1,
    (select ff + 1 as x from
       (select ff + 2 as ff from ec1
        union all
        select ff + 3 as ff from ec1) ss0
     union all
     select ff + 4 as x from ec1) as ss1
  where ss1.x = ec1.f1 and ec1.ff = 42::int8 and ec1.ff = ec1.f1;`,
				Results: []sql.Row{{`Nested Loop`}, {`Join Filter: ((((ec1_1.ff + 2) + 1)) = ec1.f1)`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: ((ff = '42'::bigint) AND (ff = '42'::bigint))`}, {`Filter: (ff = f1)`}, {`->  Append`}, {`->  Index Scan using ec1_expr2 on ec1 ec1_1`}, {`Index Cond: (((ff + 2) + 1) = '42'::bigint)`}, {`->  Index Scan using ec1_expr3 on ec1 ec1_2`}, {`Index Cond: (((ff + 3) + 1) = '42'::bigint)`}, {`->  Index Scan using ec1_expr4 on ec1 ec1_3`}, {`Index Cond: ((ff + 4) = '42'::bigint)`}},
			},
			{
				Statement: `explain (costs off)
  select * from ec1,
    (select ff + 1 as x from
       (select ff + 2 as ff from ec1
        union all
        select ff + 3 as ff from ec1) ss0
     union all
     select ff + 4 as x from ec1) as ss1,
    (select ff + 1 as x from
       (select ff + 2 as ff from ec1
        union all
        select ff + 3 as ff from ec1) ss0
     union all
     select ff + 4 as x from ec1) as ss2
  where ss1.x = ec1.f1 and ss1.x = ss2.x and ec1.ff = 42::int8;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Nested Loop`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::bigint)`}, {`->  Append`}, {`->  Index Scan using ec1_expr2 on ec1 ec1_1`}, {`Index Cond: (((ff + 2) + 1) = ec1.f1)`}, {`->  Index Scan using ec1_expr3 on ec1 ec1_2`}, {`Index Cond: (((ff + 3) + 1) = ec1.f1)`}, {`->  Index Scan using ec1_expr4 on ec1 ec1_3`}, {`Index Cond: ((ff + 4) = ec1.f1)`}, {`->  Append`}, {`->  Index Scan using ec1_expr2 on ec1 ec1_4`}, {`Index Cond: (((ff + 2) + 1) = (((ec1_1.ff + 2) + 1)))`}, {`->  Index Scan using ec1_expr3 on ec1 ec1_5`}, {`Index Cond: (((ff + 3) + 1) = (((ec1_1.ff + 2) + 1)))`}, {`->  Index Scan using ec1_expr4 on ec1 ec1_6`}, {`Index Cond: ((ff + 4) = (((ec1_1.ff + 2) + 1)))`}},
			},
			{
				Statement: `set enable_mergejoin = on;`,
			},
			{
				Statement: `set enable_nestloop = off;`,
			},
			{
				Statement: `explain (costs off)
  select * from ec1,
    (select ff + 1 as x from
       (select ff + 2 as ff from ec1
        union all
        select ff + 3 as ff from ec1) ss0
     union all
     select ff + 4 as x from ec1) as ss1,
    (select ff + 1 as x from
       (select ff + 2 as ff from ec1
        union all
        select ff + 3 as ff from ec1) ss0
     union all
     select ff + 4 as x from ec1) as ss2
  where ss1.x = ec1.f1 and ss1.x = ss2.x and ec1.ff = 42::int8;`,
				Results: []sql.Row{{`Merge Join`}, {`Merge Cond: ((((ec1_4.ff + 2) + 1)) = (((ec1_1.ff + 2) + 1)))`}, {`->  Merge Append`}, {`Sort Key: (((ec1_4.ff + 2) + 1))`}, {`->  Index Scan using ec1_expr2 on ec1 ec1_4`}, {`->  Index Scan using ec1_expr3 on ec1 ec1_5`}, {`->  Index Scan using ec1_expr4 on ec1 ec1_6`}, {`->  Materialize`}, {`->  Merge Join`}, {`Merge Cond: ((((ec1_1.ff + 2) + 1)) = ec1.f1)`}, {`->  Merge Append`}, {`Sort Key: (((ec1_1.ff + 2) + 1))`}, {`->  Index Scan using ec1_expr2 on ec1 ec1_1`}, {`->  Index Scan using ec1_expr3 on ec1 ec1_2`}, {`->  Index Scan using ec1_expr4 on ec1 ec1_3`}, {`->  Sort`}, {`Sort Key: ec1.f1 USING <`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::bigint)`}},
			},
			{
				Statement: `set enable_nestloop = on;`,
			},
			{
				Statement: `set enable_mergejoin = off;`,
			},
			{
				Statement: `drop index ec1_expr3;`,
			},
			{
				Statement: `explain (costs off)
  select * from ec1,
    (select ff + 1 as x from
       (select ff + 2 as ff from ec1
        union all
        select ff + 3 as ff from ec1) ss0
     union all
     select ff + 4 as x from ec1) as ss1
  where ss1.x = ec1.f1 and ec1.ff = 42::int8;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::bigint)`}, {`->  Append`}, {`->  Index Scan using ec1_expr2 on ec1 ec1_1`}, {`Index Cond: (((ff + 2) + 1) = ec1.f1)`}, {`->  Seq Scan on ec1 ec1_2`}, {`Filter: (((ff + 3) + 1) = ec1.f1)`}, {`->  Index Scan using ec1_expr4 on ec1 ec1_3`}, {`Index Cond: ((ff + 4) = ec1.f1)`}},
			},
			{
				Statement: `set enable_mergejoin = on;`,
			},
			{
				Statement: `set enable_nestloop = off;`,
			},
			{
				Statement: `explain (costs off)
  select * from ec1,
    (select ff + 1 as x from
       (select ff + 2 as ff from ec1
        union all
        select ff + 3 as ff from ec1) ss0
     union all
     select ff + 4 as x from ec1) as ss1
  where ss1.x = ec1.f1 and ec1.ff = 42::int8;`,
				Results: []sql.Row{{`Merge Join`}, {`Merge Cond: ((((ec1_1.ff + 2) + 1)) = ec1.f1)`}, {`->  Merge Append`}, {`Sort Key: (((ec1_1.ff + 2) + 1))`}, {`->  Index Scan using ec1_expr2 on ec1 ec1_1`}, {`->  Sort`}, {`Sort Key: (((ec1_2.ff + 3) + 1))`}, {`->  Seq Scan on ec1 ec1_2`}, {`->  Index Scan using ec1_expr4 on ec1 ec1_3`}, {`->  Sort`}, {`Sort Key: ec1.f1 USING <`}, {`->  Index Scan using ec1_pkey on ec1`}, {`Index Cond: (ff = '42'::bigint)`}},
			},
			{
				Statement: `set enable_nestloop = on;`,
			},
			{
				Statement: `set enable_mergejoin = off;`,
			},
			{
				Statement: `alter table ec1 enable row level security;`,
			},
			{
				Statement: `create policy p1 on ec1 using (f1 < '5'::int8alias1);`,
			},
			{
				Statement: `create user regress_user_ectest;`,
			},
			{
				Statement: `grant select on ec0 to regress_user_ectest;`,
			},
			{
				Statement: `grant select on ec1 to regress_user_ectest;`,
			},
			{
				Statement: `explain (costs off)
  select * from ec0 a, ec1 b
  where a.ff = b.ff and a.ff = 43::bigint::int8alias1;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Index Scan using ec0_pkey on ec0 a`}, {`Index Cond: (ff = '43'::int8alias1)`}, {`->  Index Scan using ec1_pkey on ec1 b`}, {`Index Cond: (ff = '43'::int8alias1)`}},
			},
			{
				Statement: `set session authorization regress_user_ectest;`,
			},
			{
				Statement: `explain (costs off)
  select * from ec0 a, ec1 b
  where a.ff = b.ff and a.ff = 43::bigint::int8alias1;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Index Scan using ec0_pkey on ec0 a`}, {`Index Cond: (ff = '43'::int8alias1)`}, {`->  Index Scan using ec1_pkey on ec1 b`}, {`Index Cond: (ff = a.ff)`}, {`Filter: (f1 < '5'::int8alias1)`}},
			},
			{
				Statement: `reset session authorization;`,
			},
			{
				Statement: `revoke select on ec0 from regress_user_ectest;`,
			},
			{
				Statement: `revoke select on ec1 from regress_user_ectest;`,
			},
			{
				Statement: `drop user regress_user_ectest;`,
			},
			{
				Statement: `explain (costs off)
  select * from tenk1 where unique1 = unique1 and unique2 = unique2;`,
				Results: []sql.Row{{`Seq Scan on tenk1`}, {`Filter: ((unique1 IS NOT NULL) AND (unique2 IS NOT NULL))`}},
			},
			{
				Statement: `explain (costs off)
  select * from tenk1 where unique1 = unique1 or unique2 = unique2;`,
				Results: []sql.Row{{`Seq Scan on tenk1`}, {`Filter: ((unique1 = unique1) OR (unique2 = unique2))`}},
			},
			{
				Statement: `create temp table undername (f1 name, f2 int);`,
			},
			{
				Statement: `create temp view overview as
  select f1::information_schema.sql_identifier as sqli, f2 from undername;`,
			},
			{
				Statement: `explain (costs off)  -- this should not require a sort
  select * from overview where sqli = 'foo' order by sqli;`,
				Results: []sql.Row{{`Seq Scan on undername`}, {`Filter: (f1 = 'foo'::name)`}},
			},
		},
	})
}
