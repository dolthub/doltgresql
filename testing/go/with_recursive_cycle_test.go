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

package _go

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestWithRecursive(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Documentation Examples", // https://www.postgresql.org/docs/15/queries-with.html#QUERIES-WITH-CYCLE
			SetUpScript: []string{
				`CREATE TABLE graph (
  id   integer PRIMARY KEY,
  link integer,
  data text NOT NULL
);`,
				`INSERT INTO graph (id, link, data) VALUES
  (1, 2,    'start of cyclic branch'),
  (2, 3,    'cycle node two'),
  (3, 1,    'cycle node three'),
  (4, 5,    'start of acyclic branch'),
  (5, 6,    'middle of acyclic branch'),
  (6, NULL, 'end of acyclic branch'),
  (7, 5,    'second path into node five');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `WITH RECURSIVE search_graph(id, link, data, depth, is_cycle, path) AS (
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
SELECT * FROM search_graph;`,
					ExpectedColNames: []string{"id", "link", "data", "depth", "is_cycle", "path"},
					Expected: []sql.Row{
						{1, 2, "start of cyclic branch", 0, "f", "{1}"},
						{2, 3, "cycle node two", 0, "f", "{2}"},
						{3, 1, "cycle node three", 0, "f", "{3}"},
						{4, 5, "start of acyclic branch", 0, "f", "{4}"},
						{5, 6, "middle of acyclic branch", 0, "f", "{5}"},
						{6, nil, "end of acyclic branch", 0, "f", "{6}"},
						{7, 5, "second path into node five", 0, "f", "{7}"},
						{2, 3, "cycle node two", 1, "f", "{1,2}"},
						{3, 1, "cycle node three", 1, "f", "{2,3}"},
						{1, 2, "start of cyclic branch", 1, "f", "{3,1}"},
						{5, 6, "middle of acyclic branch", 1, "f", "{4,5}"},
						{6, nil, "end of acyclic branch", 1, "f", "{5,6}"},
						{5, 6, "middle of acyclic branch", 1, "f", "{7,5}"},
						{3, 1, "cycle node three", 2, "f", "{1,2,3}"},
						{1, 2, "start of cyclic branch", 2, "f", "{2,3,1}"},
						{2, 3, "cycle node two", 2, "f", "{3,1,2}"},
						{6, nil, "end of acyclic branch", 2, "f", "{4,5,6}"},
						{6, nil, "end of acyclic branch", 2, "f", "{7,5,6}"},
						{1, 2, "start of cyclic branch", 3, "t", "{1,2,3,1}"},
						{2, 3, "cycle node two", 3, "t", "{2,3,1,2}"},
						{3, 1, "cycle node three", 3, "t", "{3,1,2,3}"},
					},
				},
				{
					Query: `WITH RECURSIVE search_graph(id, link, data, depth) AS (
  SELECT g.id, g.link, g.data, 1
  FROM graph AS g
  UNION ALL
  SELECT g.id, g.link, g.data, sg.depth + 1
  FROM graph AS g
  JOIN search_graph AS sg ON g.id = sg.link
) CYCLE id SET is_cycle USING path
SELECT * FROM search_graph ORDER BY path;`,
					ExpectedColNames: []string{"id", "link", "data", "depth", "is_cycle", "path"},
					Expected: []sql.Row{
						{1, 2, "start of cyclic branch", 1, "f", "{(1)}"},
						{2, 3, "cycle node two", 2, "f", "{(1),(2)}"},
						{3, 1, "cycle node three", 3, "f", "{(1),(2),(3)}"},
						{1, 2, "start of cyclic branch", 4, "t", "{(1),(2),(3),(1)}"},
						{2, 3, "cycle node two", 1, "f", "{(2)}"},
						{3, 1, "cycle node three", 2, "f", "{(2),(3)}"},
						{1, 2, "start of cyclic branch", 3, "f", "{(2),(3),(1)}"},
						{2, 3, "cycle node two", 4, "t", "{(2),(3),(1),(2)}"},
						{3, 1, "cycle node three", 1, "f", "{(3)}"},
						{1, 2, "start of cyclic branch", 2, "f", "{(3),(1)}"},
						{2, 3, "cycle node two", 3, "f", "{(3),(1),(2)}"},
						{3, 1, "cycle node three", 4, "t", "{(3),(1),(2),(3)}"},
						{4, 5, "start of acyclic branch", 1, "f", "{(4)}"},
						{5, 6, "middle of acyclic branch", 2, "f", "{(4),(5)}"},
						{6, nil, "end of acyclic branch", 3, "f", "{(4),(5),(6)}"},
						{5, 6, "middle of acyclic branch", 1, "f", "{(5)}"},
						{6, nil, "end of acyclic branch", 2, "f", "{(5),(6)}"},
						{6, nil, "end of acyclic branch", 1, "f", "{(6)}"},
						{7, 5, "second path into node five", 1, "f", "{(7)}"},
						{5, 6, "middle of acyclic branch", 2, "f", "{(7),(5)}"},
						{6, nil, "end of acyclic branch", 3, "f", "{(7),(5),(6)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE adds marker and path columns with PostgreSQL types",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
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
ORDER BY n;`,
					ExpectedColNames: []string{"n", "is_cycle", "path_text", "marker_type", "path_type"},
					Expected: []sql.Row{
						{1, "f", "{(1)}", "boolean", "record[]"},
						{2, "f", "{(1),(2)}", "boolean", "record[]"},
						{3, "f", "{(1),(2),(3)}", "boolean", "record[]"},
					},
				},
			},
		},
		{
			Name: "CYCLE detects a self-loop and emits the closing row",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT n, depth + 1 FROM walk
) CYCLE n SET is_cycle USING path
SELECT n, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth;`,
					Expected: []sql.Row{
						{1, 0, "f", "{(1)}"},
						{1, 1, "t", "{(1),(1)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE detects a cycle after an acyclic prefix",
			SetUpScript: []string{
				`CREATE TABLE cycle_edges (source INT, target INT);`,
				`INSERT INTO cycle_edges VALUES (1, 2), (2, 3), (3, 2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth;`,
					Expected: []sql.Row{
						{1, 0, "f", "{(1)}"},
						{2, 1, "f", "{(1),(2)}"},
						{3, 2, "f", "{(1),(2),(3)}"},
						{2, 3, "t", "{(1),(2),(3),(2)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE detection is path-local for converging branches",
			SetUpScript: []string{
				`CREATE TABLE diamond_edges (source INT, target INT);`,
				`INSERT INTO diamond_edges VALUES (1, 2), (1, 3), (2, 4), (3, 4);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM diamond_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;`,
					Expected: []sql.Row{
						{1, 0, "f", "{(1)}"},
						{2, 1, "f", "{(1),(2)}"},
						{3, 1, "f", "{(1),(3)}"},
						{4, 2, "f", "{(1),(2),(4)}"},
						{4, 2, "f", "{(1),(3),(4)}"},
					},
				},
			},
		},
		{
			Name: "A cyclic branch does not suppress an independent branch",
			SetUpScript: []string{
				`CREATE TABLE branch_edges (source INT, target INT);`,
				`INSERT INTO branch_edges VALUES (1, 2), (1, 3), (2, 4), (4, 2), (3, 5);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM branch_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node;`,
					Expected: []sql.Row{
						{1, 0, "f", "{(1)}"},
						{2, 1, "f", "{(1),(2)}"},
						{3, 1, "f", "{(1),(3)}"},
						{4, 2, "f", "{(1),(2),(4)}"},
						{5, 2, "f", "{(1),(3),(5)}"},
						{2, 3, "t", "{(1),(2),(4),(2)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE initializes independent paths for multiple anchor rows",
			Skip: true, // TODO: we hardcode the expectation of SELECT on the left side of the union
			SetUpScript: []string{
				`CREATE TABLE multi_anchor_edges (source INT, target INT);`,
				`INSERT INTO multi_anchor_edges VALUES (1, 2), (2, 1), (10, 11), (11, 10);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(root, node, depth) AS (
	VALUES (1, 1, 0), (10, 10, 0)
	UNION ALL
	SELECT w.root, e.target, w.depth + 1
	FROM multi_anchor_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT root, node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY root, depth;`,
					Expected: []sql.Row{
						{1, 1, 0, "f", "{(1)}"},
						{1, 2, 1, "f", "{(1),(2)}"},
						{1, 1, 2, "t", "{(1),(2),(1)}"},
						{10, 10, 0, "f", "{(10)}"},
						{10, 11, 1, "f", "{(10),(11)}"},
						{10, 10, 2, "t", "{(10),(11),(10)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE supports composite cycle keys",
			SetUpScript: []string{
				`CREATE TABLE composite_edges (
				from_namespace INT,
				from_node INT,
				to_namespace INT,
				to_node INT
			);`,
				`INSERT INTO composite_edges VALUES
				(1, 1, 2, 1),
				(2, 1, 2, 2),
				(2, 2, 1, 1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
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
ORDER BY depth;`,
					Expected: []sql.Row{
						{1, 1, 0, "f", "{\"(1,1)\"}"},
						{2, 1, 1, "f", "{\"(1,1)\",\"(2,1)\"}"},
						{2, 2, 2, "f", "{\"(1,1)\",\"(2,1)\",\"(2,2)\"}"},
						{1, 1, 3, "t", "{\"(1,1)\",\"(2,1)\",\"(2,2)\",\"(1,1)\"}"},
					},
				},
			},
		},
		{
			Name: "CYCLE treats repeated NULL cycle keys as equal",
			Skip: true, // TODO: https://github.com/dolthub/doltgresql/issues/2936
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(key, depth) AS (
	SELECT NULL::INT, 0
	UNION ALL
	SELECT NULL::INT, depth + 1 FROM walk
) CYCLE key SET is_cycle USING path
SELECT key, depth, is_cycle
FROM walk
ORDER BY depth;`,
					Expected: []sql.Row{
						{nil, 0, "f"},
						{nil, 1, "t"},
					},
				},
			},
		},
		{
			Name: "CYCLE supports custom text marker values",
			Skip: true, // TODO: we don't yet support CYCLE ... SET ... TO
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT n, depth + 1 FROM walk
) CYCLE n SET cycle_mark TO 'cycle' DEFAULT 'ok' USING path
SELECT n, depth, cycle_mark, pg_typeof(cycle_mark)::text AS marker_type
FROM walk
ORDER BY depth;`,
					Expected: []sql.Row{
						{1, 0, "ok", "text"},
						{1, 1, "cycle", "text"},
					},
				},
			},
		},
		{
			Name: "CYCLE supports custom integer marker values",
			Skip: true, // TODO: we don't yet support CYCLE ... SET ... TO
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT n, depth + 1 FROM walk
) CYCLE n SET cycle_mark TO 99 DEFAULT 0 USING path
SELECT n, depth, cycle_mark, pg_typeof(cycle_mark)::text AS marker_type
FROM walk
ORDER BY depth;`,
					Expected: []sql.Row{
						{1, 0, 0, "integer"},
						{1, 1, 99, "integer"},
					},
				},
			},
		},
		{
			Name: "Generated cycle marker is visible in the recursive term",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT n, depth + 1
	FROM walk
	WHERE NOT is_cycle
) CYCLE n SET is_cycle USING path
SELECT n, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth;`,
					Expected: []sql.Row{
						{1, 0, "f", "{(1)}"},
						{1, 1, "t", "{(1),(1)}"},
					},
				},
			},
		},
		{
			Name: "SELECT star expands generated cycle columns",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE n SET is_cycle USING path
SELECT * FROM walk;`,
					ExpectedColNames: []string{"n", "is_cycle", "path"},
					Expected: []sql.Row{
						{1, "f", "{(1)}"},
						{1, "t", "{(1),(1)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE works with UNION DISTINCT",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION
	SELECT n FROM walk
) CYCLE n SET is_cycle USING path
SELECT n, is_cycle, path::text AS path_text
FROM walk;`,
					Expected: []sql.Row{
						{1, "f", "{(1)}"},
						{1, "t", "{(1),(1)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE works with sibling and consuming CTEs",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
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
FROM cycles;`,
					Expected: []sql.Row{
						{1, 3, "{(1),(2),(3),(1)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE supports UUID cycle keys",
			SetUpScript: []string{
				`CREATE TABLE uuid_edges (source UUID, target UUID);`,
				`INSERT INTO uuid_edges VALUES
				('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002'),
				('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003'),
				('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(node_id, depth) AS (
	SELECT '00000000-0000-0000-0000-000000000001'::UUID, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM uuid_edges e
	JOIN walk w ON e.source = w.node_id
) CYCLE node_id SET is_cycle USING path
SELECT node_id::text, depth, is_cycle, pg_typeof(path)::text AS path_type
FROM walk
ORDER BY depth;`,
					Expected: []sql.Row{
						{"00000000-0000-0000-0000-000000000001", 0, "f", "record[]"},
						{"00000000-0000-0000-0000-000000000002", 1, "f", "record[]"},
						{"00000000-0000-0000-0000-000000000003", 2, "f", "record[]"},
						{"00000000-0000-0000-0000-000000000001", 3, "t", "record[]"},
					},
				},
			},
		},
		{
			Name: "CYCLE supports quoted identifiers",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
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
ORDER BY "Depth";`,
					ExpectedColNames: []string{"Node", "Depth", "IsCycle", "PathText"},
					Expected: []sql.Row{
						{1, 0, "f", "{(1)}"},
						{1, 1, "t", "{(1),(1)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE supports bind variables in the anchor term",
			SetUpScript: []string{
				`CREATE TABLE bound_edges (source INT, target INT);`,
				`INSERT INTO bound_edges VALUES (1, 2), (2, 3), (3, 1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(node, depth) AS (
	SELECT $1::INT, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM bound_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle
FROM walk
ORDER BY depth;`,
					BindVars: []any{1},
					Expected: []sql.Row{
						{1, 0, "f"},
						{2, 1, "f"},
						{3, 2, "f"},
						{1, 3, "t"},
					},
				},
			},
		},
		{
			Name: "CYCLE terminates a deep cycle exactly once",
			Skip: true, // TODO: we don't yet support FILTER
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n) AS (
	SELECT 0
	UNION ALL
	SELECT (n + 1) % 256 FROM walk
) CYCLE n SET is_cycle USING path
SELECT
	count(*)::INT AS row_count,
	count(*) FILTER (WHERE is_cycle)::INT AS cycle_count,
	max(array_length(path, 1)) AS max_path_length
FROM walk;`,
					Expected: []sql.Row{
						{257, 1, 257},
					},
				},
			},
		},
		{
			Name: "CYCLE works inside a view",
			SetUpScript: []string{
				`CREATE TABLE view_edges (source INT, target INT);`,
				`INSERT INTO view_edges VALUES (1, 2), (2, 1);`,
				`CREATE VIEW cycle_walk_view AS
WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM view_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM cycle_walk_view ORDER BY depth;`,
					Expected: []sql.Row{
						{1, 0, "f", "{(1)}"},
						{2, 1, "f", "{(1),(2)}"},
						{1, 2, "t", "{(1),(2),(1)}"},
					},
				},
			},
		},
		{
			Name: "CYCLE output can feed a data-modifying statement",
			SetUpScript: []string{
				`CREATE TABLE dml_edges (source INT, target INT);`,
				`INSERT INTO dml_edges VALUES (1, 2), (2, 1);`,
				`CREATE TABLE cycle_results (node INT, depth INT, is_cycle BOOLEAN);`,
				`WITH RECURSIVE walk(node, depth) AS (
	SELECT 1, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM dml_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
INSERT INTO cycle_results
SELECT node, depth, is_cycle FROM walk;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT node, depth, is_cycle FROM cycle_results ORDER BY depth;`,
					Expected: []sql.Row{
						{1, 0, "f"},
						{2, 1, "f"},
						{1, 2, "t"},
					},
				},
			},
		},
		{
			Name: "CYCLE rejects a non-recursive CTE",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n) AS (
	SELECT 1
) CYCLE n SET is_cycle USING path
SELECT * FROM walk;`,
					ExpectedErr: "WITH query is not recursive",
				},
			},
		},
		{
			Name: "CYCLE rejects an unknown cycle column",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE missing SET is_cycle USING path
SELECT * FROM walk;`,
					ExpectedErr: `cycle column "missing" not in WITH query column list`,
				},
			},
		},
		{
			Name: "CYCLE rejects duplicate cycle columns",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE n, n SET is_cycle USING path
SELECT * FROM walk;`,
					ExpectedErr: `cycle column "n" specified more than once`,
				},
			},
		},
		{
			Name: "CYCLE rejects a marker name already in the CTE output",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n, label) AS (
	SELECT 1, 'start'::TEXT
	UNION ALL
	SELECT n, label FROM walk
) CYCLE n SET label USING path
SELECT * FROM walk;`,
					ExpectedErr: `column reference "label" is ambiguous`,
				},
			},
		},
		{
			Name: "CYCLE rejects a path name already in the CTE output",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n, path) AS (
	SELECT 1, 'start'::TEXT
	UNION ALL
	SELECT n, path FROM walk
) CYCLE n SET is_cycle USING path
SELECT * FROM walk;`,
					ExpectedErr: `column reference "path" is ambiguous`,
				},
			},
		},
		{
			Name: "CYCLE rejects identical marker and path names",
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE n SET generated USING generated
SELECT * FROM walk;`,
					ExpectedErr: "cycle mark column name and cycle path column name are the same",
				},
			},
		},
		{
			Name: "CYCLE rejects incompatible marker and default types",
			Skip: true, // TODO: we don't yet support CYCLE ... SET ... TO
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(n) AS (
	SELECT 1
	UNION ALL
	SELECT n FROM walk
) CYCLE n SET is_cycle TO TRUE DEFAULT 55 USING path
SELECT * FROM walk;`,
					ExpectedErr: "CYCLE types boolean and integer cannot be matched",
				},
			},
		},
		{
			Name: "CYCLE requires a SELECT on the left side of the recursive UNION",
			SetUpScript: []string{
				`CREATE TABLE left_union_graph (source INT, target INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
WITH RECURSIVE walk(source, target) AS (
	SELECT * FROM left_union_graph
	UNION ALL
	SELECT * FROM left_union_graph
	UNION ALL
	SELECT g.*
	FROM left_union_graph g
	JOIN walk w ON g.source = w.target
) CYCLE source, target SET is_cycle USING path
SELECT * FROM walk;`,
					ExpectedErr: "with a SEARCH or CYCLE clause, the left side of the UNION must be a SELECT",
				},
			},
		},
		{
			Name: "CYCLE requires a SELECT on the right side of the recursive UNION",
			SetUpScript: []string{
				`CREATE TABLE right_union_graph (source INT, target INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
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
SELECT * FROM walk;`,
					ExpectedErr: "with a SEARCH or CYCLE clause, the right side of the UNION must be a SELECT",
				},
			},
		},
	})
}

func TestWithCycleGeneratedGraphs(t *testing.T) {
	RunScripts(t, makeWithCycleGeneratedGraphTests())
}

type generatedCycleEdge struct {
	from int
	to   int
}

type generatedCycleRow struct {
	node    int
	depth   int
	isCycle bool
	path    []int
}

func makeWithCycleGeneratedGraphTests() []ScriptTest {
	const (
		nodeCount = 5
		edgeCount = 7
	)

	seeds := []uint64{
		0x0000000000000001,
		0x0000000000000002,
		0x0000000000000003,
		0x0000000000000005,
		0x0000000000000008,
		0x000000000000000d,
		0x0000000000000015,
		0x0000000000000022,
		0x9e3779b97f4a7c15,
		0xd1b54a32d192ed03,
		0x62fec2eed91dc673,
		0x4301b2da9cc20ff8,
		0x10a0faf22c2365ff,
		0x89ebda7b349b3e38,
		0xa3421a539064428a,
		0xdeab883675571e4d,
		0x9040e9a76847ca6b,
		0x93e8341d44bc3cf3,
		0x95a36e8b834ccce4,
		0xd0ee03218946e4fe,
		0x48289f16eec514ff,
		0xe7124832cd5a455c,
		0xf0a835222287a683,
		0x122a2f3989dff1b4,
		0x0ac7bdf3d90af2fd,
		0xf10de05685aeabbb,
		0xbfa07d1a2e0d0363,
		0x11fc4c7494796cf1,
		0x7057fbbb0a9ee37e,
		0x092fdd61d4fdf5f9,
		0x1687ced1070d7f8b,
		0x2bcff8b03d78be9b,
		0xbb1477a73b25a81f,
		0xc6585c83de531aeb,
		0x2db6991b0eb906e8,
		0xed59257df91e6b2c,
		0x7887be1133a582f0,
		0x804a8876f0f601f5,
		0xfbe23113e3f39920,
		0x79c9eda27f039a47,
		0xebee2d9ccd7c7521,
	}

	tests := make([]ScriptTest, 0, len(seeds))
	for i, seed := range seeds {
		edges := makeGeneratedCycleEdges(seed, nodeCount, edgeCount)
		root := int(seed%nodeCount) + 1
		expectedRows := evaluateGeneratedCycleGraph(root, edges)

		// Keep accidental future changes to the generator from creating an unreasonably large integration test
		if len(expectedRows) > 500 {
			continue
		}

		valueStrings := make([]string, len(edges))
		for j, edge := range edges {
			valueStrings[j] = fmt.Sprintf("(%d, %d)", edge.from, edge.to)
		}

		expected := make([]sql.Row, len(expectedRows))
		for j, row := range expectedRows {
			marker := "f"
			if row.isCycle {
				marker = "t"
			}
			expected[j] = sql.Row{row.node, row.depth, marker, generatedCyclePathText(row.path)}
		}

		tests = append(tests, ScriptTest{
			Name: fmt.Sprintf("generated CYCLE graph %02d seed %016x", i+1, seed),
			SetUpScript: []string{
				`CREATE TABLE generated_cycle_edges (source INT, target INT);`,
				fmt.Sprintf("INSERT INTO generated_cycle_edges VALUES %s;", strings.Join(valueStrings, ", ")),
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: fmt.Sprintf(`
WITH RECURSIVE walk(node, depth) AS (
	SELECT %d, 0
	UNION ALL
	SELECT e.target, w.depth + 1
	FROM generated_cycle_edges e
	JOIN walk w ON e.source = w.node
) CYCLE node SET is_cycle USING path
SELECT node, depth, is_cycle, path::text AS path_text
FROM walk
ORDER BY depth, node, path::text;`, root),
					Expected: expected,
				},
			},
		})
	}

	return tests
}

func makeGeneratedCycleEdges(seed uint64, nodeCount int, edgeCount int) []generatedCycleEdge {
	edges := []generatedCycleEdge{{1, 2}, {2, 3}, {3, 1}}
	seen := map[generatedCycleEdge]struct{}{
		edges[0]: {},
		edges[1]: {},
		edges[2]: {},
	}

	state := seed
	for len(edges) < edgeCount {
		from := int(nextGeneratedCycleValue(&state)%uint64(nodeCount)) + 1
		to := int(nextGeneratedCycleValue(&state)%uint64(nodeCount)) + 1
		edge := generatedCycleEdge{from: from, to: to}
		if _, ok := seen[edge]; ok {
			continue
		}
		seen[edge] = struct{}{}
		edges = append(edges, edge)
	}

	return edges
}

func nextGeneratedCycleValue(state *uint64) uint64 {
	*state = *state*6364136223846793005 + 1442695040888963407
	return *state
}

func evaluateGeneratedCycleGraph(root int, edges []generatedCycleEdge) []generatedCycleRow {
	pending := []generatedCycleRow{{node: root, path: []int{root}}}
	rows := make([]generatedCycleRow, 0)

	for len(pending) > 0 {
		row := pending[0]
		pending = pending[1:]
		rows = append(rows, row)

		if row.isCycle {
			continue
		}

		for _, edge := range edges {
			if edge.from != row.node {
				continue
			}

			path := append(append([]int(nil), row.path...), edge.to)
			pending = append(pending, generatedCycleRow{
				node:    edge.to,
				depth:   row.depth + 1,
				isCycle: generatedCyclePathContains(row.path, edge.to),
				path:    path,
			})
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].depth != rows[j].depth {
			return rows[i].depth < rows[j].depth
		}
		if rows[i].node != rows[j].node {
			return rows[i].node < rows[j].node
		}
		return generatedCyclePathText(rows[i].path) < generatedCyclePathText(rows[j].path)
	})

	return rows
}

func generatedCyclePathContains(path []int, value int) bool {
	for _, pathValue := range path {
		if pathValue == value {
			return true
		}
	}
	return false
}

func generatedCyclePathText(path []int) string {
	values := make([]string, len(path))
	for i, value := range path {
		values[i] = fmt.Sprintf("(%d)", value)
	}
	return "{" + strings.Join(values, ",") + "}"
}
