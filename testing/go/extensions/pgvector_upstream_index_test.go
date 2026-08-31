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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"

	framework "github.com/dolthub/doltgresql/testing/go"
)

func TestPgvectorUpstreamIndex(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "vector indexes require a primary key",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE t (val vector(3));",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					// Dolt's vector index implementation currently requires keyed tables
					Query:       "CREATE INDEX ON t USING hnsw (val vector_l2_ops);",
					ExpectedErr: "vector indexes on keyless tables are not supported",
				},
			},
		},
		{
			Name: "hnsw vector",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				// L2
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING hnsw (val vector_l2_ops);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <-> (SELECT NULL::vector)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "SELECT COUNT(*) FROM t;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "TRUNCATE t;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
				// inner product
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING hnsw (val vector_ip_ops);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <#> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[1,2,4]"}, {"[1,2,3]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <#> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,2,4]"}, {"[1,2,3]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <#> (SELECT NULL::vector)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
				// cosine
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING hnsw (val vector_cosine_ops);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <=> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[0,0,0]"}, {"[1,1,1]"}, {"[1,2,3]"}, {"[1,2,4]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <=> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[0,0,0]"}, {"[1,1,1]"}, {"[1,2,3]"}, {"[1,2,4]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <=> '[0,0,0]') t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <=> (SELECT NULL::vector)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query: "SELECT t.val, t2.val FROM t CROSS JOIN LATERAL (SELECT val FROM t t3 ORDER BY val <=> t.val LIMIT 1) t2 WHERE t.val != '[0,0,0]' ORDER BY t.val;",
					Expected: []sql.Row{
						{"[1,1,1]", nil}, {"[1,2,3]", nil}, {"[1,2,4]", nil},
					},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
				// L1
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING hnsw (val vector_l1_ops);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <+> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <+> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <+> (SELECT NULL::vector)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "hnsw halfvec",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				// L2
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING hnsw (val halfvec_l2_ops);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <-> (SELECT NULL::halfvec)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "SELECT COUNT(*) FROM t;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "TRUNCATE t;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
				// inner product
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING hnsw (val halfvec_ip_ops);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <#> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[1,2,4]"}, {"[1,2,3]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <#> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,2,4]"}, {"[1,2,3]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <#> (SELECT NULL::halfvec)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
				// cosine
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING hnsw (val halfvec_cosine_ops);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <=> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[0,0,0]"}, {"[1,1,1]"}, {"[1,2,3]"}, {"[1,2,4]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <=> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,1,1]"}, {"[0,0,0]"}, {"[1,2,3]"}, {"[1,2,4]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <=> '[0,0,0]') t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <=> (SELECT NULL::halfvec)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
				// L1
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING hnsw (val halfvec_l1_ops);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <+> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <+> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <+> (SELECT NULL::halfvec)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "ivfflat vector",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				// L2
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING ivfflat (val vector_l2_ops) WITH (lists = 1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <-> (SELECT NULL::vector)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "SELECT COUNT(*) FROM t;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "TRUNCATE t;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "ivfflat halfvec",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(3));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (1, '[0,0,0]'), (2, '[1,2,3]'), (3, '[1,1,1]'), (4, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX ON t USING ivfflat (val halfvec_l2_ops) WITH (lists = 1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t VALUES (5, '[1,2,4]');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]';",
					Expected: []sql.Row{
						{nil}, {"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query: "SELECT val FROM t ORDER BY val <-> '[3,3,3]' LIMIT 4;",
					Expected: []sql.Row{
						{"[1,2,3]"}, {"[1,2,4]"}, {"[1,1,1]"}, {"[0,0,0]"},
					},
				},
				{
					Query:    "SELECT COUNT(*) FROM (SELECT val FROM t ORDER BY val <-> (SELECT NULL::halfvec)) t2;",
					Expected: []sql.Row{{5}},
				},
				{
					Query:    "DROP TABLE t;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
