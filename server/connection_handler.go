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
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync/atomic"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/sqltypes"
	querypb "github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/postgres/connection"
	"github.com/dolthub/doltgresql/postgres/messages"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ConnectionHandler is responsible for the entire lifecycle of a user connection: receiving messages they send,
// executing queries, sending the correct messages in return, and terminating the connection when appropriate.
type ConnectionHandler struct {
	mysqlConn          *mysql.Conn
	preparedStatements map[string]PreparedStatementData
	portals            map[string]PortalData
	handler            mysql.Handler
	pgTypeMap          *pgtype.Map
	waitForSync        bool
}

// Set this env var to disable panic handling in the connection, which is useful when debugging a panic
const disablePanicHandlingEnvVar = "DOLT_PGSQL_PANIC"

// handlePanics determines whether panics should be handled in the connection handler. See |disablePanicHandlingEnvVar|.
var handlePanics = true

func init() {
	if _, ok := os.LookupEnv(disablePanicHandlingEnvVar); ok {
		handlePanics = false
	}
}

// NewConnectionHandler returns a new ConnectionHandler for the connection provided
func NewConnectionHandler(conn net.Conn, handler mysql.Handler) *ConnectionHandler {
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

	return &ConnectionHandler{
		mysqlConn:          mysqlConn,
		preparedStatements: preparedStatements,
		portals:            portals,
		handler:            handler,
		pgTypeMap:          pgtype.NewMap(),
	}
}

// HandleConnection handles a connection's session, reading messages, executing queries, and sending responses.
// Expected to run in a goroutine per connection.
func (h *ConnectionHandler) HandleConnection() {
	var returnErr error
	if handlePanics {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Listener recovered panic: %v", r)

				var eomErr error
				if returnErr != nil {
					eomErr = returnErr
				} else if rErr, ok := r.(error); ok {
					eomErr = rErr
				} else {
					eomErr = fmt.Errorf("panic: %v", r)
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

			h.handler.ConnectionClosed(h.mysqlConn)
			if err := h.Conn().Close(); err != nil {
				fmt.Printf("Failed to properly close connection:\n%v\n", err)
			}
		}()
	}
	h.handler.NewConnection(h.mysqlConn)

	startupMessage, err := h.receiveStartupMessage()
	if err != nil {
		returnErr = err
		return
	}

	err = h.sendClientStartupMessages(startupMessage)
	if err != nil {
		returnErr = err
		return
	}

	err = h.chooseInitialDatabase(startupMessage)
	if err != nil {
		returnErr = err
		return
	}

	if err := connection.Send(h.Conn(), messages.ReadyForQuery{
		Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
	}); err != nil {
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
// received from the connection. Otherwise (a message is received successfully), the message is processed and any
// error is handled appropriately. The return value indicates whether the connection should be closed.
func (h *ConnectionHandler) receiveMessage() (bool, error) {
	var endOfMessages bool
	// For the time being, we handle panics in this function and treat them the same as errors so that they don't
	// forcibly close the connection. Contrast this with the panic handling logic in HandleConnection, where we treat any
	// panic as unrecoverable to the connection. As we fill out the implementation, we can revisit this decision and
	// rethink our posture over whether panics should terminate a connection.
	if handlePanics {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Listener recovered panic: %v", r)

				var eomErr error
				if rErr, ok := r.(error); ok {
					eomErr = rErr
				} else {
					eomErr = fmt.Errorf("panic: %v", r)
				}

				if !endOfMessages && h.waitForSync {
					if syncErr := connection.DiscardToSync(h.Conn()); syncErr != nil {
						fmt.Println(syncErr.Error())
					}
				}
				h.endOfMessages(eomErr)
			}
		}()
	}

	message, err := connection.Receive(h.Conn())
	if err != nil {
		return false, err
	}

	if ds, ok := message.(sql.DebugStringer); ok && logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("Received message: %s", ds.DebugString())
	} else {
		logrus.Debugf("Received message: %s", message.DefaultMessage().Name)
	}

	var stop bool
	stop, endOfMessages, err = h.handleMessage(message)
	if err != nil {
		if !endOfMessages && h.waitForSync {
			if syncErr := connection.DiscardToSync(h.Conn()); syncErr != nil {
				fmt.Println(syncErr.Error())
			}
		}
		h.endOfMessages(err)
	} else if endOfMessages {
		h.endOfMessages(nil)
	}

	return stop, nil
}

