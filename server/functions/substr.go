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

// substr represents the PostgreSQL function of the same name.
var substr = Function{
	Name:      "substr",
	Overloads: []interface{}{substr_string_int, substr_string_int_int},
}

// substr_string_int is one of the overloads of substr.
func substr_string_int(str StringType, start IntegerType) (StringType, error) {
	if str.IsNull || start.IsNull {
		return StringType{IsNull: true}, nil
	}
	runes := []rune(str.Value)
	return StringType{Value: string(runes[start.Value:])}, nil
}

// substr_string_int_int is one of the overloads of substr.
func substr_string_int_int(str StringType, start IntegerType, count IntegerType) (StringType, error) {
	if str.IsNull || start.IsNull || count.IsNull {
		return StringType{IsNull: true}, nil
	}
	runes := []rune(str.Value)
	return StringType{Value: string(runes[start.Value : start.Value+count.Value])}, nil
}
