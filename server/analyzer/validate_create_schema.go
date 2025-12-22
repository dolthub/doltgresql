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

	"github.com/dolthub/doltgresql/core"
)

// ValidateCreateSchema validates that CREATE SCHEMA is executed with a valid database context.
// In PostgreSQL, schemas exist within databases, so CREATE SCHEMA requires an active database.
// See: https://github.com/dolthub/doltgresql/issues/1863
func ValidateCreateSchema(
	ctx *sql.Context,
	a *analyzer.Analyzer,
	n sql.Node,
	scope *plan.Scope,
	sel analyzer.RuleSelector,
	qFlags *sql.QueryFlags,
) (sql.Node, transform.TreeIdentity, error) {
	cs, ok := n.(*plan.CreateSchema)
	if !ok {
		return n, transform.SameTree, nil
	}

	// Check if the current database actually has a working root.
	// GetRootFromContext will return sql.ErrDatabaseNotFound if the database
	// is not properly initialized.
	_, _, err := core.GetRootFromContext(ctx)
	if err != nil {
		if sql.ErrDatabaseNotFound.Is(err) {
			return nil, transform.SameTree, sql.ErrNoDatabaseSelected.New()
		}
		return nil, transform.SameTree, err
	}

	return cs, transform.SameTree, nil
}
