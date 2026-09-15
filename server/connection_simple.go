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
	"github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
)

// simpleQueryExecution tracks the statements being executed in a multi-statement simple query.
type simpleQueryExecution struct {
	statements    []ConvertedQuery
	nextStatement int
}

// handleQuery handles a query message, and returns a boolean flag, |endOfMessages| indicating if no other messages are
// expected as part of this query, in which case the server will send a READY FOR QUERY message back to the client so
// that it can send its next query.
func (h *ConnectionHandler) handleQuery(message *pgproto3.Query) (endOfMessages bool, err error) {
	queries, err := h.convertQuery(message.String)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("Error parsing query: %+v\n", err)
		}
		return true, err
	}

	// A query message destroys the unnamed statement and the unnamed portal
	delete(h.preparedStatements, "")
	delete(h.portals, "")

	var handled bool
	if len(queries) == 1 {
		// empty query special case
		if queries[0].AST == nil {
			return true, h.send(&pgproto3.EmptyQueryResponse{})
		}
		if err = h.rejectStatementIfTransactionFailed(queries[0]); err != nil {
			return true, err
		}
		handled, endOfMessages, err = h.handleQueryOutsideEngine(queries[0])
		if handled {
			return endOfMessages, err
		}
		injected, isInjected := queries[0].AST.(sqlparser.InjectedStatement)
		_, isDo := injected.Statement.(*node.Do)
		if isInjected && isDo {
			if err = h.startImplicitTransaction(queries[0]); err != nil {
				return true, err
			}
			if err = h.query(queries[0]); err != nil {
				return true, err
			}
			return true, h.commitImplicitTransaction()
		}
		return true, h.query(queries[0])
	}

	h.activeSimpleQuery = &simpleQueryExecution{statements: queries}
	return h.resumeSimpleQuery()
}

// resumeSimpleQuery executes statements until they finish or a COPY FROM STDIN statement needs client data.
func (h *ConnectionHandler) resumeSimpleQuery() (endOfMessages bool, err error) {
	execution := h.activeSimpleQuery
	if execution == nil {
		return true, errors.New("no active simple query to resume")
	}
	defer func() {
		if endOfMessages || err != nil {
			h.activeSimpleQuery = nil
		}
	}()

	// Multiple statements in a single Query message run in an implicit transaction block, which is committed
	// after the last statement and rolled back if any statement errors (in which case the remaining statements
	// are never executed). Transaction control statements within the message alter this behavior: see
	// handleQueryOutsideEngine for how BEGIN, COMMIT, and ROLLBACK interact with implicit transaction blocks.
	implicitTransactionControl := len(execution.statements) > 1
	for execution.nextStatement < len(execution.statements) {
		i := execution.nextStatement
		query := execution.statements[i]
		execution.nextStatement++
		if err = h.rejectStatementIfTransactionFailed(query); err != nil {
			return true, err
		}
		if implicitTransactionControl {
			if err = h.startImplicitTransaction(query); err != nil {
				return true, err
			}
		}

		var handled bool
		var statementComplete bool
		handled, statementComplete, err = h.handleQueryOutsideEngine(query)
		if err != nil {
			return true, err
		}
		if handled {
			if !statementComplete {
				return false, nil
			}
			continue
		}

		// Single statements will always be auto-committed, unless they are inside an explicit transaction block.
		// For multi-statement queries, we start an implicit transaction block before the first statement and commit
		// it on the last statement. This involves manipulating the session's auto-commit behavior so that the engine
		// automatically commits only the final statement. This is cheaper than running BEGIN and COMMIT statements
		// separately through the engine, and has the same effect.
		if implicitTransactionControl {
			if i == len(execution.statements)-1 && !h.transactionState.inExplicitTransactionBlock() {
				ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
				if err != nil {
					return false, err
				}
				ctx.SetIgnoreAutoCommit(false)
			}
		}

		err = h.query(query)
		if err != nil {
			return true, err
		}
	}

	// For some statement sequences, a final implicit COMMIT may be necessary
	if implicitTransactionControl {
		err = h.commitImplicitTransaction()
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
