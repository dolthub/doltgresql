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

	"github.com/dolthub/doltgresql/core/id"
)

// CreateArrayTypeFromBaseType create array type from given type.
func CreateArrayTypeFromBaseType(baseType *DoltgresType) *DoltgresType {
	align := TypeAlignment_Int
	if baseType.Align == TypeAlignment_Double {
		align = TypeAlignment_Double
	}
	return &DoltgresType{
		ID:            baseType.Array,
		TypLength:     int16(-1),
		PassedByVal:   false,
		TypType:       TypeType_Base,
		TypCategory:   TypeCategory_ArrayTypes,
		IsPreferred:   false,
		IsDefined:     true,
		Delimiter:     ",",
		RelID:         id.Null,
		SubscriptFunc: toFuncID("array_subscript_handler", toInternal("internal")),
		Elem:          baseType.ID,
		Array:         id.Null,
		InputFunc:     toFuncID("array_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
		OutputFunc:    toFuncID("array_out", toInternal("anyarray")),
		ReceiveFunc:   toFuncID("array_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
		SendFunc:      toFuncID("array_send", toInternal("anyarray")),
		ModInFunc:     baseType.ModInFunc,
		ModOutFunc:    baseType.ModOutFunc,
		AnalyzeFunc:   toFuncID("array_typanalyze", toInternal("internal")),
		Align:         align,
		Storage:       TypeStorage_Extended,
		NotNull:       false,
		BaseTypeID:    id.Null,
		TypMod:        -1,
		NDims:         0,
		TypCollation:  baseType.TypCollation,
		DefaulBin:     "",
		Default:       "",
		Acl:           nil,
		Checks:        nil,
		InternalName:  fmt.Sprintf("%s[]", baseType.Name()), // This will be set to the proper name in ToArrayType
		attTypMod:     baseType.attTypMod,                   // TODO: check
		CompareFunc:   toFuncID("btarraycmp", toInternal("anyarray"), toInternal("anyarray")),
	}
}
