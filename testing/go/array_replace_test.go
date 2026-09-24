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

func TestArrayReplace(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "array_replace", Assertions: []ScriptTestAssertion{
		{Query: "SELECT array_replace(ARRAY[[1,NULL],[1,4]],1,9), array_replace(ARRAY[[1,NULL],[1,4]],NULL,0);", Expected: []sql.Row{{"{{9,NULL},{9,4}}", "{{1,0},{1,4}}"}}},
		{Query: "SELECT array_replace(ARRAY['a','b'],'a',NULL), array_replace(NULL::int[],1,2), array_replace(ARRAY[]::int[],1,2);", Expected: []sql.Row{{"{NULL,b}", nil, "{}"}}},
	}}})
}
