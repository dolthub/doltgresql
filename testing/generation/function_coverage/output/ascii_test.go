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

func Test_Ascii(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "ascii",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ascii( '' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT ascii( ' ' ) ;",
					Expected: []sql.Row{{int32(32)}},
				},
				{
					Query:    "SELECT ascii( '0' ) ;",
					Expected: []sql.Row{{int32(48)}},
				},
				{
					Query:    "SELECT ascii( '1' ) ;",
					Expected: []sql.Row{{int32(49)}},
				},
				{
					Query:    "SELECT ascii( 'a' ) ;",
					Expected: []sql.Row{{int32(97)}},
				},
				{
					Query:    "SELECT ascii( 'abc' ) ;",
					Expected: []sql.Row{{int32(97)}},
				},
				{
					Query:    "SELECT ascii( '123' ) ;",
					Expected: []sql.Row{{int32(49)}},
				},
				{
					Query:    "SELECT ascii( 'value' ) ;",
					Expected: []sql.Row{{int32(118)}},
				},
				{
					Query:    "SELECT ascii( '12345' ) ;",
					Expected: []sql.Row{{int32(49)}},
				},
				{
					Query:    "SELECT ascii( 'something' ) ;",
					Expected: []sql.Row{{int32(115)}},
				},
				{
					Query:    "SELECT ascii( ' something' ) ;",
					Expected: []sql.Row{{int32(32)}},
				},
				{
					Query:    "SELECT ascii( 'something ' ) ;",
					Expected: []sql.Row{{int32(115)}},
				},
				{
					Query:    "SELECT ascii( '123456789' ) ;",
					Expected: []sql.Row{{int32(49)}},
				},
				{
					Query:    "SELECT ascii( 'a group of words' ) ;",
					Expected: []sql.Row{{int32(97)}},
				},
				{
					Query:    "SELECT ascii( '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(49)}},
				},
			},
		},
	})
}
