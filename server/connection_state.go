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

// connectionMode identifies the mutually exclusive frontend mode owning the connection.
type connectionMode byte

const (
	readyConnectionMode connectionMode = iota
	extendedQueryConnectionMode
	discardUntilSyncConnectionMode
	copyInConnectionMode
	closingConnectionMode
)

// connectionState tracks the current state of a connection, including the transaction status,
// frontend protocol state, and extended query objects, such as prepared statements.
type connectionState struct {
	txState              transactionState
	mode                 connectionMode
	extendedQueryObjects extendedQueryObjects
	activeCopy           *copyInState
}

// newConnectionState returns the initial state for a newly authenticated connection.
func newConnectionState() connectionState {
	return connectionState{
		txState:              idleTransactionState,
		mode:                 readyConnectionMode,
		extendedQueryObjects: newExtendedQueryObjects(),
	}
}

// beginExtendedQueryMode enters an extended-query batch unless another exclusive operation owns the connection.
func (s *connectionState) beginExtendedQueryMode() {
	if s.mode == readyConnectionMode {
		s.mode = extendedQueryConnectionMode
	}
}

// finishExtendedQueryMode returns the connection to normal command dispatch.
func (s *connectionState) finishExtendedQueryMode() {
	s.mode = readyConnectionMode
}

// discardUntilSync rejects the remainder of an extended-query batch.
func (s *connectionState) discardUntilSync() {
	s.mode = discardUntilSyncConnectionMode
}

// closeConnection transitions the connection into its terminal mode.
func (s *connectionState) closeConnection() {
	s.activeCopy = nil
	s.mode = closingConnectionMode
}

// beginCopyMode transfers exclusive protocol ownership to a COPY operation. Returns true if
// the mode was successfully changed, otherwise returns false if the mode change was invalid.
func (s *connectionState) beginCopyMode(copyState *copyInState) bool {
	if copyState == nil || s.activeCopy != nil {
		return false
	}
	switch s.mode {
	case readyConnectionMode, extendedQueryConnectionMode:
		s.activeCopy = copyState
		s.mode = copyInConnectionMode
		return true
	default:
		return false
	}
}

// finishCopyMode releases COPY mode
func (s *connectionState) finishCopyMode(copyState *copyInState) bool {
	if s.mode != copyInConnectionMode || s.activeCopy != copyState {
		return false
	}
	s.activeCopy = nil
	s.mode = readyConnectionMode
	return true
}

// resetExtendedQueryObjects removes every prepared statement and portal owned by the connection.
func (s *connectionState) resetExtendedQueryObjects() {
	s.extendedQueryObjects = newExtendedQueryObjects()
}
