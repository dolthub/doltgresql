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

func TestCreateAggregate(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_aggregate)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_aggregate,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE AGGREGATE newavg (
   sfunc = int4_avg_accum, basetype = int4, stype = _int8,
   finalfunc = int8_avg,
   initcond1 = '{0,0}'
);`,
			},
			{
				Statement:   `COMMENT ON AGGREGATE newavg_wrong (int4) IS 'an agg comment';`,
				ErrorString: `aggregate newavg_wrong(integer) does not exist`,
			},
			{
				Statement: `COMMENT ON AGGREGATE newavg (int4) IS 'an agg comment';`,
			},
			{
				Statement: `COMMENT ON AGGREGATE newavg (int4) IS NULL;`,
			},
			{
				Statement: `CREATE AGGREGATE newsum (
   sfunc1 = int4pl, basetype = int4, stype1 = int4,
   initcond1 = '0'
);`,
			},
			{
				Statement: `CREATE AGGREGATE newcnt (*) (
   sfunc = int8inc, stype = int8,
   initcond = '0', parallel = safe
);`,
			},
			{
				Statement: `CREATE AGGREGATE oldcnt (
   sfunc = int8inc, basetype = 'ANY', stype = int8,
   initcond = '0'
);`,
			},
			{
				Statement: `CREATE AGGREGATE newcnt ("any") (
   sfunc = int8inc_any, stype = int8,
   initcond = '0'
);`,
			},
			{
				Statement:   `COMMENT ON AGGREGATE nosuchagg (*) IS 'should fail';`,
				ErrorString: `aggregate nosuchagg(*) does not exist`,
			},
			{
				Statement: `COMMENT ON AGGREGATE newcnt (*) IS 'an agg(*) comment';`,
			},
			{
				Statement: `COMMENT ON AGGREGATE newcnt ("any") IS 'an agg(any) comment';`,
			},
			{
				Statement: `create function sum3(int8,int8,int8) returns int8 as
'select $1 + $2 + $3' language sql strict immutable;`,
			},
			{
				Statement: `create aggregate sum2(int8,int8) (
   sfunc = sum3, stype = int8,
   initcond = '0'
);`,
			},
			{
				Statement: `create type aggtype as (a integer, b integer, c text);`,
			},
			{
				Statement: `create function aggf_trans(aggtype[],integer,integer,text) returns aggtype[]
as 'select array_append($1,ROW($2,$3,$4)::aggtype)'
language sql strict immutable;`,
			},
			{
				Statement: `create function aggfns_trans(aggtype[],integer,integer,text) returns aggtype[]
as 'select array_append($1,ROW($2,$3,$4)::aggtype)'
language sql immutable;`,
			},
			{
				Statement: `create aggregate aggfstr(integer,integer,text) (
   sfunc = aggf_trans, stype = aggtype[],
   initcond = '{}'
);`,
			},
			{
				Statement: `create aggregate aggfns(integer,integer,text) (
   sfunc = aggfns_trans, stype = aggtype[], sspace = 10000,
   initcond = '{}'
);`,
			},
			{
				Statement: `create function least_accum(int8, int8) returns int8 language sql as
  'select least($1, $2)';`,
			},
			{
				Statement: `create aggregate least_agg(int4) (
  stype = int8, sfunc = least_accum
);  -- fails`,
				ErrorString: `function least_accum(bigint, bigint) requires run-time type coercion`,
			},
			{
				Statement: `drop function least_accum(int8, int8);`,
			},
			{
				Statement: `create function least_accum(anycompatible, anycompatible)
returns anycompatible language sql as
  'select least($1, $2)';`,
			},
			{
				Statement: `create aggregate least_agg(int4) (
  stype = int8, sfunc = least_accum
);  -- fails`,
				ErrorString: `function least_accum(bigint, bigint) requires run-time type coercion`,
			},
			{
				Statement: `create aggregate least_agg(int8) (
  stype = int8, sfunc = least_accum
);`,
			},
			{
				Statement: `drop function least_accum(anycompatible, anycompatible) cascade;`,
			},
			{
				Statement: `create function least_accum(anyelement, variadic anyarray)
returns anyelement language sql as
  'select least($1, min($2[i])) from generate_subscripts($2,1) g(i)';`,
			},
			{
				Statement: `create aggregate least_agg(variadic items anyarray) (
  stype = anyelement, sfunc = least_accum
);`,
			},
			{
				Statement: `create function cleast_accum(anycompatible, variadic anycompatiblearray)
