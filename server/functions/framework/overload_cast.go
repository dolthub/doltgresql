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

package framework

import pgtypes "github.com/dolthub/doltgresql/server/types"

// specialOverloadCast holds a relationship between a parameter type and an acceptable automatic cast.
type specialOverloadCast struct {
	Parameter  pgtypes.DoltgresTypeBaseID
	Expression pgtypes.DoltgresTypeBaseID
}

// specialOverloadCasts holds all valid automatic casts between a parameter and a given expression.
var specialOverloadCasts = map[specialOverloadCast]struct{}{
	{Parameter: pgtypes.Bool.BaseID(), Expression: pgtypes.VarCharMax.BaseID()}:    {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Float32.BaseID()}:    {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Float64.BaseID()}:    {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Int16.BaseID()}:      {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Int32.BaseID()}:      {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Int64.BaseID()}:      {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Numeric.BaseID()}:    {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.VarCharMax.BaseID()}: {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Float32.BaseID()}:    {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Float64.BaseID()}:    {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Int16.BaseID()}:      {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Int32.BaseID()}:      {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Int64.BaseID()}:      {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Numeric.BaseID()}:    {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.VarCharMax.BaseID()}: {},
	{Parameter: pgtypes.Int16.BaseID(), Expression: pgtypes.Float32.BaseID()}:      {},
	{Parameter: pgtypes.Int16.BaseID(), Expression: pgtypes.Float64.BaseID()}:      {},
	{Parameter: pgtypes.Int16.BaseID(), Expression: pgtypes.Int16.BaseID()}:        {},
	{Parameter: pgtypes.Int16.BaseID(), Expression: pgtypes.Int32.BaseID()}:        {},
	{Parameter: pgtypes.Int16.BaseID(), Expression: pgtypes.Int64.BaseID()}:        {},
	{Parameter: pgtypes.Int16.BaseID(), Expression: pgtypes.Numeric.BaseID()}:      {},
	{Parameter: pgtypes.Int16.BaseID(), Expression: pgtypes.VarCharMax.BaseID()}:   {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.Float32.BaseID()}:      {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.Float64.BaseID()}:      {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.Int16.BaseID()}:        {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.Int32.BaseID()}:        {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.Int64.BaseID()}:        {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.Numeric.BaseID()}:      {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.VarCharMax.BaseID()}:   {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Float32.BaseID()}:      {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Float64.BaseID()}:      {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Int16.BaseID()}:        {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Int32.BaseID()}:        {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Int64.BaseID()}:        {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Numeric.BaseID()}:      {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.VarCharMax.BaseID()}:   {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Float32.BaseID()}:    {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Float64.BaseID()}:    {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Int16.BaseID()}:      {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Int32.BaseID()}:      {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Int64.BaseID()}:      {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Numeric.BaseID()}:    {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.VarCharMax.BaseID()}: {},
	{Parameter: pgtypes.Uuid.BaseID(), Expression: pgtypes.VarCharMax.BaseID()}:    {},
}

// numericUpcasts holds all valid automatic upcasts from an expression to the parameter.
var numericUpcasts = map[specialOverloadCast]struct{}{
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Float32.BaseID()}: {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Int16.BaseID()}:   {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Int32.BaseID()}:   {},
	{Parameter: pgtypes.Float32.BaseID(), Expression: pgtypes.Int64.BaseID()}:   {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Float32.BaseID()}: {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Float64.BaseID()}: {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Int16.BaseID()}:   {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Int32.BaseID()}:   {},
	{Parameter: pgtypes.Float64.BaseID(), Expression: pgtypes.Int64.BaseID()}:   {},
	{Parameter: pgtypes.Int16.BaseID(), Expression: pgtypes.Int16.BaseID()}:     {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.Int16.BaseID()}:     {},
	{Parameter: pgtypes.Int32.BaseID(), Expression: pgtypes.Int32.BaseID()}:     {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Int16.BaseID()}:     {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Int32.BaseID()}:     {},
	{Parameter: pgtypes.Int64.BaseID(), Expression: pgtypes.Int64.BaseID()}:     {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Float32.BaseID()}: {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Float64.BaseID()}: {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Int16.BaseID()}:   {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Int32.BaseID()}:   {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Int64.BaseID()}:   {},
	{Parameter: pgtypes.Numeric.BaseID(), Expression: pgtypes.Numeric.BaseID()}: {},
}

// castPriorityForType returns the priority for the given type. The lower the value, the higher the priority. The
// priorities are slightly different if we're casting from a string literal.
func castPriorityForType(t pgtypes.DoltgresTypeBaseID, sourceStringLiteral bool) int {
	stringAdjustment := 0
	if sourceStringLiteral {
		stringAdjustment = 1
	}
	switch t {
	case pgtypes.Numeric.BaseID():
		return 1 + (2 * stringAdjustment)
	case pgtypes.Float64.BaseID():
		return 2 - stringAdjustment
	case pgtypes.Float32.BaseID():
		return 3 - (2 * stringAdjustment)
	case pgtypes.Int64.BaseID():
		return 4
	case pgtypes.Int32.BaseID():
		return 5
	case pgtypes.Int16.BaseID():
		return 6
	case pgtypes.VarCharMax.BaseID():
		return 7
	default:
		return 8
	}
}
