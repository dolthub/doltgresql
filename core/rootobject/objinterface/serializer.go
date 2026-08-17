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

	"github.com/cockroachdb/errors"
	doltserial "github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"
	flatbuffers "github.com/dolthub/flatbuffers/v23/go"

	"github.com/dolthub/doltgresql/core/storage"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// RootObjectSerializer holds function pointers for the serialization of root objects.
type RootObjectSerializer struct {
	Bytes        func(*serial.RootValue) []byte
	RootValueAdd func(builder *flatbuffers.Builder, sequences flatbuffers.UOffsetT)
}

// CanonicalHash returns the hash that represents the given contents on a root. Contents with no entries are represented
// by the empty hash, so that an emptied collection and one that was never written are the same state.
func CanonicalHash(contents prolly.AddressMap) (hash.Hash, error) {
	count, err := contents.Count()
	if err != nil {
		return hash.Hash{}, err
	}
	if count == 0 {
		return hash.Hash{}, nil
	}
	return contents.HashOf(), nil
}

// LoadProllyMap loads the contents of this serializer's collection from the given root.
func (serializer RootObjectSerializer) LoadProllyMap(ctx context.Context, root RootValue) (prolly.AddressMap, error) {
	h := serializer.RootHash(ctx, root)
	if h.IsEmpty() {
		return prolly.NewEmptyAddressMap(root.NodeStore())
	}
	val, err := root.VRW().ReadValue(ctx, h)
	if err != nil {
		return prolly.AddressMap{}, err
	}
	if val == nil {
		return prolly.NewEmptyAddressMap(root.NodeStore())
	}
	node, fileId, err := tree.NodeFromBytes(val.(types.SerialMessage))
	if err != nil {
		return prolly.AddressMap{}, err
	}
	if fileId != doltserial.AddressMapFileID {
		return prolly.AddressMap{}, errors.Errorf("invalid address map identifier, expected %s, got %s",
			doltserial.AddressMapFileID, fileId)
	}
	return prolly.NewAddressMap(node, root.NodeStore())
}

// RootHash returns the hash that the given root holds for this serializer's collection, without reading its contents.
// An empty hash means that the root holds no contents for the collection.
func (serializer RootObjectSerializer) RootHash(ctx context.Context, root RootValue) hash.Hash {
	hashBytes := serializer.Bytes(root.GetStorage(ctx).SRV)
	if len(hashBytes) != hash.ByteLen {
		return hash.Hash{}
	}
	return hash.New(hashBytes)
}

// WriteProllyMap writes the given contents to the root, returning the updated root. The root is returned unchanged when
// it already holds these contents.
func (serializer RootObjectSerializer) WriteProllyMap(ctx context.Context, root RootValue, contents prolly.AddressMap) (RootValue, error) {
	h, err := CanonicalHash(contents)
	if err != nil {
		return nil, err
	}
	if h.Equal(serializer.RootHash(ctx, root)) {
		return root, nil
	}
	if !h.IsEmpty() {
		ref, err := root.VRW().WriteValue(ctx, tree.ValueFromNode(contents.Node()))
		if err != nil {
			return nil, err
		}
		h = ref.TargetHash()
	}
	newStorage, err := root.GetStorage(ctx).SetRootObjectHash(ctx, storage.RootObjectSerialization(serializer), h)
	if err != nil {
		return nil, err
	}
	return root.WithStorage(ctx, newStorage), nil
}
