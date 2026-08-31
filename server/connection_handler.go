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
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqlserver"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	psql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/ast"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ConnectionHandler is responsible for the entire lifecycle of a user connection: receiving messages they send,
// executing queries, sending the correct messages in return, and terminating the connection when appropriate.
type ConnectionHandler struct {
	mysqlConn          *mysql.Conn
	preparedStatements map[string]PreparedStatementData
	portals            map[string]PortalData
	doltgresHandler    *DoltgresHandler
	backend            *pgproto3.Backend

	waitForSync bool
	// copyFromStdinState is set when this connection is in the COPY FROM STDIN mode, meaning it is waiting on
	// COPY DATA messages from the client to import data into tables.
	copyFromStdinState *copyFromStdinState

	// transactionState is the current transaction state of the connection, which is one of:
	// Idle (no transaction block is in progress)
	// Explicit (an explicit transaction block is in progress, opened by a BEGIN statement)
	// Implicit (an implicit transaction block is in progress, opened by a multi-statement Query message or an extended query protocol)
	// Failed (an error occurred inside an explicit transaction block, and all statements are rejected until the client ends the transaction block)
	// See https://www.postgresql.org/docs/current/protocol-flow.html for the full ruleset.
	transactionState transactionState
}

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

// handleStartup handles the entire startup routine, including SSL requests, authentication, etc. Returns false if the
// connection has been terminated, or if we should not proceed with the message loop.
func (h *ConnectionHandler) handleStartup() (bool, error) {
	startupMessage, err := h.backend.ReceiveStartupMessage()
	if err == io.EOF {
		// Receiving EOF means that the connection has terminated, so we should just return
		return false, nil
	} else if err != nil {
		return false, errors.Errorf("error receiving startup message: %w", err)
	}

	switch sm := startupMessage.(type) {
	case *pgproto3.StartupMessage:
		if err = h.handleAuthentication(sm); err != nil {
			return false, err
		}
		if err = h.sendClientStartupMessages(); err != nil {
			return false, err
		}
		if err = h.chooseInitialParameters(sm); err != nil {
			// A startup parameter (e.g. datestyle, timezone) failed validation. Without an explicit
			// ErrorResponse here, the client only sees the connection drop as an unexpected EOF instead
			// of the reason its StartupMessage was rejected.
			_ = h.send(&pgproto3.ErrorResponse{
				Severity: string(ErrorResponseSeverity_Fatal),
				Code:     "22023", // invalid_parameter_value
				Message:  err.Error(),
				Routine:  "InitPostgres",
			})
			return false, err
		}
		return true, h.send(&pgproto3.ReadyForQuery{
			TxStatus: byte(ReadyForQueryTransactionIndicator_Idle),
		})
	case *pgproto3.SSLRequest:
		hasCertificate := len(certificate.Certificate) > 0
		var performSSL = []byte("N")
		if hasCertificate {
			performSSL = []byte("S")
		}
		_, err = h.Conn().Write(performSSL)
		if err != nil {
			return false, errors.Errorf("error sending SSL request: %w", err)
		}
		// If we have a certificate and the client has asked for SSL support, then we switch here.
		// This involves swapping out our underlying net connection for a new one.
		// We can't start in SSL mode, as the client does not attempt the handshake until after our response.
		if hasCertificate {
			h.setConn(tls.Server(h.Conn(), &tls.Config{
				Certificates: []tls.Certificate{certificate},
			}))
		}
		return h.handleStartup()
	case *pgproto3.GSSEncRequest:
		// we don't support GSSAPI
		_, err = h.Conn().Write([]byte("N"))
		if err != nil {
			return false, errors.Errorf("error sending response to GSS Enc Request: %w", err)
		}
		return h.handleStartup()
	default:
		return false, errors.Errorf("terminating connection: unexpected start message: %#v", startupMessage)
	}
}

// sendClientStartupMessages sends introductory messages to the client and returns any error
func (h *ConnectionHandler) sendClientStartupMessages() error {
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "server_version",
		Value: "15.17",
	}); err != nil {
		return err
	}
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "client_encoding",
		Value: "UTF8",
	}); err != nil {
		return err
	}
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "standard_conforming_strings",
		Value: "on",
	}); err != nil {
		return err
	}
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "in_hot_standby",
		Value: "off",
	}); err != nil {
		return err
	}
	return h.send(&pgproto3.BackendKeyData{
		ProcessID: processID,
		SecretKey: make([]byte, 4), // TODO: this should represent an ID that can uniquely identify this connection, so that CancelRequest will work
	})
}

