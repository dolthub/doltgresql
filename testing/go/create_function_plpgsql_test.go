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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCreateFunctionLanguagePlpgsql(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "ALIAS",
			// TODO: Implement OpCode conversion for parsed ALIAS statements.
			Skip: true,
			SetUpScript: []string{
				`CREATE FUNCTION interpreted_alias(input TEXT)
				RETURNS TEXT AS $$
				DECLARE
					var1 TEXT;
					var2 TEXT;
				BEGIN
					DECLARE
						alias1 ALIAS FOR var1;
						alias2 ALIAS FOR alias1;
						alias3 ALIAS FOR input;
					BEGIN
						alias2 := alias3;
					END;
					RETURN var1;
				END;
				$$ LANGUAGE plpgsql;
				`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_alias('123');",
					Expected: []sql.Row{{"123"}},
				},
			},
		},
		{
			Name: "Assignment",
			SetUpScript: []string{`CREATE FUNCTION interpreted_assignment(input TEXT) RETURNS TEXT AS $$
DECLARE
	var1 TEXT;
BEGIN
	var1 := 'Initial: ' || input;
	IF input = 'Hello' THEN
		var1 := var1 || ' - Greeting';
	ELSIF input = 'Bye' THEN
		var1 := var1 || ' - Farewell';
	ELSIF length(input) > 5 THEN
		var1 := var1 || ' - Over 5';
	ELSE
		var1 := var1 || ' - Else';
	END IF;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_assignment('Hello');",
					Expected: []sql.Row{
						{"Initial: Hello - Greeting"},
					},
				},
				{
					Query: "SELECT interpreted_assignment('Bye');",
					Expected: []sql.Row{
						{"Initial: Bye - Farewell"},
					},
				},
				{
					Query: "SELECT interpreted_assignment('abc');",
					Expected: []sql.Row{
						{"Initial: abc - Else"},
					},
				},
				{
					Query: "SELECT interpreted_assignment('something');",
					Expected: []sql.Row{
						{"Initial: something - Over 5"},
					},
				},
			},
		},
		{
			Name: "CASE, with ELSE",
			SetUpScript: []string{`
CREATE FUNCTION interpreted_case(x INT) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE x
		WHEN 1, 2 THEN
			msg := 'one';
			msg := msg || ' or two';
		ELSE
			msg := 'other';
			msg := msg || ' value than one or two';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_case(1);",
					Expected: []sql.Row{{"one or two"}},
				},
				{
					Query:    "SELECT interpreted_case(2);",
					Expected: []sql.Row{{"one or two"}},
				},
				{
					Query:    "SELECT interpreted_case(0);",
					Expected: []sql.Row{{"other value than one or two"}},
				},
			},
		},
		{
			// TODO: When no CASE statements match, and there is no ELSE block,
			//       Postgres raises an exception. Unskip this test after we
			//       add support for raising exceptions from functions.
			Skip: true,
			Name: "CASE, without ELSE",
			SetUpScript: []string{`
CREATE FUNCTION interpreted_case(x INT) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE x
		WHEN 1, 2 THEN
			msg := 'one or two';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_case(1);",
					Expected: []sql.Row{{"one or two"}},
				},
				{
					Query:    "SELECT interpreted_case(2);",
					Expected: []sql.Row{{"one or two"}},
				},
				{
					Query:       "SELECT interpreted_case(0);",
					ExpectedErr: "case not found",
				},
			},
		},
		{
			Name: "Searched CASE, with ELSE",
			SetUpScript: []string{`
CREATE FUNCTION interpreted_case(x INT) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE
		WHEN x BETWEEN 0 AND 10 THEN
			msg := 'value is between zero';
			msg := msg || ' and ten';
		WHEN x BETWEEN 11 AND 20 THEN
			msg := 'value is between eleven and twenty';
		ELSE
			msg := 'value';
			msg := msg || ' is';
			msg := msg || ' out of';
			msg := msg || ' bounds';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_case(0);",
					Expected: []sql.Row{{"value is between zero and ten"}},
				},
				{
					Query:    "SELECT interpreted_case(1);",
					Expected: []sql.Row{{"value is between zero and ten"}},
				},
				{
					Query:    "SELECT interpreted_case(10);",
					Expected: []sql.Row{{"value is between zero and ten"}},
				},
				{
					Query:    "SELECT interpreted_case(11);",
					Expected: []sql.Row{{"value is between eleven and twenty"}},
				},
				{
					Query:    "SELECT interpreted_case(21);",
					Expected: []sql.Row{{"value is out of bounds"}},
				},
			},
		},
		{
			// TODO: When no CASE statements match, and there is no ELSE block,
			//       Postgres raises an exception. Unskip this test after we
			//       add support for raising exceptions from functions.
			Skip: true,
			Name: "Searched CASE, without ELSE",
			SetUpScript: []string{`
CREATE FUNCTION interpreted_case(x INT) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE
		WHEN x BETWEEN 0 AND 10 THEN
			msg := 'value is between zero and ten';
		WHEN x BETWEEN 11 AND 20 THEN
			msg := 'value';
			msg := msg || ' is between';
			msg := msg || ' eleven and';
			msg := msg || ' twenty';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_case(0);",
					Expected: []sql.Row{{"value is between zero and ten"}},
				},
				{
					Query:    "SELECT interpreted_case(1);",
					Expected: []sql.Row{{"value is between zero and ten"}},
				},
				{
					Query:    "SELECT interpreted_case(10);",
					Expected: []sql.Row{{"value is between zero and ten"}},
				},
				{
					Query:    "SELECT interpreted_case(11);",
					Expected: []sql.Row{{"value is between eleven and twenty"}},
				},
				{
					Query:       "SELECT interpreted_case(21);",
					ExpectedErr: "case not found",
				},
			},
		},
		{
			Name: "CONTINUE",
			SetUpScript: []string{`CREATE FUNCTION interpreted_continue() RETURNS INT4 AS $$
DECLARE
	var1 INT4;
BEGIN
	LOOP
		var1 := var1 + 1;
		IF var1 < 4 THEN
			CONTINUE;
		END IF;
		RETURN var1;
	END LOOP;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_continue();",
					Expected: []sql.Row{{4}},
				},
			},
		},
		{
			Name: "CONTINUE Label",
			SetUpScript: []string{`CREATE FUNCTION interpreted_continue_label() RETURNS INT4 AS $$
DECLARE
	var1 INT4;
BEGIN
	<<cont_label>>
	LOOP
		var1 := var1 + 1;
		IF var1 < 6 THEN
			CONTINUE cont_label;
		END IF;
		RETURN var1;
	END LOOP;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_continue_label();",
					Expected: []sql.Row{{6}},
				},
			},
		},
		{
			Name: "EXIT",
			SetUpScript: []string{`CREATE FUNCTION interpreted_exit() RETURNS INT4 AS $$
DECLARE
	var1 INT4;
BEGIN
	LOOP
		var1 := var1 + 1;
		IF var1 >= 8 THEN
			EXIT;
		END IF;
	END LOOP;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_exit();",
					Expected: []sql.Row{{8}},
				},
			},
		},
		{
			Name: "EXIT WHEN",
			SetUpScript: []string{`CREATE FUNCTION interpreted_exit_when() RETURNS INT4 AS $$
DECLARE
	var1 INT4;
BEGIN
	LOOP
		var1 := var1 + 1;
		EXIT WHEN var1 >= 9;
	END LOOP;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_exit_when();",
					Expected: []sql.Row{{9}},
				},
			},
		},
		{
			Name: "LOOP",
			SetUpScript: []string{`CREATE FUNCTION interpreted_loop() RETURNS INT4 AS $$
DECLARE
	var1 INT4;
BEGIN
	LOOP
		var1 := var1 + 1;
		IF var1 >= 10 THEN
			RETURN var1;
		END IF;
	END LOOP;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_loop();",
					Expected: []sql.Row{{10}},
				},
			},
		},
		{
			Name: "LOOP Label",
			SetUpScript: []string{`CREATE FUNCTION interpreted_loop_label() RETURNS INT4 AS $$
DECLARE
	var1 INT4;
BEGIN
	<<loop_label>>
	LOOP
		var1 := var1 + 1;
		IF var1 >= 12 THEN
			EXIT loop_label;
		END IF;
	END LOOP;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_loop_label();",
					Expected: []sql.Row{{12}},
				},
			},
		},
		{
			Name: "PERFORM",
			SetUpScript: []string{
				`CREATE SEQUENCE test_sequence;`,
				`CREATE FUNCTION interpreted_perform() RETURNS VOID AS $$
BEGIN
	PERFORM nextval('test_sequence');
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT nextval('test_sequence');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT interpreted_perform();",
					Expected: []sql.Row{{nil}}, // TODO: Postgres returns a value that's not null, but also not a value?
				},
				{
					Query:    "SELECT nextval('test_sequence');",
					Expected: []sql.Row{{3}},
				},
			},
		},
		{
			Name: "RETURNS SETOF",
			SetUpScript: []string{
				`CREATE TYPE user_summary AS (
					user_id   integer,
					username  text,
					is_active boolean);`,
				`CREATE OR REPLACE FUNCTION func2() RETURNS SETOF user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 1, 'username', true;
						RETURN QUERY SELECT 2, 'another', false;
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func2();",
					Expected: []sql.Row{
						{"(1,username,t)"},
						{"(2,another,f)"},
					},
				},
				{
					Query: "SELECT func2(), func2();",
					Expected: []sql.Row{
						{"(1,username,t)", "(1,username,t)"},
						{"(2,another,f)", "(2,another,f)"},
					},
				},
			},
		},
		{
			Name: "RETURNS SETOF with no results",
			SetUpScript: []string{
				`CREATE TABLE user_summary (user_id integer, username text, is_active boolean);`,
				`CREATE OR REPLACE FUNCTION func2() RETURNS SETOF user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT * from user_summary;
						RETURN QUERY SELECT * from user_summary;
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT func2();",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT func2(), func2();",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "RETURNS SETOF with type from other schema",
			SetUpScript: []string{
				`CREATE SCHEMA sch1;`,
				`CREATE TYPE sch1.user_summary AS (
					user_id   integer,
					username  text,
					is_active boolean);`,
				`CREATE OR REPLACE FUNCTION func2() RETURNS SETOF sch1.user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 1, 'username', true;
						RETURN QUERY SELECT 2, 'another', false;
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func2();",
					Expected: []sql.Row{
						{"(1,username,t)"},
						{"(2,another,f)"},
					},
				},
				{
					Query: "SELECT func2(), func2();",
					Expected: []sql.Row{
						{"(1,username,t)", "(1,username,t)"},
						{"(2,another,f)", "(2,another,f)"},
					},
				},
			},
		},
		{
			Name: "RETURNS SETOF with param",
			SetUpScript: []string{
				`CREATE TYPE user_summary AS (
					user_id   integer,
					username  text,
					is_active boolean);`,
				`CREATE OR REPLACE FUNCTION func3(user_id integer) RETURNS SETOF user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT user_id, 'username', true;
						RETURN QUERY SELECT user_id, 'another', false;
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func3(111);",
					Expected: []sql.Row{
						{"(111,username,t)"},
						{"(111,another,f)"},
					},
				},
				{
					Query: "SELECT func3(111), func3(222);",
					Expected: []sql.Row{
						{"(111,username,t)", "(222,username,t)"},
						{"(111,another,f)", "(222,another,f)"},
					},
				},
			},
		},
		{
			Name: "RETURNS TABLE",
			SetUpScript: []string{
				`CREATE FUNCTION func2() RETURNS TABLE(user_id integer, username  text, is_active boolean)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 1, 'username', true;
						RETURN QUERY SELECT 2, 'another', false;
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func2();",
					Expected: []sql.Row{
						{"(1,username,t)"},
						{"(2,another,f)"},
					},
				},
				{
					Query: "SELECT func2(), func2();",
					Expected: []sql.Row{
						{"(1,username,t)", "(1,username,t)"},
						{"(2,another,f)", "(2,another,f)"},
					},
				},
			},
		},
		{
			Name: "RETURNS TABLE with single field",
			SetUpScript: []string{
				`CREATE FUNCTION func2() RETURNS TABLE(username text)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 'username1';
						RETURN QUERY SELECT 'username2';
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func2();",
					Expected: []sql.Row{
						{"(username1)"},
						{"(username2)"},
					},
				},
				{
					Query: "SELECT func2(), func2();",
					Expected: []sql.Row{
						{"(username1)", "(username1)"},
						{"(username2)", "(username2)"},
					},
				},
			},
		},
		{
			Name: "RETURNS TABLE with types from other schema",
			SetUpScript: []string{
				`CREATE SCHEMA sch1;`,
				`CREATE TYPE sch1.mytype AS (
					user_id   integer,
					username  text);`,
				`CREATE FUNCTION func2() RETURNS TABLE(foo sch1.mytype)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 1, 'username1';
						RETURN QUERY SELECT 2, 'username2';
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func2();",
					Expected: []sql.Row{
						{"(1,username1)"},
						{"(2,username2)"},
					},
				},
				{
					Query: "SELECT func2(), func2();",
					Expected: []sql.Row{
						{"(1,username1)", "(1,username1)"},
						{"(2,username2)", "(2,username2)"},
					},
				},
			},
		},
		{
			Name: "RETURNS TABLE with param",
			SetUpScript: []string{
				`CREATE OR REPLACE FUNCTION func3(user_id integer) RETURNS TABLE(user_id integer, username  text, is_active boolean)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT user_id, 'username', true;
						RETURN QUERY SELECT user_id, 'another', false;
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func3(111);",
					Expected: []sql.Row{
						{"(111,username,t)"},
						{"(111,another,f)"},
					},
				},
				{
					Query: "SELECT func3(111), func3(222);",
					Expected: []sql.Row{
						{"(111,username,t)", "(222,username,t)"},
						{"(111,another,f)", "(222,another,f)"},
					},
				},
			},
		},
		{
			Name: "RETURNS TABLE with join query",
			SetUpScript: []string{
				`CREATE TABLE customers (
					id INT PRIMARY KEY,
					name TEXT
				);`,
				`CREATE TABLE orders (
					id SERIAL PRIMARY KEY,
					customer_id INT,
					amount INT
				);`,
				`INSERT INTO customers VALUES (1, 'John'), (2, 'Jane');`,
				`INSERT INTO orders VALUES (1, 1, 100), (2, 2, 10);`,
				`CREATE OR REPLACE FUNCTION func2(n INT) RETURNS TABLE (c_id INT, c_name TEXT, c_total_spent INT) 
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY
						SELECT c.id,
							   c.name,
							   SUM(o.amount) AS total_spent
						FROM customers c
						JOIN orders o ON o.customer_id = c.id
						GROUP BY c.id, c.name
						HAVING SUM(o.amount) >= n
						;
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func2(1);",
					Expected: []sql.Row{
						{"(1,John,100)"},
						{"(2,Jane,10)"},
					},
				},
				{
					Query: "SELECT func2(11);",
					Expected: []sql.Row{
						{"(1,John,100)"},
					},
				},
				{
					Query:    "SELECT func2(111);",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "RETURNS SETOF with composite param",
			SetUpScript: []string{
				`CREATE TYPE user_summary AS (
					user_id   integer,
					username  text,
					is_active boolean);`,
				`CREATE OR REPLACE FUNCTION func3(u user_summary) RETURNS SETOF user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT u.user_id, u.username, u.is_active;
					END;
					$$;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT func3((222,'passedin',false)::user_summary);",
					Expected: []sql.Row{
						{"(222,passedin,f)"},
					},
				},
			},
		},
		{
			Name: "RAISE",
			SetUpScript: []string{
				`CREATE FUNCTION interpreted_raise1(input TEXT) RETURNS TEXT AS $$
				DECLARE
					var1 TEXT;
				BEGIN
					RAISE WARNING 'MyMessage';
					RAISE NOTICE USING MESSAGE = 'MyNoticeMessage';
					RAISE DEBUG 'DebugTest1' USING MESSAGE = 'DebugMessage';
					var1 := input;
					RETURN var1;
				END;
				$$ LANGUAGE plpgsql;`, // TODO: this is incorrect, cannot raise both NOTICE USING MESSAGE and DEBUG USING MESSAGE, need to fix
				`CREATE FUNCTION interpreted_raise2(input TEXT) RETURNS TEXT AS $$
				DECLARE
					var1 TEXT;
				BEGIN
					RAISE EXCEPTION '% %% bar %', 'foo', 1+1;
					var1 := input;
					RETURN var1;
				END;
				$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_raise1('123');",
					Expected: []sql.Row{{"123"}},
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "MyMessage",
						},
						{
							Severity: "NOTICE",
							Message:  "'MyNoticeMessage'",
						},
						{
							Severity: "DEBUG",
							Message:  "'DebugMessage'",
						},
					},
				},
				{
					Query:       "SELECT interpreted_raise2('123');",
					ExpectedErr: "foo % bar 2",
				},
			},
		},
		{
			Name: "SELECT INTO",
			SetUpScript: []string{`CREATE FUNCTION interpreted_select_into(input INT4) RETURNS TEXT AS $$
DECLARE
	ret TEXT;
	count INT4;
BEGIN
	DROP TABLE IF EXISTS temp_table;
	CREATE TABLE temp_table (pk SERIAL PRIMARY KEY, v1 TEXT NOT NULL);
	INSERT INTO temp_table (v1) VALUES ('abc'), ('def'), ('ghi');
	SELECT COUNT(*) INTO count FROM temp_table;
	IF input > 0 AND input <= count THEN
		SELECT v1 INTO ret FROM temp_table WHERE pk = input;
	ELSE
		ret := 'out of bounds';
	END IF;
	RETURN ret;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_select_into(1);",
					Expected: []sql.Row{
						{"abc"},
					},
				},
				{
					Query: "SELECT interpreted_select_into(2);",
					Expected: []sql.Row{
						{"def"},
					},
				},
				{
					Query: "SELECT interpreted_select_into(3);",
					Expected: []sql.Row{
						{"ghi"},
					},
				},
				{
					Query: "SELECT interpreted_select_into(4);",
					Expected: []sql.Row{
						{"out of bounds"},
					},
				},
			},
		},
		{
			Name: "WHILE",
			SetUpScript: []string{
				`CREATE FUNCTION interpreted_while(input INT4) RETURNS INT AS $$
DECLARE
	counter INT4;
BEGIN
	WHILE counter + input < 100 LOOP
		-- Include more than one statement in the loop so it's not too simple 
		counter = counter + 1;
		counter = counter - 1;
		counter = counter + 1;
	END LOOP;
	RETURN counter;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_while(42);",
					Expected: []sql.Row{{58}},
				},
			},
		},
		{
			Name: "WHILE Label",
			SetUpScript: []string{
				`CREATE FUNCTION interpreted_while_label(input INT4) RETURNS INT AS $$
DECLARE
	counter INT4;
BEGIN
	<<while_label>>
	WHILE input < 1000 LOOP
		input := input + 1;
		counter := counter + 1;
		IF counter >= 10 THEN
			EXIT while_label;
		END IF;
	END LOOP;
	RETURN input;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_while_label(42);",
					Expected: []sql.Row{{52}},
				},
			},
		},
		{
			Name: "NULL",
			SetUpScript: []string{
				`CREATE FUNCTION interpreted_null(input INT) RETURNS TEXT AS $$
BEGIN
	IF input = 42 THEN
		NULL;
		NULL;
	ELSE
		RETURN 'No'; 
	END IF;
	NULL;
	RETURN 'Yes'; 
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_null(42);",
					Expected: []sql.Row{{"Yes"}},
				},
				{
					Query:    "SELECT interpreted_null(43);",
					Expected: []sql.Row{{"No"}},
				},
			},
		},
		{
			// Tests that variable names are correctly substituted with references
			// to the variables when the function is parsed.
			Name: "Variable reference substitution",
			SetUpScript: []string{`
CREATE FUNCTION test1(input TEXT) RETURNS TEXT AS $$
DECLARE
	var1 TEXT;
BEGIN
	var1 := 'input' || input;
	IF var1 = 'input' || input THEN
		RETURN var1 || 'var1';
	ELSE
		RETURN '!!!';
	END IF;
END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT test1('Hello');",
					Expected: []sql.Row{{"inputHellovar1"}},
				},
			},
		},
		{
			Name: "Overloading",
			SetUpScript: []string{`CREATE FUNCTION interpreted_overload(input TEXT) RETURNS TEXT AS $$
DECLARE
	var1 TEXT;
BEGIN
	IF length(input) > 3 THEN
		var1 := input || '_long';
	ELSE
		var1 := input;
	END IF;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION interpreted_overload(input INT4) RETURNS INT4 AS $$
DECLARE
	var1 INT4;
BEGIN
	IF input > 3 THEN
		var1 := -input;
	ELSE
		var1 := input;
	END IF;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_overload('abc');",
					Expected: []sql.Row{
						{"abc"},
					},
				},
				{
					Query: "SELECT interpreted_overload('abcd');",
					Expected: []sql.Row{
						{"abcd_long"},
					},
				},
				{
					Query: "SELECT interpreted_overload(3);",
					Expected: []sql.Row{
						{3},
					},
				},
				{
					Query: "SELECT interpreted_overload(4);",
					Expected: []sql.Row{
						{-4},
					},
				},
			},
		},
		{
			Name: "Branching",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION interpreted_as_of(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN input || '_extra';
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT interpreted_as_of('abcd');",
					Expected: []sql.Row{{"abcd_extra"}},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT dolt_checkout('-b', 'other')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'other'"}`}},
				},
				{
					Query: `CREATE OR REPLACE FUNCTION interpreted_as_of(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN input;
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'updated func')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT interpreted_as_of('abc');",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT dolt_checkout('main')",
					Expected: []sql.Row{{`{0,"Switched to branch 'main'"}`}},
				},
				{
					Query:    "SELECT interpreted_as_of('abcd');",
					Expected: []sql.Row{{"abcd_extra"}},
				},
			},
		},
		{
			Name: "Merging No Conflict",
			SetUpScript: []string{
				`CREATE TABLE test(pk INT4);`,
				`INSERT INTO test VALUES (77);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION interpreted_merging(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN input || '_extra';
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT interpreted_merging('abcd');",
					Expected: []sql.Row{{"abcd_extra"}},
				},
				{
					Query:       "SELECT interpreted_merging(55);",
					ExpectedErr: "does not exist",
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT dolt_checkout('-b', 'other')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'other'"}`}},
				},
				{
					Query: `CREATE FUNCTION interpreted_merging(input INT4) RETURNS INT4 AS $$
BEGIN
	RETURN input + 11;
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'another func')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT interpreted_merging(55);",
					Expected: []sql.Row{{66}},
				},
				{
					Query:    "SELECT dolt_checkout('main')",
					Expected: []sql.Row{{`{0,"Switched to branch 'main'"}`}},
				},
				{
					Query:    "INSERT INTO test VALUES (80);",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'updated table')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT interpreted_merging('abcde');",
					Expected: []sql.Row{{"abcde_extra"}},
				},
				{
					Query:       "SELECT interpreted_merging(67);",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "SELECT * FROM test;",
					Expected: []sql.Row{{77}, {80}},
				},
				{
					Query:    "SELECT length(dolt_merge('other')::text) = 57;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT interpreted_merging('abcdef');",
					Expected: []sql.Row{{"abcdef_extra"}},
				},
				{
					Query:    "SELECT interpreted_merging(58);",
					Expected: []sql.Row{{69}},
				},
				{
					Query:    "SELECT * FROM test;",
					Expected: []sql.Row{{77}, {80}},
				},
			},
		},
		{
			Name: "INSERT values from function",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TEXT);",
				`CREATE FUNCTION insertion_text() RETURNS TEXT AS $$
DECLARE
    var1 TEXT;
BEGIN
    var1 := 'example';
    RETURN var1;
END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (insertion_text()), (insertion_text());",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{"example"},
						{"example"},
					},
				},
			},
		},
		{
			Name: "Create function on different branch",
			Skip: true, // several issues prevent this from working yet
			SetUpScript: []string{
				`CREATE FUNCTION f1(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN input || '_extra';
END;
$$ LANGUAGE plpgsql;`,
				`call dolt_branch('b1');`,
				`CREATE FUNCTION "postgres/b1".public.f1(input INT4) RETURNS INT4 AS $$
BEGIN
	RETURN input + 11;
END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT f1('abcd');`,
					Expected: []sql.Row{{"abcd_extra"}},
				},
				{
					Query:    `SELECT "postgres/b1".public.f1(55);`,
					Expected: []sql.Row{{66}},
				},
				{
					Query: `call dolt_checkout('b1');`,
				},
				{
					Query:    `SELECT f1(55);`,
					Expected: []sql.Row{{66}},
				},
			},
		},
		{
			Name: "Nested IF statements with exceptions",
			SetUpScript: []string{
				`CREATE TABLE public.table_name (start_date DATE NOT NULL, end_date DATE);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION public.fn_name() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.start_date IS NOT NULL
       AND NEW.end_date IS NULL
    THEN
        NEW.end_date := NEW.start_date + INTERVAL '31 day';
    END IF;
    IF NEW.start_date IS NOT NULL
       AND NEW.end_date IS NOT NULL
    THEN
        IF NEW.end_date < NEW.start_date THEN
            RAISE EXCEPTION 'end_date (%) start_date (%)',
                NEW.end_date, NEW.start_date;
        END IF;
        IF NEW.end_date > (NEW.start_date + INTERVAL '31 day') THEN
            RAISE EXCEPTION 'Too far (start_date=%, end_date=%)',
                NEW.start_date, NEW.end_date;
        END IF;
    END IF;
    RETURN NEW;
END;
$$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TRIGGER trig_name BEFORE INSERT OR UPDATE ON public.table_name FOR EACH ROW EXECUTE FUNCTION public.fn_name();`,
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO public.table_name VALUES ('2025-01-02', '2025-02-02');",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO public.table_name VALUES ('2025-04-05', NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO public.table_name VALUES ('2025-09-10', '2025-07-08');",
					ExpectedErr: "end_date (2025-07-08) start_date (2025-09-10)",
				},
				{
					Query:       "INSERT INTO public.table_name VALUES ('2025-11-11', '2025-12-31');",
					ExpectedErr: "Too far (start_date=2025-11-11, end_date=2025-12-31)",
				},
			},
		},
		{
			Name: "Table as type for functions",
			SetUpScript: []string{
				// TODO: test case sensitivity of parameter names
				`CREATE TABLE test (id INT4 PRIMARY KEY, name TEXT NOT NULL, qty INT4 NOT NULL, price REAL NOT NULL);`,
				`INSERT INTO test VALUES (1, 'apple', 3, 2.5), (2, 'banana', 5, 1.2);`,
				`CREATE FUNCTION total(t test) RETURNS REAL AS $$ BEGIN RETURN t.qty * t.price; END; $$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION priceHike(t test, pricehike REAL) RETURNS test AS $$ BEGIN RETURN (t.id, t.name, t.qty, t.price + pricehike)::test; END; $$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION singleReturn() RETURNS test AS $$ DECLARE result test; BEGIN SELECT * INTO result FROM test WHERE id = 1; RETURN result; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT total(t) FROM test AS t;`,
					Expected: []sql.Row{
						{7.5},
						{6.0},
					},
				},
				{
					Query: `SELECT priceHike(t, 10.0) FROM test AS t;`,
					Expected: []sql.Row{
						{"(1,apple,3,12.5)"},
						{"(2,banana,5,11.2)"},
					},
				},
				{
					Query: `SELECT priceHike(ROW(3, 'orange', 1, 1.8)::test, 100.0);`,
					Expected: []sql.Row{
						{"(3,orange,1,101.8)"},
					},
				},
				{
					Query: `SELECT singleReturn();`,
					Skip:  true, // TODO: better PL/pgSQL internal support for non-trigger composite types
					Expected: []sql.Row{
						{"(1,apple,3,2.5)"},
					},
				},
			},
		},
		{
			Name: "Table as type for columns",
			SetUpScript: []string{
				`CREATE TABLE t1 (v1 INT4 PRIMARY KEY, v2 TEXT NOT NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE t2 (v1 INT4 PRIMARY KEY, v2 t1 NOT NULL);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO t2 VALUES (1, ROW(0, 'hello')::t1), (2, ROW(10, 'world')::t1);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM t2 ORDER BY v1;`,
					Expected: []sql.Row{
						{1, "(0,hello)"},
						{2, "(10,world)"},
					},
				},
			},
		},
		{
			Name: "AlexTransit_venderctl import dump",
			SetUpScript: []string{
				`CREATE TYPE public.tax_job_state AS ENUM (
    'sched',
    'busy',
    'final',
    'help'
);`,
				`CREATE TABLE public.catalog (
    vmid integer NOT NULL,
    code text NOT NULL,
    name text NOT NULL
);`,
				`CREATE SEQUENCE public.tax_job_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;`,
				`CREATE TABLE public.tax_job (
    id bigint NOT NULL,
    state public.tax_job_state NOT NULL,
    created timestamp with time zone NOT NULL,
    modified timestamp with time zone NOT NULL,
    scheduled timestamp with time zone,
    worker text,
    processor text,
    ext_id text,
    data jsonb,
    gross integer,
    notes text[],
    ops jsonb
);`,
				`CREATE TABLE public.trans (
    vmid integer NOT NULL,
    vmtime timestamp with time zone,
    received timestamp with time zone NOT NULL,
    menu_code text NOT NULL,
    options integer[],
    price integer NOT NULL,
    method integer NOT NULL,
    tax_job_id bigint,
    executer bigint,
    exeputer_type integer,
    executer_str text
);`,
				`ALTER TABLE ONLY public.tax_job ALTER COLUMN id SET DEFAULT nextval('public.tax_job_id_seq'::regclass);`,
				`INSERT INTO public.trans VALUES (1, '2023-04-05 06:07:08', '2023-05-06 07:08:09', 'test', ARRAY[5,7], 44, 1, NULL, 1, 1, '');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION public.tax_job_trans(t public.trans) RETURNS public.tax_job
    LANGUAGE plpgsql
    AS '
    # print_strict_params ON
DECLARE
    tjd jsonb;
    ops jsonb;
    tj tax_job;
    name text;
BEGIN
    -- lock trans row
    PERFORM
        1
    FROM
        trans
    WHERE (vmid, vmtime) = (t.vmid,
        t.vmtime)
LIMIT 1
FOR UPDATE;
    -- if trans already has tax_job assigned, just return it
    IF t.tax_job_id IS NOT NULL THEN
        SELECT
            * INTO STRICT tj
        FROM
            tax_job
        WHERE
            id = t.tax_job_id;
        RETURN tj;
    END IF;
    -- op code to human friendly name via catalog
    SELECT
        catalog.name INTO name
    FROM
        catalog
    WHERE (vmid, code) = (t.vmid,
        t.menu_code);
    IF NOT found THEN
        name := ''#'' || t.menu_code;
    END IF;
    ops := jsonb_build_array (jsonb_build_object(''vmid'', t.vmid, ''time'', t.vmtime, ''name'', name, ''code'', t.menu_code, ''amount'', 1, ''price'', t.price, ''method'', t.method));
    INSERT INTO tax_job (state, created, modified, scheduled, processor, ops, gross)
        VALUES (''sched'', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ''ru2019'', ops, t.price)
    RETURNING
        * INTO STRICT tj;
    UPDATE
        trans
    SET
        tax_job_id = tj.id
    WHERE (vmid, vmtime) = (t.vmid,
        t.vmtime);
    RETURN tj;
END;
';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT public.tax_job_trans(trans.*) FROM public.trans;`,
					Skip:     true, // TODO: implement table.* syntax
					Expected: []sql.Row{{`(1,sched,"2026-01-23 14:06:32.794817+00","2026-01-23 14:06:32.794817+00","2026-01-23 14:06:32.794817+00",,ru2019,,,44,,"[{""code"": ""test"", ""name"": ""#test"", ""time"": ""2023-04-05T06:07:08+00:00"", ""vmid"": 1, ""price"": 44, ""amount"": 1, ""method"": 1}]")`}},
				},
			},
		},
		{
			Name: "resolve type with empty search path",
			SetUpScript: []string{
				"set search_path to ''",
				`CREATE TABLE public.ambienttempdetail (tempdetailid integer NOT NULL, panelprojectid integer, threshold_value numeric(10,2), readingintervalinmin integer);`,
				`insert into public.ambienttempdetail values (1, 101, 25.5, 15);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION public.ambienttempdetail_insertupdate(p_panel_project_id integer, p_threshold_value numeric, p_reading_interval_in_min integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_rtn_value INTEGER;
BEGIN
    IF NOT EXISTS (SELECT * FROM AmbientTempDetail WHERE PanelProjectId = p_panel_project_id) THEN
        INSERT INTO AmbientTempDetail (PanelProjectId, Threshold_Value, ReadingIntervalInMin)
        VALUES (p_panel_project_id, p_threshold_value, p_reading_interval_in_min)
        RETURNING TempDetailId INTO v_rtn_value;
    ELSE
        UPDATE AmbientTempDetail
        SET PanelProjectId = p_panel_project_id,
            Threshold_Value = p_threshold_value,
            ReadingIntervalInMin = p_reading_interval_in_min
        WHERE PanelProjectId = p_panel_project_id;
        v_rtn_value := p_panel_project_id;
    END IF;
    
    RETURN v_rtn_value;
END;
$$;`,
					Expected: []sql.Row{},
				},
				{
					Query: "set search_path to 'public'",
				},
				{
					Skip:     true,
					Query:    "SELECT public.ambienttempdetail_insertupdate(101, 25.5, 15);",
					Expected: []sql.Row{{101}},
				},
			},
		},
	})
}
