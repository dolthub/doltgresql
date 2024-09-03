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
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/index"
)

// ReplaceIndexedTables replaces Dolt tables with Doltgres tables that can properly handle indexed access.
func ReplaceIndexedTables(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(node, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		if filter, ok := n.(*plan.Filter); ok {
			return transform.Node(filter, replaceIndexedTablesFilter)
		}
		return n, transform.SameTree, nil
	})
}

// replaceIndexedTablesFilter is the transform function for ReplaceIndexedTables that handles the resolved table.
func replaceIndexedTablesFilter(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
	switch n := n.(type) {
	case *plan.ResolvedTable:
		if newTable, ok, err := doltTableToIndexedTable(n.UnderlyingTable()); err != nil {
			return n, transform.SameTree, err
		} else if ok {
			nt, err := n.WithTable(newTable)
			return nt, transform.NewTree, err
		}
		return n, transform.SameTree, nil
	default:
		return n, transform.SameTree, nil
	}
}

// doltTableToIndexedTable replaces Dolt tables with Doltgres' indexed tables.
func doltTableToIndexedTable(table sql.Table) (sql.Table, bool, error) {
	switch table := table.(type) {
	case *sqle.AlterableDoltTable:
		return &index.WritableDoltgresTable{WritableDoltTable: &table.WritableDoltTable}, true, nil
	case *sqle.WritableDoltTable:
		return &index.WritableDoltgresTable{WritableDoltTable: table}, true, nil
	case *sqle.DoltTable:
		return &index.DoltgresTable{DoltTable: table}, true, nil
	case *index.DoltgresTable:
		return table, false, nil
	default:
		return table, false, nil
	}
}
