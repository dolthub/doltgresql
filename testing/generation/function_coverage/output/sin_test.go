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

func Test_Sin(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "sin",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT sin( 0::float8 ) ;",
					Expected: []sql.Row{{float64(0.000000)}},
				},
				{
					Query:    "SELECT sin( -1::float8 ) ;",
					Expected: []sql.Row{{float64(-0.841471)}},
				},
				{
					Query:    "SELECT sin( 1::float8 ) ;",
					Expected: []sql.Row{{float64(0.841471)}},
				},
				{
					Query:    "SELECT sin( 2::float8 ) ;",
					Expected: []sql.Row{{float64(0.909297)}},
				},
				{
					Query:    "SELECT sin( -2::float8 ) ;",
					Expected: []sql.Row{{float64(-0.909297)}},
				},
				{
					Query:    "SELECT sin( 5.25::float8 ) ;",
					Expected: []sql.Row{{float64(-0.858934)}},
				},
				{
					Query:    "SELECT sin( 10.87::float8 ) ;",
					Expected: []sql.Row{{float64(-0.992126)}},
				},
				{
					Query:    "SELECT sin( -10::float8 ) ;",
					Expected: []sql.Row{{float64(0.544021)}},
				},
				{
					Query:    "SELECT sin( 100::float8 ) ;",
					Expected: []sql.Row{{float64(-0.506366)}},
				},
				{
					Query:    "SELECT sin( 21050.48::float8 ) ;",
					Expected: []sql.Row{{float64(0.971711)}},
				},
				{
					Query:    "SELECT sin( 100000::float8 ) ;",
					Expected: []sql.Row{{float64(0.035749)}},
				},
				{
					Query:    "SELECT sin( -1184280::float8 ) ;",
					Expected: []sql.Row{{float64(-0.100392)}},
				},
				{
					Query:    "SELECT sin( 2525280.279::float8 ) ;",
					Expected: []sql.Row{{float64(-0.847360)}},
				},
				{
					Query:    "SELECT sin( -2147483648::float8 ) ;",
					Expected: []sql.Row{{float64(0.971310)}},
				},
				{
					Query:    "SELECT sin( 2147483647.59024::float8 ) ;",
					Expected: []sql.Row{{float64(-0.985645)}},
				},
			},
		},
	})
}
