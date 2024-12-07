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
	"sync"

	"github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the TypeCollection as a byte slice.
// If the TypeCollection is nil, then this returns a nil slice.
func (pgs *TypeCollection) Serialize(ctx context.Context) ([]byte, error) {
	if pgs == nil {
		return nil, nil
	}
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// Write all the types to the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	schemaMapKeys := utils.GetMapKeysSorted(pgs.schemaMap)
	writer.VariableUint(uint64(len(schemaMapKeys)))
	for _, schemaMapKey := range schemaMapKeys {
		nameMap := pgs.schemaMap[schemaMapKey]
		writer.String(schemaMapKey)
		nameMapKeys := utils.GetMapKeysSorted(nameMap)
		writer.VariableUint(uint64(len(nameMapKeys)))
		for _, nameMapKey := range nameMapKeys {
			typ := nameMap[nameMapKey]
			data := typ.Serialize()
			writer.ByteSlice(data)
		}
	}

	return writer.Data(), nil
}

// Deserialize returns the Collection that was serialized in the byte slice.
// Returns an empty Collection if data is nil or empty.
func Deserialize(ctx context.Context, data []byte) (*TypeCollection, error) {
	if len(data) == 0 {
		return &TypeCollection{
			schemaMap: make(map[string]map[string]*types.DoltgresType),
			mutex:     &sync.RWMutex{},
		}, nil
	}
	schemaMap := make(map[string]map[string]*types.DoltgresType)
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return nil, fmt.Errorf("version %d of types is not supported, please upgrade the server", version)
	}

	// Read from the reader
	numOfSchemas := reader.VariableUint()
	for i := uint64(0); i < numOfSchemas; i++ {
		schemaName := reader.String()
		numOfTypes := reader.VariableUint()
		nameMap := make(map[string]*types.DoltgresType)
		for j := uint64(0); j < numOfTypes; j++ {
			typData := reader.ByteSlice()
			typ, err := types.DeserializeType(typData)
			if err != nil {
				return nil, err
			}
			dt := typ.(*types.DoltgresType)
			nameMap[dt.Name] = dt
		}
		schemaMap[schemaName] = nameMap
	}
	if !reader.IsEmpty() {
		return nil, fmt.Errorf("extra data found while deserializing types")
	}

	// Return the deserialized object
	return &TypeCollection{
		schemaMap: schemaMap,
		mutex:     &sync.RWMutex{},
	}, nil
}
