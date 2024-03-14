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

func Test_Atand(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "atand",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT atand( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atand( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-45.000000)}},
				},
				{
					Query:    "SELECT atand( 1::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atand( 2::float8 ) ;",
					Expected: []sql.Row{{float64(63.434949)}},
				},
				{
					Query:    "SELECT atand( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-63.434949)}},
				},
				{
					Query:    "SELECT atand( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(79.215702)}},
				},
				{
					Query:    "SELECT atand( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(84.743794)}},
				},
				{
					Query:    "SELECT atand( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-84.289407)}},
				},
				{
					Query:    "SELECT atand( 100::float8 ) ;",
					Expected: []sql.Row{{float64(89.427061)}},
				},
				{
					Query:    "SELECT atand( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(89.997278)}},
				},
				{
					Query:    "SELECT atand( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(89.999427)}},
				},
				{
					Query:    "SELECT atand( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-89.999952)}},
				},
				{
					Query:    "SELECT atand( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(89.999977)}},
				},
				{
					Query:    "SELECT atand( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atand( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
			},
		},
	})
}
