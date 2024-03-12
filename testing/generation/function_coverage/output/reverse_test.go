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

func Test_Reverse(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "reverse",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT reverse( '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT reverse( ' ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT reverse( '0' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT reverse( '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT reverse( 'a' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT reverse( 'abc' ) ;",
					Expected: []sql.Row{{"cba"}},
				},
				{
					Query:    "SELECT reverse( '123' ) ;",
					Expected: []sql.Row{{"321"}},
				},
				{
					Query:    "SELECT reverse( 'value' ) ;",
					Expected: []sql.Row{{"eulav"}},
				},
				{
					Query:    "SELECT reverse( '12345' ) ;",
					Expected: []sql.Row{{"54321"}},
				},
				{
					Query:    "SELECT reverse( 'something' ) ;",
					Expected: []sql.Row{{"gnihtemos"}},
				},
				{
					Query:    "SELECT reverse( ' something' ) ;",
					Expected: []sql.Row{{"gnihtemos "}},
				},
				{
					Query:    "SELECT reverse( 'something ' ) ;",
					Expected: []sql.Row{{" gnihtemos"}},
				},
				{
					Query:    "SELECT reverse( '123456789' ) ;",
					Expected: []sql.Row{{"987654321"}},
				},
				{
					Query:    "SELECT reverse( 'a group of words' ) ;",
					Expected: []sql.Row{{"sdrow fo puorg a"}},
				},
				{
					Query:    "SELECT reverse( '1234567890123456' ) ;",
					Expected: []sql.Row{{"6543210987654321"}},
				},
			},
		},
	})
}
