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
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Greatest represents a GREATEST expression.
type Greatest struct {
	retType *pgtypes.DoltgresType
	Args    []sql.Expression
}

var _ vitess.Injectable = (*Greatest)(nil)
var _ sql.Expression = (*Greatest)(nil)

// Children implements the sql.Expression interface.
func (n *Greatest) Children() []sql.Expression {
	return n.Args
}

// Eval implements the sql.Expression interface.
func (n *Greatest) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return evalGreatestLeast(ctx, row, n.Args, n.retType, 1)
}

// IsNullable implements the sql.Expression interface.
func (n *Greatest) IsNullable(ctx *sql.Context) bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (n *Greatest) Resolved() bool {
	return argsResolved(n.Args) && n.retType != nil
}

// String implements the sql.Expression interface.
func (n *Greatest) String() string {
	return "GREATEST(" + argsString(n.Args) + ")"
}

// Type implements the sql.Expression interface.
func (n *Greatest) Type(ctx *sql.Context) sql.Type {
	if n.retType == nil {
		return pgtypes.Unknown
	}
	return n.retType
}

// WithChildren implements the sql.Expression interface.
func (n *Greatest) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	retType, err := compRetType(ctx, "GREATEST", children...)
	if err != nil {
		return nil, err
	}

	if retType.ID == pgtypes.Numeric.ID {
		// use Numeric type with no
		retType = pgtypes.Numeric
	}

	return &Greatest{
		retType: retType,
		Args:    children,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (n *Greatest) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	args := make([]sql.Expression, len(children))
	for i, child := range children {
		expr, ok := child.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", child)
		}
		args[i] = expr
	}
	return n.WithChildren(ctx.(*sql.Context), args...)
}

// Least represents a LEAST expression.
type Least struct {
	retType *pgtypes.DoltgresType
	Args    []sql.Expression
}

var _ vitess.Injectable = (*Least)(nil)
var _ sql.Expression = (*Least)(nil)

// Children implements the sql.Expression interface.
func (n *Least) Children() []sql.Expression {
	return n.Args
}

// Eval implements the sql.Expression interface.
func (n *Least) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return evalGreatestLeast(ctx, row, n.Args, n.retType, -1)
}

// IsNullable implements the sql.Expression interface.
func (n *Least) IsNullable(ctx *sql.Context) bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (n *Least) Resolved() bool {
	return argsResolved(n.Args) && n.retType != nil
}

// String implements the sql.Expression interface.
func (n *Least) String() string {
	return "LEAST(" + argsString(n.Args) + ")"
}

// Type implements the sql.Expression interface.
func (n *Least) Type(ctx *sql.Context) sql.Type {
	if n.retType == nil {
		return pgtypes.Unknown
	}
	return n.retType
}

// WithChildren implements the sql.Expression interface.
func (n *Least) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	retType, err := compRetType(ctx, "LEAST", children...)
	if err != nil {
		return nil, err
	}
	return &Least{
		retType: retType,
		Args:    children,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (n *Least) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	args := make([]sql.Expression, len(children))
	for i, child := range children {
		expr, ok := child.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", child)
		}
		args[i] = expr
	}
	return n.WithChildren(ctx.(*sql.Context), args...)
}

// argsResolved returns true when every expression in args is resolved.
func argsResolved(args []sql.Expression) bool {
	for _, arg := range args {
		if !arg.Resolved() {
			return false
		}
	}
	return true
}

// argsString renders args as a comma-separated list, matching the format used for function calls.
func argsString(args []sql.Expression) string {
	strs := make([]string, len(args))
	for i, arg := range args {
		strs[i] = arg.String()
	}
	return strings.Join(strs, ", ")
}

// evalGreatestLeast evaluates args, casts each non-NULL result to retType, and returns the extreme value.
// sign is 1 for GREATEST (select the maximum) and -1 for LEAST (select the minimum).
func evalGreatestLeast(ctx *sql.Context, row sql.Row, args []sql.Expression, retType *pgtypes.DoltgresType, sign int) (any, error) {
	castsColl, err := core.GetCastsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}

	var selected any
	for _, arg := range args {
		val, err := arg.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if val == nil {
			continue
		}

		argType, ok := arg.Type(ctx).(*pgtypes.DoltgresType)
		if !ok {
			return nil, errors.Errorf("expected DoltgresType, but got %s", arg.Type(ctx).String())
		}
		if argType.ID != retType.ID {
			cast, err := castsColl.GetImplicitCast(ctx, argType, retType)
			if err != nil {
				return nil, err
			}
			if !cast.ID.IsValid() {
				return nil, errors.Errorf("cannot find cast function from %s to %s", argType.String(), retType.String())
			}
			val, err = cast.Eval(ctx, val, argType, retType)
			if err != nil {
				return nil, err
			}
		}

		if selected == nil {
			selected = val
			continue
		}

		cmp, err := retType.Compare(ctx, val, selected)
		if err != nil {
			return nil, err
		}
		if cmp*sign > 0 {
			selected = val
		}
	}
	return selected, nil
}

// compRetType is used to determine the type from args based on the rules described for GREATEST/LEAST.
// https://www.postgresql.org/docs/current/typeconv-union-case.html
func compRetType(ctx *sql.Context, funcName string, args ...sql.Expression) (*pgtypes.DoltgresType, error) {
	if len(args) == 0 {
		return nil, sql.ErrInvalidArgumentNumber.New(funcName, "1 or more", 0)
	}

	castsColl, err := core.GetCastsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}

	allSameType := true
	allUnknown := true
	var candType *pgtypes.DoltgresType

	typs := make([]*pgtypes.DoltgresType, len(args))

	for i, arg := range args {
		if !arg.Resolved() {
			return nil, nil
		}
		argType := arg.Type(ctx)
		dt, ok := argType.(*pgtypes.DoltgresType)
		if !ok {
			continue
		}

		if dt.TypType == pgtypes.TypeType_Domain {
			dt = dt.DomainUnderlyingBaseType()
		}
		typs[i] = dt

		if i == 0 {
			candType = dt
		} else if dt.ID != candType.ID {
			allSameType = false
		}

		if dt.ID == pgtypes.Unknown.ID {
			continue
		}
		allUnknown = false

		if dt.TypCategory != candType.TypCategory {
			return nil, sql.ErrInvalidType.New(argType.String())
		}

		// If the candidate type can be implicitly converted to the other type,
		// but not vice-versa, select the other type as the new candidate type.
		candToArg, err := castsColl.GetImplicitCast(ctx, candType, dt)
		if err != nil {
			return nil, err
		}
		argToCand, err := castsColl.GetImplicitCast(ctx, dt, candType)
		if err != nil {
			return nil, err
		}
		if candToArg.ID.IsValid() && !argToCand.ID.IsValid() {
			candType = dt
		}
	}

	if allUnknown {
		return pgtypes.Text, nil
	}
	if allSameType {
		return candType, nil
	}

	// verify that every argument type can be implicitly cast to the candidate type.
	for _, dt := range typs {
		if dt.ID == candType.ID || dt.ID == pgtypes.Unknown.ID {
			continue
		}
		cast, err := castsColl.GetImplicitCast(ctx, dt, candType)
		if err != nil {
			return nil, err
		}
		if !cast.ID.IsValid() {
			return nil, sql.ErrInvalidType.New(dt.String())
		}
	}

	return candType, nil
}
