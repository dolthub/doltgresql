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
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync/atomic"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/mysql_db"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/postgres/connection"
	"github.com/dolthub/doltgresql/postgres/messages"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"
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
			fmt.Printf("Linener recovered panic: %v", r)
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
	preparedStatements := make(map[string]ConvertedQuery)
	portals := make(map[string]ConvertedQuery)

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
	preparedStatements, portals map[string]ConvertedQuery,
) (stop, endOfMessages bool, err error) {
	switch message := message.(type) {
	case messages.Terminate:
		return true, false, nil
	case messages.Execute:
		// TODO: implement the RowMax
		logrus.Tracef("executing portal %s with contents %v", message.Portal, portals[message.Portal])
		return false, false, l.execute(conn, mysqlConn, portals[message.Portal])
	case messages.Query:
		handled, err := l.handledPSQLCommands(conn, mysqlConn, message.String)
		if handled || err != nil {
			return false, true, err
		}

		query, err := l.convertQuery(message.String)
		if err != nil {
			return false, true, err
		}

		// The Deallocate message must not get passed to the engine, since we handle allocation / deallocation of
		// prepared statements at this layer
		switch stmt := query.AST.(type) {
		case *sqlparser.Deallocate:
			_, ok := preparedStatements[stmt.Name]
			if !ok {
				return false, true, fmt.Errorf("prepared statement %s does not exist", stmt.Name)
			}
			delete(preparedStatements, stmt.Name)

			commandComplete := messages.CommandComplete{
				Query: query.String,
				Rows:  0,
			}

			return false, true, connection.Send(conn, commandComplete)
		default:
			return false, true, l.execute(conn, mysqlConn, query)
		}
	case messages.Parse:
		// TODO: fully support prepared statements
		if query, err := l.convertQuery(message.Query); err != nil {
			return false, false, err
		} else {
			preparedStatements[message.Name] = query
		}

		return false, false, connection.Send(conn, messages.ParseComplete{})
	case messages.Describe:
		var query ConvertedQuery
		if message.IsPrepared {
			query = preparedStatements[message.Target]
		} else {
			query = portals[message.Target]
		}

		return false, false, l.describe(conn, mysqlConn, message, query)
	case messages.Sync:
		return false, true, nil
	case messages.Bind:
		logrus.Tracef("binding portal %q to prepared statement %s", message.DestinationPortal, message.SourcePreparedStatement)
		// TODO: fully support prepared statements
		portals[message.DestinationPortal] = preparedStatements[message.SourcePreparedStatement]
		return false, false, connection.Send(conn, messages.BindComplete{})
	default:
		return false, true, fmt.Errorf(`Unhandled message "%s"`, message.DefaultMessage().Name)
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

// execute handles running the given query. This will post the RowDescription, DataRow, and CommandComplete messages.
func (l *Listener) execute(conn net.Conn, mysqlConn *mysql.Conn, query ConvertedQuery) error {
	commandComplete := messages.CommandComplete{
		Query: query.String,
		Rows:  0,
	}

	if err := l.comQuery(mysqlConn, query, func(res *sqltypes.Result, more bool) error {
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
	}); err != nil {
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

// describe handles the description of the given query. This will post the ParameterDescription and RowDescription messages.
func (l *Listener) describe(conn net.Conn, mysqlConn *mysql.Conn, message messages.Describe, statement ConvertedQuery) (err error) {
	logrus.Tracef("describing statement %v", statement)

	//TODO: fully support prepared statements
	if err := connection.Send(conn, messages.ParameterDescription{
		ObjectIDs: nil,
	}); err != nil {
		return err
	}

	//TODO: properly handle these statements
	if ImplicitlyCommits(statement.String) {
		if reverseStatement, ok := HandleImplicitCommitStatement(statement.String); ok {
			// We have a reverse statement that can function as a workaround for the lack of proper rollback support.
			// This does mean that we'll still create an implicit commit, but we can fix that whenever we add proper
			// transaction support.
			defer func() {
				// If there's an error, then we don't want to execute the reverse statement
				if err == nil {
					_ = l.cfg.Handler.ComQuery(mysqlConn, reverseStatement, func(_ *sqltypes.Result, _ bool) error {
						return nil
					})
				}
			}()
		} else {
			return fmt.Errorf("We do not yet support the Describe message for the given statement")
		}
	}
	// We'll start a transaction, so that we can later rollback any changes that were made.
	//TODO: handle the case where we are already in a transaction (SAVEPOINT will sometimes fail it seems?)
	if err := l.cfg.Handler.ComQuery(mysqlConn, "START TRANSACTION;", func(_ *sqltypes.Result, _ bool) error {
		return nil
	}); err != nil {
		return err
	}
	// We need to defer the rollback, so that it will always be executed.
	defer func() {
		_ = l.cfg.Handler.ComQuery(mysqlConn, "ROLLBACK;", func(_ *sqltypes.Result, _ bool) error {
			return nil
		})
	}()
	// Execute the statement, and send the description.
	if err := l.comQuery(mysqlConn, statement, func(res *sqltypes.Result, more bool) error {
		if res != nil {
			if err := connection.Send(conn, messages.RowDescription{
				Fields: res.Fields,
			}); err != nil {
				return err
			}
		}
		return nil
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
		return true, l.execute(conn, mysqlConn, ConvertedQuery{`SELECT SCHEMA_NAME AS 'Name', 'postgres' AS 'Owner', 'UTF8' AS 'Encoding', 'English_United States.1252' AS 'Collate', 'English_United States.1252' AS 'Ctype', '' AS 'ICU Locale', 'libc' AS 'Locale Provider', '' AS 'Access privileges' FROM INFORMATION_SCHEMA.SCHEMATA ORDER BY 1;`, nil})
	}
	// Command: \l on psql 16
	if statement == "select\n  d.datname as \"name\",\n  pg_catalog.pg_get_userbyid(d.datdba) as \"owner\",\n  pg_catalog.pg_encoding_to_char(d.encoding) as \"encoding\",\n  case d.datlocprovider when 'c' then 'libc' when 'i' then 'icu' end as \"locale provider\",\n  d.datcollate as \"collate\",\n  d.datctype as \"ctype\",\n  d.daticulocale as \"icu locale\",\n  null as \"icu rules\",\n  pg_catalog.array_to_string(d.datacl, e'\\n') as \"access privileges\"\nfrom pg_catalog.pg_database d\norder by 1;" {
		return true, l.execute(conn, mysqlConn, ConvertedQuery{`SELECT SCHEMA_NAME AS 'Name', 'postgres' AS 'Owner', 'UTF8' AS 'Encoding', 'English_United States.1252' AS 'Collate', 'English_United States.1252' AS 'Ctype', '' AS 'ICU Locale', 'libc' AS 'Locale Provider', '' AS 'Access privileges' FROM INFORMATION_SCHEMA.SCHEMATA ORDER BY 1;`, nil})
	}
	// Command: \dt
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, l.execute(conn, mysqlConn, ConvertedQuery{`SELECT 'public' AS 'Schema', TABLE_NAME AS 'Name', 'table' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = database() AND TABLE_TYPE = 'BASE TABLE' ORDER BY 2;`, nil})
	}
	// Command: \d
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','v','m','s','f','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, l.execute(conn, mysqlConn, ConvertedQuery{`SELECT 'public' AS 'Schema', TABLE_NAME AS 'Name', 'table' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = database() AND TABLE_TYPE = 'BASE TABLE' ORDER BY 2;`, nil})
	}
	// Alternate \d for psql 14
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 's' then 'special' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\n     left join pg_catalog.pg_am am on am.oid = c.relam\nwhere c.relkind in ('r','p','v','m','s','f','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, l.execute(conn, mysqlConn, ConvertedQuery{`SELECT 'public' AS 'Schema', TABLE_NAME AS 'Name', 'table' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = database() AND TABLE_TYPE = 'BASE TABLE' ORDER BY 2;`, nil})
	}
	// Command: \d table_name
	if strings.HasPrefix(statement, "select c.oid,\n  n.nspname,\n  c.relname\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\nwhere c.relname operator(pg_catalog.~) '^(") && strings.HasSuffix(statement, ")$' collate pg_catalog.default\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 2, 3;") {
		// There are >at least< 15 separate statements sent for this command, which is far too much to validate and
		// implement, so we'll just return an error for now
		return true, fmt.Errorf("PSQL command not yet supported")
	}
	// Command: \dn
	if statement == "select n.nspname as \"name\",\n  pg_catalog.pg_get_userbyid(n.nspowner) as \"owner\"\nfrom pg_catalog.pg_namespace n\nwhere n.nspname !~ '^pg_' and n.nspname <> 'information_schema'\norder by 1;" {
		return true, l.execute(conn, mysqlConn, ConvertedQuery{"SELECT 'public' AS 'Name', 'pg_database_owner' AS 'Owner';", nil})
	}
	// Command: \df
	if statement == "select n.nspname as \"schema\",\n  p.proname as \"name\",\n  pg_catalog.pg_get_function_result(p.oid) as \"result data type\",\n  pg_catalog.pg_get_function_arguments(p.oid) as \"argument data types\",\n case p.prokind\n  when 'a' then 'agg'\n  when 'w' then 'window'\n  when 'p' then 'proc'\n  else 'func'\n end as \"type\"\nfrom pg_catalog.pg_proc p\n     left join pg_catalog.pg_namespace n on n.oid = p.pronamespace\nwhere pg_catalog.pg_function_is_visible(p.oid)\n      and n.nspname <> 'pg_catalog'\n      and n.nspname <> 'information_schema'\norder by 1, 2, 4;" {
		return true, l.execute(conn, mysqlConn, ConvertedQuery{"SELECT '' AS 'Schema', '' AS 'Name', '' AS 'Result data type', '' AS 'Argument data types', '' AS 'Type' FROM dual LIMIT 0;", nil})
	}
	// Command: \dv
	if statement == "select n.nspname as \"schema\",\n  c.relname as \"name\",\n  case c.relkind when 'r' then 'table' when 'v' then 'view' when 'm' then 'materialized view' when 'i' then 'index' when 's' then 'sequence' when 't' then 'toast table' when 'f' then 'foreign table' when 'p' then 'partitioned table' when 'i' then 'partitioned index' end as \"type\",\n  pg_catalog.pg_get_userbyid(c.relowner) as \"owner\"\nfrom pg_catalog.pg_class c\n     left join pg_catalog.pg_namespace n on n.oid = c.relnamespace\nwhere c.relkind in ('v','')\n      and n.nspname <> 'pg_catalog'\n      and n.nspname !~ '^pg_toast'\n      and n.nspname <> 'information_schema'\n  and pg_catalog.pg_table_is_visible(c.oid)\norder by 1,2;" {
		return true, l.execute(conn, mysqlConn, ConvertedQuery{"SELECT 'public' AS 'Schema', TABLE_NAME AS 'Name', 'view' AS 'Type', 'postgres' AS 'Owner' FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = database() AND TABLE_TYPE = 'VIEW' ORDER BY 2;", nil})
	}
	// Command: \du
	if statement == "select r.rolname, r.rolsuper, r.rolinherit,\n  r.rolcreaterole, r.rolcreatedb, r.rolcanlogin,\n  r.rolconnlimit, r.rolvaliduntil,\n  array(select b.rolname\n        from pg_catalog.pg_auth_members m\n        join pg_catalog.pg_roles b on (m.roleid = b.oid)\n        where m.member = r.oid) as memberof\n, r.rolreplication\n, r.rolbypassrls\nfrom pg_catalog.pg_roles r\nwhere r.rolname !~ '^pg_'\norder by 1;" {
		// We don't support users yet, so we'll just return nothing for now
		return true, l.execute(conn, mysqlConn, ConvertedQuery{"SELECT '' FROM dual LIMIT 0;", nil})
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

// comQuery is a shortcut that determines which version of ComQuery to call based on whether the query has been parsed.
func (l *Listener) comQuery(mysqlConn *mysql.Conn, query ConvertedQuery, callback func(res *sqltypes.Result, more bool) error) error {
	if query.AST == nil {
		return l.cfg.Handler.ComQuery(mysqlConn, query.String, callback)
	} else {
		return l.cfg.Handler.ComParsedQuery(mysqlConn, query.String, query.AST, callback)
	}
}
