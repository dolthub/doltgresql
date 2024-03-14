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

func Test_Lcm(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "lcm",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT lcm( 0::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 9223372036854775807::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 9223372036854775807::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 9223372036854775807::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(10368453162)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(10368453162)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(10525380)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(25921132905)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10525380)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(51842265810)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10525380)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(51842265810)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(526269000)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581000)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(852526729000)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(10525380)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(10525380)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(10525380)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(526269000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(52626900000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(3637730318075052)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(8973167782882020)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(52626900000000)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(518422658100000000)}},
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(85252672900000000)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(10368453162)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(10368453162)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(25921132905)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(51842265810)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(51842265810)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(3637730318075052)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(518422658100000000)}},
				},
				{
					Query:    "SELECT lcm( -5184226581::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:       "SELECT lcm( 8525267290::int8 ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:    "SELECT lcm( 2::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:    "SELECT lcm( -2::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:    "SELECT lcm( 5::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:    "SELECT lcm( 10::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:    "SELECT lcm( -10::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:    "SELECT lcm( 1000::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(852526729000)}},
				},
				{
					Query:    "SELECT lcm( 2105076::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8973167782882020)}},
				},
				{
					Query:    "SELECT lcm( 100000000::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(85252672900000000)}},
				},
				{
					Query:       "SELECT lcm( -5184226581::int8 ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 8525267290::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 0::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -1::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 1::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 2::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -2::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 5::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 10::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -10::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 1000::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 2105076::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 100000000::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -5184226581::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 8525267290::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 9223372036854775807::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 0::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT lcm( -1::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
				{
					Query:    "SELECT lcm( 1::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
				{
					Query:       "SELECT lcm( 2::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -2::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 5::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 10::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -10::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 1000::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 2105076::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 100000000::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -5184226581::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( 8525267290::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT lcm( -9223372036854775808::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT lcm( 9223372036854775807::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
			},
		},
	})
}
