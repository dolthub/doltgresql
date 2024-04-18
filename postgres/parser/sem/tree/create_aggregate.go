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

package tree

import "strings"

var _ Statement = &CreateAggregate{}

// CreateAggregate represents a CREATE AGGREGATE statement.
type CreateAggregate struct {
	Name        *UnresolvedObjectName
	Replace     bool
	Args        RoutineArgs
	SFunc       string
	SType       ResolvableTypeReference
	AggOptions  CreateAggOptions
	OrderByArgs RoutineArgs
	BaseType    ResolvableTypeReference
}

// Format implements the NodeFormatter interface.
func (node *CreateAggregate) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE ")
	if node.Replace {
		ctx.WriteString("OR REPLACE ")
	}
	ctx.WriteString("AGGREGATE ")
	ctx.FormatNode(node.Name)
	ctx.WriteString(" ( ")
	if node.OrderByArgs != nil {
		if node.Args != nil {
			ctx.FormatNode(node.Args)
		}
		ctx.WriteString(" ORDER BY ")
		ctx.FormatNode(node.OrderByArgs)
		ctx.WriteString(" ) ( ")
	} else if node.Args != nil {
		ctx.FormatNode(node.Args)
		ctx.WriteString(" ) ( ")
	} else {
		ctx.WriteString("BASETYPE = ")
		ctx.WriteString(node.BaseType.SQLString())
		ctx.WriteString(" , ")
	}
	ctx.WriteString("SFUNC = ")
	ctx.WriteString(node.SFunc)
	ctx.WriteString(" , STYPE = ")
	ctx.WriteString(node.BaseType.SQLString())
	if node.AggOptions != nil {
		ctx.FormatNode(&node.AggOptions)
	}
}

type FinalFuncModifyType string

const (
	FinalFuncModifyReadOnly  FinalFuncModifyType = "READ_ONLY"
	FinalFuncModifyShareable FinalFuncModifyType = "SHAREABLE"
	FinalFuncModifyReadWrite FinalFuncModifyType = "READ_WRITE"
)

type CreateAggOptions []CreateAggOption

// Format implements the NodeFormatter interface.
func (node *CreateAggOptions) Format(ctx *FmtCtx) {
	for _, option := range *node {
		ctx.WriteString(" , ")
		ctx.FormatNode(&option)
	}
}

type CreateAggOption struct {
	Option CreateOptionType
	// IntVal is used for SSpace and MSSpace
	IntVal Expr
	// CondVal is used for InitCond and MInitCond
	CondVal Expr
	// BoolVal is used for FinalFuncExtra and MFinalFuncExtra
	BoolVal bool
	// StrVal is used for FinalFunc, CombineFunc, SerialFunc,
	// DeserialFunc, MSFunc, MInvFunc and MFinalFunc
	StrVal string
	// FinalFuncModify is used for FinalFuncModify and MFinalFuncModify
	FinalFuncModify FinalFuncModifyType
	// TypeVal is used for MSType
	TypeVal  ResolvableTypeReference
	Parallel Parallel
	SortOp   Operator
	// Hypothetical does not define any stored value.
}

// Format implements the NodeFormatter interface.
func (node *CreateAggOption) Format(ctx *FmtCtx) {
	switch node.Option {
	case AggOptTypeSSpace:
		ctx.WriteString("SSPACE = ")
		ctx.FormatNode(node.IntVal)
	case AggOptTypeFinalFunc:
		ctx.WriteString("FINALFUNC = ")
		ctx.WriteString(node.StrVal)
	case AggOptTypeFinalFuncExtra:
		if node.BoolVal {
			ctx.WriteString("FINALFUNC_EXTRA = TRUE")
		} else {
			ctx.WriteString("FINALFUNC_EXTRA = FALSE")
		}
	case AggOptTypeFinalFuncModify:
		ctx.WriteString("FINALFUNC_MODIFY")
		ctx.WriteString(string(node.FinalFuncModify))
	case AggOptTypeCombineFunc:
		ctx.WriteString("COMBINEFUNC = ")
		ctx.WriteString(node.StrVal)
	case AggOptTypeSerialFunc:
		ctx.WriteString("SERIALFUNC = ")
		ctx.WriteString(node.StrVal)
	case AggOptTypeDeserialFunc:
		ctx.WriteString("DESERIALFUNC = ")
		ctx.WriteString(node.StrVal)
	case AggOptTypeInitCond:
		ctx.WriteString("INITCOND = ")
		ctx.FormatNode(node.CondVal)
	case AggOptTypeMSFunc:
		ctx.WriteString("MSFUNC = ")
		ctx.WriteString(node.StrVal)
	case AggOptTypeMInvFunc:
		ctx.WriteString("MINVFUNC = ")
		ctx.WriteString(node.StrVal)
	case AggOptTypeMSType:
		ctx.WriteString("MSTYPE = ")
		ctx.WriteString(node.TypeVal.SQLString())
	case AggOptTypeMSSpace:
		ctx.WriteString("MSSPACE = ")
		ctx.FormatNode(node.IntVal)
	case AggOptTypeMFinalFunc:
		ctx.WriteString("MFINALFUNC = ")
		ctx.WriteString(node.StrVal)
	case AggOptTypeMFinalFuncExtra:
		if node.BoolVal {
			ctx.WriteString("MFINALFUNC_EXTRA = TRUE")
		} else {
			ctx.WriteString("MFINALFUNC_EXTRA = FALSE")
		}
	case AggOptTypeMFinalFuncModify:
		ctx.WriteString("MFINALFUNC_MODIFY")
		ctx.WriteString(string(node.FinalFuncModify))
	case AggOptTypeMInitCond:
		ctx.WriteString("MINITCOND = ")
		ctx.FormatNode(node.CondVal)
	case AggOptTypeSortOp:
		ctx.WriteString("SORTOP = ")
		switch op := node.SortOp.(type) {
		case UnaryOperator:
			ctx.WriteString(op.String())
		case BinaryOperator:
			ctx.WriteString(op.String())
		case ComparisonOperator:
			ctx.WriteString(op.String())
		}
	case AggOptTypeParallel:
		ctx.WriteString("PARALLEL = ")
		ctx.WriteString(strings.ToUpper(string(node.Parallel)))
	case AggOptTypeHypothetical:
		ctx.WriteString("HYPOTHETICAL")
	}
}

type CreateOptionType int

const (
	AggOptTypeSSpace CreateOptionType = iota
	AggOptTypeFinalFunc
	AggOptTypeFinalFuncExtra
	AggOptTypeFinalFuncModify
	AggOptTypeCombineFunc
	AggOptTypeSerialFunc
	AggOptTypeDeserialFunc
	AggOptTypeInitCond
	AggOptTypeMSFunc
	AggOptTypeMInvFunc
	AggOptTypeMSType
	AggOptTypeMSSpace
	AggOptTypeMFinalFunc
	AggOptTypeMFinalFuncExtra
	AggOptTypeMFinalFuncModify
	AggOptTypeMInitCond
	AggOptTypeSortOp
	AggOptTypeParallel
	AggOptTypeHypothetical
)
