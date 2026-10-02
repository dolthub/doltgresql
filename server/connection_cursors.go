// Copyright 2026 Dolthub, Inc.
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

package server

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dtables"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/node"
)

// declareCursor creates the cursor from a DECLARE CURSOR statement. The cursor's query pins every table by its root, so
// FETCH can read its rows as needed without seeing later changes. When some source of rows cannot be locked to its
// current data (such as a function that reads some current value), the cursor instead reads every row immediately. Rows
// are also read in advance when using a WITH HOLD cursor declared outside an explicit transaction block, since its
// transaction commits right away.
func (h *ConnectionHandler) declareCursor(declare *node.DeclareCursor, query ConvertedQuery, simpleQuery *simpleQueryExecution) error {
	if !declare.IsHoldable && !h.inTransactionBlock(simpleQuery) {
		return noActiveTransactionError(query.StatementTag)
	}
	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, query.String)
	if err != nil {
		return err
	}
	cursors, err := core.GetCursors(sqlCtx)
	if err != nil {
		return err
	}
	if _, ok := cursors[declare.Name]; ok {
		return pgerror.Newf(pgcode.DuplicateCursor, `cursor "%s" already exists`, declare.Name)
	}
	builder := planbuilder.New(sqlCtx, h.doltgresHandler.e.Analyzer.Catalog, nil)
	if declare.Bindings != nil {
		builder.SetBindings(declare.Bindings)
	}
	boundNode, flags, err := builder.BindOnly(declare.Select, "", nil)
	if err != nil {
		return err
	}
	boundNode, isFullyLocked, err := lockTablesToRoots(sqlCtx, boundNode)
	if err != nil {
		return err
	}
	analyzedNode, err := h.doltgresHandler.e.Analyzer.Analyze(sqlCtx, boundNode, nil, flags)
	if err != nil {
		return err
	}
	schema, iter, _, err := h.doltgresHandler.e.PrepQueryPlanForExecution(sqlCtx, query.String, analyzedNode, flags)
	if err != nil {
		return err
	}
	if !isFullyLocked || (declare.IsHoldable && !h.state.txState.inExplicitTransactionBlock()) {
		rows, err := sql.RowIterToRows(sqlCtx, iter)
		if err != nil {
			return err
		}
		iter = sql.RowsToRowIter(rows...)
	}
	cursors[declare.Name] = &core.Cursor{
		Name:          declare.Name,
		Statement:     query.String,
		Schema:        schema,
		IsHoldable:    declare.IsHoldable,
		IsScrollable:  declare.IsScrollable,
		CreationTime:  time.Now(),
		Iter:          iter,
		InTransaction: h.state.txState != idleTransactionState,
	}
	return h.send(makeCommandComplete(query.StatementTag, 0))
}

// asDeclareCursor returns the DECLARE CURSOR statement that the query holds, if it holds one.
func asDeclareCursor(query ConvertedQuery) (*node.DeclareCursor, bool) {
	injected, ok := query.AST.(vitess.InjectedStatement)
	if !ok {
		return nil, false
	}
	declare, ok := injected.Statement.(*node.DeclareCursor)
	return declare, ok
}

// bindDeclareCursor handles a Bind message for a DECLARE CURSOR statement, whose parameters are given to its query
// once the cursor is declared.
func (h *ConnectionHandler) bindDeclareCursor(message *pgproto3.Bind, preparedData preparedStatementData, declare *node.DeclareCursor) error {
	sqlCtx, err := h.doltgresHandler.sm.NewContextWithQuery(context.Background(), h.mysqlConn, preparedData.Query.String)
	if err != nil {
		return err
	}
	boundDeclare := *declare
	boundDeclare.Bindings, err = h.doltgresHandler.convertBindParameters(sqlCtx, preparedData.BindVarTypes, message.ParameterFormatCodes, message.Parameters)
	if err != nil {
		return err
	}
	query := preparedData.Query
	query.AST = vitess.InjectedStatement{Statement: &boundDeclare}
	h.state.extendedQueryObjects.portals[message.DestinationPortal] = portalData{Query: query}
	return h.send(&pgproto3.BindComplete{})
}

// lockTablesToRoots replaces every table in the plan, including those in subqueries, with a version that is locked to
// the root, so that it won't see later changes. This is the same locking that Dolt uses for AS OF queries, so tables
// that are read AS OF a commit are already locked to that commit. It also returns whether every source of rows was
// locked, which is not the case for table functions and for tables that Dolt does not store, such as system tables.
func lockTablesToRoots(ctx *sql.Context, n sql.Node) (sql.Node, bool, error) {
	doltSession := dsess.DSessFromSess(ctx.Session)
	isFullyLocked := true
	n, _, err := transform.NodeWithOpaque(ctx, n, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		if _, ok := n.(sql.TableFunction); ok {
			isFullyLocked = false
			return n, transform.SameTree, nil
		}
		resolvedTable, ok := n.(*plan.ResolvedTable)
		if !ok || resolvedTable.AsOf != nil || plan.IsDualTable(resolvedTable.Table) {
			return n, transform.SameTree, nil
		}
		versionableTable, ok := resolvedTable.Table.(dtables.VersionableTable)
		if !ok {
			isFullyLocked = false
			return n, transform.SameTree, nil
		}
		roots, ok := doltSession.GetRoots(ctx, resolvedTable.Database().Name())
		if !ok {
			return nil, transform.SameTree, errors.Errorf("unable to find roots for database `%s`", resolvedTable.Database().Name())
		}
		lockedTable, err := versionableTable.LockedToRoot(ctx, roots.Working)
		if err != nil {
			return nil, transform.SameTree, err
		}
		newNode, err := resolvedTable.WithTable(ctx, lockedTable)
		return newNode, transform.NewTree, err
	})
	if err != nil {
		return nil, false, err
	}
	n, _, err = transform.NodeExprsWithOpaque(ctx, n, func(ctx *sql.Context, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		subquery, ok := expr.(*plan.Subquery)
		if !ok {
			return expr, transform.SameTree, nil
		}
		query, isQueryLocked, err := lockTablesToRoots(ctx, subquery.Query)
		if err != nil {
			return nil, transform.SameTree, err
		}
		isFullyLocked = isFullyLocked && isQueryLocked
		return subquery.WithQuery(query), transform.NewTree, nil
	})
	return n, isFullyLocked, err
}

// materializeHoldableCursors reads the remaining rows of the WITH HOLD cursors declared in the transaction that is
// committing, so that they can still be read after the commit.
func (h *ConnectionHandler) materializeHoldableCursors() error {
	ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return err
	}
	return core.MaterializeHoldableCursors(ctx)
}

// closeTransactionCursors closes the cursors that end with a transaction, given whether that transaction committed.
func (h *ConnectionHandler) closeTransactionCursors(committed bool) {
	ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		logrus.Warnf("error creating context to close transaction cursors: %s", err)
		return
	}
	if err = core.EndCursorTransaction(ctx, committed); err != nil {
		logrus.Warnf("error closing transaction cursors: %s", err)
	}
}
