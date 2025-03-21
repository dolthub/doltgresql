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

package typecollection

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TypeCollection contains a collection of types.
type TypeCollection struct {
	accessedMap   map[id.Type]*pgtypes.DoltgresType
	underlyingMap prolly.AddressMap
	ns            tree.NodeStore
	mutex         *sync.Mutex
}

// TypeWrapper is a wrapper around a type that allows it to be used as a root object.
type TypeWrapper struct {
	Type *pgtypes.DoltgresType
}

var _ rootobject.Collection = (*TypeCollection)(nil)
var _ rootobject.RootObject = TypeWrapper{}
var _ doltdb.RootObject = TypeWrapper{}

// CreateType creates a new type.
func (pgs *TypeCollection) CreateType(ctx context.Context, typ *pgtypes.DoltgresType) error {
	// We can check the built-in types without needing a lock
	if _, ok := pgtypes.IDToBuiltInDoltgresType[typ.ID]; ok {
		return pgtypes.ErrTypeAlreadyExists.New(typ.Name())
	}

	// Now we'll acquire the lock
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// Ensure that the type does not already exist in the cache or underlying map
	if _, ok := pgs.accessedMap[typ.ID]; ok {
		return pgtypes.ErrTypeAlreadyExists.New(typ.Name())
	}
	if ok, err := pgs.underlyingMap.Has(ctx, string(typ.ID)); err != nil {
		return err
	} else if ok {
		return pgtypes.ErrTypeAlreadyExists.New(typ.Name())
	}
	// Add it to our cache, which will be written when we do anything permanent
	pgs.accessedMap[typ.ID] = typ
	return nil
}

// DropType drops an existing type.
func (pgs *TypeCollection) DropType(ctx context.Context, names ...id.Type) (err error) {
	// First we'll check if we're trying to drop a built-in type
	for _, name := range names {
		if _, ok := pgtypes.IDToBuiltInDoltgresType[name]; ok {
			return errors.Newf(`cannot drop built-in type "%s"`, name.TypeName())
		}
	}

	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// We need to clear the cache so that we only need to worry about the underlying map
	if err = pgs.writeCache(ctx); err != nil {
		return err
	}
	for _, name := range names {
		if ok, err := pgs.underlyingMap.Has(ctx, string(name)); err != nil {
			return err
		} else if !ok {
			return pgtypes.ErrTypeDoesNotExist.New(name.TypeName())
		}
	}
	// Now we'll remove the types from the underlying map
	mapEditor := pgs.underlyingMap.Editor()
	for _, name := range names {
		if err = mapEditor.Delete(ctx, string(name)); err != nil {
			return err
		}
	}
	pgs.underlyingMap, err = mapEditor.Flush(ctx)
	return err
}

// GetAllTypes returns a map containing all types in the collection, grouped by the schema they're contained in.
// Each type array is also sorted by the type name. It includes built-in types.
func (pgs *TypeCollection) GetAllTypes(ctx context.Context) (typeMap map[string][]*pgtypes.DoltgresType, schemaNames []string, totalCount int, err error) {
	// IterateTypes takes a lock, so we don't take one here
	schemaNamesMap := make(map[string]struct{})
	typeMap = make(map[string][]*pgtypes.DoltgresType)
	err = pgs.IterateTypes(ctx, func(t *pgtypes.DoltgresType) (stop bool, err error) {
		schemaNamesMap[t.ID.SchemaName()] = struct{}{}
		typeMap[t.ID.SchemaName()] = append(typeMap[t.ID.SchemaName()], t)
		totalCount++
		return false, nil
	})
	if err != nil {
		return nil, nil, 0, err
	}
	// Sort the types in the type map
	for _, seqs := range typeMap {
		sort.Slice(seqs, func(i, j int) bool {
			return seqs[i].ID < seqs[j].ID
		})
	}
	// Create and sort the schema names
	schemaNames = make([]string, 0, len(schemaNamesMap))
	for name := range schemaNamesMap {
		schemaNames = append(schemaNames, name)
	}
	sort.Slice(schemaNames, func(i, j int) bool {
		return schemaNames[i] < schemaNames[j]
	})
	return
}

