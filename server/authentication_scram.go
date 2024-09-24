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

package server

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
)

// SCRAM authentication is defined in RFC-5802:
// https://datatracker.ietf.org/doc/html/rfc5802

// These are mechanisms that are used for SASL authentication.
const (
	SASLMechanism_SCRAM_SHA_256      = "SCRAM-SHA-256"
	SASLMechanism_SCRAM_SHA_256_PLUS = "SCRAM-SHA-256-PLUS"
)

// SASLBindingFlag are the flags for gs2-cbind-flag, used in SASL authentication.
type SASLBindingFlag string

const (
	SASLBindingFlag_NoClientSupport        SASLBindingFlag = "n"
	SASLBindingFlag_AssumedNoServerSupport SASLBindingFlag = "y"
	SASLBindingFlag_Used                   SASLBindingFlag = "p"
)

// SASLInitial is the structured form of the input given by *pgproto3.SASLInitialResponse.
type SASLInitial struct {
	Flag     SASLBindingFlag
	BindName string // Only set when Flag is SASLBindingFlag_Used
	Binding  string // Base64 encoding of cbind-input
	Authzid  string // Authorization ID, currently ignored in favor of the startup message's username
	Username string // Prepared using SASLprep, currently ignored in favor of the startup message's username
	Nonce    string
	RawData  []byte // The bytes that were received in the message
}

// SASLContinue is the structured form of the output for *pgproto3.SASLInitialResponse.
type SASLContinue struct {
	Nonce      string
	Salt       string // Base64 encoded salt
	Iterations uint32
}

// SASLResponse is the structured form of the input given by *pgproto3.SASLResponse.
type SASLResponse struct {
	GS2Header   string
	Nonce       string
	ClientProof string // Base64 encoded salt
	RawData     []byte // The bytes that were received in the message
}

// handleAuthentication handles authentication for the given user
func (h *ConnectionHandler) handleAuthentication(startupMessage *pgproto3.StartupMessage) error {
	var user string
	var host string
	var ok bool
	if user, ok = startupMessage.Parameters["user"]; ok && len(user) > 0 {
		if h.Conn().RemoteAddr().Network() == "unix" {
			host = "localhost"
		} else {
			host, _, _ = net.SplitHostPort(h.Conn().RemoteAddr().String())
			if len(host) == 0 {
				host = "localhost"
			}
		}
	} else {
		user = "doltgres" // TODO: should we use this, or the default "postgres" since programs may default to it?
		host = "localhost"
	}
	h.mysqlConn.User = user
	h.mysqlConn.UserData = sql.MysqlConnectionUser{
		User: user,
		Host: host,
	}
	// In order to skip the rest of the code without the static analyzer complaining, we'll add this check that will
	// always fail.
	// TODO: remove me when implementing the rest of the users
	if _, ok := startupMessage.Parameters["will_obviously_not_exist"]; !ok {
		return h.send(&pgproto3.AuthenticationOk{})
	}
	// We only support one mechanism for now.
	if err := h.send(&pgproto3.AuthenticationSASL{
		AuthMechanisms: []string{
			SASLMechanism_SCRAM_SHA_256,
		},
	}); err != nil {
		return err
	}
	if err := h.backend.SetAuthType(pgproto3.AuthTypeSASL); err != nil {
		return err
	}
	var saslInitial SASLInitial
	var saslContinue SASLContinue
	var saslResponse SASLResponse
	for {
		initialResponse, err := h.backend.Receive()
		if err != nil {
			return err
		}
		switch response := initialResponse.(type) {
		case *pgproto3.SASLInitialResponse:
			saslInitial, err = readSASLInitial(response)
			if err != nil {
				return err
			}
			saslContinue = SASLContinue{
				Nonce:      saslInitial.Nonce + "L3rfcNHYJY1ZVvWVs7j", // TODO: create a unique nonce
				Salt:       "QSXCR+Q6sek8bf92",                        // TODO: read the salt for the user
				Iterations: 4096,
			}
			if err = h.send(saslContinue.Encode()); err != nil {
				return err
			}
			if err = h.backend.SetAuthType(pgproto3.AuthTypeSASLContinue); err != nil {
				return err
			}
		case *pgproto3.SASLResponse:
			saslResponse, err = readSASLResponse(saslInitial.Base64Header(), saslContinue.Nonce, response)
			if err != nil {
				return err
			}
			serverSignature, err := verifySASLClientProof(saslInitial, saslContinue, saslResponse)
			if err != nil {
				return err
			}
			if err = h.send(&pgproto3.AuthenticationSASLFinal{
				Data: []byte("v=" + serverSignature),
			}); err != nil {
				return err
			}
			return h.send(&pgproto3.AuthenticationOk{})
		default:
			return fmt.Errorf("unknown message type encountered during SASL authentication: %T", response)
		}
	}
}

