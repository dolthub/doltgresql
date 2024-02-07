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

// Copyright 2018 The Cockroach Authors.
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

// TODO: maybe combine all of COMMENT ON ... objects below into single CommentOn interface.

// CommentOnColumn represents an COMMENT ON COLUMN statement.
type CommentOnColumn struct {
	*ColumnItem
	Comment *string
}

// Format implements the NodeFormatter interface.
func (n *CommentOnColumn) Format(ctx *FmtCtx) {
	ctx.WriteString("COMMENT ON COLUMN ")
	ctx.FormatNode(n.ColumnItem)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
	} else {
		ctx.WriteString("NULL")
	}
}

// CommentOnDatabase represents an COMMENT ON DATABASE statement.
type CommentOnDatabase struct {
	Name    Name
	Comment *string
}

// Format implements the NodeFormatter interface.
func (n *CommentOnDatabase) Format(ctx *FmtCtx) {
	ctx.WriteString("COMMENT ON DATABASE ")
	ctx.FormatNode(&n.Name)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
	} else {
		ctx.WriteString("NULL")
	}
}

// CommentOnExtension represents an COMMENT ON EXTENSION statement.
type CommentOnExtension struct {
	Name    Name
	Comment *string
}

// Format implements the NodeFormatter interface.
func (n *CommentOnExtension) Format(ctx *FmtCtx) {
	ctx.WriteString("COMMENT ON EXTENSION ")
	ctx.FormatNode(&n.Name)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
	} else {
		ctx.WriteString("NULL")
	}
}

// CommentOnIndex represents a COMMENT ON INDEX statement.
type CommentOnIndex struct {
	Index   TableIndexName
	Comment *string
}

// Format implements the NodeFormatter interface.
func (n *CommentOnIndex) Format(ctx *FmtCtx) {
	ctx.WriteString("COMMENT ON INDEX ")
	ctx.FormatNode(&n.Index)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
	} else {
		ctx.WriteString("NULL")
	}
}

// CommentOnTable represents an COMMENT ON TABLE statement.
type CommentOnTable struct {
	Table   *UnresolvedObjectName
	Comment *string
}

// Format implements the NodeFormatter interface.
func (n *CommentOnTable) Format(ctx *FmtCtx) {
	ctx.WriteString("COMMENT ON TABLE ")
	ctx.FormatNode(n.Table)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
	} else {
		ctx.WriteString("NULL")
	}
}
