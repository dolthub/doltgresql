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

// getCastFunction is used to recursively call the cast function for when the inner logic sees that it has two array
// types. This sidesteps providing
type getCastFunction func(fromType pgtypes.DoltgresType, toType pgtypes.DoltgresType) TypeCastFunction

// TypeCast is used to cast from one type to another.
type TypeCast struct {
	FromType pgtypes.DoltgresType
	ToType   pgtypes.DoltgresType
	Function TypeCastFunction
}

// explicitTypeCastMutex is used to lock the explicit type cast map and array when writing.
var explicitTypeCastMutex = &sync.RWMutex{}

// explicitTypeCastsMap is a map that maps: from -> to -> function.
var explicitTypeCastsMap = map[uint32]map[uint32]TypeCastFunction{}

// explicitTypeCastsArray is a slice that holds all registered explicit casts from the given type.
var explicitTypeCastsArray = map[uint32][]pgtypes.DoltgresType{}

// assignmentTypeCastMutex is used to lock the assignment type cast map and array when writing.
var assignmentTypeCastMutex = &sync.RWMutex{}

// assignmentTypeCastsMap is a map that maps: from -> to -> function.
var assignmentTypeCastsMap = map[uint32]map[uint32]TypeCastFunction{}

// assignmentTypeCastsArray is a slice that holds all registered assignment casts from the given type.
var assignmentTypeCastsArray = map[uint32][]pgtypes.DoltgresType{}

// implicitTypeCastMutex is used to lock the implicit type cast map and array when writing.
var implicitTypeCastMutex = &sync.RWMutex{}

// implicitTypeCastsMap is a map that maps: from -> to -> function.
var implicitTypeCastsMap = map[uint32]map[uint32]TypeCastFunction{}

// implicitTypeCastsArray is a slice that holds all registered implicit casts from the given type.
var implicitTypeCastsArray = map[uint32][]pgtypes.DoltgresType{}

// AddExplicitTypeCast registers the given explicit type cast.
func AddExplicitTypeCast(cast TypeCast) error {
	return addTypeCast(explicitTypeCastMutex, explicitTypeCastsMap, explicitTypeCastsArray, cast)
}

// AddAssignmentTypeCast registers the given assignment type cast.
func AddAssignmentTypeCast(cast TypeCast) error {
	return addTypeCast(assignmentTypeCastMutex, assignmentTypeCastsMap, assignmentTypeCastsArray, cast)
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

// MustAddAssignmentTypeCast registers the given assignment type cast. Panics if an error occurs.
func MustAddAssignmentTypeCast(cast TypeCast) {
	if err := AddAssignmentTypeCast(cast); err != nil {
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
func GetPotentialExplicitCasts(fromType uint32) []pgtypes.DoltgresType {
	return getPotentialCasts(explicitTypeCastMutex, explicitTypeCastsArray, fromType)
}

// GetPotentialAssignmentCasts returns all registered assignment and implicit type casts from the given type.
func GetPotentialAssignmentCasts(fromType uint32) []pgtypes.DoltgresType {
	assignment := getPotentialCasts(assignmentTypeCastMutex, assignmentTypeCastsArray, fromType)
	implicit := GetPotentialImplicitCasts(fromType)
	both := make([]pgtypes.DoltgresType, len(assignment)+len(implicit))
	copy(both, assignment)
	copy(both[len(assignment):], implicit)
	return both
}

// GetPotentialImplicitCasts returns all registered implicit type casts from the given type.
func GetPotentialImplicitCasts(fromType uint32) []pgtypes.DoltgresType {
	return getPotentialCasts(implicitTypeCastMutex, implicitTypeCastsArray, fromType)
}

// GetExplicitCast returns the explicit type cast function that will cast the "from" type to the "to" type. Returns nil
// if such a cast is not valid.
func GetExplicitCast(fromType pgtypes.DoltgresType, toType pgtypes.DoltgresType) TypeCastFunction {
	if tcf := getCast(explicitTypeCastMutex, explicitTypeCastsMap, fromType, toType, GetExplicitCast); tcf != nil {
		return tcf
	} else if tcf = getCast(assignmentTypeCastMutex, assignmentTypeCastsMap, fromType, toType, GetExplicitCast); tcf != nil {
		return tcf
	} else if tcf = getCast(implicitTypeCastMutex, implicitTypeCastsMap, fromType, toType, GetExplicitCast); tcf != nil {
		return tcf
	}
	// We check for the identity after checking the maps, as the identity may be overridden (such as for types that have
	// parameters). If one of the types are a string type, then we do not use the identity, and use the I/O conversions
	// below.
	if fromType.OID == toType.OID && toType.TypCategory != pgtypes.TypeCategory_StringTypes && fromType.TypCategory != pgtypes.TypeCategory_StringTypes {
		return identityCast
	}
	// All types have a built-in explicit cast from string types: https://www.postgresql.org/docs/15/sql-createcast.html
	if fromType.TypCategory == pgtypes.TypeCategory_StringTypes {
		return func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val == nil {
				return nil, nil
			}
			str, err := fromType.IoOutput(ctx, val)
			if err != nil {
				return nil, err
			}
			return targetType.IoInput(ctx, str)
		}
	} else if toType.TypCategory == pgtypes.TypeCategory_StringTypes {
		// All types have a built-in assignment cast to string types, which we can reference in an explicit cast
		return func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val == nil {
				return nil, nil
			}
			str, err := fromType.IoOutput(ctx, val)
			if err != nil {
				return nil, err
			}
			return targetType.IoInput(ctx, str)
		}
	}
	return nil
}

// GetAssignmentCast returns the assignment type cast function that will cast the "from" type to the "to" type. Returns
// nil if such a cast is not valid.
func GetAssignmentCast(fromType pgtypes.DoltgresType, toType pgtypes.DoltgresType) TypeCastFunction {
	if tcf := getCast(assignmentTypeCastMutex, assignmentTypeCastsMap, fromType, toType, GetAssignmentCast); tcf != nil {
		return tcf
	} else if tcf = getCast(implicitTypeCastMutex, implicitTypeCastsMap, fromType, toType, GetAssignmentCast); tcf != nil {
		return tcf
	}
	// We check for the identity after checking the maps, as the identity may be overridden (such as for types that have
	// parameters). If the "to" type is a string type, then we do not use the identity, and use the I/O conversion below.
	if fromType.OID == toType.OID && fromType.TypCategory != pgtypes.TypeCategory_StringTypes {
		return identityCast
	}
	// All types have a built-in assignment cast to string types: https://www.postgresql.org/docs/15/sql-createcast.html
	if toType.TypCategory == pgtypes.TypeCategory_StringTypes {
		return func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val == nil {
				return nil, nil
			}
			str, err := fromType.IoOutput(ctx, val)
			if err != nil {
				return nil, err
			}
			return targetType.IoInput(ctx, str)
		}
	}
	return nil
}

