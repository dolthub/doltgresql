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
	"math"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgtype"

	framework "github.com/dolthub/doltgresql/testing/go"
)

func TestPgvector(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "pgvector vector type",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '[1,2,3]'::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT ' [ 1.5, -0.02 , 3 ] '::vector;",
					Expected: []sql.Row{{"[1.5,-0.02,3]"}},
				},
				{
					Query:    "SELECT '[1e38,-1e-38]'::vector;",
					Expected: []sql.Row{{"[1e+38,-1e-38]"}},
				},
				{
					Query:    "SELECT '[100000,1000000,1234567]'::vector;",
					Expected: []sql.Row{{"[100000,1e+06,1.234567e+06]"}},
				},
				{
					Query:       "SELECT '[]'::vector;",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{
					Query:       "SELECT '[1,]'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[1,]"`,
				},
				{
					Query:       "SELECT 'abc'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "abc"`,
				},
				{
					Query:       "SELECT '[1,2'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[1,2"`,
				},
				{
					Query:       "SELECT '[NaN]'::vector;",
					ExpectedErr: "NaN not allowed in vector",
				},
				{
					Query:       "SELECT '[Infinity,1]'::vector;",
					ExpectedErr: "infinite value not allowed in vector",
				},
				{
					Query:       "SELECT '[1e39]'::vector;",
					ExpectedErr: `"1e39" is out of range for type vector`,
				},
			},
		},
		{
			Name: "pgvector vector functions",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT vector_dims('[1,2,3]'::vector);",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT vector_norm('[3,4]'::vector);",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_distance('[1,2,3]'::vector, '[4,5,6]'::vector);",
					Expected: []sql.Row{{5.196152422706632}},
				},
				{
					Query:    "SELECT inner_product('[1,2]'::vector, '[3,4]'::vector);",
					Expected: []sql.Row{{11.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2]'::vector, '[2,4]'::vector);",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,0]'::vector, '[0,0]'::vector)::text;",
					Expected: []sql.Row{{"NaN"}},
				},
				{
					Query:    "SELECT l1_distance('[1,2]'::vector, '[4,7]'::vector);",
					Expected: []sql.Row{{8.0}},
				},
				{
					Query:    "SELECT l2_normalize('[3,4]'::vector);",
					Expected: []sql.Row{{"[0.6,0.8]"}},
				},
				{
					Query:    "SELECT l2_normalize('[0,0]'::vector);",
					Expected: []sql.Row{{"[0,0]"}},
				},
				{
					Query:    "SELECT binary_quantize('[1,-1,0]'::vector);",
					Expected: []sql.Row{{pgtype.Bits{Bytes: []uint8{0x80}, Len: 3, Valid: true}}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::vector, 2, 3);",
					Expected: []sql.Row{{"[2,3,4]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::vector, 4, 100);",
					Expected: []sql.Row{{"[4,5]"}},
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::vector, 6, 1);",
					ExpectedErr: "vector must have at least 1 dimension",
				},
			},
		},
		{
			Name: "pgvector vector operators",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '[1,2,3]'::vector <-> '[4,5,6]'::vector;",
					Expected: []sql.Row{{5.196152422706632}},
				},
				{
					Query:    "SELECT '[1,2]'::vector <#> '[3,4]'::vector;",
					Expected: []sql.Row{{-11.0}},
				},
				{
					Query:    "SELECT '[1,2]'::vector <=> '[2,4]'::vector;",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT '[1,2]'::vector <+> '[4,7]'::vector;",
					Expected: []sql.Row{{8.0}},
				},
				{
					Query:    "SELECT '[1,2]'::vector + '[3,4]'::vector;",
					Expected: []sql.Row{{"[4,6]"}},
				},
				{
					Query:    "SELECT '[5,6]'::vector - '[1,2]'::vector;",
					Expected: []sql.Row{{"[4,4]"}},
				},
				{
					Query:    "SELECT '[1,2]'::vector * '[3,4]'::vector;",
					Expected: []sql.Row{{"[3,8]"}},
				},
				{
					Query:    "SELECT '[1,2]'::vector || '[3]'::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[1,2]'::vector < '[1,3]'::vector, '[1,2]'::vector = '[1,2]'::vector, '[1,2]'::vector >= '[1,3]'::vector, '[1,2]'::vector != '[1,2]'::vector;",
					Expected: []sql.Row{{"t", "t", "f", "f"}},
				},
				{
					Query:    "SELECT '[1,2]'::vector < '[1,2,3]'::vector;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:       "SELECT '[1,2]'::vector + '[1,2,3]'::vector;",
					ExpectedErr: "different vector dimensions 2 and 3",
				},
				{
					Query:       "SELECT '[3e38,3e38]'::vector + '[3e38,3e38]'::vector;",
					ExpectedErr: "value out of range: overflow",
				},
			},
		},
		{
			Name: "pgvector halfvec",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '[0.1,1]'::halfvec;",
					Expected: []sql.Row{{"[0.099975586,1]"}},
				},
				{
					Query:    "SELECT '[65504,-65504]'::halfvec;",
					Expected: []sql.Row{{"[65504,-65504]"}},
				},
				{
					Query:       "SELECT '[65520]'::halfvec;",
					ExpectedErr: `"65520" is out of range for type halfvec`,
				},
				{
					Query:    "SELECT '[1e-8]'::halfvec;",
					Expected: []sql.Row{{"[0]"}},
				},
				{
					Query:    "SELECT '[1,2]'::halfvec <-> '[4,6]'::halfvec;",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT '[1,2]'::halfvec <#> '[3,4]'::halfvec;",
					Expected: []sql.Row{{-11.0}},
				},
				{
					Query:    "SELECT '[1,2]'::halfvec <=> '[2,4]'::halfvec;",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT '[1,2]'::halfvec <+> '[4,7]'::halfvec;",
					Expected: []sql.Row{{8.0}},
				},
				{
					Query:    "SELECT '[1,2]'::halfvec + '[3,4]'::halfvec;",
					Expected: []sql.Row{{"[4,6]"}},
				},
				{
					Query:    "SELECT l2_normalize('[3,4]'::halfvec);",
					Expected: []sql.Row{{"[0.60009766,0.7998047]"}},
				},
				{
					Query:    "SELECT binary_quantize('[1,-1,0]'::halfvec);",
					Expected: []sql.Row{{pgtype.Bits{Bytes: []uint8{0x80}, Len: 3, Valid: true}}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::halfvec, 2, 3);",
					Expected: []sql.Row{{"[2,3,4]"}},
				},
				{
					Query:    "SELECT vector_dims('[1,2,3]'::halfvec);",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT l2_norm('[3,4]'::halfvec);",
					Expected: []sql.Row{{5.0}},
				},
			},
		},
		{
			Name: "pgvector sparsevec",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '{1:1.5,3:2}/5'::sparsevec;",
					Expected: []sql.Row{{"{1:1.5,3:2}/5"}},
				},
				{
					Query:    "SELECT '{3:1,1:2}/5'::sparsevec;",
					Expected: []sql.Row{{"{1:2,3:1}/5"}},
				},
				{
					Query:    "SELECT '{1:0,2:3}/4'::sparsevec;",
					Expected: []sql.Row{{"{2:3}/4"}},
				},
				{
					Query:    "SELECT '{}/7'::sparsevec;",
					Expected: []sql.Row{{"{}/7"}},
				},
				{
					Query:       "SELECT '{6:1}/5'::sparsevec;",
					ExpectedErr: "sparsevec index out of bounds",
				},
				{
					Query:       "SELECT '{0:1}/5'::sparsevec;",
					ExpectedErr: "sparsevec index out of bounds",
				},
				{
					Query:       "SELECT '{1:1,1:2}/5'::sparsevec;",
					ExpectedErr: "sparsevec indices must not contain duplicates",
				},
				{
					Query:       "SELECT '{1:1}/0'::sparsevec;",
					ExpectedErr: "sparsevec must have at least 1 dimension",
				},
				{
					Query:    "SELECT l2_norm('{1:3,2:4}/2'::sparsevec);",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT '{1:1,2:2}/3'::sparsevec <-> '{1:4,3:6}/3'::sparsevec;",
					Expected: []sql.Row{{7.0}},
				},
				{
					Query:    "SELECT '{1:1,2:2}/2'::sparsevec <#> '{1:3,2:4}/2'::sparsevec;",
					Expected: []sql.Row{{-11.0}},
				},
				{
					Query:    "SELECT '{1:1,2:2}/2'::sparsevec <=> '{1:2,2:4}/2'::sparsevec;",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT '{1:1,2:2}/2'::sparsevec <+> '{1:4,2:7}/2'::sparsevec;",
					Expected: []sql.Row{{8.0}},
				},
				{
					Query:    "SELECT l2_normalize('{1:3,2:4}/3'::sparsevec);",
					Expected: []sql.Row{{"{1:0.6,2:0.8}/3"}},
				},
			},
		},
		{
			Name: "pgvector bit distances",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '101'::bit(3) <~> '111'::bit(3);",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT '101'::bit(3) <%> '111'::bit(3);",
					Expected: []sql.Row{{0.33333333333333337}},
				},
				{
					Query:    "SELECT hamming_distance('101', '111');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT jaccard_distance('101', '111');",
					Expected: []sql.Row{{0.33333333333333337}},
				},
				{
					Query:    "SELECT jaccard_distance('000', '000');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT hamming_distance('10', '111');",
					ExpectedErr: "different bit lengths 2 and 3",
				},
			},
		},
		{
			Name: "pgvector casts",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '[1,2,3]'::vector::halfvec;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[0.1,1]'::halfvec::vector;",
					Expected: []sql.Row{{"[0.099975586,1]"}},
				},
				{
					Query:    "SELECT '[1,0,2]'::vector::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2}/3"}},
				},
				{
					Query:    "SELECT '{1:1,3:2}/3'::sparsevec::vector;",
					Expected: []sql.Row{{"[1,0,2]"}},
				},
				{
					Query:    "SELECT '[1,0,2]'::halfvec::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2}/3"}},
				},
				{
					Query:    "SELECT '{1:1.5}/2'::sparsevec::halfvec;",
					Expected: []sql.Row{{"[1.5,0]"}},
				},
				{
					Query:    "SELECT ARRAY[1,2,3]::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT ARRAY[1.5,2.5]::float4[]::vector;",
					Expected: []sql.Row{{"[1.5,2.5]"}},
				},
				{
					Query:    "SELECT ARRAY[1.5,2.5]::float8[]::vector;",
					Expected: []sql.Row{{"[1.5,2.5]"}},
				},
				{
					Query:    "SELECT ARRAY[1.5,2.5]::numeric[]::vector;",
					Expected: []sql.Row{{"[1.5,2.5]"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector::float4[];",
					Expected: []sql.Row{{"{1,2,3}"}},
				},
				{
					Query:    "SELECT ARRAY[1,2]::halfvec;",
					Expected: []sql.Row{{"[1,2]"}},
				},
				{
					Query:    "SELECT ARRAY[1,0,2]::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2}/3"}},
				},
			},
		},
		{
			Name: "pgvector vectors in tables",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE items (id INTEGER PRIMARY KEY, embedding vector);",
				"INSERT INTO items VALUES (1, '[1,2]'), (2, '[4,7]'), (3, '[0,0]');",
				"CREATE TABLE hitems (id INTEGER PRIMARY KEY, embedding halfvec);",
				"INSERT INTO hitems VALUES (1, '[1,2]'), (2, '[4,7]'), (3, '[0,0]');",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT embedding FROM items ORDER BY id;",
					Expected: []sql.Row{{"[1,2]"}, {"[4,7]"}, {"[0,0]"}},
				},
				{
					Query:    "SELECT id FROM items ORDER BY embedding <-> '[1,1]' LIMIT 2;",
					Expected: []sql.Row{{1}, {3}},
				},
				{
					Query:    "SELECT id, embedding <=> '[1,1]' AS d FROM items WHERE id != 3 ORDER BY d;",
					Expected: []sql.Row{{2, 0.03523617876226781}, {1, 0.05131670194948623}},
				},
				{
					Query:    "SELECT avg(embedding), sum(embedding) FROM items;",
					Expected: []sql.Row{{"[1.6666666,3]", "[5,9]"}},
				},
				{
					Query:    "SELECT avg(embedding) FROM items WHERE id > 100;",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT sum(embedding) FROM items WHERE id > 100;",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT avg(embedding), sum(embedding) FROM hitems;",
					Expected: []sql.Row{{"[1.6669922,3]", "[5,9]"}},
				},
			},
		},
		{
			Name: "pgvector vectors outside the public schema",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE SCHEMA vfx;",
				"CREATE TABLE vfx.docs (id INT PRIMARY KEY, embedding vector(3) NOT NULL, src INT REFERENCES vfx.docs (id));",
				"INSERT INTO vfx.docs VALUES (1, '[1,2,3]', NULL), (2, '[4,5,6]', 1), (3, '[10,10,10]', NULL);",
				"CREATE TABLE vfx.q (id INT PRIMARY KEY, v public.vector(2));",
				"INSERT INTO vfx.q VALUES (1, '[1,1]');",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT id FROM vfx.docs ORDER BY embedding <-> '[1,2,4]' LIMIT 2;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    "SELECT v FROM vfx.q;",
					Expected: []sql.Row{{"[1,1]"}},
				},
				{
					Query:    "SET search_path TO vfx, public;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT vector_dims(embedding) FROM docs WHERE id = 1;",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SET search_path TO vfx;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT vector_dims(embedding) FROM docs WHERE id = 1;",
					ExpectedErr: "vector_dims",
				},
			},
		},
		{
			Name: "pgvector distance operator with a bound parameter",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE items (id INT PRIMARY KEY, embedding vector(3));",
				"INSERT INTO items VALUES (1, '[1,2,3]'), (2, '[4,5,6]'), (3, '[10,10,10]');",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT id FROM items ORDER BY embedding <-> $1 LIMIT 2;",
					BindVars: []any{"[1,2,4]"},
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:    "SELECT id, embedding <-> $1 AS d FROM items ORDER BY d LIMIT 2;",
					BindVars: []any{"[1,2,4]"},
					Expected: []sql.Row{{1, 1.0}, {2, 4.69041575982343}},
				},
			},
		},
		{
			Name: "pgvector copy from csv",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE docs (id INT PRIMARY KEY, embedding vector(3) NOT NULL, src INT);",
				"CREATE TABLE hdocs (id INT PRIMARY KEY, embedding halfvec(3) NOT NULL);",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:             "COPY docs FROM STDIN (FORMAT CSV, HEADER TRUE);",
					CopyFromStdInFile: "csv-load-vector.sql",
				},
				{
					Query:    "SELECT id, src FROM docs ORDER BY id;",
					Expected: []sql.Row{{1, nil}, {2, 1}, {3, nil}},
				},
				{
					Query:    "SELECT id FROM docs ORDER BY embedding <-> '[1,2,4]' LIMIT 2;",
					Expected: []sql.Row{{1}, {3}},
				},
				{
					Query:             "COPY hdocs FROM STDIN (FORMAT CSV, HEADER TRUE);",
					CopyFromStdInFile: "csv-load-halfvec.sql",
				},
				{
					Query:    "SELECT embedding FROM hdocs ORDER BY id;",
					Expected: []sql.Row{{"[1,2,3]"}, {"[0.5,0.25,0.125]"}},
				},
			},
		},
		{
			Name: "pgvector fixture schema shape",
			SetUpScript: []string{
				"BEGIN;",
				"CREATE EXTENSION IF NOT EXISTS vector;",
				"CREATE SCHEMA vf;",
				"CREATE TYPE vf.record_status AS ENUM ('active', 'archived');",
				"CREATE TABLE vf.workspaces (id uuid PRIMARY KEY, name text NOT NULL UNIQUE);",
				`CREATE TABLE vf.documents (
					id uuid PRIMARY KEY,
					workspace_id uuid NOT NULL REFERENCES vf.workspaces (id) ON DELETE CASCADE,
					status vf.record_status NOT NULL,
					fixture_source_id uuid REFERENCES vf.documents (id),
					embedding vector(4) NOT NULL
				);`,
				"CREATE INDEX documents_filter_idx ON vf.documents (workspace_id, status, id);",
				"COMMIT;",
				"INSERT INTO vf.workspaces VALUES ('11111111-1111-1111-1111-111111111111', 'ws1'), ('22222222-2222-2222-2222-222222222222', 'ws2');",
				`INSERT INTO vf.documents VALUES
					('aaaaaaaa-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'active', NULL, '[1,0,0,0]'),
					('aaaaaaaa-0000-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', 'archived', 'aaaaaaaa-0000-0000-0000-000000000001', '[1,0.01,0,0]'),
					('aaaaaaaa-0000-0000-0000-000000000003', '22222222-2222-2222-2222-222222222222', 'active', NULL, '[0,1,0,0]');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT id FROM vf.documents WHERE workspace_id = '11111111-1111-1111-1111-111111111111' AND status = 'active' ORDER BY embedding <=> '[1,0.02,0,0]' LIMIT 1;",
					Expected: []sql.Row{{"aaaaaaaa-0000-0000-0000-000000000001"}},
				},
				{
					Query:    "DELETE FROM vf.workspaces WHERE name = 'ws1';",
					Expected: []sql.Row{},
					Skip:     true, // ON DELETE CASCADE outside the search path needs the schema names in dolt's fk.GetReferencedForeignKeys
				},
				{
					Query:    "SELECT count(*) FROM vf.documents;",
					Expected: []sql.Row{{1}},
					Skip:     true, // depends on the skipped cascading delete above
				},
			},
		},
		{
			Name: "pgvector catalog tables",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT extname, extrelocatable, extversion FROM pg_catalog.pg_extension WHERE extname = 'vector';",
					Expected: []sql.Row{{"vector", "t", "0.8.6"}},
				},
				{
					Query: "SELECT name, default_version, installed_version, comment FROM pg_catalog.pg_available_extensions WHERE name = 'vector';",
					Expected: []sql.Row{
						{"vector", "0.8.6", "0.8.6", "vector data type and ivfflat and hnsw access methods"},
					},
				},
				{
					Query: "SELECT name, version, installed, superuser, trusted, relocatable, schema, requires FROM pg_catalog.pg_available_extension_versions WHERE name = 'vector';",
					Expected: []sql.Row{
						{"vector", "0.8.6", "t", "t", "f", "t", nil, nil},
					},
				},
				{
					Query:    "SELECT count(*) FROM pg_catalog.pg_proc WHERE proname IN ('l2_distance', 'inner_product', 'cosine_distance', 'l1_distance', 'l2_norm', 'l2_normalize', 'binary_quantize', 'subvector');",
					Expected: []sql.Row{{21}},
				},
			},
		},
	})
}

