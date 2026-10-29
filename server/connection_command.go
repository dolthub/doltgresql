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
	"github.com/dolthub/go-mysql-server/sql"
)

// Dolt's GC safepoint controller considers a session busy between sql.SessionCommandBegin and
// sql.SessionCommandEnd. While a session is between commands, a GC may visit the session's in-memory roots and then
// collect everything else, so every storage access made on behalf of a session must happen inside a command.
//
// A connection does work for its session in exactly these places, and each one begins and ends the command:
//
//   - handleStartup, from the creation of the session until the end of startup.
//   - receiveMessage, from the receipt of a frontend message until its response, including any error handling, has
//     been sent. Waiting for the next message happens outside of a command.
//   - DoltgresHandler.ConnectionClosed and DoltgresHandler.ComResetConnection, which release the locks of a session
//     that is about to end. These are called while no command is open.
//
// Nothing called from within those places begins a command of its own. DISCARD ALL is the one operation that
// replaces the session during a message. A session cannot end while it has a command open, so DISCARD ALL ends the
// message's command on the old session and begins it again on the new one.

// beginCommand begins a command on the connection's current session.
func (h *ConnectionHandler) beginCommand() error {
	session := h.doltgresHandler.sm.GetSession(h.mysqlConn)
	if err := sql.SessionCommandBegin(session); err != nil {
		return err
	}
	h.commandSession = session
	return nil
}

// endCommand ends the command begun by beginCommand. It does nothing when beginCommand has not been called or has
// failed, so that it can be deferred before the command begins.
func (h *ConnectionHandler) endCommand() {
	if h.commandSession != nil {
		sql.SessionCommandEnd(h.commandSession)
		h.commandSession = nil
	}
}
