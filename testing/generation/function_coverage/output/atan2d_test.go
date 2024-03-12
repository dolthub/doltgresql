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

func Test_Atan2d(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "atan2d",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT atan2d( 0::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  0::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-135.000000)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(135.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(116.565051)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-116.565051)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(100.784298)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(95.256206)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-95.710593)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(90.572939)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(90.002722)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(90.000573)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000048)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(90.000023)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  -1::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-45.000000)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(63.434949)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-63.434949)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(79.215702)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(84.743794)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-84.289407)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(89.427061)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(89.997278)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(89.999427)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-89.999952)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(89.999977)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  1::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-26.565051)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(26.565051)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-45.000000)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(69.145542)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(79.574599)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-78.690068)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(88.854237)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(89.994556)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(89.998854)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-89.999903)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(89.999955)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  2::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-153.434949)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(153.434949)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(135.000000)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-135.000000)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(110.854458)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(100.425401)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-101.309932)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(91.145763)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(90.005444)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(90.001146)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000097)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(90.000045)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  -2::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-10.784298)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(10.784298)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(20.854458)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-20.854458)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(64.220355)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-62.300527)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(86.994731)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(89.985710)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(89.996992)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-89.999746)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(89.999881)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  5.25::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-5.256206)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(5.256206)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(10.425401)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-10.425401)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(25.779645)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-42.612914)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(83.796306)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(89.970414)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(89.993772)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-89.999474)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(89.999753)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  10.87::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-174.289407)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(174.289407)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(168.690068)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-168.690068)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(152.300527)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(132.612914)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-135.000000)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(95.710593)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(90.027218)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(90.005730)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000484)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(90.000227)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(-90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  -10::float8 ) ;",
					Expected: []sql.Row{{float64(90.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-0.572939)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(0.572939)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(1.145763)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-1.145763)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(3.005269)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(6.203694)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-5.710593)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(89.727819)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(89.942704)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-89.995162)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(89.997731)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(-89.999997)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  100::float8 ) ;",
					Expected: []sql.Row{{float64(89.999997)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.002722)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.002722)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.005444)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.005444)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.014290)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.029586)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-0.027218)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.272181)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(78.112522)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-88.981679)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(89.522399)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-89.999438)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(89.999438)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000573)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.000573)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.001146)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.001146)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.003008)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.006228)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.005730)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.057296)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(11.887478)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-85.173423)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(87.732297)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(-89.997332)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  100000::float8 ) ;",
					Expected: []sql.Row{{float64(89.997332)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-179.999952)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(179.999952)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(179.999903)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-179.999903)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(179.999746)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(179.999474)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-179.999516)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(179.995162)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(178.981679)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(175.173423)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-135.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(115.125155)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-90.031597)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(90.031597)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000023)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000023)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000045)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000045)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000119)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.000247)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000227)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.002269)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(0.477601)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(2.267703)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-25.125155)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-89.932624)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(89.932624)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-180.000000)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-180.000000)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(180.000000)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-180.000000)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(179.999997)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(179.999438)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(179.997332)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-179.968403)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(179.932624)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(-135.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(135.000000)}},
				},
				{
					Query:    "SELECT atan2d( 0::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -1::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000000)}},
				},
				{
					Query:    "SELECT atan2d( 1::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -2::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000000)}},
				},
				{
					Query:    "SELECT atan2d( 5.25::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( 10.87::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT atan2d( -10::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.000000)}},
				},
				{
					Query:    "SELECT atan2d( 100::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000003)}},
				},
				{
					Query:    "SELECT atan2d( 21050.48::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.000562)}},
				},
				{
					Query:    "SELECT atan2d( 100000::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.002668)}},
				},
				{
					Query:    "SELECT atan2d( -1184280::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.031597)}},
				},
				{
					Query:    "SELECT atan2d( 2525280.279::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(0.067376)}},
				},
				{
					Query:    "SELECT atan2d( -2147483648::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-45.000000)}},
				},
				{
					Query:    "SELECT atan2d( 2147483647.59024::float8 ,  2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(45.000000)}},
				},
			},
		},
	})
}
