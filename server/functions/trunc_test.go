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

package functions

import (
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"
)

func TestTruncNumeric(t *testing.T) {
	// Expected values were checked against PostgreSQL 16
	tests := []struct {
		input    string
		places   int32
		expected string
	}{
		{"1.567", 2, "1.56"},
		{"-2.25", 1, "-2.2"},
		{"123.456", -1, "120"},
		{"-987.654", -2, "-900"},
		{"99999.999", -5, "0"},
		{"5", -3, "0"},
		{"1.5", 0, "1"},
		{"-1.5", 0, "-1"},
		{"-0.5", 0, "0"},
		{"123.456", 0, "123"},
		{"1.5", 3, "1.500"},
		{"-0.05", 1, "0.0"},
		{"0.000", 1, "0.0"},
		{"1E+20", 2, "100000000000000000000.00"},
		{"NaN", 0, "NaN"},
		{"NaN", 2, "NaN"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			dec, _, err := apd.NewFromString(test.input)
			require.NoError(t, err)
			res, err := truncNumeric(dec, test.places)
			require.NoError(t, err)
			require.Equal(t, test.expected, res.Text('f'), "trunc(%s, %d)", test.input, test.places)
		})
	}
}
