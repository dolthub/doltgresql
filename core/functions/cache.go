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

package functions

import (
	"context"
	"sync"

	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"

	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// TODO: doc
type Cache struct {
	cache map[hash.Hash]*Collection
}

var globalCache = Cache{cache: make(map[hash.Hash]*Collection)}
var globalMutex = &sync.Mutex{}
var cacheInit = &sync.Once{}

// TODO: doc
func LoadFunctionsFromCache(ctx context.Context, root objinterface.RootValue) (*Collection, error) {
	cacheInit.Do(func() {
		// We want to create an empty map for an empty hash on the first call
		m, err := prolly.NewEmptyAddressMap(root.NodeStore())
		if err != nil {
			panic("error encountered while attempting to create an empty address map")
		}
		coll, err := NewCollection(ctx, m, root.NodeStore())
		if err != nil {
			panic("error encountered while attempting to create an empty collection")
		}
		coll.SetReadOnly()
		globalMutex.Lock()
		globalCache.cache[hash.Hash{}] = coll
		globalCache.cache[m.HashOf()] = coll // Although the map is empty, the hash will not be the empty hash
		globalMutex.Unlock()
	})
	m, ok, err := storage.GetProllyMap(ctx, root)
	if err != nil {
		return nil, err
	}
	var mapHash hash.Hash
	if ok {
		mapHash = m.HashOf()
	}
	globalMutex.Lock()
	coll, ok := globalCache.cache[mapHash]
	globalMutex.Unlock()
	if !ok {
		coll, err = NewCollection(ctx, m, root.NodeStore())
		if err != nil {
			return nil, err
		}
		coll.SetReadOnly()
		globalMutex.Lock()
		globalCache.cache[mapHash] = coll
		globalMutex.Unlock()
	}
	return coll, nil
}

// TODO: doc
func ResetCache() {
	globalMutex.Lock()
	clear(globalCache.cache)
	cacheInit = &sync.Once{}
	globalMutex.Unlock()
}
