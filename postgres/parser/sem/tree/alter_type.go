// Copyright 2023 Dolthub, Inc.
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

// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import "github.com/dolthub/doltgresql/postgres/parser/lex"

//ALTER TYPE name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
//ALTER TYPE name RENAME TO new_name
//ALTER TYPE name SET SCHEMA new_schema
//ALTER TYPE name RENAME ATTRIBUTE attribute_name TO new_attribute_name [ CASCADE | RESTRICT ]
//ALTER TYPE name action [, ... ]
//ALTER TYPE name ADD VALUE [ IF NOT EXISTS ] new_enum_value [ { BEFORE | AFTER } neighbor_enum_value ]
//ALTER TYPE name RENAME VALUE existing_enum_value TO new_enum_value
//ALTER TYPE name SET ( property = value [, ... ] )
//
//where action is one of:
//
//ADD ATTRIBUTE attribute_name data_type [ COLLATE collation ] [ CASCADE | RESTRICT ]
//DROP ATTRIBUTE [ IF EXISTS ] attribute_name [ CASCADE | RESTRICT ]
//ALTER ATTRIBUTE attribute_name [ SET DATA ] TYPE data_type [ COLLATE collation ] [ CASCADE | RESTRICT ]

var _ Statement = &AlterType{}

// AlterType represents an ALTER TYPE statement.
type AlterType struct {
	Type *UnresolvedObjectName
	Cmd  AlterTypeCmd
}

// Format implements the NodeFormatter interface.
func (node *AlterType) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER TYPE ")
	ctx.FormatNode(node.Type)
	ctx.FormatNode(node.Cmd)
}

// AlterTypeCmd represents a type modification operation.
type AlterTypeCmd interface {
	NodeFormatter
	alterTypeCmd()
}

func (*AlterTypeOwner) alterTypeCmd()           {}
func (*AlterTypeRename) alterTypeCmd()          {}
func (*AlterTypeSetSchema) alterTypeCmd()       {}
func (*AlterTypeRenameAttribute) alterTypeCmd() {}
func (*AlterTypeAlterAttribute) alterTypeCmd()  {}
func (*AlterTypeAddValue) alterTypeCmd()        {}
func (*AlterTypeRenameValue) alterTypeCmd()     {}
func (*AlterTypeSetProperty) alterTypeCmd()     {}

var _ AlterTypeCmd = &AlterTypeOwner{}
var _ AlterTypeCmd = &AlterTypeRename{}
var _ AlterTypeCmd = &AlterTypeSetSchema{}
var _ AlterTypeCmd = &AlterTypeRenameAttribute{}
var _ AlterTypeCmd = &AlterTypeAlterAttribute{}
var _ AlterTypeCmd = &AlterTypeAddValue{}
var _ AlterTypeCmd = &AlterTypeRenameValue{}
var _ AlterTypeCmd = &AlterTypeSetProperty{}

// AlterTypeOwner represents an ALTER TYPE OWNER TO command.
type AlterTypeOwner struct {
	Owner string
}

// Format implements the NodeFormatter interface.
func (node *AlterTypeOwner) Format(ctx *FmtCtx) {
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNameP(&node.Owner)
}

// AlterTypeRename represents an ALTER TYPE RENAME command.
type AlterTypeRename struct {
	NewName string
}

// Format implements the NodeFormatter interface.
func (node *AlterTypeRename) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME TO ")
	ctx.WriteString(node.NewName)
}

// AlterTypeSetSchema represents an ALTER TYPE SET SCHEMA command.
type AlterTypeSetSchema struct {
	Schema string
}

// Format implements the NodeFormatter interface.
func (node *AlterTypeSetSchema) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET SCHEMA ")
	ctx.WriteString(node.Schema)
}

