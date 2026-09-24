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

func TestArraySort(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "array_sort", Assertions: []ScriptTestAssertion{
		{
			Query:    "SELECT array_sort(ARRAY[[2,4],[3,1],[1,9]]), array_sort(ARRAY[3,NULL,1,2]);",
			Expected: []sql.Row{{"{{1,9},{2,4},{3,1}}", "{1,2,3,NULL}"}},
		},
		{
			Query:    "SELECT array_sort(ARRAY[3,NULL,1],true), array_sort(ARRAY[3,NULL,1],true,false), array_sort(ARRAY[3,NULL,1],false,true);",
			Expected: []sql.Row{{"{NULL,3,1}", "{3,1,NULL}", "{NULL,1,3}"}},
		},
		{
			Query:    "SELECT array_sort(ARRAY[[1,NULL],[1,2],[NULL,1]]), array_sort(ARRAY[]::int[]), array_sort(NULL::int[]);",
			Expected: []sql.Row{{"{{1,2},{1,NULL},{NULL,1}}", "{}", nil}},
		},
		{
			Query:    "SELECT array_sort(ARRAY['z','a','m']),array_sort(ARRAY[1],NULL);",
			Expected: []sql.Row{{"{a,m,z}", nil}},
		},
	}}})
}