// receiveStarupMessage reads a startup message from the connection given and returns it. Some startup messages will
// result in the establishment of a new connection, which is also returned.
func (h *ConnectionHandler) receiveStartupMessage() (messages.StartupMessage, error) {
	var startupMessage messages.StartupMessage
	// The initial message may be one of a few different messages, so we'll check for those.
InitialMessageLoop:
	for {
		initialMessages, err := connection.ReceiveIntoAny(h.Conn(),
			messages.StartupMessage{},
			messages.SSLRequest{},
			messages.GSSENCRequest{})
		if err != nil {
			if err == io.EOF {
				return messages.StartupMessage{}, nil
			}
			return messages.StartupMessage{}, err
		}

		if len(initialMessages) != 1 {
			return messages.StartupMessage{}, fmt.Errorf("expected a single message upon starting connection, terminating connection")
		}

		initialMessage := initialMessages[0]
		switch initialMessage := initialMessage.(type) {
		case messages.StartupMessage:
			startupMessage = initialMessage
			break InitialMessageLoop
		case messages.SSLRequest:
			hasCertificate := len(certificate.Certificate) > 0
			if err := connection.Send(h.Conn(), messages.SSLResponse{
				SupportsSSL: hasCertificate,
			}); err != nil {
				return messages.StartupMessage{}, err
			}
			// If we have a certificate and the client has asked for SSL support, then we switch here.
			// This involves swapping out our underlying net connection for a new one.
			// We can't start in SSL mode, as the client does not attempt the handshake until after our response.
			if hasCertificate {
				conn := tls.Server(h.Conn(), &tls.Config{
					Certificates: []tls.Certificate{certificate},
				})
				h.mysqlConn.Conn = conn
			}
		case messages.GSSENCRequest:
			if err = connection.Send(h.Conn(), messages.GSSENCResponse{
				SupportsGSSAPI: false,
			}); err != nil {
				return messages.StartupMessage{}, err
			}
		default:
			return messages.StartupMessage{}, fmt.Errorf("unexpected initial message, terminating connection")
		}
	}

	return startupMessage, nil
}

// chooseInitialDatabase attempts to choose the initial database for the connection, if one is specified in the
// startup message provided
func (h *ConnectionHandler) chooseInitialDatabase(startupMessage messages.StartupMessage) error {
	if db, ok := startupMessage.Parameters["database"]; ok && len(db) > 0 {
		err := h.handler.ComQuery(context.Background(), h.mysqlConn, fmt.Sprintf("USE `%s`;", db), func(res *sqltypes.Result, more bool) error {
			return nil
		})
		if err != nil {
			_ = connection.Send(h.Conn(), messages.ErrorResponse{
				Severity:     messages.ErrorResponseSeverity_Fatal,
				SqlStateCode: "3D000",
				Message:      fmt.Sprintf(`"database "%s" does not exist"`, db),
				Optional: messages.ErrorResponseOptionalFields{
					Routine: "InitPostgres",
				},
			})
			return err
		}
	} else {
		// If a database isn't specified, then we attempt to connect to a database with the same name as the user,
		// ignoring any error
		_ = h.handler.ComQuery(context.Background(), h.mysqlConn, fmt.Sprintf("USE `%s`;", h.mysqlConn.User), func(*sqltypes.Result, bool) error {
			return nil
		})
	}
	return nil
}

// handleMessages processes the message provided and returns status flags for what should happen next
func (h *ConnectionHandler) handleMessage(message connection.Message) (stop, endOfMessages bool, err error) {
	switch message := message.(type) {
	case messages.Terminate:
		return true, false, nil
	case messages.Sync:
		h.waitForSync = false
		return false, true, nil
	case messages.Query:
		return false, true, h.handleQuery(message)
	case messages.Parse:
		return false, false, h.handleParse(message)
	case messages.Describe:
		return false, false, h.handleDescribe(message)
	case messages.Bind:
		return false, false, h.handleBind(message)
	case messages.Execute:
		return false, false, h.handleExecute(message)
	case messages.Close:
		if message.ClosingPreparedStatement {
			delete(h.preparedStatements, message.Target)
		} else {
			delete(h.portals, message.Target)
		}
		return false, false, connection.Send(h.Conn(), messages.CloseComplete{})
	default:
		return false, true, fmt.Errorf(`Unhandled message "%s"`, message.DefaultMessage().Name)
	}
}

