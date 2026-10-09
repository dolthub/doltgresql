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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/node"
)

// declareCursor creates the cursor from a DECLARE CURSOR statement, reading every row of its query immediately so that
// FETCH does not see later changes.
func (h *ConnectionHandler) declareCursor(declare *node.DeclareCursor, query ConvertedQuery, simpleQuery *simpleQueryExecution) error {
	if !declare.IsHoldable && !h.isInTransactionBlock(simpleQuery) {
		return noActiveTransactionError(query.StatementTag)
	}
	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, query.String)
	if err != nil {
		return err
	}
	if _, ok, err := core.GetCursor(sqlCtx, declare.Name); err != nil {
		return err
	} else if ok {
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
	analyzedNode, err := h.doltgresHandler.e.Analyzer.Analyze(sqlCtx, boundNode, nil, flags)
	if err != nil {
		return err
	}
	schema, iter, _, err := h.doltgresHandler.e.PrepQueryPlanForExecution(sqlCtx, query.String, analyzedNode, flags)
	if err != nil {
		return err
	}
	rows, err := sql.RowIterToRows(sqlCtx, iter)
	if err != nil {
		return err
	}
	cursor := core.NewCursor(declare.Name, query.String, schema, rows, declare.IsHoldable, declare.IsScrollable)
	if err = core.AddCursor(sqlCtx, cursor); err != nil {
		return err
	}
	if h.state.txState != idleTransactionState {
		h.state.transactionCursors = append(h.state.transactionCursors, declare.Name)
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

// closeTransactionCursors closes the cursors that end with a transaction, given whether that transaction committed.
func (h *ConnectionHandler) closeTransactionCursors(committed bool) {
	names := h.state.transactionCursors
	h.state.transactionCursors = nil
	ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		logrus.Warnf("error creating context to close transaction cursors: %s", err)
		return
	}
	for _, name := range names {
		cursor, ok, err := core.GetCursor(ctx, name)
		if err == nil && ok && (!cursor.IsHoldable || !committed) {
			err = cursor.Close(ctx)
		}
		if err != nil {
			logrus.Warnf("error closing transaction cursor `%s`: %s", name, err)
		}
	}
}
