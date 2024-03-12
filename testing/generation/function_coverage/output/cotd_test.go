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

func Test_Cotd(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "cotd",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT cotd( 0::float8 ) ;",
					Expected: []sql.Row{{float64(+Inf)}},
				},
				{
					Query:    "SELECT cotd( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-57.289962)}},
				},
				{
					Query:    "SELECT cotd( 1::float8 ) ;",
					Expected: []sql.Row{{float64(57.289962)}},
				},
				{
					Query:    "SELECT cotd( 2::float8 ) ;",
					Expected: []sql.Row{{float64(28.636253)}},
				},
				{
					Query:    "SELECT cotd( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-28.636253)}},
				},
				{
					Query:    "SELECT cotd( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(10.882921)}},
				},
				{
					Query:    "SELECT cotd( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(5.207610)}},
				},
				{
					Query:    "SELECT cotd( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-5.671282)}},
				},
				{
					Query:    "SELECT cotd( 100::float8 ) ;",
					Expected: []sql.Row{{float64(-0.176327)}},
				},
				{
					Query:    "SELECT cotd( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-5.962977)}},
				},
				{
					Query:    "SELECT cotd( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.176327)}},
				},
				{
					Query:    "SELECT cotd( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-0.577350)}},
				},
				{
					Query:    "SELECT cotd( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.570876)}},
				},
				{
					Query:    "SELECT cotd( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(0.781286)}},
				},
				{
					Query:    "SELECT cotd( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.769832)}},
				},
			},
		},
	})
}
