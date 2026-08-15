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

package tree

var _ Statement = &CreateOperator{}

// CreateOperator represents a CREATE OPERATOR statement.
type CreateOperator struct {
	Name    Operator
	Options []CreateOperatorOption
}

// Format implements the NodeFormatter interface.
func (node *CreateOperator) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE OPERATOR ")
	ctx.WriteString(OperatorSymbol(node.Name))
	ctx.WriteString(" ( ")
	for i := range node.Options {
		if i > 0 {
			ctx.WriteString(" , ")
		}
		ctx.FormatNode(&node.Options[i])
	}
	ctx.WriteString(" )")
}

// CreateOperatorOption represents a single option of a CREATE OPERATOR statement.
type CreateOperatorOption struct {
	Option CreateOperatorOptionType
	// FuncName is used for Function, Restrict, and Join
	FuncName *UnresolvedObjectName
	// TypeVal is used for LeftArg and RightArg
	TypeVal ResolvableTypeReference
	// OpVal is used for Commutator and Negator
	OpVal Operator
	// Hashes and Merges do not define any stored value.
}

// Format implements the NodeFormatter interface.
func (node *CreateOperatorOption) Format(ctx *FmtCtx) {
	switch node.Option {
	case OperatorOptTypeFunction:
		ctx.WriteString("FUNCTION = ")
		ctx.FormatNode(node.FuncName)
	case OperatorOptTypeLeftArg:
		ctx.WriteString("LEFTARG = ")
		ctx.WriteString(node.TypeVal.SQLString())
	case OperatorOptTypeRightArg:
		ctx.WriteString("RIGHTARG = ")
		ctx.WriteString(node.TypeVal.SQLString())
	case OperatorOptTypeCommutator:
		ctx.WriteString("COMMUTATOR = ")
		ctx.WriteString(OperatorSymbol(node.OpVal))
	case OperatorOptTypeNegator:
		ctx.WriteString("NEGATOR = ")
		ctx.WriteString(OperatorSymbol(node.OpVal))
	case OperatorOptTypeRestrict:
		ctx.WriteString("RESTRICT = ")
		ctx.FormatNode(node.FuncName)
	case OperatorOptTypeJoin:
		ctx.WriteString("JOIN = ")
		ctx.FormatNode(node.FuncName)
	case OperatorOptTypeHashes:
		ctx.WriteString("HASHES")
	case OperatorOptTypeMerges:
		ctx.WriteString("MERGES")
	}
}

// OperatorSymbol returns the operator's symbol.
func OperatorSymbol(op Operator) string {
	switch op := op.(type) {
	case UnaryOperator:
		return op.String()
	case BinaryOperator:
		return op.String()
	case ComparisonOperator:
		return op.String()
	}
	return ""
}

// CreateOperatorOptionType represents the type of a CREATE OPERATOR option.
type CreateOperatorOptionType int

const (
	OperatorOptTypeFunction CreateOperatorOptionType = iota
	OperatorOptTypeLeftArg
	OperatorOptTypeRightArg
	OperatorOptTypeCommutator
	OperatorOptTypeNegator
	OperatorOptTypeRestrict
	OperatorOptTypeJoin
	OperatorOptTypeHashes
	OperatorOptTypeMerges
)
