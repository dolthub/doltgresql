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

package node

import (
	"context"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/settings"
)

// SetSetting executes PostgreSQL SET and SET LOCAL without routing custom
// setting names through GMS's registered system-variable planner.
type SetSetting struct {
	Name  string
	Local bool
	Reset bool
	Value sql.Expression
}

var _ sql.ExecSourceRel = (*SetSetting)(nil)
var _ sql.Expressioner = (*SetSetting)(nil)
var _ vitess.Injectable = (*SetSetting)(nil)

func (*SetSetting) Children() []sql.Node           { return nil }
func (*SetSetting) IsReadOnly() bool               { return false }
func (s *SetSetting) Resolved() bool               { return s.Reset || s.Value != nil && s.Value.Resolved() }
func (*SetSetting) Schema(*sql.Context) sql.Schema { return nil }
func (s *SetSetting) String() string               { return fmt.Sprintf("SET %s", s.Name) }
func (s *SetSetting) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(s, children...)
}
func (s *SetSetting) WithResolvedChildren(_ context.Context, children []any) (any, error) {
	if len(children) > 1 || (!s.Reset && len(children) != 1) || (s.Reset && len(children) != 0) {
		return nil, ErrVitessChildCount.New(1, len(children))
	}
	next := *s
	if len(children) == 1 {
		next.Value = children[0].(sql.Expression)
	}
	return &next, nil
}
func (s *SetSetting) Expressions() []sql.Expression {
	if s.Reset {
		return nil
	}
	return []sql.Expression{s.Value}
}
func (s *SetSetting) WithExpressions(_ *sql.Context, exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) > 1 || (!s.Reset && len(exprs) != 1) || (s.Reset && len(exprs) != 0) {
		return nil, sql.ErrInvalidChildrenNumber.New(s, len(exprs), 1)
	}
	next := *s
	if len(exprs) == 1 {
		next.Value = exprs[0]
	}
	return &next, nil
}
func (s *SetSetting) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	var value any
	if !s.Reset {
		var err error
		value, err = s.Value.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
	}
	if err := settings.Set(ctx, s.Name, value, s.Reset, s.Local); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}
