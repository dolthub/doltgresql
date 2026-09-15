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
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
)

// transactionState is the transaction block state reported by ReadyForQuery.
type transactionState byte

const (
	idleTransactionState     transactionState = 0
	explicitTransactionState transactionState = 'X'
	implicitTransactionState transactionState = 'T'
	failedTransactionState   transactionState = 'E'
)

// ReadyForQueryTransactionIndicator indicates the transaction state reported by ReadyForQuery.
type ReadyForQueryTransactionIndicator byte

const (
	ReadyForQueryTransactionIndicator_Idle                   ReadyForQueryTransactionIndicator = 'I'
	ReadyForQueryTransactionIndicator_TransactionBlock       ReadyForQueryTransactionIndicator = 'T'
	ReadyForQueryTransactionIndicator_FailedTransactionBlock ReadyForQueryTransactionIndicator = 'E'
)

// inExplicitTransactionBlock reports whether the state belongs to an explicit transaction block.
func (s transactionState) inExplicitTransactionBlock() bool {
	return s == explicitTransactionState || s == failedTransactionState
}

// readyStatus returns the wire-protocol transaction status for this state.
func (s transactionState) readyStatus() byte {
	switch s {
	case failedTransactionState:
		return byte(ReadyForQueryTransactionIndicator_FailedTransactionBlock)
	case explicitTransactionState, implicitTransactionState:
		return byte(ReadyForQueryTransactionIndicator_TransactionBlock)
	default:
		return byte(ReadyForQueryTransactionIndicator_Idle)
	}
}

// handleTransactionStatement handles transaction-control statements owned by the connection layer.
func (h *ConnectionHandler) handleTransactionStatement(query ConvertedQuery) (bool, error) {
	switch query.AST.(type) {
	case *sqlparser.Begin:
		switch h.state.transaction {
		case explicitTransactionState:
			// PostgreSQL treats a nested BEGIN as a no-op and keeps the existing transaction characteristics.
			return true, h.send(makeCommandComplete(query.StatementTag, 0))
		case implicitTransactionState:
			// BEGIN promotes the engine transaction already backing the implicit block.
			h.state.transaction = explicitTransactionState
			return true, h.send(makeCommandComplete(query.StatementTag, 0))
		default:
			h.state.transaction = explicitTransactionState
			return false, nil
		}
	case *sqlparser.Commit:
		if h.state.transaction == failedTransactionState {
			h.state.transaction = idleTransactionState
			h.clearTransactionLocalVars()
			if err := h.runEngineTransactionControl("ROLLBACK"); err != nil {
				return true, err
			}
			return true, h.send(&pgproto3.CommandComplete{CommandTag: []byte("ROLLBACK")})
		}
		// COMMIT ends either kind of active block; the engine still executes the statement itself.
		h.state.transaction = idleTransactionState
		h.clearTransactionLocalVars()
		return false, nil
	case *sqlparser.Rollback:
		h.state.transaction = idleTransactionState
		h.clearTransactionLocalVars()
		return false, nil
	case *sqlparser.Savepoint:
		if !h.state.transaction.inExplicitTransactionBlock() {
			return true, noActiveTransactionError("SAVEPOINT")
		}
	case *sqlparser.RollbackSavepoint:
		if !h.state.transaction.inExplicitTransactionBlock() {
			return true, noActiveTransactionError("ROLLBACK TO SAVEPOINT")
		}
		h.state.transaction = explicitTransactionState
	case *sqlparser.ReleaseSavepoint:
		if !h.state.transaction.inExplicitTransactionBlock() {
			return true, noActiveTransactionError("RELEASE SAVEPOINT")
		}
	}
	return false, nil
}

// startImplicitTransaction starts an implicit transaction block unless one is already active or the statement controls transactions.
func (h *ConnectionHandler) startImplicitTransaction(query ConvertedQuery) error {
	if h.state.transaction != idleTransactionState {
		return nil
	}
	switch query.AST.(type) {
	case *sqlparser.Begin, *sqlparser.Commit, *sqlparser.Rollback:
		return nil
	}
	ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return err
	}
	ctx.SetIgnoreAutoCommit(true)
	h.state.transaction = implicitTransactionState
	return nil
}

