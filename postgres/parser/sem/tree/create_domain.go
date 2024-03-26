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

var _ Statement = &CreateDomain{}

// CreateDomain represents a CREATE DOMAIN statement.
type CreateDomain struct {
	TypeName    *UnresolvedObjectName
	DataType    ResolvableTypeReference
	Collate     string
	Default     Expr
	Constraints DomainConstraints
}

// Format implements the NodeFormatter interface.
func (node *CreateDomain) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE DOMAIN ")
	ctx.FormatNode(node.TypeName)
	ctx.WriteString(" AS ")
	ctx.WriteString(node.DataType.SQLString())
	if node.Collate != "" {
		ctx.WriteString(" COLLATE ")
		ctx.WriteString(node.Collate)
	}
	if node.Default != nil {
		ctx.WriteString(" DEFAULT ")
		ctx.FormatNode(node.Default)
	}
	if node.Constraints != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.Constraints)
	}
}

type DomainConstraints []DomainConstraint

// Format implements the NodeFormatter interface.
func (node *DomainConstraints) Format(ctx *FmtCtx) {
	for i, constraint := range *node {
		if i != 0 {
			ctx.WriteByte(' ')
		}
		ctx.FormatNode(&constraint)
	}
}

// DomainConstraint represents the {NOT NULL|NULL|CHECK} constraint clauses for CREATE DOMAIN command.
type DomainConstraint struct {
	Constraint Name
	NotNull    bool
	Check      Expr
}

// Format implements the NodeFormatter interface.
func (node *DomainConstraint) Format(ctx *FmtCtx) {
	if node.Constraint != "" {
		ctx.WriteString("CONSTRAINT ")
		ctx.FormatNode(&node.Constraint)
		ctx.WriteByte(' ')
	}
	if node.Check != nil {
		ctx.WriteString("CHECK ( ")
		ctx.FormatNode(node.Check)
		ctx.WriteString(" )")
	} else if node.NotNull {
		ctx.WriteString("NOT NULL")
	} else {
		ctx.WriteString("NULL")
	}
}
