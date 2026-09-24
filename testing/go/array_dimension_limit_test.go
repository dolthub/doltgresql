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

func TestArrayDimensionLimit(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "array dimension limits",
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT array_ndims(ARRAY[[[[[[1]]]]]]);", Expected: []sql.Row{{6}}},
			{Query: "SELECT ARRAY[[[[[[[1]]]]]]];", ExpectedErr: "exceeds the maximum allowed (6)", ExpectedErrCode: "54000"},
			{Query: "SELECT '{{{{{{{1}}}}}}}'::int[];", ExpectedErr: "exceeds the maximum allowed (6)", ExpectedErrCode: "54000"},
		},
	}})
}
