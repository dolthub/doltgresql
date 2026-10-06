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
fn test_create_procedure_language_plpgsql() {
    run_scripts(&[
        ScriptTest {
            name: "Simple example",
            set_up_script: &[
                "CREATE TABLE test (v1 INT8);",
                r#"CREATE PROCEDURE example(input INT8) AS $$
				BEGIN
					INSERT INTO test VALUES (input);
				END;
				$$ LANGUAGE 'plpgsql';"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CALL example(1);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL example('2');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
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
            name: "WHILE Label",
            set_up_script: &[
                "CREATE TABLE test (v1 INT8);",
                r#"CREATE PROCEDURE interpreted_while_label(input INT4) AS $$
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
	INSERT INTO test VALUES (input);
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CALL interpreted_while_label(42);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT8)],
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
            name: "Overloading",
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT);",
                r#"CREATE PROCEDURE interpreted_overload(input TEXT) AS $$
DECLARE
	var1 TEXT;
BEGIN
	IF length(input) > 3 THEN
		var1 := input || '_long';
	ELSE
		var1 := input;
	END IF;
	INSERT INTO test VALUES (var1);
END;
$$ LANGUAGE plpgsql;"#,
                r#"CREATE PROCEDURE interpreted_overload(input INT4) AS $$
DECLARE
	var1 INT4;
BEGIN
	IF input > 3 THEN
		var1 := -input;
	ELSE
		var1 := input;
	END IF;
	INSERT INTO test VALUES (var1::text);
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CALL interpreted_overload('abc');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_overload('abcd');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_overload(3);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_overload(4);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", TEXT)],
                        rows: &[
                            &[T("abc")],
                            &[T("abcd_long")],
                            &[T("3")],
                            &[T("-4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Branching",
            set_up_script: &[
                "CREATE TABLE test(v1 INT4, v2 INT4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE PROCEDURE interpreted_branch(input INT4) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 1;
	INSERT INTO test VALUES (1, input + 100);
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_branch(4);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("104")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE v1 = 1;",
                    expected: Expected::Tag("DELETE 1"),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
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
                    query: r#"CREATE OR REPLACE PROCEDURE interpreted_branch(input INT4) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 2;
	INSERT INTO test VALUES (2, input + 1000);
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'updated func')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_branch(56);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("2"), T("1056")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM test WHERE v1 = 2;",
                    expected: Expected::Tag("DELETE 1"),
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
                    query: "CALL interpreted_branch(57);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("157")],
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
                "CREATE TABLE test(v1 INT4, v2 INT4);",
                "INSERT INTO test VALUES (1, 77);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE PROCEDURE interpreted_merging(input TEXT) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 2;
	INSERT INTO test VALUES (2, input::int4 + 100);
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_merging('12');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("77")],
                            &[T("2"), T("112")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_merging(55);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure interpreted_merging(integer) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
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
                    query: r#"CREATE PROCEDURE interpreted_merging(input INT4) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 3;
	INSERT INTO test VALUES (3, input::int4 + 1000);
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'another func')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
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
                    query: "CALL interpreted_merging(55);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function interpreted_merging(integer) does not exist", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE PROCEDURE interpreted_merging(input TEXT) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 2;
	INSERT INTO test VALUES (2, input::int4 + 10000);
END;
$$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'updated table')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_merge('other')::text) = 57;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 57", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_merging('33');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_merging(77);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", INT4), Column("v2", INT4)],
                        rows: &[
                            &[T("1"), T("77")],
                            &[T("3"), T("1077")],
                            &[T("2"), T("10033")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Procedure updates "definition" with custom body"#,
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT);",
                "CREATE PROCEDURE interpreted_example(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('1' || input); END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'initial')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
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
                    query: "CREATE OR REPLACE PROCEDURE interpreted_example(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('3' || input); END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
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
                    query: "CREATE OR REPLACE PROCEDURE interpreted_example(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('2' || input); END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_status;",
                    expected: Expected::Rows {
                        columns: &[Column("table_name", TEXT), Column("staged", BOOL), Column("status", TEXT)],
                        rows: &[
                            &[T("public.interpreted_example(text)"), T("f"), T("modified")],
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
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
                    expected: Expected::Rows {
                        columns: &[Column("length = 32", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT dolt_merge('other');",
                    expected: Expected::Rows {
                        columns: &[Column("dolt_merge", RECORD)],
                        rows: &[
                            &[T(r#"("",0,1,"conflicts found")"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM dolt_conflicts;",
                    expected: Expected::Rows {
                        columns: &[Column("table", TEXT), Column("num_conflicts", NUMERIC)],
                        rows: &[
                            &[T("public.interpreted_example(text)"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Rows {
                        columns: &[Column("base_value", TEXT), Column("our_value", TEXT), Column("our_diff_type", TEXT), Column("their_value", TEXT), Column("their_diff_type", TEXT), Column("dolt_conflict_id", TEXT)],
                        rows: &[
                            &[T("BEGIN INSERT INTO test VALUES ('1' || input); END;"), T("BEGIN INSERT INTO test VALUES ('2' || input); END;"), T("modified"), T("BEGIN INSERT INTO test VALUES ('3' || input); END;"), T("modified"), T("definition")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'BEGIN INSERT INTO test VALUES (''7'' || input); END;' WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "table not found: dolt_conflicts_interpreted_example(text)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL interpreted_example('12');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT * FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("v1", TEXT)],
                        rows: &[
                            &[T("712")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DECLARE variable with default value of literal value or parameter reference",
            set_up_script: &[
                "CREATE TABLE t (a int, b text);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE OR REPLACE PROCEDURE m (x text, y text, e int) LANGUAGE plpgsql AS $$
declare
  xx text := x;
  yy text := y;
  zz int := e;
begin
  insert into t values (zz, xx || yy);
end
$$;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL m('a', 'B', 1);",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("aB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "use nested block statements and call statement in procedure body",
            set_up_script: &[
                "CREATE TABLE tbl (a int, b text);",
                r#"CREATE PROCEDURE add_value(IN a int, IN b text)
            LANGUAGE plpgsql
            AS $$
        BEGIN
            INSERT INTO tbl VALUES (a, b);
        END;
		$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE PROCEDURE check_and_add(IN i int, IN t text)
            LANGUAGE plpgsql
            AS $$
		DECLARE d text := t;
        BEGIN
            IF LENGTH(t) < 6 THEN
                d = t || ' is too short';
            END IF;

            BEGIN
                CALL add_value(i, d);
            END;
        END;
        $$;"#,
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL check_and_add(1, 'hi');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL check_and_add(3, 'hellooo');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM tbl",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("hi is too short")],
                            &[T("3"), T("hellooo")],
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
fn test_procedure_and_do_rules() {
    run_scripts(&[
        ScriptTest {
            name: "procedures and DO",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE p(a int, OUT b int) LANGUAGE plpgsql AS $$ BEGIN b := a + 1; END $$;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE p(a int, OUT b int) LANGUAGE plpgsql AS $$ BEGIN b := a + 1; END $$;",
                    expected: Expected::Error(Diagnostic { code: "42723", message: r#"function "p" already exists with same argument types"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL p(1, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("b", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "CALL",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL p(1);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure p(integer) does not exist", hint: "No procedure matches the given name and argument types. You might need to add explicit type casts.", position: 6, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT p(1);",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "p(integer) is a procedure", hint: "To call a procedure, use CALL.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE q(INOUT x int, y text = 'd') LANGUAGE sql AS $$ SELECT x * 2 $$;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL q(4);",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "CALL",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL q(4, 'z');",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "CALL",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE r() LANGUAGE plpgsql AS $$ BEGIN RAISE NOTICE 'hi'; RETURN; END $$;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL r();",
                    expected: Expected::Tag("CALL"),
                    notices: &[Diagnostic { code: "00000", message: "hi", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE r;",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE nope();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "procedure nope() does not exist", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP FUNCTION p(int, int);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function p(integer, integer) does not exist", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE p(int);",
                    expected: Expected::Tag("DROP PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION fx() RETURNS int LANGUAGE sql AS 'SELECT 1';",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL fx();",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "fx() is not a procedure", hint: "To call a function, use SELECT.", position: 6, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP PROCEDURE fx();",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "fx() is not a procedure", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL abs(1);",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "abs(integer) is not a procedure", hint: "To call a function, use SELECT.", position: 6, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE s(a int) LANGUAGE plpgsql AS $$ BEGIN RETURN 1; END $$;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "RETURN cannot have a parameter in a procedure", position: 63, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE logged (v text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE log_it(v text) LANGUAGE plpgsql AS $$ BEGIN INSERT INTO logged VALUES (v); END $$;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CALL log_it('first');",
                    expected: Expected::Tag("CALL"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DO $$ BEGIN RAISE NOTICE 'x %', 1; INSERT INTO logged VALUES ('from do'); END $$;",
                    expected: Expected::Tag("DO"),
                    notices: &[Diagnostic { code: "00000", message: "x 1", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DO LANGUAGE plpgsql $$ DECLARE n int; BEGIN SELECT count(*) INTO n FROM logged; RAISE NOTICE 'rows %', n; END $$;",
                    expected: Expected::Tag("DO"),
                    notices: &[Diagnostic { code: "00000", message: "rows 2", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "DO LANGUAGE sql $$ SELECT 1 $$;",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"language "sql" does not support inline code execution"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DO LANGUAGE missing_language $$ SELECT 1 $$;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"language "missing_language" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM logged ORDER BY v;",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT)],
                        rows: &[
                            &[T("first")],
                            &[T("from do")],
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
