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

// CreateCastType states the type of the CREATE CAST statement.
type CreateCastType byte

const (
	CreateCastType_WithFunction CreateCastType = iota
	CreateCastType_WithoutFunction
	CreateCastType_Inout
)

// CreateCastScope states whether the CREATE CAST statement is explicit, assignment, or implicit.
type CreateCastScope byte

const (
	CreateCastScope_Explicit CreateCastScope = iota
	CreateCastScope_Assignment
	CreateCastScope_Implicit
)

// CreateCast represents a CREATE CAST statement.
type CreateCast struct {
	Source   ResolvableTypeReference
	Target   ResolvableTypeReference
	Scope    CreateCastScope
	Type     CreateCastType
	FuncName *UnresolvedObjectName
	FuncArgs RoutineArgs
}

var _ Statement = &CreateCast{}

// Format implements the NodeFormatter interface.
func (node *CreateCast) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE CAST (")
	ctx.WriteString(node.Source.SQLString())
	ctx.WriteString(" AS ")
	ctx.WriteString(node.Target.SQLString())
	ctx.WriteString(") ")
	switch node.Type {
	case CreateCastType_WithFunction:
		ctx.WriteString("WITH FUNCTION ")
		ctx.FormatNode(node.FuncName)
		if len(node.FuncArgs) > 0 {
			ctx.WriteString("(")
			ctx.FormatNode(node.FuncArgs)
			ctx.WriteString(")")
		}
	case CreateCastType_WithoutFunction:
		ctx.WriteString("WITHOUT FUNCTION")
	case CreateCastType_Inout:
		ctx.WriteString("WITH INOUT")
	}
	switch node.Scope {
	case CreateCastScope_Explicit:
		// Nothing to write
	case CreateCastScope_Assignment:
		ctx.WriteString(" AS ASSIGNMENT")
	case CreateCastScope_Implicit:
		ctx.WriteString(" AS IMPLICIT")
	}
}
