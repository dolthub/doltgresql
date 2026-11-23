// Copyright 2025 Dolthub, Inc.
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
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// NewRecordExpr creates a new record expression.
func NewRecordExpr() *RecordExpr {
	return &RecordExpr{}
}

// NewCatalogRowExpr creates a whole-row reference with a catalog's named row type.
func NewCatalogRowExpr(typ *pgtypes.DoltgresType, bareReference bool) *RecordExpr {
	return &RecordExpr{typ: typ, bareReference: bareReference}
}

// RecordExpr is a set of sql.Expressions wrapped together in a single value.
type RecordExpr struct {
	exprs         []sql.Expression
	typ           *pgtypes.DoltgresType
	bareReference bool
}

var _ sql.Expression = (*RecordExpr)(nil)
var _ vitess.Injectable = (*RecordExpr)(nil)

// Resolved implements the sql.Expression interface.
func (t *RecordExpr) Resolved() bool {
	for _, expr := range t.exprs {
		if !expr.Resolved() {
			return false
		}
	}
	return true
}

// String implements the sql.Expression interface.
func (t *RecordExpr) String() string {
	fields := make([]string, len(t.exprs))
	for i, expr := range t.exprs {
		fields[i] = expr.String()
	}
	return "ROW(" + strings.Join(fields, ", ") + ")"
}

// Type implements the sql.Expression interface.
func (t *RecordExpr) Type(ctx *sql.Context) sql.Type {
	if t.typ != nil {
		return t.typ
	}
	return pgtypes.Record
}

// IsNullable implements the sql.Expression interface.
func (t *RecordExpr) IsNullable(ctx *sql.Context) bool {
	return t.typ != nil
}

// Eval implements the sql.Expression interface.
func (t *RecordExpr) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	vals := make([]pgtypes.RecordValue, len(t.exprs))
	allNull := true
	for i, expr := range t.exprs {
		val, err := expr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		allNull = allNull && val == nil

		t := expr.Type(ctx)
		typ, ok := t.(*pgtypes.DoltgresType)
		if !ok {
			// TODO: it would be better if we had a doltgres type for NULL literals in these records
			if typ, ok := t.(sql.NullType); ok && typ.IsNullType() {
				typ = pgtypes.Unknown
			} else {
				return nil, fmt.Errorf("expected a DoltgresType, but got %T", t)
			}
		}
		vals[i] = pgtypes.RecordValue{
			Value: val,
			Type:  typ,
		}
	}
	// Catalog rows have non-null fields, so an all-null whole-row reference
	// represents the missing side of an outer join.
	if t.typ != nil && allNull {
		return nil, nil
	}

	return vals, nil
}

// Children implements the sql.Expression interface.
func (t *RecordExpr) Children() []sql.Expression {
	return t.exprs
}

// WithChildren implements the sql.Expression interface.
func (t *RecordExpr) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	tCopy := *t
	tCopy.exprs = children
	return &tCopy, nil
}

// WithResolvedChildren implements the vitess.Injectable interface
func (t *RecordExpr) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if t.bareReference {
		if len(children) == 0 {
			return nil, sql.ErrInvalidChildrenNumber.New(t, 0, 1)
		}
		// PostgreSQL resolves an unqualified column before a whole-row reference
		// with the same name. Let the builder make that decision.
		if _, ok := children[0].(*TableToComposite); !ok {
			return children[0], nil
		}
		children = children[1:]
	}
	newExpressions := make([]sql.Expression, len(children))
	for i, resolvedChild := range children {
		resolvedExpression, ok := resolvedChild.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", resolvedChild)
		}
		newExpressions[i] = resolvedExpression
	}
	return t.WithChildren(ctx.(*sql.Context), newExpressions...)
}
