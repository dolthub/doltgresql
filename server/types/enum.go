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
	srcdErrors "gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/utils"

	"github.com/dolthub/doltgresql/core/id"
)

// ErrInvalidInputValueForEnum is returned when the input value does not match given enum type's labels.
var ErrInvalidInputValueForEnum = srcdErrors.NewKind(`invalid input value for enum %s: "%s"`)

// NewEnumType creates new instance of enum DoltgresType.
func NewEnumType(ctx *sql.Context, arrayID, typeID id.Type, labels map[string]EnumLabel) *DoltgresType {
	return &DoltgresType{
		ID:                  typeID,
		TypLength:           4,
		PassedByVal:         true,
		TypType:             TypeType_Enum,
		TypCategory:         TypeCategory_EnumTypes,
		IsPreferred:         false,
		IsDefined:           true,
		Delimiter:           ",",
		RelID:               id.Null,
		SubscriptFunc:       toFuncID("-"),
		Elem:                id.NullType,
		Array:               arrayID,
		InputFunc:           toFuncID("enum_in", toInternal("cstring"), toInternal("oid")),
		OutputFunc:          toFuncID("enum_out", toInternal("anyenum")),
		ReceiveFunc:         toFuncID("enum_recv", toInternal("internal"), toInternal("oid")),
		SendFunc:            toFuncID("enum_send", toInternal("anyenum")),
		ModInFunc:           toFuncID("-"),
		ModOutFunc:          toFuncID("-"),
		AnalyzeFunc:         toFuncID("-"),
		Align:               TypeAlignment_Int,
		Storage:             TypeStorage_Plain,
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
		CompareFunc:         toFuncID("enum_cmp", toInternal("anyenum"), toInternal("anyenum")),
		EnumLabels:          labels,
		SerializationFunc:   serializeTypeEnum,
		DeserializationFunc: deserializeTypeEnum,
	}
}

// EnumLabel represents an enum type label.
// This is a pg_enum row entry.
type EnumLabel struct {
	ID        id.EnumLabel
	SortOrder float32
}

// NewEnumLabel creates new instance of enum type label.
func NewEnumLabel(ctx *sql.Context, labelID id.EnumLabel, so float32) EnumLabel {
	return EnumLabel{
		ID:        labelID,
		SortOrder: so,
	}
}

// serializeTypeEnum handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeEnum(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	str := val.(string)
	writer := utils.NewWriter(uint64(len(str) + 4))
	writer.String(str)
	return writer.Data(), nil
}

// deserializeTypeEnum handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeEnum(ctx *sql.Context, typ *DoltgresType, data []byte) (any, error) {
	// TODO: should return the index instead of label?
	if len(data) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(data)
	value := reader.String()
	if ctx == nil {
		// TODO: currently, in some places we use nil context, should fix it.
		return value, nil
	}
	if typ.TypCategory != TypeCategory_EnumTypes {
		return nil, errors.Errorf(`"%s" is not an enum type`, typ.Name())
	}
	if _, exists := typ.EnumLabels[value]; !exists {
		return nil, ErrInvalidInputValueForEnum.New(typ.Name(), value)
	}
	return value, nil
}
