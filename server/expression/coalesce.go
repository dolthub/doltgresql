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

package expression

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgCoalesce is a Doltgres-native COALESCE implementation. It uses Postgres type-resolution rules
// (FindCommonType) to compute the correct result type.
type PgCoalesce struct {
	args []sql.Expression
	typ  *pgtypes.DoltgresType
}

var _ sql.Expression = (*PgCoalesce)(nil)
var _ sql.FunctionExpression = (*PgCoalesce)(nil)
var _ sql.CollationCoercible = (*PgCoalesce)(nil)

// NewPgCoalesce creates a new PgCoalesce expression.
func NewPgCoalesce(ctx *sql.Context, args ...sql.Expression) (*PgCoalesce, error) {
	if len(args) == 0 {
		return nil, sql.ErrInvalidArgumentNumber.New("COALESCE", "1 or more", 0)
	}
	expr, err := (&PgCoalesce{typ: pgtypes.Unknown}).WithChildren(ctx, args...)
	if err != nil {
		return nil, err
	}
	return expr.(*PgCoalesce), nil
}

// FunctionName implements sql.FunctionExpression.
func (c *PgCoalesce) FunctionName() string { return "coalesce" }

// Description implements sql.FunctionExpression.
func (c *PgCoalesce) Description() string { return "returns the first non-null value in a list." }

// Type implements sql.Expression.
func (c *PgCoalesce) Type(_ *sql.Context) sql.Type {
	return c.typ
}

// CollationCoercibility implements sql.CollationCoercible.
func (c *PgCoalesce) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	if cc, ok := c.Type(ctx).(sql.CollationCoercible); ok {
		return cc.CollationCoercibility(ctx)
	}
	return sql.Collation_binary, 6
}

// IsNullable implements sql.Expression.
func (c *PgCoalesce) IsNullable(_ *sql.Context) bool {
	return true
}

// Resolved implements sql.Expression.
func (c *PgCoalesce) Resolved() bool {
	for _, arg := range c.args {
		if arg == nil || !arg.Resolved() {
			return false
		}
	}
	return true
}

// Children implements sql.Expression.
func (c *PgCoalesce) Children() []sql.Expression { return c.args }

// WithChildren implements sql.Expression.
func (c *PgCoalesce) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) == 0 {
		return nil, sql.ErrInvalidArgumentNumber.New("COALESCE", "1 or more", 0)
	}
	newC := &PgCoalesce{args: children, typ: pgtypes.Unknown}
	childTypes := make([]*pgtypes.DoltgresType, 0, len(children))
	for _, child := range children {
		dt, ok := child.Type(ctx).(*pgtypes.DoltgresType)
		if !ok {
			return newC, nil
		}
		childTypes = append(childTypes, dt)
	}
	commonType, _, err := framework.FindCommonType(ctx, childTypes)
	if err != nil {
		return nil, err
	}
	if commonType != nil {
		newC.typ = commonType
	}
	return newC, nil
}

// Eval implements sql.Expression. Returns the first non-null argument value, cast to the common type.
func (c *PgCoalesce) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	commonType := c.typ
	for _, arg := range c.args {
		if arg == nil {
			continue
		}
		val, err := arg.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if val == nil {
			continue
		}
		if commonType == pgtypes.Unknown {
			return val, nil
		}
		argType, ok := arg.Type(ctx).(*pgtypes.DoltgresType)
		if ok && argType.Equals(commonType) {
			return val, nil
		}
		// Cast the value to the common type (handles mixed-type args, e.g. int2 and int4).
		converted, _, err := commonType.ConvertToType(ctx, argType, val, 'a')
		if err != nil {
			return nil, err
		}
		return converted, nil
	}
	return nil, nil
}

// String implements sql.Expression.
func (c *PgCoalesce) String() string {
	args := make([]string, len(c.args))
	for i, arg := range c.args {
		args[i] = arg.String()
	}
	return fmt.Sprintf("coalesce(%s)", strings.Join(args, ","))
}

// DebugString implements the sql.Debuggable interface.
func (c *PgCoalesce) DebugString(ctx *sql.Context) string {
	args := make([]string, len(c.args))
	for i, arg := range c.args {
		args[i] = sql.DebugString(ctx, arg)
	}
	return fmt.Sprintf("coalesce(%s)", strings.Join(args, ","))
}
