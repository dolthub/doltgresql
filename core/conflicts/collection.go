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

// Collection contains a collection of functions.
type Collection struct {
	accessCache   map[id.Id]Conflict // This cache is used for general access when you know the exact ID
	idCache       []id.Id            // This cache simply contains the name of every root object
	mapHash       hash.Hash          // This is cached so that we don't have to calculate the hash every time
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
var _ doltdb.ConflictRootObject = Conflict{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, underlyingMap prolly.AddressMap, ns tree.NodeStore) (*Collection, error) {
	collection := &Collection{
		accessCache:   make(map[id.Id]Conflict),
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

// iterateIDs iterates over all function IDs in the collection.
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
		pgc.idCache = append(pgc.idCache, conflict.ID)
		return nil
	})
}

// GetID implements the interface objinterface.RootObject.
func (conflict Conflict) GetID() id.Id {
	return conflict.ID
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (conflict Conflict) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Conflicts
}

// HasConflicts implements the interface objinterface.RootObject.
func (conflict Conflict) HasConflicts(ctx context.Context) (bool, error) {
	return true, nil
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

// Schema implements the interface doltdb.ConflictRootObject.
func (conflict Conflict) Schema(originatingTableName string) sql.Schema {
	return sql.Schema{
		{Name: "from_root_ish", Type: pgtypes.Text, Default: nil, Nullable: false, Source: originatingTableName},
		{Name: "base_value", Type: pgtypes.Text, Default: nil, Nullable: true, Source: originatingTableName},
		{Name: "our_value", Type: pgtypes.Text, Default: nil, Nullable: true, Source: originatingTableName},
		{Name: "our_diff_type", Type: pgtypes.Text, Default: nil, Nullable: false, Source: originatingTableName},
		{Name: "their_value", Type: pgtypes.Text, Default: nil, Nullable: true, Source: originatingTableName},
		{Name: "their_diff_type", Type: pgtypes.Text, Default: nil, Nullable: false, Source: originatingTableName},
		{Name: "dolt_conflict_id", Type: pgtypes.Text, Default: nil, Nullable: false, Source: originatingTableName},
	}
}

// Diffs returns the diffs for the conflict.
func (conflict Conflict) Diffs(ctx context.Context) ([]objinterface.RootObjectDiff, error) {
	return DiffRootObjects(ctx, conflict.RootObjectID, conflict.Ours, conflict.Theirs, conflict.Ancestor)
}

// Rows implements the interface doltdb.ConflictRootObject.
func (conflict Conflict) Rows(ctx *sql.Context) (sql.RowIter, error) {
	diffs, err := conflict.Diffs(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]sql.Row, len(diffs))
	for i, diff := range diffs {
		var baseValue any
		var ourValue any
		var theirValue any
		var ourChange any
		var theirChange any
		if diff.AncestorValue != nil {
			baseValue, err = diff.Type.IoOutput(ctx, diff.AncestorValue)
			if err != nil {
				return nil, err
			}
		}
		if diff.OurValue != nil {
			ourValue, err = diff.Type.IoOutput(ctx, diff.OurValue)
			if err != nil {
				return nil, err
			}
		}
		if diff.TheirValue != nil {
			theirValue, err = diff.Type.IoOutput(ctx, diff.TheirValue)
			if err != nil {
				return nil, err
			}
		}
		switch diff.OurChange {
		case objinterface.RootObjectDiffChange_Added:
			ourChange = "added"
		case objinterface.RootObjectDiffChange_Deleted:
			ourChange = "deleted"
		case objinterface.RootObjectDiffChange_Modified:
			ourChange = "modified"
		}
		switch diff.TheirChange {
		case objinterface.RootObjectDiffChange_Added:
			theirChange = "added"
		case objinterface.RootObjectDiffChange_Deleted:
			theirChange = "deleted"
		case objinterface.RootObjectDiffChange_Modified:
			theirChange = "modified"
		}
		rows[i] = sql.Row{conflict.FromHash, baseValue, ourValue, ourChange, theirValue, theirChange, diff.FieldName}
	}
	return sql.RowsToRowIter(rows...), nil
}

// DiffRootObjects handles conflict diffs, and is declared in a different package. It is assigned here by an Init
// function to get around import cycles.
var DiffRootObjects = func(ctx context.Context, rootObjID objinterface.RootObjectID, ours, theirs, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, error) {
	return nil, errors.New("DiffRootObjects was never initialized")
}
