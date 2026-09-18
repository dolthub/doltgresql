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

// connectionModeKind identifies the mutually exclusive frontend mode owning the connection.
type connectionModeKind byte

const (
	readyConnectionMode connectionModeKind = iota
	extendedQueryConnectionMode
	discardUntilSyncConnectionMode
	copyInConnectionMode
	closingConnectionMode
)

// connectionMode records the active frontend mode and its mode-specific COPY payload.
type connectionMode struct {
	kind connectionModeKind
	copy *copyInState
}

// connectionState owns every transaction and frontend-mode transition for a connection.
type connectionState struct {
	transaction transactionState
	mode        connectionMode
}

// newConnectionState returns the initial state for a newly authenticated connection.
func newConnectionState() connectionState {
	return connectionState{
		transaction: idleTransactionState,
		mode:        connectionMode{kind: readyConnectionMode},
	}
}

// enterExtendedMode enters an extended-query batch unless another exclusive operation owns the connection.
func (s *connectionState) enterExtendedMode() {
	if s.mode.kind == readyConnectionMode {
		s.mode = connectionMode{kind: extendedQueryConnectionMode}
	}
}

// finishExtended returns the connection to normal command dispatch.
func (s *connectionState) finishExtended() {
	s.mode = connectionMode{kind: readyConnectionMode}
}

// discardUntilSync rejects the remainder of an extended-query batch.
func (s *connectionState) discardUntilSync() {
	s.mode = connectionMode{kind: discardUntilSyncConnectionMode}
}

// closeConnection transitions the connection into its terminal mode.
func (s *connectionState) closeConnection() {
	s.mode = connectionMode{kind: closingConnectionMode}
}

// beginCopy transfers exclusive protocol ownership to a valid COPY FROM STDIN operation.
func (s *connectionState) beginCopy(copyState *copyInState) bool {
	if copyState == nil {
		return false
	}
	switch s.mode.kind {
	case readyConnectionMode, extendedQueryConnectionMode:
		s.mode = connectionMode{kind: copyInConnectionMode, copy: copyState}
		return true
	default:
		return false
	}
}

// finishCopy releases COPY ownership only when copyState is the active operation.
func (s *connectionState) finishCopy(copyState *copyInState) bool {
	if s.mode.kind != copyInConnectionMode || s.mode.copy != copyState {
		return false
	}
	s.mode = connectionMode{kind: readyConnectionMode}
	return true
}
