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

func TestPgvectorIndexDDL(t *testing.T) {
	setup := []string{
		"CREATE EXTENSION vector;",
		"CREATE TABLE tv (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL, s sparsevec(3) NOT NULL, t text NOT NULL, vnodim vector NOT NULL, big vector(2001) NOT NULL, hbig halfvec(4001) NOT NULL);",
	}
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name:        "every operator class creates an index under both methods",
			SetUpScript: setup,
			Assertions: []framework.ScriptTestAssertion{
				{Query: "CREATE INDEX idx1 ON tv USING hnsw (v vector_l2_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx2 ON tv USING hnsw (v vector_ip_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx3 ON tv USING hnsw (v vector_cosine_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx4 ON tv USING hnsw (v vector_l1_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx5 ON tv USING ivfflat (v vector_l2_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx6 ON tv USING ivfflat (v vector_ip_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx7 ON tv USING ivfflat (v vector_cosine_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx8 ON tv USING hnsw (h halfvec_l2_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx9 ON tv USING hnsw (h halfvec_ip_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx10 ON tv USING hnsw (h halfvec_cosine_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx11 ON tv USING hnsw (h halfvec_l1_ops);", Expected: []sql.Row{}},
				{Query: "CREATE INDEX idx12 ON tv USING ivfflat (h halfvec_l2_ops);", Expected: []sql.Row{}},
				{
					// pgvector offers the L1 classes under hnsw only, but both emulated methods map to the same index.
					Query:    "CREATE INDEX idx13 ON tv USING ivfflat (v vector_l1_ops);",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name:        "default operator classes",
			SetUpScript: setup,
			Assertions: []framework.ScriptTestAssertion{
				{
					// vector_l2_ops is the default operator class for vector under ivfflat.
					Query:    "CREATE INDEX ON tv USING ivfflat (v);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v);",
					ExpectedErr: `data type vector has no default operator class for access method "hnsw"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING ivfflat (h);",
					ExpectedErr: `data type halfvec has no default operator class for access method "ivfflat"`,
				},
			},
		},
		{
			Name:        "operator class errors",
			SetUpScript: setup,
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v foo_ops);",
					ExpectedErr: `operator class "foo_ops" does not exist for access method "hnsw"`,
				},
				{
					// sparsevec opclasses exist under hnsw only in pgvector.
					Query:       "CREATE INDEX ON tv USING ivfflat (s sparsevec_l2_ops);",
					ExpectedErr: `operator class "sparsevec_l2_ops" does not exist for access method "ivfflat"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (s sparsevec_l2_ops);",
					ExpectedErr: `operator class "sparsevec_l2_ops" is not yet supported`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (h vector_l2_ops);",
					ExpectedErr: `operator class "vector_l2_ops" does not accept data type halfvec`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (t vector_l2_ops);",
					ExpectedErr: `operator class "vector_l2_ops" does not accept data type text`,
				},
			},
		},
		{
			Name:        "index restrictions",
			SetUpScript: setup,
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:       "CREATE UNIQUE INDEX ON tv USING hnsw (v vector_l2_ops);",
					ExpectedErr: `access method "hnsw" does not support unique indexes`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) INCLUDE (id);",
					ExpectedErr: `access method "hnsw" does not support included columns`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops, h halfvec_l2_ops);",
					ExpectedErr: `access method "hnsw" does not support multicolumn indexes`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WHERE id > 3;",
					ExpectedErr: "partial vector indexes are not yet supported",
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw ((l2_normalize(v)) vector_l2_ops);",
					ExpectedErr: "expression columns in vector indexes are not yet supported",
				},
				{
					Query:       "CREATE INDEX ON tv USING gin (v);",
					ExpectedErr: "index method gin is not yet supported",
				},
			},
		},
		{
			Name:        "column dimension requirements",
			SetUpScript: setup,
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:       "CREATE INDEX ON tv USING hnsw (vnodim vector_l2_ops);",
					ExpectedErr: "column does not have dimensions",
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (big vector_l2_ops);",
					ExpectedErr: "column cannot have more than 2000 dimensions for hnsw index",
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (hbig halfvec_l2_ops);",
					ExpectedErr: "column cannot have more than 4000 dimensions for hnsw index",
				},
				{
					Query:       "CREATE INDEX ON tv USING ivfflat (big vector_l2_ops);",
					ExpectedErr: "column cannot have more than 2000 dimensions for ivfflat index",
				},
				{
					Query:       "CREATE INDEX ON tv USING ivfflat (hbig halfvec_l2_ops);",
					ExpectedErr: "column cannot have more than 4000 dimensions for ivfflat index",
				},
			},
		},
		{
			Name:        "storage parameters",
			SetUpScript: setup,
			Assertions: []framework.ScriptTestAssertion{
				{
					// The parameter values are validated, then discarded: the underlying index has no equivalent knobs.
					Query:    "CREATE INDEX idx1 ON tv USING hnsw (v vector_l2_ops) WITH (m = 16, ef_construction = 64);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE INDEX idx2 ON tv USING ivfflat (v vector_l2_ops) WITH (lists = 100);",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 1);",
					ExpectedErr: `value 1 out of bounds for option "m"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 101);",
					ExpectedErr: `value 101 out of bounds for option "m"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (ef_construction = 3);",
					ExpectedErr: `value 3 out of bounds for option "ef_construction"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (ef_construction = 1001);",
					ExpectedErr: `value 1001 out of bounds for option "ef_construction"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 40, ef_construction = 64);",
					ExpectedErr: "ef_construction must be greater than or equal to 2 * m",
				},
				{
					// The default ef_construction of 64 applies when only m is given.
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 40);",
					ExpectedErr: "ef_construction must be greater than or equal to 2 * m",
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (foo = 1);",
					ExpectedErr: `unrecognized parameter "foo"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (lists = 100);",
					ExpectedErr: `unrecognized parameter "lists"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING ivfflat (v vector_l2_ops) WITH (m = 16);",
					ExpectedErr: `unrecognized parameter "m"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING ivfflat (v vector_l2_ops) WITH (lists = 0);",
					ExpectedErr: `value 0 out of bounds for option "lists"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING ivfflat (v vector_l2_ops) WITH (lists = 32769);",
					ExpectedErr: `value 32769 out of bounds for option "lists"`,
				},
				{
					Query:       "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 'abc');",
					ExpectedErr: `invalid value for integer option "m": abc`,
				},
			},
		},
		{
			Name: "index maintenance",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE items (id INT4 PRIMARY KEY, v vector(3) NOT NULL);",
				"INSERT INTO items VALUES (1, '[1,1,1]'), (2, '[2,2,2]'), (3, '[3,3,3]');",
				"CREATE INDEX items_idx ON items USING hnsw (v vector_cosine_ops);",
				"INSERT INTO items VALUES (4, '[4,4,4]');",
				"UPDATE items SET v = '[9,9,9]' WHERE id = 2;",
				"DELETE FROM items WHERE id = 1;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT id, v FROM items ORDER BY id;",
					Expected: []sql.Row{{2, "[9,9,9]"}, {3, "[3,3,3]"}, {4, "[4,4,4]"}},
				},
				{
					Query:    "SELECT id FROM items WHERE v = '[9,9,9]';",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "index maintenance with out-of-line vectors",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE items (id INT4 PRIMARY KEY, v vector(700) NOT NULL);",
				"INSERT INTO items VALUES (1, '" + bigDenseLiteral(700, 0) + "'), (2, '" + bigDenseLiteral(700, 1) + "');",
				"CREATE INDEX items_idx ON items USING hnsw (v vector_l2_ops);",
				"INSERT INTO items VALUES (3, '" + bigDenseLiteral(700, 2) + "');",
				"UPDATE items SET v = '" + bigDenseLiteral(700, 3) + "' WHERE id = 1;",
				"DELETE FROM items WHERE id = 2;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT id, v FROM items ORDER BY id;",
					Expected: []sql.Row{{1, bigDenseLiteral(700, 3)}, {3, bigDenseLiteral(700, 2)}},
				},
			},
		},
	})
}
