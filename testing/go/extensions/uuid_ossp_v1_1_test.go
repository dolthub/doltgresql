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

func TestUUIDOssp(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: "uuid-ossp",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT uuid_ns_url();",
					Expected: []sql.Row{{"6ba7b811-9dad-11d1-80b4-00c04fd430c8"}},
				},
				{
					Query:    "SELECT uuid_generate_v3('00000000-0000-0000-0000-000000000000'::uuid, 'example text');",
					Expected: []sql.Row{{"a55b875a-1bd9-31af-ac66-7d8323785c6e"}},
				},
				{
					Query:    "SELECT uuid_generate_v3('00000000-0000-0000-0000-000000000001'::uuid, 'example text');",
					Expected: []sql.Row{{"a319ab51-8e26-37c6-942f-7dd5fda5c3ef"}},
				},
				{
					Query:    "SELECT uuid_generate_v3(uuid_ns_url(), 'example text');",
					Expected: []sql.Row{{"6541262f-d622-3e35-8873-2b227591bf69"}},
				},
				{
					Query:    "SELECT uuid_nil();",
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
				{
					Query:    "SELECT length(uuid_nil()::text);",
					Expected: []sql.Row{{36}},
				},
				{
					Query:    "SELECT length(uuid_generate_v4()::text);",
					Expected: []sql.Row{{36}},
				},
				{
					Query:    "SELECT uuid_generate_v4() = uuid_nil();",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `WITH u1 AS (SELECT uuid_nil() AS id), u2 AS (SELECT uuid_nil() AS id) SELECT (SELECT id FROM u1) = (SELECT id FROM u2);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `WITH u1 AS (SELECT uuid_generate_v4() AS id), u2 AS (SELECT uuid_generate_v4() AS id) SELECT (SELECT id FROM u1) = (SELECT id FROM u2);`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "uuid-ossp namespace functions",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT uuid_nil();",
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
				{
					Query:    "SELECT uuid_ns_dns();",
					Expected: []sql.Row{{"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}},
				},
				{
					Query:    "SELECT uuid_ns_url();",
					Expected: []sql.Row{{"6ba7b811-9dad-11d1-80b4-00c04fd430c8"}},
				},
				{
					Query:    "SELECT uuid_ns_oid();",
					Expected: []sql.Row{{"6ba7b812-9dad-11d1-80b4-00c04fd430c8"}},
				},
				{
					Query:    "SELECT uuid_ns_x500();",
					Expected: []sql.Row{{"6ba7b814-9dad-11d1-80b4-00c04fd430c8"}},
				},
				{
					Query:    "SELECT uuid_ns_dns() = uuid_ns_dns(), uuid_ns_dns() = uuid_ns_url();",
					Expected: []sql.Row{{"t", "f"}},
				},
			},
		},
		{
			Name: "uuid-ossp uuid_generate_v3",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT uuid_generate_v3(uuid_ns_dns(), 'www.postgresql.org');",
					Expected: []sql.Row{{"9a0d5f51-76ff-394e-ba97-b28a9ff12209"}},
				},
				{
					Query:    "SELECT uuid_generate_v3(uuid_nil(), '');",
					Expected: []sql.Row{{"4ae71336-e44b-39bf-b9d2-752e234818a5"}},
				},
				{
					Query:    "SELECT uuid_generate_v3(uuid_ns_dns(), 'héllo wörld');",
					Expected: []sql.Row{{"2f301b42-2eaf-3cc3-8646-69e0ffde841f"}},
				},
				{
					Query:    "SELECT uuid_generate_v3(uuid_ns_url(), repeat('a', 1000));",
					Expected: []sql.Row{{"d8f8a14e-39ec-3186-8107-4d4e5a41d2c0"}},
				},
				{ // The version nibble is the 15th character of the textual form
					Query:    "SELECT substring(uuid_generate_v3(uuid_nil(), 'x')::text, 15, 1);",
					Expected: []sql.Row{{"3"}},
				},
				{ // The variant nibble is the 20th character, and RFC 4122 restricts it to 8, 9, a or b
					Query:    "SELECT substring(uuid_generate_v3(uuid_nil(), 'x')::text, 20, 1) IN ('8', '9', 'a', 'b');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT uuid_generate_v3(uuid_nil(), 'abc') = uuid_generate_v3(uuid_nil(), 'abc');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT uuid_generate_v3(uuid_nil(), 'abc') = uuid_generate_v3(uuid_nil(), 'ABC');",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT uuid_generate_v3(uuid_ns_dns(), 'abc') = uuid_generate_v3(uuid_ns_url(), 'abc');",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT uuid_generate_v3(NULL, 'abc') IS NULL, uuid_generate_v3(uuid_nil(), NULL) IS NULL;",
					Expected: []sql.Row{{"t", "t"}},
				},
				{
					Query:       "SELECT uuid_generate_v3('not-a-uuid', 'abc');",
					ExpectedErr: `invalid input syntax for type uuid: "not-a-uuid"`,
				},
			},
		},
		{
			Name: "uuid-ossp uuid_generate_v5",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT uuid_generate_v5(uuid_ns_url(), 'example text');",
					Expected: []sql.Row{{"59edfb26-7819-5209-86a3-79a6da9035ba"}},
				},
				{
					Query:    "SELECT uuid_generate_v5(uuid_ns_dns(), 'www.postgresql.org');",
					Expected: []sql.Row{{"1826c6c4-4d1f-534f-9dcd-7a15978dfeb9"}},
				},
				{
					Query:    "SELECT uuid_generate_v5(uuid_ns_oid(), 'example text');",
					Expected: []sql.Row{{"5758e964-d604-5cde-9b32-57368fc3b1ff"}},
				},
				{
					Query:    "SELECT uuid_generate_v5(uuid_ns_x500(), 'example text');",
					Expected: []sql.Row{{"1d8aac60-3096-5014-bfd7-1f816c37cf50"}},
				},
				{
					Query:    "SELECT uuid_generate_v5(uuid_nil(), '');",
					Expected: []sql.Row{{"e129f27c-5103-5c5c-844b-cdf0a15e160d"}},
				},
				{
					Query:    "SELECT uuid_generate_v5(uuid_ns_dns(), 'héllo wörld');",
					Expected: []sql.Row{{"d24fabfc-fb83-5476-8201-39e27376a62b"}},
				},
				{
					Query:    "SELECT uuid_generate_v5(uuid_ns_url(), repeat('a', 1000));",
					Expected: []sql.Row{{"7f46a8f9-f8ba-5a67-983a-ffc2101475df"}},
				},
				{
					Query:    "SELECT substring(uuid_generate_v5(uuid_nil(), 'x')::text, 15, 1);",
					Expected: []sql.Row{{"5"}},
				},
				{
					Query:    "SELECT substring(uuid_generate_v5(uuid_nil(), 'x')::text, 20, 1) IN ('8', '9', 'a', 'b');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT uuid_generate_v5(uuid_nil(), 'abc') = uuid_generate_v5(uuid_nil(), 'abc');",
					Expected: []sql.Row{{"t"}},
				},
				{ // Version 3 hashes with MD5 while version 5 hashes with SHA-1, so they'll never agree
					Query:    "SELECT uuid_generate_v3(uuid_nil(), 'abc') = uuid_generate_v5(uuid_nil(), 'abc');",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT uuid_generate_v5(NULL, 'abc') IS NULL, uuid_generate_v5(uuid_nil(), NULL) IS NULL;",
					Expected: []sql.Row{{"t", "t"}},
				},
			},
		},
		{
			Name: "uuid-ossp uuid_generate_v1 and uuid_generate_v1mc",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT substring(uuid_generate_v1()::text, 15, 1);",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substring(uuid_generate_v1mc()::text, 15, 1);",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substring(uuid_generate_v1()::text, 20, 1) IN ('8', '9', 'a', 'b');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT substring(uuid_generate_v1mc()::text, 20, 1) IN ('8', '9', 'a', 'b');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT substring(uuid_generate_v1mc()::text, 26, 1) IN ('3', '7', 'b', 'f');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT substring(uuid_generate_v1()::text, 25) = substring(uuid_generate_v1mc()::text, 25);",
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    "SELECT count(DISTINCT id::text) FROM (SELECT uuid_generate_v1() AS id FROM generate_series(1, 50)) t;",
					Expected: []sql.Row{{50}},
				},
				{
					Query:    "SELECT count(DISTINCT id::text) FROM (SELECT uuid_generate_v1mc() AS id FROM generate_series(1, 50)) t;",
					Expected: []sql.Row{{50}},
				},
				{
					Query:    "SELECT length(uuid_generate_v1()::text), length(uuid_generate_v1mc()::text);",
					Expected: []sql.Row{{36, 36}},
				},
			},
		},
		{
			Name: "uuid-ossp uuid_generate_v4",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT substring(uuid_generate_v4()::text, 15, 1);",
					Expected: []sql.Row{{"4"}},
				},
				{
					Query:    "SELECT substring(uuid_generate_v4()::text, 20, 1) IN ('8', '9', 'a', 'b');",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT count(DISTINCT id::text) FROM (SELECT uuid_generate_v4() AS id FROM generate_series(1, 100)) t;",
					Expected: []sql.Row{{100}},
				},
			},
		},
		{
			Name: "uuid-ossp functions used by a table",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
				`CREATE TABLE items (id uuid PRIMARY KEY DEFAULT uuid_generate_v4(), name text NOT NULL);`,
				`INSERT INTO items (name) VALUES ('first'), ('second'), ('third');`,
				`CREATE TABLE named (id uuid PRIMARY KEY, name text NOT NULL);`,
				`INSERT INTO named VALUES (uuid_generate_v5(uuid_ns_url(), 'example text'), 'example');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    "SELECT count(*), count(DISTINCT id::text) FROM items;",
					Expected: []sql.Row{{3, 3}},
				},
				{
					Query:    "SELECT name FROM items ORDER BY name;",
					Expected: []sql.Row{{"first"}, {"second"}, {"third"}},
				},
				{
					Query:    "SELECT id, name FROM named;",
					Expected: []sql.Row{{"59edfb26-7819-5209-86a3-79a6da9035ba", "example"}},
				},
				{
					Query:    "SELECT name FROM named WHERE id = uuid_generate_v5(uuid_ns_url(), 'example text');",
					Expected: []sql.Row{{"example"}},
				},
			},
		},
		{
			Name: "uuid-ossp catalog tables",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT extname, extrelocatable, extversion FROM pg_catalog.pg_extension WHERE extname = 'uuid-ossp';`,
					Expected: []sql.Row{{"uuid-ossp", "t", "1.1"}},
				},
				{
					Query: `SELECT name, default_version, installed_version, comment FROM pg_catalog.pg_available_extensions WHERE name = 'uuid-ossp';`,
					Expected: []sql.Row{
						{"uuid-ossp", "1.1", "1.1", "generate universally unique identifiers (UUIDs)"},
					},
				},
				{
					Query: `SELECT name, version, installed, superuser, trusted, relocatable, schema, requires FROM pg_catalog.pg_available_extension_versions WHERE name = 'uuid-ossp';`,
					Expected: []sql.Row{
						{"uuid-ossp", "1.1", "t", "t", "t", "t", nil, nil},
					},
				},
				{
					Query:    `SELECT count(*) FROM pg_catalog.pg_proc WHERE proname LIKE 'uuid_ns_%' OR proname LIKE 'uuid_generate_%' OR proname = 'uuid_nil';`,
					Expected: []sql.Row{{10}},
				},
				{
					Query: `SELECT proname FROM pg_catalog.pg_proc WHERE proname LIKE 'uuid_ns_%' OR proname LIKE 'uuid_generate_%' OR proname = 'uuid_nil' ORDER BY proname;`,
					Expected: []sql.Row{
						{"uuid_generate_v1"}, {"uuid_generate_v1mc"}, {"uuid_generate_v3"}, {"uuid_generate_v4"},
						{"uuid_generate_v5"}, {"uuid_nil"}, {"uuid_ns_dns"}, {"uuid_ns_oid"}, {"uuid_ns_url"},
						{"uuid_ns_x500"},
					},
				},
			},
		},
	})
}
