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

func Test_ToChar(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "to_char",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'YYYY-MM-DD HH24:MI:SS.MS');`,
					Expected: []sql.Row{
						{"2021-09-15 21:43:56.123"},
					},
				},
			},
		},
	})
}
