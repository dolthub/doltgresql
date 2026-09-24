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
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

func TestArrayFill(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "array_fill", Assertions: []ScriptTestAssertion{
		{
			Query:    "SELECT array_fill(7,ARRAY[2,3]),array_fill(NULL::int,ARRAY[2,2]),array_fill('x'::varchar,ARRAY[2],ARRAY[1]);",
			Expected: []sql.Row{{"{{7,7,7},{7,7,7}}", "{{NULL,NULL},{NULL,NULL}}", "{x,x}"}},
		},
		{
			Query:    "SELECT array_fill(1,ARRAY[]::int[]),array_fill(1,ARRAY[2,0]);",
			Expected: []sql.Row{{"{}", "{}"}},
		},
		{
			Query:           "SELECT array_fill(1,ARRAY[1,1,1,1,1,1,1]);",
			ExpectedErr:     "array",
			ExpectedErrCode: "54000",
		},
		{
			Query:           "SELECT array_fill(1,ARRAY[NULL]::int[]);",
			ExpectedErr:     "cannot be null",
			ExpectedErrCode: "22004",
		},
		{
			Query:           "SELECT array_fill(1,NULL::int[]);",
			ExpectedErr:     "cannot be null",
			ExpectedErrCode: "22004",
		},
		{
			Query:           "SELECT array_fill(1,ARRAY[2],ARRAY[0]);",
			ExpectedErr:     "array",
			ExpectedErrCode: "0A000",
		},
		{
			Query:           "SELECT array_fill(1,ARRAY[[2,2]]);",
			ExpectedErr:     "array",
			ExpectedErrCode: "2202E",
		},
		{
			Query:           "SELECT array_fill(1,ARRAY[2147483647,2]);",
			ExpectedErr:     "array",
			ExpectedErrCode: "54000",
		},
	}}})
}
