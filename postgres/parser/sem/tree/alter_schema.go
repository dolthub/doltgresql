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

var _ Statement = &AlterSchema{}

// AlterSchema represents an ALTER SCHEMA statement.
type AlterSchema struct {
	Schema string
	Cmd    AlterSchemaCmd
}

// Format implements the NodeFormatter interface.
func (node *AlterSchema) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER SCHEMA ")
	ctx.FormatNameP(&node.Schema)
	ctx.FormatNode(node.Cmd)
}

// AlterSchemaCmd represents a schema modification operation.
type AlterSchemaCmd interface {
	NodeFormatter
	alterSchemaCmd()
}

func (*AlterSchemaRename) alterSchemaCmd() {}

// AlterSchemaRename represents an ALTER SCHEMA RENAME command.
type AlterSchemaRename struct {
	NewName string
}

// Format implements the NodeFormatter interface.
func (node *AlterSchemaRename) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME TO ")
	ctx.FormatNameP(&node.NewName)
}

func (*AlterSchemaOwner) alterSchemaCmd() {}

// AlterSchemaOwner represents an ALTER SCHEMA RENAME command.
type AlterSchemaOwner struct {
	Owner string
}

// Format implements the NodeFormatter interface.
func (node *AlterSchemaOwner) Format(ctx *FmtCtx) {
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNameP(&node.Owner)
}
