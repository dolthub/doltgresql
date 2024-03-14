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

func Test_Log(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "log",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT log( 0::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT log( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.301030)}},
				},
				{
					Query:       "SELECT log( -2::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.720159)}},
				},
				{
					Query:    "SELECT log( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.036230)}},
				},
				{
					Query:       "SELECT log( -10::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 100::float8 ) ;",
					Expected: []sql.Row{{float64(2.000000)}},
				},
				{
					Query:    "SELECT log( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(4.323262)}},
				},
				{
					Query:    "SELECT log( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(5.000000)}},
				},
				{
					Query:       "SELECT log( -1184280::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(6.402310)}},
				},
				{
					Query:       "SELECT log( -2147483648::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(9.331930)}},
				},
				{
					Query:       "SELECT log( 0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT log( 2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.3010299956639812")}},
				},
				{
					Query:       "SELECT log( -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.6989700043360188")}},
				},
				{
					Query:    "SELECT log( 10::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
				{
					Query:       "SELECT log( -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("3.0000000000000000")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("6.323267779879430")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("8.0000000010187950611868560912011483")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("9.930708004175069")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("19.898853323625356")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -2::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 5::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 10::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -10::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1000::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2105076::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 8525267290::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 79223372036854775807::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 0::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -2::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 5::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 10::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -10::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1000::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2105076::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 8525267290::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 79223372036854775807::numeric ,  -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 0::numeric ,  1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.4306765580733931")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.3010299956639812")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.1003433318879937")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.04760671319690989")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.0376287494532056513890219206790570")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.03031304470309893")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.01512800716544690")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -2::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 5::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 10::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -10::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1000::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2105076::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 8525267290::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 79223372036854775807::numeric ,  -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 0::numeric ,  5::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  5::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  5::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("2.3219280948873623")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  5::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.6989700043360188")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  5::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2329900014453396")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.11053936487715004")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.0873712505308756757819606860532816")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  5::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.07038471013770194")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  5::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.03512614485710848")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("3.3219280948873623")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.4306765580733931")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.3333333333333333")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.15814607807405993")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.1249999999840813271709826067323386")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.10069775484080087")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.05025415202255538")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -2::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 5::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 10::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -10::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1000::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2105076::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 8525267290::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 79223372036854775807::numeric ,  -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 0::numeric ,  1000::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  1000::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  1000::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("9.9657842846620870")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  1000::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("4.2920296742201792")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("3.0000000000000000")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  1000::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.47443823422217980")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.3749999999522439815129478201970158")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  1000::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.30209326452240261")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  1000::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.15076245606766613")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  2105076::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  2105076::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  2105076::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("21.005440889477517")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  2105076::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("9.046550983094289")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("6.323267779879430")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  2105076::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("2.107755926626477")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.7904084723842707522222123172204982")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  2105076::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.6367388686910341")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  2105076::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.3177704602893871")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("26.5754247624832627196516620463046965")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("11.4454124660447106168818356950608592")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("8.0000000010187950611868560912011483")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("2.6666666670062650203956186970670494")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.2651686247535979104206679916202281")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000000000000000000000")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.8055820388289973409808674904634879")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.4020332162316417083118994177179252")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -2::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 5::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 10::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -10::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1000::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2105076::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 8525267290::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 79223372036854775807::numeric ,  -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 0::numeric ,  8525267290::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  8525267290::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  8525267290::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("32.989097921191967")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  8525267290::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("14.207631146645082")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("9.930708004175069")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  8525267290::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("3.310236001391690")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.5705025233589623")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.2413385003637999236785766706862784")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  8525267290::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  8525267290::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.4990593097334214")}},
				},
				{
					Query:       "SELECT log( 0::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -2::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 5::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 10::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -10::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1000::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 2105076::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 8525267290::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 79223372036854775807::numeric ,  -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 0::numeric ,  79223372036854775807::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( -1::numeric ,  79223372036854775807::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log( 1::numeric ,  79223372036854775807::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 2::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("66.102559911793837")}},
				},
				{
					Query:       "SELECT log( -2::numeric ,  79223372036854775807::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 5::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("28.468822982651622")}},
				},
				{
					Query:    "SELECT log( 10::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("19.898853323625356")}},
				},
				{
					Query:       "SELECT log( -10::numeric ,  79223372036854775807::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 1000::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("6.632951107875119")}},
				},
				{
					Query:    "SELECT log( 2105076::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("3.1469256113023226")}},
				},
				{
					Query:    "SELECT log( 100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("2.4873566651364061742271905346930906")}},
				},
				{
					Query:       "SELECT log( -5184226581::numeric ,  79223372036854775807::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 8525267290::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("2.0037698535954817")}},
				},
				{
					Query:       "SELECT log( -79223372036854775808::numeric ,  79223372036854775807::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log( 79223372036854775807::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
			},
		},
	})
}
