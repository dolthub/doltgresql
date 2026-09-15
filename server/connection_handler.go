// Copyright 2024 Dolthub, Inc.
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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqlserver"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/ast"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"
)

// ConnectionHandler is responsible for the entire lifecycle of a user connection: receiving messages they send,
// executing queries, sending the correct messages in return, and terminating the connection when appropriate.
type ConnectionHandler struct {
	mysqlConn          *mysql.Conn
	preparedStatements map[string]PreparedStatementData
	portals            map[string]PortalData
	doltgresHandler    *DoltgresHandler
	backend            *pgproto3.Backend
	convertOptions     ast.ConvertOptions

	waitForSync bool
	// copyFromStdinState is set when this connection is in the COPY FROM STDIN mode, meaning it is waiting on
	// COPY DATA messages from the client to import data into tables.
	copyFromStdinState *copyFromStdinState
	// activeSimpleQuery is the current multi-statement simple query execution.
	activeSimpleQuery *simpleQueryExecution

	// transactionState is the current transaction state of the connection, which is one of:
	// Idle (no transaction block is in progress)
	// Explicit (an explicit transaction block is in progress, opened by a BEGIN statement)
	// Implicit (an implicit transaction block is in progress, opened by a multi-statement Query message or an extended query protocol)
	// Failed (an error occurred inside an explicit transaction block, and all statements are rejected until the client ends the transaction block)
	// See https://www.postgresql.org/docs/current/protocol-flow.html for the full ruleset.
	transactionState transactionState
}

// Set this env var to disable panic handling in the connection, which is useful when debugging a panic
const disablePanicHandlingEnvVar = "DOLT_PGSQL_PANIC"

// HandlePanics determines whether panics should be handled in the connection handler. See |disablePanicHandlingEnvVar|.
var HandlePanics = true

func init() {
	if _, ok := os.LookupEnv(disablePanicHandlingEnvVar); ok {
		HandlePanics = false
	} else {
		// This checks if the Go debugger is attached, so that we can disable panic catching automatically
		pid := os.Getppid()
		for pid != 0 {
			p, err := ps.FindProcess(pid)
			if err != nil || p == nil {
				break
			} else if strings.HasPrefix(p.Executable(), "dlv") {
				HandlePanics = false
				break
			} else {
				pid = p.PPid()
			}
		}
	}
}

// NewConnectionHandler returns a new ConnectionHandler for the connection provided
func NewConnectionHandler(conn net.Conn, handler mysql.Handler, sel server.ServerEventListener) *ConnectionHandler {
	mysqlConn := &mysql.Conn{
		Conn:        conn,
		PrepareData: make(map[uint32]*mysql.PrepareData),
	}
	mysqlConn.ConnectionID = atomic.AddUint32(&connectionIDCounter, 1)

	// Postgres has a two-stage procedure for prepared queries. First the query is parsed via a |Parse| message, and
	// the result is stored in the |preparedStatements| map by the name provided. Then one or more |Bind| messages
	// provide parameters for the query, and the result is stored in |portals|. Finally, a call to |Execute| executes
	// the named portal.
	preparedStatements := make(map[string]PreparedStatementData)
	portals := make(map[string]PortalData)

	// TODO: possibly should define engine and session manager ourselves
	//  instead of depending on the GetRunningServer method.
	server := sqlserver.GetRunningServer()
	convertOptions := ast.ConvertOptions{}
	if postgresParser, ok := server.Engine.Parser.(*psql.PostgresParser); ok {
		convertOptions = postgresParser.ConvertOptions()
	}
	doltgresHandler := &DoltgresHandler{
		e:                 server.Engine,
		sm:                server.SessionManager(),
		readTimeout:       0,     // cfg.ConnReadTimeout,
		encodeLoggedQuery: false, // cfg.EncodeLoggedQuery,
		pgTypeMap:         pgtype.NewMap(),
	}
	if sel != nil {
		doltgresHandler.sel = sel
	}

	return &ConnectionHandler{
		mysqlConn:          mysqlConn,
		preparedStatements: preparedStatements,
		portals:            portals,
		doltgresHandler:    doltgresHandler,
		backend:            pgproto3.NewBackend(conn, conn),
		transactionState:   idleTransactionState,
		convertOptions:     convertOptions,
	}
}

