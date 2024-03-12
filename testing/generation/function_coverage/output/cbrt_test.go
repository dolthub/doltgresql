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

func Test_Cbrt(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "cbrt",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT cbrt( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT cbrt( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.000000)}},
				},
				{
					Query:    "SELECT cbrt( 1::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT cbrt( 2::float8 ) ;",
					Expected: []sql.Row{{float64(1.259921)}},
				},
				{
					Query:    "SELECT cbrt( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.259921)}},
				},
				{
					Query:    "SELECT cbrt( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.738013)}},
				},
				{
					Query:    "SELECT cbrt( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(2.215184)}},
				},
				{
					Query:    "SELECT cbrt( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-2.154435)}},
				},
				{
					Query:    "SELECT cbrt( 100::float8 ) ;",
					Expected: []sql.Row{{float64(4.641589)}},
				},
				{
					Query:    "SELECT cbrt( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(27.611331)}},
				},
				{
					Query:    "SELECT cbrt( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(46.415888)}},
				},
				{
					Query:    "SELECT cbrt( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-105.799788)}},
				},
				{
					Query:    "SELECT cbrt( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(136.176822)}},
				},
				{
					Query:    "SELECT cbrt( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-1290.159155)}},
				},
				{
					Query:    "SELECT cbrt( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(1290.159155)}},
				},
			},
		},
	})
}
