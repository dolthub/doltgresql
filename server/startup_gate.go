// Copyright 2026 Dolthub, Inc.
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
	"net"
	"sync"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/vitess/go/mysql"
)

// startupGate delays acceptance of external client connections until first-run
// initialization (creation of the default database) is complete. Before the
// gate is released, only internal connections created through Dial are served.
type startupGate struct {
	conns    chan net.Conn
	released chan struct{}
	once     sync.Once
}

func newStartupGate() *startupGate {
	return &startupGate{
		conns:    make(chan net.Conn),
		released: make(chan struct{}),
	}
}

// Dial returns the client half of an in-memory connection. The server half is
// handed to the gated listener's accept loop and served through the standard
// connection-handling stack, exactly as a network connection would be.
func (g *startupGate) Dial(ctx context.Context) (net.Conn, error) {
	client, srvr := net.Pipe()
	select {
	case g.conns <- srvr:
		return client, nil
	case <-ctx.Done():
		_ = client.Close()
		_ = srvr.Close()
		return nil, ctx.Err()
	}
}

// Release opens the gate: listeners gated on this startupGate begin accepting
// external connections. Safe to call more than once.
func (g *startupGate) Release() {
	g.once.Do(func() {
		close(g.released)
	})
}

// gatedListenerFactory wraps |inner| so that the listeners it produces serve
// only the gate's internal connections until |gate| is released, after which
// they delegate to the wrapped listener's normal accept loop.
func (g *startupGate) gatedListenerFactory(inner server.ProtocolListenerFunc) server.ProtocolListenerFunc {
	return func(cfg server.Config, listenerCfg mysql.ListenerConfig, sel server.ServerEventListener) (server.ProtocolListener, error) {
		pl, err := inner(cfg, listenerCfg, sel)
		if err != nil {
			return nil, err
		}
		return &gatedProtocolListener{
			inner: pl,
			gate:  g,
			cfg:   listenerCfg,
			sel:   sel,
			quit:  make(chan struct{}),
		}, nil
	}
}

// gatedProtocolListener is a server.ProtocolListener that serves a
// startupGate's internal connections until the gate is released, then hands
// control to the wrapped listener.
type gatedProtocolListener struct {
	inner     server.ProtocolListener
	gate      *startupGate
	cfg       mysql.ListenerConfig
	sel       server.ServerEventListener
	quit      chan struct{}
	closeOnce sync.Once
}

var _ server.ProtocolListener = (*gatedProtocolListener)(nil)

// Accept implements server.ProtocolListener.
func (l *gatedProtocolListener) Accept() {
	for {
		select {
		case conn := <-l.gate.conns:
			connectionHandler := NewConnectionHandler(conn, l.cfg.Handler, l.sel)
			go connectionHandler.HandleConnection()
		case <-l.gate.released:
			l.inner.Accept()
			return
		case <-l.quit:
			return
		}
	}
}

// Close implements server.ProtocolListener.
func (l *gatedProtocolListener) Close() {
	l.closeOnce.Do(func() {
		close(l.quit)
	})
	l.inner.Close()
}

// Addr implements server.ProtocolListener.
func (l *gatedProtocolListener) Addr() net.Addr {
	return l.inner.Addr()
}
