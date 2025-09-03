// Copyright 2025 Dolthub, Inc.
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

package dumps

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"
)

// ImportQueryError contains both a query and its associated error.
type ImportQueryError struct {
	Query string
	Error string
}

// InterceptImportMessages sits between PSQL and Doltgres, returning all error messages that are encountered. As we rely
// on PSQL to handle the import process, we normally wouldn't be able to associate error messages with queries, as this
// information is not returned by PSQL itself. Therefore, we create our own connection to Doltgres, and a server that
// PSQL listens to. We then forward everything from PSQL to Doltgres, while inspecting the messages as they come and go.
func InterceptImportMessages(t *testing.T, doltgresPort int, breakpointQueries []string, triggerBreakpoint func(string)) (int, chan ImportQueryError) {
	psqlPort, err := sql.GetEmptyPort()
	require.NoError(t, err)
	qeChan := make(chan ImportQueryError)
	listener, err := server.NewListener("tcp", fmt.Sprintf("127.0.0.1:%d", psqlPort), "")
	if err != nil {
		t.Fatal(err)
		return psqlPort, qeChan
	}
	timer := time.NewTimer(5 * time.Second)
	timer.Stop()
	go func() {
		<-timer.C
		_ = listener.Close()
	}()

	go func() {
		for {
			psqlConn, err := listener.Accept()
			if err != nil {
				return
			}
			timer.Stop()
			terminate := &sync.WaitGroup{}
			terminate.Add(1)
			psqlConnBackend := pgproto3.NewBackend(psqlConn, psqlConn)
			doltgresConn, err := (&net.Dialer{}).Dial("tcp", fmt.Sprintf("127.0.0.1:%d", doltgresPort))
			if err != nil {
				fmt.Println(err)
				return
			}
			doltgresConnFrontend := pgproto3.NewFrontend(doltgresConn, doltgresConn)

			if err = handleStartup(t, psqlConnBackend, doltgresConnFrontend, psqlConn); err != nil {
				fmt.Println(err)
				return
			}
			createPassthrough(qeChan, terminate, psqlConnBackend, doltgresConnFrontend, triggerBreakpoint, breakpointQueries)
			terminate.Wait()
			_ = psqlConn.Close()
			_ = doltgresConn.Close()
			timer.Reset(5 * time.Second)
		}
	}()
	return psqlPort, qeChan
}

// handleStartup handles the startup messages.
func handleStartup(t *testing.T, psqlConnBackend *pgproto3.Backend, doltgresConnFrontend *pgproto3.Frontend, clientConn net.Conn) error {
StartupLoop:
	for {
		startupMessage, err := psqlConnBackend.ReceiveStartupMessage()
		if err != nil {
			return err
		}
		switch startupMessage := startupMessage.(type) {
		case *pgproto3.SSLRequest:
			if _, err = clientConn.Write([]byte{'N'}); err != nil {
				return err
			}
		case *pgproto3.StartupMessage:
			doltgresConnFrontend.Send(startupMessage)
			if err = doltgresConnFrontend.Flush(); err != nil {
				return err
			}
			response, err := doltgresConnFrontend.Receive()
			if err != nil {
				return err
			}
			if err = setAuthType(psqlConnBackend, response); err != nil {
				return err
			}
			psqlConnBackend.Send(response)
			if err = psqlConnBackend.Flush(); err != nil {
				return err
			}
			break StartupLoop
		default:
			t.Fatalf("unexpected startup message: %v", startupMessage)
		}
	}
	return nil
}

