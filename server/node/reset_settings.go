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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/config"
)

// ResetSettings restores changed PostgreSQL GUCs to their connection defaults.
type ResetSettings struct{}

var _ sql.ExecSourceRel = (*ResetSettings)(nil)
var _ vitess.Injectable = (*ResetSettings)(nil)

func (*ResetSettings) Children() []sql.Node           { return nil }
func (*ResetSettings) IsReadOnly() bool               { return false }
func (*ResetSettings) Resolved() bool                 { return true }
func (*ResetSettings) Schema(*sql.Context) sql.Schema { return nil }
func (*ResetSettings) String() string                 { return "RESET ALL" }
func (r *ResetSettings) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(r, children...)
}
func (r *ResetSettings) WithResolvedChildren(_ context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return r, nil
}
func (*ResetSettings) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	names, err := core.SettingNames(ctx)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		if config.IsValidPostgresConfigParameter(name) {
			value, err := ctx.GetSessionVariableDefault(ctx, name)
			if err != nil {
				return nil, err
			}
			if err := config.SetPostgresParameter(ctx, name, value, false); err != nil {
				return nil, err
			}
		} else if err := core.SetSetting(ctx, name, "", false); err != nil {
			return nil, err
		}
	}
	return sql.RowsToRowIter(), nil
}
