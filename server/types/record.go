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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/utils"

	"github.com/dolthub/doltgresql/core/id"
)

// Record is a generic, anonymous record type, without field type information supplied yet. When used with RecordExpr,
// the field type information will be created once the field expressions are analyzed and type information is available,
// and a new DoltgresType instance will be created with the field type information populated.
var Record = &DoltgresType{
	ID:                  toInternal("record"),
	TypLength:           -1,
	PassedByVal:         false,
	TypType:             TypeType_Pseudo,
	TypCategory:         TypeCategory_PseudoTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_record"),
	InputFunc:           toFuncID("record_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:          toFuncID("record_out", toInternal("record")),
	ReceiveFunc:         toFuncID("record_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:            toFuncID("record_send", toInternal("record")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Double,
	Storage:             TypeStorage_Extended,
	NotNull:             false,
	BaseTypeID:          id.NullType,
	TypMod:              -1,
	NDims:               0,
	TypCollation:        id.NullCollation,
	DefaulBin:           "",
	Default:             "",
	Acl:                 nil,
	Checks:              nil,
	attTypMod:           -1,
	CompareFunc:         toFuncID("-"),
	SerializationFunc:   serializeTypeRecord,
	DeserializationFunc: deserializeTypeRecord,
}

// RecordValue represents a single value in a record, along with its
// associated type.
type RecordValue struct {
	Value any
	Type  sql.Type
}

// serializeTypeRecord handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeRecord(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	values, ok := val.([]RecordValue)
	if !ok {
		return nil, errors.Errorf("expected []RecordValue, but got %T", val)
	}
	writer := utils.NewWriter(uint64(16 * len(values)))
	writer.Byte(0) // Version
	writer.VariableUint(uint64(len(values)))
	for _, value := range values {
		dgtype, ok := value.Type.(*DoltgresType)
		if !ok {
			return nil, errors.Errorf("record_send only supports Doltgres types, but received `%T`", value.Type)
		}
		valBytes, err := dgtype.SerializeValue(ctx, value.Value)
		if err != nil {
			return nil, err
		}
		writer.Id(dgtype.ID.AsId())
		writer.ByteSlice(valBytes)
	}
	return writer.Data(), nil
}

// deserializeTypeRecord handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeRecord(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	typeColl, err := GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	reader := utils.NewReader(data)
	version := reader.Byte()
	switch version {
	case 0:
		valuesLen := reader.VariableUint()
		values := make([]RecordValue, valuesLen)
		for i := uint64(0); i < valuesLen; i++ {
			typeId := id.Type(reader.Id())
			valueData := reader.ByteSlice()
			dgtype, err := typeColl.GetType(ctx, typeId)
			if err != nil {
				return nil, err
			}
			if dgtype == nil {
				return nil, errors.Errorf("record_recv encountered type `%s.%s` which could not be found",
					typeId.SchemaName(), typeId.TypeName())
			}
			value, err := dgtype.DeserializeValue(ctx, valueData)
			if err != nil {
				return nil, err
			}
			values[i] = RecordValue{
				Value: value,
				Type:  dgtype,
			}
		}
		if reader.RemainingBytes() > 0 {
			return nil, errors.New("record_recv encountered extra data during deserialization")
		}
		return values, nil
	default:
		return nil, errors.Errorf("version %d of record serialization is not supported, please upgrade the server", version)
	}
}