func TestPgvectorVectorType(t *testing.T) {
	//TODO: port copy.sql (needs psql \copy with binary files), and hnsw_*.sql and ivfflat_*.sql (need index access methods)
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "vector_type",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '[1,2,3]'::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[-1,-2,-3]'::vector;",
					Expected: []sql.Row{{"[-1,-2,-3]"}},
				},
				{
					Query:    "SELECT '[1.,2.,3.]'::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT ' [ 1,  2 ,    3  ] '::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[1.23456]'::vector;",
					Expected: []sql.Row{{"[1.23456]"}},
				},
				{
					Query:       "SELECT '[hello,1]'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[hello,1]"`,
				},
				{
					Query:       "SELECT '[NaN,1]'::vector;",
					ExpectedErr: "NaN not allowed in vector",
				},
				{
					Query:       "SELECT '[Infinity,1]'::vector;",
					ExpectedErr: "infinite value not allowed in vector",
				},
				{
					Query:       "SELECT '[-Infinity,1]'::vector;",
					ExpectedErr: "infinite value not allowed in vector",
				},
				{
					Query:    "SELECT '[1.5e38,-1.5e38]'::vector;",
					Expected: []sql.Row{{"[1.5e+38,-1.5e+38]"}},
				},
				{
					Query:    "SELECT '[1.5e+38,-1.5e+38]'::vector;",
					Expected: []sql.Row{{"[1.5e+38,-1.5e+38]"}},
				},
				{
					Query:    "SELECT '[1.5e-38,-1.5e-38]'::vector;",
					Expected: []sql.Row{{"[1.5e-38,-1.5e-38]"}},
				},
				{
					Query:       "SELECT '[4e38,1]'::vector;",
					ExpectedErr: `"4e38" is out of range for type vector`,
				},
				{
					Query:       "SELECT '[-4e38,1]'::vector;",
					ExpectedErr: `"-4e38" is out of range for type vector`,
				},
				{
					Query:    "SELECT '[1e-46,1]'::vector;",
					Expected: []sql.Row{{"[0,1]"}},
				},
				{
					Query:    "SELECT '[-1e-46,1]'::vector;",
					Expected: []sql.Row{{"[-0,1]"}},
				},
				{
					Query:       "SELECT '[1,2,3'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[1,2,3"`,
				},
				{
					Query:       "SELECT '[1,2,3]9'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[1,2,3]9"`,
				},
				{
					Query:       "SELECT '1,2,3'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "1,2,3"`,
				},
				{
					Query:       "SELECT ''::vector;",
					ExpectedErr: `invalid input syntax for type vector: ""`,
				},
				{
					Query:       "SELECT '['::vector;",
					ExpectedErr: `invalid input syntax for type vector: "["`,
				},
				{ // The error message reports the input with its whitespace trimmed
					Query:       "SELECT '[ '::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[ "`,
					Skip:        true,
				},
				{
					Query:       "SELECT '[,'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[,"`,
				},
				{
					Query:       "SELECT '[]'::vector;",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{
					Query:       "SELECT '[ ]'::vector;",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{
					Query:       "SELECT '[,]'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[,]"`,
				},
				{
					Query:       "SELECT '[1,]'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[1,]"`,
				},
				{
					Query:       "SELECT '[1a]'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[1a]"`,
				},
				{
					Query:       "SELECT '[1,,3]'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[1,,3]"`,
				},
				{
					Query:       "SELECT '[1, ,3]'::vector;",
					ExpectedErr: `invalid input syntax for type vector: "[1, ,3]"`,
				},
				{
					Query:    "SELECT '[1,2,3]'::vector(3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:       "SELECT '[1,2,3]'::vector(2);",
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{
					Query:       "SELECT '[1,2,3]'::vector(3, 2);",
					ExpectedErr: "invalid type modifier",
				},
				{
					Query:       "SELECT '[1,2,3]'::vector('a');",
					ExpectedErr: `invalid input syntax for type integer: "a"`,
				},
				{
					Query:       "SELECT '[1,2,3]'::vector(0);",
					ExpectedErr: "dimensions for type vector must be at least 1",
				},
				{
					Query:       "SELECT '[1,2,3]'::vector(16001);",
					ExpectedErr: "dimensions for type vector cannot exceed 16000",
				},
				{
					Query:    `SELECT unnest('{"[1,2,3]", "[4,5,6]"}'::vector[]);`,
					Expected: []sql.Row{{"[1,2,3]"}, {"[4,5,6]"}},
				},
				{
					Query:       `SELECT '{"[1,2,3]"}'::vector(2)[];`,
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{
					Query:    "SELECT '[1,2,3]'::vector + '[4,5,6]';",
					Expected: []sql.Row{{"[5,7,9]"}},
				},
				{
					Query:       "SELECT '[3e38]'::vector + '[3e38]';",
					ExpectedErr: "value out of range: overflow",
				},
				{
					Query:       "SELECT '[1,2]'::vector + '[3]';",
					ExpectedErr: "different vector dimensions 2 and 1",
				},
				{
					Query:    "SELECT '[1,2,3]'::vector - '[4,5,6]';",
					Expected: []sql.Row{{"[-3,-3,-3]"}},
				},
				{
					Query:       "SELECT '[-3e38]'::vector - '[3e38]';",
					ExpectedErr: "value out of range: overflow",
				},
				{
					Query:       "SELECT '[1,2]'::vector - '[3]';",
					ExpectedErr: "different vector dimensions 2 and 1",
				},
				{
					Query:    "SELECT '[1,2,3]'::vector * '[4,5,6]';",
					Expected: []sql.Row{{"[4,10,18]"}},
				},
				{
					Query:       "SELECT '[1e37]'::vector * '[1e37]';",
					ExpectedErr: "value out of range: overflow",
				},
				{
					Query:       "SELECT '[1e-37]'::vector * '[1e-37]';",
					ExpectedErr: "value out of range: underflow",
				},
				{
					Query:       "SELECT '[1,2]'::vector * '[3]';",
					ExpectedErr: "different vector dimensions 2 and 1",
				},
				{ // || with an untyped literal does not resolve to vector_concat yet
					Query:    "SELECT '[1,2,3]'::vector || '[4,5]';",
					Expected: []sql.Row{{"[1,2,3,4,5]"}},
					Skip:     true,
				},
				{ // array_fill is not yet implemented
					Query:       "SELECT array_fill(0, ARRAY[16000])::vector || '[1]';",
					ExpectedErr: "vector cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{
					Query:    "SELECT '[1,2,3]'::vector < '[1,2,3]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector < '[1,2]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector <= '[1,2,3]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector <= '[1,2]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector = '[1,2,3]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector = '[1,2]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector != '[1,2,3]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector != '[1,2]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector >= '[1,2,3]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector >= '[1,2]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector > '[1,2,3]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector > '[1,2]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT vector_cmp('[1,2,3]', '[1,2,3]');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT vector_cmp('[1,2,3]', '[0,0,0]');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT vector_cmp('[0,0,0]', '[1,2,3]');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT vector_cmp('[1,2]', '[1,2,3]');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT vector_cmp('[1,2,3]', '[1,2]');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT vector_cmp('[1,2]', '[2,3,4]');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT vector_cmp('[2,3]', '[1,2,3]');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT vector_dims('[1,2,3]'::vector);",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT round(vector_norm('[1,1]')::numeric, 5);",
					Expected: []sql.Row{{framework.Numeric("1.41421")}},
				},
				{
					Query:    "SELECT vector_norm('[3,4]');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT vector_norm('[0,1]');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT vector_norm('[3e37,4e37]')::real;",
					Expected: []sql.Row{{float64(float32(5e37))}},
				},
				{
					Query:    "SELECT vector_norm('[0,0]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT vector_norm('[2]');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT l2_distance('[0,0]'::vector, '[3,4]');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_distance('[0,0]'::vector, '[0,1]');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT l2_distance('[1,2]'::vector, '[3]');",
					ExpectedErr: "different vector dimensions 2 and 1",
				},
				{
					Query:    "SELECT l2_distance('[3e38]'::vector, '[-3e38]');",
					Expected: []sql.Row{{math.Inf(1)}},
				},
				{
					Query:    "SELECT l2_distance('[1,1,1,1,1,1,1,1,1]'::vector, '[1,1,1,1,1,1,1,4,5]');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT '[0,0]'::vector <-> '[3,4]';",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT inner_product('[1,2]'::vector, '[3,4]');",
					Expected: []sql.Row{{11.0}},
				},
				{
					Query:       "SELECT inner_product('[1,2]'::vector, '[3]');",
					ExpectedErr: "different vector dimensions 2 and 1",
				},
				{
					Query:    "SELECT inner_product('[3e38]'::vector, '[3e38]');",
					Expected: []sql.Row{{math.Inf(1)}},
				},
				{
					Query:    "SELECT inner_product('[1,1,1,1,1,1,1,1,1]'::vector, '[1,2,3,4,5,6,7,8,9]');",
					Expected: []sql.Row{{45.0}},
				},
				{
					Query:    "SELECT '[1,2]'::vector <#> '[3,4]';",
					Expected: []sql.Row{{-11.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2]'::vector, '[2,4]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2]'::vector, '[0,0]')::text;",
					Expected: []sql.Row{{"NaN"}},
				},
				{
					Query:    "SELECT cosine_distance('[1,1]'::vector, '[1,1]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,0]'::vector, '[0,2]');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,1]'::vector, '[-1,-1]');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:       "SELECT cosine_distance('[1,2]'::vector, '[3]');",
					ExpectedErr: "different vector dimensions 2 and 1",
				},
				{
					Query:    "SELECT cosine_distance('[1,1]'::vector, '[1.1,1.1]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,1]'::vector, '[-1.1,-1.1]');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT cosine_distance('[3e38]'::vector, '[3e38]')::text;",
					Expected: []sql.Row{{"NaN"}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2,3,4,5,6,7,8,9]'::vector, '[1,2,3,4,5,6,7,8,9]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2,3,4,5,6,7,8,9]'::vector, '[-1,-2,-3,-4,-5,-6,-7,-8,-9]');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT '[1,2]'::vector <=> '[2,4]';",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT l1_distance('[0,0]'::vector, '[3,4]');",
					Expected: []sql.Row{{7.0}},
				},
				{
					Query:    "SELECT l1_distance('[0,0]'::vector, '[0,1]');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT l1_distance('[1,2]'::vector, '[3]');",
					ExpectedErr: "different vector dimensions 2 and 1",
				},
				{
					Query:    "SELECT l1_distance('[3e38]'::vector, '[-3e38]');",
					Expected: []sql.Row{{math.Inf(1)}},
				},
				{
					Query:    "SELECT l1_distance('[1,2,3,4,5,6,7,8,9]'::vector, '[1,2,3,4,5,6,7,8,9]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT l1_distance('[1,2,3,4,5,6,7,8,9]'::vector, '[0,3,2,5,4,7,6,9,8]');",
					Expected: []sql.Row{{9.0}},
				},
				{
					Query:    "SELECT '[0,0]'::vector <+> '[3,4]';",
					Expected: []sql.Row{{7.0}},
				},
				{
					Query:    "SELECT l2_normalize('[3,4]'::vector);",
					Expected: []sql.Row{{"[0.6,0.8]"}},
				},
				{
					Query:    "SELECT l2_normalize('[3,0]'::vector);",
					Expected: []sql.Row{{"[1,0]"}},
				},
				{
					Query:    "SELECT l2_normalize('[0,0.1]'::vector);",
					Expected: []sql.Row{{"[0,1]"}},
				},
				{
					Query:    "SELECT l2_normalize('[0,0]'::vector);",
					Expected: []sql.Row{{"[0,0]"}},
				},
				{
					Query:    "SELECT l2_normalize('[3e38]'::vector);",
					Expected: []sql.Row{{"[1]"}},
				},
				{
					Query:    "SELECT binary_quantize('[1,0,-1]'::vector);",
					Expected: []sql.Row{{pgtype.Bits{Bytes: []uint8{0x80}, Len: 3, Valid: true}}},
				},
				{
					Query:    "SELECT binary_quantize('[0,0.1,-0.2,-0.3,0.4,0.5,0.6,-0.7,0.8,-0.9,1]'::vector);",
					Expected: []sql.Row{{pgtype.Bits{Bytes: []uint8{0x4E, 0xA0}, Len: 11, Valid: true}}},
				},
				{
					Query:    "SELECT binary_quantize('[1,2,3,-4,5,6,-7,8,1,-2,-3,4,5,-6,7,8,-1,2,3]'::vector);",
					Expected: []sql.Row{{pgtype.Bits{Bytes: []uint8{0xED, 0x9B, 0x60}, Len: 19, Valid: true}}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::vector, 1, 3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::vector, 3, 2);",
					Expected: []sql.Row{{"[3,4]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::vector, -1, 3);",
					Expected: []sql.Row{{"[1]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::vector, 3, 9);",
					Expected: []sql.Row{{"[3,4,5]"}},
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::vector, 1, 0);",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::vector, 3, -1);",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::vector, -1, 2);",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::vector, 2147483647, 10);",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::vector, 3, 2147483647);",
					Expected: []sql.Row{{"[3,4,5]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::vector, -2147483644, 2147483647);",
					Expected: []sql.Row{{"[1,2]"}},
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:    "SELECT avg(v) FROM unnest(ARRAY['[1,2,3]'::vector, '[3,5,7]']) v;",
					Expected: []sql.Row{{"[2,3.5,5]"}},
					Skip:     true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:    "SELECT avg(v) FROM unnest(ARRAY['[1,2,3]'::vector, '[3,5,7]', NULL]) v;",
					Expected: []sql.Row{{"[2,3.5,5]"}},
					Skip:     true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:    "SELECT avg(v) FROM unnest(ARRAY[]::vector[]) v;",
					Expected: []sql.Row{{nil}},
					Skip:     true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:       "SELECT avg(v) FROM unnest(ARRAY['[1,2]'::vector, '[3]']) v;",
					ExpectedErr: "expected 2 dimensions, not 1",
					Skip:        true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:    "SELECT avg(v) FROM unnest(ARRAY['[3e38]'::vector, '[3e38]']) v;",
					Expected: []sql.Row{{"[3e+38]"}},
					Skip:     true,
				},
				{
					Query:    "SELECT vector_avg('{2,2,4,6}');",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT vector_avg('{0}');",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:       "SELECT vector_avg('{1}');",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{ // Multidimensional arrays are not yet supported
					Query:       "SELECT vector_avg('{{2,2,4,6}}');",
					ExpectedErr: "vector_avg: expected state array",
					Skip:        true,
				},
				{
					Query:       "SELECT vector_avg('{NULL,2,4,6}');",
					ExpectedErr: "vector_avg: expected state array",
				},
				{
					Query:       "SELECT vector_avg('{}');",
					ExpectedErr: "vector_avg: expected state array",
				},
				{ // int4[] does not implicitly cast to float8[] yet
					Query:       "SELECT vector_avg(array_agg(n)) FROM generate_series(1, 16002) n;",
					ExpectedErr: "vector cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{
					Query:    "SELECT vector_accum('{0}', '[1,2,3]');",
					Expected: []sql.Row{{"{1,1,2,3}"}},
				},
				{
					Query:    "SELECT vector_accum('{0,0,0,0}', '[1,2,3]');",
					Expected: []sql.Row{{"{1,1,2,3}"}},
				},
				{ // Multidimensional arrays are not yet supported
					Query:       "SELECT vector_accum('{{0}}', '[1,2,3]');",
					ExpectedErr: "vector_accum: expected state array",
					Skip:        true,
				},
				{
					Query:       "SELECT vector_accum('{NULL}', '[1,2,3]');",
					ExpectedErr: "vector_accum: expected state array",
				},
				{
					Query:       "SELECT vector_accum('{}', '[1,2,3]');",
					ExpectedErr: "vector_accum: expected state array",
				},
				{
					Query:       "SELECT vector_accum('{0,0}', '[1,2,3]');",
					ExpectedErr: "expected 1 dimensions, not 3",
				},
				{
					Query:    "SELECT vector_combine('{1,2}', '{3,4}');",
					Expected: []sql.Row{{"{4,6}"}},
				},
				{ // Multidimensional arrays are not yet supported
					Query:       "SELECT vector_combine('{{1,2}}', '{3,4}');",
					ExpectedErr: "vector_combine: expected state array",
					Skip:        true,
				},
				{ // Multidimensional arrays are not yet supported
					Query:       "SELECT vector_combine('{1,2}', '{{3,4}}');",
					ExpectedErr: "vector_combine: expected state array",
					Skip:        true,
				},
				{
					Query:       "SELECT vector_combine('{NULL,2}', '{3,4}');",
					ExpectedErr: "vector_combine: expected state array",
				},
				{
					Query:       "SELECT vector_combine('{1,2}', '{3,NULL}');",
					ExpectedErr: "vector_combine: expected state array",
				},
				{
					Query:       "SELECT vector_combine('{}', '{0}');",
					ExpectedErr: "vector_combine: expected state array",
				},
				{
					Query:       "SELECT vector_combine('{0}', '{}');",
					ExpectedErr: "vector_combine: expected state array",
				},
				{
					Query:       "SELECT vector_combine('{0}', '{0}');",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{ // int4[] does not implicitly cast to float8[] yet
					Query:       "SELECT vector_combine('{0}', (SELECT array_agg(n) FROM generate_series(1, 16002) n));",
					ExpectedErr: "vector cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{ // int4[] does not implicitly cast to float8[] yet
					Query:       "SELECT vector_combine((SELECT array_agg(n) FROM generate_series(1, 16002) n), '{0}');",
					ExpectedErr: "vector cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{ // int4[] does not implicitly cast to float8[] yet
					Query:       "SELECT vector_combine((SELECT array_agg(n) FROM generate_series(1, 16002) n), (SELECT array_agg(n) FROM generate_series(1, 16002) n));",
					ExpectedErr: "vector cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:    "SELECT sum(v) FROM unnest(ARRAY['[1,2,3]'::vector, '[3,5,7]']) v;",
					Expected: []sql.Row{{"[4,7,10]"}},
					Skip:     true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:    "SELECT sum(v) FROM unnest(ARRAY['[1,2,3]'::vector, '[3,5,7]', NULL]) v;",
					Expected: []sql.Row{{"[4,7,10]"}},
					Skip:     true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:    "SELECT sum(v) FROM unnest(ARRAY[]::vector[]) v;",
					Expected: []sql.Row{{nil}},
					Skip:     true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:       "SELECT sum(v) FROM unnest(ARRAY['[1,2]'::vector, '[3]']) v;",
					ExpectedErr: "different vector dimensions 2 and 1",
					Skip:        true,
				},
				{ // Aggregates over vector values from unnest are not yet supported
					Query:       "SELECT sum(v) FROM unnest(ARRAY['[3e38]'::vector, '[3e38]']) v;",
					ExpectedErr: "value out of range: overflow",
					Skip:        true,
				},
			},
		},
	})
}

func TestPgvectorHalfvec(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "halfvec",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '[1,2,3]'::halfvec;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[-1,-2,-3]'::halfvec;",
					Expected: []sql.Row{{"[-1,-2,-3]"}},
				},
				{
					Query:    "SELECT '[1.,2.,3.]'::halfvec;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT ' [ 1,  2 ,    3  ] '::halfvec;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[1.23456]'::halfvec;",
					Expected: []sql.Row{{"[1.234375]"}},
				},
				{
					Query:       "SELECT '[hello,1]'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[hello,1]"`,
				},
				{
					Query:       "SELECT '[NaN,1]'::halfvec;",
					ExpectedErr: "NaN not allowed in halfvec",
				},
				{
					Query:       "SELECT '[Infinity,1]'::halfvec;",
					ExpectedErr: "infinite value not allowed in halfvec",
				},
				{
					Query:       "SELECT '[-Infinity,1]'::halfvec;",
					ExpectedErr: "infinite value not allowed in halfvec",
				},
				{
					Query:    "SELECT '[65519,-65519]'::halfvec;",
					Expected: []sql.Row{{"[65504,-65504]"}},
				},
				{
					Query:       "SELECT '[65520,-65520]'::halfvec;",
					ExpectedErr: `"65520" is out of range for type halfvec`,
				},
				{
					Query:    "SELECT '[1e-8,-1e-8]'::halfvec;",
					Expected: []sql.Row{{"[0,-0]"}},
				},
				{
					Query:       "SELECT '[4e38,1]'::halfvec;",
					ExpectedErr: `"4e38" is out of range for type halfvec`,
				},
				{
					Query:    "SELECT '[1e-46,1]'::halfvec;",
					Expected: []sql.Row{{"[0,1]"}},
				},
				{
					Query:       "SELECT '[1,2,3'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[1,2,3"`,
				},
				{
					Query:       "SELECT '[1,2,3]9'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[1,2,3]9"`,
				},
				{
					Query:       "SELECT '1,2,3'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "1,2,3"`,
				},
				{
					Query:       "SELECT ''::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: ""`,
				},
				{
					Query:       "SELECT '['::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "["`,
				},
				{ // The error message reports the input with its whitespace trimmed
					Query:       "SELECT '[ '::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[ "`,
					Skip:        true,
				},
				{
					Query:       "SELECT '[,'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[,"`,
				},
				{
					Query:       "SELECT '[]'::halfvec;",
					ExpectedErr: "halfvec must have at least 1 dimension",
				},
				{
					Query:       "SELECT '[ ]'::halfvec;",
					ExpectedErr: "halfvec must have at least 1 dimension",
				},
				{
					Query:       "SELECT '[,]'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[,]"`,
				},
				{
					Query:       "SELECT '[1,]'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[1,]"`,
				},
				{
					Query:       "SELECT '[1a]'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[1a]"`,
				},
				{
					Query:       "SELECT '[1,,3]'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[1,,3]"`,
				},
				{
					Query:       "SELECT '[1, ,3]'::halfvec;",
					ExpectedErr: `invalid input syntax for type halfvec: "[1, ,3]"`,
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec(3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:       "SELECT '[1,2,3]'::halfvec(2);",
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{
					Query:       "SELECT '[1,2,3]'::halfvec(3, 2);",
					ExpectedErr: "invalid type modifier",
				},
				{
					Query:       "SELECT '[1,2,3]'::halfvec('a');",
					ExpectedErr: `invalid input syntax for type integer: "a"`,
				},
				{
					Query:       "SELECT '[1,2,3]'::halfvec(0);",
					ExpectedErr: "dimensions for type halfvec must be at least 1",
				},
				{
					Query:       "SELECT '[1,2,3]'::halfvec(16001);",
					ExpectedErr: "dimensions for type halfvec cannot exceed 16000",
				},
				{
					Query:    `SELECT unnest('{"[1,2,3]", "[4,5,6]"}'::halfvec[]);`,
					Expected: []sql.Row{{"[1,2,3]"}, {"[4,5,6]"}},
				},
				{
					Query:       `SELECT '{"[1,2,3]"}'::halfvec(2)[];`,
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec + '[4,5,6]';",
					Expected: []sql.Row{{"[5,7,9]"}},
				},
				{
					Query:       "SELECT '[65519]'::halfvec + '[65519]';",
					ExpectedErr: "value out of range: overflow",
				},
				{
					Query:       "SELECT '[1,2]'::halfvec + '[3]';",
					ExpectedErr: "different halfvec dimensions 2 and 1",
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec - '[4,5,6]';",
					Expected: []sql.Row{{"[-3,-3,-3]"}},
				},
				{
					Query:       "SELECT '[-65519]'::halfvec - '[65519]';",
					ExpectedErr: "value out of range: overflow",
				},
				{
					Query:       "SELECT '[1,2]'::halfvec - '[3]';",
					ExpectedErr: "different halfvec dimensions 2 and 1",
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec * '[4,5,6]';",
					Expected: []sql.Row{{"[4,10,18]"}},
				},
				{
					Query:       "SELECT '[65519]'::halfvec * '[65519]';",
					ExpectedErr: "value out of range: overflow",
				},
				{
					Query:       "SELECT '[1e-7]'::halfvec * '[1e-7]';",
					ExpectedErr: "value out of range: underflow",
				},
				{
					Query:       "SELECT '[1,2]'::halfvec * '[3]';",
					ExpectedErr: "different halfvec dimensions 2 and 1",
				},
				{ // || with an untyped literal does not resolve to halfvec_concat yet
					Query:    "SELECT '[1,2,3]'::halfvec || '[4,5]';",
					Expected: []sql.Row{{"[1,2,3,4,5]"}},
					Skip:     true,
				},
				{ // array_fill is not yet implemented
					Query:       "SELECT array_fill(0, ARRAY[16000])::halfvec || '[1]';",
					ExpectedErr: "halfvec cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec < '[1,2,3]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec < '[1,2]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec <= '[1,2,3]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec <= '[1,2]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec = '[1,2,3]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec = '[1,2]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec != '[1,2,3]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec != '[1,2]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec >= '[1,2,3]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec >= '[1,2]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec > '[1,2,3]';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec > '[1,2]';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT halfvec_cmp('[1,2,3]', '[1,2,3]');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT halfvec_cmp('[1,2,3]', '[0,0,0]');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT halfvec_cmp('[0,0,0]', '[1,2,3]');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT halfvec_cmp('[1,2]', '[1,2,3]');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT halfvec_cmp('[1,2,3]', '[1,2]');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT halfvec_cmp('[1,2]', '[2,3,4]');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT halfvec_cmp('[2,3]', '[1,2,3]');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT vector_dims('[1,2,3]'::halfvec);",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT round(l2_norm('[1,1]'::halfvec)::numeric, 5);",
					Expected: []sql.Row{{framework.Numeric("1.41421")}},
				},
				{
					Query:    "SELECT l2_norm('[3,4]'::halfvec);",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_norm('[0,1]'::halfvec);",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT l2_norm('[0,0]'::halfvec);",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT l2_norm('[2]'::halfvec);",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT l2_distance('[0,0]'::halfvec, '[3,4]');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_distance('[0,0]'::halfvec, '[0,1]');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT l2_distance('[1,2]'::halfvec, '[3]');",
					ExpectedErr: "different halfvec dimensions 2 and 1",
				},
				{
					Query:    "SELECT l2_distance('[1,1,1,1,1,1,1,1,1]'::halfvec, '[1,1,1,1,1,1,1,4,5]');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT '[0,0]'::halfvec <-> '[3,4]';",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT inner_product('[1,2]'::halfvec, '[3,4]');",
					Expected: []sql.Row{{11.0}},
				},
				{
					Query:       "SELECT inner_product('[1,2]'::halfvec, '[3]');",
					ExpectedErr: "different halfvec dimensions 2 and 1",
				},
				{
					Query:    "SELECT inner_product('[65504]'::halfvec, '[65504]');",
					Expected: []sql.Row{{4290774016.0}},
				},
				{
					Query:    "SELECT inner_product('[1,1,1,1,1,1,1,1,1]'::halfvec, '[1,2,3,4,5,6,7,8,9]');",
					Expected: []sql.Row{{45.0}},
				},
				{
					Query:    "SELECT '[1,2]'::halfvec <#> '[3,4]';",
					Expected: []sql.Row{{-11.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2]'::halfvec, '[2,4]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2]'::halfvec, '[0,0]')::text;",
					Expected: []sql.Row{{"NaN"}},
				},
				{
					Query:    "SELECT cosine_distance('[1,1]'::halfvec, '[1,1]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,0]'::halfvec, '[0,2]');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,1]'::halfvec, '[-1,-1]');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:       "SELECT cosine_distance('[1,2]'::halfvec, '[3]');",
					ExpectedErr: "different halfvec dimensions 2 and 1",
				},
				{
					Query:    "SELECT cosine_distance('[1,1]'::halfvec, '[1.1,1.1]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,1]'::halfvec, '[-1.1,-1.1]');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2,3,4,5,6,7,8,9]'::halfvec, '[1,2,3,4,5,6,7,8,9]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('[1,2,3,4,5,6,7,8,9]'::halfvec, '[-1,-2,-3,-4,-5,-6,-7,-8,-9]');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT '[1,2]'::halfvec <=> '[2,4]';",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT l1_distance('[0,0]'::halfvec, '[3,4]');",
					Expected: []sql.Row{{7.0}},
				},
				{
					Query:    "SELECT l1_distance('[0,0]'::halfvec, '[0,1]');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT l1_distance('[1,2]'::halfvec, '[3]');",
					ExpectedErr: "different halfvec dimensions 2 and 1",
				},
				{
					Query:    "SELECT l1_distance('[1,2,3,4,5,6,7,8,9]'::halfvec, '[1,2,3,4,5,6,7,8,9]');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT l1_distance('[1,2,3,4,5,6,7,8,9]'::halfvec, '[0,3,2,5,4,7,6,9,8]');",
					Expected: []sql.Row{{9.0}},
				},
				{
					Query:    "SELECT '[0,0]'::halfvec <+> '[3,4]';",
					Expected: []sql.Row{{7.0}},
				},
				{
					Query:    "SELECT l2_normalize('[3,4]'::halfvec);",
					Expected: []sql.Row{{"[0.60009766,0.7998047]"}},
				},
				{
					Query:    "SELECT l2_normalize('[3,0]'::halfvec);",
					Expected: []sql.Row{{"[1,0]"}},
				},
				{
					Query:    "SELECT l2_normalize('[0,0.1]'::halfvec);",
					Expected: []sql.Row{{"[0,1]"}},
				},
				{
					Query:    "SELECT l2_normalize('[0,0]'::halfvec);",
					Expected: []sql.Row{{"[0,0]"}},
				},
				{
					Query:    "SELECT l2_normalize('[65504]'::halfvec);",
					Expected: []sql.Row{{"[1]"}},
				},
				{
					Query:    "SELECT binary_quantize('[1,0,-1]'::halfvec);",
					Expected: []sql.Row{{pgtype.Bits{Bytes: []uint8{0x80}, Len: 3, Valid: true}}},
				},
				{
					Query:    "SELECT binary_quantize('[0,0.1,-0.2,-0.3,0.4,0.5,0.6,-0.7,0.8,-0.9,1]'::halfvec);",
					Expected: []sql.Row{{pgtype.Bits{Bytes: []uint8{0x4E, 0xA0}, Len: 11, Valid: true}}},
				},
				{
					Query:    "SELECT binary_quantize('[1,2,3,-4,5,6,-7,8,1,-2,-3,4,5,-6,7,8,-1,2,3]'::halfvec);",
					Expected: []sql.Row{{pgtype.Bits{Bytes: []uint8{0xED, 0x9B, 0x60}, Len: 19, Valid: true}}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::halfvec, 1, 3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::halfvec, 3, 2);",
					Expected: []sql.Row{{"[3,4]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::halfvec, -1, 3);",
					Expected: []sql.Row{{"[1]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::halfvec, 3, 9);",
					Expected: []sql.Row{{"[3,4,5]"}},
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::halfvec, 1, 0);",
					ExpectedErr: "halfvec must have at least 1 dimension",
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::halfvec, 3, -1);",
					ExpectedErr: "halfvec must have at least 1 dimension",
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::halfvec, -1, 2);",
					ExpectedErr: "halfvec must have at least 1 dimension",
				},
				{
					Query:       "SELECT subvector('[1,2,3,4,5]'::halfvec, 2147483647, 10);",
					ExpectedErr: "halfvec must have at least 1 dimension",
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::halfvec, 3, 2147483647);",
					Expected: []sql.Row{{"[3,4,5]"}},
				},
				{
					Query:    "SELECT subvector('[1,2,3,4,5]'::halfvec, -2147483644, 2147483647);",
					Expected: []sql.Row{{"[1,2]"}},
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:    "SELECT avg(v) FROM unnest(ARRAY['[1,2,3]'::halfvec, '[3,5,7]']) v;",
					Expected: []sql.Row{{"[2,3.5,5]"}},
					Skip:     true,
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:    "SELECT avg(v) FROM unnest(ARRAY['[1,2,3]'::halfvec, '[3,5,7]', NULL]) v;",
					Expected: []sql.Row{{"[2,3.5,5]"}},
					Skip:     true,
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:    "SELECT avg(v) FROM unnest(ARRAY[]::halfvec[]) v;",
					Expected: []sql.Row{{nil}},
					Skip:     true,
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:       "SELECT avg(v) FROM unnest(ARRAY['[1,2]'::halfvec, '[3]']) v;",
					ExpectedErr: "expected 2 dimensions, not 1",
					Skip:        true,
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:    "SELECT avg(v) FROM unnest(ARRAY['[65504]'::halfvec, '[65504]']) v;",
					Expected: []sql.Row{{"[65504]"}},
					Skip:     true,
				},
				{
					Query:    "SELECT halfvec_avg('{2,2,4,6}');",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT halfvec_avg('{0}');",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:       "SELECT halfvec_avg('{1}');",
					ExpectedErr: "halfvec must have at least 1 dimension",
				},
				{ // Multidimensional arrays are not yet supported
					Query:       "SELECT halfvec_avg('{{2,2,4,6}}');",
					ExpectedErr: "halfvec_avg: expected state array",
					Skip:        true,
				},
				{
					Query:       "SELECT halfvec_avg('{NULL,2,4,6}');",
					ExpectedErr: "halfvec_avg: expected state array",
				},
				{
					Query:       "SELECT halfvec_avg('{}');",
					ExpectedErr: "halfvec_avg: expected state array",
				},
				{ // int4[] does not implicitly cast to float8[] yet
					Query:       "SELECT halfvec_avg(array_agg(n)) FROM generate_series(1, 16002) n;",
					ExpectedErr: "halfvec cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{
					Query:    "SELECT halfvec_accum('{0}', '[1,2,3]');",
					Expected: []sql.Row{{"{1,1,2,3}"}},
				},
				{
					Query:    "SELECT halfvec_accum('{0,0,0,0}', '[1,2,3]');",
					Expected: []sql.Row{{"{1,1,2,3}"}},
				},
				{ // Multidimensional arrays are not yet supported
					Query:       "SELECT halfvec_accum('{{0}}', '[1,2,3]');",
					ExpectedErr: "halfvec_accum: expected state array",
					Skip:        true,
				},
				{
					Query:       "SELECT halfvec_accum('{NULL}', '[1,2,3]');",
					ExpectedErr: "halfvec_accum: expected state array",
				},
				{
					Query:       "SELECT halfvec_accum('{}', '[1,2,3]');",
					ExpectedErr: "halfvec_accum: expected state array",
				},
				{
					Query:       "SELECT halfvec_accum('{0,0}', '[1,2,3]');",
					ExpectedErr: "expected 1 dimensions, not 3",
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:    "SELECT sum(v) FROM unnest(ARRAY['[1,2,3]'::halfvec, '[3,5,7]']) v;",
					Expected: []sql.Row{{"[4,7,10]"}},
					Skip:     true,
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:    "SELECT sum(v) FROM unnest(ARRAY['[1,2,3]'::halfvec, '[3,5,7]', NULL]) v;",
					Expected: []sql.Row{{"[4,7,10]"}},
					Skip:     true,
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:    "SELECT sum(v) FROM unnest(ARRAY[]::halfvec[]) v;",
					Expected: []sql.Row{{nil}},
					Skip:     true,
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:       "SELECT sum(v) FROM unnest(ARRAY['[1,2]'::halfvec, '[3]']) v;",
					ExpectedErr: "different halfvec dimensions 2 and 1",
					Skip:        true,
				},
				{ // Aggregates over halfvec values from unnest are not yet supported
					Query:       "SELECT sum(v) FROM unnest(ARRAY['[65504]'::halfvec, '[65504]']) v;",
					ExpectedErr: "value out of range: overflow",
					Skip:        true,
				},
			},
		},
	})
}

