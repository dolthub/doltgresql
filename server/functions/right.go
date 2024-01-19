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

// right represents the PostgreSQL function of the same name.
var right = Function{
	Name:      "right",
	Overloads: []interface{}{right_string_int},
}

// right_string is one of the overloads of right.
func right_string_int(string StringType, n IntegerType) (StringType, error) {
	if string.IsNull || n.IsNull {
		return StringType{IsNull: true}, nil
	}
	if n.Value >= 0 {
		return StringType{Value: string.Value[len(string.Value)-int(n.Value):]}, nil
	} else {
		return StringType{Value: string.Value[int(-n.Value):]}, nil
	}
}
