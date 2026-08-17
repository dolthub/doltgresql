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
)

// BaseTypeDefinition describes a base type, with every support function already resolved to its registry ID.
type BaseTypeDefinition struct {
	InputFunc   uint32
	OutputFunc  uint32
	ReceiveFunc uint32
	SendFunc    uint32
	ModInFunc   uint32
	ModOutFunc  uint32
	TypLength   int16
	PassedByVal bool
	Align       TypeAlignment
	Storage     TypeStorage
	TypCategory TypeCategory
	IsPreferred bool
	Default     string
	Elem        *DoltgresType
	Delimiter   string
	Collatable  bool
}

// NewBaseTypeDefinition returns the definition of a base type whose options have all been omitted.
func NewBaseTypeDefinition() BaseTypeDefinition {
	return BaseTypeDefinition{
		TypLength:   -1,
		Align:       TypeAlignment_Int,
		Storage:     TypeStorage_Plain,
		TypCategory: TypeCategory_UserDefinedTypes,
		Delimiter:   ",",
	}
}

// NewBaseType returns a new base type from the given definition.
func NewBaseType(ctx *sql.Context, typeID id.Type, def BaseTypeDefinition) *DoltgresType {
	elem := internalNullType
	if def.Elem != nil {
		elem = def.Elem
	}
	collation := id.NullCollation
	if def.Collatable {
		collation = id.NewCollation("pg_catalog", "default")
	}
	return &DoltgresType{
		ID:                  typeID,
		TypLength:           def.TypLength,
		PassedByVal:         def.PassedByVal,
		TypType:             TypeType_Base,
		TypCategory:         def.TypCategory,
		IsPreferred:         def.IsPreferred,
		IsDefined:           true,
		Delimiter:           def.Delimiter,
		RelID:               id.Null,
		SubscriptFunc:       toFuncID("-"),
		Elem:                elem,
		Array:               internalNullType,
		InputFunc:           def.InputFunc,
		OutputFunc:          def.OutputFunc,
		ReceiveFunc:         def.ReceiveFunc,
		SendFunc:            def.SendFunc,
		ModInFunc:           def.ModInFunc,
		ModOutFunc:          def.ModOutFunc,
		AnalyzeFunc:         toFuncID("-"),
		Align:               def.Align,
		Storage:             def.Storage,
		NotNull:             false,
		BaseTypeType:        internalNullType,
		TypMod:              -1,
		NDims:               0,
		TypCollation:        collation,
		DefaulBin:           "",
		Default:             def.Default,
		Acl:                 nil,
		Checks:              nil,
		attTypMod:           -1,
		CompareFunc:         toFuncID("-"),
		SerializationFunc:   nil,
		DeserializationFunc: nil,
	}
}
