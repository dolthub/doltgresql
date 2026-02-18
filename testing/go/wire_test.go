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

package _go

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server"
)

// TestWireImplementation allows us to directly test what is received on the wire, ensuring that the wire protocol is
// correctly implemented.
func TestWireImplementation(t *testing.T) {
	RunWireScripts(t, []WireScriptTest{
		{
			Name: "Smoke Test",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT4 PRIMARY KEY);",
				"INSERT INTO test VALUES (7);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Query{String: "SELECT * FROM test;"},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("pk"),
									TableOID:             0,
									TableAttributeNumber: 0,
									DataTypeOID:          23,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.DataRow{Values: [][]byte{[]byte("7")}},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
	})
}

// IgnoreMessageParameters is used to ignore certain fields within a backend message, as they may not yet be implemented
// and therefore will return incorrect results (or variable results, such as with non-stable OIDs).
func IgnoreMessageParameters(message pgproto3.BackendMessage) pgproto3.BackendMessage {
	switch message := message.(type) {
	case *pgproto3.RowDescription:
		for i := range message.Fields {
			message.Fields[i].TableOID = 0
			message.Fields[i].TableAttributeNumber = 0
		}
		return message
	default:
		return message
	}
}

// WireScriptTest is used to test wire messages, ensuring that our wire protocol behaves as expected.
type WireScriptTest struct {
	// Name of the script.
	Name string
	// The database to create and use. If not provided, then it defaults to "postgres".
	Database string
	// The SQL statements to execute as setup, in order. Results are not checked, but statements must not error.
	SetUpScript []string
	// The set of assertions to make after setup, in order
	Assertions []WireScriptTestAssertion
	// When using RunScripts, setting this on one (or more) tests causes RunWireScripts to ignore all tests that have
	// this set to false (which is the default value). This allows a developer to easily "focus" on a specific test
	// without having to comment out other tests, pull it into a different function, etc. In addition, CI ensures that
	// this is false before passing, meaning this prevents the commented-out situation where the developer forgets to
	// uncomment their code.
	Focus bool
	// Skip is used to completely skip a test
	Skip bool
}

// WireScriptTestAssertion are the assertions upon which the script executes its main "testing" logic.
type WireScriptTestAssertion struct {
	// These are sent as a batch to the server
	Send []pgproto3.FrontendMessage
	// These are the expected results that are received from the server, and must match in both contents and order
	Receive []pgproto3.BackendMessage
	// This functions the same as Focus on WireScriptTest, except that it applies to the assertion
	Focus bool
	// This is used to skip an assertion
	Skip bool
}

// RawWireConnection is a connection that allows us to directly send and receive messages to a server.
type RawWireConnection struct {
	frontend   *pgproto3.Frontend
	connection net.Conn
	network    string
	timeout    time.Duration
	startup    *pgproto3.StartupMessage
	errChan    chan error
}

// NewRawWireConnection returns a new RawWireConnection.
func NewRawWireConnection(t *testing.T, host string, port int, timeout time.Duration) *RawWireConnection {
	network := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	connection, err := (&net.Dialer{}).Dial("tcp", network)
	require.NoError(t, err)
	c := &RawWireConnection{
		frontend:   pgproto3.NewFrontend(connection, connection),
		connection: connection,
		network:    network,
		timeout:    timeout,
		startup:    nil,
		errChan:    make(chan error),
	}
	c.init(t)
	return c
}

// Close closes the internal connection.
func (c *RawWireConnection) Close() {
	_ = c.connection.Close()
	close(c.errChan)
}

// EmptyReceiveBuffer empties the buffer used by Receive. Returns an error if the buffer was not empty.
func (c *RawWireConnection) EmptyReceiveBuffer() error {
	if c.frontend.ReadBufferLen() > 0 {
		for c.frontend.ReadBufferLen() > 0 {
			_, _ = c.frontend.Receive()
		}
		return errors.New("Doltgres sent additional messages after ReadyForQuery")
	}
	return nil
}

// Receive returns the next message from the backend.
func (c *RawWireConnection) Receive(t *testing.T) (pgproto3.BackendMessage, error) {
	var message pgproto3.BackendMessage
	go func() {
		var err error
		message, err = c.frontend.Receive()
		c.errChan <- err
	}()
	return message, c.handleErrorChannel(t, false)
}

// Send sends the given messages over the wire. If an error is returned, then the connection has been closed, and a new
// one should be opened.
func (c *RawWireConnection) Send(t *testing.T, messages ...pgproto3.FrontendMessage) error {
	if len(messages) == 0 {
		return nil
	}
	hasMessage := false
	for _, message := range messages {
		if message == nil {
			continue
		}
		hasMessage = true
		if startupMessage, ok := message.(*pgproto3.StartupMessage); ok {
			c.startup = startupMessage
		}
		c.frontend.Send(message)
	}
	if !hasMessage {
		return nil
	}
	go func() {
		c.errChan <- c.frontend.Flush()
	}()
	return c.handleErrorChannel(t, true)
}

