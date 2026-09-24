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

func TestCardinality(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "cardinality",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT cardinality(ARRAY[[1,NULL],[3,4]]), cardinality(ARRAY[]::int[]), cardinality(NULL::int[]);",
				Expected: []sql.Row{{4, 0, nil}},
			},
			{Query: "SELECT cardinality(ARRAY[[[1,2]],[[3,4]]]);", Expected: []sql.Row{{4}}},
		},
	}})
}
