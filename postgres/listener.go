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

package postgres

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/sqltypes"

	"github.com/dolthub/doltgresql/postgres/connection"
	"github.com/dolthub/doltgresql/postgres/messages"
)

var connectionIDCounter uint32

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
		if returnErr != nil {
			//TODO: return errors to the client
			fmt.Println(returnErr.Error())
		}
		l.cfg.Handler.ConnectionClosed(mysqlConn)
		if err := conn.Close(); err != nil {
			fmt.Printf("Failed to properly close connection:\n%v\n", err)
		}
	}()
	l.cfg.Handler.NewConnection(mysqlConn)

	var startupMessage messages.StartupMessage
	// The initial message may be one of a few different messages, so we'll check for those.
InitialMessageLoop:
	for {
		initialMessages, err := connection.ReceiveIntoAny(conn,
			messages.StartupMessage{},
			messages.SSLRequest{},
			messages.GSSENCRequest{})
		if err != nil {
			if err != io.EOF {
				returnErr = err
			}
			return
		}
		if len(initialMessages) != 1 {
			returnErr = fmt.Errorf("Expected a single message upon starting connection, terminating connection")
			return
		}
		initialMessage := initialMessages[0]

		switch initialMessage := initialMessage.(type) {
		case messages.StartupMessage:
			startupMessage = initialMessage
			break InitialMessageLoop
		case messages.SSLRequest:
			if err = connection.Send(conn, messages.SSLResponse{
				SupportsSSL: false,
			}); err != nil {
				returnErr = err
				return
			}
		case messages.GSSENCRequest:
			if err = connection.Send(conn, messages.GSSENCResponse{
				SupportsGSSAPI: false,
			}); err != nil {
				returnErr = err
				return
			}
		default:
			returnErr = fmt.Errorf("Unexpected initial message, terminating connection")
			return
		}
	}

	if err := connection.Send(conn, messages.AuthenticationOk{}); err != nil {
		returnErr = err
		return
	}

	if err := connection.Send(conn, messages.ParameterStatus{
		Name:  "server_version",
		Value: "15.0",
	}); err != nil {
		returnErr = err
		return
	}
	if err := connection.Send(conn, messages.ParameterStatus{
		Name:  "client_encoding",
		Value: "UTF8",
	}); err != nil {
		returnErr = err
		return
	}

	if err := connection.Send(conn, messages.BackendKeyData{
		ProcessID: 1,
		SecretKey: 0,
	}); err != nil {
		returnErr = err
		return
	}

	if err := connection.Send(conn, messages.ReadyForQuery{
		Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
	}); err != nil {
		returnErr = err
		return
	}

	if db, ok := startupMessage.Parameters["database"]; ok && len(db) > 0 {
		l.cfg.Handler.ComQuery(mysqlConn, fmt.Sprintf("USE `%s`;", db), func(res *sqltypes.Result, more bool) error {
			return nil
		})
	}

	statementCache := make(map[string]string)
	for {
		receivedMessages, err := connection.Receive(conn)
		if err != nil {
			returnErr = err
			return
		} else if len(receivedMessages) == 0 {
			returnErr = fmt.Errorf("Data received but contained no messages, terminating connection")
			return
		}

		portals := make(map[string]string)
		for _, message := range receivedMessages {
			switch message := message.(type) {
			case messages.Terminate:
				return
			case messages.Execute:
				//TODO: implement the RowMax
				if err = l.execute(conn, mysqlConn, portals[message.Portal]); err != nil {
					fmt.Println(err)
					return
				}
			case messages.Query:
				if err = l.execute(conn, mysqlConn, message.String); err != nil {
					fmt.Println(err)
					return
				}
				if err := connection.Send(conn, messages.ReadyForQuery{
					Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
				}); err != nil {
					fmt.Println(err)
					return
				}
			case messages.Parse:
				//TODO: fully support prepared statements
				statementCache[message.Name] = message.Query
				if err = connection.Send(conn, messages.ParseComplete{}); err != nil {
					fmt.Println(err)
					return
				}
			case messages.Describe:
				var query string
				if message.IsPrepared {
					query = statementCache[message.Target]
				} else {
					query = portals[message.Target]
				}
				if err = l.describe(conn, mysqlConn, message, query); err != nil {
					fmt.Println(err)
					return
				}
			case messages.Sync:
				if err = connection.Send(conn, messages.ReadyForQuery{
					Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
				}); err != nil {
					fmt.Println(err)
					return
				}
			case messages.Bind:
				//TODO: fully support prepared statements
				portals[message.DestinationPortal] = statementCache[message.SourcePreparedStatement]
				if err = connection.Send(conn, messages.BindComplete{}); err != nil {
					fmt.Println(err)
					return
				}
			case nil:
				returnErr = fmt.Errorf("Unknown message format, terminating connection")
				return
			default:
				returnErr = fmt.Errorf(`Unexpected message "%s", terminating connection\n`, message.DefaultMessage().Name)
				return
			}
		}
	}
}

// execute handles running the given query. This will post the RowDescription, DataRow, and CommandComplete messages.
func (l *Listener) execute(conn net.Conn, mysqlConn *mysql.Conn, query string) error {
	commandComplete := messages.CommandComplete{
		Query: query,
		Rows:  0,
	}

	if err := l.cfg.Handler.ComQuery(mysqlConn, query, func(res *sqltypes.Result, more bool) error {
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
		return err
	}

	if err := connection.Send(conn, commandComplete); err != nil {
		return err
	}

	return nil
}

// describe handles the description of the given query. This will post the ParameterDescription and RowDescription messages.
func (l *Listener) describe(conn net.Conn, mysqlConn *mysql.Conn, message messages.Describe, statement string) error {
	//TODO: fully support prepared statements
	if err := connection.Send(conn, messages.ParameterDescription{
		ObjectIDs: nil,
	}); err != nil {
		return err
	}

	if strings.HasPrefix(strings.TrimSpace(strings.ToLower(statement)), "select") {
		// Since it is a SELECT statement, we can run it multiple times without worry.
		if err := l.cfg.Handler.ComQuery(mysqlConn, statement, func(res *sqltypes.Result, more bool) error {
			if err := connection.Send(conn, messages.RowDescription{
				Fields: res.Fields,
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("We do not yet support returning rows from the given statement")
	}

	return nil
}
