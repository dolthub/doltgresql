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

package plpgsql

import (
	"testing"

	pg_query "github.com/dolthub/pg_query_go/v6"
	"github.com/stretchr/testify/require"
)

func TestSubstituteVariableReferencesWithComments(t *testing.T) {
	for _, test := range []struct {
		name       string
		expression string
		expected   string
		bindings   []string
	}{
		{
			name:       "line comment in subquery",
			expression: "EXISTS (SELECT 1 -- comment\n)",
			expected:   "EXISTS ( SELECT 1 ) ",
		},
		{
			name:       "variables around line comment",
			expression: "v + -- v is only a comment\nv",
			expected:   "$1 + $2 ",
			bindings:   []string{"v", "v"},
		},
		{
			name:       "trailing line comment",
			expression: "v -- comment",
			expected:   "$1 ",
			bindings:   []string{"v"},
		},
		{
			name:       "nested block comment",
			expression: "v /* outer /* v */ comment */ + v",
			expected:   "$1 + $2 ",
			bindings:   []string{"v", "v"},
		},
		{
			name:       "comment before function arguments",
			expression: "v -- comment\n(1)",
			expected:   "v ( 1 ) ",
		},
		{
			name:       "comments around record field dot",
			expression: "r /* comment */ . -- comment\nf",
			expected:   "$1 ",
			bindings:   []string{"r.f"},
		},
		{
			name:       "comment markers in literals and identifiers",
			expression: "'-- v' || $$/* v */$$ || \"-- v\"",
			expected:   "'-- v' || $$/* v */$$ || \"-- v\" ",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			stack := NewInterpreterStack(nil)
			stack.NewVariableWithValue("v", nil, nil)
			stack.NewRecord("r", Record{Name: "r", Fields: []string{"f"}}.fakeSchema(), nil)
			expression, bindings, err := substituteVariableReferences(test.expression, &stack)
			require.NoError(t, err)
			require.Equal(t, test.expected, expression)
			require.Equal(t, test.bindings, bindings)
			_, err = pg_query.Parse("SELECT " + expression + ";")
			require.NoError(t, err)
		})
	}
}
