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

// numericUpcastsMap holds all valid automatic upcasts from an expression to the parameter.
var numericUpcastsMap = map[specialOverloadCast]struct{}{
	{Expression: pgtypes.DoltgresTypeBaseID_Float32, Parameter: pgtypes.DoltgresTypeBaseID_Float32}: {},
	{Expression: pgtypes.DoltgresTypeBaseID_Float32, Parameter: pgtypes.DoltgresTypeBaseID_Float64}: {},
	{Expression: pgtypes.DoltgresTypeBaseID_Float32, Parameter: pgtypes.DoltgresTypeBaseID_Numeric}: {},
	{Expression: pgtypes.DoltgresTypeBaseID_Float64, Parameter: pgtypes.DoltgresTypeBaseID_Float64}: {},
	{Expression: pgtypes.DoltgresTypeBaseID_Float64, Parameter: pgtypes.DoltgresTypeBaseID_Numeric}: {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int16, Parameter: pgtypes.DoltgresTypeBaseID_Float32}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int16, Parameter: pgtypes.DoltgresTypeBaseID_Float64}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int16, Parameter: pgtypes.DoltgresTypeBaseID_Int16}:     {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int16, Parameter: pgtypes.DoltgresTypeBaseID_Int32}:     {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int16, Parameter: pgtypes.DoltgresTypeBaseID_Int64}:     {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int16, Parameter: pgtypes.DoltgresTypeBaseID_Numeric}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int32, Parameter: pgtypes.DoltgresTypeBaseID_Float32}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int32, Parameter: pgtypes.DoltgresTypeBaseID_Float64}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int32, Parameter: pgtypes.DoltgresTypeBaseID_Int32}:     {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int32, Parameter: pgtypes.DoltgresTypeBaseID_Int64}:     {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int32, Parameter: pgtypes.DoltgresTypeBaseID_Numeric}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int64, Parameter: pgtypes.DoltgresTypeBaseID_Float32}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int64, Parameter: pgtypes.DoltgresTypeBaseID_Float64}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int64, Parameter: pgtypes.DoltgresTypeBaseID_Int64}:     {},
	{Expression: pgtypes.DoltgresTypeBaseID_Int64, Parameter: pgtypes.DoltgresTypeBaseID_Numeric}:   {},
	{Expression: pgtypes.DoltgresTypeBaseID_Numeric, Parameter: pgtypes.DoltgresTypeBaseID_Numeric}: {},
}

// implicitOverloadCasts uses the collection of implicit casts for overload resolution.
func implicitOverloadCasts(param pgtypes.DoltgresTypeBaseID, expr pgtypes.DoltgresTypeBaseID) bool {
	f := GetImplicitCast(expr, param)
	return f != nil
}

// numericUpcasts uses the map of numeric upcasts for overload resolution.
func numericUpcasts(param pgtypes.DoltgresTypeBaseID, expr pgtypes.DoltgresTypeBaseID) bool {
	_, ok := numericUpcastsMap[specialOverloadCast{
		Parameter:  param,
		Expression: expr,
	}]
	return ok
}

// castPriorityForType returns the priority for the given type. The lower the value, the higher the priority. The
// priorities are slightly different if we're casting from a string literal.
func castPriorityForType(t pgtypes.DoltgresTypeBaseID, sourceStringLiteral bool) int {
	stringAdjustment := 0
	if sourceStringLiteral {
		stringAdjustment = 1
	}
	switch t {
	case pgtypes.DoltgresTypeBaseID_Numeric:
		return 1 + (2 * stringAdjustment)
	case pgtypes.DoltgresTypeBaseID_Float64:
		return 2 - stringAdjustment
	case pgtypes.DoltgresTypeBaseID_Float32:
		return 3 - (2 * stringAdjustment)
	case pgtypes.DoltgresTypeBaseID_Int64:
		return 4
	case pgtypes.DoltgresTypeBaseID_Int32:
		return 5
	case pgtypes.DoltgresTypeBaseID_Int16:
		return 6
	case pgtypes.DoltgresTypeBaseID_Text:
		return 7
	case pgtypes.DoltgresTypeBaseID_VarChar:
		return 8
	case pgtypes.DoltgresTypeBaseID_Char:
		return 9
	default:
		return 10
	}
}
