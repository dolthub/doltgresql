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
	"github.com/lib/pq/oid"
	"gopkg.in/src-d/go-errors.v1"
)

const (
	// StringMaxLength is the maximum number of characters (not bytes) that a Char, VarChar, or BpChar may contain.
	StringMaxLength = 10485760
	// stringInline is the maximum number of characters (not bytes) that are "guaranteed" to fit inline.
	stringInline = 16383
	// StringUnbounded is used to represent that a type does not define a limit on the strings that it accepts. Values
	// are still limited by the field size limit, but it won't be enforced by the type.
	StringUnbounded = 0
)

// ErrLengthMustBeAtLeast1 is returned when given character length is less than 1.
var ErrLengthMustBeAtLeast1 = errors.NewKind(`length for type %s must be at least 1`)

// ErrLengthCannotExceed is returned when given character length exceeds the upper bound, 10485760.
var ErrLengthCannotExceed = errors.NewKind(`length for type %s cannot exceed 10485760`)

// VarChar is a varchar that has an unbounded length.
var VarChar = &DoltgresType{
	OID:           uint32(oid.T_varchar),
	Name:          "varchar",
	Schema:        "pg_catalog",
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_StringTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__varchar),
	InputFunc:     toFuncID("varcharin", oid.T_cstring, oid.T_oid, oid.T_int4),
	OutputFunc:    toFuncID("varcharout", oid.T_varchar),
	ReceiveFunc:   toFuncID("varcharrecv", oid.T_internal, oid.T_oid, oid.T_int4),
	SendFunc:      toFuncID("varcharsend", oid.T_varchar),
	ModInFunc:     toFuncID("varchartypmodin", oid.T__cstring),
	ModOutFunc:    toFuncID("varchartypmodout", oid.T_int4),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Int,
	Storage:       TypeStorage_Extended,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  100,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	AttTypMod:     -1,
	CompareFunc:   toFuncID("bttextcmp", oid.T_text, oid.T_text), // TODO: temporarily added
}

// NewVarCharType returns VarChar type with type modifier set
// representing the maximum number of characters that the type may hold.
func NewVarCharType(maxChars int32) (*DoltgresType, error) {
	var err error
	newType := *VarChar
	newType.AttTypMod, err = GetTypModFromCharLength("varchar", maxChars)
	if err != nil {
		return nil, err
	}
	return &newType, nil
}

// MustCreateNewVarCharType panics if used with out-of-bound value.
func MustCreateNewVarCharType(maxChars int32) *DoltgresType {
	newType, err := NewVarCharType(maxChars)
	if err != nil {
		panic(err)
	}
	return newType
}

// GetTypModFromCharLength takes character type and its length and returns the type modifier value.
func GetTypModFromCharLength(typName string, l int32) (int32, error) {
	if l < 1 {
		return 0, ErrLengthMustBeAtLeast1.New(typName)
	} else if l > StringMaxLength {
		return 0, ErrLengthCannotExceed.New(typName)
	}
	return l + 4, nil
}

// GetCharLengthFromTypmod takes character type modifier and returns length value.
func GetCharLengthFromTypmod(typmod int32) int32 {
	return typmod - 4
}
