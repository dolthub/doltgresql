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
	coretypes "github.com/dolthub/doltgresql/core/types"
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

func (c *CreateDomain) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	// TODO: implement privilege checking
	return true
}

func (c *CreateDomain) Children() []sql.Node {
	return nil
}

func (c *CreateDomain) IsReadOnly() bool {
	return false
}

func (c *CreateDomain) Resolved() bool {
	return true
}

func (c *CreateDomain) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
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
	d := types.DomainType{
		Schema:      c.SchemaName,
		Name:        c.Name,
		AsType:      c.AsType,
		DefaultExpr: defExpr,
		NotNull:     c.IsNotNull,
		Checks:      checkDefs,
	}

	newType, err := coretypes.NewDomainType(ctx, d, "")
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
	return sql.RowsToRowIter(), nil
}

func (c *CreateDomain) Schema() sql.Schema {
	return nil
}

func (c *CreateDomain) String() string {
	return "CREATE DOMAIN"
}

func (c *CreateDomain) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

func (c *CreateDomain) WithResolvedChildren(children []any) (any, error) {
	checksStartAt := 0
	if c.HasDefault {
		expr, ok := children[0].(sql.Expression)
		if !ok {
			return nil, fmt.Errorf("invalid vitess child, expected sql.Expression for Default value but got %t", children[0])
		}
		c.DefaultExpr = expr
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
	c.CheckConstraints = checks
	return c, nil
}

type DomainColumn struct {
	Typ types.DoltgresType
}

var _ vitess.Injectable = (*DomainColumn)(nil)
var _ sql.Expression = (*DomainColumn)(nil)

func (d *DomainColumn) Resolved() bool {
	return true
}

func (d *DomainColumn) String() string {
	return "VALUE"
}

func (d *DomainColumn) Type() sql.Type {
	return d.Typ
}

func (d *DomainColumn) IsNullable() bool {
	return false
}

func (d *DomainColumn) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	panic("DomainColumn is a placeholder expression, but Eval() was called")
}

func (d *DomainColumn) Children() []sql.Expression {
	return nil
}

func (d *DomainColumn) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(d, len(children), 0)
	}
	return d, nil
}

func (d *DomainColumn) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, fmt.Errorf("invalid vitess child count, expected `0` but got `%d`", len(children))
	}
	return d, nil
}
