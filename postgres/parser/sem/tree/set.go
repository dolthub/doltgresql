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

var _ Statement = &SetVar{}

// SetVar represents a SET or RESET <configuration_param> statement.
type SetVar struct {
	IsLocal bool
	Name    string
	Values  Exprs
	// FromCurrent is used for SET clauses in CREATE statements only.
	FromCurrent bool
}

func (node *SetVar) SetLocalSetStmt() {
	node.IsLocal = true
}

// Format implements the NodeFormatter interface.
func (node *SetVar) Format(ctx *FmtCtx) {
	ctx.WriteString("SET ")
	if node.Name == "" {
		ctx.WriteString("ROW (")
		ctx.FormatNode(&node.Values)
		ctx.WriteString(")")
	} else {
		ctx.WithFlags(ctx.flags & ^FmtAnonymize, func() {
			// Session var names never contain PII and should be distinguished
			// for feature tracking purposes.
			ctx.FormatNameP(&node.Name)
		})

		if node.FromCurrent {
			ctx.WriteString(" FROM CURRENT")
		} else {
			ctx.WriteString(" = ")
			ctx.FormatNode(&node.Values)
		}
	}
}

var _ Statement = &SetSessionAuthorization{}

// SetSessionAuthorization represents a SET SESSION AUTHORIZATION ... statement.
type SetSessionAuthorization struct {
	Username string
	IsLocal  bool
}

// Format implements the NodeFormatter interface.
func (node *SetSessionAuthorization) Format(ctx *FmtCtx) {
	if node.Username == "" {
		// equivalent to RESET SESSION AUTHORIZATION
		ctx.WriteString("SET SESSION AUTHORIZATION DEFAULT")
	} else {
		ctx.WriteString("SET SESSION AUTHORIZATION ")
		ctx.WriteString(node.Username)
	}
}

func (node *SetSessionAuthorization) SetLocalSetStmt() {
	node.IsLocal = true
}

var _ Statement = &SetRole{}

// SetRole represents a SET ROLE ... and RESET ROLE statements.
type SetRole struct {
	IsLocal bool
	Name    string
	None    bool
	Reset   bool
}

// Format implements the NodeFormatter interface.
func (node *SetRole) Format(ctx *FmtCtx) {
	if node.Reset {
		ctx.WriteString("RESET ROLE")
	} else {
		ctx.WriteString("SET ")
		if node.IsLocal {
			ctx.WriteString("LOCAL")
		} else {
			ctx.WriteString("SESSION")
		}
		ctx.WriteString(" ROLE")
		if node.None {
			ctx.WriteString(" NONE")
		} else {
			ctx.WriteString(node.Name)
		}
	}
}

func (node *SetRole) SetLocalSetStmt() {
	node.IsLocal = true
}

// SetStmt represents a set statement that has [ SESSION | LOCAL ] setting option.
// It is used to set the |node.IsLocal| to true if LOCAL clause is defined.
type SetStmt interface {
	NodeFormatter
	SetLocalSetStmt()
}

var _ SetStmt = &SetVar{}
var _ SetStmt = &SetRole{}
var _ SetStmt = &SetSessionAuthorization{}

// The SET statements below here do not have [ SESSION | LOCAL ] setting.

var _ Statement = &SetConstraints{}

// SetConstraints represents a SET CONSTRAINTS statement.
type SetConstraints struct {
	Names    NameList
	All      bool
	Deferred bool
}

// Format implements the NodeFormatter interface.
func (node *SetConstraints) Format(ctx *FmtCtx) {
	ctx.WriteString("SET CONSTRAINTS ")
	if node.All {
		ctx.WriteString("ALL")
	} else {
		ctx.FormatNode(&node.Names)
	}
	if node.Deferred {
		ctx.WriteString(" DEFERRED")
	} else {
		ctx.WriteString(" IMMEDIATE")
	}
}

var _ Statement = &SetTransaction{}

// SetTransaction represents a SET TRANSACTION statement.
type SetTransaction struct {
	Modes TransactionModes
}

// Format implements the NodeFormatter interface.
func (node *SetTransaction) Format(ctx *FmtCtx) {
	ctx.WriteString("SET TRANSACTION")
	node.Modes.Format(ctx)
}

// SetSessionCharacteristics represents a SET SESSION CHARACTERISTICS AS TRANSACTION statement.
type SetSessionCharacteristics struct {
	Modes TransactionModes
}

// Format implements the NodeFormatter interface.
func (node *SetSessionCharacteristics) Format(ctx *FmtCtx) {
	ctx.WriteString("SET SESSION CHARACTERISTICS AS TRANSACTION")
	node.Modes.Format(ctx)
}

var _ Statement = &ResetAll{}

type ResetAll struct{}

// Format implements the NodeFormatter interface.
func (node *ResetAll) Format(ctx *FmtCtx) {
	ctx.WriteString("RESET ALL")
}
