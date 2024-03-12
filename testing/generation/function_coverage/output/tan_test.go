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

func Test_Tan(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "tan",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT tan( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT tan( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-1.557408)}},
				},
				{
					Query:    "SELECT tan( 1::float8 ) ;",
					Expected: []sql.Row{{float64(1.557408)}},
				},
				{
					Query:    "SELECT tan( 2::float8 ) ;",
					Expected: []sql.Row{{float64(-2.185040)}},
				},
				{
					Query:    "SELECT tan( -2::float8 ) ;",
					Expected: []sql.Row{{float64(2.185040)}},
				},
				{
					Query:    "SELECT tan( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-1.677326)}},
				},
				{
					Query:    "SELECT tan( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(7.921512)}},
				},
				{
					Query:    "SELECT tan( -10::float8 ) ;",
					Expected: []sql.Row{{float64(-0.648361)}},
				},
				{
					Query:    "SELECT tan( 100::float8 ) ;",
					Expected: []sql.Row{{float64(-0.587214)}},
				},
				{
					Query:    "SELECT tan( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(-4.114420)}},
				},
				{
					Query:    "SELECT tan( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(-0.035772)}},
				},
				{
					Query:    "SELECT tan( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-0.100902)}},
				},
				{
					Query:    "SELECT tan( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-1.595725)}},
				},
				{
					Query:    "SELECT tan( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(4.084289)}},
				},
				{
					Query:    "SELECT tan( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(5.838073)}},
				},
			},
		},
	})
}
