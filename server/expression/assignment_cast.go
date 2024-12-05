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
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AssignmentCast handles assignment casts.
type AssignmentCast struct {
	expr     sql.Expression
	fromType *pgtypes.DoltgresType
	toType   *pgtypes.DoltgresType
}

var _ sql.Expression = (*AssignmentCast)(nil)

// NewAssignmentCast returns a new *AssignmentCast expression.
func NewAssignmentCast(expr sql.Expression, fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType) *AssignmentCast {
	toType = checkForDomainType(toType)
	fromType = checkForDomainType(fromType)
	return &AssignmentCast{
		expr:     expr,
		fromType: fromType,
		toType:   toType,
	}
}

// Children implements the sql.Expression interface.
func (ac *AssignmentCast) Children() []sql.Expression {
	return []sql.Expression{ac.expr}
}

// Eval implements the sql.Expression interface.
func (ac *AssignmentCast) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := ac.expr.Eval(ctx, row)
	if err != nil || val == nil {
		return val, err
	}
	castFunc := framework.GetAssignmentCast(ac.fromType, ac.toType)
	if castFunc == nil {
		if ac.fromType.OID == uint32(oid.T_unknown) {
			castFunc = framework.UnknownLiteralCast
		} else {
			return nil, fmt.Errorf("ASSIGNMENT_CAST: target is of type %s but expression is of type %s: %s",
				ac.toType.String(), ac.fromType.String(), ac.expr.String())
		}
	}
	return castFunc(ctx, val, ac.toType)
}

// IsNullable implements the sql.Expression interface.
func (ac *AssignmentCast) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (ac *AssignmentCast) Resolved() bool {
	return ac.expr.Resolved()
}

// String implements the sql.Expression interface.
func (ac *AssignmentCast) String() string {
	return ac.expr.String()
}

// Type implements the sql.Expression interface.
func (ac *AssignmentCast) Type() sql.Type {
	return ac.toType
}

// WithChildren implements the sql.Expression interface.
func (ac *AssignmentCast) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(ac, len(children), 1)
	}
	return NewAssignmentCast(children[0], ac.fromType, ac.toType), nil
}

func checkForDomainType(t *pgtypes.DoltgresType) *pgtypes.DoltgresType {
	if t.TypType == pgtypes.TypeType_Domain {
		t = t.DomainUnderlyingBaseType()
	}
	return t
}
