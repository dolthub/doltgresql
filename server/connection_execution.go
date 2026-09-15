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
	"github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
	"strings"
)

// handleQueryOutsideEngine handles any queries that should be handled by the handler directly, rather than being
// passed to the engine. The response parameter |handled| is true if the query was handled, |endOfMessages| is true
// if no more messages are expected for this query and server should send the client a READY FOR QUERY message,
// and any error that occurred while handling the query.
func (h *ConnectionHandler) handleQueryOutsideEngine(query ConvertedQuery) (handled bool, endOfMessages bool, err error) {
	switch stmt := query.AST.(type) {
	case *sqlparser.Begin:
		if h.transactionState == explicitTransactionState {
			// Postgres treats a BEGIN issued while already inside a transaction block as a no-op
			// (after emitting a warning): the existing transaction, and its original characteristics
			// (isolation level, read/write mode), continue unchanged. If we forwarded this statement
			// to the engine instead, it would start a brand new transaction using this statement's
			// characteristics, silently discarding the active one (e.g. a READ WRITE transaction could
			// be replaced by a READ ONLY one, causing writes to be rejected and changes to be lost).
			return true, true, h.send(makeCommandComplete(query.StatementTag, 0))
		}
		if h.transactionState == implicitTransactionState {
			// A BEGIN inside an implicit transaction block converts it into a regular (explicit) transaction
			// block: the statements already executed in the implicit block are NOT committed, but instead are
			// retroactively included in the new explicit block. The engine transaction backing the implicit
			// block simply continues as the explicit block's transaction, so we don't involve the engine here.
			h.transactionState = explicitTransactionState
			return true, true, h.send(makeCommandComplete(query.StatementTag, 0))
		}
		h.transactionState = explicitTransactionState
	case *sqlparser.Commit:
		if h.transactionState == failedTransactionState {
			// A COMMIT issued inside a failed transaction block ends the block by rolling it back, and reports
			// ROLLBACK to the client to indicate that the transaction's effects were discarded.
			h.transactionState = idleTransactionState
			h.clearTransactionLocalVars()
			if err := h.runEngineTransactionControl("ROLLBACK"); err != nil {
				return true, true, err
			}
			return true, true, h.send(&pgproto3.CommandComplete{CommandTag: []byte("ROLLBACK")})
		}
		// A COMMIT closes the current transaction block, whether explicit or implicit. Any statements that
		// follow it in the same Query message (or extended-query batch) run in a new implicit transaction block.
		h.transactionState = idleTransactionState
		h.clearTransactionLocalVars()
	case *sqlparser.Rollback:
		// Like COMMIT, a ROLLBACK closes the current transaction block, whether explicit, implicit, or failed.
		h.transactionState = idleTransactionState
		h.clearTransactionLocalVars()
	case *sqlparser.Savepoint:
		if !h.transactionState.inExplicitTransactionBlock() {
			return true, true, noActiveTransactionError("SAVEPOINT")
		}
	case *sqlparser.RollbackSavepoint:
		if !h.transactionState.inExplicitTransactionBlock() {
			return true, true, noActiveTransactionError("ROLLBACK TO SAVEPOINT")
		}
		h.transactionState = explicitTransactionState
	case *sqlparser.ReleaseSavepoint:
		if !h.transactionState.inExplicitTransactionBlock() {
			return true, true, noActiveTransactionError("RELEASE SAVEPOINT")
		}
	case *sqlparser.Deallocate:
		return true, true, h.deallocatePreparedStatement(stmt.Name, h.preparedStatements, query, h.Conn())
	case sqlparser.InjectedStatement:
		switch injectedStmt := stmt.Statement.(type) {
		case node.DiscardStatement:
			return true, true, h.discardAll(query)
		case *node.CopyFrom:
			// When copying data from STDIN, the data is sent to the server as CopyData messages
			// We send endOfMessages=false since the server will be in COPY DATA mode and won't
			// be ready for more queries util COPY DATA mode is completed.
			if injectedStmt.Stdin {
				return true, false, h.handleCopyFromStdinQuery(injectedStmt, h.Conn())
			} else {
				// copying from a file is handled in a single message
				return true, true, h.copyFromFileQuery(injectedStmt)
			}
		case *node.CopyTo:
			// Unlike COPY FROM STDIN, the entire COPY TO STDOUT flow is driven by the server, so it completes
			// within a single message and the server is ready for the next query afterward.
			return true, true, h.handleCopyTo(injectedStmt)
		}
	}
	return false, true, nil
}

