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

var _ Statement = &CreateView{}

// CreateView represents a CREATE VIEW statement.
type CreateView struct {
	Name         TableName
	ColumnNames  NameList
	AsSource     *Select
	IfNotExists  bool
	Persistence  Persistence
	Replace      bool
	Materialized bool
}

// Format implements the NodeFormatter interface.
func (node *CreateView) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE ")

	if node.Replace {
		ctx.WriteString("OR REPLACE ")
	}

	if node.Persistence == PersistenceTemporary {
		ctx.WriteString("TEMPORARY ")
	}

	if node.Materialized {
		ctx.WriteString("MATERIALIZED ")
	}

	ctx.WriteString("VIEW ")

	if node.IfNotExists {
		ctx.WriteString("IF NOT EXISTS ")
	}
	ctx.FormatNode(&node.Name)

	if len(node.ColumnNames) > 0 {
		ctx.WriteByte(' ')
		ctx.WriteByte('(')
		ctx.FormatNode(&node.ColumnNames)
		ctx.WriteByte(')')
	}

	ctx.WriteString(" AS ")
	ctx.FormatNode(node.AsSource)
}

// RefreshMaterializedView represents a REFRESH MATERIALIZED VIEW statement.
type RefreshMaterializedView struct {
	Name              *UnresolvedObjectName
	Concurrently      bool
	RefreshDataOption RefreshDataOption
}

// RefreshDataOption corresponds to arguments for the REFRESH MATERIALIZED VIEW
// statement.
type RefreshDataOption int

const (
	// RefreshDataDefault refers to no option provided to the REFRESH MATERIALIZED
	// VIEW statement.
	RefreshDataDefault RefreshDataOption = iota
	// RefreshDataWithData refers to the WITH DATA option provided to the REFRESH
	// MATERIALIZED VIEW statement.
	RefreshDataWithData
	// RefreshDataClear refers to the WITH NO DATA option provided to the REFRESH
	// MATERIALIZED VIEW statement.
	RefreshDataClear
)

// Format implements the NodeFormatter interface.
func (node *RefreshMaterializedView) Format(ctx *FmtCtx) {
	ctx.WriteString("REFRESH MATERIALIZED VIEW ")
	if node.Concurrently {
		ctx.WriteString("CONCURRENTLY ")
	}
	ctx.FormatNode(node.Name)
	switch node.RefreshDataOption {
	case RefreshDataWithData:
		ctx.WriteString(" WITH DATA")
	case RefreshDataClear:
		ctx.WriteString(" WITH NO DATA")
	}
}
