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
fn test_count_distinct_uuid() {
    run_scripts(&[
        ScriptTest {
            name: "COUNT DISTINCT uuid",
            set_up_script: &[
                "CREATE TABLE uuid_distinct (g text, u uuid, uuid_text text, n integer);",
                r#"INSERT INTO uuid_distinct VALUES
					('a','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001',1),
					('a','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001',1),
					('a','00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002',2),
					('a',NULL,NULL,NULL),
					('b','00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002',2),
					('b','00000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000003',3),
					('b',NULL,NULL,NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT COUNT(DISTINCT u) FROM uuid_distinct;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(DISTINCT u) FROM uuid_distinct WHERE u IS NULL;",
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
                    query: "SELECT COUNT(DISTINCT u) FROM uuid_distinct WHERE false;",
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
                    query: "SELECT g, COUNT(DISTINCT u) FROM uuid_distinct GROUP BY g ORDER BY g;",
                    expected: Expected::Rows {
                        columns: &[Column("g", TEXT), Column("count", INT8)],
                        rows: &[
                            &[T("a"), T("2")],
                            &[T("b"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(DISTINCT (u::text)::uuid), COUNT(DISTINCT uuid_text::uuid) FROM uuid_distinct;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("3"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(DISTINCT ('{' || uuid_text || '}')::uuid) FROM uuid_distinct;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT COUNT(DISTINCT u), COUNT(DISTINCT uuid_text), COUNT(DISTINCT n) FROM uuid_distinct;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("3"), T("3"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(COUNT(DISTINCT u)) FROM uuid_distinct;",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("bigint")],
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
fn test_distinct_aggregate_kinds() {
    run_scripts(&[
        ScriptTest {
            name: "DISTINCT aggregates",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE da (id INT PRIMARY KEY, n NUMERIC, f FLOAT8, t TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO da VALUES (1, 1.0, 0.0, 'a'), (2, 1.00, -0.0, 'b'), (3, 2, 'NaN', 'a'), (4, NULL, 'NaN', NULL), (5, 2.000, 1.5, 'b'), (6, 3, 1.5, 'c');",
                    expected: Expected::Tag("INSERT 0 6"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(DISTINCT n), count(DISTINCT f), count(DISTINCT t), count(DISTINCT id % 2) FROM da;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("count", INT8), Column("count", INT8), Column("count", INT8)],
                        rows: &[
                            &[T("3"), T("3"), T("3"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT string_agg(DISTINCT t, ',' ORDER BY t), array_agg(DISTINCT n ORDER BY n) FROM da;",
                    expected: Expected::Rows {
                        columns: &[Column("string_agg", TEXT), Column("array_agg", NUMERIC_ARRAY)],
                        rows: &[
                            &[T("a,b,c"), T("{1.0,2,3,NULL}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(DISTINCT n), avg(DISTINCT f) FROM da;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", NUMERIC), Column("avg", FLOAT8)],
                        rows: &[
                            &[T("6.0"), T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id % 2, count(DISTINCT n) FROM da GROUP BY id % 2 ORDER BY 1;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", INT4), Column("count", INT8)],
                        rows: &[
                            &[T("0"), T("2")],
                            &[T("1"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(DISTINCT (n, t)) FROM da;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
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
    ]);
}
