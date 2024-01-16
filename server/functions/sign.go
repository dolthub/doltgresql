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

// sign represents the PostgreSQL function of the same name.
var sign = Function{
	Name:      "sign",
	Overloads: []interface{}{sign_float, sign_numeric},
}

// sign_float is one of the overloads of sign.
func sign_float(num FloatType) (FloatType, error) {
	if num.IsNull {
		return FloatType{IsNull: true}, nil
	}
	if num.Value < 0 {
		return FloatType{Value: -1}, nil
	} else if num.Value > 0 {
		return FloatType{Value: 1}, nil
	} else {
		return FloatType{Value: 0}, nil
	}
}

// sign_numeric is one of the overloads of sign.
func sign_numeric(num NumericType) (NumericType, error) {
	if num.IsNull {
		return NumericType{IsNull: true}, nil
	}
	if num.Value < 0 {
		return NumericType{Value: -1}, nil
	} else if num.Value > 0 {
		return NumericType{Value: 1}, nil
	} else {
		return NumericType{Value: 0}, nil
	}
}
