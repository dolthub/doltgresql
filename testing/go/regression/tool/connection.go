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

package main

import (
	"errors"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
)

// Connection contains the connection that is used by the replay system for interacting with the Doltgres server.
type Connection struct {
	frontend   *pgproto3.Frontend
	connection net.Conn
	timeout    time.Duration
	reader     *MessageReader
	startup    *pgproto3.StartupMessage
	bmChan     chan pgproto3.BackendMessage
	errChan    chan error
}

// NewConnection returns a new Connection.
func NewConnection(network string, reader *MessageReader, timeout time.Duration) (*Connection, error) {
	connection, err := (&net.Dialer{}).Dial("tcp", network)
	if err != nil {
		return nil, err
	}
	return &Connection{
		frontend:   pgproto3.NewFrontend(connection, connection),
		connection: connection,
		timeout:    timeout,
		reader:     reader,
		startup:    nil,
		bmChan:     make(chan pgproto3.BackendMessage),
		errChan:    make(chan error),
	}, nil
}

// Close closes the internal connection.
func (c *Connection) Close() {
	_ = c.connection.Close()
}

// EmptyReceiveBuffer empties the buffer used by Receive. Returns an error if the buffer was not empty.
func (c *Connection) EmptyReceiveBuffer() error {
	if c.frontend.ReadBufferLen() > 0 {
		for c.frontend.ReadBufferLen() > 0 {
			_, _ = c.frontend.Receive()
		}
		return errors.New("Doltgres sent additional messages after ReadyForQuery")
	}
	return nil
}

// Queue adds the given messages to the queue, which will be sent when Send is called. This is generally useful for when
// messages may conditionally be added, and we want to send all messages at once rather than one at a time.
func (c *Connection) Queue(messages ...pgproto3.FrontendMessage) {
	for _, message := range messages {
		if startupMessage, ok := message.(*pgproto3.StartupMessage); ok {
			c.startup = startupMessage
		}
		c.frontend.Send(message)
	}
}

// Receive returns the next message from the backend. If an error is returned, then the connection has been closed, and
// a new one should be opened.
func (c *Connection) Receive() (pgproto3.BackendMessage, error) {
	var message pgproto3.BackendMessage
	go func() {
		var err error
		message, err = c.frontend.Receive()
		c.errChan <- err
	}()
	return message, c.handleErrorChannel(false, false)
}

// Send sends the given messages (or messages added via Queue) over the wire. Will sync the reader to the next query on
// error. If such behavior is not desired, then use SendNoSync. If an error is returned, then the connection has been
// closed, and a new one should be opened.
func (c *Connection) Send(messages ...pgproto3.FrontendMessage) error {
	c.Queue(messages...)
	go func() {
		c.errChan <- c.frontend.Flush()
	}()
	return c.handleErrorChannel(true, true)
}

// SendNoSync is the same as Send, except that the reader will NOT sync to the next query if an error is encountered.
// This should only be called when we can guarantee that the reader is already synchronized to the next query. If an
// error is returned, then the connection has been closed, and a new one should be opened.
func (c *Connection) SendNoSync(messages ...pgproto3.FrontendMessage) error {
	c.Queue(messages...)
	go func() {
		c.errChan <- c.frontend.Flush()
	}()
	return c.handleErrorChannel(false, true)
}

// handleErrorChannel handles whether the Connection should close once an item has been sent to the error channel. If an
// error is returned, then the Connection has been closed.
func (c *Connection) handleErrorChannel(syncOnError bool, isSend bool) error {
	var err error
	select {
	case err = <-c.errChan:
	case <-time.After(c.timeout):
		if isSend {
			err = errors.New("timeout during Send")
		} else {
			err = errors.New("timeout during Receive")
		}
	}
	if err != nil {
		if syncOnError {
			c.reader.SyncToNextQuery()
		}
		c.reader.PushQueue(c.startup, &pgproto3.ReadyForQuery{TxStatus: 'I'})
		_ = c.connection.Close()
	}
	return err
}
