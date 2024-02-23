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

var _ Statement = &AlterFunction{}

// AlterFunction represents a ALTER FUNCTION statement.
type AlterFunction struct {
	Name      *UnresolvedObjectName
	Args      RoutineArgs
	Options   []RoutineOption
	Restrict  bool
	Rename    *UnresolvedObjectName
	Owner     string
	Schema    string
	No        bool
	Extension string
}

// Format implements the NodeFormatter interface.
func (node *AlterFunction) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER FUNCTION ")
	ctx.FormatNode(node.Name)
	if node.Args != nil {
		ctx.WriteString(" ( ")
		ctx.FormatNode(node.Name)
		ctx.WriteString(" )")
	}
	if node.Options != nil {
		for i, option := range node.Options {
			if i != 0 {
				ctx.WriteByte(' ')
			}
			ctx.FormatNode(option)
		}
		if node.Restrict {
			ctx.WriteString(" RESTRICT")
		}
	} else if node.Rename != nil {
		ctx.WriteString(" RENAME TO ")
		ctx.FormatNode(node.Rename)
	} else if node.Owner != "" {
		ctx.WriteString(" OWNER TO ")
		ctx.WriteString(node.Owner)
	} else if node.Schema != "" {
		ctx.WriteString(" SET SCHEMA ")
		ctx.WriteString(node.Schema)
	} else {
		if node.No {
			ctx.WriteString(" NO")
		}
		ctx.WriteString(" DEPENDS ON EXTENSION ")
		ctx.WriteString(node.Extension)
	}
}
