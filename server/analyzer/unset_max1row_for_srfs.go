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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// UnsetMax1RowForSRFs unsets the QFlagMax1Row query flag when the plan contains a set-returning function
// (e.g. unnest, generate_series). GMS sets QFlagMax1Row when it detects that a query can return at most one row
// (e.g. a point lookup on a unique index), and the handler uses that flag to take a result-spooling shortcut that
// errors if more than one row is produced. That analysis doesn't account for set-returning functions in the
// projection, which can multiply the number of output rows. See https://github.com/dolthub/doltgresql/issues/3111.
func UnsetMax1RowForSRFs(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	if qFlags == nil || !qFlags.IsSet(sql.QFlagMax1Row) {
		return node, transform.SameTree, nil
	}

	containsSRF := false
	transform.InspectExpressions(ctx, node, func(ctx *sql.Context, expr sql.Expression) bool {
		if rowIterExpr, ok := expr.(sql.RowIterExpression); ok && rowIterExpr.ReturnsRowIter() {
			containsSRF = true
		}
		return !containsSRF
	})

	if containsSRF {
		qFlags.Unset(sql.QFlagMax1Row)
	}
	return node, transform.SameTree, nil
}
