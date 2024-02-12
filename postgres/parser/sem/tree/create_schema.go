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

// CreateSchema represents a CREATE SCHEMA statement.
type CreateSchema struct {
	IfNotExists    bool
	Schema         string
	AuthRole       string
	SchemaElements []Statement
}

// Format implements the NodeFormatter interface.
func (node *CreateSchema) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE SCHEMA")

	if node.IfNotExists {
		ctx.WriteString(" IF NOT EXISTS")
	}

	if node.Schema != "" {
		ctx.WriteString(" ")
		ctx.WriteString(node.Schema)
	}

	if node.AuthRole != "" {
		ctx.WriteString(" AUTHORIZATION ")
		ctx.WriteString(node.AuthRole)
	}

	if node.SchemaElements != nil {
		for _, se := range node.SchemaElements {
			ctx.WriteByte(' ')
			ctx.FormatNode(se)
		}
	}
}
