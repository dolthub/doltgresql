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
	"net"
	"sync"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/vitess/go/mysql"
)

// LimitedListener automatically shuts down the listener after two connections close. This is intended only for testing.
// Tests will always make use of two connections, and they are as follows:
// 1. Establish a non-SSL connection, and create the database. Disconnect afterward (Postgres enforces one connection per database).
// 2. Establish a non-SSL connection, and run all of the queries relevant to the database.
// Tests that do not use the framework provided may not adhere to the above rules, and therefore must find another way
// to properly close their servers.
type LimitedListener struct {
	listener *Listener
	count    int
}

var _ server.ProtocolListener = (*LimitedListener)(nil)

// NewLimitedListener creates a new LimitedListener.
func NewLimitedListener(listenerCfg mysql.ListenerConfig) (server.ProtocolListener, error) {
	listener, err := NewListener(listenerCfg)
	if err != nil {
		return nil, err
	}
	return &LimitedListener{
		listener: listener.(*Listener),
		count:    0,
	}, nil
}

// Accept implements the interface server.ProtocolListener.
func (l *LimitedListener) Accept() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	for {
		if l.count >= 2 {
			break
		}
		conn, err := l.listener.listener.Accept()
		if err != nil {
			panic(err)
		}
		l.count++

		go func() {
			l.listener.HandleConnection(conn)
			wg.Done()
		}()
	}
	wg.Wait()
	l.Close()
}

// Close implements the interface server.ProtocolListener.
func (l *LimitedListener) Close() {
	l.listener.Close()
}

// Addr implements the interface server.ProtocolListener.
func (l *LimitedListener) Addr() net.Addr {
	return l.listener.Addr()
}
