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

// Row is a pseudo-type that is solely used as a return type for set returning functions.
var Row = &DoltgresType{
	ID:            toInternal("row"),
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Pseudo,
	TypCategory:   TypeCategory_PseudoTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.NullType,
	Array:         id.NullType,
	InputFunc:     toFuncID("_"),
	OutputFunc:    toFuncID("_"),
	ReceiveFunc:   toFuncID("-"),
	SendFunc:      toFuncID("-"),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Double,
	Storage:       TypeStorage_Extended,
	NotNull:       false,
	BaseTypeID:    id.NullType,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  id.NullCollation,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("-"),
}

// RowTypeWithReturnType returns Row type with Elem set to given type id.
// We reuse the Elem field to store the given type as it's only used for array types, which it's safely checked if
// the type is array type before used.
func RowTypeWithReturnType(baseType *DoltgresType) *DoltgresType {
	rt := *Row
	rt.Elem = baseType.ID
	rt.InputFunc = baseType.InputFunc
	rt.OutputFunc = baseType.OutputFunc
	return &rt
}

var _ sql.RowIter = (*SetReturningFunctionRowIter)(nil)

// SetReturningFunctionRowIter is used for value returned from functions that return multiple rows.
type SetReturningFunctionRowIter struct {
	next func(ctx *sql.Context) (sql.Row, error)
}

// NewSetReturningFunctionRowIter creates a new SetReturningFunctionRowIter as value returned from set returning functions that return Row Type.
func NewSetReturningFunctionRowIter(next func(ctx *sql.Context) (sql.Row, error)) *SetReturningFunctionRowIter {
	return &SetReturningFunctionRowIter{
		next: next,
	}
}

// Next implements the interface sql.RowIter.
func (s *SetReturningFunctionRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return s.next(ctx)
}

// Close implements the interface sql.RowIter.
func (s *SetReturningFunctionRowIter) Close(_ *sql.Context) error {
	return nil
}
