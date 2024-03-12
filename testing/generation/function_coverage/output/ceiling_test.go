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

func Test_Ceiling(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "ceiling",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT ceiling( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT ceiling( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.000000)}},
				},
				{
					Query:    "SELECT ceiling( 1::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT ceiling( 2::float8 ) ;",
					Expected: []sql.Row{{float64(2.000000)}},
				},
				{
					Query:    "SELECT ceiling( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-2.000000)}},
				},
				{
					Query:    "SELECT ceiling( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(6.000000)}},
				},
				{
					Query:    "SELECT ceiling( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(11.000000)}},
				},
				{
					Query:    "SELECT ceiling( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-10.000000)}},
				},
				{
					Query:    "SELECT ceiling( 100::float8 ) ;",
					Expected: []sql.Row{{float64(100.000000)}},
				},
				{
					Query:    "SELECT ceiling( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(21051.000000)}},
				},
				{
					Query:    "SELECT ceiling( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(100000.000000)}},
				},
				{
					Query:    "SELECT ceiling( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-1184280.000000)}},
				},
				{
					Query:    "SELECT ceiling( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(2525281.000000)}},
				},
				{
					Query:    "SELECT ceiling( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-2147483648.000000)}},
				},
				{
					Query:    "SELECT ceiling( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(2147483648.000000)}},
				},
				{
					Query:    "SELECT ceiling( 0::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT ceiling( -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT ceiling( 1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT ceiling( 2::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT ceiling( -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT ceiling( 5::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT ceiling( 10::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT ceiling( -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT ceiling( 1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT ceiling( 2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT ceiling( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000001")}},
				},
				{
					Query:    "SELECT ceiling( -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5184226581")}},
				},
				{
					Query:    "SELECT ceiling( 8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("8525267290")}},
				},
				{
					Query:    "SELECT ceiling( -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("-79223372036854775808")}},
				},
				{
					Query:    "SELECT ceiling( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775807")}},
				},
			},
		},
	})
}