// chooseInitialParameters attempts to choose the initial parameter settings for the connection,
// if one is specified in the startup message provided.
func (h *ConnectionHandler) chooseInitialParameters(startupMessage *pgproto3.StartupMessage) error {
	postgresParser := psql.PostgresParser{}
	for name, value := range startupMessage.Parameters {
		// TODO: handle other parameters defined in StartupMessage
		switch strings.ToLower(name) {
		case "datestyle":
			err := h.doltgresHandler.InitSessionParameterDefault(context.Background(), h.mysqlConn, "DateStyle", value)
			if err != nil {
				return err
			}
		case "timezone":
			// timezone is set via a real SET statement rather than InitSessionParameterDefault because we want
			// this value set for the current session, but NOT set as the default for all sessions.
			setStmt := fmt.Sprintf("SET timezone TO '%s';", strings.ReplaceAll(value, "'", "''"))
			parsed, err := postgresParser.ParseSimple(setStmt)
			if err != nil {
				return err
			}
			err = h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, setStmt, parsed, func(_ *sql.Context, _ *Result) error {
				return nil
			})
			if err != nil {
				return err
			}
		}
	}
	// Set the initial database. Postgres has no concept of a session without a current database: if the client
	// doesn't specify one, it defaults to the username (matching libpq). Either way, if the resolved database
	// doesn't exist we must reject the connection rather than proceed with a database-less session, which would
	// break assumptions throughout the engine.
	db, ok := startupMessage.Parameters["database"]
	if !ok || len(db) == 0 {
		db = h.mysqlConn.User
	}
	useStmt := fmt.Sprintf("SET database TO '%s';", db)
	parsed, err := postgresParser.ParseSimple(useStmt)
	if err != nil {
		return err
	}
	err = h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, useStmt, parsed, func(_ *sql.Context, _ *Result) error {
		return nil
	})
	if err != nil {
		_ = h.send(&pgproto3.ErrorResponse{
			Severity: string(ErrorResponseSeverity_Fatal),
			Code:     "3D000",
			Message:  fmt.Sprintf(`database "%s" does not exist`, db),
			Routine:  "InitPostgres",
		})
		return err
	}
	return nil
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
		return h.handleCopyDone(message)
	case *pgproto3.CopyFail:
		return h.handleCopyFail(message)
	default:
		return false, true, errors.Errorf(`unhandled message "%t"`, message)
	}
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
		return true, h.query(queries[0])
	}

	// Multiple statements in a single Query message run in an implicit transaction block, which is committed
	// after the last statement and rolled back if any statement errors (in which case the remaining statements
	// are never executed). Transaction control statements within the message alter this behavior: see
	// handleQueryOutsideEngine for how BEGIN, COMMIT, and ROLLBACK interact with implicit transaction blocks.
	implicitTransactionControl := len(queries) > 1
	for i, query := range queries {
		if err = h.rejectStatementIfTransactionFailed(query); err != nil {
			return true, err
		}

		handled, _, err = h.handleQueryOutsideEngine(query)
		if err != nil {
			return true, err
		}
		if handled {
			continue
		}

		// Single statements will always be auto-committed, unless they are inside an explicit transaction block.
		// For multi-statement queries, we start an implicit transaction block before the first statement and commit
		// it on the last statement. This involves manipulating the session's auto-commit behavior so that the engine
		// automatically commits only the final statement. This is cheaper than running BEGIN and COMMIT statements
		// separately through the engine, and has the same effect.
		if implicitTransactionControl {
			if err = h.startImplicitTransaction(query); err != nil {
				return true, err
			}
			if i == len(queries)-1 && !h.transactionState.inExplicitTransactionBlock() {
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
			if err := h.runEngineTransactionControl("ROLLBACK"); err != nil {
				return true, true, err
			}
			return true, true, h.send(&pgproto3.CommandComplete{CommandTag: []byte("ROLLBACK")})
		}
		// A COMMIT closes the current transaction block, whether explicit or implicit. Any statements that
		// follow it in the same Query message (or extended-query batch) run in a new implicit transaction block.
		h.transactionState = idleTransactionState
	case *sqlparser.Rollback:
		// Like COMMIT, a ROLLBACK closes the current transaction block, whether explicit, implicit, or failed.
		h.transactionState = idleTransactionState
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
		}
	}
	return false, true, nil
}

