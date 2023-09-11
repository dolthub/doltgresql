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
	"sync/atomic"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/sqltypes"

	"github.com/dolthub/doltgresql/postgres/messages"
)

var connectionIDCounter uint32

// TODO: doc
type Listener struct {
	listener net.Listener
	cfg      mysql.ListenerConfig
}

var _ server.ProtocolListener = (*Listener)(nil)

// TODO: doc
func NewListenerWithConfig(listenerCfg mysql.ListenerConfig) (server.ProtocolListener, error) {
	return &Listener{
		listener: listenerCfg.Listener,
		cfg:      listenerCfg,
	}, nil
}

// TODO: doc
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

// TODO: doc
func (l *Listener) Close() {
	_ = l.listener.Close()
}

// TODO: doc
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

	defer func() {
		l.cfg.Handler.ConnectionClosed(mysqlConn)
		if err := conn.Close(); err != nil {
			fmt.Printf("Failed to properly close connection:\n%v\n", err)
		}
	}()
	l.cfg.Handler.NewConnection(mysqlConn)

	buf := make([]byte, 2048)
	_, err := conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Println(err)
		}
		return
	}

	if _, err = conn.Write(messages.SSLResponse{
		SupportsSSL: false,
	}.Bytes()); err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Println(err)
		}
		return
	}
	startupMessage, err := messages.ReadStartupMessage(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err = conn.Write(messages.AuthenticationOk{}.Bytes()); err != nil {
		fmt.Println(err)
		return
	}

	if _, err = conn.Write(messages.ParameterStatus{
		Name:  "server_version",
		Value: "15.0",
	}.Bytes()); err != nil {
		fmt.Println(err)
		return
	}
	if _, err = conn.Write(messages.ParameterStatus{
		Name:  "client_encoding",
		Value: "UTF8",
	}.Bytes()); err != nil {
		fmt.Println(err)
		return
	}

	if _, err = conn.Write(messages.BackendKeyData{
		ProcessID: 1,
		SecretKey: 0,
	}.Bytes()); err != nil {
		fmt.Println(err)
		return
	}

	if _, err = conn.Write(messages.ReadyForQuery{
		Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
	}.Bytes()); err != nil {
		fmt.Println(err)
		return
	}

	if db, ok := startupMessage.Parameters["database"]; ok && len(db) > 0 {
		l.cfg.Handler.ComQuery(mysqlConn, fmt.Sprintf("USE `%s`;", db), func(res *sqltypes.Result, more bool) error {
			return nil
		})
	}

	for {
		if _, err = conn.Read(buf); err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}

		if messages.ReadTerminate(buf) {
			return
		}
		query, ok := messages.ReadQuery(buf)
		if !ok {
			fmt.Println("unknown message, terminating connection")
			return
		}
		commandCompleteTag, err := messages.QueryToCommandCompleteTag(query)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		var rowTotal int32
		if err = l.cfg.Handler.ComQuery(mysqlConn, query, func(res *sqltypes.Result, more bool) error {
			rowDescription, err := messages.NewRowDescription(res.Fields)
			if err != nil {
				return err
			}
			if _, err = conn.Write(rowDescription.Bytes()); err != nil {
				return err
			}

			for _, row := range res.Rows {
				if _, err = conn.Write(messages.NewDataRow(row).Bytes()); err != nil {
					return err
				}
			}

			if commandCompleteTag == messages.CommandCompleteTag_INSERT ||
				commandCompleteTag == messages.CommandCompleteTag_UPDATE ||
				commandCompleteTag == messages.CommandCompleteTag_DELETE {
				rowTotal = int32(res.RowsAffected)
			} else {
				rowTotal += int32(len(res.Rows))
			}
			return nil
		}); err != nil {
			fmt.Println(err.Error())
			return
		}

		if _, err = conn.Write(messages.CommandComplete{
			Tag:  commandCompleteTag,
			Rows: rowTotal,
		}.Bytes()); err != nil {
			fmt.Println(err)
			return
		}

		if _, err = conn.Write(messages.ReadyForQuery{
			Indicator: messages.ReadyForQueryTransactionIndicator_Idle,
		}.Bytes()); err != nil {
			fmt.Println(err)
			return
		}
	}
}
