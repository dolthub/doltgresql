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

package core

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
)

// SQLNodeToDoltTable takes a sql.Node and returns a *sqle.DoltTable if either the node is a Dolt table, or it is a
// wrapper or container that holds a Dolt table. Returns nil if a Dolt table could not be found. If the node is not a
// sql.Table, then this will return nil.
func SQLNodeToDoltTable(n sql.Node) *sqle.DoltTable {
	tbl, ok := n.(sql.Table)
	if !ok {
		return nil
	}
	return SQLTableToDoltTable(tbl)
}

// SQLTableToDoltTable takes a sql.Table and returns a *sqle.DoltTable if either the table is a Dolt table, or it is a
// wrapper or container that holds a Dolt table. Returns nil if a Dolt table could not be found.
func SQLTableToDoltTable(tbl sql.Table) *sqle.DoltTable {
	switch t := tbl.(type) {
	case *plan.ResolvedTable:
		return SQLTableToDoltTable(t.Table)
	case *plan.ProcessTable:
		return SQLTableToDoltTable(t.Table)
	case *plan.IndexedTableAccess:
		return SQLTableToDoltTable(t.Table)
	case *plan.ProcedureResolvedTable:
		return SQLTableToDoltTable(t.ResolvedTable.Table)
	case *sqle.WritableIndexedDoltTable:
		return t.WritableDoltTable.DoltTable
	case *sqle.IndexedDoltTable:
		return t.DoltTable
	case *sqle.AlterableDoltTable:
		return t.WritableDoltTable.DoltTable
	case *sqle.WritableDoltTable:
		return t.DoltTable
	case *sqle.DoltTable:
		return t
	default:
		if wrapper, ok := tbl.(sql.TableWrapper); ok {
			return SQLTableToDoltTable(wrapper.Underlying())
		}
		return nil
	}
}
