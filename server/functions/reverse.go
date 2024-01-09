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

// reverse represents the PostgreSQL function of the same name.
var reverse = Function{
	Name:      "reverse",
	Overloads: []interface{}{reverse_string},
}

// reverse_string is one of the overloads of reverse.
func reverse_string(text StringType) (StringType, error) {
	if text.IsNull {
		return StringType{IsNull: true}, nil
	}
	runes := []rune(text.Value)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return StringType{Value: string(runes)}, nil
}
