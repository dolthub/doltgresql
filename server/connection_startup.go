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
	"crypto/tls"
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"

	psql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/doltgresql/server/auth"
)

// handleStartup handles the entire startup routine, including SSL requests, authentication, etc. Returns false if the
// connection has been terminated, or if we should not proceed with the message loop.
func (h *ConnectionHandler) handleStartup() (bool, error) {
	startupMessage, err := h.backend.ReceiveStartupMessage()
	if err == io.EOF {
		return false, nil
	} else if err != nil {
		return false, errors.Errorf("error receiving startup message: %w", err)
	}

	switch sm := startupMessage.(type) {
	case *pgproto3.StartupMessage:
		if err = h.handleAuthentication(sm); err != nil {
			return false, err
		}
		// The first SQL can run during startup parameter selection. Install the
		// authenticated identity before that path creates or executes a plan.
		if err = h.doltgresHandler.sm.NewSession(context.Background(), h.mysqlConn); err != nil {
			return false, err
		}
		if err = auth.InitializeSessionIdentity(h.doltgresHandler.sm.GetSession(h.mysqlConn), h.mysqlConn.User); err != nil {
			return false, err
		}
		if err = h.sendClientStartupMessages(); err != nil {
			return false, err
		}
		if err = h.chooseInitialParameters(sm); err != nil {
			// Report invalid startup parameters explicitly; otherwise the client sees only an
			// unexpected EOF when the connection closes.
			_ = h.send(&pgproto3.ErrorResponse{
				Severity: string(ErrorResponseSeverity_Fatal),
				Code:     "22023", // invalid_parameter_value
				Message:  err.Error(),
				Routine:  "InitPostgres",
			})
			return false, err
		}
		return true, h.send(&pgproto3.ReadyForQuery{TxStatus: byte(ReadyForQueryTransactionIndicator_Idle)})
	case *pgproto3.SSLRequest:
		hasCertificate := len(certificate.Certificate) > 0
		performSSL := []byte("N")
		if hasCertificate {
			performSSL = []byte("S")
		}
		if _, err = h.Conn().Write(performSSL); err != nil {
			return false, errors.Errorf("error sending SSL request: %w", err)
		}
		if hasCertificate {
			h.setConn(tls.Server(h.Conn(), &tls.Config{Certificates: []tls.Certificate{certificate}}))
		}
		return h.handleStartup()
	case *pgproto3.GSSEncRequest:
		if _, err = h.Conn().Write([]byte("N")); err != nil {
			return false, errors.Errorf("error sending response to GSS Enc Request: %w", err)
		}
		return h.handleStartup()
	default:
		return false, errors.Errorf("terminating connection: unexpected start message: %#v", startupMessage)
	}
}

// sendClientStartupMessages sends introductory messages to the client and returns any error.
func (h *ConnectionHandler) sendClientStartupMessages() error {
	for _, status := range []*pgproto3.ParameterStatus{
		{Name: "server_version", Value: "15.17"},
		{Name: "client_encoding", Value: "UTF8"},
		{Name: "standard_conforming_strings", Value: "on"},
		{Name: "in_hot_standby", Value: "off"},
	} {
		if err := h.send(status); err != nil {
			return err
		}
	}
	return h.send(&pgproto3.BackendKeyData{
		ProcessID: processID,
		SecretKey: make([]byte, 4), // TODO: this should represent an ID that can uniquely identify this connection, so that CancelRequest will work
	})
}

// chooseInitialParameters attempts to choose the initial parameter settings for the connection.
func (h *ConnectionHandler) chooseInitialParameters(startupMessage *pgproto3.StartupMessage) error {
	postgresParser := psql.PostgresParser{}
	db, ok := startupMessage.Parameters["database"]
	if !ok || len(db) == 0 {
		db = h.mysqlConn.User
	}
	useStmt := fmt.Sprintf("SET database TO '%s';", db)
	parsed, err := postgresParser.ParseSimple(useStmt)
	if err != nil {
		return err
	}
	if err = h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, useStmt, parsed, func(_ *sql.Context, _ *Result) error { return nil }); err != nil {
		_ = h.send(&pgproto3.ErrorResponse{
			Severity: string(ErrorResponseSeverity_Fatal),
			Code:     "3D000",
			Message:  fmt.Sprintf(`database "%s" does not exist`, db),
			Routine:  "InitPostgres",
		})
		return err
	}
	// Settings from ALTER ROLE and ALTER DATABASE are applied before the client's parameters, so that the client's
	// parameters take precedence, which matches Postgres. Applying them requires the database to have been chosen.
	if err = h.applyRoleSettings(db); err != nil {
		return err
	}
	for name, value := range startupMessage.Parameters {
		switch strings.ToLower(name) {
		case "datestyle":
			if err = h.doltgresHandler.InitSessionParameterDefault(context.Background(), h.mysqlConn, "DateStyle", value); err != nil {
				return err
			}
		case "timezone":
			setStmt := fmt.Sprintf("SET timezone TO '%s';", strings.ReplaceAll(value, "'", "''"))
			parsed, err = postgresParser.ParseSimple(setStmt)
			if err != nil {
				return err
			}
			if err = h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, setStmt, parsed, func(_ *sql.Context, _ *Result) error { return nil }); err != nil {
				return err
			}
		}
	}
	return nil
}

// applyRoleSettings applies the settings from ALTER ROLE ... SET and ALTER DATABASE ... SET that target the session's
// user and database. As in Postgres, a setting that cannot be applied results in a warning rather than a failure to
// connect, since otherwise a bad setting could prevent the role from ever connecting to fix it.
func (h *ConnectionHandler) applyRoleSettings(db string) error {
	postgresParser := psql.PostgresParser{}
	for _, setting := range auth.SessionRoleSettings(h.mysqlConn.User, db) {
		setStmt := fmt.Sprintf("SELECT set_config('%s', '%s', false);",
			strings.ReplaceAll(setting.Name, "'", "''"), strings.ReplaceAll(setting.Value, "'", "''"))
		parsed, err := postgresParser.ParseSimple(setStmt)
		if err == nil {
			err = h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, setStmt, parsed, func(_ *sql.Context, _ *Result) error { return nil })
		}
		if err != nil {
			if err = h.send(&pgproto3.NoticeResponse{
				Severity: string(ErrorResponseSeverity_Warning),
				Message:  fmt.Sprintf(`could not apply setting "%s": %s`, setting.Name, err.Error()),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
