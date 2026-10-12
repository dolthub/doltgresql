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
fn test_parameters() {
    run_scripts(&[
        ScriptTest {
            name: "default_with_oids",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT default_with_oids;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "default_with_oids" does not exist"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET default_with_oids = false;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "DateStyle",
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
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
                    query: "SELECT timestamp '2001/02/04 04:05:06.789';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2001-02-04 04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'german';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("German, DMY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2001/02/04 04:05:06.789';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2001-02-04 04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'YMD';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("German, YMD")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'sQl';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("SQL, YMD")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2001/02/04 04:05:06.789';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2001-02-04 04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'postgreS';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("Postgres, YMD")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2001/02/04 04:05:06.789';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2001-02-04 04:05:06.789")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "RESET datestyle;",
                    expected: Expected::Tag("RESET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
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
                    query: "SET datestyle = 'unknown';",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid value for parameter "DateStyle": "unknown""#, detail: r#"Unrecognized key word: "unknown"."#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_partial_datestyle_rules() {
    run_scripts(&[
        ScriptTest {
            name: "Partial DateStyle values",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET datestyle = 'german';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("German, DMY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'YMD';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("German, YMD")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'sQl';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("SQL, YMD")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'postgreS';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("Postgres, YMD")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'dmy, german';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("German, DMY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = 'german, ymd';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("German, YMD")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET datestyle = default;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
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
                    query: "SET datestyle = 'iso';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
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
                    query: "SHOW DateStyle;",
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
                    query: "BEGIN;",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET LOCAL datestyle = 'sql';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("SQL, MDY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ROLLBACK;",
                    expected: Expected::Tag("ROLLBACK"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW DateStyle;",
                    expected: Expected::Rows {
                        columns: &[Column("DateStyle", TEXT)],
                        rows: &[
                            &[T("ISO, MDY")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
