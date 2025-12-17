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

func TestDropProcedure(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Procedure does not exist",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "DROP PROCEDURE doesnotexist;",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "DROP PROCEDURE IF EXISTS doesnotexist;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Basic cases",
			SetUpScript: []string{
				`CREATE PROCEDURE proc1() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc2(input INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CALL proc1();",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL proc2(99);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc2(INT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CALL proc1();",
					ExpectedErr: "does not exist",
				},
				{
					Query:       "CALL proc2(99);",
					ExpectedErr: "does not exist",
				},
			},
		},
		{
			Name: "Optional type information",
			SetUpScript: []string{
				`CREATE PROCEDURE proc1() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc2() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc3(input INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc4(input INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc5(input INT, foo TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CALL proc1();",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL proc2();",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL proc3(1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL proc4(2);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL proc5(3, 'abc');",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc1(OUT TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc2(OUT paramname TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc3(paramname INT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc4(IN paramname INT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc5(IN paramname INT, IN paramname TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CALL proc1();",
					ExpectedErr: "does not exist",
				},
				{
					Query:       "CALL proc2();",
					ExpectedErr: "does not exist",
				},
				{
					Query:       "CALL proc3(1);",
					ExpectedErr: "does not exist",
				},
				{
					Query:       "CALL proc4(2);",
					ExpectedErr: "does not exist",
				},
				{
					Query:       "CALL proc5(3, 'abc');",
					ExpectedErr: "does not exist",
				},
			},
		},
		{
			Name: "Qualified names",
			SetUpScript: []string{
				`CREATE PROCEDURE proc1() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc2(input TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT current_schema(), current_database();",
					Expected: []sql.Row{{"public", "postgres"}},
				},
				{
					Query:    "CALL proc1();",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL proc2('foo');",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE public.proc1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CALL proc1();",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "DROP PROCEDURE postgres.public.proc2(TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CALL proc2('bar');",
					ExpectedErr: "does not exist",
				},
			},
		},
		{
			Name: "Unspecified parameter types",
			SetUpScript: []string{
				`CREATE PROCEDURE proc1(input1 TEXT, input2 TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc2(input1 TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc2(input1 TEXT, input2 TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc3(input1 TEXT, input2 TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc3() AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DROP PROCEDURE proc1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "DROP PROCEDURE proc2;",
					ExpectedErr: "is not unique",
				},
				{
					Query:    "DROP PROCEDURE proc3;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Overloaded procedures",
			SetUpScript: []string{
				`CREATE PROCEDURE proc2(input TEXT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE proc2(input INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CALL proc2('foo');",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL proc2(42);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc2(TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CALL proc2('foo'::text);",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "CALL proc2(42);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP PROCEDURE proc2(INT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CALL proc2(42);",
					ExpectedErr: "does not exist",
				},
			},
		},
	})
}
