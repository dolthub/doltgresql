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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Interval is a constant INTERVAL expression. It's a plain interval literal (embedding
// *expression.Literal for Eval/Type/etc.) that additionally implements GMS's expression.TimeDeltaExpression
// interface, so that GMS's generic Arithmetic expression can add/subtract it to/from a date/time value via
// GMS's own calendar-correct expression.TimeDelta.
type Interval struct {
	*expression.Literal
	delta *expression.TimeDelta
}

var _ sql.Expression = (*Interval)(nil)
var _ sql.CollationCoercible = (*Interval)(nil)
var _ vitess.Injectable = (*Interval)(nil)
var _ expression.TimeDeltaExpression = (*Interval)(nil)

// NewInterval returns a new *Interval wrapping the given duration.
func NewInterval(d duration.Duration) *Interval {
	return &Interval{
		Literal: expression.NewLiteral(d, pgtypes.Interval),
		delta: &expression.TimeDelta{
			Months:       d.Months,
			Days:         d.Days,
			Microseconds: d.Nanos() / 1000,
		},
	}
}

// String implements the sql.Expression interface.
func (i *Interval) String() string {
	return fmt.Sprintf("INTERVAL '%s'", i.Val.(duration.Duration).String())
}

// WithChildren implements the sql.Expression interface. This overrides the embedded Literal's
// implementation, which would otherwise return the inner *Literal and drop this type's TimeDeltaExpression
// capability.
func (i *Interval) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(i, len(children), 0)
	}
	return i, nil
}

// WithResolvedChildren implements the vitess.Injectable interface, overriding the embedded Literal's
// implementation for the same reason as WithChildren above.
func (i *Interval) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(i, len(children), 0)
	}
	return i, nil
}

// EvalDelta implements GMS's expression.TimeDeltaExpression interface, returning this interval's duration
// pre-converted into GMS's own expression.TimeDelta (computed once in NewInterval, since it depends only
// on the immutable duration).
func (i *Interval) EvalDelta(ctx *sql.Context, row sql.Row) (*expression.TimeDelta, error) {
	return i.delta, nil
}
