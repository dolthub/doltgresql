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

package node

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/types"
)

// CreateType handles the CREATE TYPE statement.
type CreateType struct {
	SchemaName string
	Name       string

	// composite type
	AsTypes []CompositeAsType

	// enum type
	Labels []string

	typType types.TypeType
}

// CompositeAsType represents an attribute name
// and data type for a composite type.
type CompositeAsType struct {
	AttrName  string
	Typ       *types.DoltgresType
	Collation string
}

var _ sql.ExecSourceRel = (*CreateType)(nil)
var _ vitess.Injectable = (*CreateType)(nil)

// NewCreateCompositeType creates CreateType node for creating COMPOSITE type.
func NewCreateCompositeType(schema, name string, typs []CompositeAsType) *CreateType {
	return &CreateType{SchemaName: schema, Name: name, AsTypes: typs, typType: types.TypeType_Composite}
}

// NewCreateEnumType creates CreateType node for creating ENUM type.
func NewCreateEnumType(schema, name string, labels []string) *CreateType {
	return &CreateType{SchemaName: schema, Name: name, Labels: labels, typType: types.TypeType_Enum}
}

// NewCreateShellType creates CreateType node for creating
// a placeholder for a type to be defined later.
func NewCreateShellType(schema, name string) *CreateType {
	return &CreateType{SchemaName: schema, Name: name, typType: types.TypeType_Pseudo}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateType) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateType) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateType) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateType) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	var userRole auth.Role
	auth.LockRead(func() {
		userRole = auth.GetRole(ctx.Client().User)
	})
	if !userRole.IsValid() {
		return nil, fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
	}

	schema, err := core.GetSchemaName(ctx, nil, c.SchemaName)
	if err != nil {
		return nil, err
	}
	collection, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if collection.HasType(c.SchemaName, c.Name) {
		// TODO: if the existing type is array type, it updates the array type name and creates the new type.
		return nil, types.ErrTypeAlreadyExists.New(c.Name)
	}

	var newType *types.DoltgresType
	switch c.typType {
	case types.TypeType_Pseudo:
		newType = types.NewShellType(ctx, id.NewType(c.SchemaName, c.Name))
	case types.TypeType_Enum:
		typeID := id.NewType(c.SchemaName, c.Name)
		arrayID := id.NewType(c.SchemaName, "_"+c.Name)
		enumLabelMap := make(map[string]types.EnumLabel)
		for i, l := range c.Labels {
			if _, ok := enumLabelMap[l]; ok {
				// DETAIL:  Key (enumtypid, enumlabel)=(16702, ok) already exists.
				return nil, fmt.Errorf(`duplicate key value violates unique constraint "pg_enum_typid_label_index"`)
			}
			labelID := id.NewEnumLabel(typeID, l)
			el := types.NewEnumLabel(ctx, labelID, float32(i+1))
			enumLabelMap[l] = el
		}
		newType = types.NewEnumType(ctx, arrayID, typeID, enumLabelMap)
		// TODO: store labels somewhere
	case types.TypeType_Composite:
		typeID := id.NewType(c.SchemaName, c.Name)
		arrayID := id.NewType(c.SchemaName, "_"+c.Name)

		relID := id.Null // TODO: create relation with c.AsTypes
		attrs := make([]types.CompositeAttribute, len(c.AsTypes))
		for i, a := range c.AsTypes {
			attrs[i] = types.NewCompositeAttribute(ctx, relID, a.AttrName, a.Typ.ID, int16(i+1), a.Collation)
		}
		newType = types.NewCompositeType(ctx, relID, arrayID, typeID, attrs)
	default:
		return nil, fmt.Errorf("create type as %s is not supported", c.typType)
	}

	err = collection.CreateType(schema, newType)
	if err != nil {
		return nil, err
	}

	// create array type for defined types
	if newType.IsDefined {
		arrayType := types.CreateArrayTypeFromBaseType(newType)
		err = collection.CreateType(schema, arrayType)
		if err != nil {
			return nil, err
		}
	}

	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateType) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateType) String() string {
	return "CREATE TYPE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateType) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateType) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
