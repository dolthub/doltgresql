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

package aggregates

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// storage is used to read from and write to the root.
var storage = objinterface.RootObjectSerializer{
	Bytes:        (*serial.RootValue).AggregatesBytes,
	RootValueAdd: serial.RootValueAddAggregates,
}

// HandleMerge implements the interface objinterface.Collection.
func (*Collection) HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	ourAggregate := mro.OurRootObj.(Aggregate)
	theirAggregate := mro.TheirRootObj.(Aggregate)
	// Ensure that they have the same identifier
	if ourAggregate.ID != theirAggregate.ID {
		return nil, nil, errors.Newf("attempted to merge different aggregates: `%s` and `%s`",
			ourAggregate.Name().String(), theirAggregate.Name().String())
	}
	ourHash, err := ourAggregate.HashOf(ctx)
	if err != nil {
		return nil, nil, err
	}
	theirHash, err := theirAggregate.HashOf(ctx)
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
			ConstraintViolations: 0,
		}, nil
	}
	// TODO: figure out a decent merge strategy
	return nil, nil, errors.Errorf("unable to merge `%s`", theirAggregate.Name().String())
}

// LoadCollection implements the interface objinterface.Collection.
func (*Collection) LoadCollection(ctx context.Context, root objinterface.RootValue) (objinterface.Collection, error) {
	return LoadAggregates(ctx, root)
}

// LoadAggregates loads the aggregates collection from the given root.
func LoadAggregates(ctx context.Context, root objinterface.RootValue) (*Collection, error) {
	rom, err := objinterface.NewRootObjectMap(ctx, storage, root)
	if err != nil {
		return nil, err
	}
	return NewCollection(ctx, rom), nil
}

// ResolveNameFromObjects implements the interface objinterface.Collection.
func (*Collection) ResolveNameFromObjects(ctx context.Context, name doltdb.TableName, rootObjects []objinterface.RootObject) (doltdb.TableName, id.Id, error) {
	rom, err := objinterface.NewDetachedRootObjectMap(storage, tree.NewTestNodeStore())
	if err != nil {
		return doltdb.TableName{}, id.Null, err
	}
	tempCollection := NewCollection(ctx, rom)
	for _, rootObject := range rootObjects {
		if a, ok := rootObject.(Aggregate); ok {
			if err = tempCollection.AddAggregate(ctx, a); err != nil {
				return doltdb.TableName{}, id.Null, err
			}
		}
	}
	return tempCollection.ResolveName(ctx, name)
}

// Serializer implements the interface objinterface.Collection.
func (*Collection) Serializer() objinterface.RootObjectSerializer {
	return storage
}
