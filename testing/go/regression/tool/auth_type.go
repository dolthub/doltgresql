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

import "github.com/jackc/pgx/v5/pgproto3"

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
