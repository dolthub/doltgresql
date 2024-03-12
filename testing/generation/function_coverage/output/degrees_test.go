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

func Test_Degrees(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "degrees",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT degrees( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT degrees( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-57.295780)}},
				},
				{
					Query:    "SELECT degrees( 1::float8 ) ;",
					Expected: []sql.Row{{float64(57.295780)}},
				},
				{
					Query:    "SELECT degrees( 2::float8 ) ;",
					Expected: []sql.Row{{float64(114.591559)}},
				},
				{
					Query:    "SELECT degrees( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-114.591559)}},
				},
				{
					Query:    "SELECT degrees( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(300.802842)}},
				},
				{
					Query:    "SELECT degrees( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(622.805123)}},
				},
				{
					Query:    "SELECT degrees( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-572.957795)}},
				},
				{
					Query:    "SELECT degrees( 100::float8 ) ;",
					Expected: []sql.Row{{float64(5729.577951)}},
				},
				{
					Query:    "SELECT degrees( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(1206103.660725)}},
				},
				{
					Query:    "SELECT degrees( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(5729577.951308)}},
				},
				{
					Query:    "SELECT degrees( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-67854245.761753)}},
				},
				{
					Query:    "SELECT degrees( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(144687902.074319)}},
				},
				{
					Query:    "SELECT degrees( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-123041749603.757690)}},
				},
				{
					Query:    "SELECT degrees( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(123041749580.280167)}},
				},
			},
		},
	})
}
