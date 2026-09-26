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

type SetSessionAuthorization struct {
	Name  string
	Local bool
	Reset bool
}

var _ sql.ExecSourceRel = (*SetSessionAuthorization)(nil)
var _ vitess.Injectable = (*SetSessionAuthorization)(nil)

func (s *SetSessionAuthorization) Children() []sql.Node           { return nil }
func (s *SetSessionAuthorization) IsReadOnly() bool               { return false }
func (s *SetSessionAuthorization) Resolved() bool                 { return true }
func (s *SetSessionAuthorization) Schema(*sql.Context) sql.Schema { return nil }
func (s *SetSessionAuthorization) String() string                 { return "SET SESSION AUTHORIZATION" }
func (s *SetSessionAuthorization) WithChildren(_ *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(s, children...)
}
func (s *SetSessionAuthorization) WithResolvedChildren(_ context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return s, nil
}
func (s *SetSessionAuthorization) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	if err := auth.ApplySessionAuthorization(ctx, s.Name, s.Reset, s.Local); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}
