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
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/mysql_db"
	plan2 "github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/sqltypes"
	querypb "github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/postgres/connection"
	"github.com/dolthub/doltgresql/postgres/messages"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"
	_ "github.com/dolthub/doltgresql/server/functions"
)

var (
	connectionIDCounter uint32
	processID           = int32(os.Getpid())
	certificate         tls.Certificate //TODO: move this into the mysql.ListenerConfig
)

// Listener listens for connections to process PostgreSQL requests into Dolt requests.
type Listener struct {
	listener net.Listener
	cfg      mysql.ListenerConfig
}

var _ server.ProtocolListener = (*Listener)(nil)

// NewListener creates a new Listener.
func NewListener(listenerCfg mysql.ListenerConfig) (server.ProtocolListener, error) {
	return &Listener{
		listener: listenerCfg.Listener,
		cfg:      listenerCfg,
	}, nil
}

// Accept handles incoming connections.
func (l *Listener) Accept() {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			if err.Error() == "use of closed network connection" {
				break
			}
			fmt.Printf("Unable to accept connection:\n%v\n", err)
			continue
		}

		go l.HandleConnection(conn)
	}
}

// Close stops the handling of incoming connections.
func (l *Listener) Close() {
	_ = l.listener.Close()
}

// Addr returns the address that the listener is listening on.
func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

// HandleConnection handles a connection's session.
func (l *Listener) HandleConnection(conn net.Conn) {
	mysqlConn := &mysql.Conn{
		Conn:        conn,
		PrepareData: make(map[uint32]*mysql.PrepareData),
	}
	mysqlConn.ConnectionID = atomic.AddUint32(&connectionIDCounter, 1)

	var returnErr error
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Listener recovered panic: %v", r)
		}
		if returnErr != nil {
			fmt.Println(returnErr.Error())
		}
		l.cfg.Handler.ConnectionClosed(mysqlConn)
		if err := conn.Close(); err != nil {
			fmt.Printf("Failed to properly close connection:\n%v\n", err)
		}
	}()
	l.cfg.Handler.NewConnection(mysqlConn)

	startupMessage, conn, err := l.receiveStartupMessage(conn, mysqlConn)
	if err != nil {
		returnErr = err
		return
	}

	err = l.sendClientStartupMessages(conn, startupMessage, mysqlConn)
	if err != nil {
		returnErr = err
		return
	}

	err = l.chooseInitialDatabase(conn, startupMessage, mysqlConn)
	if err != nil {
		returnErr = err
		return
	}

	if err := connection.Send(conn, messages.ReadyForQuery{
		Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
	}); err != nil {
		returnErr = err
		return
	}

	// Postgres has a two-stage procedure for prepared queries. First the query is parsed via a |Parse| message, and
	// the result is stored in the |preparedStatements| map by the name provided. Then one or more |Bind| messages
	// provide parameters for the query, and the result is stored in |portals|. Finally, a call to |Execute| executes
	// the named portal.
	preparedStatements := make(map[string]PreparedStatementData)
	portals := make(map[string]PortalData)

	// Main session loop: read messages one at a time off the connection until we receive a |Terminate| message, in
	// which case we hang up, or the connection is closed by the client, which generates an io.EOF from the connection.
	for {
		message, err := connection.Receive(conn)
		if err != nil {
			returnErr = err
			return
		}

		if ds, ok := message.(sql.DebugStringer); ok && logrus.IsLevelEnabled(logrus.DebugLevel) {
			logrus.Debugf("Received message: %s", ds.DebugString())
		} else {
			logrus.Debugf("Received message: %s", message.DefaultMessage().Name)
		}

		stop, endOfMessages, err := l.handleMessage(message, conn, mysqlConn, preparedStatements, portals)
		if err != nil {
			if !endOfMessages {
				if syncErr := connection.DiscardToSync(conn); syncErr != nil {
					fmt.Println(syncErr.Error())
				}
			}
			l.endOfMessages(conn, err)
		} else if endOfMessages {
			l.endOfMessages(conn, nil)
		}

		if stop {
			returnErr = err
			break
		}
	}
}

