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

func TestCreateFunction(t *testing.T) {
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
			Name: "RAISE",
			SetUpScript: []string{
				`CREATE FUNCTION interpreted_raise(input TEXT) RETURNS TEXT AS $$
				DECLARE
					var1 TEXT;
				BEGIN
					RAISE WARNING 'MyMessage';
					RAISE NOTICE USING MESSAGE = 'MyNoticeMessage'; 
					RAISE DEBUG 'DebugTest1' USING MESSAGE = 'DebugMessage';
					RAISE EXCEPTION '% %% bar %', 'foo', 1+1;
					var1 := input;
					RETURN var1;
				END;
				$$ LANGUAGE plpgsql;
				`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_raise('123');",
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
						{
							Severity: "EXCEPTION",
							Message:  "foo % bar 2",
						},
					},
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
	})
}
