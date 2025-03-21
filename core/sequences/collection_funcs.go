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

package sequences

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/store/prolly"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// CollectionFunctions contain the functions that are used for Collection to interact as a root object.
type CollectionFunctions struct{}

var _ rootobject.CollectionFunctions = CollectionFunctions{}

// storage is used to read from and write to the root.
var storage = rootobject.RootObjectSerializer{
	Bytes:        (*serial.RootValue).SequencesBytes,
	RootValueAdd: serial.RootValueAddSequences,
}

// GetID implements the interface rootobject.CollectionFunctions.
func (c CollectionFunctions) GetID() rootobject.RootObjectID {
	return rootobject.RootObjectID_Sequences
}

// LoadCollection implements the interface rootobject.CollectionFunctions.
func (c CollectionFunctions) LoadCollection(ctx context.Context, root rootobject.RootValue) (rootobject.Collection, error) {
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
	return &Collection{
		accessedMap:   make(map[id.Sequence]*Sequence),
		underlyingMap: m,
		ns:            root.NodeStore(),
		mutex:         &sync.Mutex{},
	}, nil
}

// PutCollection implements the interface rootobject.CollectionFunctions.
func (c CollectionFunctions) PutCollection(ctx context.Context, root rootobject.RootValue, roColl rootobject.Collection) (rootobject.RootValue, error) {
	coll, ok := roColl.(*Collection)
	if !ok {
		return nil, errors.New("unknown root object collection")
	}
	m, err := coll.Map(ctx)
	if err != nil {
		return nil, err
	}
	return storage.WriteProllyMap(ctx, root, m)
}

// Serializer implements the interface rootobject.CollectionFunctions.
func (c CollectionFunctions) Serializer() rootobject.RootObjectSerializer {
	return storage
}
