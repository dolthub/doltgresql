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

func Test_Log10(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "log10",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT log10( 0::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log10( -1::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT log10( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.301030)}},
				},
				{
					Query:       "SELECT log10( -2::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(0.720159)}},
				},
				{
					Query:    "SELECT log10( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(1.036230)}},
				},
				{
					Query:       "SELECT log10( -10::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 100::float8 ) ;",
					Expected: []sql.Row{{float64(2.000000)}},
				},
				{
					Query:    "SELECT log10( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(4.323262)}},
				},
				{
					Query:    "SELECT log10( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(5.000000)}},
				},
				{
					Query:       "SELECT log10( -1184280::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(6.402310)}},
				},
				{
					Query:       "SELECT log10( -2147483648::float8 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(9.331930)}},
				},
				{
					Query:       "SELECT log10( 0::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT log10( -1::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 1::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT log10( 2::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.3010299956639812")}},
				},
				{
					Query:       "SELECT log10( -2::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 5::numeric ) ;",
					Expected: []sql.Row{{Numeric("0.6989700043360188")}},
				},
				{
					Query:    "SELECT log10( 10::numeric ) ;",
					Expected: []sql.Row{{Numeric("1.0000000000000000")}},
				},
				{
					Query:       "SELECT log10( -10::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("3.0000000000000000")}},
				},
				{
					Query:    "SELECT log10( 2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("6.323267779879430")}},
				},
				{
					Query:    "SELECT log10( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("8.0000000010187950611868560912011483")}},
				},
				{
					Query:       "SELECT log10( -5184226581::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("9.930708004175069")}},
				},
				{
					Query:       "SELECT log10( -79223372036854775808::numeric ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT log10( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("19.898853323625356")}},
				},
			},
		},
	})
}
