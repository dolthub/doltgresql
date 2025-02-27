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

func TestDropFunction(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Function does not exist",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "DROP FUNCTION doesnotexist;",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "DROP FUNCTION IF EXISTS doesnotexist;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Basic cases",
			SetUpScript: []string{`
CREATE FUNCTION func1() RETURNS TEXT AS $$
BEGIN RETURN 'func1'; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func2(input INT) RETURNS TEXT AS $$
BEGIN RETURN 'func2(INT)'; END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT func1(), func2(99);",
					Expected: []sql.Row{{"func1", "func2(INT)"}},
				},
				{
					Query:    "DROP FUNCTION func1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP FUNCTION func2(INT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT func1();",
					ExpectedErr: "not found",
				},
				{
					Query:       "SELECT func2(99);",
					ExpectedErr: "not found",
				},
			},
		},
		{
			Name: "Optional type information",
			SetUpScript: []string{`
CREATE FUNCTION func1() RETURNS TEXT AS $$
BEGIN RETURN 'func1'; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func2() RETURNS TEXT AS $$
BEGIN RETURN 'func2'; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func3(input INT) RETURNS TEXT AS $$
BEGIN RETURN 'func3(INT)'; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func4(input INT) RETURNS TEXT AS $$
BEGIN RETURN 'func4(INT)'; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func5(input INT, foo TEXT) RETURNS TEXT AS $$
BEGIN RETURN 'func5(INT, TEXT)'; END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT func1(), func2(), func3(1), func4(2);",
					Expected: []sql.Row{{"func1", "func2", "func3(INT)", "func4(INT)"}},
				},
				{
					Query:    "DROP FUNCTION func1(OUT TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP FUNCTION func2(OUT paramname TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP FUNCTION func3(paramname INT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP FUNCTION func4(IN paramname INT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP FUNCTION func5(IN paramname INT, IN paramname TEXT);",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Qualified names",
			SetUpScript: []string{`
CREATE FUNCTION func1() RETURNS TEXT AS $$
BEGIN RETURN 'func1'; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func2(input TEXT) RETURNS TEXT AS $$
BEGIN RETURN 'func2(TEXT)'; END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT current_schema(), current_database();",
					Expected: []sql.Row{{"public", "postgres"}},
				},
				{
					Query:    "SELECT func1(), func2('foo');",
					Expected: []sql.Row{{"func1", "func2(TEXT)"}},
				},
				{
					Query:    "DROP FUNCTION public.func1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT func1();",
					ExpectedErr: "not found",
				},
				{
					Query:    "DROP FUNCTION postgres.public.func2(TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT func2('w00t');",
					ExpectedErr: "not found",
				},
			},
		},
		{
			// When there is only one function with a name, the parameter types are not required,
			// but if the name is not unique, an error is returned.
			Name: "Unspecified parameter types",
			SetUpScript: []string{`
CREATE FUNCTION func1(input1 TEXT, input2 TEXT) RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func2(input1 TEXT) RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func2(input1 TEXT, input2 TEXT) RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func3(input1 TEXT, input2 TEXT) RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func3() RETURNS int AS $$
BEGIN RETURN 42; END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DROP FUNCTION func1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "DROP FUNCTION func2;",
					ExpectedErr: "is not unique",
				},
				{
					Query:    "DROP FUNCTION func3;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			// TODO: Postgres supports specifying multiple functions to drop, but our
			//       parser doesn't seem to support parsing multiple functions yet.
			Skip: true,
			Name: "Multiple functions",
			SetUpScript: []string{`
CREATE FUNCTION func1() RETURNS TEXT AS $$
BEGIN
	RETURN 'func1';
END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func2(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN 'func2(TEXT)';
END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT func1(), func2('foo');",
					Expected: []sql.Row{{"func1", "func2(TEXT)"}},
				},
				{
					Query:    "DELETE FUNCTION func1, func2(TExT);",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Overloaded functions",
			SetUpScript: []string{`
CREATE FUNCTION func2(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN 'func2(TEXT)';
END;
$$ LANGUAGE plpgsql;`, `
CREATE FUNCTION func2(input INT) RETURNS TEXT AS $$
BEGIN
	RETURN 'func2(INT)';
END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT func2('foo'), func2(42);",
					Expected: []sql.Row{{"func2(TEXT)", "func2(INT)"}},
				},
				{
					Query:    "DROP FUNCTION func2(TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT func2('foo'::text);",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "SELECT func2(42);",
					Expected: []sql.Row{{"func2(INT)"}},
				},
				{
					Query:    "DROP FUNCTION func2(INT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT func2(42);",
					ExpectedErr: "not found",
				},
			},
		},
	})
}
