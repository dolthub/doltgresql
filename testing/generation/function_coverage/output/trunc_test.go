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

func Test_Trunc(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "trunc",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT trunc( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT trunc( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.000000)}},
				},
				{
					Query:    "SELECT trunc( 1::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT trunc( 2::float8 ) ;",
					Expected: []sql.Row{{float64(2.000000)}},
				},
				{
					Query:    "SELECT trunc( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-2.000000)}},
				},
				{
					Query:    "SELECT trunc( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(5.000000)}},
				},
				{
					Query:    "SELECT trunc( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(10.000000)}},
				},
				{
					Query:    "SELECT trunc( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-10.000000)}},
				},
				{
					Query:    "SELECT trunc( 100::float8 ) ;",
					Expected: []sql.Row{{float64(100.000000)}},
				},
				{
					Query:    "SELECT trunc( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(21050.000000)}},
				},
				{
					Query:    "SELECT trunc( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(100000.000000)}},
				},
				{
					Query:    "SELECT trunc( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-1184280.000000)}},
				},
				{
					Query:    "SELECT trunc( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(2525280.000000)}},
				},
				{
					Query:    "SELECT trunc( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-2147483648.000000)}},
				},
				{
					Query:    "SELECT trunc( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(2147483647.000000)}},
				},
				{
					Query:    "SELECT trunc( 0::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT trunc( -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT trunc( 1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT trunc( 2::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT trunc( -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT trunc( 5::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT trunc( 10::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT trunc( -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT trunc( 1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT trunc( 2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000000")}},
				},
				{
					Query:    "SELECT trunc( -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5184226581")}},
				},
				{
					Query:    "SELECT trunc( 8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("8525267290")}},
				},
				{
					Query:    "SELECT trunc( -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("-79223372036854775808")}},
				},
				{
					Query:    "SELECT trunc( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775807")}},
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 0::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -1::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -2::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 5::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 10::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -10::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 1000::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 2105076::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 100000000.2345862323456346511423652312416532::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -5184226581::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 8525267290::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( -79223372036854775808::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT trunc( 79223372036854775807::numeric ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
			},
		},
	})
}
