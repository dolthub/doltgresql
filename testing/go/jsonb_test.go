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
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cockroachdb/apd/v3"
	gmstypes "github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/types"
)

// TestLegacyJsonBRoundTrip verifies that JSONB values written in the old ExtendedEnc format
// (used before JsonAdaptiveEnc was introduced for jsonb columns) deserialize correctly.
// Each JSON value type is tested to ensure no internal JsonValue* types leak into the result.
func TestLegacyJsonBRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    any // value placed into JSONDocument.Val before serialization
		expected any // expected JSONDocument.Val after deserialization
	}{
		// ── Primitives ────────────────────────────────────────────────────────────────
		{
			name:     "null",
			input:    nil,
			expected: nil,
		},
		{
			name:     "bool_true",
			input:    true,
			expected: true,
		},
		{
			name:     "bool_false",
			input:    false,
			expected: false,
		},
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string_empty",
			input:    "",
			expected: "",
		},
		{
			name:     "string_unicode",
			input:    "こんにちは",
			expected: "こんにちは",
		},
		// Numbers: the primary compatibility concern — must come back as *apd.Decimal,
		// not the internal JsonValueNumber type.
		{
			name:     "number_zero",
			input:    json.Number("0"),
			expected: mustDecimal("0"),
		},
		{
			name:     "number_integer",
			input:    json.Number("42"),
			expected: mustDecimal("42"),
		},
		{
			name:     "number_negative_integer",
			input:    json.Number("-7"),
			expected: mustDecimal("-7"),
		},
		{
			name:     "number_float",
			input:    json.Number("3.14"),
			expected: mustDecimal("3.14"),
		},
		{
			name:     "number_negative_float",
			input:    json.Number("-2.718"),
			expected: mustDecimal("-2.718"),
		},
		{
			name:     "number_large_integer",
			input:    json.Number("99999999999999999999"),
			expected: mustDecimal("99999999999999999999"),
		},
		{
			name:     "number_high_precision",
			input:    json.Number("1.23456789012345678901"),
			expected: mustDecimal("1.23456789012345678901"),
		},
		// ── Empty containers ──────────────────────────────────────────────────────────
		{
			name:     "object_empty",
			input:    map[string]any{},
			expected: map[string]any{},
		},
		{
			name:     "array_empty",
			input:    []any{},
			expected: []any{},
		},
		// ── Objects: one test per value type ─────────────────────────────────────────
		{
			name:     "object_string_val",
			input:    map[string]any{"k": "v"},
			expected: map[string]any{"k": "v"},
		},
		{
			name:     "object_number_val",
			input:    map[string]any{"n": json.Number("99")},
			expected: map[string]any{"n": mustDecimal("99")},
		},
		{
			name:     "object_bool_val",
			input:    map[string]any{"b": true},
			expected: map[string]any{"b": true},
		},
		{
			name:     "object_null_val",
			input:    map[string]any{"n": nil},
			expected: map[string]any{"n": nil},
		},
		{
			name: "object_mixed_vals",
			input: map[string]any{
				"s": "hello", "n": json.Number("1.5"), "b": false, "z": nil,
			},
			expected: map[string]any{
				"s": "hello", "n": mustDecimal("1.5"), "b": false, "z": nil,
			},
		},
		// ── Arrays: one test per value type ──────────────────────────────────────────
		{
			name:     "array_strings",
			input:    []any{"a", "b", "c"},
			expected: []any{"a", "b", "c"},
		},
		{
			name:     "array_numbers",
			input:    []any{json.Number("1"), json.Number("2"), json.Number("3")},
			expected: []any{mustDecimal("1"), mustDecimal("2"), mustDecimal("3")},
		},
		{
			name:     "array_bools",
			input:    []any{true, false, true},
			expected: []any{true, false, true},
		},
		{
			name:     "array_nulls",
			input:    []any{nil, nil},
			expected: []any{nil, nil},
		},
		{
			name:     "array_mixed",
			input:    []any{json.Number("1"), "two", true, nil},
			expected: []any{mustDecimal("1"), "two", true, nil},
		},
		// ── Nested structures ─────────────────────────────────────────────────────────
		{
			name: "object_with_nested_array",
			input: map[string]any{
				"nums": []any{json.Number("1"), json.Number("2")},
				"name": "test",
			},
			expected: map[string]any{
				"nums": []any{mustDecimal("1"), mustDecimal("2")},
				"name": "test",
			},
		},
		{
			name: "array_of_objects",
			input: []any{
				map[string]any{"x": json.Number("10")},
				map[string]any{"x": json.Number("20")},
			},
			expected: []any{
				map[string]any{"x": mustDecimal("10")},
				map[string]any{"x": mustDecimal("20")},
			},
		},
		{
			name: "deeply_nested",
			input: map[string]any{
				"a": map[string]any{
					"b": []any{
						json.Number("1"),
						map[string]any{"c": true, "d": json.Number("-9.9")},
						[]any{"x", json.Number("0")},
					},
				},
			},
			expected: map[string]any{
				"a": map[string]any{
					"b": []any{
						mustDecimal("1"),
						map[string]any{"c": true, "d": mustDecimal("-9.9")},
						[]any{"x", mustDecimal("0")},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := gmstypes.JSONDocument{Val: tt.input}

			// Simulate writing with old Doltgres (ExtendedEnc path calls serializeTypeJsonB).
			serialized, err := types.JsonB.SerializeValue(context.Background(), doc)
			require.NoError(t, err)

			// Simulate reading with new Doltgres (still calls deserializeTypeJsonB for old rows).
			result, err := types.JsonB.DeserializeValue(context.Background(), serialized)
			require.NoError(t, err)
			require.NotNil(t, result)

			resultDoc, ok := result.(gmstypes.JSONDocument)
			require.True(t, ok)
			requireNoInternalJsonTypes(t, resultDoc.Val)
			requireJsonEqual(t, tt.expected, resultDoc.Val)
		})
	}
}

