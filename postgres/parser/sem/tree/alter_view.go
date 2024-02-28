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

var _ Statement = &AlterView{}

// AlterView represents an ALTER VIEW statement.
type AlterView struct {
	Name     *UnresolvedObjectName
	IfExists bool
	Cmd      AlterViewCmd
}

func (node *AlterView) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER VIEW ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(node.Name)
	ctx.FormatNode(node.Cmd)
}

// AlterViewCmd represents a view modification operation.
type AlterViewCmd interface {
	NodeFormatter
	// Placeholder function to ensure that only desired views
	// (AlterView*) conform to the AlterViewCmd interface.
	alterViewCmd()
}

func (*AlterViewSetDefault) alterViewCmd()   {}
func (*AlterViewOwnerTo) alterViewCmd()      {}
func (*AlterViewRenameColumn) alterViewCmd() {}
func (*AlterViewRenameTo) alterViewCmd()     {}
func (*AlterViewSetSchema) alterViewCmd()    {}
func (*AlterViewSetOption) alterViewCmd()    {}

var _ AlterViewCmd = &AlterViewSetDefault{}
var _ AlterViewCmd = &AlterViewOwnerTo{}
var _ AlterViewCmd = &AlterViewRenameColumn{}
var _ AlterViewCmd = &AlterViewRenameTo{}
var _ AlterViewCmd = &AlterViewSetSchema{}
var _ AlterViewCmd = &AlterViewSetOption{}

// AlterViewSetDefault represents an ALTER VIEW ALTER COLUMN SET DEFAULT
// or DROP DEFAULT command.
type AlterViewSetDefault struct {
	Column  Name
	Default Expr
}

// Format implements the NodeFormatter interface.
func (node *AlterViewSetDefault) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	if node.Default == nil {
		ctx.WriteString(" DROP DEFAULT")
	} else {
		ctx.WriteString(" SET DEFAULT ")
		ctx.FormatNode(node.Default)
	}
}

// AlterViewOwnerTo represents an ALTER VIEW OWNER TO command.
type AlterViewOwnerTo struct {
	Owner string
}

// Format implements the NodeFormatter interface.
func (node *AlterViewOwnerTo) Format(ctx *FmtCtx) {
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNameP(&node.Owner)
}

// AlterViewRenameColumn represents an ALTER VIEW RENAME [COLUMN] command.
type AlterViewRenameColumn struct {
	Column  Name
	NewName Name
}

// Format implements the NodeFormatter interface.
func (node *AlterViewRenameColumn) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewName)
}

// AlterViewRenameTo represents an ALTER VIEW ... RENAME TO
// command.
type AlterViewRenameTo struct {
	Rename *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *AlterViewRenameTo) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME TO ")
	ctx.FormatNode(node.Rename)
}

// AlterViewSetSchema represents an ALTER VIEW ... SET SCHEMA
// command.
type AlterViewSetSchema struct {
	Schema string
}

// Format implements the NodeFormatter interface.
func (node *AlterViewSetSchema) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET SCHEMA ")
	ctx.WriteString(node.Schema)
}

// AlterViewSetOption represents an ALTER VIEW SET | RESET ... command.
type AlterViewSetOption struct {
	Reset  bool
	Params ViewOptions
}

// Format implements the NodeFormatter interface.
func (node *AlterViewSetOption) Format(ctx *FmtCtx) {
	if node.Reset {
		ctx.WriteString(" RESET ")
	} else {
		ctx.WriteString(" SET ")
	}
	ctx.FormatNode(&node.Params)
}
