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
	activeCopyFrom       *copyFromState
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

// beginDiscardUntilSyncMode rejects the remainder of an extended-query batch.
func (s *connectionState) beginDiscardUntilSyncMode() {
	s.mode = discardUntilSyncConnectionMode
}

// closeConnection transitions the connection into its terminal mode.
func (s *connectionState) beginCloseConnectionMode() {
	s.activeCopyFrom = nil
	s.mode = closingConnectionMode
}

// beginCopyInMode enters copy-in mode and records the COPY FROM  operation that owns the connection.
func (s *connectionState) beginCopyInMode(copyFrom *copyFromState) bool {
	if copyFrom == nil || s.activeCopyFrom != nil {
		return false
	}
	switch s.mode {
	case readyConnectionMode, extendedQueryConnectionMode:
		s.activeCopyFrom = copyFrom
		s.mode = copyInConnectionMode
		return true
	default:
		return false
	}
}

// finishCopyInMode releases COPY mode when copyFrom is the operation that currently owns the connection.
func (s *connectionState) finishCopyInMode(copyFrom *copyFromState) bool {
	if s.mode != copyInConnectionMode || s.activeCopyFrom != copyFrom {
		return false
	}
	s.activeCopyFrom = nil
	s.mode = readyConnectionMode
	return true
}

// resetExtendedQueryObjects removes every prepared statement and portal owned by the connection.
func (s *connectionState) resetExtendedQueryObjects() {
	s.extendedQueryObjects = newExtendedQueryObjects()
}
