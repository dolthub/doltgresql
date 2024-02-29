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

package framework

import (
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ParameterType represents the type of a parameter.
type ParameterType uint8

const (
	ParameterType_Null    ParameterType = iota // The parameter is a NULL value, and is therefore typeless
	ParameterType_Integer                      // The parameter is an IntegerType type
	ParameterType_Float                        // The parameter is a FloatType type
	ParameterType_Numeric                      // The parameter is a NumericType type
	ParameterType_String                       // The parameter is a StringType type
)

// IsParameterType returns whether the given type matches one of the given parameter types. This is meant for a broad
// comparison (such as any of the integer types), which may not be applicable to every function.
func IsParameterType(t pgtypes.DoltgresType, parameterTypes ...ParameterType) bool {
	for _, parameterType := range parameterTypes {
		switch parameterType {
		case ParameterType_Null:
			if _, ok := t.(pgtypes.NullType); ok {
				return true
			}
		case ParameterType_Integer:
			switch t.(type) {
			case pgtypes.Int16Type, pgtypes.Int32Type, pgtypes.Int64Type:
				return true
			}
		case ParameterType_Float:
			switch t.(type) {
			case pgtypes.Float32Type, pgtypes.Float64Type:
				return true
			}
		case ParameterType_Numeric:
			if _, ok := t.(pgtypes.NumericType); ok {
				return true
			}
		case ParameterType_String:
			if _, ok := t.(pgtypes.VarCharType); ok {
				return true
			}
		}
	}
	return false
}
