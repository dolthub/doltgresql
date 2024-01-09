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
	"github.com/dolthub/doltgresql/utils"
)

// abs represents the PostgreSQL function of the same name.
var abs = Function{
	Name:      "abs",
	Overloads: []interface{}{abs_int, abs_float, abs_numeric},
}

// abs_int is one of the overloads of abs.
func abs_int(num IntegerType) (IntegerType, error) {
	if num.IsNull {
		return IntegerType{IsNull: true}, nil
	}
	return IntegerType{Value: utils.Abs(num.Value)}, nil
}

// abs_float is one of the overloads of abs.
func abs_float(num FloatType) (FloatType, error) {
	if num.IsNull {
		return FloatType{IsNull: true}, nil
	}
	return FloatType{Value: utils.Abs(num.Value)}, nil
}

// abs_numeric is one of the overloads of abs.
func abs_numeric(num NumericType) (NumericType, error) {
	if num.IsNull {
		return NumericType{IsNull: true}, nil
	}
	return NumericType{Value: utils.Abs(num.Value)}, nil
}
