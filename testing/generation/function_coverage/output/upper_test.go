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

func Test_Upper(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "upper",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT upper( '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT upper( ' ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT upper( '0' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT upper( '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT upper( 'a' ) ;",
					Expected: []sql.Row{{"A"}},
				},
				{
					Query:    "SELECT upper( 'abc' ) ;",
					Expected: []sql.Row{{"ABC"}},
				},
				{
					Query:    "SELECT upper( '123' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT upper( 'value' ) ;",
					Expected: []sql.Row{{"VALUE"}},
				},
				{
					Query:    "SELECT upper( '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT upper( 'something' ) ;",
					Expected: []sql.Row{{"SOMETHING"}},
				},
				{
					Query:    "SELECT upper( ' something' ) ;",
					Expected: []sql.Row{{" SOMETHING"}},
				},
				{
					Query:    "SELECT upper( 'something ' ) ;",
					Expected: []sql.Row{{"SOMETHING "}},
				},
				{
					Query:    "SELECT upper( '123456789' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT upper( 'a group of words' ) ;",
					Expected: []sql.Row{{"A GROUP OF WORDS"}},
				},
				{
					Query:    "SELECT upper( '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
			},
		},
	})
}
