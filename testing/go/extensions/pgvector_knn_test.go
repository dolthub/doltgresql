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

package extensions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"

	framework "github.com/dolthub/doltgresql/testing/go"
)

// knnBigLiteral returns the vector literal for row i of the multi-thousand-row recall table. The
// first element is i itself, so every row's vector is distinct.
func knnBigLiteral(i int) string {
	return fmt.Sprintf("[%d,%d,%d,%d,%d,%d,%d,%d]",
		i, (i*3)%50, (i*7)%50, (i*11)%50, (i*13)%50, (i*17)%50, (i*19)%50, (i*23)%50)
}

// knnBigInsert returns a single INSERT statement covering rows 1 through count of the recall table.
func knnBigInsert(count int) string {
	sb := strings.Builder{}
	sb.WriteString("INSERT INTO big VALUES ")
	for i := 1; i <= count; i++ {
		if i > 1 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("(%d, '%s')", i, knnBigLiteral(i)))
	}
	sb.WriteString(";")
	return sb.String()
}

func TestPgvectorKnn(t *testing.T) {
	setup := []string{
		"CREATE EXTENSION vector;",
		"CREATE TABLE knn (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
		`INSERT INTO knn VALUES
			(1, '[1.2,0.3,2.1]', '[1.2,0.3,2.1]'),
			(2, '[2.5,1.1,0.4]', '[2.5,1.1,0.4]'),
			(3, '[0.9,2.2,1.7]', '[0.9,2.2,1.7]'),
			(4, '[-1.4,2.8,0.6]', '[-1.4,2.8,0.6]'),
			(5, '[3.1,-0.7,1.9]', '[3.1,-0.7,1.9]'),
			(6, '[0.2,0.8,-1.3]', '[0.2,0.8,-1.3]'),
			(7, '[4.6,3.2,2.4]', '[4.6,3.2,2.4]'),
			(8, '[-2.2,-1.1,3.3]', '[-2.2,-1.1,3.3]'),
			(9, '[1.8,1.6,0.9]', '[1.8,1.6,0.9]'),
			(10, '[0.4,-2.6,2.2]', '[0.4,-2.6,2.2]');`,
	}
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "each vector metric is served by its matching index",
			SetUpScript: append(setup,
				"CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
				"CREATE INDEX idx_ip ON knn USING hnsw (v vector_ip_ops);",
				"CREATE INDEX idx_cos ON knn USING hnsw (v vector_cosine_ops);",
				"CREATE INDEX idx_l1 ON knn USING hnsw (v vector_l1_ops);",
			),
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{
						{"Limit(4)"},
						{" └─ Project"},
						{"     ├─ columns: [knn.id]"},
						{"     └─ IndexedTableAccess(knn)"},
						{"         ├─ index: [knn.v]"},
						{"         ├─ order: knn.v <-> '[1.5,1,2]' LIMIT 4 (bigint)"},
						{"         └─ columns: [id v]"},
					},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 10;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}, {5}, {6}, {4}, {10}, {7}, {8}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <#> '[1.5,1,2]' LIMIT 10;",
					Expected: []sql.Row{{7}, {5}, {3}, {1}, {9}, {2}, {10}, {8}, {4}, {6}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <=> '[1.5,1,2]' LIMIT 10;",
					Expected: []sql.Row{{1}, {7}, {3}, {9}, {5}, {2}, {10}, {4}, {8}, {6}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <+> '[1.5,1,2]' LIMIT 10;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}, {5}, {6}, {10}, {7}, {4}, {8}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY '[1.5,1,2]' <-> v LIMIT 4;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]'::vector LIMIT 4;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4 OFFSET 3;",
					Expected: []sql.Row{{2}, {5}, {6}, {4}},
				},
			},
		},
		{
			Name: "each halfvec metric is served by its matching index",
			SetUpScript: append(setup,
				"CREATE INDEX idx_l2 ON knn USING hnsw (h halfvec_l2_ops);",
				"CREATE INDEX idx_ip ON knn USING hnsw (h halfvec_ip_ops);",
				"CREATE INDEX idx_cos ON knn USING hnsw (h halfvec_cosine_ops);",
				"CREATE INDEX idx_l1 ON knn USING hnsw (h halfvec_l1_ops);",
			),
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT id FROM knn ORDER BY h <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{
						{"Limit(4)"},
						{" └─ Project"},
						{"     ├─ columns: [knn.id]"},
						{"     └─ IndexedTableAccess(knn)"},
						{"         ├─ index: [knn.h]"},
						{"         ├─ order: knn.h <-> '[1.5,1,2]' LIMIT 4 (bigint)"},
						{"         └─ columns: [id h]"},
					},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY h <-> '[1.5,1,2]' LIMIT 10;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}, {5}, {6}, {4}, {10}, {7}, {8}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY h <#> '[1.5,1,2]' LIMIT 10;",
					Expected: []sql.Row{{7}, {5}, {3}, {1}, {9}, {2}, {10}, {8}, {4}, {6}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY h <=> '[1.5,1,2]' LIMIT 10;",
					Expected: []sql.Row{{1}, {7}, {3}, {9}, {5}, {2}, {10}, {4}, {8}, {6}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY h <+> '[1.5,1,2]' LIMIT 10;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}, {5}, {6}, {10}, {7}, {4}, {8}},
				},
			},
		},
		{
			Name: "ivfflat serves knn through the same native index",
			SetUpScript: append(setup,
				"CREATE INDEX ON knn USING ivfflat (v);",
			),
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: "EXPLAIN SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{
						{"Limit(4)"},
						{" └─ Project"},
						{"     ├─ columns: [knn.id]"},
						{"     └─ IndexedTableAccess(knn)"},
						{"         ├─ index: [knn.v]"},
						{"         ├─ order: knn.v <-> '[1.5,1,2]' LIMIT 4 (bigint)"},
						{"         └─ columns: [id v]"},
					},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}},
				},
			},
		},
		{
			Name: "queries that cannot use the index fall back to an exact scan",
			SetUpScript: append(setup,
				"CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
			),
			Assertions: []framework.ScriptTestAssertion{
				{
					// A vector index only orders nearest-first
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' DESC LIMIT 4;",
					Expected: []sql.Row{{8}, {7}, {10}, {4}},
				},
				{
					// A filter below the limit means the index's top rows could be filtered out
					Query:    "SELECT id FROM knn WHERE id > 3 ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{{9}, {5}, {6}, {4}},
				},
				{
					// The function-call form is not an operator expression, so the index is not matched
					Query: "EXPLAIN SELECT id FROM knn ORDER BY l2_distance(v, '[1.5,1,2]') LIMIT 4;",
					Expected: []sql.Row{
						{"Limit(4)"},
						{" └─ Project"},
						{"     ├─ columns: [knn.id]"},
						{"     └─ TopN(Limit: [4]; l2_distance(knn.v, '[1.5,1,2]') ASC)"},
						{"         └─ Table"},
						{"             ├─ name: knn"},
						{"             └─ columns: [id v]"},
					},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY l2_distance(v, '[1.5,1,2]') LIMIT 4;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}},
				},
				{
					// There is no cosine index, so this exact scan must still order correctly
					Query:    "SELECT id FROM knn ORDER BY v <=> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{{1}, {7}, {3}, {9}},
				},
				{
					// A NULL query vector makes every distance NULL, so the row order is arbitrary
					Query:    "SELECT count(*) FROM (SELECT id FROM knn ORDER BY v <-> NULL LIMIT 4) sq;",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT count(*) FROM (SELECT id FROM knn ORDER BY v <-> NULL::vector LIMIT 4) sq;",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> NULL, id LIMIT 4;",
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
				{
					// Column-to-column distances depend on the row, so the index cannot serve them
					Query:    "SELECT count(*) FROM (SELECT id FROM knn ORDER BY v <-> v LIMIT 4) sq;",
					Expected: []sql.Row{{4}},
				},
			},
		},
		{
			Name: "bound parameter query vectors use the index",
			SetUpScript: append(setup,
				"CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
			),
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> $1 LIMIT 4;",
					BindVars: []any{"[1.5,1,2]"},
					Expected: []sql.Row{{1}, {9}, {3}, {2}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> $1 LIMIT 10;",
					BindVars: []any{"[1.5,1,2]"},
					Expected: []sql.Row{{1}, {9}, {3}, {2}, {5}, {6}, {4}, {10}, {7}, {8}},
				},
			},
		},
		{
			Name: "index maintenance is reflected in knn results",
			SetUpScript: append(setup,
				"CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
			),
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "INSERT INTO knn VALUES (11, '[1.5,1,2]', '[1.5,1,2]');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{{11}, {1}, {9}, {3}},
				},
				{
					Query:    "UPDATE knn SET v = '[100,100,100]', h = '[100,100,100]' WHERE id = 11;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' DESC LIMIT 1;",
					Expected: []sql.Row{{11}},
				},
				{
					Query:    "DELETE FROM knn WHERE id = 11;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 4;",
					Expected: []sql.Row{{1}, {9}, {3}, {2}},
				},
			},
		},
		{
			Name: "knn across branches, merges, and AS OF",
			SetUpScript: append(setup,
				"CREATE INDEX idx_l2 ON knn USING hnsw (v vector_l2_ops);",
				"SELECT dolt_commit('-Am', 'base');",
				"SELECT dolt_branch('other');",
				"INSERT INTO knn VALUES (11, '[1.5,1,2]', '[1.5,1,2]');",
				"SELECT dolt_commit('-am', 'main adds 11');",
				"SELECT dolt_checkout('other');",
				"INSERT INTO knn VALUES (12, '[1.4,1,2]', '[1.4,1,2]');",
				"SELECT dolt_commit('-am', 'other adds 12');",
				"SELECT dolt_checkout('main');",
				"SELECT dolt_merge('other');",
			),
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT id FROM knn ORDER BY v <-> '[1.5,1,2]' LIMIT 3;",
					Expected: []sql.Row{{11}, {12}, {1}},
				},
				{
					Query:    "SELECT id FROM knn AS OF 'HEAD~2' ORDER BY v <-> '[1.5,1,2]' LIMIT 3;",
					Expected: []sql.Row{{1}, {9}, {3}},
				},
				{
					Query:    "SELECT id FROM knn AS OF 'other' ORDER BY v <-> '[1.5,1,2]' LIMIT 3;",
					Expected: []sql.Row{{12}, {1}, {9}},
				},
			},
		},
		{
			Name: "nullable vector columns",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE nulls (id INT4 PRIMARY KEY, v vector(3));",
				"INSERT INTO nulls VALUES (1, '[1.2,0.3,2.1]'), (2, NULL), (3, '[0.9,2.2,1.7]'), (4, NULL), (5, '[3.1,-0.7,1.9]');",
				"CREATE INDEX idx_l2 ON nulls USING hnsw (v vector_l2_ops);",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					// NULL vectors are not indexed, so the index scan returns fewer rows than the limit
					Query:    "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{1}, {3}, {5}},
				},
				{
					// The exact scan still includes the NULL rows. Postgres orders them last
					// ({1}, {3}, {5}, {2}, {4}) while Doltgres currently orders all NULLs first.
					Query:    "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]', id;",
					Expected: []sql.Row{{2}, {4}, {1}, {3}, {5}},
				},
				{
					Query:    "UPDATE nulls SET v = '[1.5,1,2]' WHERE id = 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{2}, {1}, {3}, {5}},
				},
				{
					Query:    "UPDATE nulls SET v = NULL WHERE id = 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{2}, {3}, {5}},
				},
				{
					Query:    "DELETE FROM nulls WHERE id = 4;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{2}, {3}, {5}},
				},
			},
		},
		{
			Name: "nullable vector columns across branches, merges, and AS OF",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE nulls (id INT4 PRIMARY KEY, v vector(3));",
				"INSERT INTO nulls VALUES (1, '[1.2,0.3,2.1]'), (2, NULL), (3, '[0.9,2.2,1.7]'), (4, NULL), (5, '[3.1,-0.7,1.9]');",
				"CREATE INDEX idx_l2 ON nulls USING hnsw (v vector_l2_ops);",
				"SELECT dolt_commit('-Am', 'base');",
				"SELECT dolt_branch('other');",
				"UPDATE nulls SET v = '[1.5,1,2]' WHERE id = 2;",
				"SELECT dolt_commit('-am', 'main fills 2');",
				"SELECT dolt_checkout('other');",
				"UPDATE nulls SET v = NULL WHERE id = 5;",
				"SELECT dolt_commit('-am', 'other nulls 5');",
				"SELECT dolt_checkout('main');",
				"SELECT dolt_merge('other');",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					// The merge combines a NULL-to-value change with a value-to-NULL change
					Query:    "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{2}, {1}, {3}},
				},
				{
					// The exact scan proves the merged rows 4 and 5 are NULL rather than missing
					Query:    "SELECT id FROM nulls ORDER BY v <-> '[1.5,1,2]', id;",
					Expected: []sql.Row{{4}, {5}, {2}, {1}, {3}},
				},
				{
					Query:    "SELECT id FROM nulls AS OF 'HEAD~2' ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{1}, {3}, {5}},
				},
				{
					Query:    "SELECT id FROM nulls AS OF 'other' ORDER BY v <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{1}, {3}},
				},
			},
		},
		{
			Name: "nullable halfvec columns",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE hnulls (id INT4 PRIMARY KEY, h halfvec(3));",
				"INSERT INTO hnulls VALUES (1, '[1.2,0.3,2.1]'), (2, NULL), (3, '[0.9,2.2,1.7]'), (4, NULL), (5, '[3.1,-0.7,1.9]');",
				"CREATE INDEX idx_l2 ON hnulls USING hnsw (h halfvec_l2_ops);",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					// NULL halfvecs are not indexed, so the index scan returns fewer rows than the limit
					Query:    "SELECT id FROM hnulls ORDER BY h <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{1}, {3}, {5}},
				},
				{
					Query:    "UPDATE hnulls SET h = '[1.5,1,2]' WHERE id = 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id FROM hnulls ORDER BY h <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{2}, {1}, {3}, {5}},
				},
				{
					Query:    "UPDATE hnulls SET h = NULL WHERE id = 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id FROM hnulls ORDER BY h <-> '[1.5,1,2]' LIMIT 5;",
					Expected: []sql.Row{{2}, {3}, {5}},
				},
			},
		},
		{
			Name: "recall sanity on a multi-thousand-row table",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE big (id INT4 PRIMARY KEY, v vector(8) NOT NULL);",
				knnBigInsert(2000),
				"CREATE INDEX ON big USING hnsw (v vector_l2_ops);",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    fmt.Sprintf("SELECT id FROM big ORDER BY v <-> '%s' LIMIT 1;", knnBigLiteral(1234)),
					Expected: []sql.Row{{1234}},
				},
				{
					// pgvector's hnsw scan caps its results at ef_search (default 40) until the knob is
					// raised while the native index always returns the full limit
					Query:    fmt.Sprintf("SELECT count(*) FROM (SELECT id FROM big ORDER BY v <-> '%s' LIMIT 100) sq;", knnBigLiteral(1234)),
					Expected: []sql.Row{{100}},
				},
				{
					// pgvector's recall knob is accepted as a custom parameter, although its value is unused
					Query:    "SET hnsw.ef_search = 100;",
					Expected: []sql.Row{},
				},
				{
					Query:    fmt.Sprintf("SELECT id FROM big ORDER BY v <-> '%s' LIMIT 1;", knnBigLiteral(1234)),
					Expected: []sql.Row{{1234}},
				},
			},
		},
	})
}
