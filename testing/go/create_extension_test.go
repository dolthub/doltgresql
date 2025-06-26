// Copyright 2025 Dolthub, Inc.
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
	"os"
	"testing"

	"github.com/dolthub/doltgresql/core/extensions"

	"github.com/dolthub/go-mysql-server/sql"
)

// SkipIfExtensionNotSupported skips if an extension (or set of extensions) is not installed locally. Missing extensions
// aren't meaningful failures, since it relies on external factors. This can still cause failures due to differences
// between versions, platforms, etc., but in that case we should write more generic tests (or use stable extensions) for
// our testing purposes.
func SkipIfExtensionNotSupported(t *testing.T, exts ...string) {
	for _, ext := range exts {
		_, err := extensions.GetExtension(ext)
		if err != nil {
			t.Skipf("Skipping test since extension `%s` is not installed", ext)
		}
	}
}

func TestCreateExtension(t *testing.T) {
	if os.Getenv("CI") != "" {
		SkipIfExtensionNotSupported(t, "uuid-ossp")
	}
	RunScripts(t, []ScriptTest{
		{
			Name: "Smoke Test",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT uuid_ns_url();",
					Expected: []sql.Row{{"6ba7b811-9dad-11d1-80b4-00c04fd430c8"}},
				},
				{
					Query:    "SELECT uuid_generate_v3('00000000-0000-0000-0000-000000000000'::uuid, 'example text');",
					Expected: []sql.Row{{"a55b875a-1bd9-31af-ac66-7d8323785c6e"}},
				},
				{
					Skip:     true, // For some reason, this returns the same result as above
					Query:    "SELECT uuid_generate_v3('00000000-0000-0000-0000-000000000001'::uuid, 'example text');",
					Expected: []sql.Row{{"a319ab51-8e26-37c6-942f-7dd5fda5c3ef"}},
				},
				{
					Skip:     true, // Need to figure out why the result is wrong
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
	})
}
