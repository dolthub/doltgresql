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

func Test_Asinh(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "asinh",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT asinh( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT asinh( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.881374)}},
				},
				{
					Query:    "SELECT asinh( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.881374)}},
				},
				{
					Query:    "SELECT asinh( 2::float8 ) ;",
					Expected: []sql.Row{{float64(1.443635)}},
				},
				{
					Query:    "SELECT asinh( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.443635)}},
				},
				{
					Query:    "SELECT asinh( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(2.360325)}},
				},
				{
					Query:    "SELECT asinh( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(3.081263)}},
				},
				{
					Query:    "SELECT asinh( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-2.998223)}},
				},
				{
					Query:    "SELECT asinh( 100::float8 ) ;",
					Expected: []sql.Row{{float64(5.298342)}},
				},
				{
					Query:    "SELECT asinh( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(10.647826)}},
				},
				{
					Query:    "SELECT asinh( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(12.206073)}},
				},
				{
					Query:    "SELECT asinh( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-14.677793)}},
				},
				{
					Query:    "SELECT asinh( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(15.435010)}},
				},
				{
					Query:    "SELECT asinh( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-22.180710)}},
				},
				{
					Query:    "SELECT asinh( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(22.180710)}},
				},
			},
		},
	})
}
