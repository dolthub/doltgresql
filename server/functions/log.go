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
	"fmt"
	"math"
)

// log represents the PostgreSQL function of the same name.
var log = Function{
	Name:      "log",
	Overloads: []interface{}{log_float, log_numeric, log_numeric_numeric},
}

// log_float is one of the overloads of log.
func log_float(num FloatType) (FloatType, error) {
	if num.IsNull {
		return FloatType{IsNull: true}, nil
	}
	if num.Value == 0 {
		return FloatType{}, fmt.Errorf("cannot take logarithm of zero")
	} else if num.Value < 0 {
		return FloatType{}, fmt.Errorf("cannot take logarithm of a negative number")
	}
	return FloatType{Value: math.Log10(num.Value)}, nil
}

// log_numeric is one of the overloads of log.
func log_numeric(num NumericType) (NumericType, error) {
	if num.IsNull {
		return NumericType{IsNull: true}, nil
	}
	if num.Value == 0 {
		return NumericType{}, fmt.Errorf("cannot take logarithm of zero")
	} else if num.Value < 0 {
		return NumericType{}, fmt.Errorf("cannot take logarithm of a negative number")
	}
	return NumericType{Value: math.Log10(num.Value)}, nil
}

// log_numeric_numeric is one of the overloads of log.
func log_numeric_numeric(base NumericType, num NumericType) (NumericType, error) {
	if base.IsNull || num.IsNull {
		return NumericType{IsNull: true}, nil
	}
	if base.Value == 0 || num.Value == 0 {
		return NumericType{}, fmt.Errorf("cannot take logarithm of zero")
	} else if base.Value < 0 || num.Value < 0 {
		return NumericType{}, fmt.Errorf("cannot take logarithm of a negative number")
	}
	return NumericType{Value: math.Log(num.Value) / math.Log(base.Value)}, nil
}
