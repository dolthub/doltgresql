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

func TestAlterOperator(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_alter_operator)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_alter_operator,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE FUNCTION alter_op_test_fn(boolean, boolean)
RETURNS boolean AS $$ SELECT NULL::BOOLEAN; $$ LANGUAGE sql IMMUTABLE;`,
			},
			{
				Statement: `CREATE FUNCTION customcontsel(internal, oid, internal, integer)
RETURNS float8 AS 'contsel' LANGUAGE internal STABLE STRICT;`,
			},
			{
				Statement: `CREATE OPERATOR === (
    LEFTARG = boolean,
    RIGHTARG = boolean,
    PROCEDURE = alter_op_test_fn,
    COMMUTATOR = ===,
    NEGATOR = !==,
    RESTRICT = customcontsel,
    JOIN = contjoinsel,
    HASHES, MERGES
);`,
			},
			{
				Statement: `SELECT pg_describe_object(refclassid,refobjid,refobjsubid) as ref, deptype
FROM pg_depend
WHERE classid = 'pg_operator'::regclass AND
      objid = '===(bool,bool)'::regoperator
ORDER BY 1;`,
				Results: []sql.Row{{`function alter_op_test_fn(boolean,boolean)`, `n`}, {`function customcontsel(internal,oid,internal,integer)`, `n`}, {`schema public`, `n`}},
			},
			{
				Statement: `ALTER OPERATOR === (boolean, boolean) SET (RESTRICT = NONE);`,
			},
			{
				Statement: `ALTER OPERATOR === (boolean, boolean) SET (JOIN = NONE);`,
			},
			{
				Statement: `SELECT oprrest, oprjoin FROM pg_operator WHERE oprname = '==='
  AND oprleft = 'boolean'::regtype AND oprright = 'boolean'::regtype;`,
				Results: []sql.Row{{`-`, `-`}},
			},
			{
				Statement: `SELECT pg_describe_object(refclassid,refobjid,refobjsubid) as ref, deptype
FROM pg_depend
WHERE classid = 'pg_operator'::regclass AND
      objid = '===(bool,bool)'::regoperator
ORDER BY 1;`,
				Results: []sql.Row{{`function alter_op_test_fn(boolean,boolean)`, `n`}, {`schema public`, `n`}},
			},
			{
				Statement: `ALTER OPERATOR === (boolean, boolean) SET (RESTRICT = contsel);`,
			},
			{
				Statement: `ALTER OPERATOR === (boolean, boolean) SET (JOIN = contjoinsel);`,
			},
			{
				Statement: `SELECT oprrest, oprjoin FROM pg_operator WHERE oprname = '==='
  AND oprleft = 'boolean'::regtype AND oprright = 'boolean'::regtype;`,
				Results: []sql.Row{{`contsel`, `contjoinsel`}},
			},
			{
				Statement: `SELECT pg_describe_object(refclassid,refobjid,refobjsubid) as ref, deptype
FROM pg_depend
WHERE classid = 'pg_operator'::regclass AND
      objid = '===(bool,bool)'::regoperator
ORDER BY 1;`,
				Results: []sql.Row{{`function alter_op_test_fn(boolean,boolean)`, `n`}, {`schema public`, `n`}},
			},
			{
				Statement: `ALTER OPERATOR === (boolean, boolean) SET (RESTRICT = NONE, JOIN = NONE);`,
			},
			{
				Statement: `SELECT oprrest, oprjoin FROM pg_operator WHERE oprname = '==='
  AND oprleft = 'boolean'::regtype AND oprright = 'boolean'::regtype;`,
				Results: []sql.Row{{`-`, `-`}},
			},
			{
				Statement: `SELECT pg_describe_object(refclassid,refobjid,refobjsubid) as ref, deptype
FROM pg_depend
WHERE classid = 'pg_operator'::regclass AND
      objid = '===(bool,bool)'::regoperator
ORDER BY 1;`,
				Results: []sql.Row{{`function alter_op_test_fn(boolean,boolean)`, `n`}, {`schema public`, `n`}},
			},
			{
				Statement: `ALTER OPERATOR === (boolean, boolean) SET (RESTRICT = customcontsel, JOIN = contjoinsel);`,
			},
			{
				Statement: `SELECT oprrest, oprjoin FROM pg_operator WHERE oprname = '==='
  AND oprleft = 'boolean'::regtype AND oprright = 'boolean'::regtype;`,
				Results: []sql.Row{{`customcontsel`, `contjoinsel`}},
			},
			{
				Statement: `SELECT pg_describe_object(refclassid,refobjid,refobjsubid) as ref, deptype
FROM pg_depend
WHERE classid = 'pg_operator'::regclass AND
      objid = '===(bool,bool)'::regoperator
ORDER BY 1;`,
				Results: []sql.Row{{`function alter_op_test_fn(boolean,boolean)`, `n`}, {`function customcontsel(internal,oid,internal,integer)`, `n`}, {`schema public`, `n`}},
			},
			{
				Statement:   `ALTER OPERATOR === (boolean, boolean) SET (COMMUTATOR = ====);`,
				ErrorString: `operator attribute "commutator" cannot be changed`,
			},
			{
				Statement:   `ALTER OPERATOR === (boolean, boolean) SET (NEGATOR = ====);`,
				ErrorString: `operator attribute "negator" cannot be changed`,
			},
			{
				Statement:   `ALTER OPERATOR === (boolean, boolean) SET (RESTRICT = non_existent_func);`,
				ErrorString: `function non_existent_func(internal, oid, internal, integer) does not exist`,
			},
			{
				Statement:   `ALTER OPERATOR === (boolean, boolean) SET (JOIN = non_existent_func);`,
				ErrorString: `function non_existent_func(internal, oid, internal, smallint, internal) does not exist`,
			},
			{
				Statement:   `ALTER OPERATOR === (boolean, boolean) SET (COMMUTATOR = !==);`,
				ErrorString: `operator attribute "commutator" cannot be changed`,
			},
			{
				Statement:   `ALTER OPERATOR === (boolean, boolean) SET (NEGATOR = !==);`,
				ErrorString: `operator attribute "negator" cannot be changed`,
			},
			{
				Statement:   `ALTER OPERATOR & (bit, bit) SET ("Restrict" = _int_contsel, "Join" = _int_contjoinsel);`,
				ErrorString: `operator attribute "Restrict" not recognized`,
			},
			{
				Statement: `CREATE USER regress_alter_op_user;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_alter_op_user;`,
			},
			{
				Statement:   `ALTER OPERATOR === (boolean, boolean) SET (RESTRICT = NONE);`,
				ErrorString: `must be owner of operator ===`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP USER regress_alter_op_user;`,
			},
			{
				Statement: `DROP OPERATOR === (boolean, boolean);`,
			},
			{
				Statement: `DROP FUNCTION customcontsel(internal, oid, internal, integer);`,
			},
			{
				Statement: `DROP FUNCTION alter_op_test_fn(boolean, boolean);`,
			},
		},
	})
}
