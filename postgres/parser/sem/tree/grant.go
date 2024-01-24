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

import (
	"fmt"
	"github.com/dolthub/doltgresql/postgres/parser/privilege"
	"strings"
)

// Grant represents a GRANT statement.
type Grant struct {
	Privileges      privilege.List
	Targets         TargetList
	Grantees        []string
	WithGrantOption bool
	GrantedBy       string

	// only used for table target with column names defined
	PrivsWithCols []PrivForCols
}

// PrivForCols is only used for table target with column names defined
type PrivForCols struct {
	Privilege privilege.Kind
	ColNames  NameList
}

// Format implements the NodeFormatter interface.
func (node *Grant) Format(ctx *FmtCtx) {
	ctx.WriteString("GRANT ")
	if node.PrivsWithCols != nil {
		for i, p := range node.PrivsWithCols {
			if i != 0 {
				ctx.WriteString(", ")
			}
			ctx.WriteString(p.Privilege.String())
			ctx.WriteString(" ( ")
			ctx.FormatNode(&p.ColNames)
			ctx.WriteString(" )")
		}
	} else {
		node.Privileges.Format(&ctx.Buffer)
	}
	ctx.WriteString(" ON ")
	ctx.FormatNode(&node.Targets)
	ctx.WriteString(" TO ")
	ctx.WriteString(strings.Join(node.Grantees, ", "))
	if node.WithGrantOption {
		ctx.WriteString(" WITH GRANT OPTION")
	}
	if node.GrantedBy != "" {
		ctx.WriteString(" GRANTED BY ")
		ctx.WriteString(node.GrantedBy)
	}
}

// TargetList represents a list of targets.
// Only one field may be non-nil.
type TargetList struct {
	TargetType privilege.ObjectType
	InSchema   []string // used with ALL only

	Tables           TablePatterns
	TableColumnNames NameList
	Sequences        NameList
	Databases        NameList
	LargeObjects     []Expr
	Routines         []Routine
	Parameters       NameList
	Types            []*UnresolvedObjectName
	// domains, foreign data wrappers, foreign servers, languages,
	// parameters, schemas, tablespaces are using string for name definition
	Names []string

	// ForRoles and Roles are used internally in the parser and not used
	// in the AST. Therefore they do not participate in pretty-printing,
	// etc.
	ForRoles bool
	Roles    NameList
}

// Routine used for { FUNCTION | PROCEDURE | ROUTINE }
type Routine struct {
	Name Name
	Args *AggregateArg
}

// Format implements the NodeFormatter interface.
func (tl *TargetList) Format(ctx *FmtCtx) {
	switch tl.TargetType {
	case privilege.Table:
		if tl.Tables == nil {
			ctx.WriteString("ALL TABLES ")
			if tl.InSchema != nil {
				ctx.WriteString("IN SCHEMA ")
				ctx.WriteString(strings.Join(tl.InSchema, ", "))
			}
		} else {
			ctx.WriteString("TABLE ")
			ctx.FormatNode(&tl.Tables)
		}
	case privilege.Sequence:
		if tl.Sequences == nil {
			ctx.WriteString("ALL SEQUENCES ")
			if tl.InSchema != nil {
				ctx.WriteString("IN SCHEMA ")
				ctx.WriteString(strings.Join(tl.InSchema, ", "))
			}
		} else {
			ctx.WriteString("SEQUENCE ")
			ctx.FormatNode(&tl.Sequences)
		}
	case privilege.Database:
		ctx.WriteString("DATABASE ")
		ctx.FormatNode(&tl.Databases)
	case privilege.Function, privilege.Procedure, privilege.Routine:
		t := strings.ToUpper(string(tl.TargetType))
		if tl.Routines == nil {
			ctx.WriteString(fmt.Sprintf("ALL %sS", t))
			if tl.InSchema != nil {
				ctx.WriteString(" IN SCHEMA ")
				ctx.WriteString(strings.Join(tl.InSchema, ", "))
			}
		} else {
			ctx.WriteString(t)
			ctx.WriteByte(' ')
			for i, r := range tl.Routines {
				if i != 0 {
					ctx.WriteString(", ")
				}
				ctx.FormatNode(&r.Name)
				if r.Args != nil {
					ctx.WriteString(" ( ")
					ctx.FormatNode(r.Args)
					ctx.WriteString(" )")
				}
			}
		}
	case privilege.LargeObject:
		ctx.WriteString("LARGE OBJECT ")
		for i, lo := range tl.LargeObjects {
			if i != 0 {
				ctx.WriteString(", ")
			}
			ctx.FormatNode(lo)
		}
	case privilege.Type:
		ctx.WriteString("TYPE ")
		for i, typ := range tl.Types {
			if i != 0 {
				ctx.WriteString(", ")
			}
			ctx.FormatNode(typ)
		}
	default:
		t := strings.ToUpper(string(tl.TargetType))
		ctx.WriteString(t)
		ctx.WriteByte(' ')
		ctx.WriteString(strings.Join(tl.Names, ", "))
	}
}

// GrantRole represents a GRANT <role> statement.
type GrantRole struct {
	Roles      NameList
	Members    []string
	WithOption string
	GrantedBy  string
}

// Format implements the NodeFormatter interface.
func (node *GrantRole) Format(ctx *FmtCtx) {
	ctx.WriteString("GRANT ")
	ctx.FormatNode(&node.Roles)
	ctx.WriteString(" TO ")
	ctx.WriteString(strings.Join(node.Members, ", "))
	if node.WithOption != "" {
		ctx.WriteString(" WITH ")
		ctx.WriteString(node.WithOption)
	}
	if node.GrantedBy != "" {
		ctx.WriteString(" GRANTED BY ")
		ctx.WriteString(node.GrantedBy)
	}
}
