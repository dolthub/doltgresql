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

package publications

import (
	"context"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"

	"github.com/dolthub/doltgresql/core/id"
	pgmerge "github.com/dolthub/doltgresql/core/merge"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// storage is used to read from and write to the root.
var storage = objinterface.RootObjectSerializer{
	Bytes:        (*serial.RootValue).PublicationsBytes,
	RootValueAdd: serial.RootValueAddPublications,
}

// HandleMerge implements the interface objinterface.Collection.
func (pgp *Collection) HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	return pgmerge.CreateConflict(ctx, mro.RightSrc, mro.OurRootObj, mro.TheirRootObj, mro.AncestorRootObj)
}

// LoadCollection implements the interface objinterface.Collection.
func (*Collection) LoadCollection(ctx context.Context, root objinterface.RootValue) (objinterface.Collection, error) {
	return LoadPublications(ctx, root)
}

// LoadPublications loads the publications collection from the given root.
func LoadPublications(ctx context.Context, root objinterface.RootValue) (*Collection, error) {
	rom, err := objinterface.NewRootObjectMap(ctx, storage, root)
	if err != nil {
		return nil, err
	}
	return NewCollection(ctx, rom)
}

// ResolveNameFromObjects implements the interface objinterface.Collection.
func (*Collection) ResolveNameFromObjects(ctx context.Context, name doltdb.TableName, rootObjects []objinterface.RootObject) (doltdb.TableName, id.Id, error) {
	tempCollection := Collection{
		accessCache: make(map[id.Publication]Publication),
	}
	for _, rootObject := range rootObjects {
		if obj, ok := rootObject.(Publication); ok {
			tempCollection.accessCache[obj.ID] = obj
		}
	}
	return tempCollection.ResolveName(ctx, name)
}

// Serializer implements the interface objinterface.Collection.
func (*Collection) Serializer() objinterface.RootObjectSerializer {
	return storage
}
