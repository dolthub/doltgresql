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
	"errors"
	"net"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/gcctx"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	sqle "github.com/dolthub/go-mysql-server"
	gmsserver "github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/dolthub/doltgresql/core"
)

// No database, socket, timers or engine is needed to prove whether the real
// PostgreSQL entry points register their session with Dolt's actual GC owner.
type lifecycleProvider struct{ dsess.DoltDatabaseProvider }

// FileSystem implements dsess.DoltDatabaseProvider without opening a database.
func (lifecycleProvider) FileSystem() filesys.Filesys { return filesys.LocalFS }

// lifecycleFixture binds the actual session manager and GC owner without SQL I/O.
func lifecycleFixture(t *testing.T, authenticated bool) (*DoltgresHandler, *mysql.Conn, *dsess.DoltSession, *gcctx.GCSafepointController) {
	t.Helper()
	controller := gcctx.NewGCSafepointController()
	session, err := dsess.NewDoltSession(sql.NewBaseSession(), lifecycleProvider{}, config.NewMapConfig(nil), nil, nil, nil, controller, nil)
	require.NoError(t, err)
	if authenticated {
		require.NoError(t, core.InitializeIdentityOnSession(session, 1, true))
	}
	manager := gmsserver.NewSessionManager(sql.NewContext, func(context.Context, *mysql.Conn, string) (sql.Session, error) { return session, nil }, noop.NewTracerProvider().Tracer("lifecycle-test"), nil, nil, sql.EmptyProcessList{}, "")
	conn := &mysql.Conn{ConnectionID: 1}
	require.NoError(t, manager.NewSession(t.Context(), conn))
	t.Cleanup(func() { sql.SessionEnd(session) })
	return &DoltgresHandler{sm: manager}, conn, session, controller
}

// requireRegisteredIdle verifies completed commands leave their session visible.
func requireRegisteredIdle(t *testing.T, controller *gcctx.GCSafepointController, session *dsess.DoltSession) {
	t.Helper()
	visits := 0
	waiter := controller.Waiter(t.Context(), nil, func(_ context.Context, provider gcctx.GCRootsProvider) error {
		require.Same(t, session, provider)
		visits++
		return nil
	})
	require.NoError(t, waiter.Wait(t.Context()))
	require.Equal(t, 1, visits, "the PG command must leave its idle transaction session visible to GC")
}

// TestPGCommandFailuresRegisterSession covers early errors in actual entry points.
func TestPGCommandFailuresRegisterSession(t *testing.T) {
	for _, operation := range []string{"query", "prepare", "bind", "bound-execute"} {
		t.Run(operation, func(t *testing.T) {
			h, conn, session, controller := lifecycleFixture(t, false)
			var err error
			switch operation {
			case "query":
				err = h.ComQuery(t.Context(), conn, "SELECT 1", nil, nil)
			case "prepare":
				_, _, err = h.ComPrepareParsed(t.Context(), conn, "SELECT 1", nil)
			case "bind":
				_, _, err = h.ComBind(t.Context(), conn, "SELECT 1", nil, BindVariables{}, nil)
			case "bound-execute":
				err = h.ComExecuteBound(t.Context(), conn, "SELECT 1", nil, nil, nil)
			}
			require.Error(t, err)
			requireRegisteredIdle(t, controller, session)
		})
	}
}

// TestPGQueryCommandProtectsExecutorAndSpooling checks both root-access phases.
func TestPGQueryCommandProtectsExecutorAndSpooling(t *testing.T) {
	h, conn, session, controller := lifecycleFixture(t, true)
	calls := []string{}
	err := h.doQuery(t.Context(), conn, "SELECT 1", nil, nil, func(ctx *sql.Context, _ string, _ sqlparser.Statement, _ sql.Node) (sql.Schema, sql.RowIter, *sql.QueryFlags, error) {
		controller.Validate(session)
		calls = append(calls, "executor")
		return nil, sql.RowsToRowIter(), nil, nil
	}, func(ctx *sql.Context, result *Result) error {
		controller.Validate(session)
		calls = append(calls, "spooled")
		return nil
	}, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"executor", "spooled"}, calls)
	requireRegisteredIdle(t, controller, session)
}

// TestPGCommandNestedHandlerKeepsOuterOwnership guards protocol error cleanup.
func TestPGCommandNestedHandlerKeepsOuterOwnership(t *testing.T) {
	h, conn, session, controller := lifecycleFixture(t, false)
	end, err := h.beginCommand(conn)
	require.NoError(t, err)
	defer end()
	// The actual nested entry point fails before execution, but must not end
	// the protocol message's command before its error handling has finished.
	require.Error(t, h.ComQuery(t.Context(), conn, "SELECT 1", nil, nil))
	controller.Validate(session)
	end()
	end() // The protocol's deferred release is safe after explicit cleanup.
	requireRegisteredIdle(t, controller, session)
}

