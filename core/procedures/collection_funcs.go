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

package procedures

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"

	"github.com/dolthub/doltgresql/core/id"
	pgmerge "github.com/dolthub/doltgresql/core/merge"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// storage is used to read from and write to the root.
var storage = objinterface.RootObjectSerializer{
	Bytes:        (*serial.RootValue).ProceduresBytes,
	RootValueAdd: serial.RootValueAddProcedures,
}

// HandleMerge implements the interface objinterface.Collection.
func (*Collection) HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	ourProc := mro.OurRootObj.(Procedure)
	theirProc := mro.TheirRootObj.(Procedure)
	// Ensure that they have the same identifier
	if ourProc.ID != theirProc.ID {
		return nil, nil, errors.Newf("attempted to merge different procedures: `%s` and `%s`",
			ourProc.Name().String(), theirProc.Name().String())
	}
	ourHash, err := ourProc.HashOf(ctx)
	if err != nil {
		return nil, nil, err
	}
	theirHash, err := theirProc.HashOf(ctx)
	if err != nil {
		return nil, nil, err
	}
	if ourHash.Equal(theirHash) {
		return mro.OurRootObj, &merge.MergeStats{
			Operation:            merge.TableUnmodified,
			Adds:                 0,
			Deletes:              0,
			Modifications:        0,
			DataConflicts:        0,
			SchemaConflicts:      0,
			RootObjectConflicts:  0,
			ConstraintViolations: 0,
		}, nil
	}
	return pgmerge.CreateConflict(ctx, mro.RightSrc, ourProc, theirProc, mro.AncestorRootObj)
}

// LoadCollection implements the interface objinterface.Collection.
func (*Collection) LoadCollection(ctx context.Context, root objinterface.RootValue) (objinterface.Collection, error) {
	return LoadProcedures(ctx, root)
}

// LoadCollectionHash implements the interface objinterface.Collection.
func (*Collection) LoadCollectionHash(ctx context.Context, root objinterface.RootValue) (hash.Hash, error) {
	m, ok, err := storage.GetProllyMap(ctx, root)
	if err != nil || !ok {
		return hash.Hash{}, err
	}
	return m.HashOf(), nil
}

// LoadProcedures loads the procedures collection from the given root.
func LoadProcedures(ctx context.Context, root objinterface.RootValue) (*Collection, error) {
	m, ok, err := storage.GetProllyMap(ctx, root)
	if err != nil {
		return nil, err
	}
	if !ok {
		m, err = prolly.NewEmptyAddressMap(root.NodeStore())
		if err != nil {
			return nil, err
		}
	}
	return NewCollection(ctx, m, root.NodeStore())
}

// ResolveNameFromObjects implements the interface objinterface.Collection.
func (*Collection) ResolveNameFromObjects(ctx context.Context, name doltdb.TableName, rootObjects []objinterface.RootObject) (doltdb.TableName, id.Id, error) {
	tempCollection := Collection{
		accessCache: make(map[id.Procedure]Procedure),
		idCache:     make([]id.Procedure, 0, len(rootObjects)),
	}
	for _, rootObject := range rootObjects {
		if obj, ok := rootObject.(Procedure); ok {
			tempCollection.accessCache[obj.ID] = obj
			tempCollection.idCache = append(tempCollection.idCache, obj.ID)
		}
	}
	return tempCollection.ResolveName(ctx, name)
}

// Serializer implements the interface objinterface.Collection.
func (*Collection) Serializer() objinterface.RootObjectSerializer {
	return storage
}

// UpdateRoot implements the interface objinterface.Collection.
func (pgp *Collection) UpdateRoot(ctx context.Context, root objinterface.RootValue) (objinterface.RootValue, error) {
	m, err := pgp.Map(ctx)
	if err != nil {
		return nil, err
	}
	return storage.WriteProllyMap(ctx, root, m)
}