// receiveStarupMessage reads a startup message from the connection given and returns it. Some startup messages will
// result in the establishment of a new connection, which is also returned.
func (l *Listener) receiveStartupMessage(conn net.Conn, mysqlConn *mysql.Conn) (messages.StartupMessage, net.Conn, error) {
	var startupMessage messages.StartupMessage
	// The initial message may be one of a few different messages, so we'll check for those.
InitialMessageLoop:
	for {
		initialMessages, err := connection.ReceiveIntoAny(conn,
			messages.StartupMessage{},
			messages.SSLRequest{},
			messages.GSSENCRequest{})
		if err != nil {
			if err == io.EOF {
				return messages.StartupMessage{}, nil, nil
			}
			return messages.StartupMessage{}, nil, err
		}

		if len(initialMessages) != 1 {
			return messages.StartupMessage{}, nil, fmt.Errorf("expected a single message upon starting connection, terminating connection")
		}

		initialMessage := initialMessages[0]
		switch initialMessage := initialMessage.(type) {
		case messages.StartupMessage:
			startupMessage = initialMessage
			break InitialMessageLoop
		case messages.SSLRequest:
			hasCertificate := len(certificate.Certificate) > 0
			if err := connection.Send(conn, messages.SSLResponse{
				SupportsSSL: hasCertificate,
			}); err != nil {
				return messages.StartupMessage{}, nil, err
			}
			// If we have a certificate and the client has asked for SSL support, then we switch here.
			// We can't start in SSL mode, as the client does not attempt the handshake until after our response.
			if hasCertificate {
				conn = tls.Server(conn, &tls.Config{
					Certificates: []tls.Certificate{certificate},
				})
				mysqlConn.Conn = conn
			}
		case messages.GSSENCRequest:
			if err = connection.Send(conn, messages.GSSENCResponse{
				SupportsGSSAPI: false,
			}); err != nil {
				return messages.StartupMessage{}, nil, err
			}
		default:
			return messages.StartupMessage{}, nil, fmt.Errorf("unexpected initial message, terminating connection")
		}
	}

	return startupMessage, conn, nil
}

