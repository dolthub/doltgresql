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

// connectionModeKind identifies the mutually exclusive frontend protocol mode owning the connection.
type connectionModeKind byte

const (
	invalidConnectionMode connectionModeKind = iota
	readyConnectionMode
	extendedConnectionMode
	discardUntilSyncConnectionMode
	copyInConnectionMode
	closingConnectionMode
)

// connectionMode records the active protocol mode and its state-specific COPY payload.
type connectionMode struct {
	kind connectionModeKind
	copy *copyInState
}

// valid reports whether the state kind and optional COPY payload agree.
func (s connectionMode) valid() bool {
	switch s.kind {
	case copyInConnectionMode:
		return s.copy != nil
	case readyConnectionMode, extendedConnectionMode, discardUntilSyncConnectionMode, closingConnectionMode:
		return s.copy == nil
	default:
		return false
	}
}

// connectionState owns every transaction and protocol phase transition for a connection.
type connectionState struct {
	transaction transactionState
	protocol    connectionMode
}

// newConnectionState returns the initial state for a newly authenticated connection.
func newConnectionState() connectionState {
	return connectionState{
		transaction: idleTransactionState,
		protocol:    connectionMode{kind: readyConnectionMode},
	}
}

// beginExtended enters an extended-query batch unless another exclusive operation owns the connection.
func (s *connectionState) beginExtended() {
	if s.protocol.kind == readyConnectionMode {
		s.protocol = connectionMode{kind: extendedConnectionMode}
	}
}

// finishExtended returns the connection to normal command dispatch.
func (s *connectionState) finishExtended() {
	s.protocol = connectionMode{kind: readyConnectionMode}
}

// discardUntilSync rejects the remainder of an extended-query batch.
func (s *connectionState) discardUntilSync() {
	s.protocol = connectionMode{kind: discardUntilSyncConnectionMode}
}

// closeProtocol transitions the connection into its terminal protocol mode.
func (s *connectionState) closeProtocol() {
	s.protocol = connectionMode{kind: closingConnectionMode}
}

// beginCopy transfers exclusive protocol ownership to a valid COPY FROM STDIN operation.
func (s *connectionState) beginCopy(copyState *copyInState) bool {
	if copyState == nil {
		return false
	}
	switch s.protocol.kind {
	case readyConnectionMode, extendedConnectionMode:
		s.protocol = connectionMode{kind: copyInConnectionMode, copy: copyState}
		return true
	default:
		return false
	}
}

// finishCopy releases COPY ownership only when copyState is the active operation.
func (s *connectionState) finishCopy(copyState *copyInState) bool {
	if s.protocol.kind != copyInConnectionMode || s.protocol.copy != copyState {
		return false
	}
	s.protocol = connectionMode{kind: readyConnectionMode}
	return true
}
