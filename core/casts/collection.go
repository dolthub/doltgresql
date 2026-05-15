// Copyright 2026 Dolthub, Inc.
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

package casts

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Collection contains a collection of casts.
type Collection struct {
	mapHash       hash.Hash // This is cached so that we don't have to calculate the hash every time
	underlyingMap prolly.AddressMap
	ns            tree.NodeStore
}

// CastType is the type of the cast, indicating which contexts it may be called in.
type CastType uint8

const (
	CastType_Explicit   CastType = 0
	CastType_Assignment CastType = 1
	CastType_Implicit   CastType = 2
)

// builtInCasts contains all casts that are built into the database by default.
var builtInCasts = map[id.Cast]Cast{}

// Cast represents a cast between two types.
type Cast struct {
	ID       id.Cast
	CastType CastType
	Function id.Function
	BuiltIn  pgtypes.TypeCastFunction
	UseInOut bool
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Cast{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, underlyingMap prolly.AddressMap, ns tree.NodeStore) (*Collection, error) {
	collection := &Collection{
		mapHash:       underlyingMap.HashOf(),
		underlyingMap: underlyingMap,
		ns:            ns,
	}
	return collection, nil
}

// GetExplicitCast returns the explicit type cast function that will cast the source type to the target type. Returns
// a Cast with an invalid ID if such a cast is not valid.
func (pgc *Collection) GetExplicitCast(ctx context.Context, sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType) (Cast, error) {
	castID := id.NewCast(sourceType.ID, targetType.ID)
	c, err := pgc.getCast(ctx, castID, sourceType, targetType, CastType_Explicit)
	if err != nil {
		return Cast{}, err
	}
	if c.ID.IsValid() {
		return c, nil
	}
	// We check for the identity and sizing casts after checking the maps, as the identity may be overridden by a user.
	if cast := pgc.getSizingOrIdentityCast(sourceType, targetType, CastType_Explicit); cast.ID.IsValid() {
		return cast, nil
	}
	// We then check for a record to composite cast
	if recordCast := pgc.getRecordCast(sourceType, targetType, CastType_Explicit); recordCast.ID.IsValid() {
		return recordCast, nil
	}
	// All types have a built-in explicit cast from string types: https://www.postgresql.org/docs/15/sql-createcast.html
	if sourceType.TypCategory == pgtypes.TypeCategory_StringTypes {
		return Cast{
			ID:       castID,
			CastType: CastType_Explicit,
			Function: id.NullFunction,
			UseInOut: true,
		}, nil
	} else if targetType.TypCategory == pgtypes.TypeCategory_StringTypes {
		// All types have a built-in assignment cast to string types, which we can reference in an explicit cast
		return Cast{
			ID:       castID,
			CastType: CastType_Explicit,
			Function: id.NullFunction,
			UseInOut: true,
		}, nil
	}
	// It is always valid to convert from the `unknown` type
	if sourceType.ID == pgtypes.Unknown.ID {
		return Cast{
			ID:       castID,
			CastType: CastType_Explicit,
			Function: id.NullFunction,
			UseInOut: true,
		}, nil
	}
	return Cast{}, nil
}

// GetAssignmentCast returns the assignment type cast function that will cast the source type to the target type.
// Returns a Cast with an invalid ID if such a cast is not valid.
func (pgc *Collection) GetAssignmentCast(ctx context.Context, sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType) (Cast, error) {
	castID := id.NewCast(sourceType.ID, targetType.ID)
	c, err := pgc.getCast(ctx, castID, sourceType, targetType, CastType_Assignment)
	if err != nil {
		return Cast{}, err
	}
	if c.ID.IsValid() {
		if c.CastType == CastType_Explicit {
			return Cast{}, nil
		}
		return c, nil
	}
	// We check for the identity and sizing casts after checking the maps, as the identity may be overridden by a user.
	if cast := pgc.getSizingOrIdentityCast(sourceType, targetType, CastType_Assignment); cast.ID.IsValid() {
		return cast, nil
	}
	// We then check for a record to composite cast
	if recordCast := pgc.getRecordCast(sourceType, targetType, CastType_Assignment); recordCast.ID.IsValid() {
		return recordCast, nil
	}
	// All types have a built-in assignment cast to string types: https://www.postgresql.org/docs/15/sql-createcast.html
	if targetType.TypCategory == pgtypes.TypeCategory_StringTypes {
		return Cast{
			ID:       castID,
			CastType: CastType_Assignment,
			Function: id.NullFunction,
			UseInOut: true,
		}, nil
	}
	// It is always valid to convert from the `unknown` type
	if sourceType.ID == pgtypes.Unknown.ID {
		return Cast{
			ID:       castID,
			CastType: CastType_Assignment,
			Function: id.NullFunction,
			UseInOut: true,
		}, nil
	}
	return Cast{}, nil
}

// GetImplicitCast returns the implicit type cast function that will cast the source type to the target type. Returns a
// Cast with an invalid ID if such a cast is not valid.
func (pgc *Collection) GetImplicitCast(ctx context.Context, sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType) (Cast, error) {
	castID := id.NewCast(sourceType.ID, targetType.ID)
	c, err := pgc.getCast(ctx, castID, sourceType, targetType, CastType_Implicit)
	if err != nil {
		return Cast{}, err
	}
	if c.ID.IsValid() {
		if c.CastType == CastType_Implicit {
			return c, nil
		}
		return Cast{}, nil
	}
	// We check for the identity and sizing casts after checking the maps, as the identity may be overridden by a user.
	if cast := pgc.getSizingOrIdentityCast(sourceType, targetType, CastType_Implicit); cast.ID.IsValid() {
		return cast, nil
	}
	// We then check for a record to composite cast
	if recordCast := pgc.getRecordCast(sourceType, targetType, CastType_Implicit); recordCast.ID.IsValid() {
		return recordCast, nil
	}
	// It is always valid to convert from the `unknown` type
	if sourceType.ID == pgtypes.Unknown.ID {
		return Cast{
			ID:       castID,
			CastType: CastType_Implicit,
			Function: id.NullFunction,
			UseInOut: true,
		}, nil
	}
	return Cast{}, nil
}

// getCast is used by each individual Get function to handle the actual fetching of the cast.
func (pgc *Collection) getCast(ctx context.Context, castID id.Cast, sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType, castType CastType) (Cast, error) {
	if c, ok := builtInCasts[castID]; ok {
		return c, nil
	}
	h, err := pgc.underlyingMap.Get(ctx, string(castID))
	if err != nil {
		return Cast{}, err
	}
	if h.IsEmpty() {
		// If there isn't a direct mapping, then we need to check if the types are array variants.
		// As long as the base types are convertable, the array variants are also convertable.
		if sourceType != nil && targetType != nil && sourceType.IsArrayType() && targetType.IsArrayType() {
			fromBaseType := sourceType.ArrayBaseType()
			toBaseType := targetType.ArrayBaseType()
			var baseCast Cast
			switch castType {
			case CastType_Explicit:
				baseCast, err = pgc.GetExplicitCast(ctx, fromBaseType, toBaseType)
				if err != nil {
					return Cast{}, err
				}
			case CastType_Assignment:
				baseCast, err = pgc.GetAssignmentCast(ctx, fromBaseType, toBaseType)
				if err != nil {
					return Cast{}, err
				}
			case CastType_Implicit:
				baseCast, err = pgc.GetImplicitCast(ctx, fromBaseType, toBaseType)
				if err != nil {
					return Cast{}, err
				}
			}
			if baseCast.ID.IsValid() {
				// We use a closure that can unwrap the slice, since conversion functions expect a singular non-nil value
				evalFunc := func(ctx *sql.Context, vals any, sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType) (any, error) {
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
						sourceBaseType := sourceType.ArrayBaseType()
						targetBaseType := targetType.ArrayBaseType()
						newVals[i], nErr = baseCast.Eval(ctx, oldVal, sourceBaseType, targetBaseType)
						if nErr != nil && err == nil {
							err = nErr
						}
					}
					return newVals, err
				}
				return Cast{
					ID:       castID,
					CastType: castType,
					Function: id.NullFunction,
					BuiltIn:  evalFunc,
					UseInOut: false,
				}, nil
			}
		}
		return Cast{}, nil
	}
	data, err := pgc.ns.ReadBytes(ctx, h)
	if err != nil {
		return Cast{}, err
	}
	return DeserializeCast(ctx, data)
}

