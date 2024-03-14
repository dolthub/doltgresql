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

func Test_Radians(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "radians",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT radians( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT radians( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.017453)}},
				},
				{
					Query:    "SELECT radians( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.017453)}},
				},
				{
					Query:    "SELECT radians( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.034907)}},
				},
				{
					Query:    "SELECT radians( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.034907)}},
				},
				{
					Query:    "SELECT radians( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.091630)}},
				},
				{
					Query:    "SELECT radians( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.189717)}},
				},
				{
					Query:    "SELECT radians( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-0.174533)}},
				},
				{
					Query:    "SELECT radians( 100::float8 ) ;",
					Expected: []sql.Row{{float64(1.745329)}},
				},
				{
					Query:    "SELECT radians( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(367.400185)}},
				},
				{
					Query:    "SELECT radians( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(1745.329252)}},
				},
				{
					Query:    "SELECT radians( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-20669.585266)}},
				},
				{
					Query:    "SELECT radians( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(44074.455404)}},
				},
				{
					Query:    "SELECT radians( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-37480660.290339)}},
				},
				{
					Query:    "SELECT radians( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(37480660.283187)}},
				},
			},
		},
	})
}