func makeCommandComplete(tag string, rows int32) *pgproto3.CommandComplete {
	switch tag {
	case "INSERT", "DELETE", "UPDATE", "MERGE", "SELECT", "CREATE TABLE AS", "MOVE", "FETCH", "COPY":
		if tag == "INSERT" {
			tag = "INSERT 0"
		}
		tag = fmt.Sprintf("%s %d", tag, rows)
	}

	return &pgproto3.CommandComplete{
		CommandTag: []byte(tag),
	}
}

// query runs the given query and sends a CommandComplete message to the client
func (h *ConnectionHandler) query(query ConvertedQuery) error {
	// |rowsAffected| gets altered by the callback below
	rowsAffected := int32(0)

	callback := h.spoolRowsCallback(query, &rowsAffected, false)
	err := h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, query.String, query.AST, callback)
	if err != nil {
		if strings.HasPrefix(err.Error(), "syntax error at position") {
			return errors.Errorf("This statement is not yet supported")
		}
		return err
	}

	return h.send(makeCommandComplete(query.StatementTag, rowsAffected))
}

// spoolRowsCallback returns a callback function that will send RowDescription message,
// then a DataRow message for each row in the result set.
func (h *ConnectionHandler) spoolRowsCallback(query ConvertedQuery, rows *int32, isExecute bool) func(ctx *sql.Context, res *Result) error {
	// IsIUD returns whether the query is either an INSERT, UPDATE, or DELETE query.
	isIUD := query.StatementTag == "INSERT" || query.StatementTag == "UPDATE" || query.StatementTag == "DELETE"

	// The RowDescription message should only be sent once, before any DataRow messages,
	// otherwise some clients will not properly handle results.
	hasSentRowDescription := false
	return func(ctx *sql.Context, res *Result) error {
		sess := dsess.DSessFromSess(ctx.Session)
		for _, notice := range sess.Notices() {
			backendMsg, ok := notice.(pgproto3.BackendMessage)
			if !ok {
				return fmt.Errorf("unexpected notice message type: %T", notice)
			}

			if err := h.send(backendMsg); err != nil {
				return err
			}
		}
		sess.ClearNotices()

		// CALL statement does not return row unless the procedure has OUT parameter, then it returns single row result.
		callWithRowReturned := query.StatementTag == "CALL" && res.RowsAffected != 0

		if returnsRow(query) || callWithRowReturned {
			// EXECUTE does not send RowDescription; instead it should be sent from DESCRIBE prior to it
			if (!isExecute && !hasSentRowDescription) || callWithRowReturned {
				hasSentRowDescription = true
				h.backend.Send(&pgproto3.RowDescription{
					Fields: res.Fields,
				})
			}
			// res.Rows should be length rowsBatch = 128
			for _, row := range res.Rows {
				h.backend.Send(&pgproto3.DataRow{
					Values: row.val,
				})
			}
			err := h.backend.Flush()
			if err != nil {
				return err
			}
		}

		if isIUD {
			*rows = int32(res.RowsAffected)
		} else {
			*rows += int32(len(res.Rows))
		}

		return nil
	}
}

// discardAll handles the DISCARD ALL command
func (h *ConnectionHandler) discardAll(query ConvertedQuery) error {
	err := h.doltgresHandler.ComResetConnection(h.mysqlConn)
	if err != nil {
		return err
	}

	return h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(query.StatementTag),
	})
}

// returnsRow returns whether the query returns set of rows such as SELECT and FETCH statements.
func returnsRow(query ConvertedQuery) bool {
	switch query.StatementTag {
	case "SELECT", "SHOW", "FETCH", "EXPLAIN", "SHOW TABLES", "SHOW CREATE", "SHOW INDEXES FROM TABLE", "SHOW DATABASES", "SHOW SCHEMAS":
		return true
	case "INSERT", "UPDATE", "DELETE":
		return hasReturningClause(query.AST)
	default:
		return false
	}
}

// hasReturningClause return true if |statement| has a RETURNING clause defined.
func hasReturningClause(statement sqlparser.Statement) bool {
	hasReturningClause := false
	sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node := node.(type) {
		case *sqlparser.Update:
			if len(node.Returning) > 0 {
				hasReturningClause = true
			}
			return false, nil
		case *sqlparser.Insert:
			if len(node.Returning) > 0 {
				hasReturningClause = true
			}
			return false, nil
		case *sqlparser.Delete:
			if len(node.Returning) > 0 {
				hasReturningClause = true
			}
		}
		return true, nil
	}, statement)

	return hasReturningClause
}

// ConvertedQuery represents a query that has been converted from the Postgres representation to the Vitess
// representation. String may contain the string version of the converted query. AST will contain the tree
// version of the converted query, and is the recommended form to use. If AST is nil, then use the String version,
// otherwise always prefer to AST.
type ConvertedQuery struct {
	String       string
	AST          vitess.Statement
	StatementTag string
}
