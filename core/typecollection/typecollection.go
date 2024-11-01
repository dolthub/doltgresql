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
	"sort"
	"sync"

	"github.com/dolthub/doltgresql/server/types"
)

// TypeCollection contains a collection of types.
type TypeCollection struct {
	schemaMap map[string]map[string]types.DoltgresType
	mutex     *sync.RWMutex
}

// GetType returns the type with the given schema and name.
// Returns nil if the type cannot be found.
func (pgs *TypeCollection) GetType(schName, typName string) (types.DoltgresType, bool) {
	pgs.mutex.RLock()
	defer pgs.mutex.RUnlock()

	if nameMap, ok := pgs.schemaMap[schName]; ok {
		if typ, ok := nameMap[typName]; ok {
			return typ, true
		}
	}
	return types.DoltgresType{}, false
}

// GetDomainType returns a domain type with the given schema and name.
// Returns nil if the type cannot be found. It checks for domain type.
func (pgs *TypeCollection) GetDomainType(schName, typName string) (types.DoltgresType, bool) {
	pgs.mutex.RLock()
	defer pgs.mutex.RUnlock()

	if nameMap, ok := pgs.schemaMap[schName]; ok {
		if typ, ok := nameMap[typName]; ok && typ.TypType == types.TypeType_Domain {
			return typ, true
		}
	}
	return types.DoltgresType{}, false
}

// GetAllTypes returns a map containing all types in the collection, grouped by the schema they're contained in.
// Each type array is also sorted by the type name.
func (pgs *TypeCollection) GetAllTypes() (typesMap map[string][]types.DoltgresType, schemaNames []string, totalCount int) {
	pgs.mutex.RLock()
	defer pgs.mutex.RUnlock()

	typesMap = make(map[string][]types.DoltgresType)
	for schemaName, nameMap := range pgs.schemaMap {
		schemaNames = append(schemaNames, schemaName)
		typs := make([]types.DoltgresType, 0, len(nameMap))
		for _, typ := range nameMap {
			typs = append(typs, typ)
		}
		totalCount += len(typs)
		sort.Slice(typs, func(i, j int) bool {
			return typs[i].Name < typs[j].Name
		})
		typesMap[schemaName] = typs
	}

	// TODO: add built-in types
	//builtInDoltgresTypes := types.GetAllTypes()
	//for _, dt := range builtInDoltgresTypes {
	//
	//}
	sort.Slice(schemaNames, func(i, j int) bool {
		return schemaNames[i] < schemaNames[j]
	})
	return
}

// CreateType creates a new type.
func (pgs *TypeCollection) CreateType(schema string, typ types.DoltgresType) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	nameMap, ok := pgs.schemaMap[schema]
	if !ok {
		nameMap = make(map[string]types.DoltgresType)
		pgs.schemaMap[schema] = nameMap
	}
	if _, ok = nameMap[typ.Name]; ok {
		return types.ErrTypeAlreadyExists.New(typ.Name)
	}
	nameMap[typ.Name] = typ
	return nil
}

// DropType drops an existing type.
func (pgs *TypeCollection) DropType(schName, typName string) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if nameMap, ok := pgs.schemaMap[schName]; ok {
		if _, ok = nameMap[typName]; ok {
			delete(nameMap, typName)
			return nil
		}
	}
	return types.ErrTypeDoesNotExist.New(typName)
}

// IterateTypes iterates over all types in the collection.
func (pgs *TypeCollection) IterateTypes(f func(schema string, typ types.DoltgresType) error) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	for schema, nameMap := range pgs.schemaMap {
		for _, t := range nameMap {
			if err := f(schema, t); err != nil {
				return err
			}
		}
	}
	return nil
}

// Clone returns a new *TypeCollection with the same contents as the original.
func (pgs *TypeCollection) Clone() *TypeCollection {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	newCollection := &TypeCollection{
		schemaMap: make(map[string]map[string]types.DoltgresType),
		mutex:     &sync.RWMutex{},
	}
	for schema, nameMap := range pgs.schemaMap {
		if len(nameMap) == 0 {
			continue
		}
		clonedNameMap := make(map[string]types.DoltgresType)
		for key, typ := range nameMap {
			clonedNameMap[key] = typ
		}
		newCollection.schemaMap[schema] = clonedNameMap
	}
	return newCollection
}
