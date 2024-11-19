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
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/types"
)

// CreateDomain handles the CREATE DOMAIN statement.
type CreateDomain struct {
	SchemaName           string
	Name                 string
	AsType               types.DoltgresType
	Collation            string
	HasDefault           bool
	DefaultExpr          sql.Expression
	IsNotNull            bool
	CheckConstraintNames []string
	CheckConstraints     sql.CheckConstraints
}

var _ sql.ExecSourceRel = (*CreateDomain)(nil)
var _ vitess.Injectable = (*CreateDomain)(nil)

// Children implements the interface sql.ExecSourceRel.
func (c *CreateDomain) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateDomain) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateDomain) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateDomain) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	var userRole auth.Role
	auth.LockRead(func() {
		userRole = auth.GetRole(ctx.Client().User)
	})
	if !userRole.IsValid() {
		return nil, fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
	}

	// TODO: create array type with this type as base type?
	var defExpr string
	if c.DefaultExpr != nil {
		defExpr = c.DefaultExpr.String()
	}
	checkDefs := make([]*sql.CheckDefinition, len(c.CheckConstraints))
	var err error
	for i, check := range c.CheckConstraints {
		checkDefs[i], err = plan.NewCheckDefinition(ctx, check)
		if err != nil {
			return nil, err
		}
	}

	newType, err := types.NewDomainType(ctx, c.SchemaName, c.Name, c.AsType, defExpr, c.IsNotNull, checkDefs, "")
	if err != nil {
		return nil, err
	}
	schema, err := core.GetSchemaName(ctx, nil, c.SchemaName)
	if err != nil {
		return nil, err
	}
	collection, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	err = collection.CreateType(schema, newType)
	if err != nil {
		return nil, err
	}

	auth.LockWrite(func() {
		auth.AddOwner(auth.OwnershipKey{
			PrivilegeObject: auth.PrivilegeObject_DOMAIN,
			Schema:          schema,
			Name:            c.Name,
		}, userRole.ID())
		err = auth.PersistChanges()
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateDomain) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateDomain) String() string {
	return "CREATE DOMAIN"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateDomain) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateDomain) WithResolvedChildren(children []any) (any, error) {
	checksStartAt := 0
	var defExpr sql.Expression
	if c.HasDefault {
		expr, ok := children[0].(sql.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid vitess child, expected sql.Expression for Default value but got %t", children[0])
		}
		defExpr = expr
		checksStartAt = 1
	}
	var checks sql.CheckConstraints
	for i, child := range children[checksStartAt:] {
		expr, ok := child.(sql.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid vitess child, expected sql.Expression for Check constraint value but got %t", children[0])
		}
		checks = append(checks, &sql.CheckConstraint{
			Name:     c.CheckConstraintNames[i],
			Expr:     expr,
			Enforced: true,
		})
	}
	return &CreateDomain{
		SchemaName:           c.SchemaName,
		Name:                 c.Name,
		AsType:               c.AsType,
		Collation:            c.Collation,
		HasDefault:           c.HasDefault,
		DefaultExpr:          defExpr,
		IsNotNull:            c.IsNotNull,
		CheckConstraintNames: c.CheckConstraintNames,
		CheckConstraints:     checks,
	}, nil
}

// DomainColumn represents the column name `VALUE.
// It is a placeholder column reference later
// used for column defined as the domain type.
type DomainColumn struct {
	Typ types.DoltgresType
}

var _ vitess.Injectable = (*DomainColumn)(nil)
var _ sql.Expression = (*DomainColumn)(nil)

// Resolved implements the interface sql.Expression.
func (d *DomainColumn) Resolved() bool {
	return true
}

// String implements the interface sql.Expression.
func (d *DomainColumn) String() string {
	return "VALUE"
}

// Type implements the interface sql.Expression.
func (d *DomainColumn) Type() sql.Type {
	return d.Typ
}

// IsNullable implements the interface sql.Expression.
func (d *DomainColumn) IsNullable() bool {
	return false
}

// Eval implements the interface sql.Expression.
func (d *DomainColumn) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	panic("DomainColumn is a placeholder expression, but Eval() was called")
}

// Children implements the interface sql.Expression.
func (d *DomainColumn) Children() []sql.Expression {
	return nil
}

// WithChildren implements the interface sql.Expression.
func (d *DomainColumn) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(d, len(children), 0)
	}
	return d, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *DomainColumn) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, fmt.Errorf("invalid vitess child count, expected `0` but got `%d`", len(children))
	}
	return d, nil
}
