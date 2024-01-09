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

// btrim represents the PostgreSQL function of the same name.
var btrim = Function{
	Name:      "btrim",
	Overloads: []interface{}{btrim_string, btrim_string_string},
}

// btrim_string is one of the overloads of btrim.
func btrim_string(str StringType) (StringType, error) {
	return btrim_string_string(str, StringType{
		Value:        " ",
		IsNull:       false,
		OriginalType: ParameterType_String,
		Source:       Source_Constant,
	})
}

// btrim_string_string is one of the overloads of btrim.
func btrim_string_string(str StringType, characters StringType) (StringType, error) {
	if str.IsNull || characters.IsNull {
		return StringType{IsNull: true}, nil
	}
	result, err := ltrim_string_string(str, characters)
	if err != nil {
		return StringType{}, err
	}
	return rtrim_string_string(result, characters)
}
