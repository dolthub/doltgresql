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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/ast"
	"github.com/dolthub/doltgresql/server/node"
)

// ConvertedQuery represents PostgreSQL text converted into the engine's statement representation.
type ConvertedQuery struct {
	String       string
	AST          sqlparser.Statement
	StatementTag string
}

// handleQueryOutsideEngine handles statements owned by the connection layer instead of the query engine.
func (h *ConnectionHandler) handleQueryOutsideEngine(query ConvertedQuery, continuation copyContinuation) (handled bool, endOfMessages bool, err error) {
	if handled, err := h.handleTransactionStatement(query); handled || err != nil {
		return handled, true, err
	}

	switch stmt := query.AST.(type) {
	case *sqlparser.Deallocate:
		return true, true, h.deallocatePreparedStatement(stmt.Name, query)
	case sqlparser.InjectedStatement:
		switch injectedStmt := stmt.Statement.(type) {
		case node.DiscardStatement:
			return true, true, h.discardAll(query)
		case *node.CopyFrom:
			if injectedStmt.Stdin {
				return true, false, h.handleCopyFromStdinQuery(injectedStmt, continuation)
			}
			return true, true, h.copyFromFileQuery(injectedStmt)
		case *node.CopyTo:
			return true, true, h.handleCopyTo(injectedStmt)
		}
	}
	return false, true, nil
}

// query runs one converted query and sends its CommandComplete response.
func (h *ConnectionHandler) query(query ConvertedQuery) error {
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

// discardAll resets all session-local resources and reports completion.
func (h *ConnectionHandler) discardAll(query ConvertedQuery) error {
	if h.state.txState != idleTransactionState {
		return pgerror.New(pgcode.ActiveSQLTransaction, "DISCARD ALL cannot run inside a transaction block")
	}
	if err := h.doltgresHandler.ComResetConnection(h.mysqlConn); err != nil {
		return err
	}
	h.state.resetExtendedQueryObjects()
	return h.send(&pgproto3.CommandComplete{CommandTag: []byte("DISCARD ALL")})
}

// spoolRowsCallback returns an engine callback that writes a statement's result messages.
func (h *ConnectionHandler) spoolRowsCallback(query ConvertedQuery, rows *int32, isExecute bool) func(ctx *sql.Context, res *Result) error {
	isIUD := query.StatementTag == "INSERT" || query.StatementTag == "UPDATE" || query.StatementTag == "DELETE"
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

		callWithRowReturned := query.StatementTag == "CALL" && res.RowsAffected != 0
		if returnsRow(query) || callWithRowReturned {
			if (!isExecute && !hasSentRowDescription) || callWithRowReturned {
				hasSentRowDescription = true
				h.backend.Send(&pgproto3.RowDescription{Fields: res.Fields})
			}
			for _, row := range res.Rows {
				h.backend.Send(&pgproto3.DataRow{Values: row.val})
			}
			if err := h.backend.Flush(); err != nil {
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

// convertQuery parses PostgreSQL text and converts it into engine statements.
func convertQuery(query string) ([]ConvertedQuery, error) {
	s, err := parser.Parse(query)
	if err != nil {
		return nil, err
	}
	if len(s) == 0 {
		return []ConvertedQuery{{String: query}}, nil
	}
	converted := make([]ConvertedQuery, len(s))
	for i := range s {
		vitessAST, err := ast.Convert(s[i])
		stmtTag := s[i].AST.StatementTag()
		if err != nil {
			return nil, err
		}
		if vitessAST == nil {
			converted[i] = ConvertedQuery{String: s[i].AST.String(), StatementTag: stmtTag}
		} else {
			converted[i] = ConvertedQuery{String: query, AST: vitessAST, StatementTag: stmtTag}
		}
	}
	return converted, nil
}

// makeCommandComplete constructs PostgreSQL's command tag for a completed statement.
func makeCommandComplete(tag string, rows int32) *pgproto3.CommandComplete {
	switch tag {
	case "INSERT":
		// PostgreSQL retains an object-ID field in INSERT tags for protocol compatibility,
		// but always reports zero.
		tag = fmt.Sprintf("INSERT 0 %d", rows)
	case "DELETE", "UPDATE", "MERGE", "SELECT", "CREATE TABLE AS", "MOVE", "FETCH", "COPY":
		tag = fmt.Sprintf("%s %d", tag, rows)
	}
	return &pgproto3.CommandComplete{CommandTag: []byte(tag)}
}

// returnsRow reports whether a statement produces a row set.
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

// hasReturningClause reports whether a data-modification statement contains RETURNING.
func hasReturningClause(statement sqlparser.Statement) bool {
	hasReturning := false
	sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node := node.(type) {
		case *sqlparser.Update:
			hasReturning = len(node.Returning) > 0
			return !hasReturning, nil
		case *sqlparser.Insert:
			hasReturning = len(node.Returning) > 0
			return !hasReturning, nil
		case *sqlparser.Delete:
			hasReturning = len(node.Returning) > 0
		}
		return true, nil
	}, statement)
	return hasReturning
}
