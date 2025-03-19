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

	doltserial "github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"
	flatbuffers "github.com/dolthub/flatbuffers/v23/go"

	"github.com/dolthub/doltgresql/core/storage"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// RootObjectID is an ID that distinguishes names and root objects from one another.
type RootObjectID int64

const (
	RootObjectID_None RootObjectID = iota
	RootObjectID_Sequences
	RootObjectID_Types
	RootObjectID_Functions
	rootObjectID_NumberOfItems // This should always be the last item
)

// RootValue is an interface to get around import cycles, since the core package references this package (and is where
// RootValue is defined).
type RootValue interface {
	doltdb.RootValue
	// GetStorage returns the storage contained in the root.
	GetStorage(context.Context) storage.RootStorage
	// WithStorage returns an updated RootValue with the given storage.
	WithStorage(context.Context, storage.RootStorage) RootValue
}

// RootObjectSerializer holds function pointers for the serialization of root objects.
type RootObjectSerializer struct {
	Bytes        func(*serial.RootValue) []byte
	RootValueAdd func(builder *flatbuffers.Builder, sequences flatbuffers.UOffsetT)
}

// RootObject is an expanded interface on Dolt's root objects.
type RootObject interface {
	doltdb.RootObject
	// GetID returns the ID associated with this root object.
	GetID() RootObjectID
}

// CreateNomsMap creates and returns a new, empty Noms map.
func (serializer RootObjectSerializer) CreateNomsMap(ctx context.Context, root RootValue) (types.Map, error) {
	return types.NewMap(ctx, root.VRW())
}

// CreateProllyMap creates and returns a new, empty Prolly map.
func (serializer RootObjectSerializer) CreateProllyMap(ctx context.Context, root RootValue) (prolly.AddressMap, error) {
	return prolly.NewEmptyAddressMap(root.NodeStore())
}

// GetNomsMap loads the Noms map from the given root, using the internal serialization functions.
func (serializer RootObjectSerializer) GetNomsMap(ctx context.Context, root RootValue) (types.Map, bool, error) {
	val, ok, err := serializer.getValue(ctx, root)
	if err != nil || !ok {
		return types.EmptyMap, ok, err
	}
	return val.(types.Map), true, nil
}

// GetProllyMap loads the Prolly map from the given root, using the internal serialization functions.
func (serializer RootObjectSerializer) GetProllyMap(ctx context.Context, root RootValue) (prolly.AddressMap, bool, error) {
	val, ok, err := serializer.getValue(ctx, root)
	if err != nil || !ok {
		return prolly.AddressMap{}, ok, err
	}
	serialMessage := val.(types.SerialMessage)
	node, fileId, err := tree.NodeFromBytes(serialMessage)
	if err != nil {
		return prolly.AddressMap{}, false, err
	}
	if fileId != doltserial.AddressMapFileID {
		return prolly.AddressMap{}, false, fmt.Errorf("invalid address map identifier, expected %s, got %s", doltserial.AddressMapFileID, fileId)
	}
	addressMap, err := prolly.NewAddressMap(node, root.NodeStore())
	return addressMap, err == nil, err
}

// WriteNomsMap writes the given Noms map to the root, returning the updated root.
func (serializer RootObjectSerializer) WriteNomsMap(ctx context.Context, root RootValue, val types.Map) (RootValue, error) {
	return serializer.writeValue(ctx, root, val)
}

// WriteProllyMap writes the given Prolly map to the root, returning the updated root.
func (serializer RootObjectSerializer) WriteProllyMap(ctx context.Context, root RootValue, val prolly.AddressMap) (RootValue, error) {
	return serializer.writeValue(ctx, root, tree.ValueFromNode(val.Node()))
}

// getValue loads the value from the given root, using the internal serialization functions.
func (serializer RootObjectSerializer) getValue(ctx context.Context, root RootValue) (types.Value, bool, error) {
	hashBytes := serializer.Bytes(root.GetStorage(ctx).SRV)
	if len(hashBytes) == 0 {
		return nil, false, nil
	}
	h := hash.New(hashBytes)
	if h.IsEmpty() {
		return nil, false, nil
	}
	val, err := root.VRW().ReadValue(ctx, h)
	return val, err == nil && val != nil, err
}

// setHash writes the given hash to storage, returning the updated storage.
func (serializer RootObjectSerializer) setHash(ctx context.Context, st storage.RootStorage, h hash.Hash) (storage.RootStorage, error) {
	if len(serializer.Bytes(st.SRV)) > 0 {
		ret := st.Clone()
		copy(serializer.Bytes(ret.SRV), h[:])
		return ret, nil
	} else {
		return st.Clone(), nil
	}
}

// writeValue writes the given value to the root, returning the updated root.
func (serializer RootObjectSerializer) writeValue(ctx context.Context, root RootValue, val types.Value) (RootValue, error) {
	ref, err := root.VRW().WriteValue(ctx, val)
	if err != nil {
		return nil, err
	}
	newStorage, err := serializer.setHash(ctx, root.GetStorage(ctx), ref.TargetHash())
	if err != nil {
		return nil, err
	}
	return root.WithStorage(ctx, newStorage), nil
}
