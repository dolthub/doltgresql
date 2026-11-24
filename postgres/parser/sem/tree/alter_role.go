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

var _ Statement = &AlterRole{}

// AlterRole represents an ALTER ROLE statement.
type AlterRole struct {
	Name      string
	IfExists  bool
	IsRole    bool
	KVOptions KVOptions
	// AllRoles is set for ALTER ROLE ALL, which only applies to configuration parameters.
	AllRoles bool
	// InDatabase is the database given by IN DATABASE, which only applies to configuration parameters.
	InDatabase string
	// SetVar is used for both SET and RESET of a configuration parameter. RESET sets SetVar.Reset.
	SetVar   *SetVar
	ResetAll bool
}

// Format implements the NodeFormatter interface.
func (node *AlterRole) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER")
	if node.IsRole {
		ctx.WriteString(" ROLE ")
	} else {
		ctx.WriteString(" USER ")
	}
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	if node.AllRoles {
		ctx.WriteString("ALL")
	} else {
		ctx.WriteString(node.Name)
	}
	if node.InDatabase != "" {
		ctx.WriteString(" IN DATABASE ")
		ctx.FormatNameP(&node.InDatabase)
	}

	if len(node.KVOptions) > 0 {
		ctx.WriteString(" WITH")
		node.KVOptions.formatAsRoleOptions(ctx)
	} else if node.SetVar != nil {
		ctx.WriteByte(' ')
		node.SetVar.Format(ctx)
	} else if node.ResetAll {
		ctx.WriteString(" RESET ALL")
	}
}
