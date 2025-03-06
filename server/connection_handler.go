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
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	psql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/ast"
	"github.com/dolthub/doltgresql/server/node"
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

	// Exposing backend as part of the context allows other code to access the connection and send pgproto messages.
	// This is required for RAISE in PL/pgSQL, for example.
	backend := pgproto3.NewBackend(conn, conn)
	doltgresHandler := &DoltgresHandler{
		e:                 server.Engine,
		sm:                server.SessionManager(),
		readTimeout:       0,     // cfg.ConnReadTimeout,
		encodeLoggedQuery: false, // cfg.EncodeLoggedQuery,
		pgTypeMap:         pgtype.NewMap(),
		backend:           backend,
	}
	if sel != nil {
		doltgresHandler.sel = sel
	}

	return &ConnectionHandler{
		mysqlConn:          mysqlConn,
		preparedStatements: preparedStatements,
		portals:            portals,
		doltgresHandler:    doltgresHandler,
		backend:            backend,
	}
}

// HandleConnection handles a connection's session, reading messages, executing queries, and sending responses.
// Expected to run in a goroutine per connection.
func (h *ConnectionHandler) HandleConnection() {
	var returnErr error
	if HandlePanics {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Listener recovered panic: %v", r)

				var eomErr error
				if returnErr != nil {
					eomErr = returnErr
				} else if rErr, ok := r.(error); ok {
					eomErr = rErr
				} else {
					eomErr = errors.Errorf("panic: %v", r)
				}

				// Sending eom can panic, which means we must recover again
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Listener recovered panic: %v", r)
					}
				}()
				h.endOfMessages(eomErr)
			}

			if returnErr != nil {
				fmt.Println(returnErr.Error())
			}

			h.doltgresHandler.ConnectionClosed(h.mysqlConn)
			if err := h.Conn().Close(); err != nil {
				fmt.Printf("Failed to properly close connection:\n%v\n", err)
			}
		}()
	}
	h.doltgresHandler.NewConnection(h.mysqlConn)

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
		if err = h.chooseInitialDatabase(sm); err != nil {
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
		Value: "15.0",
	}); err != nil {
		return err
	}
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "client_encoding",
		Value: "UTF8",
	}); err != nil {
		return err
	}
	return h.send(&pgproto3.BackendKeyData{
		ProcessID: processID,
		SecretKey: 0, // TODO: this should represent an ID that can uniquely identify this connection, so that CancelRequest will work
	})
}

