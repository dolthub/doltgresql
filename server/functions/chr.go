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

import "fmt"

// chr represents the PostgreSQL function of the same name.
var chr = Function{
	Name:      "chr",
	Overloads: []interface{}{chr_string},
}

// chr_string is one of the overloads of chr.
func chr_string(num IntegerType) (StringType, error) {
	if num.IsNull {
		return StringType{IsNull: true}, nil
	}
	if num.Value == 0 {
		return StringType{}, fmt.Errorf("null character not permitted")
	} else if num.Value < 0 {
		return StringType{}, fmt.Errorf("character number must be positive")
	}
	return StringType{Value: string(rune(num.Value))}, nil
}