// HandleConnection handles a connection's session, reading messages, executing queries, and sending responses.
// Expected to run in a goroutine per connection.
func (h *ConnectionHandler) HandleConnection() {
	var returnErr error
	if HandlePanics {
		defer func() {
			if r := recover(); r != nil {
				// debug.Stack() here prints the stack trace of the original panic, not the lexical stack of this defer function
				stackTrace := string(debug.Stack())
				logrus.Errorf("Listener recovered panic: %v: %s", r, stackTrace)

				var eomErr error
				if returnErr != nil {
					eomErr = returnErr
				} else {
					eomErr = errors.Errorf("Listener recovered panic: %v: %s", r, stackTrace)
				}

				// Sending eom can panic, which means we must recover again
				defer func() {
					if r := recover(); r != nil {
						logrus.Errorf("Listener recovered panic: %v: %s", r, string(debug.Stack()))
					}
				}()
				h.endOfMessages(eomErr)
			}

			if returnErr != nil {
				fmt.Println(returnErr.Error())
			}
		}()
	}
	defer func() {
		if err := h.Conn().Close(); err != nil {
			fmt.Printf("Failed to properly close connection:\n%v\n", err)
		}
	}()
	h.doltgresHandler.NewConnection(h.mysqlConn)
	defer func() {
		h.doltgresHandler.ConnectionClosed(h.mysqlConn)
	}()

	if proceed, err := h.handleStartup(); err != nil || !proceed {
		returnErr = err
		return
	}

	// Main session loop: read messages one at a time off the connection until we receive a |Terminate| message, in
	// which case we hang up, or the connection is closed by the client, which generates an io.EOF from the connection.
	for {
		stop, err := h.receiveMessage()
		if err != nil {
			returnErr = err
			break
		}

		if stop {
			break
		}
	}
}

// Conn returns the underlying net.Conn for this connection.
func (h *ConnectionHandler) Conn() net.Conn {
	return h.mysqlConn.Conn
}

// setConn sets a new underlying net.Conn for this connection.
func (h *ConnectionHandler) setConn(conn net.Conn) {
	h.mysqlConn.Conn = conn
	h.backend = pgproto3.NewBackend(conn, conn)
}

// receiveMessage reads a single message off the connection and processes it, returning an error if no message could be
// received from the connection. Otherwise, (a message is received successfully), the message is processed and any
// error is handled appropriately. The return value indicates whether the connection should be closed.
func (h *ConnectionHandler) receiveMessage() (bool, error) {
	var endOfMessages bool
	// For the time being, we handle panics in this function and treat them the same as errors so that they don't
	// forcibly close the connection. Contrast this with the panic handling logic in HandleConnection, where we treat any
	// panic as unrecoverable to the connection. As we fill out the implementation, we can revisit this decision and
	// rethink our posture over whether panics should terminate a connection.
	if HandlePanics {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				logrus.Errorf("Listener recovered panic: %v: %s", r, stackTrace)

				eomErr := errors.Errorf("receiveMessage recovered panic: %v: %s", r, stackTrace)
				if !endOfMessages && h.waitForSync {
					if syncErr := h.discardToSync(); syncErr != nil {
						fmt.Println(syncErr.Error())
					}
				}
				h.endOfMessages(eomErr)
			}
		}()
	}

	msg, err := h.backend.Receive()
	if err != nil {
		return false, errors.Errorf("error receiving message: %w", err)
	}

	if m, ok := msg.(json.Marshaler); ok && logrus.IsLevelEnabled(logrus.DebugLevel) {
		msgInfo, err := m.MarshalJSON()
		if err != nil {
			return false, err
		}
		logrus.Debugf("Received message: %s", string(msgInfo))
	} else {
		logrus.Debugf("Received message: %t", msg)
	}

	var stop bool
	stop, endOfMessages, err = h.handleMessage(msg)
	if err != nil {
		if !endOfMessages && h.waitForSync {
			if syncErr := h.discardToSync(); syncErr != nil {
				fmt.Println(syncErr.Error())
			}
		}
		h.endOfMessages(err)
	} else if endOfMessages {
		h.endOfMessages(nil)
	}

	return stop, nil
}

// handleMessages processes the message provided and returns status flags indicating what the connection should do next.
// If the |stop| response parameter is true, it indicates that the connection should be closed by the caller. If the
// |endOfMessages| response parameter is true, it indicates that no more messages are expected for the current operation
// and a READY FOR QUERY message should be sent back to the client, so it can send the next query.
func (h *ConnectionHandler) handleMessage(msg pgproto3.Message) (stop, endOfMessages bool, err error) {
	switch message := msg.(type) {
	case *pgproto3.Terminate:
		return true, false, nil
	case *pgproto3.Sync:
		h.waitForSync = false
		// Sync closes an implicit transaction block, committing it. An explicit transaction block (opened with
		// BEGIN) is not affected by Sync, and remains open.
		return false, true, h.commitImplicitTransaction()
	case *pgproto3.Flush:
		// We don't buffer output, so Flush is a no-op
		return false, false, nil
	case *pgproto3.Query:
		if h.activeSimpleQuery != nil {
			return false, true, errors.New("query received while a COPY FROM STDIN operation is in progress")
		}
		endOfMessages, err = h.handleQuery(message)
		return false, endOfMessages, err
	case *pgproto3.Parse:
		return false, false, h.handleParse(message)
	case *pgproto3.Describe:
		return false, false, h.handleDescribe(message)
	case *pgproto3.Bind:
		return false, false, h.handleBind(message)
	case *pgproto3.Execute:
		return false, false, h.handleExecute(message)
	case *pgproto3.Close:
		if message.ObjectType == 'S' {
			delete(h.preparedStatements, message.Name)
		} else {
			delete(h.portals, message.Name)
		}
		return false, false, h.send(&pgproto3.CloseComplete{})
	case *pgproto3.CopyData:
		return h.handleCopyData(message)
	case *pgproto3.CopyDone:
		stop, endOfMessages, err := h.handleCopyDone(message)
		if stop || err != nil || !endOfMessages || h.activeSimpleQuery == nil {
			return stop, endOfMessages, err
		}
		endOfMessages, err = h.resumeSimpleQuery()
		return false, endOfMessages, err
	case *pgproto3.CopyFail:
		return h.handleCopyFail(message)
	default:
		return false, true, errors.Errorf(`unhandled message "%t"`, message)
	}
}

