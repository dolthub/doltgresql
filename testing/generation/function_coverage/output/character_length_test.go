// Copyright 2024 Dolthub, Inc.
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

package output

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func Test_CharacterLength(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "character_length",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT character_length( '' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT character_length( ' ' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT character_length( '0' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT character_length( '1' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT character_length( 'a' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT character_length( 'abc' ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT character_length( '123' ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT character_length( 'value' ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT character_length( '12345' ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT character_length( 'something' ) ;",
					Expected: []sql.Row{{int32(9)}},
				},
				{
					Query:    "SELECT character_length( ' something' ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT character_length( 'something ' ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT character_length( '123456789' ) ;",
					Expected: []sql.Row{{int32(9)}},
				},
				{
					Query:    "SELECT character_length( 'a group of words' ) ;",
					Expected: []sql.Row{{int32(16)}},
				},
				{
					Query:    "SELECT character_length( '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(16)}},
				},
			},
		},
	})
}
