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

// atan2d represents the PostgreSQL function of the same name.
var atan2d = Function{
	Name:      "atan2d",
	Overloads: []interface{}{atan2d_float},
}

// atan2d_float is one of the overloads of atan2d.
func atan2d_float(y FloatType, x FloatType) (FloatType, error) {
	if y.IsNull || x.IsNull {
		return FloatType{IsNull: true}, nil
	}
	return FloatType{Value: toDegrees(math.Atan2(y.Value, x.Value))}, nil
}
