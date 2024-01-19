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

import "strings"

// repeat represents the PostgreSQL function of the same name.
var repeat = Function{
	Name:      "repeat",
	Overloads: []interface{}{repeat_string},
}

// repeat_string is one of the overloads of repeat.
func repeat_string(str StringType, num IntegerType) (StringType, error) {
	if str.IsNull || num.IsNull {
		return StringType{IsNull: true}, nil
	}
	return StringType{Value: strings.Repeat(str.Value, int(num.Value))}, nil
}