// init handles the startup message and initial messages from the server.
func (c *RawWireConnection) init(t *testing.T) {
	err := c.Send(t, &pgproto3.StartupMessage{
		ProtocolVersion: 196608,
		Parameters: map[string]string{
			"timezone":         "PST8PDT",
			"user":             "postgres",
			"database":         "postgres",
			"options":          " -c intervalstyle=postgres_verbose",
			"application_name": "pg_regress",
			"client_encoding":  "WIN1252",
			"datestyle":        "Postgres, MDY",
		},
	})
	require.NoError(t, err)
StartupLoop:
	for {
		postgresMessage, err := c.Receive(t)
		require.NoError(t, err)
		switch response := postgresMessage.(type) {
		case *pgproto3.AuthenticationOk:
		case *pgproto3.BackendKeyData:
		case *pgproto3.ErrorResponse:
			t.Log(response.Message)
			t.FailNow()
		case *pgproto3.ParameterStatus:
		case *pgproto3.ReadyForQuery:
			break StartupLoop
		default:
			t.Logf("unknown StartupMessage response type: %T", response)
			t.FailNow()
		}
	}
}

// handleErrorChannel handles errors while ensuring that stuck queries do not cause an infinite loop via a timeout.
func (c *RawWireConnection) handleErrorChannel(t *testing.T, isSend bool) error {
	var err error
	select {
	case err = <-c.errChan:
	case <-time.After(c.timeout):
		if isSend {
			err = errors.New("timeout during Send")
		} else {
			err = errors.New("timeout during Receive")
		}
	}
	// On error, we must create a new connection since we cut the old one
	if err != nil {
		_ = c.connection.Close()
		connection, nErr := (&net.Dialer{}).Dial("tcp", c.network)
		if nErr != nil {
			panic(fmt.Errorf("Unable to create a new connection:\n%s\n\nOriginal error:\n%s", nErr.Error(), err.Error()))
		}
		c.connection = connection
		c.frontend = pgproto3.NewFrontend(connection, connection)
		c.init(t)
	}
	return err
}

// RunWireScripts runs the given collection of scripts.
func RunWireScripts(t *testing.T, scripts []WireScriptTest) {
	// First, we'll run through the scripts to check for the Focus variable. If it's true, then append it to the new slice.
	focusScripts := make([]WireScriptTest, 0, len(scripts))
	for _, script := range scripts {
		if script.Focus {
			// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
			if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
				panic(fmt.Sprintf("The wire script `%s` has Focus set to `true`. GitHub Actions requires that "+
					"all tests are run, which Focus circumvents, leading to this error. Please disable Focus on "+
					"all tests.", script.Name))
			}
			focusScripts = append(focusScripts, script)
		}
	}
	// If we have scripts with Focus set, then we replace the normal script slice with the new slice.
	if len(focusScripts) > 0 {
		scripts = focusScripts
	}
	// TODO: for now, our wire handler can't authenticate itself, so we disable it for these tests.
	//  This prevents things such as testing multiple users, so it should be implemented at some point.
	server.EnableAuthentication = false
	defer func() {
		server.EnableAuthentication = true
	}()

	for _, script := range scripts {
		t.Run(script.Name, func(t *testing.T) {
			if script.Skip {
				t.Skip()
			}

			scriptDatabase := script.Database
			if len(scriptDatabase) == 0 {
				scriptDatabase = "postgres"
			}
			port, err := sql.GetEmptyPort()
			require.NoError(t, err)
			ctx, conn, controller := CreateServerWithPort(t, scriptDatabase, port)
			defer func() {
				controller.Stop()
				err := controller.WaitForStop()
				require.NoError(t, err)
			}()
			for _, query := range script.SetUpScript {
				_, err = conn.Exec(ctx, query)
				require.NoError(t, err, "error running setup query: %s", query)
			}
			conn.Close(ctx)
			rawConn := NewRawWireConnection(t, "localhost", port, 10*time.Second)
			defer rawConn.Close()

			// With everything set up, we now check for Focus on the assertions
			assertions := script.Assertions
			// First, we'll run through the scripts to check for the Focus variable. If it's true, then append it to the new slice.
			focusAssertions := make([]WireScriptTestAssertion, 0, len(script.Assertions))
			for _, assertion := range script.Assertions {
				if assertion.Focus {
					// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
					if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
						panic("A wire assertion has Focus set to `true`. GitHub Actions requires that " +
							"all non-skipped assertions are run, which Focus circumvents, leading to this error. " +
							"Please disable Focus on all wire assertions.")
					}
					focusAssertions = append(focusAssertions, assertion)
				}
			}
			// If we have assertions with Focus set, then we replace the original slice with the new slice.
			if len(focusAssertions) > 0 {
				assertions = focusAssertions
			}

			// Run the assertions
			for assertionIdx, assertion := range assertions {
				t.Run(fmt.Sprintf("%d", assertionIdx), func(t *testing.T) {
					if assertion.Skip {
						t.Skip("Skip has been set in the assertion")
					}
					err = rawConn.Send(t, assertion.Send...)
					require.NoError(t, err)
					for _, expectedMessage := range assertion.Receive {
						actualMessage, err := rawConn.Receive(t)
						require.NoError(t, err)
						if !assert.Equal(t, IgnoreMessageParameters(expectedMessage), IgnoreMessageParameters(actualMessage)) {
							// If the assertion fails, then we have to sync to the ReadyForQuery message
							if _, ok := actualMessage.(*pgproto3.ReadyForQuery); !ok {
								for {
									actualMessage, err := rawConn.Receive(t)
									require.NoError(t, err)
									if _, ok = actualMessage.(*pgproto3.ReadyForQuery); ok {
										return
									}
								}
							}
						}
					}
					// We then ensure that there are no other messages that were not accounted for by the assertion
					// (which we consider an error)
					_ = assert.NoError(t, rawConn.EmptyReceiveBuffer())
				})
			}
		})
	}
}
