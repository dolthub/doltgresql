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

func TestPolymorphism(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_polymorphism)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_polymorphism,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create function polyf(x anyelement) returns anyelement as $$
  select x + 1
$$ language sql;`,
			},
			{
				Statement: `select polyf(42) as int, polyf(4.5) as num;`,
				Results:   []sql.Row{{43, 5.5}},
			},
			{
				Statement:   `select polyf(point(3,4));  -- fail for lack of + operator`,
				ErrorString: `operator does not exist: point + integer`,
			},
			{
				Statement: `QUERY:  
  select x + 1
CONTEXT:  SQL function "polyf" during inlining
drop function polyf(x anyelement);`,
			},
			{
				Statement: `create function polyf(x anyelement) returns anyarray as $$
  select array[x + 1, x + 2]
$$ language sql;`,
			},
			{
				Statement: `select polyf(42) as int, polyf(4.5) as num;`,
				Results:   []sql.Row{{`{43,44}`, `{5.5,6.5}`}},
			},
			{
				Statement: `drop function polyf(x anyelement);`,
			},
			{
				Statement: `create function polyf(x anyarray) returns anyelement as $$
  select x[1]
$$ language sql;`,
			},
			{
				Statement: `select polyf(array[2,4]) as int, polyf(array[4.5, 7.7]) as num;`,
				Results:   []sql.Row{{2, 4.5}},
			},
			{
				Statement:   `select polyf(stavalues1) from pg_statistic;  -- fail, can't infer element type`,
				ErrorString: `cannot determine element type of "anyarray" argument`,
			},
			{
				Statement: `drop function polyf(x anyarray);`,
			},
			{
				Statement: `create function polyf(x anyarray) returns anyarray as $$
  select x
$$ language sql;`,
			},
			{
				Statement: `select polyf(array[2,4]) as int, polyf(array[4.5, 7.7]) as num;`,
				Results:   []sql.Row{{`{2,4}`, `{4.5,7.7}`}},
			},
			{
				Statement:   `select polyf(stavalues1) from pg_statistic;  -- fail, can't infer element type`,
				ErrorString: `return type anyarray is not supported for SQL functions`,
			},
			{
				Statement: `CONTEXT:  SQL function "polyf" during inlining
drop function polyf(x anyarray);`,
			},
			{
				Statement: `create function polyf(x anyelement) returns anyrange as $$
  select array[x + 1, x + 2]
$$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function polyf(x anyrange) returns anyarray as $$
  select array[lower(x), upper(x)]
$$ language sql;`,
			},
			{
				Statement: `select polyf(int4range(42, 49)) as int, polyf(float8range(4.5, 7.8)) as num;`,
				Results:   []sql.Row{{`{42,49}`, `{4.5,7.8}`}},
			},
			{
				Statement: `drop function polyf(x anyrange);`,
			},
			{
				Statement: `create function polyf(x anycompatible, y anycompatible) returns anycompatiblearray as $$
  select array[x, y]
$$ language sql;`,
			},
			{
				Statement: `select polyf(2, 4) as int, polyf(2, 4.5) as num;`,
				Results:   []sql.Row{{`{2,4}`, `{2,4.5}`}},
			},
			{
				Statement: `drop function polyf(x anycompatible, y anycompatible);`,
			},
			{
				Statement: `create function polyf(x anycompatiblerange, y anycompatible, z anycompatible) returns anycompatiblearray as $$
  select array[lower(x), upper(x), y, z]
$$ language sql;`,
			},
			{
				Statement: `select polyf(int4range(42, 49), 11, 2::smallint) as int, polyf(float8range(4.5, 7.8), 7.8, 11::real) as num;`,
				Results:   []sql.Row{{`{42,49,11,2}`, `{4.5,7.8,7.8,11}`}},
			},
			{
				Statement:   `select polyf(int4range(42, 49), 11, 4.5) as fail;  -- range type doesn't fit`,
				ErrorString: `function polyf(int4range, integer, numeric) does not exist`,
			},
			{
				Statement: `drop function polyf(x anycompatiblerange, y anycompatible, z anycompatible);`,
			},
			{
				Statement: `create function polyf(x anycompatiblemultirange, y anycompatible, z anycompatible) returns anycompatiblearray as $$
  select array[lower(x), upper(x), y, z]
$$ language sql;`,
			},
			{
				Statement: `select polyf(multirange(int4range(42, 49)), 11, 2::smallint) as int, polyf(multirange(float8range(4.5, 7.8)), 7.8, 11::real) as num;`,
				Results:   []sql.Row{{`{42,49,11,2}`, `{4.5,7.8,7.8,11}`}},
			},
			{
				Statement:   `select polyf(multirange(int4range(42, 49)), 11, 4.5) as fail;  -- range type doesn't fit`,
				ErrorString: `function polyf(int4multirange, integer, numeric) does not exist`,
			},
			{
				Statement: `drop function polyf(x anycompatiblemultirange, y anycompatible, z anycompatible);`,
			},
			{
				Statement: `create function polyf(x anycompatible) returns anycompatiblerange as $$
  select array[x + 1, x + 2]
$$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function polyf(x anycompatiblerange, y anycompatiblearray) returns anycompatiblerange as $$
  select x
$$ language sql;`,
			},
			{
				Statement: `select polyf(int4range(42, 49), array[11]) as int, polyf(float8range(4.5, 7.8), array[7]) as num;`,
				Results:   []sql.Row{{`[42,49)`, `[4.5,7.8)`}},
			},
			{
				Statement: `drop function polyf(x anycompatiblerange, y anycompatiblearray);`,
			},
			{
				Statement: `create function polyf(x anycompatible) returns anycompatiblemultirange as $$
  select array[x + 1, x + 2]
$$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function polyf(x anycompatiblemultirange, y anycompatiblearray) returns anycompatiblemultirange as $$
  select x
$$ language sql;`,
			},
			{
				Statement: `select polyf(multirange(int4range(42, 49)), array[11]) as int, polyf(multirange(float8range(4.5, 7.8)), array[7]) as num;`,
				Results:   []sql.Row{{`{[42,49)}`, `{[4.5,7.8)}`}},
			},
			{
				Statement: `drop function polyf(x anycompatiblemultirange, y anycompatiblearray);`,
			},
			{
				Statement: `create function polyf(a anyelement, b anyarray,
                      c anycompatible, d anycompatible,
                      OUT x anyarray, OUT y anycompatiblearray)
as $$
  select a || b, array[c, d]
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x), y, pg_typeof(y)
  from polyf(11, array[1, 2], 42, 34.5);`,
				Results: []sql.Row{{`{11,1,2}`, `integer[]`, `{42,34.5}`, `numeric[]`}},
			},
			{
				Statement: `select x, pg_typeof(x), y, pg_typeof(y)
  from polyf(11, array[1, 2], point(1,2), point(3,4));`,
				Results: []sql.Row{{`{11,1,2}`, `integer[]`, `{"(1,2)","(3,4)"}`, `point[]`}},
			},
			{
				Statement: `select x, pg_typeof(x), y, pg_typeof(y)
  from polyf(11, '{1,2}', point(1,2), '(3,4)');`,
				Results: []sql.Row{{`{11,1,2}`, `integer[]`, `{"(1,2)","(3,4)"}`, `point[]`}},
			},
			{
				Statement: `select x, pg_typeof(x), y, pg_typeof(y)
  from polyf(11, array[1, 2.2], 42, 34.5);  -- fail`,
				ErrorString: `function polyf(integer, numeric[], integer, numeric) does not exist`,
			},
			{
				Statement: `drop function polyf(a anyelement, b anyarray,
                    c anycompatible, d anycompatible);`,
			},
			{
				Statement: `create function polyf(anyrange) returns anymultirange
as 'select multirange($1);' language sql;`,
			},
			{
				Statement: `select polyf(int4range(1,10));`,
				Results:   []sql.Row{{`{[1,10)}`}},
			},
			{
				Statement:   `select polyf(null);`,
				ErrorString: `could not determine polymorphic type because input has type unknown`,
			},
			{
				Statement: `drop function polyf(anyrange);`,
			},
			{
				Statement: `create function polyf(anymultirange) returns anyelement
as 'select lower($1);' language sql;`,
			},
			{
				Statement: `select polyf(int4multirange(int4range(1,10), int4range(20,30)));`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `select polyf(null);`,
				ErrorString: `could not determine polymorphic type because input has type unknown`,
			},
			{
				Statement: `drop function polyf(anymultirange);`,
			},
			{
				Statement: `create function polyf(anycompatiblerange) returns anycompatiblemultirange
as 'select multirange($1);' language sql;`,
			},
			{
				Statement: `select polyf(int4range(1,10));`,
				Results:   []sql.Row{{`{[1,10)}`}},
			},
			{
				Statement:   `select polyf(null);`,
				ErrorString: `could not determine polymorphic type anycompatiblerange because input has type unknown`,
			},
			{
				Statement: `drop function polyf(anycompatiblerange);`,
			},
			{
				Statement: `create function polyf(anymultirange) returns anyrange
as 'select range_merge($1);' language sql;`,
			},
			{
				Statement: `select polyf(int4multirange(int4range(1,10), int4range(20,30)));`,
				Results:   []sql.Row{{`[1,30)`}},
			},
			{
				Statement:   `select polyf(null);`,
				ErrorString: `could not determine polymorphic type because input has type unknown`,
			},
			{
				Statement: `drop function polyf(anymultirange);`,
			},
			{
				Statement: `create function polyf(anycompatiblemultirange) returns anycompatiblerange
as 'select range_merge($1);' language sql;`,
			},
			{
				Statement: `select polyf(int4multirange(int4range(1,10), int4range(20,30)));`,
				Results:   []sql.Row{{`[1,30)`}},
			},
			{
				Statement:   `select polyf(null);`,
				ErrorString: `could not determine polymorphic type anycompatiblerange because input has type unknown`,
			},
			{
				Statement: `drop function polyf(anycompatiblemultirange);`,
			},
			{
				Statement: `create function polyf(anycompatiblemultirange) returns anycompatible
as 'select lower($1);' language sql;`,
			},
			{
				Statement: `select polyf(int4multirange(int4range(1,10), int4range(20,30)));`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `select polyf(null);`,
				ErrorString: `could not determine polymorphic type anycompatiblemultirange because input has type unknown`,
			},
			{
				Statement: `drop function polyf(anycompatiblemultirange);`,
			},
			{
				Statement: `-----------
-- ----------------
CREATE FUNCTION stfp(anyarray) RETURNS anyarray AS
'select $1' LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION stfnp(int[]) RETURNS int[] AS
'select $1' LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION tfp(anyarray,anyelement) RETURNS anyarray AS
'select $1 || $2' LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION tfnp(int[],int) RETURNS int[] AS
'select $1 || $2' LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION tf1p(anyarray,int) RETURNS anyarray AS
'select $1' LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION tf2p(int[],anyelement) RETURNS int[] AS
'select $1' LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION sum3(anyelement,anyelement,anyelement) returns anyelement AS
'select $1+$2+$3' language sql strict;`,
			},
			{
				Statement: `CREATE FUNCTION ffp(anyarray) RETURNS anyarray AS
'select $1' LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE FUNCTION ffnp(int[]) returns int[] as
'select $1' LANGUAGE SQL;`,
			},
			{
				Statement: `--     ------------------------
--     -------
CREATE AGGREGATE myaggp01a(*) (SFUNC = stfnp, STYPE = int4[],
  FINALFUNC = ffp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp02a(*) (SFUNC = stfnp, STYPE = anyarray,
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp03a(*) (SFUNC = stfp, STYPE = int4[],
  FINALFUNC = ffp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp03b(*) (SFUNC = stfp, STYPE = int4[],
  INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp04a(*) (SFUNC = stfp, STYPE = anyarray,
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp04b(*) (SFUNC = stfp, STYPE = anyarray,
  INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `--    -------------------------------------
--    -----------------------
CREATE AGGREGATE myaggp05a(BASETYPE = int, SFUNC = tfnp, STYPE = int[],
  FINALFUNC = ffp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp06a(BASETYPE = int, SFUNC = tf2p, STYPE = int[],
  FINALFUNC = ffp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp07a(BASETYPE = anyelement, SFUNC = tfnp, STYPE = int[],
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `function tfnp(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp08a(BASETYPE = anyelement, SFUNC = tf2p, STYPE = int[],
  FINALFUNC = ffp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp09a(BASETYPE = int, SFUNC = tf1p, STYPE = int[],
  FINALFUNC = ffp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp09b(BASETYPE = int, SFUNC = tf1p, STYPE = int[],
  INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp10a(BASETYPE = int, SFUNC = tfp, STYPE = int[],
  FINALFUNC = ffp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp10b(BASETYPE = int, SFUNC = tfp, STYPE = int[],
  INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp11a(BASETYPE = anyelement, SFUNC = tf1p, STYPE = int[],
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `function tf1p(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp11b(BASETYPE = anyelement, SFUNC = tf1p, STYPE = int[],
  INITCOND = '{}');`,
				ErrorString: `function tf1p(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp12a(BASETYPE = anyelement, SFUNC = tfp, STYPE = int[],
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `function tfp(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp12b(BASETYPE = anyelement, SFUNC = tfp, STYPE = int[],
  INITCOND = '{}');`,
				ErrorString: `function tfp(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp13a(BASETYPE = int, SFUNC = tfnp, STYPE = anyarray,
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp14a(BASETYPE = int, SFUNC = tf2p, STYPE = anyarray,
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp15a(BASETYPE = anyelement, SFUNC = tfnp,
  STYPE = anyarray, FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `function tfnp(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp16a(BASETYPE = anyelement, SFUNC = tf2p,
  STYPE = anyarray, FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `function tf2p(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp17a(BASETYPE = int, SFUNC = tf1p, STYPE = anyarray,
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp17b(BASETYPE = int, SFUNC = tf1p, STYPE = anyarray,
  INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp18a(BASETYPE = int, SFUNC = tfp, STYPE = anyarray,
  FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp18b(BASETYPE = int, SFUNC = tfp, STYPE = anyarray,
  INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp19a(BASETYPE = anyelement, SFUNC = tf1p,
  STYPE = anyarray, FINALFUNC = ffp, INITCOND = '{}');`,
				ErrorString: `function tf1p(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp19b(BASETYPE = anyelement, SFUNC = tf1p,
  STYPE = anyarray, INITCOND = '{}');`,
				ErrorString: `function tf1p(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp20a(BASETYPE = anyelement, SFUNC = tfp,
  STYPE = anyarray, FINALFUNC = ffp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggp20b(BASETYPE = anyelement, SFUNC = tfp,
  STYPE = anyarray, INITCOND = '{}');`,
			},
			{
				Statement: `--     ------------------------
--     -------
CREATE AGGREGATE myaggn01a(*) (SFUNC = stfnp, STYPE = int4[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn01b(*) (SFUNC = stfnp, STYPE = int4[],
  INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn02a(*) (SFUNC = stfnp, STYPE = anyarray,
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn02b(*) (SFUNC = stfnp, STYPE = anyarray,
  INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn03a(*) (SFUNC = stfp, STYPE = int4[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn04a(*) (SFUNC = stfp, STYPE = anyarray,
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `--    -------------------------------------
--    -----------------------
CREATE AGGREGATE myaggn05a(BASETYPE = int, SFUNC = tfnp, STYPE = int[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn05b(BASETYPE = int, SFUNC = tfnp, STYPE = int[],
  INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn06a(BASETYPE = int, SFUNC = tf2p, STYPE = int[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn06b(BASETYPE = int, SFUNC = tf2p, STYPE = int[],
  INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn07a(BASETYPE = anyelement, SFUNC = tfnp, STYPE = int[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `function tfnp(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn07b(BASETYPE = anyelement, SFUNC = tfnp, STYPE = int[],
  INITCOND = '{}');`,
				ErrorString: `function tfnp(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn08a(BASETYPE = anyelement, SFUNC = tf2p, STYPE = int[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn08b(BASETYPE = anyelement, SFUNC = tf2p, STYPE = int[],
  INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn09a(BASETYPE = int, SFUNC = tf1p, STYPE = int[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn10a(BASETYPE = int, SFUNC = tfp, STYPE = int[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn11a(BASETYPE = anyelement, SFUNC = tf1p, STYPE = int[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `function tf1p(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn12a(BASETYPE = anyelement, SFUNC = tfp, STYPE = int[],
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `function tfp(integer[], anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn13a(BASETYPE = int, SFUNC = tfnp, STYPE = anyarray,
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn13b(BASETYPE = int, SFUNC = tfnp, STYPE = anyarray,
  INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn14a(BASETYPE = int, SFUNC = tf2p, STYPE = anyarray,
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn14b(BASETYPE = int, SFUNC = tf2p, STYPE = anyarray,
  INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn15a(BASETYPE = anyelement, SFUNC = tfnp,
  STYPE = anyarray, FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `function tfnp(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn15b(BASETYPE = anyelement, SFUNC = tfnp,
  STYPE = anyarray, INITCOND = '{}');`,
				ErrorString: `function tfnp(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn16a(BASETYPE = anyelement, SFUNC = tf2p,
  STYPE = anyarray, FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `function tf2p(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn16b(BASETYPE = anyelement, SFUNC = tf2p,
  STYPE = anyarray, INITCOND = '{}');`,
				ErrorString: `function tf2p(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn17a(BASETYPE = int, SFUNC = tf1p, STYPE = anyarray,
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn18a(BASETYPE = int, SFUNC = tfp, STYPE = anyarray,
  FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `cannot determine transition data type`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn19a(BASETYPE = anyelement, SFUNC = tf1p,
  STYPE = anyarray, FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `function tf1p(anyarray, anyelement) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE myaggn20a(BASETYPE = anyelement, SFUNC = tfp,
  STYPE = anyarray, FINALFUNC = ffnp, INITCOND = '{}');`,
				ErrorString: `function ffnp(anyarray) does not exist`,
			},
			{
				Statement: `CREATE AGGREGATE mysum2(anyelement,anyelement) (SFUNC = sum3,
  STYPE = anyelement, INITCOND = '0');`,
			},
			{
				Statement: `create temp table t(f1 int, f2 int[], f3 text);`,
			},
			{
				Statement: `insert into t values(1,array[1],'a');`,
			},
			{
				Statement: `insert into t values(1,array[11],'b');`,
			},
			{
				Statement: `insert into t values(1,array[111],'c');`,
			},
			{
				Statement: `insert into t values(2,array[2],'a');`,
			},
			{
				Statement: `insert into t values(2,array[22],'b');`,
			},
			{
				Statement: `insert into t values(2,array[222],'c');`,
			},
			{
				Statement: `insert into t values(3,array[3],'a');`,
			},
			{
				Statement: `insert into t values(3,array[3],'b');`,
			},
			{
				Statement: `select f3, myaggp01a(*) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggp03a(*) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggp03b(*) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggp05a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{1,2,3}`}, {`b`, `{1,2,3}`}, {`c`, `{1,2}`}},
			},
			{
				Statement: `select f3, myaggp06a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggp08a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggp09a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggp09b(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggp10a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{1,2,3}`}, {`b`, `{1,2,3}`}, {`c`, `{1,2}`}},
			},
			{
				Statement: `select f3, myaggp10b(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{1,2,3}`}, {`b`, `{1,2,3}`}, {`c`, `{1,2}`}},
			},
			{
				Statement: `select f3, myaggp20a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{1,2,3}`}, {`b`, `{1,2,3}`}, {`c`, `{1,2}`}},
			},
			{
				Statement: `select f3, myaggp20b(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{1,2,3}`}, {`b`, `{1,2,3}`}, {`c`, `{1,2}`}},
			},
			{
				Statement: `select f3, myaggn01a(*) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggn01b(*) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggn03a(*) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggn05a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{1,2,3}`}, {`b`, `{1,2,3}`}, {`c`, `{1,2}`}},
			},
			{
				Statement: `select f3, myaggn05b(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{1,2,3}`}, {`b`, `{1,2,3}`}, {`c`, `{1,2}`}},
			},
			{
				Statement: `select f3, myaggn06a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggn06b(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggn08a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggn08b(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggn09a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{}`}, {`b`, `{}`}, {`c`, `{}`}},
			},
			{
				Statement: `select f3, myaggn10a(f1) from t group by f3 order by f3;`,
				Results:   []sql.Row{{`a`, `{1,2,3}`}, {`b`, `{1,2,3}`}, {`c`, `{1,2}`}},
			},
			{
				Statement: `select mysum2(f1, f1 + 1) from t;`,
				Results:   []sql.Row{{38}},
			},
			{
				Statement: `create function bleat(int) returns int as $$
begin
  raise notice 'bleat %', $1;`,
			},
			{
				Statement: `  return $1;`,
			},
			{
				Statement: `end$$ language plpgsql;`,
			},
			{
				Statement: `create function sql_if(bool, anyelement, anyelement) returns anyelement as $$
select case when $1 then $2 else $3 end $$ language sql;`,
			},
			{
				Statement: `select f1, sql_if(f1 > 0, bleat(f1), bleat(f1 + 1)) from int4_tbl;`,
				Results:   []sql.Row{{0, 1}, {123456, 123456}, {-123456, -123455}, {2147483647, 2147483647}, {-2147483647, -2147483646}},
			},
			{
				Statement: `select q2, sql_if(q2 > 0, q2, q2 + 1) from int8_tbl;`,
				Results:   []sql.Row{{456, 456}, {4567890123456789, 4567890123456789}, {123, 123}, {4567890123456789, 4567890123456789}, {-4567890123456789, -4567890123456788}},
			},
			{
				Statement: `CREATE AGGREGATE array_larger_accum (anyarray)
(
    sfunc = array_larger,
    stype = anyarray,
    initcond = '{}'
);`,
			},
			{
				Statement: `SELECT array_larger_accum(i)
FROM (VALUES (ARRAY[1,2]), (ARRAY[3,4])) as t(i);`,
				Results: []sql.Row{{`{3,4}`}},
			},
			{
				Statement: `SELECT array_larger_accum(i)
FROM (VALUES (ARRAY[row(1,2),row(3,4)]), (ARRAY[row(5,6),row(7,8)])) as t(i);`,
				Results: []sql.Row{{`{"(5,6)","(7,8)"}`}},
			},
			{
				Statement: `create function add_group(grp anyarray, ad anyelement, size integer)
  returns anyarray
  as $$
begin
  if grp is null then
    return array[ad];`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  if array_upper(grp, 1) < size then
    return grp || ad;`,
			},
			{
				Statement: `  end if;`,
			},
			{
				Statement: `  return grp;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$
  language plpgsql immutable;`,
			},
			{
				Statement: `create aggregate build_group(anyelement, integer) (
  SFUNC = add_group,
  STYPE = anyarray
);`,
			},
			{
				Statement: `select build_group(q1,3) from int8_tbl;`,
				Results:   []sql.Row{{`{123,123,4567890123456789}`}},
			},
			{
				Statement: `create aggregate build_group(int8, integer) (
  SFUNC = add_group,
  STYPE = int2[]
);`,
				ErrorString: `function add_group(smallint[], bigint, integer) does not exist`,
			},
			{
				Statement: `create aggregate build_group(int8, integer) (
  SFUNC = add_group,
  STYPE = int8[]
);`,
			},
			{
				Statement: `create function first_el_transfn(anyarray, anyelement) returns anyarray as
'select $1 || $2' language sql immutable;`,
			},
			{
				Statement: `create function first_el(anyarray) returns anyelement as
'select $1[1]' language sql strict immutable;`,
			},
			{
				Statement: `create aggregate first_el_agg_f8(float8) (
  SFUNC = array_append,
  STYPE = float8[],
  FINALFUNC = first_el
);`,
			},
			{
				Statement: `create aggregate first_el_agg_any(anyelement) (
  SFUNC = first_el_transfn,
  STYPE = anyarray,
  FINALFUNC = first_el
);`,
			},
			{
				Statement: `select first_el_agg_f8(x::float8) from generate_series(1,10) x;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select first_el_agg_any(x) from generate_series(1,10) x;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select first_el_agg_f8(x::float8) over(order by x) from generate_series(1,10) x;`,
				Results:   []sql.Row{{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1}, {1}, {1}},
			},
			{
				Statement: `select first_el_agg_any(x) over(order by x) from generate_series(1,10) x;`,
				Results:   []sql.Row{{1}, {1}, {1}, {1}, {1}, {1}, {1}, {1}, {1}, {1}},
			},
			{
				Statement: `select distinct array_ndims(histogram_bounds) from pg_stats
where histogram_bounds is not null;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement:   `select max(histogram_bounds) from pg_stats where tablename = 'pg_am';`,
				ErrorString: `cannot compare arrays of different element types`,
			},
			{
				Statement: `select array_in('{1,2,3}','int4'::regtype,-1);  -- this has historically worked`,
				Results:   []sql.Row{{`{1,2,3}`}},
			},
			{
				Statement:   `select * from array_in('{1,2,3}','int4'::regtype,-1);  -- this not`,
				ErrorString: `function "array_in" in FROM has unsupported return type anyarray`,
			},
			{
				Statement:   `select anyrange_in('[10,20)','int4range'::regtype,-1);`,
				ErrorString: `cannot accept a value of type anyrange`,
			},
			{
				Statement: `create function myleast(variadic anyarray) returns anyelement as $$
  select min($1[i]) from generate_subscripts($1,1) g(i)
$$ language sql immutable strict;`,
			},
			{
				Statement: `select myleast(10, 1, 20, 33);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select myleast(1.1, 0.22, 0.55);`,
				Results:   []sql.Row{{0.22}},
			},
			{
				Statement: `select myleast('z'::text);`,
				Results:   []sql.Row{{`z`}},
			},
			{
				Statement:   `select myleast(); -- fail`,
				ErrorString: `function myleast() does not exist`,
			},
			{
				Statement: `select myleast(variadic array[1,2,3,4,-1]);`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `select myleast(variadic array[1.1, -5.5]);`,
				Results:   []sql.Row{{-5.5}},
			},
			{
				Statement: `select myleast(variadic array[]::int[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `create function concat(text, variadic anyarray) returns text as $$
  select array_to_string($2, $1);`,
			},
			{
				Statement: `$$ language sql immutable strict;`,
			},
			{
				Statement: `select concat('%', 1, 2, 3, 4, 5);`,
				Results:   []sql.Row{{`1%2%3%4%5`}},
			},
			{
				Statement: `select concat('|', 'a'::text, 'b', 'c');`,
				Results:   []sql.Row{{`a|b|c`}},
			},
			{
				Statement: `select concat('|', variadic array[1,2,33]);`,
				Results:   []sql.Row{{`1|2|33`}},
			},
			{
				Statement: `select concat('|', variadic array[]::int[]);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `drop function concat(text, anyarray);`,
			},
			{
				Statement: `create function formarray(anyelement, variadic anyarray) returns anyarray as $$
  select array_prepend($1, $2);`,
			},
			{
				Statement: `$$ language sql immutable strict;`,
			},
			{
				Statement: `select formarray(1,2,3,4,5);`,
				Results:   []sql.Row{{`{1,2,3,4,5}`}},
			},
			{
				Statement: `select formarray(1.1, variadic array[1.2,55.5]);`,
				Results:   []sql.Row{{`{1.1,1.2,55.5}`}},
			},
			{
				Statement:   `select formarray(1.1, array[1.2,55.5]); -- fail without variadic`,
				ErrorString: `function formarray(numeric, numeric[]) does not exist`,
			},
			{
				Statement:   `select formarray(1, 'x'::text); -- fail, type mismatch`,
				ErrorString: `function formarray(integer, text) does not exist`,
			},
			{
				Statement:   `select formarray(1, variadic array['x'::text]); -- fail, type mismatch`,
				ErrorString: `function formarray(integer, text[]) does not exist`,
			},
			{
				Statement: `drop function formarray(anyelement, variadic anyarray);`,
			},
			{
				Statement: `select pg_typeof(null);           -- unknown`,
				Results:   []sql.Row{{`unknown`}},
			},
			{
				Statement: `select pg_typeof(0);              -- integer`,
				Results:   []sql.Row{{`integer`}},
			},
			{
				Statement: `select pg_typeof(0.0);            -- numeric`,
				Results:   []sql.Row{{`numeric`}},
			},
			{
				Statement: `select pg_typeof(1+1 = 2);        -- boolean`,
				Results:   []sql.Row{{`boolean`}},
			},
			{
				Statement: `select pg_typeof('x');            -- unknown`,
				Results:   []sql.Row{{`unknown`}},
			},
			{
				Statement: `select pg_typeof('' || '');       -- text`,
				Results:   []sql.Row{{`text`}},
			},
			{
				Statement: `select pg_typeof(pg_typeof(0));   -- regtype`,
				Results:   []sql.Row{{`regtype`}},
			},
			{
				Statement: `select pg_typeof(array[1.2,55.5]); -- numeric[]`,
				Results:   []sql.Row{{`numeric[]`}},
			},
			{
				Statement: `select pg_typeof(myleast(10, 1, 20, 33));  -- polymorphic input`,
				Results:   []sql.Row{{`integer`}},
			},
			{
				Statement: `create function dfunc(a int = 1, int = 2) returns int as $$
  select $1 + $2;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select dfunc();`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `select dfunc(10);`,
				Results:   []sql.Row{{12}},
			},
			{
				Statement: `select dfunc(10, 20);`,
				Results:   []sql.Row{{30}},
			},
			{
				Statement:   `select dfunc(10, 20, 30);  -- fail`,
				ErrorString: `function dfunc(integer, integer, integer) does not exist`,
			},
			{
				Statement:   `drop function dfunc();  -- fail`,
				ErrorString: `function dfunc() does not exist`,
			},
			{
				Statement:   `drop function dfunc(int);  -- fail`,
				ErrorString: `function dfunc(integer) does not exist`,
			},
			{
				Statement: `drop function dfunc(int, int);  -- ok`,
			},
			{
				Statement: `create function dfunc(a int = 1, b int) returns int as $$
  select $1 + $2;`,
			},
			{
				Statement:   `$$ language sql;`,
				ErrorString: `input parameters after one with a default value must also have defaults`,
			},
			{
				Statement: `create function dfunc(a int = 1, out sum int, b int = 2) as $$
  select $1 + $2;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select dfunc();`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `\df dfunc
                                          List of functions
 Schema | Name  | Result data type |                    Argument data types                    | Type 
--------+-------+------------------+-----------------------------------------------------------+------
 public | dfunc | integer          | a integer DEFAULT 1, OUT sum integer, b integer DEFAULT 2 | func
(1 row)
drop function dfunc(int, int);`,
			},
			{
				Statement: `create function dfunc(a int DEFAULT 1.0, int DEFAULT '-1') returns int as $$
  select $1 + $2;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select dfunc();`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `create function dfunc(a text DEFAULT 'Hello', b text DEFAULT 'World') returns text as $$
  select $1 || ', ' || $2;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement:   `select dfunc();  -- fail: which dfunc should be called? int or text`,
				ErrorString: `function dfunc() is not unique`,
			},
			{
				Statement: `select dfunc('Hi');  -- ok`,
				Results:   []sql.Row{{`Hi, World`}},
			},
			{
				Statement: `select dfunc('Hi', 'City');  -- ok`,
				Results:   []sql.Row{{`Hi, City`}},
			},
			{
				Statement: `select dfunc(0);  -- ok`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `select dfunc(10, 20);  -- ok`,
				Results:   []sql.Row{{30}},
			},
			{
				Statement: `drop function dfunc(int, int);`,
			},
			{
				Statement: `drop function dfunc(text, text);`,
			},
			{
				Statement: `create function dfunc(int = 1, int = 2) returns int as $$
  select 2;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `create function dfunc(int = 1, int = 2, int = 3, int = 4) returns int as $$
  select 4;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement:   `select dfunc();  -- fail`,
				ErrorString: `function dfunc() is not unique`,
			},
			{
				Statement:   `select dfunc(1);  -- fail`,
				ErrorString: `function dfunc(integer) is not unique`,
			},
			{
				Statement:   `select dfunc(1, 2);  -- fail`,
				ErrorString: `function dfunc(integer, integer) is not unique`,
			},
			{
				Statement: `select dfunc(1, 2, 3);  -- ok`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `select dfunc(1, 2, 3, 4);  -- ok`,
				Results:   []sql.Row{{4}},
			},
			{
				Statement: `drop function dfunc(int, int);`,
			},
			{
				Statement: `drop function dfunc(int, int, int, int);`,
			},
			{
				Statement: `create function dfunc(out int = 20) returns int as $$
  select 1;`,
			},
			{
				Statement:   `$$ language sql;`,
				ErrorString: `only input parameters can have default values`,
			},
			{
				Statement: `create function dfunc(anyelement = 'World'::text) returns text as $$
  select 'Hello, ' || $1::text;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select dfunc();`,
				Results:   []sql.Row{{`Hello, World`}},
			},
			{
				Statement: `select dfunc(0);`,
				Results:   []sql.Row{{`Hello, 0`}},
			},
			{
				Statement: `select dfunc(to_date('20081215','YYYYMMDD'));`,
				Results:   []sql.Row{{`Hello, 12-15-2008`}},
			},
			{
				Statement: `select dfunc('City'::text);`,
				Results:   []sql.Row{{`Hello, City`}},
			},
			{
				Statement: `drop function dfunc(anyelement);`,
			},
			{
				Statement: `create function dfunc(a variadic int[]) returns int as
$$ select array_upper($1, 1) $$ language sql;`,
			},
			{
				Statement:   `select dfunc();  -- fail`,
				ErrorString: `function dfunc() does not exist`,
			},
			{
				Statement: `select dfunc(10);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select dfunc(10,20);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `create or replace function dfunc(a variadic int[] default array[]::int[]) returns int as
$$ select array_upper($1, 1) $$ language sql;`,
			},
			{
				Statement: `select dfunc();  -- now ok`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select dfunc(10);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select dfunc(10,20);`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `create or replace function dfunc(a variadic int[]) returns int as
$$ select array_upper($1, 1) $$ language sql;`,
				ErrorString: `cannot remove parameter defaults from existing function`,
			},
			{
				Statement: `\df dfunc
                                     List of functions
 Schema | Name  | Result data type |               Argument data types               | Type 
--------+-------+------------------+-------------------------------------------------+------
 public | dfunc | integer          | VARIADIC a integer[] DEFAULT ARRAY[]::integer[] | func
(1 row)
drop function dfunc(a variadic int[]);`,
			},
			{
				Statement: `create function dfunc(int = 1, int = 2, int = 3) returns int as $$
  select 3;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `create function dfunc(int = 1, int = 2) returns int as $$
  select 2;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `create function dfunc(text) returns text as $$
  select $1;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement:   `select dfunc(1);  -- fail`,
				ErrorString: `function dfunc(integer) is not unique`,
			},
			{
				Statement: `select dfunc('Hi');`,
				Results:   []sql.Row{{`Hi`}},
			},
			{
				Statement: `drop function dfunc(int, int, int);`,
			},
			{
				Statement: `drop function dfunc(int, int);`,
			},
			{
				Statement: `drop function dfunc(text);`,
			},
			{
				Statement: `create function dfunc(a int, b int, c int = 0, d int = 0)
  returns table (a int, b int, c int, d int) as $$
  select $1, $2, $3, $4;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select (dfunc(10,20,30)).*;`,
				Results:   []sql.Row{{10, 20, 30, 0}},
			},
			{
				Statement: `select (dfunc(a := 10, b := 20, c := 30)).*;`,
				Results:   []sql.Row{{10, 20, 30, 0}},
			},
			{
				Statement: `select * from dfunc(a := 10, b := 20);`,
				Results:   []sql.Row{{10, 20, 0, 0}},
			},
			{
				Statement: `select * from dfunc(b := 10, a := 20);`,
				Results:   []sql.Row{{20, 10, 0, 0}},
			},
			{
				Statement:   `select * from dfunc(0);  -- fail`,
				ErrorString: `function dfunc(integer) does not exist`,
			},
			{
				Statement: `select * from dfunc(1,2);`,
				Results:   []sql.Row{{1, 2, 0, 0}},
			},
			{
				Statement: `select * from dfunc(1,2,c := 3);`,
				Results:   []sql.Row{{1, 2, 3, 0}},
			},
			{
				Statement: `select * from dfunc(1,2,d := 3);`,
				Results:   []sql.Row{{1, 2, 0, 3}},
			},
			{
				Statement:   `select * from dfunc(x := 20, b := 10, x := 30);  -- fail, duplicate name`,
				ErrorString: `argument name "x" used more than once`,
			},
			{
				Statement:   `select * from dfunc(10, b := 20, 30);  -- fail, named args must be last`,
				ErrorString: `positional argument cannot follow named argument`,
			},
			{
				Statement:   `select * from dfunc(x := 10, b := 20, c := 30);  -- fail, unknown param`,
				ErrorString: `function dfunc(x => integer, b => integer, c => integer) does not exist`,
			},
			{
				Statement:   `select * from dfunc(10, 10, a := 20);  -- fail, a overlaps positional parameter`,
				ErrorString: `function dfunc(integer, integer, a => integer) does not exist`,
			},
			{
				Statement:   `select * from dfunc(1,c := 2,d := 3); -- fail, no value for b`,
				ErrorString: `function dfunc(integer, c => integer, d => integer) does not exist`,
			},
			{
				Statement: `drop function dfunc(int, int, int, int);`,
			},
			{
				Statement: `create function dfunc(a varchar, b numeric, c date = current_date)
  returns table (a varchar, b numeric, c date) as $$
  select $1, $2, $3;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select (dfunc('Hello World', 20, '2009-07-25'::date)).*;`,
				Results:   []sql.Row{{`Hello World`, 20, `07-25-2009`}},
			},
			{
				Statement: `select * from dfunc('Hello World', 20, '2009-07-25'::date);`,
				Results:   []sql.Row{{`Hello World`, 20, `07-25-2009`}},
			},
			{
				Statement: `select * from dfunc(c := '2009-07-25'::date, a := 'Hello World', b := 20);`,
				Results:   []sql.Row{{`Hello World`, 20, `07-25-2009`}},
			},
			{
				Statement: `select * from dfunc('Hello World', b := 20, c := '2009-07-25'::date);`,
				Results:   []sql.Row{{`Hello World`, 20, `07-25-2009`}},
			},
			{
				Statement: `select * from dfunc('Hello World', c := '2009-07-25'::date, b := 20);`,
				Results:   []sql.Row{{`Hello World`, 20, `07-25-2009`}},
			},
			{
				Statement:   `select * from dfunc('Hello World', c := 20, b := '2009-07-25'::date);  -- fail`,
				ErrorString: `function dfunc(unknown, c => integer, b => date) does not exist`,
			},
			{
				Statement: `drop function dfunc(varchar, numeric, date);`,
			},
			{
				Statement: `create function dfunc(a varchar = 'def a', out _a varchar, c numeric = NULL, out _c numeric)
returns record as $$
  select $1, $2;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select (dfunc()).*;`,
				Results:   []sql.Row{{`def a`, ``}},
			},
			{
				Statement: `select * from dfunc();`,
				Results:   []sql.Row{{`def a`, ``}},
			},
			{
				Statement: `select * from dfunc('Hello', 100);`,
				Results:   []sql.Row{{`Hello`, 100}},
			},
			{
				Statement: `select * from dfunc(a := 'Hello', c := 100);`,
				Results:   []sql.Row{{`Hello`, 100}},
			},
			{
				Statement: `select * from dfunc(c := 100, a := 'Hello');`,
				Results:   []sql.Row{{`Hello`, 100}},
			},
			{
				Statement: `select * from dfunc('Hello');`,
				Results:   []sql.Row{{`Hello`, ``}},
			},
			{
				Statement: `select * from dfunc('Hello', c := 100);`,
				Results:   []sql.Row{{`Hello`, 100}},
			},
			{
				Statement: `select * from dfunc(c := 100);`,
				Results:   []sql.Row{{`def a`, 100}},
			},
			{
				Statement: `create or replace function dfunc(a varchar = 'def a', out _a varchar, x numeric = NULL, out _c numeric)
returns record as $$
  select $1, $2;`,
			},
			{
				Statement:   `$$ language sql;`,
				ErrorString: `cannot change name of input parameter "c"`,
			},
			{
				Statement: `create or replace function dfunc(a varchar = 'def a', out _a varchar, numeric = NULL, out _c numeric)
returns record as $$
  select $1, $2;`,
			},
			{
				Statement:   `$$ language sql;`,
				ErrorString: `cannot change name of input parameter "c"`,
			},
			{
				Statement: `drop function dfunc(varchar, numeric);`,
			},
			{
				Statement:   `create function testpolym(a int, a int) returns int as $$ select 1;$$ language sql;`,
				ErrorString: `parameter name "a" used more than once`,
			},
			{
				Statement:   `create function testpolym(int, out a int, out a int) returns int as $$ select 1;$$ language sql;`,
				ErrorString: `parameter name "a" used more than once`,
			},
			{
				Statement:   `create function testpolym(out a int, inout a int) returns int as $$ select 1;$$ language sql;`,
				ErrorString: `parameter name "a" used more than once`,
			},
			{
				Statement:   `create function testpolym(a int, inout a int) returns int as $$ select 1;$$ language sql;`,
				ErrorString: `parameter name "a" used more than once`,
			},
			{
				Statement: `create function testpolym(a int, out a int) returns int as $$ select $1;$$ language sql;`,
			},
			{
				Statement: `select testpolym(37);`,
				Results:   []sql.Row{{37}},
			},
			{
				Statement: `drop function testpolym(int);`,
			},
			{
				Statement: `create function testpolym(a int) returns table(a int) as $$ select $1;$$ language sql;`,
			},
			{
				Statement: `select * from testpolym(37);`,
				Results:   []sql.Row{{37}},
			},
			{
				Statement: `drop function testpolym(int);`,
			},
			{
				Statement: `create function dfunc(a anyelement, b anyelement = null, flag bool = true)
returns anyelement as $$
  select case when $3 then $1 else $2 end;`,
			},
			{
				Statement: `$$ language sql;`,
			},
			{
				Statement: `select dfunc(1,2);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select dfunc('a'::text, 'b'); -- positional notation with default`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a := 1, b := 2);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select dfunc(a := 'a'::text, b := 'b');`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a := 'a'::text, b := 'b', flag := false); -- named notation`,
				Results:   []sql.Row{{`b`}},
			},
			{
				Statement: `select dfunc(b := 'b'::text, a := 'a'); -- named notation with default`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a := 'a'::text, flag := true); -- named notation with default`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a := 'a'::text, flag := false); -- named notation with default`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select dfunc(b := 'b'::text, a := 'a', flag := true); -- named notation`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc('a'::text, 'b', false); -- full positional notation`,
				Results:   []sql.Row{{`b`}},
			},
			{
				Statement: `select dfunc('a'::text, 'b', flag := false); -- mixed notation`,
				Results:   []sql.Row{{`b`}},
			},
			{
				Statement: `select dfunc('a'::text, 'b', true); -- full positional notation`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc('a'::text, 'b', flag := true); -- mixed notation`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a => 1, b => 2);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select dfunc(a => 'a'::text, b => 'b');`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a => 'a'::text, b => 'b', flag => false); -- named notation`,
				Results:   []sql.Row{{`b`}},
			},
			{
				Statement: `select dfunc(b => 'b'::text, a => 'a'); -- named notation with default`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a => 'a'::text, flag => true); -- named notation with default`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a => 'a'::text, flag => false); -- named notation with default`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select dfunc(b => 'b'::text, a => 'a', flag => true); -- named notation`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc('a'::text, 'b', false); -- full positional notation`,
				Results:   []sql.Row{{`b`}},
			},
			{
				Statement: `select dfunc('a'::text, 'b', flag => false); -- mixed notation`,
				Results:   []sql.Row{{`b`}},
			},
			{
				Statement: `select dfunc('a'::text, 'b', true); -- full positional notation`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc('a'::text, 'b', flag => true); -- mixed notation`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select dfunc(a =>-1);`,
				Results:   []sql.Row{{-1}},
			},
			{
				Statement: `select dfunc(a =>+1);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select dfunc(a =>/**/1);`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `select dfunc(a =>--comment to be removed by psql
  1);`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `do $$
  declare r integer;`,
			},
			{
				Statement: `  begin
    select dfunc(a=>-- comment
      1) into r;`,
			},
			{
				Statement: `    raise info 'r = %', r;`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `INFO:  r = 1
CREATE VIEW dfview AS
   SELECT q1, q2,
     dfunc(q1,q2, flag := q1>q2) as c3,
     dfunc(q1, flag := q1<q2, b := q2) as c4
     FROM int8_tbl;`,
			},
			{
				Statement: `select * from dfview;`,
				Results:   []sql.Row{{123, 456, 456, 123}, {123, 4567890123456789, 4567890123456789, 123}, {4567890123456789, 123, 4567890123456789, 123}, {4567890123456789, 4567890123456789, 4567890123456789, 4567890123456789}, {4567890123456789, -4567890123456789, 4567890123456789, -4567890123456789}},
			},
			{
				Statement: `\d+ dfview
                           View "public.dfview"
 Column |  Type  | Collation | Nullable | Default | Storage | Description 
--------+--------+-----------+----------+---------+---------+-------------
 q1     | bigint |           |          |         | plain   | 
 q2     | bigint |           |          |         | plain   | 
 c3     | bigint |           |          |         | plain   | 
 c4     | bigint |           |          |         | plain   | 
View definition:
 SELECT int8_tbl.q1,
    int8_tbl.q2,
    dfunc(int8_tbl.q1, int8_tbl.q2, flag => int8_tbl.q1 > int8_tbl.q2) AS c3,
    dfunc(int8_tbl.q1, flag => int8_tbl.q1 < int8_tbl.q2, b => int8_tbl.q2) AS c4
   FROM int8_tbl;`,
			},
			{
				Statement: `drop view dfview;`,
			},
			{
				Statement: `drop function dfunc(anyelement, anyelement, bool);`,
			},
			{
				Statement: `create function anyctest(anycompatible, anycompatible)
returns anycompatible as $$
  select greatest($1, $2)
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, 12) x;`,
				Results:   []sql.Row{{12, `integer`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, 12.3) x;`,
				Results:   []sql.Row{{12.3, `numeric`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11, point(1,2)) x;  -- fail`,
				ErrorString: `function anyctest(integer, point) does not exist`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest('11', '12.3') x;  -- defaults to text`,
				Results:   []sql.Row{{12.3, `text`}},
			},
			{
				Statement: `drop function anyctest(anycompatible, anycompatible);`,
			},
			{
				Statement: `create function anyctest(anycompatible, anycompatible)
returns anycompatiblearray as $$
  select array[$1, $2]
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, 12) x;`,
				Results:   []sql.Row{{`{11,12}`, `integer[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, 12.3) x;`,
				Results:   []sql.Row{{`{11,12.3}`, `numeric[]`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11, array[1,2]) x;  -- fail`,
				ErrorString: `function anyctest(integer, integer[]) does not exist`,
			},
			{
				Statement: `drop function anyctest(anycompatible, anycompatible);`,
			},
			{
				Statement: `create function anyctest(anycompatible, anycompatiblearray)
returns anycompatiblearray as $$
  select array[$1] || $2
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, array[12]) x;`,
				Results:   []sql.Row{{`{11,12}`, `integer[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, array[12.3]) x;`,
				Results:   []sql.Row{{`{11,12.3}`, `numeric[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(12.3, array[13]) x;`,
				Results:   []sql.Row{{`{12.3,13}`, `numeric[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(12.3, '{13,14.4}') x;`,
				Results:   []sql.Row{{`{12.3,13,14.4}`, `numeric[]`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11, array[point(1,2)]) x;  -- fail`,
				ErrorString: `function anyctest(integer, point[]) does not exist`,
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11, 12) x;  -- fail`,
				ErrorString: `function anyctest(integer, integer) does not exist`,
			},
			{
				Statement: `drop function anyctest(anycompatible, anycompatiblearray);`,
			},
			{
				Statement: `create function anyctest(anycompatible, anycompatiblerange)
returns anycompatiblerange as $$
  select $2
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, int4range(4,7)) x;`,
				Results:   []sql.Row{{`[4,7)`, `int4range`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, numrange(4,7)) x;`,
				Results:   []sql.Row{{`[4,7)`, `numrange`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11, 12) x;  -- fail`,
				ErrorString: `function anyctest(integer, integer) does not exist`,
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11.2, int4range(4,7)) x;  -- fail`,
				ErrorString: `function anyctest(numeric, int4range) does not exist`,
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11.2, '[4,7)') x;  -- fail`,
				ErrorString: `could not determine polymorphic type anycompatiblerange because input has type unknown`,
			},
			{
				Statement: `drop function anyctest(anycompatible, anycompatiblerange);`,
			},
			{
				Statement: `create function anyctest(anycompatiblerange, anycompatiblerange)
returns anycompatible as $$
  select lower($1) + upper($2)
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(int4range(11,12), int4range(4,7)) x;`,
				Results:   []sql.Row{{18, `integer`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(int4range(11,12), numrange(4,7)) x; -- fail`,
				ErrorString: `function anyctest(int4range, numrange) does not exist`,
			},
			{
				Statement: `drop function anyctest(anycompatiblerange, anycompatiblerange);`,
			},
			{
				Statement: `create function anyctest(anycompatible)
returns anycompatiblerange as $$
  select $1
$$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function anyctest(anycompatible, anycompatiblemultirange)
returns anycompatiblemultirange as $$
  select $2
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, multirange(int4range(4,7))) x;`,
				Results:   []sql.Row{{`{[4,7)}`, `int4multirange`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, multirange(numrange(4,7))) x;`,
				Results:   []sql.Row{{`{[4,7)}`, `nummultirange`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11, 12) x;  -- fail`,
				ErrorString: `function anyctest(integer, integer) does not exist`,
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11.2, multirange(int4range(4,7))) x;  -- fail`,
				ErrorString: `function anyctest(numeric, int4multirange) does not exist`,
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11.2, '{[4,7)}') x;  -- fail`,
				ErrorString: `could not determine polymorphic type anycompatiblemultirange because input has type unknown`,
			},
			{
				Statement: `drop function anyctest(anycompatible, anycompatiblemultirange);`,
			},
			{
				Statement: `create function anyctest(anycompatiblemultirange, anycompatiblemultirange)
returns anycompatible as $$
  select lower($1) + upper($2)
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(multirange(int4range(11,12)), multirange(int4range(4,7))) x;`,
				Results:   []sql.Row{{18, `integer`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(multirange(int4range(11,12)), multirange(numrange(4,7))) x; -- fail`,
				ErrorString: `function anyctest(int4multirange, nummultirange) does not exist`,
			},
			{
				Statement: `drop function anyctest(anycompatiblemultirange, anycompatiblemultirange);`,
			},
			{
				Statement: `create function anyctest(anycompatible)
returns anycompatiblemultirange as $$
  select $1
$$ language sql;`,
				ErrorString: `cannot determine result data type`,
			},
			{
				Statement: `create function anyctest(anycompatiblenonarray, anycompatiblenonarray)
returns anycompatiblearray as $$
  select array[$1, $2]
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, 12) x;`,
				Results:   []sql.Row{{`{11,12}`, `integer[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, 12.3) x;`,
				Results:   []sql.Row{{`{11,12.3}`, `numeric[]`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(array[11], array[1,2]) x;  -- fail`,
				ErrorString: `function anyctest(integer[], integer[]) does not exist`,
			},
			{
				Statement: `drop function anyctest(anycompatiblenonarray, anycompatiblenonarray);`,
			},
			{
				Statement: `create function anyctest(a anyelement, b anyarray,
                         c anycompatible, d anycompatible)
returns anycompatiblearray as $$
  select array[c, d]
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, array[1, 2], 42, 34.5) x;`,
				Results:   []sql.Row{{`{42,34.5}`, `numeric[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, array[1, 2], point(1,2), point(3,4)) x;`,
				Results:   []sql.Row{{`{"(1,2)","(3,4)"}`, `point[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, '{1,2}', point(1,2), '(3,4)') x;`,
				Results:   []sql.Row{{`{"(1,2)","(3,4)"}`, `point[]`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11, array[1, 2.2], 42, 34.5) x;  -- fail`,
				ErrorString: `function anyctest(integer, numeric[], integer, numeric) does not exist`,
			},
			{
				Statement: `drop function anyctest(a anyelement, b anyarray,
                       c anycompatible, d anycompatible);`,
			},
			{
				Statement: `create function anyctest(variadic anycompatiblearray)
returns anycompatiblearray as $$
  select $1
$$ language sql;`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, 12) x;`,
				Results:   []sql.Row{{`{11,12}`, `integer[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, 12.2) x;`,
				Results:   []sql.Row{{`{11,12.2}`, `numeric[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(11, '12') x;`,
				Results:   []sql.Row{{`{11,12}`, `integer[]`}},
			},
			{
				Statement:   `select x, pg_typeof(x) from anyctest(11, '12.2') x;  -- fail`,
				ErrorString: `invalid input syntax for type integer: "12.2"`,
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(variadic array[11, 12]) x;`,
				Results:   []sql.Row{{`{11,12}`, `integer[]`}},
			},
			{
				Statement: `select x, pg_typeof(x) from anyctest(variadic array[11, 12.2]) x;`,
				Results:   []sql.Row{{`{11,12.2}`, `numeric[]`}},
			},
			{
				Statement: `drop function anyctest(variadic anycompatiblearray);`,
			},
		},
	})
}
