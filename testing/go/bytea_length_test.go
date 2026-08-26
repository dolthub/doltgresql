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

func TestByteaLengthFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "octet_length, length and bit_length for bytea",
			SetUpScript: []string{
				"CREATE TABLE byteas (id int primary key, v bytea);",
				`INSERT INTO byteas VALUES (1, '\x00010203'), (2, ''), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT octet_length('x'::bytea), length('x'::bytea), bit_length('x'::bytea);`,
					Expected: []sql.Row{{int32(1), int32(1), int32(8)}},
				},
				{
					Query:    `SELECT octet_length('\x00010203'::bytea), length('\x00010203'::bytea), bit_length('\x00010203'::bytea);`,
					Expected: []sql.Row{{int32(4), int32(4), int32(32)}},
				},
				{
					// Unlike length(text), which counts characters, length(bytea) counts bytes.
					Query:    `SELECT length('héllo'::bytea), length('héllo'::text), octet_length('héllo'::text);`,
					Expected: []sql.Row{{int32(6), int32(5), int32(6)}},
				},
				{
					Query:    `SELECT octet_length(''::bytea), length(''::bytea), bit_length(''::bytea);`,
					Expected: []sql.Row{{int32(0), int32(0), int32(0)}},
				},
				{
					Query:    `SELECT octet_length(NULL::bytea), length(NULL::bytea), bit_length(NULL::bytea);`,
					Expected: []sql.Row{{nil, nil, nil}},
				},
				{
					// The bytea overloads must not make an untyped literal argument ambiguous.
					Query:    `SELECT octet_length('abc'), length('abc'), bit_length('abc');`,
					Expected: []sql.Row{{int32(3), int32(3), int32(24)}},
				},
				{
					Query: "SELECT id, octet_length(v), length(v), bit_length(v) FROM byteas ORDER BY id;",
					Expected: []sql.Row{
						{int32(1), int32(4), int32(4), int32(32)},
						{int32(2), int32(0), int32(0), int32(0)},
						{int32(3), nil, nil, nil},
					},
				},
			},
		},
		{
			Name: "octet_length in a bytea CHECK constraint",
			SetUpScript: []string{
				"CREATE TABLE artifacts (id int primary key, artifact bytea NOT NULL, CONSTRAINT artifact_not_empty CHECK (octet_length(artifact) > 0));",
				`INSERT INTO artifacts VALUES (1, '\x0001');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT id, octet_length(artifact) FROM artifacts;",
					Expected: []sql.Row{{int32(1), int32(2)}},
				},
				{
					Query:       `INSERT INTO artifacts VALUES (2, '');`,
					ExpectedErr: "violated",
				},
			},
		},
	})
}
