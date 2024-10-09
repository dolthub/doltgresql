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
func nodeExplain(node *tree.Explain) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	if node.TableName != nil {
		tableName, err := nodeUnresolvedObjectName(node.TableName)
		if err != nil {
			return nil, err
		}

		var showTableOpts *vitess.ShowTablesOpt
		if node.AsOf != nil {
			asOf, err := nodeExpr(node.AsOf.Expr)
			if err != nil {
				return nil, err
			}
			showTableOpts = &vitess.ShowTablesOpt{
				AsOf: asOf,
			}
		}

		show := &vitess.Show{
			Type:          "columns",
			Table:         tableName,
			ShowTablesOpt: showTableOpts,
		}

		return show, nil
	}

	return nil, fmt.Errorf("EXPLAIN is not yet supported")
}
