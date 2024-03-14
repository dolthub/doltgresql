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

func Test_Initcap(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "initcap",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT initcap( '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT initcap( ' ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT initcap( '0' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT initcap( '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT initcap( 'a' ) ;",
					Expected: []sql.Row{{"A"}},
				},
				{
					Query:    "SELECT initcap( 'abc' ) ;",
					Expected: []sql.Row{{"Abc"}},
				},
				{
					Query:    "SELECT initcap( '123' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT initcap( 'value' ) ;",
					Expected: []sql.Row{{"Value"}},
				},
				{
					Query:    "SELECT initcap( '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT initcap( 'something' ) ;",
					Expected: []sql.Row{{"Something"}},
				},
				{
					Query:    "SELECT initcap( ' something' ) ;",
					Expected: []sql.Row{{" Something"}},
				},
				{
					Query:    "SELECT initcap( 'something ' ) ;",
					Expected: []sql.Row{{"Something "}},
				},
				{
					Query:    "SELECT initcap( '123456789' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT initcap( 'a group of words' ) ;",
					Expected: []sql.Row{{"A Group Of Words"}},
				},
				{
					Query:    "SELECT initcap( '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
			},
		},
	})
}
