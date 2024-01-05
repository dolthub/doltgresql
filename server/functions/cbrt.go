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
	"math"
)

// cbrt represents the PostgreSQL function of the same name.
var cbrt = Function{
	Name:      "cbrt",
	Overloads: []interface{}{cbrt_float},
}

// cbrt_float is one of the overloads of cbrt.
func cbrt_float(num FloatType) (FloatType, error) {
	if num.IsNull {
		return FloatType{IsNull: true}, nil
	}
	if num.OriginalType == ParameterType_String {
		return FloatType{}, fmt.Errorf("function cbrt(%s) does not exist", ParameterType_String.String())
	}
	return FloatType{Value: math.Cbrt(num.Value)}, nil
}
