// Copyright 2023 Dolthub, Inc.
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

// round represents the PostgreSQL function of the same name.
var round = Function{
	Name:      "round",
	Overloads: []interface{}{round_num, round_float, round_num_dec},
}

// round_num is one of the overloads of round.
func round_num(num NumericType) (NumericType, error) {
	if num.IsNull {
		return NumericType{IsNull: true}, nil
	}
	return NumericType{Value: math.Round(num.Value)}, nil
}

// round_float is one of the overloads of round.
func round_float(num FloatType) (FloatType, error) {
	if num.IsNull {
		return FloatType{IsNull: true}, nil
	}
	return FloatType{Value: math.RoundToEven(num.Value)}, nil
}

// round_num_dec is one of the overloads of round.
func round_num_dec(num NumericType, decimalPlaces IntegerType) (NumericType, error) {
	if num.IsNull || decimalPlaces.IsNull {
		return NumericType{IsNull: true}, nil
	}
	ratio := math.Pow10(int(decimalPlaces.Value))
	return NumericType{Value: math.Round(num.Value*ratio) / ratio}, nil
}
