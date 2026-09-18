// Copyright 2023 Dolthub, Inc.
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
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqlserver"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
)

// ConnectionHandler is responsible for the entire lifecycle of a user connection: receiving messages they send,
// executing queries, sending the correct messages in return, and terminating the connection when appropriate.
type ConnectionHandler struct {
	mysqlConn       *mysql.Conn
	doltgresHandler *DoltgresHandler
	backend         *pgproto3.Backend
	state           connectionState
	convertOptions  ast.ConvertOptions
}

// messageAction describes the connection-loop action after handling one frontend message.
type messageAction byte

const (
	continueMessages messageAction = iota
	sendReadyForQuery
	closeConnection
)

// messageResult tells the connection loop how to proceed after handling one frontend message.
type messageResult struct {
	action messageAction
	err    error
}

// continueResult returns a successful result that keeps reading frontend messages.
func continueResult() messageResult {
	return messageResult{action: continueMessages}
}

// readyResult returns a result that completes the current frontend operation.
func readyResult(err error) messageResult {
	return messageResult{action: sendReadyForQuery, err: err}
}

// closeResult returns a result that closes the connection.
func closeResult() messageResult {
	return messageResult{action: closeConnection}
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
		mysqlConn:       mysqlConn,
		doltgresHandler: doltgresHandler,
		backend:         pgproto3.NewBackend(conn, conn),
		state:           newConnectionState(),
		convertOptions:  convertOptions,
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

// receiveMessage reads a single message off the connection and processes it, returning an error if no message could be
// received from the connection. Otherwise, (a message is received successfully), the message is processed and any
// error is handled appropriately. The return value indicates whether the connection should be closed.
func (h *ConnectionHandler) receiveMessage() (stop bool, err error) {
	result := continueResult()
	// For the time being, we handle panics in this function and treat them the same as errors so that they don't
	// forcibly close the connection. Contrast this with the panic handling logic in HandleConnection, where we treat any
	// panic as unrecoverable to the connection. As we fill out the implementation, we can revisit this decision and
	// rethink our posture over whether panics should terminate a connection.
	if HandlePanics {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				logrus.Errorf("Listener recovered panic: %v: %s", r, stackTrace)

				h.handleMessageError(errors.Errorf("receiveMessage recovered panic: %v: %s",
					r, stackTrace))
				stop = h.state.mode == closingConnectionMode
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

	result = h.handleMessage(msg)
	if result.err != nil {
		h.handleMessageError(result.err)
	} else if result.action == sendReadyForQuery {
		h.endOfMessages(nil)
	}

	return result.action == closeConnection, nil
}

// handleMessage routes a frontend message according to the exclusive mode that owns the connection.
func (h *ConnectionHandler) handleMessage(msg pgproto3.Message) messageResult {
	if _, ok := msg.(*pgproto3.Terminate); ok {
		return closeResult()
	}

	switch h.state.mode {
	case copyInConnectionMode:
		if h.state.activeCopy == nil {
			h.state.closeConnection()
			return messageResult{action: closeConnection, err: errors.New("COPY mode has no active operation")}
		}
		return h.handleCopyMessage(h.state.activeCopy, msg)
	case closingConnectionMode:
		return closeResult()
	case discardUntilSyncConnectionMode:
		if _, ok := msg.(*pgproto3.Sync); ok {
			h.state.finishExtendedQueryMode()
			return readyResult(h.commitImplicitTransaction())
		}
		return continueResult()
	case readyConnectionMode, extendedQueryConnectionMode:
		return h.handleNormalMessage(msg)
	default:
		h.state.closeConnection()
		return messageResult{action: closeConnection, err: errors.New("invalid connection mode")}
	}
}

// handleNormalMessage handles messages outside COPY and extended-protocol error recovery.
func (h *ConnectionHandler) handleNormalMessage(msg pgproto3.Message) messageResult {
	switch message := msg.(type) {
	case *pgproto3.Sync:
		// Sync closes an implicit transaction block, committing it. An explicit transaction block (opened with
		// BEGIN) is not affected by Sync, and remains open.
		h.state.finishExtendedQueryMode()
		return readyResult(h.commitImplicitTransaction())
	case *pgproto3.Flush:
		// We don't buffer output, so Flush is a no-op
		return continueResult()
	case *pgproto3.Query:
		endOfMessages, err := h.handleQuery(message)
		return messageResultForCompletion(endOfMessages, err)
	case *pgproto3.Parse:
		h.state.beginExtendedQueryMode()
		return messageResult{err: h.handleParse(message)}
	case *pgproto3.Describe:
		h.state.beginExtendedQueryMode()
		return messageResult{err: h.handleDescribe(message)}
	case *pgproto3.Bind:
		h.state.beginExtendedQueryMode()
		return messageResult{err: h.handleBind(message)}
	case *pgproto3.Execute:
		h.state.beginExtendedQueryMode()
		return messageResult{err: h.handleExecute(message)}
	case *pgproto3.Close:
		h.state.beginExtendedQueryMode()
		return messageResult{err: h.handleClose(message)}
	case *pgproto3.CopyData:
		// PostgreSQL drops COPY messages that arrive after COPY has already failed and left copy-in mode.
		return continueResult()
	case *pgproto3.CopyDone:
		return continueResult()
	case *pgproto3.CopyFail:
		return continueResult()
	default:
		return readyResult(errors.Errorf(`unhandled message "%t"`, message))
	}
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

// messageResultForCompletion converts legacy statement completion into a connection-loop action.
func messageResultForCompletion(complete bool, err error) messageResult {
	if complete {
		return readyResult(err)
	}
	return messageResult{err: err}
}

// Send sends the given message over the connection.
func (h *ConnectionHandler) send(message pgproto3.BackendMessage) error {
	h.backend.Send(message)
	return h.backend.Flush()
}

// setConn replaces the underlying connection and rebuilds the protocol backend around it.
func (h *ConnectionHandler) setConn(conn net.Conn) {
	h.mysqlConn.Conn = conn
	h.backend = pgproto3.NewBackend(conn, conn)
}
