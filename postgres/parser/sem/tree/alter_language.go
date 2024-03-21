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

var _ Statement = &AlterLanguage{}

// AlterLanguage represents a ALTER LANGUAGE statement.
type AlterLanguage struct {
	Name       Name
	Procedural bool
	NewName    Name
	Owner      string
}

func (node *AlterLanguage) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER ")
	if node.Procedural {
		ctx.WriteString("PROCEDURAL ")
	}
	ctx.WriteString("LANGUAGE ")
	ctx.FormatNode(&node.Name)
	if node.NewName != "" {
		ctx.WriteString(" RENAME TO ")
		ctx.FormatNode(&node.NewName)
	} else {
		ctx.WriteString(" OWNER TO ")
		ctx.WriteString(node.Owner)
	}
}
