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
	"strings"

	"github.com/dolthub/doltgresql/postgres/parser/privilege"
)

var _ Statement = &Revoke{}

// Revoke represents a REVOKE statement.
// PrivilegeList and TargetList are defined in grant.go
type Revoke struct {
	Privileges     privilege.List
	Targets        TargetList
	Grantees       []string
	GrantOptionFor bool
	GrantedBy      string
	DropBehavior   DropBehavior

	// only used for table target with column names defined
	PrivsWithCols []PrivForCols
}

// Format implements the NodeFormatter interface.
func (node *Revoke) Format(ctx *FmtCtx) {
	ctx.WriteString("REVOKE ")
	if node.GrantOptionFor {
		ctx.WriteString(" GRANT OPTION FOR ")
	}
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
	ctx.WriteString(" FROM ")
	ctx.WriteString(strings.Join(node.Grantees, ", "))
	if node.GrantedBy != "" {
		ctx.WriteString(" GRANTED BY ")
		ctx.WriteString(node.GrantedBy)
	}
	if node.DropBehavior.String() != "" {
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	}
}

var _ Statement = &RevokeRole{}

// RevokeRole represents a REVOKE <role> statement.
type RevokeRole struct {
	Roles        NameList
	Members      []string
	Option       string
	GrantedBy    string
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *RevokeRole) Format(ctx *FmtCtx) {
	ctx.WriteString("REVOKE ")
	if node.Option != "" {
		ctx.WriteString(node.Option)
		ctx.WriteString(" OPTION FOR ")
	}
	ctx.FormatNode(&node.Roles)
	ctx.WriteString(" FROM ")
	ctx.WriteString(strings.Join(node.Members, ", "))
	if node.GrantedBy != "" {
		ctx.WriteString(" GRANTED BY ")
		ctx.WriteString(node.GrantedBy)
	}
	if node.DropBehavior.String() != "" {
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	}
}
