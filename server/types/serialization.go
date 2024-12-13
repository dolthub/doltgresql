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
	"fmt"

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

	typ.OID = reader.Uint32()
	typ.Name = reader.String()
	typ.Schema = reader.String()
	typ.TypLength = reader.Int16()
	typ.PassedByVal = reader.Bool()
	typ.TypType = TypeType(reader.String())
	typ.TypCategory = TypeCategory(reader.String())
	typ.IsPreferred = reader.Bool()
	typ.IsDefined = reader.Bool()
	typ.Delimiter = reader.String()
	typ.RelID = reader.Uint32()
	typ.SubscriptFunc = globalFunctionRegistry.StringToID(reader.String())
	typ.Elem = reader.Uint32()
	typ.Array = reader.Uint32()
	typ.InputFunc = globalFunctionRegistry.StringToID(reader.String())
	typ.OutputFunc = globalFunctionRegistry.StringToID(reader.String())
	typ.ReceiveFunc = globalFunctionRegistry.StringToID(reader.String())
	typ.SendFunc = globalFunctionRegistry.StringToID(reader.String())
	typ.ModInFunc = globalFunctionRegistry.StringToID(reader.String())
	typ.ModOutFunc = globalFunctionRegistry.StringToID(reader.String())
	typ.AnalyzeFunc = globalFunctionRegistry.StringToID(reader.String())
	typ.Align = TypeAlignment(reader.String())
	typ.Storage = TypeStorage(reader.String())
	typ.NotNull = reader.Bool()
	typ.BaseTypeOID = reader.Uint32()
	typ.TypMod = reader.Int32()
	typ.NDims = reader.Int32()
	typ.TypCollation = reader.Uint32()
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
	typ.CompareFunc = globalFunctionRegistry.StringToID(reader.String())
	numOfEnumLabels := reader.VariableUint()
	if numOfEnumLabels > 0 {
		typ.EnumLabels = make(map[string]EnumLabel)
		for k := uint64(0); k < numOfEnumLabels; k++ {
			oid := reader.Uint32()
			enumOid := reader.Uint32()
			sortOrder := reader.Float32()
			label := reader.String()
			typ.EnumLabels[label] = EnumLabel{
				OID:        oid,
				EnumTypOid: enumOid,
				SortOrder:  sortOrder,
				Label:      label,
			}
		}
	}
	numOfCompAttrs := reader.VariableUint()
	if numOfCompAttrs > 0 {
		typ.CompositeAttrs = make([]CompositeAttribute, numOfCompAttrs)
		for k := uint64(0); k < numOfCompAttrs; k++ {
			relOid := reader.Uint32()
			name := reader.String()
			typOid := reader.Uint32()
			num := reader.Int16()
			collation := reader.String()
			typ.CompositeAttrs[k] = CompositeAttribute{
				relOid:    relOid,
				name:      name,
				typOid:    typOid,
				num:       num,
				collation: collation,
			}
		}
	}
	typ.InternalName = reader.String()
	if !reader.IsEmpty() {
		return nil, fmt.Errorf("extra data found while deserializing type %s", typ.Name)
	}

	// Return the deserialized object
	return typ, nil
}

// Serialize returns the DoltgresType as a byte slice.
func (t *DoltgresType) Serialize() []byte {
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the type to the writer
	writer.Uint32(t.OID)
	writer.String(t.Name)
	writer.String(t.Schema)
	writer.Int16(t.TypLength)
	writer.Bool(t.PassedByVal)
	writer.String(string(t.TypType))
	writer.String(string(t.TypCategory))
	writer.Bool(t.IsPreferred)
	writer.Bool(t.IsDefined)
	writer.String(t.Delimiter)
	writer.Uint32(t.RelID)
	writer.String(globalFunctionRegistry.GetFullString(t.SubscriptFunc))
	writer.Uint32(t.Elem)
	writer.Uint32(t.Array)
	writer.String(globalFunctionRegistry.GetFullString(t.InputFunc))
	writer.String(globalFunctionRegistry.GetFullString(t.OutputFunc))
	writer.String(globalFunctionRegistry.GetFullString(t.ReceiveFunc))
	writer.String(globalFunctionRegistry.GetFullString(t.SendFunc))
	writer.String(globalFunctionRegistry.GetFullString(t.ModInFunc))
	writer.String(globalFunctionRegistry.GetFullString(t.ModOutFunc))
	writer.String(globalFunctionRegistry.GetFullString(t.AnalyzeFunc))
	writer.String(string(t.Align))
	writer.String(string(t.Storage))
	writer.Bool(t.NotNull)
	writer.Uint32(t.BaseTypeOID)
	writer.Int32(t.TypMod)
	writer.Int32(t.NDims)
	writer.Uint32(t.TypCollation)
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
	writer.String(globalFunctionRegistry.GetFullString(t.CompareFunc))
	writer.VariableUint(uint64(len(t.EnumLabels)))
	if t.EnumLabels != nil {
		for _, l := range t.EnumLabels {
			writer.Uint32(l.OID)
			writer.Uint32(l.EnumTypOid)
			writer.Float32(l.SortOrder)
			writer.String(l.Label)
		}
	}
	writer.VariableUint(uint64(len(t.CompositeAttrs)))
	if t.CompositeAttrs != nil {
		for _, l := range t.CompositeAttrs {
			writer.Uint32(l.relOid)
			writer.String(l.name)
			writer.Uint32(l.typOid)
			writer.Int16(l.num)
			writer.String(l.collation)
		}
	}
	writer.String(t.InternalName)
	return writer.Data()
}
