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

func Test_Scale(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "scale",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT scale( 0::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( -1::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 1::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 2::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( -2::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 5::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 10::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( -10::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 1000::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 2105076::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 100000000.2345862323456346511423652312416532::numeric ) ;",
					Expected: []sql.Row{{int32(34)}},
				},
				{
					Query:    "SELECT scale( -5184226581::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 8525267290::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( -79223372036854775808::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT scale( 79223372036854775807::numeric ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
			},
		},
	})
}
