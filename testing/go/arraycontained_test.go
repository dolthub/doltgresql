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

func TestArrayContained(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "arraycontained", Assertions: []ScriptTestAssertion{
		{
			Query:    "SELECT ARRAY[4,1] <@ ARRAY[[1,2],[3,4]],ARRAY[5] <@ ARRAY[1,2],ARRAY[]::int[] <@ ARRAY[1];",
			Expected: []sql.Row{{"t", "f", "t"}},
		},
	}}})
}