// commitImplicitTransaction commits the active implicit transaction, rolling it back if the commit fails.
func (h *ConnectionHandler) commitImplicitTransaction() error {
	if h.state.transaction != implicitTransactionState {
		return nil
	}
	h.state.transaction = idleTransactionState
	h.clearTransactionLocalVars()
	if h.restoredAutoCommitWithoutTransaction() {
		return nil
	}
	if err := h.runEngineTransactionControl("COMMIT"); err != nil {
		if rollbackErr := h.runEngineTransactionControl("ROLLBACK"); rollbackErr != nil {
			logrus.Warnf("error rolling back implicit transaction after failed commit: %s", rollbackErr)
		}
		return err
	}
	return nil
}

// rollbackImplicitTransaction rolls back the active implicit transaction.
func (h *ConnectionHandler) rollbackImplicitTransaction() {
	if h.state.transaction != implicitTransactionState {
		return
	}
	h.state.transaction = idleTransactionState
	h.clearTransactionLocalVars()
	if h.restoredAutoCommitWithoutTransaction() {
		return
	}
	if err := h.runEngineTransactionControl("ROLLBACK"); err != nil {
		logrus.Warnf("error rolling back implicit transaction: %s", err)
	}
}

// failActiveTransaction applies PostgreSQL statement-error semantics to the active transaction.
func (h *ConnectionHandler) failActiveTransaction() {
	switch h.state.transaction {
	case implicitTransactionState:
		h.rollbackImplicitTransaction()
	case explicitTransactionState:
		h.state.transaction = failedTransactionState
	}
}

// clearTransactionLocalVars restores session values for variables set with transaction-local scope.
func (h *ConnectionHandler) clearTransactionLocalVars() {
	ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		logrus.Warnf("error creating context to clear transaction-local variables: %s", err)
		return
	}
	if err = ctx.Session.ClearTransactionLocalVariables(ctx); err != nil {
		logrus.Warnf("error clearing transaction-local variables: %s", err)
		return
	}
	_ = core.SetDateStyleOutputFormat(ctx, "")
}

// restoredAutoCommitWithoutTransaction restores autocommit when no engine transaction remains.
func (h *ConnectionHandler) restoredAutoCommitWithoutTransaction() bool {
	ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil || ctx.GetTransaction() != nil {
		return false
	}
	ctx.SetIgnoreAutoCommit(false)
	return true
}

// runEngineTransactionControl executes an internal transaction-control statement without client responses.
func (h *ConnectionHandler) runEngineTransactionControl(statement string) error {
	queries, err := convertQuery(statement)
	if err != nil {
		return err
	}
	return h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, queries[0].String, queries[0].AST, func(*sql.Context, *Result) error { return nil })
}

// rejectStatementIfTransactionFailed rejects statements that cannot run in a failed transaction.
func (h *ConnectionHandler) rejectStatementIfTransactionFailed(query ConvertedQuery) error {
	if h.state.transaction != failedTransactionState || query.AST == nil {
		return nil
	}
	switch query.AST.(type) {
	case *sqlparser.Commit, *sqlparser.Rollback, *sqlparser.RollbackSavepoint:
		return nil
	}
	return &pgconn.PgError{Severity: string(ErrorResponseSeverity_Error), Code: pgcode.InFailedSQLTransaction.String(), Message: "current transaction is aborted, commands ignored until end of transaction block"}
}

// noActiveTransactionError returns PostgreSQL's error for a transaction-block-only command outside a transaction.
func noActiveTransactionError(commandName string) error {
	return &pgconn.PgError{Severity: string(ErrorResponseSeverity_Error), Code: pgcode.NoActiveSQLTransaction.String(), Message: fmt.Sprintf("%s can only be used in transaction blocks", commandName)}
}

// startTransactionIfNecessary starts a read/write transaction when the current session does not have one.
func startTransactionIfNecessary(ctx *sql.Context) error {
	doltSession, ok := ctx.Session.(*dsess.DoltSession)
	if !ok {
		return errors.Errorf("unexpected session type: %T", ctx.Session)
	}
	if doltSession.GetTransaction() == nil {
		if _, err := doltSession.StartTransaction(ctx, sql.ReadWrite); err != nil {
			return err
		}
		ctx.SetIgnoreAutoCommit(true)
	}
	return nil
}
