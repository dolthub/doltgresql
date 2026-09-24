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

func TestUnnestMultidimensionalArguments(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "multi-array unnest flattens each input", Assertions: []ScriptTestAssertion{
		{Query: "SELECT unnest('1 2'::int2vector);", Expected: []sql.Row{{1}, {2}}},

		{Query: "SELECT * FROM unnest(ARRAY[[1,2],[3,4]],ARRAY['a','b']::varchar[]) AS u(n,label);", Expected: []sql.Row{{1, "a"}, {2, "b"}, {3, nil}, {4, nil}}},
		{Query: "SELECT * FROM unnest(NULL::int[],ARRAY[[1,2],[3,4]]) AS u(a,b);", Expected: []sql.Row{{nil, 1}, {nil, 2}, {nil, 3}, {nil, 4}}},
	}}})
}
