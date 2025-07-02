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
	"runtime"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCreateExtension(t *testing.T) {
	if runtime.GOOS == "windows" && os.Getenv("CI") != "" {
		t.Skip("CI Postgres installation seems to behave weirdly, skipping for now") // TODO: look into this a bit more
	}
	RunScripts(t, []ScriptTest{
		{
			Name: "Extension Test: uuid-ossp",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT uuid_ns_url();",
					Expected: []sql.Row{{"6ba7b811-9dad-11d1-80b4-00c04fd430c8"}},
				},
				{
					Skip:     true, // This is returning different results on different platforms for some reason
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
