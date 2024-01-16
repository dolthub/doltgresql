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

// power represents the PostgreSQL function of the same name.
var power = Function{
	Name:      "power",
	Overloads: []interface{}{power_float_float, power_num_num},
}

// power_float_float is one of the overloads of power.
func power_float_float(num1 FloatType, num2 FloatType) (FloatType, error) {
	if num1.IsNull || num2.IsNull {
		return FloatType{IsNull: true}, nil
	}
	return FloatType{Value: math.Pow(num1.Value, num2.Value)}, nil
}

// power_num_num is one of the overloads of power.
func power_num_num(num1 NumericType, num2 NumericType) (NumericType, error) {
	if num1.IsNull || num2.IsNull {
		return NumericType{IsNull: true}, nil
	}
	return NumericType{Value: math.Pow(num1.Value, num2.Value)}, nil
}
