// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSerializationConsistency checks that all types serialization and deserialization.
func TestSerializationConsistency(t *testing.T) {
	for _, typ := range typesFromOID {
		t.Run(typ.String(), func(t *testing.T) {
			serializedType := typ.Serialize()
			dt, err := DeserializeType(serializedType)
			require.NoError(t, err)
			require.Equal(t, typ, dt.(DoltgresType))
		})
	}
}

// TestJsonValueType operates as a line of defense to prevent accidental changes to JSON type values. If this test
// fails, then a JsonValueType was changed that should not have been changed.
func TestJsonValueType(t *testing.T) {
	types := []struct {
		JsonValueType
		Value byte
		Name  string
	}{
		{JsonValueType_Object, 0, "Object"},
		{JsonValueType_Array, 1, "Array"},
		{JsonValueType_String, 2, "String"},
		{JsonValueType_Number, 3, "Number"},
		{JsonValueType_Boolean, 4, "Boolean"},
		{JsonValueType_Null, 5, "Null"},
	}
	allValues := make(map[byte]string)
	for _, typ := range types {
		if byte(typ.JsonValueType) != typ.Value {
			t.Logf("JSON value type `%s` has been changed from its permanent value of `%d` to `%d`",
				typ.Name, typ.Value, byte(typ.JsonValueType))
			t.Fail()
		} else if existingName, ok := allValues[typ.Value]; ok {
			t.Logf("JSON value type `%s` has the same value as `%s`: `%d`",
				typ.Name, existingName, typ.Value)
			t.Fail()
		} else {
			allValues[typ.Value] = typ.Name
		}
	}
}
