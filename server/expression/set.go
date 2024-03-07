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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// SetVar represents a SET <configuration_parameter> TO <value> expression.
type SetVar struct {
	name string
	val  sql.Expression
}

var _ vitess.InjectableExpression = (*SetVar)(nil)
var _ sql.Expression = (*SetVar)(nil)

// NewSetVar returns a new *SetVar.
func NewSetVar(n string, v sql.Expression) *SetVar {

	return &SetVar{name: n}
}

// Children implements the sql.Expression interface.
func (s *SetVar) Children() []sql.Expression {
	return []sql.Expression{s.val}
}

// Eval implements the sql.Expression interface.
func (s *SetVar) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return nil, nil
}

// IsNullable implements the sql.Expression interface.
func (s *SetVar) IsNullable() bool {
	return s.val.IsNullable()
}

// Resolved implements the sql.Expression interface.
func (s *SetVar) Resolved() bool {
	return s.val.Resolved()
}

// String implements the sql.Expression interface.
func (s *SetVar) String() string {
	return fmt.Sprintf("SET %s TO %s", s.name, s.val.String())
}

// Type implements the sql.Expression interface.
func (s *SetVar) Type() sql.Type {
	return s.val.Type()
}

// WithChildren implements the sql.Expression interface.
func (s *SetVar) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(s, len(children), 1)
	}
	return &SetVar{
		name: s.name,
		val:  children[0],
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (s *SetVar) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(s, len(children), 1)
	}
	val, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}

	return &SetVar{
		name: s.name,
		val:  val,
	}, nil
}