// TestPGQueryCommandErrorsEndOwnership checks errors in each execution phase.
func TestPGQueryCommandErrorsEndOwnership(t *testing.T) {
	for _, failure := range []string{"executor", "spooling"} {
		t.Run(failure, func(t *testing.T) {
			h, conn, session, controller := lifecycleFixture(t, true)
			want := errors.New(failure)
			err := h.doQuery(t.Context(), conn, "SELECT 1", nil, nil, func(*sql.Context, string, sqlparser.Statement, sql.Node) (sql.Schema, sql.RowIter, *sql.QueryFlags, error) {
				controller.Validate(session)
				if failure == "executor" {
					return nil, nil, nil, want
				}
				return nil, sql.RowsToRowIter(), nil, nil
			}, func(*sql.Context, *Result) error {
				controller.Validate(session)
				return want
			}, nil)
			require.ErrorIs(t, err, want)
			require.False(t, h.command.owned)
			requireRegisteredIdle(t, controller, session)
		})
	}
}

type countingCommandSession struct {
	sql.Session
	begins, ends int
	active       bool
	beginErr     error
}

// CommandBegin implements sql.LifecycleAwareSession and rejects nested commands.
func (s *countingCommandSession) CommandBegin() error {
	if s.active {
		panic("nested command begin")
	}
	s.begins++
	if s.beginErr != nil {
		return s.beginErr
	}
	s.active = true
	return nil
}

// CommandEnd implements sql.LifecycleAwareSession and rejects unmatched ends.
func (s *countingCommandSession) CommandEnd() {
	if !s.active {
		panic("unpaired command end")
	}
	s.ends++
	s.active = false
}

// SessionEnd implements sql.LifecycleAwareSession.
func (*countingCommandSession) SessionEnd() {}

// CommandSafepoint implements sql.SafepointAwareSession.
func (*countingCommandSession) CommandSafepoint() {}

// TestPGCommandFailedBeginHasNoEnd checks that failed acquisition owns nothing.
func TestPGCommandFailedBeginHasNoEnd(t *testing.T) {
	want := errors.New("begin refused")
	session := &countingCommandSession{beginErr: want}
	var command pgCommandLifecycle
	end, err := command.begin(session)
	require.ErrorIs(t, err, want)
	require.Nil(t, end)
	require.False(t, command.owned)
	require.Nil(t, command.session)
	require.Equal(t, 0, session.ends)
}

// TestPGCommandResetHandsOffReplacement checks the reset and SetDB handoff.
func TestPGCommandResetHandsOffReplacement(t *testing.T) {
	old := &countingCommandSession{}
	next := &countingCommandSession{}
	var command pgCommandLifecycle
	end, err := command.begin(old)
	require.NoError(t, err)
	// Reset disposes the old session; SetDB owns its own replacement command.
	command.releaseSession()
	require.NoError(t, sql.SessionCommandBegin(next))
	sql.SessionCommandEnd(next)
	// Error/ReadyForQuery handling may access the replacement afterwards.
	borrowed, err := command.begin(next)
	require.NoError(t, err)
	borrowed()
	require.True(t, next.active)
	end()
	require.Equal(t, 1, old.begins)
	require.Equal(t, 1, old.ends)
	require.Equal(t, 2, next.begins)
	require.Equal(t, 2, next.ends)
	require.False(t, command.owned)
}

type cleanupDatabaseProvider struct {
	sql.DatabaseProvider
	check func()
}

// Database checks protection when the real catalog resolves a locked table.
func (p cleanupDatabaseProvider) Database(*sql.Context, string) (sql.Database, error) {
	p.check()
	return nil, errors.New("synthetic table no longer exists")
}

// TestPGCommandConnectionCloseProtectsTableUnlock checks cleanup before removal.
func TestPGCommandConnectionCloseProtectsTableUnlock(t *testing.T) {
	h, conn, session, controller := lifecycleFixture(t, true)
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	conn.Conn = server
	h.sm.AddConn(conn)
	calls := 0
	h.e = sqle.NewDefault(cleanupDatabaseProvider{check: func() {
		controller.Validate(session)
		calls++
	}})
	t.Cleanup(func() { require.NoError(t, h.e.Close()) })
	catalog := h.e.Analyzer.Catalog
	ctx := sql.NewContext(t.Context(), sql.WithSession(session))
	catalog.LockTable(ctx, "locked_table")
	h.ConnectionClosed(conn)
	require.Equal(t, 1, calls)
	require.False(t, h.command.owned)
	require.Nil(t, h.sm.GetSession(conn))
	// SessionEnd must run after CommandEnd, removing the session from GC.
	visits := 0
	require.NoError(t, controller.Waiter(t.Context(), nil, func(context.Context, gcctx.GCRootsProvider) error {
		visits++
		return nil
	}).Wait(t.Context()))
	require.Zero(t, visits)
}

// TestPGCommandUnauthenticatedCloseCreatesNoSession covers startup failure.
func TestPGCommandUnauthenticatedCloseCreatesNoSession(t *testing.T) {
	built := 0
	manager := gmsserver.NewSessionManager(sql.NewContext, func(context.Context, *mysql.Conn, string) (sql.Session, error) {
		built++
		return nil, errors.New("no authenticated session")
	}, noop.NewTracerProvider().Tracer("lifecycle-test"), nil, nil, sql.EmptyProcessList{}, "")
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	conn := &mysql.Conn{ConnectionID: 1, Conn: server}
	manager.AddConn(conn)
	engine := sqle.NewDefault(cleanupDatabaseProvider{})
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	h := &DoltgresHandler{sm: manager, e: engine}
	h.ConnectionClosed(conn)
	require.Zero(t, built, "failed authentication must not create a session during cleanup")
	require.False(t, h.command.owned)
}
