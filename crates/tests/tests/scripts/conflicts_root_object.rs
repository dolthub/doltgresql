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
fn test_conflicts_root_object() {
    run_scripts(&[
        ScriptTest {
            name: r#"Function delete "definition" conflict without modification"#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("112")],
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("312")],
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                // Changed from the Go test: Postgres cannot change the return type of an existing function, so it is dropped first.
                ScriptTestAssertion {
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT2 AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", INT2)],
                        rows: &[
                            &[T("212")],
                        ],
                        tag: "SELECT 1",
                    },
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
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("BEGIN RETURN '1' || input; END;"), T("BEGIN RETURN '2' || input; END;"), T("modified"), T("BEGIN RETURN '3' || input; END;"), T("modified"), T("definition")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 15, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", INT2)],
                        rows: &[
                            &[T("212")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function update "definition" with custom body"#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("112")],
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("312")],
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("212")],
                        ],
                        tag: "SELECT 1",
                    },
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
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("BEGIN RETURN '1' || input; END;"), T("BEGIN RETURN '2' || input; END;"), T("modified"), T("BEGIN RETURN '3' || input; END;"), T("modified"), T("definition")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'BEGIN RETURN ''7'' || input; END;' WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 15, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
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
            name: r#"Function update "definition" with "theirs" body"#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("112")],
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("312")],
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("212")],
                        ],
                        tag: "SELECT 1",
                    },
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
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("BEGIN RETURN '1' || input; END;"), T("BEGIN RETURN '2' || input; END;"), T("modified"), T("BEGIN RETURN '3' || input; END;"), T("modified"), T("definition")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = their_value WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 13, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 15, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("312")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function update "definition" with "ancestor" body"#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("112")],
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("312")],
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("212")],
                        ],
                        tag: "SELECT 1",
                    },
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
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("BEGIN RETURN '1' || input; END;"), T("BEGIN RETURN '2' || input; END;"), T("modified"), T("BEGIN RETURN '3' || input; END;"), T("modified"), T("definition")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = base_value WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'definition';"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 13, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 15, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("112")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function update "return_type" with custom type"#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT4 AS $$ BEGIN RETURN input || ''; END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", INT4)],
                        rows: &[
                            &[T("12")],
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
                // Changed from the Go test: Postgres cannot change the return type of an existing function, so it is dropped first.
                ScriptTestAssertion {
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT8 AS $$ BEGIN RETURN input || ''; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", INT8)],
                        rows: &[
                            &[T("12")],
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                // Changed from the Go test: Postgres cannot change the return type of an existing function, so it is dropped first.
                ScriptTestAssertion {
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION interpreted_example(input TEXT) RETURNS FLOAT AS $$ BEGIN RETURN input || ''; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", FLOAT8)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
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
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' column naming.
                ScriptTestAssertion {
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("int4"), T("float8"), T("modified"), T("int8"), T("modified"), T("return_type")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'int2' WHERE dolt_conflict_id = 'return_type';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Rows {
                        columns: &[Column("base_value", TEXT), Column("our_value", TEXT), Column("our_diff_type", TEXT), Column("their_value", TEXT), Column("their_diff_type", TEXT), Column("dolt_conflict_id", TEXT)],
                        rows: &[
                            &[T("int4"), T("int2"), T("modified"), T("int8"), T("modified"), T("return_type")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'return_type';"#,
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 15, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", INT2)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('123456');",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#"value "123456" is out of range for type smallint"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function deleted "ours" updated "theirs", chose "ours""#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("ancestor"), Null, T("deleted"), T("theirs"), T("modified"), T("root_object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'ours';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 13, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function interpreted_example(unknown) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function deleted "ours" updated "theirs", chose "theirs""#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("ancestor"), Null, T("deleted"), T("theirs"), T("modified"), T("root_object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'theirs';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 13, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("312")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function deleted "ours" updated "theirs", chose "ancestor""#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("ancestor"), Null, T("deleted"), T("theirs"), T("modified"), T("root_object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'ancestor';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 13, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("112")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function deleted "theirs" updated "ours", chose "ours""#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("ancestor"), T("ours"), T("modified"), Null, T("deleted"), T("root_object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'ours';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("212")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function deleted "theirs" updated "ours", chose "theirs""#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("ancestor"), T("ours"), T("modified"), Null, T("deleted"), T("root_object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'theirs';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 13, ..E }),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function interpreted_example(unknown) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"Function deleted "theirs" updated "ours", chose "ancestor""#,
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                    query: "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("ancestor"), T("ours"), T("modified"), Null, T("deleted"), T("root_object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'ancestor';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "dolt_conflicts_interpreted_example(text)" does not exist"#, position: 13, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", TEXT)],
                        rows: &[
                            &[T("112")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Function update multiple conflicts",
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT4 AS $$ BEGIN RETURN input || '1'; END; $$ LANGUAGE plpgsql;",
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
                // Changed from the Go test: Postgres cannot change the return type of an existing function, so it is dropped first.
                ScriptTestAssertion {
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT8 AS $$ BEGIN RETURN input || '3'; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                // Changed from the Go test: Postgres cannot change the return type of an existing function, so it is dropped first.
                ScriptTestAssertion {
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION interpreted_example(input TEXT) RETURNS FLOAT AS $$ BEGIN RETURN input || '2'; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("public.interpreted_example(text)"), T("2")],
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
                            &[T("int4"), T("float8"), T("modified"), T("int8"), T("modified"), T("return_type")],
                            &[T("BEGIN RETURN input || '1'; END;"), T("BEGIN RETURN input || '2'; END;"), T("modified"), T("BEGIN RETURN input || '3'; END;"), T("modified"), T("definition")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = their_value;"#,
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "relation \"dolt_conflicts_interpreted_example(text)\" does not exist", position: 15, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", INT8)],
                        rows: &[
                            &[T("123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('123456789012');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", INT8)],
                        rows: &[
                            &[T("1234567890123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Function delete multiple conflicts",
            set_up_script: &[
                "CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT4 AS $$ BEGIN RETURN input || '1'; END; $$ LANGUAGE plpgsql;",
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
                // Changed from the Go test: Postgres cannot change the return type of an existing function, so it is dropped first.
                ScriptTestAssertion {
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT8 AS $$ BEGIN RETURN input || '3'; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'other')::text) = 32;",
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
                // Changed from the Go test: Postgres cannot change the return type of an existing function, so it is dropped first.
                ScriptTestAssertion {
                    query: "DROP FUNCTION interpreted_example(input TEXT);",
                    expected: Expected::Tag("DROP FUNCTION"),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "CREATE FUNCTION interpreted_example(input TEXT) RETURNS FLOAT AS $$ BEGIN RETURN input || '2'; END; $$ LANGUAGE plpgsql;",
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
                    query: "SELECT length(dolt_commit('-m', 'next')::text) = 32;",
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
                            &[T("public.interpreted_example(text)"), T("2")],
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
                            &[T("int4"), T("float8"), T("modified"), T("int8"), T("modified"), T("return_type")],
                            &[T("BEGIN RETURN input || '1'; END;"), T("BEGIN RETURN input || '2'; END;"), T("modified"), T("BEGIN RETURN input || '3'; END;"), T("modified"), T("definition")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: r#"DELETE FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Tag("DELETE 2"),
                    ..A
                },
                // Doltgres-specific: Postgres cannot run this, so this expectation follows Postgres' wording for the error.
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "dolt_conflicts_interpreted_example(text)";"#,
                    expected: Expected::Error(Diagnostic { code: "42P01", message: "relation \"dolt_conflicts_interpreted_example(text)\" does not exist", position: 15, ..E }),
                    ..A
                },
                // Doltgres-specific: an earlier Dolt statement changed state Postgres lacks, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT interpreted_example('12');",
                    expected: Expected::Rows {
                        columns: &[Column("interpreted_example", FLOAT8)],
                        rows: &[
                            &[T("122")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