// getSizingOrIdentityCast returns an identity cast if the two types are exactly the same, and a sizing cast if they
// only differ in their atttypmod values. Returns a Cast with an invalid ID if no cast is matched. This mirrors the
// behavior as described in:
// https://www.postgresql.org/docs/15/typeconv-query.html
func (pgc *Collection) getSizingOrIdentityCast(sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType, castType CastType) Cast {
	// If we receive different types, then we can return immediately
	if sourceType.ID != targetType.ID {
		return Cast{}
	}
	// If we have different atttypmod values, then we need to do a sizing cast only if one exists
	if sourceType.GetAttTypMod() != targetType.GetAttTypMod() {
		// TODO: We don't have any sizing cast functions implemented, so for now we'll approximate using output to input.
		//  We can use the query below to find all implemented sizing cast functions. It's also detailed in the link above.
		//  Lastly, not all sizing functions accept a boolean, but for those that do, we need to see whether true is
		//  used for explicit casts, or whether true is used for implicit casts.
		//      SELECT
		//        format_type(c.castsource, NULL) AS source,
		//        format_type(c.casttarget, NULL) AS target,
		//        p.oid::regprocedure AS func
		//      FROM pg_cast c JOIN pg_proc p ON p.oid = c.castfunc WHERE c.castsource = c.casttarget ORDER BY 1,2;
		return Cast{
			ID:       id.NewCast(sourceType.ID, targetType.ID),
			CastType: castType,
			Function: id.NullFunction,
			UseInOut: true,
		}
	}
	// If there is no sizing cast, then we simply use the identity cast
	return Cast{
		ID:       id.NewCast(sourceType.ID, targetType.ID),
		CastType: castType,
		Function: id.NullFunction,
		UseInOut: false,
	}
}

