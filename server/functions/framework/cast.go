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
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TODO: Right now, all casts are global. We should decide how to handle this in the presence of branches, sessions, etc.

// getCastFunction is used to recursively call the cast function for when the inner logic sees that it has two array
// types. This sidesteps providing
type getCastFunction func(fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType) pgtypes.TypeCastFunction

// TypeCast is used to cast from one type to another.
type TypeCast struct {
	FromType *pgtypes.DoltgresType
	ToType   *pgtypes.DoltgresType
	Function pgtypes.TypeCastFunction
}

// explicitTypeCastMutex is used to lock the explicit type cast map and array when writing.
var explicitTypeCastMutex = &sync.RWMutex{}

// explicitTypeCastsMap is a map that maps: from -> to -> function.
var explicitTypeCastsMap = map[id.Type]map[id.Type]pgtypes.TypeCastFunction{}

// explicitTypeCastsArray is a slice that holds all registered explicit casts from the given type.
var explicitTypeCastsArray = map[id.Type][]*pgtypes.DoltgresType{}

// assignmentTypeCastMutex is used to lock the assignment type cast map and array when writing.
var assignmentTypeCastMutex = &sync.RWMutex{}

// assignmentTypeCastsMap is a map that maps: from -> to -> function.
var assignmentTypeCastsMap = map[id.Type]map[id.Type]pgtypes.TypeCastFunction{}

// assignmentTypeCastsArray is a slice that holds all registered assignment casts from the given type.
var assignmentTypeCastsArray = map[id.Type][]*pgtypes.DoltgresType{}

// implicitTypeCastMutex is used to lock the implicit type cast map and array when writing.
var implicitTypeCastMutex = &sync.RWMutex{}

// implicitTypeCastsMap is a map that maps: from -> to -> function.
var implicitTypeCastsMap = map[id.Type]map[id.Type]pgtypes.TypeCastFunction{}

// implicitTypeCastsArray is a slice that holds all registered implicit casts from the given type.
var implicitTypeCastsArray = map[id.Type][]*pgtypes.DoltgresType{}

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
func GetPotentialExplicitCasts(fromType id.Type) []*pgtypes.DoltgresType {
	return getPotentialCasts(explicitTypeCastMutex, explicitTypeCastsArray, fromType)
}

// GetPotentialAssignmentCasts returns all registered assignment and implicit type casts from the given type.
func GetPotentialAssignmentCasts(fromType id.Type) []*pgtypes.DoltgresType {
	assignment := getPotentialCasts(assignmentTypeCastMutex, assignmentTypeCastsArray, fromType)
	implicit := GetPotentialImplicitCasts(fromType)
	both := make([]*pgtypes.DoltgresType, len(assignment)+len(implicit))
	copy(both, assignment)
	copy(both[len(assignment):], implicit)
	return both
}

// GetPotentialImplicitCasts returns all registered implicit type casts from the given type.
func GetPotentialImplicitCasts(fromType id.Type) []*pgtypes.DoltgresType {
	return getPotentialCasts(implicitTypeCastMutex, implicitTypeCastsArray, fromType)
}

