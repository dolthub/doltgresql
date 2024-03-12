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

func Test_Cosd(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "cosd",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT cosd( 0::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT cosd( -1::float8 ) ;",
					Expected: []sql.Row{{float64(0.999848)}},
				},
				{
					Query:    "SELECT cosd( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.999848)}},
				},
				{
					Query:    "SELECT cosd( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.999391)}},
				},
				{
					Query:    "SELECT cosd( -2::float8 ) ;",
					Expected: []sql.Row{{float64(0.999391)}},
				},
				{
					Query:    "SELECT cosd( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.995805)}},
				},
				{
					Query:    "SELECT cosd( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.982058)}},
				},
				{
					Query:    "SELECT cosd( -10::float8 ) ;",
					Expected: []sql.Row{{float64(0.984808)}},
				},
				{
					Query:    "SELECT cosd( 100::float8 ) ;",
					Expected: []sql.Row{{float64(-0.173648)}},
				},
				{
					Query:    "SELECT cosd( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.986228)}},
				},
				{
					Query:    "SELECT cosd( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.173648)}},
				},
				{
					Query:    "SELECT cosd( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-0.500000)}},
				},
				{
					Query:    "SELECT cosd( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.495777)}},
				},
				{
					Query:    "SELECT cosd( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-0.615661)}},
				},
				{
					Query:    "SELECT cosd( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.610010)}},
				},
			},
		},
	})
}
