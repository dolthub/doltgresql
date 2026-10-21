// Copyright 2026 Dolthub, Inc.
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

var _ Statement = &AlterExtension{}

// AlterExtension represents an ALTER EXTENSION ... UPDATE statement. Version is empty when no version is given.
type AlterExtension struct {
	Name    Name
	Version string
}

// Format implements the NodeFormatter interface.
func (node *AlterExtension) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER EXTENSION ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" UPDATE")
	if node.Version != "" {
		ctx.WriteString(" TO ")
		ctx.FormatNameP(&node.Version)
	}
}