func TestPgvectorSparsevec(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "sparsevec",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT '{1:1.5,3:3.5}/5'::sparsevec;",
					Expected: []sql.Row{{"{1:1.5,3:3.5}/5"}},
				},
				{
					Query:    "SELECT '{1:-2,3:-4}/5'::sparsevec;",
					Expected: []sql.Row{{"{1:-2,3:-4}/5"}},
				},
				{
					Query:    "SELECT '{1:2.,3:4.}/5'::sparsevec;",
					Expected: []sql.Row{{"{1:2,3:4}/5"}},
				},
				{
					Query:    "SELECT ' { 1 : 1.5 ,  3  :  3.5  } / 5 '::sparsevec;",
					Expected: []sql.Row{{"{1:1.5,3:3.5}/5"}},
				},
				{
					Query:    "SELECT '{1:1.23456}/1'::sparsevec;",
					Expected: []sql.Row{{"{1:1.23456}/1"}},
				},
				{
					Query:       "SELECT '{1:hello,2:1}/2'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{1:hello,2:1}/2"`,
				},
				{
					Query:       "SELECT '{1:NaN,2:1}/2'::sparsevec;",
					ExpectedErr: "NaN not allowed in sparsevec",
				},
				{
					Query:       "SELECT '{1:Infinity,2:1}/2'::sparsevec;",
					ExpectedErr: "infinite value not allowed in sparsevec",
				},
				{
					Query:       "SELECT '{1:-Infinity,2:1}/2'::sparsevec;",
					ExpectedErr: "infinite value not allowed in sparsevec",
				},
				{
					Query:    "SELECT '{1:1.5e38,2:-1.5e38}/2'::sparsevec;",
					Expected: []sql.Row{{"{1:1.5e+38,2:-1.5e+38}/2"}},
				},
				{
					Query:    "SELECT '{1:1.5e+38,2:-1.5e+38}/2'::sparsevec;",
					Expected: []sql.Row{{"{1:1.5e+38,2:-1.5e+38}/2"}},
				},
				{
					Query:    "SELECT '{1:1.5e-38,2:-1.5e-38}/2'::sparsevec;",
					Expected: []sql.Row{{"{1:1.5e-38,2:-1.5e-38}/2"}},
				},
				{
					Query:       "SELECT '{1:4e38,2:1}/2'::sparsevec;",
					ExpectedErr: `"4e38" is out of range for type sparsevec`,
				},
				{
					Query:       "SELECT '{1:-4e38,2:1}/2'::sparsevec;",
					ExpectedErr: `"-4e38" is out of range for type sparsevec`,
				},
				{
					Query:       "SELECT '{1:1e-46,2:1}/2'::sparsevec;",
					ExpectedErr: `"1e-46" is out of range for type sparsevec`,
				},
				{
					Query:       "SELECT '{1:-1e-46,2:1}/2'::sparsevec;",
					ExpectedErr: `"-1e-46" is out of range for type sparsevec`,
				},
				{
					Query:       "SELECT ''::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: ""`,
				},
				{
					Query:       "SELECT '{'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{"`,
				},
				{ // The error message reports the input with its whitespace trimmed
					Query:       "SELECT '{ '::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{ "`,
					Skip:        true,
				},
				{
					Query:       "SELECT '{:'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{:"`,
				},
				{
					Query:       "SELECT '{,'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{,"`,
				},
				{
					Query:       "SELECT '{}'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{}"`,
				},
				{
					Query:       "SELECT '{}/'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{}/"`,
				},
				{
					Query:    "SELECT '{}/1'::sparsevec;",
					Expected: []sql.Row{{"{}/1"}},
				},
				{
					Query:       "SELECT '{}/1a'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{}/1a"`,
				},
				{
					Query:    "SELECT '{ }/1'::sparsevec;",
					Expected: []sql.Row{{"{}/1"}},
				},
				{
					Query:       "SELECT '{:}/1'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{:}/1"`,
				},
				{
					Query:       "SELECT '{,}/1'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{,}/1"`,
				},
				{
					Query:       "SELECT '{1,}/1'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{1,}/1"`,
				},
				{
					Query:       "SELECT '{:1}/1'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{:1}/1"`,
				},
				{
					Query:       "SELECT '{1:}/1'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{1:}/1"`,
				},
				{
					Query:       "SELECT '{1a:1}/1'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{1a:1}/1"`,
				},
				{
					Query:       "SELECT '{1:1a}/1'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{1:1a}/1"`,
				},
				{
					Query:       "SELECT '{1:1,}/1'::sparsevec;",
					ExpectedErr: `invalid input syntax for type sparsevec: "{1:1,}/1"`,
				},
				{
					Query:    "SELECT '{1:0,2:1,3:0}/3'::sparsevec;",
					Expected: []sql.Row{{"{2:1}/3"}},
				},
				{
					Query:    "SELECT '{2:1,1:1}/2'::sparsevec;",
					Expected: []sql.Row{{"{1:1,2:1}/2"}},
				},
				{
					Query:       "SELECT '{1:1,1:1}/2'::sparsevec;",
					ExpectedErr: "sparsevec indices must not contain duplicates",
				},
				{
					Query:       "SELECT '{1:1,2:1,1:1}/2'::sparsevec;",
					ExpectedErr: "sparsevec indices must not contain duplicates",
				},
				{
					Query:    "SELECT '{}/5'::sparsevec;",
					Expected: []sql.Row{{"{}/5"}},
				},
				{
					Query:       "SELECT '{}/-1'::sparsevec;",
					ExpectedErr: "sparsevec must have at least 1 dimension",
				},
				{
					Query:    "SELECT '{}/1000000000'::sparsevec;",
					Expected: []sql.Row{{"{}/1000000000"}},
				},
				{
					Query:       "SELECT '{}/1000000001'::sparsevec;",
					ExpectedErr: "sparsevec cannot have more than 1000000000 dimensions",
				},
				{
					Query:       "SELECT '{}/2147483648'::sparsevec;",
					ExpectedErr: "sparsevec cannot have more than 1000000000 dimensions",
				},
				{
					Query:       "SELECT '{}/-2147483649'::sparsevec;",
					ExpectedErr: "sparsevec must have at least 1 dimension",
				},
				{
					Query:       "SELECT '{}/9223372036854775808'::sparsevec;",
					ExpectedErr: "sparsevec cannot have more than 1000000000 dimensions",
				},
				{
					Query:       "SELECT '{}/-9223372036854775809'::sparsevec;",
					ExpectedErr: "sparsevec must have at least 1 dimension",
				},
				{
					Query:       "SELECT '{2147483647:1}/1'::sparsevec;",
					ExpectedErr: "sparsevec index out of bounds",
				},
				{
					Query:       "SELECT '{2147483648:1}/1'::sparsevec;",
					ExpectedErr: "sparsevec index out of bounds",
				},
				{
					Query:       "SELECT '{-2147483648:1}/1'::sparsevec;",
					ExpectedErr: "sparsevec index out of bounds",
				},
				{
					Query:       "SELECT '{-2147483649:1}/1'::sparsevec;",
					ExpectedErr: "sparsevec index out of bounds",
				},
				{
					Query:       "SELECT '{0:1}/1'::sparsevec;",
					ExpectedErr: "sparsevec index out of bounds",
				},
				{
					Query:       "SELECT '{2:1}/1'::sparsevec;",
					ExpectedErr: "sparsevec index out of bounds",
				},
				{
					Query:    "SELECT '{}/3'::sparsevec(3);",
					Expected: []sql.Row{{"{}/3"}},
				},
				{
					Query:       "SELECT '{}/3'::sparsevec(2);",
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{
					Query:       "SELECT '{}/3'::sparsevec(3, 2);",
					ExpectedErr: "invalid type modifier",
				},
				{
					Query:       "SELECT '{}/3'::sparsevec('a');",
					ExpectedErr: `invalid input syntax for type integer: "a"`,
				},
				{
					Query:       "SELECT '{}/3'::sparsevec(0);",
					ExpectedErr: "dimensions for type sparsevec must be at least 1",
				},
				{
					Query:       "SELECT '{}/3'::sparsevec(1000000001);",
					ExpectedErr: "dimensions for type sparsevec cannot exceed 1000000000",
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec < '{1:1,2:2,3:3}/3';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec < '{1:1,2:2}/2';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec <= '{1:1,2:2,3:3}/3';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec <= '{1:1,2:2}/2';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec = '{1:1,2:2,3:3}/3';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec = '{1:1,2:2}/2';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec != '{1:1,2:2,3:3}/3';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec != '{1:1,2:2}/2';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec >= '{1:1,2:2,3:3}/3';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec >= '{1:1,2:2}/2';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec > '{1:1,2:2,3:3}/3';",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT '{1:1,2:2,3:3}/3'::sparsevec > '{1:1,2:2}/2';",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT sparsevec_cmp('{1:1,2:2,3:3}/3', '{1:1,2:2,3:3}/3');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT sparsevec_cmp('{1:1,2:2,3:3}/3', '{}/3');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT sparsevec_cmp('{}/3', '{1:1,2:2,3:3}/3');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT sparsevec_cmp('{1:1,2:2}/2', '{1:1,2:2,3:3}/3');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT sparsevec_cmp('{1:1,2:2,3:3}/3', '{1:1,2:2}/2');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT sparsevec_cmp('{1:1,2:2}/2', '{1:2,2:3,3:4}/3');",
					Expected: []sql.Row{{-1}},
				},
				{
					Query:    "SELECT sparsevec_cmp('{1:2,2:3}/2', '{1:1,2:2,3:3}/3');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT round(l2_norm('{1:1,2:1}/2'::sparsevec)::numeric, 5);",
					Expected: []sql.Row{{framework.Numeric("1.41421")}},
				},
				{
					Query:    "SELECT l2_norm('{1:3,2:4}/2'::sparsevec);",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_norm('{2:1}/2'::sparsevec);",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT l2_norm('{1:3e37,2:4e37}/2'::sparsevec)::real;",
					Expected: []sql.Row{{float64(float32(5e37))}},
				},
				{
					Query:    "SELECT l2_norm('{}/2'::sparsevec);",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT l2_norm('{1:2}/1'::sparsevec);",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT l2_distance('{}/2'::sparsevec, '{1:3,2:4}/2');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_distance('{1:3}/2'::sparsevec, '{2:4}/2');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_distance('{2:4}/2'::sparsevec, '{1:3}/2');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_distance('{1:3,2:4}/2'::sparsevec, '{}/2');",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT l2_distance('{}/2'::sparsevec, '{2:1}/2');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT '{}/2'::sparsevec <-> '{1:3,2:4}/2';",
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    "SELECT inner_product('{1:1,2:2}/2'::sparsevec, '{1:2,2:4}/2');",
					Expected: []sql.Row{{10.0}},
				},
				{
					Query:       "SELECT inner_product('{1:1,2:2}/2'::sparsevec, '{1:3}/1');",
					ExpectedErr: "different sparsevec dimensions 2 and 1",
				},
				{
					Query:    "SELECT inner_product('{1:1,3:3}/4'::sparsevec, '{2:2,4:4}/4');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT inner_product('{2:2,4:4}/4'::sparsevec, '{1:1,3:3}/4');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT inner_product('{1:1,3:3,5:5}/5'::sparsevec, '{2:4,3:6,4:8}/5');",
					Expected: []sql.Row{{18.0}},
				},
				{
					Query:    "SELECT inner_product('{1:1}/2'::sparsevec, '{}/2');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT inner_product('{}/2'::sparsevec, '{1:1}/2');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT inner_product('{1:3e38}/1'::sparsevec, '{1:3e38}/1');",
					Expected: []sql.Row{{math.Inf(1)}},
				},
				{
					Query:    "SELECT '{1:1,2:2}/2'::sparsevec <#> '{1:3,2:4}/2';",
					Expected: []sql.Row{{-11.0}},
				},
				{
					Query:    "SELECT cosine_distance('{1:1,2:2}/2'::sparsevec, '{1:2,2:4}/2');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('{1:1,2:2}/2'::sparsevec, '{}/2')::text;",
					Expected: []sql.Row{{"NaN"}},
				},
				{
					Query:    "SELECT cosine_distance('{1:1,2:1}/2'::sparsevec, '{1:1,2:1}/2');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('{1:1}/2'::sparsevec, '{2:2}/2');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT cosine_distance('{1:1,2:1}/2'::sparsevec, '{1:-1,2:-1}/2');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT cosine_distance('{1:2}/2'::sparsevec, '{2:2}/2');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT cosine_distance('{2:2}/2'::sparsevec, '{1:2}/2');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT cosine_distance('{1:1,2:2}/2'::sparsevec, '{1:3}/1');",
					ExpectedErr: "different sparsevec dimensions 2 and 1",
				},
				{
					Query:    "SELECT cosine_distance('{1:1,2:1}/2'::sparsevec, '{1:1.1,2:1.1}/2');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT cosine_distance('{1:1,2:1}/2'::sparsevec, '{1:-1.1,2:-1.1}/2');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT cosine_distance('{1:3e38}/1'::sparsevec, '{1:3e38}/1')::text;",
					Expected: []sql.Row{{"NaN"}},
				},
				{
					Query:    "SELECT cosine_distance('{}/1'::sparsevec, '{}/1')::text;",
					Expected: []sql.Row{{"NaN"}},
				},
				{
					Query:    "SELECT '{1:1,2:2}/2'::sparsevec <=> '{1:2,2:4}/2';",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT l1_distance('{}/2'::sparsevec, '{1:3,2:4}/2');",
					Expected: []sql.Row{{7.0}},
				},
				{
					Query:    "SELECT l1_distance('{}/2'::sparsevec, '{2:1}/2');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT l1_distance('{1:1,2:2}/2'::sparsevec, '{1:3}/1');",
					ExpectedErr: "different sparsevec dimensions 2 and 1",
				},
				{
					Query:    "SELECT l1_distance('{1:3e38}/1'::sparsevec, '{1:-3e38}/1');",
					Expected: []sql.Row{{math.Inf(1)}},
				},
				{
					Query:    "SELECT l1_distance('{1:1,3:3,5:5,7:7}/8'::sparsevec, '{2:2,4:4,6:6,8:8}/8');",
					Expected: []sql.Row{{36.0}},
				},
				{
					Query:    "SELECT l1_distance('{1:1,3:3,5:5,7:7,9:9}/9'::sparsevec, '{2:2,4:4,6:6,8:8}/9');",
					Expected: []sql.Row{{45.0}},
				},
				{
					Query:    "SELECT '{}/2'::sparsevec <+> '{1:3,2:4}/2';",
					Expected: []sql.Row{{7.0}},
				},
				{
					Query:    "SELECT l2_normalize('{1:3,2:4}/2'::sparsevec);",
					Expected: []sql.Row{{"{1:0.6,2:0.8}/2"}},
				},
				{
					Query:    "SELECT l2_normalize('{1:3}/2'::sparsevec);",
					Expected: []sql.Row{{"{1:1}/2"}},
				},
				{
					Query:    "SELECT l2_normalize('{2:0.1}/2'::sparsevec);",
					Expected: []sql.Row{{"{2:1}/2"}},
				},
				{
					Query:    "SELECT l2_normalize('{}/2'::sparsevec);",
					Expected: []sql.Row{{"{}/2"}},
				},
				{
					Query:    "SELECT l2_normalize('{1:3e38}/1'::sparsevec);",
					Expected: []sql.Row{{"{1:1}/1"}},
				},
				{
					Query:    "SELECT l2_normalize('{1:3e38,2:1e-37}/2'::sparsevec);",
					Expected: []sql.Row{{"{1:1}/2"}},
				},
				{
					Query:    "SELECT l2_normalize('{2:3e37,4:3e-37,6:4e37,8:4e-37}/9'::sparsevec);",
					Expected: []sql.Row{{"{2:0.6,6:0.8}/9"}},
				},
			},
		},
	})
}