// getRecordCast handles casting from a record type to a composite type (if applicable). Returns a Cast with an invalid
// ID if not applicable.
func (pgc *Collection) getRecordCast(sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType, castType CastType) Cast {
	// TODO: does casting to a record type always work for any composite type?
	//   https://www.postgresql.org/docs/15/sql-expressions.html#SQL-SYNTAX-ROW-CONSTRUCTORS seems to suggest so
	//   Also not sure if we should use the passthrough, or if we always default to implicit, assignment, or explicit
	if sourceType.IsRecordType() && targetType.IsCompositeType() {
		// When casting to a composite type, then we must match the arity and have valid casts for every position.
		if targetType.IsRecordType() {
			return Cast{
				ID:       id.NewCast(sourceType.ID, targetType.ID),
				CastType: castType,
				Function: id.NullFunction,
				UseInOut: false,
			}
		} else {
			evalFunc := func(ctx *sql.Context, val any, sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType) (any, error) {
				vals, ok := val.([]pgtypes.RecordValue)
				if !ok {
					return nil, errors.New("casting input error from record type")
				}
				if len(targetType.CompositeAttrs) != len(vals) {
					// TODO: these should go in DETAIL depending on the size
					//   Input has too few columns.
					//   Input has too many columns.
					return nil, errors.Errorf("cannot cast type %s to %s", sourceType.Name(), targetType.Name())
				}
				typeCollection, err := pgtypes.GetTypesCollectionFromContext(ctx)
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
						var positionCast Cast
						switch castType {
						case CastType_Explicit:
							positionCast, err = pgc.GetExplicitCast(ctx, valType, outputType)
							if err != nil {
								return nil, err
							}
						case CastType_Assignment:
							positionCast, err = pgc.GetAssignmentCast(ctx, valType, outputType)
							if err != nil {
								return nil, err
							}
						case CastType_Implicit:
							positionCast, err = pgc.GetImplicitCast(ctx, valType, outputType)
							if err != nil {
								return nil, err
							}
						}
						if !positionCast.ID.IsValid() {
							// TODO: this should be the DETAIL, with the actual error being "cannot cast type <FROM_TYPE> to <TO_TYPE>"
							return nil, errors.Errorf("Cannot cast type %s to %s in column %d", valType.Name(), outputType.Name(), i+1)
						}
						outputVals[i].Value, err = positionCast.Eval(ctx, vals[i].Value, valType, outputType)
						if err != nil {
							return nil, err
						}
					}
				}
				return outputVals, nil
			}
			return Cast{
				ID:       id.NewCast(sourceType.ID, targetType.ID),
				CastType: castType,
				Function: id.NullFunction,
				BuiltIn:  evalFunc,
				UseInOut: false,
			}
		}
	}
	return Cast{}
}

