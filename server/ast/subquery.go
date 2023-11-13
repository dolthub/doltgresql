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

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeSubquery handles *tree.Subquery nodes.
func nodeSubquery(node *tree.Subquery) (*vitess.Subquery, error) {
	if node == nil {
		return nil, nil
	}
	if node.Exists {
		return nil, fmt.Errorf("EXISTS is not yet supported")
	}
	selectStmt, err := nodeSelectStatement(node.Select)
	if err != nil {
		return nil, err
	}
	return &vitess.Subquery{
		Select: selectStmt,
	}, nil
}

// nodeSubquery handles *tree.Subquery nodes, returning a vitess.TableExpr.
func nodeSubqueryToTableExpr(node *tree.Subquery) (vitess.TableExpr, error) {
	subquery, err := nodeSubquery(node)
	if err != nil {
		return nil, err
	}
	return &vitess.AliasedTableExpr{
		Expr: subquery,
		As:   vitess.NewTableIdent(generateUniqueAlias()),
	}, nil
}
