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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/auth"
)

// SetRole changes the selected role. Permission is always checked against the
// session role, even when another role is currently selected.
type SetRole struct {
	Name  string
	Local bool
	None  bool
	Reset bool
}

var _ sql.ExecSourceRel = (*SetRole)(nil)
var _ vitess.Injectable = (*SetRole)(nil)

func (s *SetRole) Children() []sql.Node           { return nil }
func (s *SetRole) IsReadOnly() bool               { return false }
func (s *SetRole) Resolved() bool                 { return true }
func (s *SetRole) Schema(*sql.Context) sql.Schema { return nil }
func (s *SetRole) String() string                 { return "SET ROLE" }
func (s *SetRole) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(s, children...)
}
func (s *SetRole) WithResolvedChildren(_ context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return s, nil
}

func (s *SetRole) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	if err := auth.ApplySetRole(ctx, s.Name, s.None, s.Reset, s.Local); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}