// chooseInitialDatabase attempts to choose the initial database for the connection,
// if one is specified in the startup message provided
func (h *ConnectionHandler) chooseInitialDatabase(startupMessage *pgproto3.StartupMessage) error {
	db, ok := startupMessage.Parameters["database"]
	dbSpecified := ok && len(db) > 0
	if !dbSpecified {
		db = h.mysqlConn.User
	}
	useStmt := fmt.Sprintf("SET database TO '%s';", db)
	parsed, err := sql.GlobalParser.ParseSimple(useStmt)
	if err != nil {
		return err
	}
	err = h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, useStmt, parsed, func(res *Result) error {
		return nil
	})
	// If a database isn't specified, then we attempt to connect to a database with the same name as the user,
	// ignoring any error
	if err != nil && dbSpecified {
		_ = h.send(&pgproto3.ErrorResponse{
			Severity: string(ErrorResponseSeverity_Fatal),
			Code:     "3D000",
			Message:  fmt.Sprintf(`"database "%s" does not exist"`, db),
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
				fmt.Printf("Listener recovered panic: %v", r)

				var eomErr error
				if rErr, ok := r.(error); ok {
					eomErr = rErr
				} else {
					eomErr = errors.Errorf("panic: %v", r)
				}

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
		return false, true, nil
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
	handled, err := h.handledPSQLCommands(message.String)
	if handled || err != nil {
		return true, err
	}

	query, err := h.convertQuery(message.String)
	if err != nil {
		return true, err
	}

	// A query message destroys the unnamed statement and the unnamed portal
	delete(h.preparedStatements, "")
	delete(h.portals, "")

	// Certain statement types get handled directly by the handler instead of being passed to the engine
	handled, endOfMessages, err = h.handleQueryOutsideEngine(query)
	if handled {
		return endOfMessages, err
	}

	return true, h.query(query)
}

// handleQueryOutsideEngine handles any queries that should be handled by the handler directly, rather than being
// passed to the engine. The response parameter |handled| is true if the query was handled, |endOfMessages| is true
// if no more messages are expected for this query and server should send the client a READY FOR QUERY message,
// and any error that occurred while handling the query.
func (h *ConnectionHandler) handleQueryOutsideEngine(query ConvertedQuery) (handled bool, endOfMessages bool, err error) {
	switch stmt := query.AST.(type) {
	case *sqlparser.Deallocate:
		// TODO: handle ALL keyword
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
	query, err := h.convertQuery(message.Query)
	if err != nil {
		return err
	}

	if query.AST == nil {
		// special case: empty query
		h.preparedStatements[message.Name] = PreparedStatementData{
			Query: query,
		}
		return nil
	}

	parsedQuery, fields, err := h.doltgresHandler.ComPrepareParsed(context.Background(), h.mysqlConn, query.String, query.AST)
	if err != nil {
		return err
	}

	analyzedPlan, ok := parsedQuery.(sql.Node)
	if !ok {
		return errors.Errorf("expected a sql.Node, got %T", parsedQuery)
	}

	// A valid Parse message must have ParameterObjectIDs if there are any binding variables.
	bindVarTypes := message.ParameterOIDs
	if len(bindVarTypes) == 0 {
		// NOTE: This is used for Prepared Statement Tests only.
		bindVarTypes, err = extractBindVarTypes(analyzedPlan)
		if err != nil {
			return err
		}
	}

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
	var tag string

	h.waitForSync = true
	if message.ObjectType == 'S' {
		preparedStatementData, ok := h.preparedStatements[message.Name]
		if !ok {
			return errors.Errorf("prepared statement %s does not exist", message.Name)
		}

		fields = preparedStatementData.ReturnFields
		bindvarTypes = preparedStatementData.BindVarTypes
		tag = preparedStatementData.Query.StatementTag
	} else {
		portalData, ok := h.portals[message.Name]
		if !ok {
			return errors.Errorf("portal %s does not exist", message.Name)
		}

		fields = portalData.Fields
		tag = portalData.Query.StatementTag
	}

	return h.sendDescribeResponse(fields, bindvarTypes, tag)
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
		})
	if err != nil {
		return err
	}

	boundPlan, ok := analyzedPlan.(sql.Node)
	if !ok {
		return errors.Errorf("expected a sql.Node, got %T", analyzedPlan)
	}

	h.portals[message.DestinationPortal] = PortalData{
		Query:     preparedData.Query,
		Fields:    fields,
		BoundPlan: boundPlan,
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

	// Certain statement types get handled directly by the handler instead of being passed to the engine
	handled, _, err := h.handleQueryOutsideEngine(query)
	if handled {
		return err
	}

	// |rowsAffected| gets altered by the callback below
	rowsAffected := int32(0)

	callback := h.spoolRowsCallback(query.StatementTag, &rowsAffected, true)
	err = h.doltgresHandler.ComExecuteBound(context.Background(), h.mysqlConn, query.String, portalData.BoundPlan, callback)
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

	dataLoader := copyState.dataLoader
	if dataLoader == nil {
		copyFromStdinNode := copyState.copyFromStdinNode
		if copyFromStdinNode == nil {
			return false, false, errors.Errorf("no COPY FROM STDIN node found")
		}

		// we build an insert node to use for the full insert plan, for which the copy from node will be the row source
		builder := planbuilder.New(sqlCtx, h.doltgresHandler.e.Analyzer.Catalog, nil, psql.NewPostgresParser())
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
			dataLoader, err = dataloader.NewTabularDataLoader(insertNode.ColumnNames, tbl.Schema(), copyFromStdinNode.CopyOptions.Delimiter, "", copyFromStdinNode.CopyOptions.Header)
		case tree.CopyFormatCsv:
			dataLoader, err = dataloader.NewCsvDataLoader(insertNode.ColumnNames, tbl.Schema(), copyFromStdinNode.CopyOptions.Delimiter, copyFromStdinNode.CopyOptions.Header)
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

	callback := func(res *Result) error { return nil }
	err = h.doltgresHandler.ComExecuteBound(sqlCtx, h.mysqlConn, "COPY FROM", copyState.insertNode, callback)
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

func (h *ConnectionHandler) deallocatePreparedStatement(name string, preparedStatements map[string]PreparedStatementData, query ConvertedQuery, conn net.Conn) error {
	_, ok := preparedStatements[name]
	if !ok {
		return errors.Errorf("prepared statement %s does not exist", name)
	}
	delete(preparedStatements, name)

	return h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(query.StatementTag),
	})
}

// query runs the given query and sends a CommandComplete message to the client
func (h *ConnectionHandler) query(query ConvertedQuery) error {
	// |rowsAffected| gets altered by the callback below
	rowsAffected := int32(0)

	callback := h.spoolRowsCallback(query.StatementTag, &rowsAffected, false)
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
func (h *ConnectionHandler) spoolRowsCallback(tag string, rows *int32, isExecute bool) func(res *Result) error {
	// IsIUD returns whether the query is either an INSERT, UPDATE, or DELETE query.
	isIUD := tag == "INSERT" || tag == "UPDATE" || tag == "DELETE"
	return func(res *Result) error {
		if returnsRow(tag) {
			// EXECUTE does not send RowDescription; instead it should be sent from DESCRIBE prior to it
			if !isExecute {
				if err := h.send(&pgproto3.RowDescription{
					Fields: res.Fields,
				}); err != nil {
					return err
				}
			}

			for _, row := range res.Rows {
				if err := h.send(&pgproto3.DataRow{
					Values: row.val,
				}); err != nil {
					return err
				}
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
func (h *ConnectionHandler) sendDescribeResponse(fields []pgproto3.FieldDescription, types []uint32, tag string) error {
	// The prepared statement variant of the describe command returns the OIDs of the parameters.
	if types != nil {
		if err := h.send(&pgproto3.ParameterDescription{
			ParameterOIDs: types,
		}); err != nil {
			return err
		}
	}

	if returnsRow(tag) {
		// Both variants finish with a row description.
		return h.send(&pgproto3.RowDescription{
			Fields: fields,
		})
	} else {
		return h.send(&pgproto3.NoData{})
	}
}

// handledPSQLCommands handles the special PSQL commands, such as \l and \dt.
func (h *ConnectionHandler) handledPSQLCommands(statement string) (bool, error) {
	statement = strings.ToLower(statement)
	// Command: \l
	if statement == "select d.datname as \"name\",\n       pg_catalog.pg_get_userbyid(d.datdba) as \"owner\",\n       pg_catalog.pg_encoding_to_char(d.encoding) as \"encoding\",\n       d.datcollate as \"collate\",\n       d.datctype as \"ctype\",\n       d.daticulocale as \"icu locale\",\n       case d.datlocprovider when 'c' then 'libc' when 'i' then 'icu' end as \"locale provider\",\n       pg_catalog.array_to_string(d.datacl, e'\\n') as \"access privileges\"\nfrom pg_catalog.pg_database d\norder by 1;" {
		query, err := h.convertQuery(`select d.datname as "Name", 'postgres' as "Owner", 'UTF8' as "Encoding", 'en_US.UTF-8' as "Collate", 'en_US.UTF-8' as "Ctype", 'en-US' as "ICU Locale", case d.datlocprovider when 'c' then 'libc' when 'i' then 'icu' end as "locale provider", '' as "access privileges" from pg_catalog.pg_database d order by 1;`)
		if err != nil {
			return false, err
		}
		return true, h.query(query)
	}
	// Command: \l on psql 16
	if statement == "select\n  d.datname as \"name\",\n  pg_catalog.pg_get_userbyid(d.datdba) as \"owner\",\n  pg_catalog.pg_encoding_to_char(d.encoding) as \"encoding\",\n  case d.datlocprovider when 'c' then 'libc' when 'i' then 'icu' end as \"locale provider\",\n  d.datcollate as \"collate\",\n  d.datctype as \"ctype\",\n  d.daticulocale as \"icu locale\",\n  null as \"icu rules\",\n  pg_catalog.array_to_string(d.datacl, e'\\n') as \"access privileges\"\nfrom pg_catalog.pg_database d\norder by 1;" {
		query, err := h.convertQuery(`select d.datname as "Name", 'postgres' as "Owner", 'UTF8' as "Encoding", 'en_US.UTF-8' as "Collate", 'en_US.UTF-8' as "Ctype", 'en-US' as "ICU Locale", case d.datlocprovider when 'c' then 'libc' when 'i' then 'icu' end as "locale provider", '' as "access privileges" from pg_catalog.pg_database d order by 1;`)
		if err != nil {
			return false, err
		}
		return true, h.query(query)
	}
	// Command: \dt
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT table_schema AS "Schema", TABLE_NAME AS "Name", 'table' AS "Type", 'postgres' AS "Owner" FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA <> 'pg_catalog' AND TABLE_SCHEMA <> 'information_schema' AND TABLE_TYPE = 'BASE TABLE' ORDER BY 2;`,
			StatementTag: "SELECT",
		})
	}
	// Command: \d
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','v','m','s','f','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT table_schema AS "Schema", TABLE_NAME AS "Name", IF(TABLE_TYPE = 'VIEW', 'view', 'table') AS "Type", 'postgres' AS "Owner" FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA <> 'pg_catalog' AND TABLE_SCHEMA <> 'information_schema' AND TABLE_TYPE = 'BASE TABLE' OR TABLE_TYPE = 'VIEW' ORDER BY 2;`,
			StatementTag: "SELECT",
		})
	}
	// Alternate \d for psql 14
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 's' then 'special' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','v','m','s','f','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT table_schema AS "Schema", TABLE_NAME AS "Name", IF(TABLE_TYPE = 'VIEW', 'view', 'table') AS "Type", 'postgres' AS "Owner" FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA <> 'pg_catalog' AND TABLE_SCHEMA <> 'information_schema' AND TABLE_TYPE = 'BASE TABLE' OR TABLE_TYPE = 'VIEW' ORDER BY 2;`,
			StatementTag: "SELECT",
		})
	}
	// Command: \d table_name
	if strings.HasPrefix(statement, "select c.oid,\n  n.nspname,\n  c.relname\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\nwhere c.relname operator(pg_catalog.~) '^(") && strings.HasSuffix(statement, ")$' collate pg_catalog.default\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 2, 3;") {
		// There are >at least< 15 separate statements sent for this command, which is far too much to validate and
		// implement, so we'll just return an error for now
		return true, errors.Errorf("PSQL command not yet supported")
	}
	// Command: \dn
	if statement == "select n.nspname as \"name\",\n  pg_catalog.pg_get_userbyid(n.nspowner) as \"owner\"\nfrom pg_catalog.pg_namespace n\nwhere n.nspname !~ '^pg_' and n.nspname <> 'information_schema'\norder by 1;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT 'public' AS "Name", 'pg_database_owner' AS "Owner";`,
			StatementTag: "SELECT",
		})
	}
	// Command: \df
	if statement == "select n.nspname as \"schema\",\n  p.proname as \"name\",\n  pg_catalog.pg_get_function_result(p.oid) as \"result data type\",\n  pg_catalog.pg_get_function_arguments(p.oid) as \"argument data types\",\n case p.prokind\n  when 'a' then 'agg'\n  when 'w' then 'window'\n  when 'p' then 'proc'\n  else 'func'\n end as \"type\"\nfrom pg_catalog.pg_proc p\n     left join pg_catalog.pg_namespace n on n.oid = p.pronamespace\nwhere pg_catalog.pg_function_is_visible(p.oid)\n      and n.nspname <> 'pg_catalog'\n      and n.nspname <> 'information_schema'\norder by 1, 2, 4;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT '' AS "Schema", '' AS "Name", '' AS "Result data type", '' AS "Argument data types", '' AS "Type" LIMIT 0;`,
			StatementTag: "SELECT",
		})
	}
	// Command: \dv
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\nwhere c.relkind in ('v','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT table_schema AS "Schema", TABLE_NAME AS "Name", 'view' AS "Type", 'postgres' AS "Owner" FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA <> 'pg_catalog' AND TABLE_SCHEMA <> 'information_schema' AND TABLE_TYPE = 'VIEW' ORDER BY 2;`,
			StatementTag: "SELECT",
		})
	}
	// Command: \du
	if statement == "select r.rolname, r.rolsuper, r.rolinherit,\n  r.rolcreaterole, r.rolcreatedb, r.rolcanlogin,\n  r.rolconnlimit, r.rolvaliduntil,\n  array(select b.rolname\n        from pg_catalog.pg_auth_members m\n        join pg_catalog.pg_roles b on (m.roleid = b.oid)\n        where m.member = r.oid) as memberof\n, r.rolreplication\n, r.rolbypassrls\nfrom pg_catalog.pg_roles r\nwhere r.rolname !~ '^pg_'\norder by 1;" {
		// We don't support users yet, so we'll just return nothing for now
		return true, h.query(ConvertedQuery{
			String:       `SELECT '' FROM dual LIMIT 0;`,
			StatementTag: "SELECT",
		})
	}
	return false, nil
}

