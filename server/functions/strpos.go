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

// strpos represents the PostgreSQL function of the same name.
var strpos = Function{
	Name:      "strpos",
	Overloads: []interface{}{strpos_string},
}

// strpos_string is one of the overloads of strpos.
func strpos_string(str StringType, substring StringType) (IntegerType, error) {
	if str.IsNull || substring.IsNull {
		return IntegerType{IsNull: true}, nil
	}
	idx := strings.Index(str.Value, substring.Value)
	if idx == -1 {
		idx = 0
	}
	return IntegerType{Value: int64(idx)}, nil
}
