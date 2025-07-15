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

package rootobject

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"

	"github.com/dolthub/doltgresql/core/conflicts"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/core/triggers"
	"github.com/dolthub/doltgresql/core/typecollection"
)

var (
	// globalCollections maps each ID to the collection.
	globalCollections = []objinterface.Collection{
		nil,
		&sequences.Collection{},
		&typecollection.TypeCollection{},
		&functions.Collection{},
		&triggers.Collection{},
		&extensions.Collection{},
		&conflicts.Collection{},
	}
)

// CreateConflict creates a conflict on the given root for the two root objects.
func CreateConflict(ctx context.Context, rightSrc doltdb.Rootish, o doltdb.RootObject, t doltdb.RootObject, a doltdb.RootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	ours, ok1 := o.(objinterface.RootObject)
	theirs, ok2 := t.(objinterface.RootObject)
	if !ok1 || !ok2 {
		return nil, nil, errors.New("unsupported object found during conflict creation")
	}
	ancestor, _ := a.(objinterface.RootObject) // If this is nil, then the conversion will also be nil (which is fine)
	rightHash, err := rightSrc.HashOf()
	if err != nil {
		return nil, nil, err
	}
	if ours.GetID() != theirs.GetID() {
		return nil, nil, errors.Errorf(`cannot create a conflict between "%s" and "%s"`,
			ours.Name().String(), theirs.Name().String())
	}
	conflict := conflicts.Conflict{
		ID:           ours.GetID(),
		FromHash:     rightHash.String(),
		RootObjectID: ours.GetRootObjectID(),
		Ours:         ours,
		Theirs:       theirs,
		Ancestor:     ancestor,
	}
	diffs, err := conflict.Diffs(ctx)
	if err != nil {
		return nil, nil, err
	}
	if len(diffs) == 0 {
		return nil, nil, errors.Errorf(`cannot create a conflict between "%s" and "%s" when no diffs are produced`,
			ours.Name().String(), theirs.Name().String())
	}
	return conflict, &merge.MergeStats{
		Operation:            merge.TableUnmodified,
		Adds:                 0,
		Deletes:              0,
		Modifications:        0,
		DataConflicts:        0,
		SchemaConflicts:      0,
		RootObjectConflicts:  len(diffs),
		ConstraintViolations: 0,
	}, nil
}

// DeserializeRootObject calls the same-named function on the collection that matches the ID that was given.
func DeserializeRootObject(ctx context.Context, rootObjID objinterface.RootObjectID, data []byte) (objinterface.RootObject, error) {
	if int64(rootObjID) >= int64(len(globalCollections)) {
		return nil, errors.New("unsupported object found, please upgrade the server")
	}
	collection := globalCollections[rootObjID]
	if collection == nil {
		return nil, errors.Errorf("invalid root object ID: %d", rootObjID)
	}
	return collection.DeserializeRootObject(ctx, data)
}

// DiffRootObjects calls the same-named function on the collection that matches the ID that was given.
func DiffRootObjects(ctx context.Context, rootObjID objinterface.RootObjectID, ours, theirs, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, error) {
	if int64(rootObjID) >= int64(len(globalCollections)) {
		return nil, errors.New("unsupported object found, please upgrade the server")
	}
	collection := globalCollections[rootObjID]
	if collection == nil {
		return nil, errors.Errorf("invalid root object ID: %d", rootObjID)
	}
	return collection.DiffRootObjects(ctx, ours, theirs, ancestor)
}

// GetRootObject returns the root object that matches the given name.
func GetRootObject(ctx context.Context, root objinterface.RootValue, tName doltdb.TableName) (objinterface.RootObject, bool, error) {
	_, rawID, objID, err := ResolveName(ctx, root, tName)
	if err != nil || objID == objinterface.RootObjectID_None {
		return nil, false, err
	}
	coll, _ := globalCollections[objID].LoadCollection(ctx, root)
	return coll.GetRootObject(ctx, rawID)
}

