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

func Test_WidthBucket(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "width_bucket",
			Skip: true, // TODO: need to do another pass over these, they're generally correct but miss some edge cases
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT width_bucket( -10::float8 ,  -1::float8 ,  0::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  100000::float8 ,  0::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  10.87::float8 ,  -1::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  -10::float8 ,  -1::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  -10::float8 ,  2::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  -1::float8 ,  -2::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  2147483647.59024::float8 ,  -2::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  0::float8 ,  5.25::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  0::float8 ,  10.87::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2147483647.59024::float8 ,  -1::float8 ,  10.87::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  2::float8 ,  10.87::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  100::float8 ,  10.87::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  2147483647.59024::float8 ,  -10::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  100000::float8 ,  100::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  100::float8 ,  21050.48::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  -1::float8 ,  100000::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  5.25::float8 ,  100000::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  21050.48::float8 ,  100000::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  -2147483648::float8 ,  100000::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  10.87::float8 ,  -1184280::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -1184280::float8 ,  2525280.279::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  1::float8 ,  -2147483648::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  -10::float8 ,  -2147483648::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  100000::float8 ,  -2147483648::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  -1184280::float8 ,  -2147483648::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  2525280.279::float8 ,  -2147483648::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  2525280.279::float8 ,  -2147483648::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  -2::float8 ,  2147483647.59024::float8 ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  0::float8 ,  0::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  -1::float8 ,  0::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  -2::float8 ,  -1::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  100000::float8 ,  -1::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  -1::float8 ,  1::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  -1::float8 ,  2::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  -1::float8 ,  2::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  100000::float8 ,  2::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  2147483647.59024::float8 ,  2::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  2525280.279::float8 ,  -2::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  -2147483648::float8 ,  -2::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  -2147483648::float8 ,  5.25::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  1::float8 ,  10.87::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  5.25::float8 ,  -10::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2147483647.59024::float8 ,  5.25::float8 ,  -10::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  100::float8 ,  -10::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2147483647.59024::float8 ,  100::float8 ,  -10::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  -2147483648::float8 ,  -10::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  -1::float8 ,  100::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  0::float8 ,  21050.48::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  100000::float8 ,  21050.48::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  2525280.279::float8 ,  100000::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  100::float8 ,  -1184280::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  2147483647.59024::float8 ,  -1184280::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  -2147483648::float8 ,  2525280.279::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  -2147483648::float8 ,  2525280.279::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  -1::float8 ,  2147483647.59024::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  100000::float8 ,  2147483647.59024::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  2147483647.59024::float8 ,  2147483647.59024::float8 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  5.25::float8 ,  0::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  -2147483648::float8 ,  0::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -2::float8 ,  -1::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  -10::float8 ,  -1::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  1::float8 ,  1::float8 ,  1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  2::float8 ,  1::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  100000::float8 ,  1::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  2147483647.59024::float8 ,  1::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  2::float8 ,  2::float8 ,  1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  -2::float8 ,  -2::float8 ,  1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  2::float8 ,  5.25::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  2::float8 ,  5.25::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  -2::float8 ,  5.25::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  5.25::float8 ,  5.25::float8 ,  1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  5.25::float8 ,  10.87::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  2525280.279::float8 ,  10.87::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  -1::float8 ,  -10::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  -1::float8 ,  -10::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  1::float8 ,  -10::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  100::float8 ,  -10::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  1::float8 ,  100::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  5.25::float8 ,  100::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  0::float8 ,  100000::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  -2::float8 ,  100000::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  100::float8 ,  100000::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  100::float8 ,  100000::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  2525280.279::float8 ,  100000::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -2147483648::float8 ,  100000::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  100::float8 ,  -1184280::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  100000::float8 ,  -1184280::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -1184280::float8 ,  2525280.279::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  100000::float8 ,  -2147483648::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  2147483647.59024::float8 ,  -2147483648::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  10.87::float8 ,  2147483647.59024::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -2147483648::float8 ,  2147483647.59024::float8 ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  100::float8 ,  0::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -2147483648::float8 ,  0::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  0::float8 ,  -1::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  -1::float8 ,  -1::float8 ,  2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  1::float8 ,  1::float8 ,  2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  5.25::float8 ,  1::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  21050.48::float8 ,  1::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -1::float8 ,  2::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  2::float8 ,  2::float8 ,  2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  -2147483648::float8 ,  2::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  2147483647.59024::float8 ,  2::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  -2147483648::float8 ,  -2::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  100000::float8 ,  5.25::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -1184280::float8 ,  5.25::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  -10::float8 ,  -10::float8 ,  2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  100::float8 ,  -10::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  -10::float8 ,  100::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  100::float8 ,  100::float8 ,  2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -2147483648::float8 ,  100::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -10::float8 ,  21050.48::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  1::float8 ,  100000::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  1::float8 ,  2525280.279::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  10.87::float8 ,  2525280.279::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  2147483647.59024::float8 ,  -2147483648::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  100::float8 ,  2147483647.59024::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  100::float8 ,  2147483647.59024::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  100000::float8 ,  2147483647.59024::float8 ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  21050.48::float8 ,  0::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  -2147483648::float8 ,  0::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -1184280::float8 ,  -1::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  -2::float8 ,  1::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  100000::float8 ,  1::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  -1184280::float8 ,  2::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  2525280.279::float8 ,  2::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  10.87::float8 ,  -2::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  100::float8 ,  -2::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  -1184280::float8 ,  5.25::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  2::float8 ,  10.87::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -10::float8 ,  10.87::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  1::float8 ,  -10::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  2525280.279::float8 ,  -10::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -2147483648::float8 ,  -10::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  21050.48::float8 ,  100::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  10.87::float8 ,  21050.48::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  100::float8 ,  21050.48::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  -1::float8 ,  100000::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  5.25::float8 ,  100000::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  100::float8 ,  100000::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  -2::float8 ,  -1184280::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  5.25::float8 ,  -1184280::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  100::float8 ,  2525280.279::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  21050.48::float8 ,  2525280.279::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  -2147483648::float8 ,  2525280.279::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  1::float8 ,  -2147483648::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  2147483647.59024::float8 ,  -2147483648::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  5.25::float8 ,  2147483647.59024::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  10.87::float8 ,  2147483647.59024::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  -10::float8 ,  2147483647.59024::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  100000::float8 ,  2147483647.59024::float8 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  0::float8 ,  -1::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  10.87::float8 ,  -1::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  10.87::float8 ,  -1::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  -10::float8 ,  -1::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  100000::float8 ,  -1::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  2147483647.59024::float8 ,  -1::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -10::float8 ,  1::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  0::float8 ,  2::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  0::float8 ,  2::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  -1::float8 ,  2::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -2::float8 ,  2::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  100::float8 ,  2::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  2525280.279::float8 ,  2::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  21050.48::float8 ,  -2::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  5.25::float8 ,  5.25::float8 ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  21050.48::float8 ,  5.25::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  -10::float8 ,  10.87::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(4)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  21050.48::float8 ,  10.87::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  -1184280::float8 ,  10.87::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  -1184280::float8 ,  -10::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  -1::float8 ,  100::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  10.87::float8 ,  100::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  100::float8 ,  100::float8 ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  -1::float8 ,  21050.48::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  1::float8 ,  21050.48::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  2::float8 ,  21050.48::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -10::float8 ,  21050.48::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -2147483648::float8 ,  21050.48::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  10.87::float8 ,  100000::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  -10::float8 ,  100000::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  100000::float8 ,  100000::float8 ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  0::float8 ,  -1184280::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  -1::float8 ,  -1184280::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -1::float8 ,  -1184280::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  2525280.279::float8 ,  -1184280::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(4)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  5.25::float8 ,  2525280.279::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  -1184280::float8 ,  2525280.279::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  0::float8 ,  -2147483648::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -2::float8 ,  -2147483648::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  10.87::float8 ,  -2147483648::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  2::float8 ,  2147483647.59024::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  21050.48::float8 ,  2147483647.59024::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  2525280.279::float8 ,  2147483647.59024::float8 ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  0::float8 ,  0::float8 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  1::float8 ,  0::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  2::float8 ,  2::float8 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -2::float8 ,  -2::float8 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  2::float8 ,  5.25::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  5.25::float8 ,  5.25::float8 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  5.25::float8 ,  5.25::float8 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  10.87::float8 ,  5.25::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -10::float8 ,  5.25::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  0::float8 ,  10.87::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  10.87::float8 ,  10.87::float8 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -10::float8 ,  10.87::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  2::float8 ,  -10::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  2::float8 ,  -10::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  100::float8 ,  -10::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  2525280.279::float8 ,  -10::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -2147483648::float8 ,  -10::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  2147483647.59024::float8 ,  100::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  2::float8 ,  21050.48::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -2147483648::float8 ,  21050.48::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -1::float8 ,  100000::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  2525280.279::float8 ,  100000::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  1::float8 ,  -1184280::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  -2147483648::float8 ,  -1184280::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  2147483647.59024::float8 ,  -1184280::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  2147483647.59024::float8 ,  -1184280::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  2147483647.59024::float8 ,  -1184280::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  0::float8 ,  2525280.279::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  -10::float8 ,  -2147483648::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -1184280::float8 ,  -2147483648::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  0::float8 ,  2147483647.59024::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  1::float8 ,  2147483647.59024::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  5.25::float8 ,  2147483647.59024::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  -10::float8 ,  2147483647.59024::float8 ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  2147483647.59024::float8 ,  2147483647.59024::float8 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  2::float8 ,  0::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  2::float8 ,  -1::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2147483647.59024::float8 ,  -2::float8 ,  -1::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  21050.48::float8 ,  -1::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  1::float8 ,  2::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  10.87::float8 ,  2::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  10.87::float8 ,  2::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2147483647.59024::float8 ,  100000::float8 ,  2::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  2147483647.59024::float8 ,  2::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  -2147483648::float8 ,  -2::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  10.87::float8 ,  5.25::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  21050.48::float8 ,  5.25::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -1184280::float8 ,  5.25::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  -1184280::float8 ,  5.25::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  0::float8 ,  10.87::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  5.25::float8 ,  10.87::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  2147483647.59024::float8 ,  10.87::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  -1::float8 ,  -10::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  -1::float8 ,  -10::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  -2::float8 ,  -10::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  100::float8 ,  -10::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  100000::float8 ,  -10::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  -2147483648::float8 ,  -10::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  -2147483648::float8 ,  -10::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  -2147483648::float8 ,  -10::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  -2::float8 ,  100::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  -1::float8 ,  21050.48::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  -1184280::float8 ,  21050.48::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  100::float8 ,  100000::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  2147483647.59024::float8 ,  -1184280::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  0::float8 ,  2525280.279::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  1::float8 ,  2525280.279::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  -2::float8 ,  2525280.279::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  0::float8 ,  -2147483648::float8 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  100::float8 ,  0::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  2525280.279::float8 ,  0::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(100)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  0::float8 ,  -1::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  0::float8 ,  -1::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  21050.48::float8 ,  -1::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(100)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  -1184280::float8 ,  -1::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  5.25::float8 ,  1::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(77)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -1::float8 ,  2::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  21050.48::float8 ,  2::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(100)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  1::float8 ,  -2::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(67)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  2::float8 ,  -2::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  5.25::float8 ,  -2::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(45)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -1184280::float8 ,  -2::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  100000::float8 ,  5.25::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  -1184280::float8 ,  5.25::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  2::float8 ,  10.87::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  -10::float8 ,  10.87::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(53)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  -1184280::float8 ,  10.87::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  -10::float8 ,  -10::float8 ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  100000::float8 ,  -10::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  2::float8 ,  100::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  2147483647.59024::float8 ,  100::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  1::float8 ,  21050.48::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  2::float8 ,  21050.48::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  5.25::float8 ,  21050.48::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  -10::float8 ,  21050.48::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  21050.48::float8 ,  21050.48::float8 ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  -1::float8 ,  100000::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -2::float8 ,  100000::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  -10::float8 ,  -1184280::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  100000::float8 ,  -1184280::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(8)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  2147483647.59024::float8 ,  -1184280::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(100)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  10.87::float8 ,  2525280.279::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  -2::float8 ,  -2147483648::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  -2147483648::float8 ,  -2147483648::float8 ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  -10::float8 ,  2147483647.59024::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -2147483648::float8 ,  2147483647.59024::float8 ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(50)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  100::float8 ,  -1::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(20425)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -1::float8 ,  1::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -10::float8 ,  1::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  2147483647.59024::float8 ,  1::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -1::float8 ,  2::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  100::float8 ,  2::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(19145)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  2525280.279::float8 ,  2::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  0::float8 ,  -2::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  100::float8 ,  -2::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  21050.48::float8 ,  -2::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  21050.48::float8 ,  -2::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  -1::float8 ,  5.25::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  2::float8 ,  -10::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(5263)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  10.87::float8 ,  -10::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(8947)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  0::float8 ,  100::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  21050.48::float8 ,  100::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  2147483647.59024::float8 ,  100::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  -2::float8 ,  21050.48::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  10.87::float8 ,  21050.48::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  2147483647.59024::float8 ,  21050.48::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  2::float8 ,  100000::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  -10::float8 ,  100000::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -10::float8 ,  100000::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  0::float8 ,  -1184280::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  0::float8 ,  -1184280::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  1::float8 ,  -1184280::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  2147483647.59024::float8 ,  -1184280::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21038)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  0::float8 ,  2525280.279::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  0::float8 ,  2525280.279::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  2::float8 ,  -2147483648::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  -10::float8 ,  -2147483648::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  0::float8 ,  2147483647.59024::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  1::float8 ,  2147483647.59024::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  5.25::float8 ,  2147483647.59024::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  10.87::float8 ,  2147483647.59024::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  2525280.279::float8 ,  2147483647.59024::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  -2147483648::float8 ,  2147483647.59024::float8 ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(10525)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  1::float8 ,  0::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  5.25::float8 ,  0::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  10.87::float8 ,  -1::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  -1184280::float8 ,  -1::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  2147483647.59024::float8 ,  -1::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100000)}},
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  1::float8 ,  1::float8 ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  2::float8 ,  1::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  10.87::float8 ,  1::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  2147483647.59024::float8 ,  1::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -2::float8 ,  2::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  -10::float8 ,  2::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  -10::float8 ,  2::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  100000::float8 ,  2::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  -2147483648::float8 ,  2::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  10.87::float8 ,  -2::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(68920)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  -1184280::float8 ,  -2::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  -1184280::float8 ,  5.25::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  -1::float8 ,  10.87::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  -10::float8 ,  10.87::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(57499)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -10::float8 ,  10.87::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  2525280.279::float8 ,  10.87::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  100000::float8 ,  -10::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(99891)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  21050.48::float8 ,  100::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  2525280.279::float8 ,  100::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  2147483647.59024::float8 ,  100::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  5.25::float8 ,  21050.48::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  10.87::float8 ,  21050.48::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  21050.48::float8 ,  21050.48::float8 ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  21050.48::float8 ,  100000::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  2525280.279::float8 ,  100000::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  -1::float8 ,  -1184280::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  10.87::float8 ,  -1184280::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  10.87::float8 ,  -1184280::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  -2147483648::float8 ,  -1184280::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  21050.48::float8 ,  2525280.279::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  -2147483648::float8 ,  2525280.279::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -1::float8 ,  -2147483648::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  5.25::float8 ,  -2147483648::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  -1184280::float8 ,  -2147483648::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  -2147483648::float8 ,  -2147483648::float8 ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  5.25::float8 ,  2147483647.59024::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  100000::float8 ,  2147483647.59024::float8 ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  0::float8 ,  0::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  0::float8 ,  0::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  -10::float8 ,  0::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  2147483647.59024::float8 ,  0::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  -1::float8 ,  -1::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  100::float8 ,  -1::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  21050.48::float8 ,  -1::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  -1::float8 ,  1::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  1::float8 ,  1::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -1::float8 ,  2::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  2147483647.59024::float8 ,  2::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  100000::float8 ,  -2::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  -2147483648::float8 ,  -2::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  0::float8 ,  5.25::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  1::float8 ,  5.25::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2147483647.59024::float8 ,  100000::float8 ,  5.25::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  -1184280::float8 ,  10.87::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  -2::float8 ,  -10::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  10.87::float8 ,  -10::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  21050.48::float8 ,  -10::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -2147483648::float8 ,  -10::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  -1::float8 ,  100::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  5.25::float8 ,  21050.48::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  -2147483648::float8 ,  21050.48::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  1::float8 ,  -1184280::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  10.87::float8 ,  -1184280::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  -1184280::float8 ,  -1184280::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  2525280.279::float8 ,  -1184280::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  0::float8 ,  2525280.279::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  100::float8 ,  2525280.279::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  2525280.279::float8 ,  2525280.279::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5.25::float8 ,  100::float8 ,  -2147483648::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  100000::float8 ,  -2147483648::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2147483647.59024::float8 ,  2525280.279::float8 ,  -2147483648::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  100::float8 ,  2147483647.59024::float8 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  1::float8 ,  0::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  -2::float8 ,  0::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( -1::float8 ,  100::float8 ,  -1::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  100000::float8 ,  1::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525031)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  100000::float8 ,  1::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1993717)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  -2147483648::float8 ,  2::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  0::float8 ,  -2::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 21050.48::float8 ,  0::float8 ,  -2::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::float8 ,  -10::float8 ,  -2::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  2525280.279::float8 ,  -2::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  21050.48::float8 ,  5.25::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  0::float8 ,  10.87::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1219662)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  -1::float8 ,  10.87::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(212745)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  2147483647.59024::float8 ,  10.87::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525163)}},
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  2::float8 ,  -10::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(420881)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  100::float8 ,  -10::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2175185)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  -1184280::float8 ,  -10::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  21050.48::float8 ,  21050.48::float8 ,  2525280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  21050.48::float8 ,  21050.48::float8 ,  2525280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  0::float8 ,  100000::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(26)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  0::float8 ,  100000::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(133)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  -10::float8 ,  100000::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  -1::float8 ,  -1184280::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -10::float8 ,  -2147483648::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  -2147483648::float8 ,  -2147483648::float8 ,  2525280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  5.25::float8 ,  2147483647.59024::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  -2147483648::float8 ,  2147483647.59024::float8 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  -1184280::float8 ,  0::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  100000::float8 ,  -1::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  2147483647.59024::float8 ,  -1::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  2147483647.59024::float8 ,  -1::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  5.25::float8 ,  1::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  5.25::float8 ,  2::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  10.87::float8 ,  2::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  2147483647.59024::float8 ,  2::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10.87::float8 ,  -10::float8 ,  -2::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  2147483647.59024::float8 ,  -2::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  100::float8 ,  5.25::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::float8 ,  10.87::float8 ,  10.87::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  -10::float8 ,  10.87::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::float8 ,  -2::float8 ,  -10::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  -2147483648::float8 ,  100::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  -2::float8 ,  100000::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  -2::float8 ,  100000::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  100000::float8 ,  100000::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  2::float8 ,  -1184280::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  -10::float8 ,  -1184280::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  2525280.279::float8 ,  -1184280::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  2147483647.59024::float8 ,  -1184280::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  -2::float8 ,  2525280.279::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::float8 ,  100::float8 ,  2525280.279::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100::float8 ,  -1184280::float8 ,  2525280.279::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  100000::float8 ,  -2147483648::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::float8 ,  2147483647.59024::float8 ,  2147483647.59024::float8 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  -1::float8 ,  1::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( -1::float8 ,  2::float8 ,  2::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  2525280.279::float8 ,  2::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  2::float8 ,  -2::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2525280.279::float8 ,  -10::float8 ,  -2::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 0::float8 ,  0::float8 ,  5.25::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  100::float8 ,  5.25::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2020107836)}},
				},
				{
					Query:    "SELECT width_bucket( 100000::float8 ,  100::float8 ,  5.25::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  2525280.279::float8 ,  10.87::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( -2147483648::float8 ,  5.25::float8 ,  -10::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000::float8 ,  -10::float8 ,  -10::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2147483647.59024::float8 ,  100000::float8 ,  -10::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  2147483647.59024::float8 ,  -10::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2144958357)}},
				},
				{
					Query:       "SELECT width_bucket( 2147483647.59024::float8 ,  100::float8 ,  100::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  -1::float8 ,  21050.48::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(204023)}},
				},
				{
					Query:    "SELECT width_bucket( -2147483648::float8 ,  1::float8 ,  21050.48::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::float8 ,  5.25::float8 ,  21050.48::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  10.87::float8 ,  21050.48::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  21050.48::float8 ,  21050.48::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 21050.48::float8 ,  100000::float8 ,  21050.48::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  -1184280::float8 ,  21050.48::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2109980726)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -1184280::float8 ,  21050.48::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2110157110)}},
				},
				{
					Query:    "SELECT width_bucket( 1::float8 ,  10.87::float8 ,  100000::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100::float8 ,  -10::float8 ,  100000::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2361996)}},
				},
				{
					Query:       "SELECT width_bucket( -2::float8 ,  100000::float8 ,  100000::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1184280::float8 ,  2::float8 ,  -1184280::float8 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -10::float8 ,  2525280.279::float8 ,  -1184280::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(1461903614)}},
				},
				{
					Query:    "SELECT width_bucket( 10.87::float8 ,  10.87::float8 ,  -2147483648::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -1::float8 ,  2147483647.59024::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2525282)}},
				},
				{
					Query:    "SELECT width_bucket( 2525280.279::float8 ,  -10::float8 ,  2147483647.59024::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2525291)}},
				},
				{
					Query:    "SELECT width_bucket( 5.25::float8 ,  100000::float8 ,  2147483647.59024::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1184280::float8 ,  100000::float8 ,  2147483647.59024::float8 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  -10::numeric ,  0::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  1000::numeric ,  0::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  8525267290::numeric ,  -1::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  -10::numeric ,  1::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  -1::numeric ,  2::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  1000::numeric ,  2::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  1000::numeric ,  2::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  79223372036854775807::numeric ,  2::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  0::numeric ,  -2::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  -5184226581::numeric ,  -2::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  8525267290::numeric ,  10::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  1000::numeric ,  -10::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  -1::numeric ,  1000::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  5::numeric ,  1000::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  -2::numeric ,  2105076::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  -79223372036854775808::numeric ,  2105076::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  5::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ,  -5184226581::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  8525267290::numeric ,  -5184226581::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  10::numeric ,  8525267290::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  8525267290::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  -79223372036854775808::numeric ,  -79223372036854775808::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  79223372036854775807::numeric ,  -79223372036854775808::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  79223372036854775807::numeric ,  -79223372036854775808::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  0::numeric ,  79223372036854775807::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  2::numeric ,  79223372036854775807::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  2105076::numeric ,  79223372036854775807::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  -79223372036854775808::numeric ,  79223372036854775807::numeric ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  2105076::numeric ,  0::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  -5184226581::numeric ,  -1::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  0::numeric ,  1::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  -2::numeric ,  1::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  10::numeric ,  1::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  8525267290::numeric ,  1::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  8525267290::numeric ,  1::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -2::numeric ,  2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -10::numeric ,  2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  -79223372036854775808::numeric ,  2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  79223372036854775807::numeric ,  2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -1::numeric ,  -2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  1::numeric ,  -2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  2::numeric ,  -2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  1000::numeric ,  -2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  2105076::numeric ,  -2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  79223372036854775807::numeric ,  -2::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  -10::numeric ,  5::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  -5184226581::numeric ,  5::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  1000::numeric ,  -10::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  1::numeric ,  1000::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  1::numeric ,  1000::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  -10::numeric ,  1000::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -5184226581::numeric ,  1000::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  8525267290::numeric ,  1000::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  1::numeric ,  2105076::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  -2::numeric ,  2105076::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  -10::numeric ,  2105076::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -5184226581::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  -79223372036854775808::numeric ,  -5184226581::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  -2::numeric ,  8525267290::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  8525267290::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  -79223372036854775808::numeric ,  8525267290::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  -79223372036854775808::numeric ,  8525267290::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  10::numeric ,  -79223372036854775808::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  79223372036854775807::numeric ,  79223372036854775807::numeric ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  10::numeric ,  0::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  0::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  -10::numeric ,  -1::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  1::numeric ,  1::numeric ,  1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ,  2::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  -10::numeric ,  2::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  1000::numeric ,  2::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  79223372036854775807::numeric ,  -2::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -5184226581::numeric ,  5::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  -5184226581::numeric ,  5::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  -1::numeric ,  10::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  2::numeric ,  10::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  -2::numeric ,  -10::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -2::numeric ,  -10::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  10::numeric ,  1000::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -79223372036854775808::numeric ,  1000::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  1000::numeric ,  2105076::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  -79223372036854775808::numeric ,  2105076::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  2::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -10::numeric ,  -5184226581::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  -79223372036854775808::numeric ,  -5184226581::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  79223372036854775807::numeric ,  -5184226581::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  10::numeric ,  8525267290::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  -79223372036854775808::numeric ,  8525267290::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  -1::numeric ,  -79223372036854775808::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  -1::numeric ,  -79223372036854775808::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  -10::numeric ,  -79223372036854775808::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  -5184226581::numeric ,  -79223372036854775808::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  -1::numeric ,  79223372036854775807::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  -1::numeric ,  79223372036854775807::numeric ,  1::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -2::numeric ,  0::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  5::numeric ,  0::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  -10::numeric ,  0::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  1000::numeric ,  0::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  -5184226581::numeric ,  0::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  -2::numeric ,  -1::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  2105076::numeric ,  -1::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  -2::numeric ,  1::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  2105076::numeric ,  1::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  1000::numeric ,  2::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  1000::numeric ,  -2::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  79223372036854775807::numeric ,  -2::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  -10::numeric ,  5::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  0::numeric ,  10::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  10::numeric ,  10::numeric ,  2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  8525267290::numeric ,  10::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  2::numeric ,  1000::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -2::numeric ,  1000::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  2105076::numeric ,  2105076::numeric ,  2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -5184226581::numeric ,  2105076::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  -79223372036854775808::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  -1::numeric ,  -5184226581::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  8525267290::numeric ,  -5184226581::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  0::numeric ,  8525267290::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  -1::numeric ,  8525267290::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  1::numeric ,  8525267290::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -79223372036854775808::numeric ,  8525267290::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -79223372036854775808::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  1::numeric ,  79223372036854775807::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  1000::numeric ,  79223372036854775807::numeric ,  2::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  -10::numeric ,  0::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  1000::numeric ,  0::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  8525267290::numeric ,  0::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  -79223372036854775808::numeric ,  0::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ,  -1::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  2::numeric ,  1::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  -5184226581::numeric ,  1::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  -5184226581::numeric ,  1::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  2::numeric ,  2::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  5::numeric ,  2::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  -5184226581::numeric ,  2::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  0::numeric ,  -2::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  5::numeric ,  -2::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  -79223372036854775808::numeric ,  5::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  2::numeric ,  10::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  10::numeric ,  10::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  1000::numeric ,  10::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  -79223372036854775808::numeric ,  -10::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  0::numeric ,  2105076::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  1::numeric ,  2105076::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  2105076::numeric ,  2105076::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  0::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  10::numeric ,  -5184226581::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  2105076::numeric ,  -5184226581::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  -5184226581::numeric ,  -5184226581::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  2::numeric ,  8525267290::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  -2::numeric ,  8525267290::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  -79223372036854775808::numeric ,  8525267290::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  5::numeric ,  -79223372036854775808::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  1000::numeric ,  -79223372036854775808::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -79223372036854775808::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  2::numeric ,  79223372036854775807::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  -2::numeric ,  79223372036854775807::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  2::numeric ,  0::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  5::numeric ,  0::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  0::numeric ,  -1::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  2::numeric ,  2::numeric ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  1000::numeric ,  2::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  2105076::numeric ,  2::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  -2::numeric ,  -2::numeric ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  1::numeric ,  5::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  1000::numeric ,  5::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  -5184226581::numeric ,  5::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  -10::numeric ,  10::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(4)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -5184226581::numeric ,  10::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  -5184226581::numeric ,  10::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  1::numeric ,  -10::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  -2::numeric ,  -10::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  2105076::numeric ,  -10::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -10::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  79223372036854775807::numeric ,  -10::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  0::numeric ,  1000::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  2::numeric ,  1000::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  5::numeric ,  1000::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  2105076::numeric ,  1000::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  0::numeric ,  2105076::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  1::numeric ,  2105076::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  2::numeric ,  2105076::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  0::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  5::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(6)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  -10::numeric ,  -5184226581::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  2105076::numeric ,  -5184226581::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  10::numeric ,  8525267290::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  1000::numeric ,  8525267290::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  8525267290::numeric ,  8525267290::numeric ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ,  -79223372036854775808::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  1::numeric ,  79223372036854775807::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  10::numeric ,  79223372036854775807::numeric ,  5::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  10::numeric ,  0::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  -1::numeric ,  -1::numeric ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  2::numeric ,  -1::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  2::numeric ,  -1::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  -2::numeric ,  1::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  5::numeric ,  1::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  1000::numeric ,  1::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  -79223372036854775808::numeric ,  2::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  0::numeric ,  -2::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  5::numeric ,  -2::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  1000::numeric ,  -2::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -79223372036854775808::numeric ,  -2::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  2::numeric ,  5::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  1000::numeric ,  5::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  1000::numeric ,  5::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  -2::numeric ,  10::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  5::numeric ,  10::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  1::numeric ,  -10::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  1::numeric ,  -10::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(11)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  5::numeric ,  -10::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -5184226581::numeric ,  1000::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  5::numeric ,  2105076::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  10::numeric ,  2105076::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  2105076::numeric ,  2105076::numeric ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  2105076::numeric ,  2105076::numeric ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  5::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  -10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  -10::numeric ,  -5184226581::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  2105076::numeric ,  -5184226581::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  -5184226581::numeric ,  -5184226581::numeric ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  -1::numeric ,  8525267290::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  1::numeric ,  8525267290::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  0::numeric ,  -79223372036854775808::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  2105076::numeric ,  -79223372036854775808::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -79223372036854775808::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  5::numeric ,  79223372036854775807::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ,  10::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  0::numeric ,  0::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  1::numeric ,  0::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  2::numeric ,  0::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  8525267290::numeric ,  0::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  8525267290::numeric ,  0::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  10::numeric ,  -1::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  -10::numeric ,  -1::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  -10::numeric ,  -1::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  -1::numeric ,  1::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  -79223372036854775808::numeric ,  1::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  2::numeric ,  2::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  8525267290::numeric ,  2::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  0::numeric ,  -2::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  10::numeric ,  -2::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -2::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  79223372036854775807::numeric ,  -2::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  -2::numeric ,  5::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  -10::numeric ,  5::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  1::numeric ,  10::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  5::numeric ,  10::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  1000::numeric ,  10::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  -5184226581::numeric ,  10::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  2105076::numeric ,  -10::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  -5184226581::numeric ,  -10::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  79223372036854775807::numeric ,  -10::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  2105076::numeric ,  1000::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  -5184226581::numeric ,  1000::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  -2::numeric ,  2105076::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  79223372036854775807::numeric ,  2105076::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  2105076::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -2::numeric ,  -79223372036854775808::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  8525267290::numeric ,  -79223372036854775808::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  5::numeric ,  0::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -10::numeric ,  -1::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  2::numeric ,  1::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  2105076::numeric ,  1::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  1::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  8525267290::numeric ,  1::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  2::numeric ,  2::numeric ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -10::numeric ,  2::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(92)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -79223372036854775808::numeric ,  2::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  79223372036854775807::numeric ,  2::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  -2::numeric ,  -2::numeric ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  0::numeric ,  5::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  1000::numeric ,  10::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  -5184226581::numeric ,  10::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(100)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  8525267290::numeric ,  10::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  1::numeric ,  1000::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  10::numeric ,  1000::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  1000::numeric ,  1000::numeric ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  2::numeric ,  2105076::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  2105076::numeric ,  2105076::numeric ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  0::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  10::numeric ,  -5184226581::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  -10::numeric ,  -5184226581::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  2105076::numeric ,  -5184226581::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  79223372036854775807::numeric ,  8525267290::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(101)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  -1::numeric ,  -79223372036854775808::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  10::numeric ,  -79223372036854775808::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  1000::numeric ,  -79223372036854775808::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  79223372036854775807::numeric ,  -79223372036854775808::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(51)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  5::numeric ,  79223372036854775807::numeric ,  100::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  5::numeric ,  0::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  -1::numeric ,  -1::numeric ,  21050::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  10::numeric ,  -1::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  1000::numeric ,  -1::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21008)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  2::numeric ,  1::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  1::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(20607)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -2::numeric ,  2::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(15788)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  0::numeric ,  -2::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  1::numeric ,  -2::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  1::numeric ,  5::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  2105076::numeric ,  5::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  2::numeric ,  10::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  2::numeric ,  10::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  1000::numeric ,  10::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  0::numeric ,  -10::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  1::numeric ,  -10::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(5741)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  5::numeric ,  -10::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(7017)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  -79223372036854775808::numeric ,  -10::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -79223372036854775808::numeric ,  1000::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(21051)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  2105076::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  8525267290::numeric ,  -5184226581::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(13090)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  10::numeric ,  8525267290::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(247)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -2::numeric ,  -79223372036854775808::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  -2::numeric ,  79223372036854775807::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  1000::numeric ,  79223372036854775807::numeric ,  21050::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  10::numeric ,  0::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  2105076::numeric ,  1::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100000)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  5::numeric ,  2::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  -5184226581::numeric ,  2::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  -10::numeric ,  -2::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  -10::numeric ,  -2::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  1000::numeric ,  -2::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(99701)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  79223372036854775807::numeric ,  -2::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  -5184226581::numeric ,  5::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  -79223372036854775808::numeric ,  5::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  79223372036854775807::numeric ,  5::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  2::numeric ,  10::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  10::numeric ,  10::numeric ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  -79223372036854775808::numeric ,  10::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  79223372036854775807::numeric ,  10::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100000)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -2::numeric ,  -10::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  5::numeric ,  -10::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  0::numeric ,  1000::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  2::numeric ,  1000::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  -10::numeric ,  1000::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  -10::numeric ,  1000::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  -1::numeric ,  2105076::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  1::numeric ,  2105076::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2105076::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100001)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  5::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  -79223372036854775808::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(100000)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  10::numeric ,  -5184226581::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  2105076::numeric ,  -5184226581::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  2::numeric ,  8525267290::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  -79223372036854775808::numeric ,  8525267290::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  0::numeric ,  -79223372036854775808::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  2::numeric ,  -79223372036854775808::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  0::numeric ,  79223372036854775807::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  -1::numeric ,  79223372036854775807::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  -1::numeric ,  79223372036854775807::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -10::numeric ,  79223372036854775807::numeric ,  100000::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  1::numeric ,  0::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  2::numeric ,  0::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  2105076::numeric ,  0::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  2105076::numeric ,  0::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  79223372036854775807::numeric ,  0::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  -5184226581::numeric ,  -1::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  -5184226581::numeric ,  -1::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  -2::numeric ,  1::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  -5184226581::numeric ,  1::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  1::numeric ,  2::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  10::numeric ,  2::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  -79223372036854775808::numeric ,  2::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  -2::numeric ,  -2::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  10::numeric ,  -2::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  10::numeric ,  5::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  1000::numeric ,  10::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  1::numeric ,  -10::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  -10::numeric ,  -10::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  -10::numeric ,  1000::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  79223372036854775807::numeric ,  1000::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  -10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  -5184226581::numeric ,  -5184226581::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  8525267290::numeric ,  -5184226581::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  2::numeric ,  -79223372036854775808::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  8525267290::numeric ,  -79223372036854775808::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ,  0::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  1::numeric ,  0::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  2::numeric ,  0::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  0::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  8525267290::numeric ,  0::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525280)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  79223372036854775807::numeric ,  0::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  1000::numeric ,  -1::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  -2::numeric ,  1::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  1000::numeric ,  1::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  1000::numeric ,  1::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -79223372036854775808::numeric ,  2105076::numeric ,  1::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  1::numeric ,  2::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -79223372036854775808::numeric ,  2::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  79223372036854775807::numeric ,  -2::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  79223372036854775807::numeric ,  -2::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525280)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  1000::numeric ,  10::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  1000::numeric ,  -10::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -10::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  1::numeric ,  2105076::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2105076::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(2525281)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  2::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -1::numeric ,  -5184226581::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -5184226581::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(47790)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  0::numeric ,  -79223372036854775808::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 79223372036854775807::numeric ,  0::numeric ,  -79223372036854775808::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  -10::numeric ,  -79223372036854775808::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  2::numeric ,  79223372036854775807::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  -2::numeric ,  79223372036854775807::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( -2::numeric ,  5::numeric ,  79223372036854775807::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  -5184226581::numeric ,  79223372036854775807::numeric ,  2525280::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  -1::numeric ,  0::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  1000::numeric ,  0::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  1000::numeric ,  0::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  2105076::numeric ,  0::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  2105076::numeric ,  0::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  0::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -5184226581::numeric ,  0::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  79223372036854775807::numeric ,  0::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  5::numeric ,  -1::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  -10::numeric ,  -1::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  1000::numeric ,  -1::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  -79223372036854775808::numeric ,  -1::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  79223372036854775807::numeric ,  2::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  0::numeric ,  -2::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -2::numeric ,  5::numeric ,  5::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  10::numeric ,  5::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  10::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  -5184226581::numeric ,  10::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  79223372036854775807::numeric ,  10::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  0::numeric ,  1000::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  1::numeric ,  1000::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  10::numeric ,  1000::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  -10::numeric ,  1000::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  8525267290::numeric ,  1000::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  79223372036854775807::numeric ,  1000::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  1::numeric ,  2105076::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  2::numeric ,  2105076::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  10::numeric ,  2105076::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  -10::numeric ,  2105076::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  -10::numeric ,  2105076::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  2::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  79223372036854775807::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -1::numeric ,  79223372036854775807::numeric ,  -5184226581::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -5184226581::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  8525267290::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  0::numeric ,  -79223372036854775808::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  -10::numeric ,  -79223372036854775808::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  1000::numeric ,  -79223372036854775808::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  -1::numeric ,  79223372036854775807::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 0::numeric ,  10::numeric ,  79223372036854775807::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -10::numeric ,  79223372036854775807::numeric ,  79223372036854775807::numeric ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  -5184226581::numeric ,  0::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  0::numeric ,  -1::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  -1::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2147483411)}},
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -1::numeric ,  1::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 2::numeric ,  79223372036854775807::numeric ,  1::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( -79223372036854775808::numeric ,  79223372036854775807::numeric ,  1::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 79223372036854775807::numeric ,  -1::numeric ,  2::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 5::numeric ,  1::numeric ,  2::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -5184226581::numeric ,  2::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  79223372036854775807::numeric ,  5::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2147483647)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -2::numeric ,  10::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(1252698795)}},
				},
				{
					Query:       "SELECT width_bucket( 8525267290::numeric ,  -10::numeric ,  10::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 1::numeric ,  -10::numeric ,  -10::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT width_bucket( 10::numeric ,  -10::numeric ,  -10::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( -1::numeric ,  -2::numeric ,  1000::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2143198)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  5::numeric ,  2105076::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(5101)}},
				},
				{
					Query:       "SELECT width_bucket( 2105076::numeric ,  -5184226581::numeric ,  2105076::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  0::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(45206163)}},
				},
				{
					Query:    "SELECT width_bucket( -10::numeric ,  -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2105076::numeric ,  -1::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(45206184)}},
				},
				{
					Query:       "SELECT width_bucket( 1000::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT width_bucket( 8525267290::numeric ,  79223372036854775807::numeric ,  100000000.2345862323456346511423652312416532::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(2147483647)}},
				},
				{
					Query:    "SELECT width_bucket( 100000000.2345862323456346511423652312416532::numeric ,  1::numeric ,  -5184226581::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 0::numeric ,  8525267290::numeric ,  -5184226581::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(1335415608)}},
				},
				{
					Query:    "SELECT width_bucket( 1::numeric ,  -1::numeric ,  8525267290::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 1000::numeric ,  0::numeric ,  -79223372036854775808::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 5::numeric ,  -1::numeric ,  -79223372036854775808::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( 2::numeric ,  5::numeric ,  -79223372036854775808::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT width_bucket( 10::numeric ,  -10::numeric ,  -79223372036854775808::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT width_bucket( -5184226581::numeric ,  -5184226581::numeric ,  -79223372036854775808::numeric ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
			},
		},
	})
}
