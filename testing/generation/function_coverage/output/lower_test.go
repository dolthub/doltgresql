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

func Test_Lower(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "lower",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT lower( '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT lower( ' ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT lower( '0' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT lower( '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT lower( 'a' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT lower( 'abc' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT lower( '123' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT lower( 'value' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT lower( '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT lower( 'something' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT lower( ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT lower( 'something ' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT lower( '123456789' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT lower( 'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT lower( '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
			},
		},
	})
}