// GetDomainType returns a domain type with the given schema and name.
// Returns nil if the type cannot be found. It checks for domain type.
func (pgs *TypeCollection) GetDomainType(ctx context.Context, name id.Type) (*pgtypes.DoltgresType, error) {
	// GetType takes a lock, so we don't take one here
	t, err := pgs.GetType(ctx, name)
	if err != nil || t == nil {
		return nil, err
	}
	if t.TypType == pgtypes.TypeType_Domain {
		return t, nil
	}
	return nil, nil
}

// GetType returns the type with the given schema and name.
// Returns nil if the type cannot be found.
func (pgs *TypeCollection) GetType(ctx context.Context, name id.Type) (*pgtypes.DoltgresType, error) {
	// We can check the built-in types without needing a lock
	if t, ok := pgtypes.IDToBuiltInDoltgresType[name]; ok {
		return t, nil
	}

	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// Subsequent loads are cached
	if t, ok := pgs.accessedMap[name]; ok {
		return t, nil
	}
	// The initial load is from the internal map
	h, err := pgs.underlyingMap.Get(ctx, string(name))
	if err != nil || h.IsEmpty() {
		return nil, err
	}
	data, err := pgs.ns.ReadBytes(ctx, h)
	if err != nil {
		return nil, err
	}
	t, err := pgtypes.DeserializeType(data)
	if err != nil {
		return nil, err
	}
	pgt := t.(*pgtypes.DoltgresType)
	pgs.accessedMap[pgt.ID] = pgt
	return pgt, nil
}

// HasType checks if a type exists with given schema and type name.
func (pgs *TypeCollection) HasType(ctx context.Context, name id.Type) bool {
	// We can check the built-in types without needing a lock
	if _, ok := pgtypes.IDToBuiltInDoltgresType[name]; ok {
		return true
	}

	// Now we'll acquire the lock
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if _, ok := pgs.accessedMap[name]; ok {
		return true
	}
	ok, err := pgs.underlyingMap.Has(ctx, string(name))
	if err == nil && ok {
		return true
	}
	return false
}

// resolveName returns the fully resolved name of the given type. Returns an error if the name is ambiguous.
func (pgs *TypeCollection) resolveName(ctx context.Context, schemaName string, typeName string) (id.Type, error) {
	// First check for an exact match in the built-in types
	inputID := id.NewType(schemaName, typeName)
	if _, ok := pgtypes.IDToBuiltInDoltgresType[inputID]; ok {
		return inputID, nil
	}

	// Iterate over all the built-in names for a relative match
	var resolvedID id.Type
	for _, typ := range pgtypes.GetAllBuitInTypes() {
		if strings.EqualFold(typeName, typ.ID.TypeName()) {
			if len(schemaName) > 0 && !strings.EqualFold(schemaName, typ.ID.SchemaName()) {
				continue
			}
			if resolvedID.IsValid() {
				return id.NullType, fmt.Errorf("`%s.%s` is ambiguous, matches `%s.%s` and `%s.%s`",
					schemaName, typeName, typ.ID.SchemaName(), typ.ID.TypeName(), resolvedID.SchemaName(), resolvedID.TypeName())
			}
			resolvedID = typ.ID
		}
	}

	// Acquire the lock now that we're looking into the local data
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// We write the cache so that we only need to worry about the underlying map
	if err := pgs.writeCache(ctx); err != nil {
		return id.NullType, err
	}

	// Check for an exact match in the underlying map
	ok, err := pgs.underlyingMap.Has(ctx, string(inputID))
	if err != nil {
		return id.NullType, err
	} else if ok {
		// We don't bother looking if there's an existing match, since this is an exact match (so no ambiguity)
		return inputID, nil
	}

	// Iterate over all the names in the map
	err = pgs.underlyingMap.IterAll(ctx, func(k string, _ hash.Hash) error {
		typeID := id.Type(k)
		if strings.EqualFold(typeName, typeID.TypeName()) {
			if len(schemaName) > 0 && !strings.EqualFold(schemaName, typeID.SchemaName()) {
				return nil
			}
			if resolvedID.IsValid() {
				return fmt.Errorf("`%s.%s` is ambiguous, matches `%s.%s` and `%s.%s`",
					schemaName, typeName, typeID.SchemaName(), typeID.TypeName(), resolvedID.SchemaName(), resolvedID.TypeName())
			}
			resolvedID = typeID
		}
		return nil
	})
	if err != nil {
		return id.NullType, err
	}
	return resolvedID, nil
}