// handleQuery handles a query message, returning any error that occurs
func (h *ConnectionHandler) handleQuery(message messages.Query) error {
	handled, err := h.handledPSQLCommands(message.String)
	if handled || err != nil {
		return err
	}

	// TODO: Remove this once we support `SELECT * FROM function()` syntax
	// Github issue: https://github.com/dolthub/doltgresql/issues/464
	handled, err = h.handledWorkbenchCommands(message.String)
	if handled || err != nil {
		return err
	}

	query, err := h.convertQuery(message.String)
	if err != nil {
		return err
	}

	// A query message destroys the unnamed statement and the unnamed portal
	delete(h.preparedStatements, "")
	delete(h.portals, "")

	// Certain statement types get handled directly by the handler instead of being passed to the engine
	err, handled = h.handleQueryOutsideEngine(query)
	if handled {
		return err
	}

	return h.query(query)
}

// handleQueryOutsideEngine handles any queries that should be handled by the handler directly, rather than being
// passed to the engine. Returns true if the query was handled and any error that occurred while doing so.
func (h *ConnectionHandler) handleQueryOutsideEngine(query ConvertedQuery) (error, bool) {
	switch stmt := query.AST.(type) {
	case *sqlparser.Deallocate:
		// TODO: handle ALL keyword
		return h.deallocatePreparedStatement(stmt.Name, h.preparedStatements, query, h.Conn()), true
	case sqlparser.InjectedStatement:
		switch stmt.Statement.(type) {
		case node.DiscardStatement:
			return h.discardAll(query, h.Conn()), true
		}
	}
	return nil, false
}

// handleParse handles a parse message, returning any error that occurs
func (h *ConnectionHandler) handleParse(message messages.Parse) error {
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

	plan, fields, err := h.getPlanAndFields(query)
	if err != nil {
		return err
	}

	// TODO: bindvar types can be specified directly in the message, need tests of this
	bindVarTypes, err := extractBindVarTypes(plan)
	if err != nil {
		return err
	}

	// Nil fields means an OKResult, fill one in here
	if fields == nil {
		fields = []*querypb.Field{
			{
				Name: "Rows",
				Type: sqltypes.Int32,
			},
		}
	}

	h.preparedStatements[message.Name] = PreparedStatementData{
		Query:        query,
		ReturnFields: fields,
		BindVarTypes: bindVarTypes,
	}

	return connection.Send(h.Conn(), messages.ParseComplete{})
}

// handleDescribe handles a Describe message, returning any error that occurs
func (h *ConnectionHandler) handleDescribe(message messages.Describe) error {
	var fields []*querypb.Field
	var bindvarTypes []int32
	var tag string

	h.waitForSync = true
	if message.IsPrepared {
		preparedStatementData, ok := h.preparedStatements[message.Target]
		if !ok {
			return fmt.Errorf("prepared statement %s does not exist", message.Target)
		}

		fields = preparedStatementData.ReturnFields
		bindvarTypes = preparedStatementData.BindVarTypes
		tag = preparedStatementData.Query.StatementTag
	} else {
		portalData, ok := h.portals[message.Target]
		if !ok {
			return fmt.Errorf("portal %s does not exist", message.Target)
		}

		fields = portalData.Fields
		tag = portalData.Query.StatementTag
	}

	return h.sendDescribeResponse(h.Conn(), fields, bindvarTypes, tag)
}

