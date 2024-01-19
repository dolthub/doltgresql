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

// rtrim represents the PostgreSQL function of the same name.
var rtrim = Function{
	Name:      "rtrim",
	Overloads: []interface{}{rtrim_string, rtrim_string_string},
}

// rtrim_string is one of the overloads of rtrim.
func rtrim_string(str StringType) (StringType, error) {
	return rtrim_string_string(str, StringType{
		Value:        " ",
		IsNull:       false,
		OriginalType: ParameterType_String,
		Source:       Source_Constant,
	})
}

// rtrim_string_string is one of the overloads of rtrim.
func rtrim_string_string(str StringType, characters StringType) (StringType, error) {
	if str.IsNull || characters.IsNull {
		return StringType{IsNull: true}, nil
	}
	runes := []rune(str.Value)
	trimChars := make(map[rune]struct{})
	for _, c := range characters.Value {
		trimChars[c] = struct{}{}
	}
	trimIdx := len(runes)
	for ; trimIdx > 0; trimIdx-- {
		if _, ok := trimChars[runes[trimIdx-1]]; !ok {
			break
		}
	}
	return StringType{Value: string(runes[:trimIdx])}, nil
}
