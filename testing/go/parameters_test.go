// Copyright 2025 Dolthub, Inc.
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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestParameters(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "default_with_oids",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT default_with_oids;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SET default_with_oids = false;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "DateStyle",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SHOW DateStyle;",
					Expected: []sql.Row{{"ISO, MDY"}},
				},
				{
					Query:    "SELECT timestamp '2001/02/04 04:05:06.789';",
					Expected: []sql.Row{{"2001-02-04 04:05:06.789"}},
				},
				{
					Query:    "SET datestyle = 'german';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SHOW DateStyle;",
					Expected: []sql.Row{{"German, DMY"}},
				},
				{
					Skip:     true, // TODO: the test passes but pgx cannot parse the result
					Query:    "SELECT timestamp '2001/02/04 04:05:06.789';",
					Expected: []sql.Row{{"04.02.2001 04:05:06.789"}},
				},
				{
					Query:    "SET datestyle = 'YMD';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SHOW DateStyle;",
					Expected: []sql.Row{{"German, YMD"}},
				},
				{
					Query:    "SET datestyle = 'sQl';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SHOW DateStyle;",
					Expected: []sql.Row{{"SQL, YMD"}},
				},
				{
					Skip:     true, // TODO: the test passes but pgx cannot parse the result
					Query:    "SELECT timestamp '2001/02/04 04:05:06.789';",
					Expected: []sql.Row{{"02/04/2001 04:05:06.789"}},
				},
				{
					Query:    "SET datestyle = 'postgreS';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SHOW DateStyle;",
					Expected: []sql.Row{{"Postgres, YMD"}},
				},
				{
					Skip:     true, // TODO: the test passes but pgx cannot parse the result
					Query:    "SELECT timestamp '2001/02/04 04:05:06.789';",
					Expected: []sql.Row{{"Sun Feb 04 04:05:06.789 2001"}},
				},
				{
					Query:    "RESET datestyle;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SHOW DateStyle;",
					Expected: []sql.Row{{"ISO, MDY"}},
				},
				{
					Query:       "SET datestyle = 'unknown';",
					ExpectedErr: `invalid value for parameter "DateStyle": "unknown"`,
				},
			},
		},
	})
}
