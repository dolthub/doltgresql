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

func Test_Cosh(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "cosh",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT cosh( 0::float8 ) ;",
					Expected: []sql.Row{{float64(1.000000)}},
				},
				{
					Query:    "SELECT cosh( -1::float8 ) ;",
					Expected: []sql.Row{{float64(1.543081)}},
				},
				{
					Query:    "SELECT cosh( 1::float8 ) ;",
					Expected: []sql.Row{{float64(1.543081)}},
				},
				{
					Query:    "SELECT cosh( 2::float8 ) ;",
					Expected: []sql.Row{{float64(3.762196)}},
				},
				{
					Query:    "SELECT cosh( -2::float8 ) ;",
					Expected: []sql.Row{{float64(3.762196)}},
				},
				{
					Query:    "SELECT cosh( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(95.285758)}},
				},
				{
					Query:    "SELECT cosh( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(26287.605145)}},
				},
				{
					Query:    "SELECT cosh( -10::float8 ) ;",
					Expected: []sql.Row{{float64(11013.232920)}},
				},
				{
					Query:    "SELECT cosh( 100::float8 ) ;",
					Expected: []sql.Row{{float64(13440585709080678047126700217981451777343488.000000)}},
				},
				{
					Query:    "SELECT cosh( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(+Inf)}},
				},
				{
					Query:    "SELECT cosh( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(+Inf)}},
				},
				{
					Query:    "SELECT cosh( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(+Inf)}},
				},
				{
					Query:    "SELECT cosh( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(+Inf)}},
				},
				{
					Query:    "SELECT cosh( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(+Inf)}},
				},
				{
					Query:    "SELECT cosh( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(+Inf)}},
				},
			},
		},
	})
}
