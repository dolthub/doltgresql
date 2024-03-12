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

func Test_Acosh(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "acosh",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT acosh( 0::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT acosh( -1::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT acosh( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT acosh( 2::float8 ) ;",
					Expected: []sql.Row{{float64(1.316958)}},
				},
				{
					Query:       "SELECT acosh( -2::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT acosh( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(2.342179)}},
				},
				{
					Query:    "SELECT acosh( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(3.077031)}},
				},
				{
					Query:       "SELECT acosh( -10::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT acosh( 100::float8 ) ;",
					Expected: []sql.Row{{float64(5.298292)}},
				},
				{
					Query:    "SELECT acosh( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(10.647826)}},
				},
				{
					Query:    "SELECT acosh( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(12.206073)}},
				},
				{
					Query:       "SELECT acosh( -1184280::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT acosh( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(15.435010)}},
				},
				{
					Query:       "SELECT acosh( -2147483648::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT acosh( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(22.180710)}},
				},
			},
		},
	})
}