// GetRootObjectConflicts returns the conflict root object that matches the given name.
func GetRootObjectConflicts(ctx context.Context, root objinterface.RootValue, tName doltdb.TableName) (conflicts.Conflict, bool, error) {
	_, rawID, objID, err := ResolveName(ctx, root, tName)
	if err != nil || objID == objinterface.RootObjectID_None {
		return conflicts.Conflict{}, false, err
	}
	coll, _ := globalCollections[objinterface.RootObjectID_Conflicts].LoadCollection(ctx, root)
	ro, ok, err := coll.GetRootObject(ctx, rawID)
	if err != nil || !ok {
		return conflicts.Conflict{}, false, err
	}
	return ro.(conflicts.Conflict), true, nil
}

// HandleMerge handles merging root objects.
func HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	if mro.OurRootObj == nil {
		switch {
		case mro.TheirRootObj != nil && mro.AncestorRootObj != nil:
			theirs := mro.TheirRootObj.(objinterface.RootObject)
			ancestor := mro.AncestorRootObj.(objinterface.RootObject)
			rightHash, err := mro.RightSrc.HashOf()
			if err != nil {
				return nil, nil, err
			}
			return conflicts.Conflict{
					ID:           theirs.GetID(),
					FromHash:     rightHash.String(),
					RootObjectID: theirs.GetRootObjectID(),
					Ours:         nil,
					Theirs:       theirs,
					Ancestor:     ancestor,
				}, &merge.MergeStats{
					Operation:            merge.TableModified,
					Adds:                 0,
					Deletes:              0,
					Modifications:        0,
					DataConflicts:        0,
					SchemaConflicts:      0,
					RootObjectConflicts:  1,
					ConstraintViolations: 0,
				}, nil
		case mro.TheirRootObj != nil && mro.AncestorRootObj == nil:
			return mro.TheirRootObj, &merge.MergeStats{
				Operation:            merge.TableAdded,
				Adds:                 0,
				Deletes:              0,
				Modifications:        0,
				DataConflicts:        0,
				SchemaConflicts:      0,
				RootObjectConflicts:  0,
				ConstraintViolations: 0,
			}, nil
		case mro.TheirRootObj == nil && mro.AncestorRootObj != nil:
			return nil, &merge.MergeStats{
				Operation:            merge.TableRemoved,
				Adds:                 0,
				Deletes:              0,
				Modifications:        0,
				DataConflicts:        0,
				SchemaConflicts:      0,
				RootObjectConflicts:  0,
				ConstraintViolations: 0,
			}, nil
		case mro.TheirRootObj == nil && mro.AncestorRootObj == nil:
			return nil, &merge.MergeStats{
				Operation:            merge.TableUnmodified,
				Adds:                 0,
				Deletes:              0,
				Modifications:        0,
				DataConflicts:        0,
				SchemaConflicts:      0,
				RootObjectConflicts:  0,
				ConstraintViolations: 0,
			}, nil
		default:
			return nil, nil, errors.New("HandleMerge has somehow reached a default case")
		}
	} else if mro.TheirRootObj == nil {
		switch {
		case mro.AncestorRootObj != nil:
			ours := mro.OurRootObj.(objinterface.RootObject)
			ancestor := mro.AncestorRootObj.(objinterface.RootObject)
			rightHash, err := mro.RightSrc.HashOf()
			if err != nil {
				return nil, nil, err
			}
			return conflicts.Conflict{
					ID:           ours.GetID(),
					FromHash:     rightHash.String(),
					RootObjectID: ours.GetRootObjectID(),
					Ours:         ours,
					Theirs:       nil,
					Ancestor:     ancestor,
				}, &merge.MergeStats{
					Operation:            merge.TableModified,
					Adds:                 0,
					Deletes:              0,
					Modifications:        0,
					DataConflicts:        0,
					SchemaConflicts:      0,
					RootObjectConflicts:  1,
					ConstraintViolations: 0,
				}, nil
		case mro.AncestorRootObj == nil:
			return mro.OurRootObj, &merge.MergeStats{
				Operation:            merge.TableAdded,
				Adds:                 0,
				Deletes:              0,
				Modifications:        0,
				DataConflicts:        0,
				SchemaConflicts:      0,
				RootObjectConflicts:  0,
				ConstraintViolations: 0,
			}, nil
		default:
			return nil, nil, errors.New("HandleMerge has somehow reached a default case")
		}
	}
	identifier := mro.OurRootObj.(objinterface.RootObject).GetRootObjectID()
	if int64(identifier) >= int64(len(globalCollections)) {
		return nil, nil, errors.New("unsupported root object found, please upgrade Doltgres to the latest version")
	}
	coll := globalCollections[identifier]
	if coll == nil {
		return nil, nil, errors.Newf("invalid root object found, ID: %d", int64(identifier))
	}
	return coll.HandleMerge(ctx, mro)
}

