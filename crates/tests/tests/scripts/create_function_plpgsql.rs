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

use harness::oid::*;
use harness::pgx::Time;
use harness::plan::PlanFact;
use harness::script::Cell::{Any, Null, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_create_function_language_plpgsql() {
    run_scripts(&[
        ScriptTest {
            name: "ALIAS",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_alias(input TEXT)
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
				"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_alias('123');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_alias", TEXT)],
                        rows: &[
                            &[T("123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Assignment",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_assignment(input TEXT) RETURNS TEXT AS $$
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_assignment('Hello');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_assignment", TEXT)],
                        rows: &[
                            &[T("Initial: Hello - Greeting")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_assignment('Bye');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_assignment", TEXT)],
                        rows: &[
                            &[T("Initial: Bye - Farewell")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_assignment('abc');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_assignment", TEXT)],
                        rows: &[
                            &[T("Initial: abc - Else")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_assignment('something');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_assignment", TEXT)],
                        rows: &[
                            &[T("Initial: something - Over 5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "conditions that evaluate to NULL",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_null_if(input TEXT) RETURNS TEXT AS $$
BEGIN
	IF input = 'Hello' THEN
		RETURN 'Greeting';
	ELSIF input = 'Bye' THEN
		RETURN 'Farewell';
	ELSE
		RETURN 'Else';
	END IF;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_null_while(input TEXT) RETURNS INT AS $$
DECLARE
	count1 INT := 0;
BEGIN
	WHILE input = 'Hello' LOOP
		count1 := count1 + 1;
		EXIT WHEN count1 > 2;
	END LOOP;
	RETURN count1;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_null_exit(input TEXT) RETURNS INT AS $$
DECLARE
	count1 INT := 0;
BEGIN
	LOOP
		count1 := count1 + 1;
		EXIT WHEN input = 'Hello';
		EXIT WHEN count1 > 2;
	END LOOP;
	RETURN count1;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_null_case(x INT) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE x
		WHEN 1 THEN
			msg := 'one';
		ELSE
			msg := 'other';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_null_searched_case(x INT) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE
		WHEN x = 1 THEN
			msg := 'one';
		ELSE
			msg := 'other';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_if(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_if", TEXT)],
                        rows: &[
                            &[T("Else")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_if('Bye');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_if", TEXT)],
                        rows: &[
                            &[T("Farewell")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_while(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_while", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_while('Hello');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_while", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_exit(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_exit", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_exit('Hello');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_exit", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_case(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_case", TEXT)],
                        rows: &[
                            &[T("other")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_case(1);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_case", TEXT)],
                        rows: &[
                            &[T("one")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_searched_case(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_searched_case", TEXT)],
                        rows: &[
                            &[T("other")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null_searched_case(1);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null_searched_case", TEXT)],
                        rows: &[
                            &[T("one")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CASE, with ELSE",
            set_up_script: &[
                r#"
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(1);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("one or two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(2);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("one or two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(0);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("other value than one or two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CASE over a non-integer expression",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_case_text(x TEXT) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE x
		WHEN 'Hello', 'Hi' THEN
			msg := 'greeting';
		WHEN 'Bye' THEN
			msg := 'farewell';
		ELSE
			msg := 'other';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_case_bool(x BOOLEAN) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE x
		WHEN true THEN
			msg := 'yes';
		ELSE
			msg := 'no';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_case_numeric(x NUMERIC) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE x
		WHEN 1.5 THEN
			msg := 'one point five';
		ELSE
			msg := 'other';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_case_loop() RETURNS TEXT AS $$
DECLARE
	i INT := 0;
	msg TEXT;
	result TEXT := '';
BEGIN
	WHILE i < 3 LOOP
		i := i + 1;
		CASE i::TEXT
			WHEN '2' THEN
				msg := 'two';
			ELSE
				msg := 'n';
		END CASE;
		result := result || msg;
	END LOOP;
	RETURN result;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_text('Hello');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_text", TEXT)],
                        rows: &[
                            &[T("greeting")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_text('Bye');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_text", TEXT)],
                        rows: &[
                            &[T("farewell")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_text('zzz');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_text", TEXT)],
                        rows: &[
                            &[T("other")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_bool(true);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_bool", TEXT)],
                        rows: &[
                            &[T("yes")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_bool(false);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_bool", TEXT)],
                        rows: &[
                            &[T("no")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_numeric(1.5);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_numeric", TEXT)],
                        rows: &[
                            &[T("one point five")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_numeric(2.5);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_numeric", TEXT)],
                        rows: &[
                            &[T("other")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_loop();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_loop", TEXT)],
                        rows: &[
                            &[T("ntwon")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "NULL conditions are not met",
            set_up_script: &[
                "CREATE TABLE case_selector (v TEXT);",
                "INSERT INTO case_selector VALUES ('match');",
                r#"CREATE FUNCTION interpreted_case_null(x TEXT) RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE x
		WHEN 'match' THEN
			msg := 'matched';
		ELSE
			msg := 'fell through';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_case_empty_selector() RETURNS TEXT AS $$
DECLARE
	msg TEXT;
BEGIN
	CASE (SELECT v FROM case_selector LIMIT 1)
		WHEN 'match' THEN
			msg := 'matched';
		ELSE
			msg := 'fell through';
	END CASE;
	RETURN msg;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_if_null(x BOOLEAN) RETURNS TEXT AS $$
BEGIN
	IF x THEN
		RETURN 'true';
	ELSE
		RETURN 'not true';
	END IF;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_while_null(x BOOLEAN) RETURNS TEXT AS $$
BEGIN
	WHILE x LOOP
		RETURN 'looped';
	END LOOP;
	RETURN 'never looped';
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_null('match');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_null", TEXT)],
                        rows: &[
                            &[T("matched")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_null(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_null", TEXT)],
                        rows: &[
                            &[T("fell through")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_empty_selector();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_empty_selector", TEXT)],
                        rows: &[
                            &[T("matched")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM case_selector;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case_empty_selector();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case_empty_selector", TEXT)],
                        rows: &[
                            &[T("fell through")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_if_null(true);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_if_null", TEXT)],
                        rows: &[
                            &[T("true")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_if_null(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_if_null", TEXT)],
                        rows: &[
                            &[T("not true")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_while_null(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_while_null", TEXT)],
                        rows: &[
                            &[T("never looped")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CASE, without ELSE",
            set_up_script: &[
                r#"
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(1);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("one or two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(2);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("one or two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(0);",
                    expected: Expected::Error(Diagnostic { code: "20000", message: "case not found", hint: "CASE statement is missing ELSE part.", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Searched CASE, with ELSE",
            set_up_script: &[
                r#"
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(0);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is between zero and ten")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(1);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is between zero and ten")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(10);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is between zero and ten")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(11);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is between eleven and twenty")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(21);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is out of bounds")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Searched CASE, without ELSE",
            set_up_script: &[
                r#"
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(0);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is between zero and ten")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(1);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is between zero and ten")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(10);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is between zero and ten")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(11);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_case", TEXT)],
                        rows: &[
                            &[T("value is between eleven and twenty")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_case(21);",
                    expected: Expected::Error(Diagnostic { code: "20000", message: "case not found", hint: "CASE statement is missing ELSE part.", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_continue() RETURNS INT4 AS $$
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_continue();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_continue", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CONTINUE Label",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_continue_label() RETURNS INT4 AS $$
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_continue_label();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_continue_label", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: var1 was never initialized, so in Postgres the loop never ends.
        ScriptTest {
            name: "EXIT",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_exit() RETURNS INT4 AS $$
DECLARE
	var1 INT4 := 0;
BEGIN
	LOOP
		var1 := var1 + 1;
		IF var1 >= 8 THEN
			EXIT;
		END IF;
	END LOOP;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_exit();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_exit", INT4)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: var1 was never initialized, so in Postgres the loop never ends.
        ScriptTest {
            name: "EXIT WHEN",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_exit_when() RETURNS INT4 AS $$
DECLARE
	var1 INT4 := 0;
BEGIN
	LOOP
		var1 := var1 + 1;
		EXIT WHEN var1 >= 9;
	END LOOP;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_exit_when();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_exit_when", INT4)],
                        rows: &[
                            &[T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: var1 was never initialized, so in Postgres NULL + 1 stays NULL and the loop never ends; Doltgres treated it as 0.
        ScriptTest {
            name: "LOOP",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_loop() RETURNS INT4 AS $$
DECLARE
	var1 INT4 := 0;
BEGIN
	LOOP
		var1 := var1 + 1;
		IF var1 >= 10 THEN
			RETURN var1;
		END IF;
	END LOOP;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_uninitialized() RETURNS INT4 AS $$
DECLARE
	var1 INT4;
BEGIN
	var1 := var1 + 1;
	RETURN var1;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_loop();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_loop", INT4)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_uninitialized();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_uninitialized", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: var1 was never initialized, so in Postgres the loop never ends.
        ScriptTest {
            name: "LOOP Label",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_loop_label() RETURNS INT4 AS $$
DECLARE
	var1 INT4 := 0;
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_loop_label();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_loop_label", INT4)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "PERFORM",
            set_up_script: &[
                "CREATE SEQUENCE test_sequence;",
                r#"CREATE FUNCTION interpreted_perform() RETURNS VOID AS $$
BEGIN
	PERFORM nextval('test_sequence');
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT nextval('test_sequence');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_perform();",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_perform", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT nextval('test_sequence');",
                    expected: Expected::Rows {
                        columns: &[Column("nextval", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS SETOF",
            set_up_script: &[
                r#"CREATE TYPE user_summary AS (
					user_id   integer,
					username  text,
					is_active boolean);"#,
                r#"CREATE OR REPLACE FUNCTION func2() RETURNS SETOF user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 1, 'username', true;
						RETURN QUERY SELECT 2, 'another', false;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", USER_DEFINED)],
                        rows: &[
                            &[T("(1,username,t)")],
                            &[T("(2,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(), func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", USER_DEFINED), Column("func2", USER_DEFINED)],
                        rows: &[
                            &[T("(1,username,t)"), T("(1,username,t)")],
                            &[T("(2,another,f)"), T("(2,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS SETOF with no results",
            set_up_script: &[
                "CREATE TABLE user_summary (user_id integer, username text, is_active boolean);",
                r#"CREATE OR REPLACE FUNCTION func2() RETURNS SETOF user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT * from user_summary;
						RETURN QUERY SELECT * from user_summary;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", USER_DEFINED)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(), func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", USER_DEFINED), Column("func2", USER_DEFINED)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS SETOF with type from other schema",
            set_up_script: &[
                "CREATE SCHEMA sch1;",
                r#"CREATE TYPE sch1.user_summary AS (
					user_id   integer,
					username  text,
					is_active boolean);"#,
                r#"CREATE OR REPLACE FUNCTION func2() RETURNS SETOF sch1.user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 1, 'username', true;
						RETURN QUERY SELECT 2, 'another', false;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", USER_DEFINED)],
                        rows: &[
                            &[T("(1,username,t)")],
                            &[T("(2,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(), func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", USER_DEFINED), Column("func2", USER_DEFINED)],
                        rows: &[
                            &[T("(1,username,t)"), T("(1,username,t)")],
                            &[T("(2,another,f)"), T("(2,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS SETOF with param",
            set_up_script: &[
                r#"CREATE TYPE user_summary AS (
					user_id   integer,
					username  text,
					is_active boolean);"#,
                r#"CREATE OR REPLACE FUNCTION func3(user_id integer) RETURNS SETOF user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT user_id, 'username', true;
						RETURN QUERY SELECT user_id, 'another', false;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func3(111);",
                    expected: Expected::Rows {
                        columns: &[Column("func3", USER_DEFINED)],
                        rows: &[
                            &[T("(111,username,t)")],
                            &[T("(111,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func3(111), func3(222);",
                    expected: Expected::Rows {
                        columns: &[Column("func3", USER_DEFINED), Column("func3", USER_DEFINED)],
                        rows: &[
                            &[T("(111,username,t)"), T("(222,username,t)")],
                            &[T("(111,another,f)"), T("(222,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS TABLE",
            set_up_script: &[
                r#"CREATE FUNCTION func2() RETURNS TABLE(user_id integer, username  text, is_active boolean)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 1, 'username', true;
						RETURN QUERY SELECT 2, 'another', false;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", RECORD)],
                        rows: &[
                            &[T("(1,username,t)")],
                            &[T("(2,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(), func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", RECORD), Column("func2", RECORD)],
                        rows: &[
                            &[T("(1,username,t)"), T("(1,username,t)")],
                            &[T("(2,another,f)"), T("(2,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS TABLE with single field",
            set_up_script: &[
                r#"CREATE FUNCTION func2() RETURNS TABLE(username text)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 'username1';
						RETURN QUERY SELECT 'username2';
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", TEXT)],
                        rows: &[
                            &[T("username1")],
                            &[T("username2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(), func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", TEXT), Column("func2", TEXT)],
                        rows: &[
                            &[T("username1"), T("username1")],
                            &[T("username2"), T("username2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS TABLE with types from other schema",
            set_up_script: &[
                "CREATE SCHEMA sch1;",
                r#"CREATE TYPE sch1.mytype AS (
					user_id   integer,
					username  text);"#,
                r#"CREATE FUNCTION func2() RETURNS TABLE(foo sch1.mytype)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 1, 'username1';
						RETURN QUERY SELECT 2, 'username2';
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", USER_DEFINED)],
                        rows: &[
                            &[T("(1,username1)")],
                            &[T("(2,username2)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(), func2();",
                    expected: Expected::Rows {
                        columns: &[Column("func2", USER_DEFINED), Column("func2", USER_DEFINED)],
                        rows: &[
                            &[T("(1,username1)"), T("(1,username1)")],
                            &[T("(2,username2)"), T("(2,username2)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: Postgres rejects an input parameter with the same name as an output column, so the input is renamed.
        ScriptTest {
            name: "RETURNS TABLE with param",
            set_up_script: &[
                r#"CREATE OR REPLACE FUNCTION func3(p_user_id integer) RETURNS TABLE(user_id integer, username  text, is_active boolean)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT p_user_id, 'username', true;
						RETURN QUERY SELECT p_user_id, 'another', false;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func3(111);",
                    expected: Expected::Rows {
                        columns: &[Column("func3", RECORD)],
                        rows: &[
                            &[T("(111,username,t)")],
                            &[T("(111,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func3(111), func3(222);",
                    expected: Expected::Rows {
                        columns: &[Column("func3", RECORD), Column("func3", RECORD)],
                        rows: &[
                            &[T("(111,username,t)"), T("(222,username,t)")],
                            &[T("(111,another,f)"), T("(222,another,f)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS TABLE with join query",
            set_up_script: &[
                r#"CREATE TABLE customers (
					id INT PRIMARY KEY,
					name TEXT
				);"#,
                r#"CREATE TABLE orders (
					id SERIAL PRIMARY KEY,
					customer_id INT,
					amount INT
				);"#,
                "INSERT INTO customers VALUES (1, 'John'), (2, 'Jane');",
                "INSERT INTO orders VALUES (1, 1, 100), (2, 2, 10);",
                r#"CREATE OR REPLACE FUNCTION func2(n INT) RETURNS TABLE (c_id INT, c_name TEXT, c_total_spent INT) 
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
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func2(1);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "structure of query does not match function result type", detail: "Returned type bigint does not match expected type integer in column 3.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(11);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "structure of query does not match function result type", detail: "Returned type bigint does not match expected type integer in column 3.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func2(111);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "structure of query does not match function result type", detail: "Returned type bigint does not match expected type integer in column 3.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS TABLE with conflicting variable names",
            set_up_script: &[
                r#"CREATE TABLE customers2 (
					id INT PRIMARY KEY,
					name TEXT
				);"#,
                "INSERT INTO customers2 VALUES (1, 'John'), (2, 'Jane');",
                r#"CREATE OR REPLACE FUNCTION func_conflict(n INT) RETURNS TABLE (id INT, name TEXT) 
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY
						SELECT c.id, c.name
						FROM customers2 c
						WHERE c.id = n;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func_conflict(1);",
                    expected: Expected::Rows {
                        columns: &[Column("func_conflict", RECORD)],
                        rows: &[
                            &[T("(1,John)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func_conflict(2);",
                    expected: Expected::Rows {
                        columns: &[Column("func_conflict", RECORD)],
                        rows: &[
                            &[T("(2,Jane)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS SETOF with composite param",
            set_up_script: &[
                r#"CREATE TYPE user_summary AS (
					user_id   integer,
					username  text,
					is_active boolean);"#,
                r#"CREATE OR REPLACE FUNCTION func3(u user_summary) RETURNS SETOF user_summary
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT u.user_id, u.username, u.is_active;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT func3((222,'passedin',false)::user_summary);",
                    expected: Expected::Rows {
                        columns: &[Column("func3", USER_DEFINED)],
                        rows: &[
                            &[T("(222,passedin,f)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RAISE",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_raise1(input TEXT) RETURNS TEXT AS $$
				DECLARE
					var1 TEXT;
				BEGIN
					RAISE WARNING 'MyMessage';
					RAISE NOTICE USING MESSAGE = 'MyNoticeMessage';
					RAISE DEBUG 'DebugTest1' USING MESSAGE = 'DebugMessage';
					var1 := input;
					RETURN var1;
				END;
				$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_raise2(input TEXT) RETURNS TEXT AS $$
				DECLARE
					var1 TEXT;
				BEGIN
					RAISE EXCEPTION '% %% bar %', 'foo', 1+1;
					var1 := input;
					RETURN var1;
				END;
				$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_raise_errcode() RETURNS TEXT AS $$
				BEGIN
					RAISE EXCEPTION 'coded' USING ERRCODE = '22012';
				END;
				$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_raise_condition_name() RETURNS TEXT AS $$
				BEGIN
					RAISE EXCEPTION 'named' USING ERRCODE = 'division_by_zero';
				END;
				$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_raise1('123');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "RAISE option already specified: MESSAGE", ..E }),
                    notices: &[Diagnostic { severity: "WARNING", code: "01000", message: "MyMessage", ..E }, Diagnostic { code: "00000", message: "MyNoticeMessage", ..N }],
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_raise2('123');",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "foo % bar 2", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_raise_errcode();",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "coded", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_raise_condition_name();",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "named", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT INTO",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_select_into(input INT4) RETURNS TEXT AS $$
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_select_into(1);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_select_into", TEXT)],
                        rows: &[
                            &[T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    notices: &[Diagnostic { code: "00000", message: r#"table "temp_table" does not exist, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_select_into(2);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_select_into", TEXT)],
                        rows: &[
                            &[T("def")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_select_into(3);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_select_into", TEXT)],
                        rows: &[
                            &[T("ghi")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_select_into(4);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_select_into", TEXT)],
                        rows: &[
                            &[T("out of bounds")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "WHILE",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_while(input INT4) RETURNS INT AS $$
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_while(42);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_while", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "WHILE Label",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_while_label(input INT4) RETURNS INT AS $$
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_while_label(42);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_while_label", INT4)],
                        rows: &[
                            &[T("1000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "NULL",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_null(input INT) RETURNS TEXT AS $$
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_null(42);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null", TEXT)],
                        rows: &[
                            &[T("Yes")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_null(43);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_null", TEXT)],
                        rows: &[
                            &[T("No")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DO statement",
            set_up_script: &[
                "CREATE TABLE interpreted_do_values (v INT PRIMARY KEY)",
                r#"CREATE FUNCTION interpreted_do() RETURNS void AS $function$
BEGIN
	DO $block$ BEGIN
		INSERT INTO interpreted_do_values VALUES (42);
	END $block$;
END;
$function$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_do()",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_do", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM interpreted_do_values",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DO statement failure is atomic",
            set_up_script: &[
                "CREATE TABLE interpreted_do_atomic (v INT PRIMARY KEY)",
                r#"CREATE FUNCTION interpreted_do_error() RETURNS void AS $function$
BEGIN
	DO $block$ BEGIN
		INSERT INTO interpreted_do_atomic VALUES (1);
		RAISE EXCEPTION 'nested DO failed';
	END $block$;
END;
$function$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_do_error()",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "nested DO failed", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM interpreted_do_atomic",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Variable reference substitution",
            set_up_script: &[
                r#"
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT test1('Hello');",
                    expected: Expected::Rows {
                        columns: &[Column("test1", TEXT)],
                        rows: &[
                            &[T("inputHellovar1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Overloading",
            set_up_script: &[
                r#"CREATE FUNCTION interpreted_overload(input TEXT) RETURNS TEXT AS $$
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
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION interpreted_overload(input INT4) RETURNS INT4 AS $$
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
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_overload('abc');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_overload", TEXT)],
                        rows: &[
                            &[T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_overload('abcd');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_overload", TEXT)],
                        rows: &[
                            &[T("abcd_long")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_overload(3);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_overload", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_overload(4);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_overload", INT4)],
                        rows: &[
                            &[T("-4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Branching",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION interpreted_as_of(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN input || '_extra';
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_as_of('abcd');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_as_of", TEXT)],
                        rows: &[
                            &[T("abcd_extra")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('-b', 'other')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'other'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION interpreted_as_of(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN input;
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'updated func')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_as_of('abc');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_as_of", TEXT)],
                        rows: &[
                            &[T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('main')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_as_of('abcd');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_as_of", TEXT)],
                        rows: &[
                            &[T("abcd_extra")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Merging No Conflict",
            set_up_script: &[
                "CREATE TABLE test(pk INT4);",
                "INSERT INTO test VALUES (77);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION interpreted_merging(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN input || '_extra';
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_merging('abcd');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_merging", TEXT)],
                        rows: &[
                            &[T("abcd_extra")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_merging(55);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function interpreted_merging(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('-b', 'other')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'other'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION interpreted_merging(input INT4) RETURNS INT4 AS $$
BEGIN
	RETURN input + 11;
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'another func')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_merging(55);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_merging", INT4)],
                        rows: &[
                            &[T("66")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_checkout('main')",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_checkout", RECORD)],
                        rows: &[
                            &[T(r#"(0,"Switched to branch 'main'")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (80);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_add('.');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_add", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'updated table')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_merging('abcde');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_merging", TEXT)],
                        rows: &[
                            &[T("abcde_extra")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_merging(67);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function interpreted_merging(integer) does not exist", ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("80")],
                            &[T("77")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_merge('other')::text) = 57;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_merging('abcdef');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_merging", TEXT)],
                        rows: &[
                            &[T("abcdef_extra")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_merging(58);",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_merging", INT4)],
                        rows: &[
                            &[T("69")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("80")],
                            &[T("77")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "INSERT values from function",
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT);",
                r#"CREATE FUNCTION insertion_text() RETURNS TEXT AS $$
DECLARE
    var1 TEXT;
BEGIN
    var1 := 'example';
    RETURN var1;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (insertion_text()), (insertion_text());",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", TEXT)],
                        rows: &[
                            &[T("example")],
                            &[T("example")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Create function on different branch",
            set_up_script: &[
                r#"CREATE FUNCTION f1(input TEXT) RETURNS TEXT AS $$
BEGIN
	RETURN input || '_extra';
END;
$$ LANGUAGE plpgsql;"#,
                "call dolt_branch('b1');",
                r#"CREATE FUNCTION "postgres/b1".public.f1(input INT4) RETURNS INT4 AS $$
BEGIN
	RETURN input + 11;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            skip: Some(r#"setup fails on Postgres ("error running setup query: call dolt_branch('b1');: ERROR: procedure dolt_branch(unknown) does not exist (SQLSTATE 42883)") and on the Go server ("error running setup query: call dolt_branch('b1');: ERROR: Dolt stored procedure may only be invoked using SELECT (SQLSTATE XX000)")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT f1('abcd');",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT "postgres/b1".public.f1(55);"#,
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "call dolt_checkout('b1');",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT f1(55);",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Nested IF statements with exceptions",
            set_up_script: &[
                "CREATE TABLE public.table_name (start_date DATE NOT NULL, end_date DATE);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION public.fn_name() RETURNS trigger
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
$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER trig_name BEFORE INSERT OR UPDATE ON public.table_name FOR EACH ROW EXECUTE FUNCTION public.fn_name();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.table_name VALUES ('2025-01-02', '2025-02-02');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.table_name VALUES ('2025-04-05', NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.table_name VALUES ('2025-09-10', '2025-07-08');",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "end_date (2025-07-08) start_date (2025-09-10)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO public.table_name VALUES ('2025-11-11', '2025-12-31');",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "Too far (start_date=2025-11-11, end_date=2025-12-31)", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Table as type for functions",
            set_up_script: &[
                "CREATE TABLE test (id INT4 PRIMARY KEY, name TEXT NOT NULL, qty INT4 NOT NULL, price REAL NOT NULL);",
                "INSERT INTO test VALUES (1, 'apple', 3, 2.5), (2, 'banana', 5, 1.2);",
                "CREATE FUNCTION total(t test) RETURNS REAL AS $$ BEGIN RETURN t.qty * t.price; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION priceHike(t test, pricehike REAL) RETURNS test AS $$ BEGIN RETURN (t.id, t.name, t.qty, t.price + pricehike)::test; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION singleReturn() RETURNS test AS $$ DECLARE result test; BEGIN SELECT * INTO result FROM test WHERE id = 1; RETURN result; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT total(t) FROM test AS t;",
                    expected: Expected::Rows {
                        columns: &[Column("total", FLOAT4)],
                        rows: &[
                            &[T("7.5")],
                            &[T("6")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT priceHike(t, 10.0) FROM test AS t;",
                    expected: Expected::Rows {
                        columns: &[Column("pricehike", USER_DEFINED)],
                        rows: &[
                            &[T("(1,apple,3,12.5)")],
                            &[T("(2,banana,5,11.2)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT priceHike(ROW(3, 'orange', 1, 1.8)::test, 100.0);",
                    expected: Expected::Rows {
                        columns: &[Column("pricehike", USER_DEFINED)],
                        rows: &[
                            &[T("(3,orange,1,101.8)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT singleReturn();",
                    expected: Expected::Rows {
                        columns: &[Column("singlereturn", USER_DEFINED)],
                        rows: &[
                            &[T("(1,apple,3,2.5)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Table as type for columns",
            set_up_script: &[
                "CREATE TABLE t1 (v1 INT4 PRIMARY KEY, v2 TEXT NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE t2 (v1 INT4 PRIMARY KEY, v2 t1 NOT NULL);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t2 VALUES (1, ROW(0, 'hello')::t1), (2, ROW(10, 'world')::t1);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t2 ORDER BY v1;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("(0,hello)")],
                            &[T("2"), T("(10,world)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "AlexTransit_venderctl import dump",
            set_up_script: &[
                r#"CREATE TYPE public.tax_job_state AS ENUM (
    'sched',
    'busy',
    'final',
    'help'
);"#,
                r#"CREATE TABLE public.catalog (
    vmid integer NOT NULL,
    code text NOT NULL,
    name text NOT NULL
);"#,
                r#"CREATE SEQUENCE public.tax_job_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;"#,
                r#"CREATE TABLE public.tax_job (
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
);"#,
                r#"CREATE TABLE public.trans (
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
);"#,
                "ALTER TABLE ONLY public.tax_job ALTER COLUMN id SET DEFAULT nextval('public.tax_job_id_seq'::regclass);",
                "INSERT INTO public.trans VALUES (1, '2023-04-05 06:07:08', '2023-05-06 07:08:09', 'test', ARRAY[5,7], 44, 1, NULL, 1, 1, '');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION public.tax_job_trans(t public.trans) RETURNS public.tax_job
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
';"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.tax_job_trans(trans.*) FROM public.trans;",
                    expected: Expected::Rows {
                        columns: &[Column("tax_job_trans", USER_DEFINED)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "resolve type with empty search path",
            set_up_script: &[
                "set search_path to ''",
                "CREATE TABLE public.ambienttempdetail (tempdetailid integer NOT NULL, panelprojectid integer, threshold_value numeric(10,2), readingintervalinmin integer);",
                "insert into public.ambienttempdetail values (1, 101, 25.5, 15);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION public.ambienttempdetail_insertupdate(p_panel_project_id integer, p_threshold_value numeric, p_reading_interval_in_min integer) RETURNS integer
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
$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "set search_path to 'public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT public.ambienttempdetail_insertupdate(101, 25.5, 15);",
                    expected: Expected::Rows {
                        columns: &[Column("ambienttempdetail_insertupdate", INT4)],
                        rows: &[
                            &[T("101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "create function on non-existent table that does not exist yet with 'check_function_bodies'",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW check_function_bodies;",
                    expected: Expected::Rows {
                        columns: &[Column("check_function_bodies", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION public.film_in_stock(p_film_id integer, p_store_id integer, OUT p_film_count integer) RETURNS SETOF integer
    LANGUAGE sql
    AS $_$
     SELECT inventory_id
     FROM inventory
     WHERE film_id = $1
     AND store_id = $2
     AND inventory_in_stock(inventory_id);
$_$;"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "inventory" does not exist"#, position: 188, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET check_function_bodies = false;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION public.film_in_stock(p_film_id integer, p_store_id integer, OUT p_film_count integer) RETURNS SETOF integer
    LANGUAGE sql
    AS $_$
     SELECT inventory_id
     FROM inventory
     WHERE film_id = $1
     AND store_id = $2
     AND inventory_in_stock(inventory_id);
$_$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE TABLE public.inventory (
    inventory_id integer DEFAULT nextval('public.inventory_inventory_id_seq'::regclass) NOT NULL,
    film_id smallint NOT NULL,
    store_id smallint NOT NULL,
    last_update timestamp without time zone DEFAULT now() NOT NULL
);
"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "public.inventory_inventory_id_seq" does not exist"#, position: 74, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DECLARE variable with default value of literal value or parameter reference",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE OR REPLACE FUNCTION d() RETURNS TEXT[] AS $$ DECLARE chars TEXT[] := '{A,B,C,D,E,F,G,H}'; BEGIN RETURN chars; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d();",
                    expected: Expected::Rows {
                        columns: &[Column("d", TEXT_ARRAY)],
                        rows: &[
                            &[T("{A,B,C,D,E,F,G,H}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION
  mylt2 (x text, y text, e int) RETURNS boolean LANGUAGE plpgsql AS $$
declare
  xx text COLLATE "POSIX" := x;
  yy text := y;
  zz int := e;
begin
  return xx < yy;
end
$$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT mylt2('a', 'B', 1) as f;",
                    expected: Expected::Rows {
                        columns: &[Column("f", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DECLARE variable with default value of an expression",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION array_default() RETURNS TEXT[] AS $$ DECLARE permitted TEXT[] := ARRAY['retired_at', 'deleted_at']; BEGIN RETURN permitted; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_default();",
                    expected: Expected::Rows {
                        columns: &[Column("array_default", TEXT_ARRAY)],
                        rows: &[
                            &[T("{retired_at,deleted_at}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION quote_default() RETURNS TEXT AS $$ DECLARE x TEXT := 'it''s'; BEGIN RETURN x; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT quote_default();",
                    expected: Expected::Rows {
                        columns: &[Column("quote_default", TEXT)],
                        rows: &[
                            &[T("it's")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION call_default() RETURNS TEXT AS $$ DECLARE x TEXT := upper('abc') || length('abcd'); BEGIN RETURN x; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT call_default();",
                    expected: Expected::Rows {
                        columns: &[Column("call_default", TEXT)],
                        rows: &[
                            &[T("ABC4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION chained_default(p INT) RETURNS TEXT AS $$
DECLARE
	a INT := p * 2;
	b INT := a + 1;
	c TEXT := 'a=' || a || ' b=' || b;
	d INT := (SELECT count(*) FROM (VALUES (1), (2)) v);
BEGIN
	RETURN c || ' d=' || d;
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT chained_default(5);",
                    expected: Expected::Rows {
                        columns: &[Column("chained_default", TEXT)],
                        rows: &[
                            &[T("a=10 b=11 d=2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION cast_default() RETURNS TEXT AS $$
DECLARE
	a NUMERIC := 1.5::numeric + 1;
	b TEXT := NULL;
	c INT[] := ARRAY[1, 2, 3];
	d TIMESTAMP := '2020-01-01 00:00:00'::timestamp;
BEGIN
	RETURN a || '|' || coalesce(b, 'nil') || '|' || c[2] || '|' || d;
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cast_default();",
                    expected: Expected::Rows {
                        columns: &[Column("cast_default", TEXT)],
                        rows: &[
                            &[T("2.5|nil|2|2020-01-01 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION qualified_default() RETURNS TEXT AS $$
DECLARE
	k CONSTANT TEXT := upper('abc');
	n TEXT NOT NULL := repeat('n', 2);
BEGIN
	RETURN k || n;
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT qualified_default();",
                    expected: Expected::Rows {
                        columns: &[Column("qualified_default", TEXT)],
                        rows: &[
                            &[T("ABCnn")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "FOR I LOOP statement",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION code()
RETURNS VARCHAR AS $$
DECLARE
chars TEXT[] := '{A,B,C,D,E,F,G,H}';
  result TEXT := '';
  i INTEGER;
BEGIN
FOR i IN 1..3 LOOP
    result := result || chars[1+i];
END LOOP;
RETURN result;
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT code();",
                    expected: Expected::Rows {
                        columns: &[Column("code", VARCHAR)],
                        rows: &[
                            &[T("BCD")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "FOR S LOOP statement",
            set_up_script: &[
                r#"CREATE TABLE decks (
    id bigint NOT NULL,
    name text DEFAULT NULL::character varying,
    parent bigint
);"#,
                "INSERT INTO decks VALUES (1, 'name1', 2), (2, 'name2', 4), (5, 'name3', 1), (7, 'name4', 9);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION delete_deck_tree(p_id bigint) RETURNS void
            LANGUAGE plpgsql
            AS $$
        DECLARE
           r record;
        BEGIN
           DELETE FROM decks WHERE parent = p_id;
        
           FOR r IN SELECT id FROM decks WHERE parent = p_id LOOP
              PERFORM delete_deck_tree(r.id);
           END LOOP;
        
           DELETE FROM decks WHERE id = p_id;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from decks;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("name", TEXT), Column("parent", INT8)],
                        rows: &[
                            &[T("1"), T("name1"), T("2")],
                            &[T("2"), T("name2"), T("4")],
                            &[T("5"), T("name3"), T("1")],
                            &[T("7"), T("name4"), T("9")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT delete_deck_tree(1);",
                    expected: Expected::Rows {
                        columns: &[Column("delete_deck_tree", VOID)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from decks;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT8), Column("name", TEXT), Column("parent", INT8)],
                        rows: &[
                            &[T("2"), T("name2"), T("4")],
                            &[T("7"), T("name4"), T("9")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION fors_nested_one_var() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            r RECORD;
            result TEXT := '';
        BEGIN
            FOR r IN SELECT 1 AS n UNION ALL SELECT 2 LOOP FOR r IN SELECT 8 AS n UNION ALL SELECT 9 LOOP result := result || r.n; END LOOP; END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT fors_nested_one_var();",
                    expected: Expected::Rows {
                        columns: &[Column("fors_nested_one_var", TEXT)],
                        rows: &[
                            &[T("8989")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "FOREACH IN ARRAY statement",
            set_up_script: &[
                "CREATE TABLE tags (id int PRIMARY KEY, vals text[]);",
                "INSERT INTO tags VALUES (1, '{a,b}'), (2, '{c}');",
                "CREATE TABLE maybe_tags (id int PRIMARY KEY, vals text[]);",
                "INSERT INTO maybe_tags VALUES (1, '{a,b}'), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION concat_all(arr TEXT[]) RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            col TEXT;
            result TEXT := '';
        BEGIN
            FOREACH col IN ARRAY arr LOOP
                result := result || col;
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT concat_all('{a,b,c}');",
                    expected: Expected::Rows {
                        columns: &[Column("concat_all", TEXT)],
                        rows: &[
                            &[T("abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT concat_all('{}'), concat_all('{a,NULL,b}');",
                    expected: Expected::Rows {
                        columns: &[Column("concat_all", TEXT), Column("concat_all", TEXT)],
                        rows: &[
                            &[T(""), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, concat_all(vals) FROM tags ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("concat_all", TEXT)],
                        rows: &[
                            &[T("1"), T("ab")],
                            &[T("2"), T("c")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION sum_halves() RETURNS NUMERIC
            LANGUAGE plpgsql
            AS $$
        DECLARE
            n NUMERIC;
            total NUMERIC := 0;
        BEGIN
            FOREACH n IN ARRAY ARRAY[1, 2, 3] LOOP
                total := total + n / 2;
            END LOOP;
            RETURN total;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum_halves();",
                    expected: Expected::Rows {
                        columns: &[Column("sum_halves", NUMERIC)],
                        rows: &[
                            &[T("3.00000000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_found(arr TEXT[]) RETURNS BOOLEAN
            LANGUAGE plpgsql
            AS $$
        DECLARE
            col TEXT;
        BEGIN
            FOREACH col IN ARRAY arr LOOP
            END LOOP;
            RETURN FOUND;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_found('{a}'), foreach_found('{}');",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_found", BOOL), Column("foreach_found", BOOL)],
                        rows: &[
                            &[T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_exit(arr TEXT[]) RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            col TEXT;
            result TEXT := '';
        BEGIN
            FOREACH col IN ARRAY arr LOOP
                CONTINUE WHEN col = 'skip';
                EXIT WHEN col = 'stop';
                result := result || col;
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_exit('{a,skip,b,stop,c}');",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_exit", TEXT)],
                        rows: &[
                            &[T("ab")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_labeled() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            r RECORD;
            col TEXT;
            result TEXT := '';
        BEGIN
            <<rows>>
            FOR r IN SELECT id, vals FROM tags ORDER BY id LOOP
                FOREACH col IN ARRAY r.vals LOOP
                    EXIT rows WHEN col = 'c';
                    result := result || r.id || col;
                END LOOP;
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_labeled();",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_labeled", TEXT)],
                        rows: &[
                            &[T("1a1b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_nested() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            a TEXT;
            b TEXT;
            result TEXT := '';
        BEGIN
            FOREACH a IN ARRAY ARRAY['1', '2'] LOOP
                FOREACH b IN ARRAY ARRAY['x', 'y'] LOOP
                    result := result || a || b;
                END LOOP;
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_nested();",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_nested", TEXT)],
                        rows: &[
                            &[T("1x1y2x2y")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_nested_one_line() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            a TEXT;
            b TEXT;
            result TEXT := '';
        BEGIN
            FOREACH a IN ARRAY ARRAY['1', '2'] LOOP FOREACH b IN ARRAY ARRAY['x', 'y'] LOOP result := result || a || b; END LOOP; END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_nested_one_line();",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_nested_one_line", TEXT)],
                        rows: &[
                            &[T("1x1y2x2y")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_nested_one_var() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            x TEXT;
            result TEXT := '';
        BEGIN
            FOREACH x IN ARRAY ARRAY['1', '2'] LOOP FOREACH x IN ARRAY ARRAY['a', 'b'] LOOP result := result || x; END LOOP; END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_nested_one_var();",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_nested_one_var", TEXT)],
                        rows: &[
                            &[T("abab")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_quoted() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            "MyCol" TEXT;
            result TEXT := '';
        BEGIN
            FOREACH "MyCol" IN ARRAY ARRAY['x', 'y'] LOOP
                result := result || "MyCol";
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_quoted();",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_quoted", TEXT)],
                        rows: &[
                            &[T("xy")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_query() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            col TEXT;
            result TEXT := '';
        BEGIN
            FOREACH col IN ARRAY (SELECT vals FROM tags WHERE id = 1) LOOP
                DELETE FROM tags WHERE id = 1;
                result := result || col;
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_query();",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_query", TEXT)],
                        rows: &[
                            &[T("ab")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_null() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            col TEXT;
        BEGIN
            FOREACH col IN ARRAY CAST(NULL AS TEXT[]) LOOP
            END LOOP;
            RETURN 'no raise';
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_null();",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "FOREACH expression must not be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_null_column(tag_id INT) RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            arr TEXT[];
            col TEXT;
            result TEXT := '';
        BEGIN
            SELECT vals INTO arr FROM maybe_tags WHERE id = tag_id;
            FOREACH col IN ARRAY arr LOOP
                result := result || col;
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_null_column(1);",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_null_column", TEXT)],
                        rows: &[
                            &[T("ab")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_null_column(3);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "FOREACH expression must not be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_undefaulted() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            arr TEXT[];
            col TEXT;
            result TEXT := 'none';
        BEGIN
            FOREACH col IN ARRAY arr LOOP
                result := 'ran';
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_undefaulted();",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "FOREACH expression must not be null", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_not_array() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            col TEXT;
        BEGIN
            FOREACH col IN ARRAY 42 LOOP
            END LOOP;
            RETURN 'no raise';
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_not_array();",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "FOREACH expression must yield an array, not type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_bare_null() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            col TEXT;
        BEGIN
            FOREACH col IN ARRAY NULL LOOP
            END LOOP;
            RETURN 'no raise';
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_bare_null();",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "FOREACH expression must not be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_composite() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            r RECORD;
            result TEXT := '';
        BEGIN
            FOREACH r IN ARRAY (SELECT array_agg(maybe_tags) FROM maybe_tags) LOOP
                result := result || r.id;
            END LOOP;
            RETURN result;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_composite();",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION foreach_slice() RETURNS TEXT
            LANGUAGE plpgsql
            AS $$
        DECLARE
            row1 TEXT[];
        BEGIN
            FOREACH row1 SLICE 1 IN ARRAY ARRAY[['a', 'b']] LOOP
            END LOOP;
            RETURN 'done';
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT foreach_slice();",
                    expected: Expected::Rows {
                        columns: &[Column("foreach_slice", TEXT)],
                        rows: &[
                            &[T("done")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "use multiple types returning in block statement",
            set_up_script: &[
                "CREATE TABLE test (id int, v text);",
                "INSERT INTO test VALUES (1, 'r'), (2, 'g');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION return_table() RETURNS int AS $$ 
DECLARE ti int; tt text; 
BEGIN 
	INSERT INTO test VALUES (3, 'w') returning * INTO ti, tt; 
	RETURN ti; 
END; 
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select return_table();",
                    expected: Expected::Rows {
                        columns: &[Column("return_table", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "distinct expression over function",
            set_up_script: &[
                "CREATE TABLE test (pk SERIAL PRIMARY KEY);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION func1() RETURNS INT4 AS $$
DECLARE
  ret INT4;
BEGIN
  INSERT INTO test VALUES (DEFAULT) RETURNING pk INTO ret;
  IF ret % 2 <> 0 THEN
    RETURN NULL;
  END IF;
  RETURN ret;
END; $$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT func1() IS DISTINCT FROM func1() as DISTINCT_RESULT;",
                    expected: Expected::Rows {
                        columns: &[Column("distinct_result", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "function with single OUT parameter",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION calculate_bonus(
    IN current_salary NUMERIC,
    OUT new_total_salary NUMERIC
) AS $$
BEGIN
    -- Calculate the new total with a 10% bonus
    new_total_salary := current_salary + current_salary * 0.10;
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT calculate_bonus(5000);",
                    expected: Expected::Rows {
                        columns: &[Column("calculate_bonus", NUMERIC)],
                        rows: &[
                            &[T("5500.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM calculate_bonus(5000);",
                    expected: Expected::Rows {
                        columns: &[Column("new_total_salary", NUMERIC)],
                        rows: &[
                            &[T("5500.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT new_total_salary FROM calculate_bonus(5000);",
                    expected: Expected::Rows {
                        columns: &[Column("new_total_salary", NUMERIC)],
                        rows: &[
                            &[T("5500.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "function with multiple OUT parameter",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION calculate_bonus(
    IN current_salary NUMERIC,
    OUT bonus_amount NUMERIC,
    OUT new_total_salary NUMERIC
) AS $$
BEGIN
    -- Calculate a 10% bonus
    bonus_amount := current_salary * 0.10;
    
    -- Calculate the new total
    new_total_salary := current_salary + bonus_amount;
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT calculate_bonus(5000);",
                    expected: Expected::Rows {
                        columns: &[Column("calculate_bonus", RECORD)],
                        rows: &[
                            &[T("(500.00,5500.00)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM calculate_bonus(5000);",
                    expected: Expected::Rows {
                        columns: &[Column("bonus_amount", NUMERIC), Column("new_total_salary", NUMERIC)],
                        rows: &[
                            &[T("500.00"), T("5500.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bonus_amount FROM calculate_bonus(5000);",
                    expected: Expected::Rows {
                        columns: &[Column("bonus_amount", NUMERIC)],
                        rows: &[
                            &[T("500.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS TABLE in a FROM clause",
            set_up_script: &[
                r#"CREATE FUNCTION figures() RETURNS TABLE(shape TEXT, sides INT)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 'triangle', 3;
						RETURN QUERY SELECT 'square', 4;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM figures();",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT), Column("sides", INT4)],
                        rows: &[
                            &[T("triangle"), T("3")],
                            &[T("square"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT shape FROM figures() WHERE sides = 4;",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT)],
                        rows: &[
                            &[T("square")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT figures();",
                    expected: Expected::Rows {
                        columns: &[Column("figures", RECORD)],
                        rows: &[
                            &[T("(triangle,3)")],
                            &[T("(square,4)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS TABLE with a single column in a FROM clause",
            set_up_script: &[
                r#"CREATE FUNCTION shapes() RETURNS TABLE(shape TEXT)
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 'triangle';
						RETURN QUERY SELECT 'square';
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM shapes();",
                    expected: Expected::Rows {
                        columns: &[Column("shape", TEXT)],
                        rows: &[
                            &[T("triangle")],
                            &[T("square")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT shapes();",
                    expected: Expected::Rows {
                        columns: &[Column("shapes", TEXT)],
                        rows: &[
                            &[T("triangle")],
                            &[T("square")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNS SETOF a scalar type",
            set_up_script: &[
                r#"CREATE FUNCTION sides() RETURNS SETOF INT
					LANGUAGE plpgsql
					AS $$
					BEGIN
						RETURN QUERY SELECT 3 UNION ALL SELECT 4;
					END;
					$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM sides();",
                    expected: Expected::Rows {
                        columns: &[Column("sides", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sides();",
                    expected: Expected::Rows {
                        columns: &[Column("sides", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_plpgsql_rules() {
    run_scripts(&[
        ScriptTest {
            name: "PL/pgSQL control flow",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION classify(n int) RETURNS text LANGUAGE plpgsql AS $$ BEGIN IF n < 0 THEN RETURN 'negative'; ELSIF n = 0 THEN RETURN 'zero'; ELSE RETURN 'positive'; END IF; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT classify(-5), classify(0), classify(7);",
                    expected: Expected::Rows {
                        columns: &[Column("classify", TEXT), Column("classify", TEXT), Column("classify", TEXT)],
                        rows: &[
                            &[T("negative"), T("zero"), T("positive")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION case_of(n int) RETURNS text LANGUAGE plpgsql AS $$ BEGIN CASE n WHEN 1, 2 THEN RETURN 'small'; WHEN 3 THEN RETURN 'three'; END CASE; RETURN 'unreachable'; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT case_of(2), case_of(3);",
                    expected: Expected::Rows {
                        columns: &[Column("case_of", TEXT), Column("case_of", TEXT)],
                        rows: &[
                            &[T("small"), T("three")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT case_of(9);",
                    expected: Expected::Error(Diagnostic { code: "20000", message: "case not found", hint: "CASE statement is missing ELSE part.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION searched_case(n int) RETURNS text LANGUAGE plpgsql AS $$ BEGIN CASE WHEN n > 10 THEN RETURN 'big'; ELSE RETURN 'little'; END CASE; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT searched_case(11), searched_case(1);",
                    expected: Expected::Rows {
                        columns: &[Column("searched_case", TEXT), Column("searched_case", TEXT)],
                        rows: &[
                            &[T("big"), T("little")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION loops(n int) RETURNS text LANGUAGE plpgsql AS $$ DECLARE out text := ''; i int := 0; BEGIN <<outer>> LOOP i := i + 1; EXIT outer WHEN i > n; CONTINUE WHEN i = 2; out := out || i; END LOOP; WHILE i > 0 LOOP i := i - 2; out := out || '-'; END LOOP; FOR j IN REVERSE 10..1 BY 4 LOOP out := out || j; END LOOP; RETURN out; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT loops(4);",
                    expected: Expected::Rows {
                        columns: &[Column("loops", TEXT)],
                        rows: &[
                            &[T("134---1062")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION query_loop() RETURNS text LANGUAGE plpgsql AS $$ DECLARE r record; out text := ''; BEGIN FOR r IN SELECT * FROM (VALUES (1, 'a'), (2, 'b')) v(n, s) ORDER BY n LOOP out := out || r.n || r.s; END LOOP; RETURN out; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT query_loop();",
                    expected: Expected::Rows {
                        columns: &[Column("query_loop", TEXT)],
                        rows: &[
                            &[T("1a2b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION array_loop(a int[]) RETURNS int LANGUAGE plpgsql AS $$ DECLARE x int; total int := 0; BEGIN FOREACH x IN ARRAY a LOOP total := total + x; END LOOP; RETURN total; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_loop(ARRAY[1, 2, 3]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_loop", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_loop(NULL);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "FOREACH expression must not be null", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION exit_outside() RETURNS int LANGUAGE plpgsql AS $$ BEGIN EXIT; END $$;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "EXIT cannot be used outside a loop, unless it has a label", position: 73, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION no_return(n int) RETURNS int LANGUAGE plpgsql AS $$ BEGIN n := n + 1; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT no_return(1);",
                    expected: Expected::Error(Diagnostic { code: "2F005", message: "control reached end of function without RETURN", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "PL/pgSQL statements and FOUND",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE items (id int PRIMARY KEY, name text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items VALUES (1, 'one'), (2, 'two');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION lookup(want int) RETURNS text LANGUAGE plpgsql AS $$ DECLARE result text; BEGIN SELECT name INTO result FROM items WHERE id = want; IF NOT FOUND THEN RETURN 'missing'; END IF; RETURN result; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lookup(1), lookup(3);",
                    expected: Expected::Rows {
                        columns: &[Column("lookup", TEXT), Column("lookup", TEXT)],
                        rows: &[
                            &[T("one"), T("missing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION strict_lookup(want int) RETURNS text LANGUAGE plpgsql AS $$ DECLARE result text; BEGIN SELECT name INTO STRICT result FROM items WHERE id = want OR want = 0; RETURN result; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT strict_lookup(2);",
                    expected: Expected::Rows {
                        columns: &[Column("strict_lookup", TEXT)],
                        rows: &[
                            &[T("two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT strict_lookup(3);",
                    expected: Expected::Error(Diagnostic { code: "P0002", message: "query returned no rows", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT strict_lookup(0);",
                    expected: Expected::Error(Diagnostic { code: "P0003", message: "query returned more than one row", hint: "Make sure the query returns a single row, or use LIMIT 1.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION bump() RETURNS boolean LANGUAGE plpgsql AS $$ BEGIN UPDATE items SET name = name || '!' WHERE id = 1; RETURN FOUND; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bump();",
                    expected: Expected::Rows {
                        columns: &[Column("bump", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM items WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("one!")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION discard() RETURNS int LANGUAGE plpgsql AS $$ BEGIN SELECT 1; RETURN 1; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT discard();",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "query has no destination for result data", hint: "If you want to discard the results of a SELECT, use PERFORM instead.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION performs() RETURNS boolean LANGUAGE plpgsql AS $$ BEGIN PERFORM 1 FROM items WHERE id = 99; RETURN FOUND; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT performs();",
                    expected: Expected::Rows {
                        columns: &[Column("performs", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION dynamic(t text, want int) RETURNS text LANGUAGE plpgsql AS $$ DECLARE result text; BEGIN EXECUTE 'SELECT name FROM ' || t || ' WHERE id = $1' INTO result USING want; RETURN result; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT dynamic('items', 2);",
                    expected: Expected::Rows {
                        columns: &[Column("dynamic", TEXT)],
                        rows: &[
                            &[T("two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION null_dynamic() RETURNS void LANGUAGE plpgsql AS $$ BEGIN EXECUTE NULL; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT null_dynamic();",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "query string argument of EXECUTE is null", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION all_items() RETURNS SETOF items LANGUAGE plpgsql AS $$ BEGIN RETURN QUERY SELECT * FROM items ORDER BY id; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION wrong_shape() RETURNS TABLE(a int, b int) LANGUAGE plpgsql AS $$ BEGIN RETURN QUERY SELECT 1, 'x'::text; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM wrong_shape();",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "structure of query does not match function result type", detail: "Returned type text does not match expected type integer in column 2.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION too_few() RETURNS TABLE(a int, b int) LANGUAGE plpgsql AS $$ BEGIN RETURN QUERY SELECT 1; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM too_few();",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "structure of query does not match function result type", detail: "Number of returned columns (1) does not match expected column count (2).", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION pairs() RETURNS TABLE(a int, b text) LANGUAGE plpgsql AS $$ BEGIN RETURN QUERY SELECT id, name FROM items ORDER BY id; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pairs();",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("one!")],
                            &[T("2"), T("two")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION out_params(x int, OUT doubled int, OUT label text) LANGUAGE plpgsql AS $$ BEGIN doubled := x * 2; label := 'n' || x; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT out_params(4);",
                    expected: Expected::Rows {
                        columns: &[Column("out_params", RECORD)],
                        rows: &[
                            &[T("(8,n4)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM out_params(5);",
                    expected: Expected::Rows {
                        columns: &[Column("doubled", INT4), Column("label", TEXT)],
                        rows: &[
                            &[T("10"), T("n5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION void_result() RETURNS void LANGUAGE plpgsql AS $$ BEGIN END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT void_result(), void_result() IS NULL;",
                    expected: Expected::Rows {
                        columns: &[Column("void_result", VOID), Column("?column?", BOOL)],
                        rows: &[
                            &[T(""), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "PL/pgSQL RAISE",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION notices(n int) RETURNS int LANGUAGE plpgsql AS $$ BEGIN RAISE NOTICE 'n is %, doubled %', n, n * 2; RAISE WARNING 'careful'; RAISE INFO 'info %%'; RAISE DEBUG 'hidden'; RETURN n; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT notices(3);",
                    expected: Expected::Rows {
                        columns: &[Column("notices", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    notices: &[Diagnostic { code: "00000", message: "n is 3, doubled 6", ..N }, Diagnostic { severity: "WARNING", code: "01000", message: "careful", ..E }, Diagnostic { severity: "INFO", code: "00000", message: "info %", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION fails(n int) RETURNS int LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'bad value: %', n USING HINT = 'try ' || (n + 1), DETAIL = 'detail here', ERRCODE = '22023'; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT fails(1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "bad value: 1", detail: "detail here", hint: "try 2", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION fails_default() RETURNS int LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'plain'; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT fails_default();",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "plain", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION fails_named() RETURNS int LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'dup' USING ERRCODE = 'unique_violation'; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT fails_named();",
                    expected: Expected::Error(Diagnostic { code: "23505", message: "dup", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION null_param() RETURNS int LANGUAGE plpgsql AS $$ BEGIN RAISE NOTICE 'value: %', NULL::int; RETURN 1; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT null_param();",
                    expected: Expected::Rows {
                        columns: &[Column("null_param", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    notices: &[Diagnostic { code: "00000", message: "value: <NULL>", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION too_many() RETURNS int LANGUAGE plpgsql AS $$ BEGIN RAISE NOTICE 'x', 1; RETURN 1; END $$;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "too many parameters specified for RAISE", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "PL/pgSQL variables and declarations",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION defaults(p int) RETURNS text LANGUAGE plpgsql AS $$ DECLARE a int := p * 2; b int := a + 1; c text := 'a=' || a || ' b=' || b; BEGIN RETURN c; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT defaults(5);",
                    expected: Expected::Rows {
                        columns: &[Column("defaults", TEXT)],
                        rows: &[
                            &[T("a=10 b=11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION aliases(input text) RETURNS text LANGUAGE plpgsql AS $$ DECLARE v text; BEGIN DECLARE a1 ALIAS FOR v; a2 ALIAS FOR input; BEGIN a1 := a2 || '?'; END; RETURN v; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT aliases('x');",
                    expected: Expected::Rows {
                        columns: &[Column("aliases", TEXT)],
                        rows: &[
                            &[T("x?")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION positional(int, int) RETURNS int LANGUAGE plpgsql AS $$ BEGIN RETURN $1 * $2; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT positional(6, 7);",
                    expected: Expected::Rows {
                        columns: &[Column("positional", INT4)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION bad_type() RETURNS int LANGUAGE plpgsql AS $$ DECLARE b pg_catalog.integer; BEGIN RETURN 1; END $$;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.integer" does not exist"#, position: 73, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION bad_syntax() RETURNS int LANGUAGE plpgsql AS $$ DECLARE b pg_catalog.double precision; BEGIN RETURN 1; END $$;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "precision""#, position: 93, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION missing_type() RETURNS int LANGUAGE plpgsql AS $$ DECLARE b no_such_type; BEGIN RETURN 1; END $$;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "no_such_type" does not exist"#, position: 77, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE things (id int, label text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO things VALUES (1, 'a');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION row_types() RETURNS text LANGUAGE plpgsql AS $$ DECLARE r things; l things.label%TYPE; BEGIN SELECT * INTO r FROM things; l := r.label || '!'; RETURN r.id || l; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_types();",
                    expected: Expected::Rows {
                        columns: &[Column("row_types", TEXT)],
                        rows: &[
                            &[T("1a!")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION records() RETURNS text LANGUAGE plpgsql AS $$ DECLARE r record := ROW(1, 'a'); BEGIN SELECT 2 AS x, 'b' AS y INTO r; RETURN r.x || r.y; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT records();",
                    expected: Expected::Rows {
                        columns: &[Column("records", TEXT)],
                        rows: &[
                            &[T("2b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION unassigned() RETURNS int LANGUAGE plpgsql AS $$ DECLARE r record; BEGIN RETURN r.x; END $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unassigned();",
                    expected: Expected::Error(Diagnostic { code: "55000", message: r#"record "r" is not assigned yet"#, detail: "The tuple structure of a not-yet-assigned record is indeterminate.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
