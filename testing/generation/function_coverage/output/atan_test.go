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

func Test_Atan(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "atan",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT atan( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.785398)}},
				},
				{
					Query:    "SELECT atan( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan( 2::float8 ) ;",
					Expected: []sql.Row{{float64(1.107149)}},
				},
				{
					Query:    "SELECT atan( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.107149)}},
				},
				{
					Query:    "SELECT atan( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.382575)}},
				},
				{
					Query:    "SELECT atan( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.479058)}},
				},
				{
					Query:    "SELECT atan( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-1.471128)}},
				},
				{
					Query:    "SELECT atan( 100::float8 ) ;",
					Expected: []sql.Row{{float64(1.560797)}},
				},
				{
					Query:    "SELECT atan( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(1.570749)}},
				},
				{
					Query:    "SELECT atan( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(1.570786)}},
				},
				{
					Query:    "SELECT atan( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570795)}},
				},
				{
					Query:    "SELECT atan( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
			},
		},
	})
}
