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

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/shopspring/decimal"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TODO: Right now, all casts are global. We should decide how to handle this in the presence of branches, sessions, etc.

// TypeCastFunction is a function that takes a value of a particular kind of type, and returns it as another kind of type.
type TypeCastFunction func(ctx Context, val any) (any, error)

// TypeCast is used to cast from one type to another.
type TypeCast struct {
	FromType pgtypes.DoltgresType
	ToType   pgtypes.DoltgresType
	Function TypeCastFunction
}

// typeCastMutex is used to lock the type cast map and array when writing.
var typeCastMutex = &sync.RWMutex{}

// typeCastsMap is a map that maps: from -> to -> function.
var typeCastsMap = map[pgtypes.DoltgresTypeBaseID]map[pgtypes.DoltgresTypeBaseID]TypeCastFunction{}

// typeCastsArray is a slice that holds all registered casts from the given type.
var typeCastsArray = map[pgtypes.DoltgresTypeBaseID][]pgtypes.DoltgresType{}

// AddTypeCast registers the given type cast.
func AddTypeCast(cast TypeCast) error {
	typeCastMutex.Lock()
	defer typeCastMutex.Unlock()

	toMap, ok := typeCastsMap[cast.FromType.BaseID()]
	if !ok {
		toMap = map[pgtypes.DoltgresTypeBaseID]TypeCastFunction{}
		typeCastsMap[cast.FromType.BaseID()] = toMap
		typeCastsArray[cast.FromType.BaseID()] = nil
	}
	if _, ok := toMap[cast.ToType.BaseID()]; ok {
		// TODO: return the actual Postgres error
		return fmt.Errorf("cast from `%s` to `%s` already exists", cast.FromType.String(), cast.ToType.String())
	}
	toMap[cast.ToType.BaseID()] = cast.Function
	typeCastsArray[cast.FromType.BaseID()] = append(typeCastsArray[cast.FromType.BaseID()], cast.ToType)
	return nil
}

// MustAddTypeCast registers the given type cast. Panics if an error occurs.
func MustAddTypeCast(cast TypeCast) {
	if err := AddTypeCast(cast); err != nil {
		panic(err)
	}
}

// GetPotentialCasts returns all registered type casts from the given type.
func GetPotentialCasts(fromType pgtypes.DoltgresTypeBaseID) []pgtypes.DoltgresType {
	typeCastMutex.RLock()
	defer typeCastMutex.RUnlock()

	return typeCastsArray[fromType]
}

// GetCast returns the type cast function that will cast the "from" type to the "to" type. Returns nil if such a cast is
// not valid.
func GetCast(fromType pgtypes.DoltgresTypeBaseID, toType pgtypes.DoltgresTypeBaseID) TypeCastFunction {
	typeCastMutex.RLock()
	defer typeCastMutex.RUnlock()

	if fromType == toType {
		return identityCast
	}
	if toMap, ok := typeCastsMap[fromType]; ok {
		if f, ok := toMap[toType]; ok {
			return f
		}
	}
	return nil
}

func init() {
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float32,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			return float64(val.(float32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float32,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromFloat(float64(val.(float32))), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromFloat(val.(float64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) {
			return int32(val.(int16)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			return int64(val.(int16)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			return float32(val.(int16)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			return float64(val.(int16)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromInt(int64(val.(int16))), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			return int64(val.(int32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			return float32(val.(int32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			return float64(val.(int32)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromInt(int64(val.(int32))), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			return float32(val.(int64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			return float64(val.(int64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromInt(val.(int64)), nil
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarCharMax,
		ToType:   pgtypes.Bool,
		Function: func(ctx Context, val any) (any, error) {
			lowerVal := strings.TrimSpace(strings.ToLower(val.(string)))
			if lowerVal == "true" || lowerVal == "yes" || lowerVal == "on" || lowerVal == "1" {
				return true, nil
			} else if lowerVal == "false" || lowerVal == "no" || lowerVal == "off" || lowerVal == "0" {
				return false, nil
			} else {
				return nil, fmt.Errorf("invalid string value for boolean: %q", val)
			}
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarCharMax,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseFloat(val.(string), 32)
			return float32(out), err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarCharMax,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseFloat(val.(string), 64)
			return out, err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarCharMax,
		ToType:   pgtypes.Int16,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseInt(val.(string), 10, 16)
			return int16(out), err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarCharMax,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseInt(val.(string), 10, 32)
			return int32(out), err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarCharMax,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) {
			out, err := strconv.ParseInt(val.(string), 10, 64)
			return out, err
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.VarCharMax,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) {
			return decimal.NewFromString(val.(string))
		},
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Null,
		ToType:   pgtypes.Bool,
		Function: func(ctx Context, val any) (any, error) { return nil, nil },
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Null,
		ToType:   pgtypes.Float32,
		Function: func(ctx Context, val any) (any, error) { return nil, nil },
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Null,
		ToType:   pgtypes.Float64,
		Function: func(ctx Context, val any) (any, error) { return nil, nil },
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Null,
		ToType:   pgtypes.Int16,
		Function: func(ctx Context, val any) (any, error) { return nil, nil },
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Null,
		ToType:   pgtypes.Int32,
		Function: func(ctx Context, val any) (any, error) { return nil, nil },
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Null,
		ToType:   pgtypes.Int64,
		Function: func(ctx Context, val any) (any, error) { return nil, nil },
	})
	MustAddTypeCast(TypeCast{
		FromType: pgtypes.Null,
		ToType:   pgtypes.Numeric,
		Function: func(ctx Context, val any) (any, error) { return nil, nil },
	})
}

// identityCast simply returns the input.
func identityCast(ctx Context, val any) (any, error) {
	return val, nil
}
