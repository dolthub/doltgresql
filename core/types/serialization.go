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
			typ := nameMap[nameMapKey]
			writer.String(typ.Name)
			writer.String(typ.Owner)
			writer.Int16(typ.Length)
			writer.Bool(typ.PassedByVal)
			writer.String(string(typ.Typ))
			writer.String(string(typ.Category))
			writer.Bool(typ.IsPreferred)
			writer.Bool(typ.IsDefined)
			writer.String(typ.Delimiter)
			writer.Uint32(typ.RelID)
			writer.String(typ.Subscript)
			writer.Uint32(typ.Elem)
			writer.Uint32(typ.Array)
			writer.String(typ.Input)
			writer.String(typ.Output)
			writer.String(typ.Receive)
			writer.String(typ.Send)
			writer.String(typ.ModIn)
			writer.String(typ.ModOut)
			writer.String(typ.Analyze)
			writer.String(string(typ.Align))
			writer.String(string(typ.Storage))
			writer.Bool(typ.NotNull)
			writer.Uint32(typ.BaseTypeOID)
			writer.Int32(typ.TypMod)
			writer.Int32(typ.NDims)
			writer.Uint32(typ.Collation)
			writer.String(typ.DefaulBin)
			writer.String(typ.Default)
			writer.String(typ.Acl)
			writer.VariableUint(uint64(len(typ.Checks)))
			for _, check := range typ.Checks {
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
			mutex:     &sync.RWMutex{},
		}, nil
	}
	schemaMap := make(map[string]map[string]*Type)
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
		nameMap := make(map[string]*Type)
		for j := uint64(0); j < numOfTypes; j++ {
			typ := &Type{}
			typ.Name = reader.String()
			typ.Owner = reader.String()
			typ.Length = reader.Int16()
			typ.PassedByVal = reader.Bool()
			typ.Typ = types.TypeType(reader.String())
			typ.Category = types.TypeCategory(reader.String())
			typ.IsPreferred = reader.Bool()
			typ.IsDefined = reader.Bool()
			typ.Delimiter = reader.String()
			typ.RelID = reader.Uint32()
			typ.Subscript = reader.String()
			typ.Elem = reader.Uint32()
			typ.Array = reader.Uint32()
			typ.Input = reader.String()
			typ.Output = reader.String()
			typ.Receive = reader.String()
			typ.Send = reader.String()
			typ.ModIn = reader.String()
			typ.ModOut = reader.String()
			typ.Analyze = reader.String()
			typ.Align = types.TypeAlignment(reader.String())
			typ.Storage = types.TypeStorage(reader.String())
			typ.NotNull = reader.Bool()
			typ.BaseTypeOID = reader.Uint32()
			typ.TypMod = reader.Int32()
			typ.NDims = reader.Int32()
			typ.Collation = reader.Uint32()
			typ.DefaulBin = reader.String()
			typ.Default = reader.String()
			typ.Acl = reader.String()
			numOfChecks := reader.VariableUint()
			for k := uint64(0); k < numOfChecks; k++ {
				checkName := reader.String()
				checkExpr := reader.String()
				typ.Checks = append(typ.Checks, &sql.CheckDefinition{
					Name:            checkName,
					CheckExpression: checkExpr,
					Enforced:        true,
				})
			}
			nameMap[typ.Name] = typ
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