// handleParse handles a parse message, returning any error that occurs
func (h *ConnectionHandler) handleParse(message *pgproto3.Parse) error {
	h.waitForSync = true

	// TODO: "Named prepared statements must be explicitly closed before they can be redefined by another Parse message, but this is not required for the unnamed statement"
	queries, err := h.convertQuery(message.Query)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("Error parsing query: %+v\n", err)
		}
		return err
	}
	if len(queries) != 1 {
		return errors.Errorf("cannot insert multiple commands into a prepared statement")
	}
	query := queries[0]

	if err = h.rejectStatementIfTransactionFailed(query); err != nil {
		return err
	}

	if query.AST == nil {
		// special case: empty query
		h.preparedStatements[message.Name] = PreparedStatementData{
			Query: query,
		}
		return nil
	}

	ctx, err := h.doltgresHandler.sm.NewContextWithQuery(context.Background(), h.mysqlConn, query.String)
	if err != nil {
		return err
	}
	parsedQuery, fields, err := h.doltgresHandler.ComPrepareParsed(ctx, h.mysqlConn, query.String, query.AST)
	if err != nil {
		return err
	}

	analyzedPlan, ok := parsedQuery.(sql.Node)
	if !ok {
		return errors.Errorf("expected a sql.Node, got %T", parsedQuery)
	}

	// A valid Parse message must have ParameterObjectIDs if there are any binding variables.
	bindVarTypes := message.ParameterOIDs

	// Clients can specify an OID of zero for a parameter, or omit trailing parameters from
	// ParameterOIDs entirely (the Postgres protocol allows specifying types for only a prefix
	// of the placeholders), to indicate that a parameter's type should be inferred. We always
	// compute the plan-inferred types (we can't know whether bindVarTypes is missing
	// trailing entries without first inspecting the analyzed plan) but only use an inferred
	// type to fill a position the client left unspecified. An explicit, non-zero OID from the
	// client must never be overwritten by an inferred type, since the client will encode that
	// parameter's Bind value using its own declared type.
	inferredTypes, err := extractBindVarTypes(ctx, analyzedPlan)
	if err != nil {
		return err
	}
	merged := make([]uint32, len(inferredTypes))
	copy(merged, inferredTypes)
	for i, oid := range bindVarTypes {
		if oid != 0 && i < len(merged) {
			merged[i] = oid
		}
	}
	bindVarTypes = merged

	h.preparedStatements[message.Name] = PreparedStatementData{
		Query:        query,
		ReturnFields: fields,
		BindVarTypes: bindVarTypes,
	}
	return h.send(&pgproto3.ParseComplete{})
}

// handleDescribe handles a Describe message, returning any error that occurs
func (h *ConnectionHandler) handleDescribe(message *pgproto3.Describe) error {
	var fields []pgproto3.FieldDescription
	var bindvarTypes []uint32
	var query ConvertedQuery

	h.waitForSync = true
	if message.ObjectType == 'S' {
		preparedStatementData, ok := h.preparedStatements[message.Name]
		if !ok {
			return errors.Errorf("prepared statement %s does not exist", message.Name)
		}

		fields = preparedStatementData.ReturnFields
		bindvarTypes = preparedStatementData.BindVarTypes
		query = preparedStatementData.Query
	} else {
		portalData, ok := h.portals[message.Name]
		if !ok {
			return errors.Errorf("portal %s does not exist", message.Name)
		}

		fields = portalData.Fields
		query = portalData.Query
	}

	return h.sendDescribeResponse(fields, bindvarTypes, query)
}

// handleBind handles a bind message, returning any error that occurs
func (h *ConnectionHandler) handleBind(message *pgproto3.Bind) error {
	h.waitForSync = true

	// TODO: a named portal object lasts till the end of the current transaction, unless explicitly destroyed
	//  we need to destroy the named portal as a side effect of the transaction ending
	logrus.Tracef("binding portal %q to prepared statement %s", message.DestinationPortal, message.PreparedStatement)
	preparedData, ok := h.preparedStatements[message.PreparedStatement]
	if !ok {
		return errors.Errorf("prepared statement %s does not exist", message.PreparedStatement)
	}

	if err := h.rejectStatementIfTransactionFailed(preparedData.Query); err != nil {
		return err
	}

	if preparedData.Query.AST == nil {
		// special case: empty query
		h.portals[message.DestinationPortal] = PortalData{
			Query:        preparedData.Query,
			IsEmptyQuery: true,
		}
		return h.send(&pgproto3.BindComplete{})
	}

	analyzedPlan, fields, err := h.doltgresHandler.ComBind(
		context.Background(),
		h.mysqlConn,
		preparedData.Query.String,
		preparedData.Query.AST,
		BindVariables{
			varTypes:    preparedData.BindVarTypes,
			formatCodes: message.ParameterFormatCodes,
			parameters:  message.Parameters,
		},
		message.ResultFormatCodes)
	if err != nil {
		return err
	}

	boundPlan, ok := analyzedPlan.(sql.Node)
	if !ok {
		return errors.Errorf("expected a sql.Node, got %T", analyzedPlan)
	}

	resultFormatCodes, err := extendFormatCodes(len(fields), message.ResultFormatCodes)
	if err != nil {
		return err
	}
	h.portals[message.DestinationPortal] = PortalData{
		Query:       preparedData.Query,
		Fields:      fields,
		BoundPlan:   boundPlan,
		FormatCodes: resultFormatCodes,
	}
	return h.send(&pgproto3.BindComplete{})
}

