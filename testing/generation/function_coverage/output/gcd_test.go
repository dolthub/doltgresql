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

func Test_Gcd(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "gcd",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT gcd( 0::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  0::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(3)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(3)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(5184226581)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:       "SELECT gcd( 0::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( -1::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 1::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 2::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( -2::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 5::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 10::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( -10::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 1000::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 2105076::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 100000000::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( -5184226581::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 8525267290::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT gcd( 9223372036854775807::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 0::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
				{
					Query:    "SELECT gcd( -1::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -2::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 5::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 10::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -10::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 1000::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 2105076::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 100000000::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( -5184226581::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT gcd( 8525267290::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:       "SELECT gcd( -9223372036854775808::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT gcd( 9223372036854775807::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(9223372036854775807)}},
				},
			},
		},
	})
}
