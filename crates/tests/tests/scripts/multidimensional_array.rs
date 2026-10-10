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
fn test_multidimensional_arrays() {
    run_scripts(&[
        ScriptTest {
            name: "constructors and literals",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[1,2],[3,4]];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[ARRAY[1,2],ARRAY[3,4]]::text[];",
                    expected: Expected::Rows {
                        columns: &[Column("array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{{1,2},{3,4}}'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{ {1, 2} , {3,4} }'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{{{1},{2}},{{3},{4}}}'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{{{1},{2}},{{3},{4}}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{{"a b",c},{d,NULL}}'::text[];"#,
                    expected: Expected::Rows {
                        columns: &[Column("text", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{{"a b",c},{d,NULL}}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[['a b','c'],['d',NULL]];",
                    expected: Expected::Rows {
                        columns: &[Column("array", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{{"a b",c},{d,NULL}}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[ARRAY[]::int[]];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[1,2],[3]];",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "multidimensional arrays must have array expressions with matching dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[NULL::int[], ARRAY[1]];",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "multidimensional arrays must have array expressions with matching dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{{1,2},{3}}'::int[];",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{{1,2},{3}}""#, detail: "Multidimensional arrays must have sub-arrays with matching dimensions.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,{2}}'::int[];",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{1,{2}}""#, detail: "Multidimensional arrays must have sub-arrays with matching dimensions.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{{}}'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("int4", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{a,}'::text[];",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{a,}""#, detail: r#"Unexpected "}" character."#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(ARRAY[[1,2]]);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("integer[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[1.5,2],[3,4]]::numeric(3,1)[];",
                    expected: Expected::Rows {
                        columns: &[Column("array", NUMERIC_ARRAY)],
                        rows: &[
                            &[T("{{1.5,2.0},{3.0,4.0}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[true,false]];",
                    expected: Expected::Rows {
                        columns: &[Column("array", BOOL_ARRAY)],
                        rows: &[
                            &[T("{{t,f}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "table storage",
            set_up_script: &[
                "CREATE TABLE t (pk INT PRIMARY KEY, v INT[][], w TEXT[3][4]);",
                "INSERT INTO t VALUES (1, ARRAY[[1,2],[3,4]], ARRAY[['a','b']]), (2, ARRAY[5,6], '{{{x}}}'), (3, '{}', NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v", INT4_ARRAY), Column("w", TEXT_ARRAY)],
                        rows: &[
                            &[T("1"), T("{{1,2},{3,4}}"), T("{{a,b}}")],
                            &[T("2"), T("{5,6}"), T("{{{x}}}")],
                            &[T("3"), T("{}"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET v = ARRAY[[[1,2,3]]] WHERE pk = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk, v, array_ndims(v), array_dims(v) FROM t ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v", INT4_ARRAY), Column("array_ndims", INT4), Column("array_dims", TEXT)],
                        rows: &[
                            &[T("1"), T("{{1,2},{3,4}}"), T("2"), T("[1:2][1:2]")],
                            &[T("2"), T("{{{1,2,3}}}"), T("3"), T("[1:1][1:1][1:3]")],
                            &[T("3"), T("{}"), Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE v = ARRAY[[1,2],[3,4]];",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE v[2][1] = 3;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t ORDER BY v;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("2")],
                            &[T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM t WHERE 3 = ANY(v);",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (4, '{{1,2},{3}}', NULL);",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"malformed array literal: "{{1,2},{3}}""#, detail: "Multidimensional arrays must have sub-arrays with matching dimensions.", position: 26, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "subscripts, functions, and operators",
            set_up_script: &[
                "CREATE TABLE agg (pk INT PRIMARY KEY, v INT[]);",
                "INSERT INTO agg VALUES (1, ARRAY[1,2]), (2, ARRAY[3,4]), (3, ARRAY[5]), (4, NULL), (5, '{}');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[[1,2],[3,4]])[1][2], (ARRAY[[1,2],[3,4]])[2][1];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4), Column("array", INT4)],
                        rows: &[
                            &[T("2"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (ARRAY[[1,2],[3,4]])[1], (ARRAY[[1,2],[3,4]])[1][2][3], (ARRAY[[1,2],[3,4]])[3][1], (ARRAY[1,2])[1][1];",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4), Column("array", INT4), Column("array", INT4), Column("array", INT4)],
                        rows: &[
                            &[Null, Null, Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_length(ARRAY[[1,2,3],[4,5,6]], 1), array_length(ARRAY[[1,2,3],[4,5,6]], 2), array_length(ARRAY[[1,2,3],[4,5,6]], 3), array_length(ARRAY[[1,2,3],[4,5,6]], 0);",
                    expected: Expected::Rows {
                        columns: &[Column("array_length", INT4), Column("array_length", INT4), Column("array_length", INT4), Column("array_length", INT4)],
                        rows: &[
                            &[T("2"), T("3"), Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_upper(ARRAY[[1,2,3],[4,5,6]], 2), array_upper(ARRAY[1,2,3,4], 2);",
                    expected: Expected::Rows {
                        columns: &[Column("array_upper", INT4), Column("array_upper", INT4)],
                        rows: &[
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_ndims(ARRAY[[1,2],[3,4]]), array_ndims(ARRAY[1]), array_ndims('{}'::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_ndims", INT4), Column("array_ndims", INT4), Column("array_ndims", INT4)],
                        rows: &[
                            &[T("2"), T("1"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_dims(ARRAY[[1,2,3],[4,5,6]]), array_dims(ARRAY[1]), array_dims('{}'::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_dims", TEXT), Column("array_dims", TEXT), Column("array_dims", TEXT)],
                        rows: &[
                            &[T("[1:2][1:3]"), T("[1:1]"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_subscripts(ARRAY[[1,2,3],[4,5,6]], 2);",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_subscripts(ARRAY[[1,2,3],[4,5,6]], 3);",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(ARRAY[[1,2],[3,4]]);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 3 = ANY(ARRAY[[1,2],[3,4]]), 5 = ANY(ARRAY[[1,2],[3,4]]), 3 = ALL(ARRAY[[3,3],[3,3]]);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_string(ARRAY[[1,2],[3,4]], ',');",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("1,2,3,4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_cat(ARRAY[[1,2],[3,4]], ARRAY[5,6]), array_cat(ARRAY[5,6], ARRAY[[1,2],[3,4]]), array_cat(ARRAY[[1,2],[3,4]], ARRAY[[5,6]]), array_cat(ARRAY[[1,2],[3,4]], '{}'::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_cat", INT4_ARRAY), Column("array_cat", INT4_ARRAY), Column("array_cat", INT4_ARRAY), Column("array_cat", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4},{5,6}}"), T("{{5,6},{1,2},{3,4}}"), T("{{1,2},{3,4},{5,6}}"), T("{{1,2},{3,4}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[1,2],[3,4]] || ARRAY[5,6];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4},{5,6}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_cat(ARRAY[[1,2],[3,4]], ARRAY[5,6,7]);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "cannot concatenate incompatible arrays", detail: "Arrays with differing dimensions are not compatible for concatenation.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_append(ARRAY[[1,2],[3,4]], 5);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "argument must be empty or one-dimensional array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_prepend(5, ARRAY[[1,2],[3,4]]);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "argument must be empty or one-dimensional array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_position(ARRAY[[1,2],[3,4]], 3);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "searching for elements in multidimensional arrays is not supported", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[1,2],[3,4]] = ARRAY[1,2,3,4], ARRAY[[1,2],[3,4]] > ARRAY[1,2,3,4], ARRAY[[1,2],[3,4]] < ARRAY[[1,2,3,4]], ARRAY[[1,2],[3,4]] = ARRAY[[1,2],[3,4]];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk <= 2;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk <= 3;",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "cannot accumulate arrays of different dimensionality", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk IN (1, 4);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "cannot accumulate null arrays", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk IN (1, 5);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "cannot accumulate arrays of different dimensionality", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v ORDER BY pk DESC) FROM agg WHERE pk IN (1, 5);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "cannot accumulate empty arrays", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT v FROM agg);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "cannot accumulate arrays of different dimensionality", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "arrays of vectors are one-dimensional",
            set_up_script: &[
                "CREATE TABLE ov (pk INT PRIMARY KEY, v oidvector[]);",
                r#"INSERT INTO ov VALUES (1, '{"1 2","3"}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM ov;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("v", OIDVECTOR_ARRAY)],
                        rows: &[
                            &[T("1"), T(r#"{"1 2",3}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT pk FROM ov WHERE v = '{"1 2","3"}';"#,
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT ('{"1 2","3"}'::oidvector[])[1], array_ndims('{"1 2","3"}'::oidvector[]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("oidvector", OIDVECTOR), Column("array_ndims", INT4)],
                        rows: &[
                            &[T("1 2"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "comparison, ordering, and keys",
            set_up_script: &[
                "CREATE TABLE karr (a INT[] PRIMARY KEY, b INT);",
                "INSERT INTO karr VALUES ('{1,2}', 1), ('{{1,2},{3,4}}', 2), ('{1}', 3), ('{{1},{2}}', 4), ('{2}', 5), ('{}', 6);",
                "CREATE TABLE iarr (pk INT PRIMARY KEY, a INT[]);",
                "CREATE INDEX iarr_a ON iarr (a);",
                "INSERT INTO iarr VALUES (1, '{{1,2},{3,4}}'), (2, '{1,2,3,4}'), (3, '{{1},{2},{3},{4}}'), (4, '{{{1,2},{3,4}}}');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '{{1,2},{3,4}}'::int[] = '{1,2,3,4}'::int[], '{{1,2},{3,4}}'::int[] > '{1,2,3,4}'::int[], '{{1,2},{3,4}}'::int[] < '{1,2,3,5}'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{{1,2}}'::int[] = '{{1},{2}}'::int[], '{{1,2}}'::int[] < '{{1},{2}}'::int[], '{{1},{2}}'::int[] < '{{1,2}}'::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a FROM karr ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                            &[T("{1}")],
                            &[T("{1,2}")],
                            &[T("{{1},{2}}")],
                            &[T("{{1,2},{3,4}}")],
                            &[T("{2}")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT b FROM karr ORDER BY a DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("b", INT4)],
                        rows: &[
                            &[T("5")],
                            &[T("2")],
                            &[T("4")],
                            &[T("1")],
                            &[T("3")],
                            &[T("6")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT b FROM karr WHERE a = ARRAY[[1,2],[3,4]];",
                    expected: Expected::Rows {
                        columns: &[Column("b", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT b FROM karr WHERE a > '{1,2}' ORDER BY b;",
                    expected: Expected::Rows {
                        columns: &[Column("b", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("4")],
                            &[T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE karr SET b = 20 WHERE a = ARRAY[[1,2],[3,4]];",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM karr WHERE b = 20;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4_ARRAY), Column("b", INT4)],
                        rows: &[
                            &[T("{{1,2},{3,4}}"), T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM karr WHERE a = ARRAY[[1],[2]];",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM karr;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM iarr WHERE a = '{{1,2},{3,4}}';",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM iarr WHERE a > '{1,2,3,4}' ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM iarr ORDER BY a;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
                        rows: &[
                            &[T("2")],
                            &[T("1")],
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pk FROM iarr WHERE a[1][2] = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4)],
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
    ]);
}
