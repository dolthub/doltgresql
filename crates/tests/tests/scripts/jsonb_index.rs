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
fn test_json_b_index_sort_consistency() {
    run_scripts(&[
        ScriptTest {
            name: "JSONB index sort order matches < and > operators",
            set_up_script: &[
                "CREATE TABLE jtest (id SERIAL PRIMARY KEY, val JSONB NOT NULL)",
                "CREATE INDEX jtest_val_idx ON jtest (val)",
                r#"INSERT INTO jtest (val) VALUES
					('null'),
					('false'),
					('true'),
					('-1'),
					('0'),
					('1'),
					('2'),
					('3.14'),
					('42'),
					('100'),
					('9999'),
					('"a"'),
					('"b"'),
					('"z"'),
					('"ab"'),
					('"abc"'),
					('"foo"'),
					('"hello"'),
					('"hello world"'),
					('"longer string value"'),
					('[]'),
					('[1]'),
					('[1,2]'),
					('[1,2,3]'),
					('["a"]'),
					('[null]'),
					('[false]'),
					('[true]'),
					('["a","b","c"]'),
					('[[1,2],[3,4]]'),
					('{}'),
					('{"a":1}'),
					('{"b":2}'),
					('{"aa":1}'),
					('{"a":1,"b":2}'),
					('{"a":1,"b":2,"c":3}'),
					('{"x":{"y":1}}'),
					('{"name":"test","value":42}'),
					('{"a":{"b":{"c":1}}}'),
					('{"z":null}')"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT COUNT(*) FROM jtest a
					        JOIN jtest b ON a.val < b.val
					        JOIN jtest c ON b.val < c.val
					        WHERE NOT (a.val < c.val)"#,
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
                    query: r#"SELECT COUNT(*) FROM jtest a
					        JOIN jtest b ON a.val > b.val
					        JOIN jtest c ON b.val > c.val
					        WHERE NOT (a.val > c.val)"#,
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
                    query: r#"SELECT COUNT(*) FROM (
					          SELECT val,
					                 LAG(val) OVER (ORDER BY val) AS prev_val
					          FROM jtest
					        ) t
					        WHERE prev_val IS NOT NULL AND NOT (prev_val < val)"#,
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
                    query: r#"SELECT COUNT(*) FROM (
					          SELECT val,
					                 LEAD(val) OVER (ORDER BY val) AS next_val
					          FROM jtest
					        ) t
					        WHERE next_val IS NOT NULL AND NOT (val < next_val)"#,
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
                    query: r#"SELECT COUNT(*) FROM jtest a
					        JOIN jtest b ON a.val < b.val
					        WHERE a.val > b.val"#,
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
                    query: r#"SELECT COUNT(*) FROM jtest a
					        JOIN jtest b ON a.id < b.id
					        WHERE a.val <> b.val
					          AND NOT (a.val < b.val)
					          AND NOT (b.val < a.val)"#,
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
    ]);
}

