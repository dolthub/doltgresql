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

func Test_Tand(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "tand",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT tand( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT tand( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.017455)}},
				},
				{
					Query:    "SELECT tand( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.017455)}},
				},
				{
					Query:    "SELECT tand( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.034921)}},
				},
				{
					Query:    "SELECT tand( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.034921)}},
				},
				{
					Query:    "SELECT tand( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.091887)}},
				},
				{
					Query:    "SELECT tand( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.192027)}},
				},
				{
					Query:    "SELECT tand( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-0.176327)}},
				},
				{
					Query:    "SELECT tand( 100::float8 ) ;",
					Expected: []sql.Row{{float64(-5.671282)}},
				},
				{
					Query:    "SELECT tand( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.167701)}},
				},
				{
					Query:    "SELECT tand( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(-5.671282)}},
				},
				{
					Query:    "SELECT tand( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-1.732051)}},
				},
				{
					Query:    "SELECT tand( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(1.751695)}},
				},
				{
					Query:    "SELECT tand( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(1.279942)}},
				},
				{
					Query:    "SELECT tand( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-1.298984)}},
				},
			},
		},
	})
}
