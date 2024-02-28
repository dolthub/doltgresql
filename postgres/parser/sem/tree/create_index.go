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

type IndexElemOpClass struct {
	Name    string
	Options []IndexElemOpClassOption
}

type IndexElemOpClassOption struct {
	Param string
	Val   Expr
}

// Format implements the NodeFormatter interface.
func (node *IndexElemOpClassOption) Format(ctx *FmtCtx) {
	ctx.WriteString(node.Param)
	ctx.WriteString(" = ")
	ctx.FormatNode(node.Val)
}

// IndexElem represents a column with a direction in a CREATE INDEX statement.
type IndexElem struct {
	Column     Name
	Expr       Expr // in parentheses or function name
	Collation  string
	OpClass    *IndexElemOpClass
	Direction  Direction
	NullsOrder NullsOrder
	// only used for EXCLUDE clause
	ExcludeOp Operator
}

// Format implements the NodeFormatter interface.
func (node *IndexElem) Format(ctx *FmtCtx) {
	if node.Column != "" {
		ctx.FormatNode(&node.Column)
	} else {
		ctx.FormatNode(node.Expr)
	}
	if node.Collation != "" {
		ctx.WriteString(" COLLATE ")
		ctx.WriteString(node.Collation)
	}
	if node.OpClass != nil {
		ctx.WriteByte(' ')
		ctx.WriteString(node.OpClass.Name)
		if len(node.OpClass.Options) != 0 {
			ctx.WriteString(" (")
			for i, option := range node.OpClass.Options {
				if i != 0 {
					ctx.WriteString(", ")
				}
				ctx.FormatNode(&option)
			}
		}
	}
	if node.Direction != DefaultDirection {
		ctx.WriteByte(' ')
		ctx.WriteString(node.Direction.String())
	}
	if node.NullsOrder != DefaultNullsOrder {
		ctx.WriteByte(' ')
		ctx.WriteString(node.NullsOrder.String())
	}
	if node.ExcludeOp != nil {
		ctx.WriteString(" WITH ")
		switch op := node.ExcludeOp.(type) {
		case UnaryOperator:
			ctx.WriteString(op.String())
		case BinaryOperator:
			ctx.WriteString(op.String())
		case ComparisonOperator:
			ctx.WriteString(op.String())
		}
	}
}

// IndexElemList is list of IndexElem.
type IndexElemList []IndexElem

// Format pretty-prints the contained names separated by commas.
// Format implements the NodeFormatter interface.
func (l *IndexElemList) Format(ctx *FmtCtx) {
	for i := range *l {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&(*l)[i])
	}
}

var _ Statement = &CreateIndex{}

// CreateIndex represents a CREATE INDEX statement.
type CreateIndex struct {
	Name          Name
	Table         TableName
	Unique        bool
	Concurrently  bool
	IfNotExists   bool
	Only          bool
	Using         string
	Columns       IndexElemList
	NullsDistinct bool
	IndexParams   IndexParams
	Predicate     Expr
}

// Format implements the NodeFormatter interface.
func (node *CreateIndex) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE ")
	if node.Unique {
		ctx.WriteString("UNIQUE ")
	}
	ctx.WriteString("INDEX ")
	if node.Concurrently {
		ctx.WriteString("CONCURRENTLY ")
	}
	if node.IfNotExists {
		ctx.WriteString("IF NOT EXISTS ")
	}
	if node.Name != "" {
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	}
	ctx.WriteString("ON ")
	if node.Only {
		ctx.WriteString("ONLY ")
	}
	ctx.FormatNode(&node.Table)
	if node.Using != "" {
		ctx.WriteString(" USING ")
		ctx.WriteString(node.Using)
	}
	ctx.WriteString(" ( ")
	ctx.FormatNode(&node.Columns)
	ctx.WriteString(" )")
	if node.IndexParams.IncludeColumns != nil {
		ctx.WriteString(" INCLUDE ( ")
		ctx.FormatNode(&node.IndexParams.IncludeColumns)
		ctx.WriteString(" )")
	}
	if node.NullsDistinct {
		ctx.WriteString(" NULLS DISTINCT ")
	} else {
		ctx.WriteString(" NULLS NOT DISTINCT ")
	}
	if node.IndexParams.StorageParams != nil {
		ctx.WriteString(" WITH (")
		ctx.FormatNode(&node.IndexParams.StorageParams)
		ctx.WriteString(")")
	}
	if node.IndexParams.Tablespace != "" {
		ctx.WriteString(" TABLESPACE ")
		ctx.FormatNode(&node.IndexParams.Tablespace)
	}
	if node.Predicate != nil {
		ctx.WriteString(" WHERE ")
		ctx.FormatNode(node.Predicate)
	}
}
