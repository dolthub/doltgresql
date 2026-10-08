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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestDropExtension(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "drop extension that does not exist",
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION IF EXISTS pg_graphql;`,
					Expected:        []sql.Row{},
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `extension "pg_graphql" does not exist, skipping`}},
				},
				{
					Query: `DROP EXTENSION IF EXISTS pg_graphql, doltgres_no_such_extension;`,
					ExpectedNotices: []ExpectedNotice{
						{Severity: "NOTICE", Message: `extension "pg_graphql" does not exist, skipping`},
						{Severity: "NOTICE", Message: `extension "doltgres_no_such_extension" does not exist, skipping`},
					},
				},
				{
					Query:           `DROP EXTENSION pg_graphql;`,
					ExpectedErr:     `extension "pg_graphql" does not exist`,
					ExpectedErrCode: "42704",
				},
			},
		},
		{
			Name: "drop extension removes its functions",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT uuid_nil();`,
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
				{
					Query:    `DROP EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT count(*) FROM pg_catalog.pg_extension WHERE extname = 'uuid-ossp';`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT count(*) FROM pg_catalog.pg_proc WHERE proname LIKE 'uuid_ns_%' OR proname LIKE 'uuid_generate_%' OR proname = 'uuid_nil';`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `CREATE EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT uuid_nil();`,
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
			},
		},
		{
			Name: "drop extension with if exists drops the extensions that exist",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION IF EXISTS pg_graphql, "uuid-ossp";`,
					Expected:        []sql.Row{},
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `extension "pg_graphql" does not exist, skipping`}},
				},
				{
					Query:    `SELECT count(*) FROM pg_catalog.pg_proc WHERE proname = 'uuid_nil';`,
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "drop extension drops nothing when a name does not exist",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION "uuid-ossp", pg_graphql;`,
					ExpectedErr:     `extension "pg_graphql" does not exist`,
					ExpectedErrCode: "42704",
				},
				{
					Query:    `SELECT count(*) FROM pg_catalog.pg_proc WHERE proname = 'uuid_nil';`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "drop extension created in a schema outside the search path",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp" WITH SCHEMA public;`,
				`SELECT pg_catalog.set_config('search_path', '', false);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `DROP EXTENSION "uuid-ossp" CASCADE;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT count(*) FROM pg_catalog.pg_proc WHERE proname = 'uuid_nil';`,
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "drop extension removes its types, operators, casts, and aggregates",
			SetUpScript: []string{
				`CREATE EXTENSION vector;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '[1,2]'::vector + '[3,4]'::vector, avg(v) FROM (VALUES ('[1,2]'::vector), ('[3,4]'::vector)) t (v);`,
					Expected: []sql.Row{{"[4,6]", "[2,3]"}},
				},
				{
					Query:    `DROP EXTENSION vector;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT count(*) FROM pg_catalog.pg_type WHERE typname IN ('vector', '_vector', 'halfvec', '_halfvec', 'sparsevec', '_sparsevec');`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:           `SELECT '[1,2]'::vector;`,
					ExpectedErr:     `type "vector" does not exist`,
					ExpectedErrCode: "42704",
					Skip:            true, // Doltgres returns "unable to resolve type `vector`"
				},
				{
					Query:    `CREATE EXTENSION vector;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT '[1,2]'::vector + '[3,4]'::vector, avg(v) FROM (VALUES ('[1,2]'::vector), ('[3,4]'::vector)) t (v);`,
					Expected: []sql.Row{{"[4,6]", "[2,3]"}},
				},
			},
		},
		{
			Name: "drop extension whose types are used by a table",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
				`CREATE EXTENSION vector;`,
				`CREATE TABLE items (id INT PRIMARY KEY, embedding vector(2));`,
				`CREATE TABLE item_lists (id INT PRIMARY KEY, embeddings halfvec[]);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp", vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `SELECT uuid_nil();`,
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
				{
					Query:    `DROP TABLE items;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION vector RESTRICT;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `SELECT '[1,2]'::vector;`,
					Expected: []sql.Row{{"[1,2]"}},
				},
				{
					Query:    `DROP TABLE item_lists;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION vector;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "drop extension cascades to the columns that use its types",
			Skip: true, // TODO: CASCADE does not yet drop dependent objects
			SetUpScript: []string{
				`CREATE EXTENSION vector;`,
				`CREATE TABLE items (id INT PRIMARY KEY, embedding vector(2));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION vector CASCADE;`,
					Expected:        []sql.Row{},
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `drop cascades to column embedding of table items`}},
				},
				{
					Query:    `SELECT column_name FROM information_schema.columns WHERE table_name = 'items';`,
					Expected: []sql.Row{{"id"}},
				},
			},
		},
		{
			Name: "drop extension whose functions are used by a column default",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
				`CREATE TABLE goals (id uuid DEFAULT uuid_generate_v4(), other uuid DEFAULT gen_random_uuid());`,
				`CREATE TABLE notes (id INT PRIMARY KEY, label text DEFAULT upper(uuid_nil()::text || 'x'));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `ALTER TABLE goals ALTER COLUMN id DROP DEFAULT;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `ALTER TABLE notes ALTER COLUMN label DROP DEFAULT;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "drop extension cascades to the column defaults that use its functions",
			Skip: true, // TODO: CASCADE does not yet drop dependent objects
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
				`CREATE TABLE goals (id uuid DEFAULT uuid_generate_v4());`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION "uuid-ossp" CASCADE;`,
					Expected:        []sql.Row{},
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `drop cascades to default value for column id of table goals`}},
				},
				{
					Query:    `SELECT column_default FROM information_schema.columns WHERE table_name = 'goals';`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "drop extension whose functions are used by a generated column",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
				`CREATE TABLE gen (id INT PRIMARY KEY, nil_text text GENERATED ALWAYS AS (uuid_nil()::text) STORED);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP TABLE gen;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "drop extension whose functions are used by a view",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
				`CREATE VIEW plain_view AS SELECT gen_random_uuid() AS n;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE VIEW nil_view AS SELECT uuid_nil() AS n;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP VIEW nil_view;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE VIEW ns_view AS SELECT x.n FROM (SELECT uuid_ns_dns() AS n) x UNION SELECT gen_random_uuid();`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP VIEW ns_view;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "drop extension whose functions are used by a CHECK constraint",
			Skip: true, // Doltgres rejects extension functions in CHECK constraints
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
				`CREATE TABLE checked (t text CHECK (t <> uuid_nil()::text));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP TABLE checked;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "drop extension whose types are used by casts",
			SetUpScript: []string{
				`CREATE EXTENSION vector;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE checked (t text CHECK (t::vector IS NOT NULL));`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP TABLE checked;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE defaulted (t text DEFAULT ('[1,2]'::vector)::text);`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP TABLE defaulted;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE VIEW cast_view AS SELECT '[1,2]'::vector AS v;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP VIEW cast_view;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE VIEW array_view AS SELECT '{}'::halfvec[] AS v;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP VIEW array_view;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION vector;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "drop extension whose functions are used by a domain",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE DOMAIN defaulted_uuid AS uuid DEFAULT uuid_nil();`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP DOMAIN defaulted_uuid;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE DOMAIN checked_uuid AS uuid CHECK (VALUE <> uuid_nil());`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP DOMAIN checked_uuid;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "drop extension whose types are used by a domain",
			SetUpScript: []string{
				`CREATE EXTENSION vector;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE DOMAIN based AS vector;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP DOMAIN based;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE DOMAIN casted AS text DEFAULT ('[1,2]'::vector)::text;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP DOMAIN casted;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE DOMAIN casted_check AS text CHECK (VALUE::vector IS NOT NULL);`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION vector;`,
					ExpectedErr:     `cannot drop extension vector because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP DOMAIN casted_check;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION vector;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "drop extension whose functions are used by window functions and unions",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE VIEW window_view AS SELECT row_number() OVER (ORDER BY uuid_nil()) AS n FROM (VALUES (1)) t (i);`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP VIEW window_view;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE VIEW named_window_view AS SELECT row_number() OVER w AS n FROM (VALUES (1)) t (i) WINDOW w AS (PARTITION BY uuid_ns_url());`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP VIEW named_window_view;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE VIEW union_view AS SELECT gen_random_uuid() AS n UNION ALL SELECT uuid_ns_oid();`,
					Expected: []sql.Row{},
				},
				{
					Query:           `DROP EXTENSION "uuid-ossp";`,
					ExpectedErr:     `cannot drop extension uuid-ossp because other objects depend on it`,
					ExpectedErrCode: "2BP01",
				},
				{
					Query:    `DROP VIEW union_view;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
			},
		},
	})
}