returns anycompatible language sql as
  'select least($1, min($2[i])) from generate_subscripts($2,1) g(i)';`,
			},
			{
				Statement: `create aggregate cleast_agg(variadic items anycompatiblearray) (
  stype = anycompatible, sfunc = cleast_accum
);`,
			},
			{
				Statement: `create aggregate my_percentile_disc(float8 ORDER BY anyelement) (
  stype = internal,
  sfunc = ordered_set_transition,
  finalfunc = percentile_disc_final,
  finalfunc_extra = true,
  finalfunc_modify = read_write
);`,
			},
			{
				Statement: `create aggregate my_rank(VARIADIC "any" ORDER BY VARIADIC "any") (
  stype = internal,
  sfunc = ordered_set_transition_multi,
  finalfunc = rank_final,
  finalfunc_extra = true,
  hypothetical
);`,
			},
			{
				Statement: `alter aggregate my_percentile_disc(float8 ORDER BY anyelement)
  rename to test_percentile_disc;`,
			},
			{
				Statement: `alter aggregate my_rank(VARIADIC "any" ORDER BY VARIADIC "any")
  rename to test_rank;`,
			},
			{
				Statement: `\da test_*
                                       List of aggregate functions
 Schema |         Name         | Result data type |          Argument data types           | Description 
--------+----------------------+------------------+----------------------------------------+-------------
 public | test_percentile_disc | anyelement       | double precision ORDER BY anyelement   | 
 public | test_rank            | bigint           | VARIADIC "any" ORDER BY VARIADIC "any" | 
(2 rows)
CREATE AGGREGATE sumdouble (float8)
(
    stype = float8,
    sfunc = float8pl,
    mstype = float8,
    msfunc = float8pl,
    minvfunc = float8mi
);`,
			},
			{
				Statement: `CREATE AGGREGATE myavg (numeric)
(
	stype = internal,
	sfunc = numeric_avg_accum,
	serialfunc = numeric_avg_serialize
);`,
				ErrorString: `must specify both or neither of serialization and deserialization functions`,
			},
			{
				Statement: `CREATE AGGREGATE myavg (numeric)
(
	stype = internal,
	sfunc = numeric_avg_accum,
	serialfunc = numeric_avg_deserialize,
	deserialfunc = numeric_avg_deserialize
);`,
				ErrorString: `function numeric_avg_deserialize(internal) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myavg (numeric)
(
	stype = internal,
	sfunc = numeric_avg_accum,
	serialfunc = numeric_avg_serialize,
	deserialfunc = numeric_avg_serialize
);`,
				ErrorString: `function numeric_avg_serialize(bytea, internal) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myavg (numeric)
(
	stype = internal,
	sfunc = numeric_avg_accum,
	serialfunc = numeric_avg_serialize,
	deserialfunc = numeric_avg_deserialize,
	combinefunc = int4larger
);`,
				ErrorString: `function int4larger(internal, internal) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myavg (numeric)
(
	stype = internal,
	sfunc = numeric_avg_accum,
	finalfunc = numeric_avg,
	serialfunc = numeric_avg_serialize,
	deserialfunc = numeric_avg_deserialize,
	combinefunc = numeric_avg_combine,
	finalfunc_modify = shareable  -- just to test a non-default setting
);`,
			},
			{
				Statement: `SELECT aggfnoid, aggtransfn, aggcombinefn, aggtranstype::regtype,
       aggserialfn, aggdeserialfn, aggfinalmodify
FROM pg_aggregate
WHERE aggfnoid = 'myavg'::REGPROC;`,
				Results: []sql.Row{{`myavg`, `numeric_avg_accum`, `numeric_avg_combine`, `internal`, `numeric_avg_serialize`, `numeric_avg_deserialize`, `s`}},
			},
			{
				Statement: `DROP AGGREGATE myavg (numeric);`,
			},
			{
				Statement: `CREATE AGGREGATE myavg (numeric)
(
	stype = internal,
	sfunc = numeric_avg_accum,
	finalfunc = numeric_avg
);`,
			},
			{
				Statement: `CREATE OR REPLACE AGGREGATE myavg (numeric)
