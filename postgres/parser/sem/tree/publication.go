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

// CreatePublication represents the supported empty and FOR ALL TABLES publication forms.
type CreatePublication struct {
	Name      Name
	AllTables bool
	Options   KVOptions
}

// Format implements NodeFormatter.
func (n *CreatePublication) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE PUBLICATION ")
	ctx.FormatNode(&n.Name)
	if n.AllTables {
		ctx.WriteString(" FOR ALL TABLES")
	}
	if len(n.Options) > 0 {
		ctx.WriteString(" WITH (")
		ctx.FormatNode(&n.Options)
		ctx.WriteByte(')')
	}
}

// String implements fmt.Stringer.
func (n *CreatePublication) String() string { return AsString(n) }

// StatementType implements Statement.
func (*CreatePublication) StatementType() StatementType { return DDL }

// StatementTag implements Statement.
func (*CreatePublication) StatementTag() string { return "CREATE PUBLICATION" }

// DropPublication represents DROP PUBLICATION. Publications have no dependent objects,
// so CASCADE and RESTRICT have the same effect.
type DropPublication struct {
	Names        NameList
	IfExists     bool
	DropBehavior DropBehavior
}

// Format implements NodeFormatter.
func (n *DropPublication) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP PUBLICATION ")
	if n.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&n.Names)
	if n.DropBehavior != DropDefault {
		ctx.WriteByte(' ')
		ctx.WriteString(n.DropBehavior.String())
	}
}

// String implements fmt.Stringer.
func (n *DropPublication) String() string { return AsString(n) }

// StatementType implements Statement.
func (*DropPublication) StatementType() StatementType { return DDL }

// StatementTag implements Statement.
func (*DropPublication) StatementTag() string { return "DROP PUBLICATION" }

var _ Statement = (*CreatePublication)(nil)
var _ Statement = (*DropPublication)(nil)