func TestPgvectorCast(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "cast",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT ARRAY[1,2,3]::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT ARRAY[1.0,2.0,3.0]::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT ARRAY[1,2,3]::float4[]::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT ARRAY[1,2,3]::float8[]::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT ARRAY[1,2,3]::numeric[]::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector::real[];",
					Expected: []sql.Row{{"{1,2,3}"}},
				},
				{
					Query:    "SELECT '{1,2,3}'::real[]::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '{1,2,3}'::real[]::vector(3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:       "SELECT '{1,2,3}'::real[]::vector(2);",
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{
					Query:       "SELECT '{NULL}'::real[]::vector;",
					ExpectedErr: "array must not contain nulls",
				},
				{
					Query:       "SELECT '{NaN}'::real[]::vector;",
					ExpectedErr: "NaN not allowed in vector",
				},
				{
					Query:       "SELECT '{Infinity}'::real[]::vector;",
					ExpectedErr: "infinite value not allowed in vector",
				},
				{
					Query:       "SELECT '{-Infinity}'::real[]::vector;",
					ExpectedErr: "infinite value not allowed in vector",
				},
				{
					Query:       "SELECT '{}'::real[]::vector;",
					ExpectedErr: "vector must have at least 1 dimension",
				},
				{ // Multidimensional arrays are not yet supported
					Query:       "SELECT '{{1}}'::real[]::vector;",
					ExpectedErr: "array must be 1-D",
					Skip:        true,
				},
				{
					Query:    "SELECT '{1,2,3}'::double precision[]::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '{1,2,3}'::double precision[]::vector(3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:       "SELECT '{1,2,3}'::double precision[]::vector(2);",
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{
					Query:       "SELECT '{4e38,-4e38}'::double precision[]::vector;",
					ExpectedErr: "infinite value not allowed in vector",
				},
				{
					Query:    "SELECT '{1e-46,-1e-46}'::double precision[]::vector;",
					Expected: []sql.Row{{"[0,-0]"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector::halfvec;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::vector::halfvec(3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:       "SELECT '[1,2,3]'::vector::halfvec(2);",
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{ // Reports overflow rather than the out-of-range message
					Query:       "SELECT '[65520]'::vector::halfvec;",
					ExpectedErr: `"65520" is out of range for type halfvec`,
					Skip:        true,
				},
				{
					Query:    "SELECT '[1e-8]'::vector::halfvec;",
					Expected: []sql.Row{{"[0]"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec::vector;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '[1,2,3]'::halfvec::vector(3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:       "SELECT '[1,2,3]'::halfvec::vector(2);",
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{
					Query:    "SELECT '{1,2,3}'::real[]::halfvec;",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:    "SELECT '{1,2,3}'::real[]::halfvec(3);",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					Query:       "SELECT '{1,2,3}'::real[]::halfvec(2);",
					ExpectedErr: "expected 2 dimensions, not 3",
				},
				{ // Reports overflow rather than the out-of-range message
					Query:       "SELECT '{65520,-65520}'::real[]::halfvec;",
					ExpectedErr: `"65520" is out of range for type halfvec`,
					Skip:        true,
				},
				{
					Query:    "SELECT '{1e-8,-1e-8}'::real[]::halfvec;",
					Expected: []sql.Row{{"[0,-0]"}},
				},
				{
					Query:    "SELECT '[0,1.5,0,3.5,0]'::vector::sparsevec;",
					Expected: []sql.Row{{"{2:1.5,4:3.5}/5"}},
				},
				{
					Query:    "SELECT '[0,1.5,0,3.5,0]'::vector::sparsevec(5);",
					Expected: []sql.Row{{"{2:1.5,4:3.5}/5"}},
				},
				{
					Query:       "SELECT '[0,1.5,0,3.5,0]'::vector::sparsevec(4);",
					ExpectedErr: "expected 4 dimensions, not 5",
				},
				{
					Query:    "SELECT '{2:1.5,4:3.5}/5'::sparsevec::vector;",
					Expected: []sql.Row{{"[0,1.5,0,3.5,0]"}},
				},
				{
					Query:    "SELECT '{2:1.5,4:3.5}/5'::sparsevec::vector(5);",
					Expected: []sql.Row{{"[0,1.5,0,3.5,0]"}},
				},
				{
					Query:       "SELECT '{2:1.5,4:3.5}/5'::sparsevec::vector(4);",
					ExpectedErr: "expected 4 dimensions, not 5",
				},
				{
					Query:       "SELECT '{}/16001'::sparsevec::vector;",
					ExpectedErr: "vector cannot have more than 16000 dimensions",
				},
				{
					Query:    "SELECT '[0,1.5,0,3.5,0]'::halfvec::sparsevec;",
					Expected: []sql.Row{{"{2:1.5,4:3.5}/5"}},
				},
				{
					Query:    "SELECT '[0,1.5,0,3.5,0]'::halfvec::sparsevec(5);",
					Expected: []sql.Row{{"{2:1.5,4:3.5}/5"}},
				},
				{
					Query:       "SELECT '[0,1.5,0,3.5,0]'::halfvec::sparsevec(4);",
					ExpectedErr: "expected 4 dimensions, not 5",
				},
				{
					Query:    "SELECT '{2:1.5,4:3.5}/5'::sparsevec::halfvec;",
					Expected: []sql.Row{{"[0,1.5,0,3.5,0]"}},
				},
				{
					Query:    "SELECT '{2:1.5,4:3.5}/5'::sparsevec::halfvec(5);",
					Expected: []sql.Row{{"[0,1.5,0,3.5,0]"}},
				},
				{
					Query:       "SELECT '{2:1.5,4:3.5}/5'::sparsevec::halfvec(4);",
					ExpectedErr: "expected 4 dimensions, not 5",
				},
				{
					Query:       "SELECT '{}/16001'::sparsevec::halfvec;",
					ExpectedErr: "halfvec cannot have more than 16000 dimensions",
				},
				{ // Reports overflow rather than the out-of-range message
					Query:       "SELECT '{1:65520}/1'::sparsevec::halfvec;",
					ExpectedErr: `"65520" is out of range for type halfvec`,
					Skip:        true,
				},
				{
					Query:    "SELECT '{1:1e-8}/1'::sparsevec::halfvec;",
					Expected: []sql.Row{{"[0]"}},
				},
				{
					Query:    "SELECT ARRAY[1,0,2,0,3,0]::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2,5:3}/6"}},
				},
				{
					Query:    "SELECT ARRAY[1.0,0.0,2.0,0.0,3.0,0.0]::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2,5:3}/6"}},
				},
				{
					Query:    "SELECT ARRAY[1,0,2,0,3,0]::float4[]::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2,5:3}/6"}},
				},
				{
					Query:    "SELECT ARRAY[1,0,2,0,3,0]::float8[]::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2,5:3}/6"}},
				},
				{
					Query:    "SELECT ARRAY[1,0,2,0,3,0]::numeric[]::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2,5:3}/6"}},
				},
				{
					Query:    "SELECT '{1,0,2,0,3,0}'::real[]::sparsevec;",
					Expected: []sql.Row{{"{1:1,3:2,5:3}/6"}},
				},
				{
					Query:    "SELECT '{1,0,2,0,3,0}'::real[]::sparsevec(6);",
					Expected: []sql.Row{{"{1:1,3:2,5:3}/6"}},
				},
				{
					Query:       "SELECT '{1,0,2,0,3,0}'::real[]::sparsevec(5);",
					ExpectedErr: "expected 5 dimensions, not 6",
				},
				{
					Query:       "SELECT '{NULL}'::real[]::sparsevec;",
					ExpectedErr: "array must not contain nulls",
				},
				{
					Query:       "SELECT '{NaN}'::real[]::sparsevec;",
					ExpectedErr: "NaN not allowed in sparsevec",
				},
				{
					Query:       "SELECT '{Infinity}'::real[]::sparsevec;",
					ExpectedErr: "infinite value not allowed in sparsevec",
				},
				{
					Query:       "SELECT '{-Infinity}'::real[]::sparsevec;",
					ExpectedErr: "infinite value not allowed in sparsevec",
				},
				{
					Query:       "SELECT '{}'::real[]::sparsevec;",
					ExpectedErr: "sparsevec must have at least 1 dimension",
				},
				{ // Multidimensional arrays are not yet supported
					Query:       "SELECT '{{1}}'::real[]::sparsevec;",
					ExpectedErr: "array must be 1-D",
					Skip:        true,
				},
				{ // array_agg results cannot be cast to vector yet
					Query:       "SELECT array_agg(n)::vector FROM generate_series(1, 16001) n;",
					ExpectedErr: "vector cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{ // array_agg results do not resolve against array parameters yet
					Query:       "SELECT array_to_vector(array_agg(n), 16001, false) FROM generate_series(1, 16001) n;",
					ExpectedErr: "vector cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{ // array_agg results cannot be cast to halfvec yet
					Query:       "SELECT array_agg(n)::halfvec FROM generate_series(1, 16001) n;",
					ExpectedErr: "halfvec cannot have more than 16000 dimensions",
					Skip:        true,
				},
				{ // array_agg results cannot be cast to sparsevec yet
					Query:       "SELECT array_agg(n)::sparsevec FROM generate_series(1, 16001) n;",
					ExpectedErr: "sparsevec cannot have more than 16000 non-zero elements",
					Skip:        true,
				},
				{
					Query:    "SELECT ARRAY[1,2,3] = ARRAY[1,2,3];",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
	})
}

