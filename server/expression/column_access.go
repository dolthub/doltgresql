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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ColumnAccess represents an ARRAY[...] expression.
type ColumnAccess struct {
	colName    string
	colNameIdx int
	colTyp     *pgtypes.DoltgresType
	child      sql.Expression
}

var _ vitess.Injectable = (*ColumnAccess)(nil)
var _ sql.Expression = (*ColumnAccess)(nil)

// NewColumnAccess returns a new *ColumnAccess.
func NewColumnAccess(colName string, colIdx int) (*ColumnAccess, error) {
	if len(colName) > 0 {
		return &ColumnAccess{
			colName:    colName,
			colNameIdx: -1,
			colTyp:     nil,
			child:      nil,
		}, nil
	} else {
		return &ColumnAccess{
			colName:    "",
			colNameIdx: colIdx,
			colTyp:     nil,
			child:      nil,
		}, nil
	}
}

// Children implements the sql.Expression interface.
func (expr *ColumnAccess) Children() []sql.Expression {
	return []sql.Expression{expr.child}
}

// Eval implements the sql.Expression interface.
func (expr *ColumnAccess) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	field, err := expr.child.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if field == nil {
		return nil, nil
	}
	recordVals, ok := field.([]pgtypes.RecordValue)
	if !ok {
		if len(expr.colName) > 0 {
			return nil, errors.Errorf("column notation .%s applied to type %s, which is not a composite type",
				expr.colName, expr.child.Type().String())
		} else {
			return nil, errors.Errorf("column notation .@%d applied to type %s, which is not a composite type",
				expr.colNameIdx+1, expr.child.Type().String())
		}
	}
	return recordVals[expr.colNameIdx].Value, nil
}

// IsNullable implements the sql.Expression interface.
func (expr *ColumnAccess) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (expr *ColumnAccess) Resolved() bool {
	return expr.child != nil && expr.child.Resolved()
}

// String implements the sql.Expression interface.
func (expr *ColumnAccess) String() string {
	if expr.child == nil {
		return "COLUMN_ACCESS"
	}
	if len(expr.colName) > 0 {
		return fmt.Sprintf("(%s).%s", expr.child.String(), expr.colName)
	} else {
		return fmt.Sprintf("(%s).@%d", expr.child.String(), expr.colNameIdx+1)
	}
}

// Type implements the sql.Expression interface.
func (expr *ColumnAccess) Type() sql.Type {
	if expr.colTyp != nil {
		return expr.colTyp
	}
	if expr.child == nil {
		return nil
	}
	// We're technically returning a different type here since an unresolved type is not the same as a resolved one.
	// However, for many early analyzer steps, we only check the ID, so this at least lets us get past those cases.
	return pgtypes.NewUnresolvedDoltgresTypeFromID(expr.child.Type().(*pgtypes.DoltgresType).CompositeAttrs[expr.colNameIdx].TypeID)
}

// WithChildren implements the sql.Expression interface.
func (expr *ColumnAccess) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(expr, len(children), 1)
	}
	childType := children[0].Type()
	doltgresType, ok := childType.(*pgtypes.DoltgresType)
	if !ok {
		return nil, errors.New("column access is only valid for Doltgres types")
	}
	if !doltgresType.IsCompositeType() {
		if len(expr.colName) > 0 {
			return nil, errors.Errorf("column notation .%s applied to type %s, which is not a composite type",
				expr.colName, children[0].Type().String())
		} else {
			return nil, errors.Errorf("column notation .@%d applied to type %s, which is not a composite type",
				expr.colNameIdx+1, children[0].Type().String())
		}
	}
	var idx int
	if len(expr.colName) > 0 {
		idx = -1
		for _, attr := range doltgresType.CompositeAttrs {
			if attr.Name == expr.colName {
				idx = int(attr.Num - 1)
				break
			}
		}
		if idx == -1 {
			return nil, errors.Errorf(`column "%s" not found in data type %s`,
				expr.colName, doltgresType.String())
		}
	} else {
		if expr.colNameIdx < 0 || expr.colNameIdx >= len(doltgresType.CompositeAttrs) {
			return nil, errors.Errorf("column notation .@%d applied to type %s is out of bounds",
				expr.colNameIdx+1, children[0].Type().String())
		}
		idx = expr.colNameIdx
	}
	return &ColumnAccess{
		colName:    expr.colName,
		colNameIdx: idx,
		colTyp:     expr.colTyp,
		child:      children[0],
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (expr *ColumnAccess) WithResolvedChildren(children []any) (any, error) {
	newExpressions := make([]sql.Expression, len(children))
	for i, resolvedChild := range children {
		resolvedExpression, ok := resolvedChild.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", resolvedChild)
		}
		newExpressions[i] = resolvedExpression
	}
	return expr.WithChildren(newExpressions...)
}

// WithType returns this expression with the given type set, as it must be set within the analyzer.
func (expr *ColumnAccess) WithType(typ *pgtypes.DoltgresType) sql.Expression {
	ne := *expr
	ne.colTyp = typ
	return &ne
}
