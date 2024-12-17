// Copyright 2024 Dolthub, Inc.
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

package id

import (
	"hash/crc32"
	"strconv"
	"sync"
)

// builtinOidLimit is the largest OID that Postgres will assign to built-in items, so we use this to mitigate conflicts
// with existing and future built-in OIDs.
const builtinOidLimit = 65535

var (
	// crcTable is the table that is used for our CRC operations.
	crcTable = crc32.MakeTable(crc32.Castagnoli)
	// globalCache is the cache structure that is used for the server session.
	globalCache = &cacheStruct{
		mutex:      &sync.RWMutex{},
		toOID:      map[Internal]uint32{Null: 0},
		toInternal: map[uint32]Internal{0: Null},
	}
)

// cacheStruct is the cache structure that holds mappings between the Internal ID and external OID (used by Postgres).
// The mappings are temporary, and exist only within a server session. We must discourage users from storing converted
// OIDs, and to use the actual OID type, since the type uses Internal IDs so long as it's not returned to the user.
type cacheStruct struct {
	mutex      *sync.RWMutex
	toOID      map[Internal]uint32
	toInternal map[uint32]Internal
}

// Cache returns the global cache that is used for the server session.
func Cache() *cacheStruct {
	return globalCache
}

// ToOID returns the OID associated with the given Internal ID.
func (cache *cacheStruct) ToOID(id Internal) uint32 {
	// If the ID is in the cache, then we can just return its associated OID
	cache.mutex.RLock()
	if oid, ok := cache.toOID[id]; ok {
		cache.mutex.RUnlock()
		return oid
	}
	cache.mutex.RUnlock()
	if id.Section() == Section_OID {
		oid, _ := strconv.ParseUint(id.Segment(0), 10, 32)
		return uint32(oid)
	}
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	underlyingBytes := id.UnderlyingBytes()
	oid := crc32.Checksum(underlyingBytes, crcTable)
	// If the generated OID is valid, then we'll add it to the cache and return it
	if _, ok := cache.toInternal[oid]; !ok && oid > builtinOidLimit {
		cache.toOID[id] = oid
		cache.toInternal[oid] = id
		return oid
	}
	// In this case, the OID is not valid, so we'll run a small loop to generate an OID based on the actual ID.
	// This retains some level of determinism for OID to ID relationships.
	modifiedBytes := make([]byte, len(underlyingBytes)+1)
	copy(modifiedBytes[1:], underlyingBytes)
	for i := byte(0); i < 255; i++ {
		modifiedBytes[0] = i
		oid = crc32.Checksum(underlyingBytes, crcTable)
		if _, ok := cache.toInternal[oid]; !ok && oid > builtinOidLimit {
			cache.toOID[id] = oid
			cache.toInternal[oid] = id
			return oid
		}
	}
	// If we're here, then we'll just search for an empty OID as a last resort
	for i := uint32(4294967295); i > builtinOidLimit; i-- {
		if _, ok := cache.toInternal[oid]; !ok {
			cache.toOID[id] = oid
			cache.toInternal[oid] = id
			return oid
		}
	}
	// We must have over 4 billion items in the database, so we'll panic since there's nothing we can do
	panic("all OIDs have been taken")
}

// ToInternal returns the Internal ID associated with the given OID.
func (cache *cacheStruct) ToInternal(oid uint32) Internal {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	if id, ok := cache.toInternal[oid]; ok {
		return id
	}
	// The OID is not in the cache, so it's invalid
	return ""
}

// Exists returns whether the given Internal ID exists within the cache. This should primarily be used for the default
// functions, as it's not guaranteed that user functions will be in the cache, especially after a server restart.
func (cache *cacheStruct) Exists(id Internal) bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	_, ok := cache.toOID[id]
	return ok
}

// setBuiltIn sets the given ID to the OID. This should only be used for the built-in items.
func (cache *cacheStruct) setBuiltIn(id Internal, oid uint32) {
	if oid > builtinOidLimit {
		panic("oid is not a built-in")
	}
	cache.toOID[id] = oid
	cache.toInternal[oid] = id
}

// update is used to change the OID mapping of an existing Internal ID that has been changed (where the Internal ID
// points to the same logical item).
//
//lint:ignore U1000 For future use
func (cache *cacheStruct) update(old Internal, new Internal) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	// If the old ID doesn't exist in the cache, then we don't have anything to update
	oid, ok := cache.toOID[old]
	if !ok {
		return
	}
	// We'll delete the old entry and add the new entry, keeping the OID the same for the server session
	delete(cache.toOID, old)
	delete(cache.toInternal, oid)
	cache.toOID[new] = oid
	cache.toInternal[oid] = new
}
