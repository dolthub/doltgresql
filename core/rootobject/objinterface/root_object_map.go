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

package objinterface

import (
	"context"

	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"
)

// RootObjectMap is the storage that every root object Collection is built on. It contains the map holding the collection's
// serialized contents, along with the hash that the root held when the two were last in sync. Both hashes are needed to
// tell a change that the Collection made itself (`DiffersFrom`) from one that another writer made to the root (`IsStale`).
type RootObjectMap struct {
	serializer RootObjectSerializer
	contents   prolly.AddressMap
	ns         tree.NodeStore
	rootHash   hash.Hash
}

// NewRootObjectMap loads the contents of the given serializer's collection from the given root.
func NewRootObjectMap(ctx context.Context, serializer RootObjectSerializer, root RootValue) (RootObjectMap, error) {
	contents, err := serializer.LoadProllyMap(ctx, root)
	if err != nil {
		return RootObjectMap{}, err
	}
	return RootObjectMap{
		serializer: serializer,
		contents:   contents,
		ns:         root.NodeStore(),
		rootHash:   serializer.RootHash(ctx, root),
	}, nil
}

// NewDetachedRootObjectMap returns an empty RootObjectMap backed by the given node store rather than by a root. This is
// for collections that are assembled in memory and never written.
func NewDetachedRootObjectMap(serializer RootObjectSerializer, ns tree.NodeStore) (RootObjectMap, error) {
	contents, err := prolly.NewEmptyAddressMap(ns)
	if err != nil {
		return RootObjectMap{}, err
	}
	return RootObjectMap{
		serializer: serializer,
		contents:   contents,
		ns:         ns,
	}, nil
}

// Contents returns the map holding the collection's serialized contents.
func (rom *RootObjectMap) Contents() prolly.AddressMap {
	return rom.contents
}

// NodeStore returns the node store that the collection's contents are read from and written to.
func (rom *RootObjectMap) NodeStore() tree.NodeStore {
	return rom.ns
}

// SetContents replaces the collection's contents.
func (rom *RootObjectMap) SetContents(contents prolly.AddressMap) {
	rom.contents = contents
}

// DiffersFrom implements the interface Collection.
func (rom *RootObjectMap) DiffersFrom(ctx context.Context, root RootValue) (bool, error) {
	h, err := CanonicalHash(rom.contents)
	if err != nil {
		return false, err
	}
	return !h.Equal(rom.serializer.RootHash(ctx, root)), nil
}

// IsStale implements the interface Collection.
func (rom *RootObjectMap) IsStale(ctx context.Context, root RootValue) bool {
	return !rom.rootHash.Equal(rom.serializer.RootHash(ctx, root))
}

// UpdateRoot implements the interface Collection.
func (rom *RootObjectMap) UpdateRoot(ctx context.Context, root RootValue) (RootValue, error) {
	newRoot, err := rom.serializer.WriteProllyMap(ctx, root, rom.contents)
	if err != nil {
		return nil, err
	}
	rom.rootHash = rom.serializer.RootHash(ctx, newRoot)
	return newRoot, nil
}
