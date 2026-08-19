// Copyright 2025 Dolthub, Inc.
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
	"github.com/cockroachdb/errors"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeCopyTo handles *tree.CopyTo nodes.
func nodeCopyTo(ctx *Context, node *tree.CopyTo) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if node.Options.CopyFormat == tree.CopyFormatBinary {
		if node.Options.Header {
			return nil, errors.Errorf("cannot specify HEADER in BINARY mode")
		}
		if node.Options.Delimiter != "" {
			return nil, errors.Errorf("cannot specify DELIMITER in BINARY mode")
		}
	}

	// We create a stub select statement for the COPY TO statement, which the connection handler will build a plan
	// for at execution time, streaming the results back to the client (or to a file). When copying a table, we
	// construct a simple SELECT over the columns named (or all columns when none were).
	selectStmt := node.Statement
	if selectStmt == nil {
		var selectExprs tree.SelectExprs
		if len(node.Columns) > 0 {
			selectExprs = make(tree.SelectExprs, len(node.Columns))
			for i := range node.Columns {
				selectExprs[i] = tree.SelectExpr{Expr: tree.NewUnresolvedName(string(node.Columns[i]))}
			}
		} else {
			selectExprs = tree.SelectExprs{tree.StarSelectExpr()}
		}
		selectStmt = &tree.Select{
			Select: &tree.SelectClause{
				Exprs: selectExprs,
				From: tree.From{
					Tables: tree.TableExprs{&node.Table},
				},
			},
		}
	}

	selectStub, err := nodeSelect(ctx, selectStmt)
	if err != nil {
		return nil, err
	}

	return vitess.InjectedStatement{
		Statement: pgnodes.NewCopyTo(
			node.Table.Catalog(),
			doltdb.TableName{
				Name:   node.Table.Object(),
				Schema: node.Table.Schema(),
			},
			node.Options,
			node.File,
			node.Stdout,
			node.Columns,
			selectStub,
		),
		Children: nil,
	}, nil
}
