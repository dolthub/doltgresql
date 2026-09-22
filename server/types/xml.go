// Copyright 2026 Dolthub, Inc.
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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Xml is the XML type.
var Xml = &DoltgresType{
	ID:                  toInternal("xml"),
	TypLength:           int16(-1),
	PassedByVal:         false,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_UserDefinedTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                internalNullType,
	Array:               internalNullType,
	InputFunc:           toFuncID("xml_in", toInternal("cstring")),
	OutputFunc:          toFuncID("xml_out", toInternal("xml")),
	ReceiveFunc:         toFuncID("xml_recv", toInternal("internal")),
	SendFunc:            toFuncID("xml_send", toInternal("xml")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Int,
	Storage:             TypeStorage_Extended,
	NotNull:             false,
	BaseTypeType:        internalNullType,
	TypMod:              -1,
	NDims:               0,
	TypCollation:        id.NullCollation,
	DefaulBin:           "",
	Default:             "",
	Acl:                 nil,
	Checks:              nil,
	attTypMod:           -1,
	CompareFunc:         toFuncID("-"),
	SerializationFunc:   serializeTypeXml,
	DeserializationFunc: deserializeTypeXml,
}

// serializeTypeXml handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeXml(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	str, err := unwrapSerializationString(ctx, "xml", val)
	if err != nil {
		return nil, err
	}
	writer := utils.NewWriter(uint64(len(str) + 4))
	writer.String(str)
	return writer.Data(), nil
}

// deserializeTypeXml handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeXml(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(data)
	return reader.String(), nil
}
