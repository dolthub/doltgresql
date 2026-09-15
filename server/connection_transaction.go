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
	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

// transactionState is the transaction block state of a connection. See the field of the same name on
// ConnectionHandler for a description of the states.
type transactionState byte

const (
	idleTransactionState     transactionState = 0
	explicitTransactionState transactionState = 'X'
	implicitTransactionState transactionState = 'T'
	failedTransactionState   transactionState = 'E'
)

// inExplicitTransactionBlock returns whether this state is inside an explicit transaction block. A failed
// transaction block is still an explicit transaction block: it remains open until the client ends it.
func (s transactionState) inExplicitTransactionBlock() bool {
	return s == explicitTransactionState || s == failedTransactionState
}

// startImplicitTransaction starts an implicit transaction block for the given statement, unless a transaction
// block (implicit or explicit) is already in progress, or the statement is itself a transaction control
// statement.
func (h *ConnectionHandler) startImplicitTransaction(query ConvertedQuery) error {
	if h.transactionState != idleTransactionState {
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
	h.transactionState = implicitTransactionState
	return nil
}

// commitImplicitTransaction commits the implicit transaction block in progress, if there is one. If the commit
// fails, the transaction is rolled back instead, and the commit error is returned.
func (h *ConnectionHandler) commitImplicitTransaction() error {
	if h.transactionState != implicitTransactionState {
		return nil
	}
	h.transactionState = idleTransactionState
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

// rollbackImplicitTransaction rolls back the implicit transaction block in progress, if there is one
func (h *ConnectionHandler) rollbackImplicitTransaction() {
	if h.transactionState != implicitTransactionState {
		return
	}
	h.transactionState = idleTransactionState
	h.clearTransactionLocalVars()
	if h.restoredAutoCommitWithoutTransaction() {
		return
	}
	if err := h.runEngineTransactionControl("ROLLBACK"); err != nil {
		logrus.Warnf("error rolling back implicit transaction: %s", err)
	}
}

// clearTransactionLocalVars removes any system variable values that were set with transaction-local scope
// (SET LOCAL, or set_config with is_local), restoring the variables' session values. Called whenever the current
// transaction block ends, whether by COMMIT or ROLLBACK: Postgres reverts SET LOCAL values in both cases.
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
	// Reset any cached variables in ContextValues, in case a cached parameter (e.g. datestyle) was overridden
	_ = core.SetDateStyleOutputFormat(ctx, "")
}

// restoredAutoCommitWithoutTransaction returns whether the session no longer has an engine transaction in
// progress, restoring the session's autocommit behavior if so. Some statements end the engine transaction
// themselves as a side effect of executing (e.g. dolt_assume_cluster_role, which also poisons the session
// against any further use), and some never start one at all (e.g. DEALLOCATE, which is handled by this handler
// without involving the engine). In either case there is nothing left for an implicit transaction block to
// commit or roll back, but autocommit must still be restored, since no COMMIT or ROLLBACK statement will run
// through the engine to do it for us.
func (h *ConnectionHandler) restoredAutoCommitWithoutTransaction() bool {
	ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return false
	}
	if ctx.GetTransaction() != nil {
		return false
	}
	ctx.SetIgnoreAutoCommit(false)
	return true
}

// runEngineTransactionControl runs the given transaction control statement (BEGIN, COMMIT, or ROLLBACK) through
// the engine without sending any response messages to the client. This is used to manage the engine transaction
// backing an implicit transaction block, which is invisible to the client.
func (h *ConnectionHandler) runEngineTransactionControl(statement string) error {
	queries, err := h.convertQuery(statement)
	if err != nil {
		return err
	}
	return h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, queries[0].String, queries[0].AST,
		func(*sql.Context, *Result) error {
			return nil
		})
}

// rejectStatementIfTransactionFailed returns an error if the current transaction block is in a failed state and
// the given statement is not one that ends the transaction block.
func (h *ConnectionHandler) rejectStatementIfTransactionFailed(query ConvertedQuery) error {
	if h.transactionState != failedTransactionState || query.AST == nil {
		return nil
	}
	switch query.AST.(type) {
	case *sqlparser.Commit, *sqlparser.Rollback, *sqlparser.RollbackSavepoint:
		return nil
	}
	return &pgconn.PgError{
		Severity: string(ErrorResponseSeverity_Error),
		Code:     pgcode.InFailedSQLTransaction.String(),
		Message:  "current transaction is aborted, commands ignored until end of transaction block",
	}
}

// noActiveTransactionError returns the error that Postgres reports when the given transaction-block-only command
// (e.g. SAVEPOINT) is used outside of an explicit transaction block.
func noActiveTransactionError(commandName string) error {
	return &pgconn.PgError{
		Severity: string(ErrorResponseSeverity_Error),
		Code:     pgcode.NoActiveSQLTransaction.String(),
		Message:  fmt.Sprintf("%s can only be used in transaction blocks", commandName),
	}
}

// startTransactionIfNecessary checks to see if the current session has a transaction started yet or not, and if not,
// creates a read/write transaction for the session to use. This is necessary for handling commands that alter
// data without going through the GMS engine.
func startTransactionIfNecessary(ctx *sql.Context) error {
	doltSession, ok := ctx.Session.(*dsess.DoltSession)
	if !ok {
		return errors.Errorf("unexpected session type: %T", ctx.Session)
	}
	if doltSession.GetTransaction() == nil {
		if _, err := doltSession.StartTransaction(ctx, sql.ReadWrite); err != nil {
			return err
		}

		// When we start a transaction ourselves, we must ignore auto-commit settings for transaction
		ctx.SetIgnoreAutoCommit(true)
	}

	return nil
}

// ReadyForQueryTransactionIndicator indicates the state of the transaction related to the query.
type ReadyForQueryTransactionIndicator byte

const (
	ReadyForQueryTransactionIndicator_Idle                   ReadyForQueryTransactionIndicator = 'I'
	ReadyForQueryTransactionIndicator_TransactionBlock       ReadyForQueryTransactionIndicator = 'T'
	ReadyForQueryTransactionIndicator_FailedTransactionBlock ReadyForQueryTransactionIndicator = 'E'
)
