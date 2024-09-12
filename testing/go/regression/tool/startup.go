// Copyright 2024 Dolthub, Inc.
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

package main

import (
	"fmt"
	"net"

	"github.com/jackc/pgx/v5/pgproto3"
)

// handleStartup handles the startup messages.
func handleStartup(clientConnBackend *pgproto3.Backend, postgresConnFrontend *pgproto3.Frontend, clientConn net.Conn) error {
StartupLoop:
	for {
		startupMessage, err := clientConnBackend.ReceiveStartupMessage()
		if err != nil {
			return err
		}
		switch startupMessage := startupMessage.(type) {
		case *pgproto3.SSLRequest:
			if _, err = clientConn.Write([]byte{'N'}); err != nil {
				return err
			}
		case *pgproto3.StartupMessage:
			addMessage(startupMessage)
			postgresConnFrontend.Send(startupMessage)
			if err = postgresConnFrontend.Flush(); err != nil {
				return err
			}
			response, err := postgresConnFrontend.Receive()
			if err != nil {
				return err
			}
			if err = setAuthType(clientConnBackend, response); err != nil {
				return err
			}
			addMessage(response)
			clientConnBackend.Send(response)
			if err = clientConnBackend.Flush(); err != nil {
				return err
			}
			break StartupLoop
		default:
			return fmt.Errorf("Unexpected Startup Message: %v", startupMessage)
		}
	}
	return nil
}
