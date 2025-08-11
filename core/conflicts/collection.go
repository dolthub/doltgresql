// Copyright 2025 Dolthub, Inc.
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

package conflicts

import (
	"context"
	"maps"
	"slices"

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

// Collection contains a collection of conflicts.
type Collection struct {
	accessCache   map[id.Id]Conflict         // This cache is used for general access when you know the exact ID
	nameCache     map[doltdb.TableName]id.Id // This cache is used for ID resolution, since exact table names are always used
	idCache       []id.Id                    // This cache simply contains the name of every root object
	mapHash       hash.Hash                  // This is cached so that we don't have to calculate the hash every time
	underlyingMap prolly.AddressMap
	ns            tree.NodeStore
}

// Conflict represents a root object conflict.
type Conflict struct {
	ID           id.Id
	FromHash     string
	RootObjectID objinterface.RootObjectID
	Ours         objinterface.RootObject
	Theirs       objinterface.RootObject
	Ancestor     objinterface.RootObject
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Conflict{}
var _ objinterface.Conflict = Conflict{}
var _ doltdb.ConflictRootObject = Conflict{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, underlyingMap prolly.AddressMap, ns tree.NodeStore) (*Collection, error) {
	collection := &Collection{
		accessCache:   make(map[id.Id]Conflict),
		nameCache:     make(map[doltdb.TableName]id.Id),
		idCache:       nil,
		mapHash:       hash.Hash{},
		underlyingMap: underlyingMap,
		ns:            ns,
	}
	return collection, collection.reloadCaches(ctx)
}

// GetConflict returns the conflict with the given ID. Returns a conflict with an invalid ID if it cannot be found
// (Conflict.ID.IsValid() == false).
func (pgc *Collection) GetConflict(ctx context.Context, conflictID id.Id) (Conflict, error) {
	if conflict, ok := pgc.accessCache[conflictID]; ok {
		return conflict, nil
	}
	return Conflict{}, nil
}

// HasConflict returns whether the conflict is present.
func (pgc *Collection) HasConflict(ctx context.Context, conflictID id.Id) bool {
	_, ok := pgc.accessCache[conflictID]
	return ok
}

// AddConflict adds a new conflict.
func (pgc *Collection) AddConflict(ctx context.Context, conflict Conflict) error {
	// First we'll check to see if it exists
	if _, ok := pgc.accessCache[conflict.ID]; ok {
		return errors.Errorf(`"%s" already has a conflict`, conflict.Ours.Name())
	}

	// Now we'll add the conflict to our map
	data, err := conflict.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pgc.ns.WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pgc.underlyingMap.Editor()
	if err = mapEditor.Add(ctx, string(conflict.ID), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgc.underlyingMap = newMap
	pgc.mapHash = pgc.underlyingMap.HashOf()
	return pgc.reloadCaches(ctx)
}

// DropConflict drops an existing conflict.
func (pgc *Collection) DropConflict(ctx context.Context, conflictIDs ...id.Id) error {
	if len(conflictIDs) == 0 {
		return nil
	}
	// Check that each name exists before performing any deletions
	for _, conflictID := range conflictIDs {
		if _, ok := pgc.accessCache[conflictID]; !ok {
			return errors.Errorf(`conflict %s does not exist`, conflictID.String())
		}
	}

	// Now we'll remove the conflicts from the map
	mapEditor := pgc.underlyingMap.Editor()
	for _, conflictID := range conflictIDs {
		err := mapEditor.Delete(ctx, string(conflictID))
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
	return pgc.reloadCaches(ctx)
}

// iterateIDs iterates over all conflict IDs in the collection.
func (pgc *Collection) iterateIDs(ctx context.Context, callback func(conflictID id.Id) (stop bool, err error)) error {
	for _, conflictID := range pgc.idCache {
		stop, err := callback(conflictID)
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// IterateConflicts iterates over all conflicts in the collection.
func (pgc *Collection) IterateConflicts(ctx context.Context, callback func(conflict Conflict) (stop bool, err error)) error {
	for _, conflictID := range pgc.idCache {
		stop, err := callback(pgc.accessCache[conflictID])
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// Clone returns a new *Collection with the same contents as the original.
func (pgc *Collection) Clone(ctx context.Context) *Collection {
	return &Collection{
		accessCache:   maps.Clone(pgc.accessCache),
		nameCache:     maps.Clone(pgc.nameCache),
		idCache:       slices.Clone(pgc.idCache),
		underlyingMap: pgc.underlyingMap,
		mapHash:       pgc.mapHash,
		ns:            pgc.ns,
	}
}

// Map writes any cached sequences to the underlying map, and then returns the underlying map.
func (pgc *Collection) Map(ctx context.Context) (prolly.AddressMap, error) {
	return pgc.underlyingMap, nil
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

// reloadCaches writes the underlying map's contents to the caches.
func (pgc *Collection) reloadCaches(ctx context.Context) error {
	count, err := pgc.underlyingMap.Count()
	if err != nil {
		return err
	}

	clear(pgc.accessCache)
	clear(pgc.nameCache)
	pgc.mapHash = pgc.underlyingMap.HashOf()
	pgc.idCache = make([]id.Id, 0, count)

	return pgc.underlyingMap.IterAll(ctx, func(_ string, h hash.Hash) error {
		if h.IsEmpty() {
			return nil
		}
		data, err := pgc.ns.ReadBytes(ctx, h)
		if err != nil {
			return err
		}
		conflict, err := DeserializeConflict(ctx, data)
		if err != nil {
			return err
		}
		pgc.accessCache[conflict.ID] = conflict
		pgc.nameCache[conflict.Name()] = conflict.ID
		pgc.idCache = append(pgc.idCache, conflict.ID)
		return nil
	})
}

// DiffCount implements the interface objinterface.Conflict.
func (conflict Conflict) DiffCount(ctx *sql.Context) (int, error) {
	diffs, _, err := conflict.Diffs(ctx)
	return len(diffs), err
}

// Diffs implements the interface objinterface.Conflict.
func (conflict Conflict) Diffs(ctx context.Context) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	return DiffRootObjects(ctx, conflict.RootObjectID, conflict.FromHash, conflict.Ours, conflict.Theirs, conflict.Ancestor)
}

// FieldType implements the interface objinterface.Conflict.
func (conflict Conflict) FieldType(ctx context.Context, name string) *pgtypes.DoltgresType {
	return GetFieldType(ctx, conflict.RootObjectID, name)
}

// GetContainedRootObjectID implements the interface objinterface.RootObject.
func (conflict Conflict) GetContainedRootObjectID() objinterface.RootObjectID {
	return conflict.RootObjectID
}

// GetID implements the interface objinterface.RootObject.
func (conflict Conflict) GetID() id.Id {
	return conflict.ID
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (conflict Conflict) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Conflicts
}

// HashOf implements the interface objinterface.RootObject.
func (conflict Conflict) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := conflict.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface objinterface.RootObject.
func (conflict Conflict) Name() doltdb.TableName {
	if conflict.Ours != nil {
		return conflict.Ours.Name()
	} else {
		return conflict.Theirs.Name()
	}
}

// RemoveDiffs implements the interface doltdb.ConflictRootObject.
func (conflict Conflict) RemoveDiffs(ctx *sql.Context, diffs []doltdb.RootObjectDiff) (doltdb.ConflictRootObject, error) {
	// This is only called from within Dolt, and it specifically deals with modifying the conflict collection on the
	// root. It is therefore safe to clear context values, to ensure the root changes are not overwritten.
	ClearContextValues(ctx)
	// We need to handle the root object field name in a special way, since it's how we select which side to use in full
	if len(diffs) == 1 {
		diff := diffs[0].(objinterface.RootObjectDiff)
		if diff.FieldName == objinterface.FIELD_NAME_ROOT_OBJECT {
			if conflict.Ours == nil {
				conflict.Theirs = nil
			} else {
				conflict.Theirs = conflict.Ours
			}
			conflict.Ancestor = nil
			return conflict, nil
		}
	}
	// Deletion is equivalent to setting the value in "theirs" to "ours", as the same value cannot conflict with itself.
	// We will only reach this point if both "ours" and "theirs" are non-nil entries, so the above relies on safe assumptions.
	for _, doltDiff := range diffs {
		diff := doltDiff.(objinterface.RootObjectDiff)
		newTheirs, err := UpdateField(ctx, conflict.RootObjectID, conflict.Theirs, diff.FieldName, diff.OurValue)
		if err != nil {
			return nil, err
		}
		conflict.Theirs = newTheirs
	}
	return conflict, nil
}

// Rows implements the interface doltdb.ConflictRootObject.
func (conflict Conflict) Rows(ctx *sql.Context) (sql.RowIter, error) {
	diffs, _, err := conflict.Diffs(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]sql.Row, len(diffs))
	for i, diff := range diffs {
		rows[i], err = diff.ToRow(ctx)
		if err != nil {
			return nil, err
		}
	}
	return sql.RowsToRowIter(rows...), nil
}

// Schema implements the interface doltdb.ConflictRootObject.
func (conflict Conflict) Schema(originatingTableName string) sql.Schema {
	sch := objinterface.RootObjectDiffSchema.Copy()
	for _, col := range sch {
		col.Source = originatingTableName
	}
	return sch
}

// UpdateField implements the interface doltdb.ConflictRootObject.
func (conflict Conflict) UpdateField(ctx *sql.Context, o doltdb.RootObjectDiff, n doltdb.RootObjectDiff) (doltdb.ConflictRootObject, error) {
	// This is only called from within Dolt, and it specifically deals with modifying the conflict collection on the
	// root. It is therefore safe to clear context values, to ensure the root changes are not overwritten.
	ClearContextValues(ctx)
	oldDiff := o.(objinterface.RootObjectDiff)
	newDiff := n.(objinterface.RootObjectDiff)
	// We need to handle the root object field name in a special way, since it's how we select which side to use in full
	if oldDiff.FieldName == objinterface.FIELD_NAME_ROOT_OBJECT {
		switch newDiff.OurValue.(string) {
		case objinterface.FIELD_NAME_OURS:
			conflict.Theirs = conflict.Ours
		case objinterface.FIELD_NAME_THEIRS:
			conflict.Ours = conflict.Theirs
		case objinterface.FIELD_NAME_ANCESTOR:
			conflict.Ours = conflict.Ancestor
			conflict.Theirs = conflict.Ancestor
		default:
			return nil, errors.Newf("cannot replace the object with `%s`", newDiff.OurValue.(string))
		}
		conflict.Ancestor = nil
		return conflict, nil
	}
	newOurs, err := UpdateField(ctx, conflict.RootObjectID, conflict.Ours, oldDiff.FieldName, newDiff.OurValue)
	if err != nil {
		return nil, err
	}
	conflict.Ours = newOurs
	return conflict, nil
}

// ClearContextValues clears context values, and is declared in a different package. It is assigned here by an Init
// function to get around import cycles.
var ClearContextValues = func(ctx *sql.Context) {
	panic("ClearContextValues was never initialized")
}

// DiffRootObjects handles conflict diffs, and is declared in a different package. It is assigned here by an Init
// function to get around import cycles.
var DiffRootObjects = func(ctx context.Context, rootObjID objinterface.RootObjectID, fromHash string, ours, theirs, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	return nil, nil, errors.New("DiffRootObjects was never initialized")
}

// GetFieldType handles type fetching for fields, and is declared in a different package. It is assigned here by an Init
// function to get around import cycles.
var GetFieldType = func(ctx context.Context, rootObjID objinterface.RootObjectID, fieldName string) *pgtypes.DoltgresType {
	panic("GetFieldType was never initialized")
}

// ResolveNameExternal handles name resolution across all collection types, and is declared in a different package. It
// is assigned here by an Init function to get around import cycles.
var ResolveNameExternal = func(ctx context.Context, name doltdb.TableName, rootObjects []objinterface.RootObject) (doltdb.TableName, id.Id, error) {
	panic("ResolveNameExternal was never initialized")
}

// UpdateField handles updating fields in a root object, and is declared in a different package. It is assigned here by
// an Init function to get around import cycles.
var UpdateField = func(ctx context.Context, rootObjID objinterface.RootObjectID, rootObject objinterface.RootObject, fieldName string, newValue any) (objinterface.RootObject, error) {
	return nil, errors.New("UpdateField was never initialized")
}
