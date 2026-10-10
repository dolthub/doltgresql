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
fn test_coercion() {
    run_scripts(&[
        ScriptTest {
            name: "Raw Literals",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 0",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 0.5",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", NUMERIC)],
                        rows: &[
                            &[T("0.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 0.50",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", NUMERIC)],
                        rows: &[
                            &[T("0.50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT -0.5",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", NUMERIC)],
                        rows: &[
                            &[T("-0.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 12345671297673227365.5123624235623456",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", NUMERIC)],
                        rows: &[
                            &[T("12345671297673227365.5123624235623456")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT -1",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 70000",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("70000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 5000000000",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT8)],
                        rows: &[
                            &[T("5000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 9223372036854775808",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", NUMERIC)],
                        rows: &[
                            &[T("9223372036854775808")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ''",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'test'",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '0'",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
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
            name: "Math Functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT abs(1)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs(1.5)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", NUMERIC)],
                        rows: &[
                            &[T("1.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs(5000000000)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", INT8)],
                        rows: &[
                            &[T("5000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs(9223372036854775808)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", NUMERIC)],
                        rows: &[
                            &[T("9223372036854775808")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs('1')",
                    expected: Expected::Rows {
                        columns: &[Column("abs", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs('1.5')",
                    expected: Expected::Rows {
                        columns: &[Column("abs", FLOAT8)],
                        rows: &[
                            &[T("1.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs('12345671297673227365.5123624235623456')",
                    expected: Expected::Rows {
                        columns: &[Column("abs", FLOAT8)],
                        rows: &[
                            &[T("1.2345671297673228e+19")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs('NaN'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs('Inf'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs('-infinity'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs('0'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", NUMERIC)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs('-0.50'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("abs", NUMERIC)],
                        rows: &[
                            &[T("0.50")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT factorial('1')",
                    expected: Expected::Rows {
                        columns: &[Column("factorial", NUMERIC)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT factorial('1.5')",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type bigint: "1.5""#, position: 18, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ceil('NaN'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("ceil", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ceil('Inf'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("ceil", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ceil('-infinity'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("ceil", NUMERIC)],
                        rows: &[
                            &[T("-Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT floor('NaN'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("floor", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT floor('Inf'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("floor", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT floor('-infinity'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("floor", NUMERIC)],
                        rows: &[
                            &[T("-Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ln('NaN'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("ln", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ln('Inf'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("ln", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ln('-infinity'::numeric)",
                    expected: Expected::Error(Diagnostic { code: "2201E", message: "cannot take logarithm of a negative number", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT log('NaN'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("log", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT log('Inf'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("log", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT log('-infinity'::numeric)",
                    expected: Expected::Error(Diagnostic { code: "2201E", message: "cannot take logarithm of a negative number", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT min_scale('NaN'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("min_scale", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT min_scale('Inf'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("min_scale", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT min_scale('-infinity'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("min_scale", INT4)],
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
    ]);
}