// endOfMessages should be called from HandleConnection or a function within HandleConnection. This represents the end
// of the message slice, which may occur naturally (all relevant response messages have been sent) or on error. Once
// endOfMessages has been called, no further messages should be sent, and the connection loop should wait for the next
// query. A nil error should be provided if this is being called naturally.
func (h *ConnectionHandler) endOfMessages(err error) {
	if err != nil {
		h.sendError(err)
	}
	if sendErr := h.send(&pgproto3.ReadyForQuery{
		TxStatus: byte(ReadyForQueryTransactionIndicator_Idle),
	}); sendErr != nil {
		// We panic here for the same reason as above.
		panic(sendErr)
	}
}

// sendError sends the given error to the client. This should generally never be called directly.
func (h *ConnectionHandler) sendError(err error) {
	fmt.Println(err.Error())
	if sendErr := h.send(&pgproto3.ErrorResponse{
		Severity: string(ErrorResponseSeverity_Error),
		Code:     "XX000", // internal_error for now
		Message:  err.Error(),
	}); sendErr != nil {
		// If we're unable to send anything to the connection, then there's something wrong with the connection and
		// we should terminate it. This will be caught in HandleConnection's defer block.
		panic(sendErr)
	}
}

// convertQuery takes the given Postgres query, and converts it as an ast.ConvertedQuery that will work with the handler.
func (h *ConnectionHandler) convertQuery(query string) (ConvertedQuery, error) {
	s, err := parser.Parse(query)
	if err != nil {
		return ConvertedQuery{}, err
	}
	if len(s) > 1 {
		return ConvertedQuery{}, errors.Errorf("only a single statement at a time is currently supported")
	}
	if len(s) == 0 {
		return ConvertedQuery{String: query}, nil
	}
	vitessAST, err := ast.Convert(s[0])
	stmtTag := s[0].AST.StatementTag()
	if err != nil {
		return ConvertedQuery{}, err
	}
	if vitessAST == nil {
		return ConvertedQuery{
			String:       s[0].AST.String(),
			StatementTag: stmtTag,
		}, nil
	}
	return ConvertedQuery{
		String:       query,
		AST:          vitessAST,
		StatementTag: stmtTag,
	}, nil
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
func returnsRow(tag string) bool {
	switch tag {
	case "SELECT", "SHOW", "FETCH", "EXPLAIN", "SHOW TABLES":
		return true
	default:
		return false
	}
}
