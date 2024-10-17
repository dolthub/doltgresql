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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
)

// AccessControl handles all forms of access control to database objects and their contents. This includes privilege
// checking, ownership, Row-Level Security, etc.
func AccessControl(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return pgtransform.NodeWithOpaque(node, func(node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		accessControl, ok := node.(*pgnodes.AccessControl)
		if !ok {
			return node, transform.SameTree, nil
		}
		if err := accessControl.CheckAccess(ctx); err != nil {
			return nil, transform.NewTree, err
		}
		// TODO: implement Row-Level Security -> https://www.postgresql.org/docs/15/ddl-priv.html
		// For now, since we don't implement RLS, we'll remove the AccessControl node
		return accessControl.Child, transform.NewTree, nil
	})
}
