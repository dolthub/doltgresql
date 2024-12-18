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
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/core/id"
)

// ErrInvalidInputValueForEnum is returned when the input value does not match given enum type's labels.
var ErrInvalidInputValueForEnum = errors.NewKind(`invalid input value for enum %s: "%s"`)

// NewEnumType creates new instance of enum DoltgresType.
func NewEnumType(ctx *sql.Context, arrayID, typeID id.Internal, labels map[string]EnumLabel) *DoltgresType {
	return &DoltgresType{
		ID:            typeID,
		TypLength:     4,
		PassedByVal:   true,
		TypType:       TypeType_Enum,
		TypCategory:   TypeCategory_EnumTypes,
		IsPreferred:   false,
		IsDefined:     true,
		Delimiter:     ",",
		RelID:         id.Null,
		SubscriptFunc: toFuncID("-"),
		Elem:          id.Null,
		Array:         arrayID,
		InputFunc:     toFuncID("enum_in", toInternal("cstring"), toInternal("oid")),
		OutputFunc:    toFuncID("enum_out", toInternal("anyenum")),
		ReceiveFunc:   toFuncID("enum_recv", toInternal("internal"), toInternal("oid")),
		SendFunc:      toFuncID("enum_send", toInternal("anyenum")),
		ModInFunc:     toFuncID("-"),
		ModOutFunc:    toFuncID("-"),
		AnalyzeFunc:   toFuncID("-"),
		Align:         TypeAlignment_Int,
		Storage:       TypeStorage_Plain,
		NotNull:       false,
		BaseTypeID:    id.Null,
		TypMod:        -1,
		NDims:         0,
		TypCollation:  id.Null,
		DefaulBin:     "",
		Default:       "",
		Acl:           nil,
		Checks:        nil,
		attTypMod:     -1,
		CompareFunc:   toFuncID("enum_cmp", toInternal("anyenum"), toInternal("anyenum")),
		EnumLabels:    labels,
	}
}

// EnumLabel represents an enum type label.
// This is a pg_enum row entry.
type EnumLabel struct {
	ID        id.Internal // First segment is the ENUM parent's Internal ID, second segment is the label
	SortOrder float32
}

// NewEnumLabel creates new instance of enum type label.
func NewEnumLabel(ctx *sql.Context, typeID id.Internal, so float32) EnumLabel {
	return EnumLabel{
		ID:        typeID,
		SortOrder: so,
	}
}