func (l *Listener) chooseInitialDatabase(conn net.Conn, startupMessage messages.StartupMessage, mysqlConn *mysql.Conn) error {
	if db, ok := startupMessage.Parameters["database"]; ok && len(db) > 0 {
		err := l.cfg.Handler.ComQuery(mysqlConn, fmt.Sprintf("USE `%s`;", db), func(res *sqltypes.Result, more bool) error {
			return nil
		})
		if err != nil {
			_ = connection.Send(conn, messages.ErrorResponse{
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
		_ = l.cfg.Handler.ComQuery(mysqlConn, fmt.Sprintf("USE `%s`;", mysqlConn.User), func(*sqltypes.Result, bool) error {
			return nil
		})
	}
	return nil
}

func (l *Listener) handleMessage(
	message connection.Message,
	conn net.Conn,
	mysqlConn *mysql.Conn,
	preparedStatements map[string]PreparedStatementData,
	portals map[string]PortalData,
) (stop, endOfMessages bool, err error) {
	switch message := message.(type) {
	case messages.Terminate:
		return true, false, nil
	case messages.Sync:
		return false, true, nil
	case messages.Query:
		return l.handleQuery(message, preparedStatements, portals, mysqlConn, conn)
	case messages.Parse:
		return l.handleParse(message, preparedStatements, mysqlConn, conn)
	case messages.Describe:
		return l.handleDescribe(message, preparedStatements, portals, conn)
	case messages.Bind:
		return l.handleBind(message, preparedStatements, portals, conn, mysqlConn)
	case messages.Execute:
		return l.handleExecute(message, portals, conn, mysqlConn)
	case messages.Close:
		if message.ClosingPreparedStatement {
			delete(preparedStatements, message.Target)
		} else {
			delete(portals, message.Target)
		}

		return false, false, connection.Send(conn, messages.CloseComplete{})
	default:
		return false, true, fmt.Errorf(`Unhandled message "%s"`, message.DefaultMessage().Name)
	}
}

func (l *Listener) handleQuery(message messages.Query, preparedStatements map[string]PreparedStatementData, portals map[string]PortalData, mysqlConn *mysql.Conn, conn net.Conn) (bool, bool, error) {
	handled, err := l.handledPSQLCommands(conn, mysqlConn, message.String)
	if handled || err != nil {
		return false, true, err
	}

	query, err := l.convertQuery(message.String)
	if err != nil {
		return false, true, err
	}

	// A query message destroys the unnamed statement and the unnamed portal
	delete(preparedStatements, "")
	delete(portals, "")

	// The Deallocate message does not get passed to the engine, since we handle allocation / deallocation of
	// prepared statements at this layer
	switch stmt := query.AST.(type) {
	case *sqlparser.Deallocate:
		// TODO: handle ALL keyword
		return false, true, l.deallocatePreparedStatement(stmt.Name, preparedStatements, query, conn)
	}

	return false, true, l.query(conn, mysqlConn, query)
}

func (l *Listener) handleParse(message messages.Parse, preparedStatements map[string]PreparedStatementData, mysqlConn *mysql.Conn, conn net.Conn) (bool, bool, error) {
	// TODO: "Named prepared statements must be explicitly closed before they can be redefined by another Parse message, but this is not required for the unnamed statement"
	query, err := l.convertQuery(message.Query)
	if err != nil {
		return false, false, err
	}

	if query.AST == nil {
		// special case: empty query
		preparedStatements[message.Name] = PreparedStatementData{
			Query: query,
		}
		return false, false, nil
	}

	plan, fields, err := l.getPlanAndFields(mysqlConn, query)
	if err != nil {
		return false, false, err
	}

	// TODO: bindvar types can be specified directly in the message, need tests of this
	bindVarTypes, err := extractBindVarTypes(plan)
	if err != nil {
		return false, false, err
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

	preparedStatements[message.Name] = PreparedStatementData{
		Query:        query,
		ReturnFields: fields,
		BindVarTypes: bindVarTypes,
	}

	return false, false, connection.Send(conn, messages.ParseComplete{})
}

func (l *Listener) handleDescribe(message messages.Describe, preparedStatements map[string]PreparedStatementData, portals map[string]PortalData, conn net.Conn) (bool, bool, error) {
	var fields []*querypb.Field
	var bindvarTypes []int32

	if message.IsPrepared {
		preparedStatementData, ok := preparedStatements[message.Target]
		if !ok {
			return false, true, fmt.Errorf("prepared statement %s does not exist", message.Target)
		}

		fields = preparedStatementData.ReturnFields
		bindvarTypes = preparedStatementData.BindVarTypes
	} else {
		portalData, ok := portals[message.Target]
		if !ok {
			return false, true, fmt.Errorf("portal %s does not exist", message.Target)
		}

		fields = portalData.Fields
	}

	return false, false, l.describe(conn, fields, bindvarTypes)
}

func (l *Listener) handleBind(message messages.Bind, preparedStatements map[string]PreparedStatementData, portals map[string]PortalData, conn net.Conn, mysqlConn *mysql.Conn) (bool, bool, error) {
	// TODO: a named portal object lasts till the end of the current transaction, unless explicitly destroyed
	//  we need to destroy the named portal as a side effect of the transaction ending
	logrus.Tracef("binding portal %q to prepared statement %s", message.DestinationPortal, message.SourcePreparedStatement)
	preparedData, ok := preparedStatements[message.SourcePreparedStatement]
	if !ok {
		return false, true, fmt.Errorf("prepared statement %s does not exist", message.SourcePreparedStatement)
	}

	if preparedData.Query.AST == nil {
		// special case: empty query
		portals[message.DestinationPortal] = PortalData{
			Query:        preparedData.Query,
			IsEmptyQuery: true,
		}
		return false, false, connection.Send(conn, messages.BindComplete{})
	}

	bindVars, err := convertBindParameters(preparedData.BindVarTypes, message.ParameterValues)
	if err != nil {
		return false, false, err
	}

	boundPlan, fields, err := l.bindParams(mysqlConn, message.SourcePreparedStatement, preparedData.Query.AST, bindVars)
	if err != nil {
		return false, false, err
	}

	portals[message.DestinationPortal] = PortalData{
		Query:     preparedData.Query,
		Fields:    fields,
		BoundPlan: boundPlan,
	}
	return false, false, connection.Send(conn, messages.BindComplete{})
}

func (l *Listener) handleExecute(message messages.Execute, portals map[string]PortalData, conn net.Conn, mysqlConn *mysql.Conn) (bool, bool, error) {
	// TODO: implement the RowMax
	portalData, ok := portals[message.Portal]
	if !ok {
		return false, false, fmt.Errorf("portal %s does not exist", message.Portal)
	}

	logrus.Tracef("executing portal %s with contents %v", message.Portal, portalData)
	query := portalData.Query

	// we need the CommandComplete message defined here because it's altered by the callback below
	complete := messages.CommandComplete{
		Query: query.String,
	}

	if !portalData.IsEmptyQuery {
		err := l.cfg.Handler.(mysql.ExtendedHandler).ComExecuteBound(mysqlConn, query.String, portalData.BoundPlan, spoolRowsCallback(conn, complete))
		if err != nil {
			return false, false, err
		}
	}

	return false, false, connection.Send(conn, complete)
}

func (l *Listener) deallocatePreparedStatement(name string, preparedStatements map[string]PreparedStatementData, query ConvertedQuery, conn net.Conn) error {
	_, ok := preparedStatements[name]
	if !ok {
		return fmt.Errorf("prepared statement %s does not exist", name)
	}
	delete(preparedStatements, name)

	commandComplete := messages.CommandComplete{
		Query: query.String,
	}

	return connection.Send(conn, commandComplete)
}

func extractBindVarTypes(queryPlan sql.Node) ([]int32, error) {
	inspectNode := queryPlan
	switch queryPlan := queryPlan.(type) {
	case *plan2.InsertInto:
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
			oid, err = messages.VitessTypeToObjectID(e.Type().Type())
			if err != nil {
				err = fmt.Errorf("could not determine OID for placeholder %s: %w", e.Name, err)
				return false
			}
			types = append(types, oid)
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

func convertBindParameters(types []int32, values []messages.BindParameterValue) (map[string]*querypb.BindVariable, error) {
	bindings := make(map[string]*querypb.BindVariable, len(values))
	for i, value := range values {
		bindingName := fmt.Sprintf("v%d", i+1)
		typ := convertType(types[i])
		bindVar := &querypb.BindVariable{
			Type:   typ,
			Value:  convertBindVarValue(typ, value),
			Values: nil, // TODO
		}
		bindings[bindingName] = bindVar
	}
	return bindings, nil
}

func convertBindVarValue(typ querypb.Type, value messages.BindParameterValue) []byte {
	switch typ {
	case querypb.Type_INT8, querypb.Type_INT16, querypb.Type_INT24, querypb.Type_INT32, querypb.Type_UINT8, querypb.Type_UINT16, querypb.Type_UINT24, querypb.Type_UINT32:
		// first convert the bytes in the payload to an integer, then convert that to its base 10 string representation
		intVal := binary.BigEndian.Uint32(value.Data) // TODO: bound check
		return []byte(strconv.FormatUint(uint64(intVal), 10))
	case querypb.Type_INT64, querypb.Type_UINT64:
		// first convert the bytes in the payload to an integer, then convert that to its base 10 string representation
		intVal := binary.BigEndian.Uint64(value.Data)
		return []byte(strconv.FormatUint(intVal, 10))
	case querypb.Type_FLOAT32, querypb.Type_FLOAT64:
		// first convert the bytes in the payload to a float, then convert that to its base 10 string representation
		floatVal := binary.BigEndian.Uint64(value.Data) // TODO: bound check
		return []byte(strconv.FormatFloat(math.Float64frombits(floatVal), 'f', -1, 64))
	case querypb.Type_VARCHAR, querypb.Type_VARBINARY, querypb.Type_TEXT, querypb.Type_BLOB:
		return value.Data
	default:
		panic(fmt.Sprintf("unhandled type %v", typ))
	}
}

func convertType(oid int32) querypb.Type {
	switch oid {
	// TODO: this should never be 0
	case 0:
		return sqltypes.Int32
	case messages.OidInt4:
		return sqltypes.Int32
	case messages.OidInt8:
		return sqltypes.Int64
	case messages.OidFloat4:
		return sqltypes.Float32
	case messages.OidFloat8:
		return sqltypes.Float64
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
	default:
		panic(fmt.Sprintf("unhandled type %d", oid))
	}
}

// sendClientStartupMessages sends introductory messages to the client and returns any error
// TODO: implement users and authentication
func (l *Listener) sendClientStartupMessages(conn net.Conn, startupMessage messages.StartupMessage, mysqlConn *mysql.Conn) error {
	if user, ok := startupMessage.Parameters["user"]; ok && len(user) > 0 {
		var host string
		if conn.RemoteAddr().Network() == "unix" {
			host = "localhost"
		} else {
			host, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
			if len(host) == 0 {
				host = "localhost"
			}
		}
		mysqlConn.User = user
		mysqlConn.UserData = mysql_db.MysqlConnectionUser{
			User: user,
			Host: host,
		}
	} else {
		mysqlConn.User = "doltgres"
		mysqlConn.UserData = mysql_db.MysqlConnectionUser{
			User: "doltgres",
			Host: "localhost",
		}
	}

	if err := connection.Send(conn, messages.AuthenticationOk{}); err != nil {
		return err
	}

	if err := connection.Send(conn, messages.ParameterStatus{
		Name:  "server_version",
		Value: "15.0",
	}); err != nil {
		return err
	}

	if err := connection.Send(conn, messages.ParameterStatus{
		Name:  "client_encoding",
		Value: "UTF8",
	}); err != nil {
		return err
	}

	if err := connection.Send(conn, messages.BackendKeyData{
		ProcessID: processID,
		SecretKey: 0,
	}); err != nil {
		return err
	}

	return nil
}

// query runs the given query and sends a CommandComplete message to the client
func (l *Listener) query(conn net.Conn, mysqlConn *mysql.Conn, query ConvertedQuery) error {
	commandComplete := messages.CommandComplete{
		Query: query.String,
	}

	err := l.comQuery(mysqlConn, query, spoolRowsCallback(conn, commandComplete))

	if err != nil {
		if strings.HasPrefix(err.Error(), "syntax error at position") {
			return fmt.Errorf("This statement is not yet supported")
		}
		return err
	}

	if err := connection.Send(conn, commandComplete); err != nil {
		return err
	}

	return nil
}

// spoolRowsCallback returns a callback function that will send RowDescription message, then a DataRow message for
// each row in the result set.
func spoolRowsCallback(conn net.Conn, commandComplete messages.CommandComplete) mysql.ResultSpoolFn {
	return func(res *sqltypes.Result, more bool) error {
		if err := connection.Send(conn, messages.RowDescription{
			Fields: res.Fields,
		}); err != nil {
			return err
		}

		for _, row := range res.Rows {
			if err := connection.Send(conn, messages.DataRow{
				Values: row,
			}); err != nil {
				return err
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

// describe handles the description of the given query. This will post the ParameterDescription and RowDescription messages.
func (l *Listener) describe(conn net.Conn, fields []*querypb.Field, types []int32) (err error) {
	// The prepared statement variant of the describe command returns the OIDs of the parameters.
	if types != nil {
		if err := connection.Send(conn, messages.ParameterDescription{
			ObjectIDs: types,
		}); err != nil {
			return err
		}
	}

	// Both variants finish with a row description.
	if err := connection.Send(conn, messages.RowDescription{
		Fields: fields,
	}); err != nil {
		return err
	}

	return nil
}

// handledPSQLCommands handles the special PSQL commands, such as \l and \dt.
func (l *Listener) handledPSQLCommands(conn net.Conn, mysqlConn *mysql.Conn, statement string) (bool, error) {
	statement = strings.ToLower(statement)
	// Command: \l
	if statement == "select d.datname as \"name\",\n       pg_catalog.pg_get_userbyid(d.datdba) as \"owner\",\n       pg_catalog.pg_encoding_to_char(d.encoding) as \"encoding\",\n       d.datcollate as \"collate\",\n       d.datctype as \"ctype\",\n       d.daticulocale as \"icu locale\",\n       case d.datlocprovider when 'c' then 'libc' when 'i' then 'icu' end as \"locale provider\",\n       pg_catalog.array_to_string(d.datacl, e'\\n') as \"access privileges\"\nfrom pg_catalog.pg_database d\norder by 1;" {
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: `SELECT SCHEMA_NAME AS 'Name', 'postgres' AS 'Owner', 'UTF8' AS 'Encoding', 'English_United States.1252' AS 'Collate', 'English_United States.1252' AS 'Ctype', '' AS 'ICU Locale', 'libc' AS 'Locale Provider', '' AS 'Access privileges' FROM INFORMATION_SCHEMA.SCHEMATA ORDER BY 1;`})
	}
	// Command: \l on psql 16
	if statement == "select\n  d.datname as \"name\",\n  pg_catalog.pg_get_userbyid(d.datdba) as \"owner\",\n  pg_catalog.pg_encoding_to_char(d.encoding) as \"encoding\",\n  case d.datlocprovider when 'c' then 'libc' when 'i' then 'icu' end as \"locale provider\",\n  d.datcollate as \"collate\",\n  d.datctype as \"ctype\",\n  d.daticulocale as \"icu locale\",\n  null as \"icu rules\",\n  pg_catalog.array_to_string(d.datacl, e'\\n') as \"access privileges\"\nfrom pg_catalog.pg_database d\norder by 1;" {
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: `SELECT SCHEMA_NAME AS 'Name', 'postgres' AS 'Owner', 'UTF8' AS 'Encoding', 'English_United States.1252' AS 'Collate', 'English_United States.1252' AS 'Ctype', '' AS 'ICU Locale', 'libc' AS 'Locale Provider', '' AS 'Access privileges' FROM INFORMATION_SCHEMA.SCHEMATA ORDER BY 1;`})
	}
	// Command: \dt
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: `SELECT 'public' AS 'Schema', TABLE_NAME AS 'Name', 'table' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = database() AND TABLE_TYPE = 'BASE TABLE' ORDER BY 2;`})
	}
	// Command: \d
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','v','m','s','f','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: `SELECT 'public' AS 'Schema', TABLE_NAME AS 'Name', 'table' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = database() AND TABLE_TYPE = 'BASE TABLE' ORDER BY 2;`})
	}
	// Alternate \d for psql 14
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 's' then 'special' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','v','m','s','f','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: `SELECT 'public' AS 'Schema', TABLE_NAME AS 'Name', 'table' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = database() AND TABLE_TYPE = 'BASE TABLE' ORDER BY 2;`})
	}
	// Command: \d table_name
	if strings.HasPrefix(statement, "select c.oid,\n  n.nspname,\n  c.relname\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\nwhere c.relname operator(pg_catalog.~) '^(") && strings.HasSuffix(statement, ")$' collate pg_catalog.default\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 2, 3;") {
		// There are >at least< 15 separate statements sent for this command, which is far too much to validate and
		// implement, so we'll just return an error for now
		return true, fmt.Errorf("PSQL command not yet supported")
	}
	// Command: \dn
	if statement == "select n.nspname as \"name\",\n  pg_catalog.pg_get_userbyid(n.nspowner) as \"owner\"\nfrom pg_catalog.pg_namespace n\nwhere n.nspname !~ '^pg_' and n.nspname <> 'information_schema'\norder by 1;" {
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: "SELECT 'public' AS 'Name', 'pg_database_owner' AS 'Owner';"})
	}
	// Command: \df
	if statement == "select n.nspname as \"schema\",\n  p.proname as \"name\",\n  pg_catalog.pg_get_function_result(p.oid) as \"result data type\",\n  pg_catalog.pg_get_function_arguments(p.oid) as \"argument data types\",\n case p.prokind\n  when 'a' then 'agg'\n  when 'w' then 'window'\n  when 'p' then 'proc'\n  else 'func'\n end as \"type\"\nfrom pg_catalog.pg_proc p\n     left join pg_catalog.pg_namespace n on n.oid = p.pronamespace\nwhere pg_catalog.pg_function_is_visible(p.oid)\n      and n.nspname <> 'pg_catalog'\n      and n.nspname <> 'information_schema'\norder by 1, 2, 4;" {
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: "SELECT '' AS 'Schema', '' AS 'Name', '' AS 'Result data type', '' AS 'Argument data types', '' AS 'Type' FROM dual LIMIT 0;"})
	}
	// Command: \dv
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\nwhere c.relkind in ('v','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: "SELECT 'public' AS 'Schema', TABLE_NAME AS 'Name', 'view' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = database() AND TABLE_TYPE = 'VIEW' ORDER BY 2;"})
	}
	// Command: \du
	if statement == "select r.rolname, r.rolsuper, r.rolinherit,\n  r.rolcreaterole, r.rolcreatedb, r.rolcanlogin,\n  r.rolconnlimit, r.rolvaliduntil,\n  array(select b.rolname\n        from pg_catalog.pg_auth_members m\n        join pg_catalog.pg_roles b on (m.roleid = b.oid)\n        where m.member = r.oid) as memberof\n, r.rolreplication\n, r.rolbypassrls\nfrom pg_catalog.pg_roles r\nwhere r.rolname !~ '^pg_'\norder by 1;" {
		// We don't support users yet, so we'll just return nothing for now
		return true, l.query(conn, mysqlConn, ConvertedQuery{String: "SELECT '' FROM dual LIMIT 0;"})
	}
	return false, nil
}