#[test]
fn test_json_b_pairwise_less_than() {
    run_scripts(&[
        // Changed from the Go test: json has no btree operator class or ordering in Postgres; the test is about jsonb.
        ScriptTest {
            name: "JSONB pairwise less-than along lexical order",
            set_up_script: &[
                "CREATE TABLE jorder (val JSONB NOT NULL)",
                "CREATE INDEX jorder_val_idx ON jorder (val)",
                r#"INSERT INTO jorder (val) VALUES ('null'), ('-1'), ('0'), ('1'), ('2'), ('3.14'), ('42'), ('100'), ('9999'), ('"a"'), ('"ab"'), ('"abc"'), ('"b"'), ('"foo"'), ('"hello"'), ('"hello world"'), ('"longer string value"'), ('"z"'), ('{}'), ('{"z":null}'), ('{"x":{"y":1}}'), ('{"name":"test","value":42}'), ('{"b":2}'), ('{"aa":1}'), ('{"a":1}'), ('{"a":1,"b":2}'), ('{"a":1,"b":2,"c":3}'), ('{"a":{"b":{"c":1}}}'), ('[]'), ('[null]'), ('[1]'), ('[1,2]'), ('[1,2,3]'), ('["a"]'), ('["a","b","c"]'), ('[[1,2],[3,4]]'), ('[false]'), ('[true]'), ('false'), ('true')"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'null'::jsonb < '-1'::jsonb",
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
                    query: "SELECT '-1'::jsonb < '0'::jsonb",
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
                    query: "SELECT '0'::jsonb < '1'::jsonb",
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
                    query: "SELECT '1'::jsonb < '2'::jsonb",
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
                    query: "SELECT '2'::jsonb < '3.14'::jsonb",
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
                    query: "SELECT '3.14'::jsonb < '42'::jsonb",
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
                    query: "SELECT '42'::jsonb < '100'::jsonb",
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
                    query: "SELECT '100'::jsonb < '9999'::jsonb",
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
                    query: r#"SELECT '9999'::jsonb < '"a"'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"a"'::jsonb < '"ab"'::jsonb"#,
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
                    query: r#"SELECT '"ab"'::jsonb < '"abc"'::jsonb"#,
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
                    query: r#"SELECT '"abc"'::jsonb < '"b"'::jsonb"#,
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
                    query: r#"SELECT '"b"'::jsonb < '"foo"'::jsonb"#,
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
                    query: r#"SELECT '"foo"'::jsonb < '"hello"'::jsonb"#,
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
                    query: r#"SELECT '"hello"'::jsonb < '"hello world"'::jsonb"#,
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
                    query: r#"SELECT '"hello world"'::jsonb < '"longer string value"'::jsonb"#,
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
                    query: r#"SELECT '"longer string value"'::jsonb < '"z"'::jsonb"#,
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
                    query: r#"SELECT '"z"'::jsonb < '{}'::jsonb"#,
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
                    query: r#"SELECT '{}'::jsonb < '{"z":null}'::jsonb"#,
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
                    query: r#"SELECT '{"z":null}'::jsonb < '{"x":{"y":1}}'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"x":{"y":1}}'::jsonb < '{"name":"test","value":42}'::jsonb"#,
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
                    query: r#"SELECT '{"name":"test","value":42}'::jsonb < '{"b":2}'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"b":2}'::jsonb < '{"aa":1}'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"aa":1}'::jsonb < '{"a":1}'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1}'::jsonb < '{"a":1,"b":2}'::jsonb"#,
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
                    query: r#"SELECT '{"a":1,"b":2}'::jsonb < '{"a":1,"b":2,"c":3}'::jsonb"#,
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
                    query: r#"SELECT '{"a":1,"b":2,"c":3}'::jsonb < '{"a":{"b":{"c":1}}}'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":{"c":1}}}'::jsonb < '[]'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[]'::jsonb < '[null]'::jsonb",
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
                    query: "SELECT '[null]'::jsonb < '[1]'::jsonb",
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
                    query: "SELECT '[1]'::jsonb < '[1,2]'::jsonb",
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
                    query: "SELECT '[1,2]'::jsonb < '[1,2,3]'::jsonb",
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
                    query: r#"SELECT '[1,2,3]'::jsonb < '["a"]'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '["a"]'::jsonb < '["a","b","c"]'::jsonb"#,
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
                    query: r#"SELECT '["a","b","c"]'::jsonb < '[[1,2],[3,4]]'::jsonb"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[[1,2],[3,4]]'::jsonb < '[false]'::jsonb",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[false]'::jsonb < '[true]'::jsonb",
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
                    query: "SELECT '[true]'::jsonb < 'false'::jsonb",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'false'::jsonb < 'true'::jsonb",
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
                    query: "SELECT val FROM jorder ORDER BY val",
                    expected: Expected::Rows {
                        columns: &[Column("val", JSONB)],
                        rows: &[
                            &[T("[]")],
                            &[T("null")],
                            &[T(r#""a""#)],
                            &[T(r#""ab""#)],
                            &[T(r#""abc""#)],
                            &[T(r#""b""#)],
                            &[T(r#""foo""#)],
                            &[T(r#""hello""#)],
                            &[T(r#""hello world""#)],
                            &[T(r#""longer string value""#)],
                            &[T(r#""z""#)],
                            &[T("-1")],
                            &[T("0")],
                            &[T("1")],
                            &[T("2")],
                            &[T("3.14")],
                            &[T("42")],
                            &[T("100")],
                            &[T("9999")],
                            &[T("false")],
                            &[T("true")],
                            &[T("[null]")],
                            &[T(r#"["a"]"#)],
                            &[T("[1]")],
                            &[T("[false]")],
                            &[T("[true]")],
                            &[T("[1, 2]")],
                            &[T("[[1, 2], [3, 4]]")],
                            &[T(r#"["a", "b", "c"]"#)],
                            &[T("[1, 2, 3]")],
                            &[T("{}")],
                            &[T(r#"{"a": 1}"#)],
                            &[T(r#"{"a": {"b": {"c": 1}}}"#)],
                            &[T(r#"{"aa": 1}"#)],
                            &[T(r#"{"b": 2}"#)],
                            &[T(r#"{"x": {"y": 1}}"#)],
                            &[T(r#"{"z": null}"#)],
                            &[T(r#"{"a": 1, "b": 2}"#)],
                            &[T(r#"{"name": "test", "value": 42}"#)],
                            &[T(r#"{"a": 1, "b": 2, "c": 3}"#)],
                        ],
                        tag: "SELECT 40",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
