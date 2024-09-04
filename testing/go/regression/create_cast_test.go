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

func TestCreateCast(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_cast)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_cast,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TYPE casttesttype;`,
			},
			{
				Statement: `CREATE FUNCTION casttesttype_in(cstring)
   RETURNS casttesttype
   AS 'textin'
   LANGUAGE internal STRICT IMMUTABLE;`,
			},
			{
				Statement: `CREATE FUNCTION casttesttype_out(casttesttype)
   RETURNS cstring
   AS 'textout'
   LANGUAGE internal STRICT IMMUTABLE;`,
			},
			{
				Statement: `CREATE TYPE casttesttype (
   internallength = variable,
   input = casttesttype_in,
   output = casttesttype_out,
   alignment = int4
);`,
			},
			{
				Statement: `CREATE FUNCTION casttestfunc(casttesttype) RETURNS int4 LANGUAGE SQL AS
$$ SELECT 1; $$;`,
			},
			{
				Statement:   `SELECT casttestfunc('foo'::text); -- fails, as there's no cast`,
				ErrorString: `function casttestfunc(text) does not exist`,
			},
			{
				Statement: `CREATE CAST (text AS casttesttype) WITHOUT FUNCTION;`,
			},
			{
				Statement:   `SELECT casttestfunc('foo'::text); -- doesn't work, as the cast is explicit`,
				ErrorString: `function casttestfunc(text) does not exist`,
			},
			{
				Statement: `SELECT casttestfunc('foo'::text::casttesttype); -- should work`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `DROP CAST (text AS casttesttype); -- cleanup`,
			},
			{
				Statement: `CREATE CAST (text AS casttesttype) WITHOUT FUNCTION AS IMPLICIT;`,
			},
			{
				Statement: `SELECT casttestfunc('foo'::text); -- Should work now`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT 1234::int4::casttesttype; -- No cast yet, should fail`,
				ErrorString: `cannot cast type integer to casttesttype`,
			},
			{
				Statement: `CREATE CAST (int4 AS casttesttype) WITH INOUT;`,
			},
			{
				Statement: `SELECT 1234::int4::casttesttype; -- Should work now`,
				Results:   []sql.Row{{1234}},
			},
			{
				Statement: `DROP CAST (int4 AS casttesttype);`,
			},
			{
				Statement: `CREATE FUNCTION int4_casttesttype(int4) RETURNS casttesttype LANGUAGE SQL AS
$$ SELECT ('foo'::text || $1::text)::casttesttype; $$;`,
			},
			{
				Statement: `CREATE CAST (int4 AS casttesttype) WITH FUNCTION int4_casttesttype(int4) AS IMPLICIT;`,
			},
			{
				Statement: `SELECT 1234::int4::casttesttype; -- Should work now`,
				Results:   []sql.Row{{`foo1234`}},
			},
		},
	})
}
