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
fn test_json_array_element() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb_array_element returns array element",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::jsonb -> 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::jsonb -> 2;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::jsonb -> 5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::jsonb -> -1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::jsonb -> -3;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::jsonb -> -5;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1}'::jsonb -> 0;"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '42'::jsonb -> 0;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '[{"a":1},{"b":2}]'::jsonb -> 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T(r#"{"b": 2}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "json_array_element returns array element",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::json -> 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::json -> -1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::json -> 99;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
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
            name: "jsonb_array_element_text returns text representation",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '["alpha","beta"]'::jsonb ->> 0;"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("alpha")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::jsonb ->> -1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[10,20,30]'::jsonb ->> 99;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
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

#[test]
fn test_json_array_length() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb_array_length",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length('[]'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_array_length", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length('[1,2,3]'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_array_length", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length('[[1,2],[3,4],[5]]'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_array_length", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length('[null,null]'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_array_length", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length(null::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_array_length", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(jsonb_array_length('[]'::jsonb));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length('{}'::jsonb);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot get array length of a non-array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length('1'::jsonb);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot get array length of a scalar", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_array_length('"str"'::jsonb);"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot get array length of a scalar", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length('null'::jsonb);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot get array length of a scalar", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "json_array_length",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT json_array_length('[1,2]'::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_array_length", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_array_length('[]'::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_array_length", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_array_length(null::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_array_length", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_array_length('{"a":1}'::json);"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot get array length of a non-array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_array_length('true'::json);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot get array length of a scalar", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_json_exists() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb_exists (?) tests key/element presence",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":2}'::jsonb ? 'a';"#,
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
                    query: r#"SELECT '{"a":1,"b":2}'::jsonb ? 'z';"#,
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
                    query: r#"SELECT '{"a":null}'::jsonb ? 'a';"#,
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
                    query: r#"SELECT '["alpha","beta","gamma"]'::jsonb ? 'beta';"#,
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
                    query: r#"SELECT '["alpha","beta"]'::jsonb ? 'gamma';"#,
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
                    query: "SELECT '[1,2,3]'::jsonb ? '1';",
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
                    query: r#"SELECT '"hello"'::jsonb ? 'hello';"#,
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
                    query: r#"SELECT '"hello"'::jsonb ? 'world';"#,
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
                    query: "SELECT '42'::jsonb ? '42';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
            name: "jsonb_exists_any (?|) tests presence of any key",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":2}'::jsonb ?| ARRAY['x','b'];"#,
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
                    query: r#"SELECT '{"a":1,"b":2}'::jsonb ?| ARRAY['x','y'];"#,
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
                    query: r#"SELECT '["a","b","c"]'::jsonb ?| ARRAY['x','b'];"#,
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
                    query: r#"SELECT '["a","b","c"]'::jsonb ?| ARRAY['x','y'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
            name: "jsonb_exists_all (?&) tests presence of all keys",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":2,"c":3}'::jsonb ?& ARRAY['a','b'];"#,
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
                    query: r#"SELECT '{"a":1,"b":2}'::jsonb ?& ARRAY['a','missing'];"#,
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
                    query: r#"SELECT '["a","b","c"]'::jsonb ?& ARRAY['a','b'];"#,
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
                    query: r#"SELECT '["a","b"]'::jsonb ?& ARRAY['a','missing'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
    ]);
}

#[test]
fn test_json_extract_path() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb_extract_path follows mixed key/index paths",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":{"c":1}}}'::jsonb #> '{a,b,c}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":[10,20,30]}'::jsonb #> '{a,1}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":[10,20,30]}'::jsonb #> '{a,-1}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":[10,20]}'::jsonb #> '{a,not-an-int}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":1}}'::jsonb #> '{a,missing,c}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1}'::jsonb #> '{a,b}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
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
            name: "jsonb_extract_path_text renders the leaf as text",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":"hello"}}'::jsonb #>> '{a,b}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":{"c":1}}}'::jsonb #>> '{a,b}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T(r#"{"c": 1}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":[1,2,3]}'::jsonb #>> '{a,2}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1}'::jsonb #>> '{missing}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
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
            name: "json_extract_path follows mixed key/index paths",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":[10,20]}}'::json #> '{a,b,0}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1}'::json #> '{missing}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
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
            name: "jsonb_extract_path with multi-element text-array paths",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::jsonb #> ARRAY['a','b'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":{"c":[10,20]}}}'::jsonb #> ARRAY['a','b','c','1'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::jsonb #>> ARRAY['a','b'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"NULL":7}'::jsonb #> ARRAY['NULL'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"NULL":7}'::jsonb #> '{"NULL"}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb_extract_path returns NULL for NULL path elements",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::jsonb #> ARRAY['a',NULL];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::jsonb #> ARRAY[NULL,'b'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::jsonb #> ARRAY['a',NULL,'b'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::jsonb #> '{a,NULL,b}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"NULL":7}'::jsonb #> '{NULL}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::jsonb #>> ARRAY['a',NULL];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::jsonb #> NULL::text[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
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
            name: "json_extract_path with text-array paths and NULL elements",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::json #> ARRAY['a','b'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::json #>> ARRAY['a','b'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::json #> ARRAY['a',NULL];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":42}}'::json #>> ARRAY[NULL,'b'];"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
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

#[test]
fn test_json_inspection_stored_values() {
    run_scripts(&[
        ScriptTest {
            name: "inspection functions on large stored documents (>4 KB)",
            set_up_script: &[
                "CREATE TABLE bigdoc (id INT PRIMARY KEY, doc JSONB)",
                r#"INSERT INTO bigdoc (id, doc) VALUES (1, '{"k_0000":{"name":"value_0000","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":0},"k_0001":{"name":"value_0001","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":1},"k_0002":{"name":"value_0002","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":2},"k_0003":{"name":"value_0003","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":3},"k_0004":{"name":"value_0004","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":4},"k_0005":{"name":"value_0005","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":5},"k_0006":{"name":"value_0006","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":6},"k_0007":{"name":"value_0007","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":7},"k_0008":{"name":"value_0008","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":8},"k_0009":{"name":"value_0009","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":9},"k_0010":{"name":"value_0010","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":10},"k_0011":{"name":"value_0011","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":11},"k_0012":{"name":"value_0012","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":12},"k_0013":{"name":"value_0013","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":13},"k_0014":{"name":"value_0014","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":14},"k_0015":{"name":"value_0015","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":15},"k_0016":{"name":"value_0016","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":16},"k_0017":{"name":"value_0017","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":17},"k_0018":{"name":"value_0018","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":18},"k_0019":{"name":"value_0019","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":19},"k_0020":{"name":"value_0020","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":20},"k_0021":{"name":"value_0021","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":21},"k_0022":{"name":"value_0022","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":22},"k_0023":{"name":"value_0023","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":23},"k_0024":{"name":"value_0024","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":24},"k_0025":{"name":"value_0025","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":25},"k_0026":{"name":"value_0026","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":26},"k_0027":{"name":"value_0027","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":27},"k_0028":{"name":"value_0028","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":28},"k_0029":{"name":"value_0029","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":29},"k_0030":{"name":"value_0030","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":30},"k_0031":{"name":"value_0031","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":31},"k_0032":{"name":"value_0032","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":32},"k_0033":{"name":"value_0033","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":33},"k_0034":{"name":"value_0034","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":34},"k_0035":{"name":"value_0035","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":35},"k_0036":{"name":"value_0036","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":36},"k_0037":{"name":"value_0037","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":37},"k_0038":{"name":"value_0038","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":38},"k_0039":{"name":"value_0039","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":39},"k_0040":{"name":"value_0040","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":40},"k_0041":{"name":"value_0041","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":41},"k_0042":{"name":"value_0042","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":42},"k_0043":{"name":"value_0043","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":43},"k_0044":{"name":"value_0044","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":44},"k_0045":{"name":"value_0045","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":45},"k_0046":{"name":"value_0046","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":46},"k_0047":{"name":"value_0047","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":47},"k_0048":{"name":"value_0048","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":48},"k_0049":{"name":"value_0049","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":49},"k_0050":{"name":"value_0050","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":50},"k_0051":{"name":"value_0051","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":51},"k_0052":{"name":"value_0052","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":52},"k_0053":{"name":"value_0053","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":53},"k_0054":{"name":"value_0054","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":54},"k_0055":{"name":"value_0055","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":55},"k_0056":{"name":"value_0056","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":56},"k_0057":{"name":"value_0057","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":57},"k_0058":{"name":"value_0058","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":58},"k_0059":{"name":"value_0059","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":59},"k_0060":{"name":"value_0060","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":60},"k_0061":{"name":"value_0061","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":61},"k_0062":{"name":"value_0062","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":62},"k_0063":{"name":"value_0063","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":63},"k_0064":{"name":"value_0064","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":64},"k_0065":{"name":"value_0065","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":65},"k_0066":{"name":"value_0066","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":66},"k_0067":{"name":"value_0067","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":67},"k_0068":{"name":"value_0068","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":68},"k_0069":{"name":"value_0069","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":69},"k_0070":{"name":"value_0070","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":70},"k_0071":{"name":"value_0071","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":71},"k_0072":{"name":"value_0072","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":72},"k_0073":{"name":"value_0073","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":73},"k_0074":{"name":"value_0074","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":74},"k_0075":{"name":"value_0075","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":75},"k_0076":{"name":"value_0076","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":76},"k_0077":{"name":"value_0077","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":77},"k_0078":{"name":"value_0078","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":78},"k_0079":{"name":"value_0079","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":79},"k_0080":{"name":"value_0080","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":80},"k_0081":{"name":"value_0081","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":81},"k_0082":{"name":"value_0082","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":82},"k_0083":{"name":"value_0083","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":83},"k_0084":{"name":"value_0084","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":84},"k_0085":{"name":"value_0085","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":85},"k_0086":{"name":"value_0086","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":86},"k_0087":{"name":"value_0087","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":87},"k_0088":{"name":"value_0088","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":88},"k_0089":{"name":"value_0089","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":89},"k_0090":{"name":"value_0090","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":90},"k_0091":{"name":"value_0091","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":91},"k_0092":{"name":"value_0092","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":92},"k_0093":{"name":"value_0093","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":93},"k_0094":{"name":"value_0094","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":94},"k_0095":{"name":"value_0095","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":95},"k_0096":{"name":"value_0096","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":96},"k_0097":{"name":"value_0097","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":97},"k_0098":{"name":"value_0098","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":98},"k_0099":{"name":"value_0099","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":99}}'::jsonb)"#,
                r#"INSERT INTO bigdoc (id, doc) VALUES (2, '[{"id":0,"label":"row_0000","payload":["a","b","c","d","e"]},{"id":1,"label":"row_0001","payload":["a","b","c","d","e"]},{"id":2,"label":"row_0002","payload":["a","b","c","d","e"]},{"id":3,"label":"row_0003","payload":["a","b","c","d","e"]},{"id":4,"label":"row_0004","payload":["a","b","c","d","e"]},{"id":5,"label":"row_0005","payload":["a","b","c","d","e"]},{"id":6,"label":"row_0006","payload":["a","b","c","d","e"]},{"id":7,"label":"row_0007","payload":["a","b","c","d","e"]},{"id":8,"label":"row_0008","payload":["a","b","c","d","e"]},{"id":9,"label":"row_0009","payload":["a","b","c","d","e"]},{"id":10,"label":"row_0010","payload":["a","b","c","d","e"]},{"id":11,"label":"row_0011","payload":["a","b","c","d","e"]},{"id":12,"label":"row_0012","payload":["a","b","c","d","e"]},{"id":13,"label":"row_0013","payload":["a","b","c","d","e"]},{"id":14,"label":"row_0014","payload":["a","b","c","d","e"]},{"id":15,"label":"row_0015","payload":["a","b","c","d","e"]},{"id":16,"label":"row_0016","payload":["a","b","c","d","e"]},{"id":17,"label":"row_0017","payload":["a","b","c","d","e"]},{"id":18,"label":"row_0018","payload":["a","b","c","d","e"]},{"id":19,"label":"row_0019","payload":["a","b","c","d","e"]},{"id":20,"label":"row_0020","payload":["a","b","c","d","e"]},{"id":21,"label":"row_0021","payload":["a","b","c","d","e"]},{"id":22,"label":"row_0022","payload":["a","b","c","d","e"]},{"id":23,"label":"row_0023","payload":["a","b","c","d","e"]},{"id":24,"label":"row_0024","payload":["a","b","c","d","e"]},{"id":25,"label":"row_0025","payload":["a","b","c","d","e"]},{"id":26,"label":"row_0026","payload":["a","b","c","d","e"]},{"id":27,"label":"row_0027","payload":["a","b","c","d","e"]},{"id":28,"label":"row_0028","payload":["a","b","c","d","e"]},{"id":29,"label":"row_0029","payload":["a","b","c","d","e"]},{"id":30,"label":"row_0030","payload":["a","b","c","d","e"]},{"id":31,"label":"row_0031","payload":["a","b","c","d","e"]},{"id":32,"label":"row_0032","payload":["a","b","c","d","e"]},{"id":33,"label":"row_0033","payload":["a","b","c","d","e"]},{"id":34,"label":"row_0034","payload":["a","b","c","d","e"]},{"id":35,"label":"row_0035","payload":["a","b","c","d","e"]},{"id":36,"label":"row_0036","payload":["a","b","c","d","e"]},{"id":37,"label":"row_0037","payload":["a","b","c","d","e"]},{"id":38,"label":"row_0038","payload":["a","b","c","d","e"]},{"id":39,"label":"row_0039","payload":["a","b","c","d","e"]},{"id":40,"label":"row_0040","payload":["a","b","c","d","e"]},{"id":41,"label":"row_0041","payload":["a","b","c","d","e"]},{"id":42,"label":"row_0042","payload":["a","b","c","d","e"]},{"id":43,"label":"row_0043","payload":["a","b","c","d","e"]},{"id":44,"label":"row_0044","payload":["a","b","c","d","e"]},{"id":45,"label":"row_0045","payload":["a","b","c","d","e"]},{"id":46,"label":"row_0046","payload":["a","b","c","d","e"]},{"id":47,"label":"row_0047","payload":["a","b","c","d","e"]},{"id":48,"label":"row_0048","payload":["a","b","c","d","e"]},{"id":49,"label":"row_0049","payload":["a","b","c","d","e"]},{"id":50,"label":"row_0050","payload":["a","b","c","d","e"]},{"id":51,"label":"row_0051","payload":["a","b","c","d","e"]},{"id":52,"label":"row_0052","payload":["a","b","c","d","e"]},{"id":53,"label":"row_0053","payload":["a","b","c","d","e"]},{"id":54,"label":"row_0054","payload":["a","b","c","d","e"]},{"id":55,"label":"row_0055","payload":["a","b","c","d","e"]},{"id":56,"label":"row_0056","payload":["a","b","c","d","e"]},{"id":57,"label":"row_0057","payload":["a","b","c","d","e"]},{"id":58,"label":"row_0058","payload":["a","b","c","d","e"]},{"id":59,"label":"row_0059","payload":["a","b","c","d","e"]},{"id":60,"label":"row_0060","payload":["a","b","c","d","e"]},{"id":61,"label":"row_0061","payload":["a","b","c","d","e"]},{"id":62,"label":"row_0062","payload":["a","b","c","d","e"]},{"id":63,"label":"row_0063","payload":["a","b","c","d","e"]},{"id":64,"label":"row_0064","payload":["a","b","c","d","e"]},{"id":65,"label":"row_0065","payload":["a","b","c","d","e"]},{"id":66,"label":"row_0066","payload":["a","b","c","d","e"]},{"id":67,"label":"row_0067","payload":["a","b","c","d","e"]},{"id":68,"label":"row_0068","payload":["a","b","c","d","e"]},{"id":69,"label":"row_0069","payload":["a","b","c","d","e"]},{"id":70,"label":"row_0070","payload":["a","b","c","d","e"]},{"id":71,"label":"row_0071","payload":["a","b","c","d","e"]},{"id":72,"label":"row_0072","payload":["a","b","c","d","e"]},{"id":73,"label":"row_0073","payload":["a","b","c","d","e"]},{"id":74,"label":"row_0074","payload":["a","b","c","d","e"]},{"id":75,"label":"row_0075","payload":["a","b","c","d","e"]},{"id":76,"label":"row_0076","payload":["a","b","c","d","e"]},{"id":77,"label":"row_0077","payload":["a","b","c","d","e"]},{"id":78,"label":"row_0078","payload":["a","b","c","d","e"]},{"id":79,"label":"row_0079","payload":["a","b","c","d","e"]}]'::jsonb)"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT length(doc::text) > 4096 FROM bigdoc ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                            &[T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof(doc) FROM bigdoc ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("object")],
                            &[T("array")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_array_length(doc) FROM bigdoc WHERE id = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_array_length", INT4)],
                        rows: &[
                            &[T("80")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM jsonb_object_keys((SELECT doc FROM bigdoc WHERE id = 1)) AS k;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_strip_nulls(doc) = doc FROM bigdoc ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                            &[T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "inspection functions on a NULL column value",
            set_up_script: &[
                "CREATE TABLE nulldoc (id INT PRIMARY KEY, doc JSONB, j JSON)",
                "INSERT INTO nulldoc VALUES (1, NULL, NULL)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_typeof(doc), json_typeof(j), jsonb_array_length(doc),
					               jsonb_strip_nulls(doc), to_jsonb(doc) FROM nulldoc;"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT), Column("json_typeof", TEXT), Column("jsonb_array_length", INT4), Column("jsonb_strip_nulls", JSONB), Column("to_jsonb", JSONB)],
                        rows: &[
                            &[Null, Null, Null, Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_object_keys(doc) FROM nulldoc;",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_object_keys", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_json_large_document_access() {
    run_scripts(&[
        ScriptTest {
            name: "JSONB operators on large stored object (>4 KB)",
            set_up_script: &[
                "CREATE TABLE bigobj (id INT PRIMARY KEY, doc JSONB)",
                r#"INSERT INTO bigobj (id, doc) VALUES (1, '{"k_0000":{"name":"value_0000","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":0},"k_0001":{"name":"value_0001","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":1},"k_0002":{"name":"value_0002","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":2},"k_0003":{"name":"value_0003","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":3},"k_0004":{"name":"value_0004","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":4},"k_0005":{"name":"value_0005","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":5},"k_0006":{"name":"value_0006","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":6},"k_0007":{"name":"value_0007","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":7},"k_0008":{"name":"value_0008","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":8},"k_0009":{"name":"value_0009","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":9},"k_0010":{"name":"value_0010","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":10},"k_0011":{"name":"value_0011","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":11},"k_0012":{"name":"value_0012","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":12},"k_0013":{"name":"value_0013","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":13},"k_0014":{"name":"value_0014","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":14},"k_0015":{"name":"value_0015","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":15},"k_0016":{"name":"value_0016","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":16},"k_0017":{"name":"value_0017","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":17},"k_0018":{"name":"value_0018","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":18},"k_0019":{"name":"value_0019","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":19},"k_0020":{"name":"value_0020","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":20},"k_0021":{"name":"value_0021","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":21},"k_0022":{"name":"value_0022","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":22},"k_0023":{"name":"value_0023","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":23},"k_0024":{"name":"value_0024","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":24},"k_0025":{"name":"value_0025","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":25},"k_0026":{"name":"value_0026","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":26},"k_0027":{"name":"value_0027","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":27},"k_0028":{"name":"value_0028","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":28},"k_0029":{"name":"value_0029","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":29},"k_0030":{"name":"value_0030","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":30},"k_0031":{"name":"value_0031","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":31},"k_0032":{"name":"value_0032","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":32},"k_0033":{"name":"value_0033","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":33},"k_0034":{"name":"value_0034","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":34},"k_0035":{"name":"value_0035","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":35},"k_0036":{"name":"value_0036","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":36},"k_0037":{"name":"value_0037","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":37},"k_0038":{"name":"value_0038","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":38},"k_0039":{"name":"value_0039","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":39},"k_0040":{"name":"value_0040","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":40},"k_0041":{"name":"value_0041","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":41},"k_0042":{"name":"value_0042","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":42},"k_0043":{"name":"value_0043","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":43},"k_0044":{"name":"value_0044","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":44},"k_0045":{"name":"value_0045","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":45},"k_0046":{"name":"value_0046","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":46},"k_0047":{"name":"value_0047","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":47},"k_0048":{"name":"value_0048","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":48},"k_0049":{"name":"value_0049","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":49},"k_0050":{"name":"value_0050","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":50},"k_0051":{"name":"value_0051","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":51},"k_0052":{"name":"value_0052","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":52},"k_0053":{"name":"value_0053","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":53},"k_0054":{"name":"value_0054","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":54},"k_0055":{"name":"value_0055","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":55},"k_0056":{"name":"value_0056","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":56},"k_0057":{"name":"value_0057","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":57},"k_0058":{"name":"value_0058","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":58},"k_0059":{"name":"value_0059","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":59},"k_0060":{"name":"value_0060","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":60},"k_0061":{"name":"value_0061","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":61},"k_0062":{"name":"value_0062","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":62},"k_0063":{"name":"value_0063","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":63},"k_0064":{"name":"value_0064","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":64},"k_0065":{"name":"value_0065","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":65},"k_0066":{"name":"value_0066","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":66},"k_0067":{"name":"value_0067","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":67},"k_0068":{"name":"value_0068","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":68},"k_0069":{"name":"value_0069","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":69},"k_0070":{"name":"value_0070","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":70},"k_0071":{"name":"value_0071","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":71},"k_0072":{"name":"value_0072","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":72},"k_0073":{"name":"value_0073","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":73},"k_0074":{"name":"value_0074","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":74},"k_0075":{"name":"value_0075","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":75},"k_0076":{"name":"value_0076","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":76},"k_0077":{"name":"value_0077","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":77},"k_0078":{"name":"value_0078","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":78},"k_0079":{"name":"value_0079","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":79},"k_0080":{"name":"value_0080","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":80},"k_0081":{"name":"value_0081","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":81},"k_0082":{"name":"value_0082","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":82},"k_0083":{"name":"value_0083","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":83},"k_0084":{"name":"value_0084","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":84},"k_0085":{"name":"value_0085","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":85},"k_0086":{"name":"value_0086","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":86},"k_0087":{"name":"value_0087","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":87},"k_0088":{"name":"value_0088","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":88},"k_0089":{"name":"value_0089","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":89},"k_0090":{"name":"value_0090","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":90},"k_0091":{"name":"value_0091","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":91},"k_0092":{"name":"value_0092","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":92},"k_0093":{"name":"value_0093","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":93},"k_0094":{"name":"value_0094","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":94},"k_0095":{"name":"value_0095","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":95},"k_0096":{"name":"value_0096","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":96},"k_0097":{"name":"value_0097","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":97},"k_0098":{"name":"value_0098","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":98},"k_0099":{"name":"value_0099","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":99}}'::jsonb)"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT length(doc::text) > 4096 FROM bigobj WHERE id = 1;",
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
                    query: "SELECT doc -> 'k_0037' ->> 'name' FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("value_0037")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc -> 'k_0000' ->> 'name' FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("value_0000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc -> 'k_0099' ->> 'name' FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("value_0099")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc -> 'no_such_key' FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc -> 'k_0042' ->> 'n' FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc #>> '{k_0010, tags, 2}' FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("tag-c")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc #>> '{k_0050, tags, -1}' FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("tag-e")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc #> '{k_0001, missing}' FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc ? 'k_0017' FROM bigobj WHERE id = 1;",
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
                    query: "SELECT doc ? 'no_such_key' FROM bigobj WHERE id = 1;",
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
                    query: "SELECT doc ?| ARRAY['no_such_key', 'k_0005'] FROM bigobj WHERE id = 1;",
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
                    query: "SELECT doc ?| ARRAY['nope_1', 'nope_2'] FROM bigobj WHERE id = 1;",
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
                    query: "SELECT doc ?& ARRAY['k_0001', 'k_0099'] FROM bigobj WHERE id = 1;",
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
                    query: "SELECT doc ?& ARRAY['k_0001', 'no_such_key'] FROM bigobj WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
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
            name: "JSONB operators on large stored array (>4 KB)",
            set_up_script: &[
                "CREATE TABLE bigarr (id INT PRIMARY KEY, doc JSONB)",
                r#"INSERT INTO bigarr (id, doc) VALUES (1, '[{"id":0,"label":"row_0000","payload":["a","b","c","d","e"]},{"id":1,"label":"row_0001","payload":["a","b","c","d","e"]},{"id":2,"label":"row_0002","payload":["a","b","c","d","e"]},{"id":3,"label":"row_0003","payload":["a","b","c","d","e"]},{"id":4,"label":"row_0004","payload":["a","b","c","d","e"]},{"id":5,"label":"row_0005","payload":["a","b","c","d","e"]},{"id":6,"label":"row_0006","payload":["a","b","c","d","e"]},{"id":7,"label":"row_0007","payload":["a","b","c","d","e"]},{"id":8,"label":"row_0008","payload":["a","b","c","d","e"]},{"id":9,"label":"row_0009","payload":["a","b","c","d","e"]},{"id":10,"label":"row_0010","payload":["a","b","c","d","e"]},{"id":11,"label":"row_0011","payload":["a","b","c","d","e"]},{"id":12,"label":"row_0012","payload":["a","b","c","d","e"]},{"id":13,"label":"row_0013","payload":["a","b","c","d","e"]},{"id":14,"label":"row_0014","payload":["a","b","c","d","e"]},{"id":15,"label":"row_0015","payload":["a","b","c","d","e"]},{"id":16,"label":"row_0016","payload":["a","b","c","d","e"]},{"id":17,"label":"row_0017","payload":["a","b","c","d","e"]},{"id":18,"label":"row_0018","payload":["a","b","c","d","e"]},{"id":19,"label":"row_0019","payload":["a","b","c","d","e"]},{"id":20,"label":"row_0020","payload":["a","b","c","d","e"]},{"id":21,"label":"row_0021","payload":["a","b","c","d","e"]},{"id":22,"label":"row_0022","payload":["a","b","c","d","e"]},{"id":23,"label":"row_0023","payload":["a","b","c","d","e"]},{"id":24,"label":"row_0024","payload":["a","b","c","d","e"]},{"id":25,"label":"row_0025","payload":["a","b","c","d","e"]},{"id":26,"label":"row_0026","payload":["a","b","c","d","e"]},{"id":27,"label":"row_0027","payload":["a","b","c","d","e"]},{"id":28,"label":"row_0028","payload":["a","b","c","d","e"]},{"id":29,"label":"row_0029","payload":["a","b","c","d","e"]},{"id":30,"label":"row_0030","payload":["a","b","c","d","e"]},{"id":31,"label":"row_0031","payload":["a","b","c","d","e"]},{"id":32,"label":"row_0032","payload":["a","b","c","d","e"]},{"id":33,"label":"row_0033","payload":["a","b","c","d","e"]},{"id":34,"label":"row_0034","payload":["a","b","c","d","e"]},{"id":35,"label":"row_0035","payload":["a","b","c","d","e"]},{"id":36,"label":"row_0036","payload":["a","b","c","d","e"]},{"id":37,"label":"row_0037","payload":["a","b","c","d","e"]},{"id":38,"label":"row_0038","payload":["a","b","c","d","e"]},{"id":39,"label":"row_0039","payload":["a","b","c","d","e"]},{"id":40,"label":"row_0040","payload":["a","b","c","d","e"]},{"id":41,"label":"row_0041","payload":["a","b","c","d","e"]},{"id":42,"label":"row_0042","payload":["a","b","c","d","e"]},{"id":43,"label":"row_0043","payload":["a","b","c","d","e"]},{"id":44,"label":"row_0044","payload":["a","b","c","d","e"]},{"id":45,"label":"row_0045","payload":["a","b","c","d","e"]},{"id":46,"label":"row_0046","payload":["a","b","c","d","e"]},{"id":47,"label":"row_0047","payload":["a","b","c","d","e"]},{"id":48,"label":"row_0048","payload":["a","b","c","d","e"]},{"id":49,"label":"row_0049","payload":["a","b","c","d","e"]},{"id":50,"label":"row_0050","payload":["a","b","c","d","e"]},{"id":51,"label":"row_0051","payload":["a","b","c","d","e"]},{"id":52,"label":"row_0052","payload":["a","b","c","d","e"]},{"id":53,"label":"row_0053","payload":["a","b","c","d","e"]},{"id":54,"label":"row_0054","payload":["a","b","c","d","e"]},{"id":55,"label":"row_0055","payload":["a","b","c","d","e"]},{"id":56,"label":"row_0056","payload":["a","b","c","d","e"]},{"id":57,"label":"row_0057","payload":["a","b","c","d","e"]},{"id":58,"label":"row_0058","payload":["a","b","c","d","e"]},{"id":59,"label":"row_0059","payload":["a","b","c","d","e"]},{"id":60,"label":"row_0060","payload":["a","b","c","d","e"]},{"id":61,"label":"row_0061","payload":["a","b","c","d","e"]},{"id":62,"label":"row_0062","payload":["a","b","c","d","e"]},{"id":63,"label":"row_0063","payload":["a","b","c","d","e"]},{"id":64,"label":"row_0064","payload":["a","b","c","d","e"]},{"id":65,"label":"row_0065","payload":["a","b","c","d","e"]},{"id":66,"label":"row_0066","payload":["a","b","c","d","e"]},{"id":67,"label":"row_0067","payload":["a","b","c","d","e"]},{"id":68,"label":"row_0068","payload":["a","b","c","d","e"]},{"id":69,"label":"row_0069","payload":["a","b","c","d","e"]},{"id":70,"label":"row_0070","payload":["a","b","c","d","e"]},{"id":71,"label":"row_0071","payload":["a","b","c","d","e"]},{"id":72,"label":"row_0072","payload":["a","b","c","d","e"]},{"id":73,"label":"row_0073","payload":["a","b","c","d","e"]},{"id":74,"label":"row_0074","payload":["a","b","c","d","e"]},{"id":75,"label":"row_0075","payload":["a","b","c","d","e"]},{"id":76,"label":"row_0076","payload":["a","b","c","d","e"]},{"id":77,"label":"row_0077","payload":["a","b","c","d","e"]},{"id":78,"label":"row_0078","payload":["a","b","c","d","e"]},{"id":79,"label":"row_0079","payload":["a","b","c","d","e"]}]'::jsonb)"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT length(doc::text) > 4096 FROM bigarr WHERE id = 1;",
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
                    query: "SELECT doc -> 17 ->> 'label' FROM bigarr WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("row_0017")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc -> 0 ->> 'label' FROM bigarr WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("row_0000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc -> 79 ->> 'label' FROM bigarr WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("row_0079")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc -> -1 ->> 'label' FROM bigarr WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("row_0079")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc -> 1000 FROM bigarr WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc #>> '{42, label}' FROM bigarr WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("row_0042")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc #>> '{42, payload, 3}' FROM bigarr WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("d")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb_extract_path on large stored object with numeric keys (>4 KB)",
            set_up_script: &[
                "CREATE TABLE numkeys (id INT PRIMARY KEY, doc JSONB)",
                r#"INSERT INTO numkeys (id, doc) VALUES (1, '{"nums":{"0":"zero","1":"one","2":"two"},"k_0000":{"name":"value_0000","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":0},"k_0001":{"name":"value_0001","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":1},"k_0002":{"name":"value_0002","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":2},"k_0003":{"name":"value_0003","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":3},"k_0004":{"name":"value_0004","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":4},"k_0005":{"name":"value_0005","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":5},"k_0006":{"name":"value_0006","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":6},"k_0007":{"name":"value_0007","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":7},"k_0008":{"name":"value_0008","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":8},"k_0009":{"name":"value_0009","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":9},"k_0010":{"name":"value_0010","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":10},"k_0011":{"name":"value_0011","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":11},"k_0012":{"name":"value_0012","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":12},"k_0013":{"name":"value_0013","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":13},"k_0014":{"name":"value_0014","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":14},"k_0015":{"name":"value_0015","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":15},"k_0016":{"name":"value_0016","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":16},"k_0017":{"name":"value_0017","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":17},"k_0018":{"name":"value_0018","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":18},"k_0019":{"name":"value_0019","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":19},"k_0020":{"name":"value_0020","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":20},"k_0021":{"name":"value_0021","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":21},"k_0022":{"name":"value_0022","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":22},"k_0023":{"name":"value_0023","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":23},"k_0024":{"name":"value_0024","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":24},"k_0025":{"name":"value_0025","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":25},"k_0026":{"name":"value_0026","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":26},"k_0027":{"name":"value_0027","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":27},"k_0028":{"name":"value_0028","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":28},"k_0029":{"name":"value_0029","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":29},"k_0030":{"name":"value_0030","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":30},"k_0031":{"name":"value_0031","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":31},"k_0032":{"name":"value_0032","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":32},"k_0033":{"name":"value_0033","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":33},"k_0034":{"name":"value_0034","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":34},"k_0035":{"name":"value_0035","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":35},"k_0036":{"name":"value_0036","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":36},"k_0037":{"name":"value_0037","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":37},"k_0038":{"name":"value_0038","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":38},"k_0039":{"name":"value_0039","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":39},"k_0040":{"name":"value_0040","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":40},"k_0041":{"name":"value_0041","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":41},"k_0042":{"name":"value_0042","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":42},"k_0043":{"name":"value_0043","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":43},"k_0044":{"name":"value_0044","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":44},"k_0045":{"name":"value_0045","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":45},"k_0046":{"name":"value_0046","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":46},"k_0047":{"name":"value_0047","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":47},"k_0048":{"name":"value_0048","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":48},"k_0049":{"name":"value_0049","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":49},"k_0050":{"name":"value_0050","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":50},"k_0051":{"name":"value_0051","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":51},"k_0052":{"name":"value_0052","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":52},"k_0053":{"name":"value_0053","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":53},"k_0054":{"name":"value_0054","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":54},"k_0055":{"name":"value_0055","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":55},"k_0056":{"name":"value_0056","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":56},"k_0057":{"name":"value_0057","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":57},"k_0058":{"name":"value_0058","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":58},"k_0059":{"name":"value_0059","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":59},"k_0060":{"name":"value_0060","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":60},"k_0061":{"name":"value_0061","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":61},"k_0062":{"name":"value_0062","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":62},"k_0063":{"name":"value_0063","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":63},"k_0064":{"name":"value_0064","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":64},"k_0065":{"name":"value_0065","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":65},"k_0066":{"name":"value_0066","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":66},"k_0067":{"name":"value_0067","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":67},"k_0068":{"name":"value_0068","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":68},"k_0069":{"name":"value_0069","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":69},"k_0070":{"name":"value_0070","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":70},"k_0071":{"name":"value_0071","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":71},"k_0072":{"name":"value_0072","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":72},"k_0073":{"name":"value_0073","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":73},"k_0074":{"name":"value_0074","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":74},"k_0075":{"name":"value_0075","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":75},"k_0076":{"name":"value_0076","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":76},"k_0077":{"name":"value_0077","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":77},"k_0078":{"name":"value_0078","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":78},"k_0079":{"name":"value_0079","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":79},"k_0080":{"name":"value_0080","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":80},"k_0081":{"name":"value_0081","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":81},"k_0082":{"name":"value_0082","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":82},"k_0083":{"name":"value_0083","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":83},"k_0084":{"name":"value_0084","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":84},"k_0085":{"name":"value_0085","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":85},"k_0086":{"name":"value_0086","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":86},"k_0087":{"name":"value_0087","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":87},"k_0088":{"name":"value_0088","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":88},"k_0089":{"name":"value_0089","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":89},"k_0090":{"name":"value_0090","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":90},"k_0091":{"name":"value_0091","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":91},"k_0092":{"name":"value_0092","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":92},"k_0093":{"name":"value_0093","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":93},"k_0094":{"name":"value_0094","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":94},"k_0095":{"name":"value_0095","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":95},"k_0096":{"name":"value_0096","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":96},"k_0097":{"name":"value_0097","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":97},"k_0098":{"name":"value_0098","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":98},"k_0099":{"name":"value_0099","tags":["tag-a","tag-b","tag-c","tag-d","tag-e"],"n":99}}'::jsonb)"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT length(doc::text) > 4096 FROM numkeys WHERE id = 1;",
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
                    query: "SELECT doc #>> '{nums, 0}' FROM numkeys WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("zero")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc #>> '{nums, 2}' FROM numkeys WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc #> '{nums, 5}' FROM numkeys WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT doc #>> '{k_0001, tags, 0}' FROM numkeys WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("tag-a")],
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
fn test_json_object_field() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb_object_field returns object value",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":"two"}'::jsonb -> 'a';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":"two"}'::jsonb -> 'b';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T(r#""two""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":"two"}'::jsonb -> null;"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"nested":{"x":[1,2,3]}}'::jsonb -> 'nested';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T(r#"{"x": [1, 2, 3]}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1}'::jsonb -> 'missing';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::jsonb -> 'a';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '42'::jsonb -> 'a';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a.b":1, "a":{"b":2}}'::jsonb -> 'a.b';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a\"b":7}'::jsonb -> 'a"b';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSONB)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "json_object_field returns object value",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":"two"}'::json -> 'a';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":"two"}'::json -> 'b';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[T(r#""two""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1}'::json -> 'missing';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::json -> 'a';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", JSON)],
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
            name: "jsonb_object_field_text returns object value as text",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1,"b":"two"}'::jsonb ->> 'b';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("two")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":42}'::jsonb ->> 'a';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":{"b":1}}'::jsonb ->> 'a';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
                        rows: &[
                            &[T(r#"{"b": 1}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"a":1}'::jsonb ->> 'missing';"#,
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TEXT)],
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

#[test]
fn test_json_object_keys() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb_object_keys",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_object_keys('{"b":1,"aa":2,"c":3}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_object_keys", TEXT)],
                        rows: &[
                            &[T("b")],
                            &[T("c")],
                            &[T("aa")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_object_keys('{"a":1}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_object_keys", TEXT)],
                        rows: &[
                            &[T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_object_keys('{}'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_object_keys", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_object_keys('{"a":{"nested":1},"b":2}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_object_keys", TEXT)],
                        rows: &[
                            &[T("a")],
                            &[T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT k FROM jsonb_object_keys('{"b":1,"aa":2}'::jsonb) AS k;"#,
                    expected: Expected::Rows {
                        columns: &[Column("k", TEXT)],
                        rows: &[
                            &[T("b")],
                            &[T("aa")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_object_keys('[1,2]'::jsonb);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot call jsonb_object_keys on an array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_object_keys('42'::jsonb);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot call jsonb_object_keys on a scalar", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "json_object_keys",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT json_object_keys('{"b":1,"aa":2,"c":3}'::json);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_object_keys", TEXT)],
                        rows: &[
                            &[T("b")],
                            &[T("aa")],
                            &[T("c")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_object_keys('{}'::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_object_keys", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_object_keys('["a"]'::json);"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot call json_object_keys on an array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_object_keys('"str"'::json);"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot call json_object_keys on a scalar", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_json_strip_nulls() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb_strip_nulls",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_strip_nulls('{"a":1,"b":null}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_strip_nulls", JSONB)],
                        rows: &[
                            &[T(r#"{"a": 1}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_strip_nulls('{"a":1,"b":null,"c":{"d":null,"e":2}}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_strip_nulls", JSONB)],
                        rows: &[
                            &[T(r#"{"a": 1, "c": {"e": 2}}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_strip_nulls('{"a":[1,null,{"b":null,"c":3}]}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_strip_nulls", JSONB)],
                        rows: &[
                            &[T(r#"{"a": [1, null, {"c": 3}]}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_strip_nulls('{"a":null}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_strip_nulls", JSONB)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_strip_nulls('[1,null,2]'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_strip_nulls", JSONB)],
                        rows: &[
                            &[T("[1, null, 2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_strip_nulls('{}'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_strip_nulls", JSONB)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_strip_nulls(null::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_strip_nulls", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(jsonb_strip_nulls('{}'::jsonb));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("jsonb")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "json_strip_nulls",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT json_strip_nulls('{"a":null,"b":1}'::json);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_strip_nulls", JSON)],
                        rows: &[
                            &[T(r#"{"b":1}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_strip_nulls('{"a":{"b":null}}'::json);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_strip_nulls", JSON)],
                        rows: &[
                            &[T(r#"{"a":{}}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_strip_nulls(null::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_strip_nulls", JSON)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(json_strip_nulls('{}'::json));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("json")],
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
fn test_json_table() {
    run_scripts(&[
        // Expectations from Postgres 17, since Postgres 15 lacks this feature.
        ScriptTest {
            name: "JSON_TABLE",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc JSONB);",
                r#"INSERT INTO docs VALUES (1, '[{"a":1},{"a":2}]'), (2, '[{"a":3}]'), (3, NULL), (4, '{"a":4}');"#,
                "CREATE TABLE texts (id INT PRIMARY KEY, doc TEXT);",
                r#"INSERT INTO texts VALUES (1, '{"a":"x"}');"#,
                "CREATE VIEW v AS SELECT docs.id, jt.a FROM docs, JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]', '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::json, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::text, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::varchar, '$[*]' COLUMNS (a INT PATH '$')) AS jt;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::text FORMAT JSON, '$[*]' COLUMNS (a INT PATH '$')) AS jt;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE(1, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type integer to jsonb", position: 26, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::bytea, '$[*]' COLUMNS (a INT PATH '$')) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast type bytea to jsonb", position: 26, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb FORMAT JSON ENCODING UTF8, '$[*]' COLUMNS (a INT PATH '$')) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "JSON ENCODING clause is only allowed for bytea input type", position: 39, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('not json'::text, '$[*]' COLUMNS (a INT)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "invalid input syntax for type json", detail: r#"Token "not" is invalid."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (a INT)) AS jt(x);"#,
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_table.a FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (a INT));"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE(NULL::jsonb, '$[*]' COLUMNS (a INT)) AS jt;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE(NULL, '$[*]' COLUMNS (a INT)) AS jt;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT docs.id, jt.* FROM docs, JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT docs.id, jt.* FROM docs
						CROSS JOIN LATERAL JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt
						ORDER BY 1, 2;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT docs.id, jt.* FROM docs
						LEFT JOIN LATERAL JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT PATH '$.a')) AS jt ON TRUE
						ORDER BY 1, 2;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("3"), Null],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT docs.id, jt.* FROM docs, JSON_TABLE(
							docs.doc,
							'$[*] ? (@.a > $m)' PASSING docs.id AS m
							COLUMNS (a INT PATH '$.a' DEFAULT -1 ON EMPTY)
						) AS jt
						ORDER BY 1, 2;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM v ORDER BY 1, 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("3")],
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jt.a FROM texts, JSON_TABLE(texts.doc, '$' COLUMNS (a TEXT)) AS jt;",
                    expected: Expected::Rows {
                        columns: &[Column("a", TEXT)],
                        rows: &[
                            &[T("x")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), sum(jt.a) FROM docs, JSON_TABLE(docs.doc, '$[*]' COLUMNS (a INT)) AS jt WHERE jt.a > 1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("3"), T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSON_TABLE paths and PASSING",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*] ? (@.a > $x)' PASSING 1 AS x COLUMNS (a INT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 74, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' AS p COLUMNS (a INT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "AS""#, position: 61, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[1,2,3]'::jsonb,
							'$[*] ? (@ >= $a && @ < $b)' PASSING 2 AS a, 3.5::numeric AS b
							COLUMNS (v INT PATH '$')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 88, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'["a","b"]'::jsonb,
							'$[*] ? (@ == $s)' PASSING 'b' AS s, 'b'::varchar AS t
							COLUMNS (v TEXT PATH '$', w TEXT PATH '$t')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 80, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[true,false]'::jsonb, '$[*] ? (@ == $s)' PASSING TRUE AS s COLUMNS (v TEXT PATH '$')) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 68, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"k":1},2]'::jsonb,
							'$[*] ? (@ == $s.k)' PASSING '{"k":2}'::jsonb AS s
							COLUMNS (v TEXT PATH '$')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 84, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[1,2]'::jsonb,
							'$' PASSING '2020-01-01'::date AS d, '2020-01-01 10:00'::timestamp AS ts, '10:00'::time AS t
							COLUMNS (
								v TEXT PATH '$d',
								w JSONB PATH '$d',
								x TEXT PATH '$ts',
								y JSONB PATH '$ts',
								z TEXT PATH '$t'
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 61, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT v, w::text FROM JSON_TABLE(
							'[1,2]'::jsonb,
							'$' PASSING NULL AS d, NULL::int AS e
							COLUMNS (v TEXT PATH '$d', w JSONB PATH '$e')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 70, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[1,2]'::jsonb,
							'$' PASSING 1.5::float8 AS d, 2::int8 AS e, 3::int2 AS f, 4.5::float4 AS g
							COLUMNS (v TEXT PATH '$d', w TEXT PATH '$e', x TEXT PATH '$f', y TEXT PATH '$g')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 61, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[1]'::jsonb,
							'$' PASSING ARRAY[1] AS d, ROW(1,2) AS e
							COLUMNS (v JSONB PATH '$d' ERROR ON ERROR, w JSONB PATH '$e' ERROR ON ERROR)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 59, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[1,2]'::jsonb,
							'$' PASSING 1 AS d, 2 AS d, 3 AS "D"
							COLUMNS (v TEXT PATH '$d', w TEXT PATH '$D')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 61, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[1,2]'::jsonb,
							'$' PASSING '[1]' FORMAT JSON AS d, '[2]'::text FORMAT JSON AS e
							COLUMNS (v TEXT PATH '$d[0]', w TEXT PATH '$e[0]')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 61, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1,2]'::jsonb, '$' PASSING 1 AS D COLUMNS (v TEXT PATH '$D')) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 46, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*] ? ($x > 0)' COLUMNS (e INT PATH '$')) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 58, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('["a","b"]'::jsonb, '$[*] ? (@ == $s)' PASSING 'b'::name AS s COLUMNS (v TEXT PATH '$')) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 65, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (a INT, a TEXT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 61, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' AS a COLUMNS (a INT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "AS""#, position: 61, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (n FOR ORDINALITY, m FOR ORDINALITY)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 61, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*' COLUMNS (a INT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 60, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '' COLUMNS (a INT)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 43, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$' || '[*]' COLUMNS (a INT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 67, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('{"a":1}'::jsonb, 'strict $.b' COLUMNS (a INT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 57, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('{"a":1}'::jsonb, 'strict $.b' COLUMNS (a INT) EMPTY ARRAY ON ERROR) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 57, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('{"a":1}'::jsonb, 'strict $.b' COLUMNS (a INT) ERROR ON ERROR) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 57, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('{"a":1}'::jsonb, 'strict $.b' COLUMNS (a INT) NULL ON ERROR) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 57, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*] / 0' COLUMNS (e INT PATH '$') ERROR ON ERROR) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 51, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a') ERROR ON ERROR) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSON_TABLE columns",
            set_up_script: &[
                "CREATE TYPE pair AS (x INT, y TEXT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{"a":2}]'::jsonb, '$[*]' COLUMNS (n FOR ORDINALITY, a INT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 61, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT a, b::text, n FROM JSON_TABLE(
							'[1,"x",true,null,1.5]'::jsonb,
							'$[*]' COLUMNS (a TEXT PATH '$', b JSONB PATH '$', n FOR ORDINALITY)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 92, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT i, t, n, b, j::text, js::text FROM JSON_TABLE(
							'[
								{"a": 1.5},
								{"a": "12"},
								{"a": true},
								{"a": [1]},
								{"a": {"b": 1}},
								{"a": null},
								{}
							]'::jsonb,
							'$[*]' COLUMNS (
								i INT PATH '$.a',
								t TEXT PATH '$.a',
								n NUMERIC PATH '$.a',
								b BOOL PATH '$.a',
								j JSONB PATH '$.a',
								js JSON PATH '$.a'
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 236, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'{"a": {"b": 1}}'::jsonb,
							'$' COLUMNS (
								a TEXT PATH '$.a',
								b TEXT FORMAT JSON PATH '$.a',
								c JSON PATH '$.a',
								d TEXT PATH '$."a"."b"',
								"A" INT PATH '$.a.b'
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 71, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'{"A": 1, "a": 2, "a b": 3, "q\"x": 4}'::jsonb,
							'$' COLUMNS ("A" INT, a INT, "a b" INT, "q""x" INT)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 93, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1.5}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":[1]}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a[*]' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 57, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (i JSONB PATH '$.a[*]' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 57, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[{}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' ERROR ON EMPTY)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 48, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[{}]'::jsonb, '$[*]' COLUMNS (i INT PATH 'strict $.a' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 48, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{}]'::jsonb,
							'$[*]' COLUMNS (i INT PATH 'strict $.a' ERROR ON EMPTY, j INT PATH '$.a' DEFAULT 5 ON EMPTY)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 63, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":"x"}]'::jsonb,
							'$[*]' COLUMNS (
								i INT PATH '$.a' DEFAULT 5 ON EMPTY DEFAULT 6 ON ERROR,
								j INT PATH '$.a' DEFAULT '7' ON ERROR,
								k INT PATH '$.a' DEFAULT 7.5 ON ERROR,
								l INT PATH '$.a' DEFAULT 1 + 1 ON ERROR,
								m INT PATH '$.a' DEFAULT length('ab') ON ERROR
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 70, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i JSONB PATH '$.b' DEFAULT '{"q":1}' ON EMPTY)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' DEFAULT 'q' ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' DEFAULT (SELECT 1) ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' EMPTY ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (i INT PATH '$.a' EMPTY ARRAY ON EMPTY)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":[1,2]}]'::jsonb, '$[*]' COLUMNS (q JSONB PATH '$.a[*]' TRUE ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 57, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[
								{"a": [1, 2]},
								{"a": 3},
								{"a": "s"},
								{"a": []},
								{"a": {"k": 1}},
								{}
							]'::jsonb,
							'$[*]' COLUMNS (
								w1 JSONB PATH '$.a[*]' WITH WRAPPER,
								w2 JSONB PATH '$.a[*]' WITH CONDITIONAL WRAPPER,
								w3 JSONB PATH '$.a' WITH CONDITIONAL WRAPPER,
								w4 JSONB PATH '$.a' WITH UNCONDITIONAL ARRAY WRAPPER,
								w5 JSONB PATH '$.a' WITHOUT WRAPPER,
								q1 TEXT PATH '$.a' OMIT QUOTES,
								q2 TEXT PATH '$.a' KEEP QUOTES,
								q3 TEXT PATH '$.a' OMIT QUOTES ON SCALAR STRING,
								q4 JSONB PATH '$.a' OMIT QUOTES,
								f1 TEXT FORMAT JSON PATH '$.a',
								f2 TEXT PATH '$.a' WITH WRAPPER,
								f3 INT PATH '$.a' KEEP QUOTES,
								f4 INT PATH '$.a' OMIT QUOTES
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 185, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":[1,2]}]'::jsonb,
							'$[*]' COLUMNS (
								q JSONB PATH '$.a[*]' EMPTY ON ERROR,
								r JSONB PATH '$.a[*]' EMPTY OBJECT ON ERROR,
								s JSONB PATH '$.b' EMPTY ARRAY ON EMPTY,
								t TEXT FORMAT JSON PATH '$.b' EMPTY OBJECT ON EMPTY,
								u JSONB PATH '$.a[*]' DEFAULT '"z"' ON ERROR
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 72, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":[1,2]}]'::jsonb,
							'$[*]' COLUMNS (q JSONB PATH '$.b' WITH WRAPPER, r JSONB PATH '$.b' WITH WRAPPER ERROR ON EMPTY)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 72, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"s"}]'::jsonb, '$[*]' COLUMNS (q JSONB PATH '$.a' OMIT QUOTES ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"x"}]'::jsonb, '$[*]' COLUMNS (q TEXT PATH '$.a' WITH WRAPPER OMIT QUOTES)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a INT FORMAT JSON)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 53, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a JSON FORMAT JSON ENCODING UTF8 PATH '$.a')) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 53, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a BYTEA FORMAT JSON ENCODING UTF16)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 53, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a BYTEA FORMAT JSON ENCODING FOO)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 53, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":1}]'::jsonb,
							'$[*]' COLUMNS (
								a INT PATH '$.a' WITHOUT ARRAY WRAPPER,
								b INT PATH '$.a' WITH CONDITIONAL ARRAY WRAPPER,
								c INT PATH '$.a' WITH ARRAY WRAPPER,
								d TEXT PATH '$.a' WITHOUT WRAPPER OMIT QUOTES,
								e TEXT PATH 'lax $.a' DEFAULT 'x' ON EMPTY ERROR ON ERROR
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 68, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"d":"2020-01-02"}]'::jsonb,
							'$[*]' COLUMNS (d DATE PATH '$.d', t TEXT PATH '$.d.datetime()', j JSONB PATH '$.d.datetime()')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 79, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[
								{"a": [[1, 2], [3, 4]]},
								{"a": [["a"]]},
								{"a": [1, [2]]},
								{"a": [null, "3"]},
								{"a": []}
							]'::jsonb,
							'$[*]' COLUMNS (a INT[], t TEXT[] PATH '$.a')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 197, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":"12"},{"a":"{1,2}"}]'::jsonb, '$[*]' COLUMNS (a INT[] PATH '$.a' OMIT QUOTES)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 70, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[
								{"a": {"x": 1, "y": "q"}},
								{"a": {"x": "z"}},
								{"a": 3}
							]'::jsonb,
							'$[*]' COLUMNS (a pair)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 148, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":1},{}]'::jsonb,
							'$[*]' COLUMNS (
								e BOOL EXISTS PATH '$.a',
								i INT EXISTS PATH '$.a',
								t TEXT EXISTS PATH '$.a',
								j JSONB EXISTS PATH '$.a',
								a BOOL EXISTS
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 71, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":1},{}]'::jsonb,
							'$[*]' COLUMNS (
								e BOOL EXISTS PATH 'strict $.a',
								f BOOL EXISTS PATH 'strict $.a' TRUE ON ERROR,
								g BOOL EXISTS PATH 'strict $.a' UNKNOWN ON ERROR,
								h BOOL EXISTS PATH 'strict $.a' FALSE ON ERROR
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 71, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[1]'::jsonb,
							'$[*]' COLUMNS (
								e INT EXISTS PATH '$.a' TRUE ON ERROR,
								u INT EXISTS PATH 'strict $.a' UNKNOWN ON ERROR,
								x INT EXISTS PATH 'strict $.a' TRUE ON ERROR,
								y BOOL EXISTS PATH '$ ? (@ / 0 > 1)' ERROR ON ERROR
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 62, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{}]'::jsonb, '$[*]' COLUMNS (e BOOL EXISTS PATH 'strict $.a' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 56, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1},{}]'::jsonb, '$[*]' COLUMNS (e DATE EXISTS PATH '$.a')) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 56, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e BOOL EXISTS PATH '$.a' NULL ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e BOOL EXISTS PATH '$.a' EMPTY ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH '$ / 0')) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH '$ / 0' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSON_TABLE NESTED PATH",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[
								{"a": 1, "b": [10, 20], "c": ["x"]},
								{"a": 2, "b": [], "c": []},
								{"a": 3, "b": [30], "c": ["y", "z"]}
							]'::jsonb,
							'$[*]' COLUMNS (
								n FOR ORDINALITY,
								a INT,
								NESTED PATH '$.b[*]' COLUMNS (bn FOR ORDINALITY, b INT PATH '$'),
								NESTED '$.c[*]' AS cp COLUMNS (cn FOR ORDINALITY, c TEXT PATH '$')
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 195, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a": 1, "b": [{"x": [1, 2]}, {"x": [3]}]}]'::jsonb,
							'$[*]' COLUMNS (
								a INT,
								NESTED PATH '$.b[*]' COLUMNS (
									bn FOR ORDINALITY,
									NESTED PATH '$.x[*]' COLUMNS (xn FOR ORDINALITY, x INT PATH '$')
								)
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 103, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1,"b":[1]}]'::jsonb, '$[*]' COLUMNS (NESTED PATH '$.b[*]' COLUMNS (b INT PATH '$'))) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 61, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":1,"b":1}]'::jsonb,
							'$[*]' COLUMNS (a INT, NESTED PATH 'strict $.q' COLUMNS (b INT PATH '$'))
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 74, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":1,"b":1}]'::jsonb,
							'$[*]' COLUMNS (a INT, NESTED PATH 'strict $.q' COLUMNS (b INT PATH '$')) ERROR ON ERROR
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 74, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":1,"b":[1]}]'::jsonb,
							'$[*]' COLUMNS (a INT, NESTED PATH '$.b[*]' COLUMNS (a INT PATH '$'))
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 76, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":1,"b":[1]}]'::jsonb,
							'$[*]' AS p COLUMNS (a INT, NESTED PATH '$.b[*]' AS p COLUMNS (b INT PATH '$'))
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "AS""#, position: 76, ..E }),
                    ..A
                },
            ],
            ..S
        },
        // Expectations from Postgres 17, since Postgres 15 lacks this feature.
        ScriptTest {
            name: "JSON_TABLE composition",
            set_up_script: &[
                "CREATE TABLE orders (id INT PRIMARY KEY, doc JSONB);",
                r#"INSERT INTO orders VALUES
					(1, '{
						"customer": "ann",
						"items": [
							{"sku": "a", "qty": 2, "tags": ["x", "y"]},
							{"sku": "b", "qty": 1, "tags": []}
						]
					}'),
					(2, '{
						"customer": "bob",
						"items": [
							{"sku": "c", "qty": 5, "tags": ["z"]}
						]
					}'),
					(3, '{"customer": "cy", "items": []}');"#,
                r#"CREATE VIEW order_items AS SELECT o.id, jt.* FROM orders o, JSON_TABLE(
						o.doc,
						'$' AS root PASSING 1 AS minqty
						COLUMNS (
							customer TEXT,
							NESTED PATH '$.items[*] ? (@.qty >= $minqty)' AS items COLUMNS (
								n FOR ORDINALITY,
								sku TEXT,
								qty INT,
								big BOOL EXISTS PATH '$ ? (@.qty > 1)',
								NESTED PATH '$.tags[*]' COLUMNS (tag TEXT PATH '$')
							)
						)
					) AS jt;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM order_items ORDER BY id, n, tag;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("customer", TEXT), Column("n", INT4), Column("sku", TEXT), Column("qty", INT4), Column("big", BOOL), Column("tag", TEXT)],
                        rows: &[
                            &[T("1"), T("ann"), T("1"), T("a"), T("2"), T("t"), T("x")],
                            &[T("1"), T("ann"), T("1"), T("a"), T("2"), T("t"), T("y")],
                            &[T("1"), T("ann"), T("2"), T("b"), T("1"), T("f"), Null],
                            &[T("2"), T("bob"), T("1"), T("c"), T("5"), T("t"), T("z")],
                            &[T("3"), T("cy"), Null, Null, Null, Null, Null],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT o.id, i.sku, t.tag FROM orders o,
						JSON_TABLE(o.doc, '$.items[*]' COLUMNS (sku TEXT, tags JSONB)) AS i,
						JSON_TABLE(i.tags, '$[*]' COLUMNS (tag TEXT PATH '$')) AS t
						ORDER BY 1, 2, 3;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("sku", TEXT), Column("tag", TEXT)],
                        rows: &[
                            &[T("1"), T("a"), T("x")],
                            &[T("1"), T("a"), T("y")],
                            &[T("2"), T("c"), T("z")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jt.customer, count(*) FROM orders,
						JSON_TABLE(orders.doc, '$' COLUMNS (customer TEXT, NESTED PATH '$.items[*]' COLUMNS (sku TEXT))) AS jt
						WHERE jt.sku IS NOT NULL GROUP BY jt.customer ORDER BY 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("customer", TEXT), Column("count", INT8)],
                        rows: &[
                            &[T("ann"), T("2")],
                            &[T("bob"), T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH items AS (SELECT jt.* FROM orders, JSON_TABLE(orders.doc, '$.items[*]' COLUMNS (sku TEXT, qty INT)) AS jt)
						SELECT sum(qty) FROM items;"#,
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT8)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT id FROM orders
						WHERE EXISTS (SELECT 1 FROM JSON_TABLE(orders.doc, '$.items[*]' COLUMNS (qty INT)) AS jt WHERE jt.qty > 4);"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT max(qty) FROM JSON_TABLE(orders.doc, '$.items[*]' COLUMNS (qty INT)) AS jt) AS m FROM orders ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("m", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("5")],
                            &[Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE($1::jsonb, '$[*] ? (@.a > $x)' PASSING $2::int AS x COLUMNS (a INT)) AS jt;",
                    bind_vars: &[BindVar::Str(r#"[{"a":1},{"a":2},{"a":3}]"#), BindVar::Int(1)],
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT o.id, jt.qty FROM orders o,
						JSON_TABLE(o.doc, '$.items[*] ? (@.qty >= $m)' PASSING $1::int AS m COLUMNS (qty INT)) AS jt
						ORDER BY 1, 2;"#,
                    bind_vars: &[BindVar::Int(2)],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("qty", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("2"), T("5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1}]'::jsonb, '$[*]' COLUMNS (a INT)) AS jt WHERE a = 1 ORDER BY a DESC LIMIT 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":1}]', '$[*]' COLUMNS (a INT)) AS jt
						JOIN JSON_TABLE('[{"b":1}]', '$[*]' COLUMNS (b INT)) AS jt2 ON jt.a = jt2.b;"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT pg_typeof(a), pg_typeof(b), pg_typeof(c), pg_typeof(n) FROM JSON_TABLE(
							'[{"a":1}]',
							'$[*]' COLUMNS (a INT, b JSONB PATH '$.a', c TEXT EXISTS PATH '$.a', n FOR ORDINALITY)
						) AS jt;"#,
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("integer"), T("jsonb"), T("text"), T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSON_TABLE column types and nested behaviors",
            set_up_script: &[
                "CREATE TYPE pair AS (x INT, y TEXT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"n": "1.234", "v": "abc", "ts": "2020-01-02 03:04:05", "d": "2020-01-02"}]'::jsonb,
							'$[*]' COLUMNS (
								n NUMERIC(5,2),
								v VARCHAR(5),
								ts TIMESTAMP,
								d DATE,
								i2 SMALLINT PATH '$.n' DEFAULT 7 ON ERROR,
								f FLOAT8 PATH '$.n'
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 135, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT a, b::text, c::text FROM JSON_TABLE(
							'[1]'::bytea FORMAT JSON ENCODING UTF8,
							'$[*]' COLUMNS (a INT PATH '$', b BYTEA FORMAT JSON PATH '$', c BYTEA PATH '$')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "FORMAT""#, position: 65, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('{"é":"ü"}'::jsonb, '$' COLUMNS ("é" TEXT)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 50, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"k":{"a":1}},{"k":2}]'::jsonb,
							'$[*]' COLUMNS (
								kv TEXT PATH '$.k.keyvalue().key',
								sz INT PATH '$.k.size()',
								t TEXT PATH '$.k.type()',
								dbl FLOAT8 PATH '$.k.double()'
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 82, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[1]'::jsonb,
							'$[*]' COLUMNS (e INT PATH 'strict $.a' DEFAULT -1 ON ERROR, f INT PATH '$.a' DEFAULT -2 ON EMPTY)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 62, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":{"x":1,"y":"q"}}]'::jsonb,
							'$[*]' COLUMNS (a pair[] PATH '$[*].a' WITH WRAPPER, b TEXT PATH '$.a.y')
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 82, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":[1,2],"b":[3]}]'::jsonb,
							'$[*]' COLUMNS (
								NESTED PATH '$.a[*]' COLUMNS (a INT PATH '$', ae BOOL EXISTS PATH '$ ? (@ > 1)', aj JSONB PATH '$' WITH WRAPPER),
								NESTED PATH '$.b[*]' COLUMNS (b INT PATH '$' DEFAULT 0 ON ERROR)
							)
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 80, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":[1,2]}]'::jsonb,
							'$[*]' PASSING 2 AS lim COLUMNS (NESTED PATH '$.a[*] ? (@ < $lim)' COLUMNS (a INT PATH '$'))
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "PASSING""#, position: 72, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE(
							'[{"a":[1,2]}]'::jsonb,
							'$[*]' COLUMNS (NESTED PATH '$.a[*]' COLUMNS (a INT PATH 'strict $.x' ERROR ON ERROR))
						) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 72, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":{"k":1}}]'::jsonb, '$[*]' COLUMNS (a INT[] ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 59, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":[[1],2]}]'::jsonb, '$[*]' COLUMNS (a INT[] ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 59, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":3}]'::jsonb, '$[*]' COLUMNS (a pair ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 53, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('[{"a":[1]}]'::jsonb, '$[*]' COLUMNS (a pair ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 55, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "JSON_TABLE path error codes",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $.a' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $.*' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $[0]' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[[1]]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $[5]' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 49, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('["a"]'::jsonb, '$[*]' COLUMNS (e INT PATH '-$' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 49, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('["a"]'::jsonb, '$[*]' COLUMNS (e INT PATH '$ + 1' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 49, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('["a"]'::jsonb, '$[*]' COLUMNS (e INT PATH '$.abs()' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 49, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e INT PATH 'strict $.size()' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e JSONB PATH '$.keyvalue()' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$[*]' COLUMNS (e TEXT PATH '$.datetime()' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('["x"]'::jsonb, '$[*]' COLUMNS (e TEXT PATH '$.datetime()' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 49, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM JSON_TABLE('["x"]'::jsonb, '$[*]' COLUMNS (e FLOAT8 PATH '$.double()' ERROR ON ERROR)) AS jt;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 49, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[[]]'::jsonb, '$[*]' COLUMNS (e FLOAT8 PATH 'strict $.double()' ERROR ON ERROR)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 48, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, 'strict $.a' COLUMNS (e INT PATH '$') ERROR ON ERROR) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 53, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '$$' COLUMNS (a INT)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 45, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM JSON_TABLE('[1]'::jsonb, '@' COLUMNS (a INT)) AS jt;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "COLUMNS""#, position: 44, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_json_typeof() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb_typeof over every JSON type",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof('{}'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_typeof('{"a":1}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof('[]'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("array")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof('[1,2,3]'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("array")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_typeof('"str"'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("string")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof('1'::jsonb), jsonb_typeof('-1.5'::jsonb), jsonb_typeof('1e3'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT), Column("jsonb_typeof", TEXT), Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("number"), T("number"), T("number")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof('true'::jsonb), jsonb_typeof('false'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT), Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("boolean"), T("boolean")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof('null'::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("null")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof(null::jsonb);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT jsonb_typeof('{"a":[1,2]}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T("object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(jsonb_typeof('{}'::jsonb));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("text")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "json_typeof over every JSON type",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT json_typeof('{"a":1}'::json);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_typeof", TEXT)],
                        rows: &[
                            &[T("object")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_typeof('[1]'::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_typeof", TEXT)],
                        rows: &[
                            &[T("array")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_typeof('"str"'::json);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_typeof", TEXT)],
                        rows: &[
                            &[T("string")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_typeof('42'::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_typeof", TEXT)],
                        rows: &[
                            &[T("number")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_typeof('true'::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_typeof", TEXT)],
                        rows: &[
                            &[T("boolean")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_typeof('null'::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_typeof", TEXT)],
                        rows: &[
                            &[T("null")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_typeof(null::json);",
                    expected: Expected::Rows {
                        columns: &[Column("json_typeof", TEXT)],
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
            name: "jsonb_typeof pins the shape of a column in a CHECK constraint",
            set_up_script: &[
                r#"CREATE TABLE public.t (
				    ext jsonb,
				    evidence jsonb,
				    CONSTRAINT t_ext_object_check CHECK ((jsonb_typeof(ext) = 'object'::text)),
				    CONSTRAINT t_evidence_check   CHECK (((jsonb_typeof(evidence) = 'array'::text)
				                                          AND (jsonb_array_length(evidence) >= 1)))
				);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"INSERT INTO t VALUES ('{"a":1}', '[1]');"#,
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('[1]', '[1]');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t" violates check constraint "t_ext_object_check""#, detail: "Failing row contains ([1], [1]).", schema: "public", table: "t", constraint: "t_ext_object_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"INSERT INTO t VALUES ('{"a":1}', '[]');"#,
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "t" violates check constraint "t_evidence_check""#, detail: r#"Failing row contains ({"a": 1}, [])."#, schema: "public", table: "t", constraint: "t_evidence_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_typeof(ext), jsonb_array_length(evidence) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_typeof", TEXT), Column("jsonb_array_length", INT4)],
                        rows: &[
                            &[T("object"), T("1")],
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
fn test_jsonb_numeric_casts() {
    run_scripts(&[
        ScriptTest {
            name: "jsonb -> int2: rounding, boundaries, and out-of-range",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '12345'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-12345'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("-12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '12345.4'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '12345.5'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("12346")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '12346.5'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("12347")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '32767'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("32767")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-32768'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("-32768")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '32767.4'::jsonb::int2;",
                    expected: Expected::Rows {
                        columns: &[Column("int2", INT2)],
                        rows: &[
                            &[T("32767")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '32768'::jsonb::int2;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "smallint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '32767.5'::jsonb::int2;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "smallint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-32769'::jsonb::int2;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "smallint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1e20'::jsonb::int2;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "smallint out of range", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb -> int4: rounding, boundaries, and out-of-range",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '0'::jsonb::int4;",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2147483647'::jsonb::int4;",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4)],
                        rows: &[
                            &[T("2147483647")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-2147483648'::jsonb::int4;",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4)],
                        rows: &[
                            &[T("-2147483648")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2147483647.4'::jsonb::int4;",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4)],
                        rows: &[
                            &[T("2147483647")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '2147483648'::jsonb::int4;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "integer out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-2147483649'::jsonb::int4;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "integer out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1e20'::jsonb::int4;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "integer out of range", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb -> int8: rounding, boundaries, and out-of-range",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '0'::jsonb::int8;",
                    expected: Expected::Rows {
                        columns: &[Column("int8", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '9007199254740991'::jsonb::int8;",
                    expected: Expected::Rows {
                        columns: &[Column("int8", INT8)],
                        rows: &[
                            &[T("9007199254740991")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-9007199254740991'::jsonb::int8;",
                    expected: Expected::Rows {
                        columns: &[Column("int8", INT8)],
                        rows: &[
                            &[T("-9007199254740991")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1e20'::jsonb::int8;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-1e20'::jsonb::int8;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "bigint out of range", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb -> float4: out-of-range",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '0'::jsonb::float4;",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1.5'::jsonb::float4;",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4)],
                        rows: &[
                            &[T("1.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '3.4e38'::jsonb::float4;",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4)],
                        rows: &[
                            &[T("3.4e+38")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '3.5e38'::jsonb::float4;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""350000000000000000000000000000000000000" is out of range for type real"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-3.5e38'::jsonb::float4;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""-350000000000000000000000000000000000000" is out of range for type real"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1e40'::jsonb::float4;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""10000000000000000000000000000000000000000" is out of range for type real"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb -> float8 round-trips finite values",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '0'::jsonb::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("float8", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1.5'::jsonb::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("float8", FLOAT8)],
                        rows: &[
                            &[T("1.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1e300'::jsonb::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("float8", FLOAT8)],
                        rows: &[
                            &[T("1e+300")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb -> numeric: preserves precision",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '12345'::jsonb::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("12345")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '12345.67'::jsonb::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("12345.67")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '-12345.67'::jsonb::numeric;",
                    expected: Expected::Rows {
                        columns: &[Column("numeric", NUMERIC)],
                        rows: &[
                            &[T("-12345.67")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb non-numeric values reject numeric casts",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '{}'::jsonb::int4;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot cast jsonb object to type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[]'::jsonb::int4;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot cast jsonb array to type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '"42"'::jsonb::int4;"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot cast jsonb string to type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'true'::jsonb::int4;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot cast jsonb boolean to type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'null'::jsonb::int4;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot cast jsonb null to type integer", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}'::jsonb::float4;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot cast jsonb object to type real", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[]'::jsonb::numeric;",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "cannot cast jsonb array to type numeric", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_to_jsonb() {
    run_scripts(&[
        ScriptTest {
            name: "to_jsonb over scalar types",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_jsonb(1), to_jsonb(1.5::numeric), to_jsonb(true), to_jsonb('a'::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB), Column("to_jsonb", JSONB), Column("to_jsonb", JSONB), Column("to_jsonb", JSONB)],
                        rows: &[
                            &[T("1"), T("1.5"), T("true"), T(r#""a""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_jsonb('{"not":"parsed"}'::text);"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB)],
                        rows: &[
                            &[T(r#""{\"not\":\"parsed\"}""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_jsonb('2020-01-01'::date), jsonb_typeof(to_jsonb('2020-01-01'::date));",
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB), Column("jsonb_typeof", TEXT)],
                        rows: &[
                            &[T(r#""2020-01-01""#), T("string")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_jsonb(null::int4);",
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(to_jsonb(1));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("jsonb")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_jsonb over composite values",
            set_up_script: &[
                "CREATE TABLE t (id int4, name text)",
                "INSERT INTO t VALUES (1, 'one')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_jsonb(ARRAY[1,2,3]);",
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB)],
                        rows: &[
                            &[T("[1, 2, 3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_jsonb(ARRAY['a','b']::text[]);",
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB)],
                        rows: &[
                            &[T(r#"["a", "b"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_jsonb(t) FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB)],
                        rows: &[
                            &[T(r#"{"id": 1, "name": "one"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_jsonb('{"bbb":1,"a":2}'::json), to_json('{"bbb":1,"a":2}'::json);"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB), Column("to_json", JSON)],
                        rows: &[
                            &[T(r#"{"a": 2, "bbb": 1}"#), T(r#"{"bbb":1,"a":2}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_jsonb('{"a":1}'::jsonb);"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_jsonb", JSONB)],
                        rows: &[
                            &[T(r#"{"a": 1}"#)],
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
