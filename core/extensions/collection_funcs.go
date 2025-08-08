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

package extensions

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// storage is used to read from and write to the root.
var storage = objinterface.RootObjectSerializer{
	Bytes:        (*serial.RootValue).ExtensionsBytes,
	RootValueAdd: serial.RootValueAddExtensions,
}

// HandleMerge implements the interface objinterface.Collection.
func (*Collection) HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	ourExt := mro.OurRootObj.(Extension)
	theirExt := mro.TheirRootObj.(Extension)
	// Ensure that they have the same ID
	if ourExt.ExtName != theirExt.ExtName {
		return nil, nil, errors.Newf("attempted to merge different extensions: `%s` and `%s`",
			ourExt.Name().String(), theirExt.Name().String())
	}
	ourHash, err := ourExt.HashOf(ctx)
	if err != nil {
		return nil, nil, err
	}
	theirHash, err := theirExt.HashOf(ctx)
	if err != nil {
		return nil, nil, err
	}
	// We always keep the newest extension. I don't think this is actually valid since old extensions are very likely
	// to have invalid function signatures, but this is a start for now.
	// TODO: figure out a better method
	if ourHash.Equal(theirHash) || ourExt.CompareVersions(theirExt) >= 0 {
		return mro.OurRootObj, &merge.MergeStats{
			Operation:            merge.TableUnmodified,
			Adds:                 0,
			Deletes:              0,
			Modifications:        0,
			DataConflicts:        0,
			SchemaConflicts:      0,
			ConstraintViolations: 0,
		}, nil
	} else {
		return mro.TheirRootObj, &merge.MergeStats{
			Operation:            merge.TableModified,
			Adds:                 0,
			Deletes:              0,
			Modifications:        1,
			DataConflicts:        0,
			SchemaConflicts:      0,
			ConstraintViolations: 0,
		}, nil
	}
}

// LoadCollection implements the interface objinterface.Collection.
func (*Collection) LoadCollection(ctx context.Context, root objinterface.RootValue) (objinterface.Collection, error) {
	return LoadExtensions(ctx, root)
}

// LoadCollectionHash implements the interface objinterface.Collection.
func (*Collection) LoadCollectionHash(ctx context.Context, root objinterface.RootValue) (hash.Hash, error) {
	m, ok, err := storage.GetProllyMap(ctx, root)
	if err != nil || !ok {
		return hash.Hash{}, err
	}
	return m.HashOf(), nil
}

// LoadExtensions loads the extensions collection from the given root.
func LoadExtensions(ctx context.Context, root objinterface.RootValue) (*Collection, error) {
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
		accessCache: make(map[id.Extension]Extension),
	}
	for _, rootObject := range rootObjects {
		if obj, ok := rootObject.(Extension); ok {
			tempCollection.accessCache[obj.ExtName] = obj
		}
	}
	return tempCollection.ResolveName(ctx, name)
}

// Serializer implements the interface objinterface.Collection.
func (*Collection) Serializer() objinterface.RootObjectSerializer {
	return storage
}

// UpdateRoot implements the interface objinterface.Collection.
func (pge *Collection) UpdateRoot(ctx context.Context, root objinterface.RootValue) (objinterface.RootValue, error) {
	m, err := pge.Map(ctx)
	if err != nil {
		return nil, err
	}
	return storage.WriteProllyMap(ctx, root, m)
}