func TestPgvectorBit(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "bit",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT hamming_distance('111', '111');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT hamming_distance('111', '110');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT hamming_distance('111', '100');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT hamming_distance('111', '000');",
					Expected: []sql.Row{{3.0}},
				},
				{
					Query:    "SELECT hamming_distance('10101010101010101010', '01010101010101010101');",
					Expected: []sql.Row{{20.0}},
				},
				{
					Query:    "SELECT hamming_distance('101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101', '101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT hamming_distance('101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101', '010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010');",
					Expected: []sql.Row{{513.0}},
				},
				{
					Query:    "SELECT hamming_distance('110000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011', '100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001');",
					Expected: []sql.Row{{2.0}},
				},
				{
					Query:    "SELECT hamming_distance('', '');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:       "SELECT hamming_distance('111', '00');",
					ExpectedErr: "different bit lengths 3 and 2",
				},
				{
					Query:    "SELECT hamming_distance('111', '000'::varbit(4));",
					Expected: []sql.Row{{3.0}},
				},
				{
					Query:       "SELECT hamming_distance('111', '0000'::varbit(4));",
					ExpectedErr: "different bit lengths 3 and 4",
				},
				{
					Query:    "SELECT jaccard_distance('1111', '1111');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT jaccard_distance('1111', '1110');",
					Expected: []sql.Row{{0.25}},
				},
				{
					Query:    "SELECT jaccard_distance('1111', '1100');",
					Expected: []sql.Row{{0.5}},
				},
				{
					Query:    "SELECT jaccard_distance('1111', '1000');",
					Expected: []sql.Row{{0.75}},
				},
				{
					Query:    "SELECT jaccard_distance('1111', '0000');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT jaccard_distance('1100', '1000');",
					Expected: []sql.Row{{0.5}},
				},
				{
					Query:    "SELECT jaccard_distance('10101010101010101010', '01010101010101010101');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT jaccard_distance('101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101', '101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101');",
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    "SELECT jaccard_distance('101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101', '010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:    "SELECT jaccard_distance('110000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011', '100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001');",
					Expected: []sql.Row{{0.5}},
				},
				{
					Query:    "SELECT jaccard_distance('', '');",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT jaccard_distance('1111', '000');",
					ExpectedErr: "different bit lengths 4 and 3",
				},
				{
					Query:    "SELECT jaccard_distance('1111', '0000'::varbit(5));",
					Expected: []sql.Row{{1.0}},
				},
				{
					Query:       "SELECT jaccard_distance('1111', '00000'::varbit(5));",
					ExpectedErr: "different bit lengths 4 and 5",
				},
			},
		},
	})
}

