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
	"github.com/lib/pq/oid"
	"gopkg.in/src-d/go-errors.v1"
)

// ErrInvalidInputValueForEnum is returned when the input value does not match given enum type's labels.
var ErrInvalidInputValueForEnum = errors.NewKind(`invalid input value for enum %s: "%s"`)

// NewEnumType creates new instance of enum DoltgresType.
func NewEnumType(ctx *sql.Context, schema, name, owner string, arrayOid, typOid uint32, labels map[string]EnumLabel) *DoltgresType {
	return &DoltgresType{
		OID:           typOid,
		Name:          name,
		Schema:        schema,
		Owner:         owner,
		TypLength:     4,
		PassedByVal:   true,
		TypType:       TypeType_Enum,
		TypCategory:   TypeCategory_EnumTypes,
		IsPreferred:   false,
		IsDefined:     true,
		Delimiter:     ",",
		RelID:         0,
		SubscriptFunc: toFuncID("-"),
		Elem:          0,
		Array:         arrayOid,
		InputFunc:     toFuncID("enum_in", oid.T_cstring, oid.T_oid),
		OutputFunc:    toFuncID("enum_out", oid.T_anyenum),
		ReceiveFunc:   toFuncID("enum_recv", oid.T_internal, oid.T_oid),
		SendFunc:      toFuncID("enum_send", oid.T_anyenum),
		ModInFunc:     toFuncID("-"),
		ModOutFunc:    toFuncID("-"),
		AnalyzeFunc:   toFuncID("-"),
		Align:         TypeAlignment_Int,
		Storage:       TypeStorage_Plain,
		NotNull:       false,
		BaseTypeOID:   0,
		TypMod:        -1,
		NDims:         0,
		TypCollation:  0,
		DefaulBin:     "",
		Default:       "",
		Acl:           nil,
		Checks:        nil,
		attTypMod:     -1,
		CompareFunc:   toFuncID("enum_cmp", oid.T_anyenum, oid.T_anyenum),
		EnumLabels:    labels,
	}
}

// EnumLabel represents an enum type label.
// This is a pg_enum row entry.
type EnumLabel struct {
	OID        uint32 // unique OID for each label
	EnumTypOid uint32 // OID of DoltgresType
	SortOrder  float32
	Label      string
}

// NewEnumLabel creates new instance of enum type label.
func NewEnumLabel(ctx *sql.Context, oid, enumTypeOid uint32, so float32, label string) EnumLabel {
	return EnumLabel{
		OID:        oid,
		EnumTypOid: enumTypeOid,
		SortOrder:  so,
		Label:      label,
	}
}