// handleExecute handles an execute message, returning any error that occurs
func (h *ConnectionHandler) handleExecute(message *pgproto3.Execute) error {
	h.waitForSync = true

	// TODO: implement the RowMax
	portalData, ok := h.portals[message.Portal]
	if !ok {
		return errors.Errorf("portal %s does not exist", message.Portal)
	}

	logrus.Tracef("executing portal %s with contents %v", message.Portal, portalData)
	query := portalData.Query

	if portalData.IsEmptyQuery {
		return h.send(&pgproto3.EmptyQueryResponse{})
	}

	if err := h.rejectStatementIfTransactionFailed(query); err != nil {
		return err
	}

	// Statements executed via the extended query protocol run in an implicit transaction block, which is
	// closed (committed on success, rolled back on error) by the next Sync message
	if err := h.startImplicitTransaction(query); err != nil {
		return err
	}

	// Certain statement types get handled directly by the handler instead of being passed to the engine
	handled, _, err := h.handleQueryOutsideEngine(query)
	if handled {
		return err
	}

	// |rowsAffected| gets altered by the callback below
	rowsAffected := int32(0)

	callback := h.spoolRowsCallback(query, &rowsAffected, true)
	err = h.doltgresHandler.ComExecuteBound(context.Background(), h.mysqlConn, query.String, portalData.BoundPlan, portalData.FormatCodes, callback)
	if err != nil {
		return err
	}

	return h.send(makeCommandComplete(query.StatementTag, rowsAffected))
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

// handleCopyData handles the COPY DATA message, by loading the data sent from the client. The |stop| response parameter
// is true if the connection handler should shut down the connection, |endOfMessages| is true if no more COPY DATA
// messages are expected, and the server should tell the client that it is ready for the next query, and |err| contains
// any error that occurred while processing the COPY DATA message.
func (h *ConnectionHandler) handleCopyData(message *pgproto3.CopyData) (stop bool, endOfMessages bool, err error) {
	copyFromData := bytes.NewReader(message.Data)
	stop, endOfMessages, err = h.handleCopyDataHelper(h.copyFromStdinState, copyFromData)
	if err != nil && h.copyFromStdinState != nil {
		h.copyFromStdinState.copyErr = err
	}
	return stop, endOfMessages, err
}

// copyFromFileQuery handles a COPY FROM message that is reading from a file, returning any error that occurs
func (h *ConnectionHandler) copyFromFileQuery(stmt *node.CopyFrom) error {
	copyState := &copyFromStdinState{
		copyFromStdinNode: stmt,
	}

	// TODO: security check for file path
	// TODO: Privilege Checking: https://www.postgresql.org/docs/15/sql-copy.html
	f, err := os.Open(stmt.File)
	if err != nil {
		return err
	}
	defer f.Close()

	_, _, err = h.handleCopyDataHelper(copyState, f)
	if err != nil {
		return err
	}

	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return err
	}

	loadDataResults, err := copyState.dataLoader.Finish(sqlCtx)
	if err != nil {
		return err
	}

	if sqlCtx.GetTransaction() != nil && sqlCtx.GetIgnoreAutoCommit() {
		txSession, ok := sqlCtx.Session.(sql.TransactionSession)
		if !ok {
			return errors.Errorf("session does not implement sql.TransactionSession")
		}
		if err = txSession.CommitTransaction(sqlCtx, txSession.GetTransaction()); err != nil {
			return err
		}
		sqlCtx.SetIgnoreAutoCommit(false)
	}

	return h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(fmt.Sprintf("COPY %d", loadDataResults.RowsLoaded)),
	})
}

