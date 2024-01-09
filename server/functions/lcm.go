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

import (
	"fmt"

	"github.com/dolthub/doltgresql/utils"
)

// lcm represents the PostgreSQL function of the same name.
var lcm = Function{
	Name:      "lcm",
	Overloads: []interface{}{lcm_int_int},
}

// lcm_int_int is one of the overloads of lcm.
func lcm_int_int(num1 IntegerType, num2 IntegerType) (IntegerType, error) {
	if num1.IsNull || num2.IsNull {
		return IntegerType{IsNull: true}, nil
	}
	if num1.OriginalType == ParameterType_String || num2.OriginalType == ParameterType_String {
		return IntegerType{}, fmt.Errorf("function lcm(%s, %s) does not exist",
			num1.OriginalType.String(), num2.OriginalType.String())
	}
	gcdResult, err := gcd_int_int(num1, num2)
	if err != nil {
		return IntegerType{}, err
	}
	if gcdResult.Value == 0 {
		return IntegerType{Value: 0}, nil
	}
	return IntegerType{Value: utils.Abs((num1.Value * num2.Value) / gcdResult.Value)}, nil
}
