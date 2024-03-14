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

func Test_Mod(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "mod",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT mod( 0::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -1::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 1::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 2::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -2::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 5::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 10::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -10::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 1000::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 2105076::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 100000000::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -5184226581::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 8525267290::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 9223372036854775807::int8 ,  0::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 0::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  -1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  -1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  1::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  1::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  -2::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  -2::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  5::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  5::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(6)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  10::int8 ) ;",
					Expected: []sql.Row{{int64(7)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(6)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  -10::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  -10::int8 ) ;",
					Expected: []sql.Row{{int64(7)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(-10)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(76)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(-581)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(290)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  1000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  1000::int8 ) ;",
					Expected: []sql.Row{{int64(807)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(-10)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1061428)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(-1529469)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1814566)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  2105076::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  2105076::int8 ) ;",
					Expected: []sql.Row{{int64(1158031)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(-10)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(-84226581)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(25267290)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  100000000::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  100000000::int8 ) ;",
					Expected: []sql.Row{{int64(54775807)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(-10)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(3341040709)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  -5184226581::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  -5184226581::int8 ) ;",
					Expected: []sql.Row{{int64(1848274936)}},
				},
				{
					Query:    "SELECT mod( 0::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(-10)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(-5184226581)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  8525267290::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  8525267290::int8 ) ;",
					Expected: []sql.Row{{int64(3598291727)}},
				},
				{
					Query:       "SELECT mod( 0::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -1::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 1::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 2::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -2::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 5::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 10::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -10::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 1000::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 2105076::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 100000000::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -5184226581::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 8525267290::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 9223372036854775807::int8 ,  -9223372036854775808::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 0::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    "SELECT mod( -1::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    "SELECT mod( 1::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1)}},
				},
				{
					Query:    "SELECT mod( 2::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    "SELECT mod( -2::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(-2)}},
				},
				{
					Query:    "SELECT mod( 5::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    "SELECT mod( 10::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    "SELECT mod( -10::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(-10)}},
				},
				{
					Query:    "SELECT mod( 1000::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(1000)}},
				},
				{
					Query:    "SELECT mod( 2105076::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(2105076)}},
				},
				{
					Query:    "SELECT mod( 100000000::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(100000000)}},
				},
				{
					Query:    "SELECT mod( -5184226581::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(-5184226581)}},
				},
				{
					Query:    "SELECT mod( 8525267290::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(8525267290)}},
				},
				{
					Query:       "SELECT mod( -9223372036854775808::int8 ,  9223372036854775807::int8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 9223372036854775807::int8 ,  9223372036854775807::int8 ) ;",
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       "SELECT mod( 0::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -1::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 1::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 2::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -2::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 5::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 10::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -10::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 1000::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 2105076::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -5184226581::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 8525267290::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( -79223372036854775808::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT mod( 79223372036854775807::numeric ,  0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT mod( 0::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("-3")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  5::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("6")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-8")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  10::numeric ) ;",
					Expected: []sql.Row{{Numeric("7")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("6")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-8")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("7")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("76")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("-581")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("290")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("-808")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("807")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("1061428.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1529469")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("1814566")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1359852")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("1359851")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("2.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("5.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("10.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076.0000000000000000000000000000000000")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("-84226569.0361021503726327917393732066756868")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("25267270.0601702506210546528989553444594780")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("-7652645.5670207595773734568890609641592680")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("7652644.5670207595773734568890609641592680")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000000.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("3341040709")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1640094201")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("1640094200")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000000.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5184226581")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("-461460068")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("461460067")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000000.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5184226581")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("8525267290")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775807")}},
				},
				{
					Query:    "SELECT mod( 0::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT mod( -1::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 1::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT mod( 2::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT mod( -2::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT mod( 5::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT mod( 10::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT mod( -10::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT mod( 1000::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT mod( 2105076::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT mod( 100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000000.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT mod( -5184226581::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5184226581")}},
				},
				{
					Query:    "SELECT mod( 8525267290::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("8525267290")}},
				},
				{
					Query:    "SELECT mod( -79223372036854775808::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT mod( 79223372036854775807::numeric ,  79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
			},
		},
	})
}
