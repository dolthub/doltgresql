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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/mysql"
)

// pgCommandLifecycle belongs to one ConnectionHandler's DoltgresHandler. That
// connection executes protocol messages serially. Nested engine commands borrow
// the message's ownership; they must not register a second outstanding command.
// No transaction state is changed here, and idle socket reads own no command.
type pgCommandLifecycle struct {
	owned   bool
	session sql.Session
}

// attach begins root protection for the current session unless already owned.
func (l *pgCommandLifecycle) attach(session sql.Session) error {
	if session == nil {
		return fmt.Errorf("PostgreSQL connection has no authenticated SQL session")
	}
	if l.session != nil {
		if l.session != session {
			return fmt.Errorf("PostgreSQL command belongs to another SQL session")
		}
		return nil
	}
	if err := sql.SessionCommandBegin(session); err != nil {
		return err
	}
	l.session = session
	return nil
}

// begin borrows an outer command or returns its idempotent owning release.
func (l *pgCommandLifecycle) begin(session sql.Session) (func(), error) {
	if err := l.attach(session); err != nil {
		return nil, err
	}
	if l.owned {
		return func() {}, nil
	}
	l.owned = true
	released := false
	return func() {
		if released {
			return
		}
		released = true
		defer func() { l.owned = false }()
		l.releaseSession()
	}, nil
}

// releaseSession ends protection before a session is removed or replaced.
// SessionManager.SetDB owns the replacement session's command notifications.
// End the old command before SessionEnd removes it; the outer message still
// owns cleanup and may attach the replacement again for later root access.
func (l *pgCommandLifecycle) releaseSession() {
	if l.session != nil {
		session := l.session
		l.session = nil
		sql.SessionCommandEnd(session)
	}
}

// beginCommand owns a protocol message or a directly invoked engine command.
func (h *DoltgresHandler) beginCommand(c *mysql.Conn) (func(), error) {
	return h.command.begin(h.sm.GetSession(c))
}

// requireCommand rejects unprotected root access and reattaches after reset.
func (h *DoltgresHandler) requireCommand(c *mysql.Conn) error {
	if !h.command.owned {
		return fmt.Errorf("PostgreSQL root access requires an active command")
	}
	return h.command.attach(h.sm.GetSession(c))
}
