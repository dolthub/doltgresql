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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// applyTablesForAnalyzeAllTables finds plan.AnalyzeTable nodes that don't have any tables explicitly specified and fills in all
// tables for the current database. This enables the ANALYZE; statement to analyze all tables.
func applyTablesForAnalyzeAllTables(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	analyzeTable, ok := node.(*plan.AnalyzeTable)
	if !ok {
		return node, transform.SameTree, nil
	}

	// If a set of tables is already populated, we don't need to do anything. We only fill in all tables when
	// the caller didn't explicitly specify any tables to be analyzed.
	if len(analyzeTable.Tables) > 0 {
		return node, transform.SameTree, nil
	}

	db, err := a.Catalog.Database(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return node, transform.SameTree, err
	}
	tableNames, err := db.GetTableNames(ctx)
	if err != nil {
		return node, transform.SameTree, err
	}

	var tables []sql.Table
	for _, tableName := range tableNames {
		table, ok, err := db.GetTableInsensitive(ctx, tableName)
		if err != nil {
			return node, transform.SameTree, err
		} else if !ok {
			return node, transform.SameTree, sql.ErrTableNotFound.New(tableName)
		}
		tables = append(tables, table)
	}

	return analyzeTable.WithTables(tables), transform.NewTree, nil
}
