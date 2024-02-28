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

var _ Statement = &AlterMaterializedView{}

// AlterMaterializedView represents an ALTER MATERIALIZED VIEW statement.
type AlterMaterializedView struct {
	Name      *UnresolvedObjectName
	IfExists  bool
	Cmds      AlterTableCmds
	No        bool
	Extension string
}

func (node *AlterMaterializedView) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER MATERIALIZED VIEW ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(node.Name)
	if node.Extension != "" {
		if node.No {
			ctx.WriteString(" NO")
		}
		ctx.WriteString(" DEPENDS ON EXTENSION ")
		ctx.WriteString(node.Extension)
	} else {
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.Cmds)
	}
}

// The following statements are included as part of ALTER TABLE ... or RENAME TABLE statement:
//	ALTER MATERIALIZED VIEW [ IF EXISTS ] name RENAME TO new_name
//	ALTER MATERIALIZED VIEW [ IF EXISTS ] name SET SCHEMA new_schema
//	ALTER MATERIALIZED VIEW ALL IN TABLESPACE name [ OWNED BY role_name [, ... ] ] SET TABLESPACE new_tablespace [ NOWAIT ]