// handleCopyDataHelper is a helper function that should only be invoked by handleCopyData. handleCopyData wraps this
// function so that it can capture any returned error message and store it in the saved state.
func (h *ConnectionHandler) handleCopyDataHelper(copyState *copyFromStdinState, copyFromData io.Reader) (stop bool, endOfMessages bool, err error) {
	if copyState == nil {
		return false, true, errors.Errorf("COPY DATA message received without a COPY FROM STDIN operation in progress")
	}

	// Grab a sql.Context and ensure the session has a transaction started, otherwise the copied data
	// won't get committed correctly.
	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "COPY FROM STDIN")
	if err != nil {
		return false, false, err
	}
	if err = startTransactionIfNecessary(sqlCtx); err != nil {
		return false, false, err
	}

	if copyState.copyFromStdinNode.TableName.Schema != "" {
		originalSchema, err := sqlCtx.GetSessionVariable(sqlCtx, "search_path")
		if err != nil {
			return false, false, err
		}
		err = sqlCtx.SetSessionVariable(sqlCtx, "search_path", copyState.copyFromStdinNode.TableName.Schema)
		if err != nil {
			return false, false, err
		}
		defer func() {
			_ = sqlCtx.SetSessionVariable(sqlCtx, "search_path", originalSchema)
		}()
	}

	dataLoader := copyState.dataLoader
	if dataLoader == nil {
		copyFromStdinNode := copyState.copyFromStdinNode
		if copyFromStdinNode == nil {
			return false, false, errors.Errorf("no COPY FROM STDIN node found")
		}

		// we build an insert node to use for the full insert plan, for which the copy from node will be the row source
		builder := planbuilder.New(sqlCtx, h.doltgresHandler.e.Analyzer.Catalog, nil)
		node, flags, err := builder.BindOnly(copyFromStdinNode.InsertStub, "", nil)
		if err != nil {
			return false, false, err
		}

		insertNode, ok := node.(*plan.InsertInto)
		if !ok {
			return false, false, errors.Errorf("expected plan.InsertInto, got %T", node)
		}

		// now that we have our insert node, we can build the data loader
		tbl := getInsertableTable(insertNode.Destination)
		if tbl == nil {
			// this should be impossible, enforced by analyzer above
			return false, false, errors.Errorf("no insertable table found in %v", insertNode.Destination)
		}

		switch copyFromStdinNode.CopyOptions.CopyFormat {
		case tree.CopyFormatText:
			dataLoader, err = dataloader.NewTabularDataLoader(insertNode.ColumnNames, tbl.Schema(sqlCtx), copyFromStdinNode.CopyOptions.Delimiter, "", copyFromStdinNode.CopyOptions.Header)
		case tree.CopyFormatCsv:
			dataLoader, err = dataloader.NewCsvDataLoader(insertNode.ColumnNames, tbl.Schema(sqlCtx), copyFromStdinNode.CopyOptions.Delimiter, copyFromStdinNode.CopyOptions.Header)
		case tree.CopyFormatBinary:
			err = errors.Errorf("BINARY format is not supported for COPY FROM")
		default:
			err = errors.Errorf("unknown format specified for COPY FROM: %v",
				copyFromStdinNode.CopyOptions.CopyFormat)
		}

		if err != nil {
			return false, false, err
		}

		// we have to set the data loader on the copyFrom node before we analyze it, because we need the loader's
		// schema to analyze
		copyState.copyFromStdinNode.DataLoader = dataLoader

		// After building out stub insert node, swap out the source node with the COPY node, then analyze the entire thing
		node = insertNode.WithSource(copyFromStdinNode)
		analyzedNode, err := h.doltgresHandler.e.Analyzer.Analyze(sqlCtx, node, nil, flags)
		if err != nil {
			return false, false, err
		}

		copyState.insertNode = analyzedNode
		copyState.dataLoader = dataLoader
	}

	reader := bufio.NewReader(copyFromData)
	if err = dataLoader.SetNextDataChunk(sqlCtx, reader); err != nil {
		return false, false, err
	}

	callback := func(_ *sql.Context, _ *Result) error { return nil }
	err = h.doltgresHandler.ComExecuteBound(sqlCtx, h.mysqlConn, "COPY FROM", copyState.insertNode, nil, callback)
	if err != nil {
		return false, false, err
	}

	// We expect to see more CopyData messages until we see either a CopyDone or CopyFail message, so
	// return false for endOfMessages
	return false, false, nil
}

// Returns the first sql.InsertableTable node found in the tree provided, or nil if none is found.
func getInsertableTable(node sql.Node) sql.InsertableTable {
	var tbl sql.InsertableTable
	transform.Inspect(node, func(node sql.Node) bool {
		if rt, ok := node.(*plan.ResolvedTable); ok {
			if insertable, ok := rt.Table.(sql.InsertableTable); ok {
				tbl = insertable
				return false
			}
		}
		return true
	})

	return tbl
}

