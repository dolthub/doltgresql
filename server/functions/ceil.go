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

import (
	"math"
)

// ceil represents the PostgreSQL function of the same name.
var ceil = Function{
	Name:      "ceil",
	Overloads: []interface{}{ceil_float, ceil_numeric},
}

// ceiling represents the PostgreSQL function of the same name.
var ceiling = Function{
	Name:      "ceiling",
	Overloads: []interface{}{ceil_float, ceil_numeric},
}

// ceil_float is one of the overloads of ceil.
func ceil_float(num FloatType) (FloatType, error) {
	if num.IsNull {
		return FloatType{IsNull: true}, nil
	}
	return FloatType{Value: math.Ceil(num.Value)}, nil
}

// ceil_numeric is one of the overloads of ceil.
func ceil_numeric(num NumericType) (NumericType, error) {
	if num.IsNull {
		return NumericType{IsNull: true}, nil
	}
	return NumericType{Value: math.Ceil(num.Value)}, nil
}