// AlterTypeRenameAttribute represents an ALTER TYPE RENAME ATTRIBUTE command.
type AlterTypeRenameAttribute struct {
	ColName      Name
	NewColName   Name
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *AlterTypeRenameAttribute) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME ATTRIBUTE ")
	ctx.FormatNode(&node.ColName)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewColName)
	ctx.WriteByte(' ')
	ctx.WriteString(node.DropBehavior.String())
}

// AlterTypeAlterAttribute represents an ALTER TYPE ADD/DROP/ALTER ATTRIBUTE command.
type AlterTypeAlterAttribute struct {
	Actions []AlterAttributeAction
}

// AlterAttributeAction represents ADD/DROP/ALTER ATTRIBUTE action.
type AlterAttributeAction struct {
	Action       string
	ColName      Name
	IfExists     bool
	TypeName     ResolvableTypeReference
	Collate      string
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *AlterTypeAlterAttribute) Format(ctx *FmtCtx) {
	for i, action := range node.Actions {
		if i > 0 {
			ctx.WriteString(" , ")
		}
		switch action.Action {
		case "add":
			ctx.WriteString(" ADD ATTRIBUTE ")
			ctx.FormatNode(&action.ColName)
			ctx.WriteByte(' ')
			ctx.WriteString(action.TypeName.SQLString())
			if action.Collate != "" {
				ctx.WriteString(" COLLATE")
				ctx.WriteString(action.Collate)
			}
		case "drop":
			ctx.WriteString(" DROP ATTRIBUTE ")
			if action.IfExists {
				ctx.WriteString("IF EXISTS ")
			}
			ctx.FormatNode(&action.ColName)
		case "alter":
			ctx.WriteString(" ALTER ATTRIBUTE ")
			ctx.FormatNode(&action.ColName)
			ctx.WriteString(" SET DATA TYPE ")
			ctx.WriteString(action.TypeName.SQLString())
			if action.Collate != "" {
				ctx.WriteString(" COLLATE")
				ctx.WriteString(action.Collate)
			}
		}
		if db := action.DropBehavior.String(); db != "" {
			ctx.WriteByte(' ')
			ctx.WriteString(db)
		}
	}
}

// AlterTypeAddValue represents an ALTER TYPE ADD VALUE command.
type AlterTypeAddValue struct {
	NewVal      string
	IfNotExists bool
	Placement   *AlterTypeAddValuePlacement
}

func (node *AlterTypeAddValue) Format(ctx *FmtCtx) {
	ctx.WriteString(" ADD VALUE ")
	if node.IfNotExists {
		ctx.WriteString("IF NOT EXISTS ")
	}
	lex.EncodeSQLString(&ctx.Buffer, node.NewVal)
	if node.Placement != nil {
		if node.Placement.Before {
			ctx.WriteString(" BEFORE ")
		} else {
			ctx.WriteString(" AFTER ")
		}
		lex.EncodeSQLString(&ctx.Buffer, node.Placement.ExistingVal)
	}
}

// AlterTypeAddValuePlacement represents the placement clause for an ALTER
// TYPE ADD VALUE command ([BEFORE | AFTER] value).
type AlterTypeAddValuePlacement struct {
	Before      bool
	ExistingVal string
}

// AlterTypeRenameValue represents an ALTER TYPE RENAME VALUE command.
type AlterTypeRenameValue struct {
	OldVal string
	NewVal string
}

// Format implements the NodeFormatter interface.
func (node *AlterTypeRenameValue) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME VALUE ")
	lex.EncodeSQLString(&ctx.Buffer, node.OldVal)
	ctx.WriteString(" TO ")
	lex.EncodeSQLString(&ctx.Buffer, node.NewVal)
}

// AlterTypeSetProperty represents an ALTER TYPE SET <properties> command.
type AlterTypeSetProperty struct {
	Properties BaseTypeOptions
}

// Format implements the NodeFormatter interface.
func (node *AlterTypeSetProperty) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET ( ")
	ctx.FormatNode(&node.Properties)
	ctx.WriteString(" )")
}
