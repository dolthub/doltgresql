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

	"github.com/lib/pq/oid"
)

// CreateArrayTypeFromBaseType create array type from given type.
func CreateArrayTypeFromBaseType(baseType *DoltgresType) *DoltgresType {
	align := TypeAlignment_Int
	if baseType.Align == TypeAlignment_Double {
		align = TypeAlignment_Double
	}
	return &DoltgresType{
		OID:           baseType.Array,
		Name:          fmt.Sprintf("_%s", baseType.Name),
		Schema:        "pg_catalog",
		TypLength:     int16(-1),
		PassedByVal:   false,
		TypType:       TypeType_Base,
		TypCategory:   TypeCategory_ArrayTypes,
		IsPreferred:   false,
		IsDefined:     true,
		Delimiter:     ",",
		RelID:         0,
		SubscriptFunc: toFuncID("array_subscript_handler", oid.T_internal),
		Elem:          baseType.OID,
		Array:         0,
		InputFunc:     toFuncID("array_in", oid.T_cstring, oid.T_oid, oid.T_int4),
		OutputFunc:    toFuncID("array_out", oid.T_anyarray),
		ReceiveFunc:   toFuncID("array_recv", oid.T_internal, oid.T_oid, oid.T_int4),
		SendFunc:      toFuncID("array_send", oid.T_anyarray),
		ModInFunc:     baseType.ModInFunc,
		ModOutFunc:    baseType.ModOutFunc,
		AnalyzeFunc:   toFuncID("array_typanalyze", oid.T_internal),
		Align:         align,
		Storage:       TypeStorage_Extended,
		NotNull:       false,
		BaseTypeOID:   0,
		TypMod:        -1,
		NDims:         0,
		TypCollation:  baseType.TypCollation,
		DefaulBin:     "",
		Default:       "",
		Acl:           nil,
		Checks:        nil,
		InternalName:  fmt.Sprintf("%s[]", baseType.Name), // This will be set to the proper name in ToArrayType
		attTypMod:     baseType.attTypMod,                 // TODO: check
		CompareFunc:   toFuncID("btarraycmp", oid.T_anyarray, oid.T_anyarray),
	}
}
