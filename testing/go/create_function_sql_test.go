// Copyright 2025 Dolthub, Inc.
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

func TestCreateFunctionsLanguageSQL(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "unnamed parameter",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE FUNCTION alt_func1(int) RETURNS int LANGUAGE sql AS 'SELECT $1 + 1';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT alt_func1(3);`,
					Expected: []sql.Row{{4}},
				},
			},
		},
		{
			Name:        "named parameter",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE FUNCTION alt_func1(x int) RETURNS int LANGUAGE sql AS 'SELECT x + 1';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT alt_func1(3);`,
					Expected: []sql.Row{{4}},
				},
				{
					Query:    `CREATE FUNCTION sub_numbers(x int, y int) RETURNS int LANGUAGE sql AS 'SELECT y - x';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT sub_numbers(1, 2);`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name:        "unknown functions",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION get_grade_description(score INT)
							RETURNS TEXT
							LANGUAGE SQL
							AS $$
								SELECT
									CASE
										WHEN score >= 90 THEN 'Excellent'
										WHEN score >= 75 THEN 'Good'
										WHEN score >= 50 THEN 'Average'
									ELSE 'Fail'
									END;
							$$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT get_grade_description(92);`,
					Expected: []sql.Row{{"Excellent"}},
				},
				{
					Query:    `SELECT get_grade_description(65);`,
					Expected: []sql.Row{{"Average"}},
				},
			},
		},
		{
			Name:        "nested functions",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION calculate_double_sum(x INT, y INT)
							RETURNS INT
							LANGUAGE SQL
							AS $$
								SELECT add_numbers(x, y) * 2;
							$$;`,
					// TODO: error message should be:  function add_numbers(integer, integer) does not exist
					ExpectedErr: "function: 'add_numbers' not found",
				},
				{
					Query:    `CREATE FUNCTION add_numbers(int, int) RETURNS int LANGUAGE sql AS 'SELECT $1 + $2';`,
					Expected: []sql.Row{},
				},
				{
					Query: `CREATE FUNCTION calculate_double_sum(x INT, y INT)
							RETURNS INT
							LANGUAGE SQL
							AS $$
								SELECT add_numbers(x, y) * 2;
							$$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT calculate_double_sum(1, 2);`,
					Expected: []sql.Row{{6}},
				},
			},
		},
		{
			Name:        "function returning multiple rows",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE FUNCTION gen(a int) RETURNS SETOF INT LANGUAGE SQL AS $$ SELECT generate_series(1, a) $$ STABLE;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM gen(3);`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
			},
		},
	})
}
