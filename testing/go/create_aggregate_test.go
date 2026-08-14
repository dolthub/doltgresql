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

func TestCreateAggregate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "CREATE AGGREGATE with SFUNC and STYPE",
			SetUpScript: []string{
				`CREATE FUNCTION agg_sum_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + val $$ LANGUAGE SQL;`,
				`CREATE TABLE agg_sum_vals (pk int4 PRIMARY KEY, grp text, v int4);`,
				`INSERT INTO agg_sum_vals VALUES (1, 'a', 10), (2, 'a', 20), (3, 'b', 5), (4, 'b', NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE AGGREGATE agg_sum (int4) (SFUNC = agg_sum_step, STYPE = int4, INITCOND = '0');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT agg_sum(v) FROM agg_sum_vals WHERE grp = 'a';`,
					Expected: []sql.Row{{30}},
				},
				{ // The transition function is not STRICT, so the NULL in group 'b' nulls the state
					Query:    `SELECT grp, agg_sum(v) FROM agg_sum_vals GROUP BY grp ORDER BY grp;`,
					Expected: []sql.Row{{"a", 30}, {"b", nil}},
				},
				{ // The initial condition is the state when the aggregate sees no rows at all
					Query:    `SELECT agg_sum(v) FROM agg_sum_vals WHERE pk = 0;`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT aggkind, agginitval, aggfinalfn::text, aggcombinefn::text FROM pg_aggregate WHERE aggtransfn::text = 'agg_sum_step';`,
					Expected: []sql.Row{{"n", "0", "-", "-"}},
				},
				{
					Query:    `SELECT proname, prokind FROM pg_proc WHERE proname = 'agg_sum';`,
					Expected: []sql.Row{{"agg_sum", "a"}},
				},
			},
		},
		{
			Name: "CREATE AGGREGATE with STRICT transition function and no INITCOND",
			SetUpScript: []string{
				`CREATE FUNCTION agg_larger_step(state int4, val int4) RETURNS int4
					AS $$ SELECT CASE WHEN state > val THEN state ELSE val END $$ LANGUAGE SQL STRICT;`,
				`CREATE TABLE agg_larger_vals (pk int4 PRIMARY KEY, v int4);`,
				`INSERT INTO agg_larger_vals VALUES (1, 3), (2, NULL), (3, 8), (4, 5);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE AGGREGATE agg_larger (int4) (SFUNC = agg_larger_step, STYPE = int4);`,
					Expected: []sql.Row{},
				},
				{ // A STRICT transition function skips NULL inputs, and the first non-NULL value seeds the state
					Query:    `SELECT agg_larger(v) FROM agg_larger_vals;`,
					Expected: []sql.Row{{8}},
				},
				{
					Query:    `SELECT agg_larger(v) FROM agg_larger_vals WHERE pk = 2;`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT agg_larger(v) FROM agg_larger_vals WHERE pk = 0;`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT agginitval IS NULL FROM pg_aggregate WHERE aggtransfn::text = 'agg_larger_step';`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "CREATE AGGREGATE with FINALFUNC",
			SetUpScript: []string{
				`CREATE FUNCTION agg_charcount_step(state int4, val text) RETURNS int4
					AS $$ SELECT state + length(val) $$ LANGUAGE SQL;`,
				`CREATE FUNCTION agg_charcount_final(state int4) RETURNS text
					AS $$ SELECT 'chars: ' || state $$ LANGUAGE SQL;`,
				`CREATE TABLE agg_charcount_vals (pk int4 PRIMARY KEY, v text);`,
				`INSERT INTO agg_charcount_vals VALUES (1, 'ab'), (2, 'cde'), (3, 'f');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE AGGREGATE agg_charcount (text) (SFUNC = agg_charcount_step, STYPE = int4,
						FINALFUNC = agg_charcount_final, INITCOND = '0');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT agg_charcount(v) FROM agg_charcount_vals;`,
					Expected: []sql.Row{{"chars: 6"}},
				},
				{ // The final function still runs when the aggregate sees no rows
					Query:    `SELECT agg_charcount(v) FROM agg_charcount_vals WHERE pk = 0;`,
					Expected: []sql.Row{{"chars: 0"}},
				},
				{
					Query:    `SELECT aggfinalfn::text FROM pg_aggregate WHERE aggtransfn::text = 'agg_charcount_step';`,
					Expected: []sql.Row{{"agg_charcount_final"}},
				},
			},
		},
		{
			Name: "CREATE AGGREGATE with COMBINEFUNC",
			SetUpScript: []string{
				`CREATE FUNCTION agg_combined_step(state int8, val int4) RETURNS int8
					AS $$ SELECT state + val $$ LANGUAGE SQL;`,
				`CREATE FUNCTION agg_combined_merge(s1 int8, s2 int8) RETURNS int8
					AS $$ SELECT s1 + s2 $$ LANGUAGE SQL;`,
				`CREATE TABLE agg_combined_vals (pk int4 PRIMARY KEY, v int4);`,
				`INSERT INTO agg_combined_vals VALUES (1, 10), (2, 20), (3, 5);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE AGGREGATE agg_combined (int4) (SFUNC = agg_combined_step, STYPE = int8,
						COMBINEFUNC = agg_combined_merge, INITCOND = '0');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT agg_combined(v) FROM agg_combined_vals;`,
					Expected: []sql.Row{{35}},
				},
				{
					Query:    `SELECT aggcombinefn::text FROM pg_aggregate WHERE aggtransfn::text = 'agg_combined_step';`,
					Expected: []sql.Row{{"agg_combined_merge"}},
				},
			},
		},
		{
			Name: "CREATE AGGREGATE with multiple arguments",
			SetUpScript: []string{
				`CREATE FUNCTION agg_wsum_step(state int8, val int4, weight int4) RETURNS int8
					AS $$ SELECT state + (val * weight) $$ LANGUAGE SQL;`,
				`CREATE TABLE agg_wsum_vals (pk int4 PRIMARY KEY, v int4, w int4);`,
				`INSERT INTO agg_wsum_vals VALUES (1, 10, 1), (2, 20, 2), (3, 30, 3);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE AGGREGATE agg_wsum (int4, int4) (SFUNC = agg_wsum_step, STYPE = int8, INITCOND = '0');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT agg_wsum(v, w) FROM agg_wsum_vals;`,
					Expected: []sql.Row{{140}},
				},
			},
		},
		{
			Name: "CREATE AGGREGATE in a custom schema",
			SetUpScript: []string{
				`CREATE SCHEMA agg_nsp;`,
				`CREATE FUNCTION agg_nsp_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + 1 $$ LANGUAGE SQL;`,
				`CREATE TABLE agg_nsp_vals (pk int4 PRIMARY KEY, v int4);`,
				`INSERT INTO agg_nsp_vals VALUES (1, 10), (2, 20), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE AGGREGATE agg_nsp.agg_rows (int4) (SFUNC = agg_nsp_step, STYPE = int4, INITCOND = '0');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT agg_nsp.agg_rows(v) FROM agg_nsp_vals;`,
					Expected: []sql.Row{{3}},
				},
				{
					Query:    `SELECT n.nspname FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE p.proname = 'agg_rows';`,
					Expected: []sql.Row{{"agg_nsp"}},
				},
				{
					Query:       `SELECT agg_rows(v) FROM agg_nsp_vals;`,
					ExpectedErr: "function: 'agg_rows' not found",
				},
			},
		},
		{
			Name: "CREATE OR REPLACE AGGREGATE",
			SetUpScript: []string{
				`CREATE FUNCTION agg_replace_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + val $$ LANGUAGE SQL;`,
				`CREATE TABLE agg_replace_vals (pk int4 PRIMARY KEY, v int4);`,
				`INSERT INTO agg_replace_vals VALUES (1, 10), (2, 20);`,
				`CREATE AGGREGATE agg_replace (int4) (SFUNC = agg_replace_step, STYPE = int4, INITCOND = '0');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT agg_replace(v) FROM agg_replace_vals;`,
					Expected: []sql.Row{{30}},
				},
				{
					Query:    `CREATE OR REPLACE AGGREGATE agg_replace (int4) (SFUNC = agg_replace_step, STYPE = int4, INITCOND = '100');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT agg_replace(v) FROM agg_replace_vals;`,
					Expected: []sql.Row{{130}},
				},
				{
					Query:    `SELECT agginitval FROM pg_aggregate WHERE aggtransfn::text = 'agg_replace_step';`,
					Expected: []sql.Row{{"100"}},
				},
			},
		},
		{
			Name: "DROP AGGREGATE smoke test",
			SetUpScript: []string{
				`CREATE FUNCTION agg_drop_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + val $$ LANGUAGE SQL;`,
				`CREATE TABLE agg_drop_vals (pk int4 PRIMARY KEY, v int4);`,
				`INSERT INTO agg_drop_vals VALUES (1, 10), (2, 20);`,
				`CREATE AGGREGATE agg_drop (int4) (SFUNC = agg_drop_step, STYPE = int4, INITCOND = '0');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT agg_drop(v) FROM agg_drop_vals;`,
					Expected: []sql.Row{{30}},
				},
				{
					Query:    `DROP AGGREGATE agg_drop(int4);`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT agg_drop(v) FROM agg_drop_vals;`,
					ExpectedErr: "function: 'agg_drop' not found",
				},
				{
					Query:    `SELECT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'agg_drop');`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `CREATE AGGREGATE agg_drop (int4) (SFUNC = agg_drop_step, STYPE = int4, INITCOND = '0');`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT agg_drop(v) FROM agg_drop_vals;`,
					Expected: []sql.Row{{30}},
				},
			},
		},
		{
			Name: "CREATE AGGREGATE validation",
			SetUpScript: []string{
				`CREATE FUNCTION agg_valid_step(state int4, val int4) RETURNS int4
					AS $$ SELECT state + val $$ LANGUAGE SQL;`,
				`CREATE FUNCTION agg_one_arg(state int4) RETURNS int4
					AS $$ SELECT state $$ LANGUAGE SQL;`,
				`CREATE FUNCTION agg_wrong_ret(state int4, val int4) RETURNS text
					AS $$ SELECT 'x' $$ LANGUAGE SQL;`,
				`CREATE FUNCTION agg_taken(a int4) RETURNS int4
					AS $$ SELECT a $$ LANGUAGE SQL;`,
				`CREATE AGGREGATE agg_valid (int4) (SFUNC = agg_valid_step, STYPE = int4);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `CREATE AGGREGATE agg_bad1 (int4) (SFUNC = agg_missing_step, STYPE = int4);`,
					ExpectedErr: "function agg_missing_step(integer, integer) does not exist",
				},
				{
					Query:       `CREATE AGGREGATE agg_bad2 (int4) (SFUNC = agg_one_arg, STYPE = int4);`,
					ExpectedErr: "function agg_one_arg(integer, integer) does not exist",
				},
				{
					Query:       `CREATE AGGREGATE agg_bad3 (int4) (SFUNC = agg_wrong_ret, STYPE = int4);`,
					ExpectedErr: "return type of transition function agg_wrong_ret is not integer",
				},
				{
					Query:       `CREATE AGGREGATE agg_bad4 (int4) (SFUNC = agg_valid_step, STYPE = int4, FINALFUNC = agg_missing_final);`,
					ExpectedErr: "function agg_missing_final(integer) does not exist",
				},
				{
					Query:       `CREATE AGGREGATE agg_valid (int4) (SFUNC = agg_valid_step, STYPE = int4);`,
					ExpectedErr: `function "agg_valid" already exists with same argument types`,
				},
				{
					Query:       `CREATE AGGREGATE agg_taken (int4) (SFUNC = agg_valid_step, STYPE = int4);`,
					ExpectedErr: `function "agg_taken" already exists with same argument types`,
				},
				{
					Query:       `CREATE AGGREGATE agg_bad5 (OUT x int4) (SFUNC = agg_valid_step, STYPE = int4);`,
					ExpectedErr: "aggregates cannot have output arguments",
				},
			},
		},
	})
}
