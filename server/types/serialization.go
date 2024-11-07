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
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/utils"
)

// init sets the serialization and deserialization functions.
func init() {
	types.SetExtendedTypeSerializers(SerializeType, DeserializeType)
}

// SerializeType is able to serialize the given extended type into a byte slice. All extended types will be defined
// by DoltgreSQL.
func SerializeType(extendedType types.ExtendedType) ([]byte, error) {
	if doltgresType, ok := extendedType.(DoltgresType); ok {
		return doltgresType.Serialize(), nil
	}
	return nil, fmt.Errorf("unknown type to serialize")
}

// DeserializeType is able to deserialize the given serialized type into an appropriate extended type. All extended
// types will be defined by DoltgreSQL.
func DeserializeType(serializedType []byte) (types.ExtendedType, error) {
	return Deserialize(serializedType)
}

// Serialize returns the DoltgresType as a byte slice.
func (t DoltgresType) Serialize() []byte {
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the type to the writer
	writer.Uint32(t.OID)
	writer.String(t.Name)
	writer.String(t.Schema)
	writer.String(t.Owner)
	writer.Int16(t.TypLength)
	writer.Bool(t.PassedByVal)
	writer.String(string(t.TypType))
	writer.String(string(t.TypCategory))
	writer.Bool(t.IsPreferred)
	writer.Bool(t.IsDefined)
	writer.String(t.Delimiter)
	writer.Uint32(t.RelID)
	writer.String(t.SubscriptFunc)
	writer.Uint32(t.Elem)
	writer.Uint32(t.Array)
	writer.String(t.InputFunc)
	writer.String(t.OutputFunc)
	writer.String(t.ReceiveFunc)
	writer.String(t.SendFunc)
	writer.String(t.ModInFunc)
	writer.String(t.ModOutFunc)
	writer.String(t.AnalyzeFunc)
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
	writer.Int32(t.AttTypMod)
	// TODO: get rid this?
	writer.String(t.internalName)
	return writer.Data()
}

// Deserialize returns the Collection that was serialized in the byte slice.
// Returns an empty Collection if data is nil or empty.
func Deserialize(data []byte) (DoltgresType, error) {
	if len(data) == 0 {
		return DoltgresType{}, fmt.Errorf("deserializing empty type data")
	}

	typ := DoltgresType{}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return DoltgresType{}, fmt.Errorf("version %d of types is not supported, please upgrade the server", version)
	}

	typ.OID = reader.Uint32()
	typ.Name = reader.String()
	typ.Schema = reader.String()
	typ.Owner = reader.String()
	typ.TypLength = reader.Int16()
	typ.PassedByVal = reader.Bool()
	typ.TypType = TypeType(reader.String())
	typ.TypCategory = TypeCategory(reader.String())
	typ.IsPreferred = reader.Bool()
	typ.IsDefined = reader.Bool()
	typ.Delimiter = reader.String()
	typ.RelID = reader.Uint32()
	typ.SubscriptFunc = reader.String()
	typ.Elem = reader.Uint32()
	typ.Array = reader.Uint32()
	typ.InputFunc = reader.String()
	typ.OutputFunc = reader.String()
	typ.ReceiveFunc = reader.String()
	typ.SendFunc = reader.String()
	typ.ModInFunc = reader.String()
	typ.ModOutFunc = reader.String()
	typ.AnalyzeFunc = reader.String()
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
	typ.AttTypMod = reader.Int32()
	// TODO: get rid this?
	typ.internalName = reader.String()
	if !reader.IsEmpty() {
		return DoltgresType{}, fmt.Errorf("extra data found while deserializing type %s", typ.Name)
	}

	// Return the deserialized object
	return typ, nil
}
