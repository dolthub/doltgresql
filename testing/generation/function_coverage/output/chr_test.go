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

func Test_Chr(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "chr",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT chr( 0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT chr( -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT chr( 1::int4 ) ;",
					Expected: []sql.Row{{"\x01"}},
				},
				{
					Query:    "SELECT chr( 2::int4 ) ;",
					Expected: []sql.Row{{"\x02"}},
				},
				{
					Query:       "SELECT chr( -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT chr( 5::int4 ) ;",
					Expected: []sql.Row{{"\x05"}},
				},
				{
					Query:    "SELECT chr( 10::int4 ) ;",
					Expected: []sql.Row{{"\n"}},
				},
				{
					Query:       "SELECT chr( -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT chr( 100::int4 ) ;",
					Expected: []sql.Row{{"d"}},
				},
				{
					Query:    "SELECT chr( 21050::int4 ) ;",
					Expected: []sql.Row{{"刺"}},
				},
				{
					Query:    "SELECT chr( 100000::int4 ) ;",
					Expected: []sql.Row{{"𘚠"}},
				},
				{
					Query:       "SELECT chr( -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT chr( 2525280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT chr( -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT chr( 2147483647::int4 ) ;",
					ExpectedErr: true,
				},
			},
		},
	})
}
