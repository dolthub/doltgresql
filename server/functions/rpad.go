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

// rpad represents the PostgreSQL function of the same name.
var rpad = Function{
	Name:      "rpad",
	Overloads: []interface{}{rpad_string_int, rpad_string_int_string},
}

// rpad_string_int is one of the overloads of rpad.
func rpad_string_int(str StringType, length IntegerType) (StringType, error) {
	return rpad_string_int_string(str, length, StringType{
		Value:        " ",
		IsNull:       false,
		OriginalType: ParameterType_String,
		Source:       Source_Constant,
	})
}

// rpad_string_int_string is one of the overloads of rpad.
func rpad_string_int_string(str StringType, length IntegerType, fill StringType) (StringType, error) {
	if str.IsNull || length.IsNull || fill.IsNull {
		return StringType{IsNull: true}, nil
	}
	if length.Value <= 0 {
		return StringType{Value: ""}, nil
	}
	runes := []rune(str.Value)
	fillRunes := []rune(fill.Value)
	for int64(len(runes)) < length.Value {
		runes = append(runes, fillRunes...)
	}
	return StringType{Value: string(runes[:length.Value])}, nil
}
