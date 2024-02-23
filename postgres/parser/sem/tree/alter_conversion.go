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

var _ Statement = &AlterConversion{}

// AlterConversion represents a ALTER COLLATION statement.
type AlterConversion struct {
	Name           Name
	RefreshVersion bool
	Rename         Name
	Owner          string
	Schema         string
}

// Format implements the NodeFormatter interface.
func (node *AlterConversion) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER COLLATION ")
	ctx.FormatNode(&node.Name)
	if node.RefreshVersion {
		ctx.WriteString(" REFRESH VERSION")
	} else if node.Rename != "" {
		ctx.WriteString(" RENAME ")
		ctx.FormatNode(&node.Rename)
	} else if node.Owner != "" {
		ctx.WriteString(" OWNER TO ")
		ctx.FormatNameP(&node.Owner)
	} else if node.Schema != "" {
		ctx.WriteString(" SET SCHEMA ")
		ctx.FormatNameP(&node.Schema)
	}
}