// handleBind handles a bind message, returning any error that occurs
func (h *ConnectionHandler) handleBind(message messages.Bind) error {
	h.waitForSync = true

	// TODO: a named portal object lasts till the end of the current transaction, unless explicitly destroyed
	//  we need to destroy the named portal as a side effect of the transaction ending
	logrus.Tracef("binding portal %q to prepared statement %s", message.DestinationPortal, message.SourcePreparedStatement)
	preparedData, ok := h.preparedStatements[message.SourcePreparedStatement]
	if !ok {
		return fmt.Errorf("prepared statement %s does not exist", message.SourcePreparedStatement)
	}

	if preparedData.Query.AST == nil {
		// special case: empty query
		h.portals[message.DestinationPortal] = PortalData{
			Query:        preparedData.Query,
			IsEmptyQuery: true,
		}
		return connection.Send(h.Conn(), messages.BindComplete{})
	}

	bindVars, err := h.convertBindParameters(preparedData.BindVarTypes, message.ParameterFormatCodes, message.ParameterValues)
	if err != nil {
		return err
	}

	boundPlan, fields, err := h.bindParams(preparedData.Query.String, preparedData.Query.AST, bindVars)
	if err != nil {
		return err
	}

	h.portals[message.DestinationPortal] = PortalData{
		Query:     preparedData.Query,
		Fields:    fields,
		BoundPlan: boundPlan,
	}
	return connection.Send(h.Conn(), messages.BindComplete{})
}

// handleExecute handles an execute message, returning any error that occurs
func (h *ConnectionHandler) handleExecute(message messages.Execute) error {
	h.waitForSync = true

	// TODO: implement the RowMax
	portalData, ok := h.portals[message.Portal]
	if !ok {
		return fmt.Errorf("portal %s does not exist", message.Portal)
	}

	logrus.Tracef("executing portal %s with contents %v", message.Portal, portalData)
	query := portalData.Query

	// we need the CommandComplete message defined here because it's altered by the callback below
	complete := messages.CommandComplete{
		Query: query.String,
		Tag:   query.StatementTag,
	}

	if portalData.IsEmptyQuery {
		return connection.Send(h.Conn(), messages.EmptyQueryResponse{})
	}

	// Certain statement types get handled directly by the handler instead of being passed to the engine
	err, handled := h.handleQueryOutsideEngine(query)
	if handled {
		return err
	}

	err = h.handler.(mysql.ExtendedHandler).ComExecuteBound(context.Background(), h.mysqlConn, query.String, portalData.BoundPlan, spoolRowsCallback(h.Conn(), &complete, true))
	if err != nil {
		return err
	}

	return connection.Send(h.Conn(), complete)
}

func (h *ConnectionHandler) deallocatePreparedStatement(name string, preparedStatements map[string]PreparedStatementData, query ConvertedQuery, conn net.Conn) error {
	_, ok := preparedStatements[name]
	if !ok {
		return fmt.Errorf("prepared statement %s does not exist", name)
	}
	delete(preparedStatements, name)

	commandComplete := messages.CommandComplete{
		Query: query.String,
		Tag:   query.StatementTag,
	}

	return connection.Send(conn, commandComplete)
}

func extractBindVarTypes(queryPlan sql.Node) ([]int32, error) {
	inspectNode := queryPlan
	switch queryPlan := queryPlan.(type) {
	case *plan.InsertInto:
		inspectNode = queryPlan.Source
	}

	types := make([]int32, 0)
	var err error
	extractBindVars := func(expr sql.Expression) bool {
		if err != nil {
			return false
		}
		switch e := expr.(type) {
		case *expression.BindVar:
			var oid int32
			if doltgresType, ok := e.Type().(pgtypes.DoltgresType); ok {
				oid = int32(doltgresType.OID())
			} else {
				oid, err = messages.VitessTypeToObjectID(e.Type().Type())
				if err != nil {
					err = fmt.Errorf("could not determine OID for placeholder %s: %w", e.Name, err)
					return false
				}
			}
			types = append(types, oid)
		case *pgexprs.ExplicitCast:
			if bindVar, ok := e.Child().(*expression.BindVar); ok {
				var oid int32
				if doltgresType, ok := bindVar.Type().(pgtypes.DoltgresType); ok {
					oid = int32(doltgresType.OID())
				} else {
					oid, err = messages.VitessTypeToObjectID(e.Type().Type())
					if err != nil {
						err = fmt.Errorf("could not determine OID for placeholder %s: %w", bindVar.Name, err)
						return false
					}
				}
				types = append(types, oid)
				return false
			}
		// $1::text and similar get converted to a Convert expression wrapping the bindvar
		case *expression.Convert:
			if bindVar, ok := e.Child.(*expression.BindVar); ok {
				var oid int32
				oid, err = messages.VitessTypeToObjectID(e.Type().Type())
				if err != nil {
					err = fmt.Errorf("could not determine OID for placeholder %s: %w", bindVar.Name, err)
					return false
				}
				types = append(types, oid)
				return false
			}
		}

		return true
	}

	transform.InspectExpressions(inspectNode, extractBindVars)
	return types, err
}

