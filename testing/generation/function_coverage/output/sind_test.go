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

func Test_Sind(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "sind",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT sind( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT sind( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.017452)}},
				},
				{
					Query:    "SELECT sind( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.017452)}},
				},
				{
					Query:    "SELECT sind( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.034899)}},
				},
				{
					Query:    "SELECT sind( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.034899)}},
				},
				{
					Query:    "SELECT sind( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.091502)}},
				},
				{
					Query:    "SELECT sind( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.188581)}},
				},
				{
					Query:    "SELECT sind( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-0.173648)}},
				},
				{
					Query:    "SELECT sind( 100::float8 ) ;",
					Expected: []sql.Row{{float64(0.984808)}},
				},
				{
					Query:    "SELECT sind( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.165392)}},
				},
				{
					Query:    "SELECT sind( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.984808)}},
				},
				{
					Query:    "SELECT sind( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(0.866025)}},
				},
				{
					Query:    "SELECT sind( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.868450)}},
				},
				{
					Query:    "SELECT sind( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-0.788011)}},
				},
				{
					Query:    "SELECT sind( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.792394)}},
				},
			},
		},
	})
}
