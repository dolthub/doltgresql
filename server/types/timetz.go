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
)

// TimeTZ is the time with a time zone. Precision is unbounded.
var TimeTZ = &DoltgresType{
	ID:            toInternal("timetz"),
	TypLength:     int16(12),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_DateTimeTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.NullType,
	Array:         toInternal("_timetz"),
	InputFunc:     toFuncID("timetz_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:    toFuncID("timetz_out", toInternal("timetz")),
	ReceiveFunc:   toFuncID("timetz_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:      toFuncID("timetz_send", toInternal("timetz")),
	ModInFunc:     toFuncID("timetztypmodin", toInternal("_cstring")),
	ModOutFunc:    toFuncID("timetztypmodout", toInternal("int4")),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Double,
	Storage:       TypeStorage_Plain,
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
	CompareFunc:   toFuncID("timetz_cmp", toInternal("timetz"), toInternal("timetz")),
}

// NewTimeTZType returns TimeTZ type with typmod set. // TODO: implement precision
func NewTimeTZType(precision int32) (*DoltgresType, error) {
	typmod, err := GetTypmodFromTimePrecision(precision)
	if err != nil {
		return nil, err
	}
	newType := *TimeTZ.WithAttTypMod(typmod)
	return &newType, nil
}
