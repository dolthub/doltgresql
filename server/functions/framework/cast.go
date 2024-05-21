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
	"sync"

	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TODO: Right now, all casts are global. We should decide how to handle this in the presence of branches, sessions, etc.

// TypeCastFunction is a function that takes a value of a particular kind of type, and returns it as another kind of type.
// The targetType given should match the "To" type used to obtain the cast.
type TypeCastFunction func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error)

// TypeCast is used to cast from one type to another.
type TypeCast struct {
	FromType pgtypes.DoltgresType
	ToType   pgtypes.DoltgresType
	Function TypeCastFunction
}

// explicitTypeCastMutex is used to lock the explicit type cast map and array when writing.
var explicitTypeCastMutex = &sync.RWMutex{}

// explicitTypeCastsMap is a map that maps: from -> to -> function.
var explicitTypeCastsMap = map[pgtypes.DoltgresTypeBaseID]map[pgtypes.DoltgresTypeBaseID]TypeCastFunction{}

// explicitTypeCastsArray is a slice that holds all registered explicit casts from the given type.
var explicitTypeCastsArray = map[pgtypes.DoltgresTypeBaseID][]pgtypes.DoltgresType{}

// implicitTypeCastMutex is used to lock the implicit type cast map and array when writing.
var implicitTypeCastMutex = &sync.RWMutex{}

// implicitTypeCastsMap is a map that maps: from -> to -> function.
var implicitTypeCastsMap = map[pgtypes.DoltgresTypeBaseID]map[pgtypes.DoltgresTypeBaseID]TypeCastFunction{}

// implicitTypeCastsArray is a slice that holds all registered implicit casts from the given type.
var implicitTypeCastsArray = map[pgtypes.DoltgresTypeBaseID][]pgtypes.DoltgresType{}

// AddExplicitTypeCast registers the given explicit type cast.
func AddExplicitTypeCast(cast TypeCast) error {
	return addTypeCast(explicitTypeCastMutex, explicitTypeCastsMap, explicitTypeCastsArray, cast)
}

// AddImplicitTypeCast registers the given implicit type cast.
func AddImplicitTypeCast(cast TypeCast) error {
	return addTypeCast(implicitTypeCastMutex, implicitTypeCastsMap, implicitTypeCastsArray, cast)
}

// MustAddExplicitTypeCast registers the given explicit type cast. Panics if an error occurs.
func MustAddExplicitTypeCast(cast TypeCast) {
	if err := AddExplicitTypeCast(cast); err != nil {
		panic(err)
	}
}

// MustAddImplicitTypeCast registers the given implicit type cast. Panics if an error occurs.
func MustAddImplicitTypeCast(cast TypeCast) {
	if err := AddImplicitTypeCast(cast); err != nil {
		panic(err)
	}
}

// GetPotentialExplicitCasts returns all registered explicit type casts from the given type.
func GetPotentialExplicitCasts(fromType pgtypes.DoltgresTypeBaseID) []pgtypes.DoltgresType {
	return getPotentialCasts(explicitTypeCastMutex, explicitTypeCastsArray, fromType)
}

// GetPotentialImplicitCasts returns all registered implicit type casts from the given type.
func GetPotentialImplicitCasts(fromType pgtypes.DoltgresTypeBaseID) []pgtypes.DoltgresType {
	return getPotentialCasts(implicitTypeCastMutex, implicitTypeCastsArray, fromType)
}

// GetExplicitCast returns the explicit type cast function that will cast the "from" type to the "to" type. Returns nil
// if such a cast is not valid.
func GetExplicitCast(fromType pgtypes.DoltgresTypeBaseID, toType pgtypes.DoltgresTypeBaseID) TypeCastFunction {
	return getCast(explicitTypeCastMutex, explicitTypeCastsMap, fromType, toType)
}

// GetImplicitCast returns the implicit type cast function that will cast the "from" type to the "to" type. Returns nil
// if such a cast is not valid.
func GetImplicitCast(fromType pgtypes.DoltgresTypeBaseID, toType pgtypes.DoltgresTypeBaseID) TypeCastFunction {
	return getCast(implicitTypeCastMutex, implicitTypeCastsMap, fromType, toType)
}

// addTypeCast registers the given type cast.
func addTypeCast(mutex *sync.RWMutex,
	castMap map[pgtypes.DoltgresTypeBaseID]map[pgtypes.DoltgresTypeBaseID]TypeCastFunction,
	castArray map[pgtypes.DoltgresTypeBaseID][]pgtypes.DoltgresType, cast TypeCast) error {
	mutex.Lock()
	defer mutex.Unlock()

	toMap, ok := castMap[cast.FromType.BaseID()]
	if !ok {
		toMap = map[pgtypes.DoltgresTypeBaseID]TypeCastFunction{}
		castMap[cast.FromType.BaseID()] = toMap
		castArray[cast.FromType.BaseID()] = nil
	}
	if _, ok := toMap[cast.ToType.BaseID()]; ok {
		// TODO: return the actual Postgres error
		return fmt.Errorf("cast from `%s` to `%s` already exists", cast.FromType.String(), cast.ToType.String())
	}
	toMap[cast.ToType.BaseID()] = cast.Function
	castArray[cast.FromType.BaseID()] = append(castArray[cast.FromType.BaseID()], cast.ToType)
	return nil
}

// getPotentialCasts returns all registered type casts from the given type.
func getPotentialCasts(mutex *sync.RWMutex, castArray map[pgtypes.DoltgresTypeBaseID][]pgtypes.DoltgresType, fromType pgtypes.DoltgresTypeBaseID) []pgtypes.DoltgresType {
	mutex.RLock()
	defer mutex.RUnlock()

	return castArray[fromType]
}

// getCast returns the type cast function that will cast the "from" type to the "to" type. Returns nil if such a cast is
// not valid.
func getCast(mutex *sync.RWMutex,
	castMap map[pgtypes.DoltgresTypeBaseID]map[pgtypes.DoltgresTypeBaseID]TypeCastFunction,
	fromType pgtypes.DoltgresTypeBaseID, toType pgtypes.DoltgresTypeBaseID) TypeCastFunction {
	mutex.RLock()
	defer mutex.RUnlock()

	if toMap, ok := castMap[fromType]; ok {
		if f, ok := toMap[toType]; ok {
			return f
		}
	}
	// If there isn't a direct mapping, then we need to check if the types are array variants.
	// As long as the base types are convertable, the array variants are also convertable.
	if fromArrayType, ok := pgtypes.IsBaseIDArrayType(fromType); ok {
		if toArrayType, ok := pgtypes.IsBaseIDArrayType(toType); ok {
			if toMap, ok := castMap[fromArrayType.BaseType().BaseID()]; ok {
				if f, ok := toMap[toArrayType.BaseType().BaseID()]; ok {
					// We use a closure that can unwrap the slice, since conversion functions expect a singular non-nil value
					return func(ctx *sql.Context, vals any, targetType pgtypes.DoltgresType) (any, error) {
						var err error
						oldVals := vals.([]any)
						newVals := make([]any, len(oldVals))
						for i, oldVal := range oldVals {
							if oldVal == nil {
								continue
							}
							newVals[i], err = f(ctx, oldVal, targetType.(pgtypes.DoltgresArrayType).BaseType())
							if err != nil {
								return nil, err
							}
						}
						return newVals, nil
					}
				}
			}
		}
	}
	return nil
}
