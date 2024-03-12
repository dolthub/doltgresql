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

func Test_Sqrt(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "sqrt",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT sqrt( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:       "SELECT sqrt( -1::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 1::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT sqrt( 2::float8 ) ;",
					Expected: []sql.Row{{float64(1.414214)}},
				},
				{
					Query:       "SELECT sqrt( -2::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(2.291288)}},
				},
				{
					Query:    "SELECT sqrt( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(3.296968)}},
				},
				{
					Query:       "SELECT sqrt( -10::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 100::float8 ) ;",
					Expected: []sql.Row{{float64(10.000000)}},
				},
				{
					Query:    "SELECT sqrt( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(145.087835)}},
				},
				{
					Query:    "SELECT sqrt( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(316.227766)}},
				},
				{
					Query:       "SELECT sqrt( -1184280::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(1589.113048)}},
				},
				{
					Query:       "SELECT sqrt( -2147483648::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(46340.950007)}},
				},
				{
					Query:    "SELECT sqrt( 0::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:       "SELECT sqrt( -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.000000000000000")}},
				},
				{
					Query:    "SELECT sqrt( 2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.414213562373095")}},
				},
				{
					Query:       "SELECT sqrt( -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 5::numeric ) ;",
					Expected: []sql.Row{{Numeric("2.236067977499790")}},
				},
				{
					Query:    "SELECT sqrt( 10::numeric ) ;",
					Expected: []sql.Row{{Numeric("3.162277660168379")}},
				},
				{
					Query:       "SELECT sqrt( -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("31.622776601683793")}},
				},
				{
					Query:    "SELECT sqrt( 2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("1450.8880039479271")}},
				},
				{
					Query:    "SELECT sqrt( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("10000.0000117293116104028950144216538400")}},
				},
				{
					Query:       "SELECT sqrt( -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("92332.37400825346")}},
				},
				{
					Query:       "SELECT sqrt( -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT sqrt( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("8900751206.3226874")}},
				},
			},
		},
	})
}
