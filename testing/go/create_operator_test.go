// Copyright 2026 Dolthub, Inc.
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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCreateOperator(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "CREATE OPERATOR with LEFTARG, RIGHTARG, and FUNCTION",
			SetUpScript: []string{
				`CREATE FUNCTION op_int_dist(a int4, b int4) RETURNS int4
					AS $$ SELECT abs(a - b) $$ LANGUAGE SQL;`,
				`CREATE TABLE op_points (pk int4 PRIMARY KEY, v int4);`,
				`INSERT INTO op_points VALUES (1, 4), (2, 9), (3, 15);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_int_dist);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 3 <-> 10;`,
					Expected: []sql.Row{{7}},
				},
				{
					Query:    `SELECT 10 <-> 3;`,
					Expected: []sql.Row{{7}},
				},
				{ // NULL operands are passed through to the non-STRICT backing function
					Query:    `SELECT (NULL::int4 <-> 3) IS NULL;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pk, v <-> 10 FROM op_points ORDER BY pk;`,
					Expected: []sql.Row{{1, 6}, {2, 1}, {3, 5}},
				},
				{
					Query:    `SELECT pk FROM op_points WHERE v <-> 10 < 3 ORDER BY pk;`,
					Expected: []sql.Row{{2}},
				},
				{
					Query: `SELECT oprname, oprkind, oprcanhash, oprcanmerge, oprleft::regtype::text, oprright::regtype::text, oprresult::regtype::text
						FROM pg_operator WHERE oprcode::text = 'op_int_dist';`,
					Expected: []sql.Row{{"<->", "b", "f", "f", "integer", "integer", "integer"}},
				},
				{
					Query:    `SELECT n.nspname FROM pg_operator o JOIN pg_namespace n ON n.oid = o.oprnamespace WHERE o.oprcode::text = 'op_int_dist';`,
					Expected: []sql.Row{{"public"}},
				},
			},
		},
		{
			Name: "CREATE OPERATOR with COMMUTATOR and NEGATOR",
			SetUpScript: []string{
				`CREATE FUNCTION op_len_eq(a text, b text) RETURNS boolean
					AS $$ SELECT length(a) = length(b) $$ LANGUAGE SQL;`,
				`CREATE FUNCTION op_len_ne(a text, b text) RETURNS boolean
					AS $$ SELECT length(a) <> length(b) $$ LANGUAGE SQL;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE OPERATOR <=> (LEFTARG = text, RIGHTARG = text, FUNCTION = op_len_eq, COMMUTATOR = <=>);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE OPERATOR <~> (LEFTARG = text, RIGHTARG = text, FUNCTION = op_len_ne, COMMUTATOR = <~>, NEGATOR = <=>);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 'abc' <=> 'xyz';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'abc' <=> 'wxyz';`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'abc' <~> 'wxyz';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT c.oprname FROM pg_operator o JOIN pg_operator c ON c.oid = o.oprcom WHERE o.oprcode::text = 'op_len_eq';`,
					Expected: []sql.Row{{"<=>"}},
				},
				{ // Declaring the negator on one operator links both directions
					Query:    `SELECT n.oprname FROM pg_operator o JOIN pg_operator n ON n.oid = o.oprnegate WHERE o.oprcode::text = 'op_len_ne';`,
					Expected: []sql.Row{{"<=>"}},
				},
				{
					Query:    `SELECT n.oprname FROM pg_operator o JOIN pg_operator n ON n.oid = o.oprnegate WHERE o.oprcode::text = 'op_len_eq';`,
					Expected: []sql.Row{{"<~>"}},
				},
			},
		},
		{
			Name: "CREATE OPERATOR with HASHES and MERGES",
			SetUpScript: []string{
				`CREATE FUNCTION op_ci_eq(a text, b text) RETURNS boolean
					AS $$ SELECT lower(a) = lower(b) $$ LANGUAGE SQL;`,
			},
			Assertions: []ScriptTestAssertion{
				{ // We're not testing HASHES or MERGES, just that they parse since they're ignored options
					Query:    `CREATE OPERATOR <%> (LEFTARG = text, RIGHTARG = text, FUNCTION = op_ci_eq, COMMUTATOR = <%>, HASHES, MERGES);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 'ABC' <%> 'abc';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT oprcanhash, oprcanmerge FROM pg_operator WHERE oprcode::text = 'op_ci_eq';`,
					Expected: []sql.Row{{"t", "t"}},
				},
			},
		},
		{
			Name: "CREATE OPERATOR with composite operands",
			SetUpScript: []string{
				`CREATE TABLE op_pair (x int4, y int4);`,
				`CREATE FUNCTION op_pair_add(a op_pair, b op_pair) RETURNS op_pair
					AS $$ SELECT ROW((a).x + (b).x, (a).y + (b).y)::op_pair $$ LANGUAGE SQL;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE OPERATOR <+> (LEFTARG = op_pair, RIGHTARG = op_pair, FUNCTION = op_pair_add);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT ROW(1, 2)::op_pair <+> ROW(3, 4)::op_pair;`,
					Expected: []sql.Row{{"(4,6)"}},
				},
				{
					Query: `SELECT oprleft::regtype::text, oprright::regtype::text, oprresult::regtype::text
						FROM pg_operator WHERE oprcode::text = 'op_pair_add';`,
					Expected: []sql.Row{{"op_pair", "op_pair", "op_pair"}},
				},
			},
		},
		{
			Name: "CREATE OPERATOR with composite operands on table rows in aggregate",
			SetUpScript: []string{
				`CREATE TABLE op_pair (x int4, y int4);`,
				`CREATE FUNCTION op_pair_add(a op_pair, b op_pair) RETURNS op_pair
			AS $$ SELECT ROW((a).x + (b).x, (a).y + (b).y)::op_pair $$ LANGUAGE SQL;`,
				`CREATE OPERATOR <+> (LEFTARG = op_pair, RIGHTARG = op_pair, FUNCTION = op_pair_add);`,
				`INSERT INTO op_pair VALUES (1, 2), (3, 4), (5, 6);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT sum(((p <+> ROW(10, 20)::op_pair)).x), sum(((p <+> ROW(10, 20)::op_pair)).y)
			       FROM op_pair p;`,
					Expected: []sql.Row{{39, 72}},
				},
			},
		},
		{
			Name: "CREATE OPERATOR with mixed operand types",
			SetUpScript: []string{
				`CREATE FUNCTION op_repeat(a text, b int4) RETURNS text
					AS $$ SELECT repeat(a, b) $$ LANGUAGE SQL;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE OPERATOR <#> (LEFTARG = text, RIGHTARG = int4, FUNCTION = op_repeat);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 'ab' <#> 3;`,
					Expected: []sql.Row{{"ababab"}},
				},
				{
					Query:       `SELECT 3 <#> 'ab';`,
					ExpectedErr: "operator does not exist",
				},
			},
		},
		{
			Name: "DROP OPERATOR smoke test",
			SetUpScript: []string{
				`CREATE FUNCTION op_drop_dist(a int4, b int4) RETURNS int4
					AS $$ SELECT abs(a - b) $$ LANGUAGE SQL;`,
				`CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_drop_dist);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 3 <-> 10;`,
					Expected: []sql.Row{{7}},
				},
				{
					Query:    `DROP OPERATOR <-> (int4, int4);`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT 3 <-> 10;`,
					ExpectedErr: "operator does not exist: integer <-> integer",
				},
				{
					Query:    `SELECT EXISTS (SELECT 1 FROM pg_operator WHERE oprcode::text = 'op_drop_dist');`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_drop_dist);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT 3 <-> 10;`,
					Expected: []sql.Row{{7}},
				},
			},
		},
		{
			Name: "CREATE OPERATOR validation",
			SetUpScript: []string{
				`CREATE FUNCTION op_valid_dist(a int4, b int4) RETURNS int4
					AS $$ SELECT abs(a - b) $$ LANGUAGE SQL;`,
				`CREATE FUNCTION op_one_arg(a int4) RETURNS int4
					AS $$ SELECT a $$ LANGUAGE SQL;`,
				`CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_valid_dist);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `CREATE OPERATOR <=> (LEFTARG = int4, RIGHTARG = int4);`,
					ExpectedErr: "operator function must be specified",
				},
				{
					Query:       `CREATE OPERATOR <=> (FUNCTION = op_valid_dist);`,
					ExpectedErr: "operator argument types must be specified",
				},
				{
					Query:       `CREATE OPERATOR <=> (LEFTARG = int4, FUNCTION = op_valid_dist);`,
					ExpectedErr: "operator right argument type must be specified",
				},
				{
					Query:       `CREATE OPERATOR <=> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_missing_fn);`,
					ExpectedErr: "function op_missing_fn(integer, integer) does not exist",
				},
				{
					Query:       `CREATE OPERATOR <=> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_one_arg);`,
					ExpectedErr: "function op_one_arg(integer, integer) does not exist",
				},
				{
					Query:       `CREATE OPERATOR <-> (LEFTARG = int4, RIGHTARG = int4, FUNCTION = op_valid_dist);`,
					ExpectedErr: "operator <-> already exists",
				},
			},
		},
	})
}
