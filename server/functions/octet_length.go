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

// octet_length represents the PostgreSQL function of the same name.
var octet_length = Function{
	Name:      "octet_length",
	Overloads: []interface{}{octet_length_string},
}

// octet_length_string is one of the overloads of octet_length.
func octet_length_string(text StringType) (IntegerType, error) {
	if text.IsNull {
		return IntegerType{IsNull: true}, nil
	}
	return IntegerType{Value: int64(len(text.Value))}, nil
}