// handleCopyData handles the COPY DATA message, by loading the data sent from the client. The |stop| response parameter
// is true if the connection handler should shut down the connection, |endOfMessages| is true if no more COPY DATA
// messages are expected, and the server should tell the client that it is ready for the next query, and |err| contains
// any error that occurred while processing the COPY DATA message.
func (h *ConnectionHandler) handleCopyData(message *pgproto3.CopyData) (stop bool, endOfMessages bool, err error) {
	if h.copyFromStdinState != nil && h.copyFromStdinState.copyErr != nil {
		// A previous chunk of this COPY operation failed and its work was rolled back, so discard the remaining
		// data until the client ends the operation with a COPY DONE or COPY FAIL message. Processing further
		// chunks would insert rows for an operation that has already been rejected.
		return false, false, nil
	}
	copyFromData := bytes.NewReader(message.Data)
	stop, endOfMessages, err = h.handleCopyDataHelper(h.copyFromStdinState, copyFromData)
	if err != nil && h.copyFromStdinState != nil {
		h.copyFromStdinState.copyErr = err
		h.rollbackCopyTransaction(h.copyFromStdinState.startedTransaction)
	}
	return stop, endOfMessages, err
}

// handleCopyFail handles a COPY FAIL message by aborting the in-progress COPY DATA operation.  The |stop| response
// parameter is true if the connection handler should shut down the connection, |endOfMessages| is true if no more
// COPY DATA messages are expected, and the server should tell the client that it is ready for the next query, and
// |err| contains any error that occurred while processing the COPY DATA message.
func (h *ConnectionHandler) handleCopyFail(message *pgproto3.CopyFail) (stop bool, endOfMessages bool, err error) {
	if h.copyFromStdinState == nil {
		return false, true,
			errors.Errorf("COPY FAIL message received without a COPY FROM STDIN operation in progress")
	}

	dataLoader := h.copyFromStdinState.dataLoader
	if dataLoader == nil {
		return false, true,
			errors.Errorf("no data loader found for COPY FROM STDIN operation")
	}

	startedTransaction := h.copyFromStdinState.startedTransaction
	h.copyFromStdinState = nil
	// The client aborted the operation, so any rows loaded by chunks that were already processed must not persist
	h.rollbackCopyTransaction(startedTransaction)
	return false, true, pgerror.New(pgcode.QueryCanceled, message.Message)
}

// convertQuery takes the given Postgres query, and converts it as an ast.ConvertedQuery that will work with the handler.
// If the query string contains multiple queries, then multiple ConvertedQuery will be returned.
func (h *ConnectionHandler) convertQuery(query string) ([]ConvertedQuery, error) {
	s, err := parser.Parse(query)
	if err != nil {
		return nil, err
	}
	if len(s) == 0 {
		return []ConvertedQuery{{String: query}}, nil
	}
	converted := make([]ConvertedQuery, len(s))
	for i := range s {
		vitessAST, err := ast.ConvertWithOptions(s[i], h.convertOptions)
		stmtTag := s[i].AST.StatementTag()
		if err != nil {
			return nil, err
		}
		if vitessAST == nil {
			converted[i] = ConvertedQuery{
				String:       s[i].AST.String(),
				StatementTag: stmtTag,
			}
		} else {
			converted[i] = ConvertedQuery{
				String:       query,
				AST:          vitessAST,
				StatementTag: stmtTag,
			}
		}
	}
	return converted, nil
}

// DiscardToSync discards all messages in the buffer until a Sync has been reached. If a Sync was never sent, then this
// may cause the connection to lock until the client send a Sync, as their request structure was malformed.
func (h *ConnectionHandler) discardToSync() error {
	for {
		message, err := h.backend.Receive()
		if err != nil {
			return err
		}

		if _, ok := message.(*pgproto3.Sync); ok {
			return nil
		}
	}
}

// Send sends the given message over the connection.
func (h *ConnectionHandler) send(message pgproto3.BackendMessage) error {
	h.backend.Send(message)
	return h.backend.Flush()
}
