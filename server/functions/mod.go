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

// mod represents the PostgreSQL function of the same name.
var mod = Function{
	Name:      "mod",
	Overloads: []interface{}{mod_int_int, mod_num_num},
}

// mod_int_int is one of the overloads of mod.
func mod_int_int(num1 IntegerType, num2 IntegerType) (IntegerType, error) {
	if num1.IsNull || num2.IsNull {
		return IntegerType{IsNull: true}, nil
	}
	return IntegerType{Value: num1.Value % num2.Value}, nil
}

// mod_num_num is one of the overloads of mod.
func mod_num_num(num1 NumericType, num2 NumericType) (NumericType, error) {
	if num1.IsNull || num2.IsNull {
		return NumericType{IsNull: true}, nil
	}
	return NumericType{Value: float64(int64(num1.Value) % int64(num2.Value))}, nil
}
