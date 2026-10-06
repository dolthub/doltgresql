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
fn test_create_trigger() {
    run_scripts(&[
        ScriptTest {
            name: "BEFORE INSERT, with columns omitted from or reordered by the INSERT",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, selector TEXT DEFAULT 'DEFAULTED', result TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.result := 'saw_' || NEW.selector || '_' || NEW.pk::text;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test (pk, selector) VALUES (1, 'TRIGGER_B');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test (selector, pk) VALUES ('DIRECT_A', 2);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test (pk) VALUES (3);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("selector", TEXT), Column("result", TEXT)],
                        rows: &[
                            &[T("1"), T("TRIGGER_B"), T("saw_TRIGGER_B_1")],
                            &[T("2"), T("DIRECT_A"), T("saw_DIRECT_A_2")],
                            &[T("3"), T("DEFAULTED"), T("saw_DEFAULTED_3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE INSERT ... SELECT, with columns omitted from or reordered by the INSERT",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, selector TEXT DEFAULT 'DEFAULTED', result TEXT);",
                "CREATE TABLE source (pk INT PRIMARY KEY, selector TEXT);",
                "INSERT INTO source VALUES (1, 'FROM_QUERY'), (2, 'ALSO_FROM_QUERY');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.result := 'saw_' || NEW.selector || '_' || NEW.pk::text;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test (selector, pk) SELECT selector, pk FROM source ORDER BY pk;",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test (pk) SELECT 3;",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("selector", TEXT), Column("result", TEXT)],
                        rows: &[
                            &[T("1"), T("FROM_QUERY"), T("saw_FROM_QUERY_1")],
                            &[T("2"), T("ALSO_FROM_QUERY"), T("saw_ALSO_FROM_QUERY_2")],
                            &[T("3"), T("DEFAULTED"), T("saw_DEFAULTED_3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE INSERT",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi_1")],
                            &[T("2"), T("there_2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE UPDATE",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET v1 = v1 || '|' WHERE pk IN (1, 2);",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi|_1")],
                            &[T("2"), T("there|_2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE DELETE",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
				RETURN OLD;
			END;
			$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE DELETE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE INSERT returning NULL",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
					INSERT INTO test2 VALUES (NEW.pk, NEW.v1);
					RETURN NULL;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                    expected: Expected::Tag("INSERT 0 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi_1")],
                            &[T("2"), T("there_2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE UPDATE returning NULL",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
				INSERT INTO test2 VALUES (NEW.pk, NEW.v1);
				RETURN NULL;
			END;
			$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET v1 = v1 || '|' WHERE pk IN (1, 2);",
                    expected: Expected::Tag("UPDATE 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi|_1")],
                            &[T("2"), T("there|_2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE DELETE returning NULL",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
				RETURN NULL;
			END;
			$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE DELETE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = 1;",
                    expected: Expected::Tag("DELETE 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE UPDATE with DELETE DML",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
				RETURN OLD;
			END;
			$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "AFTER INSERT",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
					INSERT INTO test2 VALUES (NEW.pk, NEW.v1);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger AFTER INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi_1")],
                            &[T("2"), T("there_2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "AFTER UPDATE",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
				INSERT INTO test2 VALUES (NEW.pk, NEW.v1);
				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger AFTER UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET v1 = v1 || '|' WHERE pk IN (1, 2);",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi|")],
                            &[T("2"), T("there|")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi|_1")],
                            &[T("2"), T("there|_2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "AFTER DELETE returning NULL",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
				RETURN NULL;
			END;
			$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger AFTER DELETE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Cascading DELETE into INSERT, different tables",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
	RETURN OLD;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION trigger_func2() RETURNS TRIGGER AS $$
BEGIN
	NEW.pk := NEW.pk + 100;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE DELETE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test2 FOR EACH ROW EXECUTE FUNCTION trigger_func2();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("101"), T("hi")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Cascading INSERT into UPDATE, same table",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	UPDATE test SET v1 = v1 || NEW.pk::text;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION trigger_func2() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || '_u';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
                "CREATE TRIGGER test_trigger2 BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func2();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hi2_u")],
                            &[T("2"), T("there")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Multiple triggers on same table",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func_a() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || 'a';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION trigger_func_c() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || 'c';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION trigger_func_b() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || 'b';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger_b BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func_b();",
                "CREATE TRIGGER test_trigger_a BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func_a();",
                "CREATE TRIGGER test_trigger_c BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func_c();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("1"), T("hiabc")],
                            &[T("2"), T("thereabc")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Stack depth limit exceeded",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	INSERT INTO test2 VALUES (NEW.pk+2, NEW.v1 || '_');
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION trigger_func2() RETURNS TRIGGER AS $$
BEGIN
	INSERT INTO test VALUES (NEW.pk+4, NEW.v1 || '|');
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test2 FOR EACH ROW EXECUTE FUNCTION trigger_func2();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                    expected: Expected::Error(Diagnostic { code: "54001", message: "stack depth limit exceeded", hint: r#"Increase the configuration parameter "max_stack_depth" (currently 2048kB), after ensuring the platform's stack depth limit is adequate."#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DELETE TABLE deletes attached triggers",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || '_';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
                "CREATE TRIGGER test_trigger2 BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"trigger "test_trigger" for relation "test" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER test_trigger2 BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"trigger "test_trigger2" for relation "test" already exists"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE test;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER test_trigger2 BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "WHEN on BEFORE INSERT",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func1() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.pk::text || '_' || NEW.v1;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION trigger_func2() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger1 BEFORE INSERT ON test FOR EACH ROW WHEN (NEW.pk < 1) EXECUTE FUNCTION trigger_func1();",
                "CREATE TRIGGER test_trigger2 BEFORE INSERT ON test FOR EACH ROW WHEN (NEW.pk > 1) EXECUTE FUNCTION trigger_func2();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (0, 'hi'), (1, 'there'), (2, 'dude');",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v1", TEXT)],
                        rows: &[
                            &[T("0"), T("0_hi")],
                            &[T("1"), T("there")],
                            &[T("2"), T("dude_2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: Postgres rejects the non-boolean WHEN when the trigger is created, so the creation is asserted.
        ScriptTest {
            name: "WHEN with non-boolean expression",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.pk::text || '_' || NEW.v1;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW WHEN (NEW.pk + 1) EXECUTE FUNCTION trigger_func();",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of WHEN must be type boolean, not type integer", position: 70, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                    expected: Expected::Tag("INSERT 0 2"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Table as type",
            set_up_script: &[
                "CREATE TABLE test (id INT4 PRIMARY KEY, name TEXT NOT NULL, qty INT4 NOT NULL, price REAL NOT NULL);",
                r#"CREATE FUNCTION trigger_func() RETURNS trigger AS $$
DECLARE
	rec test;
BEGIN
	rec := NEW;
	IF rec.qty < 0 THEN
		rec.qty := -rec.qty;
	END IF;
	NEW := rec;
	RETURN NEW;
END; $$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'apple', 3, 2.5), (2, 'banana', -5, -1.2);",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("qty", INT4), Column("price", FLOAT4)],
                        rows: &[
                            &[T("1"), T("apple"), T("3"), T("2.5")],
                            &[T("2"), T("banana"), T("5"), T("-1.2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DECLARE default referencing the trigger records",
            set_up_script: &[
                "CREATE TABLE test (id INT4 PRIMARY KEY, val TEXT);",
                "CREATE TABLE log (msg TEXT);",
                "INSERT INTO test VALUES (7, 'a');",
                r#"CREATE FUNCTION trigger_func() RETURNS trigger AS $$
DECLARE
	old_id INT4 := OLD.id;
	changed BOOLEAN := OLD.val <> NEW.val;
BEGIN
	INSERT INTO log VALUES ('id=' || old_id || ' changed=' || changed);
	RETURN NEW;
END; $$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET val = 'b' WHERE id = 7;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT msg FROM log;",
                    expected: Expected::Rows {
                        columns: &[Column("msg", TEXT)],
                        rows: &[
                            &[T("id=7 changed=true")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "trigger to call procedure that updates another table using dynamic execute",
            set_up_script: &[
                r#"create table public."Collections"(
				 id uuid PRIMARY KEY NOT NULL,
				 name text not null,
				 username varchar(28) not null,
				 total_tracks integer DEFAULT 0);"#,
                r#"INSERT INTO public."Collections" (id, name, username, total_tracks) VALUES ('550e8400-e29b-41d4-a716-446655440000', 'My Custom Playlist', 'user_alpha', 10);"#,
                r#"create table public."CollectionItems"(
			collection_id uuid not null,
			track_id integer not null);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE FUNCTION update_collections()
  RETURNS trigger AS $$
  DECLARE
    BEGIN
    IF TG_OP = 'INSERT' THEN
      EXECUTE 'update public."Collections" set total_tracks=total_tracks+1 where id = $1;'
      USING NEW.collection_id;
    END IF;

    IF TG_OP = 'DELETE' THEN 
      EXECUTE 'update public."Collections" set total_tracks=total_tracks-1 where id = $1;'
      USING OLD.collection_id;
    END IF;
    
    RETURN NEW;
    END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE TRIGGER update_collection
				AFTER INSERT OR DELETE ON public."CollectionItems"
				FOR EACH ROW EXECUTE PROCEDURE update_collections();"#,
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO public."CollectionItems" (collection_id, track_id) VALUES ('550e8400-e29b-41d4-a716-446655440000', 101);"#,
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT total_tracks FROM public."Collections""#,
                    expected: Expected::Rows {
                        columns: &[Column("total_tracks", INT4)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "FOREACH over an array of column names",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, retired_at TEXT, note TEXT);",
                "INSERT INTO test VALUES (1, NULL, 'n');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			DECLARE
				permitted TEXT[] := '{retired_at}';
				col TEXT;
				changed INT := 0;
			BEGIN
				FOREACH col IN ARRAY permitted LOOP
					IF col = 'retired_at' AND NEW.retired_at IS DISTINCT FROM OLD.retired_at THEN
						changed := changed + 1;
						IF OLD.retired_at IS NOT NULL THEN
							RAISE EXCEPTION 'test %: % is already set', OLD.pk, col;
						END IF;
					END IF;
				END LOOP;
				IF changed = 0 THEN
					RAISE EXCEPTION 'test % is append-only', OLD.pk;
				END IF;
				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET note = 'other' WHERE pk = 1;",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "test 1 is append-only", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET retired_at = 'now' WHERE pk = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET retired_at = 'later' WHERE pk = 1;",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "test 1: retired_at is already set", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("retired_at", TEXT), Column("note", TEXT)],
                        rows: &[
                            &[T("1"), T("now"), T("n")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DROP TRIGGER",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "DROP TRIGGER test_trigger ON test;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER IF EXISTS test_trigger ON test;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    notices: &[Diagnostic { code: "00000", message: r#"trigger "test_trigger" for relation "test" does not exist, skipping"#, ..N }],
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "OLD.* IS DISTINCT FROM NEW.* in a trigger",
            set_up_script: &[
                "CREATE TABLE t3336_issue (a INT);",
                "INSERT INTO t3336_issue VALUES (1);",
                "CREATE FUNCTION f3336_issue() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE NOTICE 'the trigger ran'; RETURN NEW; END $$;",
                "CREATE TRIGGER tr BEFORE UPDATE ON t3336_issue FOR EACH ROW WHEN (old.* IS DISTINCT FROM new.*) EXECUTE FUNCTION f3336_issue();",
                "CREATE TABLE t3336 (a INT PRIMARY KEY, b TEXT);",
                "CREATE TABLE t3336_log (a INT, src TEXT);",
                "CREATE FUNCTION f3336_when() RETURNS trigger AS $$ BEGIN INSERT INTO t3336_log VALUES (NEW.a, 'when'); RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3336_body() RETURNS trigger AS $$ BEGIN IF OLD.* IS DISTINCT FROM NEW.* THEN INSERT INTO t3336_log VALUES (NEW.a, 'body'); END IF; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER tr3336_when AFTER UPDATE ON t3336 FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM NEW.*) EXECUTE FUNCTION f3336_when();",
                "CREATE TRIGGER tr3336_body AFTER UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_body();",
                "INSERT INTO t3336 VALUES (1, 'x'), (2, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t3336_issue SET a = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    notices: &[Diagnostic { code: "00000", message: "the trigger ran", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3336_issue;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ROW(2, NULL::TEXT) IS DISTINCT FROM ROW(2, 'y'::TEXT), ROW(2, NULL::TEXT) IS DISTINCT FROM ROW(2, NULL::TEXT), ROW(2, NULL::TEXT) IS NOT DISTINCT FROM ROW(2, NULL::TEXT), ROW(1, 2) IS DISTINCT FROM ROW(1, 3), ROW(1, 2) IS NOT DISTINCT FROM ROW(1, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'x' WHERE a = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(*) FROM t3336_log;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'y' WHERE a = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = NULL WHERE a = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = NULL WHERE a = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'z' WHERE a = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3336_log ORDER BY a, src;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("src", TEXT)],
                        rows: &[
                            &[T("1"), T("body")],
                            &[T("1"), T("when")],
                            &[T("2"), T("body")],
                            &[T("2"), T("body")],
                            &[T("2"), T("when")],
                            &[T("2"), T("when")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Whole-row references outside of a record comparison are rejected",
            set_up_script: &[
                "CREATE TABLE t3336 (a INT PRIMARY KEY, b TEXT);",
                "CREATE TABLE t3336_one (a INT);",
                "CREATE TABLE t3336_bool (b BOOLEAN);",
                "CREATE TABLE t3336_log2 (v TEXT);",
                "INSERT INTO t3336 VALUES (1, 'x');",
                "INSERT INTO t3336_one VALUES (1);",
                "INSERT INTO t3336_bool VALUES (true);",
                "CREATE FUNCTION f3336() RETURNS TRIGGER AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3336_assign() RETURNS TRIGGER AS $$ DECLARE v INT; BEGIN v := NEW.*; INSERT INTO t3336_log2 VALUES ('assign ' || v); RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3336_if() RETURNS TRIGGER AS $$ BEGIN IF NEW.* THEN INSERT INTO t3336_log2 VALUES ('if'); END IF; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3336_eq() RETURNS TRIGGER AS $$ BEGIN IF NEW.* = 1 THEN RETURN NEW; END IF; RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3336_raise() RETURNS TRIGGER AS $$ BEGIN RAISE EXCEPTION 'val %', NEW.*; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3336_ret() RETURNS TRIGGER AS $$ BEGIN RETURN NEW.*; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_bad BEFORE UPDATE ON t3336 FOR EACH ROW WHEN (OLD.*) EXECUTE FUNCTION f3336();",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of WHEN must be type boolean, not type t3336", position: 69, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'q' WHERE a = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_bad ON t3336;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"trigger "tr3336_bad" for table "t3336" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_bad2 BEFORE UPDATE ON t3336 FOR EACH ROW WHEN (OLD.* = 1) EXECUTE FUNCTION f3336();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "operator does not exist: t3336 = integer", hint: "No operator matches the given name and argument types. You might need to add explicit type casts.", position: 76, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'q' WHERE a = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_bad2 ON t3336;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"trigger "tr3336_bad2" for table "t3336" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_assign BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_assign();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'q' WHERE a = 1;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "assignment source returned 2 columns", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_assign ON t3336;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_if BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_if();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'q' WHERE a = 1;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "query returned 2 columns", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_if ON t3336;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_eq BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_eq();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'q' WHERE a = 1;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "operator does not exist: t3336 = integer", hint: "No operator matches the given name and argument types. You might need to add explicit type casts.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_eq ON t3336;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_raise BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_raise();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'q' WHERE a = 1;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "query returned 2 columns", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_raise ON t3336;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_ret BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_ret();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336 SET b = 'q' WHERE a = 1;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "query returned 2 columns", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_ret ON t3336;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3336;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("q")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_assign1 BEFORE UPDATE ON t3336_one FOR EACH ROW EXECUTE FUNCTION f3336_assign();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336_one SET a = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_assign1 ON t3336_one;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_raise1 BEFORE UPDATE ON t3336_one FOR EACH ROW EXECUTE FUNCTION f3336_raise();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336_one SET a = 3;",
                    expected: Expected::Error(Diagnostic { code: "P0001", message: "val 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_raise1 ON t3336_one;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3336_if1 BEFORE UPDATE ON t3336_bool FOR EACH ROW EXECUTE FUNCTION f3336_if();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336_bool SET b = true;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t3336_bool SET b = false;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TRIGGER tr3336_if1 ON t3336_bool;",
                    expected: Expected::Tag("DROP TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3336_log2 ORDER BY v;",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT)],
                        rows: &[
                            &[T("assign 2")],
                            &[T("if")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t3336_one;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE OF specific columns",
            set_up_script: &[
                "CREATE TABLE test (id INT4 PRIMARY KEY, name TEXT, data TEXT, other TEXT);",
                "INSERT INTO test VALUES (1, 'a', 'b', 'c');",
                "CREATE TABLE log (msg TEXT);",
                r#"CREATE FUNCTION trig_func1() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log VALUES ('tr1: ' || OLD.name || ',' || OLD.data || ',' || OLD.other || ' -> ' || NEW.name || ',' || NEW.data || ',' || NEW.other);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION trig_func4() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log VALUES ('tr4: ' || OLD.name || ',' || OLD.data || ',' || OLD.other || ' -> ' || NEW.name || ',' || NEW.data || ',' || NEW.other);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr1 BEFORE UPDATE OF name, data ON test FOR EACH ROW EXECUTE PROCEDURE trig_func1();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr2 BEFORE UPDATE OF nope ON test FOR EACH ROW EXECUTE PROCEDURE trig_func1();",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "nope" of relation "test" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr3 BEFORE UPDATE OF name, name ON test FOR EACH ROW EXECUTE PROCEDURE trig_func1();",
                    expected: Expected::Error(Diagnostic { code: "42701", message: r#"column "name" specified more than once"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr4 AFTER INSERT OR UPDATE OF other ON test FOR EACH ROW EXECUTE PROCEDURE trig_func4();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET other = 'd';",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET name = 'e';",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET data = data;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM log ORDER BY msg;",
                    expected: Expected::Rows {
                        columns: &[Column("msg", TEXT)],
                        rows: &[
                            &[T("tr1: a,b,d -> e,b,d")],
                            &[T("tr1: e,b,d -> e,b,d")],
                            &[T("tr4: a,b,c -> a,b,d")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT tgname, tgattr::TEXT, tgtype FROM pg_trigger WHERE tgrelid = 'test'::regclass ORDER BY tgname;",
                    expected: Expected::Rows {
                        columns: &[Column("tgname", NAME), Column("tgattr", TEXT), Column("tgtype", INT2)],
                        rows: &[
                            &[T("tr1"), T("2 3"), T("19")],
                            &[T("tr4"), T("4"), T("21")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE OF columns that are renamed or dropped",
            set_up_script: &[
                "CREATE TABLE test (id INT4 PRIMARY KEY, name TEXT, data TEXT, other TEXT);",
                "INSERT INTO test VALUES (1, 'a', 'b', 'c');",
                "CREATE TABLE log (msg TEXT);",
                r#"CREATE FUNCTION trig_func1() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log VALUES ('tr1: ' || NEW.id);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION trig_func4() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log VALUES ('tr4: ' || NEW.id);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER tr1 BEFORE UPDATE OF name, data ON public.test FOR EACH ROW EXECUTE FUNCTION trig_func1();",
                "CREATE TRIGGER tr4 AFTER INSERT OR UPDATE OF other ON public.test FOR EACH ROW EXECUTE FUNCTION trig_func4();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"ALTER TABLE test RENAME COLUMN name TO "Name2";"#,
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"UPDATE test SET "Name2" = 'x';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM log ORDER BY msg;",
                    expected: Expected::Rows {
                        columns: &[Column("msg", TEXT)],
                        rows: &[
                            &[T("tr1: 1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE tgrelid = 'test'::regclass ORDER BY tgname;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_triggerdef", TEXT)],
                        rows: &[
                            &[T(r#"CREATE TRIGGER tr1 BEFORE UPDATE OF "Name2", data ON public.test FOR EACH ROW EXECUTE FUNCTION trig_func1()"#)],
                            &[T("CREATE TRIGGER tr4 AFTER INSERT OR UPDATE OF other ON public.test FOR EACH ROW EXECUTE FUNCTION trig_func4()")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT tgname, tgattr::TEXT FROM pg_trigger WHERE tgrelid = 'test'::regclass ORDER BY tgname;",
                    expected: Expected::Rows {
                        columns: &[Column("tgname", NAME), Column("tgattr", TEXT)],
                        rows: &[
                            &[T("tr1"), T("2 3")],
                            &[T("tr4"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test DROP COLUMN data;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop column data of table test because other objects depend on it", detail: "trigger tr1 on table test depends on column data of table test", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test DROP COLUMN other;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop column other of table test because other objects depend on it", detail: "trigger tr4 on table test depends on column other of table test", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE test DROP COLUMN data CASCADE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to trigger tr1 on table test", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT tgname FROM pg_trigger WHERE tgrelid = 'test'::regclass ORDER BY tgname;",
                    expected: Expected::Rows {
                        columns: &[Column("tgname", NAME)],
                        rows: &[
                            &[T("tr4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE OF triggers alongside an ordinary UPDATE trigger",
            set_up_script: &[
                "CREATE TABLE test (id INT4 PRIMARY KEY, name TEXT, data TEXT);",
                "INSERT INTO test VALUES (1, 'initial', 'initial');",
                "CREATE TABLE log (kind TEXT, id INT4, name TEXT, data TEXT);",
                r#"CREATE FUNCTION log_ordinary() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log VALUES ('ordinary', NEW.id, NEW.name, NEW.data);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION log_named_before() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log VALUES ('named_before', NEW.id, NEW.name, NEW.data);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION log_named_after() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log VALUES ('named_after', NEW.id, NEW.name, NEW.data);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER ordinary BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION log_ordinary();",
                "CREATE TRIGGER named_before BEFORE UPDATE OF name ON test FOR EACH ROW EXECUTE FUNCTION log_named_before();",
                "CREATE TRIGGER named_after AFTER UPDATE OF name ON test FOR EACH ROW EXECUTE FUNCTION log_named_after();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET data = 'data-only' WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test AS t SET data = 'aliased' WHERE t.id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET name = name WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM log ORDER BY data, kind;",
                    expected: Expected::Rows {
                        columns: &[Column("kind", TEXT), Column("id", INT4), Column("name", TEXT), Column("data", TEXT)],
                        rows: &[
                            &[T("named_after"), T("1"), T("initial"), T("aliased")],
                            &[T("named_before"), T("1"), T("initial"), T("aliased")],
                            &[T("ordinary"), T("1"), T("initial"), T("aliased")],
                            &[T("ordinary"), T("1"), T("initial"), T("aliased")],
                            &[T("ordinary"), T("1"), T("initial"), T("data-only")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_trigger_whole_record_reference() {
    run_scripts(&[
        ScriptTest {
            name: "whole record passed to a function",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE log (id SERIAL PRIMARY KEY, which TEXT, j JSONB);",
                "INSERT INTO test VALUES (1, 'hi');",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				DECLARE
					old_v jsonb;
					new_v jsonb;
				BEGIN
					old_v := to_jsonb(OLD);
					new_v := to_jsonb(NEW);
					INSERT INTO log (which, j) VALUES ('old', old_v), ('new', new_v);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET v1 = 'bye' WHERE pk = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT which, j::text FROM log ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("which", TEXT), Column("j", TEXT)],
                        rows: &[
                            &[T("old"), T(r#"{"pk": 1, "v1": "hi"}"#)],
                            &[T("new"), T(r#"{"pk": 1, "v1": "bye"}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "whole record as text",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT, b BOOL);",
                "CREATE TABLE log (id SERIAL PRIMARY KEY, t TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (t) VALUES (NEW::text);
					RAISE NOTICE 'row: %', NEW;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"INSERT INTO test VALUES (1, 'a,b"c', true);"#,
                    expected: Expected::Tag("INSERT 0 1"),
                    notices: &[Diagnostic { code: "00000", message: r#"row: (1,"a,b""c",t)"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT t FROM log ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("t", TEXT)],
                        rows: &[
                            &[T(r#"(1,"a,b""c",t)"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "whole record with a NULL field",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE log (id SERIAL PRIMARY KEY, j TEXT, t TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (j, t) VALUES (to_jsonb(NEW)::text, NEW::text);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT j, t FROM log ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("j", TEXT), Column("t", TEXT)],
                        rows: &[
                            &[T(r#"{"pk": 1, "v1": null}"#), T("(1,)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "whole record whose field is named like the record",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, record TEXT);",
                "CREATE TABLE log (id SERIAL PRIMARY KEY, j TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (j) VALUES (to_jsonb(NEW)::text);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT j FROM log ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("j", TEXT)],
                        rows: &[
                            &[T(r#"{"pk": 1, "record": "hi"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "records compared as a whole",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
                "CREATE TABLE log (id SERIAL PRIMARY KEY, msg TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					IF NEW IS DISTINCT FROM OLD THEN
						INSERT INTO log (msg) VALUES ('changed ' || NEW.pk::text);
					ELSE
						INSERT INTO log (msg) VALUES ('same ' || NEW.pk::text);
					END IF;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET v1 = 'hi' WHERE pk IN (1, 2);",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT msg FROM log ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("msg", TEXT)],
                        rows: &[
                            &[T("same 1")],
                            &[T("changed 2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "records compared as a whole when a field leaves NULL",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT, v2 TEXT);",
                "INSERT INTO test VALUES (1, 'hi', NULL);",
                "CREATE TABLE log (id SERIAL PRIMARY KEY, msg TEXT);",
                r#"CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					IF OLD IS DISTINCT FROM NEW THEN
						INSERT INTO log (msg) VALUES ('changed ' || NEW.pk::text);
					ELSE
						INSERT INTO log (msg) VALUES ('same ' || NEW.pk::text);
					END IF;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE test SET v1 = 'bye' WHERE pk = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET v2 = 'now set' WHERE pk = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET v2 = NULL WHERE pk = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE test SET v2 = NULL WHERE pk = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT msg FROM log ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("msg", TEXT)],
                        rows: &[
                            &[T("changed 1")],
                            &[T("changed 1")],
                            &[T("changed 1")],
                            &[T("same 1")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "the record an operation does not supply",
            set_up_script: &[
                "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
                "CREATE TABLE log (id SERIAL PRIMARY KEY, o TEXT, n TEXT);",
                r#"CREATE FUNCTION insert_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (o, n) VALUES (to_jsonb(OLD)::text, to_jsonb(NEW)::text);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;"#,
                r#"CREATE FUNCTION delete_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (o, n) VALUES (to_jsonb(OLD)::text, to_jsonb(NEW)::text);
					RETURN OLD;
				END;
				$$ LANGUAGE plpgsql;"#,
                "CREATE TRIGGER t1 BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION insert_func();",
                "CREATE TRIGGER t2 BEFORE DELETE ON test FOR EACH ROW EXECUTE FUNCTION delete_func();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1, 'hi');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE pk = 1;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT o, n FROM log ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("o", TEXT), Column("n", TEXT)],
                        rows: &[
                            &[Null, T(r#"{"pk": 1, "v1": "hi"}"#)],
                            &[T(r#"{"pk": 1, "v1": "hi"}"#), Null],
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