// iterateIDs iterates over all type IDs in the collection.
func (pgs *TypeCollection) iterateIDs(ctx context.Context, f func(seqID id.Type) (stop bool, err error)) error {
	// We can iterate the built-in types without needing a lock
	for _, t := range pgtypes.GetAllBuitInTypes() {
		stop, err := f(t.ID)
		if err != nil || stop {
			return err
		}
	}

	// Now we'll acquire the lock
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// We write the cache so that we only need to worry about the underlying map
	if err := pgs.writeCache(ctx); err != nil {
		return err
	}
	err := pgs.underlyingMap.IterAll(ctx, func(k string, _ hash.Hash) error {
		stop, err := f(id.Type(k))
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
	return err
}

// IterateTypes iterates over all types in the collection.
func (pgs *TypeCollection) IterateTypes(ctx context.Context, f func(typ *pgtypes.DoltgresType) (stop bool, err error)) error {
	// We can iterate the built-in types without needing a lock
	for _, t := range pgtypes.GetAllBuitInTypes() {
		stop, err := f(t)
		if err != nil || stop {
			return err
		}
	}

	// Now we'll acquire the lock
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// We write the cache so that we only need to worry about the underlying map
	if err := pgs.writeCache(ctx); err != nil {
		return err
	}
	err := pgs.underlyingMap.IterAll(ctx, func(_ string, v hash.Hash) error {
		data, err := pgs.ns.ReadBytes(ctx, v)
		if err != nil {
			return err
		}
		t, err := pgtypes.DeserializeType(data)
		if err != nil {
			return err
		}
		stop, err := f(t.(*pgtypes.DoltgresType))
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
	return err
}

// Clone returns a new *TypeCollection with the same contents as the original.
func (pgs *TypeCollection) Clone(ctx context.Context) *TypeCollection {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	newCollection := &TypeCollection{
		accessedMap:   make(map[id.Type]*pgtypes.DoltgresType),
		underlyingMap: pgs.underlyingMap,
		mutex:         &sync.Mutex{},
	}
	for typeID, t := range pgs.accessedMap {
		newCollection.accessedMap[typeID] = t
	}
	return newCollection
}

// Map writes any cached types to the underlying map, and then returns the underlying map.
func (pgs *TypeCollection) Map(ctx context.Context) (prolly.AddressMap, error) {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if err := pgs.writeCache(ctx); err != nil {
		return prolly.AddressMap{}, err
	}
	return pgs.underlyingMap, nil
}

// GetID implements the interface rootobject.RootObject.
func (t TypeWrapper) GetID() rootobject.RootObjectID {
	return rootobject.RootObjectID_Types
}

// HashOf implements the interface rootobject.RootObject.
func (t TypeWrapper) HashOf(ctx context.Context) (hash.Hash, error) {
	if t.Type != nil {
		return hash.Of(t.Type.Serialize()), nil
	}
	return hash.Hash{}, nil
}

// Name implements the interface rootobject.RootObject.
func (t TypeWrapper) Name() doltdb.TableName {
	if t.Type != nil {
		return doltdb.TableName{
			Name:   t.Type.ID.TypeName(),
			Schema: t.Type.ID.SchemaName(),
		}
	}
	return doltdb.TableName{}
}

// writeCache writes every type in the cache to the underlying map. This does not lock the collection, as it is assumed
// that the calling function already holds the lock.
func (pgs *TypeCollection) writeCache(ctx context.Context) (err error) {
	if len(pgs.accessedMap) == 0 {
		return nil
	}
	mapEditor := pgs.underlyingMap.Editor()
	for _, t := range pgs.accessedMap {
		data := t.Serialize()
		h, err := pgs.ns.WriteBytes(ctx, data)
		if err != nil {
			return err
		}
		if err = mapEditor.Update(ctx, string(t.ID), h); err != nil {
			return err
		}
	}
	pgs.underlyingMap, err = mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	clear(pgs.accessedMap)
	return nil
}
