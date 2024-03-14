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

func Test_Ln(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "ln",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT ln( 0::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT ln( -1::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT ln( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.693147)}},
				},
				{
					Query:       "SELECT ln( -2::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(1.658228)}},
				},
				{
					Query:    "SELECT ln( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(2.386007)}},
				},
				{
					Query:       "SELECT ln( -10::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 100::float8 ) ;",
					Expected: []sql.Row{{float64(4.605170)}},
				},
				{
					Query:    "SELECT ln( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(9.954679)}},
				},
				{
					Query:    "SELECT ln( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(11.512925)}},
				},
				{
					Query:       "SELECT ln( -1184280::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(14.741863)}},
				},
				{
					Query:       "SELECT ln( -2147483648::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(21.487563)}},
				},
				{
					Query:       "SELECT ln( 0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT ln( -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT ln( 2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.6931471805599453")}},
				},
				{
					Query:       "SELECT ln( -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 5::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.6094379124341004")}},
				},
				{
					Query:    "SELECT ln( 10::numeric ) ;",
					Expected: []sql.Row{{Numeric("2.3025850929940457")}},
				},
				{
					Query:       "SELECT ln( -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("6.9077552789821371")}},
				},
				{
					Query:    "SELECT ln( 2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("14.559862128959931")}},
				},
				{
					Query:    "SELECT ln( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("18.4206807462982277928487431328957099")}},
				},
				{
					Query:       "SELECT ln( -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("22.866300213290165")}},
				},
				{
					Query:       "SELECT ln( -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT ln( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("45.818803030654766")}},
				},
			},
		},
	})
}
