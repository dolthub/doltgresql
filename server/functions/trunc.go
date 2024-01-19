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

// trunc represents the PostgreSQL function of the same name.
var trunc = Function{
	Name:      "trunc",
	Overloads: []interface{}{trunc_float, trunc_num, trunc_num_int},
}

// trunc_float is one of the overloads of trunc.
func trunc_float(num FloatType) (FloatType, error) {
	if num.IsNull {
		return FloatType{IsNull: true}, nil
	}
	return FloatType{Value: math.Trunc(num.Value)}, nil
}

// trunc_num is one of the overloads of trunc.
func trunc_num(num NumericType) (NumericType, error) {
	if num.IsNull {
		return NumericType{IsNull: true}, nil
	}
	return NumericType{Value: math.Trunc(num.Value)}, nil
}

// trunc_num_int is one of the overloads of trunc.
func trunc_num_int(num NumericType, places IntegerType) (NumericType, error) {
	if num.IsNull || places.IsNull {
		return NumericType{IsNull: true}, nil
	}
	power := math.Pow10(int(places.Value))
	return NumericType{Value: math.Trunc(num.Value*power) / power}, nil
}
