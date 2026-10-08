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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	gmstypes "github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/functions"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// testAnyWrapper simulates *val.ExtendedValueWrapper — a sql.AnyWrapper that wraps a
// sql.JSONWrapper rather than a plain string. This is what Dolt returns when reading a
// large JSONB value that was stored out-of-band with the old ExtendedAdaptiveEnc encoding
// (used before JsonAdaptiveEnc was enabled for json/jsonb columns). UnwrapAny() on such a
// wrapper calls the child handler's DeserializeValue, which for JSONB returns a
// gmstypes.JSONDocument, not a string.
type testAnyWrapper struct {
	inner interface{}
}

func (w *testAnyWrapper) UnwrapAny(_ context.Context) (interface{}, error) { return w.inner, nil }
func (w *testAnyWrapper) IsExactLength() bool                              { return true }
func (w *testAnyWrapper) MaxByteLength() int64                             { return 1000 }
func (w *testAnyWrapper) Compare(_ context.Context, _ interface{}) (int, bool, error) {
	return 0, false, nil
}
func (w *testAnyWrapper) Hash() interface{} { return nil }

// TestJsonbOutCallableWithAnyWrapper verifies that jsonb_out_callable handles a
// sql.AnyWrapper value. Large JSONB values in databases created before JsonAdaptiveEnc
// was enabled (using ExtendedAdaptiveEnc instead) are returned from storage as a
// *val.ExtendedValueWrapper, which implements sql.AnyWrapper. Its UnwrapAny() yields a
// gmstypes.JSONDocument — so jsonb_out_callable must unwrap and then handle the document,
// not panic or return an "unexpected type" error.
func TestJsonbOutCallableWithAnyWrapper(t *testing.T) {
	ctx := sql.NewEmptyContext()

	tests := []struct {
		name    string
		jsonStr string
	}{
		{"object", `{"key": "value", "num": 42}`},
		{"array", `[1, 2, 3]`},
		{"nested", `{"a": {"b": [true, null]}}`},
		{"string_scalar", `"hello"`},
		{"number_scalar", `99`},
		{"bool_scalar", `false`},
		{"null_scalar", `null`},
		// large_object exercises the out-of-band storage path: a document large
		// enough that the old ExtendedAdaptiveEnc encoding would store it off-page.
		{"large_object", makeLargeJSONObject(100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonDoc := gmstypes.MustJSON(tt.jsonStr)

			result, err := functions.JsonbOutCallable(ctx, [2]*pgtypes.DoltgresType{}, &testAnyWrapper{inner: jsonDoc})
			require.NoError(t, err, "jsonb_out_callable must handle sql.AnyWrapper wrapping a JSONDocument")
			require.NotEmpty(t, result)
		})
	}
}

// TestJsonOutCallableWithAnyWrapper is the json_out equivalent of TestJsonbOutCallableWithAnyWrapper.
// json_out_callable has the same AnyWrapper gap as jsonb_out_callable.
func TestJsonOutCallableWithAnyWrapper(t *testing.T) {
	ctx := sql.NewEmptyContext()

	tests := []struct {
		name    string
		jsonStr string
	}{
		{"object", `{"key": "value", "num": 42}`},
		{"array", `[1, 2, 3]`},
		{"nested", `{"a": {"b": [true, null]}}`},
		{"string_scalar", `"hello"`},
		{"number_scalar", `99`},
		{"bool_scalar", `false`},
		{"null_scalar", `null`},
		{"large_object", makeLargeJSONObject(100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonDoc := gmstypes.MustJSON(tt.jsonStr)

			result, err := functions.JsonOutCallable(ctx, [2]*pgtypes.DoltgresType{}, &testAnyWrapper{inner: jsonDoc})
			require.NoError(t, err, "json_out_callable must handle sql.AnyWrapper wrapping a JSONDocument")
			require.NotEmpty(t, result)
		})
	}
}

// TestJsonbSerializeWithAnyWrapper verifies that types.JsonB.SerializeValue handles a
// sql.AnyWrapper that wraps a JSONDocument containing *apd.Decimal values. This
// represents the combined Bug 1 + Bug 2 production failure path:
//
//  1. A large legacy JSONB value is read from Dolt → *val.ExtendedValueWrapper (AnyWrapper).
//  2. The wrapper is passed to serializeTypeJsonB (e.g. during replication or export).
//  3. sql.UnwrapAny extracts a gmstypes.JSONDocument.
//  4. ToInterface() returns the .Val, which after the jsonValueToInterface bugfix
//     contains *apd.Decimal for numeric fields.
//  5. ConvertToJsonDocument(*apd.Decimal) must handle that type — previously it did not.
func TestJsonbSerializeWithAnyWrapper(t *testing.T) {
	ctx := sql.NewEmptyContext()

	tests := []struct {
		name  string
		inner gmstypes.JSONDocument
	}{
		{
			"scalar_decimal",
			gmstypes.JSONDocument{Val: mustDecimal("3.14")},
		},
		{
			"object_with_decimal",
			gmstypes.JSONDocument{Val: map[string]any{
				"n": mustDecimal("42"),
				"s": "hello",
			}},
		},
		{
			"array_with_decimals",
			gmstypes.JSONDocument{Val: []any{mustDecimal("1"), mustDecimal("2"), mustDecimal("3")}},
		},
		{
			"nested_decimals",
			gmstypes.JSONDocument{Val: map[string]any{
				"a": map[string]any{"b": mustDecimal("-9.9")},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &testAnyWrapper{inner: tt.inner}
			_, err := pgtypes.JsonB.SerializeValue(ctx, wrapper)
			require.NoError(t, err,
				"SerializeValue must handle an AnyWrapper wrapping a JSONDocument with *apd.Decimal")
		})
	}
}
