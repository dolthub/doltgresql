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

// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in licenses/BSD-vitess.txt.

// Portions of this file are additionally subject to the following
// license and copyright.
//
// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// This code was derived from https://github.com/youtube/vitess.

package tree

import "github.com/dolthub/doltgresql/postgres/parser/lex"

// CreateTypeVariety represents a particular variety of user defined types.
type CreateTypeVariety int

//go:generate stringer -type=CreateTypeVariety
const (
	_ CreateTypeVariety = iota
	// Composite represents a composite user defined type.
	Composite
	// Enum represents an ENUM user defined type.
	Enum
	// Range represents a RANGE user defined type.
	Range
	// Base represents a base user defined type.
	Base
	// Shell represents a shell user defined type. Represents a `CREATE TYPE name` statement.
	Shell
	// Domain represents a DOMAIN user defined type.
	Domain
)

var _ Statement = &CreateType{}

// CreateType represents a CREATE TYPE statement.
type CreateType struct {
	TypeName  *UnresolvedObjectName
	Variety   CreateTypeVariety
	Composite CompositeType
	Enum      EnumType
	Range     RangeType
	Base      BaseType
}

var _ Statement = &CreateType{}

// Format implements the NodeFormatter interface.
func (node *CreateType) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE TYPE ")
	ctx.WriteString(node.TypeName.String())
	switch node.Variety {
	case Composite:
		ctx.FormatNode(&node.Composite)
	case Enum:
		ctx.FormatNode(&node.Enum)
	case Range:
		ctx.FormatNode(&node.Range)
	case Base:
		ctx.FormatNode(&node.Base)
	case Shell:
	case Domain:
		// Do it separate?
	}

}

// CompositeType represents a `CREATE TYPE <name> AS ( <type>, ... )` statement.
type CompositeType struct {
	Types []CompositeTypeElem
}

type CompositeTypeElem struct {
	AttrName string
	Type     ResolvableTypeReference
	// below fields are optional
	Collate string
}

func (node *CompositeType) Format(ctx *FmtCtx) {
	ctx.WriteString(" AS ( ")
	for i, t := range node.Types {
		if i > 0 {
			ctx.WriteString(" , ")
		}
		ctx.WriteString(t.AttrName)
		ctx.WriteByte(' ')
		ctx.WriteString(t.Type.SQLString())
		if t.Collate != "" {
			ctx.WriteString(" COLLATE ")
			ctx.WriteString(t.Collate)
		}
	}
	ctx.WriteString(" )")
}

// EnumType represents a `CREATE TYPE <name> AS ENUM ( 'label', ... )` statement.
type EnumType struct {
	Labels []string
}

func (node *EnumType) Format(ctx *FmtCtx) {
	ctx.WriteString(" AS ENUM ( ")
	for i := range node.Labels {
		if i > 0 {
			ctx.WriteString(" , ")
		}
		lex.EncodeSQLString(&ctx.Buffer, node.Labels[i])
	}
	ctx.WriteString(" )")
}

type RangeTypeOptionType int

const (
	RangeTypeSubtypeOpClass RangeTypeOptionType = iota
	RangeTypeCollation
	RangeTypeCanonical
	RangeTypeSubtypeDiff
	RangeTypeMultiRangeTypeName
)

// RangeType represents a `CREATE TYPE <name> AS RANGE ( <subtype>, ... )` statement.
type RangeType struct {
	Subtype ResolvableTypeReference
	Options []RangeTypeOption
}

type RangeTypeOption struct {
	Option     RangeTypeOptionType
	StrVal     string
	MRTypeName *UnresolvedObjectName
}

