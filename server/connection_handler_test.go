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
	"net"
	"testing"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
)

// TestCastSQLErrorIntegerOutOfRange verifies shared-engine integer overflow uses PostgreSQL's data-exception code.
func TestCastSQLErrorIntegerOutOfRange(t *testing.T) {
	err := sql.ErrIntegerOutOfRange.New("BIGINT", "(value + 1)")
	pgErr := castSQLError(err)
	require.Equal(t, pgcode.NumericValueOutOfRange.String(), pgErr.Code)
}

// TestConnectionStateProtocolTransitions verifies that exclusive protocol phases retain only their own data.
func TestConnectionStateProtocolTransitions(t *testing.T) {
	state := newConnectionState()
	require.Equal(t, readyProtocolState, state.protocol.kind)

	state.beginExtended()
	require.Equal(t, extendedQueryProtocolState, state.protocol.kind)
	state.discardUntilSync()
	require.Equal(t, discardUntilSyncProtocolState, state.protocol.kind)
	state.finishExtended()
	require.Equal(t, readyProtocolState, state.protocol.kind)

	execution := &simpleQueryExecution{nextStatement: 1}
	copyState := newCopyInState(nil, newSimpleQueryCopyContinuation(execution))
	require.True(t, state.beginCopy(copyState))
	require.Equal(t, copyInProtocolState, state.protocol.kind)
	require.Same(t, copyState, state.protocol.copy)
	continuation := copyState.continuation
	require.True(t, state.finishCopy(copyState))
	require.Equal(t, newSimpleQueryCopyContinuation(execution), continuation)
	require.Equal(t, readyProtocolState, state.protocol.kind)
	state.closeProtocol()
	require.Equal(t, closingProtocolState, state.protocol.kind)
}

// TestProtocolStateValidation verifies protocol-state tags accept only their corresponding payload shape.
func TestProtocolStateValidation(t *testing.T) {
	copyState := newCopyInState(nil, newExtendedQueryCopyContinuation())
	tests := []struct {
		name  string
		state protocolState
		valid bool
	}{
		{name: "zero value"},
		{name: "unknown kind", state: protocolState{kind: protocolStateKind(255)}},
		{name: "ready", state: protocolState{kind: readyProtocolState}, valid: true},
		{name: "ready with copy", state: protocolState{kind: readyProtocolState, copy: copyState}},
		{name: "extended", state: protocolState{kind: extendedQueryProtocolState}, valid: true},
		{name: "discard", state: protocolState{kind: discardUntilSyncProtocolState}, valid: true},
		{name: "closing", state: protocolState{kind: closingProtocolState}, valid: true},
		{name: "copy without state", state: protocolState{kind: copyInProtocolState}},
		{name: "copy", state: protocolState{kind: copyInProtocolState, copy: copyState}, valid: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.valid, test.state.valid())
		})
	}
}

// TestConnectionStateCopyIdentity verifies COPY transitions reject nil and stale operation state.
func TestConnectionStateCopyIdentity(t *testing.T) {
	state := newConnectionState()
	require.False(t, state.beginCopy(nil))
	require.Equal(t, readyProtocolState, state.protocol.kind)

	active := newCopyInState(nil, newExtendedQueryCopyContinuation())
	stale := newCopyInState(nil, newExtendedQueryCopyContinuation())
	require.True(t, state.beginCopy(active))
	require.False(t, state.beginCopy(stale))
	require.Same(t, active, state.protocol.copy)
	require.False(t, state.finishCopy(stale))
	require.Same(t, active, state.protocol.copy)
	require.True(t, state.finishCopy(active))
	require.Equal(t, readyProtocolState, state.protocol.kind)
	state.beginExtended()
	require.True(t, state.beginCopy(active))
	require.True(t, state.finishCopy(active))

	state.discardUntilSync()
	require.False(t, state.beginCopy(active))
	require.Equal(t, discardUntilSyncProtocolState, state.protocol.kind)
	state.closeProtocol()
	require.False(t, state.beginCopy(active))
	require.Equal(t, closingProtocolState, state.protocol.kind)
}

// TestCopyContinuationValidation verifies that only constructed, internally consistent continuations are valid.
func TestCopyContinuationValidation(t *testing.T) {
	execution := &simpleQueryExecution{nextStatement: 1}
	tests := []struct {
		name         string
		continuation copyContinuation
		valid        bool
	}{
		{name: "zero value", continuation: copyContinuation{}},
		{name: "unknown kind", continuation: copyContinuation{kind: copyContinuationKind(255)}},
		{name: "simple without execution", continuation: copyContinuation{kind: simpleQueryCopyContinuation}},
		{name: "simple constructor without execution", continuation: newSimpleQueryCopyContinuation(nil)},
		{name: "simple", continuation: newSimpleQueryCopyContinuation(execution), valid: true},
		{name: "extended", continuation: newExtendedQueryCopyContinuation(), valid: true},
		{name: "extended with execution", continuation: copyContinuation{kind: extendedQueryCopyContinuation, execution: execution}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.valid, test.continuation.valid())
		})
	}
}

