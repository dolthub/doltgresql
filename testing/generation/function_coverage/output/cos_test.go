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

func Test_Cos(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "cos",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT cos( 0::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT cos( -1::float8 ) ;",
					Expected: []sql.Row{{float64(0.540302)}},
				},
				{
					Query:    "SELECT cos( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.540302)}},
				},
				{
					Query:    "SELECT cos( 2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.416147)}},
				},
				{
					Query:    "SELECT cos( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.416147)}},
				},
				{
					Query:    "SELECT cos( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.512085)}},
				},
				{
					Query:    "SELECT cos( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-0.125245)}},
				},
				{
					Query:    "SELECT cos( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-0.839072)}},
				},
				{
					Query:    "SELECT cos( 100::float8 ) ;",
					Expected: []sql.Row{{float64(0.862319)}},
				},
				{
					Query:    "SELECT cos( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.236172)}},
				},
				{
					Query:    "SELECT cos( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.999361)}},
				},
				{
					Query:    "SELECT cos( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(0.994948)}},
				},
				{
					Query:    "SELECT cos( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.531019)}},
				},
				{
					Query:    "SELECT cos( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(0.237816)}},
				},
				{
					Query:    "SELECT cos( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.168831)}},
				},
			},
		},
	})
}