func (node *RangeType) Format(ctx *FmtCtx) {
	ctx.WriteString(" AS RANGE ( SUBTYPE ")
	ctx.WriteString(node.Subtype.SQLString())
	for i, opt := range node.Options {
		if i > 0 {
			ctx.WriteString(" , ")
		}
		switch opt.Option {
		case RangeTypeSubtypeOpClass:
			ctx.WriteString("SUBTYPE_OPCLASS = ")
			ctx.WriteString(opt.StrVal)
		case RangeTypeCollation:
			ctx.WriteString("COLLATION = ")
			ctx.WriteString(opt.StrVal)
		case RangeTypeCanonical:
			ctx.WriteString("CANONICAL = ")
			ctx.WriteString(opt.StrVal)
		case RangeTypeSubtypeDiff:
			ctx.WriteString("SUBTYPE_DIFF = ")
			ctx.WriteString(opt.StrVal)
		case RangeTypeMultiRangeTypeName:
			ctx.WriteString("MULTIRANGE_TYPE_NAME = ")
			ctx.FormatNode(opt.MRTypeName)
		}
	}
	ctx.WriteString(" )")
}

type BaseTypeOptionType int

const (
	BaseTypeReceive BaseTypeOptionType = iota
	BaseTypeSend
	BaseTypeTypModIn
	BaseTypeTypeModOut
	BaseTypeAnalyze
	BaseTypeSubscript
	BaseTypeInternalLength
	BaseTypePassedByValue
	BaseTypeAlignment
	BaseTypeStorage
	BaseTypeLikeType
	BaseTypeCategory
	BaseTypePreferred
	BaseTypeDefault
	BaseTypeElement
	BaseTypeDelimiter
	BaseTypeCollatable
)

type BaseTypeOptions []BaseTypeOption

type BaseTypeOption struct {
	Option         BaseTypeOptionType
	StrVal         string
	BoolVal        bool
	InternalLength int64
	Default        Expr
	TypeVal        ResolvableTypeReference
}

func (node *BaseTypeOptions) Format(ctx *FmtCtx) {
	for i, opt := range *node {
		if i > 0 {
			ctx.WriteString(" , ")
		}
		switch opt.Option {
		case BaseTypeReceive:
			ctx.WriteString("RECEIVE = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeSend:
			ctx.WriteString("SEND = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeTypModIn:
			ctx.WriteString("TYPMOD_IN = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeTypeModOut:
			ctx.WriteString("TYPMOD_OUT = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeAnalyze:
			ctx.WriteString("ANALYZE = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeSubscript:
			ctx.WriteString("SUBSCRIPT = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeInternalLength:
			if opt.InternalLength == -1 {
				ctx.WriteString("INTERNALLENGTH = VARIABLE")
			} else {
				ctx.Printf("INTERNALLENGTH = %d", opt.InternalLength)
			}
		case BaseTypePassedByValue:
			ctx.WriteString("PASSEDBYVALUE")
		case BaseTypeAlignment:
			ctx.WriteString("ALIGNMENT = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeStorage:
			ctx.WriteString("STORAGE = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeLikeType:
			ctx.WriteString("LIKE = ")
			ctx.WriteString(opt.TypeVal.SQLString())
		case BaseTypeCategory:
			ctx.WriteString("CATEGORY = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypePreferred:
			if opt.BoolVal {
				ctx.WriteString("PREFERRED = TRUE")
			} else {
				ctx.WriteString("PREFERRED = FALSE")
			}
		case BaseTypeDefault:
			ctx.WriteString("DEFAULT = ")
			ctx.FormatNode(opt.Default)
		case BaseTypeElement:
			ctx.WriteString("ELEMENT = ")
			ctx.WriteString(opt.TypeVal.SQLString())
		case BaseTypeDelimiter:
			ctx.WriteString("DELIMITER = ")
			ctx.WriteString(opt.StrVal)
		case BaseTypeCollatable:
			if opt.BoolVal {
				ctx.WriteString("COLLATABLE = TRUE")
			} else {
				ctx.WriteString("COLLATABLE = FALSE")
			}
		}
	}
}

// BaseType represents a `CREATE TYPE <name> ( INPUT = input_function, OUTPUT = output_function, ... )` statement.
type BaseType struct {
	Input   string
	Output  string
	Options BaseTypeOptions
}

func (node *BaseType) Format(ctx *FmtCtx) {
	ctx.WriteString(" ( INPUT = ")
	ctx.WriteString(node.Input)
	ctx.WriteString(", OUTPUT = ")
	ctx.WriteString(node.Output)
	ctx.WriteByte(' ')
	ctx.FormatNode(&node.Options)
	ctx.WriteString(" )")
}
