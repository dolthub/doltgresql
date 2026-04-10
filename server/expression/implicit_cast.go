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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ImplicitCast handles implicit casts.
type ImplicitCast struct {
	expr       sql.Expression
	sourceType *pgtypes.DoltgresType
	targetType *pgtypes.DoltgresType
}

var _ sql.Expression = (*ImplicitCast)(nil)

// NewImplicitCast returns a new *ImplicitCast expression.
func NewImplicitCast(expr sql.Expression, fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType) *ImplicitCast {
	toType = checkForDomainType(toType)
	fromType = checkForDomainType(fromType)
	return &ImplicitCast{
		expr:       expr,
		sourceType: fromType,
		targetType: toType,
	}
}

// Children implements the sql.Expression interface.
func (ic *ImplicitCast) Children() []sql.Expression {
	return []sql.Expression{ic.expr}
}

// Eval implements the sql.Expression interface.
func (ic *ImplicitCast) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := ic.expr.Eval(ctx, row)
	if err != nil || val == nil {
		return val, err
	}
	castsColl, err := core.GetCastsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	cast, err := castsColl.GetImplicitCast(ctx, ic.sourceType, ic.targetType)
	if err != nil {
		return nil, err
	}
	if !cast.ID.IsValid() {
		return nil, errors.Errorf("target is of type %s but expression is of type %s", ic.targetType.String(), ic.sourceType.String())
	}
	return cast.Eval(ctx, val, ic.sourceType, ic.targetType)
}

// IsNullable implements the sql.Expression interface.
func (ic *ImplicitCast) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (ic *ImplicitCast) Resolved() bool {
	return ic.expr.Resolved()
}

// String implements the sql.Expression interface.
func (ic *ImplicitCast) String() string {
	return ic.expr.String()
}

// Type implements the sql.Expression interface.
func (ic *ImplicitCast) Type() sql.Type {
	return ic.targetType
}

// WithChildren implements the sql.Expression interface.
func (ic *ImplicitCast) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(ic, len(children), 1)
	}
	return NewImplicitCast(children[0], ic.sourceType, ic.targetType), nil
}
