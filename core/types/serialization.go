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

package types

import (
	"context"
	"fmt"
	"sync"

	"github.com/dolthub/go-mysql-server/sql"

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

	// Write all the Types to the writer
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
			Type := nameMap[nameMapKey]
			writer.String(Type.Name)
			writer.String(Type.Owner)
			writer.Int16(Type.Length)
			writer.Bool(Type.PassedByVal)
			writer.String(string(Type.Typ))
			writer.String(string(Type.Category))
			writer.Bool(Type.IsPreferred)
			writer.Bool(Type.IsDefined)
			writer.String(Type.Delimiter)
			writer.Uint32(Type.RelID)
			writer.String(Type.Subscript)
			writer.Uint32(Type.Elem)
			writer.Uint32(Type.Array)
			writer.String(Type.Input)
			writer.String(Type.Output)
			writer.String(Type.Receive)
			writer.String(Type.Send)
			writer.String(Type.ModIn)
			writer.String(Type.ModOut)
			writer.String(Type.Analyze)
			writer.String(string(Type.Align))
			writer.String(string(Type.Storage))
			writer.Bool(Type.NotNull)
			writer.Uint32(Type.BaseTypeOID)
			writer.Int32(Type.TypMod)
			writer.Int32(Type.NDims)
			writer.Uint32(Type.Collation)
			writer.String(Type.DefaulBin)
			writer.String(Type.Default)
			writer.String(Type.Acl)
			writer.VariableUint(uint64(len(Type.Checks)))
			for _, check := range Type.Checks {
				writer.String(check.Name)
				writer.String(check.CheckExpression)
			}
		}
	}

	return writer.Data(), nil
}

// Deserialize returns the Collection that was serialized in the byte slice.
// Returns an empty Collection if data is nil or empty.
func Deserialize(ctx context.Context, data []byte) (*TypeCollection, error) {
	if len(data) == 0 {
		return &TypeCollection{
			schemaMap: make(map[string]map[string]*Type),
			mutex:     &sync.Mutex{},
		}, nil
	}
	schemaMap := make(map[string]map[string]*Type)
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return nil, fmt.Errorf("version %d of Types is not supported, please upgrade the server", version)
	}

	// Read from the reader
	numOfSchemas := reader.VariableUint()
	for i := uint64(0); i < numOfSchemas; i++ {
		schemaName := reader.String()
		numOfTypes := reader.VariableUint()
		nameMap := make(map[string]*Type)
		for j := uint64(0); j < numOfTypes; j++ {
			Type := &Type{}
			Type.Name = reader.String()
			Type.Owner = reader.String()
			Type.Length = reader.Int16()
			Type.PassedByVal = reader.Bool()
			Type.Typ = types.TypeType(reader.String())
			Type.Category = types.TypeCategory(reader.String())
			Type.IsPreferred = reader.Bool()
			Type.IsDefined = reader.Bool()
			Type.Delimiter = reader.String()
			Type.RelID = reader.Uint32()
			Type.Subscript = reader.String()
			Type.Elem = reader.Uint32()
			Type.Array = reader.Uint32()
			Type.Input = reader.String()
			Type.Output = reader.String()
			Type.Receive = reader.String()
			Type.Send = reader.String()
			Type.ModIn = reader.String()
			Type.ModOut = reader.String()
			Type.Analyze = reader.String()
			Type.Align = types.TypeAlignment(reader.String())
			Type.Storage = types.TypeStorage(reader.String())
			Type.NotNull = reader.Bool()
			Type.BaseTypeOID = reader.Uint32()
			Type.TypMod = reader.Int32()
			Type.NDims = reader.Int32()
			Type.Collation = reader.Uint32()
			Type.DefaulBin = reader.String()
			Type.Default = reader.String()
			Type.Acl = reader.String()
			numOfChecks := reader.VariableUint()
			for k := uint64(0); k < numOfChecks; k++ {
				checkName := reader.String()
				checkExpr := reader.String()
				Type.Checks = append(Type.Checks, &sql.CheckDefinition{
					Name:            checkName,
					CheckExpression: checkExpr,
					Enforced:        true,
				})
			}
			nameMap[Type.Name] = Type
		}
		schemaMap[schemaName] = nameMap
	}
	if !reader.IsEmpty() {
		return nil, fmt.Errorf("extra data found while deserializing Types")
	}

	// Return the deserialized object
	return &TypeCollection{
		schemaMap: schemaMap,
		mutex:     &sync.Mutex{},
	}, nil
}
