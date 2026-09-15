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

// protocolStateKind identifies the mutually exclusive frontend protocol state owning the connection.
type protocolStateKind byte

const (
	invalidProtocolState protocolStateKind = iota
	readyProtocolState
	extendedQueryProtocolState
	discardUntilSyncProtocolState
	copyInProtocolState
	closingProtocolState
)

// protocolState records the active protocol state and its state-specific COPY payload.
type protocolState struct {
	kind protocolStateKind
	copy *copyInState
}

// valid reports whether the state kind and optional COPY payload agree.
func (s protocolState) valid() bool {
	switch s.kind {
	case copyInProtocolState:
		return s.copy != nil
	case readyProtocolState, extendedQueryProtocolState, discardUntilSyncProtocolState, closingProtocolState:
		return s.copy == nil
	default:
		return false
	}
}

// connectionState owns every transaction and protocol phase transition for a connection.
type connectionState struct {
	transaction transactionState
	protocol    protocolState
}

// newConnectionState returns the initial state for a newly authenticated connection.
func newConnectionState() connectionState {
	return connectionState{
		transaction: idleTransactionState,
		protocol:    protocolState{kind: readyProtocolState},
	}
}

// beginExtended enters an extended-query batch unless another exclusive operation owns the connection.
func (s *connectionState) beginExtended() {
	if s.protocol.kind == readyProtocolState {
		s.protocol = protocolState{kind: extendedQueryProtocolState}
	}
}

// finishExtended returns the connection to normal command dispatch.
func (s *connectionState) finishExtended() {
	s.protocol = protocolState{kind: readyProtocolState}
}

// discardUntilSync rejects the remainder of an extended-query batch.
func (s *connectionState) discardUntilSync() {
	s.protocol = protocolState{kind: discardUntilSyncProtocolState}
}

// closeProtocol transitions the connection into its terminal protocol state.
func (s *connectionState) closeProtocol() {
	s.protocol = protocolState{kind: closingProtocolState}
}

// beginCopy transfers exclusive protocol ownership to a valid COPY FROM STDIN operation.
func (s *connectionState) beginCopy(copyState *copyInState) bool {
	if copyState == nil {
		return false
	}
	switch s.protocol.kind {
	case readyProtocolState, extendedQueryProtocolState:
		s.protocol = protocolState{kind: copyInProtocolState, copy: copyState}
		return true
	default:
		return false
	}
}

// finishCopy releases COPY ownership only when copyState is the active operation.
func (s *connectionState) finishCopy(copyState *copyInState) bool {
	if s.protocol.kind != copyInProtocolState || s.protocol.copy != copyState {
		return false
	}
	s.protocol = protocolState{kind: readyProtocolState}
	return true
}
