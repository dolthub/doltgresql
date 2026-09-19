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
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/server/node"
)

// simpleQueryExecution tracks the statements remaining in one simple-protocol Query message.
type simpleQueryExecution struct {
	statements    []ConvertedQuery
	nextStatement int
}

// handleQuery starts execution of one simple-protocol Query message.
func (h *ConnectionHandler) handleQuery(message *pgproto3.Query) (endOfMessages bool, err error) {
	queries, err := h.convertQuery(message.String)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("Error parsing query: %+v\n", err)
		}
		return true, err
	}

	// A simple Query message destroys the unnamed prepared statement and unnamed portal.
	h.state.extendedQueryObjects.clearUnnamed()
	if len(queries) == 1 && queries[0].AST == nil {
		return true, h.send(&pgproto3.EmptyQueryResponse{})
	}
	return h.resumeSimpleQuery(&simpleQueryExecution{statements: queries})
}

// resumeSimpleQuery executes statements until they finish or COPY FROM STDIN needs client data.
func (h *ConnectionHandler) resumeSimpleQuery(execution *simpleQueryExecution) (endOfMessages bool, err error) {
	if execution == nil {
		return true, errors.New("no active simple query to resume")
	}

	// Multiple statements in one Query message run in an implicit transaction block. Transaction-control statements
	// may promote or end that block through handleTransactionStatement.
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

		handled, statementComplete, err := h.handleQueryOutsideEngine(query, execution)
		if err != nil {
			return true, err
		}
		if handled {
			if !statementComplete {
				return false, nil
			}
			continue
		}

		if implicitTransactionControl && i == len(execution.statements)-1 && !h.state.txState.inExplicitTransactionBlock() {
			ctx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
			if err != nil {
				return false, err
			}
			ctx.SetIgnoreAutoCommit(false)
		}

		injected, isInjected := query.AST.(sqlparser.InjectedStatement)
		_, isDo := injected.Statement.(*node.Do)
		if !implicitTransactionControl && isInjected && isDo {
			if err = h.startImplicitTransaction(query); err != nil {
				return true, err
			}
		}

		if err = h.query(query); err != nil {
			return true, err
		}
		if !implicitTransactionControl && isInjected && isDo {
			if err = h.commitImplicitTransaction(); err != nil {
				return true, err
			}
		}
	}

	if implicitTransactionControl {
		if err = h.commitImplicitTransaction(); err != nil {
			return false, err
		}
	}
	return true, nil
}
