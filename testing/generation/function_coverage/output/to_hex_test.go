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

func Test_ToHex(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "to_hex",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT to_hex( 0::int8 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT to_hex( -1::int8 ) ;",
					Expected: []sql.Row{{"ffffffffffffffff"}},
				},
				{
					Query:    "SELECT to_hex( 1::int8 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT to_hex( 2::int8 ) ;",
					Expected: []sql.Row{{"2"}},
				},
				{
					Query:    "SELECT to_hex( -2::int8 ) ;",
					Expected: []sql.Row{{"fffffffffffffffe"}},
				},
				{
					Query:    "SELECT to_hex( 5::int8 ) ;",
					Expected: []sql.Row{{"5"}},
				},
				{
					Query:    "SELECT to_hex( 10::int8 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT to_hex( -10::int8 ) ;",
					Expected: []sql.Row{{"fffffffffffffff6"}},
				},
				{
					Query:    "SELECT to_hex( 1000::int8 ) ;",
					Expected: []sql.Row{{"3e8"}},
				},
				{
					Query:    "SELECT to_hex( 2105076::int8 ) ;",
					Expected: []sql.Row{{"201ef4"}},
				},
				{
					Query:    "SELECT to_hex( 100000000::int8 ) ;",
					Expected: []sql.Row{{"5f5e100"}},
				},
				{
					Query:    "SELECT to_hex( -5184226581::int8 ) ;",
					Expected: []sql.Row{{"fffffffecafefaeb"}},
				},
				{
					Query:    "SELECT to_hex( 8525267290::int8 ) ;",
					Expected: []sql.Row{{"1fc25415a"}},
				},
				{
					Query:       "SELECT to_hex( -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT to_hex( 9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{"7fffffffffffffff"}},
				},
			},
		},
	})
}
