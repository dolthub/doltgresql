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
)

// NewCompositeType creates new instance of composite DoltgresType.
func NewCompositeType(ctx *sql.Context, schema, name string, relId, arrayOid, typOid uint32, attrs []CompositeAttribute) *DoltgresType {
	return &DoltgresType{
		OID:            typOid,
		Name:           name,
		Schema:         schema,
		TypLength:      -1,
		PassedByVal:    false,
		TypType:        TypeType_Composite,
		TypCategory:    TypeCategory_CompositeTypes,
		IsPreferred:    false,
		IsDefined:      true,
		Delimiter:      ",",
		RelID:          relId,
		SubscriptFunc:  toFuncID("-"),
		Elem:           0,
		Array:          arrayOid,
		InputFunc:      toFuncID("record_in", oid.T_cstring, oid.T_oid, oid.T_int4),
		OutputFunc:     toFuncID("record_out", oid.T_record),
		ReceiveFunc:    toFuncID("record_recv", oid.T_internal, oid.T_oid, oid.T_int4),
		SendFunc:       toFuncID("record_send", oid.T_record),
		ModInFunc:      toFuncID("-"),
		ModOutFunc:     toFuncID("-"),
		AnalyzeFunc:    toFuncID("-"),
		Align:          TypeAlignment_Double,
		Storage:        TypeStorage_Extended,
		NotNull:        false,
		BaseTypeOID:    0,
		TypMod:         -1,
		NDims:          0,
		TypCollation:   0,
		DefaulBin:      "",
		Default:        "",
		Acl:            nil,
		Checks:         nil,
		attTypMod:      -1,
		CompareFunc:    toFuncID("btrecordcmp", oid.T_record, oid.T_record),
		CompositeAttrs: attrs,
	}
}

// CompositeAttribute represents a composite type attribute.
// This is a partial pg_attribute row entry.
type CompositeAttribute struct {
	relOid    uint32 // OID of the relation it belongs to
	name      string
	typOid    uint32 // OID of DoltgresType
	num       int16  // number of the column in relation
	collation string
}

// NewCompositeAttribute creates new instance of composite type attribute.
func NewCompositeAttribute(ctx *sql.Context, relOid uint32, name string, typOid uint32, num int16, collation string) CompositeAttribute {
	return CompositeAttribute{
		relOid:    relOid,
		name:      name,
		typOid:    typOid,
		num:       num,
		collation: collation,
	}
}