// handleCopyDone handles a COPY DONE message by finalizing the in-progress COPY DATA operation and committing the
// loaded table data. The |stop| response parameter is true if the connection handler should shut down the connection,
// |endOfMessages| is true if no more COPY DATA messages are expected, and the server should tell the client that it is
// ready for the next query, and |err| contains any error that occurred while processing the COPY DATA message.
func (h *ConnectionHandler) handleCopyDone(_ *pgproto3.CopyDone) (stop bool, endOfMessages bool, err error) {
	if h.copyFromStdinState == nil {
		return false, true,
			errors.Errorf("COPY DONE message received without a COPY FROM STDIN operation in progress")
	}

	// If there was a previous error returned from processing a CopyData message, then don't return an error here
	// and don't send endOfMessage=true, since the CopyData error already sent endOfMessage=true. If we do send
	// endOfMessage=true here, then the client gets confused about the unexpected/extra Idle message since the
	// server has already reported it was idle in the last message after the returned error.
	if h.copyFromStdinState.copyErr != nil {
		return false, false, nil
	}

	dataLoader := h.copyFromStdinState.dataLoader
	if dataLoader == nil {
		return false, true,
			errors.Errorf("no data loader found for COPY FROM STDIN operation")
	}

	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return false, false, err
	}

	loadDataResults, err := dataLoader.Finish(sqlCtx)
	if err != nil {
		return false, false, err
	}

	// TODO: rather than always committing the transaction here, we should respect whether a transaction was
	//  expliclitly started and not commit if not. In order to do that, we need to not always set
	//  ctx.GetIgnoreAutoCommit(), and instead conditionally *not* insert a transaction closing iterator during chunk
	//  processing. We need a new query flag to effectively do the latter though.
	txSession, ok := sqlCtx.Session.(sql.TransactionSession)
	if !ok {
		return false, false, errors.Errorf("session does not implement sql.TransactionSession")
	}
	if err = txSession.CommitTransaction(sqlCtx, txSession.GetTransaction()); err != nil {
		return false, false, err
	}
	sqlCtx.SetIgnoreAutoCommit(false)

	h.copyFromStdinState = nil
	// We send back endOfMessage=true, since the COPY DONE message ends the COPY DATA flow and the server is ready
	// to accept the next query now.
	return false, true, h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(fmt.Sprintf("COPY %d", loadDataResults.RowsLoaded)),
	})
}

// handleCopyFail handles a COPY FAIL message by aborting the in-progress COPY DATA operation.  The |stop| response
// parameter is true if the connection handler should shut down the connection, |endOfMessages| is true if no more
// COPY DATA messages are expected, and the server should tell the client that it is ready for the next query, and
// |err| contains any error that occurred while processing the COPY DATA message.
func (h *ConnectionHandler) handleCopyFail(_ *pgproto3.CopyFail) (stop bool, endOfMessages bool, err error) {
	if h.copyFromStdinState == nil {
		return false, true,
			errors.Errorf("COPY FAIL message received without a COPY FROM STDIN operation in progress")
	}

	dataLoader := h.copyFromStdinState.dataLoader
	if dataLoader == nil {
		return false, true,
			errors.Errorf("no data loader found for COPY FROM STDIN operation")
	}

	h.copyFromStdinState = nil
	// We send back endOfMessage=true, since the COPY FAIL message ends the COPY DATA flow and the server is ready
	// to accept the next query now.
	return false, true, nil
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
	if h.restoredAutoCommitWithoutTransaction() {
		return
	}
	if err := h.runEngineTransactionControl("ROLLBACK"); err != nil {
		logrus.Warnf("error rolling back implicit transaction: %s", err)
	}
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

// deallocatePreparedStatement handles a DEALLOCATE statement by deleting the corresponding prepared statement from the
// handler's prepared statement map, and sending a CommandComplete message back to the client. Pass an empty |name|
// for `ALL`. This matches the behavior in the parser, which doesn't include a separate field for ALL.
func (h *ConnectionHandler) deallocatePreparedStatement(name string, preparedStatements map[string]PreparedStatementData, query ConvertedQuery, conn net.Conn) error {
	if name == "" {
		for name := range preparedStatements {
			delete(preparedStatements, name)
		}
	} else {
		_, ok := preparedStatements[name]
		if !ok {
			return errors.Errorf("prepared statement %s does not exist", name)
		}
		delete(preparedStatements, name)
	}

	return h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(query.StatementTag),
	})
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

// sendDescribeResponse sends a response message for a Describe message
func (h *ConnectionHandler) sendDescribeResponse(fields []pgproto3.FieldDescription, types []uint32, query ConvertedQuery) error {
	// The prepared statement variant of the describe command returns the OIDs of the parameters.
	if types != nil {
		if err := h.send(&pgproto3.ParameterDescription{
			ParameterOIDs: types,
		}); err != nil {
			return err
		}
	}

	if returnsRow(query) {
		// Both variants finish with a row description.
		return h.send(&pgproto3.RowDescription{
			Fields: fields,
		})
	} else {
		return h.send(&pgproto3.NoData{})
	}
}

// endOfMessages should be called from HandleConnection or a function within HandleConnection. This represents the end
// of the message slice, which may occur naturally (all relevant response messages have been sent) or on error. Once
// endOfMessages has been called, no further messages should be sent, and the connection loop should wait for the next
// query. A nil error should be provided if this is being called naturally.
func (h *ConnectionHandler) endOfMessages(err error) {
	if err != nil {
		switch h.transactionState {
		case implicitTransactionState:
			h.rollbackImplicitTransaction()
		case explicitTransactionState:
			h.transactionState = failedTransactionState
		}
		h.sendError(err)
	}
	ti := ReadyForQueryTransactionIndicator_Idle
	switch h.transactionState {
	case failedTransactionState:
		ti = ReadyForQueryTransactionIndicator_FailedTransactionBlock
	case explicitTransactionState, implicitTransactionState:
		ti = ReadyForQueryTransactionIndicator_TransactionBlock
	}
	if sendErr := h.send(&pgproto3.ReadyForQuery{
		TxStatus: byte(ti),
	}); sendErr != nil {
		// We panic here for the same reason as above.
		panic(sendErr)
	}
}