// HasCast returns whether the given cast exists.
func (pgc *Collection) HasCast(ctx context.Context, castID id.Cast) bool {
	if _, ok := builtInCasts[castID]; ok {
		return true
	}
	ok, err := pgc.underlyingMap.Has(ctx, string(castID))
	if err == nil && ok {
		return true
	}
	return false
}

// AddCast adds a new cast.
func (pgc *Collection) AddCast(ctx context.Context, cast Cast) error {
	// First we'll check to see if it exists
	if pgc.HasCast(ctx, cast.ID) {
		return errors.Errorf(`cast from type %s to type %s already exists`,
			cast.ID.SourceType().TypeName(), cast.ID.TargetType().TypeName())
	}
	if cast.BuiltIn != nil {
		return errors.Errorf(`cannot create a built-in cast from type %s to type %s`,
			cast.ID.SourceType().TypeName(), cast.ID.TargetType().TypeName())
	}

	// Now we'll add the cast to our map
	data, err := cast.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pgc.ns.WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pgc.underlyingMap.Editor()
	if err = mapEditor.Add(ctx, string(cast.ID), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgc.underlyingMap = newMap
	pgc.mapHash = pgc.underlyingMap.HashOf()
	return nil
}

// DropCast drops an existing cast.
func (pgc *Collection) DropCast(ctx context.Context, castIDs ...id.Cast) error {
	if len(castIDs) == 0 {
		return nil
	}
	// Check that each name exists before performing any deletions
	for _, castID := range castIDs {
		if _, ok := builtInCasts[castID]; !ok {
			return errors.Errorf(`cannot delete built-in cast from type %s to type %s`,
				castID.SourceType().TypeName(), castID.TargetType().TypeName())
		}
		if ok, err := pgc.underlyingMap.Has(ctx, string(castID)); err != nil {
			return err
		} else if !ok {
			return errors.Errorf(`cast from type %s to type %s does not exist`,
				castID.SourceType().TypeName(), castID.TargetType().TypeName())
		}
	}

	// Now we'll remove the casts from the map
	mapEditor := pgc.underlyingMap.Editor()
	for _, castID := range castIDs {
		err := mapEditor.Delete(ctx, string(castID))
		if err != nil {
			return err
		}
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgc.underlyingMap = newMap
	pgc.mapHash = pgc.underlyingMap.HashOf()
	return nil
}

// resolveName returns the fully resolved name of the given cast. Returns an error if the name is ambiguous.
func (pgc *Collection) resolveName(ctx context.Context, schemaName string, formattedName string) (id.Cast, error) {
	if len(formattedName) == 0 {
		return id.NullCast, nil
	}

	// Check for an exact match
	fullID := pgc.tableNameToID(schemaName, formattedName)
	if pgc.HasCast(ctx, fullID) {
		return fullID, nil
	}

	// Otherwise we'll iterate over all the names
	var resolvedID id.Cast
	err := pgc.IterateCasts(ctx, func(c Cast) (stop bool, err error) {
		if !strings.EqualFold(string(c.ID), string(fullID)) {
			return false, nil
		}
		// The above matches, so this counts as a match
		if resolvedID.IsValid() {
			castTableName := CastIDToTableName(c.ID)
			resolvedTableName := CastIDToTableName(resolvedID)
			return true, fmt.Errorf("`%s` is ambiguous, matches `%s` and `%s`",
				formattedName, castTableName.String(), resolvedTableName.String())
		}
		resolvedID = c.ID
		return false, nil
	})
	return resolvedID, err
}

// IterateCasts iterates over all casts in the collection.
func (pgc *Collection) IterateCasts(ctx context.Context, callback func(c Cast) (stop bool, err error)) error {
	for _, cast := range builtInCasts {
		stop, err := callback(cast)
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return pgc.underlyingMap.IterAll(ctx, func(_ string, v hash.Hash) error {
		data, err := pgc.ns.ReadBytes(ctx, v)
		if err != nil {
			return err
		}
		c, err := DeserializeCast(ctx, data)
		if err != nil {
			return err
		}
		stop, err := callback(c)
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
}

// Clone returns a new *Collection with the same contents as the original.
func (pgc *Collection) Clone(ctx context.Context) *Collection {
	return &Collection{
		mapHash:       pgc.mapHash,
		underlyingMap: pgc.underlyingMap,
		ns:            pgc.ns,
	}
}

// Map writes any cached sequences to the underlying map, and then returns the underlying map.
func (pgc *Collection) Map(ctx context.Context) (prolly.AddressMap, error) {
	return pgc.underlyingMap, nil
}

// tableNameToID returns the ID that was encoded via the Name() call, as the returned TableName contains additional
// information (which this is able to process).
func (pgc *Collection) tableNameToID(schemaName string, formattedName string) id.Cast {
	sections := strings.Split(strings.TrimSuffix(strings.TrimPrefix(formattedName, "("), ")"), ")|(")
	if len(sections) != 4 {
		return id.NullCast
	}
	return id.NewCast(id.NewType(sections[0], sections[1]), id.NewType(sections[2], sections[3]))
}

// GetID implements the interface objinterface.RootObject.
func (cast Cast) GetID() id.Id {
	return cast.ID.AsId()
}

// DiffersFrom returns true when the hash that is associated with the underlying map for this collection is different
// from the hash in the given root.
func (pgc *Collection) DiffersFrom(ctx context.Context, root objinterface.RootValue) bool {
	hashOnGivenRoot, err := pgc.LoadCollectionHash(ctx, root)
	if err != nil {
		return true
	}
	if pgc.mapHash.Equal(hashOnGivenRoot) {
		return false
	}
	// An empty map should match an uninitialized collection on the root
	count, err := pgc.underlyingMap.Count()
	if err == nil && count == 0 && hashOnGivenRoot.IsEmpty() {
		return false
	}
	return true
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (cast Cast) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Casts
}

// HashOf implements the interface objinterface.RootObject.
func (cast Cast) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := cast.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface rootobject.RootObject.
func (cast Cast) Name() doltdb.TableName {
	return CastIDToTableName(cast.ID)
}

// Eval evaluates the cast against the given value.
func (cast Cast) Eval(ctx *sql.Context, val any, sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType) (any, error) {
	if cast.UseInOut {
		if val == nil {
			return nil, nil
		}
		output, err := sourceType.IoOutput(ctx, val)
		if err != nil {
			return nil, err
		}
		return targetType.IoInput(ctx, output)
	}
	if cast.BuiltIn != nil {
		return cast.BuiltIn(ctx, val, sourceType, targetType)
	}
	if cast.Function != id.NullFunction {
		// TODO: get the function collection and call the pointed-to function (argument count determines parameters)
		return nil, errors.Errorf(`cannot cast from type %s to type %s as CREATE CAST is not yet implemented`,
			cast.ID.SourceType().TypeName(), cast.ID.TargetType().TypeName())
	}
	// In this case, the values are binary-coercible, but we still check as we may deviate from Postgres for some reason
	if _, _, err := targetType.Convert(ctx, val); err != nil {
		return nil, errors.Errorf(`cast from type %s to type %s is mislabeled as binary-coercible`,
			cast.ID.SourceType().TypeName(), cast.ID.TargetType().TypeName())
	}
	return val, nil
}

// CastIDToTableName returns the ID in a format that's better for user consumption.
func CastIDToTableName(castID id.Cast) doltdb.TableName {
	name := fmt.Sprintf(`(%s)|(%s)|(%s)|(%s)`,
		castID.SourceType().SchemaName(),
		castID.SourceType().TypeName(),
		castID.TargetType().SchemaName(),
		castID.TargetType().TypeName())
	return doltdb.TableName{
		Name:   name,
		Schema: "",
	}
}
