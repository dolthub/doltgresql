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

func Test_SetConfig(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "set_config",
			Assertions: []ScriptTestAssertion{
				{
					// non-namespaced, non-system settings result in an error: "unrecognized configuration parameter"
					Query:       "SELECT set_config('doesnotexist', '42', false);",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT set_config('', 'bar', false);",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT set_config(NULL, 'bar', false);",
					ExpectedErr: true,
				},
				{
					// Set a system config setting
					Query:    "SELECT set_config('search_path', '123', false);",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT current_setting('search_path');",
					Expected: []sql.Row{{"123"}},
				},
				{
					// Set a user config setting in a custom namespace
					Query:    "SELECT set_config('mynamespace.var', 'bar', false);",
					Expected: []sql.Row{{"bar"}},
				},
				{
					Query:    "SELECT current_setting('mynamespace.var');",
					Expected: []sql.Row{{"bar"}},
				},
				{
					// Only text values are supported
					Query:       "SELECT set_config('myvars.boo', 3.14159, false);",
					ExpectedErr: true,
				},
				{
					// All settings must be text
					Query:    "SELECT set_config('myvars.boo', 3.14159::text, false);",
					Expected: []sql.Row{{"3.14159"}},
				},
				{
					Query:    "SELECT current_setting('myvars.boo');",
					Expected: []sql.Row{{"3.14159"}},
				},
				{
					// A NULL value is turned into the empty string
					Query:    "SELECT set_config('myvars.nullval', NULL, false);",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT current_setting('myvars.nullval');",
					Expected: []sql.Row{{""}},
				},
			},
		},
	})
}
