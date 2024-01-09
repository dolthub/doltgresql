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

// width_bucket represents the PostgreSQL function of the same name.
var width_bucket = Function{
	Name:      "width_bucket",
	Overloads: []interface{}{width_bucket_float, width_bucket_numeric},
}

// width_bucket_float is one of the overloads of width_bucket.
func width_bucket_float(operand FloatType, low FloatType, high FloatType, count IntegerType) (IntegerType, error) {
	if operand.IsNull || low.IsNull || high.IsNull || count.IsNull {
		return IntegerType{IsNull: true}, nil
	}
	if count.Value <= 0 {
		return IntegerType{}, fmt.Errorf("count must be greater than zero")
	}
	bucket := (high.Value - low.Value) / float64(count.Value)
	result := int64(math.Ceil((operand.Value - low.Value) / bucket))
	if result < 0 {
		result = 0
	} else if result > count.Value+1 {
		result = count.Value + 1
	}
	return IntegerType{Value: result}, nil
}

// width_bucket_numeric is one of the overloads of width_bucket.
func width_bucket_numeric(operand NumericType, low NumericType, high NumericType, count IntegerType) (IntegerType, error) {
	//TODO: need to implement proper numeric support
	return width_bucket_float(FloatType(operand), FloatType(low), FloatType(high), count)
}