// convertBindParameters handles the conversion from bind parameters to variable values.
func (h *ConnectionHandler) convertBindParameters(types []int32, formatCodes []int32, values []messages.BindParameterValue) (map[string]*querypb.BindVariable, error) {
	bindings := make(map[string]*querypb.BindVariable, len(values))
	for i := range values {
		bindingName := fmt.Sprintf("v%d", i+1)
		typ := convertType(types[i])
		var bindVarString string

		// TODO: need to check for byte length for given type length. E.g. int16, int32 and uint32 expects 4 bytes
		//  but currently, receives 8 bytes.

		// We'll rely on a library to decode each format, which will deal with text and binary representations for us
		if err := h.pgTypeMap.Scan(uint32(types[i]), int16(formatCodes[i]), values[i].Data, &bindVarString); err != nil {
			return nil, err
		}
		bindVar := &querypb.BindVariable{
			Type:   typ,
			Value:  []byte(bindVarString),
			Values: nil, // TODO
		}
		bindings[bindingName] = bindVar
	}
	return bindings, nil
}

// TODO: we need to migrate this away from vitess types and deal strictly with OIDs which are compatible with Postgres types
func convertType(oid int32) querypb.Type {
	switch oid {
	// TODO: this should never be 0
	case 0:
		return sqltypes.Int32
	case messages.OidInt2:
		return sqltypes.Int16
	case messages.OidInt4:
		return sqltypes.Int32
	case messages.OidInt8:
		return sqltypes.Int64
	case messages.OidFloat4:
		return sqltypes.Float32
	case messages.OidFloat8:
		return sqltypes.Float64
	case messages.OidName:
		return sqltypes.Text
	case messages.OidNumeric:
		return sqltypes.Decimal
	case messages.OidText:
		return sqltypes.Text
	case messages.OidBool:
		return sqltypes.Bit
	case messages.OidDate:
		return sqltypes.Date
	case messages.OidTimestamp:
		return sqltypes.Timestamp
	case messages.OidVarchar:
		return sqltypes.Text
	case messages.OidOid:
		return sqltypes.Uint32
	default:
		panic(fmt.Sprintf("convertType(oid): unhandled type %d", oid))
	}
}

// sendClientStartupMessages sends introductory messages to the client and returns any error
// TODO: implement users and authentication
func (h *ConnectionHandler) sendClientStartupMessages(startupMessage messages.StartupMessage) error {
	if user, ok := startupMessage.Parameters["user"]; ok && len(user) > 0 {
		var host string
		if h.Conn().RemoteAddr().Network() == "unix" {
			host = "localhost"
		} else {
			host, _, _ = net.SplitHostPort(h.Conn().RemoteAddr().String())
			if len(host) == 0 {
				host = "localhost"
			}
		}

		h.mysqlConn.User = user
		h.mysqlConn.UserData = sql.MysqlConnectionUser{
			User: user,
			Host: host,
		}
	} else {
		h.mysqlConn.User = "doltgres"
		h.mysqlConn.UserData = sql.MysqlConnectionUser{
			User: "doltgres",
			Host: "localhost",
		}
	}

	if err := connection.Send(h.Conn(), messages.AuthenticationOk{}); err != nil {
		return err
	}

	if err := connection.Send(h.Conn(), messages.ParameterStatus{
		Name:  "server_version",
		Value: "15.0",
	}); err != nil {
		return err
	}

	if err := connection.Send(h.Conn(), messages.ParameterStatus{
		Name:  "client_encoding",
		Value: "UTF8",
	}); err != nil {
		return err
	}

	if err := connection.Send(h.Conn(), messages.BackendKeyData{
		ProcessID: processID,
		SecretKey: 0,
	}); err != nil {
		return err
	}

	return nil
}

// query runs the given query and sends a CommandComplete message to the client
func (h *ConnectionHandler) query(query ConvertedQuery) error {
	commandComplete := messages.CommandComplete{
		Query: query.String,
		Tag:   query.StatementTag,
	}

	err := h.comQuery(query, spoolRowsCallback(h.Conn(), &commandComplete, false))

	if err != nil {
		if strings.HasPrefix(err.Error(), "syntax error at position") {
			return fmt.Errorf("This statement is not yet supported")
		}
		return err
	}

	if err := connection.Send(h.Conn(), commandComplete); err != nil {
		return err
	}

	return nil
}

