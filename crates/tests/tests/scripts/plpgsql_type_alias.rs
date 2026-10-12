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
fn test_plpgsql_declare_type_aliases() {
    run_scripts(&[
        // Changed from the Go test: Postgres rejects many schema-qualified type aliases when the function is created, so every setup statement is asserted.
        ScriptTest {
            name: "schema-qualified numeric type aliases",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_boolean() RETURNS text AS $$ DECLARE b pg_catalog.boolean := true; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.boolean" does not exist"#, position: 62, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_int() RETURNS text AS $$ DECLARE b pg_catalog.int := 11; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.int" does not exist"#, position: 58, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_integer() RETURNS text AS $$ DECLARE b pg_catalog.integer := 12; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.integer" does not exist"#, position: 62, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_bigint() RETURNS text AS $$ DECLARE b pg_catalog.bigint := 9223372036854775807; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.bigint" does not exist"#, position: 61, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_smallint() RETURNS text AS $$ DECLARE b pg_catalog.smallint := 32767; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.smallint" does not exist"#, position: 63, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_decimal() RETURNS text AS $$ DECLARE b pg_catalog.decimal := '1.25'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.decimal" does not exist"#, position: 62, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_dec() RETURNS text AS $$ DECLARE b pg_catalog.dec := '2.50'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.dec" does not exist"#, position: 58, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_real() RETURNS text AS $$ DECLARE b pg_catalog.real := '1.5'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.real" does not exist"#, position: 59, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_double_precision() RETURNS text AS $$ DECLARE b pg_catalog.double precision := '2.5'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "precision""#, position: 89, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_float() RETURNS text AS $$ DECLARE b pg_catalog.float := '3.5'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.float" does not exist"#, position: 60, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_boolean();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_boolean() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_int();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_int() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_integer();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_integer() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_bigint();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_bigint() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_smallint();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_smallint() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_decimal();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_decimal() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_dec();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_dec() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_real();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_real() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_double_precision();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_double_precision() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_float();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_float() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: Postgres rejects many schema-qualified type aliases when the function is created, so every setup statement is asserted.
        ScriptTest {
            name: "schema-qualified character type aliases",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_character_varying() RETURNS text AS $$ DECLARE b pg_catalog.character varying := 'abc'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "varying""#, position: 93, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_char_varying() RETURNS text AS $$ DECLARE b pg_catalog.char varying := 'def'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "varying""#, position: 83, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_national_character_varying() RETURNS text AS $$ DECLARE b pg_catalog.national character varying := 'ghi'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "character""#, position: 101, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_national_char_varying() RETURNS text AS $$ DECLARE b pg_catalog.national char varying := 'jkl'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "char""#, position: 96, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_nchar_varying() RETURNS text AS $$ DECLARE b pg_catalog.nchar varying := 'mno'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "varying""#, position: 85, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_character() RETURNS text AS $$ DECLARE b pg_catalog.character := 'pqr'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.character" does not exist"#, position: 64, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_national_character() RETURNS text AS $$ DECLARE b pg_catalog.national character := 'stu'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "character""#, position: 93, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_nchar() RETURNS text AS $$ DECLARE b pg_catalog.nchar := 'vwx'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.nchar" does not exist"#, position: 60, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_character_varying();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_character_varying() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_char_varying();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_char_varying() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_national_character_varying();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_national_character_varying() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_national_char_varying();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_national_char_varying() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_nchar_varying();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_nchar_varying() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_character();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_character() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_national_character();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_national_character() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_nchar();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_nchar() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: Postgres rejects many schema-qualified type aliases when the function is created, so every setup statement is asserted.
        ScriptTest {
            name: "schema-qualified date/time type aliases",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_timestamp_without_tz() RETURNS text AS $$ DECLARE b pg_catalog.timestamp without time zone := '2024-01-02 03:04:05'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "without""#, position: 96, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_timestamp_with_tz() RETURNS text AS $$ DECLARE b pg_catalog.timestamp with time zone := '2024-01-02 03:04:05+00'; BEGIN RETURN pg_typeof(b)::text || '|' || (b AT TIME ZONE 'UTC')::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "with""#, position: 93, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_time_without_tz() RETURNS text AS $$ DECLARE b pg_catalog.time without time zone := '03:04:05'; BEGIN RETURN b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "without""#, position: 86, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_time_with_tz() RETURNS text AS $$ DECLARE b pg_catalog.time with time zone := '03:04:05+00'; BEGIN RETURN b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "with""#, position: 83, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_timestamp_without_tz();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_timestamp_without_tz() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_timestamp_with_tz();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_timestamp_with_tz() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_time_without_tz();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_time_without_tz() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_time_with_tz();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_time_with_tz() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: Postgres rejects many schema-qualified type aliases when the function is created, so every setup statement is asserted.
        ScriptTest {
            name: "schema-qualified bit string type alias",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_bit_varying() RETURNS text AS $$ DECLARE b pg_catalog.bit varying; BEGIN RETURN (b IS NULL)::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "varying""#, position: 81, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_bit_varying();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_bit_varying() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: Postgres rejects many schema-qualified type aliases when the function is created, so every setup statement is asserted.
        ScriptTest {
            name: "multi-word type aliases with irregular whitespace",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION alias_double_precision_spaces() RETURNS text AS $$ DECLARE b pg_catalog.double   precision := '4.5'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "precision""#, position: 98, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION alias_timestamp_tz_newline() RETURNS text AS $$ DECLARE b pg_catalog.timestamp
with	time  zone := '2024-01-02 03:04:05+00'; BEGIN RETURN pg_typeof(b)::text || '|' || (b AT TIME ZONE 'UTC')::text; END; $$ LANGUAGE plpgsql;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "with""#, position: 96, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_double_precision_spaces();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_double_precision_spaces() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT alias_timestamp_tz_newline();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function alias_timestamp_tz_newline() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        // Changed from the Go test: Postgres rejects many schema-qualified type aliases when the function is created, so every setup statement is asserted.
        ScriptTest {
            name: "names whose registered spelling differs",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION canonical_bytea() RETURNS text AS $$ DECLARE b pg_catalog.bytea := '\x0102'; BEGIN RETURN b::text; END; $$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION unqualified_bytea() RETURNS text AS $$ DECLARE b bytea := '\x0304'; BEGIN RETURN b::text; END; $$ LANGUAGE plpgsql;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION canonical_bpchar() RETURNS text AS $$ DECLARE b pg_catalog.bpchar := 'yz'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION unqualified_character() RETURNS text AS $$ DECLARE b character(3) := 'abc'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION canonical_qchar() RETURNS text AS $$ DECLARE b pg_catalog.char; BEGIN RETURN (b IS NULL)::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION canonical_integer_array() RETURNS text AS $$ DECLARE b pg_catalog.integer[] := '{1,2,3}'; BEGIN RETURN pg_typeof(b)::text || '|' || b::text; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.integer[]" does not exist"#, position: 72, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT canonical_bytea();",
                    expected: Expected::Rows {
                        columns: &[Column("canonical_bytea", TEXT)],
                        rows: &[
                            &[T(r#"\x0102"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unqualified_bytea();",
                    expected: Expected::Rows {
                        columns: &[Column("unqualified_bytea", TEXT)],
                        rows: &[
                            &[T(r#"\x0304"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT canonical_bpchar();",
                    expected: Expected::Rows {
                        columns: &[Column("canonical_bpchar", TEXT)],
                        rows: &[
                            &[T("character|yz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unqualified_character();",
                    expected: Expected::Rows {
                        columns: &[Column("unqualified_character", TEXT)],
                        rows: &[
                            &[T("character|abc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT canonical_qchar();",
                    expected: Expected::Rows {
                        columns: &[Column("canonical_qchar", TEXT)],
                        rows: &[
                            &[T("true")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT canonical_integer_array();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function canonical_integer_array() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_plpgsql_declare_unknown_type() {
    run_scripts(&[
        // Changed from the Go test: Postgres rejects the unknown type when the function is created, so the creations are asserted.
        ScriptTest {
            name: "unknown schema-qualified type names are still rejected",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE FUNCTION unknown_type() RETURNS text AS $$ DECLARE b pg_catalog.not_a_type; BEGIN RETURN 'x'; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "pg_catalog.not_a_type" does not exist"#, position: 61, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION unknown_multiword_type() RETURNS text AS $$ DECLARE b pg_catalog.double imprecision; BEGIN RETURN 'x'; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "imprecision""#, position: 89, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unknown_type();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function unknown_type() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unknown_multiword_type();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function unknown_multiword_type() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
