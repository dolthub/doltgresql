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

// factorial represents the PostgreSQL function of the same name.
var factorial = Function{
	Name:      "factorial",
	Overloads: []interface{}{factorial_int},
}

// factorial_int is one of the overloads of factorial.
func factorial_int(num IntegerType) (IntegerType, error) {
	if num.IsNull {
		return IntegerType{IsNull: true}, nil
	}
	if num.Value < 0 {
		return IntegerType{}, fmt.Errorf("factorial of a negative number is undefined")
	}
	total := int64(1)
	for i := int64(2); i <= num.Value; i++ {
		total *= i
	}
	return IntegerType{Value: total}, nil
}