// spoolRowsCallback returns a callback function that will send RowDescription message, then a DataRow message for
// each row in the result set.
func spoolRowsCallback(conn net.Conn, commandComplete *messages.CommandComplete, isExecute bool) mysql.ResultSpoolFn {
	return func(res *sqltypes.Result, more bool) error {
		if messages.ReturnsRow(commandComplete.Tag) {
			// EXECUTE does not send RowDescription; instead it should be sent from DESCRIBE prior to it
			if !isExecute {
				if err := connection.Send(conn, messages.RowDescription{
					Fields: res.Fields,
				}); err != nil {
					return err
				}
			}

			for _, row := range res.Rows {
				if err := connection.Send(conn, messages.DataRow{
					Values: row,
				}); err != nil {
					return err
				}
			}
		}

		if commandComplete.IsIUD() {
			commandComplete.Rows = int32(res.RowsAffected)
		} else {
			commandComplete.Rows += int32(len(res.Rows))
		}
		return nil
	}
}

// sendDescribeResponse sends a response message for a Describe message
func (h *ConnectionHandler) sendDescribeResponse(conn net.Conn, fields []*querypb.Field, types []int32, tag string) (err error) {
	// The prepared statement variant of the describe command returns the OIDs of the parameters.
	if types != nil {
		if err := connection.Send(conn, messages.ParameterDescription{
			ObjectIDs: types,
		}); err != nil {
			return err
		}
	}

	if messages.ReturnsRow(tag) {
		// Both variants finish with a row description.
		return connection.Send(conn, messages.RowDescription{
			Fields: fields,
		})
	} else {
		return connection.Send(conn, messages.NoData{})
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
			String:       `SELECT table_schema AS 'Schema', TABLE_NAME AS 'Name', 'table' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA <> 'pg_catalog' AND CONVERT(TABLE_TYPE, CHAR) = 'BASE TABLE' ORDER BY 2;`,
			StatementTag: "SELECT",
		})
	}
	// Command: \d
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','v','m','s','f','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT table_schema AS 'Schema', TABLE_NAME AS 'Name', IF(TABLE_TYPE = 'VIEW', 'view', 'table') AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA <> 'pg_catalog' AND CONVERT(TABLE_TYPE, CHAR) = 'BASE TABLE' OR CONVERT(TABLE_TYPE, CHAR) = 'VIEW' ORDER BY 2;`,
			StatementTag: "SELECT",
		})
	}
	// Alternate \d for psql 14
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 's' then 'special' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','v','m','s','f','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT table_schema AS 'Schema', TABLE_NAME AS 'Name', IF(TABLE_TYPE = 'VIEW', 'view', 'table') AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA <> 'pg_catalog' AND CONVERT(TABLE_TYPE, CHAR) = 'BASE TABLE' OR CONVERT(TABLE_TYPE, CHAR) = 'VIEW' ORDER BY 2;`,
			StatementTag: "SELECT",
		})
	}
	// Command: \d table_name
	if strings.HasPrefix(statement, "select c.oid,\n  n.nspname,\n  c.relname\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\nwhere c.relname operator(pg_catalog.~) '^(") && strings.HasSuffix(statement, ")$' collate pg_catalog.default\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 2, 3;") {
		// There are >at least< 15 separate statements sent for this command, which is far too much to validate and
		// implement, so we'll just return an error for now
		return true, fmt.Errorf("PSQL command not yet supported")
	}
	// Command: \dn
	if statement == "select n.nspname as \"name\",\n  pg_catalog.pg_get_userbyid(n.nspowner) as \"owner\"\nfrom pg_catalog.pg_namespace n\nwhere n.nspname !~ '^pg_' and n.nspname <> 'information_schema'\norder by 1;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT 'public' AS 'Name', 'pg_database_owner' AS 'Owner';`,
			StatementTag: "SELECT",
		})
	}
	// Command: \df
	if statement == "select n.nspname as \"schema\",\n  p.proname as \"name\",\n  pg_catalog.pg_get_function_result(p.oid) as \"result data type\",\n  pg_catalog.pg_get_function_arguments(p.oid) as \"argument data types\",\n case p.prokind\n  when 'a' then 'agg'\n  when 'w' then 'window'\n  when 'p' then 'proc'\n  else 'func'\n end as \"type\"\nfrom pg_catalog.pg_proc p\n     left join pg_catalog.pg_namespace n on n.oid = p.pronamespace\nwhere pg_catalog.pg_function_is_visible(p.oid)\n      and n.nspname <> 'pg_catalog'\n      and n.nspname <> 'information_schema'\norder by 1, 2, 4;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT '' AS 'Schema', '' AS 'Name', '' AS 'Result data type', '' AS 'Argument data types', '' AS 'Type' FROM dual LIMIT 0;`,
			StatementTag: "SELECT",
		})
	}
	// Command: \dv
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\nwhere c.relkind in ('v','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT table_schema AS 'Schema', TABLE_NAME AS 'Name', 'view' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA <> 'pg_catalog' AND TABLE_TYPE = 'VIEW' ORDER BY 2;`,
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

// handledWorkbenchCommands handles commands used by some workbenches, such as dolt-workbench.
func (h *ConnectionHandler) handledWorkbenchCommands(statement string) (bool, error) {
	lower := strings.ToLower(statement)
	if lower == "select * from current_schema()" || lower == "select * from current_schema();" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT search_path AS 'current_schema';`,
			StatementTag: "SELECT",
		})
	}
	if lower == "select * from current_database()" || lower == "select * from current_database();" {
		return true, h.query(ConvertedQuery{
			String:       `SELECT DATABASE() AS 'current_database';`,
			StatementTag: "SELECT",
		})
	}
	// TODO: Fix timeout
	if strings.HasPrefix(statement, `SELECT "con"."conname" AS "constraint_name", "con"."nspname" AS "table_schema", "con"."relname" AS "table_name", "att2"."attname" AS "column_name", "ns"."nspname" AS "referenced_table_schema", "cl"."relname" AS "referenced_table_name", "att"."attname" AS "referenced_column_name", "con"."confdeltype" AS "on_delete", "con"."confupdtype" AS "on_update", "con"."condeferrable" AS "deferrable", "con"."condeferred" AS "deferred" FROM ( SELECT UNNEST ("con1"."conkey") AS "parent", UNNEST ("con1"."confkey") AS "child", "con1"."confrelid", "con1"."conrelid", "con1"."conname", "con1"."contype", "ns"."nspname", "cl"."relname", "con1"."condeferrable", CASE WHEN "con1"."condeferred" THEN 'INITIALLY DEFERRED' ELSE 'INITIALLY IMMEDIATE' END as condeferred, CASE "con1"."confdeltype" WHEN 'a' THEN 'NO ACTION' WHEN 'r' THEN 'RESTRICT' WHEN 'c' THEN 'CASCADE' WHEN 'n' THEN 'SET NULL' WHEN 'd' THEN 'SET DEFAULT' END as "confdeltype", CASE "con1"."confupdtype" WHEN 'a' THEN 'NO ACTION' WHEN 'r' THEN 'RESTRICT' WHEN 'c' THEN 'CASCADE' WHEN 'n' THEN 'SET NULL' WHEN 'd' THEN 'SET DEFAULT' END as "confupdtype" FROM "pg_class" "cl" INNER JOIN "pg_namespace" "ns" ON "cl"."relnamespace" = "ns"."oid" INNER JOIN "pg_constraint" "con1" ON "con1"."conrelid" = "cl"."oid" WHERE "con1"."contype" = 'f' AND (("ns"."nspname" = `) {
		// We'll just return nothing for now
		return true, h.query(ConvertedQuery{
			String:       `SELECT '' FROM dual LIMIT 0;`,
			StatementTag: "SELECT",
		})
	}
	// TODO: Fix timeout
	if strings.HasPrefix(statement, `SELECT "ns"."nspname" AS "table_schema", "t"."relname" AS "table_name", "i"."relname" AS "constraint_name", "a"."attname" AS "column_name", CASE "ix"."indisunique" WHEN 't' THEN 'TRUE' ELSE'FALSE' END AS "is_unique", pg_get_expr("ix"."indpred", "ix"."indrelid") AS "condition", "types"."typname" AS "type_name", "am"."amname" AS "index_type" FROM "pg_class" "t" INNER JOIN "pg_index" "ix" ON "ix"."indrelid" = "t"."oid" INNER JOIN "pg_attribute" "a" ON "a"."attrelid" = "t"."oid"  AND "a"."attnum" = ANY ("ix"."indkey") INNER JOIN "pg_namespace" "ns" ON "ns"."oid" = "t"."relnamespace" INNER JOIN "pg_class" "i" ON "i"."oid" = "ix"."indexrelid" INNER JOIN "pg_type" "types" ON "types"."oid" = "a"."atttypid" INNER JOIN "pg_am" "am" ON "i"."relam" = "am"."oid" LEFT JOIN "pg_constraint" "cnst" ON "cnst"."conname" = "i"."relname" WHERE "t"."relkind" IN ('r', 'p') AND "cnst"."contype" IS NULL AND (("ns"."nspname" `) {
		// We'll just return nothing for now
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
		h.sendError(h.Conn(), err)
	}
	if sendErr := connection.Send(h.Conn(), messages.ReadyForQuery{
		Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
	}); sendErr != nil {
		// We panic here for the same reason as above.
		panic(sendErr)
	}
}

// sendError sends the given error to the client. This should generally never be called directly.
func (h *ConnectionHandler) sendError(conn net.Conn, err error) {
	fmt.Println(err.Error())
	if sendErr := connection.Send(conn, messages.ErrorResponse{
		Severity:     messages.ErrorResponseSeverity_Error,
		SqlStateCode: "XX000", // internal_error for now
		Message:      err.Error(),
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
		return ConvertedQuery{}, fmt.Errorf("only a single statement at a time is currently supported")
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

// getPlanAndFields builds a plan and return fields for the given query
func (h *ConnectionHandler) getPlanAndFields(query ConvertedQuery) (sql.Node, []*querypb.Field, error) {
	if query.AST == nil {
		return nil, nil, fmt.Errorf("cannot prepare a query that has not been parsed")
	}

	parsedQuery, fields, err := h.handler.(mysql.ExtendedHandler).ComPrepareParsed(context.Background(), h.mysqlConn, query.String, query.AST, &mysql.PrepareData{
		PrepareStmt: query.String,
	})

	if err != nil {
		return nil, nil, err
	}

	plan, ok := parsedQuery.(sql.Node)
	if !ok {
		return nil, nil, fmt.Errorf("expected a sql.Node, got %T", parsedQuery)
	}

	return plan, fields, nil
}

// comQuery is a shortcut that determines which version of ComQuery to call based on whether the query has been parsed.
func (h *ConnectionHandler) comQuery(query ConvertedQuery, callback func(res *sqltypes.Result, more bool) error) error {
	if query.AST == nil {
		return h.handler.ComQuery(context.Background(), h.mysqlConn, query.String, callback)
	} else {
		return h.handler.(mysql.ExtendedHandler).ComParsedQuery(context.Background(), h.mysqlConn, query.String, query.AST, callback)
	}
}

// bindParams binds the paramters given to the query plan given and returns the resulting plan and fields.
func (h *ConnectionHandler) bindParams(
	query string,
	parsedQuery sqlparser.Statement,
	bindVars map[string]*querypb.BindVariable,
) (sql.Node, []*querypb.Field, error) {
	bound, fields, err := h.handler.(mysql.ExtendedHandler).ComBind(context.Background(), h.mysqlConn, query, parsedQuery, &mysql.PrepareData{
		PrepareStmt: query,
		ParamsCount: uint16(len(bindVars)),
		BindVars:    bindVars,
	})

	if err != nil {
		return nil, nil, err
	}

	plan, ok := bound.(sql.Node)
	if !ok {
		return nil, nil, fmt.Errorf("expected a sql.Node, got %T", bound)
	}

	return plan, fields, err
}

// discardAll handles the DISCARD ALL command
func (h *ConnectionHandler) discardAll(query ConvertedQuery, conn net.Conn) error {
	err := h.handler.ComResetConnection(h.mysqlConn)
	if err != nil {
		return err
	}

	commandComplete := messages.CommandComplete{
		Query: query.String,
		Tag:   query.StatementTag,
	}

	return connection.Send(conn, commandComplete)
}