// readSASLInitial reads the initial SASL response from the client.
func readSASLInitial(r *pgproto3.SASLInitialResponse) (SASLInitial, error) {
	if r.AuthMechanism != SASLMechanism_SCRAM_SHA_256 {
		return SASLInitial{}, fmt.Errorf("SASL mechanism not supported: %s", r.AuthMechanism)
	}
	saslInitial := SASLInitial{}
	sections := strings.Split(string(r.Data), ",")
	if len(sections) < 3 {
		return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: too few sections")
	}

	// gs2-cbind-flag is the first section
	gs2CbindFlag := sections[0]
	if len(gs2CbindFlag) == 0 {
		return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed gs2-cbind-flag")
	}
	switch gs2CbindFlag[0] {
	case 'n':
		saslInitial.Flag = SASLBindingFlag_NoClientSupport
	case 'p':
		if len(gs2CbindFlag) < 3 {
			return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed gs2-cbind-flag channel binding")
		}
		saslInitial.Flag = SASLBindingFlag_Used
		saslInitial.BindName = gs2CbindFlag[2:]
	case 'y':
		saslInitial.Flag = SASLBindingFlag_AssumedNoServerSupport
	default:
		return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed gs2-cbind-flag options (%c)", gs2CbindFlag[0])
	}

	// authzid is the second section
	authzid := sections[1]
	if len(authzid) > 0 {
		if len(authzid) < 3 {
			return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed authzid")
		}
		saslInitial.Authzid = authzid[2:]
	}

	// Read the gs2-header
	for i := 2; i < len(sections); i++ {
		if len(sections[i]) < 2 {
			return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed gs2-header")
		}
		switch sections[i][0] {
		case 'c':
			saslInitial.Binding = sections[i][2:]
		case 'n':
			saslInitial.Username = sections[i][2:]
		case 'r':
			saslInitial.Nonce = sections[i][2:]
		default:
			return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: unknown gs2-header option (%c)", sections[i][0])
		}
	}

	// Validate that all required options have been read
	if len(saslInitial.Nonce) == 0 {
		return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: missing nonce")
	}
	// Copy the message bytes, since the backend may re-use the slice for future responses
	saslInitial.RawData = make([]byte, len(r.Data))
	copy(saslInitial.RawData, r.Data)
	return saslInitial, nil
}

// readSASLResponse reads the second SASL response from the client.
func readSASLResponse(gs2EncodedHeader string, nonce string, r *pgproto3.SASLResponse) (SASLResponse, error) {
	saslResponse := SASLResponse{}
	for _, section := range strings.Split(string(r.Data), ",") {
		if len(section) < 3 {
			return SASLResponse{}, fmt.Errorf("invalid SASLResponse: attribute too small")
		}
		switch section[0] {
		case 'c':
			saslResponse.GS2Header = section[2:]
			if saslResponse.GS2Header != gs2EncodedHeader {
				return SASLResponse{}, fmt.Errorf("invalid SASLResponse: inconsistent GS2 header")
			}
		case 'p':
			saslResponse.ClientProof = section[2:]
		case 'r':
			saslResponse.Nonce = section[2:]
			if saslResponse.Nonce != nonce {
				return SASLResponse{}, fmt.Errorf("invalid SASLResponse: nonce does not match authentication session")
			}
		default:
			return SASLResponse{}, fmt.Errorf("invalid SASLResponse: unknown attribute (%c)", section[0])
		}
	}

	// Validate that all required options have been read
	if len(saslResponse.Nonce) == 0 {
		return SASLResponse{}, fmt.Errorf("invalid SASLResponse: missing nonce")
	}
	if len(saslResponse.ClientProof) == 0 {
		return SASLResponse{}, fmt.Errorf("invalid SASLResponse: missing nonce")
	}
	// Copy the message bytes, since the backend may re-use the slice for future responses
	saslResponse.RawData = make([]byte, len(r.Data))
	copy(saslResponse.RawData, r.Data)
	return saslResponse, nil
}

// verifySASLClientProof verifies that the proof given by the client in valid. Returns the base64-encoded
// ServerSignature, which verifies (to the client) that the server has proper access to the client's authentication
// information.
func verifySASLClientProof(saslInitial SASLInitial, saslContinue SASLContinue, saslResponse SASLResponse) (string, error) {
	// TODO: implement this
	return "", nil
}

// Base64Header returns the base64-encoded GS2 header and channel binding data.
func (si SASLInitial) Base64Header() string {
	sb := strings.Builder{}
	switch si.Flag {
	case SASLBindingFlag_NoClientSupport:
		sb.WriteString("n,")
	case SASLBindingFlag_AssumedNoServerSupport:
		sb.WriteString("y,")
	case SASLBindingFlag_Used:
		sb.WriteString(fmt.Sprintf("p=%s,", si.BindName))
	}
	sb.WriteString(si.Authzid)
	sb.WriteRune(',')
	return base64.StdEncoding.EncodeToString([]byte(sb.String()))
}

// Encode returns the struct as an AuthenticationSASLContinue message.
func (sc SASLContinue) Encode() *pgproto3.AuthenticationSASLContinue {
	return &pgproto3.AuthenticationSASLContinue{
		Data: []byte(fmt.Sprintf("r=%s,s=%s,i=%d", sc.Nonce, sc.Salt, sc.Iterations)),
	}
}
