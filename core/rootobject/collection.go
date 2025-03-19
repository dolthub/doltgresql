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

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
)

// CollectionFunctions is the collection of functions that facilitate interaction between the root value and a Collection.
type CollectionFunctions interface {
	// GetID returns the identifying ID for the associated Collection.
	GetID() RootObjectID
	// LoadCollection loads the Collection from the given root.
	LoadCollection(ctx context.Context, root RootValue) (Collection, error)
	// PutCollection updates the Collection in the given root, returning the updated root.
	PutCollection(ctx context.Context, root RootValue, coll Collection) (RootValue, error)
	// Serializer returns the serializer associated with this Collection.
	Serializer() RootObjectSerializer
}

// Collection is a collection of root objects.
type Collection interface {
	// DropRootObject removes the given root object from the collection.
	DropRootObject(ctx context.Context, identifier id.Id) error
	// GetID returns the identifying ID for the Collection.
	GetID() RootObjectID
	// GetRootObject returns the root object matching the given ID. Returns false if it cannot be found.
	GetRootObject(ctx context.Context, identifier id.Id) (RootObject, bool, error)
	// HasRootObject returns whether a root object exactly matching the given ID was found.
	HasRootObject(ctx context.Context, identifier id.Id) (bool, error)
	// IDToTableName converts the given ID to a table name. The table name will be empty for invalid IDs.
	IDToTableName(identifier id.Id) doltdb.TableName
	// IterateIDs iterates over all IDs in the Collection.
	IterateIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error
	// IterateRootObjects iterates over all root objects in the Collection.
	IterateRootObjects(ctx context.Context, callback func(rootObj RootObject) (stop bool, err error)) error
	// PutRootObject updates the Collection with the given root object. This may error if the root object already exists.
	PutRootObject(ctx context.Context, rootObj RootObject) error
	// RenameRootObject changes the ID for a root object matching the old ID.
	RenameRootObject(ctx context.Context, oldID id.Id, newID id.Id) error
	// ResolveName finds the closest matching (or exact) ID for the given name. If an exact match is not found, then
	// this may error if the name is ambiguous.
	ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error)
	// TableNameToID converts the given name to an ID. The ID will be invalid for empty/malformed names.
	TableNameToID(name doltdb.TableName) id.Id
}

var (
	// globalCollectionFunctions maps each ID to the CollectionFunctions.
	globalCollectionFunctions = make([]CollectionFunctions, rootObjectID_NumberOfItems)
	// globalCollectionInitialized tracks whether the functions have finished initializing.
	globalCollectionInitialized = false
)

// RegisterCollection registers the given set of functions.
func RegisterCollection(collection CollectionFunctions) {
	if globalCollectionInitialized {
		panic("cannot register a collection after the root object package has been initialized")
	}
	if collection == nil {
		panic("attempted to register nil collection")
	}
	if globalCollectionFunctions[collection.GetID()] != nil {
		panic("attempted to register duplicate collection")
	}
	globalCollectionFunctions[collection.GetID()] = collection
}

// GetRootObject returns the root object that matches the given name.
func GetRootObject(ctx context.Context, root RootValue, tName doltdb.TableName) (RootObject, bool, error) {
	_, rawID, objID, err := ResolveName(ctx, root, tName)
	if err != nil || objID == RootObjectID_None {
		return nil, false, err
	}
	coll, _ := globalCollectionFunctions[objID].LoadCollection(ctx, root)
	return coll.GetRootObject(ctx, rawID)
}

// LoadAllCollections loads and returns all collections from the root.
func LoadAllCollections(ctx context.Context, root RootValue) ([]Collection, error) {
	colls := make([]Collection, 0, len(globalCollectionFunctions))
	for _, collFuncs := range globalCollectionFunctions {
		if collFuncs == nil {
			continue
		}
		coll, err := collFuncs.LoadCollection(ctx, root)
		if err != nil {
			return nil, err
		}
		colls = append(colls, coll)
	}
	return colls, nil
}

// LoadCollection loads the collection matching the given ID from the root.
func LoadCollection(ctx context.Context, root RootValue, collectionID RootObjectID) (Collection, error) {
	if globalCollectionFunctions[collectionID] == nil {
		return nil, nil
	}
	return globalCollectionFunctions[collectionID].LoadCollection(ctx, root)
}

// PutCollection updates the root with the given collection, returning the updated root.
func PutCollection(ctx context.Context, root RootValue, coll Collection) (RootValue, error) {
	return globalCollectionFunctions[coll.GetID()].PutCollection(ctx, root, coll)
}

// PutRootObject adds the given root object to the respective Collection in the root, returning the updated root.
func PutRootObject(ctx context.Context, root RootValue, tName doltdb.TableName, rootObj RootObject) (RootValue, error) {
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
	return globalCollectionFunctions[rootObj.GetID()].PutCollection(ctx, root, coll)
}

// RemoveRootObject removes the matching root object from its respective Collection, returning the updated root.
func RemoveRootObject(ctx context.Context, root RootValue, identifier id.Id, rootObjectID RootObjectID) (RootValue, error) {
	coll, err := LoadCollection(ctx, root, rootObjectID)
	if err != nil {
		return nil, err
	}
	if err = coll.DropRootObject(ctx, identifier); err != nil {
		return nil, err
	}
	return globalCollectionFunctions[rootObjectID].PutCollection(ctx, root, coll)
}

// ResolveName returns the fully resolved name of the given item (if the item exists). Also returns the type of the item.
func ResolveName(ctx context.Context, root RootValue, name doltdb.TableName) (doltdb.TableName, id.Id, RootObjectID, error) {
	var resolvedName doltdb.TableName
	resolvedRawID := id.Null
	resolvedObjID := RootObjectID_None

	for _, collFuncs := range globalCollectionFunctions {
		if collFuncs == nil {
			continue
		}
		coll, err := collFuncs.LoadCollection(ctx, root)
		if err != nil {
			return doltdb.TableName{}, id.Null, RootObjectID_None, err
		}
		if coll == nil {
			continue
		}
		rName, rID, err := coll.ResolveName(ctx, name)
		if err != nil {
			return doltdb.TableName{}, id.Null, RootObjectID_None, err
		}
		if rID.IsValid() {
			if resolvedObjID != RootObjectID_None {
				return doltdb.TableName{}, id.Null, RootObjectID_None, fmt.Errorf(`"%s" is ambiguous`, name.String())
			}
			resolvedName = rName
			resolvedRawID = rID
			resolvedObjID = collFuncs.GetID()
		}
	}

	return resolvedName, resolvedRawID, resolvedObjID, nil
}
