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

var _ Statement = &AlterDefaultPrivileges{}

// AlterDefaultPrivileges represents a ALTER DEFAULT PRIVILEGES statement.
type AlterDefaultPrivileges struct {
	ForRole     bool
	TargetRoles NameList
	InSchemas   NameList
	Grant       *Grant
	Revoke      *Revoke
}

// Format implements the NodeFormatter interface.
func (node *AlterDefaultPrivileges) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER DEFAULT PRIVILEGES ")
	if len(node.TargetRoles) > 0 {
		ctx.WriteString("FOR ")
		if node.ForRole {
			ctx.WriteString("ROLE ")
		} else {
			ctx.WriteString("USER ")
		}
		ctx.FormatNode(&node.TargetRoles)
	}
	if len(node.InSchemas) > 0 {
		ctx.WriteString("IN SCHEMAS ")
		ctx.FormatNode(&node.InSchemas)
	}
	ctx.WriteByte(' ')
	if node.Grant != nil {
		node.Grant.Format(ctx)
	} else {
		node.Revoke.Format(ctx)
	}
}
