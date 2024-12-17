// Copyright 2024 Dolthub, Inc.
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

package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIoInputSections(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		err      string
	}{
		{
			input:    "character varying",
			expected: []string{"character varying"},
		},
		{
			input:    "integer[]",
			expected: []string{"integer[]"},
		},
		{
			input:    `"foo"`,
			expected: []string{"foo"},
		},
		{
			input: `"foo`,
			err:   "invalid name syntax",
		},
		{
			input:    `foo"`,
			expected: []string{`foo"`},
		},
		{
			input: `""foo`,
			err:   "invalid name syntax",
		},
		{
			input:    `"char"`,
			expected: []string{`"char"`},
		},
		{
			input:    `varchar(10)`,
			expected: []string{"varchar(10)"},
		},
		{
			input:    `"char"[]`,
			expected: []string{"\"char\"[]"},
		},
		{
			input:    "  testing",
			expected: []string{"testing"},
		},
		{
			input:    "testschema.testing",
			expected: []string{"testschema", ".", "testing"},
		},
		{
			input:    `pg_catalog."char"`,
			expected: []string{"pg_catalog", ".", `"char"`},
		},
		{
			input:    `"pg_catalog"."char"`,
			expected: []string{"pg_catalog", ".", `"char"`},
		},
		{
			input:    `"char".test`,
			expected: []string{"char", ".", "test"},
		},
		{
			input:    `"myschema".foo`,
			expected: []string{"myschema", ".", "foo"},
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual, err := ioInputSections(test.input)
			if test.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.err)
			} else {
				require.NoError(t, err)
				assert.Len(t, actual, len(test.expected))
				for i := range actual {
					if actual[i] != test.expected[i] {
						assert.Equal(t, test.expected[i], actual[i])
					}
				}
			}
		})
	}
}
