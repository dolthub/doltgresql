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

func TestPgvectorCatalog(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "pg_am lists the extension access methods once installed",
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: "SELECT amname, amtype FROM pg_catalog.pg_am ORDER BY amname;",
					Expected: []sql.Row{
						{"brin", "i"}, {"btree", "i"}, {"gin", "i"}, {"gist", "i"}, {"hash", "i"},
						{"heap", "t"}, {"spgist", "i"},
					},
				},
				{
					Query:    "CREATE EXTENSION vector;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT amname, amtype FROM pg_catalog.pg_am ORDER BY amname;",
					Expected: []sql.Row{
						{"brin", "i"}, {"btree", "i"}, {"gin", "i"}, {"gist", "i"}, {"hash", "i"},
						{"heap", "t"}, {"hnsw", "i"}, {"ivfflat", "i"}, {"spgist", "i"},
					},
				},
				{
					Query: "SELECT amname, amhandler::text FROM pg_catalog.pg_am WHERE amname IN ('hnsw', 'ivfflat') ORDER BY amname;",
					Expected: []sql.Row{
						{"hnsw", "hnswhandler"},
						{"ivfflat", "ivfflathandler"},
					},
				},
			},
		},
		{
			Name: "pg_opclass lists the extension operator classes once installed",
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT count(*) FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname IN ('hnsw', 'ivfflat');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "CREATE EXTENSION vector;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT o.opcname, a.amname, o.opcdefault FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname = 'hnsw' ORDER BY o.opcname;",
					Expected: []sql.Row{
						{"bit_hamming_ops", "hnsw", "f"},
						{"bit_jaccard_ops", "hnsw", "f"},
						{"halfvec_cosine_ops", "hnsw", "f"},
						{"halfvec_ip_ops", "hnsw", "f"},
						{"halfvec_l1_ops", "hnsw", "f"},
						{"halfvec_l2_ops", "hnsw", "f"},
						{"sparsevec_cosine_ops", "hnsw", "f"},
						{"sparsevec_ip_ops", "hnsw", "f"},
						{"sparsevec_l1_ops", "hnsw", "f"},
						{"sparsevec_l2_ops", "hnsw", "f"},
						{"vector_cosine_ops", "hnsw", "f"},
						{"vector_ip_ops", "hnsw", "f"},
						{"vector_l1_ops", "hnsw", "f"},
						{"vector_l2_ops", "hnsw", "f"},
					},
				},
				{
					// pgvector does not offer the L1 classes under ivfflat, but both emulated methods map
					// to the same index, so they appear under both.
					Query: "SELECT o.opcname, a.amname, o.opcdefault FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname = 'ivfflat' ORDER BY o.opcname;",
					Expected: []sql.Row{
						{"bit_hamming_ops", "ivfflat", "f"},
						{"halfvec_cosine_ops", "ivfflat", "f"},
						{"halfvec_ip_ops", "ivfflat", "f"},
						{"halfvec_l1_ops", "ivfflat", "f"},
						{"halfvec_l2_ops", "ivfflat", "f"},
						{"vector_cosine_ops", "ivfflat", "f"},
						{"vector_ip_ops", "ivfflat", "f"},
						{"vector_l1_ops", "ivfflat", "f"},
						{"vector_l2_ops", "ivfflat", "t"},
					},
				},
				{
					Query: "SELECT o.opcname, t.typname FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_type t ON t.oid = o.opcintype JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname = 'hnsw' ORDER BY o.opcname;",
					Expected: []sql.Row{
						{"bit_hamming_ops", "bit"},
						{"bit_jaccard_ops", "bit"},
						{"halfvec_cosine_ops", "halfvec"},
						{"halfvec_ip_ops", "halfvec"},
						{"halfvec_l1_ops", "halfvec"},
						{"halfvec_l2_ops", "halfvec"},
						{"sparsevec_cosine_ops", "sparsevec"},
						{"sparsevec_ip_ops", "sparsevec"},
						{"sparsevec_l1_ops", "sparsevec"},
						{"sparsevec_l2_ops", "sparsevec"},
						{"vector_cosine_ops", "vector"},
						{"vector_ip_ops", "vector"},
						{"vector_l1_ops", "vector"},
						{"vector_l2_ops", "vector"},
					},
				},
				{
					Query: "SELECT o.opcname, n.nspname FROM pg_catalog.pg_opclass o JOIN pg_catalog.pg_namespace n ON n.oid = o.opcnamespace JOIN pg_catalog.pg_am a ON a.oid = o.opcmethod WHERE a.amname = 'hnsw' AND o.opcname LIKE 'vector%' ORDER BY o.opcname;",
					Expected: []sql.Row{
						{"vector_cosine_ops", "public"},
						{"vector_ip_ops", "public"},
						{"vector_l1_ops", "public"},
						{"vector_l2_ops", "public"},
					},
				},
			},
		},
		{
			Name: "vector index definitions render with the hnsw method and operator class",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE items (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL);",
				"CREATE INDEX idx_cos ON items USING hnsw (v vector_cosine_ops);",
				"CREATE INDEX idx_ip ON items USING hnsw (v vector_ip_ops);",
				"CREATE INDEX idx_l1 ON items USING hnsw (v vector_l1_ops);",
				"CREATE INDEX idx_half ON items USING ivfflat (h halfvec_l2_ops);",
				"CREATE INDEX idx_def ON items USING ivfflat (v);",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					// ivfflat indexes map to the same native index as hnsw and render as hnsw, with the
					// operator class always spelled out (pgvector renders `USING ivfflat (v)` for idx_def).
					Query: "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'items' ORDER BY indexname;",
					Expected: []sql.Row{
						{"idx_cos", "CREATE INDEX idx_cos ON public.items USING hnsw (v vector_cosine_ops)"},
						{"idx_def", "CREATE INDEX idx_def ON public.items USING hnsw (v vector_l2_ops)"},
						{"idx_half", "CREATE INDEX idx_half ON public.items USING hnsw (h halfvec_l2_ops)"},
						{"idx_ip", "CREATE INDEX idx_ip ON public.items USING hnsw (v vector_ip_ops)"},
						{"idx_l1", "CREATE INDEX idx_l1 ON public.items USING hnsw (v vector_l1_ops)"},
						{"items_pkey", "CREATE UNIQUE INDEX items_pkey ON public.items USING btree (id)"},
					},
				},
				{
					Query:    "SELECT pg_get_indexdef('idx_cos'::regclass);",
					Expected: []sql.Row{{"CREATE INDEX idx_cos ON public.items USING hnsw (v vector_cosine_ops)"}},
				},
				{
					Query:    "SELECT pg_get_indexdef('idx_half'::regclass);",
					Expected: []sql.Row{{"CREATE INDEX idx_half ON public.items USING hnsw (h halfvec_l2_ops)"}},
				},
			},
		},
	})
}
