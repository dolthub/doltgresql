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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// NewCompositeType creates new instance of composite DoltgresType.
func NewCompositeType(ctx *sql.Context, relID id.Internal, arrayID, typeID id.InternalType, attrs []CompositeAttribute) *DoltgresType {
	return &DoltgresType{
		ID:             typeID,
		TypLength:      -1,
		PassedByVal:    false,
		TypType:        TypeType_Composite,
		TypCategory:    TypeCategory_CompositeTypes,
		IsPreferred:    false,
		IsDefined:      true,
		Delimiter:      ",",
		RelID:          relID,
		SubscriptFunc:  toFuncID("-"),
		Elem:           id.NullType,
		Array:          arrayID,
		InputFunc:      toFuncID("record_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
		OutputFunc:     toFuncID("record_out", toInternal("record")),
		ReceiveFunc:    toFuncID("record_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
		SendFunc:       toFuncID("record_send", toInternal("record")),
		ModInFunc:      toFuncID("-"),
		ModOutFunc:     toFuncID("-"),
		AnalyzeFunc:    toFuncID("-"),
		Align:          TypeAlignment_Double,
		Storage:        TypeStorage_Extended,
		NotNull:        false,
		BaseTypeID:     id.NullType,
		TypMod:         -1,
		NDims:          0,
		TypCollation:   id.NullCollation,
		DefaulBin:      "",
		Default:        "",
		Acl:            nil,
		Checks:         nil,
		attTypMod:      -1,
		CompareFunc:    toFuncID("btrecordcmp", toInternal("record"), toInternal("record")),
		CompositeAttrs: attrs,
	}
}

// CompositeAttribute represents a composite type attribute.
// This is a partial pg_attribute row entry.
type CompositeAttribute struct {
	relID     id.Internal // ID of the relation it belongs to
	name      string
	typeID    id.InternalType // ID of DoltgresType
	num       int16           // number of the column in relation
	collation string
}

// NewCompositeAttribute creates new instance of composite type attribute.
func NewCompositeAttribute(ctx *sql.Context, relID id.Internal, name string, typeID id.InternalType, num int16, collation string) CompositeAttribute {
	return CompositeAttribute{
		relID:     relID,
		name:      name,
		typeID:    typeID,
		num:       num,
		collation: collation,
	}
}
