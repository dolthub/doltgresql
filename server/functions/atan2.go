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

import "math"

// atan2 represents the PostgreSQL function of the same name.
var atan2 = Function{
	Name:      "atan2",
	Overloads: []interface{}{atan2_float},
}

// atan2_float is one of the overloads of atan2.
func atan2_float(y FloatType, x FloatType) (FloatType, error) {
	if y.IsNull || x.IsNull {
		return FloatType{IsNull: true}, nil
	}
	return FloatType{Value: math.Atan2(y.Value, x.Value)}, nil
}
