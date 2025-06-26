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

package objinterface

import (
	"context"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
	"github.com/dolthub/dolt/go/store/hash"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/storage"
)

// RootObjectID is an ID that distinguishes names and root objects from one another.
type RootObjectID int64

const (
	RootObjectID_None RootObjectID = iota
	RootObjectID_Sequences
	RootObjectID_Types
	RootObjectID_Functions
	RootObjectID_Triggers
	RootObjectID_Extensions
)

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
	// IterAll iterates over all root objects in the Collection.
	IterAll(ctx context.Context, callback func(rootObj RootObject) (stop bool, err error)) error
	// IterIDs iterates over all IDs in the Collection.
	IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error
	// PutRootObject updates the Collection with the given root object. This may error if the root object already exists.
	PutRootObject(ctx context.Context, rootObj RootObject) error
	// RenameRootObject changes the ID for a root object matching the old ID.
	RenameRootObject(ctx context.Context, oldID id.Id, newID id.Id) error
	// ResolveName finds the closest matching (or exact) ID for the given name. If an exact match is not found, then
	// this may error if the name is ambiguous.
	ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error)
	// TableNameToID converts the given name to an ID. The ID will be invalid for empty/malformed names.
	TableNameToID(name doltdb.TableName) id.Id

	// HandleMerge handles merging of two objects. It is guaranteed that "ours" and "theirs" will not be nil, however
	// "ancestor" may or may not be nil.
	HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error)
	// LoadCollection loads the Collection from the given root.
	LoadCollection(ctx context.Context, root RootValue) (Collection, error)
	// LoadCollectionHash loads the Collection hash from the given root. This does not load the entire collection from
	// the root, and is therefore a bit more performant if only the hash is needed.
	LoadCollectionHash(ctx context.Context, root RootValue) (hash.Hash, error)
	// Serializer returns the serializer associated with this Collection.
	Serializer() RootObjectSerializer
	// UpdateRoot updates the Collection in the given root, returning the updated root.
	UpdateRoot(ctx context.Context, root RootValue) (RootValue, error)
}

// RootValue is an interface to get around import cycles, since the core package references this package (and is where
// RootValue is defined).
type RootValue interface {
	doltdb.RootValue
	// GetStorage returns the storage contained in the root.
	GetStorage(context.Context) storage.RootStorage
	// WithStorage returns an updated RootValue with the given storage.
	WithStorage(context.Context, storage.RootStorage) RootValue
}