(
	stype = internal,
	sfunc = numeric_avg_accum,
	finalfunc = numeric_avg,
	serialfunc = numeric_avg_serialize,
	deserialfunc = numeric_avg_deserialize,
	combinefunc = numeric_avg_combine,
	finalfunc_modify = shareable  -- just to test a non-default setting
);`,
			},
			{
				Statement: `SELECT aggfnoid, aggtransfn, aggcombinefn, aggtranstype::regtype,
       aggserialfn, aggdeserialfn, aggfinalmodify
FROM pg_aggregate
WHERE aggfnoid = 'myavg'::REGPROC;`,
				Results: []sql.Row{{`myavg`, `numeric_avg_accum`, `numeric_avg_combine`, `internal`, `numeric_avg_serialize`, `numeric_avg_deserialize`, `s`}},
			},
			{
				Statement: `CREATE OR REPLACE AGGREGATE myavg (numeric)
(
	stype = numeric,
	sfunc = numeric_add
);`,
			},
			{
				Statement: `SELECT aggfnoid, aggtransfn, aggcombinefn, aggtranstype::regtype,
       aggserialfn, aggdeserialfn, aggfinalmodify
FROM pg_aggregate
WHERE aggfnoid = 'myavg'::REGPROC;`,
				Results: []sql.Row{{`myavg`, `numeric_add`, `-`, `numeric`, `-`, `-`, `r`}},
			},
			{
				Statement: `CREATE OR REPLACE AGGREGATE myavg (numeric)
(
	stype = numeric,
	sfunc = numeric_add,
	finalfunc = numeric_out
);`,
				ErrorString: `cannot change return type of existing function`,
			},
			{
				Statement: `CREATE OR REPLACE AGGREGATE myavg (order by numeric)
(
	stype = numeric,
	sfunc = numeric_add
);`,
				ErrorString: `cannot change routine kind`,
			},
			{
				Statement: `create function sum4(int8,int8,int8,int8) returns int8 as
'select $1 + $2 + $3 + $4' language sql strict immutable;`,
			},
			{
				Statement: `CREATE OR REPLACE AGGREGATE sum3 (int8,int8,int8)
(
	stype = int8,
	sfunc = sum4
);`,
				ErrorString: `cannot change routine kind`,
			},
			{
				Statement: `drop function sum4(int8,int8,int8,int8);`,
			},
			{
				Statement: `DROP AGGREGATE myavg (numeric);`,
			},
			{
				Statement: `CREATE AGGREGATE mysum (int)
(
	stype = int,
	sfunc = int4pl,
	parallel = pear
);`,
				ErrorString: `parameter "parallel" must be SAFE, RESTRICTED, or UNSAFE`,
			},
			{
				Statement: `CREATE FUNCTION float8mi_n(float8, float8) RETURNS float8 AS
$$ SELECT $1 - $2; $$
LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE AGGREGATE invalidsumdouble (float8)
(
    stype = float8,
    sfunc = float8pl,
    mstype = float8,
    msfunc = float8pl,
    minvfunc = float8mi_n
);`,
				ErrorString: `strictness of aggregate's forward and inverse transition functions must match`,
			},
			{
				Statement: `CREATE FUNCTION float8mi_int(float8, float8) RETURNS int AS
$$ SELECT CAST($1 - $2 AS INT); $$
LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE AGGREGATE wrongreturntype (float8)
(
    stype = float8,
    sfunc = float8pl,
    mstype = float8,
    msfunc = float8pl,
    minvfunc = float8mi_int
);`,
				ErrorString: `return type of inverse transition function float8mi_int is not double precision`,
			},
			{
				Statement: `CREATE AGGREGATE case_agg ( -- old syntax
	"Sfunc1" = int4pl,
	"Basetype" = int4,
	"Stype1" = int4,
	"Initcond1" = '0',
	"Parallel" = safe
);`,
				ErrorString: `aggregate stype must be specified`,
			},
			{
				Statement: `CREATE AGGREGATE case_agg(float8)
(
	"Stype" = internal,
	"Sfunc" = ordered_set_transition,
	"Finalfunc" = percentile_disc_final,
	"Finalfunc_extra" = true,
	"Finalfunc_modify" = read_write,
	"Parallel" = safe
);`,
				ErrorString: `aggregate stype must be specified`,
			},
		},
	})
}
