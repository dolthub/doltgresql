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

// nodeExplain handles *tree.Explain nodes.
func nodeExplain(ctx *Context, node *tree.Explain) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	if node.TableName != nil {
		tableName, err := nodeUnresolvedObjectName(ctx, node.TableName)
		if err != nil {
			return nil, err
		}

		var showTableOpts *vitess.ShowTablesOpt
		if node.AsOf != nil {
			asOf, err := nodeExpr(ctx, node.AsOf.Expr)
			if err != nil {
				return nil, err
			}
			showTableOpts = &vitess.ShowTablesOpt{
				AsOf:       asOf,
				SchemaName: tableName.SchemaQualifier.String(),
				DbName:     tableName.DbQualifier.String(),
			}
		}

		show := &vitess.Show{
			Type:          "columns",
			Table:         tableName,
			ShowTablesOpt: showTableOpts,
		}

		return show, nil
	}

	if stmt, ok := node.Statement.(*tree.Select); ok {
		// TODO: read tree.ExplainOptions
		selectStmt, err := nodeSelect(ctx, stmt)
		if err != nil {
			return nil, err
		}
		explain := &vitess.Explain{
			ExplainFormat: vitess.TreeStr,
			Statement:     selectStmt,
		}
		return explain, nil
	}

	return nil, fmt.Errorf("This EXPLAIN syntax is not yet supported")
}