// setAuthType sets the client's authentication type depending on the message received from the server. This is
// necessary, as the client needs the proper context to know how to parse the returned messages.
func setAuthType(clientConnBackend *pgproto3.Backend, message pgproto3.BackendMessage) error {
	switch message.(type) {
	case *pgproto3.AuthenticationOk:
		return clientConnBackend.SetAuthType(pgproto3.AuthTypeOk)
	case *pgproto3.AuthenticationCleartextPassword:
		return clientConnBackend.SetAuthType(pgproto3.AuthTypeCleartextPassword)
	case *pgproto3.AuthenticationMD5Password:
		return clientConnBackend.SetAuthType(pgproto3.AuthTypeMD5Password)
	case *pgproto3.AuthenticationGSS:
		return clientConnBackend.SetAuthType(pgproto3.AuthTypeGSS)
	case *pgproto3.AuthenticationGSSContinue:
		return clientConnBackend.SetAuthType(pgproto3.AuthTypeGSSCont)
	case *pgproto3.AuthenticationSASL:
		return clientConnBackend.SetAuthType(pgproto3.AuthTypeSASL)
	case *pgproto3.AuthenticationSASLContinue:
		return clientConnBackend.SetAuthType(pgproto3.AuthTypeSASLContinue)
	case *pgproto3.AuthenticationSASLFinal:
		return clientConnBackend.SetAuthType(pgproto3.AuthTypeSASLFinal)
	default:
		return nil
	}
}

// createPassthrough creates the go routines that will read from and write to the connections.
func createPassthrough(qeChan chan ImportQueryError, terminate *sync.WaitGroup, psqlConnBackend *pgproto3.Backend, doltgresConnFrontend *pgproto3.Frontend, triggerBreakpoint func(string), breakpointQueries []string) {
	lastQuery := ""
	writeMutex := &sync.Mutex{}
	go func() {
		defer terminate.Done()
		for {
			psqlMessage, err := psqlConnBackend.Receive()
			if err != nil {
				errStr := err.Error()
				if errStr != "unexpected EOF" && !strings.HasSuffix(errStr, "use of closed network connection") {
					fmt.Println(err)
				}
				return
			}
			switch msg := psqlMessage.(type) {
			case *pgproto3.Query:
				writeMutex.Lock()
				if len(lastQuery) == 0 {
					lastQuery = msg.String
				}
				writeMutex.Unlock()
				for _, query := range breakpointQueries {
					if strings.HasPrefix(msg.String, query) {
						triggerBreakpoint(msg.String)
						break
					}
				}
			case *pgproto3.Terminate:
				return
			}
			doltgresConnFrontend.Send(psqlMessage)
			if err = doltgresConnFrontend.Flush(); err != nil {
				errStr := err.Error()
				if errStr != "unexpected EOF" && !strings.HasSuffix(errStr, "use of closed network connection") {
					fmt.Println(err)
				}
				return
			}
		}
	}()
	go func() {
		for {
			doltgresMessage, err := doltgresConnFrontend.Receive()
			if err != nil {
				errStr := err.Error()
				if errStr != "unexpected EOF" &&
					!strings.HasSuffix(errStr, "use of closed network connection") &&
					!strings.HasSuffix(errStr, "An existing connection was forcibly closed by the remote host.") {
					fmt.Println(err)
				}
				return
			}
			switch msg := doltgresMessage.(type) {
			case *pgproto3.ErrorResponse:
				writeMutex.Lock()
				if len(lastQuery) == 0 {
					qeChan <- ImportQueryError{
						Query: "UNKNOWN QUERY HAS ERRORED",
						Error: msg.Message,
					}
				} else {
					qeChan <- ImportQueryError{
						Query: lastQuery,
						Error: msg.Message,
					}
				}
				writeMutex.Unlock()
			case *pgproto3.ReadyForQuery:
				writeMutex.Lock()
				lastQuery = ""
				writeMutex.Unlock()
			default:
				if err = setAuthType(psqlConnBackend, doltgresMessage); err != nil {
					fmt.Println(err)
					return
				}
			}
			psqlConnBackend.Send(doltgresMessage)
			if err = psqlConnBackend.Flush(); err != nil {
				errStr := err.Error()
				if errStr != "unexpected EOF" &&
					!strings.HasSuffix(errStr, "use of closed network connection") &&
					!strings.HasSuffix(errStr, "An existing connection was forcibly closed by the remote host.") {
					fmt.Println(err)
				}
				return
			}
		}
	}()
}
