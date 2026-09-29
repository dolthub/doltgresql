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

// TestArraySlices checks clipped array slices from literals, table columns, and subqueries.
func TestArraySlices(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name:        "array slices",
		SetUpScript: []string{"CREATE TABLE slice_inputs (id int PRIMARY KEY,a int[],lo int,hi int);", "INSERT INTO slice_inputs VALUES (1,ARRAY[1,2,3],2,9),(2,ARRAY[]::int[],1,2),(3,NULL,1,2);"},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT (ARRAY[[1,2,3],[4,5,6]])[2:2][2:3], (ARRAY[[1,2,3],[4,5,6]])[:][2:2];",
				Expected: []sql.Row{{"{{5,6}}", "{{2},{5}}"}},
			},
			{
				Query:    "SELECT (ARRAY[[1,2,3],[4,5,6]])[2][2:3], (ARRAY[[1,2,3],[4,5,6]])[2:9][2:9];",
				Expected: []sql.Row{{"{{2,3},{5,6}}", "{{5,6}}"}},
			},
			{
				Query:    "SELECT (ARRAY[[1,2],[3,4]])[9:10][:], (ARRAY[1,2,3])[NULL:2], (NULL::int[])[1:2];",
				Expected: []sql.Row{{"{}", nil, nil}},
			},
			{
				Query:    "SELECT (ARRAY[[[1,2]],[[3,4]]])[2:2][:][:], array_ndims((ARRAY[[1,2],[3,4]])[2:2]);",
				Expected: []sql.Row{{"{{{3,4}}}", 2}},
			},
			{
				Query:    "SELECT (ARRAY[1,2,3])[-2:2], (ARRAY[1,2,3])[3:1], (ARRAY[]::int[])[:];",
				Expected: []sql.Row{{"{1,2}", "{}", "{}"}},
			},
			{
				Query:    "SELECT id,a[lo:hi],a[lo:NULL] FROM slice_inputs ORDER BY id;",
				Expected: []sql.Row{{1, "{2,3}", nil}, {2, "{}", nil}, {3, nil, nil}},
			},
			{
				Query:    "SELECT ((SELECT a FROM slice_inputs WHERE id=1))[(SELECT 2):(SELECT 9)];",
				Expected: []sql.Row{{"{2,3}"}},
			},
			{
				Query:    "SELECT (ARRAY[1,2])[1:2][1:1],(ARRAY[[1,2],[3,4]])[1:1];",
				Expected: []sql.Row{{"{}", "{{1,2}}"}},
			},
		},
	}})
}
