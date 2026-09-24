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

func TestEmptyArrayType(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "empty array element types",
		Assertions: []ScriptTestAssertion{
			{
				Query:           "SELECT ARRAY[];",
				ExpectedErr:     "cannot determine type of empty array",
				ExpectedErrCode: "42P18",
			},
			{
				Query:           "SELECT pg_typeof(ARRAY[]);",
				ExpectedErr:     "cannot determine type of empty array",
				ExpectedErrCode: "42P18",
			},
			{
				Query:    "SELECT ARRAY[]::int[],pg_typeof(ARRAY[]::int[]),ARRAY[ARRAY[]]::int[];",
				Expected: []sql.Row{{"{}", "integer[]", "{}"}},
			},
			{Query: "SELECT ARRAY[ARRAY[]::int[]];", Expected: []sql.Row{{"{}"}}},
		},
	}})
}
