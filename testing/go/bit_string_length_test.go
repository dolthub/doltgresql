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

func TestBitStringLengthFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "length and bit_length for bit strings",
			SetUpScript: []string{
				"CREATE TABLE bit_lengths (v varbit, b bit(5));",
				"INSERT INTO bit_lengths VALUES (B'101', B'00101'), (B'', B'00000'), (NULL, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT length('101'::varbit), bit_length('101'::varbit);",
					Expected: []sql.Row{{int32(3), int32(3)}},
				},
				{
					Query:    "SELECT length(B'10101'::bit(5)), bit_length(B'10101'::bit(5));",
					Expected: []sql.Row{{int32(5), int32(5)}},
				},
				{
					Query:    "SELECT length(''::varbit), bit_length(''::varbit);",
					Expected: []sql.Row{{int32(0), int32(0)}},
				},
				{
					Query:    "SELECT length(NULL::varbit), bit_length(NULL::varbit);",
					Expected: []sql.Row{{nil, nil}},
				},
				{
					Query:    "SELECT length(v), bit_length(v), length(b), bit_length(b) FROM bit_lengths WHERE v IS NOT NULL ORDER BY v;",
					Expected: []sql.Row{{int32(0), int32(0), int32(5), int32(5)}, {int32(3), int32(3), int32(5), int32(5)}},
				},
				{
					Query:    "SELECT length(v), bit_length(v), length(b), bit_length(b) FROM bit_lengths WHERE v IS NULL;",
					Expected: []sql.Row{{nil, nil, nil, nil}},
				},
			},
		},
	})
}
