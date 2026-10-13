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

func TestStartsWith(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "starts_with",
		Assertions: []ScriptTestAssertion{
			{
				Query:    `SELECT starts_with('a.b', 'a.'), starts_with('axb', 'a.'), starts_with('abc', 'a%');`,
				Expected: []sql.Row{{"t", "f", "f"}},
			},
			{
				Query:    `SELECT starts_with('alphabet', 'alph');`,
				Expected: []sql.Row{{"t"}},
			},
			{
				Query:    `SELECT starts_with('alphabet', 'pha'), starts_with('alphabet', 'alphabetical');`,
				Expected: []sql.Row{{"f", "f"}},
			},
			{
				Query:    `SELECT starts_with('alphabet', 'alphabet'), starts_with('Alphabet', 'alph');`,
				Expected: []sql.Row{{"t", "f"}},
			},
			{
				Query:    `SELECT starts_with('', ''), starts_with('abc', ''), starts_with('', 'a');`,
				Expected: []sql.Row{{"t", "t", "f"}},
			},
			{
				Query:    `SELECT starts_with(NULL, 'a'), starts_with('abc', NULL), starts_with(NULL, NULL);`,
				Expected: []sql.Row{{nil, nil, nil}},
			},
			{
				Query:    `SELECT starts_with('éclair', 'é'), starts_with('日本語', '日本'), starts_with('éclair', 'e');`,
				Expected: []sql.Row{{"t", "t", "f"}},
			},
			{
				Query:    `SELECT pg_catalog.starts_with('alphabet'::varchar, 'alph'::text), starts_with('alphabet'::name, 'alph');`,
				Expected: []sql.Row{{"t", "t"}},
			},
			{
				Query:    `SELECT starts_with('abc '::text, 'abc '), starts_with('abc '::char(4), 'abc ');`,
				Expected: []sql.Row{{"t", "f"}},
			},
			{
				Query:    `SELECT starts_with(nspname, 'pg_') FROM pg_catalog.pg_namespace WHERE nspname = 'pg_catalog';`,
				Expected: []sql.Row{{"t"}},
			},
		},
	}})
}

// TestStartsWithErrors verifies argument validation for the text-only signature.
func TestStartsWithErrors(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "starts_with invalid arguments",
		Assertions: []ScriptTestAssertion{
			{
				Query:       `SELECT starts_with();`,
				ExpectedErr: "function starts_with() does not exist",
			},
			{
				Query:       `SELECT starts_with('abc'::text);`,
				ExpectedErr: "function starts_with(text) does not exist",
			},
			{
				Query:       `SELECT starts_with('abc'::text, 'a'::text, 'b'::text);`,
				ExpectedErr: "function starts_with(text, text, text) does not exist",
			},
			{
				Query:       `SELECT starts_with(123, '1'::text);`,
				ExpectedErr: "function starts_with(integer, text) does not exist",
			},
			{
				Query:       `SELECT starts_with('true'::text, true);`,
				ExpectedErr: "function starts_with(text, boolean) does not exist",
			},
		},
	}})
}
