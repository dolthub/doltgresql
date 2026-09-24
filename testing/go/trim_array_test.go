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

func TestTrimArray(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "trim_array",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT trim_array(ARRAY[[1,2],[3,4],[5,6]],1), trim_array(ARRAY[1,2],2), trim_array(NULL::int[],1);",
				Expected: []sql.Row{{"{{1,2},{3,4}}", "{}", nil}},
			},
			{
				Query:    "SELECT trim_array(ARRAY[]::int[],0);",
				Expected: []sql.Row{{"{}"}},
			},
			{
				Query:           "SELECT trim_array(ARRAY[[1,2],[3,4]],3);",
				ExpectedErr:     "number of elements",
				ExpectedErrCode: "2202E",
			},
			{
				Query:           "SELECT trim_array(ARRAY[1],-1);",
				ExpectedErr:     "number of elements",
				ExpectedErrCode: "2202E",
			},
		},
	}})
}
