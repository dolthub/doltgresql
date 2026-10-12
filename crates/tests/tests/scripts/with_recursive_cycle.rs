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
fn test_with_cycle_generated_graphs() {
    run_scripts(&[
        ScriptTest {
            name: "generated CYCLE graph 01 seed 0000000000000001",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 5), (1, 5), (1, 4), (4, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 2, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("2"), T("0"), T("f"), T("{(2)}")],
                            &[T("3"), T("1"), T("f"), T("{(2),(3)}")],
                            &[T("1"), T("2"), T("f"), T("{(2),(3),(1)}")],
                            &[T("5"), T("2"), T("f"), T("{(2),(3),(5)}")],
                            &[T("2"), T("3"), T("t"), T("{(2),(3),(1),(2)}")],
                            &[T("4"), T("3"), T("f"), T("{(2),(3),(1),(4)}")],
                            &[T("5"), T("3"), T("f"), T("{(2),(3),(1),(5)}")],
                            &[T("5"), T("4"), T("f"), T("{(2),(3),(1),(4),(5)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 02 seed 0000000000000002",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 3), (3, 4), (2, 2), (2, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 3, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("3"), T("0"), T("f"), T("{(3)}")],
                            &[T("1"), T("1"), T("f"), T("{(3),(1)}")],
                            &[T("3"), T("1"), T("t"), T("{(3),(3)}")],
                            &[T("4"), T("1"), T("f"), T("{(3),(4)}")],
                            &[T("2"), T("2"), T("f"), T("{(3),(1),(2)}")],
                            &[T("1"), T("3"), T("t"), T("{(3),(1),(2),(1)}")],
                            &[T("2"), T("3"), T("t"), T("{(3),(1),(2),(2)}")],
                            &[T("3"), T("3"), T("t"), T("{(3),(1),(2),(3)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 03 seed 0000000000000003",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (2, 5), (5, 3), (4, 5), (1, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("5"), T("1"), T("f"), T("{(4),(5)}")],
                            &[T("3"), T("2"), T("f"), T("{(4),(5),(3)}")],
                            &[T("1"), T("3"), T("f"), T("{(4),(5),(3),(1)}")],
                            &[T("1"), T("4"), T("t"), T("{(4),(5),(3),(1),(1)}")],
                            &[T("2"), T("4"), T("f"), T("{(4),(5),(3),(1),(2)}")],
                            &[T("3"), T("5"), T("t"), T("{(4),(5),(3),(1),(2),(3)}")],
                            &[T("5"), T("5"), T("t"), T("{(4),(5),(3),(1),(2),(5)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 04 seed 0000000000000005",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (2, 5), (4, 1), (2, 1), (3, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("1"), T("2"), T("t"), T("{(1),(2),(1)}")],
                            &[T("3"), T("2"), T("f"), T("{(1),(2),(3)}")],
                            &[T("5"), T("2"), T("f"), T("{(1),(2),(5)}")],
                            &[T("1"), T("3"), T("t"), T("{(1),(2),(3),(1)}")],
                            &[T("2"), T("3"), T("t"), T("{(1),(2),(3),(2)}")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 05 seed 0000000000000008",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (1, 3), (5, 3), (2, 5), (4, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("2"), T("1"), T("f"), T("{(4),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(4),(2),(3)}")],
                            &[T("5"), T("2"), T("f"), T("{(4),(2),(5)}")],
                            &[T("1"), T("3"), T("f"), T("{(4),(2),(3),(1)}")],
                            &[T("3"), T("3"), T("f"), T("{(4),(2),(5),(3)}")],
                            &[T("1"), T("4"), T("f"), T("{(4),(2),(5),(3),(1)}")],
                            &[T("2"), T("4"), T("t"), T("{(4),(2),(3),(1),(2)}")],
                            &[T("3"), T("4"), T("t"), T("{(4),(2),(3),(1),(3)}")],
                            &[T("2"), T("5"), T("t"), T("{(4),(2),(5),(3),(1),(2)}")],
                            &[T("3"), T("5"), T("t"), T("{(4),(2),(5),(3),(1),(3)}")],
                        ],
                        tag: "SELECT 11",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 06 seed 000000000000000d",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 1), (4, 3), (5, 1), (1, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("1"), T("1"), T("f"), T("{(4),(1)}")],
                            &[T("3"), T("1"), T("f"), T("{(4),(3)}")],
                            &[T("1"), T("2"), T("f"), T("{(4),(3),(1)}")],
                            &[T("2"), T("2"), T("f"), T("{(4),(1),(2)}")],
                            &[T("4"), T("2"), T("t"), T("{(4),(1),(4)}")],
                            &[T("2"), T("3"), T("f"), T("{(4),(3),(1),(2)}")],
                            &[T("3"), T("3"), T("f"), T("{(4),(1),(2),(3)}")],
                            &[T("4"), T("3"), T("t"), T("{(4),(3),(1),(4)}")],
                            &[T("1"), T("4"), T("t"), T("{(4),(1),(2),(3),(1)}")],
                            &[T("3"), T("4"), T("t"), T("{(4),(3),(1),(2),(3)}")],
                        ],
                        tag: "SELECT 11",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 07 seed 0000000000000015",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (5, 5), (2, 5), (4, 1), (3, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 2, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("2"), T("0"), T("f"), T("{(2)}")],
                            &[T("3"), T("1"), T("f"), T("{(2),(3)}")],
                            &[T("5"), T("1"), T("f"), T("{(2),(5)}")],
                            &[T("1"), T("2"), T("f"), T("{(2),(3),(1)}")],
                            &[T("2"), T("2"), T("t"), T("{(2),(3),(2)}")],
                            &[T("5"), T("2"), T("t"), T("{(2),(5),(5)}")],
                            &[T("2"), T("3"), T("t"), T("{(2),(3),(1),(2)}")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 08 seed 0000000000000022",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (2, 1), (5, 3), (3, 5), (5, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 5, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("5"), T("0"), T("f"), T("{(5)}")],
                            &[T("3"), T("1"), T("f"), T("{(5),(3)}")],
                            &[T("5"), T("1"), T("t"), T("{(5),(5)}")],
                            &[T("1"), T("2"), T("f"), T("{(5),(3),(1)}")],
                            &[T("5"), T("2"), T("t"), T("{(5),(3),(5)}")],
                            &[T("2"), T("3"), T("f"), T("{(5),(3),(1),(2)}")],
                            &[T("1"), T("4"), T("t"), T("{(5),(3),(1),(2),(1)}")],
                            &[T("3"), T("4"), T("t"), T("{(5),(3),(1),(2),(3)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 09 seed 9e3779b97f4a7c15",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 5), (4, 1), (2, 1), (5, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("1"), T("2"), T("t"), T("{(1),(2),(1)}")],
                            &[T("3"), T("2"), T("f"), T("{(1),(2),(3)}")],
                            &[T("1"), T("3"), T("t"), T("{(1),(2),(3),(1)}")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 10 seed d1b54a32d192ed03",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 1), (5, 2), (3, 3), (4, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 5, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("5"), T("0"), T("f"), T("{(5)}")],
                            &[T("2"), T("1"), T("f"), T("{(5),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(5),(2),(3)}")],
                            &[T("1"), T("3"), T("f"), T("{(5),(2),(3),(1)}")],
                            &[T("3"), T("3"), T("t"), T("{(5),(2),(3),(3)}")],
                            &[T("2"), T("4"), T("t"), T("{(5),(2),(3),(1),(2)}")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 11 seed 62fec2eed91dc673",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (5, 3), (4, 1), (4, 4), (3, 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("1"), T("1"), T("f"), T("{(4),(1)}")],
                            &[T("4"), T("1"), T("t"), T("{(4),(4)}")],
                            &[T("2"), T("2"), T("f"), T("{(4),(1),(2)}")],
                            &[T("3"), T("3"), T("f"), T("{(4),(1),(2),(3)}")],
                            &[T("1"), T("4"), T("t"), T("{(4),(1),(2),(3),(1)}")],
                            &[T("3"), T("4"), T("t"), T("{(4),(1),(2),(3),(3)}")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 12 seed 4301b2da9cc20ff8",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (1, 4), (3, 2), (2, 1), (5, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 3, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("3"), T("0"), T("f"), T("{(3)}")],
                            &[T("1"), T("1"), T("f"), T("{(3),(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(3),(2)}")],
                            &[T("1"), T("2"), T("f"), T("{(3),(2),(1)}")],
                            &[T("2"), T("2"), T("f"), T("{(3),(1),(2)}")],
                            &[T("3"), T("2"), T("t"), T("{(3),(2),(3)}")],
                            &[T("4"), T("2"), T("f"), T("{(3),(1),(4)}")],
                            &[T("1"), T("3"), T("t"), T("{(3),(1),(2),(1)}")],
                            &[T("2"), T("3"), T("t"), T("{(3),(2),(1),(2)}")],
                            &[T("3"), T("3"), T("t"), T("{(3),(1),(2),(3)}")],
                            &[T("4"), T("3"), T("f"), T("{(3),(2),(1),(4)}")],
                        ],
                        tag: "SELECT 11",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 13 seed 10a0faf22c2365ff",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (5, 2), (2, 4), (3, 5), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 14 seed 89ebda7b349b3e38",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 2), (4, 5), (3, 3), (2, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("2"), T("1"), T("f"), T("{(4),(2)}")],
                            &[T("5"), T("1"), T("f"), T("{(4),(5)}")],
                            &[T("3"), T("2"), T("f"), T("{(4),(2),(3)}")],
                            &[T("5"), T("2"), T("f"), T("{(4),(2),(5)}")],
                            &[T("1"), T("3"), T("f"), T("{(4),(2),(3),(1)}")],
                            &[T("3"), T("3"), T("t"), T("{(4),(2),(3),(3)}")],
                            &[T("2"), T("4"), T("t"), T("{(4),(2),(3),(1),(2)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 15 seed a3421a539064428a",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 1), (4, 5), (1, 1), (5, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 2, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("2"), T("0"), T("f"), T("{(2)}")],
                            &[T("3"), T("1"), T("f"), T("{(2),(3)}")],
                            &[T("1"), T("2"), T("f"), T("{(2),(3),(1)}")],
                            &[T("1"), T("3"), T("t"), T("{(2),(3),(1),(1)}")],
                            &[T("2"), T("3"), T("t"), T("{(2),(3),(1),(2)}")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 16 seed deab883675571e4d",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (1, 3), (4, 1), (2, 4), (2, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 5, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("5"), T("0"), T("f"), T("{(5)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 17 seed 9040e9a76847ca6b",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (1, 3), (3, 3), (4, 1), (1, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 3, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("3"), T("0"), T("f"), T("{(3)}")],
                            &[T("1"), T("1"), T("f"), T("{(3),(1)}")],
                            &[T("3"), T("1"), T("t"), T("{(3),(3)}")],
                            &[T("2"), T("2"), T("f"), T("{(3),(1),(2)}")],
                            &[T("3"), T("2"), T("t"), T("{(3),(1),(3)}")],
                            &[T("4"), T("2"), T("f"), T("{(3),(1),(4)}")],
                            &[T("1"), T("3"), T("t"), T("{(3),(1),(4),(1)}")],
                            &[T("3"), T("3"), T("t"), T("{(3),(1),(2),(3)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 18 seed 93e8341d44bc3cf3",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 5), (5, 2), (1, 5), (4, 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 5, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("5"), T("0"), T("f"), T("{(5)}")],
                            &[T("2"), T("1"), T("f"), T("{(5),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(5),(2),(3)}")],
                            &[T("1"), T("3"), T("f"), T("{(5),(2),(3),(1)}")],
                            &[T("2"), T("4"), T("t"), T("{(5),(2),(3),(1),(2)}")],
                            &[T("5"), T("4"), T("t"), T("{(5),(2),(3),(1),(5)}")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 19 seed 95a36e8b834ccce4",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (2, 1), (4, 5), (4, 2), (4, 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("1"), T("2"), T("t"), T("{(1),(2),(1)}")],
                            &[T("3"), T("2"), T("f"), T("{(1),(2),(3)}")],
                            &[T("1"), T("3"), T("t"), T("{(1),(2),(3),(1)}")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 20 seed d0ee03218946e4fe",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (5, 4), (1, 5), (2, 2), (3, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 2, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("2"), T("0"), T("f"), T("{(2)}")],
                            &[T("2"), T("1"), T("t"), T("{(2),(2)}")],
                            &[T("3"), T("1"), T("f"), T("{(2),(3)}")],
                            &[T("1"), T("2"), T("f"), T("{(2),(3),(1)}")],
                            &[T("4"), T("2"), T("f"), T("{(2),(3),(4)}")],
                            &[T("2"), T("3"), T("t"), T("{(2),(3),(1),(2)}")],
                            &[T("5"), T("3"), T("f"), T("{(2),(3),(1),(5)}")],
                            &[T("4"), T("4"), T("f"), T("{(2),(3),(1),(5),(4)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 21 seed 48289f16eec514ff",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 2), (3, 3), (5, 4), (1, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 22 seed e7124832cd5a455c",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (2, 4), (4, 1), (4, 5), (4, 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 3, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("3"), T("0"), T("f"), T("{(3)}")],
                            &[T("1"), T("1"), T("f"), T("{(3),(1)}")],
                            &[T("2"), T("2"), T("f"), T("{(3),(1),(2)}")],
                            &[T("3"), T("3"), T("t"), T("{(3),(1),(2),(3)}")],
                            &[T("4"), T("3"), T("f"), T("{(3),(1),(2),(4)}")],
                            &[T("1"), T("4"), T("t"), T("{(3),(1),(2),(4),(1)}")],
                            &[T("3"), T("4"), T("t"), T("{(3),(1),(2),(4),(3)}")],
                            &[T("5"), T("4"), T("f"), T("{(3),(1),(2),(4),(5)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 23 seed f0a835222287a683",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (1, 4), (5, 1), (1, 5), (5, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 2, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("2"), T("0"), T("f"), T("{(2)}")],
                            &[T("3"), T("1"), T("f"), T("{(2),(3)}")],
                            &[T("1"), T("2"), T("f"), T("{(2),(3),(1)}")],
                            &[T("2"), T("3"), T("t"), T("{(2),(3),(1),(2)}")],
                            &[T("4"), T("3"), T("f"), T("{(2),(3),(1),(4)}")],
                            &[T("5"), T("3"), T("f"), T("{(2),(3),(1),(5)}")],
                            &[T("1"), T("4"), T("t"), T("{(2),(3),(1),(5),(1)}")],
                            &[T("5"), T("4"), T("t"), T("{(2),(3),(1),(5),(5)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 24 seed 122a2f3989dff1b4",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 2), (5, 2), (4, 2), (3, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(1),(2),(3)}")],
                            &[T("1"), T("3"), T("t"), T("{(1),(2),(3),(1)}")],
                            &[T("2"), T("3"), T("t"), T("{(1),(2),(3),(2)}")],
                            &[T("5"), T("3"), T("f"), T("{(1),(2),(3),(5)}")],
                            &[T("2"), T("4"), T("t"), T("{(1),(2),(3),(5),(2)}")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 25 seed 0ac7bdf3d90af2fd",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 5), (2, 1), (5, 5), (5, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 26 seed f10de05685aeabbb",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 4), (4, 5), (2, 2), (4, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 5, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("5"), T("0"), T("f"), T("{(5)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 27 seed bfa07d1a2e0d0363",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (2, 2), (5, 3), (1, 3), (4, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("5"), T("1"), T("f"), T("{(4),(5)}")],
                            &[T("3"), T("2"), T("f"), T("{(4),(5),(3)}")],
                            &[T("1"), T("3"), T("f"), T("{(4),(5),(3),(1)}")],
                            &[T("2"), T("4"), T("f"), T("{(4),(5),(3),(1),(2)}")],
                            &[T("3"), T("4"), T("t"), T("{(4),(5),(3),(1),(3)}")],
                            &[T("2"), T("5"), T("t"), T("{(4),(5),(3),(1),(2),(2)}")],
                            &[T("3"), T("5"), T("t"), T("{(4),(5),(3),(1),(2),(3)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 28 seed 11fc4c7494796cf1",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (2, 5), (5, 1), (3, 5), (4, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 5, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("5"), T("0"), T("f"), T("{(5)}")],
                            &[T("1"), T("1"), T("f"), T("{(5),(1)}")],
                            &[T("2"), T("2"), T("f"), T("{(5),(1),(2)}")],
                            &[T("3"), T("3"), T("f"), T("{(5),(1),(2),(3)}")],
                            &[T("5"), T("3"), T("t"), T("{(5),(1),(2),(5)}")],
                            &[T("1"), T("4"), T("t"), T("{(5),(1),(2),(3),(1)}")],
                            &[T("5"), T("4"), T("t"), T("{(5),(1),(2),(3),(5)}")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 29 seed 7057fbbb0a9ee37e",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (5, 3), (5, 4), (3, 5), (5, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 30 seed 092fdd61d4fdf5f9",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 2), (5, 3), (2, 5), (5, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("2"), T("1"), T("f"), T("{(4),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(4),(2),(3)}")],
                            &[T("5"), T("2"), T("f"), T("{(4),(2),(5)}")],
                            &[T("1"), T("3"), T("f"), T("{(4),(2),(3),(1)}")],
                            &[T("2"), T("3"), T("t"), T("{(4),(2),(5),(2)}")],
                            &[T("3"), T("3"), T("f"), T("{(4),(2),(5),(3)}")],
                            &[T("1"), T("4"), T("f"), T("{(4),(2),(5),(3),(1)}")],
                            &[T("2"), T("4"), T("t"), T("{(4),(2),(3),(1),(2)}")],
                            &[T("2"), T("5"), T("t"), T("{(4),(2),(5),(3),(1),(2)}")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 31 seed 1687ced1070d7f8b",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (5, 4), (1, 4), (3, 3), (4, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("2"), T("1"), T("f"), T("{(4),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(4),(2),(3)}")],
                            &[T("1"), T("3"), T("f"), T("{(4),(2),(3),(1)}")],
                            &[T("3"), T("3"), T("t"), T("{(4),(2),(3),(3)}")],
                            &[T("2"), T("4"), T("t"), T("{(4),(2),(3),(1),(2)}")],
                            &[T("4"), T("4"), T("t"), T("{(4),(2),(3),(1),(4)}")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 32 seed 2bcff8b03d78be9b",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (1, 5), (3, 3), (5, 5), (4, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("5"), T("1"), T("f"), T("{(1),(5)}")],
                            &[T("3"), T("2"), T("f"), T("{(1),(2),(3)}")],
                            &[T("5"), T("2"), T("t"), T("{(1),(5),(5)}")],
                            &[T("1"), T("3"), T("t"), T("{(1),(2),(3),(1)}")],
                            &[T("3"), T("3"), T("t"), T("{(1),(2),(3),(3)}")],
                        ],
                        tag: "SELECT 7",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 33 seed bb1477a73b25a81f",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (5, 4), (3, 2), (1, 5), (2, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 34 seed c6585c83de531aeb",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 4), (4, 5), (5, 1), (5, 3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(1),(2),(3)}")],
                            &[T("1"), T("3"), T("t"), T("{(1),(2),(3),(1)}")],
                            &[T("4"), T("3"), T("f"), T("{(1),(2),(3),(4)}")],
                            &[T("5"), T("4"), T("f"), T("{(1),(2),(3),(4),(5)}")],
                            &[T("1"), T("5"), T("t"), T("{(1),(2),(3),(4),(5),(1)}")],
                            &[T("3"), T("5"), T("t"), T("{(1),(2),(3),(4),(5),(3)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 35 seed 2db6991b0eb906e8",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 3), (3, 5), (4, 2), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 5, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("5"), T("0"), T("f"), T("{(5)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 36 seed ed59257df91e6b2c",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 2), (4, 5), (2, 5), (5, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("5"), T("1"), T("f"), T("{(4),(5)}")],
                            &[T("5"), T("2"), T("t"), T("{(4),(5),(5)}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 37 seed 7887be1133a582f0",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 5), (5, 3), (1, 4), (4, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("1"), T("1"), T("f"), T("{(4),(1)}")],
                            &[T("5"), T("1"), T("f"), T("{(4),(5)}")],
                            &[T("2"), T("2"), T("f"), T("{(4),(1),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(4),(5),(3)}")],
                            &[T("4"), T("2"), T("t"), T("{(4),(1),(4)}")],
                            &[T("1"), T("3"), T("f"), T("{(4),(5),(3),(1)}")],
                            &[T("3"), T("3"), T("f"), T("{(4),(1),(2),(3)}")],
                            &[T("1"), T("4"), T("t"), T("{(4),(1),(2),(3),(1)}")],
                            &[T("2"), T("4"), T("f"), T("{(4),(5),(3),(1),(2)}")],
                            &[T("4"), T("4"), T("t"), T("{(4),(5),(3),(1),(4)}")],
                            &[T("3"), T("5"), T("t"), T("{(4),(5),(3),(1),(2),(3)}")],
                        ],
                        tag: "SELECT 12",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 38 seed 804a8876f0f601f5",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (5, 5), (2, 5), (5, 4), (3, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 39 seed fbe23113e3f39920",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (3, 3), (4, 1), (2, 4), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("2"), T("2"), T("t"), T("{(1),(2),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(1),(2),(3)}")],
                            &[T("4"), T("2"), T("f"), T("{(1),(2),(4)}")],
                            &[T("1"), T("3"), T("t"), T("{(1),(2),(3),(1)}")],
                            &[T("1"), T("3"), T("t"), T("{(1),(2),(4),(1)}")],
                            &[T("3"), T("3"), T("t"), T("{(1),(2),(3),(3)}")],
                        ],
                        tag: "SELECT 8",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 40 seed 79c9eda27f039a47",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 5), (5, 5), (3, 2), (2, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 2, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("2"), T("0"), T("f"), T("{(2)}")],
                            &[T("2"), T("1"), T("t"), T("{(2),(2)}")],
                            &[T("3"), T("1"), T("f"), T("{(2),(3)}")],
                            &[T("1"), T("2"), T("f"), T("{(2),(3),(1)}")],
                            &[T("2"), T("2"), T("t"), T("{(2),(3),(2)}")],
                            &[T("2"), T("3"), T("t"), T("{(2),(3),(1),(2)}")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated CYCLE graph 41 seed ebee2d9ccd7c7521",
            set_up_script: &[
                "CREATE TABLE generated_cycle_edges (source INT, target INT);",
                "INSERT INTO generated_cycle_edges VALUES (1, 2), (2, 3), (3, 1), (4, 1), (2, 5), (4, 2), (4, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 4, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("4"), T("0"), T("f"), T("{(4)}")],
                            &[T("1"), T("1"), T("f"), T("{(4),(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(4),(2)}")],
                            &[T("5"), T("1"), T("f"), T("{(4),(5)}")],
                            &[T("2"), T("2"), T("f"), T("{(4),(1),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(4),(2),(3)}")],
                            &[T("5"), T("2"), T("f"), T("{(4),(2),(5)}")],
                            &[T("1"), T("3"), T("f"), T("{(4),(2),(3),(1)}")],
                            &[T("3"), T("3"), T("f"), T("{(4),(1),(2),(3)}")],
                            &[T("5"), T("3"), T("f"), T("{(4),(1),(2),(5)}")],
                            &[T("1"), T("4"), T("t"), T("{(4),(1),(2),(3),(1)}")],
                            &[T("2"), T("4"), T("t"), T("{(4),(2),(3),(1),(2)}")],
                        ],
                        tag: "SELECT 12",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_with_recursive() {
    run_scripts(&[
        ScriptTest {
            name: "Documentation Examples",
            set_up_script: &[
                r#"CREATE TABLE graph (
  id   integer PRIMARY KEY,
  link integer,
  data text NOT NULL
);"#,
                r#"INSERT INTO graph (id, link, data) VALUES
  (1, 2,    'start of cyclic branch'),
  (2, 3,    'cycle node two'),
  (3, 1,    'cycle node three'),
  (4, 5,    'start of acyclic branch'),
  (5, 6,    'middle of acyclic branch'),
  (6, NULL, 'end of acyclic branch'),
  (7, 5,    'second path into node five');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE search_graph(id, link, data, depth, is_cycle, path) AS (
  SELECT g.id, g.link, g.data, 0,
    false,
    ARRAY[g.id]
  FROM graph g
UNION ALL
  SELECT g.id, g.link, g.data, sg.depth + 1,
    g.id = ANY(path),
    path || g.id
  FROM graph g, search_graph sg
  WHERE g.id = sg.link AND NOT is_cycle
)
SELECT * FROM search_graph;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("link", INT4), Column("data", TEXT), Column("depth", INT4), Column("is_cycle", BOOL), Column("path", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("2"), T("start of cyclic branch"), T("0"), T("f"), T("{1}")],
                            &[T("2"), T("3"), T("cycle node two"), T("0"), T("f"), T("{2}")],
                            &[T("3"), T("1"), T("cycle node three"), T("0"), T("f"), T("{3}")],
                            &[T("4"), T("5"), T("start of acyclic branch"), T("0"), T("f"), T("{4}")],
                            &[T("5"), T("6"), T("middle of acyclic branch"), T("0"), T("f"), T("{5}")],
                            &[T("6"), Null, T("end of acyclic branch"), T("0"), T("f"), T("{6}")],
                            &[T("7"), T("5"), T("second path into node five"), T("0"), T("f"), T("{7}")],
                            &[T("2"), T("3"), T("cycle node two"), T("1"), T("f"), T("{1,2}")],
                            &[T("3"), T("1"), T("cycle node three"), T("1"), T("f"), T("{2,3}")],
                            &[T("1"), T("2"), T("start of cyclic branch"), T("1"), T("f"), T("{3,1}")],
                            &[T("5"), T("6"), T("middle of acyclic branch"), T("1"), T("f"), T("{4,5}")],
                            &[T("6"), Null, T("end of acyclic branch"), T("1"), T("f"), T("{5,6}")],
                            &[T("5"), T("6"), T("middle of acyclic branch"), T("1"), T("f"), T("{7,5}")],
                            &[T("3"), T("1"), T("cycle node three"), T("2"), T("f"), T("{1,2,3}")],
                            &[T("1"), T("2"), T("start of cyclic branch"), T("2"), T("f"), T("{2,3,1}")],
                            &[T("2"), T("3"), T("cycle node two"), T("2"), T("f"), T("{3,1,2}")],
                            &[T("6"), Null, T("end of acyclic branch"), T("2"), T("f"), T("{4,5,6}")],
                            &[T("6"), Null, T("end of acyclic branch"), T("2"), T("f"), T("{7,5,6}")],
                            &[T("1"), T("2"), T("start of cyclic branch"), T("3"), T("t"), T("{1,2,3,1}")],
                            &[T("2"), T("3"), T("cycle node two"), T("3"), T("t"), T("{2,3,1,2}")],
                            &[T("3"), T("1"), T("cycle node three"), T("3"), T("t"), T("{3,1,2,3}")],
                        ],
                        tag: "SELECT 21",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"WITH RECURSIVE search_graph(id, link, data, depth) AS (
  SELECT g.id, g.link, g.data, 1
  FROM graph AS g
  UNION ALL
  SELECT g.id, g.link, g.data, sg.depth + 1
  FROM graph AS g
  JOIN search_graph AS sg ON g.id = sg.link
) CYCLE id SET is_cycle USING path
SELECT * FROM search_graph ORDER BY path;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("link", INT4), Column("data", TEXT), Column("depth", INT4), Column("is_cycle", BOOL), Column("path", RECORD_ARRAY)],
                        rows: &[
                            &[T("1"), T("2"), T("start of cyclic branch"), T("1"), T("f"), T("{(1)}")],
                            &[T("2"), T("3"), T("cycle node two"), T("2"), T("f"), T("{(1),(2)}")],
                            &[T("3"), T("1"), T("cycle node three"), T("3"), T("f"), T("{(1),(2),(3)}")],
                            &[T("1"), T("2"), T("start of cyclic branch"), T("4"), T("t"), T("{(1),(2),(3),(1)}")],
                            &[T("2"), T("3"), T("cycle node two"), T("1"), T("f"), T("{(2)}")],
                            &[T("3"), T("1"), T("cycle node three"), T("2"), T("f"), T("{(2),(3)}")],
                            &[T("1"), T("2"), T("start of cyclic branch"), T("3"), T("f"), T("{(2),(3),(1)}")],
                            &[T("2"), T("3"), T("cycle node two"), T("4"), T("t"), T("{(2),(3),(1),(2)}")],
                            &[T("3"), T("1"), T("cycle node three"), T("1"), T("f"), T("{(3)}")],
                            &[T("1"), T("2"), T("start of cyclic branch"), T("2"), T("f"), T("{(3),(1)}")],
                            &[T("2"), T("3"), T("cycle node two"), T("3"), T("f"), T("{(3),(1),(2)}")],
                            &[T("3"), T("1"), T("cycle node three"), T("4"), T("t"), T("{(3),(1),(2),(3)}")],
                            &[T("4"), T("5"), T("start of acyclic branch"), T("1"), T("f"), T("{(4)}")],
                            &[T("5"), T("6"), T("middle of acyclic branch"), T("2"), T("f"), T("{(4),(5)}")],
                            &[T("6"), Null, T("end of acyclic branch"), T("3"), T("f"), T("{(4),(5),(6)}")],
                            &[T("5"), T("6"), T("middle of acyclic branch"), T("1"), T("f"), T("{(5)}")],
                            &[T("6"), Null, T("end of acyclic branch"), T("2"), T("f"), T("{(5),(6)}")],
                            &[T("6"), Null, T("end of acyclic branch"), T("1"), T("f"), T("{(6)}")],
                            &[T("7"), T("5"), T("second path into node five"), T("1"), T("f"), T("{(7)}")],
                            &[T("5"), T("6"), T("middle of acyclic branch"), T("2"), T("f"), T("{(7),(5)}")],
                            &[T("6"), Null, T("end of acyclic branch"), T("3"), T("f"), T("{(7),(5),(6)}")],
                        ],
                        tag: "SELECT 21",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE adds marker and path columns with PostgreSQL types",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n + 1 FROM walk WHERE n < 3
) CYCLE n SET is_cycle USING path
SELECT
	n,
	is_cycle,
	path::text AS path_text,
	pg_typeof(is_cycle)::text AS marker_type,
	pg_typeof(path)::text AS path_type
FROM walk
ORDER BY n;"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT), Column("marker_type", TEXT), Column("path_type", TEXT)],
                        rows: &[
                            &[T("1"), T("f"), T("{(1)}"), T("boolean"), T("record[]")],
                            &[T("2"), T("f"), T("{(1),(2)}"), T("boolean"), T("record[]")],
                            &[T("3"), T("f"), T("{(1),(2),(3)}"), T("boolean"), T("record[]")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE detects a self-loop and emits the closing row",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT n, depth + 1 FROM walk
) CYCLE n SET is_cycle USING path
SELECT n, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("1"), T("1"), T("t"), T("{(1),(1)}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE detects a cycle after an acyclic prefix",
            set_up_script: &[
                "CREATE TABLE cycle_edges (source INT, target INT);",
                "INSERT INTO cycle_edges VALUES (1, 2), (2, 3), (3, 2);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("3"), T("2"), T("f"), T("{(1),(2),(3)}")],
                            &[T("2"), T("3"), T("t"), T("{(1),(2),(3),(2)}")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE detection is path-local for converging branches",
            set_up_script: &[
                "CREATE TABLE diamond_edges (source INT, target INT);",
                "INSERT INTO diamond_edges VALUES (1, 2), (1, 3), (2, 4), (3, 4);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM diamond_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("3"), T("1"), T("f"), T("{(1),(3)}")],
                            &[T("4"), T("2"), T("f"), T("{(1),(2),(4)}")],
                            &[T("4"), T("2"), T("f"), T("{(1),(3),(4)}")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "A cyclic branch does not suppress an independent branch",
            set_up_script: &[
                "CREATE TABLE branch_edges (source INT, target INT);",
                "INSERT INTO branch_edges VALUES (1, 2), (1, 3), (2, 4), (4, 2), (3, 5);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM branch_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("3"), T("1"), T("f"), T("{(1),(3)}")],
                            &[T("4"), T("2"), T("f"), T("{(1),(2),(4)}")],
                            &[T("5"), T("2"), T("f"), T("{(1),(3),(5)}")],
                            &[T("2"), T("3"), T("t"), T("{(1),(2),(4),(2)}")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE initializes independent paths for multiple anchor rows",
            set_up_script: &[
                "CREATE TABLE multi_anchor_edges (source INT, target INT);",
                "INSERT INTO multi_anchor_edges VALUES (1, 2), (2, 1), (10, 11), (11, 10);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(root, node, depth) AS (
	VALUES (1, 1, 0), (10, 10, 0)
	UNION ALL
	SELECT w.root, e.target, w.depth + 1
	FROM multi_anchor_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT root, node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY root, depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("root", INT4), Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("1"), T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("1"), T("1"), T("2"), T("t"), T("{(1),(2),(1)}")],
                            &[T("10"), T("10"), T("0"), T("f"), T("{(10)}")],
                            &[T("10"), T("11"), T("1"), T("f"), T("{(10),(11)}")],
                            &[T("10"), T("10"), T("2"), T("t"), T("{(10),(11),(10)}")],
                        ],
                        tag: "SELECT 6",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE supports composite cycle keys",
            set_up_script: &[
                r#"CREATE TABLE composite_edges (
				from_namespace INT,
				from_node INT,
				to_namespace INT,
				to_node INT
			);"#,
                r#"INSERT INTO composite_edges VALUES
				(1, 1, 2, 1),
				(2, 1, 2, 2),
				(2, 2, 1, 1);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(namespace_id, node_id, depth) AS (
	SELECT 1, 1, 0
	UNION ALL
	SELECT e.to_namespace, e.to_node, w.depth + 1
	FROM composite_edges e
	JOIN walk w
	  ON e.from_namespace = w.namespace_id
	 AND e.from_node = w.node_id
) CYCLE namespace_id, node_id SET is_cycle USING path
SELECT namespace_id, node_id, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("namespace_id", INT4), Column("node_id", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("0"), T("f"), T(r#"{"(1,1)"}"#)],
                            &[T("2"), T("1"), T("1"), T("f"), T(r#"{"(1,1)","(2,1)"}"#)],
                            &[T("2"), T("2"), T("2"), T("f"), T(r#"{"(1,1)","(2,1)","(2,2)"}"#)],
                            &[T("1"), T("1"), T("3"), T("t"), T(r#"{"(1,1)","(2,1)","(2,2)","(1,1)"}"#)],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE treats repeated NULL cycle keys as equal",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(key, depth) AS (
	SELECT NULL::INT, 0
	UNION ALL
	SELECT NULL::INT, depth + 1 FROM walk
) CYCLE key SET is_cycle USING path
SELECT key, depth, is_cycle
FROM walk
ORDER BY depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("key", INT4), Column("depth", INT4), Column("is_cycle", BOOL)],
                        rows: &[
                            &[Null, T("0"), T("f")],
                            &[Null, T("1"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE supports custom text marker values",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT n, depth + 1 FROM walk
) CYCLE n SET cycle_mark TO 'cycle' DEFAULT 'ok' USING path
SELECT n, depth, cycle_mark, pg_typeof(cycle_mark)::text AS marker_type
FROM walk
ORDER BY depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("depth", INT4), Column("cycle_mark", TEXT), Column("marker_type", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("ok"), T("text")],
                            &[T("1"), T("1"), T("cycle"), T("text")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE supports custom integer marker values",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT n, depth + 1 FROM walk
) CYCLE n SET cycle_mark TO 99 DEFAULT 0 USING path
SELECT n, depth, cycle_mark, pg_typeof(cycle_mark)::text AS marker_type
FROM walk
ORDER BY depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("depth", INT4), Column("cycle_mark", INT4), Column("marker_type", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("0"), T("integer")],
                            &[T("1"), T("1"), T("99"), T("integer")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Generated cycle marker is visible in the recursive term",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT n, depth + 1
	FROM walk
	WHERE NOT is_cycle
) CYCLE n SET is_cycle USING path
SELECT n, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("1"), T("1"), T("t"), T("{(1),(1)}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "SELECT star expands generated cycle columns",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE n SET is_cycle USING path
SELECT * FROM walk;"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("is_cycle", BOOL), Column("path", RECORD_ARRAY)],
                        rows: &[
                            &[T("1"), T("f"), T("{(1)}")],
                            &[T("1"), T("t"), T("{(1),(1)}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE works with UNION DISTINCT",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION
	SELECT n FROM walk
) CYCLE n SET is_cycle USING path
SELECT n, is_cycle, path::text AS path_text
FROM walk;"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("f"), T("{(1)}")],
                            &[T("1"), T("t"), T("{(1),(1)}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE works with sibling and consuming CTEs",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE
edges(source, target) AS (
	VALUES (1, 2), (2, 3), (3, 1)
),
walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path,
cycles AS (
	SELECT node, depth, path::text AS path_text
	FROM walk
	WHERE is_cycle
)
SELECT node, depth, path_text
FROM cycles;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("3"), T("{(1),(2),(3),(1)}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE supports UUID cycle keys",
            set_up_script: &[
                "CREATE TABLE uuid_edges (source UUID, target UUID);",
                r#"INSERT INTO uuid_edges VALUES
				('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002'),
				('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003'),
				('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node_id, depth) AS (
	SELECT '00000000-0000-0000-0000-000000000001'::UUID, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM uuid_edges e
	JOIN walk w ON e.source = w.node_id
) CYCLE node_id SET is_cycle USING path
SELECT node_id::text, depth, is_cycle, pg_typeof(path)::text AS path_type
FROM walk
ORDER BY depth;"#,
                    expected: Expected::Rows {
                        columns: &[Column("node_id", TEXT), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_type", TEXT)],
                        rows: &[
                            &[T("00000000-0000-0000-0000-000000000001"), T("0"), T("f"), T("record[]")],
                            &[T("00000000-0000-0000-0000-000000000002"), T("1"), T("f"), T("record[]")],
                            &[T("00000000-0000-0000-0000-000000000003"), T("2"), T("f"), T("record[]")],
                            &[T("00000000-0000-0000-0000-000000000001"), T("3"), T("t"), T("record[]")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE supports quoted identifiers",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk("Node", "Depth") AS (
	SELECT 1, 0
	UNION ALL
	SELECT "Node", "Depth" + 1 FROM walk
) CYCLE "Node" SET "IsCycle" USING "Path"
SELECT
	"Node",
	"Depth",
	"IsCycle",
	"Path"::text AS "PathText"
FROM walk
ORDER BY "Depth";"#,
                    expected: Expected::Rows {
                        columns: &[Column("Node", INT4), Column("Depth", INT4), Column("IsCycle", BOOL), Column("PathText", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("1"), T("1"), T("t"), T("{(1),(1)}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE supports bind variables in the anchor term",
            set_up_script: &[
                "CREATE TABLE bound_edges (source INT, target INT);",
                "INSERT INTO bound_edges VALUES (1, 2), (2, 3), (3, 1);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(node, depth) AS (
	SELECT $1::INT, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM bound_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle
FROM walk
ORDER BY depth;"#,
                    bind_vars: &[BindVar::Int(1)],
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL)],
                        rows: &[
                            &[T("1"), T("0"), T("f")],
                            &[T("2"), T("1"), T("f")],
                            &[T("3"), T("2"), T("f")],
                            &[T("1"), T("3"), T("t")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE terminates a deep cycle exactly once",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 0
	UNION ALL
	SELECT (n + 1) % 256 FROM walk
) CYCLE n SET is_cycle USING path
SELECT
	count(*)::INT AS row_count,
	count(*) FILTER (WHERE is_cycle)::INT AS cycle_count,
	max(array_length(path, 1)) AS max_path_length
FROM walk;"#,
                    expected: Expected::Rows {
                        columns: &[Column("row_count", INT4), Column("cycle_count", INT4), Column("max_path_length", INT4)],
                        rows: &[
                            &[T("257"), T("1"), T("257")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE works inside a view",
            set_up_script: &[
                "CREATE TABLE view_edges (source INT, target INT);",
                "INSERT INTO view_edges VALUES (1, 2), (2, 1);",
                r#"CREATE VIEW cycle_walk_view AS
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM view_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM cycle_walk_view ORDER BY depth;",
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL), Column("path_text", TEXT)],
                        rows: &[
                            &[T("1"), T("0"), T("f"), T("{(1)}")],
                            &[T("2"), T("1"), T("f"), T("{(1),(2)}")],
                            &[T("1"), T("2"), T("t"), T("{(1),(2),(1)}")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE output can feed a data-modifying statement",
            set_up_script: &[
                "CREATE TABLE dml_edges (source INT, target INT);",
                "INSERT INTO dml_edges VALUES (1, 2), (2, 1);",
                "CREATE TABLE cycle_results (node INT, depth INT, is_cycle BOOLEAN);",
                r#"WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM dml_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
INSERT INTO cycle_results
SELECT node, depth, is_cycle FROM walk;"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT node, depth, is_cycle FROM cycle_results ORDER BY depth;",
                    expected: Expected::Rows {
                        columns: &[Column("node", INT4), Column("depth", INT4), Column("is_cycle", BOOL)],
                        rows: &[
                            &[T("1"), T("0"), T("f")],
                            &[T("2"), T("1"), T("f")],
                            &[T("1"), T("2"), T("t")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE rejects a non-recursive CTE",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 1
) CYCLE n SET is_cycle USING path
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: "WITH query is not recursive", position: 17, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE rejects an unknown cycle column",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE missing SET is_cycle USING path
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"cycle column "missing" not in WITH query column list"#, position: 73, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE rejects duplicate cycle columns",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE n, n SET is_cycle USING path
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "42701", message: r#"cycle column "n" specified more than once"#, position: 73, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE rejects a marker name already in the CTE output",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n, label) AS (
	SELECT 1, 'start'::TEXT
	UNION ALL
	SELECT n, label FROM walk
) CYCLE n SET label USING path
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "42702", message: r#"column reference "label" is ambiguous"#, position: 84, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE rejects a path name already in the CTE output",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n, path) AS (
	SELECT 1, 'start'::TEXT
	UNION ALL
	SELECT n, path FROM walk
) CYCLE n SET is_cycle USING path
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "42702", message: r#"column reference "path" is ambiguous"#, position: 83, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE rejects identical marker and path names",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE n SET generated USING generated
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: "cycle mark column name and cycle path column name are the same", position: 73, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE rejects incompatible marker and default types",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE n SET is_cycle TO TRUE DEFAULT 55 USING path
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "42804", message: "CYCLE types boolean and integer cannot be matched", position: 110, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE requires a SELECT on the left side of the recursive UNION",
            set_up_script: &[
                "CREATE TABLE left_union_graph (source INT, target INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(source, target) AS (
	SELECT * FROM left_union_graph
	UNION ALL
	SELECT * FROM left_union_graph
	UNION ALL
	SELECT g.*
	FROM left_union_graph g
	JOIN walk w ON g.source = w.target
) CYCLE source, target SET is_cycle USING path
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "with a SEARCH or CYCLE clause, the left side of the UNION must be a SELECT", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "CYCLE requires a SELECT on the right side of the recursive UNION",
            set_up_script: &[
                "CREATE TABLE right_union_graph (source INT, target INT);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
WITH RECURSIVE walk(source, target) AS (
	SELECT * FROM right_union_graph
	UNION ALL
	(
		SELECT * FROM right_union_graph
		UNION ALL
		SELECT g.*
		FROM right_union_graph g
		JOIN walk w ON g.source = w.target
	)
) CYCLE source, target SET is_cycle USING path
SELECT * FROM walk;"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: "with a SEARCH or CYCLE clause, the right side of the UNION must be a SELECT", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
