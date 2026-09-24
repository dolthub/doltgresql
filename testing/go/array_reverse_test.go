// Copyright 2026 Dolthub, Inc.
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

func TestArrayReverse(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "array_reverse",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT array_reverse(ARRAY[[2,4],[3,1],[1,9]]), array_reverse(ARRAY[1,NULL,2]);",
				Expected: []sql.Row{{"{{1,9},{3,1},{2,4}}", "{2,NULL,1}"}},
			},
			{
				Query:    "SELECT array_reverse(ARRAY[]::int[]),array_reverse(NULL::int[]);",
				Expected: []sql.Row{{"{}", nil}},
			},
		},
	}})
}