func TestPgvectorBtree(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "btree vector",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE t (val vector(3));",
				"INSERT INTO t (val) VALUES ('[0,0,0]'), ('[1,2,3]'), ('[1,1,1]'), (NULL);",
				"CREATE INDEX ON t (val);",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT * FROM t WHERE val = '[1,2,3]';",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					// Skipped because the default ORDER BY NULL ordering places NULLs first instead of last.
					Skip:     true,
					Query:    "SELECT * FROM t ORDER BY val;",
					Expected: []sql.Row{{"[0,0,0]"}, {"[1,1,1]"}, {"[1,2,3]"}, {nil}},
				},
			},
		},
		{
			Name: "btree halfvec",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE t (val halfvec(3));",
				"INSERT INTO t (val) VALUES ('[0,0,0]'), ('[1,2,3]'), ('[1,1,1]'), (NULL);",
				"CREATE INDEX ON t (val);",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT * FROM t WHERE val = '[1,2,3]';",
					Expected: []sql.Row{{"[1,2,3]"}},
				},
				{
					// Skipped because the default ORDER BY NULL ordering places NULLs first instead of last.
					Skip:     true,
					Query:    "SELECT * FROM t ORDER BY val;",
					Expected: []sql.Row{{"[0,0,0]"}, {"[1,1,1]"}, {"[1,2,3]"}, {nil}},
				},
			},
		},
		{
			Name: "btree sparsevec",
			SetUpScript: []string{
				"CREATE EXTENSION vector;",
				"CREATE TABLE t (val sparsevec(3));",
				"INSERT INTO t (val) VALUES ('{}/3'), ('{1:1,2:2,3:3}/3'), ('{1:1,2:1,3:1}/3'), (NULL);",
				"CREATE INDEX ON t (val);",
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT * FROM t WHERE val = '{1:1,2:2,3:3}/3';",
					Expected: []sql.Row{{"{1:1,2:2,3:3}/3"}},
				},
				{
					// Skipped because the default ORDER BY NULL ordering places NULLs first instead of last.
					Skip:     true,
					Query:    "SELECT * FROM t ORDER BY val;",
					Expected: []sql.Row{{"{}/3"}, {"{1:1,2:1,3:1}/3"}, {"{1:1,2:2,3:3}/3"}, {nil}},
				},
			},
		},
	})
}