// GetExplicitCast returns the explicit type cast function that will cast the "from" type to the "to" type. Returns nil
// if such a cast is not valid.
func GetExplicitCast(fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType) pgtypes.TypeCastFunction {
	if tcf := getCast(explicitTypeCastMutex, explicitTypeCastsMap, fromType, toType, GetExplicitCast); tcf != nil {
		return tcf
	} else if tcf = getCast(assignmentTypeCastMutex, assignmentTypeCastsMap, fromType, toType, GetExplicitCast); tcf != nil {
		return tcf
	} else if tcf = getCast(implicitTypeCastMutex, implicitTypeCastsMap, fromType, toType, GetExplicitCast); tcf != nil {
		return tcf
	}
	// We check for the identity and sizing casts after checking the maps, as the identity may be overridden by a user.
	if cast := getSizingOrIdentityCast(fromType, toType, true); cast != nil {
		return cast
	}
	if recordCast := getRecordCast(fromType, toType, GetExplicitCast); recordCast != nil {
		return recordCast
	}
	// All types have a built-in explicit cast from string types: https://www.postgresql.org/docs/15/sql-createcast.html
	if fromType.TypCategory == pgtypes.TypeCategory_StringTypes {
		return func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
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
		return func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
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
	// It is always valid to convert from the `unknown` type
	if fromType.ID == pgtypes.Unknown.ID {
		return UnknownLiteralCast
	}
	return nil
}

// GetAssignmentCast returns the assignment type cast function that will cast the "from" type to the "to" type. Returns
// nil if such a cast is not valid.
func GetAssignmentCast(fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType) pgtypes.TypeCastFunction {
	if tcf := getCast(assignmentTypeCastMutex, assignmentTypeCastsMap, fromType, toType, GetAssignmentCast); tcf != nil {
		return tcf
	} else if tcf = getCast(implicitTypeCastMutex, implicitTypeCastsMap, fromType, toType, GetAssignmentCast); tcf != nil {
		return tcf
	}
	// We check for the identity and sizing casts after checking the maps, as the identity may be overridden by a user.
	if cast := getSizingOrIdentityCast(fromType, toType, false); cast != nil {
		return cast
	}
	// We then check for a record to composite cast
	if recordCast := getRecordCast(fromType, toType, GetAssignmentCast); recordCast != nil {
		return recordCast
	}
	// All types have a built-in assignment cast to string types: https://www.postgresql.org/docs/15/sql-createcast.html
	if toType.TypCategory == pgtypes.TypeCategory_StringTypes {
		return func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
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
	// It is always valid to convert from the `unknown` type
	if fromType.ID == pgtypes.Unknown.ID {
		return UnknownLiteralCast
	}
	return nil
}

// GetImplicitCast returns the implicit type cast function that will cast the "from" type to the "to" type. Returns nil
// if such a cast is not valid.
func GetImplicitCast(fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType) pgtypes.TypeCastFunction {
	if tcf := getCast(implicitTypeCastMutex, implicitTypeCastsMap, fromType, toType, GetImplicitCast); tcf != nil {
		return tcf
	}
	// We check for the identity and sizing casts after checking the maps, as the identity may be overridden by a user.
	if cast := getSizingOrIdentityCast(fromType, toType, false); cast != nil {
		return cast
	}
	// We then check for a record to composite cast
	if recordCast := getRecordCast(fromType, toType, GetImplicitCast); recordCast != nil {
		return recordCast
	}
	// It is always valid to convert from the `unknown` type
	if fromType.ID == pgtypes.Unknown.ID {
		return UnknownLiteralCast
	}
	// It is always valid to convert from the `unknown` type
	if fromType.ID == pgtypes.Unknown.ID {
		return UnknownLiteralCast
	}
	return nil
}

// addTypeCast registers the given type cast.
func addTypeCast(mutex *sync.RWMutex,
	castMap map[id.Type]map[id.Type]pgtypes.TypeCastFunction,
	castArray map[id.Type][]*pgtypes.DoltgresType, cast TypeCast) error {
	mutex.Lock()
	defer mutex.Unlock()

	toMap, ok := castMap[cast.FromType.ID]
	if !ok {
		toMap = map[id.Type]pgtypes.TypeCastFunction{}
		castMap[cast.FromType.ID] = toMap
		castArray[cast.FromType.ID] = nil
	}
	if _, ok := toMap[cast.ToType.ID]; ok {
		// TODO: return the actual Postgres error
		return errors.Errorf("cast from `%s` to `%s` already exists", cast.FromType.String(), cast.ToType.String())
	}
	toMap[cast.ToType.ID] = cast.Function
	castArray[cast.FromType.ID] = append(castArray[cast.FromType.ID], cast.ToType)
	return nil
}

// getPotentialCasts returns all registered type casts from the given type.
func getPotentialCasts(mutex *sync.RWMutex, castArray map[id.Type][]*pgtypes.DoltgresType, fromType id.Type) []*pgtypes.DoltgresType {
	mutex.RLock()
	defer mutex.RUnlock()

	return castArray[fromType]
}

// getCast returns the type cast function that will cast the "from" type to the "to" type. Returns nil if such a cast is
// not valid.
func getCast(mutex *sync.RWMutex,
	castMap map[id.Type]map[id.Type]pgtypes.TypeCastFunction,
	fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType, outerFunc getCastFunction) pgtypes.TypeCastFunction {
	mutex.RLock()
	defer mutex.RUnlock()

	if toMap, ok := castMap[fromType.ID]; ok {
		if f, ok := toMap[toType.ID]; ok {
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
			return func(ctx *sql.Context, vals any, targetType *pgtypes.DoltgresType) (any, error) {
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

// getSizingOrIdentityCast returns an identity cast if the two types are exactly the same, and a sizing cast if they
// only differ in their atttypmod values. Returns nil if no functions are matched. This mirrors the behavior as described in:
// https://www.postgresql.org/docs/15/typeconv-query.html
func getSizingOrIdentityCast(fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType, isExplicitCast bool) pgtypes.TypeCastFunction {
	// If we receive different types, then we can return immediately
	if fromType.ID != toType.ID {
		return nil
	}
	// If we have different atttypmod values, then we need to do a sizing cast only if one exists
	if fromType.GetAttTypMod() != toType.GetAttTypMod() {
		// TODO: We don't have any sizing cast functions implemented, so for now we'll approximate using output to input.
		//  We can use the query below to find all implemented sizing cast functions. It's also detailed in the link above.
		//  Lastly, not all sizing functions accept a boolean, but for those that do, we need to see whether true is
		//  used for explicit casts, or whether true is used for implicit casts.
		//      SELECT
		//        format_type(c.castsource, NULL) AS source,
		//        format_type(c.casttarget, NULL) AS target,
		//        p.oid::regprocedure AS func
		//      FROM pg_cast c JOIN pg_proc p ON p.oid = c.castfunc WHERE c.castsource = c.casttarget ORDER BY 1,2;
		return func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
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
	// If there is no sizing cast, then we simply use the identity cast
	return IdentityCast
}

// getRecordCast handles casting from a record type to a composite type (if applicable). Returns nil if not applicable.
func getRecordCast(fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType, passthrough func(*pgtypes.DoltgresType, *pgtypes.DoltgresType) pgtypes.TypeCastFunction) pgtypes.TypeCastFunction {
	// TODO: does casting to a record type always work for any composite type?
	//   https://www.postgresql.org/docs/15/sql-expressions.html#SQL-SYNTAX-ROW-CONSTRUCTORS seems to suggest so
	//   Also not sure if we should use the passthrough, or if we always default to implicit, assignment, or explicit
	if fromType.IsRecordType() && toType.IsCompositeType() {
		// When casting to a composite type, then we must match the arity and have valid casts for every position.
		if toType.IsRecordType() {
			return IdentityCast
		} else {
			return func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
				vals, ok := val.([]pgtypes.RecordValue)
				if !ok {
					return nil, errors.New("casting input error from record type")
				}
				if len(targetType.CompositeAttrs) != len(vals) {
					// TODO: these should go in DETAIL depending on the size
					//   Input has too few columns.
					//   Input has too many columns.
					return nil, errors.Newf("cannot cast type %s to %s", fromType.Name(), targetType.Name())
				}
				typeCollection, err := core.GetTypesCollectionFromContext(ctx)
				if err != nil {
					return nil, err
				}
				outputVals := make([]pgtypes.RecordValue, len(vals))
				for i := range vals {
					valType, ok := vals[i].Type.(*pgtypes.DoltgresType)
					if !ok {
						return nil, errors.New("cannot cast record containing GMS type")
					}
					outputType, err := typeCollection.GetType(ctx, targetType.CompositeAttrs[i].TypeID)
					if err != nil {
						return nil, err
					}
					outputVals[i].Type = outputType
					if vals[i].Value != nil {
						positionCast := passthrough(valType, outputType)
						if positionCast == nil {
							// TODO: this should be the DETAIL, with the actual error being "cannot cast type <FROM_TYPE> to <TO_TYPE>"
							return nil, errors.Newf("Cannot cast type %s to %s in column %d", valType.Name(), outputType.Name(), i+1)
						}
						outputVals[i].Value, err = positionCast(ctx, vals[i].Value, outputType)
						if err != nil {
							return nil, err
						}
					}
				}
				return outputVals, nil
			}
		}
	}
	return nil
}

// IdentityCast returns the input value.
func IdentityCast(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
	return val, nil
}

// UnknownLiteralCast is used when casting from an unknown literal to any type, as unknown literals are treated special in
// some contexts.
func UnknownLiteralCast(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
	if val == nil {
		return nil, nil
	}
	str, err := pgtypes.Unknown.IoOutput(ctx, val)
	if err != nil {
		return nil, err
	}
	return targetType.IoInput(ctx, str)
}
