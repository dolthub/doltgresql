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

// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import "strings"

// AlterIndex represents an ALTER INDEX statement.
type AlterIndex struct {
	IfExists bool
	Index    TableIndexName
	Cmd      AlterIndexCmd
}

var _ Statement = &AlterIndex{}

// Format implements the NodeFormatter interface.
func (node *AlterIndex) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER INDEX ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Index)
	ctx.FormatNode(node.Cmd)
}

var _ Statement = &AlterIndexAllInTablespace{}

// AlterIndexAllInTablespace represents an ALTER INDEX ALL IN TABLESPACE statement.
type AlterIndexAllInTablespace struct {
	Name       Name
	OwnedBy    []string
	Tablespace string
	NoWait     bool
}

// Format implements the NodeFormatter interface.
func (node *AlterIndexAllInTablespace) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER INDEX ALL IN TABLESPACE ")
	ctx.FormatNode(&node.Name)
	if node.OwnedBy != nil {
		ctx.WriteString(" OWNED BY ")
		ctx.WriteString(strings.Join(node.OwnedBy, ", "))
	}
	ctx.WriteString(" SET TABLESPACE ")
	ctx.WriteString(node.Tablespace)
	if node.NoWait {
		ctx.WriteString(" NOWAIT")
	}
}

// AlterIndexCmd represents an index modification operation.
type AlterIndexCmd interface {
	NodeFormatter
	// Placeholder function to ensure that only desired types
	// (AlterIndex*) conform to the AlterIndexCmd interface.
	alterIndexCmd()
}

func (*AlterIndexAttachPartition) alterIndexCmd() {}
func (*AlterIndexExtension) alterIndexCmd()       {}
func (*AlterIndexSetStatistics) alterIndexCmd()   {}
func (*AlterIndexSetStorage) alterIndexCmd()      {}
func (*AlterIndexSetTablespace) alterIndexCmd()   {}

var _ AlterIndexCmd = &AlterIndexAttachPartition{}
var _ AlterIndexCmd = &AlterIndexExtension{}
var _ AlterIndexCmd = &AlterIndexSetStatistics{}
var _ AlterIndexCmd = &AlterIndexSetStorage{}
var _ AlterIndexCmd = &AlterIndexSetTablespace{}

// AlterIndexAttachPartition represents an ALTER INDEX ... ATTACH PARTITION statement.
type AlterIndexAttachPartition struct {
	Index UnrestrictedName
}

// Format implements the NodeFormatter interface.
func (node *AlterIndexAttachPartition) Format(ctx *FmtCtx) {
	ctx.WriteString(" ATTACH PARTITION ")
	ctx.FormatNode(&node.Index)
}

// AlterIndexExtension represents an ALTER INDEX ... [NO] DEPENDS ON EXTENSION statement.
type AlterIndexExtension struct {
	No        bool
	Extension string
}

// Format implements the NodeFormatter interface.
func (node *AlterIndexExtension) Format(ctx *FmtCtx) {
	if node.No {
		ctx.WriteString(" NO")
	}
	ctx.WriteString(" DEPENDS ON EXTENSION ")
	ctx.WriteString(node.Extension)
}

// AlterIndexSetStatistics represents an ALTER INDEX ... SET TABLESPACE
// command.
type AlterIndexSetStatistics struct {
	ColumnIdx Expr
	Stats     Expr
}

// Format implements the NodeFormatter interface.
func (node *AlterIndexSetStatistics) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(node.ColumnIdx)
	ctx.WriteString(" SET STATISTICS ")
	ctx.FormatNode(node.Stats)
}

// AlterIndexSetStorage represents an ALTER INDEX ... SET TABLESPACE
// command.
type AlterIndexSetStorage struct {
	Params  StorageParams
	IsReset bool
}

// Format implements the NodeFormatter interface.
func (node *AlterIndexSetStorage) Format(ctx *FmtCtx) {
	if node.IsReset {
		ctx.WriteString(" RESET ( ")
	} else {
		ctx.WriteString(" SET ( ")
	}
	ctx.FormatNode(&node.Params)
	ctx.WriteString(" )")
}

// AlterIndexSetTablespace represents an ALTER INDEX ... SET TABLESPACE
// command.
type AlterIndexSetTablespace struct {
	Tablespace string
}

// Format implements the NodeFormatter interface.
func (node *AlterIndexSetTablespace) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET TABLESPACE ")
	ctx.WriteString(node.Tablespace)
}