// endOfMessages should be called from HandleConnection or a function within HandleConnection. This represents the end
// of the message slice, which may occur naturally (all relevant response messages have been sent) or on error. Once
// endOfMessages has been called, no further messages should be sent, and the connection loop should wait for the next
// query. A nil error should be provided if this is being called naturally.
func (l *Listener) endOfMessages(conn net.Conn, err error) {
	if err != nil {
		l.sendError(conn, err)
	}
	if sendErr := connection.Send(conn, messages.ReadyForQuery{
		Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
	}); sendErr != nil {
		// We panic here for the same reason as above.
		panic(sendErr)
	}
}

// sendError sends the given error to the client. This should generally never be called directly.
func (l *Listener) sendError(conn net.Conn, err error) {
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
func (l *Listener) convertQuery(query string) (ConvertedQuery, error) {
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
	if err != nil {
		return ConvertedQuery{}, err
	}
	if vitessAST == nil {
		return ConvertedQuery{String: s[0].AST.String()}, nil
	}
	return ConvertedQuery{
		String: query,
		AST:    vitessAST,
	}, nil
}

// getPlanAndFields builds a plan and return fields for the given query
func (l *Listener) getPlanAndFields(mysqlConn *mysql.Conn, query ConvertedQuery) (sql.Node, []*querypb.Field, error) {
	if query.AST == nil {
		return nil, nil, fmt.Errorf("cannot prepare a query that has not been parsed")
	}

	parsedQuery, fields, err := l.cfg.Handler.(mysql.ExtendedHandler).ComPrepareParsed(mysqlConn, query.String, query.AST, &mysql.PrepareData{
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
func (l *Listener) comQuery(mysqlConn *mysql.Conn, query ConvertedQuery, callback func(res *sqltypes.Result, more bool) error) error {
	if query.AST == nil {
		return l.cfg.Handler.ComQuery(mysqlConn, query.String, callback)
	} else {
		return l.cfg.Handler.(mysql.ExtendedHandler).ComParsedQuery(mysqlConn, query.String, query.AST, callback)
	}
}

func (l *Listener) bindParams(
	mysqlConn *mysql.Conn,
	query string,
	parsedQuery sqlparser.Statement,
	bindVars map[string]*querypb.BindVariable,
) (sql.Node, []*querypb.Field, error) {
	bound, fields, err := l.cfg.Handler.(mysql.ExtendedHandler).ComBind(mysqlConn, query, parsedQuery, &mysql.PrepareData{
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
