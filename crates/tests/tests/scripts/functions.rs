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
use harness::script::Cell::{Any, Approx, Null, Oid, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_aggregate_functions() {
    run_scripts(&[
        ScriptTest {
            name: "bool_and",
            set_up_script: &[
                "CREATE TABLE t1 (pk INT primary key, v1 BOOLEAN, v2 BOOLEAN);",
                "INSERT INTO t1 VALUES (1, true, false), (2, true, true), (3, true, true);",
                "CREATE TABLE t2 (v1 BOOLEAN);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT bool_and(v1), bool_and(v2) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("bool_and", BOOL), Column("bool_and", BOOL)],
                        rows: &[
                            &[T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_and(v1 and v2) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("bool_and", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_and(v1 and v2) FROM t1 where v1 and v2;",
                    expected: Expected::Rows {
                        columns: &[Column("bool_and", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_and(v1) FROM t1 where pk > 10;",
                    expected: Expected::Rows {
                        columns: &[Column("bool_and", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_and(a) FROM (VALUES(true),(false),(null)) r(a);",
                    expected: Expected::Rows {
                        columns: &[Column("bool_and", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_and(a) FROM (VALUES(true),(false),(null::bool)) r(a);",
                    expected: Expected::Rows {
                        columns: &[Column("bool_and", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_and(a) FROM (VALUES(null::bool),(true),(null::bool)) r(a);",
                    expected: Expected::Rows {
                        columns: &[Column("bool_and", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_and(v1) FROM t2",
                    expected: Expected::Rows {
                        columns: &[Column("bool_and", BOOL)],
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
            name: "bool_or",
            set_up_script: &[
                "CREATE TABLE t1 (pk INT primary key, v1 BOOLEAN, v2 BOOLEAN);",
                "INSERT INTO t1 VALUES (1, false, false), (2, true, true), (3, true, false);",
                "CREATE TABLE t2 (v1 BOOLEAN);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT bool_or(v1), bool_or(v2) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("bool_or", BOOL), Column("bool_or", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_or(v1), bool_or(v2) FROM t1 where pk <> 2;",
                    expected: Expected::Rows {
                        columns: &[Column("bool_or", BOOL), Column("bool_or", BOOL)],
                        rows: &[
                            &[T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_or(v1 and v2) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("bool_or", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_or(v1) FROM t1 where pk > 10;",
                    expected: Expected::Rows {
                        columns: &[Column("bool_or", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_or(a) FROM (VALUES(true),(false),(null::bool)) r(a);",
                    expected: Expected::Rows {
                        columns: &[Column("bool_or", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_or(a) FROM (VALUES(null::bool),(false),(null::bool)) r(a);",
                    expected: Expected::Rows {
                        columns: &[Column("bool_or", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bool_or(v1) FROM t2",
                    expected: Expected::Rows {
                        columns: &[Column("bool_or", BOOL)],
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
            name: "json_agg",
            set_up_script: &[
                "SET TIME ZONE 'UTC'",
                "CREATE TABLE json_agg_records (id int4, label text)",
                "INSERT INTO json_agg_records VALUES (1, 'one'), (2, NULL)",
                "CREATE TABLE json_agg_arrays (id int4 primary key, v int4[])",
                "INSERT INTO json_agg_arrays VALUES (1, ARRAY[1,NULL,3]), (2, ARRAY[4,5,NULL])",
                "CREATE TABLE json_agg_stored (id int4 primary key, amount numeric(40,20), payload json)",
                r#"INSERT INTO json_agg_stored VALUES (1, 12345678901234567890.12345678901234567890, '{"kind":"stored"}')"#,
                "CREATE DOMAIN json_agg_positive_int AS int4 CHECK (VALUE > 0)",
                "CREATE FUNCTION json_agg_window_step(state int4, value int4) RETURNS int4 LANGUAGE SQL IMMUTABLE AS 'SELECT state + value'",
                "CREATE AGGREGATE json_agg_window_custom (int4) (SFUNC = json_agg_window_step, STYPE = int4, INITCOND = '0')",
                "CREATE SCHEMA json_agg_collision",
                "CREATE FUNCTION json_agg_collision.json_agg(value int4) RETURNS int4 LANGUAGE SQL IMMUTABLE AS 'SELECT value'",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (VALUES (1::int4),(2),(NULL)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T("[1, 2, null]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_agg(v) FROM (VALUES ('quote"slash\line'::text),(E'line\nnext')) AS t(v);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"["quote\"slash\\line", "line\nnext"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (VALUES (true),(false),(NULL::bool)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T("[true, false, null]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (VALUES (1.2300::numeric),('-4.5'::numeric),('NaN'::numeric)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"[1.2300, -4.5, "NaN"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (VALUES ('Infinity'::float8),('-Infinity'::float8),('NaN'::float8),(1.5::float8)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"["Infinity", "-Infinity", "NaN", 1.5]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (VALUES ('2024-02-29'::date),('0001-01-01 BC'::date)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"["2024-02-29", "0001-01-01 BC"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (VALUES ('2024-02-29 12:34:56.123456'::timestamp),('2024-03-01 00:00:00'::timestamp)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"["2024-02-29T12:34:56.123456", "2024-03-01T00:00:00"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (VALUES ('550e8400-e29b-41d4-a716-446655440000'::uuid),('00000000-0000-0000-0000-000000000000'::uuid)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"["550e8400-e29b-41d4-a716-446655440000", "00000000-0000-0000-0000-000000000000"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_agg(v) FROM (VALUES ('{"b": 2, "a": 1}'::json),('null'::json)) AS t(v);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"[{"b": 2, "a": 1}, null]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_agg(v) FROM (VALUES ('{"b": 2, "a": 1}'::jsonb),('[1, null]'::jsonb)) AS t(v);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"[{"a": 1, "b": 2}, [1, null]]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_agg(v) FROM (VALUES ('1'::json),('"two"'::json),('true'::json),('null'::json),('{"k":3}'::json),('[4]'::json)) AS t(v);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"[1, "two", true, null, {"k":3}, [4]]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM json_agg_arrays;",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"[[1,null,3], 
 [4,5,null]]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(amount) FROM json_agg_stored;",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T("[12345678901234567890.12345678901234567890]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(r) FROM (SELECT * FROM json_agg_stored ORDER BY id) r;",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"[{"id":1,"amount":12345678901234567890.12345678901234567890,"payload":{"kind":"stored"}}]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(NULL::text);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T("[null]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (SELECT 1::int AS v WHERE false) AS t;",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(json_agg(1));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("json")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(r) FROM (SELECT * FROM json_agg_records ORDER BY id) r;",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"[{"id":1,"label":"one"}, 
 {"id":2,"label":null}]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT g, json_agg(v) FROM (VALUES ('a',1),('a',2),('b',3),('b',NULL)) AS t(g,v) GROUP BY g ORDER BY g;",
                    expected: Expected::Rows {
                        columns: &[Column("g", TEXT), Column("json_agg", JSON)],
                        rows: &[
                            &[T("a"), T("[1, 2]")],
                            &[T("b"), T("[3, null]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) OVER (ORDER BY id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) FROM (VALUES (1,10),(2,NULL),(3,30)) AS t(id,v) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T("[10]")],
                            &[T("[10, null]")],
                            &[T("[10, null, 30]")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) OVER (ORDER BY id ROWS BETWEEN 1 PRECEDING AND 1 PRECEDING) FROM (VALUES (1,10),(2,20)) AS t(id,v) ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[Null],
                            &[T("[10]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT json_agg(v) FROM (VALUES ('\x00ff'::bytea),(NULL::bytea)) AS t(v);"#,
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T(r#"["\\x00ff", null]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(v) FROM (VALUES (1::json_agg_positive_int),(2::json_agg_positive_int)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T("[1, 2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(DISTINCT v) FROM (VALUES (1),(1),(NULL),(NULL)) AS t(v);",
                    expected: Expected::Rows {
                        columns: &[Column("json_agg", JSON)],
                        rows: &[
                            &[T("[1, null]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(DISTINCT v) OVER (PARTITION BY p ORDER BY id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) FROM (VALUES ('a',1,1),('a',2,2),('a',3,2),('b',1,NULL),('b',2,2),('b',3,2)) AS t(p,id,v);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg(DISTINCT v) OVER (ORDER BY id ROWS BETWEEN 1 PRECEDING AND 1 PRECEDING) FROM (VALUES (1,1),(2,NULL)) AS t(id,v);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(DISTINCT v) OVER () FROM (VALUES (1),(1),(NULL)) AS t(v);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT abs(DISTINCT v) OVER () FROM (VALUES (-1)) AS t(v);",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "DISTINCT specified, but abs is not an aggregate function", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg_window_custom(DISTINCT v) OVER () FROM (VALUES (1),(1)) AS t(v);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DISTINCT is not implemented for window functions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_agg_collision.json_agg(DISTINCT v) OVER () FROM (VALUES (1)) AS t(v);",
                    expected: Expected::Error(Diagnostic { code: "42809", message: "DISTINCT specified, but json_agg_collision.json_agg is not an aggregate function", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_agg",
            set_up_script: &[
                "CREATE TABLE t1 (pk INT primary key, t timestamp, v varchar, f float[]);",
                r#"INSERT INTO t1 VALUES 
                   (1, '2023-01-01 00:00:00', 'a', '{1.0, 2.0}'),
                   (2, '2023-01-02 00:00:00', 'b', '{3.0, 4.0}'),
                   (3, '2023-01-03 00:00:00', 'c', '{5.0, 6.0}');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_agg(pk) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(t) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TIMESTAMP_ARRAY)],
                        rows: &[
                            &[T(r#"{"2023-01-01 00:00:00","2023-01-02 00:00:00","2023-01-03 00:00:00"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{a,b,c}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(f) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4},{5,6}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_agg with order by",
            set_up_script: &[
                r#"CREATE TABLE test_data (
					id INT PRIMARY KEY, 
					name VARCHAR(50), 
					age INT, 
					score FLOAT, 
					created_at TIMESTAMP, 
					category CHAR(1),
					nullable_field VARCHAR(20)
				);"#,
                r#"INSERT INTO test_data VALUES 
					(1, 'Alice', 25, 95.5, '2023-01-03 10:00:00', 'A', 'value1'),
					(2, 'Bob', 30, 87.2, '2023-01-01 09:30:00', 'B', NULL),
					(3, 'Charlie', 22, 92.8, '2023-01-02 11:15:00', 'A', 'value2'),
					(4, 'Diana', 28, 88.9, '2023-01-04 08:45:00', 'C', NULL),
					(5, 'Eve', 35, 94.1, '2023-01-05 14:20:00', 'B', 'value3'),
					(6, 'Frank', 26, 89.3, '2023-01-06 16:30:00', 'A', 'value4');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY age ASC) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Charlie,Alice,Frank,Diana,Bob,Eve}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY age DESC) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Eve,Bob,Diana,Frank,Alice,Charlie}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(id ORDER BY age) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", INT4_ARRAY)],
                        rows: &[
                            &[T("{3,1,6,4,2,5}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY score DESC) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Alice,Eve,Charlie,Frank,Diana,Bob}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY created_at ASC) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Bob,Charlie,Alice,Diana,Eve,Frank}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(age ORDER BY name) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", INT4_ARRAY)],
                        rows: &[
                            &[T("{25,30,22,28,35,26}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY category ASC, age DESC) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Frank,Alice,Charlie,Eve,Bob,Diana}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(id ORDER BY category ASC, score DESC) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,3,6,5,2,4}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY age * 2) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Charlie,Alice,Frank,Diana,Bob,Eve}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(age ORDER BY category || name) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", INT4_ARRAY)],
                        rows: &[
                            &[T("{25,22,26,30,35,28}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY CASE WHEN age > 27 THEN 1 ELSE 0 END, age) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Charlie,Alice,Frank,Diana,Bob,Eve}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY nullable_field) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Alice,Charlie,Eve,Frank,Bob,Diana}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT category, array_agg(name ORDER BY age) FROM test_data GROUP BY category ORDER BY category;",
                    expected: Expected::Rows {
                        columns: &[Column("category", BPCHAR), Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("A"), T("{Charlie,Alice,Frank}")],
                            &[T("B"), T("{Bob,Eve}")],
                            &[T("C"), T("{Diana}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT category, array_agg(name ORDER BY (SELECT COUNT(*) FROM test_data t2 WHERE t2.category = test_data.category AND t2.age < test_data.age)) FROM test_data GROUP BY category ORDER BY category;",
                    expected: Expected::Rows {
                        columns: &[Column("category", BPCHAR), Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("A"), T("{Charlie,Alice,Frank}")],
                            &[T("B"), T("{Bob,Eve}")],
                            &[T("C"), T("{Diana}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY COALESCE(nullable_field, 'zzz')) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Alice,Charlie,Eve,Frank,Bob,Diana}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY LENGTH(name) DESC, name ASC) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Charlie,Alice,Diana,Frank,Bob,Eve}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT category, array_agg(name ORDER BY score - (SELECT AVG(score) FROM test_data t2 WHERE t2.category = test_data.category)) FROM test_data GROUP BY category ORDER BY category;",
                    expected: Expected::Rows {
                        columns: &[Column("category", BPCHAR), Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("A"), T("{Frank,Charlie,Alice}")],
                            &[T("B"), T("{Bob,Eve}")],
                            &[T("C"), T("{Diana}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY EXTRACT(hour FROM created_at)) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Diana,Bob,Alice,Charlie,Eve,Frank}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY age) FROM test_data WHERE age > 100;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(name ORDER BY age > 27, age) FROM test_data;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{Charlie,Alice,Frank,Diana,Bob,Eve}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array agg with case statement",
            set_up_script: &[
                "CREATE TABLE t1 (pk INT primary key, v1 INT, v2 INT);",
                "INSERT INTO t1 VALUES (1, 10, 20), (2, 30, 40), (3, 50, 60);",
                "CREATE TABLE t2 (pk INT primary key, v1 INT, v2 TEXT);",
                "INSERT INTO t2 VALUES (1, 10, 'a'), (2, 20, 'b'), (3, 30, 'c');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_agg(CASE WHEN v1 > 20 THEN v1 ELSE NULL END) FROM t1;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", INT4_ARRAY)],
                        rows: &[
                            &[T("{NULL,30,50}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(CASE WHEN v1 >= 20 THEN v2 ELSE NULL END) FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT_ARRAY)],
                        rows: &[
                            &[T("{NULL,b,c}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(CASE WHEN v1 > 20 THEN v1::text ELSE v2 END) FROM t2;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT_ARRAY)],
                        rows: &[
                            &[T("{a,b,30}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(CASE WHEN v1 > 20 THEN v1 ELSE v2 END) FROM t2;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "CASE types text and integer cannot be matched", position: 41, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_agg with DISTINCT",
            set_up_script: &[
                r#"CREATE TABLE test_items (
					id INT PRIMARY KEY,
					category TEXT NOT NULL,
					name TEXT NOT NULL
				);"#,
                r#"INSERT INTO test_items (id, category, name) VALUES
					(1, 'A', 'foo'),
					(2, 'A', 'bar'),
					(3, 'B', 'baz'),
					(4, 'A', 'foo');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_agg(DISTINCT name) FROM test_items;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT_ARRAY)],
                        rows: &[
                            &[T("{bar,baz,foo}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(DISTINCT name ORDER BY name ASC) FROM test_items;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT_ARRAY)],
                        rows: &[
                            &[T("{bar,baz,foo}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(DISTINCT name ORDER BY name DESC) FROM test_items;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT_ARRAY)],
                        rows: &[
                            &[T("{foo,baz,bar}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT category, array_agg(DISTINCT name ORDER BY name ASC) FROM test_items GROUP BY category ORDER BY category;",
                    expected: Expected::Rows {
                        columns: &[Column("category", TEXT), Column("array_agg", TEXT_ARRAY)],
                        rows: &[
                            &[T("A"), T("{bar,foo}")],
                            &[T("B"), T("{baz}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(DISTINCT category ORDER BY category) FROM test_items;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT_ARRAY)],
                        rows: &[
                            &[T("{A,B}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_agg with DISTINCT and composite types",
            set_up_script: &[
                "CREATE TYPE point_t AS (x INT, y INT);",
                "CREATE TABLE points (id INT PRIMARY KEY, p point_t);",
                r#"INSERT INTO points VALUES
					(1, ROW(1, 2)),
					(2, ROW(3, 4)),
					(3, ROW(1, 2)),
					(4, ROW(5, 6));"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_agg(DISTINCT p ORDER BY p) FROM points;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", USER_DEFINED)],
                        rows: &[
                            &[T(r#"{"(1,2)","(3,4)","(5,6)"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_agg with DISTINCT handles NULL values",
            set_up_script: &[
                "CREATE TABLE t (v text);",
                "INSERT INTO t VALUES (NULL), (NULL), ('x');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_agg(DISTINCT v ORDER BY v NULLS FIRST)::text FROM t;",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT)],
                        rows: &[
                            &[T("{NULL,x}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(DISTINCT v)::text FROM (VALUES (NULL::text), (NULL::text)) AS vals(v);",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT)],
                        rows: &[
                            &[T("{NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_agg DISTINCT dedups semantically equal numerics",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_agg(DISTINCT v ORDER BY v)::text FROM (VALUES (1.0::numeric), (1.00::numeric)) AS vals(v);",
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT)],
                        rows: &[
                            &[T("{1.0}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"
				SELECT array_agg(DISTINCT v ORDER BY v)::text
				FROM (
					VALUES
						(1::numeric),
						(1.000000::numeric)
				) AS vals(v);
			"#,
                    expected: Expected::Rows {
                        columns: &[Column("array_agg", TEXT)],
                        rows: &[
                            &[T("{1}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "numeric SUM over COALESCE",
            set_up_script: &[
                "CREATE TABLE numeric_sum_values (grp INT, amount NUMERIC(10,2));",
                r#"INSERT INTO numeric_sum_values VALUES
					(1, 12.50),
					(1, NULL),
					(2, NULL),
					(3, 99999999.99),
					(3, 0.01);"#,
                "CREATE TABLE numeric_sum_groups (id INT PRIMARY KEY);",
                "INSERT INTO numeric_sum_groups VALUES (1), (2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*), sum(coalesce(amount, 0)), pg_typeof(sum(coalesce(amount, 0))) FROM numeric_sum_values;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("sum", NUMERIC), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("5"), T("100000012.50"), T("numeric")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT grp, sum(coalesce(amount, 0)), pg_typeof(sum(coalesce(amount, 0))) FROM numeric_sum_values GROUP BY grp ORDER BY grp;",
                    expected: Expected::Rows {
                        columns: &[Column("grp", INT4), Column("sum", NUMERIC), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("1"), T("12.50"), T("numeric")],
                            &[T("2"), T("0"), T("numeric")],
                            &[T("3"), T("100000000.00"), T("numeric")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT grp, amount,
						sum(coalesce(amount, 0)) OVER (PARTITION BY grp ORDER BY amount),
						pg_typeof(sum(coalesce(amount, 0)) OVER (PARTITION BY grp ORDER BY amount))
					FROM numeric_sum_values ORDER BY grp, amount;"#,
                    expected: Expected::Rows {
                        columns: &[Column("grp", INT4), Column("amount", NUMERIC), Column("sum", NUMERIC), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("1"), T("12.50"), T("12.50"), T("numeric")],
                            &[T("1"), Null, T("12.50"), T("numeric")],
                            &[T("2"), Null, T("0"), T("numeric")],
                            &[T("3"), T("0.01"), T("0.01"), T("numeric")],
                            &[T("3"), T("99999999.99"), T("100000000.00"), T("numeric")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT id,
						(SELECT sum(coalesce(amount, 0)) FROM numeric_sum_values WHERE grp = numeric_sum_groups.id),
						pg_typeof((SELECT sum(coalesce(amount, 0)) FROM numeric_sum_values WHERE grp = numeric_sum_groups.id))
					FROM numeric_sum_groups ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("sum", NUMERIC), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("1"), T("12.50"), T("numeric")],
                            &[T("2"), T("0"), T("numeric")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "var_pop/var_samp/stddev_pop/stddev_samp with infinite and NaN float8 input",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT sum(x::float8), avg(x::float8), var_pop(x::float8)::text FROM (VALUES ('infinity'), ('1')) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("sum", FLOAT8), Column("avg", FLOAT8), Column("var_pop", TEXT)],
                        rows: &[
                            &[T("Infinity"), T("Infinity"), T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(x::float8), avg(x::float8), var_pop(x::float8)::text FROM (VALUES ('1'), ('infinity')) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("sum", FLOAT8), Column("avg", FLOAT8), Column("var_pop", TEXT)],
                        rows: &[
                            &[T("Infinity"), T("Infinity"), T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT var_pop(x::float8)::text FROM (VALUES ('infinity'), ('infinity')) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("var_pop", TEXT)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT var_pop(x::float8)::text, var_samp(x::float8)::text, stddev_pop(x::float8)::text, stddev_samp(x::float8)::text FROM (VALUES ('infinity')) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("var_pop", TEXT), Column("var_samp", TEXT), Column("stddev_pop", TEXT), Column("stddev_samp", TEXT)],
                        rows: &[
                            &[T("NaN"), Null, T("NaN"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT var_pop(x::float8)::text, var_samp(x::float8)::text, stddev_pop(x::float8)::text, stddev_samp(x::float8)::text FROM (VALUES ('nan')) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("var_pop", TEXT), Column("var_samp", TEXT), Column("stddev_pop", TEXT), Column("stddev_samp", TEXT)],
                        rows: &[
                            &[T("NaN"), Null, T("NaN"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT var_pop(x::float8), var_samp(x::float8), stddev_pop(x::float8), stddev_samp(x::float8) FROM (VALUES (1::float8), (2::float8), (3::float8), (4::float8)) v(x);",
                    expected: Expected::Rows {
                        columns: &[Column("var_pop", FLOAT8), Column("var_samp", FLOAT8), Column("stddev_pop", FLOAT8), Column("stddev_samp", FLOAT8)],
                        rows: &[
                            &[T("1.25"), T("1.6666666666666667"), T("1.118033988749895"), T("1.2909944487358056")],
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
fn test_array_contained() {
    run_scripts(&[
        ScriptTest {
            name: "arraycontained",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[4,1] <@ ARRAY[[1,2],[3,4]],ARRAY[5] <@ ARRAY[1,2],ARRAY[]::int[] <@ ARRAY[1];",
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
                    query: "SELECT id, a <@ ARRAY[1,3], ARRAY[]::int[] <@ a FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("f"), T("t")],
                            &[T("2"), T("t"), T("t")],
                            &[T("3"), Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT ARRAY[3,3]) <@ (SELECT a FROM array_inputs WHERE id=1), ARRAY[NULL]::int[] <@ ARRAY[NULL]::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f")],
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
fn test_array_contains() {
    run_scripts(&[
        ScriptTest {
            name: "arraycontains",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[1,2],[3,4]] @> ARRAY[4,1],ARRAY[1] @> ARRAY[1,1],ARRAY[NULL]::int[] @> ARRAY[NULL]::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY['red','blue']::varchar[] @> ARRAY['red']::varchar[],ARRAY[1] @> ARRAY[]::int[],NULL::int[] @> ARRAY[1];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, a @> ARRAY[3,3], a @> ARRAY[]::int[] FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("t"), T("t")],
                            &[T("2"), T("f"), T("t")],
                            &[T("3"), Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT a FROM array_inputs WHERE id=1) @> (SELECT ARRAY[NULL]::int[]), ARRAY[]::int[] @> ARRAY[]::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("t")],
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
fn test_array_fill() {
    run_scripts(&[
        ScriptTest {
            name: "array_fill",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_fill(7,ARRAY[2,3]),array_fill(NULL::int,ARRAY[2,2]),array_fill('x'::varchar,ARRAY[2],ARRAY[1]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_fill", INT4_ARRAY), Column("array_fill", INT4_ARRAY), Column("array_fill", VARCHAR_ARRAY)],
                        rows: &[
                            &[T("{{7,7,7},{7,7,7}}"), T("{{NULL,NULL},{NULL,NULL}}"), T("{x,x}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[]::int[]),array_fill(1,ARRAY[2,0]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_fill", INT4_ARRAY), Column("array_fill", INT4_ARRAY)],
                        rows: &[
                            &[T("{}"), T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[1,1,1,1,1,1,1]);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "number of array dimensions (7) exceeds the maximum allowed (6)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[NULL]::int[]);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "dimension values cannot be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,NULL::int[]);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "dimension array or low bound array cannot be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[2],ARRAY[0]);",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[[2,2]]);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "wrong number of array subscripts", detail: "Dimension array must be one dimensional.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[2147483647,2]);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "array size exceeds the maximum allowed (134217727)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, array_fill(id,ARRAY[cardinality(a)]) FROM array_inputs WHERE id<3 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("array_fill", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{1,1,1,1}")],
                            &[T("2"), T("{}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill((SELECT id FROM array_inputs WHERE id=1),(SELECT ARRAY[2,1]),(SELECT ARRAY[1,1]));",
                    expected: Expected::Rows {
                        columns: &[Column("array_fill", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1},{1}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[2],NULL::int[]);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "dimension array or low bound array cannot be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[2],ARRAY[NULL]::int[]);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "dimension values cannot be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[2,2],ARRAY[1]);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "wrong number of array subscripts", detail: "Low bound array has different size than dimensions array.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(1,ARRAY[-1]);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "array size exceeds the maximum allowed (134217727)", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_array_functions() {
    run_scripts(&[
        ScriptTest {
            name: "unnest",
            set_up_script: &[
                "CREATE TABLE testing (id INT primary key, val1 smallint[]);",
                "INSERT INTO testing VALUES (1, '{}'), (2, '{1}'), (3, '{1, 2}');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT unnest(val1) FROM testing WHERE id=1;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT2)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(val1) FROM testing WHERE id=2;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT2)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(val1) FROM testing WHERE id=3;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT2)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from unnest(array[1,2,3]);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4)],
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
                    query: "SELECT * FROM unnest(ARRAY[1, 2]::integer[], ARRAY[3, 4]::integer[]);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4), Column("unnest", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("2"), T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[1, 2, 3], ARRAY['a', 'b']::text[]);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4), Column("unnest", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[]::int[], ARRAY['a']);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4), Column("unnest", TEXT)],
                        rows: &[
                            &[Null, T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(NULL::int[], ARRAY['a']);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4), Column("unnest", TEXT)],
                        rows: &[
                            &[Null, T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM unnest(ARRAY[1, 2], ARRAY['a', 'b']) AS t(a, b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[1, 2], ARRAY['a', 'b'], ARRAY[true, false]);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4), Column("unnest", TEXT), Column("unnest", BOOL)],
                        rows: &[
                            &[T("1"), T("a"), T("t")],
                            &[T("2"), T("b"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM (SELECT 1) s, unnest(ARRAY[1, 2], ARRAY['a', 'b']);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("unnest", INT4), Column("unnest", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("a")],
                            &[T("1"), T("2"), T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, a, b FROM testing, unnest(val1, ARRAY['x', 'y', 'z']) AS t(a, b) ORDER BY id, b;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT2), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), Null, T("x")],
                            &[T("1"), Null, T("y")],
                            &[T("1"), Null, T("z")],
                            &[T("2"), T("1"), T("x")],
                            &[T("2"), Null, T("y")],
                            &[T("2"), Null, T("z")],
                            &[T("3"), T("1"), T("x")],
                            &[T("3"), T("2"), T("y")],
                            &[T("3"), Null, T("z")],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[1, 2], 5);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function pg_catalog.unnest(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 15, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT t.* FROM unnest(ARRAY[1, 2], ARRAY['a', 'b']) t;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4), Column("unnest", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(ARRAY[1, 2], ARRAY[3, 4]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function unnest(integer[], integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[1, 2], ARRAY['a', 'b']) WITH ORDINALITY;",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4), Column("unnest", TEXT), Column("ordinality", INT8)],
                        rows: &[
                            &[T("1"), T("a"), T("1")],
                            &[T("2"), T("b"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[1, 2], ARRAY['a', 'b']) WITH ORDINALITY AS t(x, y, n);",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("y", TEXT), Column("n", INT8)],
                        rows: &[
                            &[T("1"), T("a"), T("1")],
                            &[T("2"), T("b"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, a, b, n FROM testing, unnest(val1, ARRAY['x', 'y']) WITH ORDINALITY AS t(a, b, n) ORDER BY id, n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT2), Column("b", TEXT), Column("n", INT8)],
                        rows: &[
                            &[T("1"), Null, T("x"), T("1")],
                            &[T("1"), Null, T("y"), T("2")],
                            &[T("2"), T("1"), T("x"), T("1")],
                            &[T("2"), Null, T("y"), T("2")],
                            &[T("3"), T("1"), T("x"), T("1")],
                            &[T("3"), T("2"), T("y"), T("2")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_to_json",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY[1, 2, 3])",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT array_to_json(ARRAY [jsonb '{"a":1}', jsonb '{"b":[2,3]}']);"#,
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T(r#"[{"a": 1},{"b": [2, 3]}]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY[1.5, 2.5]::float8[])",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T("[1.5,2.5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY[1e20::float8, 1e-7::float8])",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T("[1e+20,1e-07]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY[true, false])",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T("[true,false]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY[DATE '2024-02-29', DATE '2025-01-01'])",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T(r#"["2024-02-29","2025-01-01"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY['a', 'b', 'c'])",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T(r#"["a","b","c"]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY[1, NULL, 3])",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T("[1,null,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json('{{1,2},{3,4}}'::int[])",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T("[[1,2],[3,4]]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY[1, 2, 3], false);",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_json(ARRAY[1, 2, 3], true)",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_json", JSON)],
                        rows: &[
                            &[T(r#"[1,
 2,
 3]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_to_string",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_to_string(ARRAY[1, 2, 3, NULL, 5], ',', '*')",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("1,2,3,*,5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_string(ARRAY[1, 2, 3, NULL, 5], ',')",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("1,2,3,5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_string(ARRAY[37.89, 1.2], '_');",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("37.89_1.2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_string(ARRAY[37.89::int4, 1.2::int4], '_');",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("38_1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "array_to_string on vector types (int2vector, oidvector)",
            set_up_script: &[
                "CREATE TABLE vectest (pk INT PRIMARY KEY, a INT, b INT);",
                "CREATE INDEX vectest_ab ON vectest (a, b);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_to_string(indkey, ',') FROM pg_index WHERE indexrelid = 'vectest_ab'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("2,3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_string(indkey, ',') FROM pg_index WHERE indexrelid = 'vectest_pkey'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_string(indclass, ',') FROM pg_index WHERE indexrelid = 'vectest_ab'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("1978,1978")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "string_to_array",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT string_to_array('a,b,c', ',');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{a,b,c}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array('xx~^~yy~^~zz', '~^~', 'yy');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{xx,NULL,zz}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array('abc', NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{a,b,c}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array('abc', NULL, 'b');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{a,NULL,c}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array('abc', '');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{abc}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array('', ',');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array('', NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array(NULL, ',');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array(NULL, ',', 'a');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array('a,b,c', ',', NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{a,b,c}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array(',a,,b,', ',');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{"",a,"",b,""}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array(',a,', ',', '');",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", TEXT_ARRAY)],
                        rows: &[
                            &[T("{NULL,a,NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_to_array('1,2,3', ',')::int4[];",
                    expected: Expected::Rows {
                        columns: &[Column("string_to_array", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_length(string_to_array('a,b,c', ','), 1);",
                    expected: Expected::Rows {
                        columns: &[Column("array_length", INT4)],
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
            name: "array_upper",
            assertions: &[
                ScriptTestAssertion {
                    query: "select array_upper(ARRAY[1,2,3,4], 1);",
                    expected: Expected::Rows {
                        columns: &[Column("array_upper", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_upper(ARRAY[1,2,3,4], 2);",
                    expected: Expected::Rows {
                        columns: &[Column("array_upper", INT4)],
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
            name: "array_cat",
            assertions: &[
                ScriptTestAssertion {
                    query: "select array_cat(ARRAY[1,2,3], ARRAY[4,5]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_cat", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3,4,5}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_cat(NULL, ARRAY[4,5]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_cat", INT4_ARRAY)],
                        rows: &[
                            &[T("{4,5}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_cat(ARRAY[1,2,3], NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("array_cat", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_cat(NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("array_cat", TEXT_ARRAY)],
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
            name: "array_length",
            assertions: &[
                ScriptTestAssertion {
                    query: "select array_length(ARRAY[1,2,3,4,5], 1);",
                    expected: Expected::Rows {
                        columns: &[Column("array_length", INT4)],
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
            name: "array_position and array_positions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_position(ARRAY[1,2,3,4,5], 4);",
                    expected: Expected::Rows {
                        columns: &[Column("array_position", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_position(ARRAY[1,4,2,3,4,5,4], 4, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("array_position", INT4)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_position(NULL, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("array_position", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_position(ARRAY[1,4,2,3,4,5,4], NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("array_position", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_position(NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("array_position", INT4)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_positions(ARRAY[1,2,3,4,5,6,1,2,3,4,5,6], 4);",
                    expected: Expected::Rows {
                        columns: &[Column("array_positions", INT4_ARRAY)],
                        rows: &[
                            &[T("{4,10}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_positions(NULL, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("array_positions", INT4_ARRAY)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_positions(ARRAY[1,4,2,3,4,5,4], NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("array_positions", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select array_positions(NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("array_positions", INT4_ARRAY)],
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
            name: "array_prepend",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_prepend(NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("array_prepend", TEXT_ARRAY)],
                        rows: &[
                            &[T("{NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_prepend(NULL, ARRAY[6]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_prepend", INT4_ARRAY)],
                        rows: &[
                            &[T("{NULL,6}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_prepend(5, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("array_prepend", INT4_ARRAY)],
                        rows: &[
                            &[T("{5}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_prepend(5, ARRAY[6]);",
                    expected: Expected::Rows {
                        columns: &[Column("array_prepend", INT4_ARRAY)],
                        rows: &[
                            &[T("{5,6}")],
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
fn test_array_lower() {
    run_scripts(&[
        ScriptTest {
            name: "array_lower",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_lower(ARRAY[[1,2],[3,4]],1),array_lower(ARRAY[[1,2],[3,4]],2),array_lower(ARRAY[1],2);",
                    expected: Expected::Rows {
                        columns: &[Column("array_lower", INT4), Column("array_lower", INT4), Column("array_lower", INT4)],
                        rows: &[
                            &[T("1"), T("1"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_lower(ARRAY[]::int[],1),array_lower(NULL::int[],1),array_lower(ARRAY[1],0),array_lower('1 2'::int2vector,1);",
                    expected: Expected::Rows {
                        columns: &[Column("array_lower", INT4), Column("array_lower", INT4), Column("array_lower", INT4), Column("array_lower", INT4)],
                        rows: &[
                            &[Null, Null, Null, T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, array_lower(a,1) FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("array_lower", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), Null],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_lower((SELECT a FROM array_inputs WHERE id=1),NULL), array_lower(ARRAY[1],-1), array_lower('1 2'::oidvector,1);",
                    expected: Expected::Rows {
                        columns: &[Column("array_lower", INT4), Column("array_lower", INT4), Column("array_lower", INT4)],
                        rows: &[
                            &[Null, Null, T("0")],
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
fn test_array_overlap() {
    run_scripts(&[
        ScriptTest {
            name: "arrayoverlap",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[[1,2],[3,4]] && ARRAY[4,9],ARRAY[1,2] && ARRAY[9],ARRAY[NULL]::int[] && ARRAY[NULL]::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[]::int[] && ARRAY[1],NULL::int[] && ARRAY[1],ARRAY['a']::varchar[] && ARRAY['b','a']::varchar[];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), Null, T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, a && ARRAY[1], a && ARRAY[NULL]::int[] FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("t"), T("f")],
                            &[T("2"), T("f"), T("f")],
                            &[T("3"), Null, Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (SELECT a FROM array_inputs WHERE id=1) && (SELECT ARRAY[9]), ARRAY[]::int[] && ARRAY[]::int[];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("f"), T("f")],
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
fn test_array_remove() {
    run_scripts(&[
        ScriptTest {
            name: "array_remove",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_remove(ARRAY[1,2,1,NULL],1), array_remove(ARRAY[1,NULL,2],NULL), array_remove(NULL::int[],1);",
                    expected: Expected::Rows {
                        columns: &[Column("array_remove", INT4_ARRAY), Column("array_remove", INT4_ARRAY), Column("array_remove", INT4_ARRAY)],
                        rows: &[
                            &[T("{2,NULL}"), T("{1,2}"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_remove(ARRAY[1,1],1), array_remove(ARRAY[]::text[],'a');",
                    expected: Expected::Rows {
                        columns: &[Column("array_remove", INT4_ARRAY), Column("array_remove", TEXT_ARRAY)],
                        rows: &[
                            &[T("{}"), T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_remove(ARRAY[[1,2],[3,4]],2);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "removing elements from multidimensional arrays is not supported", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, array_remove(a,3) FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("array_remove", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{NULL,1}")],
                            &[T("2"), T("{}")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_remove((SELECT a FROM array_inputs WHERE id=1),(SELECT NULL::int));",
                    expected: Expected::Rows {
                        columns: &[Column("array_remove", INT4_ARRAY)],
                        rows: &[
                            &[T("{3,1,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_remove(ARRAY['NaN'::numeric,1,'NaN'::numeric],'NaN'::numeric),array_remove(ARRAY['a','b'],'z');",
                    expected: Expected::Rows {
                        columns: &[Column("array_remove", NUMERIC_ARRAY), Column("array_remove", TEXT_ARRAY)],
                        rows: &[
                            &[T("{1}"), T("{a,b}")],
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
fn test_array_replace() {
    run_scripts(&[
        ScriptTest {
            name: "array_replace",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_replace(ARRAY[[1,NULL],[1,4]],1,9), array_replace(ARRAY[[1,NULL],[1,4]],NULL,0);",
                    expected: Expected::Rows {
                        columns: &[Column("array_replace", INT4_ARRAY), Column("array_replace", INT4_ARRAY)],
                        rows: &[
                            &[T("{{9,NULL},{9,4}}"), T("{{1,0},{1,4}}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_replace(ARRAY['a','b'],'a',NULL), array_replace(NULL::int[],1,2), array_replace(ARRAY[]::int[],1,2);",
                    expected: Expected::Rows {
                        columns: &[Column("array_replace", TEXT_ARRAY), Column("array_replace", INT4_ARRAY), Column("array_replace", INT4_ARRAY)],
                        rows: &[
                            &[T("{NULL,b}"), Null, T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, array_replace(a,3,NULL) FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("array_replace", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{NULL,NULL,1,NULL}")],
                            &[T("2"), T("{}")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_replace((SELECT a FROM array_inputs WHERE id=1),(SELECT NULL::int),(SELECT 2));",
                    expected: Expected::Rows {
                        columns: &[Column("array_replace", INT4_ARRAY)],
                        rows: &[
                            &[T("{3,2,1,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_replace(ARRAY[NULL,NULL]::int[],NULL,NULL),array_replace(ARRAY[1,2],9,0);",
                    expected: Expected::Rows {
                        columns: &[Column("array_replace", INT4_ARRAY), Column("array_replace", INT4_ARRAY)],
                        rows: &[
                            &[T("{NULL,NULL}"), T("{1,2}")],
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
fn test_array_reverse() {
    run_scripts(&[
        ScriptTest {
            name: "array_reverse",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_reverse(ARRAY[[2,4],[3,1],[1,9]]), array_reverse(ARRAY[1,NULL,2]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_reverse(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_reverse(ARRAY[]::int[]),array_reverse(NULL::int[]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_reverse(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, array_reverse(a) FROM array_inputs ORDER BY id;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_reverse(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 12, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_reverse((SELECT a FROM array_inputs WHERE id=1));",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_reverse(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_reverse(ARRAY[NULL]::int[]),array_reverse(ARRAY[[[1,2]],[[3,4]]]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_reverse(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_array_sort() {
    run_scripts(&[
        ScriptTest {
            name: "array_sort",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_sort(ARRAY[[2,4],[3,1],[1,9]]), array_sort(ARRAY[3,NULL,1,2]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_sort(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_sort(ARRAY[3,NULL,1],true), array_sort(ARRAY[3,NULL,1],true,false), array_sort(ARRAY[3,NULL,1],false,true);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_sort(integer[], boolean) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_sort(ARRAY[[1,NULL],[1,2],[NULL,1]]), array_sort(ARRAY[]::int[]), array_sort(NULL::int[]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_sort(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_sort(ARRAY['z','a','m']),array_sort(ARRAY[1],NULL);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_sort(text[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, array_sort(a),array_sort(a,true),array_sort(a,true,false) FROM array_inputs ORDER BY id;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_sort(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 12, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_sort((SELECT a FROM array_inputs WHERE id=1)),array_sort((SELECT a FROM array_inputs WHERE id=1),false),array_sort((SELECT a FROM array_inputs WHERE id=1),false,true);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_sort(integer[]) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_sort(ARRAY[[1,NULL],[1,NULL],[1,2]],true,false),array_sort(ARRAY[1],false,NULL),array_sort(ARRAY[NULL,NULL]::int[]);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function array_sort(integer[], boolean, boolean) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_cardinality() {
    run_scripts(&[
        ScriptTest {
            name: "cardinality",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT cardinality(ARRAY[[1,NULL],[3,4]]), cardinality(ARRAY[]::int[]), cardinality(NULL::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("cardinality", INT4), Column("cardinality", INT4), Column("cardinality", INT4)],
                        rows: &[
                            &[T("4"), T("0"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cardinality(ARRAY[[[1,2]],[[3,4]]]);",
                    expected: Expected::Rows {
                        columns: &[Column("cardinality", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, cardinality(a) FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("cardinality", INT4)],
                        rows: &[
                            &[T("1"), T("4")],
                            &[T("2"), T("0")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cardinality((SELECT a FROM array_inputs WHERE id=1));",
                    expected: Expected::Rows {
                        columns: &[Column("cardinality", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cardinality(ARRAY[[[[[[NULL::int]]]]]]);",
                    expected: Expected::Rows {
                        columns: &[Column("cardinality", INT4)],
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

#[test]
fn test_date_and_time_function() {
    run_scripts(&[
        ScriptTest {
            name: "extract from date",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(CENTURY FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("21")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(CENTURY FROM DATE '0002-12-31 BC');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DAY FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DECADE FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("202")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOW FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOY FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("33")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(EPOCH FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("1643760000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(HOUR FROM DATE '2022-02-02');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "hour" not supported for type date"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISODOW FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISOYEAR FROM DATE '2006-01-01');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2005")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISOYEAR FROM DATE '2006-01-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2006")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT extract(julian from date '2021-06-23');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2459389")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MICROSECONDS FROM DATE '2022-02-02');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "microseconds" not supported for type date"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLENNIUM FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLISECONDS FROM DATE '2022-02-02');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "milliseconds" not supported for type date"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MINUTE FROM DATE '2022-02-02');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "minute" not supported for type date"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MONTH FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(QUARTER FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(SECOND FROM DATE '2022-02-02');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "second" not supported for type date"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE FROM DATE '2022-02-02');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone" not supported for type date"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_HOUR FROM DATE '2022-02-02');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone_hour" not supported for type date"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_MINUTE FROM DATE '2022-02-02');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone_minute" not supported for type date"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(WEEK FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(YEAR FROM DATE '2022-02-02');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2022")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "extract from time without time zone",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(CENTURY FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "century" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DAY FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "day" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DECADE FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "decade" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOW FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "dow" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOY FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "doy" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(EPOCH FROM TIME '17:12:28.5');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("61948.500000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(HOUR FROM TIME '17:12:28.5');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("17")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISODOW FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "isodow" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISOYEAR FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "isoyear" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(JULIAN FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "julian" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MICROSECONDS FROM TIME '17:12:28.5');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("28500000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLENNIUM FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "millennium" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLISECONDS FROM TIME '17:12:28.5');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("28500.000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MINUTE FROM TIME '17:12:28.5');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MONTH FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "month" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(QUARTER FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "quarter" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(SECOND FROM TIME '17:12:28.5');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("28.500000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_HOUR FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone_hour" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_MINUTE FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone_minute" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(WEEK FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "week" not supported for type time without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(YEAR FROM TIME '17:12:28.5');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "year" not supported for type time without time zone"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "extract from time with time zone",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(CENTURY FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "century" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DAY FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "day" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DECADE FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "decade" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOW FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "dow" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOY FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "doy" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(EPOCH FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("72748.500000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(HOUR FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("17")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISODOW FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "isodow" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISOYEAR FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "isoyear" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(JULIAN FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "julian" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MICROSECONDS FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("28500000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLENNIUM FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "millennium" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLISECONDS FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("28500.000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MINUTE FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MONTH FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "month" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(QUARTER FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "quarter" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(SECOND FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("28.500000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE FROM TIME WITH TIME ZONE '17:12:28.5+03');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("10800")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_HOUR FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("-3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_MINUTE FROM TIME WITH TIME ZONE '17:12:28.5-03:45');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("-45")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(WEEK FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "week" not supported for type time with time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(YEAR FROM TIME WITH TIME ZONE '17:12:28.5-03');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "year" not supported for type time with time zone"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "extract from timestamp without time zone",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(CENTURY FROM TIMESTAMP '2000-12-16 12:21:13');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(CENTURY FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("21")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DAY FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DECADE FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOW FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOY FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("47")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(EPOCH FROM TIMESTAMP '2001-02-16 20:38:40.12');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("982355920.120000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(HOUR FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISODOW FROM TIMESTAMP '2001-02-18 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISOYEAR FROM TIMESTAMP '2001-02-18 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(JULIAN FROM TIMESTAMP '2001-02-18 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2451959.86018518518518518519")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MICROSECONDS FROM TIMESTAMP '2001-02-18 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("40000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLENNIUM FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLENNIUM FROM TIMESTAMP '2000-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLISECONDS FROM TIMESTAMP '2000-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("40000.000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MINUTE FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("38")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MONTH FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(QUARTER FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(SECOND FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("40.000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone" not supported for type timestamp without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_HOUR FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone_hour" not supported for type timestamp without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_MINUTE FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"unit "timezone_minute" not supported for type timestamp without time zone"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(WEEK FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(YEAR FROM TIMESTAMP '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "extract from timestamp with time zone",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET TIMEZONE TO 'UTC';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(CENTURY FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("21")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DAY FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DECADE FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOW FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DOY FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("47")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(EPOCH FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("982345120.120000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(HOUR FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("17")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISODOW FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(ISOYEAR FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(JULIAN FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2451957.73518657407407407407")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT extract(julian from '2021-06-23 7:00:00-04'::timestamptz at time zone 'UTC+12');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2459388.95833333333333333333")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT extract(julian from '2021-06-23 8:00:00-04'::timestamptz at time zone 'UTC+12');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2459389.0000000000000000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MICROSECONDS FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("40120000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLENNIUM FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLISECONDS FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("40120.000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MINUTE FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("38")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MONTH FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(QUARTER FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(SECOND FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("40.120000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_HOUR FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(TIMEZONE_MINUTE FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05:45');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(WEEK FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(YEAR FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET TIMEZONE TO DEFAULT;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "extract from interval",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(CENTURY FROM INTERVAL '2001 years');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(DAY FROM INTERVAL '40 days 1 minute');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(decades from interval '1000 months');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(EPOCH FROM INTERVAL '5 days 3 hours');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("442800.000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(epoch from interval '10 months 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("25920010.000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(hours from interval '10 months 65 minutes 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(microsecond from interval '10 months 65 minutes 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("10000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MILLENNIUM FROM INTERVAL '2001 years');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(millenniums from interval '3000 years 65 minutes 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(millisecond from interval '10 months 65 minutes 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("10000.000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(minutes from interval '10 months 65 minutes 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MONTH FROM INTERVAL '2 years 3 months');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT EXTRACT(MONTH FROM INTERVAL '2 years 13 months');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(months from interval '20 months 65 minutes 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(quarter from interval '20 months 65 minutes 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(seconds from interval '65 minutes 10 seconds 5 millisecond');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
                        rows: &[
                            &[T("10.005000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select extract(years from interval '20 months 65 minutes 10 seconds');",
                    expected: Expected::Rows {
                        columns: &[Column("extract", NUMERIC)],
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
            name: "age",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT age(timestamp '2001-04-10', timestamp '1957-06-13');",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("43 years 9 mons 27 days")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(timestamp '1957-06-13', timestamp '2001-04-10');",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("-43 years -9 mons -27 days")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(timestamp '2001-06-13', timestamp '2001-04-10');",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("2 mons 3 days")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(timestamp '2001-04-10', timestamp '2001-06-13');",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("-2 mons -3 days")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(timestamp '2001-04-10 12:23:33', timestamp '1957-06-13 13:23:34.4');",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("43 years 9 mons 26 days 22:59:58.6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(timestamp '1957-06-13 13:23:34.4', timestamp '2001-04-10 12:23:33');",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("-43 years -9 mons -26 days -22:59:58.6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(current_date);",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(current_date::timestamp);",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(timestamptz '2013-07-01 12:00:00', timestamptz '2013-03-01 12:00:00');",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("4 mons")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT age(timestamptz '2013-03-01 00:00:00.500 UTC', timestamptz '2013-01-31 23:59:59.250 UTC');",
                    expected: Expected::Rows {
                        columns: &[Column("age", INTERVAL)],
                        rows: &[
                            &[T("28 days 00:00:01.25")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "timezone",
            set_up_script: &[
                "SET timezone = '+06:30'",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select timezone(interval '2 minutes', timestamp with time zone '2001-02-16 20:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMP)],
                        rows: &[
                            &[T("2001-02-17 01:40:40.12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timezone('UTC', timestamp with time zone '2001-02-16 20:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMP)],
                        rows: &[
                            &[T("2001-02-17 01:38:40.12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timezone('-04:45', time with time zone '20:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMETZ)],
                        rows: &[
                            &[T("06:23:40.12+04:45")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timezone(interval '2 hours 2 minutes', time with time zone '20:38:40.12-05');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMETZ)],
                        rows: &[
                            &[T("03:40:40.12+02:02")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timezone('-04:45', timestamp '2001-02-16 20:38:40.12');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-02-16 15:53:40.12+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timezone('-04:45:44', timestamp '2001-02-16 20:38:40.12');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-02-16 15:52:56.12+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '2001-02-16 20:38:40.12'::timestamp at time zone '-04:45:44';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-02-16 15:52:56.12+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timezone(interval '2 hours 2 minutes', timestamp '2001-02-16 20:38:40.12');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-02-16 18:36:40.12+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '2024-08-22 14:47:57 -07' at time zone 'utc';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMP)],
                        rows: &[
                            &[T("2024-08-22 21:47:57")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select round(extract(epoch from '2024-08-22 13:47:57-07' at time zone 'UTC')) as startup_time;",
                    expected: Expected::Rows {
                        columns: &[Column("startup_time", NUMERIC)],
                        rows: &[
                            &[T("1724359677")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timestamptz '2024-08-22 13:47:57-07' at time zone 'utc';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMP)],
                        rows: &[
                            &[T("2024-08-22 20:47:57")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timestamp '2024-08-22 13:47:57-07';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2024-08-22 13:47:57")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timestamp '2024-08-22 13:47:57-07' at time zone 'utc';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2024-08-22 13:47:57+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '2011-03-27 02:00:00'::timestamp at time zone '+01:00';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-03-27 03:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '2011-03-27 02:00:00'::timestamp at time zone 'UTC';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-03-27 02:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select timezone('MSK', timestamp '2011-03-27 02:00:00');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-03-26 22:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select '2011-03-27 02:00:00'::timestamp at time zone 'MSK';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-03-26 22:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "date_part",
            assertions: &[
                ScriptTestAssertion {
                    query: "select date_part('month', date '2001-02-16');",
                    expected: Expected::Rows {
                        columns: &[Column("date_part", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select date_part('minute', time without time zone '20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("date_part", FLOAT8)],
                        rows: &[
                            &[T("38")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select date_part('second', time with time zone '20:38:40 UTC');",
                    expected: Expected::Rows {
                        columns: &[Column("date_part", FLOAT8)],
                        rows: &[
                            &[T("40")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select date_part('year', timestamp without time zone '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("date_part", FLOAT8)],
                        rows: &[
                            &[T("2001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select date_part('day', timestamp with time zone '2001-02-16 20:38:40 UTC');",
                    expected: Expected::Rows {
                        columns: &[Column("date_part", FLOAT8)],
                        rows: &[
                            &[T("16")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select date_part('month', interval '2 years 3 months');",
                    expected: Expected::Rows {
                        columns: &[Column("date_part", FLOAT8)],
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
            name: "date_trunc",
            assertions: &[
                ScriptTestAssertion {
                    query: "select date_trunc('hour', timestamp '2001-02-16 20:38:40');",
                    expected: Expected::Rows {
                        columns: &[Column("date_trunc", TIMESTAMP)],
                        rows: &[
                            &[T("2001-02-16 20:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone to '+06:30';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select date_trunc('day', timestamp with time zone '2001-02-16 20:38:40 UTC');",
                    expected: Expected::Rows {
                        columns: &[Column("date_trunc", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-02-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select date_trunc('day', timestamp with time zone '2001-02-16 20:38:40 UTC', '-07:00');",
                    expected: Expected::Rows {
                        columns: &[Column("date_trunc", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2001-02-16 17:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone to '+06:30';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select date_trunc('hour', interval '2 days 10 hours 30 minutes');",
                    expected: Expected::Rows {
                        columns: &[Column("date_trunc", INTERVAL)],
                        rows: &[
                            &[T("2 days 10:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_date",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_date('1 4 1902', 'Q MM YYYY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("1902-04-01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('3 4 21 01', 'W MM CC YY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("2001-04-15")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('2458872', 'J');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("2020-01-23")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('44-02-01 BC','YYYY-MM-DD BC');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("0044-02-01 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('-44-02-01','YYYY-MM-DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("0044-02-01 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('2011x 12x 18', 'YYYYxMMxDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("2011-12-18")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('2015 365', 'YYYY DDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("2015-12-31")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_timestamp",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET timezone to '+06:30';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 23:38:15', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:15+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000-01-01 12:30:45', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-01-01 19:00:45+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('0097/Feb/16 --> 08:14:30', 'YYYY/Mon/DD --> HH:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("0097-02-16 14:44:30+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('97/2/16 8:14:30', 'FMYYYY/FMMM/FMDD FMHH:FMMI:FMSS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("0097-02-16 14:44:30+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011$03!18 23_38_15', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-03-19 06:08:15+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('1985 January 12', 'YYYY FMMonth DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1985-01-12 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_timestamp('1985 FMMonth 12', 'YYYY "FMMonth" DD');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1985-01-12 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_timestamp('1985 \\ 12', 'YYYY \\\\ DD');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1985-01-12 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_timestamp('My birthday-> Year: 1976, Month: May, Day: 16', '"My birthday-> Year:" YYYY, "Month:" FMMonth, "Day:" DD');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1976-05-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('1,582nd VIII 21', 'Y,YYYth FMRM DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1582-08-21 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_timestamp('15 "text between quote marks" 98 54 45',
				  E'HH24 "\\"text between quote marks\\"" YY MI SS');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1998-01-01 22:24:45+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('05121445482000', 'MMDDHH24MISSYYYY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-05-12 21:15:48+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000January09Sunday', 'YYYYFMMonthDDFMDay');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-01-09 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('97/Feb/16', 'YYMonDD');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid value "/Feb/16" for "Mon""#, detail: "The given value did not match any of the allowed values for this field.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('97/Feb/16', 'YY:Mon:DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-02-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('97/Feb/16', 'FXYY:Mon:DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-02-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('97/Feb/16', 'FXYY/Mon/DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-02-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('19971116', 'YYYYMMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-11-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('20000-1116', 'FXYYYY-MMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("20000-11-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('1997 AD 11 16', 'YYYY BC MM DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-11-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('1997 BC 11 16', 'YYYY BC MM DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-11-16 06:30:00+00 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('1997 A.D. 11 16', 'YYYY B.C. MM DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-11-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('1997 B.C. 11 16', 'YYYY B.C. MM DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1997-11-16 06:30:00+00 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('9-1116', 'Y-MMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2009-11-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('95-1116', 'YY-MMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1995-11-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('995-1116', 'YYY-MMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1995-11-16 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2005426', 'YYYYWWD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2005-10-15 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2005300', 'YYYYDDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2005-10-27 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2005527', 'IYYYIWID');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2006-01-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('005527', 'IYYIWID');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2006-01-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('05527', 'IYIWID');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2006-01-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('5527', 'IIWID');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2006-01-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2005364', 'IYYYIDDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2006-01-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('20050302', 'YYYYMMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2005-03-02 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2005 03 02', 'YYYYMMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2005-03-02 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp(' 2005 03 02', 'YYYYMMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2005-03-02 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('  20050302', 'YYYYMMDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2005-03-02 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 AM', 'YYYY-MM-DD HH12:MI PM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-18 18:08:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 PM', 'YYYY-MM-DD HH12:MI PM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 A.M.', 'YYYY-MM-DD HH12:MI P.M.');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-18 18:08:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 P.M.', 'YYYY-MM-DD HH12:MI P.M.');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 +05', 'YYYY-MM-DD HH12:MI TZH');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-18 06:38:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 -05', 'YYYY-MM-DD HH12:MI TZH');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-18 16:38:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 +05:20', 'YYYY-MM-DD HH12:MI TZH:TZM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-18 06:18:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 -05:20', 'YYYY-MM-DD HH12:MI TZH:TZM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-18 16:58:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 20', 'YYYY-MM-DD HH12:MI TZM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-18 11:18:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 11:38 PST', 'YYYY-MM-DD HH12:MI TZ');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"formatting field "TZ" is only supported in to_char"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2018-11-02 12:34:56.025', 'YYYY-MM-DD HH24:MI:SS.MS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2018-11-02 19:04:56.025+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('44-02-01 11:12:13 BC','YYYY-MM-DD HH24:MI:SS BC');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("0044-02-01 17:42:13+00 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('-44-02-01 11:12:13','YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("0044-02-01 17:42:13+00 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('-44-02-01 11:12:13 BC','YYYY-MM-DD HH24:MI:SS BC');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("0044-02-01 17:42:13+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18 23:38:15', 'YYYY-MM-DD  HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:15+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18  23:38:15', 'YYYY-MM-DD  HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:15+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18   23:38:15', 'YYYY-MM-DD  HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:15+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18  23:38:15', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:15+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18  23:38:15', 'YYYY-MM-DD  HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:15+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2011-12-18  23:38:15', 'YYYY-MM-DD   HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2011-12-19 06:08:15+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000+   JUN', 'YYYY/MON');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-06-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('  2000 +JUN', 'YYYY/MON');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-06-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp(' 2000 +JUN', 'YYYY//MON');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-06-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000  +JUN', 'YYYY//MON');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-06-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000 + JUN', 'YYYY MON');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-06-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000 ++ JUN', 'YYYY  MON');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-06-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000 + - JUN', 'YYYY  MON');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid value "-" for "MON""#, detail: "The given value did not match any of the allowed values for this field.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000 + + JUN', 'YYYY   MON');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-06-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000 -10', 'YYYY TZH');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2000-01-01 10:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2000 -10', 'YYYY  TZH');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1999-12-31 14:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2005527', 'YYYYIWID');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: "invalid combination of date conventions", hint: "Do not mix Gregorian and ISO week date conventions in a formatting template.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('19971', 'YYYYMMDD');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"source string too short for "MM" formatting field"#, detail: "Field requires 2 characters, but only 1 remain.", hint: r#"If your source string is not fixed-width, try using the "FM" modifier."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('19971)24', 'YYYYMMDD');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid value "1)" for "MM""#, detail: "Field requires 2 characters, but only 1 could be parsed.", hint: r#"If your source string is not fixed-width, try using the "FM" modifier."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('Friday 1-January-1999', 'DY DD MON YYYY');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid value "da" for "DD""#, detail: "Value must be an integer.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('Fri 1-January-1999', 'DY DD MON YYYY');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid value "uary" for "YYYY""#, detail: "Value must be an integer.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('Fri 1-Jan-1999', 'DY DD MON YYYY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1999-01-01 06:30:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('1997-11-Jan-16', 'YYYY-MM-Mon-DD');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"conflicting values for "Mon" field in formatting string"#, detail: "This value contradicts a previous setting for the same field type.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('199711xy', 'YYYYMMDD');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid value "xy" for "DD""#, detail: "Value must be an integer.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('10000000000', 'FMYYYY');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"value for "YYYY" in source string is out of range"#, detail: "Value must be in the range -2147483648 to 2147483647.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2016-06-13 25:00:00', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2016-06-13 25:00:00""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2016-06-13 15:60:00', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2016-06-13 15:60:00""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2016-06-13 15:50:60', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2016-06-13 15:50:60""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2016-06-13 15:50:55', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2016-06-13 22:20:55+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2016-06-13 15:50:55', 'YYYY-MM-DD HH:MI:SS');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"hour "15" is invalid for the 12-hour clock"#, hint: "Use the 24-hour clock, or give an hour between 1 and 12.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2016-13-01 15:50:55', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2016-13-01 15:50:55""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2016-02-30 15:50:55', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2016-02-30 15:50:55""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2016-02-29 15:50:55', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2016-02-29 22:20:55+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2015-02-29 15:50:55', 'YYYY-MM-DD HH24:MI:SS');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2015-02-29 15:50:55""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2015-02-11 86000', 'YYYY-MM-DD SSSS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2015-02-12 06:23:20+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2015-02-11 86400', 'YYYY-MM-DD SSSS');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2015-02-11 86400""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2015-02-11 86000', 'YYYY-MM-DD SSSSS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2015-02-12 06:23:20+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2015-02-11 86400', 'YYYY-MM-DD SSSSS');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2015-02-11 86400""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone to '+06:30';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "current_time and now functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT now()::timetz::text = current_time()::text;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near ")""#, position: 43, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT now()::timetz(4)::text = current_time(5)::text;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("both roundings print the same text whenever the fifth fractional digit rounds to 0, about one run in ten"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(to_char(current_date + 'now'::timetz, 'HH24:MI:SS.USTZH:TZM'));",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("21")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length('now'::timetz::text) > length('now'::time::text);",
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
                    query: "SELECT length(to_char('now'::time, 'HH24:MI:SS.US'));",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("15")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "make_timestamp",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT make_timestamp(2014, 12, 28, 6, 30, 45.887);",
                    expected: Expected::Rows {
                        columns: &[Column("make_timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("2014-12-28 06:30:45.887")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_timestamp(-44, 3, 15, 12, 30, 15);",
                    expected: Expected::Rows {
                        columns: &[Column("make_timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("0044-03-15 12:30:15 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_timestamp(-1, 3, 15, 12, 30, 15);",
                    expected: Expected::Rows {
                        columns: &[Column("make_timestamp", TIMESTAMP)],
                        rows: &[
                            &[T("0001-03-15 12:30:15 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select make_timestamp(0, 7, 15, 12, 30, 15);",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "date field value out of range: 0-07-15", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select make_timestamp(2000, 0, 15, 12, 30, 15);",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "date field value out of range: 2000-00-15", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select make_timestamp(2000, 7, 32, 12, 30, 15);",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "date field value out of range: 2000-07-32", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select make_timestamp(2000, 7, 15, 25, 30, 15);",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "time field value out of range: 25:30:15", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select make_timestamp(2000, 7, 15, 2, 61, 15);",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "time field value out of range: 2:61:15", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select make_timestamp(2000, 7, 15, 25, 30, 61);",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "time field value out of range: 25:30:61", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "make_timestamptz",
            set_up_script: &[
                "SET timezone = '+06:30'",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT make_timestamptz(2014, 12, 28, 6, 30, 45.887);",
                    expected: Expected::Rows {
                        columns: &[Column("make_timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2014-12-28 13:00:45.887+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_timestamptz(-44, 3, 15, 12, 30, 15);",
                    expected: Expected::Rows {
                        columns: &[Column("make_timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("0044-03-15 19:00:15+00 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_timestamptz(-1, 3, 15, 12, 30, 15);",
                    expected: Expected::Rows {
                        columns: &[Column("make_timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("0001-03-15 19:00:15+00 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select make_timestamptz(0, 7, 15, 12, 30, 15);",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "date field value out of range: 0-07-15", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_timestamptz(1910, 12, 24, 0, 0, 0, 'Nehwon/Lankhmar');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"time zone "Nehwon/Lankhmar" not recognized"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_timestamptz(1881, 12, 10, 0, 0, 0, 'Europe/Paris') AT TIME ZONE 'UTC';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMP)],
                        rows: &[
                            &[T("1881-12-09 23:50:39")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_timestamptz(2008, 12, 10, 10, 10, 10, 'EST');",
                    expected: Expected::Rows {
                        columns: &[Column("make_timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2008-12-10 15:10:10+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "date_bin",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT date_bin('5 min'::interval, timestamp '2020-02-01 01:01:01', timestamp '2020-02-01 00:02:30');",
                    expected: Expected::Rows {
                        columns: &[Column("date_bin", TIMESTAMP)],
                        rows: &[
                            &[T("2020-02-01 00:57:30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date_bin('5 months'::interval, timestamp '2020-02-01 01:01:01', timestamp '2001-01-01');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "timestamps cannot be binned into intervals containing months or years", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT date_bin('0 days'::interval, timestamp '1970-01-01 01:00:00' , timestamp '1970-01-01 00:00:00');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "stride must be greater than zero", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "timestamp/timestamptz minus interval with day component",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-08-15 12:00:00' - interval '90 days';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2026-05-17 12:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '2026-08-15 12:00:00+00' - interval '90 days';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2026-05-17 12:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-08-15 12:00:00' - interval '1 hour';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2026-08-15 11:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (timestamp '2026-08-15 12:00:00' - interval '90 days') = timestamp '2026-08-15 12:00:00';",
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
            name: "timestamp/timestamptz plus/minus interval normalizes month-end overflow",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-03-31 12:00:00' - interval '1 month';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2026-02-28 12:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '2026-03-31 12:00:00+00' - interval '1 month';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2026-02-28 13:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-03-15 12:00:00' - interval '1 month';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2026-02-15 12:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-01-31 00:00:00' + interval '1 month';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2026-02-28 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "timestamp/timestamptz plus/minus interval preserves sub-second precision",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-08-15 12:00:00.750000' - interval '0.250000 seconds';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2026-08-15 12:00:00.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '2026-08-15 12:00:00.750000+00' - interval '0.250000 seconds';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2026-08-15 12:00:00.5+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-08-15 12:00:00.5' + interval '0.25 seconds';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2026-08-15 12:00:00.75")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "timestamp/timestamptz plus/minus interval errors on out-of-range results",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-01-01' + interval '1000000000 years';",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", position: 42, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-01-01' - interval '1000000000 years';",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", position: 42, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '2026-01-01+00' + interval '1000000000 years';",
                    expected: Expected::Error(Diagnostic { code: "22008", message: "interval out of range", position: 47, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-01-01' + interval '1 day';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("2026-01-02 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamp '2026-01-01' + interval '290000 years';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", TIMESTAMP)],
                        rows: &[
                            &[T("292026-01-01 00:00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "time zone abbreviations whose offsets changed over time",
            set_up_script: &[
                "SET TIME ZONE 'UTC'",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT timestamp '2010-06-01 12:00:00' AT TIME ZONE 'MSK', timestamp '2015-06-01 12:00:00' AT TIME ZONE 'MSK';",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMPTZ), Column("timezone", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2010-06-01 09:00:00+00"), T("2015-06-01 09:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '2010-06-01 12:00:00 VOLT', timestamptz '2020-06-01 12:00:00 VOLT';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ), Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2010-06-01 08:00:00+00"), T("2020-06-01 08:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timestamptz '1980-01-01 00:00:00 SGT', timestamptz '2000-01-01 00:00:00 SGT';",
                    expected: Expected::Rows {
                        columns: &[Column("timestamptz", TIMESTAMPTZ), Column("timestamptz", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1979-12-31 16:30:00+00"), T("1999-12-31 16:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT timezone('IRKT', timestamptz '2012-01-01 00:00:00+00'), timezone('IRKT', timestamptz '2016-01-01 00:00:00+00');",
                    expected: Expected::Rows {
                        columns: &[Column("timezone", TIMESTAMP), Column("timezone", TIMESTAMP)],
                        rows: &[
                            &[T("2012-01-01 09:00:00"), T("2016-01-01 08:00:00")],
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
fn test_format_functions() {
    run_scripts(&[
        ScriptTest {
            name: "test to_char",
            set_up_script: &[
                "CREATE TABLE TIMESTAMP_TBL (d1 timestamp(2) without time zone);",
                "INSERT INTO TIMESTAMP_TBL VALUES ('1997-02-10 17:32:01-0800');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'YYYY-MM-DD HH24:MI:SS.MS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("2021-09-15 21:43:56.123")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'HH HH12 HH24 hh hh12 hh24 H h hH Hh');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("09 09 21 09 09 21 H h hH Hh")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'MI mi M m');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("43 43 M m")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'SS ss S s MS ms Ms mS US us Us uS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("56 56 S s 123 123 Ms mS 123457 123457 Us uS")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'Y,YYY y,yyy YYYY yyyy YYY yyy YY yy Y y');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("2,021 2,021 2021 2021 021 021 21 21 1 1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'MONTH Month month MON Mon mon MM mm Mm mM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("SEPTEMBER September september SEP Sep sep 09 09 Mm mM")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'DAY Day day DDD ddd DY Dy dy DD dd D d');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("WEDNESDAY Wednesday wednesday 258 258 WED Wed wed 15 15 4 4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'DAY Day day DDD ddd DY Dy dy DD dd D d');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("WEDNESDAY Wednesday wednesday 258 258 WED Wed wed 15 15 4 4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'IW iw');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("37 37")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'AM PM am pm A.M. P.M. a.m. p.m.');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("PM PM pm pm P.M. P.M. p.m. p.m.")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'Q q');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("3 3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char('2012-12-12 12:00'::timestamptz, 'YYYY-MM-DD SSSS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("2012-12-12 43200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone = '-06:30';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char('2012-12-12 12:00'::timestamptz, 'YYYY-MM-DD HH:MI:SS TZH:TZM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("2012-12-12 12:00:00 +06:30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char('2012-12-12 12:00 -02:00'::timestamptz, 'YYYY-MM-DD HH:MI:SS TZH:TZM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("2012-12-12 08:30:00 +06:30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char('2012-12-12 12:00 -02:00'::timestamptz, 'TZ');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone = 'UTC';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char('2012-12-12 12:00 -02:00'::timestamptz, 'TZ tz');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("UTC utc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(d1, 'Y,YYY YYYY YYY YY Y CC Q MM WW DDD DD D J') FROM TIMESTAMP_TBL;",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("1,997 1997 997 97 7 20 1 02 06 041 10 2 2450490")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(d1, 'DAY Day day DY Dy dy MONTH Month month RM MON Mon mon') FROM TIMESTAMP_TBL;",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("MONDAY    Monday    monday    MON Mon mon FEBRUARY  February  february  II   FEB Feb feb")],
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
fn test_functions_math() {
    run_scripts(&[
        ScriptTest {
            name: "cbrt",
            set_up_script: &[
                "CREATE TABLE test (pk INT primary key, v1 INT, v2 FLOAT4, v3 FLOAT8, v4 VARCHAR(255));",
                "INSERT INTO test VALUES (1, -1, -2, -3, '-5'), (2, 7, 11, 13, '17'), (3, 19, -23, 29, '-31');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT cbrt(v1), cbrt(v2), cbrt(v3) FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("cbrt", FLOAT8), Column("cbrt", FLOAT8), Column("cbrt", FLOAT8)],
                        rows: &[
                            &[T("-1"), T("-1.2599210498948732"), T("-1.4422495703074083")],
                            &[T("1.9129311827723892"), T("2.2239800905693157"), T("2.3513346877207577")],
                            &[T("2.668401648721945"), T("-2.8438669798515654"), T("3.072316825685847")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT round(cbrt(v1)::numeric, 10), round(cbrt(v2)::numeric, 10), round(cbrt(v3)::numeric, 10) FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("round", NUMERIC), Column("round", NUMERIC), Column("round", NUMERIC)],
                        rows: &[
                            &[T("-1.0000000000"), T("-1.2599210499"), T("-1.4422495703")],
                            &[T("1.9129311828"), T("2.2239800906"), T("2.3513346877")],
                            &[T("2.6684016487"), T("-2.8438669799"), T("3.0723168257")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cbrt(v4) FROM test ORDER BY pk;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function cbrt(character varying) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cbrt('64');",
                    expected: Expected::Rows {
                        columns: &[Column("cbrt", FLOAT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT round(cbrt('64'));",
                    expected: Expected::Rows {
                        columns: &[Column("round", FLOAT8)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT round('NaN'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("round", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 0::numeric::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("float8", FLOAT8)],
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
            name: "gcd",
            set_up_script: &[
                "CREATE TABLE test (pk INT primary key, v1 INT4, v2 INT8, v3 FLOAT8, v4 VARCHAR(255));",
                "INSERT INTO test VALUES (1, -2, -4, -6, '-8'), (2, 10, 12, 14.14, '16.16'), (3, 18, -20, 22.22, '-24.24');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT gcd(v1, 10), gcd(v2, 20) FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("gcd", INT4), Column("gcd", INT8)],
                        rows: &[
                            &[T("2"), T("4")],
                            &[T("10"), T("4")],
                            &[T("2"), T("20")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT gcd(v3, 10) FROM test ORDER BY pk;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function gcd(double precision, integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT gcd(v4, 10) FROM test ORDER BY pk;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function gcd(character varying, integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT gcd(36, '48');",
                    expected: Expected::Rows {
                        columns: &[Column("gcd", INT4)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT gcd('36', 48);",
                    expected: Expected::Rows {
                        columns: &[Column("gcd", INT4)],
                        rows: &[
                            &[T("12")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT gcd(1, 0), gcd(0, 1), gcd(0, 0);",
                    expected: Expected::Rows {
                        columns: &[Column("gcd", INT4), Column("gcd", INT4), Column("gcd", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "lcm",
            set_up_script: &[
                "CREATE TABLE test (pk INT primary key, v1 INT4, v2 INT8, v3 FLOAT8, v4 VARCHAR(255));",
                "INSERT INTO test VALUES (1, -2, -4, -6, '-8'), (2, 10, 12, 14.14, '16.16'), (3, 18, -20, 22.22, '-24.24');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT lcm(v1, 10), lcm(v2, 20) FROM test ORDER BY pk;",
                    expected: Expected::Rows {
                        columns: &[Column("lcm", INT4), Column("lcm", INT8)],
                        rows: &[
                            &[T("10"), T("20")],
                            &[T("10"), T("60")],
                            &[T("90"), T("20")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lcm(v3, 10) FROM test ORDER BY pk;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function lcm(double precision, integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lcm(v4, 10) FROM test ORDER BY pk;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function lcm(character varying, integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lcm(36, '48');",
                    expected: Expected::Rows {
                        columns: &[Column("lcm", INT4)],
                        rows: &[
                            &[T("144")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lcm('36', 48);",
                    expected: Expected::Rows {
                        columns: &[Column("lcm", INT4)],
                        rows: &[
                            &[T("144")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lcm(1, 0), lcm(0, 1), lcm(0, 0);",
                    expected: Expected::Rows {
                        columns: &[Column("lcm", INT4), Column("lcm", INT4), Column("lcm", INT4)],
                        rows: &[
                            &[T("0"), T("0"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "power",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT power(1::float8, 1::float8);",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(2::float8, 0.5::float8);",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("1.4142135623730951")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(0::float8, 0::float8);",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(4::float8, -1::float8);",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("0.25")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(-2::float8, -1::float8);",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("-0.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(0::float8, -1::float8);",
                    expected: Expected::Error(Diagnostic { code: "2201F", message: "zero raised to a negative power is undefined", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(1::numeric, 1::numeric)::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(2::numeric, 0.5::numeric)::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("1.414213562373095")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(0::numeric, 0::numeric)::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(4::numeric, -1::numeric)::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("0.25")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(-2::numeric, -1::numeric)::float8;",
                    expected: Expected::Rows {
                        columns: &[Column("power", FLOAT8)],
                        rows: &[
                            &[T("-0.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(0::numeric, -1::numeric);",
                    expected: Expected::Error(Diagnostic { code: "2201F", message: "zero raised to a negative power is undefined", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('3'::numeric, '-1'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("0.3333333333333333")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('3'::numeric, '-3'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("0.0370370370370370")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('Nan'::numeric, '-3'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('inf'::numeric, '-3'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('-inf'::numeric, '-3'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('-inf'::numeric, '3'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("-Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('-inf'::numeric, '4'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('0'::numeric, '3'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("0.0000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('-inf'::numeric, '-inf'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('-inf'::numeric, 'inf'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select power('-inf'::numeric, 'nan'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "greatest/least",
            set_up_script: &[
                "create table t(a decimal(6, 2), b decimal(8, 5), c decimal(5, 1));",
                "insert into t values (2.75, 8.8, 3.1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT GREATEST(25, 6, 7, 10, 20, 54);",
                    expected: Expected::Rows {
                        columns: &[Column("greatest", INT4)],
                        rows: &[
                            &[T("54")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT GREATEST(25, 6, 7, NULL, 20, 54);",
                    expected: Expected::Rows {
                        columns: &[Column("greatest", INT4)],
                        rows: &[
                            &[T("54")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT GREATEST(NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("greatest", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT LEAST(25, 6, 7, 10, 20, 54);",
                    expected: Expected::Rows {
                        columns: &[Column("least", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT LEAST(25, 6, 7, NULL, 20, 54);",
                    expected: Expected::Rows {
                        columns: &[Column("least", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT LEAST(NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("least", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select greatest(a, b, c), least(a, b, c) from t;",
                    expected: Expected::Rows {
                        columns: &[Column("greatest", NUMERIC), Column("least", NUMERIC)],
                        rows: &[
                            &[T("8.80000"), T("2.75")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "num_nonnulls and num_nulls",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 'num_nonnulls'::regproc::oid, 'num_nulls'::regproc::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("oid", OID)],
                        rows: &[
                            &[T("440"), T("438")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT num_nonnulls(1, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("num_nonnulls", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT num_nulls(1, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("num_nulls", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT num_nonnulls(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("num_nonnulls", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT num_nulls(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("num_nulls", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT num_nonnulls(1, 2, 3), num_nulls(1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("num_nonnulls", INT4), Column("num_nulls", INT4)],
                        rows: &[
                            &[T("3"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT num_nonnulls('a', NULL::int4, true, NULL::text, 1.5), num_nulls('a', NULL::int4, true, NULL::text, 1.5);",
                    expected: Expected::Rows {
                        columns: &[Column("num_nonnulls", INT4), Column("num_nulls", INT4)],
                        rows: &[
                            &[T("3"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT num_nonnulls();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function num_nonnulls() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT num_nulls();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function num_nulls() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE num_nulls_check (a uuid, b uuid, CONSTRAINT t_direction_check CHECK ((num_nonnulls(a, b) = 1)));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO num_nulls_check VALUES ('00000000-0000-0000-0000-000000000001', NULL);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO num_nulls_check VALUES ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002');",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "num_nulls_check" violates check constraint "t_direction_check""#, detail: "Failing row contains (00000000-0000-0000-0000-000000000001, 00000000-0000-0000-0000-000000000002).", schema: "public", table: "num_nulls_check", constraint: "t_direction_check", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO num_nulls_check VALUES (NULL, NULL);",
                    expected: Expected::Error(Diagnostic { code: "23514", message: r#"new row for relation "num_nulls_check" violates check constraint "t_direction_check""#, detail: "Failing row contains (null, null).", schema: "public", table: "num_nulls_check", constraint: "t_direction_check", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_functions_oid() {
    run_scripts(&[
        ScriptTest {
            name: "oid comparisons",
            set_up_script: &[
                "CREATE TABLE testing (pk INT primary key, v1 INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT 0::oid = 0, 0::oid <> 0, 0::oid < 1::oid, 845743985::oid = 845743985::oid, 845743985::oid = 845743986::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT conname FROM pg_catalog.pg_constraint WHERE conrelid = 'testing'::regclass AND conparentid = 0;",
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME)],
                        rows: &[
                            &[T("testing_pkey")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ('testing'::regclass::oid)::text::oid = 'testing'::regclass::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_regclass",
            set_up_script: &[
                "CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);",
                r#"CREATE TABLE "Testing2" (pk INT primary key, v1 INT);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_regclass('testing');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS)],
                        rows: &[
                            &[T("testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regclass('Testing2');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_regclass('"Testing2"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS)],
                        rows: &[
                            &[T(r#""Testing2""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regclass(('testing'::regclass)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS)],
                        rows: &[
                            &[T("testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regclass((('testing'::regclass)::oid)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regclass(('public.testing'::regclass)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS)],
                        rows: &[
                            &[T("testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = '';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regclass(('public.testing'::regclass)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_regclass", REGCLASS)],
                        rows: &[
                            &[T("public.testing")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_regproc",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_regproc('acos');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[T("acos")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_regproc('acos"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc(('acos'::regproc)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
                        rows: &[
                            &[T("acos")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regproc((('acos'::regproc)::oid)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_regproc", REGPROC)],
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
            name: "to_regtype",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_regtype('integer');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('integer[]');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("integer[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('int4');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('varchar');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("character varying")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('pg_catalog.varchar');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("character varying")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('varchar(10)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("character varying")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('char');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("character")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('pg_catalog.char');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T(r#""char""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('char(10)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("character")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_regtype('"char"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T(r#""char""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_regtype('pg_catalog."char"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T(r#""char""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('otherschema.char');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('timestamp');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("timestamp without time zone")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype('timestamp without time zone');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("timestamp without time zone")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_regtype('integer"');"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"unterminated quoted identifier at or near """"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype(('integer'::regtype)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype(('int'::regtype)::text);",
                    expected: Expected::Rows {
                        columns: &[Column("to_regtype", REGTYPE)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regtype((('integer'::regtype)::oid)::text);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "23""#, position: 1, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_regprocedure",
            set_up_script: &[
                "CREATE FUNCTION tf() RETURNS trigger AS $$ BEGIN RETURN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f2(a INT, b TEXT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3(INT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE FUNCTION f3(TEXT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE SCHEMA s;",
                "CREATE FUNCTION s.sf(INT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;",
                "CREATE PROCEDURE p1(INT) AS $$ BEGIN NULL; END; $$ LANGUAGE plpgsql;",
                "CREATE TABLE t1 (pk INT PRIMARY KEY);",
                "CREATE TRIGGER trg AFTER INSERT ON t1 FOR EACH ROW EXECUTE FUNCTION tf();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('tf()');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("tf()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('pg_catalog.now()');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("now()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_regprocedure('"tf"()');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("tf()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('public.tf()');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("tf()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure(' tf ( ) ');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("tf()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('f2(int, text)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("f2(integer,text)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('f2( integer , text ) ');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("f2(integer,text)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('f2(int4, varchar)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('f3(text)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("f3(text)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('f3(bool)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('s.sf(int)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("s.sf(integer)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('sf(int)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('p1(int)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("p1(integer)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('abs(float8)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("abs(double precision)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('nosuch()');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('nosuchschema.sf(int)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(to_regprocedure('tf()'));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("regprocedure")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('revision_change()') IS NULL;",
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
                    query: r#"SELECT 1 FROM pg_trigger t WHERE t.tgname = 'trg' AND t.tgfoid = to_regprocedure('"tf"()');"#,
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
                    query: "SELECT to_regprocedure('tf');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "expected a left parenthesis", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('f2(int,text');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "expected a right parenthesis", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('f2(int,)');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "expected a type name", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('f2(int))');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: "improper type name", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('abs(nosuchtype)');",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"type "nosuchtype" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = s;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('s.sf(int)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("sf(integer)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('public.f2(int,text)');",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", REGPROCEDURE)],
                        rows: &[
                            &[T("public.f2(integer,text)")],
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
fn test_json_functions() {
    run_scripts(&[
        ScriptTest {
            name: "json_build_array",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT json_build_array(1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("json_build_array", JSON)],
                        rows: &[
                            &[T("[1, 2, 3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_build_array(1, '2', 3);",
                    expected: Expected::Rows {
                        columns: &[Column("json_build_array", JSON)],
                        rows: &[
                            &[T(r#"[1, "2", 3]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_build_array();",
                    expected: Expected::Rows {
                        columns: &[Column("json_build_array", JSON)],
                        rows: &[
                            &[T("[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "json_build_object",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT json_build_object('a', 2, 'b', 4);",
                    expected: Expected::Rows {
                        columns: &[Column("json_build_object", JSON)],
                        rows: &[
                            &[T(r#"{"a" : 2, "b" : 4}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_build_object('a', 2, 'b');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "argument list must have even number of elements", hint: "The arguments of json_build_object() must consist of alternating keys and values.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT json_build_object(1, 2, 'b', 3);",
                    expected: Expected::Rows {
                        columns: &[Column("json_build_object", JSON)],
                        rows: &[
                            &[T(r#"{"1" : 2, "b" : 3}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb_build_array",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT jsonb_build_array(1, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_build_array", JSONB)],
                        rows: &[
                            &[T("[1, 2, 3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_build_array(1, '2', 3);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_build_array", JSONB)],
                        rows: &[
                            &[T(r#"[1, "2", 3]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_build_array();",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_build_array", JSONB)],
                        rows: &[
                            &[T("[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "jsonb_build_object",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT jsonb_build_object('a', 2, 'b', 4);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_build_object", JSONB)],
                        rows: &[
                            &[T(r#"{"a": 2, "b": 4}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_build_object('a', 2, 'b');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "argument list must have even number of elements", hint: "The arguments of jsonb_build_object() must consist of alternating keys and values.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jsonb_build_object(1, 2, 'b', 3);",
                    expected: Expected::Rows {
                        columns: &[Column("jsonb_build_object", JSONB)],
                        rows: &[
                            &[T(r#"{"1": 2, "b": 3}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_json",
            set_up_script: &[
                "SET TIME ZONE 'UTC'",
                "SET DateStyle = 'SQL, MDY'",
                "CREATE TABLE to_json_test (id int4, label text)",
                "INSERT INTO to_json_test VALUES (7, 'named')",
                "CREATE DOMAIN to_json_int_domain AS int4",
                "CREATE DOMAIN to_json_oid_domain AS oid",
                "CREATE DOMAIN to_json_json_domain AS json",
                "CREATE DOMAIN to_json_int_array_domain AS int4[]",
                "CREATE TYPE to_json_named_record AS (id int4, label text)",
                "CREATE DOMAIN to_json_named_record_domain AS to_json_named_record",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT to_json(E'quote " slash \\ newline\n'::text)"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""quote \" slash \\ newline\n""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(42::int4)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T("42")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(to_json(42)), pg_typeof(array_to_json(ARRAY[1])), pg_typeof(row_to_json(ROW(1)))",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("json"), T("json"), T("json")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(123.450::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T("123.450")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json('NaN'::numeric), to_json('Infinity'::numeric), to_json('-Infinity'::numeric)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON), Column("to_json", JSON), Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""NaN""#), T(r#""Infinity""#), T(r#""-Infinity""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(1.25::float8), to_json('NaN'::float8), to_json('Infinity'::float8), to_json('-Infinity'::float8)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON), Column("to_json", JSON), Column("to_json", JSON), Column("to_json", JSON)],
                        rows: &[
                            &[T("1.25"), T(r#""NaN""#), T(r#""Infinity""#), T(r#""-Infinity""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(1e20::float8), to_json(1e-7::float8)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON), Column("to_json", JSON)],
                        rows: &[
                            &[T("1e+20"), T("1e-07")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(true)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T("true")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(DATE '2024-02-29')",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""2024-02-29""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(TIMESTAMP '2024-02-29 12:34:56.123456')",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""2024-02-29T12:34:56.123456""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(TIMESTAMPTZ '2024-02-29 12:34:56.123456+05:30')",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""2024-02-29T07:04:56.123456+00:00""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(DATE '0001-01-01 BC'), to_json(TIMESTAMP '0001-01-01 12:34:56.123456 BC'), to_json(TIMESTAMPTZ '0001-01-01 12:34:56.123456+00 BC')",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON), Column("to_json", JSON), Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""0001-01-01 BC""#), T(r#""0001-01-01T12:34:56.123456 BC""#), T(r#""0001-01-01T12:34:56.123456+00:00 BC""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json('550e8400-e29b-41d4-a716-446655440000'::uuid)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""550e8400-e29b-41d4-a716-446655440000""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(23::oid), to_json('pg_class'::regclass), to_json('42'::xid)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON), Column("to_json", JSON), Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""23""#), T(r#""pg_class""#), T(r#""42""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(7::to_json_int_domain), to_json(23::to_json_oid_domain)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON), Column("to_json", JSON)],
                        rows: &[
                            &[T("7"), T(r#""23""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(ARRAY[1, NULL, 3]::to_json_int_array_domain)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T("[1,null,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json((ROW(7, 'named')::to_json_named_record)::to_json_named_record_domain)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#"{"id":7,"label":"named"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_json('{"b": 2, "a":1}'::json), to_json('{"b": 2, "a":1}'::jsonb)"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON), Column("to_json", JSON)],
                        rows: &[
                            &[T(r#"{"b": 2, "a":1}"#), T(r#"{"a": 1, "b": 2}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_json('{"b": 2, "a":1}'::to_json_json_domain)"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#"{"b": 2, "a":1}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_json(E'\\x00ff'::bytea), to_json(INTERVAL '1 day 02:03:04.5')"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON), Column("to_json", JSON)],
                        rows: &[
                            &[T(r#""\\x00ff""#), T(r#""1 day 02:03:04.5""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(ARRAY[1, NULL, 3]::int4[])",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T("[1,null,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(ROW(1, 'x'::text, NULL::bool))",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#"{"f1":1,"f2":"x","f3":null}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(t) FROM to_json_test t",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
                        rows: &[
                            &[T(r#"{"id":7,"label":"named"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_json(NULL::text)",
                    expected: Expected::Rows {
                        columns: &[Column("to_json", JSON)],
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
            name: "row_to_json anonymous row",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT row_to_json(row(1, 'foo'))",
                    expected: Expected::Rows {
                        columns: &[Column("row_to_json", JSON)],
                        rows: &[
                            &[T(r#"{"f1":1,"f2":"foo"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_to_json(row(1, true, null))",
                    expected: Expected::Rows {
                        columns: &[Column("row_to_json", JSON)],
                        rows: &[
                            &[T(r#"{"f1":1,"f2":true,"f3":null}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_to_json(row(1, 2, 3), false)",
                    expected: Expected::Rows {
                        columns: &[Column("row_to_json", JSON)],
                        rows: &[
                            &[T(r#"{"f1":1,"f2":2,"f3":3}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_to_json(row(1, 2, 3), true)",
                    expected: Expected::Rows {
                        columns: &[Column("row_to_json", JSON)],
                        rows: &[
                            &[T(r#"{"f1":1,
 "f2":2,
 "f3":3}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "row_to_json named columns",
            set_up_script: &[
                "CREATE TABLE rtj_test (id int4, name text)",
                "INSERT INTO rtj_test VALUES (1, 'Alice'), (2, 'Bob')",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT row_to_json(t) FROM rtj_test t ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("row_to_json", JSON)],
                        rows: &[
                            &[T(r#"{"id":1,"name":"Alice"}"#)],
                            &[T(r#"{"id":2,"name":"Bob"}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_to_json(t.*) FROM rtj_test t ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("row_to_json", JSON)],
                        rows: &[
                            &[T(r#"{"id":1,"name":"Alice"}"#)],
                            &[T(r#"{"id":2,"name":"Bob"}"#)],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_to_json(d.*) FROM (SELECT * FROM rtj_test) d ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("row_to_json", JSON)],
                        rows: &[
                            &[T(r#"{"id":1,"name":"Alice"}"#)],
                            &[T(r#"{"id":2,"name":"Bob"}"#)],
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
fn test_schema_visibility_inquiry_functions() {
    run_scripts(&[
        ScriptTest {
            name: "pg_function_is_visible",
            set_up_script: &[
                "CREATE SCHEMA myschema;",
                "SET search_path TO myschema;",
                "CREATE FUNCTION myfunc(a int) RETURNS int LANGUAGE sql AS 'SELECT a + 1';",
                "CREATE PROCEDURE myproc() LANGUAGE sql AS $$ SELECT 1 $$;",
                "CREATE SCHEMA testschema;",
                "SET search_path TO testschema;",
                "CREATE FUNCTION test_func(a int) RETURNS int LANGUAGE sql AS 'SELECT a + 2';",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(p.oid) FROM pg_catalog.pg_proc p WHERE p.proname = 'test_func';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(p.oid) FROM pg_catalog.pg_proc p WHERE p.proname = 'myfunc';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(p.oid) FROM pg_catalog.pg_proc p WHERE p.proname = 'myproc';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'myschema';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(p.oid) FROM pg_catalog.pg_proc p WHERE p.proname = 'myfunc';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(p.oid) FROM pg_catalog.pg_proc p WHERE p.proname = 'myproc';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(p.oid) FROM pg_catalog.pg_proc p WHERE p.proname = 'test_func';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(31);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(845743985);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_function_is_visible(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_function_is_visible", BOOL)],
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
            name: "pg_table_is_visible",
            set_up_script: &[
                "CREATE SCHEMA myschema;",
                "SET search_path TO myschema;",
                "CREATE TABLE mytable (id int, name text);",
                "INSERT INTO mytable VALUES (1,'desk'), (2,'chair');",
                "CREATE VIEW myview AS SELECT name FROM mytable;",
                "CREATE SCHEMA testschema;",
                "SET search_path TO testschema;",
                "CREATE TABLE test_table (pk INT primary key, v1 INT UNIQUE);",
                "INSERT INTO test_table VALUES (1,5), (2,7);",
                "CREATE INDEX test_index ON test_table(v1);",
                "CREATE SEQUENCE test_seq START 39;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT c.oid, c.relname AS table_name, n.nspname AS table_schema FROM pg_catalog.pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE (n.nspname='myschema' OR n.nspname='testschema') AND left(relname, 5) <> 'dolt_' order by relname;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("table_name", NAME), Column("table_schema", NAME)],
                        rows: &[
                            &[Oid(16385), T("mytable"), T("myschema")],
                            &[Oid(16390), T("myview"), T("myschema")],
                            &[Oid(16402), T("test_index"), T("testschema")],
                            &[Oid(16403), T("test_seq"), T("testschema")],
                            &[Oid(16395), T("test_table"), T("testschema")],
                            &[Oid(16398), T("test_table_pkey"), T("testschema")],
                            &[Oid(16400), T("test_table_v1_key"), T("testschema")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("testschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select pg_table_is_visible(3057657334);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_table_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select pg_table_is_visible(1952237395);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_table_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select pg_table_is_visible(1539973141);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_table_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select pg_table_is_visible(3983475213);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_table_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'myschema';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("myschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select pg_table_is_visible(3983475213);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_table_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select pg_table_is_visible(3905781870);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_table_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_type_is_visible",
            set_up_script: &[
                "CREATE SCHEMA myschema;",
                "SET search_path TO myschema;",
                "CREATE DOMAIN mydomain AS text;",
                "CREATE TYPE myenum AS ENUM ('a', 'b', 'c');",
                "CREATE SCHEMA testschema;",
                "SET search_path TO testschema;",
                "CREATE DOMAIN test_domain AS int;",
                "CREATE TYPE test_enum AS ENUM ('x', 'y', 'z');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT t.oid, t.typname, n.nspname FROM pg_catalog.pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE n.nspname='myschema' OR n.nspname='testschema' ORDER BY t.typname;",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("typname", NAME), Column("nspname", NAME)],
                        rows: &[
                            &[Oid(16385), T("_mydomain"), T("myschema")],
                            &[Oid(16387), T("_myenum"), T("myschema")],
                            &[Oid(16396), T("_test_domain"), T("testschema")],
                            &[Oid(16398), T("_test_enum"), T("testschema")],
                            &[Oid(16386), T("mydomain"), T("myschema")],
                            &[Oid(16388), T("myenum"), T("myschema")],
                            &[Oid(16397), T("test_domain"), T("testschema")],
                            &[Oid(16399), T("test_enum"), T("testschema")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("testschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(2272253470);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(1117094145);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(340132571);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(1684884017);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'myschema';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path;",
                    expected: Expected::Rows {
                        columns: &[Column("search_path", TEXT)],
                        rows: &[
                            &[T("myschema")],
                        ],
                        tag: "SHOW",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(340132571);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(1684884017);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(2272253470);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(1117094145);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("The OID is one Doltgres gives an object, which names nothing in Postgres"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible(999999);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path = 'pg_catalog';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible('text'::regtype::oid);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_type_is_visible('int4'::regtype::oid);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_type_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_collation_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_collation_is_visible(950);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_collation_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_collation_is_visible(100);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_collation_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_collation_is_visible(397);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_collation_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_collation_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_collation_is_visible", BOOL)],
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
            name: "pg_opclass_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_opclass_is_visible(15000);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_opclass_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_opclass_is_visible(397);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_opclass_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_opclass_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_opclass_is_visible", BOOL)],
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
            name: "pg_opfamily_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_opfamily_is_visible(397);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_opfamily_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_opfamily_is_visible(15000);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_opfamily_is_visible", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_opfamily_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_opfamily_is_visible", BOOL)],
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
            name: "pg_operator_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_operator_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_operator_is_visible", BOOL)],
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
            name: "pg_conversion_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_conversion_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_conversion_is_visible", BOOL)],
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
            name: "pg_ts_config_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_ts_config_is_visible(3748);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_ts_config_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_ts_config_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_ts_config_is_visible", BOOL)],
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
            name: "pg_ts_dict_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_ts_dict_is_visible(3765);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_ts_dict_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_ts_dict_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_ts_dict_is_visible", BOOL)],
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
            name: "pg_ts_template_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_ts_template_is_visible(3727);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_ts_template_is_visible", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_ts_template_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_ts_template_is_visible", BOOL)],
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
            name: "pg_statistics_obj_is_visible",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_statistics_obj_is_visible(22);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_statistics_obj_is_visible", BOOL)],
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
fn test_select_from_functions() {
    run_scripts(&[
        ScriptTest {
            name: "select * FROM functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT array_to_string(ARRAY[1, 2, 3, NULL, 5], ',', '*')",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("1,2,3,*,5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM array_to_string(ARRAY[1, 2, 3, NULL, 5], ',', '*')",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("1,2,3,*,5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM array_to_string(ARRAY[37.89, 1.2], '_');",
                    expected: Expected::Rows {
                        columns: &[Column("array_to_string", TEXT)],
                        rows: &[
                            &[T("37.89_1.2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM format_type('text'::regtype, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("text(4)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from format_type(874938247, 20);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("???")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM to_char(timestamp '2021-09-15 21:43:56.123456789', 'IW iw');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("37 37")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from format_type('text'::regtype, -1);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("text")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT "left" FROM left('name'::name, 2);"#,
                    expected: Expected::Rows {
                        columns: &[Column("left", TEXT)],
                        rows: &[
                            &[T("na")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length FROM length('name'::name);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lower FROM lower('naMe'::name);",
                    expected: Expected::Rows {
                        columns: &[Column("lower", TEXT)],
                        rows: &[
                            &[T("name")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM lpad('name'::name, 7, '*');",
                    expected: Expected::Rows {
                        columns: &[Column("lpad", TEXT)],
                        rows: &[
                            &[T("***name")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "test select  from dolt_ functions",
            set_up_script: &[
                "CREATE TABLE test (pk INT primary key, v1 INT, v2 TEXT);",
                "INSERT INTO test VALUES (1, 1, 'a'), (2, 2, 'b'), (3, 3, 'c'), (4, 4, 'd'), (5, 5, 'e');",
                "call dolt_commit('-Am', 'first table');",
            ],
            skip: Some(r#"setup fails on Postgres ("error running setup query: call dolt_commit('-Am', 'first table');: ERROR: procedure dolt_commit(unknown, unknown) does not exist (SQLSTATE 42883)") and on the Go server ("error running setup query: call dolt_commit('-Am', 'first table');: ERROR: Dolt stored procedure may only be invoked using SELECT (SQLSTATE XX000)")"#),
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from dolt_branch('newBranch')",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select status from dolt_checkout('newBranch')",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into test values (6, 6, 'f')",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select length(commit_hash) > 0 from (select commit_hash from dolt_commit('-Am', 'added f') as result)",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select dolt_checkout('main')",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select fast_forward, conflicts from dolt_merge('newBranch')",
                    flow: Flow::Query,
                    skip: Some("no observation"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_set_returning_functions() {
    run_scripts(&[
        ScriptTest {
            name: "generate_series",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT generate_series(1,3)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
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
                    query: "SELECT generate_series(1,6,2)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series(1::int4,6::int4,0::int4)",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "step size cannot equal zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series(6,1,-2)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
                        rows: &[
                            &[T("6")],
                            &[T("4")],
                            &[T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series(1.5,6,2)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", NUMERIC)],
                        rows: &[
                            &[T("1.5")],
                            &[T("3.5")],
                            &[T("5.5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series(1::int8,6::int8,0::int8)",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "step size cannot equal zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series(6,2.2,-2)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", NUMERIC)],
                        rows: &[
                            &[T("6")],
                            &[T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('2008-03-01 00:00'::timestamp,'2008-03-02 12:00', '10 hours');",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", TIMESTAMP)],
                        rows: &[
                            &[T("2008-03-01 00:00:00")],
                            &[T("2008-03-01 10:00:00")],
                            &[T("2008-03-01 20:00:00")],
                            &[T("2008-03-02 06:00:00")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('2008-03-02 12:00'::timestamp,'2008-03-01 00:00'::timestamp, '-10 hours');",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", TIMESTAMP)],
                        rows: &[
                            &[T("2008-03-02 12:00:00")],
                            &[T("2008-03-02 02:00:00")],
                            &[T("2008-03-01 16:00:00")],
                            &[T("2008-03-01 06:00:00")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_series as table function",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series(1,3)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
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
                    query: "SELECT * FROM generate_series(1,6,2)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                            &[T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series(-100::numeric, 100::numeric, 0::numeric);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "step size cannot equal zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series('2008-03-02 12:00'::timestamp,'2008-03-01 00:00'::timestamp, '-10 hours'::interval);",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", TIMESTAMP)],
                        rows: &[
                            &[T("2008-03-02 12:00:00")],
                            &[T("2008-03-02 02:00:00")],
                            &[T("2008-03-01 16:00:00")],
                            &[T("2008-03-01 06:00:00")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from generate_series('2020-01-01 00:00'::timestamp, '2020-01-02 03:00'::timestamp, '0 hour'::interval);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "step size cannot equal zero", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series('2008-03-02 12:00'::timestamp,'2008-03-01 00:00'::timestamp, '-10 hours');",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", TIMESTAMP)],
                        rows: &[
                            &[T("2008-03-02 12:00:00")],
                            &[T("2008-03-02 02:00:00")],
                            &[T("2008-03-01 16:00:00")],
                            &[T("2008-03-01 06:00:00")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('1.2'::numeric,2.4)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", NUMERIC)],
                        rows: &[
                            &[T("1.2")],
                            &[T("2.2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('1.2'::numeric,1.4,0.1)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", NUMERIC)],
                        rows: &[
                            &[T("1.2")],
                            &[T("1.3")],
                            &[T("1.4")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('Nan'::numeric,1.4,0.1)",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "start value cannot be NaN", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('NaN'::numeric,1.4)",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "start value cannot be NaN", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('1.2'::numeric,'Infinity',0.1)",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "stop value cannot be infinity", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('1.2'::numeric,'-Infinity')",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "stop value cannot be infinity", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series('1.2'::numeric,1.4,'NAN')",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "step size cannot be NaN", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "aliased SRF alongside scalar columns with ORDER BY",
            set_up_script: &[
                "CREATE TABLE srf_sort (id integer PRIMARY KEY, arr integer[]);",
                "INSERT INTO srf_sort VALUES (7, '{101,202,303}'), (8, '{44}');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id AS source_id, unnest(arr) AS elem FROM srf_sort WHERE id = 7 ORDER BY elem DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("source_id", INT4), Column("elem", INT4)],
                        rows: &[
                            &[T("7"), T("303")],
                            &[T("7"), T("202")],
                            &[T("7"), T("101")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id AS source_id, unnest(arr) AS elem FROM srf_sort WHERE id = 7 ORDER BY elem DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("source_id", INT4), Column("elem", INT4)],
                        rows: &[
                            &[T("7"), T("303")],
                            &[T("7"), T("202")],
                            &[T("7"), T("101")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id AS source_id, unnest(arr) AS elem FROM srf_sort ORDER BY elem DESC;",
                    expected: Expected::Rows {
                        columns: &[Column("source_id", INT4), Column("elem", INT4)],
                        rows: &[
                            &[T("7"), T("303")],
                            &[T("7"), T("202")],
                            &[T("7"), T("101")],
                            &[T("8"), T("44")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id AS source_id, generate_series(1, 2) AS n FROM srf_sort ORDER BY n, source_id;",
                    expected: Expected::Rows {
                        columns: &[Column("source_id", INT4), Column("n", INT4)],
                        rows: &[
                            &[T("7"), T("1")],
                            &[T("8"), T("1")],
                            &[T("7"), T("2")],
                            &[T("8"), T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id AS source_id, unnest(arr) AS elem FROM srf_sort WHERE id = 7;",
                    expected: Expected::Rows {
                        columns: &[Column("source_id", INT4), Column("elem", INT4)],
                        rows: &[
                            &[T("7"), T("101")],
                            &[T("7"), T("202")],
                            &[T("7"), T("303")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_series as table function with column alias",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series(1,3) AS s(r)",
                    expected: Expected::Rows {
                        columns: &[Column("r", INT4)],
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
                    query: "SELECT r FROM generate_series(1,3) AS s(r)",
                    expected: Expected::Rows {
                        columns: &[Column("r", INT4)],
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
                    query: "SELECT r + 1 FROM generate_series(1,3) AS s(r) WHERE r > 1",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("3")],
                            &[T("4")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series(1, array_upper(current_schemas(false), 1)) AS s(r)",
                    expected: Expected::Rows {
                        columns: &[Column("r", INT4)],
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
            name: "table function WITH ORDINALITY",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series(2,4) WITH ORDINALITY",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4), Column("ordinality", INT8)],
                        rows: &[
                            &[T("2"), T("1")],
                            &[T("3"), T("2")],
                            &[T("4"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM generate_series(2,4) WITH ORDINALITY AS a(n, ord)",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("ord", INT8)],
                        rows: &[
                            &[T("2"), T("1")],
                            &[T("3"), T("2")],
                            &[T("4"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ord, n FROM generate_series(2,4) WITH ORDINALITY AS a(n, ord) ORDER BY ord DESC",
                    expected: Expected::Rows {
                        columns: &[Column("ord", INT8), Column("n", INT4)],
                        rows: &[
                            &[T("3"), T("4")],
                            &[T("2"), T("3")],
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT n FROM generate_series(5,7) WITH ORDINALITY AS a(n, ord) WHERE ord = 2",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_partition_ancestors",
            set_up_script: &[
                "CREATE TABLE anc_test (pk INT PRIMARY KEY, v1 INT);",
                "CREATE VIEW anc_view AS SELECT pk FROM anc_test;",
                "CREATE TABLE anc_child (pk INT PRIMARY KEY, apk INT REFERENCES anc_test(pk));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_partition_ancestors('anc_test'::regclass);",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_partition_ancestors('anc_test'::regclass) WITH ORDINALITY AS a(relid, depth);",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS), Column("depth", INT8)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_partition_ancestors('anc_view'::regclass);",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM pg_partition_ancestors(845743985);",
                    expected: Expected::Rows {
                        columns: &[Column("relid", REGCLASS)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT conrelid = 'anc_child'::pg_catalog.regclass AS sametable,
       conname, pg_catalog.pg_get_constraintdef(oid, true) AS condef, conrelid::pg_catalog.regclass::text AS ontable
FROM pg_catalog.pg_constraint, pg_catalog.pg_partition_ancestors('anc_child'::regclass)
WHERE conrelid = relid AND contype = 'f' AND conparentid = 0
ORDER BY sametable DESC, conname;"#,
                    expected: Expected::Rows {
                        columns: &[Column("sametable", BOOL), Column("conname", NAME), Column("condef", TEXT), Column("ontable", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT conname, conrelid::pg_catalog.regclass::text AS ontable,
       pg_catalog.pg_get_constraintdef(oid, true) AS condef
FROM pg_catalog.pg_constraint c
WHERE confrelid IN (SELECT pg_catalog.pg_partition_ancestors('anc_test'::regclass)
                    UNION ALL VALUES ('anc_test'::pg_catalog.regclass))
      AND contype = 'f' AND conparentid = 0
ORDER BY conname;"#,
                    expected: Expected::Rows {
                        columns: &[Column("conname", NAME), Column("ontable", TEXT), Column("condef", TEXT)],
                        rows: &[
                            &[T("anc_child_apk_fkey"), T("anc_child"), T("FOREIGN KEY (apk) REFERENCES anc_test(pk)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set-returning function as join operand",
            set_up_script: &[
                "CREATE TABLE test1 (id INT PRIMARY KEY);",
                "INSERT INTO test1 VALUES (1), (2), (4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, r FROM test1 LEFT JOIN generate_series(1,3) s(r) ON id = r ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("r", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, r FROM test1 JOIN generate_series(1,3) AS s(r) ON id = s.r ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("r", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, r FROM generate_series(1,3) s(r) LEFT JOIN test1 ON id = r ORDER BY r;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("r", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[Null, T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_series as table function used in pgJDBC-style enum catalog query",
            set_up_script: &[
                "CREATE TYPE status_enum AS ENUM ('one', 'two', 'three');",
                "CREATE TABLE test1 (id INT, status status_enum);",
                "INSERT INTO test1 VALUES (1, 'one'), (2, 'two'), (3, 'three');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT
    typinput = 'pg_catalog.array_in'::regproc AS is_array,
    typtype,
    typname
FROM pg_catalog.pg_type
LEFT JOIN (
    SELECT ns.oid AS nspoid, ns.nspname, r.r
    FROM pg_namespace AS ns
    JOIN (
        SELECT
            s.r,
            (current_schemas(false))[s.r] AS nspname
        FROM generate_series(
            1,
            array_upper(current_schemas(false), 1)
        ) AS s(r)
    ) AS r USING (nspname)
) AS sp ON sp.nspoid = typnamespace
WHERE pg_type.oid = (
    SELECT atttypid
    FROM pg_attribute
    WHERE attrelid = 'test1'::regclass
      AND attname = 'status'
)
ORDER BY sp.r, pg_type.oid DESC;"#,
                    expected: Expected::Rows {
                        columns: &[Column("is_array", BOOL), Column("typtype", CHAR), Column("typname", NAME)],
                        rows: &[
                            &[T("f"), T("e"), T("status_enum")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "nested generate_series",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT generate_series(1, generate_series(1, 3))",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("1")],
                            &[T("2")],
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "limit, offset, sort",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT a, generate_series(1,2) FROM (VALUES(1),(2),(3)) r(a) LIMIT 2 OFFSET 2;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("generate_series", INT4)],
                        rows: &[
                            &[T("2"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, generate_series(1,2) FROM (VALUES(1),(2),(3)) r(a) ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("generate_series", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("1")],
                            &[T("3"), T("2")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_series with table",
            set_up_script: &[
                "CREATE TABLE t1 (pk INT primary key, v1 INT);",
                "INSERT INTO t1 VALUES (1, 1), (2, 2), (3, 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT generate_series(1,3), pk from t1",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4), Column("pk", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("1")],
                            &[T("3"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("2")],
                            &[T("3"), T("2")],
                            &[T("1"), T("3")],
                            &[T("2"), T("3")],
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_series(1,3) + pk, pk from t1",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("pk", INT4)],
                        rows: &[
                            &[T("2"), T("1")],
                            &[T("3"), T("1")],
                            &[T("4"), T("1")],
                            &[T("3"), T("2")],
                            &[T("4"), T("2")],
                            &[T("5"), T("2")],
                            &[T("4"), T("3")],
                            &[T("5"), T("3")],
                            &[T("6"), T("3")],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set returning function as table function: generate_series",
            assertions: &[
                ScriptTestAssertion {
                    query: "select * from generate_series(1,3)",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4)],
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
                    query: "select sum(null::int4) from generate_series(1,3);",
                    expected: Expected::Rows {
                        columns: &[Column("sum", INT8)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * from generate_series('2008-03-01 00:00'::timestamp,'2008-03-02 12:00', '10 hours');",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", TIMESTAMP)],
                        rows: &[
                            &[T("2008-03-01 00:00:00")],
                            &[T("2008-03-01 10:00:00")],
                            &[T("2008-03-01 20:00:00")],
                            &[T("2008-03-02 06:00:00")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_subscripts",
            set_up_script: &[
                "CREATE TABLE t1 (pk INT primary key, v1 INT[]);",
                "INSERT INTO t1 VALUES (1, ARRAY[1, 2, 3]), (2, ARRAY[4, 5]), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select generate_subscripts(v1, 1) from t1 where pk = 1",
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
                    query: "select generate_subscripts(v1, 1) + 100 from t1 where pk = 1",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("101")],
                            &[T("102")],
                            &[T("103")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select generate_subscripts(v1, 1) from t1 where pk = 3",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select generate_subscripts(v1, 1), v1 from t1",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4), Column("v1", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{1,2,3}")],
                            &[T("2"), T("{1,2,3}")],
                            &[T("3"), T("{1,2,3}")],
                            &[T("1"), T("{4,5}")],
                            &[T("2"), T("{4,5}")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_subscripts as table function with scalar alias",
            set_up_script: &[
                "CREATE TABLE array_alias_test (id INT primary key, values_array INT[]);",
                "INSERT INTO array_alias_test VALUES (1, ARRAY[10, 20, 30]), (2, NULL), (3, ARRAY[]::INT[]);",
                r#"CREATE OR REPLACE FUNCTION calculate_bonus(
    IN current_salary NUMERIC,
    OUT bonus_amount NUMERIC,
    OUT new_total_salary NUMERIC
) AS $$
BEGIN
    bonus_amount := current_salary * 0.10;
    new_total_salary := current_salary + bonus_amount;
END;
$$ LANGUAGE plpgsql;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT k FROM generate_subscripts(ARRAY[1, 2, 3], 1) AS k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4)],
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
                    query: "SELECT ARRAY(SELECT k + 1 FROM generate_subscripts(ARRAY[10, 20, 30], 1) AS k ORDER BY k);",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{2,3,4}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT values_array[k] + 1 FROM generate_subscripts(values_array, 1) AS k ORDER BY k) FROM array_alias_test WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("{11,21,31}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k FROM generate_subscripts(NULL::INT[], 1) AS k ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k FROM generate_subscripts(ARRAY[]::INT[], 1) AS k ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k FROM generate_subscripts(ARRAY[10, 20]::INT[], NULL::INT) AS k ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_subscripts(NULL::INT[], 1);",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT idx FROM generate_subscripts(NULL::INT[], 1) AS k(idx) ORDER BY idx;",
                    expected: Expected::Rows {
                        columns: &[Column("idx", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, ARRAY(SELECT k FROM generate_subscripts(values_array, 1) AS k ORDER BY k) FROM array_alias_test ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("array", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{1,2,3}")],
                            &[T("2"), T("{}")],
                            &[T("3"), T("{}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.bonus_amount, c.new_total_salary, k FROM calculate_bonus(5000) AS c CROSS JOIN generate_subscripts(ARRAY[10,20], 1) AS k ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("bonus_amount", NUMERIC), Column("new_total_salary", NUMERIC), Column("k", INT4)],
                        rows: &[
                            &[T("500.00"), T("5500.00"), T("1")],
                            &[T("500.00"), T("5500.00"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.bonus_amount, c.new_total_salary FROM calculate_bonus(5000) AS c;",
                    expected: Expected::Rows {
                        columns: &[Column("bonus_amount", NUMERIC), Column("new_total_salary", NUMERIC)],
                        rows: &[
                            &[T("500.00"), T("5500.00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k, c.bonus_amount, c.new_total_salary FROM generate_subscripts(ARRAY[10,20], 1) AS k CROSS JOIN calculate_bonus(5000) AS c ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("bonus_amount", NUMERIC), Column("new_total_salary", NUMERIC)],
                        rows: &[
                            &[T("1"), T("500.00"), T("5500.00")],
                            &[T("2"), T("500.00"), T("5500.00")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k, j FROM generate_subscripts(ARRAY[10,20], 1) AS k CROSS JOIN generate_subscripts(ARRAY[30,40], 1) AS j ORDER BY k, j;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("j", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("1"), T("2")],
                            &[T("2"), T("1")],
                            &[T("2"), T("2")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.bonus, c.total, k.idx FROM calculate_bonus(5000) AS c(bonus,total) CROSS JOIN generate_subscripts(ARRAY[10,20], 1) AS k(idx) ORDER BY k.idx;",
                    expected: Expected::Rows {
                        columns: &[Column("bonus", NUMERIC), Column("total", NUMERIC), Column("idx", INT4)],
                        rows: &[
                            &[T("500.00"), T("5500.00"), T("1")],
                            &[T("500.00"), T("5500.00"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_subscripts with join",
            set_up_script: &[
                "CREATE TABLE t1 (a INT[]);",
                "CREATE TABLE t2 (b int[]);",
                "INSERT INTO t1 VALUES (ARRAY[1]), (ARRAY[1, 2, 3])",
                "INSERT INTO t2 VALUES (ARRAY[9,10])",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select generate_subscripts(a, 1), a, generate_subscripts(b, 1), b from t1, t2;",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4), Column("a", INT4_ARRAY), Column("generate_subscripts", INT4), Column("b", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{1}"), T("1"), T("{9,10}")],
                            &[Null, T("{1}"), T("2"), T("{9,10}")],
                            &[T("1"), T("{1,2,3}"), T("1"), T("{9,10}")],
                            &[T("2"), T("{1,2,3}"), T("2"), T("{9,10}")],
                            &[T("3"), T("{1,2,3}"), Null, T("{9,10}")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_subscripts and generate_series combined",
            set_up_script: &[
                "CREATE TABLE t1 (a INT[]);",
                "INSERT INTO t1 VALUES (ARRAY[1, 2, 3]), (ARRAY[4, 5]);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "select generate_subscripts(a, 1), a, generate_series(1,4) from t1;",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4), Column("a", INT4_ARRAY), Column("generate_series", INT4)],
                        rows: &[
                            &[T("1"), T("{1,2,3}"), T("1")],
                            &[T("2"), T("{1,2,3}"), T("2")],
                            &[T("3"), T("{1,2,3}"), T("3")],
                            &[Null, T("{1,2,3}"), T("4")],
                            &[T("1"), T("{4,5}"), T("1")],
                            &[T("2"), T("{4,5}"), T("2")],
                            &[Null, T("{4,5}"), T("3")],
                            &[Null, T("{4,5}"), T("4")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_subscripts on 0-indexed array types",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT generate_subscripts('1 2 3'::int2vector, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4)],
                        rows: &[
                            &[T("0")],
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT generate_subscripts('1 2 3'::oidvector, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("generate_subscripts", INT4)],
                        rows: &[
                            &[T("0")],
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "set generation with other func calls",
            set_up_script: &[
                "CREATE sequence test_seq START WITH 1 INCREMENT BY 3;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT generate_series(1, 5), nextval('test_seq')",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4), Column("nextval", INT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("4")],
                            &[T("3"), T("7")],
                            &[T("4"), T("10")],
                            &[T("5"), T("13")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generate_series as table function and projection",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT *, unnest(ARRAY['cat', 'dog', 'bird']) AS animal FROM generate_series(1, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("generate_series", INT4), Column("animal", TEXT)],
                        rows: &[
                            &[T("1"), T("cat")],
                            &[T("1"), T("dog")],
                            &[T("1"), T("bird")],
                            &[T("2"), T("cat")],
                            &[T("2"), T("dog")],
                            &[T("2"), T("bird")],
                            &[T("3"), T("cat")],
                            &[T("3"), T("dog")],
                            &[T("3"), T("bird")],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "insert with set returning function",
            set_up_script: &[
                "create table hash_parted (a int, b int, c int);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "insert into hash_parted values(0, generate_series(1,3), generate_series(5,8));",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "insert into hash_parted values(0, generate_series(11,12), generate_series(51,54)), (1, generate_series(1,3), generate_series(5,8));",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "set-returning functions are not allowed in VALUES", position: 35, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from hash_parted;",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("0"), T("1"), T("5")],
                            &[T("0"), T("2"), T("6")],
                            &[T("0"), T("3"), T("7")],
                            &[T("0"), Null, T("8")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_string_function() {
    run_scripts(&[
        ScriptTest {
            name: "use name type for text type input",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ascii('name'::name)",
                    expected: Expected::Rows {
                        columns: &[Column("ascii", INT4)],
                        rows: &[
                            &[T("110")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT bit_length('name'::name);",
                    expected: Expected::Rows {
                        columns: &[Column("bit_length", INT4)],
                        rows: &[
                            &[T("32")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT btrim(' name  '::name);",
                    expected: Expected::Rows {
                        columns: &[Column("btrim", TEXT)],
                        rows: &[
                            &[T("name")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT initcap('name'::name);",
                    expected: Expected::Rows {
                        columns: &[Column("initcap", TEXT)],
                        rows: &[
                            &[T("Name")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT left('name'::name, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("left", TEXT)],
                        rows: &[
                            &[T("na")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length('name'::name);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lower('naMe'::name);",
                    expected: Expected::Rows {
                        columns: &[Column("lower", TEXT)],
                        rows: &[
                            &[T("name")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT lpad('name'::name, 7, '*');",
                    expected: Expected::Rows {
                        columns: &[Column("lpad", TEXT)],
                        rows: &[
                            &[T("***name")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "quote_ident",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"select quote_ident('hi"bye');"#,
                    expected: Expected::Rows {
                        columns: &[Column("quote_ident", TEXT)],
                        rows: &[
                            &[T(r#""hi""bye""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select quote_ident('hi""bye');"#,
                    expected: Expected::Rows {
                        columns: &[Column("quote_ident", TEXT)],
                        rows: &[
                            &[T(r#""hi""""bye""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select quote_ident('hi"""bye');"#,
                    expected: Expected::Rows {
                        columns: &[Column("quote_ident", TEXT)],
                        rows: &[
                            &[T(r#""hi""""""bye""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select quote_ident('hi"b"ye');"#,
                    expected: Expected::Rows {
                        columns: &[Column("quote_ident", TEXT)],
                        rows: &[
                            &[T(r#""hi""b""ye""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "translate",
            assertions: &[
                ScriptTestAssertion {
                    query: "select translate('12345', '143', 'ax');",
                    expected: Expected::Rows {
                        columns: &[Column("translate", TEXT)],
                        rows: &[
                            &[T("a2x5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select translate('12345', '143', 'axs');",
                    expected: Expected::Rows {
                        columns: &[Column("translate", TEXT)],
                        rows: &[
                            &[T("a2sx5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select translate('12345', '143', 'axsl');",
                    expected: Expected::Rows {
                        columns: &[Column("translate", TEXT)],
                        rows: &[
                            &[T("a2sx5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select translate('こんにちは', 'ん', 'a');",
                    expected: Expected::Rows {
                        columns: &[Column("translate", TEXT)],
                        rows: &[
                            &[T("こaにちは")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "substring with integer arg",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT substr('hello', 2)",
                    expected: Expected::Rows {
                        columns: &[Column("substr", TEXT)],
                        rows: &[
                            &[T("ello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring('hello', 2)",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("ello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "substring with integer args",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT substr('hello', 2, 3)",
                    expected: Expected::Rows {
                        columns: &[Column("substr", TEXT)],
                        rows: &[
                            &[T("ell")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring('hello', 2, 3)",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("ell")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "substring with integer args, expanded form",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT substr('hello' from 2 for 3)",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 23, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring('hello' from 2 for 3)",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("ell")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substr('hello' from 2)",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "from""#, position: 23, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring('hello' from 2)",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("ello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substr('hello' for 3)",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "for""#, position: 23, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring('hello' for 3)",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("hel")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "substring with regex",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT substring('hello', 'l+')",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("ll")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring('hello' FROM 'l+')",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("ll")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT substring('hello.' similar 'hello#.' escape '#')",
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("hello.")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT substring('Thomas' similar '%#"o_a#"_' escape '#')"#,
                    expected: Expected::Rows {
                        columns: &[Column("substring", TEXT)],
                        rows: &[
                            &[T("oma")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "string_agg",
            set_up_script: &[
                "CREATE TABLE test (pk INT primary key, v1 INT, v2 TEXT);",
                "INSERT INTO test VALUES (1, 1, 'a'), (2, 2, 'b'), (3, 3, 'c'), (4, 4, 'd'), (5, 5, 'e');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT string_agg(v1::text, ',') FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("string_agg", TEXT)],
                        rows: &[
                            &[T("1,2,3,4,5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_agg(v2, '|') FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("string_agg", TEXT)],
                        rows: &[
                            &[T("a|b|c|d|e")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT STRING_AGG(concat(v1::text, v2), ' * ') FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("string_agg", TEXT)],
                        rows: &[
                            &[T("1a * 2b * 3c * 4d * 5e")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT STRING_agg(concat(v1::text, v2), CONCAT(' *', ' ') ORDER BY V1 DESC) FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("string_agg", TEXT)],
                        rows: &[
                            &[T("5e * 4d * 3c * 2b * 1a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT STRING_AGG(v2, '*', v1) FROM test;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function string_agg(text, unknown, integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT STRING_AGG(v2) FROM test;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function string_agg(text) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT STRING_AGG(concat(v1::text, v2), ' * '::text) FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("string_agg", TEXT)],
                        rows: &[
                            &[T("1a * 2b * 3c * 4d * 5e")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT STRING_AGG(concat(v1::text, v2), 8::text) FROM test;",
                    expected: Expected::Rows {
                        columns: &[Column("string_agg", TEXT)],
                        rows: &[
                            &[T("1a82b83c84d85e")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "concat",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT LENGTH(CONCAT('', NULL, ''));",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT CONCAT('a', NULL, 'b');",
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT)],
                        rows: &[
                            &[T("ab")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT CONCAT(1, 2, true);",
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT)],
                        rows: &[
                            &[T("12t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT CONCAT(1::int2, 2::int4, true::bool, false::text);",
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT)],
                        rows: &[
                            &[T("12tfalse")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT CONCAT('b');",
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT)],
                        rows: &[
                            &[T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT CONCAT(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("concat", TEXT)],
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
            name: "encode",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT encode('\x1234567890abcdef00'::bytea, 'hex');"#,
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("1234567890abcdef00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT encode('\x1234567890abcdef00'::bytea, 'base64');"#,
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("EjRWeJCrze8A")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT encode('\x1234567890abcdef00'::bytea, 'escape');"#,
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("\u{12}4Vx\\220\\253\\315\\357\\000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(''::bytea, 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode('hello'::bytea, 'escape');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("hello")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT encode('\x5c'::bytea, 'escape');"#,
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T(r#"\\"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode('hello'::bytea, 'bogus');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized encoding: "bogus""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(NULL, 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
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
            name: "convert_from and decode",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\x68656c6c6f'::BYTEA, 'UTF8'), convert_from(decode('68656c6c6f', 'hex'), 'UTF8'), convert_from('\xc3a9'::BYTEA, 'UTF8'), convert_from('\xe9'::BYTEA, 'LATIN1');"#,
                    expected: Expected::Rows {
                        columns: &[Column("convert_from", TEXT), Column("convert_from", TEXT), Column("convert_from", TEXT), Column("convert_from", TEXT)],
                        rows: &[
                            &[T("hello"), T("hello"), T("é"), T("é")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\xa4a2'::BYTEA, 'EUC_JP'), convert_from('\x82a0'::BYTEA, 'SJIS');"#,
                    expected: Expected::Rows {
                        columns: &[Column("convert_from", TEXT), Column("convert_from", TEXT)],
                        rows: &[
                            &[T("あ"), T("あ")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT decode('aGVsbG8=', 'base64'), decode('abc\000', 'escape'), decode('a\\b', 'escape'), decode('68 65', 'hex');"#,
                    expected: Expected::Rows {
                        columns: &[Column("decode", BYTEA), Column("decode", BYTEA), Column("decode", BYTEA), Column("decode", BYTEA)],
                        rows: &[
                            &[T(r#"\x68656c6c6f"#), T(r#"\x61626300"#), T(r#"\x615c62"#), T(r#"\x6865"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\xff'::BYTEA, 'UTF8');"#,
                    expected: Expected::Error(Diagnostic { code: "22021", message: r#"invalid byte sequence for encoding "UTF8": 0xff"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\xa4'::BYTEA, 'EUC_JP');"#,
                    expected: Expected::Error(Diagnostic { code: "22021", message: r#"invalid byte sequence for encoding "EUC_JP": 0xa4"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\x41a4'::BYTEA, 'EUC_JP');"#,
                    expected: Expected::Error(Diagnostic { code: "22021", message: r#"invalid byte sequence for encoding "EUC_JP": 0xa4"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\xa4a2ff'::BYTEA, 'EUC_JP');"#,
                    expected: Expected::Error(Diagnostic { code: "22021", message: r#"invalid byte sequence for encoding "EUC_JP": 0xff"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\x82'::BYTEA, 'SJIS');"#,
                    expected: Expected::Error(Diagnostic { code: "22021", message: r#"invalid byte sequence for encoding "SJIS": 0x82"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\x68'::BYTEA, 'NOPE');"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid source encoding name "NOPE""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT decode('6', 'hex');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "invalid hexadecimal data: odd number of digits", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT decode('6g', 'hex');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid hexadecimal digit: "g""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT decode('abc', 'nope');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized encoding: "nope""#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "format",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT format('hello'), format('%s', 'a'), format('x %s y %s', 'a', 'b'), format('100%% %s', 'a');",
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT), Column("format", TEXT), Column("format", TEXT), Column("format", TEXT)],
                        rows: &[
                            &[T("hello"), T("a"), T("x a y b"), T("100% a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%I', 'my table'), format('INSERT INTO %I VALUES (1)', 'log');",
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT), Column("format", TEXT)],
                        rows: &[
                            &[T(r#""my table""#), T("INSERT INTO log VALUES (1)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT format('%I %I %I %I', 'Abc', 'user', 'select', 'a"b');"#,
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT)],
                        rows: &[
                            &[T(r#""Abc" "user" "select" "a""b""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT format('%L', 'it''s'), format('%L', 'a\b'), format('%L', 12);"#,
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT), Column("format", TEXT), Column("format", TEXT)],
                        rows: &[
                            &[T("'it''s'"), T(r#"E'a\\b'"#), T("'12'")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%s %s', 1, true), format('%s %L', ARRAY[true,false], true), format('%s', 'a', 'b');",
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT), Column("format", TEXT), Column("format", TEXT)],
                        rows: &[
                            &[T("1 t"), T("{t,f} 't'"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format(NULL), format(NULL, 'a'), format('%s|%L|', NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT), Column("format", TEXT), Column("format", TEXT)],
                        rows: &[
                            &[Null, Null, T("|NULL|")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%2$s %1$s %s', 'a', 'b');",
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT)],
                        rows: &[
                            &[T("b a b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('|%5s|%-5s|%*s|%-*s|%*s|', 'ab', 'cd', 4, 'ef', 4, 'gh', -4, 'ij');",
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT)],
                        rows: &[
                            &[T("|   ab|cd   |  ef|gh  |ij  |")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('|%*s|', NULL, 'ab'), format('|%*s|', '3'::text, 'ab'), format('|%3s|', 'éé');",
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT), Column("format", TEXT), Column("format", TEXT)],
                        rows: &[
                            &[T("|ab|"), T("| ab|"), T("| éé|")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format(1234.5678, 2);",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function format(numeric, integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%I', NULL);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "null values cannot be formatted as an SQL identifier", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%s');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "too few arguments for format()", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('|%*2$s|%1$*2$s|', 'ab', 5);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "too few arguments for format()", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "unterminated format() type specifier", hint: r#"For a single "%" use "%%"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%1');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "unterminated format() type specifier", hint: r#"For a single "%" use "%%"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%d', 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized format() type specifier "d""#, hint: r#"For a single "%" use "%%"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%é', 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized format() type specifier "é""#, hint: r#"For a single "%" use "%%"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%0$s', 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "format specifies argument 0, but arguments are numbered from 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%*0$s', 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "format specifies argument 0, but arguments are numbered from 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%*1s', 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"width argument position must be ended by "$""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%99999999999s', 1);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "number is out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('|%*s|', 'x'::text, 'ab');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "x""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%s %s', VARIADIC ARRAY['a', 'b']);",
                    expected: Expected::Rows {
                        columns: &[Column("format", TEXT)],
                        rows: &[
                            &[T("a b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%2147483647s', 'a');",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "out of memory", detail: "Cannot enlarge string buffer containing 0 bytes by 2147483646 more bytes.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%*s', -2147483647, 'a');",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "out of memory", detail: "Cannot enlarge string buffer containing 1 bytes by 2147483646 more bytes.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%*s', -2147483648, 'a');",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "number is out of range", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format('%1073741820s', 'a');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "invalid memory alloc request size 1073741824", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "format builds dynamic SQL in a trigger function",
            set_up_script: &[
                "CREATE TABLE t (id TEXT PRIMARY KEY);",
                "CREATE TABLE log (v TEXT);",
                r#"CREATE FUNCTION f() RETURNS TRIGGER LANGUAGE plpgsql AS $f$
BEGIN EXECUTE format('INSERT INTO %I VALUES (%L)', 'log', NEW.id); RETURN NULL; END $f$;"#,
                "CREATE TRIGGER tr AFTER INSERT ON t FOR EACH ROW EXECUTE FUNCTION f();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES ('a');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM log;",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT)],
                        rows: &[
                            &[T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "format builds dynamic SQL in functions",
            set_up_script: &[
                r#"CREATE FUNCTION make_table(name TEXT) RETURNS TEXT LANGUAGE plpgsql AS $$
BEGIN
	IF to_regclass(format('%I', name)) IS NULL THEN
		EXECUTE format('CREATE TABLE %I (id INT PRIMARY KEY, note TEXT)', name);
		RETURN 'created';
	END IF;
	RETURN 'exists';
END;
$$;"#,
                r#"CREATE FUNCTION add_note(name TEXT, id INT, note TEXT) RETURNS TEXT LANGUAGE plpgsql AS $$
DECLARE
	stmt TEXT := format('INSERT INTO %I VALUES (%s, %L)', name, id, note);
BEGIN
	EXECUTE stmt;
	RETURN stmt;
END;
$$;"#,
                r#"CREATE FUNCTION count_rows(name TEXT) RETURNS BIGINT LANGUAGE plpgsql AS $$
DECLARE
	n BIGINT;
BEGIN
	EXECUTE format('SELECT count(*) FROM %I', name) INTO n;
	RETURN n;
END;
$$;"#,
                r#"CREATE FUNCTION count_notes(name TEXT, note TEXT) RETURNS BIGINT LANGUAGE plpgsql AS $$
DECLARE
	n BIGINT;
BEGIN
	EXECUTE format('SELECT count(*) FROM %I WHERE note = $1', name) INTO n USING note;
	RETURN n;
END;
$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT make_table('My Table');",
                    expected: Expected::Rows {
                        columns: &[Column("make_table", TEXT)],
                        rows: &[
                            &[T("created")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_table('My Table');",
                    expected: Expected::Rows {
                        columns: &[Column("make_table", TEXT)],
                        rows: &[
                            &[T("exists")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_table('plain');",
                    expected: Expected::Rows {
                        columns: &[Column("make_table", TEXT)],
                        rows: &[
                            &[T("created")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT add_note('My Table', 1, 'it''s');",
                    expected: Expected::Rows {
                        columns: &[Column("add_note", TEXT)],
                        rows: &[
                            &[T(r#"INSERT INTO "My Table" VALUES (1, 'it''s')"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT add_note('My Table', 2, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("add_note", TEXT)],
                        rows: &[
                            &[T(r#"INSERT INTO "My Table" VALUES (2, NULL)"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT add_note('My Table', 3, 'back\slash');"#,
                    expected: Expected::Rows {
                        columns: &[Column("add_note", TEXT)],
                        rows: &[
                            &[T(r#"INSERT INTO "My Table" VALUES (3, E'back\\slash')"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "My Table" ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("note", TEXT)],
                        rows: &[
                            &[T("1"), T("it's")],
                            &[T("2"), Null],
                            &[T("3"), T(r#"back\slash"#)],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count_rows('My Table'), count_rows('plain');",
                    expected: Expected::Rows {
                        columns: &[Column("count_rows", INT8), Column("count_rows", INT8)],
                        rows: &[
                            &[T("3"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count_notes('My Table', 'it''s'), count_notes('My Table', 'nope');",
                    expected: Expected::Rows {
                        columns: &[Column("count_notes", INT8), Column("count_notes", INT8)],
                        rows: &[
                            &[T("1"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"DO $$
BEGIN
	EXECUTE format('UPDATE %1$I SET note = %2$L WHERE note IS NULL OR note <> %2$L', 'My Table', 'same');
END;
$$;"#,
                    expected: Expected::Tag("DO"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM "My Table" ORDER BY id;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("note", TEXT)],
                        rows: &[
                            &[T("1"), T("same")],
                            &[T("2"), T("same")],
                            &[T("3"), T("same")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT add_note('missing', 1, 'x');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "missing" does not exist"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "format builds dynamic SQL in an audit trigger",
            set_up_script: &[
                "CREATE TABLE src (id TEXT PRIMARY KEY);",
                "CREATE TABLE audit (tbl TEXT, op TEXT, id TEXT);",
                r#"CREATE FUNCTION audit_row() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
	EXECUTE format('INSERT INTO %I VALUES (%L, %L, %L)', 'audit', TG_TABLE_NAME, TG_OP, NEW.id);
	RETURN NEW;
END;
$$;"#,
                "CREATE TRIGGER src_audit AFTER INSERT OR UPDATE ON src FOR EACH ROW EXECUTE FUNCTION audit_row();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO src VALUES ('a'), ('b''c');",
                    expected: Expected::Tag("INSERT 0 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE src SET id = 'd' WHERE id = 'a';",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM audit ORDER BY op, id;",
                    expected: Expected::Rows {
                        columns: &[Column("tbl", TEXT), Column("op", TEXT), Column("id", TEXT)],
                        rows: &[
                            &[T("src"), T("INSERT"), T("a")],
                            &[T("src"), T("INSERT"), T("b'c")],
                            &[T("src"), T("UPDATE"), T("d")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_system_catalog_information_functions() {
    run_scripts(&[
        ScriptTest {
            name: "getdatabaseencoding",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT getdatabaseencoding();",
                    expected: Expected::Rows {
                        columns: &[Column("getdatabaseencoding", NAME)],
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
            name: "pg_encoding_to_char",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_encoding_to_char(encoding) FROM pg_database WHERE datname = 'postgres';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_encoding_to_char", NAME)],
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
            name: "pg_char_to_encoding",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_char_to_encoding('UTF8');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_char_to_encoding", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_char_to_encoding('utf-8');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_char_to_encoding", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_char_to_encoding('LATIN1');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_char_to_encoding", INT4)],
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
            name: "pg_get_function_arguments",
            set_up_script: &[
                "CREATE FUNCTION alt_func1(int) RETURNS int LANGUAGE sql AS 'SELECT $1 + 1';",
                "CREATE TABLE cp_test (a int, b text);",
                r#"CREATE OR REPLACE PROCEDURE ptest5(a int, b text, c int default 100)
				LANGUAGE SQL
				AS $$
					INSERT INTO cp_test VALUES(a, b);
					INSERT INTO cp_test VALUES(c, b);
				$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_function_arguments(22)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_function_arguments", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT oid, proname FROM pg_catalog.pg_proc WHERE proname = 'alt_func1' OR proname = 'ptest5';",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("proname", NAME)],
                        rows: &[
                            &[Oid(16384), T("alt_func1")],
                            &[Oid(16390), T("ptest5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_function_arguments(2891346960)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_function_arguments", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("the OID literal names a Doltgres routine, and Postgres assigns its routines other OIDs"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_function_arguments(1886569565)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_function_arguments", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("the OID literal names a Doltgres routine, and Postgres assigns its routines other OIDs"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_functiondef",
            set_up_script: &[
                "CREATE FUNCTION alt_func1(int) RETURNS int LANGUAGE sql AS 'SELECT $1 + 1';",
                "CREATE TABLE cp_test (a int, b text);",
                r#"CREATE OR REPLACE PROCEDURE ptest5(a int, b text, c int default 100)
				LANGUAGE SQL
				AS $$
					INSERT INTO cp_test VALUES(a, b);
					INSERT INTO cp_test VALUES(c, b);
				$$;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_functiondef(22)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_functiondef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_functiondef(2891346960)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_functiondef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("the OID literal names a Doltgres routine, and Postgres assigns its routines other OIDs"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT oid, proname FROM pg_catalog.pg_proc WHERE proname = 'alt_func1' OR proname = 'ptest5';",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("proname", NAME)],
                        rows: &[
                            &[Oid(16384), T("alt_func1")],
                            &[Oid(16390), T("ptest5")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_functiondef(2891346960)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_functiondef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("the OID literal names a Doltgres routine, and Postgres assigns its routines other OIDs"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_functiondef(1886569565)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_functiondef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("the OID literal names a Doltgres routine, and Postgres assigns its routines other OIDs"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_indexdef",
            set_up_script: &[
                "CREATE TABLE idx_test (pk INT PRIMARY KEY, a INT, b INT, c TEXT);",
                "CREATE UNIQUE INDEX idx_ab ON idx_test (a, b);",
                "CREATE INDEX idx_c ON idx_test (c);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef(845743985);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_test_pkey'::regclass::oid);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE UNIQUE INDEX idx_test_pkey ON public.idx_test USING btree (pk)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_ab'::regclass::oid);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE UNIQUE INDEX idx_ab ON public.idx_test USING btree (a, b)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_ab'::regclass::oid, 0, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE UNIQUE INDEX idx_ab ON idx_test USING btree (a, b)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_ab'::regclass::oid, 0, false);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE UNIQUE INDEX idx_ab ON public.idx_test USING btree (a, b)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_ab'::regclass::oid, 1, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_ab'::regclass::oid, 2, false);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_ab'::regclass::oid, 3, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_ab'::regclass::oid, -1, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_c'::regclass::oid, 1, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("c")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef(845743985, 0, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef(845743985, 1, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef(NULL, 1, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('idx_ab'::regclass::oid, NULL, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT c2.relname, i.indisprimary, i.indisunique, i.indisclustered, i.indisvalid, pg_catalog.pg_get_indexdef(i.indexrelid, 0, true),
  pg_catalog.pg_get_constraintdef(con.oid, true), contype, condeferrable, condeferred, i.indisreplident, c2.reltablespace
FROM pg_catalog.pg_class c, pg_catalog.pg_class c2, pg_catalog.pg_index i
  LEFT JOIN pg_catalog.pg_constraint con ON (conrelid = i.indrelid AND conindid = i.indexrelid AND contype IN ('p','u','x'))
WHERE c.oid = 'idx_test'::regclass AND c.oid = i.indrelid AND i.indexrelid = c2.oid
ORDER BY i.indisprimary DESC, c2.relname;"#,
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME), Column("indisprimary", BOOL), Column("indisunique", BOOL), Column("indisclustered", BOOL), Column("indisvalid", BOOL), Column("pg_get_indexdef", TEXT), Column("pg_get_constraintdef", TEXT), Column("contype", CHAR), Column("condeferrable", BOOL), Column("condeferred", BOOL), Column("indisreplident", BOOL), Column("reltablespace", OID)],
                        rows: &[
                            &[T("idx_test_pkey"), T("t"), T("t"), T("f"), T("t"), T("CREATE UNIQUE INDEX idx_test_pkey ON idx_test USING btree (pk)"), T("PRIMARY KEY (pk)"), T("p"), T("f"), T("f"), T("f"), T("0")],
                            &[T("idx_ab"), T("f"), T("t"), T("f"), T("t"), T("CREATE UNIQUE INDEX idx_ab ON idx_test USING btree (a, b)"), Null, Null, Null, Null, T("f"), T("0")],
                            &[T("idx_c"), T("f"), T("f"), T("f"), T("t"), T("CREATE INDEX idx_c ON idx_test USING btree (c)"), Null, Null, Null, Null, T("f"), T("0")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_function_result",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_function_result(22)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_function_result", TEXT)],
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
            name: "pg_get_triggerdef",
            set_up_script: &[
                "CREATE TABLE trig_test (pk INT PRIMARY KEY, v1 INT);",
                "CREATE FUNCTION trig_fn() RETURNS trigger AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;",
                "CREATE TRIGGER trig_before BEFORE INSERT ON trig_test FOR EACH ROW EXECUTE FUNCTION trig_fn();",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_triggerdef(22)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_triggerdef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_triggerdef(t.oid) FROM pg_catalog.pg_trigger t WHERE t.tgname = 'trig_before';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_triggerdef", TEXT)],
                        rows: &[
                            &[T("CREATE TRIGGER trig_before BEFORE INSERT ON public.trig_test FOR EACH ROW EXECUTE FUNCTION trig_fn()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_triggerdef(t.oid, true) FROM pg_catalog.pg_trigger t WHERE t.tgname = 'trig_before';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_triggerdef", TEXT)],
                        rows: &[
                            &[T("CREATE TRIGGER trig_before BEFORE INSERT ON trig_test FOR EACH ROW EXECUTE FUNCTION trig_fn()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_triggerdef(t.oid, false) FROM pg_catalog.pg_trigger t WHERE t.tgname = 'trig_before';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_triggerdef", TEXT)],
                        rows: &[
                            &[T("CREATE TRIGGER trig_before BEFORE INSERT ON public.trig_test FOR EACH ROW EXECUTE FUNCTION trig_fn()")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_userbyid",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_userbyid(22)",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_userbyid", NAME)],
                        rows: &[
                            &[T("unknown (OID=22)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_viewdef",
            set_up_script: &[
                "CREATE TABLE test (id int, name text)",
                "INSERT INTO test VALUES (1,'desk'), (2,'chair')",
                "CREATE VIEW test_view AS SELECT name FROM test",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT c.oid, c.relname AS table_name, n.nspname AS table_schema FROM pg_catalog.pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE (n.nspname='myschema' OR n.nspname='public') and left(relname, 5) <> 'dolt_';",
                    expected: Expected::Rows {
                        columns: &[Column("oid", OID), Column("table_name", NAME), Column("table_schema", NAME)],
                        rows: &[
                            &[Oid(16384), T("test"), T("public")],
                            &[Oid(16389), T("test_view"), T("public")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select pg_get_viewdef(2707638987);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_viewdef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("the OID literal names a Doltgres view, and Postgres assigns its views other OIDs"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_system_information_functions() {
    run_scripts(&[
        ScriptTest {
            name: "pg_typeof",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(42), pg_typeof('abc'::text), pg_typeof(ARRAY['a']);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("integer"), T("text"), T("text[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(NULL::int), pg_typeof(NULL::text[]);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("integer"), T("text[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("unknown")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "current_database",
            database: "test",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_database();",
                    expected: Expected::Rows {
                        columns: &[Column("current_database", NAME)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_database;",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "current_database" does not exist"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM current_database();",
                    expected: Expected::Rows {
                        columns: &[Column("current_database", NAME)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM current_database;",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "current_database" does not exist"#, position: 15, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "current_catalog",
            database: "test",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_catalog;",
                    expected: Expected::Rows {
                        columns: &[Column("current_catalog", NAME)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_catalog;",
                    expected: Expected::Rows {
                        columns: &[Column("current_catalog", NAME)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_catalog();",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "(""#, position: 23, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM current_catalog;",
                    expected: Expected::Rows {
                        columns: &[Column("current_catalog", NAME)],
                        rows: &[
                            &[T("test")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM current_catalog();",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "(""#, position: 30, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "current_schema",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA test_schema;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEARCH_PATH TO test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEARCH_PATH TO public, test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEARCH_PATH TO test_schema, public;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM current_schema;",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "current_schemas",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{pg_catalog,public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA test_schema;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEARCH_PATH TO test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{pg_catalog,test_schema}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{test_schema}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET SEARCH_PATH TO public, test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{pg_catalog,public,test_schema}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{public,test_schema}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "current_schema with nonexistent schemas on the search_path",
            set_up_script: &[
                "CREATE SCHEMA test_schema;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SET search_path TO does_not_exist, test_schema, public;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{test_schema,public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{pg_catalog,test_schema,public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO does_not_exist;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{pg_catalog}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO test_schema, public, test_schema;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{test_schema,public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO public, pg_catalog;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(true);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{public,pg_catalog}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO TEST_SCHEMA;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("test_schema")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: r#"current_schema with "$user" on the search_path"#,
            assertions: &[
                ScriptTestAssertion {
                    query: "SHOW search_path;",
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
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA postgres;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
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
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{postgres,public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SET search_path TO "$user", public;"#,
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SHOW search_path;",
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
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
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
            name: r#""$user" on the search_path expands to the connected user"#,
            set_up_script: &[
                "CREATE USER tester PASSWORD 'password';",
                "CREATE SCHEMA tester;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "postgres",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schema();",
                    expected: Expected::Rows {
                        columns: &[Column("current_schema", NAME)],
                        rows: &[
                            &[T("public")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "tester",
                    password: "password",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_schemas(false);",
                    expected: Expected::Rows {
                        columns: &[Column("current_schemas", NAME_ARRAY)],
                        rows: &[
                            &[T("{public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    username: "tester",
                    password: "password",
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "version",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT version();",
                    expected: Expected::Rows {
                        columns: &[Column("version", TEXT)],
                        rows: &[
                            &[T("PostgreSQL 15.19 (Homebrew) on aarch64-apple-darwin24.6.0, compiled by Apple clang version 17.0.0 (clang-1700.6.4.2), 64-bit")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("version() names the platform and compiler of the build, which differ from the machine that recorded this"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "col_description",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT col_description(100, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("col_description", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT col_description('not_a_table'::regclass, 1);",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "not_a_table" does not exist"#, position: 24, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test_table (id INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMENT ON COLUMN test_table.id IS 'This is col id';",
                    expected: Expected::Tag("COMMENT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT col_description('test_table'::regclass, 1);",
                    expected: Expected::Rows {
                        columns: &[Column("col_description", TEXT)],
                        rows: &[
                            &[T("This is col id")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "obj_description",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT obj_description(1003);",
                    expected: Expected::Rows {
                        columns: &[Column("obj_description", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT obj_description(100, 'pg_class');",
                    expected: Expected::Rows {
                        columns: &[Column("obj_description", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT obj_description('does-not-exist'::regproc, 'pg_class');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"function "does-not-exist" does not exist"#, position: 24, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT obj_description('sinh'::regproc, 'pg_proc');",
                    expected: Expected::Rows {
                        columns: &[Column("obj_description", TEXT)],
                        rows: &[
                            &[T("hyperbolic sine")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "shobj_description",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT shobj_description(100, 'pg_class');",
                    expected: Expected::Rows {
                        columns: &[Column("shobj_description", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT shobj_description('does-not-exist'::regproc, 'pg_class');",
                    expected: Expected::Error(Diagnostic { code: "42883", message: r#"function "does-not-exist" does not exist"#, position: 26, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLESPACE tblspc_2 LOCATION '/';",
                    skip: Some("Doltgres has no tablespaces, and Postgres' error here comes from the recording server's file permissions"),
                    expected: Expected::Error(Diagnostic { code: "42501", message: r#"could not set permissions on directory "/": Operation not permitted"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMENT ON TABLESPACE tblspc_2 IS 'Store a few of the things';",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"tablespace "tblspc_2" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT shobj_description(
                 (SELECT oid FROM pg_tablespace WHERE spcname = 'tblspc_2'),
                 'pg_tablespace');"#,
                    expected: Expected::Rows {
                        columns: &[Column("shobj_description", TEXT)],
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
            name: "format_type",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT format_type('integer'::regtype, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("integer")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('character varying'::regtype, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("character varying")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('varchar'::regtype, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("character varying")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('date'::regtype, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("date")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('timestamptz'::regtype, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("timestamp with time zone")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('bool'::regtype, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("boolean")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type(1007, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("integer[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT format_type('"char"'::regtype, null);"#,
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T(r#""char""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT format_type('"char"[]'::regtype, null);"#,
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T(r#""char"[]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type(1002, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T(r#""char"[]"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('real[]'::regtype, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("real[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('character varying'::regtype, 100);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("character varying(96)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('text'::regtype, 0);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("text(0)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('text'::regtype, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("text(4)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('text'::regtype, -1);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("text")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('name'::regtype, 0);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("name(0)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('bpchar'::regtype, -1);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("bpchar")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('bpchar'::regtype, 10);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("character(6)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('bpchar'::regtype, 10);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("character(6)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('character'::regtype, 4);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("character")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('varchar'::regtype, 0);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("character varying")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT format_type('"char"'::regtype, 0);"#,
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T(r#""char"(0)"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type('numeric'::regtype, 12);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("numeric(0,8)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type(874938247, 20);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("???")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT format_type(874938247, null);",
                    expected: Expected::Rows {
                        columns: &[Column("format_type", TEXT)],
                        rows: &[
                            &[T("???")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_constraintdef",
            set_up_script: &[
                "CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);",
                "CREATE TABLE testing2 (pk INT primary key, pktesting INT REFERENCES testing(pk), v1 TEXT);",
                "CREATE TABLE testing3 (pk1 INT, pk2 INT, PRIMARY KEY (pk1, pk2));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(845743985);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conrelid='testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("PRIMARY KEY (pk)")],
                            &[T("UNIQUE (v1)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conrelid='testing2'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("PRIMARY KEY (pk)")],
                            &[T("FOREIGN KEY (pktesting) REFERENCES testing(pk)")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conrelid='testing3'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("PRIMARY KEY (pk1, pk2)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid, true) FROM pg_catalog.pg_constraint WHERE conrelid='testing3'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("PRIMARY KEY (pk1, pk2)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_constraintdef(oid, false) FROM pg_catalog.pg_constraint WHERE conrelid='testing3'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_constraintdef", TEXT)],
                        rows: &[
                            &[T("PRIMARY KEY (pk1, pk2)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_ruledef",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_ruledef(845743985);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_ruledef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_ruledef(845743985, true);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_ruledef", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_ruledef(845743985, false);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_ruledef", TEXT)],
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
            name: "pg_get_expr",
            set_up_script: &[
                "CREATE TABLE testing (id INT primary key);",
                "CREATE TABLE temperature (celsius SMALLINT NOT NULL, fahrenheit SMALLINT NOT NULL GENERATED ALWAYS AS ((celsius * 9/5) + 32) STORED);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_expr(adbin, adrelid) FROM pg_catalog.pg_attrdef WHERE adrelid = 'temperature'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_expr", TEXT)],
                        rows: &[
                            &[T("(((celsius * 9) / 5) + 32)")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indexrelid, pg_get_expr(indpred, indrelid) FROM pg_catalog.pg_index WHERE indrelid='testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("indexrelid", OID), Column("pg_get_expr", TEXT)],
                        rows: &[
                            &[Oid(16387), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indexrelid, pg_get_expr(indpred, indrelid, true) FROM pg_catalog.pg_index WHERE indrelid='testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("indexrelid", OID), Column("pg_get_expr", TEXT)],
                        rows: &[
                            &[Oid(16387), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indexrelid, pg_get_expr(indpred, indrelid, NULL) FROM pg_catalog.pg_index WHERE indrelid='testing'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("indexrelid", OID), Column("pg_get_expr", TEXT)],
                        rows: &[
                            &[Oid(16387), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_serial_sequence",
            set_up_script: &[
                "create table t0 (id INTEGER NOT NULL PRIMARY KEY);",
                "create table t1 (id SERIAL PRIMARY KEY);",
                "create sequence t2_id_seq START 1 INCREMENT 3;",
                "create table t2 (id INTEGER NOT NULL DEFAULT nextval('t2_id_seq'));",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT pg_get_serial_sequence('doesnotexist.t1', 'id');",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "doesnotexist" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_serial_sequence('doesnotexist', 'id');",
                    expected: Expected::Error(Diagnostic { code: "42P01", message: r#"relation "doesnotexist" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_serial_sequence('t0', 'doesnotexist');",
                    expected: Expected::Error(Diagnostic { code: "42703", message: r#"column "doesnotexist" of relation "t0" does not exist"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_serial_sequence('t0', 'id');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_serial_sequence", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_serial_sequence('public.t1', 'id');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_serial_sequence", TEXT)],
                        rows: &[
                            &[T("public.t1_id_seq")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT pg_get_serial_sequence('"public"."t1"', 'id');"#,
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_serial_sequence", TEXT)],
                        rows: &[
                            &[T("public.t1_id_seq")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_serial_sequence('t1', 'id');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_serial_sequence", TEXT)],
                        rows: &[
                            &[T("public.t1_id_seq")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_serial_sequence('t2', 'id');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_serial_sequence", TEXT)],
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
            name: "current_setting function",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET timezone TO '+00:00';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('timezone')",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[T("+00:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wrong_input')",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "wrong_input""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wrong_input', true)",
                    expected: Expected::Rows {
                        columns: &[Column("current_setting", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT current_setting('wrong_input', false)",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"unrecognized configuration parameter "wrong_input""#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_show_all_settings",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_show_all_settings();",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("allow_in_place_tablespaces")],
                            &[T("allow_system_table_mods")],
                            &[T("application_name")],
                            &[T("archive_cleanup_command")],
                            &[T("archive_command")],
                            &[T("archive_library")],
                            &[T("archive_mode")],
                            &[T("archive_timeout")],
                            &[T("array_nulls")],
                            &[T("authentication_timeout")],
                            &[T("autovacuum")],
                            &[T("autovacuum_analyze_scale_factor")],
                            &[T("autovacuum_analyze_threshold")],
                            &[T("autovacuum_freeze_max_age")],
                            &[T("autovacuum_max_workers")],
                            &[T("autovacuum_multixact_freeze_max_age")],
                            &[T("autovacuum_naptime")],
                            &[T("autovacuum_vacuum_cost_delay")],
                            &[T("autovacuum_vacuum_cost_limit")],
                            &[T("autovacuum_vacuum_insert_scale_factor")],
                            &[T("autovacuum_vacuum_insert_threshold")],
                            &[T("autovacuum_vacuum_scale_factor")],
                            &[T("autovacuum_vacuum_threshold")],
                            &[T("autovacuum_work_mem")],
                            &[T("backend_flush_after")],
                            &[T("backslash_quote")],
                            &[T("backtrace_functions")],
                            &[T("bgwriter_delay")],
                            &[T("bgwriter_flush_after")],
                            &[T("bgwriter_lru_maxpages")],
                            &[T("bgwriter_lru_multiplier")],
                            &[T("block_size")],
                            &[T("bonjour")],
                            &[T("bonjour_name")],
                            &[T("bytea_output")],
                            &[T("check_function_bodies")],
                            &[T("checkpoint_completion_target")],
                            &[T("checkpoint_flush_after")],
                            &[T("checkpoint_timeout")],
                            &[T("checkpoint_warning")],
                            &[T("client_connection_check_interval")],
                            &[T("client_encoding")],
                            &[T("client_min_messages")],
                            &[T("cluster_name")],
                            &[T("commit_delay")],
                            &[T("commit_siblings")],
                            &[T("compute_query_id")],
                            &[T("config_file")],
                            &[T("constraint_exclusion")],
                            &[T("cpu_index_tuple_cost")],
                            &[T("cpu_operator_cost")],
                            &[T("cpu_tuple_cost")],
                            &[T("cursor_tuple_fraction")],
                            &[T("data_checksums")],
                            &[T("data_directory")],
                            &[T("data_directory_mode")],
                            &[T("data_sync_retry")],
                            &[T("DateStyle")],
                            &[T("db_user_namespace")],
                            &[T("deadlock_timeout")],
                            &[T("debug_assertions")],
                            &[T("debug_discard_caches")],
                            &[T("debug_pretty_print")],
                            &[T("debug_print_parse")],
                            &[T("debug_print_plan")],
                            &[T("debug_print_rewritten")],
                            &[T("default_statistics_target")],
                            &[T("default_table_access_method")],
                            &[T("default_tablespace")],
                            &[T("default_text_search_config")],
                            &[T("default_toast_compression")],
                            &[T("default_transaction_deferrable")],
                            &[T("default_transaction_isolation")],
                            &[T("default_transaction_read_only")],
                            &[T("dynamic_library_path")],
                            &[T("dynamic_shared_memory_type")],
                            &[T("effective_cache_size")],
                            &[T("effective_io_concurrency")],
                            &[T("enable_async_append")],
                            &[T("enable_bitmapscan")],
                            &[T("enable_gathermerge")],
                            &[T("enable_hashagg")],
                            &[T("enable_hashjoin")],
                            &[T("enable_incremental_sort")],
                            &[T("enable_indexonlyscan")],
                            &[T("enable_indexscan")],
                            &[T("enable_material")],
                            &[T("enable_memoize")],
                            &[T("enable_mergejoin")],
                            &[T("enable_nestloop")],
                            &[T("enable_parallel_append")],
                            &[T("enable_parallel_hash")],
                            &[T("enable_partition_pruning")],
                            &[T("enable_partitionwise_aggregate")],
                            &[T("enable_partitionwise_join")],
                            &[T("enable_seqscan")],
                            &[T("enable_sort")],
                            &[T("enable_tidscan")],
                            &[T("escape_string_warning")],
                            &[T("event_source")],
                            &[T("exit_on_error")],
                            &[T("external_pid_file")],
                            &[T("extra_float_digits")],
                            &[T("force_parallel_mode")],
                            &[T("from_collapse_limit")],
                            &[T("fsync")],
                            &[T("full_page_writes")],
                            &[T("geqo")],
                            &[T("geqo_effort")],
                            &[T("geqo_generations")],
                            &[T("geqo_pool_size")],
                            &[T("geqo_seed")],
                            &[T("geqo_selection_bias")],
                            &[T("geqo_threshold")],
                            &[T("gin_fuzzy_search_limit")],
                            &[T("gin_pending_list_limit")],
                            &[T("hash_mem_multiplier")],
                            &[T("hba_file")],
                            &[T("hot_standby")],
                            &[T("hot_standby_feedback")],
                            &[T("huge_page_size")],
                            &[T("huge_pages")],
                            &[T("ident_file")],
                            &[T("idle_in_transaction_session_timeout")],
                            &[T("idle_session_timeout")],
                            &[T("ignore_checksum_failure")],
                            &[T("ignore_invalid_pages")],
                            &[T("ignore_system_indexes")],
                            &[T("in_hot_standby")],
                            &[T("integer_datetimes")],
                            &[T("IntervalStyle")],
                            &[T("jit")],
                            &[T("jit_above_cost")],
                            &[T("jit_debugging_support")],
                            &[T("jit_dump_bitcode")],
                            &[T("jit_expressions")],
                            &[T("jit_inline_above_cost")],
                            &[T("jit_optimize_above_cost")],
                            &[T("jit_profiling_support")],
                            &[T("jit_provider")],
                            &[T("jit_tuple_deforming")],
                            &[T("join_collapse_limit")],
                            &[T("krb_caseins_users")],
                            &[T("krb_server_keyfile")],
                            &[T("lc_collate")],
                            &[T("lc_ctype")],
                            &[T("lc_messages")],
                            &[T("lc_monetary")],
                            &[T("lc_numeric")],
                            &[T("lc_time")],
                            &[T("listen_addresses")],
                            &[T("lo_compat_privileges")],
                            &[T("local_preload_libraries")],
                            &[T("lock_timeout")],
                            &[T("log_autovacuum_min_duration")],
                            &[T("log_checkpoints")],
                            &[T("log_connections")],
                            &[T("log_destination")],
                            &[T("log_directory")],
                            &[T("log_disconnections")],
                            &[T("log_duration")],
                            &[T("log_error_verbosity")],
                            &[T("log_executor_stats")],
                            &[T("log_file_mode")],
                            &[T("log_filename")],
                            &[T("log_hostname")],
                            &[T("log_line_prefix")],
                            &[T("log_lock_waits")],
                            &[T("log_min_duration_sample")],
                            &[T("log_min_duration_statement")],
                            &[T("log_min_error_statement")],
                            &[T("log_min_messages")],
                            &[T("log_parameter_max_length")],
                            &[T("log_parameter_max_length_on_error")],
                            &[T("log_parser_stats")],
                            &[T("log_planner_stats")],
                            &[T("log_recovery_conflict_waits")],
                            &[T("log_replication_commands")],
                            &[T("log_rotation_age")],
                            &[T("log_rotation_size")],
                            &[T("log_startup_progress_interval")],
                            &[T("log_statement")],
                            &[T("log_statement_sample_rate")],
                            &[T("log_statement_stats")],
                            &[T("log_temp_files")],
                            &[T("log_timezone")],
                            &[T("log_transaction_sample_rate")],
                            &[T("log_truncate_on_rotation")],
                            &[T("logging_collector")],
                            &[T("logical_decoding_work_mem")],
                            &[T("maintenance_io_concurrency")],
                            &[T("maintenance_work_mem")],
                            &[T("max_connections")],
                            &[T("max_files_per_process")],
                            &[T("max_function_args")],
                            &[T("max_identifier_length")],
                            &[T("max_index_keys")],
                            &[T("max_locks_per_transaction")],
                            &[T("max_logical_replication_workers")],
                            &[T("max_parallel_maintenance_workers")],
                            &[T("max_parallel_workers")],
                            &[T("max_parallel_workers_per_gather")],
                            &[T("max_pred_locks_per_page")],
                            &[T("max_pred_locks_per_relation")],
                            &[T("max_pred_locks_per_transaction")],
                            &[T("max_prepared_transactions")],
                            &[T("max_replication_slots")],
                            &[T("max_slot_wal_keep_size")],
                            &[T("max_stack_depth")],
                            &[T("max_standby_archive_delay")],
                            &[T("max_standby_streaming_delay")],
                            &[T("max_sync_workers_per_subscription")],
                            &[T("max_wal_senders")],
                            &[T("max_wal_size")],
                            &[T("max_worker_processes")],
                            &[T("min_dynamic_shared_memory")],
                            &[T("min_parallel_index_scan_size")],
                            &[T("min_parallel_table_scan_size")],
                            &[T("min_wal_size")],
                            &[T("old_snapshot_threshold")],
                            &[T("output_plugin_libraries")],
                            &[T("parallel_leader_participation")],
                            &[T("parallel_setup_cost")],
                            &[T("parallel_tuple_cost")],
                            &[T("password_encryption")],
                            &[T("plan_cache_mode")],
                            &[T("port")],
                            &[T("post_auth_delay")],
                            &[T("pre_auth_delay")],
                            &[T("primary_conninfo")],
                            &[T("primary_slot_name")],
                            &[T("promote_trigger_file")],
                            &[T("quote_all_identifiers")],
                            &[T("random_page_cost")],
                            &[T("recovery_end_command")],
                            &[T("recovery_init_sync_method")],
                            &[T("recovery_min_apply_delay")],
                            &[T("recovery_prefetch")],
                            &[T("recovery_target")],
                            &[T("recovery_target_action")],
                            &[T("recovery_target_inclusive")],
                            &[T("recovery_target_lsn")],
                            &[T("recovery_target_name")],
                            &[T("recovery_target_time")],
                            &[T("recovery_target_timeline")],
                            &[T("recovery_target_xid")],
                            &[T("recursive_worktable_factor")],
                            &[T("remove_temp_files_after_crash")],
                            &[T("restart_after_crash")],
                            &[T("restore_command")],
                            &[T("restrict_nonsystem_relation_kind")],
                            &[T("row_security")],
                            &[T("search_path")],
                            &[T("segment_size")],
                            &[T("seq_page_cost")],
                            &[T("server_encoding")],
                            &[T("server_version")],
                            &[T("server_version_num")],
                            &[T("session_preload_libraries")],
                            &[T("session_replication_role")],
                            &[T("shared_buffers")],
                            &[T("shared_memory_size")],
                            &[T("shared_memory_size_in_huge_pages")],
                            &[T("shared_memory_type")],
                            &[T("shared_preload_libraries")],
                            &[T("ssl")],
                            &[T("ssl_ca_file")],
                            &[T("ssl_cert_file")],
                            &[T("ssl_ciphers")],
                            &[T("ssl_crl_dir")],
                            &[T("ssl_crl_file")],
                            &[T("ssl_dh_params_file")],
                            &[T("ssl_ecdh_curve")],
                            &[T("ssl_key_file")],
                            &[T("ssl_library")],
                            &[T("ssl_max_protocol_version")],
                            &[T("ssl_min_protocol_version")],
                            &[T("ssl_passphrase_command")],
                            &[T("ssl_passphrase_command_supports_reload")],
                            &[T("ssl_prefer_server_ciphers")],
                            &[T("standard_conforming_strings")],
                            &[T("statement_timeout")],
                            &[T("stats_fetch_consistency")],
                            &[T("superuser_reserved_connections")],
                            &[T("synchronize_seqscans")],
                            &[T("synchronous_commit")],
                            &[T("synchronous_standby_names")],
                            &[T("syslog_facility")],
                            &[T("syslog_ident")],
                            &[T("syslog_sequence_numbers")],
                            &[T("syslog_split_messages")],
                            &[T("tcp_keepalives_count")],
                            &[T("tcp_keepalives_idle")],
                            &[T("tcp_keepalives_interval")],
                            &[T("tcp_user_timeout")],
                            &[T("temp_buffers")],
                            &[T("temp_file_limit")],
                            &[T("temp_tablespaces")],
                            &[T("TimeZone")],
                            &[T("timezone_abbreviations")],
                            &[T("trace_notify")],
                            &[T("trace_recovery_messages")],
                            &[T("trace_sort")],
                            &[T("track_activities")],
                            &[T("track_activity_query_size")],
                            &[T("track_commit_timestamp")],
                            &[T("track_counts")],
                            &[T("track_functions")],
                            &[T("track_io_timing")],
                            &[T("track_wal_io_timing")],
                            &[T("transaction_deferrable")],
                            &[T("transaction_isolation")],
                            &[T("transaction_read_only")],
                            &[T("transform_null_equals")],
                            &[T("unix_socket_directories")],
                            &[T("unix_socket_group")],
                            &[T("unix_socket_permissions")],
                            &[T("update_process_title")],
                            &[T("vacuum_cost_delay")],
                            &[T("vacuum_cost_limit")],
                            &[T("vacuum_cost_page_dirty")],
                            &[T("vacuum_cost_page_hit")],
                            &[T("vacuum_cost_page_miss")],
                            &[T("vacuum_defer_cleanup_age")],
                            &[T("vacuum_failsafe_age")],
                            &[T("vacuum_freeze_min_age")],
                            &[T("vacuum_freeze_table_age")],
                            &[T("vacuum_multixact_failsafe_age")],
                            &[T("vacuum_multixact_freeze_min_age")],
                            &[T("vacuum_multixact_freeze_table_age")],
                            &[T("wal_block_size")],
                            &[T("wal_buffers")],
                            &[T("wal_compression")],
                            &[T("wal_consistency_checking")],
                            &[T("wal_decode_buffer_size")],
                            &[T("wal_init_zero")],
                            &[T("wal_keep_size")],
                            &[T("wal_level")],
                            &[T("wal_log_hints")],
                            &[T("wal_receiver_create_temp_slot")],
                            &[T("wal_receiver_status_interval")],
                            &[T("wal_receiver_timeout")],
                            &[T("wal_recycle")],
                            &[T("wal_retrieve_retry_interval")],
                            &[T("wal_segment_size")],
                            &[T("wal_sender_timeout")],
                            &[T("wal_skip_threshold")],
                            &[T("wal_sync_method")],
                            &[T("wal_writer_delay")],
                            &[T("wal_writer_flush_after")],
                            &[T("work_mem")],
                            &[T("xmlbinary")],
                            &[T("xmloption")],
                            &[T("zero_damaged_pages")],
                        ],
                        tag: "SELECT 354",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('bytea_output','hex',false) FROM pg_show_all_settings() WHERE name = 'bytea_output';",
                    expected: Expected::Rows {
                        columns: &[Column("set_config", TEXT)],
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
    ]);
}

#[test]
fn test_trim_array() {
    run_scripts(&[
        ScriptTest {
            name: "trim_array",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT trim_array(ARRAY[[1,2],[3,4],[5,6]],1), trim_array(ARRAY[1,2],2), trim_array(NULL::int[],1);",
                    expected: Expected::Rows {
                        columns: &[Column("trim_array", INT4_ARRAY), Column("trim_array", INT4_ARRAY), Column("trim_array", INT4_ARRAY)],
                        rows: &[
                            &[T("{{1,2},{3,4}}"), T("{}"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT trim_array(ARRAY[]::int[],0);",
                    expected: Expected::Rows {
                        columns: &[Column("trim_array", INT4_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT trim_array(ARRAY[[1,2],[3,4]],3);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "number of elements to trim must be between 0 and 2", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT trim_array(ARRAY[1],-1);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "number of elements to trim must be between 0 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, trim_array(a,0) FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("trim_array", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{3,NULL,1,3}")],
                            &[T("2"), T("{}")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT trim_array((SELECT a FROM array_inputs WHERE id=1),(SELECT 2));",
                    expected: Expected::Rows {
                        columns: &[Column("trim_array", INT4_ARRAY)],
                        rows: &[
                            &[T("{3,NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT trim_array(ARRAY[1],NULL),trim_array(ARRAY[[1,2],[3,4]],2);",
                    expected: Expected::Rows {
                        columns: &[Column("trim_array", INT4_ARRAY), Column("trim_array", INT4_ARRAY)],
                        rows: &[
                            &[Null, T("{}")],
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
fn test_unknown_functions() {
    run_scripts(&[
        ScriptTest {
            name: "unknown functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT unknown_func();",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function unknown_func() does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Unsupported group_concat syntax",
            set_up_script: &[
                "CREATE TABLE x (pk int)",
                "INSERT INTO x VALUES (1),(2),(3),(4),(NULL)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT group_concat(pk ORDER BY pk) FROM x;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function group_concat(integer) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_unnest_multidimensional_arguments() {
    run_scripts(&[
        ScriptTest {
            name: "multi-array unnest flattens each input",
            set_up_script: &[
                "CREATE TABLE array_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO array_inputs VALUES (1,ARRAY[3,NULL,1,3]),(2,ARRAY[]::int[]),(3,NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT unnest('1 2'::int2vector);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT2)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[[1,2],[3,4]],ARRAY['a','b']::varchar[]) AS u(n,label);",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("label", VARCHAR)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), T("b")],
                            &[T("3"), Null],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(NULL::int[],ARRAY[[1,2],[3,4]]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[Null, T("1")],
                            &[Null, T("2")],
                            &[Null, T("3")],
                            &[Null, T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, unnest(a) FROM array_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("unnest", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("1"), Null],
                            &[T("1"), T("1")],
                            &[T("1"), T("3")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest((SELECT a FROM array_inputs WHERE id=1),(SELECT ARRAY[9,8])) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("3"), T("9")],
                            &[Null, T("8")],
                            &[T("1"), Null],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[]::int[],NULL::int[]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[[[1,2]],[[3,4]]],ARRAY[9,8,7]) WITH ORDINALITY AS u(a,b,n);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("n", INT8)],
                        rows: &[
                            &[T("1"), T("9"), T("1")],
                            &[T("2"), T("8"), T("2")],
                            &[T("3"), T("7"), T("3")],
                            &[T("4"), Null, T("4")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[[NULL,NULL],[NULL,NULL]]::int[],ARRAY[9]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[Null, T("9")],
                            &[Null, Null],
                            &[Null, Null],
                            &[Null, Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[[1,2]],ARRAY[[3],[4],[5]]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("3")],
                            &[T("2"), T("4")],
                            &[Null, T("5")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[[1,2]],ARRAY[]::text[],ARRAY[true,false,true]) AS u(a,b,c);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT), Column("c", BOOL)],
                        rows: &[
                            &[T("1"), Null, T("t")],
                            &[T("2"), Null, T("f")],
                            &[Null, Null, T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[[[[[[1,2]]]]]],ARRAY[9]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("9")],
                            &[T("2"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[[1,2]],NULL::int[]) WITH ORDINALITY AS u(a,b,n);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("n", INT8)],
                        rows: &[
                            &[T("1"), Null, T("1")],
                            &[T("2"), Null, T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[]::int[],ARRAY[[1,2]]) WITH ORDINALITY AS u(a,b,n);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("n", INT8)],
                        rows: &[
                            &[Null, T("1"), T("1")],
                            &[Null, T("2"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[[1,1],[1,1]],ARRAY[2,2]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("2")],
                            &[T("1"), T("2")],
                            &[T("1"), Null],
                            &[T("1"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id,u.v,u.label,u.n FROM array_inputs,unnest(a,ARRAY[9,8]) WITH ORDINALITY AS u(v,label,n) ORDER BY id,n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4), Column("label", INT4), Column("n", INT8)],
                        rows: &[
                            &[T("1"), T("3"), T("9"), T("1")],
                            &[T("1"), Null, T("8"), T("2")],
                            &[T("1"), T("1"), Null, T("3")],
                            &[T("1"), T("3"), Null, T("4")],
                            &[T("2"), Null, T("9"), T("1")],
                            &[T("2"), Null, T("8"), T("2")],
                            &[T("3"), Null, T("9"), T("1")],
                            &[T("3"), Null, T("8"), T("2")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest((SELECT a FROM array_inputs WHERE id=2),ARRAY[9]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[Null, T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest((SELECT a FROM array_inputs WHERE id=3),ARRAY[9]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[Null, T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest((SELECT a FROM array_inputs WHERE id=99),ARRAY[9]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[Null, T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest((SELECT ARRAY[[1,2],[3,4]]),(SELECT ARRAY[NULL,9]::int[])) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), Null],
                            &[T("2"), T("9")],
                            &[T("3"), Null],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest((SELECT a FROM array_inputs WHERE id=1));",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4)],
                        rows: &[
                            &[T("3")],
                            &[Null],
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT unnest(ARRAY[[NULL,NULL],[1,NULL]]::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("unnest", INT4)],
                        rows: &[
                            &[Null],
                            &[Null],
                            &[T("1")],
                            &[Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[NULL]::int[],ARRAY[NULL]::text[]) WITH ORDINALITY AS u(a,b,n);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", TEXT), Column("n", INT8)],
                        rows: &[
                            &[Null, Null, T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest('1 2'::oidvector,ARRAY['a']::text[]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", OID), Column("b", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("2"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest('1 2'::int2vector,ARRAY[9,8,7]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT2), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("9")],
                            &[T("2"), T("8")],
                            &[Null, T("7")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM unnest(ARRAY[['a',NULL],['b','c']]::text[],ARRAY[[true,false]]) AS u(a,b);",
                    expected: Expected::Rows {
                        columns: &[Column("a", TEXT), Column("b", BOOL)],
                        rows: &[
                            &[T("a"), T("t")],
                            &[Null, T("f")],
                            &[T("b"), Null],
                            &[T("c"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_formatting_rules() {
    run_scripts(&[
        ScriptTest {
            name: "to_char templates",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP '2021-09-05 07:03:04.123456', 'FMDDth FMMonth YYYY, FMHH12:MI:SS.MS pm');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("5th September 2021, 7:03:04.123 am")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_char(TIMESTAMP '2021-09-05 07:03:04.123456', 'Dy DD Mon YY HH24 "quoted \"text\"" US FF1 FF2 FF4 FF5');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T(r#"Sun 05 Sep 21 07 quoted "text" 123456 1 12 1234 12345"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP '2021-01-03 00:00:00', 'IYYY-IW-ID IDDD DDD WW W Q CC J');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("2020-53-7 371 003 01 1 1 21 2459218")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP '2021-12-31 23:59:59', 'RM rm FMRM SSSS SSSSS');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("XII  xii  XII 86399 86399")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP '0044-03-15 12:00:00 BC', 'YYYY BC B.C. ad a.d. CC Y,YYY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("0044 BC B.C. bc b.c. -01 0,044")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP '2001-11-21 12:00:00', 'DDTH DDth MMTH YYYYth HH12TH');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("21ST 21st 11TH 2001st 12TH")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP '2021-09-15 21:43:56', 'TZ TZH:TZM OF');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T(" +00:00 +00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP 'infinity', 'YYYY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP '2021-09-15 21:43:56', '');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(DATE '2020-02-29', 'Day, FMMonth FMDDth YYYY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("Saturday , February 29th 2020")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_char of timestamptz and interval",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET timezone TO 'America/New_York';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMPTZ '2021-07-04 12:30:00+00', 'YYYY-MM-DD HH24:MI TZ tz OF TZH:TZM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("2021-07-04 08:30 EDT edt -04 -04:00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMPTZ '2021-01-04 12:30:00+00', 'YYYY-MM-DD HH24:MI TZ OF');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("2021-01-04 07:30 EST -05")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone TO '+05:30';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMPTZ '2021-01-04 12:30:00+00', 'HH24:MI TZ OF');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("07:00  -05:30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET timezone TO 'UTC';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(INTERVAL '1 year 2 months 3 days 04:05:06.789', 'YYYY MM DD HH24 MI SS MS US');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("0001 02 03 04 05 06 789 789000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(INTERVAL '-25 hours', 'HH24 HH12 HH');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("-25 -01 -01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(INTERVAL '15 months', 'Y RM Q');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("1 III  1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(INTERVAL '1 day', 'Day');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: "invalid format specification for an interval value", hint: "Intervals are not tied to specific calendar dates.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIME '13:14:15', 'HH12:MI:SS AM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("01:14:15 PM")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_timestamp and to_date",
            assertions: &[
                ScriptTestAssertion {
                    query: "SET timezone TO 'UTC';",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021-09-05 7:03 PM', 'YYYY-MM-DD HH12:MI AM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2021-09-05 19:03:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('05 Sep 2021', 'DD Mon YYYY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2021-09-05 00:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021 248', 'YYYY DDD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2021-09-05 00:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021-09-05 12:34:56.789123', 'YYYY-MM-DD HH24:MI:SS.FF3');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2021-09-05 12:34:56.789+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021-09-05 12:34 +02:30', 'YYYY-MM-DD HH24:MI TZH:TZM');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("2021-09-05 10:04:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('20 21', 'CC YY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_timestamp", TIMESTAMPTZ)],
                        rows: &[
                            &[T("1921-01-01 00:00:00+00")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT to_date('2021-W35-7', 'IYYY-"W"IW-ID');"#,
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("2021-09-05")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('Sunday 5th September 2021', 'Day DDth Month YYYY');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("2021-09-05")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('XII 2021 1', 'RM YYYY DD');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("2021-12-01")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_date('', '');",
                    expected: Expected::Rows {
                        columns: &[Column("to_date", DATE)],
                        rows: &[
                            &[T("0001-01-01 BC")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "to_timestamp errors",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021-13-01', 'YYYY-MM-DD');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"date/time field value out of range: "2021-13-01""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('13:00 PM', 'HH12:MI AM');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"hour "13" is invalid for the 12-hour clock"#, hint: "Use the 24-hour clock, or give an hour between 1 and 12.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021 2022', 'YYYY YYYY');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"conflicting values for "YYYY" field in formatting string"#, detail: "This value contradicts a previous setting for the same field type.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('202', 'YYYYMM');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"source string too short for "YYYY" formatting field"#, detail: "Field requires 4 characters, but only 3 remain.", hint: r#"If your source string is not fixed-width, try using the "FM" modifier."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021-xx-01', 'YYYY-MM-DD');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid value "xx" for "MM""#, detail: "Value must be an integer.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021 Foo', 'YYYY Mon');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: r#"invalid value "Foo" for "Mon""#, detail: "The given value did not match any of the allowed values for this field.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021-05 3', 'YYYY-MM IW');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: "invalid combination of date conventions", hint: "Do not mix Gregorian and ISO week date conventions in a formatting template.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('12 TZ', 'HH24 TZ');",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"formatting field "TZ" is only supported in to_char"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('2021-01-01 +20', 'YYYY-MM-DD TZH');",
                    expected: Expected::Error(Diagnostic { code: "22009", message: r#"time zone displacement out of range: "2021-01-01 +20""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('99999999999', 'YYYY');",
                    expected: Expected::Error(Diagnostic { code: "22008", message: r#"value for "YYYY" in source string is out of range"#, detail: "Value must be in the range -2147483648 to 2147483647.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_timestamp('15', 'DDD');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: "cannot calculate day of year without year information", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(INTERVAL '1 day', 'TZ');",
                    expected: Expected::Error(Diagnostic { code: "22007", message: "invalid format specification for an interval value", hint: "Intervals are not tied to specific calendar dates.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_char(TIMESTAMP '2021-01-01', 'FMDDTH') ;",
                    expected: Expected::Rows {
                        columns: &[Column("to_char", TEXT)],
                        rows: &[
                            &[T("1ST")],
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
fn test_builtin_functions() {
    run_scripts(&[
        ScriptTest {
            name: "trigonometric functions in degrees and inverse and hyperbolic functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT sind(30), sind(-30), sind(90), sind(180), sind(210), sind(270), sind(360), sind(45);",
                    expected: Expected::Rows {
                        columns: &[Column("sind", FLOAT8), Column("sind", FLOAT8), Column("sind", FLOAT8), Column("sind", FLOAT8), Column("sind", FLOAT8), Column("sind", FLOAT8), Column("sind", FLOAT8), Column("sind", FLOAT8)],
                        rows: &[
                            &[T("0.5"), T("-0.5"), T("1"), T("0"), T("-0.5"), T("-1"), T("0"), T("0.7071067811865475")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosd(60), cosd(90), cosd(120), cosd(180), cosd(-60), cosd(300), cosd(45);",
                    expected: Expected::Rows {
                        columns: &[Column("cosd", FLOAT8), Column("cosd", FLOAT8), Column("cosd", FLOAT8), Column("cosd", FLOAT8), Column("cosd", FLOAT8), Column("cosd", FLOAT8), Column("cosd", FLOAT8)],
                        rows: &[
                            &[T("0.5"), T("0"), T("-0.5"), T("-1"), T("0.5"), T("0.5"), T("0.7071067811865475")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT tand(45), tand(90), tand(135), tand(180), tand(-45), tand(225), tand(30);",
                    expected: Expected::Rows {
                        columns: &[Column("tand", FLOAT8), Column("tand", FLOAT8), Column("tand", FLOAT8), Column("tand", FLOAT8), Column("tand", FLOAT8), Column("tand", FLOAT8), Column("tand", FLOAT8)],
                        rows: &[
                            &[T("1"), T("Infinity"), T("-1"), T("0"), T("-1"), T("1"), T("0.5773502691896257")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cotd(45), cotd(90), cotd(0), cotd(135), cotd(-45);",
                    expected: Expected::Rows {
                        columns: &[Column("cotd", FLOAT8), Column("cotd", FLOAT8), Column("cotd", FLOAT8), Column("cotd", FLOAT8), Column("cotd", FLOAT8)],
                        rows: &[
                            &[T("1"), T("0"), T("Infinity"), T("-1"), T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT asind(0.5), asind(-0.5), asind(1), acosd(0.5), acosd(-0.5), acosd(1), acosd(-1), atand(1), atand(-1), atan2d(1, 1), atan2d(1, -1);",
                    expected: Expected::Rows {
                        columns: &[Column("asind", FLOAT8), Column("asind", FLOAT8), Column("asind", FLOAT8), Column("acosd", FLOAT8), Column("acosd", FLOAT8), Column("acosd", FLOAT8), Column("acosd", FLOAT8), Column("atand", FLOAT8), Column("atand", FLOAT8), Column("atan2d", FLOAT8), Column("atan2d", FLOAT8)],
                        rows: &[
                            &[T("30"), T("-30"), T("90"), T("60"), T("120"), T("0"), T("180"), T("45"), T("-45"), T("45"), T("135")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                // Postgres computes acos, asin, cot, and atanh with the platform's libm, whose last digit differs between macOS and glibc.
                ScriptTestAssertion {
                    query: "SELECT acos(0.5), asin(0.5), atan(1), atan2(1, 2), cot(1), sinh(1), cosh(1), tanh(1), asinh(1), acosh(2), atanh(0.5);",
                    expected: Expected::Rows {
                        columns: &[Column("acos", FLOAT8), Column("asin", FLOAT8), Column("atan", FLOAT8), Column("atan2", FLOAT8), Column("cot", FLOAT8), Column("sinh", FLOAT8), Column("cosh", FLOAT8), Column("tanh", FLOAT8), Column("asinh", FLOAT8), Column("acosh", FLOAT8), Column("atanh", FLOAT8)],
                        rows: &[
                            &[Approx("1.0471975511965976"), Approx("0.5235987755982988"), T("0.7853981633974483"), T("0.4636476090008061"), Approx("0.6420926159343308"), T("1.1752011936438014"), T("1.5430806348152437"), T("0.7615941559557649"), T("0.881373587019543"), T("1.3169578969248166"), Approx("0.5493061443340549")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sind('NaN'), acos('NaN'), atanh('NaN'), atanh(1), atanh(-1), cosh('Infinity'), sinh('-Infinity');",
                    expected: Expected::Rows {
                        columns: &[Column("sind", FLOAT8), Column("acos", FLOAT8), Column("atanh", FLOAT8), Column("atanh", FLOAT8), Column("atanh", FLOAT8), Column("cosh", FLOAT8), Column("sinh", FLOAT8)],
                        rows: &[
                            &[T("NaN"), T("NaN"), T("NaN"), T("Infinity"), T("-Infinity"), T("Infinity"), T("-Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sind('Infinity');",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "input is out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT acos(2);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "input is out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT acosh(0.5);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "input is out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT atanh(2);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "input is out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cot(0);",
                    expected: Expected::Rows {
                        columns: &[Column("cot", FLOAT8)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "factorial, scale, trim_scale, width_bucket, and to_hex",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT factorial(0), factorial(1), factorial(5), factorial(20), factorial(25);",
                    expected: Expected::Rows {
                        columns: &[Column("factorial", NUMERIC), Column("factorial", NUMERIC), Column("factorial", NUMERIC), Column("factorial", NUMERIC), Column("factorial", NUMERIC)],
                        rows: &[
                            &[T("1"), T("1"), T("120"), T("2432902008176640000"), T("15511210043330985984000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT factorial(-1);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "factorial of a negative number is undefined", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT factorial(40000);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value overflows numeric format", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT scale(1.230), scale(5), scale('NaN'::numeric), scale('Infinity'::numeric), trim_scale(1.2300), trim_scale(10.000), trim_scale('NaN'::numeric), trim_scale(-0.00);",
                    expected: Expected::Rows {
                        columns: &[Column("scale", INT4), Column("scale", INT4), Column("scale", INT4), Column("scale", INT4), Column("trim_scale", NUMERIC), Column("trim_scale", NUMERIC), Column("trim_scale", NUMERIC), Column("trim_scale", NUMERIC)],
                        rows: &[
                            &[T("3"), T("0"), Null, Null, T("1.23"), T("10"), T("NaN"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT width_bucket(5.35, 0.024, 10.06, 5), width_bucket(5.35::float8, 0.024, 10.06, 5), width_bucket(-1, 0, 10, 5), width_bucket(10, 0, 10, 5), width_bucket(5, 10, 0, 5), width_bucket(0, 10, 0, 5), width_bucket(11, 10, 0, 5), width_bucket(9.999999999, 0, 10, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("width_bucket", INT4), Column("width_bucket", INT4), Column("width_bucket", INT4), Column("width_bucket", INT4), Column("width_bucket", INT4), Column("width_bucket", INT4), Column("width_bucket", INT4), Column("width_bucket", INT4)],
                        rows: &[
                            &[T("3"), T("3"), T("0"), T("6"), T("3"), T("6"), T("0"), T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT width_bucket(5, 0, 10, 0);",
                    expected: Expected::Error(Diagnostic { code: "2201G", message: "count must be greater than zero", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT width_bucket('NaN'::float8, 0, 10, 5);",
                    expected: Expected::Error(Diagnostic { code: "2201G", message: "operand, lower bound, and upper bound cannot be NaN", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT width_bucket(5::float8, '-Infinity', 10, 5);",
                    expected: Expected::Error(Diagnostic { code: "2201G", message: "lower and upper bounds must be finite", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT width_bucket(5, 5, 5, 5);",
                    expected: Expected::Error(Diagnostic { code: "2201G", message: "lower bound cannot equal upper bound", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT width_bucket(5::numeric, 'Infinity', 10, 5);",
                    expected: Expected::Error(Diagnostic { code: "2201G", message: "lower and upper bounds must be finite", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT width_bucket(10, 0, 10, 2147483647);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "integer out of range", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT width_bucket('Infinity'::numeric, 0, 10, 5), width_bucket('-Infinity'::float8, 0, 10, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("width_bucket", INT4), Column("width_bucket", INT4)],
                        rows: &[
                            &[T("6"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_hex(255), to_hex(-1), to_hex(-1::bigint), to_hex(0), to_hex(9223372036854775807);",
                    expected: Expected::Rows {
                        columns: &[Column("to_hex", TEXT), Column("to_hex", TEXT), Column("to_hex", TEXT), Column("to_hex", TEXT), Column("to_hex", TEXT)],
                        rows: &[
                            &[T("ff"), T("ffffffff"), T("ffffffffffffffff"), T("0"), T("7fffffffffffffff")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "server status and relation functions",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT random() >= 0 AND random() < 1;",
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
                    query: "SELECT pg_sleep(0.01), pg_sleep_for('10 milliseconds'), pg_sleep(-1);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_sleep", VOID), Column("pg_sleep_for", VOID), Column("pg_sleep", VOID)],
                        rows: &[
                            &[T(""), T(""), T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(pg_backend_pid()), pg_typeof(txid_current()), pg_typeof(pg_postmaster_start_time()), pg_postmaster_start_time() <= now(), pg_is_in_recovery();",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE), Column("pg_typeof", REGTYPE), Column("?column?", BOOL), Column("pg_is_in_recovery", BOOL)],
                        rows: &[
                            &[T("integer"), T("bigint"), T("timestamp with time zone"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_is_wal_replay_paused();",
                    expected: Expected::Error(Diagnostic { code: "55000", message: "recovery is not in progress", hint: "Recovery control functions can only be executed during recovery.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE sz (a int primary key);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_relation_size('sz'), pg_table_size('sz'), pg_relation_size('sz', 'main') >= 0, pg_relation_size(1::oid::regclass);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_relation_size", INT8), Column("pg_table_size", INT8), Column("?column?", BOOL), Column("pg_relation_size", INT8)],
                        rows: &[
                            &[T("0"), T("0"), T("t"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_relation_size('sz', 'bogus');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "invalid fork name", hint: r#"Valid fork names are "main", "fsm", "vm", and "init"."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_relation_is_publishable('sz'), pg_relation_is_publishable('pg_class'), pg_relation_is_publishable(1::oid::regclass), pg_get_partkeydef('sz'::regclass), pg_tablespace_location(1663), pg_stat_get_numscans(1);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_relation_is_publishable", BOOL), Column("pg_relation_is_publishable", BOOL), Column("pg_relation_is_publishable", BOOL), Column("pg_get_partkeydef", TEXT), Column("pg_tablespace_location", TEXT), Column("pg_stat_get_numscans", INT8)],
                        rows: &[
                            &[T("t"), T("f"), Null, Null, T(""), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "has_schema_privilege and has_database_privilege",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT has_schema_privilege('public', 'USAGE'), has_schema_privilege('public', 'CREATE'), has_schema_privilege('postgres', 'public', 'usage, create'), has_schema_privilege('public', 'USAGE WITH GRANT OPTION'), has_schema_privilege(1::oid, 'USAGE'), has_schema_privilege('pg_catalog', ' usage ');",
                    expected: Expected::Rows {
                        columns: &[Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("t"), Null, T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_schema_privilege('nosuch', 'USAGE');",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "nosuch" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_schema_privilege('nobody', 'public', 'USAGE');",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"role "nobody" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_schema_privilege('public', 'SELECT');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized privilege type: "SELECT""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_schema_privilege(1::oid, 'SELECT');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized privilege type: "SELECT""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_database_privilege('postgres', 'CONNECT'), has_database_privilege('postgres', 'temp, create'), has_database_privilege('template1', 'CONNECT'), has_database_privilege(1::oid, 'CONNECT');",
                    expected: Expected::Rows {
                        columns: &[Column("has_database_privilege", BOOL), Column("has_database_privilege", BOOL), Column("has_database_privilege", BOOL), Column("has_database_privilege", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_database_privilege('nosuch', 'CONNECT');",
                    expected: Expected::Error(Diagnostic { code: "3D000", message: r#"database "nosuch" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_database_privilege('postgres', 'USAGE');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized privilege type: "USAGE""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE ROLE limited;",
                    expected: Expected::Tag("CREATE ROLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_schema_privilege('limited', 'public', 'USAGE'), has_schema_privilege('limited', 'public', 'CREATE'), has_schema_privilege('limited', 'pg_catalog', 'USAGE'), has_database_privilege('limited', 'postgres', 'CONNECT'), has_database_privilege('limited', 'postgres', 'CREATE'), has_database_privilege('public', 'postgres', 'TEMP');",
                    expected: Expected::Rows {
                        columns: &[Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL), Column("has_database_privilege", BOOL), Column("has_database_privilege", BOOL), Column("has_database_privilege", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE SCHEMA s2;",
                    expected: Expected::Tag("CREATE SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "GRANT USAGE ON SCHEMA s2 TO limited WITH GRANT OPTION;",
                    expected: Expected::Tag("GRANT"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT has_schema_privilege('limited', 's2', 'USAGE WITH GRANT OPTION'), has_schema_privilege('limited', 's2', 'CREATE'), has_schema_privilege('limited', 's2', 'USAGE');",
                    expected: Expected::Rows {
                        columns: &[Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL), Column("has_schema_privilege", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP SCHEMA s2 CASCADE;",
                    expected: Expected::Tag("DROP SCHEMA"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP ROLE limited;",
                    expected: Expected::Tag("DROP ROLE"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_definition_functions() {
    run_scripts(&[
        ScriptTest {
            name: "pg_get_functiondef and the pg_get_function family",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE cp_test (a int, b text);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION f1(int) RETURNS int LANGUAGE sql AS 'SELECT $1 + 1';",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION f2(a int, b text DEFAULT 'x', c numeric DEFAULT 1, d varchar(10) DEFAULT 'y') RETURNS SETOF text IMMUTABLE STRICT LANGUAGE sql AS $$SELECT b$$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION f3(IN a int, OUT b int, INOUT c text) LANGUAGE plpgsql AS $$BEGIN b := a; END$$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION f4(VARIADIC xs int[]) RETURNS TABLE(n int, t text) STABLE PARALLEL SAFE COST 5 ROWS 10 LANGUAGE sql AS $$SELECT 1, 'a'$$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION f5() RETURNS SETOF cp_test SECURITY DEFINER LEAKPROOF SET search_path = public, pg_temp SET work_mem = '64MB' LANGUAGE sql AS $$SELECT * FROM cp_test$$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION f6(a int, b int) RETURNS int RETURN a + b;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION f7(x int) RETURNS int LANGUAGE plpgsql AS $function$BEGIN RETURN x; END$function$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE p1(a int, b text, c int default 100) LANGUAGE sql AS $$INSERT INTO cp_test VALUES (a, b)$$;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE PROCEDURE p2(INOUT a int, OUT b text) LANGUAGE plpgsql AS $$BEGIN b := 'x'; END$$;",
                    expected: Expected::Tag("CREATE PROCEDURE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT proname, pg_get_function_arguments(oid) AS args, pg_get_function_identity_arguments(oid) AS ident, pg_get_function_result(oid) AS res FROM pg_proc WHERE pronamespace = 'public'::regnamespace ORDER BY proname;",
                    expected: Expected::Rows {
                        columns: &[Column("proname", NAME), Column("args", TEXT), Column("ident", TEXT), Column("res", TEXT)],
                        rows: &[
                            &[T("f1"), T("integer"), T("integer"), T("integer")],
                            &[T("f2"), T("a integer, b text DEFAULT 'x'::text, c numeric DEFAULT 1, d character varying DEFAULT 'y'::character varying"), T("a integer, b text, c numeric, d character varying"), T("SETOF text")],
                            &[T("f3"), T("a integer, OUT b integer, INOUT c text"), T("a integer, OUT b integer, INOUT c text"), T("record")],
                            &[T("f4"), T("VARIADIC xs integer[]"), T("VARIADIC xs integer[]"), T("TABLE(n integer, t text)")],
                            &[T("f5"), T(""), T(""), T("SETOF cp_test")],
                            &[T("f6"), T("a integer, b integer"), T("a integer, b integer"), T("integer")],
                            &[T("f7"), T("x integer"), T("x integer"), T("integer")],
                            &[T("p1"), T("IN a integer, IN b text, IN c integer DEFAULT 100"), T("IN a integer, IN b text, IN c integer"), Null],
                            &[T("p2"), T("INOUT a integer, OUT b text"), T("INOUT a integer, OUT b text"), Null],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_functiondef(oid) FROM pg_proc WHERE pronamespace = 'public'::regnamespace ORDER BY proname;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_functiondef", TEXT)],
                        rows: &[
                            &[T(r#"CREATE OR REPLACE FUNCTION public.f1(integer)
 RETURNS integer
 LANGUAGE sql
AS $function$SELECT $1 + 1$function$
"#)],
                            &[T(r#"CREATE OR REPLACE FUNCTION public.f2(a integer, b text DEFAULT 'x'::text, c numeric DEFAULT 1, d character varying DEFAULT 'y'::character varying)
 RETURNS SETOF text
 LANGUAGE sql
 IMMUTABLE STRICT
AS $function$SELECT b$function$
"#)],
                            &[T(r#"CREATE OR REPLACE FUNCTION public.f3(a integer, OUT b integer, INOUT c text)
 RETURNS record
 LANGUAGE plpgsql
AS $function$BEGIN b := a; END$function$
"#)],
                            &[T(r#"CREATE OR REPLACE FUNCTION public.f4(VARIADIC xs integer[])
 RETURNS TABLE(n integer, t text)
 LANGUAGE sql
 STABLE PARALLEL SAFE COST 5 ROWS 10
AS $function$SELECT 1, 'a'$function$
"#)],
                            &[T(r#"CREATE OR REPLACE FUNCTION public.f5()
 RETURNS SETOF cp_test
 LANGUAGE sql
 SECURITY DEFINER LEAKPROOF
 SET search_path TO 'public', 'pg_temp'
 SET work_mem TO '64MB'
AS $function$SELECT * FROM cp_test$function$
"#)],
                            &[T(r#"CREATE OR REPLACE FUNCTION public.f6(a integer, b integer)
 RETURNS integer
 LANGUAGE sql
RETURN (a + b)
"#)],
                            &[T(r#"CREATE OR REPLACE FUNCTION public.f7(x integer)
 RETURNS integer
 LANGUAGE plpgsql
AS $function$BEGIN RETURN x; END$function$
"#)],
                            &[T(r#"CREATE OR REPLACE PROCEDURE public.p1(IN a integer, IN b text, IN c integer DEFAULT 100)
 LANGUAGE sql
AS $procedure$INSERT INTO cp_test VALUES (a, b)$procedure$
"#)],
                            &[T(r#"CREATE OR REPLACE PROCEDURE public.p2(INOUT a integer, OUT b text)
 LANGUAGE plpgsql
AS $procedure$BEGIN b := 'x'; END$procedure$
"#)],
                        ],
                        tag: "SELECT 9",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_functiondef('abs'::regproc), pg_get_function_arguments('make_interval'::regproc), pg_get_function_result('generate_series'::regproc), pg_get_function_arguments('generate_series'::regproc), pg_get_functiondef('jsonb_path_exists'::regproc);",
                    expected: Expected::Error(Diagnostic { code: "42725", message: r#"more than one function named "abs""#, position: 27, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_functiondef('pg_get_keywords'::regproc), pg_get_function_result('pg_get_keywords'::regproc);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_functiondef", TEXT), Column("pg_get_function_result", TEXT)],
                        rows: &[
                            &[T(r#"CREATE OR REPLACE FUNCTION pg_catalog.pg_get_keywords(OUT word text, OUT catcode "char", OUT barelabel boolean, OUT catdesc text, OUT baredesc text)
 RETURNS SETOF record
 LANGUAGE internal
 STABLE PARALLEL SAFE STRICT COST 10 ROWS 500
AS $function$pg_get_keywords$function$
"#), T("SETOF record")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_function_arguments(0), pg_get_function_sqlbody('f6'::regproc), pg_get_function_sqlbody('f1'::regproc);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_function_arguments", TEXT), Column("pg_get_function_sqlbody", TEXT), Column("pg_get_function_sqlbody", TEXT)],
                        rows: &[
                            &[Null, T("RETURN (a + b)"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_functiondef('sum'::regproc);",
                    expected: Expected::Error(Diagnostic { code: "42725", message: r#"more than one function named "sum""#, position: 27, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pg_get_triggerdef",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE TABLE tt (a int, b text, "C d" int);"#,
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION trig_fn() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN RETURN NEW; END$$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER t1 BEFORE INSERT ON tt FOR EACH ROW EXECUTE FUNCTION trig_fn();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER t2 AFTER UPDATE OF b, a OR DELETE OR INSERT ON tt FOR EACH ROW EXECUTE FUNCTION trig_fn('x', 'it''s');",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER t3 AFTER TRUNCATE ON tt FOR EACH STATEMENT EXECUTE PROCEDURE trig_fn();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"CREATE TRIGGER t5 BEFORE UPDATE ON tt FOR EACH ROW WHEN (new.a > 0 AND new."C d" < 5) EXECUTE FUNCTION trig_fn();"#,
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER t6 BEFORE UPDATE ON tt FOR EACH ROW WHEN (old.* IS DISTINCT FROM new.*) EXECUTE FUNCTION trig_fn();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT tgname, pg_get_triggerdef(oid), pg_get_triggerdef(oid, true) FROM pg_trigger WHERE tgrelid = 'tt'::regclass ORDER BY tgname;",
                    expected: Expected::Rows {
                        columns: &[Column("tgname", NAME), Column("pg_get_triggerdef", TEXT), Column("pg_get_triggerdef", TEXT)],
                        rows: &[
                            &[T("t1"), T("CREATE TRIGGER t1 BEFORE INSERT ON public.tt FOR EACH ROW EXECUTE FUNCTION trig_fn()"), T("CREATE TRIGGER t1 BEFORE INSERT ON tt FOR EACH ROW EXECUTE FUNCTION trig_fn()")],
                            &[T("t2"), T("CREATE TRIGGER t2 AFTER INSERT OR DELETE OR UPDATE OF b, a ON public.tt FOR EACH ROW EXECUTE FUNCTION trig_fn('x', 'it''s')"), T("CREATE TRIGGER t2 AFTER INSERT OR DELETE OR UPDATE OF b, a ON tt FOR EACH ROW EXECUTE FUNCTION trig_fn('x', 'it''s')")],
                            &[T("t3"), T("CREATE TRIGGER t3 AFTER TRUNCATE ON public.tt FOR EACH STATEMENT EXECUTE FUNCTION trig_fn()"), T("CREATE TRIGGER t3 AFTER TRUNCATE ON tt FOR EACH STATEMENT EXECUTE FUNCTION trig_fn()")],
                            &[T("t5"), T(r#"CREATE TRIGGER t5 BEFORE UPDATE ON public.tt FOR EACH ROW WHEN (((new.a > 0) AND (new."C d" < 5))) EXECUTE FUNCTION trig_fn()"#), T(r#"CREATE TRIGGER t5 BEFORE UPDATE ON tt FOR EACH ROW WHEN (new.a > 0 AND new."C d" < 5) EXECUTE FUNCTION trig_fn()"#)],
                            &[T("t6"), T("CREATE TRIGGER t6 BEFORE UPDATE ON public.tt FOR EACH ROW WHEN ((old.* IS DISTINCT FROM new.*)) EXECUTE FUNCTION trig_fn()"), T("CREATE TRIGGER t6 BEFORE UPDATE ON tt FOR EACH ROW WHEN (old.* IS DISTINCT FROM new.*) EXECUTE FUNCTION trig_fn()")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"ALTER TABLE tt RENAME COLUMN b TO "B2";"#,
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE tgname = 't2';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_triggerdef", TEXT)],
                        rows: &[
                            &[T(r#"CREATE TRIGGER t2 AFTER INSERT OR DELETE OR UPDATE OF "B2", a ON public.tt FOR EACH ROW EXECUTE FUNCTION trig_fn('x', 'it''s')"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_triggerdef(0);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_triggerdef", TEXT)],
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
            name: "pg_get_triggerdef with transition tables",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE tt (a int);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION trig_fn() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN RETURN NULL; END$$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER t4 AFTER INSERT ON tt REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION trig_fn();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    skip: Some("trigger transition tables (REFERENCING) are not supported yet"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_triggerdef(oid, true) FROM pg_trigger WHERE tgname = 't4';",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_triggerdef", TEXT)],
                        rows: &[
                            &[T("CREATE TRIGGER t4 AFTER INSERT ON tt REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION trig_fn()")],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("trigger transition tables (REFERENCING) are not supported yet"),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_numeric_math() {
    run_scripts(&[
        ScriptTest {
            name: "numeric logarithms, exponentials, powers, and square roots",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ln(2::numeric), ln(0.000123::numeric), ln(1e-20::numeric), ln(1::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("ln", NUMERIC), Column("ln", NUMERIC), Column("ln", NUMERIC), Column("ln", NUMERIC)],
                        rows: &[
                            &[T("0.6931471805599453"), T("-9.0033262025918566"), T("-46.05170185988091368036"), T("0.0000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT log(1000::numeric), log(2::numeric, 8::numeric), log(7::numeric, 49.000001::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("log", NUMERIC), Column("log", NUMERIC), Column("log", NUMERIC)],
                        rows: &[
                            &[T("3.0000000000000000"), T("3.0000000000000000"), T("2.0000000104877212")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sqrt(2::numeric), sqrt(0.00001::numeric), sqrt(99999::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("sqrt", NUMERIC), Column("sqrt", NUMERIC), Column("sqrt", NUMERIC)],
                        rows: &[
                            &[T("1.414213562373095"), T("0.0031622776601683793"), T("316.2261848740550")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT exp(0::numeric), exp(1::numeric), exp(-2.5::numeric), exp(23.456::numeric), exp(-6000::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("exp", NUMERIC), Column("exp", NUMERIC), Column("exp", NUMERIC), Column("exp", NUMERIC), Column("exp", NUMERIC)],
                        rows: &[
                            &[T("1.0000000000000000"), T("2.7182818284590452"), T("0.08208499862389880"), T("15374866997.000768"), T("0.0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(2::numeric, 0.5::numeric), power(3::numeric, -3::numeric), power(-2::numeric, 3::numeric), power(12.34::numeric, 33::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC), Column("power", NUMERIC), Column("power", NUMERIC), Column("power", NUMERIC)],
                        rows: &[
                            &[T("1.4142135623730950"), T("0.0370370370370370"), T("-8.0000000000000000"), T("1031336219503664876006972321373516886.8670741364033477")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(0.001::numeric, -0.333::numeric), power(1.5::numeric, 100::numeric), power(0::numeric, 2.5::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC), Column("power", NUMERIC), Column("power", NUMERIC)],
                        rows: &[
                            &[T("9.9770006382255332"), T("406561177535215237.3972797075670417"), T("0.0000000000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power('-inf'::numeric, '-inf'::numeric), power('nan'::numeric, 0::numeric), power(1::numeric, 'nan'::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("power", NUMERIC), Column("power", NUMERIC), Column("power", NUMERIC)],
                        rows: &[
                            &[T("0"), T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ln(0::numeric);",
                    expected: Expected::Error(Diagnostic { code: "2201E", message: "cannot take logarithm of zero", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ln(-1::numeric);",
                    expected: Expected::Error(Diagnostic { code: "2201E", message: "cannot take logarithm of a negative number", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT log(1::numeric, 10::numeric);",
                    expected: Expected::Error(Diagnostic { code: "22012", message: "division by zero", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(0::numeric, -1::numeric);",
                    expected: Expected::Error(Diagnostic { code: "2201F", message: "zero raised to a negative power is undefined", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT power(-2::numeric, 0.5::numeric);",
                    expected: Expected::Error(Diagnostic { code: "2201F", message: "a negative number raised to a non-integer power yields a complex result", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT exp(6000::numeric);",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value overflows numeric format", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sqrt(-1::numeric);",
                    expected: Expected::Error(Diagnostic { code: "2201F", message: "cannot take square root of a negative number", ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_catalog_encoding_and_ordering_rules() {
    run_scripts(&[
        ScriptTest {
            name: "Catalog, encoding, ordering, and dependency rules",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\x41a4'::BYTEA, 'EUC_JP');"#,
                    expected: Expected::Error(Diagnostic { code: "22021", message: r#"invalid byte sequence for encoding "EUC_JP": 0xa4"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\xa4a2ff'::BYTEA, 'EUC_JP');"#,
                    expected: Expected::Error(Diagnostic { code: "22021", message: r#"invalid byte sequence for encoding "EUC_JP": 0xff"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT convert_from('\xa4a241'::BYTEA, 'EUC_JP');"#,
                    expected: Expected::Rows {
                        columns: &[Column("convert_from", TEXT)],
                        rows: &[
                            &[T("あA")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name FROM pg_show_all_settings() WHERE name LIKE 'bytea%';",
                    expected: Expected::Rows {
                        columns: &[Column("name", TEXT)],
                        rows: &[
                            &[T("bytea_output")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT set_config('bytea_output','hex',false) FROM pg_show_all_settings() WHERE name = 'bytea_output';",
                    expected: Expected::Rows {
                        columns: &[Column("set_config", TEXT)],
                        rows: &[
                            &[T("hex")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE gen (a int, s text, b int GENERATED ALWAYS AS (a + 1) STORED, c text GENERATED ALWAYS AS (upper(s)) STORED);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT column_name, is_generated, generation_expression FROM information_schema.columns WHERE table_name = 'gen' ORDER BY ordinal_position;",
                    expected: Expected::Rows {
                        columns: &[Column("column_name", NAME), Column("is_generated", VARCHAR), Column("generation_expression", VARCHAR)],
                        rows: &[
                            &[T("a"), T("NEVER"), Null],
                            &[T("s"), T("NEVER"), Null],
                            &[T("b"), T("ALWAYS"), T("(a + 1)")],
                            &[T("c"), T("ALWAYS"), T("upper(s)")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE test (id INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO test VALUES (1), (3), (2);",
                    expected: Expected::Tag("INSERT 0 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT *, (SELECT id from test where id = 2) FROM test order by id;",
                    expected: Expected::Error(Diagnostic { code: "42702", message: r#"ORDER BY "id" is ambiguous"#, position: 65, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, id FROM test ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("id", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("2")],
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 OPERATOR(myschema.+) 1;",
                    expected: Expected::Error(Diagnostic { code: "3F000", message: r#"schema "myschema" does not exist"#, position: 10, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 1 OPERATOR(pg_catalog.+) 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE agg (pk INT PRIMARY KEY, v INT[]);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO agg VALUES (1, ARRAY[1,2]), (2, ARRAY[3,4]), (3, ARRAY[5]), (4, NULL), (5, '{}');",
                    expected: Expected::Tag("INSERT 0 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk IN (1, 5);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "cannot accumulate arrays of different dimensionality", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk IN (5, 1);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "cannot accumulate arrays of different dimensionality", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(v ORDER BY pk) FROM agg WHERE pk IN (1, 4);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "cannot accumulate null arrays", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT v FROM agg ORDER BY pk);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "cannot accumulate arrays of different dimensionality", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY(SELECT v FROM agg WHERE pk < 3 ORDER BY pk);",
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
                    query: "SELECT ARRAY(SELECT v FROM agg WHERE false);",
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
                    query: "SELECT pg_typeof(ARRAY(SELECT v FROM agg WHERE false));",
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
                    query: "CREATE TABLE tt (id INT4 PRIMARY KEY, name TEXT, data TEXT, other TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION tf() RETURNS TRIGGER AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER tr1 BEFORE UPDATE OF name, data ON tt FOR EACH ROW EXECUTE FUNCTION tf();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE tt DROP COLUMN data;",
                    expected: Expected::Error(Diagnostic { code: "2BP01", message: "cannot drop column data of table tt because other objects depend on it", detail: "trigger tr1 on table tt depends on column data of table tt", hint: "Use DROP ... CASCADE to drop the dependent objects too.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE tt DROP COLUMN data CASCADE;",
                    expected: Expected::Tag("ALTER TABLE"),
                    notices: &[Diagnostic { code: "00000", message: "drop cascades to trigger tr1 on table tt", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT tgname FROM pg_trigger WHERE tgrelid = 'tt'::regclass;",
                    expected: Expected::Rows {
                        columns: &[Column("tgname", NAME)],
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
fn test_regex_feature_rules() {
    run_scripts(&[
        ScriptTest {
            name: "regular expression features and forms",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"select 'Programmer' ~ '(\w).*?\1' as t;"#,
                    expected: Expected::Rows {
                        columns: &[Column("t", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 'aa bb cc' ~ '(^(?!aa))+';",
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
                    query: "select 'foobar' ~ 'foo(?=bar)', 'foobaz' ~ 'foo(?=bar)', 'foobar' ~ '(?<=foo)bar', 'xbar' ~ '(?<!foo)bar';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select regexp_replace('aaa bbb', '(a)\1', 'X', 'g');"#,
                    expected: Expected::Rows {
                        columns: &[Column("regexp_replace", TEXT)],
                        rows: &[
                            &[T("Xa bbb")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_matches('abcabc', '(b)(c)', 'g');",
                    expected: Expected::Rows {
                        columns: &[Column("regexp_matches", TEXT_ARRAY)],
                        rows: &[
                            &[T("{b,c}")],
                            &[T("{b,c}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"select regexp_count('abcabc', 'b'), regexp_split_to_array('a1b2c', '\d'), substring('foobar' from 'o(b)a');"#,
                    expected: Expected::Rows {
                        columns: &[Column("regexp_count", INT4), Column("regexp_split_to_array", TEXT_ARRAY), Column("substring", TEXT)],
                        rows: &[
                            &[T("2"), T("{a,b,c}"), T("b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 'abc' ~ '(';",
                    expected: Expected::Error(Diagnostic { code: "2201B", message: "invalid regular expression: parentheses () not balanced", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 'abc' ~ '[';",
                    expected: Expected::Error(Diagnostic { code: "2201B", message: "invalid regular expression: brackets [] not balanced", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select 'abc' ~ '*';",
                    expected: Expected::Error(Diagnostic { code: "2201B", message: "invalid regular expression: quantifier operand invalid", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_replace('A PostgreSQL function', 'a|e|i|o|u', 'X', 1, 0, 'i');",
                    expected: Expected::Rows {
                        columns: &[Column("regexp_replace", TEXT)],
                        rows: &[
                            &[T("X PXstgrXSQL fXnctXXn")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_replace('abcabcabc', 'b', 'X', 2), regexp_replace('abcabcabc', 'b', 'X', 1, 2), regexp_replace('abcabcabc', 'B', 'X', 1, 0, 'i'), regexp_replace('abcabc', '^a', 'X', 2);",
                    expected: Expected::Rows {
                        columns: &[Column("regexp_replace", TEXT), Column("regexp_replace", TEXT), Column("regexp_replace", TEXT), Column("regexp_replace", TEXT)],
                        rows: &[
                            &[T("aXcabcabc"), T("abcaXcabc"), T("aXcaXcaXc"), T("abcabc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_count('abcabc', 'b', 3), regexp_count('ABC', 'b', 1, 'i');",
                    expected: Expected::Rows {
                        columns: &[Column("regexp_count", INT4), Column("regexp_count", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_instr('abcabc', 'b'), regexp_instr('abcabc', 'b', 1, 2), regexp_instr('abcabc', 'b', 1, 1, 1), regexp_instr('abcabc', '(b)(c)', 1, 1, 0, '', 2), regexp_instr('abc', 'z');",
                    expected: Expected::Rows {
                        columns: &[Column("regexp_instr", INT4), Column("regexp_instr", INT4), Column("regexp_instr", INT4), Column("regexp_instr", INT4), Column("regexp_instr", INT4)],
                        rows: &[
                            &[T("2"), T("5"), T("3"), T("3"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_substr('abcabc', 'b.'), regexp_substr('abcabc', 'b.', 1, 2), regexp_substr('abcabc', '(b)(c)', 1, 1, '', 2), regexp_substr('abc', 'z');",
                    expected: Expected::Rows {
                        columns: &[Column("regexp_substr", TEXT), Column("regexp_substr", TEXT), Column("regexp_substr", TEXT), Column("regexp_substr", TEXT)],
                        rows: &[
                            &[T("bc"), T("bc"), T("c"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_count('abc', 'b', 0);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid value for parameter "start": 0"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_instr('abc', 'b', 1, 1, 2);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid value for parameter "endoption": 2"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "select regexp_substr('abc', 'b', 1, 1, 'g');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"regexp_substr() does not support the "global" option"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_operator_implementation_functions() {
    run_scripts(&[
        ScriptTest {
            name: "operator implementation functions and named arguments",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT int4mi(5, 3), textcat('a', 'b'), int4um(4), float8mul(2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("int4mi", INT4), Column("textcat", TEXT), Column("int4um", INT4), Column("float8mul", FLOAT8)],
                        rows: &[
                            &[T("2"), T("ab"), T("-4"), T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT make_interval(hours := -2, mins := -10, secs := -25.3), make_interval(1, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("make_interval", INTERVAL), Column("make_interval", INTERVAL)],
                        rows: &[
                            &[T("-02:10:25.3"), T("1 year 2 mons")],
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
fn test_parse_ident_splits() {
    run_scripts(&[
        ScriptTest {
            name: "parse_ident splits qualified identifiers as Postgres does",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT parse_ident('Schema.TableName');",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{schema,tablename}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('"Schema"."Table Name"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{Schema,"Table Name"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident(' first . "Sec""ond" . third ');"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{first,"Sec\"ond",third}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('a$1.b_2');",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{a$1,b_2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('naïve.Ünïcode');",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{naïve,Ünïcode}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('foo.bar()', false);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{foo,bar}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('foo.bar()');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"string is not a valid identifier: "foo.bar()""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('"unclosed');"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"string is not a valid identifier: ""unclosed""#, detail: "String has unclosed double quotes.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('""');"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"string is not a valid identifier: """""#, detail: "Quoted identifier must not be empty.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('.foo');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"string is not a valid identifier: ".foo""#, detail: r#"No valid identifier before "."."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('foo.');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"string is not a valid identifier: "foo.""#, detail: r#"No valid identifier after "."."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('1abc');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"string is not a valid identifier: "1abc""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
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
fn test_quick_functions() {
    run_scripts(&[
        ScriptTest {
            name: "overlay, quote_nullable, unistr, and pg_size_pretty",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT overlay('abcdef' placing 'xy' from 2 for 3), overlay('abcdef' placing 'XYZ' from 3), overlay('abc' placing 'x' from 2 for -1), overlay('abc' placing 'zz' from 5);",
                    expected: Expected::Rows {
                        columns: &[Column("overlay", TEXT), Column("overlay", TEXT), Column("overlay", TEXT), Column("overlay", TEXT)],
                        rows: &[
                            &[T("axyef"), T("abXYZf"), T("axabc"), T("abczz")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT overlay('\x0102030405'::bytea placing '\xff'::bytea from 2), overlay('\x0102'::bytea placing '\xaabb'::bytea from 1 for 0);"#,
                    expected: Expected::Rows {
                        columns: &[Column("overlay", BYTEA), Column("overlay", BYTEA)],
                        rows: &[
                            &[T(r#"\x01ff030405"#), T(r#"\xaabb0102"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT overlay('abc' placing 'x' from 0);",
                    expected: Expected::Error(Diagnostic { code: "22011", message: "negative substring length not allowed", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT quote_nullable(NULL), quote_nullable('it''s'), quote_nullable(42), quote_nullable('back\slash');"#,
                    expected: Expected::Rows {
                        columns: &[Column("quote_nullable", TEXT), Column("quote_nullable", TEXT), Column("quote_nullable", TEXT), Column("quote_nullable", TEXT)],
                        rows: &[
                            &[T("NULL"), T("'it''s'"), T("'42'"), T(r#"E'back\\slash'"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT unistr('d\0061t\+000061'), unistr('é\U0001F600\\'), unistr('\D83D\DE00');"#,
                    expected: Expected::Rows {
                        columns: &[Column("unistr", TEXT), Column("unistr", TEXT), Column("unistr", TEXT)],
                        rows: &[
                            &[T("data"), T(r#"é😀\"#), T("😀")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT unistr('\xyz');"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: "invalid Unicode escape", hint: r#"Unicode escapes must be \XXXX, \+XXXXXX, \uXXXX, or \UXXXXXXXX."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT unistr('\0000');"#,
                    expected: Expected::Error(Diagnostic { code: "22023", message: "invalid Unicode code point: 0000", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT unistr('\D83D');"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: "invalid Unicode surrogate pair", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_size_pretty(0::bigint), pg_size_pretty(10239::bigint), pg_size_pretty(10240::bigint), pg_size_pretty(20971519::bigint), pg_size_pretty(-123456789::bigint), pg_size_pretty(9223372036854775807);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_size_pretty", TEXT), Column("pg_size_pretty", TEXT), Column("pg_size_pretty", TEXT), Column("pg_size_pretty", TEXT), Column("pg_size_pretty", TEXT), Column("pg_size_pretty", TEXT)],
                        rows: &[
                            &[T("0 bytes"), T("10239 bytes"), T("10 kB"), T("20 MB"), T("-118 MB"), T("8192 PB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_size_pretty(1234.5::numeric), pg_size_pretty(123456789012345678901234567890::numeric), pg_size_pretty(-5000000::numeric);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_size_pretty", TEXT), Column("pg_size_pretty", TEXT), Column("pg_size_pretty", TEXT)],
                        rows: &[
                            &[T("1234.5 bytes"), T("109651655766237 PB"), T("-4883 kB")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_database_size(current_database()) > 0, pg_database_size(0::oid);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("pg_database_size", INT8)],
                        rows: &[
                            &[T("t"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_database_size('no_such_database');",
                    expected: Expected::Error(Diagnostic { code: "3D000", message: r#"database "no_such_database" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "OVERLAPS and pg_trigger_depth",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (DATE '2020-01-01', DATE '2020-02-01') OVERLAPS (DATE '2020-01-15', DATE '2020-03-01');",
                    expected: Expected::Rows {
                        columns: &[Column("overlaps", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (TIMESTAMP '2020-01-01', INTERVAL '1 day') OVERLAPS (TIMESTAMP '2020-01-02', INTERVAL '1 day');",
                    expected: Expected::Rows {
                        columns: &[Column("overlaps", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (TIME '01:00', TIME '03:00') OVERLAPS (TIME '02:00', TIME '04:00'), (TIME '01:00', INTERVAL '1 hour') OVERLAPS (TIME '01:30', TIME '00:30');",
                    expected: Expected::Rows {
                        columns: &[Column("overlaps", BOOL), Column("overlaps", BOOL)],
                        rows: &[
                            &[T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (TIMESTAMPTZ '2020-01-01 00:00+00', NULL::timestamptz) OVERLAPS (TIMESTAMPTZ '2020-01-01 00:00+00', TIMESTAMPTZ '2020-01-02 00:00+00');",
                    expected: Expected::Rows {
                        columns: &[Column("overlaps", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (NULL::timestamp, NULL::timestamp) OVERLAPS (TIMESTAMP '2020-01-01', TIMESTAMP '2020-01-02');",
                    expected: Expected::Rows {
                        columns: &[Column("overlaps", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (TIMESTAMP '2020-01-05', NULL::timestamp) OVERLAPS (TIMESTAMP '2020-01-01', TIMESTAMP '2020-01-02');",
                    expected: Expected::Rows {
                        columns: &[Column("overlaps", BOOL)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (TIMESTAMP '2020-01-03', TIMESTAMP '2020-01-01') OVERLAPS (TIMESTAMP '2020-01-02', NULL::timestamp);",
                    expected: Expected::Rows {
                        columns: &[Column("overlaps", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_trigger_depth();",
                    expected: Expected::Rows {
                        columns: &[Column("pg_trigger_depth", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE depth_log (d INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE depth_t (x INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE FUNCTION depth_f() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN INSERT INTO depth_log VALUES (pg_trigger_depth()); RETURN NEW; END; $$;",
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TRIGGER depth_tr BEFORE INSERT ON depth_t FOR EACH ROW EXECUTE FUNCTION depth_f();",
                    expected: Expected::Tag("CREATE TRIGGER"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO depth_t VALUES (1);",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT d FROM depth_log;",
                    expected: Expected::Rows {
                        columns: &[Column("d", INT4)],
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

#[test]
fn test_parse_ident() {
    run_scripts(&[
        ScriptTest {
            name: "signatures and array result",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('public'), parse_ident('"SomeSchema".some_table');"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{public}"), T("{SomeSchema,some_table}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('Schema.Table'::text, true), parse_ident('Schema.Table'::text, false);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{schema,table}"), T("{schema,table}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT pg_catalog.parse_ident('public'), pg_catalog.parse_ident('"SomeSchema".some_table', false);"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{public}"), T("{SomeSchema,some_table}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT to_regprocedure('pg_catalog.parse_ident(text, boolean)')::oid;",
                    expected: Expected::Rows {
                        columns: &[Column("to_regprocedure", OID)],
                        rows: &[
                            &[T("1268")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('Schema.Table'::varchar), parse_ident('public'::name);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{schema,table}"), T("{public}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT (parse_ident('"SomeSchema".some_table'))[1], (parse_ident('"SomeSchema".some_table'))[2], (parse_ident('public'))[2];"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT), Column("parse_ident", TEXT), Column("parse_ident", TEXT)],
                        rows: &[
                            &[T("SomeSchema"), T("some_table"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(parse_ident('public')), array_length(parse_ident('a.b.c.d.e'), 1);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE), Column("array_length", INT4)],
                        rows: &[
                            &[T("text[]"), T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident($1::text, $2::boolean);",
                    bind_vars: &[BindVar::Str(r#""SomeSchema".SomeTable(integer)"#), BindVar::Bool(false)],
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{SomeSchema,sometable}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "quoted identifiers",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('"SomeSchema".someTable'), parse_ident('"SomeSchema"."SomeTable"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{SomeSchema,sometable}"), T("{SomeSchema,SomeTable}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT (parse_ident('"schema.with.dots"."table name"'))[1], (parse_ident('"schema.with.dots"."table name"'))[2];"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT), Column("parse_ident", TEXT)],
                        rows: &[
                            &[T("schema.with.dots"), T("table name")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT (parse_ident('"a""b"."c""d"'))[1], (parse_ident('"a""b"."c""d"'))[2];"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT), Column("parse_ident", TEXT)],
                        rows: &[
                            &[T(r#"a"b"#), T(r#"c"d"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT (parse_ident('"""foo"""'))[1], (parse_ident('""""'))[1];"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT), Column("parse_ident", TEXT)],
                        rows: &[
                            &[T(r#""foo""#), T(r#"""#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT (parse_ident('"  spaced  "'))[1], (parse_ident('"123"'))[1], (parse_ident('"char"'))[1];"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT), Column("parse_ident", TEXT), Column("parse_ident", TEXT)],
                        rows: &[
                            &[T("  spaced  "), T("123"), T("char")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT (parse_ident('"comma,brace{and}slash\"'))[1], (parse_ident('"O''Brien"'))[1];"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT), Column("parse_ident", TEXT)],
                        rows: &[
                            &[T(r#"comma,brace{and}slash\"#), T("O'Brien")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT (parse_ident(E'"line\nnext"'))[1];"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT)],
                        rows: &[
                            &[T(r#"line
next"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('"NULL"."a,b"');"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{"NULL","a,b"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "unquoted identifiers and whitespace",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT parse_ident('_Schema9.Table$1'), parse_ident('select.from');",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{_schema9,table$1}"), T("{select,from}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('  First . "  Second  " . Third  ');"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T(r#"{first,"  Second  ",third}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident(E' \t\n\r\fFirst\t.\nSecond\r\f ');"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{first,second}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('ÄBC.ÖDEF'), parse_ident('日本語.テーブル'), parse_ident('🐘.TABLE');",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{Äbc,Ödef}"), T("{日本語,テーブル}"), T("{🐘,table}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (parse_ident('E\u{301}COLE'))[1];",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT)],
                        rows: &[
                            &[T("e\u{301}cole")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT (parse_ident('\u{a0}FOO\u{a0}'))[1];",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT)],
                        rows: &[
                            &[T("\u{a0}foo\u{a0}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "non-strict trailing input",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT parse_ident('Schema.Function(integer, text)', false);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{schema,function}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('foo.boo[]', false), parse_ident('aaa.a%b', false);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{foo,boo}"), T("{aaa,a}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('foo bar.baz', false), parse_ident('foo;bar', false);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{foo}"), T("{foo}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('"Foo"bar', false), parse_ident('foo"bar', false);"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{Foo}"), T("{foo}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident('"Foo"."Bar" (integer)', false);"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{Foo,Bar}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT parse_ident(E'foo\013bar', false), parse_ident('foo/*comment*/.bar', false);"#,
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("{foo}"), T("{foo}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "long identifiers",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (parse_ident(repeat('A', 100)))[1];",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT)],
                        rows: &[
                            &[T("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT length((parse_ident('"' || repeat('X', 100) || '".' || repeat('Y', 100)))[1]), length((parse_ident('"' || repeat('X', 100) || '".' || repeat('Y', 100)))[2]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("100"), T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length((parse_ident(repeat('Ä', 100)))[1]), octet_length((parse_ident(repeat('Ä', 100)))[1]);",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("octet_length", INT4)],
                        rows: &[
                            &[T("100"), T("200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "null arguments",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT parse_ident(NULL), parse_ident(NULL::text);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident(NULL, true), parse_ident(NULL, false), parse_ident(NULL, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[Null, Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT parse_ident('public', NULL::boolean), parse_ident('invalid..name', NULL::boolean);",
                    expected: Expected::Rows {
                        columns: &[Column("parse_ident", TEXT_ARRAY), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "column arguments",
            set_up_script: &[
                "CREATE TABLE parse_ident_inputs (id integer PRIMARY KEY, input text, strict_mode boolean);",
                r#"INSERT INTO parse_ident_inputs VALUES (1, 'PUBLIC', true), (2, '"SomeSchema".SomeTable', false), (3, NULL, true), (4, 'invalid..name', NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, parse_ident(input, strict_mode) FROM parse_ident_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("1"), T("{public}")],
                            &[T("2"), T("{SomeSchema,sometable}")],
                            &[T("3"), Null],
                            &[T("4"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, parse_ident(input) FROM parse_ident_inputs WHERE id < 4 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("parse_ident", TEXT_ARRAY)],
                        rows: &[
                            &[T("1"), T("{public}")],
                            &[T("2"), T("{SomeSchema,sometable}")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "PostgREST computed relationship schema",
            set_up_script: &[
                "CREATE FUNCTION public.parse_ident_computed_rel(integer) RETURNS integer LANGUAGE SQL AS 'SELECT $1';",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT (parse_ident(p.pronamespace::regnamespace::text))[1] AS schema FROM pg_catalog.pg_proc p WHERE p.proname = 'parse_ident_computed_rel';",
                    expected: Expected::Rows {
                        columns: &[Column("schema", TEXT)],
                        rows: &[
                            &[T("public")],
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
