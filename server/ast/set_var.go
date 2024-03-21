// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"fmt"
	"github.com/dolthub/doltgresql/server/config"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeSetVar handles *tree.SetVar nodes.
func nodeSetVar(node *tree.SetVar) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if !config.IsValidPostgresConfigParameter(node.Name) {
		return nil, fmt.Errorf(`ERROR: syntax error at or near "%s"'`, node.Name)
	}
	if node.IsLocal {
		// TODO: takes effect for only the current transaction rather than the current session.
		return nil, fmt.Errorf("SET LOCAL is not yet supported")
	}
	expr, err := nodeExpr(node.Values[0])
	if err != nil {
		return nil, err
	}
	setStmt := &vitess.Set{
		Exprs: vitess.SetVarExprs{&vitess.SetVarExpr{
			Scope: vitess.SetScope_Session,
			Name: &vitess.ColName{
				Name: vitess.NewColIdent(node.Name),
			},
			Expr: expr,
		}},
	}
	return setStmt, nil
}
