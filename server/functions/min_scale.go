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
	"strconv"
	"strings"
)

// min_scale represents the PostgreSQL function of the same name.
var min_scale = Function{
	Name:      "min_scale",
	Overloads: []interface{}{min_scale_numeric},
}

// min_scale_numeric is one of the overloads of min_scale.
func min_scale_numeric(num NumericType) (NumericType, error) {
	if num.IsNull {
		return NumericType{IsNull: true}, nil
	}
	str := strconv.FormatFloat(num.Value, 'f', -1, 64)
	if idx := strings.Index(str, "."); idx != -1 {
		str = str[idx+1:]
		i := len(str) - 1
		for ; i >= 0; i-- {
			if str[i] != '0' {
				break
			}
		}
		return NumericType{Value: float64(i + 1)}, nil
	}
	return NumericType{Value: 0}, nil
}
