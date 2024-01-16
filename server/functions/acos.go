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

import "math"

// acos represents the PostgreSQL function of the same name.
var acos = Function{
	Name:      "acos",
	Overloads: []interface{}{acos_float},
}

// acos_float is one of the overloads of acos.
func acos_float(num FloatType) (FloatType, error) {
	if num.IsNull {
		return FloatType{IsNull: true}, nil
	}
	return FloatType{Value: math.Acos(num.Value)}, nil
}
