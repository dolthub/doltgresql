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
	"cmp"
	"fmt"
	"maps"
	"slices"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/utils"
)

// init sets the serialization and deserialization functions.
func init() {
	types.SetExtendedTypeSerializers(SerializeType, DeserializeType)
}

// SerializeType is able to serialize the given extended type into a byte slice. All extended types will be defined
// by DoltgreSQL.
func SerializeType(extendedType types.ExtendedType) ([]byte, error) {
	if doltgresType, ok := extendedType.(*DoltgresType); ok {
		return doltgresType.Serialize(), nil
	}
	return nil, fmt.Errorf("unknown type to serialize")
}

// DeserializeType is able to deserialize the given serialized type into an appropriate extended type. All extended
// types will be defined by DoltgreSQL.
func DeserializeType(serializedType []byte) (types.ExtendedType, error) {
	if len(serializedType) == 0 {
		return nil, fmt.Errorf("deserializing empty type data")
	}

	typ := &DoltgresType{}
	reader := utils.NewReader(serializedType)
	version := reader.VariableUint()
	if version != 0 {
		return nil, fmt.Errorf("version %d of types is not supported, please upgrade the server", version)
	}

	typ.ID = reader.Internal()
	typ.TypLength = reader.Int16()
	typ.PassedByVal = reader.Bool()
	typ.TypType = TypeType(reader.String())
	typ.TypCategory = TypeCategory(reader.String())
	typ.IsPreferred = reader.Bool()
	typ.IsDefined = reader.Bool()
	typ.Delimiter = reader.String()
	typ.RelID = reader.Internal()
	typ.SubscriptFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	typ.Elem = reader.Internal()
	typ.Array = reader.Internal()
	typ.InputFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	typ.OutputFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	typ.ReceiveFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	typ.SendFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	typ.ModInFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	typ.ModOutFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	typ.AnalyzeFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	typ.Align = TypeAlignment(reader.String())
	typ.Storage = TypeStorage(reader.String())
	typ.NotNull = reader.Bool()
	typ.BaseTypeID = reader.Internal()
	typ.TypMod = reader.Int32()
	typ.NDims = reader.Int32()
	typ.TypCollation = reader.Internal()
	typ.DefaulBin = reader.String()
	typ.Default = reader.String()
	numOfAcl := reader.VariableUint()
	for k := uint64(0); k < numOfAcl; k++ {
		ac := reader.String()
		typ.Acl = append(typ.Acl, ac)
	}
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
	typ.attTypMod = reader.Int32()
	typ.CompareFunc = globalFunctionRegistry.InternalToRegistryID(reader.Internal())
	numOfEnumLabels := reader.VariableUint()
	if numOfEnumLabels > 0 {
		typ.EnumLabels = make(map[string]EnumLabel)
		for k := uint64(0); k < numOfEnumLabels; k++ {
			typeID := reader.Internal()
			sortOrder := reader.Float32()
			typ.EnumLabels[typeID.Segment(1)] = EnumLabel{
				ID:        typeID,
				SortOrder: sortOrder,
			}
		}
	}
	numOfCompAttrs := reader.VariableUint()
	if numOfCompAttrs > 0 {
		typ.CompositeAttrs = make([]CompositeAttribute, numOfCompAttrs)
		for k := uint64(0); k < numOfCompAttrs; k++ {
			relID := reader.Internal()
			name := reader.String()
			typeID := reader.Internal()
			num := reader.Int16()
			collation := reader.String()
			typ.CompositeAttrs[k] = CompositeAttribute{
				relID:     relID,
				name:      name,
				typeID:    typeID,
				num:       num,
				collation: collation,
			}
		}
	}
	typ.InternalName = reader.String()
	if !reader.IsEmpty() {
		return nil, fmt.Errorf("extra data found while deserializing type %s", typ.Name())
	}

	// Return the deserialized object
	return typ, nil
}

// Serialize returns the DoltgresType as a byte slice.
func (t *DoltgresType) Serialize() []byte {
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the type to the writer
	writer.Internal(t.ID)
	writer.Int16(t.TypLength)
	writer.Bool(t.PassedByVal)
	writer.String(string(t.TypType))
	writer.String(string(t.TypCategory))
	writer.Bool(t.IsPreferred)
	writer.Bool(t.IsDefined)
	writer.String(t.Delimiter)
	writer.Internal(t.RelID)
	writer.Internal(globalFunctionRegistry.GetInternalID(t.SubscriptFunc))
	writer.Internal(t.Elem)
	writer.Internal(t.Array)
	writer.Internal(globalFunctionRegistry.GetInternalID(t.InputFunc))
	writer.Internal(globalFunctionRegistry.GetInternalID(t.OutputFunc))
	writer.Internal(globalFunctionRegistry.GetInternalID(t.ReceiveFunc))
	writer.Internal(globalFunctionRegistry.GetInternalID(t.SendFunc))
	writer.Internal(globalFunctionRegistry.GetInternalID(t.ModInFunc))
	writer.Internal(globalFunctionRegistry.GetInternalID(t.ModOutFunc))
	writer.Internal(globalFunctionRegistry.GetInternalID(t.AnalyzeFunc))
	writer.String(string(t.Align))
	writer.String(string(t.Storage))
	writer.Bool(t.NotNull)
	writer.Internal(t.BaseTypeID)
	writer.Int32(t.TypMod)
	writer.Int32(t.NDims)
	writer.Internal(t.TypCollation)
	writer.String(t.DefaulBin)
	writer.String(t.Default)
	writer.VariableUint(uint64(len(t.Acl)))
	for _, ac := range t.Acl {
		writer.String(ac)
	}
	writer.VariableUint(uint64(len(t.Checks)))
	for _, check := range t.Checks {
		writer.String(check.Name)
		writer.String(check.CheckExpression)
	}
	writer.Int32(t.attTypMod)
	writer.Internal(globalFunctionRegistry.GetInternalID(t.CompareFunc))
	writer.VariableUint(uint64(len(t.EnumLabels)))
	if len(t.EnumLabels) > 0 {
		labels := slices.SortedFunc(maps.Values(t.EnumLabels), func(v1 EnumLabel, v2 EnumLabel) int {
			return cmp.Compare(v1.ID, v2.ID)
		})
		for _, l := range labels {
			writer.Internal(l.ID)
			writer.Float32(l.SortOrder)
		}
	}
	writer.VariableUint(uint64(len(t.CompositeAttrs)))
	if len(t.CompositeAttrs) > 0 {
		for _, l := range t.CompositeAttrs {
			writer.Internal(l.relID)
			writer.String(l.name)
			writer.Internal(l.typeID)
			writer.Int16(l.num)
			writer.String(l.collation)
		}
	}
	writer.String(t.InternalName)
	return writer.Data()
}
