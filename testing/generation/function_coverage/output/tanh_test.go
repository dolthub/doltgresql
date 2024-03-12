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

func Test_Tanh(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "tanh",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT tanh( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT tanh( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.761594)}},
				},
				{
					Query:    "SELECT tanh( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.761594)}},
				},
				{
					Query:    "SELECT tanh( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.964028)}},
				},
				{
					Query:    "SELECT tanh( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.964028)}},
				},
				{
					Query:    "SELECT tanh( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.999945)}},
				},
				{
					Query:    "SELECT tanh( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT tanh( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-1.000000)}},
				},
				{
					Query:    "SELECT tanh( 100::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT tanh( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT tanh( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT tanh( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-1.000000)}},
				},
				{
					Query:    "SELECT tanh( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT tanh( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-1.000000)}},
				},
				{
					Query:    "SELECT tanh( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
			},
		},
	})
}
