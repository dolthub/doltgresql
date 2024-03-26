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

var _ Statement = &AlterDomain{}

// AlterDomain represents a ALTER DOMAIN statement.
type AlterDomain struct {
	Name *UnresolvedObjectName
	Cmd  AlterDomainCmd
}

// Format implements the NodeFormatter interface.
func (node *AlterDomain) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER DOMAIN ")
	ctx.FormatNode(node.Name)
	ctx.FormatNode(node.Cmd)
}

// AlterDomainCmd represents a domain modification operation.
type AlterDomainCmd interface {
	NodeFormatter
	alterDomainCmd()
}

func (*AlterDomainSetDrop) alterDomainCmd()    {}
func (*AlterDomainConstraint) alterDomainCmd() {}
func (*AlterDomainOwner) alterDomainCmd()      {}
func (*AlterDomainRename) alterDomainCmd()     {}
func (*AlterDomainSetSchema) alterDomainCmd()  {}

var _ AlterDomainCmd = &AlterDomainSetDrop{}
var _ AlterDomainCmd = &AlterDomainConstraint{}
var _ AlterDomainCmd = &AlterDomainOwner{}
var _ AlterDomainCmd = &AlterDomainRename{}
var _ AlterDomainCmd = &AlterDomainSetSchema{}

// AlterDomainSetDrop represents an ALTER DOMAIN {SET|DROP} {DEFAULT|NOT NULL} command.
type AlterDomainSetDrop struct {
	IsSet   bool
	Default Expr
	NotNull bool
}

// Format implements the NodeFormatter interface.
func (node *AlterDomainSetDrop) Format(ctx *FmtCtx) {
	if node.IsSet {
		ctx.WriteString(" SET ")
	} else {
		ctx.WriteString(" DROP ")
	}
	if node.NotNull {
		ctx.WriteString("NOT NULL")
	} else {
		ctx.WriteString("DEFAULT")
		if node.Default != nil {
			ctx.WriteByte(' ')
			ctx.FormatNode(node.Default)
		}
	}
}

// AlterDomainConstraintAction represents the action for constraint clause for ALTER DOMAIN command.
type AlterDomainConstraintAction int

const (
	AlterDomainAddConstraint AlterDomainConstraintAction = iota
	AlterDomainDropConstraint
	AlterDomainRenameConstraint
	AlterDomainValidateConstraint
)

// AlterDomainConstraint represents an ALTER DOMAIN {ADD|DROP|RENAME|VALIDATE} CONSTRAINT command.
type AlterDomainConstraint struct {
	Action         AlterDomainConstraintAction
	Constraint     DomainConstraint
	NotValid       bool
	IfExists       bool
	ConstraintName Name
	DropBehavior   DropBehavior
	NewName        Name
}

// Format implements the NodeFormatter interface.
func (node *AlterDomainConstraint) Format(ctx *FmtCtx) {
	switch node.Action {
	case AlterDomainAddConstraint:
		ctx.WriteString(" ADD ")
		ctx.FormatNode(&node.Constraint)
		if node.NotValid {
			ctx.WriteString(" NOT VALID")
		}
	case AlterDomainDropConstraint:
		ctx.WriteString(" DROP ")
		if node.IfExists {
			ctx.WriteString("IF EXISTS ")
		}
		ctx.FormatNode(&node.ConstraintName)
		if db := node.DropBehavior.String(); db != "" {
			ctx.WriteByte(' ')
			ctx.WriteString(db)
		}
	case AlterDomainRenameConstraint:
		ctx.WriteString(" RENAME ")
		ctx.FormatNode(&node.ConstraintName)
		ctx.WriteString(" TO ")
		ctx.FormatNode(&node.NewName)
	case AlterDomainValidateConstraint:
		ctx.WriteString(" VALIDATE ")
		ctx.FormatNode(&node.ConstraintName)
	}
}

// AlterDomainOwner represents an ALTER DOMAIN OWNER TO command.
type AlterDomainOwner struct {
	Owner string
}

// Format implements the NodeFormatter interface.
func (node *AlterDomainOwner) Format(ctx *FmtCtx) {
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNameP(&node.Owner)
}

// AlterDomainRename represents an ALTER DOMAIN RENAME TO command.
type AlterDomainRename struct {
	NewName string
}

// Format implements the NodeFormatter interface.
func (node *AlterDomainRename) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME TO ")
	ctx.WriteString(node.NewName)
}

// AlterDomainSetSchema represents an ALTER DOMAIN SET SCHEMA command.
type AlterDomainSetSchema struct {
	Schema string
}

// Format implements the NodeFormatter interface.
func (node *AlterDomainSetSchema) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET SCHEMA ")
	ctx.WriteString(node.Schema)
}
