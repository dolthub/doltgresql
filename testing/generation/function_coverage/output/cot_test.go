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

func Test_Cot(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "cot",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT cot( 0::float8 ) ;",
					Expected: []sql.Row{{float64(+Inf)}},
				},
				{
					Query:    "SELECT cot( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.642093)}},
				},
				{
					Query:    "SELECT cot( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.642093)}},
				},
				{
					Query:    "SELECT cot( 2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.457658)}},
				},
				{
					Query:    "SELECT cot( -2::float8 ) ;",
					Expected: []sql.Row{{float64(0.457658)}},
				},
				{
					Query:    "SELECT cot( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-0.596187)}},
				},
				{
					Query:    "SELECT cot( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.126239)}},
				},
				{
					Query:    "SELECT cot( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-1.542351)}},
				},
				{
					Query:    "SELECT cot( 100::float8 ) ;",
					Expected: []sql.Row{{float64(-1.702957)}},
				},
				{
					Query:    "SELECT cot( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.243048)}},
				},
				{
					Query:    "SELECT cot( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(-27.955088)}},
				},
				{
					Query:    "SELECT cot( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-9.910614)}},
				},
				{
					Query:    "SELECT cot( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.626674)}},
				},
				{
					Query:    "SELECT cot( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(0.244841)}},
				},
				{
					Query:    "SELECT cot( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.171289)}},
				},
			},
		},
	})
}
