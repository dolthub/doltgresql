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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Text is the text type.
var Text = &DoltgresType{
	ID:                  toInternal("text"),
	TypLength:           int16(-1),
	PassedByVal:         false,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_StringTypes,
	IsPreferred:         true,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_text"),
	InputFunc:           toFuncID("textin", toInternal("cstring")),
	OutputFunc:          toFuncID("textout", toInternal("text")),
	ReceiveFunc:         toFuncID("textrecv", toInternal("internal")),
	SendFunc:            toFuncID("textsend", toInternal("text")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Int,
	Storage:             TypeStorage_Extended,
	NotNull:             false,
	BaseTypeID:          id.NullType,
	TypMod:              -1,
	NDims:               0,
	TypCollation:        id.NewCollation("pg_catalog", "default"),
	DefaulBin:           "",
	Default:             "",
	Acl:                 nil,
	Checks:              nil,
	attTypMod:           -1,
	CompareFunc:         toFuncID("bttextcmp", toInternal("text"), toInternal("text")),
	SerializationFunc:   serializeTypeText,
	DeserializationFunc: deserializeTypeText,
}

// serializeTypeText handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeText(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	str, ok, err := sql.Unwrap[string](ctx, val)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Errorf(`"text" serialization requires a string argument, got %T`, val)
	}
	writer := utils.NewWriter(uint64(len(str) + 4))
	writer.String(str)
	return writer.Data(), nil
}

// deserializeTypeText handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeText(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(data)
	return reader.String(), nil
}
