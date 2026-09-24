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

func TestArrayLower(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "array_lower",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT array_lower(ARRAY[[1,2],[3,4]],1),array_lower(ARRAY[[1,2],[3,4]],2),array_lower(ARRAY[1],2);",
				Expected: []sql.Row{{1, 1, nil}},
			},
			{
				Query:    "SELECT array_lower(ARRAY[]::int[],1),array_lower(NULL::int[],1),array_lower(ARRAY[1],0),array_lower('1 2'::int2vector,1);",
				Expected: []sql.Row{{nil, nil, nil, 0}},
			},
		},
	}})
}
