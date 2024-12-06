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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ExplicitCast represents a VALUE::TYPE expression.
type ExplicitCast struct {
	sqlChild       sql.Expression
	castToType     *pgtypes.DoltgresType
	domainNullable bool
	domainChecks   sql.CheckConstraints
}

var _ vitess.Injectable = (*ExplicitCast)(nil)
var _ sql.Expression = (*ExplicitCast)(nil)

// NewExplicitCastInjectable returns an incomplete *ExplicitCast that must be resolved through the vitess.Injectable interface.
func NewExplicitCastInjectable(castToType sql.Type) (*ExplicitCast, error) {
	pgtype, ok := castToType.(*pgtypes.DoltgresType)
	if !ok {
		return nil, fmt.Errorf("cast expects a Doltgres type as the target type")
	}
	return &ExplicitCast{
		sqlChild:   nil,
		castToType: pgtype,
	}, nil
}

// NewExplicitCast returns a new *ExplicitCast expression.
func NewExplicitCast(expr sql.Expression, toType *pgtypes.DoltgresType) *ExplicitCast {
	toType = checkForDomainType(toType)
	return &ExplicitCast{
		sqlChild:   expr,
		castToType: toType,
	}
}

// Children implements the sql.Expression interface.
func (c *ExplicitCast) Children() []sql.Expression {
	return []sql.Expression{c.sqlChild}
}

// Child returns the child that is being cast.
func (c *ExplicitCast) Child() sql.Expression {
	return c.sqlChild
}

// Eval implements the sql.Expression interface.
func (c *ExplicitCast) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := c.sqlChild.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	fromType, ok := c.sqlChild.Type().(*pgtypes.DoltgresType)
	if !ok {
		// We'll leverage GMSCast to handle the conversion from a GMS type to a Doltgres type.
		// Rather than re-evaluating the expression, we put the result in a literal.
		gmsCast := NewGMSCast(expression.NewLiteral(val, c.sqlChild.Type()))
		val, err = gmsCast.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		fromType = gmsCast.DoltgresType()
	}
	if val == nil {
		if c.castToType.TypType == pgtypes.TypeType_Domain && !c.domainNullable {
			return nil, pgtypes.ErrDomainDoesNotAllowNullValues.New(c.castToType.Name)
		}
		return nil, nil
	}

	baseCastToType := checkForDomainType(c.castToType)
	castFunction := framework.GetExplicitCast(fromType, baseCastToType)
	if castFunction == nil {
		if fromType.OID == uint32(oid.T_unknown) {
			castFunction = framework.UnknownLiteralCast
		} else {
			return nil, fmt.Errorf("EXPLICIT CAST: cast from `%s` to `%s` does not exist: %s",
				fromType.String(), c.castToType.String(), c.sqlChild.String())
		}
	}
	castResult, err := castFunction(ctx, val, c.castToType)
	if err != nil {
		// For string types and string array types, we intentionally ignore the error as using a length-restricted cast
		// is a way to intentionally truncate the data. All string types will always return the truncated result, even
		// during an error, so it's safe to use.
		castToType := c.castToType
		if c.castToType.IsArrayType() {
			castToType = c.castToType.ArrayBaseType()
		}
		// A nil result will be returned if there's a critical error, which we should never ignore.
		if castToType.TypCategory != pgtypes.TypeCategory_StringTypes || castResult == nil {
			return nil, err
		}
	}

	if c.castToType.TypType == pgtypes.TypeType_Domain {
		for _, check := range c.domainChecks {
			res, err := sql.EvaluateCondition(ctx, check.Expr, sql.Row{castResult})
			if err != nil {
				return nil, err
			}
			if sql.IsFalse(res) {
				return nil, pgtypes.ErrDomainValueViolatesCheckConstraint.New(c.castToType.Name, check.Name)
			}
		}
	}

	return castResult, nil
}

// IsNullable implements the sql.Expression interface.
func (c *ExplicitCast) IsNullable() bool {
	// TODO: verify if this is actually nullable
	return true
}

// Resolved implements the sql.Expression interface.
func (c *ExplicitCast) Resolved() bool {
	if c.sqlChild != nil && c.sqlChild.Resolved() {
		return true
	}
	return false
}

// String implements the sql.Expression interface.
func (c *ExplicitCast) String() string {
	return c.sqlChild.String() + "::" + c.castToType.String()
}

// Type implements the sql.Expression interface.
func (c *ExplicitCast) Type() sql.Type {
	return c.castToType
}

// WithChildren implements the sql.Expression interface.
func (c *ExplicitCast) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(c, len(children), 1)
	}
	return &ExplicitCast{
		sqlChild:       children[0],
		castToType:     c.castToType,
		domainNullable: c.domainNullable,
		domainChecks:   c.domainChecks,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (c *ExplicitCast) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 1 {
		return nil, fmt.Errorf("invalid vitess child count, expected `1` but got `%d`", len(children))
	}
	resolvedExpression, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	return &ExplicitCast{
		sqlChild:       resolvedExpression,
		castToType:     c.castToType,
		domainNullable: c.domainNullable,
		domainChecks:   c.domainChecks,
	}, nil
}

// WithCastToType returns a copy of the expression with castToType replaced.
func (c *ExplicitCast) WithCastToType(t *pgtypes.DoltgresType) sql.Expression {
	ec := *c
	ec.castToType = t
	return &ec
}

// WithDomainConstraints returns a copy of the expression with domain constraints defined.
func (c *ExplicitCast) WithDomainConstraints(nullable bool, checks sql.CheckConstraints) sql.Expression {
	ec := *c
	ec.domainNullable = nullable
	ec.domainChecks = checks
	return &ec
}
