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
	"github.com/cockroachdb/errors"
	psql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"io"
	"strings"
)

// handleStartup handles the entire startup routine, including SSL requests, authentication, etc. Returns false if the
// connection has been terminated, or if we should not proceed with the message loop.
func (h *ConnectionHandler) handleStartup() (bool, error) {
	startupMessage, err := h.backend.ReceiveStartupMessage()
	if err == io.EOF {
		// Receiving EOF means that the connection has terminated, so we should just return
		return false, nil
	} else if err != nil {
		return false, errors.Errorf("error receiving startup message: %w", err)
	}

	switch sm := startupMessage.(type) {
	case *pgproto3.StartupMessage:
		if err = h.handleAuthentication(sm); err != nil {
			return false, err
		}
		if err = h.sendClientStartupMessages(); err != nil {
			return false, err
		}
		if err = h.chooseInitialParameters(sm); err != nil {
			// A startup parameter (e.g. datestyle, timezone) failed validation. Without an explicit
			// ErrorResponse here, the client only sees the connection drop as an unexpected EOF instead
			// of the reason its StartupMessage was rejected.
			_ = h.send(&pgproto3.ErrorResponse{
				Severity: string(ErrorResponseSeverity_Fatal),
				Code:     "22023", // invalid_parameter_value
				Message:  err.Error(),
				Routine:  "InitPostgres",
			})
			return false, err
		}
		return true, h.send(&pgproto3.ReadyForQuery{
			TxStatus: byte(ReadyForQueryTransactionIndicator_Idle),
		})
	case *pgproto3.SSLRequest:
		hasCertificate := len(certificate.Certificate) > 0
		var performSSL = []byte("N")
		if hasCertificate {
			performSSL = []byte("S")
		}
		_, err = h.Conn().Write(performSSL)
		if err != nil {
			return false, errors.Errorf("error sending SSL request: %w", err)
		}
		// If we have a certificate and the client has asked for SSL support, then we switch here.
		// This involves swapping out our underlying net connection for a new one.
		// We can't start in SSL mode, as the client does not attempt the handshake until after our response.
		if hasCertificate {
			h.setConn(tls.Server(h.Conn(), &tls.Config{
				Certificates: []tls.Certificate{certificate},
			}))
		}
		return h.handleStartup()
	case *pgproto3.GSSEncRequest:
		// we don't support GSSAPI
		_, err = h.Conn().Write([]byte("N"))
		if err != nil {
			return false, errors.Errorf("error sending response to GSS Enc Request: %w", err)
		}
		return h.handleStartup()
	default:
		return false, errors.Errorf("terminating connection: unexpected start message: %#v", startupMessage)
	}
}

// sendClientStartupMessages sends introductory messages to the client and returns any error
func (h *ConnectionHandler) sendClientStartupMessages() error {
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "server_version",
		Value: "15.17",
	}); err != nil {
		return err
	}
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "client_encoding",
		Value: "UTF8",
	}); err != nil {
		return err
	}
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "standard_conforming_strings",
		Value: "on",
	}); err != nil {
		return err
	}
	if err := h.send(&pgproto3.ParameterStatus{
		Name:  "in_hot_standby",
		Value: "off",
	}); err != nil {
		return err
	}
	return h.send(&pgproto3.BackendKeyData{
		ProcessID: processID,
		SecretKey: make([]byte, 4), // TODO: this should represent an ID that can uniquely identify this connection, so that CancelRequest will work
	})
}

// chooseInitialParameters attempts to choose the initial parameter settings for the connection,
// if one is specified in the startup message provided.
func (h *ConnectionHandler) chooseInitialParameters(startupMessage *pgproto3.StartupMessage) error {
	postgresParser := psql.PostgresParser{}
	for name, value := range startupMessage.Parameters {
		// TODO: handle other parameters defined in StartupMessage
		switch strings.ToLower(name) {
		case "datestyle":
			err := h.doltgresHandler.InitSessionParameterDefault(context.Background(), h.mysqlConn, "DateStyle", value)
			if err != nil {
				return err
			}
		case "timezone":
			// timezone is set via a real SET statement rather than InitSessionParameterDefault because we want
			// this value set for the current session, but NOT set as the default for all sessions.
			setStmt := fmt.Sprintf("SET timezone TO '%s';", strings.ReplaceAll(value, "'", "''"))
			parsed, err := postgresParser.ParseSimple(setStmt)
			if err != nil {
				return err
			}
			err = h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, setStmt, parsed, func(_ *sql.Context, _ *Result) error {
				return nil
			})
			if err != nil {
				return err
			}
		}
	}
	// Set the initial database. Postgres has no concept of a session without a current database: if the client
	// doesn't specify one, it defaults to the username (matching libpq). Either way, if the resolved database
	// doesn't exist we must reject the connection rather than proceed with a database-less session, which would
	// break assumptions throughout the engine.
	db, ok := startupMessage.Parameters["database"]
	if !ok || len(db) == 0 {
		db = h.mysqlConn.User
	}
	useStmt := fmt.Sprintf("SET database TO '%s';", db)
	parsed, err := postgresParser.ParseSimple(useStmt)
	if err != nil {
		return err
	}
	err = h.doltgresHandler.ComQuery(context.Background(), h.mysqlConn, useStmt, parsed, func(_ *sql.Context, _ *Result) error {
		return nil
	})
	if err != nil {
		_ = h.send(&pgproto3.ErrorResponse{
			Severity: string(ErrorResponseSeverity_Fatal),
			Code:     "3D000",
			Message:  fmt.Sprintf(`database "%s" does not exist`, db),
			Routine:  "InitPostgres",
		})
		return err
	}
	return nil
}
