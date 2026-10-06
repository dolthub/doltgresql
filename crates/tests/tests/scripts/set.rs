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
fn test_set_statements() {
    run_scripts(&[
        ScriptTest {
            name: "special case for TIME ZONE",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW timezone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone TO '+00:00';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW timezone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("+00:00")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TIME ZONE LOCAL;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW timezone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TIME ZONE '+00:00';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW timezone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("+00:00")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TIME ZONE '00:00:00';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW timezone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("00:00:00")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TIME ZONE DEFAULT;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW timezone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('timezone')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "special case for SCHEMA",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SCHEMA 'postgres';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = public, pg_catalog;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("public, pg_catalog")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = postgres;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('search_path')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "special case for NAMES",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW client_encoding",
                    expected: Expected::Rows {
                        columns: &[Column("client_encoding", TEXT)],
                        rows: &[
                            &[T("UTF8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET NAMES 'LATIN1';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW client_encoding;",
                    expected: Expected::Rows {
                        columns: &[Column("client_encoding", TEXT)],
                        rows: &[
                            &[T("LATIN1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET client_encoding = DEFAULT;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW client_encoding;",
                    expected: Expected::Rows {
                        columns: &[Column("client_encoding", TEXT)],
                        rows: &[
                            &[T("UTF8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('client_encoding')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("UTF8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "special case SEED",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW geqo_seed",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_seed", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEED 1;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "1""#, position: 10, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_seed",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_seed", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo_seed')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'allow_in_place_tablespaces' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW allow_in_place_tablespaces",
                    expected: Expected::Rows {
                        columns: &[Column("allow_in_place_tablespaces", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET allow_in_place_tablespaces TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW allow_in_place_tablespaces",
                    expected: Expected::Rows {
                        columns: &[Column("allow_in_place_tablespaces", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET allow_in_place_tablespaces TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW allow_in_place_tablespaces",
                    expected: Expected::Rows {
                        columns: &[Column("allow_in_place_tablespaces", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('allow_in_place_tablespaces')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'allow_system_table_mods' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW allow_system_table_mods",
                    expected: Expected::Rows {
                        columns: &[Column("allow_system_table_mods", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET allow_system_table_mods TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW allow_system_table_mods",
                    expected: Expected::Rows {
                        columns: &[Column("allow_system_table_mods", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET allow_system_table_mods TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW allow_system_table_mods",
                    expected: Expected::Rows {
                        columns: &[Column("allow_system_table_mods", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('allow_system_table_mods')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'application_name' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW application_name",
                    expected: Expected::Rows {
                        columns: &[Column("application_name", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET application_name TO 'postgresql'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW application_name",
                    expected: Expected::Rows {
                        columns: &[Column("application_name", TEXT)],
                        rows: &[
                            &[T("postgresql")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET application_name TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW application_name",
                    expected: Expected::Rows {
                        columns: &[Column("application_name", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('application_name')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'archive_cleanup_command' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW archive_cleanup_command",
                    expected: Expected::Rows {
                        columns: &[Column("archive_cleanup_command", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET archive_cleanup_command TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "archive_cleanup_command" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('archive_cleanup_command')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'archive_command' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW archive_command",
                    expected: Expected::Rows {
                        columns: &[Column("archive_command", TEXT)],
                        rows: &[
                            &[T("(disabled)")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET archive_command TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "archive_command" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('archive_command')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("(disabled)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'archive_library' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW archive_library",
                    expected: Expected::Rows {
                        columns: &[Column("archive_library", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET archive_library TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "archive_library" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('archive_library')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'archive_mode' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW archive_mode",
                    expected: Expected::Rows {
                        columns: &[Column("archive_mode", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET archive_mode TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "archive_mode" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('archive_mode')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'archive_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW archive_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("archive_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET archive_timeout TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "archive_timeout" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('archive_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'array_nulls' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW array_nulls",
                    expected: Expected::Rows {
                        columns: &[Column("array_nulls", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET array_nulls TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW array_nulls",
                    expected: Expected::Rows {
                        columns: &[Column("array_nulls", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET array_nulls TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW array_nulls",
                    expected: Expected::Rows {
                        columns: &[Column("array_nulls", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('array_nulls')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'authentication_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW authentication_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("authentication_timeout", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET authentication_timeout TO '120'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "authentication_timeout" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('authentication_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_analyze_scale_factor' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_analyze_scale_factor",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_analyze_scale_factor", TEXT)],
                        rows: &[
                            &[T("0.1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_analyze_scale_factor TO '0.1'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_analyze_scale_factor" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_analyze_scale_factor')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_analyze_threshold' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_analyze_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_analyze_threshold", TEXT)],
                        rows: &[
                            &[T("50")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_analyze_threshold TO '50'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_analyze_threshold" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_analyze_threshold')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_freeze_max_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_freeze_max_age",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_freeze_max_age", TEXT)],
                        rows: &[
                            &[T("200000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_freeze_max_age TO '200000000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_freeze_max_age" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_freeze_max_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("200000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_max_workers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_max_workers",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_max_workers", TEXT)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_max_workers TO '3'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_max_workers" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_max_workers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'autovacuum_multixact_freeze_max_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_multixact_freeze_max_age",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_multixact_freeze_max_age", TEXT)],
                        rows: &[
                            &[T("400000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_multixact_freeze_max_age TO '400000000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_multixact_freeze_max_age" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_multixact_freeze_max_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("400000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_naptime' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_naptime",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_naptime", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_naptime TO '60'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_naptime" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_naptime')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_vacuum_cost_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_vacuum_cost_delay",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_vacuum_cost_delay", TEXT)],
                        rows: &[
                            &[T("2ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_vacuum_cost_delay TO '2'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_vacuum_cost_delay" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_vacuum_cost_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("2ms")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_vacuum_cost_limit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_vacuum_cost_limit",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_vacuum_cost_limit", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_vacuum_cost_limit TO '-1'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_vacuum_cost_limit" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_vacuum_cost_limit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_vacuum_insert_scale_factor' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_vacuum_insert_scale_factor",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_vacuum_insert_scale_factor", TEXT)],
                        rows: &[
                            &[T("0.2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_vacuum_insert_scale_factor TO '0.2'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_vacuum_insert_scale_factor" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_vacuum_insert_scale_factor')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_vacuum_insert_threshold' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_vacuum_insert_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_vacuum_insert_threshold", TEXT)],
                        rows: &[
                            &[T("1000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_vacuum_insert_threshold TO '1000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_vacuum_insert_threshold" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_vacuum_insert_threshold')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'autovacuum_vacuum_scale_factor' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_vacuum_scale_factor",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_vacuum_scale_factor", TEXT)],
                        rows: &[
                            &[T("0.2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_vacuum_scale_factor TO '0.2'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_vacuum_scale_factor" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_vacuum_scale_factor')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_vacuum_threshold' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_vacuum_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_vacuum_threshold", TEXT)],
                        rows: &[
                            &[T("50")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_vacuum_threshold TO '50'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_vacuum_threshold" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_vacuum_threshold')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'autovacuum_work_mem' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW autovacuum_work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("autovacuum_work_mem", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET autovacuum_work_mem TO '-1'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "autovacuum_work_mem" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('autovacuum_work_mem')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'backend_flush_after' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW backend_flush_after",
                    expected: Expected::Rows {
                        columns: &[Column("backend_flush_after", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET backend_flush_after TO '256'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW backend_flush_after",
                    expected: Expected::Rows {
                        columns: &[Column("backend_flush_after", TEXT)],
                        rows: &[
                            &[T("2MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET backend_flush_after TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW backend_flush_after",
                    expected: Expected::Rows {
                        columns: &[Column("backend_flush_after", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('backend_flush_after')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'backslash_quote' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW backslash_quote",
                    expected: Expected::Rows {
                        columns: &[Column("backslash_quote", TEXT)],
                        rows: &[
                            &[T("safe_encoding")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET backslash_quote TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW backslash_quote",
                    expected: Expected::Rows {
                        columns: &[Column("backslash_quote", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET backslash_quote TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW backslash_quote",
                    expected: Expected::Rows {
                        columns: &[Column("backslash_quote", TEXT)],
                        rows: &[
                            &[T("safe_encoding")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('backslash_quote')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("safe_encoding")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'backtrace_functions' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW backtrace_functions",
                    expected: Expected::Rows {
                        columns: &[Column("backtrace_functions", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET backtrace_functions TO 'default'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW backtrace_functions",
                    expected: Expected::Rows {
                        columns: &[Column("backtrace_functions", TEXT)],
                        rows: &[
                            &[T("default")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET backtrace_functions TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW backtrace_functions",
                    expected: Expected::Rows {
                        columns: &[Column("backtrace_functions", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('backtrace_functions')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'bgwriter_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW bgwriter_delay",
                    expected: Expected::Rows {
                        columns: &[Column("bgwriter_delay", TEXT)],
                        rows: &[
                            &[T("200ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bgwriter_delay TO '200'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "bgwriter_delay" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('bgwriter_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("200ms")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'bgwriter_flush_after' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW bgwriter_flush_after",
                    expected: Expected::Rows {
                        columns: &[Column("bgwriter_flush_after", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bgwriter_flush_after TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "bgwriter_flush_after" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('bgwriter_flush_after')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'bgwriter_lru_maxpages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW bgwriter_lru_maxpages",
                    expected: Expected::Rows {
                        columns: &[Column("bgwriter_lru_maxpages", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bgwriter_lru_maxpages TO '100'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "bgwriter_lru_maxpages" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('bgwriter_lru_maxpages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'bgwriter_lru_multiplier' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW bgwriter_lru_multiplier",
                    expected: Expected::Rows {
                        columns: &[Column("bgwriter_lru_multiplier", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bgwriter_lru_multiplier TO '2'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "bgwriter_lru_multiplier" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('bgwriter_lru_multiplier')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'block_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW block_size",
                    expected: Expected::Rows {
                        columns: &[Column("block_size", TEXT)],
                        rows: &[
                            &[T("8192")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET block_size TO '8192'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "block_size" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('block_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("8192")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'bonjour' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW bonjour",
                    expected: Expected::Rows {
                        columns: &[Column("bonjour", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bonjour TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "bonjour" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('bonjour')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'bonjour_name' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW bonjour_name",
                    expected: Expected::Rows {
                        columns: &[Column("bonjour_name", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bonjour_name TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "bonjour_name" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('bonjour_name')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'bytea_output' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW bytea_output",
                    expected: Expected::Rows {
                        columns: &[Column("bytea_output", TEXT)],
                        rows: &[
                            &[T("hex")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bytea_output TO 'escape'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW bytea_output",
                    expected: Expected::Rows {
                        columns: &[Column("bytea_output", TEXT)],
                        rows: &[
                            &[T("escape")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET bytea_output TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW bytea_output",
                    expected: Expected::Rows {
                        columns: &[Column("bytea_output", TEXT)],
                        rows: &[
                            &[T("hex")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('bytea_output')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("hex")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'check_function_bodies' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW check_function_bodies",
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
                    query: "SET check_function_bodies TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW check_function_bodies",
                    expected: Expected::Rows {
                        columns: &[Column("check_function_bodies", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET check_function_bodies TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW check_function_bodies",
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
                    query: "SELECT current_setting('check_function_bodies')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'checkpoint_completion_target' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW checkpoint_completion_target",
                    expected: Expected::Rows {
                        columns: &[Column("checkpoint_completion_target", TEXT)],
                        rows: &[
                            &[T("0.9")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET checkpoint_completion_target TO '0.9'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "checkpoint_completion_target" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('checkpoint_completion_target')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'checkpoint_flush_after' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW checkpoint_flush_after",
                    expected: Expected::Rows {
                        columns: &[Column("checkpoint_flush_after", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET checkpoint_flush_after TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "checkpoint_flush_after" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('checkpoint_flush_after')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'checkpoint_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW checkpoint_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("checkpoint_timeout", TEXT)],
                        rows: &[
                            &[T("5min")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET checkpoint_timeout TO '300'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "checkpoint_timeout" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('checkpoint_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("5min")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'checkpoint_warning' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW checkpoint_warning",
                    expected: Expected::Rows {
                        columns: &[Column("checkpoint_warning", TEXT)],
                        rows: &[
                            &[T("30s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET checkpoint_warning TO '30'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "checkpoint_warning" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('checkpoint_warning')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("30s")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'client_connection_check_interval' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW client_connection_check_interval",
                    expected: Expected::Rows {
                        columns: &[Column("client_connection_check_interval", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET client_connection_check_interval TO 10",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW client_connection_check_interval",
                    expected: Expected::Rows {
                        columns: &[Column("client_connection_check_interval", TEXT)],
                        rows: &[
                            &[T("10ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET client_connection_check_interval TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW client_connection_check_interval",
                    expected: Expected::Rows {
                        columns: &[Column("client_connection_check_interval", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('client_connection_check_interval')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'client_encoding' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW client_encoding",
                    expected: Expected::Rows {
                        columns: &[Column("client_encoding", TEXT)],
                        rows: &[
                            &[T("UTF8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET client_encoding TO 'LATIN1'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW client_encoding",
                    expected: Expected::Rows {
                        columns: &[Column("client_encoding", TEXT)],
                        rows: &[
                            &[T("LATIN1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET client_encoding TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW client_encoding",
                    expected: Expected::Rows {
                        columns: &[Column("client_encoding", TEXT)],
                        rows: &[
                            &[T("UTF8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('client_encoding')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("UTF8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'client_min_messages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW client_min_messages",
                    expected: Expected::Rows {
                        columns: &[Column("client_min_messages", TEXT)],
                        rows: &[
                            &[T("notice")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET client_min_messages TO 'log'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW client_min_messages",
                    expected: Expected::Rows {
                        columns: &[Column("client_min_messages", TEXT)],
                        rows: &[
                            &[T("log")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET client_min_messages TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW client_min_messages",
                    expected: Expected::Rows {
                        columns: &[Column("client_min_messages", TEXT)],
                        rows: &[
                            &[T("notice")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('client_min_messages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("notice")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'cluster_name' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW cluster_name",
                    expected: Expected::Rows {
                        columns: &[Column("cluster_name", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cluster_name TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "cluster_name" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('cluster_name')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'commit_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW commit_delay",
                    expected: Expected::Rows {
                        columns: &[Column("commit_delay", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET commit_delay TO 100000",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW commit_delay",
                    expected: Expected::Rows {
                        columns: &[Column("commit_delay", TEXT)],
                        rows: &[
                            &[T("100000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET commit_delay TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW commit_delay",
                    expected: Expected::Rows {
                        columns: &[Column("commit_delay", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('commit_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'commit_siblings' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW commit_siblings",
                    expected: Expected::Rows {
                        columns: &[Column("commit_siblings", TEXT)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET commit_siblings TO '1000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW commit_siblings",
                    expected: Expected::Rows {
                        columns: &[Column("commit_siblings", TEXT)],
                        rows: &[
                            &[T("1000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET commit_siblings TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW commit_siblings",
                    expected: Expected::Rows {
                        columns: &[Column("commit_siblings", TEXT)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('commit_siblings')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'compute_query_id' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW compute_query_id",
                    expected: Expected::Rows {
                        columns: &[Column("compute_query_id", TEXT)],
                        rows: &[
                            &[T("auto")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET compute_query_id TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW compute_query_id",
                    expected: Expected::Rows {
                        columns: &[Column("compute_query_id", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET compute_query_id TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW compute_query_id",
                    expected: Expected::Rows {
                        columns: &[Column("compute_query_id", TEXT)],
                        rows: &[
                            &[T("auto")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('compute_query_id')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("auto")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'config_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW config_file",
                    expected: Expected::Rows {
                        columns: &[Column("config_file", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET config_file TO '/Users/postgres/postgresql.conf'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "config_file" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('config_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'constraint_exclusion' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW constraint_exclusion",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_exclusion", TEXT)],
                        rows: &[
                            &[T("partition")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET constraint_exclusion TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW constraint_exclusion",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_exclusion", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET constraint_exclusion TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW constraint_exclusion",
                    expected: Expected::Rows {
                        columns: &[Column("constraint_exclusion", TEXT)],
                        rows: &[
                            &[T("partition")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('constraint_exclusion')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("partition")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'cpu_index_tuple_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW cpu_index_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_index_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.005")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cpu_index_tuple_cost TO '0.01'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW cpu_index_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_index_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.01")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cpu_index_tuple_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW cpu_index_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_index_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.005")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('cpu_index_tuple_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.005")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'cpu_operator_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW cpu_operator_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_operator_cost", TEXT)],
                        rows: &[
                            &[T("0.0025")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cpu_operator_cost TO '0.005'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW cpu_operator_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_operator_cost", TEXT)],
                        rows: &[
                            &[T("0.005")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cpu_operator_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW cpu_operator_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_operator_cost", TEXT)],
                        rows: &[
                            &[T("0.0025")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('cpu_operator_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.0025")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'cpu_tuple_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW cpu_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.01")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cpu_tuple_cost TO '0.02'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW cpu_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.02")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cpu_tuple_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW cpu_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("cpu_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.01")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('cpu_tuple_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'createrole_self_grant' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW createrole_self_grant",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "createrole_self_grant""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET createrole_self_grant TO 'inherit'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "createrole_self_grant""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW createrole_self_grant",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "createrole_self_grant""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET createrole_self_grant TO DEFAULT",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "createrole_self_grant""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW createrole_self_grant",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "createrole_self_grant""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('createrole_self_grant')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "createrole_self_grant""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'cursor_tuple_fraction' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW cursor_tuple_fraction",
                    expected: Expected::Rows {
                        columns: &[Column("cursor_tuple_fraction", TEXT)],
                        rows: &[
                            &[T("0.1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cursor_tuple_fraction TO '0.2'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW cursor_tuple_fraction",
                    expected: Expected::Rows {
                        columns: &[Column("cursor_tuple_fraction", TEXT)],
                        rows: &[
                            &[T("0.2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET cursor_tuple_fraction TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW cursor_tuple_fraction",
                    expected: Expected::Rows {
                        columns: &[Column("cursor_tuple_fraction", TEXT)],
                        rows: &[
                            &[T("0.1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('cursor_tuple_fraction')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'data_checksums' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW data_checksums",
                    expected: Expected::Rows {
                        columns: &[Column("data_checksums", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET data_checksums TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "data_checksums" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('data_checksums')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'data_directory' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW data_directory",
                    expected: Expected::Rows {
                        columns: &[Column("data_directory", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET data_directory TO '/Users/postgres'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "data_directory" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('data_directory')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'data_directory_mode' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW data_directory_mode",
                    expected: Expected::Rows {
                        columns: &[Column("data_directory_mode", TEXT)],
                        rows: &[
                            &[T("0700")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET data_directory_mode TO '448'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "data_directory_mode" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('data_directory_mode')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0700")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'data_sync_retry' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW data_sync_retry",
                    expected: Expected::Rows {
                        columns: &[Column("data_sync_retry", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET data_sync_retry TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "data_sync_retry" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('data_sync_retry')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'DateStyle' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW DateStyle",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("ISO, MDY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET DateStyle TO 'ISO, DMY'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("ISO, DMY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET DateStyle TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("ISO, MDY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('DateStyle')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("ISO, MDY")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'db_user_namespace' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW db_user_namespace",
                    expected: Expected::Rows {
                        columns: &[Column("db_user_namespace", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET db_user_namespace TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "db_user_namespace" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('db_user_namespace')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'deadlock_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW deadlock_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("deadlock_timeout", TEXT)],
                        rows: &[
                            &[T("1s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET deadlock_timeout TO '2000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW deadlock_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("deadlock_timeout", TEXT)],
                        rows: &[
                            &[T("2s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET deadlock_timeout TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW deadlock_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("deadlock_timeout", TEXT)],
                        rows: &[
                            &[T("1s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('deadlock_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1s")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_assertions' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_assertions",
                    expected: Expected::Rows {
                        columns: &[Column("debug_assertions", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_assertions TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "debug_assertions" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_assertions')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_discard_caches' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_discard_caches",
                    expected: Expected::Rows {
                        columns: &[Column("debug_discard_caches", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_discard_caches TO '0'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_discard_caches",
                    expected: Expected::Rows {
                        columns: &[Column("debug_discard_caches", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_discard_caches')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_io_direct' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_io_direct",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_io_direct""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_io_direct TO ''",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_io_direct""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_io_direct')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_io_direct""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_logical_replication_streaming' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_logical_replication_streaming",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_logical_replication_streaming""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_logical_replication_streaming TO 'immediate'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_logical_replication_streaming""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_logical_replication_streaming",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_logical_replication_streaming""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_logical_replication_streaming TO DEFAULT",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_logical_replication_streaming""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_logical_replication_streaming",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_logical_replication_streaming""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_logical_replication_streaming')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_logical_replication_streaming""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_parallel_query' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_parallel_query",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_parallel_query""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_parallel_query TO 'regress'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_parallel_query""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_parallel_query",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_parallel_query""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_parallel_query TO DEFAULT",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_parallel_query""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_parallel_query",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_parallel_query""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_parallel_query')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "debug_parallel_query""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_pretty_print' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_pretty_print",
                    expected: Expected::Rows {
                        columns: &[Column("debug_pretty_print", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_pretty_print TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_pretty_print",
                    expected: Expected::Rows {
                        columns: &[Column("debug_pretty_print", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_pretty_print TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_pretty_print",
                    expected: Expected::Rows {
                        columns: &[Column("debug_pretty_print", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_pretty_print')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_print_parse' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_print_parse",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_parse", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_print_parse TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_print_parse",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_parse", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_print_parse TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_print_parse",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_parse", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_print_parse')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_print_plan' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_print_plan",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_plan", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_print_plan TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_print_plan",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_plan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_print_plan TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_print_plan",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_plan", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_print_plan')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'debug_print_rewritten' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW debug_print_rewritten",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_rewritten", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_print_rewritten TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_print_rewritten",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_rewritten", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET debug_print_rewritten TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW debug_print_rewritten",
                    expected: Expected::Rows {
                        columns: &[Column("debug_print_rewritten", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('debug_print_rewritten')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'default_statistics_target' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW default_statistics_target",
                    expected: Expected::Rows {
                        columns: &[Column("default_statistics_target", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_statistics_target TO '10000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_statistics_target",
                    expected: Expected::Rows {
                        columns: &[Column("default_statistics_target", TEXT)],
                        rows: &[
                            &[T("10000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_statistics_target TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_statistics_target",
                    expected: Expected::Rows {
                        columns: &[Column("default_statistics_target", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('default_statistics_target')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'default_table_access_method' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW default_table_access_method",
                    expected: Expected::Rows {
                        columns: &[Column("default_table_access_method", TEXT)],
                        rows: &[
                            &[T("heap")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_table_access_method TO 'heap'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_table_access_method",
                    expected: Expected::Rows {
                        columns: &[Column("default_table_access_method", TEXT)],
                        rows: &[
                            &[T("heap")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('default_table_access_method')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("heap")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'default_tablespace' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW default_tablespace",
                    expected: Expected::Rows {
                        columns: &[Column("default_tablespace", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_tablespace TO 'pg_default'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_tablespace",
                    expected: Expected::Rows {
                        columns: &[Column("default_tablespace", TEXT)],
                        rows: &[
                            &[T("pg_default")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_tablespace TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_tablespace",
                    expected: Expected::Rows {
                        columns: &[Column("default_tablespace", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('default_tablespace')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'default_text_search_config' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW default_text_search_config",
                    expected: Expected::Rows {
                        columns: &[Column("default_text_search_config", TEXT)],
                        rows: &[
                            &[T("pg_catalog.english")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_text_search_config TO 'pg_catalog.spanish'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_text_search_config",
                    expected: Expected::Rows {
                        columns: &[Column("default_text_search_config", TEXT)],
                        rows: &[
                            &[T("pg_catalog.spanish")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_text_search_config TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_text_search_config",
                    expected: Expected::Rows {
                        columns: &[Column("default_text_search_config", TEXT)],
                        rows: &[
                            &[T("pg_catalog.english")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('default_text_search_config')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("pg_catalog.english")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'default_toast_compression' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW default_toast_compression",
                    expected: Expected::Rows {
                        columns: &[Column("default_toast_compression", TEXT)],
                        rows: &[
                            &[T("pglz")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_toast_compression TO 'lz4'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_toast_compression",
                    expected: Expected::Rows {
                        columns: &[Column("default_toast_compression", TEXT)],
                        rows: &[
                            &[T("lz4")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_toast_compression TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_toast_compression",
                    expected: Expected::Rows {
                        columns: &[Column("default_toast_compression", TEXT)],
                        rows: &[
                            &[T("pglz")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('default_toast_compression')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("pglz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'default_transaction_deferrable' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW default_transaction_deferrable",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_deferrable", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_transaction_deferrable TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_transaction_deferrable",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_deferrable", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_transaction_deferrable TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_transaction_deferrable",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_deferrable", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('default_transaction_deferrable')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'default_transaction_isolation' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW default_transaction_isolation",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_isolation", TEXT)],
                        rows: &[
                            &[T("read committed")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_transaction_isolation TO 'serializable'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_transaction_isolation",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_isolation", TEXT)],
                        rows: &[
                            &[T("serializable")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_transaction_isolation TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_transaction_isolation",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_isolation", TEXT)],
                        rows: &[
                            &[T("read committed")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('default_transaction_isolation')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("read committed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'default_transaction_read_only' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW default_transaction_read_only",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_read_only", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_transaction_read_only TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_transaction_read_only",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_read_only", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_transaction_read_only TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW default_transaction_read_only",
                    expected: Expected::Rows {
                        columns: &[Column("default_transaction_read_only", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('default_transaction_read_only')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'dynamic_library_path' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW dynamic_library_path",
                    expected: Expected::Rows {
                        columns: &[Column("dynamic_library_path", TEXT)],
                        rows: &[
                            &[T("$libdir")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET dynamic_library_path TO ''",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW dynamic_library_path",
                    expected: Expected::Rows {
                        columns: &[Column("dynamic_library_path", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET dynamic_library_path TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW dynamic_library_path",
                    expected: Expected::Rows {
                        columns: &[Column("dynamic_library_path", TEXT)],
                        rows: &[
                            &[T("$libdir")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('dynamic_library_path')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("$libdir")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'dynamic_shared_memory_type' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW dynamic_shared_memory_type",
                    expected: Expected::Rows {
                        columns: &[Column("dynamic_shared_memory_type", TEXT)],
                        rows: &[
                            &[T("posix")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET dynamic_shared_memory_type TO 'posix'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "dynamic_shared_memory_type" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('dynamic_shared_memory_type')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("posix")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'effective_cache_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW effective_cache_size",
                    expected: Expected::Rows {
                        columns: &[Column("effective_cache_size", TEXT)],
                        rows: &[
                            &[T("4GB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET effective_cache_size TO '400000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW effective_cache_size",
                    expected: Expected::Rows {
                        columns: &[Column("effective_cache_size", TEXT)],
                        rows: &[
                            &[T("3125MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET effective_cache_size TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW effective_cache_size",
                    expected: Expected::Rows {
                        columns: &[Column("effective_cache_size", TEXT)],
                        rows: &[
                            &[T("4GB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('effective_cache_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("4GB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'effective_io_concurrency' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW effective_io_concurrency",
                    expected: Expected::Rows {
                        columns: &[Column("effective_io_concurrency", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET effective_io_concurrency TO '100'",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid value for parameter "effective_io_concurrency": 100"#, detail: "effective_io_concurrency must be set to 0 on platforms that lack posix_fadvise().", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW effective_io_concurrency",
                    expected: Expected::Rows {
                        columns: &[Column("effective_io_concurrency", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET effective_io_concurrency TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW effective_io_concurrency",
                    expected: Expected::Rows {
                        columns: &[Column("effective_io_concurrency", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('effective_io_concurrency')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_async_append' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_async_append",
                    expected: Expected::Rows {
                        columns: &[Column("enable_async_append", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_async_append TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_async_append",
                    expected: Expected::Rows {
                        columns: &[Column("enable_async_append", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_async_append TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_async_append",
                    expected: Expected::Rows {
                        columns: &[Column("enable_async_append", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_async_append')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_bitmapscan' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_bitmapscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_bitmapscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_bitmapscan TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_bitmapscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_bitmapscan", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_bitmapscan TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_bitmapscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_bitmapscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_bitmapscan')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_gathermerge' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_gathermerge",
                    expected: Expected::Rows {
                        columns: &[Column("enable_gathermerge", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_gathermerge TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_gathermerge",
                    expected: Expected::Rows {
                        columns: &[Column("enable_gathermerge", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_gathermerge TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_gathermerge",
                    expected: Expected::Rows {
                        columns: &[Column("enable_gathermerge", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_gathermerge')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_hashagg' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_hashagg",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashagg", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashagg TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashagg",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashagg", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashagg TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashagg",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashagg", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_hashagg')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_hashjoin' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_hashjoin TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_hashjoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_hashjoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_hashjoin')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_incremental_sort' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_incremental_sort",
                    expected: Expected::Rows {
                        columns: &[Column("enable_incremental_sort", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_incremental_sort TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_incremental_sort",
                    expected: Expected::Rows {
                        columns: &[Column("enable_incremental_sort", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_incremental_sort TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_incremental_sort",
                    expected: Expected::Rows {
                        columns: &[Column("enable_incremental_sort", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_incremental_sort')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_indexonlyscan' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_indexonlyscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_indexonlyscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_indexonlyscan TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_indexonlyscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_indexonlyscan", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_indexonlyscan TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_indexonlyscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_indexonlyscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_indexonlyscan')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_indexscan' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_indexscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_indexscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_indexscan TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_indexscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_indexscan", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_indexscan TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_indexscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_indexscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_indexscan')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_material' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_material",
                    expected: Expected::Rows {
                        columns: &[Column("enable_material", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_material TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_material",
                    expected: Expected::Rows {
                        columns: &[Column("enable_material", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_material TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_material",
                    expected: Expected::Rows {
                        columns: &[Column("enable_material", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_material')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_memoize' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_memoize",
                    expected: Expected::Rows {
                        columns: &[Column("enable_memoize", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_memoize TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_memoize",
                    expected: Expected::Rows {
                        columns: &[Column("enable_memoize", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_memoize TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_memoize",
                    expected: Expected::Rows {
                        columns: &[Column("enable_memoize", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_memoize')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_mergejoin' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_mergejoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_mergejoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_mergejoin TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_mergejoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_mergejoin", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_mergejoin TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_mergejoin",
                    expected: Expected::Rows {
                        columns: &[Column("enable_mergejoin", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_mergejoin')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_nestloop' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_nestloop",
                    expected: Expected::Rows {
                        columns: &[Column("enable_nestloop", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_nestloop TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_nestloop",
                    expected: Expected::Rows {
                        columns: &[Column("enable_nestloop", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_nestloop TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_nestloop",
                    expected: Expected::Rows {
                        columns: &[Column("enable_nestloop", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_nestloop')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_parallel_append' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_parallel_append",
                    expected: Expected::Rows {
                        columns: &[Column("enable_parallel_append", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_parallel_append TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_parallel_append",
                    expected: Expected::Rows {
                        columns: &[Column("enable_parallel_append", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_parallel_append TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_parallel_append",
                    expected: Expected::Rows {
                        columns: &[Column("enable_parallel_append", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_parallel_append')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_parallel_hash' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_parallel_hash",
                    expected: Expected::Rows {
                        columns: &[Column("enable_parallel_hash", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_parallel_hash TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_parallel_hash",
                    expected: Expected::Rows {
                        columns: &[Column("enable_parallel_hash", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_parallel_hash TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_parallel_hash",
                    expected: Expected::Rows {
                        columns: &[Column("enable_parallel_hash", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_parallel_hash')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_partition_pruning' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_partition_pruning",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partition_pruning", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_partition_pruning TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_partition_pruning",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partition_pruning", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_partition_pruning TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_partition_pruning",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partition_pruning", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_partition_pruning')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_partitionwise_aggregate' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_partitionwise_aggregate",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partitionwise_aggregate", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_partitionwise_aggregate TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_partitionwise_aggregate",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partitionwise_aggregate", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_partitionwise_aggregate TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_partitionwise_aggregate",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partitionwise_aggregate", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_partitionwise_aggregate')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_partitionwise_join' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_partitionwise_join",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partitionwise_join", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_partitionwise_join TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_partitionwise_join",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partitionwise_join", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_partitionwise_join TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_partitionwise_join",
                    expected: Expected::Rows {
                        columns: &[Column("enable_partitionwise_join", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_partitionwise_join')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_presorted_aggregate' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_presorted_aggregate",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "enable_presorted_aggregate""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_presorted_aggregate TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "enable_presorted_aggregate""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_presorted_aggregate",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "enable_presorted_aggregate""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_presorted_aggregate TO DEFAULT",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "enable_presorted_aggregate""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_presorted_aggregate",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "enable_presorted_aggregate""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_presorted_aggregate')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "enable_presorted_aggregate""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_seqscan' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_seqscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_seqscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_seqscan TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_seqscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_seqscan", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_seqscan TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_seqscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_seqscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_seqscan')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_sort' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_sort",
                    expected: Expected::Rows {
                        columns: &[Column("enable_sort", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_sort TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_sort",
                    expected: Expected::Rows {
                        columns: &[Column("enable_sort", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_sort TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_sort",
                    expected: Expected::Rows {
                        columns: &[Column("enable_sort", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_sort')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'enable_tidscan' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW enable_tidscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_tidscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_tidscan TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_tidscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_tidscan", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET enable_tidscan TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW enable_tidscan",
                    expected: Expected::Rows {
                        columns: &[Column("enable_tidscan", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('enable_tidscan')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'escape_string_warning' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW escape_string_warning",
                    expected: Expected::Rows {
                        columns: &[Column("escape_string_warning", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET escape_string_warning TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW escape_string_warning",
                    expected: Expected::Rows {
                        columns: &[Column("escape_string_warning", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET escape_string_warning TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW escape_string_warning",
                    expected: Expected::Rows {
                        columns: &[Column("escape_string_warning", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('escape_string_warning')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'event_source' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW event_source",
                    expected: Expected::Rows {
                        columns: &[Column("event_source", TEXT)],
                        rows: &[
                            &[T("PostgreSQL")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET event_source TO 'PostgreSQL'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "event_source" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('event_source')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("PostgreSQL")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'exit_on_error' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW exit_on_error",
                    expected: Expected::Rows {
                        columns: &[Column("exit_on_error", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET exit_on_error TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW exit_on_error",
                    expected: Expected::Rows {
                        columns: &[Column("exit_on_error", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET exit_on_error TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW exit_on_error",
                    expected: Expected::Rows {
                        columns: &[Column("exit_on_error", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('exit_on_error')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'external_pid_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW external_pid_file",
                    expected: Expected::Rows {
                        columns: &[Column("external_pid_file", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET external_pid_file TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "external_pid_file" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('external_pid_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'extra_float_digits' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW extra_float_digits",
                    expected: Expected::Rows {
                        columns: &[Column("extra_float_digits", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET extra_float_digits TO -10",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW extra_float_digits",
                    expected: Expected::Rows {
                        columns: &[Column("extra_float_digits", TEXT)],
                        rows: &[
                            &[T("-10")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET extra_float_digits TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW extra_float_digits",
                    expected: Expected::Rows {
                        columns: &[Column("extra_float_digits", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('extra_float_digits')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'from_collapse_limit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW from_collapse_limit",
                    expected: Expected::Rows {
                        columns: &[Column("from_collapse_limit", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET from_collapse_limit TO 100",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW from_collapse_limit",
                    expected: Expected::Rows {
                        columns: &[Column("from_collapse_limit", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET from_collapse_limit TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW from_collapse_limit",
                    expected: Expected::Rows {
                        columns: &[Column("from_collapse_limit", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('from_collapse_limit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
        ScriptTest {
            name: "set 'fsync' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW fsync",
                    expected: Expected::Rows {
                        columns: &[Column("fsync", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET fsync TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "fsync" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('fsync')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'full_page_writes' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW full_page_writes",
                    expected: Expected::Rows {
                        columns: &[Column("full_page_writes", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET full_page_writes TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "full_page_writes" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('full_page_writes')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'geqo' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW geqo",
                    expected: Expected::Rows {
                        columns: &[Column("geqo", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo",
                    expected: Expected::Rows {
                        columns: &[Column("geqo", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo",
                    expected: Expected::Rows {
                        columns: &[Column("geqo", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'geqo_effort' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW geqo_effort",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_effort", TEXT)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_effort TO 10",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_effort",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_effort", TEXT)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_effort TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_effort",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_effort", TEXT)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo_effort')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'geqo_generations' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW geqo_generations",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_generations", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_generations TO '100'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_generations",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_generations", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_generations TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_generations",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_generations", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo_generations')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo_generations')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'geqo_pool_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW geqo_pool_size",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_pool_size", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_pool_size TO 1",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_pool_size",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_pool_size", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_pool_size TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_pool_size",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_pool_size", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo_pool_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'geqo_seed' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW geqo_seed",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_seed", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_seed TO 0.2",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_seed",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_seed", TEXT)],
                        rows: &[
                            &[T("0.2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_seed TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_seed",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_seed", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo_seed')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'geqo_selection_bias' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW geqo_selection_bias",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_selection_bias", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_selection_bias TO 1.7",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_selection_bias",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_selection_bias", TEXT)],
                        rows: &[
                            &[T("1.7")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_selection_bias TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_selection_bias",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_selection_bias", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo_selection_bias')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'geqo_threshold' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW geqo_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_threshold", TEXT)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_threshold TO 22",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_threshold", TEXT)],
                        rows: &[
                            &[T("22")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET geqo_threshold TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW geqo_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("geqo_threshold", TEXT)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('geqo_threshold')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'gin_fuzzy_search_limit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW gin_fuzzy_search_limit",
                    expected: Expected::Rows {
                        columns: &[Column("gin_fuzzy_search_limit", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET gin_fuzzy_search_limit TO 2",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW gin_fuzzy_search_limit",
                    expected: Expected::Rows {
                        columns: &[Column("gin_fuzzy_search_limit", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET gin_fuzzy_search_limit TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW gin_fuzzy_search_limit",
                    expected: Expected::Rows {
                        columns: &[Column("gin_fuzzy_search_limit", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('gin_fuzzy_search_limit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'gin_pending_list_limit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW gin_pending_list_limit",
                    expected: Expected::Rows {
                        columns: &[Column("gin_pending_list_limit", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET gin_pending_list_limit TO '4000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW gin_pending_list_limit",
                    expected: Expected::Rows {
                        columns: &[Column("gin_pending_list_limit", TEXT)],
                        rows: &[
                            &[T("4000kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET gin_pending_list_limit TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW gin_pending_list_limit",
                    expected: Expected::Rows {
                        columns: &[Column("gin_pending_list_limit", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('gin_pending_list_limit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'gss_accept_delegation' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW gss_accept_delegation",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "gss_accept_delegation""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET gss_accept_delegation TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "gss_accept_delegation""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('gss_accept_delegation')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "gss_accept_delegation""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'hash_mem_multiplier' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW hash_mem_multiplier",
                    expected: Expected::Rows {
                        columns: &[Column("hash_mem_multiplier", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET hash_mem_multiplier TO 20.1",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW hash_mem_multiplier",
                    expected: Expected::Rows {
                        columns: &[Column("hash_mem_multiplier", TEXT)],
                        rows: &[
                            &[T("20.1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET hash_mem_multiplier TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW hash_mem_multiplier",
                    expected: Expected::Rows {
                        columns: &[Column("hash_mem_multiplier", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('hash_mem_multiplier')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'hba_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW hba_file",
                    expected: Expected::Rows {
                        columns: &[Column("hba_file", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET hba_file TO '/Users/postgres/pg_hba.conf'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "hba_file" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('hba_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'hot_standby' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW hot_standby",
                    expected: Expected::Rows {
                        columns: &[Column("hot_standby", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET hot_standby TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "hot_standby" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('hot_standby')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'hot_standby_feedback' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW hot_standby_feedback",
                    expected: Expected::Rows {
                        columns: &[Column("hot_standby_feedback", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET hot_standby_feedback TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "hot_standby_feedback" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('hot_standby_feedback')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'huge_page_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW huge_page_size",
                    expected: Expected::Rows {
                        columns: &[Column("huge_page_size", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET huge_page_size TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "huge_page_size" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('huge_page_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'huge_pages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW huge_pages",
                    expected: Expected::Rows {
                        columns: &[Column("huge_pages", TEXT)],
                        rows: &[
                            &[T("try")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET huge_pages TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "huge_pages" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('huge_pages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("try")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'icu_validation_level' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW icu_validation_level",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "icu_validation_level""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET icu_validation_level TO 'disabled'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "icu_validation_level""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW icu_validation_level",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "icu_validation_level""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET icu_validation_level TO DEFAULT",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "icu_validation_level""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW icu_validation_level",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "icu_validation_level""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('icu_validation_level')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "icu_validation_level""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ident_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ident_file",
                    expected: Expected::Rows {
                        columns: &[Column("ident_file", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ident_file TO '/Users/postgres/pg_ident.conf'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ident_file" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ident_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'idle_in_transaction_session_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW idle_in_transaction_session_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("idle_in_transaction_session_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET idle_in_transaction_session_timeout TO 2",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW idle_in_transaction_session_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("idle_in_transaction_session_timeout", TEXT)],
                        rows: &[
                            &[T("2ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET idle_in_transaction_session_timeout TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW idle_in_transaction_session_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("idle_in_transaction_session_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('idle_in_transaction_session_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'idle_session_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW idle_session_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("idle_session_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET idle_session_timeout TO '3'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW idle_session_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("idle_session_timeout", TEXT)],
                        rows: &[
                            &[T("3ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET idle_session_timeout TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW idle_session_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("idle_session_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('idle_session_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ignore_checksum_failure' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ignore_checksum_failure",
                    expected: Expected::Rows {
                        columns: &[Column("ignore_checksum_failure", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ignore_checksum_failure TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW ignore_checksum_failure",
                    expected: Expected::Rows {
                        columns: &[Column("ignore_checksum_failure", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ignore_checksum_failure TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW ignore_checksum_failure",
                    expected: Expected::Rows {
                        columns: &[Column("ignore_checksum_failure", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ignore_checksum_failure')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ignore_invalid_pages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ignore_invalid_pages",
                    expected: Expected::Rows {
                        columns: &[Column("ignore_invalid_pages", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ignore_invalid_pages TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ignore_invalid_pages" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ignore_invalid_pages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ignore_system_indexes' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ignore_system_indexes",
                    expected: Expected::Rows {
                        columns: &[Column("ignore_system_indexes", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ignore_system_indexes TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ignore_system_indexes" cannot be set after connection start"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ignore_system_indexes')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'in_hot_standby' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW in_hot_standby",
                    expected: Expected::Rows {
                        columns: &[Column("in_hot_standby", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET in_hot_standby TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "in_hot_standby" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('in_hot_standby')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'integer_datetimes' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW integer_datetimes",
                    expected: Expected::Rows {
                        columns: &[Column("integer_datetimes", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET integer_datetimes TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "integer_datetimes" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('integer_datetimes')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'IntervalStyle' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW IntervalStyle",
                    expected: Expected::Rows {
                        columns: &[Column("IntervalStyle", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET IntervalStyle TO 'sql_standard'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW IntervalStyle",
                    expected: Expected::Rows {
                        columns: &[Column("IntervalStyle", TEXT)],
                        rows: &[
                            &[T("sql_standard")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET IntervalStyle TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW IntervalStyle",
                    expected: Expected::Rows {
                        columns: &[Column("IntervalStyle", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('IntervalStyle')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit",
                    expected: Expected::Rows {
                        columns: &[Column("jit", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit",
                    expected: Expected::Rows {
                        columns: &[Column("jit", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit",
                    expected: Expected::Rows {
                        columns: &[Column("jit", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_above_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_above_cost", TEXT)],
                        rows: &[
                            &[T("100000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_above_cost TO '100'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_above_cost", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_above_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_above_cost", TEXT)],
                        rows: &[
                            &[T("100000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_above_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("100000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_debugging_support' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_debugging_support",
                    expected: Expected::Rows {
                        columns: &[Column("jit_debugging_support", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_debugging_support TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "jit_debugging_support" cannot be set after connection start"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_debugging_support')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_dump_bitcode' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_dump_bitcode",
                    expected: Expected::Rows {
                        columns: &[Column("jit_dump_bitcode", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_dump_bitcode TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_dump_bitcode",
                    expected: Expected::Rows {
                        columns: &[Column("jit_dump_bitcode", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_dump_bitcode TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_dump_bitcode",
                    expected: Expected::Rows {
                        columns: &[Column("jit_dump_bitcode", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_dump_bitcode')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_expressions' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_expressions",
                    expected: Expected::Rows {
                        columns: &[Column("jit_expressions", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_expressions TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_expressions",
                    expected: Expected::Rows {
                        columns: &[Column("jit_expressions", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_expressions TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_expressions",
                    expected: Expected::Rows {
                        columns: &[Column("jit_expressions", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_expressions')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_inline_above_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_inline_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_inline_above_cost", TEXT)],
                        rows: &[
                            &[T("500000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_inline_above_cost TO '5000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_inline_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_inline_above_cost", TEXT)],
                        rows: &[
                            &[T("5000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_inline_above_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_inline_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_inline_above_cost", TEXT)],
                        rows: &[
                            &[T("500000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_inline_above_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("500000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_optimize_above_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_optimize_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_optimize_above_cost", TEXT)],
                        rows: &[
                            &[T("500000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_optimize_above_cost TO '5000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_optimize_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_optimize_above_cost", TEXT)],
                        rows: &[
                            &[T("5000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_optimize_above_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_optimize_above_cost",
                    expected: Expected::Rows {
                        columns: &[Column("jit_optimize_above_cost", TEXT)],
                        rows: &[
                            &[T("500000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_optimize_above_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("500000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_profiling_support' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_profiling_support",
                    expected: Expected::Rows {
                        columns: &[Column("jit_profiling_support", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_profiling_support TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "jit_profiling_support" cannot be set after connection start"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_profiling_support')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_provider' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_provider",
                    expected: Expected::Rows {
                        columns: &[Column("jit_provider", TEXT)],
                        rows: &[
                            &[T("llvmjit")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_provider TO 'llvmjit'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "jit_provider" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_provider')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("llvmjit")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'jit_tuple_deforming' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW jit_tuple_deforming",
                    expected: Expected::Rows {
                        columns: &[Column("jit_tuple_deforming", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_tuple_deforming TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_tuple_deforming",
                    expected: Expected::Rows {
                        columns: &[Column("jit_tuple_deforming", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET jit_tuple_deforming TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW jit_tuple_deforming",
                    expected: Expected::Rows {
                        columns: &[Column("jit_tuple_deforming", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('jit_tuple_deforming')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'join_collapse_limit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW join_collapse_limit",
                    expected: Expected::Rows {
                        columns: &[Column("join_collapse_limit", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET join_collapse_limit TO '100'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW join_collapse_limit",
                    expected: Expected::Rows {
                        columns: &[Column("join_collapse_limit", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET join_collapse_limit TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW join_collapse_limit",
                    expected: Expected::Rows {
                        columns: &[Column("join_collapse_limit", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('join_collapse_limit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
        ScriptTest {
            name: "set 'krb_caseins_users' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW krb_caseins_users",
                    expected: Expected::Rows {
                        columns: &[Column("krb_caseins_users", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET krb_caseins_users TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "krb_caseins_users" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('krb_caseins_users')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'krb_server_keyfile' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW krb_server_keyfile",
                    expected: Expected::Rows {
                        columns: &[Column("krb_server_keyfile", TEXT)],
                        rows: &[
                            &[T("FILE:/opt/homebrew/etc/postgresql/krb5.keytab")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET krb_server_keyfile TO 'FILE:'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "krb_server_keyfile" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('krb_server_keyfile')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("FILE:/opt/homebrew/etc/postgresql/krb5.keytab")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'lc_messages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW lc_messages",
                    expected: Expected::Rows {
                        columns: &[Column("lc_messages", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lc_messages TO 'en_US'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lc_messages",
                    expected: Expected::Rows {
                        columns: &[Column("lc_messages", TEXT)],
                        rows: &[
                            &[T("en_US")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lc_messages TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lc_messages",
                    expected: Expected::Rows {
                        columns: &[Column("lc_messages", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('lc_messages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'lc_monetary' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW lc_monetary",
                    expected: Expected::Rows {
                        columns: &[Column("lc_monetary", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lc_monetary TO 'en_US'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lc_monetary",
                    expected: Expected::Rows {
                        columns: &[Column("lc_monetary", TEXT)],
                        rows: &[
                            &[T("en_US")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lc_monetary TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lc_monetary",
                    expected: Expected::Rows {
                        columns: &[Column("lc_monetary", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('lc_monetary')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'lc_numeric' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW lc_numeric",
                    expected: Expected::Rows {
                        columns: &[Column("lc_numeric", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lc_numeric TO 'en_US'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lc_numeric",
                    expected: Expected::Rows {
                        columns: &[Column("lc_numeric", TEXT)],
                        rows: &[
                            &[T("en_US")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lc_numeric TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lc_numeric",
                    expected: Expected::Rows {
                        columns: &[Column("lc_numeric", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('lc_numeric')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'lc_time' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW lc_time",
                    expected: Expected::Rows {
                        columns: &[Column("lc_time", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lc_time TO 'en_US'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lc_time",
                    expected: Expected::Rows {
                        columns: &[Column("lc_time", TEXT)],
                        rows: &[
                            &[T("en_US")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lc_time TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lc_time",
                    expected: Expected::Rows {
                        columns: &[Column("lc_time", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('lc_time')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("en_US.UTF-8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'listen_addresses' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW listen_addresses",
                    expected: Expected::Rows {
                        columns: &[Column("listen_addresses", TEXT)],
                        rows: &[
                            &[T("127.0.0.1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET listen_addresses TO 'localhost'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "listen_addresses" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('listen_addresses')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("127.0.0.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'lo_compat_privileges' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW lo_compat_privileges",
                    expected: Expected::Rows {
                        columns: &[Column("lo_compat_privileges", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lo_compat_privileges TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lo_compat_privileges",
                    expected: Expected::Rows {
                        columns: &[Column("lo_compat_privileges", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lo_compat_privileges TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lo_compat_privileges",
                    expected: Expected::Rows {
                        columns: &[Column("lo_compat_privileges", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('lo_compat_privileges')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'local_preload_libraries' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW local_preload_libraries",
                    expected: Expected::Rows {
                        columns: &[Column("local_preload_libraries", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET local_preload_libraries TO '/'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW local_preload_libraries",
                    expected: Expected::Rows {
                        columns: &[Column("local_preload_libraries", TEXT)],
                        rows: &[
                            &[T(r#""/""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET local_preload_libraries TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW local_preload_libraries",
                    expected: Expected::Rows {
                        columns: &[Column("local_preload_libraries", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('local_preload_libraries')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'lock_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW lock_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("lock_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lock_timeout TO 20",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lock_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("lock_timeout", TEXT)],
                        rows: &[
                            &[T("20ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET lock_timeout TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW lock_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("lock_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('lock_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_autovacuum_min_duration' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_autovacuum_min_duration",
                    expected: Expected::Rows {
                        columns: &[Column("log_autovacuum_min_duration", TEXT)],
                        rows: &[
                            &[T("10min")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_autovacuum_min_duration TO '600'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_autovacuum_min_duration" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_autovacuum_min_duration')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("10min")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_checkpoints' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_checkpoints",
                    expected: Expected::Rows {
                        columns: &[Column("log_checkpoints", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_checkpoints TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_checkpoints" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_checkpoints')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_connections' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_connections",
                    expected: Expected::Rows {
                        columns: &[Column("log_connections", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_connections TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_connections" cannot be set after connection start"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_connections')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_destination' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_destination",
                    expected: Expected::Rows {
                        columns: &[Column("log_destination", TEXT)],
                        rows: &[
                            &[T("stderr")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_destination TO 'jsonlog'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_destination" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_destination')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("stderr")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_directory' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_directory",
                    expected: Expected::Rows {
                        columns: &[Column("log_directory", TEXT)],
                        rows: &[
                            &[T("log")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_directory TO 'log'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_directory" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_directory')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("log")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_disconnections' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_disconnections",
                    expected: Expected::Rows {
                        columns: &[Column("log_disconnections", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_disconnections TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_disconnections" cannot be set after connection start"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_disconnections')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_duration' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_duration",
                    expected: Expected::Rows {
                        columns: &[Column("log_duration", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_duration TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_duration",
                    expected: Expected::Rows {
                        columns: &[Column("log_duration", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_duration TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_duration",
                    expected: Expected::Rows {
                        columns: &[Column("log_duration", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_duration')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_error_verbosity' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_error_verbosity",
                    expected: Expected::Rows {
                        columns: &[Column("log_error_verbosity", TEXT)],
                        rows: &[
                            &[T("default")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_error_verbosity TO 'verbose'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_error_verbosity",
                    expected: Expected::Rows {
                        columns: &[Column("log_error_verbosity", TEXT)],
                        rows: &[
                            &[T("verbose")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_error_verbosity TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_error_verbosity",
                    expected: Expected::Rows {
                        columns: &[Column("log_error_verbosity", TEXT)],
                        rows: &[
                            &[T("default")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_error_verbosity')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("default")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_executor_stats' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_executor_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_executor_stats", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_executor_stats TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_executor_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_executor_stats", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_executor_stats TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_executor_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_executor_stats", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_executor_stats')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_file_mode' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_file_mode",
                    expected: Expected::Rows {
                        columns: &[Column("log_file_mode", TEXT)],
                        rows: &[
                            &[T("0600")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_file_mode TO '384'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_file_mode" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_file_mode')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0600")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_filename' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_filename",
                    expected: Expected::Rows {
                        columns: &[Column("log_filename", TEXT)],
                        rows: &[
                            &[T("postgresql-%Y-%m-%d_%H%M%S.log")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_filename TO 'postgresql-%Y-%m-%d_%H%M%S.log'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_filename" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_filename')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("postgresql-%Y-%m-%d_%H%M%S.log")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_hostname' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_hostname",
                    expected: Expected::Rows {
                        columns: &[Column("log_hostname", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_hostname TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_hostname" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_hostname')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_line_prefix' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_line_prefix",
                    expected: Expected::Rows {
                        columns: &[Column("log_line_prefix", TEXT)],
                        rows: &[
                            &[T("%m [%p] ")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_line_prefix TO '%m [%p]'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_line_prefix" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_line_prefix')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("%m [%p] ")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_lock_waits' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_lock_waits",
                    expected: Expected::Rows {
                        columns: &[Column("log_lock_waits", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_lock_waits TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_lock_waits",
                    expected: Expected::Rows {
                        columns: &[Column("log_lock_waits", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_lock_waits TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_lock_waits",
                    expected: Expected::Rows {
                        columns: &[Column("log_lock_waits", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_lock_waits')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_min_duration_sample' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_min_duration_sample",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_duration_sample", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_min_duration_sample TO 1",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_min_duration_sample",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_duration_sample", TEXT)],
                        rows: &[
                            &[T("1ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_min_duration_sample TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_min_duration_sample",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_duration_sample", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_min_duration_sample')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_min_duration_statement' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_min_duration_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_duration_statement", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_min_duration_statement TO 10",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_min_duration_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_duration_statement", TEXT)],
                        rows: &[
                            &[T("10ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_min_duration_statement TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_min_duration_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_duration_statement", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_min_duration_statement')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_min_error_statement' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_min_error_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_error_statement", TEXT)],
                        rows: &[
                            &[T("error")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_min_error_statement TO 'debug5'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_min_error_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_error_statement", TEXT)],
                        rows: &[
                            &[T("debug5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_min_error_statement TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_min_error_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_error_statement", TEXT)],
                        rows: &[
                            &[T("error")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_min_error_statement')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("error")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_min_messages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_min_messages",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_messages", TEXT)],
                        rows: &[
                            &[T("warning")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_min_messages TO 'info'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_min_messages",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_messages", TEXT)],
                        rows: &[
                            &[T("info")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_min_messages TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_min_messages",
                    expected: Expected::Rows {
                        columns: &[Column("log_min_messages", TEXT)],
                        rows: &[
                            &[T("warning")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_min_messages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("warning")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_parameter_max_length' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_parameter_max_length",
                    expected: Expected::Rows {
                        columns: &[Column("log_parameter_max_length", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_parameter_max_length TO '10'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_parameter_max_length",
                    expected: Expected::Rows {
                        columns: &[Column("log_parameter_max_length", TEXT)],
                        rows: &[
                            &[T("10B")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_parameter_max_length')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("10B")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_parameter_max_length TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_parameter_max_length",
                    expected: Expected::Rows {
                        columns: &[Column("log_parameter_max_length", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_parameter_max_length')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_parameter_max_length_on_error' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_parameter_max_length_on_error",
                    expected: Expected::Rows {
                        columns: &[Column("log_parameter_max_length_on_error", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_parameter_max_length_on_error TO '1'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_parameter_max_length_on_error",
                    expected: Expected::Rows {
                        columns: &[Column("log_parameter_max_length_on_error", TEXT)],
                        rows: &[
                            &[T("1B")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_parameter_max_length_on_error TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_parameter_max_length_on_error",
                    expected: Expected::Rows {
                        columns: &[Column("log_parameter_max_length_on_error", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_parameter_max_length_on_error')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_parser_stats' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_parser_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_parser_stats", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_parser_stats TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_parser_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_parser_stats", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_parser_stats TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_parser_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_parser_stats", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_parser_stats')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_planner_stats' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_planner_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_planner_stats", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_planner_stats TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_planner_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_planner_stats", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_planner_stats TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_planner_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_planner_stats", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_planner_stats')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_recovery_conflict_waits' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_recovery_conflict_waits",
                    expected: Expected::Rows {
                        columns: &[Column("log_recovery_conflict_waits", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_recovery_conflict_waits TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_recovery_conflict_waits" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_recovery_conflict_waits')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_replication_commands' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_replication_commands",
                    expected: Expected::Rows {
                        columns: &[Column("log_replication_commands", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_replication_commands TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_replication_commands",
                    expected: Expected::Rows {
                        columns: &[Column("log_replication_commands", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_replication_commands TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_replication_commands",
                    expected: Expected::Rows {
                        columns: &[Column("log_replication_commands", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_replication_commands')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_rotation_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_rotation_age",
                    expected: Expected::Rows {
                        columns: &[Column("log_rotation_age", TEXT)],
                        rows: &[
                            &[T("1d")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_rotation_age TO '1440'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_rotation_age" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_rotation_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1d")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_rotation_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_rotation_size",
                    expected: Expected::Rows {
                        columns: &[Column("log_rotation_size", TEXT)],
                        rows: &[
                            &[T("10MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_rotation_size TO '10240'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_rotation_size" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_rotation_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("10MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_startup_progress_interval' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_startup_progress_interval",
                    expected: Expected::Rows {
                        columns: &[Column("log_startup_progress_interval", TEXT)],
                        rows: &[
                            &[T("10s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_startup_progress_interval TO '10'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_startup_progress_interval" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_startup_progress_interval')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("10s")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_statement' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement", TEXT)],
                        rows: &[
                            &[T("none")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_statement TO 'ddl'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement", TEXT)],
                        rows: &[
                            &[T("ddl")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_statement TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_statement",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement", TEXT)],
                        rows: &[
                            &[T("none")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_statement')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_statement_sample_rate' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_statement_sample_rate",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement_sample_rate", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_statement_sample_rate TO 0.5",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_statement_sample_rate",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement_sample_rate", TEXT)],
                        rows: &[
                            &[T("0.5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_statement_sample_rate TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_statement_sample_rate",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement_sample_rate", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_statement_sample_rate')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_statement_stats' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_statement_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement_stats", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_statement_stats TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_statement_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement_stats", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_statement_stats TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_statement_stats",
                    expected: Expected::Rows {
                        columns: &[Column("log_statement_stats", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_statement_stats')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_temp_files' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_temp_files",
                    expected: Expected::Rows {
                        columns: &[Column("log_temp_files", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_temp_files TO '100'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_temp_files",
                    expected: Expected::Rows {
                        columns: &[Column("log_temp_files", TEXT)],
                        rows: &[
                            &[T("100kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_temp_files TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_temp_files",
                    expected: Expected::Rows {
                        columns: &[Column("log_temp_files", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_temp_files')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_timezone' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_timezone",
                    expected: Expected::Rows {
                        columns: &[Column("log_timezone", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_timezone TO 'America/Los_Angeles'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_timezone" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_timezone')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_transaction_sample_rate' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_transaction_sample_rate",
                    expected: Expected::Rows {
                        columns: &[Column("log_transaction_sample_rate", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_transaction_sample_rate TO '0.5'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_transaction_sample_rate",
                    expected: Expected::Rows {
                        columns: &[Column("log_transaction_sample_rate", TEXT)],
                        rows: &[
                            &[T("0.5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_transaction_sample_rate TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW log_transaction_sample_rate",
                    expected: Expected::Rows {
                        columns: &[Column("log_transaction_sample_rate", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_transaction_sample_rate')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'log_truncate_on_rotation' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW log_truncate_on_rotation",
                    expected: Expected::Rows {
                        columns: &[Column("log_truncate_on_rotation", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET log_truncate_on_rotation TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "log_truncate_on_rotation" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('log_truncate_on_rotation')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'logging_collector' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW logging_collector",
                    expected: Expected::Rows {
                        columns: &[Column("logging_collector", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET logging_collector TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "logging_collector" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('logging_collector')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'logical_decoding_work_mem' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW logical_decoding_work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("logical_decoding_work_mem", TEXT)],
                        rows: &[
                            &[T("64MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET logical_decoding_work_mem TO '64000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW logical_decoding_work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("logical_decoding_work_mem", TEXT)],
                        rows: &[
                            &[T("64000kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET logical_decoding_work_mem TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW logical_decoding_work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("logical_decoding_work_mem", TEXT)],
                        rows: &[
                            &[T("64MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('logical_decoding_work_mem')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("64MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'maintenance_io_concurrency' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW maintenance_io_concurrency",
                    expected: Expected::Rows {
                        columns: &[Column("maintenance_io_concurrency", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET maintenance_io_concurrency TO '1'",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid value for parameter "maintenance_io_concurrency": 1"#, detail: "maintenance_io_concurrency must be set to 0 on platforms that lack posix_fadvise().", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW maintenance_io_concurrency",
                    expected: Expected::Rows {
                        columns: &[Column("maintenance_io_concurrency", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET maintenance_io_concurrency TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW maintenance_io_concurrency",
                    expected: Expected::Rows {
                        columns: &[Column("maintenance_io_concurrency", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('maintenance_io_concurrency')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'maintenance_work_mem' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW maintenance_work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("maintenance_work_mem", TEXT)],
                        rows: &[
                            &[T("64MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET maintenance_work_mem TO '64000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW maintenance_work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("maintenance_work_mem", TEXT)],
                        rows: &[
                            &[T("64000kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET maintenance_work_mem TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW maintenance_work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("maintenance_work_mem", TEXT)],
                        rows: &[
                            &[T("64MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('maintenance_work_mem')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("64MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_connections' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_connections",
                    expected: Expected::Rows {
                        columns: &[Column("max_connections", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_connections TO '150'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_connections" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_connections')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_files_per_process' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_files_per_process",
                    expected: Expected::Rows {
                        columns: &[Column("max_files_per_process", TEXT)],
                        rows: &[
                            &[T("1000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_files_per_process TO '1000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_files_per_process" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_files_per_process')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'max_function_args' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_function_args",
                    expected: Expected::Rows {
                        columns: &[Column("max_function_args", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_function_args TO '100'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_function_args" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_function_args')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_identifier_length' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_identifier_length",
                    expected: Expected::Rows {
                        columns: &[Column("max_identifier_length", TEXT)],
                        rows: &[
                            &[T("63")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_identifier_length TO '63'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_identifier_length" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_identifier_length')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("63")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_index_keys' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_index_keys",
                    expected: Expected::Rows {
                        columns: &[Column("max_index_keys", TEXT)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_index_keys TO '32'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_index_keys" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_index_keys')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_locks_per_transaction' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_locks_per_transaction",
                    expected: Expected::Rows {
                        columns: &[Column("max_locks_per_transaction", TEXT)],
                        rows: &[
                            &[T("64")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_locks_per_transaction TO '64'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_locks_per_transaction" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_locks_per_transaction')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("64")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_logical_replication_workers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_logical_replication_workers",
                    expected: Expected::Rows {
                        columns: &[Column("max_logical_replication_workers", TEXT)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_logical_replication_workers TO '4'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_logical_replication_workers" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_logical_replication_workers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_parallel_apply_workers_per_subscription' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_parallel_apply_workers_per_subscription",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "max_parallel_apply_workers_per_subscription""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_parallel_apply_workers_per_subscription TO '2'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "max_parallel_apply_workers_per_subscription""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_parallel_apply_workers_per_subscription')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "max_parallel_apply_workers_per_subscription""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_parallel_maintenance_workers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_parallel_maintenance_workers",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_maintenance_workers", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_parallel_maintenance_workers TO '3'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW max_parallel_maintenance_workers",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_maintenance_workers", TEXT)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_parallel_maintenance_workers TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW max_parallel_maintenance_workers",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_maintenance_workers", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_parallel_maintenance_workers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'max_parallel_workers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_parallel_workers",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_workers", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_parallel_workers TO 11",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW max_parallel_workers",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_workers", TEXT)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_parallel_workers TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW max_parallel_workers",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_workers", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_parallel_workers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
        ScriptTest {
            name: "set 'max_parallel_workers_per_gather' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_parallel_workers_per_gather",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_workers_per_gather", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_parallel_workers_per_gather TO 3",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW max_parallel_workers_per_gather",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_workers_per_gather", TEXT)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_parallel_workers_per_gather TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW max_parallel_workers_per_gather",
                    expected: Expected::Rows {
                        columns: &[Column("max_parallel_workers_per_gather", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_parallel_workers_per_gather')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'max_pred_locks_per_page' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_pred_locks_per_page",
                    expected: Expected::Rows {
                        columns: &[Column("max_pred_locks_per_page", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_pred_locks_per_page TO '2'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_pred_locks_per_page" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_pred_locks_per_page')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'max_pred_locks_per_relation' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_pred_locks_per_relation",
                    expected: Expected::Rows {
                        columns: &[Column("max_pred_locks_per_relation", TEXT)],
                        rows: &[
                            &[T("-2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_pred_locks_per_relation TO '-2'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_pred_locks_per_relation" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_pred_locks_per_relation')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_pred_locks_per_transaction' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_pred_locks_per_transaction",
                    expected: Expected::Rows {
                        columns: &[Column("max_pred_locks_per_transaction", TEXT)],
                        rows: &[
                            &[T("64")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_pred_locks_per_transaction TO '64'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_pred_locks_per_transaction" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_pred_locks_per_transaction')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("64")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_prepared_transactions' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_prepared_transactions",
                    expected: Expected::Rows {
                        columns: &[Column("max_prepared_transactions", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_prepared_transactions TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_prepared_transactions" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_prepared_transactions')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_replication_slots' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_replication_slots",
                    expected: Expected::Rows {
                        columns: &[Column("max_replication_slots", TEXT)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_replication_slots TO '10'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_replication_slots" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_replication_slots')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_slot_wal_keep_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_slot_wal_keep_size",
                    expected: Expected::Rows {
                        columns: &[Column("max_slot_wal_keep_size", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_slot_wal_keep_size TO '-1'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_slot_wal_keep_size" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_slot_wal_keep_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_stack_depth' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_stack_depth",
                    expected: Expected::Rows {
                        columns: &[Column("max_stack_depth", TEXT)],
                        rows: &[
                            &[T("2MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_stack_depth TO '2000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW max_stack_depth",
                    expected: Expected::Rows {
                        columns: &[Column("max_stack_depth", TEXT)],
                        rows: &[
                            &[T("2000kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_stack_depth TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW max_stack_depth",
                    expected: Expected::Rows {
                        columns: &[Column("max_stack_depth", TEXT)],
                        rows: &[
                            &[T("2MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_stack_depth')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("2MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_standby_archive_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_standby_archive_delay",
                    expected: Expected::Rows {
                        columns: &[Column("max_standby_archive_delay", TEXT)],
                        rows: &[
                            &[T("30s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_standby_archive_delay TO '30'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_standby_archive_delay" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_standby_archive_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("30s")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_standby_streaming_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_standby_streaming_delay",
                    expected: Expected::Rows {
                        columns: &[Column("max_standby_streaming_delay", TEXT)],
                        rows: &[
                            &[T("30s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_standby_streaming_delay TO '30'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_standby_streaming_delay" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_standby_streaming_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("30s")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_sync_workers_per_subscription' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_sync_workers_per_subscription",
                    expected: Expected::Rows {
                        columns: &[Column("max_sync_workers_per_subscription", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_sync_workers_per_subscription TO '2'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_sync_workers_per_subscription" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_sync_workers_per_subscription')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'max_wal_senders' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_wal_senders",
                    expected: Expected::Rows {
                        columns: &[Column("max_wal_senders", TEXT)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_wal_senders TO '10'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_wal_senders" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_wal_senders')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_wal_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_wal_size",
                    expected: Expected::Rows {
                        columns: &[Column("max_wal_size", TEXT)],
                        rows: &[
                            &[T("1GB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_wal_size TO '1000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_wal_size" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_wal_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1GB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'max_worker_processes' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW max_worker_processes",
                    expected: Expected::Rows {
                        columns: &[Column("max_worker_processes", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET max_worker_processes TO '8'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "max_worker_processes" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('max_worker_processes')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
        ScriptTest {
            name: "set 'min_dynamic_shared_memory' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW min_dynamic_shared_memory",
                    expected: Expected::Rows {
                        columns: &[Column("min_dynamic_shared_memory", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET min_dynamic_shared_memory TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "min_dynamic_shared_memory" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('min_dynamic_shared_memory')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'min_parallel_index_scan_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW min_parallel_index_scan_size",
                    expected: Expected::Rows {
                        columns: &[Column("min_parallel_index_scan_size", TEXT)],
                        rows: &[
                            &[T("512kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET min_parallel_index_scan_size TO '512'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW min_parallel_index_scan_size",
                    expected: Expected::Rows {
                        columns: &[Column("min_parallel_index_scan_size", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET min_parallel_index_scan_size TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW min_parallel_index_scan_size",
                    expected: Expected::Rows {
                        columns: &[Column("min_parallel_index_scan_size", TEXT)],
                        rows: &[
                            &[T("512kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('min_parallel_index_scan_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("512kB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'min_parallel_table_scan_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW min_parallel_table_scan_size",
                    expected: Expected::Rows {
                        columns: &[Column("min_parallel_table_scan_size", TEXT)],
                        rows: &[
                            &[T("8MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET min_parallel_table_scan_size TO '800'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW min_parallel_table_scan_size",
                    expected: Expected::Rows {
                        columns: &[Column("min_parallel_table_scan_size", TEXT)],
                        rows: &[
                            &[T("6400kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET min_parallel_table_scan_size TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW min_parallel_table_scan_size",
                    expected: Expected::Rows {
                        columns: &[Column("min_parallel_table_scan_size", TEXT)],
                        rows: &[
                            &[T("8MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('min_parallel_table_scan_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("8MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'min_wal_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW min_wal_size",
                    expected: Expected::Rows {
                        columns: &[Column("min_wal_size", TEXT)],
                        rows: &[
                            &[T("80MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET min_wal_size TO '8000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "min_wal_size" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('min_wal_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("80MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'old_snapshot_threshold' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW old_snapshot_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("old_snapshot_threshold", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET old_snapshot_threshold TO '-1'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "old_snapshot_threshold" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('old_snapshot_threshold')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'parallel_leader_participation' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW parallel_leader_participation",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_leader_participation", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET parallel_leader_participation TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW parallel_leader_participation",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_leader_participation", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET parallel_leader_participation TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW parallel_leader_participation",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_leader_participation", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('parallel_leader_participation')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'parallel_setup_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW parallel_setup_cost",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_setup_cost", TEXT)],
                        rows: &[
                            &[T("1000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET parallel_setup_cost TO '10000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW parallel_setup_cost",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_setup_cost", TEXT)],
                        rows: &[
                            &[T("10000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET parallel_setup_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW parallel_setup_cost",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_setup_cost", TEXT)],
                        rows: &[
                            &[T("1000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('parallel_setup_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'parallel_tuple_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW parallel_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET parallel_tuple_cost TO '0.2'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW parallel_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET parallel_tuple_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW parallel_tuple_cost",
                    expected: Expected::Rows {
                        columns: &[Column("parallel_tuple_cost", TEXT)],
                        rows: &[
                            &[T("0.1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('parallel_tuple_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0.1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'password_encryption' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW password_encryption",
                    expected: Expected::Rows {
                        columns: &[Column("password_encryption", TEXT)],
                        rows: &[
                            &[T("scram-sha-256")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET password_encryption TO 'md5'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW password_encryption",
                    expected: Expected::Rows {
                        columns: &[Column("password_encryption", TEXT)],
                        rows: &[
                            &[T("md5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET password_encryption TO 'scram-sha-256'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW password_encryption",
                    expected: Expected::Rows {
                        columns: &[Column("password_encryption", TEXT)],
                        rows: &[
                            &[T("scram-sha-256")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('password_encryption')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("scram-sha-256")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'plan_cache_mode' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW plan_cache_mode",
                    expected: Expected::Rows {
                        columns: &[Column("plan_cache_mode", TEXT)],
                        rows: &[
                            &[T("auto")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET plan_cache_mode TO 'force_generic_plan'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW plan_cache_mode",
                    expected: Expected::Rows {
                        columns: &[Column("plan_cache_mode", TEXT)],
                        rows: &[
                            &[T("force_generic_plan")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET plan_cache_mode TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW plan_cache_mode",
                    expected: Expected::Rows {
                        columns: &[Column("plan_cache_mode", TEXT)],
                        rows: &[
                            &[T("auto")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('plan_cache_mode')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("auto")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'post_auth_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW post_auth_delay",
                    expected: Expected::Rows {
                        columns: &[Column("post_auth_delay", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET post_auth_delay TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "post_auth_delay" cannot be set after connection start"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('post_auth_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'pre_auth_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW pre_auth_delay",
                    expected: Expected::Rows {
                        columns: &[Column("pre_auth_delay", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET pre_auth_delay TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "pre_auth_delay" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('pre_auth_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'primary_conninfo' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW primary_conninfo",
                    expected: Expected::Rows {
                        columns: &[Column("primary_conninfo", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET primary_conninfo TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "primary_conninfo" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('primary_conninfo')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'primary_slot_name' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW primary_slot_name",
                    expected: Expected::Rows {
                        columns: &[Column("primary_slot_name", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET primary_slot_name TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "primary_slot_name" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('primary_slot_name')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'quote_all_identifiers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW quote_all_identifiers",
                    expected: Expected::Rows {
                        columns: &[Column("quote_all_identifiers", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET quote_all_identifiers TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW quote_all_identifiers",
                    expected: Expected::Rows {
                        columns: &[Column("quote_all_identifiers", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET quote_all_identifiers TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW quote_all_identifiers",
                    expected: Expected::Rows {
                        columns: &[Column("quote_all_identifiers", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('quote_all_identifiers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'random_page_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW random_page_cost",
                    expected: Expected::Rows {
                        columns: &[Column("random_page_cost", TEXT)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET random_page_cost TO 2.5",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW random_page_cost",
                    expected: Expected::Rows {
                        columns: &[Column("random_page_cost", TEXT)],
                        rows: &[
                            &[T("2.5")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET random_page_cost TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW random_page_cost",
                    expected: Expected::Rows {
                        columns: &[Column("random_page_cost", TEXT)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('random_page_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_end_command' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_end_command",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_end_command", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_end_command TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_end_command" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_end_command')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_init_sync_method' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_init_sync_method",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_init_sync_method", TEXT)],
                        rows: &[
                            &[T("fsync")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_init_sync_method TO 'fsync'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_init_sync_method" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_init_sync_method')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("fsync")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_min_apply_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_min_apply_delay",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_min_apply_delay", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_min_apply_delay TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_min_apply_delay" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_min_apply_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_prefetch' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_prefetch",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_prefetch", TEXT)],
                        rows: &[
                            &[T("try")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_prefetch TO 'try'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_prefetch" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_prefetch')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("try")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_target' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_target",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_target", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_target TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_target" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_target')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_target_action' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_target_action",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_target_action", TEXT)],
                        rows: &[
                            &[T("pause")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_target_action TO 'pause'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_target_action" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_target_action')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("pause")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_target_inclusive' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_target_inclusive",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_target_inclusive", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_target_inclusive TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_target_inclusive" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_target_inclusive')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_target_lsn' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_target_lsn",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_target_lsn", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_target_lsn TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_target_lsn" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_target_lsn')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_target_name' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_target_name",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_target_name", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_target_name TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_target_name" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_target_name')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_target_time' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_target_time",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_target_time", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_target_time TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_target_time" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_target_time')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_target_timeline' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_target_timeline",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_target_timeline", TEXT)],
                        rows: &[
                            &[T("latest")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_target_timeline TO 'latest'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_target_timeline" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_target_timeline')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("latest")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recovery_target_xid' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recovery_target_xid",
                    expected: Expected::Rows {
                        columns: &[Column("recovery_target_xid", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recovery_target_xid TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "recovery_target_xid" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recovery_target_xid')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'recursive_worktable_factor' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW recursive_worktable_factor",
                    expected: Expected::Rows {
                        columns: &[Column("recursive_worktable_factor", TEXT)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recursive_worktable_factor TO '1'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW recursive_worktable_factor",
                    expected: Expected::Rows {
                        columns: &[Column("recursive_worktable_factor", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET recursive_worktable_factor TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW recursive_worktable_factor",
                    expected: Expected::Rows {
                        columns: &[Column("recursive_worktable_factor", TEXT)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('recursive_worktable_factor')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'remove_temp_files_after_crash' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW remove_temp_files_after_crash",
                    expected: Expected::Rows {
                        columns: &[Column("remove_temp_files_after_crash", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET remove_temp_files_after_crash TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "remove_temp_files_after_crash" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('remove_temp_files_after_crash')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'reserved_connections' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW reserved_connections",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "reserved_connections""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET reserved_connections TO '0'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "reserved_connections""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('reserved_connections')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "reserved_connections""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'restart_after_crash' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW restart_after_crash",
                    expected: Expected::Rows {
                        columns: &[Column("restart_after_crash", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET restart_after_crash TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "restart_after_crash" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('restart_after_crash')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'restore_command' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW restore_command",
                    expected: Expected::Rows {
                        columns: &[Column("restore_command", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET restore_command TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "restore_command" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('restore_command')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'row_security' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW row_security",
                    expected: Expected::Rows {
                        columns: &[Column("row_security", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET row_security TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW row_security",
                    expected: Expected::Rows {
                        columns: &[Column("row_security", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET row_security TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW row_security",
                    expected: Expected::Rows {
                        columns: &[Column("row_security", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('row_security')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'scram_iterations' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW scram_iterations",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "scram_iterations""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET scram_iterations TO '4000'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "scram_iterations""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW scram_iterations",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "scram_iterations""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET scram_iterations TO DEFAULT",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "scram_iterations""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW scram_iterations",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "scram_iterations""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('scram_iterations')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "scram_iterations""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'search_path' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO 'postgres'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('search_path')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "search_path elements keep the quoting they need",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "$user", public"#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "$user""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO public",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "public""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "MixedCase", public"#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""MixedCase", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO MiXeD",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("mixed")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "with space""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""with space""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "with""quote""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""with""quote""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "user", public"#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO pg_catalog, public",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("pg_catalog, public")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO public, public2",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("public, public2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO 1, public",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("1, public")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO 'a, b'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""a, b""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO 'public, public2'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""public, public2""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO '"$user", public'"#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""""$user"", public""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO ''",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('search_path')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T(r#""""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "MixedCase""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET search_path",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""$user", public"#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SCHEMA 'MixedCase'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T(r#""MixedCase""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SCHEMA 'postgres'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "non-identifier parameters are not quoted",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET datestyle TO ISO, MDY",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW datestyle",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("ISO, MDY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET application_name TO "MixedCase""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW application_name",
                    expected: Expected::Rows {
                        columns: &[Column("application_name", TEXT)],
                        rows: &[
                            &[T("MixedCase")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET timezone TO "UTC""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW timezone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("UTC")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "quoted search_path elements resolve to their schemas",
            set_up_script: &[
                r#"CREATE SCHEMA "MixedCase";"#,
                "CREATE SCHEMA postgres;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SET search_path TO "MixedCase""#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema()",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("MixedCase")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t1 (a int)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n.nspname FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE c.relname = 't1'",
                    expected: Expected::Rows {
                        columns: &[Column("nspname", NAME)],
                        rows: &[
                            &[T("MixedCase")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "$user", public"#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema()",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE t2 (a int)",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n.nspname FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE c.relname = 't2'",
                    expected: Expected::Rows {
                        columns: &[Column("nspname", NAME)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO 'MixedCase, public'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false)",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'segment_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW segment_size",
                    expected: Expected::Rows {
                        columns: &[Column("segment_size", TEXT)],
                        rows: &[
                            &[T("1GB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET segment_size TO '131072'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "segment_size" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('segment_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1GB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'send_abort_for_crash' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW send_abort_for_crash",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "send_abort_for_crash""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET send_abort_for_crash TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "send_abort_for_crash""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('send_abort_for_crash')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "send_abort_for_crash""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'send_abort_for_kill' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW send_abort_for_kill",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "send_abort_for_kill""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET send_abort_for_kill TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "send_abort_for_kill""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('send_abort_for_kill')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "send_abort_for_kill""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'seq_page_cost' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW seq_page_cost",
                    expected: Expected::Rows {
                        columns: &[Column("seq_page_cost", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET seq_page_cost TO '1'",
                    expected: Expected::Tag("SET"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('seq_page_cost')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'server_encoding' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW server_encoding",
                    expected: Expected::Rows {
                        columns: &[Column("server_encoding", TEXT)],
                        rows: &[
                            &[T("UTF8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET server_encoding TO 'UTF8'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "server_encoding" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('server_encoding')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("UTF8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'server_version' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW server_version",
                    expected: Expected::Rows {
                        columns: &[Column("server_version", TEXT)],
                        rows: &[
                            &[T("15.19 (Homebrew)")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET server_version TO '15.17 (Homebrew)'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "server_version" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('server_version')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("15.19 (Homebrew)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'server_version_num' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW server_version_num",
                    expected: Expected::Rows {
                        columns: &[Column("server_version_num", TEXT)],
                        rows: &[
                            &[T("150019")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET server_version_num TO '150017'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "server_version_num" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('server_version_num')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("150019")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'session_preload_libraries' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW session_preload_libraries",
                    expected: Expected::Rows {
                        columns: &[Column("session_preload_libraries", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET session_preload_libraries TO '/'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW session_preload_libraries",
                    expected: Expected::Rows {
                        columns: &[Column("session_preload_libraries", TEXT)],
                        rows: &[
                            &[T(r#""/""#)],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET session_preload_libraries TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW session_preload_libraries",
                    expected: Expected::Rows {
                        columns: &[Column("session_preload_libraries", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('session_preload_libraries')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'session_replication_role' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW session_replication_role",
                    expected: Expected::Rows {
                        columns: &[Column("session_replication_role", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET session_replication_role TO 'local'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW session_replication_role",
                    expected: Expected::Rows {
                        columns: &[Column("session_replication_role", TEXT)],
                        rows: &[
                            &[T("local")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET session_replication_role TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW session_replication_role",
                    expected: Expected::Rows {
                        columns: &[Column("session_replication_role", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('session_replication_role')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("origin")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'shared_buffers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW shared_buffers",
                    expected: Expected::Rows {
                        columns: &[Column("shared_buffers", TEXT)],
                        rows: &[
                            &[T("128MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET shared_buffers TO '128000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "shared_buffers" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('shared_buffers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("128MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'shared_memory_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW shared_memory_size",
                    expected: Expected::Rows {
                        columns: &[Column("shared_memory_size", TEXT)],
                        rows: &[
                            &[T("143MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET shared_memory_size TO '143000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "shared_memory_size" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('shared_memory_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("143MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'shared_memory_size_in_huge_pages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW shared_memory_size_in_huge_pages",
                    expected: Expected::Rows {
                        columns: &[Column("shared_memory_size_in_huge_pages", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET shared_memory_size_in_huge_pages TO '-1'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "shared_memory_size_in_huge_pages" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('shared_memory_size_in_huge_pages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'shared_memory_type' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW shared_memory_type",
                    expected: Expected::Rows {
                        columns: &[Column("shared_memory_type", TEXT)],
                        rows: &[
                            &[T("mmap")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET shared_memory_type TO 'mmap'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "shared_memory_type" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('shared_memory_type')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("mmap")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'shared_preload_libraries' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW shared_preload_libraries",
                    expected: Expected::Rows {
                        columns: &[Column("shared_preload_libraries", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET shared_preload_libraries TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "shared_preload_libraries" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('shared_preload_libraries')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl",
                    expected: Expected::Rows {
                        columns: &[Column("ssl", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_ca_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_ca_file",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_ca_file", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_ca_file TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_ca_file" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_ca_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_cert_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_cert_file",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_cert_file", TEXT)],
                        rows: &[
                            &[T("server.crt")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_cert_file TO 'server.crt'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_cert_file" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_cert_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("server.crt")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_ciphers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_ciphers",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_ciphers", TEXT)],
                        rows: &[
                            &[T("HIGH:MEDIUM:+3DES:!aNULL")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_ciphers TO 'HIGH:MEDIUM:'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_ciphers" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_ciphers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("HIGH:MEDIUM:+3DES:!aNULL")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_crl_dir' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_crl_dir",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_crl_dir", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_crl_dir TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_crl_dir" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_crl_dir')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_crl_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_crl_file",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_crl_file", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_crl_file TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_crl_file" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_crl_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_dh_params_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_dh_params_file",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_dh_params_file", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_dh_params_file TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_dh_params_file" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_dh_params_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_ecdh_curve' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_ecdh_curve",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_ecdh_curve", TEXT)],
                        rows: &[
                            &[T("prime256v1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_ecdh_curve TO 'prime256v1'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_ecdh_curve" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_ecdh_curve')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("prime256v1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_key_file' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_key_file",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_key_file", TEXT)],
                        rows: &[
                            &[T("server.key")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_key_file TO 'server.key'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_key_file" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_key_file')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("server.key")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_library' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_library",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_library", TEXT)],
                        rows: &[
                            &[T("OpenSSL")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_library TO 'OpenSSL'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_library" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_library')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("OpenSSL")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_max_protocol_version' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_max_protocol_version",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_max_protocol_version", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_max_protocol_version TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_max_protocol_version" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_max_protocol_version')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_min_protocol_version' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_min_protocol_version",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_min_protocol_version", TEXT)],
                        rows: &[
                            &[T("TLSv1.2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_min_protocol_version TO 'TLSv1.2'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_min_protocol_version" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_min_protocol_version')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("TLSv1.2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_passphrase_command' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_passphrase_command",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_passphrase_command", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_passphrase_command TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_passphrase_command" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_passphrase_command')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_passphrase_command_supports_reload' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_passphrase_command_supports_reload",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_passphrase_command_supports_reload", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_passphrase_command_supports_reload TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_passphrase_command_supports_reload" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_passphrase_command_supports_reload')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'ssl_prefer_server_ciphers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW ssl_prefer_server_ciphers",
                    expected: Expected::Rows {
                        columns: &[Column("ssl_prefer_server_ciphers", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET ssl_prefer_server_ciphers TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "ssl_prefer_server_ciphers" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('ssl_prefer_server_ciphers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'standard_conforming_strings' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW standard_conforming_strings",
                    expected: Expected::Rows {
                        columns: &[Column("standard_conforming_strings", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET standard_conforming_strings TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW standard_conforming_strings",
                    expected: Expected::Rows {
                        columns: &[Column("standard_conforming_strings", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET standard_conforming_strings TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW standard_conforming_strings",
                    expected: Expected::Rows {
                        columns: &[Column("standard_conforming_strings", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('standard_conforming_strings')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'statement_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW statement_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("statement_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET statement_timeout TO '42'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW statement_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("statement_timeout", TEXT)],
                        rows: &[
                            &[T("42ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('statement_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("42ms")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'stats_fetch_consistency' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW stats_fetch_consistency",
                    expected: Expected::Rows {
                        columns: &[Column("stats_fetch_consistency", TEXT)],
                        rows: &[
                            &[T("cache")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET stats_fetch_consistency TO 'snapshot'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW stats_fetch_consistency",
                    expected: Expected::Rows {
                        columns: &[Column("stats_fetch_consistency", TEXT)],
                        rows: &[
                            &[T("snapshot")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET stats_fetch_consistency TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW stats_fetch_consistency",
                    expected: Expected::Rows {
                        columns: &[Column("stats_fetch_consistency", TEXT)],
                        rows: &[
                            &[T("cache")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('stats_fetch_consistency')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("cache")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'superuser_reserved_connections' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW superuser_reserved_connections",
                    expected: Expected::Rows {
                        columns: &[Column("superuser_reserved_connections", TEXT)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET superuser_reserved_connections TO '3'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "superuser_reserved_connections" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('superuser_reserved_connections')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'synchronize_seqscans' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW synchronize_seqscans",
                    expected: Expected::Rows {
                        columns: &[Column("synchronize_seqscans", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET synchronize_seqscans TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW synchronize_seqscans",
                    expected: Expected::Rows {
                        columns: &[Column("synchronize_seqscans", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET synchronize_seqscans TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW synchronize_seqscans",
                    expected: Expected::Rows {
                        columns: &[Column("synchronize_seqscans", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('synchronize_seqscans')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'synchronous_commit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW synchronous_commit",
                    expected: Expected::Rows {
                        columns: &[Column("synchronous_commit", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET synchronous_commit TO 'local'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW synchronous_commit",
                    expected: Expected::Rows {
                        columns: &[Column("synchronous_commit", TEXT)],
                        rows: &[
                            &[T("local")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET synchronous_commit TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW synchronous_commit",
                    expected: Expected::Rows {
                        columns: &[Column("synchronous_commit", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('synchronous_commit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'synchronous_standby_names' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW synchronous_standby_names",
                    expected: Expected::Rows {
                        columns: &[Column("synchronous_standby_names", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET synchronous_standby_names TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "synchronous_standby_names" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('synchronous_standby_names')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'syslog_facility' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW syslog_facility",
                    expected: Expected::Rows {
                        columns: &[Column("syslog_facility", TEXT)],
                        rows: &[
                            &[T("local0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET syslog_facility TO 'local0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "syslog_facility" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('syslog_facility')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("local0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'syslog_ident' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW syslog_ident",
                    expected: Expected::Rows {
                        columns: &[Column("syslog_ident", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET syslog_ident TO 'postgres'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "syslog_ident" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('syslog_ident')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("postgres")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'syslog_sequence_numbers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW syslog_sequence_numbers",
                    expected: Expected::Rows {
                        columns: &[Column("syslog_sequence_numbers", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET syslog_sequence_numbers TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "syslog_sequence_numbers" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('syslog_sequence_numbers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'syslog_split_messages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW syslog_split_messages",
                    expected: Expected::Rows {
                        columns: &[Column("syslog_split_messages", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET syslog_split_messages TO 'on'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "syslog_split_messages" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('syslog_split_messages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'tcp_keepalives_count' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_count",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_count", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET tcp_keepalives_count TO 100",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_count",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_count", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET tcp_keepalives_count TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_count",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_count", TEXT)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('tcp_keepalives_count')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
        ScriptTest {
            name: "set 'tcp_keepalives_idle' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_idle",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_idle", TEXT)],
                        rows: &[
                            &[T("7200")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET tcp_keepalives_idle TO 1",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_idle",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_idle", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET tcp_keepalives_idle TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_idle",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_idle", TEXT)],
                        rows: &[
                            &[T("7200")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('tcp_keepalives_idle')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("7200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'tcp_keepalives_interval' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_interval",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_interval", TEXT)],
                        rows: &[
                            &[T("75")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET tcp_keepalives_interval TO 1",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_interval",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_interval", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET tcp_keepalives_interval TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW tcp_keepalives_interval",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_keepalives_interval", TEXT)],
                        rows: &[
                            &[T("75")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('tcp_keepalives_interval')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("75")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'tcp_user_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW tcp_user_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_user_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET tcp_user_timeout TO '100000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW tcp_user_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_user_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET tcp_user_timeout TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW tcp_user_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("tcp_user_timeout", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('tcp_user_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'temp_buffers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW temp_buffers",
                    expected: Expected::Rows {
                        columns: &[Column("temp_buffers", TEXT)],
                        rows: &[
                            &[T("8MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET temp_buffers TO '8000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW temp_buffers",
                    expected: Expected::Rows {
                        columns: &[Column("temp_buffers", TEXT)],
                        rows: &[
                            &[T("64000kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET temp_buffers TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW temp_buffers",
                    expected: Expected::Rows {
                        columns: &[Column("temp_buffers", TEXT)],
                        rows: &[
                            &[T("8MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('temp_buffers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("8MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'temp_file_limit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW temp_file_limit",
                    expected: Expected::Rows {
                        columns: &[Column("temp_file_limit", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET temp_file_limit TO 100",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW temp_file_limit",
                    expected: Expected::Rows {
                        columns: &[Column("temp_file_limit", TEXT)],
                        rows: &[
                            &[T("100kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET temp_file_limit TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW temp_file_limit",
                    expected: Expected::Rows {
                        columns: &[Column("temp_file_limit", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('temp_file_limit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'temp_tablespaces' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW temp_tablespaces",
                    expected: Expected::Rows {
                        columns: &[Column("temp_tablespaces", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET temp_tablespaces TO 'pg_default'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW temp_tablespaces",
                    expected: Expected::Rows {
                        columns: &[Column("temp_tablespaces", TEXT)],
                        rows: &[
                            &[T("pg_default")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET temp_tablespaces TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW temp_tablespaces",
                    expected: Expected::Rows {
                        columns: &[Column("temp_tablespaces", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('temp_tablespaces')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'TimeZone' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW TimeZone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TimeZone TO 'UTC'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TimeZone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("UTC")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TimeZone TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW TimeZone",
                    expected: Expected::Rows {
                        columns: &[Column("TimeZone", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('TimeZone')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("America/Los_Angeles")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'timezone_abbreviations' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW timezone_abbreviations",
                    expected: Expected::Rows {
                        columns: &[Column("timezone_abbreviations", TEXT)],
                        rows: &[
                            &[T("Default")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone_abbreviations TO ''",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"could not read time zone file "": Is a directory"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW timezone_abbreviations",
                    expected: Expected::Rows {
                        columns: &[Column("timezone_abbreviations", TEXT)],
                        rows: &[
                            &[T("Default")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone_abbreviations TO 'Default'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW timezone_abbreviations",
                    expected: Expected::Rows {
                        columns: &[Column("timezone_abbreviations", TEXT)],
                        rows: &[
                            &[T("Default")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('timezone_abbreviations')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("Default")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'trace_notify' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW trace_notify",
                    expected: Expected::Rows {
                        columns: &[Column("trace_notify", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET trace_notify TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW trace_notify",
                    expected: Expected::Rows {
                        columns: &[Column("trace_notify", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET trace_notify TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW trace_notify",
                    expected: Expected::Rows {
                        columns: &[Column("trace_notify", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('trace_notify')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'trace_recovery_messages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW trace_recovery_messages",
                    expected: Expected::Rows {
                        columns: &[Column("trace_recovery_messages", TEXT)],
                        rows: &[
                            &[T("log")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET trace_recovery_messages TO 'log'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "trace_recovery_messages" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('trace_recovery_messages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("log")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'trace_sort' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW trace_sort",
                    expected: Expected::Rows {
                        columns: &[Column("trace_sort", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET trace_sort TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW trace_sort",
                    expected: Expected::Rows {
                        columns: &[Column("trace_sort", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET trace_sort TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW trace_sort",
                    expected: Expected::Rows {
                        columns: &[Column("trace_sort", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('trace_sort')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'track_activities' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW track_activities",
                    expected: Expected::Rows {
                        columns: &[Column("track_activities", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_activities TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_activities",
                    expected: Expected::Rows {
                        columns: &[Column("track_activities", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_activities TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_activities",
                    expected: Expected::Rows {
                        columns: &[Column("track_activities", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('track_activities')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'track_activity_query_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW track_activity_query_size",
                    expected: Expected::Rows {
                        columns: &[Column("track_activity_query_size", TEXT)],
                        rows: &[
                            &[T("1kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_activity_query_size TO '1024'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "track_activity_query_size" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('track_activity_query_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1kB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'track_commit_timestamp' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW track_commit_timestamp",
                    expected: Expected::Rows {
                        columns: &[Column("track_commit_timestamp", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_commit_timestamp TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "track_commit_timestamp" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('track_commit_timestamp')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'track_counts' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW track_counts",
                    expected: Expected::Rows {
                        columns: &[Column("track_counts", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_counts TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_counts",
                    expected: Expected::Rows {
                        columns: &[Column("track_counts", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_counts TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_counts",
                    expected: Expected::Rows {
                        columns: &[Column("track_counts", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('track_counts')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'track_functions' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW track_functions",
                    expected: Expected::Rows {
                        columns: &[Column("track_functions", TEXT)],
                        rows: &[
                            &[T("none")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_functions TO 'all'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_functions",
                    expected: Expected::Rows {
                        columns: &[Column("track_functions", TEXT)],
                        rows: &[
                            &[T("all")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_functions TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_functions",
                    expected: Expected::Rows {
                        columns: &[Column("track_functions", TEXT)],
                        rows: &[
                            &[T("none")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('track_functions')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("none")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'track_io_timing' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW track_io_timing",
                    expected: Expected::Rows {
                        columns: &[Column("track_io_timing", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_io_timing TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_io_timing",
                    expected: Expected::Rows {
                        columns: &[Column("track_io_timing", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_io_timing TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_io_timing",
                    expected: Expected::Rows {
                        columns: &[Column("track_io_timing", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('track_io_timing')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'track_wal_io_timing' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW track_wal_io_timing",
                    expected: Expected::Rows {
                        columns: &[Column("track_wal_io_timing", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_wal_io_timing TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_wal_io_timing",
                    expected: Expected::Rows {
                        columns: &[Column("track_wal_io_timing", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET track_wal_io_timing TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW track_wal_io_timing",
                    expected: Expected::Rows {
                        columns: &[Column("track_wal_io_timing", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('track_wal_io_timing')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'transaction_deferrable' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW transaction_deferrable",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_deferrable", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET transaction_deferrable TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_deferrable",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_deferrable", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET transaction_deferrable TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_deferrable",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_deferrable", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('transaction_deferrable')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'transaction_isolation' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW transaction_isolation",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_isolation", TEXT)],
                        rows: &[
                            &[T("read committed")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET transaction_isolation TO 'serializable'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_isolation",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_isolation", TEXT)],
                        rows: &[
                            &[T("read committed")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET transaction_isolation TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    notices: &[Diagnostic { severity: "WARNING", code: "25P01", message: "RESET TRANSACTION can only be used in transaction blocks", ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_isolation",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_isolation", TEXT)],
                        rows: &[
                            &[T("read committed")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('transaction_isolation')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("read committed")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'transaction_read_only' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW transaction_read_only",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_read_only", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET transaction_read_only TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_read_only",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_read_only", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET transaction_read_only TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transaction_read_only",
                    expected: Expected::Rows {
                        columns: &[Column("transaction_read_only", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('transaction_read_only')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'transform_null_equals' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW transform_null_equals",
                    expected: Expected::Rows {
                        columns: &[Column("transform_null_equals", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET transform_null_equals TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transform_null_equals",
                    expected: Expected::Rows {
                        columns: &[Column("transform_null_equals", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET transform_null_equals TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW transform_null_equals",
                    expected: Expected::Rows {
                        columns: &[Column("transform_null_equals", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('transform_null_equals')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'unix_socket_directories' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW unix_socket_directories",
                    expected: Expected::Rows {
                        columns: &[Column("unix_socket_directories", TEXT)],
                        rows: &[
                            &[Any],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET unix_socket_directories TO '/tmp'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "unix_socket_directories" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('unix_socket_directories')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'unix_socket_group' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW unix_socket_group",
                    expected: Expected::Rows {
                        columns: &[Column("unix_socket_group", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET unix_socket_group TO ''",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "unix_socket_group" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('unix_socket_group')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'unix_socket_permissions' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW unix_socket_permissions",
                    expected: Expected::Rows {
                        columns: &[Column("unix_socket_permissions", TEXT)],
                        rows: &[
                            &[T("0777")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET unix_socket_permissions TO '511'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "unix_socket_permissions" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('unix_socket_permissions')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0777")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'update_process_title' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW update_process_title",
                    expected: Expected::Rows {
                        columns: &[Column("update_process_title", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET update_process_title TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW update_process_title",
                    expected: Expected::Rows {
                        columns: &[Column("update_process_title", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET update_process_title TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW update_process_title",
                    expected: Expected::Rows {
                        columns: &[Column("update_process_title", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('update_process_title')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_buffer_usage_limit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_buffer_usage_limit",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "vacuum_buffer_usage_limit""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_buffer_usage_limit TO '512'",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "vacuum_buffer_usage_limit""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_buffer_usage_limit",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "vacuum_buffer_usage_limit""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_buffer_usage_limit TO DEFAULT",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "vacuum_buffer_usage_limit""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_buffer_usage_limit",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "vacuum_buffer_usage_limit""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_buffer_usage_limit')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "vacuum_buffer_usage_limit""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_cost_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_delay",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_delay", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_delay TO '0.2'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_delay",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_delay", TEXT)],
                        rows: &[
                            &[T("200us")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_delay TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_delay",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_delay", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_cost_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_cost_limit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_limit",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_limit", TEXT)],
                        rows: &[
                            &[T("200")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_limit TO '400'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_limit",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_limit", TEXT)],
                        rows: &[
                            &[T("400")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_limit TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_limit",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_limit", TEXT)],
                        rows: &[
                            &[T("200")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_cost_limit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_cost_page_dirty' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_dirty",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_dirty", TEXT)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_page_dirty TO '200'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_dirty",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_dirty", TEXT)],
                        rows: &[
                            &[T("200")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_page_dirty TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_dirty",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_dirty", TEXT)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_cost_page_dirty')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_cost_page_hit' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_hit",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_hit", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_page_hit TO '100'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_hit",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_hit", TEXT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_page_hit TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_hit",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_hit", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_cost_page_hit')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_cost_page_miss' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_miss",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_miss", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_page_miss TO '20'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_miss",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_miss", TEXT)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_cost_page_miss TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_cost_page_miss",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_cost_page_miss", TEXT)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_cost_page_miss')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
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
            name: "set 'vacuum_failsafe_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_failsafe_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_failsafe_age", TEXT)],
                        rows: &[
                            &[T("1600000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_failsafe_age TO '2100000000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_failsafe_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_failsafe_age", TEXT)],
                        rows: &[
                            &[T("2100000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_failsafe_age TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_failsafe_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_failsafe_age", TEXT)],
                        rows: &[
                            &[T("1600000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_failsafe_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1600000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_freeze_min_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_freeze_min_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_freeze_min_age", TEXT)],
                        rows: &[
                            &[T("50000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_freeze_min_age TO '20000000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_freeze_min_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_freeze_min_age", TEXT)],
                        rows: &[
                            &[T("20000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_freeze_min_age TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_freeze_min_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_freeze_min_age", TEXT)],
                        rows: &[
                            &[T("50000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_freeze_min_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("50000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_freeze_table_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_freeze_table_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_freeze_table_age", TEXT)],
                        rows: &[
                            &[T("150000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_freeze_table_age TO '100000000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_freeze_table_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_freeze_table_age", TEXT)],
                        rows: &[
                            &[T("100000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_freeze_table_age TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_freeze_table_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_freeze_table_age", TEXT)],
                        rows: &[
                            &[T("150000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_freeze_table_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("150000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_multixact_failsafe_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_failsafe_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_failsafe_age", TEXT)],
                        rows: &[
                            &[T("1600000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_multixact_failsafe_age TO '1000000000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_failsafe_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_failsafe_age", TEXT)],
                        rows: &[
                            &[T("1000000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_multixact_failsafe_age TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_failsafe_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_failsafe_age", TEXT)],
                        rows: &[
                            &[T("1600000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_multixact_failsafe_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1600000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_multixact_freeze_min_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_freeze_min_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_freeze_min_age", TEXT)],
                        rows: &[
                            &[T("5000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_multixact_freeze_min_age TO '2000000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_freeze_min_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_freeze_min_age", TEXT)],
                        rows: &[
                            &[T("2000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_multixact_freeze_min_age TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_freeze_min_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_freeze_min_age", TEXT)],
                        rows: &[
                            &[T("5000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_multixact_freeze_min_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("5000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'vacuum_multixact_freeze_table_age' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_freeze_table_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_freeze_table_age", TEXT)],
                        rows: &[
                            &[T("150000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_multixact_freeze_table_age TO '120000000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_freeze_table_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_freeze_table_age", TEXT)],
                        rows: &[
                            &[T("120000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET vacuum_multixact_freeze_table_age TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW vacuum_multixact_freeze_table_age",
                    expected: Expected::Rows {
                        columns: &[Column("vacuum_multixact_freeze_table_age", TEXT)],
                        rows: &[
                            &[T("150000000")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('vacuum_multixact_freeze_table_age')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("150000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_block_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_block_size",
                    expected: Expected::Rows {
                        columns: &[Column("wal_block_size", TEXT)],
                        rows: &[
                            &[T("8192")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_block_size TO '8192'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_block_size" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_block_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("8192")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_buffers' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_buffers",
                    expected: Expected::Rows {
                        columns: &[Column("wal_buffers", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_buffers TO '4000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_buffers" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_buffers')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_compression' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_compression",
                    expected: Expected::Rows {
                        columns: &[Column("wal_compression", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_compression TO 'lz4'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_compression",
                    expected: Expected::Rows {
                        columns: &[Column("wal_compression", TEXT)],
                        rows: &[
                            &[T("lz4")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_compression TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_compression",
                    expected: Expected::Rows {
                        columns: &[Column("wal_compression", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_compression')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_consistency_checking' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_consistency_checking",
                    expected: Expected::Rows {
                        columns: &[Column("wal_consistency_checking", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_consistency_checking TO 'generic'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_consistency_checking",
                    expected: Expected::Rows {
                        columns: &[Column("wal_consistency_checking", TEXT)],
                        rows: &[
                            &[T("generic")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_consistency_checking TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_consistency_checking",
                    expected: Expected::Rows {
                        columns: &[Column("wal_consistency_checking", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_consistency_checking')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_decode_buffer_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_decode_buffer_size",
                    expected: Expected::Rows {
                        columns: &[Column("wal_decode_buffer_size", TEXT)],
                        rows: &[
                            &[T("512kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_decode_buffer_size TO '524288'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_decode_buffer_size" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_decode_buffer_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("512kB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_init_zero' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_init_zero",
                    expected: Expected::Rows {
                        columns: &[Column("wal_init_zero", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_init_zero TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_init_zero",
                    expected: Expected::Rows {
                        columns: &[Column("wal_init_zero", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_init_zero TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_init_zero",
                    expected: Expected::Rows {
                        columns: &[Column("wal_init_zero", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_init_zero')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_keep_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_keep_size",
                    expected: Expected::Rows {
                        columns: &[Column("wal_keep_size", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_keep_size TO '0'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_keep_size" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_keep_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_level' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_level",
                    expected: Expected::Rows {
                        columns: &[Column("wal_level", TEXT)],
                        rows: &[
                            &[T("replica")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_level TO 'replica'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_level" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_level')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("replica")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_log_hints' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_log_hints",
                    expected: Expected::Rows {
                        columns: &[Column("wal_log_hints", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_log_hints TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_log_hints" cannot be changed without restarting the server"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_log_hints')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_receiver_create_temp_slot' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_receiver_create_temp_slot",
                    expected: Expected::Rows {
                        columns: &[Column("wal_receiver_create_temp_slot", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_receiver_create_temp_slot TO 'off'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_receiver_create_temp_slot" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_receiver_create_temp_slot')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_receiver_status_interval' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_receiver_status_interval",
                    expected: Expected::Rows {
                        columns: &[Column("wal_receiver_status_interval", TEXT)],
                        rows: &[
                            &[T("10s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_receiver_status_interval TO '10'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_receiver_status_interval" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_receiver_status_interval')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("10s")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_receiver_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_receiver_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("wal_receiver_timeout", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_receiver_timeout TO '60'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_receiver_timeout" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_receiver_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_recycle' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_recycle",
                    expected: Expected::Rows {
                        columns: &[Column("wal_recycle", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_recycle TO 'off'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_recycle",
                    expected: Expected::Rows {
                        columns: &[Column("wal_recycle", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_recycle TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_recycle",
                    expected: Expected::Rows {
                        columns: &[Column("wal_recycle", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_recycle')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_retrieve_retry_interval' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_retrieve_retry_interval",
                    expected: Expected::Rows {
                        columns: &[Column("wal_retrieve_retry_interval", TEXT)],
                        rows: &[
                            &[T("5s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_retrieve_retry_interval TO '5'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_retrieve_retry_interval" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_retrieve_retry_interval')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("5s")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_segment_size' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_segment_size",
                    expected: Expected::Rows {
                        columns: &[Column("wal_segment_size", TEXT)],
                        rows: &[
                            &[T("16MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_segment_size TO '16777216'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_segment_size" cannot be changed"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_segment_size')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("16MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_sender_timeout' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_sender_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("wal_sender_timeout", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_sender_timeout TO '100000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_sender_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("wal_sender_timeout", TEXT)],
                        rows: &[
                            &[T("100s")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_sender_timeout TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_sender_timeout",
                    expected: Expected::Rows {
                        columns: &[Column("wal_sender_timeout", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_sender_timeout')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1min")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_skip_threshold' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_skip_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("wal_skip_threshold", TEXT)],
                        rows: &[
                            &[T("2MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_skip_threshold TO '2000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_skip_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("wal_skip_threshold", TEXT)],
                        rows: &[
                            &[T("2000kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_skip_threshold TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW wal_skip_threshold",
                    expected: Expected::Rows {
                        columns: &[Column("wal_skip_threshold", TEXT)],
                        rows: &[
                            &[T("2MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_skip_threshold')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("2MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_sync_method' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_sync_method",
                    expected: Expected::Rows {
                        columns: &[Column("wal_sync_method", TEXT)],
                        rows: &[
                            &[T("open_datasync")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_sync_method TO 'open_datasync'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_sync_method" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_sync_method')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("open_datasync")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_writer_delay' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_writer_delay",
                    expected: Expected::Rows {
                        columns: &[Column("wal_writer_delay", TEXT)],
                        rows: &[
                            &[T("200ms")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_writer_delay TO '200'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_writer_delay" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_writer_delay')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("200ms")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'wal_writer_flush_after' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW wal_writer_flush_after",
                    expected: Expected::Rows {
                        columns: &[Column("wal_writer_flush_after", TEXT)],
                        rows: &[
                            &[T("1MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET wal_writer_flush_after TO '1000'",
                    expected: Expected::Error(Diagnostic { code: "55P02", message: r#"parameter "wal_writer_flush_after" cannot be changed now"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wal_writer_flush_after')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("1MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'work_mem' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("work_mem", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET work_mem TO '4000'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("work_mem", TEXT)],
                        rows: &[
                            &[T("4000kB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET work_mem TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW work_mem",
                    expected: Expected::Rows {
                        columns: &[Column("work_mem", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('work_mem')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("4MB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'xmlbinary' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW xmlbinary",
                    expected: Expected::Rows {
                        columns: &[Column("xmlbinary", TEXT)],
                        rows: &[
                            &[T("base64")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET xmlbinary TO 'hex'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW xmlbinary",
                    expected: Expected::Rows {
                        columns: &[Column("xmlbinary", TEXT)],
                        rows: &[
                            &[T("hex")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET xmlbinary TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW xmlbinary",
                    expected: Expected::Rows {
                        columns: &[Column("xmlbinary", TEXT)],
                        rows: &[
                            &[T("base64")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('xmlbinary')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("base64")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'xmloption' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW xmloption",
                    expected: Expected::Rows {
                        columns: &[Column("xmloption", TEXT)],
                        rows: &[
                            &[T("content")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET xmloption TO 'document'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW xmloption",
                    expected: Expected::Rows {
                        columns: &[Column("xmloption", TEXT)],
                        rows: &[
                            &[T("document")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET xmloption TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW xmloption",
                    expected: Expected::Rows {
                        columns: &[Column("xmloption", TEXT)],
                        rows: &[
                            &[T("content")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('xmloption')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("content")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set 'zero_damaged_pages' configuration variable",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW zero_damaged_pages",
                    expected: Expected::Rows {
                        columns: &[Column("zero_damaged_pages", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET zero_damaged_pages TO 'on'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW zero_damaged_pages",
                    expected: Expected::Rows {
                        columns: &[Column("zero_damaged_pages", TEXT)],
                        rows: &[
                            &[T("on")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET zero_damaged_pages TO DEFAULT",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW zero_damaged_pages",
                    expected: Expected::Rows {
                        columns: &[Column("zero_damaged_pages", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('zero_damaged_pages')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("off")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "settings with namespaces",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET myvar.var_value TO 'value'",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW myvar.var_value",
                    expected: Expected::Rows {
                        columns: &[Column("myvar.var_value", TEXT)],
                        rows: &[
                            &[T("value")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select current_setting('myvar.var_value')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("value")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select current_setting('unknown_var')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "unknown_var""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "show myvar.unknown_var",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "myvar.unknown_var""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "set myvar.var_value to (select 'a')",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "(""#, position: 24, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW myvar.var_value",
                    expected: Expected::Rows {
                        columns: &[Column("myvar.var_value", TEXT)],
                        rows: &[
                            &[T("value")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "set myvar.val2 to (select current_setting('myvar.var_value'))",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "(""#, position: 19, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW myvar.val2",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "myvar.val2""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}
