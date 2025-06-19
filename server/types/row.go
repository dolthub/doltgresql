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
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/go-mysql-server/sql"
)

// Row is a pseudo-type that is solely used as a return type for TRIGGER functions.
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

type SetRow struct {
	returnType *DoltgresType
}

type RowValues struct {
	values []any         // pointer??
	dt     *DoltgresType // array of types? if it's multiple type, it should be composite type
	count  int32
}

func NewRowValues(values []any, dt *DoltgresType, count int32) *RowValues {
	return &RowValues{
		values: values,
		dt:     dt,
		count:  count,
	}
}

func (s *RowValues) Count() int32 {
	return s.count
}

func (s *RowValues) Type() sql.Type {
	return s.dt
}

func (s *RowValues) GetRow(ctx *sql.Context, i int32) (any, error) {
	if i >= s.count {
		return "", nil // TODO: should error
	}
	return s.values[i], nil
}
