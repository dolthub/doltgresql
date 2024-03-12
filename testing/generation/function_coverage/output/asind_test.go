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

func Test_Asind(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "asind",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT asind( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT asind( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT asind( 1::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:       "SELECT asind( 2::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( -2::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( 5.25::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( 10.87::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( -10::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( 100::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( 21050.48::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( 100000::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( -1184280::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( 2525280.279::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( -2147483648::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT asind( 2147483647.59024::float8 ) ;",
					ExpectedErr: true,
				},
			},
		},
	})
}
