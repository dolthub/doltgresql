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
fn test_alter_statements() {
    run_scripts(&[
        ScriptTest {
            name: "alter database",
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER DATABASE postgres OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter sequence",
            set_up_script: &[
                "CREATE SEQUENCE testseq",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE testseq OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter type",
            set_up_script: &[
                "CREATE TYPE testtype AS ENUM ('a', 'b', 'c')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER TYPE testtype OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter function",
            set_up_script: &[
                "CREATE FUNCTION testfunc() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER FUNCTION testfunc() OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter procedure",
            set_up_script: &[
                "CREATE TABLE test (v1 INT8);",
                "CREATE PROCEDURE testproc() AS $$ BEGIN INSERT INTO test VALUES (1); END; $$ LANGUAGE plpgsql;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER PROCEDURE testproc() OWNER TO foo;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter schema",
            set_up_script: &[
                "CREATE SCHEMA testschema",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER schema testschema OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "alter view",
            set_up_script: &[
                "CREATE VIEW testview AS SELECT 1",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER VIEW testview OWNER TO foo",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "foo" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "OWNER TO checks the role before the object",
            assertions: &[
                ScriptTestAssertion {
                    query: "ALTER DATABASE alt_nope OWNER TO alt_nobody;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "alt_nobody" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TYPE alt_nope OWNER TO alt_nobody;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "alt_nobody" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SCHEMA alt_nope OWNER TO alt_nobody;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "alt_nobody" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER VIEW alt_nope OWNER TO alt_nobody;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "alt_nope" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER FUNCTION alt_nope() OWNER TO alt_nobody;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "alt_nobody" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SCHEMA alt_nope OWNER TO postgres;",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "alt_nope" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TYPE alt_nope OWNER TO postgres;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "alt_nope" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_owner_to_relations() {
    run_scripts(&[
        ScriptTest {
            name: "owner to on sequences, views, and predefined roles",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE ROLE r1;",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SEQUENCE s1;",
                    expected: Expected::Tag("CREATE SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t1 (a int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE VIEW v1 AS SELECT * FROM t1;",
                    expected: Expected::Tag("CREATE VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE s1 OWNER TO r1;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE v1 OWNER TO r1;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE public.s1 OWNER TO postgres;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SCHEMA public OWNER TO pg_database_owner;",
                    expected: Expected::Tag("ALTER SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    notices: &[Diagnostic { code: "42710", message: r#"extension "plpgsql" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE EXTENSION plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42710", message: r#"extension "plpgsql" already exists"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE s1 OWNER TO r1;",
                    expected: Expected::Tag("ALTER SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER VIEW v1 OWNER TO r1;",
                    expected: Expected::Tag("ALTER VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER SEQUENCE v1 OWNER TO r1;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#""v1" is not a sequence"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER VIEW s1 OWNER TO r1;",
                    expected: Expected::Error(Diagnostic { code: "42809", message: r#""s1" is not a view"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET statement_timeout = -1;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"-1 ms is outside the valid range for parameter "statement_timeout" (0 .. 2147483647)"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lock_timeout = '-5s';",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"-5000 ms is outside the valid range for parameter "lock_timeout" (0 .. 2147483647)"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET work_mem = 10;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"10 kB is outside the valid range for parameter "work_mem" (64 .. 2147483647)"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP VIEW v1;",
                    expected: Expected::Tag("DROP VIEW"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SEQUENCE s1;",
                    expected: Expected::Tag("DROP SEQUENCE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP TABLE t1;",
                    expected: Expected::Tag("DROP TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE r1;",
                    expected: Expected::Tag("DROP ROLE"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