// requireNoInternalJsonTypes recurses through a deserialized JSON value and asserts that no
// internal JsonValue* types have leaked out. Every value must be a native Go type.
func requireNoInternalJsonTypes(t *testing.T, val any) {
	t.Helper()
	switch v := val.(type) {
	case map[string]any:
		for _, child := range v {
			requireNoInternalJsonTypes(t, child)
		}
	case []any:
		for _, item := range v {
			requireNoInternalJsonTypes(t, item)
		}
	case types.JsonValueNumber:
		t.Errorf("number leaked as internal JsonValueNumber; expected *apd.Decimal")
	case types.JsonValueString:
		t.Errorf("string leaked as internal JsonValueString; expected string")
	case types.JsonValueBoolean:
		t.Errorf("bool leaked as internal JsonValueBoolean; expected bool")
	case types.JsonValueNull:
		t.Errorf("null leaked as internal JsonValueNull; expected nil")
	case types.JsonValueObject:
		t.Errorf("object leaked as internal JsonValueObject; expected map[string]any")
	case types.JsonValueArray:
		t.Errorf("array leaked as internal JsonValueArray; expected []any")
	}
}

// requireJsonEqual recursively compares an expected JSON value tree against actual,
// using string comparison for *apd.Decimal to avoid representation sensitivity.
func requireJsonEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if expected == nil {
		require.Nil(t, actual, "expected nil, got %T: %v", actual, actual)
		return
	}
	switch e := expected.(type) {
	case *apd.Decimal:
		a, ok := actual.(*apd.Decimal)
		require.True(t, ok, "expected *apd.Decimal, got %T: %v", actual, actual)
		require.Equal(t, e.String(), a.String())
	case map[string]any:
		a, ok := actual.(map[string]any)
		require.True(t, ok, "expected map[string]any, got %T", actual)
		require.Equal(t, len(e), len(a), "object has %d keys, want %d", len(a), len(e))
		for k, ev := range e {
			av, exists := a[k]
			require.True(t, exists, "key %q missing from result", k)
			requireJsonEqual(t, ev, av)
		}
	case []any:
		a, ok := actual.([]any)
		require.True(t, ok, "expected []any, got %T", actual)
		require.Equal(t, len(e), len(a), "array has %d elements, want %d", len(a), len(e))
		for i := range e {
			requireJsonEqual(t, e[i], a[i])
		}
	default:
		require.Equal(t, expected, actual)
	}
}

// TestLegacyJsonBReserialization verifies that a value deserialized from the legacy
// ExtendedEnc format can be re-serialized without error. This exercises the path where
// serializeTypeJsonB receives a JSONDocument whose Val contains *apd.Decimal (the type
// that jsonValueToInterface now returns for JsonValueNumber after the bugfix), and
// ConvertToJsonDocument must handle it.
func TestLegacyJsonBReserialization(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"number_integer", json.Number("42")},
		{"number_negative", json.Number("-7")},
		{"number_float", json.Number("3.14")},
		{"number_large", json.Number("99999999999999999999")},
		{"object_with_number", map[string]any{"n": json.Number("99")}},
		{"array_with_numbers", []any{json.Number("1"), json.Number("2"), json.Number("3")}},
		{"nested_number", map[string]any{"a": map[string]any{"b": json.Number("1.5")}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := gmstypes.JSONDocument{Val: tt.input}

			// Step 1: serialize (simulates old Doltgres write using the legacy ExtendedEnc path)
			serialized, err := types.JsonB.SerializeValue(context.Background(), doc)
			require.NoError(t, err)

			// Step 2: deserialize (jsonValueToInterface now returns *apd.Decimal for numbers)
			deserialized, err := types.JsonB.DeserializeValue(context.Background(), serialized)
			require.NoError(t, err)
			require.NotNil(t, deserialized)

			// Step 3: re-serialize the deserialized value.
			_, err = types.JsonB.SerializeValue(context.Background(), deserialized)
			require.NoError(t, err, "re-serializing a deserialized legacy JSONB value must not fail")
		})
	}
}

// mustDecimal parses s into an *apd.Decimal, panicking on failure.
func mustDecimal(s string) *apd.Decimal {
	d := new(apd.Decimal)
	if err := d.Scan(s); err != nil {
		panic(fmt.Sprintf("mustDecimal(%q): %v", s, err))
	}
	return d
}
