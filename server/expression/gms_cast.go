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

package expression

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// GMSCast handles the conversion from a GMS expression's type to its Doltgres type that is most similar.
type GMSCast struct {
	sqlChild sql.Expression
}

var _ sql.Expression = (*GMSCast)(nil)

// NewGMSCast returns a new *GMSCast.
func NewGMSCast(child sql.Expression) *GMSCast {
	return &GMSCast{
		sqlChild: child,
	}
}

// Children implements the sql.Expression interface.
func (c *GMSCast) Children() []sql.Expression {
	return []sql.Expression{c.sqlChild}
}

// Child returns the child that is being cast.
func (c *GMSCast) Child() sql.Expression {
	return c.sqlChild
}

// DoltgresType returns the DoltgresType that the cast evaluates to. This is the same value that is returned by Type().
func (c *GMSCast) DoltgresType(ctx *sql.Context) *pgtypes.DoltgresType {
	// GMSCast shouldn't receive a DoltgresType, but we shouldn't error if it happens
	if t, ok := c.sqlChild.Type(ctx).(*pgtypes.DoltgresType); ok {
		return t
	}

	return pgtypes.FromGmsType(c.sqlChild.Type(ctx))
}

// Eval implements the sql.Expression interface.
func (c *GMSCast) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := c.sqlChild.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	// GMSCast shouldn't receive a DoltgresType, but we shouldn't error if it happens
	if _, ok := c.sqlChild.Type(ctx).(*pgtypes.DoltgresType); ok {
		return val, nil
	}
	return framework.ConvertGMSValueToDoltgresValue(ctx, c.sqlChild.Type(ctx), val)
}

// IsNullable implements the sql.Expression interface.
func (c *GMSCast) IsNullable(ctx *sql.Context) bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (c *GMSCast) Resolved() bool {
	return c.sqlChild.Resolved()
}

// String implements the sql.Expression interface.
func (c *GMSCast) String() string {
	if gf, ok := c.sqlChild.(*expression.GetField); ok {
		return gf.Name()
	}
	return c.sqlChild.String()
}

// Type implements the sql.Expression interface.
func (c *GMSCast) Type(ctx *sql.Context) sql.Type {
	return c.DoltgresType(ctx)
}

// WithChildren implements the sql.Expression interface.
func (c *GMSCast) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(c, len(children), 1)
	}
	return &GMSCast{
		sqlChild: children[0],
	}, nil
}