// TestFinishCopyRejectsInvalidContinuations verifies invalid COPY state cannot silently resume simple-query handling.
func TestFinishCopyRejectsInvalidContinuations(t *testing.T) {
	tests := []struct {
		name         string
		continuation copyContinuation
	}{
		{name: "zero value", continuation: copyContinuation{}},
		{name: "unknown kind", continuation: copyContinuation{kind: copyContinuationKind(255)}},
		{name: "simple without execution", continuation: copyContinuation{kind: simpleQueryCopyContinuation}},
		{name: "nil simple constructor", continuation: newSimpleQueryCopyContinuation(nil)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := &ConnectionHandler{state: newConnectionState()}
			copyState := newCopyInState(nil, test.continuation)
			require.True(t, handler.state.beginCopy(copyState))
			result := handler.finishCopy(copyState, nil)
			require.ErrorContains(t, result.err, "invalid protocol continuation")
			require.Equal(t, closeConnection, result.action)
			require.Equal(t, closingProtocolState, handler.state.protocol.kind)
			require.Equal(t, closeConnection, handler.handleMessage(nil).action)
		})
	}
}

// TestReceiveMessagePanicPreservesTerminalClose verifies panic recovery propagates an existing terminal mode.
func TestReceiveMessagePanicPreservesTerminalClose(t *testing.T) {
	conn := panicReadConn{}
	handler := &ConnectionHandler{
		backend: pgproto3.NewBackend(conn, conn),
		state:   newConnectionState(),
	}
	handler.state.closeProtocol()

	stop, err := handler.receiveMessage()
	require.NoError(t, err)
	require.True(t, stop)
	require.Equal(t, closingProtocolState, handler.state.protocol.kind)
	require.Equal(t, closeConnection, handler.handleMessage(&pgproto3.Query{String: "SELECT 1"}).action)
}

// TestTransactionReadyStatus verifies that transaction state has one authoritative wire-status mapping.
func TestTransactionReadyStatus(t *testing.T) {
	tests := []struct {
		state    transactionState
		expected byte
	}{
		{state: idleTransactionState, expected: 'I'},
		{state: implicitTransactionState, expected: 'T'},
		{state: explicitTransactionState, expected: 'T'},
		{state: failedTransactionState, expected: 'E'},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, test.state.readyStatus())
	}
}

// TestExtendedQueryStateOwnsObjectLifetimes verifies statement and portal cleanup stays coordinated.
func TestExtendedQueryStateOwnsObjectLifetimes(t *testing.T) {
	state := newExtendedQueryState()
	state.preparedStatements[""] = preparedStatementData{}
	state.preparedStatements["statement"] = preparedStatementData{}
	state.portals[""] = portalData{}
	state.portals["portal"] = portalData{}

	state.clearUnnamed()
	require.NotContains(t, state.preparedStatements, "")
	require.NotContains(t, state.portals, "")
	require.Contains(t, state.preparedStatements, "statement")
	require.Contains(t, state.portals, "portal")

	state.close('S', "statement")
	state.close('P', "portal")
	require.Empty(t, state.preparedStatements)
	require.Empty(t, state.portals)

	require.ErrorContains(t, state.deallocate("missing"), "prepared statement missing does not exist")
	state.preparedStatements["one"] = preparedStatementData{}
	state.preparedStatements["two"] = preparedStatementData{}
	require.NoError(t, state.deallocate(""))
	require.Empty(t, state.preparedStatements)
}

// panicReadConn panics on reads while accepting response writes for panic-recovery tests.
type panicReadConn struct{}

// Read implements net.Conn.
func (panicReadConn) Read([]byte) (int, error) { panic("test read panic") }

// Write implements net.Conn.
func (panicReadConn) Write(p []byte) (int, error) { return len(p), nil }

// Close implements net.Conn.
func (panicReadConn) Close() error { return nil }

// LocalAddr implements net.Conn.
func (panicReadConn) LocalAddr() net.Addr { return nil }

// RemoteAddr implements net.Conn.
func (panicReadConn) RemoteAddr() net.Addr { return nil }

// SetDeadline implements net.Conn.
func (panicReadConn) SetDeadline(time.Time) error { return nil }

// SetReadDeadline implements net.Conn.
func (panicReadConn) SetReadDeadline(time.Time) error { return nil }

// SetWriteDeadline implements net.Conn.
func (panicReadConn) SetWriteDeadline(time.Time) error { return nil }
