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

// lpad represents the PostgreSQL function of the same name.
var lpad = Function{
	Name:      "lpad",
	Overloads: []interface{}{lpad_string_int, lpad_string_int_string},
}

// lpad_string_int is one of the overloads of lpad.
func lpad_string_int(str StringType, length IntegerType) (StringType, error) {
	return lpad_string_int_string(str, length, StringType{
		Value:        " ",
		IsNull:       false,
		OriginalType: ParameterType_String,
		Source:       Source_Constant,
	})
}

// lpad_string_int_string is one of the overloads of lpad.
func lpad_string_int_string(str StringType, length IntegerType, fill StringType) (StringType, error) {
	if str.IsNull || length.IsNull || fill.IsNull {
		return StringType{IsNull: true}, nil
	}
	if length.Value <= 0 {
		return StringType{Value: ""}, nil
	}
	runes := []rune(str.Value)
	fillTarget := length.Value - int64(len(runes))
	fillRunes := []rune(fill.Value)
	var result []rune
	if fillTarget > 0 {
		for int64(len(result)) < fillTarget {
			result = append(result, fillRunes...)
		}
		result = result[:fillTarget]
	}
	result = append(result, runes...)
	return StringType{Value: string(result[:length.Value])}, nil
}
