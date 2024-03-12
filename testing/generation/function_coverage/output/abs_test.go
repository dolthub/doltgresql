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

func Test_Abs(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "abs",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT abs( 0::int2 ) ;",
					Expected: []sql.Row{{int16(0)}},
				},
				{
					Query:    "SELECT abs( -1::int2 ) ;",
					Expected: []sql.Row{{int16(1)}},
				},
				{
					Query:    "SELECT abs( 1::int2 ) ;",
					Expected: []sql.Row{{int16(1)}},
				},
				{
					Query:    "SELECT abs( 2::int2 ) ;",
					Expected: []sql.Row{{int16(2)}},
				},
				{
					Query:    "SELECT abs( -2::int2 ) ;",
					Expected: []sql.Row{{int16(2)}},
				},
				{
					Query:    "SELECT abs( 5::int2 ) ;",
					Expected: []sql.Row{{int16(5)}},
				},
				{
					Query:    "SELECT abs( 10::int2 ) ;",
					Expected: []sql.Row{{int16(10)}},
				},
				{
					Query:    "SELECT abs( -10::int2 ) ;",
					Expected: []sql.Row{{int16(10)}},
				},
				{
					Query:    "SELECT abs( 100::int2 ) ;",
					Expected: []sql.Row{{int16(100)}},
				},
				{
					Query:    "SELECT abs( 2105::int2 ) ;",
					Expected: []sql.Row{{int16(2105)}},
				},
				{
					Query:    "SELECT abs( 10000::int2 ) ;",
					Expected: []sql.Row{{int16(10000)}},
				},
				{
					Query:    "SELECT abs( -11842::int2 ) ;",
					Expected: []sql.Row{{int16(11842)}},
				},
				{
					Query:    "SELECT abs( 25252::int2 ) ;",
					Expected: []sql.Row{{int16(25252)}},
				},
				{
					Query:       "SELECT abs( -32768::int2 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT abs( 32767::int2 ) ;",
					Expected: []sql.Row{{int16(32767)}},
				},
				{
					Query:    "SELECT abs( 0::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT abs( -1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT abs( 1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT abs( 2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT abs( -2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT abs( 5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT abs( 10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT abs( -10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT abs( 100::int4 ) ;",
					Expected: []sql.Row{{int32(100)}},
				},
				{
					Query:    "SELECT abs( 21050::int4 ) ;",
					Expected: []sql.Row{{int32(21050)}},
				},
				{
					Query:    "SELECT abs( 100000::int4 ) ;",
					Expected: []sql.Row{{int32(100000)}},
				},
				{
					Query:    "SELECT abs( -1184280::int4 ) ;",
					Expected: []sql.Row{{int32(1184280)}},
				},
				{
					Query:    "SELECT abs( 2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525280)}},
				},
				{
					Query:       "SELECT abs( -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT abs( 2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2147483647)}},
				},
				{
					Query:    "SELECT abs( 0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT abs( -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT abs( 1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT abs( 2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT abs( -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT abs( 5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT abs( 10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT abs( -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT abs( 1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT abs( 2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT abs( 100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT abs( -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:    "SELECT abs( 8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT abs( -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT abs( 9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
				{
					Query:    "SELECT abs( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT abs( -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT abs( 1::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT abs( 2::float8 ) ;",
					Expected: []sql.Row{{float64(2.000000)}},
				},
				{
					Query:    "SELECT abs( -2::float8 ) ;",
					Expected: []sql.Row{{float64(2.000000)}},
				},
				{
					Query:    "SELECT abs( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(5.250000)}},
				},
				{
					Query:    "SELECT abs( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(10.870000)}},
				},
				{
					Query:    "SELECT abs( -10::float8 ) ;",
					Expected: []sql.Row{{float64(10.000000)}},
				},
				{
					Query:    "SELECT abs( 100::float8 ) ;",
					Expected: []sql.Row{{float64(100.000000)}},
				},
				{
					Query:    "SELECT abs( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(21050.480000)}},
				},
				{
					Query:    "SELECT abs( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(100000.000000)}},
				},
				{
					Query:    "SELECT abs( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(1184280.000000)}},
				},
				{
					Query:    "SELECT abs( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(2525280.279000)}},
				},
				{
					Query:    "SELECT abs( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(2147483648.000000)}},
				},
				{
					Query:    "SELECT abs( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(2147483647.590240)}},
				},
				{
					Query:    "SELECT abs( 0::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT abs( -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT abs( 1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT abs( 2::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT abs( -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT abs( 5::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT abs( 10::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT abs( -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT abs( 1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT abs( 2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT abs( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000000.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT abs( -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("5184226581")}},
				},
				{
					Query:    "SELECT abs( 8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("8525267290")}},
				},
				{
					Query:    "SELECT abs( -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775808")}},
				},
				{
					Query:    "SELECT abs( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775807")}},
				},
			},
		},
	})
}