// LoadAllCollections loads and returns all collections from the root.
func LoadAllCollections(ctx context.Context, root objinterface.RootValue) ([]objinterface.Collection, error) {
	colls := make([]objinterface.Collection, 0, len(globalCollections))
	for i, emptyColl := range globalCollections {
		if emptyColl == nil || i == int(objinterface.RootObjectID_Conflicts) {
			continue
		}
		coll, err := emptyColl.LoadCollection(ctx, root)
		if err != nil {
			return nil, err
		}
		colls = append(colls, coll)
	}
	return colls, nil
}

// LoadCollection loads the collection matching the given ID from the root.
func LoadCollection(ctx context.Context, root objinterface.RootValue, collectionID objinterface.RootObjectID) (objinterface.Collection, error) {
	if globalCollections[collectionID] == nil {
		return nil, nil
	}
	return globalCollections[collectionID].LoadCollection(ctx, root)
}

// PutRootObject adds the given root object to the respective Collection in the root, returning the updated root.
func PutRootObject(ctx context.Context, root objinterface.RootValue, tName doltdb.TableName, rootObj objinterface.RootObject) (objinterface.RootValue, error) {
	coll, err := LoadCollection(ctx, root, rootObj.GetRootObjectID())
	if err != nil {
		return nil, err
	}
	identifier := coll.TableNameToID(tName)
	exists, err := coll.HasRootObject(ctx, identifier)
	if err != nil {
		return nil, err
	}
	// If this doesn't exist, it may be because the name is slightly different (e.g. missing schema), and we want to resolve it properly
	if !exists {
		_, resolvedID, err := coll.ResolveName(ctx, tName)
		if err != nil {
			return nil, err
		}
		if resolvedID.IsValid() {
			identifier = resolvedID
			exists = true
		}
	}
	if exists {
		if err = coll.DropRootObject(ctx, identifier); err != nil {
			return nil, err
		}
	}
	if err = coll.PutRootObject(ctx, rootObj); err != nil {
		return nil, err
	}
	return coll.UpdateRoot(ctx, root)
}

// RemoveRootObject removes the matching root object from its respective Collection, returning the updated root.
func RemoveRootObject(ctx context.Context, root objinterface.RootValue, identifier id.Id, rootObjectID objinterface.RootObjectID) (objinterface.RootValue, error) {
	coll, err := LoadCollection(ctx, root, rootObjectID)
	if err != nil {
		return nil, err
	}
	if err = coll.DropRootObject(ctx, identifier); err != nil {
		return nil, err
	}
	return coll.UpdateRoot(ctx, root)
}

// ResolveName returns the fully resolved name of the given item (if the item exists). Also returns the type of the item.
func ResolveName(ctx context.Context, root objinterface.RootValue, name doltdb.TableName) (doltdb.TableName, id.Id, objinterface.RootObjectID, error) {
	var resolvedName doltdb.TableName
	resolvedRawID := id.Null
	resolvedObjID := objinterface.RootObjectID_None

	for i, emptyColl := range globalCollections {
		if emptyColl == nil || i == int(objinterface.RootObjectID_Conflicts) {
			continue
		}
		coll, err := emptyColl.LoadCollection(ctx, root)
		if err != nil {
			return doltdb.TableName{}, id.Null, objinterface.RootObjectID_None, err
		}
		if coll == nil {
			continue
		}
		rName, rID, err := coll.ResolveName(ctx, name)
		if err != nil {
			return doltdb.TableName{}, id.Null, objinterface.RootObjectID_None, err
		}
		if rID.IsValid() {
			if resolvedObjID != objinterface.RootObjectID_None {
				return doltdb.TableName{}, id.Null, objinterface.RootObjectID_None, fmt.Errorf(`"%s" is ambiguous`, name.String())
			}
			resolvedName = rName
			resolvedRawID = rID
			resolvedObjID = coll.GetID()
		}
	}

	return resolvedName, resolvedRawID, resolvedObjID, nil
}
