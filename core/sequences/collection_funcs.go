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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/doltgresql/core/id"
	merge2 "github.com/dolthub/doltgresql/core/merge"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// storage is used to read from and write to the root.
var storage = objinterface.RootObjectSerializer{
	Bytes:        (*serial.RootValue).SequencesBytes,
	RootValueAdd: serial.RootValueAddSequences,
}

// HandleMerge implements the interface objinterface.Collection.
func (*Collection) HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	ourSeq := mro.OurRootObj.(*Sequence)
	theirSeq := mro.TheirRootObj.(*Sequence)
	// Ensure that they have the same identifier
	if ourSeq.Id != theirSeq.Id {
		return nil, nil, errors.Newf("attempted to merge different sequences: `%s` and `%s`",
			ourSeq.Name().String(), theirSeq.Name().String())
	}
	// Check if an ancestor is present
	var ancSeq Sequence
	hasAncestor := false
	if mro.AncestorRootObj != nil {
		ancSeq = *(mro.AncestorRootObj.(*Sequence))
		hasAncestor = true
	}
	// Take the min/max of fields that aren't dependent on the increment direction
	mergedSeq := *ourSeq
	mergedSeq.Minimum = merge2.ResolveMergeValuesVariadic(ourSeq.Minimum, theirSeq.Minimum, ancSeq.Minimum, hasAncestor, utils.Min)
	mergedSeq.Maximum = merge2.ResolveMergeValuesVariadic(ourSeq.Maximum, theirSeq.Maximum, ancSeq.Maximum, hasAncestor, utils.Max)
	mergedSeq.Cache = merge2.ResolveMergeValuesVariadic(ourSeq.Cache, theirSeq.Cache, ancSeq.Cache, hasAncestor, utils.Min)
	mergedSeq.Cycle = merge2.ResolveMergeValues(ourSeq.Cycle, theirSeq.Cycle, ancSeq.Cycle, hasAncestor, func(ourCycle, theirCycle bool) bool {
		return ourCycle || theirCycle
	})
	// Take the largest type specified
	mergedSeq.DataTypeID = merge2.ResolveMergeValues(ourSeq.DataTypeID, theirSeq.DataTypeID, ancSeq.DataTypeID, hasAncestor, func(ourID, theirID id.Type) id.Type {
		if (ourID == pgtypes.Int16.ID && (theirID == pgtypes.Int32.ID || theirID == pgtypes.Int64.ID)) ||
			(ourID == pgtypes.Int32.ID && theirID == pgtypes.Int64.ID) {
			return theirID
		} else {
			return ourID
		}
	})
	// Handle the fields that are dependent on the increment direction.
	// We'll always take the increment size that's the smallest for the most granularity, along with the one that
	// has progressed the furthest.
	// For opposing increment directions, we'll take whatever is in our collection.
	mergedSeq.Increment = merge2.ResolveMergeValues(ourSeq.Increment, theirSeq.Increment, ancSeq.Increment, hasAncestor, func(ourIncrement, theirIncrement int64) int64 {
		if ourSeq.Increment >= 0 && theirSeq.Increment >= 0 {
			return utils.Min(ourIncrement, theirIncrement)
		} else if ourSeq.Increment < 0 && theirSeq.Increment < 0 {
			return utils.Max(ourIncrement, theirIncrement)
		} else {
			return ourIncrement
		}
	})
	mergedSeq.Start = merge2.ResolveMergeValues(ourSeq.Start, theirSeq.Start, ancSeq.Start, hasAncestor, func(ourStart, theirStart int64) int64 {
		if ourSeq.Increment >= 0 && theirSeq.Increment >= 0 {
			return utils.Min(ourStart, theirStart)
		} else if ourSeq.Increment < 0 && theirSeq.Increment < 0 {
			return utils.Max(ourStart, theirStart)
		} else {
			return ourStart
		}
	})
	mergedSeq.Current = merge2.ResolveMergeValues(ourSeq.Current, theirSeq.Current, ancSeq.Current, hasAncestor, func(ourCurrent, theirCurrent int64) int64 {
		if ourSeq.Increment >= 0 && theirSeq.Increment >= 0 {
			return utils.Max(ourCurrent, theirCurrent)
		} else if ourSeq.Increment < 0 && theirSeq.Increment < 0 {
			return utils.Min(ourCurrent, theirCurrent)
		} else {
			return ourCurrent
		}
	})
	return &mergedSeq, &merge.MergeStats{
		Operation:            merge.TableModified,
		Adds:                 0,
		Deletes:              0,
		Modifications:        1,
		DataConflicts:        0,
		SchemaConflicts:      0,
		ConstraintViolations: 0,
	}, nil
}

// LoadCollection implements the interface objinterface.Collection.
func (*Collection) LoadCollection(ctx context.Context, root objinterface.RootValue) (objinterface.Collection, error) {
	return LoadSequences(ctx, root)
}

// LoadCollectionHash implements the interface objinterface.Collection.
func (*Collection) LoadCollectionHash(ctx context.Context, root objinterface.RootValue) (hash.Hash, error) {
	m, ok, err := storage.GetProllyMap(ctx, root)
	if err != nil || !ok {
		return hash.Hash{}, err
	}
	return m.HashOf(), nil
}

// LoadSequences loads the sequences collection from the given root.
func LoadSequences(ctx context.Context, root objinterface.RootValue) (*Collection, error) {
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
	}, nil
}

// ResolveNameFromObjects implements the interface objinterface.Collection.
func (*Collection) ResolveNameFromObjects(ctx context.Context, name doltdb.TableName, rootObjects []objinterface.RootObject) (doltdb.TableName, id.Id, error) {
	// First we'll check if there are any objects to search through in the first place
	accessedMap := make(map[id.Sequence]*Sequence)
	for _, rootObject := range rootObjects {
		if obj, ok := rootObject.(*Sequence); ok {
			accessedMap[obj.Id] = obj
		}
	}
	if len(accessedMap) == 0 {
		return doltdb.TableName{}, id.Null, nil
	}
	// There are root objects to search through, so we'll create a temporary store
	ns := tree.NewTestNodeStore()
	addressMap, err := prolly.NewEmptyAddressMap(ns)
	if err != nil {
		return doltdb.TableName{}, id.Null, err
	}
	tempCollection := Collection{
		accessedMap:   accessedMap,
		underlyingMap: addressMap,
		ns:            ns,
	}
	return tempCollection.ResolveName(ctx, name)
}

// Serializer implements the interface objinterface.Collection.
func (*Collection) Serializer() objinterface.RootObjectSerializer {
	return storage
}

// UpdateRoot implements the interface objinterface.Collection.
func (pgs *Collection) UpdateRoot(ctx context.Context, root objinterface.RootValue) (objinterface.RootValue, error) {
	m, err := pgs.Map(ctx)
	if err != nil {
		return nil, err
	}
	return storage.WriteProllyMap(ctx, root, m)
}