// sendError sends the given error to the client. This should generally never be called directly.
func (h *ConnectionHandler) sendError(err error) {
	pgErr := castSQLError(err)
	if sendErr := h.send(&pgproto3.ErrorResponse{
		Severity: pgErr.Severity,
		Code:     pgErr.Code,
		Message:  pgErr.Message,
	}); sendErr != nil {
		// If we're unable to send anything to the connection, then there's something wrong with the connection and
		// we should terminate it. This will be caught in HandleConnection's defer block.
		panic(sendErr)
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
		vitessAST, err := ast.Convert(s[i])
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

// handleCopyFromStdinQuery handles the COPY FROM STDIN query at the Doltgres layer, without passing it to the engine.
// COPY FROM STDIN can't be handled directly by the GMS engine, since COPY FROM STDIN relies on multiple messages sent
// over the wire.
func (h *ConnectionHandler) handleCopyFromStdinQuery(copyFrom *node.CopyFrom, conn net.Conn) error {
	h.copyFromStdinState = &copyFromStdinState{
		copyFromStdinNode: copyFrom,
	}

	return h.send(&pgproto3.CopyInResponse{
		OverallFormat: 0,
	})
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

// castSQLError returns a *pgconn.PgError with the error SQL state code, populated for the specified error object.
// Many tools (e.g. ORMs, SQL workbenches) rely on this error metadata to work correctly. If the specified error is nil,
// nil will be returned. If the error is already of type *pgconn.PgError, the error will be returned as is.
func castSQLError(err error) *pgconn.PgError {
	if err == nil {
		return nil
	}
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr
	}

	if w, ok := err.(sql.WrappedInsertError); ok {
		return castSQLError(w.Cause)
	}

	if wm, ok := err.(sql.WrappedTypeConversionError); ok {
		return castSQLError(wm.Err)
	}

	// Errors originating in our Postgres-derived parser already carry a candidate SQLSTATE (e.g. 42601
	// for syntax errors, 0A000 for unimplemented syntax); report that code directly when present.
	if pgerror.HasCandidateCode(err) {
		if code := pgerror.GetPGCode(err); code != pgcode.Uncategorized {
			return &pgconn.PgError{
				Severity: string(ErrorResponseSeverity_Error),
				Code:     code.String(),
				Message:  err.Error(),
			}
		}
	}

	// Errors that reach a client with an XX-class (internal error) code are more than cosmetic: some
	// clients (e.g. Npgsql) treat XX-class errors as critical failures and close the connection, so any
	// error a client can provoke should be mapped to its proper SQLSTATE here.
	// TODO: should update the error message to match Postgres
	var code pgcode.Code
	switch {
	// Class 0A — Feature Not Supported
	case sql.ErrUnsupportedFeature.Is(err), sql.ErrUnsupportedSyntax.Is(err):
		code = pgcode.FeatureNotSupported
	// Class 21 — Cardinality Violation
	case sql.ErrExpectedSingleRow.Is(err), sql.ErrMoreThanOneRow.Is(err):
		code = pgcode.CardinalityViolation
	// Class 22 — Data Exception
	case pgtypes.ErrDivisionByZero.Is(err):
		code = pgcode.DivisionByZero
	case sql.ErrValueOutOfRange.Is(err), pgtypes.ErrValueIsOutOfRangeForType.Is(err),
		pgtypes.ErrOutOfRange.Is(err), pgtypes.ErrInputOutOfRange.Is(err),
		errors.Is(err, pgtypes.ErrCastOutOfRange):
		code = pgcode.NumericValueOutOfRange
	case pgtypes.ErrInvalidSyntaxForType.Is(err), sql.ErrInvalidValue.Is(err):
		code = pgcode.InvalidTextRepresentation
	case pgtypes.ErrWrongLengthBit.Is(err), pgtypes.ErrVarBitLengthExceeded.Is(err):
		code = pgcode.StringDataLengthMismatch
	case sql.ErrInvalidTimeZone.Is(err), sql.ErrInvalidArgument.Is(err), sql.ErrInvalidArgumentDetails.Is(err):
		code = pgcode.InvalidParameterValue
	// Class 23 — Integrity Constraint Violation
	case sql.ErrPrimaryKeyViolation.Is(err), sql.ErrUniqueKeyViolation.Is(err),
		sql.ErrDuplicateEntry.Is(err), sql.ErrDuplicateEntrySet.Is(err):
		code = pgcode.UniqueViolation
	case sql.ErrForeignKeyChildViolation.Is(err), sql.ErrForeignKeyParentViolation.Is(err),
		sql.ErrForeignKeyNotResolved.Is(err):
		code = pgcode.ForeignKeyViolation
	case sql.ErrCheckConstraintViolated.Is(err), pgtypes.ErrDomainValueViolatesCheckConstraint.Is(err):
		code = pgcode.CheckViolation
	case sql.ErrInsertIntoNonNullableProvidedNull.Is(err),
		sql.ErrColumnDefaultReturnedNull.Is(err), pgtypes.ErrDomainDoesNotAllowNullValues.Is(err):
		code = pgcode.NotNullViolation
	// Class 25 — Invalid Transaction State
	case sql.ErrReadOnly.Is(err), sql.ErrReadOnlyTransaction.Is(err):
		code = pgcode.ReadOnlySQLTransaction
	// Classes 26, 34, 3B — statement, cursor, and savepoint names
	case sql.ErrUnknownPreparedStatement.Is(err):
		code = pgcode.InvalidSQLStatementName
	case sql.ErrCursorNotFound.Is(err):
		code = pgcode.InvalidCursorName
	case sql.ErrCursorAlreadyOpen.Is(err):
		code = pgcode.DuplicateCursor
	case sql.ErrSavepointDoesNotExist.Is(err):
		code = pgcode.InvalidSavepointSpecification
	// Classes 3D, 3F — catalog and schema names
	case sql.ErrDatabaseExists.Is(err):
		code = pgcode.DuplicateDatabase
	case sql.ErrDatabaseNotFound.Is(err):
		code = pgcode.UndefinedDatabase
	case sql.ErrDatabaseSchemaExists.Is(err):
		code = pgcode.DuplicateSchema
	case sql.ErrDatabaseSchemaNotFound.Is(err):
		code = pgcode.UndefinedSchema
	// Class 40 — Transaction Rollback. Dolt reports commit-time transaction conflicts as ErrLockDeadlock.
	case sql.ErrLockDeadlock.Is(err):
		code = pgcode.SerializationFailure
	// Class 42 — Syntax or Access Rule Violation
	case sql.ErrSyntaxError.Is(err), sql.ErrInvalidSyntax.Is(err), sql.ErrColValCountMismatch.Is(err),
		sql.ErrInsertIntoMismatchValueCount.Is(err), sql.ErrColumnNumberDoesNotMatch.Is(err):
		code = pgcode.Syntax
	case sql.ErrPrivilegeCheckFailed.Is(err), sql.ErrDatabaseAccessDeniedForUser.Is(err),
		sql.ErrTableAccessDeniedForUser.Is(err):
		code = pgcode.InsufficientPrivilege
	case sql.ErrNonAggregatedColumnWithoutGroupBy.Is(err):
		code = pgcode.Grouping
	case sql.ErrTableNotFound.Is(err), sql.ErrUnknownTable.Is(err), sql.ErrViewDoesNotExist.Is(err):
		code = pgcode.UndefinedTable
	case sql.ErrTableAlreadyExists.Is(err), sql.ErrExistingView.Is(err):
		code = pgcode.DuplicateRelation
	case sql.ErrColumnNotFound.Is(err), sql.ErrTableColumnNotFound.Is(err), sql.ErrUnknownColumn.Is(err),
		sql.ErrKeyColumnDoesNotExist.Is(err):
		code = pgcode.UndefinedColumn
	case sql.ErrColumnExists.Is(err), sql.ErrDuplicateColumn.Is(err), sql.ErrColumnSpecifiedTwice.Is(err):
		code = pgcode.DuplicateColumn
	case sql.ErrAmbiguousColumnName.Is(err), sql.ErrAmbiguousColumnOrAliasName.Is(err),
		sql.ErrAmbiguousColumnInOrderBy.Is(err):
		code = pgcode.AmbiguousColumn
	case sql.ErrFunctionNotFound.Is(err), sql.ErrTableFunctionNotFound.Is(err),
		sql.ErrInvalidArgumentNumber.Is(err), framework.ErrFunctionDoesNotExist.Is(err):
		code = pgcode.UndefinedFunction
	case sql.ErrForeignKeyDuplicateName.Is(err), pgtypes.ErrTypeAlreadyExists.Is(err):
		code = pgcode.DuplicateObject
	case sql.ErrUnknownSystemVariable.Is(err), sql.ErrUnknownConstraint.Is(err), pgtypes.ErrTypeDoesNotExist.Is(err):
		code = pgcode.UndefinedObject
	default:
		code = pgcode.Internal
	}

	return &pgconn.PgError{
		Severity: string(ErrorResponseSeverity_Error),
		Code:     code.String(),
		Message:  err.Error(),
	}
}
