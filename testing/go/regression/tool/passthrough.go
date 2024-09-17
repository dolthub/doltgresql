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
	"strings"

	"github.com/jackc/pgx/v5/pgproto3"
)

// createPassthrough creates the go routines that will read from and write to the connections.
func createPassthrough(clientConnBackend *pgproto3.Backend, postgresConnFrontend *pgproto3.Frontend) {
	go func() {
		for {
			clientMessage, err := clientConnBackend.Receive()
			if err != nil {
				errStr := err.Error()
				if errStr != "unexpected EOF" && !strings.HasSuffix(errStr, "use of closed network connection") {
					fmt.Println(err)
				}
				return
			}
			clientMessage = DuplicateMessage(clientMessage).(pgproto3.FrontendMessage)
			if query, ok := clientMessage.(*pgproto3.Query); ok {
				clientMessage, err = RewriteCopyToLocal(query)
				if err != nil {
					panic(err)
				}
			}
			addMessage(clientMessage)
			if _, ok := clientMessage.(*pgproto3.Terminate); ok {
				terminate.Done()
				return
			}
			postgresConnFrontend.Send(clientMessage)
			if err = postgresConnFrontend.Flush(); err != nil {
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
			postgresMessage, err := postgresConnFrontend.Receive()
			if err != nil {
				errStr := err.Error()
				if errStr != "unexpected EOF" &&
					!strings.HasSuffix(errStr, "use of closed network connection") &&
					!strings.HasSuffix(errStr, "An existing connection was forcibly closed by the remote host.") {
					fmt.Println(err)
				}
				return
			}
			postgresMessage = DuplicateMessage(postgresMessage).(pgproto3.BackendMessage)
			addMessage(postgresMessage)
			if err = setAuthType(clientConnBackend, postgresMessage); err != nil {
				fmt.Println(err)
				return
			}
			clientConnBackend.Send(postgresMessage)
			if err = clientConnBackend.Flush(); err != nil {
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
