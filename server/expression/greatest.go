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
	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Greatest represents a GREATEST expression.
type Greatest struct {
	casts   []casts.Cast
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
	return evalGreatestLeast(ctx, row, n.Args, n.retType, n.casts, 1)
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
	if len(children) == 0 {
		return nil, sql.ErrInvalidArgumentNumber.New("GREATEST", "1 or more", 0)
	}

	retType, castList, err := getRetTypeAndCasts(ctx, children)
	if err != nil {
		return nil, err
	}

	return &Greatest{
		casts:   castList,
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

// getRetTypeAndCasts evaluates argument types to get return type and implicit casts if needed.
func getRetTypeAndCasts(ctx *sql.Context, children []sql.Expression) (*pgtypes.DoltgresType, []casts.Cast, error) {
	castsColl, err := core.GetCastsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, nil, err
	}

	var typs = make([]*pgtypes.DoltgresType, len(children))
	var nonNullTypes []*pgtypes.DoltgresType
	for i, child := range children {
		argType, ok := child.Type(ctx).(*pgtypes.DoltgresType)
		if !ok {
			continue
		}
		typs[i] = argType
		nonNullTypes = append(nonNullTypes, argType)
	}

	retType, _, err := framework.FindCommonType(ctx, nonNullTypes)
	if err != nil {
		return nil, nil, err
	}

	if retType.ID == pgtypes.Numeric.ID {
		// use Numeric type with no precision
		retType = pgtypes.Numeric
	}

	var castList = make([]casts.Cast, len(children))
	for i, argType := range typs {
		if argType != nil && argType.ID != retType.ID {
			castList[i], err = castsColl.GetImplicitCast(ctx, argType, retType)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return retType, castList, nil
}

// evalGreatestLeast evaluates args, casts each non-NULL result to retType, and returns the extreme value.
// sign is 1 for GREATEST (select the maximum) and -1 for LEAST (select the minimum).
func evalGreatestLeast(ctx *sql.Context, row sql.Row, args []sql.Expression, retType *pgtypes.DoltgresType, casts []casts.Cast, sign int) (any, error) {
	var selected any
	for i, arg := range args {
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
			if !casts[i].ID.IsValid() {
				return nil, errors.Errorf("cannot find cast function from %s to %s", argType.String(), retType.String())
			}
			val, err = casts[i].Eval(ctx, val, argType, retType)
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
