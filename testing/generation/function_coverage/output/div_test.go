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

func Test_Div(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "div",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT div( 0::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( -1::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 1::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 2::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( -2::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 5::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 10::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( -10::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 1000::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 2105076::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( -5184226581::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 8525267290::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( -79223372036854775808::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT div( 79223372036854775807::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT div( 0::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1000")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2105076")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-100000000")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("5184226581")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-8525267290")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775808")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-79223372036854775807")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000000")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5184226581")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("8525267290")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-79223372036854775808")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775807")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("500")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1052538")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("50000000")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2592113290")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("4262633645")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-39611686018427387904")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("39611686018427387903")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-500")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1052538")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-50000000")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("2592113290")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-4262633645")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("39611686018427387904")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-39611686018427387903")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("200")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("421015")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("20000000")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1036845316")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("1705053458")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("-15844674407370955161")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("15844674407370955161")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("100")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("210507")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("10000000")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-518422658")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("852526729")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-7922337203685477580")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("7922337203685477580")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-100")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-210507")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10000000")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("518422658")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-852526729")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("7922337203685477580")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-7922337203685477580")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5184226")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("8525267")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("-79223372036854775")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("47")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2462")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("4049")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("-37634447419881")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("37634447419881")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("-51")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("85")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("-792233718510")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("792233718510")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("15281618347")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-15281618347")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("-9292772806")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("9292772806")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  -79223372036854775808::numeric ) ;",
					Skip:     true,
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 0::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -1::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -2::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 5::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 10::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -10::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 1000::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 2105076::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -5184226581::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( 8525267290::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT div( -79223372036854775808::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT div( 79223372036854775807::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
			},
		},
	})
}
