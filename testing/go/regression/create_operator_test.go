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

func TestCreateOperator(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_operator)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_operator,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE OPERATOR ## (
   leftarg = path,
   rightarg = path,
   function = path_inter,
   commutator = ##
);`,
			},
			{
				Statement: `CREATE OPERATOR @#@ (
   rightarg = int8,		-- prefix
   procedure = factorial
);`,
			},
			{
				Statement: `CREATE OPERATOR #%# (
   leftarg = int8,		-- fail, postfix is no longer supported
   procedure = factorial
);`,
				ErrorString: `operator right argument type must be specified`,
			},
			{
				Statement: `SELECT @#@ 24;`,
				Results:   []sql.Row{{620448401733239439360000.0}},
			},
			{
				Statement:   `COMMENT ON OPERATOR ###### (NONE, int4) IS 'bad prefix';`,
				ErrorString: `operator does not exist: ###### integer`,
			},
			{
				Statement:   `COMMENT ON OPERATOR ###### (int4, NONE) IS 'bad postfix';`,
				ErrorString: `postfix operators are not supported`,
			},
			{
				Statement:   `COMMENT ON OPERATOR ###### (int4, int8) IS 'bad infix';`,
				ErrorString: `operator does not exist: integer ###### bigint`,
			},
			{
				Statement:   `DROP OPERATOR ###### (NONE, int4);`,
				ErrorString: `operator does not exist: ###### integer`,
			},
			{
				Statement:   `DROP OPERATOR ###### (int4, NONE);`,
				ErrorString: `postfix operators are not supported`,
			},
			{
				Statement:   `DROP OPERATOR ###### (int4, int8);`,
				ErrorString: `operator does not exist: integer ###### bigint`,
			},
			{
				Statement: `CREATE OPERATOR => (
   rightarg = int8,
   procedure = factorial
);`,
				ErrorString: `syntax error at or near "=>"`,
			},
			{
				Statement: `CREATE OPERATOR !=- (
   rightarg = int8,
   procedure = factorial
);`,
			},
			{
				Statement: `SELECT !=- 10;`,
				Results:   []sql.Row{{3628800}},
			},
			{
				Statement:   `SELECT 10 !=-;`,
				ErrorString: `syntax error at or near ";"`,
			},
			{
				Statement: `SELECT 2 !=/**/ 1, 2 !=/**/ 2;`,
				Results:   []sql.Row{{true, false}},
			},
			{
				Statement: `SELECT 2 !=-- comment to be removed by psql
  1;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `DO $$ -- use DO to protect -- from psql
  declare r boolean;`,
			},
			{
				Statement: `  begin
    execute $e$ select 2 !=-- comment
      1 $e$ into r;`,
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
				Statement: `INFO:  r = t
SELECT true<>-1 BETWEEN 1 AND 1;  -- BETWEEN has prec. above <> but below Op`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT false<>/**/1 BETWEEN 1 AND 1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT false<=-1 BETWEEN 1 AND 1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT false>=-1 BETWEEN 1 AND 1;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT 2<=/**/3, 3>=/**/2, 2<>/**/3;`,
				Results:   []sql.Row{{true, true, true}},
			},
			{
				Statement: `SELECT 3<=/**/2, 2>=/**/3, 2<>/**/2;`,
				Results:   []sql.Row{{false, false, false}},
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE ROLE regress_rol_op1;`,
			},
			{
				Statement: `CREATE SCHEMA schema_op1;`,
			},
			{
				Statement: `GRANT USAGE ON SCHEMA schema_op1 TO PUBLIC;`,
			},
			{
				Statement: `REVOKE USAGE ON SCHEMA schema_op1 FROM regress_rol_op1;`,
			},
			{
				Statement: `SET ROLE regress_rol_op1;`,
			},
			{
				Statement: `CREATE OPERATOR schema_op1.#*# (
   rightarg = int8,
   procedure = factorial
);`,
				ErrorString: `permission denied for schema schema_op1`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE OPERATOR #*# (
   leftarg = SETOF int8,
   procedure = factorial
);`,
				ErrorString: `SETOF type not allowed for operator argument`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE OPERATOR #*# (
   rightarg = SETOF int8,
   procedure = factorial
);`,
				ErrorString: `SETOF type not allowed for operator argument`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE OR REPLACE FUNCTION fn_op2(boolean, boolean)
RETURNS boolean AS $$
    SELECT NULL::BOOLEAN;`,
			},
			{
				Statement: `$$ LANGUAGE sql IMMUTABLE;`,
			},
			{
				Statement: `CREATE OPERATOR === (
    LEFTARG = boolean,
    RIGHTARG = boolean,
    PROCEDURE = fn_op2,
    COMMUTATOR = ===,
    NEGATOR = !==,
    RESTRICT = contsel,
    JOIN = contjoinsel,
    SORT1, SORT2, LTCMP, GTCMP, HASHES, MERGES
);`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE OPERATOR #@%# (
   rightarg = int8,
   procedure = factorial,
   invalid_att = int8
);`,
			},
			{
				Statement: `CREATE OPERATOR #@%# (
   procedure = factorial
);`,
				ErrorString: `operator argument types must be specified`,
			},
			{
				Statement: `CREATE OPERATOR #@%# (
   rightarg = int8
);`,
				ErrorString: `operator function must be specified`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE ROLE regress_rol_op3;`,
			},
			{
				Statement: `CREATE TYPE type_op3 AS ENUM ('new', 'open', 'closed');`,
			},
			{
				Statement: `CREATE FUNCTION fn_op3(type_op3, int8)
RETURNS int8 AS $$
    SELECT NULL::int8;`,
			},
			{
				Statement: `$$ LANGUAGE sql IMMUTABLE;`,
			},
			{
				Statement: `REVOKE USAGE ON TYPE type_op3 FROM regress_rol_op3;`,
			},
			{
				Statement: `REVOKE USAGE ON TYPE type_op3 FROM PUBLIC;  -- Need to do this so that regress_rol_op3 is not allowed USAGE via PUBLIC`,
			},
			{
				Statement: `SET ROLE regress_rol_op3;`,
			},
			{
				Statement: `CREATE OPERATOR #*# (
   leftarg = type_op3,
   rightarg = int8,
   procedure = fn_op3
);`,
				ErrorString: `permission denied for type type_op3`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE ROLE regress_rol_op4;`,
			},
			{
				Statement: `CREATE TYPE type_op4 AS ENUM ('new', 'open', 'closed');`,
			},
			{
				Statement: `CREATE FUNCTION fn_op4(int8, type_op4)
RETURNS int8 AS $$
    SELECT NULL::int8;`,
			},
			{
				Statement: `$$ LANGUAGE sql IMMUTABLE;`,
			},
			{
				Statement: `REVOKE USAGE ON TYPE type_op4 FROM regress_rol_op4;`,
			},
			{
				Statement: `REVOKE USAGE ON TYPE type_op4 FROM PUBLIC;  -- Need to do this so that regress_rol_op3 is not allowed USAGE via PUBLIC`,
			},
			{
				Statement: `SET ROLE regress_rol_op4;`,
			},
			{
				Statement: `CREATE OPERATOR #*# (
   leftarg = int8,
   rightarg = type_op4,
   procedure = fn_op4
);`,
				ErrorString: `permission denied for type type_op4`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE ROLE regress_rol_op5;`,
			},
			{
				Statement: `CREATE TYPE type_op5 AS ENUM ('new', 'open', 'closed');`,
			},
			{
				Statement: `CREATE FUNCTION fn_op5(int8, int8)
RETURNS int8 AS $$
    SELECT NULL::int8;`,
			},
			{
				Statement: `$$ LANGUAGE sql IMMUTABLE;`,
			},
			{
				Statement: `REVOKE EXECUTE ON FUNCTION fn_op5(int8, int8) FROM regress_rol_op5;`,
			},
			{
				Statement: `REVOKE EXECUTE ON FUNCTION fn_op5(int8, int8) FROM PUBLIC;-- Need to do this so that regress_rol_op3 is not allowed EXECUTE via PUBLIC`,
			},
			{
				Statement: `SET ROLE regress_rol_op5;`,
			},
			{
				Statement: `CREATE OPERATOR #*# (
   leftarg = int8,
   rightarg = int8,
   procedure = fn_op5
);`,
				ErrorString: `permission denied for function fn_op5`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN TRANSACTION;`,
			},
			{
				Statement: `CREATE ROLE regress_rol_op6;`,
			},
			{
				Statement: `CREATE TYPE type_op6 AS ENUM ('new', 'open', 'closed');`,
			},
			{
				Statement: `CREATE FUNCTION fn_op6(int8, int8)
RETURNS type_op6 AS $$
    SELECT NULL::type_op6;`,
			},
			{
				Statement: `$$ LANGUAGE sql IMMUTABLE;`,
			},
			{
				Statement: `REVOKE USAGE ON TYPE type_op6 FROM regress_rol_op6;`,
			},
			{
				Statement: `REVOKE USAGE ON TYPE type_op6 FROM PUBLIC;  -- Need to do this so that regress_rol_op3 is not allowed USAGE via PUBLIC`,
			},
			{
				Statement: `SET ROLE regress_rol_op6;`,
			},
			{
				Statement: `CREATE OPERATOR #*# (
   leftarg = int8,
   rightarg = int8,
   procedure = fn_op6
);`,
				ErrorString: `permission denied for type type_op6`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `CREATE OPERATOR ===
(
	"Leftarg" = box,
	"Rightarg" = box,
	"Procedure" = area_equal_function,
	"Commutator" = ===,
	"Negator" = !==,
	"Restrict" = area_restriction_function,
	"Join" = area_join_function,
	"Hashes",
	"Merges"
);`,
				ErrorString: `operator function must be specified`,
			},
		},
	})
}
