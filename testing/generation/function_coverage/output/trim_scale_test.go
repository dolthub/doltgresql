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

func Test_TrimScale(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "trim_scale",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT trim_scale( 0::numeric ) ;",
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    "SELECT trim_scale( -1::numeric ) ;",
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    "SELECT trim_scale( 1::numeric ) ;",
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    "SELECT trim_scale( 2::numeric ) ;",
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    "SELECT trim_scale( -2::numeric ) ;",
					Expected: []sql.Row{{Numeric("-2")}},
				},
				{
					Query:    "SELECT trim_scale( 5::numeric ) ;",
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    "SELECT trim_scale( 10::numeric ) ;",
					Expected: []sql.Row{{Numeric("10")}},
				},
				{
					Query:    "SELECT trim_scale( -10::numeric ) ;",
					Expected: []sql.Row{{Numeric("-10")}},
				},
				{
					Query:    "SELECT trim_scale( 1000::numeric ) ;",
					Expected: []sql.Row{{Numeric("1000")}},
				},
				{
					Query:    "SELECT trim_scale( 2105076::numeric ) ;",
					Expected: []sql.Row{{Numeric("2105076")}},
				},
				{
					Query:    "SELECT trim_scale( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{Numeric("100000000.2345862323456346511423652312416532")}},
				},
				{
					Query:    "SELECT trim_scale( -5184226581::numeric ) ;",
					Expected: []sql.Row{{Numeric("-5184226581")}},
				},
				{
					Query:    "SELECT trim_scale( 8525267290::numeric ) ;",
					Expected: []sql.Row{{Numeric("8525267290")}},
				},
				{
					Query:    "SELECT trim_scale( -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{Numeric("-79223372036854775808")}},
				},
				{
					Query:    "SELECT trim_scale( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{Numeric("79223372036854775807")}},
				},
			},
		},
	})
}
