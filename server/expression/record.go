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
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// NewRecordExpr creates a new record expression.
func NewRecordExpr() *RecordExpr {
	return &RecordExpr{}
}

// RecordExpr is a set of sql.Expressions wrapped together in a single value.
type RecordExpr struct {
	exprs []sql.Expression
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
	return "RECORD EXPR"
}

// Type implements the sql.Expression interface.
func (t *RecordExpr) Type() sql.Type {
	return pgtypes.Record
}

// IsNullable implements the sql.Expression interface.
func (t *RecordExpr) IsNullable() bool {
	return false
}

// Eval implements the sql.Expression interface.
func (t *RecordExpr) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	vals := make([]pgtypes.RecordValue, len(t.exprs))
	for i, expr := range t.exprs {
		val, err := expr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}

		typ, ok := expr.Type().(*pgtypes.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("expected a DoltgresType, but got %T", expr.Type())
		}
		vals[i] = pgtypes.RecordValue{
			Value: val,
			Type:  typ,
		}
	}

	return vals, nil
}

// Children implements the sql.Expression interface.
func (t *RecordExpr) Children() []sql.Expression {
	return t.exprs
}

// WithChildren implements the sql.Expression interface.
func (t *RecordExpr) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	tCopy := *t
	tCopy.exprs = children
	return &tCopy, nil
}

// WithResolvedChildren implements the vitess.Injectable interface
func (t *RecordExpr) WithResolvedChildren(children []any) (any, error) {
	newExpressions := make([]sql.Expression, len(children))
	for i, resolvedChild := range children {
		resolvedExpression, ok := resolvedChild.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", resolvedChild)
		}
		newExpressions[i] = resolvedExpression
	}
	return t.WithChildren(newExpressions...)
}
