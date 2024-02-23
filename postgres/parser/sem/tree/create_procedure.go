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

package tree

var _ Statement = &CreateProcedure{}

// CreateProcedure represents a CREATE PROCEDURE statement.
type CreateProcedure struct {
	Name    *UnresolvedObjectName
	Replace bool
	Args    RoutineArgs
	Options []RoutineOption
}

// Format implements the NodeFormatter interface.
func (node *CreateProcedure) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE ")
	if node.Replace {
		ctx.WriteString("OR REPLACE ")
	}
	ctx.WriteString("PROCEDURE ")
	ctx.FormatNode(node.Name)
	if len(node.Args) != 0 {
		ctx.WriteString(" (")
		ctx.FormatNode(node.Args)
		ctx.WriteString(" )")
	}
	for i, option := range node.Options {
		if i != 0 {
			ctx.WriteByte(' ')
		}
		ctx.FormatNode(option)
	}
}
