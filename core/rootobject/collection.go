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

	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/core/typecollection"
)

var (
	// globalCollections maps each ID to the collection.
	globalCollections = []objinterface.Collection{
		nil,
		&sequences.Collection{},
		&typecollection.TypeCollection{},
		&functions.Collection{},
	}
)

// GetRootObject returns the root object that matches the given name.
func GetRootObject(ctx context.Context, root objinterface.RootValue, tName doltdb.TableName) (objinterface.RootObject, bool, error) {
	_, rawID, objID, err := ResolveName(ctx, root, tName)
	if err != nil || objID == objinterface.RootObjectID_None {
		return nil, false, err
	}
	coll, _ := globalCollections[objID].LoadCollection(ctx, root)
	return coll.GetRootObject(ctx, rawID)
}

// HandleMerge handles merging root objects.
func HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	if mro.OurRootObj == nil {
		switch {
		case mro.TheirRootObj != nil && mro.AncestorRootObj != nil:
			return nil, &merge.MergeStats{
				Operation:            merge.TableModified,
				Adds:                 0,
				Deletes:              0,
				Modifications:        0,
				DataConflicts:        1,
				SchemaConflicts:      0,
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
				ConstraintViolations: 0,
			}, nil
		default:
			return nil, nil, errors.New("HandleMerge has somehow reached a default case")
		}
	} else if mro.TheirRootObj == nil {
		switch {
		case mro.AncestorRootObj != nil:
			return nil, &merge.MergeStats{
				Operation:            merge.TableModified,
				Adds:                 0,
				Deletes:              0,
				Modifications:        0,
				DataConflicts:        1,
				SchemaConflicts:      0,
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
				ConstraintViolations: 0,
			}, nil
		default:
			return nil, nil, errors.New("MergeRootObjects has somehow reached a default case")
		}
	}
	identifier := mro.OurRootObj.(objinterface.RootObject).GetID()
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
	for _, emptyColl := range globalCollections {
		if emptyColl == nil {
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
	coll, err := LoadCollection(ctx, root, rootObj.GetID())
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

	for _, emptyColl := range globalCollections {
		if emptyColl == nil {
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