// GetImplicitCast returns the implicit type cast function that will cast the "from" type to the "to" type. Returns nil
// if such a cast is not valid.
func GetImplicitCast(fromType pgtypes.DoltgresType, toType pgtypes.DoltgresType) TypeCastFunction {
	if tcf := getCast(implicitTypeCastMutex, implicitTypeCastsMap, fromType, toType, GetImplicitCast); tcf != nil {
		return tcf
	}
	// We check for the identity after checking the maps, as the identity may be overridden (such as for types that have
	// parameters).
	if fromType.OID == toType.OID {
		return identityCast
	}
	return nil
}

// addTypeCast registers the given type cast.
func addTypeCast(mutex *sync.RWMutex,
	castMap map[uint32]map[uint32]TypeCastFunction,
	castArray map[uint32][]pgtypes.DoltgresType, cast TypeCast) error {
	mutex.Lock()
	defer mutex.Unlock()

	toMap, ok := castMap[cast.FromType.OID]
	if !ok {
		toMap = map[uint32]TypeCastFunction{}
		castMap[cast.FromType.OID] = toMap
		castArray[cast.FromType.OID] = nil
	}
	if _, ok := toMap[cast.ToType.OID]; ok {
		// TODO: return the actual Postgres error
		return fmt.Errorf("cast from `%s` to `%s` already exists", cast.FromType.String(), cast.ToType.String())
	}
	toMap[cast.ToType.OID] = cast.Function
	castArray[cast.FromType.OID] = append(castArray[cast.FromType.OID], cast.ToType)
	return nil
}

// getPotentialCasts returns all registered type casts from the given type.
func getPotentialCasts(mutex *sync.RWMutex, castArray map[uint32][]pgtypes.DoltgresType, fromType uint32) []pgtypes.DoltgresType {
	mutex.RLock()
	defer mutex.RUnlock()

	return castArray[fromType]
}

// getCast returns the type cast function that will cast the "from" type to the "to" type. Returns nil if such a cast is
// not valid.
func getCast(mutex *sync.RWMutex,
	castMap map[uint32]map[uint32]TypeCastFunction,
	fromType pgtypes.DoltgresType, toType pgtypes.DoltgresType, outerFunc getCastFunction) TypeCastFunction {
	mutex.RLock()
	defer mutex.RUnlock()

	if toMap, ok := castMap[fromType.OID]; ok {
		if f, ok := toMap[toType.OID]; ok {
			return f
		}
	}
	// If there isn't a direct mapping, then we need to check if the types are array variants.
	// As long as the base types are convertable, the array variants are also convertable.
	if fromType.IsArrayType() && toType.IsArrayType() {
		fromBaseType := fromType.ArrayBaseType()
		toBaseType := toType.ArrayBaseType()
		if baseCast := outerFunc(fromBaseType, toBaseType); baseCast != nil {
			// We use a closure that can unwrap the slice, since conversion functions expect a singular non-nil value
			return func(ctx *sql.Context, vals any, targetType pgtypes.DoltgresType) (any, error) {
				var err error
				oldVals := vals.([]any)
				newVals := make([]any, len(oldVals))
				for i, oldVal := range oldVals {
					if oldVal == nil {
						continue
					}
					// Some errors are optional depending on the context, so we'll still process all values even
					// after an error is received.
					var nErr error
					targetBaseType := targetType.ArrayBaseType()
					newVals[i], nErr = baseCast(ctx, oldVal, targetBaseType)
					if nErr != nil && err == nil {
						err = nErr
					}
				}
				return newVals, err
			}
		}

	}
	return nil
}

// identityCast returns the input value.
func identityCast(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	return val, nil
}

// UnknownLiteralCast is used when casting from an unknown literal to any type, as unknown literals are treated special in
// some contexts.
func UnknownLiteralCast(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	if val == nil {
		return nil, nil
	}
	str, err := pgtypes.Unknown.IoOutput(ctx, val)
	if err != nil {
		return nil, err
	}
	return targetType.IoInput(ctx, str)
}
