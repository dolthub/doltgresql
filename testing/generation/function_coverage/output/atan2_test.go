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

func Test_Atan2(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "atan2",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT atan2( 0::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-2.356194)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(2.356194)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(2.034444)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-2.034444)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.759018)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.662534)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.670465)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.580796)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.570844)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.570806)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570797)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.570797)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.785398)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(1.107149)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.107149)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(1.382575)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(1.479058)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.471128)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(1.560797)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(1.570749)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(1.570786)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570795)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.463648)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(0.463648)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.785398)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(1.206817)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(1.388839)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.373401)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(1.550799)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(1.570701)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(1.570776)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570795)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-2.677945)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(2.677945)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(2.356194)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-2.356194)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(1.934775)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(1.752754)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.768192)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(1.590794)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(1.570891)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(1.570816)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570798)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(1.570797)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-0.188222)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.188222)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.363979)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-0.363979)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.120857)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-1.087349)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.518344)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.570547)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.570744)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570792)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.570794)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-0.091738)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.091738)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.181958)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-0.181958)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.449940)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-0.743736)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.462521)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.570280)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.570688)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570787)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.570792)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-3.041924)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(3.041924)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(2.944197)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-2.944197)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(2.658146)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(2.314532)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-2.356194)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(1.670465)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(1.571271)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(1.570896)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570805)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(1.570800)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-0.010000)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(0.010000)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(0.019997)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-0.019997)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(0.052452)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(0.108275)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-0.099669)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(1.566046)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(1.569796)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570712)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(1.570757)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570796)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(1.570796)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000048)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.000048)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.000095)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000095)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.000249)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.000516)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000475)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.004750)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(1.363321)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-1.553023)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(1.562461)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570787)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(1.570787)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000010)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.000010)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.000020)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000020)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.000052)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.000109)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000100)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.001000)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.207476)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-1.486557)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(1.531217)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-1.570750)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(1.570750)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-3.141592)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(3.141592)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(3.141591)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-3.141591)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(3.141588)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(3.141583)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-3.141584)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(3.141508)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(3.123820)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(3.057353)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-2.356194)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(2.009313)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-1.571348)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(1.571348)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000000)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000001)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000001)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000002)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000004)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000004)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000040)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.008336)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.039579)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.438517)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-1.569620)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(1.569620)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-3.141593)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-3.141593)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-3.141593)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.141593)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.141583)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.141546)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-3.141041)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(3.140417)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-2.356194)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(2.356194)}},
				},
				{
					Query:    "SELECT atan2( 0::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -1::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000000)}},
				},
				{
					Query:    "SELECT atan2( 1::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( 2::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -2::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000000)}},
				},
				{
					Query:    "SELECT atan2( 5.25::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( 10.87::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( -10::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000000)}},
				},
				{
					Query:    "SELECT atan2( 100::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2( 21050.48::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000010)}},
				},
				{
					Query:    "SELECT atan2( 100000::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000047)}},
				},
				{
					Query:    "SELECT atan2( -1184280::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000551)}},
				},
				{
					Query:    "SELECT atan2( 2525280.279::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.001176)}},
				},
				{
					Query:    "SELECT atan2( -2147483648::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.785398)}},
				},
				{
					Query:    "SELECT atan2( 2147483647.59024::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.785398)}},
				},
			},
		},
	})
}
